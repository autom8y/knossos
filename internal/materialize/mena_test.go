package materialize

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestStripMenaExtension verifies all extension stripping cases.
func TestStripMenaExtension(t *testing.T) {
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
			got := StripMenaExtension(tt.input)
			if got != tt.expected {
				t.Errorf("StripMenaExtension(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestRouteMenaFile verifies routing decisions.
func TestRouteMenaFile(t *testing.T) {
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
			got := RouteMenaFile(tt.input)
			if got != tt.expected {
				t.Errorf("RouteMenaFile(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestSyncMena_Destructive verifies full projection with destructive mode:
// selectively replaces managed entries while preserving user-created content.
func TestSyncMena_Destructive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mena source with a dromena and a legomena
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

	// Create pre-existing files that should be wiped
	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "stale.md"), []byte("stale"), 0644); err != nil {
		t.Fatalf("Failed to write stale file: %v", err)
	}

	sources := []MenaSource{
		{Path: menaDir},
	}
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

	// Verify user-created file is preserved (selective write, not destructive nuke)
	if _, err := os.Stat(filepath.Join(commandsDir, "stale.md")); os.IsNotExist(err) {
		t.Errorf("Selective write should preserve user-created stale.md, but it was deleted")
	}

	// Verify dromena projected to commands/ with stripped names
	cmdIndex := filepath.Join(commandsDir, "my-cmd", "INDEX.md")
	if _, err := os.Stat(cmdIndex); os.IsNotExist(err) {
		t.Errorf("Expected dromena INDEX.md (stripped) at %s, but it does not exist", cmdIndex)
	}
	cmdHelper := filepath.Join(commandsDir, "my-cmd", "helper.md")
	if _, err := os.Stat(cmdHelper); os.IsNotExist(err) {
		t.Errorf("Expected helper.md at %s, but it does not exist", cmdHelper)
	}

	// Verify un-stripped name does NOT exist
	cmdOld := filepath.Join(commandsDir, "my-cmd", "INDEX.dro.md")
	if _, err := os.Stat(cmdOld); err == nil {
		t.Errorf("INDEX.dro.md should not exist in output (should be stripped to INDEX.md)")
	}

	// Verify legomena projected to skills/ with stripped names
	skillIndex := filepath.Join(skillsDir, "my-ref", "INDEX.md")
	if _, err := os.Stat(skillIndex); os.IsNotExist(err) {
		t.Errorf("Expected legomena INDEX.md (stripped) at %s, but it does not exist", skillIndex)
	}

	// Verify un-stripped name does NOT exist
	skillOld := filepath.Join(skillsDir, "my-ref", "INDEX.lego.md")
	if _, err := os.Stat(skillOld); err == nil {
		t.Errorf("INDEX.lego.md should not exist in output (should be stripped to INDEX.md)")
	}

	// Verify result tracking
	if len(result.CommandsProjected) == 0 {
		t.Errorf("Expected CommandsProjected to be non-empty")
	}
	if len(result.SkillsProjected) == 0 {
		t.Errorf("Expected SkillsProjected to be non-empty")
	}
}

// TestSyncMena_Additive verifies additive mode preserves existing files.
func TestSyncMena_Additive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create pre-existing files that should be preserved
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

	// Create mena source
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "new-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("Failed to create dro dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("# New Command\n"), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.dro.md: %v", err)
	}

	sources := []MenaSource{
		{Path: menaDir},
	}
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

	// Verify user-created file is preserved
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Errorf("Additive mode should preserve user-created.md, but it was deleted")
	}

	// Verify new command was projected with stripped name
	newCmd := filepath.Join(commandsDir, "new-cmd", "INDEX.md")
	if _, err := os.Stat(newCmd); os.IsNotExist(err) {
		t.Errorf("Expected new command at %s, but it does not exist", newCmd)
	}
}

