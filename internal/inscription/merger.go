package inscription

import (
	"strings"

	"github.com/autom8y/knossos/internal/errors"
)

// MergeResult contains the result of a region merge operation.
type MergeResult struct {
	// Content is the final merged content.
	Content string

	// Conflicts contains any conflicts detected during merge.
	Conflicts []Conflict

	// RegionsMerged lists the regions that were processed.
	RegionsMerged []string

	// RegionsPreserved lists satellite regions that were preserved.
	RegionsPreserved []string

	// RegionsOverwritten lists regions where user edits were overwritten.
	RegionsOverwritten []string

	// RegionsDropped lists deprecated regions that were removed during merge.
	RegionsDropped []string
}

// Conflict represents a merge conflict detected during region merging.
type Conflict struct {
	// Region is the name of the conflicting region.
	Region string

	// Type is the conflict type.
	Type ConflictType

	// Message describes the conflict.
	Message string

	// OldHash is the expected hash (from manifest).
	OldHash string

	// NewHash is the actual hash (from current content).
	NewHash string

	// Preserved indicates if the old content was preserved.
	Preserved bool
}

// ConflictType identifies the type of merge conflict.
type ConflictType string

const (
	// ConflictUserEditedKnossos indicates user edited a knossos-owned region.
	ConflictUserEditedKnossos ConflictType = "user_edited_knossos"

	// ConflictUserEditedRegenerate indicates user edited a regenerate region.
	ConflictUserEditedRegenerate ConflictType = "user_edited_regenerate"

	// ConflictOverlappingRegions indicates regions overlap (invalid state).
	ConflictOverlappingRegions ConflictType = "overlapping_regions"

	// ConflictMalformedMarkers indicates marker parsing errors.
	ConflictMalformedMarkers ConflictType = "malformed_markers"
)

// Merger handles merging of generated content with existing CLAUDE.md content.
type Merger struct {
	// Manifest is the current KNOSSOS_MANIFEST.yaml.
	Manifest *Manifest

	// Generator produces content for knossos and regenerate regions.
	Generator *Generator

	// Parser parses existing content for markers.
	Parser *MarkerParser

	// DeprecatedRegions contains region names that should be dropped during
	// merge instead of being adopted as satellite. Populated from both the
	// static DeprecatedRegions() list and dynamic detection of non-satellite
	// regions the generator can no longer produce.
	DeprecatedRegions map[string]bool
}

// NewMerger creates a new merger with the given configuration.
// Automatically detects deprecated regions from both the static list and
// dynamic analysis of the manifest against the generator's capabilities.
func NewMerger(manifest *Manifest, generator *Generator) *Merger {
	deprecated := make(map[string]bool)

	// Layer 1: static list of historically deprecated regions
	for _, name := range DeprecatedRegions() {
		deprecated[name] = true
	}

	// Layer 2: dynamic detection — non-satellite manifest regions the generator can't produce
	if generator != nil && manifest != nil {
		for name, region := range manifest.Regions {
			if region.Owner == OwnerSatellite {
				continue
			}
			if !generator.CanGenerateRegion(name) {
				deprecated[name] = true
			}
		}
	}

	// Safety: never deprecate a region the user explicitly owns as satellite.
	// The static list might contain a name that a user chose for their satellite.
	if manifest != nil {
		for name := range deprecated {
			if region := manifest.GetRegion(name); region != nil && region.Owner == OwnerSatellite {
				delete(deprecated, name)
			}
		}
	}

	return &Merger{
		Manifest:          manifest,
		Generator:         generator,
		Parser:            NewMarkerParser(),
		DeprecatedRegions: deprecated,
	}
}

