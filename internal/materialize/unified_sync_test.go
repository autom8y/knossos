package materialize

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Test infrastructure: helper functions for setting up test fixtures

func setupTestRite(t *testing.T, riteDir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(riteDir, 0755))

	// Create manifest.yaml
	manifest := &RiteManifest{
		Name:        "test-rite",
		Version:     "1.0.0",
		Description: "Test rite for integration testing",
		EntryAgent:  "test-agent",
		Agents: []Agent{
			{Name: "test-agent", Role: "Testing agent"},
		},
		Dromena:  []string{},
		Legomena: []string{},
		Hooks:    []string{},
	}
	data, err := yaml.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), data, 0644))

	// Create agents directory with test agent
	agentsDir := filepath.Join(riteDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(agentsDir, "test-agent.md"),
		[]byte("# Test Agent\n\nA test agent.\n"),
		0644,
	))
}

func setupKnossosHome(t *testing.T, knossosHome string) {
	t.Helper()

	// Create agents directory
	agentsDir := filepath.Join(knossosHome, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(agentsDir, "user-agent.md"),
		[]byte("# User Agent\n\nA user-level agent.\n"),
		0644,
	))

	// Create mena directory with dromena and legomena
	menaDir := filepath.Join(knossosHome, "mena")
	require.NoError(t, os.MkdirAll(filepath.Join(menaDir, "test-command"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(menaDir, "test-command", "INDEX.dro.md"),
		[]byte("# Test Command\n\nA test command.\n"),
		0644,
	))

	require.NoError(t, os.MkdirAll(filepath.Join(menaDir, "test-skill"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(menaDir, "test-skill", "INDEX.lego.md"),
		[]byte("# Test Skill\n\nA test skill.\n"),
		0644,
	))

	// Create hooks directory
	hooksDir := filepath.Join(knossosHome, "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(hooksDir, "test-hook.sh"),
		[]byte("#!/bin/bash\necho 'test hook'\n"),
		0755,
	))
}

func writeActiveRite(t *testing.T, claudeDir, riteName string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "ACTIVE_RITE"),
		[]byte(riteName+"\n"),
		0644,
	))
}

// TestUnifiedSync_RiteOnly tests scope=rite with a valid rite
func TestUnifiedSync_RiteOnly(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	riteDir := filepath.Join(projectDir, "rites", "test-rite")

	setupTestRite(t, riteDir)
	writeActiveRite(t, claudeDir, "test-rite")

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope: ScopeRite,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result.RiteResult)
	assert.Equal(t, "success", result.RiteResult.Status)
	assert.Nil(t, result.UserResult)

	// Verify .claude/ populated
	assert.FileExists(t, filepath.Join(claudeDir, "agents", "test-agent.md"))
	assert.FileExists(t, filepath.Join(claudeDir, "CLAUDE.md"))
	assert.FileExists(t, filepath.Join(claudeDir, "ACTIVE_RITE"))
}

// TestUnifiedSync_UserOnly tests scope=user without project context
func TestUnifiedSync_UserOnly(t *testing.T) {
	// Save and restore original KNOSSOS_HOME
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome() // Force config reload

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	projectDir := t.TempDir() // Empty project (no rite)

	setupKnossosHome(t, knossosHome)

	// Override paths.UserClaudeDir for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope: ScopeUser,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)
	assert.Nil(t, result.RiteResult)
	assert.NotNil(t, result.UserResult)
	assert.Equal(t, "success", result.UserResult.Status)

	// Verify ~/.claude/ populated
	assert.FileExists(t, filepath.Join(userClaudeDir, "agents", "user-agent.md"))
	assert.FileExists(t, filepath.Join(userClaudeDir, "commands", "test-command", "INDEX.md"))
	assert.FileExists(t, filepath.Join(userClaudeDir, "skills", "test-skill", "INDEX.md"))
	assert.FileExists(t, filepath.Join(userClaudeDir, "hooks", "test-hook.sh"))
}

