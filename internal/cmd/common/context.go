// Package common provides shared context types and helper methods for CLI commands.
package common

import (
	"os"

	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// BaseContext contains the core fields shared by all command contexts.
type BaseContext struct {
	Output     *string
	Verbose    *bool
	ProjectDir *string
}

// SessionContext extends BaseContext with session-related fields.
type SessionContext struct {
	BaseContext
	SessionID *string
}

// GetPrinter creates an output.Printer with the specified default format.
// If the Output flag is set, it overrides the default.
func (c *BaseContext) GetPrinter(defaultFormat output.Format) *output.Printer {
	format := defaultFormat
	if c.Output != nil {
		format = output.ParseFormat(*c.Output)
	}
	verbose := c.Verbose != nil && *c.Verbose
	return output.NewPrinter(format, os.Stdout, os.Stderr, verbose)
}

// GetResolver creates a paths.Resolver for the project.
func (c *BaseContext) GetResolver() *paths.Resolver {
	projectDir := ""
	if c.ProjectDir != nil {
		projectDir = *c.ProjectDir
	}
	return paths.NewResolver(projectDir)
}

// GetActiveRite reads the active rite from the project.
func (c *BaseContext) GetActiveRite() string {
	return c.GetResolver().ReadActiveRite()
}

// GetSessionID returns the session ID from the flag or reads the current session.
func (c *SessionContext) GetSessionID() (string, error) {
	if c.SessionID != nil && *c.SessionID != "" {
		return *c.SessionID, nil
	}
	return session.FindActiveSession(c.GetResolver().SessionsDir())
}

// GetLockManager creates a lock manager for session operations.
func (c *SessionContext) GetLockManager() *lock.Manager {
	resolver := c.GetResolver()
	return lock.NewManager(resolver.LocksDir())
}

