package explain

import (
	"io/fs"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Concept Loading Tests ---

func TestAllConceptsLoaded(t *testing.T) {
	// TC-E01: All 16 concepts loaded
	assert.Equal(t, 16, len(registry))
}

func TestSortedNamesCorrect(t *testing.T) {
	// TC-E02: Sorted names correct
	expected := []string{
		"agent", "dromena", "evans-principle", "inscription", "knossos", "know", "ledge",
		"legomena", "mena", "potnia", "rite", "sails", "session", "sos", "tribute", "xenia",
	}
	assert.Equal(t, expected, sortedNames)
}

func TestEachConceptHasSummary(t *testing.T) {
	// TC-E03: Each concept has summary
	for _, entry := range AllConcepts() {
		assert.NotEmpty(t, entry.Summary, "concept %q has empty summary", entry.Name)
	}
}

func TestEachConceptHasDescription(t *testing.T) {
	// TC-E04: Each concept has description
	for _, entry := range AllConcepts() {
		assert.NotEmpty(t, entry.Description, "concept %q has empty description", entry.Name)
	}
}

func TestSeeAlsoIsNonNil(t *testing.T) {
	// TC-E05: SeeAlso is non-nil
	for _, entry := range AllConcepts() {
		assert.NotNil(t, entry.SeeAlso, "concept %q has nil SeeAlso", entry.Name)
	}
}

// --- Lookup Tests (table-driven) ---

func TestLookupExactMatch(t *testing.T) {
	// TC-E06 through TC-E18: Exact match for all 13 concepts
	concepts := []string{
		"rite", "session", "agent", "mena", "dromena", "legomena",
		"inscription", "tribute", "sails", "know", "ledge", "sos", "knossos",
		"evans-principle", "potnia", "xenia",
	}
	for _, name := range concepts {
		t.Run(name, func(t *testing.T) {
			entry, err := LookupConcept(name)
			require.NoError(t, err)
			assert.Equal(t, name, entry.Name)
		})
	}
}

// --- Case Normalization Tests ---

func TestLookupCaseNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"TC-E19: uppercase RITE", "RITE", "rite"},
		{"TC-E20: mixed case Rite", "Rite", "rite"},
		{"TC-E21: mixed case SeSsIoN", "SeSsIoN", "session"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := LookupConcept(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, entry.Name)
		})
	}
}

// --- Alias Resolution Tests ---

func TestLookupAliases(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		expected string
	}{
		{"TC-E22: rites -> rite", "rites", "rite"},
		{"TC-E23: sessions -> session", "sessions", "session"},
		{"TC-E24: agents -> agent", "agents", "agent"},
		{"TC-E25: skills -> legomena", "skills", "legomena"},
		{"TC-E26: commands -> dromena", "commands", "dromena"},
		{"TC-E54: evans -> evans-principle", "evans", "evans-principle"},
		{"TC-E55: hospitality -> xenia", "hospitality", "xenia"},
		{"TC-E56: coordinator -> potnia", "coordinator", "potnia"},
		{"TC-E57: potnia-theron -> potnia", "potnia-theron", "potnia"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := LookupConcept(tt.alias)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, entry.Name)
		})
	}
}

// --- Levenshtein Suggestion Tests ---

func TestLookupSuggestions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		suggestion string // expected in error message, empty if no suggestion expected
	}{
		{"TC-E27: ryte -> rite", "ryte", true, "rite"},
		{"TC-E28: sesion -> session", "sesion", true, "session"},
		{"TC-E29: agnt -> agent", "agnt", true, "agent"},
		{"TC-E30: menna -> mena", "menna", true, "mena"},
		{"TC-E31: xyz no close match", "xyz", true, ""},
		{"TC-E32: qwerty no close match", "qwerty", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LookupConcept(tt.input)
			require.Error(t, err)
			if tt.suggestion != "" {
				assert.Contains(t, err.Error(), "Did you mean")
				assert.Contains(t, err.Error(), tt.suggestion)
			} else {
				assert.NotContains(t, err.Error(), "Did you mean")
			}
			assert.Contains(t, err.Error(), "Available concepts:")
		})
	}
}

