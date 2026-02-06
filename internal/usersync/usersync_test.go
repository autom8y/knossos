package usersync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewSyncer_ValidTypes tests syncer creation for each resource type.
func TestNewSyncer_ValidTypes(t *testing.T) {
	// Skip if KNOSSOS_HOME not set
	if os.Getenv("KNOSSOS_HOME") == "" {
		t.Skip("KNOSSOS_HOME not set")
	}

	types := []ResourceType{ResourceAgents, ResourceSkills, ResourceCommands, ResourceHooks}
	for _, rt := range types {
		syncer, err := NewSyncer(rt)
		if err != nil {
			t.Errorf("NewSyncer(%s) failed: %v", rt, err)
			continue
		}
		if syncer == nil {
			t.Errorf("NewSyncer(%s) returned nil syncer", rt)
		}
		if syncer.resourceType != rt {
			t.Errorf("syncer.resourceType = %s, want %s", syncer.resourceType, rt)
		}
	}
}

// TestNewSyncer_InvalidType tests that invalid resource type returns error.
func TestNewSyncer_InvalidType(t *testing.T) {
	// Skip if KNOSSOS_HOME not set
	if os.Getenv("KNOSSOS_HOME") == "" {
		t.Skip("KNOSSOS_HOME not set")
	}

	_, err := NewSyncer(ResourceType("invalid"))
	if err == nil {
		t.Error("NewSyncer(invalid) should return error")
	}
}

// TestResourceType_Singular tests singular form of resource types.
func TestResourceType_Singular(t *testing.T) {
	tests := []struct {
		rt   ResourceType
		want string
	}{
		{ResourceAgents, "agent"},
		{ResourceSkills, "skill"},
		{ResourceCommands, "command"},
		{ResourceHooks, "hook"},
	}
	for _, tt := range tests {
		got := tt.rt.Singular()
		if got != tt.want {
			t.Errorf("%s.Singular() = %s, want %s", tt.rt, got, tt.want)
		}
	}
}

// TestResourceType_SourceDir tests source directory name for resource types.
func TestResourceType_SourceDir(t *testing.T) {
	tests := []struct {
		rt   ResourceType
		want string
	}{
		{ResourceAgents, "user-agents"},
		{ResourceSkills, "user-skills"},
		{ResourceCommands, "mena"},
		{ResourceHooks, "user-hooks"},
	}
	for _, tt := range tests {
		got := tt.rt.SourceDir()
		if got != tt.want {
			t.Errorf("%s.SourceDir() = %s, want %s", tt.rt, got, tt.want)
		}
	}
}

// TestResourceType_RiteSubDir tests rite subdirectory name for resource types.
func TestResourceType_RiteSubDir(t *testing.T) {
	tests := []struct {
		rt   ResourceType
		want string
	}{
		{ResourceAgents, "agents"},
		{ResourceSkills, "skills"},
		{ResourceCommands, "mena"},
		{ResourceHooks, "hooks"},
	}
	for _, tt := range tests {
		got := tt.rt.RiteSubDir()
		if got != tt.want {
			t.Errorf("%s.RiteSubDir() = %s, want %s", tt.rt, got, tt.want)
		}
	}
}

// TestComputeFileChecksum tests SHA256 checksum computation.
func TestComputeFileChecksum(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	checksum, err := ComputeFileChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeFileChecksum failed: %v", err)
	}

	// Check prefix
	if len(checksum) < 7 || checksum[:7] != "sha256:" {
		t.Errorf("checksum should start with 'sha256:', got %s", checksum)
	}

	// Verify consistency
	checksum2, err := ComputeFileChecksum(testFile)
	if err != nil {
		t.Fatalf("Second ComputeFileChecksum failed: %v", err)
	}
	if checksum != checksum2 {
		t.Errorf("checksums should be consistent, got %s and %s", checksum, checksum2)
	}
}

// TestComputeContentChecksum tests checksum of byte content.
func TestComputeContentChecksum(t *testing.T) {
	content := []byte("hello world")
	checksum := ComputeContentChecksum(content)

	if len(checksum) < 7 || checksum[:7] != "sha256:" {
		t.Errorf("checksum should start with 'sha256:', got %s", checksum)
	}

	// Verify consistency
	checksum2 := ComputeContentChecksum(content)
	if checksum != checksum2 {
		t.Errorf("checksums should be consistent")
	}
}

// TestVerifyChecksum tests checksum verification.
func TestVerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	checksum, _ := ComputeFileChecksum(testFile)

	// Should match
	ok, err := VerifyChecksum(testFile, checksum)
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if !ok {
		t.Error("VerifyChecksum should return true for matching checksum")
	}

	// Should not match wrong checksum
	ok, err = VerifyChecksum(testFile, "sha256:wrong")
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if ok {
		t.Error("VerifyChecksum should return false for wrong checksum")
	}
}

