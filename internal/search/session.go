package search

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// SessionSignals holds session state used for search scoring.
// Nil SessionSignals means no active session (no-op for all modifiers).
type SessionSignals struct {
	SessionID  string
	Phase      string           // e.g., "requirements", "design", "implementation", "validation"
	Rite       string           // active rite from session
	Complexity string           // "TASK", "MODULE", "INITIATIVE"
	Initiative string
	Activity   *ActivitySummary // nil if events.jsonl missing or empty
}

// ActivitySummary holds aggregated event counts from events.jsonl tail read.
type ActivitySummary struct {
	FileChangeCount  int            // count of tool.file_change events
	AgentTasks       map[string]int // agent name -> count of agent.task_start events
	PhaseTransitions int            // count of phase.transitioned events
	LastEventTS      string         // RFC3339 timestamp of most recent event
}

// tailBufferSize is the maximum bytes to read from the tail of events.jsonl.
// 50 events * ~200 bytes/event = 10,000 bytes. Round up for safety.
const tailBufferSize = 12 * 1024 // 12KB

// defaultTailLineCap is the default maximum number of lines to parse from the tail.
var defaultTailLineCap = 50

// tailEvent is a minimal struct for parsing events.jsonl tail lines.
// Only extracts the fields needed for activity signal computation.
type tailEvent struct {
	Ts   string        `json:"ts"`
	Type string        `json:"type"`
	Meta tailEventMeta `json:"meta,omitempty"`
}

type tailEventMeta struct {
	Agent string `json:"agent,omitempty"`
}

// tailEventLegacy is a fallback struct for v1 legacy events.
type tailEventLegacy struct {
	Ts    string `json:"timestamp"`
	Event string `json:"event"`
}

// Score modifier constants.
const (
	// phaseBoostAmount is the score added when an entry matches the current phase.
	phaseBoostAmount = 150

	// activityBoostAmount is the score added when recent activity correlates with an entry.
	activityBoostAmount = 75

	// complexityPenaltyAmount is the score subtracted from heavyweight routing
	// entries in TASK-complexity sessions.
	complexityPenaltyAmount = 100

	// activityFileChangeThreshold is the minimum file_change count to trigger
	// implementation activity boost.
	activityFileChangeThreshold = 5

	// tierFloorExact is the minimum score for exact matches (Tier 1 protection).
	tierFloorExact = 900

	// tierFloorPrefix is the minimum score for prefix matches (Tier 2 protection).
	tierFloorPrefix = 500
)

// phaseBoostKeywords maps session phases to keywords that trigger boosting.
// Entries matching any keyword for the current phase receive +phaseBoostAmount.
// This is a package-level var (not const) for test override per FR-11.
var phaseBoostKeywords = map[string][]string{
	"requirements": {
		"requirements-analyst", "potnia", "prd", "requirements",
		"gather", "scope", "acceptance",
	},
	"design": {
		"architect", "tdd", "adr", "design", "technical",
		"architecture", "interface", "contract",
	},
	"implementation": {
		"principal-engineer", "build", "test", "implement",
		"code", "compile", "edit", "write",
	},
	"validation": {
		"qa-adversary", "test", "sails", "validate",
		"coverage", "adversarial", "quality",
	},
}

// complexityPenaltyKeywords are keywords in routing-domain entries that
// receive a score penalty when session complexity is TASK.
var complexityPenaltyKeywords = []string{
	"orchestration", "multi-agent", "initiative",
	"coordinate", "multi-phase",
}

// ReadSessionState reads session state for scoring from SESSION_CONTEXT.md.
// Returns nil on any error (fail-open: malformed YAML, missing file, I/O error).
// Does not acquire locks -- read-only access to a file that changes infrequently.
func ReadSessionState(contextPath string) *SessionSignals {
	content, err := os.ReadFile(contextPath)
	if err != nil {
		return nil
	}

	ctx, err := session.ParseContext(content)
	if err != nil {
		return nil
	}

	signals := &SessionSignals{
		SessionID:  ctx.SessionID,
		Phase:      ctx.CurrentPhase,
		Complexity: ctx.Complexity,
		Initiative: ctx.Initiative,
	}

	// Prefer the explicit Rite field; fall back to ActiveRite.
	if ctx.Rite != nil {
		signals.Rite = *ctx.Rite
	} else {
		signals.Rite = ctx.ActiveRite
	}

	return signals
}

