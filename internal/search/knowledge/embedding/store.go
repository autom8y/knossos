// Package embedding provides in-memory vector storage with brute-force
// cosine similarity search for the KnowledgeIndex.
//
// At 128 domains with 1536-dimensional vectors, brute-force search completes
// in <1ms. No external vector database is needed for Tier 1.
//
// Sprint 7 scope: Embeddings are generated from Haiku summaries using a
// simple hash-based vector approach for testing. Real embedding API integration
// is a future enhancement.
//
// RR-007: This package MUST NOT import internal/search/ or any parent package.
package embedding

import (
	"math"
	"sort"
)

// EmbeddingEntry holds a single domain's embedding vector.
type EmbeddingEntry struct {
	// QualifiedName is the canonical domain address.
	QualifiedName string `json:"qualified_name"`

	// Vector is the embedding vector.
	Vector []float64 `json:"vector"`

	// SourceHash is the source_hash at the time of embedding generation.
	SourceHash string `json:"source_hash"`
}

// Result holds a single embedding search result.
type Result struct {
	// QualifiedName is the canonical domain address.
	QualifiedName string

	// Similarity is the cosine similarity score in [-1.0, 1.0].
	Similarity float64

	// Freshness is the domain freshness score in [0.0, 1.0].
	// BC-12: Included from day one, zero-valued in Tier 1.
	Freshness float64
}

// Store manages domain embedding vectors with source_hash-based caching.
type Store struct {
	embeddings map[string]*EmbeddingEntry // qualifiedName -> entry
}

// NewStore creates an empty embedding store.
func NewStore() *Store {
	return &Store{
		embeddings: make(map[string]*EmbeddingEntry),
	}
}

// NewStoreFromMap creates a store pre-populated with existing embeddings.
// Used when loading persisted KnowledgeIndex JSON.
func NewStoreFromMap(entries map[string]*EmbeddingEntry) *Store {
	if entries == nil {
		entries = make(map[string]*EmbeddingEntry)
	}
	return &Store{embeddings: entries}
}

// Search finds the top-k domains most similar to the query vector.
// Uses brute-force cosine similarity (<1ms at 128 domains).
// Returns results sorted by similarity descending.
func (s *Store) Search(queryVector []float64, k int) []Result {
	if len(queryVector) == 0 || len(s.embeddings) == 0 {
		return nil
	}

	type scored struct {
		qualifiedName string
		similarity    float64
	}

	var results []scored
	for qn, entry := range s.embeddings {
		if len(entry.Vector) != len(queryVector) {
			continue // Dimension mismatch, skip.
		}
		sim := cosineSimilarity(queryVector, entry.Vector)
		if sim > 0 {
			results = append(results, scored{qualifiedName: qn, similarity: sim})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})

	if k > len(results) {
		k = len(results)
	}

	out := make([]Result, k)
	for i := 0; i < k; i++ {
		out[i] = Result{
			QualifiedName: results[i].qualifiedName,
			Similarity:    results[i].similarity,
			Freshness:     0, // BC-12: zero in Tier 1.
		}
	}
	return out
}

// Add stores or updates an embedding entry.
func (s *Store) Add(qualifiedName string, vector []float64, sourceHash string) {
	s.embeddings[qualifiedName] = &EmbeddingEntry{
		QualifiedName: qualifiedName,
		Vector:        vector,
		SourceHash:    sourceHash,
	}
}

// NeedsRecompute returns true if the embedding for the given domain is missing
// or has a different source_hash than the current one.
func (s *Store) NeedsRecompute(qualifiedName, sourceHash string) bool {
	entry, ok := s.embeddings[qualifiedName]
	if !ok || entry == nil {
		return true
	}
	return entry.SourceHash != sourceHash
}

// Get returns the embedding entry for a domain. Returns nil if not found.
func (s *Store) Get(qualifiedName string) *EmbeddingEntry {
	return s.embeddings[qualifiedName]
}

// All returns all stored embeddings. Used for persistence.
func (s *Store) All() map[string]*EmbeddingEntry {
	return s.embeddings
}

// Count returns the number of stored embeddings.
func (s *Store) Count() int {
	return len(s.embeddings)
}

// cosineSimilarity computes the cosine similarity between two vectors.
// Returns 0 if either vector has zero magnitude.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

// TextToVector generates a simple deterministic embedding from text content.
// This is a hash-based approach for Sprint 7 testing -- not a real embedding model.
// The vector captures term frequency patterns in a fixed-dimensional space.
//
// For production use, replace with a real embedding API (e.g., text-embedding-3-small).
func TextToVector(text string, dimensions int) []float64 {
	if dimensions <= 0 || text == "" {
		return nil
	}

	vec := make([]float64, dimensions)

	// Simple character-based hash distribution across dimensions.
	// This creates vectors that have non-zero cosine similarity for
	// texts with similar character distributions.
	for i, ch := range text {
		idx := (int(ch) * (i + 1)) % dimensions
		if idx < 0 {
			idx = -idx
		}
		vec[idx] += 1.0
	}

	// Normalize to unit vector.
	var mag float64
	for _, v := range vec {
		mag += v * v
	}
	if mag > 0 {
		mag = math.Sqrt(mag)
		for i := range vec {
			vec[i] /= mag
		}
	}

	return vec
}
