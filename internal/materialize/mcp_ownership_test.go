package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/paths"
)

func TestWriteMCPOwnership_CreatesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	err := writeMCPOwnership(dir, "ui", []string{"browserbase"})
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, mcpOwnershipFile))
	require.NoError(t, err)

	var ownership MCPOwnership
	require.NoError(t, json.Unmarshal(data, &ownership))
	assert.Equal(t, "ui", ownership.Rite)
	assert.Equal(t, []string{"browserbase"}, ownership.Servers)
}

func TestWriteMCPOwnership_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	err := writeMCPOwnership(dir, "ui", []string{"browserbase", "duckdb"})
	require.NoError(t, err)
	data1, _ := os.ReadFile(filepath.Join(dir, mcpOwnershipFile))

	err = writeMCPOwnership(dir, "ui", []string{"browserbase", "duckdb"})
	require.NoError(t, err)
	data2, _ := os.ReadFile(filepath.Join(dir, mcpOwnershipFile))

	assert.Equal(t, string(data1), string(data2), "writing twice should produce identical output")
}

func TestWriteMCPOwnership_SortsServerNames(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	err := writeMCPOwnership(dir, "ecosystem", []string{"terraform", "go-semantic"})
	require.NoError(t, err)

	ownership := loadMCPOwnership(dir)
	require.NotNil(t, ownership)
	assert.Equal(t, []string{"go-semantic", "terraform"}, ownership.Servers, "servers should be sorted")
}

func TestLoadMCPOwnership_NotFound(t *testing.T) {
	t.Parallel()
	ownership := loadMCPOwnership(t.TempDir())
	assert.Nil(t, ownership)
}

func TestLoadMCPOwnership_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, mcpOwnershipFile), []byte("not json"), 0644))

	ownership := loadMCPOwnership(dir)
	assert.Nil(t, ownership)
}

func TestPruneStaleMCPServers_RemovesPreviousRiteServers(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	knossosDir := filepath.Join(projectRoot, ".knossos")
	require.NoError(t, os.MkdirAll(knossosDir, 0755))

	// Write ownership from previous rite
	require.NoError(t, writeMCPOwnership(knossosDir, "ui", []string{"browserbase"}))

	// Write .mcp.json with the previous rite's server
	mcpJSON := `{"mcpServers":{"browserbase":{"command":"npx","args":["-y","stagehand-mcp-local"]},"github":{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}}}`
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, ".mcp.json"), []byte(mcpJSON), 0644))

	m := &Materializer{
		resolver: paths.NewResolver(projectRoot),
	}

	pruned := m.pruneStaleMCPServers(projectRoot)
	assert.Equal(t, 1, pruned)

	// Verify browserbase is gone, github remains
	data, err := os.ReadFile(filepath.Join(projectRoot, ".mcp.json"))
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))
	mcpServers := result["mcpServers"].(map[string]any)
	assert.NotContains(t, mcpServers, "browserbase", "previous rite's server should be removed")
	assert.Contains(t, mcpServers, "github", "satellite server should be preserved")
}

func TestPruneStaleMCPServers_PreservesSatelliteServers(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	knossosDir := filepath.Join(projectRoot, ".knossos")
	require.NoError(t, os.MkdirAll(knossosDir, 0755))

	// Ownership only includes "go-semantic"
	require.NoError(t, writeMCPOwnership(knossosDir, "ecosystem", []string{"go-semantic"}))

	// .mcp.json has go-semantic (rite-owned) + github (satellite)
	mcpJSON := `{"mcpServers":{"go-semantic":{"command":"go-semantic-mcp"},"github":{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}}}`
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, ".mcp.json"), []byte(mcpJSON), 0644))

	m := &Materializer{
		resolver: paths.NewResolver(projectRoot),
	}

	pruned := m.pruneStaleMCPServers(projectRoot)
	assert.Equal(t, 1, pruned)

	data, _ := os.ReadFile(filepath.Join(projectRoot, ".mcp.json"))
	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))
	mcpServers := result["mcpServers"].(map[string]any)
	assert.NotContains(t, mcpServers, "go-semantic")
	assert.Contains(t, mcpServers, "github", "satellite server must survive pruning")
}

func TestPruneStaleMCPServers_NoOwnershipFile(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	knossosDir := filepath.Join(projectRoot, ".knossos")
	require.NoError(t, os.MkdirAll(knossosDir, 0755))

	m := &Materializer{
		resolver: paths.NewResolver(projectRoot),
	}

	pruned := m.pruneStaleMCPServers(projectRoot)
	assert.Equal(t, 0, pruned, "no ownership file should mean no pruning")
}
