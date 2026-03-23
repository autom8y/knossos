package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

func TestDetectToolFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantHit  bool
		wantTool string
	}{
		{
			name:     "grep command",
			input:    `{"command": "grep -r 'TODO' src/"}`,
			wantHit:  true,
			wantTool: "Grep",
		},
		{
			name:     "rg command",
			input:    `{"command": "rg pattern file.go"}`,
			wantHit:  true,
			wantTool: "Grep",
		},
		{
			name:     "cat command",
			input:    `{"command": "cat /path/to/file.go"}`,
			wantHit:  true,
			wantTool: "Read",
		},
		{
			name:     "find command",
			input:    `{"command": "find . -name '*.go'"}`,
			wantHit:  true,
			wantTool: "Glob",
		},
		{
			name:     "sed command",
			input:    `{"command": "sed -i 's/old/new/g' file.go"}`,
			wantHit:  true,
			wantTool: "Edit",
		},
		{
			name:    "normal bash command",
			input:   `{"command": "go build ./cmd/ari"}`,
			wantHit: false,
		},
		{
			name:    "git command",
			input:   `{"command": "git status"}`,
			wantHit: false,
		},
		{
			name:    "ari command",
			input:   `{"command": "ari sync"}`,
			wantHit: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantHit: false,
		},
		{
			name:    "invalid JSON",
			input:   "not json",
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := detectToolFallback(tt.input)
			if tt.wantHit && result == "" {
				t.Error("expected fallback detection, got empty")
			}
			if !tt.wantHit && result != "" {
				t.Errorf("expected no fallback, got: %s", result)
			}
			if tt.wantHit && tt.wantTool != "" && !strings.Contains(result, tt.wantTool) {
				t.Errorf("expected result to mention %s, got: %s", tt.wantTool, result)
			}
		})
	}
}

func TestDetectRetrySpiralFromState(t *testing.T) {
	t.Parallel()

	t.Run("no spiral with successes", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputHash: "abc", Success: true},
				{Tool: "Bash", InputHash: "abc", Success: true},
				{Tool: "Bash", InputHash: "abc", Success: true},
			},
		}
		if result := detectRetrySpiralFromState(state); result != "" {
			t.Errorf("expected no spiral, got: %s", result)
		}
	})

	t.Run("spiral with 3 consecutive failures", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputHash: "abc", InputSnippet: "ari ask 'switch rite'", Success: false},
				{Tool: "Bash", InputHash: "abc", InputSnippet: "ari ask 'switch rite'", Success: false},
				{Tool: "Bash", InputHash: "abc", InputSnippet: "ari ask 'switch rite'", Success: false},
			},
		}
		result := detectRetrySpiralFromState(state)
		if result == "" {
			t.Error("expected spiral detection")
		}
		if !strings.Contains(result, "Bash") {
			t.Errorf("expected result to mention tool, got: %s", result)
		}
	})

	t.Run("no spiral with mixed tools", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputHash: "abc", Success: false},
				{Tool: "Read", InputHash: "def", Success: false},
				{Tool: "Bash", InputHash: "abc", Success: false},
			},
		}
		if result := detectRetrySpiralFromState(state); result != "" {
			t.Errorf("expected no spiral with mixed tools, got: %s", result)
		}
	})

	t.Run("not enough calls", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputHash: "abc", Success: false},
				{Tool: "Bash", InputHash: "abc", Success: false},
			},
		}
		if result := detectRetrySpiralFromState(state); result != "" {
			t.Errorf("expected no spiral with too few calls, got: %s", result)
		}
	})

	t.Run("spiral only checks last N calls", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputHash: "old", Success: true},
				{Tool: "Bash", InputHash: "old", Success: true},
				{Tool: "Grep", InputHash: "xyz", Success: false},
				{Tool: "Grep", InputHash: "xyz", Success: false},
				{Tool: "Grep", InputHash: "xyz", Success: false},
			},
		}
		result := detectRetrySpiralFromState(state)
		if result == "" {
			t.Error("expected spiral in last 3 calls")
		}
	})
}

