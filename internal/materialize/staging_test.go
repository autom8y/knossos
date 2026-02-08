package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	err := atomicWriteFile(target, []byte("hello"), 0644)
	require.NoError(t, err)

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))

	// Verify no .tmp file left behind
	_, err = os.Stat(target + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

func TestAtomicWriteFile_Overwrites(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	require.NoError(t, os.WriteFile(target, []byte("old"), 0644))
	require.NoError(t, atomicWriteFile(target, []byte("new"), 0644))

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))
}

func TestWriteIfChanged_SkipsIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	require.NoError(t, os.WriteFile(target, []byte("same"), 0644))

	changed, err := writeIfChanged(target, []byte("same"), 0644)
	require.NoError(t, err)
	assert.False(t, changed, "should not write identical content")
}

func TestWriteIfChanged_WritesWhenDifferent(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	require.NoError(t, os.WriteFile(target, []byte("old"), 0644))

	changed, err := writeIfChanged(target, []byte("new"), 0644)
	require.NoError(t, err)
	assert.True(t, changed)

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))
}

func TestCloneDir(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "clone")

	// Create source structure
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("bbb"), 0644))

	// Also create a .tmp file that should be skipped
	require.NoError(t, os.WriteFile(filepath.Join(src, "partial.tmp"), []byte("junk"), 0644))

	require.NoError(t, cloneDir(src, dst))

	// Verify cloned files
	content, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	require.NoError(t, err)
	assert.Equal(t, "aaa", string(content))

	content, err = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	require.NoError(t, err)
	assert.Equal(t, "bbb", string(content))

	// .tmp file should be skipped
	_, err = os.Stat(filepath.Join(dst, "partial.tmp"))
	assert.True(t, os.IsNotExist(err), ".tmp files should be skipped during clone")
}

func TestStagedMaterialize_SwapsDirectories(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")

	// Create initial .claude/ with user content
	require.NoError(t, os.MkdirAll(filepath.Join(claudeDir, "sessions"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "sessions", "session.json"),
		[]byte(`{"id":"user-session"}`), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "ACTIVE_RITE"),
		[]byte("old-rite\n"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// Run staged materialization that writes new content
	result, err := m.StagedMaterialize(func(sm *Materializer) (*Result, error) {
		dir := sm.getClaudeDir()
		// Write new content to staging
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "ACTIVE_RITE"),
			[]byte("new-rite\n"), 0644))
		// Create a new file
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "agents"), 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "agents", "test.md"),
			[]byte("# Test Agent"), 0644))
		return &Result{Status: "staged"}, nil
	})

	require.NoError(t, err)
	assert.Equal(t, "staged", result.Status)

	// Verify new content is in .claude/
	content, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "new-rite\n", string(content))

	content, err = os.ReadFile(filepath.Join(claudeDir, "agents", "test.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Test Agent", string(content))

	// Verify user content survived the swap
	content, err = os.ReadFile(filepath.Join(claudeDir, "sessions", "session.json"))
	require.NoError(t, err)
	assert.Equal(t, `{"id":"user-session"}`, string(content))

	// Verify no staging or backup dirs remain
	_, err = os.Stat(claudeDir + ".staging")
	assert.True(t, os.IsNotExist(err), "staging dir should be cleaned up")
	_, err = os.Stat(claudeDir + ".bak")
	assert.True(t, os.IsNotExist(err), "backup dir should be cleaned up")
}

func TestStagedMaterialize_RollbackOnError(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")

	// Create initial .claude/ with content
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "ACTIVE_RITE"),
		[]byte("original\n"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	// Run staged materialization that fails
	_, err := m.StagedMaterialize(func(sm *Materializer) (*Result, error) {
		return nil, assert.AnError
	})

	require.Error(t, err)

	// Verify original .claude/ is preserved
	content, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "original\n", string(content))

	// Verify staging dir was cleaned up
	_, err = os.Stat(claudeDir + ".staging")
	assert.True(t, os.IsNotExist(err), "staging dir should be cleaned up on error")
}

func TestStagedMaterialize_NoExistingClaudeDir(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")

	// No .claude/ exists — staging should still work (bootstrap case)
	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	result, err := m.StagedMaterialize(func(sm *Materializer) (*Result, error) {
		dir := sm.getClaudeDir()
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "ACTIVE_RITE"),
			[]byte("fresh\n"), 0644))
		return &Result{Status: "fresh"}, nil
	})

	require.NoError(t, err)
	assert.Equal(t, "fresh", result.Status)

	content, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_RITE"))
	require.NoError(t, err)
	assert.Equal(t, "fresh\n", string(content))
}
