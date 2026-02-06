package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/test/hooks/testutil"
)

func TestBudgetOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   BudgetOutput
		expected string
	}{
		{
			name:     "with message",
			output:   BudgetOutput{Message: "budget tracking disabled"},
			expected: "budget tracking disabled",
		},
		{
			name:     "count only",
			output:   BudgetOutput{Count: 42},
			expected: "count=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.output.Text(); got != tt.expected {
				t.Errorf("Text() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func newTestContext(t *testing.T) *cmdContext {
	t.Helper()
	outputFlag := "json"
	verboseFlag := false
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:  &outputFlag,
				Verbose: &verboseFlag,
			},
		},
	}
}

func TestRunBudget_Disabled(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: true,
	})
	t.Setenv(envBudgetDisable, "1")
	t.Setenv(envSessionKey, "disabled-test")

	// Clean up any state file from previous runs
	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-disabled-test")
	os.Remove(expectedPath)

	ctx := newTestContext(t)
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() error = %v", err)
	}

	// When disabled, no counter file should be created
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Error("Counter file should not be created when budget is disabled")
	}
}

func TestRunBudget_HooksDisabled(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: false,
	})
	t.Setenv(envSessionKey, "hooks-disabled-test")

	// Clean up any state file from previous runs
	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-hooks-disabled-test")
	os.Remove(expectedPath)

	ctx := newTestContext(t)
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() error = %v", err)
	}

	// When hooks disabled, no counter file should be created
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Error("Counter file should not be created when hooks are disabled")
	}
}

func TestIncrementCounter(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test-counter")

	// First increment: 0 → 1
	count, err := incrementCounter(stateFile)
	if err != nil {
		t.Fatalf("incrementCounter() error = %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Second increment: 1 → 2
	count, err = incrementCounter(stateFile)
	if err != nil {
		t.Fatalf("incrementCounter() error = %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// Verify file contents
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if string(data) != "2" {
		t.Errorf("file contents = %q, want %q", string(data), "2")
	}
}

func TestIncrementCounter_MultipleIncrements(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "multi-counter")

	for i := 1; i <= 10; i++ {
		count, err := incrementCounter(stateFile)
		if err != nil {
			t.Fatalf("incrementCounter() iteration %d error = %v", i, err)
		}
		if count != i {
			t.Errorf("iteration %d: count = %d, want %d", i, count, i)
		}
	}
}

func TestResolveStateFile(t *testing.T) {
	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:  &outputFlag,
				Verbose: &verboseFlag,
			},
		},
	}

	t.Run("uses ARIADNE_SESSION_KEY", func(t *testing.T) {
		t.Setenv(envSessionKey, "my-test-key")
		t.Setenv("CLAUDE_SESSION_ID", "should-not-use")

		result := resolveStateFile(ctx)
		expected := filepath.Join(os.TempDir(), "ariadne-msg-count-my-test-key")
		if result != expected {
			t.Errorf("resolveStateFile() = %q, want %q", result, expected)
		}
	})

	t.Run("falls back to CLAUDE_SESSION_ID", func(t *testing.T) {
		t.Setenv(envSessionKey, "")
		os.Unsetenv(envSessionKey)
		t.Setenv("CLAUDE_HOOK_EVENT", "PostToolUse")
		t.Setenv("CLAUDE_SESSION_ID", "sess-abc-123")

		result := resolveStateFile(ctx)
		expected := filepath.Join(os.TempDir(), "ariadne-msg-count-sess-abc-123")
		if result != expected {
			t.Errorf("resolveStateFile() = %q, want %q", result, expected)
		}
	})

	t.Run("sanitizes special characters", func(t *testing.T) {
		t.Setenv(envSessionKey, "test/key with spaces!@#")

		result := resolveStateFile(ctx)
		expected := filepath.Join(os.TempDir(), "ariadne-msg-count-test_key_with_spaces___")
		if result != expected {
			t.Errorf("resolveStateFile() = %q, want %q", result, expected)
		}
	})
}

func TestRunBudget_WarnThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "ariadne-msg-count-warn-test")

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: true,
	})
	t.Setenv(envSessionKey, "warn-test")
	t.Setenv(envMsgWarn, "3")

	// Override TempDir to use our test dir
	origTempDir := os.TempDir
	_ = origTempDir

	// Pre-seed counter to 2 (next increment → 3 = warn threshold)
	if err := os.WriteFile(stateFile, []byte("2"), 0644); err != nil {
		t.Fatalf("Failed to seed counter: %v", err)
	}

	// We need the state file to be at the resolved path
	// Since resolveStateFile uses os.TempDir(), we set ARIADNE_SESSION_KEY
	// and pre-seed the file at that path
	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-warn-test")
	if err := os.WriteFile(expectedPath, []byte("2"), 0644); err != nil {
		t.Fatalf("Failed to seed counter at temp: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(expectedPath)
		os.Remove(expectedPath + ".warned")
		os.Remove(expectedPath + ".park-warned")
	})

	ctx := newTestContext(t)
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() error = %v", err)
	}

	// Verify warn marker was created
	if _, err := os.Stat(expectedPath + ".warned"); os.IsNotExist(err) {
		t.Error("Expected .warned marker file to be created")
	}
}

