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

// TestProjectMena_Destructive verifies full projection with destructive mode:
// selectively replaces managed entries while preserving user-created content.
func TestProjectMena_Destructive(t *testing.T) {
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

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
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

// TestProjectMena_Additive verifies additive mode preserves existing files.
func TestProjectMena_Additive(t *testing.T) {
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

	_, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena additive failed: %v", err)
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

// TestProjectMena_PriorityOverride verifies that later sources override earlier
// sources for the same mena name.
func TestProjectMena_PriorityOverride(t *testing.T) {
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

	_, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena priority override failed: %v", err)
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

// TestProjectMena_EmbeddedFS verifies projection from an embedded FS source
// with extension stripping. Uses a realistic path structure matching how
// materializeMena builds embedded sources (e.g., "rites/shared/mena").
func TestProjectMena_EmbeddedFS(t *testing.T) {
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

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena embedded failed: %v", err)
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

// TestProjectMena_Filter_DroOnly verifies that ProjectDro filter only projects dromena.
func TestProjectMena_Filter_DroOnly(t *testing.T) {
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

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
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

// TestProjectMena_StandaloneFileStripping verifies that standalone files in
// grouping directories also have their extensions stripped.
func TestProjectMena_StandaloneFileStripping(t *testing.T) {
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

	_, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
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

// TestProjectMena_EmptySources verifies graceful handling of empty sources.
func TestProjectMena_EmptySources(t *testing.T) {
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

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena with empty sources should not fail: %v", err)
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

// TestProjectMena_NonexistentSource verifies graceful handling of sources that
// don't exist on disk.
func TestProjectMena_NonexistentSource(t *testing.T) {
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

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena with nonexistent source should not fail: %v", err)
	}

	if len(result.CommandsProjected) != 0 || len(result.SkillsProjected) != 0 {
		t.Errorf("Expected no projections from nonexistent source")
	}
}

// --- Scope filtering tests ---

// TestScopeIncludesPipeline verifies the truth table from TDD Section 3.
func TestScopeIncludesPipeline(t *testing.T) {
	tests := []struct {
		entryScope    MenaScope
		pipelineScope MenaScope
		want          bool
		reason        string
	}{
		{MenaScopeBoth, MenaScopeBoth, true, "no filtering on either side"},
		{MenaScopeBoth, MenaScopeProject, true, "entry goes to both, pipeline is project"},
		{MenaScopeBoth, MenaScopeUser, true, "entry goes to both, pipeline is user"},
		{MenaScopeUser, MenaScopeBoth, true, "no pipeline filtering requested"},
		{MenaScopeUser, MenaScopeUser, true, "match"},
		{MenaScopeUser, MenaScopeProject, false, "entry is user-only, pipeline is project"},
		{MenaScopeProject, MenaScopeBoth, true, "no pipeline filtering requested"},
		{MenaScopeProject, MenaScopeUser, false, "entry is project-only, pipeline is user"},
		{MenaScopeProject, MenaScopeProject, true, "match"},
	}

	for _, tt := range tests {
		name := "entry=" + tt.entryScope.String() + "_pipeline=" + tt.pipelineScope.String()
		t.Run(name, func(t *testing.T) {
			got := scopeIncludesPipeline(tt.entryScope, tt.pipelineScope)
			if got != tt.want {
				t.Errorf("scopeIncludesPipeline(%q, %q) = %v, want %v (%s)",
					string(tt.entryScope), string(tt.pipelineScope), got, tt.want, tt.reason)
			}
		})
	}
}

// TestParseMenaFrontmatterBytes verifies frontmatter extraction from raw bytes.
func TestParseMenaFrontmatterBytes(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantScope MenaScope
		wantName  string
	}{
		{
			"scope user",
			"---\nname: test\ndescription: d\nscope: user\n---\n# Body\n",
			MenaScopeUser, "test",
		},
		{
			"scope project",
			"---\nname: test\ndescription: d\nscope: project\n---\n# Body\n",
			MenaScopeProject, "test",
		},
		{
			"no scope field",
			"---\nname: test\ndescription: d\n---\n# Body\n",
			MenaScopeBoth, "test",
		},
		{
			"no frontmatter delimiters",
			"# Just a file with no frontmatter\n",
			MenaScopeBoth, "",
		},
		{
			"malformed YAML",
			"---\n: [invalid yaml\n---\n# Body\n",
			MenaScopeBoth, "",
		},
		{
			"empty content",
			"",
			MenaScopeBoth, "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := parseMenaFrontmatterBytes([]byte(tt.data))
			if fm.Scope != tt.wantScope {
				t.Errorf("scope = %q, want %q", string(fm.Scope), string(tt.wantScope))
			}
			if fm.Name != tt.wantName {
				t.Errorf("name = %q, want %q", fm.Name, tt.wantName)
			}
		})
	}
}

// TestReadMenaFrontmatterFromDir verifies directory-level frontmatter reading.
func TestReadMenaFrontmatterFromDir(t *testing.T) {
	t.Run("INDEX.dro.md with scope:user", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "INDEX.dro.md"),
			[]byte("---\nname: test\ndescription: d\nscope: user\n---\n# Body\n"), 0644); err != nil {
			t.Fatal(err)
		}
		fm := ReadMenaFrontmatterFromDir(dir)
		if fm.Scope != MenaScopeUser {
			t.Errorf("scope = %q, want %q", string(fm.Scope), string(MenaScopeUser))
		}
	})

	t.Run("INDEX.lego.md with scope:project", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "INDEX.lego.md"),
			[]byte("---\nname: test\ndescription: d\nscope: project\n---\n# Body\n"), 0644); err != nil {
			t.Fatal(err)
		}
		fm := ReadMenaFrontmatterFromDir(dir)
		if fm.Scope != MenaScopeProject {
			t.Errorf("scope = %q, want %q", string(fm.Scope), string(MenaScopeProject))
		}
	})

	t.Run("INDEX missing frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "INDEX.dro.md"),
			[]byte("# No frontmatter\n"), 0644); err != nil {
			t.Fatal(err)
		}
		fm := ReadMenaFrontmatterFromDir(dir)
		if fm.Scope != MenaScopeBoth {
			t.Errorf("scope = %q, want %q", string(fm.Scope), string(MenaScopeBoth))
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		fm := ReadMenaFrontmatterFromDir("/nonexistent/path/that/does/not/exist")
		if fm.Scope != MenaScopeBoth {
			t.Errorf("scope = %q, want %q", string(fm.Scope), string(MenaScopeBoth))
		}
	})
}

