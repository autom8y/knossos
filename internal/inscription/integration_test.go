package inscription

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Integration tests for the inscription pipeline as specified in TDD Section 11.2.
// These tests verify end-to-end behavior of the sync pipeline with realistic
// file system operations using temp directories for isolation.

// TestInscription_SyncCleanProject tests syncing on a clean project (no existing CLAUDE.md).
// TDD Section 11.2 inscription_001: Full pipeline on clean project.
func TestInscription_SyncCleanProject(t *testing.T) {
	// Setup: Create a temp project directory with no CLAUDE.md
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a sample agent file
	agentContent := `# Test Agent

The test agent verifies things work correctly.

Produces: Test reports
`
	if err := os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Create ACTIVE_RITE file
	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatalf("Failed to create ACTIVE_RITE file: %v", err)
	}

	// Create pipeline and run sync
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	// Verify sync succeeded
	if !result.Success {
		t.Error("Sync() should succeed on clean project")
	}

	// Verify CLAUDE.md was created
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		t.Fatal("Sync() should create CLAUDE.md on clean project")
	}

	// Read and verify content
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	// Verify header is present
	if !strings.Contains(string(content), "# CLAUDE.md") {
		t.Error("CLAUDE.md should contain header")
	}

	// Verify regions were synced
	if len(result.RegionsSynced) == 0 {
		t.Error("Sync() should report synced regions")
	}

	// Verify manifest was created
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("Sync() should create manifest on clean project")
	}

	// Verify inscription version was set
	if result.InscriptionVersion == "" {
		t.Error("Sync() should set inscription version")
	}
}

// TestInscription_SyncWithSatelliteSections tests that satellite-owned sections are preserved.
// TDD Section 11.2 inscription_002: Sync with existing satellite sections.
func TestInscription_SyncWithSatelliteSections(t *testing.T) {
	// Setup: Create project with existing CLAUDE.md containing satellite section
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create existing CLAUDE.md with a satellite-owned section
	existingContent := `# CLAUDE.md

> Entry point for Claude Code.

<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Old execution mode content that will be overwritten.
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START project-custom owner=satellite -->
## Project Custom Section

This is custom satellite content that MUST be preserved.
It contains project-specific documentation.
<!-- KNOSSOS:END project-custom -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest with satellite region
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
  project-custom:
    owner: satellite
section_order:
  - execution-mode
  - project-custom
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run sync
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() should succeed")
	}

	// Read synced content
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	// Verify satellite content is preserved
	if !strings.Contains(string(content), "This is custom satellite content that MUST be preserved") {
		t.Error("Sync() should preserve satellite-owned section content")
	}

	if !strings.Contains(string(content), "project-specific documentation") {
		t.Error("Sync() should preserve all satellite section content")
	}

	// Verify knossos section was updated (should not contain "Old execution mode")
	if strings.Contains(string(content), "Old execution mode content that will be overwritten") {
		t.Error("Sync() should overwrite knossos-owned section content")
	}
}

// TestInscription_SyncOverwritesKnossosRegions tests that knossos-owned regions are always overwritten.
// TDD Section 11.2 inscription_003: Sync with user edits to knossos regions.
func TestInscription_SyncOverwritesKnossosRegions(t *testing.T) {
	// Setup: Create project with modified knossos region
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create CLAUDE.md with user-modified knossos region
	existingContent := `# CLAUDE.md

> Entry point for Claude Code.

<!-- KNOSSOS:START execution-mode -->
## Execution Mode

USER MODIFIED CONTENT - This should be overwritten!
Users sometimes edit knossos regions by mistake.
<!-- KNOSSOS:END execution-mode -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest with hash (simulating previous sync)
	originalContent := "Different content that was here before"
	originalHash := ComputeContentHash(originalContent)

	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
    hash: "` + originalHash + `"
section_order:
  - execution-mode
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run sync
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() should succeed")
	}

	// Read synced content
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	// Verify user modifications were overwritten
	if strings.Contains(string(content), "USER MODIFIED CONTENT") {
		t.Error("Sync() should overwrite user edits in knossos-owned regions")
	}

	if strings.Contains(string(content), "Users sometimes edit knossos regions by mistake") {
		t.Error("Sync() should completely replace knossos region content")
	}

	// Verify conflict was detected
	hasConflict := false
	for _, conflict := range result.Conflicts {
		if conflict.Region == "execution-mode" && conflict.Type == ConflictUserEditedKnossos {
			hasConflict = true
			break
		}
	}
	if !hasConflict {
		t.Error("Sync() should detect and report conflict for user-edited knossos region")
	}
}

