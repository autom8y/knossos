// Package session implements the ari session commands.
package session

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/lock"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/session"
)

// cmdContext holds shared state for session commands.
type cmdContext struct {
	output    *string
	verbose   *bool
	projectDir *string
	sessionID  *string
}

// NewSessionCmd creates the session command group.
func NewSessionCmd(outputFlag *string, verboseFlag *bool, projectDir, sessionID *string) *cobra.Command {
	ctx := &cmdContext{
		output:     outputFlag,
		verbose:    verboseFlag,
		projectDir: projectDir,
		sessionID:  sessionID,
	}

	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage workflow sessions",
		Long:  `Create, list, park, resume, and manage Claude Code workflow sessions.`,
	}

	// Add subcommands
	cmd.AddCommand(newCreateCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newListCmd(ctx))
	cmd.AddCommand(newParkCmd(ctx))
	cmd.AddCommand(newResumeCmd(ctx))
	cmd.AddCommand(newWrapCmd(ctx))
	cmd.AddCommand(newTransitionCmd(ctx))
	cmd.AddCommand(newMigrateCmd(ctx))
	cmd.AddCommand(newAuditCmd(ctx))
	cmd.AddCommand(newLockCmd(ctx))
	cmd.AddCommand(newUnlockCmd(ctx))

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

// getLockManager creates a lock manager from the context.
func (c *cmdContext) getLockManager() *lock.Manager {
	resolver := c.getResolver()
	return lock.NewManager(resolver.LocksDir())
}

// getEventEmitter creates an event emitter for a session.
func (c *cmdContext) getEventEmitter(sessionID string) *session.EventEmitter {
	resolver := c.getResolver()
	eventsPath := resolver.SessionEventsFile(sessionID)
	auditPath := resolver.TransitionsLog()
	return session.NewEventEmitter(eventsPath, auditPath)
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

// setCurrentSessionID writes the current session ID.
func (c *cmdContext) setCurrentSessionID(sessionID string) error {
	resolver := c.getResolver()
	if err := paths.EnsureDir(resolver.SessionsDir()); err != nil {
		return err
	}
	return os.WriteFile(resolver.CurrentSessionFile(), []byte(sessionID), 0644)
}

// clearCurrentSessionID removes the current session marker.
func (c *cmdContext) clearCurrentSessionID() error {
	resolver := c.getResolver()
	err := os.Remove(resolver.CurrentSessionFile())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// getActiveTeam reads the active team from ACTIVE_TEAM file.
func (c *cmdContext) getActiveTeam() string {
	resolver := c.getResolver()
	data, err := os.ReadFile(resolver.ActiveTeamFile())
	if err != nil {
		return "none"
	}
	return string(data)
}
