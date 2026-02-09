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
	tempDir := t.TempDir()
	manifest := &RiteManifest{
		Name: "test-rite",
	}

	err := (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify settings file was created
	settingsPath := filepath.Join(tempDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	// Should have hooks
	assert.NotNil(t, settings["hooks"])

	// Should NOT have mcpServers (not needed if manifest has none)
	// Actually, we DO add it because mergeMCPServers always ensures it exists
	// But it should be empty
	if mcpServers, ok := settings["mcpServers"].(map[string]any); ok {
		assert.Empty(t, mcpServers)
	}
}

// TestMaterializeSettingsWithManifest_WithMCPServers tests that MCP servers
// from manifest are written to settings.local.json.
func TestMaterializeSettingsWithManifest_WithMCPServers(t *testing.T) {
	tempDir := t.TempDir()
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

	err := (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify settings file was created
	settingsPath := filepath.Join(tempDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	// Check mcpServers structure
	assert.NotNil(t, settings["mcpServers"])
	mcpServers, ok := settings["mcpServers"].(map[string]any)
	require.True(t, ok)

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

// TestMaterializeSettingsWithManifest_PreservesExisting tests that existing
// satellite-owned MCP servers are preserved when merging rite manifest servers.
func TestMaterializeSettingsWithManifest_PreservesExisting(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	// Create existing settings with satellite-owned server
	existingSettings := map[string]any{
		"hooks": map[string]any{},
		"mcpServers": map[string]any{
			"custom-satellite-server": map[string]any{
				"command": "custom-cmd",
				"args":    []string{"--custom"},
			},
		},
	}

	data, err := json.MarshalIndent(existingSettings, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	// Now materialize with manifest containing different servers
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

	err = (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify both servers exist
	data, err = os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	mcpServers, ok := settings["mcpServers"].(map[string]any)
	require.True(t, ok)

	// Both servers should be present
	assert.Contains(t, mcpServers, "custom-satellite-server")
	assert.Contains(t, mcpServers, "github")

	// Custom satellite server should be unchanged
	customServer, ok := mcpServers["custom-satellite-server"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "custom-cmd", customServer["command"])
}

// TestMaterializeSettingsWithManifest_UpdatesRiteOwnedServer tests that
// if a rite manifest server is updated, the settings reflect the update.
func TestMaterializeSettingsWithManifest_UpdatesRiteOwnedServer(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	// Create existing settings with old github config
	existingSettings := map[string]any{
		"hooks": map[string]any{},
		"mcpServers": map[string]any{
			"github": map[string]any{
				"command": "old-command",
			},
		},
	}

	data, err := json.MarshalIndent(existingSettings, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	// Materialize with updated github config
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
		},
	}

	err = (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify github server was updated
	data, err = os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	mcpServers, ok := settings["mcpServers"].(map[string]any)
	require.True(t, ok)

	githubServer, ok := mcpServers["github"].(map[string]any)
	require.True(t, ok)

	// Should have new config
	assert.Equal(t, "npx", githubServer["command"])
	assert.Equal(t, []any{"-y", "@modelcontextprotocol/server-github"}, githubServer["args"])
	assert.NotNil(t, githubServer["env"])
}

// TestMaterializeSettingsWithManifest_NilManifest tests that passing nil manifest
// creates minimal settings without error.
func TestMaterializeSettingsWithManifest_NilManifest(t *testing.T) {
	tempDir := t.TempDir()

	err := (&Materializer{}).materializeSettingsWithManifest(tempDir, nil, provenance.NullCollector{})
	require.NoError(t, err)

	// Verify settings file was created with hooks
	settingsPath := filepath.Join(tempDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	assert.NotNil(t, settings["hooks"])
}
