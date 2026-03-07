package rite

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/spf13/cobra"

	ritelib "github.com/autom8y/knossos/internal/rite"
)

type contextOptions struct {
	riteName string
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
  ari rite context                     # Show current rite's context
  ari rite context --rite=10x-dev # Show specific rite's context
  ari rite context --format=yaml       # Output as YAML
  ari rite context --format=json       # Output as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContext(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite name (defaults to active rite)")
	cmd.Flags().StringVar(&opts.format, "format", "markdown", "Output format: markdown, json, yaml")

	return cmd
}

func runContext(ctx *cmdContext, opts contextOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Determine rite name
	riteName := opts.riteName
	if riteName == "" {
		riteName = ctx.getActiveRite()
		if riteName == "" {
			printer.PrintLine("No active rite. Use --rite flag to specify a rite.")
			return nil
		}
	}

	// Create context loader
	loader := ritelib.NewContextLoader(resolver)

	// Load the context
	riteCtx, err := loader.Load(riteName)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Build output based on format
	switch opts.format {
	case "json":
		return printer.Print(riteContextToOutput(riteCtx, loader.HasContextFile(riteName)))
	case "yaml":
		return printer.Print(riteContextToOutput(riteCtx, loader.HasContextFile(riteName)))
	default:
		// Markdown format - raw output
		markdown := riteCtx.ToMarkdown()
		if markdown == "" {
			printer.PrintLine("No context rows defined for rite: " + riteName)
			return nil
		}
		printer.PrintText(markdown)
		return nil
	}
}

// RiteContextOutput is the JSON/YAML output structure.
type RiteContextOutput struct {
	RiteName      string            `json:"rite_name" yaml:"rite_name"`
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

func riteContextToOutput(tc *ritelib.RiteContext, hasContextFile bool) RiteContextOutput {
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

	return RiteContextOutput{
		RiteName:      tc.RiteName,
		DisplayName:   tc.DisplayName,
		Description:   tc.Description,
		Domain:        tc.Domain,
		SchemaVersion: tc.SchemaVersion,
		ContextRows:   rows,
		Metadata:      tc.Metadata,
		Source:        source,
	}
}

// Text implements output.Textable for RiteContextOutput.
func (t RiteContextOutput) Text() string {
	// For text output, defer to markdown
	return ""
}
