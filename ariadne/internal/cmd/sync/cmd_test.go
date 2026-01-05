package sync

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestNewSyncCmd verifies the sync command is properly constructed.
func TestNewSyncCmd(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	cmd := NewSyncCmd(&outputFormat, &verbose, &projectDir)

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "sync" {
		t.Errorf("Use = %q, want %q", cmd.Use, "sync")
	}

	// Verify subcommands exist
	subcommands := cmd.Commands()
	expectedSubcommands := []string{"status", "pull", "push", "diff", "resolve", "history", "reset"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, sub := range subcommands {
			if sub.Use == expected || sub.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing subcommand: %q", expected)
		}
	}
}

// TestCmdContext_GetPrinter verifies printer creation.
func TestCmdContext_GetPrinter(t *testing.T) {
	outputFormat := "text"
	verbose := true
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	printer := ctx.getPrinter()
	if printer == nil {
		t.Fatal("Expected non-nil printer")
	}
}

// TestCmdContext_GetResolver verifies resolver creation.
func TestCmdContext_GetResolver(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	resolver := ctx.getResolver()
	if resolver == nil {
		t.Fatal("Expected non-nil resolver")
	}

	if resolver.ProjectRoot() != "/tmp/test" {
		t.Errorf("ProjectRoot() = %q, want %q", resolver.ProjectRoot(), "/tmp/test")
	}
}

// TestCmdContext_NilValues verifies context handles nil values gracefully.
func TestCmdContext_NilValues(t *testing.T) {
	ctx := &cmdContext{
		output:     nil,
		verbose:    nil,
		projectDir: nil,
	}

	// Should not panic
	printer := ctx.getPrinter()
	if printer == nil {
		t.Fatal("Expected non-nil printer even with nil values")
	}

	resolver := ctx.getResolver()
	if resolver == nil {
		t.Fatal("Expected non-nil resolver even with nil values")
	}
}

// TestShortenHash_Internal verifies the package-internal shortenHash function.
func TestShortenHash_Internal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "long hash",
			input:    "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			expected: "b94d27b9",
		},
		{
			name:     "short hash",
			input:    "abc",
			expected: "abc",
		},
		{
			name:     "exactly 8 chars",
			input:    "12345678",
			expected: "12345678",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenHash(tt.input)
			if result != tt.expected {
				t.Errorf("shortenHash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestStatusCmd_Construction verifies status command construction.
func TestStatusCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newStatusCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}
}

// TestPullCmd_Construction verifies pull command construction.
func TestPullCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newPullCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "pull [remote]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "pull [remote]")
	}

	// Verify flags exist
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag")
	}

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected --dry-run flag")
	}
}

// TestPushCmd_Construction verifies push command construction.
func TestPushCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newPushCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "push" {
		t.Errorf("Use = %q, want %q", cmd.Use, "push")
	}

	// Verify flags exist
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag")
	}

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected --dry-run flag")
	}
}

// TestDiffCmd_Construction verifies diff command construction.
func TestDiffCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newDiffCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "diff [path]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "diff [path]")
	}
}

// TestResolveCmd_Construction verifies resolve command construction.
func TestResolveCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newResolveCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "resolve [path]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "resolve [path]")
	}

	// Verify strategy flag
	strategyFlag := cmd.Flags().Lookup("strategy")
	if strategyFlag == nil {
		t.Error("Expected --strategy flag")
	}
	if strategyFlag.DefValue != "ours" {
		t.Errorf("strategy default = %q, want %q", strategyFlag.DefValue, "ours")
	}
}

// TestHistoryCmd_Construction verifies history command construction.
func TestHistoryCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newHistoryCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "history" {
		t.Errorf("Use = %q, want %q", cmd.Use, "history")
	}

	// Verify limit flag
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("Expected --limit flag")
	}
	if limitFlag.DefValue != "20" {
		t.Errorf("limit default = %q, want %q", limitFlag.DefValue, "20")
	}
}

// TestResetCmd_Construction verifies reset command construction.
func TestResetCmd_Construction(t *testing.T) {
	outputFormat := "text"
	verbose := false
	projectDir := "/tmp/test"

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	cmd := newResetCmd(ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "reset" {
		t.Errorf("Use = %q, want %q", cmd.Use, "reset")
	}

	// Verify hard flag
	hardFlag := cmd.Flags().Lookup("hard")
	if hardFlag == nil {
		t.Error("Expected --hard flag")
	}
}

// TestRunStatus_NotInitialized verifies runStatus with no sync state.
func TestRunStatus_NotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Should not error - just shows not initialized
	err := runStatus(ctx)
	if err != nil {
		t.Errorf("runStatus() error = %v", err)
	}
}

// TestRunDiff_NotInitialized verifies runDiff with no sync state.
func TestRunDiff_NotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Should error - not initialized
	err := runDiff(ctx, "", diffOptions{})
	if err == nil {
		t.Error("Expected error when diff is run without initialization")
	}
}

// TestRunResolve_InvalidStrategy verifies runResolve with invalid strategy.
func TestRunResolve_InvalidStrategy(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Invalid strategy should error
	err := runResolve(ctx, "", resolveOptions{strategy: "invalid"})
	if err == nil {
		t.Error("Expected error for invalid strategy")
	}
}

// TestRunHistory_Empty verifies runHistory with no history.
func TestRunHistory_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Should not error - just shows empty history
	err := runHistory(ctx, historyOptions{limit: 10})
	if err != nil {
		t.Errorf("runHistory() error = %v", err)
	}
}

// TestRunReset_NotInitialized verifies runReset when not initialized.
func TestRunReset_NotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Should error - not initialized
	err := runReset(ctx, resetOptions{force: true})
	if err == nil {
		t.Error("Expected error when reset is run without initialization")
	}
}

// TestRunPull_NoRemote verifies runPull without remote when not initialized.
func TestRunPull_NoRemote(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Should error - no remote specified and not initialized
	err := runPull(ctx, "", pullOptions{})
	if err == nil {
		t.Error("Expected error when pull is run without remote and not initialized")
	}
}

// TestRunPush_NotInitialized verifies runPush when not initialized.
func TestRunPush_NotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	outputFormat := "json"
	verbose := false

	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &tmpDir,
	}

	// Should error - not initialized
	err := runPush(ctx, pushOptions{})
	if err == nil {
		t.Error("Expected error when push is run without initialization")
	}
}

// TestOutputCapture tests that output is captured correctly.
func TestOutputCapture(t *testing.T) {
	// This is a sanity test to verify we can capture output
	var buf bytes.Buffer
	buf.WriteString("test output")

	if buf.String() != "test output" {
		t.Errorf("Buffer content = %q, want %q", buf.String(), "test output")
	}
}
