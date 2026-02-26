package materialize

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
// and injects matching skills. For Sprint 2, only mode:"inject" is handled.
// mode:"reference" policies are skipped (implemented in Sprint 3).
//
// Injection order: policy-injected skills are prepended BEFORE agent-declared skills.
// Deduplication preserves first-occurrence order.
func applySkillPolicies(fmMap map[string]interface{}, policies []SkillPolicy) map[string]interface{} {
	if len(policies) == 0 {
		return fmMap
	}

	// Parse skill_policy_exclude directive
	excludeAll, excludeSet := parseSkillPolicyExclude(fmMap)
	if excludeAll {
		return fmMap
	}

	// Build tools set from agent frontmatter (post agent_defaults merge)
	toolsSet := parseToolsSet(fmMap, "tools")

	// Build disallowedTools set for requires_none matching
	disallowedSet := parseToolsSet(fmMap, "disallowedTools")

	// Collect skills to inject (policy order = injection order)
	toInject := make([]string, 0, len(policies))
	for _, policy := range policies {
		// Only handle inject mode in this sprint
		if policy.Mode != "inject" {
			continue
		}

		// Skip if this skill is explicitly excluded
		if excludeSet[policy.Skill] {
			continue
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

		toInject = append(toInject, policy.Skill)
	}

	if len(toInject) == 0 {
		return fmMap
	}

	// Build merged skills: policy-injected BEFORE agent-declared, deduplicated
	agentSkills := toStringSlice(fmMap["skills"])

	// Combine: injected first, then agent-declared, dedup
	combined := make([]string, 0, len(toInject)+len(agentSkills))
	combined = append(combined, toInject...)
	combined = append(combined, agentSkills...)
	deduplicated := dedup(combined)

	fmMap["skills"] = toInterfaceSlice(deduplicated)
	return fmMap
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
// Follows the same pattern as loadSharedHookDefaults in agent_transform.go.
func (m *Materializer) loadSharedSkillPolicies(resolved *ResolvedRite) []SkillPolicy {
	var data []byte
	var err error

	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.EmbeddedFS != nil {
		// Embedded FS: look for shared manifest relative to rite
		data, err = fs.ReadFile(m.sourceResolver.EmbeddedFS, "rites/shared/manifest.yaml")
	} else {
		// Filesystem: look relative to project root
		sharedPath := filepath.Join(m.resolver.ProjectRoot(), "rites", "shared", "manifest.yaml")
		data, err = os.ReadFile(sharedPath)
	}
	if err != nil {
		return nil // Shared manifest not found — graceful degradation
	}

	var manifest RiteManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		log.Printf("Warning: failed to parse shared manifest for skill_policies: %v", err)
		return nil
	}

	return manifest.SkillPolicies
}
