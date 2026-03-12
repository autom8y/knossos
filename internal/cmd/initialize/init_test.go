package initialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/paths"
)

func TestInit_FreshDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Set up context pointing to the temp directory.
	outputFlag := "text"
	verbose := false
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &outputFlag,
			Verbose:    &verbose,
			ProjectDir: &dir,
		},
		WorkDir: dir,
	}

	// Run init with no rite (minimal scaffold).
	err := runInit(ctx, "", "", false, nil)
	if err != nil {
		t.Fatalf("runInit() returned error: %v", err)
	}

	// Verify .claude/ directory was created.
	channelDir := filepath.Join(dir, paths.ClaudeChannel{}.DirName())
	if _, err := os.Stat(channelDir); os.IsNotExist(err) {
		t.Error(".claude/ directory was not created")
	}

	// Verify KNOSSOS_MANIFEST.yaml was created in .knossos/.
	manifestPath := filepath.Join(dir, ".knossos", "KNOSSOS_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("KNOSSOS_MANIFEST.yaml was not created in .knossos/")
	}

	// Verify CLAUDE.md was created.
	claudeMdPath := filepath.Join(channelDir, paths.ClaudeChannel{}.ContextFile())
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Error("CLAUDE.md was not created")
	}

	// Verify settings.local.json was created.
	settingsPath := filepath.Join(channelDir, "settings.local.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.local.json was not created")
	}

	// Verify project-level directories were scaffolded.
	for _, dirName := range []string{".knossos", ".sos", ".ledge"} {
		dirPath := filepath.Join(dir, dirName)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("%s/ directory was not created", dirName)
		}
	}

	// Verify .ledge subdirectories and .gitkeep files.
	verifyLedgeSubdirs(t, dir)
}

func TestInit_WithRite(t *testing.T) {
	// Not parallel: mutates global embedded assets via common.SetEmbeddedAssets.
	dir := t.TempDir()

	// Create a synthetic embedded rites FS with a test rite.
	embeddedRites := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte(`name: test-rite
version: "1.0"
description: "Test rite for init"
entry_agent: orchestrator
agents:
  - name: orchestrator
    role: "Coordinates work"
  - name: worker
    role: "Does work"
dromena: []
legomena: []
dependencies:
  - shared
`),
		},
		"rites/test-rite/agents/orchestrator.md": &fstest.MapFile{
			Data: []byte("# Orchestrator\nCoordinates work."),
		},
		"rites/test-rite/agents/worker.md": &fstest.MapFile{
			Data: []byte("# Worker\nDoes work."),
		},
		"rites/shared/mena/.keep": &fstest.MapFile{
			Data: []byte(""),
		},
	}

	// Set embedded assets for the test.
	common.SetEmbeddedAssets(embeddedRites, nil, nil)
	defer common.SetEmbeddedAssets(nil, nil, nil)

	outputFlag := "text"
	verbose := false
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &outputFlag,
			Verbose:    &verbose,
			ProjectDir: &dir,
		},
		WorkDir: dir,
	}

	err := runInit(ctx, "test-rite", "", false, nil)
	if err != nil {
		t.Fatalf("runInit() with rite returned error: %v", err)
	}

	channelDir := filepath.Join(dir, paths.ClaudeChannel{}.DirName())

	// Verify agents were materialized.
	agentsDir := filepath.Join(channelDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Error("agents/ directory was not created")
	}

	// Verify specific agent files exist.
	for _, agentName := range []string{"orchestrator.md", "worker.md"} {
		agentPath := filepath.Join(agentsDir, agentName)
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Errorf("agent file %s was not created", agentName)
		}
	}

	// Verify ACTIVE_RITE was written.
	knossosDir := filepath.Join(dir, ".knossos")
	activeRitePath := filepath.Join(knossosDir, "ACTIVE_RITE")
	data, err := os.ReadFile(activeRitePath)
	if err != nil {
		t.Fatalf("failed to read ACTIVE_RITE: %v", err)
	}
	if got := string(data); got != "test-rite\n" {
		t.Errorf("ACTIVE_RITE = %q, want %q", got, "test-rite\n")
	}

	// Verify KNOSSOS_MANIFEST.yaml was created in .knossos/.
	manifestPath := filepath.Join(dir, ".knossos", "KNOSSOS_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("KNOSSOS_MANIFEST.yaml was not created in .knossos/")
	}

	// Verify project-level directories were scaffolded.
	for _, dirName := range []string{".knossos", ".sos", ".ledge"} {
		dirPath := filepath.Join(dir, dirName)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("%s/ directory was not created", dirName)
		}
	}

	// Verify .ledge subdirectories and .gitkeep files.
	verifyLedgeSubdirs(t, dir)
}

