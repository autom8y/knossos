package conversation

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/autom8y/knossos/internal/llm"
)

// LLMSummarizer implements Summarizer using the shared llm.Client (BC-01).
type LLMSummarizer struct {
	client llm.Client
}

// NewLLMSummarizer creates a Summarizer backed by the Haiku LLM client.
// Returns nil if client is nil (graceful degradation: no summarization).
func NewLLMSummarizer(client llm.Client) *LLMSummarizer {
	if client == nil {
		return nil
	}
	return &LLMSummarizer{client: client}
}

const summarySystemPrompt = `You are a conversation summarizer for a knowledge-retrieval assistant.
Summarize the following conversation messages into a concise summary (max 250 tokens).
Focus on:
- What topics/questions were discussed
- What information was provided
- Any unresolved questions or follow-up context
Keep it factual and terse. Do not add commentary.`

// Summarize generates a summary of the given messages using Haiku.
// Returns empty string on failure (fail-open: caller degrades to window-only).
func (s *LLMSummarizer) Summarize(ctx context.Context, messages []ThreadMessage) string {
	if len(messages) == 0 {
		return ""
	}

	// Build the conversation text for summarization.
	var b strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&b, "%s: %s\n", msg.Role, msg.Content)
	}

	resp, err := s.client.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: summarySystemPrompt,
		UserMessage:  b.String(),
		MaxTokens:    300, // Slightly above target to account for preamble.
	})
	if err != nil {
		slog.Warn("conversation summarization failed, degrading to window-only",
			"message_count", len(messages),
			"error", err,
		)
		return ""
	}

	return strings.TrimSpace(resp)
}
