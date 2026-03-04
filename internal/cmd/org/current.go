package org

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/config"
)

func newCurrentCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the active organization",
		Long: `Display the currently active organization.

Resolution order:
  1. $KNOSSOS_ORG environment variable
  2. $XDG_CONFIG_HOME/knossos/active-org file

Examples:
  ari org current`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCurrent(ctx)
		},
	}
	return cmd
}

func runCurrent(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	activeOrg := config.ActiveOrg()
	if activeOrg == "" {
		_ = printer.Print(map[string]interface{}{
			"status":  "none",
			"message": "No active org (set with 'ari org set <name>' or KNOSSOS_ORG env var)",
		})
		return nil
	}

	_ = printer.Print(map[string]interface{}{
		"status": "ok",
		"org":    activeOrg,
	})

	return nil
}
