// Package hook implements the ari hook commands.
package hook

import (
	"regexp"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
)

// Regex patterns for git commit convention validation.
var (
	// Detects git commit commands.
	gitCommitPattern = regexp.MustCompile(`\bgit\s+commit\b`)

	// Extracts the commit message from -m flag (double or single quotes).
	// Captures message in group 1 (double quotes) or group 2 (single quotes).
	commitMessageFlag = regexp.MustCompile(`-m\s+(?:"([^"]+)"|'([^']+)')`)

	// Validates conventional commit format: type(scope): subject
	// Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build
	conventionalFormat = regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|perf|ci|build)(\([^)]+\))?: .+$`)

	// Detects heredoc-style commit messages that are too complex to validate.
	heredocPattern = regexp.MustCompile(`\$\(cat\s+<<`)

	// Detects --amend flag (modifying existing commit, skip validation).
	amendPattern = regexp.MustCompile(`--amend`)
)

// conventionDenyReason is the denial message pointing agents to the conventions skill.
const conventionDenyReason = "Commit message does not follow conventional format. " +
	"Load skill commit:behavior for full specification. " +
	"Expected: type(scope): subject where type is one of: feat|fix|docs|style|refactor|test|chore|perf|ci|build"

// newGitConventionsCmd creates the git-conventions hook subcommand.
func newGitConventionsCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-conventions",
		Short: "Validate git commit messages against conventional format",
		Long: `Validates git commit messages against conventional commit format.

This hook is triggered on PreToolUse events for Bash tools. It:
- Fast-exits for non-git-commit commands (<1ms)
- Validates commit messages match type(scope): subject format
- Allows interactive commits (no -m flag) and --amend
- Allows heredoc-style messages (too complex to parse reliably)
- Denies malformed messages with guidance pointing to commit:behavior skill
- Returns {"hookSpecificOutput": {"permissionDecision": "deny"|"allow"}}

Input (stdin JSON):
  {"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"feat: add feature\""}}

Output (stdout JSON):
  {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow"}}

Performance: <1ms for non-git-commit commands (fast-path).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runGitConventions(ctx)
			})
		},
	}

	return cmd
}

func runGitConventions(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runGitConventionsCore(ctx, printer)
}

// runGitConventionsCore contains the actual logic with injected printer for testing.
func runGitConventionsCore(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Fast-path: only handle PreToolUse events
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPreToolUse {
		return outputGitConventionsAllow(printer)
	}

	// Fast-path: only handle Bash tool
	if hookEnv.ToolName != "Bash" {
		return outputGitConventionsAllow(printer)
	}

	// Parse command from tool input (reuse parseCommand from validate.go)
	command := parseCommand(printer, hookEnv.ToolInput)
	if command == "" {
		return outputGitConventionsAllow(printer)
	}

	// Fast-path: not a git commit command
	if !gitCommitPattern.MatchString(command) {
		return outputGitConventionsAllow(printer)
	}

	// Allow --amend (modifying existing commit message)
	if amendPattern.MatchString(command) {
		return outputGitConventionsAllow(printer)
	}

	// Allow heredoc-style messages (too complex to parse reliably)
	if heredocPattern.MatchString(command) {
		return outputGitConventionsAllow(printer)
	}

	// Extract commit message from -m flag
	message := extractCommitMessage(command)
	if message == "" {
		// No -m flag means interactive commit (user will provide message)
		// or some other form we can't parse — allow
		return outputGitConventionsAllow(printer)
	}

	// Validate against conventional commit format
	if !conventionalFormat.MatchString(message) {
		return outputGitConventionsDeny(printer, conventionDenyReason)
	}

	return outputGitConventionsAllow(printer)
}

// extractCommitMessage extracts the message string from a git commit -m command.
// Returns empty string if no -m flag is found.
func extractCommitMessage(command string) string {
	matches := commitMessageFlag.FindStringSubmatch(command)
	if matches == nil {
		return ""
	}
	// Group 1 = double-quoted, Group 2 = single-quoted
	if matches[1] != "" {
		return matches[1]
	}
	return matches[2]
}

// outputGitConventionsAllow outputs an allow decision.
func outputGitConventionsAllow(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
		},
	}
	return printer.Print(result)
}

// outputGitConventionsDeny outputs a deny decision with a reason.
func outputGitConventionsDeny(printer *output.Printer, reason string) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
	return printer.Print(result)
}
