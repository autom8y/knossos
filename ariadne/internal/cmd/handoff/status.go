// Package handoff implements the ari handoff commands for agent handoff management.
package handoff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/hook/threadcontract"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/session"
)

// newStatusCmd creates the handoff status subcommand.
func newStatusCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current handoff status",
		Long: `Show the current handoff status for the active session,
including the current agent, pending handoffs, and recent transitions.

Examples:
  ari handoff status
  ari handoff status --session=session-20260104-120000-abc12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx)
		},
	}

	return cmd
}

// runStatus executes the handoff status command.
func runStatus(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

	// Get session ID
	sessionID, err := ctx.getSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.New(errors.CodeSessionNotFound, "No active session. Use 'ari session create' first.")
		printer.PrintError(err)
		return err
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	// Read events to determine current handoff state
	eventsPath := resolver.SessionEventsFile(sessionID)
	events, err := readAllEvents(eventsPath)
	if err != nil && !os.IsNotExist(err) {
		printer.VerboseLog("warn", "failed to read events", map[string]interface{}{"error": err.Error()})
	}

	// Determine current agent from most recent task_start event
	currentAgent := determineCurrentAgent(events, sessCtx.CurrentPhase)

	// Find last handoff event
	lastHandoff := findLastHandoff(events)

	// Count handoffs
	handoffCount := countHandoffs(events)

	// Build status output
	result := HandoffStatusOutput{
		SessionID:     sessionID,
		SessionStatus: string(sessCtx.Status),
		CurrentPhase:  sessCtx.CurrentPhase,
		CurrentAgent:  currentAgent,
		ActiveTeam:    sessCtx.ActiveTeam,
		HandoffCount:  handoffCount,
		LastHandoff:   lastHandoff,
		Initiative:    sessCtx.Initiative,
		CreatedAt:     sessCtx.CreatedAt.Format(time.RFC3339),
	}

	// Determine next expected handoff based on current phase
	result.NextExpectedHandoff = getNextExpectedHandoff(sessCtx.CurrentPhase)

	return printer.Print(result)
}

// readAllEvents reads both session events and thread contract events from events.jsonl.
func readAllEvents(path string) ([]GenericEvent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []GenericEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event GenericEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed lines
		}
		events = append(events, event)
	}

	return events, scanner.Err()
}

// GenericEvent represents either a session event or a thread contract event.
type GenericEvent struct {
	// Session event fields
	Timestamp string                 `json:"timestamp,omitempty"`
	Event     string                 `json:"event,omitempty"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	FromPhase string                 `json:"from_phase,omitempty"`
	ToPhase   string                 `json:"to_phase,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Thread contract event fields
	Ts      string                 `json:"ts,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Tool    string                 `json:"tool,omitempty"`
	Path    string                 `json:"path,omitempty"`
	Summary string                 `json:"summary,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// GetTimestamp returns the timestamp from either field.
func (e GenericEvent) GetTimestamp() string {
	if e.Ts != "" {
		return e.Ts
	}
	return e.Timestamp
}

// GetEventType returns the event type from either field.
func (e GenericEvent) GetEventType() string {
	if e.Type != "" {
		return e.Type
	}
	return e.Event
}

// determineCurrentAgent determines the current active agent from events.
func determineCurrentAgent(events []GenericEvent, currentPhase string) string {
	// Look for most recent task_start event
	for i := len(events) - 1; i >= 0; i-- {
		e := events[i]
		if e.Type == string(threadcontract.EventTypeTaskStart) {
			if agent, ok := e.Meta["agent"].(string); ok {
				return agent
			}
		}
		// Also check for HANDOFF_EXECUTED events
		if e.Event == "HANDOFF_EXECUTED" {
			if to := e.To; to != "" {
				return to
			}
		}
	}

	// Fallback: determine from phase
	return phaseToDefaultAgent(currentPhase)
}

// phaseToDefaultAgent returns the default agent for a phase.
func phaseToDefaultAgent(phase string) string {
	switch phase {
	case "requirements":
		return "requirements-analyst"
	case "design":
		return "architect"
	case "implementation":
		return "principal-engineer"
	case "validation", "qa":
		return "qa-adversary"
	default:
		return "unknown"
	}
}

// findLastHandoff finds the most recent handoff event.
func findLastHandoff(events []GenericEvent) *HandoffSummary {
	for i := len(events) - 1; i >= 0; i-- {
		e := events[i]

		// Check for HANDOFF_EXECUTED event
		if e.Event == "HANDOFF_EXECUTED" {
			summary := &HandoffSummary{
				Timestamp: e.Timestamp,
				ToAgent:   e.To,
			}
			if e.Metadata != nil {
				if from, ok := e.Metadata["from_phase"].(string); ok {
					summary.FromPhase = from
				}
				if to, ok := e.Metadata["target_phase"].(string); ok {
					summary.ToPhase = to
				}
				if artifact, ok := e.Metadata["artifact_id"].(string); ok {
					summary.ArtifactID = artifact
				}
			}
			return summary
		}

		// Check for task_end followed by task_start (implicit handoff)
		if e.Type == string(threadcontract.EventTypeTaskEnd) {
			summary := &HandoffSummary{
				Timestamp: e.GetTimestamp(),
			}
			if e.Meta != nil {
				if agent, ok := e.Meta["agent"].(string); ok {
					summary.FromAgent = agent
				}
			}
			return summary
		}
	}
	return nil
}

// countHandoffs counts the number of handoff events.
func countHandoffs(events []GenericEvent) int {
	count := 0
	for _, e := range events {
		if e.Event == "HANDOFF_EXECUTED" ||
			e.Type == string(threadcontract.EventTypeTaskEnd) {
			count++
		}
	}
	return count
}

// getNextExpectedHandoff returns the next expected handoff based on current phase.
func getNextExpectedHandoff(phase string) string {
	switch phase {
	case "requirements":
		return "requirements-analyst -> architect"
	case "design":
		return "architect -> principal-engineer"
	case "implementation":
		return "principal-engineer -> qa-adversary"
	case "validation", "qa":
		return "qa-adversary -> orchestrator (complete)"
	default:
		return "unknown"
	}
}

// HandoffSummary represents a summary of a handoff event.
type HandoffSummary struct {
	Timestamp  string `json:"timestamp"`
	FromAgent  string `json:"from_agent,omitempty"`
	ToAgent    string `json:"to_agent,omitempty"`
	FromPhase  string `json:"from_phase,omitempty"`
	ToPhase    string `json:"to_phase,omitempty"`
	ArtifactID string `json:"artifact_id,omitempty"`
}

// HandoffStatusOutput represents the output of handoff status.
type HandoffStatusOutput struct {
	SessionID           string          `json:"session_id"`
	SessionStatus       string          `json:"session_status"`
	CurrentPhase        string          `json:"current_phase"`
	CurrentAgent        string          `json:"current_agent"`
	ActiveTeam          string          `json:"active_team"`
	Initiative          string          `json:"initiative"`
	CreatedAt           string          `json:"created_at"`
	HandoffCount        int             `json:"handoff_count"`
	LastHandoff         *HandoffSummary `json:"last_handoff,omitempty"`
	NextExpectedHandoff string          `json:"next_expected_handoff"`
}

// Text implements output.Textable for HandoffStatusOutput.
func (h HandoffStatusOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Session: %s\n", h.SessionID))
	b.WriteString(fmt.Sprintf("Status: %s\n", h.SessionStatus))
	b.WriteString(fmt.Sprintf("Initiative: %s\n", h.Initiative))
	b.WriteString(fmt.Sprintf("Team: %s\n", h.ActiveTeam))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Current Phase: %s\n", h.CurrentPhase))
	b.WriteString(fmt.Sprintf("Current Agent: %s\n", h.CurrentAgent))
	b.WriteString(fmt.Sprintf("Handoff Count: %d\n", h.HandoffCount))
	b.WriteString("\n")

	if h.LastHandoff != nil {
		b.WriteString("Last Handoff:\n")
		b.WriteString(fmt.Sprintf("  Time: %s\n", h.LastHandoff.Timestamp))
		if h.LastHandoff.FromAgent != "" {
			b.WriteString(fmt.Sprintf("  From: %s\n", h.LastHandoff.FromAgent))
		}
		if h.LastHandoff.ToAgent != "" {
			b.WriteString(fmt.Sprintf("  To: %s\n", h.LastHandoff.ToAgent))
		}
		if h.LastHandoff.ArtifactID != "" {
			b.WriteString(fmt.Sprintf("  Artifact: %s\n", h.LastHandoff.ArtifactID))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("Next Expected: %s\n", h.NextExpectedHandoff))

	return b.String()
}

// Ensure Textable interface is implemented
var _ output.Textable = HandoffStatusOutput{}
