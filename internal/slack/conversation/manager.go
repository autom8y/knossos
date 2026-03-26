package conversation

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Manager provides thread history persistence with hybrid windowing.
// In-memory implementation with TTL-based eviction and Slack API fallback.
//
// Thread lifecycle state machine:
//
//	CREATED -> ACTIVE (first message stored)
//	ACTIVE -> DORMANT (TTL expired, cleaned up)
//	DORMANT -> RESURRECTING (GetThreadHistory called, Slack API recovery starts)
//	RESURRECTING -> ACTIVE (API recovery completes)
//
// BC-08: GetThreadHistory blocks up to 500ms during RESURRECTING state.
// Returns nil if unresolved. Never errors.
type Manager struct {
	mu      sync.RWMutex
	threads map[string]*threadEntry
	config  Config

	// Dependencies (optional, fail-open).
	summarizer Summarizer
	fetcher    SlackThreadFetcher
	emitter    MetricsEmitter

	// Metrics (counters, gauges — stub placeholders for now).
	metricsMu sync.Mutex
	metrics   *Metrics

	// Lifecycle.
	done      chan struct{}
	startedAt time.Time
}

// threadEntry is the internal representation of a tracked conversation thread.
type threadEntry struct {
	messages    []ThreadMessage
	summary     string
	state       ThreadState
	lastAccess  time.Time
	channelID   string // Needed for Slack API recovery.
	resurrectCh chan struct{}
}

// Metrics holds metric counters for observability.
// Stub implementation. Will wire to prometheus when available.
type Metrics struct {
	HitTotal        int64
	MissTotal       map[string]int64 // reason -> count
	EvictionsTotal  int64
	ActiveThreads   int64
	StartupTimeSec  float64
}

// NewMetrics creates a Metrics with initialized maps.
func NewMetrics() *Metrics {
	return &Metrics{
		MissTotal: map[string]int64{
			"cold_start":  0,
			"ttl_expired": 0,
			"deploy_gap":  0,
		},
	}
}

// NewManager creates a ConversationManager.
//
// summarizer may be nil (summaries will be skipped, degraded to window-only).
// fetcher may be nil (DORMANT threads cannot be resurrected, returns nil).
// emitter may be nil (metrics recording will be skipped).
func NewManager(config Config, summarizer Summarizer, fetcher SlackThreadFetcher, emitter ...MetricsEmitter) *Manager {
	var me MetricsEmitter
	if len(emitter) > 0 && emitter[0] != nil {
		me = emitter[0]
	}
	m := &Manager{
		threads:    make(map[string]*threadEntry),
		config:     config,
		summarizer: summarizer,
		fetcher:    fetcher,
		emitter:    me,
		metrics:    NewMetrics(),
		done:       make(chan struct{}),
		startedAt:  time.Now(),
	}

	// Emit startup timestamp metric.
	if m.emitter != nil {
		m.emitter.SetStartupTimestamp(m.startedAt)
	}

	// Start background cleanup goroutine.
	go m.cleanupLoop()

	return m
}

// GetThreadHistory returns the conversation context for a thread.
// Returns nil if no history exists (first message or expired).
// BC-08: Blocks up to 500ms during RESURRECTING state. Returns nil if unresolved.
// Never errors — returns nil on any failure (fail-open).
func (m *Manager) GetThreadHistory(ctx context.Context, threadTS string) *ThreadHistory {
	start := time.Now()

	m.mu.RLock()
	entry, exists := m.threads[threadTS]
	m.mu.RUnlock()

	if !exists {
		m.recordMiss("cold_start", start)
		return nil
	}

	switch entry.state {
	case ThreadCreated:
		// Thread exists but no messages yet.
		m.recordMiss("cold_start", start)
		return nil

	case ThreadActive:
		m.recordHit(start)
		return m.buildHistory(entry)

	case ThreadDormant:
		// Attempt resurrection via Slack API.
		return m.resurrectThread(ctx, threadTS, entry, start)

	case ThreadResurrecting:
		// BC-08: Another goroutine is already resurrecting. Wait up to timeout.
		return m.waitForResurrection(ctx, entry, start)

	default:
		return nil
	}
}

// StoreMessage appends a message to the thread's history.
// Creates the thread entry if it doesn't exist.
// channelID is required for resurrection (Slack API recovery from DORMANT state).
// Triggers summarization when message count exceeds the window size.
func (m *Manager) StoreMessage(ctx context.Context, threadTS string, channelID string, msg ThreadMessage) {
	m.mu.Lock()

	entry, exists := m.threads[threadTS]
	if !exists {
		entry = &threadEntry{
			state:      ThreadActive,
			lastAccess: time.Now(),
			channelID:  channelID,
		}
		m.threads[threadTS] = entry
	} else if entry.channelID == "" && channelID != "" {
		// Backfill channelID if it was not set during InitThread.
		entry.channelID = channelID
	}

	entry.messages = append(entry.messages, msg)
	entry.lastAccess = time.Now()

	// Transition CREATED -> ACTIVE on first message.
	if entry.state == ThreadCreated {
		entry.state = ThreadActive
	}

	shouldSummarize := len(entry.messages) > m.config.MaxRecentMessages && m.summarizer != nil
	messagesForSummary := make([]ThreadMessage, 0)
	if shouldSummarize {
		// Messages older than the window need summarization.
		prefixEnd := len(entry.messages) - m.config.MaxRecentMessages
		messagesForSummary = make([]ThreadMessage, prefixEnd)
		copy(messagesForSummary, entry.messages[:prefixEnd])
	}

	m.mu.Unlock()

	// Trigger async summarization if needed.
	if shouldSummarize && len(messagesForSummary) > 0 {
		go m.summarizePrefix(ctx, threadTS, messagesForSummary)
	}
}

