package bm25

import (
	"fmt"
	"testing"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockContentLoader is a test content loader that returns predefined content.
type mockContentLoader struct {
	content map[string]string // QualifiedName -> content
}

func (m *mockContentLoader) LoadContent(entry registryorg.DomainEntry) (string, error) {
	c, ok := m.content[entry.QualifiedName]
	if !ok {
		return "", fmt.Errorf("not found: %s", entry.QualifiedName)
	}
	return c, nil
}

func testCatalog() *registryorg.DomainCatalog {
	return &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Repos: []registryorg.RepoEntry{
			{
				Name: "repo1",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "testorg::repo1::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   "2026-03-23T18:00:00Z",
					},
					{
						QualifiedName: "testorg::repo1::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
						GeneratedAt:   "2026-03-23T18:00:00Z",
					},
				},
			},
			{
				Name: "repo2",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "testorg::repo2::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   "2026-03-22T14:00:00Z",
					},
				},
			},
		},
	}
}

func TestBuildFromCatalog_Basic(t *testing.T) {
	catalog := testCatalog()
	loader := &mockContentLoader{
		content: map[string]string{
			"testorg::repo1::architecture": `# Architecture

## Package Structure

The internal package contains all domain logic.

## Layer Boundaries

Cmd imports domain but domain never imports cmd.
`,
			"testorg::repo1::conventions": `# Conventions

## Error Handling

Use domain errors package for all errors.
`,
			"testorg::repo2::architecture": `# Architecture

## Service Design

Microservice architecture with gRPC boundaries.
`,
		},
	}

	idx, err := BuildFromCatalog(catalog, loader)
	require.NoError(t, err)

	assert.Equal(t, 3, idx.TotalDocs)
	assert.Greater(t, idx.TotalSecs, 0)
	assert.Greater(t, idx.AvgDocLen, 0.0)

	// Verify document search works.
	results := idx.SearchDocuments("architecture package", 5)
	assert.NotEmpty(t, results)

	// Verify section search works.
	secResults := idx.SearchSections("error handling domain", 5)
	assert.NotEmpty(t, secResults)
}

func TestBuildFromCatalog_FailOpen(t *testing.T) {
	catalog := testCatalog()
	// Loader that fails for everything.
	loader := &mockContentLoader{content: map[string]string{}}

	idx, err := BuildFromCatalog(catalog, loader)
	require.NoError(t, err)
	assert.Equal(t, 0, idx.TotalDocs, "should have 0 docs when all loads fail")
}

func TestBuildFromCatalog_NilCatalog(t *testing.T) {
	_, err := BuildFromCatalog(nil, &mockContentLoader{})
	assert.Error(t, err)
}

func TestBuildFromCatalog_EmptyContent(t *testing.T) {
	catalog := &registryorg.DomainCatalog{
		Repos: []registryorg.RepoEntry{
			{
				Name: "repo1",
				Domains: []registryorg.DomainEntry{
					{QualifiedName: "org::repo1::empty", Domain: "empty", Path: ".know/empty.md"},
				},
			},
		},
	}
	loader := &mockContentLoader{
		content: map[string]string{
			"org::repo1::empty": "  \n  ",
		},
	}

	idx, err := BuildFromCatalog(catalog, loader)
	require.NoError(t, err)
	assert.Equal(t, 0, idx.TotalDocs, "empty content should be skipped")
}

