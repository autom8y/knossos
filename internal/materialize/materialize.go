// Package materialize generates .claude/ directories from templates and rite manifests.
package materialize

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/materialize/compiler"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/autom8y/knossos/internal/registry"
)

// Options configures materialization behavior.
type Options struct {
	Force             bool // Skip idempotency check; proceed even if already on this rite
	DryRun            bool // Preview changes without applying
	RemoveAll         bool // Remove all orphan agents (with backup)
	KeepAll           bool // Preserve all orphan agents (default)
	PromoteAll        bool // Move orphan agents to user-level
	Minimal           bool // Generate base infrastructure only (no rite/agents/skills)
	Soft              bool // CC-safe mode: only update agents + CLAUDE.md
	OverwriteDiverged bool // Allow overwriting user-owned mena entries on flat-name collision
	ElCheapo          bool // Force haiku model override on all agents (ephemeral)
	Channel           string
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
	ElCheapoMode     bool     // true if el-cheapo model override was applied
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
	Name          string                    `yaml:"name"`
	Version       string                    `yaml:"version"`
	Description   string                    `yaml:"description"`
	EntryAgent    string                    `yaml:"entry_agent"`
	Agents        []Agent                   `yaml:"agents"`
	Dromena       []string                  `yaml:"dromena"`  // Invokable commands (project to .claude/commands/)
	Legomena      []string                  `yaml:"legomena"` // Reference knowledge (project to .claude/skills/)
	Commands      []string                  `yaml:"commands"` // Backward compat: populates from dromena+legomena if empty
	Skills        []string                  `yaml:"skills"`   // Deprecated: use Legomena instead
	Hooks         []string                  `yaml:"hooks"`
	Dependencies  []string                  `yaml:"dependencies"`
	MCPServers    []MCPServer               `yaml:"mcp_servers,omitempty"` // MCP server declarations
	HookDefaults  *HookDefaults             `yaml:"hook_defaults,omitempty"`
	AgentDefaults map[string]any            `yaml:"agent_defaults,omitempty"` // Manifest-level defaults merged into agent frontmatter during sync
	SkillPolicies []SkillPolicy             `yaml:"skill_policies,omitempty"` // Capability-driven skill wiring rules evaluated per-agent during sync
	ArchetypeData map[string]map[string]any `yaml:"archetype_data,omitempty"` // Per-archetype template data keyed by archetype name
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
	Name      string `yaml:"name"`
	Role      string `yaml:"role"`
	Archetype string `yaml:"archetype,omitempty"` // Template name: "orchestrator", etc.
}

