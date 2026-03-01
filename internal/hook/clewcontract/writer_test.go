package clewcontract

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewEventWriter(t *testing.T) {
	tmpDir := t.TempDir()

	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	expectedPath := filepath.Join(tmpDir, EventsFileName)
	if writer.Path() != expectedPath {
		t.Errorf("Path() = %v, want %v", writer.Path(), expectedPath)
	}
}

func TestNewEventWriter_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "session")

	writer, err := NewEventWriter(nestedDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	// Directory should exist
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("NewEventWriter should create session directory")
	}
}

func TestNewEventWriterPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom", "events.jsonl")

	writer, err := NewEventWriterPath(customPath)
	if err != nil {
		t.Fatalf("NewEventWriterPath failed: %v", err)
	}
	defer writer.Close()

	if writer.Path() != customPath {
		t.Errorf("Path() = %v, want %v", writer.Path(), customPath)
	}

	// Parent directory should exist
	if _, err := os.Stat(filepath.Dir(customPath)); os.IsNotExist(err) {
		t.Error("NewEventWriterPath should create parent directory")
	}
}

func TestEventWriter_Write(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	event := Event{
		Timestamp: "2024-01-04T10:23:45.123Z",
		Type:      EventTypeToolCall,
		Tool:      "Bash",
		Summary:   "Test event",
	}

	if err := writer.Write(event); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back the file
	data, err := os.ReadFile(writer.Path())
	if err != nil {
		t.Fatalf("Failed to read events file: %v", err)
	}

	// Should end with newline
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("Event line should end with newline")
	}

	// Parse the JSON
	var parsed Event
	if err := json.Unmarshal(data[:len(data)-1], &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed.Type != EventTypeToolCall {
		t.Errorf("Type = %v, want %v", parsed.Type, EventTypeToolCall)
	}
	if parsed.Tool != "Bash" {
		t.Errorf("Tool = %v, want Bash", parsed.Tool)
	}
}

func TestEventWriter_AppendBehavior(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	// Write multiple events
	events := []Event{
		{Timestamp: "2024-01-04T10:00:00.000Z", Type: EventTypeToolCall, Tool: "Read", Summary: "First"},
		{Timestamp: "2024-01-04T10:01:00.000Z", Type: EventTypeToolCall, Tool: "Edit", Summary: "Second"},
		{Timestamp: "2024-01-04T10:02:00.000Z", Type: EventTypeToolCall, Tool: "Write", Summary: "Third"},
	}

	for _, e := range events {
		if err := writer.Write(e); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	// Read all lines
	file, err := os.Open(writer.Path())
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	var readEvents []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("Failed to parse line: %v", err)
		}
		readEvents = append(readEvents, e)
	}

	if len(readEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(readEvents))
	}

	// Verify order
	if readEvents[0].Summary != "First" {
		t.Errorf("First event summary = %v, want First", readEvents[0].Summary)
	}
	if readEvents[1].Summary != "Second" {
		t.Errorf("Second event summary = %v, want Second", readEvents[1].Summary)
	}
	if readEvents[2].Summary != "Third" {
		t.Errorf("Third event summary = %v, want Third", readEvents[2].Summary)
	}
}

func TestEventWriter_WriteMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	events := []Event{
		{Timestamp: "2024-01-04T10:00:00.000Z", Type: EventTypeToolCall, Summary: "First"},
		{Timestamp: "2024-01-04T10:01:00.000Z", Type: EventTypeFileChange, Summary: "Second"},
	}

	if err := writer.WriteMultiple(events); err != nil {
		t.Fatalf("WriteMultiple failed: %v", err)
	}

	// Count lines
	file, err := os.Open(writer.Path())
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if lineCount != 2 {
		t.Errorf("Expected 2 lines, got %d", lineCount)
	}
}

func TestEventWriter_WriteMultiple_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	// Should not error with empty slice
	if err := writer.WriteMultiple([]Event{}); err != nil {
		t.Errorf("WriteMultiple with empty slice failed: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(writer.Path()); !os.IsNotExist(err) {
		t.Error("Empty WriteMultiple should not create file")
	}
}

func TestEventWriter_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	const numWriters = 10
	const eventsPerWriter = 100

	var wg sync.WaitGroup
	wg.Add(numWriters)

	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < eventsPerWriter; j++ {
				event := NewToolCallEvent("Bash", "", map[string]interface{}{
					"writer_id": writerID,
					"event_num": j,
				})
				if err := writer.Write(event); err != nil {
					t.Errorf("Concurrent write failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were written and file is not corrupt
	file, err := os.Open(writer.Path())
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	validCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Errorf("Corrupt line found: %v", err)
			continue
		}
		validCount++
	}

	expectedTotal := numWriters * eventsPerWriter
	if validCount != expectedTotal {
		t.Errorf("Expected %d valid events, got %d", expectedTotal, validCount)
	}
}

