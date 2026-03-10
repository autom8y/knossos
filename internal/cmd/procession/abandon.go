// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// newAbandonCmd creates the procession abandon subcommand.
func newAbandonCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abandon",
		Short: "Terminate the procession (session continues)",
		Long: `Terminate the active procession. The session continues and can be used
normally, but the cross-rite workflow coordination is removed. Artifact files
in the artifact directory are NOT deleted.

Use this when the cross-rite workflow is no longer needed or was started in error.

Examples:
  ari procession abandon`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAbandon(ctx)
		},
	}
	return cmd
}

// abandonOutput represents the output of procession abandon.
type abandonOutput struct {
	Message      string `json:"message"`
	ProcessionID string `json:"procession_id"`
	SessionID    string `json:"session_id"`
}

// Text implements output.Textable for abandonOutput.
func (o abandonOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Procession %s abandoned.\n", o.ProcessionID)
	fmt.Fprintf(&b, "Session %s continues.\n", o.SessionID)
	return b.String()
}

var _ output.Textable = abandonOutput{}

// runAbandon executes the procession abandon command.
func runAbandon(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get session ID
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}
	if sessionID == "" {
		return common.PrintAndReturn(printer, errors.New(errors.CodeSessionNotFound, "No active session. Use 'ari session create' first."))
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			return common.PrintAndReturn(printer, errors.ErrSessionNotFound(sessionID))
		}
		return common.PrintAndReturn(printer, err)
	}

	// Check procession is active
	if sessCtx.Procession == nil {
		return common.PrintAndReturn(printer, errors.New(errors.CodeUsageError,
			"No active procession. Nothing to abandon."))
	}

	// Capture the ID before clearing
	processionID := sessCtx.Procession.ID

	// Remove procession from session context; session continues
	sessCtx.Procession = (*session.Procession)(nil)

	// Save session context
	if err := sessCtx.Save(ctxPath); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to save session context", err))
	}

	return printer.Print(abandonOutput{
		Message:      "procession abandoned",
		ProcessionID: processionID,
		SessionID:    sessionID,
	})
}
