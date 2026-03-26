package graph

import (
	"testing"
)

func TestBuild_SameType(t *testing.T) {
	domains := []DomainInfo{
		{QualifiedName: "org::repo-a::arch", DomainType: "architecture", Repo: "repo-a"},
		{QualifiedName: "org::repo-b::arch", DomainType: "architecture", Repo: "repo-b"},
		{QualifiedName: "org::repo-a::scar", DomainType: "scar-tissue", Repo: "repo-a"},
	}

	g := Build(domains)

	// arch in repo-a should have a same_type edge to arch in repo-b.
	edges := g.GetRelationships("org::repo-a::arch")
	found := false
	for _, e := range edges {
		if e.Type == EdgeSameType && e.Target == "org::repo-b::arch" {
			found = true
			if e.Strength != 0.7 {
				t.Errorf("same_type strength = %f, want 0.7", e.Strength)
			}
		}
	}
	if !found {
		t.Error("expected same_type edge from repo-a::arch to repo-b::arch")
	}

	// scar-tissue has no same_type partner (only one domain with that type).
	scarEdges := g.GetRelationships("org::repo-a::scar")
	for _, e := range scarEdges {
		if e.Type == EdgeSameType {
			t.Error("scar-tissue should have no same_type edges (no partner)")
		}
	}
}

func TestBuild_SameType_SameRepoExcluded(t *testing.T) {
	// Two domains with the same type in the SAME repo should NOT get same_type edges.
	domains := []DomainInfo{
		{QualifiedName: "org::repo-a::arch1", DomainType: "architecture", Repo: "repo-a"},
		{QualifiedName: "org::repo-a::arch2", DomainType: "architecture", Repo: "repo-a"},
	}

	g := Build(domains)

	edges := g.GetRelationships("org::repo-a::arch1")
	for _, e := range edges {
		if e.Type == EdgeSameType {
			t.Error("same_type edge should not connect domains in the same repo")
		}
	}
}

func TestBuild_SameRepo(t *testing.T) {
	domains := []DomainInfo{
		{QualifiedName: "org::repo-a::arch", DomainType: "architecture", Repo: "repo-a"},
		{QualifiedName: "org::repo-a::scar", DomainType: "scar-tissue", Repo: "repo-a"},
		{QualifiedName: "org::repo-a::conv", DomainType: "conventions", Repo: "repo-a"},
		{QualifiedName: "org::repo-b::arch", DomainType: "architecture", Repo: "repo-b"},
	}

	g := Build(domains)

	// repo-a has 3 domains, so same_repo strength should be 0.8.
	edges := g.GetRelationships("org::repo-a::arch")
	sameRepoCount := 0
	for _, e := range edges {
		if e.Type == EdgeSameRepo {
			sameRepoCount++
			if e.Strength != 0.8 {
				t.Errorf("same_repo strength for 3-domain repo = %f, want 0.8", e.Strength)
			}
		}
	}
	if sameRepoCount != 2 {
		t.Errorf("same_repo edge count for repo-a::arch = %d, want 2", sameRepoCount)
	}

	// repo-b has only 1 domain, so no same_repo edges.
	edgesB := g.GetRelationships("org::repo-b::arch")
	for _, e := range edgesB {
		if e.Type == EdgeSameRepo {
			t.Error("single-domain repo should have no same_repo edges")
		}
	}
}

