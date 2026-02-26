// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/registry"
)

// ToolInput represents the input from Claude Code PreToolUse hook.
type ToolInput struct {
	ToolName string `json:"tool_name"`
	FilePath string `json:"file_path"`
}

// SectionClass identifies which section of SESSION_CONTEXT.md an Edit targets.
// Used by classifyEditSection to route to the appropriate permission path.
type SectionClass int

const (
	// SectionUnknown means no recognizable section indicators were found. Fail closed.
	SectionUnknown SectionClass = iota
	// SectionTimeline means the edit targets only the Timeline section. Lockless allow.
	SectionTimeline
	// SectionFrontmatter means the edit targets only YAML frontmatter. Requires Moirai lock.
	SectionFrontmatter
	// SectionOther means the edit targets only a non-Timeline body section. Requires Moirai lock.
	SectionOther
	// SectionMixed means multiple section types were detected in old_string. Fail closed.
	SectionMixed
)

// Compiled regex patterns for section detection in SESSION_CONTEXT.md.
// Compiled at package init (not per-invocation) per performance requirement D7.
var (
	// timelineEntryRe matches SESSION-2 timeline entry format: "- HH:MM | CATEGORY | summary"
	timelineEntryRe = regexp.MustCompile(`^- \d{2}:\d{2} \| `)
	// timelineHeadingRe matches the exact Timeline section heading.
	timelineHeadingRe = regexp.MustCompile(`^## Timeline$`)
	// frontmatterDelimRe matches YAML frontmatter delimiter on its own line.
	frontmatterDelimRe = regexp.MustCompile(`^---$`)
	// frontmatterKeyRe matches any known SESSION_CONTEXT.md frontmatter key at line start.
	// All 17 keys from SESSION-3 Section 2.1 are listed to prevent false negatives.
	frontmatterKeyRe = regexp.MustCompile(`^(schema_version|session_id|status|created_at|initiative|complexity|active_rite|rite|current_phase|timeline_version|parked_at|parked_reason|archived_at|resumed_at|frayed_from|fray_point|strands):`)
	// otherSectionRe matches any H2 heading. Combined with timelineHeadingRe exclusion,
	// this catches standard sections (Artifacts, Blockers, Next Steps) and custom sections.
	otherSectionRe = regexp.MustCompile(`^## .+$`)
)

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
		// SESSION_CONTEXT.md gets section-aware permission routing for Edit tool.
		// Only old_string is analyzed (decision D1: new_string does not indicate location).
		// Write tool always requires Moirai lock (decision D6: full file replacement).
		//
		// sectionClass tracks the classification for SESSION_CONTEXT Edit operations so
		// the lock-miss path can emit the correct E3/E5 advisory vs. generic W2.
		sectionClass := SectionUnknown
		if isSessionContext(filePath) && toolName == "Edit" {
			sectionClass = classifyEditSection(hookEnv.ToolInput)
			switch sectionClass {
			case SectionTimeline:
				// E1: Timeline edits are lockless. Advisory context reminds model that
				// other sections still require Moirai lock.
				return outputAllowTimeline(printer)
			case SectionMixed:
				// E6: Edit spans multiple sections. Block and advise split edits.
				return outputBlockMixed(printer)
			case SectionUnknown:
				// E7: No recognizable indicators. Fail closed.
				return outputBlockUnknown(printer)
			// SectionFrontmatter and SectionOther fall through to Moirai lock check.
			}
		}

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
		// Emit section-specific advisory for SESSION_CONTEXT Edit operations (E3/E5).
		// For Write operations and other protected files, use the generic block message.
		if isSessionContext(filePath) && toolName == "Edit" {
			switch sectionClass {
			case SectionFrontmatter:
				return outputBlockFrontmatter(printer)
			case SectionOther:
				return outputBlockOtherSection(printer)
			}
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

	var input map[string]any
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		printer.VerboseLog("warn", "failed to parse tool input JSON",
			map[string]any{"error": err.Error(), "input": toolInput})
		return ""
	}

	if fp, ok := input["file_path"].(string); ok {
		return fp
	}
	return ""
}

