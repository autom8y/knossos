package triage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ====================================================================
// QA Adversary: Sprint 3 FM-3 adversarial tests
//
// These tests target the Sprint 3 implementation risk areas:
//   - Entity extraction noise (common words matching vocabulary)
//   - Domain carryover edge cases
//   - Backward compatibility (non-follow-up path unchanged)
//   - Entity cap enforcement under stress
// ====================================================================

// ---- extractEntities adversarial tests ----

func TestExtractEntities_Adversarial_CommonWordsMatchVocabulary(t *testing.T) {
	// "search", "trust", "test" are common English words that also happen
	// to be valid domain tokens. When combined with genuinely useful tokens,
	// the 5-entity cap should limit noise propagation.
	vocabulary := []string{
		"search", "trust", "test", "coverage", "error",
		"pipeline", "knossos", "triage", "architecture", "conventions",
	}

	// Typical assistant message containing common English words.
	text := `The search results show that the test coverage is good. You can trust the pipeline
to handle errors correctly. The triage system in knossos follows standard architecture conventions.`

	entities := extractEntities(text, vocabulary)

	// The 5-entity cap should prevent all 10 vocabulary terms from leaking through.
	assert.LessOrEqual(t, len(entities), 5, "5-entity cap must limit noise from common words")

	// At least some entities should be found.
	assert.Greater(t, len(entities), 0, "should find at least one entity in matching text")
}

func TestExtractEntities_Adversarial_AllVocabularyMatchesCappedAt5(t *testing.T) {
	// Every vocabulary term appears in the text. Cap must hold.
	vocabulary := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel"}
	text := "alpha bravo charlie delta echo foxtrot golf hotel"

	entities := extractEntities(text, vocabulary)
	assert.Equal(t, 5, len(entities), "must cap at exactly 5 even when all vocabulary matches")

	// First 5 vocabulary entries should be returned (iteration order).
	assert.Equal(t, "alpha", entities[0])
	assert.Equal(t, "bravo", entities[1])
	assert.Equal(t, "charlie", entities[2])
	assert.Equal(t, "delta", entities[3])
	assert.Equal(t, "echo", entities[4])
}

func TestExtractEntities_Adversarial_EmptyText(t *testing.T) {
	vocabulary := []string{"triage", "search", "pipeline"}
	entities := extractEntities("", vocabulary)
	assert.Empty(t, entities, "empty text should return no entities")
}

func TestExtractEntities_Adversarial_WhitespaceOnlyText(t *testing.T) {
	vocabulary := []string{"triage", "search", "pipeline"}
	entities := extractEntities("   \n\t  \n  ", vocabulary)
	assert.Empty(t, entities, "whitespace-only text should return no entities")
}

func TestExtractEntities_Adversarial_EmptyVocabularyReturnsEmpty(t *testing.T) {
	// Nil vocabulary.
	entities1 := extractEntities("some meaningful text about triage", nil)
	assert.Empty(t, entities1, "nil vocabulary should return empty entities")

	// Empty slice vocabulary.
	entities2 := extractEntities("some meaningful text about triage", []string{})
	assert.Empty(t, entities2, "empty slice vocabulary should return empty entities")
}

func TestExtractEntities_Adversarial_CaseSensitivityEdge(t *testing.T) {
	vocabulary := []string{"knossos", "triage"}

	// Mixed case in text.
	entities := extractEntities("KNOSSOS uses TRIAGE for query routing", vocabulary)
	assert.Contains(t, entities, "knossos", "case-insensitive matching should find lowercase vocabulary in uppercase text")
	assert.Contains(t, entities, "triage")
}

func TestExtractEntities_Adversarial_SubstringFalsePositive(t *testing.T) {
	// "arch" is a valid 4-char token that matches inside "architecture", "search", "march".
	// This tests the substring matching behavior of Contains.
	vocabulary := []string{"arch", "search"}

	// "march" contains "arch" as a substring.
	entities := extractEntities("The march deadline approaches. Search the docs.", vocabulary)

	// "arch" should match because "march" contains "arch".
	// This IS the expected behavior (substring match), but it's a noise vector.
	assert.Contains(t, entities, "arch",
		"substring matching means 'arch' matches inside 'march' -- this is expected but noisy")
	assert.Contains(t, entities, "search")
}

