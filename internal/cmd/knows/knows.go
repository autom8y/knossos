// Package knows implements the ari knows command for inspecting .know/ codebase knowledge.
package knows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/output"
)

// ValidateOutput is the structured output for the --validate flag.
type ValidateOutput struct {
	Reports    []know.ValidationReport `json:"reports"`
	AllValid   bool                    `json:"all_valid"`
	TotalRefs  int                     `json:"total_refs"`
	BrokenRefs int                     `json:"broken_refs"`
}

// Text implements output.Textable for human-readable validation output.
func (v ValidateOutput) Text() string {
	var b strings.Builder

	// Header table.
	b.WriteString(fmt.Sprintf("%-20s %-12s %-8s %s\n", "Domain", "Total Refs", "Broken", "Status"))
	b.WriteString(strings.Repeat("-", 54) + "\n")

	for _, r := range v.Reports {
		status := "valid"
		if r.BrokenCount > 0 {
			status = "BROKEN"
		}
		b.WriteString(fmt.Sprintf("%-20s %-12d %-8d %s\n", r.Domain, r.TotalRefs, r.BrokenCount, status))
	}

	// Broken reference details.
	hasBroken := false
	for _, r := range v.Reports {
		if r.BrokenCount > 0 {
			hasBroken = true
			break
		}
	}

	if hasBroken {
		b.WriteString("\nBroken References:\n")
		for _, r := range v.Reports {
			if r.BrokenCount == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("  %s:\n", r.Domain))
			for _, br := range r.Broken {
				b.WriteString(fmt.Sprintf("    [%s] %s -- %s\n", br.Type, br.Ref, br.Error))
			}
		}
	}

	return b.String()
}

// KnowsOutput is the structured output for the knows command.
type KnowsOutput struct {
	Domains  []know.DomainStatus `json:"domains"`
	AllFresh bool                `json:"all_fresh"`
}

// stalenessLabel returns the human-readable status string for a domain:
//   - "fresh"                          -- time-fresh AND code-unchanged
//   - "stale (expired)"                -- time-expired only
//   - "stale (code changed)"           -- time-fresh but source_hash differs from HEAD
//   - "stale (expired + code changed)" -- both conditions
//   - "stale"                          -- unparseable timestamp/duration
func stalenessLabel(d know.DomainStatus) string {
	if d.Fresh {
		return "fresh"
	}
	switch {
	case d.TimeExpired && d.CodeChanged:
		return "stale (expired + code changed)"
	case d.TimeExpired:
		return "stale (expired)"
	case d.CodeChanged:
		return "stale (code changed)"
	default:
		// Stale due to unparseable timestamp or duration
		return "stale"
	}
}

// Text implements output.Textable for human-readable table output.
func (k KnowsOutput) Text() string {
	if len(k.Domains) == 0 {
		return "No codebase knowledge available. Run /know to generate."
	}

	var b strings.Builder
	// Header row
	b.WriteString(fmt.Sprintf("%-16s %-20s %-12s %-30s %-10s %-10s\n",
		"Domain", "Generated", "Expires", "Status", "Source", "Confidence"))
	b.WriteString(strings.Repeat("-", 102) + "\n")

	for _, d := range k.Domains {
		status := stalenessLabel(d)
		// Trim generated timestamp for readability: "2026-02-26T21:17:58Z" -> "2026-02-26T21:17"
		generated := d.Generated
		if len(generated) >= 16 {
			generated = generated[:16]
		}
		// Truncate source hash to 7 chars for display
		srcHash := d.SourceHash
		if len(srcHash) > 7 {
			srcHash = srcHash[:7]
		}
		b.WriteString(fmt.Sprintf("%-16s %-20s %-12s %-30s %-10s %.2f\n",
			d.Domain, generated, d.Expires, status, srcHash, d.Confidence))
	}

	return b.String()
}

type cmdContext struct {
	common.BaseContext
}