// isSessionContext returns true if filePath targets a SESSION_CONTEXT.md file.
// Section-aware permission routing only applies to SESSION_CONTEXT.md (decision D5).
func isSessionContext(filePath string) bool {
	return strings.HasSuffix(filePath, "SESSION_CONTEXT.md")
}

// parseOldString extracts the old_string field from a JSON tool_input string.
// Follows the same pattern as parseFilePath. Returns empty string if absent or
// if JSON is unparseable (which results in SectionUnknown / fail-closed).
func parseOldString(toolInput string) string {
	if toolInput == "" {
		return ""
	}
	var input map[string]any
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		return ""
	}
	if s, ok := input["old_string"].(string); ok {
		return s
	}
	return ""
}

// isTimelineIndicator returns true if line matches a Timeline section indicator.
// Checks T1 (timeline entry format) and T2 (## Timeline heading).
func isTimelineIndicator(line string) bool {
	return timelineEntryRe.MatchString(line) || timelineHeadingRe.MatchString(line)
}

// isFrontmatterIndicator returns true if line matches a frontmatter indicator.
// Checks F1 (YAML delimiter ---) and F2 (known frontmatter key at line start).
func isFrontmatterIndicator(line string) bool {
	return frontmatterDelimRe.MatchString(line) || frontmatterKeyRe.MatchString(line)
}

// isOtherSectionIndicator returns true if line is an H2 heading that is not ## Timeline.
// This matches standard sections (Artifacts, Blockers, Next Steps) and custom sections.
// Note: ## Timeline is handled by isTimelineIndicator (T2), so it is excluded here.
func isOtherSectionIndicator(line string) bool {
	return otherSectionRe.MatchString(line) && !timelineHeadingRe.MatchString(line)
}

// classifyEditSection inspects old_string from toolInput JSON and determines
// which section of SESSION_CONTEXT.md the Edit targets.
//
// Algorithm (per SESSION-4 Section 4.2):
//   - Each non-blank line is classified against Timeline, Frontmatter, or OtherSection indicators.
//   - Lines that match no indicator are "context lines" and are neutral (decision D4).
//   - If no positive indicators are found: Unknown (fail-closed, decision D2).
//   - If multiple indicator types are found: Mixed (fail-closed, decision D3).
//   - Otherwise: the single indicator type found.
func classifyEditSection(toolInput string) SectionClass {
	oldString := parseOldString(toolInput)
	if oldString == "" {
		// Empty old_string: no indicators possible. Fail closed.
		return SectionUnknown
	}

	lines := strings.Split(oldString, "\n")
	hasTimeline := false
	hasFrontmatter := false
	hasOther := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			// Blank lines are neutral — skip.
			continue
		}
		if isTimelineIndicator(trimmed) {
			hasTimeline = true
		} else if isFrontmatterIndicator(trimmed) {
			hasFrontmatter = true
		} else if isOtherSectionIndicator(trimmed) {
			hasOther = true
		}
		// Lines matching no indicator are context lines — they do not influence
		// classification. This prevents bypass via generic text (decision D4).
	}

	// Aggregate: count how many indicator types fired.
	flagCount := 0
	if hasTimeline {
		flagCount++
	}
	if hasFrontmatter {
		flagCount++
	}
	if hasOther {
		flagCount++
	}

	if flagCount == 0 {
		return SectionUnknown // No recognizable indicators: fail closed.
	}
	if flagCount > 1 {
		return SectionMixed // Multiple section types: fail closed.
	}
	if hasTimeline {
		return SectionTimeline
	}
	if hasFrontmatter {
		return SectionFrontmatter
	}
	return SectionOther
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

	if moiraiLock.Agent != registry.Ref(registry.AgentMoirai) {
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
			AdditionalContext:        "Session " + sessionID + " was previously wrapped with '" + registry.Ref(registry.CLISessionWrap) + "' and is now immutable. Archived session data is preserved at .claude/.archive/sessions/" + sessionID + "/",
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

	var input map[string]any
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		printer.VerboseLog("warn", "failed to parse tool input JSON for content field",
			map[string]any{"error": err.Error()})
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

	var fields map[string]any
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

// outputAllowTimeline outputs an allow decision with advisory context for E1 (Timeline edits).
// The advisory reminds the model that only timeline writes are lockless — other sections
// still require the Moirai lock or ari CLI commands (decision D8).
func outputAllowTimeline(printer *output.Printer) error {
	return outputAllowWithContext(printer,
		"Timeline append allowed without Moirai lock. For frontmatter or body section changes, "+
			"use "+registry.Ref(registry.CLISessionFieldSet)+" (frontmatter) or "+registry.TaskDelegation(registry.AgentMoirai)+" (lifecycle transitions).")
}

// outputBlockFrontmatter outputs a deny decision for E3 (frontmatter Edit without lock).
// Directs the model to ari session field-set or Moirai delegation.
func outputBlockFrontmatter(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "SESSION_CONTEXT frontmatter requires Moirai lock",
			AdditionalContext: "To modify frontmatter fields, use: " + registry.Ref(registry.CLISessionFieldSet) + " <field> <value> " +
				"(for settable fields: initiative, complexity, active_rite), or " +
				registry.TaskDelegation(registry.AgentMoirai) + " for lifecycle-controlled fields.",
		},
	}
	return printer.Print(result)
}

