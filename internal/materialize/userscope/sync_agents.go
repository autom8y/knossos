package userscope

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// syncUserResourceFromEmbedded syncs agents from an embedded filesystem
// when KNOSSOS_HOME is unavailable. Simplified version: no recovery mode,
// no orphan removal (embedded source is authoritative).
func (s *syncer) syncUserResourceFromEmbedded(
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

		// Check existing manifest entry
		entry, exists := manifest.Entries[manifestKey]

		if !exists {
			// Check if target exists (untracked)
			if _, statErr := os.Stat(targetPath); statErr == nil {
				// Exists untracked — mark as user-created, don't overwrite
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

			// New file — write it
			if !opts.DryRun {
				if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, content, 0644); err != nil {
					return err
				}
				manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
					provenance.ScopeUser, "embedded:"+path, "embedded", sourceChecksum,
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
				if !opts.DryRun {
					if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
						return err
					}
					if err := os.WriteFile(targetPath, content, 0644); err != nil {
						return err
					}
					manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
						provenance.ScopeUser, "embedded:"+path, "embedded", sourceChecksum,
					)
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
					manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
						provenance.ScopeUser, "embedded:"+path, "embedded", sourceChecksum,
					)
				}
				result.Changes.Added = append(result.Changes.Added, manifestKey)
			} else if entry.Checksum == sourceChecksum {
				result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
			} else {
				targetChecksum, checksumErr := checksum.File(targetPath)
				if checksumErr != nil {
					slog.Warn("checksum failed, treating as changed", "path", targetPath, "error", checksumErr)
				}
				if targetChecksum == entry.Checksum {
					// Target unchanged, update from embedded source
					if !opts.DryRun {
						if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
							return err
						}
						if err := os.WriteFile(targetPath, content, 0644); err != nil {
							return err
						}
						manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
							provenance.ScopeUser, "embedded:"+path, "embedded", sourceChecksum,
						)
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
							manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
								provenance.ScopeUser, "embedded:"+path, "embedded", sourceChecksum,
							)
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
