package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewSprintContext(t *testing.T) {
	ctx := NewSprintContext("session-20260104-160414-563c681e", "Test Goal", []string{"Task A", "Task B"})

	if ctx.SprintID == "" {
		t.Error("SprintID should not be empty")
	}
	if !IsValidSprintID(ctx.SprintID) {
		t.Errorf("SprintID %q is not valid", ctx.SprintID)
	}
	if ctx.SessionID != "session-20260104-160414-563c681e" {
		t.Errorf("SessionID = %q, want %q", ctx.SessionID, "session-20260104-160414-563c681e")
	}
	if ctx.Goal != "Test Goal" {
		t.Errorf("Goal = %q, want %q", ctx.Goal, "Test Goal")
	}
	if ctx.Status != SprintStatusActive {
		t.Errorf("Status = %v, want %v", ctx.Status, SprintStatusActive)
	}
	if ctx.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "1.0")
	}
	if len(ctx.Tasks) != 2 {
		t.Fatalf("Tasks length = %d, want 2", len(ctx.Tasks))
	}
	if ctx.Tasks[0].ID != "task-001" {
		t.Errorf("Tasks[0].ID = %q, want %q", ctx.Tasks[0].ID, "task-001")
	}
	if ctx.Tasks[0].Description != "Task A" {
		t.Errorf("Tasks[0].Description = %q, want %q", ctx.Tasks[0].Description, "Task A")
	}
	if ctx.Tasks[0].Status != string(TaskStatusPending) {
		t.Errorf("Tasks[0].Status = %q, want %q", ctx.Tasks[0].Status, TaskStatusPending)
	}
	if ctx.Tasks[1].ID != "task-002" {
		t.Errorf("Tasks[1].ID = %q, want %q", ctx.Tasks[1].ID, "task-002")
	}
}

func TestNewSprintContext_NoTasks(t *testing.T) {
	ctx := NewSprintContext("session-20260104-160414-563c681e", "Goal", nil)
	if len(ctx.Tasks) != 0 {
		t.Errorf("Tasks length = %d, want 0", len(ctx.Tasks))
	}
}

func TestParseSprintContext(t *testing.T) {
	content := `---
schema_version: "1.0"
sprint_id: "sprint-20260208-120000-abcdef01"
session_id: "session-20260104-160414-563c681e"
goal: "Test Sprint"
status: "ACTIVE"
created_at: "2026-02-08T12:00:00Z"
tasks:
  - id: "task-001"
    description: "First task"
    status: "pending"
---

# Sprint: Test Sprint
`

	ctx, err := ParseSprintContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseSprintContext() error = %v", err)
	}

	if ctx.SprintID != "sprint-20260208-120000-abcdef01" {
		t.Errorf("SprintID = %q, want %q", ctx.SprintID, "sprint-20260208-120000-abcdef01")
	}
	if ctx.Goal != "Test Sprint" {
		t.Errorf("Goal = %q, want %q", ctx.Goal, "Test Sprint")
	}
	if ctx.Status != SprintStatusActive {
		t.Errorf("Status = %v, want %v", ctx.Status, SprintStatusActive)
	}
	if len(ctx.Tasks) != 1 {
		t.Fatalf("Tasks length = %d, want 1", len(ctx.Tasks))
	}
}

func TestParseSprintContext_NoFrontmatter(t *testing.T) {
	_, err := ParseSprintContext([]byte("# Just markdown"))
	if err == nil {
		t.Error("ParseSprintContext() should error on missing frontmatter")
	}
}

func TestParseSprintContext_UnclosedFrontmatter(t *testing.T) {
	content := `---
schema_version: "1.0"
sprint_id: "test"
`
	_, err := ParseSprintContext([]byte(content))
	if err == nil {
		t.Error("ParseSprintContext() should error on unclosed frontmatter")
	}
}

func TestSprintContext_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := &SprintContext{
		SchemaVersion: "1.0",
		SprintID:      "sprint-20260208-120000-abcdef01",
		SessionID:     "session-20260104-160414-563c681e",
		Goal:          "Round Trip Test",
		Status:        SprintStatusActive,
		CreatedAt:     now,
		Tasks: []SprintTask{
			{ID: "task-001", Description: "First", Status: "pending"},
			{ID: "task-002", Description: "Second", Status: "done"},
		},
		Body: "\n# Test\n",
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	parsed, err := ParseSprintContext(data)
	if err != nil {
		t.Fatalf("ParseSprintContext() error = %v", err)
	}

	if parsed.SprintID != original.SprintID {
		t.Errorf("SprintID mismatch: got %q, want %q", parsed.SprintID, original.SprintID)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status mismatch: got %v, want %v", parsed.Status, original.Status)
	}
	if parsed.Goal != original.Goal {
		t.Errorf("Goal mismatch: got %q, want %q", parsed.Goal, original.Goal)
	}
	if len(parsed.Tasks) != 2 {
		t.Fatalf("Tasks length mismatch: got %d, want 2", len(parsed.Tasks))
	}
}

