package know

import "testing"

func FuzzComputeFileDiff(f *testing.F) {
	f.Add([]byte("package a\nfunc Foo() {}"), []byte("package a\nfunc Bar() {}"), "test.go")
	f.Add([]byte(""), []byte("package a"), "empty.go")
	f.Fuzz(func(t *testing.T, old, new []byte, path string) {
		ComputeFileDiff(old, new, path) // must not panic
	})
}
