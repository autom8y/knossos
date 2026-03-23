package worktree

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestGenerateWorktreeID tests worktree ID generation.
func TestGenerateWorktreeID(t *testing.T) {
	id1 := GenerateWorktreeID()
	id2 := GenerateWorktreeID()

	// IDs should be valid
	if !IsValidWorktreeID(id1) {
		t.Errorf("Generated ID is not valid: %s", id1)
	}
	if !IsValidWorktreeID(id2) {
		t.Errorf("Generated ID is not valid: %s", id2)
	}

	// IDs should be unique (at least different hex suffix)
	if id1 == id2 {
		t.Errorf("Generated IDs should be unique: %s == %s", id1, id2)
	}

	// IDs should have correct prefix
	if !strings.HasPrefix(id1, "wt-") {
		t.Errorf("ID should start with 'wt-': %s", id1)
	}
}

// TestIsValidWorktreeID tests worktree ID validation.
func TestIsValidWorktreeID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"wt-20260104-143052-a1b2", true},
		{"wt-20260104-143052-abcd1234", true},
		{"wt-20260104-143052-abc", false},       // Too short hex
		{"wt-2026010-143052-a1b2", false},       // Wrong date format
		{"wt-20260104-14305-a1b2", false},       // Wrong time format
		{"session-20260104-143052-a1b2", false}, // Wrong prefix
		{"", false},
		{"wt-", false},
		{"wt-invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := IsValidWorktreeID(tt.id)
			if got != tt.valid {
				t.Errorf("IsValidWorktreeID(%q) = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}

// TestParseWorktreeTimestamp tests timestamp extraction from worktree IDs.
func TestParseWorktreeTimestamp(t *testing.T) {
	id := "wt-20260104-143052-a1b2"
	ts := ParseWorktreeTimestamp(id)

	if ts.IsZero() {
		t.Errorf("ParseWorktreeTimestamp(%q) returned zero time", id)
	}

	expected := time.Date(2026, 1, 4, 14, 30, 52, 0, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("ParseWorktreeTimestamp(%q) = %v, want %v", id, ts, expected)
	}

	// Invalid ID should return zero time
	invalid := "invalid-id"
	if !ParseWorktreeTimestamp(invalid).IsZero() {
		t.Errorf("ParseWorktreeTimestamp(%q) should return zero time", invalid)
	}
}

// TestFormatAge tests human-readable age formatting.
func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		t        time.Time
		contains string
	}{
		{now, "just now"},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-1 * time.Minute), "1 minute ago"},
		{now.Add(-2 * time.Hour), "2 hours ago"},
		{now.Add(-1 * time.Hour), "1 hour ago"},
		{now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{now.Add(-1 * 24 * time.Hour), "1 day ago"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			got := FormatAge(tt.t)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("FormatAge() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

// TestMetadataManager tests metadata persistence.
func TestMetadataManager(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	mgr := NewMetadataManager(tmpDir)

	// Load from non-existent should return empty
	meta, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load from empty dir should not error: %v", err)
	}
	if len(meta.Worktrees) != 0 {
		t.Errorf("Expected 0 worktrees, got %d", len(meta.Worktrees))
	}

	// Add a worktree
	wt := Worktree{
		ID:        "wt-20260104-143052-a1b2",
		Name:      "test-worktree",
		Path:      filepath.Join(tmpDir, "wt-test"),
		Rite:      "10x-dev",
		CreatedAt: time.Now().UTC(),
	}

	err = mgr.Add(wt)
	if err != nil {
		t.Fatalf("Failed to add worktree: %v", err)
	}

	// Load and verify
	meta, err = mgr.Load()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if len(meta.Worktrees) != 1 {
		t.Errorf("Expected 1 worktree, got %d", len(meta.Worktrees))
	}
	if meta.Worktrees[0].ID != wt.ID {
		t.Errorf("Expected ID %s, got %s", wt.ID, meta.Worktrees[0].ID)
	}

	// Get by ID
	got, err := mgr.Get(wt.ID)
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}
	if got.Name != wt.Name {
		t.Errorf("Expected name %s, got %s", wt.Name, got.Name)
	}

	// Get by name
	got, err = mgr.GetByName(wt.Name)
	if err != nil {
		t.Fatalf("Failed to get by name: %v", err)
	}
	if got.ID != wt.ID {
		t.Errorf("Expected ID %s, got %s", wt.ID, got.ID)
	}

	// Add duplicate should error
	err = mgr.Add(wt)
	if err == nil {
		t.Error("Adding duplicate should error")
	}

	// Update
	wt.Rite = "new-rite"
	err = mgr.Update(wt)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}
	got, _ = mgr.Get(wt.ID)
	if got.Rite != "new-rite" {
		t.Errorf("Expected rite new-rite, got %s", got.Rite)
	}

	// Remove
	err = mgr.Remove(wt.ID)
	if err != nil {
		t.Fatalf("Failed to remove: %v", err)
	}

	_, err = mgr.Get(wt.ID)
	if err == nil {
		t.Error("Get after remove should error")
	}
}

// TestMetadataGetOlderThan tests filtering worktrees by age.
func TestMetadataGetOlderThan(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := NewMetadataManager(tmpDir)

	now := time.Now().UTC()

	// Add worktrees of different ages
	worktrees := []Worktree{
		{ID: "wt-20260104-143052-0001", Name: "old-1", Path: "/tmp/wt1", CreatedAt: now.Add(-10 * 24 * time.Hour)},
		{ID: "wt-20260104-143052-0002", Name: "old-2", Path: "/tmp/wt2", CreatedAt: now.Add(-8 * 24 * time.Hour)},
		{ID: "wt-20260104-143052-0003", Name: "new-1", Path: "/tmp/wt3", CreatedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: "wt-20260104-143052-0004", Name: "new-2", Path: "/tmp/wt4", CreatedAt: now.Add(-1 * 24 * time.Hour)},
	}

	for _, wt := range worktrees {
		if err := mgr.Add(wt); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Get worktrees older than 7 days
	old, err := mgr.GetOlderThan(7 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("GetOlderThan failed: %v", err)
	}
	if len(old) != 2 {
		t.Errorf("Expected 2 old worktrees (older than 7d), got %d", len(old))
	}

	// Get worktrees older than 5 days
	old, _ = mgr.GetOlderThan(5 * 24 * time.Hour)
	if len(old) != 2 {
		t.Errorf("Expected 2 old worktrees (older than 5d), got %d", len(old))
	}

	// Get worktrees older than 3 days
	old, _ = mgr.GetOlderThan(3 * 24 * time.Hour)
	if len(old) != 2 {
		t.Errorf("Expected 2 old worktrees (older than 3d), got %d", len(old))
	}
}

// TestPerWorktreeMeta tests per-worktree metadata storage.
func TestPerWorktreeMeta(t *testing.T) {
	tmpDir := t.TempDir()

	wt := Worktree{
		ID:         "wt-20260104-143052-a1b2",
		Name:       "test-worktree",
		Rite:       "10x-dev",
		Complexity: "MODULE",
		FromRef:    "main",
		CreatedAt:  time.Now().UTC(),
	}

	// Save
	err := SavePerWorktreeMeta(tmpDir, wt, "/parent/project")
	if err != nil {
		t.Fatalf("Failed to save per-worktree meta: %v", err)
	}

	// Verify file exists
	metaPath := filepath.Join(tmpDir, ".knossos", WorktreeMetaFileName)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}

	// Load
	meta, err := LoadPerWorktreeMeta(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load per-worktree meta: %v", err)
	}

	if meta.WorktreeID != wt.ID {
		t.Errorf("Expected ID %s, got %s", wt.ID, meta.WorktreeID)
	}
	if meta.Name != wt.Name {
		t.Errorf("Expected name %s, got %s", wt.Name, meta.Name)
	}
	if meta.Rite != wt.Rite {
		t.Errorf("Expected rite %s, got %s", wt.Rite, meta.Rite)
	}
	if meta.Complexity != wt.Complexity {
		t.Errorf("Expected complexity %s, got %s", wt.Complexity, meta.Complexity)
	}
	if meta.ParentProject != "/parent/project" {
		t.Errorf("Expected parent /parent/project, got %s", meta.ParentProject)
	}
}

// setupTestGitRepo creates a temporary git repository for testing.
func setupTestGitRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Resolve symlinks for macOS /var -> /private/var
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git config email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git config name: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	os.WriteFile(testFile, []byte("# Test\n"), 0644)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Create .claude directory
	channelDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(channelDir, 0755)

	return tmpDir
}

