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
	resolver          *paths.Resolver
	sourceResolver    *SourceResolver
	explicitSource    string // Optional explicit source from --source flag
	ritesDir          string // Deprecated: use sourceResolver
	templatesDir      string
	embeddedTemplates fs.FS  // Embedded templates filesystem
	embeddedHooksYAML []byte // Embedded hooks.yaml content
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

// WithEmbeddedFS sets the embedded rites filesystem on both the materializer's
// source resolver and stores it for rite content access. Returns the receiver.
func (m *Materializer) WithEmbeddedFS(fsys fs.FS) *Materializer {
	m.sourceResolver.WithEmbeddedFS(fsys)
	return m
}

// WithEmbeddedTemplates sets the embedded templates filesystem.
func (m *Materializer) WithEmbeddedTemplates(fsys fs.FS) *Materializer {
	m.embeddedTemplates = fsys
	return m
}

// WithEmbeddedHooks sets the embedded hooks.yaml content.
func (m *Materializer) WithEmbeddedHooks(data []byte) *Materializer {
	m.embeddedHooksYAML = data
	return m
}

// riteFS returns a filesystem rooted at the rite's directory.
// For embedded sources, returns a sub-FS of the embedded rites.
// For filesystem sources, returns os.DirFS rooted at the rite path.
func (m *Materializer) riteFS(resolved *ResolvedRite) fs.FS {
	if resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil {
		sub, err := fs.Sub(m.sourceResolver.embeddedFS, resolved.RitePath)
		if err != nil {
			return os.DirFS(resolved.RitePath)
		}
		return sub
	}
	return os.DirFS(resolved.RitePath)
}

// templatesFS returns a filesystem for templates.
// For embedded sources, returns a sub-FS of the embedded templates.
// For filesystem sources, returns os.DirFS rooted at the templates dir.
func (m *Materializer) templatesFS(resolved *ResolvedRite) fs.FS {
	if resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		sub, err := fs.Sub(m.embeddedTemplates, resolved.TemplatesDir)
		if err != nil {
			return os.DirFS(m.templatesDir)
		}
		return sub
	}
	return os.DirFS(m.templatesDir)
}

