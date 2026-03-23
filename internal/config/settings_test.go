package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSettings(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()

	// 1. Setup User scope ($KNOSSOS_HOME)
	knossosHome := filepath.Join(tmpDir, "knossos-home")
	require.NoError(t, os.MkdirAll(knossosHome, 0755))

	// Write User config
	userConfigContent := `
experimental:
  featureA: true
  featureB: false
`
	require.NoError(t, os.WriteFile(filepath.Join(knossosHome, "config.yaml"), []byte(userConfigContent), 0644))

	// Reset and set KNOSSOS_HOME for the test
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Cleanup(config.ResetKnossosHome)

	// 2. Setup Project scope
	projectRoot := filepath.Join(tmpDir, "project")
	knossosDir := filepath.Join(projectRoot, ".knossos")
	require.NoError(t, os.MkdirAll(knossosDir, 0755))

	// Write Project config (Overrides featureB, adds featureC)
	projectConfigContent := `
experimental:
  featureB: true
  featureC: true
`
	require.NoError(t, os.WriteFile(filepath.Join(knossosDir, "config.yaml"), []byte(projectConfigContent), 0644))

	// 3. Test Loading with Project scope
	t.Run("Merge User and Project scope", func(t *testing.T) {
		settings, err := config.LoadSettings(projectRoot)
		require.NoError(t, err)
		require.NotNil(t, settings)

		require.NotNil(t, settings.Experimental)
		assert.True(t, settings.Experimental["featureA"], "User setting preserved")
		assert.True(t, settings.Experimental["featureB"], "Project setting overrides User setting")
		assert.True(t, settings.Experimental["featureC"], "Project setting added")
	})

	// 4. Test Loading without Project scope (only User scope)
	t.Run("User scope only", func(t *testing.T) {
		settings, err := config.LoadSettings("")
		require.NoError(t, err)
		require.NotNil(t, settings)

		require.NotNil(t, settings.Experimental)
		assert.True(t, settings.Experimental["featureA"])
		assert.False(t, settings.Experimental["featureB"], "User setting unmodified")
		assert.False(t, settings.Experimental["featureC"])
	})

	// 5. Test Missing Files
	t.Run("Missing files gracefully handled", func(t *testing.T) {
		emptyProject := filepath.Join(tmpDir, "empty-project")

		// We need an empty home here so it doesn't find the config
		emptyHome := filepath.Join(tmpDir, "empty-home")
		config.ResetKnossosHome()
		t.Setenv("KNOSSOS_HOME", emptyHome)
		t.Cleanup(config.ResetKnossosHome)

		settings, err := config.LoadSettings(emptyProject)
		require.NoError(t, err)
		require.NotNil(t, settings)
		require.NotNil(t, settings.Experimental, "Experimental map should be initialized")
		assert.Empty(t, settings.Experimental)
	})

	// 6. Test Parse Error
	t.Run("Parse Error handling", func(t *testing.T) {
		badProject := filepath.Join(tmpDir, "bad-project")
		badKnossosDir := filepath.Join(badProject, ".knossos")
		require.NoError(t, os.MkdirAll(badKnossosDir, 0755))

		// Write invalid yaml
		require.NoError(t, os.WriteFile(filepath.Join(badKnossosDir, "config.yaml"), []byte("invalid:\n  - yaml\n- content"), 0644))

		_, err := config.LoadSettings(badProject)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to parse config.yaml")
	})
}
