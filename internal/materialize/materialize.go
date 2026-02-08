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
	"github.com/autom8y/knossos/internal/fileutil"
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
	Status           string   // Pipeline status: "success", "skipped", "minimal"
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
	Command string            `yaml:"command,omitempty" json:"command,omitempty"` // stdio only
	Args    []string          `yaml:"args,omitempty" json:"args,omitempty"`       // stdio only
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Type    string            `yaml:"type,omitempty" json:"type,omitempty"`       // stdio (default), sse, http
	URL     string            `yaml:"url,omitempty" json:"url,omitempty"`         // sse/http only
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"` // sse/http only
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
	return fileutil.WriteIfChanged(path, content, perm)
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

// MaterializeMinimal generates minimal .claude/ infrastructure without a rite.
// This is suitable for cross-cutting mode (session tracking without orchestrated workflows).
// It creates: CLAUDE.md (base sections), hooks, KNOSSOS_MANIFEST.yaml
// It does NOT create: agents/, skills/, ACTIVE_RITE
func (m *Materializer) MaterializeMinimal(opts Options) (*Result, error) {
	result := &Result{
		Status:          "minimal",
		OrphansDetected: []string{},
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
	os.Remove(filepath.Join(claudeDir, "INVOCATION_STATE.yaml"))

	return result, nil
}

// MaterializeWithOptions generates the .claude/ directory with configurable orphan handling.
func (m *Materializer) MaterializeWithOptions(activeRiteName string, opts Options) (*Result, error) {
	result := &Result{
		Status:          "success",
		OrphansDetected: []string{},
		OrphanAction:    "kept",
	}

	claudeDir := m.getClaudeDir()

	// Note: the skip guard (skip-if-same-rite) was removed. The pipeline is safe
	// to always run: selective write preserves user content, and writeIfChanged()
	// prevents unnecessary disk writes. See ADR: "ari sync is safe to run repeatedly."

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

	// 2.5. Clear stale invocation state from previous rite
	if err := m.clearInvocationState(claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to clear invocation state", err)
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
// Uses selective write: only knossos-managed agents (from manifest) are replaced.
// User-created agents not in the manifest are preserved.
func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, claudeDir string, resolved *ResolvedRite) error {
	agentsDir := filepath.Join(claudeDir, "agents")

	// Ensure agents directory exists (selective — do NOT RemoveAll)
	if err := paths.EnsureDir(agentsDir); err != nil {
		return err
	}

	// Build managed agent set from manifest
	managedAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		managedAgents[agent.Name+".md"] = true
	}

	// Remove only knossos-managed agents (will be rewritten below).
	// User-created agents not in manifest are untouched.
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && managedAgents[entry.Name()] {
				os.Remove(filepath.Join(agentsDir, entry.Name()))
			}
		}
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
	} else if resolved != nil {
		// Derive rites base directory from the resolved rite path
		ritesBase := filepath.Dir(resolved.RitePath)

		// 2. Shared rite mena
		sharedMenaDir := filepath.Join(ritesBase, "shared", "mena")
		sources = append(sources, MenaSource{Path: sharedMenaDir})

		// 3. Dependency rite mena (in order)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, MenaSource{Path: filepath.Join(ritesBase, dep, "mena")})
			}
		}

		// 4. Current rite mena (highest priority)
		currentRiteMenaDir := filepath.Join(resolved.RitePath, "mena")
		sources = append(sources, MenaSource{Path: currentRiteMenaDir})
	}

	// Delegate to ProjectMena with destructive mode
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
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
// Platform rules are overwritten from templates; user-created rules are preserved.
// On rite switch, stale knossos-managed rules are removed before writing new ones.
// Provenance is determined by template filename: any .md file whose name matches
// a template source file is knossos-managed; all others are user-created.
func (m *Materializer) materializeRules(claudeDir string, resolved *ResolvedRite) error {
	rulesDir := filepath.Join(claudeDir, "rules")
	if err := paths.EnsureDir(rulesDir); err != nil {
		return err
	}

	// Collect template rule names and content from appropriate source
	templateRules := make(map[string][]byte)

	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		tFS := m.templatesFS(resolved)
		if _, err := fs.Stat(tFS, "rules"); err != nil {
			return nil // No rules in embedded templates
		}
		entries, err := fs.ReadDir(tFS, "rules")
		if err != nil {
			return nil
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			content, err := fs.ReadFile(tFS, "rules/"+entry.Name())
			if err != nil {
				return err
			}
			templateRules[entry.Name()] = content
		}
	} else {
		sourceRulesDir := filepath.Join(m.templatesDir, "rules")
		entries, err := os.ReadDir(sourceRulesDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil // No template rules = no-op
			}
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			content, err := os.ReadFile(filepath.Join(sourceRulesDir, entry.Name()))
			if err != nil {
				return err
			}
			templateRules[entry.Name()] = content
		}
	}

	// Also collect template names from the filesystem templates dir (for the
	// provenance check even when using embedded sources). This ensures we know
	// all possible knossos-managed filenames across sources.
	allTemplateNames := make(map[string]bool)
	for name := range templateRules {
		allTemplateNames[name] = true
	}
	if fsRulesDir := filepath.Join(m.templatesDir, "rules"); fsRulesDir != "" {
		if entries, err := os.ReadDir(fsRulesDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
					allTemplateNames[entry.Name()] = true
				}
			}
		}
	}

	// Remove stale knossos-managed rules (names matching any template source)
	if existingRules, err := os.ReadDir(rulesDir); err == nil {
		for _, entry := range existingRules {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			if allTemplateNames[entry.Name()] {
				os.Remove(filepath.Join(rulesDir, entry.Name()))
			}
		}
	}

	// Write current template rules
	for name, content := range templateRules {
		dstPath := filepath.Join(rulesDir, name)
		if _, err := writeIfChanged(dstPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

// materializeHooks copies hook files from templates/hooks to .claude/hooks/
// Uses selective write: only knossos-managed hook files (from templates) are replaced.
// User-created hook files not in the template source are preserved.
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

		if err := paths.EnsureDir(hooksDir); err != nil {
			return false, err
		}

		// Build managed set from embedded source, then selectively remove
		managedHooks := collectFSFilenames(hooksSub)
		removeManagedFiles(hooksDir, managedHooks)

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

	if err := paths.EnsureDir(hooksDir); err != nil {
		return false, err
	}

	// Build managed set from source, then selectively remove
	managedHooks := collectDirFilenames(sourceHooksDir)
	removeManagedFiles(hooksDir, managedHooks)

	// Copy hooks (uses writeIfChanged internally)
	return false, m.copyDir(sourceHooksDir, hooksDir)
}

// collectDirFilenames returns a set of all filenames in a directory (non-recursive).
func collectDirFilenames(dir string) map[string]bool {
	names := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return names
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			names[entry.Name()] = true
		}
	}
	return names
}

