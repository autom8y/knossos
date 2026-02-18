// Package materialize generates .claude/ directories from templates and rite manifests.
package materialize

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/inscription"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
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
	Soft       bool // CC-safe mode: only update agents + CLAUDE.md
}

// Result contains materialization outcome details.
type Result struct {
	Status           string   // Pipeline status: "success", "skipped", "minimal"
	OrphansDetected  []string // List of orphan agent files detected
	OrphanAction     string   // Action taken: "kept", "removed", "promoted"
	BackupPath       string   // Path to backup if orphans were removed
	Source           string   // Source type used: "project", "user", "knossos", "explicit"
	SourcePath       string   // Actual path resolved for rite source
	LegacyBackupPath string   // Path to legacy CLAUDE.md backup if migration occurred
	SoftMode         bool     // true if soft mode was used
	DeferredStages   []string // stages skipped in soft mode
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
	claudeDirOverride string // If set, materialize to this directory instead of .claude/
	embeddedAgents    fs.FS  // Embedded cross-rite agents (fallback for user scope)
	embeddedMena      fs.FS  // Embedded platform mena (fallback for user scope)
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

// WithEmbeddedAgents sets the embedded agents filesystem for user scope fallback.
func (m *Materializer) WithEmbeddedAgents(fsys fs.FS) *Materializer {
	m.embeddedAgents = fsys
	return m
}

// WithEmbeddedMena sets the embedded mena filesystem for user scope fallback.
func (m *Materializer) WithEmbeddedMena(fsys fs.FS) *Materializer {
	m.embeddedMena = fsys
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

	// Provenance: load previous manifest and detect divergence
	manifestPath := provenance.ManifestPath(claudeDir)
	prevManifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		log.Printf("Warning: failed to load provenance manifest, starting fresh: %v", err)
		prevManifest = &provenance.ProvenanceManifest{
			SchemaVersion: provenance.CurrentSchemaVersion,
			Entries:       make(map[string]*provenance.ProvenanceEntry),
		}
	}
	divergenceReport, err := provenance.DetectDivergence(prevManifest, nil, claudeDir)
	if err != nil {
		log.Printf("Warning: failed to detect provenance divergence: %v", err)
	}
	collector := provenance.NewCollector()

	// Generate rules from templates (if available)
	if err := m.materializeRules(claudeDir, nil, collector); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize rules", err)
	}

	// Generate minimal CLAUDE.md (no agents)
	legacyBackupPath, err := m.materializeMinimalCLAUDEmd(claudeDir, collector)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}
	result.LegacyBackupPath = legacyBackupPath

	// Generate settings.local.json if needed (no manifest in minimal mode)
	if err := m.materializeSettingsWithManifest(claudeDir, nil, collector); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
	}

	// Remove rite-specific state files (cross-cutting mode has no rite)
	os.Remove(filepath.Join(claudeDir, "ACTIVE_RITE"))
	os.Remove(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	os.Remove(filepath.Join(claudeDir, "INVOCATION_STATE.yaml"))

	// Provenance: merge and save manifest
	if err := m.saveProvenanceManifest(manifestPath, "", collector, divergenceReport, prevManifest); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to save provenance manifest", err)
	}

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
		orphans, err := m.detectOrphans(manifest, claudeDir, resolved)
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

	// Provenance: load previous manifest and detect divergence
	manifestPath := provenance.ManifestPath(claudeDir)
	prevManifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		log.Printf("Warning: failed to load provenance manifest, starting fresh: %v", err)
		prevManifest = &provenance.ProvenanceManifest{
			SchemaVersion: provenance.CurrentSchemaVersion,
			Entries:       make(map[string]*provenance.ProvenanceEntry),
		}
	}
	divergenceReport, err := provenance.DetectDivergence(prevManifest, nil, claudeDir)
	if err != nil {
		log.Printf("Warning: failed to detect provenance divergence: %v", err)
	}
	collector := provenance.NewCollector()

	// 2.5. Clear stale invocation state from previous rite
	if err := m.clearInvocationState(claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to clear invocation state", err)
	}

	// 3. Handle orphans before materializing agents
	orphans, err := m.detectOrphans(manifest, claudeDir, resolved)
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
	if err := m.materializeAgents(manifest, ritePath, claudeDir, resolved, collector); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize agents", err)
	}

	// 4.5. Add cross-rite agents (moirai, consultant, context-engineer) that don't conflict with rite agents
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := m.materializeCrossRiteAgents(agentsDir, resolved, collector, manifest); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize cross-rite agents", err)
	}

	// 5. Generate commands/ and skills/ directories from rite + shared + dependencies + mena
	if !opts.Soft {
		if err := m.materializeMena(manifest, claudeDir, resolved, collector); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize mena", err)
		}
	}

	// 6. Generate rules/ directory from templates/rules
	if !opts.Soft {
		if err := m.materializeRules(claudeDir, resolved, collector); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize rules", err)
		}
	}

	// 7. Generate CLAUDE.md from inscription system
	legacyBackupPath, err := m.materializeCLAUDEmd(manifest, claudeDir, resolved, collector)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}
	result.LegacyBackupPath = legacyBackupPath

	// 8. Generate or update settings.local.json with MCP servers from manifest
	if !opts.Soft {
		if err := m.materializeSettingsWithManifest(claudeDir, manifest, collector); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
		}
	}

	// 9. Track state in .claude/sync/state.json
	if err := m.trackState(manifest, activeRiteName); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to track state", err)
	}

	// 9.5. Copy workflow.yaml to ACTIVE_WORKFLOW.yaml
	if !opts.Soft {
		if err := m.materializeWorkflow(claudeDir, resolved, collector); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize workflow", err)
		}
	}

	// Populate soft mode result fields
	if opts.Soft {
		result.SoftMode = true
		result.DeferredStages = []string{"mena", "rules", "settings", "workflow"}
	}

	// 10. Write ACTIVE_RITE marker
	if err := m.writeActiveRite(activeRiteName, claudeDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write ACTIVE_RITE", err)
	}

	// Provenance: merge and save manifest
	if err := m.saveProvenanceManifest(manifestPath, activeRiteName, collector, divergenceReport, prevManifest); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to save provenance manifest", err)
	}

	return result, nil
}

