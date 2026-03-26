package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// --- transformForSummon ---

func TestTransformForSummon_StripsKnossosFields(t *testing.T) {
	content := []byte(`---
name: test-agent
description: A test agent
type: specialist
role: analyst
upstream:
  - phase: analysis
tier: summonable
tools:
  - Read
  - Bash
---

# Test Agent
`)
	result, err := transformForSummon(content, "test-agent")
	if err != nil {
		t.Fatalf("transformForSummon returned error: %v", err)
	}

	// Should not contain knossos-only fields
	s := string(result)
	for _, field := range []string{"type:", "role:", "upstream:", "tier:"} {
		if strings.Contains(s, field) {
			t.Errorf("transformed output still contains field %q", field)
		}
	}

	// Should retain CC-native fields
	for _, field := range []string{"description:", "tools:", "name:"} {
		if !strings.Contains(s, field) {
			t.Errorf("transformed output missing expected field %q", field)
		}
	}
}

func TestTransformForSummon_InjectsName(t *testing.T) {
	content := []byte(`---
name: original-name
description: A test agent
tools:
  - Read
---

# Agent body
`)
	result, err := transformForSummon(content, "injected-name")
	if err != nil {
		t.Fatalf("transformForSummon returned error: %v", err)
	}

	if !strings.Contains(string(result), "name: injected-name") {
		t.Errorf("transformed output should contain injected name 'injected-name', got:\n%s", string(result))
	}
}

func TestTransformForSummon_PassesThroughInvalidFrontmatter(t *testing.T) {
	// Content without frontmatter should pass through unchanged
	content := []byte("# No frontmatter here\n\nJust body text.\n")
	result, err := transformForSummon(content, "agent-name")
	if err != nil {
		t.Fatalf("transformForSummon returned error on no-frontmatter content: %v", err)
	}
	if string(result) != string(content) {
		t.Error("content without frontmatter should pass through unchanged")
	}
}

func TestTransformForSummon_PreservesBody(t *testing.T) {
	body := "# Agent\n\nThis is the agent body content.\n\n## Section\n\nMore content.\n"
	content := []byte("---\nname: test\ndescription: Test agent\ntools:\n  - Read\n---\n" + body)

	result, err := transformForSummon(content, "test")
	if err != nil {
		t.Fatalf("transformForSummon returned error: %v", err)
	}

	if !strings.Contains(string(result), "agent body content") {
		t.Errorf("body content should be preserved, got:\n%s", string(result))
	}
}

// --- extractTierField ---

func TestExtractTierField_WithSummonableTier(t *testing.T) {
	content := []byte("---\nname: test\ntier: summonable\n---\n# Body\n")
	tier := extractTierField(content)
	if tier != "summonable" {
		t.Errorf("extractTierField = %q, want %q", tier, "summonable")
	}
}

func TestExtractTierField_WithAbsentTier(t *testing.T) {
	content := []byte("---\nname: test\n---\n# Body\n")
	tier := extractTierField(content)
	if tier != "" {
		t.Errorf("extractTierField = %q, want %q (empty)", tier, "")
	}
}

func TestExtractTierField_WithInvalidFrontmatter(t *testing.T) {
	content := []byte("No frontmatter here\n")
	tier := extractTierField(content)
	if tier != "" {
		t.Errorf("extractTierField = %q, want empty string for invalid frontmatter", tier)
	}
}

// --- standingAgents deny-list ---

func TestSummonCmd_StandingAgent_ReturnsError(t *testing.T) {
	for _, name := range []string{"pythia", "moirai", "metis"} {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputFmt := "text"
			verbose := false
			cmd := NewAgentCmd(&outputFmt, &verbose, &tmpDir)
			cmd.SetArgs([]string{"summon", name})
			err := cmd.Execute()
			if err == nil {
				t.Errorf("summon %q: expected error for standing agent, got nil", name)
			}
		})
	}
}

// --- summon with mock source dir (KNOSSOS_HOME approach) ---

