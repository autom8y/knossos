package threadcontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/autom8y/ariadne/internal/errors"
)

// EventsFileName is the standard name for the thread events log.
const EventsFileName = "events.jsonl"

// EventWriter provides thread-safe append-only JSONL writing for thread events.
type EventWriter struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
}

// NewEventWriter creates a new EventWriter for the given session directory.
// The events.jsonl file will be created in sessionDir if it doesn't exist.
func NewEventWriter(sessionDir string) (*EventWriter, error) {
	filePath := filepath.Join(sessionDir, EventsFileName)

	// Ensure session directory exists
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, errors.Wrap(errors.CodePermissionDenied, "failed to create session directory", err)
	}

	return &EventWriter{
		filePath: filePath,
	}, nil
}

// NewEventWriterPath creates a new EventWriter for an explicit file path.
// Use this when you already have the full path to events.jsonl.
func NewEventWriterPath(filePath string) (*EventWriter, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.Wrap(errors.CodePermissionDenied, "failed to create events directory", err)
	}

	return &EventWriter{
		filePath: filePath,
	}, nil
}

// Write appends an event to the JSONL file.
// Thread-safe: multiple goroutines can call Write concurrently.
func (w *EventWriter) Write(event Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal event", err)
	}

	// Open file for append (create if needed)
	f, err := os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to open events file", err)
	}
	defer f.Close()

	// Write JSON line
	if _, err := f.Write(append(data, '\n')); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write event", err)
	}

	return nil
}

// WriteMultiple appends multiple events atomically.
// All events are written in a single lock acquisition for better performance.
func (w *EventWriter) WriteMultiple(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Open file for append (create if needed)
	f, err := os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to open events file", err)
	}
	defer f.Close()

	// Write all events
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return errors.Wrap(errors.CodeGeneralError, "failed to marshal event", err)
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			return errors.Wrap(errors.CodeGeneralError, "failed to write event", err)
		}
	}

	return nil
}

// Close releases any resources held by the EventWriter.
// Currently a no-op since we open/close per write for safety.
func (w *EventWriter) Close() error {
	// No persistent file handle to close
	// Future: could add buffered writer with periodic flush
	return nil
}

// Path returns the path to the events file.
func (w *EventWriter) Path() string {
	return w.filePath
}
