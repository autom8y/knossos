package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- Test helpers ---

type mockCatalog struct {
	domains []CatalogDomainEntry
}

func (m *mockCatalog) ListDomains() []CatalogDomainEntry { return m.domains }
func (m *mockCatalog) LookupDomain(qn string) (CatalogDomainEntry, bool) {
	for _, d := range m.domains {
		if d.QualifiedName == qn {
			return d, true
		}
	}
	return CatalogDomainEntry{}, false
}
func (m *mockCatalog) DomainCount() int { return len(m.domains) }

type mockContentStore struct {
	content map[string]string
}

func (m *mockContentStore) LoadContent(qn string) (string, error) {
	c, ok := m.content[qn]
	if !ok {
		return "", os.ErrNotExist
	}
	return c, nil
}

func (m *mockContentStore) HasContent(qn string) bool {
	_, ok := m.content[qn]
	return ok
}

type mockLLMClient struct {
	response string
	err      error
	calls    int
}

func (m *mockLLMClient) Complete(_ context.Context, _, _ string, _ int) (string, error) {
	m.calls++
	return m.response, m.err
}

type mockBM25Searcher struct {
	docs     []BM25SearchHit
	sections []BM25SearchHit
}

func (m *mockBM25Searcher) SearchDocuments(_ string, k int) []BM25SearchHit {
	if k > len(m.docs) {
		k = len(m.docs)
	}
	return m.docs[:k]
}

func (m *mockBM25Searcher) SearchSections(_ string, k int) []BM25SearchHit {
	if k > len(m.sections) {
		k = len(m.sections)
	}
	return m.sections[:k]
}

// --- Build tests ---

func TestBuild_EmptyCatalog(t *testing.T) {
	idx, err := Build(context.Background(), BuildConfig{
		Catalog: &mockCatalog{domains: nil},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if idx.DomainCount() != 0 {
		t.Errorf("DomainCount() = %d, want 0", idx.DomainCount())
	}
}

func TestBuild_NilCatalog(t *testing.T) {
	_, err := Build(context.Background(), BuildConfig{})
	if err == nil {
		t.Fatal("Build() with nil catalog should return error")
	}
}

func TestBuild_WithContent(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{
				QualifiedName: "org::repo::arch",
				Domain:        "architecture",
				SourceHash:    "hash1",
				GeneratedAt:   "2026-03-26T00:00:00Z",
			},
			{
				QualifiedName: "org::repo::scar",
				Domain:        "scar-tissue",
				SourceHash:    "hash2",
				GeneratedAt:   "2026-03-26T00:00:00Z",
			},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "# Architecture\n\n## Package Structure\n\nGo packages layout.\n\n## Data Flow\n\nPipeline stages.",
			"org::repo::scar": "# Scar Tissue\n\n## Past Bugs\n\nCritical bugs found.",
		},
	}

	llm := &mockLLMClient{
		response: "Architecture domain covers Go packages and data flow.\n\nSECTION: package-structure | Describes Go package layout.\nSECTION: data-flow | Pipeline stage descriptions.",
	}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llm,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if idx.DomainCount() != 2 {
		t.Errorf("DomainCount() = %d, want 2", idx.DomainCount())
	}

	// LLM should be called for each domain.
	if llm.calls != 2 {
		t.Errorf("LLM calls = %d, want 2", llm.calls)
	}

	// Summaries should be populated.
	if idx.SummaryCount() != 2 {
		t.Errorf("SummaryCount() = %d, want 2", idx.SummaryCount())
	}

	// Embeddings should be populated.
	if idx.EmbeddingCount() != 2 {
		t.Errorf("EmbeddingCount() = %d, want 2", idx.EmbeddingCount())
	}

	// Metadata should be retrievable.
	meta, ok := idx.GetMetadata("org::repo::arch")
	if !ok {
		t.Fatal("GetMetadata() returned false for indexed domain")
	}
	if meta.DomainType != "architecture" {
		t.Errorf("DomainType = %q, want %q", meta.DomainType, "architecture")
	}
	if meta.SourceHash != "hash1" {
		t.Errorf("SourceHash = %q, want %q", meta.SourceHash, "hash1")
	}
}

func TestBuild_WithBM25(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "h1"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture content.",
		},
	}

	bm25 := &mockBM25Searcher{
		docs: []BM25SearchHit{
			{QualifiedName: "org::repo::arch", Score: 5.0, Domain: "architecture", RawText: "Content"},
		},
	}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		BM25Index:    bm25,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// BC-05: BM25 search should delegate to the wrapped index.
	results := idx.SearchByBM25("architecture", 5)
	if len(results) != 1 {
		t.Errorf("SearchByBM25() count = %d, want 1", len(results))
	}
	if results[0].QualifiedName != "org::repo::arch" {
		t.Errorf("SearchByBM25() first = %q, want %q", results[0].QualifiedName, "org::repo::arch")
	}
}

