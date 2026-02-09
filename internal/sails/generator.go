// Package sails implements the White Sails confidence signaling system for Ariadne.
// This file implements the WHITE_SAILS.yaml generator per TDD Section 5.
package sails

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
	"github.com/autom8y/knossos/internal/validation"
	"gopkg.in/yaml.v3"
)

// Generator creates WHITE_SAILS.yaml for a session.
type Generator struct {
	// SessionPath is the path to the session directory.
	SessionPath string

	// Validator is the schema validator (optional, for validation step).
	Validator *validation.Validator

	// Now is a function that returns the current time (for testing).
	Now func() time.Time
}

// GenerateResult contains the output of sails generation.
type GenerateResult struct {
	// Color is the final confidence signal after modifiers.
	Color Color

	// ComputedBase is the computed color before human modifiers.
	ComputedBase Color

	// Reasons explains why this color was computed.
	Reasons []string

	// FilePath is the path to the generated WHITE_SAILS.yaml.
	FilePath string

	// Proofs contains the collected proof items.
	Proofs map[string]ColorProof

	// SessionID is the session identifier.
	SessionID string

	// GeneratedAt is when the sails were generated.
	GeneratedAt time.Time

	// OpenQuestions from the session context.
	OpenQuestions []string

	// Modifiers applied to the color.
	Modifiers []Modifier

	// Complexity tier from session context.
	Complexity string

	// SessionType from session context.
	SessionType string
}

// WhiteSailsYAML represents the WHITE_SAILS.yaml file structure.
type WhiteSailsYAML struct {
	SchemaVersion string                  `yaml:"schema_version"`
	SessionID     string                  `yaml:"session_id"`
	GeneratedAt   string                  `yaml:"generated_at"`
	Color         string                  `yaml:"color"`
	ComputedBase  string                  `yaml:"computed_base"`
	Proofs        map[string]YAMLProof    `yaml:"proofs"`
	OpenQuestions []string                `yaml:"open_questions"`
	Modifiers     []YAMLModifier          `yaml:"modifiers,omitempty"`
	Complexity    string                  `yaml:"complexity,omitempty"`
	Type          string                  `yaml:"type,omitempty"`
	QAUpgrade     *YAMLQAUpgrade          `yaml:"qa_upgrade,omitempty"`
}

// YAMLProof represents a proof item in YAML format.
type YAMLProof struct {
	Status       string `yaml:"status"`
	EvidencePath string `yaml:"evidence_path,omitempty"`
	Summary      string `yaml:"summary,omitempty"`
	ExitCode     *int   `yaml:"exit_code,omitempty"`
	Timestamp    string `yaml:"timestamp,omitempty"`
}

// YAMLModifier represents a modifier in YAML format.
type YAMLModifier struct {
	Type          string `yaml:"type"`
	Justification string `yaml:"justification"`
	AppliedBy     string `yaml:"applied_by"`
	Timestamp     string `yaml:"timestamp,omitempty"`
}

// YAMLQAUpgrade represents a QA upgrade in YAML format.
type YAMLQAUpgrade struct {
	UpgradedAt              string   `yaml:"upgraded_at,omitempty"`
	QASessionID             string   `yaml:"qa_session_id,omitempty"`
	ConstraintResolutionLog string   `yaml:"constraint_resolution_log,omitempty"`
	AdversarialTestsAdded   []string `yaml:"adversarial_tests_added,omitempty"`
}

// NewGenerator creates a new Generator for the given session.
func NewGenerator(sessionPath string) *Generator {
	return &Generator{
		SessionPath: sessionPath,
		Now:         time.Now,
	}
}

// NewGeneratorWithValidator creates a Generator with schema validation.
func NewGeneratorWithValidator(sessionPath string, validator *validation.Validator) *Generator {
	return &Generator{
		SessionPath: sessionPath,
		Validator:   validator,
		Now:         time.Now,
	}
}

