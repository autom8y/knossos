// Package worktree provides git worktree management for parallel coding sessions.
// It enables isolated filesystem environments for running multiple sessions simultaneously.
package worktree

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"time"
)

// Worktree represents a git worktree with metadata.
type Worktree struct {
	ID         string    `json:"id"`          // wt-YYYYMMDD-HHMMSS-hex
	Name       string    `json:"name"`        // User-friendly name
	Path       string    `json:"path"`        // Filesystem path
	Branch     string    `json:"branch"`      // Git branch (empty if detached)
	Rite       string    `json:"rite"`        // Active rite
	CreatedAt  time.Time `json:"created_at"`  // Creation timestamp
	BaseBranch string    `json:"base_branch"` // Branch it was created from
	FromRef    string    `json:"from_ref"`    // Original ref used to create worktree
	Complexity string    `json:"complexity"`  // Session complexity level
}

// WorktreeMetadata stores the registry of all worktrees.
type WorktreeMetadata struct {
	Worktrees []Worktree `json:"worktrees"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// WorktreeStatus provides detailed status for a worktree.
type WorktreeStatus struct {
	Worktree
	IsDirty         bool   `json:"is_dirty"`          // Has uncommitted changes
	HasUntracked    bool   `json:"has_untracked"`     // Has untracked files
	UntrackedCount  int    `json:"untracked_count"`   // Number of untracked files
	ChangedFiles    int    `json:"changed_files"`     // Number of changed files
	CommitsAhead    int    `json:"commits_ahead"`     // Commits ahead of base branch
	CommitsBehind   int    `json:"commits_behind"`    // Commits behind base branch
	SessionStatus   string `json:"session_status"`    // none, active, parked
	CurrentSession  string `json:"current_session"`   // Session ID if any
	Age             string `json:"age"`               // Human-readable age
}

// CreateOptions specifies options for creating a worktree.
type CreateOptions struct {
	Name       string // User-friendly name
	Rite       string // Rite to activate
	FromRef    string // Git ref to create from (default: HEAD)
	Complexity string // Session complexity level
}

// CleanupOptions specifies options for cleanup operations.
type CleanupOptions struct {
	OlderThan time.Duration // Remove worktrees older than this
	DryRun    bool          // Show what would be removed
	Force     bool          // Force removal even with uncommitted changes
}

// Worktree ID format: wt-YYYYMMDD-HHMMSS-{4-char-hex}
var worktreeIDPattern = regexp.MustCompile(`^wt-[0-9]{8}-[0-9]{6}-[a-f0-9]{4,}$`)

// GenerateWorktreeID generates a new unique worktree ID.
func GenerateWorktreeID() string {
	now := time.Now()
	hex := make([]byte, 2)
	_, _ = rand.Read(hex)
	return fmt.Sprintf("wt-%s-%x",
		now.Format("20060102-150405"),
		hex,
	)
}

// IsValidWorktreeID checks if an ID matches the worktree ID pattern.
func IsValidWorktreeID(id string) bool {
	return worktreeIDPattern.MatchString(id)
}

// ParseWorktreeTimestamp extracts the timestamp from a worktree ID.
// Returns zero time if parsing fails.
func ParseWorktreeTimestamp(id string) time.Time {
	if len(id) < 21 {
		return time.Time{}
	}
	// Extract "YYYYMMDD-HHMMSS" portion
	dateStr := id[3:18] // "20260104-160414"
	t, err := time.Parse("20060102-150405", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// FormatAge returns a human-readable age string.
func FormatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
