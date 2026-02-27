package mena

import (
	"os"
	"path/filepath"
)

// ResolvePlatformMenaDir returns the filesystem path to the platform-level
// mena directory. Resolution order:
//  1. projectRoot/mena/ (knossos-on-knossos case)
//  2. XDG data dir/mena/ (installed user case)
//  3. knossosHome/mena/ (developer case)
//
// Returns "" if no platform mena directory exists on disk.
// The caller is responsible for falling back to embedded FS when "" is returned.
func ResolvePlatformMenaDir(projectRoot string, knossosHome string) string {
	// 1. Check for project-level mena first (knossos-on-knossos case)
	if projectRoot != "" {
		projectMena := filepath.Join(projectRoot, "mena")
		if _, err := os.Stat(projectMena); err == nil {
			return projectMena
		}
	}

	// 2. Check XDG data dir (installed user case)
	xdgMena := filepath.Join(xdgDataDir(), "mena")
	if _, err := os.Stat(xdgMena); err == nil {
		return xdgMena
	}

	// 3. Fall back to Knossos platform mena (developer case)
	if knossosHome != "" {
		knossosMena := filepath.Join(knossosHome, "mena")
		if _, err := os.Stat(knossosMena); err == nil {
			return knossosMena
		}
	}

	return ""
}

// xdgDataDir returns the XDG data directory for knossos.
// Replicates config.XDGDataDir() logic without importing internal/config
// to preserve this package as a stdlib-only leaf package.
//
// On macOS: ~/Library/Application Support/knossos
// On Linux: $XDG_DATA_HOME/knossos (default: ~/.local/share/knossos)
func xdgDataDir() string {
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
