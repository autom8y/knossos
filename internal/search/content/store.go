// Package content provides content loading for .know/ domain files.
//
// ContentStore abstracts content retrieval so that the BM25 index can be
// populated with real .know/ file content regardless of runtime environment:
//   - Container: reads from pre-baked /data/content/ directory
//   - Local dev: reads from local filesystem repo paths
//
// This package resolves bug C-1 (empty BM25 index in production) by providing
// a content source that works in the distroless container where no git repos
// are available on the filesystem.
package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
)

// Store loads the text content of .know/ domain files.
// Implementations may read from local filesystem, pre-baked container paths,
// or other sources. The interface is intentionally narrow for testability.
type Store interface {
	// LoadContent returns the markdown body (YAML frontmatter stripped) for a domain entry.
	// Returns an error if the content cannot be loaded.
	LoadContent(entry registryorg.DomainEntry) (string, error)

	// HasContent returns true if content is available for the given domain entry.
	HasContent(entry registryorg.DomainEntry) bool
}

// PreBakedStore reads .know/ content from a pre-baked directory structure.
//
// The directory layout mirrors the catalog's qualified name hierarchy:
//
//	{ContentDir}/{repoName}/{path}
//
// For example, for qualified name "autom8y::knossos::architecture" with
// path ".know/architecture.md", the file is at:
//
//	{ContentDir}/knossos/.know/architecture.md
//
// This store is designed for container deployment where .know/ files are
// pre-baked into the Docker image at a known path (e.g., /data/content/).
type PreBakedStore struct {
	// ContentDir is the root directory containing pre-baked .know/ files.
	ContentDir string
}

// NewPreBakedStore creates a PreBakedStore that reads from the given directory.
func NewPreBakedStore(contentDir string) *PreBakedStore {
	return &PreBakedStore{ContentDir: contentDir}
}

// LoadContent reads a .know/ file from the pre-baked directory, strips YAML
// frontmatter, and returns the body text.
func (s *PreBakedStore) LoadContent(entry registryorg.DomainEntry) (string, error) {
	path, err := s.resolvePath(entry)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read content %s: %w", path, err)
	}

	return stripFrontmatter(string(data)), nil
}

// HasContent returns true if the .know/ file exists in the pre-baked directory.
func (s *PreBakedStore) HasContent(entry registryorg.DomainEntry) bool {
	path, err := s.resolvePath(entry)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// resolvePath maps a DomainEntry to a filesystem path within the pre-baked directory.
func (s *PreBakedStore) resolvePath(entry registryorg.DomainEntry) (string, error) {
	repoName := repoFromQualifiedName(entry.QualifiedName)
	if repoName == "" {
		return "", fmt.Errorf("cannot extract repo name from qualified name: %s", entry.QualifiedName)
	}

	return filepath.Join(s.ContentDir, repoName, entry.Path), nil
}

// LocalStore reads .know/ files from local filesystem repo paths.
// This is the development-time content source — it reads directly from
// cloned repositories on the developer's machine.
type LocalStore struct {
	// RepoPaths maps repository names to their local filesystem root paths.
	// e.g., {"knossos": "/Users/tom/Code/knossos", "auth": "/Users/tom/Code/auth"}
	RepoPaths map[string]string
}

// NewLocalStore creates a LocalStore with the given repo path mappings.
func NewLocalStore(repoPaths map[string]string) *LocalStore {
	return &LocalStore{RepoPaths: repoPaths}
}

// LoadContent reads a .know/ file from a local repo, strips YAML frontmatter,
// and returns the body text.
func (s *LocalStore) LoadContent(entry registryorg.DomainEntry) (string, error) {
	repoName := repoFromQualifiedName(entry.QualifiedName)
	if repoName == "" {
		return "", fmt.Errorf("cannot extract repo name from qualified name: %s", entry.QualifiedName)
	}

	repoRoot, ok := s.RepoPaths[repoName]
	if !ok {
		return "", fmt.Errorf("no local path for repo %s", repoName)
	}

	filePath := filepath.Join(repoRoot, entry.Path)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", filePath, err)
	}

	return stripFrontmatter(string(data)), nil
}

// HasContent returns true if the .know/ file exists in the local repo.
func (s *LocalStore) HasContent(entry registryorg.DomainEntry) bool {
	repoName := repoFromQualifiedName(entry.QualifiedName)
	if repoName == "" {
		return false
	}

	repoRoot, ok := s.RepoPaths[repoName]
	if !ok {
		return false
	}

	filePath := filepath.Join(repoRoot, entry.Path)
	_, err := os.Stat(filePath)
	return err == nil
}

// repoFromQualifiedName extracts the repo component from "org::repo::domain".
func repoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// stripFrontmatter removes YAML frontmatter delimited by ---.
func stripFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---") {
		return text
	}
	rest := text[3:]
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return text
	}
	return strings.TrimSpace(rest[idx+3:])
}

// DefaultContentDir is the container path where pre-baked .know/ content is stored.
// This matches the COPY directive in deploy/Dockerfile.
const DefaultContentDir = "/data/content"