// TestClewBM25_SpecialistCompetitiveness verifies that Clew BM25 params make
// specialist domains (conventions, scar-tissue) competitive with longer
// architecture documents for queries matching specialist vocabulary.
//
// This is the core Sprint 1 validation: with b=0.55, a shorter conventions
// document with high term density for "error handling" should rank comparably
// to a longer architecture document for that query.
func TestClewBM25_SpecialistCompetitiveness(t *testing.T) {
	catalog := &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Repos: []registryorg.RepoEntry{
			{
				Name: "repo1",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "testorg::repo1::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
					},
					{
						QualifiedName: "testorg::repo1::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
					},
					{
						QualifiedName: "testorg::repo1::scar-tissue",
						Domain:        "scar-tissue",
						Path:          ".know/scar-tissue.md",
					},
				},
			},
		},
	}

	// Architecture: long document (2x+ specialist length) with broad vocabulary.
	archContent := `# Architecture

## Package Structure

The internal package contains all domain logic including error handling utilities,
session management, search indexing, and materialization pipelines. The package
structure follows a hub-and-leaf pattern where hub packages import many siblings
and leaf packages import only stdlib.

## Layer Boundaries

Cmd imports domain but domain never imports cmd. The reason package orchestrates
trust scoring, context assembly, and response generation. Error handling flows
through the dual-layer system from internal/errors.

## Data Flow

Search results flow through BM25 indexing, RRF fusion, and triage stages before
reaching the assembler for token-budgeted context packing. The pipeline handles
error cases via fail-open patterns throughout.
`

	// Conventions: shorter document but dense in error handling vocabulary.
	convContent := `# Conventions

## Error Handling Style

The project uses a dual-layer error system. Domain errors are structured and
JSON-serializable. Use errors.New for domain errors, errors.Wrap for wrapping,
and named constructors like ErrSessionNotFound for common error types.

## Error Propagation

Immediate return pattern dominates. All error handling uses if err != nil with
domain error wrapping. Never swallow errors silently.
`

	// Scar-tissue: short document about past failures.
	scarContent := `# Scar Tissue

## Past Regressions

A previous incident where mocked tests passed but production migration failed.
Error handling in the streaming path silently dropped connection errors.
The fix added explicit error propagation through the async pipeline.
`

	loader := &mockContentLoader{
		content: map[string]string{
			"testorg::repo1::architecture": archContent,
			"testorg::repo1::conventions":  convContent,
			"testorg::repo1::scar-tissue":  scarContent,
		},
	}

	// Build with default params (ari ask: b=0.25).
	defaultIdx, err := BuildFromCatalog(catalog, loader)
	require.NoError(t, err)

	// Build with Clew params (b=0.55).
	clewIdx, err := BuildFromCatalogWithScorer(catalog, loader, NewClewBM25())
	require.NoError(t, err)

	// Query: "error handling" — should favor conventions over architecture.
	query := "error handling patterns"

	defaultResults := defaultIdx.SearchDocuments(query, 5)
	clewResults := clewIdx.SearchDocuments(query, 5)

	require.NotEmpty(t, defaultResults, "default index should return results")
	require.NotEmpty(t, clewResults, "clew index should return results")

	// With Clew params, conventions should rank in top 2 for "error handling".
	clewTop2Domains := make(map[string]bool)
	for i := 0; i < 2 && i < len(clewResults); i++ {
		clewTop2Domains[clewResults[i].Domain] = true
	}
	assert.True(t, clewTop2Domains["conventions"],
		"Clew BM25 should rank conventions in top 2 for 'error handling'; got top results: %v",
		clewResults)

	// The Clew conventions score should be closer to (or exceed) architecture score
	// compared to the default index.
	var defaultConvScore, defaultArchScore float64
	for _, r := range defaultResults {
		if r.Domain == "conventions" && defaultConvScore == 0 {
			defaultConvScore = r.Score
		}
		if r.Domain == "architecture" && defaultArchScore == 0 {
			defaultArchScore = r.Score
		}
	}

	var clewConvScore, clewArchScore float64
	for _, r := range clewResults {
		if r.Domain == "conventions" && clewConvScore == 0 {
			clewConvScore = r.Score
		}
		if r.Domain == "architecture" && clewArchScore == 0 {
			clewArchScore = r.Score
		}
	}

	// Clew's conventions-to-architecture ratio should be higher than default's.
	if defaultArchScore > 0 && clewArchScore > 0 {
		defaultRatio := defaultConvScore / defaultArchScore
		clewRatio := clewConvScore / clewArchScore
		assert.Greater(t, clewRatio, defaultRatio,
			"Clew BM25 should improve conventions competitiveness vs architecture; "+
				"default ratio=%.3f, clew ratio=%.3f", defaultRatio, clewRatio)
	}
}

