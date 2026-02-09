package materialize

import (
	"io/fs"
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

// loadHooksConfig finds and parses hooks.yaml from the knossos platform.
// Resolution order:
//  1. hooks/hooks.yaml in $KNOSSOS_HOME
//  2. hooks/hooks.yaml in project root (for self-hosting)
//
// Returns nil if no hooks.yaml is found (graceful).
func (m *Materializer) loadHooksConfig() *HooksConfig {
	candidates := []string{
		// Knossos platform level
		config.KnossosHome() + "/hooks/hooks.yaml",
	}
	// Project-level fallback (for satellites that bundle hooks.yaml)
	if m.resolver != nil {
		candidates = append(candidates, m.resolver.ProjectRoot()+"/hooks/hooks.yaml")
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

	// Fallback: embedded hooks.yaml (compiled into binary)
	if m.embeddedHooks != nil {
		data, err := fs.ReadFile(m.embeddedHooks, "hooks.yaml")
		if err == nil {
			var cfg HooksConfig
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				if cfg.SchemaVersion == "2.0" {
					return &cfg
				}
			}
		}
	}

	return nil
}

// buildHooksSettings generates the hooks section for settings.local.json
// from a HooksConfig. Groups entries by event type and sorts by priority.
//
// Claude Code hook format: each event maps to an array of matcher groups.
// Each matcher group has an optional "matcher" (regex string) and a "hooks"
// array of hook handlers (each with "type" and "command").
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

// mergeHooksSettings merges knossos-managed hooks into existing settings.
// Preserves user-defined matcher groups (those without "ari hook" commands).
// Replaces all knossos-managed matcher groups with the new configuration.
// Strips legacy platform hooks (bash hooks referencing missing .sh files).
//
// Returns the merged settings and a list of stripped legacy hooks for reporting.
func mergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig) (map[string]any, []string) {
	newHooks := buildHooksSettings(hooksConfig)
	var stripped []string

	// Get existing hooks section
	existingHooks, _ := existingSettings["hooks"].(map[string]any)
	if existingHooks == nil {
		existingSettings["hooks"] = newHooks
		return existingSettings, stripped
	}

	// For each event type, merge: replace ari groups, strip legacy, preserve user groups
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
		// Three-way classification: ari (skip), legacy (strip), user (preserve)
		if existingList, ok := existingHooks[event]; ok {
			if entries, ok := existingList.([]any); ok {
				for _, e := range entries {
					if group, ok := e.(map[string]any); ok {
						if isAriManagedGroup(group) {
							// Skip -- will be replaced by new ari hooks
							continue
						} else if isLegacyPlatformHook(group) {
							// Strip legacy hook and record it
							cmd := extractCommandForReport(group)
							stripped = append(stripped, event+": stripped legacy hook: "+cmd)
							continue
						} else {
							// Genuine user hook -- preserve
							userEntries = append(userEntries, group)
						}
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
	return existingSettings, stripped
}

// isAriManagedGroup checks if a matcher group is managed by ari.
// Handles both new nested format ({hooks: [{command: "ari hook ..."}]})
// and old flat format ({command: "ari hook ..."}).
func isAriManagedGroup(group map[string]any) bool {
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

// isLegacyPlatformHook checks if a matcher group contains legacy bash hooks
// that reference missing .sh files. Returns true if ANY hook in the group
// matches ANY legacy pattern.
//
// Legacy patterns:
//  1. Command contains $CLAUDE_PROJECT_DIR (env var expansion pattern)
//  2. Command contains .claude/hooks/ (legacy hook directory)
//  3. Command ends with .sh AND does not start with "ari"
//
// Handles both new nested format ({hooks: [{command: "..."}]})
// and old flat format ({command: "..."}).
func isLegacyPlatformHook(group map[string]any) bool {
	checkCommand := func(cmd string) bool {
		if cmd == "" {
			return false
		}
		// Pattern 1: Contains $CLAUDE_PROJECT_DIR
		if strings.Contains(cmd, "$CLAUDE_PROJECT_DIR") {
			return true
		}
		// Pattern 2: Contains .claude/hooks/
		if strings.Contains(cmd, ".claude/hooks/") {
			return true
		}
		// Pattern 3: Ends with .sh AND does not start with "ari"
		if strings.HasSuffix(cmd, ".sh") && !strings.HasPrefix(cmd, "ari") {
			return true
		}
		return false
	}

	// New format: check hooks array
	if hooksArr, ok := group["hooks"]; ok {
		// After JSON unmarshal: []any
		if hooks, ok := hooksArr.([]any); ok {
			for _, h := range hooks {
				if hook, ok := h.(map[string]any); ok {
					cmd, _ := hook["command"].(string)
					if checkCommand(cmd) {
						return true
					}
				}
			}
		}
		// In-memory (before JSON round-trip): []map[string]any
		if hooks, ok := hooksArr.([]map[string]any); ok {
			for _, hook := range hooks {
				cmd, _ := hook["command"].(string)
				if checkCommand(cmd) {
					return true
				}
			}
		}
	}

	// Old flat format: check top-level command field
	if cmd, ok := group["command"].(string); ok {
		return checkCommand(cmd)
	}

	return false
}

// extractCommandForReport extracts a command string from a group for reporting purposes.
// Handles both nested and flat formats, returns a truncated substring for readability.
func extractCommandForReport(group map[string]any) string {
	// Try nested format first
	if hooksArr, ok := group["hooks"]; ok {
		if hooks, ok := hooksArr.([]any); ok && len(hooks) > 0 {
			if hook, ok := hooks[0].(map[string]any); ok {
				if cmd, ok := hook["command"].(string); ok {
					return truncate(cmd, 80)
				}
			}
		}
	}
	// Try flat format
	if cmd, ok := group["command"].(string); ok {
		return truncate(cmd, 80)
	}
	return "(unknown)"
}

// truncate returns s truncated to maxLen characters, with "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
