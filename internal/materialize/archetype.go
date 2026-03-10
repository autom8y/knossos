package materialize

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/autom8y/knossos/internal/config"
)

// OrchestratorData provides template variables for the orchestrator archetype.
type OrchestratorData struct {
	RiteName          string
	Description       string
	Color             string
	Skills            []string // All skills including orchestrator-templates
	ContractMustNot   []string // contract.must_not entries; defaults applied if empty
	PhaseRouting      string   // Pre-formatted markdown table rows (| specialist | route when |)
	HandoffCriteria   string   // Pre-formatted markdown table rows
	RiteAntiPatterns  string   // Pre-formatted bullet list
	CrossRiteProtocol string   // Markdown content or "<!-- TODO: Define how cross-rite concerns are routed and resolved -->"
	EntryPointSection string   // Optional: full markdown for rite-specific entry point section (10x-dev only)
	CustomSections    string   // Any rite-unique sections (arch back-routes, slop-chop artifact chain, etc.)

	// Exousia overrides — empty strings use the boilerplate defaults in the template.
	ExousiaYouDecide      string
	ExousiaYouEscalate    string
	ExousiaYouDoNotDecide string

	// Optional overrides for sections that are mostly identical but occasionally differ.
	ToolAccessSection          string // Override entire Tool Access section body (ecosystem has custom content)
	ConsultationProtocolInput  string // Override Input subsection (ecosystem has custom content)
	ConsultationProtocolOutput string // Override Output subsection (ecosystem has custom content)
	PositionInWorkflow         string // Override Position in Workflow content (ecosystem has ASCII diagram)
	CoreResponsibilities       string // Override Core Responsibilities bullet list (arch has extra bullet)
	SkillsReference            string // Override Skills Reference content (ecosystem has different skills)
	BehavioralConstraintsDO    string // Override second Behavioral Constraints (DO NOT) section wording
}

// defaultContractMustNot returns the standard contract.must_not entries used by most orchestrators.
func defaultContractMustNot() []string {
	return []string{
		"Execute work directly instead of generating specialist directives",
		"Use tools beyond Read",
		"Respond with prose instead of CONSULTATION_RESPONSE format",
	}
}

// renderArchetypeAgent constructs archetype-specific data from the manifest and
// renders the corresponding template. Currently only the "orchestrator" archetype
// is supported; unknown archetypes return an error.
//
// The render callback controls template resolution. Production callers pass
// Materializer.renderArchetypeResolved (DI-aware); tests may pass RenderArchetype.
func renderArchetypeAgent(projectRoot string, agent Agent, manifest *RiteManifest, render func(string, string, any) ([]byte, error)) ([]byte, error) {
	switch agent.Archetype {
	case "orchestrator":
		data := buildOrchestratorData(agent, manifest)
		return render(projectRoot, "orchestrator.md.tpl", data)
	default:
		return nil, fmt.Errorf("unknown archetype: %s", agent.Archetype)
	}
}

