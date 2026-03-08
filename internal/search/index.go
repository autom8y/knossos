package search

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/paths"
)

// SearchIndex holds all indexed entries from all data sources.
type SearchIndex struct {
	entries []SearchEntry
}

// Build creates a SearchIndex by collecting entries from all data sources.
// root is the Cobra root command; resolver may be nil if no project context.
// Commands and concepts are always collected; rite/agent/dromena/routing
// collectors run only when the resolver has a valid project root.
// Duplicate entries (same name+domain) are deduplicated, keeping the first seen.
func Build(root *cobra.Command, resolver *paths.Resolver) *SearchIndex {
	var entries []SearchEntry

	// Always available: CLI surface and concept registry.
	entries = append(entries, CollectCommands(root)...)
	entries = append(entries, CollectConcepts()...)

	// Project-scoped sources (require a valid project root).
	hasProject := resolver != nil && resolver.ProjectRoot() != ""
	if hasProject {
		entries = append(entries, CollectRites(resolver)...)
		entries = append(entries, CollectAgents(resolver)...)
		entries = append(entries, CollectDromena(resolver)...)
		entries = append(entries, CollectRouting(resolver)...)
	}

	// Deduplicate: same name+domain keeps first seen.
	seen := make(map[string]bool, len(entries))
	deduped := make([]SearchEntry, 0, len(entries))
	for _, e := range entries {
		key := string(e.Domain) + ":" + e.Name
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, e)
	}

	return &SearchIndex{entries: deduped}
}

// Search finds entries matching the query, scored and ranked.
// Entries with score=0 are excluded. Results are sorted descending by score.
// Domain filter and limit from opts are applied.
func (idx *SearchIndex) Search(query string, opts SearchOptions) []SearchResult {
	limit := opts.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}

	// Build domain filter set for O(1) lookup.
	domainFilter := make(map[Domain]bool, len(opts.Domains))
	for _, d := range opts.Domains {
		domainFilter[d] = true
	}
	filterByDomain := len(opts.Domains) > 0

	var results []SearchResult
	for _, e := range idx.entries {
		if filterByDomain && !domainFilter[e.Domain] {
			continue
		}

		score, matchType := scoreEntry(query, e)
		if score <= 0 {
			continue
		}

		results = append(results, SearchResult{
			SearchEntry: e,
			Score:       score,
			MatchType:   matchType,
		})
	}

	// Sort descending by score; stable sort preserves insertion order for ties.
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}
