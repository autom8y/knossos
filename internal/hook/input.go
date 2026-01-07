package hook

import (
	"encoding/json"
	"io"
	"os"

	"github.com/autom8y/knossos/internal/errors"
)

// ToolInput represents the parsed JSON input from Claude Code.
type ToolInput struct {
	// Raw JSON data
	Raw json.RawMessage

	// Common fields extracted from tool input
	Path        string `json:"path,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	Command     string `json:"command,omitempty"`
	Content     string `json:"content,omitempty"`
	OldString   string `json:"old_string,omitempty"`
	NewString   string `json:"new_string,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
	Query       string `json:"query,omitempty"`
	Description string `json:"description,omitempty"`

	// For nested or tool-specific data
	data map[string]interface{}
}

// ParseToolInput parses JSON tool input from a string.
func ParseToolInput(jsonStr string) (*ToolInput, error) {
	if jsonStr == "" {
		return &ToolInput{}, nil
	}
	return ParseToolInputBytes([]byte(jsonStr))
}

// ParseToolInputBytes parses JSON tool input from bytes.
func ParseToolInputBytes(data []byte) (*ToolInput, error) {
	input := &ToolInput{
		Raw:  data,
		data: make(map[string]interface{}),
	}

	// Parse into structured fields
	if err := json.Unmarshal(data, input); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse tool input JSON", err)
	}

	// Also parse into generic map for arbitrary field access
	if err := json.Unmarshal(data, &input.data); err != nil {
		// Non-fatal: structured fields were parsed successfully
		input.data = make(map[string]interface{})
	}

	return input, nil
}

// ParseToolInputFromReader parses JSON tool input from an io.Reader.
func ParseToolInputFromReader(r io.Reader) (*ToolInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read tool input", err)
	}
	return ParseToolInputBytes(data)
}

// ParseToolInputFromStdin parses JSON tool input from stdin.
func ParseToolInputFromStdin() (*ToolInput, error) {
	// Check if there's data on stdin
	stat, err := os.Stdin.Stat()
	if err != nil {
		return &ToolInput{}, nil
	}

	// If stdin is a terminal (no pipe), return empty input
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return &ToolInput{}, nil
	}

	return ParseToolInputFromReader(os.Stdin)
}

// Get returns a field value by key from the parsed data.
func (t *ToolInput) Get(key string) interface{} {
	if t.data == nil {
		return nil
	}
	return t.data[key]
}

// GetString returns a string field value by key.
func (t *ToolInput) GetString(key string) string {
	val := t.Get(key)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetBool returns a boolean field value by key.
func (t *ToolInput) GetBool(key string) bool {
	val := t.Get(key)
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetInt returns an integer field value by key.
func (t *ToolInput) GetInt(key string) int {
	val := t.Get(key)
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

// GetMap returns a map field value by key.
func (t *ToolInput) GetMap(key string) map[string]interface{} {
	val := t.Get(key)
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// GetSlice returns a slice field value by key.
func (t *ToolInput) GetSlice(key string) []interface{} {
	val := t.Get(key)
	if s, ok := val.([]interface{}); ok {
		return s
	}
	return nil
}

// GetEffectivePath returns the most likely path field from tool input.
// Different tools use different field names for paths.
func (t *ToolInput) GetEffectivePath() string {
	// Try common path field names in order of specificity
	if t.FilePath != "" {
		return t.FilePath
	}
	if t.Path != "" {
		return t.Path
	}
	// Fall back to generic lookup
	for _, key := range []string{"file_path", "path", "file", "target"} {
		if v := t.GetString(key); v != "" {
			return v
		}
	}
	return ""
}

// IsEmpty returns true if the tool input is empty or nil.
func (t *ToolInput) IsEmpty() bool {
	return t == nil || len(t.Raw) == 0
}

// String returns the raw JSON string representation.
func (t *ToolInput) String() string {
	if t.IsEmpty() {
		return "{}"
	}
	return string(t.Raw)
}
