package team

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// Rite represents a discovered rite (practice bundle).
type Rite struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	Agents       []string `json:"agents"`
	AgentCount   int      `json:"agent_count"`
	WorkflowType string   `json:"workflow_type"`
	EntryPoint   string   `json:"entry_point"`
	Active       bool     `json:"active"`
}

// Discovery locates available rites (practice bundles).
type Discovery struct {
	projectTeamsDir string
	userTeamsDir    string
	activeRite      string
}

// NewDiscovery creates a new team discovery instance.
func NewDiscovery(resolver *paths.Resolver) *Discovery {
	d := &Discovery{
		projectTeamsDir: resolver.RitesDir(),
		userTeamsDir:    paths.UserRitesDir(),
	}

	// Read active rite
	ritePath := resolver.ActiveRiteFile()
	if data, err := os.ReadFile(ritePath); err == nil {
		d.activeRite = strings.TrimSpace(string(data))
	}

	return d
}

// NewDiscoveryWithPaths creates a discovery with explicit paths.
func NewDiscoveryWithPaths(projectTeamsDir, userTeamsDir, activeRite string) *Discovery {
	return &Discovery{
		projectTeamsDir: projectTeamsDir,
		userTeamsDir:    userTeamsDir,
		activeRite:      activeRite,
	}
}

// List returns all available rites.
func (d *Discovery) List() ([]Rite, error) {
	var rites []Rite

	// Scan project rites
	if projectRites, err := d.scanDir(d.projectTeamsDir); err == nil {
		rites = append(rites, projectRites...)
	}

	// Scan user rites if present
	if d.userTeamsDir != "" {
		if userRites, err := d.scanDir(d.userTeamsDir); err == nil {
			// User rites override project rites with same name
			riteMap := make(map[string]Rite)
			for _, r := range rites {
				riteMap[r.Name] = r
			}
			for _, r := range userRites {
				riteMap[r.Name] = r
			}
			rites = make([]Rite, 0, len(riteMap))
			for _, r := range riteMap {
				rites = append(rites, r)
			}
		}
	}

	// Sort by name
	sort.Slice(rites, func(i, j int) bool {
		return rites[i].Name < rites[j].Name
	})

	// Mark active rite
	for i := range rites {
		rites[i].Active = rites[i].Name == d.activeRite
	}

	return rites, nil
}

// Get returns a specific rite by name.
func (d *Discovery) Get(name string) (*Rite, error) {
	rites, err := d.List()
	if err != nil {
		return nil, err
	}

	for _, r := range rites {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, errors.ErrRiteNotFound(name)
}

// GetActive returns the currently active rite.
func (d *Discovery) GetActive() (*Rite, error) {
	if d.activeRite == "" {
		return nil, errors.New(errors.CodeFileNotFound, "No active rite set")
	}
	return d.Get(d.activeRite)
}

// ActiveRiteName returns the name of the active rite.
func (d *Discovery) ActiveRiteName() string {
	return d.activeRite
}

// Exists checks if a rite exists.
func (d *Discovery) Exists(name string) bool {
	_, err := d.Get(name)
	return err == nil
}

// scanDir scans a directory for rites.
func (d *Discovery) scanDir(dir string) ([]Rite, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var rites []Rite
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ritePath := filepath.Join(dir, entry.Name())
		rite, err := d.loadTeam(ritePath)
		if err != nil {
			// Skip invalid rites (missing workflow.yaml, etc.)
			continue
		}

		rites = append(rites, *rite)
	}

	return rites, nil
}

// loadTeam loads a rite from a directory.
func (d *Discovery) loadTeam(ritePath string) (*Rite, error) {
	workflowPath := filepath.Join(ritePath, "workflow.yaml")
	workflow, err := LoadWorkflow(workflowPath)
	if err != nil {
		return nil, err
	}

	// Count agents in agents/ directory
	agentsDir := filepath.Join(ritePath, "agents")
	agents, err := listAgentFiles(agentsDir)
	if err != nil {
		agents = []string{}
	}

	// Extract agent names without extension
	agentNames := make([]string, len(agents))
	for i, a := range agents {
		agentNames[i] = strings.TrimSuffix(a, ".md")
	}

	rite := &Rite{
		Name:         workflow.Name,
		Path:         ritePath,
		Description:  workflow.Description,
		Agents:       agentNames,
		AgentCount:   len(agents),
		WorkflowType: workflow.WorkflowType,
		EntryPoint:   workflow.EntryPoint.Agent,
	}

	// If workflow name is empty, use directory name
	if rite.Name == "" {
		rite.Name = filepath.Base(ritePath)
	}

	return rite, nil
}

// listAgentFiles returns the list of .md files in the agents directory.
func listAgentFiles(agentsDir string) ([]string, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	var agents []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			agents = append(agents, entry.Name())
		}
	}

	sort.Strings(agents)
	return agents, nil
}
