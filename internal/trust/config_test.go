package trust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig_IsValid(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Validate()
	require.NoError(t, err, "DefaultConfig must produce a valid configuration")
}

func TestDefaultConfig_ConservativeValues(t *testing.T) {
	cfg := DefaultConfig()

	// Verify conservative thresholds per Decision #14
	assert.Equal(t, 0.7, cfg.Thresholds.HighThreshold)
	assert.Equal(t, 0.4, cfg.Thresholds.LowThreshold)

	// Verify weights
	assert.Equal(t, 0.45, cfg.Weights.Freshness)
	assert.Equal(t, 0.25, cfg.Weights.Retrieval)
	assert.Equal(t, 0.30, cfg.Weights.Coverage)

	// Verify decay defaults (empirical values from Sprint-2 parameter sweep, D-1)
	assert.Equal(t, 7.0, cfg.Decay.DefaultHalfLifeDays)
	assert.Equal(t, 14.0, cfg.Decay.HalfLives[DomainArchitecture])
	assert.Equal(t, 7.0, cfg.Decay.HalfLives[DomainConventions])
	assert.Equal(t, 14.0, cfg.Decay.HalfLives[DomainDesignConstraints])
	assert.Equal(t, 10.0, cfg.Decay.HalfLives[DomainScarTissue])
	assert.Equal(t, 5.0, cfg.Decay.HalfLives[DomainTestCoverage])
	assert.Equal(t, 10.0, cfg.Decay.HalfLives[DomainFeat])
	assert.Equal(t, 3.0, cfg.Decay.HalfLives[DomainRelease])
	assert.Equal(t, 90.0, cfg.Decay.HalfLives[DomainLiterature])
}

func TestDefaultConfig_AllDomainTypesHaveHalfLives(t *testing.T) {
	cfg := DefaultConfig()

	// Every known domain type (except unknown) should have an explicit half-life
	knownTypes := []DomainType{
		DomainArchitecture, DomainConventions, DomainDesignConstraints,
		DomainScarTissue, DomainTestCoverage, DomainFeat,
		DomainRelease, DomainLiterature,
	}
	for _, dt := range knownTypes {
		_, ok := cfg.Decay.HalfLives[dt]
		assert.True(t, ok, "domain type %s should have an explicit half-life", dt)
	}

	// DomainUnknown should NOT be in the map (falls back to default)
	_, ok := cfg.Decay.HalfLives[DomainUnknown]
	assert.False(t, ok, "DomainUnknown should fall back to DefaultHalfLifeDays")
}

func TestValidate_RejectsLowGreaterThanHigh(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Thresholds.LowThreshold = 0.8
	cfg.Thresholds.HighThreshold = 0.5
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "low_threshold")
}

func TestValidate_RejectsEqualThresholds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Thresholds.LowThreshold = 0.5
	cfg.Thresholds.HighThreshold = 0.5
	err := cfg.Validate()
	assert.Error(t, err)
}

func TestValidate_RejectsNegativeThresholds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Thresholds.LowThreshold = -0.1
	assert.Error(t, cfg.Validate())

	cfg = DefaultConfig()
	cfg.Thresholds.HighThreshold = -0.1
	assert.Error(t, cfg.Validate())
}

func TestValidate_RejectsThresholdsAboveOne(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Thresholds.HighThreshold = 1.5
	assert.Error(t, cfg.Validate())

	cfg = DefaultConfig()
	cfg.Thresholds.LowThreshold = 1.1
	assert.Error(t, cfg.Validate())
}

func TestValidate_RejectsZeroWeights(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Weights.Freshness = 0.0
	assert.Error(t, cfg.Validate())

	cfg = DefaultConfig()
	cfg.Weights.Retrieval = 0.0
	assert.Error(t, cfg.Validate())

	cfg = DefaultConfig()
	cfg.Weights.Coverage = 0.0
	assert.Error(t, cfg.Validate())
}

func TestValidate_RejectsNegativeWeights(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Weights.Freshness = -1.0
	assert.Error(t, cfg.Validate())
}

func TestValidate_RejectsNegativeHalfLife(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Decay.DefaultHalfLifeDays = -1
	assert.Error(t, cfg.Validate())
}

func TestValidate_RejectsZeroDefaultHalfLife(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Decay.DefaultHalfLifeDays = 0
	assert.Error(t, cfg.Validate())
}

