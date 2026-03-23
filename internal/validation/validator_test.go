package validation

import (
	"testing"
)

func TestValidateComplaint_Valid(t *testing.T) {
	t.Parallel()

	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	valid := []byte(`id: COMPLAINT-20260323-120000-drift-detector
filed_by: drift-detector
filed_at: "2026-03-23T12:00:00Z"
title: "tool-fallback drift: grep used via Bash"
severity: low
description: |
  Drift detection hook identified a tool-fallback pattern.
tags:
  - drift
  - tool-fallback
  - auto-filed
status: filed
`)

	if err := v.ValidateComplaint(valid); err != nil {
		t.Errorf("ValidateComplaint(valid): unexpected error: %v", err)
	}
}

func TestValidateComplaint_DeepFile(t *testing.T) {
	t.Parallel()

	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	deepFile := []byte(`id: COMPLAINT-20260311-091500-pythia
filed_by: pythia
filed_at: "2026-03-11T09:15:00Z"
title: "Missing skill for rite navigation"
severity: high
description: |
  No skill exists for cross-rite navigation guidance.
tags:
  - routing
  - skill-gap
status: triaged
zone: behavior
effort_estimate: medium
related_scars:
  - SCAR-005
evidence:
  session_id: session-20260311-012734-9847ff6f
  event_refs:
    - "event-001"
  context: "Agent attempted /consult and received empty response"
suggested_fix: |
  Add a consult skill to the shared mena.
`)

	if err := v.ValidateComplaint(deepFile); err != nil {
		t.Errorf("ValidateComplaint(deep-file): unexpected error: %v", err)
	}
}

func TestValidateComplaint_MissingRequiredField(t *testing.T) {
	t.Parallel()

	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	// Missing 'severity' which is required.
	invalid := []byte(`id: COMPLAINT-20260323-120000-test
filed_by: test
filed_at: "2026-03-23T12:00:00Z"
title: "Test complaint"
description: "Missing severity field"
status: filed
`)

	if err := v.ValidateComplaint(invalid); err == nil {
		t.Error("ValidateComplaint(missing severity): expected error, got nil")
	}
}

func TestValidateComplaint_InvalidSeverity(t *testing.T) {
	t.Parallel()

	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	invalid := []byte(`id: COMPLAINT-20260323-120000-test
filed_by: test
filed_at: "2026-03-23T12:00:00Z"
title: "Test complaint"
severity: extreme
description: "Invalid severity enum value"
status: filed
`)

	if err := v.ValidateComplaint(invalid); err == nil {
		t.Error("ValidateComplaint(invalid severity): expected error, got nil")
	}
}

func TestValidateComplaint_InvalidStatus(t *testing.T) {
	t.Parallel()

	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	invalid := []byte(`id: COMPLAINT-20260323-120000-test
filed_by: test
filed_at: "2026-03-23T12:00:00Z"
title: "Test complaint"
severity: low
description: "Invalid status"
status: pending
`)

	if err := v.ValidateComplaint(invalid); err == nil {
		t.Error("ValidateComplaint(invalid status): expected error, got nil")
	}
}

func TestValidateComplaint_MalformedYAML(t *testing.T) {
	t.Parallel()

	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	if err := v.ValidateComplaint([]byte("not: valid: yaml: {{{")); err == nil {
		t.Error("ValidateComplaint(malformed YAML): expected error, got nil")
	}
}