func TestBuild_ScopeOverlap(t *testing.T) {
	domains := []DomainInfo{
		{
			QualifiedName: "org::repo::arch",
			DomainType:    "architecture",
			Repo:          "repo",
			SourceScope:   []string{"./internal/**/*.go", "./cmd/**/*.go"},
		},
		{
			QualifiedName: "org::repo::conv",
			DomainType:    "conventions",
			Repo:          "repo",
			SourceScope:   []string{"./internal/**/*.go"},
		},
		{
			QualifiedName: "org::repo::scar",
			DomainType:    "scar-tissue",
			Repo:          "repo",
			SourceScope:   []string{"./tests/**/*.go"},
		},
	}

	g := Build(domains)

	// arch and conv share "./internal/**/*.go" -> should have scope_overlap edge.
	edges := g.GetRelationships("org::repo::arch")
	overlapFound := false
	for _, e := range edges {
		if e.Type == EdgeScopeOverlap && e.Target == "org::repo::conv" {
			overlapFound = true
			if e.Strength <= 0 || e.Strength > 1.0 {
				t.Errorf("scope_overlap strength = %f, expected (0, 1.0]", e.Strength)
			}
		}
	}
	if !overlapFound {
		t.Error("expected scope_overlap edge between arch and conv")
	}

	// scar has no overlap with arch or conv.
	scarEdges := g.GetRelationships("org::repo::scar")
	for _, e := range scarEdges {
		if e.Type == EdgeScopeOverlap {
			t.Errorf("scar should have no scope_overlap edges, got target=%s", e.Target)
		}
	}
}

func TestBuild_Empty(t *testing.T) {
	g := Build(nil)
	if g.EdgeCount() != 0 {
		t.Errorf("EdgeCount() = %d, want 0", g.EdgeCount())
	}
	if g.DomainCount() != 0 {
		t.Errorf("DomainCount() = %d, want 0", g.DomainCount())
	}
}

func TestGetRelationships_NilForMissing(t *testing.T) {
	g := New()
	edges := g.GetRelationships("nonexistent")
	if edges != nil {
		t.Error("expected nil for missing domain")
	}
}

func TestGetRelationships_ReturnsCopy(t *testing.T) {
	g := New()
	g.addEdge("src", Edge{Target: "dst", Type: EdgeSameRepo, Strength: 0.5})

	edges := g.GetRelationships("src")
	edges[0].Strength = 999

	// Original should be unchanged.
	original := g.GetRelationships("src")
	if original[0].Strength == 999 {
		t.Error("GetRelationships() did not return a copy")
	}
}

func TestSameRepoStrength(t *testing.T) {
	tests := []struct {
		size int
		want float64
	}{
		{2, 0.8},
		{3, 0.8},
		{5, 0.6},
		{7, 0.6},
		{10, 0.4},
		{15, 0.4},
		{20, 0.3},
		{50, 0.3},
	}

	for _, tt := range tests {
		got := sameRepoStrength(tt.size)
		if got != tt.want {
			t.Errorf("sameRepoStrength(%d) = %f, want %f", tt.size, got, tt.want)
		}
	}
}

func TestScopeOverlap(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want float64
	}{
		{
			name: "identical",
			a:    []string{"internal/**/*.go"},
			b:    []string{"internal/**/*.go"},
			want: 1.0,
		},
		{
			name: "no overlap",
			a:    []string{"internal/**/*.go"},
			b:    []string{"tests/**/*.go"},
			want: 0,
		},
		{
			name: "prefix containment with extra globs",
			a:    []string{"internal/**/*.go", "cmd/**/*.go"},
			b:    []string{"internal/cmd/**/*.go"},
			want: 0.5, // 1 overlap / (2+1-1) = 1/2 = 0.5
		},
		{
			name: "empty a",
			a:    nil,
			b:    []string{"internal/**/*.go"},
			want: 0,
		},
		{
			name: "empty b",
			a:    []string{"internal/**/*.go"},
			b:    nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scopeOverlap(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("scopeOverlap() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestNoDuplicateEdges(t *testing.T) {
	g := New()
	g.addEdge("src", Edge{Target: "dst", Type: EdgeSameRepo, Strength: 0.5})
	g.addEdge("src", Edge{Target: "dst", Type: EdgeSameRepo, Strength: 0.8})

	edges := g.GetRelationships("src")
	if len(edges) != 1 {
		t.Errorf("expected 1 edge (deduped), got %d", len(edges))
	}
}
