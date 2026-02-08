package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotateSessionContext_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := RotateSessionContext(tmpDir, 200, 80)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if result.Rotated {
		t.Error("expected Rotated=false for missing file")
	}
}

func TestRotateSessionContext_BelowThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	sessionContextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create a small file (10 lines)
	content := `---
session_id: test-123
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
Line 1
Line 2
`
	if err := os.WriteFile(sessionContextPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := RotateSessionContext(tmpDir, 200, 80)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Rotated {
		t.Error("expected Rotated=false for file below threshold")
	}

	// Verify original file unchanged
	afterContent, _ := os.ReadFile(sessionContextPath)
	if string(afterContent) != content {
		t.Error("file was modified when it shouldn't have been")
	}
}

func TestRotateSessionContext_ExceedsThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	sessionContextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create frontmatter (8 lines) + 100 body lines = 108 total
	var b strings.Builder
	b.WriteString(`---
session_id: test-123
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
`)
	for i := 1; i <= 100; i++ {
		b.WriteString("Body line ")
		b.WriteString(string(rune('0' + (i % 10))))
		b.WriteString("\n")
	}

	originalContent := b.String()
	if err := os.WriteFile(sessionContextPath, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Rotate with maxLines=50, keepLines=20
	result, err := RotateSessionContext(tmpDir, 50, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Rotated {
		t.Error("expected Rotated=true")
	}
	if result.ArchivedLines != 80 {
		t.Errorf("expected ArchivedLines=80, got %d", result.ArchivedLines)
	}
	if result.KeptLines != 20 {
		t.Errorf("expected KeptLines=20, got %d", result.KeptLines)
	}

	// Verify archive was created
	archivePath := filepath.Join(tmpDir, "SESSION_CONTEXT.archived.md")
	archiveContent, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("archive file not created: %v", err)
	}

	archiveStr := string(archiveContent)
	if !strings.Contains(archiveStr, "<!-- Archived at") {
		t.Error("archive missing timestamp header")
	}
	if !strings.Contains(archiveStr, "Body line") {
		t.Error("archive missing body content")
	}

	// Verify rotated file has frontmatter + last 20 lines
	rotatedContent, err := os.ReadFile(sessionContextPath)
	if err != nil {
		t.Fatal(err)
	}

	rotatedStr := string(rotatedContent)
	if !strings.HasPrefix(rotatedStr, "---\n") {
		t.Error("frontmatter missing after rotation")
	}

	// Count body lines in rotated file
	parts := strings.Split(rotatedStr, "---\n")
	if len(parts) < 3 {
		t.Fatal("invalid format after rotation")
	}
	bodyAfter := parts[2]
	bodyLinesAfter := strings.Split(strings.TrimSpace(bodyAfter), "\n")
	if len(bodyLinesAfter) != 20 {
		t.Errorf("expected 20 body lines after rotation, got %d", len(bodyLinesAfter))
	}

	// Verify we kept the LAST 20 lines (should contain "Body line" entries)
	if !strings.Contains(bodyAfter, "Body line") {
		t.Error("rotated body missing expected content")
	}
}

func TestRotateSessionContext_PreservesFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	sessionContextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create file with specific frontmatter values
	var b strings.Builder
	b.WriteString(`---
session_id: session-20260208-abc123
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: Build PreCompact hook
complexity: moderate
active_rite: forge
current_phase: implementation
rite: forge
---
`)
	for i := 1; i <= 100; i++ {
		b.WriteString("Test line\n")
	}

	if err := os.WriteFile(sessionContextPath, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}

	// Rotate
	_, err := RotateSessionContext(tmpDir, 50, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse rotated context
	ctx, err := LoadContext(sessionContextPath)
	if err != nil {
		t.Fatalf("failed to parse rotated context: %v", err)
	}

	// Verify frontmatter fields preserved
	if ctx.SessionID != "session-20260208-abc123" {
		t.Errorf("session_id not preserved: got %s", ctx.SessionID)
	}
	if ctx.Initiative != "Build PreCompact hook" {
		t.Errorf("initiative not preserved: got %s", ctx.Initiative)
	}
	if ctx.Complexity != "moderate" {
		t.Errorf("complexity not preserved: got %s", ctx.Complexity)
	}
	if ctx.ActiveRite != "forge" {
		t.Errorf("active_rite not preserved: got %s", ctx.ActiveRite)
	}
	if ctx.CurrentPhase != "implementation" {
		t.Errorf("current_phase not preserved: got %s", ctx.CurrentPhase)
	}
}

func TestRotateSessionContext_Idempotency(t *testing.T) {
	tmpDir := t.TempDir()
	sessionContextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create file that will be rotated to just under threshold
	var b strings.Builder
	b.WriteString(`---
session_id: test-123
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
`)
	for i := 1; i <= 50; i++ {
		b.WriteString("Test line\n")
	}

	if err := os.WriteFile(sessionContextPath, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}

	// First rotation (should rotate)
	result1, err := RotateSessionContext(tmpDir, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !result1.Rotated {
		t.Error("first rotation should have occurred")
	}

	// Read result
	content1, _ := os.ReadFile(sessionContextPath)

	// Second rotation (should NOT rotate - file is now small)
	result2, err := RotateSessionContext(tmpDir, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result2.Rotated {
		t.Error("second rotation should NOT have occurred")
	}

	// Verify file unchanged after second rotation
	content2, _ := os.ReadFile(sessionContextPath)
	if string(content1) != string(content2) {
		t.Error("file changed on second rotation attempt")
	}
}

func TestRotateSessionContext_EmptyBody(t *testing.T) {
	tmpDir := t.TempDir()
	sessionContextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create file with only frontmatter
	content := `---
session_id: test-123
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
`
	if err := os.WriteFile(sessionContextPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := RotateSessionContext(tmpDir, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Rotated {
		t.Error("expected no rotation for empty body")
	}
}

func TestRotateSessionContext_MultipleArchives(t *testing.T) {
	tmpDir := t.TempDir()
	sessionContextPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create and rotate twice
	for round := 1; round <= 2; round++ {
		var b strings.Builder
		b.WriteString(`---
session_id: test-123
status: ACTIVE
created_at: 2026-02-08T10:00:00Z
initiative: test
complexity: simple
active_rite: forge
current_phase: requirements
---
`)
		for i := 1; i <= 100; i++ {
			b.WriteString("Round ")
			b.WriteString(string(rune('0' + round)))
			b.WriteString(" line\n")
		}

		if err := os.WriteFile(sessionContextPath, []byte(b.String()), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := RotateSessionContext(tmpDir, 50, 10)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify archive has both rounds
	archivePath := filepath.Join(tmpDir, "SESSION_CONTEXT.archived.md")
	archiveContent, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	archiveStr := string(archiveContent)
	archiveCount := strings.Count(archiveStr, "<!-- Archived at")
	if archiveCount != 2 {
		t.Errorf("expected 2 archive headers, got %d", archiveCount)
	}

	if !strings.Contains(archiveStr, "Round 1 line") {
		t.Error("archive missing first round content")
	}
	if !strings.Contains(archiveStr, "Round 2 line") {
		t.Error("archive missing second round content")
	}
}
