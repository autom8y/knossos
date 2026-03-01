package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvenanceIntegration_BasicMaterialization verifies that PROVENANCE_MANIFEST.yaml
// is created with expected entries after a basic materialization.
func TestProvenanceIntegration_BasicMaterialization(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup templates (required for CLAUDE.md)
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templatesDir, "rules", "test-rule.md"),
		[]byte("test rule content"), 0644))

	// Setup test rite
	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Materialize
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	result, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Verify PROVENANCE_MANIFEST.yaml was created
	manifestPath := filepath.Join(claudeDir, "PROVENANCE_MANIFEST.yaml")
	require.FileExists(t, manifestPath)

	// Load and validate manifest
	manifest, err := provenance.Load(manifestPath)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	// Check schema version
	assert.Equal(t, provenance.CurrentSchemaVersion, manifest.SchemaVersion)

	// Check active rite
	assert.Equal(t, "test-rite", manifest.ActiveRite)

	// Check entries exist for expected files
	assert.NotNil(t, manifest.Entries["agents/builder.md"], "agent should be recorded")
	assert.NotNil(t, manifest.Entries["rules/test-rule.md"], "rule should be recorded")
	assert.NotNil(t, manifest.Entries["CLAUDE.md"], "CLAUDE.md should be recorded")
	assert.NotNil(t, manifest.Entries["settings.local.json"], "settings should be recorded")
	assert.NotNil(t, manifest.Entries["ACTIVE_WORKFLOW.yaml"], "workflow should be recorded")

	// ACTIVE_RITE, sync/state.json, and KNOSSOS_MANIFEST.yaml are NOT tracked in provenance
	// because they always change on every sync (timestamps, version counters) and are always
	// knossos-owned with no ownership ambiguity
	assert.Nil(t, manifest.Entries["ACTIVE_RITE"], "ACTIVE_RITE should NOT be tracked (always regenerated)")
	assert.Nil(t, manifest.Entries["sync/state.json"], "sync/state.json should NOT be tracked (always regenerated)")
	assert.Nil(t, manifest.Entries["KNOSSOS_MANIFEST.yaml"], "KNOSSOS_MANIFEST.yaml should NOT be tracked (always regenerated)")

	// Verify all recorded entries have knossos ownership and rite scope
	for path, entry := range manifest.Entries {
		assert.Equal(t, provenance.OwnerKnossos, entry.Owner, "entry %s should be knossos-owned", path)
		assert.Equal(t, provenance.ScopeRite, entry.Scope, "entry %s should have rite scope", path)
		assert.NotEmpty(t, entry.Checksum, "entry %s should have checksum", path)
		assert.NotZero(t, entry.LastSynced, "entry %s should have last_synced", path)
	}
}

// TestProvenanceIntegration_DivergenceDetection verifies that divergence detection
// promotes modified knossos files to user ownership.
func TestProvenanceIntegration_DivergenceDetection(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup templates
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templatesDir, "rules", "test-rule.md"),
		[]byte("test rule content"), 0644))

	// Setup test rite
	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Materialize first time
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	result, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Load manifest after first sync
	manifestPath := filepath.Join(claudeDir, "PROVENANCE_MANIFEST.yaml")
	manifest1, err := provenance.Load(manifestPath)
	require.NoError(t, err)
	require.NotNil(t, manifest1.Entries["rules/test-rule.md"])
	assert.Equal(t, provenance.OwnerKnossos, manifest1.Entries["rules/test-rule.md"].Owner)

	// User modifies a knossos-owned file
	rulePath := filepath.Join(claudeDir, "rules", "test-rule.md")
	require.NoError(t, os.WriteFile(rulePath, []byte("USER MODIFIED CONTENT"), 0644))

	// Materialize second time (should detect divergence)
	result, err = m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Load manifest after second sync
	manifest2, err := provenance.Load(manifestPath)
	require.NoError(t, err)

	// Verify the modified rule was promoted to user ownership
	require.NotNil(t, manifest2.Entries["rules/test-rule.md"])
	assert.Equal(t, provenance.OwnerUser, manifest2.Entries["rules/test-rule.md"].Owner,
		"modified file should be promoted to user ownership")

	// Note: Per TDD Section 6, the pipeline still writes the file on this sync
	// (divergence detection happens BEFORE writes). The manifest records user ownership,
	// and on the NEXT sync, the file will be skipped because it's user-owned.
	// For this test, we just verify the manifest promotion happened.
}

