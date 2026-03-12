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
		{"Read", "read_file"},              // agent frontmatter name
		{"ReadFiles", "read_file"},         // hook wire-protocol name
		{"Edit", "replace"},
		{"Write", "write_file"},
		{"Bash", "run_shell_command"},
		{"Glob", "glob"},
		{"Grep", "grep_search"},
		{"WebSearch", "google_web_search"},
		{"WebFetch", "web_fetch"},
		{"TodoWrite", "write_todos"},
		{"Skill", "activate_skill"},
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

	required := []string{"Task", "NotebookEdit"}
	for _, tool := range required {
		if !channel.CCOnlyTools[tool] {
			t.Errorf("CCOnlyTools missing %q", tool)
		}
	}

	// Verify tools that WERE CC-only but now have Gemini equivalents are NOT in CCOnlyTools.
	promoted := []string{"WebSearch", "WebFetch", "TodoWrite", "Skill"}
	for _, tool := range promoted {
		if channel.CCOnlyTools[tool] {
			t.Errorf("CCOnlyTools still contains %q — should be in CCToGeminiTool", tool)
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
		{"Grep", "grep_search"},
		{"WebSearch", "google_web_search"},
		{"WebFetch", "web_fetch"},
		{"TodoWrite", "write_todos"},
		{"Skill", "activate_skill"},
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

	ccOnly := []string{"Task", "NotebookEdit"}
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

	input := []string{"Read", "Bash", "Task", "Skill", "Edit", "CustomTool", "WebSearch"}
	// Expected: Read->read_file, Bash->run_shell_command, Task dropped,
	//           Skill->activate_skill, Edit->replace, CustomTool passes through,
	//           WebSearch->google_web_search
	got := channel.TranslateFrontmatterTools(input)

	expected := []string{"read_file", "run_shell_command", "activate_skill", "replace", "CustomTool", "google_web_search"}
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

	got := channel.TranslateFrontmatterTools([]string{"Task", "NotebookEdit"})
	if len(got) != 0 {
		t.Errorf("expected empty slice when all tools are CC-only, got %v", got)
	}
}

// --- Canonical tool vocabulary tests ---

func TestCanonicalToWireTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		canonical string
		channel   string
		wantWire  string
		wantOK    bool
	}{
		// Known canonical -> claude wire names
		{"read_file", "claude", "Read", true},
		{"edit_file", "claude", "Edit", true},
		{"write_file", "claude", "Write", true},
		{"list_files", "claude", "Glob", true},
		{"search_content", "claude", "Grep", true},
		{"run_shell", "claude", "Bash", true},
		{"web_search", "claude", "WebSearch", true},
		{"web_fetch", "claude", "WebFetch", true},
		{"write_todos", "claude", "TodoWrite", true},
		{"activate_skill", "claude", "Skill", true},
		{"delegate", "claude", "Task", true},
		// Known canonical -> gemini wire names
		{"read_file", "gemini", "read_file", true},
		{"edit_file", "gemini", "replace", true},
		{"write_file", "gemini", "write_file", true},
		{"list_files", "gemini", "glob", true},
		{"search_content", "gemini", "grep_search", true},
		{"run_shell", "gemini", "run_shell_command", true},
		{"web_search", "gemini", "google_web_search", true},
		{"web_fetch", "gemini", "web_fetch", true},
		{"write_todos", "gemini", "write_todos", true},
		{"activate_skill", "gemini", "activate_skill", true},
		// delegate has no gemini equivalent
		{"delegate", "gemini", "", false},
		// Unknown canonical passes through
		{"custom_tool", "claude", "custom_tool", true},
		{"custom_tool", "gemini", "custom_tool", true},
	}

	for _, tt := range tests {
		t.Run(tt.canonical+"/"+tt.channel, func(t *testing.T) {
			t.Parallel()
			got, ok := channel.CanonicalToWireTool(tt.canonical, tt.channel)
			if ok != tt.wantOK {
				t.Errorf("CanonicalToWireTool(%q, %q): ok=%v, want %v", tt.canonical, tt.channel, ok, tt.wantOK)
			}
			if got != tt.wantWire {
				t.Errorf("CanonicalToWireTool(%q, %q) = %q, want %q", tt.canonical, tt.channel, got, tt.wantWire)
			}
		})
	}
}

func TestWireToCanonicalTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wire          string
		wantCanonical string
		wantOK        bool
	}{
		// CC wire names
		{"Read", "read_file", true},
		{"Edit", "edit_file", true},
		{"Write", "write_file", true},
		{"Glob", "list_files", true},
		{"Grep", "search_content", true},
		{"Bash", "run_shell", true},
		{"WebSearch", "web_search", true},
		{"WebFetch", "web_fetch", true},
		{"TodoWrite", "write_todos", true},
		{"Skill", "activate_skill", true},
		{"Task", "delegate", true},
		// Gemini wire names
		{"read_file", "read_file", true},
		{"replace", "edit_file", true},
		{"run_shell_command", "run_shell", true},
		{"grep_search", "search_content", true},
		{"google_web_search", "web_search", true},
		// Unknown wire name passes through
		{"CustomPlugin", "CustomPlugin", false},
	}

	for _, tt := range tests {
		t.Run(tt.wire, func(t *testing.T) {
			t.Parallel()
			got, ok := channel.WireToCanonicalTool(tt.wire)
			if ok != tt.wantOK {
				t.Errorf("WireToCanonicalTool(%q): ok=%v, want %v", tt.wire, ok, tt.wantOK)
			}
			if got != tt.wantCanonical {
				t.Errorf("WireToCanonicalTool(%q) = %q, want %q", tt.wire, got, tt.wantCanonical)
			}
		})
	}
}

// TestCanonicalTool_Completeness verifies every entry in CCToGeminiTool has
// a corresponding canonical mapping, ensuring the two vocabularies stay aligned.
func TestCanonicalTool_Completeness(t *testing.T) {
	t.Parallel()

	// Wire-protocol aliases: CC hook names that map to the same tool as a
	// frontmatter name. These are CC-internal aliases, not separate canonical tools.
	wireAliases := map[string]string{
		"ReadFiles": "Read", // hook wire-protocol alias for Read
	}

	for cc, gemini := range channel.CCToGeminiTool {
		// If cc is a known wire alias, verify its target is in the canonical map instead.
		if target, isAlias := wireAliases[cc]; isAlias {
			if _, ok := channel.WireToCanonicalTool(target); !ok {
				t.Errorf("wire alias %q -> %q, but target has no canonical mapping", cc, target)
			}
			continue
		}

		ccCanonical, ccOK := channel.WireToCanonicalTool(cc)
		geminiCanonical, geminiOK := channel.WireToCanonicalTool(gemini)

		if !ccOK {
			t.Errorf("CC tool %q has no canonical mapping", cc)
			continue
		}
		if !geminiOK {
			t.Errorf("Gemini tool %q (from CC %q) has no canonical mapping", gemini, cc)
			continue
		}
		if ccCanonical != geminiCanonical {
			t.Errorf("CC %q -> canonical %q, but Gemini %q -> canonical %q (should match)",
				cc, ccCanonical, gemini, geminiCanonical)
		}
	}
}
