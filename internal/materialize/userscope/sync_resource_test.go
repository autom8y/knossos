package userscope

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/provenance"
)

// setupAgentSync creates a minimal filesystem layout for testing syncUserResource
// with ResourceAgents. Returns (knossosHome, userChannelDir).
func setupAgentSync(t *testing.T, agentFiles map[string]string) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	knossosHome := filepath.Join(tmpDir, "knossos")
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	agentsSourceDir := filepath.Join(knossosHome, "agents")
	agentsTargetDir := filepath.Join(userChannelDir, "agents")
	os.MkdirAll(agentsSourceDir, 0755)
	os.MkdirAll(agentsTargetDir, 0755)

	for name, content := range agentFiles {
		os.WriteFile(filepath.Join(agentsSourceDir, name), []byte(content), 0644)
	}
	return knossosHome, userChannelDir
}

func TestSyncUserResource_AddsNewAgent(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"pythia.md": "# Pythia Agent\nPrompt content.\n",
	})

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Summary.Added)
	}

	// Verify file was copied
	targetPath := filepath.Join(userChannelDir, "agents", "pythia.md")
	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "# Pythia Agent\nPrompt content.\n" {
		t.Errorf("unexpected content: %q", got)
	}

	// Verify manifest entry
	entry, exists := manifest.Entries["agents/pythia.md"]
	if !exists {
		t.Fatal("expected manifest entry for agents/pythia.md")
	}
	if entry.Owner != provenance.OwnerKnossos {
		t.Errorf("expected owner knossos, got %s", entry.Owner)
	}
}

func TestSyncUserResource_SkipsUserOwned(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"my-agent.md": "# New Source Content\n",
	})

	// Pre-create user-modified version at target
	targetPath := filepath.Join(userChannelDir, "agents", "my-agent.md")
	os.WriteFile(targetPath, []byte("# User Modified Content\n"), 0644)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/my-agent.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Summary.Skipped)
	}
	if result.Summary.Added != 0 {
		t.Errorf("expected 0 added, got %d", result.Summary.Added)
	}

	// User content should be preserved
	got, _ := os.ReadFile(targetPath)
	if string(got) != "# User Modified Content\n" {
		t.Errorf("user content should be preserved, got %q", got)
	}
}

func TestSyncUserResource_UpdatesKnossosOwned(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": "# Updated Source Content\n",
	})

	// Pre-create old version at target
	oldContent := []byte("# Old Content\n")
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, oldContent, 0644)
	oldChecksum := checksum.Bytes(oldContent)

	// Manifest says knossos owns it with the old checksum
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: oldChecksum,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Updated != 1 {
		t.Errorf("expected 1 updated, got %d", result.Summary.Updated)
	}

	// Target should have new content
	got, _ := os.ReadFile(targetPath)
	if string(got) != "# Updated Source Content\n" {
		t.Errorf("target should be updated, got %q", got)
	}
}

func TestSyncUserResource_UnchangedKnossosOwned(t *testing.T) {
	t.Parallel()
	content := "# Same Content\n"
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": content,
	})

	// Pre-create same version at target
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, []byte(content), 0644)
	sourceChecksum := checksum.Bytes([]byte(content))

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: sourceChecksum,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Unchanged != 1 {
		t.Errorf("expected 1 unchanged, got %d", result.Summary.Unchanged)
	}
}

func TestSyncUserResource_SkipsDiverged(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": "# New Source Content\n",
	})

	// Pre-create a locally modified version (diverged from both old and new source)
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, []byte("# Locally Modified\n"), 0644)

	// Manifest says knossos owns it with a different checksum than both source and target
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: "sha256:deadbeef",
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped (diverged), got %d", result.Summary.Skipped)
	}

	// Local content should be preserved
	got, _ := os.ReadFile(targetPath)
	if string(got) != "# Locally Modified\n" {
		t.Errorf("local content should be preserved, got %q", got)
	}
}

