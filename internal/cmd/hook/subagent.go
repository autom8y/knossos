// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
)

// subagentResult is the output of subagent hooks.
// SubagentStart/SubagentStop are side-effect hooks (logging only) -- they cannot block.
type subagentResult struct {
	Recorded bool   `json:"recorded"`
	Reason   string `json:"reason,omitempty"`
}

// subagentPayload represents the relevant fields from the stdin JSON payload
// for SubagentStart and SubagentStop events.
type subagentPayload struct {
	AgentName string `json:"agent_name"`
	AgentType string `json:"agent_type"`
	TaskID    string `json:"task_id"`
}

// newSubagentStartCmd creates the SubagentStart hook subcommand.
func newSubagentStartCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subagent-start",
		Short: "Log subagent spawn on SubagentStart",
		Long: `Logs subagent spawn events to the session clew.

This hook is triggered on SubagentStart events. It:
- Parses stdin payload for agent name/type
- Logs the subagent spawn event to events.jsonl
- Returns empty JSON (side-effect hook, no blocking)

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runSubagentStart(ctx)
			})
		},
	}

	return cmd
}

// newSubagentStopCmd creates the SubagentStop hook subcommand.
func newSubagentStopCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subagent-stop",
		Short: "Log subagent completion on SubagentStop",
		Long: `Logs subagent completion events to the session clew.

This hook is triggered on SubagentStop events. It:
- Parses stdin payload for agent completion info
- Logs the subagent completion event to events.jsonl
- Returns empty JSON (side-effect hook, no blocking)

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runSubagentStop(ctx)
			})
		},
	}

	return cmd
}

func runSubagentStart(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runSubagentStartCore(ctx, printer)
}

func runSubagentStop(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runSubagentStopCore(ctx, printer)
}

// runSubagentStartCore contains the SubagentStart hook logic.
func runSubagentStartCore(ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv()

	// Verify this is a SubagentStart event (or empty for testing)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSubagentStart {
		return outputSubagentResult(printer, false, "not a SubagentStart event")
	}

	// Get session directory for clew logging
	sessionDir := getSessionDir(ctx, hookEnv)
	if sessionDir == "" {
		return outputSubagentResult(printer, false, "no active session")
	}

	// Parse agent info from tool input (CC sends agent details in tool_input)
	agentInfo := parseSubagentInfo(hookEnv.ToolInput)

	// Log to clew using a task_start event
	event := clewcontract.Event{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Type:      clewcontract.EventTypeTaskStart,
		Summary:   fmt.Sprintf("Subagent started: %s", agentInfo.AgentName),
		Meta: map[string]interface{}{
			"agent_name": agentInfo.AgentName,
			"agent_type": agentInfo.AgentType,
			"task_id":    agentInfo.TaskID,
			"hook_event": "SubagentStart",
		},
	}

	writer, err := clewcontract.NewEventWriter(sessionDir)
	if err != nil {
		printer.VerboseLog("warn", "failed to create event writer",
			map[string]interface{}{"error": err.Error()})
		return outputSubagentResult(printer, false, "clew write failed")
	}
	defer writer.Close()

	if err := writer.Write(event); err != nil {
		printer.VerboseLog("warn", "failed to write subagent start event",
			map[string]interface{}{"error": err.Error()})
		return outputSubagentResult(printer, false, "clew write failed")
	}

	return outputSubagentResult(printer, true, "")
}

// runSubagentStopCore contains the SubagentStop hook logic.
func runSubagentStopCore(ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv()

	// Verify this is a SubagentStop event (or empty for testing)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSubagentStop {
		return outputSubagentResult(printer, false, "not a SubagentStop event")
	}

	// Get session directory for clew logging
	sessionDir := getSessionDir(ctx, hookEnv)
	if sessionDir == "" {
		return outputSubagentResult(printer, false, "no active session")
	}

	// Parse agent info from tool input
	agentInfo := parseSubagentInfo(hookEnv.ToolInput)

	// Log to clew using a task_end event
	event := clewcontract.Event{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Type:      clewcontract.EventTypeTaskEnd,
		Summary:   fmt.Sprintf("Subagent stopped: %s", agentInfo.AgentName),
		Meta: map[string]interface{}{
			"agent_name": agentInfo.AgentName,
			"agent_type": agentInfo.AgentType,
			"task_id":    agentInfo.TaskID,
			"hook_event": "SubagentStop",
		},
	}

	writer, err := clewcontract.NewEventWriter(sessionDir)
	if err != nil {
		printer.VerboseLog("warn", "failed to create event writer",
			map[string]interface{}{"error": err.Error()})
		return outputSubagentResult(printer, false, "clew write failed")
	}
	defer writer.Close()

	if err := writer.Write(event); err != nil {
		printer.VerboseLog("warn", "failed to write subagent stop event",
			map[string]interface{}{"error": err.Error()})
		return outputSubagentResult(printer, false, "clew write failed")
	}

	return outputSubagentResult(printer, true, "")
}

// outputSubagentResult outputs the subagent hook result as JSON.
func outputSubagentResult(printer *output.Printer, recorded bool, reason string) error {
	return printer.Print(subagentResult{Recorded: recorded, Reason: reason})
}

// parseSubagentInfo extracts agent information from the tool input JSON.
// Returns a subagentPayload with whatever fields could be parsed.
func parseSubagentInfo(toolInputJSON string) subagentPayload {
	var info subagentPayload
	if toolInputJSON == "" {
		info.AgentName = "unknown"
		return info
	}

	// Try to parse as JSON object
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(toolInputJSON), &raw); err != nil {
		info.AgentName = "unknown"
		return info
	}

	if name, ok := raw["agent_name"].(string); ok && name != "" {
		info.AgentName = name
	} else if name, ok := raw["name"].(string); ok && name != "" {
		info.AgentName = name
	} else {
		info.AgentName = "unknown"
	}

	if agentType, ok := raw["agent_type"].(string); ok {
		info.AgentType = agentType
	} else if agentType, ok := raw["type"].(string); ok {
		info.AgentType = agentType
	}

	if taskID, ok := raw["task_id"].(string); ok {
		info.TaskID = taskID
	}

	return info
}
