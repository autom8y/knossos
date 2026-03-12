package materialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupThroughlineIDs_RiteSwitch(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	sosDir := filepath.Join(projectDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")

	// Create templates
	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	// Setup two rites
	setupRite(t, ritesDir, "rite-alpha",
		"name: alpha-wf\n",
		[]Agent{{Name: "alpha-agent", Role: "alpha work"}})
	setupRite(t, ritesDir, "rite-beta",
		"name: beta-wf\n",
		[]Agent{{Name: "beta-agent", Role: "beta work"}})

	resolver := paths.NewResolver(projectDir)

	// Phase 1: Materialize rite-alpha
	m1 := NewMaterializer(resolver)
	m1.templatesDir = templatesDir
	_, err := m1.MaterializeWithOptions("rite-alpha", Options{})
	require.NoError(t, err)

	// Create session dirs with throughline ID files
	session1 := filepath.Join(sessionsDir, "session-20260219-100000-abcdef01")
	session2 := filepath.Join(sessionsDir, "session-20260219-110000-deadbeef")
	require.NoError(t, os.MkdirAll(session1, 0755))
	require.NoError(t, os.MkdirAll(session2, 0755))

	ids := map[string]string{"potnia": "agent-123", "moirai": "agent-456"}
	idsJSON, _ := json.Marshal(ids)
	require.NoError(t, os.WriteFile(filepath.Join(session1, ".throughline-ids.json"), idsJSON, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(session2, ".throughline-ids.json"), idsJSON, 0644))

	// Phase 2: Switch to rite-beta via Sync (which calls syncRiteScope)
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	syncResult, err := m2.Sync(SyncOptions{
		Scope:    ScopeRite,
		RiteName: "rite-beta",
	})
	require.NoError(t, err)
	require.NotNil(t, syncResult.RiteResult)

	// Verify rite-switch was detected
	assert.True(t, syncResult.RiteResult.RiteSwitched)
	assert.Equal(t, "rite-alpha", syncResult.RiteResult.PreviousRite)
	assert.Equal(t, 2, syncResult.RiteResult.ThroughlineIDsCleaned)

	// Verify files are actually gone
	_, err = os.Stat(filepath.Join(session1, ".throughline-ids.json"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(session2, ".throughline-ids.json"))
	assert.True(t, os.IsNotExist(err))
}

func TestCleanupThroughlineIDs_SameRiteNoCleanup(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	sosDir := filepath.Join(projectDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")

	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	setupRite(t, ritesDir, "same-rite",
		"name: same-wf\n",
		[]Agent{{Name: "worker", Role: "works"}})

	resolver := paths.NewResolver(projectDir)

	// Materialize the rite
	m1 := NewMaterializer(resolver)
	m1.templatesDir = templatesDir
	_, err := m1.MaterializeWithOptions("same-rite", Options{})
	require.NoError(t, err)

	// Create session with throughline IDs
	sessionDir := filepath.Join(sessionsDir, "session-20260219-120000-aabbccdd")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, ".throughline-ids.json"),
		[]byte(`{"potnia":"agent-789"}`), 0644))

	// Re-sync the SAME rite
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	syncResult, err := m2.Sync(SyncOptions{
		Scope:    ScopeRite,
		RiteName: "same-rite",
	})
	require.NoError(t, err)
	require.NotNil(t, syncResult.RiteResult)

	// No rite switch should be detected
	assert.False(t, syncResult.RiteResult.RiteSwitched)
	assert.Empty(t, syncResult.RiteResult.PreviousRite)
	assert.Equal(t, 0, syncResult.RiteResult.ThroughlineIDsCleaned)

	// Throughline IDs file should still exist
	_, err = os.Stat(filepath.Join(sessionDir, ".throughline-ids.json"))
	assert.NoError(t, err, "throughline IDs should NOT be cleaned on same-rite resync")
}

func TestCleanupThroughlineIDs_NoSessionDirs(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")

	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	setupRite(t, ritesDir, "first-rite",
		"name: first-wf\n",
		[]Agent{{Name: "agent-a", Role: "works"}})
	setupRite(t, ritesDir, "second-rite",
		"name: second-wf\n",
		[]Agent{{Name: "agent-b", Role: "works"}})

	resolver := paths.NewResolver(projectDir)

	// Materialize first rite
	m1 := NewMaterializer(resolver)
	m1.templatesDir = templatesDir
	_, err := m1.MaterializeWithOptions("first-rite", Options{})
	require.NoError(t, err)

	// Switch rites WITHOUT any session dirs existing
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	syncResult, err := m2.Sync(SyncOptions{
		Scope:    ScopeRite,
		RiteName: "second-rite",
	})
	require.NoError(t, err)
	require.NotNil(t, syncResult.RiteResult)

	// Rite switch detected, but zero files cleaned (no sessions)
	assert.True(t, syncResult.RiteResult.RiteSwitched)
	assert.Equal(t, "first-rite", syncResult.RiteResult.PreviousRite)
	assert.Equal(t, 0, syncResult.RiteResult.ThroughlineIDsCleaned)
}

