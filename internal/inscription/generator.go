package inscription

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/autom8y/knossos/internal/channel"
	"github.com/autom8y/knossos/internal/errors"
)

// AgentInfo holds agent metadata for template rendering.
// This mirrors rite.AgentInfo for the inscription package.
type AgentInfo struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Role     string `json:"role"`
	Produces string `json:"produces"`
}

// SummonableAgentInfo holds metadata for summonable agents shown in inscription.
// Summonable agents are available on demand via ari agent summon/dismiss rather
// than being permanently materialized like standing agents.
type SummonableAgentInfo struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Command string `json:"command"`
}

// RenderContext provides data for template rendering.
type RenderContext struct {
	// ActiveRite is the current rite name.
	ActiveRite string

	// AgentCount is the number of rite-native agents.
	AgentCount int

	// Agents contains metadata for rite-native agents.
	Agents []AgentInfo

	// CrossRiteAgents contains agents available across rites (moirai, pythia, etc.).
	CrossRiteAgents []AgentInfo

	// KnossosVars contains additional variables for template rendering.
	KnossosVars map[string]string

	// ProjectRoot is the root directory of the project.
	ProjectRoot string

	// IsKnossosProject is true when materializing the knossos repo itself
	// (templates are within the project). False for satellite projects.
	IsKnossosProject bool

	// ModelOverride is the active model override (e.g., "haiku" for el-cheapo mode).
	// Empty string means no override active.
	ModelOverride string

	// Channel is the target channel ("claude" or "gemini").
	// Empty string is treated as "claude" — all templates default to CC behavior.
	Channel string

	// SummonableAgents contains agents available for on-demand summoning.
	// These agents are not permanently materialized but can be summoned via
	// `ari agent summon {name}` and dismissed via `ari agent dismiss {name}`.
	SummonableAgents []SummonableAgentInfo
}

// Generator handles content generation for CLAUDE.md regions.
type Generator struct {
	// TemplateDir is the directory containing template files.
	TemplateDir string

	// TemplateFS is an optional fs.FS for reading templates (e.g., from embedded assets).
	// When set, templates are read from this FS before falling back to TemplateDir.
	TemplateFS fs.FS

	// Manifest is the current KNOSSOS_MANIFEST.yaml.
	Manifest *Manifest

	// Context is the render context for template execution.
	Context *RenderContext

	// sectionTemplates maps region names to template content.
	sectionTemplates map[string]string
}

// NewGenerator creates a new generator with the given configuration.
func NewGenerator(templateDir string, manifest *Manifest, ctx *RenderContext) *Generator {
	return &Generator{
		TemplateDir:      templateDir,
		Manifest:         manifest,
		Context:          ctx,
		sectionTemplates: make(map[string]string),
	}
}

// NewGeneratorWithFS creates a generator that reads templates from an fs.FS
// instead of the filesystem. Used when materializing from embedded assets.
func NewGeneratorWithFS(templateFS fs.FS, manifest *Manifest, ctx *RenderContext) *Generator {
	return &Generator{
		TemplateFS:       templateFS,
		Manifest:         manifest,
		Context:          ctx,
		sectionTemplates: make(map[string]string),
	}
}

// GenerateSection generates content for a region based on its owner type.
// Returns UNWRAPPED content only (no START/END markers).
// Use RenderRegion() for wrapped content suitable for direct output.
// Use this method when preparing content for the merger (MergeRegions).
func (g *Generator) GenerateSection(regionName string) (string, error) {
	region := g.Manifest.GetRegion(regionName)
	if region == nil {
		return "", errors.NewWithDetails(errors.CodeFileNotFound,
			"region not found in manifest",
			map[string]any{"region": regionName})
	}

	switch region.Owner {
	case OwnerKnossos:
		return g.renderKnossosRegion(regionName)
	case OwnerRegenerate:
		return g.regenerateFromSource(regionName, region.Source)
	case OwnerSatellite:
		// Satellite regions: render template for new files, but merger will preserve existing content
		return g.renderSatelliteRegion(regionName)
	default:
		return "", errors.NewWithDetails(errors.CodeUsageError,
			"unknown owner type",
			map[string]any{"region": regionName, "owner": string(region.Owner)})
	}
}

