package inscription

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestManifestLoader_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewManifestLoader(tmpDir)

	_, err := loader.Load()
	if err == nil {
		t.Fatal("Load() expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Load() error = %v, want 'not found'", err)
	}
}

func TestManifestLoader_Load_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(manifestDir, 0755)

	manifestPath := filepath.Join(manifestDir, ManifestFileName)
	content := `schema_version: "1.0"
inscription_version: "42"
last_sync: "2026-01-06T10:30:00Z"
active_rite: "10x-dev"
template_path: "knossos/templates/CLAUDE.md.tpl"
regions:
  execution-mode:
    owner: knossos
    hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
  quick-start:
    owner: regenerate
    source: "ACTIVE_RITE+agents"
  project-custom:
    owner: satellite
section_order:
  - execution-mode
  - quick-start
  - project-custom
`
	os.WriteFile(manifestPath, []byte(content), 0644)

	loader := NewManifestLoader(tmpDir)
	manifest, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if manifest.SchemaVersion != "1.0" {
		t.Errorf("Load() SchemaVersion = %q, want '1.0'", manifest.SchemaVersion)
	}

	if manifest.InscriptionVersion != "42" {
		t.Errorf("Load() InscriptionVersion = %q, want '42'", manifest.InscriptionVersion)
	}

	if manifest.ActiveRite != "10x-dev" {
		t.Errorf("Load() ActiveRite = %q, want '10x-dev'", manifest.ActiveRite)
	}

	if len(manifest.Regions) != 3 {
		t.Errorf("Load() got %d regions, want 3", len(manifest.Regions))
	}

	// Check region owners
	if manifest.Regions["execution-mode"].Owner != OwnerKnossos {
		t.Errorf("Load() execution-mode owner = %v, want knossos", manifest.Regions["execution-mode"].Owner)
	}

	if manifest.Regions["quick-start"].Owner != OwnerRegenerate {
		t.Errorf("Load() quick-start owner = %v, want regenerate", manifest.Regions["quick-start"].Owner)
	}

	if manifest.Regions["project-custom"].Owner != OwnerSatellite {
		t.Errorf("Load() project-custom owner = %v, want satellite", manifest.Regions["project-custom"].Owner)
	}
}

func TestManifestLoader_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(manifestDir, 0755)

	manifestPath := filepath.Join(manifestDir, ManifestFileName)
	content := `schema_version: "1.0"
  inscription_version: invalid yaml here
    regions: :::
`
	os.WriteFile(manifestPath, []byte(content), 0644)

	loader := NewManifestLoader(tmpDir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid YAML")
	}
}

func TestManifestLoader_Load_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(manifestDir, 0755)

	manifestPath := filepath.Join(manifestDir, ManifestFileName)
	// Missing inscription_version and regions
	content := `schema_version: "1.0"
`
	os.WriteFile(manifestPath, []byte(content), 0644)

	loader := NewManifestLoader(tmpDir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing required fields")
	}

	// Error should be a schema validation error
	if !strings.Contains(err.Error(), "validation") {
		t.Errorf("Load() error should be validation error, got: %v", err)
	}
}

func TestManifestLoader_Load_InvalidRegenerateWithoutSource(t *testing.T) {
	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(manifestDir, 0755)

	manifestPath := filepath.Join(manifestDir, ManifestFileName)
	content := `schema_version: "1.0"
inscription_version: "1"
regions:
  test-region:
    owner: regenerate
`
	os.WriteFile(manifestPath, []byte(content), 0644)

	loader := NewManifestLoader(tmpDir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("Load() expected error for regenerate without source")
	}

	// Error should be a validation error (source is required for regenerate)
	if !strings.Contains(err.Error(), "validation") {
		t.Errorf("Load() error should be validation error, got: %v", err)
	}
}

func TestManifestLoader_Save(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewManifestLoader(tmpDir)

	now := time.Now().UTC()
	manifest := &Manifest{
		SchemaVersion:      "1.0",
		InscriptionVersion: "1",
		LastSync:           &now,
		ActiveRite:         "test-rite",
		Regions: map[string]*Region{
			"test-region": {
				Owner: OwnerKnossos,
			},
		},
		SectionOrder: []string{"test-region"},
	}

	err := loader.Save(manifest)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	if !loader.Exists() {
		t.Error("Save() did not create manifest file")
	}

	// Load back and verify
	loaded, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.SchemaVersion != "1.0" {
		t.Errorf("Load() SchemaVersion = %q, want '1.0'", loaded.SchemaVersion)
	}

	if loaded.ActiveRite != "test-rite" {
		t.Errorf("Load() ActiveRite = %q, want 'test-rite'", loaded.ActiveRite)
	}
}

