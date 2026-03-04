package session

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
)

// Default rotation thresholds.
const (
	DefaultMaxLines  = 200 // Rotate when SESSION_CONTEXT exceeds this
	DefaultKeepLines = 80  // Keep this many lines of body after rotation
)

// RotationResult describes the outcome of a rotation operation.
type RotationResult struct {
	Rotated       bool // True if rotation occurred
	ArchivedLines int  // Number of lines archived
	KeptLines     int  // Number of body lines kept
}

// RotateSessionContext rotates SESSION_CONTEXT.md when it exceeds maxLines.
// Preserves YAML frontmatter, archives body to SESSION_CONTEXT.archived.md,
// and keeps the most recent keepLines of body content.
//
// Algorithm:
// 1. Load SESSION_CONTEXT.md and count total lines
// 2. If total lines <= maxLines, return early (no rotation)
// 3. Preserve YAML frontmatter (between --- delimiters)
// 4. Archive full body to SESSION_CONTEXT.archived.md with timestamp header
// 5. Keep last keepLines of body as new body
// 6. Write back frontmatter + trimmed body atomically (temp file + rename)
//
// Returns RotationResult describing what happened.
func RotateSessionContext(sessionDir string, maxLines int, keepLines int) (*RotationResult, error) {
	sessionContextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

	// Check if file exists
	if _, err := os.Stat(sessionContextPath); os.IsNotExist(err) {
		// No file to rotate
		return &RotationResult{Rotated: false}, nil
	}

	// Read the file
	content, err := os.ReadFile(sessionContextPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read SESSION_CONTEXT.md", err)
	}

	// Count total lines
	totalLines := countLines(string(content))
	if totalLines <= maxLines {
		// No rotation needed
		return &RotationResult{Rotated: false}, nil
	}

	// Parse frontmatter and body
	frontmatter, body, err := splitFrontmatterAndBody(string(content))
	if err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "failed to parse frontmatter", err)
	}

	bodyLines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	if len(bodyLines) == 1 && bodyLines[0] == "" {
		bodyLines = []string{}
	}

	// If body is already small enough, no rotation needed
	if len(bodyLines) <= keepLines {
		return &RotationResult{Rotated: false}, nil
	}

	// Archive the full body
	archivePath := filepath.Join(sessionDir, "SESSION_CONTEXT.archived.md")
	if err := appendToArchive(archivePath, body); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to archive body", err)
	}

	// Keep only the last keepLines of body
	keptBodyLines := bodyLines
	if len(bodyLines) > keepLines {
		keptBodyLines = bodyLines[len(bodyLines)-keepLines:]
	}
	newBody := strings.Join(keptBodyLines, "\n")
	if newBody != "" {
		newBody = newBody + "\n"
	}

	// Reconstruct the full content
	var newContent strings.Builder
	newContent.WriteString(frontmatter)
	newContent.WriteString(newBody)

	// Write atomically (temp file + rename)
	if err := fileutil.AtomicWriteFile(sessionContextPath, []byte(newContent.String()), 0644); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write rotated context", err)
	}

	archivedLines := len(bodyLines) - len(keptBodyLines)
	return &RotationResult{
		Rotated:       true,
		ArchivedLines: archivedLines,
		KeptLines:     len(keptBodyLines),
	}, nil
}

// countLines counts the number of lines in a string.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	count := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		count++
	}
	return count
}

// splitFrontmatterAndBody splits a SESSION_CONTEXT.md file into frontmatter and body.
// Returns the frontmatter (including --- delimiters and trailing newline) and body.
func splitFrontmatterAndBody(content string) (frontmatter string, body string, err error) {
	// Must start with ---
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return "", "", fmt.Errorf("no YAML frontmatter found")
	}

	// Find closing delimiter
	endIdx := strings.Index(content[4:], "\n---")
	if endIdx == -1 {
		endIdx = strings.Index(content[4:], "\r\n---")
	}
	if endIdx == -1 {
		return "", "", fmt.Errorf("unclosed YAML frontmatter")
	}

	// frontmatter includes --- delimiters and the newline after closing ---
	frontmatterEnd := endIdx + 4 + 4 // Position after "\n---"
	scanner := bufio.NewScanner(strings.NewReader(content[frontmatterEnd:]))
	if scanner.Scan() {
		// Skip the newline after closing ---
		frontmatterEnd += len(scanner.Text()) + 1
	}

	frontmatter = content[:frontmatterEnd]
	if len(content) > frontmatterEnd {
		body = content[frontmatterEnd:]
	}

	return frontmatter, body, nil
}

// appendToArchive appends content to the archive file with a timestamp header.
func appendToArchive(archivePath string, content string) error {
	f, err := os.OpenFile(archivePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	header := fmt.Sprintf("\n<!-- Archived at %s -->\n\n", timestamp)

	// Check if file is empty (new file)
	info, err := f.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		// New file, skip leading newline
		header = fmt.Sprintf("<!-- Archived at %s -->\n\n", timestamp)
	}

	if _, err := f.WriteString(header); err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		return err
	}
	if !strings.HasSuffix(content, "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}

	return nil
}

