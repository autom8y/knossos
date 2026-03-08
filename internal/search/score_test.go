package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Levenshtein ---

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"identical", "rite", "rite", 0},
		{"one substitution", "rite", "ryte", 1},
		{"one insertion", "rite", "rites", 1},
		{"one deletion", "rite", "rit", 1},
		{"empty a", "", "abc", 3},
		{"both empty", "", "", 0},
		{"completely different", "abc", "xyz", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Levenshtein(tt.a, tt.b))
		})
	}
}

// --- MinOf3 ---

func TestMinOf3(t *testing.T) {
	assert.Equal(t, 1, MinOf3(1, 2, 3))
	assert.Equal(t, 1, MinOf3(3, 2, 1))
	assert.Equal(t, 1, MinOf3(2, 1, 3))
	assert.Equal(t, 0, MinOf3(0, 0, 0))
	assert.Equal(t, -1, MinOf3(-1, 0, 1))
}

// --- tokenize ---

func TestTokenize(t *testing.T) {
	tokens := tokenize("sync pipeline quickly")
	assert.Equal(t, []string{"sync", "pipeline", "quickly"}, tokens)
}

func TestTokenizeStopWords(t *testing.T) {
	// All stop words should be filtered.
	tokens := tokenize("how do i use the ari sync command")
	// "how", "do", "i", "use", "the" are stop words; remaining: "ari", "sync", "command"
	assert.Equal(t, []string{"ari", "sync", "command"}, tokens)
}

func TestTokenizeEmpty(t *testing.T) {
	assert.Empty(t, tokenize(""))
	assert.Empty(t, tokenize("the and or"))
}

func TestTokenizeLowercases(t *testing.T) {
	tokens := tokenize("Sync Pipeline")
	assert.Equal(t, []string{"sync", "pipeline"}, tokens)
}

// --- extractKeywords ---

func TestExtractKeywordsTriggers(t *testing.T) {
	desc := "Triggers: coordinate, orchestrate, multi-phase"
	kw := extractKeywords(desc)
	assert.Contains(t, kw, "coordinate")
	assert.Contains(t, kw, "orchestrate")
	assert.Contains(t, kw, "multi-phase")
}

func TestExtractKeywordsUseWhen(t *testing.T) {
	desc := "Use when: gap analysis needed, cross-component work"
	kw := extractKeywords(desc)
	assert.Contains(t, kw, "gap analysis needed")
	assert.Contains(t, kw, "cross-component work")
}

func TestExtractKeywordsSignificantWords(t *testing.T) {
	// Words > 3 chars that aren't stop words become secondary keywords.
	desc := "Pipeline orchestration workflow"
	kw := extractKeywords(desc)
	assert.Contains(t, kw, "pipeline")
	assert.Contains(t, kw, "orchestration")
	assert.Contains(t, kw, "workflow")
}

func TestExtractKeywordsDeduplication(t *testing.T) {
	desc := "Triggers: sync, sync\nUse when: sync"
	kw := extractKeywords(desc)
	count := 0
	for _, k := range kw {
		if k == "sync" {
			count++
		}
	}
	assert.Equal(t, 1, count, "sync should appear exactly once")
}

func TestExtractKeywordsEmpty(t *testing.T) {
	assert.Empty(t, extractKeywords(""))
}

// --- scoreEntry ---

func TestScoreExactMatch(t *testing.T) {
	e := SearchEntry{Name: "session", Domain: DomainCommand, Summary: "Manage sessions"}
	score, matchType := scoreEntry("session", e)
	assert.Equal(t, 1000, score)
	assert.Equal(t, "exact", matchType)
}

func TestScoreExactMatchCaseInsensitive(t *testing.T) {
	e := SearchEntry{Name: "session", Domain: DomainCommand}
	score, matchType := scoreEntry("SESSION", e)
	assert.Equal(t, 1000, score)
	assert.Equal(t, "exact", matchType)
}

func TestScoreExactMatchAlias(t *testing.T) {
	e := SearchEntry{
		Name:    "session",
		Domain:  DomainCommand,
		Aliases: []string{"sessions"},
	}
	score, matchType := scoreEntry("sessions", e)
	assert.Equal(t, 1000, score)
	assert.Equal(t, "exact", matchType)
}

func TestScorePrefixMatch(t *testing.T) {
	e := SearchEntry{Name: "session", Domain: DomainCommand, Summary: "Manage sessions"}
	score, matchType := scoreEntry("sess", e)
	assert.Equal(t, 500, score)
	assert.Equal(t, "prefix", matchType)
}

func TestScorePrefixMatchCaseInsensitive(t *testing.T) {
	e := SearchEntry{Name: "session", Domain: DomainCommand}
	score, matchType := scoreEntry("SESS", e)
	assert.Equal(t, 500, score)
	assert.Equal(t, "prefix", matchType)
}

func TestScoreKeywordMatch(t *testing.T) {
	e := SearchEntry{
		Name:    "session-create",
		Domain:  DomainCommand,
		Summary: "Create a new session",
		Keywords: []string{"create"},
	}
	// "create" matches keyword (+150) and name word (+120) and summary (+100).
	score, matchType := scoreEntry("create", e)
	assert.Equal(t, "keyword", matchType)
	// At minimum keyword hit: 150 + 120 + 100 = 370, plus all-match bonus 100 = 470.
	assert.Greater(t, score, 0)
}

func TestScoreKeywordAllMatchBonus(t *testing.T) {
	// Use a name that won't trigger exact or prefix match for the multi-token query.
	e := SearchEntry{
		Name:        "resource-sync",
		Domain:      DomainCommand,
		Summary:     "Sync the pipeline artifacts",
		Description: "Synchronizes pipeline resources",
	}
	// Single token "pipeline" → keyword match, no all-match bonus possible with 1 token.
	// Two tokens "sync pipeline" both match → all-match bonus (+100) added.
	score1, _ := scoreEntry("pipeline", e)
	score2, _ := scoreEntry("sync pipeline", e)
	// Two matching tokens should produce a higher score than one.
	assert.Greater(t, score2, score1)
}

func TestScoreFuzzyMatch(t *testing.T) {
	e := SearchEntry{Name: "session", Domain: DomainCommand}
	// "sesion" is distance 1 from "session".
	score, matchType := scoreEntry("sesion", e)
	assert.Equal(t, "fuzzy", matchType)
	assert.Equal(t, 250, score) // 300 - 50*1
}

func TestScoreFuzzyMatchTooFar(t *testing.T) {
	e := SearchEntry{Name: "rite", Domain: DomainCommand}
	// "xyzabc" is far from "rite" — should not match.
	score, matchType := scoreEntry("xyzabc", e)
	assert.Equal(t, 0, score)
	assert.Equal(t, "none", matchType)
}

func TestScoreContextBoost(t *testing.T) {
	boosted := SearchEntry{Name: "session", Domain: DomainRite, Boosted: true}
	plain := SearchEntry{Name: "session", Domain: DomainRite, Boosted: false}

	scoreBoosted, _ := scoreEntry("session", boosted)
	scorePlain, _ := scoreEntry("session", plain)

	assert.Equal(t, scorePlain+200, scoreBoosted)
}

func TestScoreNoMatch(t *testing.T) {
	e := SearchEntry{
		Name:        "rite",
		Domain:      DomainConcept,
		Summary:     "A rite is a workflow context",
		Description: "Rites define the active workflow",
	}
	score, matchType := scoreEntry("zygote", e)
	assert.Equal(t, 0, score)
	assert.Equal(t, "none", matchType)
}
