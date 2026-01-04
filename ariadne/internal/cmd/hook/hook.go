// Package hook implements the ari hook commands.
// This package provides the hook command group for Claude Code hook integration.
package hook

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/hook"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
)

// Default timeout for hook operations (100ms target, 500ms max safety).
const (
	DefaultTimeout     = 100 * time.Millisecond
	MaxTimeout         = 500 * time.Millisecond
	EarlyExitThreshold = 5 * time.Millisecond
)

// cmdContext holds shared state for hook commands.
type cmdContext struct {
	output     *string
	verbose    *bool
	projectDir *string
	sessionID  *string
	timeout    time.Duration
}

// NewHookCmd creates the hook command group.
func NewHookCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
		sessionID:  sessionID,
		timeout:    DefaultTimeout,
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
  USE_ARI_HOOKS=1    Enable ari hook implementations
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
	cmd.AddCommand(newThreadCmd(ctx))

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	format := output.FormatJSON // Default to JSON for hook output
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

// getHookEnv parses the hook environment variables.
func (c *cmdContext) getHookEnv() *hook.Env {
	return hook.ParseEnv()
}

// shouldEarlyExit determines if hook should exit early.
// Returns true if hooks are disabled or no session context needed.
func (c *cmdContext) shouldEarlyExit() bool {
	return !hook.IsEnabled()
}

// getCurrentSessionID reads the current session ID from .current-session file.
func (c *cmdContext) getCurrentSessionID() (string, error) {
	if c.sessionID != nil && *c.sessionID != "" {
		return *c.sessionID, nil
	}
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
