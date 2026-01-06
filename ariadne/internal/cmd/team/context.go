package team

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/team"
)

type contextOptions struct {
	teamName string
	format   string
}

func newContextCmd(ctx *cmdContext) *cobra.Command {
	var opts contextOptions

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Show rite context for Claude injection",
		Long: `Displays the context injection data for a rite (practice bundle).

This context is injected into Claude sessions when the rite is active.
The output can be formatted as markdown (default), JSON, or YAML.

Examples:
  ari team context                     # Show current rite's context
  ari team context --rite=10x-dev-pack # Show specific rite's context
  ari team context --format=yaml       # Output as YAML
  ari team context --format=json       # Output as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContext(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.teamName, "rite", "r", "", "Rite name (defaults to active rite)")
	cmd.Flags().StringVarP(&opts.teamName, "team", "t", "", "Deprecated: use --rite instead")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "markdown", "Output format: markdown, json, yaml")

	cmd.Flags().MarkDeprecated("team", "use --rite instead")

	return cmd
}

func runContext(ctx *cmdContext, opts contextOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

	// Determine team name
	teamName := opts.teamName
	if teamName == "" {
		teamName = ctx.getActiveTeam()
		if teamName == "" {
			printer.PrintLine("No active team. Use --team flag to specify a team.")
			return nil
		}
	}

	// Create context loader
	loader := team.NewContextLoader(resolver)

	// Load the context
	teamCtx, err := loader.Load(teamName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output based on format
	switch opts.format {
	case "json":
		return printer.Print(teamContextToOutput(teamCtx, loader.HasContextFile(teamName)))
	case "yaml":
		return printer.Print(teamContextToOutput(teamCtx, loader.HasContextFile(teamName)))
	default:
		// Markdown format - raw output
		markdown := teamCtx.ToMarkdown()
		if markdown == "" {
			printer.PrintLine("No context rows defined for team: " + teamName)
			return nil
		}
		printer.PrintText(markdown)
		return nil
	}
}

// TeamContextOutput is the JSON/YAML output structure.
type TeamContextOutput struct {
	TeamName      string            `json:"team_name" yaml:"team_name"`
	DisplayName   string            `json:"display_name,omitempty" yaml:"display_name,omitempty"`
	Description   string            `json:"description,omitempty" yaml:"description,omitempty"`
	Domain        string            `json:"domain,omitempty" yaml:"domain,omitempty"`
	SchemaVersion string            `json:"schema_version" yaml:"schema_version"`
	ContextRows   []ContextRowOut   `json:"context_rows" yaml:"context_rows"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Source        string            `json:"source" yaml:"source"` // "context.yaml" or "orchestrator.yaml"
}

// ContextRowOut is the output representation of a context row.
type ContextRowOut struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

func teamContextToOutput(tc *team.TeamContext, hasContextFile bool) TeamContextOutput {
	rows := make([]ContextRowOut, len(tc.ContextRows))
	for i, r := range tc.ContextRows {
		rows[i] = ContextRowOut{
			Key:   r.Key,
			Value: r.Value,
		}
	}

	source := "orchestrator.yaml (fallback)"
	if hasContextFile {
		source = "context.yaml"
	}

	return TeamContextOutput{
		TeamName:      tc.TeamName,
		DisplayName:   tc.DisplayName,
		Description:   tc.Description,
		Domain:        tc.Domain,
		SchemaVersion: tc.SchemaVersion,
		ContextRows:   rows,
		Metadata:      tc.Metadata,
		Source:        source,
	}
}

// Text implements output.Textable for TeamContextOutput.
func (t TeamContextOutput) Text() string {
	// For text output, defer to markdown
	return ""
}
