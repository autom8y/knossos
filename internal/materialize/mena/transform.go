package mena

// TransformMenaFilePath applies INDEX.md promotion (dromena) and SKILL.md rename
// (legomena) to a relative file path within a mena entry. Both copyDirFS and
// user-scope walkers call this to ensure identical output structure.
//
// Parameters:
//   - base: the filename (e.g., "INDEX.md", "helper.md")
//   - dir: the directory relative to the mena entry root (e.g., ".", "schemas")
//   - isDro: true for dromena, false for legomena
//
// Returns:
//   - newBase: the potentially renamed filename
//   - promoted: true if the file was promoted to parent level (dromena INDEX.md)
func TransformMenaFilePath(base, dir string, isDro bool) (newBase string, promoted bool) {
	if base == "INDEX.md" && dir == "." {
		if isDro {
			return base, true // caller promotes to parent.md
		}
		return "SKILL.md", false // legomena rename
	}
	return base, false
}
