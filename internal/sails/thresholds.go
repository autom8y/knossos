// Package sails implements the White Sails confidence signaling system
// per Knossos Doctrine v2. White Sails provides typed contracts declaring
// computed confidence levels with explicit proof chains.
package sails

// Complexity represents the complexity tier of a session.
// Higher complexity requires stricter proof requirements.
type Complexity string

const (
	// ComplexityPatch represents small, isolated changes.
	ComplexityPatch Complexity = "PATCH"
	// ComplexityScript represents script-level changes (small automation or tooling).
	ComplexityScript Complexity = "SCRIPT"
	// ComplexityModule represents module-level changes.
	ComplexityModule Complexity = "MODULE"
	// ComplexityService represents service-level changes.
	// Note: TDD uses SERVICE, schema uses SYSTEM - aliased for compatibility.
	ComplexityService Complexity = "SERVICE"
	// ComplexitySystem is an alias for ComplexityService for schema compatibility.
	ComplexitySystem Complexity = "SYSTEM"
	// ComplexityInitiative represents cross-cutting initiative changes.
	ComplexityInitiative Complexity = "INITIATIVE"
	// ComplexityMigration represents data or schema migrations.
	ComplexityMigration Complexity = "MIGRATION"
	// ComplexityPlatform represents platform-level changes (infrastructure, core systems).
	ComplexityPlatform Complexity = "PLATFORM"
)

// IsValidComplexity checks if a complexity value is valid.
func IsValidComplexity(c Complexity) bool {
	switch c {
	case ComplexityPatch, ComplexityScript, ComplexityModule, ComplexityService,
		ComplexitySystem, ComplexityInitiative, ComplexityMigration, ComplexityPlatform:
		return true
	default:
		return false
	}
}

// ProofRequirement represents the requirement level for a proof type.
type ProofRequirement string

const (
	// ProofRequired means the proof must be present and passing for WHITE sails.
	ProofRequired ProofRequirement = "required"
	// ProofRecommended means the proof is encouraged but not required for WHITE.
	ProofRecommended ProofRequirement = "recommended"
	// ProofOptional means the proof is not tracked for this complexity level.
	ProofOptional ProofRequirement = "optional"
)

// ThresholdMatrix defines proof requirements for a complexity level.
type ThresholdMatrix struct {
	Tests       ProofRequirement
	Build       ProofRequirement
	Lint        ProofRequirement
	Adversarial ProofRequirement
	Integration ProofRequirement
}

// thresholds maps complexity levels to their proof requirements.
// Per TDD Section 4.2:
//   - PATCH/SCRIPT/MODULE: tests, build, lint required
//   - SERVICE: adds recommended adversarial, integration
//   - INITIATIVE/MIGRATION: all required
var thresholds = map[Complexity]ThresholdMatrix{
	ComplexityPatch: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofOptional,
		Integration: ProofOptional,
	},
	ComplexityScript: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofOptional,
		Integration: ProofOptional,
	},
	ComplexityModule: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofOptional,
		Integration: ProofOptional,
	},
	ComplexityService: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofRecommended,
		Integration: ProofRecommended,
	},
	ComplexitySystem: { // Alias for SERVICE
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofRecommended,
		Integration: ProofRecommended,
	},
	ComplexityInitiative: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofRequired,
		Integration: ProofRequired,
	},
	ComplexityMigration: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofRequired,
		Integration: ProofRequired,
	},
	ComplexityPlatform: {
		Tests:       ProofRequired,
		Build:       ProofRequired,
		Lint:        ProofRequired,
		Adversarial: ProofRequired,
		Integration: ProofRequired,
	},
}

// strictestThresholds returns the strictest requirements (all required).
// Used as default for unknown complexity levels - when in doubt, require everything.
var strictestThresholds = ThresholdMatrix{
	Tests:       ProofRequired,
	Build:       ProofRequired,
	Lint:        ProofRequired,
	Adversarial: ProofRequired,
	Integration: ProofRequired,
}

// GetThresholdMatrix returns the proof requirements for a complexity level.
// Unknown complexity levels default to strictest requirements (all required).
func GetThresholdMatrix(complexity Complexity) ThresholdMatrix {
	if matrix, ok := thresholds[complexity]; ok{
		return matrix
	}
	return strictestThresholds
}

// GetThresholdMatrixByString returns the proof requirements for a complexity string.
// This is a convenience wrapper for GetThresholdMatrix that accepts a string.
func GetThresholdMatrixByString(complexity string) ThresholdMatrix {
	return GetThresholdMatrix(Complexity(complexity))
}

// proofNames are the canonical names for each proof type.
var proofNames = []string{"tests", "build", "lint", "adversarial", "integration"}

// GetRequiredProofNames returns the names of proofs that are required
// for the given complexity level. Only proofs with ProofRequired status
// are included; recommended and optional proofs are excluded.
func GetRequiredProofNames(complexity Complexity) []string {
	matrix := GetThresholdMatrix(complexity)
	var required []string

	if matrix.Tests == ProofRequired {
		required = append(required, "tests")
	}
	if matrix.Build == ProofRequired {
		required = append(required, "build")
	}
	if matrix.Lint == ProofRequired {
		required = append(required, "lint")
	}
	if matrix.Adversarial == ProofRequired {
		required = append(required, "adversarial")
	}
	if matrix.Integration == ProofRequired {
		required = append(required, "integration")
	}

	return required
}

// IsProofRequired returns true if the specified proof is required for
// the given complexity level. Returns true for unknown complexity levels
// (strictest default).
func IsProofRequired(complexity Complexity, proofName string) bool {
	matrix := GetThresholdMatrix(complexity)

	switch proofName {
	case "tests":
		return matrix.Tests == ProofRequired
	case "build":
		return matrix.Build == ProofRequired
	case "lint":
		return matrix.Lint == ProofRequired
	case "adversarial":
		return matrix.Adversarial == ProofRequired
	case "integration":
		return matrix.Integration == ProofRequired
	default:
		return false
	}
}

// IsProofRecommended returns true if the specified proof is recommended
// (but not required) for the given complexity level.
func IsProofRecommended(complexity Complexity, proofName string) bool {
	matrix := GetThresholdMatrix(complexity)

	switch proofName {
	case "tests":
		return matrix.Tests == ProofRecommended
	case "build":
		return matrix.Build == ProofRecommended
	case "lint":
		return matrix.Lint == ProofRecommended
	case "adversarial":
		return matrix.Adversarial == ProofRecommended
	case "integration":
		return matrix.Integration == ProofRecommended
	default:
		return false
	}
}

// AllProofNames returns all canonical proof names.
func AllProofNames() []string {
	// Return a copy to prevent modification
	names := make([]string, len(proofNames))
	copy(names, proofNames)
	return names
}

// AllComplexityLevels returns all valid complexity levels in order of strictness.
func AllComplexityLevels() []Complexity {
	return []Complexity{
		ComplexityPatch,
		ComplexityModule,
		ComplexityService,
		ComplexityInitiative,
		ComplexityMigration,
	}
}