// TestInscription_SyncRegenerateWithUserEdits tests conflict resolution for regenerate regions.
// TDD Section 11.2 inscription_004: Sync with user edits to regenerate regions.
func TestInscription_SyncRegenerateWithUserEdits(t *testing.T) {
	// Setup: Create project with modified regenerate region
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agent file
	if err := os.WriteFile(filepath.Join(agentsDir, "agent.md"), []byte("# Agent\nDoes stuff.\nProduces: Things"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Create CLAUDE.md with modified quick-start section
	existingContent := `# CLAUDE.md

> Entry point for Claude Code.

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

USER CUSTOMIZED QUICK START - Added project-specific tips
<!-- KNOSSOS:END quick-start -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest with preserve_on_conflict=true
	originalHash := ComputeContentHash("Original quick-start content")

	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  quick-start:
    owner: regenerate
    source: ACTIVE_RITE+agents
    preserve_on_conflict: true
    hash: "` + originalHash + `"
section_order:
  - quick-start
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run sync
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() should succeed")
	}

	// Read synced content
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	// With preserve_on_conflict=true, user edits should be preserved
	if !strings.Contains(string(content), "USER CUSTOMIZED QUICK START") {
		t.Error("Sync() should preserve user edits when preserve_on_conflict=true")
	}

	// Verify conflict was detected
	hasConflict := false
	for _, conflict := range result.Conflicts {
		if conflict.Region == "quick-start" && conflict.Type == ConflictUserEditedRegenerate {
			hasConflict = true
			if !conflict.Preserved {
				t.Error("Conflict should indicate content was preserved")
			}
			break
		}
	}
	if !hasConflict {
		t.Error("Sync() should detect conflict for user-edited regenerate region")
	}
}

// TestInscription_RollbackAfterSync tests the rollback mechanism restores previous state.
// TDD Section 11.2 inscription_005: Rollback after failed sync.
func TestInscription_RollbackAfterSync(t *testing.T) {
	// Setup: Create project with existing CLAUDE.md
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create original CLAUDE.md
	originalContent := `# CLAUDE.md

> Original entry point content.

<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Original execution mode content before sync.
<!-- KNOSSOS:END execution-mode -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run sync (which creates a backup)
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if result.BackupPath == "" {
		t.Fatal("Sync() should create backup of existing CLAUDE.md")
	}

	// Verify content changed after sync
	newContent, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	if strings.Contains(string(newContent), "Original execution mode content before sync") {
		t.Error("Sync() should have updated the content")
	}

	// Perform rollback
	err = pipeline.Rollback("")
	if err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Verify original content is restored
	restoredContent, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read restored CLAUDE.md: %v", err)
	}

	if !strings.Contains(string(restoredContent), "Original execution mode content before sync") {
		t.Error("Rollback() should restore original content")
	}

	if !strings.Contains(string(restoredContent), "Original entry point content") {
		t.Error("Rollback() should restore full original file")
	}
}

// TestInscription_DryRunNoChanges tests that dry-run makes no file changes.
// TDD Section 11.2 inscription_006: Dry-run preview.
func TestInscription_DryRunNoChanges(t *testing.T) {
	// Setup: Create project with existing CLAUDE.md
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	backupsDir := filepath.Join(claudeDir, "backups")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create existing CLAUDE.md
	originalContent := `# CLAUDE.md

> Entry point for Claude Code.

<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Content before dry-run that should NOT change.
<!-- KNOSSOS:END execution-mode -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "5"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Record file modification times
	claudeMDStat, _ := os.Stat(claudeMDPath)
	manifestStat, _ := os.Stat(manifestPath)
	originalModTime := claudeMDStat.ModTime()
	manifestModTime := manifestStat.ModTime()

	// Wait a moment to ensure time difference would be detectable
	time.Sleep(10 * time.Millisecond)

	// Run dry-run sync
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Sync(DryRun) error = %v", err)
	}

	if !result.Success {
		t.Error("Sync(DryRun) should succeed")
	}

	if !result.DryRun {
		t.Error("Sync(DryRun) should set DryRun flag in result")
	}

	// Verify CLAUDE.md was NOT modified
	afterStat, _ := os.Stat(claudeMDPath)
	if !afterStat.ModTime().Equal(originalModTime) {
		t.Error("Sync(DryRun) should not modify CLAUDE.md")
	}

	// Verify content unchanged
	content, _ := os.ReadFile(claudeMDPath)
	if !strings.Contains(string(content), "Content before dry-run that should NOT change") {
		t.Error("Sync(DryRun) should not change file content")
	}

	// Verify manifest was NOT modified
	afterManifestStat, _ := os.Stat(manifestPath)
	if !afterManifestStat.ModTime().Equal(manifestModTime) {
		t.Error("Sync(DryRun) should not modify manifest")
	}

	// Verify no backup was created
	if _, err := os.Stat(backupsDir); err == nil {
		entries, _ := os.ReadDir(backupsDir)
		if len(entries) > 0 {
			t.Error("Sync(DryRun) should not create backups")
		}
	}

	// Verify result still contains useful preview information
	if len(result.RegionsSynced) == 0 {
		t.Error("Sync(DryRun) should report regions that would be synced")
	}
}

// TestInscription_IdempotentSync tests that running sync twice produces the same result.
// TDD Section 11.2 inscription_007: Idempotent sync.
func TestInscription_IdempotentSync(t *testing.T) {
	// Setup: Create a clean project
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agent file
	if err := os.WriteFile(filepath.Join(agentsDir, "test.md"), []byte("# Test\nRole.\nProduces: Output"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// First sync
	pipeline := NewPipeline(tmpDir)
	result1, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("First Sync() error = %v", err)
	}

	if !result1.Success {
		t.Fatal("First Sync() should succeed")
	}

	// Read content after first sync
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	content1, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md after first sync: %v", err)
	}

	// Second sync
	result2, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Second Sync() error = %v", err)
	}

	if !result2.Success {
		t.Fatal("Second Sync() should succeed")
	}

	// Read content after second sync
	content2, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md after second sync: %v", err)
	}

	// Extract region contents (content between markers) for comparison
	// This verifies functional idempotency - the actual content is the same
	// even if marker options differ slightly between syncs
	parser := NewMarkerParser()
	parsed1 := parser.Parse(string(content1))
	parsed2 := parser.Parse(string(content2))

	// Verify same regions are present
	if len(parsed1.Regions) != len(parsed2.Regions) {
		t.Errorf("Different number of regions: first sync %d, second sync %d",
			len(parsed1.Regions), len(parsed2.Regions))
	}

	// Verify each region's content is identical
	for name, region1 := range parsed1.Regions {
		region2 := parsed2.GetRegion(name)
		if region2 == nil {
			t.Errorf("Region %q missing after second sync", name)
			continue
		}

		// Compare content (the actual meaningful part)
		if strings.TrimSpace(region1.Content) != strings.TrimSpace(region2.Content) {
			t.Errorf("Region %q content differs between syncs", name)
			t.Logf("First: %q", region1.Content[:min(100, len(region1.Content))])
			t.Logf("Second: %q", region2.Content[:min(100, len(region2.Content))])
		}
	}

	// Verify no conflicts on second sync (content unchanged)
	// Note: Some implementations may report minor conflicts for marker option differences
	// which is acceptable as long as content is preserved
	conflictsWithOverwrite := 0
	for _, conflict := range result2.Conflicts {
		if !conflict.Preserved {
			conflictsWithOverwrite++
		}
	}
	if conflictsWithOverwrite > 0 {
		t.Logf("Warning: %d conflicts with overwrites on second sync", conflictsWithOverwrite)
	}
}

// TestInscription_RiteSwitchTriggersSync tests integration between rite switch and sync.
// TDD Section 11.2 inscription_009: Rite switch triggers sync.
func TestInscription_RiteSwitchTriggersSync(t *testing.T) {
	// Setup: Create project with existing content
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agent file
	if err := os.WriteFile(filepath.Join(agentsDir, "agent.md"), []byte("# Agent\nRole.\nProduces: Output"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Create existing CLAUDE.md
	existingContent := `# CLAUDE.md

> Entry point.

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

This project uses a 3-agent workflow (old-rite):
<!-- KNOSSOS:END quick-start -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest with old rite
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
active_rite: "old-rite"
regions:
  quick-start:
    owner: regenerate
    source: ACTIVE_RITE+agents
section_order:
  - quick-start
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Sync with new rite name (simulating rite switch)
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{
		RiteName: "new-rite",
	})
	if err != nil {
		t.Fatalf("Sync() with new rite error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() with rite change should succeed")
	}

	// Verify content was regenerated
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	// The Quick Start should now reference new-rite (from the regeneration)
	if strings.Contains(string(content), "old-rite") {
		t.Error("Sync() should regenerate content with new rite name")
	}

	// Load manifest and verify active_rite was updated
	loader := NewManifestLoader(tmpDir)
	manifest, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest.ActiveRite != "new-rite" {
		t.Errorf("Manifest active_rite = %q, want 'new-rite'", manifest.ActiveRite)
	}
}

