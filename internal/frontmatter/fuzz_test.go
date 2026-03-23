package frontmatter

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte("---\nname: test\n---\nbody\n"))
	f.Add([]byte("---\n---\n"))
	f.Add([]byte("no frontmatter"))
	f.Add([]byte(""))
	f.Add([]byte("---\n"))
	f.Fuzz(func(t *testing.T, b []byte) {
		_, _, _ = Parse(b) // must not panic
	})
}