// InitThread creates a CREATED state entry for a new thread.
// Called when assistant_thread_started fires.
func (m *Manager) InitThread(threadTS string, channelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.threads[threadTS]; !exists {
		m.threads[threadTS] = &threadEntry{
			state:      ThreadCreated,
			lastAccess: time.Now(),
			channelID:  channelID,
		}
	}
}

// Stop terminates the background cleanup goroutine.
// Must be called during server shutdown to prevent goroutine leaks.
func (m *Manager) Stop() {
	close(m.done)
}

// GetMetrics returns a snapshot of the current metrics.
func (m *Manager) GetMetrics() Metrics {
	m.metricsMu.Lock()
	missTotal := make(map[string]int64, len(m.metrics.MissTotal))
	for k, v := range m.metrics.MissTotal {
		missTotal[k] = v
	}
	snapshot := Metrics{
		HitTotal:       m.metrics.HitTotal,
		MissTotal:      missTotal,
		EvictionsTotal: m.metrics.EvictionsTotal,
		StartupTimeSec: m.metrics.StartupTimeSec,
	}
	m.metricsMu.Unlock()

	m.mu.RLock()
	snapshot.ActiveThreads = int64(len(m.threads))
	m.mu.RUnlock()

	return snapshot
}

// StartupTimestamp returns the time the manager was created.
// Powers deploy-gap detection in metrics.
func (m *Manager) StartupTimestamp() time.Time {
	return m.startedAt
}

// --- Internal methods ---

// resurrectThread transitions a DORMANT thread to RESURRECTING and attempts recovery.
func (m *Manager) resurrectThread(ctx context.Context, threadTS string, entry *threadEntry, start time.Time) *ThreadHistory {
	if m.fetcher == nil {
		m.recordMiss("ttl_expired", start)
		return nil
	}

	m.mu.Lock()
	// Double-check state under write lock (another goroutine might have resurrected).
	if entry.state == ThreadActive {
		m.mu.Unlock()
		m.recordHit(start)
		return m.buildHistory(entry)
	}
	if entry.state == ThreadResurrecting {
		m.mu.Unlock()
		return m.waitForResurrection(ctx, entry, start)
	}

	entry.state = ThreadResurrecting
	entry.resurrectCh = make(chan struct{})
	channelID := entry.channelID
	m.mu.Unlock()

	// Fetch from Slack API with a tight timeout.
	fetchCtx, cancel := context.WithTimeout(ctx, m.config.ResurrectingTimeout)
	defer cancel()

	messages, err := m.fetcher.FetchThreadMessages(fetchCtx, channelID, threadTS, 20)
	if err != nil {
		slog.Warn("thread resurrection failed",
			"thread_ts", threadTS,
			"error", err,
		)
		m.mu.Lock()
		entry.state = ThreadDormant
		entry.resurrectCh = nil
		m.mu.Unlock()
		m.recordMiss("ttl_expired", start)
		return nil
	}

	m.mu.Lock()
	entry.messages = messages
	entry.state = ThreadActive
	entry.lastAccess = time.Now()
	ch := entry.resurrectCh
	entry.resurrectCh = nil
	m.mu.Unlock()

	// Signal any waiters.
	if ch != nil {
		close(ch)
	}

	// Trigger async summarization if the resurrected thread is large.
	if len(messages) > m.config.MaxRecentMessages && m.summarizer != nil {
		prefixEnd := len(messages) - m.config.MaxRecentMessages
		go m.summarizePrefix(ctx, threadTS, messages[:prefixEnd])
	}

	m.recordHit(start)
	return m.buildHistory(entry)
}

// waitForResurrection blocks until resurrection completes or timeout.
// BC-08: Blocks max 500ms. Returns nil if unresolved.
func (m *Manager) waitForResurrection(ctx context.Context, entry *threadEntry, start time.Time) *ThreadHistory {
	m.mu.RLock()
	ch := entry.resurrectCh
	m.mu.RUnlock()

	if ch == nil {
		// Resurrection already completed or was never started.
		m.mu.RLock()
		if entry.state == ThreadActive {
			m.mu.RUnlock()
			m.recordHit(start)
			return m.buildHistory(entry)
		}
		m.mu.RUnlock()
		m.recordMiss("ttl_expired", start)
		return nil
	}

	timer := time.NewTimer(m.config.ResurrectingTimeout)
	defer timer.Stop()

	select {
	case <-ch:
		// Resurrection completed.
		m.mu.RLock()
		if entry.state == ThreadActive {
			m.mu.RUnlock()
			m.recordHit(start)
			return m.buildHistory(entry)
		}
		m.mu.RUnlock()
		m.recordMiss("ttl_expired", start)
		return nil

	case <-timer.C:
		// BC-08: Timeout exceeded. Return nil.
		slog.Debug("resurrection timeout exceeded",
			"timeout", m.config.ResurrectingTimeout,
		)
		m.recordMiss("ttl_expired", start)
		return nil

	case <-ctx.Done():
		m.recordMiss("ttl_expired", start)
		return nil
	}
}

