package config

import (
	"maps"
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"gopkg.in/yaml.v3"
)

// Settings represents the structure of the config.yaml file.
type Settings struct {
	Experimental map[string]bool `yaml:"experimental,omitempty"`
}

// LoadSettings resolves and merges configuration from the User scope
// ($KNOSSOS_HOME/config.yaml) and the Project scope (.knossos/config.yaml).
//
// Settings at the Project scope strictly override those at the User scope.
// If projectRoot is provided, the Project scope is evaluated; otherwise,
// only the User scope is used.
func LoadSettings(projectRoot string) (*Settings, error) {
	settings := &Settings{
		Experimental: make(map[string]bool),
	}

	// 1. Load User scope: $KNOSSOS_HOME/config.yaml
	userConfigPath := filepath.Join(KnossosHome(), "config.yaml")
	if err := loadAndMerge(userConfigPath, settings); err != nil {
		return nil, err
	}

	// 2. Load Project scope: .knossos/config.yaml
	if projectRoot != "" {
		resolver := paths.NewResolver(projectRoot)
		projectConfigPath := filepath.Join(resolver.KnossosDir(), "config.yaml")
		if err := loadAndMerge(projectConfigPath, settings); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

func loadAndMerge(path string, base *Settings) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing configuration files are not errors.
			return nil
		}
		return errors.Wrap(errors.CodePermissionDenied, "Failed to read config file", err)
	}

	var temp Settings
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return errors.Wrap(errors.CodeParseError, "Failed to parse config.yaml", err)
	}

	// Merge temp into base.
	// Project settings strictly override User settings.
	if temp.Experimental != nil {
		if base.Experimental == nil {
			base.Experimental = make(map[string]bool)
		}
		maps.Copy(base.Experimental, temp.Experimental)
	}

	return nil
}
