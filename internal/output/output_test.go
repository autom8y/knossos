package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// newTestPrinter returns a Printer writing to a buffer for inspection.
func newTestPrinter(format Format) (*Printer, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return NewPrinter(format, out, errOut, false), out, errOut
}

// --- JSON output contract tests ---

func TestJSON_StatusOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := StatusOutput{
		HasSession:   true,
		SessionID:    "session-20260101-120000-abcd1234",
		Status:       "ACTIVE",
		Initiative:   "test initiative",
		Complexity:   "MODULE",
		CurrentPhase: "implementation",
		ActiveRite:   "ecosystem",
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded StatusOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if decoded.SessionID != data.SessionID {
		t.Errorf("session_id = %q, want %q", decoded.SessionID, data.SessionID)
	}
	if decoded.Status != data.Status {
		t.Errorf("status = %q, want %q", decoded.Status, data.Status)
	}
	if decoded.Initiative != data.Initiative {
		t.Errorf("initiative = %q, want %q", decoded.Initiative, data.Initiative)
	}
}

func TestJSON_CreateOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := CreateOutput{
		SessionID:  "session-20260101-120000-abcd1234",
		Status:     "ACTIVE",
		Initiative: "build feature X",
		Complexity: "SYSTEM",
		Rite:       "ecosystem",
		CreatedAt:  "2026-01-01T12:00:00Z",
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded CreateOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if decoded.SessionID != data.SessionID {
		t.Errorf("session_id = %q, want %q", decoded.SessionID, data.SessionID)
	}
	if decoded.Rite != data.Rite {
		t.Errorf("rite = %q, want %q", decoded.Rite, data.Rite)
	}
}

func TestJSON_SyncResultOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := SyncResultOutput{
		Status: "ok",
		DryRun: false,
		Rite: &SyncRiteResult{
			Status:   "ok",
			RiteName: "ecosystem",
			Source:   "embedded",
		},
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded SyncResultOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if decoded.Status != "ok" {
		t.Errorf("status = %q, want %q", decoded.Status, "ok")
	}
	if decoded.Rite == nil {
		t.Fatal("rite field is nil in decoded output")
	}
	if decoded.Rite.RiteName != "ecosystem" {
		t.Errorf("rite.rite = %q, want %q", decoded.Rite.RiteName, "ecosystem")
	}
}

func TestJSON_ErrorOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, _, errBuf := newTestPrinter(FormatJSON)

	// The printer wraps non-errors.Error in {"error": {"code": ..., "message": ...}}.
	type wrappedErr struct {
		Err struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	// Use a plain error (not *errors.Error) to exercise the generic fallback path.
	plainError := &plainErrorImpl{msg: "something went wrong"}
	p.PrintError(plainError)

	if !json.Valid(errBuf.Bytes()) {
		t.Errorf("error JSON output is not valid JSON: %s", errBuf.String())
	}

	var decoded wrappedErr
	if err := json.Unmarshal(errBuf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal error JSON output: %v", err)
	}

	if decoded.Err.Code != "GENERAL_ERROR" {
		t.Errorf("error.code = %q, want %q", decoded.Err.Code, "GENERAL_ERROR")
	}
	if decoded.Err.Message != "something went wrong" {
		t.Errorf("error.message = %q, want %q", decoded.Err.Message, "something went wrong")
	}
}

// plainErrorImpl is a plain Go error with no JSON() method — tests the fallback path.
type plainErrorImpl struct{ msg string }

func (e *plainErrorImpl) Error() string { return e.msg }

// --- Format dispatch tests ---

func TestFormatDispatch_TextProducesNoJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatText)

	data := StatusOutput{HasSession: false}
	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	// Text output should NOT be valid JSON (it's human-readable prose).
	if json.Valid(buf.Bytes()) && buf.Len() > 0 {
		t.Errorf("Text format produced valid JSON — expected human-readable text")
	}
}

func TestFormatDispatch_JSONProducesValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := StatusOutput{HasSession: false, Status: "none"}
	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON format produced invalid JSON: %s", buf.String())
	}
}