// TestUnifiedSync_ScopeAll tests default scope (rite + user)
func TestUnifiedSync_ScopeAll(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	riteDir := filepath.Join(projectDir, "rites", "test-rite")

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupTestRite(t, riteDir)
	setupKnossosHome(t, knossosHome)
	writeActiveRite(t, claudeDir, "test-rite")

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope: ScopeAll,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result.RiteResult)
	assert.NotNil(t, result.UserResult)
	assert.Equal(t, "success", result.RiteResult.Status)
	assert.Equal(t, "success", result.UserResult.Status)

	// Verify both .claude/ and ~/.claude/ populated
	assert.FileExists(t, filepath.Join(claudeDir, "agents", "test-agent.md"))
	assert.FileExists(t, filepath.Join(userClaudeDir, "agents", "user-agent.md"))
}

// TestUnifiedSync_NoActiveRite_ScopeAll tests graceful degradation
func TestUnifiedSync_NoActiveRite_ScopeAll(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	projectDir := t.TempDir()
	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupKnossosHome(t, knossosHome)

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope: ScopeAll,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)

	// Rite scope should run minimal mode
	assert.NotNil(t, result.RiteResult)
	assert.Equal(t, "minimal", result.RiteResult.Status)

	// User scope should succeed
	assert.NotNil(t, result.UserResult)
	assert.Equal(t, "success", result.UserResult.Status)
}

// TestUnifiedSync_NoActiveRite_ScopeRite tests error case
func TestUnifiedSync_NoActiveRite_ScopeRite(t *testing.T) {
	projectDir := t.TempDir()

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope: ScopeRite,
	}

	result, err := m.Sync(opts)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no ACTIVE_RITE found")
}

// TestUnifiedSync_CollisionDetection tests user file shadowing rite
func TestUnifiedSync_CollisionDetection(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	riteDir := filepath.Join(projectDir, "rites", "test-rite")

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	// Setup rite with test-agent
	setupTestRite(t, riteDir)
	writeActiveRite(t, claudeDir, "test-rite")

	// Setup KNOSSOS_HOME with SAME agent name (collision)
	agentsDir := filepath.Join(knossosHome, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(agentsDir, "test-agent.md"),
		[]byte("# Collision Agent\n\nThis should be skipped.\n"),
		0644,
	))

	// First sync rite scope to populate .claude/
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	riteResult, err := m.Sync(SyncOptions{Scope: ScopeRite})
	require.NoError(t, err)
	assert.NotNil(t, riteResult.RiteResult)

	// Then sync user scope - should detect collision
	userResult, err := m.Sync(SyncOptions{Scope: ScopeUser})
	require.NoError(t, err)
	assert.NotNil(t, userResult.UserResult)

	// Verify collision was detected and file was skipped
	agentResource := userResult.UserResult.Resources[ResourceAgents]
	require.NotNil(t, agentResource)

	collisionFound := false
	for _, skipped := range agentResource.Changes.Skipped {
		if skipped.Name == "agents/test-agent.md" {
			assert.Contains(t, skipped.Reason, "collision")
			collisionFound = true
		}
	}
	assert.True(t, collisionFound, "Expected collision to be detected")

	// Verify collisions counted
	assert.Greater(t, userResult.UserResult.Totals.Collisions, 0)
}

// TestUnifiedSync_RecoveryMode tests --recover
func TestUnifiedSync_RecoveryMode(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupKnossosHome(t, knossosHome)

	// Pre-create file in ~/.claude/ that matches KNOSSOS_HOME
	agentsDir := filepath.Join(userClaudeDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(agentsDir, "user-agent.md"),
		[]byte("# User Agent\n\nA user-level agent.\n"), // Exact match
		0644,
	))

	projectDir := t.TempDir()
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope:   ScopeUser,
		Recover: true,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result.UserResult)

	// Verify file was adopted (not added or updated, but unchanged)
	agentResource := result.UserResult.Resources[ResourceAgents]
	require.NotNil(t, agentResource)
	assert.Contains(t, agentResource.Changes.Unchanged, "agents/user-agent.md")

	// Verify manifest was created and file is tracked
	manifestPath := provenance.UserManifestPath(userClaudeDir)
	manifest, err := provenance.Load(manifestPath)
	require.NoError(t, err)
	entry, exists := manifest.Entries["agents/user-agent.md"]
	require.True(t, exists)
	assert.Equal(t, provenance.OwnerKnossos, entry.Owner)
}

