package worktree

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
)

// SessionIntegration provides worktree-session context integration.
type SessionIntegration struct {
	rootDir string
}

// NewSessionIntegration creates a new SessionIntegration.
func NewSessionIntegration(rootDir string) *SessionIntegration {
	return &SessionIntegration{rootDir: rootDir}
}

// UpdateSessionWorktree updates the worktree_id field in a session's frontmatter.
func (s *SessionIntegration) UpdateSessionWorktree(sessionPath, worktreeID string) error {
	contextPath := filepath.Join(sessionPath, "SESSION_CONTEXT.md")

	// Read existing content
	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New(errors.CodeFileNotFound, "session context not found: "+contextPath)
		}
		return errors.Wrap(errors.CodeGeneralError, "failed to read session context", err)
	}

	content := string(data)

	// Parse and update frontmatter
	updated, err := updateFrontmatterField(content, "worktree_id", worktreeID)
	if err != nil {
		return err
	}

	// Write back
	if err := os.WriteFile(contextPath, []byte(updated), 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write session context", err)
	}

	return nil
}

// GetActiveWorktree retrieves the worktree_id from a session's frontmatter.
func (s *SessionIntegration) GetActiveWorktree(sessionPath string) (string, error) {
	contextPath := filepath.Join(sessionPath, "SESSION_CONTEXT.md")

	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New(errors.CodeFileNotFound, "session context not found: "+contextPath)
		}
		return "", errors.Wrap(errors.CodeGeneralError, "failed to read session context", err)
	}

	return extractFrontmatterField(string(data), "worktree_id")
}

// FindActiveSession looks for an active session in the given directory.
// Returns the session path and session ID if found.
func (s *SessionIntegration) FindActiveSession(searchPath string) (sessionPath, sessionID string, err error) {
	sessionsDir := filepath.Join(searchPath, ".sos", "sessions")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil // No sessions directory
		}
		return "", "", errors.Wrap(errors.CodeGeneralError, "failed to read sessions directory", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "session-") {
			continue
		}

		sessionDir := filepath.Join(sessionsDir, entry.Name())
		contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

		data, err := os.ReadFile(contextPath)
		if err != nil {
			continue
		}

		status, err := extractFrontmatterField(string(data), "status")
		if err != nil {
			continue
		}

		// Check if session is active (not parked, not wrapped)
		if status != "PARKED" && status != "WRAPPED" && status != "ABANDONED" {
			return sessionDir, entry.Name(), nil
		}
	}

	return "", "", nil
}

// GetSessionsForWorktree finds all sessions associated with a worktree.
func (s *SessionIntegration) GetSessionsForWorktree(searchPath, worktreeID string) ([]string, error) {
	sessionsDir := filepath.Join(searchPath, ".sos", "sessions")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read sessions directory", err)
	}

	var sessions []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "session-") {
			continue
		}

		sessionDir := filepath.Join(sessionsDir, entry.Name())
		wtID, err := s.GetActiveWorktree(sessionDir)
		if err != nil {
			continue
		}

		if wtID == worktreeID {
			sessions = append(sessions, entry.Name())
		}
	}

	return sessions, nil
}

// LinkSessionToWorktree creates a bidirectional link between session and worktree.
func (s *SessionIntegration) LinkSessionToWorktree(sessionPath string, wt *Worktree) error {
	// Update session frontmatter with worktree_id
	if err := s.UpdateSessionWorktree(sessionPath, wt.ID); err != nil {
		return err
	}

	return nil
}

// UnlinkSessionFromWorktree removes the worktree association from a session.
func (s *SessionIntegration) UnlinkSessionFromWorktree(sessionPath string) error {
	contextPath := filepath.Join(sessionPath, "SESSION_CONTEXT.md")

	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No session to unlink
		}
		return errors.Wrap(errors.CodeGeneralError, "failed to read session context", err)
	}

	// Remove worktree_id field
	updated, err := removeFrontmatterField(string(data), "worktree_id")
	if err != nil {
		return err
	}

	if err := os.WriteFile(contextPath, []byte(updated), 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write session context", err)
	}

	return nil
}

