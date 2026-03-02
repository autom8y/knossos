package mena

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/fileutil"
)

// openMenaFS returns an fs.FS for a MenaSource and the root path to walk.
//
// For embedded sources the caller must already have called fs.Sub before passing
// the FS here; fsysRoot is always "." in that case.
//
// For filesystem sources we use os.DirFS so that both code paths share the same
// fs.WalkDir implementation. Note: os.DirFS does NOT follow symlinks during
// traversal. Mena directories are structured templates without symlinks, so this
// is acceptable. If symlink support is ever required, switch back to
// filepath.WalkDir with a custom os.DirFS-compatible adapter.
func openMenaFS(src MenaSource) (fsys fs.FS, root string, err error) {
	if src.IsEmbedded {
		sub, subErr := fs.Sub(src.Fsys, src.FsysPath)
		if subErr != nil {
			return nil, "", subErr
		}
		return sub, ".", nil
	}
	return os.DirFS(src.Path), ".", nil
}

// copyDirFS copies all files from fsys (rooted at root) to dst on disk, applying
// StripMenaExtension to filenames. hideCompanions controls dromena-specific
// INDEX.md promotion and companion-hide frontmatter injection.
//
// This is the unified replacement for the two previously separate functions:
//   - copyDirWithStripping (filesystem)
//   - copyDirFromFSWithStripping (embed.FS)
func copyDirFS(fsys fs.FS, root, dst string, hideCompanions bool) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute destination-relative path after extension stripping.
		// When root=="." fs.WalkDir passes "." as the first path — skip it.
		if path == "." {
			return nil
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

		// For dromena: promote top-level INDEX.md to parent level (dst.md).
		if hideCompanions && base == "INDEX.md" && dir == "." {
			destPath = dst + ".md"
		}

		// For legomena: rename top-level INDEX.md → SKILL.md (CC entrypoint convention).
		// CC reads SKILL.md as the skill entrypoint; INDEX.md is not recognised.
		if !hideCompanions && base == "INDEX.md" && dir == "." {
			base = "SKILL.md"
			destPath = filepath.Join(dst, "SKILL.md")
		}

		// Apply companion hiding for dromena non-INDEX markdown files.
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

// collectFSFileNames builds the set of destination-relative file paths that a
// MenaSource will produce after extension stripping and legomena INDEX→SKILL
// promotion.  hideCompanions must match the value used in copyDirFS.
//
// This is the unified replacement for the two previously separate branches in
// collectSourceFileNames.
func collectFSFileNames(fsys fs.FS, hideCompanions bool) map[string]bool {
	names := make(map[string]bool)
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error { //nolint:errcheck
		if walkErr != nil || d.IsDir() || path == "." {
			return walkErr
		}
		dir := filepath.Dir(path)
		base := StripMenaExtension(filepath.Base(path))
		// Mirror legomena promotion: INDEX.md → SKILL.md at root level.
		if !hideCompanions && base == "INDEX.md" && dir == "." {
			base = "SKILL.md"
		}
		names[filepath.Join(dir, base)] = true
		return nil
	})
	return names
}
