package rite

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/output"
)

type validateOptions struct {
	riteName string
	fix      bool
}

func newValidateCmd(ctx *cmdContext) *cobra.Command {
	var opts validateOptions

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate rite integrity",
		Long:  `Validates rite (practice bundle) structure and configuration integrity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to validate (default: active)")
	cmd.Flags().StringVarP(&opts.riteName, "team", "t", "", "Deprecated: use --rite instead")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Attempt automatic repairs")

	cmd.Flags().MarkDeprecated("team", "use --rite instead")

	return cmd
}

func runValidate(ctx *cmdContext, opts validateOptions) error {
	printer := ctx.getPrinter()
	validator := ctx.getValidator()
	discovery := ctx.getDiscovery()

	// Get rite name (from flag or active)
	riteName := opts.riteName
	if riteName == "" {
		riteName = discovery.ActiveRiteName()
		if riteName == "" {
			err := errors.New(errors.CodeFileNotFound, "No active rite set. Use --rite to specify.")
			printer.PrintError(err)
			return err
		}
	}

	// Run validation
	result, err := validator.Validate(riteName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Apply fixes if requested
	if opts.fix && len(result.Fixable) > 0 {
		if err := validator.Fix(riteName); err != nil {
			printer.VerboseLog("warn", "Fix failed", map[string]interface{}{"error": err.Error()})
		}
		// Re-validate after fix
		result, _ = validator.Validate(riteName)
	}

	// Build output
	checks := make([]output.ValidationCheckOut, len(result.Checks))
	for i, c := range result.Checks {
		checks[i] = output.ValidationCheckOut{
			Check:   c.Check,
			Status:  string(c.Status),
			Message: c.Message,
		}
	}

	out := output.RiteValidateOutput{
		Rite:     result.Team,
		Valid:    result.Valid,
		Checks:   checks,
		Errors:   result.Errors,
		Warnings: result.Warnings,
		Fixable:  result.Fixable,
	}

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return error exit code if validation failed
	if !result.Valid {
		return errors.New(errors.CodeValidationFailed, "Validation failed")
	}

	return nil
}
