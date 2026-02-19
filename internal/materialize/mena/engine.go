package mena

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/provenance"
)

// SyncMena projects mena source files into commands/ and skills/ target
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
func SyncMena(sources []MenaSource, opts MenaProjectionOptions) (*MenaProjectionResult, error) {
	result := &MenaProjectionResult{}

	// Ensure target directories exist (both modes -- selective, not destructive)
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

	// Collect and resolve mena entries (shared with user-scope pipeline)
	resolution, err := CollectMena(sources, opts)
	if err != nil {
		return nil, err
	}

	// Pass 3: Write directory entries to target directories.
	for _, entry := range resolution.Entries {
		var destDir string
		if entry.MenaType == "dro" {
			destDir = filepath.Join(opts.TargetCommandsDir, entry.FlatName)
		} else {
			destDir = filepath.Join(opts.TargetSkillsDir, entry.FlatName)
		}

		// Hide companions for dromena only
		hideCompanions := entry.MenaType == "dro"

		// Collect source filenames (with extension stripping) BEFORE writing,
		// so we can remove only stale files afterwards instead of nuking the whole dir.
		var sourceFileNames map[string]bool
		if opts.Mode == MenaProjectionDestructive {
			sourceFileNames = collectSourceFileNames(entry.Source, hideCompanions)
		}

		if entry.Source.IsEmbedded {
			sub, err := fs.Sub(entry.Source.Fsys, entry.Source.FsysPath)
			if err != nil {
				return nil, err
			}
			if err := copyDirFromFSWithStripping(sub, destDir, hideCompanions); err != nil {
				return nil, err
			}
		} else {
			if err := copyDirWithStripping(entry.Source.Path, destDir, hideCompanions); err != nil {
				return nil, err
			}
		}

		// For dromena: INDEX.md was promoted to destDir.md at parent level.
		// Clean up old destDir/INDEX.md from previous syncs and remove
		// empty subdirectories left behind (dirs with only INDEX.md).
		if hideCompanions {
			oldIndex := filepath.Join(destDir, "INDEX.md")
			if _, statErr := os.Stat(oldIndex); statErr == nil {
				os.Remove(oldIndex)
			}
			CleanEmptyDirs(destDir)
			// Remove destDir itself if now empty (only had INDEX.md)
			if entries, readErr := os.ReadDir(destDir); readErr == nil && len(entries) == 0 {
				os.Remove(destDir)
			}
		}

		// In destructive mode, remove only stale files that are no longer in source.
		if opts.Mode == MenaProjectionDestructive && sourceFileNames != nil {
			removeStaleFiles(destDir, sourceFileNames)
		}

		// Record what was projected
		targetType := "commands"
		if entry.MenaType == "lego" {
			targetType = "skills"
			result.SkillsProjected = append(result.SkillsProjected, entry.FlatName)
		} else {
			result.CommandsProjected = append(result.CommandsProjected, entry.FlatName)
		}

		// Record provenance at write time with exact source attribution
		if opts.Collector != nil {
			recordMenaProvenance(opts.Collector, opts.ProjectRoot, targetType, entry.FlatName, destDir, entry.Source)
		}
	}

	// Pass 4: Write standalone files (e.g., mena/navigation/rite.dro.md)
	for _, sf := range resolution.Standalones {
		var baseDir string
		if sf.MenaType == "dro" {
			baseDir = opts.TargetCommandsDir
		} else {
			baseDir = opts.TargetSkillsDir
		}

		destPath := filepath.Join(baseDir, sf.FlatName)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(sf.SrcPath)
		if err != nil {
			return nil, err
		}
		if _, err := fileutil.WriteIfChanged(destPath, data, 0644); err != nil {
			return nil, err
		}

		targetType := "commands"
		if sf.MenaType == "lego" {
			targetType = "skills"
			result.SkillsProjected = append(result.SkillsProjected, sf.FlatName)
		} else {
			result.CommandsProjected = append(result.CommandsProjected, sf.FlatName)
		}

		// Record provenance for standalone file
		if opts.Collector != nil {
			sourcePath := sf.SrcPath
			if opts.ProjectRoot != "" {
				if rel, err := filepath.Rel(opts.ProjectRoot, sf.SrcPath); err == nil {
					sourcePath = rel
				}
			}
			opts.Collector.Record(targetType+"/"+sf.FlatName, provenance.NewKnossosEntry(
				provenance.ScopeRite,
				sourcePath,
				"project",
				checksum.Content(string(data)),
			))
		}
	}

	// Pass 5: Clean stale knossos-owned mena entries that were renamed by flattening.
	if opts.Mode == MenaProjectionDestructive {
		cleanStaleMenaEntries(opts, result)
	}

	return result, nil
}

