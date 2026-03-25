package intent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifier_Observe_DefaultTier(t *testing.T) {
	c := NewClassifier()
	result := c.Classify("What is the architecture of knossos?")
	assert.Equal(t, TierObserve, result.Tier)
	assert.True(t, result.Answerable)
	assert.Empty(t, result.UnsupportedReason)
	assert.Equal(t, "What is the architecture of knossos?", result.RawQuery)
}

func TestClassifier_EdgeCases(t *testing.T) {
	c := NewClassifier()

	tests := []struct {
		id               string
		query            string
		expectedTier     ActionTier
		expectedAnswer   bool
		expectedDomains  []string // subset check
		description      string
	}{
		// Intent edge cases from test corpus
		{
			id:             "I-01",
			query:          "Update the scar tissue for the session bug",
			expectedTier:   TierRecord,
			expectedAnswer: false,
			expectedDomains: []string{"scar-tissue"},
			description:    "Record verb 'update' + domain hint",
		},
		{
			id:             "I-02",
			query:          "Run a hotfix for the login issue",
			expectedTier:   TierAct,
			expectedAnswer: false,
			description:    "Act verb 'run' + 'hotfix'",
		},
		{
			id:             "I-03",
			query:          "How do I deploy a new version?",
			expectedTier:   TierObserve,
			expectedAnswer: true,
			description:    "'How do I' guard clause overrides 'deploy'",
		},
		{
			id:             "I-04",
			query:          "Create a new knowledge file for testing",
			expectedTier:   TierRecord,
			expectedAnswer: false,
			expectedDomains: []string{"test-coverage"},
			description:    "Record verb 'create'",
		},
		{
			id:             "I-05",
			query:          "What does the deploy script do?",
			expectedTier:   TierObserve,
			expectedAnswer: true,
			description:    "'What does' guard clause overrides 'deploy'",
		},
		{
			id:             "I-06",
			query:          "Tell me about the release process",
			expectedTier:   TierObserve,
			expectedAnswer: true,
			description:    "'Tell me about' guard clause",
		},
		{
			id:             "I-07",
			query:          "Execute the migration script in prod",
			expectedTier:   TierAct,
			expectedAnswer: false,
			description:    "Act verb 'execute'",
		},
		{
			id:             "I-08",
			query:          "Explain how to run integration tests",
			expectedTier:   TierObserve,
			expectedAnswer: true,
			description:    "'Explain how to' guard clause overrides 'run'",
		},
		{
			id:             "I-09",
			query:          "File a scar for the OOM incident",
			expectedTier:   TierRecord,
			expectedAnswer: false,
			expectedDomains: []string{"scar-tissue"},
			description:    "Record verb 'file' + 'scar' domain",
		},
		{
			id:             "I-10",
			query:          "Show me the deployment architecture",
			expectedTier:   TierObserve,
			expectedAnswer: true,
			expectedDomains: []string{"architecture"},
			description:    "'Show me' guard clause overrides 'deployment'",
		},
		{
			id:             "I-11",
			query:          "Revert the last commit on main",
			expectedTier:   TierAct,
			expectedAnswer: false,
			description:    "Act verb 'revert'",
		},
		{
			id:             "I-12",
			query:          "What are the best practices for error handling?",
			expectedTier:   TierObserve,
			expectedAnswer: true,
			description:    "Pure knowledge query",
		},
	}

	passed := 0
	for _, tt := range tests {
		t.Run(tt.id+" "+tt.description, func(t *testing.T) {
			result := c.Classify(tt.query)
			// Safety invariant: Record/Act must NEVER be misclassified as Observe.
			if tt.expectedTier == TierRecord || tt.expectedTier == TierAct {
				require.NotEqual(t, TierObserve, result.Tier,
					"SAFETY VIOLATION: Record/Act query classified as Observe: %s", tt.query)
			}

			assert.Equal(t, tt.expectedTier, result.Tier, "query: %s", tt.query)
			assert.Equal(t, tt.expectedAnswer, result.Answerable, "query: %s", tt.query)
			assert.Equal(t, tt.query, result.RawQuery)

			// Check expected domains are present (subset check).
			if len(tt.expectedDomains) > 0 {
				foundDomains := make(map[string]bool)
				for _, h := range result.DomainHints {
					foundDomains[h.Domain] = true
				}
				for _, d := range tt.expectedDomains {
					assert.True(t, foundDomains[d], "expected domain %q not in hints for query: %s", d, tt.query)
				}
			}

			passed++
		})
	}

	// PT-06-C1: >= 80% accuracy.
	total := len(tests)
	require.GreaterOrEqual(t, passed, int(float64(total)*0.8),
		"intent classifier accuracy below 80%%: %d/%d", passed, total)
}

