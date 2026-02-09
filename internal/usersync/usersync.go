package usersync

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// ResourceType identifies the type of user resource.
type ResourceType string

const (
	ResourceAgents ResourceType = "agents"
	ResourceMena   ResourceType = "mena" // Replaces ResourceSkills + ResourceCommands
	ResourceHooks  ResourceType = "hooks"
)

// Singular returns the singular form of the resource type.
func (r ResourceType) Singular() string {
	switch r {
	case ResourceAgents:
		return "agent"
	case ResourceMena:
		return "mena"
	case ResourceHooks:
		return "hook"
	default:
		return string(r)
	}
}

// SourceDir returns the source directory name for the resource type.
func (r ResourceType) SourceDir() string {
	switch r {
	case ResourceMena:
		return "mena"
	case ResourceAgents:
		return "agents"
	case ResourceHooks:
		return "hooks"
	default:
		return string(r)
	}
}

// RiteSubDir returns the subdirectory name within rites for the resource type.
func (r ResourceType) RiteSubDir() string {
	if r == ResourceMena {
		return "mena"
	}
	return string(r)
}


// Options configures sync behavior.
type Options struct {
	DryRun  bool // Preview changes without applying
	Recover bool // Adopt existing files matching knossos
	Force   bool // Overwrite diverged files
	Verbose bool // Enable verbose logging
}

// Result contains sync operation outcome.
type Result struct {
	SyncedAt time.Time    `json:"synced_at"`
	Resource ResourceType `json:"resource"`
	DryRun   bool         `json:"dry_run"`
	Source   string       `json:"source"`
	Target   string       `json:"target"`
	Changes  Changes      `json:"changes"`
	Summary  Summary      `json:"summary"`
}

// Changes categorizes sync outcomes by file.
type Changes struct {
	Added     []string       `json:"added"`
	Updated   []string       `json:"updated"`
	Skipped   []SkippedEntry `json:"skipped"`
	Unchanged []string       `json:"unchanged"`
}

// SkippedEntry explains why a file was skipped.
type SkippedEntry struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// Summary provides aggregate counts.
type Summary struct {
	Added      int `json:"added"`
	Updated    int `json:"updated"`
	Skipped    int `json:"skipped"`
	Unchanged  int `json:"unchanged"`
	Collisions int `json:"collisions"`
}

// Syncer handles user resource synchronization.
type Syncer struct {
	resourceType      ResourceType
	sourceDir         string
	targetDir         string // Used by agents, hooks (single target)
	targetCommandsDir string // Used by mena (dromena target)
	targetSkillsDir   string // Used by mena (legomena target)
	manifestPath      string
	collisionChecker  *CollisionChecker
	nested            bool // true for mena, hooks
}

// NewSyncer creates a syncer for the given resource type.
func NewSyncer(resourceType ResourceType) (*Syncer, error) {
	knossosHome := config.KnossosHome()
	if knossosHome == "" {
		return nil, ErrKnossosHomeNotSet
	}

	s := &Syncer{
		resourceType: resourceType,
	}

	// All resource types now use the unified provenance manifest
	userClaudeDir := paths.UserClaudeDir()
	s.manifestPath = provenance.UserManifestPath(userClaudeDir)

	switch resourceType {
	case ResourceAgents:
		s.sourceDir = filepath.Join(knossosHome, "agents")
		s.targetDir = paths.UserAgentsDir()
		s.nested = false
	case ResourceMena:
		s.sourceDir = filepath.Join(knossosHome, "mena")
		s.targetCommandsDir = paths.UserCommandsDir()
		s.targetSkillsDir = paths.UserSkillsDir()
		s.nested = true
	case ResourceHooks:
		s.sourceDir = filepath.Join(knossosHome, "hooks")
		s.targetDir = paths.UserHooksDir()
		s.nested = true
	default:
		return nil, ErrInvalidResourceType
	}

	// Initialize collision checker (no project context for user sync)
	s.collisionChecker = NewCollisionChecker(resourceType, s.nested, "")

	return s, nil
}

// NewSyncerWithPaths creates a syncer with explicit paths (for testing).
// For ResourceMena, targetDir is used as the commands target directory.
// Use NewMenaSyncerWithPaths for explicit dual-target testing.
func NewSyncerWithPaths(resourceType ResourceType, sourceDir, targetDir, manifestPath string) *Syncer {
	nested := resourceType != ResourceAgents
	s := &Syncer{
		resourceType:     resourceType,
		sourceDir:        sourceDir,
		targetDir:        targetDir,
		manifestPath:     manifestPath,
		collisionChecker: NewCollisionChecker(resourceType, nested, ""),
		nested:           nested,
	}
	if resourceType == ResourceMena {
		s.targetCommandsDir = targetDir
		s.targetSkillsDir = targetDir // For simple tests, both point to same dir
		s.targetDir = ""              // Mena does not use single targetDir
	}
	return s
}

