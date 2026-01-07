package sync_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sync"
)

// setupTestProject creates a temporary project structure for testing.
func setupTestProject(t *testing.T) (projectDir, remoteDir string) {
	t.Helper()

	// Create project directory
	projectDir = t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	// Create remote directory (simulating local remote)
	remoteDir = t.TempDir()
	remoteClaudeDir := filepath.Join(remoteDir, ".claude")
	if err := os.MkdirAll(remoteClaudeDir, 0755); err != nil {
		t.Fatalf("Failed to create remote .claude dir: %v", err)
	}

	return projectDir, remoteDir
}

// setupTestProjectWithFiles creates project with pre-existing files.
func setupTestProjectWithFiles(t *testing.T) (projectDir, remoteDir string) {
	t.Helper()

	projectDir, remoteDir = setupTestProject(t)

	// Create a local CLAUDE.md
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, []byte("# Local CLAUDE.md\nLocal content"), 0644); err != nil {
		t.Fatalf("Failed to write local CLAUDE.md: %v", err)
	}

	// Create a remote CLAUDE.md
	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, []byte("# Remote CLAUDE.md\nRemote content"), 0644); err != nil {
		t.Fatalf("Failed to write remote CLAUDE.md: %v", err)
	}

	return projectDir, remoteDir
}

// TestStatusCmd_NotInitialized verifies status when sync is not initialized.
func TestStatusCmd_NotInitialized(t *testing.T) {
	projectDir, _ := setupTestProject(t)

	resolver := paths.NewResolver(projectDir)
	stateManager := sync.NewStateManager(resolver)

	// Load state - should be nil
	state, err := stateManager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state != nil {
		t.Error("Expected nil state when not initialized")
	}

	// Verify not initialized
	if stateManager.IsInitialized() {
		t.Error("Expected sync to not be initialized")
	}
}

// TestPullCmd_InitializesSync verifies that pulling initializes sync state.
func TestPullCmd_InitializesSync(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create matching content in both locations to avoid conflict
	content := []byte("# Matching CLAUDE.md\nSame content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write local CLAUDE.md: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write remote CLAUDE.md: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initial pull
	result, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful pull, got conflicts: %d", result.ConflictCount)
	}

	// Verify state is initialized
	stateManager := sync.NewStateManager(resolver)
	if !stateManager.IsInitialized() {
		t.Error("Expected sync to be initialized after pull")
	}

	// Verify state has correct remote
	state, err := stateManager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Remote != remoteDir {
		t.Errorf("Remote = %q, want %q", state.Remote, remoteDir)
	}
}

// TestPullCmd_UpdatesFiles verifies that pull updates local files from remote.
func TestPullCmd_UpdatesFiles(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Force pull to overwrite conflicts
	result, err := puller.Pull(remoteDir, sync.PullOptions{Force: true})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if !result.Success {
		t.Error("Expected successful pull")
	}

	// Verify local file was updated
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(localClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read local CLAUDE.md: %v", err)
	}

	expected := "# Remote CLAUDE.md\nRemote content"
	if string(content) != expected {
		t.Errorf("Local content = %q, want %q", string(content), expected)
	}
}

