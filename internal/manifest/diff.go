// Package manifest - diff computation between manifests
package manifest

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ChangeType represents the type of change.
type ChangeType string

const (
	// ChangeAdded indicates a value was added.
	ChangeAdded ChangeType = "added"
	// ChangeModified indicates a value was modified.
	ChangeModified ChangeType = "modified"
	// ChangeRemoved indicates a value was removed.
	ChangeRemoved ChangeType = "removed"
)

// Change represents a single change between manifests.
type Change struct {
	Path     string     `json:"path"`
	Type     ChangeType `json:"type"`
	OldValue any        `json:"old_value,omitempty"`
	NewValue any        `json:"new_value,omitempty"`
}

// DiffResult holds the result of comparing two manifests.
type DiffResult struct {
	Base          string   `json:"base"`
	Compare       string   `json:"compare"`
	HasChanges    bool     `json:"has_changes"`
	Changes       []Change `json:"changes"`
	Additions     int      `json:"additions"`
	Modifications int      `json:"modifications"`
	Deletions     int      `json:"deletions"`
}

// ManifestDiffOptions configures manifest diff behavior.
type ManifestDiffOptions struct {
	IgnoreOrder bool // Treat arrays as sets for comparison
}

// Diff computes differences between two manifests.
func Diff(base, compare *Manifest, opts ManifestDiffOptions) (*DiffResult, error) {
	result := &DiffResult{
		Base:    base.Path,
		Compare: compare.Path,
		Changes: []Change{},
	}

	// Walk and compare the structures
	walkAndCompare("$", base.Content, compare.Content, &result.Changes, opts)

	// Count changes by type
	for _, c := range result.Changes {
		switch c.Type {
		case ChangeAdded:
			result.Additions++
		case ChangeModified:
			result.Modifications++
		case ChangeRemoved:
			result.Deletions++
		}
	}

	result.HasChanges = len(result.Changes) > 0

	return result, nil
}

// walkAndCompare recursively compares two values and collects changes.
func walkAndCompare(path string, base, compare any, changes *[]Change, opts ManifestDiffOptions) {
	// Handle nil cases
	if base == nil && compare == nil {
		return
	}
	if base == nil {
		*changes = append(*changes, Change{
			Path:     path,
			Type:     ChangeAdded,
			NewValue: compare,
		})
		return
	}
	if compare == nil {
		*changes = append(*changes, Change{
			Path:     path,
			Type:     ChangeRemoved,
			OldValue: base,
		})
		return
	}

	// Type mismatch
	baseType := reflect.TypeOf(base)
	compareType := reflect.TypeOf(compare)
	if baseType != compareType {
		*changes = append(*changes, Change{
			Path:     path,
			Type:     ChangeModified,
			OldValue: base,
			NewValue: compare,
		})
		return
	}

	switch baseVal := base.(type) {
	case map[string]any:
		compareVal := compare.(map[string]any)
		compareMaps(path, baseVal, compareVal, changes, opts)

	case []any:
		compareVal := compare.([]any)
		if opts.IgnoreOrder {
			compareArraysAsSet(path, baseVal, compareVal, changes)
		} else {
			compareArrays(path, baseVal, compareVal, changes, opts)
		}

	default:
		// Scalar comparison
		if !reflect.DeepEqual(base, compare) {
			*changes = append(*changes, Change{
				Path:     path,
				Type:     ChangeModified,
				OldValue: base,
				NewValue: compare,
			})
		}
	}
}

// compareMaps compares two map values.
func compareMaps(path string, base, compare map[string]any, changes *[]Change, opts ManifestDiffOptions) {
	// Get all keys from both maps
	allKeys := make(map[string]bool)
	for k := range base {
		allKeys[k] = true
	}
	for k := range compare {
		allKeys[k] = true
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		keyPath := path + "." + key
		baseVal, baseOk := base[key]
		compareVal, compareOk := compare[key]

		if !baseOk {
			*changes = append(*changes, Change{
				Path:     keyPath,
				Type:     ChangeAdded,
				NewValue: compareVal,
			})
		} else if !compareOk {
			*changes = append(*changes, Change{
				Path:     keyPath,
				Type:     ChangeRemoved,
				OldValue: baseVal,
			})
		} else {
			walkAndCompare(keyPath, baseVal, compareVal, changes, opts)
		}
	}
}