func TestManifestLoader_LoadOrCreate_Create(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewManifestLoader(tmpDir)

	manifest, err := loader.LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}

	if manifest.SchemaVersion != DefaultSchemaVersion {
		t.Errorf("LoadOrCreate() SchemaVersion = %q, want %q", manifest.SchemaVersion, DefaultSchemaVersion)
	}

	if manifest.InscriptionVersion != "1" {
		t.Errorf("LoadOrCreate() InscriptionVersion = %q, want '1'", manifest.InscriptionVersion)
	}

	// Should have default regions
	if manifest.GetRegion("execution-mode") == nil {
		t.Error("LoadOrCreate() missing default region 'execution-mode'")
	}

	if manifest.GetRegion("quick-start") == nil {
		t.Error("LoadOrCreate() missing default region 'quick-start'")
	}
}

func TestManifestLoader_LoadOrCreate_Load(t *testing.T) {
	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(manifestDir, 0755)

	manifestPath := filepath.Join(manifestDir, ManifestFileName)
	content := `schema_version: "1.0"
inscription_version: "99"
regions:
  custom-region:
    owner: satellite
`
	os.WriteFile(manifestPath, []byte(content), 0644)

	loader := NewManifestLoader(tmpDir)
	manifest, err := loader.LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}

	// Should load existing, not create default
	if manifest.InscriptionVersion != "99" {
		t.Errorf("LoadOrCreate() InscriptionVersion = %q, want '99'", manifest.InscriptionVersion)
	}

	if manifest.GetRegion("custom-region") == nil {
		t.Error("LoadOrCreate() should load existing custom-region")
	}
}

func TestManifestLoader_CreateDefault(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewManifestLoader(tmpDir)

	manifest, err := loader.CreateDefault()
	if err != nil {
		t.Fatalf("CreateDefault() error = %v", err)
	}

	// Check structure
	if manifest.SchemaVersion != DefaultSchemaVersion {
		t.Errorf("CreateDefault() SchemaVersion = %q", manifest.SchemaVersion)
	}

	if manifest.InscriptionVersion != "1" {
		t.Errorf("CreateDefault() InscriptionVersion = %q", manifest.InscriptionVersion)
	}

	if manifest.TemplatePath != DefaultTemplatePath {
		t.Errorf("CreateDefault() TemplatePath = %q", manifest.TemplatePath)
	}

	// Check default regions
	expectedKnossos := []string{
		"execution-mode",
		"agent-routing",
		"commands",
		"platform-infrastructure",
	}

	for _, name := range expectedKnossos {
		region := manifest.GetRegion(name)
		if region == nil {
			t.Errorf("CreateDefault() missing region %q", name)
			continue
		}
		if region.Owner != OwnerKnossos {
			t.Errorf("CreateDefault() region %q owner = %v, want knossos", name, region.Owner)
		}
	}

	// Check regenerate regions
	quickStart := manifest.GetRegion("quick-start")
	if quickStart == nil {
		t.Fatal("CreateDefault() missing 'quick-start' region")
	}
	if quickStart.Owner != OwnerRegenerate {
		t.Errorf("CreateDefault() quick-start owner = %v, want regenerate", quickStart.Owner)
	}
	if quickStart.Source != "ACTIVE_RITE+agents" {
		t.Errorf("CreateDefault() quick-start source = %q", quickStart.Source)
	}
}

func TestManifestLoader_IncrementVersion(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewManifestLoader(tmpDir)

	manifest := &Manifest{
		InscriptionVersion: "5",
	}

	loader.IncrementVersion(manifest)

	if manifest.InscriptionVersion != "6" {
		t.Errorf("IncrementVersion() = %q, want '6'", manifest.InscriptionVersion)
	}

	if manifest.LastSync == nil {
		t.Error("IncrementVersion() should set LastSync")
	}
}

func TestManifestLoader_UpdateRegionHash(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewManifestLoader(tmpDir)

	manifest := &Manifest{
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos},
		},
	}

	content := "test content"
	loader.UpdateRegionHash(manifest, "test", content)

	region := manifest.GetRegion("test")
	if region.Hash == "" {
		t.Error("UpdateRegionHash() should set hash")
	}

	// Hash should be "sha256:" prefix + 64 hex chars = 71 chars
	if len(region.Hash) != 71 {
		t.Errorf("UpdateRegionHash() hash length = %d, want 71", len(region.Hash))
	}

	if region.SyncedAt == nil {
		t.Error("UpdateRegionHash() should set SyncedAt")
	}
}

