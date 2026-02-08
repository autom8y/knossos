package session

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

type recoverOptions struct {
	dryRun bool
}

func newRecoverCmd(ctx *cmdContext) *cobra.Command {
	var opts recoverOptions

	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Clean up stale locks and rebuild session cache",
		Long: `Recovers from stale locks and inconsistent session state.

Actions performed:
  1. Scans all lock files for stale entries (older than 5 minutes)
  2. Removes stale lock files
  3. Scans session directories for the ACTIVE session
  4. Rebuilds .current-session cache

Use --dry-run to preview what would be fixed without making changes.

Examples:
  ari session recover
  ari session recover --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecover(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without applying")

	return cmd
}

func runRecover(ctx *cmdContext, opts recoverOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	var staleLocks []string
	var removedLocks []string

	// Step 1: Scan lock files for stale entries
	locksDir := lockMgr.LocksDir()
	if entries, err := os.ReadDir(locksDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lock") {
				continue
			}

			lockPath := filepath.Join(locksDir, entry.Name())
			if isAdvisoryLockStale(lockPath) {
				staleLocks = append(staleLocks, entry.Name())

				if !opts.dryRun {
					if err := os.Remove(lockPath); err == nil {
						removedLocks = append(removedLocks, entry.Name())
					}
				}
			}
		}
	}

	// Step 2: Scan for active session and rebuild cache
	activeID, err := session.FindActiveSession(resolver.SessionsDir())
	if err != nil {
		printer.PrintError(err)
		return err
	}

	cacheFile := resolver.CurrentSessionFile()
	cacheStatus := "unchanged"
	currentCacheID := ""
	if data, err := os.ReadFile(cacheFile); err == nil {
		currentCacheID = strings.TrimSpace(string(data))
	}

	if activeID != currentCacheID {
		cacheStatus = "rebuilt"
		if !opts.dryRun {
			if activeID != "" {
				os.WriteFile(cacheFile, []byte(activeID), 0644)
			} else {
				os.Remove(cacheFile)
			}
		}
	}

	// Build result
	result := output.RecoverOutput{
		StaleLocks:     staleLocks,
		RemovedLocks:   removedLocks,
		ActiveSession:  activeID,
		CacheStatus:    cacheStatus,
		PreviousCache:  currentCacheID,
		DryRun:         opts.dryRun,
	}

	if len(staleLocks) == 0 && cacheStatus == "unchanged" {
		result.Summary = "All healthy. No stale locks found, cache is consistent."
	} else if opts.dryRun {
		result.Summary = "Issues found. Run without --dry-run to fix."
	} else {
		result.Summary = "Recovery complete."
	}

	return printer.PrintSuccess(result)
}

// isAdvisoryLockStale checks if an advisory lock file is stale.
// Uses treatLegacyAsStale=true because recovery should aggressively clean up legacy locks.
func isAdvisoryLockStale(lockPath string) bool {
	return lock.IsStaleFile(lockPath, true)
}