// TestManifest_LoadSave tests manifest load and save operations.
func TestManifest_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create manifest
	manifest := &Manifest{
		Version:  ManifestVersion,
		LastSync: time.Now().UTC(),
		Entries: map[string]Entry{
			"test.md": {
				Source:      SourceKnossos,
				InstalledAt: time.Now().UTC(),
				Checksum:    "sha256:abc123",
			},
		},
	}

	// Save
	err := SaveManifest(manifestPath, ResourceAgents, manifest)
	if err != nil {
		t.Fatalf("SaveManifest failed: %v", err)
	}

	// Load
	loaded, err := LoadManifest(manifestPath, ResourceAgents)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if loaded.Version != manifest.Version {
		t.Errorf("Version mismatch: got %s, want %s", loaded.Version, manifest.Version)
	}

	if len(loaded.Entries) != 1 {
		t.Errorf("Entries count: got %d, want 1", len(loaded.Entries))
	}

	entry, ok := loaded.Entries["test.md"]
	if !ok {
		t.Error("Entry 'test.md' not found")
	}
	if entry.Source != SourceKnossos {
		t.Errorf("Source: got %s, want %s", entry.Source, SourceKnossos)
	}
	if entry.Checksum != "sha256:abc123" {
		t.Errorf("Checksum: got %s, want sha256:abc123", entry.Checksum)
	}
}

// TestManifest_LoadNotExists tests loading non-existent manifest.
func TestManifest_LoadNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "nonexistent.json")

	manifest, err := LoadManifest(manifestPath, ResourceAgents)
	if err != nil {
		t.Fatalf("LoadManifest should not error for non-existent file: %v", err)
	}

	if manifest.Version != ManifestVersion {
		t.Errorf("Version: got %s, want %s", manifest.Version, ManifestVersion)
	}

	if len(manifest.Entries) != 0 {
		t.Errorf("Entries should be empty, got %d", len(manifest.Entries))
	}
}

// TestManifest_LoadCorrupt tests loading corrupt manifest.
func TestManifest_LoadCorrupt(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "corrupt.json")

	// Write corrupt JSON
	if err := os.WriteFile(manifestPath, []byte("not valid json{"), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := LoadManifest(manifestPath, ResourceAgents)
	if err != nil {
		t.Fatalf("LoadManifest should return empty manifest for corrupt file: %v", err)
	}

	// Should return empty manifest
	if len(manifest.Entries) != 0 {
		t.Error("Should return empty manifest for corrupt file")
	}

	// Should backup corrupt file
	backupPath := manifestPath + ".corrupt"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Should backup corrupt manifest")
	}
}

// TestSyncer_DryRun tests that dry-run doesn't modify files.
func TestSyncer_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create source directory with test file
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create syncer with explicit paths
	syncer := NewSyncerWithPaths(ResourceAgents, sourceDir, targetDir, manifestPath)

	// Sync with dry-run
	result, err := syncer.Sync(Options{DryRun: true})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Should report added
	if len(result.Changes.Added) != 1 {
		t.Errorf("Should report 1 added, got %d", len(result.Changes.Added))
	}

	// Target should not exist
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Error("Target directory should not be created in dry-run")
	}

	// Manifest should not exist
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("Manifest should not be created in dry-run")
	}
}

// TestSyncer_AddNew tests adding new files.
func TestSyncer_AddNew(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create source directory with test file
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	syncer := NewSyncerWithPaths(ResourceAgents, sourceDir, targetDir, manifestPath)

	result, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if len(result.Changes.Added) != 1 {
		t.Errorf("Should add 1 file, got %d", len(result.Changes.Added))
	}

	// Verify file exists in target
	if _, err := os.Stat(filepath.Join(targetDir, "test.md")); os.IsNotExist(err) {
		t.Error("File should exist in target")
	}

	// Verify manifest
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Manifest should exist")
	}
}

// TestSyncer_UpdateChanged tests updating changed files.
func TestSyncer_UpdateChanged(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create initial file and sync
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}

	syncer := NewSyncerWithPaths(ResourceAgents, sourceDir, targetDir, manifestPath)
	_, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatal(err)
	}

	// Update source file
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("v2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Sync again
	result, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if len(result.Changes.Updated) != 1 {
		t.Errorf("Should update 1 file, got %d", len(result.Changes.Updated))
	}

	// Verify content
	content, _ := os.ReadFile(filepath.Join(targetDir, "test.md"))
	if string(content) != "v2" {
		t.Errorf("Content should be updated, got %s", string(content))
	}
}

// TestSyncer_SkipUserCreated tests that user-created files are skipped.
func TestSyncer_SkipUserCreated(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create source file
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create target file with different content (user-created)
	if err := os.WriteFile(filepath.Join(targetDir, "test.md"), []byte("user"), 0644); err != nil {
		t.Fatal(err)
	}

	syncer := NewSyncerWithPaths(ResourceAgents, sourceDir, targetDir, manifestPath)

	result, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Should be skipped as user-created
	if len(result.Changes.Skipped) != 1 {
		t.Errorf("Should skip 1 file, got %d", len(result.Changes.Skipped))
	}

	// Content should not change
	content, _ := os.ReadFile(filepath.Join(targetDir, "test.md"))
	if string(content) != "user" {
		t.Errorf("User content should be preserved, got %s", string(content))
	}
}

