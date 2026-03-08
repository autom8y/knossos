package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/hook/clewcontract"
)

// SnapshotRole identifies the agent role for budget allocation.
type SnapshotRole string

const (
	// RoleOrchestrator is for agents that coordinate workflows (potnia, moirai, etc.).
	// Budget: last 10 timeline entries, all decisions, blockers.
	RoleOrchestrator SnapshotRole = "orchestrator"

	// RoleSpecialist is for agents that appear in the active rite's workflow phases.
	// Budget: last 5 + up to 3 agent-scoped timeline entries, blockers, no decisions.
	RoleSpecialist SnapshotRole = "specialist"

	// RoleBackground is for known agents not in the orchestrator set or rite workflow.
	// Budget: phase name only — no timeline, no decisions, no blockers.
	RoleBackground SnapshotRole = "background"
)

// SnapshotConfig holds the role and agent identity for snapshot generation.
type SnapshotConfig struct {
	Role      SnapshotRole
	AgentName string // For specialist scoping. May be empty.
}

// DecisionEntry holds an extracted decision from events.jsonl.
type DecisionEntry struct {
	Decision  string
	Rationale string
	Timestamp time.Time
}

// Snapshot holds the role-adaptive session context projection.
type Snapshot struct {
	// Session metadata (always present)
	Phase      string
	Complexity string
	Status     string
	Initiative string
	Rite       string // Omitted for specialist/background in JSON

	// Filtered by role
	Timeline  []TimelineEntry // Empty for background
	Decisions []DecisionEntry // Only for orchestrator
	Blockers  string          // Empty string when absent; omitted for background

	// Identity
	Role        SnapshotRole
	AgentName   string
	GeneratedAt time.Time
}

// GenerateSnapshot builds a role-adaptive session context projection.
// It reads SESSION_CONTEXT.md (via the already-loaded ctx) for metadata and body,
// and reads events.jsonl from sessionDir for timeline and decisions.
//
// eventsPath is the absolute path to events.jsonl for the session.
// Returns a valid (possibly minimal) Snapshot on any error — it never fails.
func GenerateSnapshot(ctx *Context, eventsPath string, config SnapshotConfig) (*Snapshot, error) {
	snap := &Snapshot{
		Phase:       ctx.CurrentPhase,
		Complexity:  ctx.Complexity,
		Status:      string(ctx.Status),
		Initiative:  ctx.Initiative,
		Rite:        ctx.ActiveRite,
		Role:        config.Role,
		AgentName:   config.AgentName,
		GeneratedAt: time.Now().UTC(),
	}

	switch config.Role {
	case RoleOrchestrator:
		// Full budget: last 10 timeline entries, all decisions, blockers.
		typedEvents, _ := readSnapshotEvents(eventsPath)
		snap.Timeline = buildOrchestratorTimeline(typedEvents, 10)
		snap.Decisions = extractDecisions(typedEvents)
		snap.Blockers = extractBlockers(ctx.Body)

	case RoleSpecialist:
		// Medium budget: scoped timeline (max 8), blockers, no decisions section.
		typedEvents, _ := readSnapshotEvents(eventsPath)
		snap.Timeline = buildSpecialistTimeline(typedEvents, config.AgentName)
		snap.Blockers = extractBlockers(ctx.Body)
		// Decisions omitted for specialists (they see decisions inline in timeline).

	case RoleBackground:
		// Minimal: phase/complexity/status only. No timeline, no decisions, no blockers.
	}

	return snap, nil
}

// RenderMarkdown renders the snapshot as role-adaptive markdown for injection into agent context.
// Follows the templates from SESSION-5 Section 5.
func (s *Snapshot) RenderMarkdown() string {
	var b strings.Builder

	// Background: single-line minimal template.
	if s.Role == RoleBackground {
		fmt.Fprintf(&b, "## Session Context (auto-injected)\n")
		fmt.Fprintf(&b, "Phase: %s | Complexity: %s | Status: %s\n",
			s.Phase, s.Complexity, s.Status)
		return b.String()
	}

	// Orchestrator and specialist header.
	fmt.Fprintf(&b, "## Session Context (auto-injected)\n")
	fmt.Fprintf(&b, "Phase: %s | Complexity: %s | Status: %s\n",
		s.Phase, s.Complexity, s.Status)

	if s.Role == RoleOrchestrator {
		// Orchestrator shows initiative AND rite.
		rite := s.Rite
		if rite == "" {
			rite = "none"
		}
		fmt.Fprintf(&b, "Initiative: %s | Rite: %s\n", s.Initiative, rite)
	} else {
		// Specialist shows initiative only (already knows their rite).
		fmt.Fprintf(&b, "Initiative: %s\n", s.Initiative)
	}

	// Timeline section — omit when empty (fresh session).
	if len(s.Timeline) > 0 {
		if s.Role == RoleOrchestrator {
			fmt.Fprintf(&b, "\n### Timeline (last %d)\n", len(s.Timeline))
		} else {
			fmt.Fprintf(&b, "\n### Timeline (recent)\n")
		}
		for _, e := range s.Timeline {
			// Category is already 8-char padded per FormatTimelineEntry.
			fmt.Fprintf(&b, "- %s | %s | %s\n",
				e.Time.Format("15:04"), e.Category, e.Summary)
		}
	}

	// Decisions section — orchestrator only.
	if s.Role == RoleOrchestrator && len(s.Decisions) > 0 {
		fmt.Fprintf(&b, "\n### Decisions\n")
		for _, d := range s.Decisions {
			if d.Rationale != "" {
				// Truncate rationale to 30 chars per SESSION-2 Section 3.2 rule 4.
				rat := d.Rationale
				if len(rat) > 30 {
					rat = rat[:27] + "..."
				}
				fmt.Fprintf(&b, "- %s (%s)\n", d.Decision, rat)
			} else {
				fmt.Fprintf(&b, "- %s\n", d.Decision)
			}
		}
	}

	// Blockers section — orchestrator and specialist, but only when non-empty.
	if s.Blockers != "" {
		fmt.Fprintf(&b, "\n### Blockers\n%s", s.Blockers)
		// Ensure trailing newline.
		if !strings.HasSuffix(s.Blockers, "\n") {
			fmt.Fprintf(&b, "\n")
		}
	}

	return b.String()
}

