package explain

import (
	"testing"

	"github.com/autom8y/knossos/internal/concept"
	"github.com/autom8y/knossos/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Concept Loading Tests ---
// These tests verify the concept package through the explain delegation layer.

func TestAllConceptsLoaded(t *testing.T) {
	// TC-E01: All 16 concepts loaded
	assert.Equal(t, 16, len(AllConcepts()))
}

func TestSortedNamesCorrect(t *testing.T) {
	// TC-E02: Sorted names correct
	expected := []string{
		"agent", "dromena", "evans-principle", "inscription", "knossos", "know", "ledge",
		"legomena", "mena", "potnia", "rite", "sails", "session", "sos", "tribute", "xenia",
	}
	all := AllConcepts()
	names := make([]string, len(all))
	for i, c := range all {
		names[i] = c.Name
	}
	assert.Equal(t, expected, names)
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
	// TC-E06 through TC-E18: Exact match for all 16 concepts
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

// --- Frontmatter Parsing Tests ---
// These tests now use the concept package's exported ParseConceptForTest helper
// since parseConcept is internal to the concept package.

func TestParseConceptValid(t *testing.T) {
	// TC-E40: Valid concept entry via LookupConcept
	entry, err := concept.LookupConcept("rite")
	require.NoError(t, err)
	assert.Equal(t, "rite", entry.Name)
	assert.NotEmpty(t, entry.Summary)
	assert.NotEmpty(t, entry.Description)
	assert.NotNil(t, entry.SeeAlso)
}

// --- Display Name Tests ---

func TestDisplayNames(t *testing.T) {
	tests := []struct {
		name            string
		conceptName     string
		expectedDisplay string
	}{
		{"TC-E46: legomena", "legomena", "legomena (skills)"},
		{"TC-E47: dromena", "dromena", "dromena (commands)"},
		{"TC-E48: rite (no CC term)", "rite", "rite"},
		{"TC-E49: session (no CC term)", "session", "session"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := concept.LookupConcept(tt.conceptName)
			require.NoError(t, err)
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

func TestAllConceptsHave16Entries(t *testing.T) {
	// TC-E52: 16 concepts loaded
	assert.Equal(t, 16, len(AllConcepts()))
}

func TestAllExpectedConceptsPresent(t *testing.T) {
	// TC-E53: All expected concepts present
	expectedNames := []string{
		"agent", "dromena", "evans-principle", "inscription", "knossos",
		"know", "ledge", "legomena", "mena", "potnia",
		"rite", "sails", "session", "sos", "tribute", "xenia",
	}

	conceptMap := make(map[string]bool)
	for _, c := range AllConcepts() {
		conceptMap[c.Name] = true
	}

	for _, expected := range expectedNames {
		assert.True(t, conceptMap[expected], "missing concept: %s", expected)
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
