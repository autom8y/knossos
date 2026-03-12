package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sync"
)

func TestStateManager_InitializeAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	channelDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	manager := sync.NewStateManager(resolver)

	// Initially not initialized
	if manager.IsInitialized() {
		t.Error("expected sync not to be initialized")
	}

	// Initialize
	state, err := manager.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if state.SchemaVersion != "1.1" {
		t.Errorf("SchemaVersion = %q, want %q", state.SchemaVersion, "1.1")
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

	if loaded.SchemaVersion != "1.1" {
		t.Errorf("Loaded SchemaVersion = %q, want %q", loaded.SchemaVersion, "1.1")
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
