// Package explain implements the ari explain command for concept lookup.
package explain

import (
	"fmt"
	"strings"
)

// ConceptEntry holds a fully parsed concept definition.
type ConceptEntry struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	SeeAlso     []string `json:"see_also"`
	Aliases     []string `json:"aliases"`
	HarnessTerm string   `json:"cc_term,omitempty"`
}

// ConceptOutput represents a single concept lookup result.
// Implements output.Textable for human-readable output.
// Serializes to JSON/YAML via struct tags for machine-readable output.
type ConceptOutput struct {
	Concept        string   `json:"concept"`
	DisplayName    string   `json:"display_name"`
	Summary        string   `json:"summary"`
	Description    string   `json:"description"`
	SeeAlso        []string `json:"see_also"`
	ProjectContext string   `json:"project_context,omitempty"`
}

// Text implements output.Textable.
func (c ConceptOutput) Text() string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== %s ===\n", c.DisplayName)
	b.WriteString("\n")
	b.WriteString(c.Description)
	b.WriteString("\n")

	if len(c.SeeAlso) > 0 {
		b.WriteString("\n")
		fmt.Fprintf(&b, "See also: %s\n", strings.Join(c.SeeAlso, ", "))
	}

	if c.ProjectContext != "" {
		b.WriteString("\n")
		b.WriteString(c.ProjectContext)
		b.WriteString("\n")
	}

	return b.String()
}

// ConceptListOutput represents all concepts for table listing.
// Implements output.Tabular for text table output.
// Serializes to JSON/YAML via struct tags.
type ConceptListOutput struct {
	Concepts []ConceptSummary `json:"concepts"`
}

// ConceptSummary is a concept name+summary pair for listing.
type ConceptSummary struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// Headers implements output.Tabular.
func (c ConceptListOutput) Headers() []string {
	return []string{"CONCEPT", "SUMMARY"}
}

// Rows implements output.Tabular.
func (c ConceptListOutput) Rows() [][]string {
	rows := make([][]string, len(c.Concepts))
	for i, cs := range c.Concepts {
		rows[i] = []string{cs.Name, cs.Summary}
	}
	return rows
}
