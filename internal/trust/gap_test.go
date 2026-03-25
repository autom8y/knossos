package trust

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C-2: Gap Admission Functionality

func TestGapAdmission_MissingDomain(t *testing.T) {
	gap := NewGapAdmission(
		[]string{"kubernetes-migration"},
		nil,
	)

	assert.True(t, gap.HasGaps())
	assert.False(t, gap.IsEmpty())
	require.Len(t, gap.Suggestions, 1)
	assert.Contains(t, gap.Suggestions[0], "/know --domain=kubernetes-migration")
	assert.Contains(t, gap.Reason, "no knowledge found for")
	assert.Contains(t, gap.Reason, "kubernetes-migration")
}

func TestGapAdmission_StaleDomain(t *testing.T) {
	gap := NewGapAdmission(
		nil,
		[]StaleDomainInfo{{
			QualifiedName:      "autom8y::knossos::test-coverage",
			Domain:             "test-coverage",
			Repo:               "knossos",
			Freshness:          0.10,
			DaysSinceGenerated: 21,
		}},
	)

	assert.True(t, gap.HasGaps())
	require.Len(t, gap.Suggestions, 1)
	assert.Contains(t, gap.Suggestions[0], "/know --domain=test-coverage")
	assert.Contains(t, gap.Suggestions[0], "repo knossos")
	assert.Contains(t, gap.Suggestions[0], "21 days ago")
	assert.Contains(t, gap.Reason, "stale knowledge in")
	assert.Contains(t, gap.Reason, "autom8y::knossos::test-coverage")
}

func TestGapAdmission_MixedMissingAndStale(t *testing.T) {
	gap := NewGapAdmission(
		[]string{"security-policy"},
		[]StaleDomainInfo{{
			QualifiedName:      "autom8y::auth::design-constraints",
			Domain:             "design-constraints",
			Repo:               "auth",
			Freshness:          0.05,
			DaysSinceGenerated: 60,
		}},
	)

	assert.True(t, gap.HasGaps())
	require.Len(t, gap.Suggestions, 2)
	assert.Contains(t, gap.Reason, "no knowledge found for")
	assert.Contains(t, gap.Reason, "stale knowledge in")

	// First suggestion is for missing, second for stale
	assert.Contains(t, gap.Suggestions[0], "security-policy")
	assert.Contains(t, gap.Suggestions[1], "design-constraints")
	assert.Contains(t, gap.Suggestions[1], "repo auth")
}

func TestGapAdmission_Empty(t *testing.T) {
	gap := NewGapAdmission(nil, nil)

	assert.False(t, gap.HasGaps())
	assert.True(t, gap.IsEmpty())
	assert.Empty(t, gap.Suggestions)
	assert.Equal(t, "insufficient knowledge to answer this question reliably", gap.Reason)
}

func TestGapAdmission_MultipleMissingDomains(t *testing.T) {
	gap := NewGapAdmission(
		[]string{"kubernetes-migration", "security-policy", "ci-pipeline"},
		nil,
	)

	require.Len(t, gap.Suggestions, 3)
	for _, s := range gap.Suggestions {
		assert.True(t, strings.Contains(s, "/know --domain="), "suggestion should contain /know command")
	}
	assert.Contains(t, gap.Reason, "kubernetes-migration")
	assert.Contains(t, gap.Reason, "security-policy")
	assert.Contains(t, gap.Reason, "ci-pipeline")
}

func TestGapAdmission_MultipleStaleDomains(t *testing.T) {
	gap := NewGapAdmission(
		nil,
		[]StaleDomainInfo{
			{
				QualifiedName:      "autom8y::knossos::test-coverage",
				Domain:             "test-coverage",
				Repo:               "knossos",
				Freshness:          0.10,
				DaysSinceGenerated: 21,
			},
			{
				QualifiedName:      "autom8y::auth::conventions",
				Domain:             "conventions",
				Repo:               "auth",
				Freshness:          0.20,
				DaysSinceGenerated: 45,
			},
		},
	)

	require.Len(t, gap.Suggestions, 2)
	assert.Contains(t, gap.Suggestions[0], "repo knossos")
	assert.Contains(t, gap.Suggestions[1], "repo auth")
}

func TestSuggestionFor_WithRepo(t *testing.T) {
	s := SuggestionFor("architecture", "knossos")
	assert.Contains(t, s, "/know --domain=architecture")
	assert.Contains(t, s, "repo knossos")
}

func TestSuggestionFor_WithoutRepo(t *testing.T) {
	s := SuggestionFor("architecture", "")
	assert.Contains(t, s, "/know --domain=architecture")
	assert.Contains(t, s, "relevant repository")
	assert.NotContains(t, s, "repo ")
}

func TestSuggestionFor_FeatDomain(t *testing.T) {
	s := SuggestionFor("feat/materialization", "knossos")
	assert.Contains(t, s, "/know --domain=feat/materialization")
	assert.Contains(t, s, "repo knossos")
}

func TestGapAdmission_HasGaps_OnlyMissing(t *testing.T) {
	gap := GapAdmission{MissingDomains: []string{"something"}}
	assert.True(t, gap.HasGaps())
	assert.False(t, gap.IsEmpty())
}

func TestGapAdmission_HasGaps_OnlyStale(t *testing.T) {
	gap := GapAdmission{StaleDomains: []StaleDomainInfo{{QualifiedName: "a::b::c"}}}
	assert.True(t, gap.HasGaps())
	assert.False(t, gap.IsEmpty())
}

func TestGapAdmission_ReasonFormat(t *testing.T) {
	// Mixed: should have both reasons joined by semicolon
	gap := NewGapAdmission(
		[]string{"missing-domain"},
		[]StaleDomainInfo{{QualifiedName: "org::repo::stale", Domain: "stale", Repo: "repo", DaysSinceGenerated: 10}},
	)
	parts := strings.Split(gap.Reason, "; ")
	require.Len(t, parts, 2)
	assert.True(t, strings.HasPrefix(parts[0], "no knowledge found for"))
	assert.True(t, strings.HasPrefix(parts[1], "stale knowledge in"))
}
