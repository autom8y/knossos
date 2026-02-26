package materialize

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
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
	ExousiaYouDecide    string
	ExousiaYouEscalate  string
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

// RenderArchetype loads a template from knossos/archetypes/ relative to projectRoot
// and renders it with the provided data.
func RenderArchetype(projectRoot, templateName string, data interface{}) ([]byte, error) {
	tplPath := filepath.Join(projectRoot, "knossos", "archetypes", templateName)

	tplContent, err := os.ReadFile(tplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archetype template %s: %w", templateName, err)
	}

	return RenderArchetypeFromString(string(tplContent), templateName, data)
}

// RenderArchetypeFromString parses and renders a template from a raw string.
// Useful for testing without filesystem dependency.
func RenderArchetypeFromString(tplContent, templateName string, data interface{}) ([]byte, error) {
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