// outputBlockOtherSection outputs a deny decision for E5 (body section Edit without lock).
// Directs the model to Moirai delegation or ari session log for timeline appends.
func outputBlockOtherSection(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "SESSION_CONTEXT body section requires Moirai lock",
			AdditionalContext: "Body sections (Artifacts, Blockers, Next Steps) are Moirai-managed. " +
				"Use " + registry.TaskDelegation(registry.AgentMoirai) + " for section updates. " +
				"For timeline appends, use: " + registry.Ref(registry.CLISessionLog) + " --type=<type> '<summary>'.",
		},
	}
	return printer.Print(result)
}

// outputBlockMixed outputs a deny decision for E6 (Mixed section edit).
// The edit's old_string spans multiple sections; model must split into separate edits.
func outputBlockMixed(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "Edit targets multiple SESSION_CONTEXT sections",
			AdditionalContext: "This edit spans multiple sections (e.g., timeline + frontmatter or timeline + body). " +
				"Split into separate edits: use " + registry.Ref(registry.CLISessionLog) + " for timeline appends and " + registry.TaskDelegation(registry.AgentMoirai) + " for other mutations.",
		},
	}
	return printer.Print(result)
}

// outputBlockUnknown outputs a deny decision for E7 (Unknown section).
// No recognizable section indicators were found in old_string. Fail closed.
func outputBlockUnknown(printer *output.Printer) error {
	result := hook.PreToolUseOutput{
		HookSpecificOutput: hook.HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: "Cannot determine target section in SESSION_CONTEXT",
			AdditionalContext: "The edit content did not match any recognized section pattern. " +
				"Use " + registry.Ref(registry.CLISessionLog) + " --type=<type> '<summary>' for timeline entries, " +
				registry.Ref(registry.CLISessionFieldSet) + " for frontmatter fields, or " + registry.TaskDelegation(registry.AgentMoirai) + " for lifecycle operations.",
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
		moiraiOp = registry.TaskDelegation(registry.AgentMoirai, "transition_phase", "update_field", "park_session", "resume_session", "handoff", "record_decision")
	} else if strings.HasSuffix(filePath, "SPRINT_CONTEXT.md") {
		contextType = "SPRINT_CONTEXT"
		moiraiOp = registry.TaskDelegation(registry.AgentMoirai, "create_sprint", "mark_complete", "update_field")
	} else {
		contextType = "context file"
		moiraiOp = registry.TaskDelegation(registry.AgentMoirai)
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
