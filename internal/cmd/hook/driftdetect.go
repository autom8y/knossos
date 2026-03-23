// Package hook implements the ari hook commands.
package hook

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/validation"
)

// Drift detection constants.
const (
	DriftStateFile           = ".drift-state.json"
	DriftDedupStateFile      = ".drift-dedup-state.json"
	DriftMaxRecentCalls      = 10
	DriftRetryThreshold      = 3
	DriftExplorationThreshold = 3
)

// DriftState tracks recent tool calls for pattern detection across hook invocations.
type DriftState struct {
	RecentCalls []DriftCall `json:"recent_calls"`
}

// DriftCall records a single tool invocation for drift analysis.
type DriftCall struct {
	Tool         string `json:"tool"`
	InputHash    string `json:"input_hash"`
	InputSnippet string `json:"input_snippet"`
	Success      bool   `json:"success"`
	At           string `json:"at"`
}

// DriftOutput is the JSON output for the drift detection hook.
type DriftOutput struct {
	Pattern   string `json:"pattern,omitempty"`
	Filed     bool   `json:"filed"`
	Message   string `json:"message"`
	Complaint string `json:"complaint,omitempty"`
}

// Text implements output.Textable.
func (d DriftOutput) Text() string {
	return d.Message
}

// DedupState tracks which complaint dedup keys have already been filed.
// Persisted at .sos/wip/complaints/.drift-dedup-state.json per ADR-cassandra-dedup-boundary.
type DedupState struct {
	Version int                    `json:"version"`
	Entries map[string]DedupEntry  `json:"entries"`
}

// DedupEntry records a single dedup key's filing state.
type DedupEntry struct {
	FirstFiled  string `json:"first_filed"`
	LastSeen    string `json:"last_seen"`
	Count       int    `json:"count"`
	ComplaintID string `json:"complaint_id"`
}

// loadDedupState reads the dedup state file. Returns a fresh state on any error (fail-open).
func loadDedupState(path string) DedupState {
	data, err := os.ReadFile(path)
	if err != nil {
		return DedupState{Version: 1, Entries: make(map[string]DedupEntry)}
	}
	var state DedupState
	if err := json.Unmarshal(data, &state); err != nil {
		return DedupState{Version: 1, Entries: make(map[string]DedupEntry)}
	}
	if state.Entries == nil {
		state.Entries = make(map[string]DedupEntry)
	}
	return state
}

// saveDedupState writes the dedup state file. Errors are logged but do not block.
func saveDedupState(path string, state DedupState, printer *output.Printer) {
	data, err := json.Marshal(state)
	if err != nil {
		printer.VerboseLog("warn", "driftdetect: failed to marshal dedup state",
			map[string]any{"error": err.Error()})
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		printer.VerboseLog("warn", "driftdetect: failed to write dedup state",
			map[string]any{"error": err.Error()})
	}
}

// dedupKey computes the dedup key for a complaint. Format: pattern_type:target_tool.
func dedupKey(pattern, toolFallbackTarget string) string {
	return pattern + ":" + toolFallbackTarget
}

// toolFallbackPatterns maps Bash command prefixes to their dedicated tool equivalents.
var toolFallbackPatterns = map[string]string{
	"grep ":  "Grep",
	"grep -": "Grep",
	"rg ":    "Grep",
	"cat ":   "Read",
	"head ":  "Read",
	"head -": "Read",
	"tail ":  "Read",
	"tail -": "Read",
	"find ":  "Glob",
	"find .": "Glob",
	"sed ":   "Edit",
	"sed -":  "Edit",
	"awk ":   "Edit",
}