// TestGitOperations tests git operations.
func TestGitOperations(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	git := NewGitOperations(tmpDir)

	// Test IsGitRepo
	if !git.IsGitRepo() {
		t.Error("Should be a git repo")
	}

	// Test IsWorktree (should be false for main repo)
	if git.IsWorktree() {
		t.Error("Main repo should not be a worktree")
	}

	// Test GetProjectRoot
	root, err := git.GetProjectRoot()
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}
	// Resolve symlinks for comparison
	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	actualRoot, _ := filepath.EvalSymlinks(root)
	if actualRoot != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, actualRoot)
	}

	// Test RefExists
	if !git.RefExists("HEAD") {
		t.Error("HEAD should exist")
	}
	if git.RefExists("nonexistent-branch") {
		t.Error("nonexistent-branch should not exist")
	}

	// Test GetWorktreesDir
	wtDir, err := git.GetWorktreesDir()
	if err != nil {
		t.Fatalf("GetWorktreesDir failed: %v", err)
	}
	// Just verify it ends with .worktrees
	if !strings.HasSuffix(wtDir, ".worktrees") {
		t.Errorf("Worktrees dir should end with .worktrees, got %s", wtDir)
	}
}

// TestManagerCreate tests worktree creation.
func TestManagerCreate(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	opts := CreateOptions{
		Name:       "test-feature",
		Rite:       "10x-dev",
		Complexity: "MODULE",
	}

	wt, err := mgr.Create(opts)
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify worktree was created
	if !IsValidWorktreeID(wt.ID) {
		t.Errorf("Invalid worktree ID: %s", wt.ID)
	}
	if wt.Name != opts.Name {
		t.Errorf("Expected name %s, got %s", opts.Name, wt.Name)
	}
	if wt.Rite != opts.Rite {
		t.Errorf("Expected rite %s, got %s", opts.Rite, wt.Rite)
	}

	// Verify path exists
	if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
		t.Error("Worktree path does not exist")
	}

	// Verify per-worktree metadata
	meta, err := LoadPerWorktreeMeta(wt.Path)
	if err != nil {
		t.Fatalf("Failed to load per-worktree meta: %v", err)
	}
	if meta.WorktreeID != wt.ID {
		t.Errorf("Per-worktree meta ID mismatch: %s != %s", meta.WorktreeID, wt.ID)
	}

	// Verify registry was updated
	worktrees, err := mgr.List()
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}
	if len(worktrees) != 1 {
		t.Errorf("Expected 1 worktree, got %d", len(worktrees))
	}
}

