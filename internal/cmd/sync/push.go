package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/sync"
)

type pushOptions struct {
	force  bool
	dryRun bool
	paths  []string
}

func newPushCmd(ctx *cmdContext) *cobra.Command {
	var opts pushOptions

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push local changes to remote",
		Long: `Pushes local configuration changes to the remote.

Currently only supports local filesystem remotes.

Pre-push checks:
  - Verifies no unresolved conflicts
  - Checks remote hasn't changed since last pull
  - Use --force to override safety checks`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(ctx, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force push even with conflicts or remote changes")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().StringSliceVar(&opts.paths, "path", nil, "Specific paths to push (can be repeated)")

	return cmd
}

func runPush(ctx *cmdContext, opts pushOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Create history manager
	historyManager := sync.NewHistoryManager(resolver)

	// Perform push
	pusher := sync.NewPusher(resolver)
	pushOpts := sync.PushOptions{
		Force:  opts.force,
		DryRun: opts.dryRun,
		Paths:  opts.paths,
	}

	result, err := pusher.Push(pushOpts)
	if err != nil {
		printer.PrintError(err)
		historyManager.RecordPush("", nil, false, err.Error())
		return err
	}

	// Build output
	out := output.SyncPushOutput{
		Remote:       result.Remote,
		Success:      result.Success,
		FilesPushed:  make([]output.SyncFileChange, len(result.FilesPushed)),
		PushedCount:  result.PushedCount,
		Rejected:     result.Rejected,
		RejectReason: result.RejectReason,
	}

	for i, f := range result.FilesPushed {
		out.FilesPushed[i] = output.SyncFileChange{
			Path:    f.Path,
			Action:  f.Action,
			OldHash: shortenHash(f.OldHash),
			NewHash: shortenHash(f.NewHash),
		}
	}

	// Record in history
	fileNames := make([]string, len(result.FilesPushed))
	for i, f := range result.FilesPushed {
		fileNames[i] = f.Path
	}
	errMsg := ""
	if result.Rejected {
		errMsg = result.RejectReason
	}
	historyManager.RecordPush(result.Remote, fileNames, result.Success, errMsg)

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return error if rejected
	if result.Rejected {
		return errors.ErrRemoteRejected(result.Remote, result.RejectReason)
	}

	return nil
}