// compareArrays compares two arrays maintaining order.
func compareArrays(path string, base, compare []any, changes *[]Change, opts ManifestDiffOptions) {
	maxLen := max(len(compare), len(base))

	for i := range maxLen {
		indexPath := path + "[" + strconv.Itoa(i) + "]"

		if i >= len(base) {
			*changes = append(*changes, Change{
				Path:     indexPath,
				Type:     ChangeAdded,
				NewValue: compare[i],
			})
		} else if i >= len(compare) {
			*changes = append(*changes, Change{
				Path:     indexPath,
				Type:     ChangeRemoved,
				OldValue: base[i],
			})
		} else {
			walkAndCompare(indexPath, base[i], compare[i], changes, opts)
		}
	}
}

// compareArraysAsSet compares arrays treating them as sets (ignoring order).
func compareArraysAsSet(path string, base, compare []any, changes *[]Change) {
	baseSet := make(map[string]any)
	compareSet := make(map[string]any)

	// Convert to sets using JSON as key
	for _, v := range base {
		key := toJSONKey(v)
		baseSet[key] = v
	}
	for _, v := range compare {
		key := toJSONKey(v)
		compareSet[key] = v
	}

	// Find additions
	for key, val := range compareSet {
		if _, ok := baseSet[key]; !ok {
			*changes = append(*changes, Change{
				Path:     path + "[]",
				Type:     ChangeAdded,
				NewValue: val,
			})
		}
	}

	// Find deletions
	for key, val := range baseSet {
		if _, ok := compareSet[key]; !ok {
			*changes = append(*changes, Change{
				Path:     path + "[]",
				Type:     ChangeRemoved,
				OldValue: val,
			})
		}
	}
}

// toJSONKey converts a value to a JSON string for use as a map key.
func toJSONKey(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// FormatUnified formats the diff result as unified diff output.
func (d *DiffResult) FormatUnified() string {
	if !d.HasChanges {
		return ""
	}

	var result strings.Builder
	result.WriteString("--- " + d.Base + "\n")
	result.WriteString("+++ " + d.Compare + "\n")
	result.WriteString("\n")

	// Group changes by top-level path
	groups := groupChangesBySection(d.Changes)
	for section, sectionChanges := range groups {
		result.WriteString("@@ " + section + " @@\n")
		for _, c := range sectionChanges {
			switch c.Type {
			case ChangeRemoved:
				result.WriteString(formatValue("-", c.OldValue))
			case ChangeAdded:
				result.WriteString(formatValue("+", c.NewValue))
			case ChangeModified:
				result.WriteString(formatValue("-", c.OldValue))
				result.WriteString(formatValue("+", c.NewValue))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// groupChangesBySection groups changes by their top-level key.
func groupChangesBySection(changes []Change) map[string][]Change {
	groups := make(map[string][]Change)
	for _, c := range changes {
		// Extract section from path ($.section.field -> section)
		section := extractSection(c.Path)
		groups[section] = append(groups[section], c)
	}
	return groups
}

// extractSection extracts the top-level section from a path.
func extractSection(path string) string {
	// Remove leading "$."
	if len(path) > 2 && path[:2] == "$." {
		path = path[2:]
	}
	// Get first segment
	for i, r := range path {
		if r == '.' || r == '[' {
			return path[:i]
		}
	}
	return path
}

// formatValue formats a value for unified diff output.
func formatValue(prefix string, val any) string {
	data, _ := json.MarshalIndent(val, "", "  ")
	lines := splitLines(string(data))
	var result strings.Builder
	for _, line := range lines {
		result.WriteString(prefix + "  " + line + "\n")
	}
	return result.String()
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
