package clewcontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/autom8y/knossos/internal/errors"
)

// DefaultFlushInterval is the default interval for BufferedEventWriter flush operations.
// Events are buffered and written in batches at this interval for improved performance.
// Set to under 10 seconds per design requirements (acceptable loss window on crash).
const DefaultFlushInterval = 5 * time.Second

// EventsFileName is the standard name for the clew events log.
// Each session gets its own events.jsonl at .claude/sessions/<session-id>/events.jsonl,
// providing natural session-scoped event isolation without explicit rotation.
const EventsFileName = "events.jsonl"

// EventWriter provides thread-safe append-only JSONL writing for clew events.
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

// BufferedEventWriter provides async event writing with periodic flush.
// Events are buffered in memory and flushed to disk at configurable intervals.
// This provides better performance than synchronous writes while accepting
// a bounded data loss window (up to flushInterval) on crash.
//
// Usage:
//
//	w := NewBufferedEventWriter(sessionDir, 5*time.Second)
//	w.Write(event) // Non-blocking, returns immediately
//	// ... later
//	w.Close() // Ensures final flush before shutdown
type BufferedEventWriter struct {
	sessionDir    string
	events        []Event
	mu            sync.Mutex
	done          chan struct{}
	flushed       chan struct{} // Signals final flush complete
	closed        bool
	ticker        *time.Ticker
	flushInterval time.Duration
	flushErr      error // Last flush error, for diagnostic purposes
}

// NewBufferedEventWriter creates a writer that buffers events and flushes periodically.
// The flushInterval determines how often buffered events are written to disk.
// Use DefaultFlushInterval for the standard 5-second interval.
//
// The writer starts a background goroutine that must be stopped by calling Close().
func NewBufferedEventWriter(sessionDir string, flushInterval time.Duration) *BufferedEventWriter {
	w := &BufferedEventWriter{
		sessionDir:    sessionDir,
		events:        make([]Event, 0, 100),
		done:          make(chan struct{}),
		flushed:       make(chan struct{}),
		flushInterval: flushInterval,
	}
	w.ticker = time.NewTicker(flushInterval)
	go w.flushLoop()
	return w
}

// flushLoop runs in a background goroutine, periodically flushing buffered events.
func (w *BufferedEventWriter) flushLoop() {
	for {
		select {
		case <-w.ticker.C:
			_ = w.Flush() // Errors are stored in w.flushErr for diagnostics
		case <-w.done:
			w.ticker.Stop()
			_ = w.Flush() // Final flush on close
			close(w.flushed)
			return
		}
	}
}

// Write buffers an event for async flush. This method is non-blocking and returns immediately.
// Thread-safe: multiple goroutines can call Write concurrently.
func (w *BufferedEventWriter) Write(event Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return // Silently drop writes after close
	}

	w.events = append(w.events, event)
}

// Flush writes all buffered events to disk.
// This is called automatically by the background goroutine, but can also be called
// manually if immediate persistence is required.
// Thread-safe: can be called concurrently with Write.
func (w *BufferedEventWriter) Flush() error {
	w.mu.Lock()
	if len(w.events) == 0 {
		w.mu.Unlock()
		return nil
	}

	// Swap buffer to minimize lock hold time
	toFlush := w.events
	w.events = make([]Event, 0, 100)
	w.mu.Unlock()

	// Use existing EventWriter for atomic batch write
	syncWriter, err := NewEventWriter(w.sessionDir)
	if err != nil {
		w.mu.Lock()
		w.flushErr = err
		// Re-queue events that failed to flush (prepend to preserve order)
		w.events = append(toFlush, w.events...)
		w.mu.Unlock()
		return err
	}
	defer syncWriter.Close()

	err = syncWriter.WriteMultiple(toFlush)
	if err != nil {
		w.mu.Lock()
		w.flushErr = err
		// Re-queue events that failed to flush
		w.events = append(toFlush, w.events...)
		w.mu.Unlock()
	}
	return err
}

// Close stops the background flush goroutine and performs a final flush.
// After Close returns, all buffered events will have been written to disk
// (or an error will be returned if the final flush failed).
// Close is idempotent - calling it multiple times is safe.
func (w *BufferedEventWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	close(w.done)

	// Wait for the final flush to complete
	<-w.flushed

	// Return any flush error that occurred
	w.mu.Lock()
	err := w.flushErr
	w.mu.Unlock()
	return err
}

// Path returns the path to the events file.
func (w *BufferedEventWriter) Path() string {
	return filepath.Join(w.sessionDir, EventsFileName)
}

// FlushError returns the last error from a background flush operation, if any.
// This is useful for diagnostics when async writes may have failed.
func (w *BufferedEventWriter) FlushError() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushErr
}

// Len returns the number of events currently buffered (not yet flushed).
// Useful for testing and diagnostics.
func (w *BufferedEventWriter) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.events)
}