// Sync performs a unified sync operation across rite and/or user scopes.
func (m *Materializer) Sync(opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{}

	// Normalize defaults
	if opts.Scope == "" {
		opts.Scope = ScopeAll
	}
	if !opts.Scope.IsValid() {
		return nil, fmt.Errorf("invalid scope: %q", opts.Scope)
	}
	if !opts.Resource.IsValid() {
		return nil, fmt.Errorf("invalid resource: %q", opts.Resource)
	}

	// Phase 1: Rite scope
	if opts.Scope == ScopeAll || opts.Scope == ScopeRite {
		riteResult, err := m.syncRiteScope(opts)
		if err != nil {
			if opts.Scope == ScopeRite {
				return nil, err
			}
			// scope=all: skip rite, continue to user
			result.RiteResult = &RiteScopeResult{Status: "skipped"}
		} else {
			result.RiteResult = riteResult
		}
	}

	// Phase 2: User scope
	if opts.Scope == ScopeAll || opts.Scope == ScopeUser {
		userResult, err := m.syncUserScope(opts)
		if err != nil {
			if opts.Scope == ScopeUser {
				return nil, err // hard fail only if explicitly user-only
			}
			// scope=all: log and skip, don't block rite results
			result.UserResult = &UserScopeResult{
				Status: "skipped",
				Errors: []UserResourceError{{Resource: ResourceAll, Err: err.Error()}},
			}
		} else {
			result.UserResult = userResult
		}
	}

	return result, nil
}

// syncRiteScope delegates to existing MaterializeWithOptions.
func (m *Materializer) syncRiteScope(opts SyncOptions) (*RiteScopeResult, error) {
	riteName := opts.RiteName
	if riteName == "" {
		// Try to read ACTIVE_RITE
		activeRitePath := filepath.Join(m.resolver.ClaudeDir(), "ACTIVE_RITE")
		data, err := os.ReadFile(activeRitePath)
		if err != nil {
			if opts.Scope == ScopeRite {
				return nil, fmt.Errorf("no ACTIVE_RITE found, specify --rite")
			}
			// scope=all with no rite: run minimal
			return m.syncRiteScopeMinimal(opts)
		}
		riteName = strings.TrimSpace(string(data))
	}

	legacyOpts := Options{
		DryRun:    opts.DryRun,
		RemoveAll: !opts.KeepOrphans,
		KeepAll:   opts.KeepOrphans,
		Soft:      opts.Soft,
	}

	legacyResult, err := m.MaterializeWithOptions(riteName, legacyOpts)
	if err != nil {
		return nil, err
	}

	return &RiteScopeResult{
		Status:           legacyResult.Status,
		RiteName:         riteName,
		Source:           legacyResult.Source,
		SourcePath:       legacyResult.SourcePath,
		OrphansDetected:  legacyResult.OrphansDetected,
		OrphanAction:     legacyResult.OrphanAction,
		BackupPath:       legacyResult.BackupPath,
		LegacyBackupPath: legacyResult.LegacyBackupPath,
		SoftMode:         legacyResult.SoftMode,
		DeferredStages:   legacyResult.DeferredStages,
	}, nil
}

