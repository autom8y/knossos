// Package output provides format-aware output printing for Ariadne.
// This file contains manifest-domain specific output structures.
package output

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// --- Manifest Output Structures ---

// ManifestShowOutput represents the manifest show command output.
type ManifestShowOutput struct {
	Path    string              `json:"path"`
	Exists  bool                `json:"exists"`
	Format  string              `json:"format,omitempty"`
	Schema  *ManifestSchemaInfo `json:"schema,omitempty"`
	Content map[string]any      `json:"content,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// ManifestSchemaInfo holds schema metadata.
type ManifestSchemaInfo struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Valid   bool   `json:"valid"`
}

// Text implements Textable for ManifestShowOutput.
func (m ManifestShowOutput) Text() string {
	if !m.Exists {
		return fmt.Sprintf("Manifest: %s\nStatus: Not found", m.Path)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Manifest: %s\n", m.Path)
	fmt.Fprintf(&b, "Format: %s\n", strings.ToUpper(m.Format))

	if m.Schema != nil {
		valid := "valid"
		if !m.Schema.Valid {
			valid = "invalid"
		}
		fmt.Fprintf(&b, "Schema: %s v%s (%s)\n", m.Schema.Type, m.Schema.Version, valid)
	}

	b.WriteString("\n")

	// Format content sections
	if m.Content != nil {
		if project, ok := m.Content["project"].(map[string]any); ok {
			if name, ok := project["name"].(string); ok {
				fmt.Fprintf(&b, "Project: %s\n", name)
			}
			if desc, ok := project["description"].(string); ok {
				fmt.Fprintf(&b, "Description: %s\n", desc)
			}
		}

		b.WriteString("\n")

		if teams, ok := m.Content["teams"].(map[string]any); ok {
			b.WriteString("Rites:\n")
			if def, ok := teams["default"].(string); ok {
				fmt.Fprintf(&b, "  Default: %s\n", def)
			}
			if avail, ok := teams["available"].([]any); ok {
				names := make([]string, len(avail))
				for i, v := range avail {
					names[i] = fmt.Sprintf("%v", v)
				}
				fmt.Fprintf(&b, "  Available: %s\n", strings.Join(names, ", "))
			}
		}

		if paths, ok := m.Content["paths"].(map[string]any); ok {
			b.WriteString("\nPaths:\n")
			for key, val := range paths {
				fmt.Fprintf(&b, "  %s: %v\n", cases.Title(language.English).String(key), val)
			}
		}
	}

	return b.String()
}

// ManifestValidateOutput represents the manifest validate command output.
type ManifestValidateOutput struct {
	Path     string                    `json:"path"`
	Schema   string                    `json:"schema"`
	Valid    bool                      `json:"valid"`
	Issues   []ManifestValidationIssue `json:"issues"`
	Warnings []ManifestValidationIssue `json:"warnings"`
}

// ManifestValidationIssue represents a validation issue.
type ManifestValidationIssue struct {
	Path     string `json:"path"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// Text implements Textable for ManifestValidateOutput.
func (v ManifestValidateOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Validating: %s\n", v.Path)
	fmt.Fprintf(&b, "Schema: %s\n\n", v.Schema)

	for _, issue := range v.Issues {
		fmt.Fprintf(&b, "[ERROR] %s: %s\n", issue.Path, issue.Message)
	}

	for _, warn := range v.Warnings {
		fmt.Fprintf(&b, "[WARN]  %s: %s\n", warn.Path, warn.Message)
	}

	result := "VALID"
	if !v.Valid {
		result = "INVALID"
	}
	fmt.Fprintf(&b, "\nResult: %s (%d errors, %d warnings)\n", result, len(v.Issues), len(v.Warnings))

	return b.String()
}

// ManifestDiffOutput represents the manifest diff command output.
type ManifestDiffOutput struct {
	Base          string               `json:"base"`
	Compare       string               `json:"compare"`
	HasChanges    bool                 `json:"has_changes"`
	Changes       []ManifestDiffChange `json:"changes"`
	Additions     int                  `json:"additions"`
	Modifications int                  `json:"modifications"`
	Deletions     int                  `json:"deletions"`
	UnifiedDiff   string               `json:"-"` // For text output
}

// ManifestDiffChange represents a single change.
type ManifestDiffChange struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	OldValue any    `json:"old_value,omitempty"`
	NewValue any    `json:"new_value,omitempty"`
}

