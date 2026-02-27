// Package registry — validate.go provides rite reference validation.
// It parses rite manifests and checks that referenced agents, legomena,
// and dromena files exist on disk.
package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/mena"
)

// Warning describes a reference in a rite manifest that could not be resolved
// to an existing file.
type Warning struct {
	// File is the relative path within the rite (e.g., "agents/ghost.md").
	File string
	// RefName is the name that could not be resolved (e.g., "ghost").
	RefName string
	// Message is a human-readable description of the resolution failure.
	Message string
}

// riteManifest is a minimal struct for parsing manifest.yaml.
// It avoids importing the full materialize package.
type riteManifest struct {
	Name          string                `yaml:"name"`
	EntryAgent    string                `yaml:"entry_agent"`
	Agents        []manifestAgent       `yaml:"agents"`
	Legomena      []string              `yaml:"legomena"`
	Dromena       []string              `yaml:"dromena"`
	Dependencies  []string              `yaml:"dependencies"`
	AgentDefaults manifestAgentDefaults `yaml:"agent_defaults,omitempty"`
	SkillPolicies []manifestSkillPolicy `yaml:"skill_policies,omitempty"`
}

// manifestAgentDefaults captures only the skills field from agent_defaults.
type manifestAgentDefaults struct {
	Skills []string `yaml:"skills"`
}

// manifestSkillPolicy captures only the skill name from a skill_policies entry.
type manifestSkillPolicy struct {
	Skill string `yaml:"skill"`
}

// agentSkills is a minimal struct for extracting the skills field from
// agent frontmatter. Avoids importing internal/agent.
type agentSkills struct {
	Skills []string `yaml:"skills"`
}

// manifestAgent represents a single agent entry in a manifest.
type manifestAgent struct {
	Name string `yaml:"name"`
}

// ValidateRiteReferences parses the manifest.yaml in ritePath and checks that
// every referenced agent, legomena, and dromena file exists on disk.
//
// ritesBase is the parent directory containing all rites (e.g., "rites/").
// When non-empty, the validator checks shared and dependency mena sources
// in addition to the rite-local mena directory. When empty, only rite-local
// mena is checked.
//
// platformMenaDir is the resolved platform mena directory path (e.g., from
// the materializer's getMenaDir()). When non-empty, it is added as the
// lowest-priority source in the chain. Pass "" if not available.
//
// Returns a (possibly empty) slice of warnings and nil on success.
// Returns a non-nil error only if the manifest is missing or unparseable.
func ValidateRiteReferences(ritePath string, ritesBase string, platformMenaDir string) ([]Warning, error) {
	manifestPath := filepath.Join(ritePath, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("manifest not found at %s: %w", manifestPath, err)
	}

	var m riteManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest %s: %w", manifestPath, err)
	}

	var warnings []Warning

	// Build a set of declared agent names for entry_agent validation.
	declaredAgents := make(map[string]bool, len(m.Agents))

	// Check each agent file exists.
	for _, agent := range m.Agents {
		declaredAgents[agent.Name] = true
		agentFile := filepath.Join(ritePath, "agents", agent.Name+".md")
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			warnings = append(warnings, Warning{
				File:    filepath.Join("agents", agent.Name+".md"),
				RefName: agent.Name,
				Message: fmt.Sprintf("agent file not found: agents/%s.md", agent.Name),
			})
		}
	}

	// Build mena source chain using the unified internal/mena package.
	// Priority order (lowest to highest): platform -> shared -> dependencies -> rite-local.
	// This mirrors the materializeMena() source priority order.
	sources := mena.BuildSourceChain(mena.SourceChainOptions{
		RitePath:        ritePath,
		RitesBase:       ritesBase,
		Dependencies:    m.Dependencies,
		PlatformMenaDir: platformMenaDir,
	})

	// Check each legomena INDEX file exists in any source.
	for _, lego := range m.Legomena {
		if !mena.Exists(lego, "lego", sources) {
			warnings = append(warnings, Warning{
				File:    filepath.Join("mena", lego, "INDEX.lego.md"),
				RefName: lego,
				Message: fmt.Sprintf("legomena index not found: mena/%s/INDEX.lego.md", lego),
			})
		}
	}

	// Check each dromena exists in any source — try directory-based INDEX
	// pattern first, then flat file pattern. Both are valid knossos conventions.
	for _, dro := range m.Dromena {
		if !mena.Exists(dro, "dro", sources) {
			warnings = append(warnings, Warning{
				File:    filepath.Join("mena", dro),
				RefName: dro,
				Message: fmt.Sprintf(
					"dromena not found: neither mena/%s/INDEX.dro.md nor mena/%s.dro.md exists",
					dro, dro,
				),
			})
		}
	}

	// Validate entry_agent is declared in the agents list.
	if m.EntryAgent != "" && !declaredAgents[m.EntryAgent] {
		warnings = append(warnings, Warning{
			File:    "manifest.yaml",
			RefName: m.EntryAgent,
			Message: fmt.Sprintf("entry_agent %q is not in the agents list", m.EntryAgent),
		})
	}

	// Validate agent_defaults.skills — manifest-level, one check per skill.
	for _, skill := range m.AgentDefaults.Skills {
		if !mena.Exists(skill, "lego", sources) {
			warnings = append(warnings, Warning{
				File:    "manifest.yaml",
				RefName: skill,
				Message: fmt.Sprintf("agent_defaults.skills: skill %q not found in mena sources", skill),
			})
		}
	}

	// Validate skill_policies[].skill — manifest-level.
	for _, policy := range m.SkillPolicies {
		if policy.Skill == "" {
			continue
		}
		if !mena.Exists(policy.Skill, "lego", sources) {
			warnings = append(warnings, Warning{
				File:    "manifest.yaml",
				RefName: policy.Skill,
				Message: fmt.Sprintf("skill_policies: skill %q not found in mena sources", policy.Skill),
			})
		}
	}

	// Validate per-agent frontmatter skills.
	for _, agent := range m.Agents {
		agentFile := filepath.Join(ritePath, "agents", agent.Name+".md")
		content, err := os.ReadFile(agentFile)
		if err != nil {
			continue // already warned above if file is missing
		}
		yamlBytes, _, err := frontmatter.Parse(content)
		if err != nil {
			continue // not a skill validation concern
		}
		var as agentSkills
		if err := yaml.Unmarshal(yamlBytes, &as); err != nil {
			continue
		}
		for _, skill := range as.Skills {
			if !mena.Exists(skill, "lego", sources) {
				warnings = append(warnings, Warning{
					File:    filepath.Join("agents", agent.Name+".md"),
					RefName: skill,
					Message: fmt.Sprintf("skill %q not found in mena sources", skill),
				})
			}
		}
	}

	return warnings, nil
}
