package userscope

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// syncer holds dependencies for user-scope sync operations.
// Created internally by SyncUserScope from the params struct.
type syncer struct {
	resolver       *paths.Resolver
	embeddedAgents fs.FS
	embeddedMena   fs.FS
	embeddedRites  fs.FS
}

// SyncUserScope is the primary entry point for user-scope sync.
// It syncs resources from KNOSSOS_HOME/{agents,mena,hooks} to ~/.claude/ directories.
func SyncUserScope(params SyncUserScopeParams) (*UserScopeResult, error) {
	s := &syncer{
		resolver:       params.Resolver,
		embeddedAgents: params.EmbeddedAgents,
		embeddedMena:   params.EmbeddedMena,
		embeddedRites:  params.EmbeddedRites,
	}
	return s.syncUserScope(params.Opts)
}

// syncUserScope implements the full user-scope sync logic.
// It syncs resources from KNOSSOS_HOME/{agents,mena,hooks} to ~/.claude/ directories.
//
// Note: This file uses os.WriteFile rather than writeIfChanged because user-scope
// targets (~/.claude/) are outside CC's project-level file watcher scope. The
// writeIfChanged optimization prevents unnecessary file watcher triggers in the
// project .claude/ directory, but that concern does not apply here.
func (s *syncer) syncUserScope(opts SyncOptions) (*UserScopeResult, error) {
	result := &UserScopeResult{
		Status:    "success",
		Resources: make(map[SyncResource]*UserResourceResult),
		Errors:    []UserResourceError{},
	}

	// Resolve KNOSSOS_HOME
	knossosHome := config.KnossosHome()
	if knossosHome == "" && s.embeddedAgents == nil && s.embeddedMena == nil {
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

	// Wipe stale knossos-produced mena entries before sync.
	// One-time migration from old non-flattening pipeline; subsequent runs are no-ops.
	if knossosHome != "" {
		wipeKnossosOwnedMenaEntries(knossosHome, userClaudeDir, manifest, opts.DryRun)
	}

	// Initialize collision checker with project .claude/ directory.
	// The checker requires a rite PROVENANCE_MANIFEST.yaml to detect collisions.
	// When the manifest is missing, try to fall back to the main worktree's
	// provenance (for linked git worktrees). If that also fails and an ACTIVE_RITE
	// is present, fail-closed: skip all user-scope writes to prevent cross-project
	// contamination via USER_PROVENANCE_MANIFEST.yaml.
	// When no ACTIVE_RITE exists, there is no rite to protect against; proceed.
	projectClaudeDir := s.resolver.ClaudeDir()
	collisionChecker := NewCollisionChecker(projectClaudeDir)
	if !collisionChecker.IsEffective() {
		// Attempt to use the main worktree's provenance manifest when we are
		// running inside a linked git worktree. This avoids the blanket "skip all"
		// behaviour when the main worktree has a valid rite provenance.
		if mainDir, err := worktreeMainDir(s.resolver.ProjectRoot()); err == nil {
			mainClaudeDir := filepath.Join(mainDir, ".claude")
			mainChecker := NewCollisionChecker(mainClaudeDir)
			if mainChecker.IsEffective() {
				slog.Info("userscope: collision checker fell back to main worktree provenance", "path", mainClaudeDir)
				collisionChecker = mainChecker
			}
		}

		// If still not effective, apply the fail-closed guard.
		if !collisionChecker.IsEffective() {
			activeRite := s.resolver.ReadActiveRite()
			if activeRite != "" {
				slog.Warn("userscope: collision checker not effective, skipping user-scope writes to prevent contamination", "path", projectClaudeDir, "active_rite", activeRite)
				return result, nil
			}
			// No ACTIVE_RITE: no rite to protect. Proceed without collision checking.
		}
	}

	// Determine which resource types to sync
	var resourcesToSync []SyncResource
	if opts.Resource == ResourceAll || opts.Resource == "" {
		resourcesToSync = []SyncResource{ResourceAgents, ResourceMena, ResourceHooks}
	} else {
		resourcesToSync = []SyncResource{opts.Resource}
	}

	// Sync each resource type
	for _, resourceType := range resourcesToSync {
		resourceResult, err := s.syncUserResource(
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
func (s *syncer) syncUserResource(
	resourceType SyncResource,
	knossosHome string,
	userClaudeDir string,
	manifest *provenance.ProvenanceManifest,
	collisionChecker *CollisionChecker,
	opts SyncOptions,
) (*UserResourceResult, error) {
	// Mena uses dedicated CollectMena-based pipeline for namespace flattening
	// and companion hiding parity with rite-scope.
	if resourceType == ResourceMena {
		return s.syncUserMena(knossosHome, userClaudeDir, manifest, collisionChecker, opts)
	}

	// Resolve paths based on resource type
	var sourceDir string
	var targetDirs []string
	var nested bool

	switch resourceType {
	case ResourceAgents:
		sourceDir = filepath.Join(knossosHome, "agents")
		targetDirs = []string{filepath.Join(userClaudeDir, "agents")}
		nested = false
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
	sourceExists := true
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		sourceExists = false
	}

	// If filesystem source doesn't exist, try embedded fallback
	if !sourceExists {
		if resourceType == ResourceAgents && s.embeddedAgents != nil {
			return s.syncUserResourceFromEmbedded(
				resourceType, s.embeddedAgents, "agents",
				userClaudeDir, manifest, collisionChecker, opts,
			)
		}
		// Hooks: no embedded fallback (KNOSSOS_HOME-only)
		return result, nil
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
		if entry.Owner == provenance.OwnerKnossos && prefix != "" && strings.HasPrefix(key, prefix) {
			snapshot[key] = false
		}
	}

	// Phase 2: Walk source directory and sync files
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

		// Build manifest key (agents/hooks only; mena early-returns above)
		var manifestKey string
		var targetPath string

		if nested {
			manifestKey = prefix + relPath
			targetPath = filepath.Join(targetDirs[0], relPath)
		} else {
			manifestKey = prefix + filepath.Base(relPath)
			targetPath = filepath.Join(targetDirs[0], filepath.Base(relPath))
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
					targetChecksum, checksumErr := checksum.File(targetPath)
					if checksumErr != nil {
						slog.Warn("checksum failed, treating as changed", "path", targetPath, "error", checksumErr)
					}
					if checksumErr == nil && targetChecksum == sourceChecksum {
						// Adopt as knossos-owned
						if !opts.DryRun {
							manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
								provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
							)
						}
						result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
					} else {
						// Adopt as user-owned (modified)
						if !opts.DryRun {
							manifest.Entries[manifestKey] = provenance.NewUserEntry(
								provenance.ScopeUser, targetChecksum,
							)
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
					targetChecksum, checksumErr := checksum.File(targetPath)
					if checksumErr != nil {
						slog.Warn("checksum failed, treating as changed", "path", targetPath, "error", checksumErr)
					}
					manifest.Entries[manifestKey] = provenance.NewUserEntry(
						provenance.ScopeUser, targetChecksum,
					)
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
				manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
					provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
				)
			}
			result.Changes.Added = append(result.Changes.Added, manifestKey)
			return nil
		}

		// Existing entry
		switch entry.Owner {
		case provenance.OwnerUser, provenance.OwnerUntracked:
			// If file was deleted from disk, clear stale manifest entry so it can be re-created
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				delete(manifest.Entries, manifestKey)
				// Re-create from source
				if !opts.DryRun {
					if err := copyUserFile(path, targetPath); err != nil {
						return err
					}
					manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
						provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
					)
				}
				result.Changes.Added = append(result.Changes.Added, manifestKey)
			} else {
				// File exists on disk — never touch user-created files
				result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
					Name:   manifestKey,
					Reason: "user-created",
				})
			}

		case provenance.OwnerKnossos:
			// If target was deleted from disk, re-create unconditionally
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				if !opts.DryRun {
					if err := copyUserFile(path, targetPath); err != nil {
						return err
					}
					manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
						provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
					)
				}
				result.Changes.Added = append(result.Changes.Added, manifestKey)
			} else if entry.Checksum == sourceChecksum {
				// No change in source
				result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
			} else {
				// Source changed - check if target diverged
				targetChecksum, checksumErr := checksum.File(targetPath)
				if checksumErr != nil {
					slog.Warn("checksum failed, treating as changed", "path", targetPath, "error", checksumErr)
				}
				if targetChecksum == entry.Checksum {
					// Target unchanged, update from source
					if !opts.DryRun {
						if err := copyUserFile(path, targetPath); err != nil {
							return err
						}
						manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
							provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
						)
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
							manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
								provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
							)
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
