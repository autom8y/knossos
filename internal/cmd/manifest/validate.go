package manifest

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/manifest"
	"github.com/autom8y/knossos/internal/output"
)

type validateOptions struct {
	schema string
	strict bool
}

func newValidateCmd(ctx *cmdContext) *cobra.Command {
	var opts validateOptions

	cmd := &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate manifest against schema",
		Long:  `Validates a manifest file against its JSON schema.`,
		Args:  common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(ctx, args[0], opts)
		},
		SilenceUsage: true, // Don't print usage on errors
	}

	cmd.Flags().StringVar(&opts.schema, "schema", "", "Schema name to validate against (auto-detects if not specified)")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Fail on additional properties not in schema")

	return cmd
}

func runValidate(ctx *cmdContext, path string, opts validateOptions) error {
	printer := ctx.getPrinter()

	// Load manifest
	m, err := manifest.Load(path)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Determine schema name
	schemaName := opts.schema
	if schemaName == "" {
		detected, err := manifest.DetectSchemaFromPath(path)
		if err != nil {
			return common.PrintAndReturn(printer, err)
		}
		schemaName = detected
	}

	// Create validator
	validator, err := ctx.getSchemaValidator()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Validate
	result, err := validator.Validate(m, schemaName, opts.strict)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Convert to output format
	issues := make([]output.ManifestValidationIssue, len(result.Issues))
	for i, issue := range result.Issues {
		issues[i] = output.ManifestValidationIssue{
			Path:     issue.Path,
			Message:  issue.Message,
			Severity: issue.Severity,
		}
	}

	warnings := make([]output.ManifestValidationIssue, len(result.Warnings))
	for i, warn := range result.Warnings {
		warnings[i] = output.ManifestValidationIssue{
			Path:     warn.Path,
			Message:  warn.Message,
			Severity: warn.Severity,
		}
	}

	out := output.ManifestValidateOutput{
		Path:     result.Path,
		Schema:   result.Schema,
		Valid:    result.Valid,
		Issues:   issues,
		Warnings: warnings,
	}

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return error exit code if validation failed
	if !result.Valid {
		return errors.New(errors.CodeSchemaInvalid, "Schema validation failed")
	}

	return nil
}
