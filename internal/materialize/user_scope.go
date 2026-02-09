package materialize

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// syncUserScope implements the full user-scope sync logic.
// It syncs resources from KNOSSOS_HOME/{agents,mena,hooks} to ~/.claude/ directories.
func (m *Materializer) syncUserScope(opts SyncOptions) (*UserScopeResult, error) {
	result := &UserScopeResult{
		Status:    "success",
		Resources: make(map[SyncResource]*UserResourceResult),
		Errors:    []UserResourceError{},
	}

	// Resolve KNOSSOS_HOME
	knossosHome := config.KnossosHome()
	if knossosHome == "" {
		return nil, ErrKnossosHomeNotSet()
	}

	// Resolve user ~/.claude/ directory
	userClaudeDir := paths.UserClaudeDir()

	// Load or bootstrap USER_PROVENANCE_MANIFEST.yaml
	manifestPath := provenance.UserManifestPath(userClaudeDir)
	manifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		return nil, err
	}

	// Initialize collision checker with project .claude/ directory
	projectClaudeDir := m.resolver.ClaudeDir()
	collisionChecker := NewCollisionChecker(projectClaudeDir)

	// Determine which resource types to sync
	resourcesToSync := []SyncResource{}
	if opts.Resource == ResourceAll || opts.Resource == "" {
		resourcesToSync = []SyncResource{ResourceAgents, ResourceMena, ResourceHooks}
	} else {
		resourcesToSync = []SyncResource{opts.Resource}
	}

	// Sync each resource type
	for _, resourceType := range resourcesToSync {
		resourceResult, err := syncUserResource(
			resourceType,
			knossosHome,
			userClaudeDir,
			manifest,
			collisionChecker,
			opts,
		)
		if err != nil {
			result.Errors = append(result.Errors, UserResourceError{
				Resource: resourceType,
				Err:      err.Error(),
			})
			continue
		}
		result.Resources[resourceType] = resourceResult

		// Aggregate totals
		result.Totals.Added += resourceResult.Summary.Added
		result.Totals.Updated += resourceResult.Summary.Updated
		result.Totals.Skipped += resourceResult.Summary.Skipped
		result.Totals.Unchanged += resourceResult.Summary.Unchanged
		result.Totals.Collisions += resourceResult.Summary.Collisions
	}

	// Save manifest if not dry-run
	if !opts.DryRun {
		manifest.LastSync = time.Now().UTC()
		if err := provenance.Save(manifestPath, manifest); err != nil {
			return nil, err
		}

		// Clean up old JSON manifests
		cleanupOldManifests(userClaudeDir)
	}

	return result, nil
}