func TestInit_AlreadyInitialized(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create .knossos/ with a valid KNOSSOS_MANIFEST.yaml to simulate existing init.
	knossosDir := filepath.Join(dir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(knossosDir, "KNOSSOS_MANIFEST.yaml")
	validManifest := `schema_version: "1.0"
inscription_version: "1"
regions:
  user-content:
    owner: satellite
section_order:
  - user-content
`
	if err := os.WriteFile(manifestPath, []byte(validManifest), 0644); err != nil {
		t.Fatal(err)
	}

	outputFlag := "text"
	verbose := false
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &outputFlag,
			Verbose:    &verbose,
			ProjectDir: &dir,
		},
		WorkDir: dir,
	}

	// Running without --force should succeed (exit 0) without modifying anything.
	err := runInit(ctx, "", "", false, nil)
	if err != nil {
		t.Fatalf("runInit() should succeed for already initialized, got error: %v", err)
	}

	// Verify the original manifest is unchanged (not overwritten).
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != validManifest {
		t.Errorf("manifest was modified: got %q, want %q", got, validManifest)
	}
}

func TestInit_Force(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create .knossos/ with a valid KNOSSOS_MANIFEST.yaml.
	knossosDir := filepath.Join(dir, ".knossos")
	channelDir := filepath.Join(dir, paths.ClaudeChannel{}.DirName())
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(knossosDir, "KNOSSOS_MANIFEST.yaml")
	validManifest := `schema_version: "1.0"
inscription_version: "1"
regions:
  user-content:
    owner: satellite
section_order:
  - user-content
`
	if err := os.WriteFile(manifestPath, []byte(validManifest), 0644); err != nil {
		t.Fatal(err)
	}

	outputFlag := "text"
	verbose := false
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &outputFlag,
			Verbose:    &verbose,
			ProjectDir: &dir,
		},
		WorkDir: dir,
	}

	// Running with --force should reinitialize.
	err := runInit(ctx, "", "", true, nil)
	if err != nil {
		t.Fatalf("runInit() with --force returned error: %v", err)
	}

	// Verify CLAUDE.md was generated (proves re-materialization happened).
	claudeMdPath := filepath.Join(channelDir, paths.ClaudeChannel{}.ContextFile())
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Error("CLAUDE.md was not created after --force reinitialize")
	}

	// Verify settings.local.json was created.
	settingsPath := filepath.Join(channelDir, "settings.local.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.local.json was not created after --force reinitialize")
	}
}

func TestInit_NonKnossosClaudeDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create .claude/ without KNOSSOS_MANIFEST.yaml (non-Knossos project).
	channelDir := filepath.Join(dir, paths.ClaudeChannel{}.DirName())
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write some existing content (like a manual CLAUDE.md).
	if err := os.WriteFile(filepath.Join(channelDir, paths.ClaudeChannel{}.ContextFile()), []byte("# Custom"), 0644); err != nil {
		t.Fatal(err)
	}

	outputFlag := "text"
	verbose := false
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &outputFlag,
			Verbose:    &verbose,
			ProjectDir: &dir,
		},
		WorkDir: dir,
	}

	// Should error without --force.
	err := runInit(ctx, "", "", false, nil)
	if err == nil {
		t.Fatal("runInit() should error for non-Knossos .claude/ without --force")
	}

	// Should succeed with --force.
	err = runInit(ctx, "", "", true, nil)
	if err != nil {
		t.Fatalf("runInit() with --force should succeed, got error: %v", err)
	}
}

func TestInit_NeedsProjectFalse(t *testing.T) {
	t.Parallel()
	outputFlag := "text"
	verbose := false
	projectDir := ""

	cmd := NewInitCmd(&outputFlag, &verbose, &projectDir)
	if common.NeedsProject(cmd) {
		t.Error("init command should have needsProject=false, got true")
	}
}

func TestInit_LedgeIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Run scaffoldProjectDirs twice — should not error or overwrite.
	scaffoldProjectDirs(dir)
	scaffoldProjectDirs(dir)

	verifyLedgeSubdirs(t, dir)
}

