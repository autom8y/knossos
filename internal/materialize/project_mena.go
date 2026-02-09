package materialize

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/provenance"
)

// StripMenaExtension removes the .dro or .lego infix from a filename.
// Examples:
//
//	"INDEX.dro.md"      -> "INDEX.md"
//	"INDEX.lego.md"     -> "INDEX.md"
//	"commit.dro.md"     -> "commit.md"
//	"prompting.lego.md" -> "prompting.md"
//	"helper.md"         -> "helper.md"    (no infix, unchanged)
//	"README.md"         -> "README.md"    (no infix, unchanged)
//	"data.json"         -> "data.json"    (no infix, unchanged)
//
// Only the first infix is stripped (handles pathological "foo.dro.dro.md").
func StripMenaExtension(filename string) string {
	if strings.Contains(filename, ".dro.") {
		return strings.Replace(filename, ".dro.", ".", 1)
	}
	if strings.Contains(filename, ".lego.") {
		return strings.Replace(filename, ".lego.", ".", 1)
	}
	return filename
}

// RouteMenaFile determines whether a file routes to commands/ or skills/.
// Returns "commands" or "skills".
func RouteMenaFile(filename string) string {
	menaType := DetectMenaType(filename)
	if menaType == "lego" {
		return "skills"
	}
	return "commands"
}

// MenaSource represents a source for mena files. It can be either a
// filesystem path or an embedded FS path.
type MenaSource struct {
	Path       string // Filesystem path (for os-based sources)
	Fsys       fs.FS  // Embedded filesystem (nil for os-based sources)
	FsysPath   string // Path within Fsys (e.g., "rites/shared/mena")
	IsEmbedded bool
}

// MenaProjectionMode controls whether projection is additive or destructive.
type MenaProjectionMode int

const (
	// MenaProjectionAdditive adds/updates files without removing unmanaged content.
	// Used by usersync (ari sync user mena).
	MenaProjectionAdditive MenaProjectionMode = iota

	// MenaProjectionDestructive wipes target commands/ and skills/ directories
	// before projecting. Used by materialize (ari rite start).
	MenaProjectionDestructive
)

// MenaFilter controls which mena types to project.
type MenaFilter int

const (
	ProjectDro  MenaFilter = 1 << iota // Project dromena only (commands/)
	ProjectLego                        // Project legomena only (skills/)
	ProjectAll  = ProjectDro | ProjectLego
)

// MenaProjectionOptions configures the projection operation.
type MenaProjectionOptions struct {
	Mode   MenaProjectionMode
	Filter MenaFilter

	// TargetCommandsDir is the absolute path to the commands/ output directory.
	TargetCommandsDir string

	// TargetSkillsDir is the absolute path to the skills/ output directory.
	TargetSkillsDir string

	// Collector records provenance at write time. If nil, provenance is not recorded.
	Collector provenance.Collector

	// ProjectRoot is the project root for computing relative source paths.
	// Required when Collector is non-nil.
	ProjectRoot string
}

// MenaProjectionResult reports what the projection did.
type MenaProjectionResult struct {
	CommandsProjected []string // Relative paths of files written to commands/
	SkillsProjected   []string // Relative paths of files written to skills/
}

// menaCollectedEntry represents a leaf mena directory collected for routing.
type menaCollectedEntry struct {
	source MenaSource
	name   string
}

// menaStandaloneFile represents a standalone file in a grouping directory.
type menaStandaloneFile struct {
	srcPath string
	relPath string // e.g., "navigation/rite.dro.md"
}

// ReadMenaFrontmatterFromDir reads the INDEX file from a filesystem directory,
// parses its YAML frontmatter, and returns the result.
// Returns a zero-value MenaFrontmatter if the INDEX file has no
// frontmatter or if parsing fails (with a logged warning for parse failures).
func ReadMenaFrontmatterFromDir(dirPath string) MenaFrontmatter {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return MenaFrontmatter{}
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			data, err := os.ReadFile(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				return MenaFrontmatter{}
			}
			return parseMenaFrontmatterBytes(data)
		}
	}
	return MenaFrontmatter{}
}

// ReadMenaFrontmatterFromFile reads a standalone mena file and parses its
// YAML frontmatter. Returns a zero-value MenaFrontmatter if no frontmatter
// is present or parsing fails.
func ReadMenaFrontmatterFromFile(filePath string) MenaFrontmatter {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return MenaFrontmatter{}
	}
	return parseMenaFrontmatterBytes(data)
}

