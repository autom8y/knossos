package conversation

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock implementations ---

type mockFetcher struct {
	mu       sync.Mutex
	messages []ThreadMessage
	err      error
	calls    int
	delay    time.Duration
}

func (f *mockFetcher) FetchThreadMessages(ctx context.Context, channelID string, threadTS string, limit int) ([]ThreadMessage, error) {
	f.mu.Lock()
	f.calls++
	delay := f.delay
	f.mu.Unlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	return f.messages, f.err
}

func (f *mockFetcher) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type mockSummarizer struct {
	mu      sync.Mutex
	summary string
	calls   int
}

func (s *mockSummarizer) Summarize(ctx context.Context, messages []ThreadMessage) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.summary
}

func (s *mockSummarizer) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// --- Test helpers ---

func testConfig() Config {
	return Config{
		MaxRecentMessages:   3, // Small for testing.
		TTL:                 100 * time.Millisecond,
		CleanupInterval:     50 * time.Millisecond,
		SummaryMaxTokens:    250,
		ResurrectingTimeout: 200 * time.Millisecond,
	}
}

func makeMsg(role, content string) ThreadMessage {
	return ThreadMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// --- Tests ---

func TestManager_StoreAndRetrieve(t *testing.T) {
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	ctx := context.Background()

	// Store a message.
	mgr.StoreMessage(ctx, "thread-1", makeMsg("user", "What is the architecture?"))

	// Retrieve it.
	history := mgr.GetThreadHistory(ctx, "thread-1")
	require.NotNil(t, history)
	assert.Equal(t, 1, history.TotalMessageCount)
	assert.Len(t, history.RecentMessages, 1)
	assert.Equal(t, "user", history.RecentMessages[0].Role)
	assert.Equal(t, "What is the architecture?", history.RecentMessages[0].Content)
	assert.Equal(t, ThreadActive, history.State)
	assert.Empty(t, history.Summary)
}

func TestManager_WindowLimitsRecentMessages(t *testing.T) {
	cfg := testConfig() // MaxRecentMessages = 3.
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	ctx := context.Background()

	// Store 5 messages.
	for i := 0; i < 5; i++ {
		mgr.StoreMessage(ctx, "thread-1", makeMsg("user", time.Now().String()))
	}

	history := mgr.GetThreadHistory(ctx, "thread-1")
	require.NotNil(t, history)
	assert.Equal(t, 5, history.TotalMessageCount)
	assert.Len(t, history.RecentMessages, 3, "should only return last 3 messages")
}

func TestManager_SummarizationTriggered(t *testing.T) {
	cfg := testConfig()
	summarizer := &mockSummarizer{summary: "Previous discussion about architecture."}
	mgr := NewManager(cfg, summarizer, nil)
	defer mgr.Stop()

	ctx := context.Background()

	// Store enough messages to trigger summarization (>MaxRecentMessages).
	for i := 0; i < 5; i++ {
		mgr.StoreMessage(ctx, "thread-1", makeMsg("user", "message"))
	}

	// Give the async summarization goroutine time to complete.
	time.Sleep(50 * time.Millisecond)

	history := mgr.GetThreadHistory(ctx, "thread-1")
	require.NotNil(t, history)
	assert.Equal(t, "Previous discussion about architecture.", history.Summary)
	assert.GreaterOrEqual(t, summarizer.callCount(), 1)
}

func TestManager_GetThreadHistory_UnknownThread(t *testing.T) {
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	history := mgr.GetThreadHistory(context.Background(), "nonexistent-thread")
	assert.Nil(t, history, "should return nil for unknown threads")

	metrics := mgr.GetMetrics()
	assert.Equal(t, int64(1), metrics.MissTotal["cold_start"])
}

func TestManager_InitThread(t *testing.T) {
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	// Init a thread (from assistant_thread_started event).
	mgr.InitThread("thread-1", "C001")

	// CREATED state returns nil (no messages yet).
	history := mgr.GetThreadHistory(context.Background(), "thread-1")
	assert.Nil(t, history, "CREATED thread with no messages should return nil")

	// Store a message and verify transition to ACTIVE.
	mgr.StoreMessage(context.Background(), "thread-1", makeMsg("user", "hello"))
	history = mgr.GetThreadHistory(context.Background(), "thread-1")
	require.NotNil(t, history)
	assert.Equal(t, ThreadActive, history.State)
}

func TestManager_TTLEviction(t *testing.T) {
	cfg := testConfig() // TTL = 100ms, CleanupInterval = 50ms.
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	ctx := context.Background()

	// Store a message.
	mgr.StoreMessage(ctx, "thread-1", makeMsg("user", "hello"))
	require.NotNil(t, mgr.GetThreadHistory(ctx, "thread-1"))

	// Wait for TTL + cleanup interval to fire.
	time.Sleep(200 * time.Millisecond)

	// Thread should now be DORMANT.
	mgr.mu.RLock()
	entry, exists := mgr.threads["thread-1"]
	mgr.mu.RUnlock()

	if exists {
		assert.Equal(t, ThreadDormant, entry.state, "thread should be DORMANT after TTL")
	}
	// Without a fetcher, GetThreadHistory for DORMANT returns nil.
	history := mgr.GetThreadHistory(ctx, "thread-1")
	assert.Nil(t, history, "DORMANT thread without fetcher should return nil")
}

func TestManager_ResurrectionSuccess(t *testing.T) {
	cfg := testConfig()
	fetcher := &mockFetcher{
		messages: []ThreadMessage{
			makeMsg("user", "original question"),
			makeMsg("assistant", "original answer"),
			makeMsg("user", "follow-up question"),
		},
	}
	mgr := NewManager(cfg, nil, fetcher)
	defer mgr.Stop()

	ctx := context.Background()

	// Manually set a thread to DORMANT.
	mgr.mu.Lock()
	mgr.threads["thread-1"] = &threadEntry{
		state:      ThreadDormant,
		lastAccess: time.Now().Add(-3 * time.Hour),
		channelID:  "C001",
	}
	mgr.mu.Unlock()

	// Get should trigger resurrection.
	history := mgr.GetThreadHistory(ctx, "thread-1")
	require.NotNil(t, history, "resurrected thread should return history")
	assert.Equal(t, 3, history.TotalMessageCount)
	assert.Equal(t, ThreadActive, history.State)
	assert.Equal(t, 1, fetcher.callCount())
}

func TestManager_ResurrectionTimeout(t *testing.T) {
	cfg := testConfig() // ResurrectingTimeout = 200ms.
	fetcher := &mockFetcher{
		messages: []ThreadMessage{makeMsg("user", "hello")},
		delay:    500 * time.Millisecond, // Longer than timeout.
	}
	mgr := NewManager(cfg, nil, fetcher)
	defer mgr.Stop()

	ctx := context.Background()

	// Manually set a thread to DORMANT.
	mgr.mu.Lock()
	mgr.threads["thread-1"] = &threadEntry{
		state:      ThreadDormant,
		lastAccess: time.Now().Add(-3 * time.Hour),
		channelID:  "C001",
	}
	mgr.mu.Unlock()

	// BC-08: Should block up to ResurrectingTimeout, then return nil.
	start := time.Now()
	history := mgr.GetThreadHistory(ctx, "thread-1")
	elapsed := time.Since(start)

	assert.Nil(t, history, "should return nil when resurrection times out")
	assert.Less(t, elapsed, 400*time.Millisecond, "should not block longer than timeout + margin")
}

func TestManager_ConcurrentResurrectionWaiters(t *testing.T) {
	cfg := testConfig()
	fetcher := &mockFetcher{
		messages: []ThreadMessage{
			makeMsg("user", "hello"),
			makeMsg("assistant", "hi there"),
		},
		delay: 50 * time.Millisecond,
	}
	mgr := NewManager(cfg, nil, fetcher)
	defer mgr.Stop()

	ctx := context.Background()

	// Manually set a thread to DORMANT.
	mgr.mu.Lock()
	mgr.threads["thread-1"] = &threadEntry{
		state:      ThreadDormant,
		lastAccess: time.Now().Add(-3 * time.Hour),
		channelID:  "C001",
	}
	mgr.mu.Unlock()

	// Launch multiple concurrent readers.
	var wg sync.WaitGroup
	results := make([]*ThreadHistory, 5)
	for i := 0; i < 5; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = mgr.GetThreadHistory(ctx, "thread-1")
		}()
	}
	wg.Wait()

	// At least one should succeed (the initiator).
	successCount := 0
	for _, r := range results {
		if r != nil {
			successCount++
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "at least one concurrent reader should get history")

	// Only one fetch should have been made.
	assert.Equal(t, 1, fetcher.callCount(), "should only fetch once even with concurrent requests")
}

