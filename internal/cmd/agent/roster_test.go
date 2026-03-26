package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// --- roster subcommand: metadata ---

func TestRosterCmd_Use(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, err := cmd.Find([]string{"roster"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("agent command missing 'roster' subcommand")
	}
	if !strings.HasPrefix(sub.Use, "roster") {
		t.Errorf("roster subcommand Use = %q, want prefix 'roster'", sub.Use)
	}
}

func TestRosterCmd_ShortDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"roster"})
	if sub.Short == "" {
		t.Error("roster subcommand Short is empty")
	}
}

func TestRosterCmd_LongDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"roster"})
	if sub.Long == "" {
		t.Error("roster subcommand Long is empty")
	}
}

// --- rosterOutput.Text ---

func TestRosterOutput_Text_ContainsThreeSections(t *testing.T) {
	r := rosterOutput{
		Standing:  []rosterEntry{{Name: "pythia", Section: "standing"}},
		Summoned:  []rosterEntry{{Name: "theoros", Section: "summoned", SummonedAt: "2026-03-26T00:00:00Z"}},
		Available: []rosterEntry{{Name: "naxos", Section: "available", Description: "An available agent"}},
	}
	text := r.Text()

	if !strings.Contains(text, "Standing") {
		t.Error("Text() should contain 'Standing' section header")
	}
	if !strings.Contains(text, "Summoned") {
		t.Error("Text() should contain 'Summoned' section header")
	}
	if !strings.Contains(text, "Available") {
		t.Error("Text() should contain 'Available' section header")
	}

	if !strings.Contains(text, "pythia") {
		t.Error("Text() should contain 'pythia' in Standing section")
	}
	if !strings.Contains(text, "theoros") {
		t.Error("Text() should contain 'theoros' in Summoned section")
	}
	if !strings.Contains(text, "naxos") {
		t.Error("Text() should contain 'naxos' in Available section")
	}
}

func TestRosterOutput_Text_EmptySectionsShowNone(t *testing.T) {
	r := rosterOutput{
		Standing:  []rosterEntry{{Name: "pythia", Section: "standing"}},
		Summoned:  nil,
		Available: nil,
	}
	text := r.Text()

	// Both empty sections should say "(none)"
	noneCount := strings.Count(text, "(none)")
	if noneCount < 2 {
		t.Errorf("expected at least 2 '(none)' placeholders, got %d in:\n%s", noneCount, text)
	}
}

// --- buildStandingSection ---

func TestBuildStandingSection_ContainsAllThreeAgents(t *testing.T) {
	entries := buildStandingSection("")
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
	}
	for _, expected := range []string{"pythia", "moirai", "metis"} {
		if !names[expected] {
			t.Errorf("buildStandingSection: missing expected agent %q", expected)
		}
	}
}

func TestBuildStandingSection_SectionIsStanding(t *testing.T) {
	entries := buildStandingSection("")
	for _, e := range entries {
		if e.Section != "standing" {
			t.Errorf("entry %q has section %q, want 'standing'", e.Name, e.Section)
		}
	}
}

// --- buildSummonedSection ---

func TestBuildSummonedSection_EmptyManifest_ReturnsEmpty(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	// No manifest file → should degrade gracefully
	entries := buildSummonedSection(claudeDir)
	if len(entries) != 0 {
		t.Errorf("buildSummonedSection with no manifest should return empty, got %d entries", len(entries))
	}
}

func TestBuildSummonedSection_FiltersSummonEntries(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/summoned-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "summon:summoned-agent",
				SourceType: "summon",
				Checksum:   "sha256:" + strings.Repeat("a", 64),
				LastSynced: time.Now().UTC(),
			},
			"agents/rite-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "rites/ecosystem/agents/rite-agent.md", // NOT summon
				SourceType: "project",
				Checksum:   "sha256:" + strings.Repeat("b", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to save test manifest: %v", err)
	}

	entries := buildSummonedSection(claudeDir)
	if len(entries) != 1 {
		t.Fatalf("buildSummonedSection should return 1 summoned entry, got %d", len(entries))
	}
	if entries[0].Name != "summoned-agent" {
		t.Errorf("summoned entry name = %q, want %q", entries[0].Name, "summoned-agent")
	}
	if entries[0].Section != "summoned" {
		t.Errorf("summoned entry section = %q, want %q", entries[0].Section, "summoned")
	}
}

// --- buildAvailableSection ---

