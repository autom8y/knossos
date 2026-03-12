// Package mena provides the mena projection engine for the materialize pipeline.
// It handles collection, type detection, namespace resolution, extension stripping,
// and file routing for dromena (commands) and legomena (skills).
package mena

import (
	"io/fs"

	"github.com/autom8y/knossos/internal/materialize/compiler"
	menapkg "github.com/autom8y/knossos/internal/mena"
	"github.com/autom8y/knossos/internal/provenance"
)

type ChannelCompiler = compiler.ChannelCompiler

// MenaSource is a re-export alias for menapkg.MenaSource.
// The type is defined in internal/mena to allow shared use by the registry validator.
type MenaSource = menapkg.MenaSource

// MenaProjectionMode controls whether projection is additive or destructive.
type MenaProjectionMode int

const (
	// MenaProjectionAdditive adds/updates files without removing unmanaged content.
	// Used by user scope sync (ari sync --scope=user).
	MenaProjectionAdditive MenaProjectionMode = iota

	// MenaProjectionDestructive wipes target commands/ and skills/ directories
	// before projecting. Used by rite scope sync (ari sync --scope=rite).
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

	// KnossosDir is the .knossos/ directory for the project.
	// Used to locate PROVENANCE_MANIFEST.yaml (now at .knossos/PROVENANCE_MANIFEST.yaml).
	// When empty, falls back to filepath.Join(filepath.Dir(filepath.Dir(TargetCommandsDir)), ".knossos").
	KnossosDir string

	// OverwriteDiverged allows overwriting user-owned/untracked entries
	// that collide with flat-name projection. When false (default),
	// knossos yields and falls back to source-path routing.
	OverwriteDiverged bool

	// RiteName is the name of the rite being synced. When non-empty,
	// stale mena cleanup removes entries from rites NOT in the active
	// dependency chain ({RiteName} ∪ ActiveDeps).
	RiteName string

	// ActiveDeps is the rite dependency chain (e.g., ["shared", "dep1"]).
	// Used with RiteName to determine which rite sources are active.
	// Entries from rites outside this chain are cleaned as stale.
	ActiveDeps []string

	// Compiler handles channel-specific format transforms.
	Compiler ChannelCompiler

	// Channel identifies the AI assistant channel (e.g., "gemini").
	// Used for provenance recording. Empty defaults to claude behavior.
	Channel string
}

// MenaProjectionResult reports what the projection did.
type MenaProjectionResult struct {
	CommandsProjected []string // Relative paths of files written to commands/
	SkillsProjected   []string // Relative paths of files written to skills/
	Warnings          []string // Non-fatal diagnostic messages (e.g., namespace collisions)
}

// menaCollectedEntry represents a leaf mena directory collected for routing.
type menaCollectedEntry struct {
	source      MenaSource
	name        string
	sourceIndex int    // index into sources array (higher = higher priority)
	menaType    string // "dro" or "lego", detected during Pass 1 collection
}

// menaStandaloneFile represents a standalone file in a grouping directory.
type menaStandaloneFile struct {
	srcPath     string
	relPath     string // e.g., "navigation/rite.dro.md"
	sourceIndex int    // index into sources array (higher = higher priority)
	isEmbedded  bool   // true for standalone files from embedded FS
	fsys        fs.FS  // non-nil for embedded FS standalones (used for reading)
}

// MenaResolution holds the resolved mena entries after collection and namespace flattening.
// Returned by CollectMena for reuse by both rite-scope (SyncMena) and user-scope (syncUserMena).
type MenaResolution struct {
	Entries     map[string]MenaResolvedEntry      // source key -> directory entry
	Standalones map[string]MenaResolvedStandalone // source key -> standalone file
	Warnings    []string                          // Non-fatal diagnostics from namespace resolution (e.g., collisions)
}

// MenaResolvedEntry represents a resolved leaf mena directory with flat name and type.
type MenaResolvedEntry struct {
	Source   MenaSource
	FlatName string // after resolveNamespace (e.g., "spike" from "operations/spike")
	MenaType string // "dro" or "lego"
}

// MenaResolvedStandalone represents a resolved standalone mena file with flat name and type.
type MenaResolvedStandalone struct {
	SrcPath    string
	RelPath    string // original relative path (e.g., "operations/architect.dro.md")
	FlatName   string // after resolveNamespace + strip (e.g., "architect.md")
	MenaType   string // "dro" or "lego"
	isEmbedded bool   // true for standalone files from embedded FS
	fsys       fs.FS  // non-nil for embedded FS standalones (used for reading)
}

// Re-export moved utility functions from internal/mena/ leaf package.
// These are re-exported so that existing callers within this package and
// via internal/materialize/mena.go continue to work without changes.
var (
	StripMenaExtension = menapkg.StripMenaExtension
	RouteMenaFile      = menapkg.RouteMenaFile
	DetectMenaType     = menapkg.DetectMenaType
)

// ReadMenaFrontmatterFromDir reads the INDEX file from a filesystem directory,
// parses its YAML frontmatter, and returns the result.
func ReadMenaFrontmatterFromDir(dirPath string) MenaFrontmatter {
	return readMenaFrontmatterFromDir(dirPath)
}

// ReadMenaFrontmatterFromFile reads a standalone mena file and parses its
// YAML frontmatter.
func ReadMenaFrontmatterFromFile(filePath string) MenaFrontmatter {
	return readMenaFrontmatterFromFile(filePath)
}
