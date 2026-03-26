// Package agent implements ari agent commands.
// dismiss.go implements the `ari agent dismiss {name}` subcommand.
package agent

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

type dismissOptions struct {
	name  string // populated from args[0]
	force bool
}

func newDismissCmd(ctx *cmdContext) *cobra.Command {
	var opts dismissOptions

	cmd := &cobra.Command{
		Use:   "dismiss <name>",
		Short: "Dismiss a summoned agent from your user-level Claude config",
		Long: `Dismisses a previously summoned agent from your user-level Claude configuration.

Only agents that were summoned via 'ari agent summon' can be dismissed.
Standing agents (pythia, moirai, metis) and manually created agents are
not affected.

Examples:
  ari agent dismiss theoros      # Dismiss the theoros agent
  ari agent roster               # See which agents are currently summoned`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.name = args[0]
			return runDismiss(ctx, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "Remove the file even if provenance is inconsistent")

	return cmd
}

func runDismiss(ctx *cmdContext, args []string, opts dismissOptions) error {
	printer := ctx.GetPrinter(output.FormatText)

	name := opts.name

	// Guard: check standing agent deny-list
	if standingAgents[name] {
		err := errors.NewWithDetails(errors.CodeValidationFailed,
			fmt.Sprintf("%q is a standing agent and cannot be dismissed", name),
			map[string]any{"agent": name, "standing_agents": []string{"pythia", "moirai", "metis"}})
		return common.PrintAndReturn(printer, err)
	}

	// Resolve user channel dir
	userChannelDir, pathErr := paths.UserChannelDir("claude")
	if pathErr != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to resolve user channel directory", pathErr)
		return common.PrintAndReturn(printer, err)
	}

	// Load user provenance manifest
	manifestPath := provenance.UserManifestPath(userChannelDir)
	manifest, loadErr := provenance.LoadOrBootstrap(manifestPath)
	if loadErr != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to load user provenance manifest", loadErr)
		return common.PrintAndReturn(printer, err)
	}

	manifestKey := "agents/" + name + ".md"

	// Check entry: must exist and be a summon-sourced agent
	entry, exists := manifest.Entries[manifestKey]
	if !exists {
		if opts.force {
			// Force mode: attempt file removal even without manifest entry
			printer.PrintLine(fmt.Sprintf("warning: no provenance entry for %q — proceeding with force", name))
		} else {
			err := errors.NewWithDetails(errors.CodeFileNotFound,
				fmt.Sprintf("agent %q was not summoned via 'ari agent summon' (no provenance entry)", name),
				map[string]any{"agent": name, "key": manifestKey})
			return common.PrintAndReturn(printer, err)
		}
	} else {
		// Verify it was summoned (source matches "summon:*" pattern)
		if !strings.HasPrefix(entry.SourcePath, "summon:") {
			if opts.force {
				printer.PrintLine(fmt.Sprintf("warning: agent %q has source %q, not a summon — proceeding with force", name, entry.SourcePath))
			} else {
				err := errors.NewWithDetails(errors.CodeValidationFailed,
					fmt.Sprintf("agent %q was not summoned via 'ari agent summon' (source: %q)", name, entry.SourcePath),
					map[string]any{
						"agent":       name,
						"source_path": entry.SourcePath,
						"hint":        "use --force to remove anyway",
					})
				return common.PrintAndReturn(printer, err)
			}
		}
	}

	// Remove the agent file
	agentFilePath := filepath.Join(userChannelDir, "agents", name+".md")
	if removeErr := os.Remove(agentFilePath); removeErr != nil {
		if os.IsNotExist(removeErr) {
			// File already gone — warn but continue to clean up provenance
			printer.PrintLine(fmt.Sprintf("warning: agent file %s not found — cleaning up provenance", agentFilePath))
		} else {
			err := errors.Wrap(errors.CodePermissionDenied,
				fmt.Sprintf("failed to remove agent file: %s", agentFilePath), removeErr)
			return common.PrintAndReturn(printer, err)
		}
	}

	// Remove provenance entry
	if exists {
		delete(manifest.Entries, manifestKey)
	}
	manifest.LastSync = time.Now().UTC()

	if saveErr := provenance.Save(manifestPath, manifest); saveErr != nil {
		// Non-fatal: file was removed; provenance is tracking metadata only
		slog.Warn("failed to save user provenance manifest after dismiss", "error", saveErr)
		printer.PrintLine(fmt.Sprintf("warning: agent removed but provenance update failed: %s", saveErr))
	}

	printer.PrintLine(fmt.Sprintf("%s dismissed. Takes effect on next CC restart.", name))
	return nil
}
