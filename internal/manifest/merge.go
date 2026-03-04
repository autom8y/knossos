// Package manifest - three-way merge logic
package manifest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// MergeStrategy defines how conflicts are resolved.
type MergeStrategy string

const (
	// StrategySmart performs field-level three-way merge.
	StrategySmart MergeStrategy = "smart"
	// StrategyOurs prefers our changes on conflict.
	StrategyOurs MergeStrategy = "ours"
	// StrategyTheirs prefers their changes on conflict.
	StrategyTheirs MergeStrategy = "theirs"
	// StrategyUnion merges arrays with union (no duplicates).
	StrategyUnion MergeStrategy = "union"
)

// ManifestMergeOptions configures manifest merge behavior.
type ManifestMergeOptions struct {
	Strategy MergeStrategy
	DryRun   bool
}

// Conflict represents a merge conflict.
type Conflict struct {
	Path        string      `json:"path"`
	BaseValue   interface{} `json:"base_value"`
	OursValue   interface{} `json:"ours_value"`
	TheirsValue interface{} `json:"theirs_value"`
}

// MergeChanges tracks which changes came from which source.
type MergeChanges struct {
	FromOurs   []string `json:"from_ours"`
	FromTheirs []string `json:"from_theirs"`
}

// MergeResult holds the result of a three-way merge.
type MergeResult struct {
	Base          string                 `json:"base"`
	Ours          string                 `json:"ours"`
	Theirs        string                 `json:"theirs"`
	Strategy      string                 `json:"strategy"`
	HasConflicts  bool                   `json:"has_conflicts"`
	Conflicts     []Conflict             `json:"conflicts,omitempty"`
	Merged        map[string]interface{} `json:"merged,omitempty"`
	MergedMarkers string                 `json:"merged_with_markers,omitempty"`
	Changes       *MergeChanges          `json:"changes,omitempty"`
}

// Merge performs a three-way merge of manifests.
func Merge(base, ours, theirs *Manifest, opts ManifestMergeOptions) (*MergeResult, error) {
	result := &MergeResult{
		Base:     base.Path,
		Ours:     ours.Path,
		Theirs:   theirs.Path,
		Strategy: string(opts.Strategy),
		Changes: &MergeChanges{
			FromOurs:   []string{},
			FromTheirs: []string{},
		},
	}

	switch opts.Strategy {
	case StrategyOurs:
		return mergePreferOurs(result, ours)
	case StrategyTheirs:
		return mergePreferTheirs(result, theirs)
	case StrategyUnion:
		return mergeUnion(result, base, ours, theirs)
	default:
		return mergeSmart(result, base, ours, theirs)
	}
}

// mergePreferOurs returns ours content.
func mergePreferOurs(result *MergeResult, ours *Manifest) (*MergeResult, error) {
	result.Merged = deepCopy(ours.Content)
	result.HasConflicts = false
	return result, nil
}

// mergePreferTheirs returns theirs content.
func mergePreferTheirs(result *MergeResult, theirs *Manifest) (*MergeResult, error) {
	result.Merged = deepCopy(theirs.Content)
	result.HasConflicts = false
	return result, nil
}

// mergeUnion merges arrays as sets.
func mergeUnion(result *MergeResult, base, ours, theirs *Manifest) (*MergeResult, error) {
	merged := make(map[string]interface{})
	conflicts := []Conflict{}

	mergeFieldsUnion("$", base.Content, ours.Content, theirs.Content, merged, &conflicts, result.Changes)

	result.Merged = merged
	result.HasConflicts = len(conflicts) > 0
	result.Conflicts = conflicts

	if result.HasConflicts {
		result.MergedMarkers = generateConflictMarkers(merged, conflicts)
	}

	return result, nil
}

// mergeSmart performs field-level three-way merge.
func mergeSmart(result *MergeResult, base, ours, theirs *Manifest) (*MergeResult, error) {
	merged := make(map[string]interface{})
	conflicts := []Conflict{}

	mergeFields("$", base.Content, ours.Content, theirs.Content, merged, &conflicts, result.Changes)

	result.Merged = merged
	result.HasConflicts = len(conflicts) > 0
	result.Conflicts = conflicts

	if result.HasConflicts {
		result.MergedMarkers = generateConflictMarkers(merged, conflicts)
	}

	return result, nil
}

