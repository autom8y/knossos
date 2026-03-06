// Package sails implements the White Sails confidence signaling system per Knossos Doctrine v2.
// White Sails provides honest confidence signals at session wrap to prevent Aegeus failures
// (false confidence leading to production issues).
//
// This file implements the color computation algorithm per TDD Section 4.1.
package sails

import (
	"strconv"
	"strings"
	"time"
)

// Color represents the session confidence level.
// There are only three states: WHITE (high confidence), GRAY (unknown), BLACK (known failure).
type Color string

const (
	// ColorWhite indicates high confidence - ship without QA.
	// Requires all proofs present + tests pass + lint clean + no open questions.
	ColorWhite Color = "WHITE"

	// ColorGray indicates unknown confidence - needs QA review.
	// Missing proofs OR open questions OR complexity ceiling OR declared uncertainty.
	ColorGray Color = "GRAY"

	// ColorBlack indicates known failure - do not ship.
	// Tests failing OR build broken OR explicit blocker.
	ColorBlack Color = "BLACK"
)

// String returns the string representation of the color.
func (c Color) String() string {
	return string(c)
}

// IsValid checks if the color is a valid value.
func (c Color) IsValid() bool {
	switch c {
	case ColorWhite, ColorGray, ColorBlack:
		return true
	default:
		return false
	}
}

// ParseColorFromYAML extracts the sails color from YAML content.
// Looks for a "color:" key and maps its value to a Color constant.
// Returns ColorGray if the color cannot be determined.
func ParseColorFromYAML(content []byte) Color {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "color:"); ok {
			value := strings.TrimSpace(after)
			value = strings.Trim(value, "\"'")
			switch strings.ToUpper(value) {
			case "WHITE":
				return ColorWhite
			case "GRAY":
				return ColorGray
			case "BLACK":
				return ColorBlack
			}
		}
	}
	return ColorGray
}

// ModifierType represents the type of color modifier.
type ModifierType string

const (
	// ModifierDowngradeToGray downgrades WHITE to GRAY.
	ModifierDowngradeToGray ModifierType = "DOWNGRADE_TO_GRAY"
	// ModifierDowngradeToBlack downgrades any color to BLACK.
	ModifierDowngradeToBlack ModifierType = "DOWNGRADE_TO_BLACK"
	// ModifierHumanOverrideGray forces color to GRAY regardless of computed base.
	ModifierHumanOverrideGray ModifierType = "HUMAN_OVERRIDE_GRAY"
)

// String returns the string representation of the modifier type.
func (m ModifierType) String() string {
	return string(m)
}

// IsValid checks if the modifier type is valid.
func (m ModifierType) IsValid() bool {
	switch m {
	case ModifierDowngradeToGray, ModifierDowngradeToBlack, ModifierHumanOverrideGray:
		return true
	default:
		return false
	}
}

// ProofStatus represents the status of a proof item.
type ProofStatus string

const (
	// ProofPass indicates the proof passed validation.
	ProofPass ProofStatus = "PASS"
	// ProofFail indicates the proof failed validation.
	ProofFail ProofStatus = "FAIL"
	// ProofSkip indicates the proof was intentionally skipped.
	ProofSkip ProofStatus = "SKIP"
	// ProofUnknown indicates the proof status could not be determined.
	ProofUnknown ProofStatus = "UNKNOWN"
)

// IsPassing returns true if this status represents a passing state.
func (s ProofStatus) IsPassing() bool {
	return s == ProofPass || s == ProofSkip
}

// Modifier represents a human-declared adjustment to the computed color.
type Modifier struct {
	Type          ModifierType `yaml:"type" json:"type"`
	Justification string       `yaml:"justification" json:"justification"`
	AppliedBy     string       `yaml:"applied_by" json:"applied_by"` // "agent" or "human"
	Timestamp     *time.Time   `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
}

// QAUpgrade represents a QA session that upgraded gray to white.
type QAUpgrade struct {
	UpgradedAt              *time.Time `yaml:"upgraded_at,omitempty" json:"upgraded_at,omitempty"`
	QASessionID             string     `yaml:"qa_session_id,omitempty" json:"qa_session_id,omitempty"`
	ConstraintResolutionLog string     `yaml:"constraint_resolution_log,omitempty" json:"constraint_resolution_log,omitempty"`
	AdversarialTestsAdded   []string   `yaml:"adversarial_tests_added,omitempty" json:"adversarial_tests_added,omitempty"`
}

// ColorProof represents evidence of a quality check for color computation.
// This is a thin wrapper that uses ProofStatus from the sails package.
type ColorProof struct {
	Status       ProofStatus `yaml:"status" json:"status"`
	EvidencePath string      `yaml:"evidence_path,omitempty" json:"evidence_path,omitempty"`
	Summary      string      `yaml:"summary,omitempty" json:"summary,omitempty"`
	ExitCode     *int        `yaml:"exit_code,omitempty" json:"exit_code,omitempty"`
	Timestamp    *time.Time  `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
}

