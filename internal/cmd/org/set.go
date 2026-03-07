package org

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

func newSetCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <org-name>",
		Short: "Set the active organization",
		Long: `Set the active org by writing to $XDG_CONFIG_HOME/knossos/active-org.

The active org is used by ari sync to resolve org-level resources.
Can also be set via the KNOSSOS_ORG environment variable.

Examples:
  ari org set autom8y
  ari org set my-team`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(ctx, args[0])
		},
	}
	return cmd
}

func runSet(ctx *cmdContext, orgName string) error {
	printer := ctx.getPrinter()

	// Validate org exists
	orgDir := paths.OrgDataDir(orgName)
	if _, err := os.Stat(orgDir); os.IsNotExist(err) {
		return errors.New(errors.CodeFileNotFound, fmt.Sprintf("org %q does not exist (run 'ari org init %s' first)", orgName, orgName))
	}

	// Write active-org file
	configDir := paths.ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to create config directory", err)
	}

	activeOrgPath := filepath.Join(configDir, "active-org")
	if err := os.WriteFile(activeOrgPath, []byte(orgName+"\n"), 0644); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to write active-org", err)
	}

	_ = printer.Print(map[string]any{
		"status":  "set",
		"org":     orgName,
		"message": fmt.Sprintf("Active org set to %q", orgName),
	})

	return nil
}
