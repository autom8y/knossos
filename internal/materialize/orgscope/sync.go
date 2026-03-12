// Package orgscope implements the org-scope sync pipeline for the materialize
// system. It syncs resources from an org directory to ~/.claude/.
package orgscope

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/materialize/compiler"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// SyncOrgScopeParams provides all dependencies for org-scope sync.
type SyncOrgScopeParams struct {
	OrgName		string	// Explicit org name (empty = use config.ActiveOrg())
	OrgDir		string	// Override org data directory (for testing; empty = use paths.OrgDataDir)
	UserChannelDir	string	// Override user .claude directory (for testing; empty = use paths.UserChannelDir)
	DryRun		bool
	Channel		string	// Target channel: "claude" (default) or "gemini"
}

// OrgScopeResult wraps org scope sync outcome.
type OrgScopeResult struct {
	Status	string	`json:"status"`
	Error	string	`json:"error,omitempty"`
	OrgName	string	`json:"org_name,omitempty"`
	Source	string	`json:"source,omitempty"`
	Agents	int	`json:"agents,omitempty"`
	Mena	int	`json:"mena,omitempty"`
}

// SyncOrgScope is the convenience entry point for org-scope sync.
// It resolves org name from config.ActiveOrg() when not provided in params,
// then delegates to syncOrgScopeResolved.
func SyncOrgScope(params SyncOrgScopeParams) (*OrgScopeResult, error) {
	if params.OrgName == "" {
		params.OrgName = config.ActiveOrg()
	}
	return syncOrgScopeResolved(params)
}

// syncOrgScopeResolved performs org-scope sync with all params resolved.
// Requires OrgName to be explicitly set (empty means no org → skip).
// Tests use this directly to avoid config.ActiveOrg() side effects.
func syncOrgScopeResolved(params SyncOrgScopeParams) (*OrgScopeResult, error) {
	orgName := params.OrgName
	if orgName == "" {
		return &OrgScopeResult{
			Status:	"skipped",
			Error:	"no active org configured",
		}, nil
	}

	orgDir := params.OrgDir
	if orgDir == "" {
		orgDir = paths.OrgDataDir(orgName)
	}
	if _, err := os.Stat(orgDir); os.IsNotExist(err) {
		return &OrgScopeResult{
			Status:		"skipped",
			OrgName:	orgName,
			Error:		"org directory does not exist: " + orgDir,
		}, nil
	}

	userChannelDir := params.UserChannelDir
	if userChannelDir == "" {
		userChannelDir = paths.UserChannelDir(params.Channel)
	}

	// Load or bootstrap ORG_PROVENANCE_MANIFEST.yaml
	manifestPath := provenance.OrgManifestPath(userChannelDir)
	manifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		return nil, err
	}

	result := &OrgScopeResult{
		Status:		"success",
		OrgName:	orgName,
		Source:		orgDir,
	}

	// Sync agents
	agentsDir := filepath.Join(orgDir, "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		count, err := syncOrgResource(agentsDir, filepath.Join(userChannelDir, "agents"), manifest, params.DryRun, params.Channel)
		if err != nil {
			slog.Warn("orgscope: error syncing agents", "error", err)
		}
		result.Agents = count
	}

	// Sync mena (commands + skills)
	menaDir := filepath.Join(orgDir, "mena")
	if _, err := os.Stat(menaDir); err == nil {
		count, err := syncOrgResource(menaDir, filepath.Join(userChannelDir, "skills"), manifest, params.DryRun, "")
		if err != nil {
			slog.Warn("orgscope: error syncing mena", "error", err)
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
// tracking provenance. When channel is "gemini" and files are agent markdown files,
// applies GeminiCompiler transformation before writing.
// The sourceChecksum is always computed from the original source so subsequent
// syncs correctly detect source changes.
// Returns the count of files synced.
func syncOrgResource(sourceDir, targetDir string, manifest *provenance.ProvenanceManifest, dryRun bool, channel string) (int, error) {
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
			continue	// Flat files only for now
		}
		name := entry.Name()
		sourcePath := filepath.Join(sourceDir, name)
		targetPath := filepath.Join(targetDir, name)

		sourceData, err := os.ReadFile(sourcePath)
		if err != nil {
			slog.Warn("orgscope: failed to read source", "path", sourcePath, "error", err)
			continue
		}

		// Provenance checksum always tracks the source file, not the compiled output.
		sourceChecksum := checksum.Bytes(sourceData)

		// Check if target already exists and is org-owned with same checksum
		relPath := name
		if existing, ok := manifest.Entries[relPath]; ok {
			if existing.Scope == provenance.ScopeOrg && existing.Checksum == sourceChecksum {
				continue	// Unchanged
			}
		}

		if dryRun {
			count++
			continue
		}

		// Apply gemini compilation for agent markdown files
		writeData := sourceData
		if channel == "gemini" {
			compiled, compileErr := compileOrgAgentForGemini(name, sourceData)
			if compileErr == nil {
				writeData = compiled
			} else {
				slog.Warn("orgscope: agent gemini compile failed, using raw source", "path", sourcePath, "error", compileErr)
			}
		}

		if err := os.WriteFile(targetPath, writeData, 0644); err != nil {
			slog.Warn("orgscope: failed to write target", "path", targetPath, "error", err)
			continue
		}

		// Track in provenance manifest
		manifest.Entries[relPath] = provenance.NewKnossosEntry(
			provenance.ScopeOrg,
			sourcePath,
			"org",
			sourceChecksum, "",
		)
		count++
	}

	return count, nil
}

// compileOrgAgentForGemini parses an agent file's frontmatter and applies the
// GeminiCompiler transformation (tool name translation + CC key stripping).
// Only applies to markdown files — other file types are returned unchanged.
func compileOrgAgentForGemini(name string, content []byte) ([]byte, error) {
	if filepath.Ext(name) != ".md" {
		return content, nil
	}
	yamlBytes, body, err := frontmatter.Parse(content)
	if err != nil {
		// No frontmatter — pass through unchanged
		return content, nil
	}
	var fmMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &fmMap); err != nil {
		return nil, err
	}
	if fmMap == nil {
		fmMap = make(map[string]any)
	}
	return (&compiler.GeminiCompiler{}).CompileAgent(name, fmMap, string(body))
}
