// Package hook implements the ari hook commands.
package hook

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
)

// newCallCmd creates the call hook subcommand.
// It acts as a signing wrapper for other hook commands.
func newCallCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "call",
		Short:              "Sign and execute a hook command",
		Hidden:             true, // internal platform use only
		DisableFlagParsing: true, // forward all flags to subcommand
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCall(ctx, args)
		},
	}

	return cmd
}

func runCall(ctx *cmdContext, args []string) error {
	// 1. Capture stdin
	payload, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	// 2. Compute signature
	sig := hook.Sign(payload)

	// 3. Reconstruct command
	// The args will look like ["ari", "hook", "validate", "--output", "json"]
	// We want to insert --signature <sig>

	finalArgs := make([]string, 0, len(args)+2)
	finalArgs = append(finalArgs, args...)
	if sig != "" {
		finalArgs = append(finalArgs, "--signature", sig)
	}

	// 4. Execute sub-process
	// Use the same binary (ari)
	exe, err := os.Executable()
	if err != nil {
		exe = "ari" // fallback
	}

	// If the first arg is "ari", replace it with the actual executable path
	if len(finalArgs) > 0 && finalArgs[0] == "ari" {
		finalArgs = finalArgs[1:]
	}

	subCmd := exec.Command(exe, finalArgs...)
	subCmd.Stdin = bytes.NewReader(payload)
	subCmd.Stdout = os.Stdout
	subCmd.Stderr = os.Stderr

	// Forward exit code
	err = subCmd.Run()
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	return err
}