func TestComputeContentHash(t *testing.T) {
	tests := []struct {
		content string
		want    string
	}{
		{
			content: "hello",
			// SHA256 of "hello" with sha256: prefix
			want: "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			content: "",
			// SHA256 of "" with sha256: prefix
			want: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := ComputeContentHash(tt.content)
			if got != tt.want {
				t.Errorf("ComputeContentHash(%q) = %q, want %q", tt.content, got, tt.want)
			}
		})
	}
}

func TestManifest_GetRegion_SetRegion(t *testing.T) {
	manifest := &Manifest{}

	// GetRegion on empty manifest
	if manifest.GetRegion("test") != nil {
		t.Error("GetRegion() on empty manifest should return nil")
	}

	// SetRegion creates map
	region := &Region{Owner: OwnerKnossos}
	manifest.SetRegion("test", region)

	if got := manifest.GetRegion("test"); got != region {
		t.Error("GetRegion() should return set region")
	}
}

func TestManifest_HasRegion(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"exists": {Owner: OwnerKnossos},
		},
	}

	if !manifest.HasRegion("exists") {
		t.Error("HasRegion('exists') = false, want true")
	}

	if manifest.HasRegion("missing") {
		t.Error("HasRegion('missing') = true, want false")
	}
}

func TestManifest_RegionNames(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"region-a": {Owner: OwnerKnossos},
			"region-b": {Owner: OwnerSatellite},
		},
	}

	names := manifest.RegionNames()
	if len(names) != 2 {
		t.Errorf("RegionNames() got %d, want 2", len(names))
	}

	// Check both names present (order not guaranteed)
	hasA, hasB := false, false
	for _, name := range names {
		if name == "region-a" {
			hasA = true
		}
		if name == "region-b" {
			hasB = true
		}
	}
	if !hasA || !hasB {
		t.Errorf("RegionNames() = %v, want [region-a, region-b]", names)
	}
}

func TestManifest_GetRegionsByOwner(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"knossos-1":    {Owner: OwnerKnossos},
			"knossos-2":    {Owner: OwnerKnossos},
			"satellite-1":  {Owner: OwnerSatellite},
			"regenerate-1": {Owner: OwnerRegenerate, Source: "test"},
		},
	}

	knossos := manifest.GetKnossosRegions()
	if len(knossos) != 2 {
		t.Errorf("GetKnossosRegions() got %d, want 2", len(knossos))
	}

	satellite := manifest.GetSatelliteRegions()
	if len(satellite) != 1 {
		t.Errorf("GetSatelliteRegions() got %d, want 1", len(satellite))
	}

	regenerate := manifest.GetRegenerateRegions()
	if len(regenerate) != 1 {
		t.Errorf("GetRegenerateRegions() got %d, want 1", len(regenerate))
	}
}

func TestManifest_AddRegion(t *testing.T) {
	manifest := &Manifest{
		Regions: make(map[string]*Region),
	}

	// Add valid region
	err := manifest.AddRegion("new-region", &Region{Owner: OwnerKnossos})
	if err != nil {
		t.Errorf("AddRegion() error = %v", err)
	}

	if manifest.GetRegion("new-region") == nil {
		t.Error("AddRegion() did not add region")
	}

	// Add duplicate should fail
	err = manifest.AddRegion("new-region", &Region{Owner: OwnerSatellite})
	if err == nil {
		t.Error("AddRegion() should error on duplicate")
	}

	// Add invalid name should fail
	err = manifest.AddRegion("Invalid Name", &Region{Owner: OwnerKnossos})
	if err == nil {
		t.Error("AddRegion() should error on invalid name")
	}
}

func TestManifest_RemoveRegion(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"keep":   {Owner: OwnerKnossos},
			"remove": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"keep", "remove"},
	}

	manifest.RemoveRegion("remove")

	if manifest.HasRegion("remove") {
		t.Error("RemoveRegion() did not remove region")
	}

	if !manifest.HasRegion("keep") {
		t.Error("RemoveRegion() removed wrong region")
	}

	// Check section_order updated
	if len(manifest.SectionOrder) != 1 || manifest.SectionOrder[0] != "keep" {
		t.Errorf("RemoveRegion() SectionOrder = %v, want [keep]", manifest.SectionOrder)
	}
}

