package inscription

import (
	"strings"
	"testing"
)

func TestNewMerger(t *testing.T) {
	manifest := &Manifest{}
	gen := &Generator{}

	merger := NewMerger(manifest, gen)

	if merger.Manifest != manifest {
		t.Error("NewMerger() Manifest not set")
	}
	if merger.Generator != gen {
		t.Error("NewMerger() Generator not set")
	}
	if merger.Parser == nil {
		t.Error("NewMerger() Parser should be initialized")
	}
}

func TestMerger_MergeRegions_Basic(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"section-a": {Owner: OwnerKnossos},
			"section-b": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"section-a", "section-b"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START section-a -->
Old content A
<!-- KNOSSOS:END section-a -->

<!-- KNOSSOS:START section-b -->
Old content B
<!-- KNOSSOS:END section-b -->`

	generated := map[string]string{
		"section-a": "New content A",
		"section-b": "New content B",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	if !strings.Contains(result.Content, "New content A") {
		t.Error("MergeRegions() should contain new content A")
	}
	if !strings.Contains(result.Content, "New content B") {
		t.Error("MergeRegions() should contain new content B")
	}
	if strings.Contains(result.Content, "Old content") {
		t.Error("MergeRegions() should not contain old content (knossos regions overwritten)")
	}
}

func TestMerger_MergeRegions_SatellitePreserved(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"knossos-section":   {Owner: OwnerKnossos},
			"satellite-section": {Owner: OwnerSatellite},
		},
		SectionOrder: []string{"knossos-section", "satellite-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START knossos-section -->
Old knossos content
<!-- KNOSSOS:END knossos-section -->

<!-- KNOSSOS:START satellite-section -->
Custom satellite content - should be preserved
<!-- KNOSSOS:END satellite-section -->`

	generated := map[string]string{
		"knossos-section": "New knossos content",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// Knossos content should be new
	if !strings.Contains(result.Content, "New knossos content") {
		t.Error("MergeRegions() should contain new knossos content")
	}

	// Satellite content should be preserved
	if !strings.Contains(result.Content, "Custom satellite content - should be preserved") {
		t.Error("MergeRegions() should preserve satellite content")
	}
}

func TestMerger_MergeRegions_RegenerateClean(t *testing.T) {
	// Hash of "Old regenerate content"
	oldHash := ComputeContentHash("Old regenerate content")

	manifest := &Manifest{
		Regions: map[string]*Region{
			"regenerate-section": {
				Owner:  OwnerRegenerate,
				Source: "test",
				Hash:   oldHash,
			},
		},
		SectionOrder: []string{"regenerate-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START regenerate-section regenerate=true source=test -->
Old regenerate content
<!-- KNOSSOS:END regenerate-section -->`

	generated := map[string]string{
		"regenerate-section": "New regenerate content",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// Clean regenerate - should use new content
	if !strings.Contains(result.Content, "New regenerate content") {
		t.Error("MergeRegions() clean regenerate should use new content")
	}

	// No conflicts for clean regenerate
	if len(result.Conflicts) > 0 {
		t.Errorf("MergeRegions() clean regenerate should have no conflicts, got %v", result.Conflicts)
	}
}

func TestMerger_MergeRegions_RegenerateModified_PreserveOnConflict(t *testing.T) {
	// Hash of original content (not what's in file)
	oldHash := ComputeContentHash("Original content")

	manifest := &Manifest{
		Regions: map[string]*Region{
			"regenerate-section": {
				Owner:              OwnerRegenerate,
				Source:             "test",
				Hash:               oldHash,
				PreserveOnConflict: true,
			},
		},
		SectionOrder: []string{"regenerate-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	// User modified the content
	existing := `<!-- KNOSSOS:START regenerate-section regenerate=true source=test -->
User modified content
<!-- KNOSSOS:END regenerate-section -->`

	generated := map[string]string{
		"regenerate-section": "New regenerate content",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// User edits should be preserved
	if !strings.Contains(result.Content, "User modified content") {
		t.Error("MergeRegions() should preserve user modified content")
	}

	// Should have conflict
	if len(result.Conflicts) != 1 {
		t.Fatalf("MergeRegions() expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Type != ConflictUserEditedRegenerate {
		t.Errorf("MergeRegions() conflict type = %v, want ConflictUserEditedRegenerate", result.Conflicts[0].Type)
	}
	if !result.Conflicts[0].Preserved {
		t.Error("MergeRegions() conflict should indicate preserved=true")
	}
}

func TestMerger_MergeRegions_RegenerateModified_Overwrite(t *testing.T) {
	oldHash := ComputeContentHash("Original content")

	manifest := &Manifest{
		Regions: map[string]*Region{
			"regenerate-section": {
				Owner:              OwnerRegenerate,
				Source:             "test",
				Hash:               oldHash,
				PreserveOnConflict: false, // Default - overwrite user edits
			},
		},
		SectionOrder: []string{"regenerate-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START regenerate-section regenerate=true source=test -->
User modified content
<!-- KNOSSOS:END regenerate-section -->`

	generated := map[string]string{
		"regenerate-section": "New regenerate content",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// User edits should be overwritten
	if !strings.Contains(result.Content, "New regenerate content") {
		t.Error("MergeRegions() should overwrite with new content")
	}

	// Should have conflict
	if len(result.Conflicts) != 1 {
		t.Fatalf("MergeRegions() expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Preserved {
		t.Error("MergeRegions() conflict should indicate preserved=false")
	}
}

func TestMerger_MergeRegions_KnossosModified(t *testing.T) {
	oldHash := ComputeContentHash("Original knossos content")

	manifest := &Manifest{
		Regions: map[string]*Region{
			"knossos-section": {
				Owner: OwnerKnossos,
				Hash:  oldHash,
			},
		},
		SectionOrder: []string{"knossos-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	// User edited knossos region (shouldn't do this!)
	existing := `<!-- KNOSSOS:START knossos-section -->
User edited knossos content
<!-- KNOSSOS:END knossos-section -->`

	generated := map[string]string{
		"knossos-section": "New knossos content",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// Knossos always overwrites
	if !strings.Contains(result.Content, "New knossos content") {
		t.Error("MergeRegions() knossos should always use new content")
	}

	// Should have conflict warning
	if len(result.Conflicts) != 1 {
		t.Fatalf("MergeRegions() expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Type != ConflictUserEditedKnossos {
		t.Errorf("MergeRegions() conflict type = %v, want ConflictUserEditedKnossos", result.Conflicts[0].Type)
	}
}

func TestMerger_MergeRegions_UnknownRegionPreserved(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"known-section": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"known-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	// Existing has a region not in manifest
	existing := `<!-- KNOSSOS:START known-section -->
Known content
<!-- KNOSSOS:END known-section -->

<!-- KNOSSOS:START unknown-section -->
Custom unknown content
<!-- KNOSSOS:END unknown-section -->`

	generated := map[string]string{
		"known-section": "New known content",
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// Unknown region should be preserved
	if !strings.Contains(result.Content, "Custom unknown content") {
		t.Error("MergeRegions() should preserve unknown regions")
	}

	// Unknown region should be added to manifest as satellite
	if !manifest.HasRegion("unknown-section") {
		t.Error("MergeRegions() should add unknown region to manifest")
	}
	if manifest.GetRegion("unknown-section").Owner != OwnerSatellite {
		t.Error("MergeRegions() unknown region should be satellite")
	}
}

func TestMerger_MergeRegions_MalformedMarkers(t *testing.T) {
	manifest := &Manifest{
		Regions:      map[string]*Region{},
		SectionOrder: []string{},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	// Malformed markers (START without END)
	existing := `<!-- KNOSSOS:START orphan-section -->
Content without end`

	result, err := merger.MergeRegions(existing, map[string]string{})
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// Should have conflict for malformed markers
	if len(result.Conflicts) == 0 {
		t.Error("MergeRegions() should detect malformed markers")
	}

	foundMalformed := false
	for _, c := range result.Conflicts {
		if c.Type == ConflictMalformedMarkers {
			foundMalformed = true
			break
		}
	}
	if !foundMalformed {
		t.Error("MergeRegions() should have ConflictMalformedMarkers")
	}
}

func TestMerger_DetectConflicts(t *testing.T) {
	oldHash := ComputeContentHash("Original content")

	manifest := &Manifest{
		Regions: map[string]*Region{
			"modified-section": {
				Owner: OwnerKnossos,
				Hash:  oldHash,
			},
			"unchanged-section": {
				Owner: OwnerKnossos,
				Hash:  ComputeContentHash("Unchanged content"),
			},
		},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START modified-section -->
User edited content
<!-- KNOSSOS:END modified-section -->

<!-- KNOSSOS:START unchanged-section -->
Unchanged content
<!-- KNOSSOS:END unchanged-section -->`

	conflicts, err := merger.DetectConflicts(existing)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	// Should detect only the modified section
	if len(conflicts) != 1 {
		t.Fatalf("DetectConflicts() expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].Region != "modified-section" {
		t.Errorf("DetectConflicts() conflict region = %q, want 'modified-section'", conflicts[0].Region)
	}
}

func TestMerger_MergeWithOptions_Force(t *testing.T) {
	oldHash := ComputeContentHash("Original content")

	manifest := &Manifest{
		Regions: map[string]*Region{
			"knossos-section": {Owner: OwnerKnossos, Hash: oldHash},
			"satellite-section": {Owner: OwnerSatellite},
		},
		SectionOrder: []string{"knossos-section", "satellite-section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START knossos-section -->
User edited knossos
<!-- KNOSSOS:END knossos-section -->

<!-- KNOSSOS:START satellite-section -->
Satellite content
<!-- KNOSSOS:END satellite-section -->`

	generated := map[string]string{
		"knossos-section": "Forced new content",
	}

	result, err := merger.MergeWithOptions(existing, generated, InscriptionMergeOptions{Force: true})
	if err != nil {
		t.Fatalf("MergeWithOptions() error = %v", err)
	}

	// Force should overwrite knossos
	if !strings.Contains(result.Content, "Forced new content") {
		t.Error("MergeWithOptions() force should use new content")
	}

	// Satellite should still be preserved
	if !strings.Contains(result.Content, "Satellite content") {
		t.Error("MergeWithOptions() force should still preserve satellite")
	}
}

func TestMerger_MergeWithOptions_PreserveAllConflicts(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"section": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START section -->
Existing content
<!-- KNOSSOS:END section -->`

	generated := map[string]string{
		"section": "New content",
	}

	result, err := merger.MergeWithOptions(existing, generated, InscriptionMergeOptions{PreserveAllConflicts: true})
	if err != nil {
		t.Fatalf("MergeWithOptions() error = %v", err)
	}

	// Preserve mode should keep existing content
	if !strings.Contains(result.Content, "Existing content") {
		t.Error("MergeWithOptions() preserve mode should keep existing")
	}
}

func TestMerger_ValidateMerge(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"known": {Owner: OwnerKnossos},
		},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	// Valid content
	err := merger.ValidateMerge(`<!-- KNOSSOS:START known -->
Content
<!-- KNOSSOS:END known -->`, map[string]string{"known": "new"})
	if err != nil {
		t.Errorf("ValidateMerge() valid content error = %v", err)
	}

	// Invalid - malformed markers
	err = merger.ValidateMerge(`<!-- KNOSSOS:START orphan -->
No end`, map[string]string{})
	if err == nil {
		t.Error("ValidateMerge() should error on malformed markers")
	}

	// Invalid - unknown region in generated content
	err = merger.ValidateMerge(``, map[string]string{"unknown": "content"})
	if err == nil {
		t.Error("ValidateMerge() should error on unknown generated region")
	}
}

func TestMerger_UpdateManifestHashes(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos},
		},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	result := &MergeResult{
		Content: `<!-- KNOSSOS:START test -->
New content here
<!-- KNOSSOS:END test -->`,
	}

	merger.UpdateManifestHashes(result)

	// Hash should be updated
	region := manifest.GetRegion("test")
	if region.Hash == "" {
		t.Error("UpdateManifestHashes() should set hash")
	}

	expectedHash := ComputeContentHash("New content here")
	if region.Hash != expectedHash {
		t.Errorf("UpdateManifestHashes() hash = %q, want %q", region.Hash, expectedHash)
	}
}

func TestMerger_WrapWithMarkers(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"normal":     {Owner: OwnerKnossos},
			"regenerate": {Owner: OwnerRegenerate, Source: "test"},
		},
	}
	merger := NewMerger(manifest, nil)

	// Normal region
	wrapped := merger.wrapWithMarkers("normal", manifest.GetRegion("normal"), "content")
	if !strings.HasPrefix(wrapped, "<!-- KNOSSOS:START normal -->") {
		t.Error("wrapWithMarkers() normal region marker incorrect")
	}
	if strings.Contains(wrapped, "regenerate=true") {
		t.Error("wrapWithMarkers() normal region should not have regenerate option")
	}

	// Regenerate region
	wrapped = merger.wrapWithMarkers("regenerate", manifest.GetRegion("regenerate"), "content")
	if !strings.Contains(wrapped, "regenerate=true") {
		t.Error("wrapWithMarkers() regenerate region should have regenerate option")
	}
	if !strings.Contains(wrapped, "source=test") {
		t.Error("wrapWithMarkers() regenerate region should have source option")
	}
}

func TestIsWrapped(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		regionName string
		want       bool
	}{
		{
			name:       "wrapped content",
			content:    "<!-- KNOSSOS:START test-region -->\nContent here\n<!-- KNOSSOS:END test-region -->",
			regionName: "test-region",
			want:       true,
		},
		{
			name:       "unwrapped content",
			content:    "Just plain content without markers",
			regionName: "test-region",
			want:       false,
		},
		{
			name:       "different region name",
			content:    "<!-- KNOSSOS:START other-region -->\nContent\n<!-- KNOSSOS:END other-region -->",
			regionName: "test-region",
			want:       false,
		},
		{
			name:       "only start marker",
			content:    "<!-- KNOSSOS:START test-region -->\nContent without end",
			regionName: "test-region",
			want:       false,
		},
		{
			name:       "only end marker",
			content:    "Content without start\n<!-- KNOSSOS:END test-region -->",
			regionName: "test-region",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWrapped(tt.content, tt.regionName)
			if got != tt.want {
				t.Errorf("isWrapped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerger_WrapWithMarkers_PreventsDoubleWrapping(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test-region": {Owner: OwnerKnossos},
		},
	}
	merger := NewMerger(manifest, nil)

	// Pre-wrapped content
	preWrapped := "<!-- KNOSSOS:START test-region -->\nOriginal content\n<!-- KNOSSOS:END test-region -->"

	// Attempt to wrap already-wrapped content
	result := merger.wrapWithMarkers("test-region", manifest.GetRegion("test-region"), preWrapped)

	// Should return content as-is, not double-wrapped
	if result != preWrapped {
		t.Errorf("wrapWithMarkers() should not double-wrap; got:\n%s", result)
	}

	// Verify no double markers
	startCount := strings.Count(result, "<!-- KNOSSOS:START test-region")
	if startCount != 1 {
		t.Errorf("Expected 1 START marker, got %d", startCount)
	}

	endCount := strings.Count(result, "<!-- KNOSSOS:END test-region")
	if endCount != 1 {
		t.Errorf("Expected 1 END marker, got %d", endCount)
	}
}

func TestMerger_WrapWithMarkers_WrapsUnwrappedContent(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test-region": {Owner: OwnerKnossos},
		},
	}
	merger := NewMerger(manifest, nil)

	// Unwrapped content
	unwrapped := "Just plain content"

	result := merger.wrapWithMarkers("test-region", manifest.GetRegion("test-region"), unwrapped)

	// Should be wrapped
	if !strings.HasPrefix(result, "<!-- KNOSSOS:START test-region -->") {
		t.Error("wrapWithMarkers() should wrap unwrapped content with START marker")
	}
	if !strings.HasSuffix(result, "<!-- KNOSSOS:END test-region -->") {
		t.Error("wrapWithMarkers() should wrap unwrapped content with END marker")
	}
	if !strings.Contains(result, "Just plain content") {
		t.Error("wrapWithMarkers() should preserve content")
	}
}

// TestMerger_FullPipeline_TemplateWithMarkers tests the scenario where
// templates already include KNOSSOS markers (like real production templates).
// This reproduces the double-marker bug found in production.
func TestMerger_FullPipeline_TemplateWithMarkers(t *testing.T) {
	// Simulate what happens when GenerateSection returns content WITH markers
	// (because the template file includes them)
	manifest := &Manifest{
		Regions: map[string]*Region{
			"execution-mode": {Owner: OwnerKnossos},
			"quick-start":    {Owner: OwnerRegenerate, Source: "ACTIVE_RITE"},
		},
		SectionOrder: []string{"execution-mode", "quick-start"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	// Existing content (clean, single markers per region)
	existing := `<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Old content.
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

Old quick start.
<!-- KNOSSOS:END quick-start -->`

	// Generated content WITH markers (simulating template output)
	// This is the key - templates include markers but generator still returns them
	generated := map[string]string{
		"execution-mode": `<!-- KNOSSOS:START execution-mode -->
## Execution Mode

New content from template.
<!-- KNOSSOS:END execution-mode -->`,
		"quick-start": `<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

New quick start from template.
<!-- KNOSSOS:END quick-start -->`,
	}

	result, err := merger.MergeRegions(existing, generated)
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	// Count markers - should be exactly 2 START and 2 END
	startCount := strings.Count(result.Content, "<!-- KNOSSOS:START")
	endCount := strings.Count(result.Content, "<!-- KNOSSOS:END")

	t.Logf("Result content:\n%s", result.Content)
	t.Logf("START markers: %d, END markers: %d", startCount, endCount)

	if startCount != 2 {
		t.Errorf("Expected 2 START markers, got %d - DOUBLE WRAP BUG!", startCount)
	}
	if endCount != 2 {
		t.Errorf("Expected 2 END markers, got %d - DOUBLE WRAP BUG!", endCount)
	}

	// Verify content is present
	if !strings.Contains(result.Content, "New content from template") {
		t.Error("Result should contain new content")
	}
}

func TestMergeResult_Fields(t *testing.T) {
	result := &MergeResult{
		Content:            "test content",
		Conflicts:          []Conflict{{Region: "test", Type: ConflictUserEditedKnossos}},
		RegionsMerged:      []string{"region1", "region2"},
		RegionsPreserved:   []string{"satellite"},
		RegionsOverwritten: []string{"region1"},
	}

	if result.Content != "test content" {
		t.Error("MergeResult Content not set")
	}
	if len(result.Conflicts) != 1 {
		t.Error("MergeResult Conflicts not set")
	}
	if len(result.RegionsMerged) != 2 {
		t.Error("MergeResult RegionsMerged not set")
	}
	if len(result.RegionsPreserved) != 1 {
		t.Error("MergeResult RegionsPreserved not set")
	}
	if len(result.RegionsOverwritten) != 1 {
		t.Error("MergeResult RegionsOverwritten not set")
	}
}

func TestConflict_Fields(t *testing.T) {
	conflict := Conflict{
		Region:    "test-region",
		Type:      ConflictUserEditedKnossos,
		Message:   "User edits overwritten",
		OldHash:   "abc123",
		NewHash:   "def456",
		Preserved: false,
	}

	if conflict.Region != "test-region" {
		t.Error("Conflict Region not set")
	}
	if conflict.Type != ConflictUserEditedKnossos {
		t.Error("Conflict Type not set")
	}
	if conflict.Message != "User edits overwritten" {
		t.Error("Conflict Message not set")
	}
	if conflict.OldHash != "abc123" {
		t.Error("Conflict OldHash not set")
	}
	if conflict.NewHash != "def456" {
		t.Error("Conflict NewHash not set")
	}
	if conflict.Preserved {
		t.Error("Conflict Preserved should be false")
	}
}

func TestConflictType_Values(t *testing.T) {
	if ConflictUserEditedKnossos != "user_edited_knossos" {
		t.Error("ConflictUserEditedKnossos value incorrect")
	}
	if ConflictUserEditedRegenerate != "user_edited_regenerate" {
		t.Error("ConflictUserEditedRegenerate value incorrect")
	}
	if ConflictOverlappingRegions != "overlapping_regions" {
		t.Error("ConflictOverlappingRegions value incorrect")
	}
	if ConflictMalformedMarkers != "malformed_markers" {
		t.Error("ConflictMalformedMarkers value incorrect")
	}
}

func TestMerger_CheckOverlappingRegions_NoOverlap(t *testing.T) {
	manifest := &Manifest{}
	merger := NewMerger(manifest, nil)

	parseResult := &ParseResult{
		Regions: map[string]*ParsedRegion{
			"a": {Name: "a", StartLine: 1, EndLine: 5},
			"b": {Name: "b", StartLine: 7, EndLine: 10},
		},
	}

	err := merger.checkOverlappingRegions(parseResult)
	if err != nil {
		t.Errorf("checkOverlappingRegions() non-overlapping error = %v", err)
	}
}

func TestMerger_MergeRegions_EmptyExisting(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"section": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	result, err := merger.MergeRegions("", map[string]string{
		"section": "New content",
	})
	if err != nil {
		t.Fatalf("MergeRegions() error = %v", err)
	}

	if !strings.Contains(result.Content, "New content") {
		t.Error("MergeRegions() empty existing should use new content")
	}
}

func TestMerger_ForceMerge(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"knossos":   {Owner: OwnerKnossos},
			"satellite": {Owner: OwnerSatellite},
		},
		SectionOrder: []string{"knossos", "satellite"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START knossos -->
Old knossos
<!-- KNOSSOS:END knossos -->

<!-- KNOSSOS:START satellite -->
My custom content
<!-- KNOSSOS:END satellite -->`

	result, err := merger.forceMerge(existing, map[string]string{
		"knossos": "Forced knossos",
	})
	if err != nil {
		t.Fatalf("forceMerge() error = %v", err)
	}

	// Force overwrites non-satellite
	if !strings.Contains(result.Content, "Forced knossos") {
		t.Error("forceMerge() should force overwrite knossos")
	}

	// Satellite still preserved
	if !strings.Contains(result.Content, "My custom content") {
		t.Error("forceMerge() should preserve satellite")
	}

	// Check tracking arrays
	if len(result.RegionsOverwritten) != 1 || result.RegionsOverwritten[0] != "knossos" {
		t.Error("forceMerge() should track overwritten regions")
	}
	if len(result.RegionsPreserved) != 1 || result.RegionsPreserved[0] != "satellite" {
		t.Error("forceMerge() should track preserved regions")
	}
}

func TestMerger_PreserveMerge(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"section": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	existing := `<!-- KNOSSOS:START section -->
My existing content
<!-- KNOSSOS:END section -->`

	result, err := merger.preserveMerge(existing, map[string]string{
		"section": "New content",
	})
	if err != nil {
		t.Fatalf("preserveMerge() error = %v", err)
	}

	// Preserve keeps existing
	if !strings.Contains(result.Content, "My existing content") {
		t.Error("preserveMerge() should keep existing content")
	}
	if strings.Contains(result.Content, "New content") {
		t.Error("preserveMerge() should not use new content when existing exists")
	}
}

func TestMerger_PreserveMerge_NoExisting(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"section": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"section"},
	}
	gen := NewGenerator("", manifest, nil)
	merger := NewMerger(manifest, gen)

	result, err := merger.preserveMerge("", map[string]string{
		"section": "New content",
	})
	if err != nil {
		t.Fatalf("preserveMerge() error = %v", err)
	}

	// When no existing, use new content
	if !strings.Contains(result.Content, "New content") {
		t.Error("preserveMerge() should use new content when no existing")
	}
}
