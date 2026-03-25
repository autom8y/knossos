package bm25

import (
	"math"
	"strings"
)

// BM25 implements the Okapi BM25 scoring algorithm.
//
//	score(D, Q) = sum_i IDF(qi) * (f(qi, D) * (k1 + 1)) / (f(qi, D) + k1 * (1 - b + b * |D|/avgdl))
//
// where IDF(qi) = ln((N - n(qi) + 0.5) / (n(qi) + 0.5) + 1)
// (Robertson-Sparck Jones formulation with +1 to prevent negative IDF)
type BM25 struct {
	K1 float64 // Term frequency saturation parameter
	B  float64 // Length normalization parameter
}

// NewBM25 returns a BM25 scorer with the empirically validated parameters.
func NewBM25() *BM25 {
	return &BM25{K1: BM25K1, B: BM25B}
}

// IDF computes the Inverse Document Frequency using the Robertson-Sparck Jones formula.
// docFreq is the number of documents containing the term.
// totalDocs is the total number of documents in the corpus.
func (bm *BM25) IDF(docFreq, totalDocs int) float64 {
	n := float64(docFreq)
	N := float64(totalDocs)
	return math.Log((N-n+0.5)/(n+0.5) + 1.0)
}

// ScoreDocument computes the total BM25 score for a query against an indexed unit.
// queryTerms are pre-tokenized query terms (lowercase).
// freqMap is the corpus-level document frequency map (DocFreq or SecFreq).
// totalDocs is the total document/section count.
// avgDocLen is the average document/section length.
func (bm *BM25) ScoreDocument(queryTerms []string, unit *IndexedUnit,
	freqMap map[string]int, totalDocs int, avgDocLen float64) float64 {

	score := 0.0
	for _, term := range queryTerms {
		term = strings.ToLower(term)
		tf := unit.TermFreqs[term]
		if tf == 0 {
			continue
		}
		n := freqMap[term]
		idf := bm.IDF(n, totalDocs)

		ftf := float64(tf)
		fdl := float64(unit.TotalTerms)
		numerator := ftf * (bm.K1 + 1.0)
		denominator := ftf + bm.K1*(1.0-bm.B+bm.B*fdl/avgDocLen)
		score += idf * numerator / denominator
	}
	return score
}