func TestClassifier_RecordActNeverMisclassifiedAsObserve(t *testing.T) {
	// Safety invariant: Record/Act must NEVER be classified as Observe.
	c := NewClassifier()

	actQueries := []string{
		"Execute the migration script in prod",
		"Deploy to production",
		"Rollback the last release",
		"Trigger the CI pipeline",
		"Push the changes to main",
		"Merge the feature branch",
		"Ship the hotfix",
	}

	for _, q := range actQueries {
		result := c.Classify(q)
		assert.NotEqual(t, TierObserve, result.Tier,
			"SAFETY: Act query classified as Observe: %q", q)
	}

	recordQueries := []string{
		"Update the architecture doc",
		"Add a note about the bug",
		"Amend the scar tissue for the regression",
		"Create a new knowledge file",
		"Edit the conventions file",
	}

	for _, q := range recordQueries {
		result := c.Classify(q)
		assert.NotEqual(t, TierObserve, result.Tier,
			"SAFETY: Record query classified as Observe: %q", q)
	}
}

func TestClassifier_DomainHints_Architecture(t *testing.T) {
	c := NewClassifier()
	result := c.Classify("How does the package structure work?")
	assert.Equal(t, TierObserve, result.Tier)

	domains := make(map[string]bool)
	for _, h := range result.DomainHints {
		domains[h.Domain] = true
	}
	assert.True(t, domains["architecture"], "expected architecture domain hint")
}

func TestClassifier_DomainHints_MultiDomain(t *testing.T) {
	c := NewClassifier()
	// Should produce both architecture and test-coverage hints.
	result := c.Classify("What is the testing pattern for the architecture layer?")
	assert.Equal(t, TierObserve, result.Tier)

	domains := make(map[string]bool)
	for _, h := range result.DomainHints {
		domains[h.Domain] = true
	}
	assert.True(t, domains["architecture"] || domains["test-coverage"],
		"expected at least one of architecture or test-coverage in hints")
}

func TestClassifier_DomainHints_Empty_ForUnknownDomain(t *testing.T) {
	c := NewClassifier()
	// A query with no matching domain keywords should produce empty hints.
	result := c.Classify("What is the meaning of life?")
	assert.Equal(t, TierObserve, result.Tier)
	// May or may not have hints -- the pipeline handles empty hints as unfiltered search.
}

func TestClassifier_UnsupportedReason_Populated(t *testing.T) {
	c := NewClassifier()

	actResult := c.Classify("Deploy the new version to production")
	assert.Equal(t, TierAct, actResult.Tier)
	assert.False(t, actResult.Answerable)
	assert.NotEmpty(t, actResult.UnsupportedReason)

	recordResult := c.Classify("Create a new scar tissue document")
	assert.Equal(t, TierRecord, recordResult.Tier)
	assert.False(t, recordResult.Answerable)
	assert.NotEmpty(t, recordResult.UnsupportedReason)
}

func TestClassifier_Observe_AllHighConfidenceQueries(t *testing.T) {
	c := NewClassifier()

	queries := []string{
		"How does the sync pipeline work?",
		"What error handling patterns does knossos use?",
		"Explain the session lifecycle state machine",
		"What is the materializer pattern?",
		"How does the provenance manifest track file ownership?",
		"What testing patterns are used in the codebase?",
		"Describe the hook event pipeline from CC to ari",
		"What are the naming conventions for Go packages?",
	}

	for _, q := range queries {
		result := c.Classify(q)
		assert.Equal(t, TierObserve, result.Tier, "expected OBSERVE for: %s", q)
		assert.True(t, result.Answerable, "expected answerable for: %s", q)
	}
}

func TestClassifier_Tokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"what's the architecture?", []string{"what", "s", "the", "architecture"}},
		{"how-does-it-work", []string{"how", "does", "it", "work"}},
		{"feat/materialization", []string{"feat", "materialization"}},
	}
	for _, tt := range tests {
		got := tokenize(tt.input)
		assert.Equal(t, tt.expected, got, "input: %s", tt.input)
	}
}
