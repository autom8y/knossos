package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewSprintContext(t *testing.T) {
	ctx := NewSprintContext("session-20260208-120000-abcdef01", "Test Goal", []string{"Task A", "Task B"})

	if ctx.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "1.0")
	}
	if !IsValidSprintID(ctx.SprintID) {
		t.Errorf("SprintID %q is not valid", ctx.SprintID)
	}
	if ctx.SessionID != "session-20260208-120000-abcdef01" {
		t.Errorf("SessionID = %q, want %q", ctx.SessionID, "session-20260208-120000-abcdef01")
	}
	if ctx.Goal != "Test Goal" {
		t.Errorf("Goal = %q, want %q", ctx.Goal, "Test Goal")
	}
	if ctx.Status != SprintStatusActive {
		t.Errorf("Status = %q, want %q", ctx.Status, SprintStatusActive)
	}
	if ctx.StartedAt == nil {
		t.Error("StartedAt should not be nil for active sprint")
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
	if ctx.Tasks[0].Status != "pending" {
		t.Errorf("Tasks[0].Status = %q, want %q", ctx.Tasks[0].Status, "pending")
	}
	if ctx.Tasks[1].ID != "task-002" {
		t.Errorf("Tasks[1].ID = %q, want %q", ctx.Tasks[1].ID, "task-002")
	}
}

func TestNewSprintContext_NoTasks(t *testing.T) {
	ctx := NewSprintContext("session-20260208-120000-abcdef01", "Goal", nil)

	if len(ctx.Tasks) != 0 {
		t.Errorf("Tasks length = %d, want 0", len(ctx.Tasks))
	}
}

