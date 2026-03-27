package reason

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/trust"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/intent"
	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/reason/testdata"
)

// TestClassifier_IntentCorpus validates intent classifier accuracy against
// the full 12-query edge case corpus (PT-06-C1 gate: >= 80% accuracy).
func TestClassifier_IntentCorpus(t *testing.T) {
	c := intent.NewClassifier()

	total := len(testdata.IntentEdgeCaseQueries)
	passed := 0

	for _, q := range testdata.IntentEdgeCaseQueries {
		q := q // capture
		t.Run(q.ID+"_"+q.Description, func(t *testing.T) {
			result := c.Classify(q.Query)

			// Map expected string to ActionTier.
			var expectedTier intent.ActionTier
			switch q.ExpectedIntent {
			case "OBSERVE":
				expectedTier = intent.TierObserve
			case "RECORD":
				expectedTier = intent.TierRecord
			case "ACT":
				expectedTier = intent.TierAct
			}

			// Safety invariant: Record/Act must NEVER be misclassified as Observe.
			if expectedTier == intent.TierRecord || expectedTier == intent.TierAct {
				if result.Tier == intent.TierObserve {
					t.Errorf("SAFETY VIOLATION: %s classified as Observe (expected %s): %q",
						expectedTier, expectedTier, q.Query)
					return
				}
			}

			if result.Tier == expectedTier {
				passed++
			} else {
				t.Errorf("query %s: expected %s, got %s: %q",
					q.ID, expectedTier, result.Tier, q.Query)
			}

			// Answerability check.
			assert.Equal(t, q.ExpectedAnswerable, result.Answerable,
				"query %s: answerability mismatch", q.ID)
		})
	}

	// PT-06-C1: >= 80% accuracy on edge cases.
	accuracy := float64(passed) / float64(total)
	require.GreaterOrEqualf(t, accuracy, 0.80,
		"intent classifier accuracy %.0f%% is below PT-06-C1 threshold (80%%): %d/%d passed",
		accuracy*100, passed, total)
	t.Logf("intent classifier accuracy: %.0f%% (%d/%d)", accuracy*100, passed, total)
}

// TestClassifier_ObserveCorpus validates that all HIGH and MEDIUM corpus queries
// are classified as OBSERVE.
func TestClassifier_ObserveCorpus(t *testing.T) {
	c := intent.NewClassifier()

	var allObserveQueries []testdata.TestQuery
	allObserveQueries = append(allObserveQueries, testdata.HighConfidenceQueries...)
	allObserveQueries = append(allObserveQueries, testdata.MediumConfidenceQueries...)
	allObserveQueries = append(allObserveQueries, testdata.LowConfidenceQueries...)

	for _, q := range allObserveQueries {
		q := q
		t.Run(q.ID, func(t *testing.T) {
			result := c.Classify(q.Query)
			assert.Equal(t, intent.TierObserve, result.Tier,
				"expected OBSERVE for corpus query %s: %q", q.ID, q.Query)
			assert.True(t, result.Answerable,
				"expected answerable for corpus query %s", q.ID)
		})
	}
}

// TestPipeline_LowConfidence_NeverCallsClaude_CorpusValidation validates that
// with an empty search index, all queries produce LOW tier and never call Claude (D-9).
func TestPipeline_LowConfidence_NeverCallsClaude_CorpusValidation(t *testing.T) {
	t.Setenv("CLEW_CONTENT_DIR", "")
	for _, q := range testdata.LowConfidenceQueries {
		q := q
		t.Run(q.ID, func(t *testing.T) {
			mock := &response.MockClaudeClient{}
			p := buildTestPipeline(mock)

			resp, err := p.Query(context.Background(), q.Query)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, 0, mock.CallCount,
				"Claude must NOT be called for query %s (expected LOW): %q", q.ID, q.Query)
			// With empty index, all queries produce LOW tier.
			assert.Equal(t, trust.TierLow, resp.Tier,
				"expected LOW tier for unregistered domain query %s", q.ID)
		})
	}
}

