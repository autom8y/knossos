// Package hook implements the ari hook commands.
// This package provides the hook command group for Claude Code hook integration.
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// Default timeout for hook operations (100ms target, 500ms max safety).
const (
	DefaultTimeout     = 100 * time.Millisecond
	MaxTimeout         = 500 * time.Millisecond
	EarlyExitThreshold = 5 * time.Millisecond
)

// cmdContext holds shared state for hook commands.
type cmdContext struct {
	common.SessionContext
	timeout time.Duration
}

// NewHookCmd creates the hook command group.
func NewHookCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     outputFlag,
				Verbose:    verboseFlag,
				ProjectDir: projectDir,
			},
			SessionID: sessionID,
		},
		timeout: DefaultTimeout,
	}

	var timeoutMs int

	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Claude Code hook infrastructure",
		Long: `Hook command group for Claude Code hook integration.

This command group provides infrastructure for Claude Code hooks,
enabling Go-based hook implementations with consistent behavior.

Hooks process Claude Code tool events and can modify, validate,
or enrich tool operations. Use subcommands for specific hook types.

Environment Variables:
  CLAUDE_HOOK_*      Standard Claude Code hook environment variables

Performance Targets:
  Early exit: <5ms   (when hooks disabled or no session)
  Full execution: <100ms (with all processing)`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Convert timeout flag to duration
			if timeoutMs > 0 {
				ctx.timeout = time.Duration(timeoutMs) * time.Millisecond
				if ctx.timeout > MaxTimeout {
					ctx.timeout = MaxTimeout
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand specified
			cmd.Help()
		},
	}

	// Add persistent flags for all hook subcommands
	cmd.PersistentFlags().IntVar(&timeoutMs, "timeout", 100,
		"Hook operation timeout in milliseconds (max 500)")

	// Add hook subcommands
	cmd.AddCommand(newContextCmd(ctx))
	cmd.AddCommand(newAutoparkCmd(ctx))
	cmd.AddCommand(newWriteguardCmd(ctx))
	cmd.AddCommand(newRouteCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newClewCmd(ctx))
	cmd.AddCommand(newBudgetCmd(ctx))
	cmd.AddCommand(newPrecompactCmd(ctx))
	cmd.AddCommand(newSubagentStartCmd(ctx))
	cmd.AddCommand(newSubagentStopCmd(ctx))

	// Hook commands do NOT require project context
	common.SetNeedsProject(cmd, false, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
// Hook commands default to JSON format for machine-readable output.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatJSON)
}

// getHookEnv parses the hook environment variables.
func (c *cmdContext) getHookEnv() *hook.Env {
	return hook.ParseEnv()
}

// withTimeout wraps a command execution function with context.WithTimeout.
// This ensures all hook commands respect the configured timeout.
func (c *cmdContext) withTimeout(fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Create a channel to signal completion
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("hook operation timed out after %v", c.timeout)
	}
}

// resolveSession resolves the project resolver and current session ID from hook context.
// Returns the resolver, trimmed session ID, and any error.
// If no resolver can be established, resolver.ProjectRoot() will be empty.
// If no session exists, sessionID will be empty and err will be nil.
func (c *cmdContext) resolveSession(hookEnv *hook.Env) (*paths.Resolver, string, error) {
	// Get resolver for path lookups
	resolver := c.GetResolver()
	if resolver.ProjectRoot() == "" {
		// Try to discover project from environment
		if hookEnv.ProjectDir != "" {
			resolver = paths.NewResolver(hookEnv.ProjectDir)
		}
	}

	// Use ResolveSession priority chain: explicit flag > CC map > smart scan
	explicitID := ""
	if c.SessionID != nil {
		explicitID = *c.SessionID
	}
	sessionID, err := session.ResolveSession(resolver, hookEnv.SessionID, explicitID)
	if err != nil {
		return resolver, "", err
	}

	return resolver, sessionID, nil
}
