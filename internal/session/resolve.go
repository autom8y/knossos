package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/paths"
)

// ResolveSession resolves a Knossos session ID using a priority chain:
//   1. explicitID (from --session-id flag) — returned directly if non-empty
//   2. ccSessionID (from CC stdin payload) — looked up in .cc-map/
//   3. Smart scan — FindActiveSessions() fallback
//
// For priority 3 with 2+ active sessions, returns an error listing the sessions.
// The interactive prompt for CLI disambiguation is handled by callers, not here.
func ResolveSession(resolver *paths.Resolver, ccSessionID string, explicitID string) (string, error) {
	// Priority 1: Explicit ID from flag
	explicitID = strings.TrimSpace(explicitID)
	if explicitID != "" {
		return explicitID, nil
	}

	// Priority 2: CC map lookup
	ccSessionID = strings.TrimSpace(ccSessionID)
	if ccSessionID != "" {
		sanitized := sanitizeCCSessionID(ccSessionID)
		if sanitized != "" {
			mapFile := filepath.Join(resolver.CCMapDir(), sanitized)
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

	// Multiple active sessions — return error with list
	return "", fmt.Errorf("multiple active sessions detected: %v — use --session-id to specify or park extras", activeIDs)
}

// SetCCMap creates or updates a CC-to-Knossos session mapping.
// Creates the .cc-map/ directory if it doesn't exist.
func SetCCMap(resolver *paths.Resolver, ccSessionID string, knossosSessionID string) error {
	sanitized := sanitizeCCSessionID(ccSessionID)
	if sanitized == "" {
		return fmt.Errorf("invalid CC session ID: %q", ccSessionID)
	}

	ccMapDir := resolver.CCMapDir()
	if err := paths.EnsureDir(ccMapDir); err != nil {
		return fmt.Errorf("failed to create cc-map directory: %w", err)
	}

	mapFile := filepath.Join(ccMapDir, sanitized)
	if err := os.WriteFile(mapFile, []byte(knossosSessionID), 0644); err != nil {
		return fmt.Errorf("failed to write cc-map file: %w", err)
	}

	return nil
}

// ClearCCMap removes a CC-to-Knossos session mapping.
// Returns nil if the mapping doesn't exist.
func ClearCCMap(resolver *paths.Resolver, ccSessionID string) error {
	sanitized := sanitizeCCSessionID(ccSessionID)
	if sanitized == "" {
		return nil // No-op for invalid ID
	}

	mapFile := filepath.Join(resolver.CCMapDir(), sanitized)
	err := os.Remove(mapFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cc-map file: %w", err)
	}

	return nil
}

// sanitizeCCSessionID sanitizes a CC session ID for use as a filename.
// Uses filepath.Base to prevent directory traversal.
// Returns empty string for invalid inputs (empty, ".", "..").
func sanitizeCCSessionID(id string) string {
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
