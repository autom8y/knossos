// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

// newAgentGuardCmd creates the agent-guard hook subcommand.
func newAgentGuardCmd(ctx *cmdContext) *cobra.Command {
	var agentName string
	var allowPaths []string

	cmd := &cobra.Command{
		Use:   "agent-guard",
		Short: "Enforce path-based write boundaries for agents",
		Long: `Enforces per-agent write path restrictions on PreToolUse Write events.

This hook is triggered on PreToolUse events for the Write tool. It:
- Checks if the target file_path matches any --allow-path prefix
- Returns {"hookSpecificOutput": {"permissionDecision": "deny", ...}} if the path is not allowed
- Returns {"hookSpecificOutput": {"permissionDecision": "allow"}} if the path matches any prefix
- Allows all non-Write tool events to pass through

Input (stdin JSON):
  {"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"internal/agent/f.go"}}

Output (stdout JSON):
  {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "ecosystem-analyst is not permitted to Write files: internal/agent/f.go is outside allowed paths"}}

Flags:
  --agent       Agent name for deny reason messages (default: "this agent")
  --allow-path  Path prefix that exempts the write (repeatable)

Performance: <5ms for all paths.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runAgentGuard(cmd, ctx, agentName, allowPaths)
			})
		},
	}

	cmd.Flags().StringVar(&agentName, "agent", "this agent",
		"Agent name used in deny reason messages")
	cmd.Flags().StringArrayVar(&allowPaths, "allow-path", nil,
		"Path prefix that allows the write (repeatable; trailing slash recommended)")

	return cmd
}

// runAgentGuard is the entry point called by RunE via withTimeout.
func runAgentGuard(cmd *cobra.Command, ctx *cmdContext, agentName string, allowPaths []string) error {
	printer := ctx.getPrinter()
	return runAgentGuardCore(cmd, ctx, printer, agentName, allowPaths)
}

// runAgentGuardCore contains the testable core logic for agent-guard.
// It enforces path-based write boundaries: only writes to paths matching an
// --allow-path prefix are permitted. All other writes are denied with an
// agent-specific reason message.
func runAgentGuardCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer, agentName string, allowPaths []string) error {
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Pass through non-PreToolUse events without inspection.
	// Empty event (direct CLI invocation or test) is treated as PreToolUse.
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreTool {
		return outputAllow(printer)
	}

	// Distinguish JSON parse error (ALLOW) from missing file_path (DENY).
	// parseFilePath() returns "" for both cases; we need to distinguish them.
	filePath, parseError := parseFilePathStrict(printer, hookEnv.ToolInput)
	if parseError {
		// Bad JSON or empty stdin: graceful degradation per hook contract.
		// "Errors default to allow -- never block on hook failure."
		return outputAllow(printer)
	}

	if filePath == "" {
		// Well-formed JSON but no file_path field: model sent a malformed Write.
		// Fail closed to prevent bypass-by-omission.
		return outputAgentDeny(printer, agentName, hookEnv.ToolName, "file path is missing from tool input")
	}

	// Check whether the file path matches any allowed prefix.
	if isAllowedPath(filePath, allowPaths) {
		return outputAllow(printer)
	}

	// No allowed prefix matched (includes the empty allowPaths case, where the
	// loop body never executes and we fall through to deny).
	return outputAgentDeny(printer, agentName, hookEnv.ToolName, filePath+" is outside allowed paths")
}

// parseFilePathStrict extracts file_path from JSON tool input, distinguishing
// a JSON parse error from a missing field. Unlike parseFilePath() in writeguard.go,
// it returns a (path, parseError) pair:
//   - parseError=true means the JSON was unparseable or toolInput was empty and
//     toolName was also empty (corrupt/missing stdin) -- caller should ALLOW.
//   - parseError=false, path="" means the JSON parsed but file_path was absent
//     or empty -- caller should DENY (well-formed Write without a path).
//   - parseError=false, path!="" means normal case -- proceed to prefix check.
func parseFilePathStrict(printer *output.Printer, toolInput string) (filePath string, parseError bool) {
	if toolInput == "" {
		// Empty toolInput with no toolName means no stdin payload at all.
		// Treat as graceful degradation (parse error = ALLOW).
		return "", true
	}

	var input map[string]any
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		printer.VerboseLog("warn", "agent-guard: failed to parse tool input JSON",
			map[string]any{"error": err.Error(), "input": toolInput})
		return "", true
	}

	// JSON parsed successfully. Extract file_path (may be absent).
	if fp, ok := input["file_path"].(string); ok {
		return fp, false
	}
	// file_path field absent or not a string: well-formed JSON, missing path.
	return "", false
}

// isAllowedPath reports whether filePath matches any of the given path prefixes.
// Two match conditions per prefix:
//  1. strings.HasPrefix: handles relative paths (e.g. ".sos/wip/gap.md" vs ".sos/wip/")
//  2. strings.Contains with "/"+prefix: handles absolute paths
//     (e.g. "/Users/tom/project/.sos/wip/gap.md" contains "/.sos/wip/")
//
// No path normalization is applied -- callers control exact prefix strings.
// A trailing slash in the prefix prevents sibling-directory false positives
// (e.g. ".sos/wip/" does not match ".sos/wip-private/file.md").
func isAllowedPath(filePath string, allowPaths []string) bool {
	for _, prefix := range allowPaths {
		if strings.HasPrefix(filePath, prefix) || strings.Contains(filePath, "/"+prefix) {
			return true
		}
	}
	return false
}

// outputAgentDeny outputs a deny decision with an agent-specific reason message.
// The reason format is: "<agentName> is not permitted to <toolName> files: <reason>"
// AdditionalContext directs the agent to check its hooks configuration.
func outputAgentDeny(printer *output.Printer, agentName, toolName, reason string) error {
	if toolName == "" {
		toolName = "Write"
	}
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: agentName + " is not permitted to " + toolName + " files: " + reason,
			AdditionalContext:        "This agent's write access is restricted to specific paths. Check the agent's hooks configuration for allowed paths.",
		},
	}
	return printer.Print(result)
}
