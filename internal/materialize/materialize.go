// Package materialize generates .claude/ directories from templates and rite manifests.
package materialize

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/inscription"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sync"
)

// Options configures materialization behavior.
type Options struct {
	Force      bool // Skip idempotency check; proceed even if already on this rite
	DryRun     bool // Preview changes without applying
	RemoveAll  bool // Remove all orphan agents (with backup)
	KeepAll    bool // Preserve all orphan agents (default)
	PromoteAll bool // Move orphan agents to user-level
	Minimal    bool // Generate base infrastructure only (no rite/agents/skills)
}

// Result contains materialization outcome details.
type Result struct {
	OrphansDetected  []string // List of orphan agent files detected
	OrphanAction     string   // Action taken: "kept", "removed", "promoted"
	BackupPath       string   // Path to backup if orphans were removed
	HooksSkipped     bool     // True if hooks were skipped (no templates/hooks dir)
	Source           string   // Source type used: "project", "user", "knossos", "explicit"
	SourcePath       string   // Actual path resolved for rite source
	LegacyBackupPath string   // Path to legacy CLAUDE.md backup if migration occurred
}

// MCPServer represents an MCP server declaration in a rite manifest.
type MCPServer struct {
	Name    string            `yaml:"name" json:"name"`
	Command string            `yaml:"command" json:"command"`
	Args    []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

// RiteManifest represents a rite manifest.yaml file.
type RiteManifest struct {
	Name         string      `yaml:"name"`
	Version      string      `yaml:"version"`
	Description  string      `yaml:"description"`
	EntryAgent   string      `yaml:"entry_agent"`
	Agents       []Agent     `yaml:"agents"`
	Dromena      []string    `yaml:"dromena"`                // Invokable commands (project to .claude/commands/)
	Legomena     []string    `yaml:"legomena"`               // Reference knowledge (project to .claude/skills/)
	Commands     []string    `yaml:"commands"`               // Backward compat: populates from dromena+legomena if empty
	Skills       []string    `yaml:"skills"`                 // Deprecated: use Legomena instead
	Hooks        []string    `yaml:"hooks"`
	Dependencies []string    `yaml:"dependencies"`
	MCPServers   []MCPServer `yaml:"mcp_servers,omitempty"` // MCP server declarations
}

// MCPServerNames returns the list of MCP server names declared in the manifest.
func (m *RiteManifest) MCPServerNames() []string {
	names := make([]string, len(m.MCPServers))
	for i, server := range m.MCPServers {
		names[i] = server.Name
	}
	return names
}

// Agent represents an agent definition in a rite manifest.
type Agent struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
}

// Materializer handles .claude/ directory generation.
type Materializer struct {
	resolver       *paths.Resolver
	sourceResolver *SourceResolver
	explicitSource string // Optional explicit source from --source flag
	ritesDir       string // Deprecated: use sourceResolver
	templatesDir   string
}

// NewMaterializer creates a new materializer with default source resolution.
// Uses 4-tier resolution: project > user > knossos.
func NewMaterializer(resolver *paths.Resolver) *Materializer {
	projectRoot := resolver.ProjectRoot()
	return &Materializer{
		resolver:       resolver,
		sourceResolver: NewSourceResolver(projectRoot),
		ritesDir:       filepath.Join(projectRoot, "rites"),
		templatesDir:   filepath.Join(projectRoot, "templates"),
	}
}

// NewMaterializerWithSource creates a materializer with an explicit source path.
// The source can be a path (absolute or ~-relative) or "knossos" to use $KNOSSOS_HOME.
func NewMaterializerWithSource(resolver *paths.Resolver, source string) *Materializer {
	projectRoot := resolver.ProjectRoot()
	return &Materializer{
		resolver:       resolver,
		sourceResolver: NewSourceResolver(projectRoot),
		explicitSource: source,
		ritesDir:       filepath.Join(projectRoot, "rites"),
		templatesDir:   filepath.Join(projectRoot, "templates"),
	}
}

// Materialize generates the .claude/ directory from templates and the active rite.
// This is the legacy method that uses default options (keep orphans).
func (m *Materializer) Materialize(activeRiteName string) error {
	_, err := m.MaterializeWithOptions(activeRiteName, Options{KeepAll: true})
	return err
}

