package provenance

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
)

// TestCollectorAccumulation tests that Record() accumulates entries correctly.
func TestCollectorAccumulation(t *testing.T) {
	collector := NewCollector()

	now := time.Now().UTC()
	entry1 := &ProvenanceEntry{
		Owner:          OwnerKnossos,
		Scope: ScopeRite,
		SourcePath:     "rites/ecosystem/agents/orchestrator.md",
		SourceType:     "project",
		Checksum:       "sha256:abc123",
		LastSynced:     now,
	}
	entry2 := &ProvenanceEntry{
		Owner:          OwnerKnossos,
		Scope: ScopeRite,
		SourcePath:     "mena/operations/commit/",
		SourceType:     "project",
		Checksum:       "sha256:def456",
		LastSynced:     now,
	}

	collector.Record("agents/orchestrator.md", entry1)
	collector.Record("commands/commit/", entry2)

	entries := collector.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries["agents/orchestrator.md"] != entry1 {
		t.Error("entry1 not found or incorrect")
	}
	if entries["commands/commit/"] != entry2 {
		t.Error("entry2 not found or incorrect")
	}
}

// TestCollectorThreadSafety tests that concurrent Record calls are safe.
func TestCollectorThreadSafety(t *testing.T) {
	collector := NewCollector()
	now := time.Now().UTC()

	const numGoroutines = 10
	const numRecordsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range numRecordsPerGoroutine {
				path := filepath.Join("test", string(rune('a'+id)), string(rune('A'+j)))
				entry := &ProvenanceEntry{
					Owner:      OwnerKnossos,
					Checksum:   "sha256:0000000000000000000000000000000000000000000000000000000000000000",
					LastSynced: now,
				}
				collector.Record(path, entry)
			}
		}(i)
	}

	wg.Wait()

	entries := collector.Entries()
	if len(entries) != numGoroutines*numRecordsPerGoroutine {
		t.Errorf("expected %d entries, got %d", numGoroutines*numRecordsPerGoroutine, len(entries))
	}
}

// TestManifestRoundTrip tests that Save then Load produces identical struct.
func TestManifestRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ManifestFileName)

	now := time.Now().UTC().Truncate(time.Second) // YAML loses sub-second precision

	original := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		ActiveRite:    "ecosystem",
		Entries: map[string]*ProvenanceEntry{
			"agents/orchestrator.md": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "rites/ecosystem/agents/orchestrator.md",
				SourceType:     "project",
				Checksum:       "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced:     now,
			},
			"commands/commit/": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "mena/operations/commit/",
				SourceType:     "project",
				Checksum:       "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				LastSynced:     now,
			},
		},
	}

	// Save
	if err := Save(manifestPath, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Compare
	if loaded.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion mismatch: got %s, want %s", loaded.SchemaVersion, original.SchemaVersion)
	}
	if !loaded.LastSync.Equal(original.LastSync) {
		t.Errorf("LastSync mismatch: got %v, want %v", loaded.LastSync, original.LastSync)
	}
	if loaded.ActiveRite != original.ActiveRite {
		t.Errorf("ActiveRite mismatch: got %s, want %s", loaded.ActiveRite, original.ActiveRite)
	}
	if len(loaded.Entries) != len(original.Entries) {
		t.Fatalf("Entries length mismatch: got %d, want %d", len(loaded.Entries), len(original.Entries))
	}
	for path, origEntry := range original.Entries {
		loadedEntry, ok := loaded.Entries[path]
		if !ok {
			t.Errorf("Entry %s missing from loaded manifest", path)
			continue
		}
		if loadedEntry.Owner != origEntry.Owner {
			t.Errorf("Entry %s Owner mismatch: got %s, want %s", path, loadedEntry.Owner, origEntry.Owner)
		}
		if loadedEntry.Checksum != origEntry.Checksum {
			t.Errorf("Entry %s Checksum mismatch: got %s, want %s", path, loadedEntry.Checksum, origEntry.Checksum)
		}
	}
}