func TestSummonCmd_ValidAgent_WritesFile(t *testing.T) {
	// Set up KNOSSOS_HOME with a fake summonable agent
	knossosHome := t.TempDir()
	agentsDir := filepath.Join(knossosHome, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	agentContent := `---
name: fake-agent
description: A fake summonable agent for testing
tools:
  - Read
tier: summonable
---

# Fake Agent

Test body.
`
	agentPath := filepath.Join(agentsDir, "fake-agent.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up fake ~/.claude/ user channel dir
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	// Override HOME and KNOSSOS_HOME for this test
	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Setenv("HOME", fakeHome)

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"summon", "fake-agent"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("summon returned error: %v", err)
	}

	// Verify the file was written
	targetPath := filepath.Join(claudeDir, "agents", "fake-agent.md")
	data, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatalf("expected agent file at %s, got error: %v", targetPath, readErr)
	}

	// Verify knossos fields were stripped
	if strings.Contains(string(data), "tier:") {
		t.Error("summoned agent file should not contain 'tier:' field")
	}
	// Verify name was injected
	if !strings.Contains(string(data), "name: fake-agent") {
		t.Errorf("summoned agent should contain 'name: fake-agent', got:\n%s", string(data))
	}
}

func TestSummonCmd_AgentNotFound_ReturnsError(t *testing.T) {
	knossosHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(knossosHome, "agents"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KNOSSOS_HOME", knossosHome)

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"summon", "nonexistent-agent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent agent, got nil")
	}
}

func TestSummonCmd_UpdatesProvenanceManifest(t *testing.T) {
	knossosHome := t.TempDir()
	agentsDir := filepath.Join(knossosHome, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	agentContent := `---
name: prov-agent
description: Provenance test agent
tools:
  - Read
tier: summonable
---

# Prov Agent
`
	if err := os.WriteFile(filepath.Join(agentsDir, "prov-agent.md"), []byte(agentContent), 0644); err != nil {
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
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"summon", "prov-agent"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("summon returned error: %v", err)
	}

	// Read provenance manifest and verify entry
	manifestPath := provenance.UserManifestPath(claudeDir)
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load provenance manifest: %v", err)
	}

	entry, ok := manifest.Entries["agents/prov-agent.md"]
	if !ok {
		t.Fatal("expected provenance entry for 'agents/prov-agent.md'")
	}
	if entry.SourcePath != "summon:prov-agent" {
		t.Errorf("entry.SourcePath = %q, want %q", entry.SourcePath, "summon:prov-agent")
	}
	if entry.SourceType != "summon" {
		t.Errorf("entry.SourceType = %q, want %q", entry.SourceType, "summon")
	}
	if entry.Owner != provenance.OwnerKnossos {
		t.Errorf("entry.Owner = %q, want %q", entry.Owner, provenance.OwnerKnossos)
	}
	if entry.Scope != provenance.ScopeUser {
		t.Errorf("entry.Scope = %q, want %q", entry.Scope, provenance.ScopeUser)
	}
	if !strings.HasPrefix(entry.Checksum, "sha256:") {
		t.Errorf("entry.Checksum = %q, want sha256: prefix", entry.Checksum)
	}
	if entry.LastSynced.IsZero() {
		t.Error("entry.LastSynced should not be zero")
	}
}

// --- collision checker ---

func TestRiteCollisionChecker_NoknossosDirReturnsIneffective(t *testing.T) {
	checker := newCollisionCheckerForSummon("/nonexistent/knossos/dir")
	if checker.IsEffective() {
		t.Error("checker for nonexistent dir should not be effective")
	}
	if collides, _ := checker.CheckCollision("agents/test.md"); collides {
		t.Error("ineffective checker should never report collision")
	}
}

func TestRiteCollisionChecker_EffectiveWithManifest(t *testing.T) {
	// Build a minimal provenance manifest with a rite-owned agent
	knossosDir := t.TempDir()
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/rite-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeRite,
				SourcePath: "rites/test/agents/rite-agent.md",
				SourceType: "project",
				Checksum:   "sha256:" + strings.Repeat("a", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.ManifestPath(knossosDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to save test manifest: %v", err)
	}

	checker := newCollisionCheckerForSummon(knossosDir)
	if !checker.IsEffective() {
		t.Error("checker with valid manifest should be effective")
	}

	// Collides with rite-owned agent
	collides, reason := checker.CheckCollision("agents/rite-agent.md")
	if !collides {
		t.Error("should detect collision with rite-owned agent")
	}
	if reason == "" {
		t.Error("collision reason should not be empty")
	}

	// Does not collide with different agent
	collides, _ = checker.CheckCollision("agents/other-agent.md")
	if collides {
		t.Error("should not detect collision for agent not in rite manifest")
	}
}
