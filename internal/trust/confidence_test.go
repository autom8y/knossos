package trust

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C-1: Confidence Model Differentiation -- 25 test scenarios producing all 3 tiers.
func TestConfidenceScoring_C1_TierDifferentiation(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	tests := []struct {
		name      string
		freshness float64
		retrieval float64
		coverage  float64
		wantTier  ConfidenceTier
		rationale string
	}{
		// === HIGH tier scenarios (8 cases) ===
		{"fresh-architecture-full-coverage", 0.95, 0.85, 1.0, TierHigh,
			"Recently generated architecture with good search match and full coverage"},
		{"fresh-conventions-full-coverage", 0.90, 0.80, 1.0, TierHigh,
			"Fresh conventions with solid match"},
		{"moderate-freshness-excellent-match", 0.70, 0.95, 1.0, TierHigh,
			"Slightly aged but excellent retrieval compensates"},
		{"perfect-inputs", 1.0, 1.0, 1.0, TierHigh,
			"Theoretical maximum: all signals perfect"},
		{"fresh-multi-domain-full-coverage", 0.85, 0.75, 1.0, TierHigh,
			"Multiple domains all fresh"},
		{"literature-query-high-freshness", 0.98, 0.70, 1.0, TierHigh,
			"Literature domain with slow decay still very fresh"},
		{"scar-tissue-query-moderate-age", 0.80, 0.80, 1.0, TierHigh,
			"Scar tissue with 60-day halflife stays fresh longer"},
		{"borderline-high", 0.75, 0.72, 0.90, TierHigh,
			"Just above HIGH threshold"},

		// === MEDIUM tier scenarios (10 cases) ===
		// Note: The weighted geometric mean with weights (0.45, 0.25, 0.30) produces
		// higher scores than simple arithmetic mean. These values are calibrated to
		// the actual formula to produce MEDIUM tier (0.4 <= overall < 0.7).
		{"half-life-reached-poor-retrieval", 0.50, 0.50, 1.0, TierMedium,
			"Domain at half-life with moderate retrieval"},
		{"stale-test-coverage", 0.25, 0.90, 1.0, TierMedium,
			"Test-coverage at 2x half-life (14 days on 7-day halflife)"},
		{"good-freshness-very-poor-retrieval", 0.90, 0.25, 1.0, TierMedium,
			"Fresh knowledge but very weak search match"},
		{"partial-coverage-moderate-rest", 0.70, 0.70, 0.40, TierMedium,
			"Moderate freshness and retrieval but low coverage"},
		{"all-moderate", 0.60, 0.60, 0.60, TierMedium,
			"Uniformly moderate across all signals"},
		{"fresh-very-poor-coverage", 0.95, 0.90, 0.30, TierMedium,
			"Excellent freshness and retrieval but very low coverage"},
		{"stale-good-retrieval-full-coverage", 0.35, 0.90, 1.0, TierMedium,
			"Old knowledge but perfect match and coverage"},
		{"moderate-all-slightly-low", 0.55, 0.65, 0.80, TierMedium,
			"Moderate freshness drags score below HIGH"},
		{"borderline-medium-low", 0.45, 0.50, 0.70, TierMedium,
			"Just above LOW threshold"},
		{"mixed-freshness-across-sources", 0.42, 0.75, 0.80, TierMedium,
			"Min freshness is low but retrieval and coverage are solid"},

		// === LOW tier scenarios (7 cases) ===
		{"zero-freshness", 0.0, 0.90, 1.0, TierLow,
			"Unparseable timestamp -> freshness 0.0 -> overall 0.0"},
		{"zero-coverage", 0.90, 0.85, 0.0, TierLow,
			"No matching domains found"},
		{"zero-retrieval", 0.90, 0.0, 1.0, TierLow,
			"Completely irrelevant results"},
		{"all-zeros", 0.0, 0.0, 0.0, TierLow,
			"No knowledge, no match, no coverage"},
		{"very-stale-poor-match", 0.10, 0.20, 0.30, TierLow,
			"Everything is weak"},
		{"unknown-domain-no-coverage", 0.0, 0.10, 0.0, TierLow,
			"Query about unknown topic with no registry entries"},
		{"slightly-above-zero-all-signals", 0.15, 0.15, 0.15, TierLow,
			"All signals barely present but insufficient"},
	}

	// Count tiers for the C-1 minimum: 20+ test cases with all 3 tiers
	tierCounts := map[ConfidenceTier]int{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.Score(ScoreInput{
				Freshness:        tt.freshness,
				RetrievalQuality: tt.retrieval,
				DomainCoverage:   tt.coverage,
			})

			assert.Equal(t, tt.wantTier, score.Tier,
				"tier mismatch for %s (overall=%.4f): %s", tt.name, score.Overall, tt.rationale)
			tierCounts[score.Tier]++

			// Verify LOW tier always has GapAdmission
			if tt.wantTier == TierLow {
				assert.NotNil(t, score.Gap, "LOW tier must have GapAdmission")
			} else {
				assert.Nil(t, score.Gap, "non-LOW tier must not have GapAdmission")
			}

			// Verify overall is in [0.0, 1.0]
			assert.GreaterOrEqual(t, score.Overall, 0.0)
			assert.LessOrEqual(t, score.Overall, 1.0)
		})
	}

	// Verify we tested all 3 tiers and hit 20+ total
	assert.GreaterOrEqual(t, len(tests), 20, "C-1 requires 20+ test cases")
	assert.Greater(t, tierCounts[TierHigh], 0, "must have HIGH tier test cases")
	assert.Greater(t, tierCounts[TierMedium], 0, "must have MEDIUM tier test cases")
	assert.Greater(t, tierCounts[TierLow], 0, "must have LOW tier test cases")
}