// RenderRegion renders a specific region and wraps it with START/END markers.
// Returns WRAPPED content suitable for direct output to file.
// WARNING: Do not pass the result of this method to MergeRegions - use GenerateSection instead.
// This is the main entry point per TDD Section 5.2 Stage 3.
func (g *Generator) RenderRegion(regionName string) (string, error) {
	content, err := g.GenerateSection(regionName)
	if err != nil {
		return "", err
	}

	// For satellite regions without templates, return empty (backward compatibility)
	// Satellite regions with templates will have content from renderSatelliteRegion
	if content == "" {
		region := g.Manifest.GetRegion(regionName)
		if region != nil && region.Owner == OwnerSatellite {
			return "", nil
		}
	}

	// Build marker options
	options := g.buildMarkerOptions(regionName)

	return WrapContent(regionName, content, options), nil
}

// buildMarkerOptions constructs options map for a region's start marker.
func (g *Generator) buildMarkerOptions(regionName string) map[string]string {
	region := g.Manifest.GetRegion(regionName)
	if region == nil {
		return nil
	}

	options := make(map[string]string)

	// Add owner hint for regenerate regions
	if region.Owner == OwnerRegenerate {
		options["regenerate"] = "true"
		if region.Source != "" {
			options["source"] = region.Source
		}
	}

	return options
}

// renderKnossosRegion renders a knossos-owned region from templates.
func (g *Generator) renderKnossosRegion(regionName string) (string, error) {
	// Try section-specific template first
	templatePath := g.getSectionTemplatePath(regionName)
	if templatePath != "" {
		return g.renderTemplateFile(templatePath)
	}

	// Fall back to inline template from sectionTemplates
	if tmpl, ok := g.sectionTemplates[regionName]; ok {
		return g.renderTemplateString(regionName, tmpl)
	}

	// Use default content for known sections
	return g.getDefaultSectionContent(regionName)
}

// renderSatelliteRegion renders a satellite region's template.
// Returns template content for new files; merger will preserve existing content.
// Returns empty string if no template exists (backward compatibility).
func (g *Generator) renderSatelliteRegion(regionName string) (string, error) {
	// Try to find a template file for this satellite region
	templatePath := g.getSectionTemplatePath(regionName)
	if templatePath != "" {
		return g.renderTemplateFile(templatePath)
	}

	// No template for this satellite region - return empty (backward compatible)
	return "", nil
}

