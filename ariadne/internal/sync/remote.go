// Package sync provides remote fetching for Ariadne.
// It handles HTTP(S), git refs, and local paths.
package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
	"gopkg.in/yaml.v3"
)

// RemoteType represents the type of remote source.
type RemoteType string

const (
	// RemoteTypeHTTP is an HTTP(S) URL.
	RemoteTypeHTTP RemoteType = "http"
	// RemoteTypeGit is a git reference.
	RemoteTypeGit RemoteType = "git"
	// RemoteTypeLocal is a local filesystem path.
	RemoteTypeLocal RemoteType = "local"
)

// Remote represents a remote source for sync.
type Remote struct {
	Type     RemoteType
	URL      string
	BasePath string // For git remotes, the base path in the repo
}

// RemoteFetcher handles fetching content from remotes.
type RemoteFetcher struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewRemoteFetcher creates a new remote fetcher.
func NewRemoteFetcher() *RemoteFetcher {
	return &RemoteFetcher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// ParseRemote parses a remote string into a Remote struct.
func ParseRemote(remote string) (*Remote, error) {
	// Check for HTTP(S) URL
	if strings.HasPrefix(remote, "http://") || strings.HasPrefix(remote, "https://") {
		return &Remote{
			Type: RemoteTypeHTTP,
			URL:  remote,
		}, nil
	}

	// Check for git URL patterns (git@host:path or git://host/path)
	if strings.HasPrefix(remote, "git@") || strings.HasPrefix(remote, "git://") {
		return &Remote{
			Type: RemoteTypeGit,
			URL:  remote,
		}, nil
	}

	// Check for GitHub shorthand (org/repo)
	if !strings.HasPrefix(remote, "/") && !strings.HasPrefix(remote, ".") &&
		strings.Count(remote, "/") == 1 && !strings.Contains(remote, ":") {
		// Convert to GitHub raw URL
		parts := strings.Split(remote, "/")
		if len(parts) == 2 {
			return &Remote{
				Type: RemoteTypeHTTP,
				URL:  fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main", parts[0], parts[1]),
			}, nil
		}
	}

	// Check if it's a local path
	if strings.HasPrefix(remote, "/") || strings.HasPrefix(remote, ".") ||
		strings.HasPrefix(remote, "~") {
		// Expand ~ to home directory
		if strings.HasPrefix(remote, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to get home directory", err)
			}
			remote = filepath.Join(home, remote[1:])
		}

		absPath, err := filepath.Abs(remote)
		if err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to resolve path", err)
		}

		return &Remote{
			Type: RemoteTypeLocal,
			URL:  absPath,
		}, nil
	}

	return nil, errors.NewWithDetails(errors.CodeUsageError,
		"invalid remote format",
		map[string]interface{}{
			"remote":   remote,
			"expected": "URL, git remote, org/repo, or local path",
		})
}

// FetchFile fetches a single file from the remote.
func (f *RemoteFetcher) FetchFile(remote *Remote, path string) ([]byte, error) {
	switch remote.Type {
	case RemoteTypeHTTP:
		return f.fetchHTTP(remote.URL, path)
	case RemoteTypeGit:
		return f.fetchGit(remote.URL, path)
	case RemoteTypeLocal:
		return f.fetchLocal(remote.URL, path)
	default:
		return nil, errors.New(errors.CodeGeneralError, "unknown remote type")
	}
}

// fetchHTTP fetches a file over HTTP(S).
func (f *RemoteFetcher) fetchHTTP(baseURL, path string) ([]byte, error) {
	fullURL := baseURL
	if path != "" {
		// Join URL and path properly
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "invalid base URL", err)
		}
		u.Path = filepath.Join(u.Path, path)
		fullURL = u.String()
	}

	resp, err := f.httpClient.Get(fullURL)
	if err != nil {
		return nil, errors.ErrNetworkError(fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrRemoteNotFound(fullURL)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrNetworkError(fullURL, fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read response body", err)
	}

	return data, nil
}

// fetchGit fetches a file from a git repository.
func (f *RemoteFetcher) fetchGit(gitURL, path string) ([]byte, error) {
	// For git, we need to clone or fetch the repo first
	// This is a simplified implementation - in production you'd want caching

	// Check if this is a git ref format (commit:path)
	if strings.Contains(gitURL, ":") && !strings.HasPrefix(gitURL, "git@") && !strings.HasPrefix(gitURL, "git://") {
		parts := strings.SplitN(gitURL, ":", 2)
		if len(parts) == 2 {
			// Use git show
			ref := parts[0]
			filePath := parts[1]
			if path != "" {
				filePath = filepath.Join(filepath.Dir(filePath), path)
			}

			cmd := exec.Command("git", "show", ref+":"+filePath)
			data, err := cmd.Output()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return nil, errors.NewWithDetails(errors.CodeFileNotFound,
						"git ref not found",
						map[string]interface{}{
							"ref":    ref + ":" + filePath,
							"stderr": string(exitErr.Stderr),
						})
				}
				return nil, errors.ErrNetworkError(gitURL, err)
			}
			return data, nil
		}
	}

	// For remote git URLs, we'd need to clone/fetch
	// For now, return an error suggesting local git refs
	return nil, errors.NewWithDetails(errors.CodeUsageError,
		"remote git URLs not yet supported",
		map[string]interface{}{
			"git_url":    gitURL,
			"suggestion": "Use local git ref format: HEAD:.claude/path or origin/main:.claude/path",
		})
}

// fetchLocal fetches a file from the local filesystem.
func (f *RemoteFetcher) fetchLocal(basePath, path string) ([]byte, error) {
	fullPath := basePath
	if path != "" {
		fullPath = filepath.Join(basePath, path)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ErrRemoteNotFound(fullPath)
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read file", err)
	}

	return data, nil
}

// FetchManifest fetches and parses a manifest file from the remote.
func (f *RemoteFetcher) FetchManifest(remote *Remote, path string) (map[string]interface{}, error) {
	data, err := f.FetchFile(remote, path)
	if err != nil {
		return nil, err
	}

	var content map[string]interface{}

	// Try JSON first, then YAML
	if err := json.Unmarshal(data, &content); err != nil {
		if err := yaml.Unmarshal(data, &content); err != nil {
			return nil, errors.NewWithDetails(errors.CodeParseError,
				"failed to parse manifest",
				map[string]interface{}{
					"path":  path,
					"cause": err.Error(),
				})
		}
	}

	return content, nil
}

// ListFiles lists files at a remote path (for directories).
// Only supported for local remotes.
func (f *RemoteFetcher) ListFiles(remote *Remote, path string) ([]string, error) {
	if remote.Type != RemoteTypeLocal {
		return nil, errors.New(errors.CodeUsageError, "listing files only supported for local remotes")
	}

	fullPath := remote.URL
	if path != "" {
		fullPath = filepath.Join(remote.URL, path)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ErrRemoteNotFound(fullPath)
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to list directory", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// Exists checks if a remote path exists.
func (f *RemoteFetcher) Exists(remote *Remote, path string) (bool, error) {
	_, err := f.FetchFile(remote, path)
	if err != nil {
		if errors.IsRemoteNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
