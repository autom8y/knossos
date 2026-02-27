package mena

import "strings"

// StripMenaExtension removes the .dro or .lego infix from a filename.
// Examples:
//
//	"INDEX.dro.md"      -> "INDEX.md"
//	"INDEX.lego.md"     -> "INDEX.md"
//	"commit.dro.md"     -> "commit.md"
//	"prompting.lego.md" -> "prompting.md"
//	"helper.md"         -> "helper.md"    (no infix, unchanged)
//	"README.md"         -> "README.md"    (no infix, unchanged)
//	"data.json"         -> "data.json"    (no infix, unchanged)
//
// Only the first infix is stripped (handles pathological "foo.dro.dro.md").
func StripMenaExtension(filename string) string {
	if strings.Contains(filename, ".dro.") {
		return strings.Replace(filename, ".dro.", ".", 1)
	}
	if strings.Contains(filename, ".lego.") {
		return strings.Replace(filename, ".lego.", ".", 1)
	}
	return filename
}

// RouteMenaFile determines whether a file routes to commands/ or skills/.
// Returns "commands" or "skills".
func RouteMenaFile(filename string) string {
	menaType := DetectMenaType(filename)
	if menaType == "lego" {
		return "skills"
	}
	return "commands"
}

// DetectMenaType determines content type from file extension convention.
// Files with .dro.md extension are dromena (invokable, project to .claude/commands/).
// Files with .lego.md extension are legomena (reference, project to .claude/skills/).
// Returns "dro" as default for backward compatibility.
func DetectMenaType(filename string) string {
	if strings.Contains(filename, ".dro.") {
		return "dro"
	}
	if strings.Contains(filename, ".lego.") {
		return "lego"
	}
	return "dro" // default for backward compat
}