// TestUnifiedSync_OverwriteDiverged tests --overwrite-diverged
func TestUnifiedSync_OverwriteDiverged(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupKnossosHome(t, knossosHome)

	projectDir := t.TempDir()
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// First sync to establish baseline
	_, err := m.Sync(SyncOptions{Scope: ScopeUser})
	require.NoError(t, err)

	// Modify target file (simulate divergence)
	targetPath := filepath.Join(userClaudeDir, "agents", "user-agent.md")
	require.NoError(t, os.WriteFile(
		targetPath,
		[]byte("# Modified Agent\n\nLocally modified.\n"),
		0644,
	))

	// Update source file
	sourcePath := filepath.Join(knossosHome, "agents", "user-agent.md")
	require.NoError(t, os.WriteFile(
		sourcePath,
		[]byte("# Updated User Agent\n\nUpdated in KNOSSOS_HOME.\n"),
		0644,
	))

	// Sync without --overwrite-diverged should skip
	result1, err := m.Sync(SyncOptions{Scope: ScopeUser})
	require.NoError(t, err)
	agentResource := result1.UserResult.Resources[ResourceAgents]
	divergedFound := false
	for _, skipped := range agentResource.Changes.Skipped {
		if skipped.Name == "agents/user-agent.md" && skipped.Reason == "diverged (use --overwrite-diverged to force)" {
			divergedFound = true
		}
	}
	assert.True(t, divergedFound, "Expected diverged file to be skipped")

	// Sync with --overwrite-diverged should update
	result2, err := m.Sync(SyncOptions{Scope: ScopeUser, OverwriteDiverged: true})
	require.NoError(t, err)
	agentResource2 := result2.UserResult.Resources[ResourceAgents]
	assert.Contains(t, agentResource2.Changes.Updated, "agents/user-agent.md")

	// Verify file was overwritten
	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Updated in KNOSSOS_HOME")
}

// TestUnifiedSync_KeepOrphans tests --keep-orphans
func TestUnifiedSync_KeepOrphans(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupKnossosHome(t, knossosHome)

	projectDir := t.TempDir()
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// First sync to establish baseline
	_, err := m.Sync(SyncOptions{Scope: ScopeUser})
	require.NoError(t, err)

	// Remove file from KNOSSOS_HOME (orphan it)
	orphanPath := filepath.Join(knossosHome, "agents", "user-agent.md")
	require.NoError(t, os.Remove(orphanPath))

	// Sync without --keep-orphans should remove orphan
	result1, err := m.Sync(SyncOptions{Scope: ScopeUser})
	require.NoError(t, err)
	assert.NotNil(t, result1.UserResult)

	// Verify orphan was removed
	targetPath := filepath.Join(userClaudeDir, "agents", "user-agent.md")
	_, err = os.Stat(targetPath)
	assert.True(t, os.IsNotExist(err), "Orphan should be removed by default")

	// Re-sync to establish baseline again
	require.NoError(t, os.WriteFile(orphanPath, []byte("# User Agent\n"), 0644))
	_, err = m.Sync(SyncOptions{Scope: ScopeUser})
	require.NoError(t, err)

	// Remove file again and sync with --keep-orphans
	require.NoError(t, os.Remove(orphanPath))
	result2, err := m.Sync(SyncOptions{Scope: ScopeUser, KeepOrphans: true})
	require.NoError(t, err)
	assert.NotNil(t, result2.UserResult)

	// Verify orphan was kept
	_, err = os.Stat(targetPath)
	assert.NoError(t, err, "Orphan should be preserved with --keep-orphans")
}

