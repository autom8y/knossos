package org

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

func newListCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available organizations",
		Long: `Discover all organizations at the XDG data path.

Lists all directories under $XDG_DATA_HOME/knossos/orgs/.
Marks the active org with an asterisk.

Examples:
  ari org list`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx)
		},
	}
	return cmd
}

func runList(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	orgsDir := filepath.Join(paths.DataDir(), "orgs")
	entries, err := os.ReadDir(orgsDir)
	if err != nil {
		if os.IsNotExist(err) {
			printer.Print(map[string]interface{}{
				"status": "empty",
				"orgs":   []string{},
			})
			return nil
		}
		return errors.Wrap(errors.CodeFileNotFound, "failed to read orgs directory", err)
	}

	activeOrg := config.ActiveOrg()
	var orgs []map[string]interface{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		org := map[string]interface{}{
			"name": entry.Name(),
		}
		if entry.Name() == activeOrg {
			org["active"] = true
		}
		orgs = append(orgs, org)
	}

	printer.Print(map[string]interface{}{
		"status": "ok",
		"orgs":   orgs,
		"count":  len(orgs),
	})

	return nil
}
