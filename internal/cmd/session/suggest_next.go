package session

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/naxos"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// SuggestNextOutput is the structured output for ari session suggest-next.
type SuggestNextOutput struct {
	HasTriage       bool     `json:"has_triage"`
	CriticalOrphans int      `json:"critical_orphans"`
	HighOrphans     int      `json:"high_orphans"`
	TotalOrphans    int      `json:"total_orphans"`
	SuggestedAction string   `json:"suggested_action"`
	Rationale       string   `json:"rationale"`
	RelatedOrphans  []string `json:"related_orphans,omitempty"`
}

// Text implements output.Textable.
func (s SuggestNextOutput) Text() string {
	var b strings.Builder

	if !s.HasTriage {
		b.WriteString("No triage data available.\n")
		fmt.Fprintf(&b, "Suggested: %s\n", s.SuggestedAction)
		fmt.Fprintf(&b, "Rationale: %s\n", s.Rationale)
		return b.String()
	}

	fmt.Fprintf(&b, "Orphaned sessions: %d total (%d CRITICAL, %d HIGH)\n",
		s.TotalOrphans, s.CriticalOrphans, s.HighOrphans)

	if len(s.RelatedOrphans) > 0 {
		fmt.Fprintf(&b, "Related to current initiative: %s\n",
			strings.Join(s.RelatedOrphans, ", "))
	}

	fmt.Fprintf(&b, "Suggested: %s\n", s.SuggestedAction)
	fmt.Fprintf(&b, "Rationale: %s\n", s.Rationale)

	return b.String()
}

func newSuggestNextCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suggest-next",
		Short: "Suggest next session action based on triage data",
		Long: `Analyze NAXOS_TRIAGE.md and current session context to recommend
the most appropriate next action for session management.

Returns a structured recommendation including critical/high orphan counts,
a suggested action (resume/wrap/new session), and supporting rationale.
When a current session exists and has initiative keywords matching an orphaned
session, the suggestion prioritizes resuming that related session.

Examples:
  ari session suggest-next
  ari session suggest-next -o json

Context:
  Consumed by Moirai to drive session routing decisions.
  Output is stable and machine-parseable with -o json.
  Run 'ari naxos triage' first to generate fresh triage data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSuggestNext(ctx)
		},
	}

	return cmd
}

// runSuggestNext reads the triage artifact and current session to produce a
// recommendation for what to do next.
func runSuggestNext(ctx *cmdContext) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()
	sessionsDir := resolver.SessionsDir()

	// Step 1: Read triage artifact. Non-existence is not an error — it means
	// no triage has been run yet, which is a valid state for fresh projects.
	triageResult, err := naxos.ReadTriageArtifact(sessionsDir)
	if err != nil {
		// No artifact: return a safe default recommendation.
		return printer.Print(SuggestNextOutput{
			HasTriage:       false,
			SuggestedAction: "new session",
			Rationale:       "No triage data available",
		})
	}

	// Step 2: Read current session initiative (best-effort; empty string is fine).
	currentInitiative := readCurrentInitiative(ctx)

	// Step 3: Enrich entries with initiative data by loading each orphan's
	// SESSION_CONTEXT.md. The artifact only stores a summary — Initiative is
	// not persisted in the markdown table, so we resolve it here.
	enrichedEntries := enrichWithInitiatives(triageResult.Entries, resolver)

	// Step 4: Count orphans by severity.
	critical := triageResult.BySeverity[naxos.SeverityCritical]
	high := triageResult.BySeverity[naxos.SeverityHigh]
	total := triageResult.TotalTriaged

	// Step 5: Find orphan entries whose initiatives overlap the current initiative keywords.
	related := matchingOrphans(enrichedEntries, currentInitiative)

	// Step 6: Build the recommendation.
	action, rationale := buildRecommendation(enrichedEntries, critical, high, related)

	return printer.Print(SuggestNextOutput{
		HasTriage:       true,
		CriticalOrphans: critical,
		HighOrphans:     high,
		TotalOrphans:    total,
		SuggestedAction: action,
		Rationale:       rationale,
		RelatedOrphans:  related,
	})
}

// readCurrentInitiative looks up the active session's initiative. Returns an
// empty string when no session exists or the context cannot be read — callers
// must treat this as "unknown initiative".
func readCurrentInitiative(ctx *cmdContext) string {
	sessionID, err := ctx.GetSessionID()
	if err != nil || sessionID == "" {
		return ""
	}

	resolver := ctx.GetResolver()
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		return ""
	}

	return sessCtx.Initiative
}

// enrichWithInitiatives loads SESSION_CONTEXT.md for each triage entry to fill
// in the Initiative field that is not preserved in the artifact markdown table.
// Entries whose session directory no longer exists or cannot be parsed are
// returned unchanged (Initiative remains empty).
func enrichWithInitiatives(entries []naxos.TriageEntry, resolver *paths.Resolver) []naxos.TriageEntry {
	if len(entries) == 0 {
		return entries
	}

	enriched := make([]naxos.TriageEntry, len(entries))
	copy(enriched, entries)

	for i, entry := range enriched {
		if entry.Initiative != "" {
			// Already populated — skip.
			continue
		}
		ctxPath := resolver.SessionContextFile(entry.SessionID)
		sessCtx, err := session.LoadContext(ctxPath)
		if err != nil {
			// Session may have been deleted; skip silently.
			continue
		}
		enriched[i].Initiative = sessCtx.Initiative
	}

	return enriched
}

// matchingOrphans returns the session IDs of triage entries whose initiative
// shares at least one non-trivial keyword with the current initiative.
// Returns nil when currentInitiative is empty or no matches are found.
func matchingOrphans(entries []naxos.TriageEntry, currentInitiative string) []string {
	if currentInitiative == "" {
		return nil
	}

	keywords := extractKeywords(currentInitiative)
	if len(keywords) == 0 {
		return nil
	}

	var matched []string
	for _, entry := range entries {
		if entry.Initiative == "" {
			continue
		}
		entryKeywords := extractKeywords(entry.Initiative)
		if hasOverlap(keywords, entryKeywords) {
			matched = append(matched, entry.SessionID)
		}
	}

	return matched
}

// extractKeywords lowercases and splits a string into words, filtering out
// common stop-words and short tokens that are not useful for matching.
func extractKeywords(s string) map[string]struct{} {
	stopWords := map[string]struct{}{
		"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "for": {},
		"in": {}, "on": {}, "at": {}, "to": {}, "of": {}, "is": {},
		"it": {}, "as": {}, "be": {}, "by": {}, "do": {},
	}

	words := strings.Fields(strings.ToLower(s))
	result := make(map[string]struct{}, len(words))
	for _, w := range words {
		// Strip punctuation from word edges.
		w = strings.Trim(w, ".,;:!?\"'()-")
		if len(w) < 3 {
			continue
		}
		if _, isStop := stopWords[w]; isStop {
			continue
		}
		result[w] = struct{}{}
	}

	return result
}

// hasOverlap returns true when the two keyword sets share at least one element.
func hasOverlap(a, b map[string]struct{}) bool {
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return false
}

// buildRecommendation applies the priority rules from the design spec to choose
// a suggested action and rationale string.
//
// Priority order:
//  1. Critical orphans with initiative match → "resume {id}"
//  2. Critical orphans without match → "wrap {id}"
//  3. Only high orphans → "Consider triaging with /naxos"
//  4. No actionable orphans → "new session"
func buildRecommendation(
	entries []naxos.TriageEntry,
	critical, high int,
	relatedIDs []string,
) (action, rationale string) {
	relatedSet := make(map[string]struct{}, len(relatedIDs))
	for _, id := range relatedIDs {
		relatedSet[id] = struct{}{}
	}

	if critical > 0 {
		// Walk entries in priority order (already sorted by triage).
		for _, entry := range entries {
			if entry.Severity != naxos.SeverityCritical {
				continue
			}
			if _, related := relatedSet[entry.SessionID]; related {
				return fmt.Sprintf("resume %s", entry.SessionID),
					fmt.Sprintf("Critical orphan %s matches current initiative", entry.SessionID)
			}
		}
		// Critical orphan exists but no initiative match — recommend wrap.
		for _, entry := range entries {
			if entry.Severity == naxos.SeverityCritical {
				return fmt.Sprintf("wrap %s", entry.SessionID),
					fmt.Sprintf("Critical orphan %s requires attention", entry.SessionID)
			}
		}
	}

	if high > 0 {
		return "Consider triaging with /naxos",
			fmt.Sprintf("%d HIGH severity orphan(s) need review", high)
	}

	return "new session", "All sessions healthy"
}
