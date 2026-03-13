package inscription

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
)

// SyncInscriptionOptions configures the core inscription sync operation.
// Both Pipeline.Sync() and the materialization pipeline delegate to SyncInscription.
type SyncInscriptionOptions struct {
	// ChannelDir is the path to the channel directory (e.g. .claude/ or .gemini/).
	ChannelDir string

	// RenderCtx provides template rendering data (agents, rite name, vars).
	RenderCtx *RenderContext

	// ActiveRite is the rite name to set in the manifest (empty = no change).
	ActiveRite string

	// TemplateDir is the filesystem path to template files (used when TemplateFS is nil).
	TemplateDir string

	// TemplateFS is an optional embedded FS for templates. Takes priority over TemplateDir.
	TemplateFS fs.FS

	// UpdateManifest controls whether manifest hashes are updated after merge.
	// Set to true for full syncs, false for minimal/cross-cutting mode.
	UpdateManifest bool

	// ContextFilename is the name of the context file (e.g. "CLAUDE.md" or "GEMINI.md").
	// Defaults to "CLAUDE.md" if empty.
	ContextFilename string
}

// SyncInscriptionResult contains the result of an inscription sync.
type SyncInscriptionResult struct {
	// Written is true if CLAUDE.md content changed and was written to disk.
	Written bool

	// MergeResult contains merge details (regions synced, conflicts, etc.).
	MergeResult *MergeResult

	// LegacyBackupPath is the path to a legacy CLAUDE.md backup, if migration occurred.
	LegacyBackupPath string

	// ManifestVersion is the current inscription version after sync.
	ManifestVersion string
}

// SyncInscription is the single canonical path for inscription (context file) generation.
// Both the standalone `ari inscription sync` and the materialization pipeline
// delegate to this function for the core merge/write logic.
//
// The function:
//  1. Loads or creates KNOSSOS_MANIFEST.yaml
//  2. Creates a generator (from TemplateFS or TemplateDir)
//  3. Generates all section content
//  4. Detects and backs up legacy (pre-marker) context files
//  5. Merges generated content with existing context file
//  6. Writes only if content changed (avoids harness file watcher triggers)
//  7. Optionally updates manifest hashes and version
func SyncInscription(opts SyncInscriptionOptions) (*SyncInscriptionResult, error) {
	// 1. Load or create KNOSSOS_MANIFEST.yaml in .knossos/
	projectRoot := filepath.Dir(opts.ChannelDir)
	knossosManifestPath := filepath.Join(projectRoot, ".knossos", "KNOSSOS_MANIFEST.yaml")

	loader := NewManifestLoader(projectRoot)
	loader.ManifestPath = knossosManifestPath

	manifestExists := loader.Exists()

	manifest, err := loader.LoadOrCreate()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load or create KNOSSOS_MANIFEST.yaml", err)
	}

	// Adopt new default sections that may have been added in newer knossos versions.
	// This is how existing satellites pick up new inscription sections (e.g. "know").
	manifest.AdoptNewDefaults()

	// Update active rite in manifest if provided
	if opts.ActiveRite != "" {
		manifest.SetActiveRite(opts.ActiveRite)
	}

	// Save manifest if newly created or defaults were adopted
	if !manifestExists {
		if err := loader.Save(manifest); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
	}

	// 2. Create generator from TemplateFS or TemplateDir
	var generator *Generator
	if opts.TemplateFS != nil {
		generator = NewGeneratorWithFS(opts.TemplateFS, manifest, opts.RenderCtx)
	} else {
		generator = NewGenerator(opts.TemplateDir, manifest, opts.RenderCtx)
	}

	// 3. Generate all sections
	sections, err := generator.GenerateAll()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to generate CLAUDE.md sections", err)
	}

	// 4. Read existing context file and detect legacy format
	contextFilename := opts.ContextFilename
	if contextFilename == "" {
		contextFilename = "CLAUDE.md"
	}
	contextFilePath := filepath.Join(opts.ChannelDir, contextFilename)
	existingContent := ""
	legacyBackupPath := ""

	if data, err := os.ReadFile(contextFilePath); err == nil {
		existingContent = string(data)

		// Detect legacy context file (no KNOSSOS markers) and backup before overwriting
		if !strings.Contains(existingContent, "<!-- KNOSSOS:START") && len(existingContent) > 0 {
			legacyBackupPath = contextFilePath + ".legacy-backup"
			if err := os.WriteFile(legacyBackupPath, data, 0644); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to backup legacy context file", err)
			}
			existingContent = ""
		}
	}

	// 5. Merge sections
	merger := NewMerger(manifest, generator)
	mergeResult, err := merger.MergeRegions(existingContent, sections)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to merge CLAUDE.md regions", err)
	}

	// 6. Write only if content changed
	written, err := fileutil.WriteIfChanged(contextFilePath, []byte(mergeResult.Content), 0644)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write context file", err)
	}

	// 7. Update manifest hashes and version
	if opts.UpdateManifest {
		merger.UpdateManifestHashes(mergeResult)
		loader.IncrementVersion(manifest)
		if err := loader.Save(manifest); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
	}

	return &SyncInscriptionResult{
		Written:          written,
		MergeResult:      mergeResult,
		LegacyBackupPath: legacyBackupPath,
		ManifestVersion:  manifest.InscriptionVersion,
	}, nil
}
