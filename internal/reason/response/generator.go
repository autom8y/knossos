package response

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/citation"
	"github.com/autom8y/knossos/internal/trust"
)

// GeneratorConfig controls response generation behavior.
type GeneratorConfig struct {
	// Model is the Claude model identifier.
	// Default: "claude-sonnet-4-6".
	Model string

	// MaxResponseTokens is the maximum response tokens.
	// Default: 2000.
	MaxResponseTokens int

	// Temperature controls response randomness.
	// Default: 0.2 (low creativity, high factuality).
	Temperature float64

	// TimeoutSeconds is the per-query Claude API timeout.
	// Default: 30.
	TimeoutSeconds int
}

// DefaultGeneratorConfig returns production defaults.
func DefaultGeneratorConfig() GeneratorConfig {
	return GeneratorConfig{
		Model:             "claude-sonnet-4-6",
		MaxResponseTokens: 2000,
		Temperature:       0.2,
		TimeoutSeconds:    60,
	}
}

// Generator produces ReasoningResponses from assembled contexts.
type Generator struct {
	client ClaudeClient
	config GeneratorConfig
}

// NewGenerator creates a Generator with the given client and config.
func NewGenerator(client ClaudeClient, config GeneratorConfig) *Generator {
	return &Generator{
		client: client,
		config: config,
	}
}

// clewAnswerSchema is the JSON schema for structured Claude output.
var clewAnswerSchema = &JSONSchema{
	Name:        "clew_answer",
	Description: "Structured answer with provenance citations",
	Schema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"answer": map[string]any{
				"type":        "string",
				"description": "The response text with inline [repo::domain] citations",
			},
			"citations": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"qualified_name": map[string]any{
							"type":        "string",
							"description": "Canonical cross-repo address: org::repo::domain",
						},
						"section": map[string]any{
							"type":        "string",
							"description": "Specific section heading, if applicable",
						},
						"excerpt": map[string]any{
							"type":        "string",
							"description": "Brief supporting excerpt from the source (1-2 sentences)",
						},
					},
					"required": []string{"qualified_name", "excerpt"},
				},
				"minItems": 1,
			},
			"caveats": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"answer", "citations"},
	},
}

// Generate produces a ReasoningResponse from an assembled context and trust assessment.
// Implements three-tier behavior:
//   - HIGH: call Claude + validate citations
//   - MEDIUM: call Claude + staleness caveats
//   - Degraded: fallback on Claude failure (citation list, Degraded=true)
//
// LOW tier is handled by the Pipeline before reaching the Generator.
func (g *Generator) Generate(
	ctx context.Context,
	assembled *reasoncontext.AssembledContext,
	confidence trust.ConfidenceScore,
	chain *trust.ProvenanceChain,
	intentSummary IntentSummary,
) (*ReasoningResponse, error) {
	// Apply per-query timeout from config.
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.TimeoutSeconds)*time.Second)
	defer cancel()

	// Call Claude API without tool forcing -- free-form text generation is ~2x faster.
	// The system prompt instructs Claude to use inline [org::repo::domain] citations
	// which we extract post-hoc using the same regex as the streaming path.
	completionReq := CompletionRequest{
		SystemPrompt:   assembled.SystemPrompt,
		UserMessage:    assembled.UserMessage,
		Model:          g.config.Model,
		MaxTokens:      g.config.MaxResponseTokens,
		Temperature:    g.config.Temperature,
		ResponseSchema: nil,
	}

	resp, err := g.client.Complete(queryCtx, completionReq)
	if err != nil {
		// Claude failure: return degraded response, not an error.
		slog.Error("claude API failure",
			"error", err,
			"tier", confidence.Tier.String(),
			"overall", confidence.Overall,
		)
		return g.buildDegradedResponse(err.Error(), confidence, chain, intentSummary), nil
	}

	if resp == nil || resp.Content == "" {
		return g.buildDegradedResponse("empty response from Claude", confidence, chain, intentSummary), nil
	}

	// Parse response: try JSON first (fallback for structured output), then free-form text.
	structured, ok := parseResponse(resp.Content)
	if !ok {
		return g.buildDegradedResponse("empty answer in response", confidence, chain, intentSummary), nil
	}

	// Validate citations against provenance chain (PT-06-C3).
	valid, invalid := ValidateCitations(structured.Citations, chain)
	if len(invalid) > 0 {
		slog.Debug("stripped fabricated citations",
			"stripped", len(invalid),
			"kept", len(valid),
		)
	}

	// If all citations stripped, degrade to citation-only response.
	if len(valid) == 0 && len(structured.Citations) > 0 {
		return g.buildCitationOnlyResponse(confidence, chain, intentSummary), nil
	}

	// Build token report.
	tokensUsed := TokenReport{
		PromptTokens:     resp.Usage.InputTokens,
		CompletionTokens: resp.Usage.OutputTokens,
		TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		EstimatedCostUSD: EstimateCost(g.config.Model, resp.Usage),
	}

	answer := structured.Answer

	// MEDIUM tier: add staleness footer and per-source caveats.
	if confidence.Tier == trust.TierMedium {
		answer = addStalenessFooter(answer, chain, structured.Caveats)
	}

	return &ReasoningResponse{
		Answer:     answer,
		Confidence: confidence,
		Provenance: chain,
		Citations:  valid,
		TokensUsed: tokensUsed,
		Tier:       confidence.Tier,
		Intent:     intentSummary,
	}, nil
}

