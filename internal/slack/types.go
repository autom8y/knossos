// Package slack implements the Slack surface for Clew, routing events to the
// reasoning pipeline and rendering trust-tiered responses as Block Kit messages.
//
// Layer boundaries:
//   - slack/ imports internal/reason/response (types only), internal/trust (tier constants)
//   - slack/ imports internal/serve/webhook (challenge handling)
//   - slack/ does NOT import internal/reason (pipeline) -- uses QueryRunner interface
//   - slack/ does NOT import internal/cmd/
package slack

import "encoding/json"

// EventEnvelope is the top-level JSON structure sent by the Slack Events API.
// The Type field discriminates between challenge requests and event callbacks.
type EventEnvelope struct {
	// Type is "event_callback" for real events, "url_verification" for challenges.
	Type string `json:"type"`

	// Challenge is populated only for url_verification requests.
	Challenge string `json:"challenge,omitempty"`

	// Event is the raw inner event payload. Parsed based on event type.
	Event json.RawMessage `json:"event"`

	// TeamID is the Slack workspace identifier.
	TeamID string `json:"team_id"`

	// EventID is the unique identifier for this event delivery.
	EventID string `json:"event_id"`
}

// MessageEvent represents a Slack message event from the Events API.
type MessageEvent struct {
	// Type is always "message".
	Type string `json:"type"`

	// SubType is empty for normal user messages. Non-empty values include
	// "bot_message", "message_changed", etc.
	SubType string `json:"subtype"`

	// Text is the message content.
	Text string `json:"text"`

	// User is the Slack user ID of the sender. Empty for bot messages.
	User string `json:"user"`

	// BotID is non-empty when the message was sent by a bot.
	// SECURITY: Messages with a non-empty BotID MUST be filtered to prevent
	// prompt injection via bot-to-bot message chains.
	BotID string `json:"bot_id"`

	// Channel is the channel or DM ID where the message was posted.
	Channel string `json:"channel"`

	// ThreadTS is the timestamp of the parent message in a thread.
	// Empty for top-level messages.
	ThreadTS string `json:"thread_ts"`

	// TS is the message's own timestamp (unique message ID within a channel).
	TS string `json:"ts"`
}

// AssistantThreadEvent represents a Slack assistant_thread_started event.
// Fired when a user opens a new AI assistant thread.
type AssistantThreadEvent struct {
	// Type is "assistant_thread_started".
	Type string `json:"type"`

	// AssistantThread contains the thread details.
	AssistantThread AssistantThreadInfo `json:"assistant_thread"`
}

// AssistantThreadInfo holds the channel and thread identifiers for an assistant thread.
type AssistantThreadInfo struct {
	// ChannelID is the channel where the thread was started.
	ChannelID string `json:"channel_id"`

	// ThreadTS is the thread timestamp.
	ThreadTS string `json:"thread_ts"`

	// Context holds any initial context (may be empty on first start).
	Context json.RawMessage `json:"context,omitempty"`
}

// innerEventType is a minimal struct for peeking at the event type field
// before full deserialization.
type innerEventType struct {
	Type string `json:"type"`
}
