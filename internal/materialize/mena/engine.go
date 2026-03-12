package mena

import (
	"fmt"
	"io/fs"
	"log/slog"
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
//  1. Platform mena/ (core features: commands, skills, knowledge tools)
//  2. rites/shared/mena/ (cross-rite overlay: interview, smell-detection, etc.)
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
	// Propagate namespace collision warnings from resolution to result
	result.Warnings = append(result.Warnings, resolution.Warnings...)

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

		// Open a unified fs.FS view of this source (handles both embedded and
		// filesystem sources) so the copy and stale-file-collection paths share
		// one implementation.
		srcFS, srcRoot, err := openMenaFS(entry.Source)
		if err != nil {
			return nil, err
		}

		// Collect source filenames (with extension stripping) BEFORE writing,
		// so we can remove only stale files afterwards instead of nuking the whole dir.
		var sourceFileNames map[string]bool
		if opts.Mode == MenaProjectionDestructive {
			sourceFileNames = collectFSFileNames(srcFS, hideCompanions)
		}

		if err := copyDirFS(srcFS, srcRoot, destDir, hideCompanions, opts.Compiler); err != nil {
			return nil, err
		}

		// For dromena: INDEX.md was promoted to destDir.md at parent level.
		// Clean up old destDir/INDEX.md from previous syncs and remove
		// empty subdirectories left behind (dirs with only INDEX.md).
		if hideCompanions {
			oldIndex := filepath.Join(destDir, "INDEX.md")
			if _, statErr := os.Stat(oldIndex); statErr == nil {
				if rmErr := os.Remove(oldIndex); rmErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove old INDEX.md in %s: %v", destDir, rmErr))
				}
			}
			for _, cleanErr := range CleanEmptyDirs(destDir) {
				result.Warnings = append(result.Warnings, cleanErr.Error())
			}
			// Remove destDir itself if now empty (only had INDEX.md)
			if entries, readErr := os.ReadDir(destDir); readErr == nil && len(entries) == 0 {
				if rmErr := os.Remove(destDir); rmErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove empty directory %s: %v", destDir, rmErr))
				}
			}
		}

		// In destructive mode, remove only stale files that are no longer in source.
		// Guard: destDir may not exist for INDEX-only dromena (no companions).
		if opts.Mode == MenaProjectionDestructive && sourceFileNames != nil {
			if info, statErr := os.Stat(destDir); statErr == nil && info.IsDir() {
				removeStaleFiles(destDir, sourceFileNames)
			}
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
			recordMenaProvenance(opts.Collector, opts.ProjectRoot, targetType, entry.FlatName, destDir, opts.Channel, entry.Source)
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
		var data []byte
		if sf.isEmbedded && sf.fsys != nil {
			data, err = fs.ReadFile(sf.fsys, sf.SrcPath)
		} else {
			data, err = os.ReadFile(sf.SrcPath)
		}
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(destPath, ".md") {
			data = RewriteMenaContentPaths(data)
		}

		// Apply channel compiler transforms for standalone files
		if opts.Compiler != nil && strings.HasSuffix(destPath, ".md") {
			fm := ParseMenaFrontmatterBytes(data)
			name := fm.Name
			if name == "" {
				name = strings.TrimSuffix(sf.FlatName, ".md")
			}

			if sf.MenaType == "dro" {
				newFilename, newContent, err := opts.Compiler.CompileCommand(name, fm.Description, fm.ArgumentHint, string(data))
				if err != nil {
					return nil, err
				}
				destPath = filepath.Join(filepath.Dir(destPath), newFilename)
				data = newContent
			} else {
				_, newFilename, newContent, err := opts.Compiler.CompileSkill(name, fm.Description, string(data))
				if err != nil {
					return nil, err
				}
				destPath = filepath.Join(filepath.Dir(destPath), newFilename)
				data = newContent
			}
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
				checksum.Content(string(data)), opts.Channel,
			))
		}
	}

	// Pass 5: Clean stale knossos-owned mena entries that were renamed by flattening.
	if opts.Mode == MenaProjectionDestructive {
		cleanStaleMenaEntries(opts, result)
	}

	// Pass 6: Reconcile untracked entries — files on disk with no provenance.
	// These are artifacts from before provenance tracking that were never registered.
	// Only runs in destructive mode (rite-scope sync).
	if opts.Mode == MenaProjectionDestructive {
		reconcileUntrackedEntries(opts, result)
	}

	return result, nil
}

