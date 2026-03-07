package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

func newClaimCmd(ctx *cmdContext) *cobra.Command {
	var ccSessionID string

	cmd := &cobra.Command{
		Use:   "claim <session-id>",
		Short: "Bind this CC instance to a knossos session",
		Long: `Bind the current Claude Code instance to a specific knossos session.

Writes a CC-map entry so that subsequent SessionStart hooks resolve
to the specified session. The target session must exist and not be
in ARCHIVED state.

The CC session ID is obtained from the SessionStart hook output
(cc_session_id field). Pass it via --cc-session-id.

Examples:
  ari session claim session-20260305-140000-abc12345 --cc-session-id <cc-id>

Context:
  Use after 'ari session fray' to bind a new CC instance to the child session.
  Use to disambiguate when multiple active sessions exist.
  Idempotent: re-claiming overwrites any previous binding.`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClaim(ctx, args[0], ccSessionID)
		},
	}

	cmd.Flags().StringVar(&ccSessionID, "cc-session-id", "",
		"CC session ID (from SessionStart hook cc_session_id field)")
	_ = cmd.MarkFlagRequired("cc-session-id")

	return cmd
}

func runClaim(ctx *cmdContext, targetSessionID, ccSessionID string) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Validate target session exists
	ctxPath := resolver.SessionContextFile(targetSessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.ErrSessionNotFound(targetSessionID)
		}
		return err
	}

	// Reject ARCHIVED (terminal) sessions
	if sessCtx.Status == sess.StatusArchived {
		return errors.New(errors.CodeValidationFailed,
			fmt.Sprintf("cannot claim archived session %s (terminal state)", targetSessionID))
	}

	// Write CC map entry
	if err := sess.SetCCMap(resolver, ccSessionID, targetSessionID); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write CC map", err)
	}

	return printer.Print(output.ClaimOutput{
		SessionID:   targetSessionID,
		CCSessionID: ccSessionID,
		Status:      string(sessCtx.Status),
	})
}
