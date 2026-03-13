package inscription

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/paths"
	"gopkg.in/yaml.v3"
)

// AgentFrontmatter represents the YAML frontmatter in agent files.
// Agent files use YAML frontmatter between --- delimiters.
type AgentFrontmatter struct {
	Name        string `yaml:"name"`
	Role        string `yaml:"role"`
	Description string `yaml:"description"`
	Tools       string `yaml:"tools"`
	Model       string `yaml:"model"`
	Color       string `yaml:"color"`
}

// extractFrontmatter extracts and parses YAML frontmatter from markdown content.
// Returns nil if no valid frontmatter is found.
// Delegates to frontmatter.Parse which handles both \n and \r\n line endings.
func extractFrontmatter(content string) *AgentFrontmatter {
	yamlBytes, _, err := frontmatter.Parse([]byte(content))
	if err != nil {
		return nil
	}

	var fm AgentFrontmatter
	if err := yaml.Unmarshal(yamlBytes, &fm); err != nil {
		return nil
	}

	return &fm
}

// Pipeline orchestrates the full CLAUDE.md sync process.
// It coordinates the generator, merger, and backup manager to produce
// a synchronized CLAUDE.md file based on templates and project state.
type Pipeline struct {
	// InscriptionPath is the path to the context file (e.g. CLAUDE.md, GEMINI.md).
	InscriptionPath string

	// ManifestPath is the path to the KNOSSOS_MANIFEST.yaml file.
	ManifestPath string

	// TemplateDir is the directory containing template files.
	TemplateDir string

	// BackupDir is the directory for storing backups.
	BackupDir string

	// ProjectRoot is the root directory of the project.
	ProjectRoot string
}

// InscriptionSyncOptions configures the inscription sync operation.
type InscriptionSyncOptions struct {
	// Force forces full regeneration regardless of hashes.
	Force bool

	// RiteName is the new rite name (empty = use current).
	RiteName string

	// DryRun previews changes without writing.
	DryRun bool

	// Verbose enables detailed output.
	Verbose bool

	// NoBackup skips backup creation.
	NoBackup bool
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	// Success indicates if the sync completed successfully.
	Success bool `json:"success"`

	// RegionsSynced lists the regions that were updated.
	RegionsSynced []string `json:"regions_synced"`

	// Conflicts contains any conflicts detected during merge.
	Conflicts []Conflict `json:"conflicts,omitempty"`

	// BackupPath is the path to the backup file (empty if no backup).
	BackupPath string `json:"backup_path,omitempty"`

	// Duration is how long the sync took.
	Duration time.Duration `json:"duration"`

	// InscriptionVersion is the new inscription version number.
	InscriptionVersion string `json:"inscription_version"`

	// DryRun indicates this was a preview only.
	DryRun bool `json:"dry_run,omitempty"`
}

// SyncPreview contains a preview of what sync would change.
type SyncPreview struct {
	// WouldSync lists regions that would be updated.
	WouldSync []string `json:"would_sync"`

	// WouldPreserve lists regions that would be preserved.
	WouldPreserve []string `json:"would_preserve"`

	// Conflicts lists conflicts that would occur.
	Conflicts []Conflict `json:"conflicts,omitempty"`

	// Diff is a unified diff of the changes (if requested).
	Diff string `json:"diff,omitempty"`

	// CurrentVersion is the current inscription version.
	CurrentVersion string `json:"current_version"`

	// NewVersion is what the new version would be.
	NewVersion string `json:"new_version"`
}

// ValidationResult contains the result of manifest validation.
type ValidationResult struct {
	// Valid indicates if the manifest is valid.
	Valid bool `json:"valid"`

	// Issues lists any validation problems found.
	Issues []ValidationIssue `json:"issues,omitempty"`

	// RegionCount is the number of defined regions.
	RegionCount int `json:"region_count"`

	// SchemaVersion is the manifest schema version.
	SchemaVersion string `json:"schema_version"`
}

