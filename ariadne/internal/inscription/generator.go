package inscription

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/autom8y/ariadne/internal/errors"
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
	// ActiveRite is the current rite/team name.
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
		// Satellite regions are pass-through; content comes from existing file
		return "", nil
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

	// For satellite regions, return empty (will be preserved from existing file)
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
	return template.FuncMap{
		"include": g.includePartial,
		"ifdef":   g.conditionalInclude,
		"agents":  g.loadAgentTable,
		"term":    g.lookupTerminology,
		"join":    strings.Join,
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"title":   strings.Title,
	}
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
		"ariadne":     "CLI binary (`ari`) - the thread ensuring return",
		"theseus":     "Claude Code agent - the navigator with amnesia",
		"moirai":      "Session lifecycle agent - the Fates who spin, measure, and cut",
		"white-sails": "Confidence signal - honest return indicator",
		"rites":       "Practice bundles - invokable ceremonies (formerly 'team packs')",
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
	sb.WriteString("**New here?** Use the `prompting` skill for copy-paste patterns, or `initiative-scoping` to start a new project.")

	return sb.String(), nil
}

// generateAgentConfigsContent generates the Agent Configurations section.
func (g *Generator) generateAgentConfigsContent() (string, error) {
	if g.Context == nil || len(g.Context.Agents) == 0 {
		return g.getDefaultAgentConfigsContent(), nil
	}

	var sb strings.Builder
	sb.WriteString("## Agent Configurations\n\n")
	sb.WriteString("Full agent prompts live in `.claude/agents/`:\n\n")

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
		"execution-mode":     g.getDefaultExecutionModeContent(),
		"knossos-identity":   g.getDefaultKnossosIdentityContent(),
		"agent-routing":      g.getDefaultAgentRoutingContent(),
		"skills":             g.getDefaultSkillsContent(),
		"hooks":              g.getDefaultHooksContent(),
		"dynamic-context":    g.getDefaultDynamicContextContent(),
		"ariadne-cli":        g.getDefaultAriadneCliContent(),
		"getting-help":       g.getDefaultGettingHelpContent(),
		"state-management":   g.getDefaultStateManagementContent(),
		"slash-commands":     g.getDefaultSlashCommandsContent(),
		"quick-start":        g.getDefaultQuickStartContent(),
		"agent-configurations": g.getDefaultAgentConfigsContent(),
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

This project supports three operating modes (see PRD-hybrid-session-model for details):

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Coach pattern, delegate via Task tool |

**Unsure?** Use ` + "`/consult`" + ` for workflow routing.

For enforcement rules: ` + "`orchestration/execution-mode.md`"
}

func (g *Generator) getDefaultKnossosIdentityContent() string {
	return `## Knossos Identity

> **roster/.claude/ IS Knossos.** This repository is the Knossos platform.

The naming reflects Greek mythology (see ` + "`docs/philosophy/knossos-doctrine.md`" + ` for the full doctrine):

| Myth | Component | Function |
|------|-----------|----------|
| **Knossos** | The platform | The labyrinth itself |
| **Ariadne** | CLI binary (` + "`ari`" + `) | The thread ensuring return |
| **Theseus** | Claude Code agent | The navigator with amnesia |
| **Moirai** | Session lifecycle agent | The Fates who spin, measure, and cut |
| **White Sails** | Confidence signal | Honest return indicator |
| **Rites** | Practice bundles | Invokable ceremonies (formerly "team packs") |

For full details: ` + "`docs/guides/knossos-integration.md`" + ` and ` + "`docs/decisions/ADR-0009-knossos-roster-identity.md`"
}

func (g *Generator) getDefaultAgentRoutingContent() string {
	return `## Agent Routing

When working within an orchestrated session, the main thread coordinates via Task tool delegation to specialist agents. Without an active session, direct execution or ` + "`/task`" + ` initialization are both valid approaches.

For routing guidance: ` + "`/consult`"
}

func (g *Generator) getDefaultSkillsContent() string {
	return `## Skills

Skills are invoked via the **Skill tool**. Key skills: ` + "`orchestration`" + ` (workflow coordination), ` + "`documentation`" + ` (templates), ` + "`prompting`" + ` (agent invocation), ` + "`standards`" + ` (conventions), ` + "`ecosystem-ref`" + ` (roster ecosystem patterns). See ` + "`.claude/skills/`" + ` and ` + "`~/.claude/skills/`" + ` for full list.`
}

func (g *Generator) getDefaultHooksContent() string {
	return `## Hooks

Hooks auto-inject context (SessionStart, Stop, PostToolUse). No manual context needed. See ` + "`.claude/hooks/`" + `.`
}

func (g *Generator) getDefaultDynamicContextContent() string {
	return `## Dynamic Context

Commands use ` + "`!`" + ` prefix for live context: ` + "`` `!`cat .claude/ACTIVE_RITE` ``" + `. Prefer hooks for complex context.`
}