// Generate creates WHITE_SAILS.yaml and returns the result.
func (g *Generator) Generate() (*GenerateResult, error) {
	if g.SessionPath == "" {
		return nil, errors.New(errors.CodeUsageError, "session path is required")
	}

	// Verify session directory exists
	info, err := os.Stat(g.SessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeSessionNotFound,
				"session directory not found",
				map[string]interface{}{"path": g.SessionPath})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to access session directory", err)
	}
	if !info.IsDir() {
		return nil, errors.NewWithDetails(errors.CodeUsageError,
			"path is not a directory",
			map[string]interface{}{"path": g.SessionPath})
	}

	// Step 1: Collect proofs from session directory
	proofSet, err := CollectProofs(g.SessionPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to collect proofs", err)
	}

	// Step 2: Load session context to get open questions, blockers, modifiers, complexity, and session type
	sessionID, complexity, sessionType, openQuestions, blockers, modifiers, qaUpgrade, err := g.loadSessionContext()
	if err != nil {
		// Session context is optional - use defaults if not found
		sessionID = g.extractSessionIDFromPath()
		complexity = "MODULE" // Default complexity
		sessionType = "standard"
		openQuestions = nil
		blockers = nil
		modifiers = nil
		qaUpgrade = nil
	}

	// Step 3: Convert ProofSet to ColorProof map
	proofs := g.proofSetToColorProofs(proofSet)

	// Step 4: Build ColorInput and compute color
	colorInput := ColorInput{
		SessionType:   sessionType,
		Complexity:    complexity,
		Proofs:        proofs,
		OpenQuestions: openQuestions,
		Blockers:      blockers,
		Modifiers:     modifiers,
		QAUpgrade:     qaUpgrade,
	}

	colorResult := ComputeColor(colorInput)

	// Step 5: Prepare result
	now := g.Now().UTC()
	filePath := filepath.Join(g.SessionPath, "WHITE_SAILS.yaml")

	result := &GenerateResult{
		Color:         colorResult.Color,
		ComputedBase:  colorResult.ComputedBase,
		Reasons:       colorResult.Reasons,
		FilePath:      filePath,
		Proofs:        proofs,
		SessionID:     sessionID,
		GeneratedAt:   now,
		OpenQuestions: openQuestions,
		Modifiers:     modifiers,
		Complexity:    complexity,
		SessionType:   sessionType,
	}

	// Step 6: Generate WHITE_SAILS.yaml content
	yamlContent, err := g.generateYAML(result, qaUpgrade)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to generate YAML", err)
	}

	// Step 7: Validate against schema if validator is available
	if g.Validator != nil {
		validationResult, err := g.Validator.ValidateWhiteSails(yamlContent)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "schema validation error", err)
		}
		if !validationResult.Valid {
			issues := make([]string, len(validationResult.Issues))
			for i, issue := range validationResult.Issues {
				issues[i] = issue.Message
			}
			return nil, errors.NewWithDetails(errors.CodeSchemaInvalid,
				"generated WHITE_SAILS.yaml failed schema validation",
				map[string]interface{}{"issues": issues})
		}
	}

	// Step 8: Write WHITE_SAILS.yaml to session directory
	if err := os.WriteFile(filePath, yamlContent, 0644); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write WHITE_SAILS.yaml", err)
	}

	return result, nil
}

// loadSessionContext reads SESSION_CONTEXT.md to extract metadata.
func (g *Generator) loadSessionContext() (sessionID, complexity, sessionType string, openQuestions []string, blockers []string, modifiers []Modifier, qaUpgrade *QAUpgrade, err error) {
	contextPath := filepath.Join(g.SessionPath, "SESSION_CONTEXT.md")

	content, readErr := os.ReadFile(contextPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return "", "", "", nil, nil, nil, nil, errors.New(errors.CodeFileNotFound, "SESSION_CONTEXT.md not found")
		}
		return "", "", "", nil, nil, nil, nil, errors.Wrap(errors.CodeGeneralError, "failed to read SESSION_CONTEXT.md", readErr)
	}

	ctx, parseErr := session.ParseContext(content)
	if parseErr != nil {
		return "", "", "", nil, nil, nil, nil, errors.Wrap(errors.CodeSchemaInvalid, "failed to parse SESSION_CONTEXT.md", parseErr)
	}

	// Extract session ID and complexity from context
	sessionID = ctx.SessionID
	complexity = ctx.Complexity
	if complexity == "" {
		complexity = "MODULE" // Default
	}

	// Extract session type from body (defaults to standard)
	sessionType = extractSessionType(ctx.Body)

	// Extract open questions from the body
	openQuestions = extractOpenQuestions(ctx.Body)

	// Extract blockers from the body
	blockers = extractBlockers(ctx.Body)

	// Extract modifiers from the body (if any declared)
	modifiers = extractModifiers(ctx.Body)

	// QA upgrade would be extracted if present (typically from a QA session)
	qaUpgrade = extractQAUpgrade(ctx.Body)

	return sessionID, complexity, sessionType, openQuestions, blockers, modifiers, qaUpgrade, nil
}

// extractSessionIDFromPath extracts session ID from the directory name.
func (g *Generator) extractSessionIDFromPath() string {
	return filepath.Base(g.SessionPath)
}

