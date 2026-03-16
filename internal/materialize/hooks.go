// Package materialize re-exports hooks types from the hooks sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/hooks"
)

// Type aliases for backward compatibility.
type (
	HooksConfig    = hooks.HooksConfig
	HookEntry      = hooks.HookEntry
	MCPPoolsConfig = hooks.MCPPoolsConfig
	MCPPool        = hooks.MCPPool
	MCPPoolRef     = hooks.MCPPoolRef
	MCPServerConfig = hooks.MCPServerConfig
)

// Re-export functions (used by core tests and other core code).
var (
	mergeHooksSettings   = hooks.MergeHooksSettings
	loadExistingSettings = hooks.LoadExistingSettings
	saveSettings         = hooks.SaveSettings
)

// loadHooksConfig delegates to hooks.LoadHooksConfig with the Materializer's project root.
func (m *Materializer) loadHooksConfig() *HooksConfig {
	var projectRoot string
	if m.resolver != nil {
		projectRoot = m.resolver.ProjectRoot()
	}
	return hooks.LoadHooksConfig(projectRoot)
}

// loadMCPPoolsConfig delegates to hooks.LoadMCPPoolsConfig with the Materializer's project root.
func (m *Materializer) loadMCPPoolsConfig() *MCPPoolsConfig {
	var projectRoot string
	if m.resolver != nil {
		projectRoot = m.resolver.ProjectRoot()
	}
	return hooks.LoadMCPPoolsConfig(projectRoot)
}
