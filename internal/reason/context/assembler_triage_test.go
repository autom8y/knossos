package context

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/trust"
)

// ---- BC-07: WeightedMeanFreshness tests ----

func TestWeightedMeanFreshness_EmptyCandidates(t *testing.T) {
	result := WeightedMeanFreshness(nil)
	assert.Equal(t, 0.0, result, "empty candidates should return 0.0")
}

func TestWeightedMeanFreshness_SingleCandidate(t *testing.T) {
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::c", RelevanceScore: 0.9, Freshness: 0.8},
	}
	result := WeightedMeanFreshness(candidates)
	assert.InDelta(t, 0.8, result, 0.001,
		"single candidate: freshness equals candidate's freshness")
}

func TestWeightedMeanFreshness_WeightedByRelevance(t *testing.T) {
	// BC-07: sum(RelevanceScore_i * FreshnessScore_i) / sum(RelevanceScore_i)
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::fresh", RelevanceScore: 0.9, Freshness: 0.95},
		{QualifiedName: "a::b::stale", RelevanceScore: 0.3, Freshness: 0.2},
	}

	result := WeightedMeanFreshness(candidates)

	// Expected: (0.9*0.95 + 0.3*0.2) / (0.9+0.3) = (0.855 + 0.06) / 1.2 = 0.7625
	expected := (0.9*0.95 + 0.3*0.2) / (0.9 + 0.3)
	assert.InDelta(t, expected, result, 0.001,
		"BC-07: weighted mean should favor high-relevance candidates")
}

func TestWeightedMeanFreshness_HighRelevanceDominates(t *testing.T) {
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::primary", RelevanceScore: 0.95, Freshness: 0.9},
		{QualifiedName: "a::b::secondary", RelevanceScore: 0.1, Freshness: 0.1},
	}

	result := WeightedMeanFreshness(candidates)

	// High-relevance candidate should dominate: result should be close to 0.9, not 0.5.
	assert.Greater(t, result, 0.8, "high-relevance candidate should dominate freshness")
}

func TestWeightedMeanFreshness_ZeroRelevanceScores(t *testing.T) {
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::c", RelevanceScore: 0.0, Freshness: 0.8},
		{QualifiedName: "a::b::d", RelevanceScore: 0.0, Freshness: 0.5},
	}

	result := WeightedMeanFreshness(candidates)
	assert.Equal(t, 0.0, result, "all-zero relevance scores should return 0.0")
}

func TestWeightedMeanFreshness_ZeroFreshness_Tier1Default(t *testing.T) {
	// BC-12: Freshness is zero-valued in Tier 1 until populated.
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::c", RelevanceScore: 0.9, Freshness: 0.0},
		{QualifiedName: "a::b::d", RelevanceScore: 0.7, Freshness: 0.0},
	}

	result := WeightedMeanFreshness(candidates)
	assert.Equal(t, 0.0, result, "Tier 1 zero freshness should propagate")
}

func TestWeightedMeanFreshness_ThreeCandidates(t *testing.T) {
	candidates := []TriageCandidateInfo{
		{QualifiedName: "a::b::arch", RelevanceScore: 0.95, Freshness: 0.9},
		{QualifiedName: "a::b::conv", RelevanceScore: 0.80, Freshness: 0.85},
		{QualifiedName: "a::b::scar", RelevanceScore: 0.60, Freshness: 0.5},
	}

	result := WeightedMeanFreshness(candidates)

	expected := (0.95*0.9 + 0.80*0.85 + 0.60*0.5) / (0.95 + 0.80 + 0.60)
	assert.InDelta(t, expected, result, 0.001)
}

// ---- FM-3: resolveSourceBudget tests ----

