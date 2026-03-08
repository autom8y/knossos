package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// newQueryTestContext builds a cmdContext pointing at projectDir with optional
// explicit session ID. Pass nil sessionID to rely on smart scan resolution.
func newQueryTestContext(projectDir string, sessionID *string) *cmdContext {
	outFlag := "text"
	verboseFlag := false
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: sessionID,
		},
	}
}

// setupQuerySession creates a temp project with one active session on disk.
// Returns (projectDir, sessionID, *session.Context).
func setupQuerySession(t *testing.T) (string, string, *session.Context) {
	t.Helper()
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	os.MkdirAll(sessionsDir, 0755)
	os.MkdirAll(locksDir, 0755)

	sessCtx := session.NewContext("S3 query initiative", "MODULE", "ecosystem")
	sessionDir := filepath.Join(sessionsDir, sessCtx.SessionID)
	os.MkdirAll(sessionDir, 0755)
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := sessCtx.Save(ctxPath); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}
	// Write .current-session so FindActiveSession resolves via smart scan
	os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessCtx.SessionID), 0644)

	return tmpDir, sessCtx.SessionID, sessCtx
}

// --- runQuery integration-level tests ---

func TestQuery_RunQuery_NoSession_FullOutput(t *testing.T) {
	// Empty project — no sessions. Full output path returns has_session: false,
	// which is not an error.
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".sos", "sessions"), 0755)

	ctx := newQueryTestContext(tmpDir, nil)
	opts := queryOptions{}

	if err := runQuery(ctx, opts); err != nil {
		t.Fatalf("runQuery() with no session should not error, got: %v", err)
	}
}

func TestQuery_RunQuery_NoSession_FieldQuery_Errors(t *testing.T) {
	// Empty project — querying a specific field with no session should error.
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".sos", "sessions"), 0755)

	ctx := newQueryTestContext(tmpDir, nil)
	opts := queryOptions{field: "complexity"}

	err := runQuery(ctx, opts)
	if err == nil {
		t.Fatal("runQuery() with --field and no session should return error")
	}
}

func TestQuery_RunQuery_SessionNotFound(t *testing.T) {
	// Explicit session ID that doesn't exist on disk should error.
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".sos", "sessions"), 0755)

	nonexistent := "session-20260306-000000-deadbeefdeadbeef"
	ctx := newQueryTestContext(tmpDir, &nonexistent)
	opts := queryOptions{}

	err := runQuery(ctx, opts)
	if err == nil {
		t.Fatal("runQuery() with nonexistent session should return error")
	}
}

func TestQuery_RunQuery_ActiveSession(t *testing.T) {
	projectDir, sessionID, _ := setupQuerySession(t)
	ctx := newQueryTestContext(projectDir, &sessionID)
	opts := queryOptions{}

	// Must not error with a valid active session
	if err := runQuery(ctx, opts); err != nil {
		t.Fatalf("runQuery() error = %v", err)
	}
}

func TestQuery_RunQuery_FieldQuery_UnknownField(t *testing.T) {
	projectDir, sessionID, _ := setupQuerySession(t)
	ctx := newQueryTestContext(projectDir, &sessionID)
	opts := queryOptions{field: "does_not_exist"}

	err := runQuery(ctx, opts)
	if err == nil {
		t.Fatal("runQuery() with unknown --field should return error")
	}
}

// --- getQueryField unit tests ---

func TestQuery_GetQueryField_AllKnownFields(t *testing.T) {
	_, sessionID, _ := setupQuerySession(t)

	// Build a synthetic session context for testing
	sessCtx := &session.Context{
		SessionID:    sessionID,
		Status:       session.StatusActive,
		Initiative:   "test initiative",
		Complexity:   "SYSTEM",
		ActiveRite:   "ecosystem",
		CurrentPhase: "design",
		FrayedFrom:   "parent-session-123",
		FrameRef:     "frame-abc",
		ParkSource:   "manual",
		ClaimedBy:    "cc-agent-xyz",
	}
	activeRite := "ecosystem"
	mode := "orchestrated"

	cases := []struct {
		field string
		want  string
	}{
		{"session_id", sessionID},
		{"status", "ACTIVE"},
		{"initiative", "test initiative"},
		{"complexity", "SYSTEM"},
		{"active_rite", "ecosystem"},
		{"execution_mode", "orchestrated"},
		{"current_phase", "design"},
		{"frayed_from", "parent-session-123"},
		{"frame_ref", "frame-abc"},
		{"park_source", "manual"},
		{"claimed_by", "cc-agent-xyz"},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			got, ok := getQueryField(sessCtx, activeRite, mode, tc.field)
			if !ok {
				t.Fatalf("getQueryField(%q) returned ok=false", tc.field)
			}
			if got != tc.want {
				t.Errorf("getQueryField(%q) = %q, want %q", tc.field, got, tc.want)
			}
		})
	}
}

func TestQuery_GetQueryField_UnknownField(t *testing.T) {
	sessCtx := &session.Context{}
	_, ok := getQueryField(sessCtx, "", "native", "nonexistent_field")
	if ok {
		t.Error("getQueryField(nonexistent) should return ok=false")
	}
}

// --- DeriveExecutionMode unit tests (canonical function in session package) ---

