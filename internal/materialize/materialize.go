// Package materialize generates .claude/ directories from templates and rite manifests.
package materialize

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/inscription"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sync"
)

// RiteManifest represents a rite manifest.yaml file.
type RiteManifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	EntryAgent  string   `yaml:"entry_agent"`
	Agents      []Agent  `yaml:"agents"`
	Skills      []string `yaml:"skills"`
	Hooks       []string `yaml:"hooks"`
	Dependencies []string `yaml:"dependencies"`
}

// Agent represents an agent definition in a rite manifest.
type Agent struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
}

// Materializer handles .claude/ directory generation.
type Materializer struct {
	resolver *paths.Resolver
	ritesDir string
	templatesDir string
}

// NewMaterializer creates a new materializer.
func NewMaterializer(resolver *paths.Resolver) *Materializer {
	projectRoot := resolver.ProjectRoot()
	return &Materializer{
		resolver:     resolver,
		ritesDir:     filepath.Join(projectRoot, "rites"),
		templatesDir: filepath.Join(projectRoot, "templates"),
	}
}

// Materialize generates the .claude/ directory from templates and the active rite.
func (m *Materializer) Materialize(activeRiteName string) error {
	// 1. Load the active rite manifest
	ritePath := filepath.Join(m.ritesDir, activeRiteName)
	manifest, err := m.loadRiteManifest(ritePath)
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, "failed to load rite manifest", err)
	}

	claudeDir := m.resolver.ClaudeDir()

	// 2. Ensure .claude/ directory exists
	if err := paths.EnsureDir(claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create .claude directory", err)
	}

	// 3. Generate agents/ directory from rite
	if err := m.materializeAgents(manifest, ritePath, claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to materialize agents", err)
	}

	// 4. Generate skills/ directory from rite + shared + dependencies
	if err := m.materializeSkills(manifest, claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to materialize skills", err)
	}

	// 5. Generate hooks/ directory from templates/hooks
	if err := m.materializeHooks(claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to materialize hooks", err)
	}

	// 6. Generate CLAUDE.md from inscription system
	if err := m.materializeCLAUDEmd(manifest, claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}

	// 7. Generate settings.local.json if not exists
	if err := m.materializeSettings(claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
	}

	// 8. Track state in .claude/sync/state.json
	if err := m.trackState(manifest, activeRiteName); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to track state", err)
	}

	// 9. Write ACTIVE_RITE marker
	if err := m.writeActiveRite(activeRiteName, claudeDir); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write ACTIVE_RITE", err)
	}

	return nil
}

