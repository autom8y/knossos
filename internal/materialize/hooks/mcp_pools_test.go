package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMCPPoolsConfig_NotFound(t *testing.T) {
	t.Parallel()
	cfg := LoadMCPPoolsConfigWithPaths("/nonexistent", "/also-nonexistent")
	assert.Nil(t, cfg)
}

func TestLoadMCPPoolsConfig_ValidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	yaml := `schema_version: "1.0"
pools:
  browser-local:
    description: "Local browser"
    server:
      name: browserbase
      command: npx
      args: ["-y", "@autom8y/mcp-stagehand"]
      env:
        STAGEHAND_ENV: "${STAGEHAND_ENV}"
  browser-cloud:
    description: "Cloud browser"
    server:
      name: browserbase
      command: npx
      args: ["-y", "@browserbasehq/mcp-browserbase"]
      env:
        BROWSERBASE_API_KEY: "${BROWSERBASE_API_KEY}"
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "mcp-pools.yaml"), []byte(yaml), 0644))

	cfg := LoadMCPPoolsConfigWithPaths("", dir)
	require.NotNil(t, cfg)
	assert.Equal(t, "1.0", cfg.SchemaVersion)
	assert.Len(t, cfg.Pools, 2)

	localPool, ok := cfg.Pools["browser-local"]
	require.True(t, ok)
	assert.Equal(t, "browserbase", localPool.Server.Name)
	assert.Equal(t, "npx", localPool.Server.Command)
	assert.Equal(t, []string{"-y", "@autom8y/mcp-stagehand"}, localPool.Server.Args)
	assert.Equal(t, "${STAGEHAND_ENV}", localPool.Server.Env["STAGEHAND_ENV"])
}

func TestLoadMCPPoolsConfig_InvalidSchema(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	yaml := `schema_version: "2.0"
pools:
  test:
    server:
      name: test
      command: echo
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "mcp-pools.yaml"), []byte(yaml), 0644))

	cfg := LoadMCPPoolsConfigWithPaths("", dir)
	assert.Nil(t, cfg, "schema version 2.0 should not be accepted")
}

func TestLoadMCPPoolsConfig_KnossosHomePrecedence(t *testing.T) {
	t.Parallel()
	knossosHome := t.TempDir()
	projectRoot := t.TempDir()

	for _, dir := range []string{knossosHome, projectRoot} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "config"), 0755))
	}

	// Write different configs to each location
	homeYAML := `schema_version: "1.0"
pools:
  home-pool:
    server:
      name: home-server
      command: echo
`
	projectYAML := `schema_version: "1.0"
pools:
  project-pool:
    server:
      name: project-server
      command: echo
`
	require.NoError(t, os.WriteFile(filepath.Join(knossosHome, "config/mcp-pools.yaml"), []byte(homeYAML), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, "config/mcp-pools.yaml"), []byte(projectYAML), 0644))

	cfg := LoadMCPPoolsConfigWithPaths(knossosHome, projectRoot)
	require.NotNil(t, cfg)
	_, hasHomePool := cfg.Pools["home-pool"]
	assert.True(t, hasHomePool, "KNOSSOS_HOME should take precedence")
}

func TestResolvePoolServers_BasicResolution(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Args:    []string{"-y", "@autom8y/mcp-stagehand"},
					Env:     map[string]string{"STAGEHAND_ENV": "${STAGEHAND_ENV}"},
				},
			},
		},
	}

	refs := []MCPPoolRef{{Pool: "browser-local"}}
	servers, err := ResolvePoolServers(pools, refs)
	require.NoError(t, err)
	require.Len(t, servers, 1)

	assert.Equal(t, "browserbase", servers[0].Name)
	assert.Equal(t, "npx", servers[0].Command)
	assert.Equal(t, []string{"-y", "@autom8y/mcp-stagehand"}, servers[0].Args)
	assert.Equal(t, "${STAGEHAND_ENV}", servers[0].Env["STAGEHAND_ENV"])
}

func TestResolvePoolServers_ArgsAppend(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Args:    []string{"-y", "@autom8y/mcp-stagehand"},
				},
			},
		},
	}

	refs := []MCPPoolRef{{
		Pool:       "browser-local",
		ArgsAppend: []string{"--headless", "--timeout", "30"},
	}}
	servers, err := ResolvePoolServers(pools, refs)
	require.NoError(t, err)
	require.Len(t, servers, 1)

	assert.Equal(t, []string{"-y", "@autom8y/mcp-stagehand", "--headless", "--timeout", "30"}, servers[0].Args)
}

func TestResolvePoolServers_EnvMerge(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Env: map[string]string{
						"STAGEHAND_ENV": "${STAGEHAND_ENV}",
						"KEEP_ME":       "original",
					},
				},
			},
		},
	}

	refs := []MCPPoolRef{{
		Pool: "browser-local",
		EnvMerge: map[string]string{
			"STAGEHAND_ENV":          "LOCAL",  // Override existing
			"STAGEHAND_MODEL_API_KEY": "${KEY}", // Add new
		},
	}}
	servers, err := ResolvePoolServers(pools, refs)
	require.NoError(t, err)
	require.Len(t, servers, 1)

	assert.Equal(t, "LOCAL", servers[0].Env["STAGEHAND_ENV"], "rite override should win")
	assert.Equal(t, "${KEY}", servers[0].Env["STAGEHAND_MODEL_API_KEY"], "new env should be added")
	assert.Equal(t, "original", servers[0].Env["KEEP_ME"], "existing non-overridden env should be preserved")
}