func TestDetectCommandExplorationFromState(t *testing.T) {
	t.Parallel()

	t.Run("exploration with 3 ari variations", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputSnippet: "ari rite list"},
				{Tool: "Bash", InputSnippet: "ari rites"},
				{Tool: "Bash", InputSnippet: "ari list-rites"},
			},
		}
		result := detectCommandExplorationFromState(state)
		if result == "" {
			t.Error("expected exploration detection")
		}
		if !strings.Contains(result, "3") {
			t.Errorf("expected result to mention count, got: %s", result)
		}
	})

	t.Run("no exploration with non-ari commands", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputSnippet: "go build ./cmd/ari"},
				{Tool: "Bash", InputSnippet: "go test ./..."},
				{Tool: "Bash", InputSnippet: "git status"},
			},
		}
		if result := detectCommandExplorationFromState(state); result != "" {
			t.Errorf("expected no exploration for non-ari commands, got: %s", result)
		}
	})

	t.Run("no exploration with too few ari commands", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputSnippet: "ari sync"},
				{Tool: "Bash", InputSnippet: "ari ask 'help'"},
			},
		}
		if result := detectCommandExplorationFromState(state); result != "" {
			t.Errorf("expected no exploration with too few commands, got: %s", result)
		}
	})

	t.Run("no exploration when same command repeated", func(t *testing.T) {
		t.Parallel()
		state := &DriftState{
			RecentCalls: []DriftCall{
				{Tool: "Bash", InputSnippet: "ari sync"},
				{Tool: "Bash", InputSnippet: "ari sync"},
				{Tool: "Bash", InputSnippet: "ari sync"},
			},
		}
		if result := detectCommandExplorationFromState(state); result != "" {
			t.Errorf("expected no exploration when same command repeated, got: %s", result)
		}
	})
}

func TestFileDriftComplaint(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fixedTime := time.Date(2026, 3, 11, 14, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedTime }

	env := &hook.Env{
		ToolName:   "Bash",
		ProjectDir: tmpDir,
	}

	path, err := fileDriftComplaint(env, "retry-spiral", "Bash failed 3 times with similar input", nowFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("complaint file not found: %v", err)
	}

	content := string(data)

	// Check required fields
	if !strings.Contains(content, "id: COMPLAINT-20260311-143000-drift-detector") {
		t.Error("missing or wrong id")
	}
	if !strings.Contains(content, "filed_by: drift-detector") {
		t.Error("missing filed_by")
	}
	if !strings.Contains(content, "severity: medium") {
		t.Error("missing severity")
	}
	if !strings.Contains(content, "status: filed") {
		t.Error("missing status")
	}
	if !strings.Contains(content, "retry-spiral") {
		t.Error("missing pattern tag")
	}

	// Verify directory was created
	complaintsDir := filepath.Join(tmpDir, ".sos", "wip", "complaints")
	if _, err := os.Stat(complaintsDir); os.IsNotExist(err) {
		t.Error("complaints directory was not created")
	}
}

func TestFileDriftComplaint_ToolFallbackSeverity(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fixedTime := time.Date(2026, 3, 11, 14, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedTime }

	env := &hook.Env{
		ToolName:   "Bash",
		ProjectDir: tmpDir,
	}

	path, err := fileDriftComplaint(env, "tool-fallback", "used Bash 'grep...' instead of Grep tool", nowFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("complaint file not found: %v", err)
	}

	if !strings.Contains(string(data), "severity: low") {
		t.Error("tool-fallback should have low severity")
	}
}

