package inscription

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/autom8y/knossos/internal/errors"
)

// AgentInfo holds agent metadata for template rendering.
// This mirrors team.AgentInfo for the inscription package.
type AgentInfo struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Role     string `json:"role"`
	Produces string `json:"produces"`
}

// RenderContext provides data for template rendering.
type RenderContext struct {
	// ActiveRite is the current rite name.
	ActiveRite string

	// AgentCount is the number of agents in the active rite.
	AgentCount int

	// Agents contains metadata for each agent.
	Agents []AgentInfo

	// KnossosVars contains additional variables for template rendering.
	KnossosVars map[string]string

	// ProjectRoot is the root directory of the project.
	ProjectRoot string
}

// Generator handles content generation for CLAUDE.md regions.
type Generator struct {
	// TemplateDir is the directory containing template files.
	TemplateDir string

	// Manifest is the current KNOSSOS_MANIFEST.yaml.
	Manifest *Manifest

	// Context is the render context for template execution.
	Context *RenderContext

	// templates is the parsed template cache.
	templates *template.Template

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

// GenerateSection generates content for a region based on its owner type.
// Returns UNWRAPPED content only (no START/END markers).
// Use RenderRegion() for wrapped content suitable for direct output.
// Use this method when preparing content for the merger (MergeRegions).
func (g *Generator) GenerateSection(regionName string) (string, error) {
	region := g.Manifest.GetRegion(regionName)
	if region == nil {
		return "", errors.NewWithDetails(errors.CodeFileNotFound,
			"region not found in manifest",
			map[string]interface{}{"region": regionName})
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
			map[string]interface{}{"region": regionName, "owner": string(region.Owner)})
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
func (g *Generator) getSectionTemplatePath(regionName string) string {
	if g.TemplateDir == "" {
		return ""
	}

	// Check for section-specific template
	path := filepath.Join(g.TemplateDir, "sections", regionName+".md.tpl")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// renderTemplateFile renders a template from a file path.
func (g *Generator) renderTemplateFile(path string) (string, error) {
	data, err := os.ReadFile(path)
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
			map[string]interface{}{"name": name, "cause": err.Error()})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.Context); err != nil {
		return "", errors.NewWithDetails(errors.CodeGeneralError,
			"failed to execute template",
			map[string]interface{}{"name": name, "cause": err.Error()})
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

	// Note: Sprig already provides join, lower, upper, title, and many more

	return funcs
}

// includePartial loads and renders a partial template.
func (g *Generator) includePartial(partialPath string) (string, error) {
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
		return "| Agent | Role | Produces |\n| ----- | ---- | -------- |\n"
	}

	var sb strings.Builder
	sb.WriteString("| Agent | Role | Produces |\n")
	sb.WriteString("| ----- | ---- | -------- |\n")

	for _, agent := range g.Context.Agents {
		sb.WriteString("| **")
		sb.WriteString(agent.Name)
		sb.WriteString("** | ")
		sb.WriteString(agent.Role)
		sb.WriteString(" | ")
		sb.WriteString(agent.Produces)
		sb.WriteString(" |\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// lookupTerminology returns the definition for a term.
func (g *Generator) lookupTerminology(term string) string {
	// Default terminology definitions
	terminology := map[string]string{
		"knossos":     "The platform (the labyrinth itself)",
		"ariadne":     "CLI binary (`ari`) - the clew ensuring return",
		"theseus":     "Claude Code agent - the navigator with amnesia",
		"moirai":      "Session lifecycle agent - the Fates who spin, measure, and cut",
		"white-sails": "Confidence signal - honest return indicator",
		"rites":       "Practice bundles - invokable ceremonies",
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
			map[string]interface{}{"region": regionName, "source": source})
	}
}

// generateQuickStartContent generates the Quick Start section content.
func (g *Generator) generateQuickStartContent() (string, error) {
	if g.Context == nil {
		return g.getDefaultQuickStartContent(), nil
	}

	var sb strings.Builder

	// Header with team info
	sb.WriteString("## Quick Start\n\n")

	if g.Context.ActiveRite != "" {
		sb.WriteString("This project uses a ")
		sb.WriteString(itoa(g.Context.AgentCount))
		sb.WriteString("-agent workflow (")
		sb.WriteString(g.Context.ActiveRite)
		sb.WriteString("):\n\n")
	} else {
		sb.WriteString("This project uses a multi-agent workflow:\n\n")
	}

	// Agent table
	sb.WriteString(g.loadAgentTable())
	sb.WriteString("\n\n")

	// Footer
	sb.WriteString("Use `prompting` for agent invocation patterns. Use `initiative-scoping` for new projects.")

	return sb.String(), nil
}

// generateAgentConfigsContent generates the Agent Configurations section.
func (g *Generator) generateAgentConfigsContent() (string, error) {
	if g.Context == nil || len(g.Context.Agents) == 0 {
		return g.getDefaultAgentConfigsContent(), nil
	}

	var sb strings.Builder
	sb.WriteString("## Agents\n\n")
	sb.WriteString("Prompts in `.claude/agents/`:\n\n")

	for _, agent := range g.Context.Agents {
		sb.WriteString("- `")
		sb.WriteString(agent.File)
		sb.WriteString("` - ")
		sb.WriteString(agent.Role)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// getDefaultSectionContent returns default content for known sections.
func (g *Generator) getDefaultSectionContent(regionName string) (string, error) {
	defaults := map[string]string{
		"execution-mode":          g.getDefaultExecutionModeContent(),
		"agent-routing":           g.getDefaultAgentRoutingContent(),
		"commands":                g.getDefaultCommandsContent(),
		"skills":                  g.getDefaultCommandsContent(), // Alias for backward compatibility
		"platform-infrastructure": g.getDefaultPlatformInfrastructureContent(),
		"navigation":              g.getDefaultNavigationContent(),
		"slash-commands":          g.getDefaultSlashCommandsContent(),
		"quick-start":             g.getDefaultQuickStartContent(),
		"agent-configurations":    g.getDefaultAgentConfigsContent(),
	}

	if content, ok := defaults[regionName]; ok {
		return content, nil
	}

	return "", errors.NewWithDetails(errors.CodeFileNotFound,
		"no template found for region",
		map[string]interface{}{"region": regionName})
}

// SetSectionTemplate sets an inline template for a section.
func (g *Generator) SetSectionTemplate(regionName, template string) {
	g.sectionTemplates[regionName] = template
}

// Default content generators for each section type

func (g *Generator) getDefaultExecutionModeContent() string {
	return `## Execution Mode

Three operating modes:

| Mode | Session | Rite | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Coach pattern, delegate via Task tool |

Use ` + "`/consult`" + ` for mode selection. Enforcement rules: ` + "`orchestration/execution-mode.md`"
}


func (g *Generator) getDefaultAgentRoutingContent() string {
	return `## Agent Routing

In orchestrated sessions, delegate to specialists via Task tool. Without a session, execute directly or use ` + "`/task`" + `.

Routing guidance: ` + "`/consult`"
}

func (g *Generator) getDefaultCommandsContent() string {
	return `## Commands

Invoke via the **Skill tool**. Two types:

- **Invokable** (` + "`/name`" + `): User-callable actions (` + "`/start`" + `, ` + "`/commit`" + `, ` + "`/pr`" + `)
- **Reference** (Skill tool): Domain knowledge (` + "`prompting`" + `, ` + "`doc-artifacts`" + `, ` + "`standards`" + `)

Full list: ` + "`.claude/commands/`" + ` and ` + "`.claude/skills/`"
}


func (g *Generator) getDefaultPlatformInfrastructureContent() string {
	return `## Platform

Hooks auto-inject session context. CLI reference: ` + "`ari --help`" + `.
Mutate ` + "`*_CONTEXT.md`" + ` only via ` + "`Task(moirai, \"...\")`" + `.`
}

func (g *Generator) getDefaultNavigationContent() string {
	return `## Navigation

Workflow routing: ` + "`/consult`" + `. Domain knowledge: Skill tool. File locations: ` + "`MEMORY.md`" + `.`
}

func (g *Generator) getDefaultSlashCommandsContent() string {
	return `## Slash Commands

Always respond with outcome. "No response" is never correct for explicit user requests.`
}

func (g *Generator) getDefaultQuickStartContent() string {
	return `## Quick Start

This project uses a multi-agent workflow:

| Agent | Role | Produces |
| ----- | ---- | -------- |

Use ` + "`prompting`" + ` for agent invocation patterns. Use ` + "`initiative-scoping`" + ` for new projects.`
}

func (g *Generator) getDefaultAgentConfigsContent() string {
	return `## Agents

Prompts in ` + "`.claude/agents/`" + `.`
}

// GenerateAll generates content for all sections in section_order.
func (g *Generator) GenerateAll() (map[string]string, error) {
	result := make(map[string]string)

	for _, regionName := range g.Manifest.SectionOrder {
		region := g.Manifest.GetRegion(regionName)
		if region == nil {
			// Skip regions not in manifest
			continue
		}

		// Generate all regions including satellite
		// For satellite regions, the merger will preserve existing content if present,
		// but for new files we need the initial template content
		content, err := g.GenerateSection(regionName)
		if err != nil {
			return nil, err
		}

		result[regionName] = content
	}

	return result, nil
}
