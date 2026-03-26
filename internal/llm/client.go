// Package llm provides a shared LLM client for transport-only API calls.
//
// BC-01: This package lives at internal/llm/ (infrastructure layer), NOT in
// internal/triage/haiku/. All Haiku callsites import from here:
//   - Query refinement (Stage 0, triage)
//   - Domain reasoning (Stage 3, triage)
//   - Summary generation (Sprint 7, knowledge index)
//   - Conversation summarization (Sprint 6, conversation manager)
//
// This package is transport-only: API key management, HTTP transport, retry,
// rate limiting. ZERO prompt engineering -- callers own their prompts.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client abstracts LLM completion calls for all pipeline callsites.
// Implementations handle transport concerns (API key, HTTP, retry).
// Callers handle prompt engineering.
type Client interface {
	// Complete sends a completion request and returns the response text.
	// Returns an error for API failures, timeouts, and rate limits.
	// The caller owns context cancellation (timeout enforcement).
	Complete(ctx context.Context, req CompletionRequest) (string, error)
}

// CompletionRequest holds parameters for a single LLM API call.
// Callers construct these with their own prompt engineering.
type CompletionRequest struct {
	// SystemPrompt is the system-level instruction.
	SystemPrompt string

	// UserMessage is the user-facing message content.
	UserMessage string

	// MaxTokens is the maximum response tokens. Must be > 0.
	MaxTokens int

	// Model is the model identifier (e.g., "claude-haiku-4-5-20250315").
	// When empty, defaults to the client's configured model.
	Model string
}

// ClientConfig holds configuration for the Anthropic LLM client.
type ClientConfig struct {
	// APIKey is the Anthropic API key. If empty, reads from ANTHROPIC_API_KEY.
	APIKey string

	// DefaultModel is the default model for requests that don't specify one.
	// Default: "claude-haiku-4-5-20250315".
	DefaultModel string

	// DefaultMaxTokens is used when CompletionRequest.MaxTokens is 0.
	// Default: 800.
	DefaultMaxTokens int
}

// DefaultClientConfig returns production defaults for the Haiku client.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		DefaultModel:     "claude-haiku-4-5-20250315",
		DefaultMaxTokens: 800,
	}
}

// AnthropicClient implements Client using the Anthropic SDK.
type AnthropicClient struct {
	apiKey string
	config ClientConfig
}

// NewAnthropicClient creates a Client backed by the Anthropic API.
// Returns an error if no API key is available.
func NewAnthropicClient(config ClientConfig) (*AnthropicClient, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required for llm.Client")
	}

	if config.DefaultModel == "" {
		config.DefaultModel = DefaultClientConfig().DefaultModel
	}
	if config.DefaultMaxTokens <= 0 {
		config.DefaultMaxTokens = DefaultClientConfig().DefaultMaxTokens
	}

	return &AnthropicClient{
		apiKey: apiKey,
		config: config,
	}, nil
}

// Complete sends a completion request to the Anthropic API.
func (c *AnthropicClient) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	model := req.Model
	if model == "" {
		model = c.config.DefaultModel
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = c.config.DefaultMaxTokens
	}

	// Create client per-call (matches existing pattern in response/claude.go).
	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(maxTokens),
		System: []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.UserMessage)),
		},
	}

	msg, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("llm completion failed: %w", err)
	}

	// Extract text from response content blocks.
	var text string
	for _, block := range msg.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			if b.Text != "" {
				text += b.Text
			}
		case anthropic.ToolUseBlock:
			inputBytes, marshalErr := json.Marshal(b.Input)
			if marshalErr == nil {
				text += string(inputBytes)
			}
		}
	}

	return text, nil
}
