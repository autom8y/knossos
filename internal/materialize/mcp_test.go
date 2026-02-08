package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeMCPServers_EmptySettings(t *testing.T) {
	settings := make(map[string]any)
	mcpServers := []MCPServer{
		{
			Name:    "github",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			Env: map[string]string{
				"GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}",
			},
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	// Check structure
	assert.NotNil(t, result["mcpServers"])
	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	// Check github server
	assert.Contains(t, mcpMap, "github")
	githubServer, ok := mcpMap["github"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "npx", githubServer["command"])
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-github"}, githubServer["args"])

	envMap, ok := githubServer["env"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "${GITHUB_TOKEN}", envMap["GITHUB_PERSONAL_ACCESS_TOKEN"])
}

func TestMergeMCPServers_PreservesExistingServers(t *testing.T) {
	settings := map[string]any{
		"mcpServers": map[string]any{
			"custom-server": map[string]any{
				"command": "custom-cmd",
				"args":    []string{"--flag"},
			},
		},
	}

	mcpServers := []MCPServer{
		{
			Name:    "terraform",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-terraform"},
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	// Both servers should exist
	assert.Contains(t, mcpMap, "custom-server")
	assert.Contains(t, mcpMap, "terraform")

	// Custom server should be unchanged
	customServer, ok := mcpMap["custom-server"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "custom-cmd", customServer["command"])
}

func TestMergeMCPServers_UpdatesExistingServer(t *testing.T) {
	settings := map[string]any{
		"mcpServers": map[string]any{
			"github": map[string]any{
				"command": "old-command",
			},
		},
	}

	mcpServers := []MCPServer{
		{
			Name:    "github",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-github"},
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	githubServer, ok := mcpMap["github"].(map[string]any)
	require.True(t, ok)

	// Should be updated to new command
	assert.Equal(t, "npx", githubServer["command"])
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-github"}, githubServer["args"])
}

func TestMergeMCPServers_PreservesOtherSettings(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"events": []string{"pre-commit"},
		},
		"otherSetting": "value",
	}

	mcpServers := []MCPServer{
		{
			Name:    "go-semantic",
			Command: "go-semantic-mcp",
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	// Other settings should be preserved
	assert.Equal(t, map[string]any{"events": []string{"pre-commit"}}, result["hooks"])
	assert.Equal(t, "value", result["otherSetting"])
}

func TestLoadExistingSettings_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	settings, err := loadExistingSettings(settingsPath)
	require.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Empty(t, settings)
}

func TestLoadExistingSettings_ValidJSON(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	existingData := map[string]any{
		"hooks": map[string]any{},
		"mcpServers": map[string]any{
			"custom": map[string]any{
				"command": "custom-cmd",
			},
		},
	}

	data, err := json.MarshalIndent(existingData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	settings, err := loadExistingSettings(settingsPath)
	require.NoError(t, err)
	assert.NotNil(t, settings["hooks"])
	assert.NotNil(t, settings["mcpServers"])
}

func TestLoadExistingSettings_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	require.NoError(t, os.WriteFile(settingsPath, []byte("invalid json"), 0644))

	settings, err := loadExistingSettings(settingsPath)
	require.NoError(t, err) // Should not error, returns empty map
	assert.NotNil(t, settings)
	assert.Empty(t, settings) // Invalid JSON returns empty map
}

func TestSaveSettings(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.local.json")

	settings := map[string]any{
		"hooks": map[string]any{},
		"mcpServers": map[string]any{
			"github": map[string]any{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-github"},
			},
		},
	}

	err := saveSettings(settingsPath, settings)
	require.NoError(t, err)

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var loaded map[string]any
	require.NoError(t, json.Unmarshal(data, &loaded))

	assert.NotNil(t, loaded["hooks"])
	assert.NotNil(t, loaded["mcpServers"])
}

func TestRiteManifest_MCPServerNames(t *testing.T) {
	manifest := &RiteManifest{
		MCPServers: []MCPServer{
			{Name: "github"},
			{Name: "terraform"},
			{Name: "go-semantic"},
		},
	}

	names := manifest.MCPServerNames()
	assert.Equal(t, []string{"github", "terraform", "go-semantic"}, names)
}

func TestRiteManifest_MCPServerNames_Empty(t *testing.T) {
	manifest := &RiteManifest{}

	names := manifest.MCPServerNames()
	assert.NotNil(t, names)
	assert.Empty(t, names)
}

func TestMergeMCPServers_StdioDefault(t *testing.T) {
	// When no Type is set, server should be treated as stdio (backward compat)
	settings := make(map[string]any)
	mcpServers := []MCPServer{
		{
			Name:    "stdio-server",
			Command: "npx",
			Args:    []string{"-y", "server"},
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	server, ok := mcpMap["stdio-server"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "npx", server["command"])
	assert.Equal(t, []string{"-y", "server"}, server["args"])
	// Type should NOT be emitted for stdio default
	assert.NotContains(t, server, "type")
	assert.NotContains(t, server, "url")
	assert.NotContains(t, server, "headers")
}

func TestMergeMCPServers_SSETransport(t *testing.T) {
	settings := make(map[string]any)
	mcpServers := []MCPServer{
		{
			Name: "sse-server",
			Type: "sse",
			URL:  "https://api.example.com/sse",
			Headers: map[string]string{
				"Authorization": "Bearer ${TOKEN}",
			},
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	server, ok := mcpMap["sse-server"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "sse", server["type"])
	assert.Equal(t, "https://api.example.com/sse", server["url"])
	headers, ok := server["headers"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "Bearer ${TOKEN}", headers["Authorization"])
	// stdio fields should NOT be present
	assert.NotContains(t, server, "command")
	assert.NotContains(t, server, "args")
}

func TestMergeMCPServers_HTTPTransport(t *testing.T) {
	settings := make(map[string]any)
	mcpServers := []MCPServer{
		{
			Name: "http-server",
			Type: "http",
			URL:  "https://api.example.com/mcp",
			Headers: map[string]string{
				"X-API-Key": "${API_KEY}",
				"Accept":    "application/json",
			},
			Env: map[string]string{
				"API_KEY": "${SECRET}",
			},
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	server, ok := mcpMap["http-server"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "http", server["type"])
	assert.Equal(t, "https://api.example.com/mcp", server["url"])
	headers, ok := server["headers"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "${API_KEY}", headers["X-API-Key"])
	assert.Equal(t, "application/json", headers["Accept"])
	// Env should still work with http transport
	envMap, ok := server["env"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "${SECRET}", envMap["API_KEY"])
	// stdio fields should NOT be present
	assert.NotContains(t, server, "command")
	assert.NotContains(t, server, "args")
}

func TestMergeMCPServers_BackwardCompat_ExistingStdio(t *testing.T) {
	// Verify that existing stdio servers without Type field still work
	settings := map[string]any{
		"mcpServers": map[string]any{
			"existing-stdio": map[string]any{
				"command": "old-cmd",
				"args":    []string{"--old"},
			},
		},
	}

	mcpServers := []MCPServer{
		{
			Name: "new-sse",
			Type: "sse",
			URL:  "https://example.com/sse",
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	// Existing stdio server preserved
	existingServer, ok := mcpMap["existing-stdio"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "old-cmd", existingServer["command"])

	// New SSE server added
	newServer, ok := mcpMap["new-sse"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "sse", newServer["type"])
	assert.Equal(t, "https://example.com/sse", newServer["url"])
}

func TestMergeMCPServers_NoEnvOrArgs(t *testing.T) {
	settings := make(map[string]any)
	mcpServers := []MCPServer{
		{
			Name:    "simple-server",
			Command: "simple-cmd",
			// No Args or Env
		},
	}

	result := mergeMCPServers(settings, mcpServers)

	mcpMap, ok := result["mcpServers"].(map[string]any)
	require.True(t, ok)

	simpleServer, ok := mcpMap["simple-server"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "simple-cmd", simpleServer["command"])
	// Args and Env should not be present if they were empty in source
	assert.NotContains(t, simpleServer, "args")
	assert.NotContains(t, simpleServer, "env")
}
