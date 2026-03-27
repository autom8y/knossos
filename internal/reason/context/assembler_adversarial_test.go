package context

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/autom8y/knossos/internal/search"
)

// ====================================================================
// QA Adversary: Sprint 3 FM-3 resolveSourceBudget adversarial tests
//
// Targets boundary conditions, large domain counts, and fallback behavior.
// ====================================================================

// ---- resolveSourceBudget boundary and edge cases ----

func TestResolveSourceBudget_Adversarial_ExactBoundaryAt2(t *testing.T) {
	// 2 is the upper boundary for "uses fallback" tier.
	result := resolveSourceBudget(2, 8000)
	assert.Equal(t, 8000, result, "exactly 2 domains should use fallback")
}

func TestResolveSourceBudget_Adversarial_ExactBoundaryAt3(t *testing.T) {
	// 3 is the lower boundary for the 12000 tier.
	result := resolveSourceBudget(3, 8000)
	assert.Equal(t, 12000, result, "exactly 3 domains should return 12000")
}

func TestResolveSourceBudget_Adversarial_ExactBoundaryAt4(t *testing.T) {
	// 4 is the upper boundary for the 12000 tier.
	result := resolveSourceBudget(4, 8000)
	assert.Equal(t, 12000, result, "exactly 4 domains should return 12000")
}

func TestResolveSourceBudget_Adversarial_ExactBoundaryAt5(t *testing.T) {
	// 5 is the lower boundary for the 16000 tier.
	result := resolveSourceBudget(5, 8000)
	assert.Equal(t, 16000, result, "exactly 5 domains should return 16000")
}

func TestResolveSourceBudget_Adversarial_ZeroDomains(t *testing.T) {
	result := resolveSourceBudget(0, 8000)
	assert.Equal(t, 8000, result, "0 domains should use fallback")
}

func TestResolveSourceBudget_Adversarial_NegativeDomains(t *testing.T) {
	result := resolveSourceBudget(-1, 8000)
	assert.Equal(t, 8000, result, "negative domains should use fallback")

	result2 := resolveSourceBudget(-100, 4000)
	assert.Equal(t, 4000, result2, "large negative should use fallback")
}

func TestResolveSourceBudget_Adversarial_LargeDomainCounts(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"20 domains", 20, 16000},
		{"50 domains", 50, 16000},
		{"100 domains", 100, 16000},
		{"1000 domains", 1000, 16000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveSourceBudget(tt.count, 8000)
			assert.Equal(t, tt.expected, result, "large domain counts should all return 16000")
		})
	}
}

func TestResolveSourceBudget_Adversarial_ZeroFallback(t *testing.T) {
	// What if the fallback itself is 0?
	result := resolveSourceBudget(0, 0)
	assert.Equal(t, 0, result, "zero fallback with zero domains should return 0")

	// But with enough domains, should still return the fixed tier values.
	result2 := resolveSourceBudget(3, 0)
	assert.Equal(t, 12000, result2, "3 domains should return 12000 regardless of fallback value")
}

func TestResolveSourceBudget_Adversarial_OneDomain(t *testing.T) {
	result := resolveSourceBudget(1, 8000)
	assert.Equal(t, 8000, result, "single domain should use fallback")
}

// ---- Summary tier adversarial tests ----

func TestAssembleSummaryTier_Adversarial_EmptySummaryFromLookup(t *testing.T) {
	// SummaryLookup returns ok=true but empty string -- should fall back to full content.
	lookupFn := func(qn string) (string, bool) {
		return "", true // Found but empty.
	}

	counter := &testTokenCounter{tokensPerChar: 1}
	config := AssemblerConfig{
		SourceBudgetTokens: 50000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		SummaryLookup:      lookupFn,
	}
	assembler := NewAssembler(counter, config)

	results := makeTestSearchResults(5, "Full content for domain")
	chain, score := makeTestTrustData(5)

	ctx := assembler.Assemble(results, chain, score, "test question", "testorg")

	// Position 3+ should still use full content because the summary is empty.
	for i := 3; i < len(ctx.Sources); i++ {
		assert.Contains(t, ctx.Sources[i].Content, "Full content for domain",
			"empty summary (ok=true, empty string) should fall back to full content at position %d", i)
	}
}

func TestAssembleSummaryTier_Adversarial_SummaryLargerThanFullContent(t *testing.T) {
	// Pathological case: summary is LARGER than the original content.
	// The assembler should still use it (no size comparison logic exists).
	summaryStore := map[string]string{
		"a::b::domain-3": strings.Repeat("X", 10000), // Much larger than original.
		"a::b::domain-4": strings.Repeat("Y", 10000),
	}
	lookupFn := func(qn string) (string, bool) {
		s, ok := summaryStore[qn]
		return s, ok
	}

	counter := &testTokenCounter{tokensPerChar: 1}
	config := AssemblerConfig{
		SourceBudgetTokens: 50000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		SummaryLookup:      lookupFn,
	}
	assembler := NewAssembler(counter, config)

	results := makeTestSearchResults(5, "Short")
	chain, score := makeTestTrustData(5)

	ctx := assembler.Assemble(results, chain, score, "test question", "testorg")

	// Verify that summary content was used for position 3+, even though it's larger.
	for i := 3; i < len(ctx.Sources); i++ {
		qn := ctx.Sources[i].QualifiedName
		if _, hasSummary := summaryStore[qn]; hasSummary {
			assert.NotContains(t, ctx.Sources[i].Content, "Short",
				"position %d should use summary content even if larger than original", i)
		}
	}
}

