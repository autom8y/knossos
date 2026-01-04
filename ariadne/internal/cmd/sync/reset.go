package sync

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/sync"
)

type resetOptions struct {
	hard    bool
	force   bool
}

func newResetCmd(ctx *cmdContext) *cobra.Command {
	var opts resetOptions

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset sync state (dangerous)",
		Long: `Resets the sync state, removing tracking information.

Without --hard, only clears state.json.
With --hard, also reverts tracked files to their remote versions.

WARNING: This operation can cause data loss. Use --force to skip confirmation.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReset(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.hard, "hard", false, "Also revert files to remote versions")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runReset(ctx *cmdContext, opts resetOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

	stateManager := sync.NewStateManager(resolver)
	historyManager := sync.NewHistoryManager(resolver)

	// Check if initialized
	if !stateManager.IsInitialized() {
		err := errors.New(errors.CodeSyncStateCorrupt, "Sync not initialized. Nothing to reset.")
		printer.PrintError(err)
		return err
	}

	// Load state for hard reset
	state, err := stateManager.Load()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Confirmation prompt unless --force
	if !opts.force {
		action := "clear sync state"
		if opts.hard {
			action = "clear sync state and revert files"
		}

		fmt.Printf("This will %s. Continue? [y/N] ", action)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			out := output.SyncResetOutput{
				Reset:   false,
				Message: "Reset cancelled",
			}
			return printer.Print(out)
		}
	}

	var filesReset []string

	// Hard reset: revert files
	if opts.hard && state != nil {
		remote, err := sync.ParseRemote(state.Remote)
		if err != nil {
			printer.PrintError(err)
			return err
		}

		fetcher := sync.NewRemoteFetcher()

		for path, tracked := range state.TrackedFiles {
			if tracked.RemoteHash != "" {
				// Fetch and restore remote version
				remoteContent, err := fetcher.FetchFile(remote, path)
				if err != nil {
					continue // Skip files that can't be fetched
				}

				localPath := resolver.ProjectRoot() + "/" + path
				if err := os.WriteFile(localPath, remoteContent, 0644); err != nil {
					continue // Skip files that can't be written
				}

				filesReset = append(filesReset, path)
			}
		}
	}

	// Clear state
	if err := stateManager.Reset(); err != nil {
		printer.PrintError(err)
		return err
	}

	// Record in history
	historyManager.RecordReset(opts.hard, filesReset)

	// Build output
	out := output.SyncResetOutput{
		Reset:        true,
		Hard:         opts.hard,
		FilesReset:   filesReset,
		StateCleared: true,
	}

	return printer.Print(out)
}
