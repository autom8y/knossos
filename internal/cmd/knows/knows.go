// Package knows implements the ari knows command for inspecting .know/ codebase knowledge.
package knows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/output"
)

// DeltaOutput is the structured output for the --delta flag.
type DeltaOutput struct {
	Domain    string               `json:"domain"`
	Manifest  *know.ChangeManifest `json:"manifest"`
	Mode      string               `json:"mode"`
	ForceFull bool                 `json:"force_full"`
}

// Text implements output.Textable for human-readable delta output.
func (d DeltaOutput) Text() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Domain:     %s\n", d.Domain))
	b.WriteString(fmt.Sprintf("Mode:       %s\n", d.Mode))
	b.WriteString(fmt.Sprintf("ForceFull:  %v\n", d.ForceFull))

	if d.Manifest == nil {
		b.WriteString("Manifest:   (none)\n")
		return b.String()
	}

	m := d.Manifest
	changedCount := len(m.NewFiles) + len(m.ModifiedFiles) + len(m.DeletedFiles) + len(m.RenamedFiles)
	b.WriteString(fmt.Sprintf("Manifest:   %s..%s\n", m.FromHash, m.ToHash))
	b.WriteString(fmt.Sprintf("  New:      %d\n", len(m.NewFiles)))
	b.WriteString(fmt.Sprintf("  Modified: %d\n", len(m.ModifiedFiles)))
	b.WriteString(fmt.Sprintf("  Deleted:  %d\n", len(m.DeletedFiles)))
	b.WriteString(fmt.Sprintf("  Renamed:  %d\n", len(m.RenamedFiles)))
	b.WriteString(fmt.Sprintf("  Total:    %d files changed\n", changedCount))
	b.WriteString(fmt.Sprintf("  Delta:    %d lines (ratio %.2f)\n", m.DeltaLines, m.DeltaRatio))

	if m.CommitLog != "" {
		b.WriteString("CommitLog:\n")
		lines := strings.Split(m.CommitLog, "\n")
		if len(lines) > 10 {
			lines = lines[:10]
		}
		for _, line := range lines {
			b.WriteString(fmt.Sprintf("  %s\n", line))
		}
		if logLines := strings.Count(m.CommitLog, "\n") + 1; logLines > 10 {
			b.WriteString(fmt.Sprintf("  ... (%d more commits)\n", logLines-10))
		}
	}

	if len(m.NewFiles) > 0 {
		b.WriteString("New files:\n")
		for _, f := range m.NewFiles {
			b.WriteString(fmt.Sprintf("  + %s\n", f))
		}
	}
	if len(m.ModifiedFiles) > 0 {
		b.WriteString("Modified files:\n")
		for _, f := range m.ModifiedFiles {
			b.WriteString(fmt.Sprintf("  ~ %s\n", f))
		}
	}
	if len(m.DeletedFiles) > 0 {
		b.WriteString("Deleted files:\n")
		for _, f := range m.DeletedFiles {
			b.WriteString(fmt.Sprintf("  - %s\n", f))
		}
	}
	if len(m.RenamedFiles) > 0 {
		b.WriteString("Renamed files:\n")
		for _, rf := range m.RenamedFiles {
			b.WriteString(fmt.Sprintf("  %s -> %s\n", rf.OldPath, rf.NewPath))
		}
	}

	return b.String()
}

// DeltaAllOutput wraps multiple DeltaOutput results.
type DeltaAllOutput struct {
	Domains []DeltaOutput `json:"domains"`
}

// Text implements output.Textable for human-readable multi-domain delta summary.
func (da DeltaAllOutput) Text() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%-20s %-14s %-10s %-14s %s\n",
		"Domain", "Mode", "ForceFull", "Files Changed", "Delta Lines"))
	b.WriteString(strings.Repeat("-", 70) + "\n")

	for _, d := range da.Domains {
		filesChanged := 0
		deltaLines := 0
		if d.Manifest != nil {
			filesChanged = len(d.Manifest.NewFiles) + len(d.Manifest.ModifiedFiles) +
				len(d.Manifest.DeletedFiles) + len(d.Manifest.RenamedFiles)
			deltaLines = d.Manifest.DeltaLines
		}
		b.WriteString(fmt.Sprintf("%-20s %-14s %-10v %-14d %d\n",
			d.Domain, d.Mode, d.ForceFull, filesChanged, deltaLines))
	}

	return b.String()
}

