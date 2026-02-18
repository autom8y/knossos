package materialize

import (
	"io/fs"
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
//
// Note: This file uses os.WriteFile rather than writeIfChanged because user-scope
// targets (~/.claude/) are outside CC's project-level file watcher scope. The
// writeIfChanged optimization prevents unnecessary file watcher triggers in the
// project .claude/ directory, but that concern does not apply here.
func (m *Materializer) syncUserScope(opts SyncOptions) (*UserScopeResult, error) {
	result := &UserScopeResult{
		Status:    "success",
		Resources: make(map[SyncResource]*UserResourceResult),
		Errors:    []UserResourceError{},
	}

	// Resolve KNOSSOS_HOME
	knossosHome := config.KnossosHome()
	if knossosHome == "" && m.embeddedAgents == nil && m.embeddedMena == nil {
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
		resourceResult, err := m.syncUserResource(
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
func (m *Materializer) syncUserResource(
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
		return m.syncUserMena(knossosHome, userClaudeDir, manifest, collisionChecker, opts)
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
		if resourceType == ResourceAgents && m.embeddedAgents != nil {
			return m.syncUserResourceFromEmbedded(
				resourceType, m.embeddedAgents, "agents",
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
			// If file was deleted from disk, clear stale manifest entry so it can be re-created
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				delete(manifest.Entries, manifestKey)
				// Re-create from source
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
			} else if entry.Checksum == sourceChecksum {
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

// syncUserResourceFromEmbedded syncs agents from an embedded filesystem
// when KNOSSOS_HOME is unavailable. Simplified version: no recovery mode,
// no orphan removal (embedded source is authoritative).
func (m *Materializer) syncUserResourceFromEmbedded(
	resourceType SyncResource,
	embeddedFS fs.FS,
	embeddedRoot string,
	userClaudeDir string,
	manifest *provenance.ProvenanceManifest,
	collisionChecker *CollisionChecker,
	opts SyncOptions,
) (*UserResourceResult, error) {
	prefix := resourcePrefixForType(resourceType)
	targetDir := filepath.Join(userClaudeDir, string(resourceType))

	result := &UserResourceResult{
		Source: "embedded:" + embeddedRoot,
		Target: targetDir,
		Changes: UserSyncChanges{
			Added:     []string{},
			Updated:   []string{},
			Skipped:   []UserSkippedEntry{},
			Unchanged: []string{},
		},
	}

	// Walk embedded filesystem
	err := fs.WalkDir(embeddedFS, embeddedRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Compute relative path from embedded root
		relPath, err := filepath.Rel(embeddedRoot, path)
		if err != nil {
			return err
		}

		manifestKey := prefix + filepath.Base(relPath)
		targetPath := filepath.Join(targetDir, filepath.Base(relPath))

		// Check for collision with rite
		if collision, _ := collisionChecker.CheckCollision(manifestKey); collision {
			result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
				Name:   manifestKey,
				Reason: "collision with rite resource",
			})
			return nil
		}

		// Read embedded content
		content, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			return err
		}

		sourceChecksum := checksum.Bytes(content)
		now := time.Now().UTC()

		// Check existing manifest entry
		entry, exists := manifest.Entries[manifestKey]

		if !exists {
			// Check if target exists (untracked)
			if _, statErr := os.Stat(targetPath); statErr == nil {
				// Exists untracked — mark as user-created, don't overwrite
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

			// New file — write it
			if !opts.DryRun {
				if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, content, 0644); err != nil {
					return err
				}
				manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeUser,
					SourcePath: "embedded:" + path,
					SourceType: "embedded",
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
			// If file was deleted from disk, clear stale manifest entry so it can be re-created
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				delete(manifest.Entries, manifestKey)
				if !opts.DryRun {
					if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
						return err
					}
					if err := os.WriteFile(targetPath, content, 0644); err != nil {
						return err
					}
					manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerKnossos,
						Scope:      provenance.ScopeUser,
						SourcePath: "embedded:" + path,
						SourceType: "embedded",
						Checksum:   sourceChecksum,
						LastSynced: now,
					}
				}
				result.Changes.Added = append(result.Changes.Added, manifestKey)
			} else {
				result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
					Name:   manifestKey,
					Reason: "user-created",
				})
			}

		case provenance.OwnerKnossos:
			// If target was deleted from disk, re-create unconditionally
			if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
				if !opts.DryRun {
					if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
						return err
					}
					if err := os.WriteFile(targetPath, content, 0644); err != nil {
						return err
					}
					manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerKnossos,
						Scope:      provenance.ScopeUser,
						SourcePath: "embedded:" + path,
						SourceType: "embedded",
						Checksum:   sourceChecksum,
						LastSynced: now,
					}
				}
				result.Changes.Added = append(result.Changes.Added, manifestKey)
			} else if entry.Checksum == sourceChecksum {
				result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
			} else {
				targetChecksum, _ := checksum.File(targetPath)
				if targetChecksum == entry.Checksum {
					// Target unchanged, update from embedded source
					if !opts.DryRun {
						if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
							return err
						}
						if err := os.WriteFile(targetPath, content, 0644); err != nil {
							return err
						}
						manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
							Owner:      provenance.OwnerKnossos,
							Scope:      provenance.ScopeUser,
							SourcePath: "embedded:" + path,
							SourceType: "embedded",
							Checksum:   sourceChecksum,
							LastSynced: now,
						}
					}
					result.Changes.Updated = append(result.Changes.Updated, manifestKey)
				} else {
					if opts.OverwriteDiverged {
						if !opts.DryRun {
							if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
								return err
							}
							if err := os.WriteFile(targetPath, content, 0644); err != nil {
								return err
							}
							manifest.Entries[manifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerKnossos,
								Scope:      provenance.ScopeUser,
								SourcePath: "embedded:" + path,
								SourceType: "embedded",
								Checksum:   sourceChecksum,
								LastSynced: now,
							}
						}
						result.Changes.Updated = append(result.Changes.Updated, manifestKey)
					} else {
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

	result.Summary = UserSyncSummary{
		Added:      len(result.Changes.Added),
		Updated:    len(result.Changes.Updated),
		Skipped:    len(result.Changes.Skipped),
		Unchanged:  len(result.Changes.Unchanged),
		Collisions: countUserCollisions(result.Changes.Skipped),
	}

	return result, nil
}

