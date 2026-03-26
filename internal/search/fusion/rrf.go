// Package fusion implements Reciprocal Rank Fusion (RRF) for merging
// search results from multiple channels (BM25 knowledge + structural CLI surface).
package fusion

import (
	"sort"
	"strings"

	"github.com/autom8y/knossos/internal/search/bm25"
)

// FusedResult holds a result from the RRF fusion of multiple search channels.
type FusedResult struct {
	// QualifiedName is the canonical ID.
	// BM25 results: "org::repo::domain" or "org::repo::domain##section"
	// Structural results: "structural::name::domain"
	QualifiedName string

	// RRFScore is the fused Reciprocal Rank Fusion score.
	RRFScore float64

	// ChannelRanks maps channel name to rank (1-indexed) in that channel.
	// Possible keys: "bm25-doc", "bm25-sec", "structural"
	ChannelRanks map[string]int

	// MatchType indicates the primary source: "bm25-document", "bm25-section",
	// "structural", or "fused" (when multiple channels contributed).
	MatchType string

	// Domain is the .know/ domain or structural domain.
	Domain string

	// SourceChannel is "bm25" or "structural".
	SourceChannel string

	// Freshness is the search-level freshness score (0.0-1.0).
	// Only populated for BM25 results.
	Freshness float64

	// RawText is the full .know/ content (frontmatter stripped).
	// Populated from BM25 IndexedUnit.RawText; empty for structural results.
	RawText string
}

// StructuralResult represents a result from the existing 4-tier structural scorer.
// This wraps the production SearchResult for RRF input.
type StructuralResult struct {
	Name      string
	Domain    string
	Score     int
	MatchType string
}

// StripSection extracts the parent QN from a section-qualified name.
// "org::repo::domain##section" -> "org::repo::domain"
// "org::repo::domain" -> "org::repo::domain" (unchanged)
func StripSection(qn string) string {
	if idx := strings.Index(qn, "##"); idx >= 0 {
		return qn[:idx]
	}
	return qn
}

// deduplicateBM25 applies section-wins dedup with a cap of maxSecPerParent
// sections per parent document.
//
// Strategy:
//  1. Group sections by parent QN (already score-sorted).
//  2. Keep top maxSecPerParent sections per parent.
//  3. Drop the parent document result when it has section children.
//  4. Add documents with no section children (document-as-fallback).
//  5. Sort merged result by BM25 score descending.
func deduplicateBM25(docResults, secResults []bm25.SearchResult, maxSecPerParent int) []bm25.SearchResult {
	parentSections := map[string][]bm25.SearchResult{}
	for _, sec := range secResults {
		parent := StripSection(sec.QualifiedName)
		parentSections[parent] = append(parentSections[parent], sec)
	}

	var merged []bm25.SearchResult
	selectedParents := map[string]bool{}

	// Keep top-N sections per parent.
	for parent, secs := range parentSections {
		selectedParents[parent] = true
		limit := maxSecPerParent
		if limit > len(secs) {
			limit = len(secs)
		}
		merged = append(merged, secs[:limit]...)
	}

	// Add documents with no section children.
	for _, doc := range docResults {
		if !selectedParents[doc.QualifiedName] {
			merged = append(merged, doc)
		}
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	return merged
}

// applyDomainNameBoosting multiplies BM25 scores by 2.0x for documents whose
// domain name contains a query term. This is applied before RRF rank assignment.
// D-5: Domain-name boosting only; no freshness re-ranking.
func applyDomainNameBoosting(results []bm25.SearchResult, queryTerms []string) []bm25.SearchResult {
	if len(queryTerms) == 0 {
		return results
	}

	boosted := make([]bm25.SearchResult, len(results))
	copy(boosted, results)

	for i := range boosted {
		domainLower := strings.ToLower(boosted[i].Domain)
		for _, term := range queryTerms {
			termLower := strings.ToLower(term)
			if len(termLower) >= 3 && strings.Contains(domainLower, termLower) {
				boosted[i].Score *= 2.0
				break
			}
		}
	}

	// Re-sort after boosting.
	sort.Slice(boosted, func(i, j int) bool {
		return boosted[i].Score > boosted[j].Score
	})

	return boosted
}

// RRFMerge combines BM25 and structural search results using Reciprocal Rank Fusion.
//
//	RRF(d) = sum_i [ 1 / (k + rank_i(d)) ]
//
// Domain-name boosting is applied to BM25 results before rank assignment (D-5).
// Section deduplication limits sections to top-2 per parent document.
//
// The BM25 and structural channels index different corpora (.know/ files vs CLI surface),
// so RRF degenerates to weighted rank interleaving. This is still the correct approach:
// it gives a principled, parameter-controlled merge without comparing heterogeneous scores.
func RRFMerge(bm25Doc []bm25.SearchResult, bm25Sec []bm25.SearchResult,
	structural []StructuralResult, query string, k float64) []FusedResult {

	queryTerms := bm25.Tokenize(query)

	// Apply domain-name boosting to BM25 results before dedup/rank assignment.
	bm25Doc = applyDomainNameBoosting(bm25Doc, queryTerms)
	bm25Sec = applyDomainNameBoosting(bm25Sec, queryTerms)

	// Apply section dedup within the BM25 channel.
	bm25Merged := deduplicateBM25(bm25Doc, bm25Sec, 2)

	// Map: key -> *FusedResult
	fused := map[string]*FusedResult{}

	// Channel A: BM25 (knowledge documents and sections, post-dedup).
	for rank, r := range bm25Merged {
		rrfContrib := 1.0 / (k + float64(rank+1))
		key := r.QualifiedName

		if _, exists := fused[key]; !exists {
			mt := "bm25-document"
			if r.MatchType == "section" {
				mt = "bm25-section"
			}
			fused[key] = &FusedResult{
				QualifiedName: r.QualifiedName,
				ChannelRanks:  map[string]int{},
				MatchType:     mt,
				Domain:        r.Domain,
				SourceChannel: "bm25",
				RawText:       r.RawText,
			}
		}

		fr := fused[key]
		fr.RRFScore += rrfContrib

		channelName := "bm25-doc"
		if r.MatchType == "section" {
			channelName = "bm25-sec"
		}
		fr.ChannelRanks[channelName] = rank + 1
	}

	// Channel B: Structural (CLI surface entries).
	for rank, r := range structural {
		key := "structural::" + r.Name + "::" + r.Domain
		rrfContrib := 1.0 / (k + float64(rank+1))

		if _, exists := fused[key]; !exists {
			fused[key] = &FusedResult{
				QualifiedName: key,
				ChannelRanks:  map[string]int{},
				MatchType:     "structural",
				Domain:        r.Domain,
				SourceChannel: "structural",
			}
		}

		fr := fused[key]
		fr.RRFScore += rrfContrib
		fr.ChannelRanks["structural"] = rank + 1
	}

	// Collect and sort by RRF score descending.
	results := make([]FusedResult, 0, len(fused))
	for _, fr := range fused {
		if len(fr.ChannelRanks) > 1 {
			fr.MatchType = "fused"
		}
		results = append(results, *fr)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].RRFScore != results[j].RRFScore {
			return results[i].RRFScore > results[j].RRFScore
		}
		return results[i].QualifiedName < results[j].QualifiedName
	})

	return results
}
