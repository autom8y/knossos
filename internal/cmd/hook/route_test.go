package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"
)

func TestRouteOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   RouteOutput
		expected string
	}{
		{
			name: "routed with args",
			output: RouteOutput{
				Routed:   true,
				Command:  "/start",
				Args:     "Add dark mode",
				Category: CategorySession,
			},
			expected: "Routed: /start Add dark mode (session)",
		},
		{
			name: "routed without args",
			output: RouteOutput{
				Routed:   true,
				Command:  "/wrap",
				Args:     "",
				Category: CategorySession,
			},
			expected: "Routed: /wrap (session)",
		},
		{
			name:     "not routed",
			output:   RouteOutput{Routed: false},
			expected: "not routed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			if result != tt.expected {
				t.Errorf("Text() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseSlashCommand(t *testing.T) {
	tests := []struct {
		name             string
		message          string
		expectedCommand  string
		expectedArgs     string
		expectedCategory CommandCategory
	}{
		// Session commands
		{
			name:             "start with args",
			message:          "/start Add dark mode toggle",
			expectedCommand:  "/start",
			expectedArgs:     "Add dark mode toggle",
			expectedCategory: CategorySession,
		},
		{
			name:             "park without args",
			message:          "/park",
			expectedCommand:  "/park",
			expectedArgs:     "",
			expectedCategory: CategorySession,
		},
		{
			name:             "resume with args",
			message:          "/resume session-123",
			expectedCommand:  "/resume",
			expectedArgs:     "session-123",
			expectedCategory: CategorySession,
		},
		{
			name:             "wrap without args",
			message:          "/wrap",
			expectedCommand:  "/wrap",
			expectedArgs:     "",
			expectedCategory: CategorySession,
		},
		// Orchestrator commands
		{
			name:             "consult with question",
			message:          "/consult Which team for database work?",
			expectedCommand:  "/consult",
			expectedArgs:     "Which team for database work?",
			expectedCategory: CategoryOrchestrator,
		},
		// Initiative commands
		{
			name:             "task with description",
			message:          "/task Implement user authentication",
			expectedCommand:  "/task",
			expectedArgs:     "Implement user authentication",
			expectedCategory: CategoryInitiative,
		},
		{
			name:             "sprint with scope",
			message:          "/sprint Q1 feature set",
			expectedCommand:  "/sprint",
			expectedArgs:     "Q1 feature set",
			expectedCategory: CategoryInitiative,
		},
		// Git commands
		{
			name:             "commit with message",
			message:          "/commit Fix authentication bug",
			expectedCommand:  "/commit",
			expectedArgs:     "Fix authentication bug",
			expectedCategory: CategoryGit,
		},
		{
			name:             "pr without args",
			message:          "/pr",
			expectedCommand:  "/pr",
			expectedArgs:     "",
			expectedCategory: CategoryGit,
		},
		// Thread commands
		{
			name:             "stamp with decision",
			message:          "/stamp Use PostgreSQL over MongoDB for ACID compliance",
			expectedCommand:  "/stamp",
			expectedArgs:     "Use PostgreSQL over MongoDB for ACID compliance",
			expectedCategory: CategoryClew,
		},
		{
			name:             "stamp without args",
			message:          "/stamp",
			expectedCommand:  "/stamp",
			expectedArgs:     "",
			expectedCategory: CategoryClew,
		},
		// Case insensitivity
		{
			name:             "uppercase command",
			message:          "/START Add feature",
			expectedCommand:  "/start",
			expectedArgs:     "Add feature",
			expectedCategory: CategorySession,
		},
		{
			name:             "mixed case command",
			message:          "/CoNsUlT help me",
			expectedCommand:  "/consult",
			expectedArgs:     "help me",
			expectedCategory: CategoryOrchestrator,
		},
		// Whitespace handling
		{
			name:             "leading whitespace",
			message:          "   /start Initialize",
			expectedCommand:  "/start",
			expectedArgs:     "Initialize",
			expectedCategory: CategorySession,
		},
		{
			name:             "multiple spaces in args",
			message:          "/task   Build   feature",
			expectedCommand:  "/task",
			expectedArgs:     "Build   feature",
			expectedCategory: CategoryInitiative,
		},
		// Non-routed cases
		{
			name:             "regular message",
			message:          "Help me with this code",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
		{
			name:             "empty message",
			message:          "",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
		{
			name:             "only whitespace",
			message:          "   ",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
		{
			name:             "slash in middle of message",
			message:          "Please run /start for me",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
		{
			name:             "unknown command",
			message:          "/unknown command here",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
		{
			name:             "partial command match",
			message:          "/star",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
		{
			name:             "command prefix without space",
			message:          "/startNow",
			expectedCommand:  "",
			expectedArgs:     "",
			expectedCategory: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, cat := parseSlashCommand(tt.message)
			if cmd != tt.expectedCommand {
				t.Errorf("command = %q, want %q", cmd, tt.expectedCommand)
			}
			if args != tt.expectedArgs {
				t.Errorf("args = %q, want %q", args, tt.expectedArgs)
			}
			if cat != tt.expectedCategory {
				t.Errorf("category = %q, want %q", cat, tt.expectedCategory)
			}
		})
	}
}

func TestRunRoute_SessionCommands(t *testing.T) {
	commands := []struct {
		command  string
		category CommandCategory
	}{
		{"/start", CategorySession},
		{"/park", CategorySession},
		{"/resume", CategorySession},
		{"/wrap", CategorySession},
	}

	for _, tc := range commands {
		t.Run(tc.command, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "UserPromptSubmit",
				UserMessage: tc.command + " some args",
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			ctx := &cmdContext{
				output:  &outputFlag,
				verbose: &verboseFlag,
			}

			err := runRouteWithPrinter(ctx, printer)
			if err != nil {
				t.Fatalf("runRoute() error = %v", err)
			}

			var result RouteOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
			}

			if !result.Routed {
				t.Error("Expected Routed=true")
			}
			if result.Command != tc.command {
				t.Errorf("Command = %q, want %q", result.Command, tc.command)
			}
			if result.Category != tc.category {
				t.Errorf("Category = %q, want %q", result.Category, tc.category)
			}
		})
	}
}

func TestRunRoute_OrchestratorCommands(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "UserPromptSubmit",
		UserMessage: "/consult Which workflow should I use?",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runRouteWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runRoute() error = %v", err)
	}

	var result RouteOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Routed {
		t.Error("Expected Routed=true")
	}
	if result.Command != "/consult" {
		t.Errorf("Command = %q, want %q", result.Command, "/consult")
	}
	if result.Category != CategoryOrchestrator {
		t.Errorf("Category = %q, want %q", result.Category, CategoryOrchestrator)
	}
	if result.Args != "Which workflow should I use?" {
		t.Errorf("Args = %q, want %q", result.Args, "Which workflow should I use?")
	}
}

func TestRunRoute_InitiativeCommands(t *testing.T) {
	commands := []string{"/task", "/sprint"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "UserPromptSubmit",
				UserMessage: cmd + " Feature implementation",
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			ctx := &cmdContext{
				output:  &outputFlag,
				verbose: &verboseFlag,
			}

			err := runRouteWithPrinter(ctx, printer)
			if err != nil {
				t.Fatalf("runRoute() error = %v", err)
			}

			var result RouteOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
			}

			if !result.Routed {
				t.Error("Expected Routed=true")
			}
			if result.Command != cmd {
				t.Errorf("Command = %q, want %q", result.Command, cmd)
			}
			if result.Category != CategoryInitiative {
				t.Errorf("Category = %q, want %q", result.Category, CategoryInitiative)
			}
		})
	}
}

func TestRunRoute_GitCommands(t *testing.T) {
	commands := []string{"/commit", "/pr"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "UserPromptSubmit",
				UserMessage: cmd,
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			ctx := &cmdContext{
				output:  &outputFlag,
				verbose: &verboseFlag,
			}

			err := runRouteWithPrinter(ctx, printer)
			if err != nil {
				t.Fatalf("runRoute() error = %v", err)
			}

			var result RouteOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
			}

			if !result.Routed {
				t.Error("Expected Routed=true")
			}
			if result.Command != cmd {
				t.Errorf("Command = %q, want %q", result.Command, cmd)
			}
			if result.Category != CategoryGit {
				t.Errorf("Category = %q, want %q", result.Category, CategoryGit)
			}
		})
	}
}

