package provenance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// validKnossosEntry returns a knossos-owned entry with a valid checksum for testing.
func validKnossosEntry(sourcePath string) *ProvenanceEntry {
	return &ProvenanceEntry{
		Owner:      OwnerKnossos,
		Scope:      ScopeRite,
		SourcePath: sourcePath,
		SourceType: "project",
		Checksum:   "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		LastSynced: time.Now().UTC(),
	}
}

// validUserEntry returns a user-owned entry with a valid checksum for testing.
func validUserEntry() *ProvenanceEntry {
	return &ProvenanceEntry{
		Owner:      OwnerUser,
		Scope:      ScopeRite,
		Checksum:   "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		LastSynced: time.Now().UTC(),
	}
}

// TestMerge_EmptyInputs verifies Merge with nil prev, nil divergence, and empty collector
// produces a valid manifest containing only the collector's entries.
func TestMerge_EmptyInputs(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	collector := NewCollector()
	entry := validKnossosEntry("rites/eco/agents/foo.md")
	collector.Record("agents/foo.md", entry)

	result := Merge(claudeDir, "", "eco", collector, nil, nil, false)

	if result == nil {
		t.Fatal("expected non-nil manifest")
	}
	if result.ActiveRite != "eco" {
		t.Errorf("expected ActiveRite=eco, got %s", result.ActiveRite)
	}
	if result.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected SchemaVersion=%s, got %s", CurrentSchemaVersion, result.SchemaVersion)
	}
	if result.LastSync.IsZero() {
		t.Error("expected non-zero LastSync")
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if _, ok := result.Entries["agents/foo.md"]; !ok {
		t.Error("expected agents/foo.md in entries")
	}
}

// TestMerge_Step0_CarryForwardKnossos verifies that knossos entries from prevManifest
// that still exist on disk are carried forward, while entries for deleted files are dropped.
func TestMerge_Step0_CarryForwardKnossos(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create one file on disk, leave the other absent
	existingFile := filepath.Join(claudeDir, "agents", "existing.md")
	if err := os.MkdirAll(filepath.Dir(existingFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// "agents/deleted.md" intentionally not created

	now := time.Now().UTC()
	prevManifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"agents/existing.md": validKnossosEntry("rites/eco/agents/existing.md"),
			"agents/deleted.md":  validKnossosEntry("rites/eco/agents/deleted.md"),
		},
	}

	// Empty collector — nothing written this sync
	collector := NewCollector()

	result := Merge(claudeDir, "", "eco", collector, nil, prevManifest, false)

	// Only the on-disk entry should carry forward
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry (on-disk only), got %d: %v", len(result.Entries), result.Entries)
	}
	if _, ok := result.Entries["agents/existing.md"]; !ok {
		t.Error("expected agents/existing.md to be carried forward")
	}
	if _, ok := result.Entries["agents/deleted.md"]; ok {
		t.Error("agents/deleted.md should NOT be carried forward (file missing on disk)")
	}
}

// TestMerge_Step0_CarryForwardKnossosDirEntry verifies that directory entries (trailing slash)
// are correctly resolved on disk (trailing slash stripped before stat).
func TestMerge_Step0_CarryForwardKnossosDirEntry(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create the directory on disk
	menaDir := filepath.Join(claudeDir, "commands", "commit")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	prevManifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"commands/commit/": validKnossosEntry("mena/operations/commit/"),
		},
	}

	collector := NewCollector()
	result := Merge(claudeDir, "", "eco", collector, nil, prevManifest, false)

	if _, ok := result.Entries["commands/commit/"]; !ok {
		t.Error("expected commands/commit/ directory entry to be carried forward")
	}
}

// TestMerge_Step1_DivergencePromoted verifies that divergence report promoted entries are
// layered in Step 1, and that entries with empty checksums are skipped.
func TestMerge_Step1_DivergencePromoted(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	divergenceReport := &DivergenceReport{
		Promoted: map[string]*ProvenanceEntry{
			"agents/modified.md": {
				Owner:      OwnerUser,
				Scope:      ScopeRite,
				SourcePath: "rites/eco/agents/modified.md",
				SourceType: "project",
				Checksum:   "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				LastSynced: time.Now().UTC(),
			},
			// Deleted file — empty checksum — should be skipped
			"agents/deleted.md": {
				Owner:      OwnerUser,
				Scope:      ScopeRite,
				SourcePath: "rites/eco/agents/deleted.md",
				SourceType: "project",
				Checksum:   "", // empty — must be skipped
				LastSynced: time.Now().UTC(),
			},
		},
		CarriedForward: map[string]*ProvenanceEntry{
			"custom-agent.md": validUserEntry(),
		},
		Removed: []string{"agents/deleted.md"},
	}

	collector := NewCollector()
	result := Merge(claudeDir, "", "eco", collector, divergenceReport, nil, false)

	// promoted with checksum should be included
	if _, ok := result.Entries["agents/modified.md"]; !ok {
		t.Error("expected agents/modified.md in final entries")
	}
	// promoted with empty checksum should be skipped
	if _, ok := result.Entries["agents/deleted.md"]; ok {
		t.Error("agents/deleted.md with empty checksum should be excluded")
	}
	// carried-forward should be included
	if _, ok := result.Entries["custom-agent.md"]; !ok {
		t.Error("expected custom-agent.md (carried forward) in final entries")
	}
}

