package serve

import (
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestNewServeCmd(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""

	cmd := NewServeCmd(&output, &verbose, &projectDir)

	if cmd.Use != "serve" {
		t.Errorf("expected Use 'serve', got %q", cmd.Use)
	}

	// Verify flags exist with correct defaults.
	// Port and max-concurrent default to 0 (meaning "resolve from hierarchy").
	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("expected --port flag")
	}
	if portFlag.DefValue != "0" {
		t.Errorf("expected port default 0, got %q", portFlag.DefValue)
	}

	secretFlag := cmd.Flags().Lookup("slack-signing-secret")
	if secretFlag == nil {
		t.Fatal("expected --slack-signing-secret flag")
	}

	tokenFlag := cmd.Flags().Lookup("slack-bot-token")
	if tokenFlag == nil {
		t.Fatal("expected --slack-bot-token flag")
	}

	drainFlag := cmd.Flags().Lookup("drain-timeout")
	if drainFlag == nil {
		t.Fatal("expected --drain-timeout flag")
	}
	if drainFlag.DefValue != "0s" {
		t.Errorf("expected drain-timeout default 0s, got %q", drainFlag.DefValue)
	}

	maxConcurrentFlag := cmd.Flags().Lookup("max-concurrent")
	if maxConcurrentFlag == nil {
		t.Fatal("expected --max-concurrent flag")
	}
	if maxConcurrentFlag.DefValue != "0" {
		t.Errorf("expected max-concurrent default 0, got %q", maxConcurrentFlag.DefValue)
	}

	// Verify new flags exist.
	orgFlag := cmd.Flags().Lookup("org")
	if orgFlag == nil {
		t.Fatal("expected --org flag")
	}
	if orgFlag.DefValue != "" {
		t.Errorf("expected org default empty, got %q", orgFlag.DefValue)
	}

	envFileFlag := cmd.Flags().Lookup("env-file")
	if envFileFlag == nil {
		t.Fatal("expected --env-file flag")
	}
	if envFileFlag.DefValue != "" {
		t.Errorf("expected env-file default empty, got %q", envFileFlag.DefValue)
	}
}

func TestNewServeCmd_NeedsProject(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""

	cmd := NewServeCmd(&output, &verbose, &projectDir)

	// ari serve should NOT require project context
	val, ok := cmd.Annotations["needsProject"]
	if !ok {
		t.Fatal("expected needsProject annotation to be set")
	}
	if val != "false" {
		t.Errorf("expected needsProject=false, got %q", val)
	}
}

func TestRunServe_MissingSigningSecret(t *testing.T) {
	// Ensure env vars are clean.
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("KNOSSOS_ORG", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	output := "text"
	verbose := false
	projectDir := ""

	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &output,
			Verbose:    &verbose,
			ProjectDir: &projectDir,
		},
	}

	opts := serveOptions{}
	err := runServe(ctx, opts)
	if err == nil {
		t.Fatal("expected error for missing signing secret")
	}
}

func TestRunServe_MissingBotToken(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "test-secret")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("KNOSSOS_ORG", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	output := "text"
	verbose := false
	projectDir := ""

	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &output,
			Verbose:    &verbose,
			ProjectDir: &projectDir,
		},
	}

	opts := serveOptions{}
	err := runServe(ctx, opts)
	if err == nil {
		t.Fatal("expected error for missing bot token")
	}
}
