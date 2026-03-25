package bm25

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSections(t *testing.T) {
	content := `Intro text.

## Package Structure

The internal packages.

## Layer Boundaries

Cmd imports domain.
`
	sections := SplitSections(content)
	assert.Len(t, sections, 3) // intro + 2 headed sections

	// Intro section (no heading).
	assert.Equal(t, "", sections[0].Heading)
	assert.Contains(t, sections[0].Body, "Intro text")

	// Package Structure section.
	assert.Equal(t, "Package Structure", sections[1].Heading)
	assert.Contains(t, sections[1].Body, "internal packages")
	assert.Equal(t, "package-structure", sections[1].Slug)

	// Layer Boundaries section.
	assert.Equal(t, "Layer Boundaries", sections[2].Heading)
	assert.Contains(t, sections[2].Body, "Cmd imports domain")
}

func TestSplitSections_Empty(t *testing.T) {
	assert.Nil(t, SplitSections(""))
	assert.Nil(t, SplitSections("   "))
}

func TestSplitSections_NoHeadings(t *testing.T) {
	sections := SplitSections("Just some plain text.\nNo headings here.")
	assert.Len(t, sections, 1)
	assert.Equal(t, "", sections[0].Heading)
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Package Structure", "package-structure"},
		{"Layer Boundaries", "layer-boundaries"},
		{"Error Handling Style", "error-handling-style"},
		{"Domain-Specific Idioms", "domain-specific-idioms"},
		{"", ""},
		{"A B C D E", "a-b-c-d-e"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, Slugify(tt.input))
		})
	}
}

func TestSectionQualifiedName(t *testing.T) {
	assert.Equal(t, "org::repo::arch##pkg-struct",
		SectionQualifiedName("org::repo::arch", "pkg-struct"))
	assert.Equal(t, "org::repo::arch",
		SectionQualifiedName("org::repo::arch", ""))
}
