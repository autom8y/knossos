// Package materialize provides unified sync types for both rite and user scopes.
package materialize

// SyncScope determines which scopes to execute during sync.
type SyncScope string

const (
	ScopeAll  SyncScope = "all"
	ScopeRite SyncScope = "rite"
	ScopeUser SyncScope = "user"
)

// IsValid returns true if the scope is a recognized value.
func (s SyncScope) IsValid() bool {
	switch s {
	case ScopeAll, ScopeRite, ScopeUser:
		return true
	default:
		return false
	}
}

// SyncResource identifies a filterable resource type.
type SyncResource string

const (
	ResourceAll    SyncResource = ""
	ResourceAgents SyncResource = "agents"
	ResourceMena   SyncResource = "mena"
	ResourceHooks  SyncResource = "hooks"
)

// IsValid returns true if the resource is a recognized value.
func (r SyncResource) IsValid() bool {
	switch r {
	case ResourceAll, ResourceAgents, ResourceMena, ResourceHooks:
		return true
	default:
		return false
	}
}

// SyncOptions configures the unified sync pipeline.
type SyncOptions struct {
	Scope             SyncScope
	RiteName          string
	Source            string
	Resource          SyncResource
	DryRun            bool
	Recover           bool
	OverwriteDiverged bool
	KeepOrphans       bool
}

// SyncResult contains unified outcome.
type SyncResult struct {
	RiteResult *RiteScopeResult `json:"rite,omitempty"`
	UserResult *UserScopeResult `json:"user,omitempty"`
}

// RiteScopeResult wraps rite scope outcome.
type RiteScopeResult struct {
	Status           string   `json:"status"`
	RiteName         string   `json:"rite_name,omitempty"`
	Source           string   `json:"source,omitempty"`
	SourcePath       string   `json:"source_path,omitempty"`
	OrphansDetected  []string `json:"orphans_detected,omitempty"`
	OrphanAction     string   `json:"orphan_action,omitempty"`
	BackupPath       string   `json:"backup_path,omitempty"`
	LegacyBackupPath string   `json:"legacy_backup_path,omitempty"`
}

// UserScopeResult wraps user scope outcome.
type UserScopeResult struct {
	Status    string                               `json:"status"`
	Resources map[SyncResource]*UserResourceResult `json:"resources,omitempty"`
	Totals    UserSyncSummary                      `json:"totals"`
	Errors    []UserResourceError                  `json:"errors,omitempty"`
}

// UserResourceResult is per-resource outcome.
type UserResourceResult struct {
	Source  string          `json:"source"`
	Target  string          `json:"target"`
	Changes UserSyncChanges `json:"changes"`
	Summary UserSyncSummary `json:"summary"`
}

// UserSyncChanges tracks files changed during user sync.
type UserSyncChanges struct {
	Added     []string           `json:"added"`
	Updated   []string           `json:"updated"`
	Skipped   []UserSkippedEntry `json:"skipped"`
	Unchanged []string           `json:"unchanged"`
}

// UserSkippedEntry explains why a file was skipped.
type UserSkippedEntry struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// UserSyncSummary provides aggregate counts.
type UserSyncSummary struct {
	Added      int `json:"added"`
	Updated    int `json:"updated"`
	Skipped    int `json:"skipped"`
	Unchanged  int `json:"unchanged"`
	Collisions int `json:"collisions"`
}

// UserResourceError captures errors for a specific resource.
type UserResourceError struct {
	Resource SyncResource `json:"resource"`
	Err      string       `json:"error"`
}
