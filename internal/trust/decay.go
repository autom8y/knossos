package trust

import (
	"math"
	"strings"
	"time"
)

// DomainType classifies a .know/ domain for decay rate lookup.
type DomainType string

const (
	DomainArchitecture      DomainType = "architecture"
	DomainConventions       DomainType = "conventions"
	DomainDesignConstraints DomainType = "design-constraints"
	DomainScarTissue        DomainType = "scar-tissue"
	DomainTestCoverage      DomainType = "test-coverage"
	DomainFeat              DomainType = "feat"
	DomainRelease           DomainType = "release"
	DomainLiterature        DomainType = "literature"
	DomainUnknown           DomainType = "unknown"
)

// DecayConfig holds domain-type-specific half-lives for the exponential decay model.
// All durations are in days. Every value is configurable via TrustConfig.
type DecayConfig struct {
	// HalfLives maps DomainType to half-life in days.
	// Missing entries fall back to DefaultHalfLifeDays.
	HalfLives map[DomainType]float64 `yaml:"half_lives"`

	// DefaultHalfLifeDays is the fallback half-life for domains not in HalfLives.
	DefaultHalfLifeDays float64 `yaml:"default_half_life_days"`
}

// ClassifyDomain determines the DomainType from a domain name string.
// Handles sub-namespaced domains: "feat/materialization" -> DomainFeat,
// "release/platform-profile" -> DomainRelease, "literature-*" -> DomainLiterature.
func ClassifyDomain(domain string) DomainType {
	switch {
	case strings.HasPrefix(domain, "feat/") || domain == "feat":
		return DomainFeat
	case strings.HasPrefix(domain, "release/") || domain == "release":
		return DomainRelease
	case strings.HasPrefix(domain, "literature-") || strings.HasPrefix(domain, "literature/"):
		return DomainLiterature
	case domain == "architecture":
		return DomainArchitecture
	case domain == "conventions":
		return DomainConventions
	case domain == "design-constraints":
		return DomainDesignConstraints
	case domain == "scar-tissue":
		return DomainScarTissue
	case domain == "test-coverage":
		return DomainTestCoverage
	default:
		return DomainUnknown
	}
}

// DomainHalfLife returns the half-life in days for a given domain.
// Uses the domain name to classify, then looks up the half-life.
// Falls back to DefaultHalfLifeDays if the domain type is not in the map.
func (dc *DecayConfig) DomainHalfLife(domain string) float64 {
	domainType := ClassifyDomain(domain)
	if hl, ok := dc.HalfLives[domainType]; ok {
		return hl
	}
	return dc.DefaultHalfLifeDays
}

// Decay computes the freshness score for a domain given its generation time.
//
// Mathematical model:
//
//	freshness(t) = exp(-ln(2) / half_life * t)
//
// Where:
//
//	t = elapsed time since generatedAt (in fractional days)
//	half_life = domain-type-specific half-life in days
//
// Properties:
//   - freshness(0) = 1.0 (just generated)
//   - freshness(half_life) = 0.5 (half-life reached)
//   - freshness(2 * half_life) ~= 0.25
//   - freshness is continuous, monotonically decreasing, asymptotically approaching 0
//
// Edge cases:
//   - generatedAt is zero value: returns 0.0 (fail-safe per DomainEntry.IsStale pattern)
//   - generatedAt is in the future: returns 1.0 (clamped, not penalized)
//   - half_life <= 0: returns 0.0 (invalid config, fail-safe)
func Decay(generatedAt time.Time, now time.Time, halfLifeDays float64) float64 {
	if generatedAt.IsZero() {
		return 0.0
	}
	if halfLifeDays <= 0 {
		return 0.0
	}

	elapsed := now.Sub(generatedAt)
	if elapsed < 0 {
		return 1.0 // future timestamp: clamp to fresh
	}

	elapsedDays := elapsed.Hours() / 24.0
	lambda := math.Ln2 / halfLifeDays
	return math.Exp(-lambda * elapsedDays)
}

// DecayFromString computes freshness from an RFC3339 timestamp string and domain name.
// Returns 0.0 if the timestamp cannot be parsed (fail-safe: unparseable = maximally stale).
// This is the primary entry point for consumers that have string-typed timestamps
// (e.g., DomainEntry.GeneratedAt).
func (dc *DecayConfig) DecayFromString(generatedAtStr string, domain string, now time.Time) float64 {
	generatedAt, err := time.Parse(time.RFC3339, generatedAtStr)
	if err != nil {
		return 0.0 // fail-safe: unparseable -> maximally stale
	}
	halfLife := dc.DomainHalfLife(domain)
	return Decay(generatedAt, now, halfLife)
}
