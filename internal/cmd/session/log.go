package session

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

// logOptions holds the flag values for `ari session log`.
type logOptions struct {
	eventType string
	agent     string
	sha       string
	rationale string
}

func newLogCmd(ctx *cmdContext) *cobra.Command {
	var opts logOptions

	cmd := &cobra.Command{
		Use:   "log <message>",
		Short: "Append a typed event to the session timeline",
		Long: `Appends a typed event to events.jsonl AND a formatted entry to the
## Timeline section of SESSION_CONTEXT.md.

This command does NOT acquire a Moirai lock. Timeline appends are
intentionally lock-free because hooks may call this command at high frequency.
If two concurrent appends race, both events are preserved in events.jsonl and
the timeline entry can be reconstructed via 'ari session timeline --from-events'.

Context:
  Use this command to record significant events in the session timeline.
  Agents should log decisions with --type=decision and include --rationale.
  Hook handlers should log agent delegations and commit events automatically.

Examples:
  ari session log "started architect handoff"
  ari session log --type=decision "chose CSS vars over styled-components" --rationale="runtime perf"
  ari session log --type=agent --agent=architect "designing components"
  ari session log --type=commit --sha=abc123f "feat: theme provider"
  ari session log --type=command "invoked /consult"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLog(ctx, args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.eventType, "type", "t", "general",
		"Event type: general, decision, agent, commit, command (default: general)")
	cmd.Flags().StringVar(&opts.agent, "agent", "",
		"Agent name (required when --type=agent)")
	cmd.Flags().StringVar(&opts.sha, "sha", "",
		"Commit SHA (required when --type=commit)")
	cmd.Flags().StringVar(&opts.rationale, "rationale", "",
		"Decision rationale (used when --type=decision)")

	return cmd
}

func runLog(ctx *cmdContext, message string, opts logOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Validate flags before resolving session ID.
	if err := validateLogFlags(opts); err != nil {
		printer.PrintError(err)
		return err
	}

	// Resolve session ID. No lock is acquired -- timeline is append-only.
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

	// Build the typed event based on --type flag.
	event := buildLogEvent(message, opts)

	// Write event to events.jsonl.
	sessionDir := resolver.SessionDir(sessionID)
	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	writer.WriteTyped(event)
	if flushErr := writer.Flush(); flushErr != nil {
		// Non-fatal: log and continue. Event will be flushed by background goroutine.
		printer.VerboseLog("warn", "failed to flush typed event", map[string]any{
			"error": flushErr.Error(),
		})
	}
	writer.Close()

	// Format the timeline entry and append to SESSION_CONTEXT.md.
	entry := sess.FormatTimelineEntry(event)
	ctxPath := resolver.SessionContextFile(sessionID)
	if appendErr := sess.AppendEntry(ctxPath, entry); appendErr != nil {
		// Non-fatal: the event is already in events.jsonl and can be reconstructed.
		printer.VerboseLog("warn", "failed to append timeline entry", map[string]any{
			"error": appendErr.Error(),
			"entry": entry,
		})
	}

	// Output the result.
	return printer.PrintSuccess(output.LogOutput{
		SessionID: sessionID,
		Type:      string(event.Type),
		Entry:     entry,
	})
}

// validateLogFlags validates that required flags for each --type value are present.
func validateLogFlags(opts logOptions) error {
	switch opts.eventType {
	case "agent":
		if opts.agent == "" {
			return errors.New(errors.CodeUsageError,
				"--agent is required when --type=agent")
		}
	case "commit":
		if opts.sha == "" {
			return errors.New(errors.CodeUsageError,
				"--sha is required when --type=commit")
		}
	case "general", "decision", "command":
		// No additional required flags.
	default:
		return errors.New(errors.CodeUsageError,
			"--type must be one of: general, decision, agent, commit, command")
	}
	return nil
}

// buildLogEvent constructs a TypedEvent from the message and flag options.
// Source is always SourceAgent since `ari session log` is called by agents via CLI.
func buildLogEvent(message string, opts logOptions) clewcontract.TypedEvent {
	switch opts.eventType {
	case "decision":
		return clewcontract.NewTypedDecisionRecordedEvent(message, opts.rationale, nil)

	case "agent":
		// Pass message as the task ID so it appears in the timeline summary.
		// agentType and agentID are not supplied from CLI; they're empty.
		return clewcontract.NewTypedAgentDelegatedEvent(
			clewcontract.SourceAgent, opts.agent, "", message, "")

	case "commit":
		return clewcontract.NewTypedCommitCreatedEvent(opts.sha, message)

	case "command":
		// Override source to SourceAgent since this is an explicit CLI call, not a hook.
		event := clewcontract.NewTypedCommandInvokedEvent(message, "manual")
		event.Source = clewcontract.SourceAgent
		return event

	default: // "general"
		// General notes become decision.recorded with empty rationale.
		// This produces a NOTE category in the timeline via the fallback path.
		// Per spec: general → note event with source=agent.
		// We emit as decision.recorded (which is curated) with no rationale
		// so it appears on the timeline as a NOTE-like DECISION entry.
		return clewcontract.NewTypedDecisionRecordedEvent(message, "", nil)
	}
}