// NewMenaSyncerWithPaths creates a mena syncer with explicit dual-target paths (for testing).
func NewMenaSyncerWithPaths(sourceDir, targetCommandsDir, targetSkillsDir, manifestPath string) *Syncer {
	return &Syncer{
		resourceType:      ResourceMena,
		sourceDir:         sourceDir,
		targetCommandsDir: targetCommandsDir,
		targetSkillsDir:   targetSkillsDir,
		manifestPath:      manifestPath,
		collisionChecker:  NewCollisionChecker(ResourceMena, true, ""),
		nested:            true,
	}
}

// Sync performs the synchronization operation.
func (s *Syncer) Sync(opts Options) (*Result, error) {
	// Determine target display string
	target := s.targetDir
	if s.resourceType == ResourceMena {
		target = s.targetCommandsDir + " + " + s.targetSkillsDir
	}

	result := &Result{
		SyncedAt: time.Now().UTC(),
		Resource: s.resourceType,
		DryRun:   opts.DryRun,
		Source:   s.sourceDir,
		Target:   target,
		Changes: Changes{
			Added:     []string{},
			Updated:   []string{},
			Skipped:   []SkippedEntry{},
			Unchanged: []string{},
		},
	}

	// Check source directory exists
	if _, err := os.Stat(s.sourceDir); os.IsNotExist(err) {
		return nil, ErrSourceNotFound(s.sourceDir)
	}

	// Ensure target directory exists
	if !opts.DryRun {
		if s.resourceType == ResourceMena {
			if err := paths.EnsureDir(s.targetCommandsDir); err != nil {
				return nil, ErrTargetCreateFailed(s.targetCommandsDir, err)
			}
			if err := paths.EnsureDir(s.targetSkillsDir); err != nil {
				return nil, ErrTargetCreateFailed(s.targetSkillsDir, err)
			}
		} else {
			if err := paths.EnsureDir(s.targetDir); err != nil {
				return nil, ErrTargetCreateFailed(s.targetDir, err)
			}
		}
	}

	// Load or create manifest
	manifest, err := s.loadManifest()
	if err != nil {
		return nil, err
	}

	// Handle recovery mode first
	if opts.Recover {
		if err := s.recover(manifest, result, opts); err != nil {
			return nil, err
		}
	}

	// Sync source files to target
	if err := s.syncFiles(manifest, result, opts); err != nil {
		return nil, err
	}

	// Update manifest
	if !opts.DryRun {
		manifest.LastSync = result.SyncedAt
		if err := s.saveManifest(manifest); err != nil {
			return nil, err
		}
		s.cleanupOldManifests()
	}

	// Calculate summary
	result.Summary = Summary{
		Added:      len(result.Changes.Added),
		Updated:    len(result.Changes.Updated),
		Skipped:    len(result.Changes.Skipped),
		Unchanged:  len(result.Changes.Unchanged),
		Collisions: s.countCollisions(result.Changes.Skipped),
	}

	return result, nil
}

// Status returns what would be synced without actually syncing.
func (s *Syncer) Status() (*Result, error) {
	return s.Sync(Options{DryRun: true})
}


// recover adopts existing target files that match knossos sources.
func (s *Syncer) recover(manifest *provenance.ProvenanceManifest, result *Result, opts Options) error {
	if s.resourceType == ResourceMena {
		// For mena, walk both commands and skills directories
		if err := s.recoverDir(s.targetCommandsDir, "commands", manifest, result, opts); err != nil {
			return err
		}
		return s.recoverDir(s.targetSkillsDir, "skills", manifest, result, opts)
	}
	return s.recoverDir(s.targetDir, "", manifest, result, opts)
}

