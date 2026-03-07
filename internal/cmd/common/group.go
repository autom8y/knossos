package common

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
)

// SetGroupCommand configures a command as a group (parent-only) command.
// When invoked with arguments (unknown subcommand), it returns an error.
// When invoked without arguments, it shows help.
func SetGroupCommand(cmd *cobra.Command) {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New(errors.CodeUsageError,
				fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath()))
		}
		return cmd.Help()
	}
}
