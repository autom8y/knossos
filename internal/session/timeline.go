package session

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
)

// timelineSection is the heading used for the timeline section in SESSION_CONTEXT.md.
const timelineSection = "## Timeline"

// TimelineEntry is a parsed timeline entry from SESSION_CONTEXT.md.
type TimelineEntry struct {
	// Time is the parsed timestamp from the HH:MM field (date is set to zero).
	Time time.Time
	// Category is the 8-char padded category string (e.g., "SESSION ", "PHASE   ").
	Category string
	// Summary is the free-form summary text (already truncated to 80 chars max).
	Summary string
	// Raw is the original unmodified line.
	Raw string
}

// IsCuratedType returns true if the event type projects to the timeline.
// Only the 11 canonical curated types defined in SESSION-2 Section 1.3 are curated.
func IsCuratedType(eventType clewcontract.EventType) bool {
	switch eventType {
	case clewcontract.EventTypeSessionCreated,
		clewcontract.EventTypeSessionParked,
		clewcontract.EventTypeSessionResumed,
		clewcontract.EventTypeSessionWrapped,
		clewcontract.EventTypeSessionFrayed,
		clewcontract.EventTypePhaseTransitioned,
		clewcontract.EventTypeAgentDelegated,
		clewcontract.EventTypeAgentCompleted,
		clewcontract.EventTypeCommitCreated,
		clewcontract.EventTypeDecisionRecorded,
		clewcontract.EventTypeCommandInvoked:
		return true
	default:
		return false
	}
}

// EventTypeToCategory maps a v3 event type to its 8-char padded timeline category string.
// The format string "%-8s" left-aligns and pads to exactly 8 characters.
// Falls through to "NOTE" for any unrecognized type.
func EventTypeToCategory(eventType clewcontract.EventType) string {
	var raw string
	switch eventType {
	case clewcontract.EventTypeSessionCreated,
		clewcontract.EventTypeSessionParked,
		clewcontract.EventTypeSessionResumed,
		clewcontract.EventTypeSessionWrapped,
		clewcontract.EventTypeSessionFrayed:
		raw = "SESSION"
	case clewcontract.EventTypePhaseTransitioned:
		raw = "PHASE"
	case clewcontract.EventTypeAgentDelegated,
		clewcontract.EventTypeAgentCompleted:
		raw = "AGENT"
	case clewcontract.EventTypeCommitCreated:
		raw = "COMMIT"
	case clewcontract.EventTypeDecisionRecorded:
		raw = "DECISION"
	case clewcontract.EventTypeCommandInvoked:
		raw = "COMMAND"
	default:
		raw = "NOTE"
	}
	// Left-align, space-pad to exactly 8 characters per spec Section 2.2.
	return fmt.Sprintf("%-8s", raw)
}

// truncateSummary truncates s to maxLen characters, appending "..." if truncated.
// If len(s) <= maxLen, s is returned unchanged.
func truncateSummary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Truncate to maxLen-3 and append "..."
	return s[:maxLen-3] + "..."
}