func TestValidate_RejectsNegativeDomainHalfLife(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Decay.HalfLives[DomainArchitecture] = -5
	assert.Error(t, cfg.Validate())
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	cfg, err := LoadConfig("")
	require.NoError(t, err)
	assert.Equal(t, DefaultConfig(), cfg)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/trust.yaml")
	require.NoError(t, err)
	assert.Equal(t, DefaultConfig(), cfg)
}

func TestLoadConfig_OverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")

	yamlContent := `
thresholds:
  high_threshold: 0.5
  low_threshold: 0.3
`
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	// Overridden values
	assert.Equal(t, 0.5, cfg.Thresholds.HighThreshold)
	assert.Equal(t, 0.3, cfg.Thresholds.LowThreshold)

	// Defaults preserved for non-overridden fields
	assert.Equal(t, 7.0, cfg.Decay.DefaultHalfLifeDays)
	assert.Equal(t, 0.45, cfg.Weights.Freshness)
}

func TestLoadConfig_FullOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")

	yamlContent := `
decay:
  default_half_life_days: 14
  half_lives:
    architecture: 45
    test-coverage: 3
thresholds:
  high_threshold: 0.8
  low_threshold: 0.5
weights:
  freshness: 0.50
  retrieval: 0.30
  coverage: 0.20
`
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, 14.0, cfg.Decay.DefaultHalfLifeDays)
	assert.Equal(t, 45.0, cfg.Decay.HalfLives[DomainArchitecture])
	assert.Equal(t, 3.0, cfg.Decay.HalfLives[DomainTestCoverage])
	assert.Equal(t, 0.8, cfg.Thresholds.HighThreshold)
	assert.Equal(t, 0.5, cfg.Thresholds.LowThreshold)
	assert.Equal(t, 0.50, cfg.Weights.Freshness)
	assert.Equal(t, 0.30, cfg.Weights.Retrieval)
	assert.Equal(t, 0.20, cfg.Weights.Coverage)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")

	err := os.WriteFile(path, []byte("not: [valid: yaml: :::"), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse trust config")
}

func TestLoadConfig_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")

	// Low > High: fails validation
	yamlContent := `
thresholds:
  high_threshold: 0.3
  low_threshold: 0.8
`
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid trust config")
}

// C-3: Threshold configurability -- changing thresholds changes tier classification
func TestThresholdConfigurability(t *testing.T) {
	// This test is here in config_test.go because it tests the C-3 criterion
	// that configuration changes affect behavior. The actual scoring test is in
	// confidence_test.go. This test verifies the config loading path.
	dir := t.TempDir()

	// Default config
	defaultPath := filepath.Join(dir, "default.yaml")
	err := os.WriteFile(defaultPath, []byte(`
thresholds:
  high_threshold: 0.7
  low_threshold: 0.4
`), 0644)
	require.NoError(t, err)

	cfg1, err := LoadConfig(defaultPath)
	require.NoError(t, err)
	assert.Equal(t, 0.7, cfg1.Thresholds.HighThreshold)

	// Modified config
	modifiedPath := filepath.Join(dir, "modified.yaml")
	err = os.WriteFile(modifiedPath, []byte(`
thresholds:
  high_threshold: 0.5
  low_threshold: 0.3
`), 0644)
	require.NoError(t, err)

	cfg2, err := LoadConfig(modifiedPath)
	require.NoError(t, err)
	assert.Equal(t, 0.5, cfg2.Thresholds.HighThreshold)
	assert.Equal(t, 0.3, cfg2.Thresholds.LowThreshold)
}

func TestLoadConfig_RoundTrip(t *testing.T) {
	// Write default config to YAML, read it back, verify equality
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")

	// Manually write the complete default config (empirical values)
	yamlContent := `
decay:
  default_half_life_days: 7
  half_lives:
    architecture: 14
    conventions: 7
    design-constraints: 14
    scar-tissue: 10
    test-coverage: 5
    feat: 10
    release: 3
    literature: 90
thresholds:
  high_threshold: 0.7
  low_threshold: 0.4
weights:
  freshness: 0.45
  retrieval: 0.25
  coverage: 0.30
`
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	expected := DefaultConfig()
	assert.Equal(t, expected.Thresholds, cfg.Thresholds)
	assert.Equal(t, expected.Weights, cfg.Weights)
	assert.Equal(t, expected.Decay.DefaultHalfLifeDays, cfg.Decay.DefaultHalfLifeDays)
}
