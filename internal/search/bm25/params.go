// Package bm25 implements BM25 text retrieval and search-level freshness decay
// for cross-repo knowledge search. Parameters are empirically validated from the
// Sprint-2 parameter sweep (100 combinations: 5 k1 x 4 b x 5 RRF-k, 187-doc corpus).
package bm25

// BM25 scoring parameters.
const (
	// BM25K1 is the term frequency saturation parameter.
	// Higher values increase the contribution of high-frequency terms.
	// Validated at 1.2 — higher values do not improve P@5, lower values slightly worse.
	BM25K1 = 1.2

	// BM25B is the document length normalization parameter.
	// Lower than standard 0.75 because .know/ documents have high length variance
	// (25-450 lines) and the most important documents are the longest.
	// b=0.25 consistently outperforms b=0.75 across all k1 values.
	BM25B = 0.25

	// RRFConstK is the Reciprocal Rank Fusion smoothing constant.
	// Standard range; lower k gives more weight to top-ranked results.
	RRFConstK = 40.0
)

// Decay half-life constants in days, empirically calibrated from a 187-document corpus.
// All values are shorter than academic benchmarks — conservative bias per Decision #14:
// "wrong answers with confidence are worse than refusals."
const (
	DecayArchitecture      = 14.0 // Structural knowledge, moderate decay
	DecayConventions       = 7.0  // Practices evolve each sprint
	DecayScarTissue        = 10.0 // Lessons age as code changes
	DecayDesignConstraints = 14.0 // Architectural constraints persist
	DecayTestCoverage      = 5.0  // Coverage changes with every PR
	DecayFeat              = 10.0 // Feature docs track sprint cadence
	DecayRelease           = 3.0  // Release info is highly time-sensitive
	DecayLiterature        = 90.0 // Literature reviews are write-once-read-many
	DecayRadar             = 7.0  // Opportunities are time-sensitive
	DecayUnknown           = 7.0  // Conservative default for unclassified domains
)