// newDriftdetectCmd creates the drift detection hook subcommand.
func newDriftdetectCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "driftdetect",
		Short: "Detect CLI-agent interop drift and auto-file complaints",
		Long: `Detects patterns indicating CLI-agent surface drift and auto-files
complaints when patterns are identified.

This hook fires on PostToolUse events (async). It maintains a
session-scoped state file to track tool calls across invocations
and detects three drift patterns:

  1. Retry spiral: Same tool called 3+ times with similar args, all failing
  2. Command exploration: 3+ variations of ari commands in sequence
  3. Tool fallback: Agent uses Bash for what a dedicated tool handles

When a pattern is detected, a quick-file complaint is written to
.sos/wip/complaints/.

Performance: <50ms (async, no latency impact on tool execution).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runDriftdetect(cmd, ctx)
			})
		},
	}

	return cmd
}

func runDriftdetect(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runDriftdetectCore(cmd, ctx, printer, time.Now)
}

// runDriftdetectCore contains the drift detection logic with injected dependencies for testing.
func runDriftdetectCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer, nowFn func() time.Time) error {
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Accept PostToolUse and PostToolUseFailure events
	isFailure := hookEnv.Event == hook.EventPostToolFailure
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPostTool && !isFailure {
		return printer.Print(DriftOutput{Message: "not a post_tool event"})
	}

	// Must have tool information
	if hookEnv.ToolName == "" {
		return printer.Print(DriftOutput{Message: "no tool information"})
	}

	// Check for single-event tool fallback pattern with dedup gate.
	if hookEnv.ToolName == "Bash" {
		if pattern := detectToolFallback(hookEnv.ToolInput); pattern != "" {
			// Dedup gate: check if this pattern+tool was already filed.
			targetTool := detectToolFallbackTarget(hookEnv.ToolInput)
			projectDir := hookEnv.GetProjectDir()
			dedupPath := filepath.Join(projectDir, ".sos", "wip", "complaints", DriftDedupStateFile)
			dedupState := loadDedupState(dedupPath)
			key := dedupKey("tool-fallback", targetTool)
			now := nowFn()

			if entry, exists := dedupState.Entries[key]; exists {
				// Duplicate: update count and last_seen, skip filing.
				entry.Count++
				entry.LastSeen = now.Format(time.RFC3339)
				dedupState.Entries[key] = entry
				saveDedupState(dedupPath, dedupState, printer)
				return printer.Print(DriftOutput{
					Pattern: "tool-fallback",
					Filed:   false,
					Message: fmt.Sprintf("tool fallback: %s (duplicate #%d, suppressed)", pattern, entry.Count),
				})
			}

			// New key: file the complaint.
			complaintPath, err := fileDriftComplaint(hookEnv, "tool-fallback", pattern, nowFn)
			if err != nil {
				printer.VerboseLog("warn", "driftdetect: failed to file complaint",
					map[string]any{"error": err.Error()})
				return printer.Print(DriftOutput{Message: "tool fallback detected, complaint filing failed"})
			}
			validateFiledComplaint(complaintPath, printer)

			// Record in dedup state.
			complaintID := fmt.Sprintf("COMPLAINT-%s-drift-detector", now.Format("20060102-150405"))
			dedupState.Entries[key] = DedupEntry{
				FirstFiled:  now.Format(time.RFC3339),
				LastSeen:    now.Format(time.RFC3339),
				Count:       1,
				ComplaintID: complaintID,
			}
			saveDedupState(dedupPath, dedupState, printer)

			return printer.Print(DriftOutput{
				Pattern:   "tool-fallback",
				Filed:     true,
				Message:   "tool fallback: " + pattern,
				Complaint: complaintPath,
			})
		}
	}

	// Resolve session for state file path
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil || sessionID == "" || resolver.ProjectRoot() == "" {
		return printer.Print(DriftOutput{Message: "no active session for state tracking"})
	}

	// Load drift state
	sessionDir := resolver.SessionDir(sessionID)
	statePath := filepath.Join(sessionDir, DriftStateFile)
	state := loadDriftState(statePath)

	// Record current call
	snippet := extractInputSnippet(hookEnv.ToolName, hookEnv.ToolInput)
	call := DriftCall{
		Tool:         hookEnv.ToolName,
		InputHash:    hashInput(snippet),
		InputSnippet: snippet,
		Success:      !isFailure,
		At:           nowFn().Format(time.RFC3339),
	}
	state.RecentCalls = append(state.RecentCalls, call)
	if len(state.RecentCalls) > DriftMaxRecentCalls {
		state.RecentCalls = state.RecentCalls[len(state.RecentCalls)-DriftMaxRecentCalls:]
	}

	// Check for retry spiral (consecutive failures with similar input)
	if pattern := detectRetrySpiralFromState(state); pattern != "" {
		saveDriftState(statePath, state, printer)
		complaintPath, err := fileDriftComplaint(hookEnv, "retry-spiral", pattern, nowFn)
		if err != nil {
			printer.VerboseLog("warn", "driftdetect: failed to file complaint",
				map[string]any{"error": err.Error()})
		} else {
			validateFiledComplaint(complaintPath, printer)
		}
		return printer.Print(DriftOutput{
			Pattern:   "retry-spiral",
			Filed:     err == nil,
			Message:   "retry spiral: " + pattern,
			Complaint: complaintPath,
		})
	}

	// Check for command exploration (3+ ari command variations)
	if pattern := detectCommandExplorationFromState(state); pattern != "" {
		saveDriftState(statePath, state, printer)
		complaintPath, err := fileDriftComplaint(hookEnv, "command-exploration", pattern, nowFn)
		if err != nil {
			printer.VerboseLog("warn", "driftdetect: failed to file complaint",
				map[string]any{"error": err.Error()})
		} else {
			validateFiledComplaint(complaintPath, printer)
		}
		return printer.Print(DriftOutput{
			Pattern:   "command-exploration",
			Filed:     err == nil,
			Message:   "command exploration: " + pattern,
			Complaint: complaintPath,
		})
	}

	// No pattern detected — save state and exit
	saveDriftState(statePath, state, printer)
	return printer.Print(DriftOutput{Message: "no drift pattern detected"})
}

// detectToolFallback checks if a Bash command uses a tool that has a dedicated CC equivalent.
// Returns a description of the fallback, or empty string if no fallback detected.
func detectToolFallback(toolInput string) string {
	cmd := extractBashCommand(toolInput)
	if cmd == "" {
		return ""
	}

	for prefix, dedicatedTool := range toolFallbackPatterns {
		if strings.HasPrefix(cmd, prefix) {
			return fmt.Sprintf("used Bash '%s...' instead of %s tool", truncate(cmd, 60), dedicatedTool)
		}
	}
	return ""
}

// detectToolFallbackTarget returns the dedicated tool name for a tool-fallback command.
// Returns empty string if no fallback pattern matches.
func detectToolFallbackTarget(toolInput string) string {
	cmd := extractBashCommand(toolInput)
	if cmd == "" {
		return ""
	}
	for prefix, dedicatedTool := range toolFallbackPatterns {
		if strings.HasPrefix(cmd, prefix) {
			return dedicatedTool
		}
	}
	return ""
}

// detectRetrySpiralFromState checks the last N calls for consecutive failures with similar input.
func detectRetrySpiralFromState(state *DriftState) string {
	calls := state.RecentCalls
	if len(calls) < DriftRetryThreshold {
		return ""
	}

	// Check the last DriftRetryThreshold calls
	tail := calls[len(calls)-DriftRetryThreshold:]

	// All must be failures
	for _, c := range tail {
		if c.Success {
			return ""
		}
	}

	// All must be the same tool
	tool := tail[0].Tool
	for _, c := range tail[1:] {
		if c.Tool != tool {
			return ""
		}
	}

	// Input hashes should be similar (at least 2 of 3 match)
	hashCounts := make(map[string]int)
	for _, c := range tail {
		hashCounts[c.InputHash]++
	}

	// If the most common hash appears in majority of calls, it's a spiral
	maxCount := 0
	for _, count := range hashCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	if maxCount >= DriftRetryThreshold-1 {
		return fmt.Sprintf("%s failed %d times with similar input: %s",
			tool, DriftRetryThreshold, tail[0].InputSnippet)
	}

	return ""
}

// detectCommandExplorationFromState checks for 3+ ari command variations in recent Bash calls.
func detectCommandExplorationFromState(state *DriftState) string {
	calls := state.RecentCalls

	// Collect recent Bash calls with ari commands
	var ariCmds []string
	for i := len(calls) - 1; i >= 0 && len(ariCmds) < DriftExplorationThreshold+2; i-- {
		if calls[i].Tool != "Bash" {
			continue
		}
		snippet := calls[i].InputSnippet
		if strings.HasPrefix(snippet, "ari ") {
			ariCmds = append(ariCmds, snippet)
		}
	}

	if len(ariCmds) < DriftExplorationThreshold {
		return ""
	}

	// Check if the recent ari commands are variations (different commands, same intent)
	// Heuristic: 3+ different ari subcommand invocations in a short window
	uniqueCmds := make(map[string]bool)
	for _, cmd := range ariCmds[:DriftExplorationThreshold] {
		uniqueCmds[cmd] = true
	}

	if len(uniqueCmds) >= DriftExplorationThreshold {
		return fmt.Sprintf("explored %d ari command variations: %s",
			len(uniqueCmds), strings.Join(ariCmds[:DriftExplorationThreshold], " | "))
	}

	return ""
}

// complaintYAML mirrors the quick-file complaint structure for safe YAML marshaling.
// Uses yaml.v3 struct marshaling to eliminate YAML injection risk from unescaped
// Bash command snippets (fixes H1 from Cassandra health review).
type complaintYAML struct {
	ID          string   `yaml:"id"`
	FiledBy     string   `yaml:"filed_by"`
	FiledAt     string   `yaml:"filed_at"`
	Title       string   `yaml:"title"`
	Severity    string   `yaml:"severity"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Status      string   `yaml:"status"`
}