func TestResolvePoolServers_EnvMerge_RewritesArgs(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Args: []string{
						"-y", "@autom8y/mcp-stagehand",
						"--modelName", "${STAGEHAND_MODEL_NAME}",
						"--modelApiKey", "${STAGEHAND_MODEL_API_KEY}",
					},
					Env: map[string]string{
						"STAGEHAND_ENV":           "${STAGEHAND_ENV}",
						"STAGEHAND_MODEL_API_KEY": "${STAGEHAND_MODEL_API_KEY}",
					},
				},
			},
		},
	}

	// Bridge: map agnostic STAGEHAND_MODEL_API_KEY to provider-specific ANTHROPIC_API_KEY
	refs := []MCPPoolRef{{
		Pool: "browser-local",
		EnvMerge: map[string]string{
			"STAGEHAND_MODEL_API_KEY": "${ANTHROPIC_API_KEY}",
		},
	}}
	servers, err := ResolvePoolServers(pools, refs)
	require.NoError(t, err)
	require.Len(t, servers, 1)

	// Env should be bridged
	assert.Equal(t, "${ANTHROPIC_API_KEY}", servers[0].Env["STAGEHAND_MODEL_API_KEY"],
		"env_merge should bridge to provider-specific var")

	// Args should also be rewritten
	assert.Equal(t, []string{
		"-y", "@autom8y/mcp-stagehand",
		"--modelName", "${STAGEHAND_MODEL_NAME}",
		"--modelApiKey", "${ANTHROPIC_API_KEY}",
	}, servers[0].Args, "args referencing overridden env var should be rewritten")

	// Non-overridden env and args should be unchanged
	assert.Equal(t, "${STAGEHAND_ENV}", servers[0].Env["STAGEHAND_ENV"])
}

func TestResolvePoolServers_UnknownPool(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{},
	}

	refs := []MCPPoolRef{{Pool: "does-not-exist"}}
	servers, err := ResolvePoolServers(pools, refs)
	assert.Error(t, err)
	assert.Nil(t, servers)
	assert.Contains(t, err.Error(), "unknown MCP pool")
}

func TestResolvePoolServers_EmptyRefs(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{Name: "browserbase"},
			},
		},
	}

	servers, err := ResolvePoolServers(pools, nil)
	require.NoError(t, err)
	assert.Nil(t, servers)

	servers, err = ResolvePoolServers(pools, []MCPPoolRef{})
	require.NoError(t, err)
	assert.Nil(t, servers)
}

func TestResolvePoolServers_NilPools(t *testing.T) {
	t.Parallel()
	servers, err := ResolvePoolServers(nil, []MCPPoolRef{{Pool: "anything"}})
	require.NoError(t, err)
	assert.Nil(t, servers)
}

func TestResolvePoolServers_DoesNotMutatePoolDefinition(t *testing.T) {
	t.Parallel()
	pools := &MCPPoolsConfig{
		Pools: map[string]MCPPool{
			"browser-local": {
				Server: MCPServerConfig{
					Name:    "browserbase",
					Command: "npx",
					Args:    []string{"-y", "@autom8y/mcp-stagehand"},
					Env:     map[string]string{"STAGEHAND_ENV": "${STAGEHAND_ENV}"},
				},
			},
		},
	}

	refs := []MCPPoolRef{{
		Pool:       "browser-local",
		ArgsAppend: []string{"--extra"},
		EnvMerge:   map[string]string{"NEW_VAR": "new_value"},
	}}

	_, err := ResolvePoolServers(pools, refs)
	require.NoError(t, err)

	// Original pool must be unchanged
	pool := pools.Pools["browser-local"]
	assert.Equal(t, []string{"-y", "@autom8y/mcp-stagehand"}, pool.Server.Args, "pool args must not be mutated")
	assert.NotContains(t, pool.Server.Env, "NEW_VAR", "pool env must not be mutated")
}

func TestValidateMCPEnvVars_WarnOnUnset(t *testing.T) {
	// No t.Parallel — uses t.Setenv which requires sequential execution
	os.Unsetenv("UNLIKELY_VAR_FOR_TEST_ABC123")

	servers := []MCPServerConfig{{
		Name: "test-server",
		Env:  map[string]string{"UNLIKELY_VAR_FOR_TEST_ABC123": "${UNLIKELY_VAR_FOR_TEST_ABC123}"},
	}}

	// ValidateMCPEnvVars logs warnings but doesn't error — just verify it doesn't panic
	ValidateMCPEnvVars(servers)
}

func TestValidateMCPEnvVars_NoWarnOnSet(t *testing.T) {
	// No t.Parallel — uses t.Setenv which requires sequential execution
	t.Setenv("MCP_POOL_TEST_SET_VAR", "some-value")

	servers := []MCPServerConfig{{
		Name: "test-server",
		Env:  map[string]string{"MCP_POOL_TEST_SET_VAR": "${MCP_POOL_TEST_SET_VAR}"},
	}}

	// Should not warn — just verify no panic
	ValidateMCPEnvVars(servers)
}

func TestValidateMCPEnvVars_ScansArgsAndEnv(t *testing.T) {
	t.Parallel()
	servers := []MCPServerConfig{{
		Name: "test-server",
		Args: []string{"--key", "${ARG_SCANNED_VAR_XYZ}"},
		Env:  map[string]string{"ENV_SCANNED_VAR_XYZ": "${ENV_SCANNED_VAR_XYZ}"},
	}}

	// Just verify no panic — both args and env are scanned
	ValidateMCPEnvVars(servers)
}
