package materialize

import (
	"log/slog"
	"strings"
)

// SkillPolicy defines a single capability-driven skill wiring rule.
// Declared at manifest level (shared or rite), evaluated per-agent during sync.
type SkillPolicy struct {
	Skill         string   `yaml:"skill"`
	Mode          string   `yaml:"mode"`                    // "inject" or "reference"
	RequiresTools []string `yaml:"requires_tools,omitempty"`
	RequiresNone  []string `yaml:"requires_none,omitempty"`
}

// MergeSkillPolicies merges shared and rite-level skill policies.
// Rite policies win on same skill name (rite overrides shared).
// Preserves order: shared-first, rite-appended (for non-overridden rite policies).
func MergeSkillPolicies(shared, rite []SkillPolicy) []SkillPolicy {
	if len(shared) == 0 && len(rite) == 0 {
		return nil
	}

	// Build a map of shared policies by skill name for O(1) override detection
	sharedBySkill := make(map[string]int, len(shared)) // skill -> index in result
	result := make([]SkillPolicy, 0, len(shared)+len(rite))

	for i, p := range shared {
		result = append(result, p)
		sharedBySkill[p.Skill] = i
	}

	// For each rite policy: replace shared entry if same skill, else append
	for _, p := range rite {
		if idx, exists := sharedBySkill[p.Skill]; exists {
			// Rite wins: replace in-place to preserve shared ordering
			result[idx] = p
		} else {
			// New skill not in shared: append
			result = append(result, p)
		}
	}

	return result
}