// TestSyncer_ForceDiverged tests force overwrite of diverged files.
func TestSyncer_ForceDiverged(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create and sync initial file
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}

	syncer := NewSyncerWithPaths(ResourceAgents, sourceDir, targetDir, manifestPath)
	_, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatal(err)
	}

	// Modify target (diverge)
	if err := os.WriteFile(filepath.Join(targetDir, "test.md"), []byte("diverged"), 0644); err != nil {
		t.Fatal(err)
	}

	// Update source
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), []byte("v2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Sync without force - should skip
	result, _ := syncer.Sync(Options{})
	if len(result.Changes.Skipped) != 1 {
		t.Errorf("Should skip diverged file without --force")
	}

	// Sync with force - should update
	result, err = syncer.Sync(Options{Force: true})
	if err != nil {
		t.Fatalf("Force sync failed: %v", err)
	}

	if len(result.Changes.Updated) != 1 {
		t.Errorf("Should update 1 file with --force, got %d", len(result.Changes.Updated))
	}

	content, _ := os.ReadFile(filepath.Join(targetDir, "test.md"))
	if string(content) != "v2" {
		t.Errorf("Content should be updated with --force, got %s", string(content))
	}
}

// TestSyncer_RecoverExisting tests recovery mode.
func TestSyncer_RecoverExisting(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create same content in source and target (as if user copied manually)
	content := []byte("same content")
	if err := os.WriteFile(filepath.Join(sourceDir, "test.md"), content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "test.md"), content, 0644); err != nil {
		t.Fatal(err)
	}

	syncer := NewSyncerWithPaths(ResourceAgents, sourceDir, targetDir, manifestPath)

	// Sync with recover
	result, err := syncer.Sync(Options{Recover: true})
	if err != nil {
		t.Fatalf("Recover sync failed: %v", err)
	}

	// Should be adopted as unchanged (checksums match)
	if len(result.Changes.Unchanged) != 1 {
		t.Errorf("Should adopt 1 file as unchanged, got %d unchanged, %d added",
			len(result.Changes.Unchanged), len(result.Changes.Added))
	}
}

// TestSyncer_NestedDirectories tests syncing nested directory structures.
func TestSyncer_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create nested source structure
	nestedDir := filepath.Join(sourceDir, "category", "subcategory")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "skill.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Skills use nested structure
	syncer := NewSyncerWithPaths(ResourceSkills, sourceDir, targetDir, manifestPath)

	result, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if len(result.Changes.Added) != 1 {
		t.Errorf("Should add 1 file, got %d", len(result.Changes.Added))
	}

	// Verify nested path in result
	expectedPath := filepath.Join("category", "subcategory", "skill.md")
	if result.Changes.Added[0] != expectedPath {
		t.Errorf("Path should be %s, got %s", expectedPath, result.Changes.Added[0])
	}

	// Verify file exists in target
	if _, err := os.Stat(filepath.Join(targetDir, expectedPath)); os.IsNotExist(err) {
		t.Error("Nested file should exist in target")
	}
}

// TestSyncer_PreservesExecutable tests that executable permissions are preserved.
func TestSyncer_PreservesExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create source with executable script
	libDir := filepath.Join(sourceDir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(libDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hi"), 0755); err != nil {
		t.Fatal(err)
	}

	syncer := NewSyncerWithPaths(ResourceHooks, sourceDir, targetDir, manifestPath)

	_, err := syncer.Sync(Options{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify executable permission
	targetScript := filepath.Join(targetDir, "lib", "script.sh")
	info, err := os.Stat(targetScript)
	if err != nil {
		t.Fatalf("Failed to stat target: %v", err)
	}

	if info.Mode()&0111 == 0 {
		t.Error("Script should have execute permission")
	}
}

// TestResult_Text tests text output generation.
func TestResult_Text(t *testing.T) {
	result := Result{
		Resource: ResourceAgents,
		DryRun:   false,
		Source:   "/source",
		Target:   "/target",
		Changes: Changes{
			Added:   []string{"new.md"},
			Updated: []string{"changed.md"},
			Skipped: []SkippedEntry{
				{Name: "user.md", Reason: "user-created"},
			},
			Unchanged: []string{"same.md"},
		},
		Summary: Summary{
			Added:      1,
			Updated:    1,
			Skipped:    1,
			Unchanged:  1,
			Collisions: 0,
		},
	}

	text := result.Text()

	if text == "" {
		t.Error("Text() should return non-empty string")
	}

	// Check key elements
	if !containsString(text, "Syncing user agents") {
		t.Error("Should contain resource type")
	}
	if !containsString(text, "Added: new.md") {
		t.Error("Should contain added file")
	}
	if !containsString(text, "Updated: changed.md") {
		t.Error("Should contain updated file")
	}
	if !containsString(text, "Skipped: user.md") {
		t.Error("Should contain skipped file")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
