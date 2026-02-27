package mena

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Exists checks whether a mena entry with the given name and type can be
// resolved from any of the provided sources.
//
// For legomena ("lego"): checks for INDEX directory pattern.
// For dromena ("dro"): checks for INDEX directory pattern and flat file pattern.
//
// Filesystem sources are checked via os.Stat. Embedded FS sources are
// checked via fs.Stat/fs.ReadDir. Sources with empty/nonexistent paths
// are silently skipped.
func Exists(name string, menaType string, sources []MenaSource) bool {
	for _, src := range sources {
		if src.IsEmbedded {
			if fsExistsInSource(name, menaType, src) {
				return true
			}
		} else {
			if fsPathExistsInSource(name, menaType, src) {
				return true
			}
		}
	}
	return false
}

// fsPathExistsInSource checks a filesystem (non-embedded) source.
func fsPathExistsInSource(name string, menaType string, src MenaSource) bool {
	if src.Path == "" {
		return false
	}

	// INDEX dir pattern: check for any file with strings.HasPrefix(name, "INDEX")
	// in the {src.Path}/{name}/ directory.
	dirPath := filepath.Join(src.Path, name)
	if dirHasIndex(dirPath) {
		return true
	}

	// Flat file pattern: dromena only
	if menaType == "dro" {
		if _, err := os.Stat(filepath.Join(src.Path, name+".dro.md")); err == nil {
			return true
		}
	}

	return false
}

// fsExistsInSource checks an embedded FS source.
func fsExistsInSource(name string, menaType string, src MenaSource) bool {
	// INDEX dir pattern: check for any file with strings.HasPrefix(name, "INDEX")
	// in the {src.FsysPath}/{name}/ directory.
	dirPath := src.FsysPath + "/" + name
	if fsHasIndex(src.Fsys, dirPath) {
		return true
	}

	// Flat file pattern: dromena only
	if menaType == "dro" {
		if _, err := fs.Stat(src.Fsys, src.FsysPath+"/"+name+".dro.md"); err == nil {
			return true
		}
	}

	return false
}

// dirHasIndex checks if a filesystem directory contains an INDEX file.
// Uses strings.HasPrefix to match any INDEX* file, aligning with the
// materializer's dirHasIndexFile() behavior.
func dirHasIndex(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			return true
		}
	}
	return false
}

// fsHasIndex checks if an embedded FS directory contains an INDEX file.
func fsHasIndex(fsys fs.FS, dirPath string) bool {
	entries, err := fs.ReadDir(fsys, dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			return true
		}
	}
	return false
}