// buildHistory constructs a ThreadHistory from a threadEntry.
//
// WS-2 (Gap B2): Uses user-turn windowing instead of raw message count windowing.
// Counting user turns backward ensures each kept user turn retains its paired
// assistant response at the window boundary. This prevents splitting a user message
// from its paired assistant response, which breaks conversational coherence.
//
// Algorithm: walk backward counting user-role messages. When the count exceeds
// MaxRecentMessages (used as user-turn limit), slice from the last seen user index.
// Minimum-message fallback: if the result would be empty (e.g., all-assistant thread),
// fall back to the last MaxRecentMessages messages.
func (m *Manager) buildHistory(entry *threadEntry) *ThreadHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := len(entry.messages)
	if total == 0 {
		return nil
	}

	// User-turn windowing: walk backward, count user messages.
	// MaxRecentMessages serves as the user-turn limit.
	limit := m.config.MaxRecentMessages
	windowStart := 0

	userCount := 0
	lastUserIdx := total
	for i := total - 1; i >= 0; i-- {
		if entry.messages[i].Role == "user" {
			userCount++
			if userCount > limit {
				// Exceeded limit: slice from lastUserIdx (the start of the
				// Nth-from-end user turn, which includes its assistant response).
				windowStart = lastUserIdx
				break
			}
			lastUserIdx = i
		}
	}

	// Minimum-message fallback: if no user messages found (all-assistant thread),
	// fall back to raw message count windowing to return something useful.
	if userCount == 0 && total > limit {
		windowStart = total - limit
	}

	recent := make([]ThreadMessage, total-windowStart)
	copy(recent, entry.messages[windowStart:])

	return &ThreadHistory{
		Summary:           entry.summary,
		RecentMessages:    recent,
		TotalMessageCount: total,
		State:             entry.state,
	}
}

// summarizePrefix generates a summary for messages older than the window.
func (m *Manager) summarizePrefix(ctx context.Context, threadTS string, messages []ThreadMessage) {
	if m.summarizer == nil || len(messages) == 0 {
		return
	}

	summary := m.summarizer.Summarize(ctx, messages)
	if summary == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.threads[threadTS]
	if !exists {
		return
	}
	entry.summary = summary
}

// cleanupLoop periodically evicts expired threads.
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.done:
			return
		}
	}
}

// cleanup evicts threads that have exceeded the TTL.
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-m.config.TTL)
	evicted := 0

	for ts, entry := range m.threads {
		if entry.lastAccess.Before(cutoff) {
			// Preserve channelID for potential resurrection.
			channelID := entry.channelID

			// Transition to DORMANT (do not delete — allows resurrection).
			m.threads[ts] = &threadEntry{
				state:      ThreadDormant,
				lastAccess: entry.lastAccess,
				channelID:  channelID,
			}
			evicted++
		}
	}

	// Clean up DORMANT entries that are very old (2x TTL).
	longCutoff := time.Now().Add(-2 * m.config.TTL)
	for ts, entry := range m.threads {
		if entry.state == ThreadDormant && entry.lastAccess.Before(longCutoff) {
			delete(m.threads, ts)
		}
	}

	if evicted > 0 {
		m.metricsMu.Lock()
		m.metrics.EvictionsTotal += int64(evicted)
		m.metricsMu.Unlock()

		if m.emitter != nil {
			for i := 0; i < evicted; i++ {
				m.emitter.IncrEvictions()
			}
			m.emitter.SetActiveThreads(len(m.threads))
		}

		slog.Debug("conversation cleanup",
			"evicted", evicted,
			"remaining", len(m.threads),
		)
	}
}

// recordHit increments the hit counter and emits the EMF metric.
func (m *Manager) recordHit(start time.Time) {
	m.metricsMu.Lock()
	m.metrics.HitTotal++
	m.metricsMu.Unlock()

	if m.emitter != nil {
		m.emitter.IncrConversationHit()
		m.emitter.RecordConversationGetLatency("hit", time.Since(start))
	}
}

// recordMiss increments the miss counter with the given reason and emits the EMF metric.
func (m *Manager) recordMiss(reason string, start time.Time) {
	m.metricsMu.Lock()
	m.metrics.MissTotal[reason]++
	m.metricsMu.Unlock()

	if m.emitter != nil {
		m.emitter.IncrConversationMiss(reason)
		m.emitter.RecordConversationGetLatency("miss", time.Since(start))
	}
}
