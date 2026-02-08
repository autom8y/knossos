// Package hook implements the ari hook commands.
package hook

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
)

// ClewTriggerOutput represents the trigger detection result in clew output.
type ClewTriggerOutput struct {
	Triggered bool   `json:"triggered"`
	Type      string `json:"type,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Suggest   string `json:"suggest,omitempty"`
}

// ClewOutput represents the output of the clew hook.
type ClewOutput struct {
	Recorded   bool               `json:"recorded"`
	Reason     string             `json:"reason,omitempty"`
	EventsFile string             `json:"events_file,omitempty"`
	Trigger    *ClewTriggerOutput `json:"trigger,omitempty"`
}

// Text implements output.Textable for text output.
func (t ClewOutput) Text() string {
	if t.Recorded {
		msg := "Event recorded to " + t.EventsFile
		if t.Trigger != nil && t.Trigger.Triggered {
			msg += "\n" + t.Trigger.Suggest
		}
		return msg
	}
	return "Not recorded: " + t.Reason
}

// newClewCmd creates the clew hook subcommand.
func newClewCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clew",
		Short: "Record tool events on PostToolUse",
		Long: `Records tool events to events.jsonl for Clew Contract v2.

This hook is triggered on PostToolUse events. It:
- Reads CLAUDE_HOOK_TOOL_INPUT environment variable
- Parses the tool input JSON
- Calls RecordToolEvent to write to events.jsonl
- Returns JSON: {"recorded": true} or {"recorded": false, "reason": "..."}

The clew hook implements the "Clew Contract" pattern:
"Theseus has amnesia; the Clew remembers"

Events provide the factual route through decisions for session recovery
and debugging. (Ariadne gave Theseus a CLEW - ball of thread - to navigate the labyrinth)

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runClew(ctx)
			})
		},
	}

	return cmd
}

// runClew is the cobra RunE handler that creates the printer and delegates to runClewCore.
func runClew(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runClewCore(ctx, printer)
}

// getSessionDir determines the session directory from context and environment.
func getSessionDir(ctx *cmdContext, hookEnv *hook.Env) string {
	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil || sessionID == "" || resolver.ProjectRoot() == "" {
		return ""
	}

	// Return the session directory path
	return resolver.SessionDir(sessionID)
}

// outputNotRecorded outputs the not-recorded response with the given printer.
func outputNotRecorded(printer *output.Printer, reason string) error {
	result := ClewOutput{
		Recorded: false,
		Reason:   reason,
	}
	return printer.Print(result)
}

