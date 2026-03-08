package search

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- StaticSynonymSource ---

func TestStaticSynonymSource_KnownMappings(t *testing.T) {
	src := NewStaticSynonymSource()

	tests := []struct {
		token    string
		expected []string
	}{
		{"deploy", []string{"sre", "operations", "reliability", "infrastructure"}},
		{"refactor", []string{"hygiene", "cleanup", "code-quality"}},
		{"ship", []string{"releaser", "release", "publish"}},
		{"audit", []string{"review", "security", "compliance"}},
		{"test", []string{"qa", "validation", "testing"}},
		{"docs", []string{"documentation", "technical-writing"}},
		{"document", []string{"documentation", "docs", "technical-writing"}},
		{"build", []string{"forge", "compile", "implementation"}},
		{"debug", []string{"clinic", "diagnose", "troubleshoot"}},
		{"plan", []string{"strategy", "planning", "roadmap"}},
		{"research", []string{"rnd", "spike", "investigation"}},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			result := src.Expand(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStaticSynonymSource_UnknownToken(t *testing.T) {
	src := NewStaticSynonymSource()
	assert.Nil(t, src.Expand("xyzzy"))
}

func TestStaticSynonymSource_CaseInsensitive(t *testing.T) {
	// Input is expected lowercase (tokenize handles uppercasing).
	// StaticSynonymSource lowercases input for safety.
	src := NewStaticSynonymSource()
	result := src.Expand("Deploy")
	assert.Equal(t, []string{"sre", "operations", "reliability", "infrastructure"}, result)
}

func TestStaticSynonymSource_NilSource(t *testing.T) {
	var src *StaticSynonymSource
	assert.Nil(t, src.Expand("deploy"))
}

// --- OrchestratorSynonymSource ---

func TestOrchestratorSynonymSource_FromTriggers(t *testing.T) {
	root := t.TempDir()
	riteDir := filepath.Join(root, "my-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))

	orchContent := `rite:
  name: my-rite
  domain: testing
frontmatter:
  description: "Coordinates phases. Triggers: coordinate, orchestrate, release workflow"
routing:
  analyst: "Gap analysis"
`
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "orchestrator.yaml"), []byte(orchContent), 0644))

	src := NewOrchestratorSynonymSource(root)

	// "coordinate" should map to the rite name.
	expansions := src.Expand("coordinate")
	assert.Contains(t, expansions, "my-rite")

	// "orchestrate" should map to the rite name.
	expansions = src.Expand("orchestrate")
	assert.Contains(t, expansions, "my-rite")
}

func TestOrchestratorSynonymSource_IncludesDomain(t *testing.T) {
	root := t.TempDir()
	riteDir := filepath.Join(root, "hygiene")
	require.NoError(t, os.MkdirAll(riteDir, 0755))

	orchContent := `rite:
  name: hygiene
  domain: code quality
frontmatter:
  description: "Triggers: cleanup, lint"
`
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "orchestrator.yaml"), []byte(orchContent), 0644))

	src := NewOrchestratorSynonymSource(root)

	// "cleanup" should expand to rite name AND domain.
	expansions := src.Expand("cleanup")
	assert.Contains(t, expansions, "hygiene")
	assert.Contains(t, expansions, "code quality")
}

func TestOrchestratorSynonymSource_MissingDir(t *testing.T) {
	src := NewOrchestratorSynonymSource("/nonexistent/path")
	require.NotNil(t, src, "should return empty source, not nil")
	assert.Nil(t, src.Expand("anything"))
}

func TestOrchestratorSynonymSource_MalformedYAML(t *testing.T) {
	root := t.TempDir()

	// Create a valid rite and a malformed one.
	goodDir := filepath.Join(root, "good-rite")
	require.NoError(t, os.MkdirAll(goodDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(goodDir, "orchestrator.yaml"), []byte(`rite:
  name: good-rite
  domain: testing
frontmatter:
  description: "Triggers: validate"
`), 0644))

	badDir := filepath.Join(root, "bad-rite")
	require.NoError(t, os.MkdirAll(badDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(badDir, "orchestrator.yaml"), []byte(`{{{not yaml`), 0644))

	src := NewOrchestratorSynonymSource(root)

	// Good rite's triggers should be available.
	expansions := src.Expand("validate")
	assert.Contains(t, expansions, "good-rite")

	// Bad rite should be silently skipped — no panic.
	assert.Nil(t, src.Expand("not"))
}

func TestOrchestratorSynonymSource_NilSource(t *testing.T) {
	var src *OrchestratorSynonymSource
	assert.Nil(t, src.Expand("anything"))
}

// --- CompositeSynonymSource ---

func TestCompositeSynonymSource_MergesDeduplicated(t *testing.T) {
	static := NewStaticSynonymSource()

	// Create an orchestrator source that also maps "ship" to "releaser".
	orch := &OrchestratorSynonymSource{
		synonyms: map[string][]string{
			"ship": {"releaser", "deploy-tool"},
		},
	}

	composite := NewCompositeSynonymSource(static, orch)
	expansions := composite.Expand("ship")

	// Should have static expansions + orchestrator's "deploy-tool" but deduplicated "releaser".
	assert.Contains(t, expansions, "releaser")
	assert.Contains(t, expansions, "release")
	assert.Contains(t, expansions, "publish")
	assert.Contains(t, expansions, "deploy-tool")

	// Count "releaser" — should appear exactly once.
	count := 0
	for _, e := range expansions {
		if e == "releaser" {
			count++
		}
	}
	assert.Equal(t, 1, count, "releaser should appear exactly once after deduplication")
}

func TestCompositeSynonymSource_Empty(t *testing.T) {
	composite := NewCompositeSynonymSource()
	assert.Nil(t, composite.Expand("anything"))
}

func TestCompositeSynonymSource_NilSources(t *testing.T) {
	composite := NewCompositeSynonymSource(nil, nil)
	assert.Nil(t, composite.Expand("deploy"))
}

func TestCompositeSynonymSource_NilComposite(t *testing.T) {
	var composite *CompositeSynonymSource
	assert.Nil(t, composite.Expand("deploy"))
}

// --- expandSynonyms ---

func TestMaxExpansionsPerToken(t *testing.T) {
	// Create a source with more than 6 expansions for a single token.
	src := &StaticSynonymSource{
		synonyms: map[string][]string{
			"deploy": {"a", "b", "c", "d", "e", "f", "g", "h"},
		},
	}
	expansions := expandSynonyms("deploy", src)
	assert.Len(t, expansions, maxExpansionsPerToken, "should cap at %d expansions", maxExpansionsPerToken)
}

func TestMinTokenLengthForExpansion(t *testing.T) {
	src := NewStaticSynonymSource()

	// 2-char token should not be expanded.
	assert.Nil(t, expandSynonyms("qa", src))

	// 1-char token should not be expanded.
	assert.Nil(t, expandSynonyms("a", src))

	// 3-char token should be expanded if it has synonyms.
	// "docs" is 4 chars, so use a custom source for 3-char test.
	customSrc := &StaticSynonymSource{
		synonyms: map[string][]string{
			"fix": {"repair", "patch"},
		},
	}
	result := expandSynonyms("fix", customSrc)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
}

func TestExpandSynonyms_NilSource(t *testing.T) {
	assert.Nil(t, expandSynonyms("deploy", nil))
}

func TestExpandSynonyms_NoSynonymsForToken(t *testing.T) {
	src := NewStaticSynonymSource()
	assert.Nil(t, expandSynonyms("xyzzy", src))
}
