package explain

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

type cmdContext struct {
	common.BaseContext
}

// NewExplainCmd creates the ari explain command.
func NewExplainCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "explain [concept]",
		Short: "Explain a knossos concept",
		Long: `Look up definitions for knossos domain concepts like rite, session,
agent, mena, dromena, legomena, and more.

With no argument, lists all concepts with one-line summaries.
With a concept name, shows the full definition with project-aware context.

Examples:
  ari explain              # List all concepts
  ari explain rite         # Full definition of "rite"
  ari explain -o json      # All concepts as JSON
  ari explain rite -o json # Single concept as JSON`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.GetPrinter(output.FormatText)

			// Build resolver -- GetContext checks resolver.ProjectRoot() == "" and
			// returns empty string when no project context is available.
			resolver := ctx.GetResolver()

			if len(args) == 0 {
				// List all concepts
				return listConcepts(printer)
			}

			// Lookup single concept
			return lookupConcept(printer, resolver, args[0])
		},
	}

	common.SetNeedsProject(cmd, false, true)
	return cmd
}

// listConcepts renders the all-concepts table.
func listConcepts(printer *output.Printer) error {
	concepts := AllConcepts()
	list := ConceptListOutput{
		Concepts: make([]ConceptSummary, len(concepts)),
	}
	for i, c := range concepts {
		list.Concepts[i] = ConceptSummary{
			Name:    c.Name,
			Summary: c.Summary,
		}
	}
	return printer.Print(list)
}

// lookupConcept resolves and renders a single concept.
func lookupConcept(printer *output.Printer, resolver *paths.Resolver, input string) error {
	entry, err := LookupConcept(input)
	if err != nil {
		return err
	}

	result := ConceptOutput{
		Concept:     entry.Name,
		DisplayName: entry.DisplayName,
		Summary:     entry.Summary,
		Description: entry.Description,
		SeeAlso:     entry.SeeAlso,
	}

	// Inject project context
	result.ProjectContext = GetContext(entry.Name, resolver)

	return printer.Print(result)
}
