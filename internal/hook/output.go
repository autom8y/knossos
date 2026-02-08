package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// OutputFormat represents the hook output format.
type OutputFormat int

const (
	// FormatJSON outputs structured JSON (for bash wrappers to parse).
	FormatJSON OutputFormat = iota
	// FormatText outputs human-readable text (for debugging).
	FormatText
)

// Decision represents a hook's decision about a tool operation.
type Decision string

const (
	// DecisionAllow permits the tool operation to proceed.
	DecisionAllow Decision = "allow"
	// DecisionBlock prevents the tool operation.
	DecisionBlock Decision = "block"
	// DecisionModify changes the tool input before execution.
	DecisionModify Decision = "modify"
)

// Result represents the structured output of a hook execution.
type Result struct {
	// Decision indicates what action to take (legacy field)
	Decision Decision `json:"decision"`

	// PermissionDecision is the CC-native field for PreToolUse hooks.
	// CC reads this field; value must be exactly "allow" or "deny".
	// Auto-populated from Decision field during JSON encoding.
	PermissionDecision string `json:"permissionDecision,omitempty"`

	// Reason explains the decision (for logging/debugging)
	Reason string `json:"reason,omitempty"`

	// Message is output to show the user
	Message string `json:"message,omitempty"`

	// ModifiedInput contains changed tool input (when Decision is Modify)
	ModifiedInput json.RawMessage `json:"modified_input,omitempty"`

	// Context contains additional data injected by the hook
	Context map[string]interface{} `json:"context,omitempty"`

	// Error contains error information if hook failed
	Error *HookError `json:"error,omitempty"`

	// Performance tracking
	DurationMs int64 `json:"duration_ms,omitempty"`
}

// PreToolUseOutput is the CC-native output format for PreToolUse hooks.
// CC expects decisions wrapped in a hookSpecificOutput envelope.
type PreToolUseOutput struct {
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput"`
}

// HookSpecificOutput contains the PreToolUse decision fields CC reads.
type HookSpecificOutput struct {
	HookEventName            string          `json:"hookEventName"`
	PermissionDecision       string          `json:"permissionDecision"`
	PermissionDecisionReason string          `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             json.RawMessage `json:"updatedInput,omitempty"`
}

// HookError represents an error from hook execution.
type HookError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Writer handles hook output formatting.
type Writer struct {
	format OutputFormat
	out    io.Writer
	err    io.Writer
}

// NewWriter creates a new hook output writer.
func NewWriter(format OutputFormat, out, errOut io.Writer) *Writer {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	return &Writer{
		format: format,
		out:    out,
		err:    errOut,
	}
}

// DefaultWriter returns a Writer with JSON format to stdout.
func DefaultWriter() *Writer {
	return NewWriter(FormatJSON, os.Stdout, os.Stderr)
}

// WriteResult outputs a hook result in the configured format.
func (w *Writer) WriteResult(r *Result) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(r)
	default:
		return w.writeText(r)
	}
}

// WriteAllow outputs an allow decision.
func (w *Writer) WriteAllow(reason string) error {
	return w.WriteResult(&Result{
		Decision: DecisionAllow,
		Reason:   reason,
	})
}

// WriteBlock outputs a block decision with a user message.
func (w *Writer) WriteBlock(reason, message string) error {
	return w.WriteResult(&Result{
		Decision: DecisionBlock,
		Reason:   reason,
		Message:  message,
	})
}

// WriteModify outputs a modify decision with changed input.
func (w *Writer) WriteModify(reason string, modifiedInput interface{}) error {
	var rawInput json.RawMessage
	if modifiedInput != nil {
		data, err := json.Marshal(modifiedInput)
		if err != nil {
			return w.WriteError("MARSHAL_ERROR", "failed to marshal modified input")
		}
		rawInput = data
	}
	return w.WriteResult(&Result{
		Decision:      DecisionModify,
		Reason:        reason,
		ModifiedInput: rawInput,
	})
}

// WriteError outputs an error result.
func (w *Writer) WriteError(code, message string) error {
	return w.WriteResult(&Result{
		Decision: DecisionAllow, // Errors should not block by default (graceful degradation)
		Error: &HookError{
			Code:    code,
			Message: message,
		},
	})
}

// WritePreToolUseAllow outputs an allow decision in CC's hookSpecificOutput format.
func (w *Writer) WritePreToolUseAllow(reason string) error {
	output := PreToolUseOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "allow",
			PermissionDecisionReason: reason,
		},
	}
	enc := json.NewEncoder(w.out)
	return enc.Encode(output)
}

// WritePreToolUseBlock outputs a deny decision in CC's hookSpecificOutput format.
func (w *Writer) WritePreToolUseBlock(reason string) error {
	output := PreToolUseOutput{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
	enc := json.NewEncoder(w.out)
	return enc.Encode(output)
}

// WriteContext outputs an allow decision with injected context.
func (w *Writer) WriteContext(context map[string]interface{}) error {
	return w.WriteResult(&Result{
		Decision: DecisionAllow,
		Context:  context,
	})
}

func (w *Writer) writeJSON(r *Result) error {
	// Auto-populate PermissionDecision from Decision for CC compatibility.
	// This ensures dual output: both legacy "decision" and CC-native "permissionDecision".
	switch r.Decision {
	case DecisionAllow:
		r.PermissionDecision = "allow"
	case DecisionBlock:
		r.PermissionDecision = "deny" // CC uses "deny", not "block"
	case DecisionModify:
		r.PermissionDecision = "allow" // CC does not support modify
	default:
		r.PermissionDecision = "allow" // Default to allow for unknown decisions
	}

	enc := json.NewEncoder(w.out)
	// Compact JSON for easier parsing by bash wrappers
	return enc.Encode(r)
}

func (w *Writer) writeText(r *Result) error {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Decision: %s\n", r.Decision))
	if r.Reason != "" {
		b.WriteString(fmt.Sprintf("Reason: %s\n", r.Reason))
	}
	if r.Message != "" {
		b.WriteString(fmt.Sprintf("Message: %s\n", r.Message))
	}
	if r.Error != nil {
		b.WriteString(fmt.Sprintf("Error: [%s] %s\n", r.Error.Code, r.Error.Message))
	}
	if r.DurationMs > 0 {
		b.WriteString(fmt.Sprintf("Duration: %dms\n", r.DurationMs))
	}

	_, err := fmt.Fprint(w.out, b.String())
	return err
}

// WriteDebug writes debug information to stderr.
func (w *Writer) WriteDebug(format string, args ...interface{}) {
	fmt.Fprintf(w.err, "[DEBUG] "+format+"\n", args...)
}

// Allow is a helper that creates an allow result.
func Allow(reason string) *Result {
	return &Result{Decision: DecisionAllow, Reason: reason}
}

// Block is a helper that creates a block result.
func Block(reason, message string) *Result {
	return &Result{Decision: DecisionBlock, Reason: reason, Message: message}
}

// Modify is a helper that creates a modify result.
func Modify(reason string, modifiedInput json.RawMessage) *Result {
	return &Result{Decision: DecisionModify, Reason: reason, ModifiedInput: modifiedInput}
}

// WithContext adds context to a result.
func (r *Result) WithContext(ctx map[string]interface{}) *Result {
	r.Context = ctx
	return r
}

// WithDuration adds timing information to a result.
func (r *Result) WithDuration(ms int64) *Result {
	r.DurationMs = ms
	return r
}