func TestQuery_DeriveExecutionMode(t *testing.T) {
	cases := []struct {
		name       string
		status     session.Status
		activeRite string
		want       string
	}{
		{"active with rite", session.StatusActive, "ecosystem", "orchestrated"},
		{"active empty rite", session.StatusActive, "", "cross-cutting"},
		{"active none rite", session.StatusActive, "none", "cross-cutting"},
		{"parked with rite", session.StatusParked, "ecosystem", "cross-cutting"},
		{"archived with rite", session.StatusArchived, "ecosystem", "native"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := session.DeriveExecutionMode(tc.status, tc.activeRite)
			if got != tc.want {
				t.Errorf("DeriveExecutionMode(%q, %q) = %q, want %q", tc.status, tc.activeRite, got, tc.want)
			}
		})
	}
}

// --- convertQueryStrands unit tests ---

func TestQuery_ConvertQueryStrands_Nil(t *testing.T) {
	if result := convertQueryStrands(nil); result != nil {
		t.Errorf("convertQueryStrands(nil) = %v, want nil", result)
	}
}

func TestQuery_ConvertQueryStrands_Empty(t *testing.T) {
	if result := convertQueryStrands([]session.Strand{}); result != nil {
		t.Errorf("convertQueryStrands([]) = %v, want nil", result)
	}
}

func TestQuery_ConvertQueryStrands_Populated(t *testing.T) {
	strands := []session.Strand{
		{SessionID: "child-001", Status: "ACTIVE", FrameRef: "frame-abc"},
		{SessionID: "child-002", Status: "PARKED", LandedAt: "2026-03-06T12:00:00Z"},
	}
	result := convertQueryStrands(strands)

	if len(result) != 2 {
		t.Fatalf("convertQueryStrands() len = %d, want 2", len(result))
	}
	if result[0].SessionID != "child-001" || result[0].FrameRef != "frame-abc" {
		t.Errorf("result[0] = %+v, unexpected", result[0])
	}
	if result[1].Status != "PARKED" || result[1].LandedAt != "2026-03-06T12:00:00Z" {
		t.Errorf("result[1] = %+v, unexpected", result[1])
	}
}

// --- QueryOutput.Text() unit tests ---

func TestQuery_OutputText_HasSession(t *testing.T) {
	q := output.QueryOutput{
		SessionID:     "session-20260306-122256-4fc1e1cc",
		Status:        "ACTIVE",
		Initiative:    "S3 query initiative",
		Complexity:    "MODULE",
		ActiveRite:    "ecosystem",
		ExecutionMode: "orchestrated",
		CurrentPhase:  "requirements",
		HasSession:    true,
	}

	text := q.Text()

	checks := []string{
		"---\n",
		"session_id: session-20260306-122256-4fc1e1cc",
		"status: ACTIVE",
		`initiative: "S3 query initiative"`,
		"active_rite: ecosystem",
		"execution_mode: orchestrated",
		"current_phase: requirements",
		"complexity: MODULE",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Errorf("Text() missing %q, output:\n%s", check, text)
		}
	}
	if !strings.HasSuffix(text, "---\n") {
		t.Errorf("Text() should end with '---\\n', got: %q", text[len(text)-10:])
	}
}

func TestQuery_OutputText_NoSession(t *testing.T) {
	q := output.QueryOutput{HasSession: false}
	text := q.Text()

	if !strings.Contains(text, "has_session: false") {
		t.Errorf("no-session Text() missing has_session: false")
	}
	// Should not contain any session-specific fields
	for _, absent := range []string{"session_id:", "status:", "initiative:", "active_rite:"} {
		if strings.Contains(text, absent) {
			t.Errorf("no-session Text() should not contain %q", absent)
		}
	}
}

func TestQuery_OutputText_OmitsEmptyOptionalFields(t *testing.T) {
	q := output.QueryOutput{
		SessionID:     "session-20260306-122256-4fc1e1cc",
		Status:        "ACTIVE",
		Initiative:    "test",
		ActiveRite:    "none",
		ExecutionMode: "cross-cutting",
		HasSession:    true,
		// All optional fields empty: FrayedFrom, FrameRef, ParkSource, ClaimedBy, Strands
	}

	text := q.Text()

	for _, absent := range []string{"frayed_from:", "frame_ref:", "park_source:", "claimed_by:", "strands:"} {
		if strings.Contains(text, absent) {
			t.Errorf("Text() should omit empty field %q, output:\n%s", absent, text)
		}
	}
}

func TestQuery_OutputText_WithStrands(t *testing.T) {
	q := output.QueryOutput{
		SessionID:     "session-20260306-122256-4fc1e1cc",
		Status:        "ACTIVE",
		Initiative:    "test",
		ActiveRite:    "ecosystem",
		ExecutionMode: "orchestrated",
		HasSession:    true,
		Strands: []output.QueryStrand{
			{SessionID: "child-001", Status: "ACTIVE"},
			{SessionID: "child-002", Status: "PARKED", FrameRef: "frame-xyz"},
		},
	}

	text := q.Text()

	if !strings.Contains(text, "strands:") {
		t.Errorf("Text() missing strands: section")
	}
	if !strings.Contains(text, "  - session_id: child-001") {
		t.Errorf("Text() missing child-001 strand")
	}
	if !strings.Contains(text, "    frame_ref: frame-xyz") {
		t.Errorf("Text() missing frame_ref on child-002")
	}
}
