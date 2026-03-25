package observe

import (
	"log/slog"
	"os"
	"strings"
)

// ConfigureStructuredLogging sets the slog default to a JSON handler with the
// specified log level. The level string is case-insensitive and accepts
// DEBUG, INFO, WARN, ERROR. Defaults to INFO on unrecognised input.
func ConfigureStructuredLogging(level string) {
	var lvl slog.Level

	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "WARN", "WARNING":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})
	slog.SetDefault(slog.New(handler))
}