// makeTestMenaFS returns a minimal embedded mena FS suitable for
// extractEmbeddedMenaToXDG tests.
func makeTestMenaFS() fstest.MapFS {
	return fstest.MapFS{
		"mena/navigation/go.dro.md": &fstest.MapFile{
			Data: []byte("# go\nNavigate."),
		},
		"mena/guidance/intro.lego.md": &fstest.MapFile{
			Data: []byte("# Intro\nGuidance."),
		},
	}
}

// testXDGDataDir creates a temp directory suitable for use as an XDG data dir
// and returns both the base "knossos" dir (to pass to extractEmbeddedMenaToXDG)
// and the expected mena subdirectory path.
func testXDGDataDir(t *testing.T) (knossosDir string, xdgMena string) {
	t.Helper()
	xdgBase := t.TempDir()
	knossosDir = filepath.Join(xdgBase, "knossos")
	xdgMena = filepath.Join(knossosDir, "mena")
	return knossosDir, xdgMena
}

// TestExtractEmbeddedMenaToXDG_FreshExtraction verifies that when neither the
// XDG mena directory nor the sentinel exist, extraction runs and the sentinel
// is written with the current version.
func TestExtractEmbeddedMenaToXDG_FreshExtraction(t *testing.T) {
	t.Parallel()
	knossosDir, xdgMena := testXDGDataDir(t)

	embMena := makeTestMenaFS()
	extractEmbeddedMenaToXDG(embMena, extractOpts{xdgDataDir: knossosDir, version: "v1.0.0"})

	// Directory must exist.
	if _, err := os.Stat(xdgMena); os.IsNotExist(err) {
		t.Fatal("XDG mena directory was not created")
	}

	// Sentinel must contain the version.
	sentinel := filepath.Join(xdgMena, xdgVersionSentinel)
	data, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("sentinel file not written: %v", err)
	}
	if got := string(data); got != "v1.0.0" {
		t.Errorf("sentinel = %q, want %q", got, "v1.0.0")
	}

	// At least one content file must be present.
	goFile := filepath.Join(xdgMena, "navigation", "go.dro.md")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Error("extracted mena content file missing")
	}
}

