// Package materialize re-exports source resolution types from the source sub-package.
package materialize

import (
	"path/filepath"

	"github.com/autom8y/knossos/internal/materialize/source"
	"github.com/autom8y/knossos/internal/paths"
)

// Type aliases for backward compatibility. Core and test code can continue
// using materialize.SourceType, materialize.ResolvedRite, etc.
type (
	SourceType     = source.SourceType
	RiteSource     = source.RiteSource
	ResolvedRite   = source.ResolvedRite
	SourceResolver = source.SourceResolver
)

// Re-export source type constants.
const (
	SourceProject  = source.SourceProject
	SourceUser     = source.SourceUser
	SourceKnossos  = source.SourceKnossos
	SourceExplicit = source.SourceExplicit
	SourceEmbedded = source.SourceEmbedded
)

// Re-export constructors.
var NewSourceResolver = source.NewSourceResolver
var NewSourceResolverWithPaths = source.NewSourceResolverWithPaths

// UserResourcePaths contains resolved source/target paths for user-scope resources.
// Kept in core because it references SyncResource from sync_types.go.
type UserResourcePaths struct {
	ResourceType SyncResource
	SourceDir    string // $KNOSSOS_HOME/{agents,mena,hooks}
	TargetDir    string // ~/.claude/{agents,hooks} (single target)
	CommandsDir  string // ~/.claude/commands/ (mena only)
	SkillsDir    string // ~/.claude/skills/ (mena only)
	Nested       bool   // true for mena, hooks
}

// ResolveUserResources resolves source/target paths for user-scope resources.
func ResolveUserResources(knossosHome string) ([]UserResourcePaths, error) {
	return []UserResourcePaths{
		{
			ResourceType: ResourceAgents,
			SourceDir:    filepath.Join(knossosHome, "agents"),
			TargetDir:    paths.UserAgentsDir(),
			Nested:       false,
		},
		{
			ResourceType: ResourceMena,
			SourceDir:    filepath.Join(knossosHome, "mena"),
			CommandsDir:  paths.UserCommandsDir(),
			SkillsDir:    paths.UserSkillsDir(),
			Nested:       true,
		},
		{
			ResourceType: ResourceHooks,
			SourceDir:    filepath.Join(knossosHome, "hooks"),
			TargetDir:    paths.UserHooksDir(),
			Nested:       true,
		},
	}, nil
}
