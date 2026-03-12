package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/autom8y/knossos/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRite creates a minimal rite directory with manifest, workflow, and agent.
func setupRite(t *testing.T, ritesDir, riteName, workflowContent string, agents []Agent) {
	t.Helper()
	riteDir := filepath.Join(ritesDir, riteName)
	require.NoError(t, os.MkdirAll(filepath.Join(riteDir, "agents"), 0755))

	// Build manifest YAML
	var manifest strings.Builder
	manifest.WriteString("name: " + riteName + "\nversion: \"1.0\"\ndescription: test rite\nentry_agent: " + agents[0].Name + "\nagents:\n")
	for _, a := range agents {
		manifest.WriteString("  - name: " + a.Name + "\n    role: " + a.Role + "\n")
	}
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest.String()), 0644))

	if workflowContent != "" {
		require.NoError(t, os.WriteFile(filepath.Join(riteDir, "workflow.yaml"), []byte(workflowContent), 0644))
	}

	for _, a := range agents {
		content := "# " + a.Name + "\n\nRole: " + a.Role + "\n"
		require.NoError(t, os.WriteFile(filepath.Join(riteDir, "agents", a.Name+".md"), []byte(content), 0644))
	}
}

func TestRiteSwitchIntegration_StateConsistency(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	knossosDir := filepath.Join(projectDir, ".knossos")
	ritesDir := filepath.Join(knossosDir, "rites")

	// Create templates/rules with a known template rule
	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templatesDir, "rules", "internal-session.md"),
		[]byte("session rule v1"), 0644))

	// Create templates/sections for CLAUDE.md (required by inscription)
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	// Setup rite-a
	setupRite(t, ritesDir, "rite-a",
		"name: rite-a-workflow\nphases:\n  - design\n  - build\n",
		[]Agent{{Name: "designer", Role: "designs things"}})

	// Setup rite-b
	setupRite(t, ritesDir, "rite-b",
		"name: rite-b-workflow\nphases:\n  - deploy\n",
		[]Agent{{Name: "deployer", Role: "deploys things"}})

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	// Phase 1: Materialize rite-a
	result, err := m.MaterializeWithOptions("rite-a", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "project", result.Source)

	// Verify ACTIVE_RITE
	activeRite, err := os.ReadFile(filepath.Join(knossosDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "rite-a\n", string(activeRite))

	// Verify ACTIVE_WORKFLOW.yaml
	workflow, err := os.ReadFile(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(workflow), "rite-a-workflow")

	// Verify sync/state.json is valid (active_rite removed in PKG-008; ACTIVE_RITE file is authoritative)
	stateManager := sync.NewStateManager(resolver)
	state, err := stateManager.Load()
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.NotEmpty(t, state.LastSync)

	// Verify template rule was written
	ruleContent, err := os.ReadFile(filepath.Join(claudeDir, "rules", "internal-session.md"))
	require.NoError(t, err)
	assert.Equal(t, "session rule v1", string(ruleContent))

	// Simulate actions between rite switches:
	// 1. Create INVOCATION_STATE.yaml in .knossos/ (simulating `ari rite invoke`)
	require.NoError(t, os.MkdirAll(knossosDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(knossosDir, "INVOCATION_STATE.yaml"),
		[]byte("current_rite: rite-a\n"), 0644))
	// 2. Create a user rule
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "rules", "my-custom.md"),
		[]byte("my custom rule"), 0644))

	// Phase 2: Switch to rite-b
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	_, err = m2.MaterializeWithOptions("rite-b", Options{Force: true})
	require.NoError(t, err)

	// Verify ACTIVE_RITE = rite-b
	activeRite, err = os.ReadFile(filepath.Join(knossosDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "rite-b\n", string(activeRite))

	// Verify ACTIVE_WORKFLOW.yaml matches rite-b
	workflow, err = os.ReadFile(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(workflow), "rite-b-workflow")

	// Verify sync/state.json is still valid after rite switch
	state, err = stateManager.Load()
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.NotEmpty(t, state.LastSync)

	// Verify INVOCATION_STATE.yaml is gone from .knossos/
	_, err = os.Stat(filepath.Join(knossosDir, "INVOCATION_STATE.yaml"))
	assert.True(t, os.IsNotExist(err), "INVOCATION_STATE.yaml should be removed on rite switch")

	// Verify user rule survived
	userRule, err := os.ReadFile(filepath.Join(claudeDir, "rules", "my-custom.md"))
	require.NoError(t, err)
	assert.Equal(t, "my custom rule", string(userRule))

	// Verify template rule still exists (same template name, refreshed content)
	ruleContent, err = os.ReadFile(filepath.Join(claudeDir, "rules", "internal-session.md"))
	require.NoError(t, err)
	assert.Equal(t, "session rule v1", string(ruleContent))

	// Phase 3: Idempotency — materialize rite-b again with Force
	m3 := NewMaterializer(resolver)
	m3.templatesDir = templatesDir
	result2, err := m3.MaterializeWithOptions("rite-b", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "project", result2.Source)

	// Everything should be identical
	activeRite2, err := os.ReadFile(filepath.Join(knossosDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "rite-b\n", string(activeRite2))

	workflow2, err := os.ReadFile(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(workflow), string(workflow2))
}

func TestRiteSwitchIntegration_NoWorkflow(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	knossosDir := filepath.Join(projectDir, ".knossos")
	ritesDir := filepath.Join(knossosDir, "rites")

	// Create templates dir
	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	// Setup rite-a WITH workflow, and no-workflow rite WITHOUT
	setupRite(t, ritesDir, "has-workflow",
		"name: has-wf\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds"}})
	setupRite(t, ritesDir, "no-workflow", "",
		[]Agent{{Name: "worker", Role: "does work"}})

	resolver := paths.NewResolver(projectDir)

	// Phase 1: Materialize rite with workflow
	m1 := NewMaterializer(resolver)
	m1.templatesDir = templatesDir
	_, err := m1.MaterializeWithOptions("has-workflow", Options{Force: true})
	require.NoError(t, err)

	// Verify ACTIVE_WORKFLOW.yaml exists from rite with workflow
	_, err = os.Stat(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err, "ACTIVE_WORKFLOW.yaml should exist after materializing rite with workflow")

	// Phase 2: Switch to rite without workflow
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	_, err = m2.MaterializeWithOptions("no-workflow", Options{Force: true})
	require.NoError(t, err)

	// ACTIVE_WORKFLOW.yaml must be REMOVED (not stale from previous rite)
	_, err = os.Stat(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	assert.True(t, os.IsNotExist(err), "stale ACTIVE_WORKFLOW.yaml must be removed when switching to no-workflow rite")

	// ACTIVE_RITE should exist
	activeRite, err := os.ReadFile(filepath.Join(knossosDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "no-workflow\n", string(activeRite))
}

func TestRiteSwitchIntegration_EmbeddedSource(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	knossosDir := filepath.Join(projectDir, ".knossos")

	workflowContent := []byte("name: embedded-wf\nphases:\n  - test\n")
	agentContent := []byte("# tester\n\nRole: tests\n")
	manifestContent := []byte("name: embedded-rite\nversion: \"1.0\"\ndescription: test\nentry_agent: tester\nagents:\n  - name: tester\n    role: tests\n")

	embeddedFS := fstest.MapFS{
		"rites/embedded-rite/manifest.yaml":    &fstest.MapFile{Data: manifestContent},
		"rites/embedded-rite/workflow.yaml":    &fstest.MapFile{Data: workflowContent},
		"rites/embedded-rite/agents/tester.md": &fstest.MapFile{Data: agentContent},
	}
	// embeddedTemplates still includes a rules/ entry to prove the guard works --
	// materializeRules must skip rules for embedded source regardless of template content.
	embeddedTemplates := fstest.MapFS{
		"sections/.gitkeep":         &fstest.MapFile{Data: []byte{}},
		"rules/internal-session.md": &fstest.MapFile{Data: []byte("embedded session rule")},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver).
		WithEmbeddedFS(embeddedFS).
		WithEmbeddedTemplates(embeddedTemplates)

	// Need to set the source resolver to use embedded FS
	resolved := &ResolvedRite{
		Name:         "embedded-rite",
		RitePath:     "rites/embedded-rite",
		ManifestPath: "rites/embedded-rite/manifest.yaml",
		TemplatesDir: ".",
		Source:       RiteSource{Type: SourceEmbedded, Path: "embedded"},
	}

	// Directly test the sub-methods since MaterializeWithOptions goes through
	// source resolution which needs filesystem rites.
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.MkdirAll(knossosDir, 0755))

	// Test workflow materialization from embedded
	err := m.materializeWorkflow(knossosDir, resolved, provenance.NullCollector{}, "")
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(workflowContent), string(got))

	// Test rules materialization from embedded -- expect NO rules written.
	// Embedded rites are for foreign projects; knossos-internal rules (internal/**,
	// rites/**, etc.) are harmful noise on non-knossos codebases.
	err = m.materializeRules(claudeDir, resolved, provenance.NullCollector{}, "")
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(claudeDir, "rules", "internal-session.md"))
	assert.True(t, os.IsNotExist(err), "embedded rules must NOT be written to foreign projects")

	// Test agents materialization from embedded
	manifest := &RiteManifest{
		Name:       "embedded-rite",
		Agents:     []Agent{{Name: "tester", Role: "tests"}},
		EntryAgent: "tester",
	}
	err = m.materializeAgents(manifest, resolved.RitePath, claudeDir, resolved, provenance.NullCollector{}, nil, nil, "", "", nil)
	require.NoError(t, err)
	got, err = os.ReadFile(filepath.Join(claudeDir, "agents", "tester.md"))
	require.NoError(t, err)
	assert.Equal(t, string(agentContent), string(got))
}

func TestRiteSwitchIntegration_SyncStateJSON(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")

	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	setupRite(t, ritesDir, "state-test",
		"name: state-wf\n",
		[]Agent{{Name: "agent", Role: "works"}})

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	_, err := m.MaterializeWithOptions("state-test", Options{Force: true})
	require.NoError(t, err)

	// Read state.json and verify it's valid JSON with last_sync.
	// active_rite was removed from state.json (PKG-008): ACTIVE_RITE file is the authoritative store.
	stateData, err := os.ReadFile(filepath.Join(projectDir, ".knossos", "sync", "state.json"))
	require.NoError(t, err)

	var rawState map[string]any
	require.NoError(t, json.Unmarshal(stateData, &rawState))
	assert.NotEmpty(t, rawState["last_sync"])
	assert.NotContains(t, rawState, "active_rite", "active_rite must not be written to state.json")
}
