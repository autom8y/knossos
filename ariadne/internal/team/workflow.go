// Package team implements team pack discovery, management, and switching for Ariadne.
package team

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Workflow represents a parsed workflow.yaml file.
type Workflow struct {
	Name         string       `yaml:"name" json:"name"`
	WorkflowType string       `yaml:"workflow_type" json:"workflow_type"`
	Description  string       `yaml:"description" json:"description"`
	EntryPoint   EntryPoint   `yaml:"entry_point" json:"entry_point"`
	Phases       []Phase      `yaml:"phases" json:"phases"`
}

// EntryPoint defines the workflow entry point configuration.
type EntryPoint struct {
	Agent string `yaml:"agent" json:"agent"`
}

// Phase represents a workflow phase.
type Phase struct {
	Name      string  `yaml:"name" json:"name"`
	Agent     string  `yaml:"agent" json:"agent"`
	Produces  string  `yaml:"produces" json:"produces"`
	Next      *string `yaml:"next" json:"next,omitempty"`
	Condition string  `yaml:"condition,omitempty" json:"condition,omitempty"`
}

// LoadWorkflow reads and parses a workflow.yaml file.
func LoadWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// PhaseNames returns the list of phase names.
func (w *Workflow) PhaseNames() []string {
	names := make([]string, len(w.Phases))
	for i, p := range w.Phases {
		names[i] = p.Name
	}
	return names
}

// AgentNames returns the unique agent names from phases.
func (w *Workflow) AgentNames() []string {
	seen := make(map[string]bool)
	var agents []string
	for _, p := range w.Phases {
		if !seen[p.Agent] {
			seen[p.Agent] = true
			agents = append(agents, p.Agent)
		}
	}
	return agents
}

// GetPhase returns a phase by name.
func (w *Workflow) GetPhase(name string) *Phase {
	for i := range w.Phases {
		if w.Phases[i].Name == name {
			return &w.Phases[i]
		}
	}
	return nil
}

// AgentInfo holds agent metadata extracted from workflow.
type AgentInfo struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Role     string `json:"role"`
	Produces string `json:"produces"`
}

// GetAgentInfo returns agent info from phases with role/produces mapping.
func (w *Workflow) GetAgentInfo() []AgentInfo {
	var infos []AgentInfo
	seen := make(map[string]bool)

	for _, p := range w.Phases {
		if seen[p.Agent] {
			continue
		}
		seen[p.Agent] = true

		info := AgentInfo{
			Name:     p.Agent,
			File:     p.Agent + ".md",
			Produces: p.Produces,
		}

		// Derive role from agent name (will be overridden from agent file if available)
		info.Role = deriveRoleFromName(p.Agent)

		infos = append(infos, info)
	}

	return infos
}

// deriveRoleFromName provides a default role description based on agent name.
func deriveRoleFromName(name string) string {
	roleMap := map[string]string{
		"architect":            "Evaluates tradeoffs and designs systems",
		"orchestrator":         "Coordinates development lifecycle",
		"principal-engineer":   "Transforms designs into production code",
		"qa-adversary":         "Breaks things so users don't",
		"requirements-analyst": "Extracts stakeholder needs",
	}

	if role, ok := roleMap[name]; ok {
		return role
	}

	// Convert kebab-case to title case
	words := strings.Split(name, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(string(w[0])) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