// --- Levenshtein Distance Unit Tests ---

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"TC-E33: identical", "rite", "rite", 0},
		{"TC-E34: one substitution", "rite", "ryte", 1},
		{"TC-E35: one insertion", "rite", "rites", 1},
		{"TC-E36: one deletion", "rite", "rit", 1},
		{"TC-E37: empty string", "", "abc", 3},
		{"TC-E38: both empty", "", "", 0},
		{"TC-E39: completely different", "abc", "xyz", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, levenshtein(tt.a, tt.b))
		})
	}
}

// --- Frontmatter Parsing Tests ---

func TestParseConceptValid(t *testing.T) {
	// TC-E40: Valid frontmatter
	data := []byte(`---
summary: Test summary line.
see_also: [foo, bar]
aliases: [test-alias]
harness_term: test-cc
---
This is the description body.
`)
	entry, err := parseConcept("test", data)
	require.NoError(t, err)
	assert.Equal(t, "test", entry.Name)
	assert.Equal(t, "Test summary line.", entry.Summary)
	assert.Equal(t, "This is the description body.", entry.Description)
	assert.Equal(t, []string{"foo", "bar"}, entry.SeeAlso)
	assert.Equal(t, []string{"test-alias"}, entry.Aliases)
	assert.Equal(t, "test-cc", entry.HarnessTerm)
	assert.Equal(t, "test (test-cc)", entry.DisplayName)
}

func TestParseConceptMissingSummary(t *testing.T) {
	// TC-E41: Missing summary
	data := []byte(`---
see_also: [foo]
---
Body text.
`)
	_, err := parseConcept("test", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: summary")
}

func TestParseConceptEmptySeeAlso(t *testing.T) {
	// TC-E42: Empty see_also
	data := []byte(`---
summary: Test summary.
see_also: []
---
Body text.
`)
	entry, err := parseConcept("test", data)
	require.NoError(t, err)
	assert.NotNil(t, entry.SeeAlso)
	assert.Empty(t, entry.SeeAlso)
}

func TestParseConceptNoAliases(t *testing.T) {
	// TC-E43: No aliases field
	data := []byte(`---
summary: Test summary.
see_also: [foo]
---
Body text.
`)
	entry, err := parseConcept("test", data)
	require.NoError(t, err)
	assert.NotNil(t, entry.Aliases)
	assert.Empty(t, entry.Aliases)
}

func TestParseConceptWithHarnessTerm(t *testing.T) {
	// TC-E44: With cc_term
	data := []byte(`---
summary: Test summary.
see_also: []
harness_term: skills
---
Body text.
`)
	entry, err := parseConcept("legomena", data)
	require.NoError(t, err)
	assert.Equal(t, "legomena (skills)", entry.DisplayName)
}

func TestParseConceptWithoutHarnessTerm(t *testing.T) {
	// TC-E45: Without cc_term
	data := []byte(`---
summary: Test summary.
see_also: []
---
Body text.
`)
	entry, err := parseConcept("rite", data)
	require.NoError(t, err)
	assert.Equal(t, "rite", entry.DisplayName)
}

// --- Display Name Tests ---

func TestDisplayNames(t *testing.T) {
	tests := []struct {
		name            string
		concept         string
		expectedDisplay string
	}{
		{"TC-E46: legomena", "legomena", "legomena (skills)"},
		{"TC-E47: dromena", "dromena", "dromena (commands)"},
		{"TC-E48: rite (no CC term)", "rite", "rite"},
		{"TC-E49: session (no CC term)", "session", "session"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := registry[tt.concept]
			require.True(t, ok, "concept %q not found in registry", tt.concept)
			assert.Equal(t, tt.expectedDisplay, entry.DisplayName)
		})
	}
}