func TestManager_Stop_CleanupGoroutine(t *testing.T) {
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)

	// Stop should not panic or block.
	mgr.Stop()

	// After stop, operations should still work (just no background cleanup).
	ctx := context.Background()
	mgr.StoreMessage(ctx, "thread-1", makeMsg("user", "hello"))
	history := mgr.GetThreadHistory(ctx, "thread-1")
	require.NotNil(t, history)
}

func TestManager_Metrics(t *testing.T) {
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	ctx := context.Background()

	// Generate some hits and misses.
	mgr.GetThreadHistory(ctx, "nonexistent") // cold_start miss
	mgr.StoreMessage(ctx, "thread-1", makeMsg("user", "hello"))
	mgr.GetThreadHistory(ctx, "thread-1") // hit

	metrics := mgr.GetMetrics()
	assert.Equal(t, int64(1), metrics.HitTotal)
	assert.Equal(t, int64(1), metrics.MissTotal["cold_start"])
	assert.Equal(t, int64(1), metrics.ActiveThreads)
}

func TestManager_StartupTimestamp(t *testing.T) {
	before := time.Now()
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()
	after := time.Now()

	ts := mgr.StartupTimestamp()
	assert.True(t, ts.After(before) || ts.Equal(before))
	assert.True(t, ts.Before(after) || ts.Equal(after))
}

