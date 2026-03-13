// Package provenance implements the ari provenance commands.
package provenance

import (
	"os"
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
		Short: "Inspect file provenance in channel directory",
		Long: `Inspect the provenance manifest to see the origin and ownership state of files in the channel directory.

The provenance manifest tracks every file Knossos places in the channel directory, recording:
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
	common.SetGroupCommand(cmd)

	return cmd
}

// newShowCmd creates the 'provenance show' subcommand.
func newShowCmd(ctx *cmdContext) *cobra.Command {
	var scopeFilter string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display provenance manifest",
		Long: `Display the provenance manifest showing origin and ownership for all files in the channel directory.

The status column shows:
  - match:    File on disk matches the expected checksum
  - diverged: File has been modified (knossos -> user ownership promotion)
  - -:        User or untracked file (no checksum validation)

Output formats:
  - text: Tabular output (default)
  - json: Full manifest as JSON
  - yaml: Full manifest as YAML

Flags:
  --scope: Filter by scope (rite, user). Default shows both scopes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(ctx, scopeFilter)
		},
	}

	cmd.Flags().StringVar(&scopeFilter, "scope", "", "Filter by scope: rite, user (default: show both)")

	common.SetNeedsProject(cmd, true, false)
	return cmd
}

// runShow implements the 'provenance show' command.
func runShow(ctx *cmdContext, scopeFilter string) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	var allEntries []*ShowEntry
	var combinedOutput CombinedOutput

	// Load rite-scope manifest (from project .knossos/)
	if scopeFilter == "" || scopeFilter == "rite" {
		knossosDir := filepath.Join(resolver.ProjectRoot(), ".knossos")
		channelDir := filepath.Join(resolver.ProjectRoot(), ".claude")
		manifestPath := provenance.ManifestPath(knossosDir)
		manifest, err := provenance.LoadOrBootstrap(manifestPath)
		if err == nil && len(manifest.Entries) > 0 {
			combinedOutput.Rite = manifest
			for path, entry := range manifest.Entries {
				allEntries = append(allEntries, makeShowEntry(
					path, entry, channelDir, *ctx.Verbose))
			}
		}
	}

	// Load user-scope manifest (from user channel dir)
	if scopeFilter == "" || scopeFilter == "user" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			userChannelDir := filepath.Join(homeDir, ".claude")
			userManifestPath := provenance.UserManifestPath(userChannelDir)
			userManifest, loadErr := provenance.LoadOrBootstrap(userManifestPath)
			if loadErr == nil && len(userManifest.Entries) > 0 {
				combinedOutput.User = userManifest
				for path, entry := range userManifest.Entries {
					displayPath := "~/" + path
					showEntry := makeShowEntryWithDisplayPath(
						displayPath, path, entry, userChannelDir, *ctx.Verbose)
					allEntries = append(allEntries, showEntry)
				}
			}
		}
	}

	// If no entries found, show helpful message
	if len(allEntries) == 0 {
		printer.PrintLine("No provenance manifest found. Run 'ari sync' first.")
		return nil
	}

	// Output based on format
	format := output.ParseFormat(*ctx.Output)
	if format == output.FormatJSON || format == output.FormatYAML {
		// For JSON/YAML, output the combined structure
		return printer.Print(&combinedOutput)
	}

	// For text, output as table
	return printer.Print(&ShowOutput{Entries: allEntries})
}

// ShowEntry represents a single provenance entry for display.
type ShowEntry struct {
	Path       string `json:"path"`
	Owner      string `json:"owner"`
	Scope      string `json:"scope"`
	SourcePath string `json:"source_path,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	Status     string `json:"status"`
	Checksum   string `json:"checksum,omitempty"`
}

// ShowOutput contains the provenance table for display.
type ShowOutput struct {
	Entries []*ShowEntry
}

// CombinedOutput contains both rite and user provenance manifests for structured output.
type CombinedOutput struct {
	Rite *provenance.ProvenanceManifest `json:"rite,omitempty" yaml:"rite,omitempty"`
	User *provenance.ProvenanceManifest `json:"user,omitempty" yaml:"user,omitempty"`
}

// Headers implements output.Tabular.
func (s *ShowOutput) Headers() []string {
	return []string{"PATH", "OWNER", "SCOPE", "SOURCE", "STATUS"}
}

// Rows implements output.Tabular.
func (s *ShowOutput) Rows() [][]string {
	rows := make([][]string, len(s.Entries))
	for i, e := range s.Entries {
		source := formatSource(e.SourcePath, e.SourceType)
		rows[i] = []string{e.Path, e.Owner, e.Scope, source, e.Status}
	}
	return rows
}

// makeShowEntry creates a ShowEntry from a ProvenanceEntry.
func makeShowEntry(path string, entry *provenance.ProvenanceEntry, channelDir string, verbose bool) *ShowEntry {
	showEntry := &ShowEntry{
		Path:       path,
		Owner:      string(entry.Owner),
		Scope:      string(entry.Scope),
		SourcePath: entry.SourcePath,
		SourceType: entry.SourceType,
		Status:     computeStatus(channelDir, path, entry),
	}

	// Add checksum if verbose
	if verbose {
		showEntry.Checksum = entry.Checksum
	}

	return showEntry
}

// makeShowEntryWithDisplayPath creates a ShowEntry with separate display and actual paths.
// Used for user-scope entries where displayPath has "~/" prefix but actualPath is used for status.
func makeShowEntryWithDisplayPath(displayPath, actualPath string, entry *provenance.ProvenanceEntry, channelDir string, verbose bool) *ShowEntry {
	showEntry := &ShowEntry{
		Path:       displayPath,
		Owner:      string(entry.Owner),
		Scope:      string(entry.Scope),
		SourcePath: entry.SourcePath,
		SourceType: entry.SourceType,
		Status:     computeStatus(channelDir, actualPath, entry),
	}

	// Add checksum if verbose
	if verbose {
		showEntry.Checksum = entry.Checksum
	}

	return showEntry
}

// computeStatus determines the status of a file based on its checksum.
func computeStatus(channelDir, path string, entry *provenance.ProvenanceEntry) string {
	// User and untracked files have no validation
	if entry.Owner != provenance.OwnerKnossos {
		return "-"
	}

	// Compute current checksum
	fullPath := filepath.Join(channelDir, path)
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
func formatSource(sourcePath, _ string) string {
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
