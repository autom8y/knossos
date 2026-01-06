// Package threadcontract provides Thread Contract v2 event recording for Claude Code hooks.
package clewcontract

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TriggerType represents the type of auto-trigger condition.
type TriggerType string

// Auto-trigger types for detecting when /stamp should be prompted.
const (
	TriggerFileCount     TriggerType = "file_count_threshold"
	TriggerContextSwitch TriggerType = "context_switch"
	TriggerFailureRepeat TriggerType = "failure_repeat"
	TriggerSacredPath    TriggerType = "sacred_path"
)

// TriggerResult represents the outcome of trigger detection.
type TriggerResult struct {
	Triggered bool        `json:"triggered"`
	Type      TriggerType `json:"type,omitempty"`
	Reason    string      `json:"reason,omitempty"`
	Suggest   string      `json:"suggest,omitempty"` // Suggested /stamp prompt
}

// TriggerConfig contains configuration for auto-trigger detection.
type TriggerConfig struct {
	FileCountThreshold int      // default 5
	SacredPaths        []string // patterns to watch
}

// DefaultSacredPaths contains the default patterns for sacred paths.
var DefaultSacredPaths = []string{
	".claude/",
	"*_CONTEXT.md",
	"CLAUDE.md",
	"docs/decisions/",
	"docs/requirements/",
}

// DefaultTriggerConfig returns a TriggerConfig with sensible defaults.
func DefaultTriggerConfig() TriggerConfig {
	return TriggerConfig{
		FileCountThreshold: 5,
		SacredPaths:        DefaultSacredPaths,
	}
}

// CheckTriggers evaluates all trigger conditions and returns a result.
// It checks the event stream and current event against configured thresholds.
func CheckTriggers(eventsPath string, currentEvent Event, config TriggerConfig) TriggerResult {
	// Check sacred path trigger first (highest priority)
	if result := checkSacredPath(currentEvent, config.SacredPaths); result.Triggered {
		return result
	}

	// Check failure repeat pattern
	if result := checkFailureRepeat(eventsPath, currentEvent); result.Triggered {
		return result
	}

	// Check file count threshold
	if result := checkFileCount(eventsPath, currentEvent, config.FileCountThreshold); result.Triggered {
		return result
	}

	// Check context switch
	if result := checkContextSwitch(currentEvent); result.Triggered {
		return result
	}

	return TriggerResult{Triggered: false}
}

// checkSacredPath checks if the current event touches a sacred path.
func checkSacredPath(event Event, sacredPaths []string) TriggerResult {
	// Only check for write operations (Edit, Write tools)
	if event.Tool != "Edit" && event.Tool != "Write" {
		return TriggerResult{Triggered: false}
	}

	if event.Path == "" {
		return TriggerResult{Triggered: false}
	}

	for _, pattern := range sacredPaths {
		if matchSacredPattern(event.Path, pattern) {
			return TriggerResult{
				Triggered: true,
				Type:      TriggerSacredPath,
				Reason:    fmt.Sprintf("Writing to sacred path: %s", event.Path),
				Suggest:   "Consider /stamp: what change are you making to this configuration/decision?",
			}
		}
	}

	return TriggerResult{Triggered: false}
}

// matchSacredPattern checks if a path matches a sacred path pattern.
func matchSacredPattern(path, pattern string) bool {
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		// Check if path contains this directory
		return strings.Contains(path, pattern) || strings.Contains(path, strings.TrimSuffix(pattern, "/"))
	}

	// Handle wildcard prefix patterns (*_CONTEXT.md)
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	// Handle exact filename matches
	base := filepath.Base(path)
	return base == pattern || strings.Contains(path, pattern)
}

// checkFailureRepeat detects repeated failures (same error 2+ times).
func checkFailureRepeat(eventsPath string, currentEvent Event) TriggerResult {
	// Only check Bash commands with failures
	if currentEvent.Tool != "Bash" {
		return TriggerResult{Triggered: false}
	}

	// Check if current event indicates a failure
	exitCode, hasExitCode := currentEvent.Meta["exit_code"]
	if !hasExitCode {
		return TriggerResult{Triggered: false}
	}

	// Exit code 0 is success
	exitCodeFloat, ok := exitCode.(float64)
	if !ok {
		// Try int
		exitCodeInt, ok := exitCode.(int)
		if !ok || exitCodeInt == 0 {
			return TriggerResult{Triggered: false}
		}
	} else if exitCodeFloat == 0 {
		return TriggerResult{Triggered: false}
	}

	// Look for repeated failures in recent history
	if DetectRepeatedFailures(eventsPath, currentEvent) {
		command := ""
		if cmd, ok := currentEvent.Meta["command"].(string); ok {
			command = truncateString(cmd, 50)
		}
		return TriggerResult{
			Triggered: true,
			Type:      TriggerFailureRepeat,
			Reason:    fmt.Sprintf("Repeated failure detected: %s", command),
			Suggest:   "Consider /stamp: what approach are you trying? What's not working?",
		}
	}

	return TriggerResult{Triggered: false}
}

