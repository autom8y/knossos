package search

import (
	"log/slog"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/search/bm25"
	"github.com/autom8y/knossos/internal/search/fusion"
)

// SearchIndex holds all indexed entries from all data sources.
type SearchIndex struct {
	entries  []SearchEntry
	synonyms SynonymSource // nil = no expansion (backward compatible)

	// bm25Index is the optional BM25 knowledge index.
	// Non-nil when a registry catalog was successfully loaded.
	bm25Index *bm25.Index

	// catalog is the loaded domain catalog (for provenance lookup).
	catalog *registryorg.DomainCatalog
}

// Build creates a SearchIndex by collecting entries from all data sources.
// root is the Cobra root command; resolver may be nil if no project context.
// Commands and concepts are always collected; rite/agent/dromena/routing
// collectors run only when the resolver has a valid project root.
// Duplicate entries (same name+domain) are deduplicated, keeping the first seen.
//
// If a registry catalog is available, Build also creates a BM25 knowledge index
// from the catalog domains. The BM25 channel is additive-only — if the registry
// is absent, Search() returns structural-only results (backward compatible).
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
		entries = append(entries, CollectProcessions(resolver)...)
		entries = append(entries, CollectParkedSessions(resolver)...)
	}

	// Knowledge domain entries from org registry (fail-open).
	entries = append(entries, CollectKnowledgeDomains()...)

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

	// Construct synonym source: static always available, orchestrator when
	// project context exists. Preserves invariant: ari ask works without a project.
	var synonyms SynonymSource
	static := NewStaticSynonymSource()
	if hasProject {
		orch := NewOrchestratorSynonymSource(resolver.RitesDir())
		synonyms = NewCompositeSynonymSource(static, orch)
	} else {
		synonyms = static
	}

	idx := &SearchIndex{entries: deduped, synonyms: synonyms}

	// Attempt to build BM25 index from registry catalog (fail-open).
	idx.catalog, idx.bm25Index = tryBuildBM25Index()

	return idx
}

// tryBuildBM25Index attempts to load the org catalog and build a BM25 index.
// Returns nil for both values if the registry is not available.
func tryBuildBM25Index() (*registryorg.DomainCatalog, *bm25.Index) {
	orgCtx, err := config.DefaultOrgContext()
	if err != nil {
		return nil, nil
	}

	catalogPath := registryorg.CatalogPath(orgCtx)
	catalog, err := registryorg.LoadCatalog(catalogPath)
	if err != nil {
		return nil, nil
	}

	if catalog.DomainCount() == 0 {
		return catalog, nil
	}

	// Build repo path map for content loading.
	// For now, only local repos are supported. Cross-repo content loading
	// via GitHub API is deferred to a future sprint.
	repoPaths := make(map[string]string)
	// Repo path population is deferred to sprint-4+ when cross-repo
	// path resolution from org config is implemented. For now, content
	// loading relies on repos being accessible locally.

	loader := &bm25.LocalContentLoader{RepoPaths: repoPaths}
	bm25Idx, err := bm25.BuildFromCatalog(catalog, loader)
	if err != nil {
		slog.Debug("BM25 index build failed", "error", err)
		return catalog, nil
	}

	if bm25Idx.TotalDocs == 0 {
		return catalog, nil
	}

	return catalog, bm25Idx
}

// HasBM25 returns true if a BM25 knowledge index is available.
func (idx *SearchIndex) HasBM25() bool {
	return idx.bm25Index != nil
}