// TestExtractEmbeddedMenaToXDG_VersionMatch verifies that when the sentinel
// matches the current version, extraction is skipped (idempotent).
func TestExtractEmbeddedMenaToXDG_VersionMatch(t *testing.T) {
	t.Parallel()
	knossosDir, xdgMena := testXDGDataDir(t)

	// Pre-create directory with sentinel matching current version.
	if err := os.MkdirAll(xdgMena, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(xdgMena, xdgVersionSentinel)
	if err := os.WriteFile(sentinel, []byte("v1.0.0"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a marker file that should NOT be overwritten.
	markerPath := filepath.Join(xdgMena, "marker.txt")
	if err := os.WriteFile(markerPath, []byte("keep-me"), 0644); err != nil {
		t.Fatal(err)
	}

	embMena := makeTestMenaFS()
	extractEmbeddedMenaToXDG(embMena, extractOpts{xdgDataDir: knossosDir, version: "v1.0.0"})

	// Marker file must still exist (extraction was skipped).
	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("marker file missing after version-match skip: %v", err)
	}
	if string(data) != "keep-me" {
		t.Errorf("marker file content changed: %q", string(data))
	}

	// Sentinel must remain unchanged.
	data, err = os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("sentinel missing: %v", err)
	}
	if string(data) != "v1.0.0" {
		t.Errorf("sentinel = %q, want %q", string(data), "v1.0.0")
	}
}

// TestExtractEmbeddedMenaToXDG_VersionMismatch verifies that when the sentinel
// records a different version, the XDG mena directory is wiped and re-extracted,
// and the sentinel is updated to the current version.
func TestExtractEmbeddedMenaToXDG_VersionMismatch(t *testing.T) {
	t.Parallel()
	knossosDir, xdgMena := testXDGDataDir(t)

	// Pre-create directory with sentinel for old version.
	if err := os.MkdirAll(xdgMena, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(xdgMena, xdgVersionSentinel)
	if err := os.WriteFile(sentinel, []byte("v1.0.0"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a stale file that should be removed on re-extraction.
	stalePath := filepath.Join(xdgMena, "stale.md")
	if err := os.WriteFile(stalePath, []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}

	embMena := makeTestMenaFS()
	extractEmbeddedMenaToXDG(embMena, extractOpts{xdgDataDir: knossosDir, version: "v2.0.0"})

	// Stale file must be gone (directory was wiped).
	if _, err := os.Stat(stalePath); err == nil {
		t.Error("stale file survived re-extraction -- wipe did not happen")
	}

	// Sentinel must be updated to current version.
	data, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("sentinel missing after re-extraction: %v", err)
	}
	if string(data) != "v2.0.0" {
		t.Errorf("sentinel = %q, want %q", string(data), "v2.0.0")
	}

	// Content from embedded FS must be present.
	goFile := filepath.Join(xdgMena, "navigation", "go.dro.md")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Error("re-extracted mena content file missing")
	}
}

// TestExtractEmbeddedMenaToXDG_DirectoryExistsNoSentinel verifies that when the
// XDG mena directory exists but has no version sentinel (legacy state), extraction
// is treated as stale and re-runs.
func TestExtractEmbeddedMenaToXDG_DirectoryExistsNoSentinel(t *testing.T) {
	t.Parallel()
	knossosDir, xdgMena := testXDGDataDir(t)

	// Pre-create directory with content but NO sentinel.
	if err := os.MkdirAll(xdgMena, 0755); err != nil {
		t.Fatal(err)
	}
	stalePath := filepath.Join(xdgMena, "legacy-file.md")
	if err := os.WriteFile(stalePath, []byte("legacy"), 0644); err != nil {
		t.Fatal(err)
	}

	embMena := makeTestMenaFS()
	extractEmbeddedMenaToXDG(embMena, extractOpts{xdgDataDir: knossosDir, version: "v1.0.0"})

	// Legacy file must be gone (directory was wiped and re-extracted).
	if _, err := os.Stat(stalePath); err == nil {
		t.Error("legacy file survived re-extraction -- stale directory not wiped")
	}

	// Sentinel must now exist with current version.
	sentinel := filepath.Join(xdgMena, xdgVersionSentinel)
	data, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("sentinel not written after re-extraction of sentinel-less dir: %v", err)
	}
	if string(data) != "v1.0.0" {
		t.Errorf("sentinel = %q, want %q", string(data), "v1.0.0")
	}

	// Embedded content must be present.
	goFile := filepath.Join(xdgMena, "navigation", "go.dro.md")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Error("re-extracted mena content file missing after sentinel-less wipe")
	}
}

func TestWriteProjectGitignore_NewFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeProjectGitignore(dir)

	gitignorePath := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	content := string(data)

	for _, pattern := range []string{"# Knossos", "# End Knossos", ".knossos/", ".claude/CLAUDE.md", "**/.sos/*", "!**/.sos/land/", "**/.ledge/*", "!**/.ledge/shelf/"} {
		if !strings.Contains(content, pattern) {
			t.Errorf(".gitignore missing pattern %q", pattern)
		}
	}
}

func TestWriteProjectGitignore_AppendToExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create .gitignore with user content.
	gitignorePath := filepath.Join(dir, ".gitignore")
	userContent := "# My project\nnode_modules/\n*.log\n"
	if err := os.WriteFile(gitignorePath, []byte(userContent), 0644); err != nil {
		t.Fatal(err)
	}

	writeProjectGitignore(dir)

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	content := string(data)

	// User content preserved.
	if !strings.Contains(content, "node_modules/") {
		t.Error("user content was not preserved")
	}
	// Knossos block appended.
	if !strings.Contains(content, "# Knossos") {
		t.Error("Knossos block was not appended")
	}
	// Blank line separator between user content and block.
	if !strings.Contains(content, "*.log\n\n# Knossos") {
		t.Error("missing blank line separator before Knossos block")
	}
}

