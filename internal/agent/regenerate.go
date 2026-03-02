package agent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/autom8y/knossos/internal/errors"
)

// RegeneratePlatformSections regenerates platform and derived sections from templates.
// Author-owned sections are preserved exactly as-is.
func RegeneratePlatformSections(agent *ParsedAgent, archetype *Archetype) (*ParsedAgent, error) {
	if agent == nil {
		return nil, errors.New(errors.CodeUsageError, "agent is required")
	}
	if archetype == nil {
		return nil, errors.New(errors.CodeUsageError, "archetype is required")
	}

	// Load the archetype template content
	templateContent, err := loadArchetypeTemplate(archetype.Name)
	if err != nil {
		return nil, err
	}

	// Parse the template to extract platform section content
	platformSections, err := extractPlatformSectionsFromTemplate(templateContent)
	if err != nil {
		return nil, err
	}

	// Build a new section list following archetype order
	newSections := []ParsedSection{}
	processedNames := make(map[string]bool)

	// Process sections in archetype order
	for _, sectionDef := range archetype.Sections {
		heading := strings.TrimPrefix(sectionDef.Heading, "## ")

		// Find existing section
		existingSection := agent.FindSection(sectionDef.Name)

		var newSection ParsedSection

		switch sectionDef.Ownership {
		case OwnerPlatform:
			// Replace with template content
			content, found := platformSections[sectionDef.Name]
			if !found {
				content = fmt.Sprintf("<!-- Platform section %q content not found in template -->", sectionDef.Name)
			}
			newSection = ParsedSection{
				Name:      sectionDef.Name,
				Heading:   heading,
				Content:   strings.TrimSpace(content),
				Ownership: OwnerPlatform,
			}

		case OwnerDerived:
			// Generate from frontmatter
			content := generateDerivedContent(agent, sectionDef.Name)
			newSection = ParsedSection{
				Name:      sectionDef.Name,
				Heading:   heading,
				Content:   strings.TrimSpace(content),
				Ownership: OwnerDerived,
			}

		case OwnerAuthor:
			// Preserve existing content or add TODO marker
			if existingSection != nil {
				newSection = *existingSection
				newSection.Name = sectionDef.Name
				newSection.Heading = heading
			} else {
				// Add with TODO marker
				todoContent := fmt.Sprintf("<!-- TODO: %s -->", sectionDef.TodoHint)
				newSection = ParsedSection{
					Name:      sectionDef.Name,
					Heading:   heading,
					Content:   todoContent,
					Ownership: OwnerAuthor,
				}
			}
		}

		newSections = append(newSections, newSection)
		processedNames[sectionDef.Name] = true
	}

	// Preserve any unknown sections (not in archetype) at the end
	for _, section := range agent.Sections {
		if section.Name == "" || !processedNames[section.Name] {
			newSections = append(newSections, section)
		}
	}

	// Return updated agent
	updatedAgent := &ParsedAgent{
		Frontmatter:    agent.Frontmatter,
		RawFrontmatter: agent.RawFrontmatter,
		Title:          agent.Title,
		Preamble:       agent.Preamble,
		Sections:       newSections,
	}

	return updatedAgent, nil
}

// loadArchetypeTemplate loads the template content for an archetype.
func loadArchetypeTemplate(archetypeName string) (string, error) {
	templatePath := fmt.Sprintf("templates/%s.md.tpl", archetypeName)
	content, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return "", errors.Wrap(errors.CodeFileNotFound,
			fmt.Sprintf("template not found for archetype %q", archetypeName), err)
	}
	return string(content), nil
}

// extractPlatformSectionsFromTemplate parses a template and extracts platform section content.
func extractPlatformSectionsFromTemplate(templateContent string) (map[string]string, error) {
	// Render the template with minimal data to extract structure
	data := ScaffoldData{
		Name:        "placeholder",
		Title:       "Placeholder",
		Description: "Placeholder description",
		RiteName:    "placeholder-rite",
		Tools:       []string{"Read"},
		Model:       "opus",
		Color:       "blue",
	}

	funcMap := sprig.TxtFuncMap()
	tmpl, err := template.New("extract").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse template", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to render template", err)
	}

	// Parse the rendered template as an agent file
	renderedContent := buf.Bytes()
	parsedTemplate, err := ParseAgentSections(renderedContent)
	if err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse rendered template", err)
	}

	// Extract platform and derived sections
	sections := make(map[string]string)
	for _, section := range parsedTemplate.Sections {
		if section.Ownership == OwnerPlatform {
			sections[section.Name] = section.Content
		}
	}

	return sections, nil
}

