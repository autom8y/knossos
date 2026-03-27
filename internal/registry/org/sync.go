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
	"github.com/autom8y/knossos/internal/know"
)

// GitHubClient abstracts GitHub API calls for testability.
type GitHubClient interface {
	ListOrgRepos(org string) ([]GitHubRepo, error)
	ListDirectoryContents(owner, repo, dirPath string) ([]GitHubContent, error)
	GetFileContent(owner, repo, filePath string) ([]byte, error)
	// GetTree returns the full recursive tree for a repo at the given SHA/branch.
	// Returns nil, nil if the tree endpoint is not available or returns an error.
	GetTree(owner, repo, sha string) ([]GitHubTreeEntry, error)
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

// GitHubTreeEntry represents a single entry in a GitHub recursive tree response.
type GitHubTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"` // "blob" or "tree"
	SHA  string `json:"sha"`
}

// gitHubTreeResponse is the response from the Git Trees API.
type gitHubTreeResponse struct {
	SHA       string            `json:"sha"`
	Tree      []GitHubTreeEntry `json:"tree"`
	Truncated bool              `json:"truncated"`
}

// httpGitHubClient implements GitHubClient using stdlib net/http.
type httpGitHubClient struct {
	httpClient *http.Client
	token      string
	baseURL    string // default: "https://api.github.com"
}

// NewGitHubClient creates a GitHubClient using the given http.Client and token.
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

func (c *httpGitHubClient) GetTree(owner, repo, sha string) ([]GitHubTreeEntry, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1", c.baseURL, owner, repo, sha)
	var resp gitHubTreeResponse
	if err := c.get(url, &resp); err != nil {
		return nil, err
	}
	return resp.Tree, nil
}

// MaxScopeDepth is the maximum directory depth for .know/ discovery.
// The autom8y monorepo's deepest .know/ is at depth 3 (services/ads/.know/).
const MaxScopeDepth = 5

// knowFrontmatter mirrors the fields needed from a .know/ file's frontmatter.
type knowFrontmatter struct {
	Domain        string  `yaml:"domain"`
	GeneratedAt   string  `yaml:"generated_at"`
	ExpiresAfter  string  `yaml:"expires_after"`
	SourceHash    string  `yaml:"source_hash"`
	Confidence    float64 `yaml:"confidence"`
	FormatVersion string  `yaml:"format_version"`
}

