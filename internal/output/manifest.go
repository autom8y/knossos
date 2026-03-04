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
	Path    string                 `json:"path"`
	Exists  bool                   `json:"exists"`
	Format  string                 `json:"format,omitempty"`
	Schema  *ManifestSchemaInfo    `json:"schema,omitempty"`
	Content map[string]interface{} `json:"content,omitempty"`
	Error   string                 `json:"error,omitempty"`
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
	b.WriteString(fmt.Sprintf("Manifest: %s\n", m.Path))
	b.WriteString(fmt.Sprintf("Format: %s\n", strings.ToUpper(m.Format)))

	if m.Schema != nil {
		valid := "valid"
		if !m.Schema.Valid {
			valid = "invalid"
		}
		b.WriteString(fmt.Sprintf("Schema: %s v%s (%s)\n", m.Schema.Type, m.Schema.Version, valid))
	}

	b.WriteString("\n")

	// Format content sections
	if m.Content != nil {
		if project, ok := m.Content["project"].(map[string]interface{}); ok {
			if name, ok := project["name"].(string); ok {
				b.WriteString(fmt.Sprintf("Project: %s\n", name))
			}
			if desc, ok := project["description"].(string); ok {
				b.WriteString(fmt.Sprintf("Description: %s\n", desc))
			}
		}

		b.WriteString("\n")

		if teams, ok := m.Content["teams"].(map[string]interface{}); ok {
			b.WriteString("Rites:\n")
			if def, ok := teams["default"].(string); ok {
				b.WriteString(fmt.Sprintf("  Default: %s\n", def))
			}
			if avail, ok := teams["available"].([]interface{}); ok {
				names := make([]string, len(avail))
				for i, v := range avail {
					names[i] = fmt.Sprintf("%v", v)
				}
				b.WriteString(fmt.Sprintf("  Available: %s\n", strings.Join(names, ", ")))
			}
		}

		if paths, ok := m.Content["paths"].(map[string]interface{}); ok {
			b.WriteString("\nPaths:\n")
			for key, val := range paths {
				b.WriteString(fmt.Sprintf("  %s: %v\n", cases.Title(language.English).String(key), val))
			}
		}
	}

	return b.String()
}

// ManifestValidateOutput represents the manifest validate command output.
type ManifestValidateOutput struct {
	Path     string                     `json:"path"`
	Schema   string                     `json:"schema"`
	Valid    bool                       `json:"valid"`
	Issues   []ManifestValidationIssue  `json:"issues"`
	Warnings []ManifestValidationIssue  `json:"warnings"`
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
	b.WriteString(fmt.Sprintf("Validating: %s\n", v.Path))
	b.WriteString(fmt.Sprintf("Schema: %s\n\n", v.Schema))

	for _, issue := range v.Issues {
		b.WriteString(fmt.Sprintf("[ERROR] %s: %s\n", issue.Path, issue.Message))
	}

	for _, warn := range v.Warnings {
		b.WriteString(fmt.Sprintf("[WARN]  %s: %s\n", warn.Path, warn.Message))
	}

	result := "VALID"
	if !v.Valid {
		result = "INVALID"
	}
	b.WriteString(fmt.Sprintf("\nResult: %s (%d errors, %d warnings)\n", result, len(v.Issues), len(v.Warnings)))

	return b.String()
}

// ManifestDiffOutput represents the manifest diff command output.
type ManifestDiffOutput struct {
	Base          string                `json:"base"`
	Compare       string                `json:"compare"`
	HasChanges    bool                  `json:"has_changes"`
	Changes       []ManifestDiffChange  `json:"changes"`
	Additions     int                   `json:"additions"`
	Modifications int                   `json:"modifications"`
	Deletions     int                   `json:"deletions"`
	UnifiedDiff   string                `json:"-"` // For text output
}

// ManifestDiffChange represents a single change.
type ManifestDiffChange struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
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
	b.WriteString(fmt.Sprintf("--- %s\n", d.Base))
	b.WriteString(fmt.Sprintf("+++ %s\n\n", d.Compare))

	for _, c := range d.Changes {
		switch c.Type {
		case "added":
			b.WriteString(fmt.Sprintf("+ %s: %v\n", c.Path, c.NewValue))
		case "removed":
			b.WriteString(fmt.Sprintf("- %s: %v\n", c.Path, c.OldValue))
		case "modified":
			b.WriteString(fmt.Sprintf("~ %s: %v -> %v\n", c.Path, c.OldValue, c.NewValue))
		}
	}

	b.WriteString(fmt.Sprintf("\n%d additions, %d modifications, %d deletions\n",
		d.Additions, d.Modifications, d.Deletions))

	return b.String()
}

// ManifestMergeOutput represents the manifest merge command output.
type ManifestMergeOutput struct {
	Base          string                   `json:"base"`
	Ours          string                   `json:"ours"`
	Theirs        string                   `json:"theirs"`
	Strategy      string                   `json:"strategy"`
	HasConflicts  bool                     `json:"has_conflicts"`
	Conflicts     []ManifestMergeConflict  `json:"conflicts,omitempty"`
	Merged        map[string]interface{}   `json:"merged,omitempty"`
	MergedMarkers string                   `json:"merged_with_markers,omitempty"`
	Changes       *ManifestMergeChanges    `json:"changes,omitempty"`
	OutputPath    string                   `json:"output_path,omitempty"`
}

// ManifestMergeConflict represents a merge conflict.
type ManifestMergeConflict struct {
	Path        string      `json:"path"`
	BaseValue   interface{} `json:"base_value"`
	OursValue   interface{} `json:"ours_value"`
	TheirsValue interface{} `json:"theirs_value"`
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
	b.WriteString(fmt.Sprintf("  Base: %s\n", m.Base))
	b.WriteString(fmt.Sprintf("  Ours: %s\n", m.Ours))
	b.WriteString(fmt.Sprintf("  Theirs: %s\n", m.Theirs))
	b.WriteString(fmt.Sprintf("  Strategy: %s\n\n", m.Strategy))

	if m.Changes != nil {
		if len(m.Changes.FromOurs) > 0 || len(m.Changes.FromTheirs) > 0 {
			b.WriteString("Changes:\n")
			for _, path := range m.Changes.FromOurs {
				b.WriteString(fmt.Sprintf("  [OURS]   %s\n", path))
			}
			for _, path := range m.Changes.FromTheirs {
				b.WriteString(fmt.Sprintf("  [THEIRS] %s\n", path))
			}
			b.WriteString("\n")
		}
	}

	if m.HasConflicts {
		b.WriteString("Conflicts:\n")
		for _, c := range m.Conflicts {
			b.WriteString(fmt.Sprintf("  %s:\n", c.Path))
			b.WriteString(fmt.Sprintf("    Base:   %v\n", c.BaseValue))
			b.WriteString(fmt.Sprintf("    Ours:   %v\n", c.OursValue))
			b.WriteString(fmt.Sprintf("    Theirs: %v\n", c.TheirsValue))
		}
		b.WriteString(fmt.Sprintf("\nResult: CONFLICTS (%d conflicts)\n", len(m.Conflicts)))
	} else {
		b.WriteString("Result: MERGED (no conflicts)\n")
	}

	if m.OutputPath != "" {
		b.WriteString(fmt.Sprintf("Output: %s\n", m.OutputPath))
	}

	return b.String()
}
