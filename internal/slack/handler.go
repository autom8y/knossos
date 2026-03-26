package slack

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/serve/webhook"
	"github.com/autom8y/knossos/internal/slack/conversation"
	"github.com/autom8y/knossos/internal/slack/streaming"
)

// eventDedup provides event ID deduplication with TTL-based expiration (TD-02 fix).
// Slack may retry events; this prevents duplicate pipeline invocations.
type eventDedup struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
	done chan struct{}
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
		done: make(chan struct{}),
	}
	// Background cleanup every TTL/2. Stops when done is closed.
	go func() {
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d.cleanup()
			case <-d.done:
				return
			}
		}
	}()
	return d
}

// Stop terminates the background cleanup goroutine.
func (d *eventDedup) Stop() {
	close(d.done)
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

// size returns the current number of entries in the dedup map.
func (d *eventDedup) size() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
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

// active returns the number of currently acquired slots.
func (l *concurrencyLimiter) active() int {
	return len(l.sem)
}

// QueryRunner abstracts the reasoning pipeline for testability.
// The concrete implementation is *reason.Pipeline.
type QueryRunner interface {
	Query(ctx context.Context, question string) (*response.ReasoningResponse, error)
}

// TriageQueryRunner abstracts the reasoning pipeline's triage-aware entry point.
// WARNING-03 fix: processWithTriage MUST use this interface (not QueryRunner.Query)
// so that triage candidates reach the assembler for BC-07 weighted-mean freshness.
type TriageQueryRunner interface {
	QueryWithTriage(ctx context.Context, triageInput *TriageResultInputData) (*response.ReasoningResponse, error)
}

// StreamingQueryRunner extends QueryRunner with streaming support.
// BC-03: Uses onChunk callback. reason/ does NOT import slack/.
type StreamingQueryRunner interface {
	QueryStream(ctx context.Context, triageInput *TriageResultInputData, onChunk func(chunk string)) (*response.ReasoningResponse, error)
}

// TriageResultInputData is the handler-local TriageResultInput for the pipeline.
// Avoids importing reason/ types directly in the interface.
type TriageResultInputData struct {
	RefinedQuery   string
	Candidates     []TriageCandidateData
	ModelCallCount int

	// ConversationHistory holds recent conversation turns for multi-turn synthesis.
	// WS-2: Populated from ConversationManager's thread history. Passed through
	// to the reasoning pipeline which injects it into the system prompt.
	ConversationHistory []reasoncontext.ConversationTurn
}

// TriageRunner abstracts the triage orchestrator for testability.
// The concrete implementation is *triage.Orchestrator (passed as data, not import).
type TriageRunner interface {
	Assess(ctx context.Context, query string, threadHistory []TriageThreadMessage) (*TriageResultData, error)
}

// TriageThreadMessage is the handler-local thread message type.
// Converted from triage.ThreadMessage by the handler.
type TriageThreadMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// TriageResultData holds triage results in a handler-consumable form.
// The handler converts this to reason.TriageResultInput for the pipeline.
type TriageResultData struct {
	RefinedQuery   string
	Candidates     []TriageCandidateData
	ModelCallCount int
}

// TriageCandidateData holds a single triage candidate in handler-consumable form.
type TriageCandidateData struct {
	QualifiedName       string
	RelevanceScore      float64
	EmbeddingSimilarity float64
	Freshness           float64
	Rationale           string
	DomainType          string
	RelatedDomains      []string
}

// defaultSuggestedPrompts are the initial prompts shown when a user starts an assistant thread.
var defaultSuggestedPrompts = []string{
	"How are our projects structured?",
	"What practices and conventions do we follow?",
	"What decisions have shaped our technical direction?",
}

// HandlerMetrics abstracts the observability recorder to avoid circular imports.
// Satisfied by observe.EMFRecorder and observe.NopRecorder.
type HandlerMetrics interface {
	RecordPrePipelineLatency(cmResult string, duration time.Duration)
	SetConcurrentQueries(count int)
	IncrDropped(reason string)
	SetDedupMapSize(size int)
	IncrDedupDrops()
	RecordHaikuCalls(count int)
}

// HandlerDeps holds all dependencies for the Slack event handler.
type HandlerDeps struct {
	Pipeline     QueryRunner
	Client       *SlackClient
	Config       SlackConfig
	TriageRunner TriageRunner

	// TriagePipeline handles queries with pre-computed triage candidates.
	// WARNING-03 fix: processWithTriage uses this to pass candidate data
	// through to the assembler for BC-07 weighted-mean freshness.
	// May be nil -- when nil, falls back to QueryRunner.Query (v1 path).
	TriagePipeline TriageQueryRunner

	// ConversationMgr tracks multi-turn conversation state.
	// May be nil -- when nil, conversation history is not available.
	ConversationMgr *conversation.Manager

	// StreamSender renders progressive streaming responses.
	// May be nil -- when nil, responses are posted as single messages.
	StreamSender *streaming.Sender

	// StreamingRunner executes the streaming pipeline.
	// May be nil -- when nil, uses non-streaming pipeline.
	StreamingRunner StreamingQueryRunner

	// Metrics records pre-pipeline and handler-level metrics.
	// May be nil -- when nil, metrics recording is skipped.
	Metrics HandlerMetrics

	// MaxConcurrent is the maximum number of concurrent pipeline queries.
	// Default: 5 if zero.
	MaxConcurrent int
}

// NewSlackHandler returns an http.HandlerFunc that processes Slack Events API payloads
// and a ThreadContextStore for pipeline access to assistant thread context.
// The handler routes events to the reasoning pipeline and renders responses.
//
// triageRunner may be nil -- when nil, the handler falls back to v1 pipeline (Query only).
//
// The request body has already been restored by the upstream verification middleware
// (webhook.Verifier.Handler).
func NewSlackHandler(pipeline QueryRunner, client *SlackClient, cfg SlackConfig, triageRunner TriageRunner) (http.HandlerFunc, *ThreadContextStore, func()) {
	return NewSlackHandlerWithDeps(HandlerDeps{
		Pipeline:     pipeline,
		Client:       client,
		Config:       cfg,
		TriageRunner: triageRunner,
	})
}

// NewSlackHandlerWithDeps creates the handler with full Sprint 6 dependencies.
// Returns the handler, a ThreadContextStore, and a stop function that must be
// called during server shutdown to terminate background goroutines.
func NewSlackHandlerWithDeps(deps HandlerDeps) (http.HandlerFunc, *ThreadContextStore, func()) {
	dedup := newEventDedup(5 * time.Minute)             // TD-02: 5-minute dedup window
	maxConcurrent := deps.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}
	limiter := newConcurrencyLimiter(maxConcurrent)      // TD-03: configurable concurrent pipeline queries
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
			if deps.Metrics != nil {
				deps.Metrics.IncrDedupDrops()
				deps.Metrics.SetDedupMapSize(dedup.size())
			}
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
			handleAssistantThreadStarted(w, envelope.Event, deps.Client, deps.ConversationMgr)
		case "assistant_thread_context_changed":
			handleAssistantThreadContextChanged(w, envelope.Event, ctxStore)
		case "message":
			handleMessage(w, envelope.Event, deps, limiter, ctxStore)
		default:
			slog.Debug("unhandled event type", "type", inner.Type)
			w.WriteHeader(http.StatusOK)
		}
	}, ctxStore, func() { dedup.Stop() }
}

