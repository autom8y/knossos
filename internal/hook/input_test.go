package hook

import (
	"strings"
	"testing"
)

func TestParseToolInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(*testing.T, *ToolInput)
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
			checkFunc: func(t *testing.T, ti *ToolInput) {
				if !ti.IsEmpty() {
					t.Error("Expected empty ToolInput")
				}
			},
		},
		{
			name:    "simple bash command",
			input:   `{"command":"git status","description":"Show status"}`,
			wantErr: false,
			checkFunc: func(t *testing.T, ti *ToolInput) {
				if ti.Command != "git status" {
					t.Errorf("Expected command 'git status', got %q", ti.Command)
				}
				if ti.Description != "Show status" {
					t.Errorf("Expected description 'Show status', got %q", ti.Description)
				}
			},
		},
		{
			name:    "write tool input",
			input:   `{"file_path":"/tmp/test.txt","content":"hello world"}`,
			wantErr: false,
			checkFunc: func(t *testing.T, ti *ToolInput) {
				if ti.FilePath != "/tmp/test.txt" {
					t.Errorf("Expected file_path '/tmp/test.txt', got %q", ti.FilePath)
				}
				if ti.Content != "hello world" {
					t.Errorf("Expected content 'hello world', got %q", ti.Content)
				}
			},
		},
		{
			name:    "edit tool input",
			input:   `{"file_path":"/tmp/test.txt","old_string":"hello","new_string":"goodbye"}`,
			wantErr: false,
			checkFunc: func(t *testing.T, ti *ToolInput) {
				if ti.FilePath != "/tmp/test.txt" {
					t.Errorf("Expected file_path '/tmp/test.txt', got %q", ti.FilePath)
				}
				if ti.OldString != "hello" {
					t.Errorf("Expected old_string 'hello', got %q", ti.OldString)
				}
				if ti.NewString != "goodbye" {
					t.Errorf("Expected new_string 'goodbye', got %q", ti.NewString)
				}
			},
		},
		{
			name:    "glob pattern",
			input:   `{"pattern":"**/*.go","path":"/project"}`,
			wantErr: false,
			checkFunc: func(t *testing.T, ti *ToolInput) {
				if ti.Pattern != "**/*.go" {
					t.Errorf("Expected pattern '**/*.go', got %q", ti.Pattern)
				}
				if ti.Path != "/project" {
					t.Errorf("Expected path '/project', got %q", ti.Path)
				}
			},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseToolInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToolInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil && got != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func TestToolInputGet(t *testing.T) {
	input := `{"name":"test","count":42,"enabled":true,"nested":{"key":"value"}}`
	ti, err := ParseToolInput(input)
	if err != nil {
		t.Fatalf("ParseToolInput() error = %v", err)
	}

	t.Run("GetString", func(t *testing.T) {
		if got := ti.GetString("name"); got != "test" {
			t.Errorf("GetString(name) = %q, want 'test'", got)
		}
		if got := ti.GetString("missing"); got != "" {
			t.Errorf("GetString(missing) = %q, want ''", got)
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		if got := ti.GetInt("count"); got != 42 {
			t.Errorf("GetInt(count) = %d, want 42", got)
		}
		if got := ti.GetInt("missing"); got != 0 {
			t.Errorf("GetInt(missing) = %d, want 0", got)
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		if got := ti.GetBool("enabled"); !got {
			t.Error("GetBool(enabled) = false, want true")
		}
		if got := ti.GetBool("missing"); got {
			t.Error("GetBool(missing) = true, want false")
		}
	})

	t.Run("GetMap", func(t *testing.T) {
		nested := ti.GetMap("nested")
		if nested == nil {
			t.Fatal("GetMap(nested) = nil, want map")
		}
		if nested["key"] != "value" {
			t.Errorf("nested[key] = %v, want 'value'", nested["key"])
		}
	})
}

func TestToolInputGetEffectivePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "file_path takes precedence",
			input:    `{"file_path":"/a/b.txt","path":"/c/d"}`,
			expected: "/a/b.txt",
		},
		{
			name:     "fallback to path",
			input:    `{"path":"/c/d"}`,
			expected: "/c/d",
		},
		{
			name:     "empty when no path",
			input:    `{"command":"ls"}`,
			expected: "",
		},
		{
			name:     "file field fallback",
			input:    `{"file":"/e/f.txt"}`,
			expected: "/e/f.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti, err := ParseToolInput(tt.input)
			if err != nil {
				t.Fatalf("ParseToolInput() error = %v", err)
			}
			if got := ti.GetEffectivePath(); got != tt.expected {
				t.Errorf("GetEffectivePath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestToolInputString(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		ti, _ := ParseToolInput(`{"command":"ls"}`)
		got := ti.String()
		if !strings.Contains(got, "command") {
			t.Errorf("String() = %q, expected to contain 'command'", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		ti := &ToolInput{}
		if got := ti.String(); got != "{}" {
			t.Errorf("String() = %q, want '{}'", got)
		}
	})
}

func TestParseToolInputFromReader(t *testing.T) {
	input := `{"command":"test"}`
	reader := strings.NewReader(input)

	ti, err := ParseToolInputFromReader(reader)
	if err != nil {
		t.Fatalf("ParseToolInputFromReader() error = %v", err)
	}

	if ti.Command != "test" {
		t.Errorf("Command = %q, want 'test'", ti.Command)
	}
}