func TestExtractInputSnippet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tool     string
		input    string
		contains string
	}{
		{
			name:     "bash command",
			tool:     "Bash",
			input:    `{"command": "go build ./cmd/ari"}`,
			contains: "go build",
		},
		{
			name:     "read file",
			tool:     "Read",
			input:    `{"file_path": "/path/to/file.go"}`,
			contains: "read:/path/to/file.go",
		},
		{
			name:     "grep pattern",
			tool:     "Grep",
			input:    `{"pattern": "func main"}`,
			contains: "grep:func main",
		},
		{
			name:     "empty input",
			tool:     "Bash",
			input:    "",
			contains: "Bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractInputSnippet(tt.tool, tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("extractInputSnippet(%s, ...) = %q, want to contain %q", tt.tool, result, tt.contains)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestDriftState_LoadSave(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "drift-state.json")

	// Load from non-existent file returns empty state
	state := loadDriftState(statePath)
	if len(state.RecentCalls) != 0 {
		t.Error("expected empty state for missing file")
	}

	// Save and reload
	state.RecentCalls = []DriftCall{
		{Tool: "Bash", InputHash: "abc", InputSnippet: "go build", Success: true, At: "2026-03-11T14:00:00Z"},
	}
	var stdout, stderr bytes.Buffer
	p := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)
	saveDriftState(statePath, state, p)

	loaded := loadDriftState(statePath)
	if len(loaded.RecentCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(loaded.RecentCalls))
	}
	if loaded.RecentCalls[0].Tool != "Bash" {
		t.Errorf("tool = %q, want Bash", loaded.RecentCalls[0].Tool)
	}
}

func TestExtractBashCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"valid command", `{"command": "go build ./cmd/ari"}`, "go build ./cmd/ari"},
		{"empty input", "", ""},
		{"invalid JSON", "not json", ""},
		{"no command field", `{"file_path": "/tmp/x"}`, ""},
		{"whitespace trimmed", `{"command": "  git status  "}`, "git status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractBashCommand(tt.input)
			if got != tt.want {
				t.Errorf("extractBashCommand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHashInput(t *testing.T) {
	t.Parallel()

	// Same input should produce same hash
	h1 := hashInput("go build ./cmd/ari")
	h2 := hashInput("go build ./cmd/ari")
	if h1 != h2 {
		t.Error("same input should produce same hash")
	}

	// Different input should produce different hash
	h3 := hashInput("go test ./...")
	if h1 == h3 {
		t.Error("different input should produce different hash")
	}

	// Hash should be 8 hex chars (4 bytes)
	if len(h1) != 8 {
		t.Errorf("hash length = %d, want 8", len(h1))
	}
}

func TestDriftOutput_Text(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		out  DriftOutput
		want string
	}{
		{"message only", DriftOutput{Message: "no drift"}, "no drift"},
		{"filed complaint", DriftOutput{Message: "tool fallback detected", Filed: true}, "tool fallback detected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.out.Text()
			if got != tt.want {
				t.Errorf("Text() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRunDriftdetectCore_ToolFallback tests the full hook flow for tool fallback detection.
func TestRunDriftdetectCore_ToolFallback(t *testing.T) {
	tmpDir := t.TempDir()
	fixedTime := time.Date(2026, 3, 11, 14, 30, 0, 0, time.UTC)

	// Pipe stdin with a Bash tool event containing a grep command
	pipeHookStdinWithTool(t, string(hook.EventPostTool), tmpDir, "Bash",
		`{"command": "grep -r 'TODO' src/"}`)

	ctx := &cmdContext{}
	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	if err := runDriftdetectCore(nil, ctx, printer, func() time.Time { return fixedTime }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result DriftOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v (%s)", err, stdout.String())
	}

	if result.Pattern != "tool-fallback" {
		t.Errorf("pattern = %q, want tool-fallback", result.Pattern)
	}
	if !result.Filed {
		t.Error("expected complaint to be filed")
	}
}

// pipeHookStdinWithTool sets up os.Stdin with a JSON payload including tool_name for hook testing.
func pipeHookStdinWithTool(t *testing.T, event, projectDir, toolName, toolInput string) {
	t.Helper()
	oldStdin := os.Stdin
	payload := map[string]any{
		"hook_event_name": event,
		"cwd":             projectDir,
		"session_id":      "",
		"tool_name":       toolName,
	}
	if toolInput != "" {
		payload["tool_input"] = json.RawMessage(toolInput)
	}
	data, _ := json.Marshal(payload)
	r, w, _ := os.Pipe()
	go func() {
		w.Write(data)
		w.Close()
	}()
	os.Stdin = r
	os.Setenv("CLAUDE_PROJECT_DIR", projectDir)
	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	})
}
