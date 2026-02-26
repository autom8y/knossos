package session

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

type listOptions struct {
	all    bool
	status string
	limit  int
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions with status filtering",
		Long: `List sessions with optional filtering by status.

Parked sessions older than ARIADNE_STALE_SESSION_DAYS (default: 2 days)
are annotated as stale with a suggestion to wrap them. Use --all to
include archived sessions from the archive directory.

Examples:
  ari session list
  ari session list --status PARKED
  ari session list --all -n 50
  ari session list -o json

Context:
  Use this to discover parked sessions before creating new ones.
  Stale annotations help agents identify sessions to wrap or gc.
  Prefer 'ari session status' for current session detail.
  Use 'ari session gc' to batch-archive stale parked sessions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Include archived sessions")
	cmd.Flags().StringVar(&opts.status, "status", "", "Filter by status: ACTIVE, PARKED, ARCHIVED")
	cmd.Flags().IntVarP(&opts.limit, "limit", "n", 20, "Maximum sessions to return")

	return cmd
}

func runList(ctx *cmdContext, opts listOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get current session ID
	currentID, _ := session.FindActiveSession(resolver.SessionsDir())

	var sessions []output.SessionSummary

	// Scan sessions directory
	sessionsDir := resolver.SessionsDir()
	if entries, err := os.ReadDir(sessionsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || !paths.IsSessionDir(entry.Name()) {
				continue
			}

			summary, ok := loadSessionSummary(resolver, entry.Name(), currentID)
			if !ok {
				continue
			}

			// Apply status filter
			if opts.status != "" && summary.Status != opts.status {
				continue
			}

			// Skip archived if not --all
			if !opts.all && summary.Status == "ARCHIVED" {
				continue
			}

			sessions = append(sessions, summary)
		}
	}

	// If --all, also scan archive directory
	if opts.all {
		archiveDir := resolver.ArchiveDir()
		if entries, err := os.ReadDir(archiveDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() || !paths.IsSessionDir(entry.Name()) {
					continue
				}

				// Load from archive
				archiveResolver := paths.NewResolver(filepath.Dir(filepath.Dir(archiveDir)))
				summary, ok := loadSessionSummaryFromPath(
					filepath.Join(archiveDir, entry.Name(), "SESSION_CONTEXT.md"),
					entry.Name(),
					currentID,
				)
				if !ok {
					// Try with archive resolver
					_ = archiveResolver
					continue
				}

				// Apply status filter
				if opts.status != "" && summary.Status != opts.status {
					continue
				}

				sessions = append(sessions, summary)
			}
		}
	}

	// Sort by created_at descending
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt > sessions[j].CreatedAt
	})

	// Apply limit
	if opts.limit > 0 && len(sessions) > opts.limit {
		sessions = sessions[:opts.limit]
	}

	result := output.SessionListOutput{
		Sessions:       sessions,
		Total:          len(sessions),
		CurrentSession: currentID,
	}

	return printer.Print(result)
}

func loadSessionSummary(resolver *paths.Resolver, sessionID, currentID string) (output.SessionSummary, bool) {
	ctxPath := resolver.SessionContextFile(sessionID)
	return loadSessionSummaryFromPath(ctxPath, sessionID, currentID)
}

func loadSessionSummaryFromPath(ctxPath, sessionID, currentID string) (output.SessionSummary, bool) {
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		return output.SessionSummary{}, false
	}

	summary := output.SessionSummary{
		SessionID:  sessionID,
		Status:     string(sessCtx.Status),
		Initiative: sessCtx.Initiative,
		Complexity: sessCtx.Complexity,
		CreatedAt:  sessCtx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Current:    sessionID == strings.TrimSpace(currentID),
	}

	if sessCtx.ParkedAt != nil {
		summary.ParkedAt = sessCtx.ParkedAt.Format("2006-01-02T15:04:05Z")
	}

	// Annotate stale parked sessions
	if sessCtx.Status == session.StatusParked && sessCtx.ParkedAt != nil {
		threshold := staleSessionThreshold()
		if time.Since(*sessCtx.ParkedAt) > threshold {
			summary.Stale = true
			summary.StaleHint = "consider: ari session wrap " + sessionID
		}
	}

	return summary, true
}

// staleSessionThreshold returns the duration after which a parked session
// is considered stale. Configurable via ARIADNE_STALE_SESSION_DAYS (default: 2).
func staleSessionThreshold() time.Duration {
	days := 2
	if env := os.Getenv("ARIADNE_STALE_SESSION_DAYS"); env != "" {
		if d, err := strconv.Atoi(env); err == nil && d > 0 {
			days = d
		}
	}
	return time.Duration(days) * 24 * time.Hour
}
