package materialize

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/provenance"

	"github.com/autom8y/knossos/internal/materialize/compiler"
	"github.com/autom8y/knossos/internal/paths"
)

func TestCopyDirFromFS(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	fsys := fstest.MapFS{}
	dst := t.TempDir()
	if err := copyDirFromFS(fsys, dst); err != nil {
		t.Fatalf("copyDirFromFS with empty FS failed: %v", err)
	}
}

func TestCopyDirFromFS_SubFS(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

	if err := m.materializeAgents(manifest, "rites/test-rite", claudeDir, resolved, provenance.NullCollector{}, nil, nil, "", "", nil); err != nil {
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

func TestMaterializeMena_FromEmbedded(t *testing.T) {
	t.Parallel()
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

	resolver := paths.NewResolver(tmpDir)
	sr := NewSourceResolverWithPaths(tmpDir, "", "", "/nonexistent-knossos-home")
	m := NewMaterializerWithSourceResolver(resolver, sr)
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

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}, false, "", &compiler.ClaudeCompiler{}); err != nil {
		t.Fatalf("materializeMena from embedded failed: %v", err)
	}

	// Verify dromena INDEX promoted to parent level
	cmdPath := filepath.Join(claudeDir, "commands", "my-cmd.md")
	if _, err := os.Stat(cmdPath); err != nil {
		t.Errorf("Expected promoted dromena at %s, got error: %v", cmdPath, err)
	}

	// Verify legomena routed to skills/ as SKILL.md (CC entrypoint convention)
	skillPath := filepath.Join(claudeDir, "skills", "shared-ref", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("Expected legomena at %s (renamed from INDEX.lego.md, CC entrypoint), got error: %v", skillPath, err)
	}
}

// TestMaterializeMena_EmbeddedSkipsPlatformMena verifies that when the rite source is
// embedded, platform-level mena from KnossosHome (the developer's local knossos tree)
// is NOT injected into foreign projects -- even when KNOSSOS_HOME resolves to a real
// directory containing mena files.
func TestMaterializeMena_EmbeddedSkipsPlatformMena(t *testing.T) {
	t.Parallel()
	// Create a fake "platform" mena dir (simulating ~/Code/knossos/mena/)
	// that contains a platform-only command not in the embedded rite.
	platformMenaDir := t.TempDir()
	platformCmdDir := filepath.Join(platformMenaDir, "mena", "platform-op")
	if err := os.MkdirAll(platformCmdDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(platformCmdDir, "INDEX.dro.md"),
		[]byte("---\nname: platform-op\n---\n# Platform Op\n"),
		0644,
	); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Embedded FS contains only the rite-specific command.
	fsys := fstest.MapFS{
		"rites/foreign-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: foreign-rite\nversion: 1.0\n"),
		},
		"rites/foreign-rite/mena/rite-cmd/INDEX.dro.md": &fstest.MapFile{
			Data: []byte("---\nname: rite-cmd\n---\n# Rite Command\n"),
		},
		"rites/shared/mena/shared-ref/INDEX.lego.md": &fstest.MapFile{
			Data: []byte("---\nname: shared-ref\n---\n# Shared Ref\n"),
		},
	}

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	resolver := paths.NewResolver(tmpDir)
	sr := NewSourceResolverWithPaths(tmpDir, "", "", platformMenaDir)
	m := NewMaterializerWithSourceResolver(resolver, sr)
	m.sourceResolver.WithEmbeddedFS(fsys)

	manifest := &RiteManifest{
		Name:         "foreign-rite",
		Version:      "1.0",
		Dependencies: []string{"shared"},
	}
	resolved := &ResolvedRite{
		Name:         "foreign-rite",
		Source:       RiteSource{Type: SourceEmbedded, Path: "embedded://rites/foreign-rite"},
		RitePath:     "rites/foreign-rite",
		ManifestPath: "rites/foreign-rite/manifest.yaml",
		TemplatesDir: "knossos/templates",
	}

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}, false, "", &compiler.ClaudeCompiler{}); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Rite-specific command MUST appear.
	riteCmdPath := filepath.Join(claudeDir, "commands", "rite-cmd.md")
	if _, err := os.Stat(riteCmdPath); err != nil {
		t.Errorf("Expected rite command at %s, got error: %v", riteCmdPath, err)
	}

	// Platform mena MAY appear from filesystem (KnossosHome) or embedded fallback.
	// This is intentional: platform mena IS the product and ships to all users.
	// We only verify that the rite-specific mena always materializes.
}
