// Package trust provides the confidence model, provenance chain, and gap admission
// infrastructure for the Clew organizational intelligence system.
//
// Every Clew response must carry a trust assessment: a ConfidenceScore composed from
// temporal freshness, retrieval quality, and domain coverage, mapped to a ConfidenceTier
// (HIGH/MEDIUM/LOW) that determines response behavior.
//
// This package is a pure domain library. It does not import internal/search/ (to avoid
// bidirectional dependencies) or internal/cmd/* (to respect layer boundaries). It MAY
// import internal/know/ for domain type classification and duration parsing.
//
// Usage:
//
//	cfg := trust.DefaultConfig()
//	scorer := trust.NewScorer(cfg)
//	score := scorer.Score(trust.ScoreInput{
//	    Freshness:        0.85,
//	    RetrievalQuality: 0.72,
//	    DomainCoverage:   1.0,
//	})
//	// score.Tier == trust.TierHigh
package trust