func TestManifest_SetActiveRite(t *testing.T) {
	manifest := &Manifest{
		ActiveRite: "old-rite",
	}

	old := manifest.SetActiveRite("new-rite")

	if old != "old-rite" {
		t.Errorf("SetActiveRite() returned %q, want 'old-rite'", old)
	}

	if manifest.ActiveRite != "new-rite" {
		t.Errorf("SetActiveRite() ActiveRite = %q, want 'new-rite'", manifest.ActiveRite)
	}
}

func TestRegion_ContentChanged(t *testing.T) {
	content := "test content"
	hash := ComputeContentHash(content)

	region := &Region{
		Owner: OwnerKnossos,
		Hash:  hash,
	}

	// Same content should not be changed
	if region.ContentChanged(content) {
		t.Error("ContentChanged() = true for same content, want false")
	}

	// Different content should be changed
	if !region.ContentChanged("different content") {
		t.Error("ContentChanged() = false for different content, want true")
	}

	// Empty hash should always be changed
	region.Hash = ""
	if !region.ContentChanged(content) {
		t.Error("ContentChanged() = false for empty hash, want true")
	}
}

func TestManifest_Clone(t *testing.T) {
	now := time.Now().UTC()
	original := &Manifest{
		SchemaVersion:      "1.0",
		InscriptionVersion: "5",
		LastSync:           &now,
		ActiveRite:         "test-rite",
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos, Hash: "abc"},
		},
		SectionOrder: []string{"test"},
	}

	clone, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone() unexpected error: %v", err)
	}

	// Verify values copied
	if clone.SchemaVersion != original.SchemaVersion {
		t.Error("Clone() SchemaVersion mismatch")
	}

	if clone.ActiveRite != original.ActiveRite {
		t.Error("Clone() ActiveRite mismatch")
	}

	// Verify deep copy (modifying clone doesn't affect original)
	clone.ActiveRite = "modified"
	if original.ActiveRite == "modified" {
		t.Error("Clone() should be deep copy")
	}

	clone.Regions["test"].Hash = "modified"
	if original.Regions["test"].Hash == "modified" {
		t.Error("Clone() regions should be deep copy")
	}
}

func TestMergeManifests(t *testing.T) {
	base := &Manifest{
		SchemaVersion:      "1.0",
		InscriptionVersion: "1",
		ActiveRite:         "base-rite",
		Regions: map[string]*Region{
			"base-region": {Owner: OwnerKnossos},
		},
		SectionOrder: []string{"base-region"},
	}

	overlay := &Manifest{
		InscriptionVersion: "2",
		ActiveRite:         "overlay-rite",
		Regions: map[string]*Region{
			"overlay-region": {Owner: OwnerSatellite},
		},
	}

	merged, err := MergeManifests(base, overlay)
	if err != nil {
		t.Fatalf("MergeManifests() unexpected error: %v", err)
	}

	// Should keep base schema_version
	if merged.SchemaVersion != "1.0" {
		t.Errorf("MergeManifests() SchemaVersion = %q, want '1.0'", merged.SchemaVersion)
	}

	// Should use overlay inscription_version
	if merged.InscriptionVersion != "2" {
		t.Errorf("MergeManifests() InscriptionVersion = %q, want '2'", merged.InscriptionVersion)
	}

	// Should use overlay active_rite
	if merged.ActiveRite != "overlay-rite" {
		t.Errorf("MergeManifests() ActiveRite = %q, want 'overlay-rite'", merged.ActiveRite)
	}

	// Should have both regions
	if merged.GetRegion("base-region") == nil {
		t.Error("MergeManifests() missing base-region")
	}
	if merged.GetRegion("overlay-region") == nil {
		t.Error("MergeManifests() missing overlay-region")
	}
}

func TestMergeManifests_NilInputs(t *testing.T) {
	base := &Manifest{SchemaVersion: "1.0", InscriptionVersion: "1", Regions: make(map[string]*Region)}

	// Nil overlay returns base clone
	result, err := MergeManifests(base, nil)
	if err != nil {
		t.Fatalf("MergeManifests(base, nil) unexpected error: %v", err)
	}
	if result.SchemaVersion != "1.0" {
		t.Error("MergeManifests(base, nil) should return base clone")
	}

	// Nil base returns overlay clone
	result, err = MergeManifests(nil, base)
	if err != nil {
		t.Fatalf("MergeManifests(nil, overlay) unexpected error: %v", err)
	}
	if result.SchemaVersion != "1.0" {
		t.Error("MergeManifests(nil, overlay) should return overlay clone")
	}
}

func TestIsValidSchemaVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0", true},
		{"2.1", true},
		{"10.20", true},
		{"0.1", true},

		{"", false},
		{"1", false},
		{"1.0.0", false},
		{"1.", false},
		{".1", false},
		{"a.b", false},
		{"1.0a", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := isValidSchemaVersion(tt.version); got != tt.want {
				t.Errorf("isValidSchemaVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestDefaultSectionOrder(t *testing.T) {
	order := DefaultSectionOrder()

	if len(order) == 0 {
		t.Error("DefaultSectionOrder() returned empty slice")
	}

	// Check first section is execution-mode
	if order[0] != "execution-mode" {
		t.Errorf("DefaultSectionOrder()[0] = %q, want 'execution-mode'", order[0])
	}

	// Check all entries are valid region names
	for _, name := range order {
		if err := ValidateRegionName(name); err != nil {
			t.Errorf("DefaultSectionOrder() contains invalid name %q: %v", name, err)
		}
	}
}

func TestManifest_ToJSON(t *testing.T) {
	manifest := &Manifest{
		SchemaVersion:      "1.0",
		InscriptionVersion: "1",
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos},
		},
	}

	data, err := manifest.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if !strings.Contains(string(data), `"schema_version":"1.0"`) {
		t.Error("ToJSON() missing schema_version")
	}

	if !strings.Contains(string(data), `"owner":"knossos"`) {
		t.Error("ToJSON() missing region owner")
	}
}

func TestAdoptNewDefaults_AddsMissingRegion(t *testing.T) {
	// Simulate an existing manifest without the "know" region (pre-Layer 2)
	manifest := &Manifest{
		SchemaVersion:      "1.0",
		InscriptionVersion: "42",
		Regions: map[string]*Region{
			"execution-mode":       {Owner: OwnerKnossos},
			"agent-routing":        {Owner: OwnerKnossos},
			"commands":             {Owner: OwnerKnossos},
			"platform-infrastructure": {Owner: OwnerKnossos},
			"quick-start":          {Owner: OwnerRegenerate, Source: "ACTIVE_RITE+agents"},
			"agent-configurations": {Owner: OwnerRegenerate, Source: "agents/*.md"},
			"user-content":         {Owner: OwnerSatellite},
		},
		SectionOrder: []string{
			"execution-mode", "quick-start", "agent-routing",
			"commands", "agent-configurations", "platform-infrastructure", "user-content",
		},
	}

	manifest.AdoptNewDefaults()

	// "know" should now be present
	if !manifest.HasRegion("know") {
		t.Fatal("AdoptNewDefaults() did not add 'know' region")
	}
	knowRegion := manifest.GetRegion("know")
	if knowRegion.Owner != OwnerKnossos {
		t.Errorf("know region owner = %q, want %q", knowRegion.Owner, OwnerKnossos)
	}

	// Section order should include "know"
	found := false
	for _, s := range manifest.SectionOrder {
		if s == "know" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AdoptNewDefaults() did not add 'know' to SectionOrder")
	}

	// Existing regions should be untouched
	if manifest.GetRegion("execution-mode").Owner != OwnerKnossos {
		t.Error("AdoptNewDefaults() modified existing region")
	}
}

func TestAdoptNewDefaults_DoesNotOverwriteExisting(t *testing.T) {
	manifest := &Manifest{
		SchemaVersion:      "1.0",
		InscriptionVersion: "5",
		Regions: map[string]*Region{
			"execution-mode": {Owner: OwnerKnossos, Hash: "sha256:abc123"},
			"know":           {Owner: OwnerKnossos, Hash: "sha256:existing"},
			"user-content":   {Owner: OwnerSatellite},
		},
		SectionOrder: []string{"execution-mode", "know", "user-content"},
	}

	manifest.AdoptNewDefaults()

	// Existing "know" hash should be preserved
	if manifest.GetRegion("know").Hash != "sha256:existing" {
		t.Error("AdoptNewDefaults() overwrote existing region metadata")
	}
}

func TestDeprecatedRegions_NotInDefaultSectionOrder(t *testing.T) {
	defaults := make(map[string]bool)
	for _, name := range DefaultSectionOrder() {
		defaults[name] = true
	}

	for _, name := range DeprecatedRegions() {
		if defaults[name] {
			t.Errorf("DeprecatedRegions() contains %q which is still in DefaultSectionOrder()", name)
		}
	}
}
