package hooks

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/config"
)

// MCPPoolsConfig represents a parsed config/mcp-pools.yaml file.
type MCPPoolsConfig struct {
	SchemaVersion string             `yaml:"schema_version"`
	Pools         map[string]MCPPool `yaml:"pools"`
}

// MCPPool defines a reusable MCP server pool.
type MCPPool struct {
	Description string          `yaml:"description,omitempty"`
	Server      MCPServerConfig `yaml:"server"`
}

// MCPPoolRef references a pool from config/mcp-pools.yaml with optional overrides.
// Defined in the hooks sub-package to avoid circular imports; re-exported by the
// parent materialize package.
type MCPPoolRef struct {
	Pool       string            `yaml:"pool"`                  // Pool name from config/mcp-pools.yaml
	ArgsAppend []string          `yaml:"args_append,omitempty"` // Appended to pool server args
	EnvMerge   map[string]string `yaml:"env_merge,omitempty"`   // Merged into pool server env (rite wins)
}

// LoadMCPPoolsConfig finds and parses mcp-pools.yaml from the filesystem.
// Resolution order:
//  1. config/mcp-pools.yaml in $KNOSSOS_HOME
//  2. config/mcp-pools.yaml in projectRoot (for self-hosting and satellites)
//
// Returns nil if no mcp-pools.yaml is found (graceful).
func LoadMCPPoolsConfig(projectRoot string) *MCPPoolsConfig {
	return LoadMCPPoolsConfigWithPaths(config.KnossosHome(), projectRoot)
}

// LoadMCPPoolsConfigWithPaths finds and parses mcp-pools.yaml using explicit paths.
// This is the DI-capable variant that avoids reading config globals.
// Resolution order:
//  1. config/mcp-pools.yaml in knossosHome
//  2. config/mcp-pools.yaml in projectRoot (for self-hosting and satellites)
//
// Returns nil if no mcp-pools.yaml is found (graceful).
func LoadMCPPoolsConfigWithPaths(knossosHome, projectRoot string) *MCPPoolsConfig {
	var candidates []string
	if knossosHome != "" {
		candidates = append(candidates, knossosHome+"/config/mcp-pools.yaml")
	}
	if projectRoot != "" {
		candidates = append(candidates, projectRoot+"/config/mcp-pools.yaml")
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cfg MCPPoolsConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue
		}

		if cfg.SchemaVersion != "1.0" {
			continue
		}

		return &cfg
	}

	return nil
}

// ResolvePoolServers resolves MCPPoolRef entries into MCPServerConfig entries.
// For each pool ref:
//  1. Look up pool by name in MCPPoolsConfig
//  2. Clone the pool's canonical server definition
//  3. Apply args_append: append to server.Args
//  4. Apply env_merge: merge into server.Env (rite values win on conflict)
//
// Returns error on unknown pool name (misconfiguration).
func ResolvePoolServers(pools *MCPPoolsConfig, refs []MCPPoolRef) ([]MCPServerConfig, error) {
	if pools == nil || len(refs) == 0 {
		return nil, nil
	}

	servers := make([]MCPServerConfig, 0, len(refs))

	for _, ref := range refs {
		pool, ok := pools.Pools[ref.Pool]
		if !ok {
			return nil, fmt.Errorf("unknown MCP pool %q", ref.Pool)
		}

		// Clone the canonical server
		server := MCPServerConfig{
			Name:    pool.Server.Name,
			Command: pool.Server.Command,
			Type:    pool.Server.Type,
			URL:     pool.Server.URL,
		}

		// Clone args (don't mutate pool definition)
		if len(pool.Server.Args) > 0 {
			server.Args = make([]string, len(pool.Server.Args))
			copy(server.Args, pool.Server.Args)
		}

		// Clone env
		if len(pool.Server.Env) > 0 {
			server.Env = make(map[string]string, len(pool.Server.Env))
			maps.Copy(server.Env, pool.Server.Env)
		}

		// Clone headers
		if len(pool.Server.Headers) > 0 {
			server.Headers = make(map[string]string, len(pool.Server.Headers))
			maps.Copy(server.Headers, pool.Server.Headers)
		}

		// Apply args_append
		if len(ref.ArgsAppend) > 0 {
			server.Args = append(server.Args, ref.ArgsAppend...)
		}

		// Apply env_merge (rite values win on conflict).
		// Also rewrites ${VAR} references in args: when env_merge overrides a key
		// whose pool value was "${KEY}", any "${KEY}" in args is rewritten to the
		// new value. This enables provider bridging: the pool uses agnostic var
		// names (STAGEHAND_MODEL_API_KEY) and the rite bridges to provider-specific
		// vars (${ANTHROPIC_API_KEY}) via env_merge — args are rewritten to match
		// so CC resolves them from the parent process env.
		if len(ref.EnvMerge) > 0 {
			if server.Env == nil {
				server.Env = make(map[string]string)
			}
			for key, newVal := range ref.EnvMerge {
				oldVal, existed := server.Env[key]
				server.Env[key] = newVal

				// Rewrite args: if the pool's env value was a ${VAR} reference,
				// replace all occurrences of that reference in args with the new value.
				if existed && oldVal != newVal {
					for i, arg := range server.Args {
						if arg == oldVal {
							server.Args[i] = newVal
						}
					}
				}
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// envVarPattern matches ${VAR_NAME} references in strings.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// ValidateMCPEnvVars scans MCP server configs for ${VAR} patterns and emits
// slog.Warn for each unset variable. Non-blocking: never returns an error.
// CC/Gemini resolve ${VAR} at runtime via their own env; this validation
// catches missing direnv setup during development.
func ValidateMCPEnvVars(servers []MCPServerConfig) {
	seen := make(map[string]bool)

	for _, server := range servers {
		// Scan env values
		for _, v := range server.Env {
			extractAndWarn(v, server.Name, &seen)
		}
		// Scan args (some MCPs take ${VAR} in CLI flags)
		for _, arg := range server.Args {
			extractAndWarn(arg, server.Name, &seen)
		}
	}
}

// extractAndWarn extracts ${VAR} references from a string and warns on unset vars.
func extractAndWarn(s, serverName string, seen *map[string]bool) {
	matches := envVarPattern.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		varName := match[1]
		if (*seen)[varName] {
			continue
		}
		(*seen)[varName] = true
		if _, ok := os.LookupEnv(varName); !ok {
			slog.Warn("MCP env var not set", "var", varName, "server", serverName)
		}
	}
}
