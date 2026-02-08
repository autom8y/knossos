// Package materialize generates .claude/ directories from templates and rite manifests.
package materialize

import (
	"bytes"
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
	embeddedHooks     fs.FS  // Embedded hooks filesystem
	claudeDirOverride string // If set, materialize to this directory instead of .claude/
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

// WithEmbeddedHooks sets the embedded hooks filesystem.
func (m *Materializer) WithEmbeddedHooks(fsys fs.FS) *Materializer {
	m.embeddedHooks = fsys
	return m
}

// getClaudeDir returns the target .claude/ directory, respecting any override.
func (m *Materializer) getClaudeDir() string {
	if m.claudeDirOverride != "" {
		return m.claudeDirOverride
	}
	return m.resolver.ClaudeDir()
}

// StagedMaterialize builds the .claude/ directory in a staging copy then
// atomically swaps it into place. This prevents Claude Code's file watcher
// from seeing intermediate states during multi-file writes.
//
// The flow:
//  1. Clone current .claude/ → .claude.staging/ (preserves user content)
//  2. Run materializeFn against the staging directory
//  3. Rename .claude/ → .claude.bak/, .claude.staging/ → .claude/
//  4. Clean up .claude.bak/
//
// The two-rename gap is microseconds — well below CC's file watcher debounce.
func (m *Materializer) StagedMaterialize(materializeFn func(m *Materializer) (*Result, error)) (*Result, error) {
	claudeDir := m.resolver.ClaudeDir()
	stagingDir := claudeDir + ".staging"
	backupDir := claudeDir + ".bak"

	// Clean up any leftover staging/backup dirs from previous failed runs
	os.RemoveAll(stagingDir)
	os.RemoveAll(backupDir)

	// Clone current .claude/ to staging (preserves sessions, user content, etc.)
	if _, err := os.Stat(claudeDir); err == nil {
		if err := cloneDir(claudeDir, stagingDir); err != nil {
			os.RemoveAll(stagingDir)
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to create staging directory", err)
		}
	} else {
		// No existing .claude/ — staging starts empty
		if err := os.MkdirAll(stagingDir, 0755); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to create staging directory", err)
		}
	}

	// Point materializer at staging directory
	m.claudeDirOverride = stagingDir
	defer func() { m.claudeDirOverride = "" }()

	// Run the actual materialization into staging
	result, err := materializeFn(m)
	if err != nil {
		os.RemoveAll(stagingDir)
		return nil, err
	}

	// Atomic swap: .claude → .claude.bak, .claude.staging → .claude
	if _, err := os.Stat(claudeDir); err == nil {
		if err := os.Rename(claudeDir, backupDir); err != nil {
			os.RemoveAll(stagingDir)
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to move .claude to backup", err)
		}
	}
	if err := os.Rename(stagingDir, claudeDir); err != nil {
		// Rollback: restore from backup
		os.Rename(backupDir, claudeDir)
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to swap staging into place", err)
	}

	// Clean up backup
	os.RemoveAll(backupDir)

	return result, nil
}

// cloneDir recursively copies src to dst, preserving directory structure.
func cloneDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Skip .tmp files from interrupted atomic writes
		if strings.HasSuffix(path, ".tmp") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(destPath, content, 0644)
	})
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

// writeIfChanged writes content to path only if it differs from existing content.
// Uses atomic writes (write to temp file, then rename) to prevent Claude Code's
// file watcher from seeing partially-written files.
func writeIfChanged(path string, content []byte, perm os.FileMode) (bool, error) {
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, content) {
		return false, nil
	}
	return true, atomicWriteFile(path, content, perm)
}

// atomicWriteFile writes content to a temp file in the same directory then
// renames it over the target. rename(2) is atomic on POSIX, so the file
// watcher never sees a partially-written file.
func atomicWriteFile(path string, content []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // best-effort cleanup
		return err
	}
	return nil
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
		_, err = writeIfChanged(destPath, content, 0644)
		return err
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

	claudeDir := m.getClaudeDir()

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

	// Generate rules from templates (if available)
	if err := m.materializeRules(claudeDir, nil); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize rules", err)
	}

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

	// Remove rite-specific state files (cross-cutting mode has no rite)
	os.Remove(filepath.Join(claudeDir, "ACTIVE_RITE"))
	os.Remove(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))

	return result, nil
}

// MaterializeWithOptions generates the .claude/ directory with configurable orphan handling.
func (m *Materializer) MaterializeWithOptions(activeRiteName string, opts Options) (*Result, error) {
	result := &Result{
		OrphansDetected: []string{},
		OrphanAction:    "kept",
	}

	claudeDir := m.getClaudeDir()

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

	// 6.5. Generate rules/ directory from templates/rules
	if err := m.materializeRules(claudeDir, resolved); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize rules", err)
	}

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

	// 9.5. Copy workflow.yaml to ACTIVE_WORKFLOW.yaml
	if err := m.materializeWorkflow(claudeDir, resolved); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize workflow", err)
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

		// Write to destination (only if changed)
		destPath := filepath.Join(agentsDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		_, err = writeIfChanged(destPath, content, 0644)
		return err
	})
}

