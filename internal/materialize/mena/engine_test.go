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
		name		string
		sourcePath	string
		riteName	string
		want		bool
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

// TestIsFromActiveChain verifies the dependency chain matching helper.
func TestIsFromActiveChain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		sourcePath string
		riteName   string
		deps       []string
		want       bool
	}{
		{"current rite", "rites/security/mena/sec-ref", "security", nil, true},
		{"shared dependency", "rites/shared/mena/smell-detection", "security", []string{"shared"}, true},
		{"explicit dependency", "rites/releaser/mena/release", "security", []string{"shared", "releaser"}, true},
		{"not in chain", "rites/forge/mena/forge-rite", "security", []string{"shared"}, false},
		{"platform mena", "mena/operations/commit/", "security", []string{"shared"}, true},
		{"procession mena", ".knossos/procession-mena/sec-rem", "security", nil, true},
		{"empty deps foreign rite", "rites/forge/mena/build", "security", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isFromActiveChain(tt.sourcePath, tt.riteName, tt.deps)
			if got != tt.want {
				t.Errorf("isFromActiveChain(%q, %q, %v) = %v, want %v",
					tt.sourcePath, tt.riteName, tt.deps, got, tt.want)
			}
		})
	}
}

// TestReconcileUntrackedEntries verifies that untracked mena entries with
// knossos frontmatter are cleaned, while user-created entries are preserved.
func TestReconcileUntrackedEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("mkdir knossos: %v", err)
	}
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir commands: %v", err)
	}
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}

	// Create an untracked knossos command (has mena frontmatter)
	knossosCmd := "---\nname: stale-cmd\ndescription: Legacy command\n---\n# Body\n"
	if err := os.WriteFile(filepath.Join(commandsDir, "stale-cmd.md"), []byte(knossosCmd), 0644); err != nil {
		t.Fatalf("write knossos cmd: %v", err)
	}

	// Create a user-created command (no mena frontmatter)
	userCmd := "# My Custom Command\nDo the thing.\n"
	if err := os.WriteFile(filepath.Join(commandsDir, "my-custom.md"), []byte(userCmd), 0644); err != nil {
		t.Fatalf("write user cmd: %v", err)
	}

	// Create a .gitkeep (should be ignored)
	if err := os.WriteFile(filepath.Join(commandsDir, ".gitkeep"), []byte(""), 0644); err != nil {
		t.Fatalf("write gitkeep: %v", err)
	}

	// Empty provenance manifest (nothing tracked)
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "security",
		Entries:       map[string]*provenance.ProvenanceEntry{},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	result := &MenaProjectionResult{}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		KnossosDir:        knossosDir,
		RiteName:          "security",
	}

	reconcileUntrackedEntries(opts, result)

	// Knossos-formatted command should be removed
	if exists(filepath.Join(commandsDir, "stale-cmd.md")) {
		t.Errorf("untracked knossos command stale-cmd.md should have been removed")
	}

	// User-created command should be preserved
	if !exists(filepath.Join(commandsDir, "my-custom.md")) {
		t.Errorf("user-created command my-custom.md should have been preserved")
	}

	// .gitkeep should be preserved
	if !exists(filepath.Join(commandsDir, ".gitkeep")) {
		t.Errorf(".gitkeep should have been preserved")
	}
}

