package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/sync"
)

type historyOptions struct {
	limit     int
	operation string
	since     string
}

func newHistoryCmd(ctx *cmdContext) *cobra.Command {
	var opts historyOptions

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show sync history/audit log",
		Long: `Shows the history of sync operations.

Displays:
  - Timestamp of each operation
  - Operation type (pull, push, resolve, reset)
  - Files affected
  - Success/failure status`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(ctx, opts)
		},
	}

	cmd.Flags().IntVarP(&opts.limit, "limit", "n", 20, "Maximum entries to show")
	cmd.Flags().StringVar(&opts.operation, "operation", "", "Filter by operation type")
	cmd.Flags().StringVar(&opts.since, "since", "", "Show entries since timestamp (RFC3339)")

	return cmd
}

func runHistory(ctx *cmdContext, opts historyOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	historyManager := sync.NewHistoryManager(resolver)
	listOpts := sync.ListOptions{
		Limit:     opts.limit,
		Operation: opts.operation,
		Since:     opts.since,
	}

	entries, err := historyManager.List(listOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	totalCount, _ := historyManager.Count()

	// Build output
	out := output.SyncHistoryOutput{
		Entries: make([]output.SyncHistoryEntry, len(entries)),
		Total:   totalCount,
		Limit:   opts.limit,
	}

	for i, e := range entries {
		out.Entries[i] = output.SyncHistoryEntry{
			Timestamp: e.Timestamp,
			Operation: e.Operation,
			Remote:    e.Remote,
			Files:     e.Files,
			FileCount: e.FileCount,
			Success:   e.Success,
			Error:     e.Error,
			Metadata:  e.Metadata,
		}
	}

	return printer.Print(out)
}