// C-3: Threshold Configurability -- changing thresholds changes tier classification.
func TestConfidenceScoring_C3_ThresholdConfigurability(t *testing.T) {
	input := ScoreInput{
		Freshness:        0.60,
		RetrievalQuality: 0.60,
		DomainCoverage:   0.60,
	}

	// With defaults (high=0.7, low=0.4), overall ~0.60 -> MEDIUM
	defaultCfg := DefaultConfig()
	scorer1 := NewScorer(defaultCfg)
	score1 := scorer1.Score(input)
	assert.Equal(t, TierMedium, score1.Tier)

	// Change high threshold to 0.5 -- same inputs now produce HIGH
	modifiedCfg := defaultCfg
	modifiedCfg.Thresholds.HighThreshold = 0.5
	scorer2 := NewScorer(modifiedCfg)
	score2 := scorer2.Score(input)
	assert.Equal(t, TierHigh, score2.Tier)

	// Change low threshold to 0.7 -- same inputs now produce LOW
	modifiedCfg2 := defaultCfg
	modifiedCfg2.Thresholds.LowThreshold = 0.7
	modifiedCfg2.Thresholds.HighThreshold = 0.9
	scorer3 := NewScorer(modifiedCfg2)
	score3 := scorer3.Score(input)
	assert.Equal(t, TierLow, score3.Tier)
}

func TestConfidenceTier_String(t *testing.T) {
	assert.Equal(t, "HIGH", TierHigh.String())
	assert.Equal(t, "MEDIUM", TierMedium.String())
	assert.Equal(t, "LOW", TierLow.String())
	assert.Equal(t, "UNKNOWN", ConfidenceTier(99).String())
}

