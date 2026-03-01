package initialize

import (
	"os"
	"path/filepath"
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

	// Verify KNOSSOS_MANIFEST.yaml was created.
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("KNOSSOS_MANIFEST.yaml was not created")
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
	activeRitePath := filepath.Join(claudeDir, "ACTIVE_RITE")
	data, err := os.ReadFile(activeRitePath)
	if err != nil {
		t.Fatalf("failed to read ACTIVE_RITE: %v", err)
	}
	if got := string(data); got != "test-rite\n" {
		t.Errorf("ACTIVE_RITE = %q, want %q", got, "test-rite\n")
	}

	// Verify KNOSSOS_MANIFEST.yaml was created.
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("KNOSSOS_MANIFEST.yaml was not created")
	}

	// Verify project-level directories were scaffolded.
	for _, dirName := range []string{".knossos", ".sos", ".ledge"} {
		dirPath := filepath.Join(dir, dirName)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("%s/ directory was not created", dirName)
		}
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()

	// Pre-create the .claude/ with a valid KNOSSOS_MANIFEST.yaml to simulate existing init.
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
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

	// Pre-create the .claude/ with a valid KNOSSOS_MANIFEST.yaml.
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
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
