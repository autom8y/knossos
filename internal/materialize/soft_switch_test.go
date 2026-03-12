package materialize

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

// TestMaterializeWithOptions_SoftMode_AgentsUpdated verifies that agents/ directory is updated in soft mode.
func TestMaterializeWithOptions_SoftMode_AgentsUpdated(t *testing.T) {
	t.Parallel()
	// Setup test directory
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create test rite with an agent
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "test-rite")
	if err := os.MkdirAll(filepath.Join(riteDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create manifest
	manifestContent := `name: test-rite
version: 1.0.0
description: Test rite for soft mode
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test agent role
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create agent file
	agentContent := "# Test Agent\n\nTest agent content\n"
	if err := os.WriteFile(filepath.Join(riteDir, "agents", "test-agent.md"), []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create materializer with explicit source
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	// Execute soft mode sync
	opts := Options{
		Soft: true,
	}
	result, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify soft mode was recorded
	if !result.SoftMode {
		t.Error("Expected SoftMode=true, got false")
	}

	// Verify agents directory was created and contains the agent
	agentPath := filepath.Join(claudeDir, "agents", "test-agent.md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Error("Expected agent file to be created in soft mode")
	}

	// Verify agent content matches
	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("Failed to read agent file: %v", err)
	}
	if string(content) != agentContent {
		t.Errorf("Agent content mismatch.\nExpected: %s\nGot: %s", agentContent, string(content))
	}
}

// TestMaterializeWithOptions_SoftMode_InscriptionUpdated verifies that the inscription (context file) is updated in soft mode.
func TestMaterializeWithOptions_SoftMode_InscriptionUpdated(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create test rite
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "test-rite")
	if err := os.MkdirAll(filepath.Join(riteDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `name: test-rite
version: 1.0.0
description: Test rite
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test role
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	opts := Options{Soft: true}
	_, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify CLAUDE.md was created
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Error("Expected CLAUDE.md to be created in soft mode")
	}
}

