// Package hook implements the ari hook commands.
package hook

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/hook"
	"github.com/autom8y/ariadne/internal/hook/threadcontract"
	"github.com/autom8y/ariadne/internal/output"
)

// ThreadTriggerOutput represents the trigger detection result in thread output.
type ThreadTriggerOutput struct {
	Triggered bool   `json:"triggered"`
	Type      string `json:"type,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Suggest   string `json:"suggest,omitempty"`
}

// ThreadOutput represents the output of the thread hook.
type ThreadOutput struct {
	Recorded   bool                 `json:"recorded"`
	Reason     string               `json:"reason,omitempty"`
	EventsFile string               `json:"events_file,omitempty"`
	Trigger    *ThreadTriggerOutput `json:"trigger,omitempty"`
}

// Text implements output.Textable for text output.
func (t ThreadOutput) Text() string {
	if t.Recorded {
		msg := "Event recorded to " + t.EventsFile
		if t.Trigger != nil && t.Trigger.Triggered {
			msg += "\n" + t.Trigger.Suggest
		}
		return msg
	}
	return "Not recorded: " + t.Reason
}

// newThreadCmd creates the thread hook subcommand.
func newThreadCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread",
		Short: "Record tool events on PostToolUse",
		Long: `Records tool events to events.jsonl for Thread Contract v2.

This hook is triggered on PostToolUse events. It:
- Reads CLAUDE_HOOK_TOOL_INPUT environment variable
- Parses the tool input JSON
- Calls RecordToolEvent to write to events.jsonl
- Returns JSON: {"recorded": true} or {"recorded": false, "reason": "..."}

The thread hook implements the "Thread Contract" pattern:
"Theseus has amnesia; the Thread remembers"

Events provide the factual route through decisions for session recovery
and debugging.

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runThread(ctx)
			})
		},
	}

	return cmd
}

