package mena

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/provenance"
)

// resolveNamespace computes flat command names for dromena entries by reading
// frontmatter name fields. Returns a map from source key to flat name and a
// slice of diagnostic warnings for namespace collisions.
// On name collision between dromena from different sources, the highest-priority
// source wins (later in sources array = higher priority: user < shared < dep < rite).
// On collision with user-owned commands in target dir, knossos entry falls back
// and a warning is emitted so the user knows why the flat name was not applied.
func resolveNamespace(collected map[string]menaCollectedEntry, standalones map[string]menaStandaloneFile, opts MenaProjectionOptions) (map[string]string, []string) {
	flatNames := make(map[string]string)
	var warnings []string

	type nameCandidate struct {
		sourceKey   string
		sourceIndex int
	}
	nameToSources := make(map[string][]nameCandidate) // flat name -> candidates

	// Step 1: Read frontmatter names from collected entries (directories with INDEX files)
	for sourceKey, ce := range collected {
		// Only process dromena (commands/)
		var menaType string
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

		if menaType != "dro" {
			continue // Only flatten dromena
		}

		// Read frontmatter name
		var fm MenaFrontmatter
		if ce.source.IsEmbedded {
			// Read INDEX file from embedded FS
			entries, err := fs.ReadDir(ce.source.Fsys, ce.source.FsysPath)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
						indexPath := ce.source.FsysPath + "/" + entry.Name()
						data, err := fs.ReadFile(ce.source.Fsys, indexPath)
						if err == nil {
							fm = ParseMenaFrontmatterBytes(data)
						}
						break
					}
				}
			}
		} else {
			fm = readMenaFrontmatterFromDir(ce.source.Path)
		}

		if fm.Name != "" {
			nameToSources[fm.Name] = append(nameToSources[fm.Name], nameCandidate{
				sourceKey:   sourceKey,
				sourceIndex: ce.sourceIndex,
			})
		}
	}

	// Step 2: Read frontmatter names from standalone files
	for sourceKey, sf := range standalones {
		menaType := DetectMenaType(filepath.Base(sf.srcPath))
		if menaType != "dro" {
			continue // Only flatten dromena
		}

		fm := readMenaFrontmatterFromFile(sf.srcPath)
		if fm.Name != "" {
			nameToSources[fm.Name] = append(nameToSources[fm.Name], nameCandidate{
				sourceKey:   sourceKey,
				sourceIndex: sf.sourceIndex,
			})
		}
	}

	// Step 3: Build flat name mapping, resolve collisions by source priority
	for flatName, candidates := range nameToSources {
		if len(candidates) > 1 {
			// Multiple sources want same flat name -- highest sourceIndex wins
			winner := candidates[0]
			for _, c := range candidates[1:] {
				if c.sourceIndex > winner.sourceIndex {
					winner = c
				}
			}
			flatNames[winner.sourceKey] = flatName
			// Losers keep their source paths (no flat name assigned)
			continue
		}
		flatNames[candidates[0].sourceKey] = flatName
	}

	// Step 4: Pre-scan target commands/ for user-created entries.
	// If a flat name collides with an existing user-owned entry, knossos yields.
	// Uses provenance manifest to distinguish knossos-owned (safe to overwrite) from user-owned.
	if opts.TargetCommandsDir != "" {
		// Load existing provenance manifest to identify ownership.
		// Manifest lives in .knossos/ (migrated from .claude/).
		knossosDir := opts.KnossosDir
		if knossosDir == "" {
			// Fallback: derive from TargetCommandsDir (.claude/commands/ -> parent -> sibling .knossos/)
			channelDir := filepath.Dir(opts.TargetCommandsDir)
			knossosDir = filepath.Join(filepath.Dir(channelDir), ".knossos")
		}
		manifestPath := provenance.ManifestPathForChannel(knossosDir, opts.Channel)
		oldManifest, loadErr := provenance.Load(manifestPath)
		if loadErr != nil && !errors.IsNotFound(loadErr) {
			slog.Warn("failed to load provenance manifest for collision check", "error", loadErr)
		}

		entries, err := os.ReadDir(opts.TargetCommandsDir)
		if err == nil {
			// Build reverse map: flat name -> source keys that want this name
			flatToSource := make(map[string][]string)
			for sourceKey, flatName := range flatNames {
				flatToSource[flatName] = append(flatToSource[flatName], sourceKey)
			}

			seen := make(map[string]bool)
			for _, entry := range entries {
				entryName := entry.Name()
				// Strip .md extension for file entries to match flat name
				if !entry.IsDir() && strings.HasSuffix(entryName, ".md") {
					entryName = strings.TrimSuffix(entryName, ".md")
				}

				// A command with companions exists as both dir and promoted .md file.
				// Skip the second occurrence to avoid duplicate warnings/log lines.
				if seen[entryName] {
					continue
				}
				seen[entryName] = true

				sourceKeys, isFlat := flatToSource[entryName]
				if !isFlat {
					continue // Not a name we're trying to flatten to
				}

				// Check if the existing entry is knossos-owned via provenance
				isKnossosOwned := false
				if oldManifest != nil {
					// Check both dir and file provenance keys
					for _, provenanceKey := range []string{
						"commands/" + entryName + "/",
						"commands/" + entryName + ".md",
						"commands/" + entryName,
					} {
						if pe, ok := oldManifest.Entries[provenanceKey]; ok && pe.Owner == provenance.OwnerKnossos {
							isKnossosOwned = true
							break
						}
					}
				}

				if isKnossosOwned {
					continue // Safe to overwrite knossos-owned entries
				}

				if opts.OverwriteDiverged {
					slog.Info("flat name overwriting user entry", "name", entryName, "reason", "--overwrite-diverged")
					continue
				}

				// User-owned or untracked entry -- knossos yields.
				// Surface a warning so the user knows why the flat name was not applied
				// and can resolve the conflict manually (rename or remove the user entry).
				for _, sourceKey := range sourceKeys {
					warnings = append(warnings, fmt.Sprintf(
						"namespace collision: flat name %q conflicts with existing user-owned entry %q; knossos falling back to source-path routing for %q",
						entryName, entryName, sourceKey,
					))
					slog.Warn("flat name collides with existing user entry, falling back to source path", "name", entryName, "source", sourceKey)
					delete(flatNames, sourceKey)
				}
			}
		}
	}

	return flatNames, warnings
}