// Text implements Textable for ManifestDiffOutput.
func (d ManifestDiffOutput) Text() string {
	if !d.HasChanges {
		return "No differences found"
	}
	if d.UnifiedDiff != "" {
		return d.UnifiedDiff
	}

	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n", d.Base)
	fmt.Fprintf(&b, "+++ %s\n\n", d.Compare)

	for _, c := range d.Changes {
		switch c.Type {
		case "added":
			fmt.Fprintf(&b, "+ %s: %v\n", c.Path, c.NewValue)
		case "removed":
			fmt.Fprintf(&b, "- %s: %v\n", c.Path, c.OldValue)
		case "modified":
			fmt.Fprintf(&b, "~ %s: %v -> %v\n", c.Path, c.OldValue, c.NewValue)
		}
	}

	fmt.Fprintf(&b, "\n%d additions, %d modifications, %d deletions\n",
		d.Additions, d.Modifications, d.Deletions)

	return b.String()
}

// ManifestMergeOutput represents the manifest merge command output.
type ManifestMergeOutput struct {
	Base          string                  `json:"base"`
	Ours          string                  `json:"ours"`
	Theirs        string                  `json:"theirs"`
	Strategy      string                  `json:"strategy"`
	HasConflicts  bool                    `json:"has_conflicts"`
	Conflicts     []ManifestMergeConflict `json:"conflicts,omitempty"`
	Merged        map[string]any          `json:"merged,omitempty"`
	MergedMarkers string                  `json:"merged_with_markers,omitempty"`
	Changes       *ManifestMergeChanges   `json:"changes,omitempty"`
	OutputPath    string                  `json:"output_path,omitempty"`
}

// ManifestMergeConflict represents a merge conflict.
type ManifestMergeConflict struct {
	Path        string `json:"path"`
	BaseValue   any    `json:"base_value"`
	OursValue   any    `json:"ours_value"`
	TheirsValue any    `json:"theirs_value"`
}

// ManifestMergeChanges tracks change sources.
type ManifestMergeChanges struct {
	FromOurs   []string `json:"from_ours"`
	FromTheirs []string `json:"from_theirs"`
}

// Text implements Textable for ManifestMergeOutput.
func (m ManifestMergeOutput) Text() string {
	var b strings.Builder
	b.WriteString("Merging manifests...\n")
	fmt.Fprintf(&b, "  Base: %s\n", m.Base)
	fmt.Fprintf(&b, "  Ours: %s\n", m.Ours)
	fmt.Fprintf(&b, "  Theirs: %s\n", m.Theirs)
	fmt.Fprintf(&b, "  Strategy: %s\n\n", m.Strategy)

	if m.Changes != nil {
		if len(m.Changes.FromOurs) > 0 || len(m.Changes.FromTheirs) > 0 {
			b.WriteString("Changes:\n")
			for _, path := range m.Changes.FromOurs {
				fmt.Fprintf(&b, "  [OURS]   %s\n", path)
			}
			for _, path := range m.Changes.FromTheirs {
				fmt.Fprintf(&b, "  [THEIRS] %s\n", path)
			}
			b.WriteString("\n")
		}
	}

	if m.HasConflicts {
		b.WriteString("Conflicts:\n")
		for _, c := range m.Conflicts {
			fmt.Fprintf(&b, "  %s:\n", c.Path)
			fmt.Fprintf(&b, "    Base:   %v\n", c.BaseValue)
			fmt.Fprintf(&b, "    Ours:   %v\n", c.OursValue)
			fmt.Fprintf(&b, "    Theirs: %v\n", c.TheirsValue)
		}
		fmt.Fprintf(&b, "\nResult: CONFLICTS (%d conflicts)\n", len(m.Conflicts))
	} else {
		b.WriteString("Result: MERGED (no conflicts)\n")
	}

	if m.OutputPath != "" {
		fmt.Fprintf(&b, "Output: %s\n", m.OutputPath)
	}

	return b.String()
}