// MaterializeMinimal generates minimal .claude/ infrastructure without a rite.
// This is suitable for cross-cutting mode (session tracking without orchestrated workflows).
// It creates: CLAUDE.md (base sections), hooks, KNOSSOS_MANIFEST.yaml
// It does NOT create: agents/, skills/, ACTIVE_RITE
func (m *Materializer) MaterializeMinimal(opts Options) (*Result, error) {
	result := &Result{
		OrphansDetected: []string{},
		OrphanAction:    "minimal",
		Source:          "minimal",
	}

	claudeDir := m.resolver.ClaudeDir()

	// Dry-run: just return success
	if opts.DryRun {
		return result, nil
	}

	// Ensure .claude/ directory exists
	if err := paths.EnsureDir(claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create .claude directory", err)
	}

	// Generate hooks from templates (if available)
	hooksSkipped, err := m.materializeHooks(claudeDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize hooks", err)
	}
	result.HooksSkipped = hooksSkipped

	// Generate minimal CLAUDE.md (no agents)
	legacyBackupPath, err := m.materializeMinimalCLAUDEmd(claudeDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}
	result.LegacyBackupPath = legacyBackupPath

	// Generate settings.local.json if needed (no manifest in minimal mode)
	if err := m.materializeSettingsWithManifest(claudeDir, nil); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
	}

	// Remove ACTIVE_RITE file if it exists (cross-cutting mode has no rite)
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	os.Remove(activeRitePath) // Ignore error if doesn't exist

	return result, nil
}

// MaterializeWithOptions generates the .claude/ directory with configurable orphan handling.
func (m *Materializer) MaterializeWithOptions(activeRiteName string, opts Options) (*Result, error) {
	result := &Result{
		OrphansDetected: []string{},
		OrphanAction:    "kept",
	}

	claudeDir := m.resolver.ClaudeDir()

	// Check if already on this rite (skip unless --force)
	if !opts.Force && !opts.DryRun {
		currentRite, err := m.getCurrentRite(claudeDir)
		if err == nil && currentRite == activeRiteName {
			result.OrphanAction = "skipped"
			return result, nil
		}
	}

	// 1. Resolve rite source using 4-tier resolution
	resolved, err := m.sourceResolver.ResolveRite(activeRiteName, m.explicitSource)
	if err != nil {
		return nil, err // Error already has good context from SourceResolver
	}

	// Use resolved rite path and record source info
	ritePath := resolved.RitePath
	result.Source = string(resolved.Source.Type)
	result.SourcePath = resolved.Source.Path

	// Update templates dir if resolved from different source
	if resolved.TemplatesDir != "" {
		m.templatesDir = resolved.TemplatesDir
	}

	// Load the rite manifest from resolved path
	manifest, err := m.loadRiteManifest(ritePath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to load rite manifest", err)
	}

	// Dry-run: just detect orphans and return
	if opts.DryRun {
		orphans, err := m.detectOrphans(manifest, claudeDir)
		if err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to detect orphans", err)
		}
		result.OrphansDetected = orphans
		return result, nil
	}

	// 2. Ensure .claude/ directory exists
	if err := paths.EnsureDir(claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create .claude directory", err)
	}

	// 3. Handle orphans before materializing agents
	orphans, err := m.detectOrphans(manifest, claudeDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to detect orphans", err)
	}
	result.OrphansDetected = orphans

	if len(orphans) > 0 {
		if opts.RemoveAll {
			backupPath, err := m.backupAndRemoveOrphans(orphans, claudeDir)
			if err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to remove orphans", err)
			}
			result.OrphanAction = "removed"
			result.BackupPath = backupPath
		} else if opts.PromoteAll {
			if err := m.promoteOrphans(orphans, claudeDir); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to promote orphans", err)
			}
			result.OrphanAction = "promoted"
		}
		// KeepAll: do nothing (orphans remain)
	}

	// 4. Generate agents/ directory from rite
	if err := m.materializeAgents(manifest, ritePath, claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize agents", err)
	}

	// 5. Generate commands/ and skills/ directories from rite + shared + dependencies + mena
	if err := m.materializeMena(manifest, claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize mena", err)
	}

	// 6. Generate hooks/ directory from templates/hooks
	hooksSkipped, err := m.materializeHooks(claudeDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize hooks", err)
	}
	result.HooksSkipped = hooksSkipped

	// 7. Generate CLAUDE.md from inscription system
	legacyBackupPath, err := m.materializeCLAUDEmd(manifest, claudeDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}
	result.LegacyBackupPath = legacyBackupPath

	// 8. Generate or update settings.local.json with MCP servers from manifest
	if err := m.materializeSettingsWithManifest(claudeDir, manifest); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
	}

	// 9. Track state in .claude/sync/state.json
	if err := m.trackState(manifest, activeRiteName); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to track state", err)
	}

	// 10. Write ACTIVE_RITE marker
	if err := m.writeActiveRite(activeRiteName, claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write ACTIVE_RITE", err)
	}

	return result, nil
}

