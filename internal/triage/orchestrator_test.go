package triage

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/llm"
)

// ---- Test infrastructure ----

// mockLLMClient is a test double for llm.Client within triage tests.
type mockLLMClient struct {
	response string
	err      error
}

func (m *mockLLMClient) Complete(_ context.Context, _ llm.CompletionRequest) (string, error) {
	return m.response, m.err
}

type mockSearchIndex struct {
	bm25Results []BM25Result
	metadata    map[string]*DomainMetadata
	allDomains  []DomainMetadata
}

func (m *mockSearchIndex) SearchByBM25(query string, k int) []BM25Result {
	if k > len(m.bm25Results) {
		return m.bm25Results
	}
	return m.bm25Results[:k]
}

func (m *mockSearchIndex) GetMetadata(qn string) (*DomainMetadata, bool) {
	md, ok := m.metadata[qn]
	return md, ok
}

func (m *mockSearchIndex) ListAllDomains() []DomainMetadata {
	return m.allDomains
}

func testDomains() []DomainMetadata {
	return []DomainMetadata{
		{QualifiedName: "autom8y::knossos::architecture", DomainType: "architecture", Repo: "knossos", FreshnessScore: 0.95},
		{QualifiedName: "autom8y::knossos::scar-tissue", DomainType: "scar-tissue", Repo: "knossos", FreshnessScore: 0.6},
		{QualifiedName: "autom8y::knossos::conventions", DomainType: "conventions", Repo: "knossos", FreshnessScore: 0.85},
		{QualifiedName: "autom8y::autom8y-web::architecture", DomainType: "architecture", Repo: "autom8y-web", FreshnessScore: 0.75},
		{QualifiedName: "autom8y::platform-infra::release", DomainType: "release", Repo: "platform-infra", FreshnessScore: 0.4},
	}
}

func testMetadataMap() map[string]*DomainMetadata {
	domains := testDomains()
	m := make(map[string]*DomainMetadata, len(domains))
	for i := range domains {
		m[domains[i].QualifiedName] = &domains[i]
	}
	return m
}

func testBM25Results() []BM25Result {
	return []BM25Result{
		{QualifiedName: "autom8y::knossos::architecture", Score: 0.9, Domain: "architecture"},
		{QualifiedName: "autom8y::knossos::conventions", Score: 0.7, Domain: "conventions"},
		{QualifiedName: "autom8y::autom8y-web::architecture", Score: 0.5, Domain: "architecture"},
	}
}

func testSearchIndex() *mockSearchIndex {
	return &mockSearchIndex{
		bm25Results: testBM25Results(),
		metadata:    testMetadataMap(),
		allDomains:  testDomains(),
	}
}

func validStage3JSON() string {
	resp := stage3Response{
		Candidates: []stage3Candidate{
			{QualifiedName: "autom8y::knossos::architecture", RelevanceScore: 0.95, Rationale: "Primary architecture source", DomainType: "architecture"},
			{QualifiedName: "autom8y::knossos::conventions", RelevanceScore: 0.8, Rationale: "Coding patterns", DomainType: "conventions"},
			{QualifiedName: "autom8y::autom8y-web::architecture", RelevanceScore: 0.6, Rationale: "Web architecture for comparison", DomainType: "architecture"},
		},
		Intent: stage3Intent{
			Type:              "architecture",
			TargetDomainTypes: []string{"architecture", "conventions"},
			Repos:             []string{"knossos"},
		},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

// ---- Stage 0 tests ----

func TestStage0_SkippedWhenNoHistory(t *testing.T) {
	mock := &mockLLMClient{response: "refined query"}
	orch := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})

	result, err := orch.Assess(context.Background(), "What is the architecture?", nil)

	require.NoError(t, err)
	// Stage 0 should be skipped, so no LLM calls for refinement.
	// Stage 3 makes one call.
	assert.Equal(t, "What is the architecture?", result.RefinedQuery)
}

func TestStage0_RefinesFollowUpQuery(t *testing.T) {
	callCount := 0

	// Override mock to track calls and return different responses.
	orch := &Orchestrator{
		llmClient: &multiResponseMock{
			responses: []string{
				"Compare knossos architecture to autom8y-web architecture",
				validStage3JSON(),
			},
			callCount: &callCount,
		},
		searchIndex:    testSearchIndex(),
		embeddingModel: &StubEmbeddingModel{},
	}

	history := []ThreadMessage{
		{Role: "user", Content: "How is knossos structured?", Timestamp: time.Now()},
		{Role: "assistant", Content: "Knossos uses a layered architecture...", Timestamp: time.Now()},
	}

	result, err := orch.Assess(context.Background(), "Now compare that to autom8y-web", history)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Compare knossos architecture to autom8y-web architecture", result.RefinedQuery)
	assert.True(t, result.Intent.IsFollowUp)
}

// ---- Stage 1 tests ----

