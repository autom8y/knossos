package mena

import (
	"os"
	"path/filepath"
	"strings"
)

// WalkEntry holds the data passed to a Walk callback.
type WalkEntry struct {
	Path    string // Absolute filesystem path to the file
	RelPath string // Path relative to the MenaSource.Path directory
	Data    []byte // File content
}

// Walk iterates all files matching suffix within the given sources.
// For each matching file, it reads the content and invokes fn.
//
// Sources with empty or nonexistent paths are silently skipped.
// Files that cannot be read are silently skipped (consistent with Exists
// behavior where os.ReadDir/os.Stat errors return false, not error).
//
// Walk does NOT support embedded FS sources (IsEmbedded=true are skipped).
// Lint operates on filesystem sources only; embedded FS iteration can be
// added later if a consumer needs it.
//
// The suffix filter matches against the full filename, not just the extension.
// Example suffixes: ".dro.md", ".lego.md".
func Walk(sources []MenaSource, suffix string, fn func(WalkEntry)) {
	for _, src := range sources {
		// Skip embedded sources — no filesystem path to walk.
		if src.IsEmbedded {
			continue
		}
		// Skip sources with no path set.
		if src.Path == "" {
			continue
		}

		// filepath.WalkDir (Go 1.16+) is preferred over filepath.Walk:
		// it avoids an os.Lstat call per entry, which is measurably faster
		// on large directory trees. Knossos requires Go 1.22+.
		_ = filepath.WalkDir(src.Path, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				// If the root itself does not exist, WalkDir returns the error
				// immediately. Silently skip the entire source.
				return nil
			}
			// Skip directories; continue walking into them.
			if d.IsDir() {
				return nil
			}
			// Filter by suffix on the full filename.
			if !strings.HasSuffix(d.Name(), suffix) {
				return nil
			}

			data, readErr := os.ReadFile(path)
			if readErr != nil {
				// Unreadable files are silently skipped.
				return nil
			}

			// relPath is relative to the source directory, not to any project root.
			// Callers that need project-relative display paths compute them externally.
			relPath, _ := filepath.Rel(src.Path, path)

			fn(WalkEntry{Path: path, RelPath: relPath, Data: data})
			return nil
		})
	}
}