// recordMenaProvenance records a provenance entry for a projected mena directory.
func recordMenaProvenance(collector provenance.Collector, projectRoot, targetType, name, destDir, channel string, src MenaSource) {
	hash, err := checksum.Dir(destDir)
	if err != nil {
		// Directory may not exist if INDEX.md was promoted and there were no companions.
		// Try .md first, then .toml (gemini channel compiles to .toml).
		promotedFile := destDir + ".md"
		if data, readErr := os.ReadFile(promotedFile); readErr == nil {
			hash = checksum.Content(string(data))
		} else if data, readErr := os.ReadFile(destDir + ".toml"); readErr == nil {
			hash = checksum.Content(string(data))
		} else {
			return // best-effort: skip if all fail
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
		hash, channel,
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

	// Load existing provenance manifest to identify knossos-owned entries.
	// Manifest lives in .knossos/ (migrated from .claude/).
	// Uses channel-keyed path so gemini reads its own manifest, not claude's.
	claudeDir := filepath.Dir(opts.TargetCommandsDir)
	knossosDir := opts.KnossosDir
	if knossosDir == "" {
		// Fallback: derive from TargetCommandsDir (.claude/commands/ -> parent -> sibling .knossos/)
		knossosDir = filepath.Join(filepath.Dir(claudeDir), ".knossos")
	}
	manifestPath := provenance.ManifestPathForChannel(knossosDir, opts.Channel)
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		return	// No manifest = no stale entries to clean
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

		// No rite scoping needed — the projected set from SyncMena contains all
		// entries from all active sources (platform + shared + deps + current rite).
		// Any knossos-owned entry not in the projected set is genuinely stale,
		// regardless of which rite originally projected it.

		// Stale knossos-owned entry -- remove it
		absPath := filepath.Join(claudeDir, key)
		absPath = strings.TrimRight(absPath, "/")
		if info, err := os.Stat(absPath); err == nil {
			if info.IsDir() {
				if rmErr := os.RemoveAll(absPath); rmErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove stale mena directory %s: %v", key, rmErr))
				} else {
					slog.Info("removed stale mena entry", "key", key)
				}
			} else {
				if rmErr := os.Remove(absPath); rmErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove stale mena file %s: %v", key, rmErr))
				} else {
					slog.Info("removed stale mena entry", "key", key)
				}
			}
		}
		// Try promoted file variants: .md (claude) and .toml (gemini)
		for _, ext := range []string{".md", ".toml"} {
			promotedFile := absPath + ext
			if _, statErr := os.Stat(promotedFile); statErr == nil {
				if rmErr := os.Remove(promotedFile); rmErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove stale promoted file %s%s: %v", key, ext, rmErr))
				} else {
					slog.Info("removed stale promoted file", "key", key+ext)
				}
			}
		}
	}

	// Also clean empty parent directories left behind by removal.
	// Surface non-permission errors as warnings -- permission errors on shared
	// or read-only directories are acceptable and silently ignored.
	for _, dir := range []string{opts.TargetCommandsDir, opts.TargetSkillsDir} {
		for _, cleanErr := range CleanEmptyDirs(dir) {
			result.Warnings = append(result.Warnings, cleanErr.Error())
		}
	}
}

// isFromRite checks whether a provenance source_path belongs to a specific rite.
// It matches the pattern "rites/{riteName}/mena/" anywhere in the path, handling
// both relative (rites/10x-dev/mena/) and absolute paths.
func isFromRite(sourcePath, riteName string) bool {
	return strings.Contains(sourcePath, "rites/"+riteName+"/mena/")
}

// isFromActiveChain checks whether a source path belongs to the active rite,
// any of its dependencies, or a non-rite source (platform/procession mena).
// Entries from rites outside the chain are considered stale on rite switch.
func isFromActiveChain(sourcePath, riteName string, deps []string) bool {
	if isFromRite(sourcePath, riteName) {
		return true
	}
	for _, dep := range deps {
		if isFromRite(sourcePath, dep) {
			return true
		}
	}
	// Non-rite sources (platform mena, procession mena) are always active
	return !strings.Contains(sourcePath, "rites/")
}

// reconcileUntrackedEntries removes files in commands/ and skills/ that are
// not in the current projection AND not tracked in provenance. These are
// artifacts from before provenance tracking was introduced, or from rites
// whose provenance was lost. Only removes files that have knossos-style
// mena frontmatter (name: + description: fields), preserving user-created entries.
func reconcileUntrackedEntries(opts MenaProjectionOptions, result *MenaProjectionResult) {
	// Build projected set
	projected := make(map[string]bool)
	for _, name := range result.CommandsProjected {
		projected[name] = true
	}
	for _, name := range result.SkillsProjected {
		projected[name] = true
	}

	// Load provenance manifest to check tracking status
	claudeDir := filepath.Dir(opts.TargetCommandsDir)
	knossosDir := opts.KnossosDir
	if knossosDir == "" {
		knossosDir = filepath.Join(filepath.Dir(claudeDir), ".knossos")
	}
	manifestPath := provenance.ManifestPathForChannel(knossosDir, opts.Channel)
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		return // No manifest = can't determine tracking status
	}

	for _, pair := range []struct{ dir, prefix string }{
		{opts.TargetCommandsDir, "commands/"},
		{opts.TargetSkillsDir, "skills/"},
	} {
		entries, readErr := os.ReadDir(pair.dir)
		if readErr != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			// Strip extensions to get the command/skill name
			baseName := strings.TrimSuffix(strings.TrimSuffix(name, ".md"), ".toml")

			// Skip if in current projection
			if projected[baseName] || projected[baseName+"/"] || projected[baseName+".md"] || projected[name] {
				continue
			}

			// Skip .gitkeep and hidden files
			if strings.HasPrefix(name, ".") {
				continue
			}

			// Check if tracked in provenance (try multiple key formats)
			tracked := false
			for _, keyVariant := range []string{
				pair.prefix + name,
				pair.prefix + name + "/",
				pair.prefix + baseName + ".md",
				pair.prefix + baseName + "/",
			} {
				if _, ok := manifest.Entries[keyVariant]; ok {
					tracked = true
					break
				}
			}
			if tracked {
				continue // Already handled by cleanStaleMenaEntries
			}

			// Untracked entry — check if it has knossos mena frontmatter
			absPath := filepath.Join(pair.dir, name)
			if e.IsDir() {
				// Directory: check for INDEX file with mena frontmatter
				if hasMenaFrontmatter(absPath) {
					if rmErr := os.RemoveAll(absPath); rmErr != nil {
						result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove untracked mena directory %s: %v", pair.prefix+name, rmErr))
					} else {
						slog.Info("removed untracked mena entry", "key", pair.prefix+name)
					}
				}
			} else {
				// File: check for mena frontmatter
				if hasMenaFrontmatterFile(absPath) {
					if rmErr := os.Remove(absPath); rmErr != nil {
						result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove untracked mena file %s: %v", pair.prefix+name, rmErr))
					} else {
						slog.Info("removed untracked mena entry", "key", pair.prefix+name)
					}
				}
			}
		}
	}
}

