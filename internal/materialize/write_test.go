package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteIfChanged_SkipsIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	require.NoError(t, os.WriteFile(target, []byte("same"), 0644))

	changed, err := fileutil.WriteIfChanged(target, []byte("same"), 0644)
	require.NoError(t, err)
	assert.False(t, changed, "should not write identical content")
}

func TestWriteIfChanged_WritesWhenDifferent(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	require.NoError(t, os.WriteFile(target, []byte("old"), 0644))

	changed, err := fileutil.WriteIfChanged(target, []byte("new"), 0644)
	require.NoError(t, err)
	assert.True(t, changed)

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))
}