func TestScore_GeometricMeanProperties(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	// Property: equal inputs with equal weights should produce that value
	cfg := DefaultConfig()
	cfg.Weights = ScoringWeights{Freshness: 1.0, Retrieval: 1.0, Coverage: 1.0}
	equalScorer := NewScorer(cfg)

	score := equalScorer.Score(ScoreInput{
		Freshness: 0.5, RetrievalQuality: 0.5, DomainCoverage: 0.5,
	})
	assert.InDelta(t, 0.5, score.Overall, 0.001)

	// Property: perfect inputs -> overall = 1.0
	score = scorer.Score(ScoreInput{
		Freshness: 1.0, RetrievalQuality: 1.0, DomainCoverage: 1.0,
	})
	assert.InDelta(t, 1.0, score.Overall, 0.001)

	// Property: geometric mean is always <= arithmetic mean
	score = scorer.Score(ScoreInput{
		Freshness: 0.8, RetrievalQuality: 0.4, DomainCoverage: 0.9,
	})
	wf := DefaultConfig().Weights.Freshness
	wr := DefaultConfig().Weights.Retrieval
	wc := DefaultConfig().Weights.Coverage
	wSum := wf + wr + wc
	arithmeticMean := (wf*0.8 + wr*0.4 + wc*0.9) / wSum
	assert.LessOrEqual(t, score.Overall, arithmeticMean)
}

func TestScore_ZeroIntolerance(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	// Any single zero input forces overall to zero
	tests := []struct {
		name string
		f, r, c float64
	}{
		{"zero-freshness", 0.0, 0.9, 0.9},
		{"zero-retrieval", 0.9, 0.0, 0.9},
		{"zero-coverage", 0.9, 0.9, 0.0},
		{"all-zero", 0.0, 0.0, 0.0},
		{"two-zeros", 0.0, 0.0, 0.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.Score(ScoreInput{
				Freshness:        tt.f,
				RetrievalQuality: tt.r,
				DomainCoverage:   tt.c,
			})
			assert.Equal(t, 0.0, score.Overall)
			assert.Equal(t, TierLow, score.Tier)
		})
	}
}

func TestScore_ClampOutOfRange(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	// Values > 1.0 clamped to 1.0; values < 0.0 clamped to 0.0
	score := scorer.Score(ScoreInput{
		Freshness:        1.5,
		RetrievalQuality: -0.1,
		DomainCoverage:   0.8,
	})
	assert.Equal(t, 1.0, score.Freshness)
	assert.Equal(t, 0.0, score.Retrieval) // clamped -> zero -> overall = 0
	assert.Equal(t, 0.0, score.Overall)
	assert.Equal(t, TierLow, score.Tier)
}

func TestScore_EqualWeights(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Weights = ScoringWeights{Freshness: 1.0, Retrieval: 1.0, Coverage: 1.0}
	scorer := NewScorer(cfg)

	// With equal weights, geometric mean of (0.5, 0.5, 0.5) = 0.5
	score := scorer.Score(ScoreInput{
		Freshness: 0.5, RetrievalQuality: 0.5, DomainCoverage: 0.5,
	})
	assert.InDelta(t, 0.5, score.Overall, 0.001)
}

func TestScore_WeightInfluence(t *testing.T) {
	// Higher weight on freshness means freshness has more influence
	cfg := DefaultConfig()
	cfg.Weights = ScoringWeights{Freshness: 0.90, Retrieval: 0.05, Coverage: 0.05}
	scorer := NewScorer(cfg)

	// High freshness, low everything else -> overall pulled toward freshness
	score := scorer.Score(ScoreInput{
		Freshness: 0.9, RetrievalQuality: 0.3, DomainCoverage: 0.3,
	})
	// With heavy freshness weight, overall should be closer to 0.9 than to 0.3
	assert.Greater(t, score.Overall, 0.6)
}

func TestScore_LowTier_HasGapAdmission(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	score := scorer.Score(ScoreInput{
		Freshness:        0.0,
		RetrievalQuality: 0.5,
		DomainCoverage:   0.5,
		MissingDomains:   []string{"kubernetes-migration"},
		StaleDomains: []StaleDomainInfo{{
			QualifiedName:      "autom8y::knossos::test-coverage",
			Domain:             "test-coverage",
			Repo:               "knossos",
			Freshness:          0.1,
			DaysSinceGenerated: 21,
		}},
	})

	require.Equal(t, TierLow, score.Tier)
	require.NotNil(t, score.Gap)
	assert.True(t, score.Gap.HasGaps())
	assert.Len(t, score.Gap.MissingDomains, 1)
	assert.Len(t, score.Gap.StaleDomains, 1)
	assert.Len(t, score.Gap.Suggestions, 2)
}

