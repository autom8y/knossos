package materialize

import (
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// OrchestratorConfig represents the parsed orchestrator.yaml file.
// This is the structured configuration that rite authors maintain alongside
// manifest.yaml. The materialization pipeline reads it and converts it to
// the flat map[string]any format expected by buildOrchestratorData().
type OrchestratorConfig struct {
	Rite struct {
		Name   string `yaml:"name"`
		Domain string `yaml:"domain"`
		Color  string `yaml:"color"`
	} `yaml:"rite"`
	Frontmatter struct {
		Role        string `yaml:"role"`
		Description string `yaml:"description"`
	} `yaml:"frontmatter"`
	Routing          map[string]string   `yaml:"routing"`
	WorkflowPosition struct {
		Upstream   string `yaml:"upstream"`
		Downstream string `yaml:"downstream"`
	} `yaml:"workflow_position"`
	HandoffCriteria   map[string][]string `yaml:"handoff_criteria"`
	Skills            []string            `yaml:"skills"`
	Antipatterns      []string            `yaml:"antipatterns"`
	CrossRiteProtocol string              `yaml:"cross_rite_protocol"`
}

// loadOrchestratorConfig reads orchestrator.yaml from the given fs.FS.
// Returns nil and an error if the file is missing or malformed.
func loadOrchestratorConfig(riteFS fs.FS) (*OrchestratorConfig, error) {
	data, err := fs.ReadFile(riteFS, "orchestrator.yaml")
	if err != nil {
		return nil, err
	}

	var config OrchestratorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse orchestrator.yaml: %w", err)
	}

	return &config, nil
}

// convertOrchestratorConfigToData transforms a structured OrchestratorConfig
// into the flat map[string]any format consumed by buildOrchestratorData().
// Field keys match the ArchetypeData schema used by the orchestrator archetype.
//
// Markdown formatting follows the conventions established in archetype_test.go
// test fixtures (tenxDevData, forgeData, etc.):
//   - phase_routing: "| specialist | condition |\n" rows (sorted by key)
//   - handoff_criteria: "| phase | - item1<- item2< |\n" rows (ordered by routing keys)
//   - rite_anti_patterns: "- **pattern**\n" bullets
func convertOrchestratorConfigToData(config *OrchestratorConfig) map[string]any {
	data := make(map[string]any)

	if config.Frontmatter.Description != "" {
		// Collapse multi-line descriptions to a single line. The archetype template
		// uses `description: |\n  {{.Description}}` which only indents the first line;
		// subsequent lines would spill out as top-level YAML keys.
		data["description"] = collapseLines(config.Frontmatter.Description)
	}
	if config.Rite.Color != "" {
		data["color"] = config.Rite.Color
	}
	if len(config.Skills) > 0 {
		// Extract bare skill names. orchestrator.yaml uses documentation strings
		// like "review-ref for methodology, severity model, ..." — we need just
		// the skill name (text before the first space).
		skills := make([]any, len(config.Skills))
		for i, s := range config.Skills {
			skills[i] = extractSkillName(s)
		}
		data["skills"] = skills
	}
	if config.CrossRiteProtocol != "" {
		data["cross_rite_protocol"] = strings.TrimSpace(config.CrossRiteProtocol)
	}

	// Routing → phase_routing markdown table rows (sorted by specialist name)
	if len(config.Routing) > 0 {
		data["phase_routing"] = formatRoutingTable(config.Routing)
	}

	// Handoff criteria → markdown table rows (ordered by routing keys for consistency)
	if len(config.HandoffCriteria) > 0 {
		data["handoff_criteria"] = formatHandoffTable(config.HandoffCriteria, config.Routing)
	}

	// Antipatterns → bullet list
	if len(config.Antipatterns) > 0 {
		data["rite_anti_patterns"] = formatAntipatterns(config.Antipatterns)
	}

	// Workflow position → formatted string
	if config.WorkflowPosition.Upstream != "" || config.WorkflowPosition.Downstream != "" {
		data["position_in_workflow"] = fmt.Sprintf("**Upstream**: %s\n**Downstream**: %s",
			config.WorkflowPosition.Upstream, config.WorkflowPosition.Downstream)
	}

	return data
}

