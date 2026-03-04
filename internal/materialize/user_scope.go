// Package materialize re-exports user-scope types from the userscope sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/userscope"
)

// syncUserScope delegates to the userscope sub-package.
func (m *Materializer) syncUserScope(opts SyncOptions) (*UserScopeResult, error) {
	return userscope.SyncUserScope(userscope.SyncUserScopeParams{
		Resolver:       m.resolver,
		EmbeddedAgents: m.embeddedAgents,
		EmbeddedMena:   m.embeddedMena,
		EmbeddedRites:  m.sourceResolver.EmbeddedFS,
		Opts: userscope.SyncOptions{
			Resource:          opts.Resource,
			DryRun:            opts.DryRun,
			Recover:           opts.Recover,
			OverwriteDiverged: opts.OverwriteDiverged,
			KeepOrphans:       opts.KeepOrphans,
		},
	})
}
