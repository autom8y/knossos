package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicWriteFile_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	err := AtomicWriteFile(target, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("AtomicWriteFile() error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("content = %q, want %q", string(content), "hello")
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWriteFile(target, []byte("new"), 0644); err != nil {
		t.Fatalf("AtomicWriteFile() error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want %q", string(content), "new")
	}
}

func TestAtomicWriteFile_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "a", "b", "c", "test.txt")

	err := AtomicWriteFile(target, []byte("deep"), 0644)
	if err != nil {
		t.Fatalf("AtomicWriteFile() error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(content) != "deep" {
		t.Errorf("content = %q, want %q", string(content), "deep")
	}
}

func TestAtomicWriteFile_NoTmpLeftBehind(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	if err := AtomicWriteFile(target, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp") {
			t.Errorf("found leftover tmp file: %s", entry.Name())
		}
	}
}

func TestAtomicWriteFile_PermissionsApplied(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	if err := AtomicWriteFile(target, []byte("content"), 0600); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("permissions = %o, want %o", perm, 0600)
	}
}

func TestWriteIfChanged_SkipsIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(target, []byte("same"), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err := WriteIfChanged(target, []byte("same"), 0644)
	if err != nil {
		t.Fatalf("WriteIfChanged() error: %v", err)
	}
	if changed {
		t.Error("WriteIfChanged() = true for identical content, want false")
	}
}

func TestWriteIfChanged_WritesDifferent(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err := WriteIfChanged(target, []byte("new"), 0644)
	if err != nil {
		t.Fatalf("WriteIfChanged() error: %v", err)
	}
	if !changed {
		t.Error("WriteIfChanged() = false for different content, want true")
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want %q", string(content), "new")
	}
}
