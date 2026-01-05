package sails

import (
	"reflect"
	"testing"
)

func TestIsValidComplexity(t *testing.T) {
	tests := []struct {
		complexity Complexity
		want       bool
	}{
		{ComplexityPatch, true},
		{ComplexityModule, true},
		{ComplexityService, true},
		{ComplexitySystem, true},
		{ComplexityInitiative, true},
		{ComplexityMigration, true},
		{Complexity("INVALID"), false},
		{Complexity(""), false},
		{Complexity("patch"), false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(string(tt.complexity), func(t *testing.T) {
			if got := IsValidComplexity(tt.complexity); got != tt.want {
				t.Errorf("IsValidComplexity(%q) = %v, want %v", tt.complexity, got, tt.want)
			}
		})
	}
}

func TestGetThresholdMatrix_PATCH(t *testing.T) {
	matrix := GetThresholdMatrix(ComplexityPatch)

	// PATCH: tests, build, lint required; adversarial, integration optional
	if matrix.Tests != ProofRequired {
		t.Errorf("PATCH.Tests = %v, want %v", matrix.Tests, ProofRequired)
	}
	if matrix.Build != ProofRequired {
		t.Errorf("PATCH.Build = %v, want %v", matrix.Build, ProofRequired)
	}
	if matrix.Lint != ProofRequired {
		t.Errorf("PATCH.Lint = %v, want %v", matrix.Lint, ProofRequired)
	}
	if matrix.Adversarial != ProofOptional {
		t.Errorf("PATCH.Adversarial = %v, want %v", matrix.Adversarial, ProofOptional)
	}
	if matrix.Integration != ProofOptional {
		t.Errorf("PATCH.Integration = %v, want %v", matrix.Integration, ProofOptional)
	}
}

func TestGetThresholdMatrix_MODULE(t *testing.T) {
	matrix := GetThresholdMatrix(ComplexityModule)

	// MODULE: same as PATCH - tests, build, lint required
	if matrix.Tests != ProofRequired {
		t.Errorf("MODULE.Tests = %v, want %v", matrix.Tests, ProofRequired)
	}
	if matrix.Build != ProofRequired {
		t.Errorf("MODULE.Build = %v, want %v", matrix.Build, ProofRequired)
	}
	if matrix.Lint != ProofRequired {
		t.Errorf("MODULE.Lint = %v, want %v", matrix.Lint, ProofRequired)
	}
	if matrix.Adversarial != ProofOptional {
		t.Errorf("MODULE.Adversarial = %v, want %v", matrix.Adversarial, ProofOptional)
	}
	if matrix.Integration != ProofOptional {
		t.Errorf("MODULE.Integration = %v, want %v", matrix.Integration, ProofOptional)
	}
}

func TestGetThresholdMatrix_SERVICE(t *testing.T) {
	matrix := GetThresholdMatrix(ComplexityService)

	// SERVICE: tests, build, lint required; adversarial, integration recommended
	if matrix.Tests != ProofRequired {
		t.Errorf("SERVICE.Tests = %v, want %v", matrix.Tests, ProofRequired)
	}
	if matrix.Build != ProofRequired {
		t.Errorf("SERVICE.Build = %v, want %v", matrix.Build, ProofRequired)
	}
	if matrix.Lint != ProofRequired {
		t.Errorf("SERVICE.Lint = %v, want %v", matrix.Lint, ProofRequired)
	}
	if matrix.Adversarial != ProofRecommended {
		t.Errorf("SERVICE.Adversarial = %v, want %v", matrix.Adversarial, ProofRecommended)
	}
	if matrix.Integration != ProofRecommended {
		t.Errorf("SERVICE.Integration = %v, want %v", matrix.Integration, ProofRecommended)
	}
}

func TestGetThresholdMatrix_SYSTEM_Alias(t *testing.T) {
	// SYSTEM should be identical to SERVICE (alias for schema compatibility)
	serviceMatrix := GetThresholdMatrix(ComplexityService)
	systemMatrix := GetThresholdMatrix(ComplexitySystem)

	if !reflect.DeepEqual(serviceMatrix, systemMatrix) {
		t.Errorf("SYSTEM matrix differs from SERVICE: got %+v, want %+v", systemMatrix, serviceMatrix)
	}
}

func TestGetThresholdMatrix_INITIATIVE(t *testing.T) {
	matrix := GetThresholdMatrix(ComplexityInitiative)

	// INITIATIVE: all required
	if matrix.Tests != ProofRequired {
		t.Errorf("INITIATIVE.Tests = %v, want %v", matrix.Tests, ProofRequired)
	}
	if matrix.Build != ProofRequired {
		t.Errorf("INITIATIVE.Build = %v, want %v", matrix.Build, ProofRequired)
	}
	if matrix.Lint != ProofRequired {
		t.Errorf("INITIATIVE.Lint = %v, want %v", matrix.Lint, ProofRequired)
	}
	if matrix.Adversarial != ProofRequired {
		t.Errorf("INITIATIVE.Adversarial = %v, want %v", matrix.Adversarial, ProofRequired)
	}
	if matrix.Integration != ProofRequired {
		t.Errorf("INITIATIVE.Integration = %v, want %v", matrix.Integration, ProofRequired)
	}
}

func TestGetThresholdMatrix_MIGRATION(t *testing.T) {
	matrix := GetThresholdMatrix(ComplexityMigration)

	// MIGRATION: all required (same as INITIATIVE)
	if matrix.Tests != ProofRequired {
		t.Errorf("MIGRATION.Tests = %v, want %v", matrix.Tests, ProofRequired)
	}
	if matrix.Build != ProofRequired {
		t.Errorf("MIGRATION.Build = %v, want %v", matrix.Build, ProofRequired)
	}
	if matrix.Lint != ProofRequired {
		t.Errorf("MIGRATION.Lint = %v, want %v", matrix.Lint, ProofRequired)
	}
	if matrix.Adversarial != ProofRequired {
		t.Errorf("MIGRATION.Adversarial = %v, want %v", matrix.Adversarial, ProofRequired)
	}
	if matrix.Integration != ProofRequired {
		t.Errorf("MIGRATION.Integration = %v, want %v", matrix.Integration, ProofRequired)
	}
}

func TestGetThresholdMatrix_UnknownDefaultsToStrictest(t *testing.T) {
	matrix := GetThresholdMatrix(Complexity("UNKNOWN"))

	// Unknown complexity defaults to strictest (all required)
	if matrix.Tests != ProofRequired {
		t.Errorf("UNKNOWN.Tests = %v, want %v", matrix.Tests, ProofRequired)
	}
	if matrix.Build != ProofRequired {
		t.Errorf("UNKNOWN.Build = %v, want %v", matrix.Build, ProofRequired)
	}
	if matrix.Lint != ProofRequired {
		t.Errorf("UNKNOWN.Lint = %v, want %v", matrix.Lint, ProofRequired)
	}
	if matrix.Adversarial != ProofRequired {
		t.Errorf("UNKNOWN.Adversarial = %v, want %v", matrix.Adversarial, ProofRequired)
	}
	if matrix.Integration != ProofRequired {
		t.Errorf("UNKNOWN.Integration = %v, want %v", matrix.Integration, ProofRequired)
	}
}

func TestGetRequiredProofNames_PATCH(t *testing.T) {
	names := GetRequiredProofNames(ComplexityPatch)
	expected := []string{"tests", "build", "lint"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetRequiredProofNames(PATCH) = %v, want %v", names, expected)
	}
}

func TestGetRequiredProofNames_MODULE(t *testing.T) {
	names := GetRequiredProofNames(ComplexityModule)
	expected := []string{"tests", "build", "lint"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetRequiredProofNames(MODULE) = %v, want %v", names, expected)
	}
}

func TestGetRequiredProofNames_SERVICE(t *testing.T) {
	// SERVICE has required tests/build/lint, but adversarial/integration are only recommended
	names := GetRequiredProofNames(ComplexityService)
	expected := []string{"tests", "build", "lint"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetRequiredProofNames(SERVICE) = %v, want %v", names, expected)
	}
}

func TestGetRequiredProofNames_INITIATIVE(t *testing.T) {
	names := GetRequiredProofNames(ComplexityInitiative)
	expected := []string{"tests", "build", "lint", "adversarial", "integration"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetRequiredProofNames(INITIATIVE) = %v, want %v", names, expected)
	}
}

func TestGetRequiredProofNames_MIGRATION(t *testing.T) {
	names := GetRequiredProofNames(ComplexityMigration)
	expected := []string{"tests", "build", "lint", "adversarial", "integration"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetRequiredProofNames(MIGRATION) = %v, want %v", names, expected)
	}
}

func TestGetRequiredProofNames_UnknownDefaultsToStrictest(t *testing.T) {
	names := GetRequiredProofNames(Complexity("UNKNOWN"))
	expected := []string{"tests", "build", "lint", "adversarial", "integration"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetRequiredProofNames(UNKNOWN) = %v, want %v", names, expected)
	}
}

func TestIsProofRequired(t *testing.T) {
	tests := []struct {
		complexity Complexity
		proofName  string
		want       bool
	}{
		// PATCH - only tests, build, lint required
		{ComplexityPatch, "tests", true},
		{ComplexityPatch, "build", true},
		{ComplexityPatch, "lint", true},
		{ComplexityPatch, "adversarial", false},
		{ComplexityPatch, "integration", false},

		// MODULE - same as PATCH
		{ComplexityModule, "tests", true},
		{ComplexityModule, "build", true},
		{ComplexityModule, "lint", true},
		{ComplexityModule, "adversarial", false},
		{ComplexityModule, "integration", false},

		// SERVICE - adversarial/integration are recommended, not required
		{ComplexityService, "tests", true},
		{ComplexityService, "build", true},
		{ComplexityService, "lint", true},
		{ComplexityService, "adversarial", false},
		{ComplexityService, "integration", false},

		// INITIATIVE - all required
		{ComplexityInitiative, "tests", true},
		{ComplexityInitiative, "build", true},
		{ComplexityInitiative, "lint", true},
		{ComplexityInitiative, "adversarial", true},
		{ComplexityInitiative, "integration", true},

		// MIGRATION - all required
		{ComplexityMigration, "tests", true},
		{ComplexityMigration, "build", true},
		{ComplexityMigration, "lint", true},
		{ComplexityMigration, "adversarial", true},
		{ComplexityMigration, "integration", true},

		// Unknown proof name
		{ComplexityPatch, "unknown", false},
		{ComplexityInitiative, "unknown", false},
	}

	for _, tt := range tests {
		name := string(tt.complexity) + "_" + tt.proofName
		t.Run(name, func(t *testing.T) {
			if got := IsProofRequired(tt.complexity, tt.proofName); got != tt.want {
				t.Errorf("IsProofRequired(%s, %s) = %v, want %v",
					tt.complexity, tt.proofName, got, tt.want)
			}
		})
	}
}

func TestIsProofRecommended(t *testing.T) {
	tests := []struct {
		complexity Complexity
		proofName  string
		want       bool
	}{
		// PATCH - nothing recommended (required or optional)
		{ComplexityPatch, "tests", false},
		{ComplexityPatch, "adversarial", false},
		{ComplexityPatch, "integration", false},

		// SERVICE - adversarial/integration recommended
		{ComplexityService, "tests", false}, // required, not recommended
		{ComplexityService, "adversarial", true},
		{ComplexityService, "integration", true},

		// INITIATIVE - nothing recommended (all required)
		{ComplexityInitiative, "adversarial", false},
		{ComplexityInitiative, "integration", false},

		// Unknown proof name
		{ComplexityService, "unknown", false},
	}

	for _, tt := range tests {
		name := string(tt.complexity) + "_" + tt.proofName
		t.Run(name, func(t *testing.T) {
			if got := IsProofRecommended(tt.complexity, tt.proofName); got != tt.want {
				t.Errorf("IsProofRecommended(%s, %s) = %v, want %v",
					tt.complexity, tt.proofName, got, tt.want)
			}
		})
	}
}

func TestAllProofNames(t *testing.T) {
	names := AllProofNames()
	expected := []string{"tests", "build", "lint", "adversarial", "integration"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("AllProofNames() = %v, want %v", names, expected)
	}

	// Verify modification of returned slice doesn't affect original
	names[0] = "modified"
	names2 := AllProofNames()
	if names2[0] != "tests" {
		t.Error("AllProofNames() returned slice that shares backing array")
	}
}

func TestAllComplexityLevels(t *testing.T) {
	levels := AllComplexityLevels()

	// Should include all main levels (excluding SYSTEM alias)
	expected := []Complexity{
		ComplexityPatch,
		ComplexityModule,
		ComplexityService,
		ComplexityInitiative,
		ComplexityMigration,
	}

	if !reflect.DeepEqual(levels, expected) {
		t.Errorf("AllComplexityLevels() = %v, want %v", levels, expected)
	}
}

// Table-driven test for the complete threshold matrix per TDD Section 4.2
func TestThresholdMatrix_PerTDD(t *testing.T) {
	// Per TDD Section 4.2:
	// | Complexity | tests | build | lint | adversarial | integration |
	// |------------|-------|-------|------|-------------|-------------|
	// | PATCH      | Req   | Req   | Req  | -           | -           |
	// | MODULE     | Req   | Req   | Req  | -           | -           |
	// | SERVICE    | Req   | Req   | Req  | Rec         | Rec         |
	// | INITIATIVE | Req   | Req   | Req  | Req         | Req         |
	// | MIGRATION  | Req   | Req   | Req  | Req         | Req         |

	tests := []struct {
		complexity  Complexity
		tests       ProofRequirement
		build       ProofRequirement
		lint        ProofRequirement
		adversarial ProofRequirement
		integration ProofRequirement
	}{
		{ComplexityPatch, ProofRequired, ProofRequired, ProofRequired, ProofOptional, ProofOptional},
		{ComplexityModule, ProofRequired, ProofRequired, ProofRequired, ProofOptional, ProofOptional},
		{ComplexityService, ProofRequired, ProofRequired, ProofRequired, ProofRecommended, ProofRecommended},
		{ComplexityInitiative, ProofRequired, ProofRequired, ProofRequired, ProofRequired, ProofRequired},
		{ComplexityMigration, ProofRequired, ProofRequired, ProofRequired, ProofRequired, ProofRequired},
	}

	for _, tt := range tests {
		t.Run(string(tt.complexity), func(t *testing.T) {
			matrix := GetThresholdMatrix(tt.complexity)

			if matrix.Tests != tt.tests {
				t.Errorf("%s.Tests = %v, want %v", tt.complexity, matrix.Tests, tt.tests)
			}
			if matrix.Build != tt.build {
				t.Errorf("%s.Build = %v, want %v", tt.complexity, matrix.Build, tt.build)
			}
			if matrix.Lint != tt.lint {
				t.Errorf("%s.Lint = %v, want %v", tt.complexity, matrix.Lint, tt.lint)
			}
			if matrix.Adversarial != tt.adversarial {
				t.Errorf("%s.Adversarial = %v, want %v", tt.complexity, matrix.Adversarial, tt.adversarial)
			}
			if matrix.Integration != tt.integration {
				t.Errorf("%s.Integration = %v, want %v", tt.complexity, matrix.Integration, tt.integration)
			}
		})
	}
}
