package checksum

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "hello",
			content: "hello",
			want:    "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:    "empty string",
			content: "",
			want:    "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:    "hello world",
			content: "hello world",
			want:    "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Content(tt.content)
			if got != tt.want {
				t.Errorf("Content(%q) = %q, want %q", tt.content, got, tt.want)
			}
		})
	}
}

func TestBytes(t *testing.T) {
	t.Parallel()
	got := Bytes([]byte("hello world"))
	want := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if got != want {
		t.Errorf("Bytes(hello world) = %q, want %q", got, want)
	}
}

func TestContent_HasPrefix(t *testing.T) {
	t.Parallel()
	got := Content("test")
	if !strings.HasPrefix(got, Prefix) {
		t.Errorf("Content() = %q, missing prefix %q", got, Prefix)
	}
}

func TestBytes_HasPrefix(t *testing.T) {
	t.Parallel()
	got := Bytes([]byte("test"))
	if !strings.HasPrefix(got, Prefix) {
		t.Errorf("Bytes() = %q, missing prefix %q", got, Prefix)
	}
}

func TestContent_MatchesBytes(t *testing.T) {
	t.Parallel()
	input := "test content"
	fromContent := Content(input)
	fromBytes := Bytes([]byte(input))
	if fromContent != fromBytes {
		t.Errorf("Content and Bytes disagree: %q vs %q", fromContent, fromBytes)
	}
}

func TestFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	got, err := File(testFile)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}

	want := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if got != want {
		t.Errorf("File() = %q, want %q", got, want)
	}
}

func TestFile_NonExistent(t *testing.T) {
	t.Parallel()
	got, err := File("/nonexistent/path/file.txt")
	if err != nil {
		t.Fatalf("File() error = %v, want nil for nonexistent", err)
	}
	if got != "" {
		t.Errorf("File() = %q, want empty for nonexistent", got)
	}
}

func TestFile_MatchesContent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "consistent hash test"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fromFile, err := File(testFile)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}

	fromBytes := Bytes([]byte(content))
	if fromFile != fromBytes {
		t.Errorf("File and Bytes disagree: %q vs %q", fromFile, fromBytes)
	}
}

func TestDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("aaa"), 0644); err != nil {
		t.Fatalf("failed to write a.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "sub", "b.txt"), []byte("bbb"), 0644); err != nil {
		t.Fatalf("failed to write b.txt: %v", err)
	}

	got, err := Dir(tmpDir)
	if err != nil {
		t.Fatalf("Dir() error = %v", err)
	}

	if !strings.HasPrefix(got, Prefix) {
		t.Errorf("Dir() = %q, missing prefix %q", got, Prefix)
	}

	// Deterministic: same directory should produce same hash
	got2, err := Dir(tmpDir)
	if err != nil {
		t.Fatalf("Dir() second call error = %v", err)
	}
	if got != got2 {
		t.Errorf("Dir() not deterministic: %q vs %q", got, got2)
	}
}
