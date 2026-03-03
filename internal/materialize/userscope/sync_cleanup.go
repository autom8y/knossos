package userscope

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/provenance"
)

// copyUserFile copies a file preserving permissions and ensuring parent directories exist.
func copyUserFile(src, dst string) error {
	// Ensure parent directory exists
	if err := ensureDirForFile(dst); err != nil {
		return err
	}

	// Read source
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Get source permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	perm := info.Mode()

	// For hooks, ensure executable bit for scripts
	if isExecutableFile(src) && perm&0111 == 0 {
		perm |= 0755
	}

	// Write destination with appropriate permissions
	return os.WriteFile(dst, content, perm)
}

// ensureDirForFile ensures the parent directory of a file path exists.
func ensureDirForFile(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

// isExecutableFile checks if a file should be executable (hooks).
func isExecutableFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	executableExtensions := map[string]bool{
		".sh":   true,
		".bash": true,
		".zsh":  true,
		".py":   true,
		".rb":   true,
		".pl":   true,
	}
	if executableExtensions[ext] {
		return true
	}
	// Check if in lib/ directory or has hook- prefix
	if strings.Contains(path, string(filepath.Separator)+"lib"+string(filepath.Separator)) {
		return true
	}
	base := filepath.Base(path)
	return strings.HasPrefix(base, "hook-") || strings.HasPrefix(base, "pre-") || strings.HasPrefix(base, "post-")
}

// removeUserOrphan removes an orphaned knossos-owned entry.
func removeUserOrphan(key string, manifest *provenance.ProvenanceManifest, userClaudeDir string) {
	entry := manifest.Entries[key]
	if entry == nil || entry.Owner != provenance.OwnerKnossos {
		return // Safety: only remove knossos-owned orphans
	}

	// Determine target path from key
	var targetPath string
	if strings.HasPrefix(key, "commands/") {
		targetPath = filepath.Join(userClaudeDir, key)
	} else if strings.HasPrefix(key, "skills/") {
		targetPath = filepath.Join(userClaudeDir, key)
	} else if strings.HasPrefix(key, "agents/") {
		targetPath = filepath.Join(userClaudeDir, key)
	} else if strings.HasPrefix(key, "hooks/") {
		targetPath = filepath.Join(userClaudeDir, key)
	}

	if targetPath == "" {
		return
	}

	// Remove file or directory
	if strings.HasSuffix(key, "/") {
		os.RemoveAll(targetPath)
	} else {
		os.Remove(targetPath)
	}

	// Remove from manifest
	delete(manifest.Entries, key)
}

// cleanupOldManifests removes legacy JSON manifest files and their v2-backup remnants.
// The v1 JSON manifests were superseded by USER_PROVENANCE_MANIFEST.yaml.
// The .v2-backup files were created by this function during v1-to-v2 migration
// and serve no rollback purpose now that migration is complete.
func cleanupOldManifests(userClaudeDir string) {
	oldManifests := []string{
		filepath.Join(userClaudeDir, "USER_AGENT_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_MENA_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_HOOKS_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_COMMAND_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_SKILL_MANIFEST.json"),
	}
	for _, path := range oldManifests {
		// Remove the original JSON manifest if still present.
		// Skip backup creation -- migration is complete, backups serve no purpose.
		os.Remove(path)

		// Remove .v2-backup remnants from previous migration runs.
		os.Remove(path + ".v2-backup")
	}
}

// recoverUserResource adopts existing files matching knossos sources.
func recoverUserResource(
	resourceType SyncResource,
	sourceDir string,
	targetDir string,
	manifest *provenance.ProvenanceManifest,
	nested bool,
	opts SyncOptions,
) error {
	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil // Nothing to recover
	}

	return filepath.WalkDir(targetDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(targetDir, path)

		// Build manifest key
		var manifestKey string
		if resourceType == ResourceMena {
			// Determine which target directory we're in
			if strings.Contains(targetDir, "commands") {
				manifestKey = "commands/" + relPath
			} else {
				manifestKey = "skills/" + relPath
			}
		} else {
			prefix := resourcePrefixForType(resourceType)
			if nested {
				manifestKey = prefix + relPath
			} else {
				manifestKey = prefix + filepath.Base(relPath)
			}
		}

		// Skip if already in manifest
		if _, exists := manifest.Entries[manifestKey]; exists {
			return nil
		}

		// Find source file
		var sourcePath string
		if resourceType == ResourceMena {
			// For mena, find source file with .dro/.lego variants
			sourcePath, _ = findMenaSource(sourceDir, relPath)
			if sourcePath == "" {
				// Not in knossos source - mark as user
				if !opts.DryRun {
					targetChecksum, checksumErr := checksum.File(path)
					if checksumErr != nil {
						slog.Warn("checksum failed, treating as changed", "path", path, "error", checksumErr)
					}
					manifest.Entries[manifestKey] = provenance.NewUserEntry(
						provenance.ScopeUser, targetChecksum,
					)
				}
				return nil
			}
		} else {
			sourcePath = filepath.Join(sourceDir, relPath)
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				// Not in knossos - mark as user
				if !opts.DryRun {
					targetChecksum, checksumErr := checksum.File(path)
					if checksumErr != nil {
						slog.Warn("checksum failed, treating as changed", "path", path, "error", checksumErr)
					}
					manifest.Entries[manifestKey] = provenance.NewUserEntry(
						provenance.ScopeUser, targetChecksum,
					)
				}
				return nil
			}
		}

		// Compare checksums
		sourceChecksum, srcErr := checksum.File(sourcePath)
		if srcErr != nil {
			slog.Warn("checksum failed, treating as changed", "path", sourcePath, "error", srcErr)
		}
		targetChecksum, tgtErr := checksum.File(path)
		if tgtErr != nil {
			slog.Warn("checksum failed, treating as changed", "path", path, "error", tgtErr)
		}
		sourceRelPath, _ := filepath.Rel(config.KnossosHome(), sourcePath)

		if !opts.DryRun {
			if srcErr == nil && tgtErr == nil && sourceChecksum == targetChecksum {
				manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
					provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
				)
			} else {
				manifest.Entries[manifestKey] = provenance.NewUserEntry(
					provenance.ScopeUser, targetChecksum,
				)
			}
		}

		return nil
	})
}

// resourcePrefixForType returns the manifest key prefix for a resource type.
func resourcePrefixForType(resourceType SyncResource) string {
	switch resourceType {
	case ResourceAgents:
		return "agents/"
	case ResourceMena:
		return "" // Mena already includes commands/ or skills/ prefix
	case ResourceHooks:
		return "hooks/"
	default:
		return ""
	}
}

// countUserCollisions counts collision entries in the skipped list.
func countUserCollisions(skipped []UserSkippedEntry) int {
	count := 0
	for _, entry := range skipped {
		if strings.Contains(entry.Reason, "collision") {
			count++
		}
	}
	return count
}