// syncRiteScopeMinimal handles cross-cutting mode (no rite).
func (m *Materializer) syncRiteScopeMinimal(opts SyncOptions) (*RiteScopeResult, error) {
	legacyOpts := Options{DryRun: opts.DryRun, Minimal: true}
	legacyResult, err := m.MaterializeMinimal(legacyOpts)
	if err != nil {
		return nil, err
	}
	return &RiteScopeResult{
		Status: legacyResult.Status,
		Source: "minimal",
	}, nil
}

// syncUserScope is implemented in user_scope.go

// detectOrphans finds agent files that are not in the incoming rite's manifest.
// If a provenance manifest exists, uses manifest-based detection: files with
// owner=user or files not in the provenance manifest are orphans.
// Otherwise, falls back to rite manifest membership check (backward compatible).
func (m *Materializer) detectOrphans(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite) ([]string, error) {
	agentsDir := filepath.Join(claudeDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Build complete expected agent set: rite agents + cross-rite agents
	expectedAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		expectedAgents[agent.Name+".md"] = true
	}
	for _, agentName := range m.listCrossRiteAgents(resolved) {
		expectedAgents[agentName] = true
	}

	// Try loading provenance manifest for manifest-based detection
	manifestPath := provenance.ManifestPath(claudeDir)
	provenanceManifest, err := provenance.Load(manifestPath)

	// If provenance manifest exists, use manifest-based orphan detection
	if err == nil && provenanceManifest != nil {
		return m.detectOrphansFromProvenance(expectedAgents, claudeDir, provenanceManifest)
	}

	// Fallback: rite manifest membership check (backward compatible)
	return m.detectOrphansLegacy(expectedAgents, agentsDir)
}

// detectOrphansFromProvenance detects orphans using the provenance manifest.
// An agent file is an orphan if:
//   - It is NOT in the provenance manifest, OR
//   - It has owner=user in the provenance manifest, OR
//   - It is knossos-owned BUT not in the expected agents set (rite + cross-rite)
func (m *Materializer) detectOrphansFromProvenance(expectedAgents map[string]bool, claudeDir string, provenanceManifest *provenance.ProvenanceManifest) ([]string, error) {
	agentsDir := filepath.Join(claudeDir, "agents")

	orphans := []string{}
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only consider .md files
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// Construct relative path within .claude/
		relativePath := filepath.Join("agents", entry.Name())

		// Check if file is in provenance manifest
		provenanceEntry, exists := provenanceManifest.Entries[relativePath]

		// File not in provenance manifest -> orphan
		if !exists {
			orphans = append(orphans, entry.Name())
			continue
		}

		// File with owner=user -> orphan (user-created or previously promoted)
		if provenanceEntry.Owner == provenance.OwnerUser {
			orphans = append(orphans, entry.Name())
			continue
		}

		// Knossos-owned file not in current rite manifest -> orphan
		if provenanceEntry.Owner == provenance.OwnerKnossos && !expectedAgents[entry.Name()] {
			orphans = append(orphans, entry.Name())
			continue
		}

		// Knossos-owned entries still in rite manifest are NOT orphans
	}

	return orphans, nil
}