// updateFrontmatterField updates or adds a field in YAML frontmatter.
func updateFrontmatterField(content, field, value string) (string, error) {
	lines := strings.Split(content, "\n")

	// Find frontmatter boundaries
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", errors.New(errors.CodeParseError, "invalid frontmatter: missing opening ---")
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return "", errors.New(errors.CodeParseError, "invalid frontmatter: missing closing ---")
	}

	// Look for existing field
	fieldPattern := regexp.MustCompile(`^(\s*)` + regexp.QuoteMeta(field) + `:\s*`)
	fieldFound := false

	for i := 1; i < endIdx; i++ {
		if fieldPattern.MatchString(lines[i]) {
			// Update existing field
			matches := fieldPattern.FindStringSubmatch(lines[i])
			indent := ""
			if len(matches) > 1 {
				indent = matches[1]
			}
			lines[i] = indent + field + ": " + quoteIfNeeded(value)
			fieldFound = true
			break
		}
	}

	// Add new field if not found
	if !fieldFound {
		newLine := field + ": " + quoteIfNeeded(value)
		// Insert before closing ---
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:endIdx]...)
		newLines = append(newLines, newLine)
		newLines = append(newLines, lines[endIdx:]...)
		lines = newLines
	}

	return strings.Join(lines, "\n"), nil
}

// extractFrontmatterField extracts a field value from YAML frontmatter.
func extractFrontmatterField(content, field string) (string, error) {
	lines := strings.Split(content, "\n")

	// Find frontmatter boundaries
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", nil // No frontmatter
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return "", nil // Invalid frontmatter
	}

	// Look for field
	fieldPattern := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(field) + `:\s*(.*)$`)

	for i := 1; i < endIdx; i++ {
		matches := fieldPattern.FindStringSubmatch(lines[i])
		if len(matches) > 1 {
			value := strings.TrimSpace(matches[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			return value, nil
		}
	}

	return "", nil // Field not found
}

// removeFrontmatterField removes a field from YAML frontmatter.
func removeFrontmatterField(content, field string) (string, error) {
	lines := strings.Split(content, "\n")

	// Find frontmatter boundaries
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return content, nil // No frontmatter
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return content, nil // Invalid frontmatter
	}

	// Remove field line
	fieldPattern := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(field) + `:`)
	newLines := make([]string, 0, len(lines))

	for i, line := range lines {
		if i > 0 && i < endIdx && fieldPattern.MatchString(line) {
			continue // Skip this line
		}
		newLines = append(newLines, line)
	}

	return strings.Join(newLines, "\n"), nil
}

// quoteIfNeeded adds quotes around a value if it contains special characters.
func quoteIfNeeded(value string) string {
	// Quote if empty or contains special characters
	if value == "" {
		return `""`
	}

	needsQuotes := false
	for _, c := range value {
		if c == ':' || c == '#' || c == '[' || c == ']' || c == '{' || c == '}' ||
			c == ',' || c == '&' || c == '*' || c == '!' || c == '|' || c == '>' ||
			c == '\'' || c == '"' || c == '%' || c == '@' || c == '`' {
			needsQuotes = true
			break
		}
	}

	// Quote if starts with special characters
	if len(value) > 0 {
		first := value[0]
		if first == '-' || first == '?' || first == ' ' {
			needsQuotes = true
		}
	}

	if needsQuotes {
		// Escape existing quotes and wrap
		escaped := strings.ReplaceAll(value, `"`, `\"`)
		return `"` + escaped + `"`
	}

	return value
}

// ParseSessionContext reads and parses a session context file.
type SessionContext struct {
	SchemaVersion string `json:"schema_version"`
	SessionID     string `json:"session_id"`
	Status        string `json:"status"`
	WorktreeID    string `json:"worktree_id,omitempty"`
	ActiveRite    string `json:"active_rite,omitempty"`
	Initiative    string `json:"initiative,omitempty"`
	Complexity    string `json:"complexity,omitempty"`
	CurrentPhase  string `json:"current_phase,omitempty"`
}

// ParseSessionContextFile parses a SESSION_CONTEXT.md file.
func ParseSessionContextFile(path string) (*SessionContext, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to open session context", err)
	}
	defer func() { _ = file.Close() }()

	ctx := &SessionContext{}
	scanner := bufio.NewScanner(file)
	inFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break // End of frontmatter
			}
		}

		if !inFrontmatter {
			continue
		}

		// Parse key: value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		switch key {
		case "schema_version":
			ctx.SchemaVersion = value
		case "session_id":
			ctx.SessionID = value
		case "status":
			ctx.Status = value
		case "worktree_id":
			ctx.WorktreeID = value
		case "active_rite", "active_team": // Support both new and legacy names
			ctx.ActiveRite = value
		case "initiative":
			ctx.Initiative = value
		case "complexity":
			ctx.Complexity = value
		case "current_phase":
			ctx.CurrentPhase = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse session context", err)
	}

	return ctx, nil
}
