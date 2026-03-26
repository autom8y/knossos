package embedding

import (
	"math"
	"testing"
)

func TestStore_Search(t *testing.T) {
	s := NewStore()

	// Add some embeddings with known vectors.
	s.Add("org::repo::arch", []float64{1, 0, 0}, "hash1")
	s.Add("org::repo::scar", []float64{0, 1, 0}, "hash2")
	s.Add("org::repo::conv", []float64{0.7, 0.7, 0}, "hash3")

	tests := []struct {
		name      string
		query     []float64
		k         int
		wantFirst string
		wantCount int
	}{
		{
			name:      "exact match",
			query:     []float64{1, 0, 0},
			k:         3,
			wantFirst: "org::repo::arch",
			wantCount: 2, // arch (1.0) and conv (~0.7), scar is 0
		},
		{
			name:      "partial match",
			query:     []float64{0.5, 0.5, 0},
			k:         3,
			wantFirst: "org::repo::conv",
			wantCount: 3,
		},
		{
			name:      "k limits results",
			query:     []float64{0.5, 0.5, 0},
			k:         1,
			wantFirst: "org::repo::conv",
			wantCount: 1,
		},
		{
			name:      "empty query",
			query:     []float64{},
			k:         3,
			wantCount: 0,
		},
		{
			name:      "orthogonal query returns nothing",
			query:     []float64{0, 0, 1},
			k:         3,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := s.Search(tt.query, tt.k)
			if len(results) != tt.wantCount {
				t.Errorf("Search() count = %d, want %d", len(results), tt.wantCount)
			}
			if tt.wantFirst != "" && len(results) > 0 {
				if results[0].QualifiedName != tt.wantFirst {
					t.Errorf("Search() first = %q, want %q", results[0].QualifiedName, tt.wantFirst)
				}
			}
			// BC-12: Freshness should be zero in Tier 1.
			for _, r := range results {
				if r.Freshness != 0 {
					t.Errorf("Freshness = %f, want 0 (BC-12)", r.Freshness)
				}
			}
		})
	}
}

func TestStore_Search_SortedDescending(t *testing.T) {
	s := NewStore()
	s.Add("a", []float64{1, 0}, "h1")
	s.Add("b", []float64{0.8, 0.2}, "h2")
	s.Add("c", []float64{0.5, 0.5}, "h3")

	results := s.Search([]float64{1, 0}, 3)

	for i := 1; i < len(results); i++ {
		if results[i].Similarity > results[i-1].Similarity {
			t.Errorf("results not sorted descending: [%d]=%f > [%d]=%f",
				i, results[i].Similarity, i-1, results[i-1].Similarity)
		}
	}
}

