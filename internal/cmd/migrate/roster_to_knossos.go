package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/spf13/cobra"
)

type rosterToKnossosOptions struct {
	apply          bool
	backup         bool
	noBackup       bool
	generateScript bool
	scriptFile     string
	projectDir     string
	skipProject    bool
	skipUser       bool
}

// EnvVarMapping records a ROSTER_* environment variable mapping.
type EnvVarMapping struct {
	Old   string // e.g., "ROSTER_HOME"
	New   string // e.g., "KNOSSOS_HOME"
	Value string // current value
}

// newRosterToKnossosCmd creates the roster-to-knossos subcommand.
func newRosterToKnossosCmd(ctx *cmdContext) *cobra.Command {
	var opts rosterToKnossosOptions

	cmd := &cobra.Command{
		Use:   "roster-to-knossos",
		Short: "Migrate satellite manifests from roster to knossos branding",
		Long: `Migrates satellite manifests from "roster" to "knossos" branding.

Rewrites source fields in USER_*_MANIFEST.json files and CEM manifest
metadata keys. Safe to run multiple times (idempotent).

By default, runs in dry-run mode. Use --apply to execute changes.

Targets:
  User manifests   ~/.claude/USER_{AGENT,SKILL,COMMAND,HOOKS}_MANIFEST.json
  CEM manifest     .claude/.cem/manifest.json (project-level)
  Env variables    Advisory for ROSTER_* environment variables

Examples:
  ari migrate roster-to-knossos                     # Preview all changes
  ari migrate roster-to-knossos --apply             # Execute migration
  ari migrate roster-to-knossos --generate-script   # Output env var migration script
  ari migrate roster-to-knossos --apply --no-backup # Migrate without backups
  ari migrate roster-to-knossos --skip-project      # Only migrate user manifests`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRosterToKnossos(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.apply, "apply", false, "Execute the migration (sets dry-run to false)")
	cmd.Flags().BoolVar(&opts.backup, "backup", true, "Create backup files before rewriting")
	cmd.Flags().BoolVar(&opts.noBackup, "no-backup", false, "Skip backup creation")
	cmd.Flags().BoolVar(&opts.generateScript, "generate-script", false, "Output shell profile migration script to stdout")
	cmd.Flags().StringVar(&opts.scriptFile, "script-file", "", "Write shell profile migration script to file instead of stdout")
	cmd.Flags().StringVar(&opts.projectDir, "project", "", "Target project directory for CEM manifest")
	cmd.Flags().BoolVar(&opts.skipProject, "skip-project", false, "Skip CEM manifest migration")
	cmd.Flags().BoolVar(&opts.skipUser, "skip-user", false, "Skip user-level manifest migration")

	return cmd
}

// runRosterToKnossos executes the migration.
func runRosterToKnossos(ctx *cmdContext, opts rosterToKnossosOptions) error {
	printer := ctx.getPrinter()

	// Validate flag combinations
	if opts.skipUser && opts.skipProject {
		err := errors.New(errors.CodeUsageError, "cannot skip both user and project migrations (nothing to migrate)")
		printer.PrintError(err)
		return err
	}

	if opts.skipProject && opts.projectDir != "" {
		err := errors.New(errors.CodeUsageError, "cannot specify both --skip-project and --project")
		printer.PrintError(err)
		return err
	}

	// Determine effective flags
	dryRun := !opts.apply
	backup := opts.backup && !opts.noBackup

	// Initialize result
	result := output.RosterMigrateOutput{
		DryRun: dryRun,
	}

	// Migrate user manifests
	if !opts.skipUser {
		userResults, err := migrateUserManifests(dryRun, backup)
		if err != nil {
			result.Errors = append(result.Errors, "User manifest migration: "+err.Error())
		}
		result.UserManifests = userResults
		for _, r := range userResults {
			result.ManifestsFound++
			if !r.Skipped {
				result.ManifestsChanged++
				result.EntriesRewritten += r.EntriesRewritten
				if r.BackupPath != "" {
					result.BackupsCreated = append(result.BackupsCreated, r.BackupPath)
				}
			} else {
				result.ManifestsSkipped++
			}
		}
	}

	// Migrate CEM manifest
	if !opts.skipProject {
		// Resolve project dir
		projectDir := opts.projectDir
		if projectDir == "" {
			if ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
				projectDir = *ctx.ProjectDir
			} else {
				// Try cwd
				cwd, _ := os.Getwd()
				projectDir = cwd
			}
		}

		cemResult, err := migrateCEMManifest(projectDir, dryRun, backup)
		if err != nil {
			result.Errors = append(result.Errors, "CEM manifest migration: "+err.Error())
		}
		if cemResult != nil {
			result.CEMManifest = cemResult
			result.ManifestsFound++
			if !cemResult.Skipped {
				result.ManifestsChanged++
				result.EntriesRewritten += cemResult.EntriesRewritten
				if cemResult.BackupPath != "" {
					result.BackupsCreated = append(result.BackupsCreated, cemResult.BackupPath)
				}
			} else {
				result.ManifestsSkipped++
			}
		}
	}

	// Scan environment variables
	envVars := scanRosterEnvVars()
	for _, ev := range envVars {
		result.EnvVarsDetected = append(result.EnvVarsDetected, output.EnvVarDetected{
			Current: ev.Old,
			Replace: ev.New,
			Value:   ev.Value,
		})
	}

	// Generate migration script
	if opts.generateScript || opts.scriptFile != "" {
		script := generateMigrationScript(envVars)
		if opts.scriptFile != "" {
			if err := os.WriteFile(opts.scriptFile, []byte(script), 0755); err != nil {
				result.Errors = append(result.Errors, "Failed to write script: "+err.Error())
			} else {
				result.ScriptGenerated = true
				result.ScriptPath = opts.scriptFile
			}
		} else {
			// Output to stdout
			printer.PrintLine(script)
			result.ScriptGenerated = true
		}
	}

	return printer.Print(result)
}