// RenderJSON marshals the snapshot to JSON with role-based field omission.
// Fields use omitempty semantics per SESSION-5 Section 6.2.
func (s *Snapshot) RenderJSON() ([]byte, error) {
	type timelineJSON struct {
		Time     string `json:"time"`
		Category string `json:"category"`
		Summary  string `json:"summary"`
	}
	type decisionJSON struct {
		Decision  string `json:"decision"`
		Rationale string `json:"rationale,omitempty"`
	}

	// Build role-scoped JSON output per SESSION-5 Section 6.2.
	out := map[string]any{
		"role":          string(s.Role),
		"agent_name":    s.AgentName,
		"status":        s.Status,
		"initiative":    s.Initiative,
		"complexity":    s.Complexity,
		"current_phase": s.Phase,
		"generated_at":  s.GeneratedAt.UTC().Format(time.RFC3339Nano),
	}

	// active_rite: present for orchestrator only.
	if s.Role == RoleOrchestrator && s.Rite != "" {
		out["active_rite"] = s.Rite
	}

	// timeline: present for orchestrator and specialist only.
	if s.Role != RoleBackground && len(s.Timeline) > 0 {
		tl := make([]timelineJSON, len(s.Timeline))
		for i, e := range s.Timeline {
			cat := strings.TrimRight(e.Category, " ")
			tl[i] = timelineJSON{
				Time:     e.Time.Format("15:04"),
				Category: cat,
				Summary:  e.Summary,
			}
		}
		out["timeline"] = tl
	}

	// decisions: orchestrator only.
	if s.Role == RoleOrchestrator && len(s.Decisions) > 0 {
		ds := make([]decisionJSON, len(s.Decisions))
		for i, d := range s.Decisions {
			ds[i] = decisionJSON{
				Decision:  d.Decision,
				Rationale: d.Rationale,
			}
		}
		out["decisions"] = ds
	}

	// blockers: orchestrator and specialist, omit for background and when empty.
	if s.Role != RoleBackground && s.Blockers != "" {
		out["blockers"] = s.Blockers
	}

	return json.Marshal(out)
}

// --- internal helpers ---

// readSnapshotEvents reads v3 TypedEvents from events.jsonl.
// Returns empty slice (not error) if the file is absent or unreadable.
// This is a snapshot-local reader; it reads only TypedEvents (v3, has "data" field).
func readSnapshotEvents(eventsPath string) ([]clewcontract.TypedEvent, error) {
	f, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	type detector struct {
		Data json.RawMessage `json:"data"`
	}

	var events []clewcontract.TypedEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Only v3 TypedEvents have a "data" field.
		var det detector
		if err := json.Unmarshal(line, &det); err != nil || det.Data == nil {
			continue
		}

		var te clewcontract.TypedEvent
		if err := json.Unmarshal(line, &te); err != nil || string(te.Type) == "" {
			continue
		}

		events = append(events, te)
	}

	return events, scanner.Err()
}

// buildOrchestratorTimeline extracts the last N curated TypedEvents as TimelineEntry values.
func buildOrchestratorTimeline(events []clewcontract.TypedEvent, limit int) []TimelineEntry {
	var curated []clewcontract.TypedEvent
	for _, e := range events {
		if IsCuratedType(e.Type) {
			curated = append(curated, e)
		}
	}

	// Take last N.
	if len(curated) > limit {
		curated = curated[len(curated)-limit:]
	}

	return typedEventsToTimelineEntries(curated)
}