func TestExtractEntities_Adversarial_VeryLongText(t *testing.T) {
	// Ensure no pathological behavior with large input text.
	vocabulary := []string{"needle", "haystack"}

	// Build a 50KB text with the needle at the end.
	bigText := ""
	for i := 0; i < 500; i++ {
		bigText += "This is line number that contains no vocabulary matches at all. "
	}
	bigText += "The needle is in the haystack."

	entities := extractEntities(bigText, vocabulary)
	assert.Contains(t, entities, "needle")
	assert.Contains(t, entities, "haystack")
	assert.LessOrEqual(t, len(entities), 5)
}

func TestExtractEntities_Adversarial_DuplicateVocabularyTokens(t *testing.T) {
	// Vocabulary with duplicate entries.
	vocabulary := []string{"triage", "triage", "search", "search", "triage"}
	text := "The triage and search systems work together."

	entities := extractEntities(text, vocabulary)

	// Should deduplicate: only one "triage" and one "search".
	triageCount := 0
	searchCount := 0
	for _, e := range entities {
		if e == "triage" {
			triageCount++
		}
		if e == "search" {
			searchCount++
		}
	}
	assert.Equal(t, 1, triageCount, "should deduplicate 'triage' even with duplicate vocabulary entries")
	assert.Equal(t, 1, searchCount, "should deduplicate 'search' even with duplicate vocabulary entries")
}

// ---- buildDomainVocabulary adversarial tests ----

func TestBuildDomainVocabulary_Adversarial_EmptyDomains(t *testing.T) {
	vocab := buildDomainVocabulary(nil)
	assert.Empty(t, vocab, "nil domains should return empty vocabulary")

	vocab2 := buildDomainVocabulary([]DomainMetadata{})
	assert.Empty(t, vocab2, "empty domains slice should return empty vocabulary")
}

func TestBuildDomainVocabulary_Adversarial_AllShortTokens(t *testing.T) {
	// Qualified name where ALL tokens are < 3 chars after splitting.
	domains := []DomainMetadata{
		{QualifiedName: "ab::cd::ef-gh"},
	}

	vocab := buildDomainVocabulary(domains)

	// "ab" (2), "cd" (2), "ef" (2), "gh" (2) -- all filtered out.
	assert.Empty(t, vocab, "all tokens < 3 chars should be filtered, leaving empty vocabulary")
}

func TestBuildDomainVocabulary_Adversarial_SpecialCharsInQualifiedName(t *testing.T) {
	// Qualified names with unexpected characters (spaces, dots).
	domains := []DomainMetadata{
		{QualifiedName: "autom8y::my.repo::domain name"},
	}

	vocab := buildDomainVocabulary(domains)

	// "my.repo" does not get split on "." -- it stays as one token.
	// "domain name" has a space but doesn't get split on space -- it stays as one token.
	// Both are >= 3 chars so they pass the filter.
	assert.NotEmpty(t, vocab)
}

// ---- extractPriorEntities adversarial tests ----

func TestExtractPriorEntities_NoAssistantMessage(t *testing.T) {
	idx := testSearchIndex()
	orch := &Orchestrator{searchIndex: idx}

	// History with only user messages -- no assistant.
	history := []ThreadMessage{
		{Role: "user", Content: "What is knossos?", Timestamp: time.Now()},
		{Role: "user", Content: "Tell me more", Timestamp: time.Now()},
	}

	entities := orch.extractPriorEntities(history)
	assert.Empty(t, entities, "no assistant message should return no entities")
}

func TestExtractPriorEntities_EmptyHistory(t *testing.T) {
	idx := testSearchIndex()
	orch := &Orchestrator{searchIndex: idx}

	entities := orch.extractPriorEntities(nil)
	assert.Empty(t, entities, "nil history should return no entities")

	entities2 := orch.extractPriorEntities([]ThreadMessage{})
	assert.Empty(t, entities2, "empty history should return no entities")
}

func TestExtractPriorEntities_UsesLastAssistantMessage(t *testing.T) {
	idx := testSearchIndex()
	orch := &Orchestrator{searchIndex: idx}

	history := []ThreadMessage{
		{Role: "user", Content: "What is knossos?", Timestamp: time.Now()},
		{Role: "assistant", Content: "Knossos is an architecture platform.", Timestamp: time.Now()},
		{Role: "user", Content: "What about autom8y-web?", Timestamp: time.Now()},
		{Role: "assistant", Content: "autom8y-web is the web frontend with conventions.", Timestamp: time.Now()},
	}

	entities := orch.extractPriorEntities(history)

	// Should extract from the LAST assistant message (the autom8y-web one).
	// "autom8y" and "conventions" should be among the entities.
	assert.NotEmpty(t, entities, "should extract entities from last assistant message")
}

