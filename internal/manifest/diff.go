// Package manifest - diff computation between manifests
package manifest

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
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
	Path     string      `json:"path"`
	Type     ChangeType  `json:"type"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
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
func walkAndCompare(path string, base, compare interface{}, changes *[]Change, opts ManifestDiffOptions) {
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
	case map[string]interface{}:
		compareVal := compare.(map[string]interface{})
		compareMaps(path, baseVal, compareVal, changes, opts)

	case []interface{}:
		compareVal := compare.([]interface{})
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
func compareMaps(path string, base, compare map[string]interface{}, changes *[]Change, opts ManifestDiffOptions) {
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
func compareArrays(path string, base, compare []interface{}, changes *[]Change, opts ManifestDiffOptions) {
	maxLen := len(base)
	if len(compare) > maxLen {
		maxLen = len(compare)
	}

	for i := 0; i < maxLen; i++ {
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
func compareArraysAsSet(path string, base, compare []interface{}, changes *[]Change) {
	baseSet := make(map[string]interface{})
	compareSet := make(map[string]interface{})

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
func toJSONKey(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// FormatUnified formats the diff result as unified diff output.
func (d *DiffResult) FormatUnified() string {
	if !d.HasChanges {
		return ""
	}

	var result string
	result += "--- " + d.Base + "\n"
	result += "+++ " + d.Compare + "\n"
	result += "\n"

	// Group changes by top-level path
	groups := groupChangesBySection(d.Changes)
	for section, sectionChanges := range groups {
		result += "@@ " + section + " @@\n"
		for _, c := range sectionChanges {
			switch c.Type {
			case ChangeRemoved:
				result += formatValue("-", c.OldValue)
			case ChangeAdded:
				result += formatValue("+", c.NewValue)
			case ChangeModified:
				result += formatValue("-", c.OldValue)
				result += formatValue("+", c.NewValue)
			}
		}
		result += "\n"
	}

	return result
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
func formatValue(prefix string, val interface{}) string {
	data, _ := json.MarshalIndent(val, "", "  ")
	lines := splitLines(string(data))
	var result string
	for _, line := range lines {
		result += prefix + "  " + line + "\n"
	}
	return result
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