func TestBuild_IncrementalSkipsUnchanged(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "hash1"},
			{QualifiedName: "org::repo::scar", Domain: "scar-tissue", SourceHash: "hash2"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture content.",
			"org::repo::scar": "Scar content.",
		},
	}

	llm := &mockLLMClient{response: "Summary text."}

	// First build: indexes all domains.
	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llm,
	})
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}

	// Persist to temp file.
	tmpDir := t.TempDir()
	persistPath := filepath.Join(tmpDir, "ki.json")
	if err := Save(persistPath, idx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reset LLM call counter.
	llm.calls = 0

	// Second build with same hashes: should skip all domains.
	_, err = Build(context.Background(), BuildConfig{
		Catalog:       catalog,
		ContentStore:  contentStore,
		LLMClient:     llm,
		PersistedPath: persistPath,
	})
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}

	if llm.calls != 0 {
		t.Errorf("LLM calls on unchanged rebuild = %d, want 0", llm.calls)
	}
}

func TestBuild_IncrementalReindexesChanged(t *testing.T) {
	// First build.
	catalog1 := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "hash1"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture content.",
		},
	}

	llm := &mockLLMClient{response: "Summary text."}

	tmpDir := t.TempDir()
	persistPath := filepath.Join(tmpDir, "ki.json")

	_, err := Build(context.Background(), BuildConfig{
		Catalog:       catalog1,
		ContentStore:  contentStore,
		LLMClient:     llm,
		PersistedPath: persistPath,
	})
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}

	// Second build with changed hash.
	catalog2 := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "hash2-changed"},
		},
	}

	llm.calls = 0

	_, err = Build(context.Background(), BuildConfig{
		Catalog:       catalog2,
		ContentStore:  contentStore,
		LLMClient:     llm,
		PersistedPath: persistPath,
	})
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}

	if llm.calls != 1 {
		t.Errorf("LLM calls on changed domain = %d, want 1", llm.calls)
	}
}

// --- Coordinator method tests ---

func TestKnowledgeIndex_SearchByEmbedding(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "h1"},
			{QualifiedName: "org::repo::scar", Domain: "scar-tissue", SourceHash: "h2"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture and Go packages.",
			"org::repo::scar": "Past bugs and defensive patterns.",
		},
	}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Generate a query vector from text similar to "arch".
	archContent, _ := contentStore.LoadContent("org::repo::arch")
	queryVec := make([]float64, embeddingDimensions)
	// Use the embedding module's text-to-vector for the query.
	// We expect arch to be more similar to arch-like query.
	from := truncateForEmbedding(archContent, 2000)
	if from != "" {
		// Simple: search with the arch embedding itself.
		entry := idx.embeddings.Get("org::repo::arch")
		if entry != nil {
			queryVec = entry.Vector
		}
	}

	results := idx.SearchByEmbedding(queryVec, 5)
	if len(results) == 0 {
		t.Fatal("SearchByEmbedding() returned empty results")
	}

	// First result should be arch (exact match).
	if results[0].QualifiedName != "org::repo::arch" {
		t.Errorf("first result = %q, want %q", results[0].QualifiedName, "org::repo::arch")
	}

	// BC-12: Freshness should be zero.
	for _, r := range results {
		if r.Freshness != 0 {
			t.Errorf("Freshness = %f, want 0 (BC-12)", r.Freshness)
		}
	}
}

func TestKnowledgeIndex_GetSummary(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "h1"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture content.",
		},
	}

	llm := &mockLLMClient{response: "A summary of the architecture domain."}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llm,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	summary, ok := idx.GetSummary("org::repo::arch")
	if !ok {
		t.Fatal("GetSummary() returned false")
	}
	if summary == "" {
		t.Error("GetSummary() returned empty string")
	}

	// Missing domain.
	_, ok = idx.GetSummary("nonexistent")
	if ok {
		t.Error("GetSummary() returned true for nonexistent domain")
	}
}

