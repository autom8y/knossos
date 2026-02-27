package mena

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolvePlatformMenaDir_ProjectRoot verifies that a mena/ directory
// under projectRoot is returned first.
func TestResolvePlatformMenaDir_ProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	menaDir := filepath.Join(projectRoot, "mena")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatal(err)
	}

	got := ResolvePlatformMenaDir(projectRoot, "")
	if got != menaDir {
		t.Errorf("expected %q, got %q", menaDir, got)
	}
}

// TestResolvePlatformMenaDir_KnossosHome verifies that when no projectRoot
// mena exists, the knossosHome/mena/ is returned.
func TestResolvePlatformMenaDir_KnossosHome(t *testing.T) {
	projectRoot := t.TempDir() // no mena/ inside
	knossosHome := t.TempDir()
	menaDir := filepath.Join(knossosHome, "mena")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatal(err)
	}

	// XDG_DATA_HOME must not interfere — set it to a dir without mena
	xdgBase := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdgBase)

	got := ResolvePlatformMenaDir(projectRoot, knossosHome)
	if got != menaDir {
		t.Errorf("expected %q, got %q", menaDir, got)
	}
}

// TestResolvePlatformMenaDir_None verifies that "" is returned when no
// mena directory exists in any resolution path.
func TestResolvePlatformMenaDir_None(t *testing.T) {
	projectRoot := t.TempDir() // no mena/ inside
	knossosHome := t.TempDir() // no mena/ inside

	// XDG_DATA_HOME must not interfere — set it to a dir without mena
	xdgBase := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdgBase)

	got := ResolvePlatformMenaDir(projectRoot, knossosHome)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
