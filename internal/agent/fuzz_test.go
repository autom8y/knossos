package agent

import "testing"

func FuzzParseAgentFrontmatter(f *testing.F) {
	f.Add([]byte("---\nname: test\ndescription: desc\n---\nbody"))
	f.Add([]byte("---\n---\n"))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, b []byte) {
		_, _ = ParseAgentFrontmatter(b) // must not panic
	})
}
