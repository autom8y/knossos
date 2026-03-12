package materialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testManifest builds a ProvenanceManifest from a map of path→owner pairs.
// All entries get valid provenance fields suitable for testing.
func testManifest(entries map[string]provenance.OwnerType) *provenance.ProvenanceManifest {
	now := time.Now().UTC()
	m := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      now,
		ActiveRite:    "test-rite",
		Entries:       make(map[string]*provenance.ProvenanceEntry, len(entries)),
	}
	for path, owner := range entries {
		entry := &provenance.ProvenanceEntry{
			Owner:      owner,
			Scope:      provenance.ScopeRite,
			Checksum:   "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			LastSynced: now,
		}
		if owner == provenance.OwnerKnossos {
			entry.SourcePath = "rites/test/" + path
			entry.SourceType = "project"
		}
		m.Entries[path] = entry
	}
	return m
}

func TestGenerateChannelGitignore_BasicEntries(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/builder.md":     provenance.OwnerKnossos,
		"commands/commit/":      provenance.OwnerKnossos,
		"skills/conventions/":   provenance.OwnerKnossos,
		"agents/my-custom.md":   provenance.OwnerUser,
	})

	written, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)
	assert.True(t, written)

	content, err := os.ReadFile(filepath.Join(channelDir, ".gitignore"))
	require.NoError(t, err)
	text := string(content)

	// Knossos entries present with / prefix
	assert.Contains(t, text, "/agents/builder.md\n")
	assert.Contains(t, text, "/commands/commit/\n")
	assert.Contains(t, text, "/skills/conventions/\n")

	// User entry absent
	assert.NotContains(t, text, "my-custom")

	// Header present
	assert.True(t, strings.HasPrefix(text, "# Auto-generated"))

	// Sorted: agents before commands before skills
	agentsIdx := strings.Index(text, "/agents/builder.md")
	commandsIdx := strings.Index(text, "/commands/commit/")
	skillsIdx := strings.Index(text, "/skills/conventions/")
	assert.Less(t, agentsIdx, commandsIdx)
	assert.Less(t, commandsIdx, skillsIdx)
}

func TestGenerateChannelGitignore_ExcludesOutliers(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/builder.md":      provenance.OwnerKnossos,
		"ACTIVE_WORKFLOW.yaml":   provenance.OwnerKnossos,
		".mcp.json":              provenance.OwnerKnossos,
		"CLAUDE.md":              provenance.OwnerKnossos,
		"GEMINI.md":              provenance.OwnerKnossos,
	})

	written, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)
	assert.True(t, written)

	content, err := os.ReadFile(filepath.Join(channelDir, ".gitignore"))
	require.NoError(t, err)
	text := string(content)

	// Normal entry present
	assert.Contains(t, text, "/agents/builder.md\n")

	// Outliers excluded
	assert.NotContains(t, text, "ACTIVE_WORKFLOW")
	assert.NotContains(t, text, ".mcp.json")
	assert.NotContains(t, text, "CLAUDE.md")
	assert.NotContains(t, text, "GEMINI.md")
}

func TestGenerateChannelGitignore_DirectoryEntries(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	manifest := testManifest(map[string]provenance.OwnerType{
		"commands/commit/":     provenance.OwnerKnossos,
		"skills/conventions/":  provenance.OwnerKnossos,
	})

	written, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)
	assert.True(t, written)

	content, err := os.ReadFile(filepath.Join(channelDir, ".gitignore"))
	require.NoError(t, err)
	text := string(content)

	// Trailing slash preserved (git interprets as directory match)
	assert.Contains(t, text, "/commands/commit/\n")
	assert.Contains(t, text, "/skills/conventions/\n")
}

func TestGenerateChannelGitignore_Idempotent(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/builder.md": provenance.OwnerKnossos,
		"settings.local.json": provenance.OwnerKnossos,
	})

	// First call writes
	written1, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)
	assert.True(t, written1)

	content1, err := os.ReadFile(filepath.Join(channelDir, ".gitignore"))
	require.NoError(t, err)

	// Second call with same manifest — no write
	written2, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)
	assert.False(t, written2, "second call should not write (identical content)")

	content2, err := os.ReadFile(filepath.Join(channelDir, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, string(content1), string(content2))
}

func TestGenerateChannelGitignore_NilManifest(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	written, err := generateChannelGitignore(channelDir, nil)
	require.NoError(t, err)
	assert.False(t, written)

	// No file created
	_, err = os.Stat(filepath.Join(channelDir, ".gitignore"))
	assert.True(t, os.IsNotExist(err))
}

func TestGenerateChannelGitignore_SelfEntry(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/builder.md": provenance.OwnerKnossos,
	})

	_, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(channelDir, ".gitignore"))
	require.NoError(t, err)
	text := string(content)

	// Self-entry present (makes the file invisible to git)
	assert.Contains(t, text, ".gitignore\n")
	// Self-entry should NOT have the / prefix (it's the file itself, not a tracked entry)
	lines := strings.Split(text, "\n")
	foundSelfEntry := false
	for _, line := range lines {
		if line == ".gitignore" {
			foundSelfEntry = true
			break
		}
	}
	assert.True(t, foundSelfEntry, "self-entry '.gitignore' should be present without / prefix")
}

func TestGenerateChannelGitignore_UserOnlyManifest(t *testing.T) {
	t.Parallel()
	channelDir := t.TempDir()

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/my-custom.md": provenance.OwnerUser,
		"commands/my-cmd/":    provenance.OwnerUser,
	})

	written, err := generateChannelGitignore(channelDir, manifest)
	require.NoError(t, err)
	assert.False(t, written, "no knossos entries means no gitignore needed")

	// No file created
	_, err = os.Stat(filepath.Join(channelDir, ".gitignore"))
	assert.True(t, os.IsNotExist(err))
}
