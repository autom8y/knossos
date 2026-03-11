package complaint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/output"
)

func TestLoadComplaints(t *testing.T) {
	t.Parallel()

	complaints, err := loadComplaints("testdata")
	if err != nil {
		t.Fatalf("loadComplaints(testdata): unexpected error: %v", err)
	}

	if len(complaints) != 3 {
		t.Errorf("loadComplaints(testdata): got %d complaints, want 3", len(complaints))
	}

	// Verify specific fields on a known fixture.
	var found bool
	for _, c := range complaints {
		if c.ID == "COMPLAINT-20260311-091500-pythia" {
			found = true
			if c.FiledBy != "pythia" {
				t.Errorf("filed_by: got %q, want %q", c.FiledBy, "pythia")
			}
			if c.Severity != "high" {
				t.Errorf("severity: got %q, want %q", c.Severity, "high")
			}
			if c.Status != "triaged" {
				t.Errorf("status: got %q, want %q", c.Status, "triaged")
			}
			if c.Evidence == nil {
				t.Error("evidence: expected non-nil for deep-file fixture")
			} else if c.Evidence.SessionID != "session-20260311-012734-9847ff6f" {
				t.Errorf("evidence.session_id: got %q, want %q",
					c.Evidence.SessionID, "session-20260311-012734-9847ff6f")
			}
		}
	}
	if !found {
		t.Error("loadComplaints: COMPLAINT-20260311-091500-pythia not found in results")
	}
}

func TestLoadComplaints_MissingDir(t *testing.T) {
	t.Parallel()

	_, err := loadComplaints("/tmp/nonexistent-knossos-complaints-xyz")
	if err == nil {
		t.Error("loadComplaints(missing dir): expected error, got nil")
	}
}

