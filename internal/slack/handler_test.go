package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	slackapi "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/slack/streaming"
	"github.com/autom8y/knossos/internal/trust"
)

// --- mock types ---

// mockPipeline implements QueryRunner for tests.
type mockPipeline struct {
	mu       sync.Mutex
	response *response.ReasoningResponse
	err      error
	calls    []string
}

func (m *mockPipeline) Query(_ context.Context, question string) (*response.ReasoningResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, question)
	return m.response, m.err
}

func (m *mockPipeline) queryCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]string, len(m.calls))
	copy(dst, m.calls)
	return dst
}

// mockSlackClient records method calls for verification.
type mockSlackClient struct {
	mu             sync.Mutex
	sentBlocks     []sentBlocksCall
	statusCalls    []statusCall
	promptsCalls   []promptsCall
	titleCalls     []titleCall
	sendBlocksErr  error
	setStatusErr   error
	setPromptsErr  error
	setTitleErr    error
}

type sentBlocksCall struct {
	channelID string
	threadTS  string
}

type statusCall struct {
	channelID  string
	threadTS   string
	statusText string
}

type promptsCall struct {
	channelID string
	threadTS  string
	prompts   []string
}

type titleCall struct {
	channelID string
	threadTS  string
	title     string
}

func (m *mockSlackClient) recordSentBlocks(channelID, threadTS string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentBlocks = append(m.sentBlocks, sentBlocksCall{channelID, threadTS})
}

func (m *mockSlackClient) recordStatus(channelID, threadTS, statusText string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusCalls = append(m.statusCalls, statusCall{channelID, threadTS, statusText})
}

func (m *mockSlackClient) recordPrompts(channelID, threadTS string, prompts []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.promptsCalls = append(m.promptsCalls, promptsCall{channelID, threadTS, prompts})
}

func (m *mockSlackClient) recordTitle(channelID, threadTS, title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.titleCalls = append(m.titleCalls, titleCall{channelID, threadTS, title})
}

func (m *mockSlackClient) getSentBlocks() []sentBlocksCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]sentBlocksCall, len(m.sentBlocks))
	copy(dst, m.sentBlocks)
	return dst
}

func (m *mockSlackClient) getStatusCalls() []statusCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]statusCall, len(m.statusCalls))
	copy(dst, m.statusCalls)
	return dst
}

func (m *mockSlackClient) getPromptsCalls() []promptsCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]promptsCall, len(m.promptsCalls))
	copy(dst, m.promptsCalls)
	return dst
}

func (m *mockSlackClient) getTitleCalls() []titleCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]titleCall, len(m.titleCalls))
	copy(dst, m.titleCalls)
	return dst
}

// testableSlackClient creates a SlackClient with methods overridden via a mock transport.
// Since SlackClient methods call Slack APIs directly, we use a different approach:
// We create a testable handler that captures calls.
type testableHandler struct {
	pipeline *mockPipeline
	mock     *mockSlackClient
}

func newTestableHandler(t *testing.T) *testableHandler {
	t.Helper()
	return &testableHandler{
		pipeline: &mockPipeline{
			response: &response.ReasoningResponse{
				Answer: "Test answer",
				Tier:   trust.TierHigh,
			},
		},
		mock: &mockSlackClient{},
	}
}

// --- Tests ---