// detectOrphansLegacy uses manifest membership check.
// This is the fallback when no provenance manifest exists (backward compatible).
func (m *Materializer) detectOrphansLegacy(expectedAgents map[string]bool, agentsDir string) ([]string, error) {
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
// Also materializes cross-rite agents from top-level agents/ directory.
func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error {
	agentsDir := filepath.Join(claudeDir, "agents")

	// Ensure agents directory exists (selective — do NOT RemoveAll)
	if err := paths.EnsureDir(agentsDir); err != nil {
		return err
	}

	// Build managed agent set from manifest (rite agents)
	managedAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		managedAgents[agent.Name+".md"] = true
	}

	// Add cross-rite agents to managed set
	crossRiteAgents := m.listCrossRiteAgents(resolved)
	for _, agentName := range crossRiteAgents {
		managedAgents[agentName] = true
	}

	// NOTE: We intentionally do NOT pre-delete managed agents before rewriting.
	// writeIfChanged() handles overwrite-if-different atomically. Pre-deletion
	// causes CC's file watcher to see DELETE events for files that are immediately
	// recreated, which crashes/disrupts active Claude Code sessions.

	now := time.Now().UTC()

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
		// Copy agents and record provenance
		err = fs.WalkDir(agentsSub, ".", func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return walkErr
			}
			content, err := fs.ReadFile(agentsSub, path)
			if err != nil {
				return err
			}
			destPath := filepath.Join(agentsDir, path)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			written, err := writeIfChanged(destPath, content, 0644)
			if err != nil {
				return err
			}
			if written {
				relPath := "agents/" + path
				sourcePath := resolved.RitePath + "/agents/" + path
				collector.Record(relPath, &provenance.ProvenanceEntry{
					Owner:          provenance.OwnerKnossos,
					Scope: provenance.ScopeRite,
					SourcePath:     sourcePath,
					SourceType:     string(resolved.Source.Type),
					Checksum:       checksum.Bytes(content),
					LastSynced:     now,
				})
			}
			return nil
		})
		return err
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

		written, err := writeIfChanged(destPath, content, 0644)
		if err != nil {
			return err
		}

		// Record provenance after successful write
		if written {
			srcRelPath, _ := filepath.Rel(m.resolver.ProjectRoot(), path)
			collector.Record("agents/"+relPath, &provenance.ProvenanceEntry{
				Owner:          provenance.OwnerKnossos,
				Scope: provenance.ScopeRite,
				SourcePath:     srcRelPath,
				SourceType:     string(resolved.Source.Type),
				Checksum:       checksum.Bytes(content),
				LastSynced:     now,
			})
		}

		return nil
	})
}

// listCrossRiteAgents returns a list of cross-rite agent filenames from top-level agents/.
// Cross-rite agents are agents that should be available regardless of active rite.
func (m *Materializer) listCrossRiteAgents(resolved *ResolvedRite) []string {
	var agents []string

	// For embedded sources, read from embedded FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil {
		entries, err := fs.ReadDir(m.sourceResolver.embeddedFS, "agents")
		if err != nil {
			return agents // No cross-rite agents in embedded FS
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				agents = append(agents, entry.Name())
			}
		}
		return agents
	}

	// For filesystem sources, read from project root agents/
	projectRoot := m.resolver.ProjectRoot()
	crossRiteDir := filepath.Join(projectRoot, "agents")
	entries, err := os.ReadDir(crossRiteDir)
	if err != nil {
		return agents // No cross-rite agents directory
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			agents = append(agents, entry.Name())
		}
	}
	return agents
}

