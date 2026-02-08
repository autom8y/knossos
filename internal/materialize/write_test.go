package materialize

import (
	"os"
	"path/filepath"
	"testing"

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