func TestRunRoute_ThreadCommands(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "UserPromptSubmit",
		UserMessage: "/stamp Chose PostgreSQL for ACID compliance",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runRouteWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runRoute() error = %v", err)
	}

	var result RouteOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.Routed {
		t.Error("Expected Routed=true")
	}
	if result.Command != "/stamp" {
		t.Errorf("Command = %q, want %q", result.Command, "/stamp")
	}
	if result.Category != CategoryClew {
		t.Errorf("Category = %q, want %q", result.Category, CategoryClew)
	}
	if result.Args != "Chose PostgreSQL for ACID compliance" {
		t.Errorf("Args = %q, want %q", result.Args, "Chose PostgreSQL for ACID compliance")
	}
}

func TestRunRoute_RegularMessage(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "UserPromptSubmit",
		UserMessage: "Help me understand this code",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runRouteWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runRoute() error = %v", err)
	}

	var result RouteOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Routed {
		t.Error("Expected Routed=false for regular message")
	}
}

func TestRunRoute_EmptyMessage(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "UserPromptSubmit",
		UserMessage: "",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runRouteWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runRoute() error = %v", err)
	}

	var result RouteOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Routed {
		t.Error("Expected Routed=false for empty message")
	}
}

func TestRunRoute_WrongEventType(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
		UserMessage: "/start Initialize session",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runRouteWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runRoute() error = %v", err)
	}

	var result RouteOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Routed {
		t.Error("Expected Routed=false for wrong event type")
	}
}

