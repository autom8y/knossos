package bm25

import (
	"math"
	"time"

	"github.com/autom8y/knossos/internal/trust"
)

// halfLifeForDomain returns the search-level half-life for a domain.
// Uses trust.ClassifyDomain() as the single source of domain classification (D-6),
// then maps the result to the empirically calibrated half-life constants.
func halfLifeForDomain(domain string) float64 {
	dt := trust.ClassifyDomain(domain)
	switch dt {
	case trust.DomainArchitecture:
		return DecayArchitecture
	case trust.DomainConventions:
		return DecayConventions
	case trust.DomainDesignConstraints:
		return DecayDesignConstraints
	case trust.DomainScarTissue:
		return DecayScarTissue
	case trust.DomainTestCoverage:
		return DecayTestCoverage
	case trust.DomainFeat:
		return DecayFeat
	case trust.DomainRelease:
		return DecayRelease
	case trust.DomainLiterature:
		return DecayLiterature
	default:
		return DecayUnknown
	}
}

// SearchFreshness computes the search-level freshness score for a domain document.
//
// Mathematical model:
//
//	freshness(t) = exp(-ln(2) / halfLife * elapsedDays)
//
// This is the same exponential decay formula as trust.Decay(), but with different
// edge-case behavior per D-2 (split fail-safe by pipeline stage):
//
//   - Search level (this function): unparseable timestamp returns 1.0 (OPTIMISTIC).
//     At the search stage, the goal is retrieval coverage. An unparseable timestamp
//     should not silently remove a document from search results.
//
//   - Trust level (trust.DecayFromString): unparseable timestamp returns 0.0 (PESSIMISTIC).
//     At the trust stage, unknown freshness = uncertain provenance.
//
// The user sees both signals: the document appears in results (search lets it through),
// but its confidence score is penalized (trust flags the unknown freshness).
func SearchFreshness(generatedAtStr string, domain string, now time.Time) float64 {
	if generatedAtStr == "" {
		return 1.0 // optimistic: prefer showing results
	}

	generatedAt, err := time.Parse(time.RFC3339, generatedAtStr)
	if err != nil {
		return 1.0 // optimistic: unparseable -> treat as fresh for search ranking
	}

	elapsed := now.Sub(generatedAt)
	if elapsed < 0 {
		return 1.0 // future timestamp: fully fresh
	}

	halfLife := halfLifeForDomain(domain)
	if halfLife <= 0 {
		return 1.0 // invalid config: fail-safe to fresh
	}

	elapsedDays := elapsed.Hours() / 24.0
	lambda := math.Ln2 / halfLife
	return math.Exp(-lambda * elapsedDays)
}
