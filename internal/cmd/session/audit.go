package session

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

type auditOptions struct {
	limit     int
	eventType string
	since     string
}

func newAuditCmd(ctx *cmdContext) *cobra.Command {
	var opts auditOptions

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Show session event history",
		Long: `Display session event history from events.jsonl.

Events include session lifecycle transitions, handoffs, phase changes,
and other significant actions recorded during the session. Reads all
event types, not just curated timeline events.

Examples:
  ari session audit
  ari session audit -n 10
  ari session audit -e HANDOFF_EXECUTED
  ari session audit --since 2026-01-05T14:00:00Z
  ari session audit -o json

Context:
  Low-level event inspection. Prefer 'ari session timeline' for curated view.
  Use -e to filter by event type for debugging specific lifecycle issues.
  Use --since with ISO8601 timestamps for time-bounded queries.
  Acquires shared lock -- safe for concurrent reads.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(ctx, opts)
		},
	}

	cmd.Flags().IntVarP(&opts.limit, "limit", "n", 50, "Maximum events to return")
	cmd.Flags().StringVarP(&opts.eventType, "event-type", "e", "", "Filter by event type")
	cmd.Flags().StringVar(&opts.since, "since", "", "Only events after this ISO8601 timestamp")

	return cmd
}

func runAudit(ctx *cmdContext, opts auditOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	// Acquire shared lock for consistent read
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Shared, lock.DefaultTimeout, "ari-session-audit")
	if err != nil {
		// Non-fatal - continue without lock
		printer.VerboseLog("warn", "failed to acquire lock", map[string]any{"error": err.Error()})
	} else {
		defer sessionLock.Release()
	}

	// Read events
	eventsPath := resolver.SessionEventsFile(sessionID)
	events, err := sess.ReadEvents(eventsPath)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Parse since timestamp
	var sinceTime time.Time
	if opts.since != "" {
		t, err := time.Parse(time.RFC3339, opts.since)
		if err != nil {
			printer.PrintError(errors.Wrap(errors.CodeUsageError, "invalid --since timestamp", err))
			return err
		}
		sinceTime = t
	}

	// Filter events
	filtered := sess.FilterEvents(events, opts.eventType, sinceTime)

	// Apply limit
	if opts.limit > 0 && len(filtered) > opts.limit {
		filtered = filtered[:opts.limit]
	}

	// Convert to output format
	outputEvents := make([]output.AuditEvent, len(filtered))
	for i, e := range filtered {
		outputEvents[i] = output.AuditEvent{
			Timestamp: e.Timestamp,
			Event:     string(e.Event),
			From:      e.From,
			To:        e.To,
			FromPhase: e.FromPhase,
			ToPhase:   e.ToPhase,
			Metadata:  e.Metadata,
		}
	}

	result := output.AuditOutput{
		SessionID: sessionID,
		Events:    outputEvents,
		Total:     len(outputEvents),
		FiltersApplied: output.AuditFilters{
			Limit:     opts.limit,
			EventType: opts.eventType,
			Since:     opts.since,
		},
	}

	return printer.Print(result)
}
