package reason

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/trust"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/intent"
	"github.com/autom8y/knossos/internal/reason/response"
)

// ---- Test infrastructure ----

// testTokenCounter approximates 1 token per word.
type testTokenCounter struct{}

func (t *testTokenCounter) Count(text string) int {
	if text == "" {
		return 0
	}
	return len(text)/5 + 1
}

// buildTestPipeline constructs a minimal but real pipeline for integration tests.
// Uses MockClaudeClient (no real API calls).
func buildTestPipeline(mock *response.MockClaudeClient) *Pipeline {
	classifier := intent.NewClassifier()
	assembler := reasoncontext.NewAssembler(&testTokenCounter{}, reasoncontext.DefaultAssemblerConfig())
	generator := response.NewGenerator(mock, response.DefaultGeneratorConfig())
	scorer := trust.NewScorer(trust.DefaultConfig())

	// Minimal search index with knowledge domain entries.
	searchIndex := buildTestSearchIndex()
	catalog := buildTestCatalog()

	config := DefaultReasoningConfig()
	config.SearchLimit = 10

	return NewPipeline(classifier, assembler, generator, scorer, searchIndex, catalog, config)
}

// buildTestSearchIndex creates a minimal SearchIndex for tests.
// Uses a stub root command to avoid the nil-command panic in Build().
func buildTestSearchIndex() *search.SearchIndex {
	// Build() requires a non-nil cobra.Command to collect CLI commands.
	rootCmd := &cobra.Command{Use: "test"}
	return search.Build(rootCmd, nil)
}

// buildTestCatalog creates a minimal DomainCatalog for provenance resolution.
func buildTestCatalog() *registryorg.DomainCatalog {
	now := time.Now().UTC()
	return &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      now.Format(time.RFC3339),
		Repos: []registryorg.RepoEntry{
			{
				Name: "knossos",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "autom8y::knossos::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   now.Add(-24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "abc1234",
					},
					{
						QualifiedName: "autom8y::knossos::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
						GeneratedAt:   now.Add(-48 * time.Hour).Format(time.RFC3339),
						SourceHash:    "def5678",
					},
				},
			},
		},
	}
}

// mockResponse builds a valid structured JSON response for MockClaudeClient.
func mockResponse(qn string) *response.CompletionResponse {
	sa := response.StructuredAnswer{
		Answer: "This is a synthesized answer about " + qn,
		Citations: []response.Citation{
			{QualifiedName: qn, Excerpt: "Relevant excerpt from the source."},
		},
	}
	b, _ := json.Marshal(sa)
	return &response.CompletionResponse{
		Content:    string(b),
		StopReason: "end_turn",
		Usage:      response.TokenUsage{InputTokens: 500, OutputTokens: 200},
	}
}

// ---- Integration tests ----

func TestPipeline_NilDependencies(t *testing.T) {
	p := &Pipeline{
		classifier: nil,
		assembler:  nil,
		generator:  nil,
		scorer:     nil,
		search:     nil,
	}
	resp, err := p.Query(context.Background(), "question")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPipeline_RecordIntent_NotCallsClaude(t *testing.T) {
	// "Update the scar tissue" -> TierRecord -> short-circuit, Claude NOT called.
	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	resp, err := p.Query(context.Background(), "Update the scar tissue for the session bug")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called for Record intent")
	assert.Equal(t, trust.TierLow, resp.Tier, "Record intent returns TierLow response")
	assert.False(t, resp.Intent.Answerable)
	assert.Equal(t, "RECORD", resp.Intent.Tier)
}

func TestPipeline_ActIntent_NotCallsClaude(t *testing.T) {
	// "Execute the migration script" -> TierAct -> short-circuit, Claude NOT called.
	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	resp, err := p.Query(context.Background(), "Execute the migration script in prod")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called for Act intent")
	assert.False(t, resp.Intent.Answerable)
	assert.Equal(t, "ACT", resp.Intent.Tier)
}

func TestPipeline_LowConfidence_NeverCallsClaude(t *testing.T) {
	// Prevent CLEW_CONTENT_DIR from polluting BM25 with section candidates
	// that resolve to catalog parents via provenance chain.
	t.Setenv("CLEW_CONTENT_DIR", "")

	// Empty search results -> zero retrieval quality -> TierLow -> Claude NOT called.
	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	// A question about an unregistered topic will produce no search results
	// and no provenance links -> TierLow.
	resp, err := p.Query(context.Background(), "How does the Kubernetes migration work?")

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Claude must not be called for LOW tier (D-9).
	assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called for LOW confidence (D-9)")
	assert.Equal(t, trust.TierLow, resp.Tier)
	assert.NotNil(t, resp.Gap, "LOW tier response must have GapAdmission")
}

func TestPipeline_EmptySearchResults_LowConfidence(t *testing.T) {
	// Explicit zero-search scenario.
	// Clear CLEW_CONTENT_DIR to prevent environment-dependent BM25 results
	// from producing section candidates that resolve to catalog parents.
	t.Setenv("CLEW_CONTENT_DIR", "")

	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	// Generic question with no matching knowledge.
	resp, err := p.Query(context.Background(), "What is the database schema for user accounts?")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, mock.CallCount, "no search results -> LOW tier -> no Claude call")
	assert.Equal(t, trust.TierLow, resp.Tier)
}