// recoverDir walks a single target directory recovering untracked files.
// For mena resources, menaTarget should be "commands" or "skills".
func (s *Syncer) recoverDir(recoverDir, menaTarget string, manifest *provenance.ProvenanceManifest, result *Result, opts Options) error {
	// Check if target directory exists
	if _, err := os.Stat(recoverDir); os.IsNotExist(err) {
		return nil // Nothing to recover
	}

	// Walk target directory looking for untracked files
	return filepath.WalkDir(recoverDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(recoverDir, path)
		manifestKey := relPath
		if !s.nested {
			manifestKey = filepath.Base(relPath)
		}

		// Add resource type prefix
		fullManifestKey := s.prefixManifestKey(manifestKey)
		if s.resourceType == ResourceMena {
			if menaTarget == "commands" {
				fullManifestKey = "commands/" + manifestKey
			} else {
				fullManifestKey = "skills/" + manifestKey
			}
		}

		// Skip if already in manifest
		if _, exists := manifest.Entries[fullManifestKey]; exists {
			return nil
		}

		// For mena, we need to find the source file which may have .dro or .lego infix.
		// The target file has the stripped name; we must search for the original source.
		sourcePath := filepath.Join(s.sourceDir, relPath)
		if s.resourceType == ResourceMena {
			// Try to find the source file with any mena extension
			sourcePath, _ = s.findMenaSource(relPath)
			if sourcePath == "" {
				// Not in knossos source - mark as user
				if !opts.DryRun {
					targetChecksum, _ := ComputeFileChecksum(path)
					manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerUser,
						Scope:      provenance.ScopeUser,
						Checksum:   targetChecksum,
						LastSynced: result.SyncedAt,
					}
				}
				return nil
			}
		} else if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			// Not in knossos - mark as user
			if !opts.DryRun {
				targetChecksum, _ := ComputeFileChecksum(path)
				manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerUser,
					Scope:      provenance.ScopeUser,
					Checksum:   targetChecksum,
					LastSynced: result.SyncedAt,
				}
			}
			return nil
		}

		// Compare checksums
		sourceChecksum, _ := ComputeFileChecksum(sourcePath)
		targetChecksum, _ := ComputeFileChecksum(path)
		sourceRelPath, _ := filepath.Rel(s.sourceDir, sourcePath)

		if !opts.DryRun {
			if sourceChecksum == targetChecksum {
				manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeUser,
					SourcePath: sourceRelPath,
					SourceType: "user-sync",
					Checksum:   sourceChecksum,
					LastSynced: result.SyncedAt,
				}
			} else {
				manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerUser,
					Scope:      provenance.ScopeUser,
					Checksum:   targetChecksum,
					LastSynced: result.SyncedAt,
				}
			}
		}

		return nil
	})
}

// findMenaSource locates the source file for a stripped manifest key.
// It checks for .dro and .lego variants of the filename.
// Returns the source path and detected mena type, or ("", "") if not found.
func (s *Syncer) findMenaSource(strippedRelPath string) (string, string) {
	dir := filepath.Dir(strippedRelPath)
	base := filepath.Base(strippedRelPath)

	// Try original path (no infix)
	candidate := filepath.Join(s.sourceDir, strippedRelPath)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, materialize.DetectMenaType(base)
	}

	// Try .dro variant: insert .dro before the last extension
	ext := filepath.Ext(base)
	nameNoExt := base[:len(base)-len(ext)]
	droName := nameNoExt + ".dro" + ext
	candidate = filepath.Join(s.sourceDir, dir, droName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, "dro"
	}

	// Try .lego variant
	legoName := nameNoExt + ".lego" + ext
	candidate = filepath.Join(s.sourceDir, dir, legoName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, "lego"
	}

	return "", ""
}


// copyFile copies a file preserving permissions.
func (s *Syncer) copyFile(src, dst string) error {
	// Ensure parent directory exists
	if err := paths.EnsureDir(filepath.Dir(dst)); err != nil {
		return ErrCopy(src, dst, err)
	}

	// Read source
	content, err := os.ReadFile(src)
	if err != nil {
		return ErrCopy(src, dst, err)
	}

	// Get source permissions
	info, err := os.Stat(src)
	if err != nil {
		return ErrCopy(src, dst, err)
	}

	perm := info.Mode()

	// For hooks, ensure executable bit for scripts
	if s.resourceType == ResourceHooks && isExecutable(src) && perm&0111 == 0 {
		perm |= 0755
	}

	// Write destination with appropriate permissions
	if err := os.WriteFile(dst, content, perm); err != nil {
		return ErrCopy(src, dst, err)
	}

	return nil
}

// countCollisions counts collision entries in skipped list.
func (s *Syncer) countCollisions(skipped []SkippedEntry) int {
	count := 0
	for _, entry := range skipped {
		if strings.Contains(entry.Reason, "collision") {
			count++
		}
	}
	return count
}

// SourceDir returns the source directory path.
func (s *Syncer) SourceDir() string {
	return s.sourceDir
}

// TargetDir returns the target directory path.
// For ResourceMena, returns empty string (use TargetCommandsDir/TargetSkillsDir instead).
func (s *Syncer) TargetDir() string {
	return s.targetDir
}

// TargetCommandsDir returns the commands target directory (mena only).
func (s *Syncer) TargetCommandsDir() string {
	return s.targetCommandsDir
}