// recordMenaProvenance records a provenance entry for a projected mena directory.
func recordMenaProvenance(collector provenance.Collector, projectRoot, targetType, name, destDir string, src MenaSource) {
	hash, err := checksum.Dir(destDir)
	if err != nil {
		// Directory may not exist if INDEX.md was promoted and there were no companions.
		promotedFile := destDir + ".md"
		if data, readErr := os.ReadFile(promotedFile); readErr == nil {
			hash = checksum.Content(string(data))
		} else {
			return // best-effort: skip if both fail
		}
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
		if projectRoot != "" {
			if rel, err := filepath.Rel(projectRoot, src.Path); err == nil {
				sourcePath = rel
			}
		}
		if sourcePath == "" {
			sourcePath = "mena/" + name + "/"
		}
	}

	collector.Record(targetType+"/"+name+"/", provenance.NewKnossosEntry(
		provenance.ScopeRite,
		sourcePath,
		sourceType,
		hash,
	))
}

// cleanStaleMenaEntries removes knossos-owned command/skill directories that are
// no longer in the current projection result.
func cleanStaleMenaEntries(opts MenaProjectionOptions, result *MenaProjectionResult) {
	// Build set of currently projected entries
	projected := make(map[string]bool)
	for _, name := range result.CommandsProjected {
		projected["commands/"+name+"/"] = true
		projected["commands/"+name] = true
	}
	for _, name := range result.SkillsProjected {
		projected["skills/"+name+"/"] = true
		projected["skills/"+name] = true
	}

	// Load existing provenance manifest to identify knossos-owned entries
	claudeDir := filepath.Dir(opts.TargetCommandsDir)
	manifestPath := filepath.Join(claudeDir, provenance.ManifestFileName)
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		return // No manifest = no stale entries to clean
	}

	// Find knossos-owned mena entries not in current projection
	for key, entry := range manifest.Entries {
		if entry.Owner != provenance.OwnerKnossos {
			continue
		}
		if !strings.HasPrefix(key, "commands/") && !strings.HasPrefix(key, "skills/") {
			continue
		}
		if projected[key] {
			continue
		}

		// Stale knossos-owned entry -- remove it
		absPath := filepath.Join(claudeDir, key)
		absPath = strings.TrimRight(absPath, "/")
		if info, err := os.Stat(absPath); err == nil {
			if info.IsDir() {
				os.RemoveAll(absPath)
			} else {
				os.Remove(absPath)
			}
			log.Printf("Removed stale mena entry: %s", key)
		}
		promotedFile := absPath + ".md"
		if _, statErr := os.Stat(promotedFile); statErr == nil {
			os.Remove(promotedFile)
			log.Printf("Removed stale promoted file: %s.md", key)
		}
	}

	// Also clean empty parent directories left behind by removal
	for _, dir := range []string{opts.TargetCommandsDir, opts.TargetSkillsDir} {
		CleanEmptyDirs(dir)
	}
}

// CleanEmptyDirs removes empty subdirectories within a directory.
func CleanEmptyDirs(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subdir := filepath.Join(root, entry.Name())
		subEntries, err := os.ReadDir(subdir)
		if err != nil {
			continue
		}
		if len(subEntries) == 0 {
			os.Remove(subdir)
		} else {
			CleanEmptyDirs(subdir)
			subEntries, _ = os.ReadDir(subdir)
			if len(subEntries) == 0 {
				os.Remove(subdir)
			}
		}
	}
}

