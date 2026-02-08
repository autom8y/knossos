// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

// ToolInput represents the input from Claude Code PreToolUse hook.
type ToolInput struct {
	ToolName string `json:"tool_name"`
	FilePath string `json:"file_path"`
}


// Protected file patterns for context files.
var protectedPatterns = []string{
	"SESSION_CONTEXT.md",
	"SPRINT_CONTEXT.md",
}

// newWriteguardCmd creates the writeguard hook subcommand.
func newWriteguardCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "writeguard",
		Short: "Block direct writes to context files",
		Long: `Enforces Moirai for context file mutations.

This hook is triggered on PreToolUse events for Write/Edit tools. It:
- Checks if the target file is a protected context file (*_CONTEXT.md)
- Returns {"hookSpecificOutput": {"permissionDecision": "deny", "permissionDecisionReason": "..."}} to prevent the write
- Returns {"hookSpecificOutput": {"permissionDecision": "allow"}} for all other files
- Allows writes when Moirai holds a valid session lock

Input (env vars):
  CLAUDE_TOOL_INPUT: {"file_path": ".claude/sessions/.../SESSION_CONTEXT.md"}
  CLAUDE_PROJECT_DIR: project root directory

Output (stdout JSON):
  {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "Use Moirai for SESSION_CONTEXT mutations"}}

Performance: <5ms for passthrough path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runWriteguard(ctx)
			})
		},
	}

	return cmd
}

func runWriteguard(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runWriteguardCore(ctx, printer, "")
}

// runWriteguardCore contains the actual logic with injected printer for testing.
// stdinInput is used by tests to simulate stdin input.
func runWriteguardCore(ctx *cmdContext, printer *output.Printer, stdinInput string) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PreToolUse event
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreToolUse {
		return outputAllow(printer)
	}

	// Check if this is a Write or Edit tool
	toolName := hookEnv.ToolName
	if toolName != "Write" && toolName != "Edit" {
		return outputAllow(printer)
	}

	// Parse file path from tool input
	filePath := parseFilePath(printer, hookEnv.ToolInput)
	if filePath == "" && stdinInput != "" {
		// Try stdin input for testing
		filePath = parseFilePath(printer, stdinInput)
	}

	if filePath == "" {
		return outputAllow(printer)
	}

	// Check if file is protected
	if isProtectedFile(filePath) {
		// Check if Moirai lock is held
		if isMoiraiLockHeld(hookEnv.GetProjectDir()) {
			return outputAllow(printer)
		}
		return outputBlock(printer, filePath)
	}

	return outputAllow(printer)
}

// parseFilePath extracts file_path from JSON tool input.
func parseFilePath(printer *output.Printer, toolInput string) string {
	if toolInput == "" {
		return ""
	}

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		printer.VerboseLog("warn", "failed to parse tool input JSON",
			map[string]interface{}{"error": err.Error(), "input": toolInput})
		return ""
	}

	if fp, ok := input["file_path"].(string); ok {
		return fp
	}
	return ""
}

// isProtectedFile checks if the file path matches a protected pattern.
func isProtectedFile(filePath string) bool {
	for _, pattern := range protectedPatterns {
		if strings.HasSuffix(filePath, pattern) {
			return true
		}
	}
	return false
}

// isMoiraiLockHeld checks if a valid Moirai lock exists for the current session.
// Returns true only if:
// - Current session can be resolved
// - Lock file exists at .claude/sessions/{session-id}/.moirai-lock
// - Lock agent field is "moirai"
// - Lock is not stale (acquired_at + stale_after_seconds > now)
// Returns false on any error (fail closed).
func isMoiraiLockHeld(projectDir string) bool {
	if projectDir == "" {
		return false
	}

	// Read current session ID
	currentSessionPath := strings.TrimSpace(projectDir) + "/.claude/sessions/.current-session"
	sessionIDBytes, err := os.ReadFile(currentSessionPath)
	if err != nil {
		return false
	}
	sessionID := strings.TrimSpace(string(sessionIDBytes))
	if sessionID == "" {
		return false
	}

	// Check for lock file
	lockPath := strings.TrimSpace(projectDir) + "/.claude/sessions/" + sessionID + "/.moirai-lock"
	lockData, err := os.ReadFile(lockPath)
	if err != nil {
		return false
	}

	// Parse lock JSON
	var lock struct {
		Agent             string `json:"agent"`
		AcquiredAt        string `json:"acquired_at"`
		StaleAfterSeconds int    `json:"stale_after_seconds"`
	}
	if err := json.Unmarshal(lockData, &lock); err != nil {
		return false
	}

	// Verify agent is moirai
	if lock.Agent != "moirai" {
		return false
	}

	// Check if stale
	acquiredAt, err := time.Parse(time.RFC3339, lock.AcquiredAt)
	if err != nil {
		return false
	}
	staleThreshold := acquiredAt.Add(time.Duration(lock.StaleAfterSeconds) * time.Second)
	if time.Now().UTC().After(staleThreshold) {
		return false // stale lock
	}

	return true
}

// outputAllow outputs an allow decision in CC's hookSpecificOutput format.
func outputAllow(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
		},
	}
	return printer.Print(result)
}

// outputBlock outputs a block decision in CC's hookSpecificOutput format.
func outputBlock(printer *output.Printer, filePath string) error {
	// Determine which context file type
	var contextType string
	if strings.HasSuffix(filePath, "SESSION_CONTEXT.md") {
		contextType = "SESSION_CONTEXT"
	} else if strings.HasSuffix(filePath, "SPRINT_CONTEXT.md") {
		contextType = "SPRINT_CONTEXT"
	} else {
		contextType = "context file"
	}

	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "Use Moirai for " + contextType + " mutations",
		},
	}
	return printer.Print(result)
}