// Proof is an alias for ColorProof for test compatibility.
type Proof = ColorProof

// ColorInput contains all inputs for color computation.
type ColorInput struct {
	// SessionType is the type of session: "standard", "spike", "hotfix"
	SessionType string `yaml:"session_type" json:"session_type"`

	// Complexity is the complexity tier: "PATCH", "SCRIPT", "MODULE", "SERVICE", "INITIATIVE", "MIGRATION", "PLATFORM"
	Complexity string `yaml:"complexity" json:"complexity"`

	// Proofs is a map of proof name to proof result.
	// Standard proof names: "tests", "build", "lint", "adversarial", "integration"
	Proofs map[string]ColorProof `yaml:"proofs" json:"proofs"`

	// OpenQuestions is a list of unresolved questions.
	// Any open question triggers gray ceiling.
	OpenQuestions []string `yaml:"open_questions" json:"open_questions"`

	// Blockers is a list of explicit blockers from the session context.
	// Any blocker triggers BLACK (known failure).
	Blockers []string `yaml:"blockers,omitempty" json:"blockers,omitempty"`

	// Modifiers are human-declared adjustments.
	Modifiers []Modifier `yaml:"modifiers" json:"modifiers"`

	// QAUpgrade is present if a QA session upgraded this from gray to white.
	QAUpgrade *QAUpgrade `yaml:"qa_upgrade,omitempty" json:"qa_upgrade,omitempty"`
}

// ColorResult contains the computed color and reasoning.
type ColorResult struct {
	// Color is the final confidence signal after modifiers.
	Color Color `yaml:"color" json:"color"`

	// ComputedBase is the computed color before human modifiers.
	ComputedBase Color `yaml:"computed_base" json:"computed_base"`

	// Reasons explains why this color was computed.
	Reasons []string `yaml:"reasons" json:"reasons"`
}

// ComputeColor computes the sails color based on the algorithm in TDD 4.1.
//
// Algorithm summary:
//  1. Check for failures (BLACK): explicit blockers OR proof failures
//  2. Check for open questions (GRAY ceiling)
//  3. Check session type ceiling (spike/hotfix = GRAY)
//  4. Check proof completeness per complexity
//  5. All proofs present = WHITE
//  6. Apply modifiers (downgrade only)
//  7. QA upgrade path (only from GRAY)
func ComputeColor(input ColorInput) ColorResult {
	var reasons []string

	// Step 1a: Check for explicit blockers (BLACK)
	if len(input.Blockers) > 0 {
		reasons = append(reasons, "explicit blockers present: black sails (do not ship)")
		for _, blocker := range input.Blockers {
			reasons = append(reasons, "  - "+blocker)
		}
		return ColorResult{
			Color:        ColorBlack,
			ComputedBase: ColorBlack,
			Reasons:      reasons,
		}
	}

	// Step 1b: Check for proof failures (BLACK)
	for name, proof := range input.Proofs {
		if proof.Status == ProofFail {
			reasons = append(reasons, "proof '"+name+"' has status FAIL")
			return ColorResult{
				Color:        ColorBlack,
				ComputedBase: ColorBlack,
				Reasons:      reasons,
			}
		}
	}

	// Step 2: Check for open questions (GRAY ceiling)
	if len(input.OpenQuestions) > 0 {
		reasons = append(reasons, "open questions present: gray ceiling applied")
		computedBase := ColorGray
		finalColor := applyModifiers(computedBase, input.Modifiers, &reasons)
		finalColor = applyQAUpgrade(computedBase, finalColor, input.QAUpgrade, &reasons)
		return ColorResult{
			Color:        finalColor,
			ComputedBase: computedBase,
			Reasons:      reasons,
		}
	}

	// Step 3: Check session type ceiling
	sessionType := normalizeSessionType(input.SessionType)
	if sessionType == "spike" {
		reasons = append(reasons, "session type 'spike' has gray ceiling (spikes never white)")
		computedBase := ColorGray
		finalColor := applyModifiers(computedBase, input.Modifiers, &reasons)
		finalColor = applyQAUpgrade(computedBase, finalColor, input.QAUpgrade, &reasons)
		return ColorResult{
			Color:        finalColor,
			ComputedBase: computedBase,
			Reasons:      reasons,
		}
	}

	if sessionType == "hotfix" {
		reasons = append(reasons, "session type 'hotfix' has gray ceiling (expedited gray)")
		computedBase := ColorGray
		finalColor := applyModifiers(computedBase, input.Modifiers, &reasons)
		finalColor = applyQAUpgrade(computedBase, finalColor, input.QAUpgrade, &reasons)
		return ColorResult{
			Color:        finalColor,
			ComputedBase: computedBase,
			Reasons:      reasons,
		}
	}

	// Step 4: Check proof completeness per complexity
	requiredProofs := GetRequiredProofsForColor(input.Complexity)
	for _, proofName := range requiredProofs {
		proof, exists := input.Proofs[proofName]
		if !exists {
			reasons = append(reasons, "required proof '"+proofName+"' is missing")
			computedBase := ColorGray
			finalColor := applyModifiers(computedBase, input.Modifiers, &reasons)
			finalColor = applyQAUpgrade(computedBase, finalColor, input.QAUpgrade, &reasons)
			return ColorResult{
				Color:        finalColor,
				ComputedBase: computedBase,
				Reasons:      reasons,
			}
		}
		if !isProofStatusPassing(proof.Status) {
			reasons = append(reasons, "required proof '"+proofName+"' has status "+string(proof.Status)+" (not PASS or SKIP)")
			computedBase := ColorGray
			finalColor := applyModifiers(computedBase, input.Modifiers, &reasons)
			finalColor = applyQAUpgrade(computedBase, finalColor, input.QAUpgrade, &reasons)
			return ColorResult{
				Color:        finalColor,
				ComputedBase: computedBase,
				Reasons:      reasons,
			}
		}
	}

	// Step 5: All proofs present and passing
	reasons = append(reasons, "all required proofs present and passing")
	computedBase := ColorWhite
	finalColor := applyModifiers(computedBase, input.Modifiers, &reasons)

	// Note: QA upgrade only applies when computed_base is GRAY, so skip here
	// since computed_base is WHITE

	return ColorResult{
		Color:        finalColor,
		ComputedBase: computedBase,
		Reasons:      reasons,
	}
}

