package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/search/knowledge/embedding"
	"github.com/autom8y/knossos/internal/search/knowledge/graph"
	"github.com/autom8y/knossos/internal/search/knowledge/summary"
)

// persistedIndex is the JSON-serializable representation of a KnowledgeIndex.
// This is separate from KnowledgeIndex because the index holds interfaces
// (BM25Searcher) that cannot be serialized.
type persistedIndex struct {
	// Version identifies the persistence schema for forward-compat.
	Version string `json:"version"`

	// Catalog holds domain metadata keyed by qualified name.
	Catalog map[string]*DomainMetadata `json:"catalog"`

	// Summaries holds domain summaries keyed by qualified name.
	Summaries map[string]*summary.DomainSummary `json:"summaries"`

	// Embeddings holds domain embedding vectors keyed by qualified name.
	Embeddings map[string]*embedding.EmbeddingEntry `json:"embeddings"`

	// Edges holds entity graph edges keyed by source qualified name.
	Edges map[string][]graph.Edge `json:"edges"`
}

const persistVersion = "1.0"

// DefaultPersistedPath is the container path for pre-baked KnowledgeIndex JSON.
// BC-11: Pre-baked in container image alongside ContentStore.
const DefaultPersistedPath = "/app/data/knowledge-index.json"

// Save serializes a KnowledgeIndex to JSON at the given path.
// Creates parent directories if needed.
func Save(path string, idx *KnowledgeIndex) error {
	if idx == nil {
		return fmt.Errorf("cannot save nil KnowledgeIndex")
	}

	pi := persistedIndex{
		Version:    persistVersion,
		Catalog:    idx.catalog,
		Summaries:  idx.summaries.All(),
		Embeddings: idx.embeddings.All(),
		Edges:      idx.graph.AllEdges(),
	}

	data, err := json.MarshalIndent(pi, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal knowledge index: %w", err)
	}

	// Ensure parent directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	// Write atomically via temp file.
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write knowledge index: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure.
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename knowledge index: %w", err)
	}

	return nil
}

// Load deserializes a KnowledgeIndex from JSON at the given path.
// The returned KnowledgeIndex has no BM25 index; the caller must set it.
// Target: load <5s for 128 domains.
func Load(path string) (*KnowledgeIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read knowledge index: %w", err)
	}

	var pi persistedIndex
	if err := json.Unmarshal(data, &pi); err != nil {
		return nil, fmt.Errorf("unmarshal knowledge index: %w", err)
	}

	if pi.Version != persistVersion {
		return nil, fmt.Errorf("unsupported knowledge index version: %s (expected %s)",
			pi.Version, persistVersion)
	}

	idx := &KnowledgeIndex{
		bm25:       nil, // Caller must set BM25 index.
		summaries:  summary.NewStoreFromMap(pi.Summaries),
		embeddings: embedding.NewStoreFromMap(pi.Embeddings),
		graph:      graph.NewFromEdges(pi.Edges),
		catalog:    pi.Catalog,
	}

	if idx.catalog == nil {
		idx.catalog = make(map[string]*DomainMetadata)
	}

	return idx, nil
}
