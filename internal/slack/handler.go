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

// threadContextEntry holds a stored thread context with expiration metadata.
type threadContextEntry struct {
	context  json.RawMessage
	storedAt time.Time
}

// ThreadContextStore provides goroutine-safe storage for assistant thread context.
// When users navigate to a different channel while the assistant container is open,
// Slack sends an assistant_thread_context_changed event with channel context.
// This store holds that context keyed by thread timestamp for pipeline retrieval.
type ThreadContextStore struct {
	mu      sync.Mutex
	entries map[string]threadContextEntry
	ttl     time.Duration
	done    chan struct{}
}

// newThreadContextStore creates a ThreadContextStore with TTL-based expiration.
// The cleanup goroutine runs every ttl/2 and stops when the done channel is closed.
func newThreadContextStore(ttl time.Duration) *ThreadContextStore {
	s := &ThreadContextStore{
		entries: make(map[string]threadContextEntry),
		ttl:     ttl,
		done:    make(chan struct{}),
	}
	go func() {
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanup()
			case <-s.done:
				return
			}
		}
	}()
	return s
}

// Set stores thread context for the given thread timestamp.
func (s *ThreadContextStore) Set(threadTS string, ctx json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[threadTS] = threadContextEntry{
		context:  ctx,
		storedAt: time.Now(),
	}
}

// Get retrieves the stored context for a thread timestamp.
// Returns the context and true if found and not expired, or nil and false otherwise.
func (s *ThreadContextStore) Get(threadTS string) (json.RawMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[threadTS]
	if !ok {
		return nil, false
	}
	if time.Since(entry.storedAt) > s.ttl {
		delete(s.entries, threadTS)
		return nil, false
	}
	return entry.context, true
}

// Stop terminates the background cleanup goroutine.
func (s *ThreadContextStore) Stop() {
	close(s.done)
}

func (s *ThreadContextStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-s.ttl)
	for ts, entry := range s.entries {
		if entry.storedAt.Before(cutoff) {
			delete(s.entries, ts)
		}
	}
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

// NewSlackHandler returns an http.HandlerFunc that processes Slack Events API payloads
// and a ThreadContextStore for pipeline access to assistant thread context.
// The handler routes events to the reasoning pipeline and renders responses.
//
// The request body has already been restored by the upstream verification middleware
// (webhook.Verifier.Handler).
func NewSlackHandler(pipeline QueryRunner, client *SlackClient, cfg SlackConfig) (http.HandlerFunc, *ThreadContextStore) {
	dedup := newEventDedup(5 * time.Minute)             // TD-02: 5-minute dedup window
	limiter := newConcurrencyLimiter(5)                  // TD-03: max 5 concurrent pipeline queries
	ctxStore := newThreadContextStore(30 * time.Minute)  // GAP-10: 30-minute thread context TTL

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
		case "assistant_thread_context_changed":
			handleAssistantThreadContextChanged(w, envelope.Event, ctxStore)
		case "message":
			handleMessage(w, envelope.Event, pipeline, client, limiter)
		default:
			slog.Debug("unhandled event type", "type", inner.Type)
			w.WriteHeader(http.StatusOK)
		}
	}, ctxStore
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

// handleAssistantThreadContextChanged processes an assistant_thread_context_changed event.
// Stores the updated channel context so the pipeline can provide context-aware responses.
// This is an acknowledge-only event -- no response is sent to Slack.
func handleAssistantThreadContextChanged(w http.ResponseWriter, eventData json.RawMessage, ctxStore *ThreadContextStore) {
	var event AssistantThreadContextChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		slog.Error("failed to parse assistant_thread_context_changed event", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	threadTS := event.AssistantThread.ThreadTS
	threadCtx := event.AssistantThread.Context

	if threadTS != "" && len(threadCtx) > 0 {
		ctxStore.Set(threadTS, threadCtx)
		slog.Info("stored assistant thread context",
			"thread_ts", threadTS,
			"channel_id", event.AssistantThread.ChannelID,
		)
	} else {
		slog.Debug("assistant_thread_context_changed with empty thread_ts or context",
			"thread_ts", threadTS,
		)
	}

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

	// GAP-6: Run the reasoning pipeline with a 60-second timeout.
	// Prevents hanging API calls from exhausting concurrency limiter slots.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resp, err := pipeline.Query(ctx, question)
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
