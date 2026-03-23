package materialize

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/paths"
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
		"agents/builder.md":   provenance.OwnerKnossos,
		"commands/commit/":    provenance.OwnerKnossos,
		"skills/conventions/": provenance.OwnerKnossos,
		"agents/my-custom.md": provenance.OwnerUser,
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
		"agents/builder.md":    provenance.OwnerKnossos,
		"ACTIVE_WORKFLOW.yaml": provenance.OwnerKnossos,
		".mcp.json":            provenance.OwnerKnossos,
		"CLAUDE.md":            provenance.OwnerKnossos,
		"GEMINI.md":            provenance.OwnerKnossos,
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
		"commands/commit/":    provenance.OwnerKnossos,
		"skills/conventions/": provenance.OwnerKnossos,
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
		"agents/builder.md":   provenance.OwnerKnossos,
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
	foundSelfEntry := slices.Contains(lines, ".gitignore")
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

// --- untrackKnossosFiles tests ---

// setupGitRepo initializes a temporary git repository for testing.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	return dir
}

// runGit runs a git command in the given directory and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, output)
	return strings.TrimSpace(string(output))
}

// gitTrackedFiles returns the set of tracked files under a path in the repo.
func gitTrackedFiles(t *testing.T, projectRoot, path string) map[string]bool {
	t.Helper()
	cmd := exec.Command("git", "-C", projectRoot, "ls-files", path)
	output, err := cmd.Output()
	require.NoError(t, err)
	tracked := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			tracked[line] = true
		}
	}
	return tracked
}

func TestUntrackKnossosFiles_NilManifest(t *testing.T) {
	t.Parallel()
	count := untrackKnossosFiles("/tmp", "/tmp/"+paths.ClaudeChannel{}.DirName(), nil)
	assert.Equal(t, 0, count)
}

func TestUntrackKnossosFiles_NoGitRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	channelDir := filepath.Join(dir, paths.ClaudeChannel{}.DirName())
	require.NoError(t, os.MkdirAll(channelDir, 0755))

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/potnia.md": provenance.OwnerKnossos,
	})

	count := untrackKnossosFiles(dir, channelDir, manifest)
	assert.Equal(t, 0, count, "no git repo means no untracking")
}

func TestUntrackKnossosFiles_TrackedFiles(t *testing.T) {
	t.Parallel()
	projectRoot := setupGitRepo(t)
	channelDir := filepath.Join(projectRoot, paths.ClaudeChannel{}.DirName())
	require.NoError(t, os.MkdirAll(filepath.Join(channelDir, "agents"), 0755))

	// Create and commit both knossos-owned and user-owned files
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "agents", "potnia.md"), []byte("knossos"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "agents", "custom.md"), []byte("user"), 0644))
	chDir := paths.ClaudeChannel{}.DirName()
	runGit(t, projectRoot, "add", chDir+"/agents/potnia.md", chDir+"/agents/custom.md")
	runGit(t, projectRoot, "commit", "-m", "initial")

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/potnia.md": provenance.OwnerKnossos,
		"agents/custom.md": provenance.OwnerUser,
	})

	count := untrackKnossosFiles(projectRoot, channelDir, manifest)
	assert.Equal(t, 1, count)

	// Verify potnia is untracked from index
	tracked := gitTrackedFiles(t, projectRoot, paths.ClaudeChannel{}.DirName())
	assert.False(t, tracked[paths.ClaudeChannel{}.DirName()+"/agents/potnia.md"], "knossos-owned file should be untracked")
	assert.True(t, tracked[paths.ClaudeChannel{}.DirName()+"/agents/custom.md"], "user-owned file should remain tracked")

	// Verify files still exist on disk (git rm --cached doesn't delete)
	assert.FileExists(t, filepath.Join(channelDir, "agents", "potnia.md"))
	assert.FileExists(t, filepath.Join(channelDir, "agents", "custom.md"))
}

func TestUntrackKnossosFiles_DirectoryEntries(t *testing.T) {
	t.Parallel()
	projectRoot := setupGitRepo(t)
	channelDir := filepath.Join(projectRoot, paths.ClaudeChannel{}.DirName())
	require.NoError(t, os.MkdirAll(filepath.Join(channelDir, "commands", "commit"), 0755))

	// Create and commit files under a directory entry
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "commands", "commit", "prompt.md"), []byte("test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "commands", "commit", "SKILL.md"), []byte("test"), 0644))
	runGit(t, projectRoot, "add", paths.ClaudeChannel{}.DirName()+"/commands/")
	runGit(t, projectRoot, "commit", "-m", "initial")

	// Directory entry in provenance (trailing slash)
	manifest := testManifest(map[string]provenance.OwnerType{
		"commands/commit/": provenance.OwnerKnossos,
	})

	count := untrackKnossosFiles(projectRoot, channelDir, manifest)
	assert.Equal(t, 2, count)

	// All files under the directory should be untracked
	tracked := gitTrackedFiles(t, projectRoot, paths.ClaudeChannel{}.DirName())
	assert.Empty(t, tracked, "all files under directory entry should be untracked")

	// Files still exist on disk
	assert.FileExists(t, filepath.Join(channelDir, "commands", "commit", "prompt.md"))
	assert.FileExists(t, filepath.Join(channelDir, "commands", "commit", "SKILL.md"))
}

func TestUntrackKnossosFiles_NoTrackedFiles(t *testing.T) {
	t.Parallel()
	projectRoot := setupGitRepo(t)
	channelDir := filepath.Join(projectRoot, paths.ClaudeChannel{}.DirName())
	require.NoError(t, os.MkdirAll(filepath.Join(channelDir, "agents"), 0755))

	// Create files but don't commit them (untracked)
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "agents", "potnia.md"), []byte("test"), 0644))
	// Need at least one commit for git to be functional
	require.NoError(t, os.WriteFile(filepath.Join(projectRoot, "README.md"), []byte("test"), 0644))
	runGit(t, projectRoot, "add", "README.md")
	runGit(t, projectRoot, "commit", "-m", "initial")

	manifest := testManifest(map[string]provenance.OwnerType{
		"agents/potnia.md": provenance.OwnerKnossos,
	})

	count := untrackKnossosFiles(projectRoot, channelDir, manifest)
	assert.Equal(t, 0, count, "untracked files need no removal from index")
}

func TestUntrackKnossosFiles_ExcludesOutliers(t *testing.T) {
	t.Parallel()
	projectRoot := setupGitRepo(t)
	channelDir := filepath.Join(projectRoot, paths.ClaudeChannel{}.DirName())
	require.NoError(t, os.MkdirAll(channelDir, 0755))

	// Create and commit CLAUDE.md (an outlier — inscription deferred as commitable)
	require.NoError(t, os.WriteFile(filepath.Join(channelDir, "CLAUDE.md"), []byte("inscription"), 0644))
	runGit(t, projectRoot, "add", paths.ClaudeChannel{}.DirName()+"/"+paths.ClaudeChannel{}.ContextFile())
	runGit(t, projectRoot, "commit", "-m", "initial")

	manifest := testManifest(map[string]provenance.OwnerType{
		"CLAUDE.md": provenance.OwnerKnossos,
	})

	count := untrackKnossosFiles(projectRoot, channelDir, manifest)
	assert.Equal(t, 0, count, "outlier entries should not be untracked")

	// CLAUDE.md should still be tracked
	tracked := gitTrackedFiles(t, projectRoot, paths.ClaudeChannel{}.DirName())
	assert.True(t, tracked[paths.ClaudeChannel{}.DirName()+"/"+paths.ClaudeChannel{}.ContextFile()], "CLAUDE.md should remain tracked")
}