// materializeMena copies mena files to .claude/commands/ or .claude/skills/
// based on the filename convention (.dro.md for dromena, .lego.md for legomena).
// Sources: mena/, rites/{rite}/mena/, rites/shared/mena/
// Priority order (later sources override earlier): mena < shared < dependencies < current rite
//
// This method builds the source list and delegates to ProjectMena() for the
// actual collection, routing, extension stripping, and file copying.
func (m *Materializer) materializeMena(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite) error {
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	isEmbedded := resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil

	// Build priority-ordered source list (later sources override earlier)
	var sources []MenaSource

	// 1. User-level mena (lowest priority, can be overridden)
	// Always from filesystem, never from embedded
	if menaDir := m.getMenaDir(); menaDir != "" {
		sources = append(sources, MenaSource{Path: menaDir})
	}

	if isEmbedded {
		// For embedded sources, read mena from embedded FS
		embFS := m.sourceResolver.embeddedFS

		// 2. Shared rite mena
		sources = append(sources, MenaSource{Fsys: embFS, FsysPath: "rites/shared/mena", IsEmbedded: true})

		// 3. Dependency rite mena (in order)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, MenaSource{Fsys: embFS, FsysPath: "rites/" + dep + "/mena", IsEmbedded: true})
			}
		}

		// 4. Current rite mena (highest priority)
		sources = append(sources, MenaSource{Fsys: embFS, FsysPath: "rites/" + manifest.Name + "/mena", IsEmbedded: true})
	} else {
		// 2. Shared rite mena
		sharedMenaDir := filepath.Join(m.ritesDir, "shared", "mena")
		sources = append(sources, MenaSource{Path: sharedMenaDir})

		// 3. Dependency rite mena (in order)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, MenaSource{Path: filepath.Join(m.ritesDir, dep, "mena")})
			}
		}

		// 4. Current rite mena (highest priority)
		currentRiteMenaDir := filepath.Join(m.ritesDir, manifest.Name, "mena")
		sources = append(sources, MenaSource{Path: currentRiteMenaDir})
	}

	// Delegate to ProjectMena with destructive mode
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject, // Filter out scope:user entries
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := ProjectMena(sources, opts)
	return err
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
					source: MenaSource{Path: childPath},
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
				source: MenaSource{
					Fsys:       fsys,
					FsysPath:   childPath,
					IsEmbedded: true,
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

// materializeRules copies rule files from templates/rules to .claude/rules/
// Platform rules (internal-*.md, mena.md) are overwritten from templates.
// User-created rules (other .md files) are preserved.
func (m *Materializer) materializeRules(claudeDir string, resolved *ResolvedRite) error {
	rulesDir := filepath.Join(claudeDir, "rules")
	if err := paths.EnsureDir(rulesDir); err != nil {
		return err
	}

	// For embedded sources, try embedded templates FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		tFS := m.templatesFS(resolved)
		if _, err := fs.Stat(tFS, "rules"); err != nil {
			return nil // No rules in embedded templates
		}
		entries, err := fs.ReadDir(tFS, "rules")
		if err != nil {
			return nil // Path doesn't exist
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			content, err := fs.ReadFile(tFS, "rules/"+entry.Name())
			if err != nil {
				return err
			}
			dstPath := filepath.Join(rulesDir, entry.Name())
			_, err = writeIfChanged(dstPath, content, 0644)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Filesystem path: templates/rules/
	sourceRulesDir := filepath.Join(m.templatesDir, "rules")

	// Check if templates/rules exists
	entries, err := os.ReadDir(sourceRulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No template rules = no-op
		}
		return err
	}

	// Copy each template rule file using writeIfChanged for idempotency
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(sourceRulesDir, entry.Name()))
		if err != nil {
			return err
		}
		dstPath := filepath.Join(rulesDir, entry.Name())
		_, err = writeIfChanged(dstPath, content, 0644)
		if err != nil {
			return err
		}
	}

	return nil
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

	// Write CLAUDE.md (only if content changed, to avoid triggering Claude Code file watcher)
	if _, err := writeIfChanged(claudeMdPath, []byte(mergeResult.Content), 0644); err != nil {
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

	// Write settings (only if content changed, to avoid triggering Claude Code file watcher)
	return saveSettings(settingsPath, existingSettings)
}

// trackState updates .claude/sync/state.json with materialization metadata.
func (m *Materializer) trackState(manifest *RiteManifest, activeRiteName string) error {
	stateManager := sync.NewStateManager(m.resolver)

	// During staged materialization, override the sync dir to target the staging directory.
	if m.claudeDirOverride != "" {
		stateManager.SetSyncDir(filepath.Join(m.claudeDirOverride, "sync"))
	}

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

	// Update active rite and last sync time
	state.ActiveRite = activeRiteName
	state.LastSync = time.Now().UTC()
	return stateManager.Save(state)
}

// materializeWorkflow copies workflow.yaml from the rite to .claude/ACTIVE_WORKFLOW.yaml.
// Returns nil if the rite has no workflow.yaml (non-fatal).
func (m *Materializer) materializeWorkflow(claudeDir string, resolved *ResolvedRite) error {
	rFS := m.riteFS(resolved)
	content, err := fs.ReadFile(rFS, "workflow.yaml")
	if err != nil {
		// No workflow.yaml in this rite — not an error
		return nil
	}
	dstPath := filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml")
	_, err = writeIfChanged(dstPath, content, 0644)
	return err
}

// writeActiveRite writes the ACTIVE_RITE marker file.
func (m *Materializer) writeActiveRite(riteName, claudeDir string) error {
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	_, err := writeIfChanged(activeRitePath, []byte(riteName+"\n"), 0644)
	return err
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

		// Read and write file (only if changed)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = writeIfChanged(destPath, content, 0644)
		return err
	})
}