// Search finds entries matching the query, scored and ranked.
// Entries with score=0 are excluded. Results are sorted descending by score.
// Domain filter and limit from opts are applied.
//
// When a BM25 index is available, Search runs both structural and BM25 channels,
// then fuses results via RRF. When no BM25 index is available, Search returns
// structural-only results (backward compatible).
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

	// Run the structural channel (always).
	var structuralResults []SearchResult
	for _, e := range idx.entries {
		if filterByDomain && !domainFilter[e.Domain] {
			continue
		}

		score, matchType := scoreEntry(query, e, idx.synonyms)
		if score <= 0 {
			continue
		}

		// H4: Apply session score modifier (no-op when opts.Session is nil).
		score += sessionScoreModifier(e, matchType, score, opts.Session)

		structuralResults = append(structuralResults, SearchResult{
			SearchEntry: e,
			Score:       score,
			MatchType:   matchType,
		})
	}

	// Sort structural results.
	sort.SliceStable(structuralResults, func(i, j int) bool {
		return structuralResults[i].Score > structuralResults[j].Score
	})

	// If no BM25 index or knowledge domain filter excludes BM25, return structural only.
	if idx.bm25Index == nil || (filterByDomain && !domainFilter[DomainKnowledge]) {
		if len(structuralResults) > limit {
			structuralResults = structuralResults[:limit]
		}
		return structuralResults
	}

	// Run BM25 channel.
	topK := limit * 3 // Fetch more to allow for dedup and fusion.
	bm25Docs := idx.bm25Index.SearchDocuments(query, topK)
	bm25Secs := idx.bm25Index.SearchSections(query, topK)

	// If BM25 returned nothing, return structural only.
	if len(bm25Docs) == 0 && len(bm25Secs) == 0 {
		if len(structuralResults) > limit {
			structuralResults = structuralResults[:limit]
		}
		return structuralResults
	}

	// Convert structural results to fusion input format.
	fusionStructural := make([]fusion.StructuralResult, len(structuralResults))
	for i, r := range structuralResults {
		fusionStructural[i] = fusion.StructuralResult{
			Name:      r.Name,
			Domain:    string(r.Domain),
			Score:     r.Score,
			MatchType: r.MatchType,
		}
	}

	// Fuse channels via RRF.
	fused := fusion.RRFMerge(bm25Docs, bm25Secs, fusionStructural, query, bm25.RRFConstK)

	// Convert fused results back to SearchResult format.
	now := time.Now()
	var fusedResults []SearchResult
	for _, fr := range fused {
		if filterByDomain {
			d := Domain(fr.Domain)
			if fr.SourceChannel == "bm25" {
				d = DomainKnowledge
			}
			if !domainFilter[d] {
				continue
			}
		}

		sr := SearchResult{
			SearchEntry: SearchEntry{
				Name:   fr.QualifiedName,
				Domain: Domain(fr.Domain),
			},
			Score:     int(fr.RRFScore * 1000), // Scale RRF score for integer comparison.
			MatchType: fr.MatchType,
		}

		if fr.SourceChannel == "bm25" {
			sr.Domain = DomainKnowledge
			sr.SearchEntry.Summary = fr.Domain
			sr.SearchEntry.Action = "ari knows read " + fr.Domain

			// Add freshness annotation (display only, D-5).
			sr.SearchEntry.Description = freshnessAnnotation(fr, now)
		} else {
			// Reconstruct structural entry fields from the original results.
			for _, orig := range structuralResults {
				key := "structural::" + orig.Name + "::" + string(orig.Domain)
				if key == fr.QualifiedName {
					sr.SearchEntry = orig.SearchEntry
					sr.MatchType = orig.MatchType
					break
				}
			}
		}

		fusedResults = append(fusedResults, sr)
	}

	if len(fusedResults) > limit {
		fusedResults = fusedResults[:limit]
	}

	return fusedResults
}

// freshnessAnnotation returns a display-only annotation for a knowledge result.
// D-5: Display only, NOT ranking. D-4: Freshness thresholds match trust tiers.
func freshnessAnnotation(fr fusion.FusedResult, now time.Time) string {
	if fr.Freshness <= 0 {
		// Try to compute from catalog if available.
		return fr.Domain
	}
	if fr.Freshness >= 0.7 {
		return fr.Domain // No annotation needed.
	}
	if fr.Freshness >= 0.4 {
		return fr.Domain + " (moderately stale)"
	}
	return fr.Domain + " (stale)"
}