// collapseLines joins a multi-line string into a single line, replacing
// newlines with spaces and collapsing consecutive whitespace.
func collapseLines(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	// Collapse multiple spaces into one
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// extractSkillName extracts the bare skill name from a documentation string.
// orchestrator.yaml uses "skill-name for description..." — we need just "skill-name".
// If the string has no space, it's already a bare name.
func extractSkillName(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, ' '); idx > 0 {
		return s[:idx]
	}
	return s
}

// formatRoutingTable converts a specialist→condition map to markdown table rows.
// Keys are sorted alphabetically for deterministic output.
func formatRoutingTable(routing map[string]string) string {
	keys := make([]string, 0, len(routing))
	for k := range routing {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "| %s | %s |\n", k, routing[k])
	}
	return b.String()
}

// formatHandoffTable converts phase→criteria map to markdown table rows.
// Phase ordering uses routing keys (sorted) to maintain consistency with
// the phase_routing table. Phases not in routing are appended alphabetically.
//
// Multi-item criteria use the "<-" separator convention established in
// existing test fixtures (see tenxDevData in archetype_test.go).
func formatHandoffTable(criteria map[string][]string, routing map[string]string) string {
	// Determine phase order: routing keys first (sorted), then remaining phases
	routingKeys := make([]string, 0, len(routing))
	for k := range routing {
		routingKeys = append(routingKeys, k)
	}
	sort.Strings(routingKeys)

	// Collect phases present in criteria but not in routing
	inRouting := make(map[string]bool, len(routingKeys))
	for _, k := range routingKeys {
		inRouting[k] = true
	}
	var extra []string
	for k := range criteria {
		if !inRouting[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)

	// Build ordered phase list: handoff criteria phases that appear in routing order,
	// then any extra phases. Only include phases that have criteria entries.
	var orderedPhases []string
	for _, k := range routingKeys {
		if _, has := criteria[k]; has {
			orderedPhases = append(orderedPhases, k)
		}
	}
	orderedPhases = append(orderedPhases, extra...)

	var b strings.Builder
	for _, phase := range orderedPhases {
		items := criteria[phase]
		if len(items) == 0 {
			continue
		}
		// Format: | phase | - item1<- item2<- item3< |
		fmt.Fprintf(&b, "| %s | - %s", phase, items[0])
		for _, item := range items[1:] {
			fmt.Fprintf(&b, "<- %s", item)
		}
		b.WriteString("< |\n")
	}
	return b.String()
}

// formatAntipatterns converts a list of antipattern strings to a markdown bullet list.
// Each item is wrapped in bold markers: "- **item**".
func formatAntipatterns(patterns []string) string {
	var b strings.Builder
	for _, p := range patterns {
		fmt.Fprintf(&b, "- **%s**\n", p)
	}
	return strings.TrimRight(b.String(), "\n")
}

// enrichArchetypeData loads archetype config files (e.g., orchestrator.yaml) from the
// rite directory and merges the converted data into manifest.ArchetypeData.
//
// Precedence: config file provides the base layer; manifest.archetype_data overrides
// individual fields. This follows the "convention < explicit" pattern used throughout
// the resolution chain.
//
// Today only the "orchestrator" archetype has a config file convention. Future archetypes
// can add their own by extending the switch below.
func enrichArchetypeData(manifest *RiteManifest, riteFS fs.FS) {
	if riteFS == nil {
		return
	}

	// Deduplicate: only load config once per archetype name
	loaded := make(map[string]bool)

	for _, agent := range manifest.Agents {
		if agent.Archetype == "" || loaded[agent.Archetype] {
			continue
		}
		loaded[agent.Archetype] = true

		switch agent.Archetype {
		case "orchestrator":
			config, err := loadOrchestratorConfig(riteFS)
			if err != nil {
				slog.Info("no orchestrator.yaml found for archetype agent",
					"agent", agent.Name, "rite", manifest.Name)
				continue
			}

			configData := convertOrchestratorConfigToData(config)

			if manifest.ArchetypeData == nil {
				manifest.ArchetypeData = make(map[string]map[string]any)
			}

			// Config provides base; existing manifest archetype_data overrides per-field
			existing := manifest.ArchetypeData[agent.Archetype]
			if existing == nil {
				manifest.ArchetypeData[agent.Archetype] = configData
			} else {
				for k, v := range configData {
					if _, has := existing[k]; !has {
						existing[k] = v
					}
				}
			}

		default:
			// Unknown archetype config convention — skip silently.
			// The archetype template renderer will handle unknown archetypes.
		}
	}
}
