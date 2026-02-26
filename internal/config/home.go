// Package config provides configuration utilities for the Knossos platform.
package config

import (
	"os"
	"path/filepath"
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

// ResetKnossosHome resets the cached home directory (for testing only).
func ResetKnossosHome() {
	knossosHomeOnce = sync.Once{}
	knossosHome = ""
}