// rewriteUserManifestBytes rewrites source fields in a user manifest JSON blob.
// Returns the rewritten JSON bytes and the count of entries changed.
// Returns the input unchanged if no roster references found (idempotent).
func rewriteUserManifestBytes(data []byte) ([]byte, int, error) {
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, 0, err
	}

	count := 0

	// Iterate over resource type keys (agents, skills, commands, hooks)
	for _, resourceVal := range manifest {
		if resourceMap, ok := resourceVal.(map[string]interface{}); ok {
			// For each entry in the resource map
			for _, entryVal := range resourceMap {
				if entryMap, ok := entryVal.(map[string]interface{}); ok {
					if source, ok := entryMap["source"].(string); ok {
						if source == "roster" {
							entryMap["source"] = "knossos"
							count++
						} else if source == "roster-diverged" {
							entryMap["source"] = "knossos-diverged"
							count++
						}
					}
				}
			}
		}
	}

	if count == 0 {
		return data, 0, nil
	}

	// Marshal back with formatting
	rewritten, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, 0, err
	}

	// Add trailing newline
	rewritten = append(rewritten, '\n')

	return rewritten, count, nil
}

// rewriteCEMManifestBytes rewrites roster references in a CEM manifest JSON blob.
// Returns the rewritten JSON bytes and the count of fields changed.
func rewriteCEMManifestBytes(data []byte) ([]byte, int, error) {
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, 0, err
	}

	count := 0

	// Rename top-level "roster" key to "knossos"
	if rosterVal, ok := manifest["roster"]; ok {
		manifest["knossos"] = rosterVal
		delete(manifest, "roster")
		count++
	}

	// Rename "team.roster_path" to "team.knossos_path"
	if teamVal, ok := manifest["team"].(map[string]interface{}); ok {
		if rosterPathVal, ok := teamVal["roster_path"]; ok {
			teamVal["knossos_path"] = rosterPathVal
			delete(teamVal, "roster_path")
			count++
		}
	}

	// Update "managed_files" entries
	if managedFilesVal, ok := manifest["managed_files"].([]interface{}); ok {
		for _, fileVal := range managedFilesVal {
			if fileMap, ok := fileVal.(map[string]interface{}); ok {
				if source, ok := fileMap["source"].(string); ok {
					if source == "roster" {
						fileMap["source"] = "knossos"
						count++
					}
				}
			}
		}
	}

	if count == 0 {
		return data, 0, nil
	}

	// Marshal back with formatting
	rewritten, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, 0, err
	}

	// Add trailing newline
	rewritten = append(rewritten, '\n')

	return rewritten, count, nil
}

// scanRosterEnvVars returns a list of ROSTER_* environment variables currently set.
func scanRosterEnvVars() []EnvVarMapping {
	var mappings []EnvVarMapping

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(key, "ROSTER_") {
			newKey := strings.Replace(key, "ROSTER_", "KNOSSOS_", 1)
			mappings = append(mappings, EnvVarMapping{
				Old:   key,
				New:   newKey,
				Value: value,
			})
		}
	}

	return mappings
}

