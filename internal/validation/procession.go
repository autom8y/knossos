package validation

import (
	"fmt"
)

// requiredHandoffFields lists the required fields for a procession handoff artifact.
// Matches the handoff-procession.schema.json required array.
var requiredHandoffFields = []string{
	"type",
	"procession_id",
	"source_station",
	"source_rite",
	"target_station",
	"target_rite",
	"produced_at",
	"artifacts",
}

// ValidateProcessionHandoffFields performs lightweight validation of required
// procession handoff artifact fields. Returns a list of issues (empty = valid).
func ValidateProcessionHandoffFields(data map[string]any) []string {
	var issues []string

	for _, field := range requiredHandoffFields {
		val, ok := data[field]
		if !ok {
			issues = append(issues, fmt.Sprintf("missing required field: %s", field))
			continue
		}
		if isEmpty(val) {
			issues = append(issues, fmt.Sprintf("field %s must not be empty", field))
		}
	}

	// Validate type field value
	if typ, ok := data["type"].(string); ok {
		if typ != "handoff" {
			issues = append(issues, fmt.Sprintf("field type must be \"handoff\", got %q", typ))
		}
	}

	// Validate artifacts is an array with at least one item
	if artifacts, ok := data["artifacts"]; ok {
		count := getItemCount(artifacts)
		if count < 1 {
			issues = append(issues, "field artifacts must have at least 1 item")
		}
	}

	return issues
}