// InjectCompanionHideFrontmatter adds user-invocable: false to companion file content.
// If the file has existing YAML frontmatter, it merges the field into the existing block.
// If the file has no frontmatter, it prepends a new frontmatter block.
func InjectCompanionHideFrontmatter(content []byte) []byte {
	// Check if content starts with frontmatter delimiter
	if bytes.HasPrefix(content, []byte("---\n")) {
		// Find closing delimiter
		searchStart := 4
		var endIndex int
		if idx := bytes.Index(content[searchStart:], []byte("\n---\n")); idx != -1 {
			endIndex = searchStart + idx + 1 // +1 to include the \n before ---
			// Insert "user-invocable: false\n" just before closing delimiter
			result := make([]byte, 0, len(content)+len("user-invocable: false\n"))
			result = append(result, content[:endIndex]...)
			result = append(result, []byte("user-invocable: false\n")...)
			result = append(result, content[endIndex:]...)
			return result
		} else if idx := bytes.Index(content[searchStart:], []byte("\n---\r\n")); idx != -1 {
			endIndex = searchStart + idx + 1 // +1 to include the \n before ---
			result := make([]byte, 0, len(content)+len("user-invocable: false\n"))
			result = append(result, content[:endIndex]...)
			result = append(result, []byte("user-invocable: false\n")...)
			result = append(result, content[endIndex:]...)
			return result
		}
		// No closing delimiter found, fall through to prepend
	} else if bytes.HasPrefix(content, []byte("---\r\n")) {
		searchStart := 5
		var endIndex int
		if idx := bytes.Index(content[searchStart:], []byte("\r\n---\r\n")); idx != -1 {
			endIndex = searchStart + idx + 2 // +2 to include the \r\n before ---
			result := make([]byte, 0, len(content)+len("user-invocable: false\r\n"))
			result = append(result, content[:endIndex]...)
			result = append(result, []byte("user-invocable: false\r\n")...)
			result = append(result, content[endIndex:]...)
			return result
		} else if idx := bytes.Index(content[searchStart:], []byte("\r\n---\n")); idx != -1 {
			endIndex = searchStart + idx + 2 // +2 to include the \r\n before ---
			result := make([]byte, 0, len(content)+len("user-invocable: false\r\n"))
			result = append(result, content[:endIndex]...)
			result = append(result, []byte("user-invocable: false\r\n")...)
			result = append(result, content[endIndex:]...)
			return result
		}
		// No closing delimiter found, fall through to prepend
	}

	// No frontmatter: prepend a new frontmatter block
	prefix := []byte("---\nuser-invocable: false\n---\n\n")
	result := make([]byte, 0, len(prefix)+len(content))
	result = append(result, prefix...)
	result = append(result, content...)
	return result
}
