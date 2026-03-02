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
	p, buf, _ := newTestPrinter(FormatJSON)

	data := CreateOutput{
		SessionID: "session-20260101-120000-abcd1234",
		Status:    "ACTIVE",
		Initiative: "build feature X",
		Complexity: "SYSTEM",
		Rite:      "ecosystem",
		CreatedAt: "2026-01-01T12:00:00Z",
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
	if pErr := p.PrintError(plainError); pErr != nil {
		t.Fatalf("PrintError() error = %v", pErr)
	}

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
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	quietPrinter := NewPrinter(FormatText, outBuf, errBuf, false)
	quietPrinter.VerboseLog("info", "quiet message", nil)
	if errBuf.Len() != 0 {
		t.Errorf("VerboseLog wrote to errOut when verbose=false")
	}

	verbosePrinter := NewPrinter(FormatText, outBuf, errBuf, true)
	verbosePrinter.VerboseLog("info", "verbose message", map[string]interface{}{"key": "value"})
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
