package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClient is a test double for llm.Client.
// Records calls and returns preconfigured responses.
type MockClient struct {
	// Response is returned by Complete(). Set before test.
	Response string

	// Err is returned by Complete(). Set for error scenarios.
	Err error

	// CallCount tracks how many times Complete() was called.
	CallCount int

	// LastRequest stores the most recent CompletionRequest for assertion.
	LastRequest CompletionRequest

	// Requests stores all requests in order (for multi-call tests).
	Requests []CompletionRequest
}

// Complete records the call and returns the preconfigured response.
func (m *MockClient) Complete(_ context.Context, req CompletionRequest) (string, error) {
	m.CallCount++
	m.LastRequest = req
	m.Requests = append(m.Requests, req)
	return m.Response, m.Err
}

func TestDefaultClientConfig(t *testing.T) {
	cfg := DefaultClientConfig()
	assert.Equal(t, "claude-haiku-4-5", cfg.DefaultModel)
	assert.Equal(t, 800, cfg.DefaultMaxTokens)
}

func TestNewAnthropicClient_NoAPIKey(t *testing.T) {
	// Temporarily clear ANTHROPIC_API_KEY to test missing key.
	t.Setenv("ANTHROPIC_API_KEY", "")

	_, err := NewAnthropicClient(ClientConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ANTHROPIC_API_KEY")
}

func TestNewAnthropicClient_WithExplicitAPIKey(t *testing.T) {
	// Ensure env is clear so we only test explicit key.
	t.Setenv("ANTHROPIC_API_KEY", "")

	client, err := NewAnthropicClient(ClientConfig{
		APIKey: "test-key-123",
	})
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "claude-haiku-4-5", client.config.DefaultModel)
	assert.Equal(t, 800, client.config.DefaultMaxTokens)
}

func TestNewAnthropicClient_FromEnvVar(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "env-key-456")

	client, err := NewAnthropicClient(ClientConfig{})
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewAnthropicClient_CustomConfig(t *testing.T) {
	client, err := NewAnthropicClient(ClientConfig{
		APIKey:           "test-key",
		DefaultModel:     "claude-custom-model",
		DefaultMaxTokens: 1200,
	})
	require.NoError(t, err)
	assert.Equal(t, "claude-custom-model", client.config.DefaultModel)
	assert.Equal(t, 1200, client.config.DefaultMaxTokens)
}

func TestMockClient_RecordsCalls(t *testing.T) {
	mock := &MockClient{Response: "test response"}

	result, err := mock.Complete(context.Background(), CompletionRequest{
		SystemPrompt: "system",
		UserMessage:  "user",
		MaxTokens:    100,
	})

	require.NoError(t, err)
	assert.Equal(t, "test response", result)
	assert.Equal(t, 1, mock.CallCount)
	assert.Equal(t, "system", mock.LastRequest.SystemPrompt)
	assert.Equal(t, "user", mock.LastRequest.UserMessage)
}

func TestMockClient_ReturnsError(t *testing.T) {
	mock := &MockClient{
		Err: assert.AnError,
	}

	_, err := mock.Complete(context.Background(), CompletionRequest{
		UserMessage: "test",
	})

	require.Error(t, err)
	assert.Equal(t, 1, mock.CallCount)
}

func TestMockClient_MultipleRequests(t *testing.T) {
	mock := &MockClient{Response: "ok"}

	_, _ = mock.Complete(context.Background(), CompletionRequest{UserMessage: "first"})
	_, _ = mock.Complete(context.Background(), CompletionRequest{UserMessage: "second"})

	assert.Equal(t, 2, mock.CallCount)
	require.Len(t, mock.Requests, 2)
	assert.Equal(t, "first", mock.Requests[0].UserMessage)
	assert.Equal(t, "second", mock.Requests[1].UserMessage)
	assert.Equal(t, "second", mock.LastRequest.UserMessage)
}