func TestSyncUserResource_OverwritesDiverged(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": "# New Source\n",
	})

	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, []byte("# Locally Modified\n"), 0644)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: "sha256:deadbeef",
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{
		OverwriteDiverged: true,
	})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Updated != 1 {
		t.Errorf("expected 1 updated (force overwrite), got %d", result.Summary.Updated)
	}

	got, _ := os.ReadFile(targetPath)
	if string(got) != "# New Source\n" {
		t.Errorf("target should be overwritten, got %q", got)
	}
}

func TestSyncUserResource_CollisionSkipped(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"pythia.md": "# Pythia\n",
	})

	checker := &CollisionChecker{
		manifestLoaded: true,
		riteEntries:    map[string]bool{"agents/pythia.md": true},
	}
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Collisions != 1 {
		t.Errorf("expected 1 collision, got %d", result.Summary.Collisions)
	}

	// File should NOT be copied
	targetPath := filepath.Join(userChannelDir, "agents", "pythia.md")
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Error("colliding file should not be copied to target")
	}
}

func TestSyncUserResource_OrphanRemoval(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"keeper.md": "# Keeper\n",
	})

	// Pre-create an orphan file (in manifest but not in source anymore)
	orphanPath := filepath.Join(userChannelDir, "agents", "orphan.md")
	os.WriteFile(orphanPath, []byte("# Orphan\n"), 0644)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/orphan.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: "sha256:whatever",
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	_, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	// Orphan file should be removed
	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Error("expected orphan file to be removed")
	}
	// Orphan manifest entry should be removed
	if _, exists := manifest.Entries["agents/orphan.md"]; exists {
		t.Error("expected orphan manifest entry to be removed")
	}
}

func TestSyncUserResource_DryRun(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"new-agent.md": "# New Agent\n",
	})

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Added != 1 {
		t.Errorf("dry-run should still report 1 added, got %d", result.Summary.Added)
	}

	// File should NOT be created
	targetPath := filepath.Join(userChannelDir, "agents", "new-agent.md")
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Error("dry-run should not create files")
	}

	// Manifest should not be updated
	if len(manifest.Entries) != 0 {
		t.Error("dry-run should not update manifest entries")
	}
}

func TestSyncUserResource_RecreateDeletion_KnossosOwned(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": "# Agent Content\n",
	})

	// Manifest says knossos owns it, but target file was deleted from disk
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: "sha256:old-checksum",
			},
		},
	}
	// Note: target file does NOT exist
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added (re-created), got %d", result.Summary.Added)
	}

	// File should be re-created
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("expected deleted knossos-owned file to be re-created")
	}
}

func TestSyncUserResource_RecreateDeletion_UserOwned(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": "# Source Content\n",
	})

	// Manifest says user owns it, but target file was deleted from disk
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	// Deleted user-owned file should be re-created from source
	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added (re-created from source), got %d", result.Summary.Added)
	}
}

func TestSyncUserResource_NewFileTargetExists_Untracked(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent.md": "# Source Content\n",
	})

	// Pre-create a file at the target (not in manifest — untracked)
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, []byte("# User Pre-existing\n"), 0644)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped (user-created), got %d", result.Summary.Skipped)
	}

	// User content should be preserved
	got, _ := os.ReadFile(targetPath)
	if string(got) != "# User Pre-existing\n" {
		t.Errorf("untracked user content should be preserved, got %q", got)
	}

	// Manifest should now track it as user-owned
	entry, exists := manifest.Entries["agents/agent.md"]
	if !exists {
		t.Fatal("expected manifest entry for newly tracked user file")
	}
	if entry.Owner != provenance.OwnerUser {
		t.Errorf("expected owner user, got %s", entry.Owner)
	}
}

