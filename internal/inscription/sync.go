package inscription

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
)

// CLAUDEmdSyncOptions configures the core CLAUDE.md sync operation.
// Both Pipeline.Sync() and the materialization pipeline delegate to SyncCLAUDEmd.
type CLAUDEmdSyncOptions struct {
	// ClaudeDir is the path to the .claude/ directory.
	ClaudeDir string

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
}

// CLAUDEmdSyncResult contains the result of a CLAUDE.md sync.
type CLAUDEmdSyncResult struct {
	// Written is true if CLAUDE.md content changed and was written to disk.
	Written bool

	// MergeResult contains merge details (regions synced, conflicts, etc.).
	MergeResult *MergeResult

	// LegacyBackupPath is the path to a legacy CLAUDE.md backup, if migration occurred.
	LegacyBackupPath string

	// ManifestVersion is the current inscription version after sync.
	ManifestVersion string
}

// SyncCLAUDEmd is the single canonical path for CLAUDE.md generation.
// Both the standalone `ari inscription sync` and the materialization pipeline
// delegate to this function for the core merge/write logic.
//
// The function:
//  1. Loads or creates KNOSSOS_MANIFEST.yaml
//  2. Creates a generator (from TemplateFS or TemplateDir)
//  3. Generates all section content
//  4. Detects and backs up legacy (pre-marker) CLAUDE.md files
//  5. Merges generated content with existing CLAUDE.md
//  6. Writes only if content changed (avoids CC file watcher triggers)
//  7. Optionally updates manifest hashes and version
func SyncCLAUDEmd(opts CLAUDEmdSyncOptions) (*CLAUDEmdSyncResult, error) {
	// 1. Load or create KNOSSOS_MANIFEST.yaml
	knossosManifestPath := filepath.Join(opts.ClaudeDir, "KNOSSOS_MANIFEST.yaml")
	projectRoot := filepath.Dir(opts.ClaudeDir)

	loader := NewManifestLoader(projectRoot)
	loader.ManifestPath = knossosManifestPath

	manifestExists := loader.Exists()

	manifest, err := loader.LoadOrCreate()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load or create KNOSSOS_MANIFEST.yaml", err)
	}

	// Update active rite in manifest if provided
	if opts.ActiveRite != "" {
		manifest.SetActiveRite(opts.ActiveRite)
	}

	// Save manifest if newly created
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

	// 4. Read existing CLAUDE.md and detect legacy format
	claudeMdPath := filepath.Join(opts.ClaudeDir, "CLAUDE.md")
	existingContent := ""
	legacyBackupPath := ""

	if data, err := os.ReadFile(claudeMdPath); err == nil {
		existingContent = string(data)

		// Detect legacy CLAUDE.md (no KNOSSOS markers) and backup before overwriting
		if !strings.Contains(existingContent, "<!-- KNOSSOS:START") && len(existingContent) > 0 {
			legacyBackupPath = claudeMdPath + ".legacy-backup"
			if err := os.WriteFile(legacyBackupPath, data, 0644); err != nil {
				return nil, errors.Wrap(errors.CodeGeneralError, "failed to backup legacy CLAUDE.md", err)
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
	written, err := fileutil.WriteIfChanged(claudeMdPath, []byte(mergeResult.Content), 0644)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write CLAUDE.md", err)
	}

	// 7. Update manifest hashes and version
	if opts.UpdateManifest {
		merger.UpdateManifestHashes(mergeResult)
		loader.IncrementVersion(manifest)
		if err := loader.Save(manifest); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to save KNOSSOS_MANIFEST.yaml", err)
		}
	}

	return &CLAUDEmdSyncResult{
		Written:          written,
		MergeResult:      mergeResult,
		LegacyBackupPath: legacyBackupPath,
		ManifestVersion:  manifest.InscriptionVersion,
	}, nil
}
