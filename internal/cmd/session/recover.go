package session

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
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
						// Emit recovery event to affected session (non-fatal)
						sessionID := strings.TrimSuffix(entry.Name(), ".lock")
						if paths.IsSessionDir(sessionID) {
							sessionDir := resolver.SessionDir(sessionID)
							w := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
							w.Write(clewcontract.NewToolCallEvent("recover", lockPath, map[string]interface{}{
								"action":     "stale_lock_removed",
								"session_id": sessionID,
							}))
							w.Flush()
							w.Close()
						}
					}
				}
			}
		}
	}

	// Step 2: Clean up orphaned CC map entries
	ccMapDir := resolver.CCMapDir()
	var ccMapOrphans []string
	var removedCCMapOrphans []string
	if entries, err := os.ReadDir(ccMapDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			mapFile := filepath.Join(ccMapDir, entry.Name())
			data, readErr := os.ReadFile(mapFile)
			if readErr != nil {
				continue
			}
			knossosID := strings.TrimSpace(string(data))
			// Check if the mapped session still exists
			sessionDir := resolver.SessionDir(knossosID)
			if _, statErr := os.Stat(sessionDir); os.IsNotExist(statErr) {
				ccMapOrphans = append(ccMapOrphans, entry.Name()+" -> "+knossosID)
				if !opts.dryRun {
					if removeErr := os.Remove(mapFile); removeErr == nil {
						removedCCMapOrphans = append(removedCCMapOrphans, entry.Name())
					}
				}
			}
		}
	}

	// Step 3: Also clean up stale .current-session file if it still exists
	currentSessionFile := resolver.CurrentSessionFile()
	currentSessionCleaned := false
	if _, err := os.Stat(currentSessionFile); err == nil {
		if !opts.dryRun {
			if removeErr := os.Remove(currentSessionFile); removeErr == nil {
				currentSessionCleaned = true
			}
		} else {
			currentSessionCleaned = true // would be cleaned
		}
	}

	// Find active session
	activeID, err := session.FindActiveSession(resolver.SessionsDir())
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build result
	result := output.RecoverOutput{
		StaleLocks:            staleLocks,
		RemovedLocks:          removedLocks,
		ActiveSession:         activeID,
		CCMapOrphans:          ccMapOrphans,
		RemovedCCMapOrphans:   removedCCMapOrphans,
		CurrentSessionCleaned: currentSessionCleaned,
		DryRun:                opts.dryRun,
	}

	if len(staleLocks) == 0 && len(ccMapOrphans) == 0 && !currentSessionCleaned {
		result.Summary = "All healthy. No stale locks, CC map orphans, or legacy cache files found."
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
