package initialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestInit_FreshDirectory(t *testing.T) {
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
	}

	// Run init with no rite (minimal scaffold).
	// We need to simulate running from the temp dir since runInit uses os.Getwd().
	// Instead, we directly call the materializer logic the same way.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	err = runInit(ctx, "", "", false, nil)
	if err != nil {
		t.Fatalf("runInit() returned error: %v", err)
	}

	// Verify .claude/ directory was created.
	claudeDir := filepath.Join(dir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Error(".claude/ directory was not created")
	}

	// Verify KNOSSOS_MANIFEST.yaml was created in .knossos/.
	manifestPath := filepath.Join(dir, ".knossos", "KNOSSOS_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("KNOSSOS_MANIFEST.yaml was not created in .knossos/")
	}

	// Verify CLAUDE.md was created.
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Error("CLAUDE.md was not created")
	}

	// Verify settings.local.json was created.
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
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
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	err = runInit(ctx, "test-rite", "", false, nil)
	if err != nil {
		t.Fatalf("runInit() with rite returned error: %v", err)
	}

	claudeDir := filepath.Join(dir, ".claude")

	// Verify agents were materialized.
	agentsDir := filepath.Join(claudeDir, "agents")
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
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Running without --force should succeed (exit 0) without modifying anything.
	err = runInit(ctx, "", "", false, nil)
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
	dir := t.TempDir()

	// Pre-create .knossos/ with a valid KNOSSOS_MANIFEST.yaml.
	knossosDir := filepath.Join(dir, ".knossos")
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
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
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Running with --force should reinitialize.
	err = runInit(ctx, "", "", true, nil)
	if err != nil {
		t.Fatalf("runInit() with --force returned error: %v", err)
	}

	// Verify CLAUDE.md was generated (proves re-materialization happened).
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Error("CLAUDE.md was not created after --force reinitialize")
	}

	// Verify settings.local.json was created.
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.local.json was not created after --force reinitialize")
	}
}

func TestInit_NonKnossosClaudeDir(t *testing.T) {
	dir := t.TempDir()

	// Pre-create .claude/ without KNOSSOS_MANIFEST.yaml (non-Knossos project).
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write some existing content (like a manual CLAUDE.md).
	if err := os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("# Custom"), 0644); err != nil {
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
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Should error without --force.
	err = runInit(ctx, "", "", false, nil)
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
	outputFlag := "text"
	verbose := false
	projectDir := ""

	cmd := NewInitCmd(&outputFlag, &verbose, &projectDir)
	if common.NeedsProject(cmd) {
		t.Error("init command should have needsProject=false, got true")
	}
}

func TestInit_LedgeIdempotent(t *testing.T) {
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

// withXDGDataDir redirects config.XDGDataDir to a temp directory by setting
// XDG_DATA_HOME, then restores the original value on cleanup.
func withXDGDataDir(t *testing.T) string {
	t.Helper()
	xdgBase := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdgBase)
	// config.XDGDataDir() reads XDG_DATA_HOME directly (no caching), so
	// the env var is sufficient -- no reset needed.
	return filepath.Join(xdgBase, "knossos", "mena")
}

// TestExtractEmbeddedMenaToXDG_FreshExtraction verifies that when neither the
// XDG mena directory nor the sentinel exist, extraction runs and the sentinel
// is written with the current version.
func TestExtractEmbeddedMenaToXDG_FreshExtraction(t *testing.T) {
	xdgMena := withXDGDataDir(t)
	common.SetBuildVersion("v1.0.0")
	defer common.SetBuildVersion("dev")

	embMena := makeTestMenaFS()
	extractEmbeddedMenaToXDG(embMena)

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
	xdgMena := withXDGDataDir(t)
	common.SetBuildVersion("v1.0.0")
	defer common.SetBuildVersion("dev")

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
	extractEmbeddedMenaToXDG(embMena)

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
	xdgMena := withXDGDataDir(t)
	common.SetBuildVersion("v2.0.0")
	defer common.SetBuildVersion("dev")

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
	extractEmbeddedMenaToXDG(embMena)

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
	xdgMena := withXDGDataDir(t)
	common.SetBuildVersion("v1.0.0")
	defer common.SetBuildVersion("dev")

	// Pre-create directory with content but NO sentinel.
	if err := os.MkdirAll(xdgMena, 0755); err != nil {
		t.Fatal(err)
	}
	stalePath := filepath.Join(xdgMena, "legacy-file.md")
	if err := os.WriteFile(stalePath, []byte("legacy"), 0644); err != nil {
		t.Fatal(err)
	}

	embMena := makeTestMenaFS()
	extractEmbeddedMenaToXDG(embMena)

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
}
