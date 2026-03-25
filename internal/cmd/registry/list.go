package registry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
)

type listOptions struct {
	org   string
	repo  string
	stale bool
}

// domainListOutput is the structured output for a single domain row.
type domainListOutput struct {
	Repo          string  `json:"repo" yaml:"repo"`
	Domain        string  `json:"domain" yaml:"domain"`
	QualifiedName string  `json:"qualified_name" yaml:"qualified_name"`
	GeneratedAt   string  `json:"generated_at" yaml:"generated_at"`
	ExpiresAfter  string  `json:"expires_after" yaml:"expires_after"`
	SourceHash    string  `json:"source_hash" yaml:"source_hash"`
	Confidence    float64 `json:"confidence" yaml:"confidence"`
	Stale         bool    `json:"stale" yaml:"stale"`
}

// Text implements output.Textable for a single domain row.
func (d domainListOutput) Text() string {
	staleMarker := " "
	if d.Stale {
		staleMarker = "*"
	}
	return fmt.Sprintf("%s %-20s %-30s %-45s %-20s %-15s %s",
		staleMarker,
		truncate(d.Repo, 20),
		truncate(d.Domain, 30),
		truncate(d.QualifiedName, 45),
		truncate(d.GeneratedAt, 20),
		truncate(d.ExpiresAfter, 15),
		truncate(d.SourceHash, 8),
	)
}

// domainListResultOutput is the top-level output for the list command.
type domainListResultOutput struct {
	Org     string             `json:"org" yaml:"org"`
	Count   int                `json:"count" yaml:"count"`
	Domains []domainListOutput `json:"domains" yaml:"domains"`
}

// Text implements output.Textable — prints header + rows.
func (r domainListResultOutput) Text() string {
	if len(r.Domains) == 0 {
		return fmt.Sprintf("No domains found for org %s. Run 'ari registry sync' first.\n", r.Org)
	}
	var sb strings.Builder
	header := fmt.Sprintf("  %-20s %-30s %-45s %-20s %-15s %s",
		"REPO", "DOMAIN", "QUALIFIED NAME", "GENERATED", "EXPIRES", "HASH")
	sb.WriteString(header + "\n")
	sb.WriteString(strings.Repeat("-", len(header)) + "\n")
	for _, d := range r.Domains {
		sb.WriteString(d.Text() + "\n")
	}
	sb.WriteString(fmt.Sprintf("\n%d domain(s) — * = stale\n", r.Count))
	return sb.String()
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	opts := listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cataloged knowledge domains",
		Long: `Display the domain catalog for the active org.

Reads from the persisted catalog at $XDG_DATA_HOME/knossos/registry/{org}/domains.yaml.
Run 'ari registry sync' first to populate the catalog.

Output columns: REPO, DOMAIN, QUALIFIED NAME, GENERATED, EXPIRES, HASH

Examples:
  ari registry list
  ari registry list --repo knossos
  ari registry list --stale
  ari registry list -o json`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.org, "org", "", "Override active org")
	cmd.Flags().StringVar(&opts.repo, "repo", "", "Filter by repository name")
	cmd.Flags().BoolVar(&opts.stale, "stale", false, "Show only stale domains")

	return cmd
}

func runList(ctx *cmdContext, opts listOptions) error {
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

	// Build output rows, applying filters.
	var rows []domainListOutput
	for _, repo := range catalog.Repos {
		if opts.repo != "" && repo.Name != opts.repo {
			continue
		}
		for _, d := range repo.Domains {
			isStale := d.IsStale()
			if opts.stale && !isStale {
				continue
			}
			rows = append(rows, domainListOutput{
				Repo:          repo.Name,
				Domain:        d.Domain,
				QualifiedName: d.QualifiedName,
				GeneratedAt:   d.GeneratedAt,
				ExpiresAfter:  d.ExpiresAfter,
				SourceHash:    d.SourceHash,
				Confidence:    d.Confidence,
				Stale:         isStale,
			})
		}
	}

	result := domainListResultOutput{
		Org:     orgName,
		Count:   len(rows),
		Domains: rows,
	}

	_ = printer.Print(result)
	return nil
}

// truncate shortens a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