// Materializer handles .claude/ directory generation.
type Materializer struct {
	resolver          *paths.Resolver
	sourceResolver    *SourceResolver
	explicitSource    string // Optional explicit source from --source flag
	templatesDir      string
	embeddedTemplates fs.FS  // Embedded templates filesystem
	channelDirOverride string // If set, materialize to this directory instead of .claude/
	embeddedAgents      fs.FS   // Embedded cross-rite agents (fallback for user scope)
	embeddedMena        fs.FS   // Embedded platform mena (fallback for user scope)
	embeddedProcessions fs.FS   // Embedded procession templates (fallback for resolution)
	userChannelDir      string  // If set, user-scope sync writes here instead of paths.UserChannelDir("claude")
	xdgDataDir          string  // If set, used for XDG mena path instead of config.XDGDataDir()
	knossosHome         string  // If set, used for archetype/procession resolution instead of config.KnossosHome()
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

// NewMaterializerWithSourceResolver creates a materializer with an explicit source resolver.
// This enables test injection of tier paths without global state mutation.
func NewMaterializerWithSourceResolver(resolver *paths.Resolver, sr *SourceResolver) *Materializer {
	return &Materializer{
		resolver:       resolver,
		sourceResolver: sr,
		templatesDir:   filepath.Join(resolver.ProjectRoot(), "templates"),
	}
}

func compilerForChannel(channel string) compiler.ChannelCompiler {
	if channel == "gemini" {
		return &compiler.GeminiCompiler{}
	}
	// Claude and empty channel require no compilation pass: agent files are written
	// as-is after transformAgentContent(). Returning nil skips the compile step.
	return nil
}

// WithSourceResolver replaces the materializer's source resolver.
// Enables test injection without global state mutation. Returns the receiver.
func (m *Materializer) WithSourceResolver(sr *SourceResolver) *Materializer {
	m.sourceResolver = sr
	return m
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

// WithEmbeddedProcessions sets the embedded procession templates filesystem.
// Used during sync to resolve procession templates when not available on disk.
func (m *Materializer) WithEmbeddedProcessions(fsys fs.FS) *Materializer {
	m.embeddedProcessions = fsys
	return m
}

// getChannelDir returns the target .claude/ directory, respecting any override.
func (m *Materializer) getChannelDir() string {
	if m.channelDirOverride != "" {
		return m.channelDirOverride
	}
	return m.resolver.ClaudeDir()
}

// getKnossosDir returns the .knossos/ directory for the project.
// The provenance manifest lives here (PROVENANCE_MANIFEST.yaml).
func (m *Materializer) getKnossosDir() string {
	return m.resolver.KnossosDir()
}

// riteFS returns a filesystem rooted at the rite's directory.
// For embedded sources, returns a sub-FS of the embedded rites.
// For filesystem sources, returns os.DirFS rooted at the rite path.
func (m *Materializer) riteFS(resolved *ResolvedRite) fs.FS {
	if resolved.Source.Type == SourceEmbedded && m.sourceResolver.EmbeddedFS != nil {
		sub, err := fs.Sub(m.sourceResolver.EmbeddedFS, resolved.RitePath)
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
		_, err = fileutil.WriteIfChanged(destPath, content, 0644)
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

	// Save existing override and restore it when done to prevent mutation leaking
	originalOverride := m.channelDirOverride
	defer func() { m.channelDirOverride = originalOverride }()

	if opts.Channel == "gemini" {
		m.channelDirOverride = filepath.Join(filepath.Dir(m.resolver.ClaudeDir()), ".gemini")
	}

	channelDir := m.getChannelDir()

	// Dry-run: just return success
	if opts.DryRun {
		return result, nil
	}

	// Ensure channel directory exists
	if err := paths.EnsureDir(channelDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create .claude directory", err)
	}

	// Provenance: load previous manifest and detect divergence.
	// LoadOrBootstrap returns an empty manifest only on file-not-found; all other errors
	// (parse failure, schema validation) propagate and abort the pipeline. A corrupted
	// provenance manifest must be fixed or removed manually -- silent bootstrapping would
	// mask data corruption and defeat the purpose of the manifest.
	knossosDir := m.getKnossosDir()
	if err := paths.EnsureDir(knossosDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create .knossos directory", err)
	}
	manifestPath := provenance.ManifestPathForChannel(knossosDir, opts.Channel)
	prevManifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load provenance manifest", err)
	}
	divergenceReport, err := provenance.DetectDivergence(prevManifest, nil, channelDir)
	if err != nil {
		slog.Warn("failed to detect provenance divergence", "error", err)
	}
	collector := provenance.NewCollector()

	// Remove stale settings.json created by the deleted writeDefaultSettings() function.
	// Must run after prevManifest is loaded (needed for the provenance gate) and before
	// materializeSettingsWithManifest() writes settings.local.json.
	m.cleanupStaleBlanketSettings(channelDir, prevManifest)

	// Generate rules from templates (if available)
	if err := m.materializeRules(channelDir, nil, collector, opts.Channel); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize rules", err)
	}

	comp := compilerForChannel(opts.Channel)

	// Generate minimal CLAUDE.md (no agents)
	legacyBackupPath, err := m.materializeMinimalInscription(channelDir, collector, opts.Channel, comp)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}
	result.LegacyBackupPath = legacyBackupPath

	// Generate settings.local.json if needed (no manifest in minimal mode)
	if err := m.materializeSettingsWithManifest(channelDir, nil, collector, opts.Channel); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
	}

	// Project platform mena + shared rite mena so cross-cutting mode still
	// has core features (/know, /radar, /research, etc.).
	if err := m.materializeMinimalMena(channelDir, collector, opts.OverwriteDiverged, opts.Channel, comp); err != nil {
		slog.Warn("failed to materialize mena in minimal mode", "error", err)
		// Non-fatal: mena is a best-effort enhancement in minimal mode
	}

	// Remove rite-specific state files (cross-cutting mode has no rite)
	_ = os.Remove(filepath.Join(knossosDir, "ACTIVE_RITE"))
	_ = os.Remove(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	_ = os.Remove(filepath.Join(channelDir, "INVOCATION_STATE.yaml"))

	// Provenance: merge and save manifest
	finalManifest, err := m.saveProvenanceManifest(manifestPath, channelDir, "", collector, divergenceReport, prevManifest, opts.OverwriteDiverged)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to save provenance manifest", err)
	}

	// Generate channel .gitignore from provenance (non-fatal)
	if _, gitErr := generateChannelGitignore(channelDir, finalManifest); gitErr != nil {
		slog.Warn("failed to generate channel .gitignore", "channel_dir", channelDir, "error", gitErr)
	}

	// Untrack knossos-owned files already in git index (non-fatal, best-effort)
	untrackKnossosFiles(m.resolver.ProjectRoot(), channelDir, finalManifest)

	return result, nil
}

