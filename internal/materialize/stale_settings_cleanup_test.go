package materialize

import (
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestManifest creates a minimal ProvenanceManifest for use in tests.
func buildTestManifest(entries map[string]*provenance.ProvenanceEntry) *provenance.ProvenanceManifest {
	m := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		Entries:       make(map[string]*provenance.ProvenanceEntry),
	}
	maps.Copy(m.Entries, entries)
	return m
}

// newMaterializerForTest creates a Materializer pointing at the given project directory.
func newMaterializerForTest(projectDir string) *Materializer {
	resolver := paths.NewResolver(projectDir)
	return NewMaterializer(resolver)
}

// --- cleanupStaleBlanketSettings tests ---

func TestCleanupStaleSettings_NoFile(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil)

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.False(t, removed, "should return false when settings.json does not exist")
}

func TestCleanupStaleSettings_AgentGuardFingerprint(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Blanket-deny agent-guard hook — the original writeDefaultSettings() template
	content := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook agent-guard"
          }
        ]
      }
    ]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte(content), 0644))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil) // No provenance entry for settings.json

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.True(t, removed, "should remove blanket-deny agent-guard settings.json")

	_, err := os.Stat(settingsPath)
	assert.True(t, os.IsNotExist(err), "settings.json should be deleted")
}

func TestCleanupStaleSettings_EmptyStubFingerprint(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Empty CC-default stub found in satellites
	content := `{"permissions":{"allow":[],"additionalDirectories":[]},"hooks":{}}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte(content), 0644))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil)

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.True(t, removed, "should remove empty CC-default stub settings.json")

	_, err := os.Stat(settingsPath)
	assert.True(t, os.IsNotExist(err), "settings.json should be deleted")
}

func TestCleanupStaleSettings_UserModified(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// User added additional custom hooks alongside the agent-guard hook
	content := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook agent-guard"
          }
        ]
      }
    ]
  },
  "permissions": {
    "allow": ["Bash(git:*)"]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte(content), 0644))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil)

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.False(t, removed, "should not remove settings.json with extra user fields")

	_, err := os.Stat(settingsPath)
	assert.NoError(t, err, "settings.json should still exist")
}

func TestCleanupStaleSettings_HasProvenance(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Matches fingerprint 2 but has a provenance entry — pipeline-managed
	content := `{"permissions":{"allow":[],"additionalDirectories":[]},"hooks":{}}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte(content), 0644))

	m := newMaterializerForTest(projectDir)
	// settings.json is tracked in the provenance manifest
	manifest := buildTestManifest(map[string]*provenance.ProvenanceEntry{
		"settings.json": provenance.NewKnossosEntry(
			provenance.ScopeRite,
			"rites/eco/templates/settings.json",
			"project",
			"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		),
	})

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.False(t, removed, "should not remove settings.json that has a provenance entry")

	_, err := os.Stat(settingsPath)
	assert.NoError(t, err, "settings.json should still exist")
}

func TestCleanupStaleSettings_InvalidJSON(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte("not valid json {{{}"), 0644))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil)

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.False(t, removed, "should not remove settings.json with invalid JSON")

	_, err := os.Stat(settingsPath)
	assert.NoError(t, err, "settings.json should still exist")
}

func TestCleanupStaleSettings_WhitespaceVariant(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Reformatted version of the blanket-deny pattern (different indentation/newlines)
	content := `{"hooks":{"PreToolUse":[{"matcher":".*","hooks":[{"type":"command","command":"ari hook agent-guard"}]}]}}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte(content), 0644))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil)

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.True(t, removed, "structural comparison should be whitespace-insensitive")

	_, err := os.Stat(settingsPath)
	assert.True(t, os.IsNotExist(err), "settings.json should be deleted")
}

func TestCleanupStaleSettings_PartialMatch(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// agent-guard hook that includes --allow-path: this is the current pipeline output, not stale
	content := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook agent-guard --allow-path /Users/dev/project"
          }
        ]
      }
    ]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	require.NoError(t, os.WriteFile(settingsPath, []byte(content), 0644))

	m := newMaterializerForTest(projectDir)
	manifest := buildTestManifest(nil)

	removed := m.cleanupStaleBlanketSettings(claudeDir, manifest)
	assert.False(t, removed, "agent-guard with --allow-path is not the stale pattern")

	_, err := os.Stat(settingsPath)
	assert.NoError(t, err, "settings.json should still exist")
}

// --- matchesStaleSettingsFingerprint unit tests ---

func TestMatchesStaleSettingsFingerprint_AgentGuardBlanketDeny(t *testing.T) {
	parsed := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": ".*",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "ari hook agent-guard",
						},
					},
				},
			},
		},
	}
	assert.True(t, matchesStaleSettingsFingerprint(parsed))
}

func TestMatchesStaleSettingsFingerprint_AgentGuardWithAllowPath(t *testing.T) {
	parsed := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": ".*",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "ari hook agent-guard --allow-path /some/path",
						},
					},
				},
			},
		},
	}
	assert.False(t, matchesStaleSettingsFingerprint(parsed))
}

func TestMatchesStaleSettingsFingerprint_EmptyStub(t *testing.T) {
	parsed := map[string]any{
		"permissions": map[string]any{
			"allow":                 []any{},
			"additionalDirectories": []any{},
		},
		"hooks": map[string]any{},
	}
	assert.True(t, matchesStaleSettingsFingerprint(parsed))
}

func TestMatchesStaleSettingsFingerprint_EmptyStubWithExtraPermissions(t *testing.T) {
	// permissions has extra fields beyond allow + additionalDirectories
	parsed := map[string]any{
		"permissions": map[string]any{
			"allow":                 []any{},
			"additionalDirectories": []any{},
			"deny":                  []any{"Bash(rm:*)"},
		},
		"hooks": map[string]any{},
	}
	assert.False(t, matchesStaleSettingsFingerprint(parsed))
}

func TestMatchesStaleSettingsFingerprint_EmptyStubNonEmptyAllow(t *testing.T) {
	parsed := map[string]any{
		"permissions": map[string]any{
			"allow":                 []any{"Bash(git:*)"},
			"additionalDirectories": []any{},
		},
		"hooks": map[string]any{},
	}
	assert.False(t, matchesStaleSettingsFingerprint(parsed))
}

func TestMatchesStaleSettingsFingerprint_EmptyStubWithNonEmptyHooks(t *testing.T) {
	// hooks is not empty — user added something
	parsed := map[string]any{
		"permissions": map[string]any{
			"allow":                 []any{},
			"additionalDirectories": []any{},
		},
		"hooks": map[string]any{
			"PostToolUse": []any{},
		},
	}
	assert.False(t, matchesStaleSettingsFingerprint(parsed))
}

func TestMatchesStaleSettingsFingerprint_EmptyObject(t *testing.T) {
	parsed := map[string]any{}
	assert.False(t, matchesStaleSettingsFingerprint(parsed))
}
