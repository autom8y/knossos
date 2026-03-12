package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClearInvocationState_RemovesFile(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	channelDir := filepath.Join(projectDir, ".claude")
	knossosDir := filepath.Join(projectDir, ".knossos")
	require.NoError(t, os.MkdirAll(channelDir, 0755))
	require.NoError(t, os.MkdirAll(knossosDir, 0755))

	// Create INVOCATION_STATE.yaml in .knossos/
	invPath := filepath.Join(knossosDir, "INVOCATION_STATE.yaml")
	require.NoError(t, os.WriteFile(invPath, []byte("current_rite: old-rite\n"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	err := m.clearInvocationState(channelDir)
	require.NoError(t, err)

	_, err = os.Stat(invPath)
	assert.True(t, os.IsNotExist(err), "INVOCATION_STATE.yaml should be removed from .knossos/")
}

func TestClearInvocationState_NoFile(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	channelDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(channelDir, 0755))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	err := m.clearInvocationState(channelDir)
	require.NoError(t, err, "should not error when file doesn't exist")
}
