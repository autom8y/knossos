package bm25

import (
	"sort"
	"strings"
)

// IndexedUnit represents a single indexable unit (document or section).
type IndexedUnit struct {
	// QualifiedName is the canonical cross-repo address.
	// Documents: "org::repo::domain"
	// Sections:  "org::repo::domain##section-slug"
	QualifiedName string

	// Domain is the bare domain name from .know/ frontmatter.
	Domain string

	// Title is the document or section heading.
	Title string

	// RawText is the original text content (for display snippets).
	RawText string

	// TermFreqs maps each term to its frequency in this unit.
	TermFreqs map[string]int

	// TotalTerms is the total number of terms in this unit.
	TotalTerms int

	// GeneratedAt is the RFC3339 timestamp from frontmatter (for freshness).
	GeneratedAt string
}

// Index holds the complete BM25 index with dual document/section indexing.
type Index struct {
	Documents []*IndexedUnit // All document-level indexed units
	Sections  []*IndexedUnit // All section-level indexed units

	DocFreq map[string]int // Term -> number of documents containing it
	SecFreq map[string]int // Term -> number of sections containing it

	TotalDocs int // Total document count
	TotalSecs int // Total section count

	AvgDocLen float64 // Average document length (terms)
	AvgSecLen float64 // Average section length (terms)
}

// NewIndex creates an empty index.
func NewIndex() *Index {
	return &Index{
		DocFreq: make(map[string]int),
		SecFreq: make(map[string]int),
	}
}

// AddDocument indexes a document-level unit.
func (idx *Index) AddDocument(unit *IndexedUnit) {
	idx.Documents = append(idx.Documents, unit)
	idx.TotalDocs++

	for term := range unit.TermFreqs {
		idx.DocFreq[term]++
	}
}

// AddSection indexes a section-level unit.
func (idx *Index) AddSection(unit *IndexedUnit) {
	idx.Sections = append(idx.Sections, unit)
	idx.TotalSecs++

	for term := range unit.TermFreqs {
		idx.SecFreq[term]++
	}
}

// Finalize computes average document/section lengths after all units are added.
// Must be called before searching.
func (idx *Index) Finalize() {
	if idx.TotalDocs > 0 {
		totalTerms := 0
		for _, doc := range idx.Documents {
			totalTerms += doc.TotalTerms
		}
		idx.AvgDocLen = float64(totalTerms) / float64(idx.TotalDocs)
	}

	if idx.TotalSecs > 0 {
		totalTerms := 0
		for _, sec := range idx.Sections {
			totalTerms += sec.TotalTerms
		}
		idx.AvgSecLen = float64(totalTerms) / float64(idx.TotalSecs)
	}
}

// SearchResult holds a single BM25 search result with scoring details.
type SearchResult struct {
	QualifiedName string
	Score         float64
	MatchType     string // "document" or "section"
	Domain        string
}

// SearchDocuments searches the document-level index and returns top-k ranked results.
func (idx *Index) SearchDocuments(query string, k int) []SearchResult {
	queryTerms := Tokenize(query)
	if len(queryTerms) == 0 {
		return nil
	}

	bm := NewBM25()

	type scored struct {
		unit  *IndexedUnit
		score float64
	}

	var results []scored
	for _, doc := range idx.Documents {
		s := bm.ScoreDocument(queryTerms, doc, idx.DocFreq, idx.TotalDocs, idx.AvgDocLen)
		if s > 0 {
			results = append(results, scored{doc, s})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if k > len(results) {
		k = len(results)
	}

	out := make([]SearchResult, k)
	for i := 0; i < k; i++ {
		r := results[i]
		out[i] = SearchResult{
			QualifiedName: r.unit.QualifiedName,
			Score:         r.score,
			MatchType:     "document",
			Domain:        r.unit.Domain,
		}
	}
	return out
}

// SearchSections searches the section-level index and returns top-k ranked results.
func (idx *Index) SearchSections(query string, k int) []SearchResult {
	queryTerms := Tokenize(query)
	if len(queryTerms) == 0 {
		return nil
	}

	bm := NewBM25()

	type scored struct {
		unit  *IndexedUnit
		score float64
	}

	var results []scored
	for _, sec := range idx.Sections {
		s := bm.ScoreDocument(queryTerms, sec, idx.SecFreq, idx.TotalSecs, idx.AvgSecLen)
		if s > 0 {
			results = append(results, scored{sec, s})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if k > len(results) {
		k = len(results)
	}

	out := make([]SearchResult, k)
	for i := 0; i < k; i++ {
		r := results[i]
		out[i] = SearchResult{
			QualifiedName: r.unit.QualifiedName,
			Score:         r.score,
			MatchType:     "section",
			Domain:        r.unit.Domain,
		}
	}
	return out
}

// Tokenize splits text into lowercase terms, filtering out short tokens,
// file extensions, and markdown punctuation.
func Tokenize(text string) []string {
	// Replace common markdown and code artifacts with spaces.
	replacer := strings.NewReplacer(
		"|", " ", "`", " ", "#", " ", "*", " ", "-", " ",
		"(", " ", ")", " ", "[", " ", "]", " ", "{", " ",
		"}", " ", "<", " ", ">", " ", "/", " ", "\\", " ",
		":", " ", ";", " ", ",", " ", "\"", " ", "'", " ",
		"=", " ", "+", " ", "~", " ", "^", " ", "!", " ",
		"@", " ", "$", " ", "%", " ", "&", " ", "?", " ",
		"\t", " ", "\n", " ", "\r", " ",
	)
	cleaned := replacer.Replace(text)

	words := strings.Fields(cleaned)
	tokens := make([]string, 0, len(words))

	// File extension set for filtering.
	extFilter := map[string]bool{
		"md": true, "go": true, "py": true, "yaml": true,
		"json": true, "toml": true, "yml": true,
	}

	for _, w := range words {
		w = strings.ToLower(strings.TrimSpace(w))
		if len(w) < 2 {
			continue
		}
		if extFilter[w] {
			continue
		}
		tokens = append(tokens, w)
	}
	return tokens
}

// BuildTermFreqs computes term frequencies for a text.
// Returns the frequency map and total term count.
func BuildTermFreqs(text string) (map[string]int, int) {
	tokens := Tokenize(text)
	freqs := make(map[string]int, len(tokens))
	for _, t := range tokens {
		freqs[t]++
	}
	return freqs, len(tokens)
}