// checkFileCount checks if unique file count exceeds threshold.
func checkFileCount(eventsPath string, currentEvent Event, threshold int) TriggerResult {
	count := CountUniqueFiles(eventsPath)

	// Include current event's path if it's a write operation
	if currentEvent.Path != "" && (currentEvent.Tool == "Edit" || currentEvent.Tool == "Write") {
		// The count already includes files from events.jsonl
		// We just need to check against threshold
	}

	if count >= threshold {
		return TriggerResult{
			Triggered: true,
			Type:      TriggerFileCount,
			Reason:    fmt.Sprintf("%d files modified (threshold: %d)", count, threshold),
			Suggest:   "Consider /stamp: what approach are you taking across these files?",
		}
	}

	return TriggerResult{Triggered: false}
}

// checkContextSwitch checks if the current event represents a context switch.
func checkContextSwitch(event Event) TriggerResult {
	if event.Type != EventTypeContextSwitch {
		return TriggerResult{Triggered: false}
	}

	reason := "Context switch detected"
	if event.Summary != "" {
		reason = event.Summary
	}

	return TriggerResult{
		Triggered: true,
		Type:      TriggerContextSwitch,
		Reason:    reason,
		Suggest:   "Consider /stamp: why are you switching context?",
	}
}

// CountUniqueFiles counts unique file paths from write operations in events.jsonl.
func CountUniqueFiles(eventsPath string) int {
	events, err := readEventsFromFile(eventsPath)
	if err != nil {
		return 0
	}

	uniqueFiles := make(map[string]bool)
	for _, event := range events {
		// Only count write operations
		if event.Tool == "Edit" || event.Tool == "Write" {
			if event.Path != "" {
				uniqueFiles[event.Path] = true
			}
		}
	}

	return len(uniqueFiles)
}

// DetectRepeatedFailures checks if the current event represents a repeated failure.
// Returns true if the same type of failure has occurred before in the session.
func DetectRepeatedFailures(eventsPath string, currentEvent Event) bool {
	events, err := readEventsFromFile(eventsPath)
	if err != nil {
		return false
	}

	// Get current failure signature
	currentSig := getFailureSignature(currentEvent)
	if currentSig == "" {
		return false
	}

	// Count matching failures in history
	failureCount := 0
	for _, event := range events {
		sig := getFailureSignature(event)
		if sig == currentSig {
			failureCount++
			if failureCount >= 1 { // Current makes 2
				return true
			}
		}
	}

	return false
}

// getFailureSignature extracts a comparable signature from a failed command.
func getFailureSignature(event Event) string {
	if event.Tool != "Bash" {
		return ""
	}

	// Check for failure
	exitCode, hasExitCode := event.Meta["exit_code"]
	if !hasExitCode {
		return ""
	}

	exitCodeFloat, ok := exitCode.(float64)
	if ok && exitCodeFloat == 0 {
		return ""
	}
	exitCodeInt, ok := exitCode.(int)
	if ok && exitCodeInt == 0 {
		return ""
	}

	// Get command for signature
	command, ok := event.Meta["command"].(string)
	if !ok || command == "" {
		return ""
	}

	// Normalize command for comparison
	// Extract base command (first word) and key patterns
	return normalizeCommandForComparison(command)
}

// normalizeCommandForComparison extracts key parts of a command for comparison.
func normalizeCommandForComparison(command string) string {
	// Extract the base command (first word)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	baseCmd := parts[0]

	// For test commands, include test pattern if present
	// Check if any part of command is "test" or base command contains "test"
	for _, part := range parts {
		if part == "test" || strings.Contains(part, "test") {
			return baseCmd + ":test"
		}
	}

	// For build commands
	// Check if any part of command is "build" or base command contains "build"
	for _, part := range parts {
		if part == "build" || strings.Contains(part, "build") {
			return baseCmd + ":build"
		}
	}

	// For other commands, just use the base command
	return baseCmd
}

// readEventsFromFile reads all events from an events.jsonl file.
func readEventsFromFile(eventsPath string) ([]Event, error) {
	file, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			// Skip malformed lines
			continue
		}
		events = append(events, e)
	}

	return events, scanner.Err()
}