func runThread(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	// Early exit if hooks disabled
	if ctx.shouldEarlyExit() {
		return outputNotRecorded(printer, "hooks disabled")
	}

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PostToolUse event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPostToolUse {
		printer.VerboseLog("debug", "skipping thread hook for non-PostToolUse event",
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
	if toolInputJSON == "" {
		toolInputJSON = os.Getenv("CLAUDE_HOOK_TOOL_INPUT")
	}

	toolInput, err := hook.ParseToolInput(toolInputJSON)
	if err != nil {
		printer.VerboseLog("warn", "failed to parse tool input",
			map[string]interface{}{"error": err.Error()})
		return outputNotRecorded(printer, "invalid tool input: "+err.Error())
	}

	// Build event for trigger checking (before recording)
	event := threadcontract.BuildEventFromToolInput(hookEnv, toolInput)

	// Record the tool event
	if err := threadcontract.RecordToolEvent(sessionDir, hookEnv, toolInput); err != nil {
		printer.VerboseLog("error", "failed to record tool event",
			map[string]interface{}{"error": err.Error()})
		return outputNotRecorded(printer, "write failed: "+err.Error())
	}

	// Check if this is an orchestrator Task completion and record throughline stamp
	if hookEnv.ToolName == "Task" && hookEnv.ToolResult != "" {
		if throughline := threadcontract.ExtractThroughline(hookEnv.ToolResult); throughline != nil {
			// Record decision stamp (fail silently - don't break hook if stamp fails)
			if err := threadcontract.RecordStamp(sessionDir, throughline.Decision, throughline.Rationale, nil); err != nil {
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
	eventsPath := threadcontract.GetEventsPath(sessionDir)
	triggerConfig := threadcontract.DefaultTriggerConfig()
	triggerResult := threadcontract.CheckTriggers(eventsPath, event, triggerConfig)

	// Build output
	result := ThreadOutput{
		Recorded:   true,
		EventsFile: eventsPath,
	}

	// Include trigger if triggered
	if triggerResult.Triggered {
		result.Trigger = &ThreadTriggerOutput{
			Triggered: true,
			Type:      string(triggerResult.Type),
			Reason:    triggerResult.Reason,
			Suggest:   triggerResult.Suggest,
		}
	}

	return printer.Print(result)
}

// getSessionDir determines the session directory from context and environment.
func getSessionDir(ctx *cmdContext, hookEnv *hook.Env) string {
	// Try to get session ID from context
	sessionID, err := ctx.getCurrentSessionID()
	if err != nil || sessionID == "" {
		return ""
	}

	sessionID = strings.TrimSpace(sessionID)

	// Get resolver for path lookups
	resolver := ctx.getResolver()
	if resolver.ProjectRoot() == "" {
		// Try to discover project from environment
		if hookEnv.ProjectDir != "" {
			resolver = newResolverFromPath(hookEnv.ProjectDir)
		} else {
			return ""
		}
	}

	// Return the session directory path
	return resolver.SessionDir(sessionID)
}

// outputNotRecorded outputs the not-recorded response.
func outputNotRecorded(printer *output.Printer, reason string) error {
	result := ThreadOutput{
		Recorded: false,
		Reason:   reason,
	}
	return printer.Print(result)
}

// runThreadWithPrinter is a helper that uses an injected printer for testing.
func runThreadWithPrinter(ctx *cmdContext, printer *output.Printer) error {
	// Early exit if hooks disabled
	if ctx.shouldEarlyExit() {
		return outputNotRecordedWithPrinter(printer, "hooks disabled")
	}

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a PostToolUse event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPostToolUse {
		printer.VerboseLog("debug", "skipping thread hook for non-PostToolUse event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputNotRecordedWithPrinter(printer, "not a PostToolUse event")
	}

	// Get session directory
	sessionDir := getSessionDir(ctx, hookEnv)
	if sessionDir == "" {
		return outputNotRecordedWithPrinter(printer, "no active session")
	}

	// Parse tool input from environment
	toolInputJSON := hookEnv.ToolInput
	if toolInputJSON == "" {
		toolInputJSON = os.Getenv("CLAUDE_HOOK_TOOL_INPUT")
	}

	toolInput, err := hook.ParseToolInput(toolInputJSON)
	if err != nil {
		printer.VerboseLog("warn", "failed to parse tool input",
			map[string]interface{}{"error": err.Error()})
		return outputNotRecordedWithPrinter(printer, "invalid tool input: "+err.Error())
	}

	// Build event for trigger checking (before recording)
	event := threadcontract.BuildEventFromToolInput(hookEnv, toolInput)

	// Record the tool event
	if err := threadcontract.RecordToolEvent(sessionDir, hookEnv, toolInput); err != nil {
		printer.VerboseLog("error", "failed to record tool event",
			map[string]interface{}{"error": err.Error()})
		return outputNotRecordedWithPrinter(printer, "write failed: "+err.Error())
	}

	// Check if this is an orchestrator Task completion and record throughline stamp
	if hookEnv.ToolName == "Task" && hookEnv.ToolResult != "" {
		if throughline := threadcontract.ExtractThroughline(hookEnv.ToolResult); throughline != nil {
			// Record decision stamp (fail silently - don't break hook if stamp fails)
			if err := threadcontract.RecordStamp(sessionDir, throughline.Decision, throughline.Rationale, nil); err != nil {
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
	eventsPath := threadcontract.GetEventsPath(sessionDir)
	triggerConfig := threadcontract.DefaultTriggerConfig()
	triggerResult := threadcontract.CheckTriggers(eventsPath, event, triggerConfig)

	// Build output
	result := ThreadOutput{
		Recorded:   true,
		EventsFile: eventsPath,
	}

	// Include trigger if triggered
	if triggerResult.Triggered {
		result.Trigger = &ThreadTriggerOutput{
			Triggered: true,
			Type:      string(triggerResult.Type),
			Reason:    triggerResult.Reason,
			Suggest:   triggerResult.Suggest,
		}
	}

	return printer.Print(result)
}

// outputNotRecordedWithPrinter outputs the not-recorded response with an injected printer.
func outputNotRecordedWithPrinter(printer *output.Printer, reason string) error {
	result := ThreadOutput{
		Recorded: false,
		Reason:   reason,
	}
	return printer.Print(result)
}
