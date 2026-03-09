package suggest

import "fmt"

// maxSuggestionsPerEvent caps the number of suggestions returned per event
// to avoid overwhelming the model with noise.
const maxSuggestionsPerEvent = 2

// SessionStartSuggestions generates suggestions for a SessionStart event.
// Returns nil when no suggestions apply (common case for no-session starts).
func SessionStartSuggestions(input *SessionInput) []Suggestion {
	if input == nil || input.SessionID == "" {
		return nil
	}

	var suggestions []Suggestion

	// Resumed from park suggestion (highest priority for session start)
	if input.ParkSource != "" {
		suggestions = append(suggestions, Suggestion{
			Kind:   KindSessionStart,
			Text:   "Resuming parked session. Review parked state above, then continue.",
			Reason: "session was parked and is being resumed",
		})
	}

	// Initiative + complexity suggestion
	if input.Initiative != "" && len(suggestions) < maxSuggestionsPerEvent {
		switch input.Complexity {
		case "TASK":
			suggestions = append(suggestions, Suggestion{
				Kind:   KindSessionStart,
				Text:   fmt.Sprintf("Your initiative is %q. Consider starting with /task.", input.Initiative),
				Action: "/task",
				Reason: "TASK complexity detected",
			})
		case "MODULE":
			suggestions = append(suggestions, Suggestion{
				Kind:   KindSessionStart,
				Text:   fmt.Sprintf("Your initiative is %q. Consider starting with /sprint or /task.", input.Initiative),
				Action: "/sprint",
				Reason: "MODULE complexity detected",
			})
		case "INITIATIVE":
			suggestions = append(suggestions, Suggestion{
				Kind:   KindSessionStart,
				Text:   fmt.Sprintf("Your initiative is %q. Consider starting with /10x for full orchestration.", input.Initiative),
				Action: "/10x",
				Reason: "INITIATIVE complexity detected",
			})
		}
	}

	// Strand suggestion
	if input.StrandCount > 0 && len(suggestions) < maxSuggestionsPerEvent {
		suggestions = append(suggestions, Suggestion{
			Kind:   KindSessionStart,
			Text:   fmt.Sprintf("This session has %d strand(s). Use /fray to manage parallel work.", input.StrandCount),
			Action: "/fray",
			Reason: "session has active strands",
		})
	}

	if len(suggestions) > maxSuggestionsPerEvent {
		suggestions = suggestions[:maxSuggestionsPerEvent]
	}

	return suggestions
}

// PhaseTransitionSuggestions generates suggestions when a phase change is detected.
// Returns nil when previousPhase == currentPhase or when either is empty.
func PhaseTransitionSuggestions(input *PhaseTransitionInput) []Suggestion {
	if input == nil {
		return nil
	}
	if input.PreviousPhase == "" || input.CurrentPhase == "" {
		return nil
	}
	if input.PreviousPhase == input.CurrentPhase {
		return nil
	}

	transition := input.PreviousPhase + " -> " + input.CurrentPhase

	var text string
	var action string

	switch {
	case input.PreviousPhase == "requirements" && input.CurrentPhase == "design":
		text = "Requirements phase complete. The architect agent creates TDDs and ADRs at design boundaries."
	case input.PreviousPhase == "design" && input.CurrentPhase == "implementation":
		text = "Design phase complete. The principal-engineer implements from TDD specifications."
	case input.PreviousPhase == "implementation" && input.CurrentPhase == "validation":
		text = "Implementation phase complete. The qa-adversary validates against PRD acceptance criteria."
	case input.PreviousPhase == "validation":
		text = "Validation phase complete. Consider /wrap if all criteria pass, or loop back if defects remain."
		action = "/wrap"
	default:
		text = fmt.Sprintf("Phase transitioned: %s.", transition)
	}

	return []Suggestion{{
		Kind:   KindPhaseTransition,
		Text:   text,
		Action: action,
		Reason: fmt.Sprintf("phase changed from %s to %s", input.PreviousPhase, input.CurrentPhase),
	}}
}