func TestLoadComplaints_SkipsNonYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write a non-YAML file that starts with COMPLAINT-.
	if err := os.WriteFile(filepath.Join(dir, "COMPLAINT-not-yaml.txt"), []byte("not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write a YAML that does NOT start with COMPLAINT-.
	if err := os.WriteFile(filepath.Join(dir, "other.yaml"), []byte("id: other"), 0o644); err != nil {
		t.Fatal(err)
	}

	complaints, err := loadComplaints(dir)
	if err != nil {
		t.Fatalf("loadComplaints: unexpected error: %v", err)
	}
	if len(complaints) != 0 {
		t.Errorf("expected 0 complaints, got %d", len(complaints))
	}
}

func TestFilterComplaints_ByStatus(t *testing.T) {
	t.Parallel()

	complaints := []Complaint{
		{ID: "A", Status: "filed", Severity: "low"},
		{ID: "B", Status: "triaged", Severity: "high"},
		{ID: "C", Status: "filed", Severity: "medium"},
	}

	got := filterComplaints(complaints, listOptions{status: "filed"})
	if len(got) != 2 {
		t.Errorf("filter by status=filed: got %d, want 2", len(got))
	}
	for _, c := range got {
		if c.Status != "filed" {
			t.Errorf("filter by status=filed: unexpected status %q in result", c.Status)
		}
	}
}

func TestFilterComplaints_BySeverity(t *testing.T) {
	t.Parallel()

	complaints := []Complaint{
		{ID: "A", Status: "filed", Severity: "low"},
		{ID: "B", Status: "triaged", Severity: "high"},
		{ID: "C", Status: "filed", Severity: "high"},
	}

	got := filterComplaints(complaints, listOptions{severity: "high"})
	if len(got) != 2 {
		t.Errorf("filter by severity=high: got %d, want 2", len(got))
	}
}

func TestFilterComplaints_Combined(t *testing.T) {
	t.Parallel()

	complaints := []Complaint{
		{ID: "A", Status: "filed", Severity: "high"},
		{ID: "B", Status: "triaged", Severity: "high"},
		{ID: "C", Status: "filed", Severity: "medium"},
	}

	got := filterComplaints(complaints, listOptions{status: "filed", severity: "high"})
	if len(got) != 1 {
		t.Errorf("filter by status=filed,severity=high: got %d, want 1", len(got))
	}
	if got[0].ID != "A" {
		t.Errorf("wrong complaint returned: got %q, want %q", got[0].ID, "A")
	}
}

func TestFilterComplaints_NoFilters(t *testing.T) {
	t.Parallel()

	complaints := []Complaint{
		{ID: "A", Status: "filed"},
		{ID: "B", Status: "triaged"},
	}

	got := filterComplaints(complaints, listOptions{})
	if len(got) != 2 {
		t.Errorf("no filter: got %d, want 2", len(got))
	}
}

func TestTruncateTitle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 40, "short"},
		{"exactly40chars1234567890123456789012345", 40, "exactly40chars1234567890123456789012345"},
		{"this is a very long title that exceeds forty characters easily", 40, "this is a very long title that exceed..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tc := range cases {
		got := truncateTitle(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncateTitle(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}

func TestComplaintListOutput_Text_Empty(t *testing.T) {
	t.Parallel()

	o := complaintListOutput{Complaints: nil, Total: 0}
	if o.Text() != "No complaints found.\n" {
		t.Errorf("empty output Text(): got %q, want %q", o.Text(), "No complaints found.\n")
	}
}

func TestComplaintListOutput_Text_NonEmpty(t *testing.T) {
	t.Parallel()

	complaints := []Complaint{
		{
			ID:       "COMPLAINT-20260311-143022-drift-detect",
			Severity: "medium",
			Title:    "retry-spiral drift: Bash tool failed",
			Status:   "filed",
			FiledAt:  "2026-03-11T14:30:22Z",
		},
	}
	o := complaintListOutput{Complaints: complaints, Total: 1}
	text := o.Text()

	if !strings.Contains(text, "COMPLAINT-20260311-143022-drift-detect") {
		t.Error("Text() missing complaint ID")
	}
	if !strings.Contains(text, "medium") {
		t.Error("Text() missing severity")
	}
	if !strings.Contains(text, "filed") {
		t.Error("Text() missing status")
	}
	if !strings.Contains(text, "2026-03-11") {
		t.Error("Text() missing filed date")
	}
}

func TestComplaintListOutput_JSON_EmptyArray(t *testing.T) {
	t.Parallel()

	o := complaintListOutput{Complaints: nil, Total: 0}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	arr, ok := result["complaints"].([]any)
	if !ok {
		t.Errorf("complaints field should be an array, got %T", result["complaints"])
	}
	if len(arr) != 0 {
		t.Errorf("complaints array should be empty, got %d elements", len(arr))
	}
}

func TestRunList_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	complaintsDir := filepath.Join(dir, ".sos", "wip", "complaints")
	if err := os.MkdirAll(complaintsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	printer := output.NewPrinter(output.FormatText, &buf, nil, false)

	complaints, _ := loadComplaints(complaintsDir)
	if err := printComplaints(printer, complaints); err != nil {
		t.Fatalf("printComplaints: %v", err)
	}

	if !strings.Contains(buf.String(), "No complaints found.") {
		t.Errorf("empty dir output: got %q, want 'No complaints found.'", buf.String())
	}
}

func TestRunList_MissingDir(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printer := output.NewPrinter(output.FormatText, &buf, nil, false)

	// loadComplaints on missing dir returns error; runList treats it as empty.
	complaints, _ := loadComplaints("/tmp/nonexistent-knossos-complaints-xyz")
	if err := printComplaints(printer, complaints); err != nil {
		t.Fatalf("printComplaints: %v", err)
	}

	if !strings.Contains(buf.String(), "No complaints found.") {
		t.Errorf("missing dir output: got %q, want 'No complaints found.'", buf.String())
	}
}

func TestRunList_WithFixtures_JSON(t *testing.T) {
	t.Parallel()

	complaints, err := loadComplaints("testdata")
	if err != nil {
		t.Fatalf("loadComplaints: %v", err)
	}

	var buf bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &buf, nil, false)
	if err := printComplaints(printer, complaints); err != nil {
		t.Fatalf("printComplaints: %v", err)
	}

	var result struct {
		Complaints []Complaint `json:"complaints"`
		Total      int         `json:"total"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("JSON total: got %d, want 3", result.Total)
	}
	if len(result.Complaints) != 3 {
		t.Errorf("JSON complaints: got %d, want 3", len(result.Complaints))
	}
}

func TestComplaintListOutput_Headers(t *testing.T) {
	t.Parallel()

	o := complaintListOutput{}
	headers := o.Headers()
	expected := []string{"ID", "SEVERITY", "TITLE", "STATUS", "FILED"}
	if len(headers) != len(expected) {
		t.Fatalf("Headers() len: got %d, want %d", len(headers), len(expected))
	}
	for i, h := range expected {
		if headers[i] != h {
			t.Errorf("Headers()[%d]: got %q, want %q", i, headers[i], h)
		}
	}
}

func TestComplaintListOutput_Rows_Empty(t *testing.T) {
	t.Parallel()

	o := complaintListOutput{Complaints: nil}
	rows := o.Rows()
	if len(rows) != 0 {
		t.Errorf("Rows() on empty: got %d rows, want 0", len(rows))
	}
}