func TestScore_HighTier_NoGapAdmission(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	score := scorer.Score(ScoreInput{
		Freshness:        0.95,
		RetrievalQuality: 0.85,
		DomainCoverage:   1.0,
	})

	assert.Equal(t, TierHigh, score.Tier)
	assert.Nil(t, score.Gap)
}

func TestScore_MediumTier_NoGapAdmission(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	score := scorer.Score(ScoreInput{
		Freshness:        0.50,
		RetrievalQuality: 0.50,
		DomainCoverage:   1.0,
	})

	assert.Equal(t, TierMedium, score.Tier)
	assert.Nil(t, score.Gap)
}

func TestScore_StoredInputValues(t *testing.T) {
	scorer := NewScorer(DefaultConfig())

	score := scorer.Score(ScoreInput{
		Freshness:        0.75,
		RetrievalQuality: 0.60,
		DomainCoverage:   0.90,
	})

	assert.Equal(t, 0.75, score.Freshness)
	assert.Equal(t, 0.60, score.Retrieval)
	assert.Equal(t, 0.90, score.Coverage)
}

func TestScore_WorkedExample1_FreshArchitecture(t *testing.T) {
	// TDD Example 1: Fresh architecture query, full coverage
	// Architecture half-life is now 14 days (empirical)
	scorer := NewScorer(DefaultConfig())

	// Freshness: exp(-ln(2)/14 * 5) = 0.781
	freshness := math.Exp(-math.Ln2 / 14.0 * 5.0)
	score := scorer.Score(ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: 0.85,
		DomainCoverage:   1.0,
	})

	assert.InDelta(t, 0.861, score.Overall, 0.02)
	assert.Equal(t, TierHigh, score.Tier)
}

func TestScore_WorkedExample2_StaleTestCoverage(t *testing.T) {
	// TDD Example 2: Stale test-coverage query
	// Test-coverage half-life is now 5 days (empirical)
	scorer := NewScorer(DefaultConfig())

	// Freshness: exp(-ln(2)/5 * 10) = 0.25
	freshness := math.Exp(-math.Ln2 / 5.0 * 10.0)
	score := scorer.Score(ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: 0.90,
		DomainCoverage:   1.0,
	})

	assert.InDelta(t, 0.517, score.Overall, 0.02)
	assert.Equal(t, TierMedium, score.Tier)
}

func TestScore_WorkedExample4_CompletelyUnknown(t *testing.T) {
	// TDD Example 4: Completely unknown topic
	scorer := NewScorer(DefaultConfig())

	score := scorer.Score(ScoreInput{
		Freshness:        0.0,
		RetrievalQuality: 0.1,
		DomainCoverage:   0.0,
		MissingDomains:   []string{"kubernetes-migration"},
	})

	assert.Equal(t, 0.0, score.Overall)
	assert.Equal(t, TierLow, score.Tier)
	require.NotNil(t, score.Gap)
	assert.Contains(t, score.Gap.Suggestions[0], "kubernetes-migration")
}

func TestScore_WorkedExample5_ThresholdChange(t *testing.T) {
	// TDD Example 5: Threshold configurability
	// Overall ~0.517 from example 2 inputs (test-coverage half-life now 5d)

	freshness := math.Exp(-math.Ln2 / 5.0 * 10.0) // 0.25
	input := ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: 0.90,
		DomainCoverage:   1.0,
	}

	// Default (high=0.7, low=0.4): MEDIUM
	scorer1 := NewScorer(DefaultConfig())
	score1 := scorer1.Score(input)
	assert.Equal(t, TierMedium, score1.Tier)

	// Changed (high=0.5, low=0.3): HIGH (0.517 >= 0.5)
	cfg2 := DefaultConfig()
	cfg2.Thresholds.HighThreshold = 0.5
	cfg2.Thresholds.LowThreshold = 0.3
	scorer2 := NewScorer(cfg2)
	score2 := scorer2.Score(input)
	assert.Equal(t, TierHigh, score2.Tier)
}

