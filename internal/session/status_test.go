package session

import "testing"

func TestNormalizeStatus_ValidPassthrough(t *testing.T) {
	tests := []struct {
		input string
		want  Status
	}{
		{"ACTIVE", StatusActive},
		{"PARKED", StatusParked},
		{"ARCHIVED", StatusArchived},
		{"NONE", StatusNone},
	}

	for _, tt := range tests {
		got := NormalizeStatus(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeStatus_AliasMapping(t *testing.T) {
	tests := []struct {
		input string
		want  Status
	}{
		{"COMPLETE", StatusArchived},
		{"COMPLETED", StatusArchived},
		{"complete", StatusArchived},   // case-insensitive
		{"Completed", StatusArchived},  // mixed case
	}

	for _, tt := range tests {
		got := NormalizeStatus(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeStatus_UnknownPassthrough(t *testing.T) {
	// Unknown values pass through as-is (uppercased) for downstream validation
	got := NormalizeStatus("BOGUS")
	if got != Status("BOGUS") {
		t.Errorf("NormalizeStatus(\"BOGUS\") = %q, want %q", got, "BOGUS")
	}
}

func TestNormalizeStatus_WhitespaceAndCase(t *testing.T) {
	tests := []struct {
		input string
		want  Status
	}{
		{"  ACTIVE  ", StatusActive},
		{"active", StatusActive},
		{"Active", StatusActive},
		{"  parked ", StatusParked},
	}

	for _, tt := range tests {
		got := NormalizeStatus(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