// SyncRegistry discovers repos for the org and catalogs all .know/ domains.
// Discovers .know/ directories at any depth within each repo using the recursive
// tree API (single API call per repo).
func SyncRegistry(ctx OrgContext, client GitHubClient) (*DomainCatalog, error) {
	catalog := NewCatalog(ctx)

	var repoConfigs []config.RepoConfig
	if repos := ctx.Repos(); len(repos) > 0 {
		repoConfigs = repos
	} else {
		ghRepos, err := client.ListOrgRepos(ctx.Name())
		if err != nil {
			return nil, fmt.Errorf("discover repos for org %s: %w", ctx.Name(), err)
		}
		for _, r := range ghRepos {
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
			fmt.Printf("warn: sync repo %s/%s: %v\n", ctx.Name(), rc.Name, err)
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

// syncRepo syncs a single repo's .know/ domains using recursive tree discovery.
func syncRepo(orgName string, rc config.RepoConfig, client GitHubClient) (RepoEntry, error) {
	entry := RepoEntry{
		Name:          rc.Name,
		URL:           rc.URL,
		DefaultBranch: rc.DefaultBranch,
	}

	// Discover all .know/ directory paths using recursive tree API.
	knowPaths := discoverKnowPaths(rc, client)

	// For each .know/ directory, scan its contents.
	for _, kp := range knowPaths {
		scope := scopeFromKnowPath(kp)
		domains, err := scanKnowDir(orgName, rc, kp, scope, client)
		if err != nil {
			// Non-fatal: skip this .know/ directory.
			continue
		}
		entry.Domains = append(entry.Domains, domains...)
	}

	entry.LastSynced = time.Now().UTC().Format(time.RFC3339)
	return entry, nil
}

// discoverKnowPaths finds all .know/ directories in a repo.
// Uses the recursive tree API (single API call) then filters client-side.
// Falls back to checking just root .know/ if tree API fails.
func discoverKnowPaths(rc config.RepoConfig, client GitHubClient) []string {
	branch := rc.DefaultBranch
	if branch == "" {
		branch = "main"
	}

	tree, err := client.GetTree(rc.Name, rc.Name, branch)
	if err != nil {
		// Fallback: just try root .know/
		return []string{".know"}
	}

	var paths []string
	for _, entry := range tree {
		if entry.Type != "tree" {
			continue
		}
		// Match directories named ".know" at any depth within MaxScopeDepth.
		if path.Base(entry.Path) != ".know" {
			continue
		}
		depth := strings.Count(entry.Path, "/")
		if depth > MaxScopeDepth {
			continue
		}
		// Exclude common vendor/generated directories.
		if shouldExcludePath(entry.Path) {
			continue
		}
		paths = append(paths, entry.Path)
	}

	if len(paths) == 0 {
		// Still try root .know/ in case tree was empty/truncated.
		return []string{".know"}
	}
	return paths
}

// shouldExcludePath returns true for paths that should never be scanned for .know/.
func shouldExcludePath(p string) bool {
	excludePrefixes := []string{
		"vendor/",
		"node_modules/",
		".git/",
		".terraform/",
		".knossos/worktrees/",
	}
	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(p, prefix) || strings.Contains(p, "/"+prefix) {
			return true
		}
	}
	return false
}

// scopeFromKnowPath derives the scope from a .know/ directory path.
// ".know" -> "" (root scope)
// "services/ads/.know" -> "services/ads"
// "sdks/python/autom8y-meta/.know" -> "sdks/python/autom8y-meta"
func scopeFromKnowPath(knowPath string) string {
	trimmed := strings.TrimSuffix(knowPath, "/.know")
	if trimmed == knowPath || trimmed == "" {
		// It was just ".know" (root) or empty after trimming.
		return ""
	}
	return trimmed
}

// scanKnowDir scans a single .know/ directory and its feat/ and release/ subdirs.
func scanKnowDir(orgName string, rc config.RepoConfig, knowPath, scope string, client GitHubClient) ([]DomainEntry, error) {
	var domains []DomainEntry

	contents, err := client.ListDirectoryContents(orgName, rc.Name, knowPath)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("list %s in %s: %w", knowPath, rc.Name, err)
	}

	for _, item := range contents {
		switch item.Type {
		case "file":
			if !strings.HasSuffix(item.Name, ".md") {
				continue
			}
			domain, err := fetchDomain(orgName, rc, item.Path, scope, client)
			if err != nil {
				continue
			}
			domains = append(domains, domain)

		case "dir":
			if item.Name == "feat" || item.Name == "release" {
				subContents, err := client.ListDirectoryContents(orgName, rc.Name, item.Path)
				if err != nil {
					continue
				}
				for _, sub := range subContents {
					if sub.Type != "file" || !strings.HasSuffix(sub.Name, ".md") {
						continue
					}
					domain, err := fetchDomain(orgName, rc, sub.Path, scope, client)
					if err != nil {
						continue
					}
					domains = append(domains, domain)
				}
			}
		}
	}

	return domains, nil
}

// fetchDomain fetches and parses a single .know/ file, building a DomainEntry.
func fetchDomain(orgName string, rc config.RepoConfig, filePath, scope string, client GitHubClient) (DomainEntry, error) {
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
		domainName = deriveDomainName(filePath, scope)
	}

	qdn := know.NewScoped(orgName, rc.Name, scope, domainName)

	return DomainEntry{
		QualifiedName: qdn.String(),
		Domain:        domainName,
		Scope:         scope,
		Path:          filePath,
		GeneratedAt:   frontmatter.GeneratedAt,
		ExpiresAfter:  frontmatter.ExpiresAfter,
		SourceHash:    frontmatter.SourceHash,
		Confidence:    frontmatter.Confidence,
		FormatVersion: frontmatter.FormatVersion,
	}, nil
}

// deriveDomainName computes the domain name from a file path when frontmatter lacks it.
// For ".know/architecture.md" -> "architecture"
// For ".know/feat/materialization.md" -> "feat/materialization"
// For "services/ads/.know/architecture.md" -> "architecture"
// For "services/ads/.know/feat/materialization.md" -> "feat/materialization"
func deriveDomainName(filePath, scope string) string {
	// Strip scope prefix to get the .know/-relative path.
	knowRelative := filePath
	if scope != "" {
		knowRelative = strings.TrimPrefix(filePath, scope+"/")
	}
	// Strip ".know/" prefix.
	knowRelative = strings.TrimPrefix(knowRelative, ".know/")
	// Strip .md extension.
	return strings.TrimSuffix(knowRelative, ".md")
}

// extractFrontmatter parses YAML frontmatter from a markdown file.
func extractFrontmatter(data []byte) (knowFrontmatter, error) {
	content := strings.TrimSpace(string(data))

	if !strings.HasPrefix(content, "---") {
		return knowFrontmatter{}, fmt.Errorf("no frontmatter found")
	}

	rest := strings.TrimPrefix(content, "---")
	rest = strings.TrimPrefix(rest, "\n")

	end := strings.Index(rest, "\n---")
	if end == -1 {
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
