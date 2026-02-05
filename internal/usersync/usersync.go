package usersync

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
)

// ResourceType identifies the type of user resource.
type ResourceType string

const (
	ResourceAgents   ResourceType = "agents"
	ResourceSkills   ResourceType = "skills"
	ResourceCommands ResourceType = "commands"
	ResourceHooks    ResourceType = "hooks"
)

// Singular returns the singular form of the resource type.
func (r ResourceType) Singular() string {
	switch r {
	case ResourceAgents:
		return "agent"
	case ResourceSkills:
		return "skill"
	case ResourceCommands:
		return "command"
	case ResourceHooks:
		return "hook"
	default:
		return string(r)
	}
}

// SourceDir returns the source directory name for the resource type.
func (r ResourceType) SourceDir() string {
	if r == ResourceCommands {
		return "mena"
	}
	return "user-" + string(r)
}

// SourceType identifies the origin of a synced resource.
type SourceType string

const (
	SourceRoster   SourceType = "roster"          // Synced from roster, unchanged
	SourceDiverged SourceType = "roster-diverged" // From roster but locally modified
	SourceUser     SourceType = "user"            // User-created, not in roster
)

// Options configures sync behavior.
type Options struct {
	DryRun  bool // Preview changes without applying
	Recover bool // Adopt existing files matching roster
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
	resourceType     ResourceType
	sourceDir        string
	targetDir        string
	manifestPath     string
	collisionChecker *CollisionChecker
	nested           bool // true for skills, commands, hooks
}

