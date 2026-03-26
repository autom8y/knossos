package response

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/trust"
)

// helpers

func makeHighConfidence() trust.ConfidenceScore {
	return trust.ConfidenceScore{
		Overall:   0.85,
		Freshness: 0.9,
		Retrieval: 0.8,
		Coverage:  0.85,
		Tier:      trust.TierHigh,
	}
}

func makeMediumConfidence() trust.ConfidenceScore {
	return trust.ConfidenceScore{
		Overall:   0.55,
		Freshness: 0.5,
		Retrieval: 0.6,
		Coverage:  0.6,
		Tier:      trust.TierMedium,
	}
}

func makeChainWithSource(qn, domain, repo string, freshness float64) *trust.ProvenanceChain {
	return &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{
				QualifiedName:    qn,
				Domain:           domain,
				Repo:             repo,
				FreshnessAtQuery: freshness,
				FilePath:         ".know/" + domain + ".md",
				GeneratedAt:      time.Now().Add(-24 * time.Hour),
			},
		},
		BuiltAt: time.Now(),
	}
}

func makeAssembledContext(tier trust.ConfidenceTier) *reasoncontext.AssembledContext {
	return &reasoncontext.AssembledContext{
		SystemPrompt: "You are Clew...",
		UserMessage:  "test question",
		Tier:         tier,
	}
}

func makeIntentSummary() IntentSummary {
	return IntentSummary{
		Tier:       "OBSERVE",
		Domains:    []string{"architecture"},
		Answerable: true,
	}
}

func validStructuredJSON(qn string) string {
	sa := StructuredAnswer{
		Answer: "This is a valid answer referencing [knossos::architecture].",
		Citations: []Citation{
			{
				QualifiedName: qn,
				Excerpt:       "This excerpt supports the claim.",
			},
		},
	}
	b, _ := json.Marshal(sa)
	return string(b)
}

// Tests

func TestGenerator_HighTier_CallsClaude_ValidCitations(t *testing.T) {
	qn := "autom8y::knossos::architecture"
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    validStructuredJSON(qn),
			StopReason: "end_turn",
			Usage:      TokenUsage{InputTokens: 100, OutputTokens: 50},
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource(qn, "architecture", "knossos", 0.9)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Claude should have been called exactly once.
	assert.Equal(t, 1, mock.CallCount, "Claude should be called once for HIGH tier")
	assert.Equal(t, trust.TierHigh, resp.Tier)
	assert.NotEmpty(t, resp.Answer)
	assert.False(t, resp.Degraded)
	assert.NotNil(t, resp.Provenance)

	// Citation should be validated and present.
	require.Len(t, resp.Citations, 1)
	assert.Equal(t, qn, resp.Citations[0].QualifiedName)

	// Token tracking.
	assert.Equal(t, 100, resp.TokensUsed.PromptTokens)
	assert.Equal(t, 50, resp.TokensUsed.CompletionTokens)
}

func TestGenerator_MediumTier_CallsClaude_StalenessFooter(t *testing.T) {
	qn := "autom8y::knossos::test-coverage"
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    validStructuredJSON(qn),
			StopReason: "end_turn",
			Usage:      TokenUsage{InputTokens: 80, OutputTokens: 40},
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource(qn, "test-coverage", "knossos", 0.3) // stale

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierMedium), makeMediumConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, mock.CallCount)
	assert.Equal(t, trust.TierMedium, resp.Tier)
	assert.False(t, resp.Degraded)

	// MEDIUM tier should append staleness footer.
	assert.Contains(t, resp.Answer, "Note:", "MEDIUM tier should include staleness note")
}

func TestGenerator_LowTier_NeverCallsClaude(t *testing.T) {
	// LOW tier should never reach the Generator -- it's handled in Pipeline.
	// But if it does, Generator handles it gracefully.
	// This test validates the Generator itself doesn't special-case LOW:
	// It still calls Claude (because the tier check is Pipeline's job).
	// The Pipeline test covers D-9 more directly.
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    `{"answer": "answer", "citations": [{"qualified_name": "x::y::z", "excerpt": "e"}]}`,
			StopReason: "end_turn",
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())

	lowConfidence := trust.ConfidenceScore{
		Overall: 0.2,
		Tier:    trust.TierLow,
	}
	chain := &trust.ProvenanceChain{}

	// Generator calls Claude even for LOW (the Pipeline should prevent this).
	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierLow), lowConfidence, chain, makeIntentSummary())
	require.NoError(t, err)
	require.NotNil(t, resp)
	// Citation validation will strip "x::y::z" since chain is empty.
	assert.True(t, resp.Degraded, "all citations stripped -> degraded")
}