func TestHandler_URLVerificationChallenge(t *testing.T) {
	handler, _, _ := NewSlackHandler(nil, nil, DefaultSlackConfig(), nil)

	body := `{"type":"url_verification","challenge":"test-challenge-value"}`
	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Challenge string `json:"challenge"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "test-challenge-value", resp.Challenge)
}

func TestHandler_EventCallbackBotFiltered(t *testing.T) {
	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "Should not be called",
			Tier:   trust.TierHigh,
		},
	}
	client := NewSlackClientWithAPI(&noopSlackAPI{}, "xoxb-test")
	handler, _, _ := NewSlackHandler(pipeline, client, DefaultSlackConfig(), nil)

	event := MessageEvent{
		Type:    "message",
		Text:    "hello from a bot",
		BotID:   "B12345",
		Channel: "C001",
		TS:      "1234567890.123456",
	}
	eventJSON, _ := json.Marshal(event)
	envelope := EventEnvelope{
		Type:    "event_callback",
		Event:   eventJSON,
		TeamID:  "T001",
		EventID: "Ev001",
	}
	body, _ := json.Marshal(envelope)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(string(body)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Pipeline should NOT have been called
	assert.Empty(t, pipeline.queryCalls(), "pipeline should not be invoked for bot messages")
}

func TestHandler_EventCallbackMessageSubtype(t *testing.T) {
	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "Should not be called",
			Tier:   trust.TierHigh,
		},
	}
	client := NewSlackClientWithAPI(&noopSlackAPI{}, "xoxb-test")
	handler, _, _ := NewSlackHandler(pipeline, client, DefaultSlackConfig(), nil)

	event := MessageEvent{
		Type:    "message",
		SubType: "message_changed",
		Text:    "edited message",
		User:    "U001",
		Channel: "C001",
		TS:      "1234567890.123456",
	}
	body := makeEnvelope(t, event)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, pipeline.queryCalls(), "pipeline should not be invoked for message subtypes")
}

func TestHandler_EventCallbackUserMessage(t *testing.T) {
	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "The architecture follows a 3-tier model.",
			Tier:   trust.TierHigh,
		},
	}
	client := NewSlackClientWithAPI(&recordingSlackAPI{}, "xoxb-test")
	handler, _, _ := NewSlackHandler(pipeline, client, DefaultSlackConfig(), nil)

	event := MessageEvent{
		Type:    "message",
		Text:    "What is the architecture?",
		User:    "U001",
		Channel: "C001",
		TS:      "1234567890.123456",
	}
	body := makeEnvelope(t, event)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should respond 200 immediately
	assert.Equal(t, http.StatusOK, w.Code)

	// Give the goroutine time to complete
	waitForCalls(t, pipeline, 1, 2*time.Second)

	calls := pipeline.queryCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "What is the architecture?", calls[0])
}

func TestHandler_NonEventCallback(t *testing.T) {
	handler, _, _ := NewSlackHandler(nil, nil, DefaultSlackConfig(), nil)

	body := `{"type":"app_rate_limited","team_id":"T001"}`
	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_AssistantThreadStarted(t *testing.T) {
	client := NewSlackClientWithAPI(&noopSlackAPI{}, "xoxb-test")
	handler, _, _ := NewSlackHandler(nil, client, DefaultSlackConfig(), nil)

	event := AssistantThreadEvent{
		Type: "assistant_thread_started",
		AssistantThread: AssistantThreadInfo{
			ChannelID: "C001",
			ThreadTS:  "1234567890.123456",
		},
	}
	eventJSON, _ := json.Marshal(event)
	envelope := EventEnvelope{
		Type:    "event_callback",
		Event:   eventJSON,
		TeamID:  "T001",
		EventID: "Ev001",
	}
	body, _ := json.Marshal(envelope)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(string(body)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_EmptyMessage(t *testing.T) {
	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "Should not be called",
			Tier:   trust.TierHigh,
		},
	}
	client := NewSlackClientWithAPI(&noopSlackAPI{}, "xoxb-test")
	handler, _, _ := NewSlackHandler(pipeline, client, DefaultSlackConfig(), nil)

	event := MessageEvent{
		Type:    "message",
		Text:    "",
		User:    "U001",
		Channel: "C001",
		TS:      "1234567890.123456",
	}
	body := makeEnvelope(t, event)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Brief wait to ensure no goroutine spawned
	time.Sleep(50 * time.Millisecond)
	assert.Empty(t, pipeline.queryCalls(), "pipeline should not be invoked for empty messages")
}

func TestHandler_InvalidJSON(t *testing.T) {
	handler, _, _ := NewSlackHandler(nil, nil, DefaultSlackConfig(), nil)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- test helpers ---

// noopSlackAPI is a no-op implementation of SlackAPI for tests.
type noopSlackAPI struct{}

func (n *noopSlackAPI) PostMessage(channelID string, options ...slackapi.MsgOption) (string, string, error) {
	return "", "", nil
}

func (n *noopSlackAPI) UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error) {
	return "", "", "", nil
}

func (n *noopSlackAPI) AuthTest() (*slackapi.AuthTestResponse, error) {
	return &slackapi.AuthTestResponse{}, nil
}

// recordingSlackAPI records PostMessage calls.
type recordingSlackAPI struct {
	mu    sync.Mutex
	calls []string
}

func (r *recordingSlackAPI) PostMessage(channelID string, options ...slackapi.MsgOption) (string, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, channelID)
	return "", "", nil
}

func (r *recordingSlackAPI) UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error) {
	return "", "", "", nil
}

func (r *recordingSlackAPI) AuthTest() (*slackapi.AuthTestResponse, error) {
	return &slackapi.AuthTestResponse{}, nil
}

func makeEnvelope(t *testing.T, event interface{}) string {
	t.Helper()
	eventJSON, err := json.Marshal(event)
	require.NoError(t, err)
	envelope := EventEnvelope{
		Type:    "event_callback",
		Event:   eventJSON,
		TeamID:  "T001",
		EventID: "Ev001",
	}
	body, err := json.Marshal(envelope)
	require.NoError(t, err)
	return string(body)
}

func waitForCalls(t *testing.T, pipeline *mockPipeline, expected int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(pipeline.queryCalls()) >= expected {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d pipeline calls, got %d", expected, len(pipeline.queryCalls()))
}

// --- Streaming mock types ---

// mockStreamingRunner implements StreamingQueryRunner for tests.
type mockStreamingRunner struct {
	mu       sync.Mutex
	calls    []string
	chunks   []string // Chunks to emit via onChunk callback.
	response *response.ReasoningResponse
	err      error
}

func (m *mockStreamingRunner) QueryStream(_ context.Context, triageInput *TriageResultInputData, onChunk func(chunk string)) (*response.ReasoningResponse, error) {
	m.mu.Lock()
	query := ""
	if triageInput != nil {
		query = triageInput.RefinedQuery
	}
	m.calls = append(m.calls, query)
	chunks := make([]string, len(m.chunks))
	copy(chunks, m.chunks)
	resp := m.response
	err := m.err
	m.mu.Unlock()

	// Emit test chunks via callback.
	for _, chunk := range chunks {
		onChunk(chunk)
	}

	return resp, err
}

func (m *mockStreamingRunner) queryStreamCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]string, len(m.calls))
	copy(dst, m.calls)
	return dst
}

// mockTriageRunner implements TriageRunner for tests.
type mockTriageRunner struct {
	mu     sync.Mutex
	calls  []string
	result *TriageResultData
	err    error
}

func (m *mockTriageRunner) Assess(_ context.Context, query string, _ []TriageThreadMessage, _ ...TriageAssessOptions) (*TriageResultData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, query)
	return m.result, m.err
}

func (m *mockTriageRunner) assessCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]string, len(m.calls))
	copy(dst, m.calls)
	return dst
}

// streamingSlackServer creates a mock Slack API server that records streaming API calls.
type streamingSlackServer struct {
	mu     sync.Mutex
	calls  []streamAPICall
	server *httptest.Server
}

type streamAPICall struct {
	method string
	body   map[string]interface{}
}

func newStreamingSlackServer(t *testing.T) *streamingSlackServer {
	t.Helper()
	m := &streamingSlackServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/api/")
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		m.mu.Lock()
		m.calls = append(m.calls, streamAPICall{method: method, body: body})
		m.mu.Unlock()

		// Return appropriate responses for each Slack API method.
		switch method {
		case "chat.startStream":
			streamID := fmt.Sprintf("stream-%d", time.Now().UnixNano())
			_, _ = fmt.Fprintf(w, `{"ok":true,"stream_id":"%s"}`, streamID)
		case "chat.appendStream":
			_, _ = fmt.Fprint(w, `{"ok":true}`)
		case "chat.stopStream":
			_, _ = fmt.Fprint(w, `{"ok":true}`)
		case "chat.postMessage":
			_, _ = fmt.Fprint(w, `{"ok":true,"ts":"1234567890.999999"}`)
		case "chat.update":
			_, _ = fmt.Fprint(w, `{"ok":true}`)
		default:
			_, _ = fmt.Fprint(w, `{"ok":true}`)
		}
	})

	m.server = httptest.NewServer(mux)
	return m
}

func (m *streamingSlackServer) getCalls() []streamAPICall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]streamAPICall, len(m.calls))
	copy(dst, m.calls)
	return dst
}

func (m *streamingSlackServer) getCallMethods() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	methods := make([]string, len(m.calls))
	for i, c := range m.calls {
		methods[i] = c.method
	}
	return methods
}

// waitForStreamCalls polls until the streaming runner has at least `expected` calls.
func waitForStreamCalls(t *testing.T, runner *mockStreamingRunner, expected int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(runner.queryStreamCalls()) >= expected {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d streaming calls, got %d", expected, len(runner.queryStreamCalls()))
}

// --- Streaming tests ---

func TestHandler_StreamingPath(t *testing.T) {
	// Set up streaming infrastructure.
	mockServer := newStreamingSlackServer(t)
	defer mockServer.server.Close()

	streamSender := streaming.NewSenderForTest("xoxb-test", "T001", mockServer.server.URL+"/api/")

	streamRunner := &mockStreamingRunner{
		chunks: []string{"Hello ", "World!"},
		response: &response.ReasoningResponse{
			Answer: "Hello World!",
			Tier:   trust.TierHigh,
		},
	}

	triageRunner := &mockTriageRunner{
		result: &TriageResultData{
			RefinedQuery: "refined test question",
			Candidates: []TriageCandidateData{
				{QualifiedName: "test::domain", RelevanceScore: 0.9},
			},
			ModelCallCount: 1,
		},
	}

	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "Sync fallback - should not be used",
			Tier:   trust.TierHigh,
		},
	}

	client := NewSlackClientWithAPI(&recordingSlackAPI{}, "xoxb-test")
	cfg := DefaultSlackConfig()
	cfg.StreamingEnabled = true

	handler, _, stop := NewSlackHandlerWithDeps(HandlerDeps{
		Pipeline:        pipeline,
		Client:          client,
		Config:          cfg,
		TriageRunner:    triageRunner,
		StreamingRunner: streamRunner,
		StreamSender:    streamSender,
	})
	defer stop()

	// Send a user message.
	event := MessageEvent{
		Type:    "message",
		Text:    "What is streaming?",
		User:    "U001",
		Channel: "C001",
		TS:      "1234567890.123456",
	}
	body := makeEnvelope(t, event)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Wait for the streaming pipeline to be called.
	waitForStreamCalls(t, streamRunner, 1, 3*time.Second)

	// Verify StreamingRunner.QueryStream was called with the refined query.
	streamCalls := streamRunner.queryStreamCalls()
	require.Len(t, streamCalls, 1)
	assert.Equal(t, "refined test question", streamCalls[0])

	// Verify triage was called.
	triageCalls := triageRunner.assessCalls()
	require.Len(t, triageCalls, 1)
	assert.Equal(t, "What is streaming?", triageCalls[0])

	// Verify sync pipeline was NOT called (streaming path was taken).
	assert.Empty(t, pipeline.queryCalls(), "sync pipeline should not be called when streaming path is taken")

	// Verify streaming API calls: startStream, appendStream (x2 chunks), stopStream.
	// Poll for stopStream — it runs in a deferred cleanup after the goroutine completes.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		methods := mockServer.getCallMethods()
		hasStop := false
		for _, m := range methods {
			if m == "chat.stopStream" {
				hasStop = true
				break
			}
		}
		if hasStop {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	methods := mockServer.getCallMethods()
	assert.Contains(t, methods, "chat.startStream", "should call startStream")
	assert.Contains(t, methods, "chat.appendStream", "should call appendStream for chunks")
	assert.Contains(t, methods, "chat.stopStream", "should call stopStream")
}

func TestHandler_StreamingFallbackWhenDisabled(t *testing.T) {
	// Set up streaming infrastructure but disable streaming.
	mockServer := newStreamingSlackServer(t)
	defer mockServer.server.Close()

	streamSender := streaming.NewSenderForTest("xoxb-test", "T001", mockServer.server.URL+"/api/")

	streamRunner := &mockStreamingRunner{
		chunks: []string{"Should not be called"},
		response: &response.ReasoningResponse{
			Answer: "Streaming response",
			Tier:   trust.TierHigh,
		},
	}

	triageRunner := &mockTriageRunner{
		result: &TriageResultData{
			RefinedQuery: "refined",
			Candidates: []TriageCandidateData{
				{QualifiedName: "test::domain", RelevanceScore: 0.9},
			},
			ModelCallCount: 1,
		},
	}

	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "Sync pipeline response",
			Tier:   trust.TierHigh,
		},
	}

	client := NewSlackClientWithAPI(&recordingSlackAPI{}, "xoxb-test")
	cfg := DefaultSlackConfig()
	cfg.StreamingEnabled = false // Disable streaming.

	handler, _, stop := NewSlackHandlerWithDeps(HandlerDeps{
		Pipeline:        pipeline,
		Client:          client,
		Config:          cfg,
		TriageRunner:    triageRunner,
		StreamingRunner: streamRunner,
		StreamSender:    streamSender,
	})
	defer stop()

	event := MessageEvent{
		Type:    "message",
		Text:    "What is the architecture?",
		User:    "U001",
		Channel: "C001",
		TS:      "1234567890.222222",
	}
	body := makeEnvelope(t, event)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Wait for sync pipeline to be called (triage path uses it via processWithTriage).
	// The triage runner returns candidates, so processWithTriage calls Pipeline.Query
	// (since TriagePipeline is nil).
	waitForCalls(t, pipeline, 1, 3*time.Second)

	// Verify streaming runner was NOT called.
	assert.Empty(t, streamRunner.queryStreamCalls(), "streaming runner should not be called when disabled")

	// Verify streaming API was NOT called.
	assert.Empty(t, mockServer.getCalls(), "streaming API should not be called when disabled")
}

func TestHandler_StreamingFallbackWhenRunnerNil(t *testing.T) {
	// Streaming enabled but StreamingRunner is nil — should fall back to sync.
	pipeline := &mockPipeline{
		response: &response.ReasoningResponse{
			Answer: "Sync pipeline response",
			Tier:   trust.TierHigh,
		},
	}

	triageRunner := &mockTriageRunner{
		result: &TriageResultData{
			RefinedQuery: "refined",
			Candidates: []TriageCandidateData{
				{QualifiedName: "test::domain", RelevanceScore: 0.9},
			},
			ModelCallCount: 1,
		},
	}

	client := NewSlackClientWithAPI(&recordingSlackAPI{}, "xoxb-test")
	cfg := DefaultSlackConfig()
	cfg.StreamingEnabled = true

	handler, _, stop := NewSlackHandlerWithDeps(HandlerDeps{
		Pipeline:        pipeline,
		Client:          client,
		Config:          cfg,
		TriageRunner:    triageRunner,
		StreamingRunner: nil, // Nil streaming runner.
		StreamSender:    nil, // Nil stream sender.
	})
	defer stop()

	event := MessageEvent{
		Type:    "message",
		Text:    "What is the architecture?",
		User:    "U001",
		Channel: "C001",
		TS:      "1234567890.333333",
	}
	body := makeEnvelope(t, event)

	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Wait for sync pipeline to be called.
	waitForCalls(t, pipeline, 1, 3*time.Second)

	// Verify sync pipeline was used.
	calls := pipeline.queryCalls()
	require.Len(t, calls, 1)
}
