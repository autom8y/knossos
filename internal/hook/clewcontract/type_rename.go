package clewcontract

// type_rename.go -- v2 to v3 type string rename map.
//
// When reading v2 flat events, RenameV2Type() maps the old type string to the
// new v3 canonical string. This normalizes the type field for consumers that
// only know the v3 type strings.
//
// See SESSION-1 spec Section 5.4 for the canonical rename map.

// v2TypeRenames maps deprecated v2 type strings to their v3 equivalents.
// Only RENAMED types are in this map. Unchanged types are not listed because
// they pass through as-is. See Section 4 cross-reference table for the full picture.
var v2TypeRenames = map[string]string{
	"tool.call":             "tool.invoked",
	"tool.file_change":      "file.modified",
	"tool.artifact_created": "artifact.created",
	"tool.error":            "error.occurred",
	"agent.decision":        "decision.recorded",
	"agent.task_start":      "agent.delegated",
	"agent.task_end":        "agent.completed",
}

// RenameV2Type maps a v2 event type string to its v3 canonical equivalent.
// If the type was renamed, returns the new string. If unchanged, returns the
// input string unmodified. Never returns empty string.
//
// Example:
//
//	RenameV2Type("tool.call")    -> "tool.invoked"
//	RenameV2Type("session.created") -> "session.created"  (unchanged)
func RenameV2Type(v2Type string) string {
	if v3Type, ok := v2TypeRenames[v2Type]; ok {
		return v3Type
	}
	return v2Type
}