// NewSyncer creates a syncer for the given resource type.
func NewSyncer(resourceType ResourceType) (*Syncer, error) {
	knossosHome := config.KnossosHome()
	if knossosHome == "" {
		return nil, ErrKnossosHomeNotSet
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	s := &Syncer{
		resourceType: resourceType,
	}

	switch resourceType {
	case ResourceAgents:
		s.sourceDir = filepath.Join(knossosHome, "user-agents")
		s.targetDir = filepath.Join(homeDir, ".claude", "agents")
		s.manifestPath = filepath.Join(homeDir, ".claude", "USER_AGENT_MANIFEST.json")
		s.nested = false
	case ResourceSkills:
		s.sourceDir = filepath.Join(knossosHome, "user-skills")
		s.targetDir = filepath.Join(homeDir, ".claude", "skills")
		s.manifestPath = filepath.Join(homeDir, ".claude", "USER_SKILL_MANIFEST.json")
		s.nested = true
	case ResourceCommands:
		s.sourceDir = filepath.Join(knossosHome, "mena")
		s.targetDir = filepath.Join(homeDir, ".claude", "commands")
		s.manifestPath = filepath.Join(homeDir, ".claude", "USER_COMMAND_MANIFEST.json")
		s.nested = true
	case ResourceHooks:
		s.sourceDir = filepath.Join(knossosHome, "user-hooks")
		s.targetDir = filepath.Join(homeDir, ".claude", "hooks")
		s.manifestPath = filepath.Join(homeDir, ".claude", "USER_HOOKS_MANIFEST.json")
		s.nested = true
	default:
		return nil, ErrInvalidResourceType
	}

	// Initialize collision checker
	s.collisionChecker = NewCollisionChecker(resourceType, s.nested)

	return s, nil
}

// NewSyncerWithPaths creates a syncer with explicit paths (for testing).
func NewSyncerWithPaths(resourceType ResourceType, sourceDir, targetDir, manifestPath string) *Syncer {
	nested := resourceType != ResourceAgents
	return &Syncer{
		resourceType:     resourceType,
		sourceDir:        sourceDir,
		targetDir:        targetDir,
		manifestPath:     manifestPath,
		collisionChecker: NewCollisionChecker(resourceType, nested),
		nested:           nested,
	}
}

// Sync performs the synchronization operation.
func (s *Syncer) Sync(opts Options) (*Result, error) {
	result := &Result{
		SyncedAt: time.Now().UTC(),
		Resource: s.resourceType,
		DryRun:   opts.DryRun,
		Source:   s.sourceDir,
		Target:   s.targetDir,
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
		if err := paths.EnsureDir(s.targetDir); err != nil {
			return nil, ErrTargetCreateFailed(s.targetDir, err)
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

// syncFiles iterates source files and syncs to target.
func (s *Syncer) syncFiles(manifest *Manifest, result *Result, opts Options) error {
	return filepath.WalkDir(s.sourceDir, func(path string, d os.DirEntry, err error) error {
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
		entry, exists := manifest.Entries[manifestKey]
		targetPath := filepath.Join(s.targetDir, manifestKey)

		if !exists {
			// New file - check if target exists (untracked)
			if _, err := os.Stat(targetPath); err == nil {
				// Target exists but not in manifest
				if opts.Recover {
					targetChecksum, _ := ComputeFileChecksum(targetPath)
					if targetChecksum == sourceChecksum {
						// Exact match - adopt as roster
						if !opts.DryRun {
							manifest.Entries[manifestKey] = Entry{
								Source:      SourceRoster,
								InstalledAt: result.SyncedAt,
								Checksum:    sourceChecksum,
							}
						}
						result.Changes.Unchanged = append(result.Changes.Unchanged, manifestKey)
					} else {
						// Different - adopt as diverged
						if !opts.DryRun {
							manifest.Entries[manifestKey] = Entry{
								Source:      SourceDiverged,
								InstalledAt: result.SyncedAt,
								Checksum:    targetChecksum,
							}
						}
						result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
							Name:   manifestKey,
							Reason: "adopted as diverged (local modifications)",
						})
					}
					return nil
				}
				// Not recovering - skip as user-created
				if !opts.DryRun {
					manifest.Entries[manifestKey] = Entry{
						Source:      SourceUser,
						InstalledAt: result.SyncedAt,
						Checksum:    "",
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
				manifest.Entries[manifestKey] = Entry{
					Source:      SourceRoster,
					InstalledAt: result.SyncedAt,
					Checksum:    sourceChecksum,
				}
			}
			result.Changes.Added = append(result.Changes.Added, manifestKey)
			return nil
		}

		// Existing entry
		switch entry.Source {
		case SourceUser:
			// Never touch user-created files
			result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
				Name:   manifestKey,
				Reason: "user-created",
			})

		case SourceDiverged:
			if opts.Force {
				// Force overwrite
				if !opts.DryRun {
					if err := s.copyFile(path, targetPath); err != nil {
						return err
					}
					manifest.Entries[manifestKey] = Entry{
						Source:      SourceRoster,
						InstalledAt: result.SyncedAt,
						Checksum:    sourceChecksum,
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

		case SourceRoster:
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
						manifest.Entries[manifestKey] = Entry{
							Source:      SourceRoster,
							InstalledAt: result.SyncedAt,
							Checksum:    sourceChecksum,
						}
					}
					result.Changes.Updated = append(result.Changes.Updated, manifestKey)
				} else {
					// Target diverged - mark as diverged
					if !opts.DryRun {
						manifest.Entries[manifestKey] = Entry{
							Source:      SourceDiverged,
							InstalledAt: entry.InstalledAt,
							Checksum:    targetChecksum,
						}
					}
					result.Changes.Skipped = append(result.Changes.Skipped, SkippedEntry{
						Name:   manifestKey,
						Reason: "diverged (local modifications)",
					})
				}
			}
		}

		return nil
	})
}

// recover adopts existing target files that match roster sources.
func (s *Syncer) recover(manifest *Manifest, result *Result, opts Options) error {
	// Check if target directory exists
	if _, err := os.Stat(s.targetDir); os.IsNotExist(err) {
		return nil // Nothing to recover
	}

	// Walk target directory looking for untracked files
	return filepath.WalkDir(s.targetDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(s.targetDir, path)
		manifestKey := relPath
		if !s.nested {
			manifestKey = filepath.Base(relPath)
		}

		// Skip if already in manifest
		if _, exists := manifest.Entries[manifestKey]; exists {
			return nil
		}

		// Check if source exists
		sourcePath := filepath.Join(s.sourceDir, relPath)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			// Not in roster - mark as user
			if !opts.DryRun {
				targetChecksum, _ := ComputeFileChecksum(path)
				manifest.Entries[manifestKey] = Entry{
					Source:      SourceUser,
					InstalledAt: result.SyncedAt,
					Checksum:    targetChecksum,
				}
			}
			return nil
		}

		// Compare checksums
		sourceChecksum, _ := ComputeFileChecksum(sourcePath)
		targetChecksum, _ := ComputeFileChecksum(path)

		if !opts.DryRun {
			if sourceChecksum == targetChecksum {
				manifest.Entries[manifestKey] = Entry{
					Source:      SourceRoster,
					InstalledAt: result.SyncedAt,
					Checksum:    sourceChecksum,
				}
			} else {
				manifest.Entries[manifestKey] = Entry{
					Source:      SourceDiverged,
					InstalledAt: result.SyncedAt,
					Checksum:    targetChecksum,
				}
			}
		}

		return nil
	})
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
func (s *Syncer) TargetDir() string {
	return s.targetDir
}

// ManifestPath returns the manifest file path.
func (s *Syncer) ManifestPath() string {
	return s.manifestPath
}