func TestAssembleSummaryTier_Adversarial_Position3IsFirstSummaryPosition(t *testing.T) {
	// Verify that position 3 (0-indexed) IS the first position using summary,
	// and positions 0-2 NEVER use summary.
	callLog := make(map[string]bool)
	lookupFn := func(qn string) (string, bool) {
		callLog[qn] = true
		return "summary for " + qn, true
	}

	counter := &testTokenCounter{tokensPerChar: 1}
	config := AssemblerConfig{
		SourceBudgetTokens: 50000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		SummaryLookup:      lookupFn,
	}
	assembler := NewAssembler(counter, config)

	results := makeTestSearchResults(6, "Full content")
	chain, score := makeTestTrustData(6)

	ctx := assembler.Assemble(results, chain, score, "test", "testorg")

	// First 3 positions should use full content.
	for i := 0; i < 3 && i < len(ctx.Sources); i++ {
		assert.Contains(t, ctx.Sources[i].Content, "Full content",
			"position %d should use full content, not summary", i)
	}

	// Position 3+ should use summary.
	for i := 3; i < len(ctx.Sources); i++ {
		assert.Contains(t, ctx.Sources[i].Content, "summary for",
			"position %d should use summary content", i)
	}
}

// ---- TriageDomainCount integration adversarial tests ----

func TestAssembleBudget_Adversarial_TriageDomainCountZeroUsesDefaultBudget(t *testing.T) {
	counter := &testTokenCounter{tokensPerChar: 1}

	config := AssemblerConfig{
		SourceBudgetTokens: 8000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		TriageDomainCount:  0, // Zero = use default.
	}
	assembler := NewAssembler(counter, config)

	// 3 domains at 4000 tokens each = 12000, which exceeds 8000 budget.
	results := makeTestSearchResultsWithSize(3, 4000)
	chain, score := makeTestTrustData(3)

	ctx := assembler.Assemble(results, chain, score, "test", "testorg")

	// With 8000 budget, should fit 2 sources max (2 * 4000 = 8000).
	assert.LessOrEqual(t, len(ctx.Sources), 2,
		"8000 budget should NOT fit 3 sources at 4000 tokens each")
}

func TestAssembleBudget_Adversarial_BudgetExpansionAllowsMoreSources(t *testing.T) {
	counter := &testTokenCounter{tokensPerChar: 1}

	// With TriageDomainCount=5, budget should be 16000.
	config := AssemblerConfig{
		SourceBudgetTokens: 8000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		TriageDomainCount:  5,
	}
	assembler := NewAssembler(counter, config)

	// 4 domains at 3000 tokens each = 12000, fits in 16000 but not 8000.
	results := makeTestSearchResultsWithSize(4, 3000)
	chain, score := makeTestTrustData(4)

	ctx := assembler.Assemble(results, chain, score, "test", "testorg")

	assert.Len(t, ctx.Sources, 4,
		"16K budget should fit 4 sources at 3000 tokens each (12000 total)")
}

// ---- Assembler with conversation history adversarial tests ----

func TestAssemble_Adversarial_ConversationHistoryDoesNotReduceBudget(t *testing.T) {
	counter := &testTokenCounter{tokensPerChar: 1}
	config := AssemblerConfig{
		SourceBudgetTokens: 8000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
	}
	assembler := NewAssembler(counter, config)

	results := makeTestSearchResultsWithSize(2, 3000) // 6000 total, fits in 8000.
	chain, score := makeTestTrustData(2)

	// Run without history.
	ctx1 := assembler.Assemble(results, chain, score, "question", "testorg")

	// Run with large history.
	history := []ConversationTurn{}
	for i := 0; i < 10; i++ {
		history = append(history, ConversationTurn{
			Role:    "user",
			Content: fmt.Sprintf("Question %d about various topics", i),
		})
		history = append(history, ConversationTurn{
			Role:    "assistant",
			Content: strings.Repeat("Answer content ", 50), // ~750 chars per answer.
		})
	}

	ctx2 := assembler.Assemble(results, chain, score, "question", "testorg", history)

	// Both should include the same number of sources -- history is ADDITIONAL context.
	assert.Equal(t, len(ctx1.Sources), len(ctx2.Sources),
		"conversation history should NOT reduce the source material budget")
}

func TestAssemble_Adversarial_NilHistoryIdenticalToOmitted(t *testing.T) {
	counter := &testTokenCounter{tokensPerChar: 1}
	config := AssemblerConfig{
		SourceBudgetTokens: 8000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
	}
	assembler := NewAssembler(counter, config)

	results := []search.SearchResult{
		makeResult("a::b::arch", 900, "arch content"),
	}
	chain, score := makeTestTrustData(1)

	// No history parameter at all.
	ctx1 := assembler.Assemble(results, chain, score, "question", "testorg")
	// Explicit nil history.
	ctx2 := assembler.Assemble(results, chain, score, "question", "testorg", nil)

	// System prompts should be identical (no CONVERSATION HISTORY section in either).
	assert.Equal(t, ctx1.SystemPrompt, ctx2.SystemPrompt,
		"nil history should produce identical system prompt to omitted history")
}

// ---- WeightedMeanFreshness adversarial tests ----

func TestWeightedMeanFreshness_Adversarial_SingleZeroRelevance(t *testing.T) {
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::c", RelevanceScore: 0.0, Freshness: 1.0},
	}
	result := WeightedMeanFreshness(candidates)
	assert.Equal(t, 0.0, result, "single zero-relevance candidate should return 0.0")
}

func TestWeightedMeanFreshness_Adversarial_ExtremeValues(t *testing.T) {
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::c", RelevanceScore: 1.0, Freshness: 1.0},
		{QualifiedName: "a::b::d", RelevanceScore: 1.0, Freshness: 0.0},
	}
	result := WeightedMeanFreshness(candidates)
	assert.InDelta(t, 0.5, result, 0.001, "equal weights with 1.0 and 0.0 freshness should average to 0.5")
}