// TestBuildFromCatalogWithScorer_Isolation verifies that the custom scorer is
// used by the built index and does not affect other indexes.
func TestBuildFromCatalogWithScorer_Isolation(t *testing.T) {
	// Use documents with significantly different lengths so b-parameter
	// differences produce observable score differences.
	catalog := &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Repos: []registryorg.RepoEntry{
			{
				Name: "repo1",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "testorg::repo1::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
					},
					{
						QualifiedName: "testorg::repo1::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
					},
				},
			},
		},
	}
	loader := &mockContentLoader{
		content: map[string]string{
			// Long document (architecture is ~4x conventions length).
			"testorg::repo1::architecture": `# Architecture

## Package Structure

The internal package contains all domain logic including search indexing
and materialization pipelines. The codebase follows hub-and-leaf patterns
with explicit import boundaries documented in source. All domain logic
resides under internal while cmd provides CLI surface only.

## Layer Boundaries

Commands import domain packages but domain packages never import commands.
The reason package orchestrates trust scoring and context assembly. Search
results flow through multiple triage stages before context packing.

## Data Flow

The pipeline processes queries through BM25 indexing, embedding search,
reciprocal rank fusion, and multi-stage triage assessment before reaching
the assembler for token-budgeted context window construction.
`,
			// Short document.
			"testorg::repo1::conventions": `# Conventions

## Errors

Domain error wrapping patterns for search indexing.
`,
		},
	}

	// Build two indexes with different b params.
	lowB, err := BuildFromCatalogWithScorer(catalog, loader, NewBM25WithParams(1.2, 0.10))
	require.NoError(t, err)

	highB, err := BuildFromCatalogWithScorer(catalog, loader, NewBM25WithParams(1.2, 0.90))
	require.NoError(t, err)

	assert.Equal(t, lowB.TotalDocs, highB.TotalDocs)

	// Query that matches the long architecture document.
	query := "package structure domain logic"
	lowResults := lowB.SearchDocuments(query, 3)
	highResults := highB.SearchDocuments(query, 3)

	require.NotEmpty(t, lowResults)
	require.NotEmpty(t, highResults)

	// Find architecture scores in both — high b should penalize the long doc more.
	var lowArchScore, highArchScore float64
	for _, r := range lowResults {
		if r.Domain == "architecture" {
			lowArchScore = r.Score
		}
	}
	for _, r := range highResults {
		if r.Domain == "architecture" {
			highArchScore = r.Score
		}
	}

	assert.Greater(t, lowArchScore, highArchScore,
		"low b should score long architecture doc higher than high b; low=%.3f, high=%.3f",
		lowArchScore, highArchScore)
}

// TestDomainTypeVocabulary_Differentiation verifies that the metadata
// amplification injects domain-type-specific vocabulary that differentiates
// domain types in BM25 scoring.
func TestDomainTypeVocabulary_Differentiation(t *testing.T) {
	catalog := &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Repos: []registryorg.RepoEntry{
			{
				Name: "repo1",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "testorg::repo1::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
					},
					{
						QualifiedName: "testorg::repo1::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
					},
					{
						QualifiedName: "testorg::repo1::scar-tissue",
						Domain:        "scar-tissue",
						Path:          ".know/scar-tissue.md",
					},
				},
			},
		},
	}

	// Minimal content — the vocabulary injection is what matters.
	loader := &mockContentLoader{
		content: map[string]string{
			"testorg::repo1::architecture": "# Architecture\n\nGeneral overview.",
			"testorg::repo1::conventions":  "# Conventions\n\nGeneral overview.",
			"testorg::repo1::scar-tissue":  "# Scar Tissue\n\nGeneral overview.",
		},
	}

	idx, err := BuildFromCatalog(catalog, loader)
	require.NoError(t, err)

	// Query for domain-type vocabulary should surface the correct domain.
	tests := []struct {
		query          string
		expectedDomain string
	}{
		{"layers packages components modules", "architecture"},
		{"practices standards guidelines idioms", "conventions"},
		{"bugs failures regressions incidents", "scar-tissue"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := idx.SearchDocuments(tt.query, 3)
			require.NotEmpty(t, results, "should find results for %q", tt.query)
			assert.Equal(t, tt.expectedDomain, results[0].Domain,
				"query %q should rank %s first", tt.query, tt.expectedDomain)
		})
	}
}

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with frontmatter", "---\ndomain: arch\n---\n# Title\nBody text", "# Title\nBody text"},
		{"no frontmatter", "# Title\nBody text", "# Title\nBody text"},
		{"empty", "", ""},
		{"only frontmatter", "---\nfoo: bar\n---", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, stripFrontmatter(tt.input))
		})
	}
}
