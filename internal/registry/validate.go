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
	Name         string          `yaml:"name"`
	EntryAgent   string          `yaml:"entry_agent"`
	Agents       []manifestAgent `yaml:"agents"`
	Legomena     []string        `yaml:"legomena"`
	Dromena      []string        `yaml:"dromena"`
	Dependencies []string        `yaml:"dependencies"`
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
// Returns a (possibly empty) slice of warnings and nil on success.
// Returns a non-nil error only if the manifest is missing or unparseable.
func ValidateRiteReferences(ritePath string, ritesBase string) ([]Warning, error) {
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

	// Build mena lookup chain: rite-local -> shared -> dependencies.
	// This mirrors the materializeMena() source priority order.
	menaDirs := buildMenaDirs(ritePath, ritesBase, m.Dependencies)

	// Check each legomena INDEX file exists in any source.
	for _, lego := range m.Legomena {
		if !menaExistsInSources(lego, "lego", menaDirs) {
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
		if !menaExistsInSources(dro, "dro", menaDirs) {
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

// buildMenaDirs returns the ordered list of mena directories to search.
// Priority: rite-local -> shared -> each dependency.
func buildMenaDirs(ritePath string, ritesBase string, dependencies []string) []string {
	dirs := []string{filepath.Join(ritePath, "mena")}

	if ritesBase == "" {
		return dirs
	}

	// Shared rite mena.
	dirs = append(dirs, filepath.Join(ritesBase, "shared", "mena"))

	// Dependency rite mena (in manifest order).
	for _, dep := range dependencies {
		if dep != "shared" {
			dirs = append(dirs, filepath.Join(ritesBase, dep, "mena"))
		}
	}

	return dirs
}

// menaExistsInSources checks whether a mena entry can be resolved from any of
// the provided mena source directories.
// For legomena: checks for {dir}/{name}/INDEX.lego.md
// For dromena: checks for {dir}/{name}/INDEX.dro.md or {dir}/{name}.dro.md
func menaExistsInSources(name string, menaType string, menaDirs []string) bool {
	for _, dir := range menaDirs {
		switch menaType {
		case "lego":
			if _, err := os.Stat(filepath.Join(dir, name, "INDEX.lego.md")); err == nil {
				return true
			}
		case "dro":
			if _, err := os.Stat(filepath.Join(dir, name, "INDEX.dro.md")); err == nil {
				return true
			}
			if _, err := os.Stat(filepath.Join(dir, name+".dro.md")); err == nil {
				return true
			}
		}
	}
	return false
}