func TestParseSprintContext(t *testing.T) {
	content := `---
schema_version: "1.0"
sprint_id: "sprint-20260208-120000-abcdef01"
session_id: "session-20260208-100000-11111111"
goal: "Test Sprint"
status: "ACTIVE"
created_at: "2026-02-08T12:00:00Z"
started_at: "2026-02-08T12:00:00Z"
tasks:
  - id: "task-001"
    description: "First task"
    status: "pending"
  - id: "task-002"
    description: "Second task"
    status: "done"
---

# Sprint: Test Sprint

## Progress
`

	ctx, err := ParseSprintContext([]byte(content))
	if err != nil {
		t.Fatalf("ParseSprintContext() error = %v", err)
	}

	if ctx.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "1.0")
	}
	if ctx.SprintID != "sprint-20260208-120000-abcdef01" {
		t.Errorf("SprintID = %q, want %q", ctx.SprintID, "sprint-20260208-120000-abcdef01")
	}
	if ctx.Status != SprintStatusActive {
		t.Errorf("Status = %q, want %q", ctx.Status, SprintStatusActive)
	}
	if len(ctx.Tasks) != 2 {
		t.Fatalf("Tasks length = %d, want 2", len(ctx.Tasks))
	}
	if ctx.Tasks[1].Status != "done" {
		t.Errorf("Tasks[1].Status = %q, want %q", ctx.Tasks[1].Status, "done")
	}
	if ctx.StartedAt == nil {
		t.Error("StartedAt should not be nil")
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
sprint_id: "test"
# Never closed
`
	_, err := ParseSprintContext([]byte(content))
	if err == nil {
		t.Error("ParseSprintContext() should error on unclosed frontmatter")
	}
}

func TestSprintContext_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	startedAt := now

	original := &SprintContext{
		SchemaVersion: "1.0",
		SprintID:      "sprint-20260208-120000-abcdef01",
		SessionID:     "session-20260208-100000-11111111",
		Goal:          "Round Trip Test",
		Status:        SprintStatusActive,
		CreatedAt:     now,
		StartedAt:     &startedAt,
		Tasks: []SprintTask{
			{ID: "task-001", Description: "First", Status: "pending"},
			{ID: "task-002", Description: "Second", Status: "done", Agent: "engineer"},
		},
		Body: "\n# Sprint: Round Trip Test\n",
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
	if parsed.SessionID != original.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", parsed.SessionID, original.SessionID)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", parsed.Status, original.Status)
	}
	if parsed.Goal != original.Goal {
		t.Errorf("Goal mismatch: got %q, want %q", parsed.Goal, original.Goal)
	}
	if len(parsed.Tasks) != len(original.Tasks) {
		t.Fatalf("Tasks length mismatch: got %d, want %d", len(parsed.Tasks), len(original.Tasks))
	}
	if parsed.Tasks[1].Agent != "engineer" {
		t.Errorf("Tasks[1].Agent mismatch: got %q, want %q", parsed.Tasks[1].Agent, "engineer")
	}
}

func TestSprintContext_Serialize_CompletedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	completed := now.Add(time.Hour)

	ctx := &SprintContext{
		SchemaVersion: "1.0",
		SprintID:      "sprint-20260208-120000-abcdef01",
		SessionID:     "session-20260208-100000-11111111",
		Goal:          "Test",
		Status:        SprintStatusCompleted,
		CreatedAt:     now,
		CompletedAt:   &completed,
		Body:          "\n# Test\n",
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
}

func TestSprintContext_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "SPRINT_CONTEXT.md")

	original := NewSprintContext("session-20260208-120000-abcdef01", "Save Test", []string{"task1"})
	if err := original.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("File should exist after Save(): %v", err)
	}

	loaded, err := LoadSprintContext(path)
	if err != nil {
		t.Fatalf("LoadSprintContext() error = %v", err)
	}

	if loaded.SprintID != original.SprintID {
		t.Errorf("SprintID mismatch: got %q, want %q", loaded.SprintID, original.SprintID)
	}
	if loaded.Goal != original.Goal {
		t.Errorf("Goal mismatch: got %q, want %q", loaded.Goal, original.Goal)
	}
}

func TestLoadSprintContext_NotFound(t *testing.T) {
	_, err := LoadSprintContext("/nonexistent/path/SPRINT_CONTEXT.md")
	if err == nil {
		t.Error("LoadSprintContext() should error on missing file")
	}
}

func TestSprintContext_MarkTaskComplete(t *testing.T) {
	ctx := NewSprintContext("session-20260208-120000-abcdef01", "Test", []string{"A", "B", "C"})

	if err := ctx.MarkTaskComplete("task-001"); err != nil {
		t.Fatalf("MarkTaskComplete() error = %v", err)
	}
	if ctx.Tasks[0].Status != "done" {
		t.Errorf("Tasks[0].Status = %q, want %q", ctx.Tasks[0].Status, "done")
	}

	if err := ctx.MarkTaskComplete("task-001"); err == nil {
		t.Error("MarkTaskComplete() should error on already completed task")
	}

	if err := ctx.MarkTaskComplete("task-999"); err == nil {
		t.Error("MarkTaskComplete() should error on non-existent task")
	}
}

func TestSprintContext_AllTasksDone(t *testing.T) {
	ctx := NewSprintContext("session-20260208-120000-abcdef01", "Test", []string{"A", "B"})

	if ctx.AllTasksDone() {
		t.Error("AllTasksDone() should be false with pending tasks")
	}

	ctx.Tasks[0].Status = "done"
	if ctx.AllTasksDone() {
		t.Error("AllTasksDone() should be false with one pending task")
	}

	ctx.Tasks[1].Status = "skipped"
	if !ctx.AllTasksDone() {
		t.Error("AllTasksDone() should be true when all tasks are done/skipped")
	}
}

func TestSprintContext_AllTasksDone_Empty(t *testing.T) {
	ctx := NewSprintContext("session-20260208-120000-abcdef01", "Test", nil)
	if ctx.AllTasksDone() {
		t.Error("AllTasksDone() should be false with no tasks")
	}
}

func TestGenerateSprintID(t *testing.T) {
	id := GenerateSprintID()
	if !IsValidSprintID(id) {
		t.Errorf("GenerateSprintID() = %q, does not match expected pattern", id)
	}

	id2 := GenerateSprintID()
	if id == id2 {
		t.Errorf("GenerateSprintID() produced duplicate IDs: %q", id)
	}
}

func TestIsValidSprintID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"sprint-20260208-120000-abcdef01", true},
		{"sprint-20251231-235959-deadbeef", true},
		{"sprint-20260101-000000-00000000", true},
		{"session-20260208-120000-abcdef01", false},
		{"sprint-2026020-120000-abcdef01", false},
		{"sprint-20260208-12000-abcdef01", false},
		{"sprint-20260208-120000-abcdef0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := IsValidSprintID(tt.id); got != tt.want {
				t.Errorf("IsValidSprintID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}
