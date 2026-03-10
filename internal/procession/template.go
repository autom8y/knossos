// Package procession implements cross-rite coordinated workflow templates.
// A procession is a multi-station workflow that spans multiple rites,
// analogous to how a rite workflow spans multiple phases within a single rite.
package procession

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"gopkg.in/yaml.v3"
)

// namePattern validates station and template names: lowercase letters, digits, hyphens.
// Must start with a lowercase letter.
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Template represents a procession template that defines a cross-rite
// coordinated workflow. Templates are YAML files in the processions/
// directory, analogous to how workflow.yaml defines intra-rite workflows.
type Template struct {
	Name        string    `yaml:"name"         json:"name"`
	Description string    `yaml:"description"  json:"description"`
	Stations    []Station `yaml:"stations"     json:"stations"`
	ArtifactDir string    `yaml:"artifact_dir" json:"artifact_dir"`
}

// Station represents a single station in a procession template.
// Each station maps to a rite that handles that phase of the workflow.
type Station struct {
	Name     string   `yaml:"name"               json:"name"`
	Rite     string   `yaml:"rite"               json:"rite"`
	AltRite  string   `yaml:"alt_rite,omitempty" json:"alt_rite,omitempty"`
	Goal     string   `yaml:"goal"               json:"goal"`
	Produces []string `yaml:"produces"           json:"produces"`
	LoopTo   string   `yaml:"loop_to,omitempty"  json:"loop_to,omitempty"`
}

// Validate checks the template against schema requirements.
// Returns nil if valid; returns an error listing all violations if invalid.
// All violations are collected before returning (no fail-fast) to provide
// complete diagnostic output.
func (t *Template) Validate() error {
	var issues []string

	// 1. Validate template name
	if t.Name == "" {
		issues = append(issues, "name: must not be empty")
	} else if len(t.Name) > 64 {
		issues = append(issues, fmt.Sprintf("name: must be at most 64 characters (got %d)", len(t.Name)))
	} else if !namePattern.MatchString(t.Name) {
		issues = append(issues, fmt.Sprintf("name: %q does not match pattern ^[a-z][a-z0-9-]*$", t.Name))
	}

	// 2. Validate description
	if t.Description == "" {
		issues = append(issues, "description: must not be empty")
	} else if len(t.Description) > 200 {
		issues = append(issues, fmt.Sprintf("description: must be at most 200 characters (got %d)", len(t.Description)))
	}

	// 3. Validate stations count — a single-station procession is just a regular rite workflow
	if len(t.Stations) < 2 {
		issues = append(issues, fmt.Sprintf("stations: at least 2 stations required (got %d)", len(t.Stations)))
	}

	// 4. Validate artifact_dir
	if !strings.HasPrefix(t.ArtifactDir, ".sos/wip/") {
		issues = append(issues, fmt.Sprintf("artifact_dir: must start with .sos/wip/ (got %q)", t.ArtifactDir))
	}

	// Build station name set for loop_to validation and duplicate detection.
	// We build this before the per-station loop so loop_to can reference any station
	// in the template, not just those seen so far.
	stationNames := make(map[string]int) // name -> first occurrence index
	for i, s := range t.Stations {
		if s.Name != "" {
			if prev, exists := stationNames[s.Name]; exists {
				issues = append(issues, fmt.Sprintf("stations[%d].name: duplicate station name %q (first seen at index %d)", i, s.Name, prev))
			} else {
				stationNames[s.Name] = i
			}
		}
	}

	// 5. Validate each station's fields
	for i, s := range t.Stations {
		prefix := fmt.Sprintf("stations[%d]", i)

		// 5a. Station name pattern and length
		if s.Name == "" {
			issues = append(issues, fmt.Sprintf("%s.name: must not be empty", prefix))
		} else if len(s.Name) > 32 {
			issues = append(issues, fmt.Sprintf("%s.name: must be at most 32 characters (got %d)", prefix, len(s.Name)))
		} else if !namePattern.MatchString(s.Name) {
			issues = append(issues, fmt.Sprintf("%s.name: %q does not match pattern ^[a-z][a-z0-9-]*$", prefix, s.Name))
		}

		// 5b. Rite must be non-empty (existence validated at ari procession create time)
		if s.Rite == "" {
			issues = append(issues, fmt.Sprintf("%s.rite: must not be empty", prefix))
		}

		// 5c. Goal
		if s.Goal == "" {
			issues = append(issues, fmt.Sprintf("%s.goal: must not be empty", prefix))
		} else if len(s.Goal) > 500 {
			issues = append(issues, fmt.Sprintf("%s.goal: must be at most 500 characters (got %d)", prefix, len(s.Goal)))
		}

		// 5d. Produces must have at least 1 element; each element non-empty
		if len(s.Produces) == 0 {
			issues = append(issues, fmt.Sprintf("%s.produces: must have at least 1 element", prefix))
		} else {
			for j, p := range s.Produces {
				if p == "" {
					issues = append(issues, fmt.Sprintf("%s.produces[%d]: must not be empty", prefix, j))
				}
			}
		}

		// 5e. loop_to, if set, must reference a station that exists in this template
		if s.LoopTo != "" {
			if _, exists := stationNames[s.LoopTo]; !exists {
				issues = append(issues, fmt.Sprintf("%s.loop_to: references unknown station %q", prefix, s.LoopTo))
			}
		}
	}

	if len(issues) > 0 {
		return errors.New(errors.CodeSchemaInvalid,
			fmt.Sprintf("procession template validation failed:\n  - %s", strings.Join(issues, "\n  - ")))
	}
	return nil
}

