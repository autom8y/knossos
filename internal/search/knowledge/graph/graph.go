// Package graph provides a deterministic entity graph for cross-domain
// relationship discovery in the KnowledgeIndex.
//
// Three edge types are computed from domain metadata (zero LLM cost):
//   - same_type: domains sharing the same domain type across repos
//   - same_repo: domains within the same repository
//   - scope_overlap: domains with overlapping source_scope globs
//
// D-L6: No edge re-computation needed. All edges are metadata-derived.
// When a domain changes, its edges are trivially recomputed from metadata.
//
// RR-007: This package MUST NOT import internal/search/ or any parent package.
package graph

import (
	"sort"
	"strings"

	"github.com/autom8y/knossos/internal/know"
)

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

// Edge represents a directed edge in the entity graph.
type Edge struct {
	// Target is the qualified name of the connected domain.
	Target string `json:"target"`

	// Type is the edge classification.
	Type EdgeType `json:"type"`

	// Strength is the edge weight in [0.0, 1.0].
	Strength float64 `json:"strength"`

	// Evidence is a human-readable explanation of why this edge exists.
	Evidence string `json:"evidence"`
}

// DomainInfo holds the metadata needed for edge computation.
// This is a narrow struct to avoid importing external types.
type DomainInfo struct {
	QualifiedName string
	DomainType    string
	Repo          string
	SourceScope   []string
}

// Graph holds the entity relationship graph.
type Graph struct {
	edges map[string][]Edge // qualifiedName -> outbound edges
}

// New creates an empty graph.
func New() *Graph {
	return &Graph{
		edges: make(map[string][]Edge),
	}
}

// NewFromEdges creates a graph pre-populated with existing edges.
// Used when loading persisted KnowledgeIndex JSON.
func NewFromEdges(edges map[string][]Edge) *Graph {
	if edges == nil {
		edges = make(map[string][]Edge)
	}
	return &Graph{edges: edges}
}

// Build constructs the entity graph from domain metadata.
// All edges are deterministic and computed from metadata fields.
// D-L6: Zero LLM cost.
func Build(domains []DomainInfo) *Graph {
	g := New()

	// Index: domain type -> list of QNs with that type.
	typeIndex := make(map[string][]string)
	// Index: repo -> list of QNs in that repo.
	repoIndex := make(map[string][]string)

	for _, d := range domains {
		if d.DomainType != "" {
			typeIndex[d.DomainType] = append(typeIndex[d.DomainType], d.QualifiedName)
		}
		if d.Repo != "" {
			repoIndex[d.Repo] = append(repoIndex[d.Repo], d.QualifiedName)
		}
	}

	// Edge type 1: same_type — domains with the same type across repos.
	for domainType, qns := range typeIndex {
		if len(qns) < 2 {
			continue
		}
		for i, a := range qns {
			repoA := know.RepoFromQualifiedName(a)
			for j, b := range qns {
				if i == j {
					continue
				}
				repoB := know.RepoFromQualifiedName(b)
				// Only create same_type edges across different repos.
				if repoA == repoB {
					continue
				}
				g.addEdge(a, Edge{
					Target:   b,
					Type:     EdgeSameType,
					Strength: 0.7,
					Evidence: "same domain type: " + domainType,
				})
			}
		}
	}

	// Edge type 2: same_repo — domains in the same repository.
	for repo, qns := range repoIndex {
		if len(qns) < 2 {
			continue
		}
		// Strength decreases with repo size to avoid noise.
		strength := sameRepoStrength(len(qns))
		for i, a := range qns {
			for j, b := range qns {
				if i == j {
					continue
				}
				g.addEdge(a, Edge{
					Target:   b,
					Type:     EdgeSameRepo,
					Strength: strength,
					Evidence: "same repository: " + repo,
				})
			}
		}
	}

	// Edge type 3: scope_overlap — domains with overlapping source_scope globs.
	for i, a := range domains {
		if len(a.SourceScope) == 0 {
			continue
		}
		for j, b := range domains {
			if i == j || len(b.SourceScope) == 0 {
				continue
			}
			overlap := scopeOverlap(a.SourceScope, b.SourceScope)
			if overlap > 0 {
				g.addEdge(a.QualifiedName, Edge{
					Target:   b.QualifiedName,
					Type:     EdgeScopeOverlap,
					Strength: overlap,
					Evidence: "overlapping source scope",
				})
			}
		}
	}

	return g
}

