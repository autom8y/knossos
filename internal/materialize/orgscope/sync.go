// Package orgscope implements the org-scope sync pipeline for the materialize
// system. It syncs resources from an org directory to ~/.claude/.
package orgscope

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// SyncOrgScopeParams provides all dependencies for org-scope sync.
type SyncOrgScopeParams struct {
	OrgName      string // Explicit org name (empty = use config.ActiveOrg())
	OrgDir       string // Override org data directory (for testing; empty = use paths.OrgDataDir)
	UserClaudeDir string // Override user .claude directory (for testing; empty = use paths.UserClaudeDir)
	DryRun       bool
}

// OrgScopeResult wraps org scope sync outcome.
type OrgScopeResult struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	OrgName string `json:"org_name,omitempty"`
	Source  string `json:"source,omitempty"`
	Agents  int    `json:"agents,omitempty"`
	Mena    int    `json:"mena,omitempty"`
}

// SyncOrgScope is the primary entry point for org-scope sync.
// It syncs agents and mena from an org directory to ~/.claude/.
func SyncOrgScope(params SyncOrgScopeParams) (*OrgScopeResult, error) {
	orgName := params.OrgName
	if orgName == "" {
		orgName = config.ActiveOrg()
	}
	if orgName == "" {
		return &OrgScopeResult{
			Status: "skipped",
			Error:  "no active org configured",
		}, nil
	}

	orgDir := params.OrgDir
	if orgDir == "" {
		orgDir = paths.OrgDataDir(orgName)
	}
	if _, err := os.Stat(orgDir); os.IsNotExist(err) {
		return &OrgScopeResult{
			Status:  "skipped",
			OrgName: orgName,
			Error:   "org directory does not exist: " + orgDir,
		}, nil
	}

	userClaudeDir := params.UserClaudeDir
	if userClaudeDir == "" {
		userClaudeDir = paths.UserClaudeDir()
	}

	// Load or bootstrap ORG_PROVENANCE_MANIFEST.yaml
	manifestPath := provenance.OrgManifestPath(userClaudeDir)
	manifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		return nil, err
	}

	result := &OrgScopeResult{
		Status:  "success",
		OrgName: orgName,
		Source:  orgDir,
	}

	// Sync agents
	agentsDir := filepath.Join(orgDir, "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		count, err := syncOrgResource(agentsDir, filepath.Join(userClaudeDir, "agents"), manifest, params.DryRun)
		if err != nil {
			log.Printf("orgscope: error syncing agents: %v", err)
		}
		result.Agents = count
	}

	// Sync mena (commands + skills)
	menaDir := filepath.Join(orgDir, "mena")
	if _, err := os.Stat(menaDir); err == nil {
		count, err := syncOrgResource(menaDir, filepath.Join(userClaudeDir, "skills"), manifest, params.DryRun)
		if err != nil {
			log.Printf("orgscope: error syncing mena: %v", err)
		}
		result.Mena = count
	}

	// Save manifest if not dry-run
	if !params.DryRun {
		manifest.LastSync = time.Now().UTC()
		if err := provenance.Save(manifestPath, manifest); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// syncOrgResource copies files from an org source directory to a target directory,
// tracking provenance. Returns the count of files synced.
func syncOrgResource(sourceDir, targetDir string, manifest *provenance.ProvenanceManifest, dryRun bool) (int, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return 0, err
	}

	if !dryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return 0, err
		}
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Flat files only for now
		}
		name := entry.Name()
		sourcePath := filepath.Join(sourceDir, name)
		targetPath := filepath.Join(targetDir, name)

		sourceData, err := os.ReadFile(sourcePath)
		if err != nil {
			log.Printf("orgscope: failed to read %s: %v", sourcePath, err)
			continue
		}

		sourceChecksum := checksum.Bytes(sourceData)

		// Check if target already exists and is org-owned with same checksum
		relPath := name
		if existing, ok := manifest.Entries[relPath]; ok {
			if existing.Scope == provenance.ScopeOrg && existing.Checksum == sourceChecksum {
				continue // Unchanged
			}
		}

		if dryRun {
			count++
			continue
		}

		if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
			log.Printf("orgscope: failed to write %s: %v", targetPath, err)
			continue
		}

		// Track in provenance manifest
		manifest.Entries[relPath] = provenance.NewKnossosEntry(
			provenance.ScopeOrg,
			sourcePath,
			"org",
			sourceChecksum,
		)
		count++
	}

	return count, nil
}
