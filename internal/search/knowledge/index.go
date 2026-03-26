package knowledge

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/autom8y/knossos/internal/search/knowledge/embedding"
	"github.com/autom8y/knossos/internal/search/knowledge/graph"
	"github.com/autom8y/knossos/internal/search/knowledge/summary"
)

// KnowledgeIndex is the unified coordinator wrapping four sub-components:
// BM25, summaries, embeddings, and entity graph.
//
// BC-05: ONE BM25 index. Existing internal/search/bm25/ is wrapped, not duplicated.
// BC-10: Tier 1 uses restart-required for cache coherence. No hot-reload.
// BC-12: EmbeddingResult includes Freshness from day one (zero in Tier 1).
//
// RR-006 guard: This file must stay under 500 lines.
type KnowledgeIndex struct {
	bm25       BM25Searcher
	summaries  *summary.Store
	embeddings *embedding.Store
	graph      *graph.Graph
	catalog    map[string]*DomainMetadata // qualifiedName -> metadata
}

// SearchByEmbedding performs cosine similarity search against domain embeddings.
// Returns the top-k results sorted by similarity descending.
// BC-12: Results include Freshness field (zero in Tier 1).
func (ki *KnowledgeIndex) SearchByEmbedding(queryEmbedding []float64, k int) []EmbeddingResult {
	if ki.embeddings == nil {
		return nil
	}

	results := ki.embeddings.Search(queryEmbedding, k)
	out := make([]EmbeddingResult, len(results))
	for i, r := range results {
		out[i] = EmbeddingResult{
			QualifiedName: r.QualifiedName,
			Similarity:    r.Similarity,
			Freshness:     r.Freshness, // BC-12: zero in Tier 1.
		}
	}
	return out
}

// SearchByBM25 delegates to the wrapped BM25 index for text-based search.
// BC-05: ONE BM25 index -- this is the same index used by internal/search/.
func (ki *KnowledgeIndex) SearchByBM25(queryText string, k int) []BM25Result {
	if ki.bm25 == nil {
		return nil
	}

	hits := ki.bm25.SearchDocuments(queryText, k)
	out := make([]BM25Result, len(hits))
	for i, h := range hits {
		out[i] = BM25Result{
			QualifiedName: h.QualifiedName,
			Score:         h.Score,
			Domain:        h.Domain,
			RawText:       h.RawText,
		}
	}
	return out
}

// GetSummary returns the domain summary text for the given qualified name.
func (ki *KnowledgeIndex) GetSummary(qualifiedName string) (string, bool) {
	if ki.summaries == nil {
		return "", false
	}
	return ki.summaries.GetSummary(qualifiedName)
}

// GetRelationships returns entity graph relationships for a domain.
func (ki *KnowledgeIndex) GetRelationships(qualifiedName string) []Relationship {
	if ki.graph == nil {
		return nil
	}

	edges := ki.graph.GetRelationships(qualifiedName)
	out := make([]Relationship, len(edges))
	for i, e := range edges {
		out[i] = Relationship{
			Target:   e.Target,
			Type:     EdgeType(e.Type),
			Strength: e.Strength,
			Evidence: e.Evidence,
		}
	}
	return out
}

// GetMetadata returns the domain metadata for the given qualified name.
func (ki *KnowledgeIndex) GetMetadata(qualifiedName string) (*DomainMetadata, bool) {
	meta, ok := ki.catalog[qualifiedName]
	return meta, ok
}

// DomainCount returns the number of indexed domains.
func (ki *KnowledgeIndex) DomainCount() int {
	return len(ki.catalog)
}

// SummaryCount returns the number of stored domain summaries.
func (ki *KnowledgeIndex) SummaryCount() int {
	if ki.summaries == nil {
		return 0
	}
	return ki.summaries.Count()
}

// EmbeddingCount returns the number of stored domain embeddings.
func (ki *KnowledgeIndex) EmbeddingCount() int {
	if ki.embeddings == nil {
		return 0
	}
	return ki.embeddings.Count()
}

// EdgeCount returns the total number of entity graph edges.
func (ki *KnowledgeIndex) EdgeCount() int {
	if ki.graph == nil {
		return 0
	}
	return ki.graph.EdgeCount()
}

// NeedsReindex returns true if the domain's current source_hash differs from
// the indexed hash. Returns true for domains not yet in the index.
func (ki *KnowledgeIndex) NeedsReindex(qualifiedName, currentSourceHash string) bool {
	meta, ok := ki.catalog[qualifiedName]
	if !ok || meta == nil {
		return true
	}
	return meta.SourceHash != currentSourceHash
}

// Reindex regenerates the index entry for a single domain.
// This performs eager rebuild: summary + embedding + metadata update.
// D-L5: Eager rebuild on source_hash change.
func (ki *KnowledgeIndex) Reindex(ctx context.Context, qualifiedName, content string, meta *DomainMetadata, llmClient LLMClient) error {
	if meta == nil {
		return fmt.Errorf("metadata is required for reindex of %s", qualifiedName)
	}

	// Update metadata.
	meta.IndexedAt = time.Now()
	ki.catalog[qualifiedName] = meta

	// Regenerate summary if LLM client is available.
	if llmClient != nil {
		sections := parseSections(content)
		adapter := &llmClientAdapter{client: llmClient}

		genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := ki.summaries.Generate(genCtx, qualifiedName, content, meta.SourceHash, sections, adapter)
		cancel()

		if err != nil {
			slog.Warn("summary regeneration failed during reindex",
				"domain", qualifiedName, "error", err)
		}
	}

	// Regenerate embedding.
	summaryText, hasSummary := ki.summaries.GetSummary(qualifiedName)
	embedText := summaryText
	if !hasSummary {
		embedText = truncateForEmbedding(content, 2000)
	}
	if embedText != "" {
		vec := embedding.TextToVector(embedText, embeddingDimensions)
		if vec != nil {
			ki.embeddings.Add(qualifiedName, vec, meta.SourceHash)
		}
	}

	return nil
}

// AllDomainNames returns all indexed domain qualified names.
func (ki *KnowledgeIndex) AllDomainNames() []string {
	names := make([]string, 0, len(ki.catalog))
	for qn := range ki.catalog {
		names = append(names, qn)
	}
	return names
}

// HasBM25 returns true if a BM25 index is available.
func (ki *KnowledgeIndex) HasBM25() bool {
	return ki.bm25 != nil
}

// Summaries returns the underlying summary store. Used for persistence.
func (ki *KnowledgeIndex) Summaries() *summary.Store {
	return ki.summaries
}

// Embeddings returns the underlying embedding store. Used for persistence.
func (ki *KnowledgeIndex) Embeddings() *embedding.Store {
	return ki.embeddings
}

// Graph returns the underlying entity graph. Used for persistence.
func (ki *KnowledgeIndex) Graph() *graph.Graph {
	return ki.graph
}

// Catalog returns the domain metadata catalog. Used for persistence.
func (ki *KnowledgeIndex) Catalog() map[string]*DomainMetadata {
	return ki.catalog
}
