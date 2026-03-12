package channel_test

import (
	"testing"

	"github.com/autom8y/knossos/internal/channel"
)

// TestCCToGeminiTool_Coverage verifies both key spaces (frontmatter and wire-protocol)
// are present in the map.
func TestCCToGeminiTool_Coverage(t *testing.T) {
	t.Parallel()

	required := []struct {
		key      string
		expected string
	}{
		{"Read", "read_file"},       // agent frontmatter name
		{"ReadFiles", "read_file"},  // hook wire-protocol name
		{"Edit", "replace"},
		{"Write", "write_file"},
		{"Bash", "run_shell_command"},
		{"Glob", "glob"},
		{"Grep", "search_files"},
	}

	for _, tc := range required {
		got, ok := channel.CCToGeminiTool[tc.key]
		if !ok {
			t.Errorf("CCToGeminiTool[%q] missing", tc.key)
			continue
		}
		if got != tc.expected {
			t.Errorf("CCToGeminiTool[%q] = %q, want %q", tc.key, got, tc.expected)
		}
	}
}

// TestCCOnlyTools_AllPresent verifies the CC-only tool set is complete.
func TestCCOnlyTools_AllPresent(t *testing.T) {
	t.Parallel()

	required := []string{"Task", "TodoWrite", "Skill", "WebSearch", "WebFetch", "NotebookEdit"}
	for _, tool := range required {
		if !channel.CCOnlyTools[tool] {
			t.Errorf("CCOnlyTools missing %q", tool)
		}
	}
}

// TestTranslateTool_KnownMapping verifies successful translation of known tools.
func TestTranslateTool_KnownMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected string
	}{
		{"Read", "read_file"},
		{"ReadFiles", "read_file"},
		{"Bash", "run_shell_command"},
		{"Edit", "replace"},
		{"Write", "write_file"},
		{"Glob", "glob"},
		{"Grep", "search_files"},
	}

	for _, tc := range cases {
		got, ok := channel.TranslateTool(tc.input)
		if !ok {
			t.Errorf("TranslateTool(%q): ok=false, want true", tc.input)
		}
		if got != tc.expected {
			t.Errorf("TranslateTool(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

// TestTranslateTool_CCOnlyDropped verifies CC-only tools return ("", false).
func TestTranslateTool_CCOnlyDropped(t *testing.T) {
	t.Parallel()

	ccOnly := []string{"Task", "TodoWrite", "Skill", "WebSearch", "WebFetch", "NotebookEdit"}
	for _, tool := range ccOnly {
		got, ok := channel.TranslateTool(tool)
		if ok {
			t.Errorf("TranslateTool(%q): ok=true, want false (CC-only tool should be dropped)", tool)
		}
		if got != "" {
			t.Errorf("TranslateTool(%q) = %q, want \"\" for CC-only tool", tool, got)
		}
	}
}

// TestTranslateTool_UnknownPassthrough verifies unknown tools pass through unchanged.
func TestTranslateTool_UnknownPassthrough(t *testing.T) {
	t.Parallel()

	unknowns := []string{"CustomTool", "FutureTool", "SomePlugin"}
	for _, tool := range unknowns {
		got, ok := channel.TranslateTool(tool)
		if !ok {
			t.Errorf("TranslateTool(%q): ok=false, want true for unknown tool passthrough", tool)
		}
		if got != tool {
			t.Errorf("TranslateTool(%q) = %q, want %q (passthrough)", tool, got, tool)
		}
	}
}

// TestTranslateFrontmatterTools_MixedInput verifies the slice translation helper.
func TestTranslateFrontmatterTools_MixedInput(t *testing.T) {
	t.Parallel()

	input := []string{"Read", "Bash", "Task", "Skill", "Edit", "CustomTool"}
	// Expected: Read->read_file, Bash->run_shell_command, Task dropped, Skill dropped,
	//           Edit->replace, CustomTool passes through
	got := channel.TranslateFrontmatterTools(input)

	expected := []string{"read_file", "run_shell_command", "replace", "CustomTool"}
	if len(got) != len(expected) {
		t.Fatalf("TranslateFrontmatterTools: len=%d, want %d; got %v", len(got), len(expected), got)
	}
	for i, e := range expected {
		if got[i] != e {
			t.Errorf("TranslateFrontmatterTools[%d] = %q, want %q", i, got[i], e)
		}
	}
}

// TestTranslateFrontmatterTools_NilInput verifies nil input returns nil.
func TestTranslateFrontmatterTools_NilInput(t *testing.T) {
	t.Parallel()

	if got := channel.TranslateFrontmatterTools(nil); got != nil {
		t.Errorf("TranslateFrontmatterTools(nil) = %v, want nil", got)
	}
}

// TestTranslateFrontmatterTools_EmptyInput verifies empty input returns empty slice.
func TestTranslateFrontmatterTools_EmptyInput(t *testing.T) {
	t.Parallel()

	got := channel.TranslateFrontmatterTools([]string{})
	if got == nil {
		t.Error("TranslateFrontmatterTools([]): returned nil, want empty slice")
	}
	if len(got) != 0 {
		t.Errorf("TranslateFrontmatterTools([]): len=%d, want 0", len(got))
	}
}

// TestTranslateFrontmatterTools_AllCCOnly verifies all-CC-only input produces empty slice.
func TestTranslateFrontmatterTools_AllCCOnly(t *testing.T) {
	t.Parallel()

	got := channel.TranslateFrontmatterTools([]string{"Task", "TodoWrite", "Skill"})
	if len(got) != 0 {
		t.Errorf("expected empty slice when all tools are CC-only, got %v", got)
	}
}