func TestSyncUserResource_SourceMissing_NoEmbedded(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	knossosHome := filepath.Join(tmpDir, "knossos")
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)
	// Note: knossosHome/agents does NOT exist

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{} // no embeddedAgents

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	// Should return empty result (no-op)
	if result.Summary.Added+result.Summary.Updated+result.Summary.Skipped != 0 {
		t.Error("expected no-op when source directory missing and no embedded fallback")
	}
}

func TestSyncUserResource_InvalidResourceType(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	_, err := s.syncUserResource(SyncResource("invalid"), tmpDir, tmpDir, manifest, checker, SyncOptions{})
	if err == nil {
		t.Error("expected error for invalid resource type")
	}
}

func TestSyncUserResource_MultipleAgents(t *testing.T) {
	t.Parallel()
	knossosHome, userChannelDir := setupAgentSync(t, map[string]string{
		"agent-a.md": "# Agent A\n",
		"agent-b.md": "# Agent B\n",
		"agent-c.md": "# Agent C\n",
	})

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Added != 3 {
		t.Errorf("expected 3 added, got %d", result.Summary.Added)
	}
	if len(manifest.Entries) != 3 {
		t.Errorf("expected 3 manifest entries, got %d", len(manifest.Entries))
	}
}

func TestSyncUserResource_Hooks_NestedStructure(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	knossosHome := filepath.Join(tmpDir, "knossos")
	userChannelDir := filepath.Join(tmpDir, "user-claude")

	// Create nested hooks structure
	hooksDir := filepath.Join(knossosHome, "hooks")
	os.MkdirAll(filepath.Join(hooksDir, "lib"), 0755)
	os.WriteFile(filepath.Join(hooksDir, "pre-commit.sh"), []byte("#!/bin/sh\ncheck"), 0644)
	os.WriteFile(filepath.Join(hooksDir, "lib", "helpers.sh"), []byte("#!/bin/sh\nhelp"), 0644)

	os.MkdirAll(filepath.Join(userChannelDir, "hooks"), 0755)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResource(ResourceHooks, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Added != 2 {
		t.Errorf("expected 2 added (nested hooks), got %d", result.Summary.Added)
	}

	// Verify nested path preserved
	libHelper := filepath.Join(userChannelDir, "hooks", "lib", "helpers.sh")
	if _, err := os.Stat(libHelper); os.IsNotExist(err) {
		t.Error("expected nested hook file to be created")
	}

	// Verify manifest keys use hooks/ prefix with nested path
	if _, exists := manifest.Entries["hooks/lib/helpers.sh"]; !exists {
		t.Errorf("expected manifest key hooks/lib/helpers.sh, got keys: %v", manifestKeys(manifest))
	}
}

// --- Tests for syncUserResourceFromEmbedded ---

func TestSyncUserResourceFromEmbedded_AddsNewAgent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	embeddedFS := fstest.MapFS{
		"agents/pythia.md": &fstest.MapFile{Data: []byte("# Embedded Pythia\n")},
		"agents/architect.md":  &fstest.MapFile{Data: []byte("# Embedded Architect\n")},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{embeddedAgents: embeddedFS}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Added != 2 {
		t.Errorf("expected 2 added, got %d", result.Summary.Added)
	}

	// Verify files written
	got, err := os.ReadFile(filepath.Join(userChannelDir, "agents", "pythia.md"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "# Embedded Pythia\n" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestSyncUserResourceFromEmbedded_SkipsUserOwned(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	// Pre-create user file
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, []byte("# User Content\n"), 0644)

	embeddedFS := fstest.MapFS{
		"agents/agent.md": &fstest.MapFile{Data: []byte("# Embedded Content\n")},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Summary.Skipped)
	}

	// User content preserved
	got, _ := os.ReadFile(targetPath)
	if string(got) != "# User Content\n" {
		t.Errorf("user content should be preserved, got %q", got)
	}
}

func TestSyncUserResourceFromEmbedded_UpdatesKnossosOwned(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	oldContent := []byte("# Old Embedded Content\n")
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, oldContent, 0644)
	oldChecksum := checksum.Bytes(oldContent)

	embeddedFS := fstest.MapFS{
		"agents/agent.md": &fstest.MapFile{Data: []byte("# New Embedded Content\n")},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: oldChecksum,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Updated != 1 {
		t.Errorf("expected 1 updated, got %d", result.Summary.Updated)
	}

	got, _ := os.ReadFile(targetPath)
	if string(got) != "# New Embedded Content\n" {
		t.Errorf("target should be updated, got %q", got)
	}
}

func TestSyncUserResourceFromEmbedded_CollisionSkipped(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	embeddedFS := fstest.MapFS{
		"agents/pythia.md": &fstest.MapFile{Data: []byte("# Pythia\n")},
	}

	checker := &CollisionChecker{
		manifestLoaded: true,
		riteEntries:    map[string]bool{"agents/pythia.md": true},
	}
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	s := &syncer{}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Collisions != 1 {
		t.Errorf("expected 1 collision, got %d", result.Summary.Collisions)
	}
}

func TestSyncUserResourceFromEmbedded_RecreatesDeleted(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	embeddedFS := fstest.MapFS{
		"agents/agent.md": &fstest.MapFile{Data: []byte("# Content\n")},
	}

	// Manifest says knossos owns it, but file is deleted from disk
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/agent.md": {
				Owner:    provenance.OwnerKnossos,
				Scope:    provenance.ScopeUser,
				Checksum: "sha256:old",
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added (re-created), got %d", result.Summary.Added)
	}
}

func TestSyncUserResourceFromEmbedded_DryRun(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	embeddedFS := fstest.MapFS{
		"agents/agent.md": &fstest.MapFile{Data: []byte("# Content\n")},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{DryRun: true},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Added != 1 {
		t.Errorf("dry-run should report 1 added, got %d", result.Summary.Added)
	}

	// File should not exist
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Error("dry-run should not create files")
	}
}

func TestSyncUserResourceFromEmbedded_UntrackedUserFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	// Pre-create untracked file
	targetPath := filepath.Join(userChannelDir, "agents", "agent.md")
	os.WriteFile(targetPath, []byte("# User Pre-existing\n"), 0644)

	embeddedFS := fstest.MapFS{
		"agents/agent.md": &fstest.MapFile{Data: []byte("# Embedded Content\n")},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserResourceFromEmbedded(
		ResourceAgents, embeddedFS, "agents",
		userChannelDir, manifest, checker, SyncOptions{},
	)
	if err != nil {
		t.Fatalf("syncUserResourceFromEmbedded: %v", err)
	}

	if result.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped (user-created), got %d", result.Summary.Skipped)
	}

	// User content preserved
	got, _ := os.ReadFile(targetPath)
	if string(got) != "# User Pre-existing\n" {
		t.Errorf("untracked user content should be preserved, got %q", got)
	}
}

// --- Tests for syncUserResource with embedded fallback ---

func TestSyncUserResource_FallsBackToEmbedded(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	knossosHome := filepath.Join(tmpDir, "knossos")
	userChannelDir := filepath.Join(tmpDir, "user-claude")
	// knossosHome/agents does NOT exist (no filesystem source)
	os.MkdirAll(filepath.Join(userChannelDir, "agents"), 0755)

	embeddedFS := fstest.MapFS{
		"agents/pythia.md": &fstest.MapFile{Data: []byte("# Embedded\n")},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{embeddedAgents: embeddedFS}

	result, err := s.syncUserResource(ResourceAgents, knossosHome, userChannelDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserResource: %v", err)
	}

	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added from embedded fallback, got %d", result.Summary.Added)
	}
}

// manifestKeys returns all keys from a provenance manifest for debugging.
func manifestKeys(m *provenance.ProvenanceManifest) []string {
	keys := make([]string, 0, len(m.Entries))
	for k := range m.Entries {
		keys = append(keys, k)
	}
	return keys
}