// copyDirFromFS copies all files from an fs.FS to a destination directory on disk.
func copyDirFromFS(fsys fs.FS, dst string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, path)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(destPath, content, 0644)
	})
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
	hooksSkipped, err := m.materializeHooks(claudeDir, nil)
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
	manifest, err := m.loadRiteManifest(ritePath, resolved)
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
	if err := m.materializeAgents(manifest, ritePath, claudeDir, resolved); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize agents", err)
	}

	// 5. Generate commands/ and skills/ directories from rite + shared + dependencies + mena
	if err := m.materializeMena(manifest, claudeDir, resolved); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize mena", err)
	}

	// 6. Generate hooks/ directory from templates/hooks
	hooksSkipped, err := m.materializeHooks(claudeDir, resolved)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize hooks", err)
	}
	result.HooksSkipped = hooksSkipped

	// 7. Generate CLAUDE.md from inscription system
	legacyBackupPath, err := m.materializeCLAUDEmd(manifest, claudeDir, resolved)
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
// When resolved is non-nil and the source is embedded, reads from the embedded FS.
func (m *Materializer) loadRiteManifest(ritePath string, resolved *ResolvedRite) (*RiteManifest, error) {
	var data []byte
	var err error

	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil {
		data, err = fs.ReadFile(m.sourceResolver.embeddedFS, resolved.ManifestPath)
	} else {
		manifestPath := filepath.Join(ritePath, "manifest.yaml")
		data, err = os.ReadFile(manifestPath)
	}
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
func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, claudeDir string, resolved *ResolvedRite) error {
	agentsDir := filepath.Join(claudeDir, "agents")

	// Remove existing agents directory
	if err := os.RemoveAll(agentsDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create fresh agents directory
	if err := paths.EnsureDir(agentsDir); err != nil {
		return err
	}

	// For embedded sources, use fs.FS to read agent files
	if resolved != nil && resolved.Source.Type == SourceEmbedded {
		rFS := m.riteFS(resolved)
		agentsSub, err := fs.Sub(rFS, "agents")
		if err != nil {
			return nil // No agents sub-dir in embedded FS
		}
		// Check if agents dir exists in embedded FS
		if _, err := fs.Stat(rFS, "agents"); err != nil {
			return nil // No agents in this rite
		}
		return copyDirFromFS(agentsSub, agentsDir)
	}

	// Filesystem path: use existing os-based copy
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

// menaSource represents a source for mena files, which can be either
// a filesystem path or an embedded FS path.
type menaSource struct {
	path       string // Filesystem path (for os-based sources)
	fsys       fs.FS  // Embedded filesystem (nil for os-based sources)
	fsysPath   string // Path within fsys (e.g., "rites/shared/mena")
	isEmbedded bool
}

// menaCollectedEntry represents a leaf mena directory collected for routing.
type menaCollectedEntry struct {
	source menaSource
	name   string
}

// menaStandaloneFile represents a standalone file in a grouping directory.
type menaStandaloneFile struct {
	srcPath string
	relPath string // e.g., "navigation/rite.dro.md"
}

// materializeMena copies mena files to .claude/commands/ or .claude/skills/
// based on the filename convention (.dro.md for dromena, .lego.md for legomena).
// Sources: mena/, rites/{rite}/mena/, rites/shared/mena/
// Priority order (later sources override earlier): mena < shared < dependencies < current rite
func (m *Materializer) materializeMena(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite) error {
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

	isEmbedded := resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil

	// Priority order for sources (later sources can override earlier)
	var sources []menaSource

	// 1. User-level mena (lowest priority, can be overridden)
	// Always from filesystem, never from embedded
	if menaDir := m.getMenaDir(); menaDir != "" {
		sources = append(sources, menaSource{path: menaDir})
	}

	if isEmbedded {
		// For embedded sources, read mena from embedded FS
		embFS := m.sourceResolver.embeddedFS

		// 2. Shared rite mena
		sources = append(sources, menaSource{fsys: embFS, fsysPath: "rites/shared/mena", isEmbedded: true})

		// 3. Dependency rite mena (in order)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, menaSource{fsys: embFS, fsysPath: "rites/" + dep + "/mena", isEmbedded: true})
			}
		}

		// 4. Current rite mena (highest priority)
		sources = append(sources, menaSource{fsys: embFS, fsysPath: "rites/" + manifest.Name + "/mena", isEmbedded: true})
	} else {
		// 2. Shared rite mena
		sharedMenaDir := filepath.Join(m.ritesDir, "shared", "mena")
		sources = append(sources, menaSource{path: sharedMenaDir})

		// 3. Dependency rite mena (in order)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, menaSource{path: filepath.Join(m.ritesDir, dep, "mena")})
			}
		}

		// 4. Current rite mena (highest priority)
		currentRiteMenaDir := filepath.Join(m.ritesDir, manifest.Name, "mena")
		sources = append(sources, menaSource{path: currentRiteMenaDir})
	}

	// Pass 1: Collect mena entries from all sources.
	// Later sources override earlier ones for the same command name.
	// Directories with INDEX files are leaf entries; directories without
	// INDEX files are grouping directories that are recursively descended.
	collected := make(map[string]menaCollectedEntry)
	standalones := make(map[string]menaStandaloneFile) // key: relPath

	for _, src := range sources {
		if src.isEmbedded {
			collectMenaEntriesFS(src.fsys, src.fsysPath, "", collected)
		} else {
			if src.path == "" {
				continue
			}
			if _, err := os.Stat(src.path); os.IsNotExist(err) {
				continue
			}
			if err := collectMenaEntriesDir(src.path, "", collected, standalones); err != nil {
				return err
			}
		}
	}

	// Pass 2: Route each collected command directory by filename convention.
	for name, ce := range collected {
		menaType := "dro" // default: route to commands/

		if ce.source.isEmbedded {
			entries, err := fs.ReadDir(ce.source.fsys, ce.source.fsysPath)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
						menaType = DetectMenaType(entry.Name())
						break
					}
				}
			}
		} else {
			if entries, err := os.ReadDir(ce.source.path); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
						menaType = DetectMenaType(entry.Name())
						break
					}
				}
			}
		}

		var destDir string
		if menaType == "dro" {
			destDir = filepath.Join(commandsDir, name)
		} else {
			destDir = filepath.Join(skillsDir, name)
		}

		if ce.source.isEmbedded {
			sub, err := fs.Sub(ce.source.fsys, ce.source.fsysPath)
			if err != nil {
				return err
			}
			if err := copyDirFromFS(sub, destDir); err != nil {
				return err
			}
		} else {
			if err := m.copyDir(ce.source.path, destDir); err != nil {
				return err
			}
		}
	}

	// Copy standalone files (e.g., mena/navigation/rite.dro.md)
	// Route by extension: .dro.md → commands/, .lego.md → skills/
	for _, sf := range standalones {
		menaType := DetectMenaType(filepath.Base(sf.srcPath))
		var baseDir string
		if menaType == "dro" {
			baseDir = commandsDir
		} else {
			baseDir = skillsDir
		}
		destPath := filepath.Join(baseDir, sf.relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		data, err := os.ReadFile(sf.srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// collectMenaEntriesDir recursively collects mena entries from a filesystem directory.
// Leaf directories (containing INDEX files) are collected for routing.
// Standalone files in grouping directories are collected separately.
func collectMenaEntriesDir(dirPath string, prefix string, collected map[string]menaCollectedEntry, standalones map[string]menaStandaloneFile) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if prefix != "" {
			name = prefix + "/" + entry.Name()
		}
		if entry.IsDir() {
			childPath := filepath.Join(dirPath, entry.Name())
			if dirHasIndexFile(childPath) {
				collected[name] = menaCollectedEntry{
					source: menaSource{path: childPath},
					name:   name,
				}
			} else {
				if err := collectMenaEntriesDir(childPath, name, collected, standalones); err != nil {
					return err
				}
			}
		} else {
			// Standalone file in a grouping directory
			standalones[name] = menaStandaloneFile{
				srcPath: filepath.Join(dirPath, entry.Name()),
				relPath: name,
			}
		}
	}
	return nil
}

// collectMenaEntriesFS recursively collects mena entries from an embedded filesystem.
func collectMenaEntriesFS(fsys fs.FS, fsysPath string, prefix string, collected map[string]menaCollectedEntry) {
	entries, err := fs.ReadDir(fsys, fsysPath)
	if err != nil {
		return // Path doesn't exist in embedded FS
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Embedded rite mena doesn't have standalone files
		}
		name := entry.Name()
		if prefix != "" {
			name = prefix + "/" + entry.Name()
		}
		childPath := fsysPath + "/" + entry.Name()
		if fsHasIndexFile(fsys, childPath) {
			collected[name] = menaCollectedEntry{
				source: menaSource{
					fsys:       fsys,
					fsysPath:   childPath,
					isEmbedded: true,
				},
				name: name,
			}
		} else {
			collectMenaEntriesFS(fsys, childPath, name, collected)
		}
	}
}

