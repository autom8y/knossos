package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/provenance"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaterializeSettingsWithManifest_NoMCPServers tests that settings are created
// with minimal content when manifest has no MCP servers.
func TestMaterializeSettingsWithManifest_NoMCPServers(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	manifest := &RiteManifest{
		Name: "test-rite",
	}

	err := (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{}, "claude")
	require.NoError(t, err)

	// Verify settings file was created
	settingsPath := filepath.Join(tempDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	// Should have hooks
	assert.NotNil(t, settings["hooks"])

	// SCAR-028: mcpServers must NOT be in settings.local.json
	assert.Nil(t, settings["mcpServers"], "mcpServers must not be in settings.local.json (SCAR-028)")
}

// TestMaterializeSettingsWithManifest_StaleMcpServersRemoved tests that
// stale mcpServers in settings.local.json are cleaned up (SCAR-028).
func TestMaterializeSettingsWithManifest_StaleMcpServersRemoved(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	// Create existing settings WITH stale mcpServers (from before SCAR-028 fix)
	existingSettings := map[string]any{
		"hooks": map[string]any{},
		"mcpServers": map[string]any{
			"github": map[string]any{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-github"},
			},
		},
	}

	data, err := json.MarshalIndent(existingSettings, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	manifest := &RiteManifest{Name: "test-rite"}
	err = (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{}, "claude")
	require.NoError(t, err)

	// Verify mcpServers was removed
	data, err = os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))
	assert.Nil(t, settings["mcpServers"], "stale mcpServers must be removed from settings.local.json (SCAR-028)")
	assert.NotNil(t, settings["hooks"], "hooks must be preserved")
}

// TestMaterializeMcpJson_WritesToProjectRoot tests that MCP servers from
// the rite manifest are written to .mcp.json at project root (SCAR-028).
func TestMaterializeMcpJson_WritesToProjectRoot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	manifest := &RiteManifest{
		Name: "test-rite",
		MCPServers: []MCPServer{
			{
				Name:    "github",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
				Env: map[string]string{
					"GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}",
				},
			},
			{
				Name:    "terraform",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-terraform"},
			},
		},
	}

	err := (&Materializer{}).materializeMcpJson(projectRoot, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify .mcp.json was created
	mcpJsonPath := filepath.Join(projectRoot, ".mcp.json")
	data, err := os.ReadFile(mcpJsonPath)
	require.NoError(t, err)

	var mcpFile map[string]any
	require.NoError(t, json.Unmarshal(data, &mcpFile))

	mcpServers, ok := mcpFile["mcpServers"].(map[string]any)
	require.True(t, ok, "mcpServers key must exist in .mcp.json")

	// Check github server
	assert.Contains(t, mcpServers, "github")
	githubServer, ok := mcpServers["github"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "npx", githubServer["command"])
	assert.Equal(t, []any{"-y", "@modelcontextprotocol/server-github"}, githubServer["args"])

	githubEnv, ok := githubServer["env"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "${GITHUB_TOKEN}", githubEnv["GITHUB_PERSONAL_ACCESS_TOKEN"])

	// Check terraform server
	assert.Contains(t, mcpServers, "terraform")
	terraformServer, ok := mcpServers["terraform"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "npx", terraformServer["command"])
}