// TestProvenanceIntegration_Idempotency verifies that materializing twice produces
// identical manifests (no spurious divergence).
func TestProvenanceIntegration_Idempotency(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup templates
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	// Setup test rite
	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Materialize first time
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	result, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Load first manifest
	manifestPath := filepath.Join(claudeDir, "PROVENANCE_MANIFEST.yaml")
	manifest1, err := provenance.Load(manifestPath)
	require.NoError(t, err)

	// Materialize second time (idempotency test)
	result, err = m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Load second manifest
	manifest2, err := provenance.Load(manifestPath)
	require.NoError(t, err)

	// Verify same entries exist (ignore CLAUDE.md which may have generation timestamps)
	// Note: sync/state.json, KNOSSOS_MANIFEST.yaml, and ACTIVE_RITE are not tracked in
	// provenance manifest because they always regenerate with new timestamps/versions
	for path, entry1 := range manifest1.Entries {
		if path == "CLAUDE.md" {
			continue // May have generation timestamps in content, so checksum might differ
		}
		entry2, ok := manifest2.Entries[path]
		require.True(t, ok, "entry %s should exist in second manifest", path)
		assert.Equal(t, entry1.Owner, entry2.Owner, "owner should match for %s", path)
		assert.Equal(t, entry1.SourcePath, entry2.SourcePath, "source path should match for %s", path)
		assert.Equal(t, entry1.SourceType, entry2.SourceType, "source type should match for %s", path)
		assert.Equal(t, entry1.Checksum, entry2.Checksum, "checksum should match for %s", path)
	}
}

// TestProvenanceIntegration_DryRun verifies that dry-run mode does not write manifest.
func TestProvenanceIntegration_DryRun(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup templates
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	// Setup test rite
	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Materialize in dry-run mode
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	result, err := m.MaterializeWithOptions("test-rite", Options{DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Verify PROVENANCE_MANIFEST.yaml was NOT created
	manifestPath := filepath.Join(claudeDir, "PROVENANCE_MANIFEST.yaml")
	_, err = os.Stat(manifestPath)
	assert.True(t, os.IsNotExist(err), "manifest should not exist in dry-run mode")
}

// TestProvenanceIntegration_MenaDirectories verifies directory-level provenance
// for mena entries (commands/ and skills/).
func TestProvenanceIntegration_MenaDirectories(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")
	menaDir := filepath.Join(projectDir, ".knossos", "mena")

	// Setup templates
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))

	// Setup mena entries
	testCommandDir := filepath.Join(menaDir, "test-command")
	require.NoError(t, os.MkdirAll(testCommandDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(testCommandDir, "INDEX.dro.md"),
		[]byte("# Test Command\n\nCommand content"), 0644))

	testSkillDir := filepath.Join(menaDir, "test-skill")
	require.NoError(t, os.MkdirAll(testSkillDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(testSkillDir, "INDEX.lego.md"),
		[]byte("# Test Skill\n\nSkill content"), 0644))

	// Setup test rite (with empty manifest)
	setupRite(t, ritesDir, "test-rite",
		"",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Materialize
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	result, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Load manifest
	manifestPath := filepath.Join(claudeDir, "PROVENANCE_MANIFEST.yaml")
	manifest, err := provenance.Load(manifestPath)
	require.NoError(t, err)

	// Verify mena directory entries (with trailing slash)
	commandEntry := manifest.Entries["commands/test-command/"]
	require.NotNil(t, commandEntry, "command directory should be recorded")
	assert.Equal(t, provenance.OwnerKnossos, commandEntry.Owner)
	assert.NotEmpty(t, commandEntry.Checksum, "command directory should have checksum")

	skillEntry := manifest.Entries["skills/test-skill/"]
	require.NotNil(t, skillEntry, "skill directory should be recorded")
	assert.Equal(t, provenance.OwnerKnossos, skillEntry.Owner)
	assert.NotEmpty(t, skillEntry.Checksum, "skill directory should have checksum")

	// Verify command checksum matches promoted file (INDEX.md promoted to test-command.md)
	// Dromena without companions have no subdirectory — checksum is from promoted file
	promotedCommandPath := filepath.Join(claudeDir, "commands", "test-command.md")
	promotedData, err := os.ReadFile(promotedCommandPath)
	require.NoError(t, err, "promoted command file should exist")
	commandHash := checksum.Content(string(promotedData))
	assert.Equal(t, commandHash, commandEntry.Checksum, "command checksum should match promoted file")

	// Verify skill checksum matches directory (legomena are not promoted)
	skillDirPath := filepath.Join(claudeDir, "skills", "test-skill")
	skillHash, err := checksum.Dir(skillDirPath)
	require.NoError(t, err)
	assert.Equal(t, skillHash, skillEntry.Checksum, "skill directory checksum should match")
}
