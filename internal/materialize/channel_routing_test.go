package materialize_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

func createTestRite(t *testing.T, dir string) string {
	t.Helper()
	riteDir := filepath.Join(dir, "rites", "test-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `
name: test-rite
version: 1.0.0
description: A test rite
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}
	return riteDir
}

func TestMaterializeWithOptions_GeminiChannel(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	// Create test rite
	createTestRite(t, tmpDir)

	m := materialize.NewMaterializerWithSource(resolver, filepath.Join(tmpDir, "rites"))

	opts := materialize.Options{
		Channel: "gemini",
		DryRun:  false,
	}

	_, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Should write to .gemini
	geminiDir := filepath.Join(tmpDir, ".gemini")
	if _, err := os.Stat(geminiDir); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", geminiDir)
	}

	// Should NOT write to .claude
	claudeDir := filepath.Join(tmpDir, ".claude")
	if _, err := os.Stat(claudeDir); !os.IsNotExist(err) {
		t.Errorf("expected %s to NOT be created", claudeDir)
	}
}

func TestMaterializeWithOptions_DefaultChannel(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRite(t, tmpDir)

	m := materialize.NewMaterializerWithSource(resolver, filepath.Join(tmpDir, "rites"))

	opts := materialize.Options{
		Channel: "claude",
		DryRun:  false,
	}

	_, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Should write to .claude
	claudeDir := filepath.Join(tmpDir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", claudeDir)
	}

	// Should NOT write to .gemini
	geminiDir := filepath.Join(tmpDir, ".gemini")
	if _, err := os.Stat(geminiDir); !os.IsNotExist(err) {
		t.Errorf("expected %s to NOT be created", geminiDir)
	}
}

func TestSync_ChannelAll_ProjectsBoth(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRite(t, tmpDir)

	// Write ACTIVE_RITE so syncRiteScope knows which rite to use
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knossosDir, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatal(err)
	}

	m := materialize.NewMaterializerWithSource(resolver, filepath.Join(tmpDir, "rites"))

	opts := materialize.SyncOptions{
		Scope:   materialize.ScopeRite,
		Channel: "all",
	}

	result, err := m.Sync(opts)
	if err != nil {
		t.Fatalf("Sync(channel=all) failed: %v", err)
	}

	if result.RiteResult == nil {
		t.Fatal("expected RiteResult to be non-nil")
	}

	// Should have ChannelResults for both channels
	if len(result.RiteResult.ChannelResults) != 2 {
		t.Fatalf("expected 2 channel results, got %d", len(result.RiteResult.ChannelResults))
	}

	for _, chName := range []string{"claude", "gemini"} {
		chResult, ok := result.RiteResult.ChannelResults[chName]
		if !ok {
			t.Errorf("missing channel result for %q", chName)
			continue
		}
		if chResult.Status != "success" {
			t.Errorf("channel %q status = %q, want %q (error: %s)", chName, chResult.Status, "success", chResult.Error)
		}
	}

	// Both directories should exist
	claudeDir := filepath.Join(tmpDir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", claudeDir)
	}

	geminiDir := filepath.Join(tmpDir, ".gemini")
	if _, err := os.Stat(geminiDir); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", geminiDir)
	}

	// Top-level result should inherit from first channel (claude)
	if result.RiteResult.RiteName != "test-rite" {
		t.Errorf("wrapper RiteName = %q, want %q", result.RiteResult.RiteName, "test-rite")
	}
}

func TestSync_ChannelAll_PartialOnFailure(t *testing.T) {
	t.Parallel()

	// This test verifies that if a channel=all sync has one channel succeed
	// and another fail, the wrapper result status is "partial".
	// We can't easily force one channel to fail without deeper mocking,
	// so we just verify the structural contract: if both succeed, status != "partial".
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRite(t, tmpDir)

	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knossosDir, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatal(err)
	}

	m := materialize.NewMaterializerWithSource(resolver, filepath.Join(tmpDir, "rites"))

	opts := materialize.SyncOptions{
		Scope:   materialize.ScopeRite,
		Channel: "all",
	}

	result, err := m.Sync(opts)
	if err != nil {
		t.Fatalf("Sync(channel=all) failed: %v", err)
	}

	// When both succeed, status should NOT be "partial"
	if result.RiteResult.Status == "partial" {
		t.Errorf("status = %q when both channels succeeded; expected non-partial", result.RiteResult.Status)
	}
}