// TestMerge_Step2_CollectorLayering verifies that collector entries overwrite Step 0/1 entries,
// EXCEPT entries that were promoted to user ownership in the divergence report.
func TestMerge_Step2_CollectorLayering(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create the knossos-owned file on disk so Step 0 carries it forward
	agentFile := filepath.Join(claudeDir, "agents", "overwritable.md")
	if err := os.MkdirAll(filepath.Dir(agentFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentFile, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	prevManifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"agents/overwritable.md": validKnossosEntry("rites/eco/agents/overwritable.md"),
		},
	}

	// Divergence report promotes "agents/user-protected.md" to user — collector must NOT overwrite it
	userChecksum := "sha256:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	divergenceReport := &DivergenceReport{
		Promoted: map[string]*ProvenanceEntry{
			"agents/user-protected.md": {
				Owner:      OwnerUser,
				Scope:      ScopeRite,
				SourcePath: "rites/eco/agents/user-protected.md",
				SourceType: "project",
				Checksum:   userChecksum,
				LastSynced: now,
			},
		},
		CarriedForward: map[string]*ProvenanceEntry{},
		Removed:        []string{},
	}

	// Collector writes fresh entries for both paths
	newKnossosChecksum := "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	newCollectorEntry := &ProvenanceEntry{
		Owner:      OwnerKnossos,
		Scope:      ScopeRite,
		SourcePath: "rites/eco/agents/overwritable.md",
		SourceType: "project",
		Checksum:   newKnossosChecksum,
		LastSynced: now,
	}
	collectorProtectedEntry := &ProvenanceEntry{
		Owner:      OwnerKnossos, // pipeline thinks it owns this
		Scope:      ScopeRite,
		SourcePath: "rites/eco/agents/user-protected.md",
		SourceType: "project",
		Checksum:   "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		LastSynced: now,
	}

	collector := NewCollector()
	collector.Record("agents/overwritable.md", newCollectorEntry)
	collector.Record("agents/user-protected.md", collectorProtectedEntry)

	result := Merge(claudeDir, "", "eco", collector, divergenceReport, prevManifest, false)

	// "agents/overwritable.md": collector should overwrite Step 0 entry
	overwritable, ok := result.Entries["agents/overwritable.md"]
	if !ok {
		t.Fatal("expected agents/overwritable.md in final entries")
	}
	if overwritable.Checksum != newKnossosChecksum {
		t.Errorf("expected collector checksum %s, got %s", newKnossosChecksum, overwritable.Checksum)
	}

	// "agents/user-protected.md": user-promoted entry must NOT be overwritten
	protected, ok := result.Entries["agents/user-protected.md"]
	if !ok {
		t.Fatal("expected agents/user-protected.md in final entries")
	}
	if protected.Owner != OwnerUser {
		t.Errorf("expected Owner=user, got %s", protected.Owner)
	}
	if protected.Checksum != userChecksum {
		t.Errorf("expected user checksum %s, got %s", userChecksum, protected.Checksum)
	}
}

// TestMerge_Step3_UntrackedPromotion verifies that prev untracked entries not written by the
// collector this sync are promoted to owner:user in Step 3.
func TestMerge_Step3_UntrackedPromotion(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	now := time.Now().UTC()
	untrackedEntry := &ProvenanceEntry{
		Owner:      OwnerUntracked,
		Scope:      ScopeRite,
		Checksum:   "sha256:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
		LastSynced: now,
	}
	prevManifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"legacy-file.md":    untrackedEntry,
			"pipeline-owned.md": validKnossosEntry("rites/eco/pipeline-owned.md"),
		},
	}

	// Collector writes "pipeline-owned.md" but NOT "legacy-file.md"
	collector := NewCollector()
	collectorEntry := &ProvenanceEntry{
		Owner:      OwnerKnossos,
		Scope:      ScopeRite,
		SourcePath: "rites/eco/pipeline-owned.md",
		SourceType: "project",
		Checksum:   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		LastSynced: now,
	}
	collector.Record("pipeline-owned.md", collectorEntry)

	result := Merge(claudeDir, "", "eco", collector, nil, prevManifest, false)

	// "legacy-file.md" was untracked and NOT written this sync → must be promoted to user
	legacy, ok := result.Entries["legacy-file.md"]
	if !ok {
		t.Fatal("expected legacy-file.md in final entries after untracked promotion")
	}
	if legacy.Owner != OwnerUser {
		t.Errorf("expected Owner=user after promotion, got %s", legacy.Owner)
	}
	if legacy.Checksum != untrackedEntry.Checksum {
		t.Errorf("expected checksum preserved: %s, got %s", untrackedEntry.Checksum, legacy.Checksum)
	}

	// "pipeline-owned.md" was written by collector this sync → must NOT be promoted
	pipelineOwned, ok := result.Entries["pipeline-owned.md"]
	if !ok {
		t.Fatal("expected pipeline-owned.md in final entries")
	}
	if pipelineOwned.Owner != OwnerKnossos {
		t.Errorf("expected pipeline entry to remain knossos, got %s", pipelineOwned.Owner)
	}
}

