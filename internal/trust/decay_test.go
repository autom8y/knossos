package trust

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyDomain(t *testing.T) {
	tests := []struct {
		domain   string
		wantType DomainType
	}{
		{"architecture", DomainArchitecture},
		{"conventions", DomainConventions},
		{"design-constraints", DomainDesignConstraints},
		{"scar-tissue", DomainScarTissue},
		{"test-coverage", DomainTestCoverage},
		{"feat/materialization", DomainFeat},
		{"feat/session-management", DomainFeat},
		{"feat", DomainFeat},
		{"release/platform-profile", DomainRelease},
		{"release/history", DomainRelease},
		{"release", DomainRelease},
		{"literature-agentic-knowledge-retrieval", DomainLiterature},
		{"literature-enterprise-ai-slack-integration", DomainLiterature},
		{"literature/some-topic", DomainLiterature},
		{"something-unknown", DomainUnknown},
		{"custom-domain", DomainUnknown},
		{"", DomainUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := ClassifyDomain(tt.domain)
			assert.Equal(t, tt.wantType, got)
		})
	}
}

func TestDecayConfig_DomainHalfLife(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		domain       string
		wantHalfLife float64
	}{
		{"architecture", 14.0},
		{"conventions", 7.0},
		{"design-constraints", 14.0},
		{"scar-tissue", 10.0},
		{"test-coverage", 5.0},
		{"feat/materialization", 10.0},
		{"release/platform-profile", 3.0},
		{"literature-agentic-knowledge-retrieval", 90.0},
		{"unknown-domain", 7.0}, // default
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := cfg.Decay.DomainHalfLife(tt.domain)
			assert.Equal(t, tt.wantHalfLife, got)
		})
	}
}

func TestDecay_MathematicalProperties(t *testing.T) {
	halfLife := 30.0
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	// Property 1: freshness ~= 1.0 at t=0
	f0 := Decay(base, base, halfLife)
	assert.InDelta(t, 1.0, f0, 0.001)

	// Property 2: freshness ~= 0.5 at t=half_life
	fHL := Decay(base, base.Add(30*24*time.Hour), halfLife)
	assert.InDelta(t, 0.5, fHL, 0.001)

	// Property 3: freshness ~= 0.25 at t=2*half_life
	f2HL := Decay(base, base.Add(60*24*time.Hour), halfLife)
	assert.InDelta(t, 0.25, f2HL, 0.001)

	// Property 4: monotonically decreasing and always positive
	prev := 1.0
	for days := 1; days <= 120; days++ {
		f := Decay(base, base.Add(time.Duration(days)*24*time.Hour), halfLife)
		assert.Less(t, f, prev, "freshness must decrease over time at day %d", days)
		assert.Greater(t, f, 0.0, "freshness must remain positive at day %d", days)
		prev = f
	}

	// Property 5: continuous (no cliff edges)
	for days := 0; days < 120; days++ {
		f1 := Decay(base, base.Add(time.Duration(days)*24*time.Hour), halfLife)
		f2 := Decay(base, base.Add(time.Duration(days+1)*24*time.Hour), halfLife)
		delta := f1 - f2
		assert.Greater(t, delta, 0.0, "each day should decrease freshness")
		assert.Less(t, delta, 0.1, "no single day should produce a >0.1 jump")
	}
}

func TestDecay_AllDomainTypes(t *testing.T) {
	// Verify freshness ~= 1.0 at t=0 for all domain types
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	halfLives := []float64{30.0, 21.0, 60.0, 7.0, 14.0, 180.0}
	for _, hl := range halfLives {
		f := Decay(base, base, hl)
		assert.InDelta(t, 1.0, f, 0.001, "freshness at t=0 for halflife=%f", hl)
	}
}

func TestDecay_ZeroTime(t *testing.T) {
	// Zero time -> freshness 0.0 (fail-safe)
	f := Decay(time.Time{}, time.Now(), 30.0)
	assert.Equal(t, 0.0, f)
}

func TestDecay_FutureTimestamp(t *testing.T) {
	// Future -> clamped to 1.0
	now := time.Now()
	future := now.Add(24 * time.Hour)
	f := Decay(future, now, 30.0)
	assert.Equal(t, 1.0, f)
}

