package ask

import (
	"encoding/json"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Interface Compliance Tests ---

func TestAskOutputImplementsTextable(t *testing.T) {
	// TC-A01: AskOutput implements output.Textable
	var _ output.Textable = AskOutput{}
}

// --- Text Output Format Tests ---

func TestAskOutputTextNoResults(t *testing.T) {
	// TC-A02: Empty results shows helpful "No matches" message
	out := AskOutput{
		Query:   "xyzzy",
		Results: nil,
		Total:   0,
	}

	text := out.Text()
	assert.Contains(t, text, `Query: "xyzzy"`)
	assert.Contains(t, text, "No matches found")
	assert.Contains(t, text, "Suggestions")
}

func TestAskOutputTextWithResults(t *testing.T) {
	// TC-A03: Numbered list format with rank, name, domain, summary, and action
	out := AskOutput{
		Query: "release",
		Results: []AskResultEntry{
			{
				Rank:    1,
				Name:    "releaser",
				Domain:  "rite",
				Summary: "Release engineering with dependency analysis",
				Action:  "/releaser",
				Score:   90,
			},
			{
				Rank:    2,
				Name:    "ari session wrap",
				Domain:  "command",
				Summary: "Archive the current session",
				Action:  "ari session wrap",
				Score:   45,
			},
		},
		Total: 2,
	}

	text := out.Text()
	assert.Contains(t, text, `Query: "release"`)
	assert.Contains(t, text, "1. releaser [rite]")
	assert.Contains(t, text, "Release engineering with dependency analysis")
	assert.Contains(t, text, "Try: /releaser")
	assert.Contains(t, text, "2. ari session wrap [command]")
	assert.Contains(t, text, "Try: ari session wrap")
	assert.Contains(t, text, "2 results found")
}

func TestAskOutputTextShowsContext(t *testing.T) {
	// TC-A04: Active rite context appears in text output
	out := AskOutput{
		Query:   "session",
		Results: []AskResultEntry{},
		Total:   0,
		Context: "Active rite: ecosystem",
	}

	text := out.Text()
	assert.Contains(t, text, "Active rite: ecosystem")
}

func TestAskOutputTextSingularResult(t *testing.T) {
	// TC-A05: Singular "result" (not "results") when total is 1
	out := AskOutput{
		Query: "sails",
		Results: []AskResultEntry{
			{Rank: 1, Name: "sails", Domain: "concept", Summary: "Ship readiness indicator", Action: "ari explain sails"},
		},
		Total: 1,
	}

	text := out.Text()
	assert.Contains(t, text, "1 result found")
	assert.NotContains(t, text, "1 results found")
}

func TestAskOutputTextNoActionSkipped(t *testing.T) {
	// TC-A06: Result with no action does not print "Try:" line
	out := AskOutput{
		Query: "concept",
		Results: []AskResultEntry{
			{Rank: 1, Name: "rite", Domain: "concept", Summary: "A workflow context", Action: ""},
		},
		Total: 1,
	}

	text := out.Text()
	assert.NotContains(t, text, "Try:")
}

func TestAskOutputTextNoContextWhenEmpty(t *testing.T) {
	// TC-A07: No context line when Context field is empty
	out := AskOutput{
		Query: "session",
		Results: []AskResultEntry{
			{Rank: 1, Name: "session", Domain: "concept", Summary: "A work unit", Action: "ari explain session"},
		},
		Total:   1,
		Context: "",
	}

	text := out.Text()
	assert.NotContains(t, text, "Active rite:")
}

// --- JSON Structure Tests ---

func TestAskOutputJSONFields(t *testing.T) {
	// TC-A08: JSON output contains all required fields
	out := AskOutput{
		Query: "release",
		Results: []AskResultEntry{
			{Rank: 1, Name: "releaser", Domain: "rite", Summary: "Release engineering", Action: "/releaser", Score: 80},
		},
		Total:   1,
		Context: "Active rite: ecosystem",
	}

	data, err := json.Marshal(out)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "release", decoded["query"])
	assert.EqualValues(t, 1, decoded["total"])
	assert.Equal(t, "Active rite: ecosystem", decoded["context"])

	results, ok := decoded["results"].([]any)
	require.True(t, ok, "results should be an array")
	require.Len(t, results, 1)

	entry, ok := results[0].(map[string]any)
	require.True(t, ok)
	assert.EqualValues(t, 1, entry["rank"])
	assert.Equal(t, "releaser", entry["name"])
	assert.Equal(t, "rite", entry["domain"])
	assert.Equal(t, "Release engineering", entry["summary"])
	assert.Equal(t, "/releaser", entry["action"])
	assert.EqualValues(t, 80, entry["score"])
}

func TestAskOutputJSONOmitsEmptyContext(t *testing.T) {
	// TC-A09: JSON omits context field when empty (omitempty)
	out := AskOutput{
		Query:   "session",
		Results: []AskResultEntry{},
		Total:   0,
		Context: "",
	}

	data, err := json.Marshal(out)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	_, hasContext := decoded["context"]
	assert.False(t, hasContext, "context field should be omitted when empty")
}

func TestAskResultEntryJSONOmitsZeroScore(t *testing.T) {
	// TC-A10: Score field omitted in JSON when zero (omitempty)
	entry := AskResultEntry{
		Rank:    1,
		Name:    "test",
		Domain:  "command",
		Summary: "Test summary",
		Action:  "ari test",
		Score:   0,
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	_, hasScore := decoded["score"]
	assert.False(t, hasScore, "score field should be omitted when zero")
}
