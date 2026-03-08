package search

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/paths"
)

// buildTestRoot is defined in collectors_test.go (same package).

// minimalCobraRoot returns a minimal Cobra command tree for testing.
func minimalCobraRoot() *cobra.Command {
	root := &cobra.Command{Use: "ari", Short: "Ariadne"}
	root.AddCommand(
		&cobra.Command{Use: "sync", Short: "Sync resources"},
		&cobra.Command{Use: "explain", Short: "Explain a concept"},
	)
	return root
}

// --- Build ---

func TestBuildWithoutProject(t *testing.T) {
	root := minimalCobraRoot()
	idx := Build(root, nil)
	require.NotNil(t, idx)

	// Should have commands + 13 concepts at minimum.
	assert.Greater(t, len(idx.entries), 13)

	// Verify command domain present.
	hasCommand := false
	hasConcept := false
	for _, e := range idx.entries {
		if e.Domain == DomainCommand {
			hasCommand = true
		}
		if e.Domain == DomainConcept {
			hasConcept = true
		}
	}
	assert.True(t, hasCommand, "should have command entries")
	assert.True(t, hasConcept, "should have concept entries")
}

func TestBuildWithEmptyProjectRoot(t *testing.T) {
	root := minimalCobraRoot()
	resolver := paths.NewResolver("")
	idx := Build(root, resolver)
	require.NotNil(t, idx)

	// Empty project root — no project-scoped entries, but commands and concepts present.
	hasConcept := false
	for _, e := range idx.entries {
		hasConcept = hasConcept || e.Domain == DomainConcept
		// No rite/agent/dromena/routing entries expected.
		assert.NotEqual(t, DomainRite, e.Domain)
		assert.NotEqual(t, DomainAgent, e.Domain)
		assert.NotEqual(t, DomainDromena, e.Domain)
	}
	assert.True(t, hasConcept)
}

func TestBuildWithProject(t *testing.T) {
	root := minimalCobraRoot()
	projectRoot := buildTestRoot(t)
	resolver := paths.NewResolver(projectRoot)
	idx := Build(root, resolver)
	require.NotNil(t, idx)

	// Should succeed even with empty rites/agents dirs.
	assert.NotNil(t, idx.entries)
}

// --- Search ---

func TestSearchExactMatch(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session management"},
			{Name: "rite", Domain: DomainConcept, Summary: "Rite workflow"},
		},
	}
	results := idx.Search("session", SearchOptions{Limit: 5})
	require.NotEmpty(t, results)
	assert.Equal(t, "session", results[0].Name)
	assert.Equal(t, "exact", results[0].MatchType)
	assert.Equal(t, 1000, results[0].Score)
}

func TestSearchNoResults(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session management"},
		},
	}
	results := idx.Search("zygote", SearchOptions{Limit: 5})
	assert.Empty(t, results)
}

func TestSearchSortedByScore(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			// prefix match for "sess"
			{Name: "session", Domain: DomainConcept, Summary: "Manage sessions"},
			// keyword match for "sess" in summary
			{Name: "other", Domain: DomainConcept, Summary: "sess related info"},
		},
	}
	results := idx.Search("sess", SearchOptions{Limit: 5})
	require.Len(t, results, 2)
	// Prefix match (500) should beat keyword match.
	assert.Equal(t, "session", results[0].Name)
	assert.GreaterOrEqual(t, results[0].Score, results[1].Score)
}

func TestSearchDefaultLimit(t *testing.T) {
	entries := make([]SearchEntry, 20)
	for i := range entries {
		entries[i] = SearchEntry{
			Name:    "session",
			Domain:  DomainConcept,
			Summary: "Manages session lifecycle",
		}
	}
	idx := &SearchIndex{entries: entries}
	// Limit 0 → use DefaultLimit (5).
	results := idx.Search("session", SearchOptions{Limit: 0})
	assert.Len(t, results, DefaultLimit)
}

func TestSearchCustomLimit(t *testing.T) {
	entries := make([]SearchEntry, 10)
	for i := range entries {
		entries[i] = SearchEntry{
			Name:    "session",
			Domain:  DomainConcept,
			Summary: "Manages session lifecycle",
		}
	}
	idx := &SearchIndex{entries: entries}
	results := idx.Search("session", SearchOptions{Limit: 3})
	assert.Len(t, results, 3)
}

func TestSearchDomainFilter(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session concept"},
			{Name: "session", Domain: DomainCommand, Summary: "Session command"},
			{Name: "session", Domain: DomainRite, Summary: "Session rite"},
		},
	}
	results := idx.Search("session", SearchOptions{
		Limit:   10,
		Domains: []Domain{DomainCommand},
	})
	require.Len(t, results, 1)
	assert.Equal(t, DomainCommand, results[0].Domain)
}