func TestEventWriter_FileCreationOnFirstWrite(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}
	defer writer.Close()

	// File should not exist yet
	if _, err := os.Stat(writer.Path()); !os.IsNotExist(err) {
		t.Error("File should not exist before first write")
	}

	// Write first event
	event := NewToolCallEvent("Read", "/some/file", nil)
	if err := writer.Write(event); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Now file should exist
	if _, err := os.Stat(writer.Path()); os.IsNotExist(err) {
		t.Error("File should exist after first write")
	}
}

func TestEventWriter_Close(t *testing.T) {
	tmpDir := t.TempDir()
	writer, err := NewEventWriter(tmpDir)
	if err != nil {
		t.Fatalf("NewEventWriter failed: %v", err)
	}

	// Close should not error
	if err := writer.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Close again should not error (idempotent)
	if err := writer.Close(); err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

// --- BufferedEventWriter Tests ---

func TestNewBufferedEventWriter(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, DefaultFlushInterval)
	defer writer.Close()

	expectedPath := filepath.Join(tmpDir, EventsFileName)
	if writer.Path() != expectedPath {
		t.Errorf("Path() = %v, want %v", writer.Path(), expectedPath)
	}
}

func TestBufferedEventWriter_Write_NonBlocking(t *testing.T) {
	tmpDir := t.TempDir()

	// Use a long flush interval so events stay buffered
	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)
	defer writer.Close()

	event := NewToolCallEvent("Bash", "/test", nil)

	// Write should be non-blocking (returns void)
	writer.Write(event)

	// Event should be buffered, not yet written to disk
	if writer.Len() != 1 {
		t.Errorf("Expected 1 buffered event, got %d", writer.Len())
	}

	// File should not exist yet (no flush)
	if _, err := os.Stat(writer.Path()); !os.IsNotExist(err) {
		t.Error("File should not exist before flush")
	}
}

func TestBufferedEventWriter_ManualFlush(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)
	defer writer.Close()

	event := NewToolCallEvent("Read", "/some/file", nil)
	writer.Write(event)

	// Manual flush should write events to disk
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Buffer should be empty after flush
	if writer.Len() != 0 {
		t.Errorf("Expected 0 buffered events after flush, got %d", writer.Len())
	}

	// File should now exist with the event
	data, err := os.ReadFile(writer.Path())
	if err != nil {
		t.Fatalf("Failed to read events file: %v", err)
	}

	var parsed Event
	if err := json.Unmarshal(data[:len(data)-1], &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed.Tool != "Read" {
		t.Errorf("Tool = %v, want Read", parsed.Tool)
	}
}

func TestBufferedEventWriter_AutoFlush(t *testing.T) {
	tmpDir := t.TempDir()

	// Use a short flush interval
	flushInterval := 100 * time.Millisecond
	writer := NewBufferedEventWriter(tmpDir, flushInterval)
	defer writer.Close()

	event := NewToolCallEvent("Edit", "/test/file", nil)
	writer.Write(event)

	// Wait for auto-flush (2x interval to be safe)
	time.Sleep(flushInterval * 3)

	// Buffer should be empty after auto-flush
	if writer.Len() != 0 {
		t.Errorf("Expected 0 buffered events after auto-flush, got %d", writer.Len())
	}

	// File should exist
	if _, err := os.Stat(writer.Path()); os.IsNotExist(err) {
		t.Error("File should exist after auto-flush")
	}
}

func TestBufferedEventWriter_CloseFlushes(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)

	events := []Event{
		NewToolCallEvent("Read", "/a", nil),
		NewToolCallEvent("Edit", "/b", nil),
		NewToolCallEvent("Write", "/c", nil),
	}

	for _, e := range events {
		writer.Write(e)
	}

	// Events should be buffered
	if writer.Len() != 3 {
		t.Errorf("Expected 3 buffered events, got %d", writer.Len())
	}

	// Close should perform final flush
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// All events should be written
	file, err := os.Open(writer.Path())
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if lineCount != 3 {
		t.Errorf("Expected 3 events after close, got %d", lineCount)
	}
}

func TestBufferedEventWriter_CloseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, DefaultFlushInterval)

	// First close
	if err := writer.Close(); err != nil {
		t.Errorf("First Close failed: %v", err)
	}

	// Second close should not error
	if err := writer.Close(); err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

func TestBufferedEventWriter_WriteAfterClose(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)
	writer.Close()

	// Write after close should not panic
	event := NewToolCallEvent("Bash", "/test", nil)
	writer.Write(event) // Should silently drop

	// Nothing should be buffered
	if writer.Len() != 0 {
		t.Errorf("Expected 0 buffered events after close, got %d", writer.Len())
	}
}

func TestBufferedEventWriter_FlushEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)
	defer writer.Close()

	// Flush with no events should succeed
	if err := writer.Flush(); err != nil {
		t.Errorf("Flush on empty buffer failed: %v", err)
	}

	// File should not be created
	if _, err := os.Stat(writer.Path()); !os.IsNotExist(err) {
		t.Error("Empty flush should not create file")
	}
}

