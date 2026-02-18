package session

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/naxos"
)

type gcOptions struct {
	staleDays int
	force     bool
	dryRun    bool
}

func newGcCmd(ctx *cmdContext) *cobra.Command {
	var opts gcOptions

	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Archive stale parked sessions",
		Long: `Scans for PARKED sessions older than the stale threshold and archives them.

The default stale threshold is set by ARIADNE_STALE_SESSION_DAYS (default: 2 days).
Override with --stale-after for a one-off run.

Without --force, prompts for confirmation before wrapping each session.
With --dry-run, only lists stale sessions without archiving them.

Examples:
  ari session gc
  ari session gc --stale-after 7
  ari session gc --force
  ari session gc --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGc(ctx, opts)
		},
	}

	cmd.Flags().IntVar(&opts.staleDays, "stale-after", 0, "Days parked before considered stale (default: ARIADNE_STALE_SESSION_DAYS or 2)")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Archive all stale sessions without prompting")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "List stale sessions without archiving")

	return cmd
}

func runGc(ctx *cmdContext, opts gcOptions) error {
	resolver := ctx.GetResolver()

	// Determine stale threshold: --stale-after flag > env var > default (2 days)
	var threshold time.Duration
	if opts.staleDays > 0 {
		threshold = time.Duration(opts.staleDays) * 24 * time.Hour
	} else {
		threshold = staleSessionThreshold()
	}

	// Discover stale PARKED sessions
	staleSessions := naxos.ScanStaleSessions(resolver.SessionsDir(), threshold, "")
	if len(staleSessions) == 0 {
		fmt.Fprintln(os.Stdout, "No stale sessions found.")
		return nil
	}

	// Print what was found
	fmt.Fprintf(os.Stdout, "Found %d stale parked session(s) (parked > %s):\n\n", len(staleSessions), formatThreshold(threshold))
	for i, s := range staleSessions {
		fmt.Fprintf(os.Stdout, "  %d. %s  (parked %s ago)\n", i+1, s.ID, naxos.FormatDuration(s.Age))
	}
	fmt.Fprintln(os.Stdout)

	if opts.dryRun {
		fmt.Fprintln(os.Stdout, "Dry-run: no sessions archived. Use --force or confirm interactively.")
		return nil
	}

	// Confirm unless --force
	if !opts.force {
		fmt.Fprintf(os.Stdout, "Archive %d session(s)? [y/N] ", len(staleSessions))
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			fmt.Fprintln(os.Stdout, "Aborted.")
			return nil
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stdout, "Aborted.")
			return nil
		}
	}

	// Archive each stale session via runWrap (gets all Session A boundary fixes)
	archived := 0
	for _, s := range staleSessions {
		wrapCtx := newGcWrapContext(ctx, s.ID)
		if err := runWrap(wrapCtx, wrapOptions{noArchive: false, force: true}); err != nil {
			fmt.Fprintf(os.Stderr, "  warn: failed to archive %s: %v\n", s.ID, err)
			continue
		}
		archived++
	}

	fmt.Fprintf(os.Stdout, "\nArchived %d/%d session(s).\n", archived, len(staleSessions))
	return nil
}

// newGcWrapContext creates a cmdContext scoped to a specific session ID
// for use by the gc command when wrapping stale sessions.
// Inherits output format and project dir from the parent context.
func newGcWrapContext(parent *cmdContext, sessionID string) *cmdContext {
	id := sessionID
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: parent.SessionContext.BaseContext,
			SessionID:   &id,
		},
	}
}

// formatThreshold formats a duration as a human-readable threshold description.
func formatThreshold(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
