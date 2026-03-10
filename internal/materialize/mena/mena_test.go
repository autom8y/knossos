package mena

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// TestStripMenaExtension verifies all extension stripping cases.
func TestStripMenaExtension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"dro.md stripped", "INDEX.dro.md", "INDEX.md"},
		{"lego.md stripped", "INDEX.lego.md", "INDEX.md"},
		{"standalone dro stripped", "commit.dro.md", "commit.md"},
		{"standalone lego stripped", "prompting.lego.md", "prompting.md"},
		{"plain md unchanged", "helper.md", "helper.md"},
		{"plain INDEX unchanged", "INDEX.md", "INDEX.md"},
		{"readme unchanged", "README.md", "README.md"},
		{"non-md unchanged", "data.json", "data.json"},
		{"double infix strips first only", "foo.dro.dro.md", "foo.dro.md"},
		{"double lego strips first only", "bar.lego.lego.md", "bar.lego.md"},
		{"no extension at all", "Makefile", "Makefile"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StripMenaExtension(tt.input)
			if got != tt.expected {
				t.Errorf("StripMenaExtension(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestRouteMenaFile verifies routing decisions.
func TestRouteMenaFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"dro routes to commands", "INDEX.dro.md", "commands"},
		{"lego routes to skills", "INDEX.lego.md", "skills"},
		{"plain md defaults to commands", "INDEX.md", "commands"},
		{"standalone dro routes to commands", "commit.dro.md", "commands"},
		{"standalone lego routes to skills", "prompting.lego.md", "skills"},
		{"plain file defaults to commands", "helper.md", "commands"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RouteMenaFile(tt.input)
			if got != tt.expected {
				t.Errorf("RouteMenaFile(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestSyncMena_Destructive verifies full projection with destructive mode.
func TestSyncMena_Destructive(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "my-cmd")
	legoDir := filepath.Join(menaDir, "my-ref")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("Failed to create dro dir: %v", err)
	}
	if err := os.MkdirAll(legoDir, 0755); err != nil {
		t.Fatalf("Failed to create lego dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("# My Command\n"), 0644); err != nil {
		t.Fatalf("Failed to write dro INDEX: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "helper.md"), []byte("# Helper\n"), 0644); err != nil {
		t.Fatalf("Failed to write helper: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte("# My Ref\n"), 0644); err != nil {
		t.Fatalf("Failed to write lego INDEX: %v", err)
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "stale.md"), []byte("stale"), 0644); err != nil {
		t.Fatalf("Failed to write stale file: %v", err)
	}

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(commandsDir, "stale.md")); os.IsNotExist(err) {
		t.Errorf("Selective write should preserve user-created stale.md")
	}

	cmdPromoted := filepath.Join(commandsDir, "my-cmd.md")
	if _, err := os.Stat(cmdPromoted); os.IsNotExist(err) {
		t.Errorf("Expected promoted dromena at %s", cmdPromoted)
	}

	cmdOldIndex := filepath.Join(commandsDir, "my-cmd", "INDEX.md")
	if _, err := os.Stat(cmdOldIndex); err == nil {
		t.Errorf("INDEX.md should not exist in subdirectory")
	}

	cmdHelper := filepath.Join(commandsDir, "my-cmd", "helper.md")
	if _, err := os.Stat(cmdHelper); os.IsNotExist(err) {
		t.Errorf("Expected helper.md at %s", cmdHelper)
	}

	skillEntrypoint := filepath.Join(skillsDir, "my-ref", "SKILL.md")
	if _, err := os.Stat(skillEntrypoint); os.IsNotExist(err) {
		t.Errorf("Expected legomena entrypoint at %s (CC expects SKILL.md, not INDEX.md)", skillEntrypoint)
	}

	// Verify INDEX.md was NOT produced (CC does not read it as skill entrypoint)
	skillOldIndex := filepath.Join(skillsDir, "my-ref", "INDEX.md")
	if _, err := os.Stat(skillOldIndex); err == nil {
		t.Errorf("INDEX.md must not exist at %s; legomena entrypoint must be SKILL.md", skillOldIndex)
	}

	if len(result.CommandsProjected) == 0 {
		t.Errorf("Expected CommandsProjected to be non-empty")
	}
	if len(result.SkillsProjected) == 0 {
		t.Errorf("Expected SkillsProjected to be non-empty")
	}
}

// TestSyncMena_Additive verifies additive mode preserves existing files.
func TestSyncMena_Additive(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	userFile := filepath.Join(commandsDir, "user-created.md")
	if err := os.WriteFile(userFile, []byte("user content"), 0644); err != nil {
		t.Fatalf("Failed to write user file: %v", err)
	}

	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "new-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("Failed to create dro dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("# New Command\n"), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.dro.md: %v", err)
	}

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionAdditive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena additive failed: %v", err)
	}

	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Errorf("Additive mode should preserve user-created.md")
	}

	newCmd := filepath.Join(commandsDir, "new-cmd.md")
	if _, err := os.Stat(newCmd); os.IsNotExist(err) {
		t.Errorf("Expected promoted command at %s", newCmd)
	}
}

// TestSyncMena_PriorityOverride verifies later sources override earlier ones.
func TestSyncMena_PriorityOverride(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	lowDir := filepath.Join(tmpDir, "low", "my-cmd")
	if err := os.MkdirAll(lowDir, 0755); err != nil {
		t.Fatalf("Failed to create low dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lowDir, "INDEX.dro.md"), []byte("low-priority content\n"), 0644); err != nil {
		t.Fatalf("Failed to write low INDEX: %v", err)
	}

	highDir := filepath.Join(tmpDir, "high", "my-cmd")
	if err := os.MkdirAll(highDir, 0755); err != nil {
		t.Fatalf("Failed to create high dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(highDir, "INDEX.dro.md"), []byte("high-priority content\n"), 0644); err != nil {
		t.Fatalf("Failed to write high INDEX: %v", err)
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{
		{Path: filepath.Join(tmpDir, "low")},
		{Path: filepath.Join(tmpDir, "high")},
	}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena priority override failed: %v", err)
	}

	cmdPromoted := filepath.Join(commandsDir, "my-cmd.md")
	content, err := os.ReadFile(cmdPromoted)
	if err != nil {
		t.Fatalf("Failed to read promoted command file: %v", err)
	}

	if string(content) != "high-priority content\n" {
		t.Errorf("Expected high-priority content, got %q", string(content))
	}
}

// TestSyncMena_EmbeddedFS verifies projection from an embedded FS source.
func TestSyncMena_EmbeddedFS(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	fsys := fstest.MapFS{
		"rites/test-rite/mena/my-cmd/INDEX.dro.md": &fstest.MapFile{
			Data: []byte("---\nname: my-cmd\n---\n# Embedded Command\n"),
		},
		"rites/shared/mena/my-ref/INDEX.lego.md": &fstest.MapFile{
			Data: []byte("---\nname: my-ref\n---\n# Embedded Ref\n"),
		},
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{
		{Fsys: fsys, FsysPath: "rites/shared/mena", IsEmbedded: true},
		{Fsys: fsys, FsysPath: "rites/test-rite/mena", IsEmbedded: true},
	}

	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena embedded failed: %v", err)
	}

	cmdPromoted := filepath.Join(commandsDir, "my-cmd.md")
	if _, err := os.Stat(cmdPromoted); os.IsNotExist(err) {
		t.Errorf("Expected promoted dromena at %s", cmdPromoted)
	}

	skillEntrypoint := filepath.Join(skillsDir, "my-ref", "SKILL.md")
	if _, err := os.Stat(skillEntrypoint); os.IsNotExist(err) {
		t.Errorf("Expected embedded legomena entrypoint at %s (CC expects SKILL.md, not INDEX.md)", skillEntrypoint)
	}

	// Verify INDEX.md was NOT produced (CC does not read it as skill entrypoint)
	skillOldIndex := filepath.Join(skillsDir, "my-ref", "INDEX.md")
	if _, err := os.Stat(skillOldIndex); err == nil {
		t.Errorf("INDEX.md must not exist at %s; legomena entrypoint must be SKILL.md", skillOldIndex)
	}

	if len(result.CommandsProjected) == 0 {
		t.Errorf("Expected CommandsProjected to be non-empty")
	}
	if len(result.SkillsProjected) == 0 {
		t.Errorf("Expected SkillsProjected to be non-empty")
	}
}

// TestSyncMena_Filter_DroOnly verifies that ProjectDro filter only projects dromena.
func TestSyncMena_Filter_DroOnly(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "cmd1")
	legoDir := filepath.Join(menaDir, "ref1")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("Failed to create dro dir: %v", err)
	}
	if err := os.MkdirAll(legoDir, 0755); err != nil {
		t.Fatalf("Failed to create lego dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("# Cmd\n"), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte("# Ref\n"), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectDro,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(commandsDir, "cmd1.md")); os.IsNotExist(err) {
		t.Errorf("Expected cmd1.md to be projected to commands/")
	}

	if _, err := os.Stat(filepath.Join(skillsDir, "ref1", "SKILL.md")); !os.IsNotExist(err) {
		t.Errorf("Expected ref1 to NOT be projected when filter is ProjectDro")
	}

	if len(result.SkillsProjected) != 0 {
		t.Errorf("Expected no skills projected, got %v", result.SkillsProjected)
	}
}

// TestSyncMena_StandaloneFileStripping verifies standalone files have extensions stripped.
func TestSyncMena_StandaloneFileStripping(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	menaDir := filepath.Join(tmpDir, "mena")
	groupDir := filepath.Join(menaDir, "navigation")
	if err := os.MkdirAll(groupDir, 0755); err != nil {
		t.Fatalf("Failed to create group dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(groupDir, "rite.dro.md"), []byte("# Rite Navigation\n"), 0644); err != nil {
		t.Fatalf("Failed to write standalone dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(groupDir, "reference.lego.md"), []byte("# Reference\n"), 0644); err != nil {
		t.Fatalf("Failed to write standalone lego: %v", err)
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena failed: %v", err)
	}

	droStripped := filepath.Join(commandsDir, "navigation", "rite.md")
	if _, err := os.Stat(droStripped); os.IsNotExist(err) {
		t.Errorf("Expected standalone dro at %s", droStripped)
	}
	legoStripped := filepath.Join(skillsDir, "navigation", "reference.md")
	if _, err := os.Stat(legoStripped); os.IsNotExist(err) {
		t.Errorf("Expected standalone lego at %s", legoStripped)
	}
}

// TestSyncMena_EmptySources verifies graceful handling of empty sources.
func TestSyncMena_EmptySources(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena with empty sources should not fail: %v", err)
	}

	if len(result.CommandsProjected) != 0 || len(result.SkillsProjected) != 0 {
		t.Errorf("Expected no projections from empty sources")
	}
}

// TestSyncMena_NonexistentSource verifies graceful handling of missing sources.
func TestSyncMena_NonexistentSource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: filepath.Join(tmpDir, "nonexistent-mena")}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena with nonexistent source should not fail: %v", err)
	}

	if len(result.CommandsProjected) != 0 || len(result.SkillsProjected) != 0 {
		t.Errorf("Expected no projections from nonexistent source")
	}
}