// copyDirWithStripping copies all files from src to dst, applying
// StripMenaExtension to filenames during copy.
func copyDirWithStripping(src, dst string, hideCompanions bool) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

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

		// For dromena: promote top-level INDEX.md to parent level (dst.md)
		if hideCompanions && base == "INDEX.md" && dir == "." {
			destPath = dst + ".md"
		}

		// For legomena: rename top-level INDEX.md → SKILL.md (CC entrypoint convention).
		// CC reads SKILL.md as the skill entrypoint; INDEX.md is not recognized.
		if !hideCompanions && base == "INDEX.md" && dir == "." {
			base = "SKILL.md"
			destPath = filepath.Join(dst, "SKILL.md")
		}

		// Apply companion hiding for dromena non-INDEX markdown files
		if hideCompanions && base != "INDEX.md" && strings.HasSuffix(base, ".md") {
			content = InjectCompanionHideFrontmatter(content)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		_, err = fileutil.WriteIfChanged(destPath, content, 0644)
		return err
	})
}

// copyDirFromFSWithStripping copies all files from an fs.FS to a destination
// directory on disk, applying StripMenaExtension to filenames during copy.
func copyDirFromFSWithStripping(fsys fs.FS, dst string, hideCompanions bool) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

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

		// For dromena: promote top-level INDEX.md to parent level (dst.md)
		if hideCompanions && base == "INDEX.md" && dir == "." {
			destPath = dst + ".md"
		}

		// For legomena: rename top-level INDEX.md → SKILL.md (CC entrypoint convention).
		// CC reads SKILL.md as the skill entrypoint; INDEX.md is not recognized.
		if !hideCompanions && base == "INDEX.md" && dir == "." {
			base = "SKILL.md"
			destPath = filepath.Join(dst, "SKILL.md")
		}

		// Apply companion hiding for dromena non-INDEX markdown files
		if hideCompanions && base != "INDEX.md" && strings.HasSuffix(base, ".md") {
			content = InjectCompanionHideFrontmatter(content)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		_, err = fileutil.WriteIfChanged(destPath, content, 0644)
		return err
	})
}

// collectSourceFileNames builds the set of destination-relative file paths
// that a mena source will produce (after extension stripping and promotion).
// hideCompanions must match the value passed to copyDirWithStripping/copyDirFromFSWithStripping
// so that the legomena INDEX.md → SKILL.md rename is reflected in the expected filenames.
func collectSourceFileNames(src MenaSource, hideCompanions bool) map[string]bool {
	names := make(map[string]bool)

	if src.IsEmbedded {
		sub, err := fs.Sub(src.Fsys, src.FsysPath)
		if err != nil {
			return names
		}
		fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return walkErr
			}
			dir := filepath.Dir(path)
			base := StripMenaExtension(filepath.Base(path))
			// Mirror legomena promotion: INDEX.md → SKILL.md at root level
			if !hideCompanions && base == "INDEX.md" && dir == "." {
				base = "SKILL.md"
			}
			names[filepath.Join(dir, base)] = true
			return nil
		})
	} else if src.Path != "" {
		filepath.WalkDir(src.Path, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return walkErr
			}
			relPath, relErr := filepath.Rel(src.Path, path)
			if relErr != nil {
				return nil
			}
			dir := filepath.Dir(relPath)
			base := StripMenaExtension(filepath.Base(relPath))
			// Mirror legomena promotion: INDEX.md → SKILL.md at root level
			if !hideCompanions && base == "INDEX.md" && dir == "." {
				base = "SKILL.md"
			}
			names[filepath.Join(dir, base)] = true
			return nil
		})
	}

	return names
}

// removeStaleFiles removes files in destDir that are NOT in the sourceFileNames set.
func removeStaleFiles(destDir string, sourceFileNames map[string]bool) {
	filepath.WalkDir(destDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		relPath, relErr := filepath.Rel(destDir, path)
		if relErr != nil {
			return nil
		}
		if !sourceFileNames[relPath] {
			os.Remove(path)
		}
		return nil
	})
	CleanEmptyDirs(destDir)
}
