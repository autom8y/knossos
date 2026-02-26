package materialize

import (
	"fmt"
	"strings"
)

// HookDefaults configures rite-level hook defaults merged into agent frontmatter during sync.
// Declared in manifest.yaml under the hook_defaults key.
type HookDefaults struct {
	WriteGuard *WriteGuardDefaults `yaml:"write_guard,omitempty"`
}

// WriteGuardDefaults defines the write-guard hook configuration at manifest level.
type WriteGuardDefaults struct {
	AllowPaths []string `yaml:"allow_paths,omitempty"` // Base allowed paths (shared-level)
	ExtraPaths []string `yaml:"extra_paths,omitempty"` // Additional paths (rite-level extends shared)
	Timeout    int      `yaml:"timeout,omitempty"`     // Hook timeout in seconds (default 3)
}

// WriteGuardAgent holds agent-level write-guard overrides parsed from frontmatter.
type WriteGuardAgent struct {
	ExtraPaths []string `yaml:"extra-paths,omitempty"`
}

// ResolvedWriteGuard is the final merged write-guard configuration for a single agent.
type ResolvedWriteGuard struct {
	AllowPaths []string
	Timeout    int
	AgentName  string
}

const defaultWriteGuardTimeout = 3

// ResolveHookDefaults merges shared and rite-level hook defaults (3-tier cascade, tiers 1-2).
// Returns nil if no write-guard defaults exist at either level.
func ResolveHookDefaults(shared, rite *HookDefaults) *WriteGuardDefaults {
	var sharedWG, riteWG *WriteGuardDefaults
	if shared != nil {
		sharedWG = shared.WriteGuard
	}
	if rite != nil {
		riteWG = rite.WriteGuard
	}

	if sharedWG == nil && riteWG == nil {
		return nil
	}

	result := &WriteGuardDefaults{}

	// Start with shared base paths
	if sharedWG != nil {
		result.AllowPaths = append(result.AllowPaths, sharedWG.AllowPaths...)
		result.Timeout = sharedWG.Timeout
	}

	// Append rite extra paths
	if riteWG != nil {
		result.AllowPaths = append(result.AllowPaths, riteWG.ExtraPaths...)
		// Rite-level allow_paths also contribute (if a rite defines base paths directly)
		result.AllowPaths = append(result.AllowPaths, riteWG.AllowPaths...)
		// Rite timeout overrides shared if set
		if riteWG.Timeout > 0 {
			result.Timeout = riteWG.Timeout
		}
	}

	// Deduplicate paths preserving order
	result.AllowPaths = dedup(result.AllowPaths)

	// Apply default timeout
	if result.Timeout == 0 {
		result.Timeout = defaultWriteGuardTimeout
	}

	return result
}

// ResolveWriteGuard merges manifest-level defaults with agent-level overrides (tier 3).
// Returns nil if:
//   - defaults is nil (no hook defaults at manifest level)
//   - agent opted out (write-guard: false)
//   - agent has no write-guard key (no opt-in)
func ResolveWriteGuard(defaults *WriteGuardDefaults, agentName string, agentWG interface{}) *ResolvedWriteGuard {
	if defaults == nil {
		return nil
	}
	if agentWG == nil {
		return nil // No write-guard key → no write-guard
	}

	// Handle opt-out: write-guard: false
	if b, ok := agentWG.(bool); ok {
		if !b {
			return nil // Explicit opt-out
		}
		// write-guard: true → use defaults as-is
		return &ResolvedWriteGuard{
			AllowPaths: append([]string{}, defaults.AllowPaths...),
			Timeout:    defaults.Timeout,
			AgentName:  agentName,
		}
	}

	// Handle struct form: write-guard: {extra-paths: [...]}
	var agentConfig WriteGuardAgent
	if m, ok := agentWG.(map[string]interface{}); ok {
		if eps, ok := m["extra-paths"]; ok {
			switch v := eps.(type) {
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok {
						agentConfig.ExtraPaths = append(agentConfig.ExtraPaths, s)
					}
				}
			case []string:
				agentConfig.ExtraPaths = v
			}
		}
	}

	merged := append([]string{}, defaults.AllowPaths...)
	merged = append(merged, agentConfig.ExtraPaths...)
	merged = dedup(merged)

	return &ResolvedWriteGuard{
		AllowPaths: merged,
		Timeout:    defaults.Timeout,
		AgentName:  agentName,
	}
}

// GenerateWriteGuardHooks produces a CC-compatible hooks map from a ResolvedWriteGuard.
// The output matches the format Claude Code expects in agent frontmatter:
//
//	hooks:
//	  PreToolUse:
//	    - matcher: "Write"
//	      hooks:
//	        - type: command
//	          command: "ari hook agent-guard --agent {name} --allow-path ... --output json"
//	          timeout: 3
func GenerateWriteGuardHooks(resolved *ResolvedWriteGuard) map[string]interface{} {
	if resolved == nil {
		return nil
	}

	// Build the command string
	var parts []string
	parts = append(parts, "ari hook agent-guard")
	parts = append(parts, fmt.Sprintf("--agent %s", resolved.AgentName))
	for _, p := range resolved.AllowPaths {
		parts = append(parts, fmt.Sprintf("--allow-path %s", p))
	}
	parts = append(parts, "--output json")
	command := strings.Join(parts, " ")

	hookEntry := map[string]interface{}{
		"type":    "command",
		"command": command,
		"timeout": resolved.Timeout,
	}

	matcherGroup := map[string]interface{}{
		"matcher": "Write",
		"hooks":   []interface{}{hookEntry},
	}

	return map[string]interface{}{
		"PreToolUse": []interface{}{matcherGroup},
	}
}

// dedup removes duplicate strings preserving first-occurrence order.
func dedup(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
