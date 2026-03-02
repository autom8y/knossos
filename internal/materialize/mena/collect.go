package mena

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CollectMena collects and resolves mena entries from sources without writing files.
// Returns the resolved entries with flat names and mena types.
// Reused by both rite-scope (SyncMena) and user-scope (syncUserMena).
func CollectMena(sources []MenaSource, opts MenaProjectionOptions) (*MenaResolution, error) {
	resolution := &MenaResolution{
		Entries:     make(map[string]MenaResolvedEntry),
		Standalones: make(map[string]MenaResolvedStandalone),
	}

	// Pass 1: Collect mena entries from all sources.
	// Later sources override earlier ones for the same command name.
	collected := make(map[string]menaCollectedEntry)
	standalones := make(map[string]menaStandaloneFile)

	for srcIdx, src := range sources {
		if src.IsEmbedded {
			collectMenaEntriesFS(src.Fsys, src.FsysPath, "", collected, srcIdx)
		} else {
			if src.Path == "" {
				continue
			}
			if _, err := os.Stat(src.Path); os.IsNotExist(err) {
				// Log at verbose level so callers can diagnose missing mena sources
				// (e.g., shared mena pointing to wrong base directory for satellite rites).
				// Graceful skip is intentional — not all sources are required to exist.
				log.Printf("mena: source path does not exist, skipping: %s", src.Path)
				continue
			}
			if err := collectMenaEntriesDir(src.Path, "", collected, standalones, srcIdx); err != nil {
				return nil, err
			}
		}
	}

	// Pass 1.5: Resolve flat namespace for dromena.
	// resolveNamespace returns warnings for user-owned collisions so callers can
	// surface them as diagnostic output rather than silently falling back.
	flatNames, nsWarnings := resolveNamespace(collected, standalones, opts)
	resolution.Warnings = append(resolution.Warnings, nsWarnings...)

	// Pass 2: Apply flat names for each directory entry using the cached mena type.
	for name, ce := range collected {
		menaType := ce.menaType

		// Apply filter
		if menaType == "dro" && opts.Filter&ProjectDro == 0 {
			continue
		}
		if menaType == "lego" && opts.Filter&ProjectLego == 0 {
			continue
		}

		// Use flat name for dromena if available
		flatName := name
		if menaType == "dro" {
			if fn, ok := flatNames[name]; ok {
				flatName = fn
			}
		}

		resolution.Entries[name] = MenaResolvedEntry{
			Source:   ce.source,
			FlatName: flatName,
			MenaType: menaType,
		}
	}

	// Resolve standalones.
	for key, sf := range standalones {
		menaType := DetectMenaType(filepath.Base(sf.srcPath))

		// Apply filter
		if menaType == "dro" && opts.Filter&ProjectDro == 0 {
			continue
		}
		if menaType == "lego" && opts.Filter&ProjectLego == 0 {
			continue
		}

		// Strip the mena extension from the relative path's filename
		dir := filepath.Dir(sf.relPath)
		base := StripMenaExtension(filepath.Base(sf.relPath))

		// Use flat name for dromena if available
		var strippedRel string
		if menaType == "dro" {
			if flatName, ok := flatNames[sf.relPath]; ok {
				strippedRel = flatName + ".md"
			} else {
				strippedRel = filepath.Join(dir, base)
			}
		} else {
			strippedRel = filepath.Join(dir, base)
		}

		resolution.Standalones[key] = MenaResolvedStandalone{
			SrcPath:  sf.srcPath,
			RelPath:  sf.relPath,
			FlatName: strippedRel,
			MenaType: menaType,
		}
	}

	return resolution, nil
}

// collectMenaEntriesDir recursively collects mena entries from a filesystem directory.
// Leaf directories (containing INDEX files) are collected for routing.
// Standalone files in grouping directories are collected separately.
func collectMenaEntriesDir(dirPath string, prefix string, collected map[string]menaCollectedEntry, standalones map[string]menaStandaloneFile, srcIdx int) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if prefix != "" {
			name = prefix + "/" + entry.Name()
		}
		if entry.IsDir() {
			childPath := filepath.Join(dirPath, entry.Name())
			if dirHasIndexFile(childPath) {
				ce := menaCollectedEntry{
					source:      MenaSource{Path: childPath},
					name:        name,
					sourceIndex: srcIdx,
				}
				ce.menaType = detectEntryMenaType(ce)
				collected[name] = ce
			} else {
				if err := collectMenaEntriesDir(childPath, name, collected, standalones, srcIdx); err != nil {
					return err
				}
			}
		} else {
			// Standalone file in a grouping directory
			standalones[name] = menaStandaloneFile{
				srcPath:     filepath.Join(dirPath, entry.Name()),
				relPath:     name,
				sourceIndex: srcIdx,
			}
		}
	}
	return nil
}

// collectMenaEntriesFS recursively collects mena entries from an embedded filesystem.
func collectMenaEntriesFS(fsys fs.FS, fsysPath string, prefix string, collected map[string]menaCollectedEntry, srcIdx int) {
	entries, err := fs.ReadDir(fsys, fsysPath)
	if err != nil {
		return // Path doesn't exist in embedded FS
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Embedded rite mena doesn't have standalone files
		}
		name := entry.Name()
		if prefix != "" {
			name = prefix + "/" + entry.Name()
		}
		childPath := fsysPath + "/" + entry.Name()
		if fsHasIndexFile(fsys, childPath) {
			ce := menaCollectedEntry{
				source: MenaSource{
					Fsys:       fsys,
					FsysPath:   childPath,
					IsEmbedded: true,
				},
				name:        name,
				sourceIndex: srcIdx,
			}
			ce.menaType = detectEntryMenaType(ce)
			collected[name] = ce
		} else {
			collectMenaEntriesFS(fsys, childPath, name, collected, srcIdx)
		}
	}
}

// dirHasIndexFile checks if a filesystem directory contains an INDEX file.
func dirHasIndexFile(dirPath string) bool {
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

// fsHasIndexFile checks if an embedded FS directory contains an INDEX file.
func fsHasIndexFile(fsys fs.FS, dirPath string) bool {
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

// detectEntryMenaType determines the mena type for a collected directory entry
// by examining its INDEX file. Returns "dro" as default.
func detectEntryMenaType(ce menaCollectedEntry) string {
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

	return menaType
}
