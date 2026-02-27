package mena

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestExists_FilesystemIndexDir_Lego verifies that a legomena directory
// with INDEX.lego.md resolves to true.
func TestExists_FilesystemIndexDir_Lego(t *testing.T) {
	dir := t.TempDir()
	// Create mena/{name}/INDEX.lego.md
	if err := os.MkdirAll(filepath.Join(dir, "conventions"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "conventions", "INDEX.lego.md"), []byte("# placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	sources := []MenaSource{{Path: dir}}
	if !Exists("conventions", "lego", sources) {
		t.Error("expected Exists to return true for legomena with INDEX.lego.md")
	}
}

// TestExists_FilesystemIndexDir_Dro verifies that a dromena directory
// with INDEX.dro.md resolves to true.
func TestExists_FilesystemIndexDir_Dro(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "go"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go", "INDEX.dro.md"), []byte("# placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	sources := []MenaSource{{Path: dir}}
	if !Exists("go", "dro", sources) {
		t.Error("expected Exists to return true for dromena with INDEX.dro.md")
	}
}

// TestExists_FilesystemIndexDir_GenericIndex verifies that a directory
// with INDEX.md only (no type infix) resolves to true (broadened detection).
func TestExists_FilesystemIndexDir_GenericIndex(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "generic"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "generic", "INDEX.md"), []byte("# placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	sources := []MenaSource{{Path: dir}}
	if !Exists("generic", "lego", sources) {
		t.Error("expected Exists to return true for directory with generic INDEX.md")
	}
}

// TestExists_FilesystemFlatFile_Dro verifies that a flat file pattern
// {name}.dro.md resolves for dromena.
func TestExists_FilesystemFlatFile_Dro(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "park.dro.md"), []byte("# placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	sources := []MenaSource{{Path: dir}}
	if !Exists("park", "dro", sources) {
		t.Error("expected Exists to return true for flat dromena file")
	}
}

// TestExists_FilesystemFlatFile_Lego verifies that a flat file pattern
// {name}.lego.md does NOT resolve for legomena (flat file pattern is dromena-only).
func TestExists_FilesystemFlatFile_Lego(t *testing.T) {
	dir := t.TempDir()
	// Create only a flat .lego.md file, no directory with INDEX
	if err := os.WriteFile(filepath.Join(dir, "skill.lego.md"), []byte("# placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	sources := []MenaSource{{Path: dir}}
	if Exists("skill", "lego", sources) {
		t.Error("expected Exists to return false for flat legomena file (flat pattern is dromena-only)")
	}
}

// TestExists_EmptyPath_Skipped verifies that a source with Path="" is skipped
// without panicking.
func TestExists_EmptyPath_Skipped(t *testing.T) {
	sources := []MenaSource{{Path: ""}}
	result := Exists("anything", "lego", sources)
	if result {
		t.Error("expected Exists to return false for empty path source")
	}
}

// TestExists_NonexistentDir_Skipped verifies that a source with a nonexistent
// path is silently skipped without panicking.
func TestExists_NonexistentDir_Skipped(t *testing.T) {
	sources := []MenaSource{{Path: "/nonexistent/path/that/does/not/exist"}}
	result := Exists("anything", "lego", sources)
	if result {
		t.Error("expected Exists to return false for nonexistent path")
	}
}

// TestExists_MultipleSourcesFirstMatch verifies that when an entry is absent
// from the first source but present in the second source, Exists returns true.
func TestExists_MultipleSourcesFirstMatch(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Entry only in second source
	if err := os.MkdirAll(filepath.Join(dir2, "provider-ref"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "provider-ref", "INDEX.lego.md"), []byte("# placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	sources := []MenaSource{{Path: dir1}, {Path: dir2}}
	if !Exists("provider-ref", "lego", sources) {
		t.Error("expected Exists to return true when entry is in second source")
	}
}

// TestExists_EmbeddedFS_IndexDir verifies that Exists resolves from an
// embedded FS source containing an INDEX file.
func TestExists_EmbeddedFS_IndexDir(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/shared/mena/conventions/INDEX.lego.md": &fstest.MapFile{
			Data: []byte("# placeholder"),
		},
	}

	sources := []MenaSource{
		{
			Fsys:       fsys,
			FsysPath:   "rites/shared/mena",
			IsEmbedded: true,
		},
	}
	if !Exists("conventions", "lego", sources) {
		t.Error("expected Exists to return true for embedded FS with INDEX file")
	}
}

// TestExists_EmbeddedFS_FlatDro verifies that Exists resolves a flat dromena
// file from an embedded FS source.
func TestExists_EmbeddedFS_FlatDro(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/shared/mena/park.dro.md": &fstest.MapFile{
			Data: []byte("# placeholder"),
		},
	}

	sources := []MenaSource{
		{
			Fsys:       fsys,
			FsysPath:   "rites/shared/mena",
			IsEmbedded: true,
		},
	}
	if !Exists("park", "dro", sources) {
		t.Error("expected Exists to return true for embedded FS with flat dromena file")
	}
}
