package threadcontract

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
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