// detectOrphans finds agent files that are not in the incoming rite's manifest.
func (m *Materializer) detectOrphans(manifest *RiteManifest, claudeDir string) ([]string, error) {
	agentsDir := filepath.Join(claudeDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Build set of expected agents from manifest
	expectedAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		expectedAgents[agent.Name+".md"] = true
	}

	// Find files that aren't expected
	orphans := []string{}
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !expectedAgents[entry.Name()] {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				orphans = append(orphans, entry.Name())
			}
		}
	}

	return orphans, nil
}

// backupAndRemoveOrphans creates a backup of orphan agents then removes them.
func (m *Materializer) backupAndRemoveOrphans(orphans []string, claudeDir string) (string, error) {
	if len(orphans) == 0 {
		return "", nil
	}

	agentsDir := filepath.Join(claudeDir, "agents")
	backupDir := filepath.Join(claudeDir, ".orphan-backup", time.Now().Format("20060102-150405"))

	if err := paths.EnsureDir(backupDir); err != nil {
		return "", err
	}

	for _, orphan := range orphans {
		srcPath := filepath.Join(agentsDir, orphan)
		dstPath := filepath.Join(backupDir, orphan)

		// Copy to backup
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return "", err
		}

		// Remove original
		if err := os.Remove(srcPath); err != nil {
			return "", err
		}
	}

	return backupDir, nil
}

// promoteOrphans moves orphan agents to user-level ~/.claude/agents/.
func (m *Materializer) promoteOrphans(orphans []string, claudeDir string) error {
	if len(orphans) == 0 {
		return nil
	}

	agentsDir := filepath.Join(claudeDir, "agents")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	userAgentsDir := filepath.Join(homeDir, ".claude", "agents")

	if err := paths.EnsureDir(userAgentsDir); err != nil {
		return err
	}

	for _, orphan := range orphans {
		srcPath := filepath.Join(agentsDir, orphan)
		dstPath := filepath.Join(userAgentsDir, orphan)

		// Copy to user-level
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return err
		}

		// Remove from project-level
		if err := os.Remove(srcPath); err != nil {
			return err
		}
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

// materializeMena copies mena files to .claude/commands/ or .claude/skills/
// based on the filename convention (.dro.md for dromena, .lego.md for legomena).
// Sources: mena/, rites/{rite}/mena/, rites/shared/mena/
// Priority order (later sources override earlier): mena < shared < dependencies < current rite
func (m *Materializer) materializeMena(manifest *RiteManifest, claudeDir string) error {
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	// Remove and recreate commands directory
	if err := os.RemoveAll(commandsDir); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := paths.EnsureDir(commandsDir); err != nil {
		return err
	}

	// Remove and recreate skills directory (routing places legomena here)
	if err := os.RemoveAll(skillsDir); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := paths.EnsureDir(skillsDir); err != nil {
		return err
	}

	// Priority order for sources (later sources can override earlier)
	sources := []string{}

	// 1. User-level mena (lowest priority, can be overridden)
	if menaDir := m.getMenaDir(); menaDir != "" {
		sources = append(sources, menaDir)
	}

	// 2. Shared rite mena
	sharedMenaDir := filepath.Join(m.ritesDir, "shared", "mena")
	sources = append(sources, sharedMenaDir)

	// 3. Dependency rite mena (in order)
	for _, dep := range manifest.Dependencies {
		if dep != "shared" { // Already added shared
			sources = append(sources, filepath.Join(m.ritesDir, dep, "mena"))
		}
	}

	// 4. Current rite mena (highest priority)
	currentRiteMenaDir := filepath.Join(m.ritesDir, manifest.Name, "mena")
	sources = append(sources, currentRiteMenaDir)

	// Pass 1: Collect command directories from all sources.
	// Later sources override earlier ones for the same command name.
	// Each entry maps commandName -> source directory path.
	collected := make(map[string]string)
	for _, source := range sources {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			// Later sources override earlier: same key gets overwritten
			collected[entry.Name()] = filepath.Join(source, entry.Name())
		}
	}

	// Pass 2: Route each collected command directory by filename convention.
	// Files with .dro.md -> .claude/commands/ (dromena, invokable)
	// Files with .lego.md -> .claude/skills/ (legomena, reference)
	// Default (INDEX.md or other) -> .claude/commands/ for backward compatibility
	for name, srcDir := range collected {
		menaType := "dro" // default: route to commands/

		// Check for INDEX.md variants to determine mena type
		if entries, err := os.ReadDir(srcDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
					menaType = DetectMenaType(entry.Name())
					break
				}
			}
		}

		var destDir string
		if menaType == "dro" {
			destDir = filepath.Join(commandsDir, name)
		} else {
			destDir = filepath.Join(skillsDir, name)
		}

		if err := m.copyDir(srcDir, destDir); err != nil {
			return err
		}
	}

	return nil
}

