// Package hook implements the ari hook commands.
package hook

import (
	"regexp"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/registry"
)

// Regex patterns for AI attribution marker detection.
var (
	// Case-insensitive Co-Authored-By in any form.
	coAuthoredByPattern = regexp.MustCompile(`(?i)co-authored-by\s*:`)

	// "Generated with" Claude/AI/Anthropic footers.
	generatedWithPattern = regexp.MustCompile(`(?i)generated\s+with\s+.*(claude|ai|anthropic)`)

	// anthropic.com email in trailers.
	anthropicEmailPattern = regexp.MustCompile(`(?i)noreply@anthropic\.com`)
)

// attributionDenyReason is the denial message for commits containing AI attribution markers.
var attributionDenyReason = "Commit contains AI attribution marker. " +
	"Per platform conventions, commits use user-only attribution. " +
	registry.Recovery(registry.SkillAttributionGuard)

// newAttributionGuardCmd creates the attribution-guard hook subcommand.
func newAttributionGuardCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attribution-guard",
		Short: "Block git commits containing AI attribution markers",
		Long: `Blocks git commits that contain AI attribution markers.

This hook is triggered on PreToolUse events for Bash tools. It:
- Scans git commit commands (including heredoc bodies) for AI markers
- Detects Co-Authored-By, "Generated with" footers, anthropic.com emails
- Returns deny to block commits with attribution markers
- Does NOT skip heredocs (unlike git-conventions) — heredocs are the primary vector
- Returns {"hookSpecificOutput": {"permissionDecision": "deny"|"allow"}}

Input (stdin JSON):
  {"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"...\""}}

Output (stdout JSON):
  {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow"}}

Performance: <1ms for non-git-commit commands (fast-path).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runAttributionGuard(cmd, ctx)
			})
		},
	}

	return cmd
}

func runAttributionGuard(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runAttributionGuardCore(cmd, ctx, printer)
}

// runAttributionGuardCore contains the actual logic with injected printer for testing.
func runAttributionGuardCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Fast-path: only handle PreToolUse events
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreTool {
		return outputAttributionAllow(printer)
	}

	// Fast-path: only handle Bash tool
	if hookEnv.ToolName != "Bash" {
		return outputAttributionAllow(printer)
	}

	// Parse command from tool input
	command := parseCommand(printer, hookEnv.ToolInput)
	if command == "" {
		return outputAttributionAllow(printer)
	}

	// Fast-path: not a git commit command
	if !gitCommitPattern.MatchString(command) {
		return outputAttributionAllow(printer)
	}

	// Scan full command string (including heredoc body) for AI attribution markers
	if coAuthoredByPattern.MatchString(command) ||
		generatedWithPattern.MatchString(command) ||
		anthropicEmailPattern.MatchString(command) {
		return outputAttributionDeny(printer, attributionDenyReason)
	}

	return outputAttributionAllow(printer)
}

// outputAttributionAllow outputs an allow decision.
func outputAttributionAllow(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
		},
	}
	return printer.Print(result)
}

// outputAttributionDeny outputs a deny decision with a reason.
func outputAttributionDeny(printer *output.Printer, reason string) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
	return printer.Print(result)
}