// fileDriftComplaint writes a quick-file complaint YAML to .sos/wip/complaints/.
func fileDriftComplaint(hookEnv *hook.Env, pattern, detail string, nowFn func() time.Time) (string, error) {
	now := nowFn()

	// Ensure complaints directory exists
	projectDir := hookEnv.GetProjectDir()
	complaintsDir := filepath.Join(projectDir, ".sos", "wip", "complaints")
	if err := os.MkdirAll(complaintsDir, 0755); err != nil {
		return "", fmt.Errorf("create complaints dir: %w", err)
	}

	timestamp := now.Format("20060102-150405")
	id := fmt.Sprintf("COMPLAINT-%s-drift-detector", timestamp)
	filename := id + ".yaml"
	path := filepath.Join(complaintsDir, filename)

	severity := "medium"
	if pattern == "tool-fallback" {
		severity = "low"
	}

	c := complaintYAML{
		ID:       id,
		FiledBy:  "drift-detector",
		FiledAt:  now.Format(time.RFC3339),
		Title:    fmt.Sprintf("%s drift: %s", pattern, truncate(detail, 120)),
		Severity: severity,
		Description: fmt.Sprintf("Drift detection hook identified a %s pattern.\n%s\nTool: %s",
			pattern, detail, hookEnv.ToolName),
		Tags:   []string{"drift", pattern, "auto-filed"},
		Status: "filed",
	}

	data, err := yaml.Marshal(&c)
	if err != nil {
		return "", fmt.Errorf("marshal complaint: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write complaint: %w", err)
	}

	return path, nil
}

