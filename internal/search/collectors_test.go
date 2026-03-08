package search

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/paths"
)

// --- helpers ---

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

// buildTestRoot creates a minimal project structure and returns the root path.
func buildTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".claude", "agents"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".claude", "commands"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".knossos", "rites"), 0755))
	return root
}

// --- CollectCommands ---

func TestCollectCommands(t *testing.T) {
	root := &cobra.Command{Use: "ari", Short: "Ariadne CLI"}
	sub1 := &cobra.Command{Use: "sync", Short: "Sync resources"}
	sub2 := &cobra.Command{Use: "create", Short: "Create a session"}
	hidden := &cobra.Command{Use: "internal", Short: "Hidden", Hidden: true}

	session := &cobra.Command{Use: "session", Short: "Manage sessions"}
	session.AddCommand(sub2)
	root.AddCommand(sub1, session, hidden)

	entries := CollectCommands(root)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
		assert.Equal(t, DomainCommand, e.Domain)
	}

	assert.True(t, names["sync"], "sync should be collected")
	assert.True(t, names["session"], "session should be collected")
	assert.True(t, names["session create"], "session create should be collected")
	assert.False(t, names["internal"], "hidden commands should not be collected")
}

func TestCollectCommandsAction(t *testing.T) {
	root := &cobra.Command{Use: "ari"}
	sub := &cobra.Command{Use: "explain", Short: "Explain a concept"}
	root.AddCommand(sub)

	entries := CollectCommands(root)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Action, "--help")
}

func TestCollectCommandsEmpty(t *testing.T) {
	root := &cobra.Command{Use: "ari"}
	entries := CollectCommands(root)
	assert.Empty(t, entries)
}

// --- CollectConcepts ---

func TestCollectConcepts(t *testing.T) {
	entries := CollectConcepts()
	// The explain package embeds exactly 13 concepts.
	assert.Len(t, entries, 13)
}

func TestCollectConceptsDomain(t *testing.T) {
	entries := CollectConcepts()
	for _, e := range entries {
		assert.Equal(t, DomainConcept, e.Domain, "concept %q should have DomainConcept", e.Name)
	}
}

func TestCollectConceptsAction(t *testing.T) {
	entries := CollectConcepts()
	for _, e := range entries {
		assert.Contains(t, e.Action, "ari explain", "concept %q action should reference ari explain", e.Name)
	}
}

// --- CollectRites ---

func TestCollectRitesNoProject(t *testing.T) {
	// Nil resolver → empty slice.
	assert.Empty(t, CollectRites(nil))
}

func TestCollectRitesEmptyProjectRoot(t *testing.T) {
	resolver := paths.NewResolver("")
	assert.Empty(t, CollectRites(resolver))
}

func TestCollectRitesNoRitesDir(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".claude"), 0755))
	resolver := paths.NewResolver(root)
	// .knossos/rites doesn't exist — should return empty.
	assert.Empty(t, CollectRites(resolver))
}

func TestCollectRitesWithRite(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "test-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeTestFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: test-rite\ndescription: A test rite")

	resolver := paths.NewResolver(root)
	entries := CollectRites(resolver)

	require.Len(t, entries, 1)
	assert.Equal(t, "test-rite", entries[0].Name)
	assert.Equal(t, DomainRite, entries[0].Domain)
	assert.Equal(t, "/test-rite", entries[0].Action)
}

func TestCollectRitesActiveRiteBoosted(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "active-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeTestFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: active-rite")
	// Mark it active.
	writeTestFile(t, filepath.Join(root, ".knossos", "ACTIVE_RITE"), "active-rite")

	resolver := paths.NewResolver(root)
	entries := CollectRites(resolver)

	require.Len(t, entries, 1)
	assert.True(t, entries[0].Boosted, "active rite should be boosted")
}

// --- CollectAgents ---

func TestCollectAgentsNilResolver(t *testing.T) {
	assert.Empty(t, CollectAgents(nil))
}

func TestCollectAgentsEmptyProjectRoot(t *testing.T) {
	resolver := paths.NewResolver("")
	assert.Empty(t, CollectAgents(resolver))
}

