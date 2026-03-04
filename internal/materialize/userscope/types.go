// Package userscope implements the user-scope sync pipeline for the
// materialize system. It syncs resources from KNOSSOS_HOME to ~/.claude/.
package userscope

import (
	"fmt"
	"io/fs"

	"github.com/autom8y/knossos/internal/paths"
)

// SyncUserScopeParams provides all dependencies for user-scope sync.
// Constructed by the core Sync() method from Materializer fields.
type SyncUserScopeParams struct {
	Resolver       *paths.Resolver // for ClaudeDir() (collision checker)
	EmbeddedAgents fs.FS           // fallback agent source
	EmbeddedMena   fs.FS           // fallback mena source (platform mena/)
	EmbeddedRites  fs.FS           // fallback rites source (for rites/shared/mena/)
	Opts           SyncOptions     // sync configuration
}

// SyncOptions mirrors the relevant fields from materialize.SyncOptions.
// Defined locally to avoid importing the parent package.
type SyncOptions struct {
	Resource          SyncResource
	DryRun            bool
	Recover           bool
	OverwriteDiverged bool
	KeepOrphans       bool
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

// UserSyncError is a typed error for user sync failures.
type UserSyncError struct {
	Message string
}

func (e *UserSyncError) Error() string {
	return e.Message
}

// ErrKnossosHomeNotSet returns an error when KNOSSOS_HOME is not set.
func ErrKnossosHomeNotSet() error {
	return &UserSyncError{Message: "KNOSSOS_HOME not set and no embedded fallback available"}
}

// ErrInvalidResourceType returns an error for unrecognized resource types.
func ErrInvalidResourceType() error {
	return fmt.Errorf("invalid resource type")
}