// mergeFields performs three-way merge on map fields.
func mergeFields(path string, base, ours, theirs map[string]interface{},
	merged map[string]interface{}, conflicts *[]Conflict, changes *MergeChanges) {

	allKeys := collectKeys(base, ours, theirs)

	for _, key := range allKeys {
		fieldPath := path + "." + key
		baseVal, baseOk := base[key]
		oursVal, oursOk := ours[key]
		theirsVal, theirsOk := theirs[key]

		// Apply three-way merge semantics per TDD Section 7.1
		switch {
		case !baseOk && !oursOk && theirsOk:
			// New in theirs only -> accept theirs
			merged[key] = deepCopyValue(theirsVal)
			changes.FromTheirs = append(changes.FromTheirs, fieldPath)

		case !baseOk && oursOk && !theirsOk:
			// New in ours only -> accept ours
			merged[key] = deepCopyValue(oursVal)
			changes.FromOurs = append(changes.FromOurs, fieldPath)

		case baseOk && !oursOk && !theirsOk:
			// Deleted in both -> delete (don't add to merged)

		case baseOk && !oursOk && theirsOk:
			// Deleted in ours, possibly modified in theirs
			if !equal(baseVal, theirsVal) {
				// Conflict: ours deleted, theirs modified
				*conflicts = append(*conflicts, Conflict{
					Path:        fieldPath,
					BaseValue:   baseVal,
					OursValue:   nil,
					TheirsValue: theirsVal,
				})
				// Keep theirs as default
				merged[key] = deepCopyValue(theirsVal)
			}
			// If theirs unchanged from base, accept deletion

		case baseOk && oursOk && !theirsOk:
			// Modified in ours, deleted in theirs
			if !equal(baseVal, oursVal) {
				// Conflict: ours modified, theirs deleted
				*conflicts = append(*conflicts, Conflict{
					Path:        fieldPath,
					BaseValue:   baseVal,
					OursValue:   oursVal,
					TheirsValue: nil,
				})
				// Keep ours as default
				merged[key] = deepCopyValue(oursVal)
				changes.FromOurs = append(changes.FromOurs, fieldPath)
			}
			// If ours unchanged from base, accept deletion

		case baseOk && oursOk && theirsOk:
			oursChanged := !equal(baseVal, oursVal)
			theirsChanged := !equal(baseVal, theirsVal)

			switch {
			case !oursChanged && !theirsChanged:
				// Neither modified -> keep base
				merged[key] = deepCopyValue(baseVal)

			case oursChanged && !theirsChanged:
				// Only ours modified -> accept ours
				merged[key] = deepCopyValue(oursVal)
				changes.FromOurs = append(changes.FromOurs, fieldPath)

			case !oursChanged && theirsChanged:
				// Only theirs modified -> accept theirs
				merged[key] = deepCopyValue(theirsVal)
				changes.FromTheirs = append(changes.FromTheirs, fieldPath)

			case oursChanged && theirsChanged:
				if equal(oursVal, theirsVal) {
					// Both changed to same value -> use it
					merged[key] = deepCopyValue(oursVal)
				} else {
					// Both changed differently - check if nested object
					if oursMap, ok := oursVal.(map[string]interface{}); ok {
						if theirsMap, ok := theirsVal.(map[string]interface{}); ok {
							if baseMap, ok := baseVal.(map[string]interface{}); ok {
								// Recursively merge nested objects
								nestedMerged := make(map[string]interface{})
								mergeFields(fieldPath, baseMap, oursMap, theirsMap, nestedMerged, conflicts, changes)
								merged[key] = nestedMerged
								continue
							}
						}
					}
					// Scalar conflict
					*conflicts = append(*conflicts, Conflict{
						Path:        fieldPath,
						BaseValue:   baseVal,
						OursValue:   oursVal,
						TheirsValue: theirsVal,
					})
					// Use ours as default, add markers later
					merged[key] = deepCopyValue(oursVal)
				}
			}

		case !baseOk && oursOk && theirsOk:
			// New in both
			if equal(oursVal, theirsVal) {
				merged[key] = deepCopyValue(oursVal)
			} else {
				*conflicts = append(*conflicts, Conflict{
					Path:        fieldPath,
					BaseValue:   nil,
					OursValue:   oursVal,
					TheirsValue: theirsVal,
				})
				merged[key] = deepCopyValue(oursVal)
			}
		}
	}
}

