// Package materialize re-exports mena types from the mena sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/mena"
)

// Type aliases for backward compatibility.
type (
	MenaSource             = mena.MenaSource
	MenaProjectionMode     = mena.MenaProjectionMode
	MenaFilter             = mena.MenaFilter
	MenaProjectionOptions  = mena.MenaProjectionOptions
	MenaProjectionResult   = mena.MenaProjectionResult
	MenaResolution         = mena.MenaResolution
	MenaResolvedEntry      = mena.MenaResolvedEntry
	MenaResolvedStandalone = mena.MenaResolvedStandalone
)

// Re-export constants.
const (
	MenaProjectionAdditive    = mena.MenaProjectionAdditive
	MenaProjectionDestructive = mena.MenaProjectionDestructive
	ProjectDro                = mena.ProjectDro
	ProjectLego               = mena.ProjectLego
	ProjectAll                = mena.ProjectAll
)

// Re-export functions.
var (
	CollectMena                 = mena.CollectMena
	SyncMena                    = mena.SyncMena
	StripMenaExtension          = mena.StripMenaExtension
	RouteMenaFile               = mena.RouteMenaFile
	DetectMenaType              = mena.DetectMenaType
	ReadMenaFrontmatterFromDir  = mena.ReadMenaFrontmatterFromDir
	ReadMenaFrontmatterFromFile = mena.ReadMenaFrontmatterFromFile
)
