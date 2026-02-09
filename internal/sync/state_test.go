package sync_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sync"
)

func TestStateManager_InitializeAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	manager := sync.NewStateManager(resolver)

	// Initially not initialized
	if manager.IsInitialized() {
		t.Error("expected sync not to be initialized")
	}

	// Initialize
	state, err := manager.Initialize("https://example.com/config")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if state.Remote != "https://example.com/config" {
		t.Errorf("Remote = %q, want %q", state.Remote, "https://example.com/config")
	}

	// Now should be initialized
	if !manager.IsInitialized() {
		t.Error("expected sync to be initialized after Initialize()")
	}

	// Load should return the state
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Remote != "https://example.com/config" {
		t.Errorf("Loaded Remote = %q, want %q", loaded.Remote, "https://example.com/config")
	}
}

func TestStateManager_TrackedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	manager := sync.NewStateManager(resolver)

	state, err := manager.Initialize("https://example.com")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Add tracked file
	manager.UpdateTrackedFile(state, ".claude/CLAUDE.md", "abc123", "abc123", "abc123", "synced")

	if len(state.TrackedFiles) != 1 {
		t.Errorf("TrackedFiles count = %d, want 1", len(state.TrackedFiles))
	}

	tracked := state.TrackedFiles[".claude/CLAUDE.md"]
	if tracked.LocalHash != "abc123" {
		t.Errorf("LocalHash = %q, want %q", tracked.LocalHash, "abc123")
	}

	if tracked.Status != "synced" {
		t.Errorf("Status = %q, want %q", tracked.Status, "synced")
	}
}

func TestState_ActiveRiteField(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	manager := sync.NewStateManager(resolver)

	state, err := manager.Initialize("local:hygiene")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Set ActiveRite
	state.ActiveRite = "hygiene"
	if err := manager.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.ActiveRite != "hygiene" {
		t.Errorf("ActiveRite = %q, want %q", loaded.ActiveRite, "hygiene")
	}
}

func TestState_ActiveRiteOmittedWhenEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	manager := sync.NewStateManager(resolver)

	state, err := manager.Initialize("local:none")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Save without setting ActiveRite
	if err := manager.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read raw JSON and verify active_rite is absent
	data, err := os.ReadFile(manager.StatePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if strings.Contains(string(data), "active_rite") {
		t.Error("expected active_rite to be omitted from JSON when empty")
	}
}

func TestComputeContentHash(t *testing.T) {
	content := []byte("hello world")
	hash := sync.ComputeContentHash(content)

	// SHA-256 of "hello world" with sha256: prefix
	expected := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("hash = %q, want %q", hash, expected)
	}
}

func TestComputeFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write test content
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hash, err := sync.ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("ComputeFileHash() error = %v", err)
	}

	expected := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("hash = %q, want %q", hash, expected)
	}

	// Non-existent file should return empty hash
	hash, err = sync.ComputeFileHash(filepath.Join(tmpDir, "nonexistent.txt"))
	if err != nil {
		t.Fatalf("ComputeFileHash() for nonexistent error = %v", err)
	}
	if hash != "" {
		t.Errorf("hash for nonexistent = %q, want empty", hash)
	}
}