// buildOrchestratorData constructs an OrchestratorData from the manifest's
// ArchetypeData["orchestrator"] section combined with top-level manifest fields.
//
// Manifest-level fields provide RiteName and Description.
// The archetype_data.orchestrator map provides all template-specific fields:
// color, skills, phase_routing, handoff_criteria, rite_anti_patterns,
// cross_rite_protocol, entry_point_section, custom_sections, and all
// optional override fields.
func buildOrchestratorData(agent Agent, manifest *RiteManifest) *OrchestratorData {
	data := &OrchestratorData{
		RiteName: manifest.Name,
	}

	raw, ok := manifest.ArchetypeData["orchestrator"]
	if !ok {
		return data
	}

	// Helper to extract a string field from the raw map.
	str := func(key string) string {
		if v, ok := raw[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	// Helper to extract a []string from the raw map (YAML list).
	strSlice := func(key string) []string {
		v, ok := raw[key]
		if !ok {
			return nil
		}
		switch typed := v.(type) {
		case []any:
			result := make([]string, 0, len(typed))
			for _, item := range typed {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		case []string:
			return typed
		default:
			return nil
		}
	}

	data.Description = str("description")
	data.Color = str("color")
	data.Skills = strSlice("skills")
	data.ContractMustNot = strSlice("contract_must_not")
	data.PhaseRouting = str("phase_routing")
	data.HandoffCriteria = str("handoff_criteria")
	data.RiteAntiPatterns = str("rite_anti_patterns")
	data.CrossRiteProtocol = str("cross_rite_protocol")
	data.EntryPointSection = str("entry_point_section")
	data.CustomSections = str("custom_sections")

	// Exousia overrides
	data.ExousiaYouDecide = str("exousia_you_decide")
	data.ExousiaYouEscalate = str("exousia_you_escalate")
	data.ExousiaYouDoNotDecide = str("exousia_you_do_not_decide")

	// Optional section overrides
	data.ToolAccessSection = str("tool_access_section")
	data.ConsultationProtocolInput = str("consultation_protocol_input")
	data.ConsultationProtocolOutput = str("consultation_protocol_output")
	data.PositionInWorkflow = str("position_in_workflow")
	data.CoreResponsibilities = str("core_responsibilities")
	data.SkillsReference = str("skills_reference")
	data.BehavioralConstraintsDO = str("behavioral_constraints_do")

	return data
}

// RenderArchetype loads a template from knossos/archetypes/ and renders it.
// Resolution order: projectRoot → KnossosHome → embedded FS.
func RenderArchetype(projectRoot, templateName string, data any) ([]byte, error) {
	relPath := filepath.Join("knossos", "archetypes", templateName)

	// 1. Try project root (knossos-on-knossos case)
	tplPath := filepath.Join(projectRoot, relPath)
	if tplContent, err := os.ReadFile(tplPath); err == nil {
		return RenderArchetypeFromString(string(tplContent), templateName, data)
	}

	// 2. Try KnossosHome (developer case, foreign project)
	if knossosHome := config.KnossosHome(); knossosHome != "" {
		tplPath = filepath.Join(knossosHome, relPath)
		if tplContent, err := os.ReadFile(tplPath); err == nil {
			return RenderArchetypeFromString(string(tplContent), templateName, data)
		}
	}

	// 3. Try XDG data dir (installed user case)
	xdgPath := filepath.Join(config.XDGDataDir(), "archetypes", templateName)
	if tplContent, err := os.ReadFile(xdgPath); err == nil {
		return RenderArchetypeFromString(string(tplContent), templateName, data)
	}

	return nil, fmt.Errorf("archetype template %s not found in project, KnossosHome, or XDG data dir", templateName)
}

// renderArchetypeResolved loads a template using Materializer path fields.
// Uses m.knossosHome and m.xdgDataDir with fallback to source resolver / config globals.
// This avoids direct config.KnossosHome() calls in the materialize pipeline.
func (m *Materializer) renderArchetypeResolved(projectRoot, templateName string, data any) ([]byte, error) {
	relPath := filepath.Join("knossos", "archetypes", templateName)

	// 1. Try project root (knossos-on-knossos case)
	tplPath := filepath.Join(projectRoot, relPath)
	if tplContent, err := os.ReadFile(tplPath); err == nil {
		return RenderArchetypeFromString(string(tplContent), templateName, data)
	}

	// 2. Try KnossosHome (prefer struct field, fall back to source resolver)
	knossosHome := m.knossosHome
	if knossosHome == "" {
		knossosHome = m.sourceResolver.KnossosHome()
	}
	if knossosHome != "" {
		tplPath = filepath.Join(knossosHome, relPath)
		if tplContent, err := os.ReadFile(tplPath); err == nil {
			return RenderArchetypeFromString(string(tplContent), templateName, data)
		}
	}

	// 3. Try XDG data dir (prefer struct field, fall back to config)
	xdgDataDir := m.xdgDataDir
	if xdgDataDir == "" {
		xdgDataDir = config.XDGDataDir()
	}
	xdgPath := filepath.Join(xdgDataDir, "archetypes", templateName)
	if tplContent, err := os.ReadFile(xdgPath); err == nil {
		return RenderArchetypeFromString(string(tplContent), templateName, data)
	}

	return nil, fmt.Errorf("archetype template %s not found in project, KnossosHome, or XDG data dir", templateName)
}

// RenderArchetypeFromString parses and renders a template from a raw string.
// Useful for testing without filesystem dependency.
func RenderArchetypeFromString(tplContent, templateName string, data any) ([]byte, error) {
	tmpl, err := template.New(templateName).Parse(tplContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse archetype template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render archetype template %s: %w", templateName, err)
	}

	return buf.Bytes(), nil
}