// ExtractSummary produces the summary text for a v3 TypedEvent.
// Template definitions from SESSION-2 Section 3.1.
func ExtractSummary(event clewcontract.TypedEvent) string {
	var summary string

	switch event.Type {
	case clewcontract.EventTypeSessionCreated:
		var d clewcontract.SessionCreatedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			summary = "created: (unknown)"
		} else {
			switch {
			case d.Complexity != "":
				summary = fmt.Sprintf("created: %s (%s)", d.Initiative, d.Complexity)
			case d.Initiative != "":
				summary = fmt.Sprintf("created: %s", d.Initiative)
			default:
				summary = "created: (unknown)"
			}
		}

	case clewcontract.EventTypeSessionParked:
		var d clewcontract.SessionParkedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.Reason == "" {
			summary = "parked"
		} else {
			summary = fmt.Sprintf("parked: %s", d.Reason)
		}

	case clewcontract.EventTypeSessionResumed:
		summary = "resumed"

	case clewcontract.EventTypeSessionWrapped:
		var d clewcontract.SessionWrappedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.SailsColor == "" {
			summary = "wrapped"
		} else {
			summary = fmt.Sprintf("wrapped (%s)", d.SailsColor)
		}

	case clewcontract.EventTypeSessionFrayed:
		var d clewcontract.SessionFrayedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.ChildID == "" {
			summary = "frayed"
		} else {
			summary = fmt.Sprintf("frayed -> %s", d.ChildID)
		}

	case clewcontract.EventTypePhaseTransitioned:
		var d clewcontract.PhaseTransitionedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			summary = "phase transition"
		} else {
			summary = fmt.Sprintf("%s -> %s", d.From, d.To)
		}

	case clewcontract.EventTypeAgentDelegated:
		var d clewcontract.AgentDelegatedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.AgentName == "" {
			summary = "delegated"
		} else if d.TaskID != "" {
			summary = fmt.Sprintf("delegated %s: %s", d.AgentName, d.TaskID)
		} else {
			summary = fmt.Sprintf("delegated %s", d.AgentName)
		}

	case clewcontract.EventTypeAgentCompleted:
		var d clewcontract.AgentCompletedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.AgentName == "" {
			summary = "completed"
		} else if d.Outcome != "" {
			summary = fmt.Sprintf("completed %s: %s", d.AgentName, d.Outcome)
		} else {
			summary = fmt.Sprintf("completed %s", d.AgentName)
		}

	case clewcontract.EventTypeCommitCreated:
		var d clewcontract.CommitCreatedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.SHA == "" {
			summary = "commit"
		} else {
			// Always use the first 7 characters of SHA (spec Section 3.2 rule 3).
			sha7 := d.SHA
			if len(sha7) > 7 {
				sha7 = sha7[:7]
			}
			summary = fmt.Sprintf("%s: %s", sha7, d.Message)
		}

	case clewcontract.EventTypeDecisionRecorded:
		var d clewcontract.DecisionRecordedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.Decision == "" {
			summary = "decision"
		} else if d.Rationale != "" {
			// Rationale shortened to 30 chars per spec Section 3.2 rule 4.
			rationale := truncateSummary(d.Rationale, 30)
			summary = fmt.Sprintf("%s (%s)", d.Decision, rationale)
		} else {
			summary = d.Decision
		}

	case clewcontract.EventTypeCommandInvoked:
		var d clewcontract.CommandInvokedData
		if err := json.Unmarshal(event.Data, &d); err != nil || d.Command == "" {
			summary = "command"
		} else {
			summary = d.Command
		}

	default:
		// Fallback: attempt to extract a "message" field from raw Data.
		var raw map[string]any
		if err := json.Unmarshal(event.Data, &raw); err == nil {
			if msg, ok := raw["message"].(string); ok && msg != "" {
				summary = msg
			} else if msg, ok := raw["summary"].(string); ok && msg != "" {
				summary = msg
			}
		}
		if summary == "" {
			summary = string(event.Type)
		}
	}

	return truncateSummary(summary, 80)
}

// FormatTimelineEntry converts a TypedEvent into a formatted timeline entry string.
// Format: "- HH:MM | CATEGORY | summary" per SESSION-2 Section 2.1.
//
// The canonical Go format string is: fmt.Sprintf("- %s | %-8s | %s", timeHHMM, category, summary)
func FormatTimelineEntry(event clewcontract.TypedEvent) string {
	// Parse timestamp - extract HH:MM in UTC.
	timeHHMM := "00:00"
	if ts, err := time.Parse("2006-01-02T15:04:05.000Z", event.Ts); err == nil {
		timeHHMM = ts.UTC().Format("15:04")
	} else if ts, err := time.Parse(time.RFC3339, event.Ts); err == nil {
		// Defensive: also handle RFC3339 (which may have timezone offset).
		timeHHMM = ts.UTC().Format("15:04")
	}

	// Derive category (already padded to 8 chars by EventTypeToCategory).
	category := EventTypeToCategory(event.Type)

	// Extract summary (already truncated to 80 chars by ExtractSummary).
	summary := ExtractSummary(event)

	return fmt.Sprintf("- %s | %s | %s", timeHHMM, category, summary)
}