// validateFiledComplaint performs non-blocking schema validation on a filed complaint.
// Logs a warning if the complaint fails validation but does not prevent filing.
func validateFiledComplaint(path string, printer *output.Printer) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // best-effort: skip validation if read fails
	}
	v, err := validation.NewValidator()
	if err != nil {
		return // best-effort: skip if validator unavailable
	}
	if err := v.ValidateComplaint(data); err != nil {
		printer.VerboseLog("warn", "driftdetect: filed complaint failed schema validation",
			map[string]any{"path": path, "error": err.Error()})
	}
}

// extractBashCommand extracts the command string from Bash tool input JSON.
func extractBashCommand(toolInput string) string {
	if toolInput == "" {
		return ""
	}
	var input map[string]any
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		return ""
	}
	if cmd, ok := input["command"].(string); ok {
		return strings.TrimSpace(cmd)
	}
	return ""
}

// extractInputSnippet creates a short, hashable representation of a tool call's input.
func extractInputSnippet(toolName, toolInput string) string {
	if toolInput == "" {
		return toolName
	}

	var input map[string]any
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		return toolName
	}

	switch toolName {
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			return truncate(strings.TrimSpace(cmd), 100)
		}
	case "Read":
		if fp, ok := input["file_path"].(string); ok {
			return "read:" + fp
		}
	case "Write", "Edit":
		if fp, ok := input["file_path"].(string); ok {
			return toolName + ":" + fp
		}
	case "Grep":
		if pattern, ok := input["pattern"].(string); ok {
			return "grep:" + pattern
		}
	case "Glob":
		if pattern, ok := input["pattern"].(string); ok {
			return "glob:" + pattern
		}
	}

	return toolName
}

// hashInput produces a short hash for dedup comparison of tool inputs.
func hashInput(snippet string) string {
	h := sha256.Sum256([]byte(snippet))
	return fmt.Sprintf("%x", h[:4])
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// loadDriftState reads the drift state file from the session directory.
// Returns an empty state if the file doesn't exist or can't be parsed.
func loadDriftState(path string) *DriftState {
	data, err := os.ReadFile(path)
	if err != nil {
		return &DriftState{}
	}
	var state DriftState
	if err := json.Unmarshal(data, &state); err != nil {
		return &DriftState{}
	}
	return &state
}

// saveDriftState writes the drift state to the session directory.
// Best-effort: errors are logged but not fatal (async hook, no blocking).
func saveDriftState(path string, state *DriftState, printer *output.Printer) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		printer.VerboseLog("warn", "driftdetect: failed to save drift state",
			map[string]any{"path": path, "error": err.Error()})
	}
}
