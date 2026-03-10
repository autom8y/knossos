package mena

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// TestResolveNamespace_DirAndFileDeduplicated verifies that when a command exists
// as both a directory (companions) and a promoted .md file on disk, the Step 4
// collision check processes it only once, producing at most one warning.
func TestResolveNamespace_DirAndFileDeduplicated(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	commandsDir := filepath.Join(claudeDir, "commands")

	// Create both forms on disk: commands/my-cmd/ (dir) and commands/my-cmd.md (promoted)
	if err := os.MkdirAll(filepath.Join(commandsDir, "my-cmd"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "my-cmd", "behavior.md"), []byte("companion"), 0644); err != nil {
		t.Fatalf("write companion: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "my-cmd.md"), []byte("promoted INDEX"), 0644); err != nil {
		t.Fatalf("write promoted: %v", err)
	}

	// No provenance manifest -> entries are treated as user-owned/untracked.
	// This triggers the warning path in Step 4.

	// Build collected entries: a single dromenon wanting flat name "my-cmd"
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "group", "my-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("mkdir dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\ndescription: test\n---\n# Cmd\n"), 0644); err != nil {
		t.Fatalf("write INDEX: %v", err)
	}

	collected := map[string]menaCollectedEntry{
		"group/my-cmd": {
			source:      MenaSource{Path: droDir},
			name:        "my-cmd",
			sourceIndex: 0,
			menaType:    "dro",
		},
	}

	opts := MenaProjectionOptions{
		TargetCommandsDir: commandsDir,
		OverwriteDiverged: false,
	}

	_, warnings := resolveNamespace(collected, nil, opts)

	// Count warnings mentioning "my-cmd"
	count := 0
	for _, w := range warnings {
		if strings.Contains(w, "my-cmd") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 warning for 'my-cmd', got %d; warnings: %v", count, warnings)
	}
}

// TestResolveNamespace_DirAndFileDeduplicated_KnossosOwned verifies that when
// both dir and file forms exist but are knossos-owned, neither produces a warning
// and the dedup does not introduce regressions.
func TestResolveNamespace_DirAndFileDeduplicated_KnossosOwned(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	knossosDir := filepath.Join(tmpDir, ".knossos")
	commandsDir := filepath.Join(claudeDir, "commands")

	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("mkdir knossos: %v", err)
	}

	// Create both forms on disk
	if err := os.MkdirAll(filepath.Join(commandsDir, "my-cmd"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "my-cmd.md"), []byte("promoted"), 0644); err != nil {
		t.Fatalf("write promoted: %v", err)
	}

	// Provenance marks it as knossos-owned
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "test",
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/my-cmd/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/test/mena/my-cmd",
				"project",
				"sha256:0000000000000000000000000000000000000000000000000000000000000000",
			),
		},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	droDir := filepath.Join(tmpDir, "mena", "my-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("mkdir dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\ndescription: test\n---\n# Cmd\n"), 0644); err != nil {
		t.Fatalf("write INDEX: %v", err)
	}

	collected := map[string]menaCollectedEntry{
		"my-cmd": {
			source:      MenaSource{Path: droDir},
			name:        "my-cmd",
			sourceIndex: 0,
			menaType:    "dro",
		},
	}

	opts := MenaProjectionOptions{
		TargetCommandsDir: commandsDir,
		KnossosDir:        knossosDir,
		OverwriteDiverged: false,
	}

	flatNames, warnings := resolveNamespace(collected, nil, opts)

	if len(warnings) != 0 {
		t.Errorf("expected no warnings for knossos-owned entry, got %v", warnings)
	}
	if flatNames["my-cmd"] != "my-cmd" {
		t.Errorf("expected flat name 'my-cmd' to be preserved, got %v", flatNames)
	}
}
