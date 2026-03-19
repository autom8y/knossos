package channel_test

import (
	"testing"

	"github.com/autom8y/knossos/internal/channel"
)

// TestTranslateTool_CCOnlyBehavior verifies tools with a claude entry but no gemini
// entry in CanonicalTool are treated as channel-only and dropped.
func TestTranslateTool_CCOnlyBehavior(t *testing.T) {
	t.Parallel()

	// "delegate" canonical entry has claude="Task" but no gemini entry.
	// TranslateTool("Task") must return ("", false).
	got, ok := channel.TranslateTool("Task")
	if ok {
		t.Errorf("TranslateTool(Task): ok=true, want false (no gemini entry in CanonicalTool)")
	}
	if got != "" {
		t.Errorf("TranslateTool(Task) = %q, want \"\"", got)
	}

	// Verify tools that have Gemini equivalents are NOT treated as CC-only.
	translated := []string{"WebSearch", "WebFetch", "TodoWrite", "Skill"}
	for _, tool := range translated {
		_, ok := channel.TranslateTool(tool)
		if !ok {
			t.Errorf("TranslateTool(%q): ok=false, tool has a Gemini equivalent", tool)
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

// TestTranslateTool_CCOnlyDropped verifies tools with a claude entry but no gemini
// entry in CanonicalTool return ("", false) — they are dropped during translation.
func TestTranslateTool_CCOnlyDropped(t *testing.T) {
	t.Parallel()

	// "Task" maps to the "delegate" canonical entry which has claude="Task" but no gemini.
	got, ok := channel.TranslateTool("Task")
	if ok {
		t.Errorf("TranslateTool(Task): ok=true, want false (no gemini wire in CanonicalTool)")
	}
	if got != "" {
		t.Errorf("TranslateTool(Task) = %q, want \"\"", got)
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

// TestTranslateTool_MCP verifies MCP tool translation.
func TestTranslateTool_MCP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected string
	}{
		{"mcp:github/create_issue", "mcp_github_create_issue"},
		{"mcp:browserbase/session_create", "mcp_browserbase_session_create"},
		{"mcp:server-name/tool-name", "mcp_server-name_tool-name"},
		{"mcp:server", "mcp_server"},
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

// TestTranslateFrontmatterTools_MixedInput verifies the slice translation helper.
func TestTranslateFrontmatterTools_MixedInput(t *testing.T) {
	t.Parallel()

	input := []string{"Read", "Bash", "Task", "Skill", "Edit", "CustomTool", "WebSearch", "mcp:github/create_issue"}
	// Expected: Read->read_file, Bash->run_shell_command, Task dropped,
	//           Skill->activate_skill, Edit->replace, CustomTool passes through,
	//           WebSearch->google_web_search, mcp:github/create_issue->mcp_github_create_issue
	got := channel.TranslateFrontmatterTools(input)

	expected := []string{"read_file", "run_shell_command", "activate_skill", "replace", "CustomTool", "google_web_search", "mcp_github_create_issue"}
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

	// "Task" has a claude entry in CanonicalTool but no gemini entry — it is dropped.
	got := channel.TranslateFrontmatterTools([]string{"Task"})
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

// TestCanonicalTool_Completeness verifies that every canonical entry with a "gemini"
// wire name also has a "claude" wire name, and that TranslateTool round-trips correctly
// for all bidirectional tools.
func TestCanonicalTool_Completeness(t *testing.T) {
	t.Parallel()

	for canonical, wires := range channel.CanonicalTool {
		ccWire, hasCC := wires["claude"]
		geminiWire, hasGemini := wires["gemini"]

		if !hasCC && !hasGemini {
			t.Errorf("canonical %q has no wire entries for any channel", canonical)
			continue
		}

		if hasCC && hasGemini {
			// Bidirectional: TranslateTool(ccWire) must return (geminiWire, true).
			got, ok := channel.TranslateTool(ccWire)
			if !ok {
				t.Errorf("TranslateTool(%q) [canonical %q]: ok=false, want true", ccWire, canonical)
			}
			if got != geminiWire {
				t.Errorf("TranslateTool(%q) [canonical %q] = %q, want %q", ccWire, canonical, got, geminiWire)
			}
		}

		if hasCC && !hasGemini {
			// CC-only: TranslateTool(ccWire) must return ("", false).
			got, ok := channel.TranslateTool(ccWire)
			if ok {
				t.Errorf("TranslateTool(%q) [canonical %q, CC-only]: ok=true, want false", ccWire, canonical)
			}
			if got != "" {
				t.Errorf("TranslateTool(%q) [canonical %q, CC-only] = %q, want \"\"", ccWire, canonical, got)
			}
		}
	}
}