func TestStage1_MetadataFilter_PassesDomainTypeMatch(t *testing.T) {
	orch := &Orchestrator{}
	domains := testDomains()

	// "bugs" should match scar-tissue via domain type signals.
	passed := orch.stage1MetadataFilter("tell me about bugs in knossos", domains)

	foundScarTissue := false
	for _, d := range passed {
		if d.DomainType == "scar-tissue" {
			foundScarTissue = true
		}
	}
	assert.True(t, foundScarTissue, "scar-tissue should pass for bug-related queries")
}

func TestStage1_MetadataFilter_PassesQualifiedNameMatch(t *testing.T) {
	orch := &Orchestrator{}
	domains := testDomains()

	passed := orch.stage1MetadataFilter("autom8y-web architecture", domains)

	foundWebArch := false
	for _, d := range passed {
		if d.QualifiedName == "autom8y::autom8y-web::architecture" {
			foundWebArch = true
		}
	}
	assert.True(t, foundWebArch, "autom8y-web should match by qualified name substring")
}

func TestStage1_MetadataFilter_ExcludesSeverelyStale(t *testing.T) {
	orch := &Orchestrator{}
	domains := []DomainMetadata{
		{QualifiedName: "a::b::fresh", DomainType: "architecture", FreshnessScore: 0.9},
		{QualifiedName: "a::b::stale", DomainType: "architecture", FreshnessScore: 0.05}, // Below 0.1 threshold.
	}

	passed := orch.stage1MetadataFilter("tell me about architecture", domains)

	for _, d := range passed {
		assert.NotEqual(t, "a::b::stale", d.QualifiedName,
			"severely stale domains (freshness < 0.1) should be excluded")
	}
}

func TestStage1_MetadataFilter_BroadQueryPassesAll(t *testing.T) {
	orch := &Orchestrator{}
	domains := testDomains()

	passed := orch.stage1MetadataFilter("tell me about everything", domains)

	// Broad query should pass most/all domains through.
	assert.GreaterOrEqual(t, len(passed), 3, "broad query should pass most domains")
}

// ---- Stage 2 tests (BC-06: BM25 fallback) ----

func TestStage2_BM25Fallback_WhenEmbeddingFails(t *testing.T) {
	orch := &Orchestrator{
		searchIndex:    testSearchIndex(),
		embeddingModel: &StubEmbeddingModel{}, // Always returns error.
	}

	candidates := testDomains()
	result, usedBM25 := orch.stage2PreFilter(context.Background(), "architecture", toMetadata(candidates))

	assert.True(t, usedBM25, "BC-06: must use BM25 fallback when embedding fails")
	assert.Greater(t, len(result), 0, "BM25 fallback must return candidates")
}

func TestStage2_BM25Fallback_CapsAt20(t *testing.T) {
	// Create 30 BM25 results to test the cap.
	var bm25Results []BM25Result
	for i := 0; i < 30; i++ {
		bm25Results = append(bm25Results, BM25Result{
			QualifiedName: "a::b::domain-" + string(rune('A'+i)),
			Score:         float64(30 - i),
			Domain:        "test",
		})
	}

	idx := &mockSearchIndex{
		bm25Results: bm25Results,
		metadata:    make(map[string]*DomainMetadata),
		allDomains:  testDomains(),
	}

	orch := &Orchestrator{
		searchIndex:    idx,
		embeddingModel: &StubEmbeddingModel{},
	}

	result := orch.stage2BM25Fallback("test", testDomains())
	assert.LessOrEqual(t, len(result), 20, "BM25 fallback must cap at 20 candidates")
}

func TestStage2_NilEmbeddingModel_UsesBM25(t *testing.T) {
	orch := &Orchestrator{
		searchIndex:    testSearchIndex(),
		embeddingModel: nil,
	}

	result, usedBM25 := orch.stage2PreFilter(context.Background(), "architecture", toMetadata(testDomains()))

	assert.True(t, usedBM25)
	assert.Greater(t, len(result), 0)
}

// ---- Stage 3 tests ----

func TestStage3_ParsesValidJSON(t *testing.T) {
	candidates := []stage2Candidate{
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::architecture", DomainType: "architecture", FreshnessScore: 0.95}},
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::conventions", DomainType: "conventions", FreshnessScore: 0.85}},
		{metadata: DomainMetadata{QualifiedName: "autom8y::autom8y-web::architecture", DomainType: "architecture", FreshnessScore: 0.75}},
	}

	result, err := parseStage3Response(validStage3JSON(), candidates)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Candidates), 3)
	assert.Equal(t, "architecture", result.Intent.Type)

	// Verify candidates are sorted by relevance descending.
	for i := 1; i < len(result.Candidates); i++ {
		assert.GreaterOrEqual(t, result.Candidates[i-1].RelevanceScore, result.Candidates[i].RelevanceScore,
			"candidates must be sorted by relevance descending")
	}
}