// applySkillPolicies evaluates skill policies against a single agent's frontmatter map
// and applies matching skills. Supports two modes:
//   - inject: adds skill to fmMap["skills"] (prepended before agent-declared skills)
//   - reference: prepends an HTML comment to body directing agent to invoke via Skill tool
//
// Dead reference guard: if mode is "reference" and the agent has "Skill" in
// disallowedTools, the reference comment is silently skipped (agent cannot use Skill tool).
//
// Agent overrides: skill_policy_override frontmatter field allows per-agent mode overrides.
// If an agent overrides a policy's mode, the override mode is used instead.
// Exclude always wins over override (excluded skills are never applied).
//
// Returns the modified fmMap AND the modified body.
func applySkillPolicies(fmMap map[string]interface{}, body []byte, policies []SkillPolicy) (map[string]interface{}, []byte) {
	if len(policies) == 0 {
		return fmMap, body
	}

	// Parse skill_policy_exclude directive
	excludeAll, excludeSet := parseSkillPolicyExclude(fmMap)
	if excludeAll {
		return fmMap, body
	}

	// Build tools set from agent frontmatter (post agent_defaults merge)
	toolsSet := parseToolsSet(fmMap, "tools")

	// Build disallowedTools set for requires_none matching and dead reference guard
	disallowedSet := parseToolsSet(fmMap, "disallowedTools")

	// Parse agent-level overrides: skill_policy_override: [{skill: foo, mode: inject}]
	overrideMap := parseSkillPolicyOverride(fmMap)

	// Collect skills to inject and reference comments to prepend
	toInject := make([]string, 0, len(policies))
	toReference := make([]string, 0, len(policies))

	for _, policy := range policies {
		// Exclude wins over everything — skip entirely
		if excludeSet[policy.Skill] {
			continue
		}

		// Determine effective mode: agent override wins over policy mode
		effectiveMode := policy.Mode
		if override, hasOverride := overrideMap[policy.Skill]; hasOverride {
			effectiveMode = override
		}

		// requires_tools: skip if agent lacks ANY required tool
		if len(policy.RequiresTools) > 0 {
			allPresent := true
			for _, req := range policy.RequiresTools {
				if !toolsSet[req] {
					allPresent = false
					break
				}
			}
			if !allPresent {
				continue
			}
		}

		// requires_none: skip if agent HAS any of these in disallowedTools
		if len(policy.RequiresNone) > 0 {
			blocked := false
			for _, none := range policy.RequiresNone {
				if disallowedSet[none] {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}
		}

		switch effectiveMode {
		case "inject":
			toInject = append(toInject, policy.Skill)
		case "reference":
			// Dead reference guard: if agent cannot use Skill tool, skip silently
			if disallowedSet["Skill"] {
				continue
			}
			toReference = append(toReference, policy.Skill)
		}
	}

	// Apply inject: prepend policy skills before agent-declared skills, deduplicated
	if len(toInject) > 0 {
		agentSkills := toStringSlice(fmMap["skills"])
		combined := make([]string, 0, len(toInject)+len(agentSkills))
		combined = append(combined, toInject...)
		combined = append(combined, agentSkills...)
		fmMap["skills"] = toInterfaceSlice(dedup(combined))
	}

	// Apply reference: prepend HTML comments to body, one per skill
	if len(toReference) > 0 {
		var commentLines []byte
		for _, skillName := range toReference {
			line := "<!-- skill_policies: " + skillName + " (invoke via Skill tool when needed) -->\n"
			commentLines = append(commentLines, []byte(line)...)
		}
		// Prepend comments before existing body
		body = append(commentLines, body...)
	}

	return fmMap, body
}

// parseSkillPolicyOverride reads the skill_policy_override field from the frontmatter map.
// Returns a map of skill-name → override-mode for O(1) lookup during policy evaluation.
// The field is expected to be a list of {skill: name, mode: mode} objects.
func parseSkillPolicyOverride(fmMap map[string]interface{}) map[string]string {
	val, ok := fmMap["skill_policy_override"]
	if !ok {
		return nil
	}

	// Expected YAML structure: list of maps with "skill" and "mode" keys
	items, ok := val.([]interface{})
	if !ok {
		return nil
	}

	overrides := make(map[string]string, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		skillVal, hasSkill := entry["skill"]
		modeVal, hasMode := entry["mode"]
		if !hasSkill || !hasMode {
			continue
		}
		skillName, ok1 := skillVal.(string)
		modeName, ok2 := modeVal.(string)
		if ok1 && ok2 && skillName != "" && modeName != "" {
			overrides[skillName] = modeName
		}
	}

	if len(overrides) == 0 {
		return nil
	}
	return overrides
}

// parseSkillPolicyExclude reads the skill_policy_exclude field from the frontmatter map.
// Returns (true, nil) if "all" exclusion, or (false, set) for a specific exclusion list.
func parseSkillPolicyExclude(fmMap map[string]interface{}) (excludeAll bool, excludeSet map[string]bool) {
	val, ok := fmMap["skill_policy_exclude"]
	if !ok {
		return false, nil
	}

	// String "all" means exclude all policies
	if s, ok := val.(string); ok {
		if s == "all" {
			return true, nil
		}
		// Single string (not "all") — treat as single-item exclusion set
		return false, map[string]bool{s: true}
	}

	// String slice: build exclusion set
	items := toStringSlice(val)
	if len(items) == 0 {
		return false, nil
	}
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return false, set
}

// parseToolsSet parses a tools-like field from fmMap into a set for O(1) lookup.
// The field can be a YAML list ["Bash", "Read"] or a comma-separated string "Bash, Read".
func parseToolsSet(fmMap map[string]interface{}, fieldName string) map[string]bool {
	val, ok := fmMap[fieldName]
	if !ok {
		return nil
	}

	// Try YAML list first
	items := toStringSlice(val)
	if items != nil {
		set := make(map[string]bool, len(items))
		for _, item := range items {
			set[strings.TrimSpace(item)] = true
		}
		return set
	}

	// Try comma-separated string
	if s, ok := val.(string); ok {
		parts := strings.Split(s, ",")
		set := make(map[string]bool, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				set[trimmed] = true
			}
		}
		return set
	}

	return nil
}

// loadSharedSkillPolicies loads skill_policies from the shared rite manifest.
// Returns nil if the shared manifest doesn't exist or has no skill_policies.
// Delegates to loadSharedManifest (see agent_transform.go) for the shared load path.
func (m *Materializer) loadSharedSkillPolicies(resolved *ResolvedRite) []SkillPolicy {
	manifest, err := m.loadSharedManifest(resolved)
	if err != nil {
		slog.Warn("failed to parse shared manifest for skill_policies", "error", err)
		return nil
	}
	if manifest == nil {
		return nil
	}
	return manifest.SkillPolicies
}