// ---- injectPriorDomains adversarial tests ----

func TestInjectPriorDomains_Adversarial_AllPriorDomainsAlreadyPresent(t *testing.T) {
	idx := testSearchIndex()
	orch := &Orchestrator{searchIndex: idx}

	candidates := []stage2Candidate{
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::architecture"}, bm25Score: 0.9},
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::scar-tissue"}, bm25Score: 0.7},
	}

	// Both prior domains are already in Stage 2.
	result := orch.injectPriorDomains("architecture", candidates, []string{
		"autom8y::knossos::architecture",
		"autom8y::knossos::scar-tissue",
	})

	assert.Len(t, result, 2, "should NOT duplicate when all prior domains already present")
}

func TestInjectPriorDomains_Adversarial_PriorDomainNotInMetadata(t *testing.T) {
	// Prior domain that doesn't exist in the metadata index.
	idx := &mockSearchIndex{
		bm25Results: testBM25Results(),
		metadata:    testMetadataMap(), // Does NOT contain "nonexistent::domain".
		allDomains:  testDomains(),
	}
	orch := &Orchestrator{searchIndex: idx}

	candidates := []stage2Candidate{
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::architecture"}, bm25Score: 0.9},
	}

	result := orch.injectPriorDomains("architecture", candidates, []string{
		"nonexistent::domain::doesnotexist",
	})

	// The nonexistent domain should be silently skipped (GetMetadata returns false).
	assert.Len(t, result, 1, "nonexistent prior domain should be silently skipped")
}

func TestInjectPriorDomains_Adversarial_SoftFloorScoring(t *testing.T) {
	// BM25 returns NO results at all.
	idx := &mockSearchIndex{
		bm25Results: nil, // Empty BM25.
		metadata:    testMetadataMap(),
		allDomains:  testDomains(),
	}
	orch := &Orchestrator{searchIndex: idx}

	candidates := []stage2Candidate{
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::architecture"}, bm25Score: 0.9},
	}

	result := orch.injectPriorDomains("unrelated query", candidates, []string{
		"autom8y::knossos::scar-tissue",
	})

	assert.Len(t, result, 2, "should inject prior domain even when BM25 returns nothing")

	// The injected domain should have the soft floor score.
	for _, c := range result {
		if c.metadata.QualifiedName == "autom8y::knossos::scar-tissue" {
			assert.Equal(t, 0.1, c.bm25Score, "soft floor should be 0.1 when not in BM25 results")
		}
	}
}

func TestInjectPriorDomains_Adversarial_LargePriorDomainList(t *testing.T) {
	// What happens with a large list of prior domains (more than typical 3-5).
	var priorDomains []string
	metadata := testMetadataMap()

	// Create 20 "prior" domains, only 5 of which exist in metadata.
	for i := 0; i < 20; i++ {
		priorDomains = append(priorDomains, "nonexistent::repo::domain-"+string(rune('A'+i)))
	}
	// Add the 5 real domains.
	for _, d := range testDomains() {
		priorDomains = append(priorDomains, d.QualifiedName)
	}

	idx := &mockSearchIndex{
		bm25Results: testBM25Results(),
		metadata:    metadata,
		allDomains:  testDomains(),
	}
	orch := &Orchestrator{searchIndex: idx}

	candidates := []stage2Candidate{} // Empty starting candidates.

	result := orch.injectPriorDomains("broad query", candidates, priorDomains)

	// Only the 5 real domains should be injected (nonexistent ones silently skipped).
	assert.Equal(t, 5, len(result), "only domains with valid metadata should be injected")
}

func TestInjectPriorDomains_Adversarial_OrderAfterInjection(t *testing.T) {
	// Verify descending sort is maintained after injection.
	idx := &mockSearchIndex{
		bm25Results: []BM25Result{
			{QualifiedName: "autom8y::knossos::conventions", Score: 0.3},
		},
		metadata:   testMetadataMap(),
		allDomains: testDomains(),
	}
	orch := &Orchestrator{searchIndex: idx}

	candidates := []stage2Candidate{
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::architecture"}, bm25Score: 0.9},
		{metadata: DomainMetadata{QualifiedName: "autom8y::knossos::scar-tissue"}, bm25Score: 0.5},
	}

	result := orch.injectPriorDomains("query", candidates, []string{
		"autom8y::knossos::conventions",
	})

	// Verify descending order by BM25 score.
	for i := 1; i < len(result); i++ {
		assert.GreaterOrEqual(t, result[i-1].bm25Score, result[i].bm25Score,
			"candidates must be sorted by BM25 score descending after injection")
	}
}

