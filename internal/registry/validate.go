// Package registry — validate.go provides rite reference validation.
// It parses rite manifests and checks that referenced agents, legomena,
// and dromena files exist on disk.
package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	Name       string          `yaml:"name"`
	EntryAgent string          `yaml:"entry_agent"`
	Agents     []manifestAgent `yaml:"agents"`
	Legomena   []string        `yaml:"legomena"`
	Dromena    []string        `yaml:"dromena"`
}

// manifestAgent represents a single agent entry in a manifest.
type manifestAgent struct {
	Name string `yaml:"name"`
}

// ValidateRiteReferences parses the manifest.yaml in ritePath and checks that
// every referenced agent, legomena, and dromena file exists on disk.
//
// Returns a (possibly empty) slice of warnings and nil on success.
// Returns a non-nil error only if the manifest is missing or unparseable.
func ValidateRiteReferences(ritePath string) ([]Warning, error) {
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

	// Check each legomena INDEX file exists.
	for _, lego := range m.Legomena {
		indexFile := filepath.Join(ritePath, "mena", lego, "INDEX.lego.md")
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			warnings = append(warnings, Warning{
				File:    filepath.Join("mena", lego, "INDEX.lego.md"),
				RefName: lego,
				Message: fmt.Sprintf("legomena index not found: mena/%s/INDEX.lego.md", lego),
			})
		}
	}

	// Check each dromena exists — try directory-based INDEX pattern first,
	// then flat file pattern. Both are valid knossos conventions.
	for _, dro := range m.Dromena {
		indexPattern := filepath.Join(ritePath, "mena", dro, "INDEX.dro.md")
		flatPattern := filepath.Join(ritePath, "mena", dro+".dro.md")

		_, errIndex := os.Stat(indexPattern)
		_, errFlat := os.Stat(flatPattern)

		if os.IsNotExist(errIndex) && os.IsNotExist(errFlat) {
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

	return warnings, nil
}