func TestGenerator_ClaudeTimeout_DegradedResponse(t *testing.T) {
	mock := &MockClaudeClient{
		Err: context.DeadlineExceeded,
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource("autom8y::knossos::architecture", "architecture", "knossos", 0.9)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	// Should NOT return an error -- degraded response instead.
	require.NoError(t, err, "Claude timeout should produce degraded response, not error")
	require.NotNil(t, resp)
	assert.True(t, resp.Degraded)
	assert.Contains(t, resp.DegradedReason, "deadline exceeded",
		"degraded reason should contain timeout info")

	// Citations should come from ProvenanceChain.
	assert.NotEmpty(t, resp.Citations, "degraded response should include citations from chain")
}

func TestGenerator_ClaudeError_DegradedResponse(t *testing.T) {
	mock := &MockClaudeClient{
		Err: errors.New("API rate limit exceeded"),
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource("autom8y::knossos::conventions", "conventions", "knossos", 0.8)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Degraded)
	assert.NotEmpty(t, resp.DegradedReason)
	// Provenance chain should be attached.
	assert.NotNil(t, resp.Provenance)
}

func TestGenerator_InvalidCitations_Stripped(t *testing.T) {
	// Claude returns a citation for a source NOT in the provenance chain.
	fabricatedQN := "fabricated::org::nonexistent"
	realQN := "autom8y::knossos::architecture"

	sa := StructuredAnswer{
		Answer: "Answer with real and fake citations.",
		Citations: []Citation{
			{QualifiedName: realQN, Excerpt: "real excerpt"},
			{QualifiedName: fabricatedQN, Excerpt: "fabricated excerpt"},
		},
	}
	b, _ := json.Marshal(sa)

	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    string(b),
			StopReason: "end_turn",
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource(realQN, "architecture", "knossos", 0.9)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Degraded, "partial citation strip should not degrade")

	// Only valid citation should remain.
	require.Len(t, resp.Citations, 1)
	assert.Equal(t, realQN, resp.Citations[0].QualifiedName)
}

func TestGenerator_AllCitationsFabricated_DegradesToCitationOnly(t *testing.T) {
	sa := StructuredAnswer{
		Answer: "Answer with only fabricated citations.",
		Citations: []Citation{
			{QualifiedName: "fake::org::fake", Excerpt: "fake"},
		},
	}
	b, _ := json.Marshal(sa)

	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    string(b),
			StopReason: "end_turn",
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource("autom8y::knossos::architecture", "architecture", "knossos", 0.9)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Degraded, "all citations stripped should degrade")
	assert.Contains(t, resp.DegradedReason, "fabricated")
}

func TestGenerator_EmptyResponse_Degraded(t *testing.T) {
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    "",
			StopReason: "max_tokens",
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := &trust.ProvenanceChain{}

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Degraded)
}

func TestGenerator_PlainTextResponse_ParsedAsFreeForm(t *testing.T) {
	// With tool forcing removed, plain text from Claude is valid free-form output.
	// No inline citations means no citations extracted, but the answer is preserved.
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    "this is a plain text response with no citations",
			StopReason: "end_turn",
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource("autom8y::knossos::architecture", "architecture", "knossos", 0.9)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Plain text without citations: no citations to validate, so not degraded.
	assert.False(t, resp.Degraded, "plain text with no citations is valid free-form output")
	assert.Equal(t, "this is a plain text response with no citations", resp.Answer)
	assert.Empty(t, resp.Citations, "no inline citations to extract")
}

func TestValidateCitations_NilChain(t *testing.T) {
	citations := []Citation{
		{QualifiedName: "x::y::z", Excerpt: "e"},
	}
	valid, invalid := ValidateCitations(citations, nil)
	assert.Empty(t, valid)
	assert.Len(t, invalid, 1)
}

func TestValidateCitations_EmptyChain(t *testing.T) {
	citations := []Citation{
		{QualifiedName: "x::y::z", Excerpt: "e"},
	}
	chain := &trust.ProvenanceChain{}
	valid, invalid := ValidateCitations(citations, chain)
	assert.Empty(t, valid)
	assert.Len(t, invalid, 1)
}

func TestValidateCitations_AllValid(t *testing.T) {
	citations := []Citation{
		{QualifiedName: "autom8y::knossos::architecture", Excerpt: "e1"},
		{QualifiedName: "autom8y::knossos::conventions", Excerpt: "e2"},
	}
	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::architecture"},
			{QualifiedName: "autom8y::knossos::conventions"},
		},
	}
	valid, invalid := ValidateCitations(citations, chain)
	assert.Len(t, valid, 2)
	assert.Empty(t, invalid)
}

func TestEstimateCost(t *testing.T) {
	cost := EstimateCost("claude-sonnet-4-6", TokenUsage{
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	})
	// 1M input * $3/M + 1M output * $15/M = $18.00
	assert.InDelta(t, 18.0, cost, 0.01)
}

