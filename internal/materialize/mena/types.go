// Package mena provides the mena projection engine for the materialize pipeline.
// It handles collection, type detection, namespace resolution, extension stripping,
// and file routing for dromena (commands) and legomena (skills).
package mena

import (
	"io/fs"
	"strings"

	"github.com/autom8y/knossos/internal/provenance"
)

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
}

// MenaProjectionResult reports what the projection did.
type MenaProjectionResult struct {
	CommandsProjected []string // Relative paths of files written to commands/
	SkillsProjected   []string // Relative paths of files written to skills/
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
}

// MenaResolution holds the resolved mena entries after collection and namespace flattening.
// Returned by CollectMena for reuse by both rite-scope (SyncMena) and user-scope (syncUserMena).
type MenaResolution struct {
	Entries     map[string]MenaResolvedEntry      // source key -> directory entry
	Standalones map[string]MenaResolvedStandalone // source key -> standalone file
}

// MenaResolvedEntry represents a resolved leaf mena directory with flat name and type.
type MenaResolvedEntry struct {
	Source   MenaSource
	FlatName string // after resolveNamespace (e.g., "spike" from "operations/spike")
	MenaType string // "dro" or "lego"
}

// MenaResolvedStandalone represents a resolved standalone mena file with flat name and type.
type MenaResolvedStandalone struct {
	SrcPath  string
	RelPath  string // original relative path (e.g., "operations/architect.dro.md")
	FlatName string // after resolveNamespace + strip (e.g., "architect.md")
	MenaType string // "dro" or "lego"
}

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

// DetectMenaType determines content type from file extension convention.
// Files with .dro.md extension are dromena (invokable, project to .claude/commands/).
// Files with .lego.md extension are legomena (reference, project to .claude/skills/).
// Returns "dro" as default for backward compatibility.
func DetectMenaType(filename string) string {
	if strings.Contains(filename, ".dro.") {
		return "dro"
	}
	if strings.Contains(filename, ".lego.") {
		return "lego"
	}
	return "dro" // default for backward compat
}

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
