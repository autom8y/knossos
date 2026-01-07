package sync

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/sync"
)

func newStatusCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status for all tracked paths",
		Long: `Shows the sync status of all tracked configuration files.

Displays:
  - Remote source URL
  - Last sync timestamp
  - Status of each tracked file (synced, modified, conflict)
  - Any unresolved conflicts`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx)
		},
	}

	return cmd
}

func runStatus(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

	stateManager := sync.NewStateManager(resolver)
	state, err := stateManager.Load()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := output.SyncStatusOutput{
		Initialized:  state != nil,
		TrackedPaths: []output.SyncTrackedPath{},
		Conflicts:    []output.SyncConflictEntry{},
	}

	if state == nil {
		return printer.Print(out)
	}

	out.Remote = state.Remote
	if !state.LastSync.IsZero() {
		out.LastSync = state.LastSync.Format("2006-01-02 15:04:05")
	}

	// Build tracker to refresh state
	tracker := sync.NewTracker(resolver, stateManager)
	if err := tracker.RefreshAll(state); err != nil {
		printer.PrintError(err)
		return err
	}

	// Build tracked paths
	for path, tracked := range state.TrackedFiles {
		out.TrackedPaths = append(out.TrackedPaths, output.SyncTrackedPath{
			Path:         path,
			Status:       tracked.Status,
			LocalHash:    shortenHash(tracked.LocalHash),
			RemoteHash:   shortenHash(tracked.RemoteHash),
			BaseHash:     shortenHash(tracked.BaseHash),
			LastModified: tracked.LastModified.Format("2006-01-02 15:04:05"),
		})
	}

	// Build conflicts
	if state.HasConflicts() {
		out.HasConflicts = true
		for _, c := range state.Conflicts {
			out.Conflicts = append(out.Conflicts, output.SyncConflictEntry{
				Path:        c.Path,
				Description: c.Description,
				LocalHash:   shortenHash(c.LocalHash),
				RemoteHash:  shortenHash(c.RemoteHash),
				BaseHash:    shortenHash(c.BaseHash),
			})
		}
	}

	return printer.Print(out)
}

// shortenHash returns first 8 characters of a hash.
func shortenHash(hash string) string {
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}