// TestLoadOrBootstrapMissingFile tests that LoadOrBootstrap returns empty manifest when file missing.
func TestLoadOrBootstrapMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ManifestFileName)

	manifest, err := LoadOrBootstrap(manifestPath)
	if err != nil {
		t.Fatalf("LoadOrBootstrap failed: %v", err)
	}

	if manifest.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected SchemaVersion %s, got %s", CurrentSchemaVersion, manifest.SchemaVersion)
	}
	if !manifest.LastSync.IsZero() {
		t.Errorf("expected zero LastSync, got %v", manifest.LastSync)
	}
	if manifest.ActiveRite != "" {
		t.Errorf("expected empty ActiveRite, got %s", manifest.ActiveRite)
	}
	if manifest.Entries == nil {
		t.Error("expected non-nil Entries map")
	}
	if len(manifest.Entries) != 0 {
		t.Errorf("expected empty Entries map, got %d entries", len(manifest.Entries))
	}
}

// TestLoadOrBootstrapExistingFile tests that LoadOrBootstrap returns parsed manifest when file exists.
func TestLoadOrBootstrapExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ManifestFileName)

	now := time.Now().UTC().Truncate(time.Second)
	original := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		ActiveRite:    "ecosystem",
		Entries: map[string]*ProvenanceEntry{
			"agents/orchestrator.md": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "rites/ecosystem/agents/orchestrator.md",
				SourceType:     "project",
				Checksum:       "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced:     now,
			},
		},
	}

	// Save first
	if err := Save(manifestPath, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load via LoadOrBootstrap
	loaded, err := LoadOrBootstrap(manifestPath)
	if err != nil {
		t.Fatalf("LoadOrBootstrap failed: %v", err)
	}

	if loaded.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion mismatch: got %s, want %s", loaded.SchemaVersion, original.SchemaVersion)
	}
	if len(loaded.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(loaded.Entries))
	}
}

// TestDivergenceChecksumMatch tests that unchanged knossos files stay knossos.
func TestDivergenceChecksumMatch(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file with known content
	testFilePath := filepath.Join(claudeDir, "test.md")
	testContent := []byte("test content")
	if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Compute checksum
	hash := checksum.Bytes(testContent)

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"test.md": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "test/source.md",
				SourceType:     "project",
				Checksum:       hash,
				LastSynced:     now,
			},
		},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// Checksum matches: file should NOT be promoted
	if len(report.Promoted) != 0 {
		t.Errorf("expected no promoted entries, got %d", len(report.Promoted))
	}
	if len(report.CarriedForward) != 0 {
		t.Errorf("expected no carried forward entries, got %d", len(report.CarriedForward))
	}
}

// TestDivergenceChecksumMismatch tests that modified knossos files are promoted to user.
func TestDivergenceChecksumMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file with modified content
	testFilePath := filepath.Join(claudeDir, "test.md")
	testContent := []byte("modified content")
	if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Original checksum for different content
	originalHash := checksum.Bytes([]byte("original content"))

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"test.md": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "test/source.md",
				SourceType:     "project",
				Checksum:       originalHash,
				LastSynced:     now,
			},
		},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// Checksum mismatch: file should be promoted to user
	if len(report.Promoted) != 1 {
		t.Fatalf("expected 1 promoted entry, got %d", len(report.Promoted))
	}
	promoted, ok := report.Promoted["test.md"]
	if !ok {
		t.Fatal("test.md not in promoted entries")
	}
	if promoted.Owner != OwnerUser {
		t.Errorf("expected Owner=user, got %s", promoted.Owner)
	}
	if promoted.SourcePath != "test/source.md" {
		t.Errorf("expected SourcePath retained, got %s", promoted.SourcePath)
	}
}

// TestDivergenceUserOwnedCarriedForward tests that user-owned entries are carried forward unchanged.
func TestDivergenceUserOwnedCarriedForward(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"custom-agent.md": {
				Owner:      OwnerUser,
				Checksum:   "sha256:userchecksum1234567890abcdef1234567890abcdef1234567890abcdef123456",
				LastSynced: now,
			},
		},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// User-owned files are carried forward, not promoted
	if len(report.Promoted) != 0 {
		t.Errorf("expected no promoted entries, got %d", len(report.Promoted))
	}
	if len(report.CarriedForward) != 1 {
		t.Fatalf("expected 1 carried forward entry, got %d", len(report.CarriedForward))
	}
	carried, ok := report.CarriedForward["custom-agent.md"]
	if !ok {
		t.Fatal("custom-agent.md not in carried forward entries")
	}
	if carried.Owner != OwnerUser {
		t.Errorf("expected Owner=user, got %s", carried.Owner)
	}
}

