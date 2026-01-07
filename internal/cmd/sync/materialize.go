package sync

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/materialize"
)

func newMaterializeCmd(ctx *cmdContext) *cobra.Command {
	var force bool
	var riteName string
	var dryRun bool
	var removeAll bool
	var keepAll bool
	var promoteAll bool

	cmd := &cobra.Command{
		Use:   "materialize",
		Short: "Generate .claude/ directory from templates and rite manifests",
		Long: `Materialize generates the complete .claude/ directory structure from:
  - templates/ (hooks, CLAUDE.md sections)
  - rites/{active}/ (agents, skills)
  - rites/shared/ (shared skills)

This is an idempotent operation - it can be run multiple times safely.

By default, it uses the current ACTIVE_RITE. Use --rite to specify a different rite.

Orphan Handling:
  When switching rites, agents from the previous rite become "orphans".
  By default, orphans are preserved (--keep-all behavior).
  Use --remove-all to delete orphans (with backup) or --promote-all to move them to user-level.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate mutually exclusive flags
			exclusiveCount := 0
			if removeAll {
				exclusiveCount++
			}
			if keepAll {
				exclusiveCount++
			}
			if promoteAll {
				exclusiveCount++
			}
			if exclusiveCount > 1 {
				return fmt.Errorf("--remove-all, --keep-all, and --promote-all are mutually exclusive")
			}

			// Build options
			opts := materialize.Options{
				Force:      force,
				DryRun:     dryRun,
				RemoveAll:  removeAll,
				KeepAll:    keepAll || exclusiveCount == 0, // Default to keep
				PromoteAll: promoteAll,
			}

			return runMaterialize(ctx, riteName, opts)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force regeneration, overwriting local changes")
	cmd.Flags().StringVar(&riteName, "rite", "", "Rite to materialize (defaults to current ACTIVE_RITE)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&removeAll, "remove-all", false, "Remove all orphan agents (with backup)")
	cmd.Flags().BoolVar(&keepAll, "keep-all", false, "Preserve all orphan agents (default)")
	cmd.Flags().BoolVar(&promoteAll, "promote-all", false, "Move orphan agents to user-level (~/.claude/agents/)")

	return cmd
}

func runMaterialize(ctx *cmdContext, riteName string, opts materialize.Options) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Determine which rite to materialize
	if riteName == "" {
		// Read current ACTIVE_RITE
		activeRitePath := resolver.ActiveRiteFile()
		data, err := os.ReadFile(activeRitePath)
		if err != nil {
			if os.IsNotExist(err) {
				printer.PrintError(fmt.Errorf("no ACTIVE_RITE found, please specify --rite"))
				return err
			}
			printer.PrintError(err)
			return err
		}
		riteName = string(data)
		// Trim whitespace/newlines
		riteName = string(bytes.TrimSpace([]byte(riteName)))
	}

	if riteName == "" {
		err := fmt.Errorf("rite name cannot be empty")
		printer.PrintError(err)
		return err
	}

	// Create materializer
	m := materialize.NewMaterializer(resolver)

	// Run materialization
	printer.VerboseLog("info", fmt.Sprintf("Materializing .claude/ for rite: %s", riteName), nil)

	result, err := m.MaterializeWithOptions(riteName, opts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	output := map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("Successfully materialized .claude/ for rite '%s'", riteName),
		"rite":       riteName,
		"claude_dir": resolver.ClaudeDir(),
	}

	// Check if materialization was skipped (already on this rite)
	if result.OrphanAction == "skipped" {
		output["status"] = "skipped"
		output["message"] = fmt.Sprintf("Already on rite '%s' (use --force to re-materialize)", riteName)
		return printer.Print(output)
	}

	// Add orphan info if present
	if len(result.OrphansDetected) > 0 {
		output["orphans_detected"] = result.OrphansDetected
		output["orphan_action"] = result.OrphanAction
	}

	// Add dry-run indicator
	if opts.DryRun {
		output["dry_run"] = true
		output["message"] = fmt.Sprintf("[DRY RUN] Would materialize .claude/ for rite '%s'", riteName)
	}

	return printer.Print(output)
}