func TestSearchDomainFilterMultiple(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session concept"},
			{Name: "session", Domain: DomainCommand, Summary: "Session command"},
			{Name: "session", Domain: DomainRite, Summary: "Session rite"},
		},
	}
	results := idx.Search("session", SearchOptions{
		Limit:   10,
		Domains: []Domain{DomainCommand, DomainConcept},
	})
	assert.Len(t, results, 2)
	for _, r := range results {
		assert.NotEqual(t, DomainRite, r.Domain)
	}
}

func TestSearchExcludesZeroScore(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session management"},
			// This entry should not match "session" at all.
			{Name: "zygote-xylophone", Domain: DomainConcept, Summary: "Unrelated entry"},
		},
	}
	results := idx.Search("session", SearchOptions{Limit: 10})
	for _, r := range results {
		assert.Greater(t, r.Score, 0)
	}
}

// TestBuildIntegration runs Build against real CLI and concept data (no project).
func TestBuildIntegration(t *testing.T) {
	root := &cobra.Command{Use: "ari", Short: "Ariadne CLI"}
	root.AddCommand(&cobra.Command{Use: "sync", Short: "Sync resources to .claude/"})
	root.AddCommand(&cobra.Command{Use: "explain", Short: "Explain a knossos concept"})

	idx := Build(root, nil)
	require.NotNil(t, idx)

	// Search for "sync" — should find the sync command.
	results := idx.Search("sync", SearchOptions{Limit: 5})
	require.NotEmpty(t, results)
	found := false
	for _, r := range results {
		if r.Name == "sync" && r.Domain == DomainCommand {
			found = true
			break
		}
	}
	assert.True(t, found, "sync command should appear in results")
}

// TestSearchWithSynonymExpansion_Integration verifies end-to-end synonym expansion.
func TestSearchWithSynonymExpansion_Integration(t *testing.T) {
	// Build an index with manually populated entries and a synonym source.
	idx := &SearchIndex{
		entries: []SearchEntry{
			{
				Name:        "sre",
				Domain:      DomainRite,
				Summary:     "Site reliability engineering",
				Description: "Manages reliability and operations",
				Keywords:    []string{"reliability", "operations", "incident response"},
				Action:      "/sre",
			},
			{
				Name:        "hygiene",
				Domain:      DomainRite,
				Summary:     "Code quality and cleanup",
				Description: "Refactoring and cleanup workflows",
				Keywords:    []string{"cleanup", "lint", "code-quality"},
				Action:      "/hygiene",
			},
			{
				Name:        "releaser",
				Domain:      DomainRite,
				Summary:     "Release engineering",
				Description: "Publishing and release workflows",
				Keywords:    []string{"publish", "release", "versioning"},
				Action:      "/releaser",
			},
		},
		synonyms: NewStaticSynonymSource(),
	}

	// "deploy" should find SRE via synonym expansion.
	results := idx.Search("deploy", SearchOptions{Limit: 10})
	found := false
	for _, r := range results {
		if r.Name == "sre" {
			found = true
			assert.Equal(t, "keyword", r.MatchType)
			assert.Greater(t, r.Score, 0)
			break
		}
	}
	assert.True(t, found, "deploy should find SRE rite via synonym expansion")

	// "refactor" should find hygiene via synonym expansion.
	results = idx.Search("refactor", SearchOptions{Limit: 10})
	found = false
	for _, r := range results {
		if r.Name == "hygiene" {
			found = true
			assert.Equal(t, "keyword", r.MatchType)
			break
		}
	}
	assert.True(t, found, "refactor should find hygiene rite via synonym expansion")

	// "ship" should find releaser via synonym expansion.
	results = idx.Search("ship", SearchOptions{Limit: 10})
	found = false
	for _, r := range results {
		if r.Name == "releaser" {
			found = true
			assert.Equal(t, "keyword", r.MatchType)
			break
		}
	}
	assert.True(t, found, "ship should find releaser rite via synonym expansion")

	// Direct query should outscore expanded query.
	directResults := idx.Search("sre", SearchOptions{Limit: 10})
	expandedResults := idx.Search("deploy", SearchOptions{Limit: 10})

	var directSREScore, expandedSREScore int
	for _, r := range directResults {
		if r.Name == "sre" {
			directSREScore = r.Score
		}
	}
	for _, r := range expandedResults {
		if r.Name == "sre" {
			expandedSREScore = r.Score
		}
	}
	assert.Greater(t, directSREScore, expandedSREScore,
		"direct 'sre' query (score=%d) should outscore expanded 'deploy' query (score=%d)",
		directSREScore, expandedSREScore)
}

// TestSearchWithoutSynonyms_BackwardCompatible verifies nil synonym source works.
func TestSearchWithoutSynonyms_BackwardCompatible(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session management"},
		},
		synonyms: nil,
	}
	results := idx.Search("session", SearchOptions{Limit: 5})
	require.Len(t, results, 1)
	assert.Equal(t, 1000, results[0].Score)
	assert.Equal(t, "exact", results[0].MatchType)
}
