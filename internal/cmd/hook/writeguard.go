// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/frontmatter"
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


// Protected file patterns for context and platform infrastructure files.
var protectedPatterns = []string{
	"SESSION_CONTEXT.md",
	"SPRINT_CONTEXT.md",
	"PROVENANCE_MANIFEST.yaml",
	"KNOSSOS_MANIFEST.yaml",
	"settings.local.json",
}

// validWipTypes is the closed taxonomy of valid .wip/ artifact type values.
// Values are case-sensitive lowercase. The set is intentionally small to keep
// the classification meaningful.
var validWipTypes = map[string]bool{
	"spike":   true,
	"spec":    true,
	"audit":   true,
	"design":  true,
	"triage":  true,
	"qa":      true,
	"scratch": true,
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

	// .wip/ validation: only applies to Write tool (not Edit — edits to existing files
	// skip frontmatter re-validation per design). .wip/ paths are never protected, so
	// we return early here and skip the protected-file check entirely.
	if toolName == "Write" && isWipPath(filePath) {
		content := parseContentField(printer, hookEnv.ToolInput)
		valid, _, reason := validateWipFrontmatter(content)
		if valid {
			return outputAllow(printer)
		}
		return outputAllowWithContext(printer, reason)
	}

	// Check if file is protected
	if isProtectedFile(filePath) {
		// Resolve session via priority chain, then check Moirai lock
		resolver, sessionID, _ := ctx.resolveSession(hookEnv)
		// Fallback: extract session ID from file path for PARKED sessions
		// that aren't visible to FindActiveSessions()
		if sessionID == "" {
			sessionID = extractSessionIDFromPath(filePath)
		}
		if sessionID != "" && isMoiraiLockHeld(resolver, sessionID) {
			return outputAllow(printer)
		}
		// If Moirai lock not held, check whether the session is archived.
		// Archived sessions are terminal — deny with a clear message instead
		// of the generic "Use Moirai" guidance which implies a recoverable state.
		if sessionID != "" && isSessionArchived(resolver, sessionID) {
			return outputBlockArchived(printer, sessionID)
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

// extractSessionIDFromPath extracts a session ID from a file path.
// Looks for path segments matching the session-YYYYMMDD-HHMMSS-{hex} pattern.
// Returns empty string if no session ID found.
func extractSessionIDFromPath(filePath string) string {
	parts := strings.Split(filePath, "/")
	for _, part := range parts {
		if paths.IsSessionDir(part) {
			return part
		}
	}
	return ""
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

// isSessionArchived checks if a session exists in the archive directory.
// Returns true only if the archive directory for the given session ID exists.
// This check is O(1) — a single stat call. Only runs when Moirai lock is not held.
func isSessionArchived(resolver *paths.Resolver, sessionID string) bool {
	archivePath := resolver.ArchiveDir() + "/" + sessionID
	_, err := os.Stat(archivePath)
	return err == nil
}

// outputBlockArchived outputs a deny decision with a message specific to archived sessions.
// Archived sessions are terminal — the "Use Moirai" delegation guidance is inappropriate
// because Moirai cannot operate on archived sessions either.
func outputBlockArchived(printer *output.Printer, sessionID string) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "Session " + sessionID + " is archived (terminal state). Context files cannot be mutated after archiving.",
			AdditionalContext:        "Session " + sessionID + " was previously wrapped with 'ari session wrap' and is now immutable. Archived session data is preserved at .claude/.archive/sessions/" + sessionID + "/",
		},
	}
	return printer.Print(result)
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

// isWipPath returns true if filePath targets a .wip/ directory.
// Matches both relative paths starting with ".wip/" and absolute paths containing
// "/.wip/" as a path segment. The matching is intentionally broad: a false positive
// (validating a non-root .wip/ write) is harmless — it just advises on frontmatter.
func isWipPath(filePath string) bool {
	return strings.HasPrefix(filePath, ".wip/") || strings.Contains(filePath, "/.wip/")
}

// parseContentField extracts the "content" field from a JSON tool_input string.
// Reuses the same pattern as parseFilePath. Returns empty string if the field is
// missing or the JSON is unparseable. Logs a verbose warning on parse failure.
func parseContentField(printer *output.Printer, toolInput string) string {
	if toolInput == "" {
		return ""
	}

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		printer.VerboseLog("warn", "failed to parse tool input JSON for content field",
			map[string]interface{}{"error": err.Error()})
		return ""
	}

	if content, ok := input["content"].(string); ok {
		return content
	}
	return ""
}

// validateWipFrontmatter parses YAML frontmatter from content and validates the type field.
// Returns (true, typeValue, "") on success.
// Returns (false, "", reason) on failure with an actionable reason string.
func validateWipFrontmatter(content string) (bool, string, string) {
	yamlBytes, _, err := frontmatter.Parse([]byte(content))
	if err != nil {
		return false, "", ".wip/ files require YAML frontmatter. Add to the top of your file:\n---\ntype: <spike|spec|audit|design|triage|qa|scratch>\n---"
	}

	var fields map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &fields); err != nil {
		return false, "", ".wip/ files require YAML frontmatter. Add to the top of your file:\n---\ntype: <spike|spec|audit|design|triage|qa|scratch>\n---"
	}

	typeVal, ok := fields["type"]
	if !ok {
		return false, "", ".wip/ frontmatter must include a type field. Valid types: spike, spec, audit, design, triage, qa, scratch"
	}

	typeStr, ok := typeVal.(string)
	if !ok || typeStr == "" {
		return false, "", ".wip/ frontmatter must include a type field. Valid types: spike, spec, audit, design, triage, qa, scratch"
	}

	if !validWipTypes[typeStr] {
		return false, "", fmt.Sprintf(".wip/ frontmatter type %q is not valid. Valid types: spike, spec, audit, design, triage, qa, scratch", typeStr)
	}

	return true, typeStr, ""
}

// outputAllowWithContext outputs an allow decision with advisory additionalContext.
// CC surfaces additionalContext to the model so it can self-correct on the next write.
func outputAllowWithContext(printer *output.Printer, context string) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			AdditionalContext:  context,
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
		moiraiOp = `Task(moirai, "<operation> ...") — operations: transition_phase, update_field, park_session, resume_session, handoff, record_decision`
	} else if strings.HasSuffix(filePath, "SPRINT_CONTEXT.md") {
		contextType = "SPRINT_CONTEXT"
		moiraiOp = `Task(moirai, "<operation> ...") — operations: create_sprint, mark_complete, update_field`
	} else {
		contextType = "context file"
		moiraiOp = `Task(moirai, "<operation> ...")`
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