// isProofStatusPassing returns true if the proof is in a passing state (PASS or SKIP).
func isProofStatusPassing(status ProofStatus) bool {
	return status == ProofPass || status == ProofSkip
}

// isProofStatusValid checks if the proof status is valid.
func isProofStatusValid(status ProofStatus) bool {
	switch status {
	case ProofPass, ProofFail, ProofSkip, ProofUnknown:
		return true
	default:
		return false
	}
}

// normalizeSessionType normalizes session type string to lowercase.
// Empty string defaults to "standard".
func normalizeSessionType(sessionType string) string {
	if sessionType == "" {
		return "standard"
	}
	// Convert to lowercase for comparison
	switch sessionType {
	case "spike", "SPIKE", "Spike":
		return "spike"
	case "hotfix", "HOTFIX", "Hotfix":
		return "hotfix"
	default:
		return "standard"
	}
}

// applyModifiers applies modifiers to the color (downgrade only).
// Returns the final color after all modifiers are applied.
func applyModifiers(color Color, modifiers []Modifier, reasons *[]string) Color {
	finalColor := color

	for _, mod := range modifiers {
		switch mod.Type {
		case ModifierDowngradeToGray:
			if finalColor == ColorWhite {
				finalColor = ColorGray
				*reasons = append(*reasons, "modifier DOWNGRADE_TO_GRAY applied: "+mod.Justification)
			}
		case ModifierDowngradeToBlack:
			finalColor = ColorBlack
			*reasons = append(*reasons, "modifier DOWNGRADE_TO_BLACK applied: "+mod.Justification)
		case ModifierHumanOverrideGray:
			finalColor = ColorGray
			*reasons = append(*reasons, "modifier HUMAN_OVERRIDE_GRAY applied: "+mod.Justification)
		}
	}

	return finalColor
}

// applyQAUpgrade applies QA upgrade if conditions are met.
// QA upgrade can only promote GRAY to WHITE, and only if:
// - computed_base is GRAY
// - current color is still GRAY (not downgraded to BLACK)
// - QA upgrade has constraint_resolution_log
// - QA upgrade has at least one adversarial test added
func applyQAUpgrade(computedBase, currentColor Color, qaUpgrade *QAUpgrade, reasons *[]string) Color {
	if qaUpgrade == nil {
		return currentColor
	}

	// QA upgrade only applies from GRAY base
	if computedBase != ColorGray {
		return currentColor
	}

	// Can only upgrade if still GRAY (not downgraded to BLACK)
	if currentColor != ColorGray {
		return currentColor
	}

	// Must have constraint resolution log
	if qaUpgrade.ConstraintResolutionLog == "" {
		*reasons = append(*reasons, "QA upgrade missing constraint_resolution_log: cannot upgrade")
		return currentColor
	}

	// Must have adversarial tests added
	if len(qaUpgrade.AdversarialTestsAdded) == 0 {
		*reasons = append(*reasons, "QA upgrade has no adversarial_tests_added: cannot upgrade")
		return currentColor
	}

	// All conditions met: upgrade to WHITE
	*reasons = append(*reasons, "QA upgrade applied: gray -> white via QA session "+qaUpgrade.QASessionID)
	return ColorWhite
}