// SubagentStopSuggestions generates suggestions after a subagent completes.
// Returns nil when the completed agent has no actionable follow-up.
func SubagentStopSuggestions(input *SubagentInput) []Suggestion {
	if input == nil || input.AgentName == "" {
		return nil
	}

	var text string
	var action string

	switch input.AgentName {
	case "qa-adversary":
		text = "QA completed. Review findings, address defects, or /wrap if done."
		action = "/wrap"
	case "architect":
		text = "Architecture design complete. Review TDD, then proceed to implementation."
	case "requirements-analyst":
		text = "Requirements gathered. Review PRD, then proceed to design."
	case "principal-engineer":
		text = "Implementation complete. Proceed to validation with qa-adversary."
	default:
		return nil
	}

	return []Suggestion{{
		Kind:   KindSubagentComplete,
		Text:   text,
		Action: action,
		Reason: fmt.Sprintf("%s agent completed", input.AgentName),
	}}
}

// OrphanHygieneSuggestions generates suggestions based on Naxos triage state.
// Returns nil when input is nil, TotalTriaged is 0, or no actionable severity is present.
// Naxos suggestions are lower priority than session-start suggestions; callers should
// append these after other suggestions and respect the overall cap.
func OrphanHygieneSuggestions(input *NaxosInput) []Suggestion {
	if input == nil || input.TotalTriaged == 0 {
		return nil
	}

	var suggestions []Suggestion

	// CRITICAL orphans have the highest urgency — surface first.
	if n := input.BySeverity["CRITICAL"]; n > 0 {
		sessionWord := "session"
		if n != 1 {
			sessionWord = "sessions"
		}
		suggestions = append(suggestions, Suggestion{
			Kind:   KindOrphanHygiene,
			Text:   fmt.Sprintf("%d critical orphaned %s need attention. Run /naxos to triage.", n, sessionWord),
			Action: "/naxos",
			Reason: fmt.Sprintf("%d CRITICAL orphaned sessions found in triage artifact", n),
		})
	}

	// HIGH orphans are surfaced if there is room under the cap.
	if n := input.BySeverity["HIGH"]; n > 0 && len(suggestions) < maxSuggestionsPerEvent {
		sessionWord := "session"
		if n != 1 {
			sessionWord = "sessions"
		}
		suggestions = append(suggestions, Suggestion{
			Kind:   KindOrphanHygiene,
			Text:   fmt.Sprintf("%d high-priority orphaned %s. Consider /naxos.", n, sessionWord),
			Action: "/naxos",
			Reason: fmt.Sprintf("%d HIGH orphaned sessions found in triage artifact", n),
		})
	}

	if len(suggestions) > maxSuggestionsPerEvent {
		suggestions = suggestions[:maxSuggestionsPerEvent]
	}

	return suggestions
}

// BudgetWarningSuggestions generates suggestions when budget thresholds are crossed.
// Returns nil when count is below all thresholds.
func BudgetWarningSuggestions(input *SessionInput) []Suggestion {
	if input == nil {
		return nil
	}

	// Park threshold takes priority over warn threshold
	if input.ParkThreshold > 0 && input.ToolCount >= input.ParkThreshold {
		return []Suggestion{{
			Kind:   KindBudgetWarning,
			Text:   "Session is deep. Recommend /park now or /handoff to continue in a fresh context.",
			Action: "/park",
			Reason: fmt.Sprintf("tool count %d reached park threshold %d", input.ToolCount, input.ParkThreshold),
		}}
	}

	if input.WarnThreshold > 0 && input.ToolCount >= input.WarnThreshold {
		return []Suggestion{{
			Kind:   KindBudgetWarning,
			Text:   "Consider /park to preserve session state before context degrades.",
			Action: "/park",
			Reason: fmt.Sprintf("tool count %d reached warn threshold %d", input.ToolCount, input.WarnThreshold),
		}}
	}

	return nil
}