// generateDerivedContent generates content for derived sections from frontmatter.
func generateDerivedContent(agent *ParsedAgent, sectionName string) string {
	switch sectionName {
	case "tool-access":
		return generateToolAccessTable(agent)
	case "position-in-workflow":
		return generateWorkflowPosition(agent)
	case "skills-reference":
		return generateSkillsReference(agent.Frontmatter.Skills)
	default:
		return fmt.Sprintf("<!-- Derived section %q not implemented -->", sectionName)
	}
}

// generateToolAccessTable generates the tool access table from frontmatter tools.
func generateToolAccessTable(agent *ParsedAgent) string {
	if len(agent.Frontmatter.Tools) == 0 {
		return "No tools configured."
	}

	var buf bytes.Buffer
	buf.WriteString("You have: ")
	for i, tool := range agent.Frontmatter.Tools {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("`")
		buf.WriteString(tool)
		buf.WriteString("`")
	}
	buf.WriteString("\n\n")

	buf.WriteString("| Tool | When to Use |\n")
	buf.WriteString("|------|-------------|\n")
	for _, tool := range agent.Frontmatter.Tools {
		buf.WriteString(fmt.Sprintf("| **%s** | *Use for %s operations* |\n", tool, strings.ToLower(tool)))
	}

	return buf.String()
}

// generateWorkflowPosition generates the workflow position section from upstream/downstream.
func generateWorkflowPosition(agent *ParsedAgent) string {
	var buf bytes.Buffer

	if len(agent.Frontmatter.Upstream) > 0 {
		buf.WriteString("**Upstream**: ")
		var names []string
		for _, u := range agent.Frontmatter.Upstream {
			names = append(names, u.Source)
		}
		buf.WriteString(strings.Join(names, ", "))
		buf.WriteString("\n")
	} else {
		buf.WriteString("**Upstream**: Not specified\n")
	}

	if len(agent.Frontmatter.Downstream) > 0 {
		buf.WriteString("**Downstream**: ")
		var names []string
		for _, d := range agent.Frontmatter.Downstream {
			names = append(names, d.Agent)
		}
		buf.WriteString(strings.Join(names, ", "))
		buf.WriteString("\n")
	} else {
		buf.WriteString("**Downstream**: Not specified\n")
	}

	return buf.String()
}

// generateSkillsReference generates the skills reference section from the agent's frontmatter
// skills list. If the agent has skills defined, each skill is listed by its actual name without
// the @-prefix anti-pattern. If no skills are defined, a generic instruction is returned.
func generateSkillsReference(skills []string) string {
	if len(skills) == 0 {
		return "Load skills on demand via Skill tool as needed."
	}

	var buf strings.Builder
	buf.WriteString("Reference these skills as appropriate:\n")
	for _, skill := range skills {
		buf.WriteString("- ")
		buf.WriteString(skill)
		buf.WriteString("\n")
	}
	return strings.TrimRight(buf.String(), "\n")
}

// AssembleAgentFile reassembles a parsed agent into markdown content.
func AssembleAgentFile(agent *ParsedAgent) []byte {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")
	buf.WriteString(agent.RawFrontmatter)
	if !strings.HasSuffix(agent.RawFrontmatter, "\n") {
		buf.WriteString("\n")
	}
	buf.WriteString("---\n\n")

	// Write title
	buf.WriteString("# ")
	buf.WriteString(agent.Title)
	buf.WriteString("\n\n")

	// Write preamble
	if agent.Preamble != "" {
		buf.WriteString(agent.Preamble)
		buf.WriteString("\n\n")
	}

	// Write sections
	for i, section := range agent.Sections {
		buf.WriteString("## ")
		buf.WriteString(section.Heading)
		buf.WriteString("\n\n")
		buf.WriteString(section.Content)
		if !strings.HasSuffix(section.Content, "\n") {
			buf.WriteString("\n")
		}
		if i < len(agent.Sections)-1 {
			buf.WriteString("\n")
		}
	}

	// Ensure file ends with single newline
	content := buf.Bytes()
	content = bytes.TrimRight(content, "\n")
	content = append(content, '\n')

	return content
}