// TargetSkillsDir returns the skills target directory (mena only).
func (s *Syncer) TargetSkillsDir() string {
	return s.targetSkillsDir
}

// ManifestPath returns the manifest file path.
func (s *Syncer) ManifestPath() string {
	return s.manifestPath
}

// prefixManifestKey adds the resource type prefix to the manifest key.
func (s *Syncer) prefixManifestKey(key string) string {
	switch s.resourceType {
	case ResourceAgents:
		return "agents/" + key
	case ResourceMena:
		// Mena keys need the target directory prefix (commands/ or skills/)
		// This is handled in syncFiles where we know the mena target
		return key
	case ResourceHooks:
		return "hooks/" + key
	default:
		return key
	}
}

// keyToTargetPath converts a manifest key back to a target path.
func (s *Syncer) keyToTargetPath(key string) string {
	// Strip resource prefix
	switch s.resourceType {
	case ResourceAgents:
		key = strings.TrimPrefix(key, "agents/")
		return filepath.Join(s.targetDir, key)
	case ResourceMena:
		if strings.HasPrefix(key, "commands/") {
			key = strings.TrimPrefix(key, "commands/")
			return filepath.Join(s.targetCommandsDir, key)
		} else if strings.HasPrefix(key, "skills/") {
			key = strings.TrimPrefix(key, "skills/")
			return filepath.Join(s.targetSkillsDir, key)
		}
		return ""
	case ResourceHooks:
		key = strings.TrimPrefix(key, "hooks/")
		return filepath.Join(s.targetDir, key)
	default:
		return ""
	}
}

// resourcePrefix returns the manifest key prefix for this resource type.
func (s *Syncer) resourcePrefix() string {
	switch s.resourceType {
	case ResourceAgents:
		return "agents/"
	case ResourceMena:
		return "" // Mena has two prefixes: commands/ and skills/
	case ResourceHooks:
		return "hooks/"
	default:
		return ""
	}
}