// TestParseMenaFrontmatterBytes verifies frontmatter extraction from raw bytes.
func TestParseMenaFrontmatterBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		data     string
		wantName string
	}{
		{"with frontmatter", "---\nname: test\ndescription: d\n---\n# Body\n", "test"},
		{"no frontmatter delimiters", "# Just a file with no frontmatter\n", ""},
		{"malformed YAML", "---\n: [invalid yaml\n---\n# Body\n", ""},
		{"empty content", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := ParseMenaFrontmatterBytes([]byte(tt.data))
			if fm.Name != tt.wantName {
				t.Errorf("name = %q, want %q", fm.Name, tt.wantName)
			}
		})
	}
}

// TestDetectMenaType verifies the extension-based type detection
func TestDetectMenaType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename string
		expected string
	}{
		{"INDEX.dro.md", "dro"},
		{"INDEX.lego.md", "lego"},
		{"INDEX.md", "dro"},
		{"commit.dro.md", "dro"},
		{"standards.lego.md", "lego"},
		{"behavior.md", "dro"},
		{"README.md", "dro"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			got := DetectMenaType(tt.filename)
			if got != tt.expected {
				t.Errorf("DetectMenaType(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}

// TestSyncMena_NamespaceCollision_YieldsToUserEntry verifies that a flat name
// collision with a user-owned entry in commands/ causes the dromenon to fall back
// to its source path instead of overwriting the user file.
func TestSyncMena_NamespaceCollision_YieldsToUserEntry(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir // provenance manifest lives here
	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create a user-owned entry at the flat path
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "my-cmd.md"), []byte("user content"), 0644); err != nil {
		t.Fatalf("write user file: %v", err)
	}

	// Write a provenance manifest marking it as user-owned
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "test",
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/my-cmd.md": provenance.NewUserEntry(provenance.ScopeRite, "sha256:0000000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	if err := provenance.Save(filepath.Join(claudeDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Create a mena source with a dromenon that wants flat name "my-cmd"
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "group", "my-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("mkdir dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\ndescription: test\n---\n# Platform cmd\n"), 0644); err != nil {
		t.Fatalf("write INDEX: %v", err)
	}

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectDro,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		OverwriteDiverged: false, // default: yield to user
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena failed: %v", err)
	}

	// The flat path should NOT be overwritten (user file preserved)
	content, err := os.ReadFile(filepath.Join(commandsDir, "my-cmd.md"))
	if err != nil {
		t.Fatalf("read user file: %v", err)
	}
	if string(content) != "user content" {
		t.Errorf("user file was overwritten; got %q, want %q", string(content), "user content")
	}

	// The dromenon should have fallen back to source path
	found := slices.Contains(result.CommandsProjected, "group/my-cmd")
	if !found {
		t.Errorf("expected dromenon to fall back to source path 'group/my-cmd', projected: %v", result.CommandsProjected)
	}
}

// TestSyncMena_NamespaceCollision_OverwriteDivergedReclaims verifies that
// --overwrite-diverged allows the platform to reclaim a flat name from a
// user-owned entry.
func TestSyncMena_NamespaceCollision_OverwriteDivergedReclaims(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := tmpDir
	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create a user-owned entry at the flat path
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "my-cmd.md"), []byte("user content"), 0644); err != nil {
		t.Fatalf("write user file: %v", err)
	}

	// Write provenance manifest marking it user-owned
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		ActiveRite:    "test",
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/my-cmd.md": provenance.NewUserEntry(provenance.ScopeRite, "sha256:0000000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	if err := provenance.Save(filepath.Join(claudeDir, provenance.ManifestFileName), manifest); err != nil {
		t.Fatalf("save provenance: %v", err)
	}

	// Create a mena source with a dromenon that wants flat name "my-cmd"
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "group", "my-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("mkdir dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\ndescription: test\n---\n# Platform cmd\n"), 0644); err != nil {
		t.Fatalf("write INDEX: %v", err)
	}

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectDro,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		OverwriteDiverged: true, // reclaim flat name
	}

	result, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena failed: %v", err)
	}

	// The flat path should now contain the platform content (overwritten)
	content, err := os.ReadFile(filepath.Join(commandsDir, "my-cmd.md"))
	if err != nil {
		t.Fatalf("read reclaimed file: %v", err)
	}
	if string(content) == "user content" {
		t.Errorf("user file was NOT overwritten; OverwriteDiverged should have reclaimed the flat name")
	}

	// The dromenon should be projected at the flat name, not the source path
	foundFlat := slices.Contains(result.CommandsProjected, "my-cmd")
	if !foundFlat {
		t.Errorf("expected dromenon at flat name 'my-cmd', projected: %v", result.CommandsProjected)
	}
}