// TestPullCmd_DryRunDoesNotModify verifies dry-run mode doesn't modify files.
func TestPullCmd_DryRunDoesNotModify(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	// Get original content
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	originalContent, err := os.ReadFile(localClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read local CLAUDE.md: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Dry run pull
	result, err := puller.Pull(remoteDir, sync.PullOptions{DryRun: true, Force: true})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if !result.Success {
		t.Error("Expected successful dry-run pull")
	}

	// Verify local file was NOT modified
	newContent, err := os.ReadFile(localClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read local CLAUDE.md after dry-run: %v", err)
	}

	if string(newContent) != string(originalContent) {
		t.Errorf("Content changed during dry-run: got %q, want %q", string(newContent), string(originalContent))
	}
}

// TestPushCmd_RequiresInitialization verifies push fails without initialization.
func TestPushCmd_RequiresInitialization(t *testing.T) {
	projectDir, _ := setupTestProject(t)

	resolver := paths.NewResolver(projectDir)
	pusher := sync.NewPusher(resolver)

	// Push without initialization should fail
	_, err := pusher.Push(sync.PushOptions{})
	if err == nil {
		t.Error("Expected error when pushing without initialization")
	}
}

// TestPushCmd_PushesLocalChanges verifies push sends local changes to remote.
func TestPushCmd_PushesLocalChanges(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize with force pull
	_, err := puller.Pull(remoteDir, sync.PullOptions{Force: true})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Modify local file
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	newContent := []byte("# Modified Local CLAUDE.md\nNew local content")
	if err := os.WriteFile(localClaudeMD, newContent, 0644); err != nil {
		t.Fatalf("Failed to write modified CLAUDE.md: %v", err)
	}

	// Push changes
	pusher := sync.NewPusher(resolver)
	result, err := pusher.Push(sync.PushOptions{})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Push failed: %s", result.RejectReason)
	}

	// Verify remote file was updated
	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	remoteContent, err := os.ReadFile(remoteClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read remote CLAUDE.md: %v", err)
	}

	if string(remoteContent) != string(newContent) {
		t.Errorf("Remote content = %q, want %q", string(remoteContent), string(newContent))
	}
}

// TestPushCmd_DryRunDoesNotModify verifies dry-run doesn't modify remote.
func TestPushCmd_DryRunDoesNotModify(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize
	_, err := puller.Pull(remoteDir, sync.PullOptions{Force: true})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Get original remote content
	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	originalContent, err := os.ReadFile(remoteClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read remote CLAUDE.md: %v", err)
	}

	// Modify local file
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, []byte("Modified content"), 0644); err != nil {
		t.Fatalf("Failed to write modified CLAUDE.md: %v", err)
	}

	// Dry run push
	pusher := sync.NewPusher(resolver)
	result, err := pusher.Push(sync.PushOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Dry-run push failed: %s", result.RejectReason)
	}

	// Verify remote file was NOT modified
	newRemoteContent, err := os.ReadFile(remoteClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read remote CLAUDE.md after dry-run: %v", err)
	}

	if string(newRemoteContent) != string(originalContent) {
		t.Errorf("Remote content changed during dry-run: got %q, want %q", string(newRemoteContent), string(originalContent))
	}
}

// TestConflictDetection_LocalAndRemoteChanged verifies three-way conflict detection.
func TestConflictDetection_LocalAndRemoteChanged(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create initial content in both locations
	baseContent := []byte("# Base CLAUDE.md\nOriginal content")

	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write local CLAUDE.md: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write remote CLAUDE.md: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initial sync
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Initial Pull() error = %v", err)
	}

	// Modify local file
	localModified := []byte("# Base CLAUDE.md\nLocal modification")
	if err := os.WriteFile(localClaudeMD, localModified, 0644); err != nil {
		t.Fatalf("Failed to modify local CLAUDE.md: %v", err)
	}

	// Modify remote file
	remoteModified := []byte("# Base CLAUDE.md\nRemote modification")
	if err := os.WriteFile(remoteClaudeMD, remoteModified, 0644); err != nil {
		t.Fatalf("Failed to modify remote CLAUDE.md: %v", err)
	}

	// Pull should detect conflict
	result, err := puller.Pull(remoteDir, sync.PullOptions{})
	// Pull may return an error for conflicts
	_ = err

	// Check for conflicts
	if result.ConflictCount == 0 {
		t.Error("Expected conflict when both local and remote changed")
	}

	// Verify conflict details
	if len(result.FilesConflict) == 0 {
		t.Error("Expected conflict entries in result")
	}
}

