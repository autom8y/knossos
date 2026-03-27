package ask

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/session"
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
	var sessionFlag string

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
  ari ask --limit 10 "session"
  ari ask --session=session-20260308-143022-a1b2c3d4 "what next?"`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			query := args[0]
			printer := ctx.GetPrinter(output.FormatText)
			resolver := ctx.GetResolver()

			// H4: Detect session and build session signals.
			var sessionSignals *search.SessionSignals
			hasProject := resolver != nil && resolver.ProjectRoot() != ""
			if hasProject {
				var sessionID string

				if sessionFlag != "" {
					// Explicit session override.
					sessionID = sessionFlag
				} else {
					// Auto-detect active session.
					id, err := session.FindActiveSession(resolver.SessionsDir())
					if err == nil {
						// err != nil means multiple active sessions -- fall back to no-session (AC-7.4).
						sessionID = id
					}
				}

				if sessionID != "" {
					sessionSignals = readSessionSignals(sessionID, resolver)
				}
			}

			// Build search index from the root command and project resolver.
			idx := search.Build(ctx.rootCmd, resolver)

			// Parse domain filter (comma-separated list).
			opts := search.SearchOptions{
				Limit:   limit,
				Session: sessionSignals,
			}
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
			if hasProject {
				activeRite := resolver.ReadActiveRite()
				if activeRite != "" {
					out.Context = "Active rite: " + activeRite
				}
			}

			// H4: Populate session context in output.
			if sessionSignals != nil {
				out.SessionContext = buildAskSessionContext(sessionSignals)
			}

			for i, r := range results {
				entry := AskResultEntry{
					Rank:    i + 1,
					Name:    r.Name,
					Domain:  string(r.Domain),
					Summary: r.Summary,
					Action:  r.Action,
					Score:   r.Score,
				}

				// For knowledge results, extract repo source and freshness annotation.
				if r.Domain == search.DomainKnowledge {
					entry.Source = know.RepoFromQualifiedName(r.Name)
					entry.Freshness = r.Description // Contains freshness annotation.
				}

				out.Results = append(out.Results, entry)
			}

			return printer.Print(out)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", search.DefaultLimit, "Maximum results to return")
	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain (comma-separated): command,concept,rite,agent,dromena,routing,session,knowledge")
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session ID override (auto-detects if omitted)")

	common.SetNeedsProject(cmd, false, true)
	return cmd
}

// readSessionSignals loads session signals for an active session.
// sessionID: either auto-detected or from --session flag.
// resolver: path resolver for computing file paths.
// Returns nil if session state cannot be read (fail-open).
func readSessionSignals(sessionID string, resolver interface {
	SessionContextFile(string) string
	SessionEventsFile(string) string
}) *search.SessionSignals {
	contextPath := resolver.SessionContextFile(sessionID)
	signals := search.ReadSessionState(contextPath)
	if signals == nil {
		return nil
	}

	eventsPath := resolver.SessionEventsFile(sessionID)
	signals.Activity = search.TailReadEvents(eventsPath, 0)

	return signals
}

// buildAskSessionContext converts SessionSignals to the output AskSessionContext.
func buildAskSessionContext(signals *search.SessionSignals) *AskSessionContext {
	sc := &AskSessionContext{
		SessionID:  signals.SessionID,
		Phase:      signals.Phase,
		Rite:       signals.Rite,
		Complexity: signals.Complexity,
		Initiative: signals.Initiative,
	}

	if signals.Activity != nil {
		totalAgentTasks := 0
		for _, count := range signals.Activity.AgentTasks {
			totalAgentTasks += count
		}

		as := &AskActivitySummary{
			FileChanges: signals.Activity.FileChangeCount,
			AgentTasks:  totalAgentTasks,
		}

		// Compute last event age.
		if signals.Activity.LastEventTS != "" {
			if t, err := time.Parse(time.RFC3339, signals.Activity.LastEventTS); err == nil {
				age := time.Since(t)
				as.LastEventAge = formatDuration(age)
			} else if t, err := time.Parse("2006-01-02T15:04:05.000Z", signals.Activity.LastEventTS); err == nil {
				age := time.Since(t)
				as.LastEventAge = formatDuration(age)
			}
		}

		// Only include activity summary if it has meaningful data.
		if as.FileChanges > 0 || as.AgentTasks > 0 || as.LastEventAge != "" {
			sc.ActivitySummary = as
		}
	}

	return sc
}

// formatDuration converts a duration to a human-readable string like "2m", "1h", "3d".
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	switch {
	case d < time.Minute:
		return "<1m"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
