package bm25

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenize_Basic(t *testing.T) {
	tokens := Tokenize("Hello World")
	assert.Equal(t, []string{"hello", "world"}, tokens)
}

func TestTokenize_MarkdownPunctuation(t *testing.T) {
	tokens := Tokenize("## Package Structure\n\n- `internal/search` | something")
	assert.Contains(t, tokens, "package")
	assert.Contains(t, tokens, "structure")
	assert.Contains(t, tokens, "internal")
	assert.Contains(t, tokens, "search")
	assert.Contains(t, tokens, "something")
}

func TestTokenize_FiltersShortTokens(t *testing.T) {
	tokens := Tokenize("a b cd efg")
	assert.NotContains(t, tokens, "a")
	assert.NotContains(t, tokens, "b")
	assert.Contains(t, tokens, "cd")
	assert.Contains(t, tokens, "efg")
}

func TestTokenize_FiltersExtensions(t *testing.T) {
	tokens := Tokenize("read the go file and yaml config")
	assert.NotContains(t, tokens, "go")
	assert.NotContains(t, tokens, "yaml")
	assert.Contains(t, tokens, "read")
	assert.Contains(t, tokens, "file")
}

func TestTokenize_Empty(t *testing.T) {
	assert.Empty(t, Tokenize(""))
	assert.Empty(t, Tokenize("   "))
}

func TestBuildTermFreqs(t *testing.T) {
	freqs, total := BuildTermFreqs("hello world hello test")
	assert.Equal(t, 2, freqs["hello"])
	assert.Equal(t, 1, freqs["world"])
	assert.Equal(t, 1, freqs["test"])
	assert.Equal(t, 4, total)
}

func TestIndex_DocumentSearch(t *testing.T) {
	idx := NewIndex()

	doc1 := &IndexedUnit{
		QualifiedName: "org::repo::architecture",
		Domain:        "architecture",
		Title:         "Architecture",
	}
	doc1.TermFreqs, doc1.TotalTerms = BuildTermFreqs("package structure layers data flow architecture")

	doc2 := &IndexedUnit{
		QualifiedName: "org::repo::conventions",
		Domain:        "conventions",
		Title:         "Conventions",
	}
	doc2.TermFreqs, doc2.TotalTerms = BuildTermFreqs("error handling conventions naming patterns")

	idx.AddDocument(doc1)
	idx.AddDocument(doc2)
	idx.Finalize()

	results := idx.SearchDocuments("architecture package", 5)
	require.NotEmpty(t, results)
	assert.Equal(t, "org::repo::architecture", results[0].QualifiedName)
	assert.Equal(t, "document", results[0].MatchType)
}

func TestIndex_SectionSearch(t *testing.T) {
	idx := NewIndex()

	sec1 := &IndexedUnit{
		QualifiedName: "org::repo::architecture##package-structure",
		Domain:        "architecture",
		Title:         "Package Structure",
	}
	sec1.TermFreqs, sec1.TotalTerms = BuildTermFreqs("the internal package contains all domain logic")

	sec2 := &IndexedUnit{
		QualifiedName: "org::repo::architecture##layer-boundaries",
		Domain:        "architecture",
		Title:         "Layer Boundaries",
	}
	sec2.TermFreqs, sec2.TotalTerms = BuildTermFreqs("cmd imports domain but domain never imports cmd")

	idx.AddSection(sec1)
	idx.AddSection(sec2)
	idx.Finalize()

	results := idx.SearchSections("domain logic package", 5)
	require.NotEmpty(t, results)
	assert.Equal(t, "section", results[0].MatchType)
}

func TestIndex_EmptyQuery(t *testing.T) {
	idx := NewIndex()
	doc := &IndexedUnit{
		QualifiedName: "org::repo::test",
		Domain:        "test",
		TermFreqs:     map[string]int{"hello": 1},
		TotalTerms:    1,
	}
	idx.AddDocument(doc)
	idx.Finalize()

	assert.Empty(t, idx.SearchDocuments("", 5))
	assert.Empty(t, idx.SearchSections("", 5))
}

func TestIndex_Finalize(t *testing.T) {
	idx := NewIndex()

	doc1 := &IndexedUnit{QualifiedName: "a", TermFreqs: map[string]int{"x": 1}, TotalTerms: 10}
	doc2 := &IndexedUnit{QualifiedName: "b", TermFreqs: map[string]int{"y": 1}, TotalTerms: 20}
	idx.AddDocument(doc1)
	idx.AddDocument(doc2)

	sec1 := &IndexedUnit{QualifiedName: "c", TermFreqs: map[string]int{"z": 1}, TotalTerms: 5}
	idx.AddSection(sec1)

	idx.Finalize()

	assert.Equal(t, 2, idx.TotalDocs)
	assert.Equal(t, 1, idx.TotalSecs)
	assert.InDelta(t, 15.0, idx.AvgDocLen, 0.001) // (10+20)/2
	assert.InDelta(t, 5.0, idx.AvgSecLen, 0.001)
}