func TestBuildAvailableSection_ExcludesSummonedAgents(t *testing.T) {
	knossosHome := t.TempDir()
	agentsDir := filepath.Join(knossosHome, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KNOSSOS_HOME", knossosHome)

	// Write a summonable agent
	summonableContent := `---
name: available-agent
description: Available agent for testing
tools:
  - Read
tier: summonable
---
# Body
`
	if err := os.WriteFile(filepath.Join(agentsDir, "available-agent.md"), []byte(summonableContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Mark it as already summoned
	summonedNames := map[string]bool{"available-agent": true}
	entries := buildAvailableSection(summonedNames)

	for _, e := range entries {
		if e.Name == "available-agent" {
			t.Error("buildAvailableSection should exclude already-summoned agents")
		}
	}
}

func TestBuildAvailableSection_OnlyIncludesSummonableTier(t *testing.T) {
	knossosHome := t.TempDir()
	agentsDir := filepath.Join(knossosHome, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KNOSSOS_HOME", knossosHome)

	// Write a summonable agent
	summonableContent := `---
name: yes-agent
description: Summonable agent
tools:
  - Read
tier: summonable
---
# Body
`
	// Write a non-summonable agent (no tier field)
	notSummonableContent := `---
name: no-agent
description: Not summonable agent
tools:
  - Read
---
# Body
`
	if err := os.WriteFile(filepath.Join(agentsDir, "yes-agent.md"), []byte(summonableContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "no-agent.md"), []byte(notSummonableContent), 0644); err != nil {
		t.Fatal(err)
	}

	entries := buildAvailableSection(nil)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
	}

	if !names["yes-agent"] {
		t.Error("buildAvailableSection should include agent with tier: summonable")
	}
	if names["no-agent"] {
		t.Error("buildAvailableSection should exclude agent without tier: summonable")
	}
}

func TestBuildAvailableSection_ExcludesStandingAgents(t *testing.T) {
	knossosHome := t.TempDir()
	agentsDir := filepath.Join(knossosHome, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KNOSSOS_HOME", knossosHome)

	// Write a standing agent with tier: summonable (should still be excluded)
	pythiaContent := `---
name: pythia
description: Standing agent
tools:
  - Read
tier: summonable
---
# Body
`
	if err := os.WriteFile(filepath.Join(agentsDir, "pythia.md"), []byte(pythiaContent), 0644); err != nil {
		t.Fatal(err)
	}

	entries := buildAvailableSection(nil)
	for _, e := range entries {
		if standingAgents[e.Name] {
			t.Errorf("buildAvailableSection should exclude standing agent %q", e.Name)
		}
	}
}

// --- truncateDescription ---

func TestTruncateDescription_ShortString_Unchanged(t *testing.T) {
	s := "Short description"
	result := truncateDescription(s)
	if result != s {
		t.Errorf("truncateDescription(%q) = %q, want %q", s, result, s)
	}
}

func TestTruncateDescription_LongString_Truncated(t *testing.T) {
	s := strings.Repeat("a", 70)
	result := truncateDescription(s)
	if len(result) > 60 {
		t.Errorf("truncateDescription should truncate to <=60 chars, got %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("truncated description should end with '...'")
	}
}

func TestTruncateDescription_MultiLine_TakesFirstLine(t *testing.T) {
	s := "First line\nSecond line\nThird line"
	result := truncateDescription(s)
	if strings.Contains(result, "Second") || strings.Contains(result, "Third") {
		t.Error("truncateDescription should only use first line of multi-line description")
	}
	if !strings.Contains(result, "First") {
		t.Error("truncateDescription should contain first line content")
	}
}

// --- extractTierFromBytes ---

func TestExtractTierFromBytes_WithSummonableTier(t *testing.T) {
	content := []byte("---\nname: test\ntier: summonable\n---\n# Body\n")
	tier := extractTierFromBytes(content)
	if tier != "summonable" {
		t.Errorf("extractTierFromBytes = %q, want %q", tier, "summonable")
	}
}

func TestExtractTierFromBytes_WithAbsentTier(t *testing.T) {
	content := []byte("---\nname: test\n---\n# Body\n")
	tier := extractTierFromBytes(content)
	if tier != "" {
		t.Errorf("extractTierFromBytes = %q, want empty string", tier)
	}
}

// --- Integration test: summon -> roster -> dismiss -> roster ---

func TestIntegration_SummonRosterDismissRoster(t *testing.T) {
	// Set up KNOSSOS_HOME with a summonable agent
	knossosHome := t.TempDir()
	agentsDir := filepath.Join(knossosHome, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	agentContent := `---
name: integration-agent
description: Integration test agent for summon/dismiss cycle
tools:
  - Read
tier: summonable
---

# Integration Agent

Test body for integration test.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "integration-agent.md"), []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Setenv("HOME", fakeHome)

	projectDir := t.TempDir()

	// Step 1: Summon the agent
	{
		outputFmt := "text"
		verbose := false
		cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
		cmd.SetArgs([]string{"summon", "integration-agent"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("summon failed: %v", err)
		}
	}

	// Step 2: Roster should show it in Summoned section
	{
		summoned := buildSummonedSection(claudeDir)
		found := false
		for _, e := range summoned {
			if e.Name == "integration-agent" {
				found = true
				break
			}
		}
		if !found {
			t.Error("after summon, roster should show integration-agent in Summoned section")
		}

		// Should not be in Available (already summoned)
		available := buildAvailableSection(map[string]bool{"integration-agent": true})
		for _, e := range available {
			if e.Name == "integration-agent" {
				t.Error("after summon, roster should not show integration-agent in Available section")
			}
		}
	}

	// Step 3: Dismiss the agent
	{
		outputFmt := "text"
		verbose := false
		cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
		cmd.SetArgs([]string{"dismiss", "integration-agent"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("dismiss failed: %v", err)
		}
	}

	// Step 4: Roster should show it in Available section again (not Summoned)
	{
		summoned := buildSummonedSection(claudeDir)
		for _, e := range summoned {
			if e.Name == "integration-agent" {
				t.Error("after dismiss, roster should not show integration-agent in Summoned section")
			}
		}

		available := buildAvailableSection(nil)
		found := false
		for _, e := range available {
			if e.Name == "integration-agent" {
				found = true
				break
			}
		}
		if !found {
			t.Error("after dismiss, roster should show integration-agent in Available section")
		}
	}
}
