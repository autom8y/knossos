package ask

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/search"
)

type cmdContext struct {
	common.BaseContext
	rootCmd *cobra.Command
}

// NewAskCmd creates the ari ask command.
func NewAskCmd(rootCmd *cobra.Command, outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
		rootCmd: rootCmd,
	}

	var limit int
	var domain string

	cmd := &cobra.Command{
		Use:   "ask [query]",
		Short: "Find the right command for any task",
		Long: `Natural language query interface to the knossos CLI surface.

Ask a question in plain English and get ranked suggestions for commands,
rites, agents, and workflows that match your intent.

Without a project context, searches CLI commands and concepts.
With a project, also searches rites, agents, dromena, and routing.`,
		Example: `  ari ask "how do I release my project?"
  ari ask "code quality"
  ari ask "start a session"
  ari ask -o json "release"
  ari ask --domain=rite "ecosystem"
  ari ask --limit 10 "session"`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			query := args[0]
			printer := ctx.GetPrinter(output.FormatText)
			resolver := ctx.GetResolver()

			// Build search index from the root command and project resolver.
			idx := search.Build(ctx.rootCmd, resolver)

			// Parse domain filter (comma-separated list).
			opts := search.SearchOptions{Limit: limit}
			if domain != "" {
				for _, d := range strings.Split(domain, ",") {
					opts.Domains = append(opts.Domains, search.Domain(strings.TrimSpace(d)))
				}
			}

			// Execute the search.
			results := idx.Search(query, opts)

			// Build the output struct.
			out := AskOutput{
				Query: query,
				Total: len(results),
			}

			// Inject active rite context when a project is available.
			if resolver != nil && resolver.ProjectRoot() != "" {
				activeRite := resolver.ReadActiveRite()
				if activeRite != "" {
					out.Context = "Active rite: " + activeRite
				}
			}

			for i, r := range results {
				out.Results = append(out.Results, AskResultEntry{
					Rank:    i + 1,
					Name:    r.Name,
					Domain:  string(r.Domain),
					Summary: r.Summary,
					Action:  r.Action,
					Score:   r.Score,
				})
			}

			return printer.Print(out)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", search.DefaultLimit, "Maximum results to return")
	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain (comma-separated): command,concept,rite,agent,dromena,routing")

	common.SetNeedsProject(cmd, false, true)
	return cmd
}
