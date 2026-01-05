// Package handoff implements the ari handoff commands for agent handoff management.
package handoff

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/hook/threadcontract"
	"github.com/autom8y/ariadne/internal/output"
)

// historyOptions holds options for the history command.
type historyOptions struct {
	limit int
}

// newHistoryCmd creates the handoff history subcommand.
func newHistoryCmd(ctx *cmdContext) *cobra.Command {
	var opts historyOptions

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show handoff history for a session",
		Long: `Show the history of handoffs that have occurred in the
current or specified session, queried from events.jsonl.

The history includes task_start, task_end, and HANDOFF_EXECUTED events
that represent agent transitions during the session.

Examples:
  ari handoff history
  ari handoff history --limit=10
  ari handoff history --session=session-20260104-120000-abc12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(ctx, opts)
		},
	}

	cmd.Flags().IntVar(&opts.limit, "limit", 0, "Limit number of history entries (0 = unlimited)")

	return cmd
}

// runHistory executes the handoff history command.
func runHistory(ctx *cmdContext, opts historyOptions) error {
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

	// Read events from events.jsonl
	eventsPath := resolver.SessionEventsFile(sessionID)
	events, err := readAllEvents(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No events file yet - return empty history
			result := HandoffHistoryOutput{
				SessionID: sessionID,
				Entries:   []HandoffHistoryEntry{},
				Total:     0,
			}
			return printer.Print(result)
		}
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to read events", err))
		return err
	}

	// Filter to handoff-related events
	var entries []HandoffHistoryEntry
	for _, e := range events {
		entry := eventToHistoryEntry(e)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	// Apply limit if specified
	total := len(entries)
	if opts.limit > 0 && len(entries) > opts.limit {
		entries = entries[len(entries)-opts.limit:]
	}

	result := HandoffHistoryOutput{
		SessionID: sessionID,
		Entries:   entries,
		Total:     total,
		Showing:   len(entries),
	}

	return printer.Print(result)
}

// eventToHistoryEntry converts an event to a history entry if it's handoff-related.
func eventToHistoryEntry(e GenericEvent) *HandoffHistoryEntry {
	eventType := e.GetEventType()
	timestamp := e.GetTimestamp()

	switch eventType {
	case string(threadcontract.EventTypeTaskStart):
		entry := &HandoffHistoryEntry{
			Timestamp: timestamp,
			EventType: "task_start",
			Summary:   e.Summary,
		}
		if e.Meta != nil {
			if agent, ok := e.Meta["agent"].(string); ok {
				entry.Agent = agent
				entry.Summary = fmt.Sprintf("Task started for %s", agent)
			}
			if desc, ok := e.Meta["description"].(string); ok {
				entry.Description = desc
			}
		}
		return entry

	case string(threadcontract.EventTypeTaskEnd):
		entry := &HandoffHistoryEntry{
			Timestamp: timestamp,
			EventType: "task_end",
			Summary:   e.Summary,
		}
		if e.Meta != nil {
			if agent, ok := e.Meta["agent"].(string); ok {
				entry.Agent = agent
				entry.Summary = fmt.Sprintf("Task completed by %s", agent)
			}
			if status, ok := e.Meta["status"].(string); ok {
				entry.Status = status
			}
			if throughline, ok := e.Meta["throughline"].(string); ok {
				entry.Throughline = throughline
			}
			if artifacts, ok := e.Meta["artifacts"].([]interface{}); ok {
				for _, a := range artifacts {
					if s, ok := a.(string); ok {
						entry.Artifacts = append(entry.Artifacts, s)
					}
				}
			}
			if durationMs, ok := e.Meta["duration_ms"].(float64); ok {
				entry.DurationMs = int64(durationMs)
			}
		}
		return entry

	case "HANDOFF_EXECUTED":
		entry := &HandoffHistoryEntry{
			Timestamp: timestamp,
			EventType: "handoff_executed",
			ToAgent:   e.To,
		}
		if e.Metadata != nil {
			if artifact, ok := e.Metadata["artifact_id"].(string); ok {
				entry.Artifacts = []string{artifact}
			}
			if fromPhase, ok := e.Metadata["from_phase"].(string); ok {
				entry.FromPhase = fromPhase
			}
			if toPhase, ok := e.Metadata["target_phase"].(string); ok {
				entry.ToPhase = toPhase
			}
		}
		entry.Summary = fmt.Sprintf("Handoff to %s", e.To)
		return entry

	case "PHASE_TRANSITIONED":
		entry := &HandoffHistoryEntry{
			Timestamp: timestamp,
			EventType: "phase_transition",
			FromPhase: e.FromPhase,
			ToPhase:   e.ToPhase,
			Summary:   fmt.Sprintf("Phase transition: %s -> %s", e.FromPhase, e.ToPhase),
		}
		return entry

	default:
		return nil
	}
}

// HandoffHistoryEntry represents a single handoff history entry.
type HandoffHistoryEntry struct {
	Timestamp   string   `json:"timestamp"`
	EventType   string   `json:"event_type"`
	Agent       string   `json:"agent,omitempty"`
	ToAgent     string   `json:"to_agent,omitempty"`
	FromPhase   string   `json:"from_phase,omitempty"`
	ToPhase     string   `json:"to_phase,omitempty"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Throughline string   `json:"throughline,omitempty"`
	Artifacts   []string `json:"artifacts,omitempty"`
	DurationMs  int64    `json:"duration_ms,omitempty"`
}

// HandoffHistoryOutput represents the output of handoff history.
type HandoffHistoryOutput struct {
	SessionID string                `json:"session_id"`
	Entries   []HandoffHistoryEntry `json:"entries"`
	Total     int                   `json:"total"`
	Showing   int                   `json:"showing,omitempty"`
}

// Headers implements output.Tabular for HandoffHistoryOutput.
func (h HandoffHistoryOutput) Headers() []string {
	return []string{"TIMESTAMP", "EVENT", "AGENT", "SUMMARY"}
}

// Rows implements output.Tabular for HandoffHistoryOutput.
func (h HandoffHistoryOutput) Rows() [][]string {
	rows := make([][]string, len(h.Entries))
	for i, e := range h.Entries {
		agent := e.Agent
		if agent == "" {
			agent = e.ToAgent
		}
		if agent == "" && e.FromPhase != "" {
			agent = e.FromPhase + " -> " + e.ToPhase
		}

		// Truncate timestamp to just time portion
		ts := e.Timestamp
		if len(ts) > 19 {
			ts = ts[11:19] // Extract HH:MM:SS
		}

		// Truncate summary
		summary := e.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}

		rows[i] = []string{ts, e.EventType, agent, summary}
	}
	return rows
}

// Text implements output.Textable for HandoffHistoryOutput.
func (h HandoffHistoryOutput) Text() string {
	if len(h.Entries) == 0 {
		return fmt.Sprintf("No handoff history for session %s", h.SessionID)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Handoff History for %s\n", h.SessionID))
	b.WriteString(fmt.Sprintf("Total entries: %d", h.Total))
	if h.Showing > 0 && h.Showing < h.Total {
		b.WriteString(fmt.Sprintf(" (showing last %d)", h.Showing))
	}
	b.WriteString("\n\n")

	return b.String()
}

// Ensure Tabular and Textable interfaces are implemented
var _ output.Tabular = HandoffHistoryOutput{}
var _ output.Textable = HandoffHistoryOutput{}
