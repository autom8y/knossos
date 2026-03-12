package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaterializeAgents_PreservesUserAgents verifies that user-created agents
// in channel agents/ survive materialization (selective write, not destructive nuke).
func TestMaterializeAgents_PreservesUserAgents(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	channelDir := filepath.Join(projectDir, paths.ClaudeChannel{}.DirName())

	// Setup a rite with one agent
	agents := []Agent{{Name: "designer", Role: "designs things"}}
	setupRite(t, ritesDir, "test-rite", "", agents)

	// Pre-create a user agent that is NOT in the rite manifest
	agentsDir := filepath.Join(channelDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	userAgent := filepath.Join(agentsDir, "my-custom-agent.md")
	require.NoError(t, os.WriteFile(userAgent, []byte("# My Custom Agent\n"), 0644))

	// Materialize
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	// Verify rite agent was written
	riteAgent := filepath.Join(agentsDir, "designer.md")
	assert.FileExists(t, riteAgent, "rite agent should be materialized")

	// Verify user agent survived
	assert.FileExists(t, userAgent, "user-created agent should survive materialization")
	content, err := os.ReadFile(userAgent)
	require.NoError(t, err)
	assert.Equal(t, "# My Custom Agent\n", string(content))
}

// TestMaterializeAgents_KeepAllPreservesOrphans verifies that orphan agents
// from a previous rite survive with KeepAll (the default) after rite switch.
func TestMaterializeAgents_KeepAllPreservesOrphans(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	channelDir := filepath.Join(projectDir, paths.ClaudeChannel{}.DirName())

	// Setup rite-a with agent "designer"
	setupRite(t, ritesDir, "rite-a", "", []Agent{{Name: "designer", Role: "designs"}})
	// Setup rite-b with agent "deployer"
	setupRite(t, ritesDir, "rite-b", "", []Agent{{Name: "deployer", Role: "deploys"}})

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// Materialize rite-a
	_, err := m.MaterializeWithOptions("rite-a", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	designerPath := filepath.Join(channelDir, "agents", "designer.md")
	assert.FileExists(t, designerPath, "designer should exist after rite-a materialization")

	// Switch to rite-b with KeepAll
	result, err := m.MaterializeWithOptions("rite-b", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	// designer.md should be detected as orphan
	assert.Contains(t, result.OrphansDetected, "designer.md", "designer should be detected as orphan")

	// With KeepAll, designer should survive (no longer destroyed by os.RemoveAll)
	assert.FileExists(t, designerPath, "orphan agent should survive with KeepAll")

	// deployer should be written
	deployerPath := filepath.Join(channelDir, "agents", "deployer.md")
	assert.FileExists(t, deployerPath, "deployer should exist after rite-b materialization")
}

// TestSyncMena_PreservesUserCommands verifies that user-created commands
// in channel commands/ survive destructive mode projection.
func TestSyncMena_PreservesUserCommands(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create mena source with one command
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "my-cmd")
	require.NoError(t, os.MkdirAll(droDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(droDir, "INDEX.dro.md"),
		[]byte("# My Command\n"), 0644))

	// Create target dirs with a user-created command
	commandsDir := filepath.Join(tmpDir, "commands")
	userCmdDir := filepath.Join(commandsDir, "user-custom")
	require.NoError(t, os.MkdirAll(userCmdDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(userCmdDir, "INDEX.md"),
		[]byte("# User Custom Command\n"), 0644))

	skillsDir := filepath.Join(tmpDir, "skills")

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := SyncMena(sources, opts)
	require.NoError(t, err)

	// Verify projected command exists (promoted to parent level)
	assert.FileExists(t,
		filepath.Join(commandsDir, "my-cmd.md"),
		"projected command should exist (promoted)")

	// Verify user-created command survived
	assert.FileExists(t,
		filepath.Join(userCmdDir, "INDEX.md"),
		"user-created command should survive destructive projection")
	content, err := os.ReadFile(filepath.Join(userCmdDir, "INDEX.md"))
	require.NoError(t, err)
	assert.Equal(t, "# User Custom Command\n", string(content))
}

// TestSyncMena_PreservesUserSkills verifies that user-created skills
// in channel skills/ survive destructive mode projection.
func TestSyncMena_PreservesUserSkills(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create mena source with one legomena
	menaDir := filepath.Join(tmpDir, "mena")
	legoDir := filepath.Join(menaDir, "my-ref")
	require.NoError(t, os.MkdirAll(legoDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(legoDir, "INDEX.lego.md"),
		[]byte("# My Ref\n"), 0644))

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create a user skill
	userSkillDir := filepath.Join(skillsDir, "user-ref")
	require.NoError(t, os.MkdirAll(userSkillDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(userSkillDir, "INDEX.md"),
		[]byte("# User Ref\n"), 0644))

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := SyncMena(sources, opts)
	require.NoError(t, err)

	// Verify projected skill exists (CC entrypoint is SKILL.md, not INDEX.md)
	assert.FileExists(t,
		filepath.Join(skillsDir, "my-ref", "SKILL.md"),
		"projected skill should exist")

	// Verify user-created skill survived
	assert.FileExists(t,
		filepath.Join(userSkillDir, "INDEX.md"),
		"user-created skill should survive destructive projection")
}

// TestSyncMena_CleansStaleCompanionFiles verifies that when a managed entry
// is re-projected with fewer files, stale companion files are cleaned.
func TestSyncMena_CleansStaleCompanionFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, "commands")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Pre-create a managed entry with a companion file
	entryDir := filepath.Join(commandsDir, "my-cmd")
	require.NoError(t, os.MkdirAll(entryDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(entryDir, "INDEX.md"),
		[]byte("# Old Index\n"), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(entryDir, "companion.md"),
		[]byte("# Old Companion\n"), 0644))

	// Create mena source that only has INDEX (no companion)
	menaDir := filepath.Join(tmpDir, "mena")
	droDir := filepath.Join(menaDir, "my-cmd")
	require.NoError(t, os.MkdirAll(droDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(droDir, "INDEX.dro.md"),
		[]byte("# New Index\n"), 0644))

	sources := []MenaSource{{Path: menaDir}}
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
	}

	_, err := SyncMena(sources, opts)
	require.NoError(t, err)

	// New INDEX should be promoted to parent level (commands/my-cmd.md)
	promotedPath := filepath.Join(commandsDir, "my-cmd.md")
	content, err := os.ReadFile(promotedPath)
	require.NoError(t, err)
	assert.Equal(t, "# New Index\n", string(content))

	// Old INDEX.md in subdirectory should be cleaned up
	assert.NoFileExists(t,
		filepath.Join(entryDir, "INDEX.md"),
		"old INDEX.md should be cleaned from subdirectory after promotion")

	// Old companion should be gone (entry subdir was cleaned before rewrite)
	assert.NoFileExists(t,
		filepath.Join(entryDir, "companion.md"),
		"stale companion file should be cleaned on re-projection")
}

