// Package config provides configuration utilities for the Knossos platform.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	knossosHome      string
	knossosHomeOnce  sync.Once
	deprecationShown bool
)

// KnossosHome returns the resolved Knossos platform home directory.
// Falls back to ROSTER_HOME with deprecation warning (shown once per process).
// Default: $HOME/Code/roster
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

	// Fallback: ROSTER_HOME (deprecated)
	if home := os.Getenv("ROSTER_HOME"); home != "" {
		if !deprecationShown && os.Getenv("KNOSSOS_SUPPRESS_DEPRECATION") != "1" {
			fmt.Fprintln(os.Stderr, "[DEPRECATED] ROSTER_HOME is deprecated. Set KNOSSOS_HOME instead.")
			fmt.Fprintln(os.Stderr, "  ROSTER_HOME support will be removed in version 3.0")
			deprecationShown = true
		}
		return home
	}

	// Default
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Code", "roster")
}

// ResetKnossosHome resets the cached home directory (for testing only).
func ResetKnossosHome() {
	knossosHomeOnce = sync.Once{}
	knossosHome = ""
	deprecationShown = false
}
