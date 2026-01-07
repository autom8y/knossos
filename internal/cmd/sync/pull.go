package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/sync"
)

type pullOptions struct {
	force  bool
	dryRun bool
	paths  []string
}

func newPullCmd(ctx *cmdContext) *cobra.Command {
	var opts pullOptions

	cmd := &cobra.Command{
		Use:   "pull [remote]",
		Short: "Pull remote changes with conflict detection",
		Long: `Pulls changes from the remote source with three-way conflict detection.

If remote is not specified, uses the previously configured remote.

Remote formats:
  - Local path: /path/to/source or ./relative
  - HTTP(S) URL: https://example.com/config
  - GitHub: org/repo (uses raw.githubusercontent.com)
  - Git ref: HEAD:.claude/path

Conflict handling:
  - Detects when both local and remote have changed
  - Files with conflicts are marked but not overwritten
  - Use 'ari sync resolve' to handle conflicts`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			remote := ""
			if len(args) > 0 {
				remote = args[0]
			}
			return runPull(ctx, remote, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force overwrite even with conflicts")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().StringSliceVar(&opts.paths, "path", nil, "Specific paths to pull (can be repeated)")

	return cmd
}

func runPull(ctx *cmdContext, remoteArg string, opts pullOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Load existing state to get remote if not specified
	stateManager := sync.NewStateManager(resolver)
	state, err := stateManager.Load()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	remote := remoteArg
	if remote == "" {
		if state == nil {
			err := errors.New(errors.CodeUsageError, "No remote specified and sync not initialized. Usage: ari sync pull <remote>")
			printer.PrintError(err)
			return err
		}
		remote = state.Remote
	}

	// Create history manager
	historyManager := sync.NewHistoryManager(resolver)

	// Perform pull
	puller := sync.NewPuller(resolver)
	pullOpts := sync.PullOptions{
		Force:  opts.force,
		DryRun: opts.dryRun,
		Paths:  opts.paths,
	}

	result, err := puller.Pull(remote, pullOpts)
	if err != nil {
		printer.PrintError(err)
		// Record failed pull
		historyManager.RecordPull(remote, nil, false, err.Error())
		return err
	}

	// Build output
	out := output.SyncPullOutput{
		Remote:        result.Remote,
		Success:       result.Success,
		FilesUpdated:  make([]output.SyncFileChange, len(result.FilesUpdated)),
		FilesConflict: make([]output.SyncConflictEntry, len(result.FilesConflict)),
		HasConflicts:  result.ConflictCount > 0,
		UpdatedCount:  result.UpdatedCount,
		ConflictCount: result.ConflictCount,
	}

	for i, f := range result.FilesUpdated {
		out.FilesUpdated[i] = output.SyncFileChange{
			Path:    f.Path,
			Action:  f.Action,
			OldHash: shortenHash(f.OldHash),
			NewHash: shortenHash(f.NewHash),
		}
	}

	for i, c := range result.FilesConflict {
		out.FilesConflict[i] = output.SyncConflictEntry{
			Path:        c.Path,
			Description: c.Description,
			LocalHash:   shortenHash(c.LocalHash),
			RemoteHash:  shortenHash(c.RemoteHash),
			BaseHash:    shortenHash(c.BaseHash),
		}
	}

	// Record in history
	fileNames := make([]string, len(result.FilesUpdated))
	for i, f := range result.FilesUpdated {
		fileNames[i] = f.Path
	}
	historyManager.RecordPull(remote, fileNames, result.Success, "")

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return error if conflicts exist
	if result.ConflictCount > 0 {
		conflictPaths := make([]string, len(result.FilesConflict))
		for i, c := range result.FilesConflict {
			conflictPaths[i] = c.Path
		}
		return errors.ErrSyncConflict(conflictPaths)
	}

	return nil
}