// TestMaterializeWithOptions_SoftMode_MenaSkipped verifies that commands/ and skills/ are NOT updated in soft mode.
func TestMaterializeWithOptions_SoftMode_MenaSkipped(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Pre-create commands and skills directories with marker files
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	markerContent := "ORIGINAL_CONTENT"
	commandMarker := filepath.Join(commandsDir, ".marker")
	skillMarker := filepath.Join(skillsDir, ".marker")
	if err := os.WriteFile(commandMarker, []byte(markerContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillMarker, []byte(markerContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test rite with mena
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "test-rite")
	menaDirs := []string{
		filepath.Join(riteDir, "mena", "test-command"),
		filepath.Join(riteDir, "mena", "test-skill"),
	}
	for _, dir := range menaDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create INDEX files with routing
	if err := os.WriteFile(filepath.Join(riteDir, "mena", "test-command", "INDEX.dro.md"), []byte("# Test Command"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(riteDir, "mena", "test-skill", "INDEX.lego.md"), []byte("# Test Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	manifestContent := `name: test-rite
version: 1.0.0
description: Test rite
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test role
dromena:
  - test-command
legomena:
  - test-skill
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	opts := Options{Soft: true}
	result, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify mena stage was deferred
	if !contains(result.DeferredStages, "mena") {
		t.Error("Expected 'mena' in DeferredStages")
	}

	// Verify marker files still exist with original content
	content, err := os.ReadFile(commandMarker)
	if err != nil {
		t.Fatalf("Command marker file was deleted in soft mode: %v", err)
	}
	if string(content) != markerContent {
		t.Error("Command marker file was modified in soft mode")
	}

	content, err = os.ReadFile(skillMarker)
	if err != nil {
		t.Fatalf("Skill marker file was deleted in soft mode: %v", err)
	}
	if string(content) != markerContent {
		t.Error("Skill marker file was modified in soft mode")
	}
}

// TestMaterializeWithOptions_SoftMode_SettingsSkipped verifies that settings.local.json is NOT updated in soft mode.
func TestMaterializeWithOptions_SoftMode_SettingsSkipped(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-create settings with marker content
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	originalSettings := `{"marker":"ORIGINAL"}`
	if err := os.WriteFile(settingsPath, []byte(originalSettings), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test rite
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "test-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `name: test-rite
version: 1.0.0
description: Test rite
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test role
mcp_servers:
  - name: test-server
    command: test
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	opts := Options{Soft: true}
	result, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify settings stage was deferred
	if !contains(result.DeferredStages, "settings") {
		t.Error("Expected 'settings' in DeferredStages")
	}

	// Verify settings file still has original content
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Settings file read failed: %v", err)
	}
	if string(content) != originalSettings {
		t.Errorf("Settings file was modified in soft mode.\nExpected: %s\nGot: %s", originalSettings, string(content))
	}
}

// TestMaterializeWithOptions_SoftMode_ActiveRiteUpdated verifies that ACTIVE_RITE is updated in soft mode.
func TestMaterializeWithOptions_SoftMode_ActiveRiteUpdated(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create test rite
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "new-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `name: new-rite
version: 1.0.0
description: New rite
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test role
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	opts := Options{Soft: true}
	_, err := m.MaterializeWithOptions("new-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify ACTIVE_RITE was updated
	activeRitePath := filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE")
	content, err := os.ReadFile(activeRitePath)
	if err != nil {
		t.Fatalf("ACTIVE_RITE file not created: %v", err)
	}

	expected := "new-rite\n"
	if string(content) != expected {
		t.Errorf("ACTIVE_RITE content mismatch.\nExpected: %q\nGot: %q", expected, string(content))
	}
}

// TestMaterializeWithOptions_SoftMode_ResultReporting verifies that result correctly reports soft mode status.
func TestMaterializeWithOptions_SoftMode_ResultReporting(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create minimal test rite
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "test-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `name: test-rite
version: 1.0.0
description: Test rite
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test role
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	opts := Options{Soft: true}
	result, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify SoftMode flag
	if !result.SoftMode {
		t.Error("Expected SoftMode=true, got false")
	}

	// Verify DeferredStages has exactly 4 entries
	expectedStages := []string{"mena", "rules", "settings", "workflow"}
	if len(result.DeferredStages) != len(expectedStages) {
		t.Errorf("Expected %d deferred stages, got %d", len(expectedStages), len(result.DeferredStages))
	}

	// Verify all expected stages are present
	for _, stage := range expectedStages {
		if !contains(result.DeferredStages, stage) {
			t.Errorf("Expected '%s' in DeferredStages, but it was missing", stage)
		}
	}

	// Verify status is still success
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", result.Status)
	}
}

// TestMaterializeWithOptions_FullMode_Unchanged verifies that full mode still runs all stages.
func TestMaterializeWithOptions_FullMode_Unchanged(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create test rite with mena
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "test-rite")
	menaDir := filepath.Join(riteDir, "mena", "test-command")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mena INDEX file
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.dro.md"), []byte("# Test Command"), 0644); err != nil {
		t.Fatal(err)
	}

	manifestContent := `name: test-rite
version: 1.0.0
description: Test rite
entry_agent: test-agent
agents:
  - name: test-agent
    role: Test role
dromena:
  - test-command
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializerWithSource(resolver, filepath.Dir(riteDir))
	m.claudeDirOverride = claudeDir

	// Execute WITHOUT soft mode (full mode)
	opts := Options{Soft: false}
	result, err := m.MaterializeWithOptions("test-rite", opts)
	if err != nil {
		t.Fatalf("MaterializeWithOptions failed: %v", err)
	}

	// Verify soft mode was NOT used
	if result.SoftMode {
		t.Error("Expected SoftMode=false in full mode, got true")
	}

	// Verify DeferredStages is empty
	if len(result.DeferredStages) > 0 {
		t.Errorf("Expected no deferred stages in full mode, got %v", result.DeferredStages)
	}

	// Verify mena was materialized (promoted command file should exist)
	commandFile := filepath.Join(claudeDir, "commands", "test-command.md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		t.Error("Expected promoted command file to be created in full mode")
	}

	// Verify settings was created
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Expected settings.local.json to be created in full mode")
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
