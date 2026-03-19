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
//  4. Apply channel-specific env overrides (e.g. GEMINI_CUA_MODEL -> CUA_MODEL)
//  5. Apply env_merge: merge into server.Env (rite values win on conflict)
//
// Returns error on unknown pool name (misconfiguration).
func ResolvePoolServers(pools *MCPPoolsConfig, refs []MCPPoolRef, channel string) ([]MCPServerConfig, error) {
	if pools == nil || len(refs) == 0 {
		return nil, nil
	}

	// Determine channel prefix for env overrides
	// gemini -> GEMINI_
	// claude -> ANTHROPIC_ (legacy/default)
	channelPrefix := "ANTHROPIC"
	if channel == "gemini" {
		channelPrefix = "GEMINI"
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

		// Build merge map: start with implicit channel overrides, then apply explicit env_merge.
		// Channel overrides allow parameterizing CUA_MODEL/API keys by channel.
		// Logic: if server has env var VAR, check for {CHANNEL}_VAR in os.Env.
		// If found, override VAR -> "${{CHANNEL}_VAR}".
		mergeMap := make(map[string]string)

		// 1. Implicit channel overrides
		if server.Env != nil {
			for key := range server.Env {
				overrideVar := fmt.Sprintf("%s_%s", channelPrefix, key)
				if _, exists := os.LookupEnv(overrideVar); exists {
					mergeMap[key] = fmt.Sprintf("${%s}", overrideVar)
				}
			}
		}

		// 2. Explicit env_merge from rite (wins over implicit)
		if len(ref.EnvMerge) > 0 {
			maps.Copy(mergeMap, ref.EnvMerge)
		}

		// Apply final merge map
		if len(mergeMap) > 0 {
			if server.Env == nil {
				server.Env = make(map[string]string)
			}
			for key, newVal := range mergeMap {
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
