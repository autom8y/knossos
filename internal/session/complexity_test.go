package session

import "testing"

func TestIsValidComplexity_ValidValues(t *testing.T) {
	valid := []string{"PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION"}
	for _, c := range valid {
		if !IsValidComplexity(c) {
			t.Errorf("IsValidComplexity(%q) = false, want true", c)
		}
	}
}

func TestIsValidComplexity_InvalidValues(t *testing.T) {
	invalid := []string{"SCRIPT", "SERVICE", "PLATFORM", "", "module"}
	for _, c := range invalid {
		if IsValidComplexity(c) {
			t.Errorf("IsValidComplexity(%q) = true, want false", c)
		}
	}
}
