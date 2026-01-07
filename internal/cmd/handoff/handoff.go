// Package handoff implements the ari handoff commands for agent handoff management.
package handoff

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
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

Commands:
  prepare  - Validate readiness and emit task_end event
  execute  - Trigger transition and emit task_start event
  status   - Query current handoff state
  history  - Query handoff events from events.jsonl

Examples:
  ari handoff prepare --from=architect --to=principal-engineer
  ari handoff execute --artifact=TDD-user-auth --to=principal-engineer
  ari handoff status
  ari handoff history --limit=10`,
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
	return strings.TrimSpace(string(data)), nil
}
