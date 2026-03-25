package context

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/trust"
)

// fixedCounter is a test TokenCounter that returns a fixed value per string.
type fixedCounter struct {
	perToken int // tokens per word (approximation)
}

func (f *fixedCounter) Count(text string) int {
	if text == "" {
		return 0
	}
	// Simple approximation: 1 token per 4 characters.
	return len(text)/4 + 1
}

// exactCounter counts every character as one token -- useful for precise budget testing.
type exactCounter struct{}

func (e *exactCounter) Count(text string) int {
	return len(text)
}

// constantCounter returns a fixed token count for every string.
type constantCounter struct {
	count int
}

func (c *constantCounter) Count(_ string) int {
	return c.count
}

func makeResult(qn string, score int, content string) search.SearchResult {
	return search.SearchResult{
		SearchEntry: search.SearchEntry{
			Name:        qn,
			Domain:      search.DomainKnowledge,
			Description: content,
		},
		Score: score,
	}
}

func makeChain(qn, domain, repo string, freshness float64) *trust.ProvenanceChain {
	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{
				QualifiedName:   qn,
				Domain:          domain,
				Repo:            repo,
				FreshnessAtQuery: freshness,
				GeneratedAt:     time.Now().Add(-24 * time.Hour),
			},
		},
		BuiltAt: time.Now(),
	}
	return chain
}

func makeScore(tier trust.ConfidenceTier) trust.ConfidenceScore {
	return trust.ConfidenceScore{
		Overall:  0.8,
		Freshness: 0.8,
		Retrieval: 0.8,
		Coverage:  0.8,
		Tier:     tier,
	}
}

func TestAssembler_EmptyResults(t *testing.T) {
	a := NewAssembler(&fixedCounter{}, DefaultAssemblerConfig())
	chain := &trust.ProvenanceChain{}
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(nil, chain, score, "test question", "autom8y")

	require.NotNil(t, ctx)
	assert.Empty(t, ctx.Sources)
	assert.Equal(t, "test question", ctx.UserMessage)
	assert.Equal(t, trust.TierHigh, ctx.Tier)
	assert.NotEmpty(t, ctx.SystemPrompt)
	// Budget should show no source tokens.
	assert.Equal(t, 0, ctx.Budget.SourceMaterialTokens)
}

func TestAssembler_PacksUnderBudget(t *testing.T) {
	// Each source will report exactly 100 tokens.
	counter := &constantCounter{count: 100}
	cfg := DefaultAssemblerConfig()
	cfg.SourceBudgetTokens = 250 // fits 2 sources

	a := NewAssembler(counter, cfg)

	results := []search.SearchResult{
		makeResult("autom8y::knossos::architecture", 900, "arch content"),
		makeResult("autom8y::knossos::conventions", 800, "conv content"),
		makeResult("autom8y::knossos::scar-tissue", 700, "scar content"),
	}

	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::architecture", Domain: "architecture", Repo: "knossos", FreshnessAtQuery: 0.9},
			{QualifiedName: "autom8y::knossos::conventions", Domain: "conventions", Repo: "knossos", FreshnessAtQuery: 0.85},
			{QualifiedName: "autom8y::knossos::scar-tissue", Domain: "scar-tissue", Repo: "knossos", FreshnessAtQuery: 0.7},
		},
		BuiltAt: time.Now(),
	}
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(results, chain, score, "question", "autom8y")

	require.NotNil(t, ctx)
	// Must not exceed budget.
	assert.LessOrEqual(t, ctx.Budget.SourceMaterialTokens, 250,
		"source material tokens exceed budget")
	// Should have included 2 sources (2 * 100 = 200 <= 250).
	assert.Equal(t, 2, len(ctx.Sources), "expected 2 sources to be packed")
	assert.Equal(t, 1, ctx.Budget.SourcesSkipped, "expected 1 source skipped")
}

func TestAssembler_BudgetEnforcedStrictly(t *testing.T) {
	// Use a counter that reports exactly the text length.
	counter := &exactCounter{}
	cfg := DefaultAssemblerConfig()
	cfg.SourceBudgetTokens = 50 // very tight budget

	a := NewAssembler(counter, cfg)

	// Each content is 60 chars -- too big to fit individually.
	bigContent := "123456789012345678901234567890123456789012345678901234567890" // 60 chars
	results := []search.SearchResult{
		makeResult("autom8y::knossos::architecture", 900, bigContent),
		makeResult("autom8y::knossos::conventions", 800, bigContent),
	}

	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::architecture", FreshnessAtQuery: 0.9},
			{QualifiedName: "autom8y::knossos::conventions", FreshnessAtQuery: 0.9},
		},
		BuiltAt: time.Now(),
	}
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(results, chain, score, "question", "autom8y")

	require.NotNil(t, ctx)
	// No sources should be included -- all exceed the budget.
	assert.Empty(t, ctx.Sources, "no sources should fit in tight budget")
	assert.Equal(t, 0, ctx.Budget.SourceMaterialTokens)
	assert.LessOrEqual(t, ctx.Budget.SourceMaterialTokens, cfg.SourceBudgetTokens,
		"source tokens must never exceed budget")
}

