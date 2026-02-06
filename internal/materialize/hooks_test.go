package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
)

func TestBuildHooksSettings(t *testing.T) {
	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "PreToolUse", Matcher: "Edit|Write", Command: "ari hook writeguard --output json", Priority: 3},
			{Event: "PreToolUse", Matcher: "Bash", Command: "ari hook validate --output json", Priority: 5},
			{Event: "PostToolUse", Matcher: "Edit|Write|Bash", Command: "ari hook clew --output json", Priority: 5},
			{Event: "PostToolUse", Command: "ari hook budget --output json", Priority: 90},
			{Event: "SessionStart", Command: "ari hook context --output json", Priority: 5},
			{Event: "Stop", Command: "ari hook autopark --output json", Priority: 5},
			{Event: "UserPromptSubmit", Matcher: "^/", Command: "ari hook route --output json", Priority: 3},
		},
	}

	hooks := buildHooksSettings(cfg)

	// Check all event types are present
	expectedEvents := []string{"PreToolUse", "PostToolUse", "SessionStart", "Stop", "UserPromptSubmit"}
	for _, event := range expectedEvents {
		if hooks[event] == nil {
			t.Errorf("Missing event type %s", event)
		}
	}

	// Check PreToolUse has 2 entries ordered by priority (3 before 5)
	preToolUse, ok := hooks["PreToolUse"].([]map[string]any)
	if !ok {
		t.Fatalf("PreToolUse is not []map[string]any")
	}
	if len(preToolUse) != 2 {
		t.Fatalf("PreToolUse has %d entries, want 2", len(preToolUse))
	}
	if preToolUse[0]["command"] != "ari hook writeguard --output json" {
		t.Errorf("PreToolUse[0] command = %v, want writeguard (priority 3)", preToolUse[0]["command"])
	}
	if preToolUse[1]["command"] != "ari hook validate --output json" {
		t.Errorf("PreToolUse[1] command = %v, want validate (priority 5)", preToolUse[1]["command"])
	}

	// Check PostToolUse sorted: clew (5) before budget (90)
	postToolUse, ok := hooks["PostToolUse"].([]map[string]any)
	if !ok {
		t.Fatalf("PostToolUse is not []map[string]any")
	}
	if len(postToolUse) != 2 {
		t.Fatalf("PostToolUse has %d entries, want 2", len(postToolUse))
	}
	if postToolUse[0]["command"] != "ari hook clew --output json" {
		t.Errorf("PostToolUse[0] command = %v, want clew (priority 5)", postToolUse[0]["command"])
	}
	if postToolUse[1]["command"] != "ari hook budget --output json" {
		t.Errorf("PostToolUse[1] command = %v, want budget (priority 90)", postToolUse[1]["command"])
	}

	// Check matcher is included when present, absent when not
	if preToolUse[0]["matcher"] != "Edit|Write" {
		t.Errorf("PreToolUse[0] matcher = %v, want Edit|Write", preToolUse[0]["matcher"])
	}
	if postToolUse[1]["matcher"] != nil {
		t.Errorf("PostToolUse[1] should not have matcher, got %v", postToolUse[1]["matcher"])
	}
}

func TestBuildHooksSettings_SkipsEmptyCommand(t *testing.T) {
	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "PreToolUse", Command: "ari hook writeguard --output json"},
			{Event: "PreToolUse", Command: ""}, // Should be skipped
		},
	}

	hooks := buildHooksSettings(cfg)
	preToolUse, ok := hooks["PreToolUse"].([]map[string]any)
	if !ok {
		t.Fatalf("PreToolUse is not []map[string]any")
	}
	if len(preToolUse) != 1 {
		t.Errorf("Expected 1 entry (empty command skipped), got %d", len(preToolUse))
	}
}

func TestMergeHooksSettings_FreshSettings(t *testing.T) {
	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "PreToolUse", Command: "ari hook writeguard --output json", Priority: 3},
		},
	}

	settings := make(map[string]any)
	result := mergeHooksSettings(settings, cfg)

	hooks, ok := result["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks is not map[string]any")
	}

	preToolUse, ok := hooks["PreToolUse"].([]map[string]any)
	if !ok {
		t.Fatalf("PreToolUse is not []map[string]any")
	}
	if len(preToolUse) != 1 {
		t.Fatalf("Expected 1 hook entry, got %d", len(preToolUse))
	}
}

func TestMergeHooksSettings_PreservesUserHooks(t *testing.T) {
	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "PreToolUse", Command: "ari hook writeguard --output json", Priority: 3},
		},
	}

	// Simulate existing settings with a user-defined hook
	settings := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{"command": "my-custom-hook.sh", "matcher": "Bash"},
				map[string]any{"command": "ari hook writeguard --output json"}, // old ari hook
			},
		},
	}

	result := mergeHooksSettings(settings, cfg)

	hooks := result["hooks"].(map[string]any)
	preToolUse := hooks["PreToolUse"].([]map[string]any)

	// Should have 2 entries: new ari hook + preserved user hook
	if len(preToolUse) != 2 {
		t.Fatalf("Expected 2 entries (1 ari + 1 user), got %d", len(preToolUse))
	}

	// Ari hook should come first
	if preToolUse[0]["command"] != "ari hook writeguard --output json" {
		t.Errorf("First entry should be ari hook, got %v", preToolUse[0]["command"])
	}

	// User hook should be preserved
	if preToolUse[1]["command"] != "my-custom-hook.sh" {
		t.Errorf("Second entry should be user hook, got %v", preToolUse[1]["command"])
	}
}