func TestSprintContext_Serialize_CompletedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	completed := now.Add(time.Hour)

	ctx := &SprintContext{
		SchemaVersion: "1.0",
		SprintID:      "sprint-20260208-120000-abcdef01",
		SessionID:     "session-20260104-160414-563c681e",
		Goal:          "Completed Sprint",
		Status:        SprintStatusCompleted,
		CreatedAt:     now,
		CompletedAt:   &completed,
	}

	data, err := ctx.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	str := string(data)
	if !strings.Contains(str, "completed_at:") {
		t.Error("Serialized content should contain completed_at")
	}

	parsed, err := ParseSprintContext(data)
	if err != nil {
		t.Fatalf("ParseSprintContext() error = %v", err)
	}
	if parsed.CompletedAt == nil {
		t.Fatal("CompletedAt should not be nil")
	}
	if !parsed.CompletedAt.Equal(completed) {
		t.Errorf("CompletedAt mismatch: got %v, want %v", parsed.CompletedAt, completed)
	}
}

func TestSprintContext_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SPRINT_CONTEXT.md")

	ctx := NewSprintContext("session-20260104-160414-563c681e", "Save Test", []string{"Task A"})
	if err := ctx.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadSprintContext(path)
	if err != nil {
		t.Fatalf("LoadSprintContext() error = %v", err)
	}

	if loaded.SprintID != ctx.SprintID {
		t.Errorf("SprintID mismatch: got %q, want %q", loaded.SprintID, ctx.SprintID)
	}
	if loaded.Goal != ctx.Goal {
		t.Errorf("Goal mismatch: got %q, want %q", loaded.Goal, ctx.Goal)
	}
}

func TestLoadSprintContext_NotFound(t *testing.T) {
	_, err := LoadSprintContext("/nonexistent/path/SPRINT_CONTEXT.md")
	if err == nil {
		t.Error("LoadSprintContext() should error on missing file")
	}
}

func TestSprintContext_MarkTaskComplete(t *testing.T) {
	ctx := NewSprintContext("session-20260104-160414-563c681e", "Test", []string{"Task A", "Task B"})

	if err := ctx.MarkTaskComplete("task-001"); err != nil {
		t.Fatalf("MarkTaskComplete() error = %v", err)
	}
	if ctx.Tasks[0].Status != string(TaskStatusDone) {
		t.Errorf("Tasks[0].Status = %q, want %q", ctx.Tasks[0].Status, TaskStatusDone)
	}

	// Already completed
	if err := ctx.MarkTaskComplete("task-001"); err == nil {
		t.Error("MarkTaskComplete() should error on already completed task")
	}

	// Not found
	if err := ctx.MarkTaskComplete("task-999"); err == nil {
		t.Error("MarkTaskComplete() should error on missing task")
	}
}

func TestSprintContext_AllTasksDone(t *testing.T) {
	ctx := NewSprintContext("session-20260104-160414-563c681e", "Test", []string{"A", "B"})

	if ctx.AllTasksDone() {
		t.Error("AllTasksDone() should return false when tasks are pending")
	}

	ctx.Tasks[0].Status = string(TaskStatusDone)
	if ctx.AllTasksDone() {
		t.Error("AllTasksDone() should return false when not all tasks are done")
	}

	ctx.Tasks[1].Status = string(TaskStatusSkipped)
	if !ctx.AllTasksDone() {
		t.Error("AllTasksDone() should return true when all tasks are done/skipped")
	}
}

func TestSprintContext_AllTasksDone_Empty(t *testing.T) {
	ctx := NewSprintContext("session-20260104-160414-563c681e", "Test", nil)
	if ctx.AllTasksDone() {
		t.Error("AllTasksDone() should return false for empty tasks")
	}
}

func TestGenerateSprintID(t *testing.T) {
	id := GenerateSprintID()
	if !strings.HasPrefix(id, "sprint-") {
		t.Errorf("SprintID %q should start with 'sprint-'", id)
	}
	if !IsValidSprintID(id) {
		t.Errorf("Generated SprintID %q is not valid", id)
	}
}

func TestIsValidSprintID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"sprint-20260208-120000-abcdef01", true},
		{"sprint-20260208-120000-12345678", true},
		{"session-20260208-120000-abcdef01", false},
		{"sprint-2026020-120000-abcdef01", false},
		{"", false},
		{"sprint", false},
	}

	for _, tt := range tests {
		if got := IsValidSprintID(tt.id); got != tt.valid {
			t.Errorf("IsValidSprintID(%q) = %v, want %v", tt.id, got, tt.valid)
		}
	}
}

func TestSprintContext_SaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SPRINT_CONTEXT.md")

	ctx := NewSprintContext("test-session", "Test", nil)
	if err := ctx.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() should create the file")
	}
}
