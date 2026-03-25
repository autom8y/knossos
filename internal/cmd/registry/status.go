package registry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
)

type statusOptions struct {
	org string
}

// registryStatusOutput is the structured output for ari registry status.
type registryStatusOutput struct {
	Org          string `json:"org" yaml:"org"`
	LastSynced   string `json:"last_synced" yaml:"last_synced"`
	DomainCount  int    `json:"domain_count" yaml:"domain_count"`
	RepoCount    int    `json:"repo_count" yaml:"repo_count"`
	StaleCount   int    `json:"stale_count" yaml:"stale_count"`
	CatalogPath  string `json:"catalog_path" yaml:"catalog_path"`
	SchemaVersion string `json:"schema_version" yaml:"schema_version"`
}

// Text implements output.Textable.
func (s registryStatusOutput) Text() string {
	staleNote := ""
	if s.StaleCount > 0 {
		staleNote = fmt.Sprintf(" (%d stale — run 'ari registry sync' to refresh)", s.StaleCount)
	}
	return fmt.Sprintf(`Registry status for org: %s
  Catalog path:  %s
  Schema:        %s
  Last synced:   %s
  Repos:         %d
  Domains:       %d%s
`,
		s.Org,
		s.CatalogPath,
		s.SchemaVersion,
		s.LastSynced,
		s.RepoCount,
		s.DomainCount,
		staleNote,
	)
}

func newStatusCmd(ctx *cmdContext) *cobra.Command {
	opts := statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show registry sync summary",
		Long: `Display a summary of the knowledge domain catalog.

Shows last sync time, domain count, repo count, and stale domain count.

Examples:
  ari registry status
  ari registry status --org autom8y
  ari registry status -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.org, "org", "", "Override active org")

	return cmd
}

func runStatus(ctx *cmdContext, opts statusOptions) error {
	printer := ctx.GetPrinter(output.FormatText)

	orgName := opts.org
	if orgName == "" {
		orgName = config.ActiveOrg()
	}
	if orgName == "" {
		return common.PrintAndReturn(printer,
			errors.New(errors.CodeUsageError, "no active org configured; use --org or set KNOSSOS_ORG"))
	}

	orgCtx, err := config.NewOrgContext(orgName)
	if err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeGeneralError, "build org context", err))
	}

	catalogPath := registryorg.CatalogPath(orgCtx)
	catalog, err := registryorg.LoadCatalog(catalogPath)
	if err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeFileNotFound, fmt.Sprintf("load catalog for org %s", orgName), err))
	}

	result := registryStatusOutput{
		Org:           orgName,
		LastSynced:    catalog.SyncedAt,
		DomainCount:   catalog.DomainCount(),
		RepoCount:     catalog.RepoCount(),
		StaleCount:    catalog.StaleCount(),
		CatalogPath:   catalogPath,
		SchemaVersion: catalog.SchemaVersion,
	}

	_ = printer.Print(result)
	return nil
}
