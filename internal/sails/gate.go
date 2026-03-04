// Package sails implements the White Sails confidence signaling system per Knossos Doctrine v2.
// This file implements the quality gate check function per TDD Section 7.
package sails

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"gopkg.in/yaml.v3"
)

// GateResult represents the result of a quality gate check.
// Pass criteria: WHITE color only (GRAY and BLACK fail the gate).
type GateResult struct {
	// Pass indicates whether the gate check passed.
	// True only if color is WHITE.
	Pass bool `json:"pass" yaml:"pass"`

	// Color is the sails color from WHITE_SAILS.yaml.
	Color Color `json:"color" yaml:"color"`

	// SessionID is the session identifier from WHITE_SAILS.yaml.
	SessionID string `json:"session_id" yaml:"session_id"`

	// Reasons explains why the gate passed or failed.
	Reasons []string `json:"reasons" yaml:"reasons"`

	// FilePath is the path to the WHITE_SAILS.yaml file.
	FilePath string `json:"file_path" yaml:"file_path"`

	// ComputedBase is the computed color before modifiers.
	ComputedBase Color `json:"computed_base,omitempty" yaml:"computed_base,omitempty"`

	// OpenQuestions from the sails file (if any).
	OpenQuestions []string `json:"open_questions,omitempty" yaml:"open_questions,omitempty"`

	// ContractViolations lists thread contract violations found (if any).
	ContractViolations []ContractViolation `json:"contract_violations,omitempty" yaml:"contract_violations,omitempty"`
}

// CheckGate reads WHITE_SAILS.yaml from a session and returns the gate result.
// Pass criteria: WHITE color only (GRAY and BLACK fail the gate).
//
// The sessionPath can be:
//   - A session directory containing WHITE_SAILS.yaml
//   - A direct path to WHITE_SAILS.yaml
//
// This function also validates the thread contract from events.jsonl.
// Thread contract violations degrade sails to GRAY at minimum.
func CheckGate(sessionPath string) (*GateResult, error) {
	if sessionPath == "" {
		return nil, errors.New(errors.CodeUsageError, "session path is required")
	}

	// Determine the WHITE_SAILS.yaml path and session directory
	sailsPath := sessionPath
	sessionDir := sessionPath
	info, err := os.Stat(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"path not found",
				map[string]any{"path": sessionPath})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to access path", err)
	}

	// If it's a directory, look for WHITE_SAILS.yaml inside
	if info.IsDir() {
		sailsPath = filepath.Join(sessionPath, "WHITE_SAILS.yaml")
		if _, err := os.Stat(sailsPath); err != nil {
			if os.IsNotExist(err) {
				return nil, errors.NewWithDetails(errors.CodeFileNotFound,
					"WHITE_SAILS.yaml not found in session directory",
					map[string]any{"session_path": sessionPath})
			}
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to access WHITE_SAILS.yaml", err)
		}
	} else {
		// If sailsPath is a file, extract the directory
		sessionDir = filepath.Dir(sailsPath)
	}

	// Read the WHITE_SAILS.yaml file
	content, err := os.ReadFile(sailsPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read WHITE_SAILS.yaml", err)
	}

	// Parse the YAML
	var sails WhiteSailsYAML
	if err := yaml.Unmarshal(content, &sails); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse WHITE_SAILS.yaml", err)
	}

	// Validate clew contract from events.jsonl
	violations, err := ValidateClewContract(sessionDir)
	if err != nil {
		// Log the error but don't fail the gate check
		// Clew contract validation is best-effort
		violations = nil
	}

	// Build the gate result
	color := Color(sails.Color)
	computedBase := Color(sails.ComputedBase)

	// Apply clew contract violations to color
	if len(violations) > 0 {
		// Clew contract violations degrade to GRAY minimum
		if color == ColorWhite {
			color = ColorGray
		}
	}

	result := &GateResult{
		Pass:               color == ColorWhite,
		Color:              color,
		SessionID:          sails.SessionID,
		FilePath:           sailsPath,
		ComputedBase:       computedBase,
		OpenQuestions:      sails.OpenQuestions,
		ContractViolations: violations,
	}

	// Build reasons based on the color
	switch Color(sails.Color) { // Use original color from YAML
	case ColorWhite:
		result.Reasons = append(result.Reasons, "sails color is WHITE: high confidence, ship without QA")
	case ColorGray:
		result.Reasons = append(result.Reasons, "sails color is GRAY: unknown confidence, needs QA review")
		if len(sails.OpenQuestions) > 0 {
			result.Reasons = append(result.Reasons, "open questions present")
		}
		if sails.Type == "spike" {
			result.Reasons = append(result.Reasons, "session type is spike (gray ceiling)")
		}
		if sails.Type == "hotfix" {
			result.Reasons = append(result.Reasons, "session type is hotfix (expedited gray)")
		}
	case ColorBlack:
		result.Reasons = append(result.Reasons, "sails color is BLACK: known failure, do not ship")
	default:
		result.Reasons = append(result.Reasons, "unknown sails color: "+string(sails.Color))
	}

	// Add reason for clew contract violations
	if len(violations) > 0 {
		if Color(sails.Color) == ColorWhite {
			result.Reasons = append(result.Reasons, "clew contract violations present: downgraded to GRAY")
		}
		result.Reasons = append(result.Reasons, "clew contract has violations (see contract_violations)")
	}

	return result, nil
}

// CheckGateForSession checks the gate for a specific session.
func CheckGateForSession(projectRoot string, sessionID string) (*GateResult, error) {
	if projectRoot == "" {
		return nil, errors.New(errors.CodeUsageError, "project root is required")
	}
	if sessionID == "" {
		return nil, errors.New(errors.CodeSessionNotFound, "no session ID provided")
	}

	resolver := paths.NewResolver(projectRoot)
	sessionDir := resolver.SessionDir(strings.TrimSpace(sessionID))
	return CheckGate(sessionDir)
}

// trimWhitespace removes leading and trailing whitespace from a string.
func trimWhitespace(s string) string {
	start := 0
	end := len(s)
	for start < end && isWhitespace(s[start]) {
		start++
	}
	for end > start && isWhitespace(s[end-1]) {
		end--
	}
	return s[start:end]
}

// isWhitespace returns true if the byte is a whitespace character.
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// GateExitCode returns the appropriate exit code for a gate result.
// Exit code 0 for WHITE (pass), non-zero for GRAY/BLACK (fail).
func GateExitCode(result *GateResult) int {
	if result == nil {
		return errors.ExitGeneralError
	}
	if result.Pass {
		return errors.ExitSuccess
	}
	// Use validation failed exit code for gate failures
	return errors.ExitValidationFailed
}
