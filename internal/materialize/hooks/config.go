// Package hooks provides hook and settings generation for the materialize pipeline.
// It reads hooks.yaml from the filesystem and produces settings.local.json content.
package hooks

import (
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/config"
)

// HooksConfig represents a parsed hooks.yaml file.
type HooksConfig struct {
	SchemaVersion string      `yaml:"schema_version"`
	Hooks         []HookEntry `yaml:"hooks"`
}

// HookEntry represents a single hook entry in hooks.yaml.
type HookEntry struct {
	Event       string `yaml:"event"`
	Matcher     string `yaml:"matcher,omitempty"`
	Command     string `yaml:"command"`
	Timeout     int    `yaml:"timeout,omitempty"`
	Async       bool   `yaml:"async,omitempty"`
	Priority    int    `yaml:"priority,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// ariHookPrefix identifies knossos-managed hooks in settings.local.json.
const ariHookPrefix = "ari hook"

// LoadHooksConfig finds and parses hooks.yaml from the filesystem.
// Resolution order:
//  1. config/hooks.yaml in $KNOSSOS_HOME
//  2. config/hooks.yaml in projectRoot (for self-hosting and satellites)
//
// Returns nil if no hooks.yaml is found (graceful).
// For fresh projects, "ari init" bootstraps config/hooks.yaml from embedded bytes.
func LoadHooksConfig(projectRoot string) *HooksConfig {
	return LoadHooksConfigWithPaths(config.KnossosHome(), projectRoot)
}

// LoadHooksConfigWithPaths finds and parses hooks.yaml using explicit paths.
// This is the DI-capable variant that avoids reading config globals.
// Resolution order:
//  1. config/hooks.yaml in knossosHome
//  2. config/hooks.yaml in projectRoot (for self-hosting and satellites)
//
// Returns nil if no hooks.yaml is found (graceful).
func LoadHooksConfigWithPaths(knossosHome, projectRoot string) *HooksConfig {
	var candidates []string
	// Knossos platform level
	if knossosHome != "" {
		candidates = append(candidates, knossosHome+"/config/hooks.yaml")
	}
	// Project-level (for self-hosting and satellites bootstrapped by ari init)
	if projectRoot != "" {
		candidates = append(candidates, projectRoot+"/config/hooks.yaml")
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cfg HooksConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue
		}

		// Validate: must have command field (v2 schema)
		if cfg.SchemaVersion != "2.0" {
			continue
		}

		return &cfg
	}

	return nil
}

// BuildHooksSettings generates the hooks section for settings.local.json
// from a HooksConfig. Groups entries by event type and sorts by priority.
//
// Claude Code hook format: each event maps to an array of matcher groups.
// Each matcher group has an optional "matcher" (regex string) and a "hooks"
// array of hook handlers (each with "type" and "command").
func BuildHooksSettings(cfg *HooksConfig, channel string) map[string]any {
	hooks := make(map[string]any)

	// Group entries by event type
	byEvent := make(map[string][]HookEntry)
	for _, entry := range cfg.Hooks {
		if entry.Command == "" {
			continue
		}
		byEvent[entry.Event] = append(byEvent[entry.Event], entry)
	}

	// Sort each event's entries by priority (lower = first)
	for event, entries := range byEvent {
		sort.Slice(entries, func(i, j int) bool {
			pi, pj := entries[i].Priority, entries[j].Priority
			if pi == 0 {
				pi = 50
			}
			if pj == 0 {
				pj = 50
			}
			return pi < pj
		})

		matcherGroups := make([]map[string]any, 0, len(entries))
		for _, entry := range entries {
			hookHandler := map[string]any{
				"type":    "command",
				"command": entry.Command,
			}
			if entry.Timeout > 0 {
				hookHandler["timeout"] = entry.Timeout
			}
			if entry.Async {
				hookHandler["async"] = true
			}
			if channel == "gemini" {
				hookHandler["env"] = map[string]string{
					"KNOSSOS_CHANNEL": "gemini",
				}
			}

			matcherGroup := map[string]any{
				"hooks": []map[string]any{hookHandler},
			}
			if entry.Matcher != "" {
				matcherGroup["matcher"] = entry.Matcher
			}
			matcherGroups = append(matcherGroups, matcherGroup)
		}

		hooks[event] = matcherGroups
	}

	return hooks
}

// MergeHooksSettings merges knossos-managed hooks into existing settings.
// Preserves user-defined matcher groups (those without "ari hook" commands).
// Replaces all knossos-managed matcher groups with the new configuration.
func MergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig, channel string) map[string]any {
	newHooks := BuildHooksSettings(hooksConfig, channel)

	// Get existing hooks section
	existingHooks, _ := existingSettings["hooks"].(map[string]any)
	if existingHooks == nil {
		existingSettings["hooks"] = newHooks
		return existingSettings
	}

	// For each event type, merge: replace ari groups, preserve user groups
	mergedHooks := make(map[string]any)

	// First, collect all event types from both sources
	allEvents := make(map[string]bool)
	for event := range existingHooks {
		allEvents[event] = true
	}
	for event := range newHooks {
		allEvents[event] = true
	}

	for event := range allEvents {
		var userEntries []map[string]any

		// Extract user-defined matcher groups from existing settings
		// Two-way classification: ari (replace), user (preserve)
		if existingList, ok := existingHooks[event]; ok {
			if entries, ok := existingList.([]any); ok {
				for _, e := range entries {
					if group, ok := e.(map[string]any); ok {
						if IsAriManagedGroup(group) {
							// Skip -- will be replaced by new ari hooks
							continue
						}
						// User hook -- preserve
						userEntries = append(userEntries, group)
					}
				}
			}
		}

		// Get new ari matcher groups for this event
		var ariEntries []map[string]any
		if newList, ok := newHooks[event]; ok {
			if entries, ok := newList.([]map[string]any); ok {
				ariEntries = entries
			}
		}

		// Combine: ari groups first (sorted by priority), then user groups
		combined := make([]map[string]any, 0, len(ariEntries)+len(userEntries))
		combined = append(combined, ariEntries...)
		combined = append(combined, userEntries...)

		if len(combined) > 0 {
			mergedHooks[event] = combined
		}
	}

	existingSettings["hooks"] = mergedHooks
	return existingSettings
}

// IsAriManagedGroup checks if a matcher group is managed by ari.
// Handles both new nested format ({hooks: [{command: "ari hook ..."}]})
// and old flat format ({command: "ari hook ..."}).
func IsAriManagedGroup(group map[string]any) bool {
	// New format: check hooks array
	if hooksArr, ok := group["hooks"]; ok {
		// After JSON unmarshal: []any
		if hooks, ok := hooksArr.([]any); ok {
			allAri := len(hooks) > 0
			for _, h := range hooks {
				if hook, ok := h.(map[string]any); ok {
					cmd, _ := hook["command"].(string)
					if !strings.HasPrefix(cmd, ariHookPrefix) {
						allAri = false
						break
					}
				}
			}
			return allAri
		}
		// In-memory (before JSON round-trip): []map[string]any
		if hooks, ok := hooksArr.([]map[string]any); ok {
			allAri := len(hooks) > 0
			for _, hook := range hooks {
				cmd, _ := hook["command"].(string)
				if !strings.HasPrefix(cmd, ariHookPrefix) {
					allAri = false
					break
				}
			}
			return allAri
		}
	}

	// Old flat format: check top-level command field
	if cmd, ok := group["command"].(string); ok {
		return strings.HasPrefix(cmd, ariHookPrefix)
	}

	return false
}
