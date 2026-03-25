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
