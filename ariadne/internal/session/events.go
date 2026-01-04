package session

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
)

// EventType represents the type of session event.
type EventType string

const (
	EventSessionCreated    EventType = "SESSION_CREATED"
	EventSessionParked     EventType = "SESSION_PARKED"
	EventSessionResumed    EventType = "SESSION_RESUMED"
	EventSessionArchived   EventType = "SESSION_ARCHIVED"
	EventPhaseTransitioned EventType = "PHASE_TRANSITIONED"
	EventLockAcquired      EventType = "LOCK_ACQUIRED"
	EventLockReleased      EventType = "LOCK_RELEASED"
	EventSchemaMigrated    EventType = "SCHEMA_MIGRATED"
)

// Event represents a session lifecycle event.
type Event struct {
	Timestamp string                 `json:"timestamp"`
	Event     EventType              `json:"event"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	FromPhase string                 `json:"from_phase,omitempty"`
	ToPhase   string                 `json:"to_phase,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventEmitter handles event emission to JSONL files.
type EventEmitter struct {
	eventsPath string
	auditPath  string
}

// NewEventEmitter creates a new event emitter.
func NewEventEmitter(eventsPath, auditPath string) *EventEmitter {
	return &EventEmitter{
		eventsPath: eventsPath,
		auditPath:  auditPath,
	}
}

// Emit writes an event to the events file.
func (e *EventEmitter) Emit(event Event) error {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal event", err)
	}

	// Append to events file
	f, err := os.OpenFile(e.eventsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to open events file", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write event", err)
	}

	return nil
}

// EmitToAudit writes an event to the global audit log.
func (e *EventEmitter) EmitToAudit(sessionID string, event Event) error {
	if e.auditPath == "" {
		return nil
	}

	// Format: timestamp | session_id | event_type | from -> to
	transition := ""
	if event.From != "" || event.To != "" {
		transition = event.From + " -> " + event.To
	} else if event.FromPhase != "" || event.ToPhase != "" {
		transition = event.FromPhase + " -> " + event.ToPhase
	}

	line := event.Timestamp + " | " + sessionID + " | " + string(event.Event) + " | " + transition + "\n"

	f, err := os.OpenFile(e.auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to open audit file", err)
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write audit entry", err)
	}

	return nil
}

// EmitCreated emits a SESSION_CREATED event.
func (e *EventEmitter) EmitCreated(sessionID, initiative, complexity, team string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSessionCreated,
		From:      "NONE",
		To:        "ACTIVE",
		Metadata: map[string]interface{}{
			"initiative": initiative,
			"complexity": complexity,
			"team":       team,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitParked emits a SESSION_PARKED event.
func (e *EventEmitter) EmitParked(sessionID, reason, gitStatus string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSessionParked,
		From:      "ACTIVE",
		To:        "PARKED",
		Metadata: map[string]interface{}{
			"reason":     reason,
			"git_status": gitStatus,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitResumed emits a SESSION_RESUMED event.
func (e *EventEmitter) EmitResumed(sessionID string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSessionResumed,
		From:      "PARKED",
		To:        "ACTIVE",
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitArchived emits a SESSION_ARCHIVED event.
func (e *EventEmitter) EmitArchived(sessionID, fromStatus string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSessionArchived,
		From:      fromStatus,
		To:        "ARCHIVED",
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitPhaseTransition emits a PHASE_TRANSITIONED event.
func (e *EventEmitter) EmitPhaseTransition(sessionID, fromPhase, toPhase string, artifactsValidated bool) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventPhaseTransitioned,
		FromPhase: fromPhase,
		ToPhase:   toPhase,
		Metadata: map[string]interface{}{
			"artifacts_validated": artifactsValidated,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitLockAcquired emits a LOCK_ACQUIRED event.
func (e *EventEmitter) EmitLockAcquired(sessionID string, pid int) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventLockAcquired,
		Metadata: map[string]interface{}{
			"pid": pid,
		},
	}
	return e.Emit(event)
}

// EmitLockReleased emits a LOCK_RELEASED event.
func (e *EventEmitter) EmitLockReleased(sessionID string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventLockReleased,
	}
	return e.Emit(event)
}

// EmitSchemaMigrated emits a SCHEMA_MIGRATED event.
func (e *EventEmitter) EmitSchemaMigrated(sessionID, fromVersion, toVersion string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSchemaMigrated,
		Metadata: map[string]interface{}{
			"from_version": fromVersion,
			"to_version":   toVersion,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// ReadEvents reads events from a JSONL file.
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
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue // Skip malformed lines
		}
		events = append(events, event)
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
		if eventType != "" && string(e.Event) != eventType {
			continue
		}
		if !since.IsZero() {
			eventTime, err := time.Parse(time.RFC3339, e.Timestamp)
			if err != nil || eventTime.Before(since) {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	return filtered
}
