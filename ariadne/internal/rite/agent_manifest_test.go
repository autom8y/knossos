package rite

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewEmptyManifest(t *testing.T) {
	m := NewEmptyAgentManifest()

	if m.Version != AgentManifestVersion {
		t.Errorf("Version = %q, want %q", m.Version, AgentManifestVersion)
	}

	if m.Agents == nil {
		t.Error("Agents is nil, want initialized map")
	}

	if m.Orphans == nil {
		t.Error("Orphans is nil, want initialized slice")
	}
}

func TestManifest_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "AGENT_MANIFEST.json")

	// Create and save manifest
	m := NewEmptyAgentManifest()
	m.ActiveRite = "test-team"
	m.AddAgent("agent-a.md", "rite", "test-team", "sha256:abc123")

	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadAgentManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadAgentManifest() error = %v", err)
	}

	if loaded.ActiveRite != "test-team" {
		t.Errorf("ActiveRite = %q, want %q", loaded.ActiveRite, "test-team")
	}

	if len(loaded.Agents) != 1 {
		t.Errorf("len(Agents) = %d, want 1", len(loaded.Agents))
	}

	agent, ok := loaded.Agents["agent-a.md"]
	if !ok {
		t.Fatal("agent-a.md not found in Agents")
	}

	if agent.Source != "rite" {
		t.Errorf("agent.Source = %q, want %q", agent.Source, "rite")
	}
	if agent.Origin != "test-team" {
		t.Errorf("agent.Origin = %q, want %q", agent.Origin, "test-team")
	}
	if agent.Checksum != "sha256:abc123" {
		t.Errorf("agent.Checksum = %q, want %q", agent.Checksum, "sha256:abc123")
	}
}

func TestManifest_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "does-not-exist.json")

	m, err := LoadAgentManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadAgentManifest() error = %v (should return empty manifest)", err)
	}

	if m == nil {
		t.Fatal("LoadAgentManifest() returned nil for non-existent file")
	}

	if m.Version != AgentManifestVersion {
		t.Errorf("Version = %q, want %q", m.Version, AgentManifestVersion)
	}
}

func TestManifest_DetectOrphans(t *testing.T) {
	m := NewEmptyAgentManifest()
	m.ActiveRite = "team-a"
	m.AddAgent("agent-a.md", "rite", "team-a", "sha256:a")
	m.AddAgent("agent-b.md", "rite", "team-a", "sha256:b")
	m.AddAgent("agent-c.md", "rite", "team-b", "sha256:c")
	m.AddAgent("project-agent.md", "project", "", "sha256:d")

	// Switch to team-b - agents from team-a become orphans
	orphans := m.DetectOrphans("team-b")

	if len(orphans) != 2 {
		t.Errorf("DetectOrphans() returned %d orphans, want 2", len(orphans))
	}

	// Check that agent-a and agent-b are orphans
	orphanSet := make(map[string]bool)
	for _, o := range orphans {
		orphanSet[o] = true
	}

	if !orphanSet["agent-a.md"] {
		t.Error("agent-a.md should be orphan")
	}
	if !orphanSet["agent-b.md"] {
		t.Error("agent-b.md should be orphan")
	}
	if orphanSet["agent-c.md"] {
		t.Error("agent-c.md should not be orphan (from team-b)")
	}
	if orphanSet["project-agent.md"] {
		t.Error("project-agent.md should not be orphan (project source)")
	}
}

func TestManifest_MarkOrphaned(t *testing.T) {
	m := NewEmptyAgentManifest()
	m.AddAgent("agent-a.md", "rite", "team-a", "sha256:a")

	m.MarkOrphaned("agent-a.md")

	agent := m.Agents["agent-a.md"]
	if !agent.Orphaned {
		t.Error("agent.Orphaned = false, want true")
	}

	if len(m.Orphans) != 1 || m.Orphans[0] != "agent-a.md" {
		t.Errorf("Orphans = %v, want [agent-a.md]", m.Orphans)
	}
}

func TestManifest_PromoteToProject(t *testing.T) {
	m := NewEmptyAgentManifest()
	m.AddAgent("agent-a.md", "rite", "team-a", "sha256:a")
	m.MarkOrphaned("agent-a.md")

	m.PromoteToProject("agent-a.md")

	agent := m.Agents["agent-a.md"]
	if agent.Source != "project" {
		t.Errorf("agent.Source = %q, want %q", agent.Source, "project")
	}
	if agent.Origin != "" {
		t.Errorf("agent.Origin = %q, want empty", agent.Origin)
	}
	if agent.Orphaned {
		t.Error("agent.Orphaned = true, want false")
	}

	if len(m.Orphans) != 0 {
		t.Errorf("len(Orphans) = %d, want 0", len(m.Orphans))
	}
}

func TestManifest_RemoveAgent(t *testing.T) {
	m := NewEmptyAgentManifest()
	m.AddAgent("agent-a.md", "rite", "team-a", "sha256:a")

	if len(m.Agents) != 1 {
		t.Fatalf("len(Agents) = %d before remove, want 1", len(m.Agents))
	}

	m.RemoveAgent("agent-a.md")

	if len(m.Agents) != 0 {
		t.Errorf("len(Agents) = %d after remove, want 0", len(m.Agents))
	}
}

func TestManifest_GetRiteAgents(t *testing.T) {
	m := NewEmptyAgentManifest()
	m.AddAgent("agent-a.md", "rite", "team-a", "sha256:a")
	m.AddAgent("agent-b.md", "rite", "team-a", "sha256:b")
	m.AddAgent("agent-c.md", "rite", "team-b", "sha256:c")
	m.AddAgent("project-agent.md", "project", "", "sha256:d")

	agents := m.GetRiteAgents("team-a")

	if len(agents) != 2 {
		t.Errorf("GetRiteAgents(team-a) returned %d agents, want 2", len(agents))
	}
}

func TestComputeChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	checksum, err := ComputeChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeChecksum() error = %v", err)
	}

	if checksum == "" {
		t.Error("ComputeChecksum() returned empty string")
	}

	if len(checksum) < 10 {
		t.Errorf("ComputeChecksum() returned short checksum: %q", checksum)
	}

	// Verify checksum starts with sha256:
	if checksum[:7] != "sha256:" {
		t.Errorf("checksum doesn't start with sha256: %q", checksum)
	}
}

func TestManifest_AgentEntry_InstalledAt(t *testing.T) {
	m := NewEmptyAgentManifest()
	before := time.Now().Add(-time.Second)
	m.AddAgent("agent.md", "rite", "test", "sha256:x")
	after := time.Now().Add(time.Second)

	agent := m.Agents["agent.md"]
	if agent.InstalledAt.Before(before) || agent.InstalledAt.After(after) {
		t.Errorf("InstalledAt = %v, expected between %v and %v", agent.InstalledAt, before, after)
	}
}
