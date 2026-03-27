package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

// --- dismiss standing agent guard ---

func TestDismissCmd_StandingAgent_ReturnsError(t *testing.T) {
	for _, name := range []string{"pythia", "moirai"} {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputFmt := "text"
			verbose := false
			cmd := NewAgentCmd(&outputFmt, &verbose, &tmpDir)
			cmd.SetArgs([]string{"dismiss", name})
			err := cmd.Execute()
			if err == nil {
				t.Errorf("dismiss %q: expected error for standing agent, got nil", name)
			}
		})
	}
}

// --- dismiss non-summoned agent ---

func TestDismissCmd_NoManifestEntry_ReturnsError(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"dismiss", "some-agent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when dismissing agent with no provenance entry")
	}
}

func TestDismissCmd_ManifestEntryNotSummon_ReturnsError(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Pre-populate manifest with non-summon entry
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/non-summon-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "rites/ecosystem/agents/non-summon-agent.md", // NOT summon:
				SourceType: "project",
				Checksum:   "sha256:" + strings.Repeat("a", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to save test manifest: %v", err)
	}

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"dismiss", "non-summon-agent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when dismissing non-summon agent")
	}
}

// --- dismiss with pre-populated manifest ---

func TestDismissCmd_SummonedAgent_RemovesFileAndEntry(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Write a fake summoned agent file
	agentFilePath := filepath.Join(agentsDir, "my-agent.md")
	if err := os.WriteFile(agentFilePath, []byte("---\nname: my-agent\n---\n# Body\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-populate manifest with summon entry
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/my-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "summon:my-agent",
				SourceType: "summon",
				Checksum:   "sha256:" + strings.Repeat("b", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to save test manifest: %v", err)
	}

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"dismiss", "my-agent"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("dismiss returned error: %v", err)
	}

	// File should be gone
	if _, statErr := os.Stat(agentFilePath); !os.IsNotExist(statErr) {
		t.Error("agent file should have been removed after dismiss")
	}

	// Provenance entry should be removed
	updatedManifest, loadErr := provenance.Load(manifestPath)
	if loadErr != nil {
		t.Fatalf("failed to load manifest after dismiss: %v", loadErr)
	}
	if _, ok := updatedManifest.Entries["agents/my-agent.md"]; ok {
		t.Error("provenance entry should have been removed after dismiss")
	}
}

func TestDismissCmd_SummonedAgent_FileMissing_SucceedsWithWarning(t *testing.T) {
	// Dismiss should succeed even if the file was already manually deleted.
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Pre-populate manifest with summon entry — but no file on disk
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/gone-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "summon:gone-agent",
				SourceType: "summon",
				Checksum:   "sha256:" + strings.Repeat("c", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to save test manifest: %v", err)
	}

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"dismiss", "gone-agent"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("dismiss of missing file should succeed (warn), got error: %v", err)
	}

	// Provenance entry should still be cleaned up
	updatedManifest, loadErr := provenance.Load(manifestPath)
	if loadErr != nil {
		t.Fatalf("failed to load manifest after dismiss: %v", loadErr)
	}
	if _, ok := updatedManifest.Entries["agents/gone-agent.md"]; ok {
		t.Error("provenance entry should have been removed even when file was missing")
	}
}

// --- force flag ---

func TestDismissCmd_Force_RemovesWithoutManifestEntry(t *testing.T) {
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Write a fake agent file with no manifest entry
	agentFilePath := filepath.Join(agentsDir, "force-agent.md")
	if err := os.WriteFile(agentFilePath, []byte("---\nname: force-agent\n---\n# Body\n"), 0644); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	outputFmt := "text"
	verbose := false
	cmd := NewAgentCmd(&outputFmt, &verbose, &projectDir)
	cmd.SetArgs([]string{"dismiss", "force-agent", "--force"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("dismiss --force returned error: %v", err)
	}

	// File should be gone
	if _, statErr := os.Stat(agentFilePath); !os.IsNotExist(statErr) {
		t.Error("agent file should have been removed with --force even without manifest entry")
	}
}
