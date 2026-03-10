package mena

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// TestCleanEmptyDirs_NonExistentRoot verifies that CleanEmptyDirs returns nil
// errors when called with a path that does not exist on disk.
func TestCleanEmptyDirs_NonExistentRoot(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	errs := CleanEmptyDirs(nonExistent)
	if len(errs) != 0 {
		t.Errorf("CleanEmptyDirs(%q) returned %d errors, want 0: %v", nonExistent, len(errs), errs)
	}
}

// TestCleanEmptyDirs_ExistingEmptySubdir verifies normal behavior:
// empty subdirectories are removed successfully.
func TestCleanEmptyDirs_ExistingEmptySubdir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, "root")
	subdir := filepath.Join(root, "empty-child")

	if err := mkdirAll(subdir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	errs := CleanEmptyDirs(root)
	if len(errs) != 0 {
		t.Errorf("CleanEmptyDirs returned errors: %v", errs)
	}

	// The empty subdir should have been removed
	if exists(subdir) {
		t.Errorf("expected empty subdirectory %q to be removed", subdir)
	}
}

// helpers

func mkdirAll(path string) error {
	return mkdirAllMode(path, 0755)
}

func mkdirAllMode(path string, mode uint32) error {
	return os.MkdirAll(path, os.FileMode(mode))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestIsFromRite verifies the source_path rite origin matching helper.
func TestIsFromRite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		sourcePath string
		riteName   string
		want       bool
	}{
		{"relative match", "rites/10x-dev/mena/commit/", "10x-dev", true},
		{"absolute match", "../../Code/knossos/rites/ecosystem/mena/spike/", "ecosystem", true},
		{"shared rite", "rites/shared/mena/smell-detection/", "10x-dev", false},
		{"different rite", "rites/forge/mena/build-ref/", "10x-dev", false},
		{"platform mena", "mena/operations/commit/", "10x-dev", false},
		{"empty source path", "", "10x-dev", false},
		{"empty rite name", "rites/10x-dev/mena/commit/", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isFromRite(tt.sourcePath, tt.riteName)
			if got != tt.want {
				t.Errorf("isFromRite(%q, %q) = %v, want %v", tt.sourcePath, tt.riteName, got, tt.want)
			}
		})
	}
}

// TestCleanStaleMena_CrossRitePreserved verifies that stale cleanup scoped to
// rite B does not delete entries originating from rite A.
func TestCleanStaleMena_CrossRitePreserved(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("mkdir knossos: %v", err)
	}
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	// Create directories for rite A entries on disk
	riteASkillDir := filepath.Join(skillsDir, "ecosystem-ref")
	if err := os.MkdirAll(riteASkillDir, 0755); err != nil {
		t.Fatalf("mkdir rite A skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(riteASkillDir, "SKILL.md"), []byte("# Ecosystem Ref\n"), 0644); err != nil {
		t.Fatalf("write rite A skill: %v", err)
	}

	riteACmdFile := filepath.Join(commandsDir, "eco-cmd.md")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir commands: %v", err)
	}
	if err := os.WriteFile(riteACmdFile, []byte("# Eco Cmd\n"), 0644); err != nil {
		t.Fatalf("write rite A cmd: %v", err)
	}

	// Create entries for rite B on disk
	riteBSkillDir := filepath.Join(skillsDir, "forge-ref")
	if err := os.MkdirAll(riteBSkillDir, 0755); err != nil {
		t.Fatalf("mkdir rite B skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(riteBSkillDir, "SKILL.md"), []byte("# Forge Ref\n"), 0644); err != nil {
		t.Fatalf("write rite B skill: %v", err)
	}

	// Write a provenance manifest with entries from BOTH rites
	fullChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "ecosystem",
		Entries: map[string]*provenance.ProvenanceEntry{
			// Rite A (ecosystem) entries
			"skills/ecosystem-ref/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/ecosystem/mena/ecosystem-ref",
				"project",
				fullChecksum,
			),
			"commands/eco-cmd": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/ecosystem/mena/eco-cmd",
				"project",
				fullChecksum,
			),
			// Rite B (forge) entries
			"skills/forge-ref/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/forge/mena/forge-ref",
				"project",
				fullChecksum,
			),
		},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Simulate syncing rite B (forge). Only forge-ref is projected.
	result := &MenaProjectionResult{
		SkillsProjected: []string{"forge-ref"},
	}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		KnossosDir:        knossosDir,
		RiteName:          "forge",
	}

	cleanStaleMenaEntries(opts, result)

	// Rite A entries should be PRESERVED (not deleted by forge sync)
	if !exists(filepath.Join(skillsDir, "ecosystem-ref", "SKILL.md")) {
		t.Errorf("rite A skill ecosystem-ref was incorrectly deleted during rite B sync")
	}
	if !exists(filepath.Join(commandsDir, "eco-cmd.md")) {
		t.Errorf("rite A command eco-cmd.md was incorrectly deleted during rite B sync")
	}

	// Rite B entries that are projected should be preserved
	if !exists(filepath.Join(skillsDir, "forge-ref", "SKILL.md")) {
		t.Errorf("projected rite B skill forge-ref was incorrectly deleted")
	}
}

// TestCleanStaleMena_SameRiteStaleRemoved verifies that stale entries from the
// SAME rite are still correctly removed.
func TestCleanStaleMena_SameRiteStaleRemoved(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("mkdir knossos: %v", err)
	}
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	// Create a stale entry from rite A on disk
	staleDir := filepath.Join(skillsDir, "old-skill")
	if err := os.MkdirAll(staleDir, 0755); err != nil {
		t.Fatalf("mkdir stale: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staleDir, "SKILL.md"), []byte("# Old\n"), 0644); err != nil {
		t.Fatalf("write stale: %v", err)
	}
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir commands: %v", err)
	}

	fullChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "ecosystem",
		Entries: map[string]*provenance.ProvenanceEntry{
			"skills/old-skill/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/ecosystem/mena/old-skill",
				"project",
				fullChecksum,
			),
		},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Sync ecosystem with no skills projected (old-skill was renamed/removed)
	result := &MenaProjectionResult{}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		KnossosDir:        knossosDir,
		RiteName:          "ecosystem",
	}

	cleanStaleMenaEntries(opts, result)

	// The stale entry from the same rite SHOULD be removed
	if exists(filepath.Join(skillsDir, "old-skill", "SKILL.md")) {
		t.Errorf("stale same-rite entry old-skill should have been removed")
	}
}