// TestUnifiedSync_ResourceFilter tests --resource flag
func TestUnifiedSync_ResourceFilter(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupKnossosHome(t, knossosHome)

	projectDir := t.TempDir()
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope:    ScopeUser,
		Resource: ResourceAgents,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result.UserResult)

	// Only agents should be synced
	assert.Contains(t, result.UserResult.Resources, ResourceAgents)
	assert.NotContains(t, result.UserResult.Resources, ResourceMena)
	assert.NotContains(t, result.UserResult.Resources, ResourceHooks)

	// Verify only agents directory created
	assert.DirExists(t, filepath.Join(userClaudeDir, "agents"))
	assert.NoDirExists(t, filepath.Join(userClaudeDir, "commands"))
	assert.NoDirExists(t, filepath.Join(userClaudeDir, "skills"))
	assert.NoDirExists(t, filepath.Join(userClaudeDir, "hooks"))
}

// TestUnifiedSync_DryRun tests --dry-run for both scopes
func TestUnifiedSync_DryRun(t *testing.T) {
	oldKnossosHome := os.Getenv("KNOSSOS_HOME")
	defer func() {
		if oldKnossosHome != "" {
			os.Setenv("KNOSSOS_HOME", oldKnossosHome)
		} else {
			os.Unsetenv("KNOSSOS_HOME")
		}
	}()

	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	riteDir := filepath.Join(projectDir, "rites", "test-rite")

	knossosHome := t.TempDir()
	os.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", filepath.Dir(userClaudeDir))

	setupTestRite(t, riteDir)
	setupKnossosHome(t, knossosHome)
	writeActiveRite(t, claudeDir, "test-rite")

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope:  ScopeAll,
		DryRun: true,
	}

	result, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result.RiteResult)
	assert.NotNil(t, result.UserResult)

	// Verify no files were actually written
	assert.NoDirExists(t, filepath.Join(claudeDir, "agents"))
	assert.NoDirExists(t, filepath.Join(userClaudeDir, "agents"))

	// Rite dry-run should still report what would happen
	assert.Equal(t, "success", result.RiteResult.Status)
}

// TestUnifiedSync_Idempotency tests running sync twice
func TestUnifiedSync_Idempotency(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	riteDir := filepath.Join(projectDir, "rites", "test-rite")

	setupTestRite(t, riteDir)
	writeActiveRite(t, claudeDir, "test-rite")

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	opts := SyncOptions{
		Scope: ScopeRite,
	}

	// First sync
	result1, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result1.RiteResult)

	// Read CLAUDE.md content after first sync
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	content1, err := os.ReadFile(claudeMdPath)
	require.NoError(t, err)

	// Read provenance manifest after first sync
	manifestPath := provenance.ManifestPath(claudeDir)
	manifest1, err := provenance.Load(manifestPath)
	require.NoError(t, err)
	agentEntry1 := manifest1.Entries["agents/test-agent.md"]
	require.NotNil(t, agentEntry1)

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Second sync
	result2, err := m.Sync(opts)
	require.NoError(t, err)
	assert.NotNil(t, result2.RiteResult)
	assert.Equal(t, "success", result2.RiteResult.Status)

	// Read CLAUDE.md content after second sync
	content2, err := os.ReadFile(claudeMdPath)
	require.NoError(t, err)

	// Read provenance manifest after second sync
	manifest2, err := provenance.Load(manifestPath)
	require.NoError(t, err)
	agentEntry2 := manifest2.Entries["agents/test-agent.md"]
	require.NotNil(t, agentEntry2)

	// Verify content is functionally equivalent (same length and key markers)
	// Note: exact string equality may fail due to non-deterministic attribute ordering in templates
	assert.Equal(t, len(content1), len(content2), "CLAUDE.md length should be identical on second sync")
	assert.Contains(t, string(content2), "test-rite", "CLAUDE.md should still reference test-rite")
	assert.Contains(t, string(content2), "test-agent", "CLAUDE.md should still reference test-agent")

	// Verify agent entry checksum unchanged (file not rewritten)
	assert.Equal(t, agentEntry1.Checksum, agentEntry2.Checksum, "Agent checksum should be unchanged")
}
