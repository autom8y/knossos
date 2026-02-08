package session

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindActiveSession scans session directories to find the currently ACTIVE session.
// Returns empty string if no active session exists.
// Returns error if multiple ACTIVE sessions detected (data corruption/race condition).
// This is the authoritative source of truth — .current-session is a cache.
func FindActiveSession(sessionsDir string) (string, error) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var activeIDs []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Quick check: session dirs start with "session-" and are 32+ chars
		if len(name) < 32 || name[:8] != "session-" {
			continue
		}

		contextPath := filepath.Join(sessionsDir, name, "SESSION_CONTEXT.md")
		status := readStatusFromFrontmatter(contextPath)
		if status == "ACTIVE" {
			activeIDs = append(activeIDs, name)
		}
	}

	if len(activeIDs) > 1 {
		return "", fmt.Errorf("multiple active sessions detected: %v — resolve with 'ari session park' to park extras", activeIDs)
	}

	if len(activeIDs) == 1 {
		return activeIDs[0], nil
	}

	return "", nil
}

// readStatusFromFrontmatter reads only the status field from YAML frontmatter.
// Reads only until the closing "---" to minimize I/O. Returns empty string on any error.
func readStatusFromFrontmatter(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// First line must be "---"
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return ""
	}

	// Scan frontmatter lines until closing "---"
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// End of frontmatter
		if trimmed == "---" {
			break
		}

		// Look for "status: VALUE"
		if strings.HasPrefix(trimmed, "status:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "status:"))
			return value
		}
	}

	return ""
}