// getSectionTemplatePath returns the path to a section template if it exists.
// When TemplateFS is set, returns the fs.FS-relative path if the template exists there.
// Otherwise falls back to the filesystem TemplateDir.
func (g *Generator) getSectionTemplatePath(regionName string) string {
	// Try embedded FS first
	if g.TemplateFS != nil {
		fsPath := "sections/" + regionName + ".md.tpl"
		if _, err := fs.Stat(g.TemplateFS, fsPath); err == nil {
			return fsPath // Return FS-relative path
		}
	}

	// Fall back to filesystem
	if g.TemplateDir == "" {
		return ""
	}

	path := filepath.Join(g.TemplateDir, "sections", regionName+".md.tpl")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// renderTemplateFile renders a template from a file path.
// When TemplateFS is set, reads from the embedded FS; otherwise reads from os filesystem.
func (g *Generator) renderTemplateFile(path string) (string, error) {
	var data []byte
	var err error

	if g.TemplateFS != nil {
		data, err = fs.ReadFile(g.TemplateFS, path)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", errors.Wrap(errors.CodeFileNotFound, "failed to read template file", err)
	}

	return g.renderTemplateString(filepath.Base(path), string(data))
}

// renderTemplateString executes a template string with the render context.
func (g *Generator) renderTemplateString(name string, tmplStr string) (string, error) {
	tmpl, err := template.New(name).Funcs(g.templateFuncs()).Parse(tmplStr)
	if err != nil {
		return "", errors.NewWithDetails(errors.CodeParseError,
			"failed to parse template",
			map[string]any{"name": name, "cause": err.Error()})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.Context); err != nil {
		return "", errors.NewWithDetails(errors.CodeGeneralError,
			"failed to execute template",
			map[string]any{"name": name, "cause": err.Error()})
	}

	return strings.TrimSpace(buf.String()), nil
}

// templateFuncs returns the custom template functions.
func (g *Generator) templateFuncs() template.FuncMap {
	// Start with Sprig functions (100+ utility functions)
	funcs := sprig.TxtFuncMap()

	// Add Knossos-specific functions
	funcs["include"] = g.includePartial
	funcs["ifdef"] = g.conditionalInclude
	funcs["agents"] = g.loadAgentTable
	funcs["term"] = g.lookupTerminology

	// channelDir returns the channel-appropriate dot-directory name.
	// Used in templates as: `{{ channelDir }}/agents/`
	funcs["channelDir"] = func() string {
		if g.Context != nil && g.Context.Channel == "gemini" {
			return ".gemini"
		}
		return ".claude"
	}

	// toolName translates a CC tool name to the channel-appropriate wire name.
	// For Gemini, translates known tools; unknown tools pass through.
	// For Claude (or empty channel), returns the name unchanged.
	// Used in templates as: `{{ toolName "Read" }}(".know/architecture.md")`
	funcs["toolName"] = func(name string) string {
		if g.Context != nil && g.Context.Channel == "gemini" {
			if gemini, ok := channel.TranslateTool(name); ok && gemini != "" {
				return gemini
			}
		}
		return name
	}

	// Note: Sprig already provides join, lower, upper, title, and many more

	return funcs
}

// includePartial loads and renders a partial template.
func (g *Generator) includePartial(partialPath string) (string, error) {
	// When using TemplateFS, the path is already FS-relative
	if g.TemplateFS != nil {
		return g.renderTemplateFile(partialPath)
	}

	if g.TemplateDir == "" {
		return "", errors.New(errors.CodeFileNotFound, "template directory not configured")
	}

	fullPath := filepath.Join(g.TemplateDir, partialPath)
	return g.renderTemplateFile(fullPath)
}

// conditionalInclude returns content if condition is true.
func (g *Generator) conditionalInclude(condition bool, content string) string {
	if condition {
		return content
	}
	return ""
}

// loadAgentTable generates a markdown table of agents.
func (g *Generator) loadAgentTable() string {
	if g.Context == nil || len(g.Context.Agents) == 0 {
		return "| Agent | Role |\n| ----- | ---- |\n"
	}

	var sb strings.Builder
	sb.WriteString("| Agent | Role |\n")
	sb.WriteString("| ----- | ---- |\n")

	for _, agent := range g.Context.Agents {
		sb.WriteString("| **")
		sb.WriteString(agent.Name)
		sb.WriteString("** | ")
		sb.WriteString(agent.Role)
		sb.WriteString(" |\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// lookupTerminology returns the definition for a term.
func (g *Generator) lookupTerminology(term string) string {
	// Default terminology definitions
	terminology := map[string]string{
		"knossos":         "The platform (the labyrinth itself)",
		"ariadne":         "CLI binary (`ari`) - the clew ensuring return",
		"theseus":         "Claude Code agent - the navigator with amnesia",
		"moirai":          "Session lifecycle agent - the Fates who spin, measure, and cut",
		"white-sails":     "Confidence signal - honest return indicator",
		"rites":           "Practice bundles - invokable ceremonies",
		"pantheon":        "Agent collection within a rite",
		"dromena":         "Slash commands - user-invoked, transient",
		"legomena":        "Skills - model-invoked, persistent",
		"mena":            "Source directory for dromena + legomena",
		"inscription":     "CLAUDE.md generation from templates",
		"materialization": "Source to .claude/ projection",
	}

	if def, ok := terminology[strings.ToLower(term)]; ok {
		return def
	}

	// Check KnossosVars for custom terminology
	if g.Context != nil && g.Context.KnossosVars != nil {
		if def, ok := g.Context.KnossosVars["term_"+strings.ToLower(term)]; ok {
			return def
		}
	}

	return term
}

// regenerateFromSource generates content from a dynamic source.
func (g *Generator) regenerateFromSource(regionName string, source string) (string, error) {
	switch {
	case strings.HasPrefix(source, "ACTIVE_RITE"):
		return g.generateQuickStartContent()
	case strings.HasPrefix(source, "agents/"):
		return g.generateAgentConfigsContent()
	default:
		return "", errors.NewWithDetails(errors.CodeUsageError,
			"unknown regenerate source",
			map[string]any{"region": regionName, "source": source})
	}
}

// generateQuickStartContent generates the Quick Start section content.
func (g *Generator) generateQuickStartContent() (string, error) {
	if g.Context == nil {
		return g.getDefaultQuickStartContent(), nil
	}

	var sb strings.Builder

	// Header with rite info
	sb.WriteString("## Quick Start\n\n")

	if g.Context.ActiveRite != "" && g.Context.AgentCount > 0 {
		sb.WriteString(itoa(g.Context.AgentCount))
		sb.WriteString("-agent workflow (")
		sb.WriteString(g.Context.ActiveRite)
		sb.WriteString("):\n\n")

		// Rite-native agent table
		sb.WriteString(g.loadAgentTable())
		sb.WriteString("\n\n")

		// Cross-rite agents (single-line summary)
		if len(g.Context.CrossRiteAgents) > 0 {
			var names []string
			for _, a := range g.Context.CrossRiteAgents {
				names = append(names, "`"+a.Name+"`")
			}
			sb.WriteString("Cross-rite agents also available: ")
			sb.WriteString(strings.Join(names, ", "))
			sb.WriteString("\n\n")
		}
	} else {
		// Cross-cutting mode (no rite): show cross-rite agents only
		sb.WriteString("Multi-agent workflow:\n\n")
		if len(g.Context.CrossRiteAgents) > 0 {
			sb.WriteString(g.loadCrossRiteAgentTable())
			sb.WriteString("\n\n")
		}
	}

	// Footer
	switch {
	case g.Context.IsKnossosProject:
		sb.WriteString("Entry point: `/go`. Agent invocation patterns: `prompting` skill. Routing guidance: `/consult`.")
	case g.Context.Channel == "gemini":
		sb.WriteString("Agents activate when your prompt matches their description.")
	default:
		sb.WriteString("Delegate to specialists via Task tool.")
	}

	return sb.String(), nil
}

// loadCrossRiteAgentTable generates a markdown table of cross-rite agents.
func (g *Generator) loadCrossRiteAgentTable() string {
	if g.Context == nil || len(g.Context.CrossRiteAgents) == 0 {
		return "| Agent | Role |\n| ----- | ---- |\n"
	}

	var sb strings.Builder
	sb.WriteString("| Agent | Role |\n")
	sb.WriteString("| ----- | ---- |\n")

	for _, agent := range g.Context.CrossRiteAgents {
		sb.WriteString("| **")
		sb.WriteString(agent.Name)
		sb.WriteString("** | ")
		sb.WriteString(agent.Role)
		sb.WriteString(" |\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// generateAgentConfigsContent generates the Agent Configurations section.
func (g *Generator) generateAgentConfigsContent() (string, error) {
	hasRiteAgents := g.Context != nil && len(g.Context.Agents) > 0
	hasCrossRiteAgents := g.Context != nil && len(g.Context.CrossRiteAgents) > 0
	hasSummonableAgents := g.Context != nil && len(g.Context.SummonableAgents) > 0

	if !hasRiteAgents && !hasCrossRiteAgents && !hasSummonableAgents {
		return g.getDefaultAgentConfigsContent(), nil
	}

	var sb strings.Builder
	sb.WriteString("## Agents\n\n")

	if hasRiteAgents || hasCrossRiteAgents {
		if g.Context != nil && g.Context.Channel == "gemini" {
			sb.WriteString("Prompts in `.gemini/agents/`:\n\n")
		} else {
			sb.WriteString("Prompts in `.claude/agents/`:\n\n")
		}

		if hasRiteAgents {
			for _, agent := range g.Context.Agents {
				sb.WriteString("- `")
				sb.WriteString(agent.File)
				sb.WriteString("` - ")
				sb.WriteString(agent.Role)
				sb.WriteString("\n")
			}
		}

		if hasCrossRiteAgents {
			if hasRiteAgents {
				sb.WriteString("\nCross-rite agents:\n\n")
			}
			for _, agent := range g.Context.CrossRiteAgents {
				sb.WriteString("- `")
				sb.WriteString(agent.File)
				sb.WriteString("` - ")
				sb.WriteString(agent.Role)
				sb.WriteString("\n")
			}
		}
	}

	// Summonable Heroes section: agents available on demand via ari agent summon/dismiss
	if hasSummonableAgents {
		sb.WriteString("\n### Summonable Heroes\n")
		sb.WriteString("Operational agents available on demand. Their commands handle the lifecycle:\n")
		for _, agent := range g.Context.SummonableAgents {
			sb.WriteString("- **")
			sb.WriteString(agent.Name)
			sb.WriteString("** - ")
			sb.WriteString(agent.Role)
			sb.WriteString(" -> `")
			sb.WriteString(agent.Command)
			sb.WriteString("`\n")
		}
		restartText := "restart CC"
		if g.Context != nil && g.Context.Channel == "gemini" {
			restartText = "restart Gemini Code Assist"
		}
		sb.WriteString("\nSummon: `ari agent summon {name}` then ")
		sb.WriteString(restartText)
		sb.WriteString(".\n")
		sb.WriteString("Dismiss: `ari agent dismiss {name}` then ")
		sb.WriteString(restartText)
		sb.WriteString(".")
	}

	return sb.String(), nil
}

// getDefaultSectionContent returns default content for known sections.
func (g *Generator) getDefaultSectionContent(regionName string) (string, error) {
	defaults := map[string]string{
		"execution-mode":          g.getDefaultExecutionModeContent(),
		"model-override":          g.getDefaultModelOverrideContent(),
		"agent-routing":           g.getDefaultAgentRoutingContent(),
		"commands":                g.getDefaultCommandsContent(),
		"platform-infrastructure": g.getDefaultPlatformInfrastructureContent(),
		"quick-start":             g.getDefaultQuickStartContent(),
		"agent-configurations":    g.getDefaultAgentConfigsContent(),
	}

	if content, ok := defaults[regionName]; ok {
		return content, nil
	}

	return "", errors.NewWithDetails(errors.CodeFileNotFound,
		"no template found for region",
		map[string]any{"region": regionName})
}

// SetSectionTemplate sets an inline template for a section.
func (g *Generator) SetSectionTemplate(regionName, template string) {
	g.sectionTemplates[regionName] = template
}

// Default content generators for each section type

func (g *Generator) getDefaultModelOverrideContent() string {
	if g.Context == nil || g.Context.ModelOverride == "" {
		return ""
	}
	return `## Model Override

All agents forced to **` + g.Context.ModelOverride + `** (el-cheapo mode). Ephemeral -- reverts on session exit.`
}

func (g *Generator) getDefaultExecutionModeContent() string {
	isGemini := g.Context != nil && g.Context.Channel == "gemini"
	if g.Context != nil && !g.Context.IsKnossosProject {
		if isGemini {
			return `## Execution Mode

Use the available agents and slash commands. Agents activate automatically when your prompt matches their description.`
		}
		return `## Execution Mode

Use the available agents and slash commands. Delegate complex work to specialists via Task tool.`
	}
	if isGemini {
		return `## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Potnia consulted via description matching |

Use ` + "`/go`" + ` to start any session. Use ` + "`/consult`" + ` for mode selection.`
	}
	return `## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Potnia coordinates; delegate via Task tool |

Use ` + "`/go`" + ` to start any session. Use ` + "`/consult`" + ` for mode selection.`
}

func (g *Generator) getDefaultAgentRoutingContent() string {
	isGemini := g.Context != nil && g.Context.Channel == "gemini"
	if g.Context != nil && !g.Context.IsKnossosProject {
		if isGemini {
			return `## Agent Routing

Agents activate automatically based on description matching. Write prompts that align with specialist descriptions for effective routing.`
		}
		return `## Agent Routing

Delegate to specialists via Task tool.
Agents cannot spawn agents — only the main thread has Task tool access.`
	}
	if isGemini {
		return `## Agent Routing

Agents activate automatically based on description matching. Write prompts that align with specialist descriptions for effective routing.
Without a session, execute directly. Routing guidance: ` + "`/consult`" + `.`
	}
	return `## Agent Routing

Delegate to specialists via Task tool. Potnia coordinates phases and handoffs.
Agents cannot spawn agents — only the main thread has Task tool access.
Without a session, execute directly or use ` + "`/task`" + `. Routing guidance: ` + "`/consult`" + `.`
}

func (g *Generator) getDefaultCommandsContent() string {
	isGemini := g.Context != nil && g.Context.Channel == "gemini"
	if g.Context != nil && !g.Context.IsKnossosProject {
		if isGemini {
			return `## Gemini Primitives

| Primitive | Invocation | Source |
|---|---|---|
| Slash command | User types ` + "`/name`" + ` | ` + "`.gemini/commands/`" + ` |
| Skill | Loaded into context | ` + "`.gemini/skills/`" + ` |
| Agent | Activates on description match | ` + "`.gemini/agents/`" + ` |
| Hook | Auto-fires on lifecycle events | ` + "`.gemini/settings.local.json`" + ` |

Agents cannot spawn other agents — only the main thread can dispatch sub-agents.`
		}
		return `## CC Primitives

| CC Primitive | Invocation | Source |
|---|---|---|
| Slash command | User types ` + "`/name`" + ` | ` + "`.claude/commands/`" + ` |
| Skill tool | Model calls ` + "`Skill(\"name\")`" + ` | ` + "`.claude/skills/`" + ` |
| Task tool | Model calls ` + "`Task(subagent_type)`" + ` | ` + "`.claude/agents/`" + ` |
| Hook | Auto-fires on lifecycle events | ` + "`.claude/settings.json`" + ` |

Agents cannot spawn other agents — only the main thread has Task tool access.`
	}
	if isGemini {
		return `## Gemini Primitives

| Primitive | Knossos Name | Invocation | Source |
|---|---|---|---|
| Slash command | **Dromena** | User types ` + "`/name`" + ` | ` + "`.gemini/commands/`" + ` |
| Skill | **Legomena** | Loaded into context | ` + "`.gemini/skills/`" + ` |
| Agent | **Agent** | Activates on description match | ` + "`.gemini/agents/`" + ` |
| Hook | **Hook** | Auto-fires on lifecycle events | ` + "`.gemini/settings.local.json`" + ` |
| GEMINI.md | **Inscription** | Always in context | ` + "`knossos/templates/`" + ` |

Agents cannot spawn other agents — only the main thread can dispatch sub-agents.`
	}
	return `## CC Primitives

| CC Primitive | Knossos Name | Invocation | Source |
|---|---|---|---|
| Slash command | **Dromena** | User types ` + "`/name`" + ` | ` + "`.claude/commands/`" + ` |
| Skill tool | **Legomena** | Model calls ` + "`Skill(\"name\")`" + ` | ` + "`.claude/skills/`" + ` |
| Task tool | **Agent** | Model calls ` + "`Task(subagent_type)`" + ` | ` + "`.claude/agents/`" + ` |
| Hook | **Hook** | Auto-fires on lifecycle events | ` + "`.claude/settings.json`" + ` |
| CLAUDE.md | **Inscription** | Always in context | ` + "`knossos/templates/`" + ` |

Agents cannot spawn other agents — only the main thread has Task tool access.`
}

func (g *Generator) getDefaultPlatformInfrastructureContent() string {
	isGemini := g.Context != nil && g.Context.Channel == "gemini"
	if g.Context != nil && !g.Context.IsKnossosProject {
		// Non-knossos: same content for both channels
		return `## Platform

CLI reference: ` + "`ari --help`" + `.`
	}
	if isGemini {
		return `## Platform

**Entry**: ` + "`/go`" + ` — detects session state, resumes parked work, or routes new tasks.

**Sessions**: ` + "`/sos`" + ` (start, park, resume, wrap), ` + "`/handoff`" + `, ` + "`/fray`" + `. Mutate ` + "`*_CONTEXT.md`" + ` only via the moirai agent.

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: ` + "`ari --help`" + `.`
	}
	return `## Platform

**Entry**: ` + "`/go`" + ` — detects session state, resumes parked work, or routes new tasks.

**Sessions**: ` + "`/start`" + `, ` + "`/park`" + `, ` + "`/continue`" + `, ` + "`/wrap`" + `. Mutate ` + "`*_CONTEXT.md`" + ` only via ` + "`Task(moirai, \"...\")`" + `.

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: ` + "`ari --help`" + `.`
}

func (g *Generator) getDefaultQuickStartContent() string {
	if g.Context != nil && !g.Context.IsKnossosProject {
		return `## Quick Start

No active rite. Use ` + "`ari sync --rite=<name>`" + ` to activate a rite.`
	}
	return `## Quick Start

No active rite. Use ` + "`/go`" + ` to get started, or ` + "`ari sync --rite=<name>`" + ` to activate directly.`
}

func (g *Generator) getDefaultAgentConfigsContent() string {
	if g.Context != nil && g.Context.Channel == "gemini" {
		return `## Agents

Prompts in ` + "`.gemini/agents/`" + `.`
	}
	return `## Agents

Prompts in ` + "`.claude/agents/`" + `.`
}

// GenerateAll generates content for all sections in section_order.
// Non-satellite regions that the generator cannot produce (removed templates,
// deprecated sections) are silently skipped — the merger will clean them from
// the manifest during the merge phase.
func (g *Generator) GenerateAll() (map[string]string, error) {
	result := make(map[string]string)

	for _, regionName := range g.Manifest.SectionOrder {
		region := g.Manifest.GetRegion(regionName)
		if region == nil {
			// Skip regions not in manifest
			continue
		}

		// Skip non-satellite regions the generator can't produce (deprecated/removed).
		// Satellite regions without templates return empty string, which is fine.
		if region.Owner != OwnerSatellite && !g.CanGenerateRegion(regionName) {
			continue
		}

		content, err := g.GenerateSection(regionName)
		if err != nil {
			return nil, err
		}

		result[regionName] = content
	}

	return result, nil
}

// CanGenerateRegion returns true if the generator can produce content for the
// named region via template file, inline template, or Go default.
func (g *Generator) CanGenerateRegion(name string) bool {
	if g.getSectionTemplatePath(name) != "" {
		return true
	}
	if _, ok := g.sectionTemplates[name]; ok {
		return true
	}
	_, err := g.getDefaultSectionContent(name)
	return err == nil
}
