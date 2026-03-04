package know

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverServiceBoundaries_GoMod(t *testing.T) {
	root := t.TempDir()
	svcDir := filepath.Join(root, "services", "auth")
	if err := os.MkdirAll(svcDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(svcDir, "go.mod"), []byte("module svc/auth\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	boundaries, err := DiscoverServiceBoundaries(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(boundaries) != 1 {
		t.Fatalf("want 1 boundary, got %d", len(boundaries))
	}
	if boundaries[0].Path != "services/auth" {
		t.Errorf("path = %q, want %q", boundaries[0].Path, "services/auth")
	}
	if boundaries[0].MarkerType != "go" {
		t.Errorf("type = %q, want %q", boundaries[0].MarkerType, "go")
	}
	if boundaries[0].HasKnow {
		t.Error("HasKnow should be false")
	}
}

func TestDiscoverServiceBoundaries_Multiple(t *testing.T) {
	root := t.TempDir()
	// Go service
	goDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module svc/api\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Node service
	nodeDir := filepath.Join(root, "frontend", "web")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	boundaries, err := DiscoverServiceBoundaries(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(boundaries) != 2 {
		t.Fatalf("want 2 boundaries, got %d", len(boundaries))
	}
	// Sorted by path
	if boundaries[0].Path != "frontend/web" {
		t.Errorf("first path = %q, want %q", boundaries[0].Path, "frontend/web")
	}
	if boundaries[1].Path != "services/api" {
		t.Errorf("second path = %q, want %q", boundaries[1].Path, "services/api")
	}
}

func TestDiscoverServiceBoundaries_SkipsRoot(t *testing.T) {
	root := t.TempDir()
	// Root go.mod should be excluded
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module monorepo\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	boundaries, err := DiscoverServiceBoundaries(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(boundaries) != 0 {
		t.Errorf("want 0 boundaries (root excluded), got %d", len(boundaries))
	}
}

func TestDiscoverServiceBoundaries_DetectsExistingKnow(t *testing.T) {
	root := t.TempDir()
	svcDir := filepath.Join(root, "services", "payments")
	knowDir := filepath.Join(svcDir, ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(svcDir, "go.mod"), []byte("module svc/payments\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	boundaries, err := DiscoverServiceBoundaries(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(boundaries) != 1 {
		t.Fatalf("want 1 boundary, got %d", len(boundaries))
	}
	if !boundaries[0].HasKnow {
		t.Error("HasKnow should be true")
	}
}

func TestDiscoverServiceBoundaries_SkipsHidden(t *testing.T) {
	root := t.TempDir()
	// Hidden dir with go.mod should be skipped
	hiddenDir := filepath.Join(root, ".hidden", "service")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "go.mod"), []byte("module hidden\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// node_modules should be skipped
	nmDir := filepath.Join(root, "node_modules", "some-pkg")
	if err := os.MkdirAll(nmDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nmDir, "package.json"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	boundaries, err := DiscoverServiceBoundaries(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(boundaries) != 0 {
		t.Errorf("want 0 boundaries (hidden/node_modules skipped), got %d", len(boundaries))
	}
}

func TestDiscoverServiceBoundaries_Empty(t *testing.T) {
	root := t.TempDir()

	boundaries, err := DiscoverServiceBoundaries(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(boundaries) != 0 {
		t.Errorf("want 0 boundaries, got %d", len(boundaries))
	}
}