// TestSyncMena_PriorityOverride verifies that later sources override earlier
// sources for the same mena name.
func TestSyncMena_PriorityOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// Create low-priority source
	lowDir := filepath.Join(tmpDir, "low", "my-cmd")
	if err := os.MkdirAll(lowDir, 0755); err != nil {
		t.Fatalf("Failed to create low dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lowDir, "INDEX.dro.md"), []byte("low-priority content\n"), 0644); err != nil {
		t.Fatalf("Failed to write low INDEX: %v", err)
	}

	// Create high-priority source
	highDir := filepath.Join(tmpDir, "high", "my-cmd")
	if err := os.MkdirAll(highDir, 0755); err != nil {
		t.Fatalf("Failed to create high dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(highDir, "INDEX.dro.md"), []byte("high-priority content\n"), 0644); err != nil {
		t.Fatalf("Failed to write high INDEX: %v", err)
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Low priority first, high priority second (later overrides)
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

	// Verify the high-priority content wins (stripped to INDEX.md)
	cmdIndex := filepath.Join(commandsDir, "my-cmd", "INDEX.md")
	content, err := os.ReadFile(cmdIndex)
	if err != nil {
		t.Fatalf("Failed to read projected INDEX.md: %v", err)
	}

	if string(content) != "high-priority content\n" {
		t.Errorf("Expected high-priority content, got %q", string(content))
	}
}

// TestSyncMena_EmbeddedFS verifies projection from an embedded FS source
// with extension stripping. Uses a realistic path structure matching how
// materializeMena builds embedded sources (e.g., "rites/shared/mena").
func TestSyncMena_EmbeddedFS(t *testing.T) {
	tmpDir := t.TempDir()

	// Build an in-memory FS mimicking real embedded rite structure
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

	// Sources use FsysPath matching the real embedded directory structure
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

	// Verify dromena projected to commands/ with stripped name
	cmdIndex := filepath.Join(commandsDir, "my-cmd", "INDEX.md")
	if _, err := os.Stat(cmdIndex); os.IsNotExist(err) {
		t.Errorf("Expected embedded dromena at %s (stripped), but it does not exist", cmdIndex)
	}

	// Verify legomena projected to skills/ with stripped name
	skillIndex := filepath.Join(skillsDir, "my-ref", "INDEX.md")
	if _, err := os.Stat(skillIndex); os.IsNotExist(err) {
		t.Errorf("Expected embedded legomena at %s (stripped), but it does not exist", skillIndex)
	}

	// Verify un-stripped names do NOT exist
	cmdOld := filepath.Join(commandsDir, "my-cmd", "INDEX.dro.md")
	if _, err := os.Stat(cmdOld); err == nil {
		t.Errorf("INDEX.dro.md should not exist in embedded output")
	}
	skillOld := filepath.Join(skillsDir, "my-ref", "INDEX.lego.md")
	if _, err := os.Stat(skillOld); err == nil {
		t.Errorf("INDEX.lego.md should not exist in embedded output")
	}

	// Verify result tracking
	if len(result.CommandsProjected) == 0 {
		t.Errorf("Expected CommandsProjected to be non-empty for embedded source")
	}
	if len(result.SkillsProjected) == 0 {
		t.Errorf("Expected SkillsProjected to be non-empty for embedded source")
	}
}

// TestSyncMena_Filter_DroOnly verifies that ProjectDro filter only projects dromena.
func TestSyncMena_Filter_DroOnly(t *testing.T) {
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

	// Commands should exist
	if _, err := os.Stat(filepath.Join(commandsDir, "cmd1", "INDEX.md")); os.IsNotExist(err) {
		t.Errorf("Expected cmd1 to be projected to commands/")
	}

	// Skills should NOT be created (filter excludes lego)
	if _, err := os.Stat(filepath.Join(skillsDir, "ref1", "INDEX.md")); !os.IsNotExist(err) {
		t.Errorf("Expected ref1 to NOT be projected when filter is ProjectDro")
	}

	if len(result.SkillsProjected) != 0 {
		t.Errorf("Expected no skills projected, got %v", result.SkillsProjected)
	}
}

// TestSyncMena_StandaloneFileStripping verifies that standalone files in
// grouping directories also have their extensions stripped.
func TestSyncMena_StandaloneFileStripping(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a grouping directory with standalone files
	menaDir := filepath.Join(tmpDir, "mena")
	groupDir := filepath.Join(menaDir, "navigation")
	if err := os.MkdirAll(groupDir, 0755); err != nil {
		t.Fatalf("Failed to create group dir: %v", err)
	}

	// Standalone file with .dro extension
	if err := os.WriteFile(filepath.Join(groupDir, "rite.dro.md"), []byte("# Rite Navigation\n"), 0644); err != nil {
		t.Fatalf("Failed to write standalone dro: %v", err)
	}

	// Standalone file with .lego extension
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

	// Verify standalone dro stripped and routed to commands/
	droStripped := filepath.Join(commandsDir, "navigation", "rite.md")
	if _, err := os.Stat(droStripped); os.IsNotExist(err) {
		t.Errorf("Expected standalone dro at %s (stripped from rite.dro.md), but it does not exist", droStripped)
	}
	droOld := filepath.Join(commandsDir, "navigation", "rite.dro.md")
	if _, err := os.Stat(droOld); err == nil {
		t.Errorf("Un-stripped rite.dro.md should not exist in output")
	}

	// Verify standalone lego stripped and routed to skills/
	legoStripped := filepath.Join(skillsDir, "navigation", "reference.md")
	if _, err := os.Stat(legoStripped); os.IsNotExist(err) {
		t.Errorf("Expected standalone lego at %s (stripped from reference.lego.md), but it does not exist", legoStripped)
	}
	legoOld := filepath.Join(skillsDir, "navigation", "reference.lego.md")
	if _, err := os.Stat(legoOld); err == nil {
		t.Errorf("Un-stripped reference.lego.md should not exist in output")
	}
}

// TestSyncMena_EmptySources verifies graceful handling of empty sources.
func TestSyncMena_EmptySources(t *testing.T) {
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

	if len(result.CommandsProjected) != 0 {
		t.Errorf("Expected no commands projected from empty sources")
	}
	if len(result.SkillsProjected) != 0 {
		t.Errorf("Expected no skills projected from empty sources")
	}

	// Directories should still be created
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Errorf("commands/ dir should be created even with empty sources")
	}
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Errorf("skills/ dir should be created even with empty sources")
	}
}

// TestSyncMena_NonexistentSource verifies graceful handling of sources that
// don't exist on disk.
func TestSyncMena_NonexistentSource(t *testing.T) {
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{
		{Path: filepath.Join(tmpDir, "nonexistent-mena")},
	}
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
	tests := []struct {
		name     string
		data     string
		wantName string
	}{
		{
			"with frontmatter",
			"---\nname: test\ndescription: d\n---\n# Body\n",
			"test",
		},
		{
			"no frontmatter delimiters",
			"# Just a file with no frontmatter\n",
			"",
		},
		{
			"malformed YAML",
			"---\n: [invalid yaml\n---\n# Body\n",
			"",
		},
		{
			"empty content",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := parseMenaFrontmatterBytes([]byte(tt.data))
			if fm.Name != tt.wantName {
				t.Errorf("name = %q, want %q", fm.Name, tt.wantName)
			}
		})
	}
}