// ValidationIssue describes a single validation problem.
type ValidationIssue struct {
	// Severity is "error", "warning", or "info".
	Severity string `json:"severity"`

	// Region is the affected region (empty if global).
	Region string `json:"region,omitempty"`

	// Message describes the issue.
	Message string `json:"message"`
}

// NewPipeline creates a new pipeline for the given project root.
// HA-FS: InscriptionPath targets the actual CC channel context file (SCAR-002: never rename .claude/)
func NewPipeline(projectRoot string) *Pipeline {
	return &Pipeline{
		InscriptionPath: filepath.Join(projectRoot, ".claude", "CLAUDE.md"),
		ManifestPath: DefaultManifestPath(projectRoot),
		TemplateDir:  filepath.Join(projectRoot, "knossos", "templates"),
		BackupDir:    filepath.Join(projectRoot, ".knossos", "backups"),
		ProjectRoot:  projectRoot,
	}
}

// NewPipelineWithPaths creates a pipeline with custom paths.
func NewPipelineWithPaths(inscriptionPath, manifestPath, templateDir, backupDir string) *Pipeline {
	projectRoot := filepath.Dir(filepath.Dir(inscriptionPath))
	return &Pipeline{
		InscriptionPath: inscriptionPath,
		ManifestPath: manifestPath,
		TemplateDir:  templateDir,
		BackupDir:    backupDir,
		ProjectRoot:  projectRoot,
	}
}

// Sync performs the full sync pipeline.
// This delegates to SyncInscription for the core merge/write logic,
// adding Pipeline-specific features (backup, dry-run).
func (p *Pipeline) Sync(opts InscriptionSyncOptions) (*SyncResult, error) {
	start := time.Now()

	// Handle dry-run via DryRun method
	if opts.DryRun {
		preview, err := p.DryRun(opts)
		if err != nil {
			return nil, err
		}
		return &SyncResult{
			Success:            true,
			RegionsSynced:      preview.WouldSync,
			Conflicts:          preview.Conflicts,
			Duration:           time.Since(start),
			InscriptionVersion: preview.CurrentVersion,
			DryRun:             true,
		}, nil
	}

	// Build render context from project state
	ctx, err := p.buildRenderContext(nil)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to build render context", err)
	}

	// Create backup before writing (Pipeline-specific feature)
	backupPath := ""
	if !opts.NoBackup {
		if data, _ := os.ReadFile(p.InscriptionPath); len(data) > 0 {
			backupMgr := NewBackupManager(p.ProjectRoot)
			backupPath, _ = backupMgr.CreateBackup()
		}
	}

	// Delegate to unified SyncInscription
	channelDir := filepath.Dir(p.InscriptionPath)
	result, err := SyncInscription(SyncInscriptionOptions{
		ChannelDir:     channelDir,
		RenderCtx:      ctx,
		ActiveRite:     opts.RiteName,
		TemplateDir:    p.TemplateDir,
		UpdateManifest: true,
	})
	if err != nil {
		return nil, err
	}

	return &SyncResult{
		Success:            true,
		RegionsSynced:      result.MergeResult.RegionsMerged,
		Conflicts:          result.MergeResult.Conflicts,
		BackupPath:         backupPath,
		Duration:           time.Since(start),
		InscriptionVersion: result.ManifestVersion,
	}, nil
}

