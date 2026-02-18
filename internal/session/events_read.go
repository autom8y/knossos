package session

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/autom8y/knossos/internal/errors"
)

// Dual Event Structs — ADR-0027 Backward Compatibility Bridge
//
// This file maintains two event structs (Event and ClewEvent) to support reading
// events.jsonl files that contain both pre-ADR-0027 (EventEmitter) and post-ADR-0027
// (clewcontract) format entries. The canonical write type is clewcontract.Event —
// all new event emission goes through clewcontract constructors and BufferedEventWriter.
//
// This dual-read capability exists solely for the `ari session audit` command, which
// must read historical event logs spanning the format migration. It is a deliberate
// ADR-0027 exception: the write path is fully unified, but the read path bridges both.
//
// Removal trigger: once all sessions created before ADR-0027 sprint 3 (commit 1a0e8f7)
// have been wrapped and archived, legacy Event entries will no longer appear in any
// active session's events.jsonl. At that point, the legacy Event struct and the
// format-sniffing logic in ReadEvents() can be removed.

// Event represents a session event in the pre-ADR-0027 EventEmitter format.
type Event struct {
	Timestamp string                 `json:"timestamp"`
	Event     string                 `json:"event"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	FromPhase string                 `json:"from_phase,omitempty"`
	ToPhase   string                 `json:"to_phase,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ClewEvent represents a Clew Contract v2 event (new format).
// This is a minimal struct for reading Clew events.jsonl entries.
type ClewEvent struct {
	Timestamp string                 `json:"ts"`
	Type      string                 `json:"type"`
	Tool      string                 `json:"tool,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Summary   string                 `json:"summary"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// ReadEvents reads events from a JSONL file, supporting both legacy and Clew formats.
// It attempts to parse each line as a legacy Event first, then as a ClewEvent.
// Returns normalized Event structs for backward compatibility with audit command.
func ReadEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to open events file", err)
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Try legacy format first
		var legacyEvent Event
		if err := json.Unmarshal(line, &legacyEvent); err == nil && legacyEvent.Event != "" {
			events = append(events, legacyEvent)
			continue
		}

		// Try Clew format
		var clewEvent ClewEvent
		if err := json.Unmarshal(line, &clewEvent); err == nil && clewEvent.Type != "" {
			// Convert Clew event to legacy format for audit compatibility
			events = append(events, Event{
				Timestamp: clewEvent.Timestamp,
				Event:     clewEvent.Type,
				Metadata: map[string]interface{}{
					"tool":    clewEvent.Tool,
					"path":    clewEvent.Path,
					"summary": clewEvent.Summary,
					"meta":    clewEvent.Meta,
				},
			})
			continue
		}

		// Skip malformed lines (matches previous behavior)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read events", err)
	}

	return events, nil
}

// FilterEvents filters events by type and/or timestamp.
func FilterEvents(events []Event, eventType string, since time.Time) []Event {
	var filtered []Event
	for _, e := range events {
		if eventType != "" && e.Event != eventType {
			continue
		}
		if !since.IsZero() {
			// Try RFC3339 (legacy format) first
			eventTime, err := time.Parse(time.RFC3339, e.Timestamp)
			if err != nil {
				// Try RFC3339 with milliseconds (Clew format)
				eventTime, err = time.Parse("2006-01-02T15:04:05.000Z", e.Timestamp)
				if err != nil || eventTime.Before(since) {
					continue
				}
			} else if eventTime.Before(since) {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	return filtered
}