// GetRequiredProofsForColor returns the list of required proof names for a given complexity.
// This function is specific to color computation and uses the thresholds package.
// Based on TDD 4.2 Required Proofs by Complexity table.
//
// | Complexity | tests | build | lint | adversarial | integration |
// |------------|-------|-------|------|-------------|-------------|
// | PATCH      | Req   | Req   | Req  | -           | -           |
// | SCRIPT     | Req   | Req   | Req  | -           | -           |
// | MODULE     | Req   | Req   | Req  | -           | -           |
// | SERVICE    | Req   | Req   | Req  | Rec         | Rec         |
// | INITIATIVE | Req   | Req   | Req  | Req         | Req         |
// | MIGRATION  | Req   | Req   | Req  | Req         | Req         |
// | PLATFORM   | Req   | Req   | Req  | Req         | Req         |
//
// Note: "Recommended" proofs are not required for WHITE, only "Required".
func GetRequiredProofsForColor(complexity string) []string {
	// Use the existing thresholds function
	return GetRequiredProofNames(Complexity(complexity))
}

// IsRequiredProofForColor checks if a proof is required for the given complexity.
func IsRequiredProofForColor(complexity, proofName string) bool {
	return IsProofRequired(Complexity(complexity), proofName)
}

// ValidateColorInput validates the input for color computation.
// Returns a list of validation errors (empty if valid).
func ValidateColorInput(input ColorInput) []string {
	var errors []string

	// Validate proofs
	for name, proof := range input.Proofs {
		if !isProofStatusValid(proof.Status) {
			errors = append(errors, "proof '"+name+"' has invalid status: "+string(proof.Status))
		}
	}

	// Validate modifiers
	for i, mod := range input.Modifiers {
		if !mod.Type.IsValid() {
			errors = append(errors, "modifier["+strconv.Itoa(i)+"] has invalid type: "+string(mod.Type))
		}
		if mod.Justification == "" {
			errors = append(errors, "modifier["+strconv.Itoa(i)+"] missing justification")
		}
		if mod.AppliedBy != "agent" && mod.AppliedBy != "human" {
			errors = append(errors, "modifier["+strconv.Itoa(i)+"] has invalid applied_by: "+mod.AppliedBy)
		}
	}

	return errors
}

// NewColorInputFromProofSet creates a ColorInput from a ProofSet.
// This is a convenience function for integration with the proof collection system.
func NewColorInputFromProofSet(proofSet *ProofSet, sessionType, complexity string, openQuestions []string) ColorInput {
	proofs := make(map[string]ColorProof)

	if proofSet.Tests != nil {
		proofs["tests"] = ColorProof{
			Status:       proofSet.Tests.Status,
			EvidencePath: proofSet.Tests.EvidencePath,
			Summary:      proofSet.Tests.Summary,
			ExitCode:     &proofSet.Tests.ExitCode,
			Timestamp:    &proofSet.Tests.Timestamp,
		}
	}
	if proofSet.Build != nil {
		proofs["build"] = ColorProof{
			Status:       proofSet.Build.Status,
			EvidencePath: proofSet.Build.EvidencePath,
			Summary:      proofSet.Build.Summary,
			ExitCode:     &proofSet.Build.ExitCode,
			Timestamp:    &proofSet.Build.Timestamp,
		}
	}
	if proofSet.Lint != nil {
		proofs["lint"] = ColorProof{
			Status:       proofSet.Lint.Status,
			EvidencePath: proofSet.Lint.EvidencePath,
			Summary:      proofSet.Lint.Summary,
			ExitCode:     &proofSet.Lint.ExitCode,
			Timestamp:    &proofSet.Lint.Timestamp,
		}
	}
	if proofSet.Adversarial != nil {
		proofs["adversarial"] = ColorProof{
			Status:       proofSet.Adversarial.Status,
			EvidencePath: proofSet.Adversarial.EvidencePath,
			Summary:      proofSet.Adversarial.Summary,
			ExitCode:     &proofSet.Adversarial.ExitCode,
			Timestamp:    &proofSet.Adversarial.Timestamp,
		}
	}
	if proofSet.Integration != nil {
		proofs["integration"] = ColorProof{
			Status:       proofSet.Integration.Status,
			EvidencePath: proofSet.Integration.EvidencePath,
			Summary:      proofSet.Integration.Summary,
			ExitCode:     &proofSet.Integration.ExitCode,
			Timestamp:    &proofSet.Integration.Timestamp,
		}
	}

	return ColorInput{
		SessionType:   sessionType,
		Complexity:    complexity,
		Proofs:        proofs,
		OpenQuestions: openQuestions,
	}
}