// NewKnowsCmd creates the knows command.
func NewKnowsCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	var checkFlag bool
	var validateFlag bool

	cmd := &cobra.Command{
		Use:   "knows [domain]",
		Short: "Inspect .know/ codebase knowledge freshness",
		Long: `Inspect the codebase knowledge stored in .know/.

Without arguments, lists all domains with freshness status.
With a domain name, prints the full content of that domain file to stdout.

Examples:
  ari knows                    # List all domains with freshness
  ari knows architecture       # Print full .know/architecture.md content
  ari knows --check            # Exit 0 if all fresh, exit 1 if any stale
  ari knows --validate         # Validate references in all .know/ files
  ari knows --validate arch    # Validate references in a single domain
  ari knows -o json            # JSON output for scripting`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKnows(ctx, args, checkFlag, validateFlag)
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Exit 1 if any domain is stale (for CI/hooks)")
	cmd.Flags().BoolVar(&validateFlag, "validate", false, "Validate references in .know/ files against codebase")

	common.SetNeedsProject(cmd, true, true)

	return cmd
}

func runKnows(ctx *cmdContext, args []string, checkFlag, validateFlag bool) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()
	projectDir := resolver.ProjectRoot()
	knowDir := filepath.Join(projectDir, ".know")

	// --validate mode: check references in .know/ files against the codebase.
	if validateFlag {
		return runValidate(printer, projectDir, args)
	}

	// Single domain read: just cat the file to stdout
	if len(args) == 1 {
		return readSingleDomain(knowDir, args[0])
	}

	// Read all domain metadata
	domains, err := know.ReadMeta(knowDir)
	if err != nil {
		return fmt.Errorf("reading .know/ metadata: %w", err)
	}

	if len(domains) == 0 {
		printer.PrintLine("No codebase knowledge available. Run /know to generate.")
		return nil
	}

	allFresh := true
	for _, d := range domains {
		if !d.Fresh {
			allFresh = false
			break
		}
	}

	// --check mode: exit 1 if any stale
	if checkFlag {
		if allFresh {
			printer.PrintLine("OK: all domains fresh")
			return nil
		}
		// Print stale domains with reason before exiting
		for _, d := range domains {
			if !d.Fresh {
				printer.PrintLine(fmt.Sprintf("STALE: %s (%s, expires %s)", d.Domain, stalenessLabel(d), d.Expires))
			}
		}
		os.Exit(1)
		return nil // unreachable, but required for compiler
	}

	result := KnowsOutput{
		Domains:  domains,
		AllFresh: allFresh,
	}
	return printer.Print(result)
}

// runValidate executes the --validate flag logic.
// With a domain arg: validates that single domain and prints its report.
// Without a domain arg: validates all domains and prints a summary table.
// Exits with code 1 if any broken references are found.
func runValidate(printer interface {
	Print(data any) error
	PrintLine(text string)
}, projectDir string, args []string) error {
	var validateOut ValidateOutput

	if len(args) == 1 {
		// Single domain validation.
		report, err := know.ValidateDomain(projectDir, args[0])
		if err != nil {
			return fmt.Errorf("validating domain %q: %w", args[0], err)
		}
		validateOut.Reports = []know.ValidationReport{*report}
		validateOut.TotalRefs = report.TotalRefs
		validateOut.BrokenRefs = report.BrokenCount
		validateOut.AllValid = report.BrokenCount == 0
	} else {
		// All domains validation.
		reports, err := know.ValidateAll(projectDir)
		if err != nil {
			return fmt.Errorf("validating .know/: %w", err)
		}
		validateOut.Reports = reports
		for _, r := range reports {
			validateOut.TotalRefs += r.TotalRefs
			validateOut.BrokenRefs += r.BrokenCount
		}
		validateOut.AllValid = validateOut.BrokenRefs == 0
	}

	if err := printer.Print(validateOut); err != nil {
		return err
	}

	// Exit 1 if any broken references found, matching the --check pattern.
	if validateOut.BrokenRefs > 0 {
		os.Exit(1)
	}

	return nil
}

// readSingleDomain reads and prints a single domain file to stdout.
func readSingleDomain(knowDir, domain string) error {
	path := filepath.Join(knowDir, domain+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("domain %q not found in .know/ (expected %s)", domain, path)
		}
		return fmt.Errorf("reading .know/%s.md: %w", domain, err)
	}
	_, err = os.Stdout.Write(data)
	return err
}