// buildSpecialistTimeline applies the specialist scoping algorithm (SESSION-5 Section 4).
// Returns up to 8 timeline entries: last 5 recent + up to 3 agent-scoped (deduped, sorted).
func buildSpecialistTimeline(events []clewcontract.TypedEvent, agentName string) []TimelineEntry {
	// Filter to curated types.
	var curated []clewcontract.TypedEvent
	for _, e := range events {
		if IsCuratedType(e.Type) {
			curated = append(curated, e)
		}
	}

	if len(curated) == 0 {
		return nil
	}

	// Pass 2: last 5 chronological entries.
	recentStart := max(len(curated)-5, 0)
	recent := curated[recentStart:]

	// Build a set of recent indices for deduplication.
	recentSet := make(map[int]bool, len(recent))
	for i := recentStart; i < len(curated); i++ {
		recentSet[i] = true
	}

	// Pass 3: agent-scoped entries not already in recent.
	var agentScoped []clewcontract.TypedEvent
	if agentName != "" {
		for i, e := range curated {
			if recentSet[i] {
				continue
			}
			if agentNameInSummary(ExtractSummary(e), agentName) {
				agentScoped = append(agentScoped, e)
			}
		}
	}

	// Pass 4: cap agent-scoped at 3 (most recent).
	if len(agentScoped) > 3 {
		agentScoped = agentScoped[len(agentScoped)-3:]
	}

	// Pass 5: merge and deduplicate.
	merged := make([]clewcontract.TypedEvent, 0, len(agentScoped)+len(recent))
	merged = append(merged, agentScoped...)
	merged = append(merged, recent...)

	// Pass 6: hard cap at 8 total (take most recent after sort by timestamp).
	// Since events are appended chronologically, merged is already roughly sorted.
	// We already know agentScoped precedes recent chronologically by construction.
	if len(merged) > 8 {
		merged = merged[len(merged)-8:]
	}

	return typedEventsToTimelineEntries(merged)
}

// agentNameInSummary returns true if agentName appears as a substring in summary.
// Case-sensitive substring match: SESSION-5 Section 4.2.
func agentNameInSummary(summary, agentName string) bool {
	if agentName == "" {
		return false
	}
	return strings.Contains(summary, agentName)
}

// typedEventsToTimelineEntries converts v3 TypedEvents to TimelineEntry values.
// Uses the same timestamp parsing as FormatTimelineEntry.
func typedEventsToTimelineEntries(events []clewcontract.TypedEvent) []TimelineEntry {
	if len(events) == 0 {
		return nil
	}

	entries := make([]TimelineEntry, 0, len(events))
	for _, e := range events {
		// Parse timestamp to extract HH:MM (same as FormatTimelineEntry).
		var t time.Time
		if ts, err := time.Parse("2006-01-02T15:04:05.000Z", e.Ts); err == nil {
			t = ts.UTC()
		} else if ts, err := time.Parse(time.RFC3339, e.Ts); err == nil {
			t = ts.UTC()
		}
		// t.Time portion is zero-dated, matching TimelineEntry.Time convention.
		parsed := time.Date(0, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)

		entries = append(entries, TimelineEntry{
			Time:     parsed,
			Category: EventTypeToCategory(e.Type), // already 8-char padded
			Summary:  ExtractSummary(e),
			Raw:      FormatTimelineEntry(e),
		})
	}

	return entries
}

// extractDecisions scans TypedEvents for decision.recorded events and returns them.
// Used only by orchestrator snapshots.
func extractDecisions(events []clewcontract.TypedEvent) []DecisionEntry {
	var decisions []DecisionEntry
	for _, e := range events {
		if e.Type != clewcontract.EventTypeDecisionRecorded {
			continue
		}

		var d clewcontract.DecisionRecordedData
		if err := json.Unmarshal(e.Data, &d); err != nil || d.Decision == "" {
			continue
		}

		// Parse timestamp for ordering (best-effort).
		var ts time.Time
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", e.Ts); err == nil {
			ts = t
		} else if t, err := time.Parse(time.RFC3339, e.Ts); err == nil {
			ts = t
		}

		decisions = append(decisions, DecisionEntry{
			Decision:  d.Decision,
			Rationale: d.Rationale,
			Timestamp: ts,
		})
	}
	return decisions
}

// extractBlockers parses the ## Blockers section from SESSION_CONTEXT.md body.
// Returns empty string when the section is absent, empty, or contains only "None"/"None yet.".
// Returns the section content verbatim (no truncation) when it has real blockers.
func extractBlockers(body string) string {
	lines := strings.Split(body, "\n")

	// Find ## Blockers section.
	start := -1
	for i, line := range lines {
		if strings.TrimRight(line, " \t") == "## Blockers" {
			start = i + 1
			break
		}
	}

	if start == -1 {
		return ""
	}

	// Collect lines until the next ## heading or EOF.
	var blockLines []string
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			break
		}
		blockLines = append(blockLines, lines[i])
	}

	// Join and trim whitespace.
	content := strings.TrimSpace(strings.Join(blockLines, "\n"))

	// Skip empty or "None" variants.
	if content == "" || content == "None" || content == "None yet." || content == "None yet" {
		return ""
	}

	return content
}
