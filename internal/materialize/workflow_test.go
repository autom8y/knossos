package materialize

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterializeWorkflow_WritesFile(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Create a rite with workflow.yaml
	riteDir := filepath.Join(projectDir, "rites", "test-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	workflowContent := []byte("name: test-workflow\nphases:\n  - build\n")
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "workflow.yaml"), workflowContent, 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		RitePath: riteDir,
		Source:   RiteSource{Type: SourceProject, Path: riteDir},
	}

	err := m.materializeWorkflow(claudeDir, resolved)
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(workflowContent), string(got))
}

func TestMaterializeWorkflow_NoWorkflowFile(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Create a rite without workflow.yaml
	riteDir := filepath.Join(projectDir, "rites", "no-workflow")
	require.NoError(t, os.MkdirAll(riteDir, 0755))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		RitePath: riteDir,
		Source:   RiteSource{Type: SourceProject, Path: riteDir},
	}

	err := m.materializeWorkflow(claudeDir, resolved)
	require.NoError(t, err)

	// ACTIVE_WORKFLOW.yaml should not exist
	_, err = os.Stat(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	assert.True(t, os.IsNotExist(err))
}

func TestMaterializeWorkflow_RemovesStaleOnNoWorkflow(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Pre-existing ACTIVE_WORKFLOW.yaml from a previous rite
	staleContent := []byte("name: old-rite-workflow\nphases:\n  - stale\n")
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"), staleContent, 0644))

	// New rite has no workflow.yaml
	riteDir := filepath.Join(projectDir, "rites", "no-workflow")
	require.NoError(t, os.MkdirAll(riteDir, 0755))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		RitePath: riteDir,
		Source:   RiteSource{Type: SourceProject, Path: riteDir},
	}

	err := m.materializeWorkflow(claudeDir, resolved)
	require.NoError(t, err)

	// Stale file must be removed
	_, err = os.Stat(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	assert.True(t, os.IsNotExist(err), "stale ACTIVE_WORKFLOW.yaml should be removed when new rite has no workflow")
}

func TestMaterializeWorkflow_Idempotent(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	riteDir := filepath.Join(projectDir, "rites", "test-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	workflowContent := []byte("name: test-workflow\n")
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "workflow.yaml"), workflowContent, 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		RitePath: riteDir,
		Source:   RiteSource{Type: SourceProject, Path: riteDir},
	}

	// First write
	require.NoError(t, m.materializeWorkflow(claudeDir, resolved))
	// Second write should be a no-op (writeIfChanged returns false)
	require.NoError(t, m.materializeWorkflow(claudeDir, resolved))

	got, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(workflowContent), string(got))
}

func TestMaterializeWorkflow_EmbeddedSource(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	workflowContent := []byte("name: embedded-workflow\n")
	embeddedFS := fstest.MapFS{
		"rites/embedded-rite/workflow.yaml": &fstest.MapFile{Data: workflowContent},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver).WithEmbeddedFS(embeddedFS)

	resolved := &ResolvedRite{
		RitePath: "rites/embedded-rite",
		Source:   RiteSource{Type: SourceEmbedded, Path: "embedded"},
	}

	err := m.materializeWorkflow(claudeDir, resolved)
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(workflowContent), string(got))
}

func TestMaterializeWorkflow_OverwritesOld(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	// Pre-existing stale workflow
	require.NoError(t, os.WriteFile(
		filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"),
		[]byte("name: old-workflow\n"), 0644))

	// New rite with different workflow
	riteDir := filepath.Join(projectDir, "rites", "new-rite")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	newContent := []byte("name: new-workflow\nphases:\n  - deploy\n")
	require.NoError(t, os.WriteFile(filepath.Join(riteDir, "workflow.yaml"), newContent, 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		RitePath: riteDir,
		Source:   RiteSource{Type: SourceProject, Path: riteDir},
	}

	err := m.materializeWorkflow(claudeDir, resolved)
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(claudeDir, "ACTIVE_WORKFLOW.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(newContent), string(got))
}