// SemanticDiffOutput is the structured output for the --semantic-diff flag.
type SemanticDiffOutput struct {
	Domain string             `json:"domain"`
	Diff   *know.SemanticDiff `json:"diff"`
}

// Text implements output.Textable for human-readable semantic diff output.
func (s SemanticDiffOutput) Text() string {
	if s.Diff == nil {
		return fmt.Sprintf("Domain: %s\nNo changes detected.\n", s.Domain)
	}
	return fmt.Sprintf("Domain: %s\n\n%s", s.Domain, know.FormatSemanticDiff(s.Diff))
}

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
	var deltaFlag bool
	var semanticDiffFlag bool
	var scopeDir string

	cmd := &cobra.Command{
		Use:   "knows [domain]",
		Short: "Inspect .know/ codebase knowledge freshness",
		Long: `Inspect the codebase knowledge stored in .know/.

Without arguments, lists all domains with freshness status.
With a domain name, prints the full content of that domain file to stdout.

Examples:
  ari knows                                   # List all domains with freshness
  ari knows architecture                      # Print full .know/architecture.md content
  ari knows --check                           # Exit 0 if all fresh, exit 1 if any stale
  ari knows --validate                        # Validate references in all .know/ files
  ari knows --validate arch                   # Validate references in a single domain
  ari knows --delta                           # Show change manifests for all domains
  ari knows --delta architecture              # Show change manifest for one domain
  ari knows --semantic-diff arch              # AST-based semantic diff for Go files
  ari knows --scope-dir services/payments/    # Hierarchical view from service dir
  ari knows -o json                           # JSON output for scripting`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKnows(ctx, args, scopeDir, checkFlag, validateFlag, deltaFlag, semanticDiffFlag)
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Exit 1 if any domain is stale (for CI/hooks)")
	cmd.Flags().BoolVar(&validateFlag, "validate", false, "Validate references in .know/ files against codebase")
	cmd.Flags().BoolVar(&deltaFlag, "delta", false, "Show change manifest and recommended update mode")
	cmd.Flags().BoolVar(&semanticDiffFlag, "semantic-diff", false, "Show AST-based semantic diff for Go files (compressed change context)")
	cmd.Flags().StringVar(&scopeDir, "scope-dir", "", "Starting directory for hierarchical .know/ discovery (defaults to project root)")

	common.SetNeedsProject(cmd, true, true)

	return cmd
}