// --- ParseFormat and ValidateFormat ---

func TestParseFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"yaml", FormatYAML},
		{"YAML", FormatYAML},
		{"text", FormatText},
		{"", FormatText},
		{"unknown", FormatText},
	}

	for _, tt := range tests {
		got := ParseFormat(tt.input)
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"json", false},
		{"yaml", false},
		{"text", false},
		{"", false},
		{"invalid", true},
		{"xml", true},
	}

	for _, tt := range tests {
		err := ValidateFormat(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

// --- SessionListOutput contract ---

func TestJSON_SessionListOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := SessionListOutput{
		Sessions: []SessionSummary{
			{
				SessionID:  "session-20260101-120000-abcd1234",
				Status:     "ACTIVE",
				Initiative: "test",
				Complexity: "MODULE",
				CreatedAt:  "2026-01-01T12:00:00Z",
				Current:    true,
			},
		},
		Total:          1,
		CurrentSession: "session-20260101-120000-abcd1234",
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded SessionListOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Total != 1 {
		t.Errorf("total = %d, want 1", decoded.Total)
	}
	if len(decoded.Sessions) != 1 {
		t.Fatalf("sessions length = %d, want 1", len(decoded.Sessions))
	}
	if !decoded.Sessions[0].Current {
		t.Errorf("sessions[0].current = false, want true")
	}
}

// --- VerboseLog ---

func TestVerboseLog_OnlyWritesWhenVerbose(t *testing.T) {
	t.Parallel()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	quietPrinter := NewPrinter(FormatText, outBuf, errBuf, false)
	quietPrinter.VerboseLog("info", "quiet message", nil)
	if errBuf.Len() != 0 {
		t.Errorf("VerboseLog wrote to errOut when verbose=false")
	}

	verbosePrinter := NewPrinter(FormatText, outBuf, errBuf, true)
	verbosePrinter.VerboseLog("info", "verbose message", map[string]any{"key": "value"})
	if errBuf.Len() == 0 {
		t.Errorf("VerboseLog wrote nothing when verbose=true")
	}
	if !json.Valid(errBuf.Bytes()) {
		t.Errorf("VerboseLog output is not valid JSON: %s", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "verbose message") {
		t.Errorf("VerboseLog output missing message: %s", errBuf.String())
	}
}

// --- SyncResultOutput Text tests ---

func TestTextOutput_OrgSkipped_ShowsReason(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		status     string
		errMsg     string
		wantLabel  string
		wantAbsent string
	}{
		{
			name:       "org skipped shows Reason",
			status:     "skipped",
			errMsg:     "no active org configured",
			wantLabel:  "Reason:",
			wantAbsent: "Error:",
		},
		{
			name:       "org error shows Error",
			status:     "error",
			errMsg:     "permission denied",
			wantLabel:  "Error:",
			wantAbsent: "Reason:",
		},
		{
			name:       "rite skipped shows Reason",
			status:     "skipped",
			errMsg:     "no rite configured",
			wantLabel:  "Reason:",
			wantAbsent: "Error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var data SyncResultOutput
			if strings.HasPrefix(tt.name, "rite") {
				data = SyncResultOutput{
					Status: "ok",
					Rite: &SyncRiteResult{
						Status: tt.status,
						Error:  tt.errMsg,
					},
				}
			} else {
				data = SyncResultOutput{
					Status: "ok",
					Org: &SyncOrgResult{
						Status: tt.status,
						Error:  tt.errMsg,
					},
				}
			}

			text := data.Text()
			if !strings.Contains(text, tt.wantLabel) {
				t.Errorf("output missing %q; got:\n%s", tt.wantLabel, text)
			}
			if strings.Contains(text, tt.wantAbsent) {
				t.Errorf("output should not contain %q; got:\n%s", tt.wantAbsent, text)
			}
		})
	}
}

// --- RiteListOutput JSON contract ---

func TestJSON_RiteListOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := RiteListOutput{
		Rites: []RiteSummary{
			{
				Name:       "ecosystem",
				Form:       "full",
				AgentCount: 5,
				SkillCount: 3,
				Path:       "rites/ecosystem",
				Source:     "project",
				Active:     true,
			},
			{
				Name:       "releaser",
				Form:       "lite",
				AgentCount: 2,
				SkillCount: 1,
				Path:       "rites/releaser",
				Source:     "project",
				Active:     false,
			},
		},
		Total:      2,
		ActiveRite: "ecosystem",
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded RiteListOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Total != 2 {
		t.Errorf("total = %d, want 2", decoded.Total)
	}
	if decoded.ActiveRite != "ecosystem" {
		t.Errorf("active_rite = %q, want %q", decoded.ActiveRite, "ecosystem")
	}
	if len(decoded.Rites) != 2 {
		t.Fatalf("rites length = %d, want 2", len(decoded.Rites))
	}
	if !decoded.Rites[0].Active {
		t.Errorf("rites[0].active = false, want true")
	}
	if decoded.Rites[1].Active {
		t.Errorf("rites[1].active = true, want false")
	}
}

// --- AuditOutput JSON contract ---

func TestJSON_AuditOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := AuditOutput{
		SessionID: "session-20260101-120000-abcd1234",
		Events: []AuditEvent{
			{
				Timestamp: "2026-01-01T12:00:00Z",
				Event:     "state_transition",
				From:      "NONE",
				To:        "ACTIVE",
			},
			{
				Timestamp: "2026-01-01T13:00:00Z",
				Event:     "phase_transition",
				FromPhase: "planning",
				ToPhase:   "implementation",
				Metadata:  map[string]any{"reason": "design complete"},
			},
		},
		Total: 2,
		FiltersApplied: AuditFilters{
			Limit:     50,
			EventType: "",
			Since:     "",
		},
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded AuditOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.SessionID != data.SessionID {
		t.Errorf("session_id = %q, want %q", decoded.SessionID, data.SessionID)
	}
	if decoded.Total != 2 {
		t.Errorf("total = %d, want 2", decoded.Total)
	}
	if len(decoded.Events) != 2 {
		t.Fatalf("events length = %d, want 2", len(decoded.Events))
	}
	if decoded.Events[0].To != "ACTIVE" {
		t.Errorf("events[0].to = %q, want %q", decoded.Events[0].To, "ACTIVE")
	}
	if decoded.Events[1].ToPhase != "implementation" {
		t.Errorf("events[1].to_phase = %q, want %q", decoded.Events[1].ToPhase, "implementation")
	}
}

// --- ManifestShowOutput JSON contract ---

func TestJSON_ManifestShowOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := ManifestShowOutput{
		Path:   "knossos.yaml",
		Exists: true,
		Format: "yaml",
		Schema: &ManifestSchemaInfo{
			Type:    "rite-manifest",
			Version: "1.0",
			Valid:   true,
		},
		Content: map[string]any{
			"project": map[string]any{
				"name":        "knossos",
				"description": "test project",
			},
		},
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded ManifestShowOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Path != "knossos.yaml" {
		t.Errorf("path = %q, want %q", decoded.Path, "knossos.yaml")
	}
	if !decoded.Exists {
		t.Errorf("exists = false, want true")
	}
	if decoded.Schema == nil {
		t.Fatal("schema is nil")
	}
	if !decoded.Schema.Valid {
		t.Errorf("schema.valid = false, want true")
	}
}

// --- ManifestValidateOutput JSON contract ---

func TestJSON_ManifestValidateOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := ManifestValidateOutput{
		Path:   "knossos.yaml",
		Schema: "rite-manifest-v1",
		Valid:  false,
		Issues: []ManifestValidationIssue{
			{Path: "project.name", Message: "required field missing", Severity: "error"},
		},
		Warnings: []ManifestValidationIssue{
			{Path: "teams.available", Message: "empty list", Severity: "warning"},
		},
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded ManifestValidateOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Valid {
		t.Errorf("valid = true, want false")
	}
	if len(decoded.Issues) != 1 {
		t.Fatalf("issues length = %d, want 1", len(decoded.Issues))
	}
	if len(decoded.Warnings) != 1 {
		t.Fatalf("warnings length = %d, want 1", len(decoded.Warnings))
	}
}

// --- RiteInfoOutput JSON contract ---

func TestJSON_RiteInfoOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := RiteInfoOutput{
		Name:   "ecosystem",
		Form:   "full",
		Path:   "rites/ecosystem",
		Source: "project",
		Active: true,
		Agents: []RiteAgentInfo{
			{Name: "pythia", File: "pythia.md", Role: "orchestrator"},
		},
		Skills: []RiteSkillInfo{
			{Ref: "conventions", Path: ".claude/skills/conventions", External: false},
		},
		Workflow: &RiteWorkflowInfo{
			Type:       "orchestrated",
			EntryPoint: "pythia",
			Phases:     []string{"planning", "implementation", "review"},
		},
		Budget: &RiteBudgetInfo{
			EstimatedTokens: 15000,
			AgentsCost:      10000,
			SkillsCost:      3000,
			WorkflowCost:    2000,
		},
		SchemaVersion: "1.0",
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded RiteInfoOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != "ecosystem" {
		t.Errorf("name = %q, want %q", decoded.Name, "ecosystem")
	}
	if !decoded.Active {
		t.Errorf("active = false, want true")
	}
	if len(decoded.Agents) != 1 {
		t.Fatalf("agents length = %d, want 1", len(decoded.Agents))
	}
	if decoded.Workflow == nil {
		t.Fatal("workflow is nil")
	}
	if decoded.Budget == nil {
		t.Fatal("budget is nil")
	}
}

// --- TransitionOutput JSON contract ---

func TestJSON_TransitionOutput_ValidJSON(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)

	data := TransitionOutput{
		SessionID:      "session-20260101-120000-abcd1234",
		Status:         "ARCHIVED",
		PreviousStatus: "ACTIVE",
		SailsColor:     "WHITE",
		SailsBase:      "WHITE",
		Archived:       true,
		ArchivePath:    "/archive/session.tar.gz",
	}

	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("JSON output is not valid JSON: %s", buf.String())
	}

	var decoded TransitionOutput
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Status != "ARCHIVED" {
		t.Errorf("status = %q, want %q", decoded.Status, "ARCHIVED")
	}
	if decoded.SailsColor != "WHITE" {
		t.Errorf("sails_color = %q, want %q", decoded.SailsColor, "WHITE")
	}
}

