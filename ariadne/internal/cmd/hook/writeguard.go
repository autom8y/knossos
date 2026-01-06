// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/hook"
	"github.com/autom8y/ariadne/internal/output"
)

// WriteGuardDecision represents the decision for a write operation.
type WriteGuardDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

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
		Long: `Enforces state-mate for context file mutations.

This hook is triggered on PreToolUse events for Write/Edit tools. It:
- Checks if the target file is a protected context file (*_CONTEXT.md)
- Returns {"decision": "block", "reason": "..."} to prevent the write
- Returns {"decision": "allow"} for all other files
- Respects MOIRAI_BYPASS env var for override

Input (stdin JSON):
  {"tool_name": "Write", "file_path": ".claude/sessions/.../SESSION_CONTEXT.md"}

Output (stdout JSON):
  {"decision": "block", "reason": "Use state-mate for SESSION_CONTEXT mutations"}

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

	// Early exit if hooks disabled - allow all writes
	if ctx.shouldEarlyExit() {
		return outputAllow(printer)
	}

	// Check bypass env var
	if os.Getenv(BypassEnvVar) == "1" {
		printer.VerboseLog("debug", "writeguard bypassed via env var", nil)
		return outputAllow(printer)
	}

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PreToolUse event
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreToolUse {
		printer.VerboseLog("debug", "skipping writeguard for non-PreToolUse event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputAllow(printer)
	}

	// Check if this is a Write or Edit tool
	toolName := hookEnv.ToolName
	if toolName != "Write" && toolName != "Edit" {
		// Not a write operation, allow
		return outputAllow(printer)
	}

	// Parse file path from tool input (try env var first, then stdin)
	filePath := parseFilePath(printer, hookEnv.ToolInput)
	if filePath == "" {
		// Try reading from stdin as fallback
		filePath = parseFilePathFromStdin(printer)
	}

	if filePath == "" {
		printer.VerboseLog("debug", "no file path in tool input", nil)
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

// parseFilePathFromStdin reads JSON input from stdin with a timeout.
func parseFilePathFromStdin(printer *output.Printer) string {
	// Create a context with timeout for stdin read (50ms should be plenty for piped input)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Channel to receive read result
	type readResult struct {
		data []byte
		err  error
	}
	resultCh := make(chan readResult, 1)

	// Read stdin in a goroutine
	go func() {
		// Read stdin with a limit to prevent excessive memory usage
		data, err := io.ReadAll(io.LimitReader(os.Stdin, 8192))
		resultCh <- readResult{data: data, err: err}
	}()

	// Wait for either read completion or timeout
	var data []byte
	var err error
	select {
	case result := <-resultCh:
		data = result.data
		err = result.err
	case <-ctx.Done():
		// Timeout - stdin is likely not piped or hung
		printer.VerboseLog("debug", "stdin read timed out (no piped input)", nil)
		return ""
	}

	if err != nil || len(data) == 0 {
		return ""
	}

	var input ToolInput
	if err := json.Unmarshal(data, &input); err != nil {
		// Try the map form
		var mapInput map[string]interface{}
		if err2 := json.Unmarshal(data, &mapInput); err2 != nil {
			printer.VerboseLog("warn", "failed to parse stdin JSON",
				map[string]interface{}{"error": err.Error(), "error2": err2.Error(), "data": string(data)})
			return ""
		}
		if fp, ok := mapInput["file_path"].(string); ok {
			return fp
		}
		return ""
	}

	return input.FilePath
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

// outputAllow outputs an allow decision.
func outputAllow(printer interface{ Print(interface{}) error }) error {
	result := WriteGuardDecision{
		Decision: "allow",
	}
	return printer.Print(result)
}

// outputBlock outputs a block decision with the reason.
func outputBlock(printer interface{ Print(interface{}) error }, filePath string) error {
	// Determine which context file type
	var contextType string
	if strings.HasSuffix(filePath, "SESSION_CONTEXT.md") {
		contextType = "SESSION_CONTEXT"
	} else if strings.HasSuffix(filePath, "SPRINT_CONTEXT.md") {
		contextType = "SPRINT_CONTEXT"
	} else {
		contextType = "context file"
	}

	result := WriteGuardDecision{
		Decision: "block",
		Reason:   "Use state-mate for " + contextType + " mutations",
	}
	return printer.Print(result)
}
