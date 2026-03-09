// Package sails implements the ari sails commands for White Sails quality gates.
package sails

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sails"
	"github.com/autom8y/knossos/internal/session"
)

// checkFlags holds the flags for the check command.
type checkFlags struct {
	quiet bool
}

// newCheckCmd creates the sails check command.
func newCheckCmd(ctx *cmdContext) *cobra.Command {
	flags := &checkFlags{}

	cmd := &cobra.Command{
		Use:   "check [session-path]",
		Short: "Check quality gate for a session",
		Long: `Check the quality gate for a session's WHITE_SAILS.yaml.

Returns exit code 0 for WHITE (pass), non-zero for GRAY/BLACK (fail).

The session-path can be:
  - A session directory containing WHITE_SAILS.yaml
  - A direct path to WHITE_SAILS.yaml
  - Omitted to use the current active session

Examples:
  # Check current session
  ari sails check

  # Check specific session directory
  ari sails check .sos/sessions/session-20260105-143000-abc12345

  # Check specific WHITE_SAILS.yaml file
  ari sails check path/to/WHITE_SAILS.yaml

  # Quiet mode (exit code only)
  ari sails check --quiet`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(ctx, flags, args)
		},
	}

	cmd.Flags().BoolVarP(&flags.quiet, "quiet", "q", false, "Quiet mode: only set exit code, no output")

	return cmd
}

// runCheck executes the sails check command.
func runCheck(ctx *cmdContext, flags *checkFlags, args []string) error {
	printer := ctx.getPrinter()

	var result *sails.GateResult
	var err error

	if len(args) > 0 {
		// Check specified path
		result, err = sails.CheckGate(args[0])
	} else {
		// Check current session
		projectDir := ""
		if ctx.ProjectDir != nil {
			projectDir = *ctx.ProjectDir
		}
		if projectDir == "" {
			return errors.New(errors.CodeProjectNotFound, "no project directory specified and none discovered")
		}
		// Find active session for gate check
		activeID, findErr := session.FindActiveSession(paths.NewResolver(projectDir).SessionsDir())
		if findErr != nil {
			return findErr
		}
		if activeID == "" {
			return errors.New(errors.CodeSessionNotFound, "no active session")
		}
		result, err = sails.CheckGateForSession(projectDir, activeID)
	}

	if err != nil {
		if !flags.quiet {
			printer.PrintError(err)
		}
		return errors.Handled(err)
	}

	// Output the result
	if !flags.quiet {
		_ = printer.Print(formatGateResult(result))
	}

	// Return error with appropriate exit code for gate failures
	if exitCode := sails.GateExitCode(result); exitCode != 0 {
		return errors.New(errors.CodeValidationFailed, "quality gate failed")
	}

	return nil
}

// formatGateResult formats the gate result for output.
func formatGateResult(result *sails.GateResult) any {
	// For JSON/YAML output, return the struct directly
	// For text output, this will be formatted by the printer

	return &gateOutput{
		Pass:               result.Pass,
		Color:              string(result.Color),
		SessionID:          result.SessionID,
		Reasons:            result.Reasons,
		FilePath:           result.FilePath,
		ComputedBase:       string(result.ComputedBase),
		OpenQuestions:      result.OpenQuestions,
		ContractViolations: result.ContractViolations,
		Summary:            buildSummary(result),
	}
}

// gateOutput is the structured output for the check command.
type gateOutput struct {
	Pass               bool                      `json:"pass" yaml:"pass"`
	Color              string                    `json:"color" yaml:"color"`
	SessionID          string                    `json:"session_id" yaml:"session_id"`
	Reasons            []string                  `json:"reasons" yaml:"reasons"`
	FilePath           string                    `json:"file_path" yaml:"file_path"`
	ComputedBase       string                    `json:"computed_base,omitempty" yaml:"computed_base,omitempty"`
	OpenQuestions      []string                  `json:"open_questions,omitempty" yaml:"open_questions,omitempty"`
	ContractViolations []sails.ContractViolation `json:"contract_violations,omitempty" yaml:"contract_violations,omitempty"`
	Summary            string                    `json:"summary" yaml:"summary"`
}

// Text implements output.Textable for the gate output.
func (g *gateOutput) Text() string {
	var b strings.Builder

	// Header with pass/fail indicator
	if g.Pass {
		b.WriteString("PASS: Quality gate passed\n")
	} else {
		b.WriteString("FAIL: Quality gate failed\n")
	}

	b.WriteString("\n")

	// Color information
	fmt.Fprintf(&b, "Color:        %s\n", g.Color)
	if g.ComputedBase != "" && g.ComputedBase != g.Color {
		fmt.Fprintf(&b, "Computed:     %s (before modifiers)\n", g.ComputedBase)
	}

	// Session info
	if g.SessionID != "" {
		fmt.Fprintf(&b, "Session:      %s\n", g.SessionID)
	}

	// File path
	fmt.Fprintf(&b, "File:         %s\n", g.FilePath)

	// Reasons
	if len(g.Reasons) > 0 {
		b.WriteString("\nReasons:\n")
		for _, reason := range g.Reasons {
			fmt.Fprintf(&b, "  - %s\n", reason)
		}
	}

	// Open questions (if any)
	if len(g.OpenQuestions) > 0 {
		b.WriteString("\nOpen Questions:\n")
		for _, q := range g.OpenQuestions {
			fmt.Fprintf(&b, "  - %s\n", q)
		}
	}

	// Contract violations (if any)
	if len(g.ContractViolations) > 0 {
		b.WriteString("\nClew Contract Violations:\n")
		for _, v := range g.ContractViolations {
			severityLabel := "ERROR"
			if v.Severity == "warning" {
				severityLabel = "WARN"
			}
			fmt.Fprintf(&b, "  [%s] %s: %s\n", severityLabel, v.Type, v.Description)
		}
	}

	return b.String()
}

// buildSummary creates a one-line summary for the result.
func buildSummary(result *sails.GateResult) string {
	if result.Pass {
		return "WHITE sails: high confidence, ship without QA"
	}

	switch result.Color {
	case sails.ColorGray:
		if len(result.OpenQuestions) > 0 {
			return fmt.Sprintf("GRAY sails: %d open question(s), needs QA review", len(result.OpenQuestions))
		}
		return "GRAY sails: unknown confidence, needs QA review"
	case sails.ColorBlack:
		return "BLACK sails: known failure, do not ship"
	default:
		return fmt.Sprintf("Unknown sails color: %s", result.Color)
	}
}
