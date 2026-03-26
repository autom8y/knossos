package response

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeClient abstracts the Claude API for testing and future model substitution.
// Implementations must handle their own timeout/retry logic.
type ClaudeClient interface {
	// Complete sends a completion request and returns a structured response.
	// Returns an error for API failures, timeouts, and rate limits.
	// The caller is responsible for context cancellation (timeout enforcement).
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}

// CompletionRequest holds all parameters for a single Claude API call.
type CompletionRequest struct {
	// SystemPrompt is the system-level instruction (identity + tier behavior + sources).
	SystemPrompt string

	// UserMessage is the user's question.
	UserMessage string

	// Model is the Claude model identifier (e.g., "claude-sonnet-4-6").
	Model string

	// MaxTokens is the maximum response tokens. Must be > 0.
	MaxTokens int

	// Temperature controls response randomness. 0.0-1.0.
	// For knowledge retrieval: 0.2 (low creativity, high factuality).
	Temperature float64

	// ResponseSchema is the JSON schema for structured output.
	// When non-nil, the API returns structured JSON matching this schema.
	ResponseSchema *JSONSchema
}

// JSONSchema defines the expected response structure for structured outputs.
type JSONSchema struct {
	Name        string         // Schema name for the API
	Description string         // Human-readable description
	Schema      map[string]any // JSON Schema definition
}

// CompletionResponse holds the raw Claude API response.
type CompletionResponse struct {
	// Content is the response text (or structured JSON string).
	Content string

	// StopReason indicates why generation stopped ("end_turn", "max_tokens", etc.).
	StopReason string

	// Usage contains token consumption metrics.
	Usage TokenUsage
}

// TokenUsage tracks token consumption for a single API call.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// AnthropicClient implements ClaudeClient using the Anthropic Go SDK.
// Constructed with an API key; reuses the SDK client across calls for connection pooling.
type AnthropicClient struct {
	apiKey string
	client anthropic.Client
}

// NewAnthropicClient creates a production ClaudeClient.
// Returns an error if ANTHROPIC_API_KEY is not set.
func NewAnthropicClient() (*AnthropicClient, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
	}
	return &AnthropicClient{
		apiKey: key,
		client: anthropic.NewClient(option.WithAPIKey(key)),
	}, nil
}

// Complete sends a completion request to the Claude API using the Anthropic Go SDK.
// When ResponseSchema is set, uses tool forcing to obtain structured JSON output.
// Context cancellation propagates through the SDK call.
func (c *AnthropicClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	client := c.client

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: int64(req.MaxTokens),
		System: []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.UserMessage)),
		},
	}

	// Structured output: use tool forcing to enforce JSON schema response.
	// The model is forced to call the named tool, which yields the structured JSON.
	if req.ResponseSchema != nil {
		properties := req.ResponseSchema.Schema["properties"]
		required, _ := req.ResponseSchema.Schema["required"].([]string)

		inputSchema := anthropic.ToolInputSchemaParam{
			Properties: properties,
			Required:   required,
		}
		params.Tools = []anthropic.ToolUnionParam{
			anthropic.ToolUnionParamOfTool(inputSchema, req.ResponseSchema.Name),
		}
		// Override description after construction -- ToolUnionParamOfTool doesn't take description.
		if params.Tools[0].OfTool != nil {
			params.Tools[0].OfTool.Description = anthropic.String(req.ResponseSchema.Description)
		}
		params.ToolChoice = anthropic.ToolChoiceParamOfTool(req.ResponseSchema.Name)
	}

	msg, err := client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("claude API call: %w", err)
	}

	// Extract content from response.
	var content string
	stopReason := string(msg.StopReason)

	for _, block := range msg.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			if b.Text != "" {
				content = b.Text
			}
		case anthropic.ToolUseBlock:
			// Structured output: the tool input IS the structured JSON.
			inputBytes, marshalErr := json.Marshal(b.Input)
			if marshalErr == nil {
				content = string(inputBytes)
			}
		}
		if content != "" {
			break
		}
	}

	return &CompletionResponse{
		Content:    content,
		StopReason: stopReason,
		Usage: TokenUsage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
	}, nil
}

// MockClaudeClient is a test double for ClaudeClient.
// Records calls and returns preconfigured responses.
type MockClaudeClient struct {
	// Response is returned by Complete(). Set before test.
	Response *CompletionResponse

	// Err is returned by Complete(). Set before test for error scenarios.
	Err error

	// CallCount tracks how many times Complete() was called.
	CallCount int

	// LastRequest stores the most recent CompletionRequest for assertion.
	LastRequest CompletionRequest

	// Requests stores all requests in order (for multi-call tests).
	Requests []CompletionRequest
}

// Complete records the call and returns preconfigured response/error.
func (m *MockClaudeClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	m.CallCount++
	m.LastRequest = req
	m.Requests = append(m.Requests, req)
	return m.Response, m.Err
}
