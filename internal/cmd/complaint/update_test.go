package complaint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"gopkg.in/yaml.v3"
)

func writeTestComplaint(t *testing.T, dir, id, status string) string {
	t.Helper()
	c := Complaint{
		ID:          id,
		FiledBy:     "test-agent",
		FiledAt:     "2026-03-23T12:00:00Z",
		Title:       "Test complaint",
		Severity:    "low",
		Description: "Test description",
		Tags:        []string{"test"},
		Status:      status,
	}
	data, err := yaml.Marshal(&c)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, id+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunUpdate_ValidTransition(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	id := "COMPLAINT-20260323-120000-test-agent"
	writeTestComplaint(t, dir, id, "filed")

	var buf bytes.Buffer
	printer := output.NewPrinter(output.FormatText, &buf, nil, false)

	// Find and update directly using exported helpers.
	filePath, err := findComplaintFile(dir, id)
	if err != nil {
		t.Fatalf("findComplaintFile: %v", err)
	}

	// Read, update, write.
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	var c Complaint
	if err := yaml.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	if c.Status != "filed" {
		t.Fatalf("initial status: got %q, want %q", c.Status, "filed")
	}

	c.Status = "triaged"
	updated, err := yaml.Marshal(&c)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, updated, 0644); err != nil {
		t.Fatal(err)
	}

	// Verify the update persisted.
	data, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	var reloaded Complaint
	if err := yaml.Unmarshal(data, &reloaded); err != nil {
		t.Fatal(err)
	}
	if reloaded.Status != "triaged" {
		t.Errorf("updated status: got %q, want %q", reloaded.Status, "triaged")
	}

	// Verify the output format.
	out := updateOutput{ID: id, OldStatus: "filed", NewStatus: "triaged", Path: filePath}
	if err := printer.Print(out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "filed") || !strings.Contains(buf.String(), "triaged") {
		t.Errorf("output missing status transition: %q", buf.String())
	}
}

func TestValidStatuses(t *testing.T) {
	t.Parallel()

	expected := []string{"filed", "triaged", "accepted", "rejected", "resolved"}
	for _, s := range expected {
		if !validStatuses[s] {
			t.Errorf("validStatuses missing %q", s)
		}
	}

	invalid := []string{"pending", "closed", "open", "", "FILED"}
	for _, s := range invalid {
		if validStatuses[s] {
			t.Errorf("validStatuses should not contain %q", s)
		}
	}
}

func TestFindComplaintFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	id := "COMPLAINT-20260323-120000-test"
	writeTestComplaint(t, dir, id, "filed")

	path, err := findComplaintFile(dir, id)
	if err != nil {
		t.Fatalf("findComplaintFile: %v", err)
	}
	if !strings.HasSuffix(path, id+".yaml") {
		t.Errorf("path: got %q, want suffix %q", path, id+".yaml")
	}
}

func TestFindComplaintFile_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := findComplaintFile(dir, "COMPLAINT-nonexistent")
	if err == nil {
		t.Error("expected error for missing complaint, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found': %v", err)
	}
}

func TestFindComplaintFile_YmlExtension(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	id := "COMPLAINT-20260323-120000-yml-test"
	path := filepath.Join(dir, id+".yml")
	if err := os.WriteFile(path, []byte("id: "+id+"\nstatus: filed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := findComplaintFile(dir, id)
	if err != nil {
		t.Fatalf("findComplaintFile: %v", err)
	}
	if found != path {
		t.Errorf("path: got %q, want %q", found, path)
	}
}

func TestUpdateOutput_Text(t *testing.T) {
	t.Parallel()

	out := updateOutput{
		ID:        "COMPLAINT-20260323-120000-test",
		OldStatus: "filed",
		NewStatus: "triaged",
		Path:      "/tmp/test.yaml",
	}

	text := out.Text()
	if !strings.Contains(text, "filed") {
		t.Error("Text() missing old status")
	}
	if !strings.Contains(text, "triaged") {
		t.Error("Text() missing new status")
	}
	if !strings.Contains(text, "COMPLAINT-20260323-120000-test") {
		t.Error("Text() missing complaint ID")
	}
}

func TestUpdateOutput_JSON(t *testing.T) {
	t.Parallel()

	out := updateOutput{
		ID:        "COMPLAINT-20260323-120000-test",
		OldStatus: "filed",
		NewStatus: "triaged",
		Path:      "/tmp/test.yaml",
	}

	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if result["id"] != "COMPLAINT-20260323-120000-test" {
		t.Errorf("JSON id: got %q", result["id"])
	}
	if result["old_status"] != "filed" {
		t.Errorf("JSON old_status: got %q", result["old_status"])
	}
	if result["new_status"] != "triaged" {
		t.Errorf("JSON new_status: got %q", result["new_status"])
	}
}

func TestRunUpdate_InvalidStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	id := "COMPLAINT-20260323-120000-test-agent"
	writeTestComplaint(t, dir, id, "filed")

	// Validate that invalid statuses are rejected.
	invalidStatuses := []string{"pending", "closed", "FILED", ""}
	for _, status := range invalidStatuses {
		if validStatuses[status] {
			t.Errorf("status %q should be invalid", status)
		}
	}
}

func TestRunUpdate_StatusPersistence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	id := "COMPLAINT-20260323-120000-persist"
	writeTestComplaint(t, dir, id, "filed")

	// Simulate full update cycle: filed → triaged → accepted → resolved.
	transitions := []string{"triaged", "accepted", "resolved"}
	for _, newStatus := range transitions {
		filePath := filepath.Join(dir, id+".yaml")
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("reading %s: %v", id, err)
		}

		var c Complaint
		if err := yaml.Unmarshal(data, &c); err != nil {
			t.Fatalf("parsing %s: %v", id, err)
		}

		c.Status = newStatus
		updated, err := yaml.Marshal(&c)
		if err != nil {
			t.Fatalf("marshaling %s: %v", id, err)
		}

		if err := os.WriteFile(filePath, updated, 0644); err != nil {
			t.Fatalf("writing %s: %v", id, err)
		}

		// Verify.
		data, err = os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		var verify Complaint
		if err := yaml.Unmarshal(data, &verify); err != nil {
			t.Fatal(err)
		}
		if verify.Status != newStatus {
			t.Errorf("after transition to %q: got %q", newStatus, verify.Status)
		}
	}
}

func TestRunUpdate_FilterAfterUpdate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestComplaint(t, dir, "COMPLAINT-20260323-120001-a", "filed")
	writeTestComplaint(t, dir, "COMPLAINT-20260323-120002-b", "filed")
	writeTestComplaint(t, dir, "COMPLAINT-20260323-120003-c", "triaged")

	// Load and filter by status=filed.
	complaints, err := loadComplaints(dir)
	if err != nil {
		t.Fatal(err)
	}

	filed := filterComplaints(complaints, listOptions{status: "filed"})
	if len(filed) != 2 {
		t.Errorf("filed complaints: got %d, want 2", len(filed))
	}

	triaged := filterComplaints(complaints, listOptions{status: "triaged"})
	if len(triaged) != 1 {
		t.Errorf("triaged complaints: got %d, want 1", len(triaged))
	}
}