// TestMerge_Step3_UntrackedAlreadyInFinal verifies that if an untracked entry is already in
// the final map (e.g., via divergence CarriedForward in Step 1), Step 3 does not overwrite it.
func TestMerge_Step3_UntrackedAlreadyInFinal(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	now := time.Now().UTC()
	untrackedChecksum := "sha256:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	divergenceCarriedForwardChecksum := "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	prevManifest := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"legacy-file.md": {
				Owner:      OwnerUntracked,
				Scope:      ScopeRite,
				Checksum:   untrackedChecksum,
				LastSynced: now,
			},
		},
	}

	// Divergence report carries forward "legacy-file.md" (e.g., it was in a carried-forward set)
	divergenceReport := &DivergenceReport{
		Promoted: map[string]*ProvenanceEntry{},
		CarriedForward: map[string]*ProvenanceEntry{
			"legacy-file.md": {
				Owner:      OwnerUntracked,
				Scope:      ScopeRite,
				Checksum:   divergenceCarriedForwardChecksum, // different checksum to detect overwrite
				LastSynced: now,
			},
		},
		Removed: []string{},
	}

	collector := NewCollector() // nothing written this sync

	result := Merge(claudeDir, "", "eco", collector, divergenceReport, prevManifest, false)

	legacy, ok := result.Entries["legacy-file.md"]
	if !ok {
		t.Fatal("expected legacy-file.md in final entries")
	}
	// Step 3 must NOT overwrite the entry already in final map from Step 1
	if legacy.Checksum != divergenceCarriedForwardChecksum {
		t.Errorf("Step 3 must not overwrite existing final entry: expected %s, got %s",
			divergenceCarriedForwardChecksum, legacy.Checksum)
	}
}

// TestMerge_Step2_OverwriteDivergedReclaims verifies that when overwriteDiverged is true,
// collector entries reclaim user-promoted entries instead of skipping them.
func TestMerge_Step2_OverwriteDivergedReclaims(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	now := time.Now().UTC()
	userChecksum := "sha256:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

	// Divergence report promotes entry to user-owned
	divergenceReport := &DivergenceReport{
		Promoted: map[string]*ProvenanceEntry{
			"commands/commit/": {
				Owner:      OwnerUser,
				Scope:      ScopeRite,
				SourcePath: "mena/operations/commit",
				SourceType: "project",
				Checksum:   userChecksum,
				LastSynced: now,
			},
		},
		CarriedForward: map[string]*ProvenanceEntry{},
		Removed:        []string{},
	}

	// Collector writes a knossos entry for the same path
	knossosChecksum := "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	collector := NewCollector()
	collector.Record("commands/commit/", &ProvenanceEntry{
		Owner:      OwnerKnossos,
		Scope:      ScopeRite,
		SourcePath: "mena/operations/commit",
		SourceType: "project",
		Checksum:   knossosChecksum,
		LastSynced: now,
	})

	// Without overwriteDiverged: user entry wins
	resultDefault := Merge(claudeDir, "", "eco", collector, divergenceReport, nil, false)
	entry := resultDefault.Entries["commands/commit/"]
	if entry == nil {
		t.Fatal("expected commands/commit/ in default merge result")
	}
	if entry.Owner != OwnerUser {
		t.Errorf("default merge: expected Owner=user, got %s", entry.Owner)
	}

	// With overwriteDiverged: collector reclaims ownership
	resultOverwrite := Merge(claudeDir, "", "eco", collector, divergenceReport, nil, true)
	entry = resultOverwrite.Entries["commands/commit/"]
	if entry == nil {
		t.Fatal("expected commands/commit/ in overwrite merge result")
	}
	if entry.Owner != OwnerKnossos {
		t.Errorf("overwrite merge: expected Owner=knossos (reclaimed), got %s", entry.Owner)
	}
	if entry.Checksum != knossosChecksum {
		t.Errorf("overwrite merge: expected collector checksum %s, got %s", knossosChecksum, entry.Checksum)
	}
}
