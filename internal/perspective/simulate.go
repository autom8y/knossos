package perspective

import (
	"strings"
)

// simulateKeywordMap maps common natural language keywords to CC tool names.
var simulateKeywordMap = map[string][]string{
	"read":     {"Read"},
	"file":     {"Read", "Write", "Edit", "Glob"},
	"write":    {"Write"},
	"edit":     {"Edit"},
	"search":   {"Grep", "Glob", "WebSearch"},
	"find":     {"Grep", "Glob"},
	"grep":     {"Grep"},
	"glob":     {"Glob"},
	"bash":     {"Bash"},
	"shell":    {"Bash"},
	"command":  {"Bash"},
	"run":      {"Bash"},
	"execute":  {"Bash"},
	"terminal": {"Bash"},
	"web":      {"WebSearch", "WebFetch"},
	"fetch":    {"WebFetch"},
	"browse":   {"WebSearch", "WebFetch"},
	"url":      {"WebFetch"},
	"todo":     {"TodoWrite", "TodoRead"},
	"task":     {"Task"},
	"delegate": {"Task"},
	"agent":    {"Task"},
	"notebook": {"NotebookEdit"},
	"jupyter":  {"NotebookEdit"},
	"skill":    {"Skill"},
	"ask":      {"AskUserQuestion"},
	"question": {"AskUserQuestion"},
	"user":     {"AskUserQuestion"},
}

// RunSimulate evaluates a natural language prompt against the agent's resolved
// perspective to predict what the agent can and cannot attempt.
func RunSimulate(doc *PerspectiveDocument, prompt string) *SimulateOverlay {
	overlay := &SimulateOverlay{
		Prompt: prompt,
	}

	keywords := tokenize(prompt)

	// Extract layer data
	l2 := getLayerData[*PerceptionData](doc, "L2")
	l3 := getLayerData[*CapabilityData](doc, "L3")
	l4 := getLayerData[*ConstraintData](doc, "L4")
	l6 := getLayerData[*PositionData](doc, "L6")

	// Build lookup sets
	var toolSet map[string]bool
	if l3 != nil {
		toolSet = perceptionBuildSet(l3.Tools)
	}
	var disallowedSet map[string]bool
	if l4 != nil {
		disallowedSet = perceptionBuildSet(l4.DisallowedTools)
	}

	// --- Tool matching ---
	toolMatchSet := make(map[string]SimulateMatch) // dedupe by tool name
	for _, kw := range keywords {
		// Exact tool name match (case-insensitive)
		for tool := range knownChannelTools {
			if strings.EqualFold(kw, tool) {
				toolMatchSet[tool] = SimulateMatch{Name: tool, MatchType: "exact", Relevance: "high"}
			}
		}
		// Keyword map match
		if tools, ok := simulateKeywordMap[kw]; ok {
			for _, tool := range tools {
				if _, already := toolMatchSet[tool]; !already {
					toolMatchSet[tool] = SimulateMatch{Name: tool, MatchType: "keyword", Relevance: "medium"}
				}
			}
		}
	}

	// Classify tool matches as can/cannot attempt
	for _, match := range toolMatchSet {
		overlay.ToolMatches = append(overlay.ToolMatches, match)
		switch {
		case disallowedSet[match.Name]:
			overlay.CannotAttempt = append(overlay.CannotAttempt, match.Name+" (disallowed)")
		case toolSet[match.Name]:
			overlay.CanAttempt = append(overlay.CanAttempt, match.Name)
		default:
			overlay.CannotAttempt = append(overlay.CannotAttempt, match.Name+" (not in tools)")
		}
	}

	// --- Skill matching ---
	if l2 != nil {
		allSkills := make([]string, 0)
		allSkills = append(allSkills, l2.ExplicitSkills...)
		allSkills = append(allSkills, l2.PolicyInjectedSkills...)
		allSkills = append(allSkills, l2.PolicyReferencedSkills...)
		allSkills = append(allSkills, l2.OnDemandSkills...)

		for _, skill := range allSkills {
			for _, kw := range keywords {
				if strings.Contains(strings.ToLower(skill), kw) {
					overlay.SkillMatches = append(overlay.SkillMatches, SimulateMatch{
						Name:      skill,
						MatchType: "partial",
						Relevance: "medium",
					})
					break // one match per skill is enough
				}
			}
		}
	}

	// --- Constraint hits ---
	if l4 != nil && l4.BehavioralContract != nil {
		for _, mustNot := range l4.BehavioralContract.MustNot {
			for _, kw := range keywords {
				if strings.Contains(strings.ToLower(mustNot), kw) {
					overlay.ConstraintHits = append(overlay.ConstraintHits, SimulateMatch{
						Name:      mustNot,
						MatchType: "partial",
						Relevance: "high",
					})
					break
				}
			}
		}
	}

	// --- Handoff needed ---
	if l6 != nil && l6.PhaseSuccessor != "" {
		// If prompt seems to be about the next phase's work, suggest handoff
		overlay.HandoffNeeded = []string{}
		if l6.PhaseSuccessor != "" {
			// Simple heuristic: if we have more cannot_attempt than can_attempt, suggest handoff
			if len(overlay.CannotAttempt) > len(overlay.CanAttempt) && len(overlay.CannotAttempt) > 0 {
				overlay.HandoffNeeded = append(overlay.HandoffNeeded,
					l6.PhaseSuccessor+" (agent lacks capabilities for matched tools)")
			}
		}
	}

	// Normalize nil slices
	if overlay.ToolMatches == nil {
		overlay.ToolMatches = []SimulateMatch{}
	}
	if overlay.SkillMatches == nil {
		overlay.SkillMatches = []SimulateMatch{}
	}
	if overlay.ConstraintHits == nil {
		overlay.ConstraintHits = []SimulateMatch{}
	}
	if overlay.CanAttempt == nil {
		overlay.CanAttempt = []string{}
	}
	if overlay.CannotAttempt == nil {
		overlay.CannotAttempt = []string{}
	}
	if overlay.HandoffNeeded == nil {
		overlay.HandoffNeeded = []string{}
	}

	return overlay
}

// tokenize splits a prompt into lowercase keywords for matching.
func tokenize(prompt string) []string {
	// Split on whitespace and common punctuation
	words := strings.FieldsFunc(strings.ToLower(prompt), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_'
	})
	// Deduplicate
	seen := make(map[string]bool, len(words))
	result := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) > 1 && !seen[w] { // skip single-char tokens
			seen[w] = true
			result = append(result, w)
		}
	}
	return result
}