// TestReadMenaFrontmatterFromFile verifies file-level frontmatter reading.
func TestReadMenaFrontmatterFromFile(t *testing.T) {
	t.Run("file with scope:user", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "test.dro.md")
		if err := os.WriteFile(p,
			[]byte("---\nname: test\ndescription: d\nscope: user\n---\n# Body\n"), 0644); err != nil {
			t.Fatal(err)
		}
		fm := ReadMenaFrontmatterFromFile(p)
		if fm.Scope != MenaScopeUser {
			t.Errorf("scope = %q, want %q", string(fm.Scope), string(MenaScopeUser))
		}
	})

	t.Run("file with no frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "test.dro.md")
		if err := os.WriteFile(p, []byte("# No frontmatter\n"), 0644); err != nil {
			t.Fatal(err)
		}
		fm := ReadMenaFrontmatterFromFile(p)
		if fm.Scope != MenaScopeBoth {
			t.Errorf("scope = %q, want %q", string(fm.Scope), string(MenaScopeBoth))
		}
	})
}

// helper: create a leaf mena directory with INDEX file containing optional scope.
func createMenaLeaf(t *testing.T, baseDir, name, indexName, scope string) {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create dir %s: %v", dir, err)
	}
	var content string
	if scope != "" {
		content = "---\nname: " + name + "\ndescription: test\nscope: " + scope + "\n---\n# Body\n"
	} else {
		content = "---\nname: " + name + "\ndescription: test\n---\n# Body\n"
	}
	if err := os.WriteFile(filepath.Join(dir, indexName), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", indexName, err)
	}
}

// TestProjectMena_ScopeUser_ExcludedFromProject verifies scope:user entries are
// excluded when PipelineScope is MenaScopeProject.
func TestProjectMena_ScopeUser_ExcludedFromProject(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")
	createMenaLeaf(t, menaDir, "user-cmd", "INDEX.dro.md", "user")

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	if len(result.CommandsProjected) != 0 {
		t.Errorf("scope:user entry should be excluded from project pipeline, got projected: %v", result.CommandsProjected)
	}

	if _, err := os.Stat(filepath.Join(commandsDir, "user-cmd", "INDEX.md")); !os.IsNotExist(err) {
		t.Errorf("scope:user entry should not exist in commands/")
	}
}