// syncUserResource syncs a single resource type from KNOSSOS_HOME to ~/.claude/.
func syncUserResource(
	resourceType SyncResource,
	knossosHome string,
	userClaudeDir string,
	manifest *provenance.ProvenanceManifest,
	collisionChecker *CollisionChecker,
	opts SyncOptions,
) (*UserResourceResult, error) {
	// Resolve paths based on resource type
	var sourceDir string
	var targetDirs []string
	var nested bool

	switch resourceType {
	case ResourceAgents:
		sourceDir = filepath.Join(knossosHome, "agents")
		targetDirs = []string{filepath.Join(userClaudeDir, "agents")}
		nested = false
	case ResourceMena:
		sourceDir = filepath.Join(knossosHome, "mena")
		targetDirs = []string{
			filepath.Join(userClaudeDir, "commands"),
			filepath.Join(userClaudeDir, "skills"),
		}
		nested = true
	case ResourceHooks:
		sourceDir = filepath.Join(knossosHome, "hooks")
		targetDirs = []string{filepath.Join(userClaudeDir, "hooks")}
		nested = true
	default:
		return nil, ErrInvalidResourceType()
	}

	result := &UserResourceResult{
		Source: sourceDir,
		Target: strings.Join(targetDirs, " + "),
		Changes: UserSyncChanges{
			Added:     []string{},
			Updated:   []string{},
			Skipped:   []UserSkippedEntry{},
			Unchanged: []string{},
		},
	}

	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return result, nil // No source = no-op
	}

	// Ensure target directories exist (unless dry-run)
	if !opts.DryRun {
		for _, targetDir := range targetDirs {
			if err := paths.EnsureDir(targetDir); err != nil {
				return nil, err
			}
		}
	}

	// Recovery mode: adopt existing untracked files
	if opts.Recover {
		for _, targetDir := range targetDirs {
			if err := recoverUserResource(
				resourceType,
				sourceDir,
				targetDir,
				manifest,
				nested,
				opts,
			); err != nil {
				return nil, err
			}
		}
	}

	// Phase 1: Snapshot existing knossos-owned entries for this resource
	snapshot := make(map[string]bool)
	prefix := resourcePrefixForType(resourceType)

	for key, entry := range manifest.Entries {
		if entry.Owner == provenance.OwnerKnossos {
			// For mena, match both "commands/" and "skills/" prefixes
			if resourceType == ResourceMena {
				if strings.HasPrefix(key, "commands/") || strings.HasPrefix(key, "skills/") {
					snapshot[key] = false // not yet seen
				}
			} else if prefix != "" && strings.HasPrefix(key, prefix) {
				snapshot[key] = false
			}
		}
	}

	// Phase 2: Walk source directory and sync files
	now := time.Now().UTC()
	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Compute relative path from source
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Build manifest key
		var manifestKey string
		var targetDir string
		var targetPath string

		if resourceType == ResourceMena {
			// Mena: detect type, route to target, strip extension
			menaType := DetectMenaType(filepath.Base(relPath))
			targetSubdir := RouteMenaFile(filepath.Base(relPath))
			strippedName := StripMenaExtension(filepath.Base(relPath))
			dir := filepath.Dir(relPath)

			if targetSubdir == "commands" {
				targetDir = targetDirs[0]
				manifestKey = "commands/" + filepath.Join(dir, strippedName)
			} else {
				targetDir = targetDirs[1]
				manifestKey = "skills/" + filepath.Join(dir, strippedName)
			}
			targetPath = filepath.Join(targetDir, dir, strippedName)

			// Skip if mena type is unknown
			if menaType == "" {
				return nil
			}
		} else {
			// Agents/Hooks: use basename or relPath
			if nested {
				manifestKey = prefix + relPath
				targetPath = filepath.Join(targetDirs[0], relPath)
			} else {
				manifestKey = prefix + filepath.Base(relPath)
				targetPath = filepath.Join(targetDirs[0], filepath.Base(relPath))
			}
		}

		// Mark as seen in snapshot
		if _, tracked := snapshot[manifestKey]; tracked {
			snapshot[manifestKey] = true
		}

		// Check for collision with rite
		if collision, _ := collisionChecker.CheckCollision(manifestKey); collision {
			result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
				Name:   manifestKey,
				Reason: "collision with rite resource",
			})
			return nil
		}

		// Compute source checksum
		sourceChecksum, err := checksum.File(path)
		if err != nil {
			return err
		}

		// Check existing manifest entry
		entry, exists := manifest.Entries[manifestKey]

		// Compute source path relative to knossos home
		sourceRelPath, _ := filepath.Rel(knossosHome, path)

		// Handle new file (not in manifest)
		if !exists {
			// Check if target exists (untracked)
			if _, statErr := os.Stat(targetPath); statErr == nil {
				if opts.Recover {
					targetChecksum, _ := checksum.File(targetPath)
					if targetChecksum == sourceChecksum {
						// Adopt as knossos-owned
						if !opts.DryRun {
							manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerKnossos,
								Scope:      provenance.ScopeUser,
								SourcePath: sourceRelPath,
								SourceType: "user-sync",
								Checksum:   sourceChecksum,
								LastSynced: now,
							}
						}
						result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
					} else {
						// Adopt as user-owned (modified)
						if !opts.DryRun {
							manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerUser,
								Scope:      provenance.ScopeUser,
								Checksum:   targetChecksum,
								LastSynced: now,
							}
						}
						result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
							Name:   manifestKey,
							Reason: "adopted as user (local modifications)",
						})
					}
					return nil
				}
				// Not recovering - mark as user-created
				if !opts.DryRun {
					targetChecksum, _ := checksum.File(targetPath)
					manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerUser,
						Scope:      provenance.ScopeUser,
						Checksum:   targetChecksum,
						LastSynced: now,
					}
				}
				result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
					Name:   manifestKey,
					Reason: "user-created",
				})
				return nil
			}

			// New file, target doesn't exist - copy it
			if !opts.DryRun {
				if err := copyUserFile(path, targetPath); err != nil {
					return err
				}
				manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeUser,
					SourcePath: sourceRelPath,
					SourceType: "user-sync",
					Checksum:   sourceChecksum,
					LastSynced: now,
				}
			}
			result.Changes.Added = append(result.Changes.Added, manifestKey)
			return nil
		}

		// Existing entry
		switch entry.Owner {
		case provenance.OwnerUser, provenance.OwnerUntracked:
			// Never touch user-created files
			result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
				Name:   manifestKey,
				Reason: "user-created",
			})

		case provenance.OwnerKnossos:
			// Check if source changed
			if entry.Checksum == sourceChecksum {
				// No change in source
				result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
			} else {
				// Source changed - check if target diverged
				targetChecksum, _ := checksum.File(targetPath)
				if targetChecksum == entry.Checksum {
					// Target unchanged, update from source
					if !opts.DryRun {
						if err := copyUserFile(path, targetPath); err != nil {
							return err
						}
						manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
							Owner:      provenance.OwnerKnossos,
							Scope:      provenance.ScopeUser,
							SourcePath: sourceRelPath,
							SourceType: "user-sync",
							Checksum:   sourceChecksum,
							LastSynced: now,
						}
					}
					result.Changes.Updated = append(result.Changes.Updated, manifestKey)
				} else {
					// Target diverged
					if opts.OverwriteDiverged {
						// Force overwrite
						if !opts.DryRun {
							if err := copyUserFile(path, targetPath); err != nil {
								return err
							}
							manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerKnossos,
								Scope:      provenance.ScopeUser,
								SourcePath: sourceRelPath,
								SourceType: "user-sync",
								Checksum:   sourceChecksum,
								LastSynced: now,
							}
						}
						result.Changes.Updated = append(result.Changes.Updated, manifestKey)
					} else {
						// Skip diverged without --overwrite-diverged
						result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
							Name:   manifestKey,
							Reason: "diverged (use --overwrite-diverged to force)",
						})
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Phase 3: Orphan removal
	if !opts.DryRun && !opts.KeepOrphans {
		for key, seen := range snapshot {
			if !seen {
				removeUserOrphan(key, manifest, userClaudeDir)
			}
		}
	}

	// Calculate summary
	result.Summary = UserSyncSummary{
		Added:      len(result.Changes.Added),
		Updated:    len(result.Changes.Updated),
		Skipped:    len(result.Changes.Skipped),
		Unchanged:  len(result.Changes.Unchanged),
		Collisions: countUserCollisions(result.Changes.Skipped),
	}

	return result, nil
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

	now := time.Now().UTC()

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
					targetChecksum, _ := checksum.File(path)
					manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerUser,
						Scope:      provenance.ScopeUser,
						Checksum:   targetChecksum,
						LastSynced: now,
					}
				}
				return nil
			}
		} else {
			sourcePath = filepath.Join(sourceDir, relPath)
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				// Not in knossos - mark as user
				if !opts.DryRun {
					targetChecksum, _ := checksum.File(path)
					manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerUser,
						Scope:      provenance.ScopeUser,
						Checksum:   targetChecksum,
						LastSynced: now,
					}
				}
				return nil
			}
		}

		// Compare checksums
		sourceChecksum, _ := checksum.File(sourcePath)
		targetChecksum, _ := checksum.File(path)
		sourceRelPath, _ := filepath.Rel(config.KnossosHome(), sourcePath)

		if !opts.DryRun {
			if sourceChecksum == targetChecksum {
				manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeUser,
					SourcePath: sourceRelPath,
					SourceType: "user-sync",
					Checksum:   sourceChecksum,
					LastSynced: now,
				}
			} else {
				manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerUser,
					Scope:      provenance.ScopeUser,
					Checksum:   targetChecksum,
					LastSynced: now,
				}
			}
		}

		return nil
	})
}

