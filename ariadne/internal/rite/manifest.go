// Package rite implements rite manifest handling, discovery, and invocation for Ariadne.
// Rites are composable practice bundles that can be invoked (additive) or swapped (replacement).
package rite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/ariadne/internal/errors"
)

// RiteForm classifies a rite by its composition of components.
type RiteForm string

const (
	// FormSimple represents a rite with skills only, no agents.
	FormSimple RiteForm = "simple"
	// FormPractitioner represents a rite with agents + skills.
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
	for _, valid := range ValidForms {
		if f == valid {
			return true
		}
	}
	return false
}

// RiteManifest represents a parsed rite.yaml file.
type RiteManifest struct {
	SchemaVersion string `yaml:"schema_version" json:"schema_version"`
	Name          string `yaml:"name" json:"name"`
	DisplayName   string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Description   string `yaml:"description,omitempty" json:"description,omitempty"`
	Form          RiteForm `yaml:"form" json:"form"`

	// Component references
	Agents []AgentRef `yaml:"agents,omitempty" json:"agents,omitempty"`
	Skills []SkillRef `yaml:"skills,omitempty" json:"skills,omitempty"`

	// Optional workflow configuration
	Workflow *WorkflowConfig `yaml:"workflow,omitempty" json:"workflow,omitempty"`

	// Optional lifecycle hooks
	Hooks *HooksConfig `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	// Context budget metadata
	Budget *BudgetInfo `yaml:"budget,omitempty" json:"budget,omitempty"`

	// Migration metadata (for transition from legacy naming)
	Migration *MigrationInfo `yaml:"migration,omitempty" json:"migration,omitempty"`

	// Path is the directory containing this manifest (set during load, not serialized)
	Path string `yaml:"-" json:"path,omitempty"`
}

// AgentRef references an agent within a rite.
type AgentRef struct {
	Name     string `yaml:"name" json:"name"`
	File     string `yaml:"file" json:"file"`
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

// MigrationInfo contains migration metadata from legacy naming.
type MigrationInfo struct {
	FromRite   string `yaml:"from_rite,omitempty" json:"from_rite,omitempty"`
	MigratedAt string `yaml:"migrated_at,omitempty" json:"migrated_at,omitempty"`
}

// LoadManifest reads and parses a rite.yaml file.
func LoadManifest(path string) (*RiteManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to read rite manifest", err)
	}

	var manifest RiteManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, errors.ErrParseError(path, "yaml", err)
	}

	// Set the path to the containing directory
	manifest.Path = filepath.Dir(path)

	return &manifest, nil
}

// LoadManifestFromDir loads a rite.yaml from a directory.
func LoadManifestFromDir(dir string) (*RiteManifest, error) {
	manifestPath := filepath.Join(dir, "rite.yaml")
	return LoadManifest(manifestPath)
}

// Validate checks if the manifest is valid according to schema rules.
func (m *RiteManifest) Validate() []string {
	var issues []string

	// Required fields
	if m.SchemaVersion == "" {
		issues = append(issues, "schema_version is required")
	}
	if m.Name == "" {
		issues = append(issues, "name is required")
	} else if !isKebabCase(m.Name) {
		issues = append(issues, "name must be kebab-case")
	}
	if m.Form == "" {
		issues = append(issues, "form is required")
	} else if !IsValidForm(m.Form) {
		issues = append(issues, "form must be one of: simple, practitioner, procedural, full")
	}

	// Form-specific validation
	switch m.Form {
	case FormSimple:
		// Simple forms should not have agents
		if len(m.Agents) > 0 {
			issues = append(issues, "simple form should not have agents")
		}
	case FormPractitioner, FormFull:
		// Practitioner and full forms require agents
		if len(m.Agents) == 0 {
			issues = append(issues, "practitioner and full forms require agents")
		}
	case FormProcedural:
		// Procedural forms should not have dedicated agents
		if len(m.Agents) > 0 {
			issues = append(issues, "procedural form should not have dedicated agents")
		}
	}

	// Validate agent references
	for i, agent := range m.Agents {
		if agent.Name == "" {
			issues = append(issues, fmt.Sprintf("agents[%d].name is required", i))
		}
		if agent.File == "" {
			issues = append(issues, fmt.Sprintf("agents[%d].file is required", i))
		}
	}

	// Validate skill references
	for i, skill := range m.Skills {
		if skill.Ref == "" {
			issues = append(issues, fmt.Sprintf("skills[%d].ref is required", i))
		}
		// External skills should not have path
		if skill.External && skill.Path != "" {
			issues = append(issues, fmt.Sprintf("skills[%d]: external skills should not have path", i))
		}
		// Non-external skills should have path
		if !skill.External && skill.Path == "" {
			issues = append(issues, fmt.Sprintf("skills[%d]: local skills require path", i))
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
func (m *RiteManifest) SkillRefs() []string {
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
	return len(m.Skills) > 0
}

// HasWorkflow returns true if the rite has workflow configuration.
func (m *RiteManifest) HasWorkflow() bool {
	return m.Workflow != nil
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
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	// No consecutive hyphens
	return !strings.Contains(s, "--")
}
