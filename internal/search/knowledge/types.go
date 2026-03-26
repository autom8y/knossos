// Package knowledge implements the KnowledgeIndex coordinator for the Clew v2
// intelligence pipeline (Sprint 7).
//
// KnowledgeIndex wraps four sub-components into a unified index:
//   - BM25 term frequencies (existing internal/search/bm25/, wrapped not duplicated -- BC-05)
//   - Section summaries (Haiku-generated, cached by source_hash)
//   - Domain embeddings (in-memory vectors with cosine similarity)
//   - Entity graph with typed edges (deterministic, zero LLM cost)
//
// RR-007 guard: This package and its sub-packages MUST NOT import the parent
// internal/search/ package. This prevents Cobra dependency contamination.
//
// BC-10: Tier 1 uses restart-required for cache coherence. No hot-reload.
// BC-11: KnowledgeIndex JSON is pre-baked in the container image.
// BC-12: EmbeddingResult includes Freshness from day one (zero in Tier 1).
package knowledge

import (
	"context"
	"time"
)

// DomainMetadata holds enriched metadata for a single domain in the KnowledgeIndex.
// This extends the catalog's DomainEntry with KI-specific fields.
type DomainMetadata struct {
	// QualifiedName is the canonical cross-repo address: "org::repo::domain".
	QualifiedName string `json:"qualified_name"`

	// DomainType is the domain classification (architecture, scar-tissue, etc.).
	DomainType string `json:"domain_type"`

	// Repo is the repository name extracted from the qualified name.
	Repo string `json:"repo"`

	// SourceHash is the git short SHA from the domain file's frontmatter.
	// Used for incremental rebuild detection.
	SourceHash string `json:"source_hash"`

	// SourceScope lists the file globs that this domain covers.
	// Used for scope_overlap edge computation in the entity graph.
	SourceScope []string `json:"source_scope,omitempty"`

	// GeneratedAt is the RFC3339 timestamp from the domain file's frontmatter.
	GeneratedAt string `json:"generated_at"`

	// FreshnessScore is the domain freshness in [0.0, 1.0].
	// BC-12: Included from day one, zero-valued in Tier 1.
	FreshnessScore float64 `json:"freshness_score"`

	// IndexedAt is when this domain was last indexed into the KnowledgeIndex.
	IndexedAt time.Time `json:"indexed_at"`
}

// EmbeddingResult holds a single embedding search result.
type EmbeddingResult struct {
	// QualifiedName is the canonical domain address.
	QualifiedName string

	// Similarity is the cosine similarity score in [-1.0, 1.0].
	Similarity float64

	// Freshness is the domain freshness score in [0.0, 1.0].
	// BC-12: Included from day one, zero-valued in Tier 1.
	Freshness float64
}

// BM25Result holds a single BM25 search result from the wrapped index.
type BM25Result struct {
	// QualifiedName is the canonical domain address.
	QualifiedName string

	// Score is the BM25 relevance score.
	Score float64

	// Domain is the bare domain name.
	Domain string

	// RawText is the full .know/ content (frontmatter stripped).
	RawText string
}

// Relationship represents a typed edge in the entity graph.
type Relationship struct {
	// Target is the qualified name of the connected domain.
	Target string `json:"target"`

	// Type is the edge classification.
	Type EdgeType `json:"type"`

	// Strength is the edge weight in [0.0, 1.0].
	Strength float64 `json:"strength"`

	// Evidence is a human-readable explanation of why this edge exists.
	Evidence string `json:"evidence"`
}

// EdgeType classifies entity graph relationships.
type EdgeType string

const (
	// EdgeSameType connects domains with the same domain type across repos.
	EdgeSameType EdgeType = "same_type"

	// EdgeSameRepo connects domains within the same repository.
	EdgeSameRepo EdgeType = "same_repo"

	// EdgeScopeOverlap connects domains with overlapping source_scope globs.
	EdgeScopeOverlap EdgeType = "scope_overlap"
)

// BuildConfig holds configuration for the KnowledgeIndex builder.
type BuildConfig struct {
	// Catalog provides domain metadata from the org registry.
	Catalog DomainCatalog

	// ContentStore loads .know/ file content for indexing.
	ContentStore ContentStore

	// LLMClient generates domain and section summaries via Haiku.
	// May be nil -- summary generation is skipped when unavailable.
	LLMClient LLMClient

	// PersistedPath is the filesystem path for KnowledgeIndex JSON persistence.
	// Empty string means no persistence (pure in-memory build).
	PersistedPath string

	// BM25Index is the existing BM25 index to wrap (BC-05: ONE index, not duplicated).
	// May be nil -- BM25 search returns empty results when unavailable.
	BM25Index BM25Searcher
}

// DomainCatalog is the interface for accessing the org domain registry.
// This is a narrow interface to avoid importing internal/registry/org directly.
type DomainCatalog interface {
	ListDomains() []CatalogDomainEntry
	LookupDomain(qualifiedName string) (CatalogDomainEntry, bool)
	DomainCount() int
}

// CatalogDomainEntry mirrors registryorg.DomainEntry for dependency isolation.
// RR-007: knowledge/ sub-packages must not import parent search/ or registry/.
type CatalogDomainEntry struct {
	QualifiedName string
	Domain        string
	Path          string
	GeneratedAt   string
	ExpiresAfter  string
	SourceHash    string
	Confidence    float64
}

// ContentStore abstracts content loading for .know/ domain files.
// RR-007: Narrow interface avoids importing content.Store directly.
type ContentStore interface {
	// LoadContent returns the markdown body (frontmatter stripped) for a domain.
	LoadContent(qualifiedName string) (string, error)

	// HasContent returns true if content is available for the domain.
	HasContent(qualifiedName string) bool
}

// LLMClient abstracts LLM completion calls for summary generation.
// RR-007: Narrow interface avoids importing internal/llm directly.
type LLMClient interface {
	// Complete sends a completion request and returns the response text.
	Complete(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error)
}

// BM25Searcher abstracts the BM25 search interface.
// RR-007: Narrow interface avoids importing internal/search/bm25 directly.
type BM25Searcher interface {
	// SearchDocuments searches document-level BM25 index for top-k results.
	SearchDocuments(query string, k int) []BM25SearchHit

	// SearchSections searches section-level BM25 index for top-k results.
	SearchSections(query string, k int) []BM25SearchHit
}

// BM25SearchHit is a single hit from the BM25 searcher.
type BM25SearchHit struct {
	QualifiedName string
	Score         float64
	Domain        string
	RawText       string
	MatchType     string // "document" or "section"
}