// --- Text() smoke tests ---

func TestText_Smoke_NoPanic(t *testing.T) {
	t.Parallel()
	// Verify that Text() methods do not panic and produce non-empty output
	// for representative inputs. Table-driven to cover maximum output types.
	tests := []struct {
		name     string
		textable Textable
		wantNon  bool // if true, expect non-empty output
	}{
		{
			name:     "StatusOutput/active",
			textable: StatusOutput{HasSession: true, SessionID: "s1", Status: "ACTIVE", Initiative: "test", CurrentPhase: "impl", ActiveRite: "eco"},
			wantNon:  true,
		},
		{
			name:     "StatusOutput/no-session",
			textable: StatusOutput{HasSession: false},
			wantNon:  true,
		},
		{
			name:     "CreateOutput",
			textable: CreateOutput{SessionID: "s1", Initiative: "x", Complexity: "M", Rite: "eco"},
			wantNon:  true,
		},
		{
			name:     "SyncResultOutput/full",
			textable: SyncResultOutput{Status: "ok", Rite: &SyncRiteResult{Status: "ok", RiteName: "eco"}, Org: &SyncOrgResult{Status: "ok", OrgName: "team"}, User: &SyncUserResult{Status: "ok"}},
			wantNon:  true,
		},
		{
			name:     "SyncResultOutput/dry-run",
			textable: SyncResultOutput{Status: "ok", DryRun: true},
			wantNon:  true,
		},
		{
			name:     "RiteListOutput/empty",
			textable: RiteListOutput{},
			wantNon:  true,
		},
		{
			name:     "RiteListOutput/populated",
			textable: RiteListOutput{Rites: []RiteSummary{{Name: "eco", Active: true}}, Total: 1},
			wantNon:  true,
		},
		{
			name:     "RiteInfoOutput",
			textable: RiteInfoOutput{Name: "eco", Form: "full", Path: "p", Source: "project", Active: true, Agents: []RiteAgentInfo{{Name: "a", Role: "r"}}},
			wantNon:  true,
		},
		{
			name:     "RiteCurrentOutput/no-rite",
			textable: RiteCurrentOutput{},
			wantNon:  true,
		},
		{
			name:     "RiteCurrentOutput/active",
			textable: RiteCurrentOutput{ActiveRite: "eco", NativeAgents: []string{"a"}, Budget: CurrentBudgetOutput{TotalTokens: 1000, BudgetLimit: 50000}},
			wantNon:  true,
		},
		{
			name:     "RiteStatusOutput",
			textable: RiteStatusOutput{Rite: "eco", IsActive: true, Path: "p", Description: "d", WorkflowType: "orchestrated", EntryPoint: "pythia", ManifestValid: true, ClaudeMDSynced: true},
			wantNon:  true,
		},
		{
			name:     "RiteValidateOutput",
			textable: RiteValidateOutput{Rite: "eco", Valid: true, Checks: []ValidationCheckOut{{Check: "schema", Status: "pass", Message: "ok"}}},
			wantNon:  true,
		},
		{
			name:     "PantheonOutput",
			textable: PantheonOutput{Rite: "eco", Agents: []PantheonAgent{{Name: "a"}}, Count: 1},
			wantNon:  true,
		},
		{
			name:     "ManifestShowOutput/exists",
			textable: ManifestShowOutput{Path: "k.yaml", Exists: true, Format: "yaml"},
			wantNon:  true,
		},
		{
			name:     "ManifestShowOutput/missing",
			textable: ManifestShowOutput{Path: "k.yaml", Exists: false},
			wantNon:  true,
		},
		{
			name:     "ManifestValidateOutput",
			textable: ManifestValidateOutput{Path: "k.yaml", Schema: "v1", Valid: true},
			wantNon:  true,
		},
		{
			name:     "ManifestDiffOutput/no-changes",
			textable: ManifestDiffOutput{HasChanges: false},
			wantNon:  true,
		},
		{
			name:     "ManifestDiffOutput/with-changes",
			textable: ManifestDiffOutput{HasChanges: true, Base: "a", Compare: "b", Changes: []ManifestDiffChange{{Path: "p", Type: "added", NewValue: "v"}}},
			wantNon:  true,
		},
		{
			name:     "ManifestMergeOutput",
			textable: ManifestMergeOutput{Base: "a", Ours: "b", Theirs: "c", Strategy: "ours"},
			wantNon:  true,
		},
		{
			name:     "TimelineOutput/empty",
			textable: TimelineOutput{SessionID: "s1"},
			wantNon:  true,
		},
		{
			name:     "TimelineOutput/populated",
			textable: TimelineOutput{SessionID: "s1", Entries: []TimelineEntryOutput{{Time: "12:00", Category: "state", Summary: "created"}}},
			wantNon:  true,
		},
		{
			name:     "FrayOutput",
			textable: FrayOutput{ParentID: "p1", ChildID: "c1", FrayPoint: "fp", Status: "ok"},
			wantNon:  true,
		},
		{
			name:     "LogOutput",
			textable: LogOutput{SessionID: "s1", Type: "note", Entry: "logged"},
			wantNon:  true,
		},
		{
			name:     "FieldOutput",
			textable: FieldOutput{Key: "status", Value: "ACTIVE"},
			wantNon:  true,
		},
		{
			name:     "FieldAllOutput",
			textable: FieldAllOutput{SessionID: "s1", Status: "ACTIVE"},
			wantNon:  true,
		},
		{
			name:     "TransitionOutput/archived",
			textable: TransitionOutput{SessionID: "s1", Status: "ARCHIVED", SailsColor: "WHITE"},
			wantNon:  true,
		},
		{
			name:     "TransitionOutput/black-sails",
			textable: TransitionOutput{SessionID: "s1", Status: "ARCHIVED", SailsColor: "BLACK", SailsReasons: []string{"tests failed"}},
			wantNon:  true,
		},
		{
			name:     "TransitionOutput/gray-sails",
			textable: TransitionOutput{SessionID: "s1", Status: "ARCHIVED", SailsColor: "GRAY"},
			wantNon:  true,
		},
		{
			name:     "SeedCreateOutput",
			textable: SeedCreateOutput{SessionID: "s1", Status: "PARKED", Seeded: true, SeededTo: "/tmp/seed"},
			wantNon:  true,
		},
		{
			name:     "SnapshotOutput",
			textable: SnapshotOutput{Markdown: "# Snapshot\nContent here"},
			wantNon:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.textable.Text()
			if tt.wantNon && got == "" {
				t.Errorf("Text() returned empty string, want non-empty")
			}
		})
	}
}