// MergeRegions merges generated content with existing CLAUDE.md content.
// Implements the merge algorithm from TDD Section 5.2 Stage 4.
func (m *Merger) MergeRegions(existingContent string, generatedContent map[string]string) (*MergeResult, error) {
	result := &MergeResult{
		Conflicts:          make([]Conflict, 0),
		RegionsMerged:      make([]string, 0),
		RegionsPreserved:   make([]string, 0),
		RegionsOverwritten: make([]string, 0),
		RegionsDropped:     make([]string, 0),
	}

	// Parse existing content to extract regions
	parseResult := m.Parser.Parse(existingContent)

	// Check for parse errors (treat malformed content as satellite)
	if parseResult.HasErrors() {
		for _, err := range parseResult.Errors {
			result.Conflicts = append(result.Conflicts, Conflict{
				Region:  err.Raw,
				Type:    ConflictMalformedMarkers,
				Message: err.Message,
			})
		}
	}

	// Build the merged output
	var output strings.Builder

	// Track which regions we've processed
	processedRegions := make(map[string]bool)

	// Process sections in order from manifest
	for _, regionName := range m.Manifest.SectionOrder {
		region := m.Manifest.GetRegion(regionName)
		if region == nil {
			continue
		}

		// Get existing content for this region
		existingRegion := parseResult.GetRegion(regionName)
		oldContent := ""
		if existingRegion != nil {
			oldContent = existingRegion.Content
		}

		// Get generated content for this region
		newContent := ""
		if content, ok := generatedContent[regionName]; ok {
			newContent = content
		}

		// Merge based on owner type
		mergedContent, conflict := m.mergeRegion(regionName, region, oldContent, newContent)

		if conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			if conflict.Preserved {
				result.RegionsPreserved = append(result.RegionsPreserved, regionName)
			} else {
				result.RegionsOverwritten = append(result.RegionsOverwritten, regionName)
			}
		}

		// Wrap content with markers and add to output
		if mergedContent != "" {
			wrapped := m.wrapWithMarkers(regionName, region, mergedContent)
			output.WriteString(wrapped)
			output.WriteString("\n\n")
		}

		processedRegions[regionName] = true
		result.RegionsMerged = append(result.RegionsMerged, regionName)
	}

	// Append unknown sections from existing content (satellite-owned by default).
	// Skip deprecated regions to prevent zombie adoption of stale knossos content.
	for name, parsedRegion := range parseResult.Regions {
		if processedRegions[name] {
			continue
		}

		// Drop deprecated regions instead of adopting as satellite
		if m.DeprecatedRegions[name] {
			if m.Manifest.HasRegion(name) {
				m.Manifest.RemoveRegion(name)
			}
			result.RegionsDropped = append(result.RegionsDropped, name)
			continue
		}

		// Unknown region - treat as satellite
		if !m.Manifest.HasRegion(name) {
			// Add to manifest as satellite
			m.Manifest.SetRegion(name, &Region{
				Owner: OwnerSatellite,
			})
		}

		// Preserve the content
		region := m.Manifest.GetRegion(name)
		wrapped := m.wrapWithMarkers(name, region, parsedRegion.Content)
		output.WriteString(wrapped)
		output.WriteString("\n\n")

		result.RegionsPreserved = append(result.RegionsPreserved, name)
		result.RegionsMerged = append(result.RegionsMerged, name)
	}

	// Clean deprecated regions from manifest (covers regions that were in
	// SectionOrder and got processed above but should still be removed)
	for name := range m.DeprecatedRegions {
		if m.Manifest.HasRegion(name) {
			m.Manifest.RemoveRegion(name)
			result.RegionsDropped = append(result.RegionsDropped, name)
		}
	}

	result.Content = strings.TrimSpace(output.String())
	return result, nil
}

// mergeRegion merges a single region based on owner type.
// Returns the merged content and any conflict detected.
func (m *Merger) mergeRegion(regionName string, region *Region, oldContent, newContent string) (string, *Conflict) {
	switch region.Owner {
	case OwnerSatellite:
		// Never overwrite satellite regions
		if oldContent != "" {
			return oldContent, nil
		}
		// No existing content, use new content if available
		return newContent, nil

	case OwnerKnossos:
		// Always sync knossos regions
		// Check if user edited (for warning)
		if oldContent != "" && region.Hash != "" {
			currentHash := ComputeContentHash(oldContent)
			if currentHash != region.Hash {
				// User modified knossos-owned content
				return newContent, &Conflict{
					Region:    regionName,
					Type:      ConflictUserEditedKnossos,
					Message:   "User edits overwritten in knossos-owned region",
					OldHash:   region.Hash,
					NewHash:   currentHash,
					Preserved: false,
				}
			}
		}
		return newContent, nil

	case OwnerRegenerate:
		// Check if user modified regenerated content
		if oldContent != "" && region.Hash != "" {
			currentHash := ComputeContentHash(oldContent)
			if currentHash != region.Hash {
				// User modified regenerated content
				if region.PreserveOnConflict {
					// Preserve user edits
					return oldContent, &Conflict{
						Region:    regionName,
						Type:      ConflictUserEditedRegenerate,
						Message:   "User edits preserved in regenerate region",
						OldHash:   region.Hash,
						NewHash:   currentHash,
						Preserved: true,
					}
				}
				// Overwrite user edits
				return newContent, &Conflict{
					Region:    regionName,
					Type:      ConflictUserEditedRegenerate,
					Message:   "User edits overwritten in regenerate region",
					OldHash:   region.Hash,
					NewHash:   currentHash,
					Preserved: false,
				}
			}
		}
		// Clean regenerate (no user modifications)
		return newContent, nil

	default:
		// Unknown owner - treat as satellite
		return oldContent, nil
	}
}