func TestStage3_PartialJSONRecovery(t *testing.T) {
	// Simulate truncated JSON (G-3).
	truncated := `{"candidates": [{"qualified_name": "autom8y::knossos::architecture", "relevance_score": 0.95, "rationale": "main arch", "domain_type": "architecture"}, {"qualified_name": "autom8y::knossos::conventions", "relevance_score": 0.8, "rationale": "con`

	candidates := []stage2Candidate{
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::architecture"}},
	}

	result, err := parseStage3Response(truncated, candidates)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Candidates), 1,
		"partial JSON recovery should recover at least the first complete candidate")
}

func TestStage3_CapsAt5Candidates(t *testing.T) {
	// Build response with 7 candidates.
	var candidates []stage3Candidate
	for i := 0; i < 7; i++ {
		candidates = append(candidates, stage3Candidate{
			QualifiedName:  "a::b::domain-" + string(rune('A'+i)),
			RelevanceScore: float64(7-i) / 7.0,
			DomainType:     "test",
		})
	}

	resp := stage3Response{Candidates: candidates, Intent: stage3Intent{Type: "exploration"}}
	b, _ := json.Marshal(resp)

	result, err := parseStage3Response(string(b), nil)

	require.NoError(t, err)
	assert.LessOrEqual(t, len(result.Candidates), 5, "stage 3 must cap at 5 candidates")
}

// ---- Fail-open chain tests ----

func TestFailOpen_Stage3Fails_UsesStage2Scores(t *testing.T) {
	mock := &mockLLMClient{err: errors.New("haiku unavailable")}
	orch := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})

	result, err := orch.Assess(context.Background(), "architecture overview", nil)

	require.NoError(t, err)
	require.NotNil(t, result, "fail-open: stage 3 failure should still return result")
	assert.Greater(t, len(result.Candidates), 0, "should have candidates from stage 2 fallback")
}

func TestFailOpen_Stage2Fails_BM25Fallback(t *testing.T) {
	mock := &mockLLMClient{response: validStage3JSON()}

	// Embedding model that always fails.
	orch := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})

	result, err := orch.Assess(context.Background(), "architecture overview", nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result.Candidates), 0, "BC-06: BM25 fallback should provide candidates")
}

func TestFailOpen_AllFail_ReturnsNil(t *testing.T) {
	// All stages fail: empty search index + failing LLM.
	emptyIdx := &mockSearchIndex{
		allDomains: nil,
	}
	mock := &mockLLMClient{err: errors.New("fail")}
	orch := NewOrchestrator(mock, emptyIdx, &StubEmbeddingModel{})

	result, err := orch.Assess(context.Background(), "anything", nil)

	require.NoError(t, err)
	assert.Nil(t, result, "all stages fail should return nil for v1 fallback")
}

// ---- End-to-end triage test ----

func TestAssess_EndToEnd_FirstMessage(t *testing.T) {
	mock := &mockLLMClient{response: validStage3JSON()}
	orch := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})

	result, err := orch.Assess(context.Background(), "How is the knossos architecture structured?", nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "How is the knossos architecture structured?", result.RefinedQuery,
		"first message should not be refined (no thread history)")
	assert.False(t, result.Intent.IsFollowUp)
	assert.Greater(t, len(result.Candidates), 0)
	assert.True(t, result.TriageLatency > 0, "latency must be recorded")

	// Verify candidates have required fields.
	for _, c := range result.Candidates {
		assert.NotEmpty(t, c.QualifiedName)
		assert.GreaterOrEqual(t, c.RelevanceScore, 0.0)
		assert.LessOrEqual(t, c.RelevanceScore, 1.0)
	}
}

func TestAssess_EndToEnd_FollowUpMessage(t *testing.T) {
	callCount := 0
	orch := &Orchestrator{
		llmClient: &multiResponseMock{
			responses: []string{
				"Compare knossos architecture to autom8y-web architecture",
				validStage3JSON(),
			},
			callCount: &callCount,
		},
		searchIndex:    testSearchIndex(),
		embeddingModel: &StubEmbeddingModel{},
	}

	history := []ThreadMessage{
		{Role: "user", Content: "How is knossos structured?"},
	}

	result, err := orch.Assess(context.Background(), "Compare to autom8y-web", history)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Intent.IsFollowUp)
	assert.Equal(t, 2, result.ModelCallCount, "follow-up needs 2 model calls (stage 0 + stage 3)")
}

// ---- Helpers ----

func toMetadata(domains []DomainMetadata) []DomainMetadata {
	return domains
}

// multiResponseMock returns different responses for sequential calls.
type multiResponseMock struct {
	responses []string
	callCount *int
}

func (m *multiResponseMock) Complete(_ context.Context, _ llm.CompletionRequest) (string, error) {
	idx := *m.callCount
	*m.callCount++
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return "", errors.New("no more responses configured")
}
