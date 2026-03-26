package response

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/trust"
)

// GenerateStream produces a streaming ReasoningResponse.
// BC-03: Uses onChunk callback to send text chunks. reason/ does NOT import slack/.
//
// The streaming path uses free-form text with inline [org::repo::domain] citation
// markers (NOT tool-forced structured output, which is incompatible with streaming).
// Novel Discovery 5: Citation quality asymmetry between streaming and non-streaming.
//
// Goroutine topology:
//   - Goroutine A: SSE reader -> tokenCh (buffered 32)
//   - Goroutine B: Throttled batcher -> accumulate tokens, flush >= 100 chars
//     or 300ms timeout -> call onChunk
//
// BC-09: On error after partial streaming, returns partial text in response.
// The handler is responsible for deferred StopStream.
func (g *Generator) GenerateStream(
	ctx context.Context,
	assembled *reasoncontext.AssembledContext,
	confidence trust.ConfidenceScore,
	chain *trust.ProvenanceChain,
	intentSummary IntentSummary,
	onChunk func(chunk string),
) (*ReasoningResponse, error) {
	if g.client == nil {
		return nil, fmt.Errorf("generator has nil client")
	}

	// Apply per-query timeout.
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.TimeoutSeconds)*time.Second)
	defer cancel()

	// Build the streaming-specific system prompt.
	// Streaming path uses free-form text with inline citations instead of tool forcing.
	streamingPrompt := buildStreamingPrompt(assembled.SystemPrompt, chain)

	// Get the concrete Anthropic client for streaming.
	anthropicClient, ok := g.client.(*AnthropicClient)
	if !ok {
		// Fallback to non-streaming for non-Anthropic clients (e.g., mock).
		slog.Info("streaming not available for non-Anthropic client, falling back to Generate")
		return g.Generate(ctx, assembled, confidence, chain, intentSummary)
	}

	client := anthropicClient.client

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(g.config.Model),
		MaxTokens: int64(g.config.MaxResponseTokens),
		System: []anthropic.TextBlockParam{
			{Text: streamingPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(assembled.UserMessage)),
		},
	}

	// Start the SSE stream.
	stream := client.Messages.NewStreaming(queryCtx, params)
	defer stream.Close()

	// Channel for SSE reader -> batcher communication.
	tokenCh := make(chan string, 32)

	// Goroutine A: SSE reader.
	var streamErr error
	go func() {
		defer close(tokenCh)
		for stream.Next() {
			event := stream.Current()
			switch evt := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				delta := evt.Delta.AsTextDelta()
				if delta.Text != "" {
					select {
					case tokenCh <- delta.Text:
					case <-queryCtx.Done():
						return
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			streamErr = err
		}
	}()

	// Goroutine B: Throttled batcher (runs inline, not a separate goroutine).
	var fullText strings.Builder
	var buf strings.Builder
	const minChunkSize = 100
	const flushInterval = 300 * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		chunk := buf.String()
		buf.Reset()
		fullText.WriteString(chunk)
		if onChunk != nil {
			onChunk(chunk)
		}
		ticker.Reset(flushInterval)
	}

	// Read from tokenCh until closed.
	for {
		select {
		case token, ok := <-tokenCh:
			if !ok {
				// Stream complete. Flush remaining buffer.
				flush()
				goto done
			}
			buf.WriteString(token)
			if buf.Len() >= minChunkSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-queryCtx.Done():
			flush()
			goto done
		}
	}

done:
	if streamErr != nil {
		slog.Error("streaming generation failed",
			"error", streamErr,
			"partial_text_len", fullText.Len(),
		)
		// BC-09: Return partial text if we have some.
		if fullText.Len() > 0 {
			return &ReasoningResponse{
				Answer:         fullText.String(),
				Confidence:     confidence,
				Provenance:     chain,
				Tier:           confidence.Tier,
				Intent:         intentSummary,
				Degraded:       true,
				DegradedReason: fmt.Sprintf("streaming interrupted: %v", streamErr),
			}, nil
		}
		return nil, fmt.Errorf("streaming generation failed: %w", streamErr)
	}

	answer := fullText.String()
	if answer == "" {
		return g.buildDegradedResponse("empty streaming response", confidence, chain, intentSummary), nil
	}

	// Post-hoc citation parsing is done by the caller (handler) using
	// streaming.ExtractCitations() -- not done here to avoid importing slack/.

	return &ReasoningResponse{
		Answer:     answer,
		Confidence: confidence,
		Provenance: chain,
		Tier:       confidence.Tier,
		Intent:     intentSummary,
	}, nil
}

// buildStreamingPrompt modifies the system prompt for the streaming path.
// Instructs Claude to use inline [org::repo::domain] citation markers
// instead of tool-forced structured output.
func buildStreamingPrompt(basePrompt string, chain *trust.ProvenanceChain) string {
	var b strings.Builder
	b.WriteString(basePrompt)
	b.WriteString("\n\n## Response Format (Streaming)\n\n")
	b.WriteString("Write your response as free-form markdown text.\n")
	b.WriteString("Embed source citations inline using the format [org::repo::domain].\n")
	b.WriteString("Example: According to [autom8y::knossos::architecture], the system uses a 3-tier model.\n\n")

	if chain != nil && len(chain.Sources) > 0 {
		b.WriteString("Available source identifiers for citation:\n")
		for _, s := range chain.Sources {
			fmt.Fprintf(&b, "- %s\n", s.QualifiedName)
		}
		b.WriteString("\nOnly cite sources from this list. Do not fabricate citations.\n")
	}

	return b.String()
}
