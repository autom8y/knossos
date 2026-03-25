package bm25

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchFreshness_FreshAtZero(t *testing.T) {
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	f := SearchFreshness("2026-03-24T00:00:00Z", "architecture", now)
	assert.InDelta(t, 1.0, f, 0.001)
}

func TestSearchFreshness_HalfLifeReached(t *testing.T) {
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		domain   string
		daysAgo  int
		halfLife float64
	}{
		{"architecture 14d", "architecture", 14, DecayArchitecture},
		{"conventions 7d", "conventions", 7, DecayConventions},
		{"test-coverage 5d", "test-coverage", 5, DecayTestCoverage},
		{"release 3d", "release/history", 3, DecayRelease},
		{"literature 90d", "literature-agentic", 90, DecayLiterature},
		{"feat 10d", "feat/session", 10, DecayFeat},
		{"scar-tissue 10d", "scar-tissue", 10, DecayScarTissue},
		{"design-constraints 14d", "design-constraints", 14, DecayDesignConstraints},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genAt := now.Add(-time.Duration(tt.daysAgo) * 24 * time.Hour).Format(time.RFC3339)
			f := SearchFreshness(genAt, tt.domain, now)
			assert.InDelta(t, 0.5, f, 0.001, "expected ~0.5 at half-life for %s", tt.domain)
		})
	}
}

func TestSearchFreshness_EmptyTimestamp_Optimistic(t *testing.T) {
	// D-2: Search-level decay returns 1.0 for empty timestamps
	f := SearchFreshness("", "architecture", time.Now())
	assert.Equal(t, 1.0, f)
}

func TestSearchFreshness_UnparseableTimestamp_Optimistic(t *testing.T) {
	// D-2: Search-level decay returns 1.0 for unparseable timestamps
	f := SearchFreshness("not-a-date", "architecture", time.Now())
	assert.Equal(t, 1.0, f)
}

func TestSearchFreshness_FutureTimestamp(t *testing.T) {
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	f := SearchFreshness("2026-04-01T00:00:00Z", "architecture", now)
	assert.Equal(t, 1.0, f)
}

func TestSearchFreshness_DomainClassification(t *testing.T) {
	// Verify each domain type resolves to the correct half-life by checking
	// that a document aged exactly at the domain's half-life produces ~0.5.
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		domain      string
		expectedHL  float64
	}{
		{"architecture", DecayArchitecture},
		{"conventions", DecayConventions},
		{"design-constraints", DecayDesignConstraints},
		{"scar-tissue", DecayScarTissue},
		{"test-coverage", DecayTestCoverage},
		{"feat/materialization", DecayFeat},
		{"release/platform-profile", DecayRelease},
		{"literature-agentic-knowledge-retrieval", DecayLiterature},
		{"unknown-domain", DecayUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			// Generate timestamp at exactly the half-life in the past
			genAt := now.Add(-time.Duration(tt.expectedHL*24) * time.Hour).Format(time.RFC3339)
			f := SearchFreshness(genAt, tt.domain, now)
			assert.InDelta(t, 0.5, f, 0.001,
				"domain %s should have half-life %.0f days", tt.domain, tt.expectedHL)
		})
	}
}

func TestSearchFreshness_MonotonicallyDecreasing(t *testing.T) {
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	prev := 2.0
	for days := 0; days <= 90; days++ {
		genAt := now.Add(-time.Duration(days) * 24 * time.Hour).Format(time.RFC3339)
		f := SearchFreshness(genAt, "architecture", now)
		assert.LessOrEqual(t, f, prev, "freshness must be non-increasing at day %d", days)
		assert.GreaterOrEqual(t, f, 0.0, "freshness must be non-negative at day %d", days)
		prev = f
	}
}