func TestMergeHooksSettings_RemovesOldAriHooks(t *testing.T) {
	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "PostToolUse", Command: "ari hook budget --output json", Priority: 90},
		},
	}

	// Simulate existing settings with old ari hooks
	settings := map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{"command": "ari hook clew --output json"},    // old, will be replaced
				map[string]any{"command": "ari hook budget --output json"},  // old, will be replaced
			},
		},
	}

	result := mergeHooksSettings(settings, cfg)

	hooks := result["hooks"].(map[string]any)
	postToolUse := hooks["PostToolUse"].([]map[string]any)

	// Should have only 1 entry (the new budget hook, old clew removed since not in new config)
	if len(postToolUse) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(postToolUse))
	}
	if postToolUse[0]["command"] != "ari hook budget --output json" {
		t.Errorf("Expected budget hook, got %v", postToolUse[0]["command"])
	}
}

func TestMergeHooksSettings_Idempotent(t *testing.T) {
	cfg := &HooksConfig{
		SchemaVersion: "2.0",
		Hooks: []HookEntry{
			{Event: "PreToolUse", Matcher: "Edit|Write", Command: "ari hook writeguard --output json", Priority: 3},
			{Event: "PostToolUse", Command: "ari hook budget --output json", Priority: 90},
		},
	}

	settings := make(map[string]any)
	result1 := mergeHooksSettings(settings, cfg)

	// Serialize to JSON and back (simulates load/save cycle)
	data, _ := json.Marshal(result1)
	var settings2 map[string]any
	json.Unmarshal(data, &settings2)

	result2 := mergeHooksSettings(settings2, cfg)

	// Marshal both and compare
	data1, _ := json.MarshalIndent(result1, "", "  ")
	data2, _ := json.MarshalIndent(result2, "", "  ")

	if string(data1) != string(data2) {
		t.Errorf("Merge is not idempotent.\nFirst:\n%s\nSecond:\n%s", data1, data2)
	}
}

func TestLoadHooksConfig(t *testing.T) {
	// Create a temp directory structure with hooks.yaml
	tmpDir := t.TempDir()
	hooksDir := filepath.Join(tmpDir, "user-hooks", "ari")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	hooksYAML := `schema_version: "2.0"
hooks:
  - event: PreToolUse
    matcher: "Edit|Write"
    command: "ari hook writeguard --output json"
    timeout: 3
    priority: 3
    description: "Guards writes"
`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.yaml"), []byte(hooksYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a materializer pointing at the temp dir
	resolver := paths.NewResolver(tmpDir)
	mat := NewMaterializer(resolver)

	// Override knossosHome to point at tmpDir via env var
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", tmpDir)
	t.Cleanup(config.ResetKnossosHome)

	cfg := mat.loadHooksConfig()
	if cfg == nil {
		t.Fatal("Expected hooks config, got nil")
	}

	if cfg.SchemaVersion != "2.0" {
		t.Errorf("SchemaVersion = %q, want 2.0", cfg.SchemaVersion)
	}
	if len(cfg.Hooks) != 1 {
		t.Fatalf("Expected 1 hook entry, got %d", len(cfg.Hooks))
	}
	if cfg.Hooks[0].Command != "ari hook writeguard --output json" {
		t.Errorf("Command = %q, want ari hook writeguard", cfg.Hooks[0].Command)
	}
}

func TestLoadHooksConfig_RejectsV1Schema(t *testing.T) {
	tmpDir := t.TempDir()
	hooksDir := filepath.Join(tmpDir, "user-hooks", "ari")
	os.MkdirAll(hooksDir, 0755)

	// v1 schema (no command field, has path field)
	hooksYAML := `schema_version: "1.0"
hooks:
  - event: PreToolUse
    path: ari/writeguard.sh
    timeout: 3
`
	os.WriteFile(filepath.Join(hooksDir, "hooks.yaml"), []byte(hooksYAML), 0644)

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", tmpDir)
	t.Cleanup(config.ResetKnossosHome)

	resolver := paths.NewResolver(tmpDir)
	mat := NewMaterializer(resolver)

	cfg := mat.loadHooksConfig()
	if cfg != nil {
		t.Error("Expected nil for v1 schema, got config")
	}
}

func TestLoadHooksConfig_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", tmpDir)
	t.Cleanup(config.ResetKnossosHome)

	resolver := paths.NewResolver(tmpDir)
	mat := NewMaterializer(resolver)

	cfg := mat.loadHooksConfig()
	if cfg != nil {
		t.Error("Expected nil when no hooks.yaml exists")
	}
}