// handleAssistantThreadStarted processes an assistant_thread_started event.
// Sets default suggested prompts and initializes conversation tracking.
func handleAssistantThreadStarted(w http.ResponseWriter, eventData json.RawMessage, client *SlackClient, convMgr *conversation.Manager) {
	var event AssistantThreadEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		slog.Error("failed to parse assistant_thread_started event", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	channelID := event.AssistantThread.ChannelID
	threadTS := event.AssistantThread.ThreadTS

	// Initialize conversation tracking for this thread.
	if convMgr != nil {
		convMgr.InitThread(threadTS, channelID)
	}

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
func handleMessage(w http.ResponseWriter, eventData json.RawMessage, deps HandlerDeps, limiter *concurrencyLimiter, ctxStore *ThreadContextStore) {
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

	// WS-2.3: Retrieve stored thread context from ThreadContextStore.
	if threadCtx, ok := ctxStore.Get(threadTS); ok {
		slog.Info("thread context available for message",
			"thread_ts", threadTS,
			"channel", msg.Channel,
			"context_bytes", len(threadCtx),
		)
	}

	// TD-03: Rate limit concurrent pipeline queries.
	if !limiter.tryAcquire() {
		slog.Warn("rate limit exceeded, dropping message",
			"channel", msg.Channel,
			"user", msg.User,
		)
		if deps.Metrics != nil {
			deps.Metrics.IncrDropped("rate_limited")
		}
		go func() {
			_ = deps.Client.SendBlocks(msg.Channel, threadTS, RenderRateLimited())
		}()
		return
	}

	// Record concurrent query gauge.
	if deps.Metrics != nil {
		deps.Metrics.SetConcurrentQueries(limiter.active())
	}

	// Process asynchronously with limiter release on completion.
	go func() {
		defer func() {
			limiter.release()
			if deps.Metrics != nil {
				deps.Metrics.SetConcurrentQueries(limiter.active())
			}
		}()

		// Pre-pipeline timing starts here (includes ConversationManager lookup).
		prePipelineStart := time.Now()
		cmResult := "no_cm"

		// Get conversation history for multi-turn context.
		var threadHistory []TriageThreadMessage
		var conversationTurns []reasoncontext.ConversationTurn
		if deps.ConversationMgr != nil {
			ctx := context.Background()
			history := deps.ConversationMgr.GetThreadHistory(ctx, threadTS)
			if history != nil {
				cmResult = "hit"
				// BC-04: Convert conversation.ThreadMessage to handler-local type.
				threadHistory = convertThreadHistory(history)
				// WS-2: Convert to ConversationTurn for synthesis prompt injection.
				conversationTurns = convertToConversationTurns(history)
				slog.Info("conversation history retrieved",
					"thread_ts", threadTS,
					"total_messages", history.TotalMessageCount,
					"recent_messages", len(history.RecentMessages),
					"conversation_turns", len(conversationTurns),
					"has_summary", history.Summary != "",
				)
			} else {
				cmResult = "miss"
			}
		}

		// Record pre-pipeline latency (ConversationManager lookup cost).
		if deps.Metrics != nil {
			deps.Metrics.RecordPrePipelineLatency(cmResult, time.Since(prePipelineStart))
		}

		processMessage(msg.Channel, threadTS, msg.Text, deps, threadHistory, conversationTurns)
	}()
}

// convertThreadHistory converts conversation.ThreadHistory to handler-local thread messages.
// BC-04: This is the conversion point between conversation/ and triage/ types.
func convertThreadHistory(history *conversation.ThreadHistory) []TriageThreadMessage {
	if history == nil {
		return nil
	}

	var messages []TriageThreadMessage
	for _, m := range history.RecentMessages {
		messages = append(messages, TriageThreadMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}
	return messages
}

// convertToConversationTurns converts conversation.ThreadHistory to ConversationTurn
// for injection into the synthesis system prompt.
// WS-2: Uses the reason/context.ConversationTurn type (consumer-side definition).
// Default: 3 user turns (windowed by buildHistory's user-turn algorithm).
func convertToConversationTurns(history *conversation.ThreadHistory) []reasoncontext.ConversationTurn {
	if history == nil {
		return nil
	}

	var turns []reasoncontext.ConversationTurn
	for _, m := range history.RecentMessages {
		turns = append(turns, reasoncontext.ConversationTurn{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return turns
}

// processMessage runs the reasoning pipeline and posts the response.
// Runs in a goroutine -- must not reference the http.ResponseWriter.
// WS-2: conversationTurns carries recent thread turns for synthesis prompt injection.
func processMessage(channelID, threadTS, question string, deps HandlerDeps, threadHistory []TriageThreadMessage, conversationTurns []reasoncontext.ConversationTurn) {
	client := deps.Client

	// Set "thinking" status.
	if err := client.SetStatus(channelID, threadTS, "", "Searching knowledge..."); err != nil {
		slog.Warn("failed to set processing status",
			"channel", channelID,
			"error", err,
		)
	}

	// GAP-6: Run the reasoning pipeline with a 60-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var resp *response.ReasoningResponse
	var err error

	// Sprint 5/6: Wire triage with thread history into message handling.
	// WS-2: Pass conversation turns for synthesis prompt injection.
	if deps.TriageRunner != nil {
		resp, err = processWithTriage(ctx, question, deps, threadHistory, conversationTurns)
	} else {
		resp, err = deps.Pipeline.Query(ctx, question)
	}
	if err != nil {
		slog.Error("pipeline query failed",
			"channel", channelID,
			"question", question,
			"error", err,
		)
		_ = client.SetStatus(channelID, threadTS, "", "Error: "+err.Error())
		return
	}

	// Store the user message and assistant response in conversation history.
	if deps.ConversationMgr != nil {
		deps.ConversationMgr.StoreMessage(ctx, threadTS, channelID, conversation.ThreadMessage{
			Role:      "user",
			Content:   question,
			Timestamp: time.Now(),
		})
		if resp != nil && resp.Answer != "" {
			deps.ConversationMgr.StoreMessage(ctx, threadTS, channelID, conversation.ThreadMessage{
				Role:      "assistant",
				Content:   resp.Answer,
				Timestamp: time.Now(),
			})
		}
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

// processWithTriage runs triage first, then passes candidates to the pipeline.
// WARNING-03 fix: uses TriagePipeline.QueryWithTriage to pass candidate data
// through to the assembler for BC-07 weighted-mean freshness computation.
// Falls back to v1 pipeline (Query) when triage returns nil, errors, or
// TriagePipeline is not wired.
// WS-2: conversationTurns is forwarded to the pipeline for synthesis prompt injection.
func processWithTriage(ctx context.Context, question string, deps HandlerDeps, threadHistory []TriageThreadMessage, conversationTurns []reasoncontext.ConversationTurn) (*response.ReasoningResponse, error) {
	// Run triage (Stages 0-3) with thread history for multi-turn context.
	triageStart := time.Now()
	triageResult, err := deps.TriageRunner.Assess(ctx, question, threadHistory)
	triageElapsed := time.Since(triageStart)
	if err != nil {
		slog.Warn("triage failed, falling back to v1 pipeline",
			"error", err,
		)
		return deps.Pipeline.Query(ctx, question)
	}

	if triageResult == nil || len(triageResult.Candidates) == 0 {
		slog.Info("triage returned no candidates, falling back to v1 pipeline")
		return deps.Pipeline.Query(ctx, question)
	}

	// Record triage metrics: Haiku call count and stage 3 latency.
	if deps.Metrics != nil {
		deps.Metrics.RecordHaikuCalls(triageResult.ModelCallCount)
		// Record triage latency as stage3 (the dominant triage cost).
		deps.Metrics.RecordPrePipelineLatency("triage", triageElapsed)
	}

	// Use the refined query from triage for improved search quality.
	refinedQuery := triageResult.RefinedQuery
	if refinedQuery == "" {
		refinedQuery = question
	}

	slog.Info("triage complete, using refined query",
		"original", question,
		"refined", refinedQuery,
		"candidates", len(triageResult.Candidates),
		"model_calls", triageResult.ModelCallCount,
	)

	// WARNING-03 fix: Pass triage candidates through to the pipeline so the
	// assembler receives RelevanceScores for BC-07 weighted-mean freshness.
	// Without this, triage candidate data is discarded and freshness falls
	// back to position-weighted chain computation.
	if deps.TriagePipeline != nil {
		triageInput := &TriageResultInputData{
			RefinedQuery:        refinedQuery,
			Candidates:          triageResult.Candidates,
			ModelCallCount:      triageResult.ModelCallCount,
			ConversationHistory: conversationTurns, // WS-2: forward conversation turns to pipeline.
		}
		return deps.TriagePipeline.QueryWithTriage(ctx, triageInput)
	}

	// Fallback: TriagePipeline not wired. Use v1 pipeline with refined query.
	slog.Warn("TriagePipeline not wired, triage candidates discarded")
	return deps.Pipeline.Query(ctx, refinedQuery)
}