func TestCollectAgentsMissingDir(t *testing.T) {
	root := t.TempDir()
	resolver := paths.NewResolver(root)
	assert.Empty(t, CollectAgents(resolver))
}

func TestCollectAgentsWithAgent(t *testing.T) {
	root := buildTestRoot(t)
	agentContent := `---
name: my-agent
description: Does important work
tools: [Bash, Read]
---
This agent does important work.
`
	writeTestFile(t, filepath.Join(root, ".claude", "agents", "my-agent.md"), agentContent)

	resolver := paths.NewResolver(root)
	entries := CollectAgents(resolver)

	require.Len(t, entries, 1)
	assert.Equal(t, "my-agent", entries[0].Name)
	assert.Equal(t, DomainAgent, entries[0].Domain)
	assert.NotEmpty(t, entries[0].Summary)
}

func TestCollectAgentsSkipsInvalidFrontmatter(t *testing.T) {
	root := buildTestRoot(t)
	// File without frontmatter should be skipped.
	writeTestFile(t, filepath.Join(root, ".claude", "agents", "bad.md"), "no frontmatter here")

	resolver := paths.NewResolver(root)
	entries := CollectAgents(resolver)
	assert.Empty(t, entries)
}

// --- CollectDromena ---

func TestCollectDromenaNilResolver(t *testing.T) {
	assert.Empty(t, CollectDromena(nil))
}

func TestCollectDromenaEmptyProjectRoot(t *testing.T) {
	resolver := paths.NewResolver("")
	assert.Empty(t, CollectDromena(resolver))
}

func TestCollectDromenaMissingDir(t *testing.T) {
	root := t.TempDir()
	resolver := paths.NewResolver(root)
	assert.Empty(t, CollectDromena(resolver))
}

func TestCollectDromenaWithCommand(t *testing.T) {
	root := buildTestRoot(t)
	// Note: description must be quoted in YAML if it contains colons.
	cmdContent := `---
name: my-command
description: "Does something useful. Triggers: deploy, release"
---
Body text.
`
	writeTestFile(t, filepath.Join(root, ".claude", "commands", "my-command.dro.md"), cmdContent)

	resolver := paths.NewResolver(root)
	entries := CollectDromena(resolver)

	require.Len(t, entries, 1)
	assert.Equal(t, "my-command", entries[0].Name)
	assert.Equal(t, DomainDromena, entries[0].Domain)
	assert.Equal(t, "/my-command", entries[0].Action)
	assert.Contains(t, entries[0].Keywords, "deploy")
}

func TestCollectDromenaSkipsNoFrontmatter(t *testing.T) {
	root := buildTestRoot(t)
	writeTestFile(t, filepath.Join(root, ".claude", "commands", "plain.md"), "No frontmatter")

	resolver := paths.NewResolver(root)
	entries := CollectDromena(resolver)
	assert.Empty(t, entries)
}

// --- CollectRouting ---

func TestCollectRoutingNilResolver(t *testing.T) {
	assert.Empty(t, CollectRouting(nil))
}

func TestCollectRoutingEmptyProjectRoot(t *testing.T) {
	resolver := paths.NewResolver("")
	assert.Empty(t, CollectRouting(resolver))
}

func TestCollectRoutingMissingDir(t *testing.T) {
	root := t.TempDir()
	resolver := paths.NewResolver(root)
	assert.Empty(t, CollectRouting(resolver))
}

func TestCollectRoutingWithOrchestrator(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "my-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))

	orchContent := `rite:
  name: my-rite
  domain: testing
frontmatter:
  description: "Coordinates phases. Triggers: coordinate, orchestrate"
routing:
  analyst: "Gap analysis needed"
  engineer: "Implementation needed"
`
	writeTestFile(t, filepath.Join(riteDir, "orchestrator.yaml"), orchContent)

	resolver := paths.NewResolver(root)
	entries := CollectRouting(resolver)

	assert.Len(t, entries, 2)
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
		assert.Equal(t, DomainRouting, e.Domain)
	}
	assert.True(t, names["analyst"])
	assert.True(t, names["engineer"])
}
