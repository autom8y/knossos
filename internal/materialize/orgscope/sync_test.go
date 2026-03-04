package orgscope

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncOrgScope_NoOrg(t *testing.T) {
	t.Setenv("KNOSSOS_ORG", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // Prevent ActiveOrg() from reading active-org file
	result, err := SyncOrgScope(SyncOrgScopeParams{})
	if err != nil {
		t.Fatalf("SyncOrgScope failed: %v", err)
	}
	if result.Status != "skipped" {
		t.Errorf("Expected status 'skipped', got %q", result.Status)
	}
}

func TestSyncOrgScope_OrgDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	result, err := SyncOrgScope(SyncOrgScopeParams{
		OrgName: "nonexistent-org",
		OrgDir:  filepath.Join(tmpDir, "no-such-dir"),
	})
	if err != nil {
		t.Fatalf("SyncOrgScope failed: %v", err)
	}
	if result.Status != "skipped" {
		t.Errorf("Expected status 'skipped', got %q", result.Status)
	}
	if result.OrgName != "nonexistent-org" {
		t.Errorf("Expected org name 'nonexistent-org', got %q", result.OrgName)
	}
}

func TestSyncOrgScope_SyncsAgents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create org directory with an agent
	orgDir := filepath.Join(tmpDir, "org-data")
	agentsDir := filepath.Join(orgDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), []byte("# Test Agent\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a fake user .claude dir
	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	if err := os.MkdirAll(userClaudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := SyncOrgScope(SyncOrgScopeParams{
		OrgName:       "test-org",
		OrgDir:        orgDir,
		UserClaudeDir: userClaudeDir,
	})
	if err != nil {
		t.Fatalf("SyncOrgScope failed: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got %q (error: %s)", result.Status, result.Error)
	}
	if result.Agents != 1 {
		t.Errorf("Expected 1 agent synced, got %d", result.Agents)
	}

	// Verify the agent was actually written
	targetAgent := filepath.Join(userClaudeDir, "agents", "test-agent.md")
	if _, err := os.Stat(targetAgent); os.IsNotExist(err) {
		t.Error("Expected agent file to exist at target")
	}

	// Verify provenance manifest was created
	manifestPath := filepath.Join(userClaudeDir, "ORG_PROVENANCE_MANIFEST.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Expected ORG_PROVENANCE_MANIFEST.yaml to be created")
	}
}

func TestSyncOrgScope_SyncsMena(t *testing.T) {
	tmpDir := t.TempDir()

	// Create org directory with mena
	orgDir := filepath.Join(tmpDir, "org-data")
	menaDir := filepath.Join(orgDir, "mena")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(menaDir, "standards.lego.md"), []byte("# Standards\n"), 0644); err != nil {
		t.Fatal(err)
	}

	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	if err := os.MkdirAll(userClaudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := SyncOrgScope(SyncOrgScopeParams{
		OrgName:       "test-org",
		OrgDir:        orgDir,
		UserClaudeDir: userClaudeDir,
	})
	if err != nil {
		t.Fatalf("SyncOrgScope failed: %v", err)
	}
	if result.Mena != 1 {
		t.Errorf("Expected 1 mena synced, got %d", result.Mena)
	}

	// Verify the mena was written to skills/
	targetMena := filepath.Join(userClaudeDir, "skills", "standards.lego.md")
	if _, err := os.Stat(targetMena); os.IsNotExist(err) {
		t.Error("Expected mena file to exist at target skills/")
	}
}

func TestSyncOrgScope_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create org with an agent
	orgDir := filepath.Join(tmpDir, "org-data")
	agentsDir := filepath.Join(orgDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "agent.md"), []byte("# Agent\n"), 0644); err != nil {
		t.Fatal(err)
	}

	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	if err := os.MkdirAll(userClaudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := SyncOrgScope(SyncOrgScopeParams{
		OrgName:       "dry-org",
		OrgDir:        orgDir,
		UserClaudeDir: userClaudeDir,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("SyncOrgScope failed: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got %q", result.Status)
	}
	if result.Agents != 1 {
		t.Errorf("Expected 1 agent in dry run, got %d", result.Agents)
	}

	// Verify target was NOT created (dry run)
	targetAgent := filepath.Join(userClaudeDir, "agents", "agent.md")
	if _, err := os.Stat(targetAgent); err == nil {
		t.Error("Expected agent file to NOT exist in dry-run mode")
	}
}

func TestSyncOrgScope_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create org with an agent
	orgDir := filepath.Join(tmpDir, "org-data")
	agentsDir := filepath.Join(orgDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "agent.md"), []byte("# Agent\n"), 0644); err != nil {
		t.Fatal(err)
	}

	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	if err := os.MkdirAll(userClaudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	params := SyncOrgScopeParams{
		OrgName:       "test-org",
		OrgDir:        orgDir,
		UserClaudeDir: userClaudeDir,
	}

	// First sync
	result1, err := SyncOrgScope(params)
	if err != nil {
		t.Fatalf("First sync failed: %v", err)
	}
	if result1.Agents != 1 {
		t.Errorf("First sync: expected 1 agent, got %d", result1.Agents)
	}

	// Second sync (should be idempotent — 0 changes since checksum matches)
	result2, err := SyncOrgScope(params)
	if err != nil {
		t.Fatalf("Second sync failed: %v", err)
	}
	if result2.Agents != 0 {
		t.Errorf("Second sync: expected 0 agents (unchanged), got %d", result2.Agents)
	}
}
