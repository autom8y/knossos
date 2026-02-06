package materialize

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/config"
)

func TestSourceResolver_EmbeddedFallback(t *testing.T) {
	// Create in-memory fs.FS with a test rite
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 1.0\n"),
		},
		"rites/test-rite/agents/agent-one.md": &fstest.MapFile{
			Data: []byte("# Agent One\n"),
		},
	}

	// Point all filesystem sources at nonexistent paths
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver("/nonexistent-project")
	resolver.WithEmbeddedFS(fsys)

	// Should find rite in embedded FS when filesystem sources don't exist
	resolved, err := resolver.ResolveRite("test-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}

	if resolved.Source.Type != SourceEmbedded {
		t.Errorf("Expected source type %q, got %q", SourceEmbedded, resolved.Source.Type)
	}
	if resolved.Name != "test-rite" {
		t.Errorf("Expected rite name %q, got %q", "test-rite", resolved.Name)
	}
	if resolved.RitePath != "rites/test-rite" {
		t.Errorf("Expected rite path %q, got %q", "rites/test-rite", resolved.RitePath)
	}
	if resolved.ManifestPath != "rites/test-rite/manifest.yaml" {
		t.Errorf("Expected manifest path %q, got %q", "rites/test-rite/manifest.yaml", resolved.ManifestPath)
	}
	if resolved.TemplatesDir != "knossos/templates" {
		t.Errorf("Expected templates dir %q, got %q", "knossos/templates", resolved.TemplatesDir)
	}
}

func TestSourceResolver_FilesystemOverridesEmbedded(t *testing.T) {
	// Create a temp directory with a project rite
	tmpDir := t.TempDir()
	createTestRite(t, tmpDir, "test-rite")

	// Create embedded FS with the same rite
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 2.0\n"),
		},
	}

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver(tmpDir)
	resolver.WithEmbeddedFS(fsys)

	resolved, err := resolver.ResolveRite("test-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}

	// Filesystem (project) should override embedded
	if resolved.Source.Type != SourceProject {
		t.Errorf("Expected source type %q (filesystem wins), got %q", SourceProject, resolved.Source.Type)
	}
}

func TestSourceResolver_EmbeddedNotFoundReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/other-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: other-rite\nversion: 1.0\n"),
		},
	}

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver("/nonexistent-project")
	resolver.WithEmbeddedFS(fsys)

	_, err := resolver.ResolveRite("missing-rite", "")
	if err == nil {
		t.Fatal("Expected error for missing rite, got nil")
	}
}

func TestSourceResolver_NoEmbeddedFS(t *testing.T) {
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver("/nonexistent-project")
	// No embedded FS set

	_, err := resolver.ResolveRite("any-rite", "")
	if err == nil {
		t.Fatal("Expected error when no sources available, got nil")
	}
}

func TestSourceResolver_ListIncludesEmbedded(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/embedded-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: embedded-rite\nversion: 1.0\n"),
		},
		"rites/another-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: another-rite\nversion: 1.0\n"),
		},
	}

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver("/nonexistent-project")
	resolver.WithEmbeddedFS(fsys)

	rites, err := resolver.ListAvailableRites()
	if err != nil {
		t.Fatalf("ListAvailableRites failed: %v", err)
	}

	if len(rites) != 2 {
		t.Fatalf("Expected 2 rites, got %d", len(rites))
	}

	// Both should be embedded source type
	for _, r := range rites {
		if r.Source.Type != SourceEmbedded {
			t.Errorf("Expected source type %q for rite %q, got %q", SourceEmbedded, r.Name, r.Source.Type)
		}
	}
}

func TestSourceResolver_ListShadowsEmbedded(t *testing.T) {
	// Create a temp dir with a project rite that shadows an embedded one
	tmpDir := t.TempDir()
	createTestRite(t, tmpDir, "shadowed-rite")

	fsys := fstest.MapFS{
		"rites/shadowed-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: shadowed-rite\nversion: 2.0\n"),
		},
		"rites/embedded-only/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: embedded-only\nversion: 1.0\n"),
		},
	}

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver(tmpDir)
	resolver.WithEmbeddedFS(fsys)

	rites, err := resolver.ListAvailableRites()
	if err != nil {
		t.Fatalf("ListAvailableRites failed: %v", err)
	}

	if len(rites) != 2 {
		t.Fatalf("Expected 2 rites (1 project + 1 embedded-only), got %d", len(rites))
	}

	// Find the shadowed rite and verify it comes from project, not embedded
	for _, r := range rites {
		if r.Name == "shadowed-rite" && r.Source.Type != SourceProject {
			t.Errorf("Expected project source for shadowed-rite, got %q", r.Source.Type)
		}
		if r.Name == "embedded-only" && r.Source.Type != SourceEmbedded {
			t.Errorf("Expected embedded source for embedded-only, got %q", r.Source.Type)
		}
	}
}

func TestSourceResolver_EmbeddedCaching(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/cached-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: cached-rite\nversion: 1.0\n"),
		},
	}

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := NewSourceResolver("/nonexistent-project")
	resolver.WithEmbeddedFS(fsys)

	// First resolution
	resolved1, err := resolver.ResolveRite("cached-rite", "")
	if err != nil {
		t.Fatalf("First resolve failed: %v", err)
	}

	// Second resolution should use cache
	resolved2, err := resolver.ResolveRite("cached-rite", "")
	if err != nil {
		t.Fatalf("Second resolve failed: %v", err)
	}

	if resolved1 != resolved2 {
		t.Error("Expected cached result to return same pointer")
	}
}

// createTestRite creates a minimal rite directory structure for testing.
func createTestRite(t *testing.T, baseDir, riteName string) {
	t.Helper()
	riteDir := filepath.Join(baseDir, "rites", riteName)
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("Failed to create rite dir: %v", err)
	}
	manifest := "name: " + riteName + "\nversion: 1.0\n"
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}
}