// findMenaSource locates the source file for a stripped manifest key.
// Returns the source path and mena type, or ("", "") if not found.
func findMenaSource(sourceDir, strippedRelPath string) (string, string) {
	dir := filepath.Dir(strippedRelPath)
	base := filepath.Base(strippedRelPath)

	// Try original path (no infix)
	candidate := filepath.Join(sourceDir, strippedRelPath)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, DetectMenaType(base)
	}

	// Try .dro variant: insert .dro before the last extension
	ext := filepath.Ext(base)
	nameNoExt := base[:len(base)-len(ext)]
	droName := nameNoExt + ".dro" + ext
	candidate = filepath.Join(sourceDir, dir, droName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, "dro"
	}

	// Try .lego variant
	legoName := nameNoExt + ".lego" + ext
	candidate = filepath.Join(sourceDir, dir, legoName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, "lego"
	}

	return "", ""
}

// copyUserFile copies a file preserving permissions and ensuring parent directories exist.
func copyUserFile(src, dst string) error {
	// Ensure parent directory exists
	if err := paths.EnsureDir(filepath.Dir(dst)); err != nil {
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

// cleanupOldManifests removes legacy JSON manifest files.
func cleanupOldManifests(userClaudeDir string) {
	oldManifests := []string{
		filepath.Join(userClaudeDir, "USER_AGENT_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_MENA_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_HOOKS_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_COMMAND_MANIFEST.json"),
		filepath.Join(userClaudeDir, "USER_SKILL_MANIFEST.json"),
	}
	for _, path := range oldManifests {
		// Backup before removal for safety
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Already gone or unreadable
		}
		backupPath := path + ".v2-backup"
		os.WriteFile(backupPath, data, 0644) // Best effort
		os.Remove(path)
	}
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

// ErrKnossosHomeNotSet returns an error when KNOSSOS_HOME is not set.
func ErrKnossosHomeNotSet() error {
	return &UserSyncError{Message: "KNOSSOS_HOME not set. Set KNOSSOS_HOME environment variable."}
}

// ErrInvalidResourceType returns an error for invalid resource types.
func ErrInvalidResourceType() error {
	return &UserSyncError{Message: "invalid resource type"}
}

// UserSyncError represents a user sync error.
type UserSyncError struct {
	Message string
}

func (e *UserSyncError) Error() string {
	return e.Message
}