// LoadTemplate reads and parses a procession template from a YAML file on disk.
// Returns the validated template or an error.
func LoadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.CodeFileNotFound,
				fmt.Sprintf("procession template not found: %s", path))
		}
		return nil, errors.Wrap(errors.CodeGeneralError,
			fmt.Sprintf("failed to read procession template: %s", path), err)
	}

	var t Template
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid,
			fmt.Sprintf("invalid YAML in procession template: %s", path), err)
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}
	return &t, nil
}

// LoadEmbeddedTemplate reads and parses a procession template from an fs.FS
// (typically an embed.FS declared by the caller with //go:embed processions/).
// The name parameter is the template name without path prefix or extension;
// e.g., "security-remediation" resolves to "processions/security-remediation.yaml".
func LoadEmbeddedTemplate(name string, embedded fs.FS) (*Template, error) {
	path := fmt.Sprintf("processions/%s.yaml", name)

	data, err := fs.ReadFile(embedded, path)
	if err != nil {
		return nil, errors.New(errors.CodeFileNotFound,
			fmt.Sprintf("embedded procession template not found: %s", name))
	}

	var t Template
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid,
			fmt.Sprintf("invalid YAML in embedded procession template: %s", name), err)
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}
	return &t, nil
}

// GetStation returns a pointer to the station with the given name, or nil if not found.
// Follows the rite.Workflow.GetPhase() pattern.
func (t *Template) GetStation(name string) *Station {
	for i := range t.Stations {
		if t.Stations[i].Name == name {
			return &t.Stations[i]
		}
	}
	return nil
}

// StationNames returns the ordered list of station names.
// Follows the rite.Workflow.PhaseNames() pattern.
func (t *Template) StationNames() []string {
	names := make([]string, len(t.Stations))
	for i, s := range t.Stations {
		names[i] = s.Name
	}
	return names
}

// NextStation returns the name of the station after current in the ordered sequence,
// or an empty string if current is the last station or not found.
// Used by ari procession proceed to compute NextStation and NextRite session fields.
func (t *Template) NextStation(current string) string {
	for i, s := range t.Stations {
		if s.Name == current && i+1 < len(t.Stations) {
			return t.Stations[i+1].Name
		}
	}
	return ""
}
