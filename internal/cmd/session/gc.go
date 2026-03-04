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
	"github.com/autom8y/knossos/internal/output"
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
		Long: `Scan for PARKED sessions older than the stale threshold and archive them.

The default stale threshold is set by ARI_STALE_SESSION_DAYS (default: 2 days).
Override with --stale-after for a one-off run.

Without --force, prompt for confirmation before wrapping each session.
With --dry-run, only list stale sessions without archiving them.

Examples:
  ari session gc
  ari session gc --stale-after 7
  ari session gc --force
  ari session gc --dry-run

Context:
  Batch housekeeping command. Run periodically or after 'ari session wrap'.
  Delegates to runWrap internally -- each archived session gets full wrap treatment.
  Use --dry-run first to preview which sessions would be archived.
  Agents should not call this directly -- it is a user-facing maintenance tool.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGc(ctx, opts)
		},
	}

	cmd.Flags().IntVar(&opts.staleDays, "stale-after", 0, "Days parked before considered stale (default: ARI_STALE_SESSION_DAYS or 2)")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Archive all stale sessions without prompting")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "List stale sessions without archiving")

	return cmd
}

// gcOutput is the structured output for ari session gc.
type gcOutput struct {
	StaleCount int             `json:"stale_count"`
	Threshold  string          `json:"threshold"`
	Sessions   []gcSessionInfo `json:"sessions"`
	Archived   int             `json:"archived"`
	Total      int             `json:"total"`
	DryRun     bool            `json:"dry_run,omitempty"`
	Aborted    bool            `json:"aborted,omitempty"`
}

type gcSessionInfo struct {
	ID        string `json:"id"`
	ParkedAge string `json:"parked_age"`
}

// Text implements output.Textable.
func (g gcOutput) Text() string {
	var b strings.Builder

	if g.StaleCount == 0 {
		b.WriteString("No stale sessions found.\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Found %d stale parked session(s) (parked > %s):\n\n", g.StaleCount, g.Threshold))
	for i, s := range g.Sessions {
		b.WriteString(fmt.Sprintf("  %d. %s  (parked %s ago)\n", i+1, s.ID, s.ParkedAge))
	}
	b.WriteString("\n")

	if g.DryRun {
		b.WriteString("Dry-run: no sessions archived. Use --force or confirm interactively.\n")
		return b.String()
	}

	if g.Aborted {
		b.WriteString("Aborted.\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("\nArchived %d/%d session(s).\n", g.Archived, g.Total))
	return b.String()
}

func runGc(ctx *cmdContext, opts gcOptions) error {
	printer := ctx.GetPrinter(output.FormatText)
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
		return printer.Print(gcOutput{StaleCount: 0})
	}

	// Build session info
	sessions := make([]gcSessionInfo, len(staleSessions))
	for i, s := range staleSessions {
		sessions[i] = gcSessionInfo{ID: s.ID, ParkedAge: naxos.FormatDuration(s.Age)}
	}

	thresholdStr := formatThreshold(threshold)

	if opts.dryRun {
		return printer.Print(gcOutput{
			StaleCount: len(staleSessions),
			Threshold:  thresholdStr,
			Sessions:   sessions,
			DryRun:     true,
		})
	}

	// Print what was found (use text for interactive prompt)
	printer.PrintLine(fmt.Sprintf("Found %d stale parked session(s) (parked > %s):\n", len(staleSessions), thresholdStr))
	for i, s := range staleSessions {
		printer.PrintLine(fmt.Sprintf("  %d. %s  (parked %s ago)", i+1, s.ID, naxos.FormatDuration(s.Age)))
	}

	// Confirm unless --force
	if !opts.force {
		fmt.Fprintf(os.Stdout, "\nArchive %d session(s)? [y/N] ", len(staleSessions))
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return printer.Print(gcOutput{
				StaleCount: len(staleSessions),
				Threshold:  thresholdStr,
				Sessions:   sessions,
				Aborted:    true,
			})
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			return printer.Print(gcOutput{
				StaleCount: len(staleSessions),
				Threshold:  thresholdStr,
				Sessions:   sessions,
				Aborted:    true,
			})
		}
	}

	// Archive each stale session via runWrap (gets all Session A boundary fixes)
	archived := 0
	for _, s := range staleSessions {
		wrapCtx := newGcWrapContext(ctx, s.ID)
		if err := runWrap(wrapCtx, wrapOptions{noArchive: false, force: true}); err != nil {
			printer.VerboseLog("warn", fmt.Sprintf("failed to archive %s", s.ID), map[string]any{"error": err.Error()})
			fmt.Fprintf(os.Stderr, "  warn: failed to archive %s: %v\n", s.ID, err)
			continue
		}
		archived++
	}

	return printer.Print(gcOutput{
		StaleCount: len(staleSessions),
		Threshold:  thresholdStr,
		Sessions:   sessions,
		Archived:   archived,
		Total:      len(staleSessions),
	})
}

// newGcWrapContext creates a cmdContext scoped to a specific session ID
// for use by the gc command when wrapping stale sessions.
// Inherits output format and project dir from the parent context.
func newGcWrapContext(parent *cmdContext, sessionID string) *cmdContext {
	id := sessionID
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: parent.BaseContext,
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