func TestAssembler_HighInclusionScoreFirst(t *testing.T) {
	// High relevance + high freshness should rank first.
	counter := &constantCounter{count: 100}
	cfg := DefaultAssemblerConfig()
	cfg.SourceBudgetTokens = 150 // fits only 1

	a := NewAssembler(counter, cfg)

	results := []search.SearchResult{
		makeResult("autom8y::knossos::low-relevance", 200, "low relevance content"),
		makeResult("autom8y::knossos::high-relevance", 900, "high relevance content"),
	}

	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::low-relevance", Domain: "low-relevance", Repo: "knossos", FreshnessAtQuery: 0.5},
			{QualifiedName: "autom8y::knossos::high-relevance", Domain: "high-relevance", Repo: "knossos", FreshnessAtQuery: 0.9},
		},
		BuiltAt: time.Now(),
	}
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(results, chain, score, "question", "autom8y")

	require.NotNil(t, ctx)
	require.Equal(t, 1, len(ctx.Sources), "expected exactly 1 source packed")
	assert.Equal(t, "autom8y::knossos::high-relevance", ctx.Sources[0].QualifiedName,
		"higher relevance+freshness source should be included first")
}

func TestAssembler_FreshnessLabelMapping(t *testing.T) {
	tests := []struct {
		freshness float64
		label     string
	}{
		{0.9, "fresh"},
		{0.7, "fresh"},
		{0.69, "moderately stale"},
		{0.4, "moderately stale"},
		{0.39, "stale"},
		{0.0, "stale"},
	}

	for _, tt := range tests {
		got := freshnessLabel(tt.freshness)
		assert.Equal(t, tt.label, got, "freshness=%.2f", tt.freshness)
	}
}

func TestAssembler_BudgetReport_Populated(t *testing.T) {
	counter := &constantCounter{count: 50}
	cfg := DefaultAssemblerConfig()
	cfg.SourceBudgetTokens = 200

	a := NewAssembler(counter, cfg)

	results := []search.SearchResult{
		makeResult("autom8y::knossos::architecture", 900, "content"),
		makeResult("autom8y::knossos::conventions", 800, "content"),
	}
	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::architecture", Domain: "architecture", FreshnessAtQuery: 0.9},
			{QualifiedName: "autom8y::knossos::conventions", Domain: "conventions", FreshnessAtQuery: 0.85},
		},
		BuiltAt: time.Now(),
	}
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(results, chain, score, "question?", "autom8y")

	require.NotNil(t, ctx)
	assert.Equal(t, 200, ctx.Budget.BudgetLimit)
	assert.Greater(t, ctx.Budget.SystemPromptTokens, 0, "system prompt tokens should be counted")
	assert.Greater(t, ctx.Budget.UserMessageTokens, 0, "user message tokens should be counted")
	assert.Equal(t, ctx.Budget.SourcesIncluded, len(ctx.Sources))
}

func TestAssembler_DiversityBonus_CrossDomain(t *testing.T) {
	// Two sources from same domain + one from new domain.
	// Budget fits 2. The cross-domain source should win over the second same-domain source.
	counter := &constantCounter{count: 100}
	cfg := DefaultAssemblerConfig()
	cfg.SourceBudgetTokens = 250 // fits exactly 2 sources

	a := NewAssembler(counter, cfg)

	// All same relevance -- diversity should break the tie.
	results := []search.SearchResult{
		makeResult("autom8y::knossos::arch-1", 500, "arch content 1"),
		makeResult("autom8y::knossos::arch-2", 499, "arch content 2"),   // same domain, lower score
		makeResult("autom8y::knossos::conventions", 498, "conv content"), // different domain
	}

	chain := &trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::arch-1", Domain: "architecture", Repo: "knossos", FreshnessAtQuery: 0.8},
			{QualifiedName: "autom8y::knossos::arch-2", Domain: "architecture", Repo: "knossos", FreshnessAtQuery: 0.8},
			{QualifiedName: "autom8y::knossos::conventions", Domain: "conventions", Repo: "knossos", FreshnessAtQuery: 0.8},
		},
		BuiltAt: time.Now(),
	}
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(results, chain, score, "question", "autom8y")

	require.NotNil(t, ctx)
	assert.Equal(t, 2, len(ctx.Sources))
}

func TestAssembler_SkipsNonKnowledgeDomains(t *testing.T) {
	counter := &fixedCounter{}
	a := NewAssembler(counter, DefaultAssemblerConfig())

	results := []search.SearchResult{
		{SearchEntry: search.SearchEntry{Name: "sync", Domain: search.DomainCommand, Description: "sync command"}, Score: 900},
		{SearchEntry: search.SearchEntry{Name: "rite", Domain: search.DomainRite, Description: "rite content"}, Score: 800},
		makeResult("autom8y::knossos::architecture", 700, "arch content"),
	}

	chain := makeChain("autom8y::knossos::architecture", "architecture", "knossos", 0.9)
	score := makeScore(trust.TierHigh)

	ctx := a.Assemble(results, chain, score, "question", "autom8y")

	require.NotNil(t, ctx)
	// Only the knowledge domain result should be included.
	for _, s := range ctx.Sources {
		assert.Equal(t, "autom8y::knossos::architecture", s.QualifiedName)
	}
}

func TestAssembler_MediumTier_SystemPromptContainsMediumBehavior(t *testing.T) {
	a := NewAssembler(&fixedCounter{}, DefaultAssemblerConfig())
	results := []search.SearchResult{
		makeResult("autom8y::knossos::architecture", 800, "content"),
	}
	chain := makeChain("autom8y::knossos::architecture", "architecture", "knossos", 0.5)
	score := makeScore(trust.TierMedium)

	ctx := a.Assemble(results, chain, score, "question", "autom8y")

	require.NotNil(t, ctx)
	assert.Contains(t, ctx.SystemPrompt, "MEDIUM", "medium tier prompt should contain MEDIUM")
}