// hasMenaFrontmatter checks if a directory contains an INDEX file with
// knossos-style mena frontmatter (name: and description: fields).
func hasMenaFrontmatter(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "INDEX") {
			return true // INDEX file presence is sufficient
		}
	}
	// Check for promoted file (standalone command)
	return false
}

// hasMenaFrontmatterFile checks if a standalone file has knossos-style
// mena frontmatter by looking for name: and description: in YAML frontmatter.
func hasMenaFrontmatterFile(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	content := string(data)
	// Quick heuristic: mena-generated files have YAML frontmatter with name: and description:
	if !strings.HasPrefix(content, "---") {
		return false
	}
	// Check for both required frontmatter fields
	return strings.Contains(content, "\nname:") && strings.Contains(content, "\ndescription:")
}

// CleanEmptyDirs removes empty subdirectories within a directory.
// Returns non-permission errors encountered during cleanup (permission errors
// are acceptable on shared/read-only directories and are silently ignored).
// Callers should surface these errors as warnings, not abort the pipeline.
func CleanEmptyDirs(root string) []error {
	if _, err := os.Stat(root); err != nil {
		return nil	// Directory doesn't exist, nothing to clean
	}
	return cleanEmptyDirsRecursive(root)
}

// cleanEmptyDirsRecursive is the internal recursive implementation of CleanEmptyDirs.
func cleanEmptyDirsRecursive(root string) []error {
	var errs []error
	entries, err := os.ReadDir(root)
	if err != nil {
		if !os.IsPermission(err) {
			errs = append(errs, fmt.Errorf("CleanEmptyDirs: failed to read directory %s: %w", root, err))
		}
		return errs
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subdir := filepath.Join(root, entry.Name())
		subEntries, err := os.ReadDir(subdir)
		if err != nil {
			if !os.IsPermission(err) {
				errs = append(errs, fmt.Errorf("CleanEmptyDirs: failed to read subdirectory %s: %w", subdir, err))
			}
			continue
		}
		if len(subEntries) == 0 {
			if rmErr := os.Remove(subdir); rmErr != nil && !os.IsPermission(rmErr) {
				errs = append(errs, fmt.Errorf("CleanEmptyDirs: failed to remove empty directory %s: %w", subdir, rmErr))
			}
		} else {
			errs = append(errs, cleanEmptyDirsRecursive(subdir)...)
			// Re-read after recursive cleanup to check if now empty
			subEntries, _ = os.ReadDir(subdir)
			if len(subEntries) == 0 {
				if rmErr := os.Remove(subdir); rmErr != nil && !os.IsPermission(rmErr) {
					errs = append(errs, fmt.Errorf("CleanEmptyDirs: failed to remove empty directory %s: %w", subdir, rmErr))
				}
			}
		}
	}
	return errs
}

// removeStaleFiles removes files in destDir that are NOT in the sourceFileNames set.
// Removal errors are logged as warnings (non-fatal).
func removeStaleFiles(destDir string, sourceFileNames map[string]bool) {
	_ = filepath.WalkDir(destDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		relPath, relErr := filepath.Rel(destDir, path)
		if relErr != nil {
			return nil
		}
		if !sourceFileNames[relPath] {
			if rmErr := os.Remove(path); rmErr != nil {
				slog.Warn("failed to remove stale file", "path", path, "error", rmErr)
			}
		}
		return nil
	})
	// Log non-permission errors from CleanEmptyDirs. Permission errors are
	// acceptable on shared/read-only directories and are silently ignored.
	for _, cleanErr := range CleanEmptyDirs(destDir) {
		slog.Warn("clean empty dirs issue", "error", cleanErr)
	}
}
