package slack

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/serve/webhook"
)

// eventDedup provides event ID deduplication with TTL-based expiration (TD-02 fix).
// Slack may retry events; this prevents duplicate pipeline invocations.
type eventDedup struct {
	mu    sync.Mutex
	seen  map[string]time.Time
	ttl   time.Duration
}

func newEventDedup(ttl time.Duration) *eventDedup {
	d := &eventDedup{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
	// Background cleanup every TTL/2.
	go func() {
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()
		for range ticker.C {
			d.cleanup()
		}
	}()
	return d
}

// isDuplicate returns true if the event ID was already seen within the TTL window.
func (d *eventDedup) isDuplicate(eventID string) bool {
	if eventID == "" {
		return false // no event ID = can't dedup, process it
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.seen[eventID]; ok {
		return true
	}
	d.seen[eventID] = time.Now()
	return false
}

func (d *eventDedup) cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := time.Now().Add(-d.ttl)
	for id, ts := range d.seen {
		if ts.Before(cutoff) {
			delete(d.seen, id)
		}
	}
}

// concurrencyLimiter limits concurrent pipeline invocations (TD-03 fix).
// Prevents burst events from spawning unbounded Claude API calls.
type concurrencyLimiter struct {
	sem chan struct{}
}

func newConcurrencyLimiter(maxConcurrent int) *concurrencyLimiter {
	return &concurrencyLimiter{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// tryAcquire attempts to acquire a slot without blocking. Returns false if at capacity.
func (l *concurrencyLimiter) tryAcquire() bool {
	select {
	case l.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l *concurrencyLimiter) release() {
	<-l.sem
}

// QueryRunner abstracts the reasoning pipeline for testability.
// The concrete implementation is *reason.Pipeline.
type QueryRunner interface {
	Query(ctx context.Context, question string) (*response.ReasoningResponse, error)
}

// defaultSuggestedPrompts are the initial prompts shown when a user starts an assistant thread.
var defaultSuggestedPrompts = []string{
	"What is the architecture of this project?",
	"What conventions does this codebase follow?",
	"What are the known design constraints?",
}

// NewSlackHandler returns an http.HandlerFunc that processes Slack Events API payloads.
// The handler routes events to the reasoning pipeline and renders responses.
//
// The request body has already been restored by the upstream verification middleware
// (webhook.Verifier.Handler).
func NewSlackHandler(pipeline QueryRunner, client *SlackClient, cfg SlackConfig) http.HandlerFunc {
	dedup := newEventDedup(5 * time.Minute)       // TD-02: 5-minute dedup window
	limiter := newConcurrencyLimiter(5)            // TD-03: max 5 concurrent pipeline queries

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("failed to read request body", "error", err)
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		// Handle URL verification challenge (already authenticated by middleware).
		if webhook.HandleChallenge(w, r, body) {
			return
		}

		// Parse the event envelope.
		var envelope EventEnvelope
		if err := json.Unmarshal(body, &envelope); err != nil {
			slog.Error("failed to parse event envelope", "error", err)
			http.Error(w, "invalid event payload", http.StatusBadRequest)
			return
		}

		// Only process event_callback type.
		if envelope.Type != "event_callback" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// TD-02: Event deduplication — reject already-seen event IDs.
		if dedup.isDuplicate(envelope.EventID) {
			slog.Debug("duplicate event filtered", "event_id", envelope.EventID)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Peek at the inner event type.
		var inner innerEventType
		if err := json.Unmarshal(envelope.Event, &inner); err != nil {
			slog.Error("failed to parse inner event type", "error", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		switch inner.Type {
		case "assistant_thread_started":
			handleAssistantThreadStarted(w, envelope.Event, client)
		case "message":
			handleMessage(w, envelope.Event, pipeline, client, limiter)
		default:
			slog.Debug("unhandled event type", "type", inner.Type)
			w.WriteHeader(http.StatusOK)
		}
	}
}

// handleAssistantThreadStarted processes an assistant_thread_started event.
// Sets default suggested prompts and acknowledges the event.
func handleAssistantThreadStarted(w http.ResponseWriter, eventData json.RawMessage, client *SlackClient) {
	var event AssistantThreadEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		slog.Error("failed to parse assistant_thread_started event", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	channelID := event.AssistantThread.ChannelID
	threadTS := event.AssistantThread.ThreadTS

	// Set suggested prompts asynchronously (non-blocking to Slack's 3s timeout).
	go func() {
		if err := client.SetSuggestedPrompts(channelID, threadTS, defaultSuggestedPrompts); err != nil {
			slog.Warn("failed to set suggested prompts",
				"channel", channelID,
				"thread_ts", threadTS,
				"error", err,
			)
		}
	}()

	w.WriteHeader(http.StatusOK)
}

// handleMessage processes a Slack message event.
// SECURITY: Bot messages are filtered before pipeline invocation to prevent
// prompt injection via bot-to-bot message chains.
func handleMessage(w http.ResponseWriter, eventData json.RawMessage, pipeline QueryRunner, client *SlackClient, limiter *concurrencyLimiter) {
	var msg MessageEvent
	if err := json.Unmarshal(eventData, &msg); err != nil {
		slog.Error("failed to parse message event", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// BOT FILTER: Reject messages from bots to prevent prompt injection.
	if msg.BotID != "" {
		slog.Info("filtered bot message",
			"bot_id", msg.BotID,
			"channel", msg.Channel,
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Reject messages with a subtype (edited, deleted, etc.) -- only process plain messages.
	if msg.SubType != "" {
		slog.Debug("skipping message with subtype", "subtype", msg.SubType)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Reject empty messages.
	if msg.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Respond 200 immediately to meet Slack's 3-second acknowledgment deadline.
	w.WriteHeader(http.StatusOK)

	// Determine thread context: reply in thread if already threaded, else start a thread.
	threadTS := msg.ThreadTS
	if threadTS == "" {
		threadTS = msg.TS
	}

	// TD-03: Rate limit concurrent pipeline queries.
	if !limiter.tryAcquire() {
		slog.Warn("rate limit exceeded, dropping message",
			"channel", msg.Channel,
			"user", msg.User,
		)
		go func() {
			_ = client.SendBlocks(msg.Channel, threadTS, RenderRateLimited())
		}()
		return
	}

	// Process asynchronously with limiter release on completion.
	go func() {
		defer limiter.release()
		processMessage(msg.Channel, threadTS, msg.Text, pipeline, client)
	}()
}

// processMessage runs the reasoning pipeline and posts the response.
// Runs in a goroutine -- must not reference the http.ResponseWriter.
func processMessage(channelID, threadTS, question string, pipeline QueryRunner, client *SlackClient) {
	// Set "thinking" status.
	if err := client.SetStatus(channelID, threadTS, "", "Searching knowledge..."); err != nil {
		slog.Warn("failed to set processing status",
			"channel", channelID,
			"error", err,
		)
	}

	// Run the reasoning pipeline.
	resp, err := pipeline.Query(context.Background(), question)
	if err != nil {
		slog.Error("pipeline query failed",
			"channel", channelID,
			"question", question,
			"error", err,
		)
		_ = client.SetStatus(channelID, threadTS, "", "Error: "+err.Error())
		return
	}

	// Render response as Block Kit blocks.
	blocks := RenderResponse(resp)

	// Send the rendered response.
	if err := client.SendBlocks(channelID, threadTS, blocks); err != nil {
		slog.Error("failed to send response blocks",
			"channel", channelID,
			"error", err,
		)
		_ = client.SetStatus(channelID, threadTS, "", "Error sending response")
		return
	}

	// Set thread title (first 60 chars of question).
	title := question
	if len(title) > 60 {
		title = title[:60]
	}
	if err := client.SetTitle(channelID, threadTS, title); err != nil {
		slog.Warn("failed to set thread title",
			"channel", channelID,
			"error", err,
		)
	}

	// Mark processing complete.
	if err := client.SetStatus(channelID, threadTS, "", ""); err != nil {
		slog.Warn("failed to clear status",
			"channel", channelID,
			"error", err,
		)
	}
}
