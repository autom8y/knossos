package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
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

// newMCPTestMaterializer creates a Materializer with a paths.Resolver rooted at projectRoot.
// Ensures .knossos/ directory exists for ownership tracking.
func newMCPTestMaterializer(t *testing.T, projectRoot string) *Materializer {
	t.Helper()
	knossosDir := filepath.Join(projectRoot, ".knossos")
	require.NoError(t, os.MkdirAll(knossosDir, 0755))
	return &Materializer{resolver: paths.NewResolver(projectRoot)}
}

// TestMaterializeMcpJson_WritesToProjectRoot tests that MCP servers from
// the rite manifest are written to .mcp.json at project root (SCAR-028).
func TestMaterializeMcpJson_WritesToProjectRoot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	m := newMCPTestMaterializer(t, projectRoot)
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

	err := m.materializeMcpJsonFromResolved(projectRoot, manifest, manifest.MCPServers, provenance.NullCollector{}, "")
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
	m := newMCPTestMaterializer(t, projectRoot)
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

	err = m.materializeMcpJsonFromResolved(projectRoot, manifest, manifest.MCPServers, provenance.NullCollector{}, "")
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
	m := newMCPTestMaterializer(t, projectRoot)

	err := m.materializeMcpJsonFromResolved(projectRoot, nil, nil, provenance.NullCollector{}, "")
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
	m := newMCPTestMaterializer(t, projectRoot)
	manifest := &RiteManifest{Name: "test-rite"}

	err := m.materializeMcpJsonFromResolved(projectRoot, manifest, manifest.MCPServers, provenance.NullCollector{}, "")
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
	m := newMCPTestMaterializer(t, projectRoot)
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

	err = m.materializeMcpJsonFromResolved(projectRoot, manifest, manifest.MCPServers, provenance.NullCollector{}, "")
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

// TestMaterializeMcpJsonWithPools_ResolvesAndWrites tests end-to-end pool resolution to .mcp.json.
func TestMaterializeMcpJsonWithPools_ResolvesAndWrites(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	m := newMCPTestMaterializer(t, projectRoot)

	poolsConfig := &MCPPoolsConfig{
		SchemaVersion: "1.0",
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Args:    []string{"-y", "stagehand-mcp-local"},
					Env:     map[string]string{"STAGEHAND_ENV": "${STAGEHAND_ENV}"},
				},
			},
		},
	}

	manifest := &RiteManifest{
		Name:     "ui",
		MCPPools: []MCPPoolRef{{Pool: "browser-local"}},
	}

	err := func() error {
		resolved, err := resolveAllMCPServers(manifest, poolsConfig)
		if err != nil {
			return err
		}
		return m.materializeMcpJsonFromResolved(projectRoot, manifest, resolved, provenance.NullCollector{}, "")
	}()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(projectRoot, ".mcp.json"))
	require.NoError(t, err)

	var mcpFile map[string]any
	require.NoError(t, json.Unmarshal(data, &mcpFile))

	mcpServers, ok := mcpFile["mcpServers"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, mcpServers, "browserbase", "pool-resolved server must be in .mcp.json")

	browserbase, ok := mcpServers["browserbase"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "npx", browserbase["command"])
}

// TestMaterializeMcpJsonWithPools_DirectServerOverridesPool tests name collision:
// direct mcp_servers win over pool-resolved servers.
func TestMaterializeMcpJsonWithPools_DirectServerOverridesPool(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	m := newMCPTestMaterializer(t, projectRoot)

	poolsConfig := &MCPPoolsConfig{
		SchemaVersion: "1.0",
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Args:    []string{"-y", "stagehand-mcp-local"},
				},
			},
		},
	}

	manifest := &RiteManifest{
		Name:     "ui",
		MCPPools: []MCPPoolRef{{Pool: "browser-local"}},
		MCPServers: []MCPServer{
			{Name: "browserbase", Command: "custom-command", Args: []string{"--custom"}},
		},
	}

	err := func() error {
		resolved, err := resolveAllMCPServers(manifest, poolsConfig)
		if err != nil {
			return err
		}
		return m.materializeMcpJsonFromResolved(projectRoot, manifest, resolved, provenance.NullCollector{}, "")
	}()
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(projectRoot, ".mcp.json"))
	var mcpFile map[string]any
	require.NoError(t, json.Unmarshal(data, &mcpFile))

	mcpServers := mcpFile["mcpServers"].(map[string]any)
	browserbase := mcpServers["browserbase"].(map[string]any)
	assert.Equal(t, "custom-command", browserbase["command"], "direct server must override pool server on name collision")
}

