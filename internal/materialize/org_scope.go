// Package materialize delegates org-scope sync to the orgscope sub-package.
package materialize

import (
	"fmt"

	"github.com/autom8y/knossos/internal/materialize/orgscope"
	"github.com/autom8y/knossos/internal/paths"
)

// syncOrgScope dispatches to single-channel or all-channels depending on opts.Channel.
func (m *Materializer) syncOrgScope(opts SyncOptions) (*OrgScopeResult, error) {
	if opts.Channel == "all" {
		return m.syncOrgScopeAllChannels(opts)
	}
	return m.syncOrgScopeSingleChannel(opts)
}

// syncOrgScopeSingleChannel syncs org resources to a single channel directory.
func (m *Materializer) syncOrgScopeSingleChannel(opts SyncOptions) (*OrgScopeResult, error) {
	// Resolve target directory: use m.userChannelDir override (for testing) when set;
	// otherwise resolve from the channel name (e.g., ~/.claude or ~/.gemini).
	userChannelDir := m.userChannelDir
	if userChannelDir == "" {
		userChannelDir = paths.UserChannelDir(opts.Channel)
	}

	result, err := orgscope.SyncOrgScope(orgscope.SyncOrgScopeParams{
		OrgName:       opts.OrgName,
		UserChannelDir: userChannelDir,
		DryRun:        opts.DryRun,
		Channel:       opts.Channel,
	})
	if err != nil {
		return nil, err
	}

	// Convert orgscope result to materialize result type
	return &OrgScopeResult{
		Status:  result.Status,
		Error:   result.Error,
		OrgName: result.OrgName,
		Source:  result.Source,
		Agents:  result.Agents,
		Mena:    result.Mena,
	}, nil
}

// syncOrgScopeAllChannels iterates over all channels and calls syncOrgScopeSingleChannel
// for each. Partial failures are non-fatal (same pattern as rite-scope all-channels).
func (m *Materializer) syncOrgScopeAllChannels(opts SyncOptions) (*OrgScopeResult, error) {
	channels := paths.AllChannels()
	aggregate := &OrgScopeResult{
		Status:  "success",
		OrgName: opts.OrgName,
	}

	var lastSource string
	for _, ch := range channels {
		perChannelOpts := opts
		perChannelOpts.Channel = ch.Name()

		chResult, err := m.syncOrgScopeSingleChannel(perChannelOpts)
		if err != nil {
			// Non-fatal: accumulate error and continue
			if aggregate.Error == "" {
				aggregate.Error = fmt.Sprintf("channel %s: %s", ch.Name(), err.Error())
			} else {
				aggregate.Error += fmt.Sprintf("; channel %s: %s", ch.Name(), err.Error())
			}
			aggregate.Status = "partial"
			continue
		}

		// Aggregate counts from each channel
		aggregate.Agents += chResult.Agents
		aggregate.Mena += chResult.Mena
		if aggregate.OrgName == "" {
			aggregate.OrgName = chResult.OrgName
		}
		if chResult.Source != "" {
			lastSource = chResult.Source
		}
		// Propagate skipped status if org is not configured
		if chResult.Status == "skipped" {
			aggregate.Status = "skipped"
			aggregate.Error = chResult.Error
			break // Both channels will have the same skip reason — no need to continue
		}
	}

	if lastSource != "" {
		aggregate.Source = lastSource
	}

	return aggregate, nil
}