// getMenaDir returns the mena directory path.
// Checks project-level first, then falls back to Knossos platform level.
func (m *Materializer) getMenaDir() string {
	// Check for project-level mena first
	projectMena := filepath.Join(m.resolver.ProjectRoot(), "mena")
	if _, err := os.Stat(projectMena); err == nil {
		return projectMena
	}

	// Fall back to Knossos platform mena
	if m.sourceResolver.knossosHome != "" {
		knossosMena := filepath.Join(m.sourceResolver.knossosHome, "mena")
		if _, err := os.Stat(knossosMena); err == nil {
			return knossosMena
		}
	}

	return ""
}

// materializeHooks copies hook files from templates/hooks to .claude/hooks/
// Returns (skipped bool, err error) where skipped=true if no templates/hooks dir exists.
func (m *Materializer) materializeHooks(claudeDir string) (bool, error) {
	hooksDir := filepath.Join(claudeDir, "hooks")
	sourceHooksDir := filepath.Join(m.templatesDir, "hooks")

	// Check if templates/hooks exists (it may not yet)
	if _, err := os.Stat(sourceHooksDir); os.IsNotExist(err) {
		// No templates/hooks directory - this is expected for consumer projects
		// or when hooks are managed separately. Preserve existing hooks if any.
		return true, nil
	}

	// Remove existing hooks directory
	if err := os.RemoveAll(hooksDir); err != nil && !os.IsNotExist(err) {
		return false, err
	}

	// Create fresh hooks directory
	if err := paths.EnsureDir(hooksDir); err != nil {
		return false, err
	}

	// Copy hooks
	return false, m.copyDir(sourceHooksDir, hooksDir)
}

// materializeCLAUDEmd generates CLAUDE.md using the inscription system.
// Returns the path to legacy backup if migration occurred, or empty string if no backup.
func (m *Materializer) materializeCLAUDEmd(manifest *RiteManifest, claudeDir string) (string, error) {
	// Load or create KNOSSOS_MANIFEST.yaml
	knossosManifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	loader := inscription.NewManifestLoader(m.resolver.ProjectRoot())
	loader.ManifestPath = knossosManifestPath

	// Check if manifest exists before loading (for save decision later)
	manifestExists := loader.Exists()

	inscriptionManifest, err := loader.LoadOrCreate()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to load or create KNOSSOS_MANIFEST.yaml", err)
	}

	// Save manifest if newly created (ensures it persists for future runs)
	if !manifestExists {
		if err := loader.Save(inscriptionManifest); err != nil {
			return "", errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
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

	// Create generator with template directory for section rendering
	generator := inscription.NewGenerator(m.templatesDir, inscriptionManifest, renderCtx)

	// Generate all sections
	sections, err := generator.GenerateAll()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to generate CLAUDE.md sections", err)
	}

	// Read existing CLAUDE.md and check for legacy format
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	existingContent := ""
	legacyBackupPath := ""

	if data, err := os.ReadFile(claudeMdPath); err == nil {
		existingContent = string(data)

		// Detect legacy CLAUDE.md (no KNOSSOS markers) and backup before overwriting
		if !strings.Contains(existingContent, "<!-- KNOSSOS:START") && len(existingContent) > 0 {
			legacyBackupPath = fmt.Sprintf("%s.legacy-%s", claudeMdPath, time.Now().Format("20060102-150405"))
			if err := os.WriteFile(legacyBackupPath, data, 0644); err != nil {
				return "", errors.Wrap(errors.CodeGeneralError, "failed to backup legacy CLAUDE.md", err)
			}
			// Clear existing content so merger generates fresh file
			existingContent = ""
		}
	}

	// Merge sections
	merger := inscription.NewMerger(inscriptionManifest, generator)
	mergeResult, err := merger.MergeRegions(existingContent, sections)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to merge CLAUDE.md regions", err)
	}

	// Write CLAUDE.md
	if err := os.WriteFile(claudeMdPath, []byte(mergeResult.Content), 0644); err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to write CLAUDE.md", err)
	}

	return legacyBackupPath, nil
}