func runKnows(ctx *cmdContext, args []string, scopeDir string, checkFlag, validateFlag, deltaFlag, semanticDiffFlag bool) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()
	projectDir := resolver.ProjectRoot()

	// Resolve scope directory for hierarchical discovery
	startDir := projectDir
	if scopeDir != "" {
		startDir = scopeDir
	}
	knowDir := filepath.Join(startDir, ".know")

	// --validate mode: check references in .know/ files against the codebase.
	if validateFlag {
		return runValidate(printer, projectDir, args)
	}

	// --semantic-diff mode: AST-based semantic diff for Go files.
	if semanticDiffFlag {
		return runSemanticDiff(printer, startDir, projectDir, args)
	}

	// --delta mode: show change manifests and recommended update modes.
	if deltaFlag {
		return runDelta(printer, startDir, projectDir, args)
	}

	// Single domain read: just cat the file to stdout
	if len(args) == 1 {
		return readSingleDomain(knowDir, args[0])
	}

	// Read all domain metadata
	domains, err := know.ReadMeta(startDir, projectDir)
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, "reading .know/ metadata", err)
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
			return errors.Wrap(errors.CodeValidationFailed, fmt.Sprintf("validating domain %q", args[0]), err)
		}
		validateOut.Reports = []know.ValidationReport{*report}
		validateOut.TotalRefs = report.TotalRefs
		validateOut.BrokenRefs = report.BrokenCount
		validateOut.AllValid = report.BrokenCount == 0
	} else {
		// All domains validation.
		reports, err := know.ValidateAll(projectDir)
		if err != nil {
			return errors.Wrap(errors.CodeValidationFailed, "validating .know/", err)
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

// runDelta executes the --delta flag logic.
// With a domain arg: computes the change manifest for that single domain.
// Without a domain arg: computes manifests for all domains and prints a summary table.
func runDelta(printer interface {
	Print(data any) error
	PrintLine(text string)
}, startDir, projectDir string, args []string) error {
	knowDir := filepath.Join(startDir, ".know")
	// Read all domain statuses first to get the list of known domains.
	allStatuses, err := know.ReadMeta(startDir, projectDir)
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, "reading .know/ metadata", err)
	}

	if len(allStatuses) == 0 {
		printer.PrintLine("No codebase knowledge available. Run /know to generate.")
		return nil
	}

	// Filter to a single domain when one is specified.
	statuses := allStatuses
	if len(args) == 1 {
		var found []know.DomainStatus
		for _, s := range allStatuses {
			if s.Domain == args[0] {
				found = append(found, s)
				break
			}
		}
		if len(found) == 0 {
			return errors.New(errors.CodeFileNotFound, fmt.Sprintf("domain %q not found in .know/", args[0]))
		}
		statuses = found
	}

	var results []DeltaOutput

	// Cache unscoped manifests by (fromHash, toHash) to avoid redundant git
	// operations when multiple domains share the same hash pair.
	manifestCache := make(map[string]*know.ChangeManifest)

	for _, status := range statuses {
		out := DeltaOutput{
			Domain:    status.Domain,
			ForceFull: status.ForceFull,
		}

		// When either hash is empty we cannot compute a diff — force full.
		if status.SourceHash == "" || status.CurrentHash == "" {
			out.Mode = "full"
			results = append(results, out)
			continue
		}

		// When hashes match, there are no code changes — check for time expiry only.
		if status.SourceHash == status.CurrentHash {
			if status.TimeExpired {
				out.Mode = "time-only"
			} else {
				out.Mode = "skip"
			}
			results = append(results, out)
			continue
		}

		// Hashes differ: compute the change manifest.
		// Read full Meta to get SourceScope and incremental cycle info.
		meta, err := know.ReadSingleMeta(knowDir, status.Domain)
		if err != nil {
			// Graceful degradation: report full mode if we cannot read the meta.
			fmt.Fprintf(os.Stderr, "warn: cannot read meta for %q: %v\n", status.Domain, err)
			out.Mode = "full"
			results = append(results, out)
			continue
		}

		// Get or compute the unscoped manifest for this hash pair.
		cacheKey := status.SourceHash + ":" + status.CurrentHash
		raw, ok := manifestCache[cacheKey]
		if !ok {
			raw, err = know.ComputeChangeManifest(status.SourceHash, status.CurrentHash, nil)
			if err != nil {
				// Graceful degradation: git may be unavailable.
				fmt.Fprintf(os.Stderr, "warn: cannot compute manifest for %q: %v\n", status.Domain, err)
				out.Mode = "full"
				results = append(results, out)
				continue
			}
			manifestCache[cacheKey] = raw
		}

		// Filter the cached manifest by this domain's source scope.
		manifest := know.FilterChangeManifest(raw, meta.SourceScope)

		out.Manifest = manifest
		out.Mode = know.RecommendedMode(manifest, meta)
		results = append(results, out)
	}

	// Single domain: print detailed DeltaOutput.
	if len(args) == 1 {
		return printer.Print(results[0])
	}

	// All domains: print summary table.
	return printer.Print(DeltaAllOutput{Domains: results})
}

// readSingleDomain reads and prints a single domain file to stdout.
// domain may be a plain name ("architecture") or a feat namespace ("feat/materialization").
// The feat namespace resolves to .know/feat/{slug}.md.
func readSingleDomain(knowDir, domain string) error {
	path := know.DomainFilePath(knowDir, domain)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New(errors.CodeFileNotFound, fmt.Sprintf("domain %q not found in .know/ (expected %s)", domain, path))
		}
		return errors.Wrap(errors.CodeFileNotFound, fmt.Sprintf("reading .know/%s.md", domain), err)
	}
	_, err = os.Stdout.Write(data)
	return err
}