// TestMaterializeMcpJsonWithPools_NilPoolsConfig tests graceful degradation
// when no mcp-pools.yaml exists.
func TestMaterializeMcpJsonWithPools_NilPoolsConfig(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	m := newMCPTestMaterializer(t, projectRoot)

	manifest := &RiteManifest{
		Name: "ecosystem",
		MCPServers: []MCPServer{
			{Name: "go-semantic", Command: "go-semantic-mcp"},
		},
	}

	err := m.materializeMcpJsonFromResolved(projectRoot, manifest, manifest.MCPServers, provenance.NullCollector{}, "")
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(projectRoot, ".mcp.json"))
	var mcpFile map[string]any
	require.NoError(t, json.Unmarshal(data, &mcpFile))

	mcpServers := mcpFile["mcpServers"].(map[string]any)
	assert.Contains(t, mcpServers, "go-semantic", "direct servers must work without pools config")
}

// TestMaterializeMcpJsonWithPools_WritesOwnership tests that ownership file
// is written alongside .mcp.json.
func TestMaterializeMcpJsonWithPools_WritesOwnership(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	m := newMCPTestMaterializer(t, projectRoot)

	manifest := &RiteManifest{
		Name: "ui",
		MCPServers: []MCPServer{
			{Name: "browserbase", Command: "npx"},
		},
	}

	err := m.materializeMcpJsonFromResolved(projectRoot, manifest, manifest.MCPServers, provenance.NullCollector{}, "")
	require.NoError(t, err)

	ownership := loadMCPOwnership(filepath.Join(projectRoot, ".knossos"))
	require.NotNil(t, ownership)
	assert.Equal(t, "ui", ownership.Rite)
	assert.Equal(t, []string{"browserbase"}, ownership.Servers)
}

// TestSCAR_MCPPoolResolution_NeverWritesToSettingsLocalJson is a SCAR regression
// test verifying that pool-resolved servers still go to .mcp.json only.
func TestSCAR_MCPPoolResolution_NeverWritesToSettingsLocalJson(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	poolsConfig := &MCPPoolsConfig{
		SchemaVersion: "1.0",
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{Name: "browserbase", Command: "npx"},
			},
		},
	}

	manifest := &RiteManifest{
		Name:     "ui",
		MCPPools: []MCPPoolRef{{Pool: "browser-local"}},
	}

	// Write settings (hooks only)
	err := (&Materializer{}).materializeSettingsWithManifest(tempDir, manifest, provenance.NullCollector{}, "claude")
	require.NoError(t, err)

	settingsPath := filepath.Join(tempDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))
	assert.Nil(t, settings["mcpServers"],
		"SCAR-028: pool-resolved MCP servers must NOT be in settings.local.json")

	// Write MCP servers to .mcp.json
	m := newMCPTestMaterializer(t, tempDir)
	resolved, resolveErr := resolveAllMCPServers(manifest, poolsConfig)
	require.NoError(t, resolveErr)
	err = m.materializeMcpJsonFromResolved(tempDir, manifest, resolved, provenance.NullCollector{}, "")
	require.NoError(t, err)

	// Re-read settings — must still have no mcpServers
	data, err = os.ReadFile(settingsPath)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &settings))
	assert.Nil(t, settings["mcpServers"])
}
