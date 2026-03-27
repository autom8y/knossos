package citation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractCitations(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single citation",
			text:     "The architecture is described in [autom8y::knossos::architecture].",
			expected: []string{"autom8y::knossos::architecture"},
		},
		{
			name:     "multiple citations",
			text:     "See [autom8y::knossos::architecture] and [autom8y::data::conventions].",
			expected: []string{"autom8y::knossos::architecture", "autom8y::data::conventions"},
		},
		{
			name:     "duplicate citations deduplicated",
			text:     "[autom8y::knossos::architecture] mentions [autom8y::knossos::architecture] again.",
			expected: []string{"autom8y::knossos::architecture"},
		},
		{
			name:     "no citations",
			text:     "This is a plain response with no citations.",
			expected: nil,
		},
		{
			name:     "citation with hyphens and underscores",
			text:     "Found in [autom8y::my-repo::scar-tissue] and [org_2::repo_1::design-constraints].",
			expected: []string{"autom8y::my-repo::scar-tissue", "org_2::repo_1::design-constraints"},
		},
		{
			name:     "partial citation not matched",
			text:     "Reference: [autom8y::knossos] is incomplete.",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractCitations(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}
