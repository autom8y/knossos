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

// KnowsOutput is the structured output for the knows command.
type KnowsOutput struct {
	Domains  []know.DomainStatus `json:"domains"`
	AllFresh bool                `json:"all_fresh"`
}

// Text implements output.Textable for human-readable table output.
func (k KnowsOutput) Text() string {
	if len(k.Domains) == 0 {
		return "No codebase knowledge available. Run /know to generate."
	}

	var b strings.Builder
	// Header row
	b.WriteString(fmt.Sprintf("%-16s %-20s %-12s %-10s %-10s %-10s\n",
		"Domain", "Generated", "Expires", "Status", "Source", "Confidence"))
	b.WriteString(strings.Repeat("-", 82) + "\n")

	for _, d := range k.Domains {
		status := "fresh"
		if !d.Fresh {
			status = "STALE"
		}
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
		b.WriteString(fmt.Sprintf("%-16s %-20s %-12s %-10s %-10s %.2f\n",
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
  ari knows -o json            # JSON output for scripting`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKnows(ctx, args, checkFlag)
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Exit 1 if any domain is stale (for CI/hooks)")

	common.SetNeedsProject(cmd, true, true)

	return cmd
}

func runKnows(ctx *cmdContext, args []string, checkFlag bool) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()
	projectDir := resolver.ProjectRoot()
	knowDir := filepath.Join(projectDir, ".know")

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
		// Print stale domains before exiting
		for _, d := range domains {
			if !d.Fresh {
				printer.PrintLine(fmt.Sprintf("STALE: %s (expired %s)", d.Domain, d.Expires))
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