// runSemanticDiff computes AST-based semantic diffs for a domain's changed Go files.
func runSemanticDiff(printer interface {
	Print(data any) error
	PrintLine(text string)
}, startDir, projectDir string, args []string) error {
	knowDir := filepath.Join(startDir, ".know")
	allStatuses, err := know.ReadMeta(startDir, projectDir)
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, "reading .know/ metadata", err)
	}

	if len(allStatuses) == 0 {
		printer.PrintLine("No codebase knowledge available. Run /know to generate.")
		return nil
	}

	statuses := allStatuses
	if len(args) == 1 {
		var found []know.DomainStatus
		for _, s := range allStatuses {
			if s.Domain == args[0] {
				found = append(found, s)
				break
			}
		}
		if len(found) == 0 {
			return errors.New(errors.CodeFileNotFound, fmt.Sprintf("domain %q not found in .know/", args[0]))
		}
		statuses = found
	}

	// Cache unscoped manifests by (fromHash, toHash) to avoid redundant git
	// operations when multiple domains share the same hash pair.
	manifestCache := make(map[string]*know.ChangeManifest)

	for _, status := range statuses {
		if status.SourceHash == "" || status.CurrentHash == "" {
			continue
		}
		if status.SourceHash == status.CurrentHash {
			continue
		}

		meta, err := know.ReadSingleMeta(knowDir, status.Domain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: cannot read meta for %q: %v\n", status.Domain, err)
			continue
		}

		cacheKey := status.SourceHash + ":" + status.CurrentHash
		raw, ok := manifestCache[cacheKey]
		if !ok {
			raw, err = know.ComputeChangeManifest(status.SourceHash, status.CurrentHash, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warn: cannot compute manifest for %q: %v\n", status.Domain, err)
				continue
			}
			manifestCache[cacheKey] = raw
		}

		manifest := know.FilterChangeManifest(raw, meta.SourceScope)

		sd := &know.SemanticDiff{
			FromHash: status.SourceHash,
			ToHash:   status.CurrentHash,
		}

		for _, f := range manifest.ModifiedFiles {
			if !strings.HasSuffix(f, ".go") {
				sd.NonGoFiles = append(sd.NonGoFiles, f)
				continue
			}

			oldSource, err := know.GitShowFileFunc()(status.SourceHash, f)
			if err != nil {
				sd.SkippedFiles = append(sd.SkippedFiles, f)
				continue
			}

			newSource, err := os.ReadFile(f)
			if err != nil {
				sd.SkippedFiles = append(sd.SkippedFiles, f)
				continue
			}

			fileDiff, err := know.ComputeFileDiff(oldSource, newSource, f)
			if err != nil {
				sd.SkippedFiles = append(sd.SkippedFiles, f)
				continue
			}

			if fileDiff != nil {
				sd.Files = append(sd.Files, *fileDiff)
			}
		}

		for _, f := range manifest.NewFiles {
			if !strings.HasSuffix(f, ".go") {
				sd.NonGoFiles = append(sd.NonGoFiles, f)
			}
		}

		for _, f := range manifest.DeletedFiles {
			if strings.HasSuffix(f, ".go") {
				oldSource, err := know.GitShowFileFunc()(status.SourceHash, f)
				if err != nil {
					sd.SkippedFiles = append(sd.SkippedFiles, f)
					continue
				}
				fileDiff, err := know.ComputeFileDiff(oldSource, nil, f)
				if err != nil {
					sd.SkippedFiles = append(sd.SkippedFiles, f)
					continue
				}
				if fileDiff != nil {
					sd.Files = append(sd.Files, *fileDiff)
				}
			} else {
				sd.NonGoFiles = append(sd.NonGoFiles, f)
			}
		}

		out := SemanticDiffOutput{
			Domain: status.Domain,
			Diff:   sd,
		}

		if len(sd.Files) == 0 && len(sd.NonGoFiles) == 0 && len(sd.SkippedFiles) == 0 {
			out.Diff = nil
		}

		if err := printer.Print(out); err != nil {
			return err
		}
	}

	return nil
}
