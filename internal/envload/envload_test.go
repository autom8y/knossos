package envload

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/config"
)

// mockOrgContext implements config.OrgContext for testing.
type mockOrgContext struct {
	name    string
	dataDir string
}

func (m *mockOrgContext) Name() string                { return m.name }
func (m *mockOrgContext) DataDir() string              { return m.dataDir }
func (m *mockOrgContext) RegistryDir() string          { return filepath.Join(m.dataDir, "registry") }
func (m *mockOrgContext) Repos() []config.RepoConfig   { return nil }

func TestDefaultServeConfig(t *testing.T) {
	cfg := DefaultServeConfig()

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("LogLevel = %q, want INFO", cfg.LogLevel)
	}
	if cfg.MaxConcurrent != 10 {
		t.Errorf("MaxConcurrent = %d, want 10", cfg.MaxConcurrent)
	}
	if cfg.DrainTimeout != 30*time.Second {
		t.Errorf("DrainTimeout = %v, want 30s", cfg.DrainTimeout)
	}
	if cfg.SlackSigningSecret != "" {
		t.Error("SlackSigningSecret should be empty")
	}
	if cfg.SlackBotToken != "" {
		t.Error("SlackBotToken should be empty")
	}
	if cfg.AnthropicAPIKey != "" {
		t.Error("AnthropicAPIKey should be empty")
	}
}

func TestLoad_DefaultsOnly(t *testing.T) {
	// Clear any env vars that would interfere.
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	cfg, err := Load(nil, Overrides{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("LogLevel = %q, want INFO", cfg.LogLevel)
	}
	if cfg.MaxConcurrent != 10 {
		t.Errorf("MaxConcurrent = %d, want 10", cfg.MaxConcurrent)
	}
}

func TestLoad_NilOrgContext(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "env-secret")
	t.Setenv("SLACK_BOT_TOKEN", "env-token")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	cfg, err := Load(nil, Overrides{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SlackSigningSecret != "env-secret" {
		t.Errorf("SlackSigningSecret = %q, want env-secret", cfg.SlackSigningSecret)
	}
	if cfg.SlackBotToken != "env-token" {
		t.Errorf("SlackBotToken = %q, want env-token", cfg.SlackBotToken)
	}
}

func TestLoad_FileLayer(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ServeEnvFile)
	content := `SLACK_SIGNING_SECRET=file-secret
SLACK_BOT_TOKEN=file-token
ANTHROPIC_API_KEY=file-key
PORT=9090
LOG_LEVEL=DEBUG
MAX_CONCURRENT=20
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
`
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	orgCtx := &mockOrgContext{name: "test-org", dataDir: tmpDir}

	cfg, err := Load(orgCtx, Overrides{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SlackSigningSecret != "file-secret" {
		t.Errorf("SlackSigningSecret = %q, want file-secret", cfg.SlackSigningSecret)
	}
	if cfg.SlackBotToken != "file-token" {
		t.Errorf("SlackBotToken = %q, want file-token", cfg.SlackBotToken)
	}
	if cfg.AnthropicAPIKey != "file-key" {
		t.Errorf("AnthropicAPIKey = %q, want file-key", cfg.AnthropicAPIKey)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("LogLevel = %q, want DEBUG", cfg.LogLevel)
	}
	if cfg.MaxConcurrent != 20 {
		t.Errorf("MaxConcurrent = %d, want 20", cfg.MaxConcurrent)
	}
	if cfg.OTELEndpoint != "http://localhost:4318" {
		t.Errorf("OTELEndpoint = %q, want http://localhost:4318", cfg.OTELEndpoint)
	}
}

func TestLoad_EnvBeatsFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ServeEnvFile)
	content := `SLACK_SIGNING_SECRET=file-secret
PORT=9090
`
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Set env vars (tier 2) which should beat file (tier 3).
	t.Setenv("SLACK_SIGNING_SECRET", "env-secret")
	t.Setenv("PORT", "3000")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	orgCtx := &mockOrgContext{name: "test-org", dataDir: tmpDir}

	cfg, err := Load(orgCtx, Overrides{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SlackSigningSecret != "env-secret" {
		t.Errorf("SlackSigningSecret = %q, want env-secret (env beats file)", cfg.SlackSigningSecret)
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want 3000 (env beats file)", cfg.Port)
	}
}

func TestLoad_OverrideBeatsAll(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ServeEnvFile)
	content := `SLACK_SIGNING_SECRET=file-secret
PORT=9090
MAX_CONCURRENT=20
`
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SLACK_SIGNING_SECRET", "env-secret")
	t.Setenv("PORT", "3000")
	t.Setenv("MAX_CONCURRENT", "50")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	orgCtx := &mockOrgContext{name: "test-org", dataDir: tmpDir}

	cfg, err := Load(orgCtx, Overrides{
		SlackSigningSecret: "flag-secret",
		Port:               4000,
		MaxConcurrent:      5,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SlackSigningSecret != "flag-secret" {
		t.Errorf("SlackSigningSecret = %q, want flag-secret (override beats all)", cfg.SlackSigningSecret)
	}
	if cfg.Port != 4000 {
		t.Errorf("Port = %d, want 4000 (override beats all)", cfg.Port)
	}
	if cfg.MaxConcurrent != 5 {
		t.Errorf("MaxConcurrent = %d, want 5 (override beats all)", cfg.MaxConcurrent)
	}
}

func TestLoad_MissingFileNotError(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	tmpDir := t.TempDir()
	// Do NOT create serve.env -- it should be silently skipped.
	orgCtx := &mockOrgContext{name: "test-org", dataDir: tmpDir}

	cfg, err := Load(orgCtx, Overrides{})
	if err != nil {
		t.Fatalf("Load() should not error for missing file, got: %v", err)
	}

	// Should fall through to defaults.
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080 (default)", cfg.Port)
	}
}

func TestLoad_MalformedFileIsError(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ServeEnvFile)
	content := "THIS IS NOT VALID"
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	orgCtx := &mockOrgContext{name: "test-org", dataDir: tmpDir}

	_, err := Load(orgCtx, Overrides{})
	if err == nil {
		t.Fatal("Load() should error for malformed file")
	}
}

func TestLoad_EnvFileOverride(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	tmpDir := t.TempDir()
	customFile := filepath.Join(tmpDir, "custom.env")
	content := `PORT=7777
`
	if err := os.WriteFile(customFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// --env-file should be used even with nil orgCtx.
	cfg, err := Load(nil, Overrides{EnvFile: customFile})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 7777 {
		t.Errorf("Port = %d, want 7777 (from custom env file)", cfg.Port)
	}
}

func TestLoadFile_NonExistent(t *testing.T) {
	vars, err := LoadFile("/nonexistent/path/serve.env")
	if err != nil {
		t.Fatalf("LoadFile() should not error for missing file, got: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("LoadFile() should return empty map, got %d entries", len(vars))
	}
}

func TestLoadFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.env")
	content := "KEY=value\nOTHER=stuff\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	vars, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if vars["KEY"] != "value" {
		t.Errorf("KEY = %q, want value", vars["KEY"])
	}
	if vars["OTHER"] != "stuff" {
		t.Errorf("OTHER = %q, want stuff", vars["OTHER"])
	}
}

func TestLoadFile_MalformedFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.env")
	content := "NOPE"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("LoadFile() should error for malformed file")
	}
}
