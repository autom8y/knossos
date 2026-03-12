package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/paths"
)

// ResolveSession resolves a Knossos session ID using a priority chain:
//   1. explicitID (from --session-id flag) -- returned directly if non-empty
//   2. harnessSessionID (from harness stdin payload) -- looked up in .harness-map/
//   3. Smart scan -- FindActiveSessions() fallback
//
// For priority 3 with 2+ active sessions, returns an error listing the sessions.
// The interactive prompt for CLI disambiguation is handled by callers, not here.
func ResolveSession(resolver *paths.Resolver, harnessSessionID string, explicitID string) (string, error) {
	// Priority 1: Explicit ID from flag
	explicitID = strings.TrimSpace(explicitID)
	if explicitID != "" {
		return explicitID, nil
	}

	// Priority 2: Harness session map lookup
	harnessSessionID = strings.TrimSpace(harnessSessionID)
	if harnessSessionID != "" {
		sanitized := sanitizeHarnessSessionID(harnessSessionID)
		if sanitized != "" {
			mapFile := filepath.Join(resolver.HarnessMapDir(), sanitized)
			data, err := os.ReadFile(mapFile)
			if err == nil {
				return strings.TrimSpace(string(data)), nil
			}
			// If map file doesn't exist, fall through to Priority 3
		}
	}

	// Priority 3: Smart scan
	activeIDs, err := FindActiveSessions(resolver.SessionsDir())
	if err != nil {
		return "", err
	}

	if len(activeIDs) == 0 {
		return "", nil
	}

	if len(activeIDs) == 1 {
		return activeIDs[0], nil
	}

	// Multiple active sessions -- return error with list
	return "", fmt.Errorf("multiple active sessions detected: %v — use --session-id to specify or park extras", activeIDs)
}

// SetHarnessSessionMap creates or updates a harness-to-Knossos session mapping.
// Creates the .harness-map/ directory if it doesn't exist.
func SetHarnessSessionMap(resolver *paths.Resolver, harnessSessionID string, knossosSessionID string) error {
	sanitized := sanitizeHarnessSessionID(harnessSessionID)
	if sanitized == "" {
		return fmt.Errorf("invalid harness session ID: %q", harnessSessionID)
	}

	mapDir := resolver.HarnessMapDir()
	if err := paths.EnsureDir(mapDir); err != nil {
		return fmt.Errorf("failed to create harness-map directory: %w", err)
	}

	mapFile := filepath.Join(mapDir, sanitized)
	if err := os.WriteFile(mapFile, []byte(knossosSessionID), 0644); err != nil {
		return fmt.Errorf("failed to write harness-map file: %w", err)
	}

	return nil
}

// SetCCMap is a deprecated alias for SetHarnessSessionMap.
func SetCCMap(resolver *paths.Resolver, ccSessionID string, knossosSessionID string) error {
	return SetHarnessSessionMap(resolver, ccSessionID, knossosSessionID)
}

// ClearHarnessSessionMap removes a harness-to-Knossos session mapping.
// Returns nil if the mapping doesn't exist.
func ClearHarnessSessionMap(resolver *paths.Resolver, harnessSessionID string) error {
	sanitized := sanitizeHarnessSessionID(harnessSessionID)
	if sanitized == "" {
		return nil // No-op for invalid ID
	}

	mapFile := filepath.Join(resolver.HarnessMapDir(), sanitized)
	err := os.Remove(mapFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove harness-map file: %w", err)
	}

	return nil
}

// ClearCCMap is a deprecated alias for ClearHarnessSessionMap.
func ClearCCMap(resolver *paths.Resolver, ccSessionID string) error {
	return ClearHarnessSessionMap(resolver, ccSessionID)
}

// ClearHarnessSessionMapForSession scans the harness map directory and removes
// all entries that map to the given Knossos session ID. This is used during
// session wrap to prevent stale map entries from accumulating.
//
// The scan is O(n) where n is the number of entries in .harness-map/. In practice
// there is at most one entry per active harness conversation.
// Returns nil if no entries found or directory doesn't exist.
func ClearHarnessSessionMapForSession(resolver *paths.Resolver, knossosSessionID string) error {
	mapDir := resolver.HarnessMapDir()
	entries, err := os.ReadDir(mapDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read harness-map directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		mapFile := filepath.Join(mapDir, entry.Name())
		data, readErr := os.ReadFile(mapFile)
		if readErr != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == knossosSessionID {
			_ = os.Remove(mapFile) // best-effort, ignore errors
		}
	}
	return nil
}

// ClearCCMapForSession is a deprecated alias for ClearHarnessSessionMapForSession.
func ClearCCMapForSession(resolver *paths.Resolver, knossosSessionID string) error {
	return ClearHarnessSessionMapForSession(resolver, knossosSessionID)
}

// sanitizeHarnessSessionID sanitizes a harness session ID for use as a filename.
// Uses filepath.Base to prevent directory traversal.
// Returns empty string for invalid inputs (empty, ".", "..").
func sanitizeHarnessSessionID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}

	// Use filepath.Base to get just the filename component
	// This prevents directory traversal attacks
	base := filepath.Base(id)

	// Reject "." and ".." which are special directory entries
	if base == "." || base == ".." {
		return ""
	}

	return base
}