func (g *Generator) getDefaultAriadneCliContent() string {
	return `## Ariadne CLI

The ` + "`ari`" + ` binary provides session and hook operations:

` + "```bash" + `
# Session management
ari session create "initiative" COMPLEXITY
ari session status
ari session park "reason"

# Hook operations
ari hook thread
ari hook context

# Quality gates
ari sails check

# Agent handoffs
ari handoff prepare --from <agent> --to <agent>
ari handoff execute --from <agent> --to <agent>
ari handoff status
ari handoff history
` + "```" + `

### Cognitive Budget

Tool usage tracking with configurable thresholds:
- ` + "`ARIADNE_MSG_WARN=250`" + ` - Warning threshold (default)
- ` + "`ARIADNE_MSG_PARK`" + ` - Park suggestion threshold
- ` + "`ARIADNE_BUDGET_DISABLE=1`" + ` - Disable tracking

Build: ` + "`cd ariadne && just build`" + `

Full reference: ` + "`docs/guides/knossos-integration.md`"
}

func (g *Generator) getDefaultGettingHelpContent() string {
	return `## Getting Help

| Question | Skill |
|----------|-------|
| Invoke agents | ` + "`prompting`" + ` |
| Templates | ` + "`documentation`" + ` or ` + "`doc-ecosystem`" + ` |
| Conventions | ` + "`standards`" + ` |
| Workflow coordination | ` + "`orchestration`" + ` |
| Roster ecosystem | ` + "`ecosystem-ref`" + ` |
| User preferences | See ` + "`docs/guides/user-preferences.md`" + ` |
| Knossos integration | ` + "`docs/guides/knossos-integration.md`" + ` |
| Migration path | ` + "`docs/guides/knossos-migration.md`" + ` |
| Unsure where to start | ` + "`/consult`" + ` |`
}

func (g *Generator) getDefaultStateManagementContent() string {
	return `## State Management

**Mutating session/sprint state?** Use the **Moirai** (the Fates) for all ` + "`SESSION_CONTEXT.md`" + ` and ` + "`SPRINT_CONTEXT.md`" + ` changes.

### Moirai Usage

The Moirai are the centralized authority for session lifecycle--spinning sessions into existence (Clotho), measuring their allotment (Lachesis), and cutting when complete (Atropos). They enforce schema validation, lifecycle transitions, and maintain audit trails.

**When to Use**:
- Updating session state (park, resume, wrap)
- Marking tasks complete
- Transitioning workflow phases
- Creating or managing sprints
- Any modification to ` + "`*_CONTEXT.md`" + ` files
- Generating White Sails confidence signals

**Invocation Pattern** (requires session context):
` + "```" + `
Task(moirai, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md

Session Context:
- Session ID: {from session-manager.sh status}
- Session Path: .claude/sessions/{session-id}/SESSION_CONTEXT.md")
` + "```" + `

Get session context: ` + "`.claude/hooks/lib/session-manager.sh status | jq -r '.session_id'`" + `

**Natural Language Supported**:
` + "```" + `
Task(moirai, "Mark the PRD task complete with artifact at docs/requirements/PRD-foo.md")
` + "```" + `

**Control Flags**:
- ` + "`--dry-run`" + `: Preview changes without applying
- ` + "`--emergency`" + `: Bypass non-critical validations (logged)
- ` + "`--override=reason`" + `: Bypass lifecycle rules with explicit reason

**Direct writes blocked**: PreToolUse hook intercepts ` + "`Write`" + `/` + "`Edit`" + ` to ` + "`*_CONTEXT.md`" + ` and instructs use of Moirai.

**Full documentation**: See ` + "`user-agents/moirai.md`" + ` and ` + "`docs/philosophy/knossos-doctrine.md`"
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

**New here?** Use the ` + "`prompting`" + ` skill for copy-paste patterns, or ` + "`initiative-scoping`" + ` to start a new project.`
}

func (g *Generator) getDefaultAgentConfigsContent() string {
	return `## Agent Configurations

Full agent prompts live in ` + "`.claude/agents/`" + `.`
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

		// Skip satellite regions (they're preserved, not generated)
		if region.Owner == OwnerSatellite {
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
