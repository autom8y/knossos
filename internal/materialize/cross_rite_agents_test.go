package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCrossRiteAgents_NotMaterializedToProject verifies cross-rite agents from
// top-level agents/ are NOT copied to project .claude/agents/.
// Cross-rite agents are user-scope owned (synced to ~/.claude/agents/ by user-scope sync).
func TestCrossRiteAgents_NotMaterializedToProject(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, "rites")
	claudeDir := filepath.Join(projectDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")

	// Setup a rite with one agent
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "designer", Role: "designs"}})

	// Create cross-rite agents in top-level agents/
	crossRiteDir := filepath.Join(projectDir, "agents")
	require.NoError(t, os.MkdirAll(crossRiteDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "moirai.md"), []byte("# Moirai\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "consultant.md"), []byte("# Consultant\n"), 0644))

	// Materialize
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	// Verify rite agent was written
	assert.FileExists(t, filepath.Join(agentsDir, "designer.md"))

	// Verify cross-rite agents were NOT written to project level
	assert.NoFileExists(t, filepath.Join(agentsDir, "moirai.md"),
		"cross-rite agents should not be materialized to project .claude/agents/")
	assert.NoFileExists(t, filepath.Join(agentsDir, "consultant.md"),
		"cross-rite agents should not be materialized to project .claude/agents/")
}

// TestCrossRiteAgents_OrphanedOnRiteSwitch verifies that previously-materialized
// cross-rite agents at project level are detected as orphans after this change.
func TestCrossRiteAgents_OrphanedOnRiteSwitch(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, "rites")
	claudeDir := filepath.Join(projectDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")

	// Setup a rite
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "designer", Role: "designs"}})

	// Simulate stale cross-rite agents left from a previous sync
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "moirai.md"), []byte("# Moirai\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "consultant.md"), []byte("# Consultant\n"), 0644))

	// Materialize with RemoveAll orphan strategy
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	result, err := m.MaterializeWithOptions("test-rite", Options{Force: true, RemoveAll: true})
	require.NoError(t, err)

	// Stale cross-rite agents should be detected as orphans
	assert.Contains(t, result.OrphansDetected, "moirai.md")
	assert.Contains(t, result.OrphansDetected, "consultant.md")

	// And removed from project level
	assert.NoFileExists(t, filepath.Join(agentsDir, "moirai.md"))
	assert.NoFileExists(t, filepath.Join(agentsDir, "consultant.md"))

	// Rite agent preserved
	assert.FileExists(t, filepath.Join(agentsDir, "designer.md"))
}