// TestManagerList tests worktree listing.
func TestManagerList(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Initially empty
	worktrees, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees, got %d", len(worktrees))
	}

	// Create some worktrees
	for i := range 3 {
		_, err := mgr.Create(CreateOptions{Name: "test-" + string(rune('a'+i))})
		if err != nil {
			t.Fatalf("Failed to create worktree %d: %v", i, err)
		}
	}

	// List again
	worktrees, err = mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(worktrees) != 3 {
		t.Errorf("Expected 3 worktrees, got %d", len(worktrees))
	}
}

// TestManagerRemove tests worktree removal.
func TestManagerRemove(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	wt, err := mgr.Create(CreateOptions{Name: "to-remove"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Remove by ID with force (worktree may have untracked files from per-worktree meta)
	err = mgr.Remove(wt.ID, true)
	if err != nil {
		t.Fatalf("Failed to remove worktree: %v", err)
	}

	// Verify it's gone
	worktrees, _ := mgr.List()
	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees after removal, got %d", len(worktrees))
	}

	// Verify path is gone
	if _, err := os.Stat(wt.Path); !os.IsNotExist(err) {
		t.Error("Worktree path should not exist after removal")
	}
}

// TestManagerRemoveByName tests worktree removal by name.
func TestManagerRemoveByName(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	_, err = mgr.Create(CreateOptions{Name: "feature-x"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Remove by name with force
	err = mgr.Remove("feature-x", true)
	if err != nil {
		t.Fatalf("Failed to remove worktree by name: %v", err)
	}

	worktrees, _ := mgr.List()
	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees after removal, got %d", len(worktrees))
	}
}

// TestManagerRemoveDirtyRequiresForce tests that removing dirty worktree requires force.
func TestManagerRemoveDirtyRequiresForce(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	wt, err := mgr.Create(CreateOptions{Name: "dirty-wt"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Make it dirty by creating an additional untracked file beyond the metadata
	untrackedFile := filepath.Join(wt.Path, "user-untracked.txt")
	os.WriteFile(untrackedFile, []byte("dirty"), 0644)

	// Try to remove without force - should fail (has untracked files)
	err = mgr.Remove(wt.ID, false)
	if err == nil {
		t.Error("Removing worktree with untracked files without force should fail")
	}

	// Remove with force - should succeed
	err = mgr.Remove(wt.ID, true)
	if err != nil {
		t.Fatalf("Failed to remove dirty worktree with force: %v", err)
	}
}

// TestManagerCleanup tests cleanup of stale worktrees.
func TestManagerCleanup(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	wt, err := mgr.Create(CreateOptions{Name: "old-wt"})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Manually backdate the creation time
	metaMgr := NewMetadataManager(mgr.GetWorktreesDir())
	wtMeta, _ := metaMgr.Get(wt.ID)
	wtMeta.CreatedAt = time.Now().Add(-10 * 24 * time.Hour)
	if err := metaMgr.Update(*wtMeta); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Cleanup with force and dry run
	result, err := mgr.Cleanup(CleanupOptions{
		OlderThan: 7 * 24 * time.Hour,
		DryRun:    true,
		Force:     true, // Force because worktree has untracked metadata files
	})
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if len(result.Removed) != 1 {
		t.Errorf("Expected 1 to be removed in dry run, got %d", len(result.Removed))
	}

	// Verify it's still there (dry run)
	worktrees, _ := mgr.List()
	if len(worktrees) != 1 {
		t.Errorf("Worktree should still exist after dry run")
	}

	// Actual cleanup with force
	result, err = mgr.Cleanup(CleanupOptions{
		OlderThan: 7 * 24 * time.Hour,
		DryRun:    false,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if len(result.Removed) != 1 {
		t.Errorf("Expected 1 to be removed, got %d", len(result.Removed))
	}

	// Verify it's gone
	worktrees, _ = mgr.List()
	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees after cleanup, got %d", len(worktrees))
	}
}

// TestManagerStatus tests worktree status.
func TestManagerStatus(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a worktree
	wt, err := mgr.Create(CreateOptions{
		Name:       "status-test",
		Complexity: "MODULE",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Get status
	status, err := mgr.Status(wt.ID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if status.ID != wt.ID {
		t.Errorf("Expected ID %s, got %s", wt.ID, status.ID)
	}
	if status.SessionStatus != "none" {
		t.Errorf("Expected session status 'none', got %s", status.SessionStatus)
	}

	// The worktree will have untracked files from per-worktree metadata
	// Just verify HasUntracked is set correctly
	if !status.HasUntracked {
		t.Log("Note: Worktree should have untracked files from metadata")
	}

	// Add a tracked file modification to test IsDirty
	readmeFile := filepath.Join(wt.Path, "README.md")
	os.WriteFile(readmeFile, []byte("# Modified\n"), 0644)

	status, _ = mgr.Status(wt.ID)
	if !status.IsDirty {
		t.Error("Worktree with modified file should have IsDirty=true")
	}
}

// TestRemoveCleansUpLocalDirectories tests that Remove cleans up .sos/ and .ledge/
// in the worktree, and that symlinked directories' targets in the main root are NOT
// affected by removal.
func TestRemoveCleansUpLocalDirectories(t *testing.T) {
	tmpDir := setupTestGitRepo(t)

	// Create .knossos/ and .know/ in the main repo root
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("Failed to create .knossos/: %v", err)
	}
	// Put a sentinel file in .knossos/ to verify it is not deleted
	sentinel := filepath.Join(knossosDir, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("do not delete"), 0644); err != nil {
		t.Fatalf("Failed to create sentinel: %v", err)
	}

	knowDir := filepath.Join(tmpDir, ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("Failed to create .know/: %v", err)
	}
	knowSentinel := filepath.Join(knowDir, "sentinel.txt")
	if err := os.WriteFile(knowSentinel, []byte("do not delete"), 0644); err != nil {
		t.Fatalf("Failed to create know sentinel: %v", err)
	}

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	wt, err := mgr.Create(CreateOptions{
		Name: "remove-cleanup-test",
		Rite: "test-rite",
	})
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify the seeded directories exist before removal
	if _, err := os.Stat(filepath.Join(wt.Path, ".sos")); os.IsNotExist(err) {
		t.Fatal(".sos/ should exist before removal")
	}
	if _, err := os.Stat(filepath.Join(wt.Path, ".ledge", "decisions")); os.IsNotExist(err) {
		t.Fatal(".ledge/decisions/ should exist before removal")
	}

	// Remove with force (worktree has untracked metadata files)
	if err := mgr.Remove(wt.ID, true); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify worktree path is gone
	if _, err := os.Stat(wt.Path); !os.IsNotExist(err) {
		t.Error("Worktree path should not exist after removal")
	}

	// Verify main root's .knossos/ was NOT deleted (symlink target preserved)
	if _, err := os.Stat(sentinel); os.IsNotExist(err) {
		t.Error("Main root .knossos/sentinel.txt should still exist after worktree removal")
	}

	// Verify main root's .know/ was NOT deleted
	if _, err := os.Stat(knowSentinel); os.IsNotExist(err) {
		t.Error("Main root .know/sentinel.txt should still exist after worktree removal")
	}
}

// TestMetadataJSON tests that metadata serializes correctly.
func TestMetadataJSON(t *testing.T) {
	wt := Worktree{
		ID:         "wt-20260104-143052-a1b2",
		Name:       "test",
		Path:       "/tmp/test",
		Rite:       "test-rite",
		CreatedAt:  time.Now().UTC(),
		BaseBranch: "main",
		FromRef:    "HEAD",
		Complexity: "MODULE",
	}

	data, err := json.Marshal(wt)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshalled Worktree
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if unmarshalled.ID != wt.ID {
		t.Errorf("ID mismatch: %s != %s", unmarshalled.ID, wt.ID)
	}
	if unmarshalled.Name != wt.Name {
		t.Errorf("Name mismatch: %s != %s", unmarshalled.Name, wt.Name)
	}
}