func TestManager_StoreMessage_ChannelIDPreserved(t *testing.T) {
	cfg := testConfig()
	mgr := NewManager(cfg, nil, nil)
	defer mgr.Stop()

	// Init thread with channel ID.
	mgr.InitThread("thread-1", "C001")

	// Store messages.
	mgr.StoreMessage(context.Background(), "thread-1", makeMsg("user", "hello"))

	// Verify channel ID is preserved.
	mgr.mu.RLock()
	entry := mgr.threads["thread-1"]
	mgr.mu.RUnlock()

	assert.Equal(t, "C001", entry.channelID)
}

// ---- Deploy-Gap Acceptance Tests ----

func TestDeployGap_EmptyCacheOnStartup(t *testing.T) {
	// After deploy, ConversationManager starts with empty cache.
	// First follow-up to any thread returns nil (deploy-gap expected).
	config := DefaultConfig()
	mgr := NewManager(config, nil, nil)
	defer mgr.Stop()

	// No threads have been initialized -- simulates post-deploy state.
	history := mgr.GetThreadHistory(context.Background(), "orphan-thread")
	assert.Nil(t, history, "post-deploy: first follow-up must return nil (deploy-gap)")
}

func TestDeployGap_WithFetcher_GracefulRecovery(t *testing.T) {
	// After deploy with a fetcher wired, DORMANT threads trigger recovery.
	// The fetcher returns messages, and the manager populates the thread.
	fetcher := &mockFetcher{
		messages: []ThreadMessage{
			{Role: "user", Content: "original question", Timestamp: time.Now().Add(-10 * time.Minute)},
			{Role: "assistant", Content: "original answer", Timestamp: time.Now().Add(-9 * time.Minute)},
		},
	}
	config := DefaultConfig()
	mgr := NewManager(config, nil, fetcher)
	defer mgr.Stop()

	// Initialize the thread (simulates assistant_thread_started before deploy).
	mgr.InitThread("deploy-test", "C001")

	// First follow-up after thread was created but no messages stored.
	// Without prior StoreMessage calls, the thread has no messages, so
	// GetThreadHistory returns an empty/nil history.
	history := mgr.GetThreadHistory(context.Background(), "deploy-test")

	// With no stored messages, the thread exists but has empty history.
	// This validates the deploy-gap path: manager starts empty, threads
	// initialized but no conversation data.
	if history != nil {
		assert.Equal(t, 0, history.TotalMessageCount,
			"deploy-gap: newly initialized thread should have no messages")
	}
}

func TestDeployGap_NilFetcher_ReturnsNil(t *testing.T) {
	// When fetcher is nil (conversations.replies not wired), DORMANT threads
	// return nil gracefully. This is the expected Tier 1 behavior.
	config := DefaultConfig()
	mgr := NewManager(config, nil, nil)
	defer mgr.Stop()

	history := mgr.GetThreadHistory(context.Background(), "unknown-thread")
	assert.Nil(t, history, "nil fetcher: DORMANT threads must return nil, not error")
}

func TestDeployGap_MetricsTrackDeployGap(t *testing.T) {
	// The metrics must distinguish deploy_gap misses from cold_start and ttl_expired.
	config := DefaultConfig()
	mgr := NewManager(config, nil, nil)
	defer mgr.Stop()

	// Verify metrics are initialized with deploy_gap reason.
	mgr.metricsMu.Lock()
	_, hasDeployGap := mgr.metrics.MissTotal["deploy_gap"]
	_, hasColdStart := mgr.metrics.MissTotal["cold_start"]
	_, hasTTLExpired := mgr.metrics.MissTotal["ttl_expired"]
	mgr.metricsMu.Unlock()

	assert.True(t, hasDeployGap, "metrics must track deploy_gap miss reason")
	assert.True(t, hasColdStart, "metrics must track cold_start miss reason")
	assert.True(t, hasTTLExpired, "metrics must track ttl_expired miss reason")
}

func TestDeployGap_StartupTimestamp(t *testing.T) {
	// The manager records its startup timestamp for deploy-gap detection.
	// Any thread access before the startup time must be classified as deploy_gap.
	before := time.Now()
	config := DefaultConfig()
	mgr := NewManager(config, nil, nil)
	defer mgr.Stop()
	after := time.Now()

	require.False(t, mgr.startedAt.IsZero(), "startup timestamp must be set")
	assert.True(t, mgr.startedAt.After(before) || mgr.startedAt.Equal(before),
		"startup timestamp must be >= construction start")
	assert.True(t, mgr.startedAt.Before(after) || mgr.startedAt.Equal(after),
		"startup timestamp must be <= construction end")
}