func TestRunRoute_UnknownCommand(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "UserPromptSubmit",
		UserMessage: "/unknown do something",
		UseAriHooks: true,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	err := runRouteWithPrinter(ctx, printer)
	if err != nil {
		t.Fatalf("runRoute() error = %v", err)
	}

	var result RouteOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Routed {
		t.Error("Expected Routed=false for unknown command")
	}
}

func TestRunRoute_PartialCommandMatch(t *testing.T) {
	tests := []string{
		"/star",      // Partial match of /start
		"/startNow",  // Command prefix without space
		"/commit-it", // Command with suffix
	}

	for _, msg := range tests {
		t.Run(msg, func(t *testing.T) {
			testutil.SetupEnv(t, &testutil.HookEnv{
				Event:       "UserPromptSubmit",
				UserMessage: msg,
				UseAriHooks: true,
			})

			var stdout, stderr bytes.Buffer
			printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

			outputFlag := "json"
			verboseFlag := false
			ctx := &cmdContext{
				output:  &outputFlag,
				verbose: &verboseFlag,
			}

			err := runRouteWithPrinter(ctx, printer)
			if err != nil {
				t.Fatalf("runRoute() error = %v", err)
			}

			var result RouteOutput
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
			}

			if result.Routed {
				t.Errorf("Expected Routed=false for partial match %q", msg)
			}
		})
	}
}

// BenchmarkRouteHook_SlashCommand benchmarks slash command routing (<5ms target).
func BenchmarkRouteHook_SlashCommand(b *testing.B) {
	os.Setenv("USE_ARI_HOOKS", "1")
	os.Setenv("CLAUDE_HOOK_EVENT", "UserPromptSubmit")
	os.Setenv("CLAUDE_USER_MESSAGE", "/start Add dark mode toggle to settings")
	defer func() {
		os.Unsetenv("USE_ARI_HOOKS")
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_USER_MESSAGE")
	}()

	outputFlag := "json"
	verboseFlag := false

	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runRouteWithPrinter(ctx, printer)
	}

	// Report timing
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Slash command routing took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkRouteHook_RegularMessage benchmarks regular message processing (<5ms target).
func BenchmarkRouteHook_RegularMessage(b *testing.B) {
	os.Setenv("USE_ARI_HOOKS", "1")
	os.Setenv("CLAUDE_HOOK_EVENT", "UserPromptSubmit")
	os.Setenv("CLAUDE_USER_MESSAGE", "Help me understand this code and fix the bug in the authentication module")
	defer func() {
		os.Unsetenv("USE_ARI_HOOKS")
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_USER_MESSAGE")
	}()

	outputFlag := "json"
	verboseFlag := false

	ctx := &cmdContext{
		output:  &outputFlag,
		verbose: &verboseFlag,
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		runRouteWithPrinter(ctx, printer)
	}

	// Report timing
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Regular message routing took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkParseSlashCommand benchmarks the command parsing function directly.
func BenchmarkParseSlashCommand(b *testing.B) {
	messages := []string{
		"/start Add dark mode toggle",
		"/commit Fix authentication bug",
		"Help me with this code",
		"/consult Which team should handle this?",
		"   /task   Implement user registration",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range messages {
			parseSlashCommand(msg)
		}
	}
}
