// Package triage implements the multi-stage query triage orchestrator.
//
// BC-02: This is a new top-level package between slack/ and reason/ in the
// dependency DAG. It imports llm/ and uses search/ interfaces.
//
// The triage pipeline runs Stages 0-3 of the Clew v2 intelligence pipeline:
//   - Stage 0: Multi-Turn Context Resolution (conditional, uses llm.Client)
//   - Stage 1: Metadata Pre-filter (zero cost, domain type matching)
//   - Stage 2: Embedding Pre-filter (cosine similarity or BM25 fallback)
//   - Stage 3: Haiku Deep Assessment (single LLM call for ranking)
//
// Fail-open chain:
//   - Stage 3 fails -> embedding/BM25 scores only
//   - Stage 2 fails -> BM25 fallback (BC-06: explicit, required)
//   - All fail -> return nil (caller falls back to v1 behavior)
package triage

import (
	"context"
	"time"
)

// ThreadMessage represents a single message in a Slack thread.
// BC-04: Duplicated in triage/, NOT shared with other packages.
// The conversion from conversation.ThreadMessage happens in slack/handler.go.
type ThreadMessage struct {
	Role      string    // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// TriageResult is the output of the triage layer, consumed by Pipeline.QueryWithTriage().
type TriageResult struct {
	// RefinedQuery is the query after multi-turn context resolution.
	// Equal to the original query when Stage 0 is skipped.
	RefinedQuery string

	// Intent is the query classification for retrieval strategy.
	Intent QueryIntent

	// Candidates are the ranked candidates (3-5, descending relevance).
	Candidates []TriageCandidate

	// TriageLatency is the wall-clock triage duration.
	TriageLatency time.Duration

	// ModelCallCount is the number of LLM API calls made during triage.
	ModelCallCount int
}

// TriageCandidate is a single candidate domain selected by the triage layer.
type TriageCandidate struct {
	// QualifiedName is the canonical address: "org::repo::domain".
	QualifiedName string

	// RelevanceScore is the triage relevance score in [0.0, 1.0].
	RelevanceScore float64

	// EmbeddingSimilarity is the cosine similarity from Stage 2.
	// Zero when Stage 2 was skipped or fell back to BM25.
	EmbeddingSimilarity float64

	// Freshness is the domain freshness score in [0.0, 1.0].
	// BC-12: Included from day one, zero-valued in Tier 1 until populated.
	Freshness float64

	// Rationale is the LLM-generated explanation (empty if embedding-only).
	Rationale string

	// DomainType is the domain classification (architecture, scar-tissue, etc.).
	DomainType string

	// RelatedDomains are domains connected via the entity graph.
	RelatedDomains []string
}

// QueryIntent classifies the query for downstream routing.
type QueryIntent struct {
	// Type is the intent classification.
	// Values: "architecture", "debugging", "comparison", "how-to", "exploration"
	Type string

	// TargetDomainTypes are domain types most likely relevant.
	TargetDomainTypes []string

	// Repos are repositories mentioned or implied in the query.
	Repos []string

	// IsFollowUp indicates this query builds on prior thread context.
	IsFollowUp bool
}

// DomainMetadata holds metadata for a single domain, used during triage.
type DomainMetadata struct {
	QualifiedName  string
	DomainType     string
	Repo           string
	FreshnessScore float64
	GeneratedAt    string // RFC3339
}

// BM25Result holds a single BM25 search result for the triage fallback path.
type BM25Result struct {
	QualifiedName string
	Score         float64
	Domain        string
	RawText       string
}

// SearchIndex is the interface the triage orchestrator uses to access search.
// This is a narrow interface adapted from the existing search package, avoiding
// a direct import dependency on the full search.SearchIndex struct.
type SearchIndex interface {
	// SearchByBM25 performs a BM25 search and returns the top-k results.
	SearchByBM25(query string, k int) []BM25Result

	// GetMetadata returns domain metadata by qualified name.
	// Returns false if the domain is not found.
	GetMetadata(qualifiedName string) (*DomainMetadata, bool)

	// ListAllDomains returns metadata for all indexed domains.
	ListAllDomains() []DomainMetadata
}

// EmbeddingModel abstracts embedding computation for Stage 2.
// Sprint 5 uses a stub that returns errors (triggering BM25 fallback).
// Sprint 7 provides the real implementation.
type EmbeddingModel interface {
	// Embed computes an embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float64, error)

	// Dimensions returns the embedding vector dimensionality.
	Dimensions() int
}

// StubEmbeddingModel always returns an error, triggering the BM25 fallback path.
// This is the Sprint 5 implementation -- real embeddings come in Sprint 7.
type StubEmbeddingModel struct{}

// Embed always returns an error to trigger BM25 fallback (BC-06).
func (s *StubEmbeddingModel) Embed(_ context.Context, _ string) ([]float64, error) {
	return nil, errEmbeddingsNotAvailable
}

// Dimensions returns the expected dimensionality for text-embedding-3-small.
func (s *StubEmbeddingModel) Dimensions() int {
	return 1536
}