// TestMaterialize_NoSkipGuard_AlwaysRuns verifies that MaterializeWithOptions
// runs the full pipeline even when already on the same rite (skip guard removed).
func TestMaterialize_NoSkipGuard_AlwaysRuns(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	channelDir := filepath.Join(projectDir, paths.ClaudeChannel{}.DirName())

	agents := []Agent{{Name: "tester", Role: "tests things"}}
	setupRite(t, ritesDir, "test-rite", "", agents)

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// First materialization
	result, err := m.MaterializeWithOptions("test-rite", Options{KeepAll: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status, "first materialization should succeed")

	agentPath := filepath.Join(channelDir, "agents", "tester.md")
	assert.FileExists(t, agentPath)

	// Modify the source agent file
	riteAgentPath := filepath.Join(ritesDir, "test-rite", "agents", "tester.md")
	require.NoError(t, os.WriteFile(riteAgentPath, []byte("# Updated Tester\n"), 0644))

	// Second materialization WITHOUT --force — should still run (no skip guard)
	result, err = m.MaterializeWithOptions("test-rite", Options{KeepAll: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status, "second materialization should succeed (no skip)")

	// Verify updated content was picked up
	content, err := os.ReadFile(agentPath)
	require.NoError(t, err)
	assert.Equal(t, "# Updated Tester\n", string(content),
		"modified source should be picked up without --force")
}