// generateMigrationScript produces a shell script that updates shell profile env vars.
func generateMigrationScript(envVars []EnvVarMapping) string {
	var b strings.Builder

	b.WriteString("#!/bin/bash\n")
	b.WriteString("# Generated by: ari migrate roster-to-knossos\n")
	b.WriteString("# This script updates your shell profile to replace ROSTER_* env vars\n")
	b.WriteString("# with their KNOSSOS_* equivalents. Review before running.\n")
	b.WriteString("#\n")

	if len(envVars) == 0 {
		b.WriteString("# No ROSTER_* environment variables detected. Nothing to do.\n")
		return b.String()
	}

	// Detect profile files
	homeDir, _ := os.UserHomeDir()
	profiles := []string{}
	for _, profile := range []string{".zshrc", ".bashrc", ".bash_profile", ".profile"} {
		profilePath := filepath.Join(homeDir, profile)
		if _, err := os.Stat(profilePath); err == nil {
			profiles = append(profiles, profilePath)
		}
	}

	if len(profiles) == 0 {
		b.WriteString("# No shell profile files detected (.zshrc, .bashrc, etc.)\n")
		return b.String()
	}

	b.WriteString("# Detected profile files:\n")
	for _, profile := range profiles {
		b.WriteString(fmt.Sprintf("#   %s\n", profile))
	}
	b.WriteString("#\n")
	b.WriteString("# Detected variables:\n")
	for _, ev := range envVars {
		b.WriteString(fmt.Sprintf("#   %s -> %s\n", ev.Old, ev.New))
	}
	b.WriteString("\n")

	b.WriteString("set -euo pipefail\n\n")

	// Detect platform for sed syntax
	b.WriteString("# Detect platform for sed compatibility\n")
	b.WriteString("if [[ \"$(uname -s)\" == \"Darwin\" ]]; then\n")
	b.WriteString("  SED_INPLACE=('-i' '')\n")
	b.WriteString("else\n")
	b.WriteString("  SED_INPLACE=('-i')\n")
	b.WriteString("fi\n\n")

	// Back up profiles
	for _, profile := range profiles {
		b.WriteString(fmt.Sprintf("# Back up %s\n", profile))
		b.WriteString(fmt.Sprintf("cp %s %s.pre-knossos-migrate\n", profile, profile))
	}
	b.WriteString("\n")

	// Replace each variable
	for _, ev := range envVars {
		b.WriteString(fmt.Sprintf("# %s -> %s\n", ev.Old, ev.New))
		for _, profile := range profiles {
			b.WriteString(fmt.Sprintf("sed \"${SED_INPLACE[@]}\" 's/%s/%s/g' %s\n", ev.Old, ev.New, profile))
		}
		b.WriteString("\n")
	}

	// Completion message
	b.WriteString("echo \"Profile updated. Run: source ")
	if len(profiles) > 0 {
		b.WriteString(profiles[0])
	}
	b.WriteString("\"\n")
	b.WriteString("echo \"Backups saved with .pre-knossos-migrate extension\"\n")

	return b.String()
}

// migrateUserManifests discovers and rewrites all USER_*_MANIFEST.json files.
func migrateUserManifests(dryRun, backup bool) ([]output.ManifestMigResult, error) {
	var results []output.ManifestMigResult

	// Get user .claude directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	claudeDir := filepath.Join(homeDir, ".claude")

	// Check if .claude directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		// Not an error, just no manifests to migrate
		return results, nil
	}

	// Glob for USER_*_MANIFEST.json files
	pattern := filepath.Join(claudeDir, "USER_*_MANIFEST.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, manifestPath := range matches {
		result := migrateManifestFile(manifestPath, dryRun, backup, rewriteUserManifestBytes)
		results = append(results, result)
	}

	return results, nil
}

// migrateCEMManifest discovers and rewrites the CEM manifest in a project.
func migrateCEMManifest(projectDir string, dryRun, backup bool) (*output.ManifestMigResult, error) {
	if projectDir == "" {
		return nil, nil
	}

	// Construct CEM manifest path
	cemPath := filepath.Join(projectDir, ".claude", ".cem", "manifest.json")

	// Check if file exists
	if _, err := os.Stat(cemPath); os.IsNotExist(err) {
		// No CEM manifest to migrate
		return nil, nil
	}

	result := migrateManifestFile(cemPath, dryRun, backup, rewriteCEMManifestBytes)
	return &result, nil
}

// migrateManifestFile migrates a single manifest file using the provided rewrite function.
func migrateManifestFile(
	manifestPath string,
	dryRun, backup bool,
	rewriteFunc func([]byte) ([]byte, int, error),
) output.ManifestMigResult {
	result := output.ManifestMigResult{
		Path: manifestPath,
	}

	// Read file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Skipped = true
		result.SkipReason = "failed to read: " + err.Error()
		return result
	}

	// Rewrite
	rewritten, count, err := rewriteFunc(data)
	if err != nil {
		result.Skipped = true
		result.SkipReason = "failed to parse: " + err.Error()
		return result
	}

	// Check if already migrated
	if count == 0 {
		result.Skipped = true
		result.SkipReason = "already migrated"
		return result
	}

	result.EntriesRewritten = count

	// If dry-run, stop here
	if dryRun {
		return result
	}

	// Create backup if requested
	if backup {
		backupPath := manifestPath + ".roster-backup"
		// Check if backup already exists
		if _, err := os.Stat(backupPath); err == nil {
			// Backup exists, don't overwrite
			result.BackupPath = filepath.Base(backupPath) + " (exists)"
		} else {
			if err := os.WriteFile(backupPath, data, 0644); err != nil {
				result.Skipped = true
				result.SkipReason = "failed to create backup: " + err.Error()
				return result
			}
			result.BackupPath = filepath.Base(backupPath)
		}
	}

	// Write rewritten manifest atomically
	tmpPath := manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, rewritten, 0644); err != nil {
		result.Skipped = true
		result.SkipReason = "failed to write temp file: " + err.Error()
		return result
	}

	if err := os.Rename(tmpPath, manifestPath); err != nil {
		os.Remove(tmpPath)
		result.Skipped = true
		result.SkipReason = "failed to rename: " + err.Error()
		return result
	}

	return result
}
