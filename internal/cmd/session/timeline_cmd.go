package session

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

// timelineOptions holds the flag values for `ari session timeline`.
type timelineOptions struct {
	since      time.Duration
	eventType  string
	last       int
	fromEvents bool
}

func newTimelineCmd(ctx *cmdContext) *cobra.Command {
	var opts timelineOptions

	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show session timeline",
		Long: `Reads and displays the curated timeline for the current session.

The default source is the ## Timeline section of SESSION_CONTEXT.md, which
contains the curated view of significant events (decisions, agent delegations,
commits, phase transitions, and session lifecycle events).

Use --from-events to query the raw event backplane (events.jsonl) instead.
This includes all v3 TypedEvents that project to the timeline, even if the
timeline section has not been updated yet.

Context:
  Use this command to review what happened during the session.
  The default source is the Timeline section of SESSION_CONTEXT.md (curated view).
  Use --from-events for the full raw event stream (includes backplane-only events).
  Agents should prefer the curated view unless they need low-level detail.
  Use 'ari session status' for session metadata (status, phase, initiative).

Examples:
  ari session timeline
  ari session timeline --since=1h
  ari session timeline --type=DECISION
  ari session timeline --last=10
  ari session timeline --from-events
  ari session timeline --from-events --type=agent.delegated
  ari session timeline -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimeline(ctx, opts)
		},
	}

	cmd.Flags().DurationVar(&opts.since, "since", 0,
		"Show entries from the last N duration (e.g., 1h, 30m)")
	cmd.Flags().StringVar(&opts.eventType, "type", "",
		"Filter by category: SESSION, AGENT, COMMIT, DECISION, PHASE, COMMAND")
	cmd.Flags().IntVar(&opts.last, "last", 0,
		"Show only the last N entries (default: all)")
	cmd.Flags().BoolVar(&opts.fromEvents, "from-events", false,
		"Read from events.jsonl instead of Timeline section")

	return cmd
}

func runTimeline(ctx *cmdContext, opts timelineOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err)
	}

	if sessionID == "" {
		return errors.ErrSessionNotFound("")
	}

	var entries []output.TimelineEntryOutput
	var total int

	if opts.fromEvents {
		// Path B: read curated TypedEvents from events.jsonl.
		eventsPath := resolver.SessionEventsFile(sessionID)
		entries, total, err = readFromEvents(eventsPath, opts)
		if err != nil {
			return err
		}
	} else {
		// Path A: read ## Timeline section from SESSION_CONTEXT.md.
		ctxPath := resolver.SessionContextFile(sessionID)
		entries, total, err = readFromTimeline(ctxPath, opts)
		if err != nil {
			return err
		}
	}

	result := output.TimelineOutput{
		SessionID: sessionID,
		Entries:   entries,
		Total:     total,
		Filtered:  len(entries),
	}

	return printer.Print(result)
}

// readFromTimeline reads and filters entries from the ## Timeline section of SESSION_CONTEXT.md.
func readFromTimeline(ctxPath string, opts timelineOptions) ([]output.TimelineEntryOutput, int, error) {
	rawEntries, err := sess.ReadTimeline(ctxPath)
	if err != nil {
		return nil, 0, err
	}

	total := len(rawEntries)
	filtered := filterTimeline(rawEntries, opts)

	entries := make([]output.TimelineEntryOutput, len(filtered))
	for i, e := range filtered {
		entries[i] = output.TimelineEntryOutput{
			Time:     e.Time.Format("15:04"),
			Category: strings.TrimRight(e.Category, " "),
			Summary:  e.Summary,
		}
	}

	return entries, total, nil
}

// filterTimeline applies --since, --type, and --last filters to timeline entries.
// Filter application order: --since, --type, then --last.
func filterTimeline(entries []sess.TimelineEntry, opts timelineOptions) []sess.TimelineEntry {
	var filtered []sess.TimelineEntry

	// Compute since cutoff using today's date with the HH:MM from entries.
	// Timeline entries only have HH:MM (no date), so we assume today's date.
	var sinceCutoff time.Time
	if opts.since > 0 {
		sinceCutoff = time.Now().Add(-opts.since)
	}

	for _, e := range entries {
		// --since filter: compare HH:MM against today's cutoff time.
		if !sinceCutoff.IsZero() {
			// Reconstruct today's date with entry's HH:MM.
			now := time.Now()
			entryTime := time.Date(
				now.Year(), now.Month(), now.Day(),
				e.Time.Hour(), e.Time.Minute(), 0, 0,
				now.Location(),
			)
			if entryTime.Before(sinceCutoff) {
				continue
			}
		}

		// --type filter: match category case-insensitively.
		// Category is 8-char padded (e.g., "SESSION "), so trim before compare.
		if opts.eventType != "" {
			trimmedCategory := strings.TrimRight(e.Category, " ")
			if !strings.EqualFold(trimmedCategory, opts.eventType) {
				continue
			}
		}

		filtered = append(filtered, e)
	}

	// --last filter: take the last N from remaining entries.
	if opts.last > 0 && len(filtered) > opts.last {
		filtered = filtered[len(filtered)-opts.last:]
	}

	return filtered
}

// readFromEvents reads curated TypedEvents from events.jsonl and formats them as timeline entries.
// Only events for which IsCuratedType returns true are included.
func readFromEvents(eventsPath string, opts timelineOptions) ([]output.TimelineEntryOutput, int, error) {
	typedEvents, err := readTypedEventsFromPath(eventsPath)
	if err != nil {
		return nil, 0, err
	}

	// Filter to curated types only.
	var curated []clewcontract.TypedEvent
	for _, e := range typedEvents {
		if sess.IsCuratedType(e.Type) {
			curated = append(curated, e)
		}
	}

	total := len(curated)

	// Apply --since filter using full timestamps from events.
	var sinceCutoff time.Time
	if opts.since > 0 {
		sinceCutoff = time.Now().Add(-opts.since)
	}

	var filtered []clewcontract.TypedEvent
	for _, e := range curated {
		if !sinceCutoff.IsZero() {
			// Parse event timestamp (millisecond UTC format or RFC3339).
			eventTime, parseErr := time.Parse("2006-01-02T15:04:05.000Z", e.Ts)
			if parseErr != nil {
				eventTime, parseErr = time.Parse(time.RFC3339, e.Ts)
				if parseErr != nil {
					// Unparseable timestamp — include the event (don't silently exclude).
					filtered = append(filtered, e)
					continue
				}
			}
			if eventTime.Before(sinceCutoff) {
				continue
			}
		}

		// --type filter: match against the full event type string (e.g. "agent.delegated")
		// or against the derived category (e.g. "AGENT").
		if opts.eventType != "" {
			category := strings.TrimRight(sess.EventTypeToCategory(e.Type), " ")
			eventTypeStr := string(e.Type)
			if !strings.EqualFold(category, opts.eventType) &&
				!strings.EqualFold(eventTypeStr, opts.eventType) {
				continue
			}
		}

		filtered = append(filtered, e)
	}

	// --last filter.
	if opts.last > 0 && len(filtered) > opts.last {
		filtered = filtered[len(filtered)-opts.last:]
	}

	entries := make([]output.TimelineEntryOutput, len(filtered))
	for i, e := range filtered {
		// Parse HH:MM from event timestamp.
		timeHHMM := "00:00"
		if ts, parseErr := time.Parse("2006-01-02T15:04:05.000Z", e.Ts); parseErr == nil {
			timeHHMM = ts.UTC().Format("15:04")
		} else if ts, parseErr := time.Parse(time.RFC3339, e.Ts); parseErr == nil {
			timeHHMM = ts.UTC().Format("15:04")
		}

		category := strings.TrimRight(sess.EventTypeToCategory(e.Type), " ")
		summary := sess.ExtractSummary(e)

		entries[i] = output.TimelineEntryOutput{
			Time:     timeHHMM,
			Category: category,
			Summary:  summary,
		}
	}

	return entries, total, nil
}

// readTypedEventsFromPath reads v3 TypedEvents from a JSONL file.
// Lines without a "data" field are skipped (not v3 TypedEvents).
// Returns empty slice (not error) if the file does not exist.
func readTypedEventsFromPath(path string) ([]clewcontract.TypedEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to open events file", err)
	}
	defer f.Close()

	// typedEventDetector checks for the "data" field to identify v3 TypedEvents.
	type typedEventDetector struct {
		Data json.RawMessage `json:"data"`
	}

	var events []clewcontract.TypedEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Only parse lines with "data" field (v3 TypedEvent).
		var detector typedEventDetector
		if err := json.Unmarshal(line, &detector); err != nil || detector.Data == nil {
			continue
		}

		var te clewcontract.TypedEvent
		if err := json.Unmarshal(line, &te); err != nil || string(te.Type) == "" {
			continue
		}

		events = append(events, te)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read events file", err)
	}

	return events, nil
}

