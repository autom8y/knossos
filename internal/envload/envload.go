package envload

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/autom8y/knossos/internal/config"
)

// ServeEnvFile is the filename for ari serve configuration.
const ServeEnvFile = "serve.env"

// ServeConfig holds resolved, typed configuration for ari serve.
// All fields are populated after Load() applies the resolution hierarchy.
type ServeConfig struct {
	// Secrets (required -- caller validates after Load returns)
	SlackSigningSecret string
	SlackBotToken      string
	AnthropicAPIKey    string

	// Server config (defaults applied if not set)
	Port          int
	LogLevel      string
	MaxConcurrent int
	DrainTimeout  time.Duration

	// Observability (optional, empty = disabled)
	OTELEndpoint string
}

// Overrides holds values from CLI flags. Zero values mean "not set by flag."
// The caller populates this from cobra flag bindings.
type Overrides struct {
	SlackSigningSecret string
	SlackBotToken      string
	Port               int
	MaxConcurrent      int
	DrainTimeout       time.Duration
	EnvFile            string
}

// DefaultServeConfig returns ServeConfig with hardcoded defaults for non-secret fields.
func DefaultServeConfig() ServeConfig {
	return ServeConfig{
		Port:          8080,
		LogLevel:      "INFO",
		MaxConcurrent: 10,
		DrainTimeout:  30 * time.Second,
	}
}

// Load resolves ServeConfig using the hierarchy: overrides > process env > org env file > defaults.
//
// If orgCtx is nil, the org env file layer is skipped (pure env-var mode for backward compat).
// If the org env file does not exist, the file layer is silently skipped.
// Returns an error only if the env file exists but is malformed.
//
// Load does NOT validate that required secrets are present -- the caller checks that
// after Load returns, because the error messages are command-specific.
func Load(orgCtx config.OrgContext, overrides Overrides) (ServeConfig, error) {
	cfg := DefaultServeConfig()

	// Load org env file (tier 3).
	var fileVars map[string]string
	envFilePath := resolveEnvFilePath(orgCtx, overrides.EnvFile)
	if envFilePath != "" {
		var err error
		fileVars, err = LoadFile(envFilePath)
		if err != nil {
			return ServeConfig{}, err
		}
	}

	// Apply the four-tier hierarchy for each config field.
	// Tier order: overrides (1) > os.Getenv (2) > file (3) > default (4).

	// SlackSigningSecret
	cfg.SlackSigningSecret = resolveString(
		overrides.SlackSigningSecret,
		os.Getenv("SLACK_SIGNING_SECRET"),
		fileVars["SLACK_SIGNING_SECRET"],
		"",
	)

	// SlackBotToken
	cfg.SlackBotToken = resolveString(
		overrides.SlackBotToken,
		os.Getenv("SLACK_BOT_TOKEN"),
		fileVars["SLACK_BOT_TOKEN"],
		"",
	)

	// AnthropicAPIKey (no CLI flag override -- secrets should not be on the command line)
	cfg.AnthropicAPIKey = resolveString(
		"",
		os.Getenv("ANTHROPIC_API_KEY"),
		fileVars["ANTHROPIC_API_KEY"],
		"",
	)

	// Port
	cfg.Port = resolveInt(
		overrides.Port,
		envInt("PORT"),
		fileInt(fileVars, "PORT"),
		cfg.Port,
	)

	// LogLevel
	cfg.LogLevel = resolveString(
		"",
		os.Getenv("LOG_LEVEL"),
		fileVars["LOG_LEVEL"],
		cfg.LogLevel,
	)

	// MaxConcurrent
	cfg.MaxConcurrent = resolveInt(
		overrides.MaxConcurrent,
		envInt("MAX_CONCURRENT"),
		fileInt(fileVars, "MAX_CONCURRENT"),
		cfg.MaxConcurrent,
	)

	// DrainTimeout
	cfg.DrainTimeout = resolveDuration(
		overrides.DrainTimeout,
		envDuration("DRAIN_TIMEOUT"),
		fileDuration(fileVars, "DRAIN_TIMEOUT"),
		cfg.DrainTimeout,
	)

	// OTELEndpoint
	cfg.OTELEndpoint = resolveString(
		"",
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		fileVars["OTEL_EXPORTER_OTLP_ENDPOINT"],
		"",
	)

	return cfg, nil
}

// LoadFile parses a dotenv file into a map[string]string.
// Returns an empty map (not an error) if the file does not exist.
// Returns an error if the file exists but contains syntax errors.
func LoadFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("opening env file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	CheckPermissions(path)

	vars, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing env file %s: %w", path, err)
	}

	return vars, nil
}

// resolveEnvFilePath determines the env file path from explicit override or org context.
func resolveEnvFilePath(orgCtx config.OrgContext, envFileOverride string) string {
	if envFileOverride != "" {
		return envFileOverride
	}
	if orgCtx == nil {
		return ""
	}
	return filepath.Join(orgCtx.DataDir(), ServeEnvFile)
}

// resolveString returns the first non-empty string in priority order.
func resolveString(flag, env, file, def string) string {
	if flag != "" {
		return flag
	}
	if env != "" {
		return env
	}
	if file != "" {
		return file
	}
	return def
}

// resolveInt returns the first non-zero int in priority order.
func resolveInt(flag, env, file, def int) int {
	if flag != 0 {
		return flag
	}
	if env != 0 {
		return env
	}
	if file != 0 {
		return file
	}
	return def
}

// resolveDuration returns the first non-zero duration in priority order.
func resolveDuration(flag, env, file, def time.Duration) time.Duration {
	if flag != 0 {
		return flag
	}
	if env != 0 {
		return env
	}
	if file != 0 {
		return file
	}
	return def
}

// envInt reads an env var as int, returns 0 if missing or unparseable.
func envInt(key string) int {
	s := os.Getenv(key)
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		slog.Warn("invalid integer in environment variable", "key", key, "value", s, "error", err)
		return 0
	}
	return v
}

// envDuration reads an env var as duration, returns 0 if missing or unparseable.
func envDuration(key string) time.Duration {
	s := os.Getenv(key)
	if s == "" {
		return 0
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		slog.Warn("invalid duration in environment variable", "key", key, "value", s, "error", err)
		return 0
	}
	return v
}

// fileInt reads a key from file vars as int, returns 0 if missing or unparseable.
func fileInt(vars map[string]string, key string) int {
	s, ok := vars[key]
	if !ok || s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		slog.Warn("invalid integer in env file", "key", key, "value", s, "error", err)
		return 0
	}
	return v
}

// fileDuration reads a key from file vars as duration, returns 0 if missing or unparseable.
func fileDuration(vars map[string]string, key string) time.Duration {
	s, ok := vars[key]
	if !ok || s == "" {
		return 0
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		slog.Warn("invalid duration in env file", "key", key, "value", s, "error", err)
		return 0
	}
	return v
}
