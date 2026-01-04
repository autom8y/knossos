// Package handoff implements the ari handoff commands for agent handoff management.
package handoff

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
)

// cmdContext holds shared state for handoff commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
	sessionID  *string
}

// NewHandoffCmd creates the handoff command group.
func NewHandoffCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
		sessionID:  sessionID,
	}

	cmd := &cobra.Command{
		Use:   "handoff",
		Short: "Manage agent handoffs between workflow phases",
		Long: `Manage handoffs between agents during workflow execution.

Handoffs transfer work from one agent to another within a session,
ensuring proper artifact validation and context preservation.

Examples:
  ari handoff prepare --from=architect --to=principal-engineer
  ari handoff execute --artifact=TDD-user-auth
  ari handoff status`,
	}

	// Add subcommands
	cmd.AddCommand(newPrepareCmd(ctx))
	cmd.AddCommand(newExecuteCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newHistoryCmd(ctx))

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	format := output.FormatText
	if c.output != nil {
		format = output.ParseFormat(*c.output)
	}
	verbose := false
	if c.verbose != nil {
		verbose = *c.verbose
	}
	return output.NewPrinter(format, os.Stdout, os.Stderr, verbose)
}

// getResolver creates a path resolver from the context.
func (c *cmdContext) getResolver() *paths.Resolver {
	projectDir := ""
	if c.projectDir != nil {
		projectDir = *c.projectDir
	}
	return paths.NewResolver(projectDir)
}

// getSessionID returns the session ID to use (from flag or current).
func (c *cmdContext) getSessionID() (string, error) {
	if c.sessionID != nil && *c.sessionID != "" {
		return *c.sessionID, nil
	}
	return c.getCurrentSessionID()
}

// getCurrentSessionID reads the current session ID from .current-session file.
func (c *cmdContext) getCurrentSessionID() (string, error) {
	resolver := c.getResolver()
	data, err := os.ReadFile(resolver.CurrentSessionFile())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// newPrepareCmd creates the handoff prepare subcommand.
func newPrepareCmd(ctx *cmdContext) *cobra.Command {
	var (
		fromAgent string
		toAgent   string
	)

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepare for handoff between agents",
		Long: `Prepare for a handoff by validating current agent's output
and checking readiness for the receiving agent.

This command validates handoff criteria and generates a handoff
context that can be passed to the receiving agent.

Examples:
  ari handoff prepare --from=architect --to=principal-engineer
  ari handoff prepare --from=requirements-analyst --to=architect`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			printer.PrintError(errors.New(errors.CodeGeneralError, "handoff prepare not yet implemented"))
			return fmt.Errorf("not implemented")
		},
	}

	cmd.Flags().StringVar(&fromAgent, "from", "", "Source agent (e.g., architect)")
	cmd.Flags().StringVar(&toAgent, "to", "", "Target agent (e.g., principal-engineer)")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

// newExecuteCmd creates the handoff execute subcommand.
func newExecuteCmd(ctx *cmdContext) *cobra.Command {
	var artifactID string

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a prepared handoff",
		Long: `Execute a handoff that has been prepared, recording the
transition in the session audit log.

Examples:
  ari handoff execute --artifact=TDD-user-auth
  ari handoff execute --artifact=PRD-user-auth`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			printer.PrintError(errors.New(errors.CodeGeneralError, "handoff execute not yet implemented"))
			return fmt.Errorf("not implemented")
		},
	}

	cmd.Flags().StringVar(&artifactID, "artifact", "", "Artifact ID being handed off")
	_ = cmd.MarkFlagRequired("artifact")

	return cmd
}

// newStatusCmd creates the handoff status subcommand.
func newStatusCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current handoff status",
		Long: `Show the current handoff status for the active session,
including pending handoffs and validation state.

Examples:
  ari handoff status
  ari handoff status --session=session-20260104-120000-abc12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			printer.PrintError(errors.New(errors.CodeGeneralError, "handoff status not yet implemented"))
			return fmt.Errorf("not implemented")
		},
	}

	return cmd
}

// newHistoryCmd creates the handoff history subcommand.
func newHistoryCmd(ctx *cmdContext) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show handoff history for a session",
		Long: `Show the history of handoffs that have occurred in the
current or specified session.

Examples:
  ari handoff history
  ari handoff history --limit=10
  ari handoff history --session=session-20260104-120000-abc12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			printer.PrintError(errors.New(errors.CodeGeneralError, "handoff history not yet implemented"))
			return fmt.Errorf("not implemented")
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of history entries (0 = unlimited)")

	return cmd
}
