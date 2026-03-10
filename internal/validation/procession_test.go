package validation

import (
	"testing"
)

func TestValidateProcessionHandoffFields_Valid(t *testing.T) {
	data := map[string]any{
		"type":           "handoff",
		"procession_id":  "security-remediation-2026-03-10",
		"source_station": "audit",
		"source_rite":    "security",
		"target_station": "assess",
		"target_rite":    "debt-triage",
		"produced_at":    "2026-03-10T10:00:00Z",
		"artifacts": []any{
			map[string]any{"type": "threat-model", "path": ".sos/wip/sr/threat-model.md"},
		},
	}

	issues := ValidateProcessionHandoffFields(data)
	if len(issues) > 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}
}

func TestValidateProcessionHandoffFields_MissingRequired(t *testing.T) {
	// Missing all required fields
	data := map[string]any{}
	issues := ValidateProcessionHandoffFields(data)
	if len(issues) != 8 {
		t.Errorf("expected 8 issues for empty data, got %d: %v", len(issues), issues)
	}
}

func TestValidateProcessionHandoffFields_WrongType(t *testing.T) {
	data := map[string]any{
		"type":           "not-handoff",
		"procession_id":  "security-remediation-2026-03-10",
		"source_station": "audit",
		"source_rite":    "security",
		"target_station": "assess",
		"target_rite":    "debt-triage",
		"produced_at":    "2026-03-10T10:00:00Z",
		"artifacts": []any{
			map[string]any{"type": "threat-model", "path": ".sos/wip/sr/threat-model.md"},
		},
	}

	issues := ValidateProcessionHandoffFields(data)
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for wrong type, got %d: %v", len(issues), issues)
	}
}

func TestValidateProcessionHandoffFields_EmptyArtifacts(t *testing.T) {
	data := map[string]any{
		"type":           "handoff",
		"procession_id":  "security-remediation-2026-03-10",
		"source_station": "audit",
		"source_rite":    "security",
		"target_station": "assess",
		"target_rite":    "debt-triage",
		"produced_at":    "2026-03-10T10:00:00Z",
		"artifacts":      []any{},
	}

	issues := ValidateProcessionHandoffFields(data)
	// Should have 2 issues: empty artifacts field + min items
	hasMinItemsIssue := false
	for _, issue := range issues {
		if issue == "field artifacts must have at least 1 item" {
			hasMinItemsIssue = true
		}
	}
	if !hasMinItemsIssue {
		t.Errorf("expected min items issue for empty artifacts, got: %v", issues)
	}
}

func TestValidateProcessionHandoffFields_PartialMissing(t *testing.T) {
	data := map[string]any{
		"type":           "handoff",
		"procession_id":  "security-remediation-2026-03-10",
		"source_station": "audit",
		// missing source_rite, target_station, target_rite, produced_at, artifacts
	}

	issues := ValidateProcessionHandoffFields(data)
	if len(issues) != 5 {
		t.Errorf("expected 5 issues for partial data, got %d: %v", len(issues), issues)
	}
}