func TestRunBudget_ParkThreshold(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: true,
	})
	t.Setenv(envSessionKey, "park-test")
	t.Setenv(envMsgWarn, "1000") // high warn so we only test park
	t.Setenv(envMsgPark, "5")

	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-park-test")
	if err := os.WriteFile(expectedPath, []byte("4"), 0644); err != nil {
		t.Fatalf("Failed to seed counter: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(expectedPath)
		os.Remove(expectedPath + ".warned")
		os.Remove(expectedPath + ".park-warned")
	})

	ctx := newTestContext(t)
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() error = %v", err)
	}

	// Verify park marker was created
	if _, err := os.Stat(expectedPath + ".park-warned"); os.IsNotExist(err) {
		t.Error("Expected .park-warned marker file to be created")
	}
}

func TestRunBudget_OneShot(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: true,
	})
	t.Setenv(envSessionKey, "oneshot-test")
	t.Setenv(envMsgWarn, "1")

	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-oneshot-test")
	// Start clean
	os.Remove(expectedPath)
	os.Remove(expectedPath + ".warned")
	t.Cleanup(func() {
		os.Remove(expectedPath)
		os.Remove(expectedPath + ".warned")
		os.Remove(expectedPath + ".park-warned")
	})

	ctx := newTestContext(t)

	// First call: count=1 >= warn=1 → should create marker
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() first call error = %v", err)
	}
	if _, err := os.Stat(expectedPath + ".warned"); os.IsNotExist(err) {
		t.Fatal("Expected .warned marker after first call")
	}

	// Second call: count=2 >= warn=1 → marker exists, should NOT re-warn
	// (We can't easily check the severity from here without capturing output,
	// but we verify the marker file is unchanged)
	markerInfo, _ := os.Stat(expectedPath + ".warned")
	markerModTime := markerInfo.ModTime()

	err = runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() second call error = %v", err)
	}

	markerInfo2, _ := os.Stat(expectedPath + ".warned")
	if !markerInfo2.ModTime().Equal(markerModTime) {
		t.Error("Warn marker was rewritten — one-shot behavior broken")
	}
}

func TestRunBudget_CustomThresholds(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: true,
	})
	t.Setenv(envSessionKey, "custom-thresh")
	t.Setenv(envMsgWarn, "100")
	t.Setenv(envMsgPark, "200")

	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-custom-thresh")
	os.Remove(expectedPath)
	t.Cleanup(func() {
		os.Remove(expectedPath)
		os.Remove(expectedPath + ".warned")
		os.Remove(expectedPath + ".park-warned")
	})

	ctx := newTestContext(t)

	// count=1, below both thresholds
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() error = %v", err)
	}

	// Verify no markers
	if _, err := os.Stat(expectedPath + ".warned"); !os.IsNotExist(err) {
		t.Error("Warn marker should not exist at count=1")
	}
	if _, err := os.Stat(expectedPath + ".park-warned"); !os.IsNotExist(err) {
		t.Error("Park marker should not exist at count=1")
	}
}

func TestRunBudget_InvalidThresholds(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "PostToolUse",
		ToolName:    "Bash",
		UseAriHooks: true,
	})
	t.Setenv(envSessionKey, "invalid-thresh")
	t.Setenv(envMsgWarn, "not-a-number")
	t.Setenv(envMsgPark, "-5")

	expectedPath := filepath.Join(os.TempDir(), "ariadne-msg-count-invalid-thresh")
	os.Remove(expectedPath)
	t.Cleanup(func() {
		os.Remove(expectedPath)
		os.Remove(expectedPath + ".warned")
		os.Remove(expectedPath + ".park-warned")
	})

	ctx := newTestContext(t)

	// Should use defaultWarn (250) for invalid warn, ignore invalid park
	err := runBudget(ctx)
	if err != nil {
		t.Fatalf("runBudget() error = %v", err)
	}

	// count=1, well below default 250 — no markers
	if _, err := os.Stat(expectedPath + ".warned"); !os.IsNotExist(err) {
		t.Error("Warn marker should not exist with default threshold")
	}
}