// TestDivergenceUnknownOwnedCarriedForward tests that unknown-owned entries are carried forward.
func TestDivergenceUnknownOwnedCarriedForward(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"legacy-file.md": {
				Owner:      OwnerUntracked,
				Checksum:   "sha256:unknownchecksum1234567890abcdef1234567890abcdef1234567890abcdef12",
				LastSynced: now,
			},
		},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// Unknown-owned files are carried forward
	if len(report.CarriedForward) != 1 {
		t.Fatalf("expected 1 carried forward entry, got %d", len(report.CarriedForward))
	}
	carried, ok := report.CarriedForward["legacy-file.md"]
	if !ok {
		t.Fatal("legacy-file.md not in carried forward entries")
	}
	if carried.Owner != OwnerUntracked {
		t.Errorf("expected Owner=unknown, got %s", carried.Owner)
	}
}

// TestDivergenceFileMissing tests that deleted knossos files are promoted to user with empty checksum.
func TestDivergenceFileMissing(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"deleted.md": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "test/deleted.md",
				SourceType:     "project",
				Checksum:       "sha256:deadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678",
				LastSynced:     now,
			},
		},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// File missing: promoted to user with empty checksum
	if len(report.Promoted) != 1 {
		t.Fatalf("expected 1 promoted entry, got %d", len(report.Promoted))
	}
	promoted, ok := report.Promoted["deleted.md"]
	if !ok {
		t.Fatal("deleted.md not in promoted entries")
	}
	if promoted.Owner != OwnerUser {
		t.Errorf("expected Owner=user, got %s", promoted.Owner)
	}
	if promoted.Checksum != "" {
		t.Errorf("expected empty Checksum, got %s", promoted.Checksum)
	}
	if len(report.Removed) != 1 || report.Removed[0] != "deleted.md" {
		t.Errorf("expected deleted.md in Removed list")
	}
}

// TestDivergenceNewFile tests that new files in current sync are NOT in divergence report.
func TestDivergenceNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries:       map[string]*ProvenanceEntry{},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// No existing entries: no divergence
	if len(report.Promoted) != 0 {
		t.Errorf("expected no promoted entries, got %d", len(report.Promoted))
	}
	if len(report.CarriedForward) != 0 {
		t.Errorf("expected no carried forward entries, got %d", len(report.CarriedForward))
	}
}

// TestMenaDirectoryChecksum tests that directory-level checksums work correctly.
func TestMenaDirectoryChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	menaDir := filepath.Join(claudeDir, "commands", "commit")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files in mena directory
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.md"), []byte("index content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(menaDir, "support.md"), []byte("support content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Compute directory checksum
	dirHash, err := checksum.Dir(menaDir)
	if err != nil {
		t.Fatalf("checksum.Dir failed: %v", err)
	}

	now := time.Now().UTC()
	previous := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"commands/commit/": {
				Owner:          OwnerKnossos,
				Scope: ScopeRite,
				SourcePath:     "mena/operations/commit/",
				SourceType:     "project",
				Checksum:       dirHash,
				LastSynced:     now,
			},
		},
	}

	report, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// Directory unchanged: no promotion
	if len(report.Promoted) != 0 {
		t.Errorf("expected no promoted entries, got %d", len(report.Promoted))
	}

	// Modify a file in the directory
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.md"), []byte("modified index"), 0644); err != nil {
		t.Fatal(err)
	}

	report2, err := DetectDivergence(previous, nil, claudeDir)
	if err != nil {
		t.Fatalf("DetectDivergence failed: %v", err)
	}

	// Directory modified: should be promoted
	if len(report2.Promoted) != 1 {
		t.Fatalf("expected 1 promoted entry, got %d", len(report2.Promoted))
	}
	promoted, ok := report2.Promoted["commands/commit/"]
	if !ok {
		t.Fatal("commands/commit/ not in promoted entries")
	}
	if promoted.Owner != OwnerUser {
		t.Errorf("expected Owner=user, got %s", promoted.Owner)
	}
}