func TestPipeline_ClaudeTimeout_DegradedResponse(t *testing.T) {
	// Mock returns timeout error -> degraded response.
	mock := &response.MockClaudeClient{
		Err: context.DeadlineExceeded,
	}

	// Force HIGH/MEDIUM path by injecting a pipeline that will find results.
	// We use a custom pipeline where the scorer is set to always return HIGH.
	// Since buildTestPipeline uses real search (empty index), we need to test
	// the generator degraded path through the pipeline.
	// The simplest approach: use a query that produces HIGH/MEDIUM from the mock flow.
	// But with empty search, we'll always get LOW. We need to test the degraded path
	// more directly through the generator test -- which already does.
	// This integration test validates the full flow by using a pipeline where search
	// returns some results.

	// Build pipeline with catalog that has results.
	p := buildTestPipeline(mock)

	// The pipeline will produce LOW tier because the real search index has no content.
	// The Generator degraded path is tested in generator_test.go.
	// Here we verify the pipeline returns gracefully even when mock fails.
	resp, err := p.Query(context.Background(), "How does the sync pipeline work?")

	require.NoError(t, err)
	require.NotNil(t, resp, "pipeline must always return a response")
	// With empty search index, will be LOW tier (no results).
	// The pipeline contract: always return a response, never nil.
}

func TestPipeline_ClaudeError_GracefulResponse(t *testing.T) {
	mock := &response.MockClaudeClient{
		Err: errors.New("API rate limit"),
	}
	p := buildTestPipeline(mock)

	resp, err := p.Query(context.Background(), "What error handling patterns are used?")

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestPipeline_IntentSummary_Populated(t *testing.T) {
	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	resp, err := p.Query(context.Background(), "What is the architecture?")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "OBSERVE", resp.Intent.Tier)
	assert.True(t, resp.Intent.Answerable)
}

func TestPipeline_AllQueriesReturnNonNilResponse(t *testing.T) {
	// Every query must return a non-nil response.
	queries := []string{
		"How does the sync pipeline work?",
		"Update the architecture documentation",
		"Deploy the new release to production",
		"What is the materializer pattern?",
		"How does billing work?",
		"",
	}

	mock := &response.MockClaudeClient{
		Response: mockResponse("autom8y::knossos::architecture"),
	}
	p := buildTestPipeline(mock)

	for _, q := range queries {
		resp, err := p.Query(context.Background(), q)
		require.NoError(t, err, "pipeline should not error for query: %q", q)
		require.NotNil(t, resp, "pipeline must return non-nil response for query: %q", q)
	}
}