func TestDecay_InvalidHalfLife(t *testing.T) {
	now := time.Now()
	// Zero half-life -> 0.0
	assert.Equal(t, 0.0, Decay(now, now, 0.0))
	// Negative half-life -> 0.0
	assert.Equal(t, 0.0, Decay(now, now, -1.0))
}

func TestDecay_SpecificHalfLifeValues(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		halfLife float64
		days     int
		wantF    float64
		delta    float64
	}{
		{"architecture at 30 days", 30.0, 30, 0.5, 0.001},
		{"test-coverage at 7 days", 7.0, 7, 0.5, 0.001},
		{"test-coverage at 14 days (2x halflife)", 7.0, 14, 0.25, 0.001},
		{"scar-tissue at 60 days", 60.0, 60, 0.5, 0.001},
		{"literature at 180 days", 180.0, 180, 0.5, 0.001},
		{"feat at 14 days", 14.0, 14, 0.5, 0.001},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Decay(base, base.Add(time.Duration(tt.days)*24*time.Hour), tt.halfLife)
			assert.InDelta(t, tt.wantF, f, tt.delta)
		})
	}
}

func TestDecayFromString_Integration(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	// Architecture generated 5 days ago: exp(-ln(2)/14 * 5) ~ 0.781
	f := cfg.Decay.DecayFromString("2026-03-19T00:00:00Z", "architecture", now)
	assert.InDelta(t, 0.781, f, 0.01)

	// Test-coverage generated 10 days ago: exp(-ln(2)/5 * 10) = 0.25
	f = cfg.Decay.DecayFromString("2026-03-14T00:00:00Z", "test-coverage", now)
	assert.InDelta(t, 0.25, f, 0.01)

	// Unparseable timestamp -> 0.0
	f = cfg.Decay.DecayFromString("not-a-date", "architecture", now)
	assert.Equal(t, 0.0, f)

	// Empty timestamp -> 0.0
	f = cfg.Decay.DecayFromString("", "architecture", now)
	assert.Equal(t, 0.0, f)
}

func TestDecayFromString_AllDomainTypes(t *testing.T) {
	cfg := DefaultConfig()
	// Generated exactly at "now" -- all should be ~1.0
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	genAt := "2026-03-24T00:00:00Z"

	domains := []string{
		"architecture", "conventions", "design-constraints",
		"scar-tissue", "test-coverage", "feat/materialization",
		"release/platform-profile", "literature-agentic-knowledge-retrieval",
		"unknown-domain",
	}
	for _, d := range domains {
		t.Run(d, func(t *testing.T) {
			f := cfg.Decay.DecayFromString(genAt, d, now)
			assert.InDelta(t, 1.0, f, 0.001)
		})
	}
}

func TestDecay_SevenDayAndStandardDurationEquivalence(t *testing.T) {
	// "7d" and "168h" should represent the same duration conceptually.
	// The decay model works in days, so we verify that 7 days elapsed
	// produces the same result regardless of how the caller arrived at the time.
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	halfLife := 7.0

	// 7 days via day count
	f1 := Decay(base, base.Add(7*24*time.Hour), halfLife)
	// 168 hours
	f2 := Decay(base, base.Add(168*time.Hour), halfLife)

	assert.Equal(t, f1, f2)
	assert.InDelta(t, 0.5, f1, 0.001)
}

func TestDecay_VeryLargeElapsed(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	// 10 years later
	f := Decay(base, base.Add(3650*24*time.Hour), 30.0)
	assert.Greater(t, f, 0.0, "freshness never reaches zero")
	assert.Less(t, f, 0.001, "freshness is negligible after 10 years")
}

func TestDecay_ExactMathematicalFormula(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	halfLife := 21.0
	days := 10.0

	expected := math.Exp(-math.Ln2 / halfLife * days)
	actual := Decay(base, base.Add(time.Duration(days*24)*time.Hour), halfLife)

	require.InDelta(t, expected, actual, 1e-10, "should match the mathematical formula exactly")
}