// collectFSFilenames returns a set of all filenames in an fs.FS root (non-recursive).
func collectFSFilenames(fsys fs.FS) map[string]bool {
	names := make(map[string]bool)
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return names
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			names[entry.Name()] = true
		}
	}
	return names
}

// removeManagedFiles removes files in dir whose names are in the managed set.
// Files not in the managed set (user-created) are preserved.
func removeManagedFiles(dir string, managed map[string]bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() && managed[entry.Name()] {
			os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
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

// clearInvocationState removes INVOCATION_STATE.yaml which becomes stale on rite switch.
// The file tracks borrowed components from the previous rite's invocations.
func (m *Materializer) clearInvocationState(claudeDir string) error {
	err := os.Remove(filepath.Join(claudeDir, "INVOCATION_STATE.yaml"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// materializeWorkflow copies workflow.yaml from the rite to .claude/ACTIVE_WORKFLOW.yaml.
// If the rite has no workflow.yaml, any existing ACTIVE_WORKFLOW.yaml is removed to
// prevent stale workflow data from a previous rite persisting after switch.
func (m *Materializer) materializeWorkflow(claudeDir string, resolved *ResolvedRite) error {
	dstPath := filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml")
	rFS := m.riteFS(resolved)
	content, err := fs.ReadFile(rFS, "workflow.yaml")
	if err != nil {
		// No workflow.yaml in this rite — remove any stale file from previous rite
		if removeErr := os.Remove(dstPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return removeErr
		}
		return nil
	}
	_, err = writeIfChanged(dstPath, content, 0644)
	return err
}

// writeActiveRite writes the ACTIVE_RITE marker file.
func (m *Materializer) writeActiveRite(riteName, claudeDir string) error {
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	_, err := writeIfChanged(activeRitePath, []byte(riteName+"\n"), 0644)
	return err
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