func TestPipeline_BudgetCompliance(t *testing.T) {
	// All queries must produce assembled contexts with SourceMaterialTokens <= BudgetLimit.
	// We test this by building the assembler directly and asserting budget compliance.
	assembler := reasoncontext.NewAssembler(&testTokenCounter{}, reasoncontext.DefaultAssemblerConfig())

	// Simulate various search result sets.
	testCases := []struct {
		name    string
		results []search.SearchResult
	}{
		{
			name:    "empty",
			results: nil,
		},
		{
			name: "single result",
			results: []search.SearchResult{
				{SearchEntry: search.SearchEntry{Name: "autom8y::knossos::architecture", Domain: search.DomainKnowledge, Description: "arch content"}, Score: 900},
			},
		},
		{
			name: "many results",
			results: func() []search.SearchResult {
				var rs []search.SearchResult
				for i := 0; i < 20; i++ {
					rs = append(rs, search.SearchResult{
						SearchEntry: search.SearchEntry{
							Name:        fmt.Sprintf("autom8y::knossos::domain-%d", i),
							Domain:      search.DomainKnowledge,
							Description: "content for domain " + fmt.Sprintf("%d", i),
						},
						Score: 900 - i*10,
					})
				}
				return rs
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chain := &trust.ProvenanceChain{BuiltAt: time.Now()}
			score := trust.ConfidenceScore{
				Overall: 0.8,
				Tier:    trust.TierHigh,
			}

			ctx := assembler.Assemble(tc.results, chain, score, "test question", "autom8y")
			require.NotNil(t, ctx)

			assert.LessOrEqual(t,
				ctx.Budget.SourceMaterialTokens,
				ctx.Budget.BudgetLimit,
				"source material must not exceed budget for case: %s", tc.name,
			)
		})
	}
}

func TestPipeline_UnsupportedResponse_HasReason(t *testing.T) {
	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	// Record intent.
	resp, err := p.Query(context.Background(), "Create a new architecture doc")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, mock.CallCount)
	assert.NotEmpty(t, resp.Answer)

	// Act intent.
	resp2, err2 := p.Query(context.Background(), "Run the deployment script")
	require.NoError(t, err2)
	require.NotNil(t, resp2)
	assert.Equal(t, 0, mock.CallCount)
	assert.NotEmpty(t, resp2.Answer)
}

// ---- Helper function tests ----

func TestNormalizeRetrievalQuality(t *testing.T) {
	// Table-driven: verify RRF scores (x1000 scaled) normalize correctly.
	// Only DomainKnowledge results are considered; others are ignored.
	// Max theoretical RRF: 3000/(40+1) ≈ 73 (top hit in all 3 retrieval lists).
	tests := []struct {
		name     string
		results  []search.SearchResult
		expected float64
	}{
		{"empty results", nil, 0.0},
		{"no knowledge results", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainCommand}, Score: 500},
		}, 0.0},
		{"non-knowledge results ignored", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainConcept}, Score: 1000},
			{SearchEntry: search.SearchEntry{Domain: search.DomainKnowledge}, Score: 50},
		}, 0.50},
		{"clamped at 1.0", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainKnowledge}, Score: 200},
		}, 1.0},
		{"top hit in 3 lists (max)", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainKnowledge}, Score: 73},
		}, 0.73},
		{"top hit in 2 lists", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainKnowledge}, Score: 49},
		}, 0.49},
		{"top hit in 1 list", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainKnowledge}, Score: 24},
		}, 0.24},
		{"weak single-list hit", []search.SearchResult{
			{SearchEntry: search.SearchEntry{Domain: search.DomainKnowledge}, Score: 16},
		}, 0.16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, normalizeRetrievalQuality(tt.results), 0.001)
		})
	}
}

func TestComputeDomainCoverage_EmptyHints(t *testing.T) {
	chain := trust.ProvenanceChain{}
	assert.Equal(t, 1.0, computeDomainCoverage(nil, chain), "empty hints -> 1.0 (unfiltered)")
}

func TestComputeDomainCoverage_AllMissing(t *testing.T) {
	hints := []intent.DomainHint{{Domain: "missing"}}
	chain := trust.ProvenanceChain{}
	assert.Equal(t, 0.0, computeDomainCoverage(hints, chain))
}

func TestComputeDomainCoverage_Partial(t *testing.T) {
	hints := []intent.DomainHint{
		{Domain: "architecture"},
		{Domain: "missing"},
	}
	chain := trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{Domain: "architecture"},
		},
	}
	assert.InDelta(t, 0.5, computeDomainCoverage(hints, chain), 0.001)
}

