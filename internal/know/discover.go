package know

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ServiceBoundary represents a discovered service boundary in the repository.
type ServiceBoundary struct {
	Path       string `json:"path"`        // relative path from repo root
	MarkerFile string `json:"marker_file"` // which marker was found
	MarkerType string `json:"marker_type"` // "go", "node", "rust", "python", "java", "bazel"
	HasKnow    bool   `json:"has_know"`    // whether a .know/ directory already exists here
}

// serviceMarkers maps marker filenames to their type classification.
var serviceMarkers = map[string]string{
	"go.mod":         "go",
	"package.json":   "node",
	"Cargo.toml":     "rust",
	"pyproject.toml": "python",
	"pom.xml":        "java",
	"BUILD":          "bazel",
	"BUILD.bazel":    "bazel",
}

// skipDirs contains directory names to skip during discovery.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"testdata":     true,
}

// DiscoverServiceBoundaries walks the repository tree from repoRoot,
// identifying directories that contain service boundary markers (go.mod,
// package.json, Cargo.toml, etc.). The repo root itself is excluded
// (it is the monorepo root, not a service). Returns candidates sorted by path.
func DiscoverServiceBoundaries(repoRoot string) ([]ServiceBoundary, error) {
	rootAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve repo root: %w", err)
	}

	var boundaries []ServiceBoundary
	// Track directories already added to avoid duplicates (multiple markers in same dir).
	seen := make(map[string]bool)

	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}

		if d.IsDir() {
			name := d.Name()
			// Skip hidden directories and common non-source directories.
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this file is a service marker.
		markerType, isMarker := serviceMarkers[d.Name()]
		if !isMarker {
			return nil
		}

		dir := filepath.Dir(path)
		// Skip repo root — it's the monorepo root, not a service.
		if dir == rootAbs {
			return nil
		}

		// Skip if we already recorded this directory (e.g., BUILD + BUILD.bazel).
		if seen[dir] {
			return nil
		}
		seen[dir] = true

		rel, err := filepath.Rel(rootAbs, dir)
		if err != nil {
			return nil
		}

		hasKnow := false
		if info, statErr := os.Stat(filepath.Join(dir, ".know")); statErr == nil && info.IsDir() {
			hasKnow = true
		}

		boundaries = append(boundaries, ServiceBoundary{
			Path:       rel,
			MarkerFile: d.Name(),
			MarkerType: markerType,
			HasKnow:    hasKnow,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repository: %w", err)
	}

	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i].Path < boundaries[j].Path
	})

	return boundaries, nil
}
