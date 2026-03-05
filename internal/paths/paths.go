// Package paths provides path resolution and project discovery for Ariadne.
// It handles XDG base directories and project root discovery.
package paths

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/autom8y/knossos/internal/errors"
)

// Resolver handles path resolution relative to a project root.
type Resolver struct {
	projectRoot string
}

// NewResolver creates a new path resolver for the given project root.
func NewResolver(projectRoot string) *Resolver {
	return &Resolver{projectRoot: projectRoot}
}

// FindProjectRoot walks up from the given directory looking for .claude/ or .knossos/.
// If startDir is empty, uses the current working directory.
// Returns an error if neither directory is found.
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
		// Check .claude/ first (CC projects)
		claudeDir := filepath.Join(dir, ".claude")
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			return dir, nil
		}
		// Fallback: .knossos/ (knossos-managed projects)
		knossosDir := filepath.Join(dir, ".knossos")
		if info, err := os.Stat(knossosDir); err == nil && info.IsDir() {
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

// SOSDir returns the path to the .sos/ directory (Session Or State).
func (r *Resolver) SOSDir() string {
	return filepath.Join(r.projectRoot, ".sos")
}

// SessionsDir returns the path to the sessions directory.
func (r *Resolver) SessionsDir() string {
	return filepath.Join(r.SOSDir(), "sessions")
}

// LocksDir returns the path to the locks directory.
func (r *Resolver) LocksDir() string {
	return filepath.Join(r.SessionsDir(), ".locks")
}

// CCMapDir returns the path to the CC session map directory.
func (r *Resolver) CCMapDir() string {
	return filepath.Join(r.SessionsDir(), ".cc-map")
}

// WipDir returns the path to the .sos/wip/ directory (ephemeral working artifacts).
func (r *Resolver) WipDir() string {
	return filepath.Join(r.SOSDir(), "wip")
}

// ArchiveDir returns the path to the archive directory.
func (r *Resolver) ArchiveDir() string {
	return filepath.Join(r.SOSDir(), "archive")
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

// ActiveRiteFile returns the path to the ACTIVE_RITE file.
func (r *Resolver) ActiveRiteFile() string {
	return filepath.Join(r.KnossosDir(), "ACTIVE_RITE")
}

// ReadActiveRite reads the ACTIVE_RITE file and returns its content.
// Trims whitespace and returns empty string on error.
func (r *Resolver) ReadActiveRite() string {
	data, err := os.ReadFile(r.ActiveRiteFile())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ActiveWorkflowFile returns the path to the ACTIVE_WORKFLOW.yaml file.
func (r *Resolver) ActiveWorkflowFile() string {
	return filepath.Join(r.KnossosDir(), "ACTIVE_WORKFLOW.yaml")
}

// KnossosManifestFile returns the path to the KNOSSOS_MANIFEST.yaml file.
func (r *Resolver) KnossosManifestFile() string {
	return filepath.Join(r.KnossosDir(), "KNOSSOS_MANIFEST.yaml")
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

// KnossosDir returns the path to the .knossos/ directory (framework configuration).
func (r *Resolver) KnossosDir() string {
	return filepath.Join(r.projectRoot, ".knossos")
}

// RitesDir returns the path to the project's satellite rites directory.
func (r *Resolver) RitesDir() string {
	return filepath.Join(r.KnossosDir(), "rites")
}

// LedgeDir returns the path to the .ledge/ directory (work product artifacts).
func (r *Resolver) LedgeDir() string {
	return filepath.Join(r.projectRoot, ".ledge")
}

// LedgeDecisionsDir returns the path to the .ledge/decisions/ directory.
func (r *Resolver) LedgeDecisionsDir() string {
	return filepath.Join(r.LedgeDir(), "decisions")
}

// LedgeSpecsDir returns the path to the .ledge/specs/ directory.
func (r *Resolver) LedgeSpecsDir() string {
	return filepath.Join(r.LedgeDir(), "specs")
}

// LedgeReviewsDir returns the path to the .ledge/reviews/ directory.
func (r *Resolver) LedgeReviewsDir() string {
	return filepath.Join(r.LedgeDir(), "reviews")
}

// LedgeSpikesDir returns the path to the .ledge/spikes/ directory.
func (r *Resolver) LedgeSpikesDir() string {
	return filepath.Join(r.LedgeDir(), "spikes")
}

// LedgeShelfDir returns the path to the .ledge/shelf/ directory.
func (r *Resolver) LedgeShelfDir() string {
	return filepath.Join(r.LedgeDir(), "shelf")
}


// --- Rite Path Methods ---

// InvocationStateFile returns the path to the INVOCATION_STATE.yaml file.
func (r *Resolver) InvocationStateFile() string {
	return filepath.Join(r.KnossosDir(), "INVOCATION_STATE.yaml")
}

// KnossosSyncDir returns the path to the .knossos/sync/ directory.
func (r *Resolver) KnossosSyncDir() string {
	return filepath.Join(r.KnossosDir(), "sync")
}

// KnossosBackupsDir returns the path to the .knossos/backups/ directory.
func (r *Resolver) KnossosBackupsDir() string {
	return filepath.Join(r.KnossosDir(), "backups")
}

// ElCheapoMarkerFile returns the path to the .knossos/.el-cheapo-active marker.
func (r *Resolver) ElCheapoMarkerFile() string {
	return filepath.Join(r.KnossosDir(), ".el-cheapo-active")
}

// WorktreeMetaFile returns the path to per-worktree metadata in .knossos/.
func (r *Resolver) WorktreeMetaFile() string {
	return filepath.Join(r.KnossosDir(), ".worktree-meta.json")
}

// WorktreesDir returns the path to the .knossos/worktrees/ directory.
func (r *Resolver) WorktreesDir() string {
	return filepath.Join(r.KnossosDir(), "worktrees")
}

// RiteDir returns the path to a rite directory.
// Checks project satellite rites (.knossos/rites/) first, then user rites.
func (r *Resolver) RiteDir(riteName string) string {
	// Check project satellite rites first
	projectPath := filepath.Join(r.RitesDir(), riteName)
	if _, err := os.Stat(filepath.Join(projectPath, "manifest.yaml")); err == nil {
		return projectPath
	}
	// Fall back to user rites
	return filepath.Join(UserRitesDir(), riteName)
}

// RiteManifestFile returns the path to a rite's manifest file.
func (r *Resolver) RiteManifestFile(riteName string) string {
	return filepath.Join(r.RiteDir(riteName), "manifest.yaml")
}

// RiteAgentsDir returns the path to a rite's agents directory.
func (r *Resolver) RiteAgentsDir(riteName string) string {
	return filepath.Join(r.RiteDir(riteName), "agents")
}

// RiteSkillsDir returns the path to a rite's skills directory.
func (r *Resolver) RiteSkillsDir(riteName string) string {
	return filepath.Join(r.RiteDir(riteName), "skills")
}

// RiteWorkflowFile returns the path to a rite's workflow.yaml file.
func (r *Resolver) RiteWorkflowFile(riteName string) string {
	return filepath.Join(r.RiteDir(riteName), "workflow.yaml")
}

// RiteOrchestratorFile returns the path to a rite's orchestrator.yaml file.
func (r *Resolver) RiteOrchestratorFile(riteName string) string {
	return filepath.Join(r.RiteDir(riteName), "orchestrator.yaml")
}

// RiteContextFile returns the path to a rite's context.yaml file.
func (r *Resolver) RiteContextFile(riteName string) string {
	return filepath.Join(r.RiteDir(riteName), "context.yaml")
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// --- XDG Directory Helpers ---

// ConfigDir returns the XDG config directory for knossos.
func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "knossos")
}

// StateDir returns the XDG state directory for knossos.
func StateDir() string {
	return filepath.Join(xdg.StateHome, "knossos")
}

// CacheDir returns the XDG cache directory for knossos.
func CacheDir() string {
	return filepath.Join(xdg.CacheHome, "knossos")
}

// DataDir returns the XDG data directory for knossos.
func DataDir() string {
	return filepath.Join(xdg.DataHome, "knossos")
}

// UserRitesDir returns the user-level rites directory.
func UserRitesDir() string {
	return filepath.Join(DataDir(), "rites")
}

// --- Org-Level Resource Paths ---

// OrgDataDir returns the data directory for a named organization.
// Location: $XDG_DATA_HOME/knossos/orgs/{orgName}/
func OrgDataDir(orgName string) string {
	return filepath.Join(DataDir(), "orgs", orgName)
}

// OrgRitesDir returns the org-level rites directory.
// Returns empty string if orgName is empty.
func OrgRitesDir(orgName string) string {
	if orgName == "" {
		return ""
	}
	return filepath.Join(OrgDataDir(orgName), "rites")
}

// OrgAgentsDir returns the org-level agents directory.
func OrgAgentsDir(orgName string) string {
	return filepath.Join(OrgDataDir(orgName), "agents")
}

// OrgMenaDir returns the org-level mena directory.
func OrgMenaDir(orgName string) string {
	return filepath.Join(OrgDataDir(orgName), "mena")
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

// --- User-Level Resource Paths ---

// UserClaudeDir returns the user-level .claude directory.
func UserClaudeDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude")
}

// UserAgentsDir returns the user-level agents directory.
func UserAgentsDir() string {
	return filepath.Join(UserClaudeDir(), "agents")
}

// UserSkillsDir returns the user-level skills directory.
func UserSkillsDir() string {
	return filepath.Join(UserClaudeDir(), "skills")
}

// UserCommandsDir returns the user-level commands directory.
func UserCommandsDir() string {
	return filepath.Join(UserClaudeDir(), "commands")
}

// UserHooksDir returns the user-level hooks directory.
func UserHooksDir() string {
	return filepath.Join(UserClaudeDir(), "hooks")
}

// UserProvenanceManifest returns the path to the user-level provenance manifest.
func UserProvenanceManifest() string {
	return filepath.Join(UserClaudeDir(), "USER_PROVENANCE_MANIFEST.yaml")
}

// OrgProvenanceManifest returns the path to the org-level provenance manifest.
func OrgProvenanceManifest() string {
	return filepath.Join(UserClaudeDir(), "ORG_PROVENANCE_MANIFEST.yaml")
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
