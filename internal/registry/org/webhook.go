package org

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/config"
)

// pushEventPayload is the minimal subset of GitHub's push webhook payload
// needed to detect .know/ file changes.
type pushEventPayload struct {
	Repository struct {
		Name          string `json:"name"`
		HTMLURL       string `json:"html_url"`
		DefaultBranch string `json:"default_branch"`
	} `json:"repository"`
	Commits []struct {
		Added    []string `json:"added"`
		Removed  []string `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"commits"`
}

// HandlePushEvent processes a GitHub push webhook payload and incrementally
// updates the catalog for any repo whose .know/ files changed.
// If no .know/ files changed in the push, the catalog is returned unchanged.
func HandlePushEvent(payload []byte, catalog *DomainCatalog, client GitHubClient) error {
	var event pushEventPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("parse push event payload: %w", err)
	}

	repoName := event.Repository.Name
	if repoName == "" {
		return fmt.Errorf("push event missing repository.name")
	}

	// Collect all changed file paths from the push.
	var changedFiles []string
	for _, commit := range event.Commits {
		changedFiles = append(changedFiles, commit.Added...)
		changedFiles = append(changedFiles, commit.Modified...)
		changedFiles = append(changedFiles, commit.Removed...)
	}

	// Check if any .know/ files were affected.
	knowChanged := slices.ContainsFunc(changedFiles, isKnowFile)

	if !knowChanged {
		// No .know/ changes — catalog is current.
		return nil
	}

	// Re-sync the affected repo only.
	rc := config.RepoConfig{
		Name:          repoName,
		URL:           event.Repository.HTMLURL,
		DefaultBranch: event.Repository.DefaultBranch,
	}

	updatedRepo, err := syncRepo(catalog.Org, rc, client)
	if err != nil {
		return fmt.Errorf("re-sync repo %s: %w", repoName, err)
	}

	// Replace or insert the repo entry in the catalog.
	found := false
	for i, r := range catalog.Repos {
		if r.Name == repoName {
			catalog.Repos[i] = updatedRepo
			found = true
			break
		}
	}
	if !found {
		catalog.Repos = append(catalog.Repos, updatedRepo)
	}

	catalog.SyncedAt = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// isKnowFile returns true if the file path is inside a .know/ directory.
func isKnowFile(filePath string) bool {
	return strings.HasPrefix(filePath, ".know/") || filePath == ".know"
}