// materializeCrossRiteAgents copies cross-rite agents from top-level agents/ to .claude/agents/.
// Cross-rite agents are available in all rites but do NOT override rite-scoped agents of the same name.
func (m *Materializer) materializeCrossRiteAgents(agentsDir string, resolved *ResolvedRite, collector provenance.Collector, manifest *RiteManifest) error {
	now := time.Now().UTC()

	// Build set of rite agent names (these take priority)
	riteAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		riteAgents[agent.Name+".md"] = true
	}

	// For embedded sources, copy from embedded FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil {
		entries, err := fs.ReadDir(m.sourceResolver.embeddedFS, "agents")
		if err != nil {
			return nil // No cross-rite agents in embedded FS
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}
			// Skip if rite already has an agent with this name
			if riteAgents[entry.Name()] {
				continue
			}
			content, err := fs.ReadFile(m.sourceResolver.embeddedFS, "agents/"+entry.Name())
			if err != nil {
				return err
			}
			destPath := filepath.Join(agentsDir, entry.Name())
			written, err := writeIfChanged(destPath, content, 0644)
			if err != nil {
				return err
			}
			if written {
				collector.Record("agents/"+entry.Name(), &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeRite,
					SourcePath: "agents/" + entry.Name(),
					SourceType: string(resolved.Source.Type),
					Checksum:   checksum.Bytes(content),
					LastSynced: now,
				})
			}
		}
		return nil
	}

	// For filesystem sources, copy from project root agents/
	projectRoot := m.resolver.ProjectRoot()
	crossRiteDir := filepath.Join(projectRoot, "agents")
	entries, err := os.ReadDir(crossRiteDir)
	if err != nil {
		return nil // No cross-rite agents directory
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		// Skip if rite already has an agent with this name
		if riteAgents[entry.Name()] {
			continue
		}

		srcPath := filepath.Join(crossRiteDir, entry.Name())
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		destPath := filepath.Join(agentsDir, entry.Name())
		written, err := writeIfChanged(destPath, content, 0644)
		if err != nil {
			return err
		}

		if written {
			srcRelPath, _ := filepath.Rel(projectRoot, srcPath)
			collector.Record("agents/"+entry.Name(), &provenance.ProvenanceEntry{
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeRite,
				SourcePath: srcRelPath,
				SourceType: "project",
				Checksum:   checksum.Bytes(content),
				LastSynced: now,
			})
		}
	}

	return nil
}

// materializeMena copies mena files to .claude/commands/ or .claude/skills/
// based on the filename convention (.dro.md for dromena, .lego.md for legomena).
// Sources: mena/, rites/{rite}/mena/, rites/shared/mena/
// Priority order (later sources override earlier): mena < shared < dependencies < current rite
//
// This method builds the source list and delegates to SyncMena() for the
// actual collection, routing, extension stripping, and file copying.
func (m *Materializer) materializeMena(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error {
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

	// Delegate to SyncMena with destructive mode and provenance collector
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		Collector:         collector,
		ProjectRoot:       m.resolver.ProjectRoot(),
	}

	_, err := SyncMena(sources, opts)
	return err
}

// collectMenaEntriesDir recursively collects mena entries from a filesystem directory.
// Leaf directories (containing INDEX files) are collected for routing.
// Standalone files in grouping directories are collected separately.
func collectMenaEntriesDir(dirPath string, prefix string, collected map[string]menaCollectedEntry, standalones map[string]menaStandaloneFile, srcIdx int) error {
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
					source:      MenaSource{Path: childPath},
					name:        name,
					sourceIndex: srcIdx,
				}
			} else {
				if err := collectMenaEntriesDir(childPath, name, collected, standalones, srcIdx); err != nil {
					return err
				}
			}
		} else {
			// Standalone file in a grouping directory
			standalones[name] = menaStandaloneFile{
				srcPath:     filepath.Join(dirPath, entry.Name()),
				relPath:     name,
				sourceIndex: srcIdx,
			}
		}
	}
	return nil
}