// --- Tabular interface tests ---

func TestTabular_Headers_NonEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tabular Tabular
	}{
		{"SessionListOutput", SessionListOutput{Sessions: []SessionSummary{{SessionID: "s1"}}, Total: 1}},
		{"AuditOutput", AuditOutput{Events: []AuditEvent{{Event: "e1"}}}},
		{"RiteListOutput", RiteListOutput{Rites: []RiteSummary{{Name: "eco"}}}},
		{"RiteStatusOutput", RiteStatusOutput{Rite: "eco"}},
		{"RiteSwitchOutput", RiteSwitchOutput{Rite: "eco"}},
		{"RiteSwitchDryRunOutput", RiteSwitchDryRunOutput{WouldSwitchTo: "eco"}},
		{"RiteValidateOutput", RiteValidateOutput{Rite: "eco"}},
		{"PantheonOutput", PantheonOutput{Rite: "eco"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			headers := tt.tabular.Headers()
			if len(headers) == 0 {
				t.Error("Headers() returned empty slice")
			}
		})
	}
}

func TestTabular_Rows_MatchHeaders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tabular Tabular
	}{
		{
			"SessionListOutput",
			SessionListOutput{
				Sessions: []SessionSummary{
					{SessionID: "s1", Status: "ACTIVE", Initiative: "test", CreatedAt: "2026-01-01T12:00:00Z", Current: true},
				},
				Total: 1,
			},
		},
		{
			"AuditOutput",
			AuditOutput{
				Events: []AuditEvent{
					{Timestamp: "2026-01-01T12:00:00Z", Event: "state_transition", From: "NONE", To: "ACTIVE"},
				},
			},
		},
		{
			"RiteListOutput",
			RiteListOutput{
				Rites: []RiteSummary{
					{Name: "eco", Form: "full", AgentCount: 5, SkillCount: 3, Source: "project", Active: true},
				},
			},
		},
		{
			"PantheonOutput",
			PantheonOutput{
				Rite:   "eco",
				Agents: []PantheonAgent{{Name: "pythia", Model: "opus", Description: "orchestrator"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			headers := tt.tabular.Headers()
			rows := tt.tabular.Rows()
			if len(rows) == 0 {
				t.Fatal("Rows() returned empty slice")
			}
			for i, row := range rows {
				if len(row) != len(headers) {
					t.Errorf("row[%d] has %d columns, want %d (matching headers)", i, len(row), len(headers))
				}
			}
		})
	}
}

// --- PrintSuccess tests ---

func TestPrintSuccess_JSONProducesOutput(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON)
	data := StatusOutput{HasSession: true, Status: "ACTIVE"}
	if err := p.PrintSuccess(data); err != nil {
		t.Fatalf("PrintSuccess() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("PrintSuccess() produced no output in JSON mode")
	}
	if !json.Valid(buf.Bytes()) {
		t.Errorf("PrintSuccess() JSON output is not valid: %s", buf.String())
	}
}

func TestPrintSuccess_TextIsSilent(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatText)
	data := StatusOutput{HasSession: true, Status: "ACTIVE"}
	if err := p.PrintSuccess(data); err != nil {
		t.Fatalf("PrintSuccess() error = %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("PrintSuccess() wrote output in text mode, expected silent: %s", buf.String())
	}
}

// --- PrintError text mode test ---

func TestPrintError_TextMode(t *testing.T) {
	t.Parallel()
	p, _, errBuf := newTestPrinter(FormatText)
	err := &plainErrorImpl{msg: "test error"}
	p.PrintError(err)
	if !strings.Contains(errBuf.String(), "test error") {
		t.Errorf("PrintError() text output missing error message: %s", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Error:") {
		t.Errorf("PrintError() text output missing 'Error:' prefix: %s", errBuf.String())
	}
}

// --- PrintText and PrintLine ---

func TestPrintText_WritesRaw(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON) // format should not matter for PrintText
	p.PrintText("raw output")
	if buf.String() != "raw output" {
		t.Errorf("PrintText() = %q, want %q", buf.String(), "raw output")
	}
}

func TestPrintLine_WritesWithNewline(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatJSON) // format should not matter for PrintLine
	p.PrintLine("a line")
	if buf.String() != "a line\n" {
		t.Errorf("PrintLine() = %q, want %q", buf.String(), "a line\n")
	}
}

