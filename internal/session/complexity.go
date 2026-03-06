package session

// Complexity represents a session complexity level.
type Complexity string

const (
	ComplexityPatch      Complexity = "PATCH"
	ComplexityModule     Complexity = "MODULE"
	ComplexitySystem     Complexity = "SYSTEM"
	ComplexityInitiative Complexity = "INITIATIVE"
	ComplexityMigration  Complexity = "MIGRATION"
)

// IsValidComplexity checks if a complexity value is canonical.
func IsValidComplexity(c string) bool {
	switch Complexity(c) {
	case ComplexityPatch, ComplexityModule, ComplexitySystem,
		ComplexityInitiative, ComplexityMigration:
		return true
	default:
		return false
	}
}
