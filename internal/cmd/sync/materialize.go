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

	cmd := &cobra.Command{
		Use:   "materialize",
		Short: "Generate .claude/ directory from templates and rite manifests",
		Long: `Materialize generates the complete .claude/ directory structure from:
  - templates/ (hooks, CLAUDE.md sections)
  - rites/{active}/ (agents, skills)
  - rites/shared/ (shared skills)

This is an idempotent operation - it can be run multiple times safely.

By default, it uses the current ACTIVE_RITE. Use --rite to specify a different rite.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMaterialize(ctx, riteName, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force regeneration, overwriting local changes")
	cmd.Flags().StringVar(&riteName, "rite", "", "Rite to materialize (defaults to current ACTIVE_RITE)")

	return cmd
}

func runMaterialize(ctx *cmdContext, riteName string, force bool) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

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

	if err := m.Materialize(riteName); err != nil {
		printer.PrintError(err)
		return err
	}

	// Success output
	result := map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("Successfully materialized .claude/ for rite '%s'", riteName),
		"rite":       riteName,
		"claude_dir": resolver.ClaudeDir(),
	}

	return printer.Print(result)
}