// syncUserMena syncs mena files from KNOSSOS_HOME/mena to ~/.claude/{commands,skills}
// using the CollectMena pipeline for namespace flattening and companion hiding parity
// with the rite-scope pipeline (SyncMena).
func (m *Materializer) syncUserMena(
	knossosHome string,
	userClaudeDir string,
	manifest *provenance.ProvenanceManifest,
	collisionChecker *CollisionChecker,
	opts SyncOptions,
) (*UserResourceResult, error) {
	sourceDir := filepath.Join(knossosHome, "mena")
	commandsDir := filepath.Join(userClaudeDir, "commands")
	skillsDir := filepath.Join(userClaudeDir, "skills")

	result := &UserResourceResult{
		Source: sourceDir,
		Target: commandsDir + " + " + skillsDir,
		Changes: UserSyncChanges{
			Added:     []string{},
			Updated:   []string{},
			Skipped:   []UserSkippedEntry{},
			Unchanged: []string{},
		},
	}

	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		// Try embedded mena fallback
		if m.embeddedMena != nil {
			return m.syncUserMenaFromEmbedded(userClaudeDir, manifest, collisionChecker, opts)
		}
		return result, nil // No source = no-op
	}

	// Ensure target directories exist (unless dry-run)
	if !opts.DryRun {
		if err := paths.EnsureDir(commandsDir); err != nil {
			return nil, err
		}
		if err := paths.EnsureDir(skillsDir); err != nil {
			return nil, err
		}
	}

	// Call CollectMena to get flattened, type-resolved entries.
	// Empty target dirs: user-scope provenance handles user-collision protection.
	sources := []MenaSource{{Path: sourceDir}}
	collectOpts := MenaProjectionOptions{
		Filter: ProjectAll,
	}
	resolution, err := CollectMena(sources, collectOpts)
	if err != nil {
		return nil, err
	}

	// Phase 1: Snapshot existing knossos-owned mena entries for orphan detection
	snapshot := make(map[string]bool)
	for key, entry := range manifest.Entries {
		if entry.Owner == provenance.OwnerKnossos {
			if strings.HasPrefix(key, "commands/") || strings.HasPrefix(key, "skills/") {
				snapshot[key] = false // not yet seen
			}
		}
	}

	now := time.Now().UTC()

	// Phase 2a: Sync resolved directory entries
	for _, entry := range resolution.Entries {
		var targetBaseDir string
		var manifestPrefix string
		if entry.MenaType == "dro" {
			targetBaseDir = commandsDir
			manifestPrefix = "commands/"
		} else {
			targetBaseDir = skillsDir
			manifestPrefix = "skills/"
		}

		// Only filesystem sources in user-scope (no embedded)
		if entry.Source.Path == "" {
			continue
		}

		// Walk source directory files
		walkErr := filepath.WalkDir(entry.Source.Path, func(path string, d os.DirEntry, wErr error) error {
			if wErr != nil {
				return wErr
			}
			if d.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(entry.Source.Path, path)
			if err != nil {
				return err
			}

			// Strip mena extension
			strippedName := StripMenaExtension(filepath.Base(relPath))
			fileDir := filepath.Dir(relPath)
			strippedRel := filepath.Join(fileDir, strippedName)

			// Manifest key uses flat name
			manifestKey := manifestPrefix + filepath.Join(entry.FlatName, strippedRel)
			targetPath := filepath.Join(targetBaseDir, entry.FlatName, strippedRel)

			// Read source content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Apply companion hiding for dro non-INDEX markdown files
			if entry.MenaType == "dro" && strippedName != "INDEX.md" && strings.HasSuffix(strippedName, ".md") {
				content = injectCompanionHideFrontmatter(content)
			}

			// Checksum of transformed content (what we'll write)
			sourceChecksum := checksum.Bytes(content)

			return syncUserMenaFile(
				manifestKey, targetPath, content, sourceChecksum,
				filepath.Join("mena", entry.FlatName, strippedRel),
				manifest, collisionChecker, snapshot, result, opts, now,
			)
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}

	// Phase 2b: Sync resolved standalone files
	for _, sf := range resolution.Standalones {
		var targetBaseDir string
		var manifestPrefix string
		if sf.MenaType == "dro" {
			targetBaseDir = commandsDir
			manifestPrefix = "commands/"
		} else {
			targetBaseDir = skillsDir
			manifestPrefix = "skills/"
		}

		manifestKey := manifestPrefix + sf.FlatName
		targetPath := filepath.Join(targetBaseDir, sf.FlatName)

		content, err := os.ReadFile(sf.SrcPath)
		if err != nil {
			return nil, err
		}

		sourceChecksum := checksum.Bytes(content)
		sourceRelPath := filepath.Join("mena", sf.RelPath)

		if err := syncUserMenaFile(
			manifestKey, targetPath, content, sourceChecksum,
			sourceRelPath,
			manifest, collisionChecker, snapshot, result, opts, now,
		); err != nil {
			return nil, err
		}
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

// syncUserMenaFromEmbedded syncs mena from the embedded filesystem
// when KNOSSOS_HOME is unavailable.
func (m *Materializer) syncUserMenaFromEmbedded(
	userClaudeDir string,
	manifest *provenance.ProvenanceManifest,
	collisionChecker *CollisionChecker,
	opts SyncOptions,
) (*UserResourceResult, error) {
	commandsDir := filepath.Join(userClaudeDir, "commands")
	skillsDir := filepath.Join(userClaudeDir, "skills")

	result := &UserResourceResult{
		Source: "embedded:mena",
		Target: commandsDir + " + " + skillsDir,
		Changes: UserSyncChanges{
			Added:     []string{},
			Updated:   []string{},
			Skipped:   []UserSkippedEntry{},
			Unchanged: []string{},
		},
	}

	// Ensure target directories exist (unless dry-run)
	if !opts.DryRun {
		if err := paths.EnsureDir(commandsDir); err != nil {
			return nil, err
		}
		if err := paths.EnsureDir(skillsDir); err != nil {
			return nil, err
		}
	}

	// Use embedded FS as a MenaSource via the CollectMena pipeline.
	// CollectMena supports fs.FS sources via MenaSource.Fsys field.
	sources := []MenaSource{{Fsys: m.embeddedMena, FsysPath: "mena", IsEmbedded: true}}
	collectOpts := MenaProjectionOptions{
		Filter: ProjectAll,
	}
	resolution, err := CollectMena(sources, collectOpts)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	// Process resolved entries — read content from embedded FS
	for _, entry := range resolution.Entries {
		var targetBaseDir string
		var manifestPrefix string
		if entry.MenaType == "dro" {
			targetBaseDir = commandsDir
			manifestPrefix = "commands/"
		} else {
			targetBaseDir = skillsDir
			manifestPrefix = "skills/"
		}

		// Walk embedded source directory
		if entry.Source.Fsys == nil {
			continue
		}
		walkErr := fs.WalkDir(entry.Source.Fsys, entry.Source.FsysPath, func(path string, d fs.DirEntry, wErr error) error {
			if wErr != nil {
				return wErr
			}
			if d.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(entry.Source.FsysPath, path)
			if err != nil {
				return err
			}

			// Strip mena extension
			strippedName := StripMenaExtension(filepath.Base(relPath))
			fileDir := filepath.Dir(relPath)
			strippedRel := filepath.Join(fileDir, strippedName)

			// Manifest key uses flat name
			manifestKey := manifestPrefix + filepath.Join(entry.FlatName, strippedRel)
			targetPath := filepath.Join(targetBaseDir, entry.FlatName, strippedRel)

			// Read embedded content
			content, err := fs.ReadFile(entry.Source.Fsys, path)
			if err != nil {
				return err
			}

			// Apply companion hiding for dro non-INDEX markdown files
			if entry.MenaType == "dro" && strippedName != "INDEX.md" && strings.HasSuffix(strippedName, ".md") {
				content = injectCompanionHideFrontmatter(content)
			}

			sourceChecksum := checksum.Bytes(content)

			return syncUserMenaFile(
				manifestKey, targetPath, content, sourceChecksum,
				"embedded:mena/"+filepath.Join(entry.FlatName, strippedRel),
				manifest, collisionChecker, nil, result, opts, now,
			)
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}

	// Process standalone files
	for _, sf := range resolution.Standalones {
		var targetBaseDir string
		var manifestPrefix string
		if sf.MenaType == "dro" {
			targetBaseDir = commandsDir
			manifestPrefix = "commands/"
		} else {
			targetBaseDir = skillsDir
			manifestPrefix = "skills/"
		}

		manifestKey := manifestPrefix + sf.FlatName
		targetPath := filepath.Join(targetBaseDir, sf.FlatName)

		// Read from embedded FS
		content, err := fs.ReadFile(m.embeddedMena, sf.SrcPath)
		if err != nil {
			return nil, err
		}

		sourceChecksum := checksum.Bytes(content)

		if err := syncUserMenaFile(
			manifestKey, targetPath, content, sourceChecksum,
			"embedded:mena/"+sf.RelPath,
			manifest, collisionChecker, nil, result, opts, now,
		); err != nil {
			return nil, err
		}
	}

	result.Summary = UserSyncSummary{
		Added:      len(result.Changes.Added),
		Updated:    len(result.Changes.Updated),
		Skipped:    len(result.Changes.Skipped),
		Unchanged:  len(result.Changes.Unchanged),
		Collisions: countUserCollisions(result.Changes.Skipped),
	}

	return result, nil
}

// syncUserMenaFile handles provenance-tracked sync of a single mena file.
// Shared by both directory entries and standalone files in syncUserMena.
func syncUserMenaFile(
	manifestKey string,
	targetPath string,
	content []byte,
	sourceChecksum string,
	sourceRelPath string,
	manifest *provenance.ProvenanceManifest,
	collisionChecker *CollisionChecker,
	snapshot map[string]bool,
	result *UserResourceResult,
	opts SyncOptions,
	now time.Time,
) error {
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

	// Check existing manifest entry
	entry, exists := manifest.Entries[manifestKey]

	// Handle new file (not in manifest)
	if !exists {
		// Check if target exists (untracked)
		if _, statErr := os.Stat(targetPath); statErr == nil {
			if opts.Recover {
				targetChecksum, _ := checksum.File(targetPath)
				if targetChecksum == sourceChecksum {
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

		// New file, target doesn't exist - write it
		if !opts.DryRun {
			if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
				return err
			}
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
		// If file was deleted from disk, clear stale manifest entry so it can be re-created
		if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
			delete(manifest.Entries, manifestKey)
			if !opts.DryRun {
				if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
		} else {
			result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
				Name:   manifestKey,
				Reason: "user-created",
			})
		}

	case provenance.OwnerKnossos:
		// If target was deleted from disk, re-create unconditionally
		if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
			if !opts.DryRun {
				if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
		} else if entry.Checksum == sourceChecksum {
			result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
		} else {
			targetChecksum, _ := checksum.File(targetPath)
			if targetChecksum == entry.Checksum {
				// Target unchanged, update from source
				if !opts.DryRun {
					if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
						return err
					}
					if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
					if !opts.DryRun {
						if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
							return err
						}
						if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
					result.Changes.Skipped = append(result.Changes.Skipped, UserSkippedEntry{
						Name:   manifestKey,
						Reason: "diverged (use --overwrite-diverged to force)",
					})
				}
			}
		}
	}

	return nil
}

// wipeKnossosOwnedMenaEntries removes ALL knossos-owned entries with commands/
// or skills/ prefix from the user manifest and deletes corresponding files.
// This is a one-time migration from the old non-flattening user-scope pipeline.
// On subsequent runs with clean entries, this is a no-op.
// Only touches owner=knossos entries. User content is NEVER destroyed.
func wipeKnossosOwnedMenaEntries(knossosHome, userClaudeDir string, manifest *provenance.ProvenanceManifest, dryRun bool) {
	// Collect mena entries to identify knossos-producible manifest key patterns.
	// The old pipeline used non-flattened paths; the new pipeline uses flattened paths.
	// Any existing manifest entry matching either pattern is knossos-produced, not user-created.
	menaSourceDir := filepath.Join(knossosHome, "mena")
	sources := []MenaSource{{Path: menaSourceDir}}
	resolution, err := CollectMena(sources, MenaProjectionOptions{Filter: ProjectAll})
	if err != nil {
		return // Best effort: if CollectMena fails, skip wipe
	}

	// Build set of manifest key prefixes that knossos produces (old + new style).
	knossosKeys := make(map[string]bool)

	for sourceKey, entry := range resolution.Entries {
		prefix := "commands/"
		if entry.MenaType == "lego" {
			prefix = "skills/"
		}
		// Old-style directory: "commands/operations/spike/" (prefix)
		knossosKeys[prefix+sourceKey+"/"] = true
		// Old-style collapsed: "commands/operations/spike.md" (exact)
		knossosKeys[prefix+sourceKey+".md"] = true
		// New-style directory: "commands/spike/" (prefix)
		knossosKeys[prefix+entry.FlatName+"/"] = true
	}
	for _, sf := range resolution.Standalones {
		prefix := "commands/"
		if sf.MenaType == "lego" {
			prefix = "skills/"
		}
		// Old-style: "commands/navigation/rite.md"
		dir := filepath.Dir(sf.RelPath)
		base := StripMenaExtension(filepath.Base(sf.RelPath))
		knossosKeys[prefix+filepath.Join(dir, base)] = true
		// New-style (flattened): "commands/rite.md"
		knossosKeys[prefix+sf.FlatName] = true
	}

	// Phase 1: Walk manifest — any commands/ or skills/ entry matching a
	// knossos-producible key pattern gets wiped, regardless of current owner
	// (old pipeline set owner: user for all entries).
	for key := range manifest.Entries {
		if !strings.HasPrefix(key, "commands/") && !strings.HasPrefix(key, "skills/") {
			continue
		}
		if !matchesKnossosKey(key, knossosKeys) {
			continue
		}
		if !dryRun {
			targetPath := filepath.Join(userClaudeDir, key)
			os.Remove(targetPath)
			delete(manifest.Entries, key)
		}
	}

	// Phase 2: Scan disk for untracked orphans. The old pipeline created files
	// that may not have manifest entries (or entries under different keys).
	// Remove any file on disk whose path matches a knossos-producible pattern.
	if !dryRun {
		for _, dir := range []string{"commands", "skills"} {
			targetDir := filepath.Join(userClaudeDir, dir)
			filepath.WalkDir(targetDir, func(path string, d os.DirEntry, wErr error) error {
				if wErr != nil || d.IsDir() {
					return nil
				}
				relPath, err := filepath.Rel(userClaudeDir, path)
				if err != nil {
					return nil
				}
				// Skip files that are still tracked in the manifest (already handled)
				if _, tracked := manifest.Entries[relPath]; tracked {
					return nil
				}
				if matchesKnossosKey(relPath, knossosKeys) {
					os.Remove(path)
				}
				return nil
			})
		}
		// Clean empty parent directories
		cleanEmptyDirs(filepath.Join(userClaudeDir, "commands"))
		cleanEmptyDirs(filepath.Join(userClaudeDir, "skills"))
	}
}

// matchesKnossosKey checks if a manifest key matches any knossos-producible pattern.
// Patterns can be exact matches or directory prefixes (ending with "/").
func matchesKnossosKey(key string, knossosKeys map[string]bool) bool {
	if knossosKeys[key] {
		return true
	}
	for pattern := range knossosKeys {
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(key, pattern) {
			return true
		}
	}
	return false
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
