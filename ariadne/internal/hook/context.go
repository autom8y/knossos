package hook

import (
	"os"
	"path/filepath"

	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/session"
)

// Context holds session and project context for hook operations.
type Context struct {
	// Project paths
	ProjectDir string
	ClaudeDir  string

	// Session state
	SessionID  string
	Session    *session.Context
	HasSession bool

	// Team information
	ActiveTeam string

	// Execution mode
	ExecutionMode string

	// Path resolver for consistent path operations
	resolver *paths.Resolver
}

// LoadContext loads session and project context for hook operations.
// This function is optimized for performance - it returns early when no context is available.
func LoadContext(env *Env) (*Context, error) {
	ctx := &Context{}

	// Determine project directory
	projectDir := env.GetProjectDir()
	if projectDir == "" || projectDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return ctx, nil // Return empty context, don't fail
		}
		projectDir = cwd
	}

	// Try to find project root (walk up looking for .claude/)
	projectRoot, err := paths.FindProjectRoot(projectDir)
	if err != nil {
		// No project found - return minimal context
		ctx.ProjectDir = projectDir
		return ctx, nil
	}

	ctx.ProjectDir = projectRoot
	ctx.ClaudeDir = filepath.Join(projectRoot, ".claude")
	ctx.resolver = paths.NewResolver(projectRoot)

	// Load active team (quick file read)
	if data, err := os.ReadFile(ctx.resolver.ActiveTeamFile()); err == nil {
		ctx.ActiveTeam = string(data)
	}

	// Determine execution mode based on team
	ctx.ExecutionMode = determineExecutionMode(ctx.ActiveTeam, ctx.HasSession)

	// Load current session ID
	sessionID := env.SessionID
	if sessionID == "" {
		// Try to read from .current-session file
		if data, err := os.ReadFile(ctx.resolver.CurrentSessionFile()); err == nil {
			sessionID = string(data)
		}
	}

	if sessionID == "" {
		// No session - return context without session data
		return ctx, nil
	}

	ctx.SessionID = sessionID
	ctx.HasSession = true

	// Load session context
	sessionPath := ctx.resolver.SessionContextFile(sessionID)
	sessionCtx, err := session.LoadContext(sessionPath)
	if err != nil {
		// Session file exists but couldn't load - still return context
		return ctx, nil
	}

	ctx.Session = sessionCtx

	// Update execution mode based on session
	ctx.ExecutionMode = determineExecutionMode(ctx.ActiveTeam, true)

	return ctx, nil
}

// LoadContextFromProjectDir loads context from a specific project directory.
func LoadContextFromProjectDir(projectDir string) (*Context, error) {
	env := &Env{ProjectDir: projectDir}
	return LoadContext(env)
}

// determineExecutionMode determines the execution mode based on team and session state.
func determineExecutionMode(activeTeam string, hasSession bool) string {
	if !hasSession {
		return "native"
	}
	if activeTeam == "" || activeTeam == "none" {
		return "cross-cutting"
	}
	return "orchestrated"
}

// GetResolver returns the path resolver for this context.
func (c *Context) GetResolver() *paths.Resolver {
	if c.resolver == nil {
		c.resolver = paths.NewResolver(c.ProjectDir)
	}
	return c.resolver
}

// IsOrchestrated returns true if in orchestrated execution mode.
func (c *Context) IsOrchestrated() bool {
	return c.ExecutionMode == "orchestrated"
}

// IsCrossCutting returns true if in cross-cutting execution mode.
func (c *Context) IsCrossCutting() bool {
	return c.ExecutionMode == "cross-cutting"
}

// IsNative returns true if in native execution mode (no session).
func (c *Context) IsNative() bool {
	return c.ExecutionMode == "native"
}

// GetSessionStatus returns the session status or empty string if no session.
func (c *Context) GetSessionStatus() string {
	if c.Session == nil {
		return ""
	}
	return string(c.Session.Status)
}

// GetCurrentPhase returns the current workflow phase or empty string.
func (c *Context) GetCurrentPhase() string {
	if c.Session == nil {
		return ""
	}
	return c.Session.CurrentPhase
}

// ContextSummary returns a map of context data suitable for hook output.
func (c *Context) ContextSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"execution_mode": c.ExecutionMode,
		"has_session":    c.HasSession,
	}

	if c.ProjectDir != "" {
		summary["project_dir"] = c.ProjectDir
	}

	if c.ActiveTeam != "" {
		summary["active_team"] = c.ActiveTeam
	}

	if c.SessionID != "" {
		summary["session_id"] = c.SessionID
	}

	if c.Session != nil {
		summary["session_status"] = string(c.Session.Status)
		summary["current_phase"] = c.Session.CurrentPhase
		summary["initiative"] = c.Session.Initiative
	}

	return summary
}
