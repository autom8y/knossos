package streaming

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSlackServer creates an httptest server that simulates Slack API responses.
// It records all API calls and can be configured to return errors for specific methods.
type mockSlackServer struct {
	mu     sync.Mutex
	calls  []apiCall
	server *httptest.Server

	// Per-method error responses.
	errors map[string]string
}

type apiCall struct {
	method  string
	payload map[string]any
}

func newMockSlackServer(t *testing.T) *mockSlackServer {
	t.Helper()
	m := &mockSlackServer{
		errors: make(map[string]string),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Path[len("/api/"):]

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			payload = map[string]any{}
		}

		m.mu.Lock()
		m.calls = append(m.calls, apiCall{method: method, payload: payload})
		errMsg := m.errors[method]
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if errMsg != "" {
			resp := map[string]any{"ok": false, "error": errMsg}
			json.NewEncoder(w).Encode(resp)
			return
		}

		switch method {
		case "chat.startStream":
			resp := map[string]any{"ok": true, "stream_id": "stream-123"}
			json.NewEncoder(w).Encode(resp)
		case "chat.appendStream":
			resp := map[string]any{"ok": true}
			json.NewEncoder(w).Encode(resp)
		case "chat.stopStream":
			resp := map[string]any{"ok": true}
			json.NewEncoder(w).Encode(resp)
		case "chat.postMessage":
			resp := map[string]any{"ok": true, "ts": "1234567890.000001"}
			json.NewEncoder(w).Encode(resp)
		case "chat.update":
			resp := map[string]any{"ok": true}
			json.NewEncoder(w).Encode(resp)
		default:
			resp := map[string]any{"ok": false, "error": "unknown_method"}
			json.NewEncoder(w).Encode(resp)
		}
	})

	m.server = httptest.NewServer(mux)
	return m
}

func (m *mockSlackServer) getCalls() []apiCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]apiCall, len(m.calls))
	copy(dst, m.calls)
	return dst
}

func (m *mockSlackServer) setError(method, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = errMsg
}

// senderWithMockServer creates a Sender whose API calls go to the mock server.
// This overrides the base URL used by slackAPICall.
func senderWithMockServer(t *testing.T, mock *mockSlackServer) *Sender {
	t.Helper()
	s := NewSender("xoxb-test-token", "T001")
	// Override the API call method to use mock server.
	s.apiBaseURL = mock.server.URL + "/api/"
	return s
}

func TestSender_NativeStreaming_HappyPath(t *testing.T) {
	mock := newMockSlackServer(t)
	defer mock.server.Close()

	sender := senderWithMockServer(t, mock)
	ctx := context.Background()

	// Start stream.
	streamID, err := sender.StartStream(ctx, "C001", "1234567890.000000")
	require.NoError(t, err)
	assert.NotEmpty(t, streamID)

	// Append chunks.
	err = sender.AppendStream(ctx, streamID, "Hello ")
	require.NoError(t, err)
	err = sender.AppendStream(ctx, streamID, "World!")
	require.NoError(t, err)

	// Stop stream.
	err = sender.StopStream(ctx, streamID)
	require.NoError(t, err)

	// Verify API calls.
	calls := mock.getCalls()
	require.Len(t, calls, 4)
	assert.Equal(t, "chat.startStream", calls[0].method)
	assert.Equal(t, "chat.appendStream", calls[1].method)
	assert.Equal(t, "chat.appendStream", calls[2].method)
	assert.Equal(t, "chat.stopStream", calls[3].method)
}

