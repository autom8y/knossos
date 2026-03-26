// Package conversation manages thread history for multi-turn interactions.
//
// BC-04: ThreadMessage is duplicated here, NOT shared with triage/.
// The conversion from conversation.ThreadMessage to triage.ThreadMessage
// happens in internal/slack/handler.go.
//
// This package sits at the interface layer: internal/slack/conversation/.
// It imports internal/llm/ (infrastructure) for summarization.
// It does NOT import triage/, reason/, or search/.
package conversation

import (
	"context"
	"time"
)

// ThreadMessage represents a single message in a thread.
// BC-04: This type is intentionally duplicated from triage.ThreadMessage.
// The handler converts between the two.
type ThreadMessage struct {
	Role      string    // "user" or "assistant"
	Content   string    // Full text (user) or truncated (assistant)
	Timestamp time.Time // Slack message timestamp
}

// ThreadHistory is the resolved conversation context for a thread.
// Contains the hybrid window: recent messages verbatim + summary of older messages.
type ThreadHistory struct {
	// Summary is the Haiku-generated summary of messages older than the window.
	// Empty string if the thread has fewer than WindowSize+1 messages.
	Summary string

	// RecentMessages is the last N messages in chronological order.
	RecentMessages []ThreadMessage

	// TotalMessageCount is the full thread length (not just the window).
	TotalMessageCount int

	// State is the current lifecycle state of this thread.
	State ThreadState
}

// ThreadState represents the lifecycle state of a conversation thread.
type ThreadState int

const (
	// ThreadCreated: thread initialized, no messages stored yet.
	ThreadCreated ThreadState = iota

	// ThreadActive: messages accumulating, cache populated, within TTL.
	ThreadActive

	// ThreadDormant: cache evicted (TTL expired), thread exists in Slack.
	ThreadDormant

	// ThreadResurrecting: conversations.replies fetch in progress.
	// Transient state during cache miss recovery.
	ThreadResurrecting
)

// String returns a human-readable name for the thread state.
func (s ThreadState) String() string {
	switch s {
	case ThreadCreated:
		return "CREATED"
	case ThreadActive:
		return "ACTIVE"
	case ThreadDormant:
		return "DORMANT"
	case ThreadResurrecting:
		return "RESURRECTING"
	default:
		return "UNKNOWN"
	}
}

// Config holds tuning parameters for the ConversationManager.
type Config struct {
	// MaxRecentMessages is the number of recent messages retained verbatim (N).
	// Default: 5.
	MaxRecentMessages int

	// TTL is the cache entry TTL.
	// Default: 2 hours.
	TTL time.Duration

	// CleanupInterval is the background cleanup frequency.
	// Default: 5 minutes.
	CleanupInterval time.Duration

	// SummaryMaxTokens is the max tokens for thread summary.
	// Default: 250.
	SummaryMaxTokens int

	// ResurrectingTimeout is the max time to block during RESURRECTING state.
	// BC-08: 500ms maximum.
	// Default: 500ms.
	ResurrectingTimeout time.Duration
}

// DefaultConfig returns production defaults.
func DefaultConfig() Config {
	return Config{
		MaxRecentMessages:   5,
		TTL:                 2 * time.Hour,
		CleanupInterval:     5 * time.Minute,
		SummaryMaxTokens:    250,
		ResurrectingTimeout: 500 * time.Millisecond,
	}
}

// SlackThreadFetcher abstracts the Slack conversations.replies API.
// Used for cache miss recovery (DORMANT -> RESURRECTING -> ACTIVE).
type SlackThreadFetcher interface {
	// FetchThreadMessages retrieves messages from a Slack thread.
	// Returns up to limit messages in chronological order.
	// channelID is required for the Slack API call.
	FetchThreadMessages(ctx context.Context, channelID string, threadTS string, limit int) ([]ThreadMessage, error)
}

// Summarizer abstracts the LLM summarization capability.
// Uses the shared llm.Client from internal/llm/ (BC-01).
type Summarizer interface {
	// Summarize generates a summary of the given messages.
	// Returns the summary text, or empty string on failure.
	Summarize(ctx context.Context, messages []ThreadMessage) string
}

// MetricsEmitter abstracts the observability recorder to avoid circular imports.
// Satisfied by observe.EMFRecorder and observe.NopRecorder.
type MetricsEmitter interface {
	IncrConversationHit()
	IncrConversationMiss(reason string)
	RecordConversationGetLatency(result string, duration time.Duration)
	SetActiveThreads(count int)
	IncrEvictions()
	SetConversationMemoryBytes(bytes int64)
	SetStartupTimestamp(t time.Time)
}
