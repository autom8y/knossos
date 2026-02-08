// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

// ToolInput represents the input from Claude Code PreToolUse hook.
type ToolInput struct {
	ToolName string `json:"tool_name"`
	FilePath string `json:"file_path"`
}

// BypassEnvVar is the environment variable to bypass writeguard.
const BypassEnvVar = "MOIRAI_BYPASS"

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
- Respects MOIRAI_BYPASS env var for override

Input (env vars):
  CLAUDE_TOOL_INPUT: {"file_path": ".claude/sessions/.../SESSION_CONTEXT.md"}

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
	// Check bypass env var
	if os.Getenv(BypassEnvVar) == "1" {
		return outputAllow(printer)
	}

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
