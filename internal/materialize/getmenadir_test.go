package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

// TestGetMenaDir_ResolutionOrder_KnossosHomeBeforeXDG verifies that getMenaDir()
// returns the KnossosHome mena directory when both KnossosHome and an XDG data
// directory contain a mena/ subdirectory.
//
// Resolution order (as of WS-1): project-level → KnossosHome → XDG data dir.
// Before WS-1 the order was project-level → XDG → KnossosHome, meaning an XDG
// install could shadow the developer's working knossos checkout. This regression
// test ensures KnossosHome wins over XDG when both are present.
func TestGetMenaDir_ResolutionOrder_KnossosHomeBeforeXDG(t *testing.T) {
	t.Parallel()
	// Create a KnossosHome directory with a mena/ subdirectory.
	knossosHome := t.TempDir()
	knossosMenaDir := filepath.Join(knossosHome, "mena")
	if err := os.MkdirAll(knossosMenaDir, 0755); err != nil {
		t.Fatalf("mkdir knossosMenaDir: %v", err)
	}

	// Create an XDG data directory with a mena/ subdirectory.
	xdgDataHome := t.TempDir()
	xdgMenaDir := filepath.Join(xdgDataHome, "knossos", "mena")
	if err := os.MkdirAll(xdgMenaDir, 0755); err != nil {
		t.Fatalf("mkdir xdgMenaDir: %v", err)
	}

	// Create a project root with no .knossos/mena/ so project-level is skipped.
	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sr := NewSourceResolverWithPaths(projectRoot, "", "", knossosHome)
	m := NewMaterializerWithSourceResolver(resolver, sr)
	// xdgDataDir replaces config.XDGDataDir() which returns $XDG_DATA_HOME/knossos
	m.xdgDataDir = filepath.Join(xdgDataHome, "knossos")

	got := m.getMenaDir()

	if got != knossosMenaDir {
		t.Errorf("getMenaDir() = %q, want KnossosHome mena %q\n"+
			"(XDG mena at %q must not shadow KnossosHome)",
			got, knossosMenaDir, xdgMenaDir)
	}
}

// TestGetMenaDir_ResolutionOrder_ProjectOverridesKnossosHome verifies that the
// project-level mena directory (.knossos/mena/) wins over KnossosHome when both
// exist. Project level is the highest-priority filesystem tier.
func TestGetMenaDir_ResolutionOrder_ProjectOverridesKnossosHome(t *testing.T) {
	t.Parallel()
	// Create a KnossosHome directory with a mena/ subdirectory.
	knossosHome := t.TempDir()
	knossosMenaDir := filepath.Join(knossosHome, "mena")
	if err := os.MkdirAll(knossosMenaDir, 0755); err != nil {
		t.Fatalf("mkdir knossosMenaDir: %v", err)
	}

	// Create a project root with a .knossos/mena/ directory.
	projectRoot := t.TempDir()
	projectMenaDir := filepath.Join(projectRoot, ".knossos", "mena")
	if err := os.MkdirAll(projectMenaDir, 0755); err != nil {
		t.Fatalf("mkdir projectMenaDir: %v", err)
	}

	resolver := paths.NewResolver(projectRoot)
	sr := NewSourceResolverWithPaths(projectRoot, "", "", knossosHome)
	m := NewMaterializerWithSourceResolver(resolver, sr)

	got := m.getMenaDir()

	if got != projectMenaDir {
		t.Errorf("getMenaDir() = %q, want project mena %q\n"+
			"(KnossosHome mena at %q must not shadow project-level)",
			got, projectMenaDir, knossosMenaDir)
	}
}

// TestGetMenaDir_ResolutionOrder_XDGFallbackWhenNoKnossosHome verifies that getMenaDir()
// falls back to XDG data dir when KnossosHome has no mena/ directory.
func TestGetMenaDir_ResolutionOrder_XDGFallbackWhenNoKnossosHome(t *testing.T) {
	t.Parallel()
	// KnossosHome exists but has NO mena/ subdirectory.
	knossosHome := t.TempDir()

	// XDG data directory has a mena/ subdirectory.
	xdgDataHome := t.TempDir()
	xdgMenaDir := filepath.Join(xdgDataHome, "knossos", "mena")
	if err := os.MkdirAll(xdgMenaDir, 0755); err != nil {
		t.Fatalf("mkdir xdgMenaDir: %v", err)
	}

	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sr := NewSourceResolverWithPaths(projectRoot, "", "", knossosHome)
	m := NewMaterializerWithSourceResolver(resolver, sr)
	// xdgDataDir replaces config.XDGDataDir() which returns $XDG_DATA_HOME/knossos
	m.xdgDataDir = filepath.Join(xdgDataHome, "knossos")

	got := m.getMenaDir()

	if got != xdgMenaDir {
		t.Errorf("getMenaDir() = %q, want XDG mena %q\n"+
			"(XDG must be returned when KnossosHome has no mena/)",
			got, xdgMenaDir)
	}
}

// TestGetMenaDir_ResolutionOrder_EmptyWhenNoneExist verifies that getMenaDir()
// returns "" when neither project-level, KnossosHome, nor XDG mena directories exist.
func TestGetMenaDir_ResolutionOrder_EmptyWhenNoneExist(t *testing.T) {
	t.Parallel()
	// All temp dirs exist as roots but none have a mena/ subdirectory.
	knossosHome := t.TempDir()
	xdgDataHome := t.TempDir()

	projectRoot := t.TempDir()
	resolver := paths.NewResolver(projectRoot)
	sr := NewSourceResolverWithPaths(projectRoot, "", "", knossosHome)
	m := NewMaterializerWithSourceResolver(resolver, sr)
	// xdgDataDir replaces config.XDGDataDir() which returns $XDG_DATA_HOME/knossos
	m.xdgDataDir = filepath.Join(xdgDataHome, "knossos")

	got := m.getMenaDir()

	if got != "" {
		t.Errorf("getMenaDir() = %q, want \"\" when no mena directory exists", got)
	}
}
