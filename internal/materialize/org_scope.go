// Package materialize delegates org-scope sync to the orgscope sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/orgscope"
)

// syncOrgScope delegates to the orgscope sub-package.
func (m *Materializer) syncOrgScope(opts SyncOptions) (*OrgScopeResult, error) {
	result, err := orgscope.SyncOrgScope(orgscope.SyncOrgScopeParams{
		OrgName: opts.OrgName,
		DryRun:  opts.DryRun,
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