// parseMenaFrontmatterBytes extracts YAML frontmatter from raw file bytes.
// Returns a zero-value MenaFrontmatter if no frontmatter delimiters are found
// or if YAML parsing fails. Parse failures are silent (the entry is treated
// as unscoped per EC-7 in the PRD).
func parseMenaFrontmatterBytes(data []byte) MenaFrontmatter {
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		return MenaFrontmatter{}
	}

	// Find closing delimiter — searchStart skips past the opening "---\n" or "---\r\n"
	var endIndex int
	searchStart := 4
	if bytes.HasPrefix(data, []byte("---\r\n")) {
		searchStart = 5
	}
	if idx := bytes.Index(data[searchStart:], []byte("\n---\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\r\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\r\n---\n")); idx != -1 {
		endIndex = idx
	} else {
		return MenaFrontmatter{}
	}

	var fm MenaFrontmatter
	if err := yaml.Unmarshal(data[searchStart:searchStart+endIndex], &fm); err != nil {
		log.Printf("Warning: malformed YAML frontmatter, treating as unscoped: %v", err)
		return MenaFrontmatter{}
	}
	return fm
}

// ProjectMena projects mena source files into commands/ and skills/ target
// directories. It handles extension stripping, mena type routing, and supports
// both filesystem and embedded FS sources.
//
// Sources are processed in priority order (later overrides earlier):
//  1. Distribution-level mena/ (from knossosHome or projectRoot)
//  2. rites/shared/mena/
//  3. rites/{dependency}/mena/ (in manifest dependency order)
//  4. rites/{active}/mena/ (highest priority)
//
// In Additive mode, existing files in target directories are preserved.
// In Destructive mode, target directories are wiped before projection.
func ProjectMena(sources []MenaSource, opts MenaProjectionOptions) (*MenaProjectionResult, error) {
	result := &MenaProjectionResult{}

	// Ensure target directories exist (both modes — selective, not destructive)
	if opts.Filter&ProjectDro != 0 {
		if err := os.MkdirAll(opts.TargetCommandsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create commands dir: %w", err)
		}
	}
	if opts.Filter&ProjectLego != 0 {
		if err := os.MkdirAll(opts.TargetSkillsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create skills dir: %w", err)
		}
	}

	// Pass 1: Collect mena entries from all sources.
	// Later sources override earlier ones for the same command name.
	collected := make(map[string]menaCollectedEntry)
	standalones := make(map[string]menaStandaloneFile)

	for _, src := range sources {
		if src.IsEmbedded {
			collectMenaEntriesFS(src.Fsys, src.FsysPath, "", collected)
		} else {
			if src.Path == "" {
				continue
			}
			if _, err := os.Stat(src.Path); os.IsNotExist(err) {
				continue
			}
			if err := collectMenaEntriesDir(src.Path, "", collected, standalones); err != nil {
				return nil, err
			}
		}
	}

	// Pass 2: Route each collected leaf directory by filename convention.
	for name, ce := range collected {
		menaType := "dro" // default: route to commands/

		if ce.source.IsEmbedded {
			entries, err := fs.ReadDir(ce.source.Fsys, ce.source.FsysPath)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
						menaType = DetectMenaType(entry.Name())
						break
					}
				}
			}
		} else {
			if entries, err := os.ReadDir(ce.source.Path); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
						menaType = DetectMenaType(entry.Name())
						break
					}
				}
			}
		}

		// Apply filter
		if menaType == "dro" && opts.Filter&ProjectDro == 0 {
			continue
		}
		if menaType == "lego" && opts.Filter&ProjectLego == 0 {
			continue
		}

		var destDir string
		if menaType == "dro" {
			destDir = filepath.Join(opts.TargetCommandsDir, name)
		} else {
			destDir = filepath.Join(opts.TargetSkillsDir, name)
		}

		// In destructive mode, clean this specific entry's subdir before writing.
		// This removes stale companion files from a previous version of the same entry.
		// User-created entries (not in collected set) are never touched.
		if opts.Mode == MenaProjectionDestructive {
			os.RemoveAll(destDir)
		}

		if ce.source.IsEmbedded {
			sub, err := fs.Sub(ce.source.Fsys, ce.source.FsysPath)
			if err != nil {
				return nil, err
			}
			if err := copyDirFromFSWithStripping(sub, destDir); err != nil {
				return nil, err
			}
		} else {
			if err := copyDirWithStripping(ce.source.Path, destDir); err != nil {
				return nil, err
			}
		}

		// Record what was projected
		targetType := "commands"
		if menaType == "lego" {
			targetType = "skills"
			result.SkillsProjected = append(result.SkillsProjected, name)
		} else {
			result.CommandsProjected = append(result.CommandsProjected, name)
		}

		// Record provenance at write time with exact source attribution
		if opts.Collector != nil {
			recordMenaProvenance(opts.Collector, opts.ProjectRoot, targetType, name, destDir, ce.source)
		}
	}

	// Copy standalone files (e.g., mena/navigation/rite.dro.md)
	// Route by extension: .dro.md -> commands/, .lego.md -> skills/
	for _, sf := range standalones {
		menaType := DetectMenaType(filepath.Base(sf.srcPath))

		// Apply filter
		if menaType == "dro" && opts.Filter&ProjectDro == 0 {
			continue
		}
		if menaType == "lego" && opts.Filter&ProjectLego == 0 {
			continue
		}

		var baseDir string
		if menaType == "dro" {
			baseDir = opts.TargetCommandsDir
		} else {
			baseDir = opts.TargetSkillsDir
		}

		// Strip the mena extension from the relative path's filename
		dir := filepath.Dir(sf.relPath)
		base := StripMenaExtension(filepath.Base(sf.relPath))
		strippedRel := filepath.Join(dir, base)

		destPath := filepath.Join(baseDir, strippedRel)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(sf.srcPath)
		if err != nil {
			return nil, err
		}
		if _, err := writeIfChanged(destPath, data, 0644); err != nil {
			return nil, err
		}

		targetType := "commands"
		if menaType == "lego" {
			targetType = "skills"
			result.SkillsProjected = append(result.SkillsProjected, strippedRel)
		} else {
			result.CommandsProjected = append(result.CommandsProjected, strippedRel)
		}

		// Record provenance for standalone file
		if opts.Collector != nil {
			now := time.Now().UTC()
			sourcePath := sf.srcPath
			if opts.ProjectRoot != "" {
				if rel, err := filepath.Rel(opts.ProjectRoot, sf.srcPath); err == nil {
					sourcePath = rel
				}
			}
			collector := opts.Collector
			collector.Record(targetType+"/"+strippedRel, &provenance.ProvenanceEntry{
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeRite,
				SourcePath: sourcePath,
				SourceType: "project",
				Checksum:   checksum.Content(string(data)),
				LastSynced: now,
			})
		}
	}

	// Note: stale entries from previous rites (no longer in projection set) persist.
	// Without provenance tracking, we cannot distinguish "stale knossos" from
	// "user-created" entries. This is an acceptable tradeoff — user content is
	// preserved at the cost of potential stale entries on rite switch.

	return result, nil
}