func TestSender_FallbackToEditBased(t *testing.T) {
	mock := newMockSlackServer(t)
	defer mock.server.Close()

	// Make native streaming fail.
	mock.setError("chat.startStream", "not_allowed")

	sender := senderWithMockServer(t, mock)
	ctx := context.Background()

	// Start stream (should fall back to edit-based).
	streamID, err := sender.StartStream(ctx, "C001", "1234567890.000000")
	require.NoError(t, err)
	assert.Contains(t, streamID, "edit-")

	// Append chunks.
	err = sender.AppendStream(ctx, streamID, "Hello ")
	require.NoError(t, err)
	err = sender.AppendStream(ctx, streamID, "World!")
	require.NoError(t, err)

	// Stop stream.
	err = sender.StopStream(ctx, streamID)
	require.NoError(t, err)

	// Verify: startStream failed, then postMessage, then two updates, then final update.
	calls := mock.getCalls()
	assert.Equal(t, "chat.startStream", calls[0].method) // Failed.
	assert.Equal(t, "chat.postMessage", calls[1].method)  // Initial message.
	assert.Equal(t, "chat.update", calls[2].method)        // First append.
	assert.Equal(t, "chat.update", calls[3].method)        // Second append.
	assert.Equal(t, "chat.update", calls[4].method)        // Final on stop.
}

func TestSender_FallbackToSingleMessage(t *testing.T) {
	mock := newMockSlackServer(t)
	defer mock.server.Close()

	// Make both native and edit-based fail.
	mock.setError("chat.startStream", "not_allowed")
	mock.setError("chat.postMessage", "channel_not_found")

	sender := senderWithMockServer(t, mock)
	ctx := context.Background()

	// Start stream (should fall back to single-message mode).
	streamID, err := sender.StartStream(ctx, "C001", "1234567890.000000")
	require.NoError(t, err)
	assert.Contains(t, streamID, "single-")

	// Append chunks (accumulated in buffer).
	err = sender.AppendStream(ctx, streamID, "Hello ")
	require.NoError(t, err)
	err = sender.AppendStream(ctx, streamID, "World!")
	require.NoError(t, err)

	// Clear postMessage error for the final send.
	mock.setError("chat.postMessage", "")

	// Stop stream (should post accumulated message).
	err = sender.StopStream(ctx, streamID)
	require.NoError(t, err)
}

func TestSender_StopStreamWithError(t *testing.T) {
	mock := newMockSlackServer(t)
	defer mock.server.Close()

	sender := senderWithMockServer(t, mock)
	ctx := context.Background()

	// Start a native stream.
	streamID, err := sender.StartStream(ctx, "C001", "1234567890.000000")
	require.NoError(t, err)

	// Append partial content.
	err = sender.AppendStream(ctx, streamID, "Partial response...")
	require.NoError(t, err)

	// BC-09: Stop with error indicator.
	err = sender.StopStreamWithError(ctx, streamID, "\n\n_Response interrupted. Please try again._")
	require.NoError(t, err)

	// Verify stopStream was called with markdown_text.
	calls := mock.getCalls()
	lastCall := calls[len(calls)-1]
	assert.Equal(t, "chat.stopStream", lastCall.method)
	if text, ok := lastCall.payload["markdown_text"]; ok {
		assert.Contains(t, text, "Response interrupted")
	}
}

func TestSender_StopStream_UnknownID(t *testing.T) {
	sender := NewSender("xoxb-test", "T001")

	// Should not error for unknown stream ID.
	err := sender.StopStream(context.Background(), "nonexistent")
	assert.NoError(t, err)
}

func TestExtractCitations(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single citation",
			text:     "The architecture is described in [autom8y::knossos::architecture].",
			expected: []string{"autom8y::knossos::architecture"},
		},
		{
			name:     "multiple citations",
			text:     "See [autom8y::knossos::architecture] and [autom8y::data::conventions].",
			expected: []string{"autom8y::knossos::architecture", "autom8y::data::conventions"},
		},
		{
			name:     "duplicate citations deduplicated",
			text:     "[autom8y::knossos::architecture] mentions [autom8y::knossos::architecture] again.",
			expected: []string{"autom8y::knossos::architecture"},
		},
		{
			name:     "no citations",
			text:     "This is a plain response with no citations.",
			expected: nil,
		},
		{
			name:     "citation with hyphens and underscores",
			text:     "Found in [autom8y::my-repo::scar-tissue] and [org_2::repo_1::design-constraints].",
			expected: []string{"autom8y::my-repo::scar-tissue", "org_2::repo_1::design-constraints"},
		},
		{
			name:     "partial citation not matched",
			text:     "Reference: [autom8y::knossos] is incomplete.",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractCitations(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}
