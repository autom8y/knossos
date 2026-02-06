// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

// ValidateDecision represents the decision for a bash command validation.
type ValidateDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// ValidateBypassEnvVar is the environment variable to bypass validate hook.
const ValidateBypassEnvVar = "ARI_VALIDATE_BYPASS"

// BashToolInput represents the input from Claude Code Bash tool.
type BashToolInput struct {
	Command string `json:"command"`
}

// Protected paths that should not be deleted with rm -rf.
// Ordered from longest to shortest to match more specific paths first.
var protectedPaths = []string{
	".github/",
	".github",
	".claude/",
	".claude",
	".git/",
	".git",
	"node_modules/",
	"node_modules",
}

// Regex patterns for dangerous commands.
var (
	// rm -rf on protected paths
	rmRfPattern = regexp.MustCompile(`\brm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+`)

	// git push --force or -f to main/master
	forcePushMainPattern = regexp.MustCompile(`\bgit\s+push\s+[^|;]*?(--force|-f)\s+[^|;]*(main|master)\b`)
	forcePushPattern     = regexp.MustCompile(`\bgit\s+push\s+[^|;]*(--force|-f)`)

	// git commit with --no-verify
	noVerifyPattern = regexp.MustCompile(`\bgit\s+(commit|push)\s+[^|;]*--no-verify`)

	// git reset --hard
	resetHardPattern = regexp.MustCompile(`\bgit\s+reset\s+[^|;]*--hard`)

	// git clean -fd
	cleanFdPattern = regexp.MustCompile(`\bgit\s+clean\s+[^|;]*(-[a-zA-Z]*f[a-zA-Z]*d[a-zA-Z]*|-[a-zA-Z]*d[a-zA-Z]*f[a-zA-Z]*)`)
)

// newValidateCmd creates the validate hook subcommand.
func newValidateCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate bash commands against security rules",
		Long: `Validates bash commands for potentially dangerous operations.

This hook is triggered on PreToolUse events for Bash tools. It:
- Blocks rm -rf on protected paths (.claude/, .git/, etc.)
- Blocks force push to main/master branches
- Blocks --no-verify on commits
- Blocks destructive git commands (reset --hard, clean -fd)
- Returns {"decision": "block", "reason": "..."} to prevent execution
- Returns {"decision": "allow"} for safe commands
- Respects ARI_VALIDATE_BYPASS env var for override

Input (stdin JSON):
  {"command": "rm -rf .git", "description": "..."}

Output (stdout JSON):
  {"decision": "block", "reason": "Cannot rm -rf protected path: .git"}

Performance: <5ms for passthrough path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runValidate(ctx)
			})
		},
	}

	return cmd
}

func runValidate(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	// Check bypass env var
	if os.Getenv(ValidateBypassEnvVar) == "1" {
		printer.VerboseLog("debug", "validate bypassed via env var", nil)
		return outputValidateAllow(printer)
	}

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PreToolUse event
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreToolUse {
		printer.VerboseLog("debug", "skipping validate for non-PreToolUse event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputValidateAllow(printer)
	}

	// Check if this is a Bash tool
	if hookEnv.ToolName != "Bash" {
		// Not a bash command, allow
		return outputValidateAllow(printer)
	}

	// Parse command from tool input
	command := parseCommand(printer, hookEnv.ToolInput)
	if command == "" {
		printer.VerboseLog("debug", "no command in tool input", nil)
		return outputValidateAllow(printer)
	}

	// Validate the command
	if blocked, reason := validateCommand(command); blocked {
		return outputValidateBlock(printer, reason)
	}

	return outputValidateAllow(printer)
}

// parseCommand extracts command from JSON tool input.
func parseCommand(printer *output.Printer, toolInput string) string {
	if toolInput == "" {
		return ""
	}

	var input BashToolInput
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		// Try the map form
		var mapInput map[string]interface{}
		if err2 := json.Unmarshal([]byte(toolInput), &mapInput); err2 != nil {
			printer.VerboseLog("warn", "failed to parse tool input JSON",
				map[string]interface{}{"error": err.Error(), "error2": err2.Error(), "input": toolInput})
			return ""
		}
		if cmd, ok := mapInput["command"].(string); ok {
			return cmd
		}
		return ""
	}

	return input.Command
}

// validateCommand checks if a command is dangerous.
// Returns (blocked, reason) where blocked is true if the command should be blocked.
func validateCommand(command string) (bool, string) {
	// Normalize command for matching
	cmd := strings.ToLower(command)

	// Check for rm -rf on protected paths
	if rmRfPattern.MatchString(cmd) {
		for _, path := range protectedPaths {
			// Check if the protected path appears after rm -rf
			if strings.Contains(command, path) {
				return true, "Cannot rm -rf protected path: " + strings.TrimSuffix(path, "/")
			}
		}
	}

	// Check for force push to main/master
	if forcePushMainPattern.MatchString(cmd) {
		return true, "Force push to main/master is blocked. Use --force-with-lease or push to a feature branch."
	}

	// Check for --no-verify on commits or pushes
	if noVerifyPattern.MatchString(cmd) {
		return true, "Skipping hooks with --no-verify is blocked. Pre-commit hooks exist for a reason."
	}

	// Check for git reset --hard
	if resetHardPattern.MatchString(cmd) {
		return true, "git reset --hard is blocked. Use git stash or git checkout for safer alternatives."
	}

	// Check for git clean -fd
	if cleanFdPattern.MatchString(cmd) {
		return true, "git clean -fd is blocked on protected branches. Use git stash or manual cleanup."
	}

	return false, ""
}

// outputValidateAllow outputs an allow decision.
func outputValidateAllow(printer interface{ Print(interface{}) error }) error {
	result := ValidateDecision{
		Decision: "allow",
	}
	return printer.Print(result)
}

// outputValidateBlock outputs a block decision with the reason.
func outputValidateBlock(printer interface{ Print(interface{}) error }, reason string) error {
	result := ValidateDecision{
		Decision: "block",
		Reason:   reason,
	}
	return printer.Print(result)
}

// runValidateWithPrinter is a helper for testing with injected printer.
func runValidateWithPrinter(ctx *cmdContext, printer interface{ Print(interface{}) error }, stdinInput string) error {
	// Check bypass env var
	if os.Getenv(ValidateBypassEnvVar) == "1" {
		return outputValidateAllow(printer)
	}

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PreToolUse event
	if hookEnv.Event != "" && string(hookEnv.Event) != "PreToolUse" {
		return outputValidateAllow(printer)
	}

	// Check if this is a Bash tool
	if hookEnv.ToolName != "Bash" {
		return outputValidateAllow(printer)
	}

	// Parse command from tool input
	testPrinter := printer.(*output.Printer)
	command := parseCommand(testPrinter, hookEnv.ToolInput)
	if command == "" && stdinInput != "" {
		// Try stdin input for testing
		command = parseCommand(testPrinter, stdinInput)
	}

	if command == "" {
		return outputValidateAllow(printer)
	}

	// Validate the command
	if blocked, reason := validateCommand(command); blocked {
		return outputValidateBlock(printer, reason)
	}

	return outputValidateAllow(printer)
}