// TestPipeline_BudgetCompliance_CorpusValidation validates that all 20 corpus queries
// produce assembled contexts that comply with the source material budget (PT-06-C2).
func TestPipeline_BudgetCompliance_CorpusValidation(t *testing.T) {
	assembler := reasoncontext.NewAssembler(&testTokenCounter{}, reasoncontext.DefaultAssemblerConfig())

	allQueries := testdata.AllQueries()
	for _, q := range allQueries {
		q := q
		t.Run(q.ID, func(t *testing.T) {
			assembled := assembler.Assemble(nil, nil, trust.ConfidenceScore{
				Tier: trust.TierHigh,
			}, q.Query, "autom8y")

			require.NotNil(t, assembled)
			assert.LessOrEqual(t,
				assembled.Budget.SourceMaterialTokens,
				assembled.Budget.BudgetLimit,
				"budget violation for query %s: %d > %d",
				q.ID, assembled.Budget.SourceMaterialTokens, assembled.Budget.BudgetLimit,
			)
		})
	}
}

// TestPipeline_ThreeTierDistinction validates that the three tiers produce
// distinct response shapes (PT-06-C5).
func TestPipeline_ThreeTierDistinction(t *testing.T) {
	t.Run("LOW_tier_no_claude_no_provenance", func(t *testing.T) {
		t.Setenv("CLEW_CONTENT_DIR", "")
		// With empty search index, all queries produce LOW tier.
		mock := &response.MockClaudeClient{}
		p := buildTestPipeline(mock)

		resp, err := p.Query(context.Background(), "How does billing work?")
		require.NoError(t, err)
		require.NotNil(t, resp)

		// LOW tier shape: GapAdmission, no Provenance, Claude not called.
		assert.Equal(t, trust.TierLow, resp.Tier, "LOW tier for unknown domain")
		assert.NotNil(t, resp.Gap, "LOW tier must have GapAdmission")
		assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called for LOW tier (D-9)")
		assert.Nil(t, resp.Provenance, "LOW tier response has no provenance")
		assert.NotEmpty(t, resp.Answer, "LOW tier must have an explanation")
	})

	t.Run("Record_intent_returns_unsupported", func(t *testing.T) {
		mock := &response.MockClaudeClient{}
		p := buildTestPipeline(mock)

		resp, err := p.Query(context.Background(), "Update the architecture documentation")
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, 0, mock.CallCount, "Claude not called for Record intent")
		assert.False(t, resp.Intent.Answerable)
		assert.Equal(t, "RECORD", resp.Intent.Tier)
		assert.NotEmpty(t, resp.Answer)
	})

	t.Run("Act_intent_returns_unsupported", func(t *testing.T) {
		mock := &response.MockClaudeClient{}
		p := buildTestPipeline(mock)

		resp, err := p.Query(context.Background(), "Deploy the new version to production")
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, 0, mock.CallCount, "Claude not called for Act intent")
		assert.False(t, resp.Intent.Answerable)
		assert.Equal(t, "ACT", resp.Intent.Tier)
	})
}

// TestPipeline_GapDetection validates that LOW confidence responses contain
// actionable gap information (PT-06-C4).
func TestPipeline_GapDetection(t *testing.T) {
	t.Setenv("CLEW_CONTENT_DIR", "")
	mock := &response.MockClaudeClient{}
	p := buildTestPipeline(mock)

	// All LOW corpus queries should produce gap admissions.
	for _, q := range testdata.LowConfidenceQueries {
		q := q
		t.Run(q.ID, func(t *testing.T) {
			resp, err := p.Query(context.Background(), q.Query)
			require.NoError(t, err)
			require.NotNil(t, resp)

			// With empty search index -> always LOW.
			assert.Equal(t, trust.TierLow, resp.Tier)
			assert.NotNil(t, resp.Gap, "LOW tier must have GapAdmission")
			assert.NotEmpty(t, resp.Answer, "LOW tier must have explanation in answer")
		})
	}
}

// TestPipeline_AllQueriesNeverNil validates that no corpus query returns nil (PT-06).
func TestPipeline_AllQueriesNeverNil(t *testing.T) {
	mock := &response.MockClaudeClient{
		Response: &response.CompletionResponse{
			Content:    `{"answer":"test","citations":[{"qualified_name":"x","excerpt":"e"}]}`,
			StopReason: "end_turn",
		},
	}
	p := buildTestPipeline(mock)

	allQueries := testdata.AllQueries()
	for _, q := range allQueries {
		q := q
		t.Run(q.ID, func(t *testing.T) {
			resp, err := p.Query(context.Background(), q.Query)
			require.NoError(t, err, "pipeline must not error for corpus query %s", q.ID)
			require.NotNil(t, resp, "pipeline must return non-nil for corpus query %s", q.ID)
		})
	}
}