// TestValidationMissingSchemaVersion tests that manifest without schema_version fails validation.
func TestValidationMissingSchemaVersion(t *testing.T) {
	manifest := &ProvenanceManifest{
		SchemaVersion: "",
		LastSync:      time.Now().UTC(),
		Entries:       make(map[string]*ProvenanceEntry),
	}

	err := validateManifest(manifest)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestValidationInvalidChecksum tests that invalid checksum format fails validation.
func TestValidationInvalidChecksum(t *testing.T) {
	now := time.Now().UTC()
	manifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"test.md": {
				Owner:      OwnerKnossos,
				Checksum:   "invalid-checksum",
				LastSynced: now,
			},
		},
	}

	err := validateManifest(manifest)
	if err == nil {
		t.Fatal("expected validation error for invalid checksum, got nil")
	}
}

// TestValidationKnossosRequiresSource tests that knossos entries require source_path and source_type.
func TestValidationKnossosRequiresSource(t *testing.T) {
	now := time.Now().UTC()
	manifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"test.md": {
				Owner:      OwnerKnossos,
				Checksum:   "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced: now,
				// Missing SourcePath and SourceType
			},
		},
	}

	err := validateManifest(manifest)
	if err == nil {
		t.Fatal("expected validation error for missing source_path, got nil")
	}
}

// TestScopeTypeIsValid tests that ScopeType.IsValid() returns correct values.
func TestScopeTypeIsValid(t *testing.T) {
	tests := []struct {
		scope ScopeType
		valid bool
	}{
		{ScopeRite, true},
		{ScopeUser, true},
		{ScopeType("invalid"), false},
		{ScopeType(""), false},
	}

	for _, tt := range tests {
		got := tt.scope.IsValid()
		if got != tt.valid {
			t.Errorf("ScopeType(%q).IsValid() = %v, want %v", tt.scope, got, tt.valid)
		}
	}
}

// TestOwnerTypeUnknownInvalid tests that the old "unknown" owner value is rejected.
func TestOwnerTypeUnknownInvalid(t *testing.T) {
	oldUnknown := OwnerType("unknown")
	if oldUnknown.IsValid() {
		t.Error("OwnerType(\"unknown\").IsValid() should return false")
	}
}

// TestMigrateV1ToV2 tests that v1.0 manifests are correctly migrated to v2.0.
func TestMigrateV1ToV2(t *testing.T) {
	now := time.Now().UTC()
	v1Manifest := &ProvenanceManifest{
		SchemaVersion: "1.0",
		LastSync:      now,
		ActiveRite:    "ecosystem",
		Entries: map[string]*ProvenanceEntry{
			"agents/orchestrator.md": {
				Owner:      OwnerKnossos,
				SourcePath: "rites/ecosystem/agents/orchestrator.md",
				SourceType: "project",
				Checksum:   "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced: now,
				// Scope is empty (v1.0 format)
			},
			"custom-agent.md": {
				Owner:      OwnerType("unknown"), // v1.0 used "unknown"
				Checksum:   "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				LastSynced: now,
			},
		},
	}

	// Migrate
	migrateV1ToV2(v1Manifest)

	// Check schema version updated
	if v1Manifest.SchemaVersion != "2.0" {
		t.Errorf("expected SchemaVersion 2.0, got %s", v1Manifest.SchemaVersion)
	}

	// Check Scope set to ScopeRite
	if v1Manifest.Entries["agents/orchestrator.md"].Scope != ScopeRite {
		t.Errorf("expected Scope=rite, got %s", v1Manifest.Entries["agents/orchestrator.md"].Scope)
	}

	// Check "unknown" owner converted to "untracked"
	if v1Manifest.Entries["custom-agent.md"].Owner != OwnerUntracked {
		t.Errorf("expected Owner=untracked, got %s", v1Manifest.Entries["custom-agent.md"].Owner)
	}
}

// TestValidationMissingScope tests that entries without Scope fail validation.
func TestValidationMissingScope(t *testing.T) {
	now := time.Now().UTC()
	manifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"test.md": {
				Owner:      OwnerKnossos,
				SourcePath: "test/source.md",
				SourceType: "project",
				Checksum:   "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced: now,
				// Missing Scope
			},
		},
	}

	err := validateManifest(manifest)
	if err == nil {
		t.Fatal("expected validation error for missing scope, got nil")
	}
}