// TestCleanStaleMena_CrossRiteCleaned verifies that stale entries from rites
// NOT in the active dependency chain are cleaned on rite switch.
func TestCleanStaleMena_CrossRiteCleaned(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("mkdir knossos: %v", err)
	}
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	// Create entries from ecosystem rite on disk
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

	// Create entries for forge rite on disk
	riteBSkillDir := filepath.Join(skillsDir, "forge-ref")
	if err := os.MkdirAll(riteBSkillDir, 0755); err != nil {
		t.Fatalf("mkdir rite B skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(riteBSkillDir, "SKILL.md"), []byte("# Forge Ref\n"), 0644); err != nil {
		t.Fatalf("write rite B skill: %v", err)
	}

	// Write provenance with entries from BOTH rites
	fullChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion:	provenance.CurrentSchemaVersion,
		LastSync:	time.Now().UTC(),
		ActiveRite:	"ecosystem",
		Entries: map[string]*provenance.ProvenanceEntry{
			"skills/ecosystem-ref/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/ecosystem/mena/ecosystem-ref",
				"project",
				fullChecksum, "",
			),
			"commands/eco-cmd": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/ecosystem/mena/eco-cmd",
				"project",
				fullChecksum, "",
			),
			"skills/forge-ref/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/forge/mena/forge-ref",
				"project",
				fullChecksum, "",
			),
		},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Sync forge rite with no ecosystem dependency — ecosystem entries should be cleaned
	result := &MenaProjectionResult{
		SkillsProjected: []string{"forge-ref"},
	}
	opts := MenaProjectionOptions{
		Mode:			MenaProjectionDestructive,
		TargetCommandsDir:	commandsDir,
		TargetSkillsDir:	skillsDir,
		KnossosDir:		knossosDir,
		RiteName:		"forge",
		ActiveDeps:		[]string{"shared"},
	}

	cleanStaleMenaEntries(opts, result)

	// Ecosystem entries should be CLEANED (not in forge's dependency chain)
	if exists(filepath.Join(skillsDir, "ecosystem-ref", "SKILL.md")) {
		t.Errorf("cross-rite entry ecosystem-ref should have been cleaned on rite switch")
	}
	if exists(filepath.Join(commandsDir, "eco-cmd.md")) {
		t.Errorf("cross-rite command eco-cmd.md should have been cleaned on rite switch")
	}

	// Projected forge entries should be preserved
	if !exists(filepath.Join(skillsDir, "forge-ref", "SKILL.md")) {
		t.Errorf("projected forge-ref was incorrectly deleted")
	}
}

// TestCleanStaleMena_DependencyPreserved verifies that entries from rites
// in the active dependency chain are preserved.
func TestCleanStaleMena_DependencyPreserved(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("mkdir knossos: %v", err)
	}
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	// Create entries from shared rite (dependency)
	sharedSkillDir := filepath.Join(skillsDir, "smell-detection")
	if err := os.MkdirAll(sharedSkillDir, 0755); err != nil {
		t.Fatalf("mkdir shared skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedSkillDir, "SKILL.md"), []byte("# Smell\n"), 0644); err != nil {
		t.Fatalf("write shared skill: %v", err)
	}
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir commands: %v", err)
	}

	fullChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion:	provenance.CurrentSchemaVersion,
		LastSync:	time.Now().UTC(),
		ActiveRite:	"10x-dev",
		Entries: map[string]*provenance.ProvenanceEntry{
			"skills/smell-detection/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/shared/mena/smell-detection",
				"shared",
				fullChecksum, "",
			),
		},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Sync security rite — SyncMena always projects all active sources,
	// so shared entries appear in the projected set.
	result := &MenaProjectionResult{
		SkillsProjected: []string{"smell-detection"},
	}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		KnossosDir:        knossosDir,
		RiteName:          "security",
		ActiveDeps:        []string{"shared"},
	}

	cleanStaleMenaEntries(opts, result)

	// Shared entries should be PRESERVED (in projected set from SyncMena)
	if !exists(filepath.Join(skillsDir, "smell-detection", "SKILL.md")) {
		t.Errorf("projected shared dependency smell-detection should be preserved")
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
		SchemaVersion:	provenance.CurrentSchemaVersion,
		LastSync:	time.Now().UTC(),
		ActiveRite:	"ecosystem",
		Entries: map[string]*provenance.ProvenanceEntry{
			"skills/old-skill/": provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/ecosystem/mena/old-skill",
				"project",
				fullChecksum, "",
			),
		},
	}
	if err := provenance.Save(filepath.Join(knossosDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Sync ecosystem with no skills projected (old-skill was renamed/removed)
	result := &MenaProjectionResult{}
	opts := MenaProjectionOptions{
		Mode:			MenaProjectionDestructive,
		TargetCommandsDir:	commandsDir,
		TargetSkillsDir:	skillsDir,
		KnossosDir:		knossosDir,
		RiteName:		"ecosystem",
	}

	cleanStaleMenaEntries(opts, result)

	// The stale entry from the same rite SHOULD be removed
	if exists(filepath.Join(skillsDir, "old-skill", "SKILL.md")) {
		t.Errorf("stale same-rite entry old-skill should have been removed")
	}
}