// MaterializeWithOptions generates the .claude/ directory with configurable orphan handling.
func (m *Materializer) MaterializeWithOptions(activeRiteName string, opts Options) (*Result, error) {
	result := &Result{
		Status:          "success",
		OrphansDetected: []string{},
		OrphanAction:    "kept",
	}

	// Save existing override and restore it when done to prevent mutation leaking
	originalOverride := m.channelDirOverride
	defer func() { m.channelDirOverride = originalOverride }()

	if opts.Channel == "gemini" {
		m.channelDirOverride = filepath.Join(filepath.Dir(m.resolver.ClaudeDir()), ".gemini")
	}

	channelDir := m.getChannelDir()

	// Remove el-cheapo marker on normal sync (revert path)
	if !opts.ElCheapo {
		knossosDir := filepath.Join(filepath.Dir(channelDir), ".knossos")
		_ = os.Remove(filepath.Join(knossosDir, ".el-cheapo-active"))
	}

	// Note: the skip guard (skip-if-same-rite) was removed. The pipeline is safe
	// to always run: selective write preserves user content, and fileutil.WriteIfChanged()
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

	// Validate rite references: warn about stale agents/mena entries in the manifest.
	// Non-blocking: validation errors (missing manifest, parse failure) are silently skipped.
	ritesBase := filepath.Dir(ritePath)
	platformMenaDir := m.getMenaDir()
	if warnings, err := registry.ValidateRiteReferences(ritePath, ritesBase, platformMenaDir); err == nil {
		for _, w := range warnings {
			slog.Warn("stale rite reference", "file", w.File, "message", w.Message, "ref", w.RefName)
		}
	}

	// Update templates dir if resolved from different source
	if resolved.TemplatesDir != "" {
		m.templatesDir = resolved.TemplatesDir
	}

	// Load the rite manifest from resolved path
	manifest, err := m.loadRiteManifest(ritePath, resolved)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to load rite manifest", err)
	}

	// Compute model override from options (needed early for CLAUDE.md pre-validation)
	modelOverride := ""
	if opts.ElCheapo {
		modelOverride = "haiku"
	}

	comp := compilerForChannel(opts.Channel)

	// Dry-run: just detect orphans and return
	if opts.DryRun {
		orphans, err := m.detectOrphans(manifest, channelDir, resolved, opts.Channel)
		if err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to detect orphans", err)
		}
		result.OrphansDetected = orphans
		return result, nil
	}

	// Pre-validate CLAUDE.md generation before any disk writes.
	// Template rendering is the most failure-prone step. Validating it first
	// prevents partial state where agents are on disk but CLAUDE.md is stale.
	if err := m.prevalidateInscription(manifest, channelDir, resolved, modelOverride, opts.Channel); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "CLAUDE.md pre-validation failed (no files written)", err)
	}

	// 2. Ensure .claude/ and .knossos/ directories exist
	if err := paths.EnsureDir(channelDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create .claude directory", err)
	}
	knossosDir := m.getKnossosDir()
	if err := paths.EnsureDir(knossosDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create .knossos directory", err)
	}

	// Provenance: load previous manifest and detect divergence.
	// LoadOrBootstrap returns an empty manifest only on file-not-found; all other errors
	// (parse failure, schema validation) propagate and abort the pipeline. A corrupted
	// provenance manifest must be fixed or removed manually -- silent bootstrapping would
	// mask data corruption and defeat the purpose of the manifest.
	manifestPath := provenance.ManifestPathForChannel(knossosDir, opts.Channel)
	prevManifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load provenance manifest", err)
	}
	divergenceReport, err := provenance.DetectDivergence(prevManifest, nil, channelDir)
	if err != nil {
		slog.Warn("failed to detect provenance divergence", "error", err)
	}
	collector := provenance.NewCollector()

	// 2.5. Clear stale invocation state from previous rite
	if err := m.clearInvocationState(channelDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to clear invocation state", err)
	}

	// 2.6. Remove stale settings.json created by the deleted writeDefaultSettings() function.
	// Must run after prevManifest is loaded (needed for the provenance gate) and before
	// materializeSettingsWithManifest() writes settings.local.json.
	if !opts.DryRun {
		m.cleanupStaleBlanketSettings(channelDir, prevManifest)
	}

	// 3. Handle orphans before materializing agents
	orphans, err := m.detectOrphans(manifest, channelDir, resolved, opts.Channel)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to detect orphans", err)
	}
	result.OrphansDetected = orphans

	if len(orphans) > 0 {
		if opts.RemoveAll {
			backupPath, err := m.backupAndRemoveOrphans(orphans, channelDir, knossosDir)
			if err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to remove orphans", err)
			}
			result.OrphanAction = "removed"
			result.BackupPath = backupPath
		} else if opts.PromoteAll {
			if err := m.promoteOrphans(orphans, channelDir); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to promote orphans", err)
			}
			result.OrphanAction = "promoted"
		}
		// KeepAll: do nothing (orphans remain)
	}

	// 3.5. Resolve hook defaults: shared → rite cascade
	sharedHookDefaults := m.loadSharedHookDefaults(resolved)
	mergedWriteGuardDefaults := ResolveHookDefaults(sharedHookDefaults, manifest.HookDefaults)

	// 3.6. Resolve skill policies: shared → rite cascade
	sharedSkillPolicies := m.loadSharedSkillPolicies(resolved)
	mergedSkillPolicies := MergeSkillPolicies(sharedSkillPolicies, manifest.SkillPolicies)

	// 4. Generate agents/ directory from rite
	if err := m.materializeAgents(manifest, ritePath, channelDir, resolved, collector, mergedWriteGuardDefaults, mergedSkillPolicies, modelOverride, opts.Channel, comp); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize agents", err)
	}

	// 5. Generate commands/ and skills/ directories from rite + shared + dependencies + mena
	if !opts.Soft {
		if err := m.materializeMena(manifest, channelDir, resolved, collector, opts.OverwriteDiverged, opts.Channel, comp); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize mena", err)
		}
	}

	// 6. Generate rules/ directory from templates/rules
	if !opts.Soft {
		if err := m.materializeRules(channelDir, resolved, collector, opts.Channel); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize rules", err)
		}
	}

	// 7. Generate CLAUDE.md from inscription system
	legacyBackupPath, err := m.materializeInscription(manifest, channelDir, resolved, collector, modelOverride, opts.Channel, comp)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize CLAUDE.md", err)
	}
	result.LegacyBackupPath = legacyBackupPath

	// 8. Generate or update settings.local.json (hooks only; MCP servers moved to .mcp.json per SCAR-028)
	if !opts.Soft {
		if err := m.materializeSettingsWithManifest(channelDir, manifest, collector, opts.Channel); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize settings", err)
		}
	}

	// 8.1. Write MCP servers to .mcp.json at project root (SCAR-028)
	if !opts.Soft {
		projectRoot := filepath.Dir(channelDir)
		if err := m.materializeMcpJson(projectRoot, manifest, collector, opts.Channel); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize .mcp.json", err)
		}
	}

	// 8.5. El-cheapo mode: inject model override and revert hook into settings
	if opts.ElCheapo {
		if err := m.injectElCheapoSettings(channelDir); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to inject el-cheapo settings", err)
		}
	}

	// 9. Track state in .knossos/sync/state.json
	if err := m.trackState(manifest, activeRiteName); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to track state", err)
	}

	// 9.5. Copy workflow.yaml to ACTIVE_WORKFLOW.yaml
	if !opts.Soft {
		if err := m.materializeWorkflow(knossosDir, resolved, collector, opts.Channel); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to materialize workflow", err)
		}
	}

	// Populate soft mode result fields
	if opts.Soft {
		result.SoftMode = true
		result.DeferredStages = []string{"mena", "rules", "settings", "workflow"}
	}

	// Populate el-cheapo mode result field
	if opts.ElCheapo {
		result.ElCheapoMode = true
	}

	// 10. Write ACTIVE_RITE marker
	if err := m.writeActiveRite(activeRiteName); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write ACTIVE_RITE", err)
	}

	// Provenance: merge and save manifest
	finalManifest, err := m.saveProvenanceManifest(manifestPath, channelDir, activeRiteName, collector, divergenceReport, prevManifest, opts.OverwriteDiverged)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to save provenance manifest", err)
	}

	// Generate channel .gitignore from provenance (non-fatal)
	if _, gitErr := generateChannelGitignore(channelDir, finalManifest); gitErr != nil {
		slog.Warn("failed to generate channel .gitignore", "channel_dir", channelDir, "error", gitErr)
	}

	// Untrack knossos-owned files already in git index (non-fatal, best-effort)
	untrackKnossosFiles(m.resolver.ProjectRoot(), channelDir, finalManifest)

	return result, nil
}