// proofSetToColorProofs converts a ProofSet to a map of ColorProof.
func (g *Generator) proofSetToColorProofs(ps *ProofSet) map[string]ColorProof {
	proofs := make(map[string]ColorProof)

	if ps.Tests != nil {
		proofs["tests"] = g.proofItemToColorProof(ps.Tests)
	}
	if ps.Build != nil {
		proofs["build"] = g.proofItemToColorProof(ps.Build)
	}
	if ps.Lint != nil {
		proofs["lint"] = g.proofItemToColorProof(ps.Lint)
	}
	if ps.Adversarial != nil {
		proofs["adversarial"] = g.proofItemToColorProof(ps.Adversarial)
	}
	if ps.Integration != nil {
		proofs["integration"] = g.proofItemToColorProof(ps.Integration)
	}

	return proofs
}

// proofItemToColorProof converts a ProofItem to a ColorProof.
func (g *Generator) proofItemToColorProof(item *ProofItem) ColorProof {
	exitCode := item.ExitCode
	timestamp := item.Timestamp
	return ColorProof{
		Status:       item.Status,
		EvidencePath: item.EvidencePath,
		Summary:      item.Summary,
		ExitCode:     &exitCode,
		Timestamp:    &timestamp,
	}
}

// generateYAML creates the WHITE_SAILS.yaml content.
func (g *Generator) generateYAML(result *GenerateResult, qaUpgrade *QAUpgrade) ([]byte, error) {
	// Convert proofs to YAML format
	yamlProofs := make(map[string]YAMLProof)
	for name, proof := range result.Proofs {
		yamlProof := YAMLProof{
			Status:       string(proof.Status),
			EvidencePath: proof.EvidencePath,
			Summary:      proof.Summary,
		}
		if proof.ExitCode != nil {
			yamlProof.ExitCode = proof.ExitCode
		}
		if proof.Timestamp != nil {
			yamlProof.Timestamp = proof.Timestamp.UTC().Format(time.RFC3339)
		}
		yamlProofs[name] = yamlProof
	}

	// Convert modifiers to YAML format
	var yamlModifiers []YAMLModifier
	for _, mod := range result.Modifiers {
		yamlMod := YAMLModifier{
			Type:          string(mod.Type),
			Justification: mod.Justification,
			AppliedBy:     mod.AppliedBy,
		}
		if mod.Timestamp != nil {
			yamlMod.Timestamp = mod.Timestamp.UTC().Format(time.RFC3339)
		}
		yamlModifiers = append(yamlModifiers, yamlMod)
	}

	// Convert QA upgrade to YAML format
	var yamlQAUpgrade *YAMLQAUpgrade
	if qaUpgrade != nil {
		yamlQAUpgrade = &YAMLQAUpgrade{
			QASessionID:             qaUpgrade.QASessionID,
			ConstraintResolutionLog: qaUpgrade.ConstraintResolutionLog,
			AdversarialTestsAdded:   qaUpgrade.AdversarialTestsAdded,
		}
		if qaUpgrade.UpgradedAt != nil {
			yamlQAUpgrade.UpgradedAt = qaUpgrade.UpgradedAt.UTC().Format(time.RFC3339)
		}
	}

	// Build the YAML structure
	sails := WhiteSailsYAML{
		SchemaVersion: "1.0",
		SessionID:     result.SessionID,
		GeneratedAt:   result.GeneratedAt.Format(time.RFC3339),
		Color:         string(result.Color),
		ComputedBase:  string(result.ComputedBase),
		Proofs:        yamlProofs,
		OpenQuestions: result.OpenQuestions,
		Modifiers:     yamlModifiers,
		Complexity:    result.Complexity,
		Type:          result.SessionType,
		QAUpgrade:     yamlQAUpgrade,
	}

	// Ensure open_questions is an empty array, not null
	if sails.OpenQuestions == nil {
		sails.OpenQuestions = []string{}
	}

	return yaml.Marshal(sails)
}

// extractSessionType parses the session body for session type.
// Looks for a section like "## Session Type" followed by the type (spike, hotfix, standard).
func extractSessionType(body string) string {
	// Pattern to find Session Type section
	sessionTypePattern := regexp.MustCompile(`(?i)##\s*Session\s*Type\s*\n`)
	match := sessionTypePattern.FindStringIndex(body)
	if match == nil {
		return "standard"
	}

	// Extract content after the header until the next section or end
	startIdx := match[1]
	remaining := body[startIdx:]

	// Find the next section header (## Something)
	nextSectionPattern := regexp.MustCompile(`\n##\s+`)
	nextMatch := nextSectionPattern.FindStringIndex(remaining)

	var sectionContent string
	if nextMatch != nil {
		sectionContent = remaining[:nextMatch[0]]
	} else {
		sectionContent = remaining
	}

	// Normalize and check for known types
	sectionContent = strings.TrimSpace(strings.ToLower(sectionContent))

	if strings.HasPrefix(sectionContent, "spike") {
		return "spike"
	}
	if strings.HasPrefix(sectionContent, "hotfix") {
		return "hotfix"
	}

	return "standard"
}