// loadRiteManifest loads a rite's manifest.yaml file.
func (m *Materializer) loadRiteManifest(ritePath string) (*RiteManifest, error) {
	manifestPath := filepath.Join(ritePath, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest RiteManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// materializeAgents copies agent files from rite to .claude/agents/
func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, claudeDir string) error {
	agentsDir := filepath.Join(claudeDir, "agents")

	// Remove existing agents directory
	if err := os.RemoveAll(agentsDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create fresh agents directory
	if err := paths.EnsureDir(agentsDir); err != nil {
		return err
	}

	// Copy agent files
	sourceAgentsDir := filepath.Join(ritePath, "agents")
	if _, err := os.Stat(sourceAgentsDir); os.IsNotExist(err) {
		return nil // No agents in this rite
	}

	return filepath.WalkDir(sourceAgentsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Read source file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Compute relative path
		relPath, err := filepath.Rel(sourceAgentsDir, path)
		if err != nil {
			return err
		}

		// Write to destination
		destPath := filepath.Join(agentsDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(destPath, content, 0644)
	})
}

// materializeSkills copies skill files from rite, shared, and dependencies to .claude/skills/
func (m *Materializer) materializeSkills(manifest *RiteManifest, claudeDir string) error {
	skillsDir := filepath.Join(claudeDir, "skills")

	// Remove existing skills directory
	if err := os.RemoveAll(skillsDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create fresh skills directory
	if err := paths.EnsureDir(skillsDir); err != nil {
		return err
	}

	// Collect all skill sources: current rite + dependencies + shared
	sources := []string{filepath.Join(m.ritesDir, manifest.Name, "skills")}

	// Add dependency rites
	for _, dep := range manifest.Dependencies {
		sources = append(sources, filepath.Join(m.ritesDir, dep, "skills"))
	}

	// Copy skills from all sources
	for _, source := range sources {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue // Skip if source doesn't exist
		}

		if err := m.copyDir(source, skillsDir); err != nil {
			return err
		}
	}

	return nil
}

// materializeHooks copies hook files from templates/hooks to .claude/hooks/
func (m *Materializer) materializeHooks(claudeDir string) error {
	hooksDir := filepath.Join(claudeDir, "hooks")
	sourceHooksDir := filepath.Join(m.templatesDir, "hooks")

	// Check if templates/hooks exists (it may not yet)
	if _, err := os.Stat(sourceHooksDir); os.IsNotExist(err) {
		// For now, just copy the existing hooks from .claude/hooks if they exist
		// This preserves existing functionality until we have templates/hooks
		return nil
	}

	// Remove existing hooks directory
	if err := os.RemoveAll(hooksDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create fresh hooks directory
	if err := paths.EnsureDir(hooksDir); err != nil {
		return err
	}

	// Copy hooks
	return m.copyDir(sourceHooksDir, hooksDir)
}

// materializeCLAUDEmd generates CLAUDE.md using the inscription system.
func (m *Materializer) materializeCLAUDEmd(manifest *RiteManifest, claudeDir string) error {
	// Load KNOSSOS_MANIFEST.yaml
	knossosManifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	loader := inscription.NewManifestLoader(m.resolver.ProjectRoot())
	loader.ManifestPath = knossosManifestPath
	inscriptionManifest, err := loader.Load()
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, "failed to load KNOSSOS_MANIFEST.yaml", err)
	}

	// Build render context
	agents := make([]inscription.AgentInfo, 0, len(manifest.Agents))
	for _, agent := range manifest.Agents {
		agents = append(agents, inscription.AgentInfo{
			Name:     agent.Name,
			File:     agent.Name + ".md",
			Role:     agent.Role,
			Produces: "", // Not in minimal manifest
		})
	}

	renderCtx := &inscription.RenderContext{
		ActiveRite:  manifest.Name,
		AgentCount:  len(manifest.Agents),
		Agents:      agents,
		KnossosVars: make(map[string]string),
		ProjectRoot: m.resolver.ProjectRoot(),
	}

	// Create generator
	generator := inscription.NewGenerator("", inscriptionManifest, renderCtx)

	// Generate all sections
	sections, err := generator.GenerateAll()
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to generate CLAUDE.md sections", err)
	}

	// Read existing CLAUDE.md if it exists
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	existingContent := ""
	if data, err := os.ReadFile(claudeMdPath); err == nil {
		existingContent = string(data)
	}

	// Merge sections
	merger := inscription.NewMerger(inscriptionManifest, generator)
	mergeResult, err := merger.MergeRegions(existingContent, sections)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to merge CLAUDE.md regions", err)
	}

	// Write CLAUDE.md
	return os.WriteFile(claudeMdPath, []byte(mergeResult.Content), 0644)
}

// materializeSettings generates settings.local.json if it doesn't exist.
func (m *Materializer) materializeSettings(claudeDir string) error {
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Don't overwrite existing settings
	if _, err := os.Stat(settingsPath); err == nil {
		return nil
	}

	// Create minimal settings with hooks configuration
	settings := map[string]interface{}{
		"hooks": map[string]interface{}{},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

// trackState updates .claude/sync/state.json with materialization metadata.
func (m *Materializer) trackState(manifest *RiteManifest, activeRiteName string) error {
	stateManager := sync.NewStateManager(m.resolver)

	// Load or initialize state
	state, err := stateManager.Load()
	if err != nil {
		return err
	}

	if state == nil {
		// Initialize new state
		state, err = stateManager.Initialize(fmt.Sprintf("local:%s", activeRiteName))
		if err != nil {
			return err
		}
	}

	// Update active rite in state (we'll need to extend the State struct for this)
	// For now, just save the state
	return stateManager.Save(state)
}

// writeActiveRite writes the ACTIVE_RITE marker file.
func (m *Materializer) writeActiveRite(riteName, claudeDir string) error {
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	return os.WriteFile(activeRitePath, []byte(riteName+"\n"), 0644)
}

// copyDir recursively copies a directory.
func (m *Materializer) copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read and write file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, content, 0644)
	})
}