// collectMenaEntriesFS recursively collects mena entries from an embedded filesystem.
func collectMenaEntriesFS(fsys fs.FS, fsysPath string, prefix string, collected map[string]menaCollectedEntry, srcIdx int) {
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
				name:        name,
				sourceIndex: srcIdx,
			}
		} else {
			collectMenaEntriesFS(fsys, childPath, name, collected, srcIdx)
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
func (m *Materializer) materializeRules(claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error {
	rulesDir := filepath.Join(claudeDir, "rules")
	if err := paths.EnsureDir(rulesDir); err != nil {
		return err
	}

	now := time.Now().UTC()
	projectRoot := m.resolver.ProjectRoot()

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

	// Remove only STALE knossos-managed rules: files that match a known template name
	// but are NOT in the current rite's template set. Do NOT pre-delete rules that will
	// be rewritten — writeIfChanged() handles atomic overwrite. Pre-deletion causes
	// CC's file watcher to see DELETE events that crash active sessions.
	if existingRules, err := os.ReadDir(rulesDir); err == nil {
		for _, entry := range existingRules {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			if allTemplateNames[entry.Name()] && templateRules[entry.Name()] == nil {
				os.Remove(filepath.Join(rulesDir, entry.Name()))
			}
		}
	}

	// Write current template rules and record provenance
	for name, content := range templateRules {
		dstPath := filepath.Join(rulesDir, name)
		written, err := writeIfChanged(dstPath, content, 0644)
		if err != nil {
			return err
		}
		if written {
			sourcePath := filepath.Join(m.templatesDir, "rules", name)
			srcRelPath, _ := filepath.Rel(projectRoot, sourcePath)
			collector.Record("rules/"+name, &provenance.ProvenanceEntry{
				Owner:          provenance.OwnerKnossos,
				Scope: provenance.ScopeRite,
				SourcePath:     srcRelPath,
				SourceType:     "template",
				Checksum:       checksum.Bytes(content),
				LastSynced:     now,
			})
		}
	}

	return nil
}

// materializeCLAUDEmd generates CLAUDE.md using the inscription system.
// Delegates to inscription.SyncCLAUDEmd for the core merge/write logic,
// then records provenance for the written file.
// Returns the path to legacy backup if migration occurred, or empty string if no backup.
func (m *Materializer) materializeCLAUDEmd(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, collector provenance.Collector) (string, error) {
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

	// Resolve template source: embedded FS or filesystem directory
	var templateFS fs.FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		templateFS = m.templatesFS(resolved)
	}

	// Delegate to canonical SyncCLAUDEmd
	result, err := inscription.SyncCLAUDEmd(inscription.CLAUDEmdSyncOptions{
		ClaudeDir:      claudeDir,
		RenderCtx:      renderCtx,
		ActiveRite:     manifest.Name,
		TemplateDir:    m.templatesDir,
		TemplateFS:     templateFS,
		UpdateManifest: true,
	})
	if err != nil {
		return "", err
	}

	// Record provenance after successful write (materialization-specific concern)
	if result.Written {
		now := time.Now().UTC()
		srcRelPath := "(generated)"
		if m.templatesDir != "" {
			projectRoot := m.resolver.ProjectRoot()
			sourcePath := filepath.Join(m.templatesDir, "CLAUDE.md.tpl")
			if rel, err := filepath.Rel(projectRoot, sourcePath); err == nil && rel != "" {
				srcRelPath = rel
			}
		}
		collector.Record("CLAUDE.md", &provenance.ProvenanceEntry{
			Owner:      provenance.OwnerKnossos,
			Scope:      provenance.ScopeRite,
			SourcePath: srcRelPath,
			SourceType: "template",
			Checksum:   checksum.Content(result.MergeResult.Content),
			LastSynced: now,
		})
	}

	return result.LegacyBackupPath, nil
}

// materializeMinimalCLAUDEmd generates CLAUDE.md for cross-cutting mode (no agents).
// Delegates to inscription.SyncCLAUDEmd without manifest updates.
func (m *Materializer) materializeMinimalCLAUDEmd(claudeDir string, collector provenance.Collector) (string, error) {
	renderCtx := &inscription.RenderContext{
		ActiveRite:  "",
		AgentCount:  0,
		Agents:      []inscription.AgentInfo{},
		KnossosVars: make(map[string]string),
		ProjectRoot: m.resolver.ProjectRoot(),
	}

	result, err := inscription.SyncCLAUDEmd(inscription.CLAUDEmdSyncOptions{
		ClaudeDir:      claudeDir,
		RenderCtx:      renderCtx,
		TemplateDir:    m.templatesDir,
		UpdateManifest: false,
	})
	if err != nil {
		return "", err
	}

	return result.LegacyBackupPath, nil
}

// materializeSettingsWithManifest generates or updates settings.local.json.
// If manifest has MCP servers, merges them into existing settings.
// Loads hooks.yaml and merges hook registrations into settings.
// If no manifest or no MCP servers, creates minimal settings if needed.
func (m *Materializer) materializeSettingsWithManifest(claudeDir string, manifest *RiteManifest, collector provenance.Collector) error {
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
	err = saveSettings(settingsPath, existingSettings)
	if err != nil {
		return err
	}

	// Record provenance after successful write
	hash, err := checksum.File(settingsPath)
	if err == nil && hash != "" {
		now := time.Now().UTC()
		collector.Record("settings.local.json", &provenance.ProvenanceEntry{
			Owner:          provenance.OwnerKnossos,
			Scope: provenance.ScopeRite,
			SourcePath:     "(generated)",
			SourceType:     "template",
			Checksum:       hash,
			LastSynced:     now,
		})
	}

	return nil
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
		state, err = stateManager.Initialize()
		if err != nil {
			return err
		}
	}

	// Update active rite and last sync time
	state.ActiveRite = activeRiteName
	state.LastSync = time.Now().UTC()
	err = stateManager.Save(state)
	if err != nil {
		return err
	}

	return nil
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
func (m *Materializer) materializeWorkflow(claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error {
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
	written, err := writeIfChanged(dstPath, content, 0644)
	if err != nil {
		return err
	}

	// Record provenance after successful write
	if written {
		now := time.Now().UTC()
		projectRoot := m.resolver.ProjectRoot()
		sourcePath := resolved.RitePath + "/workflow.yaml"
		srcRelPath, _ := filepath.Rel(projectRoot, sourcePath)
		collector.Record("ACTIVE_WORKFLOW.yaml", &provenance.ProvenanceEntry{
			Owner:          provenance.OwnerKnossos,
			Scope: provenance.ScopeRite,
			SourcePath:     srcRelPath,
			SourceType:     string(resolved.Source.Type),
			Checksum:       checksum.Bytes(content),
			LastSynced:     now,
		})
	}

	return nil
}

// writeActiveRite writes the ACTIVE_RITE marker file.
func (m *Materializer) writeActiveRite(riteName, claudeDir string) error {
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	content := []byte(riteName + "\n")
	_, err := writeIfChanged(activeRitePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
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

// saveProvenanceManifest merges collector entries with divergence report and previous manifest,
// then writes the final manifest to disk. Implements the merge algorithm from TDD Section 6.
func (m *Materializer) saveProvenanceManifest(
	manifestPath string,
	activeRite string,
	collector provenance.Collector,
	divergenceReport *provenance.DivergenceReport,
	prevManifest *provenance.ProvenanceManifest,
) error {
	finalEntries := make(map[string]*provenance.ProvenanceEntry)

	// Step 0: Carry forward knossos entries from previous manifest that still exist on disk
	// but weren't re-written this sync (idempotency - files that didn't change)
	if prevManifest != nil {
		claudeDir := filepath.Dir(manifestPath)
		for path, entry := range prevManifest.Entries {
			if entry.Owner == provenance.OwnerKnossos {
				// Check if file/directory still exists on disk (not removed)
				fullPath := filepath.Join(claudeDir, path)
				// For directory entries (mena), remove trailing slash before stat
				if strings.HasSuffix(path, "/") {
					fullPath = strings.TrimSuffix(fullPath, "/")
				}
				if _, err := os.Stat(fullPath); err == nil {
					// File/directory exists, carry forward entry (will be overwritten if rewritten in Step 2)
					finalEntries[path] = entry
				}
			}
		}
	}

	// Step 1: Carry forward promoted entries (user-owned + unknown) from divergence detection
	// Skip entries with empty checksums (deleted files) as they fail validation
	if divergenceReport != nil {
		for path, entry := range divergenceReport.Promoted {
			if entry.Checksum != "" {
				finalEntries[path] = entry
			}
		}
		for path, entry := range divergenceReport.CarriedForward {
			if entry.Checksum != "" {
				finalEntries[path] = entry
			}
		}
	}

	// Step 2: Layer current sync entries on top
	// Pipeline-written files take precedence unless path was promoted to user-owned in Step 1
	for path, entry := range collector.Entries() {
		if existing, ok := finalEntries[path]; ok {
			if existing.Owner == provenance.OwnerUser {
				// User promoted this file via divergence detection.
				// Do NOT overwrite with the pipeline entry.
				continue
			}
		}
		finalEntries[path] = entry
	}

	// Step 3: Resolve unknown entries from previous manifest
	// Files with owner:untracked that the pipeline did NOT write this sync are promoted to owner:user
	if prevManifest != nil {
		for path, entry := range prevManifest.Entries {
			if entry.Owner == provenance.OwnerUntracked {
				if _, writtenThisSync := collector.Entries()[path]; !writtenThisSync {
					promotedEntry := *entry
					promotedEntry.Owner = provenance.OwnerUser
					if _, alreadyInFinal := finalEntries[path]; !alreadyInFinal {
						finalEntries[path] = &promotedEntry
					}
				}
			}
		}
	}

	// Build final manifest
	finalManifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    activeRite,
		Entries:       finalEntries,
	}

	// Save manifest
	return provenance.Save(manifestPath, finalManifest)
}
