package org

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/config"
)

// GitHubClient abstracts GitHub API calls for testability.
// The httpGitHubClient struct implements this interface using stdlib net/http.
type GitHubClient interface {
	// ListOrgRepos returns all repositories for the given GitHub org.
	ListOrgRepos(org string) ([]GitHubRepo, error)
	// ListDirectoryContents returns the contents of a directory in a repo.
	ListDirectoryContents(owner, repo, dirPath string) ([]GitHubContent, error)
	// GetFileContent fetches the raw content of a file from a repo.
	GetFileContent(owner, repo, filePath string) ([]byte, error)
}

// GitHubRepo is a minimal representation of a GitHub repository.
type GitHubRepo struct {
	Name          string `json:"name"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Archived      bool   `json:"archived"`
	Fork          bool   `json:"fork"`
}

// GitHubContent represents an entry in a GitHub directory listing.
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
}

// httpGitHubClient implements GitHubClient using stdlib net/http.
// It sends a GitHub token for authentication when configured.
type httpGitHubClient struct {
	httpClient *http.Client
	token      string
	baseURL    string // default: "https://api.github.com"
}

// NewGitHubClient creates a GitHubClient using the given http.Client and token.
// Pass nil for httpClient to use http.DefaultClient.
// token may be empty for unauthenticated access (rate-limited).
func NewGitHubClient(httpClient *http.Client, token string) GitHubClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &httpGitHubClient{
		httpClient: httpClient,
		token:      token,
		baseURL:    "https://api.github.com",
	}
}

// NewGitHubClientWithBase creates a client with a custom base URL (for testing).
func NewGitHubClientWithBase(httpClient *http.Client, token, baseURL string) GitHubClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &httpGitHubClient{
		httpClient: httpClient,
		token:      token,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (c *httpGitHubClient) get(url string, out any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request %s: %w", url, err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response from %s: %w", url, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API %s returned status %d: %s", url, resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response from %s: %w", url, err)
	}
	return nil
}

func (c *httpGitHubClient) ListOrgRepos(org string) ([]GitHubRepo, error) {
	// Paginate: GitHub returns up to 100 per page.
	var allRepos []GitHubRepo
	page := 1
	for {
		url := fmt.Sprintf("%s/orgs/%s/repos?per_page=100&page=%d", c.baseURL, org, page)
		var repos []GitHubRepo
		if err := c.get(url, &repos); err != nil {
			return nil, fmt.Errorf("list repos for org %s (page %d): %w", org, page, err)
		}
		if len(repos) == 0 {
			break
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < 100 {
			// Last page
			break
		}
		page++
	}
	return allRepos, nil
}

func (c *httpGitHubClient) ListDirectoryContents(owner, repo, dirPath string) ([]GitHubContent, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.baseURL, owner, repo, dirPath)
	var contents []GitHubContent
	if err := c.get(url, &contents); err != nil {
		return nil, err
	}
	return contents, nil
}

func (c *httpGitHubClient) GetFileContent(owner, repo, filePath string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.baseURL, owner, repo, filePath)

	// For file content, use the raw media type to get raw bytes directly.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request %s: %w", url, err)
	}
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %w", url, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found: %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API %s returned status %d", url, resp.StatusCode)
	}

	return body, nil
}

// knowFrontmatter mirrors the fields needed from a .know/ file's frontmatter.
// Only the fields present in know.Meta are extracted; unrecognized fields are ignored.
type knowFrontmatter struct {
	Domain        string  `yaml:"domain"`
	GeneratedAt   string  `yaml:"generated_at"`
	ExpiresAfter  string  `yaml:"expires_after"`
	SourceHash    string  `yaml:"source_hash"`
	Confidence    float64 `yaml:"confidence"`
	FormatVersion string  `yaml:"format_version"`
}

// SyncRegistry discovers repos for the org and catalogs all .know/ domains.
// Strategy:
//  1. If the org context has repos configured (from org.yaml), use those.
//  2. Otherwise, discover repos via the GitHub API.
//
// For each repo, lists .know/*.md, .know/feat/*.md, and .know/release/*.md,
// parses their frontmatter, and builds DomainEntry records.
func SyncRegistry(ctx OrgContext, client GitHubClient) (*DomainCatalog, error) {
	catalog := NewCatalog(ctx)

	var repoConfigs []config.RepoConfig
	if repos := ctx.Repos(); len(repos) > 0 {
		// Explicit repo list from org.yaml takes priority.
		repoConfigs = repos
	} else {
		// Fall back to GitHub API discovery.
		ghRepos, err := client.ListOrgRepos(ctx.Name())
		if err != nil {
			return nil, fmt.Errorf("discover repos for org %s: %w", ctx.Name(), err)
		}
		for _, r := range ghRepos {
			// Skip archived and forked repos for knowledge discovery.
			if r.Archived || r.Fork {
				continue
			}
			repoConfigs = append(repoConfigs, config.RepoConfig{
				Name:          r.Name,
				URL:           r.HTMLURL,
				DefaultBranch: r.DefaultBranch,
			})
		}
	}

	for _, rc := range repoConfigs {
		repoEntry, err := syncRepo(ctx.Name(), rc, client)
		if err != nil {
			// Non-fatal: a single repo failure does not abort the whole sync.
			// Log to stderr is the project convention for non-user-visible warnings.
			fmt.Printf("warn: sync repo %s/%s: %v\n", ctx.Name(), rc.Name, err)
			// Include a partial entry so the catalog shows the repo was attempted.
			repoEntry = RepoEntry{
				Name:          rc.Name,
				URL:           rc.URL,
				DefaultBranch: rc.DefaultBranch,
				LastSynced:    "",
				Domains:       nil,
			}
		}
		catalog.Repos = append(catalog.Repos, repoEntry)
	}

	catalog.SyncedAt = time.Now().UTC().Format(time.RFC3339)
	return catalog, nil
}

// syncRepo syncs a single repo's .know/ domains.
func syncRepo(orgName string, rc config.RepoConfig, client GitHubClient) (RepoEntry, error) {
	entry := RepoEntry{
		Name:          rc.Name,
		URL:           rc.URL,
		DefaultBranch: rc.DefaultBranch,
	}

	// Discover .know/ paths to scan: root, feat/, release/
	knowPaths := []string{".know"}

	for _, knowBase := range knowPaths {
		contents, err := client.ListDirectoryContents(orgName, rc.Name, knowBase)
		if err != nil {
			// .know/ directory not present — not an error, just no domains here.
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			return entry, fmt.Errorf("list %s in %s: %w", knowBase, rc.Name, err)
		}

		for _, item := range contents {
			switch item.Type {
			case "file":
				if !strings.HasSuffix(item.Name, ".md") {
					continue
				}
				domain, err := fetchDomain(orgName, rc, item.Path, client)
				if err != nil {
					// Non-fatal: skip unparseable files.
					continue
				}
				entry.Domains = append(entry.Domains, domain)

			case "dir":
				// Scan one level deep for feat/ and release/ subdirectories.
				if item.Name == "feat" || item.Name == "release" {
					subContents, err := client.ListDirectoryContents(orgName, rc.Name, item.Path)
					if err != nil {
						continue
					}
					for _, sub := range subContents {
						if sub.Type != "file" || !strings.HasSuffix(sub.Name, ".md") {
							continue
						}
						domain, err := fetchDomain(orgName, rc, sub.Path, client)
						if err != nil {
							continue
						}
						entry.Domains = append(entry.Domains, domain)
					}
				}
			}
		}
	}

	entry.LastSynced = time.Now().UTC().Format(time.RFC3339)
	return entry, nil
}

// fetchDomain fetches and parses a single .know/ file, building a DomainEntry.
func fetchDomain(orgName string, rc config.RepoConfig, filePath string, client GitHubClient) (DomainEntry, error) {
	content, err := client.GetFileContent(orgName, rc.Name, filePath)
	if err != nil {
		return DomainEntry{}, fmt.Errorf("fetch %s: %w", filePath, err)
	}

	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return DomainEntry{}, fmt.Errorf("parse frontmatter in %s: %w", filePath, err)
	}

	// Compute domain name: use frontmatter domain if set; otherwise derive from path.
	domainName := frontmatter.Domain
	if domainName == "" {
		// Derive from file path: ".know/feat/materialization.md" -> "feat/materialization"
		baseName := strings.TrimSuffix(path.Base(filePath), ".md")
		dirName := path.Dir(filePath) // e.g., ".know/feat"
		knowDir := ".know"
		if dirName != knowDir {
			rel := strings.TrimPrefix(dirName, knowDir+"/")
			domainName = rel + "/" + baseName
		} else {
			domainName = baseName
		}
	}

	qdn := orgName + "::" + rc.Name + "::" + domainName

	return DomainEntry{
		QualifiedName: qdn,
		Domain:        domainName,
		Path:          filePath,
		GeneratedAt:   frontmatter.GeneratedAt,
		ExpiresAfter:  frontmatter.ExpiresAfter,
		SourceHash:    frontmatter.SourceHash,
		Confidence:    frontmatter.Confidence,
		FormatVersion: frontmatter.FormatVersion,
	}, nil
}

// extractFrontmatter parses YAML frontmatter from a markdown file.
// Frontmatter is delimited by "---\n" lines.
func extractFrontmatter(data []byte) (knowFrontmatter, error) {
	content := strings.TrimSpace(string(data))

	if !strings.HasPrefix(content, "---") {
		return knowFrontmatter{}, fmt.Errorf("no frontmatter found")
	}

	// Strip the opening "---"
	rest := strings.TrimPrefix(content, "---")
	rest = strings.TrimPrefix(rest, "\n")

	// Find the closing "---"
	end := strings.Index(rest, "\n---")
	if end == -1 {
		// Try without a trailing newline (some files end file with just "---")
		if strings.HasSuffix(strings.TrimSpace(rest), "---") {
			idx := strings.LastIndex(rest, "---")
			rest = rest[:idx]
		} else {
			return knowFrontmatter{}, fmt.Errorf("frontmatter closing delimiter not found")
		}
	} else {
		rest = rest[:end]
	}

	var fm knowFrontmatter
	if err := yaml.Unmarshal([]byte(rest), &fm); err != nil {
		return knowFrontmatter{}, fmt.Errorf("unmarshal frontmatter: %w", err)
	}

	return fm, nil
}
