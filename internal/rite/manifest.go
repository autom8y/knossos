// Package rite implements rite manifest handling, discovery, and invocation for Ariadne.
// Rites are composable practice bundles that can be invoked (additive) or swapped (replacement).
package rite

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
)

// RiteForm classifies a rite by its composition of components.
type RiteForm string

const (
	// FormSimple represents a rite with mena only, no agents.
	FormSimple RiteForm = "simple"
	// FormPractitioner represents a rite with agents + mena.
	FormPractitioner RiteForm = "practitioner"
	// FormProcedural represents a rite with hooks + workflows, no dedicated agents.
	FormProcedural RiteForm = "procedural"
	// FormFull represents a rite with all components.
	FormFull RiteForm = "full"
)

// ValidForms contains all valid rite forms.
var ValidForms = []RiteForm{FormSimple, FormPractitioner, FormProcedural, FormFull}

// IsValidForm returns true if the form is valid.
func IsValidForm(f RiteForm) bool {
	return slices.Contains(ValidForms, f)
}

// RiteManifest represents a parsed manifest.yaml file.
// This struct supports both the new actual format used in rites/ directory
// and maintains backward compatibility with the original planned schema.
type RiteManifest struct {
	// Primary fields (actual manifest.yaml format)
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version,omitempty" json:"version,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	EntryAgent  string `yaml:"entry_agent,omitempty" json:"entry_agent,omitempty"`

	// Phases (actual format has phases at top level)
	Phases []ManifestPhase `yaml:"phases,omitempty" json:"phases,omitempty"`

	// Component references - supports both string list and object list
	Agents     []AgentRef `yaml:"agents,omitempty" json:"agents,omitempty"`
	SkillNames []string   `yaml:"-" json:"-"` // Parsed from skills when they're strings
	Skills     []SkillRef `yaml:"-" json:"skills,omitempty"`

	// Dependencies on other rites
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Complexity levels
	ComplexityLevels []ComplexityLevel `yaml:"complexity_levels,omitempty" json:"complexity_levels,omitempty"`

	// Metadata
	Metadata map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Legacy/planned fields for backward compatibility
	SchemaVersion string   `yaml:"schema_version,omitempty" json:"schema_version,omitempty"`
	DisplayName   string   `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Form          RiteForm `yaml:"form,omitempty" json:"form,omitempty"`

	// Optional workflow configuration (legacy format)
	Workflow *WorkflowConfig `yaml:"workflow,omitempty" json:"workflow,omitempty"`

	// Optional lifecycle hooks
	Hooks any `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	// Context budget metadata
	Budget *BudgetInfo `yaml:"budget,omitempty" json:"budget,omitempty"`

	// Path is the directory containing this manifest (set during load, not serialized)
	Path string `yaml:"-" json:"path,omitempty"`
}

// ManifestPhase represents a phase in the actual manifest format.
type ManifestPhase struct {
	Name        string `yaml:"name" json:"name"`
	Agent       string `yaml:"agent" json:"agent"`
	Produces    string `yaml:"produces,omitempty" json:"produces,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Condition   string `yaml:"condition,omitempty" json:"condition,omitempty"`
}

// ComplexityLevel represents a complexity level definition.
type ComplexityLevel struct {
	Name  string `yaml:"name" json:"name"`
	Scope string `yaml:"scope" json:"scope"`
}

// AgentRef references an agent within a rite.
type AgentRef struct {
	Name     string `yaml:"name" json:"name"`
	File     string `yaml:"file,omitempty" json:"file,omitempty"`
	Role     string `yaml:"role,omitempty" json:"role,omitempty"`
	Produces string `yaml:"produces,omitempty" json:"produces,omitempty"`
}

// SkillRef references a skill within a rite.
type SkillRef struct {
	Ref      string `yaml:"ref" json:"ref"`
	Path     string `yaml:"path,omitempty" json:"path,omitempty"`
	External bool   `yaml:"external,omitempty" json:"external,omitempty"`
}

// WorkflowConfig defines workflow configuration for a rite.
type WorkflowConfig struct {
	Type       string        `yaml:"type" json:"type"`
	EntryPoint string        `yaml:"entry_point" json:"entry_point"`
	Phases     []PhaseConfig `yaml:"phases,omitempty" json:"phases,omitempty"`
}

// PhaseConfig defines a workflow phase.
type PhaseConfig struct {
	Name     string `yaml:"name" json:"name"`
	Agent    string `yaml:"agent" json:"agent"`
	Produces string `yaml:"produces,omitempty" json:"produces,omitempty"`
	Next     string `yaml:"next,omitempty" json:"next,omitempty"`
}

// HooksConfig defines lifecycle hooks.
type HooksConfig struct {
	OnInvoke  []HookAction `yaml:"on_invoke,omitempty" json:"on_invoke,omitempty"`
	OnRelease []HookAction `yaml:"on_release,omitempty" json:"on_release,omitempty"`
}

// HookAction defines a hook action.
type HookAction struct {
	Action string `yaml:"action" json:"action"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
}

