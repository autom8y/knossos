package rite

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/resolution"
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
	Source       string   `json:"source"` // "project", "user", "org", or "platform"
}

// Discovery locates available rites via the unified resolution chain.
type Discovery struct {
	chain      *resolution.Chain
	activeRite string
}

// PlatformRitesDirIn returns the platform-level rites directory for the given knossos home.
func PlatformRitesDirIn(knossosHome string) string {
	if knossosHome == "" {
		return ""
	}
	return filepath.Join(knossosHome, "rites")
}

// PlatformRitesDir returns the platform-level rites directory ($KNOSSOS_HOME/rites/).
func PlatformRitesDir() string {
	return PlatformRitesDirIn(config.KnossosHome())
}

// NewDiscovery creates a new rite discovery instance.
func NewDiscovery(resolver *paths.Resolver) *Discovery {
	return &Discovery{
		chain: resolution.RiteChain(
			resolver.RitesDir(),
			paths.UserRitesDir(),
			paths.OrgRitesDir(config.ActiveOrg()),
			PlatformRitesDir(),
			nil, // no embedded FS for Discovery
		),
		activeRite: resolver.ReadActiveRite(),
	}
}

// NewDiscoveryWithPaths creates a discovery with explicit paths.
func NewDiscoveryWithPaths(projectRitesDir, userRitesDir, orgRitesDir, platformRitesDir, activeRite string) *Discovery {
	return &Discovery{
		chain: resolution.RiteChain(
			projectRitesDir,
			userRitesDir,
			orgRitesDir,
			platformRitesDir,
			nil,
		),
		activeRite: activeRite,
	}
}

// riteValidator checks that an entry is a valid rite directory (has manifest.yaml).
func riteValidator(item resolution.ResolvedItem) bool {
	_, err := os.Stat(filepath.Join(item.Path, "manifest.yaml"))
	return err == nil
}

// List returns all available rites.
// Resolution order (highest priority wins): project > user > org > platform.
func (d *Discovery) List() ([]Rite, error) {
	items, err := d.chain.ResolveAll(riteValidator)
	if err != nil {
		return nil, err
	}

	rites := make([]Rite, 0, len(items))
	for _, item := range items {
		r, err := loadRite(item.Path, item.Source)
		if err != nil {
			continue // skip invalid rites
		}
		rites = append(rites, *r)
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

// GetRitePath returns the path to a rite directory.
// Uses the resolution chain for full 4-tier lookup.
func (d *Discovery) GetRitePath(name string) (string, error) {
	item, err := d.chain.Resolve(name, riteValidator)
	if err != nil {
		return "", errors.ErrRiteNotFound(name)
	}
	return item.Path, nil
}

// loadRite loads a rite from a directory.
func loadRite(ritePath, source string) (*Rite, error) {
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
