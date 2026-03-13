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
	TargetDir    string // ~/{channel}/{agents,hooks} (single target)
	CommandsDir  string // ~/{channel}/commands/ (mena only)
	SkillsDir    string // ~/{channel}/skills/ (mena only)
	Nested       bool   // true for mena, hooks
}

// ResolveUserResourcesForChannel resolves source/target paths for user-scope resources
// targeting a specific channel.
func ResolveUserResourcesForChannel(knossosHome, channel string) ([]UserResourcePaths, error) {
	return []UserResourcePaths{
		{
			ResourceType: ResourceAgents,
			SourceDir:    filepath.Join(knossosHome, "agents"),
			TargetDir:    paths.UserAgentsDirForChannel(channel),
			Nested:       false,
		},
		{
			ResourceType: ResourceMena,
			SourceDir:    filepath.Join(knossosHome, "mena"),
			CommandsDir:  paths.UserCommandsDirForChannel(channel),
			SkillsDir:    paths.UserSkillsDirForChannel(channel),
			Nested:       true,
		},
		{
			ResourceType: ResourceHooks,
			SourceDir:    filepath.Join(knossosHome, "hooks"),
			TargetDir:    paths.UserHooksDirForChannel(channel),
			Nested:       true,
		},
	}, nil
}

// ResolveUserResources resolves source/target paths for user-scope resources.
// Deprecated: Use ResolveUserResourcesForChannel for channel-aware paths.
// HA-CC: "claude" is the default CC channel name; this deprecated wrapper hard-codes it.
func ResolveUserResources(knossosHome string) ([]UserResourcePaths, error) {
	return ResolveUserResourcesForChannel(knossosHome, "claude")
}