func TestCleanupThroughlineIDs_SkipsNonSessionDirs(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	sosDir := filepath.Join(projectDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")

	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	setupRite(t, ritesDir, "rite-x",
		"name: x-wf\n",
		[]Agent{{Name: "x-agent", Role: "x work"}})
	setupRite(t, ritesDir, "rite-y",
		"name: y-wf\n",
		[]Agent{{Name: "y-agent", Role: "y work"}})

	resolver := paths.NewResolver(projectDir)

	// Materialize rite-x
	m1 := NewMaterializer(resolver)
	m1.templatesDir = templatesDir
	_, err := m1.MaterializeWithOptions("rite-x", Options{})
	require.NoError(t, err)

	// Create non-session dirs that should be skipped
	for _, name := range []string{".locks", ".harness-map", ".audit", ".current-session"} {
		dir := filepath.Join(sessionsDir, name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		// Plant a decoy file that should NOT be removed
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, ".throughline-ids.json"),
			[]byte(`{"decoy":"should-survive"}`), 0644))
	}

	// Create one real session dir
	realSession := filepath.Join(sessionsDir, "session-20260219-130000-11223344")
	require.NoError(t, os.MkdirAll(realSession, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(realSession, ".throughline-ids.json"),
		[]byte(`{"potnia":"agent-abc"}`), 0644))

	// Switch to rite-y
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	syncResult, err := m2.Sync(SyncOptions{
		Scope:    ScopeRite,
		RiteName: "rite-y",
	})
	require.NoError(t, err)
	require.NotNil(t, syncResult.RiteResult)

	// Only the real session's file should have been cleaned
	assert.Equal(t, 1, syncResult.RiteResult.ThroughlineIDsCleaned)

	// Decoy files in non-session dirs should survive
	for _, name := range []string{".locks", ".harness-map", ".audit", ".current-session"} {
		_, err := os.Stat(filepath.Join(sessionsDir, name, ".throughline-ids.json"))
		assert.NoError(t, err, "decoy in %s should survive cleanup", name)
	}

	// Real session file should be gone
	_, err = os.Stat(filepath.Join(realSession, ".throughline-ids.json"))
	assert.True(t, os.IsNotExist(err))
}

func TestCleanupThroughlineIDs_DryRunSkipsCleanup(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	sosDir := filepath.Join(projectDir, ".sos")
	sessionsDir := filepath.Join(sosDir, "sessions")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")

	templatesDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	setupRite(t, ritesDir, "dry-a",
		"name: dry-a-wf\n",
		[]Agent{{Name: "agent-dry-a", Role: "works"}})
	setupRite(t, ritesDir, "dry-b",
		"name: dry-b-wf\n",
		[]Agent{{Name: "agent-dry-b", Role: "works"}})

	resolver := paths.NewResolver(projectDir)

	// Materialize dry-a
	m1 := NewMaterializer(resolver)
	m1.templatesDir = templatesDir
	_, err := m1.MaterializeWithOptions("dry-a", Options{})
	require.NoError(t, err)

	// Create session with throughline IDs
	sessionDir := filepath.Join(sessionsDir, "session-20260219-140000-dryrun01")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, ".throughline-ids.json"),
		[]byte(`{"potnia":"agent-dry"}`), 0644))

	// Switch with DryRun
	m2 := NewMaterializer(resolver)
	m2.templatesDir = templatesDir
	syncResult, err := m2.Sync(SyncOptions{
		Scope:    ScopeRite,
		RiteName: "dry-b",
		DryRun:   true,
	})
	require.NoError(t, err)
	require.NotNil(t, syncResult.RiteResult)

	// DryRun should skip cleanup entirely
	assert.False(t, syncResult.RiteResult.RiteSwitched)
	assert.Equal(t, 0, syncResult.RiteResult.ThroughlineIDsCleaned)

	// File should still exist
	_, err = os.Stat(filepath.Join(sessionDir, ".throughline-ids.json"))
	assert.NoError(t, err, "throughline IDs should survive dry run")
}