// mergeFieldsUnion merges with array union strategy.
func mergeFieldsUnion(path string, base, ours, theirs map[string]interface{},
	merged map[string]interface{}, conflicts *[]Conflict, changes *MergeChanges) {

	allKeys := collectKeys(base, ours, theirs)

	for _, key := range allKeys {
		fieldPath := path + "." + key
		baseVal, baseOk := base[key]
		oursVal, oursOk := ours[key]
		theirsVal, theirsOk := theirs[key]

		// Handle arrays with union
		if oursArr, oursIsArr := oursVal.([]interface{}); oursIsArr {
			if theirsArr, theirsIsArr := theirsVal.([]interface{}); theirsIsArr {
				merged[key] = unionArrays(oursArr, theirsArr)
				continue
			}
		}

		// Fall back to smart merge for non-arrays
		switch {
		case !baseOk && !oursOk && theirsOk:
			merged[key] = deepCopyValue(theirsVal)
			changes.FromTheirs = append(changes.FromTheirs, fieldPath)
		case !baseOk && oursOk && !theirsOk:
			merged[key] = deepCopyValue(oursVal)
			changes.FromOurs = append(changes.FromOurs, fieldPath)
		case baseOk && oursOk && theirsOk:
			if equal(oursVal, theirsVal) {
				merged[key] = deepCopyValue(oursVal)
			} else if equal(baseVal, oursVal) {
				merged[key] = deepCopyValue(theirsVal)
				changes.FromTheirs = append(changes.FromTheirs, fieldPath)
			} else if equal(baseVal, theirsVal) {
				merged[key] = deepCopyValue(oursVal)
				changes.FromOurs = append(changes.FromOurs, fieldPath)
			} else {
				*conflicts = append(*conflicts, Conflict{
					Path:        fieldPath,
					BaseValue:   baseVal,
					OursValue:   oursVal,
					TheirsValue: theirsVal,
				})
				merged[key] = deepCopyValue(oursVal)
			}
		case oursOk:
			merged[key] = deepCopyValue(oursVal)
		case theirsOk:
			merged[key] = deepCopyValue(theirsVal)
		}
	}
}

// unionArrays merges two arrays without duplicates.
func unionArrays(ours, theirs []interface{}) []interface{} {
	seen := make(map[string]bool)
	result := []interface{}{}

	for _, v := range ours {
		key := toJSONKey(v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	for _, v := range theirs {
		key := toJSONKey(v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	return result
}

// collectKeys returns all unique keys from the maps, sorted.
func collectKeys(maps ...map[string]interface{}) []string {
	seen := make(map[string]bool)
	for _, m := range maps {
		for k := range m {
			seen[k] = true
		}
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// equal checks if two values are equal.
func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// deepCopy creates a deep copy of a map.
func deepCopy(m map[string]interface{}) map[string]interface{} {
	data, _ := json.Marshal(m)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	return result
}

// deepCopyValue creates a deep copy of any value.
func deepCopyValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	data, _ := json.Marshal(v)
	var result interface{}
	_ = json.Unmarshal(data, &result)
	return result
}

// generateConflictMarkers creates Git-style conflict markers.
func generateConflictMarkers(merged map[string]interface{}, conflicts []Conflict) string {
	// Create a copy with conflict markers embedded
	data, _ := json.MarshalIndent(merged, "", "  ")
	result := string(data)

	// For each conflict, find the value in the JSON and add markers
	for _, conflict := range conflicts {
		oursJSON, _ := json.Marshal(conflict.OursValue)
		theirsJSON, _ := json.Marshal(conflict.TheirsValue)

		marker := fmt.Sprintf("<<<<<<< ours\n%s\n=======\n%s\n>>>>>>> theirs",
			string(oursJSON), string(theirsJSON))

		// This is a simplified replacement - it works for simple values
		// A more robust implementation would use path-aware replacement
		result = strings.Replace(result, string(oursJSON), marker, 1)
	}

	return result
}

// ToManifest converts a merge result to a Manifest.
func (r *MergeResult) ToManifest(path string, format Format) *Manifest {
	data, _ := json.Marshal(r.Merged)
	return &Manifest{
		Path:    path,
		Format:  format,
		Content: r.Merged,
		Raw:     data,
	}
}