func TestGenerator_HighTier_SystemPromptContainsHIGH(t *testing.T) {
	qn := "autom8y::knossos::architecture"
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    validStructuredJSON(qn),
			StopReason: "end_turn",
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource(qn, "architecture", "knossos", 0.9)

	// Use a context that has a proper system prompt containing HIGH.
	assembled := &reasoncontext.AssembledContext{
		SystemPrompt: "Confidence: HIGH -- the knowledge sources are current.",
		UserMessage:  "test question",
		Tier:         trust.TierHigh,
	}

	_, err := g.Generate(context.Background(), assembled, makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.Equal(t, 1, mock.CallCount)
	// The system prompt passed to Claude should contain HIGH tier behavior.
	assert.Contains(t, mock.LastRequest.SystemPrompt, "HIGH",
		"HIGH tier prompt should contain HIGH keyword")
}

func TestGenerator_NoToolForcing(t *testing.T) {
	// Verify that Generate() sends nil ResponseSchema (no tool forcing).
	qn := "autom8y::knossos::architecture"
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    validStructuredJSON(qn),
			StopReason: "end_turn",
			Usage:      TokenUsage{InputTokens: 100, OutputTokens: 50},
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource(qn, "architecture", "knossos", 0.9)

	_, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	assert.Nil(t, mock.LastRequest.ResponseSchema,
		"Generate should not use tool forcing (ResponseSchema must be nil)")
}

func TestGenerator_FreeFormWithInlineCitations(t *testing.T) {
	// Claude returns free-form markdown with inline [org::repo::domain] citations.
	qn := "autom8y::knossos::architecture"
	freeFormText := "The architecture uses a layered design [autom8y::knossos::architecture] with clear separation of concerns."

	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    freeFormText,
			StopReason: "end_turn",
			Usage:      TokenUsage{InputTokens: 200, OutputTokens: 80},
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := makeChainWithSource(qn, "architecture", "knossos", 0.9)

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Degraded)
	assert.Equal(t, freeFormText, resp.Answer)

	// Inline citation should be extracted and validated.
	require.Len(t, resp.Citations, 1)
	assert.Equal(t, qn, resp.Citations[0].QualifiedName)
}

func TestGenerator_FreeFormMultipleCitations(t *testing.T) {
	// Claude returns free-form text referencing multiple sources.
	qn1 := "autom8y::knossos::architecture"
	qn2 := "autom8y::knossos::conventions"

	freeFormText := "The architecture [autom8y::knossos::architecture] follows conventions [autom8y::knossos::conventions] closely."
	mock := &MockClaudeClient{
		Response: &CompletionResponse{
			Content:    freeFormText,
			StopReason: "end_turn",
			Usage:      TokenUsage{InputTokens: 200, OutputTokens: 80},
		},
	}
	g := NewGenerator(mock, DefaultGeneratorConfig())
	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: qn1, Domain: "architecture", Repo: "knossos", FreshnessAtQuery: 0.9},
			{QualifiedName: qn2, Domain: "conventions", Repo: "knossos", FreshnessAtQuery: 0.8},
		},
		BuiltAt: time.Now(),
	}

	resp, err := g.Generate(context.Background(), makeAssembledContext(trust.TierHigh), makeHighConfidence(), chain, makeIntentSummary())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Degraded)
	require.Len(t, resp.Citations, 2)
	assert.Equal(t, qn1, resp.Citations[0].QualifiedName)
	assert.Equal(t, qn2, resp.Citations[1].QualifiedName)
}

func TestParseResponse_JSONFallback(t *testing.T) {
	// parseResponse should accept valid JSON as a fallback.
	qn := "autom8y::knossos::architecture"
	jsonContent := validStructuredJSON(qn)

	result, ok := parseResponse(jsonContent)

	assert.True(t, ok)
	assert.Contains(t, result.Answer, "valid answer")
	require.Len(t, result.Citations, 1)
	assert.Equal(t, qn, result.Citations[0].QualifiedName)
}

func TestParseResponse_FreeFormText(t *testing.T) {
	// parseResponse should extract inline citations from free-form text.
	text := "The system uses [autom8y::knossos::architecture] patterns."

	result, ok := parseResponse(text)

	assert.True(t, ok)
	assert.Equal(t, text, result.Answer)
	require.Len(t, result.Citations, 1)
	assert.Equal(t, "autom8y::knossos::architecture", result.Citations[0].QualifiedName)
}

func TestParseResponse_EmptyContent(t *testing.T) {
	_, ok := parseResponse("")
	assert.False(t, ok)

	_, ok = parseResponse("   ")
	assert.False(t, ok)
}

func TestParseResponse_NoCitations(t *testing.T) {
	// Free-form text without any inline citations.
	result, ok := parseResponse("A plain answer with no citation markers.")
	assert.True(t, ok)
	assert.Equal(t, "A plain answer with no citation markers.", result.Answer)
	assert.Empty(t, result.Citations)
}