// recordMenaProvenance records a provenance entry for a projected mena directory.
// Uses directory checksum and exact source attribution.
func recordMenaProvenance(collector provenance.Collector, projectRoot, targetType, name, destDir string, src MenaSource) {
	now := time.Now().UTC()

	hash, err := checksum.Dir(destDir)
	if err != nil {
		return // best-effort: skip if checksum fails
	}

	sourcePath := ""
	sourceType := "project"

	if src.IsEmbedded {
		sourcePath = src.FsysPath
		if strings.Contains(src.FsysPath, "/shared/") {
			sourceType = "shared"
		}
	} else if src.Path != "" {
		sourceType = "project"
		// src.Path is already the full path to the leaf directory
		if projectRoot != "" {
			if rel, err := filepath.Rel(projectRoot, src.Path); err == nil {
				sourcePath = rel
			}
		}
		if sourcePath == "" {
			sourcePath = "mena/" + name + "/"
		}
	}

	collector.Record(targetType+"/"+name+"/", &provenance.ProvenanceEntry{
		Owner:      provenance.OwnerKnossos,
		Scope:      provenance.ScopeRite,
		SourcePath: sourcePath,
		SourceType: sourceType,
		Checksum:   hash,
		LastSynced: now,
	})
}

// copyDirWithStripping copies all files from src to dst, applying
// StripMenaExtension to filenames during copy.
func copyDirWithStripping(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Strip extension from the filename component
		dir := filepath.Dir(relPath)
		base := StripMenaExtension(filepath.Base(relPath))
		strippedRel := filepath.Join(dir, base)
		destPath := filepath.Join(dst, strippedRel)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		_, err = writeIfChanged(destPath, content, 0644)
		return err
	})
}

// copyDirFromFSWithStripping copies all files from an fs.FS to a destination
// directory on disk, applying StripMenaExtension to filenames during copy.
func copyDirFromFSWithStripping(fsys fs.FS, dst string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Strip extension from the filename component
		dir := filepath.Dir(path)
		base := StripMenaExtension(filepath.Base(path))
		strippedPath := filepath.Join(dir, base)
		destPath := filepath.Join(dst, strippedPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		_, err = writeIfChanged(destPath, content, 0644)
		return err
	})
}