// ---- Assess backward compatibility ----

func TestAssess_Adversarial_EmptyOptsIdenticalToNoOpts(t *testing.T) {
	// Passing empty AssessOptions should be functionally identical to passing none.
	mock := &mockLLMClient{response: validStage3JSON()}

	// Run without opts.
	orch1 := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})
	result1, err1 := orch1.Assess(context.Background(), "architecture overview", nil)

	// Run with empty opts.
	orch2 := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})
	result2, err2 := orch2.Assess(context.Background(), "architecture overview", nil, AssessOptions{})

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, result1)
	require.NotNil(t, result2)

	assert.Equal(t, result1.RefinedQuery, result2.RefinedQuery,
		"empty opts should produce same refined query as no opts")
	assert.Equal(t, result1.Intent.IsFollowUp, result2.Intent.IsFollowUp,
		"empty opts should produce same follow-up flag as no opts")
	assert.Equal(t, len(result1.Candidates), len(result2.Candidates),
		"empty opts should produce same candidate count as no opts")
}

func TestAssess_Adversarial_PriorDomainsWithoutHistory(t *testing.T) {
	// PriorTurnDomains without thread history should NOT activate enhancement.
	mock := &mockLLMClient{response: validStage3JSON()}
	orch := NewOrchestrator(mock, testSearchIndex(), &StubEmbeddingModel{})

	opts := AssessOptions{
		PriorTurnDomains: []string{"autom8y::knossos::conventions"},
	}

	result, err := orch.Assess(context.Background(), "architecture overview", nil, opts)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Intent.IsFollowUp,
		"prior domains without history should NOT mark as follow-up")
}

func TestAssess_Adversarial_HistoryWithoutPriorDomains(t *testing.T) {
	// Thread history WITHOUT prior domains: standard follow-up, no enhancement.
	callCount := 0
	orch := &Orchestrator{
		llmClient: &multiResponseMock{
			responses: []string{
				"What is knossos architecture", // Stage 0 refinement.
				validStage3JSON(),              // Stage 3.
			},
			callCount: &callCount,
		},
		searchIndex:    testSearchIndex(),
		embeddingModel: &StubEmbeddingModel{},
	}

	history := []ThreadMessage{
		{Role: "user", Content: "Hi", Timestamp: time.Now()},
	}

	// No opts -- standard follow-up path.
	result, err := orch.Assess(context.Background(), "What about architecture?", history)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Intent.IsFollowUp)
	// Should NOT have entity extraction (no prior domains).
}

// ---- Stage 0 user message adversarial tests ----

func TestStage0UserMessage_Adversarial_VeryLongHistory(t *testing.T) {
	// Stress test: 50-turn history.
	var history []ThreadMessage
	for i := 0; i < 50; i++ {
		history = append(history, ThreadMessage{
			Role:    "user",
			Content: "Question number " + string(rune('0'+i%10)),
		})
		history = append(history, ThreadMessage{
			Role:    "assistant",
			Content: "Answer number " + string(rune('0'+i%10)),
		})
	}

	entities := []string{"knossos", "triage"}
	msg := stage0UserMessage("Final question", history, entities)

	// Should not panic or produce malformed output.
	assert.Contains(t, msg, "Final question")
	assert.Contains(t, msg, "Prior turn entities:")
	assert.Contains(t, msg, "Conversation history:")
}

func TestStage0UserMessage_Adversarial_EmptyHistory(t *testing.T) {
	msg := stage0UserMessage("Question", nil)
	assert.Contains(t, msg, "Conversation history:")
	assert.Contains(t, msg, "Question")
	assert.NotContains(t, msg, "Prior turn entities:")
}

func TestStage0UserMessage_Adversarial_EntitiesWithSpecialChars(t *testing.T) {
	entities := []string{"autom8y", "scar-tissue", "test::coverage"}
	msg := stage0UserMessage("Question", []ThreadMessage{
		{Role: "user", Content: "Hi"},
	}, entities)

	assert.Contains(t, msg, "autom8y, scar-tissue, test::coverage")
}
