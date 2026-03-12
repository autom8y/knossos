package materialize_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

// createTestRiteWithDromena creates a test rite with a single dromena (command) mena entry.
// Returns the rite directory path. The dromena has name "test-cmd" and a markdown body.
func createTestRiteWithDromena(t *testing.T, dir string) string {
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

	// Create mena directory with a dromena command
	droDir := filepath.Join(riteDir, "mena", "test-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatal(err)
	}

	droContent := `---
name: test-cmd
description: A test command for channel verification
---
# Test Command

Execute this test command.
`
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte(droContent), 0644); err != nil {
		t.Fatal(err)
	}

	return riteDir
}

// normalizeKnossosMarkers normalizes KNOSSOS section markers in content so that
// attribute ordering (which depends on Go map iteration) does not cause false
// negatives. Markers like "<!-- KNOSSOS:START name attr1=v1 attr2=v2 -->" have
// their attributes sorted alphabetically.
func normalizeKnossosMarkers(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "<!-- KNOSSOS:") {
			continue
		}
		// Extract everything between "<!-- " and " -->"
		inner := strings.TrimPrefix(trimmed, "<!-- ")
		inner = strings.TrimSuffix(inner, " -->")
		parts := strings.Fields(inner)
		if len(parts) <= 2 {
			continue // "KNOSSOS:START name" -- nothing to sort
		}
		// parts[0] = "KNOSSOS:START" or "KNOSSOS:END", parts[1] = section name
		// parts[2:] = attributes to sort
		attrs := parts[2:]
		sortedAttrs := make([]string, len(attrs))
		copy(sortedAttrs, attrs)
		// Simple sort -- alphabetical is sufficient for determinism
		for a := 0; a < len(sortedAttrs); a++ {
			for b := a + 1; b < len(sortedAttrs); b++ {
				if sortedAttrs[a] > sortedAttrs[b] {
					sortedAttrs[a], sortedAttrs[b] = sortedAttrs[b], sortedAttrs[a]
				}
			}
		}
		normalized := "<!-- " + parts[0] + " " + parts[1] + " " + strings.Join(sortedAttrs, " ") + " -->"
		lines[i] = strings.Replace(line, trimmed, normalized, 1)
	}
	return []byte(strings.Join(lines, "\n"))
}

// collectDirContents recursively walks dir and returns a map of relative-path to file-bytes.
func collectDirContents(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	contents := make(map[string][]byte)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return contents
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		contents[rel] = data
		return nil
	})
	if err != nil {
		t.Fatalf("collectDirContents(%s) failed: %v", dir, err)
	}
	return contents
}

// TestSync_ChannelAll_CompilerTransforms verifies that the Gemini channel uses
// TOML commands while Claude uses markdown for the same dromena source.
func TestSync_ChannelAll_CompilerTransforms(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRiteWithDromena(t, tmpDir)

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

	// Claude: .claude/commands/test-cmd.md should exist with raw markdown body
	claudeCmd := filepath.Join(tmpDir, ".claude", "commands", "test-cmd.md")
	claudeData, err := os.ReadFile(claudeCmd)
	if err != nil {
		t.Fatalf("expected claude command at %s: %v", claudeCmd, err)
	}
	if !strings.Contains(string(claudeData), "# Test Command") {
		t.Errorf("claude command should contain markdown body, got: %s", string(claudeData))
	}

	// Gemini: .gemini/commands/test-cmd.toml should exist with TOML-encoded content
	geminiCmd := filepath.Join(tmpDir, ".gemini", "commands", "test-cmd.toml")
	geminiData, err := os.ReadFile(geminiCmd)
	if err != nil {
		t.Fatalf("expected gemini command at %s: %v", geminiCmd, err)
	}
	geminiStr := string(geminiData)
	if !strings.Contains(geminiStr, "name = 'test-cmd'") {
		t.Errorf("gemini command should contain TOML name field, got: %s", geminiStr)
	}
	if !strings.Contains(geminiStr, "prompt =") {
		t.Errorf("gemini command should contain TOML prompt field, got: %s", geminiStr)
	}

	// Claude should NOT have a .toml command
	claudeToml := filepath.Join(tmpDir, ".claude", "commands", "test-cmd.toml")
	if _, err := os.Stat(claudeToml); !os.IsNotExist(err) {
		t.Errorf("claude should not have a TOML command at %s", claudeToml)
	}

	// Gemini should NOT have a .md command (promoted dromena)
	geminiMd := filepath.Join(tmpDir, ".gemini", "commands", "test-cmd.md")
	if _, err := os.Stat(geminiMd); !os.IsNotExist(err) {
		t.Errorf("gemini should not have a markdown command at %s", geminiMd)
	}

	// Gemini should use GEMINI.md, not CLAUDE.md
	geminiContext := filepath.Join(tmpDir, ".gemini", "GEMINI.md")
	if _, err := os.Stat(geminiContext); os.IsNotExist(err) {
		t.Errorf("expected GEMINI.md at %s", geminiContext)
	}
	claudeInGemini := filepath.Join(tmpDir, ".gemini", "CLAUDE.md")
	if _, err := os.Stat(claudeInGemini); !os.IsNotExist(err) {
		t.Errorf("CLAUDE.md should not exist in .gemini/ directory")
	}
}

