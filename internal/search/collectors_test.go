package search

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/config"
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
	// The explain package embeds exactly 16 concepts.
	assert.Len(t, entries, 16)
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

// --- CollectRites enrichment ---

func TestCollectRitesWithOrchestratorKeywords(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "sre-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeTestFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: sre-rite\ndescription: SRE management")

	orchContent := `rite:
  name: sre-rite
  domain: site reliability engineering
frontmatter:
  description: "Coordinates reliability. Triggers: coordinate, orchestrate, reliability"
routing:
  engineer: "Implementation needed"
`
	writeTestFile(t, filepath.Join(riteDir, "orchestrator.yaml"), orchContent)

	resolver := paths.NewResolver(root)
	entries := CollectRites(resolver)

	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Keywords, "coordinate")
	assert.Contains(t, entries[0].Keywords, "orchestrate")
	assert.Contains(t, entries[0].Keywords, "reliability")
}

func TestCollectRitesWithDomainAlias(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "hygiene")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeTestFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: hygiene\ndescription: Code quality")

	orchContent := `rite:
  name: hygiene
  domain: code quality
frontmatter:
  description: "Triggers: lint, cleanup"
`
	writeTestFile(t, filepath.Join(riteDir, "orchestrator.yaml"), orchContent)

	resolver := paths.NewResolver(root)
	entries := CollectRites(resolver)

	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Aliases, "code quality", "domain should be added as alias")
}

func TestCollectRitesSeparatesSummaryDescription(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "test-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeTestFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: test-rite\ndescription: A test rite")

	// Use YAML literal block scalar (|) to preserve newlines in description.
	orchContent := `rite:
  name: test-rite
  domain: testing
frontmatter:
  description: |
    Coordinates testing phases.
    Triggers: validate, verify
    Use when: testing needed
`
	writeTestFile(t, filepath.Join(riteDir, "orchestrator.yaml"), orchContent)

	resolver := paths.NewResolver(root)
	entries := CollectRites(resolver)

	require.Len(t, entries, 1)
	// Summary should be first line only.
	assert.Equal(t, "Coordinates testing phases.", entries[0].Summary)
	// Description should be the full text.
	assert.Contains(t, entries[0].Description, "Triggers: validate, verify")
	assert.Contains(t, entries[0].Description, "Use when: testing needed")
}

func TestCollectRitesNoOrchestratorGraceful(t *testing.T) {
	root := buildTestRoot(t)
	riteDir := filepath.Join(root, ".knossos", "rites", "plain-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeTestFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: plain-rite\ndescription: A plain rite")
	// No orchestrator.yaml — enrichment should be skipped gracefully.

	resolver := paths.NewResolver(root)
	entries := CollectRites(resolver)

	require.Len(t, entries, 1)
	assert.Equal(t, "plain-rite", entries[0].Name)
	assert.Equal(t, "A plain rite", entries[0].Summary)
	assert.Equal(t, "A plain rite", entries[0].Description)
	assert.Empty(t, entries[0].Keywords, "no orchestrator means no keywords")
	assert.Empty(t, entries[0].Aliases, "no orchestrator means no aliases")
}

// --- CollectProcessions ---

// isolateProcessionHome overrides KNOSSOS_HOME to prevent picking up real templates.
func isolateProcessionHome(t *testing.T) {
	t.Helper()
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", t.TempDir())
	t.Cleanup(config.ResetKnossosHome)
}

func writeProcessionTemplate(t *testing.T, root, name, description string, stations int) {
	t.Helper()
	dir := filepath.Join(root, "processions")
	require.NoError(t, os.MkdirAll(dir, 0755))

	content := "name: " + name + "\ndescription: \"" + description + "\"\nstations:\n"
	rites := []string{"security", "debt-triage", "10x-dev", "hygiene", "security"}
	for i := 0; i < stations; i++ {
		r := rites[i%len(rites)]
		content += "  - name: station-" + string(rune('a'+i)) + "\n"
		content += "    rite: " + r + "\n"
		content += "    goal: \"Goal " + string(rune('a'+i)) + "\"\n"
		content += "    produces: [artifact-" + string(rune('a'+i)) + "]\n"
	}
	content += "artifact_dir: .sos/wip/" + name + "/\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0644))
}

func TestCollectProcessions_NilResolver(t *testing.T) {
	entries := CollectProcessions(nil)
	assert.Nil(t, entries)
}

func TestCollectProcessions_NoTemplates(t *testing.T) {
	isolateProcessionHome(t)
	root := buildTestRoot(t)
	resolver := paths.NewResolver(root)
	entries := CollectProcessions(resolver)
	assert.Empty(t, entries)
}

func TestCollectProcessions_SingleTemplate(t *testing.T) {
	isolateProcessionHome(t)
	root := buildTestRoot(t)
	writeProcessionTemplate(t, root, "sec-rem", "Security remediation lifecycle", 5)

	resolver := paths.NewResolver(root)
	entries := CollectProcessions(resolver)

	require.Len(t, entries, 1)
	assert.Equal(t, "sec-rem", entries[0].Name)
	assert.Equal(t, DomainProcession, entries[0].Domain)
	assert.Equal(t, "Security remediation lifecycle", entries[0].Summary)
	assert.Equal(t, "/sec-rem", entries[0].Action)
	assert.True(t, entries[0].Boosted, "5-station template should be boosted (>2)")

	// Keywords should include station names, rite names, and fixed terms
	assert.Contains(t, entries[0].Keywords, "station-a")
	assert.Contains(t, entries[0].Keywords, "security")
	assert.Contains(t, entries[0].Keywords, "procession")
	assert.Contains(t, entries[0].Keywords, "cross-rite")
	assert.Contains(t, entries[0].Keywords, "workflow")
}

func TestCollectProcessions_NotBoostedWhenFewStations(t *testing.T) {
	isolateProcessionHome(t)
	root := buildTestRoot(t)
	writeProcessionTemplate(t, root, "small", "Two station workflow", 2)

	resolver := paths.NewResolver(root)
	entries := CollectProcessions(resolver)

	require.Len(t, entries, 1)
	assert.False(t, entries[0].Boosted, "2-station template should not be boosted")
}

func TestCollectProcessions_MultipleTemplates(t *testing.T) {
	isolateProcessionHome(t)
	root := buildTestRoot(t)
	writeProcessionTemplate(t, root, "workflow-a", "First workflow", 3)
	writeProcessionTemplate(t, root, "workflow-b", "Second workflow", 2)

	resolver := paths.NewResolver(root)
	entries := CollectProcessions(resolver)

	require.Len(t, entries, 2)
	names := []string{entries[0].Name, entries[1].Name}
	assert.Contains(t, names, "workflow-a")
	assert.Contains(t, names, "workflow-b")
}