func TestBufferedEventWriter_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()

	// Short flush interval to exercise concurrent flush + write
	writer := NewBufferedEventWriter(tmpDir, 50*time.Millisecond)
	defer writer.Close()

	const numWriters = 10
	const eventsPerWriter = 100

	var wg sync.WaitGroup
	wg.Add(numWriters)

	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < eventsPerWriter; j++ {
				event := NewToolCallEvent("Bash", "", map[string]interface{}{
					"writer_id": writerID,
					"event_num": j,
				})
				writer.Write(event)
			}
		}(i)
	}

	wg.Wait()

	// Close to ensure final flush
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify all events were written
	file, err := os.Open(writer.Path())
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	validCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Errorf("Corrupt line found: %v", err)
			continue
		}
		validCount++
	}

	expectedTotal := numWriters * eventsPerWriter
	if validCount != expectedTotal {
		t.Errorf("Expected %d valid events, got %d", expectedTotal, validCount)
	}
}

func TestBufferedEventWriter_MultipleFlushes(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)
	defer writer.Close()

	// Write and flush multiple batches
	for batch := 0; batch < 3; batch++ {
		for i := 0; i < 10; i++ {
			event := NewToolCallEvent("Bash", "", map[string]interface{}{
				"batch": batch,
				"index": i,
			})
			writer.Write(event)
		}
		if err := writer.Flush(); err != nil {
			t.Fatalf("Flush %d failed: %v", batch, err)
		}
	}

	// Should have 30 events total
	file, err := os.Open(writer.Path())
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if lineCount != 30 {
		t.Errorf("Expected 30 events, got %d", lineCount)
	}
}

func TestBufferedEventWriter_DefaultFlushInterval(t *testing.T) {
	// Verify the default flush interval is under 10 seconds as required
	if DefaultFlushInterval >= 10*time.Second {
		t.Errorf("DefaultFlushInterval = %v, should be < 10s", DefaultFlushInterval)
	}

	// Verify it's a reasonable value (not too short)
	if DefaultFlushInterval < 1*time.Second {
		t.Errorf("DefaultFlushInterval = %v, should be >= 1s for performance", DefaultFlushInterval)
	}
}

func TestBufferedEventWriter_FlushError(t *testing.T) {
	tmpDir := t.TempDir()

	writer := NewBufferedEventWriter(tmpDir, 1*time.Hour)
	defer writer.Close()

	// Initially no error
	if err := writer.FlushError(); err != nil {
		t.Errorf("Expected no initial flush error, got: %v", err)
	}

	// Write and flush successfully
	event := NewToolCallEvent("Test", "", nil)
	writer.Write(event)
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Still no error after successful flush
	if err := writer.FlushError(); err != nil {
		t.Errorf("Expected no flush error after success, got: %v", err)
	}
}

func TestSessionScopedEventIsolation(t *testing.T) {
	// Verify that two separate session directories produce separate event files.
	// This confirms the session-scoped design: each session writes to its own
	// events.jsonl at .sos/sessions/<session-id>/events.jsonl.
	tmpDir := t.TempDir()

	sessionDirA := filepath.Join(tmpDir, "sessions", "session-aaa")
	sessionDirB := filepath.Join(tmpDir, "sessions", "session-bbb")

	// Write events to session A
	writerA := NewBufferedEventWriter(sessionDirA, DefaultFlushInterval)
	writerA.Write(NewToolCallEvent("Edit", "/file-a.go", nil))
	writerA.Write(NewToolCallEvent("Edit", "/file-a2.go", nil))
	if err := writerA.Close(); err != nil {
		t.Fatalf("Close A failed: %v", err)
	}

	// Write events to session B
	writerB := NewBufferedEventWriter(sessionDirB, DefaultFlushInterval)
	writerB.Write(NewToolCallEvent("Write", "/file-b.go", nil))
	if err := writerB.Close(); err != nil {
		t.Fatalf("Close B failed: %v", err)
	}

	// Verify session A has 2 events and session B has 1 event (no interleaving)
	eventsA, err := os.ReadFile(filepath.Join(sessionDirA, EventsFileName))
	if err != nil {
		t.Fatalf("Failed to read session A events: %v", err)
	}
	eventsB, err := os.ReadFile(filepath.Join(sessionDirB, EventsFileName))
	if err != nil {
		t.Fatalf("Failed to read session B events: %v", err)
	}

	linesA := countLines(eventsA)
	linesB := countLines(eventsB)

	if linesA != 2 {
		t.Errorf("Session A: expected 2 events, got %d", linesA)
	}
	if linesB != 1 {
		t.Errorf("Session B: expected 1 event, got %d", linesB)
	}

	// Verify the paths are distinct
	pathA := filepath.Join(sessionDirA, EventsFileName)
	pathB := filepath.Join(sessionDirB, EventsFileName)
	if pathA == pathB {
		t.Error("Session event file paths should be different")
	}
}

// countLines counts non-empty lines in data.
func countLines(data []byte) int {
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			count++
		}
	}
	return count
}