// GetRelationships returns all outbound edges for a domain.
// Returns nil if no edges exist.
func (g *Graph) GetRelationships(qualifiedName string) []Edge {
	edges := g.edges[qualifiedName]
	if len(edges) == 0 {
		return nil
	}
	// Return a copy to prevent mutation.
	out := make([]Edge, len(edges))
	copy(out, edges)
	return out
}

// AllEdges returns the full edge map. Used for persistence.
func (g *Graph) AllEdges() map[string][]Edge {
	return g.edges
}

// EdgeCount returns the total number of edges across all domains.
func (g *Graph) EdgeCount() int {
	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}

// DomainCount returns the number of domains with at least one edge.
func (g *Graph) DomainCount() int {
	return len(g.edges)
}

// addEdge appends an edge, avoiding duplicates (same target + type).
func (g *Graph) addEdge(source string, edge Edge) {
	for _, existing := range g.edges[source] {
		if existing.Target == edge.Target && existing.Type == edge.Type {
			return // Skip duplicate.
		}
	}
	g.edges[source] = append(g.edges[source], edge)
}

// sameRepoStrength computes edge strength inversely proportional to repo size.
// Small repos (2-3 domains) get strength 0.8, large repos (20+) get strength 0.3.
func sameRepoStrength(repoSize int) float64 {
	switch {
	case repoSize <= 3:
		return 0.8
	case repoSize <= 7:
		return 0.6
	case repoSize <= 15:
		return 0.4
	default:
		return 0.3
	}
}

// scopeOverlap computes the Jaccard similarity of two source_scope glob sets.
// Returns 0 if no overlap, up to 1.0 for identical sets.
//
// This uses simple string matching on glob patterns, not full glob expansion.
// Patterns like "./internal/**/*.go" and "./internal/cmd/**/*.go" are detected
// as overlapping via prefix containment.
func scopeOverlap(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	// Normalize globs: strip leading "./"
	normA := normalizeGlobs(a)
	normB := normalizeGlobs(b)

	// Count overlaps via exact match or prefix containment.
	overlapCount := 0
	for _, ga := range normA {
		for _, gb := range normB {
			if ga == gb || isGlobPrefix(ga, gb) || isGlobPrefix(gb, ga) {
				overlapCount++
				break
			}
		}
	}

	if overlapCount == 0 {
		return 0
	}

	// Jaccard-like: overlap / union.
	union := len(normA) + len(normB) - overlapCount
	if union <= 0 {
		return 0
	}

	return float64(overlapCount) / float64(union)
}

// normalizeGlobs strips leading "./" from glob patterns and sorts them.
func normalizeGlobs(globs []string) []string {
	out := make([]string, len(globs))
	for i, g := range globs {
		out[i] = strings.TrimPrefix(g, "./")
	}
	sort.Strings(out)
	return out
}

// isGlobPrefix returns true if pattern a is a prefix path of pattern b.
// e.g., "internal/**/*.go" is a prefix of "internal/cmd/**/*.go".
func isGlobPrefix(a, b string) bool {
	// Strip glob wildcards to get the directory prefix.
	dirA := stripGlobWildcards(a)
	dirB := stripGlobWildcards(b)

	if dirA == "" || dirB == "" {
		return false
	}

	// a's directory must be a prefix of b's directory.
	return strings.HasPrefix(dirB, dirA)
}

// stripGlobWildcards returns the directory portion before any wildcards.
func stripGlobWildcards(pattern string) string {
	// Find first wildcard character.
	for i, ch := range pattern {
		if ch == '*' || ch == '?' || ch == '[' {
			// Return everything up to the last path separator before the wildcard.
			prefix := pattern[:i]
			lastSlash := strings.LastIndex(prefix, "/")
			if lastSlash >= 0 {
				return prefix[:lastSlash+1]
			}
			return ""
		}
	}
	// No wildcards — return the full path as a directory prefix.
	lastSlash := strings.LastIndex(pattern, "/")
	if lastSlash >= 0 {
		return pattern[:lastSlash+1]
	}
	return pattern
}
