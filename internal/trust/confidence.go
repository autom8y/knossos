package trust

import (
	"math"
	"time"
)

// ConfidenceTier determines response behavior.
type ConfidenceTier int

const (
	// TierHigh: direct answer with sources. No caveats needed.
	TierHigh ConfidenceTier = iota
	// TierMedium: answer with caveats and staleness warnings.
	TierMedium
	// TierLow: refuse to answer. Emit GapAdmission instead.
	TierLow
)

// String returns the human-readable tier name.
func (ct ConfidenceTier) String() string {
	switch ct {
	case TierHigh:
		return "HIGH"
	case TierMedium:
		return "MEDIUM"
	case TierLow:
		return "LOW"
	default:
		return "UNKNOWN"
	}
}

// ConfidenceScore is the composite trust assessment for a Clew response.
// Computed from three input signals and mapped to a tier.
type ConfidenceScore struct {
	// Overall is the composite confidence score in [0.0, 1.0].
	Overall float64

	// Freshness is the temporal decay signal in [0.0, 1.0].
	Freshness float64

	// Retrieval is the search quality signal in [0.0, 1.0].
	Retrieval float64

	// Coverage is the domain coverage signal in [0.0, 1.0].
	Coverage float64

	// Tier is the classified confidence tier (HIGH, MEDIUM, LOW).
	Tier ConfidenceTier

	// Gap is populated when Tier == TierLow.
	// Contains refusal explanation and actionable suggestions.
	// Nil for HIGH and MEDIUM tiers.
	Gap *GapAdmission
}

// Scorer computes ConfidenceScores from input signals.
// Constructed with a TrustConfig; reusable across multiple scorings.
type Scorer struct {
	config TrustConfig
}

// NewScorer creates a Scorer with the given configuration.
func NewScorer(config TrustConfig) *Scorer {
	return &Scorer{config: config}
}

// Config returns the scorer's configuration (for inspection/testing).
func (s *Scorer) Config() TrustConfig {
	return s.config
}

// ScoreInput holds the raw signals for a single confidence scoring operation.
type ScoreInput struct {
	// Freshness: the temporal decay signal. Use MinFreshness or MeanFreshness
	// from the ProvenanceChain, or compute via DecayConfig.
	Freshness float64

	// RetrievalQuality: the search relevance signal from the search layer.
	// A float64 in [0.0, 1.0]. The search layer is responsible for normalizing
	// its internal scores to this range.
	RetrievalQuality float64

	// DomainCoverage: fraction of query-relevant domains found.
	// 1.0 means all requested domains are present; 0.0 means none.
	DomainCoverage float64

	// Chain is the provenance chain for the response.
	Chain *ProvenanceChain

	// MissingDomains are domains the query needs but the registry lacks.
	MissingDomains []string

	// StaleDomains are domains found but below freshness threshold.
	StaleDomains []StaleDomainInfo
}

// Score computes a ConfidenceScore from the input signals.
//
// ## Composite Scoring Algorithm: Weighted Geometric Mean
//
// Formula:
//
//	Overall = (F^wf * R^wr * C^wc) ^ (1 / (wf + wr + wc))
//
// Where:
//
//	F = Freshness (0.0-1.0)
//	R = RetrievalQuality (0.0-1.0)
//	C = DomainCoverage (0.0-1.0)
//	wf, wr, wc = configured weights (default: 0.45, 0.25, 0.30)
//
// Properties:
//   - Zero-intolerance: any zero input produces zero overall
//   - Weighted sensitivity: higher-weighted inputs have more influence
//   - Diminishing returns: improving already-strong signals has less impact
func (s *Scorer) Score(input ScoreInput) ConfidenceScore {
	// Clamp inputs to [0.0, 1.0]
	freshness := clamp01(input.Freshness)
	retrieval := clamp01(input.RetrievalQuality)
	coverage := clamp01(input.DomainCoverage)

	// Compute weighted geometric mean
	wf := s.config.Weights.Freshness
	wr := s.config.Weights.Retrieval
	wc := s.config.Weights.Coverage
	wSum := wf + wr + wc

	var overall float64
	if freshness == 0 || retrieval == 0 || coverage == 0 {
		// Short-circuit: any zero input -> zero overall (geometric mean property)
		overall = 0.0
	} else {
		// Weighted geometric mean via log-space computation for numerical stability:
		// log(G) = (wf*log(F) + wr*log(R) + wc*log(C)) / (wf+wr+wc)
		logG := (wf*math.Log(freshness) + wr*math.Log(retrieval) + wc*math.Log(coverage)) / wSum
		overall = math.Exp(logG)
	}

	// Classify tier
	tier := s.classify(overall)

	score := ConfidenceScore{
		Overall:   overall,
		Freshness: freshness,
		Retrieval: retrieval,
		Coverage:  coverage,
		Tier:      tier,
	}

	// Build GapAdmission for LOW tier
	if tier == TierLow {
		gap := NewGapAdmission(input.MissingDomains, input.StaleDomains)
		score.Gap = &gap
	}

	return score
}

// FreshnessFromChain computes the freshness input from a ProvenanceChain.
// Uses the minimum freshness across all sources (weakest-link model).
// Returns 0.0 for empty or nil chains.
func FreshnessFromChain(chain *ProvenanceChain) float64 {
	if chain == nil {
		return 0.0
	}
	return chain.MinFreshness()
}

// ComputeFreshness is a convenience function that computes the freshness score
// for a set of domain entries. Returns the minimum freshness across all entries.
// This is the recommended way to compute the freshness input for Score().
func ComputeFreshness(entries []ProvenanceLinkInput, decay *DecayConfig, now time.Time) float64 {
	if len(entries) == 0 {
		return 0.0
	}
	chain := NewProvenanceChain(entries, decay, now)
	return FreshnessFromChain(&chain)
}

// classify maps an overall score to a ConfidenceTier.
func (s *Scorer) classify(overall float64) ConfidenceTier {
	switch {
	case overall >= s.config.Thresholds.HighThreshold:
		return TierHigh
	case overall >= s.config.Thresholds.LowThreshold:
		return TierMedium
	default:
		return TierLow
	}
}

// clamp01 constrains a value to the [0.0, 1.0] range.
func clamp01(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}