func TestFindMissingDomains(t *testing.T) {
	hints := []intent.DomainHint{
		{Domain: "architecture"},
		{Domain: "conventions"},
		{Domain: "missing"},
	}
	chain := trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{Domain: "architecture"},
			{Domain: "conventions"},
		},
	}
	missing := findMissingDomains(hints, chain)
	assert.Equal(t, []string{"missing"}, missing)
}

func TestFindStaleDomains_BelowThreshold(t *testing.T) {
	chain := trust.ProvenanceChain{
		Sources: []trust.ProvenanceLink{
			{QualifiedName: "autom8y::knossos::fresh", Domain: "fresh", Repo: "knossos", FreshnessAtQuery: 0.9},
			{QualifiedName: "autom8y::knossos::stale", Domain: "stale", Repo: "knossos", FreshnessAtQuery: 0.1},
		},
	}
	stale := findStaleDomains(chain, 0.4)
	assert.Len(t, stale, 1)
	assert.Equal(t, "stale", stale[0].Domain)
}

func TestBuildProvenanceLinkInputs_NilCatalog(t *testing.T) {
	results := []search.SearchResult{
		{SearchEntry: search.SearchEntry{Name: "x", Domain: search.DomainKnowledge}},
	}
	inputs := buildProvenanceLinkInputs(results, nil)
	assert.Nil(t, inputs)
}

func TestExtractSearchDomains_EmptyHints(t *testing.T) {
	domains := extractSearchDomains(nil)
	assert.Nil(t, domains)
}

func TestExtractSearchDomains_WithHints(t *testing.T) {
	hints := []intent.DomainHint{{Domain: "architecture"}}
	domains := extractSearchDomains(hints)
	assert.Equal(t, []search.Domain{search.DomainKnowledge}, domains)
}

// ---- triageCandidatesToSearchResults unit tests ----

func TestTriageCandidatesToSearchResults_WithContentLookup(t *testing.T) {
	// When a content lookup function is provided, Description should be populated.
	contentStore := map[string]string{
		"org::repo::architecture": "# Architecture\nPackage structure and layers.",
		"org::repo::conventions":  "# Conventions\nError handling patterns.",
	}
	lookup := func(qn string) (string, bool) {
		text, ok := contentStore[qn]
		return text, ok
	}

	candidates := []TriageCandidateInput{
		{QualifiedName: "org::repo::architecture", RelevanceScore: 0.95},
		{QualifiedName: "org::repo::conventions", RelevanceScore: 0.7},
	}

	results := triageCandidatesToSearchResults(candidates, lookup)

	require.Len(t, results, 2)
	assert.Equal(t, "# Architecture\nPackage structure and layers.", results[0].Description,
		"Description must contain .know/ content from BM25 index")
	assert.Equal(t, "# Conventions\nError handling patterns.", results[1].Description)
	assert.Equal(t, search.DomainKnowledge, results[0].Domain)
	assert.Equal(t, "triage", results[0].MatchType)
}

func TestTriageCandidatesToSearchResults_NilLookup(t *testing.T) {
	// When no content lookup is available, Description should remain empty.
	// This is backward compatible with the original behavior.
	candidates := []TriageCandidateInput{
		{QualifiedName: "org::repo::architecture", RelevanceScore: 0.95},
	}

	results := triageCandidatesToSearchResults(candidates, nil)

	require.Len(t, results, 1)
	assert.Empty(t, results[0].Description, "nil lookup must leave Description empty")
}

func TestTriageCandidatesToSearchResults_MissingContent(t *testing.T) {
	// When a candidate's qualified name is not in the content store,
	// Description should remain empty for that candidate.
	lookup := func(qn string) (string, bool) {
		if qn == "org::repo::architecture" {
			return "Architecture content", true
		}
		return "", false
	}

	candidates := []TriageCandidateInput{
		{QualifiedName: "org::repo::architecture", RelevanceScore: 0.95},
		{QualifiedName: "org::repo::unknown", RelevanceScore: 0.5},
	}

	results := triageCandidatesToSearchResults(candidates, lookup)

	require.Len(t, results, 2)
	assert.Equal(t, "Architecture content", results[0].Description)
	assert.Empty(t, results[1].Description, "unknown domain must have empty Description")
}