func TestResolveSourceBudget_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name         string
		domainCount  int
		fallback     int
		expected     int
	}{
		{name: "zero domains uses fallback", domainCount: 0, fallback: 8000, expected: 8000},
		{name: "negative domains uses fallback", domainCount: -1, fallback: 8000, expected: 8000},
		{name: "1 domain uses fallback", domainCount: 1, fallback: 8000, expected: 8000},
		{name: "2 domains uses fallback", domainCount: 2, fallback: 8000, expected: 8000},
		{name: "3 domains returns 12000", domainCount: 3, fallback: 8000, expected: 12000},
		{name: "4 domains returns 12000", domainCount: 4, fallback: 8000, expected: 12000},
		{name: "5 domains returns 16000", domainCount: 5, fallback: 8000, expected: 16000},
		{name: "10 domains returns 16000", domainCount: 10, fallback: 8000, expected: 16000},
		{name: "custom fallback used for 2 domains", domainCount: 2, fallback: 6000, expected: 6000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveSourceBudget(tt.domainCount, tt.fallback)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---- FM-3: Summary tier tests ----

func TestAssembleSummaryTier_Position4PlusFallsBack(t *testing.T) {
	// When SummaryLookup is nil, full content is used for all positions.
	counter := &testTokenCounter{tokensPerChar: 1}
	config := AssemblerConfig{
		SourceBudgetTokens: 50000, // large budget to fit everything
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		SummaryLookup:      nil,
	}
	assembler := NewAssembler(counter, config)

	// Create 5 search results with varying content lengths.
	results := makeTestSearchResults(5, "Full content for domain")

	chain, score := makeTestTrustData(5)

	ctx := assembler.Assemble(results, chain, score, "test question", "testorg")

	// All sources should use full content (no summary substitution).
	for _, src := range ctx.Sources {
		assert.Contains(t, src.Content, "Full content for domain",
			"nil SummaryLookup should use full content for all positions")
	}
}

func TestAssembleSummaryTier_Position4PlusUsesSummary(t *testing.T) {
	// When SummaryLookup is provided, positions 4+ should use summary content.
	summaryStore := map[string]string{
		"a::b::domain-3": "Summary of domain 3",
		"a::b::domain-4": "Summary of domain 4",
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

	results := makeTestSearchResults(5, "Full content for domain")
	chain, score := makeTestTrustData(5)

	ctx := assembler.Assemble(results, chain, score, "test question", "testorg")

	// Verify we got some sources.
	assert.GreaterOrEqual(t, len(ctx.Sources), 3, "should include at least 3 sources")

	// The first 3 sources (positions 0-2) should have full content.
	for i := 0; i < 3 && i < len(ctx.Sources); i++ {
		assert.Contains(t, ctx.Sources[i].Content, "Full content for domain",
			"position %d should use full content", i)
	}

	// Positions 3+ that have summaries should use summary content.
	for i := 3; i < len(ctx.Sources); i++ {
		qn := ctx.Sources[i].QualifiedName
		if _, hasSummary := summaryStore[qn]; hasSummary {
			assert.NotContains(t, ctx.Sources[i].Content, "Full content for domain",
				"position %d should use summary content", i)
			assert.Contains(t, ctx.Sources[i].Content, "Summary of",
				"position %d should contain summary text", i)
		}
	}
}

func TestAssembleSummaryTier_MissingSummaryUsesFullContent(t *testing.T) {
	// SummaryLookup returns false for some domains -- those should use full content.
	lookupFn := func(qn string) (string, bool) {
		return "", false // Always returns not-found.
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

	results := makeTestSearchResults(5, "Full content here")
	chain, score := makeTestTrustData(5)

	ctx := assembler.Assemble(results, chain, score, "test question", "testorg")

	// All sources should use full content because lookup always returns false.
	for _, src := range ctx.Sources {
		assert.Contains(t, src.Content, "Full content here",
			"missing summary should fall back to full content")
	}
}

// ---- FM-3: TriageDomainCount integration with budget ----

func TestAssembleBudget_TriageDomainCountAffectsBudget(t *testing.T) {
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

	// Create results with content that would exceed 8000 but fit in 16000.
	results := makeTestSearchResultsWithSize(3, 4000) // 12000 total
	chain, score := makeTestTrustData(3)

	ctx := assembler.Assemble(results, chain, score, "test", "testorg")

	// With a 16000 budget, all 3 sources (12000 tokens) should fit.
	assert.Len(t, ctx.Sources, 3, "16K budget should fit 3 sources at 4000 tokens each")
}

// ---- Test helpers ----

type testTokenCounter struct {
	tokensPerChar int
}

func (c *testTokenCounter) Count(text string) int {
	if c.tokensPerChar == 0 {
		return len(text)
	}
	return len(text) * c.tokensPerChar
}

// makeTestSearchResults creates n search results with the given content prefix.
func makeTestSearchResults(n int, contentPrefix string) []search.SearchResult {
	var results []search.SearchResult
	for i := 0; i < n; i++ {
		qn := fmt.Sprintf("a::b::domain-%d", i)
		results = append(results, search.SearchResult{
			SearchEntry: search.SearchEntry{
				Name:        qn,
				Domain:      search.DomainKnowledge,
				Description: fmt.Sprintf("%s %d with some additional text", contentPrefix, i),
			},
			Score: 900 - i*100,
		})
	}
	return results
}

// makeTestSearchResultsWithSize creates n search results with content of approximately
// the given token count (using 1:1 char-to-token mapping).
func makeTestSearchResultsWithSize(n int, tokenCount int) []search.SearchResult {
	var results []search.SearchResult
	for i := 0; i < n; i++ {
		qn := fmt.Sprintf("a::b::domain-%d", i)
		// Create content of approximately the desired token count.
		content := strings.Repeat("x", tokenCount)
		results = append(results, search.SearchResult{
			SearchEntry: search.SearchEntry{
				Name:        qn,
				Domain:      search.DomainKnowledge,
				Description: content,
			},
			Score: 900 - i*100,
		})
	}
	return results
}

// makeTestTrustData creates a ProvenanceChain and ConfidenceScore for n domains.
func makeTestTrustData(n int) (*trust.ProvenanceChain, trust.ConfidenceScore) {
	var sources []trust.ProvenanceLink
	for i := 0; i < n; i++ {
		sources = append(sources, trust.ProvenanceLink{
			QualifiedName:    fmt.Sprintf("a::b::domain-%d", i),
			Domain:           "test",
			Repo:             "b",
			FreshnessAtQuery: 0.8,
			GeneratedAt:      time.Now().Add(-24 * time.Hour),
		})
	}
	chain := &trust.ProvenanceChain{
		Sources: sources,
		BuiltAt: time.Now(),
	}
	score := trust.ConfidenceScore{
		Overall:   0.85,
		Freshness: 0.8,
		Retrieval: 0.9,
		Coverage:  1.0,
		Tier:      trust.TierHigh,
	}
	return chain, score
}
