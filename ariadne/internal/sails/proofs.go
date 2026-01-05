// Package sails implements the White Sails confidence signaling system for Ariadne.
// It provides proof collection, color computation, and confidence attestation per Knossos Doctrine v2.
package sails

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// ProofItem represents a single proof artifact from session validation.
// Note: ProofStatus constants (ProofPass, ProofFail, ProofSkip, ProofUnknown)
// are defined in color.go to avoid duplication.
type ProofItem struct {
	Status       ProofStatus `yaml:"status" json:"status"`
	EvidencePath string      `yaml:"evidence_path,omitempty" json:"evidence_path,omitempty"`
	Summary      string      `yaml:"summary,omitempty" json:"summary,omitempty"`
	ExitCode     int         `yaml:"exit_code,omitempty" json:"exit_code,omitempty"`
	Timestamp    time.Time   `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
}

// ProofSet contains all proof items collected from a session.
type ProofSet struct {
	Tests       *ProofItem `yaml:"tests,omitempty" json:"tests,omitempty"`
	Build       *ProofItem `yaml:"build,omitempty" json:"build,omitempty"`
	Lint        *ProofItem `yaml:"lint,omitempty" json:"lint,omitempty"`
	Adversarial *ProofItem `yaml:"adversarial,omitempty" json:"adversarial,omitempty"`
	Integration *ProofItem `yaml:"integration,omitempty" json:"integration,omitempty"`
}

// ProofType identifies the kind of proof being collected.
type ProofType string

const (
	ProofTypeTest        ProofType = "test"
	ProofTypeBuild       ProofType = "build"
	ProofTypeLint        ProofType = "lint"
	ProofTypeAdversarial ProofType = "adversarial"
	ProofTypeIntegration ProofType = "integration"
)

// proofLogFiles maps proof types to their expected log file names.
var proofLogFiles = map[ProofType]string{
	ProofTypeTest:        "test-output.log",
	ProofTypeBuild:       "build-output.log",
	ProofTypeLint:        "lint-output.log",
	ProofTypeAdversarial: "adversarial-output.log",
	ProofTypeIntegration: "integration-output.log",
}

// Exit code detection patterns.
var (
	// exitCodeExplicit matches "exit code: N" or "exit code N"
	exitCodeExplicit = regexp.MustCompile(`(?i)exit\s+code[:\s]+(\d+)`)
	// exitCodeExited matches "exited with N" or "exited N"
	exitCodeExited = regexp.MustCompile(`(?i)exited\s+(?:with\s+)?(\d+)`)
	// exitCodeStatus matches trailing "exit status N"
	exitCodeStatus = regexp.MustCompile(`(?i)exit\s+status\s+(\d+)`)
)

// Test summary detection patterns.
var (
	// goTestSummary matches Go test output like "ok  pkg 0.123s" or "FAIL pkg 0.123s"
	goTestOK   = regexp.MustCompile(`(?m)^ok\s+\S+`)
	goTestFail = regexp.MustCompile(`(?m)^FAIL\s+\S+`)

	// genericTestCounts matches "N tests passed", "N passed", "N failed", "N skipped"
	testPassedCount  = regexp.MustCompile(`(?i)(\d+)\s+(?:tests?\s+)?passed`)
	testFailedCount  = regexp.MustCompile(`(?i)(\d+)\s+(?:tests?\s+)?failed`)
	testSkippedCount = regexp.MustCompile(`(?i)(\d+)\s+(?:tests?\s+)?skipped`)

	// pytestSummary matches pytest output like "5 passed, 2 failed, 1 skipped"
	pytestSummary = regexp.MustCompile(`(?i)(\d+)\s+passed(?:,\s*(\d+)\s+failed)?(?:,\s*(\d+)\s+skipped)?`)

	// jestSummary matches Jest output like "Tests: 5 passed, 2 failed, 7 total"
	jestSummary = regexp.MustCompile(`(?i)Tests?:\s*(\d+)\s+passed(?:,\s*(\d+)\s+failed)?`)
)

// CollectProofs reads proof artifacts from a session directory.
// It looks for test-output.log, build-output.log, lint-output.log,
// and optional adversarial-output.log, integration-output.log files.
func CollectProofs(sessionDir string) (*ProofSet, error) {
	if sessionDir == "" {
		return nil, errors.New(errors.CodeUsageError, "session directory is required")
	}

	// Verify session directory exists
	info, err := os.Stat(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeSessionNotFound,
				"session directory not found",
				map[string]interface{}{"path": sessionDir})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to access session directory", err)
	}
	if !info.IsDir() {
		return nil, errors.NewWithDetails(errors.CodeUsageError,
			"path is not a directory",
			map[string]interface{}{"path": sessionDir})
	}

	proofSet := &ProofSet{}

	// Collect required proofs
	proofSet.Tests = parseProofFromDir(sessionDir, ProofTypeTest)
	proofSet.Build = parseProofFromDir(sessionDir, ProofTypeBuild)
	proofSet.Lint = parseProofFromDir(sessionDir, ProofTypeLint)

	// Collect optional proofs
	proofSet.Adversarial = parseProofFromDir(sessionDir, ProofTypeAdversarial)
	proofSet.Integration = parseProofFromDir(sessionDir, ProofTypeIntegration)

	return proofSet, nil
}

// parseProofFromDir reads and parses a proof log from a session directory.
func parseProofFromDir(sessionDir string, proofType ProofType) *ProofItem {
	logFileName, ok := proofLogFiles[proofType]
	if !ok {
		return &ProofItem{
			Status:    ProofUnknown,
			Summary:   "unknown proof type",
			Timestamp: time.Now().UTC(),
		}
	}

	logPath := filepath.Join(sessionDir, logFileName)
	proof, _ := ParseProofLog(logPath, string(proofType))
	return proof
}

// ParseProofLog parses a log file to extract proof status.
// Returns a ProofItem with UNKNOWN status if the file is missing or empty.
func ParseProofLog(path string, proofType string) (*ProofItem, error) {
	now := time.Now().UTC()

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProofItem{
				Status:    ProofUnknown,
				Summary:   "log file not found",
				Timestamp: now,
			}, nil
		}
		return &ProofItem{
			Status:    ProofUnknown,
			Summary:   "failed to access log file: " + err.Error(),
			Timestamp: now,
		}, errors.Wrap(errors.CodeGeneralError, "failed to access log file", err)
	}

	// Check for empty file
	if info.Size() == 0 {
		return &ProofItem{
			Status:       ProofUnknown,
			EvidencePath: path,
			Summary:      "log file is empty",
			Timestamp:    now,
		}, nil
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return &ProofItem{
			Status:    ProofUnknown,
			Summary:   "failed to read log file: " + err.Error(),
			Timestamp: now,
		}, errors.Wrap(errors.CodeGeneralError, "failed to read log file", err)
	}

	// Use file modification time as timestamp
	timestamp := info.ModTime().UTC()

	// Parse based on proof type
	return parseLogContent(string(content), path, proofType, timestamp), nil
}

// parseLogContent analyzes log content to determine proof status.
func parseLogContent(content, path, proofType string, timestamp time.Time) *ProofItem {
	proof := &ProofItem{
		EvidencePath: path,
		Timestamp:    timestamp,
	}

	// Detect exit code
	exitCode, hasExitCode := DetectExitCode(content)
	if hasExitCode {
		proof.ExitCode = exitCode
	}

	// Handle based on proof type
	switch proofType {
	case string(ProofTypeTest), "tests":
		proof.Status, proof.Summary = analyzeTestOutput(content, exitCode, hasExitCode)
	case string(ProofTypeBuild):
		proof.Status, proof.Summary = analyzeBuildOutput(content, exitCode, hasExitCode)
	case string(ProofTypeLint):
		proof.Status, proof.Summary = analyzeLintOutput(content, exitCode, hasExitCode)
	case string(ProofTypeAdversarial):
		proof.Status, proof.Summary = analyzeTestOutput(content, exitCode, hasExitCode)
	case string(ProofTypeIntegration):
		proof.Status, proof.Summary = analyzeTestOutput(content, exitCode, hasExitCode)
	default:
		proof.Status, proof.Summary = analyzeGenericOutput(content, exitCode, hasExitCode)
	}

	return proof
}

// DetectExitCode extracts exit code from log content.
// Returns the exit code and true if found, or (0, false) if not found.
func DetectExitCode(content string) (int, bool) {
	// Try patterns in order of specificity

	// "exit code: N" or "exit code N"
	if matches := exitCodeExplicit.FindStringSubmatch(content); len(matches) > 1 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			return code, true
		}
	}

	// "exited with N" or "exited N"
	if matches := exitCodeExited.FindStringSubmatch(content); len(matches) > 1 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			return code, true
		}
	}

	// "exit status N" (often at end of output)
	if matches := exitCodeStatus.FindStringSubmatch(content); len(matches) > 1 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			return code, true
		}
	}

	return 0, false
}

// DetectTestSummary extracts test pass/fail/skip counts from output.
// Returns (passed, failed, skipped, ok) where ok is true if any counts were detected.
func DetectTestSummary(content string) (passed, failed, skipped int, ok bool) {
	// Try pytest-style summary first: "5 passed, 2 failed, 1 skipped"
	if matches := pytestSummary.FindStringSubmatch(content); len(matches) > 1 {
		passed = parseIntOrZero(matches[1])
		if len(matches) > 2 {
			failed = parseIntOrZero(matches[2])
		}
		if len(matches) > 3 {
			skipped = parseIntOrZero(matches[3])
		}
		return passed, failed, skipped, true
	}

	// Try Jest-style summary: "Tests: 5 passed, 2 failed"
	if matches := jestSummary.FindStringSubmatch(content); len(matches) > 1 {
		passed = parseIntOrZero(matches[1])
		if len(matches) > 2 {
			failed = parseIntOrZero(matches[2])
		}
		return passed, failed, skipped, true
	}

	// Try generic patterns
	if matches := testPassedCount.FindStringSubmatch(content); len(matches) > 1 {
		passed = parseIntOrZero(matches[1])
		ok = true
	}
	if matches := testFailedCount.FindStringSubmatch(content); len(matches) > 1 {
		failed = parseIntOrZero(matches[1])
		ok = true
	}
	if matches := testSkippedCount.FindStringSubmatch(content); len(matches) > 1 {
		skipped = parseIntOrZero(matches[1])
		ok = true
	}

	// Try Go test output: count "ok" and "FAIL" lines
	if !ok {
		okCount := len(goTestOK.FindAllString(content, -1))
		failCount := len(goTestFail.FindAllString(content, -1))
		if okCount > 0 || failCount > 0 {
			passed = okCount
			failed = failCount
			ok = true
		}
	}

	return passed, failed, skipped, ok
}

// analyzeTestOutput determines test proof status from log content.
func analyzeTestOutput(content string, exitCode int, hasExitCode bool) (ProofStatus, string) {
	passed, failed, skipped, hasSummary := DetectTestSummary(content)

	// If we have test summary, use it
	if hasSummary {
		var summary strings.Builder
		if passed > 0 {
			summary.WriteString(strconv.Itoa(passed))
			summary.WriteString(" passed")
		}
		if failed > 0 {
			if summary.Len() > 0 {
				summary.WriteString(", ")
			}
			summary.WriteString(strconv.Itoa(failed))
			summary.WriteString(" failed")
		}
		if skipped > 0 {
			if summary.Len() > 0 {
				summary.WriteString(", ")
			}
			summary.WriteString(strconv.Itoa(skipped))
			summary.WriteString(" skipped")
		}

		if failed > 0 {
			return ProofFail, summary.String()
		}
		if passed > 0 {
			return ProofPass, summary.String()
		}
		if skipped > 0 {
			return ProofSkip, summary.String()
		}
	}

	// Fall back to exit code
	if hasExitCode {
		if exitCode == 0 {
			return ProofPass, "tests passed (exit code 0)"
		}
		return ProofFail, "tests failed (exit code " + strconv.Itoa(exitCode) + ")"
	}

	// Check for explicit PASS/FAIL markers
	upperContent := strings.ToUpper(content)
	if strings.Contains(upperContent, "PASS") && !strings.Contains(upperContent, "FAIL") {
		return ProofPass, "tests passed"
	}
	if strings.Contains(upperContent, "FAIL") {
		return ProofFail, "tests failed"
	}

	return ProofUnknown, "could not determine test outcome"
}

// analyzeBuildOutput determines build proof status from log content.
func analyzeBuildOutput(content string, exitCode int, hasExitCode bool) (ProofStatus, string) {
	// Exit code is the primary indicator for builds
	if hasExitCode {
		if exitCode == 0 {
			return ProofPass, "build succeeded"
		}
		return ProofFail, "build failed (exit code " + strconv.Itoa(exitCode) + ")"
	}

	// Look for common build success/failure patterns
	lowerContent := strings.ToLower(content)

	// Check for explicit errors
	if strings.Contains(lowerContent, "error:") ||
		strings.Contains(lowerContent, "fatal error") ||
		strings.Contains(lowerContent, "build failed") ||
		strings.Contains(lowerContent, "compilation failed") {
		return ProofFail, "build errors detected"
	}

	// Check for success patterns
	if strings.Contains(lowerContent, "build successful") ||
		strings.Contains(lowerContent, "build succeeded") ||
		strings.Contains(lowerContent, "compilation successful") {
		return ProofPass, "build succeeded"
	}

	// No clear indicators
	return ProofUnknown, "could not determine build outcome"
}

// analyzeLintOutput determines lint proof status from log content.
func analyzeLintOutput(content string, exitCode int, hasExitCode bool) (ProofStatus, string) {
	// Exit code is primary indicator
	if hasExitCode {
		if exitCode == 0 {
			return ProofPass, "lint clean"
		}
		return ProofFail, "lint issues found (exit code " + strconv.Itoa(exitCode) + ")"
	}

	// Look for warning/error counts
	lowerContent := strings.ToLower(content)

	// Check for explicit issues
	if strings.Contains(lowerContent, "error") ||
		strings.Contains(lowerContent, "warning") {
		// Count occurrences
		errorCount := strings.Count(lowerContent, "error")
		warningCount := strings.Count(lowerContent, "warning")
		if errorCount > 0 || warningCount > 0 {
			var summary strings.Builder
			if errorCount > 0 {
				summary.WriteString(strconv.Itoa(errorCount))
				summary.WriteString(" error(s)")
			}
			if warningCount > 0 {
				if summary.Len() > 0 {
					summary.WriteString(", ")
				}
				summary.WriteString(strconv.Itoa(warningCount))
				summary.WriteString(" warning(s)")
			}
			return ProofFail, summary.String()
		}
	}

	// Check for clean patterns
	if strings.Contains(lowerContent, "no issues") ||
		strings.Contains(lowerContent, "no problems") ||
		strings.Contains(lowerContent, "clean") {
		return ProofPass, "lint clean"
	}

	return ProofUnknown, "could not determine lint outcome"
}

// analyzeGenericOutput analyzes output without specific type hints.
func analyzeGenericOutput(content string, exitCode int, hasExitCode bool) (ProofStatus, string) {
	if hasExitCode {
		if exitCode == 0 {
			return ProofPass, "completed successfully"
		}
		return ProofFail, "failed (exit code " + strconv.Itoa(exitCode) + ")"
	}

	return ProofUnknown, "could not determine outcome"
}

// parseIntOrZero safely parses a string to int, returning 0 on failure.
func parseIntOrZero(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// ProofCollector provides session-aware proof collection.
type ProofCollector struct {
	projectRoot string
	paths       *paths.Resolver
}

// NewProofCollector creates a new ProofCollector for the given project root.
func NewProofCollector(projectRoot string) *ProofCollector {
	return &ProofCollector{
		projectRoot: projectRoot,
		paths:       paths.NewResolver(projectRoot),
	}
}

// CollectForSession collects proofs from a specific session.
func (c *ProofCollector) CollectForSession(sessionID string) (*ProofSet, error) {
	sessionDir := c.paths.SessionDir(sessionID)
	return CollectProofs(sessionDir)
}

// HasRequiredProofs checks if a ProofSet has the minimum required proofs.
// Required proofs are: tests, build, lint (all must not be nil and not UNKNOWN).
func (ps *ProofSet) HasRequiredProofs() bool {
	if ps.Tests == nil || ps.Tests.Status == ProofUnknown {
		return false
	}
	if ps.Build == nil || ps.Build.Status == ProofUnknown {
		return false
	}
	if ps.Lint == nil || ps.Lint.Status == ProofUnknown {
		return false
	}
	return true
}

// AllPass returns true if all present proofs have PASS status.
func (ps *ProofSet) AllPass() bool {
	if ps.Tests != nil && ps.Tests.Status != ProofPass && ps.Tests.Status != ProofSkip {
		return false
	}
	if ps.Build != nil && ps.Build.Status != ProofPass && ps.Build.Status != ProofSkip {
		return false
	}
	if ps.Lint != nil && ps.Lint.Status != ProofPass && ps.Lint.Status != ProofSkip {
		return false
	}
	if ps.Adversarial != nil && ps.Adversarial.Status != ProofPass && ps.Adversarial.Status != ProofSkip {
		return false
	}
	if ps.Integration != nil && ps.Integration.Status != ProofPass && ps.Integration.Status != ProofSkip {
		return false
	}
	return true
}

// AnyFail returns true if any proof has FAIL status.
func (ps *ProofSet) AnyFail() bool {
	if ps.Tests != nil && ps.Tests.Status == ProofFail {
		return true
	}
	if ps.Build != nil && ps.Build.Status == ProofFail {
		return true
	}
	if ps.Lint != nil && ps.Lint.Status == ProofFail {
		return true
	}
	if ps.Adversarial != nil && ps.Adversarial.Status == ProofFail {
		return true
	}
	if ps.Integration != nil && ps.Integration.Status == ProofFail {
		return true
	}
	return false
}

// Summary returns a human-readable summary of all proofs.
func (ps *ProofSet) Summary() string {
	var parts []string

	if ps.Tests != nil {
		parts = append(parts, "tests: "+string(ps.Tests.Status))
	}
	if ps.Build != nil {
		parts = append(parts, "build: "+string(ps.Build.Status))
	}
	if ps.Lint != nil {
		parts = append(parts, "lint: "+string(ps.Lint.Status))
	}
	if ps.Adversarial != nil {
		parts = append(parts, "adversarial: "+string(ps.Adversarial.Status))
	}
	if ps.Integration != nil {
		parts = append(parts, "integration: "+string(ps.Integration.Status))
	}

	if len(parts) == 0 {
		return "no proofs collected"
	}

	return strings.Join(parts, ", ")
}