func TestStore_NeedsRecompute(t *testing.T) {
	tests := []struct {
		name          string
		qualifiedName string
		sourceHash    string
		preload       map[string]*EmbeddingEntry
		want          bool
	}{
		{
			name:          "missing needs recompute",
			qualifiedName: "org::repo::new",
			sourceHash:    "hash1",
			preload:       map[string]*EmbeddingEntry{},
			want:          true,
		},
		{
			name:          "hash match skips recompute",
			qualifiedName: "org::repo::arch",
			sourceHash:    "abc123",
			preload: map[string]*EmbeddingEntry{
				"org::repo::arch": {
					QualifiedName: "org::repo::arch",
					Vector:        []float64{1, 0, 0},
					SourceHash:    "abc123",
				},
			},
			want: false,
		},
		{
			name:          "hash mismatch needs recompute",
			qualifiedName: "org::repo::arch",
			sourceHash:    "new-hash",
			preload: map[string]*EmbeddingEntry{
				"org::repo::arch": {
					QualifiedName: "org::repo::arch",
					Vector:        []float64{1, 0, 0},
					SourceHash:    "old-hash",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStoreFromMap(tt.preload)
			got := s.NeedsRecompute(tt.qualifiedName, tt.sourceHash)
			if got != tt.want {
				t.Errorf("NeedsRecompute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Add(t *testing.T) {
	s := NewStore()
	s.Add("org::repo::test", []float64{1, 2, 3}, "hash1")

	if s.Count() != 1 {
		t.Errorf("Count() = %d, want 1", s.Count())
	}

	entry := s.Get("org::repo::test")
	if entry == nil {
		t.Fatal("Get() returned nil after Add()")
	}
	if entry.SourceHash != "hash1" {
		t.Errorf("SourceHash = %q, want %q", entry.SourceHash, "hash1")
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{
			name: "identical",
			a:    []float64{1, 0, 0},
			b:    []float64{1, 0, 0},
			want: 1.0,
		},
		{
			name: "orthogonal",
			a:    []float64{1, 0, 0},
			b:    []float64{0, 1, 0},
			want: 0.0,
		},
		{
			name: "opposite",
			a:    []float64{1, 0, 0},
			b:    []float64{-1, 0, 0},
			want: -1.0,
		},
		{
			name: "45 degrees",
			a:    []float64{1, 0},
			b:    []float64{1, 1},
			want: 1.0 / math.Sqrt(2),
		},
		{
			name: "empty",
			a:    []float64{},
			b:    []float64{},
			want: 0.0,
		},
		{
			name: "dimension mismatch",
			a:    []float64{1, 0},
			b:    []float64{1, 0, 0},
			want: 0.0,
		},
		{
			name: "zero vector",
			a:    []float64{0, 0, 0},
			b:    []float64{1, 0, 0},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-10 {
				t.Errorf("cosineSimilarity() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestTextToVector(t *testing.T) {
	tests := []struct {
		name string
		text string
		dims int
		want int // expected vector length (0 for nil)
	}{
		{
			name: "normal text",
			text: "architecture overview",
			dims: 64,
			want: 64,
		},
		{
			name: "empty text",
			text: "",
			dims: 64,
			want: 0,
		},
		{
			name: "zero dimensions",
			text: "hello",
			dims: 0,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := TextToVector(tt.text, tt.dims)
			if tt.want == 0 {
				if vec != nil {
					t.Errorf("TextToVector() = %v, want nil", vec)
				}
				return
			}
			if len(vec) != tt.want {
				t.Errorf("TextToVector() len = %d, want %d", len(vec), tt.want)
			}

			// Verify unit vector (magnitude ~1.0).
			var mag float64
			for _, v := range vec {
				mag += v * v
			}
			mag = math.Sqrt(mag)
			if math.Abs(mag-1.0) > 1e-10 {
				t.Errorf("TextToVector() magnitude = %f, want 1.0", mag)
			}
		})
	}

	// Same text should produce same vector (deterministic).
	t.Run("deterministic", func(t *testing.T) {
		v1 := TextToVector("test content", 64)
		v2 := TextToVector("test content", 64)
		for i := range v1 {
			if v1[i] != v2[i] {
				t.Errorf("TextToVector not deterministic at index %d: %f != %f", i, v1[i], v2[i])
				break
			}
		}
	})

	// Different texts should produce different vectors.
	t.Run("different texts different vectors", func(t *testing.T) {
		v1 := TextToVector("architecture overview", 64)
		v2 := TextToVector("scar tissue and bugs", 64)
		same := true
		for i := range v1 {
			if v1[i] != v2[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("TextToVector produced identical vectors for different texts")
		}
	})

	// Similar texts should have positive cosine similarity.
	t.Run("similar texts positive similarity", func(t *testing.T) {
		v1 := TextToVector("Go architecture and package structure", 64)
		v2 := TextToVector("Go architecture and module layout", 64)
		sim := cosineSimilarity(v1, v2)
		if sim <= 0 {
			t.Errorf("expected positive similarity for similar texts, got %f", sim)
		}
	})
}