// extractOpenQuestions parses the session body for open questions.
// Looks for a section like "## Open Questions" followed by bullet points.
func extractOpenQuestions(body string) []string {
	var questions []string

	// Pattern to find Open Questions section
	openQuestionsPattern := regexp.MustCompile(`(?i)##\s*Open\s*Questions?\s*\n`)
	match := openQuestionsPattern.FindStringIndex(body)
	if match == nil {
		return questions
	}

	// Extract content after the header until the next section or end
	startIdx := match[1]
	remaining := body[startIdx:]

	// Find the next section header (## Something)
	nextSectionPattern := regexp.MustCompile(`\n##\s+`)
	nextMatch := nextSectionPattern.FindStringIndex(remaining)

	var sectionContent string
	if nextMatch != nil {
		sectionContent = remaining[:nextMatch[0]]
	} else {
		sectionContent = remaining
	}

	// Extract bullet points (lines starting with - or *)
	bulletPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*(.+)$`)
	matches := bulletPattern.FindAllStringSubmatch(sectionContent, -1)
	for _, m := range matches {
		if len(m) > 1 {
			question := strings.TrimSpace(m[1])
			if question != "" && question != "None" && question != "None yet." && question != "N/A" {
				questions = append(questions, question)
			}
		}
	}

	return questions
}

// extractBlockers parses the session body for blockers.
// Looks for a section like "## Blockers" followed by bullet points.
// Ignores common "no blocker" phrases like "None", "None yet", "N/A".
func extractBlockers(body string) []string {
	var blockers []string

	// Pattern to find Blockers section
	blockersPattern := regexp.MustCompile(`(?i)##\s*Blockers?\s*\n`)
	match := blockersPattern.FindStringIndex(body)
	if match == nil {
		return blockers
	}

	// Extract content after the header until the next section or end
	startIdx := match[1]
	remaining := body[startIdx:]

	// Find the next section header (## Something)
	nextSectionPattern := regexp.MustCompile(`\n##\s+`)
	nextMatch := nextSectionPattern.FindStringIndex(remaining)

	var sectionContent string
	if nextMatch != nil {
		sectionContent = remaining[:nextMatch[0]]
	} else {
		sectionContent = remaining
	}

	// Extract bullet points (lines starting with - or *)
	bulletPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*(.+)$`)
	matches := bulletPattern.FindAllStringSubmatch(sectionContent, -1)
	for _, m := range matches {
		if len(m) > 1 {
			blocker := strings.TrimSpace(m[1])
			// Filter out common "no blocker" phrases
			blockerLower := strings.ToLower(blocker)
			if blocker != "" &&
				blockerLower != "none" &&
				blockerLower != "none yet" &&
				blockerLower != "none yet." &&
				blockerLower != "n/a" &&
				!strings.HasPrefix(blockerLower, "none ") {
				blockers = append(blockers, blocker)
			}
		}
	}

	return blockers
}

// extractModifiers parses the session body for declared modifiers.
// Looks for a section like "## Modifiers" or inline modifier declarations.
func extractModifiers(body string) []Modifier {
	var modifiers []Modifier

	// Pattern to find Modifiers section
	modifiersPattern := regexp.MustCompile(`(?i)##\s*Modifiers?\s*\n`)
	match := modifiersPattern.FindStringIndex(body)
	if match == nil {
		return modifiers
	}

	// Extract content after the header until the next section or end
	startIdx := match[1]
	remaining := body[startIdx:]

	// Find the next section header
	nextSectionPattern := regexp.MustCompile(`\n##\s+`)
	nextMatch := nextSectionPattern.FindStringIndex(remaining)

	var sectionContent string
	if nextMatch != nil {
		sectionContent = remaining[:nextMatch[0]]
	} else {
		sectionContent = remaining
	}

	// Pattern to extract modifier declarations
	// Format: - DOWNGRADE_TO_GRAY: justification text (applied_by: agent|human)
	modifierPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*(DOWNGRADE_TO_GRAY|DOWNGRADE_TO_BLACK|HUMAN_OVERRIDE_GRAY):\s*(.+?)(?:\s*\(applied_by:\s*(agent|human)\))?$`)
	matches := modifierPattern.FindAllStringSubmatch(sectionContent, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			modType := ModifierType(m[1])
			justification := strings.TrimSpace(m[2])
			appliedBy := "agent" // Default
			if len(m) >= 4 && m[3] != "" {
				appliedBy = m[3]
			}

			if justification != "" && modType.IsValid() {
				modifiers = append(modifiers, Modifier{
					Type:          modType,
					Justification: justification,
					AppliedBy:     appliedBy,
				})
			}
		}
	}

	return modifiers
}

// extractQAUpgrade extracts QA upgrade information from the SESSION_CONTEXT.md body.
// Format:
// ## QA Upgrade
// - qa_session_id: session-20260105-123456-qa
// - upgraded_at: 2026-01-05T12:34:56Z
// - constraint_resolution_log: path/to/log.md
// - adversarial_tests_added:
//   - tests/adversarial/test1.go
//   - tests/adversarial/test2.go
func extractQAUpgrade(body string) *QAUpgrade {
	// Pattern to find QA Upgrade section
	qaUpgradePattern := regexp.MustCompile(`(?i)##\s*QA\s*Upgrade\s*\n`)
	match := qaUpgradePattern.FindStringIndex(body)
	if match == nil {
		return nil
	}

	// Extract content after the header until the next section or end
	startIdx := match[1]
	remaining := body[startIdx:]

	// Find the next section header (## Something)
	nextSectionPattern := regexp.MustCompile(`\n##\s+`)
	nextMatch := nextSectionPattern.FindStringIndex(remaining)

	var sectionContent string
	if nextMatch != nil {
		sectionContent = remaining[:nextMatch[0]]
	} else {
		sectionContent = remaining
	}

	upgrade := &QAUpgrade{}
	hasData := false

	// Extract qa_session_id
	qaSessionIDPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*qa_session_id:\s*(.+)$`)
	if m := qaSessionIDPattern.FindStringSubmatch(sectionContent); m != nil && len(m) > 1 {
		upgrade.QASessionID = strings.TrimSpace(m[1])
		hasData = true
	}

	// Extract upgraded_at timestamp
	upgradedAtPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*upgraded_at:\s*(.+)$`)
	if m := upgradedAtPattern.FindStringSubmatch(sectionContent); m != nil && len(m) > 1 {
		timestampStr := strings.TrimSpace(m[1])
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			upgrade.UpgradedAt = &t
			hasData = true
		}
	}

	// Extract constraint_resolution_log
	logPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*constraint_resolution_log:\s*(.+)$`)
	if m := logPattern.FindStringSubmatch(sectionContent); m != nil && len(m) > 1 {
		upgrade.ConstraintResolutionLog = strings.TrimSpace(m[1])
		hasData = true
	}

	// Extract adversarial_tests_added (multi-line list)
	testsPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*adversarial_tests_added:\s*$`)
	if testsMatch := testsPattern.FindStringIndex(sectionContent); testsMatch != nil {
		// Find all indented list items after adversarial_tests_added
		afterTests := sectionContent[testsMatch[1]:]
		testItemPattern := regexp.MustCompile(`(?m)^[\s]{2,}[-*]\s*(.+)$`)
		testMatches := testItemPattern.FindAllStringSubmatch(afterTests, -1)
		for _, tm := range testMatches {
			if len(tm) > 1 {
				testPath := strings.TrimSpace(tm[1])
				if testPath != "" {
					upgrade.AdversarialTestsAdded = append(upgrade.AdversarialTestsAdded, testPath)
					hasData = true
				}
			}
		}
	}

	// Return nil if no data was found
	if !hasData {
		return nil
	}

	return upgrade
}

// GeneratorFromProject creates a Generator for a specific session in a project.
func GeneratorFromProject(projectRoot string, sessionID string) (*Generator, error) {
	if projectRoot == "" {
		return nil, errors.New(errors.CodeUsageError, "project root is required")
	}
	if sessionID == "" {
		return nil, errors.New(errors.CodeSessionNotFound, "no session ID provided")
	}

	resolver := paths.NewResolver(projectRoot)
	sessionDir := resolver.SessionDir(strings.TrimSpace(sessionID))

	// Verify session directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return nil, errors.New(errors.CodeSessionNotFound, "session directory not found: "+sessionID)
	}

	return NewGenerator(sessionDir), nil
}

// GeneratorFromProjectWithValidator creates a Generator with validation for a specific session.
func GeneratorFromProjectWithValidator(projectRoot string, sessionID string, validator *validation.Validator) (*Generator, error) {
	gen, err := GeneratorFromProject(projectRoot, sessionID)
	if err != nil {
		return nil, err
	}
	gen.Validator = validator
	return gen, nil
}