// dirHasIndexFile checks if a filesystem directory contains an INDEX file.
func dirHasIndexFile(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			return true
		}
	}
	return false
}

// fsHasIndexFile checks if an embedded FS directory contains an INDEX file.
func fsHasIndexFile(fsys fs.FS, dirPath string) bool {
	entries, err := fs.ReadDir(fsys, dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			return true
		}
	}
	return false
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
func (m *Materializer) materializeHooks(claudeDir string, resolved *ResolvedRite) (bool, error) {
	hooksDir := filepath.Join(claudeDir, "hooks")

	// For embedded sources, try embedded templates FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		tFS := m.templatesFS(resolved)
		if _, err := fs.Stat(tFS, "hooks"); err != nil {
			return true, nil // No hooks in embedded templates
		}
		hooksSub, err := fs.Sub(tFS, "hooks")
		if err != nil {
			return true, nil
		}

		// Remove existing hooks directory
		if err := os.RemoveAll(hooksDir); err != nil && !os.IsNotExist(err) {
			return false, err
		}
		if err := paths.EnsureDir(hooksDir); err != nil {
			return false, err
		}
		return false, copyDirFromFS(hooksSub, hooksDir)
	}

	// Filesystem path
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
func (m *Materializer) materializeCLAUDEmd(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite) (string, error) {
	// Build render context with full agent details
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

	// Delegate to shared helper with manifest update and hash tracking enabled
	return m.mergeCLAUDEmd(claudeDir, renderCtx, manifest.Name, resolved, true)
}