// ensureProjectDirs creates the minimum directory structure required for sync
// to function. It is idempotent and zero-cost when the directories already exist.
// This covers worktrees, fresh clones, and any scenario where gitignored
// directories are absent. Errors are intentionally ignored: if a directory
// cannot be created, the subsequent sync steps will fail with actionable errors.
func (m *Materializer) ensureProjectDirs() {
	dirs := []string{
		m.resolver.ClaudeDir(),   // channel dir
		m.resolver.SessionsDir(), // .sos/sessions/ (implies .sos/)
		m.resolver.KnossosDir(),  // .knossos/
	}
	for _, d := range dirs {
		_ = os.MkdirAll(d, 0755)
	}
}

// Sync performs a unified sync operation across rite and/or user scopes.
func (m *Materializer) Sync(opts SyncOptions) (*SyncResult, error) {
	// Pre-flight: ensure framework directories exist before any scope dispatch.
	// This is idempotent and handles worktrees, fresh clones, and any env where
	// the gitignored directories (.claude/, .sos/, .knossos/) are absent.
	m.ensureProjectDirs()

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
			// scope=all: surface error but continue to user scope
			result.RiteResult = &RiteScopeResult{
				Status: "error",
				Error:  err.Error(),
			}
		} else {
			result.RiteResult = riteResult
		}
	}

	// Phase 1.5: Org scope
	if opts.Scope == ScopeAll || opts.Scope == ScopeOrg {
		orgResult, err := m.syncOrgScope(opts)
		if err != nil {
			if opts.Scope == ScopeOrg {
				return nil, err
			}
			// scope=all: log and skip, don't block other results
			result.OrgResult = &OrgScopeResult{
				Status: "error",
				Error:  err.Error(),
			}
		} else {
			result.OrgResult = orgResult
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
	// Dispatch multi-channel: resolve "all" into per-channel iterations.
	// The value "all" is resolved ONLY here -- MaterializeWithOptions never sees it.
	if opts.Channel == "all" {
		return m.syncRiteScopeAllChannels(opts)
	}

	riteName := opts.RiteName

	// Always read previous ACTIVE_RITE for rite-switch detection
	previousRite := m.resolver.ReadActiveRite()

	if riteName == "" {
		if previousRite == "" {
			// Before falling to minimal, check if we are in a linked git worktree
			// and inherit the rite from the main worktree's ACTIVE_RITE.
			if isGitWorktree(m.resolver.ProjectRoot()) {
				if mainDir, err := getMainWorktreeDir(m.resolver.ProjectRoot()); err == nil {
					if inherited := inheritRiteFromMainWorktree(mainDir); inherited != "" {
						riteName = inherited
					}
				}
			}
		}
		if riteName == "" && previousRite == "" {
			if opts.Scope == ScopeRite {
				return nil, fmt.Errorf("no ACTIVE_RITE found, specify --rite")
			}
			// scope=all with no rite: run minimal
			return m.syncRiteScopeMinimal(opts)
		}
		if riteName == "" {
			riteName = previousRite
		}
	}

	legacyOpts := Options{
		DryRun:            opts.DryRun,
		RemoveAll:         !opts.KeepOrphans,
		KeepAll:           opts.KeepOrphans,
		Soft:              opts.Soft,
		OverwriteDiverged: opts.OverwriteDiverged,
		ElCheapo:          opts.ElCheapo,
		Channel:           opts.Channel,
	}

	legacyResult, err := m.MaterializeWithOptions(riteName, legacyOpts)
	if err != nil {
		return nil, err
	}

	result := &RiteScopeResult{
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
		ElCheapoMode:     legacyResult.ElCheapoMode,
	}

	// Rite-switch cleanup: remove stale throughline IDs from all sessions
	if previousRite != "" && previousRite != riteName && !opts.DryRun {
		result.RiteSwitched = true
		result.PreviousRite = previousRite
		result.ThroughlineIDsCleaned = m.cleanupThroughlineIDs()
	}

	return result, nil
}

// syncRiteScopeAllChannels iterates over all channels, calling the single-channel
// syncRiteScope path for each. It aggregates per-channel results into ChannelResults.
// Rite-switch cleanup runs only on the first channel iteration (via the existing guard
// in syncRiteScope which reads ACTIVE_RITE state -- the first channel writes the new
// ACTIVE_RITE, so subsequent channels see matching rites and skip cleanup).
func (m *Materializer) syncRiteScopeAllChannels(opts SyncOptions) (*RiteScopeResult, error) {
	channels := paths.AllChannels()
	channelResults := make(map[string]*RiteScopeResult, len(channels))
	var firstResult *RiteScopeResult
	hasFailure := false

	for _, ch := range channels {
		perChannelOpts := opts
		perChannelOpts.Channel = ch.Name()

		chResult, err := m.syncRiteScope(perChannelOpts)
		if err != nil {
			channelResults[ch.Name()] = &RiteScopeResult{
				Status: "error",
				Error:  err.Error(),
			}
			hasFailure = true
			continue
		}
		channelResults[ch.Name()] = chResult
		if firstResult == nil {
			firstResult = chResult
		}
	}

	// Build wrapper result: top-level fields populated from first successful channel
	wrapper := &RiteScopeResult{
		ChannelResults: channelResults,
	}
	if firstResult != nil {
		wrapper.Status = firstResult.Status
		wrapper.RiteName = firstResult.RiteName
		wrapper.Source = firstResult.Source
		wrapper.SourcePath = firstResult.SourcePath
		wrapper.OrphansDetected = firstResult.OrphansDetected
		wrapper.OrphanAction = firstResult.OrphanAction
		wrapper.BackupPath = firstResult.BackupPath
		wrapper.LegacyBackupPath = firstResult.LegacyBackupPath
		wrapper.SoftMode = firstResult.SoftMode
		wrapper.DeferredStages = firstResult.DeferredStages
		wrapper.ElCheapoMode = firstResult.ElCheapoMode
		wrapper.RiteSwitched = firstResult.RiteSwitched
		wrapper.PreviousRite = firstResult.PreviousRite
		wrapper.ThroughlineIDsCleaned = firstResult.ThroughlineIDsCleaned
	}

	// If any channel failed but at least one succeeded, mark as partial
	if hasFailure && firstResult != nil {
		wrapper.Status = "partial"
	} else if hasFailure && firstResult == nil {
		// All channels failed -- return the first channel's error
		for _, ch := range channels {
			if r, ok := channelResults[ch.Name()]; ok && r.Error != "" {
				return nil, fmt.Errorf("all channels failed; %s: %s", ch.Name(), r.Error)
			}
		}
	}

	return wrapper, nil
}

// syncRiteScopeMinimal handles cross-cutting mode (no rite).
func (m *Materializer) syncRiteScopeMinimal(opts SyncOptions) (*RiteScopeResult, error) {
	// Dispatch multi-channel for minimal path
	if opts.Channel == "all" {
		return m.syncRiteScopeMinimalAllChannels(opts)
	}

	legacyOpts := Options{DryRun: opts.DryRun, Minimal: true, Channel: opts.Channel}
	legacyResult, err := m.MaterializeMinimal(legacyOpts)
	if err != nil {
		return nil, err
	}
	return &RiteScopeResult{
		Status: legacyResult.Status,
		Source: "minimal",
	}, nil
}

// syncRiteScopeMinimalAllChannels iterates over all channels for minimal sync.
func (m *Materializer) syncRiteScopeMinimalAllChannels(opts SyncOptions) (*RiteScopeResult, error) {
	channels := paths.AllChannels()
	channelResults := make(map[string]*RiteScopeResult, len(channels))
	var firstResult *RiteScopeResult
	hasFailure := false

	for _, ch := range channels {
		perChannelOpts := opts
		perChannelOpts.Channel = ch.Name()

		chResult, err := m.syncRiteScopeMinimal(perChannelOpts)
		if err != nil {
			channelResults[ch.Name()] = &RiteScopeResult{
				Status: "error",
				Error:  err.Error(),
			}
			hasFailure = true
			continue
		}
		channelResults[ch.Name()] = chResult
		if firstResult == nil {
			firstResult = chResult
		}
	}

	wrapper := &RiteScopeResult{
		ChannelResults: channelResults,
	}
	if firstResult != nil {
		wrapper.Status = firstResult.Status
		wrapper.Source = firstResult.Source
	}
	if hasFailure && firstResult != nil {
		wrapper.Status = "partial"
	} else if hasFailure && firstResult == nil {
		for _, ch := range channels {
			if r, ok := channelResults[ch.Name()]; ok && r.Error != "" {
				return nil, fmt.Errorf("all channels failed; %s: %s", ch.Name(), r.Error)
			}
		}
	}

	return wrapper, nil
}

// syncUserScope is implemented in user_scope.go (delegates to userscope sub-package)

// loadRiteManifest loads a rite's manifest.yaml file.
// When resolved is non-nil and the source is embedded, reads from the embedded FS.
func (m *Materializer) loadRiteManifest(ritePath string, resolved *ResolvedRite) (*RiteManifest, error) {
	var data []byte
	var err error

	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.EmbeddedFS != nil {
		data, err = fs.ReadFile(m.sourceResolver.EmbeddedFS, resolved.ManifestPath)
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
