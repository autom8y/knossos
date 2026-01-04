// Package paths provides path resolution and project discovery for Ariadne.
// It handles XDG base directories and project root discovery.
package paths

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/autom8y/ariadne/internal/errors"
)

// Resolver handles path resolution relative to a project root.
type Resolver struct {
	projectRoot string
}

// NewResolver creates a new path resolver for the given project root.
func NewResolver(projectRoot string) *Resolver {
	return &Resolver{projectRoot: projectRoot}
}

// FindProjectRoot walks up from the given directory looking for .claude/.
// If startDir is empty, uses the current working directory.
// Returns an error if no .claude/ directory is found.
func FindProjectRoot(startDir string) (string, error) {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return "", errors.Wrap(errors.CodeGeneralError, "failed to get working directory", err)
		}
	}

	dir := startDir
	for {
		claudeDir := filepath.Join(dir, ".claude")
		info, err := os.Stat(claudeDir)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", errors.ErrProjectNotFound()
		}
		dir = parent
	}
}

// ProjectRoot returns the project root directory.
func (r *Resolver) ProjectRoot() string {
	return r.projectRoot
}

// ClaudeDir returns the path to the .claude directory.
func (r *Resolver) ClaudeDir() string {
	return filepath.Join(r.projectRoot, ".claude")
}

// SessionsDir returns the path to the sessions directory.
func (r *Resolver) SessionsDir() string {
	return filepath.Join(r.projectRoot, ".claude", "sessions")
}

// LocksDir returns the path to the locks directory.
func (r *Resolver) LocksDir() string {
	return filepath.Join(r.projectRoot, ".claude", "sessions", ".locks")
}

// AuditDir returns the path to the audit directory.
func (r *Resolver) AuditDir() string {
	return filepath.Join(r.projectRoot, ".claude", "sessions", ".audit")
}

// ArchiveDir returns the path to the archive directory.
func (r *Resolver) ArchiveDir() string {
	return filepath.Join(r.projectRoot, ".claude", ".archive", "sessions")
}

// SessionDir returns the path to a specific session directory.
func (r *Resolver) SessionDir(sessionID string) string {
	return filepath.Join(r.SessionsDir(), sessionID)
}

// SessionContextFile returns the path to a session's SESSION_CONTEXT.md.
func (r *Resolver) SessionContextFile(sessionID string) string {
	return filepath.Join(r.SessionDir(sessionID), "SESSION_CONTEXT.md")
}

// SessionEventsFile returns the path to a session's events.jsonl.
func (r *Resolver) SessionEventsFile(sessionID string) string {
	return filepath.Join(r.SessionDir(sessionID), "events.jsonl")
}

// LockFile returns the path to a session's lock file.
func (r *Resolver) LockFile(sessionID string) string {
	return filepath.Join(r.LocksDir(), sessionID+".lock")
}

// CurrentSessionFile returns the path to the .current-session file.
func (r *Resolver) CurrentSessionFile() string {
	return filepath.Join(r.SessionsDir(), ".current-session")
}

// ActiveTeamFile returns the path to the ACTIVE_TEAM file.
func (r *Resolver) ActiveTeamFile() string {
	return filepath.Join(r.ClaudeDir(), "ACTIVE_TEAM")
}

// ActiveWorkflowFile returns the path to the ACTIVE_WORKFLOW.yaml file.
func (r *Resolver) ActiveWorkflowFile() string {
	return filepath.Join(r.ClaudeDir(), "ACTIVE_WORKFLOW.yaml")
}

// AgentManifestFile returns the path to the AGENT_MANIFEST.json file.
func (r *Resolver) AgentManifestFile() string {
	return filepath.Join(r.ClaudeDir(), "AGENT_MANIFEST.json")
}

// AgentsDir returns the path to the .claude/agents/ directory.
func (r *Resolver) AgentsDir() string {
	return filepath.Join(r.ClaudeDir(), "agents")
}

// AgentFile returns the path to a specific agent file.
func (r *Resolver) AgentFile(name string) string {
	return filepath.Join(r.AgentsDir(), name)
}

// ClaudeMDFile returns the path to the .claude/CLAUDE.md file.
func (r *Resolver) ClaudeMDFile() string {
	return filepath.Join(r.ClaudeDir(), "CLAUDE.md")
}

// TeamsDir returns the path to the project's teams/ directory.
func (r *Resolver) TeamsDir() string {
	return filepath.Join(r.projectRoot, "teams")
}

// TeamDir returns the path to a specific team directory.
func (r *Resolver) TeamDir(teamName string) string {
	return filepath.Join(r.TeamsDir(), teamName)
}

// TeamAgentsDir returns the path to a team's agents/ directory.
func (r *Resolver) TeamAgentsDir(teamName string) string {
	return filepath.Join(r.TeamDir(teamName), "agents")
}

// TeamWorkflowFile returns the path to a team's workflow.yaml file.
func (r *Resolver) TeamWorkflowFile(teamName string) string {
	return filepath.Join(r.TeamDir(teamName), "workflow.yaml")
}

// TeamOrchestratorFile returns the path to a team's orchestrator.yaml file.
func (r *Resolver) TeamOrchestratorFile(teamName string) string {
	return filepath.Join(r.TeamDir(teamName), "orchestrator.yaml")
}

// TeamContextFile returns the path to a team's context.yaml file.
func (r *Resolver) TeamContextFile(teamName string) string {
	return filepath.Join(r.TeamDir(teamName), "context.yaml")
}

// TransitionsLog returns the path to the global transitions log.
func (r *Resolver) TransitionsLog() string {
	return filepath.Join(r.AuditDir(), "transitions.log")
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// --- XDG Directory Helpers ---

// ConfigDir returns the XDG config directory for ariadne.
func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "ariadne")
}

// StateDir returns the XDG state directory for ariadne.
func StateDir() string {
	return filepath.Join(xdg.StateHome, "ariadne")
}

// CacheDir returns the XDG cache directory for ariadne.
func CacheDir() string {
	return filepath.Join(xdg.CacheHome, "ariadne")
}

// DataDir returns the XDG data directory for ariadne.
func DataDir() string {
	return filepath.Join(xdg.DataHome, "ariadne")
}

// UserTeamsDir returns the user-level teams directory.
func UserTeamsDir() string {
	return filepath.Join(DataDir(), "teams")
}

// ConfigFile returns the path to a file in the config directory.
func ConfigFile(name string) string {
	return filepath.Join(ConfigDir(), name)
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	return EnsureDir(ConfigDir())
}

// EnsureStateDir creates the state directory if it doesn't exist.
func EnsureStateDir() error {
	return EnsureDir(StateDir())
}

// --- Session ID Helpers ---

// SessionIDFromDir extracts the session ID from a directory name.
func SessionIDFromDir(dir string) string {
	return filepath.Base(dir)
}

// IsSessionDir checks if a directory name looks like a session directory.
func IsSessionDir(name string) bool {
	// Session IDs follow pattern: session-YYYYMMDD-HHMMSS-{8-hex}
	if len(name) < 32 {
		return false
	}
	return len(name) >= 7 && name[:8] == "session-"
}
