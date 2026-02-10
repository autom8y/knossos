package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaterializeCrossRiteAgents_Basic verifies cross-rite agents from top-level
// agents/ are materialized to .claude/agents/.
func TestMaterializeCrossRiteAgents_Basic(t *testing.T) {
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

	// Verify cross-rite agents were written
	assert.FileExists(t, filepath.Join(agentsDir, "moirai.md"))
	content, err := os.ReadFile(filepath.Join(agentsDir, "moirai.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Moirai\n", string(content))

	assert.FileExists(t, filepath.Join(agentsDir, "consultant.md"))
}

// TestMaterializeCrossRiteAgents_RitePriority verifies that rite-scoped agents
// take priority over cross-rite agents of the same name.
func TestMaterializeCrossRiteAgents_RitePriority(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, "rites")
	claudeDir := filepath.Join(projectDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")

	// Setup a rite that has an agent named "orchestrator"
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "orchestrator", Role: "coordinates"}})

	// Also create a cross-rite agent named "orchestrator" (should be skipped)
	crossRiteDir := filepath.Join(projectDir, "agents")
	require.NoError(t, os.MkdirAll(crossRiteDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "orchestrator.md"), []byte("# Cross-Rite Orchestrator\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "moirai.md"), []byte("# Moirai\n"), 0644))

	// Materialize
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	// Verify rite's orchestrator was written (NOT the cross-rite one)
	content, err := os.ReadFile(filepath.Join(agentsDir, "orchestrator.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "coordinates", "rite agent content should be used")
	assert.NotContains(t, string(content), "Cross-Rite", "cross-rite agent should NOT override rite agent")

	// Verify non-conflicting cross-rite agent was still written
	assert.FileExists(t, filepath.Join(agentsDir, "moirai.md"))
}

// TestMaterializeCrossRiteAgents_Provenance verifies cross-rite agents get
// provenance entries recorded.
func TestMaterializeCrossRiteAgents_Provenance(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, "rites")
	claudeDir := filepath.Join(projectDir, ".claude")

	// Setup a rite
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "designer", Role: "designs"}})

	// Create cross-rite agent
	crossRiteDir := filepath.Join(projectDir, "agents")
	require.NoError(t, os.MkdirAll(crossRiteDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "moirai.md"), []byte("# Moirai\n"), 0644))

	// Materialize
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	// Load provenance manifest and check cross-rite agent entry
	manifestPath := provenance.ManifestPath(claudeDir)
	manifest, err := provenance.Load(manifestPath)
	require.NoError(t, err)

	entry, exists := manifest.Entries["agents/moirai.md"]
	require.True(t, exists, "cross-rite agent should have provenance entry")
	assert.Equal(t, provenance.OwnerKnossos, entry.Owner)
	assert.Equal(t, "agents/moirai.md", entry.SourcePath)
	assert.NotEmpty(t, entry.Checksum)
}

// TestMaterializeCrossRiteAgents_NoCrossRiteDir verifies graceful handling
// when no top-level agents/ directory exists.
func TestMaterializeCrossRiteAgents_NoCrossRiteDir(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, "rites")

	// Setup a rite (no top-level agents/ created)
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "designer", Role: "designs"}})

	// Materialize — should succeed without error
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true, KeepAll: true})
	require.NoError(t, err)
}

// TestListCrossRiteAgents verifies the listing of cross-rite agent filenames.
func TestListCrossRiteAgents(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, "rites")

	// Setup rite for resolution
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "designer", Role: "designs"}})

	// Create cross-rite agents dir with mixed files
	crossRiteDir := filepath.Join(projectDir, "agents")
	require.NoError(t, os.MkdirAll(crossRiteDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "moirai.md"), []byte("# Moirai\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "consultant.md"), []byte("# Consultant\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(crossRiteDir, "README.txt"), []byte("not an agent"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(crossRiteDir, "subdir"), 0755))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// Resolve rite to get ResolvedRite
	resolved, err := m.sourceResolver.ResolveRite("test-rite", "")
	require.NoError(t, err)

	agents := m.listCrossRiteAgents(resolved)
	assert.Contains(t, agents, "moirai.md")
	assert.Contains(t, agents, "consultant.md")
	assert.NotContains(t, agents, "README.txt", "non-.md files should be excluded")
	assert.Len(t, agents, 2)
}