// isWrapped checks if content already contains KNOSSOS markers for the given region.
// This is a safety check to prevent double-wrapping.
func isWrapped(content, regionName string) bool {
	startMarker := "<!-- KNOSSOS:START " + regionName
	endMarker := "<!-- KNOSSOS:END " + regionName
	return strings.Contains(content, startMarker) && strings.Contains(content, endMarker)
}

// wrapWithMarkers wraps content with KNOSSOS START/END markers.
// Detects and prevents double-wrapping of already-wrapped content.
func (m *Merger) wrapWithMarkers(regionName string, region *Region, content string) string {
	// Safety check: detect pre-wrapped content to prevent double-wrapping
	if isWrapped(content, regionName) {
		// Content already has markers - return as-is to prevent duplicates
		return content
	}

	options := make(map[string]string)

	// Add options for regenerate regions
	if region != nil && region.Owner == OwnerRegenerate {
		options["regenerate"] = "true"
		if region.Source != "" {
			options["source"] = region.Source
		}
	}

	return WrapContent(regionName, content, options)
}

// InscriptionMergeOptions configures the inscription merge behavior.
type InscriptionMergeOptions struct {
	// Force overwrites all content regardless of ownership.
	Force bool

	// PreserveAllConflicts preserves user edits even in non-preservable regions.
	PreserveAllConflicts bool

	// DryRun previews changes without applying.
	DryRun bool
}

// MergeWithOptions merges with custom options.
func (m *Merger) MergeWithOptions(existingContent string, generatedContent map[string]string, opts InscriptionMergeOptions) (*MergeResult, error) {
	if opts.Force {
		// Force mode: overwrite everything except satellite
		return m.forceMerge(existingContent, generatedContent)
	}

	if opts.PreserveAllConflicts {
		// Preserve all user edits
		return m.preserveMerge(existingContent, generatedContent)
	}

	// Standard merge
	return m.MergeRegions(existingContent, generatedContent)
}

// forceMerge overwrites all non-satellite content.
func (m *Merger) forceMerge(existingContent string, generatedContent map[string]string) (*MergeResult, error) {
	// Parse existing for satellite content
	parseResult := m.Parser.Parse(existingContent)

	result := &MergeResult{
		Conflicts:          make([]Conflict, 0),
		RegionsMerged:      make([]string, 0),
		RegionsPreserved:   make([]string, 0),
		RegionsOverwritten: make([]string, 0),
		RegionsDropped:     make([]string, 0),
	}

	var output strings.Builder

	// Process all sections in order
	for _, regionName := range m.Manifest.SectionOrder {
		region := m.Manifest.GetRegion(regionName)
		if region == nil {
			continue
		}

		var content string
		if region.Owner == OwnerSatellite {
			// Preserve satellite content
			if existing := parseResult.GetRegion(regionName); existing != nil {
				content = existing.Content
				result.RegionsPreserved = append(result.RegionsPreserved, regionName)
			}
		} else {
			// Use generated content
			if generated, ok := generatedContent[regionName]; ok {
				content = generated
			}
			result.RegionsOverwritten = append(result.RegionsOverwritten, regionName)
		}

		if content != "" {
			wrapped := m.wrapWithMarkers(regionName, region, content)
			output.WriteString(wrapped)
			output.WriteString("\n\n")
		}

		result.RegionsMerged = append(result.RegionsMerged, regionName)
	}

	// Clean deprecated regions from manifest
	for name := range m.DeprecatedRegions {
		if m.Manifest.HasRegion(name) {
			m.Manifest.RemoveRegion(name)
			result.RegionsDropped = append(result.RegionsDropped, name)
		}
	}

	result.Content = strings.TrimSpace(output.String())
	return result, nil
}

// preserveMerge preserves all user modifications.
func (m *Merger) preserveMerge(existingContent string, generatedContent map[string]string) (*MergeResult, error) {
	parseResult := m.Parser.Parse(existingContent)

	result := &MergeResult{
		Conflicts:          make([]Conflict, 0),
		RegionsMerged:      make([]string, 0),
		RegionsPreserved:   make([]string, 0),
		RegionsOverwritten: make([]string, 0),
		RegionsDropped:     make([]string, 0),
	}

	var output strings.Builder

	for _, regionName := range m.Manifest.SectionOrder {
		region := m.Manifest.GetRegion(regionName)
		if region == nil {
			continue
		}

		existingRegion := parseResult.GetRegion(regionName)
		var content string

		if existingRegion != nil && existingRegion.Content != "" {
			// Preserve existing content if present
			content = existingRegion.Content
			result.RegionsPreserved = append(result.RegionsPreserved, regionName)
		} else if generated, ok := generatedContent[regionName]; ok {
			// Use generated content if no existing
			content = generated
		}

		if content != "" {
			wrapped := m.wrapWithMarkers(regionName, region, content)
			output.WriteString(wrapped)
			output.WriteString("\n\n")
		}

		result.RegionsMerged = append(result.RegionsMerged, regionName)
	}

	// Clean deprecated regions from manifest
	for name := range m.DeprecatedRegions {
		if m.Manifest.HasRegion(name) {
			m.Manifest.RemoveRegion(name)
			result.RegionsDropped = append(result.RegionsDropped, name)
		}
	}

	result.Content = strings.TrimSpace(output.String())
	return result, nil
}