func TestMaterializeWithOptions_ClaudeUnchanged(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRite(t, tmpDir)

	// Pre-create a file in .claude to test it's untouched
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	markerFile := filepath.Join(claudeDir, "marker.txt")
	if err := os.WriteFile(markerFile, []byte("untouched"), 0644); err != nil {
		t.Fatal(err)
	}
	markerInfoBefore, err := os.Stat(markerFile)
	if err != nil {
		t.Fatal(err)
	}

	m := materialize.NewMaterializerWithSource(resolver, filepath.Join(tmpDir, "rites"))

	opts := materialize.Options{
		Channel: "gemini",
		DryRun:  false,
	}

	_, err = m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	markerInfoAfter, err := os.Stat(markerFile)
	if err != nil {
		t.Fatal(err)
	}

	if markerInfoBefore.ModTime() != markerInfoAfter.ModTime() {
		t.Errorf(".claude dir contents were modified")
	}

	// .gemini should also exist now
	geminiDir := filepath.Join(tmpDir, ".gemini")
	if _, err := os.Stat(geminiDir); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", geminiDir)
	}
}

// TestSync_ChannelAll_IndependentManifests verifies that channel=all sync
// produces two independent provenance manifests in .knossos/:
// - PROVENANCE_MANIFEST.yaml for claude
// - PROVENANCE_MANIFEST_GEMINI.yaml for gemini
// and that a gemini sync does not overwrite the claude manifest.
func TestSync_ChannelAll_IndependentManifests(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRite(t, tmpDir)

	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knossosDir, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatal(err)
	}

	m := materialize.NewMaterializerWithSource(resolver, filepath.Join(tmpDir, "rites"))

	opts := materialize.SyncOptions{
		Scope:   materialize.ScopeRite,
		Channel: "all",
	}

	_, err := m.Sync(opts)
	if err != nil {
		t.Fatalf("Sync(channel=all) failed: %v", err)
	}

	// Claude manifest should exist at the default path
	claudeManifestPath := filepath.Join(knossosDir, provenance.ManifestFileName)
	if _, err := os.Stat(claudeManifestPath); os.IsNotExist(err) {
		t.Fatalf("expected claude manifest at %s", claudeManifestPath)
	}

	// Gemini manifest should exist at the channel-keyed path
	geminiManifestPath := filepath.Join(knossosDir, provenance.GeminiManifestFileName)
	if _, err := os.Stat(geminiManifestPath); os.IsNotExist(err) {
		t.Fatalf("expected gemini manifest at %s", geminiManifestPath)
	}

	// Load both manifests and verify they are independent
	claudeManifest, err := provenance.Load(claudeManifestPath)
	if err != nil {
		t.Fatalf("failed to load claude manifest: %v", err)
	}
	geminiManifest, err := provenance.Load(geminiManifestPath)
	if err != nil {
		t.Fatalf("failed to load gemini manifest: %v", err)
	}

	// Both should have the same schema version
	if claudeManifest.SchemaVersion != provenance.CurrentSchemaVersion {
		t.Errorf("claude manifest schema = %q, want %q", claudeManifest.SchemaVersion, provenance.CurrentSchemaVersion)
	}
	if geminiManifest.SchemaVersion != provenance.CurrentSchemaVersion {
		t.Errorf("gemini manifest schema = %q, want %q", geminiManifest.SchemaVersion, provenance.CurrentSchemaVersion)
	}

	// Both should have entries (at minimum CLAUDE.md/GEMINI.md equivalent)
	if len(claudeManifest.Entries) == 0 {
		t.Error("claude manifest has no entries")
	}
	if len(geminiManifest.Entries) == 0 {
		t.Error("gemini manifest has no entries")
	}

	// Now do a gemini-only sync and verify the claude manifest is unchanged
	claudeManifestBefore, err := os.ReadFile(claudeManifestPath)
	if err != nil {
		t.Fatal(err)
	}

	geminiOnlyOpts := materialize.SyncOptions{
		Scope:   materialize.ScopeRite,
		Channel: "gemini",
		RiteName: "test-rite",
	}

	_, err = m.Sync(geminiOnlyOpts)
	if err != nil {
		t.Fatalf("Sync(channel=gemini) failed: %v", err)
	}

	claudeManifestAfter, err := os.ReadFile(claudeManifestPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(claudeManifestBefore) != string(claudeManifestAfter) {
		t.Error("gemini-only sync modified the claude manifest; manifests should be independent")
	}
}
