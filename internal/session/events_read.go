package session

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
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

// typedClewEventDetector is a minimal struct used to detect v3 TypedEvent lines.
// Detection rule per SESSION-1 spec Section 5.1: presence of "data" field -> v3.
// We use json.RawMessage so we can check for the field without full parsing.
type typedClewEventDetector struct {
	Data json.RawMessage `json:"data"`
}

// ReadEvents reads events from a JSONL file, supporting v1 legacy, v2 flat, and v3 typed formats.
// All three formats may be interleaved in the same file.
//
// Format detection per SESSION-1 spec Section 5.1:
//   - JSON has "data" field -> v3 TypedEvent (highest precedence)
//   - JSON has "event" field with non-empty value -> v1 legacy Event
//   - JSON has "type" field with non-empty value -> v2 flat ClewEvent
//   - Otherwise -> malformed line, skipped
//
// Returns normalized Event structs for backward compatibility with the audit command.
func ReadEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to open events file", err)
	}
	defer func() { _ = f.Close() }()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()

		// v3 detection: check for "data" field first (highest precedence per spec).
		// A line with "data" is a TypedEvent regardless of whether "meta" is also present.
		var detector typedClewEventDetector
		if err := json.Unmarshal(line, &detector); err == nil && detector.Data != nil {
			// Parse as full v3 TypedEvent and normalize to audit-compatible Event.
			var te clewcontract.TypedEvent
			if err := json.Unmarshal(line, &te); err == nil && string(te.Type) != "" {
				events = append(events, typedEventToLegacy(te))
				continue
			}
		}

		// v1 detection: "event" field with non-empty string value.
		var legacyEvent Event
		if err := json.Unmarshal(line, &legacyEvent); err == nil && legacyEvent.Event != "" {
			events = append(events, legacyEvent)
			continue
		}

		// v2 detection: "type" field with non-empty string value.
		var clewEvent ClewEvent
		if err := json.Unmarshal(line, &clewEvent); err == nil && clewEvent.Type != "" {
			// Convert v2 Clew event to legacy format for audit compatibility.
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

		// Skip malformed lines (matches previous behavior).
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read events", err)
	}

	return events, nil
}

// typedEventToLegacy converts a v3 TypedEvent to a normalized legacy Event struct.
// This provides backward compatibility for the audit command which expects the legacy format.
//
// Conversion rules per SESSION-1 spec Section 5.3:
//   - type: apply v2->v3 rename map (RenameV2Type is a no-op on already-renamed types)
//   - source: stored in Metadata["source"]
//   - data: stored in Metadata["data"] as the parsed JSON object
func typedEventToLegacy(te clewcontract.TypedEvent) Event {
	// Parse the Data field to include in Metadata.
	var dataMap map[string]interface{}
	_ = json.Unmarshal(te.Data, &dataMap)

	// Apply rename: v3 type strings are already canonical; this is a no-op for v3 events
	// but ensures consistency if a hybrid event is encountered.
	normalizedType := clewcontract.RenameV2Type(string(te.Type))

	return Event{
		Timestamp: te.Ts,
		Event:     normalizedType,
		Metadata: map[string]interface{}{
			"source": string(te.Source),
			"data":   dataMap,
		},
	}
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
