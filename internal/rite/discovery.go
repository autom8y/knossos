package rite

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// Rite represents a discovered rite.
type Rite struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name,omitempty"`
	Description  string   `json:"description,omitempty"`
	Form         RiteForm `json:"form"`
	Path         string   `json:"path"`
	Agents       []string `json:"agents,omitempty"`
	AgentCount   int      `json:"agent_count"`
	Skills       []string `json:"skills,omitempty"`
	SkillCount   int      `json:"skill_count"`
	HasWorkflow  bool     `json:"has_workflow"`
	WorkflowType string   `json:"workflow_type,omitempty"`
	EntryPoint   string   `json:"entry_point,omitempty"`
	Active       bool     `json:"active"`
	Source       string   `json:"source"` // "project" or "user"
}

// Discovery locates available rites.
type Discovery struct {
	projectRitesDir string
	userRitesDir    string
	orgRitesDir     string
	activeRite      string
}

// NewDiscovery creates a new rite discovery instance.
func NewDiscovery(resolver *paths.Resolver) *Discovery {
	return &Discovery{
		projectRitesDir: resolver.RitesDir(),
		userRitesDir:    paths.UserRitesDir(),
		orgRitesDir:     paths.OrgRitesDir(config.ActiveOrg()),
		activeRite:      resolver.ReadActiveRite(),
	}
}

// NewDiscoveryWithPaths creates a discovery with explicit paths.
func NewDiscoveryWithPaths(projectRitesDir, userRitesDir, activeRite string) *Discovery {
	return &Discovery{
		projectRitesDir: projectRitesDir,
		userRitesDir:    userRitesDir,
		activeRite:      activeRite,
	}
}

// List returns all available rites.
func (d *Discovery) List() ([]Rite, error) {
	var rites []Rite

	// Scan project rites
	if projectRites, err := d.scanDir(d.projectRitesDir, "project"); err == nil {
		rites = append(rites, projectRites...)
	}

	// Scan org rites if present (org < user < project in precedence)
	if d.orgRitesDir != "" {
		if orgRites, err := d.scanDir(d.orgRitesDir, "org"); err == nil {
			riteMap := make(map[string]Rite)
			for _, r := range rites {
				riteMap[r.Name] = r
			}
			for _, r := range orgRites {
				if _, exists := riteMap[r.Name]; !exists {
					riteMap[r.Name] = r
				}
			}
			rites = make([]Rite, 0, len(riteMap))
			for _, r := range riteMap {
				rites = append(rites, r)
			}
		}
	}

	// Scan user rites if present (user > org, user < project)
	if d.userRitesDir != "" {
		if userRites, err := d.scanDir(d.userRitesDir, "user"); err == nil {
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

// ListByForm returns rites filtered by form.
func (d *Discovery) ListByForm(form RiteForm) ([]Rite, error) {
	all, err := d.List()
	if err != nil {
		return nil, err
	}

	var filtered []Rite
	for _, r := range all {
		if r.Form == form {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// ListBySource returns rites filtered by source (project or user).
func (d *Discovery) ListBySource(source string) ([]Rite, error) {
	all, err := d.List()
	if err != nil {
		return nil, err
	}

	var filtered []Rite
	for _, r := range all {
		if r.Source == source {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
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

// GetManifest returns the full manifest for a rite.
func (d *Discovery) GetManifest(name string) (*RiteManifest, error) {
	rite, err := d.Get(name)
	if err != nil {
		return nil, err
	}

	manifest, err := LoadManifestFromDir(rite.Path)
	if err != nil {
		return nil, err
	}

	// Derive form if not explicitly set (consistent with loadRite)
	if manifest.Form == "" {
		manifest.Form = deriveForm(manifest)
	}

	return manifest, nil
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
func (d *Discovery) scanDir(dir, source string) ([]Rite, error) {
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
		rite, err := d.loadRite(ritePath, source)
		if err != nil {
			// Skip invalid rites (missing rite.yaml, etc.)
			continue
		}

		rites = append(rites, *rite)
	}

	return rites, nil
}

// loadRite loads a rite from a directory.
func (d *Discovery) loadRite(ritePath, source string) (*Rite, error) {
	manifest, err := LoadManifestFromDir(ritePath)
	if err != nil {
		return nil, err
	}

	// Derive form from manifest structure if not explicitly set
	form := manifest.Form
	if form == "" {
		form = deriveForm(manifest)
	}

	// Calculate skill count from both formats
	skillCount := len(manifest.Skills)
	if len(manifest.SkillNames) > 0 {
		skillCount = len(manifest.SkillNames)
	}

	rite := &Rite{
		Name:        manifest.Name,
		DisplayName: manifest.DisplayName,
		Description: manifest.Description,
		Form:        form,
		Path:        ritePath,
		Agents:      manifest.AgentNames(),
		AgentCount:  len(manifest.Agents),
		Skills:      manifest.SkillRefs(),
		SkillCount:  skillCount,
		HasWorkflow: manifest.HasWorkflow(),
		Source:      source,
	}

	// If name is empty, use directory name
	if rite.Name == "" {
		rite.Name = filepath.Base(ritePath)
	}

	// Set workflow type from phases if available
	if len(manifest.Phases) > 0 {
		rite.WorkflowType = "sequential"
	} else if manifest.Workflow != nil {
		rite.WorkflowType = manifest.Workflow.Type
	}

	// Set entry point
	if manifest.EntryAgent != "" {
		rite.EntryPoint = manifest.EntryAgent
	} else if manifest.Workflow != nil {
		rite.EntryPoint = manifest.Workflow.EntryPoint
	}

	return rite, nil
}

// deriveForm determines the rite form from its components.
func deriveForm(m *RiteManifest) RiteForm {
	hasAgents := len(m.Agents) > 0
	hasSkills := len(m.Skills) > 0 || len(m.SkillNames) > 0
	hasWorkflow := m.HasWorkflow()

	if hasAgents && hasSkills && hasWorkflow {
		return FormFull
	}
	if hasAgents && hasSkills {
		return FormPractitioner
	}
	if hasWorkflow && !hasAgents {
		return FormProcedural
	}
	if hasSkills && !hasAgents {
		return FormSimple
	}
	// Default to practitioner for rites with agents
	if hasAgents {
		return FormPractitioner
	}
	return FormSimple
}

// GetRitePath returns the path to a rite directory.
// Returns the first match from project or user directories.
func (d *Discovery) GetRitePath(name string) (string, error) {
	// Check project rites first
	projectPath := filepath.Join(d.projectRitesDir, name)
	if _, err := os.Stat(filepath.Join(projectPath, "manifest.yaml")); err == nil {
		return projectPath, nil
	}

	// Check user rites
	if d.userRitesDir != "" {
		userPath := filepath.Join(d.userRitesDir, name)
		if _, err := os.Stat(filepath.Join(userPath, "manifest.yaml")); err == nil {
			return userPath, nil
		}
	}

	// Check org rites
	if d.orgRitesDir != "" {
		orgPath := filepath.Join(d.orgRitesDir, name)
		if _, err := os.Stat(filepath.Join(orgPath, "manifest.yaml")); err == nil {
			return orgPath, nil
		}
	}

	return "", errors.ErrRiteNotFound(name)
}
