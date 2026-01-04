package hook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriterWriteResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *Result
		format   OutputFormat
		contains []string
	}{
		{
			name: "allow decision json",
			result: &Result{
				Decision: DecisionAllow,
				Reason:   "test reason",
			},
			format:   FormatJSON,
			contains: []string{`"decision":"allow"`, `"reason":"test reason"`},
		},
		{
			name: "block decision json",
			result: &Result{
				Decision: DecisionBlock,
				Reason:   "blocked",
				Message:  "Cannot proceed",
			},
			format:   FormatJSON,
			contains: []string{`"decision":"block"`, `"message":"Cannot proceed"`},
		},
		{
			name: "allow decision text",
			result: &Result{
				Decision: DecisionAllow,
				Reason:   "test reason",
			},
			format:   FormatText,
			contains: []string{"Decision: allow", "Reason: test reason"},
		},
		{
			name: "with error json",
			result: &Result{
				Decision: DecisionAllow,
				Error: &HookError{
					Code:    "TEST_ERROR",
					Message: "Something went wrong",
				},
			},
			format:   FormatJSON,
			contains: []string{`"code":"TEST_ERROR"`, `"message":"Something went wrong"`},
		},
		{
			name: "with duration",
			result: &Result{
				Decision:   DecisionAllow,
				DurationMs: 42,
			},
			format:   FormatJSON,
			contains: []string{`"duration_ms":42`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewWriter(tt.format, &buf, nil)

			err := w.WriteResult(tt.result)
			if err != nil {
				t.Fatalf("WriteResult() error = %v", err)
			}

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestWriterHelpers(t *testing.T) {
	t.Run("WriteAllow", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(FormatJSON, &buf, nil)

		err := w.WriteAllow("allowed")
		if err != nil {
			t.Fatalf("WriteAllow() error = %v", err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if result.Decision != DecisionAllow {
			t.Errorf("Decision = %v, want %v", result.Decision, DecisionAllow)
		}
		if result.Reason != "allowed" {
			t.Errorf("Reason = %v, want 'allowed'", result.Reason)
		}
	})

	t.Run("WriteBlock", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(FormatJSON, &buf, nil)

		err := w.WriteBlock("security", "Operation blocked")
		if err != nil {
			t.Fatalf("WriteBlock() error = %v", err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if result.Decision != DecisionBlock {
			t.Errorf("Decision = %v, want %v", result.Decision, DecisionBlock)
		}
		if result.Message != "Operation blocked" {
			t.Errorf("Message = %v, want 'Operation blocked'", result.Message)
		}
	})

	t.Run("WriteModify", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(FormatJSON, &buf, nil)

		modified := map[string]string{"command": "safe-ls"}
		err := w.WriteModify("sanitized", modified)
		if err != nil {
			t.Fatalf("WriteModify() error = %v", err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if result.Decision != DecisionModify {
			t.Errorf("Decision = %v, want %v", result.Decision, DecisionModify)
		}
		if result.ModifiedInput == nil {
			t.Error("ModifiedInput is nil")
		}
	})

	t.Run("WriteError", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(FormatJSON, &buf, nil)

		err := w.WriteError("HOOK_FAILED", "Hook crashed")
		if err != nil {
			t.Fatalf("WriteError() error = %v", err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		// Errors should allow by default (graceful degradation)
		if result.Decision != DecisionAllow {
			t.Errorf("Decision = %v, want %v (graceful degradation)", result.Decision, DecisionAllow)
		}
		if result.Error == nil {
			t.Fatal("Error is nil")
		}
		if result.Error.Code != "HOOK_FAILED" {
			t.Errorf("Error.Code = %v, want 'HOOK_FAILED'", result.Error.Code)
		}
	})

	t.Run("WriteContext", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(FormatJSON, &buf, nil)

		ctx := map[string]interface{}{
			"session_id": "test-123",
			"team":       "dev",
		}
		err := w.WriteContext(ctx)
		if err != nil {
			t.Fatalf("WriteContext() error = %v", err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if result.Context["session_id"] != "test-123" {
			t.Errorf("Context[session_id] = %v, want 'test-123'", result.Context["session_id"])
		}
	})
}

func TestResultHelpers(t *testing.T) {
	t.Run("Allow", func(t *testing.T) {
		r := Allow("test")
		if r.Decision != DecisionAllow {
			t.Errorf("Decision = %v, want %v", r.Decision, DecisionAllow)
		}
	})

	t.Run("Block", func(t *testing.T) {
		r := Block("reason", "message")
		if r.Decision != DecisionBlock {
			t.Errorf("Decision = %v, want %v", r.Decision, DecisionBlock)
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		r := Allow("test").WithContext(map[string]interface{}{"key": "value"})
		if r.Context["key"] != "value" {
			t.Error("Context not set correctly")
		}
	})

	t.Run("WithDuration", func(t *testing.T) {
		r := Allow("test").WithDuration(50)
		if r.DurationMs != 50 {
			t.Errorf("DurationMs = %d, want 50", r.DurationMs)
		}
	})
}

func TestDefaultWriter(t *testing.T) {
	w := DefaultWriter()
	if w == nil {
		t.Fatal("DefaultWriter() returned nil")
	}
	if w.format != FormatJSON {
		t.Errorf("Default format = %v, want JSON", w.format)
	}
}