// BudgetInfo contains context budget metadata.
type BudgetInfo struct {
	EstimatedTokens int `yaml:"estimated_tokens,omitempty" json:"estimated_tokens,omitempty"`
	AgentsCost      int `yaml:"agents_cost,omitempty" json:"agents_cost,omitempty"`
	SkillsCost      int `yaml:"skills_cost,omitempty" json:"skills_cost,omitempty"`
	WorkflowCost    int `yaml:"workflow_cost,omitempty" json:"workflow_cost,omitempty"`
}

// rawManifest is an intermediate struct for parsing manifests with flexible skills field.
type rawManifest struct {
	Name             string            `yaml:"name"`
	Version          string            `yaml:"version,omitempty"`
	Description      string            `yaml:"description,omitempty"`
	EntryAgent       string            `yaml:"entry_agent,omitempty"`
	Phases           []ManifestPhase   `yaml:"phases,omitempty"`
	Agents           []AgentRef        `yaml:"agents,omitempty"`
	Skills           any               `yaml:"skills,omitempty"` // Can be []string or []SkillRef
	Dependencies     []string          `yaml:"dependencies,omitempty"`
	ComplexityLevels []ComplexityLevel `yaml:"complexity_levels,omitempty"`
	Metadata         map[string]any    `yaml:"metadata,omitempty"`
	SchemaVersion    string            `yaml:"schema_version,omitempty"`
	DisplayName      string            `yaml:"display_name,omitempty"`
	Form             RiteForm          `yaml:"form,omitempty"`
	Workflow         *WorkflowConfig   `yaml:"workflow,omitempty"`
	Hooks            any               `yaml:"hooks,omitempty"`
	Budget           *BudgetInfo       `yaml:"budget,omitempty"`
}

// LoadManifest reads and parses a manifest.yaml file.
func LoadManifest(path string) (*RiteManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to read rite manifest", err)
	}

	var raw rawManifest
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.ErrParseError(path, "yaml", err)
	}

	manifest := &RiteManifest{
		Name:             raw.Name,
		Version:          raw.Version,
		Description:      raw.Description,
		EntryAgent:       raw.EntryAgent,
		Phases:           raw.Phases,
		Agents:           raw.Agents,
		Dependencies:     raw.Dependencies,
		ComplexityLevels: raw.ComplexityLevels,
		Metadata:         raw.Metadata,
		SchemaVersion:    raw.SchemaVersion,
		DisplayName:      raw.DisplayName,
		Form:             raw.Form,
		Workflow:         raw.Workflow,
		Hooks:            raw.Hooks,
		Budget:           raw.Budget,
		Path:             filepath.Dir(path),
	}

	// Parse skills - can be []string or []SkillRef
	if raw.Skills != nil {
		if skills, ok := raw.Skills.([]any); ok {
			for _, s := range skills {
				switch skill := s.(type) {
				case string:
					manifest.SkillNames = append(manifest.SkillNames, skill)
				case map[string]any:
					ref := SkillRef{}
					if v, ok := skill["ref"].(string); ok {
						ref.Ref = v
					}
					if v, ok := skill["path"].(string); ok {
						ref.Path = v
					}
					if v, ok := skill["external"].(bool); ok {
						ref.External = v
					}
					manifest.Skills = append(manifest.Skills, ref)
				}
			}
		}
	}

	return manifest, nil
}

// LoadManifestFromDir loads a manifest.yaml from a directory.
func LoadManifestFromDir(dir string) (*RiteManifest, error) {
	manifestPath := filepath.Join(dir, "manifest.yaml")
	return LoadManifest(manifestPath)
}

