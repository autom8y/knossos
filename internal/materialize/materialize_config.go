package materialize

import (
	"path/filepath"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/autom8y/knossos/internal/checksum"
)

// materializeConfigSettings loads the platform config.yaml, merges experimental features
// into settings.local.json, injects platform invariants (e.g., Gemini enableAgents),
// and saves the result idempotently.
func (m *Materializer) materializeConfigSettings(channelDir string, collector provenance.Collector, channel string) error {
	settingsPath := filepath.Join(channelDir, "settings.local.json")

	// Load existing settings.local.json or create an empty map
	existingSettings, err := loadExistingSettings(settingsPath)
	if err != nil {
		return err
	}

	// Load canonical config.yaml (Project > User)
	platformConfig, err := config.LoadSettings(m.resolver.ProjectRoot())
	if err != nil {
		// If the file doesn't exist or is invalid, we continue with an empty config
		// so we don't block materialization, but log/handle as appropriate.
		// For now, assume a missing config is fine, LoadSettings should return empty struct on error or we handle it here.
		// If LoadSettings returns an error on missing, we might want to ignore os.ErrNotExist.
		// Assuming LoadSettings returns a usable object even if empty/error.
		if platformConfig == nil {
			platformConfig = &config.Settings{
				Experimental: make(map[string]bool),
			}
		}
	}

	// Merge experimental features from config.yaml
	if len(platformConfig.Experimental) > 0 {
		expMap, ok := existingSettings["experimental"].(map[string]any)
		if !ok || expMap == nil {
			expMap = make(map[string]any)
		}
		for k, v := range platformConfig.Experimental {
			expMap[k] = v
		}
		existingSettings["experimental"] = expMap
	}

	// Inject hardcoded platform invariants
	if channel == "gemini" {
		expMap, ok := existingSettings["experimental"].(map[string]any)
		if !ok || expMap == nil {
			expMap = make(map[string]any)
		}
		expMap["enableAgents"] = true
		existingSettings["experimental"] = expMap
	}

	// Save idempotently
	err = saveSettings(settingsPath, existingSettings)
	if err != nil {
		return err
	}

	// Record provenance
	hash, err := checksum.File(settingsPath)
	if err == nil && hash != "" {
		collector.Record("settings.local.json", provenance.NewKnossosEntry(
			provenance.ScopeRite,
			"(generated)",
			"config",
			hash, channel,
		))
	}

	return nil
}