func TestKnowledgeIndex_GetRelationships(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo-a::arch", Domain: "architecture", SourceHash: "h1"},
			{QualifiedName: "org::repo-b::arch", Domain: "architecture", SourceHash: "h2"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo-a::arch": "Content A.",
			"org::repo-b::arch": "Content B.",
		},
	}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	rels := idx.GetRelationships("org::repo-a::arch")
	if len(rels) == 0 {
		t.Fatal("expected relationships for same domain type across repos")
	}

	// Should have a same_type edge to repo-b::arch.
	found := false
	for _, r := range rels {
		if r.Target == "org::repo-b::arch" && r.Type == EdgeSameType {
			found = true
		}
	}
	if !found {
		t.Error("expected same_type edge from repo-a::arch to repo-b::arch")
	}
}

func TestKnowledgeIndex_NeedsReindex(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "h1"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Content.",
		},
	}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Same hash: no reindex needed.
	if idx.NeedsReindex("org::repo::arch", "h1") {
		t.Error("NeedsReindex() returned true for matching hash")
	}

	// Different hash: reindex needed.
	if !idx.NeedsReindex("org::repo::arch", "h2-new") {
		t.Error("NeedsReindex() returned false for changed hash")
	}

	// Unknown domain: reindex needed.
	if !idx.NeedsReindex("org::repo::unknown", "any") {
		t.Error("NeedsReindex() returned false for unknown domain")
	}
}

// --- Persistence tests ---

func TestSaveLoad_RoundTrip(t *testing.T) {
	catalog := &mockCatalog{
		domains: []CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "h1"},
			{QualifiedName: "org::repo::scar", Domain: "scar-tissue", SourceHash: "h2"},
		},
	}

	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture content with sections.\n\n## Overview\n\nThe overview section.",
			"org::repo::scar": "Scar tissue content.",
		},
	}

	llm := &mockLLMClient{response: "Domain summary text.\n\nSECTION: overview | Overview section summary."}

	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llm,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ki.json")

	// Save.
	if err := Save(path, idx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat() error = %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("persisted file is empty")
	}

	// Load.
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded state matches.
	if loaded.DomainCount() != idx.DomainCount() {
		t.Errorf("loaded DomainCount = %d, want %d", loaded.DomainCount(), idx.DomainCount())
	}
	if loaded.SummaryCount() != idx.SummaryCount() {
		t.Errorf("loaded SummaryCount = %d, want %d", loaded.SummaryCount(), idx.SummaryCount())
	}
	if loaded.EmbeddingCount() != idx.EmbeddingCount() {
		t.Errorf("loaded EmbeddingCount = %d, want %d", loaded.EmbeddingCount(), idx.EmbeddingCount())
	}

	// Metadata round-trip.
	meta, ok := loaded.GetMetadata("org::repo::arch")
	if !ok {
		t.Fatal("loaded index missing org::repo::arch metadata")
	}
	if meta.SourceHash != "h1" {
		t.Errorf("loaded SourceHash = %q, want %q", meta.SourceHash, "h1")
	}

	// Summary round-trip.
	summary, ok := loaded.GetSummary("org::repo::arch")
	if !ok {
		t.Fatal("loaded index missing org::repo::arch summary")
	}
	if summary == "" {
		t.Error("loaded summary is empty")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/ki.json")
	if err == nil {
		t.Fatal("Load() should return error for missing file")
	}
}

func TestSave_NilIndex(t *testing.T) {
	err := Save("/tmp/test-ki.json", nil)
	if err == nil {
		t.Fatal("Save() should return error for nil index")
	}
}

// --- Helper function tests ---

func TestParseSections(t *testing.T) {
	content := "# Title\n\nIntro text.\n\n## Overview\n\nOverview body.\n\n## Details\n\nDetails body.\n"

	sections := parseSections(content)
	if len(sections) != 2 {
		t.Errorf("parseSections() count = %d, want 2", len(sections))
	}

	if _, ok := sections["overview"]; !ok {
		t.Error("missing 'overview' section")
	}
	if _, ok := sections["details"]; !ok {
		t.Error("missing 'details' section")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Package Structure", "package-structure"},
		{"Data Flow & Patterns", "data-flow--patterns"},
		{"", ""},
		{"A Very Long Heading That Exceeds The Maximum Allowed Characters For Slugification Purposes", "a-very-long-heading-that-exceeds-the-maximum-allowed-charact"},
	}

	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRepoFromQualifiedName(t *testing.T) {
	tests := []struct {
		qn   string
		want string
	}{
		{"org::repo::domain", "repo"},
		{"org::knossos::arch", "knossos"},
		{"single", ""},
		{"org::repo", "repo"},
	}

	for _, tt := range tests {
		got := repoFromQualifiedName(tt.qn)
		if got != tt.want {
			t.Errorf("repoFromQualifiedName(%q) = %q, want %q", tt.qn, got, tt.want)
		}
	}
}
