// Package config provides configuration utilities for the Knossos platform.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	knossosHome     string
	knossosHomeOnce sync.Once
)

// KnossosHome returns the resolved Knossos platform home directory.
// Default: $HOME/Code/knossos
func KnossosHome() string {
	knossosHomeOnce.Do(func() {
		knossosHome = resolveKnossosHome()
	})
	return knossosHome
}

func resolveKnossosHome() string {
	// Primary: KNOSSOS_HOME
	if home := os.Getenv("KNOSSOS_HOME"); home != "" {
		return home
	}

	// Default
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Code", "knossos")
}

// XDGDataDir returns the XDG data directory for knossos.
// On macOS: ~/Library/Application Support/knossos
// On Linux: $XDG_DATA_HOME/knossos (default: ~/.local/share/knossos)
func XDGDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "knossos")
	}
	homeDir, _ := os.UserHomeDir()
	// macOS convention
	if _, err := os.Stat(filepath.Join(homeDir, "Library")); err == nil {
		return filepath.Join(homeDir, "Library", "Application Support", "knossos")
	}
	// Linux/default
	return filepath.Join(homeDir, ".local", "share", "knossos")
}

// ActiveOrg returns the currently active organization name.
// Resolution: $KNOSSOS_ORG env var, then $XDG_CONFIG_HOME/knossos/active-org file.
// Returns empty string if no org is configured.
func ActiveOrg() string {
	if org := os.Getenv("KNOSSOS_ORG"); org != "" {
		return org
	}

	// Inline XDG config path resolution to avoid circular import with paths package.
	// Must match paths.ConfigDir() logic: macOS uses ~/Library/Application Support.
	var configDir string
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		configDir = filepath.Join(xdg, "knossos")
	} else {
		homeDir, _ := os.UserHomeDir()
		if _, err := os.Stat(filepath.Join(homeDir, "Library")); err == nil {
			configDir = filepath.Join(homeDir, "Library", "Application Support", "knossos")
		} else {
			configDir = filepath.Join(homeDir, ".config", "knossos")
		}
	}

	data, err := os.ReadFile(filepath.Join(configDir, "active-org"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ResetKnossosHome resets the cached home directory (for testing only).
//
// Tests that need a custom KNOSSOS_HOME must follow this pattern to avoid
// cache poisoning across test cases:
//
//	config.ResetKnossosHome()
//	t.Setenv("KNOSSOS_HOME", tmpDir)
//	t.Cleanup(config.ResetKnossosHome)
//
// RISK-003: KnossosHome is cached via sync.Once. A test that calls KnossosHome
// (directly or transitively) before setting KNOSSOS_HOME will silently poison
// the cache for all subsequent tests in the same process. Always reset before
// and after any test that requires a custom home directory.
func ResetKnossosHome() {
	knossosHomeOnce = sync.Once{}
	knossosHome = ""
}
