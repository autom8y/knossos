package fusion

import (
	"testing"

	"github.com/autom8y/knossos/internal/search/bm25"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripSection(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"org::repo::domain##section", "org::repo::domain"},
		{"org::repo::domain", "org::repo::domain"},
		{"no-separator", "no-separator"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, StripSection(tt.input))
		})
	}
}

func TestRRFMerge_BM25Only(t *testing.T) {
	docs := []bm25.SearchResult{
		{QualifiedName: "org::repo::arch", Score: 5.0, MatchType: "document", Domain: "architecture"},
		{QualifiedName: "org::repo::conv", Score: 3.0, MatchType: "document", Domain: "conventions"},
	}
	secs := []bm25.SearchResult{
		{QualifiedName: "org::repo::arch##pkg", Score: 4.0, MatchType: "section", Domain: "architecture"},
	}

	results := RRFMerge(docs, secs, nil, "architecture package", 40.0)
	require.NotEmpty(t, results)

	// All results should be from BM25 channel.
	for _, r := range results {
		assert.Equal(t, "bm25", r.SourceChannel)
	}
}

func TestRRFMerge_StructuralOnly(t *testing.T) {
	structural := []StructuralResult{
		{Name: "session create", Domain: "command", Score: 1000, MatchType: "exact"},
		{Name: "session park", Domain: "command", Score: 700, MatchType: "prefix"},
	}

	results := RRFMerge(nil, nil, structural, "session", 40.0)
	require.Len(t, results, 2)

	for _, r := range results {
		assert.Equal(t, "structural", r.SourceChannel)
	}
}

func TestRRFMerge_MixedChannels(t *testing.T) {
	docs := []bm25.SearchResult{
		{QualifiedName: "org::repo::arch", Score: 5.0, MatchType: "document", Domain: "architecture"},
	}
	structural := []StructuralResult{
		{Name: "session create", Domain: "command", Score: 1000, MatchType: "exact"},
	}

	results := RRFMerge(docs, nil, structural, "architecture session", 40.0)
	require.NotEmpty(t, results)

	// Both channels should be represented.
	channels := map[string]bool{}
	for _, r := range results {
		channels[r.SourceChannel] = true
	}
	assert.True(t, channels["bm25"])
	assert.True(t, channels["structural"])
}

func TestRRFMerge_SectionDedup(t *testing.T) {
	// Three sections from the same parent -- only top-2 should survive.
	secs := []bm25.SearchResult{
		{QualifiedName: "org::repo::arch##sec1", Score: 5.0, MatchType: "section", Domain: "architecture"},
		{QualifiedName: "org::repo::arch##sec2", Score: 4.0, MatchType: "section", Domain: "architecture"},
		{QualifiedName: "org::repo::arch##sec3", Score: 3.0, MatchType: "section", Domain: "architecture"},
	}
	// Parent doc should be dropped because it has section children.
	docs := []bm25.SearchResult{
		{QualifiedName: "org::repo::arch", Score: 6.0, MatchType: "document", Domain: "architecture"},
	}

	results := RRFMerge(docs, secs, nil, "architecture", 40.0)

	// Should have 2 section results (top-2) and no parent doc.
	archCount := 0
	for _, r := range results {
		if r.Domain == "architecture" {
			archCount++
		}
	}
	assert.Equal(t, 2, archCount, "should have exactly 2 results for architecture (top-2 sections)")
}

func TestRRFMerge_DomainNameBoosting(t *testing.T) {
	// Two documents: one whose domain matches the query term, one that doesn't.
	docs := []bm25.SearchResult{
		{QualifiedName: "org::repo::conventions", Score: 5.0, MatchType: "document", Domain: "conventions"},
		{QualifiedName: "org::repo::architecture", Score: 4.8, MatchType: "document", Domain: "architecture"},
	}

	// Query "architecture" should boost the architecture doc.
	results := RRFMerge(docs, nil, nil, "architecture patterns", 40.0)
	require.NotEmpty(t, results)

	// Architecture should be ranked first because of domain-name boost (4.8 * 2.0 > 5.0).
	assert.Equal(t, "org::repo::architecture", results[0].QualifiedName)
}

func TestRRFMerge_EmptyInputs(t *testing.T) {
	results := RRFMerge(nil, nil, nil, "test", 40.0)
	assert.Empty(t, results)
}

func TestDeduplicateBM25_DocumentFallback(t *testing.T) {
	// Document with no sections should survive.
	docs := []bm25.SearchResult{
		{QualifiedName: "org::repo::conv", Score: 5.0, MatchType: "document", Domain: "conventions"},
		{QualifiedName: "org::repo::arch", Score: 4.0, MatchType: "document", Domain: "architecture"},
	}
	secs := []bm25.SearchResult{
		{QualifiedName: "org::repo::arch##sec1", Score: 6.0, MatchType: "section", Domain: "architecture"},
	}

	merged := deduplicateBM25(docs, secs, 2)

	// conv (no sections, kept) + arch##sec1 (section, kept). arch doc dropped.
	qns := map[string]bool{}
	for _, r := range merged {
		qns[r.QualifiedName] = true
	}
	assert.True(t, qns["org::repo::conv"], "conv doc should survive (no sections)")
	assert.True(t, qns["org::repo::arch##sec1"], "arch section should survive")
	assert.False(t, qns["org::repo::arch"], "arch doc should be dropped (has section children)")
}
