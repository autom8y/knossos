package materialize

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
	Priority    int    `yaml:"priority,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// ariHookPrefix identifies knossos-managed hooks in settings.local.json.
const ariHookPrefix = "ari hook"

// loadHooksConfig finds and parses hooks.yaml from the knossos platform.
// Resolution order:
//  1. user-hooks/ari/hooks.yaml in $KNOSSOS_HOME
//  2. user-hooks/ari/hooks.yaml in project root (for self-hosting)
//
// Returns nil if no hooks.yaml is found (graceful).
func (m *Materializer) loadHooksConfig() *HooksConfig {
	candidates := []string{
		// Knossos platform level
		config.KnossosHome() + "/user-hooks/ari/hooks.yaml",
	}
	// Project-level fallback (for satellites that bundle hooks.yaml)
	if m.resolver != nil {
		candidates = append(candidates, m.resolver.ProjectRoot()+"/user-hooks/ari/hooks.yaml")
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

// buildHooksSettings generates the hooks section for settings.local.json
// from a HooksConfig. Groups entries by event type and sorts by priority.
func buildHooksSettings(cfg *HooksConfig) map[string]any {
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

		hookList := make([]map[string]any, 0, len(entries))
		for _, entry := range entries {
			hookEntry := map[string]any{
				"command": entry.Command,
			}
			if entry.Matcher != "" {
				hookEntry["matcher"] = entry.Matcher
			}
			hookList = append(hookList, hookEntry)
		}

		hooks[event] = hookList
	}

	return hooks
}

// mergeHooksSettings merges knossos-managed hooks into existing settings.
// Preserves user-defined hooks (commands not starting with "ari hook").
// Replaces all knossos-managed hooks with the new configuration.
func mergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig) map[string]any {
	newHooks := buildHooksSettings(hooksConfig)

	// Get existing hooks section
	existingHooks, _ := existingSettings["hooks"].(map[string]any)
	if existingHooks == nil {
		existingSettings["hooks"] = newHooks
		return existingSettings
	}

	// For each event type, merge: replace ari hooks, preserve user hooks
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

		// Extract user-defined hooks from existing (non-ari commands)
		if existingList, ok := existingHooks[event]; ok {
			if entries, ok := existingList.([]any); ok {
				for _, e := range entries {
					if entry, ok := e.(map[string]any); ok {
						cmd, _ := entry["command"].(string)
						if !strings.HasPrefix(cmd, ariHookPrefix) {
							userEntries = append(userEntries, entry)
						}
					}
				}
			}
		}

		// Get new ari hooks for this event
		var ariEntries []map[string]any
		if newList, ok := newHooks[event]; ok {
			if entries, ok := newList.([]map[string]any); ok {
				ariEntries = entries
			}
		}

		// Combine: ari hooks first (sorted by priority), then user hooks
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
