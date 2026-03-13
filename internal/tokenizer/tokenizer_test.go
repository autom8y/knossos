package tokenizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCount(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	// Empty string
	assert.Equal(t, 0, c.Count(""))

	// Known token counts (cl100k_base approximation)
	n := c.Count("hello world")
	assert.Greater(t, n, 0)
	assert.Less(t, n, 10) // "hello world" should be ~2 tokens
}

func TestCountFile(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	require.NoError(t, os.WriteFile(path, []byte("# Hello\n\nThis is a test file with some content."), 0644))

	n, err := c.CountFile(path)
	require.NoError(t, err)
	assert.Greater(t, n, 0)

	// Non-existent file
	_, err = c.CountFile(filepath.Join(dir, "nope.md"))
	assert.Error(t, err)
}

func TestCalculateBudget(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	// Create a mock channel directory
	channelDir := t.TempDir()

	// CLAUDE.md with sections
	claudeMd := `<!-- KNOSSOS:START quick-start -->
## Quick Start
This project uses a multi-agent workflow.
<!-- KNOSSOS:END quick-start -->

<!-- KNOSSOS:START agents -->
## Agents
- orchestrator
- analyst
<!-- KNOSSOS:END agents -->
`
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "CLAUDE.md"), []byte(claudeMd), 0644))

	// agents/
	agentsDir := filepath.Join(channelDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "orchestrator.md"), []byte("You are the orchestrator agent."), 0644))

	// skills/
	skillsDir := filepath.Join(channelDir, "skills", "nav")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "worktree.md"), []byte("Worktree management skill content."), 0644))

	// rules/
	rulesDir := filepath.Join(channelDir, "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "go.md"), []byte("Use gofmt."), 0644))

	report, err := c.CalculateBudget(channelDir)
	require.NoError(t, err)

	assert.Greater(t, report.TotalTokens, 0)
	assert.Contains(t, report.Categories, "CLAUDE.md")
	assert.Contains(t, report.Categories, "agents")
	assert.Contains(t, report.Categories, "skills")
	assert.Contains(t, report.Categories, "rules")
	assert.Greater(t, len(report.Files), 0)

	// Sections parsed
	assert.Greater(t, len(report.Sections), 0)
	sectionNames := make([]string, len(report.Sections))
	for i, s := range report.Sections {
		sectionNames[i] = s.Name
	}
	assert.Contains(t, sectionNames, "quick-start")
	assert.Contains(t, sectionNames, "agents")

	// Files sorted by token count descending
	for i := 1; i < len(report.Files); i++ {
		assert.GreaterOrEqual(t, report.Files[i-1].Tokens, report.Files[i].Tokens)
	}
}

func TestCalculateBudget_Empty(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	channelDir := t.TempDir()
	report, err := c.CalculateBudget(channelDir)
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalTokens)
}
