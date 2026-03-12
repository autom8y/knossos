package inscription

import (
	"fmt"
	"github.com/autom8y/knossos/internal/cmd/common"
	"strings"

	"github.com/spf13/cobra"
)

func newValidateCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate inscription manifest and context file",
		Long: `Validate the KNOSSOS_MANIFEST.yaml and context file.

This command checks:
  - Manifest schema version and required fields
  - Region definitions and ownership
  - Marker syntax in context file
  - Region consistency between manifest and file
  - Template directory presence

Examples:
  ari inscription validate          # Validate current state
  ari inscription validate --json   # JSON output for scripting`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(ctx)
		},
	}

	return cmd
}

func runValidate(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	pipeline := ctx.getPipeline()

	result, err := pipeline.Validate()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	out := ValidateOutput{
		Valid:         result.Valid,
		SchemaVersion: result.SchemaVersion,
		RegionCount:   result.RegionCount,
	}

	if len(result.Issues) > 0 {
		out.Issues = make([]ValidationIssueOutput, len(result.Issues))
		for i, issue := range result.Issues {
			out.Issues[i] = ValidationIssueOutput{
				Severity: issue.Severity,
				Region:   issue.Region,
				Message:  issue.Message,
			}
		}
	}

	return printer.Print(out)
}

// ValidateOutput represents validation result for output.
type ValidateOutput struct {
	Valid         bool                    `json:"valid"`
	SchemaVersion string                  `json:"schema_version,omitempty"`
	RegionCount   int                     `json:"region_count"`
	Issues        []ValidationIssueOutput `json:"issues,omitempty"`
}

// Text implements output.Textable for ValidateOutput.
func (v ValidateOutput) Text() string {
	var b strings.Builder

	if v.Valid {
		b.WriteString("Validation passed\n")
	} else {
		b.WriteString("Validation failed\n")
	}

	if v.SchemaVersion != "" {
		fmt.Fprintf(&b, "Schema version: %s\n", v.SchemaVersion)
	}

	fmt.Fprintf(&b, "Regions defined: %d\n", v.RegionCount)

	if len(v.Issues) > 0 {
		fmt.Fprintf(&b, "\nIssues (%d):\n", len(v.Issues))
		for _, issue := range v.Issues {
			icon := "!"
			switch issue.Severity {
			case "error":
				icon = "X"
			case "warning":
				icon = "!"
			case "info":
				icon = "i"
			}

			if issue.Region != "" {
				fmt.Fprintf(&b, "  %s [%s] %s: %s\n", icon, issue.Severity, issue.Region, issue.Message)
			} else {
				fmt.Fprintf(&b, "  %s [%s] %s\n", icon, issue.Severity, issue.Message)
			}
		}
	}

	return b.String()
}

// ValidationIssueOutput represents a validation issue for output.
type ValidationIssueOutput struct {
	Severity string `json:"severity"`
	Region   string `json:"region,omitempty"`
	Message  string `json:"message"`
}
