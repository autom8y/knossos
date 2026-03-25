package envload

import (
	"fmt"
	"log/slog"
	"os"
)

// CheckPermissions verifies that the env file has restrictive permissions.
// Logs a warning via slog if the file is group- or world-readable.
// Does not return an error -- overly permissive files are a warning, not a blocker.
func CheckPermissions(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	mode := info.Mode().Perm()
	if mode&0o077 != 0 {
		slog.Warn("env file has overly permissive permissions",
			"path", path,
			"mode", fmt.Sprintf("%04o", mode),
			"recommended", "0600",
		)
	}
}

// CreateTemplate writes a serve.env template to the given path with 0600 permissions.
// Returns an error if the file already exists to prevent accidental overwrite.
func CreateTemplate(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("env file already exists at %s", path)
	}

	content := `# Org-level configuration for ari serve.
# Location: $XDG_DATA_HOME/knossos/orgs/{org}/serve.env
# Permissions: 0600 (owner read/write only)
#
# Resolution priority: CLI flags > process env > this file > defaults

# Required: Anthropic Claude API key (reasoning pipeline)
ANTHROPIC_API_KEY=

# Required: Slack app signing secret (webhook verification)
SLACK_SIGNING_SECRET=

# Required: Slack bot OAuth token (posting responses)
SLACK_BOT_TOKEN=

# Optional: Server configuration (defaults shown)
# PORT=8080
# LOG_LEVEL=INFO
# MAX_CONCURRENT=10

# Optional: OTEL collector endpoint (empty = noop tracing)
# OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
`
	return os.WriteFile(path, []byte(content), 0o600)
}