func TestWriteProjectGitignore_ReplaceExistingBlock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create .gitignore with stale Knossos block surrounded by user content.
	gitignorePath := filepath.Join(dir, ".gitignore")
	staleContent := "# My project\nnode_modules/\n\n# Knossos\n.knossos/\n.claude/CLAUDE.md\n# End Knossos\n\n# Custom rules\n*.log\n"
	if err := os.WriteFile(gitignorePath, []byte(staleContent), 0644); err != nil {
		t.Fatal(err)
	}

	writeProjectGitignore(dir)

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	content := string(data)

	// User content before and after preserved.
	if !strings.Contains(content, "node_modules/") {
		t.Error("user content before block was not preserved")
	}
	if !strings.Contains(content, "*.log") {
		t.Error("user content after block was not preserved")
	}
	// New patterns present (were missing in stale block).
	if !strings.Contains(content, "**/.ledge/*") {
		t.Error("updated patterns not present after replacement")
	}
	if !strings.Contains(content, "!**/.ledge/shelf/") {
		t.Error("shelf negation pattern not present after replacement")
	}
	// Only one Knossos block.
	if strings.Count(content, "# Knossos\n") != 1 {
		t.Errorf("expected exactly 1 Knossos block, got %d", strings.Count(content, "# Knossos\n"))
	}
}

func TestWriteProjectGitignore_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeProjectGitignore(dir)

	gitignorePath := filepath.Join(dir, ".gitignore")
	first, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}

	writeProjectGitignore(dir)

	second, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(first) != string(second) {
		t.Errorf("content changed on second run:\nfirst:  %q\nsecond: %q", string(first), string(second))
	}
}

// verifyLedgeSubdirs checks that .ledge subdirectories, .gitkeep files,
// and .gitignore policies were created correctly.
func verifyLedgeSubdirs(t *testing.T, projectDir string) {
	t.Helper()
	ledgeDir := filepath.Join(projectDir, ".ledge")

	// Verify subdirectories exist with .gitkeep.
	for _, sub := range []string{"decisions", "specs", "reviews", "spikes"} {
		subDir := filepath.Join(ledgeDir, sub)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			t.Errorf(".ledge/%s/ directory was not created", sub)
		}
		gitkeep := filepath.Join(subDir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			t.Errorf(".ledge/%s/.gitkeep was not created", sub)
		}
	}

	// Verify root .ledge/.gitignore exists and has expected content.
	rootGitignore := filepath.Join(ledgeDir, ".gitignore")
	data, err := os.ReadFile(rootGitignore)
	if err != nil {
		t.Fatalf("failed to read .ledge/.gitignore: %v", err)
	}
	rootContent := string(data)
	for _, pattern := range []string{"*.scratch", "*.tmp", "*.wip"} {
		if !strings.Contains(rootContent, pattern) {
			t.Errorf(".ledge/.gitignore missing pattern %q", pattern)
		}
	}

	// Verify .ledge/spikes/.gitignore exists with opt-in policy.
	spikesGitignore := filepath.Join(ledgeDir, "spikes", ".gitignore")
	data, err = os.ReadFile(spikesGitignore)
	if err != nil {
		t.Fatalf("failed to read .ledge/spikes/.gitignore: %v", err)
	}
	spikesContent := string(data)
	for _, pattern := range []string{"# Spike artifacts may be large", "*", "!.gitignore", "!.gitkeep"} {
		if !strings.Contains(spikesContent, pattern) {
			t.Errorf(".ledge/spikes/.gitignore missing pattern %q", pattern)
		}
	}

	// Verify .sos/land/ exists with .gitkeep.
	landDir := filepath.Join(projectDir, ".sos", "land")
	if _, err := os.Stat(landDir); os.IsNotExist(err) {
		t.Error(".sos/land/ directory was not created")
	}
	landGitkeep := filepath.Join(landDir, ".gitkeep")
	if _, err := os.Stat(landGitkeep); os.IsNotExist(err) {
		t.Error(".sos/land/.gitkeep was not created")
	}

	// Verify .ledge/shelf/ has mirrored category subdirectories.
	shelfDir := filepath.Join(ledgeDir, "shelf")
	for _, sub := range []string{"decisions", "specs", "reviews"} {
		subDir := filepath.Join(shelfDir, sub)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			t.Errorf(".ledge/shelf/%s/ directory was not created", sub)
		}
		gitkeep := filepath.Join(subDir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			t.Errorf(".ledge/shelf/%s/.gitkeep was not created", sub)
		}
	}

	// Verify root .gitignore has Knossos block.
	rootGitignore2 := filepath.Join(projectDir, ".gitignore")
	data, err = os.ReadFile(rootGitignore2)
	if err != nil {
		t.Fatalf("failed to read root .gitignore: %v", err)
	}
	rootContent2 := string(data)
	if !strings.Contains(rootContent2, "# Knossos") {
		t.Error("root .gitignore missing Knossos block")
	}
}