// DryRun previews changes without writing.
func (p *Pipeline) DryRun(opts InscriptionSyncOptions) (*SyncPreview, error) {
	opts.DryRun = true

	// Load manifest
	loader := NewManifestLoader(p.ProjectRoot)
	manifest, err := loader.LoadOrCreate()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load manifest", err)
	}

	currentVersion := manifest.InscriptionVersion

	// Update rite if specified
	if opts.RiteName != "" {
		manifest.SetActiveRite(opts.RiteName)
	}

	// Build context
	ctx, err := p.buildRenderContext(manifest)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to build render context", err)
	}

	// Generate content
	generator := NewGenerator(p.TemplateDir, manifest, ctx)
	generatedContent, err := generator.GenerateAll()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to generate content", err)
	}

	// Load existing content
	existingContent := ""
	if data, err := os.ReadFile(p.InscriptionPath); err == nil {
		existingContent = string(data)
	}

	// Create merger and detect conflicts
	merger := NewMerger(manifest, generator)
	conflicts, err := merger.DetectConflicts(existingContent)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to detect conflicts", err)
	}

	// Determine which regions would be synced vs preserved
	var wouldSync, wouldPreserve []string
	for _, name := range manifest.SectionOrder {
		region := manifest.GetRegion(name)
		if region == nil {
			continue
		}
		if region.Owner == OwnerSatellite {
			wouldPreserve = append(wouldPreserve, name)
		} else {
			if _, ok := generatedContent[name]; ok {
				wouldSync = append(wouldSync, name)
			}
		}
	}

	// Compute new version
	version := 1
	if currentVersion != "" {
		if v, err := parseVersion(currentVersion); err == nil {
			version = v + 1
		}
	}

	return &SyncPreview{
		WouldSync:      wouldSync,
		WouldPreserve:  wouldPreserve,
		Conflicts:      conflicts,
		CurrentVersion: currentVersion,
		NewVersion:     itoa(version),
	}, nil
}

// Rollback reverts to a previous backup.
// If timestamp is empty, reverts to the most recent backup.
func (p *Pipeline) Rollback(timestamp string) error {
	backupMgr := NewBackupManager(p.ProjectRoot)
	return backupMgr.RestoreBackup(timestamp)
}

// Validate checks the current state without syncing.
func (p *Pipeline) Validate() (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Issues: make([]ValidationIssue, 0),
	}

	// Load manifest
	loader := NewManifestLoader(p.ProjectRoot)
	manifest, err := loader.Load()
	if err != nil {
		if errors.IsNotFound(err) {
			result.Issues = append(result.Issues, ValidationIssue{
				Severity: "warning",
				Message:  "No KNOSSOS_MANIFEST.yaml found - will be created on first sync",
			})
			return result, nil
		}
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: "error",
			Message:  "Failed to load manifest: " + err.Error(),
		})
		return result, nil
	}

	result.SchemaVersion = manifest.SchemaVersion
	result.RegionCount = len(manifest.Regions)

	// Check CLAUDE.md exists
	if _, err := os.Stat(p.InscriptionPath); os.IsNotExist(err) {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: "warning",
			Message:  "Context file not found at " + p.InscriptionPath,
		})
	} else if err != nil {
		// Handle other stat errors (permissions, etc.)
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: "error",
			Message:  "Cannot access CLAUDE.md: " + err.Error(),
		})
	} else {
		// Parse CLAUDE.md for markers
		content, err := os.ReadFile(p.InscriptionPath)
		if err != nil {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity: "error",
				Message:  "Cannot read CLAUDE.md: " + err.Error(),
			})
			return result, nil
		}
		parser := NewMarkerParser()
		parseResult := parser.Parse(string(content))

		if parseResult.HasErrors() {
			result.Valid = false
			for _, pe := range parseResult.Errors {
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: "error",
					Message:  pe.Message,
				})
			}
		}

		// Check for regions in manifest but not in file
		for name := range manifest.Regions {
			if parseResult.GetRegion(name) == nil {
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: "info",
					Region:   name,
					Message:  "Region defined in manifest but not found in CLAUDE.md",
				})
			}
		}

		// Check for regions in file but not in manifest
		deprecated := make(map[string]bool)
		for _, name := range DeprecatedRegions() {
			deprecated[name] = true
		}
		for name := range parseResult.Regions {
			if deprecated[name] {
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: "warning",
					Region:   name,
					Message:  "Deprecated region found in CLAUDE.md — will be removed on next sync",
				})
			} else if !manifest.HasRegion(name) {
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: "info",
					Region:   name,
					Message:  "Region found in CLAUDE.md but not defined in manifest",
				})
			}
		}
	}

	// Check template directory
	if _, err := os.Stat(p.TemplateDir); os.IsNotExist(err) {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: "info",
			Message:  "Template directory not found at " + p.TemplateDir + " - will use defaults",
		})
	}

	// Validate each region
	for name, region := range manifest.Regions {
		// Check regenerate regions have source
		if region.Owner == OwnerRegenerate && region.Source == "" {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity: "error",
				Region:   name,
				Message:  "Regenerate region missing source",
			})
		}
	}

	return result, nil
}