// TestProjectMena_ScopeProject_ExcludedFromUser verifies scope:project entries are
// excluded when PipelineScope is MenaScopeUser.
func TestProjectMena_ScopeProject_ExcludedFromUser(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")
	createMenaLeaf(t, menaDir, "proj-cmd", "INDEX.dro.md", "project")

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeUser,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	if len(result.CommandsProjected) != 0 {
		t.Errorf("scope:project entry should be excluded from user pipeline, got projected: %v", result.CommandsProjected)
	}
}

// TestProjectMena_ScopeUser_IncludedInUser verifies scope:user entries are
// included when PipelineScope is MenaScopeUser.
func TestProjectMena_ScopeUser_IncludedInUser(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")
	createMenaLeaf(t, menaDir, "user-cmd", "INDEX.dro.md", "user")

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeUser,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	if len(result.CommandsProjected) != 1 {
		t.Errorf("scope:user entry should be included in user pipeline, got %d projected", len(result.CommandsProjected))
	}
}

// TestProjectMena_ScopeProject_IncludedInProject verifies scope:project entries are
// included when PipelineScope is MenaScopeProject.
func TestProjectMena_ScopeProject_IncludedInProject(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")
	createMenaLeaf(t, menaDir, "proj-cmd", "INDEX.dro.md", "project")

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	if len(result.CommandsProjected) != 1 {
		t.Errorf("scope:project entry should be included in project pipeline, got %d projected", len(result.CommandsProjected))
	}
}

// TestProjectMena_NoScope_IncludedInBoth verifies entries without scope are
// included in both pipelines.
func TestProjectMena_NoScope_IncludedInBoth(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")
	createMenaLeaf(t, menaDir, "both-cmd", "INDEX.dro.md", "") // no scope

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}

	// Test with PipelineScope: MenaScopeProject
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena (project) failed: %v", err)
	}
	if len(result.CommandsProjected) != 1 {
		t.Errorf("no-scope entry should be included in project pipeline, got %d projected", len(result.CommandsProjected))
	}

	// Test with PipelineScope: MenaScopeUser
	commandsDir2 := filepath.Join(tmpDir, "commands2")
	skillsDir2 := filepath.Join(tmpDir, "skills2")
	opts2 := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeUser,
		TargetCommandsDir: commandsDir2,
		TargetSkillsDir:   skillsDir2,
	}

	result2, err := ProjectMena(sources, opts2)
	if err != nil {
		t.Fatalf("ProjectMena (user) failed: %v", err)
	}
	if len(result2.CommandsProjected) != 1 {
		t.Errorf("no-scope entry should be included in user pipeline, got %d projected", len(result2.CommandsProjected))
	}
}

// TestProjectMena_NoPipelineScope_NoFiltering verifies that when PipelineScope
// is zero value (MenaScopeBoth), no scope filtering occurs (EC-8).
func TestProjectMena_NoPipelineScope_NoFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")
	createMenaLeaf(t, menaDir, "user-cmd", "INDEX.dro.md", "user")

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeBoth, // zero value -- no filtering
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	if len(result.CommandsProjected) != 1 {
		t.Errorf("no pipeline scope should include all entries, got %d projected", len(result.CommandsProjected))
	}
}