// runClewCore contains the actual clew hook logic. It accepts an injected printer
// for testing purposes, allowing tests to capture output without creating a full
// command context.
func runClewCore(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PostToolUse event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPostToolUse {
		printer.VerboseLog("debug", "skipping clew hook for non-PostToolUse event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputNotRecorded(printer, "not a PostToolUse event")
	}

	// Get session directory
	sessionDir := getSessionDir(ctx, hookEnv)
	if sessionDir == "" {
		return outputNotRecorded(printer, "no active session")
	}

	// Parse tool input from environment
	toolInputJSON := hookEnv.ToolInput
	toolInput, err := hook.ParseToolInput(toolInputJSON)
	if err != nil {
		printer.VerboseLog("warn", "failed to parse tool input",
			map[string]interface{}{"error": err.Error()})
		// Emit error event for parse failure (graceful degradation)
		emitErrorEvent(sessionDir, "TOOL_INPUT_PARSE", "failed to parse tool input: "+err.Error(), true, printer)
		return outputNotRecorded(printer, "invalid tool input: "+err.Error())
	}

	// Build event for trigger checking (before recording)
	event := clewcontract.BuildEventFromToolInput(hookEnv, toolInput)

	// Record the tool event
	if err := clewcontract.RecordToolEvent(sessionDir, hookEnv, toolInput); err != nil {
		printer.VerboseLog("error", "failed to record tool event",
			map[string]interface{}{"error": err.Error()})
		// Emit error event for write failure (graceful degradation)
		emitErrorEvent(sessionDir, "CLEW_WRITE", "failed to record tool event: "+err.Error(), true, printer)
		return outputNotRecorded(printer, "write failed: "+err.Error())
	}

	// Emit supplemental events for Edit/Write tools
	emitSupplementalEvents(sessionDir, hookEnv.ToolName, toolInput, printer)

	// Check if this is an orchestrator Task completion and record throughline stamp
	if hookEnv.ToolName == "Task" && hookEnv.ToolResult != "" {
		if throughline := clewcontract.ExtractThroughline(hookEnv.ToolResult); throughline != nil {
			// Record decision stamp (fail silently - don't break hook if stamp fails)
			if err := clewcontract.RecordStamp(sessionDir, throughline.Decision, throughline.Rationale, nil); err != nil {
				printer.VerboseLog("warn", "failed to record orchestrator stamp",
					map[string]interface{}{"error": err.Error()})
				// Continue - stamp failure is not critical
			} else {
				printer.VerboseLog("debug", "recorded orchestrator decision stamp",
					map[string]interface{}{"decision": throughline.Decision})
			}
		}
	}

	// Check triggers after recording
	eventsPath := clewcontract.GetEventsPath(sessionDir)
	triggerConfig := clewcontract.DefaultTriggerConfig()
	triggerResult := clewcontract.CheckTriggers(eventsPath, event, triggerConfig)

	// Build output
	result := ClewOutput{
		Recorded:   true,
		EventsFile: eventsPath,
	}

	// Include trigger if triggered
	if triggerResult.Triggered {
		result.Trigger = &ClewTriggerOutput{
			Triggered: true,
			Type:      string(triggerResult.Type),
			Reason:    triggerResult.Reason,
			Suggest:   triggerResult.Suggest,
		}
	}

	return printer.Print(result)
}

// artifactPatterns maps filename patterns to artifact types.
// Used to detect when a Write tool creates a known artifact.
var artifactPatterns = map[string]clewcontract.ArtifactType{
	"PRD-":            clewcontract.ArtifactTypePRD,
	"TDD-":            clewcontract.ArtifactTypeTDD,
	"ADR-":            clewcontract.ArtifactTypeADR,
	"SESSION_CONTEXT": clewcontract.ArtifactTypePRD, // session artifact
	"SPRINT_CONTEXT":  clewcontract.ArtifactTypePRD, // sprint artifact
}

// emitSupplementalEvents emits file_change and artifact_created events
// for Edit/Write tools. These supplement the primary tool_call event.
// All emissions are best-effort -- failures are logged but do not affect the hook result.
func emitSupplementalEvents(sessionDir, toolName string, toolInput *hook.ToolInput, printer *output.Printer) {
	if toolName != "Edit" && toolName != "Write" {
		return
	}

	path := toolInput.GetEffectivePath()
	if path == "" {
		return
	}

	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer writer.Close()

	// Emit file_change event for Edit/Write tools
	linesChanged := 0
	if toolInput.Content != "" {
		linesChanged = strings.Count(toolInput.Content, "\n") + 1
	} else if toolInput.NewString != "" {
		linesChanged = strings.Count(toolInput.NewString, "\n") + 1
	}
	fileChangeEvent := clewcontract.NewFileChangeEvent(path, linesChanged)
	writer.Write(fileChangeEvent)

	// Emit artifact_created event for Write to known artifact paths
	if toolName == "Write" {
		if artType, phase := matchArtifactPattern(path); artType != "" {
			artifactEvent := clewcontract.NewArtifactCreatedEvent(artType, path, phase, "")
			writer.Write(artifactEvent)
		}
	}

	if flushErr := writer.Flush(); flushErr != nil {
		printer.VerboseLog("warn", "failed to emit supplemental events",
			map[string]interface{}{"error": flushErr.Error()})
	}
}

// matchArtifactPattern checks if a file path matches a known artifact pattern.
// Returns the artifact type and inferred phase, or empty strings if no match.
func matchArtifactPattern(path string) (clewcontract.ArtifactType, string) {
	base := filepath.Base(path)

	for prefix, artType := range artifactPatterns {
		if strings.HasPrefix(base, prefix) {
			phase := artifactTypeToPhase(artType)
			return artType, phase
		}
	}
	return "", ""
}

// artifactTypeToPhase maps artifact types to workflow phases.
func artifactTypeToPhase(artType clewcontract.ArtifactType) string {
	switch artType {
	case clewcontract.ArtifactTypePRD:
		return "requirements"
	case clewcontract.ArtifactTypeTDD:
		return "design"
	case clewcontract.ArtifactTypeADR:
		return "design"
	case clewcontract.ArtifactTypeTestPlan:
		return "validation"
	default:
		return ""
	}
}

// emitErrorEvent emits a structured error event to the clew log.
// This provides error visibility in the event stream for diagnostics.
// All emissions are best-effort -- failures are logged but do not propagate.
func emitErrorEvent(sessionDir, errorCode, message string, recoverable bool, printer *output.Printer) {
	if sessionDir == "" {
		return
	}

	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer writer.Close()

	errorEvent := clewcontract.NewErrorEvent(errorCode, message, "clew hook", recoverable, "check hook logs")
	writer.Write(errorEvent)

	if flushErr := writer.Flush(); flushErr != nil {
		printer.VerboseLog("warn", "failed to emit error event",
			map[string]interface{}{"error": flushErr.Error()})
	}
}
