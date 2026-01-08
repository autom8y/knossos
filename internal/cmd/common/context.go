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
	resolver := c.GetResolver()
	ritePath := resolver.ActiveRiteFile()
	data, err := os.ReadFile(ritePath)
	if err != nil {
		return ""
	}
	return string(data)
}

// GetSessionID returns the session ID from the flag or reads the current session.
func (c *SessionContext) GetSessionID() (string, error) {
	if c.SessionID != nil && *c.SessionID != "" {
		return *c.SessionID, nil
	}
	return c.GetCurrentSessionID()
}

// GetCurrentSessionID reads the current session ID from the project.
// Returns empty string with no error if no session is active (file doesn't exist).
func (c *SessionContext) GetCurrentSessionID() (string, error) {
	resolver := c.GetResolver()
	data, err := os.ReadFile(resolver.CurrentSessionFile())
	if err != nil {
		if os.IsNotExist(err) {
			// No current session is a valid state, not an error
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// SetCurrentSessionID writes the current session ID.
func (c *SessionContext) SetCurrentSessionID(sessionID string) error {
	resolver := c.GetResolver()
	if err := paths.EnsureDir(resolver.SessionsDir()); err != nil {
		return err
	}
	return os.WriteFile(resolver.CurrentSessionFile(), []byte(sessionID), 0644)
}

// ClearCurrentSessionID removes the current session ID file.
func (c *SessionContext) ClearCurrentSessionID() error {
	resolver := c.GetResolver()
	err := os.Remove(resolver.CurrentSessionFile())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// GetLockManager creates a lock manager for session operations.
func (c *SessionContext) GetLockManager() *lock.Manager {
	resolver := c.GetResolver()
	return lock.NewManager(resolver.LocksDir())
}

// GetEventEmitter creates an event emitter for the given session.
func (c *SessionContext) GetEventEmitter(sessionID string) *session.EventEmitter {
	resolver := c.GetResolver()
	eventsPath := resolver.SessionEventsFile(sessionID)
	auditPath := resolver.TransitionsLog()
	return session.NewEventEmitter(eventsPath, auditPath)
}
