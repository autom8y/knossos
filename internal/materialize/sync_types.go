// Package materialize provides unified sync types for both rite and user scopes.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/userscope"
)

// SyncScope determines which scopes to execute during sync.
type SyncScope string

const (
	ScopeAll  SyncScope = "all"
	ScopeRite SyncScope = "rite"
	ScopeOrg  SyncScope = "org"
	ScopeUser SyncScope = "user"
)

// IsValid returns true if the scope is a recognized value.
func (s SyncScope) IsValid() bool {
	switch s {
	case ScopeAll, ScopeRite, ScopeOrg, ScopeUser:
		return true
	default:
		return false
	}
}

// SyncResource is a type alias to userscope.SyncResource for backward compatibility.
// This ensures the map key type in UserScopeResult.Resources matches core constants.
type SyncResource = userscope.SyncResource

// Re-export resource constants from the userscope package.
const (
	ResourceAll    = userscope.ResourceAll
	ResourceAgents = userscope.ResourceAgents
	ResourceMena   = userscope.ResourceMena
	ResourceHooks  = userscope.ResourceHooks
)

// SyncOptions configures the unified sync pipeline.
type SyncOptions struct {
	Scope             SyncScope
	RiteName          string
	Source            string
	Resource          SyncResource
	OrgName           string // Org name for org-scope sync (empty = use config.ActiveOrg())
	DryRun            bool
	Recover           bool
	OverwriteDiverged bool
	KeepOrphans       bool
	Soft              bool // CC-safe mode: only update agents + CLAUDE.md
	ElCheapo          bool // Force all agents to haiku model (ephemeral, reverted on session exit)
}

// SyncResult contains unified outcome.
type SyncResult struct {
	RiteResult *RiteScopeResult `json:"rite,omitempty"`
	OrgResult  *OrgScopeResult  `json:"org,omitempty"`
	UserResult *UserScopeResult `json:"user,omitempty"`
}

// OrgScopeResult wraps org scope sync outcome.
type OrgScopeResult struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	OrgName string `json:"org_name,omitempty"`
	Source  string `json:"source,omitempty"`
	Agents  int    `json:"agents,omitempty"`
	Mena    int    `json:"mena,omitempty"`
}

// RiteScopeResult wraps rite scope outcome.
type RiteScopeResult struct {
	Status                string   `json:"status"`
	Error                 string   `json:"error,omitempty"`
	RiteName              string   `json:"rite_name,omitempty"`
	Source                string   `json:"source,omitempty"`
	SourcePath            string   `json:"source_path,omitempty"`
	OrphansDetected       []string `json:"orphans_detected,omitempty"`
	OrphanAction          string   `json:"orphan_action,omitempty"`
	BackupPath            string   `json:"backup_path,omitempty"`
	LegacyBackupPath      string   `json:"legacy_backup_path,omitempty"`
	SoftMode              bool     `json:"soft_mode,omitempty"`              // true if soft mode was used
	DeferredStages        []string `json:"deferred_stages,omitempty"`        // stages skipped in soft mode
	RiteSwitched          bool     `json:"rite_switched,omitempty"`          // true if rite changed from previous
	PreviousRite          string   `json:"previous_rite,omitempty"`          // previous ACTIVE_RITE name
	ThroughlineIDsCleaned int      `json:"throughline_ids_cleaned,omitempty"` // count of .throughline-ids.json files removed
	ElCheapoMode          bool     `json:"el_cheapo_mode,omitempty"`          // true if el-cheapo model override active
}

// Type aliases for user-scope types from the userscope sub-package.
type (
	UserScopeResult    = userscope.UserScopeResult
	UserResourceResult = userscope.UserResourceResult
	UserSyncChanges    = userscope.UserSyncChanges
	UserSkippedEntry   = userscope.UserSkippedEntry
	UserSyncSummary    = userscope.UserSyncSummary
	UserResourceError  = userscope.UserResourceError
)
