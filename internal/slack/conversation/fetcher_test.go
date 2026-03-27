package conversation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchThreadMessages_Success(t *testing.T) {
	const threadTS = "1716000000.000000"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))

		resp := map[string]any{
			"ok": true,
			"messages": []map[string]any{
				// Parent message — should be filtered out.
				{"user": "U111", "text": "parent", "ts": threadTS},
				// User message — should be included as role "user".
				{"user": "U222", "text": "hello from user", "ts": "1716000001.000000"},
				// Bot message — should be filtered out.
				{"bot_id": "B999", "text": "bot reply", "ts": "1716000002.000000"},
				// Assistant message (no user field) — should be included as role "assistant".
				{"text": "assistant reply", "ts": "1716000003.000000"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	fetcher := NewSlackThreadFetcherForTest("test-token", srv.URL+"/")
	messages, err := fetcher.FetchThreadMessages(context.Background(), "C123", threadTS, 20)

	require.NoError(t, err)
	require.Len(t, messages, 2)

	// First: user message.
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "hello from user", messages[0].Content)
	assert.False(t, messages[0].Timestamp.IsZero())

	// Second: assistant message.
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "assistant reply", messages[1].Content)
	assert.False(t, messages[1].Timestamp.IsZero())
}

func TestFetchThreadMessages_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	fetcher := NewSlackThreadFetcherForTest("test-token", srv.URL+"/")
	messages, err := fetcher.FetchThreadMessages(context.Background(), "C123", "1716000000.000000", 20)

	assert.Nil(t, messages)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestFetchThreadMessages_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"ok":    false,
			"error": "channel_not_found",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	fetcher := NewSlackThreadFetcherForTest("test-token", srv.URL+"/")
	messages, err := fetcher.FetchThreadMessages(context.Background(), "C123", "1716000000.000000", 20)

	assert.Nil(t, messages)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "channel_not_found"))
}

func TestFetchThreadMessages_EmptyThread(t *testing.T) {
	const threadTS = "1716000000.000000"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only the parent message — no replies.
		resp := map[string]any{
			"ok": true,
			"messages": []map[string]any{
				{"user": "U111", "text": "parent", "ts": threadTS},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	fetcher := NewSlackThreadFetcherForTest("test-token", srv.URL+"/")
	messages, err := fetcher.FetchThreadMessages(context.Background(), "C123", threadTS, 20)

	require.NoError(t, err)
	require.NotNil(t, messages, "should return empty slice, not nil")
	assert.Empty(t, messages)
}
