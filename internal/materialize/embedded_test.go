package materialize

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/provenance"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
)

func TestCopyDirFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"foo.md":     &fstest.MapFile{Data: []byte("# Foo\n")},
		"bar.md":     &fstest.MapFile{Data: []byte("# Bar\n")},
		"sub/baz.md": &fstest.MapFile{Data: []byte("# Baz\n")},
	}

	dst := t.TempDir()
	if err := copyDirFromFS(fsys, dst); err != nil {
		t.Fatalf("copyDirFromFS failed: %v", err)
	}

	// Verify top-level files
	for _, name := range []string{"foo.md", "bar.md"} {
		data, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Errorf("Expected file %s, got error: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("File %s is empty", name)
		}
	}

	// Verify nested file
	data, err := os.ReadFile(filepath.Join(dst, "sub", "baz.md"))
	if err != nil {
		t.Fatalf("Expected sub/baz.md, got error: %v", err)
	}
	if string(data) != "# Baz\n" {
		t.Errorf("sub/baz.md content = %q, want %q", string(data), "# Baz\n")
	}
}

func TestCopyDirFromFS_EmptyFS(t *testing.T) {
	fsys := fstest.MapFS{}
	dst := t.TempDir()
	if err := copyDirFromFS(fsys, dst); err != nil {
		t.Fatalf("copyDirFromFS with empty FS failed: %v", err)
	}
}

func TestCopyDirFromFS_SubFS(t *testing.T) {
	fsys := fstest.MapFS{
		"agents/one.md": &fstest.MapFile{Data: []byte("# One\n")},
		"agents/two.md": &fstest.MapFile{Data: []byte("# Two\n")},
		"other/x.md":    &fstest.MapFile{Data: []byte("# X\n")},
	}

	sub, err := fs.Sub(fsys, "agents")
	if err != nil {
		t.Fatalf("fs.Sub failed: %v", err)
	}

	dst := t.TempDir()
	if err := copyDirFromFS(sub, dst); err != nil {
		t.Fatalf("copyDirFromFS with sub-FS failed: %v", err)
	}

	// Should contain one.md and two.md, but NOT other/x.md
	if _, err := os.Stat(filepath.Join(dst, "one.md")); err != nil {
		t.Error("Expected one.md in output")
	}
	if _, err := os.Stat(filepath.Join(dst, "two.md")); err != nil {
		t.Error("Expected two.md in output")
	}
	if _, err := os.Stat(filepath.Join(dst, "x.md")); err == nil {
		t.Error("Did not expect x.md in output (should be filtered by sub-FS)")
	}
}

func TestLoadRiteManifest_FromEmbedded(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 1.0\nagents:\n  - name: agent-one\n    role: tester\n"),
		},
	}

	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)
	m.sourceResolver.WithEmbeddedFS(fsys)

	resolved := &ResolvedRite{
		Name: "test-rite",
		Source: RiteSource{
			Type: SourceEmbedded,
			Path: "embedded://rites/test-rite",
		},
		RitePath:     "rites/test-rite",
		ManifestPath: "rites/test-rite/manifest.yaml",
		TemplatesDir: "knossos/templates",
	}

	manifest, err := m.loadRiteManifest("rites/test-rite", resolved)
	if err != nil {
		t.Fatalf("loadRiteManifest from embedded failed: %v", err)
	}

	if manifest.Name != "test-rite" {
		t.Errorf("Expected manifest name %q, got %q", "test-rite", manifest.Name)
	}
	if len(manifest.Agents) != 1 {
		t.Fatalf("Expected 1 agent, got %d", len(manifest.Agents))
	}
	if manifest.Agents[0].Name != "agent-one" {
		t.Errorf("Expected agent name %q, got %q", "agent-one", manifest.Agents[0].Name)
	}
}

func TestLoadRiteManifest_FilesystemFallback(t *testing.T) {
	tmpDir := t.TempDir()
	riteDir := filepath.Join(tmpDir, "rites", "fs-rite")
	os.MkdirAll(riteDir, 0755)
	os.WriteFile(filepath.Join(riteDir, "manifest.yaml"),
		[]byte("name: fs-rite\nversion: 1.0\n"), 0644)

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	// No embedded FS, resolved is nil (filesystem path)
	manifest, err := m.loadRiteManifest(riteDir, nil)
	if err != nil {
		t.Fatalf("loadRiteManifest from filesystem failed: %v", err)
	}

	if manifest.Name != "fs-rite" {
		t.Errorf("Expected manifest name %q, got %q", "fs-rite", manifest.Name)
	}
}