// Validate checks if the manifest is valid according to schema rules.
// Supports both the actual manifest format and the legacy planned format.
func (m *RiteManifest) Validate() []string {
	var issues []string

	// Required fields - name is always required
	if m.Name == "" {
		issues = append(issues, "name is required")
	} else if !isKebabCase(m.Name) {
		issues = append(issues, "name must be kebab-case")
	}

	// For legacy format, validate schema_version and form
	if m.SchemaVersion != "" || m.Form != "" {
		if m.SchemaVersion == "" {
			issues = append(issues, "schema_version is required (legacy format)")
		}
		if m.Form == "" {
			issues = append(issues, "form is required (legacy format)")
		} else if !IsValidForm(m.Form) {
			issues = append(issues, "form must be one of: simple, practitioner, procedural, full")
		}

		// Form-specific validation (legacy format only)
		switch m.Form {
		case FormSimple:
			if len(m.Agents) > 0 {
				issues = append(issues, "simple form should not have agents")
			}
		case FormPractitioner, FormFull:
			if len(m.Agents) == 0 {
				issues = append(issues, "practitioner and full forms require agents")
			}
		case FormProcedural:
			if len(m.Agents) > 0 {
				issues = append(issues, "procedural form should not have dedicated agents")
			}
		}

		// Validate agent references (legacy format requires file)
		for i, agent := range m.Agents {
			if agent.Name == "" {
				issues = append(issues, fmt.Sprintf("agents[%d].name is required", i))
			}
			if agent.File == "" {
				issues = append(issues, fmt.Sprintf("agents[%d].file is required", i))
			}
		}

		// Validate skill references (legacy format)
		for i, skill := range m.Skills {
			if skill.Ref == "" {
				issues = append(issues, fmt.Sprintf("skills[%d].ref is required", i))
			}
			if skill.External && skill.Path != "" {
				issues = append(issues, fmt.Sprintf("skills[%d]: external skills should not have path", i))
			}
			if !skill.External && skill.Path == "" {
				issues = append(issues, fmt.Sprintf("skills[%d]: local skills require path", i))
			}
		}
	} else {
		// Actual format validation
		// Validate agent references (name only required)
		for i, agent := range m.Agents {
			if agent.Name == "" {
				issues = append(issues, fmt.Sprintf("agents[%d].name is required", i))
			}
		}

		// Validate phases
		for i, phase := range m.Phases {
			if phase.Name == "" {
				issues = append(issues, fmt.Sprintf("phases[%d].name is required", i))
			}
			if phase.Agent == "" {
				issues = append(issues, fmt.Sprintf("phases[%d].agent is required", i))
			}
		}
	}

	// Validate budget fields (if present)
	if m.Budget != nil {
		if m.Budget.EstimatedTokens < 0 {
			issues = append(issues, "budget.estimated_tokens must be non-negative")
		}
		if m.Budget.AgentsCost < 0 {
			issues = append(issues, "budget.agents_cost must be non-negative")
		}
		if m.Budget.SkillsCost < 0 {
			issues = append(issues, "budget.skills_cost must be non-negative")
		}
	}

	return issues
}

// IsValid returns true if the manifest passes validation.
func (m *RiteManifest) IsValid() bool {
	return len(m.Validate()) == 0
}

// AgentNames returns the list of agent names.
func (m *RiteManifest) AgentNames() []string {
	names := make([]string, len(m.Agents))
	for i, a := range m.Agents {
		names[i] = a.Name
	}
	return names
}

// SkillRefs returns the list of skill references.
// Returns SkillNames if skills were parsed as strings, otherwise refs from SkillRef objects.
func (m *RiteManifest) SkillRefs() []string {
	// If skills were parsed as strings, return those
	if len(m.SkillNames) > 0 {
		return m.SkillNames
	}
	// Otherwise, extract refs from SkillRef objects
	refs := make([]string, len(m.Skills))
	for i, s := range m.Skills {
		refs[i] = s.Ref
	}
	return refs
}

// GetAgent returns an agent by name, or nil if not found.
func (m *RiteManifest) GetAgent(name string) *AgentRef {
	for i := range m.Agents {
		if m.Agents[i].Name == name {
			return &m.Agents[i]
		}
	}
	return nil
}

// GetSkill returns a skill by ref, or nil if not found.
func (m *RiteManifest) GetSkill(ref string) *SkillRef {
	for i := range m.Skills {
		if m.Skills[i].Ref == ref {
			return &m.Skills[i]
		}
	}
	return nil
}

// HasAgents returns true if the rite has agents.
func (m *RiteManifest) HasAgents() bool {
	return len(m.Agents) > 0
}

// HasSkills returns true if the rite has skills.
func (m *RiteManifest) HasSkills() bool {
	return len(m.Skills) > 0 || len(m.SkillNames) > 0
}

// HasWorkflow returns true if the rite has workflow configuration.
// Checks both the legacy Workflow field and the new Phases field.
func (m *RiteManifest) HasWorkflow() bool {
	return m.Workflow != nil || len(m.Phases) > 0
}

// GetEstimatedTokens returns the estimated token cost.
func (m *RiteManifest) GetEstimatedTokens() int {
	if m.Budget != nil && m.Budget.EstimatedTokens > 0 {
		return m.Budget.EstimatedTokens
	}
	// Calculate from components if not specified
	total := 0
	if m.Budget != nil {
		total += m.Budget.AgentsCost
		total += m.Budget.SkillsCost
		total += m.Budget.WorkflowCost
	}
	return total
}

// Save writes the manifest to a file.
func (m *RiteManifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal rite manifest", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to write rite manifest", err)
	}
	return nil
}

// isKebabCase validates that a string is kebab-case.
func isKebabCase(s string) bool {
	if s == "" {
		return false
	}
	// Must not start or end with hyphen
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	// Must be lowercase alphanumeric with hyphens
	for _, c := range s {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	// No consecutive hyphens
	return !strings.Contains(s, "--")
}