// TestConflictDetection_FirstSyncWithExisting verifies conflict on first sync with existing local file.
func TestConflictDetection_FirstSyncWithExisting(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create different content in local and remote
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, []byte("Local unique content"), 0644); err != nil {
		t.Fatalf("Failed to write local CLAUDE.md: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, []byte("Remote unique content"), 0644); err != nil {
		t.Fatalf("Failed to write remote CLAUDE.md: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// First pull with different content should detect conflict
	result, _ := puller.Pull(remoteDir, sync.PullOptions{})

	// Should report conflict on first sync when local file differs
	if result.ConflictCount == 0 {
		t.Error("Expected conflict on first sync with different local content")
	}
}

// TestResolve_OursStrategy verifies resolving conflict with 'ours' strategy.
func TestResolve_OursStrategy(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create base content
	baseContent := []byte("# Base content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Initial pull failed: %v", err)
	}

	// Create conflict
	localModified := []byte("# Local changes")
	if err := os.WriteFile(localClaudeMD, localModified, 0644); err != nil {
		t.Fatalf("Failed to modify local file: %v", err)
	}

	remoteModified := []byte("# Remote changes")
	if err := os.WriteFile(remoteClaudeMD, remoteModified, 0644); err != nil {
		t.Fatalf("Failed to modify remote file: %v", err)
	}

	// Pull to detect conflict
	puller.Pull(remoteDir, sync.PullOptions{})

	// Resolve with 'ours' strategy
	syncResolver := sync.NewResolver(resolver)
	result, err := syncResolver.Resolve(sync.ResolveOptions{
		Strategy: sync.ResolveOurs,
		Path:     ".claude/CLAUDE.md",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !result.Resolved {
		t.Error("Expected conflict to be resolved")
	}

	// Verify local content was preserved
	content, err := os.ReadFile(localClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read local file: %v", err)
	}

	if string(content) != string(localModified) {
		t.Errorf("Content = %q, want %q (local)", string(content), string(localModified))
	}
}

// TestResolve_TheirsStrategy verifies resolving conflict with 'theirs' strategy.
func TestResolve_TheirsStrategy(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create base content
	baseContent := []byte("# Base content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Initial pull failed: %v", err)
	}

	// Create conflict
	localModified := []byte("# Local changes")
	if err := os.WriteFile(localClaudeMD, localModified, 0644); err != nil {
		t.Fatalf("Failed to modify local file: %v", err)
	}

	remoteModified := []byte("# Remote changes")
	if err := os.WriteFile(remoteClaudeMD, remoteModified, 0644); err != nil {
		t.Fatalf("Failed to modify remote file: %v", err)
	}

	// Pull to detect conflict
	puller.Pull(remoteDir, sync.PullOptions{})

	// Resolve with 'theirs' strategy
	syncResolver := sync.NewResolver(resolver)
	result, err := syncResolver.Resolve(sync.ResolveOptions{
		Strategy: sync.ResolveTheirs,
		Path:     ".claude/CLAUDE.md",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !result.Resolved {
		t.Error("Expected conflict to be resolved")
	}

	// Verify local content was replaced with remote
	content, err := os.ReadFile(localClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read local file: %v", err)
	}

	if string(content) != string(remoteModified) {
		t.Errorf("Content = %q, want %q (remote)", string(content), string(remoteModified))
	}
}

// TestHistory_RecordsPullOperations verifies pull operations are recorded in history.
func TestHistory_RecordsPullOperations(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	historyManager := sync.NewHistoryManager(resolver)

	// Record a pull
	err := historyManager.RecordPull(remoteDir, []string{".claude/CLAUDE.md"}, true, "")
	if err != nil {
		t.Fatalf("RecordPull() error = %v", err)
	}

	// Verify history
	entries, err := historyManager.List(sync.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Operation != "pull" {
		t.Errorf("Operation = %q, want %q", entries[0].Operation, "pull")
	}

	if entries[0].Remote != remoteDir {
		t.Errorf("Remote = %q, want %q", entries[0].Remote, remoteDir)
	}

	if !entries[0].Success {
		t.Error("Expected success = true")
	}
}

// TestHistory_RecordsPushOperations verifies push operations are recorded in history.
func TestHistory_RecordsPushOperations(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	historyManager := sync.NewHistoryManager(resolver)

	// Record a push
	err := historyManager.RecordPush(remoteDir, []string{".claude/CLAUDE.md"}, true, "")
	if err != nil {
		t.Fatalf("RecordPush() error = %v", err)
	}

	// Verify history
	entries, err := historyManager.List(sync.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Operation != "push" {
		t.Errorf("Operation = %q, want %q", entries[0].Operation, "push")
	}
}

// TestHistory_FiltersByOperation verifies history can be filtered by operation.
func TestHistory_FiltersByOperation(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	historyManager := sync.NewHistoryManager(resolver)

	// Record multiple operations
	historyManager.RecordPull(remoteDir, []string{"file1"}, true, "")
	historyManager.RecordPush(remoteDir, []string{"file2"}, true, "")
	historyManager.RecordPull(remoteDir, []string{"file3"}, true, "")

	// Filter by pull
	entries, err := historyManager.List(sync.ListOptions{Operation: "pull"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 pull entries, got %d", len(entries))
	}

	for _, e := range entries {
		if e.Operation != "pull" {
			t.Errorf("Expected pull operation, got %q", e.Operation)
		}
	}
}

// TestDiff_DetectsModifiedFiles verifies diff detects local modifications.
func TestDiff_DetectsModifiedFiles(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create matching content
	content := []byte("# Original content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize sync
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Modify local file
	modifiedContent := []byte("# Modified content")
	if err := os.WriteFile(localClaudeMD, modifiedContent, 0644); err != nil {
		t.Fatalf("Failed to modify local file: %v", err)
	}

	// Check diff
	differ := sync.NewDiffer(resolver)
	result, err := differ.Diff(sync.DiffOptions{})
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if !result.HasChanges {
		t.Error("Expected diff to detect changes")
	}

	if result.ChangedFiles != 1 {
		t.Errorf("ChangedFiles = %d, want 1", result.ChangedFiles)
	}
}

// TestDiff_NoChangesWhenSynced verifies diff reports no changes when in sync.
func TestDiff_NoChangesWhenSynced(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create matching content
	content := []byte("# Same content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize sync
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Check diff (no modifications)
	differ := sync.NewDiffer(resolver)
	result, err := differ.Diff(sync.DiffOptions{})
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if result.HasChanges {
		t.Error("Expected no changes when synced")
	}
}

// TestReset_ClearsState verifies reset clears sync state.
func TestReset_ClearsState(t *testing.T) {
	projectDir, remoteDir := setupTestProjectWithFiles(t)

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize
	_, err := puller.Pull(remoteDir, sync.PullOptions{Force: true})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Verify initialized
	stateManager := sync.NewStateManager(resolver)
	if !stateManager.IsInitialized() {
		t.Fatal("Expected sync to be initialized")
	}

	// Reset
	if err := stateManager.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Verify not initialized
	if stateManager.IsInitialized() {
		t.Error("Expected sync to not be initialized after reset")
	}
}

// TestPushCmd_RejectsWithConflicts verifies push is rejected when conflicts exist.
func TestPushCmd_RejectsWithConflicts(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create base content
	baseContent := []byte("# Base content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Initial pull failed: %v", err)
	}

	// Create conflict
	localModified := []byte("# Local changes")
	if err := os.WriteFile(localClaudeMD, localModified, 0644); err != nil {
		t.Fatalf("Failed to modify local file: %v", err)
	}

	remoteModified := []byte("# Remote changes")
	if err := os.WriteFile(remoteClaudeMD, remoteModified, 0644); err != nil {
		t.Fatalf("Failed to modify remote file: %v", err)
	}

	// Pull to detect conflict
	puller.Pull(remoteDir, sync.PullOptions{})

	// Attempt push with conflicts
	pusher := sync.NewPusher(resolver)
	result, err := pusher.Push(sync.PushOptions{})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !result.Rejected {
		t.Error("Expected push to be rejected with conflicts")
	}

	if result.RejectReason == "" {
		t.Error("Expected reject reason to be set")
	}
}

// TestPushCmd_ForceOverridesConflicts verifies force flag bypasses conflict check.
func TestPushCmd_ForceOverridesConflicts(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create base content
	baseContent := []byte("# Base content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, baseContent, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Initial pull failed: %v", err)
	}

	// Create conflict
	localModified := []byte("# Local changes - force push")
	if err := os.WriteFile(localClaudeMD, localModified, 0644); err != nil {
		t.Fatalf("Failed to modify local file: %v", err)
	}

	remoteModified := []byte("# Remote changes")
	if err := os.WriteFile(remoteClaudeMD, remoteModified, 0644); err != nil {
		t.Fatalf("Failed to modify remote file: %v", err)
	}

	// Pull to detect conflict
	puller.Pull(remoteDir, sync.PullOptions{})

	// Force push with conflicts
	pusher := sync.NewPusher(resolver)
	result, err := pusher.Push(sync.PushOptions{Force: true})
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if result.Rejected {
		t.Errorf("Expected force push to succeed, got rejected: %s", result.RejectReason)
	}

	// Verify remote was updated
	remoteContent, err := os.ReadFile(remoteClaudeMD)
	if err != nil {
		t.Fatalf("Failed to read remote file: %v", err)
	}

	if string(remoteContent) != string(localModified) {
		t.Errorf("Remote content = %q, want %q", string(remoteContent), string(localModified))
	}
}

// TestSyncOutput_StatusTextFormat verifies status output in text format.
func TestSyncOutput_StatusTextFormat(t *testing.T) {
	out := output.SyncStatusOutput{
		Initialized:  true,
		Remote:       "/tmp/remote",
		LastSync:     "2026-01-05 12:00:00",
		TrackedPaths: []output.SyncTrackedPath{
			{Path: ".claude/CLAUDE.md", Status: "synced"},
			{Path: ".claude/settings.json", Status: "modified"},
		},
		HasConflicts: false,
	}

	text := out.Text()

	if text == "" {
		t.Error("Expected non-empty text output")
	}

	if !bytes.Contains([]byte(text), []byte("Remote:")) {
		t.Error("Expected text to contain 'Remote:'")
	}
}

// TestSyncOutput_PullOutputWithConflicts verifies pull output format with conflicts.
func TestSyncOutput_PullOutputWithConflicts(t *testing.T) {
	out := output.SyncPullOutput{
		Remote:       "/tmp/remote",
		Success:      false,
		HasConflicts: true,
		ConflictCount: 1,
		FilesConflict: []output.SyncConflictEntry{
			{
				Path:        ".claude/CLAUDE.md",
				Description: "Both modified",
			},
		},
	}

	text := out.Text()

	if !bytes.Contains([]byte(text), []byte("Conflicts:")) {
		t.Error("Expected text to contain 'Conflicts:'")
	}
}

// TestTrackedFiles_SyncedStatus verifies tracked files maintain synced status.
func TestTrackedFiles_SyncedStatus(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create matching content
	content := []byte("# Same content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize sync
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Load state and check status
	stateManager := sync.NewStateManager(resolver)
	state, err := stateManager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tracked, exists := state.TrackedFiles[".claude/CLAUDE.md"]
	if !exists {
		t.Fatal("Expected file to be tracked")
	}

	if tracked.Status != "synced" {
		t.Errorf("Status = %q, want %q", tracked.Status, "synced")
	}
}

// TestTrackedFiles_ModifiedStatus verifies tracked files detect modified status.
func TestTrackedFiles_ModifiedStatus(t *testing.T) {
	projectDir, remoteDir := setupTestProject(t)

	// Create matching content
	content := []byte("# Same content")
	localClaudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(localClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write local file: %v", err)
	}

	remoteClaudeMD := filepath.Join(remoteDir, ".claude", "CLAUDE.md")
	if err := os.WriteFile(remoteClaudeMD, content, 0644); err != nil {
		t.Fatalf("Failed to write remote file: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	puller := sync.NewPuller(resolver)

	// Initialize sync
	_, err := puller.Pull(remoteDir, sync.PullOptions{})
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	// Modify local file
	modifiedContent := []byte("# Modified content")
	if err := os.WriteFile(localClaudeMD, modifiedContent, 0644); err != nil {
		t.Fatalf("Failed to modify local file: %v", err)
	}

	// Refresh state
	stateManager := sync.NewStateManager(resolver)
	state, err := stateManager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tracker := sync.NewTracker(resolver, stateManager)
	tracker.RefreshAll(state)

	tracked, exists := state.TrackedFiles[".claude/CLAUDE.md"]
	if !exists {
		t.Fatal("Expected file to be tracked")
	}

	if tracked.Status != "modified" {
		t.Errorf("Status = %q, want %q", tracked.Status, "modified")
	}
}

// TestParseRemote_HTTPUrl verifies HTTP URL parsing.
func TestParseRemote_HTTPUrl(t *testing.T) {
	remote, err := sync.ParseRemote("https://example.com/config")
	if err != nil {
		t.Fatalf("ParseRemote() error = %v", err)
	}

	if remote.Type != sync.RemoteTypeHTTP {
		t.Errorf("Type = %q, want %q", remote.Type, sync.RemoteTypeHTTP)
	}

	if remote.URL != "https://example.com/config" {
		t.Errorf("URL = %q, want %q", remote.URL, "https://example.com/config")
	}
}

// TestParseRemote_LocalPath verifies local path parsing.
func TestParseRemote_LocalPath(t *testing.T) {
	remote, err := sync.ParseRemote("/tmp/config")
	if err != nil {
		t.Fatalf("ParseRemote() error = %v", err)
	}

	if remote.Type != sync.RemoteTypeLocal {
		t.Errorf("Type = %q, want %q", remote.Type, sync.RemoteTypeLocal)
	}
}

// TestParseRemote_GitHubShorthand verifies GitHub shorthand parsing.
func TestParseRemote_GitHubShorthand(t *testing.T) {
	remote, err := sync.ParseRemote("anthropic/ariadne")
	if err != nil {
		t.Fatalf("ParseRemote() error = %v", err)
	}

	if remote.Type != sync.RemoteTypeHTTP {
		t.Errorf("Type = %q, want %q", remote.Type, sync.RemoteTypeHTTP)
	}

	expectedURL := "https://raw.githubusercontent.com/anthropic/ariadne/main"
	if remote.URL != expectedURL {
		t.Errorf("URL = %q, want %q", remote.URL, expectedURL)
	}
}

// TestShortenHash verifies hash shortening.
func TestShortenHash(t *testing.T) {
	fullHash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	shortHash := shortenHash(fullHash)

	if len(shortHash) != 8 {
		t.Errorf("Short hash length = %d, want 8", len(shortHash))
	}

	if shortHash != "b94d27b9" {
		t.Errorf("Short hash = %q, want %q", shortHash, "b94d27b9")
	}
}

// TestShortenHash_ShortInput verifies handling of already short hashes.
func TestShortenHash_ShortInput(t *testing.T) {
	shortHash := shortenHash("abc")

	if shortHash != "abc" {
		t.Errorf("Short hash = %q, want %q", shortHash, "abc")
	}
}

// shortenHash returns first 8 characters of a hash (copied from status.go for testing).
func shortenHash(hash string) string {
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}

// TestComputeContentHash verifies content hashing.
func TestComputeContentHash(t *testing.T) {
	content := []byte("hello world")
	hash := sync.ComputeContentHash(content)

	// SHA-256 of "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("hash = %q, want %q", hash, expected)
	}
}

// TestHistory_Timestamp verifies history entries have timestamps.
func TestHistory_Timestamp(t *testing.T) {
	projectDir, _ := setupTestProject(t)

	resolver := paths.NewResolver(projectDir)
	historyManager := sync.NewHistoryManager(resolver)

	// Record operation
	err := historyManager.RecordPull("/tmp/remote", []string{"file1"}, true, "")
	if err != nil {
		t.Fatalf("RecordPull() error = %v", err)
	}

	// Verify timestamp
	entries, err := historyManager.List(sync.ListOptions{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	// Parse timestamp
	_, err = time.Parse(time.RFC3339, entries[0].Timestamp)
	if err != nil {
		t.Errorf("Failed to parse timestamp %q: %v", entries[0].Timestamp, err)
	}
}

// TestJSON_OutputMarshaling verifies JSON output can be marshaled.
func TestJSON_OutputMarshaling(t *testing.T) {
	out := output.SyncStatusOutput{
		Initialized:  true,
		Remote:       "/tmp/remote",
		TrackedPaths: []output.SyncTrackedPath{
			{Path: ".claude/CLAUDE.md", Status: "synced"},
		},
	}

	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Verify it can be unmarshaled
	var parsed output.SyncStatusOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if parsed.Remote != out.Remote {
		t.Errorf("Remote = %q, want %q", parsed.Remote, out.Remote)
	}
}