// TestSyncMena_NamespaceCollision_UntrackedEntry verifies that entries on disk
// without provenance (untracked) are also treated as user content and yield,
// unless OverwriteDiverged is set.
func TestSyncMena_NamespaceCollision_UntrackedEntry(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create an entry on disk with NO provenance manifest at all
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "my-cmd.md"), []byte("untracked content"), 0644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}

	// Create a mena source with a dromenon that wants flat name "my-cmd"
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "group", "my-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("mkdir dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\ndescription: test\n---\n# Platform cmd\n"), 0644); err != nil {
		t.Fatalf("write INDEX: %v", err)
	}

	sources := []MenaSource{{Path: menaDir}}

	// Without OverwriteDiverged: should yield to untracked entry
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectDro,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		OverwriteDiverged: false,
	}

	_, err := SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(commandsDir, "my-cmd.md"))
	if string(content) != "untracked content" {
		t.Errorf("untracked file was overwritten without --overwrite-diverged")
	}

	// With OverwriteDiverged: should reclaim
	opts.OverwriteDiverged = true
	_, err = SyncMena(sources, opts)
	if err != nil {
		t.Fatalf("SyncMena with OverwriteDiverged failed: %v", err)
	}

	content, _ = os.ReadFile(filepath.Join(commandsDir, "my-cmd.md"))
	if string(content) == "untracked content" {
		t.Errorf("untracked file was NOT overwritten with --overwrite-diverged")
	}
}
