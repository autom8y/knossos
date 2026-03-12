// Package hook implements the ari hook commands.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/registry"
	"github.com/autom8y/knossos/internal/suggest"
)

// subagentResult is the output of subagent hooks.
// SubagentStart/SubagentStop are side-effect hooks (logging only) -- they cannot block.
type subagentResult struct {
	Recorded    bool                   `json:"recorded"`
	Reason      string                 `json:"reason,omitempty"`
	Suggestions []suggest.Suggestion   `json:"suggestions,omitempty"` // H5: next-step suggestions
}

// subagentPayload represents the relevant fields from the stdin JSON payload
// for SubagentStart and SubagentStop events.
type subagentPayload struct {
	AgentName string `json:"agent_name"`
	AgentType string `json:"agent_type"`
	TaskID    string `json:"task_id"`
	AgentID   string `json:"agent_id"`
}

// throughlineAgentNames is the set of agent names tracked as throughline agents.
// These are long-running orchestrator agents whose IDs must survive compaction.
var throughlineAgentNames = registry.ThroughlineAgents()

// ThroughlineIDsFile is the filename for the session-scoped throughline agent ID map.
const ThroughlineIDsFile = ".throughline-ids.json"

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
				return runSubagentStart(cmd, ctx)
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
				return runSubagentStop(cmd, ctx)
			})
		},
	}

	return cmd
}

func runSubagentStart(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runSubagentStartCore(cmd, ctx, printer)
}

func runSubagentStop(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runSubagentStopCore(cmd, ctx, printer)
}

// runSubagentStartCore contains the SubagentStart hook logic.
func runSubagentStartCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Verify this is a SubagentStart event (or empty for testing)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSubagentStart {
		return outputSubagentResult(printer, false, "not a subagent_start event")
	}

	// Get session directory for clew logging
	sessionDir := getSessionDir(ctx, hookEnv)
	if sessionDir == "" {
		return outputSubagentResult(printer, false, "no active session")
	}

	// Parse agent info from tool input (CC sends agent details in tool_input)
	agentInfo := parseSubagentInfo(hookEnv.ToolInput)

	// Log to clew using a task_start event
	meta := map[string]any{
		"agent_name": agentInfo.AgentName,
		"agent_type": agentInfo.AgentType,
		"task_id":    agentInfo.TaskID,
		"hook_event": "SubagentStart",
	}
	if agentInfo.AgentID != "" {
		meta["agent_id"] = agentInfo.AgentID
	}

	event := clewcontract.Event{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Type:      clewcontract.EventTypeTaskStart,
		Summary:   fmt.Sprintf("Subagent started: %s", agentInfo.AgentName),
		Meta:      meta,
	}

	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = writer.Close() }()

	writer.Write(event)
	if flushErr := writer.Flush(); flushErr != nil {
		printer.VerboseLog("warn", "failed to write subagent start event",
			map[string]any{"error": flushErr.Error()})
		return outputSubagentResult(printer, false, "clew write failed")
	}

	// Persist throughline agent ID if this agent is a throughline agent.
	// Best-effort: failures are logged but never block the hook.
	if agentInfo.AgentID != "" && throughlineAgentNames[agentInfo.AgentName] {
		if err := upsertThroughlineID(sessionDir, agentInfo.AgentName, agentInfo.AgentID); err != nil {
			printer.VerboseLog("warn", "failed to persist throughline agent ID",
				map[string]any{
					"agent_name": agentInfo.AgentName,
					"error":      err.Error(),
				})
		}
	}

	return outputSubagentResult(printer, true, "")
}

// runSubagentStopCore contains the SubagentStop hook logic.
func runSubagentStopCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Verify this is a SubagentStop event (or empty for testing)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSubagentStop {
		return outputSubagentResult(printer, false, "not a subagent_stop event")
	}

	// Get session directory for clew logging
	sessionDir := getSessionDir(ctx, hookEnv)
	if sessionDir == "" {
		return outputSubagentResult(printer, false, "no active session")
	}

	// Parse agent info from tool input
	agentInfo := parseSubagentInfo(hookEnv.ToolInput)

	// Log to clew using a task_end event
	stopMeta := map[string]any{
		"agent_name": agentInfo.AgentName,
		"agent_type": agentInfo.AgentType,
		"task_id":    agentInfo.TaskID,
		"hook_event": "SubagentStop",
	}
	if agentInfo.AgentID != "" {
		stopMeta["agent_id"] = agentInfo.AgentID
	}

	event := clewcontract.Event{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Type:      clewcontract.EventTypeTaskEnd,
		Summary:   fmt.Sprintf("Subagent stopped: %s", agentInfo.AgentName),
		Meta:      stopMeta,
	}

	stopWriter := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = stopWriter.Close() }()

	stopWriter.Write(event)
	if flushErr := stopWriter.Flush(); flushErr != nil {
		printer.VerboseLog("warn", "failed to write subagent stop event",
			map[string]any{"error": flushErr.Error()})
		return outputSubagentResult(printer, false, "clew write failed")
	}

	// H5: Generate subagent completion suggestions (fail-open, <2ms pure function)
	suggestInput := &suggest.SubagentInput{
		AgentName: agentInfo.AgentName,
		AgentType: agentInfo.AgentType,
	}
	suggestions := suggest.SubagentStopSuggestions(suggestInput)

	result := subagentResult{Recorded: true, Suggestions: suggestions}
	return printer.Print(result)
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
	var raw map[string]any
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

	// Extract agent_id with fallback to "id" field, matching the same fallback
	// pattern used for agent_name above.
	if agentID, ok := raw["agent_id"].(string); ok && agentID != "" {
		info.AgentID = agentID
	} else if agentID, ok := raw["id"].(string); ok && agentID != "" {
		info.AgentID = agentID
	}

	return info
}

// upsertThroughlineID atomically writes the agent_id for a throughline agent
// into the session-scoped .throughline-ids.json file.
// It reads the existing file (if any), updates the key, and writes back via
// a temp-file rename for atomicity.
func upsertThroughlineID(sessionDir, agentName, agentID string) error {
	idFile := filepath.Join(sessionDir, ThroughlineIDsFile)

	// Read existing IDs (if any).
	ids := make(map[string]string)
	if data, err := os.ReadFile(idFile); err == nil {
		// Best-effort unmarshal — corrupt file is treated as empty.
		_ = json.Unmarshal(data, &ids)
	}

	// Update the key for this agent.
	ids[agentName] = agentID

	// Marshal to JSON.
	data, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("marshal throughline IDs: %w", err)
	}

	// Write to a temp file in the same directory, then rename for atomicity.
	tmpFile := idFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write throughline IDs temp file: %w", err)
	}
	if err := os.Rename(tmpFile, idFile); err != nil {
		return fmt.Errorf("rename throughline IDs file: %w", err)
	}

	return nil
}

// readThroughlineIDs reads the .throughline-ids.json file from sessionDir.
// Returns an empty map if the file does not exist or cannot be parsed.
func readThroughlineIDs(sessionDir string) map[string]string {
	idFile := filepath.Join(sessionDir, ThroughlineIDsFile)
	data, err := os.ReadFile(idFile)
	if err != nil {
		return nil
	}
	var ids map[string]string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil
	}
	return ids
}
