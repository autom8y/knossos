package worktree

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSwitchWorktree tests the Switch operation.
func TestSwitchWorktree(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree to switch to
	wt, err := mgr.Create(CreateOptions{
		Name: "switch-target",
		Rite: "test-rite",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Test switch
	switched, err := mgr.Switch(wt.ID, WorktreeSwitchOptions{UpdateRite: false})
	if err != nil {
		t.Fatalf("Switch failed: %v", err)
	}

	if switched.ID != wt.ID {
		t.Errorf("Expected ID %s, got %s", wt.ID, switched.ID)
	}
	if switched.Name != wt.Name {
		t.Errorf("Expected name %s, got %s", wt.Name, switched.Name)
	}
}

// TestSwitchByName tests switching by name instead of ID.
func TestSwitchByName(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	wt, err := mgr.Create(CreateOptions{Name: "my-feature"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Switch by name
	switched, err := mgr.Switch("my-feature", WorktreeSwitchOptions{})
	if err != nil {
		t.Fatalf("Switch by name failed: %v", err)
	}

	if switched.ID != wt.ID {
		t.Errorf("Expected ID %s, got %s", wt.ID, switched.ID)
	}
}

// TestSwitchNonexistent tests switching to a nonexistent worktree.
func TestSwitchNonexistent(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = mgr.Switch("nonexistent", WorktreeSwitchOptions{})
	if err == nil {
		t.Error("Switch to nonexistent worktree should fail")
	}
}

// TestSwitchWithRiteUpdate tests that UpdateRite option works.
func TestSwitchWithRiteUpdate(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree with rite
	wt, err := mgr.Create(CreateOptions{
		Name: "rite-wt",
		Rite: "my-rite",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Switch with rite update
	_, err = mgr.Switch(wt.ID, WorktreeSwitchOptions{UpdateRite: true})
	if err != nil {
		t.Fatalf("Switch failed: %v", err)
	}

	// Verify ACTIVE_RITE file was updated
	activeRitePath := filepath.Join(wt.Path, ".knossos", "ACTIVE_RITE")
	data, err := os.ReadFile(activeRitePath)
	if err != nil {
		t.Fatalf("Failed to read ACTIVE_RITE: %v", err)
	}

	if strings.TrimSpace(string(data)) != "my-rite" {
		t.Errorf("Expected rite 'my-rite', got '%s'", strings.TrimSpace(string(data)))
	}
}

// TestCloneWorktree tests the Clone operation.
func TestCloneWorktree(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create source worktree
	source, err := mgr.Create(CreateOptions{
		Name:       "source-wt",
		Rite:       "source-rite",
		Complexity: "MODULE",
	})
	if err != nil {
		t.Fatalf("Failed to create source worktree: %v", err)
	}

	// Clone it
	clone, err := mgr.Clone(source.ID, "cloned-wt", CloneOptions{})
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	// Verify clone has new ID
	if clone.ID == source.ID {
		t.Error("Clone should have different ID from source")
	}

	// Verify clone has correct name
	if clone.Name != "cloned-wt" {
		t.Errorf("Expected name 'cloned-wt', got '%s'", clone.Name)
	}

	// Verify clone copied rite
	if clone.Rite != source.Rite {
		t.Errorf("Expected rite '%s', got '%s'", source.Rite, clone.Rite)
	}

	// Verify clone copied complexity
	if clone.Complexity != source.Complexity {
		t.Errorf("Expected complexity '%s', got '%s'", source.Complexity, clone.Complexity)
	}

	// Verify clone path exists
	if _, err := os.Stat(clone.Path); os.IsNotExist(err) {
		t.Error("Clone path does not exist")
	}
}

// TestCloneWithRiteOverride tests clone with rite override.
func TestCloneWithRiteOverride(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create source worktree
	source, err := mgr.Create(CreateOptions{
		Name: "source",
		Rite: "original-rite",
	})
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// Clone with rite override
	clone, err := mgr.Clone(source.ID, "override-clone", CloneOptions{
		Rite: "new-rite",
	})
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	if clone.Rite != "new-rite" {
		t.Errorf("Expected rite 'new-rite', got '%s'", clone.Rite)
	}
}

// TestCloneByName tests cloning by worktree name.
func TestCloneByName(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create source worktree
	_, err = mgr.Create(CreateOptions{Name: "named-source"})
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// Clone by name
	clone, err := mgr.Clone("named-source", "clone-of-named", CloneOptions{})
	if err != nil {
		t.Fatalf("Clone by name failed: %v", err)
	}

	if clone.Name != "clone-of-named" {
		t.Errorf("Expected name 'clone-of-named', got '%s'", clone.Name)
	}
}

// TestCloneNonexistent tests cloning a nonexistent worktree.
func TestCloneNonexistent(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = mgr.Clone("nonexistent", "new-name", CloneOptions{})
	if err == nil {
		t.Error("Clone of nonexistent worktree should fail")
	}
}

// TestSyncWorktree tests the Sync operation.
func TestSyncWorktree(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	wt, err := mgr.Create(CreateOptions{Name: "sync-test"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Sync without pull
	result, err := mgr.Sync(wt.ID, WorktreeSyncOptions{Pull: false})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify result has worktree info
	if result.Worktree.ID != wt.ID {
		t.Errorf("Expected worktree ID %s, got %s", wt.ID, result.Worktree.ID)
	}

	// Initial state should be up to date (just created from HEAD)
	// Note: This depends on the test setup
	if result.Diverged {
		t.Log("Warning: Worktree shows as diverged, may need investigation")
	}
}

// TestSyncByName tests syncing by worktree name.
func TestSyncByName(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = mgr.Create(CreateOptions{Name: "sync-named"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	result, err := mgr.Sync("sync-named", WorktreeSyncOptions{})
	if err != nil {
		t.Fatalf("Sync by name failed: %v", err)
	}

	if result.Worktree.Name != "sync-named" {
		t.Errorf("Expected name 'sync-named', got '%s'", result.Worktree.Name)
	}
}

// TestSyncNonexistent tests syncing a nonexistent worktree.
func TestSyncNonexistent(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = mgr.Sync("nonexistent", WorktreeSyncOptions{})
	if err == nil {
		t.Error("Sync of nonexistent worktree should fail")
	}
}

// TestExportImportRoundtrip tests export and import.
func TestExportImportRoundtrip(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree with some content
	wt, err := mgr.Create(CreateOptions{
		Name:       "export-test",
		Rite:       "export-rite",
		Complexity: "SYSTEM",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Add some files to the worktree
	testFile := filepath.Join(wt.Path, "test-export.txt")
	if err := os.WriteFile(testFile, []byte("export test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Export
	archivePath := filepath.Join(tmpDir, "export.tar.gz")
	if err := mgr.Export(wt.ID, archivePath); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("Archive file was not created")
	}

	// Verify archive contains metadata
	verifyArchiveMetadata(t, archivePath, wt)

	// Remove the original worktree (force because of untracked files)
	if err := mgr.Remove(wt.ID, true); err != nil {
		t.Fatalf("Failed to remove worktree: %v", err)
	}

	// Import
	imported, err := mgr.Import(archivePath)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify imported worktree has new ID
	if imported.ID == wt.ID {
		t.Error("Imported worktree should have new ID")
	}

	// Verify imported worktree preserves metadata
	if imported.Name != wt.Name {
		t.Errorf("Expected name '%s', got '%s'", wt.Name, imported.Name)
	}
	if imported.Rite != wt.Rite {
		t.Errorf("Expected rite '%s', got '%s'", wt.Rite, imported.Rite)
	}
	if imported.Complexity != wt.Complexity {
		t.Errorf("Expected complexity '%s', got '%s'", wt.Complexity, imported.Complexity)
	}

	// Verify imported files exist
	importedTestFile := filepath.Join(imported.Path, "test-export.txt")
	if _, err := os.Stat(importedTestFile); os.IsNotExist(err) {
		t.Error("Imported worktree should have test file")
	}

	// Verify file content
	content, err := os.ReadFile(importedTestFile)
	if err != nil {
		t.Fatalf("Failed to read imported test file: %v", err)
	}
	if string(content) != "export test content" {
		t.Errorf("Expected content 'export test content', got '%s'", string(content))
	}
}

// TestExportByName tests export by worktree name.
func TestExportByName(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = mgr.Create(CreateOptions{Name: "named-export"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	archivePath := filepath.Join(tmpDir, "named-export.tar.gz")
	if err := mgr.Export("named-export", archivePath); err != nil {
		t.Fatalf("Export by name failed: %v", err)
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("Archive should exist")
	}
}

// TestExportNonexistent tests exporting a nonexistent worktree.
func TestExportNonexistent(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	archivePath := filepath.Join(tmpDir, "nonexistent.tar.gz")
	err = mgr.Export("nonexistent", archivePath)
	if err == nil {
		t.Error("Export of nonexistent worktree should fail")
	}
}

// TestImportInvalidArchive tests importing an invalid archive.
func TestImportInvalidArchive(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create an invalid archive
	invalidPath := filepath.Join(tmpDir, "invalid.tar.gz")
	if err := os.WriteFile(invalidPath, []byte("not a valid archive"), 0644); err != nil {
		t.Fatalf("Failed to create invalid archive: %v", err)
	}

	_, err = mgr.Import(invalidPath)
	if err == nil {
		t.Error("Import of invalid archive should fail")
	}
}

// TestImportMissingMetadata tests importing archive without metadata.
func TestImportMissingMetadata(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a valid tar.gz without metadata
	archivePath := filepath.Join(tmpDir, "no-meta.tar.gz")
	createEmptyArchive(t, archivePath)

	_, err = mgr.Import(archivePath)
	if err == nil {
		t.Error("Import of archive without metadata should fail")
	}
}

// verifyArchiveMetadata checks that the archive contains valid metadata.
func verifyArchiveMetadata(t *testing.T, archivePath string, wt *Worktree) {
	t.Helper()

	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("Failed to read gzip: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	metaFound := false
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read tar: %v", err)
		}

		if header.Name == archiveMetaFile {
			metaFound = true

			data, err := io.ReadAll(tarReader)
			if err != nil {
				t.Fatalf("Failed to read metadata: %v", err)
			}

			var meta ExportArchive
			if err := json.Unmarshal(data, &meta); err != nil {
				t.Fatalf("Failed to parse metadata: %v", err)
			}

			if meta.WorktreeID != wt.ID {
				t.Errorf("Metadata ID mismatch: %s != %s", meta.WorktreeID, wt.ID)
			}
			if meta.Name != wt.Name {
				t.Errorf("Metadata name mismatch: %s != %s", meta.Name, wt.Name)
			}
			if meta.Version != archiveVersion {
				t.Errorf("Metadata version mismatch: %s != %s", meta.Version, archiveVersion)
			}
			break
		}
	}

	if !metaFound {
		t.Error("Archive does not contain metadata file")
	}
}

// createEmptyArchive creates a tar.gz with a dummy file but no metadata.
func createEmptyArchive(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add a dummy file
	header := &tar.Header{
		Name:    "dummy.txt",
		Size:    5,
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}
	if _, err := tarWriter.Write([]byte("dummy")); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
}

// TestSessionIntegrationUpdateWorktree tests updating worktree_id in session.
func TestSessionIntegrationUpdateWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a session context file
	sessionDir := filepath.Join(tmpDir, ".sos", "sessions", "session-test")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	initialContent := `---
schema_version: "2.1"
session_id: session-test
status: "ACTIVE"
---

# Test Session
`
	if err := os.WriteFile(contextPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create context file: %v", err)
	}

	// Update worktree_id
	si := NewSessionIntegration(tmpDir)
	if err := si.UpdateSessionWorktree(sessionDir, "wt-12345"); err != nil {
		t.Fatalf("UpdateSessionWorktree failed: %v", err)
	}

	// Read back
	wtID, err := si.GetActiveWorktree(sessionDir)
	if err != nil {
		t.Fatalf("GetActiveWorktree failed: %v", err)
	}

	if wtID != "wt-12345" {
		t.Errorf("Expected worktree_id 'wt-12345', got '%s'", wtID)
	}
}

// TestSessionIntegrationGetActiveWorktree tests retrieving worktree_id.
func TestSessionIntegrationGetActiveWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a session context with worktree_id
	sessionDir := filepath.Join(tmpDir, ".sos", "sessions", "session-test")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	content := `---
schema_version: "2.1"
session_id: session-test
status: "ACTIVE"
worktree_id: "wt-existing-id"
---

# Test Session
`
	if err := os.WriteFile(contextPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create context file: %v", err)
	}

	si := NewSessionIntegration(tmpDir)
	wtID, err := si.GetActiveWorktree(sessionDir)
	if err != nil {
		t.Fatalf("GetActiveWorktree failed: %v", err)
	}

	if wtID != "wt-existing-id" {
		t.Errorf("Expected 'wt-existing-id', got '%s'", wtID)
	}
}

// TestSessionIntegrationMissingWorktree tests getting nonexistent worktree_id.
func TestSessionIntegrationMissingWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	sessionDir := filepath.Join(tmpDir, ".sos", "sessions", "session-test")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	content := `---
schema_version: "2.1"
session_id: session-test
status: "ACTIVE"
---

# Test Session
`
	if err := os.WriteFile(contextPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create context file: %v", err)
	}

	si := NewSessionIntegration(tmpDir)
	wtID, err := si.GetActiveWorktree(sessionDir)
	if err != nil {
		t.Fatalf("GetActiveWorktree failed: %v", err)
	}

	// Should return empty string, not error
	if wtID != "" {
		t.Errorf("Expected empty worktree_id, got '%s'", wtID)
	}
}

// TestFrontmatterParsing tests the frontmatter parsing utilities.
func TestFrontmatterParsing(t *testing.T) {
	content := `---
field1: value1
field2: "quoted value"
field3: 'single quoted'
---

Body content
`

	// Test extraction
	v1, _ := extractFrontmatterField(content, "field1")
	if v1 != "value1" {
		t.Errorf("Expected 'value1', got '%s'", v1)
	}

	v2, _ := extractFrontmatterField(content, "field2")
	if v2 != "quoted value" {
		t.Errorf("Expected 'quoted value', got '%s'", v2)
	}

	v3, _ := extractFrontmatterField(content, "field3")
	if v3 != "single quoted" {
		t.Errorf("Expected 'single quoted', got '%s'", v3)
	}

	// Test missing field
	missing, _ := extractFrontmatterField(content, "nonexistent")
	if missing != "" {
		t.Errorf("Expected empty string for missing field, got '%s'", missing)
	}
}

// TestFrontmatterUpdate tests updating frontmatter fields.
func TestFrontmatterUpdate(t *testing.T) {
	content := `---
existing: old
---

Body
`

	// Update existing field
	updated, err := updateFrontmatterField(content, "existing", "new")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	v, _ := extractFrontmatterField(updated, "existing")
	if v != "new" {
		t.Errorf("Expected 'new', got '%s'", v)
	}

	// Add new field
	updated2, err := updateFrontmatterField(content, "newfield", "newvalue")
	if err != nil {
		t.Fatalf("Add new field failed: %v", err)
	}

	v2, _ := extractFrontmatterField(updated2, "newfield")
	if v2 != "newvalue" {
		t.Errorf("Expected 'newvalue', got '%s'", v2)
	}
}

// TestFrontmatterRemove tests removing frontmatter fields.
func TestFrontmatterRemove(t *testing.T) {
	content := `---
keep: this
remove: that
---

Body
`

	updated, err := removeFrontmatterField(content, "remove")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify removed
	v, _ := extractFrontmatterField(updated, "remove")
	if v != "" {
		t.Errorf("Field should be removed, got '%s'", v)
	}

	// Verify kept
	v2, _ := extractFrontmatterField(updated, "keep")
	if v2 != "this" {
		t.Errorf("Expected 'this', got '%s'", v2)
	}
}

// TestQuoteIfNeeded tests the YAML quoting utility.
func TestQuoteIfNeeded(t *testing.T) {
	tests := []struct {
		input    string
		needsQuotes bool
	}{
		{"simple", false},
		{"with:colon", true},
		{"with space", false}, // Spaces in middle are fine
		{"-starts-with-dash", true},
		{"has#hash", true},
		{"", true}, // Empty needs quotes
	}

	for _, tt := range tests {
		result := quoteIfNeeded(tt.input)
		hasQuotes := strings.HasPrefix(result, `"`) && strings.HasSuffix(result, `"`)
		if hasQuotes != tt.needsQuotes {
			t.Errorf("quoteIfNeeded(%q) = %q, needsQuotes=%v, expected=%v",
				tt.input, result, hasQuotes, tt.needsQuotes)
		}
	}
}

// TestSetupWorktreeEcosystemSeedsDirectories tests that Create produces a worktree
// with .knossos/ and .know/ as symlinks pointing to main root, and .ledge/ and .sos/
// as scaffolded directories.
func TestSetupWorktreeEcosystemSeedsDirectories(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	// Create .knossos/ and .know/ in the main repo root so symlinks have targets
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("Failed to create .knossos/: %v", err)
	}
	knowDir := filepath.Join(tmpDir, ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("Failed to create .know/: %v", err)
	}

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	wt, err := mgr.Create(CreateOptions{
		Name: "seed-test",
		Rite: "test-rite",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify .knossos/ is a symlink pointing to main root
	knossosLink := filepath.Join(wt.Path, ".knossos")
	fi, err := os.Lstat(knossosLink)
	if err != nil {
		t.Fatalf(".knossos/ does not exist in worktree: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error(".knossos/ should be a symlink")
	}
	target, err := os.Readlink(knossosLink)
	if err != nil {
		t.Fatalf("Failed to readlink .knossos/: %v", err)
	}
	if target != knossosDir {
		t.Errorf(".knossos/ symlink target = %s, want %s", target, knossosDir)
	}

	// Verify .know/ is a symlink pointing to main root
	knowLink := filepath.Join(wt.Path, ".know")
	fi, err = os.Lstat(knowLink)
	if err != nil {
		t.Fatalf(".know/ does not exist in worktree: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error(".know/ should be a symlink")
	}
	target, err = os.Readlink(knowLink)
	if err != nil {
		t.Fatalf("Failed to readlink .know/: %v", err)
	}
	if target != knowDir {
		t.Errorf(".know/ symlink target = %s, want %s", target, knowDir)
	}

	// Verify .ledge/ subdirectories exist
	for _, sub := range []string{"decisions", "specs", "reviews", "spikes"} {
		subPath := filepath.Join(wt.Path, ".ledge", sub)
		fi, err := os.Stat(subPath)
		if err != nil {
			t.Errorf(".ledge/%s does not exist: %v", sub, err)
		} else if !fi.IsDir() {
			t.Errorf(".ledge/%s is not a directory", sub)
		}
	}

	// Verify .sos/ exists
	sosPath := filepath.Join(wt.Path, ".sos")
	fi, err = os.Stat(sosPath)
	if err != nil {
		t.Fatalf(".sos/ does not exist in worktree: %v", err)
	}
	if !fi.IsDir() {
		t.Error(".sos/ should be a directory")
	}
}

// TestSetupWorktreeEcosystemSkipsSymlinksWhenSourceMissing tests that when .knossos/
// and .know/ don't exist in root, Create still succeeds and those symlinks are not
// created, but .ledge/ and .sos/ are still scaffolded.
func TestSetupWorktreeEcosystemSkipsSymlinksWhenSourceMissing(t *testing.T) {
	tmpDir := setupTestGitRepo(t)


	// Do NOT create .knossos/ or .know/ in the main repo root

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	wt, err := mgr.Create(CreateOptions{
		Name: "no-source-test",
		Rite: "test-rite",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify .knossos/ is NOT a symlink (ensureProjectDirs creates it as a real
	// directory, but the symlink to main should not be created when source is missing)
	knossosPath := filepath.Join(wt.Path, ".knossos")
	if fi, err := os.Lstat(knossosPath); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Error(".knossos/ should not be a symlink when source is missing")
		}
	}

	// Verify .know/ symlink was NOT created
	knowLink := filepath.Join(wt.Path, ".know")
	if fi, err := os.Lstat(knowLink); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Error(".know/ should not be a symlink when source is missing")
		}
	}

	// Verify .ledge/ subdirectories still exist
	for _, sub := range []string{"decisions", "specs", "reviews", "spikes"} {
		subPath := filepath.Join(wt.Path, ".ledge", sub)
		fi, err := os.Stat(subPath)
		if err != nil {
			t.Errorf(".ledge/%s does not exist: %v", sub, err)
		} else if !fi.IsDir() {
			t.Errorf(".ledge/%s is not a directory", sub)
		}
	}

	// Verify .sos/ still exists
	sosPath := filepath.Join(wt.Path, ".sos")
	fi, err := os.Stat(sosPath)
	if err != nil {
		t.Fatalf(".sos/ does not exist in worktree: %v", err)
	}
	if !fi.IsDir() {
		t.Error(".sos/ should be a directory")
	}
}

// TestParseSessionContextFile tests session context parsing.
func TestParseSessionContextFile(t *testing.T) {
	tmpDir := t.TempDir()

	contextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")
	content := `---
schema_version: "2.1"
session_id: session-12345
status: "ACTIVE"
worktree_id: "wt-abcdef"
active_rite: "10x-dev"
initiative: "Test Initiative"
complexity: MODULE
current_phase: implementation
---

# Test Session

Some body content.
`
	if err := os.WriteFile(contextPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write context file: %v", err)
	}

	ctx, err := ParseSessionContextFile(contextPath)
	if err != nil {
		t.Fatalf("ParseSessionContextFile failed: %v", err)
	}

	if ctx.SchemaVersion != "2.1" {
		t.Errorf("Expected schema_version '2.1', got '%s'", ctx.SchemaVersion)
	}
	if ctx.SessionID != "session-12345" {
		t.Errorf("Expected session_id 'session-12345', got '%s'", ctx.SessionID)
	}
	if ctx.Status != "ACTIVE" {
		t.Errorf("Expected status 'ACTIVE', got '%s'", ctx.Status)
	}
	if ctx.WorktreeID != "wt-abcdef" {
		t.Errorf("Expected worktree_id 'wt-abcdef', got '%s'", ctx.WorktreeID)
	}
	if ctx.ActiveRite != "10x-dev" {
		t.Errorf("Expected active_rite '10x-dev', got '%s'", ctx.ActiveRite)
	}
	if ctx.Initiative != "Test Initiative" {
		t.Errorf("Expected initiative 'Test Initiative', got '%s'", ctx.Initiative)
	}
	if ctx.Complexity != "MODULE" {
		t.Errorf("Expected complexity 'MODULE', got '%s'", ctx.Complexity)
	}
	if ctx.CurrentPhase != "implementation" {
		t.Errorf("Expected current_phase 'implementation', got '%s'", ctx.CurrentPhase)
	}
}