// DetectConflicts checks for conflicts without merging.
func (m *Merger) DetectConflicts(existingContent string) ([]Conflict, error) {
	parseResult := m.Parser.Parse(existingContent)
	conflicts := make([]Conflict, 0)

	// Check for parse errors
	if parseResult.HasErrors() {
		for _, err := range parseResult.Errors {
			conflicts = append(conflicts, Conflict{
				Region:  err.Raw,
				Type:    ConflictMalformedMarkers,
				Message: err.Message,
			})
		}
	}

	// Check each region for hash mismatches
	for regionName, existingRegion := range parseResult.Regions {
		region := m.Manifest.GetRegion(regionName)
		if region == nil {
			continue
		}

		// Skip satellite regions (no conflict possible)
		if region.Owner == OwnerSatellite {
			continue
		}

		// Check hash
		if region.Hash != "" {
			currentHash := ComputeContentHash(existingRegion.Content)
			if currentHash != region.Hash {
				conflictType := ConflictUserEditedKnossos
				if region.Owner == OwnerRegenerate {
					conflictType = ConflictUserEditedRegenerate
				}

				conflicts = append(conflicts, Conflict{
					Region:    regionName,
					Type:      conflictType,
					Message:   "Content modified since last sync",
					OldHash:   region.Hash,
					NewHash:   currentHash,
					Preserved: region.Owner == OwnerRegenerate && region.PreserveOnConflict,
				})
			}
		}
	}

	// Check for overlapping regions (shouldn't happen with proper markers)
	if err := m.checkOverlappingRegions(parseResult); err != nil {
		conflicts = append(conflicts, Conflict{
			Type:    ConflictOverlappingRegions,
			Message: err.Error(),
		})
	}

	return conflicts, nil
}

// checkOverlappingRegions validates that no regions overlap.
func (m *Merger) checkOverlappingRegions(parseResult *ParseResult) error {
	// Sort regions by start line
	type regionLine struct {
		name  string
		start int
		end   int
	}

	regions := make([]regionLine, 0, len(parseResult.Regions))
	for name, region := range parseResult.Regions {
		regions = append(regions, regionLine{
			name:  name,
			start: region.StartLine,
			end:   region.EndLine,
		})
	}

	// Simple O(n^2) check for overlaps
	for i := 0; i < len(regions); i++ {
		for j := i + 1; j < len(regions); j++ {
			a, b := regions[i], regions[j]

			// Check if regions overlap
			if a.start < b.end && b.start < a.end {
				// Regions overlap (but adjacent is OK)
				if !(a.end == b.start || b.end == a.start) {
					return errors.NewWithDetails(errors.CodeUsageError,
						"overlapping regions detected",
						map[string]interface{}{
							"region_a": a.name,
							"region_b": b.name,
						})
				}
			}
		}
	}

	return nil
}

// ValidateMerge validates that a merge can proceed.
func (m *Merger) ValidateMerge(existingContent string, generatedContent map[string]string) error {
	// Parse existing content
	parseResult := m.Parser.Parse(existingContent)

	// Check for fatal parse errors
	if parseResult.HasErrors() {
		for _, err := range parseResult.Errors {
			// Unclosed regions are fatal
			if strings.Contains(err.Message, "without matching") {
				return errors.NewWithDetails(errors.CodeParseError,
					"malformed markers prevent merge",
					map[string]interface{}{
						"error": err.Message,
						"line":  err.Line,
					})
			}
		}
	}

	// Check for overlapping regions
	if err := m.checkOverlappingRegions(parseResult); err != nil {
		return err
	}

	// Validate all generated content regions exist in manifest
	for regionName := range generatedContent {
		if !m.Manifest.HasRegion(regionName) {
			return errors.NewWithDetails(errors.CodeUsageError,
				"generated content for unknown region",
				map[string]interface{}{"region": regionName})
		}
	}

	return nil
}

// UpdateManifestHashes updates the manifest with new content hashes.
func (m *Merger) UpdateManifestHashes(result *MergeResult) {
	parseResult := m.Parser.Parse(result.Content)

	for regionName, region := range parseResult.Regions {
		manifestRegion := m.Manifest.GetRegion(regionName)
		if manifestRegion == nil {
			continue
		}

		// Update hash for non-satellite regions
		if manifestRegion.Owner != OwnerSatellite {
			manifestRegion.Hash = ComputeContentHash(region.Content)
		}
	}
}
