package mena

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestWalk_MatchesSuffix verifies that Walk invokes the callback only for
// files whose name has the requested suffix.
func TestWalk_MatchesSuffix(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.dro.md"), "dro content")
	writeFile(t, filepath.Join(dir, "b.lego.md"), "lego content")
	writeFile(t, filepath.Join(dir, "c.txt"), "txt content")

	sources := []MenaSource{{Path: dir}}
	var got []string
	Walk(sources, ".dro.md", func(e WalkEntry) {
		got = append(got, e.RelPath)
	})

	if len(got) != 1 || got[0] != "a.dro.md" {
		t.Errorf("expected [a.dro.md], got %v", got)
	}
}

// TestWalk_RecursesIntoSubdirs verifies that Walk descends into subdirectories
// and reports the correct relPath for nested files.
func TestWalk_RecursesIntoSubdirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub", "nested.dro.md"), "nested")

	sources := []MenaSource{{Path: dir}}
	var got []string
	Walk(sources, ".dro.md", func(e WalkEntry) {
		got = append(got, e.RelPath)
	})

	want := filepath.Join("sub", "nested.dro.md")
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

// TestWalk_SkipsNonexistentSource verifies that a source pointing to a
// nonexistent directory is silently skipped without panicking.
func TestWalk_SkipsNonexistentSource(t *testing.T) {
	sources := []MenaSource{{Path: "/nonexistent/path/that/does/not/exist"}}
	var count int
	Walk(sources, ".dro.md", func(e WalkEntry) {
		count++
	})
	if count != 0 {
		t.Errorf("expected 0 callbacks for nonexistent source, got %d", count)
	}
}

// TestWalk_SkipsEmptyPath verifies that a source with Path="" is silently
// skipped without panicking.
func TestWalk_SkipsEmptyPath(t *testing.T) {
	sources := []MenaSource{{Path: ""}}
	var count int
	Walk(sources, ".dro.md", func(e WalkEntry) {
		count++
	})
	if count != 0 {
		t.Errorf("expected 0 callbacks for empty path source, got %d", count)
	}
}

// TestWalk_SkipsEmbeddedSource verifies that sources with IsEmbedded=true are
// skipped entirely.
func TestWalk_SkipsEmbeddedSource(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.dro.md"), "content")

	// Mark the source as embedded — Walk must not process it.
	sources := []MenaSource{{Path: dir, IsEmbedded: true}}
	var count int
	Walk(sources, ".dro.md", func(e WalkEntry) {
		count++
	})
	if count != 0 {
		t.Errorf("expected 0 callbacks for embedded source, got %d", count)
	}
}

// TestWalk_MultipleSources verifies that Walk visits files across multiple
// sources and the callback is invoked once per matching file.
func TestWalk_MultipleSources(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	writeFile(t, filepath.Join(dir1, "first.dro.md"), "first")
	writeFile(t, filepath.Join(dir2, "second.dro.md"), "second")

	sources := []MenaSource{{Path: dir1}, {Path: dir2}}
	var got []string
	Walk(sources, ".dro.md", func(e WalkEntry) {
		got = append(got, e.RelPath)
	})

	if len(got) != 2 {
		t.Errorf("expected 2 callbacks, got %d: %v", len(got), got)
	}
}

// TestWalk_UnreadableFileSkipped verifies that files which cannot be read
// are silently skipped. On macOS, running as root bypasses permissions,
// so this test is skipped on darwin when uid == 0.
func TestWalk_UnreadableFileSkipped(t *testing.T) {
	if runtime.GOOS == "darwin" {
		// macOS root can read 000 files; skip to avoid false pass/fail.
		// Check euid via os.Getuid rather than importing syscall.
		if os.Getuid() == 0 {
			t.Skip("skipping on macOS root: chmod 000 does not restrict root")
		}
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "unreadable.dro.md")
	writeFile(t, path, "secret")
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0644) })

	sources := []MenaSource{{Path: dir}}
	var count int
	Walk(sources, ".dro.md", func(e WalkEntry) {
		count++
	})
	if count != 0 {
		t.Errorf("expected 0 callbacks for unreadable file, got %d", count)
	}
}

// TestWalk_RelPathIsSourceRelative verifies that entry.RelPath is relative to
// the MenaSource.Path directory, not to any project root.
func TestWalk_RelPathIsSourceRelative(t *testing.T) {
	// Simulate: source at <tmpdir>/mena/, file at <tmpdir>/mena/session/park/INDEX.dro.md
	base := t.TempDir()
	menaDir := filepath.Join(base, "mena")
	if err := os.MkdirAll(filepath.Join(menaDir, "session", "park"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(menaDir, "session", "park", "INDEX.dro.md"), "park content")

	sources := []MenaSource{{Path: menaDir}}
	var gotRel []string
	Walk(sources, ".dro.md", func(e WalkEntry) {
		gotRel = append(gotRel, e.RelPath)
	})

	want := filepath.Join("session", "park", "INDEX.dro.md")
	if len(gotRel) != 1 || gotRel[0] != want {
		t.Errorf("expected relPath %q, got %v", want, gotRel)
	}
}

// TestWalk_IndexDirectoryFiles verifies that the suffix filter correctly
// discriminates between files with different suffixes in the same directory.
func TestWalk_IndexDirectoryFiles(t *testing.T) {
	dir := t.TempDir()
	parkDir := filepath.Join(dir, "park")
	if err := os.MkdirAll(parkDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(parkDir, "INDEX.dro.md"), "index dro")
	writeFile(t, filepath.Join(parkDir, "behavior.md"), "behavior")

	sources := []MenaSource{{Path: dir}}

	// Walk with .dro.md: only INDEX.dro.md
	var dro []string
	Walk(sources, ".dro.md", func(e WalkEntry) {
		dro = append(dro, e.RelPath)
	})
	if len(dro) != 1 {
		t.Errorf("expected 1 .dro.md file, got %d: %v", len(dro), dro)
	}

	// Walk with .md: both files
	var md []string
	Walk(sources, ".md", func(e WalkEntry) {
		md = append(md, e.RelPath)
	})
	if len(md) != 2 {
		t.Errorf("expected 2 .md files, got %d: %v", len(md), md)
	}
}

// TestWalk_ReadsContent verifies that entry.Data contains the exact bytes
// written to the file.
func TestWalk_ReadsContent(t *testing.T) {
	dir := t.TempDir()
	want := []byte("# Test skill content\nLine two.")
	writeFile(t, filepath.Join(dir, "skill.lego.md"), string(want))

	sources := []MenaSource{{Path: dir}}
	var got []byte
	Walk(sources, ".lego.md", func(e WalkEntry) {
		got = e.Data
	})

	if string(got) != string(want) {
		t.Errorf("content mismatch\nwant: %q\n got: %q", want, got)
	}
}

// writeFile is a test helper that creates a file with the given content,
// failing the test on any error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%q): %v", path, err)
	}
}
