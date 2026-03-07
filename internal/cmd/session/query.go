package session

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

type queryOptions struct {
	field string
}

func newQueryCmd(ctx *cmdContext) *cobra.Command {
	var opts queryOptions

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query current session state (secondary channel for agents)",
		Long: `Query the active session state on demand.

Returns the same YAML frontmatter format as the hook context injection so
agents can parse query output identically to hook-injected context. Use
this when you need a fresh view of session state mid-conversation without
waiting for the next SessionStart hook.

This is a read-only command — it does not mutate session state or emit
lifecycle events.

Resolution chain (same as hook):
  1. --session-id flag (explicit)
  2. CC session map (.cc-map/)
  3. Smart scan (single active session)

Examples:
  ari session query
  ari session query -o json
  ari session query --field complexity
  ari session query --field status
  ari session query --session-id session-20260306-122256-4fc1e1cc

Context:
  Agents call this mid-conversation when session state may have changed.
  Prefer hook injection for initial load; use query for fresh pulls.
  Use --field for single-value reads (e.g., scripted phase checks).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.field, "field", "",
		"Return only the named field value (e.g. complexity, status, current_phase)")

	return cmd
}

func runQuery(ctx *cmdContext, opts queryOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Resolve session ID via the standard chain: explicit flag > CC map > smart scan.
	// We pass empty ccSessionID since this is a user-invoked CLI command, not a hook.
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		err = errors.Wrap(errors.CodeGeneralError, "failed to resolve session", err)
		return common.PrintAndReturn(printer, err)
	}

	if sessionID == "" {
		// No active session — mirror hook's no-session path
		if opts.field != "" {
			err := errors.New(errors.CodeSessionNotFound, "no active session")
			return common.PrintAndReturn(printer, err)
		}
		return printer.Print(output.QueryOutput{HasSession: false})
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.New(errors.CodeSessionNotFound,
				fmt.Sprintf("session not found: %s", sessionID))
		} else {
			err = errors.Wrap(errors.CodeGeneralError, "failed to load session context", err)
		}
		return common.PrintAndReturn(printer, err)
	}

	// Read active rite with backward compatibility (same as hook)
	activeRite := resolver.ReadActiveRite()
	if activeRite == "" {
		activeRite = sessCtx.ActiveRite
	}

	// Determine execution mode using the same logic as the hook
	mode := determineQueryExecutionMode(sessCtx, activeRite)

	// --field: single field lookup
	if opts.field != "" {
		value, ok := getQueryField(sessCtx, activeRite, mode, opts.field)
		if !ok {
			err := errors.New(errors.CodeUsageError,
				fmt.Sprintf("unknown field %q: valid fields are session_id, status, initiative, complexity, active_rite, execution_mode, current_phase, frayed_from, frame_ref, park_source, claimed_by", opts.field))
			return common.PrintAndReturn(printer, err)
		}
		printer.PrintLine(value)
		return nil
	}

	// Full output — build QueryOutput mirroring hook ContextOutput
	result := output.QueryOutput{
		SessionID:     sessCtx.SessionID,
		Status:        string(sessCtx.Status),
		Initiative:    sessCtx.Initiative,
		Complexity:    sessCtx.Complexity,
		ActiveRite:    activeRite,
		ExecutionMode: mode,
		CurrentPhase:  sessCtx.CurrentPhase,
		FrayedFrom:    sessCtx.FrayedFrom,
		FrameRef:      sessCtx.FrameRef,
		ParkSource:    sessCtx.ParkSource,
		ClaimedBy:     sessCtx.ClaimedBy,
		Strands:       convertQueryStrands(sessCtx.Strands),
		HasSession:    true,
	}

	return printer.Print(result)
}

// determineQueryExecutionMode mirrors hook.determineExecutionMode.
// It lives here to avoid a cross-package import of the hook package from the
// session command package (which would create an import cycle).
func determineQueryExecutionMode(sessCtx *sess.Context, activeRite string) string {
	if sessCtx == nil {
		return "native"
	}
	if activeRite != "" && activeRite != "none" {
		return "orchestrated"
	}
	return "cross-cutting"
}

// convertQueryStrands converts session.Strand slice to output.QueryStrand slice.
// Returns nil when input is nil or empty (omitempty suppresses the field).
func convertQueryStrands(strands []sess.Strand) []output.QueryStrand {
	if len(strands) == 0 {
		return nil
	}
	out := make([]output.QueryStrand, len(strands))
	for i, s := range strands {
		out[i] = output.QueryStrand{
			SessionID: s.SessionID,
			Status:    s.Status,
			FrameRef:  s.FrameRef,
			LandedAt:  s.LandedAt,
		}
	}
	return out
}

// getQueryField returns the string value for the named field from session state.
// Returns ("", false) for unknown field names.
func getQueryField(sessCtx *sess.Context, activeRite, mode, field string) (string, bool) {
	switch field {
	case "session_id":
		return sessCtx.SessionID, true
	case "status":
		return string(sessCtx.Status), true
	case "initiative":
		return sessCtx.Initiative, true
	case "complexity":
		return sessCtx.Complexity, true
	case "active_rite":
		return activeRite, true
	case "execution_mode":
		return mode, true
	case "current_phase":
		return sessCtx.CurrentPhase, true
	case "frayed_from":
		return sessCtx.FrayedFrom, true
	case "frame_ref":
		return sessCtx.FrameRef, true
	case "park_source":
		return sessCtx.ParkSource, true
	case "claimed_by":
		return sessCtx.ClaimedBy, true
	default:
		return "", false
	}
}