// materializeMinimalCLAUDEmd generates CLAUDE.md for cross-cutting mode (no agents).
func (m *Materializer) materializeMinimalCLAUDEmd(claudeDir string) (string, error) {
	// Build minimal render context (no agents, no rite)
	renderCtx := &inscription.RenderContext{
		ActiveRite:  "", // Cross-cutting mode has no rite
		AgentCount:  0,
		Agents:      []inscription.AgentInfo{},
		KnossosVars: make(map[string]string),
		ProjectRoot: m.resolver.ProjectRoot(),
	}

	// Delegate to shared helper without manifest update or hash tracking
	return m.mergeCLAUDEmd(claudeDir, renderCtx, "", nil, false)
}

// mergeCLAUDEmd contains the shared logic for generating and merging CLAUDE.md content.
// Parameters:
//   - claudeDir: path to .claude/ directory
//   - renderCtx: pre-built render context (full or minimal)
//   - activeRite: name of active rite (empty string for minimal mode)
//   - resolved: resolved rite information (nil for minimal mode)
//   - updateManifest: if true, updates manifest hashes and saves; if false, only saves if newly created
//
// Returns the path to legacy backup if migration occurred, or empty string if no backup.
func (m *Materializer) mergeCLAUDEmd(claudeDir string, renderCtx *inscription.RenderContext, activeRite string, resolved *ResolvedRite, updateManifest bool) (string, error) {
	// Load or create KNOSSOS_MANIFEST.yaml
	knossosManifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	loader := inscription.NewManifestLoader(m.resolver.ProjectRoot())
	loader.ManifestPath = knossosManifestPath

	manifestExists := loader.Exists()

	inscriptionManifest, err := loader.LoadOrCreate()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to load or create KNOSSOS_MANIFEST.yaml", err)
	}

	// Update active rite in manifest if provided
	if activeRite != "" {
		inscriptionManifest.SetActiveRite(activeRite)
	}

	// Save manifest if newly created (needed for minimal mode)
	if !manifestExists {
		if err := loader.Save(inscriptionManifest); err != nil {
			return "", errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
	}

	// Create generator with template directory for section rendering.
	// When the source is embedded, pass the embedded templates FS to the generator.
	var generator *inscription.Generator
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		tFS := m.templatesFS(resolved)
		generator = inscription.NewGeneratorWithFS(tFS, inscriptionManifest, renderCtx)
	} else {
		generator = inscription.NewGenerator(m.templatesDir, inscriptionManifest, renderCtx)
	}

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

	// Update manifest hashes and save (for full mode only)
	if updateManifest {
		merger.UpdateManifestHashes(mergeResult)
		loader.IncrementVersion(inscriptionManifest)
		if err := loader.Save(inscriptionManifest); err != nil {
			return "", errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
	}

	return legacyBackupPath, nil
}

// materializeSettings generates or updates settings.local.json with MCP servers from manifest.
func (m *Materializer) materializeSettings(claudeDir string) error {
	return m.materializeSettingsWithManifest(claudeDir, nil)
}

// materializeSettingsWithManifest generates or updates settings.local.json.
// If manifest has MCP servers, merges them into existing settings.
// Loads hooks.yaml and merges hook registrations into settings.
// If no manifest or no MCP servers, creates minimal settings if needed.
func (m *Materializer) materializeSettingsWithManifest(claudeDir string, manifest *RiteManifest) error {
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Load existing settings or create empty map
	existingSettings, err := loadExistingSettings(settingsPath)
	if err != nil {
		return err
	}

	// Load hooks.yaml and merge hook registrations
	if hooksConfig := m.loadHooksConfig(); hooksConfig != nil {
		existingSettings = mergeHooksSettings(existingSettings, hooksConfig)
	} else {
		// No hooks.yaml found — ensure hooks key exists (empty)
		if existingSettings["hooks"] == nil {
			existingSettings["hooks"] = make(map[string]any)
		}
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