// TestProjectMena_StandaloneFile_ScopeFiltered verifies scope filtering for standalone files.
func TestProjectMena_StandaloneFile_ScopeFiltered(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a grouping directory with standalone file having scope:user
	menaDir := filepath.Join(tmpDir, "mena")
	groupDir := filepath.Join(menaDir, "navigation")
	if err := os.MkdirAll(groupDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(groupDir, "rite.dro.md"),
		[]byte("---\nname: rite\ndescription: test\nscope: user\n---\n# Body\n"), 0644); err != nil {
		t.Fatal(err)
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	if len(result.CommandsProjected) != 0 {
		t.Errorf("scope:user standalone should be excluded from project pipeline, got: %v", result.CommandsProjected)
	}

	// Verify the file does not exist in output
	if _, err := os.Stat(filepath.Join(commandsDir, "navigation", "rite.md")); !os.IsNotExist(err) {
		t.Error("scope:user standalone file should not exist in commands/")
	}
}

// TestProjectMena_ScopeWithEmbeddedFS verifies scope filtering with embedded FS sources.
func TestProjectMena_ScopeWithEmbeddedFS(t *testing.T) {
	tmpDir := t.TempDir()

	fsys := fstest.MapFS{
		"rites/test-rite/mena/user-cmd/INDEX.dro.md": &fstest.MapFile{
			Data: []byte("---\nname: user-cmd\ndescription: test\nscope: user\n---\n# Body\n"),
		},
	}

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{
		{Fsys: fsys, FsysPath: "rites/test-rite/mena", IsEmbedded: true},
	}

	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena embedded failed: %v", err)
	}

	if len(result.CommandsProjected) != 0 {
		t.Errorf("scope:user embedded entry should be excluded from project pipeline, got: %v", result.CommandsProjected)
	}
}

// TestProjectMena_MixedScopes verifies mixed scope entries are filtered correctly.
func TestProjectMena_MixedScopes(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")

	createMenaLeaf(t, menaDir, "user-only", "INDEX.dro.md", "user")
	createMenaLeaf(t, menaDir, "proj-only", "INDEX.dro.md", "project")
	createMenaLeaf(t, menaDir, "both-cmd", "INDEX.dro.md", "")

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	result, err := ProjectMena(sources, opts)
	if err != nil {
		t.Fatalf("ProjectMena failed: %v", err)
	}

	// Should have proj-only and both-cmd, but NOT user-only
	projected := make(map[string]bool)
	for _, p := range result.CommandsProjected {
		projected[p] = true
	}

	if projected["user-only"] {
		t.Error("scope:user should be excluded from project pipeline")
	}
	if !projected["proj-only"] {
		t.Error("scope:project should be included in project pipeline")
	}
	if !projected["both-cmd"] {
		t.Error("no-scope should be included in project pipeline")
	}
	if len(result.CommandsProjected) != 2 {
		t.Errorf("expected 2 projected commands, got %d: %v", len(result.CommandsProjected), result.CommandsProjected)
	}
}

// TestProjectMena_ScopeEndToEnd tests the full scope filtering lifecycle.
func TestProjectMena_ScopeEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")

	createMenaLeaf(t, menaDir, "user-entry", "INDEX.dro.md", "user")
	createMenaLeaf(t, menaDir, "project-entry", "INDEX.dro.md", "project")
	createMenaLeaf(t, menaDir, "both-entry", "INDEX.dro.md", "")

	// Test 1: Project pipeline
	cmdDir1 := filepath.Join(tmpDir, "cmd1")
	skillDir1 := filepath.Join(tmpDir, "skill1")
	r1, err := ProjectMena([]MenaSource{{Path: menaDir}}, MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeProject,
		TargetCommandsDir: cmdDir1,
		TargetSkillsDir:   skillDir1,
	})
	if err != nil {
		t.Fatalf("ProjectMena (project) failed: %v", err)
	}

	p1 := make(map[string]bool)
	for _, p := range r1.CommandsProjected {
		p1[p] = true
	}
	if !p1["project-entry"] || !p1["both-entry"] {
		t.Errorf("project pipeline should include project-entry and both-entry, got: %v", r1.CommandsProjected)
	}
	if p1["user-entry"] {
		t.Error("project pipeline should NOT include user-entry")
	}

	// Test 2: User pipeline
	cmdDir2 := filepath.Join(tmpDir, "cmd2")
	skillDir2 := filepath.Join(tmpDir, "skill2")
	r2, err := ProjectMena([]MenaSource{{Path: menaDir}}, MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		PipelineScope:     MenaScopeUser,
		TargetCommandsDir: cmdDir2,
		TargetSkillsDir:   skillDir2,
	})
	if err != nil {
		t.Fatalf("ProjectMena (user) failed: %v", err)
	}

	p2 := make(map[string]bool)
	for _, p := range r2.CommandsProjected {
		p2[p] = true
	}
	if !p2["user-entry"] || !p2["both-entry"] {
		t.Errorf("user pipeline should include user-entry and both-entry, got: %v", r2.CommandsProjected)
	}
	if p2["project-entry"] {
		t.Error("user pipeline should NOT include project-entry")
	}
}
