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

// ResetKnossosHome resets the cached home directory (for testing only).
func ResetKnossosHome() {
	knossosHomeOnce = sync.Once{}
	knossosHome = ""
}