// syncFiles iterates source files and syncs to target.
func (s *Syncer) syncFiles(manifest *provenance.ProvenanceManifest, result *Result, opts Options) error {
	// Phase 1: Snapshot current keys for this resource type
	prefix := s.resourcePrefix()
	existingKeys := make(map[string]bool)

	for key, entry := range manifest.Entries {
		// For mena, check both prefixes
		if s.resourceType == ResourceMena {
			if strings.HasPrefix(key, "commands/") || strings.HasPrefix(key, "skills/") {
				if entry.Owner == provenance.OwnerKnossos {
					existingKeys[key] = false
				}
			}
		} else if prefix != "" && strings.HasPrefix(key, prefix) {
			if entry.Owner == provenance.OwnerKnossos {
				existingKeys[key] = false
			}
		}
	}

	// Phase 2: Walk source and sync
	err := filepath.WalkDir(s.sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (we process files)
		if d.IsDir() {
			return nil
		}

		// Compute relative path for manifest key
		relPath, err := filepath.Rel(s.sourceDir, path)
		if err != nil {
			return err
		}

		// For flat resources, use just the filename
		manifestKey := relPath
		if !s.nested {
			manifestKey = filepath.Base(relPath)
		}

		// For mena: strip extension from manifest key and route to correct target.
		// Manifest keys use STRIPPED filenames (e.g., "commit/INDEX.md" not "commit/INDEX.dro.md").
		var menaTarget string
		if s.resourceType == ResourceMena {
			dir := filepath.Dir(manifestKey)
			base := materialize.StripMenaExtension(filepath.Base(manifestKey))
			manifestKey = filepath.Join(dir, base)
			// Use ORIGINAL filename for routing
			menaTarget = materialize.RouteMenaFile(filepath.Base(relPath))
		}

		// Add resource type prefix to manifest key
		fullManifestKey := s.prefixManifestKey(manifestKey)
		if s.resourceType == ResourceMena {
			// Mena keys need the target directory prefix
			if menaTarget == "commands" {
				fullManifestKey = "commands/" + manifestKey
			} else {
				fullManifestKey = "skills/" + manifestKey
			}
		}

		// Mark this key as seen
		if _, tracked := existingKeys[fullManifestKey]; tracked {
			existingKeys[fullManifestKey] = true
		}

		// Check for rite collision
		if collision, riteName := s.collisionChecker.CheckCollision(manifestKey); collision {
			result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
				Name:   manifestKey,
				Reason: "collision with rite " + s.resourceType.Singular() + " (" + riteName + ")",
			})
			return nil
		}

		// Calculate source checksum
		sourceChecksum, err := ComputeFileChecksum(path)
		if err != nil {
			return ErrChecksum(path, err)
		}

		// Check existing manifest entry
		entry, exists := manifest.Entries[fullManifestKey]

		// Determine target path based on resource type and mena routing
		targetBase := s.targetDir
		if s.resourceType == ResourceMena {
			if menaTarget == "commands" {
				targetBase = s.targetCommandsDir
			} else {
				targetBase = s.targetSkillsDir
			}
		}
		targetPath := filepath.Join(targetBase, manifestKey)

		// Compute source path relative to knossos home for SourcePath field
		sourceRelPath, _ := filepath.Rel(s.sourceDir, path)

		if !exists {
			// New file - check if target exists (untracked)
			if _, err := os.Stat(targetPath); err == nil {
				// Target exists but not in manifest
				if opts.Recover {
					targetChecksum, _ := ComputeFileChecksum(targetPath)
					if targetChecksum == sourceChecksum {
						// Exact match - adopt as knossos
						if !opts.DryRun {
							manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerKnossos,
								Scope:      provenance.ScopeUser,
								SourcePath: sourceRelPath,
								SourceType: "user-sync",
								Checksum:   sourceChecksum,
								LastSynced: result.SyncedAt,
							}
						}
						result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
					} else {
						// Different - adopt as user (target has been modified)
						if !opts.DryRun {
							manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerUser,
								Scope:      provenance.ScopeUser,
								Checksum:   targetChecksum,
								LastSynced: result.SyncedAt,
							}
						}
						result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
							Name:   manifestKey,
							Reason: "adopted as user (local modifications)",
						})
					}
					return nil
				}
				// Not recovering - skip as user-created
				if !opts.DryRun {
					targetChecksum, _ := ComputeFileChecksum(targetPath)
					manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
						Owner:      provenance.OwnerUser,
						Scope:      provenance.ScopeUser,
						Checksum:   targetChecksum,
						LastSynced: result.SyncedAt,
					}
				}
				result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
					Name:   manifestKey,
					Reason: "user-created",
				})
				return nil
			}

			// New file, target doesn't exist - add it
			if !opts.DryRun {
				if err := s.copyFile(path, targetPath); err != nil {
					return err
				}
				manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeUser,
					SourcePath: sourceRelPath,
					SourceType: "user-sync",
					Checksum:   sourceChecksum,
					LastSynced: result.SyncedAt,
				}
			}
			result.Changes.Added = append(result.Changes.Added, manifestKey)
			return nil
		}

		// Existing entry
		switch entry.Owner {
		case provenance.OwnerUser, provenance.OwnerUntracked:
			// Never touch user-created files
			result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
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
				targetChecksum, _ := ComputeFileChecksum(targetPath)
				if targetChecksum == entry.Checksum {
					// Target unchanged, update from source
					if !opts.DryRun {
						if err := s.copyFile(path, targetPath); err != nil {
							return err
						}
						manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
							Owner:      provenance.OwnerKnossos,
							Scope:      provenance.ScopeUser,
							SourcePath: sourceRelPath,
							SourceType: "user-sync",
							Checksum:   sourceChecksum,
							LastSynced: result.SyncedAt,
						}
					}
					result.Changes.Updated = append(result.Changes.Updated, manifestKey)
				} else {
					// Target diverged - check if --force
					if opts.Force {
						// Force overwrite
						if !opts.DryRun {
							if err := s.copyFile(path, targetPath); err != nil {
								return err
							}
							manifest.Entries[fullManifestKey] = &provenance.ProvenanceEntry{
								Owner:      provenance.OwnerKnossos,
								Scope:      provenance.ScopeUser,
								SourcePath: sourceRelPath,
								SourceType: "user-sync",
								Checksum:   sourceChecksum,
								LastSynced: result.SyncedAt,
							}
						}
						result.Changes.Updated = append(result.Changes.Updated, manifestKey)
					} else {
						// Skip diverged without force
						result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
							Name:   manifestKey,
							Reason: "diverged (use --force to overwrite)",
						})
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Phase 3: Orphan removal
	if !opts.DryRun {
		for key, seen := range existingKeys {
			if !seen {
				s.removeOrphan(key, manifest)
			}
		}
	}

	return nil
}

// removeOrphan removes an orphaned knossos-owned entry.
func (s *Syncer) removeOrphan(key string, manifest *provenance.ProvenanceManifest) {
	entry := manifest.Entries[key]
	if entry == nil || entry.Owner != provenance.OwnerKnossos {
		return // Safety: only remove knossos-owned orphans
	}

	// Determine target path from key
	targetPath := s.keyToTargetPath(key)
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
