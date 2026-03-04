package materialize

import "maps"

// appendListFields are list fields where agent values are APPENDED to defaults.
// Deduplication preserves first-occurrence order (defaults before agent).
// skills was removed: skill injection is now handled exclusively by skill_policies.
var appendListFields = map[string]bool{}

// replaceListFields are list fields where agent values REPLACE defaults entirely.
// If the agent doesn't define the field, the default is used.
var replaceListFields = map[string]bool{
	"disallowedTools": true,
	"tools":           true,
	"allowedTools":    true,
}

// MergeAgentDefaults merges manifest-level agent_defaults into agent frontmatter.
// Agent-level values take precedence over defaults.
//
// Merge semantics:
//   - Scalars: agent overrides default
//   - skills: scalar semantics (agent replaces default entirely); skill injection is
//     handled exclusively by skill_policies — not agent_defaults
//   - disallowedTools, tools, allowedTools: REPLACE (agent overrides entirely)
//   - Maps: deep merge, agent keys win
//   - Missing agent value: use default
//   - Missing default: use agent value as-is
//   - Nil/empty defaults: no-op, return agent frontmatter unchanged
func MergeAgentDefaults(defaults map[string]any, agentFM map[string]any) map[string]any {
	if len(defaults) == 0 {
		return agentFM
	}
	if agentFM == nil {
		agentFM = make(map[string]any)
	}

	result := make(map[string]any, len(agentFM)+len(defaults))

	// Start with all agent values
	maps.Copy(result, agentFM)

	// Merge defaults for keys not in agent, or apply field-specific merge
	for key, defaultVal := range defaults {
		agentVal, agentHasKey := result[key]

		if !agentHasKey {
			// Agent doesn't have this field — use default
			result[key] = defaultVal
			continue
		}

		// Field exists in both — apply merge strategy by field type
		if appendListFields[key] {
			result[key] = mergeAppendLists(defaultVal, agentVal)
			continue
		}

		if replaceListFields[key] {
			// Agent value replaces default entirely (already in result)
			continue
		}

		// Check if both values are maps — deep merge
		defaultMap, defaultIsMap := toStringMap(defaultVal)
		agentMap, agentIsMap := toStringMap(agentVal)
		if defaultIsMap && agentIsMap {
			result[key] = deepMergeMaps(defaultMap, agentMap)
			continue
		}

		// Scalar: agent value wins (already in result)
	}

	return result
}

// mergeAppendLists appends agent list items to default list items, deduplicating.
// Returns the merged list preserving first-occurrence order (defaults first, then agent).
func mergeAppendLists(defaultVal, agentVal any) any {
	defaultItems := toStringSlice(defaultVal)
	agentItems := toStringSlice(agentVal)

	merged := make([]string, 0, len(defaultItems)+len(agentItems))
	merged = append(merged, defaultItems...)
	merged = append(merged, agentItems...)

	return toInterfaceSlice(dedup(merged))
}

// deepMergeMaps merges two string-keyed maps. Agent keys win on conflict.
func deepMergeMaps(base, override map[string]any) map[string]any {
	result := make(map[string]any, len(base)+len(override))

	maps.Copy(result, base)

	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			baseMap, baseIsMap := toStringMap(baseVal)
			overrideMap, overrideIsMap := toStringMap(v)
			if baseIsMap && overrideIsMap {
				result[k] = deepMergeMaps(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}

	return result
}

// toStringMap attempts to convert an interface{} to map[string]interface{}.
// Handles both Go-native maps and YAML-unmarshaled maps.
func toStringMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		// YAML sometimes produces this type
		result := make(map[string]any, len(m))
		for k, val := range m {
			if ks, ok := k.(string); ok {
				result[ks] = val
			}
		}
		return result, true
	}
	return nil, false
}

// toStringSlice converts an interface{} that may be a []interface{} or []string
// into a []string. Returns nil if the value is not a recognized list type.
func toStringSlice(v any) []string {
	switch s := v.(type) {
	case []any:
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case []string:
		return s
	}
	return nil
}

// toInterfaceSlice converts []string to []interface{} for YAML map compatibility.
func toInterfaceSlice(items []string) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
