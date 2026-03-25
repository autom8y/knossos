package slack

import "time"

// SlackConfig holds configuration for the Slack integration.
type SlackConfig struct {
	// BotToken is the Slack bot OAuth token (xoxb-...).
	BotToken string

	// AppToken is the Slack app-level token (xapp-...).
	// Used for Socket Mode (not used in HTTP mode).
	AppToken string

	// StreamingEnabled controls whether progressive response streaming is used.
	// When true, messages are posted incrementally as the pipeline produces output.
	// Default: true.
	StreamingEnabled bool

	// StreamChunkSize is the minimum character count before sending a streaming update.
	// Default: 100.
	StreamChunkSize int

	// StreamChunkDelay is the minimum delay between streaming updates.
	// Default: 50ms.
	StreamChunkDelay time.Duration
}

// DefaultSlackConfig returns a SlackConfig with production defaults.
func DefaultSlackConfig() SlackConfig {
	return SlackConfig{
		StreamingEnabled: true,
		StreamChunkSize:  100,
		StreamChunkDelay: 50 * time.Millisecond,
	}
}
