package trust

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C-4: Provenance Chain Integrity

func TestNewProvenanceChain_TracesToKnowFiles(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC)

	inputs := []ProvenanceLinkInput{
		{
			QualifiedName: "autom8y::knossos::architecture",
			GeneratedAt:   "2026-03-23T18:00:00Z",
			SourceHash:    "78abb186",
			FilePath:      ".know/architecture.md",
			Domain:        "architecture",
			Repo:          "knossos",
		},
		{
			QualifiedName: "autom8y::knossos::conventions",
			GeneratedAt:   "2026-03-23T18:00:00Z",
			SourceHash:    "78abb186",
			FilePath:      ".know/conventions.md",
			Domain:        "conventions",
			Repo:          "knossos",
		},
	}

	chain := NewProvenanceChain(inputs, &cfg.Decay, now)

	// Verify all C-4 required fields
	assert.Equal(t, 2, chain.Len())
	assert.False(t, chain.IsEmpty())

	for _, link := range chain.Sources {
		assert.NotEmpty(t, link.QualifiedName)
		assert.False(t, link.GeneratedAt.IsZero())
		assert.NotEmpty(t, link.SourceHash)
		assert.NotEmpty(t, link.FilePath)
		assert.Greater(t, link.FreshnessAtQuery, 0.0)
		assert.LessOrEqual(t, link.FreshnessAtQuery, 1.0)
	}

	// Verify data is from actual input, not fabricated
	assert.Equal(t, "autom8y::knossos::architecture", chain.Sources[0].QualifiedName)
	assert.Equal(t, ".know/architecture.md", chain.Sources[0].FilePath)
	assert.Equal(t, "78abb186", chain.Sources[0].SourceHash)
	assert.Equal(t, "architecture", chain.Sources[0].Domain)
	assert.Equal(t, "knossos", chain.Sources[0].Repo)
}

func TestNewProvenanceChain_UnparseableTimestamp(t *testing.T) {
	cfg := DefaultConfig()
	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "org::repo::domain", GeneratedAt: "not-a-timestamp", SourceHash: "abc"},
	}, &cfg.Decay, time.Now())

	// Link is included (provenance still valid) but freshness is 0.0
	require.Equal(t, 1, chain.Len())
	assert.Equal(t, 0.0, chain.Sources[0].FreshnessAtQuery)
	assert.True(t, chain.Sources[0].GeneratedAt.IsZero())
}

func TestNewProvenanceChain_EmptyInputs(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()
	chain := NewProvenanceChain(nil, &cfg.Decay, now)

	assert.True(t, chain.IsEmpty())
	assert.Equal(t, 0, chain.Len())
	assert.Equal(t, now, chain.BuiltAt)
}

func TestNewProvenanceChain_WithExcerpt(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{
			QualifiedName: "autom8y::knossos::architecture",
			GeneratedAt:   "2026-03-24T00:00:00Z",
			SourceHash:    "abc123",
			FilePath:      ".know/architecture.md",
			Domain:        "architecture",
			Repo:          "knossos",
			Excerpt:       "## Package Structure",
		},
	}, &cfg.Decay, now)

	require.Equal(t, 1, chain.Len())
	assert.Equal(t, "## Package Structure", chain.Sources[0].Excerpt)
}

func TestProvenanceChain_QualifiedNames(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "autom8y::knossos::architecture", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},
		{QualifiedName: "autom8y::knossos::conventions", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "conventions"},
		{QualifiedName: "autom8y::auth::design-constraints", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "design-constraints"},
	}, &cfg.Decay, now)

	names := chain.QualifiedNames()
	assert.Equal(t, []string{
		"autom8y::knossos::architecture",
		"autom8y::knossos::conventions",
		"autom8y::auth::design-constraints",
	}, names)
}

func TestProvenanceChain_MinFreshness(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	// One fresh, one stale
	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "a", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},  // ~1.0
		{QualifiedName: "b", GeneratedAt: "2026-02-22T00:00:00Z", Domain: "test-coverage"}, // very stale
	}, &cfg.Decay, now)

	min := chain.MinFreshness()
	// The test-coverage entry is 30 days old with 7-day halflife: ~0.0625
	assert.Less(t, min, 0.1, "min should be the stale entry")
}

func TestProvenanceChain_MinFreshness_Empty(t *testing.T) {
	chain := ProvenanceChain{}
	assert.Equal(t, 0.0, chain.MinFreshness())
}

func TestProvenanceChain_MeanFreshness(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "a", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},
		{QualifiedName: "b", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "conventions"},
	}, &cfg.Decay, now)

	mean := chain.MeanFreshness()
	// Both generated at "now", so both ~1.0, mean ~1.0
	assert.InDelta(t, 1.0, mean, 0.01)
}

func TestProvenanceChain_MeanFreshness_Empty(t *testing.T) {
	chain := ProvenanceChain{}
	assert.Equal(t, 0.0, chain.MeanFreshness())
}

func TestProvenanceChain_MeanFreshness_Mixed(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "a", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},  // ~1.0
		{QualifiedName: "b", GeneratedAt: "not-valid", Domain: "conventions"},                // 0.0
	}, &cfg.Decay, now)

	mean := chain.MeanFreshness()
	// (1.0 + 0.0) / 2 = 0.5
	assert.InDelta(t, 0.5, mean, 0.01)
}

func TestProvenanceChain_StaleSources(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "fresh", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},
		{QualifiedName: "stale", GeneratedAt: "not-valid", Domain: "conventions"},
	}, &cfg.Decay, now)

	stale := chain.StaleSources(0.5)
	require.Len(t, stale, 1)
	assert.Equal(t, "stale", stale[0].QualifiedName)
}

func TestProvenanceChain_StaleSources_NoneStale(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{QualifiedName: "a", GeneratedAt: "2026-03-24T00:00:00Z", Domain: "architecture"},
	}, &cfg.Decay, now)

	stale := chain.StaleSources(0.5)
	assert.Empty(t, stale)
}

func TestProvenanceChain_SingleSource(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	chain := NewProvenanceChain([]ProvenanceLinkInput{
		{
			QualifiedName: "autom8y::knossos::scar-tissue",
			GeneratedAt:   "2026-03-24T00:00:00Z",
			SourceHash:    "deadbeef",
			FilePath:      ".know/scar-tissue.md",
			Domain:        "scar-tissue",
			Repo:          "knossos",
		},
	}, &cfg.Decay, now)

	require.Equal(t, 1, chain.Len())
	assert.InDelta(t, 1.0, chain.MinFreshness(), 0.001)
	assert.InDelta(t, 1.0, chain.MeanFreshness(), 0.001)
	assert.Equal(t, chain.MinFreshness(), chain.MeanFreshness())
}

func TestProvenanceChain_BuiltAtTimestamp(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Date(2026, 3, 24, 15, 30, 0, 0, time.UTC)

	chain := NewProvenanceChain(nil, &cfg.Decay, now)
	assert.Equal(t, now, chain.BuiltAt)
}