// addStalenessFooter appends a staleness footer and caveats to the answer for MEDIUM tier.
func addStalenessFooter(answer string, chain *trust.ProvenanceChain, existingCaveats []string) string {
	footer := "\n\n_Note: Some sources may not reflect the latest changes._"
	if chain != nil {
		stale := chain.StaleSources(0.4)
		for _, s := range stale {
			display := strings.ReplaceAll(s.Domain, "-", " ")
			display = simpleTitleCase(display)
			if s.Repo != "" {
				footer += fmt.Sprintf("\n_The %s information from %s may not reflect recent changes._", display, s.Repo)
			} else {
				footer += fmt.Sprintf("\n_The %s information may not reflect recent changes._", display)
			}
		}
	}
	return answer + footer
}

// buildDegradedResponse constructs a degraded response when Claude fails.
// Returns citations from ProvenanceChain so the user has useful information.
func (g *Generator) buildDegradedResponse(
	reason string,
	confidence trust.ConfidenceScore,
	chain *trust.ProvenanceChain,
	intentSummary IntentSummary,
) *ReasoningResponse {
	answer := "I found relevant knowledge sources but was unable to generate a complete synthesis at this time. Here are the sources I found:"

	// Convert ProvenanceChain sources to citations with human-readable names.
	var citations []Citation
	if chain != nil {
		for _, s := range chain.Sources {
			citations = append(citations, Citation{
				QualifiedName: s.QualifiedName,
				Excerpt:       humanReadableSourceName(s.QualifiedName, s.Domain, s.Repo),
			})
		}
	}

	return &ReasoningResponse{
		Answer:         answer,
		Confidence:     confidence,
		Provenance:     chain,
		Citations:      citations,
		Tier:           confidence.Tier,
		Intent:         intentSummary,
		Degraded:       true,
		DegradedReason: reason,
	}
}

// buildCitationOnlyResponse is used when all citations are fabricated.
func (g *Generator) buildCitationOnlyResponse(
	confidence trust.ConfidenceScore,
	chain *trust.ProvenanceChain,
	intentSummary IntentSummary,
) *ReasoningResponse {
	answer := "I found relevant sources but was unable to synthesize a verified answer."

	var citations []Citation
	if chain != nil {
		for _, s := range chain.Sources {
			citations = append(citations, Citation{
				QualifiedName: s.QualifiedName,
				Excerpt:       humanReadableSourceName(s.QualifiedName, s.Domain, s.Repo),
			})
		}
	}

	return &ReasoningResponse{
		Answer:         answer,
		Confidence:     confidence,
		Provenance:     chain,
		Citations:      citations,
		Tier:           confidence.Tier,
		Intent:         intentSummary,
		Degraded:       true,
		DegradedReason: "all citations were fabricated by Claude and stripped",
	}
}