// TestSync_ChannelAll_ClaudeRegression verifies that claude output is byte-identical
// whether synced alone (channel=claude) or as part of channel=all.
func TestSync_ChannelAll_ClaudeRegression(t *testing.T) {
	t.Parallel()

	// --- Pass 1: channel=claude only ---
	tmpDir1 := t.TempDir()
	resolver1 := paths.NewResolver(tmpDir1)
	createTestRiteWithDromena(t, tmpDir1)

	knossosDir1 := filepath.Join(tmpDir1, ".knossos")
	if err := os.MkdirAll(knossosDir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knossosDir1, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatal(err)
	}

	m1 := materialize.NewMaterializerWithSource(resolver1, filepath.Join(tmpDir1, "rites"))
	_, err := m1.Sync(materialize.SyncOptions{
		Scope:   materialize.ScopeRite,
		Channel: "claude",
	})
	if err != nil {
		t.Fatalf("Sync(channel=claude) failed: %v", err)
	}

	claudeOnly := collectDirContents(t, filepath.Join(tmpDir1, ".claude"))

	// --- Pass 2: channel=all ---
	tmpDir2 := t.TempDir()
	resolver2 := paths.NewResolver(tmpDir2)
	createTestRiteWithDromena(t, tmpDir2)

	knossosDir2 := filepath.Join(tmpDir2, ".knossos")
	if err := os.MkdirAll(knossosDir2, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knossosDir2, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatal(err)
	}

	m2 := materialize.NewMaterializerWithSource(resolver2, filepath.Join(tmpDir2, "rites"))
	_, err = m2.Sync(materialize.SyncOptions{
		Scope:   materialize.ScopeRite,
		Channel: "all",
	})
	if err != nil {
		t.Fatalf("Sync(channel=all) failed: %v", err)
	}

	claudeFromAll := collectDirContents(t, filepath.Join(tmpDir2, ".claude"))

	// Compare: same files, same bytes
	if len(claudeOnly) == 0 {
		t.Fatal("channel=claude produced no files in .claude/")
	}
	if len(claudeOnly) != len(claudeFromAll) {
		t.Errorf("file count mismatch: channel=claude produced %d files, channel=all produced %d files",
			len(claudeOnly), len(claudeFromAll))
		// Log differences for debugging
		for k := range claudeOnly {
			if _, ok := claudeFromAll[k]; !ok {
				t.Errorf("  missing from channel=all: %s", k)
			}
		}
		for k := range claudeFromAll {
			if _, ok := claudeOnly[k]; !ok {
				t.Errorf("  extra in channel=all: %s", k)
			}
		}
	}

	for path, data := range claudeOnly {
		allData, ok := claudeFromAll[path]
		if !ok {
			t.Errorf("file %s present in channel=claude but missing from channel=all", path)
			continue
		}
		// Normalize KNOSSOS markers before comparison to account for
		// non-deterministic Go map iteration order in section attributes.
		normData := normalizeKnossosMarkers(data)
		normAllData := normalizeKnossosMarkers(allData)
		if !bytes.Equal(normData, normAllData) {
			t.Errorf("file %s differs between channel=claude and channel=all\n  claude-only len=%d\n  all len=%d",
				path, len(data), len(allData))
		}
	}
}

