package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