// AppendEntry appends a formatted timeline entry to the ## Timeline section of contextPath.
//
// Algorithm:
//  1. Read the full file.
//  2. Locate the ## Timeline section.
//  3. Find the boundary of the section (next ## heading or EOF).
//  4. Insert entry before that boundary.
//  5. Write back atomically.
//
// If the ## Timeline section does not exist, it is appended at the end of the file.
// This function does NOT acquire any lock -- the caller is responsible for
// accepting last-writer-wins semantics (see SESSION-2 Section 7.3).
func AppendEntry(contextPath string, entry string) error {
	content, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New(errors.CodeFileNotFound, "SESSION_CONTEXT.md not found: "+contextPath)
		}
		return errors.Wrap(errors.CodeGeneralError, "failed to read SESSION_CONTEXT.md", err)
	}

	body := string(content)
	newBody, err := insertIntoTimelineSection(body, entry)
	if err != nil {
		return err
	}

	if err := fileutil.AtomicWriteFile(contextPath, []byte(newBody), 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write timeline entry", err)
	}
	return nil
}

// insertIntoTimelineSection finds the ## Timeline section in body and appends entry to it.
// Returns the updated body string.
func insertIntoTimelineSection(body string, entry string) (string, error) {
	lines := strings.Split(body, "\n")

	// Find the ## Timeline section.
	timelineLineIdx := -1
	for i, line := range lines {
		if strings.TrimRight(line, " \t") == timelineSection {
			timelineLineIdx = i
			break
		}
	}

	if timelineLineIdx == -1 {
		// ## Timeline section not found: append it at the end.
		// Ensure there's a blank line before the section.
		result := strings.TrimRight(body, "\n")
		result += "\n\n" + timelineSection + "\n" + entry + "\n"
		return result, nil
	}

	// Find the end of the ## Timeline section: the next ## heading or EOF.
	insertIdx := len(lines) // default: insert at EOF
	for i := timelineLineIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			insertIdx = i
			break
		}
	}

	// Build the new lines by inserting entry before the boundary.
	// Entry goes immediately before insertIdx (at end of timeline section content).
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, entry)
	newLines = append(newLines, lines[insertIdx:]...)

	return strings.Join(newLines, "\n"), nil
}

// ReadTimeline reads and parses all timeline entries from the ## Timeline section.
// Returns an empty slice (not an error) if the section is missing or empty.
func ReadTimeline(contextPath string) ([]TimelineEntry, error) {
	content, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.CodeFileNotFound, "SESSION_CONTEXT.md not found: "+contextPath)
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read SESSION_CONTEXT.md", err)
	}

	return parseTimelineEntries(string(content)), nil
}

// parseTimelineEntries extracts timeline entry lines from a full SESSION_CONTEXT.md body.
// An entry line matches the pattern "- HH:MM | ...".
func parseTimelineEntries(body string) []TimelineEntry {
	lines := strings.Split(body, "\n")

	// Find the ## Timeline section boundaries.
	start := -1
	end := len(lines)
	for i, line := range lines {
		if strings.TrimRight(line, " \t") == timelineSection {
			start = i + 1
			continue
		}
		if start > 0 && strings.HasPrefix(line, "## ") {
			end = i
			break
		}
	}

	if start == -1 {
		return nil
	}

	var entries []TimelineEntry
	for _, line := range lines[start:end] {
		entry, ok := parseTimelineEntry(line)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// parseTimelineEntry attempts to parse a single timeline entry line.
// Returns (entry, true) on success or (zero, false) if the line is not a valid entry.
func parseTimelineEntry(line string) (TimelineEntry, bool) {
	// Entry format: "- HH:MM | CATEGORY | summary"
	// Minimum valid line: "- HH:MM | NOTE     | x" = 23 chars
	if len(line) < 23 {
		return TimelineEntry{}, false
	}
	if !strings.HasPrefix(line, "- ") {
		return TimelineEntry{}, false
	}

	rest := line[2:] // strip "- "
	// Expected: "HH:MM | ..."
	if len(rest) < 5 || rest[2] != ':' {
		return TimelineEntry{}, false
	}

	// Parse HH:MM
	hhmmStr := rest[:5]
	t, err := time.Parse("15:04", hhmmStr)
	if err != nil {
		return TimelineEntry{}, false
	}

	// Check separator " | "
	if len(rest) < 9 || rest[5:8] != " | " {
		return TimelineEntry{}, false
	}

	afterTime := rest[8:] // "CATEGORY | summary"
	// Category is exactly 8 chars followed by " | "
	if len(afterTime) < 11 {
		return TimelineEntry{}, false
	}
	category := afterTime[:8]
	if afterTime[8:11] != " | " {
		return TimelineEntry{}, false
	}

	summary := afterTime[11:]

	return TimelineEntry{
		Time:     t,
		Category: category,
		Summary:  summary,
		Raw:      line,
	}, true
}