// ListBackups returns all available backups.
func (p *Pipeline) ListBackups() ([]BackupInfo, error) {
	backupMgr := NewBackupManager(p.ProjectRoot)
	return backupMgr.ListBackups()
}

// GetDiff returns the diff between current CLAUDE.md and what would be generated.
func (p *Pipeline) GetDiff(regionName string) (string, error) {
	// Load manifest
	loader := NewManifestLoader(p.ProjectRoot)
	manifest, err := loader.LoadOrCreate()
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to load manifest", err)
	}

	// Build context
	ctx, err := p.buildRenderContext(manifest)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to build render context", err)
	}

	// Generate content
	generator := NewGenerator(p.TemplateDir, manifest, ctx)

	// Load existing content
	existingContent := ""
	if data, err := os.ReadFile(p.InscriptionPath); err == nil {
		existingContent = string(data)
	}

	parser := NewMarkerParser()
	parseResult := parser.Parse(existingContent)

	if regionName != "" {
		// Diff specific region
		existingRegion := parseResult.GetRegion(regionName)
		oldContent := ""
		if existingRegion != nil {
			oldContent = existingRegion.Content
		}

		newContent, err := generator.GenerateSection(regionName)
		if err != nil {
			return "", err
		}

		return simpleDiff(regionName, oldContent, newContent), nil
	}

	// Diff all regions
	var sb strings.Builder
	for _, name := range manifest.SectionOrder {
		region := manifest.GetRegion(name)
		if region == nil || region.Owner == OwnerSatellite {
			continue
		}

		existingRegion := parseResult.GetRegion(name)
		oldContent := ""
		if existingRegion != nil {
			oldContent = existingRegion.Content
		}

		newContent, err := generator.GenerateSection(name)
		if err != nil {
			continue
		}

		if oldContent != newContent {
			sb.WriteString(simpleDiff(name, oldContent, newContent))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// buildRenderContext creates a RenderContext from project state.
// If manifest is nil, rite name is loaded from ACTIVE_RITE file.
func (p *Pipeline) buildRenderContext(manifest *Manifest) (*RenderContext, error) {
	ctx := &RenderContext{
		ProjectRoot:      p.ProjectRoot,
		IsKnossosProject: p.TemplateDir != "" && strings.HasPrefix(p.TemplateDir, p.ProjectRoot),
		KnossosVars:      make(map[string]string),
	}

	if manifest != nil {
		ctx.ActiveRite = manifest.ActiveRite
	}

	// Try to load active rite if not set
	if ctx.ActiveRite == "" {
		ctx.ActiveRite = paths.NewResolver(p.ProjectRoot).ReadActiveRite()
	}

	// Load agent information from the channel agents directory
	agentsDir := filepath.Join(p.ProjectRoot, ".claude", "agents") // HA-FS: actual CC channel agents directory path
	allAgents, err := p.loadAgents(agentsDir)
	if err != nil {
		return ctx, nil
	}

	// Scope agents by rite manifest if a rite is active
	if ctx.ActiveRite != "" {
		riteAgentNames := p.loadRiteAgentNames(ctx.ActiveRite)
		if len(riteAgentNames) > 0 {
			for _, agent := range allAgents {
				if riteAgentNames[agent.Name] {
					ctx.Agents = append(ctx.Agents, agent)
				} else {
					ctx.CrossRiteAgents = append(ctx.CrossRiteAgents, agent)
				}
			}
			ctx.AgentCount = len(ctx.Agents)
		} else {
			// Fallback: couldn't load rite manifest, show all agents
			ctx.Agents = allAgents
			ctx.AgentCount = len(allAgents)
		}
	} else {
		// No active rite (cross-cutting mode): all agents are cross-rite
		ctx.CrossRiteAgents = allAgents
		ctx.AgentCount = 0
	}

	return ctx, nil
}

// loadRiteAgentNames loads agent names from a rite's manifest.yaml.
func (p *Pipeline) loadRiteAgentNames(riteName string) map[string]bool {
	manifestPath := filepath.Join(p.ProjectRoot, "rites", riteName, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil
	}

	var manifest struct {
		Agents []struct {
			Name string `yaml:"name"`
		} `yaml:"agents"`
	}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil
	}

	names := make(map[string]bool, len(manifest.Agents))
	for _, a := range manifest.Agents {
		names[a.Name] = true
	}
	return names
}

// loadAgents reads agent metadata from a directory.
// Parses YAML frontmatter for role/description, falls back to legacy scanning.
func (p *Pipeline) loadAgents(agentsDir string) ([]AgentInfo, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	var agents []AgentInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		agent := AgentInfo{
			Name: strings.TrimSuffix(entry.Name(), ".md"),
			File: entry.Name(),
		}

		// Try to parse YAML frontmatter first
		if fm := extractFrontmatter(string(content)); fm != nil {
			// Use description as Role if role is empty (description is more verbose)
			if fm.Description != "" {
				agent.Role = firstSentence(fm.Description)
			} else if fm.Role != "" {
				agent.Role = firstSentence(fm.Role)
			}
			// Look for produces in frontmatter (future extension)
			agents = append(agents, agent)
			continue
		}

		// Fallback: legacy scanning for files without frontmatter
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Skip YAML delimiters, headers, and empty lines
			if line == "---" || strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			// Skip lines that look like YAML key:value
			if strings.Contains(line, ":") && !strings.HasPrefix(line, ">") {
				continue
			}
			if agent.Role == "" {
				// First non-header, non-YAML line is likely the role description
				agent.Role = truncate(line, 80)
			}
			// Look for "Produces:" or similar patterns
			if strings.Contains(strings.ToLower(line), "produces") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					agent.Produces = strings.TrimSpace(parts[1])
				}
			}
			if agent.Role != "" && agent.Produces != "" {
				break
			}
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// simpleDiff creates a simple diff output for a region.
func simpleDiff(regionName, oldContent, newContent string) string {
	var sb strings.Builder
	sb.WriteString("--- ")
	sb.WriteString(regionName)
	sb.WriteString(" (current)\n")
	sb.WriteString("+++ ")
	sb.WriteString(regionName)
	sb.WriteString(" (generated)\n")

	switch {
	case oldContent == "" && newContent != "":
		sb.WriteString("+ (new region)\n")
		for _, line := range strings.Split(newContent, "\n") {
			sb.WriteString("+ ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	case oldContent != "" && newContent == "":
		sb.WriteString("- (region removed)\n")
		for _, line := range strings.Split(oldContent, "\n") {
			sb.WriteString("- ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	case oldContent != newContent:
		sb.WriteString("@@ region changed @@\n")
		// Simple line-by-line comparison
		oldLines := strings.Split(oldContent, "\n")
		newLines := strings.Split(newContent, "\n")

		// Very basic diff - show removed then added
		for _, line := range oldLines {
			if !contains(newLines, line) {
				sb.WriteString("- ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
		for _, line := range newLines {
			if !contains(oldLines, line) {
				sb.WriteString("+ ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// contains checks if a slice contains a string.
func contains(slice []string, s string) bool {
	return slices.Contains(slice, s)
}

// firstSentence extracts the first sentence from a string.
// Returns up to the first ". " or "." at end, or truncates to 80 chars.
func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	// Check for first sentence boundary
	if idx := strings.Index(s, ". "); idx > 0 {
		return s[:idx+1]
	}
	// Check for period at end
	if strings.HasSuffix(s, ".") && len(s) <= 80 {
		return s
	}
	return truncate(s, 80)
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// parseVersion parses a version string to int.
func parseVersion(v string) (int, error) {
	var result int
	for _, c := range v {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			break
		}
	}
	return result, nil
}
