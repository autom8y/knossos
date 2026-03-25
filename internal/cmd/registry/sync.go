package registry

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
)

type syncOptions struct {
	org   string
	token string
}

func newSyncCmd(ctx *cmdContext) *cobra.Command {
	opts := syncOptions{}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync the knowledge domain catalog from GitHub",
		Long: `Discover repos and catalog .know/ domains for the active org.

Reads org.yaml repos if configured; otherwise discovers repos via GitHub API.
Persists the catalog to $XDG_DATA_HOME/knossos/registry/{org}/domains.yaml.

The GitHub token is read from --token flag or GITHUB_TOKEN environment variable.
Without a token, requests are rate-limited to 60/hour.

Examples:
  ari registry sync
  ari registry sync --org autom8y
  ari registry sync --token ghp_xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.org, "org", "", "Override active org (default: KNOSSOS_ORG or active-org file)")
	cmd.Flags().StringVar(&opts.token, "token", "", "GitHub token (default: GITHUB_TOKEN env)")

	return cmd
}

func runSync(ctx *cmdContext, opts syncOptions) error {
	printer := ctx.getPrinter()

	// Resolve org name
	orgName := opts.org
	if orgName == "" {
		orgName = config.ActiveOrg()
	}
	if orgName == "" {
		return common.PrintAndReturn(printer,
			errors.New(errors.CodeUsageError, "no active org configured; use --org or set KNOSSOS_ORG"))
	}

	// Resolve GitHub token
	token := opts.token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	// Build org context
	orgCtx, err := config.NewOrgContext(orgName)
	if err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeGeneralError, "build org context", err))
	}

	// Build GitHub client
	ghClient := registryorg.NewGitHubClient(http.DefaultClient, token)

	// Run sync
	catalog, err := registryorg.SyncRegistry(orgCtx, ghClient)
	if err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeNetworkError, fmt.Sprintf("sync registry for org %s", orgName), err))
	}

	// Persist catalog
	catalogPath := registryorg.CatalogPath(orgCtx)
	if err := registryorg.SaveCatalog(catalogPath, catalog); err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodePermissionDenied, fmt.Sprintf("save catalog to %s", catalogPath), err))
	}

	_ = printer.Print(map[string]any{
		"status":       "synced",
		"org":          orgName,
		"repos":        catalog.RepoCount(),
		"domains":      catalog.DomainCount(),
		"stale":        catalog.StaleCount(),
		"synced_at":    catalog.SyncedAt,
		"catalog_path": catalogPath,
	})

	return nil
}