// TailReadEvents reads the last N events from events.jsonl using seek-to-end.
// Returns nil on any error (fail-open). Allocates at most tailBufferSize bytes.
// lineCap controls the maximum number of parsed lines (default 50).
func TailReadEvents(eventsPath string, lineCap int) *ActivitySummary {
	f, err := os.Open(eventsPath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return nil
	}

	size := stat.Size()
	if size == 0 {
		return nil
	}

	// Determine read offset and buffer size.
	readSize := size
	offset := int64(0)
	if size > int64(tailBufferSize) {
		offset = size - int64(tailBufferSize)
		readSize = int64(tailBufferSize)
	}

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return nil
		}
	}

	buf := make([]byte, readSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil
	}
	buf = buf[:n]

	// Split into lines.
	rawLines := bytes.Split(buf, []byte{'\n'})

	// If we seeked past the beginning, discard the first line (likely partial).
	if offset > 0 && len(rawLines) > 0 {
		rawLines = rawLines[1:]
	}

	// Remove trailing empty line from final newline.
	if len(rawLines) > 0 && len(rawLines[len(rawLines)-1]) == 0 {
		rawLines = rawLines[:len(rawLines)-1]
	}

	// Take only the last lineCap lines.
	if lineCap <= 0 {
		lineCap = defaultTailLineCap
	}
	if len(rawLines) > lineCap {
		rawLines = rawLines[len(rawLines)-lineCap:]
	}

	// Parse events and aggregate.
	summary := &ActivitySummary{
		AgentTasks: make(map[string]int),
	}
	parsed := 0

	for _, line := range rawLines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var evt tailEvent
		if err := json.Unmarshal(line, &evt); err != nil {
			continue // skip unparseable lines
		}

		// If primary parse succeeded but Type is empty, try legacy format
		// (v1 events use "event" and "timestamp" instead of "type" and "ts").
		if evt.Type == "" {
			var legacy tailEventLegacy
			if err := json.Unmarshal(line, &legacy); err == nil && legacy.Event != "" {
				evt.Type = legacy.Event
				if legacy.Ts != "" {
					evt.Ts = legacy.Ts
				}
			}
		}

		if evt.Type == "" {
			continue
		}

		parsed++

		// Track last event timestamp.
		if evt.Ts != "" {
			summary.LastEventTS = evt.Ts
		}

		// Aggregate by event type.
		switch evt.Type {
		case "tool.file_change":
			summary.FileChangeCount++
		case "agent.task_start":
			if evt.Meta.Agent != "" {
				summary.AgentTasks[evt.Meta.Agent]++
			}
		case "phase.transitioned":
			summary.PhaseTransitions++
		}
	}

	if parsed == 0 {
		return nil
	}

	return summary
}

// entryMatchesKeywords checks if an entry's name, keywords, or summary
// contain any of the provided keywords. Used by phase and activity boost logic.
func entryMatchesKeywords(entry SearchEntry, keywords []string) bool {
	nameLower := strings.ToLower(entry.Name)
	summaryLower := strings.ToLower(entry.Summary)

	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)

		if strings.Contains(nameLower, kwLower) {
			return true
		}

		for _, entryKW := range entry.Keywords {
			if strings.Contains(strings.ToLower(entryKW), kwLower) {
				return true
			}
		}

		if strings.Contains(summaryLower, kwLower) {
			return true
		}
	}

	return false
}