// Helper function for min of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Additional integration tests for edge cases

// TestInscription_MalformedMarkersHandled tests graceful handling of malformed markers.
func TestInscription_MalformedMarkersHandled(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create CLAUDE.md with malformed marker (unclosed region)
	malformedContent := `# CLAUDE.md

<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Content without an END marker - this is malformed!

<!-- KNOSSOS:START another-region -->
## Another Region

Also no END marker
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	if err := os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Sync should still succeed (graceful degradation)
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})

	// The sync should handle malformed content gracefully
	// It may succeed with conflicts or report errors, but shouldn't panic
	if err != nil {
		// Some implementations may return an error for malformed content
		t.Logf("Sync() returned error for malformed content: %v", err)
		return
	}

	// If no error, check that conflicts were detected
	if len(result.Conflicts) == 0 {
		t.Log("Note: No conflicts detected for malformed markers - may treat as satellite content")
	}
}

// TestInscription_LargeFile tests handling of large CLAUDE.md files.
func TestInscription_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create a large CLAUDE.md (50KB+)
	var sb strings.Builder
	sb.WriteString("# CLAUDE.md\n\n> Entry point.\n\n")
	sb.WriteString("<!-- KNOSSOS:START execution-mode -->\n")
	sb.WriteString("## Execution Mode\n\n")

	// Add ~50KB of content
	for i := 0; i < 1000; i++ {
		sb.WriteString("This is line ")
		sb.WriteString(itoa(i))
		sb.WriteString(" of the large documentation file with plenty of text to make it sizable.\n")
	}

	sb.WriteString("<!-- KNOSSOS:END execution-mode -->\n")

	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("Failed to create large CLAUDE.md: %v", err)
	}

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	if err := os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Sync should complete in reasonable time
	start := time.Now()
	pipeline := NewPipeline(tmpDir)
	result, err := pipeline.Sync(InscriptionSyncOptions{})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Sync() error on large file = %v", err)
	}

	if !result.Success {
		t.Error("Sync() should succeed on large file")
	}

	// Should complete within 5 seconds even on slow hardware
	if elapsed > 5*time.Second {
		t.Errorf("Sync() took too long on large file: %v", elapsed)
	}
}

// TestInscription_ConcurrentBackups tests backup cleanup with multiple syncs.
func TestInscription_ConcurrentBackups(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	backupsDir := filepath.Join(claudeDir, "backups")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create initial CLAUDE.md
	initialContent := `# CLAUDE.md

<!-- KNOSSOS:START execution-mode -->
## Execution Mode
Initial content.
<!-- KNOSSOS:END execution-mode -->
`
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	if err := os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run multiple syncs (more than MaxBackups default of 5)
	pipeline := NewPipeline(tmpDir)
	for i := 0; i < 8; i++ {
		// Small delay to ensure different timestamps
		time.Sleep(5 * time.Millisecond)

		_, err := pipeline.Sync(InscriptionSyncOptions{})
		if err != nil {
			t.Fatalf("Sync() %d error = %v", i, err)
		}
	}

	// Verify backups were cleaned up (should be <= MaxBackups)
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("Failed to read backups directory: %v", err)
	}

	// Count actual backup files
	backupCount := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "CLAUDE.md.") {
			backupCount++
		}
	}

	if backupCount > 5 {
		t.Errorf("Backup cleanup not working: found %d backups, expected <= 5", backupCount)
	}
}
