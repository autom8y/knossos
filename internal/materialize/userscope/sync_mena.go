package userscope

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/materialize/mena"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// syncUserMena syncs mena files from KNOSSOS_HOME/mena and rites/shared/mena to
// ~/.claude/{commands,skills} using the CollectMena pipeline for namespace flattening
// and companion hiding parity with the rite-scope pipeline (SyncMena).
func (s *syncer) syncUserMena(
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
		if s.embeddedMena != nil {
			return s.syncUserMenaFromEmbedded(userClaudeDir, manifest, collisionChecker, opts)
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
	// Sources: platform mena (lowest priority) + shared rite mena (higher priority).
	// Shared rite mena provides cross-rite features (/know, /radar, /research, etc.)
	// that should be available globally, not just within a rite.
	sources := []mena.MenaSource{{Path: sourceDir}}
	sharedMenaDir := filepath.Join(knossosHome, "rites", "shared", "mena")
	if _, err := os.Stat(sharedMenaDir); err == nil {
		sources = append(sources, mena.MenaSource{Path: sharedMenaDir})
	}
	collectOpts := mena.MenaProjectionOptions{
		Filter: mena.ProjectAll,
	}
	resolution, err := mena.CollectMena(sources, collectOpts)
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
			strippedName := mena.StripMenaExtension(filepath.Base(relPath))
			fileDir := filepath.Dir(relPath)
			strippedRel := filepath.Join(fileDir, strippedName)

			// Preserve original stripped path for provenance source tracking
			sourceStrippedRel := strippedRel

			// Manifest key uses flat name
			manifestKey := manifestPrefix + filepath.Join(entry.FlatName, strippedRel)
			targetPath := filepath.Join(targetBaseDir, entry.FlatName, strippedRel)

			// Dromena INDEX.md promotion: top-level INDEX.md → parent-level {flatName}.md
			if entry.MenaType == "dro" && strippedName == "INDEX.md" && fileDir == "." {
				manifestKey = manifestPrefix + entry.FlatName + ".md"
				targetPath = filepath.Join(targetBaseDir, entry.FlatName+".md")
			}

			// Legomena INDEX.md → SKILL.md rename (CC entrypoint convention)
			if entry.MenaType == "lego" && strippedName == "INDEX.md" && fileDir == "." {
				strippedName = "SKILL.md"
				manifestKey = manifestPrefix + filepath.Join(entry.FlatName, "SKILL.md")
				targetPath = filepath.Join(targetBaseDir, entry.FlatName, "SKILL.md")
			}

			// Read source content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Apply companion hiding for dro non-INDEX markdown files
			if entry.MenaType == "dro" && strippedName != "INDEX.md" && strings.HasSuffix(strippedName, ".md") {
				content = mena.InjectCompanionHideFrontmatter(content)
			}

			// Checksum of transformed content (what we'll write)
			sourceChecksum := checksum.Bytes(content)

			return syncUserMenaFile(
				manifestKey, targetPath, content, sourceChecksum,
				filepath.Join("mena", entry.FlatName, sourceStrippedRel),
				manifest, collisionChecker, snapshot, result, opts,
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
			manifest, collisionChecker, snapshot, result, opts,
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
func (s *syncer) syncUserMenaFromEmbedded(
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
	sources := []mena.MenaSource{{Fsys: s.embeddedMena, FsysPath: "mena", IsEmbedded: true}}
	collectOpts := mena.MenaProjectionOptions{
		Filter: mena.ProjectAll,
	}
	resolution, err := mena.CollectMena(sources, collectOpts)
	if err != nil {
		return nil, err
	}

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
			strippedName := mena.StripMenaExtension(filepath.Base(relPath))
			fileDir := filepath.Dir(relPath)
			strippedRel := filepath.Join(fileDir, strippedName)

			// Preserve original stripped path for provenance source tracking
			sourceStrippedRel := strippedRel

			// Manifest key uses flat name
			manifestKey := manifestPrefix + filepath.Join(entry.FlatName, strippedRel)
			targetPath := filepath.Join(targetBaseDir, entry.FlatName, strippedRel)

			// Dromena INDEX.md promotion: top-level INDEX.md → parent-level {flatName}.md
			if entry.MenaType == "dro" && strippedName == "INDEX.md" && fileDir == "." {
				manifestKey = manifestPrefix + entry.FlatName + ".md"
				targetPath = filepath.Join(targetBaseDir, entry.FlatName+".md")
			}

			// Legomena INDEX.md → SKILL.md rename (CC entrypoint convention)
			if entry.MenaType == "lego" && strippedName == "INDEX.md" && fileDir == "." {
				strippedName = "SKILL.md"
				manifestKey = manifestPrefix + filepath.Join(entry.FlatName, "SKILL.md")
				targetPath = filepath.Join(targetBaseDir, entry.FlatName, "SKILL.md")
			}

			// Read embedded content
			content, err := fs.ReadFile(entry.Source.Fsys, path)
			if err != nil {
				return err
			}

			// Apply companion hiding for dro non-INDEX markdown files
			if entry.MenaType == "dro" && strippedName != "INDEX.md" && strings.HasSuffix(strippedName, ".md") {
				content = mena.InjectCompanionHideFrontmatter(content)
			}

			sourceChecksum := checksum.Bytes(content)

			return syncUserMenaFile(
				manifestKey, targetPath, content, sourceChecksum,
				"embedded:mena/"+filepath.Join(entry.FlatName, sourceStrippedRel),
				manifest, collisionChecker, nil, result, opts,
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
		content, err := fs.ReadFile(s.embeddedMena, sf.SrcPath)
		if err != nil {
			return nil, err
		}

		sourceChecksum := checksum.Bytes(content)

		if err := syncUserMenaFile(
			manifestKey, targetPath, content, sourceChecksum,
			"embedded:mena/"+sf.RelPath,
			manifest, collisionChecker, nil, result, opts,
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
				targetChecksum, checksumErr := checksum.File(targetPath)
				if checksumErr != nil {
					slog.Warn("checksum failed, treating as changed", "path", targetPath, "error", checksumErr)
				}
				if checksumErr == nil && targetChecksum == sourceChecksum {
					if !opts.DryRun {
						manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
							provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
						)
					}
					result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
				} else {
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

		// New file, target doesn't exist - write it
		if !opts.DryRun {
			if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
				return err
			}
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
			if !opts.DryRun {
				if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, content, 0644); err != nil {
					return err
				}
				manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
					provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
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
					provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
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
				// Target unchanged, update from source
				if !opts.DryRun {
					if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
						return err
					}
					if err := os.WriteFile(targetPath, content, 0644); err != nil {
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
					if !opts.DryRun {
						if err := paths.EnsureDir(filepath.Dir(targetPath)); err != nil {
							return err
						}
						if err := os.WriteFile(targetPath, content, 0644); err != nil {
							return err
						}
						manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
							provenance.ScopeUser, sourceRelPath, "user-sync", sourceChecksum,
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
	sources := []mena.MenaSource{{Path: menaSourceDir}}
	resolution, err := mena.CollectMena(sources, mena.MenaProjectionOptions{Filter: mena.ProjectAll})
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
		base := mena.StripMenaExtension(filepath.Base(sf.RelPath))
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
			_ = os.Remove(targetPath)
			delete(manifest.Entries, key)
		}
	}

	// Phase 2: Scan disk for untracked orphans. The old pipeline created files
	// that may not have manifest entries (or entries under different keys).
	// Remove any file on disk whose path matches a knossos-producible pattern.
	if !dryRun {
		for _, dir := range []string{"commands", "skills"} {
			targetDir := filepath.Join(userClaudeDir, dir)
			_ = filepath.WalkDir(targetDir, func(path string, d os.DirEntry, wErr error) error {
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
					_ = os.Remove(path)
				}
				return nil
			})
		}
		// Clean empty parent directories
		mena.CleanEmptyDirs(filepath.Join(userClaudeDir, "commands"))
		mena.CleanEmptyDirs(filepath.Join(userClaudeDir, "skills"))
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

// findMenaSource locates the source file for a stripped manifest key.
// Returns the source path and mena type, or ("", "") if not found.
func findMenaSource(sourceDir, strippedRelPath string) (string, string) {
	dir := filepath.Dir(strippedRelPath)
	base := filepath.Base(strippedRelPath)

	// Try original path (no infix)
	candidate := filepath.Join(sourceDir, strippedRelPath)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, mena.DetectMenaType(base)
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