// sessionScoreModifier computes the additive score adjustment for an entry
// given session signals. Returns 0 when signals is nil (no-session case).
//
// The modifier applies three independent adjustments:
//  1. Phase boost:       +150 if entry keywords/name match current phase
//  2. Activity boost:    +75  if recent events correlate with entry category
//  3. Complexity penalty: -100 if TASK session + heavyweight routing entry
//
// Phase and activity boosts targeting the same entry use max(phase, activity),
// not sum, to prevent double-boosting (PRD EC-13).
//
// Tier floor protection: the modifier returns a value that, when added to
// baseScore, will not reduce it below tierFloorExact (for exact matches) or
// tierFloorPrefix (for prefix matches).
func sessionScoreModifier(entry SearchEntry, matchType string, baseScore int, signals *SessionSignals) int {
	if signals == nil {
		return 0
	}

	// 1. Compute phase boost.
	phaseAdj := 0
	if keywords, ok := phaseBoostKeywords[signals.Phase]; ok {
		if entryMatchesKeywords(entry, keywords) {
			phaseAdj = phaseBoostAmount
		}
	}

	// 2. Compute activity boost.
	activityAdj := 0
	if signals.Activity != nil {
		if signals.Activity.FileChangeCount > activityFileChangeThreshold {
			if implKeywords, ok := phaseBoostKeywords["implementation"]; ok {
				if entryMatchesKeywords(entry, implKeywords) {
					activityAdj = activityBoostAmount
				}
			}
		}

		if signals.Activity.AgentTasks["qa-adversary"] > 0 {
			if valKeywords, ok := phaseBoostKeywords["validation"]; ok {
				if entryMatchesKeywords(entry, valKeywords) {
					activityAdj = max(activityAdj, activityBoostAmount)
				}
			}
		}

		if signals.Activity.AgentTasks["architect"] > 0 {
			if designKeywords, ok := phaseBoostKeywords["design"]; ok {
				if entryMatchesKeywords(entry, designKeywords) {
					activityAdj = max(activityAdj, activityBoostAmount)
				}
			}
		}
	}

	// 3. Combine phase + activity: max, not sum (EC-13).
	positiveAdj := max(phaseAdj, activityAdj)

	// 4. Compute complexity penalty.
	penalty := 0
	if signals.Complexity == "TASK" && entry.Domain == DomainRouting {
		nameLower := strings.ToLower(entry.Name)
		summaryLower := strings.ToLower(entry.Summary)
		for _, kw := range complexityPenaltyKeywords {
			kwLower := strings.ToLower(kw)
			if strings.Contains(nameLower, kwLower) || strings.Contains(summaryLower, kwLower) {
				penalty = complexityPenaltyAmount
				break
			}
			for _, entryKW := range entry.Keywords {
				if strings.Contains(strings.ToLower(entryKW), kwLower) {
					penalty = complexityPenaltyAmount
					break
				}
			}
			if penalty > 0 {
				break
			}
		}
	}

	// 5. Raw modifier.
	rawModifier := positiveAdj - penalty

	// 6. Tier floor protection.
	switch matchType {
	case "exact":
		floor := tierFloorExact - baseScore
		if rawModifier < floor {
			return floor
		}
	case "prefix":
		floor := tierFloorPrefix - baseScore
		if rawModifier < floor {
			return floor
		}
	}

	return rawModifier
}

// CollectParkedSessions scans .sos/sessions/ for PARKED sessions and returns
// SearchEntry items for each. Returns nil if resolver is nil, has no project root,
// or the sessions directory is unreadable (fail-open).
func CollectParkedSessions(resolver *paths.Resolver) []SearchEntry {
	if resolver == nil || resolver.ProjectRoot() == "" {
		return nil
	}

	sessionsDir := resolver.SessionsDir()
	dirEntries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil
	}

	var entries []SearchEntry
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}

		name := de.Name()
		// Session dirs start with "session-" and are 32+ chars.
		if !paths.IsSessionDir(name) {
			continue
		}

		contextPath := filepath.Join(sessionsDir, name, "SESSION_CONTEXT.md")
		content, err := os.ReadFile(contextPath)
		if err != nil {
			continue
		}

		ctx, err := session.ParseContext(content)
		if err != nil {
			continue
		}

		if ctx.Status != session.StatusParked {
			continue
		}

		// Build entry name: prefer initiative, fall back to session ID.
		entryName := ctx.Initiative
		if entryName == "" {
			entryName = ctx.SessionID
		}

		// Build summary.
		summary := "Parked session on " + ctx.ActiveRite + ": " + ctx.Initiative
		if ctx.ParkedReason != "" {
			summary += " (parked: " + ctx.ParkedReason + ")"
		}
		if ctx.Initiative == "" {
			summary = "Parked session (unknown initiative)"
		}

		// Build keywords from initiative text.
		kw := tokenize(ctx.Initiative)
		kw = append(kw, ctx.ActiveRite, "parked", "deferred")

		entries = append(entries, SearchEntry{
			Name:    entryName,
			Domain:  DomainSession,
			Summary: summary,
			Action:  "ari session resume --session-id=" + ctx.SessionID,
			Keywords: kw,
		})
	}

	return entries
}
