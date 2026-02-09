// Package provenance implements the ari provenance commands.
package provenance

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/provenance"
)

// cmdContext holds shared state for provenance commands.
type cmdContext struct {
	common.BaseContext
}

// NewProvenanceCmd creates the provenance command group.
func NewProvenanceCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "provenance",
		Short: "Inspect file provenance in .claude/",
		Long: `Inspect the provenance manifest to see the origin and ownership state of files in .claude/.

The provenance manifest tracks every file Knossos places in .claude/, recording:
  - Owner: who owns the file (knossos, user, unknown)
  - Source: where the file came from (rite path, template path, mena path)
  - Status: whether the file matches the expected checksum (match, diverged)

Examples:
  ari provenance show              # Display provenance table
  ari provenance show -o json      # JSON output for tooling
  ari provenance show --verbose    # Show full checksums`,
	}

	// Add subcommands
	cmd.AddCommand(newShowCmd(ctx))

	// Provenance commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// newShowCmd creates the 'provenance show' subcommand.
func newShowCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display provenance manifest",
		Long: `Display the provenance manifest showing origin and ownership for all files in .claude/.

The status column shows:
  - match:    File on disk matches the expected checksum
  - diverged: File has been modified (knossos -> user ownership promotion)
  - -:        User or unknown file (no checksum validation)

Output formats:
  - text: Tabular output (default)
  - json: Full manifest as JSON
  - yaml: Full manifest as YAML`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(ctx)
		},
	}

	common.SetNeedsProject(cmd, true, false)
	return cmd
}

// runShow implements the 'provenance show' command.
func runShow(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	claudeDir := filepath.Join(resolver.ProjectRoot(), ".claude")
	manifestPath := provenance.ManifestPath(claudeDir)

	// Load or bootstrap manifest
	manifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// If manifest is empty (bootstrap case), show helpful message
	if len(manifest.Entries) == 0 {
		printer.PrintLine("No provenance manifest found. Run 'ari sync materialize' first.")
		return nil
	}

	// Compute status for each entry
	entries := make([]*ShowEntry, 0, len(manifest.Entries))
	for path, entry := range manifest.Entries {
		showEntry := &ShowEntry{
			Path:       path,
			Owner:      string(entry.Owner),
			SourcePath: entry.SourcePath,
			SourceType: entry.SourceType,
			Status:     computeStatus(claudeDir, path, entry, *ctx.Verbose),
		}

		// Add checksum if verbose
		if *ctx.Verbose {
			showEntry.Checksum = entry.Checksum
		}

		entries = append(entries, showEntry)
	}

	// Output based on format
	format := output.ParseFormat(*ctx.Output)
	if format == output.FormatJSON || format == output.FormatYAML {
		// For JSON/YAML, output the raw manifest
		return printer.Print(manifest)
	}

	// For text, output as table
	return printer.Print(&ShowOutput{Entries: entries})
}

// ShowEntry represents a single provenance entry for display.
type ShowEntry struct {
	Path       string `json:"path"`
	Owner      string `json:"owner"`
	SourcePath string `json:"source_path,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	Status     string `json:"status"`
	Checksum   string `json:"checksum,omitempty"`
}

// ShowOutput contains the provenance table for display.
type ShowOutput struct {
	Entries []*ShowEntry
}

// Headers implements output.Tabular.
func (s *ShowOutput) Headers() []string {
	return []string{"PATH", "OWNER", "SOURCE", "STATUS"}
}

// Rows implements output.Tabular.
func (s *ShowOutput) Rows() [][]string {
	rows := make([][]string, len(s.Entries))
	for i, e := range s.Entries {
		source := formatSource(e.SourcePath, e.SourceType)
		rows[i] = []string{e.Path, e.Owner, source, e.Status}
	}
	return rows
}

// computeStatus determines the status of a file based on its checksum.
func computeStatus(claudeDir, path string, entry *provenance.ProvenanceEntry, verbose bool) string {
	// User and unknown files have no validation
	if entry.Owner != provenance.OwnerKnossos {
		return "-"
	}

	// Compute current checksum
	fullPath := filepath.Join(claudeDir, path)
	var currentChecksum string
	var err error

	if strings.HasSuffix(path, "/") {
		// Directory-level entry (mena)
		dirPath := strings.TrimSuffix(fullPath, "/")
		currentChecksum, err = checksum.Dir(dirPath)
	} else {
		// File-level entry
		currentChecksum, err = checksum.File(fullPath)
	}

	if err != nil || currentChecksum == "" {
		return "missing"
	}

	if currentChecksum == entry.Checksum {
		return "match"
	}

	return "diverged"
}

// formatSource formats the source information for display.
func formatSource(sourcePath, sourceType string) string {
	if sourcePath == "" {
		return "(user-created)"
	}

	// For concise display, show source path with truncation if needed
	if len(sourcePath) > 50 {
		return sourcePath[:47] + "..."
	}

	return sourcePath
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}