// materializeMinimalCLAUDEmd generates CLAUDE.md for cross-cutting mode (no agents).
func (m *Materializer) materializeMinimalCLAUDEmd(claudeDir string) (string, error) {
	// Load or create KNOSSOS_MANIFEST.yaml
	knossosManifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	loader := inscription.NewManifestLoader(m.resolver.ProjectRoot())
	loader.ManifestPath = knossosManifestPath

	// Check if manifest exists before loading
	manifestExists := loader.Exists()

	inscriptionManifest, err := loader.LoadOrCreate()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to load or create KNOSSOS_MANIFEST.yaml", err)
	}

	// Save manifest if newly created
	if !manifestExists {
		if err := loader.Save(inscriptionManifest); err != nil {
			return "", errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
	}

	// Build minimal render context (no agents, no rite)
	renderCtx := &inscription.RenderContext{
		ActiveRite:  "", // Cross-cutting mode has no rite
		AgentCount:  0,
		Agents:      []inscription.AgentInfo{},
		KnossosVars: make(map[string]string),
		ProjectRoot: m.resolver.ProjectRoot(),
	}

	// Create generator with template directory
	generator := inscription.NewGenerator(m.templatesDir, inscriptionManifest, renderCtx)

	// Generate all sections
	sections, err := generator.GenerateAll()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to generate CLAUDE.md sections", err)
	}

	// Read existing CLAUDE.md and check for legacy format
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	existingContent := ""
	legacyBackupPath := ""

	if data, err := os.ReadFile(claudeMdPath); err == nil {
		existingContent = string(data)

		// Detect legacy CLAUDE.md (no KNOSSOS markers) and backup before overwriting
		if !strings.Contains(existingContent, "<!-- KNOSSOS:START") && len(existingContent) > 0 {
			legacyBackupPath = fmt.Sprintf("%s.legacy-%s", claudeMdPath, time.Now().Format("20060102-150405"))
			if err := os.WriteFile(legacyBackupPath, data, 0644); err != nil {
				return "", errors.Wrap(errors.CodeGeneralError, "failed to backup legacy CLAUDE.md", err)
			}
			existingContent = ""
		}
	}

	// Merge sections
	merger := inscription.NewMerger(inscriptionManifest, generator)
	mergeResult, err := merger.MergeRegions(existingContent, sections)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to merge CLAUDE.md regions", err)
	}

	// Write CLAUDE.md
	if err := os.WriteFile(claudeMdPath, []byte(mergeResult.Content), 0644); err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to write CLAUDE.md", err)
	}

	return legacyBackupPath, nil
}

// materializeSettings generates or updates settings.local.json with MCP servers from manifest.
func (m *Materializer) materializeSettings(claudeDir string) error {
	return m.materializeSettingsWithManifest(claudeDir, nil)
}

// materializeSettingsWithManifest generates or updates settings.local.json.
// If manifest has MCP servers, merges them into existing settings.
// If no manifest or no MCP servers, creates minimal settings if needed.
func (m *Materializer) materializeSettingsWithManifest(claudeDir string, manifest *RiteManifest) error {
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Load existing settings or create empty map
	existingSettings, err := loadExistingSettings(settingsPath)
	if err != nil {
		return err
	}

	// Ensure hooks key exists
	if existingSettings["hooks"] == nil {
		existingSettings["hooks"] = make(map[string]any)
	}

	// If manifest has MCP servers, merge them
	if manifest != nil && len(manifest.MCPServers) > 0 {
		existingSettings = mergeMCPServers(existingSettings, manifest.MCPServers)
	}

	// Write settings (always write, even if no changes, to ensure file exists)
	return saveSettings(settingsPath, existingSettings)
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

// getCurrentRite reads the current active rite from ACTIVE_RITE file.
func (m *Materializer) getCurrentRite(claudeDir string) (string, error) {
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	data, err := os.ReadFile(activeRitePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
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