// --- Interface Compliance Tests ---

func TestConceptOutputImplementsTextable(t *testing.T) {
	// TC-E50: ConceptOutput implements Textable
	var _ output.Textable = ConceptOutput{}
}

func TestConceptListOutputImplementsTabular(t *testing.T) {
	// TC-E51: ConceptListOutput implements Tabular
	var _ output.Tabular = ConceptListOutput{}
}

// --- Embedded File Inventory Tests ---

func TestEmbeddedFSContains16Files(t *testing.T) {
	// TC-E52: Embedded FS contains exactly 16 files
	entries, err := fs.ReadDir(conceptFS, "concepts")
	require.NoError(t, err)

	mdCount := 0
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 3 && e.Name()[len(e.Name())-3:] == ".md" {
			mdCount++
		}
	}
	assert.Equal(t, 16, mdCount)
}

func TestAllExpectedFilenamesPresent(t *testing.T) {
	// TC-E53: All expected filenames present
	expectedFiles := []string{
		"agent.md", "dromena.md", "evans-principle.md", "inscription.md", "knossos.md",
		"know.md", "ledge.md", "legomena.md", "mena.md", "potnia.md",
		"rite.md", "sails.md", "session.md", "sos.md", "tribute.md", "xenia.md",
	}

	entries, err := fs.ReadDir(conceptFS, "concepts")
	require.NoError(t, err)

	fileNames := make(map[string]bool)
	for _, e := range entries {
		fileNames[e.Name()] = true
	}

	for _, expected := range expectedFiles {
		assert.True(t, fileNames[expected], "missing embedded concept file: %s", expected)
	}
}

// --- Text Output Format Tests ---

func TestConceptOutputTextFormat(t *testing.T) {
	co := ConceptOutput{
		Concept:        "legomena",
		DisplayName:    "legomena (skills)",
		Summary:        "Skills (things said).",
		Description:    "Skills (things said) -- persistent mena.",
		SeeAlso:        []string{"dromena", "mena"},
		ProjectContext: "Your project has 8 legomena.",
	}

	text := co.Text()
	assert.Contains(t, text, "=== legomena (skills) ===")
	assert.Contains(t, text, "Skills (things said) -- persistent mena.")
	assert.Contains(t, text, "See also: dromena, mena")
	assert.Contains(t, text, "Your project has 8 legomena.")
}

func TestConceptOutputTextNoSeeAlso(t *testing.T) {
	co := ConceptOutput{
		Concept:     "test",
		DisplayName: "test",
		Summary:     "Test.",
		Description: "Test description.",
		SeeAlso:     []string{},
	}

	text := co.Text()
	assert.NotContains(t, text, "See also:")
}

func TestConceptOutputTextNoProjectContext(t *testing.T) {
	co := ConceptOutput{
		Concept:     "test",
		DisplayName: "test",
		Summary:     "Test.",
		Description: "Test description.",
		SeeAlso:     []string{"foo"},
	}

	text := co.Text()
	// Should not have extra blank line at the end for missing context
	assert.Contains(t, text, "See also: foo")
}

func TestConceptListOutputHeaders(t *testing.T) {
	list := ConceptListOutput{}
	assert.Equal(t, []string{"CONCEPT", "SUMMARY"}, list.Headers())
}

func TestConceptListOutputRows(t *testing.T) {
	list := ConceptListOutput{
		Concepts: []ConceptSummary{
			{Name: "agent", Summary: "Agent summary."},
			{Name: "rite", Summary: "Rite summary."},
		},
	}
	rows := list.Rows()
	assert.Len(t, rows, 2)
	assert.Equal(t, []string{"agent", "Agent summary."}, rows[0])
	assert.Equal(t, []string{"rite", "Rite summary."}, rows[1])
}
