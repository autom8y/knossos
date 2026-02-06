// Package common provides shared context types and helper methods for CLI commands.
package common

import (
	"os"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// CacheTTL is how long the .current-session cache is trusted before re-scanning.
const CacheTTL = 5 * time.Second

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

// GetCurrentSessionID returns the active session ID.
// Strategy: read .current-session cache, validate against scan.
// If cache is fresh (< CacheTTL) and non-empty, trust it.
// If cache is stale, missing, or empty, scan and rebuild.
func (c *SessionContext) GetCurrentSessionID() (string, error) {
	resolver := c.GetResolver()
	cacheFile := resolver.CurrentSessionFile()

	// Try cache first
	if info, err := os.Stat(cacheFile); err == nil {
		if time.Since(info.ModTime()) < CacheTTL {
			data, err := os.ReadFile(cacheFile)
			if err == nil && len(strings.TrimSpace(string(data))) > 0 {
				return strings.TrimSpace(string(data)), nil
			}
		}
	}

	// Cache miss/stale — scan for active session
	activeID, err := session.FindActiveSession(resolver.SessionsDir())
	if err != nil {
		return "", err
	}

	// Rebuild cache
	if activeID != "" {
		os.WriteFile(cacheFile, []byte(activeID), 0644)
	} else {
		os.Remove(cacheFile) // No active session — remove stale cache
	}

	return activeID, nil
}

// SetCurrentSessionID updates the .current-session cache.
// This is a performance optimization, not the source of truth.
func (c *SessionContext) SetCurrentSessionID(sessionID string) error {
	resolver := c.GetResolver()
	if err := paths.EnsureDir(resolver.SessionsDir()); err != nil {
		return err
	}
	return os.WriteFile(resolver.CurrentSessionFile(), []byte(sessionID), 0644)
}

// ClearCurrentSessionID removes the .current-session cache.
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