func TestFreshnessFromChain(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "a", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},
		{QualifiedName: "b", GeneratedAt: "2026-02-22T00:00:00Z", Domain: "test-coverage"},
	}, &cfg.Decay, now)

	freshness := FreshnessFromChain(&chain)
	// Should be the minimum (the stale test-coverage entry)
	assert.Less(t, freshness, 0.1)
}

func TestFreshnessFromChain_NilChain(t *testing.T) {
	assert.Equal(t, 0.0, FreshnessFromChain(nil))
}

func TestFreshnessFromChain_EmptyChain(t *testing.T) {
	chain := ProvenanceChain{}
	assert.Equal(t, 0.0, FreshnessFromChain(&chain))
}

func TestComputeFreshness(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	entries := []ProvenanceLinkInput{
		{QualifiedName: "a", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},
	}

	freshness := ComputeFreshness(entries, &cfg.Decay, now)
	assert.InDelta(t, 1.0, freshness, 0.001)
}

func TestComputeFreshness_EmptyEntries(t *testing.T) {
	cfg := DefaultConfig()
	freshness := ComputeFreshness(nil, &cfg.Decay, time.Now())
	assert.Equal(t, 0.0, freshness)
}

func TestNewScorer_Config(t *testing.T) {
	cfg := DefaultConfig()
	scorer := NewScorer(cfg)
	assert.Equal(t, cfg, scorer.Config())
}

func TestScore_MonotonicallyDecreasing(t *testing.T) {
	// As freshness decreases, overall should decrease (holding others constant)
	scorer := NewScorer(DefaultConfig())

	prev := 2.0
	for f := 100; f >= 0; f -= 5 {
		score := scorer.Score(ScoreInput{
			Freshness:        float64(f) / 100.0,
			RetrievalQuality: 0.8,
			DomainCoverage:   0.9,
		})
		assert.LessOrEqual(t, score.Overall, prev,
			"overall should decrease as freshness decreases (f=%d)", f)
		prev = score.Overall
	}
}

func TestScore_SymmetricWithEqualWeights(t *testing.T) {
	// With equal weights, permuting inputs should produce the same overall
	cfg := DefaultConfig()
	cfg.Weights = ScoringWeights{Freshness: 1.0, Retrieval: 1.0, Coverage: 1.0}
	scorer := NewScorer(cfg)

	s1 := scorer.Score(ScoreInput{Freshness: 0.8, RetrievalQuality: 0.5, DomainCoverage: 0.3})
	s2 := scorer.Score(ScoreInput{Freshness: 0.5, RetrievalQuality: 0.3, DomainCoverage: 0.8})
	s3 := scorer.Score(ScoreInput{Freshness: 0.3, RetrievalQuality: 0.8, DomainCoverage: 0.5})

	assert.InDelta(t, s1.Overall, s2.Overall, 0.001)
	assert.InDelta(t, s2.Overall, s3.Overall, 0.001)
}

// Integration: end-to-end from ProvenanceLinkInput through Score
func TestEndToEnd_ProvenanceThroughScoring(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	// Build provenance chain
	inputs := []ProvenanceLinkInput{
		{
			QualifiedName: "autom8y::knossos::architecture",
			GeneratedAt:   "2026-03-19T00:00:00Z", // 5 days ago
			SourceHash:    "78abb186",
			FilePath:      ".know/architecture.md",
			Domain:        "architecture",
			Repo:          "knossos",
		},
	}
	chain := NewProvenanceChain(inputs, &cfg.Decay, now)
	freshness := FreshnessFromChain(&chain)

	// Score
	scorer := NewScorer(cfg)
	score := scorer.Score(ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: 0.85,
		DomainCoverage:   1.0,
		Chain:            &chain,
	})

	assert.Equal(t, TierHigh, score.Tier)
	assert.Greater(t, score.Overall, 0.7)
	assert.InDelta(t, 0.781, freshness, 0.01) // 5 days on 14-day halflife
}
