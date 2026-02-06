package agent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/autom8y/knossos/internal/errors"
)

// ScaffoldData contains the template rendering context for agent scaffolding.
type ScaffoldData struct {
	// Name is the kebab-case agent name (e.g., "technology-scout").
	Name string

	// Title is the title-case agent name for the markdown heading (e.g., "Technology Scout").
	Title string

	// Description is the agent description.
	Description string

	// RiteName is the rite this agent belongs to.
	RiteName string

	// Tools is the list of tool names.
	Tools []string

	// Model is the Claude model name.
	Model string

	// Color is the agent badge color.
	Color string
}

// ScaffoldAgent renders an archetype template with the provided agent metadata.
// It returns the complete agent markdown file content as bytes.
func ScaffoldAgent(archetype *Archetype, name, riteName, description string) ([]byte, error) {
	if archetype == nil {
		return nil, errors.New(errors.CodeUsageError, "archetype is required")
	}
	if name == "" {
		return nil, errors.New(errors.CodeUsageError, "agent name is required")
	}
	if riteName == "" {
		return nil, errors.New(errors.CodeUsageError, "rite name is required")
	}

	// Use a default description if none provided
	if description == "" {
		description = fmt.Sprintf("%s agent for the %s rite", toTitleCase(archetype.Name), riteName)
	}

	data := ScaffoldData{
		Name:        name,
		Title:       toTitleCase(name),
		Description: description,
		RiteName:    riteName,
		Tools:       archetype.Defaults.Tools,
		Model:       archetype.Defaults.Model,
		Color:       archetype.Defaults.Color,
	}

	// Load the template from embedded filesystem
	templatePath := fmt.Sprintf("templates/%s.md.tpl", archetype.Name)
	tmplContent, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound,
			fmt.Sprintf("template not found for archetype %q", archetype.Name), err)
	}

	// Parse and execute the template with Sprig functions
	funcMap := sprig.TxtFuncMap()
	tmpl, err := template.New(archetype.Name).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return nil, errors.Wrap(errors.CodeParseError,
			fmt.Sprintf("failed to parse template for archetype %q", archetype.Name), err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError,
			fmt.Sprintf("failed to render template for archetype %q", archetype.Name), err)
	}

	result := buf.Bytes()

	// Validate the generated output passes WARN-mode validation
	if err := validateScaffoldOutput(result); err != nil {
		return nil, errors.Wrap(errors.CodeValidationFailed,
			"generated agent file failed validation", err)
	}

	return result, nil
}

// validateScaffoldOutput parses and validates the generated agent markdown.
// This ensures scaffold output always produces valid agents.
func validateScaffoldOutput(content []byte) error {
	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		return fmt.Errorf("frontmatter parse error: %w", err)
	}

	if err := fm.Validate(); err != nil {
		return fmt.Errorf("frontmatter validation error: %w", err)
	}

	return nil
}

// toTitleCase converts a kebab-case string to Title Case.
// "technology-scout" becomes "Technology Scout".
func toTitleCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