// TestMaterializeMcpJson_PreservesExistingSatelliteServers tests union merge:
// rite servers are added, satellite servers in .mcp.json are preserved.
func TestMaterializeMcpJson_PreservesExistingSatelliteServers(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	mcpJsonPath := filepath.Join(projectRoot, ".mcp.json")

	// Create existing .mcp.json with a user-owned satellite server
	existing := map[string]any{
		"mcpServers": map[string]any{
			"my-custom-server": map[string]any{
				"command": "my-server-binary",
				"args":    []any{"--port", "8080"},
			},
		},
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(mcpJsonPath, data, 0644))

	// Materialize rite MCP servers
	manifest := &RiteManifest{
		Name: "test-rite",
		MCPServers: []MCPServer{
			{Name: "duckdb", Command: "uvx", Args: []string{"mcp-server-motherduck"}},
		},
	}

	err = (&Materializer{}).materializeMcpJson(projectRoot, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify both servers exist
	data, err = os.ReadFile(mcpJsonPath)
	require.NoError(t, err)

	var mcpFile map[string]any
	require.NoError(t, json.Unmarshal(data, &mcpFile))

	mcpServers, ok := mcpFile["mcpServers"].(map[string]any)
	require.True(t, ok)

	assert.Contains(t, mcpServers, "my-custom-server", "satellite server must be preserved")
	assert.Contains(t, mcpServers, "duckdb", "rite server must be added")

	// Satellite server should be unchanged
	customServer, ok := mcpServers["my-custom-server"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "my-server-binary", customServer["command"])
}

// TestMaterializeMcpJson_NilManifest tests that nil manifest is a no-op.
func TestMaterializeMcpJson_NilManifest(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()

	err := (&Materializer{}).materializeMcpJson(projectRoot, nil, provenance.NullCollector{})
	require.NoError(t, err)

	// .mcp.json should NOT be created
	mcpJsonPath := filepath.Join(projectRoot, ".mcp.json")
	_, err = os.Stat(mcpJsonPath)
	assert.True(t, os.IsNotExist(err), ".mcp.json must not be created when manifest is nil")
}

// TestMaterializeMcpJson_EmptyMCPServers tests that empty MCP servers list is a no-op.
func TestMaterializeMcpJson_EmptyMCPServers(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	manifest := &RiteManifest{Name: "test-rite"}

	err := (&Materializer{}).materializeMcpJson(projectRoot, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// .mcp.json should NOT be created
	mcpJsonPath := filepath.Join(projectRoot, ".mcp.json")
	_, err = os.Stat(mcpJsonPath)
	assert.True(t, os.IsNotExist(err), ".mcp.json must not be created when no MCP servers declared")
}

// TestMaterializeMcpJson_UpdatesExistingRiteServer tests that rite-owned
// servers are updated when the manifest changes.
func TestMaterializeMcpJson_UpdatesExistingRiteServer(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	mcpJsonPath := filepath.Join(projectRoot, ".mcp.json")

	// Create existing .mcp.json with old github config
	existing := map[string]any{
		"mcpServers": map[string]any{
			"github": map[string]any{
				"command": "old-command",
			},
		},
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(mcpJsonPath, data, 0644))

	// Materialize with updated github config
	manifest := &RiteManifest{
		Name: "test-rite",
		MCPServers: []MCPServer{
			{
				Name:    "github",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			},
		},
	}

	err = (&Materializer{}).materializeMcpJson(projectRoot, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	data, err = os.ReadFile(mcpJsonPath)
	require.NoError(t, err)

	var mcpFile map[string]any
	require.NoError(t, json.Unmarshal(data, &mcpFile))

	mcpServers, ok := mcpFile["mcpServers"].(map[string]any)
	require.True(t, ok)

	githubServer, ok := mcpServers["github"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "npx", githubServer["command"])
	assert.Equal(t, []any{"-y", "@modelcontextprotocol/server-github"}, githubServer["args"])
}

// TestSCAR028_MCPServers_NotInSettingsLocalJson is a SCAR regression test
// verifying that MCP servers are never written to settings.local.json.
func TestSCAR028_MCPServers_NotInSettingsLocalJson(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	manifest := &RiteManifest{
		Name: "test-rite",
		MCPServers: []MCPServer{
			{Name: "duckdb", Command: "uvx", Args: []string{"mcp-server-motherduck"}},
		},
	}

	err := (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{}, "claude")
	require.NoError(t, err)

	settingsPath := filepath.Join(tempDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	assert.Nil(t, settings["mcpServers"],
		"SCAR-028: mcpServers must NEVER be in settings.local.json — CC reads from .mcp.json")
}