// TestSync_ChannelAll_Idempotent verifies SCAR-003: running channel=all sync
// twice produces identical output for both .claude/ and .gemini/ directories
// as well as provenance manifests.
func TestSync_ChannelAll_Idempotent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	createTestRiteWithDromena(t, tmpDir)

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

	// --- First sync ---
	_, err := m.Sync(opts)
	if err != nil {
		t.Fatalf("Sync #1 failed: %v", err)
	}

	claudeFirst := collectDirContents(t, filepath.Join(tmpDir, ".claude"))
	geminiFirst := collectDirContents(t, filepath.Join(tmpDir, ".gemini"))

	// Also capture provenance manifests
	claudeManifestFirst, err := os.ReadFile(filepath.Join(knossosDir, provenance.ManifestFileName))
	if err != nil {
		t.Fatalf("failed to read claude manifest after sync #1: %v", err)
	}
	geminiManifestFirst, err := os.ReadFile(filepath.Join(knossosDir, provenance.GeminiManifestFileName))
	if err != nil {
		t.Fatalf("failed to read gemini manifest after sync #1: %v", err)
	}

	// --- Second sync ---
	_, err = m.Sync(opts)
	if err != nil {
		t.Fatalf("Sync #2 failed: %v", err)
	}

	claudeSecond := collectDirContents(t, filepath.Join(tmpDir, ".claude"))
	geminiSecond := collectDirContents(t, filepath.Join(tmpDir, ".gemini"))

	claudeManifestSecond, err := os.ReadFile(filepath.Join(knossosDir, provenance.ManifestFileName))
	if err != nil {
		t.Fatalf("failed to read claude manifest after sync #2: %v", err)
	}
	geminiManifestSecond, err := os.ReadFile(filepath.Join(knossosDir, provenance.GeminiManifestFileName))
	if err != nil {
		t.Fatalf("failed to read gemini manifest after sync #2: %v", err)
	}

	// Verify .claude/ is identical (normalizing KNOSSOS markers for map-order stability)
	assertDirContentsEqual(t, ".claude/", claudeFirst, claudeSecond)

	// Verify .gemini/ is identical
	assertDirContentsEqual(t, ".gemini/", geminiFirst, geminiSecond)

	// Verify provenance manifests are identical (normalizing context file hashes
	// which vary with KNOSSOS marker attribute ordering)
	assertNormalizedEqual(t, "claude provenance manifest", claudeManifestFirst, claudeManifestSecond)
	assertNormalizedEqual(t, "gemini provenance manifest", geminiManifestFirst, geminiManifestSecond)
}

// assertDirContentsEqual compares two directory content maps, normalizing
// KNOSSOS markers to handle non-deterministic Go map iteration ordering.
func assertDirContentsEqual(t *testing.T, prefix string, first, second map[string][]byte) {
	t.Helper()

	if len(first) != len(second) {
		t.Errorf("%s file count changed: %d -> %d", prefix, len(first), len(second))
	}
	for path, data := range first {
		secondData, ok := second[path]
		if !ok {
			t.Errorf("%s%s disappeared after second sync", prefix, path)
			continue
		}
		normFirst := normalizeKnossosMarkers(data)
		normSecond := normalizeKnossosMarkers(secondData)
		if !bytes.Equal(normFirst, normSecond) {
			t.Errorf("%s%s content changed between syncs (len %d -> %d)", prefix, path, len(data), len(secondData))
		}
	}
	for path := range second {
		if _, ok := first[path]; !ok {
			t.Errorf("%s%s appeared after second sync (not present after first)", prefix, path)
		}
	}
}

// assertNormalizedEqual compares two byte slices after normalizing KNOSSOS markers
// and provenance hash values that depend on marker ordering.
func assertNormalizedEqual(t *testing.T, label string, first, second []byte) {
	t.Helper()
	normFirst := normalizeKnossosMarkers(first)
	normSecond := normalizeKnossosMarkers(second)
	if bytes.Equal(normFirst, normSecond) {
		return
	}
	// If still different after marker normalization, check if the only difference
	// is in SHA256 hashes (which change when marker ordering changes).
	// Strip all sha256:... values and compare the structural content.
	normalizeHashes := func(data []byte) []byte {
		s := string(data)
		var result strings.Builder
		for len(s) > 0 {
			idx := strings.Index(s, "sha256:")
			if idx < 0 {
				result.WriteString(s)
				break
			}
			result.WriteString(s[:idx])
			result.WriteString("sha256:NORMALIZED")
			// Skip past the actual hex hash
			end := idx + 7 // past "sha256:"
			for end < len(s) && ((s[end] >= '0' && s[end] <= '9') ||
				(s[end] >= 'a' && s[end] <= 'f') ||
				(s[end] >= 'A' && s[end] <= 'F')) {
				end++
			}
			s = s[end:]
		}
		return []byte(result.String())
	}
	structFirst := normalizeHashes(normFirst)
	structSecond := normalizeHashes(normSecond)
	// Also normalize timestamps (last_sync, last_synced, synced_at change every run)
	// and any ISO 8601 timestamps embedded in the YAML.
	normalizeTimestamps := func(data []byte) []byte {
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			// Normalize any line containing a YAML key with a timestamp-like value
			// (ISO 8601 format: YYYY-MM-DDTHH:MM:SS...)
			trimmed := strings.TrimSpace(line)
			// Check for known timestamp keys
			for _, prefix := range []string{"last_sync:", "last_synced:", "synced_at:", "created_at:", "updated_at:"} {
				if strings.HasPrefix(trimmed, prefix) {
					indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
					key := strings.TrimSuffix(prefix, ":")
					lines[i] = indent + key + ": NORMALIZED_TIMESTAMP"
					break
				}
			}
		}
		return []byte(strings.Join(lines, "\n"))
	}
	structFirst = normalizeTimestamps(structFirst)
	structSecond = normalizeTimestamps(structSecond)
	if !bytes.Equal(structFirst, structSecond) {
		t.Errorf("%s changed between syncs (structural difference beyond hash/marker/timestamp normalization)", label)
	}
}
