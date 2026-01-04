package threadcontract

import (
	"regexp"
	"strings"
)

// Throughline represents the decision summary from orchestrator CONSULTATION_RESPONSE.
type Throughline struct {
	Decision  string
	Rationale string
}

// ExtractThroughline parses YAML throughline from orchestrator Task tool result.
// Returns nil if no throughline is found (not an error - just not an orchestrator response).
//
// Expected YAML format:
//
//	throughline:
//	  decision: "What was decided (one sentence)"
//	  rationale: "Why this decision"
//
// The function performs lightweight regex extraction to avoid full YAML parsing overhead.
func ExtractThroughline(toolResult string) *Throughline {
	if toolResult == "" {
		return nil
	}

	// Check if this looks like an orchestrator response (has "throughline:" key)
	if !strings.Contains(toolResult, "throughline:") {
		return nil
	}

	// Extract decision field using regex
	// Match: decision: "text" or decision: text (with or without quotes)
	decisionPattern := regexp.MustCompile(`(?m)^\s*decision:\s*(?:"([^"]+)"|(.+))$`)
	decisionMatch := decisionPattern.FindStringSubmatch(toolResult)

	var decision string
	if len(decisionMatch) > 1 {
		// Use quoted match (group 1) if present, otherwise unquoted (group 2)
		if decisionMatch[1] != "" {
			decision = decisionMatch[1]
		} else if decisionMatch[2] != "" {
			decision = strings.TrimSpace(decisionMatch[2])
		}
	}

	// Extract rationale field using regex
	// Match: rationale: "text" or rationale: text (with or without quotes)
	rationalePattern := regexp.MustCompile(`(?m)^\s*rationale:\s*(?:"([^"]+)"|(.+))$`)
	rationaleMatch := rationalePattern.FindStringSubmatch(toolResult)

	var rationale string
	if len(rationaleMatch) > 1 {
		// Use quoted match (group 1) if present, otherwise unquoted (group 2)
		if rationaleMatch[1] != "" {
			rationale = rationaleMatch[1]
		} else if rationaleMatch[2] != "" {
			rationale = strings.TrimSpace(rationaleMatch[2])
		}
	}

	// Only return throughline if we found at least decision
	if decision == "" {
		return nil
	}

	return &Throughline{
		Decision:  decision,
		Rationale: rationale,
	}
}

// IsOrchestratorAgent checks if the tool result appears to be from an orchestrator agent.
// This is a heuristic check looking for CONSULTATION_RESPONSE markers.
func IsOrchestratorAgent(toolResult string) bool {
	if toolResult == "" {
		return false
	}

	// Look for key orchestrator response patterns
	return strings.Contains(toolResult, "throughline:") ||
		strings.Contains(toolResult, "directive:") ||
		strings.Contains(toolResult, "CONSULTATION_RESPONSE")
}