func TestMaterializeAgents_FromEmbedded(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 1.0\nagents:\n  - name: agent-one\n    role: tester\n"),
		},
		"rites/test-rite/agents/agent-one.md": &fstest.MapFile{
			Data: []byte("# Agent One\n\nThis is agent one.\n"),
		},
	}

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)
	m.sourceResolver.WithEmbeddedFS(fsys)

	manifest := &RiteManifest{
		Name:    "test-rite",
		Version: "1.0",
		Agents:  []Agent{{Name: "agent-one", Role: "tester"}},
	}

	resolved := &ResolvedRite{
		Name:         "test-rite",
		Source:       RiteSource{Type: SourceEmbedded, Path: "embedded://rites/test-rite"},
		RitePath:     "rites/test-rite",
		ManifestPath: "rites/test-rite/manifest.yaml",
		TemplatesDir: "knossos/templates",
	}

	if err := m.materializeAgents(manifest, "rites/test-rite", claudeDir, resolved, provenance.NullCollector{}); err != nil {
		t.Fatalf("materializeAgents from embedded failed: %v", err)
	}

	// Verify agent file was written
	agentPath := filepath.Join(claudeDir, "agents", "agent-one.md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("Expected agent file at %s, got error: %v", agentPath, err)
	}
	if string(data) != "# Agent One\n\nThis is agent one.\n" {
		t.Errorf("Agent content = %q, want %q", string(data), "# Agent One\n\nThis is agent one.\n")
	}
}

func TestEmbeddedHooks_Fallback(t *testing.T) {
	hooksFS := fstest.MapFS{
		"hooks.yaml": &fstest.MapFile{
			Data: []byte(`schema_version: "2.0"
hooks:
  - event: PreToolUse
    command: "ari hook writeguard --output json"
    priority: 3
`),
		},
	}

	tmpDir := t.TempDir()

	// Ensure no filesystem hooks.yaml is found
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)
	m.embeddedHooks = hooksFS

	cfg := m.loadHooksConfig()
	if cfg == nil {
		t.Fatal("Expected hooks config from embedded fallback, got nil")
	}
	if cfg.SchemaVersion != "2.0" {
		t.Errorf("SchemaVersion = %q, want 2.0", cfg.SchemaVersion)
	}
	if len(cfg.Hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(cfg.Hooks))
	}
	if cfg.Hooks[0].Command != "ari hook writeguard --output json" {
		t.Errorf("Hook command = %q, want writeguard", cfg.Hooks[0].Command)
	}
}

func TestEmbeddedHooks_FilesystemOverrides(t *testing.T) {
	// Create filesystem hooks.yaml at project level
	tmpDir := t.TempDir()
	hooksDir := filepath.Join(tmpDir, "hooks")
	os.MkdirAll(hooksDir, 0755)
	os.WriteFile(filepath.Join(hooksDir, "hooks.yaml"), []byte(`schema_version: "2.0"
hooks:
  - event: Stop
    command: "ari hook autopark --output json"
    priority: 5
`), 0644)

	// Set embedded hooks (different from filesystem)
	embeddedHooksFS := fstest.MapFS{
		"hooks.yaml": &fstest.MapFile{
			Data: []byte(`schema_version: "2.0"
hooks:
  - event: PreToolUse
    command: "ari hook embedded --output json"
    priority: 3
`),
		},
	}

	// Point KNOSSOS_HOME at nonexistent so only project-level is found
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)
	m.embeddedHooks = embeddedHooksFS

	cfg := m.loadHooksConfig()
	if cfg == nil {
		t.Fatal("Expected hooks config, got nil")
	}

	// Filesystem should win -- should have Stop event, not PreToolUse
	if len(cfg.Hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(cfg.Hooks))
	}
	if cfg.Hooks[0].Event != "Stop" {
		t.Errorf("Expected filesystem hook event %q, got %q (embedded leaked through)", "Stop", cfg.Hooks[0].Event)
	}
}

func TestMaterializeMena_FromEmbedded(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 1.0\ndependencies:\n  - shared\n"),
		},
		"rites/test-rite/mena/my-cmd/INDEX.dro.md": &fstest.MapFile{
			Data: []byte("---\nname: my-cmd\n---\n# My Command\n"),
		},
		"rites/shared/mena/shared-ref/INDEX.lego.md": &fstest.MapFile{
			Data: []byte("---\nname: shared-ref\n---\n# Shared Ref\n"),
		},
	}

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", "/nonexistent-knossos-home")
	t.Cleanup(config.ResetKnossosHome)

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)
	m.sourceResolver.WithEmbeddedFS(fsys)

	manifest := &RiteManifest{
		Name:         "test-rite",
		Version:      "1.0",
		Dependencies: []string{"shared"},
	}

	resolved := &ResolvedRite{
		Name:         "test-rite",
		Source:       RiteSource{Type: SourceEmbedded, Path: "embedded://rites/test-rite"},
		RitePath:     "rites/test-rite",
		ManifestPath: "rites/test-rite/manifest.yaml",
		TemplatesDir: "knossos/templates",
	}

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}); err != nil {
		t.Fatalf("materializeMena from embedded failed: %v", err)
	}

	// Verify dromena routed to commands/ with extension stripped
	cmdPath := filepath.Join(claudeDir, "commands", "my-cmd", "INDEX.md")
	if _, err := os.Stat(cmdPath); err != nil {
		t.Errorf("Expected dromena at %s (stripped from INDEX.dro.md), got error: %v", cmdPath, err)
	}

	// Verify legomena routed to skills/ with extension stripped
	skillPath := filepath.Join(claudeDir, "skills", "shared-ref", "INDEX.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("Expected legomena at %s (stripped from INDEX.lego.md), got error: %v", skillPath, err)
	}
}
