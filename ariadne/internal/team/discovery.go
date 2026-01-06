package team

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// Team represents a discovered team pack.
type Team struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	Agents       []string `json:"agents"`
	AgentCount   int      `json:"agent_count"`
	WorkflowType string   `json:"workflow_type"`
	EntryPoint   string   `json:"entry_point"`
	Active       bool     `json:"active"`
}

// Discovery locates available team packs.
type Discovery struct {
	projectTeamsDir string
	userTeamsDir    string
	activeTeam      string
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
		d.activeTeam = strings.TrimSpace(string(data))
	}

	return d
}

// NewDiscoveryWithPaths creates a discovery with explicit paths.
func NewDiscoveryWithPaths(projectTeamsDir, userTeamsDir, activeTeam string) *Discovery {
	return &Discovery{
		projectTeamsDir: projectTeamsDir,
		userTeamsDir:    userTeamsDir,
		activeTeam:      activeTeam,
	}
}

// List returns all available teams.
func (d *Discovery) List() ([]Team, error) {
	var teams []Team

	// Scan project teams/
	if projectTeams, err := d.scanDir(d.projectTeamsDir); err == nil {
		teams = append(teams, projectTeams...)
	}

	// Scan user teams if present
	if d.userTeamsDir != "" {
		if userTeams, err := d.scanDir(d.userTeamsDir); err == nil {
			// User teams override project teams with same name
			teamMap := make(map[string]Team)
			for _, t := range teams {
				teamMap[t.Name] = t
			}
			for _, t := range userTeams {
				teamMap[t.Name] = t
			}
			teams = make([]Team, 0, len(teamMap))
			for _, t := range teamMap {
				teams = append(teams, t)
			}
		}
	}

	// Sort by name
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Name < teams[j].Name
	})

	// Mark active team
	for i := range teams {
		teams[i].Active = teams[i].Name == d.activeTeam
	}

	return teams, nil
}

// Get returns a specific team by name.
func (d *Discovery) Get(name string) (*Team, error) {
	teams, err := d.List()
	if err != nil {
		return nil, err
	}

	for _, t := range teams {
		if t.Name == name {
			return &t, nil
		}
	}

	return nil, errors.ErrTeamNotFound(name)
}

// GetActive returns the currently active team.
func (d *Discovery) GetActive() (*Team, error) {
	if d.activeTeam == "" {
		return nil, errors.New(errors.CodeFileNotFound, "No active team set")
	}
	return d.Get(d.activeTeam)
}

// ActiveTeamName returns the name of the active team.
func (d *Discovery) ActiveTeamName() string {
	return d.activeTeam
}

// Exists checks if a team exists.
func (d *Discovery) Exists(name string) bool {
	_, err := d.Get(name)
	return err == nil
}

// scanDir scans a directory for team packs.
func (d *Discovery) scanDir(dir string) ([]Team, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var teams []Team
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		teamPath := filepath.Join(dir, entry.Name())
		team, err := d.loadTeam(teamPath)
		if err != nil {
			// Skip invalid teams (missing workflow.yaml, etc.)
			continue
		}

		teams = append(teams, *team)
	}

	return teams, nil
}

// loadTeam loads a team from a directory.
func (d *Discovery) loadTeam(teamPath string) (*Team, error) {
	workflowPath := filepath.Join(teamPath, "workflow.yaml")
	workflow, err := LoadWorkflow(workflowPath)
	if err != nil {
		return nil, err
	}

	// Count agents in agents/ directory
	agentsDir := filepath.Join(teamPath, "agents")
	agents, err := listAgentFiles(agentsDir)
	if err != nil {
		agents = []string{}
	}

	// Extract agent names without extension
	agentNames := make([]string, len(agents))
	for i, a := range agents {
		agentNames[i] = strings.TrimSuffix(a, ".md")
	}

	team := &Team{
		Name:         workflow.Name,
		Path:         teamPath,
		Description:  workflow.Description,
		Agents:       agentNames,
		AgentCount:   len(agents),
		WorkflowType: workflow.WorkflowType,
		EntryPoint:   workflow.EntryPoint.Agent,
	}

	// If workflow name is empty, use directory name
	if team.Name == "" {
		team.Name = filepath.Base(teamPath)
	}

	return team, nil
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