// --- truncateDescription ---

func TestTruncateDescription(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"Short description.", "Short description."},
		{"First sentence. Second sentence.", "First sentence."},
		{strings.Repeat("a", 100), strings.Repeat("a", 77) + "..."},
		{"Line one\nLine two\nLine three", "Line one Line two"},
	}

	for _, tt := range tests {
		got := truncateDescription(tt.input)
		if got != tt.want {
			t.Errorf("truncateDescription(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- printTable via Printer ---

func TestPrintTable_RendersTabularData(t *testing.T) {
	t.Parallel()
	p, buf, _ := newTestPrinter(FormatText)

	data := RiteListOutput{
		Rites: []RiteSummary{
			{Name: "ecosystem", Form: "full", AgentCount: 5, SkillCount: 3, Source: "project", Active: true},
		},
		Total: 1,
	}

	// Text print should go through printTable because RiteListOutput implements Tabular
	if err := p.Print(data); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ecosystem") {
		t.Errorf("table output missing rite name: %s", output)
	}
	if !strings.Contains(output, "RITE") {
		t.Errorf("table output missing header: %s", output)
	}
}

// TestTextOutput_RiteSwitchOrphans verifies that orphan output uses descriptive
// rite-switch phrasing when RiteSwitched is true.
func TestTextOutput_RiteSwitchOrphans(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		riteSwitched bool
		previousRite string
		riteName     string
		orphanAction string
		wantContains string
		wantAbsent   string
	}{
		{
			name:         "rite switch shows replaced message",
			riteSwitched: true,
			previousRite: "releaser",
			riteName:     "10x-dev",
			orphanAction: "removed",
			wantContains: "Agents: 3 replaced (rite switch: releaser -> 10x-dev)",
			wantAbsent:   "Orphans:",
		},
		{
			name:         "non-switch shows orphans message",
			riteSwitched: false,
			riteName:     "10x-dev",
			orphanAction: "removed",
			wantContains: "Orphans: 3 detected (removed)",
			wantAbsent:   "Agents:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := SyncResultOutput{
				Status: "success",
				Rite: &SyncRiteResult{
					Status:          "success",
					RiteName:        tt.riteName,
					OrphansDetected: []string{"a.md", "b.md", "c.md"},
					OrphanAction:    tt.orphanAction,
					RiteSwitched:    tt.riteSwitched,
					PreviousRite:    tt.previousRite,
				},
			}
			text := data.Text()
			if !strings.Contains(text, tt.wantContains) {
				t.Errorf("output missing %q; got:\n%s", tt.wantContains, text)
			}
			if strings.Contains(text, tt.wantAbsent) {
				t.Errorf("output should not contain %q; got:\n%s", tt.wantAbsent, text)
			}
		})
	}
}