// humanReadableSourceName converts a qualified name into a readable label for degraded responses.
// Example: "autom8y::knossos::architecture" -> "Architecture (knossos)"
func humanReadableSourceName(qualifiedName, domain, repo string) string {
	if domain != "" && repo != "" {
		display := strings.ReplaceAll(domain, "-", " ")
		display = simpleTitleCase(display)
		return fmt.Sprintf("%s (%s)", display, repo)
	}
	return qualifiedName
}

// ValidateCitations cross-checks Claude's citations against the provenance chain.
// Returns valid citations and a list of stripped invalid citations.
//
// Normalization: Claude sometimes includes ##section suffixes in qualified_name
// (e.g., "autom8y::knossos::architecture##key-abstractions"). These are stripped
// before matching against the provenance chain.
func ValidateCitations(citations []Citation, chain *trust.ProvenanceChain) (valid []Citation, invalid []Citation) {
	if chain == nil || chain.IsEmpty() {
		return nil, citations
	}

	chainNames := make(map[string]bool, len(chain.Sources))
	for _, s := range chain.Sources {
		chainNames[s.QualifiedName] = true
	}

	for _, c := range citations {
		if chainNames[c.QualifiedName] {
			valid = append(valid, c)
			continue
		}
		// Strip ##section suffix and retry.
		qn := c.QualifiedName
		if idx := strings.Index(qn, "##"); idx >= 0 {
			qn = qn[:idx]
		}
		if chainNames[qn] {
			if c.Section == "" {
				c.Section = c.QualifiedName[len(qn)+2:]
			}
			c.QualifiedName = qn
			valid = append(valid, c)
			continue
		}
		slog.Debug("citation not in provenance chain",
			"cited", c.QualifiedName, "normalized", qn)
		invalid = append(invalid, c)
	}
	return valid, invalid
}

// EstimateCost computes approximate USD cost from token usage.
// Pricing is hardcoded; will be made configurable in a future iteration.
func EstimateCost(model string, usage TokenUsage) float64 {
	// Sonnet 4.5 pricing (approximate):
	// Input:  $3.00 per million tokens
	// Output: $15.00 per million tokens
	inputCost := float64(usage.InputTokens) * 3.0 / 1_000_000.0
	outputCost := float64(usage.OutputTokens) * 15.0 / 1_000_000.0
	return inputCost + outputCost
}

// simpleTitleCase capitalizes the first letter of each word, where word
// boundaries are spaces and slashes. Replaces deprecated strings.Title
// for simple ASCII domain names (e.g., "feat/materialization" -> "Feat/Materialization").
func simpleTitleCase(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	capitalizeNext := true
	for _, r := range s {
		if r == ' ' || r == '/' {
			b.WriteRune(r)
			capitalizeNext = true
		} else if capitalizeNext {
			b.WriteString(strings.ToUpper(string(r)))
			capitalizeNext = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// parseResponse attempts to extract a StructuredAnswer from Claude's response.
// Strategy: try JSON first (in case Claude returned structured output anyway),
// then fall back to free-form text parsing with inline citation extraction.
// Returns (answer, true) on success, or (zero, false) if the response is empty.
func parseResponse(content string) (StructuredAnswer, bool) {
	// Strategy 1: Try JSON parse (backward compat / if Claude returns JSON voluntarily).
	var structured StructuredAnswer
	if err := json.Unmarshal([]byte(content), &structured); err == nil && structured.Answer != "" {
		return structured, true
	}

	// Strategy 2: Free-form text with inline [org::repo::domain] citations.
	text := strings.TrimSpace(content)
	if text == "" {
		return StructuredAnswer{}, false
	}

	// Extract inline citations using the same regex as the streaming path.
	qualifiedNames := citation.ExtractCitations(text)
	var citations []Citation
	for _, qn := range qualifiedNames {
		citations = append(citations, Citation{
			QualifiedName: qn,
			Excerpt:       "Referenced inline in response",
		})
	}

	return StructuredAnswer{
		Answer:    text,
		Citations: citations,
	}, true
}

// truncate returns at most n characters of s with "..." suffix if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
