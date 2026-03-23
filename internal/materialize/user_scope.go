// Package materialize re-exports user-scope types from the userscope sub-package.
package materialize

import (
	"fmt"
	"maps"

	"github.com/autom8y/knossos/internal/materialize/userscope"
	"github.com/autom8y/knossos/internal/paths"
)

// syncUserScope dispatches to single-channel or all-channels depending on opts.Channel.
func (m *Materializer) syncUserScope(opts SyncOptions) (*UserScopeResult, error) {
	if opts.Channel == "all" {
		return m.syncUserScopeAllChannels(opts)
	}
	return m.syncUserScopeSingleChannel(opts)
}

// syncUserScopeSingleChannel syncs user resources to a single channel directory.
func (m *Materializer) syncUserScopeSingleChannel(opts SyncOptions) (*UserScopeResult, error) {
	// Resolve target directory: use m.userChannelDir override (for testing) when set;
	// otherwise resolve from the channel name (e.g., ~/.claude or ~/.gemini).
	userChannelDir := m.userChannelDir
	if userChannelDir == "" {
		var err error
		userChannelDir, err = paths.UserChannelDir(opts.Channel)
		if err != nil {
			return nil, err
		}
	}

	return userscope.SyncUserScope(userscope.SyncUserScopeParams{
		Resolver:       m.resolver,
		EmbeddedAgents: m.embeddedAgents,
		EmbeddedMena:   m.embeddedMena,
		EmbeddedRites:  m.sourceResolver.EmbeddedFS,
		KnossosHome:    m.sourceResolver.KnossosHome(),
		UserChannelDir: userChannelDir,
		Opts: userscope.SyncOptions{
			Resource:          opts.Resource,
			DryRun:            opts.DryRun,
			Recover:           opts.Recover,
			OverwriteDiverged: opts.OverwriteDiverged,
			KeepOrphans:       opts.KeepOrphans,
			Channel:           opts.Channel,
		},
	})
}

// syncUserScopeAllChannels iterates over all channels and calls syncUserScopeSingleChannel
// for each, aggregating totals. On per-channel failure, logs a warning and continues
// (same behaviour as syncRiteScopeAllChannels).
func (m *Materializer) syncUserScopeAllChannels(opts SyncOptions) (*UserScopeResult, error) {
	channels := paths.AllChannels()
	aggregate := &UserScopeResult{
		Status:    "success",
		Resources: make(map[SyncResource]*UserResourceResult),
		Errors:    []UserResourceError{},
	}

	for _, ch := range channels {
		perChannelOpts := opts
		perChannelOpts.Channel = ch.Name()

		chResult, err := m.syncUserScopeSingleChannel(perChannelOpts)
		if err != nil {
			// Non-fatal: log and continue so other channels complete
			aggregate.Errors = append(aggregate.Errors, UserResourceError{
				Resource: userscope.ResourceAll,
				Err:      fmt.Sprintf("channel %s: %s", ch.Name(), err.Error()),
			})
			aggregate.Status = "partial"
			continue
		}

		// Merge resource results (keep channel-specific target paths for the last channel
		// that wrote each resource — aggregation is totals-only for multi-channel)
		maps.Copy(aggregate.Resources, chResult.Resources)
		aggregate.Errors = append(aggregate.Errors, chResult.Errors...)
		aggregate.Totals.Added += chResult.Totals.Added
		aggregate.Totals.Updated += chResult.Totals.Updated
		aggregate.Totals.Skipped += chResult.Totals.Skipped
		aggregate.Totals.Unchanged += chResult.Totals.Unchanged
		aggregate.Totals.Collisions += chResult.Totals.Collisions
	}

	return aggregate, nil
}
