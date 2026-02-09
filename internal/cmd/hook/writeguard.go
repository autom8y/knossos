// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
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

Input (stdin JSON):
  {"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":".claude/sessions/.../SESSION_CONTEXT.md"}}

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
	return runWriteguardCore(ctx, printer)
}

// runWriteguardCore contains the actual logic with injected printer for testing.
func runWriteguardCore(ctx *cmdContext, printer *output.Printer) error {
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
	if filePath == "" {
		return outputAllow(printer)
	}

	// Check if file is protected
	if isProtectedFile(filePath) {
		// Resolve session via priority chain, then check Moirai lock
		resolver, sessionID, _ := ctx.resolveSession(hookEnv)
		if sessionID != "" && isMoiraiLockHeld(resolver, sessionID) {
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

// isMoiraiLockHeld checks if a valid Moirai lock exists for the given session.
// Returns true only if:
// - Lock file exists at .claude/sessions/{session-id}/.moirai-lock
// - Lock agent field is "moirai"
// - Lock is not stale (acquired_at + stale_after_seconds > now)
// Returns false on any error (fail closed).
func isMoiraiLockHeld(resolver *paths.Resolver, sessionID string) bool {
	lockPath := resolver.SessionDir(sessionID) + "/.moirai-lock"
	moiraiLock, err := lock.ReadMoiraiLock(lockPath)
	if err != nil {
		return false
	}

	if moiraiLock.Agent != "moirai" {
		return false
	}

	if lock.IsMoiraiLockStale(moiraiLock) {
		return false
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
// Includes additionalContext with the exact Moirai Task invocation pattern
// so Claude knows how to delegate instead of retrying the blocked write.
func outputBlock(printer *output.Printer, filePath string) error {
	var contextType string
	var moiraiOp string
	if strings.HasSuffix(filePath, "SESSION_CONTEXT.md") {
		contextType = "SESSION_CONTEXT"
		moiraiOp = `Task(moirai, "update_session ...")`
	} else if strings.HasSuffix(filePath, "SPRINT_CONTEXT.md") {
		contextType = "SPRINT_CONTEXT"
		moiraiOp = `Task(moirai, "create_sprint ...")`
	} else {
		contextType = "context file"
		moiraiOp = `Task(moirai, "<operation>")`
	}

	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "Use Moirai for " + contextType + " mutations",
			AdditionalContext:        "To mutate " + contextType + ", delegate via: " + moiraiOp + ". Moirai is the session lifecycle agent that handles all *_CONTEXT.md mutations with proper locking and validation.",
		},
	}
	return printer.Print(result)
}
