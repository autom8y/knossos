package materialize

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/inscription"
	"github.com/autom8y/knossos/internal/materialize/compiler"
	"github.com/autom8y/knossos/internal/provenance"
)

// materializeCLAUDEmd generates CLAUDE.md using the inscription system.
// Delegates to inscription.SyncCLAUDEmd for the core merge/write logic,
// then records provenance for the written file.
// Returns the path to legacy backup if migration occurred, or empty string if no backup.
func (m *Materializer) materializeCLAUDEmd(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, collector provenance.Collector, modelOverride string, comp compiler.ChannelCompiler) (string, error) {
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

	projectRoot := m.resolver.ProjectRoot()
	renderCtx := &inscription.RenderContext{
		ActiveRite:       manifest.Name,
		AgentCount:       len(manifest.Agents),
		Agents:           agents,
		KnossosVars:      make(map[string]string),
		ProjectRoot:      projectRoot,
		IsKnossosProject: m.templatesDir != "" && strings.HasPrefix(m.templatesDir, projectRoot),
		ModelOverride:    modelOverride,
	}

	// Resolve template source: embedded FS or filesystem directory
	var templateFS fs.FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		templateFS = m.templatesFS(resolved)
	}

	contextFilename := "CLAUDE.md"
	if comp != nil {
		contextFilename = comp.ContextFilename()
	}

	// Delegate to canonical SyncCLAUDEmd
	result, err := inscription.SyncCLAUDEmd(inscription.CLAUDEmdSyncOptions{
		ClaudeDir:       claudeDir,
		RenderCtx:       renderCtx,
		ActiveRite:      manifest.Name,
		TemplateDir:     m.templatesDir,
		TemplateFS:      templateFS,
		UpdateManifest:  true,
		ContextFilename: contextFilename,
	})
	if err != nil {
		return "", err
	}

	// Record provenance after successful write (materialization-specific concern)
	if result.Written {
		srcRelPath := "(generated)"
		if m.templatesDir != "" {
			projectRoot := m.resolver.ProjectRoot()
			sourcePath := filepath.Join(m.templatesDir, "CLAUDE.md.tpl")
			if rel, err := filepath.Rel(projectRoot, sourcePath); err == nil && rel != "" {
				srcRelPath = rel
			}
		}
		collector.Record(contextFilename, provenance.NewKnossosEntry(
			provenance.ScopeRite,
			srcRelPath,
			"template",
			checksum.Content(result.MergeResult.Content),
		))
	}

	return result.LegacyBackupPath, nil
}

// materializeMinimalCLAUDEmd generates CLAUDE.md for cross-cutting mode (no agents).
// Delegates to inscription.SyncCLAUDEmd without manifest updates.
func (m *Materializer) materializeMinimalCLAUDEmd(claudeDir string, collector provenance.Collector) (string, error) {
	projectRoot := m.resolver.ProjectRoot()
	renderCtx := &inscription.RenderContext{
		ActiveRite:       "",
		AgentCount:       0,
		Agents:           []inscription.AgentInfo{},
		KnossosVars:      make(map[string]string),
		ProjectRoot:      projectRoot,
		IsKnossosProject: m.templatesDir != "" && strings.HasPrefix(m.templatesDir, projectRoot),
	}

	contextFilename := "CLAUDE.md"
	if m.claudeDirOverride != "" && strings.HasSuffix(m.claudeDirOverride, ".gemini") {
		contextFilename = "GEMINI.md"
	}

	result, err := inscription.SyncCLAUDEmd(inscription.CLAUDEmdSyncOptions{
		ClaudeDir:       claudeDir,
		RenderCtx:       renderCtx,
		TemplateDir:     m.templatesDir,
		UpdateManifest:  false,
		ContextFilename: contextFilename,
	})
	if err != nil {
		return "", err
	}

	return result.LegacyBackupPath, nil
}

// prevalidateCLAUDEmd validates that CLAUDE.md generation will succeed without
// writing any files. This is called BEFORE destructive operations (agent writes,
// orphan removal) to prevent partial state when template rendering fails.
func (m *Materializer) prevalidateCLAUDEmd(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, modelOverride string) error {
	agents := make([]inscription.AgentInfo, 0, len(manifest.Agents))
	for _, agent := range manifest.Agents {
		agents = append(agents, inscription.AgentInfo{
			Name: agent.Name,
			File: agent.Name + ".md",
			Role: agent.Role,
		})
	}

	projectRoot := m.resolver.ProjectRoot()
	renderCtx := &inscription.RenderContext{
		ActiveRite:       manifest.Name,
		AgentCount:       len(manifest.Agents),
		Agents:           agents,
		KnossosVars:      make(map[string]string),
		ProjectRoot:      projectRoot,
		IsKnossosProject: m.templatesDir != "" && strings.HasPrefix(m.templatesDir, projectRoot),
		ModelOverride:    modelOverride,
	}

	// Resolve template source
	var templateFS fs.FS
	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.embeddedTemplates != nil {
		templateFS = m.templatesFS(resolved)
	}

	// Load or create manifest (read-only validation)
	knossosManifestPath := filepath.Join(projectRoot, ".knossos", "KNOSSOS_MANIFEST.yaml")
	loader := inscription.NewManifestLoader(projectRoot)
	loader.ManifestPath = knossosManifestPath
	insManifest, err := loader.LoadOrCreate()
	if err != nil {
		return err
	}
	insManifest.AdoptNewDefaults()

	// Create generator and validate all sections render without error
	var generator *inscription.Generator
	if templateFS != nil {
		generator = inscription.NewGeneratorWithFS(templateFS, insManifest, renderCtx)
	} else {
		generator = inscription.NewGenerator(m.templatesDir, insManifest, renderCtx)
	}

	_, err = generator.GenerateAll()
	return err
}
