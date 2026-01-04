// Package output provides format-aware output printing for Ariadne.
// This file contains sync-domain specific output structures.
package output

import (
	"fmt"
	"strings"
)

// --- Sync Output Structures ---

// SyncStatusOutput represents the sync status command output.
type SyncStatusOutput struct {
	Initialized  bool                `json:"initialized"`
	Remote       string              `json:"remote,omitempty"`
	LastSync     string              `json:"last_sync,omitempty"`
	TrackedPaths []SyncTrackedPath   `json:"tracked_paths"`
	HasConflicts bool                `json:"has_conflicts"`
	Conflicts    []SyncConflictEntry `json:"conflicts,omitempty"`
}

// SyncTrackedPath represents a tracked file's sync status.
type SyncTrackedPath struct {
	Path         string `json:"path"`
	Status       string `json:"status"` // synced, modified, conflict, untracked
	LocalHash    string `json:"local_hash,omitempty"`
	RemoteHash   string `json:"remote_hash,omitempty"`
	BaseHash     string `json:"base_hash,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
}

// SyncConflictEntry represents a conflict in the sync status.
type SyncConflictEntry struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	LocalHash   string `json:"local_hash,omitempty"`
	RemoteHash  string `json:"remote_hash,omitempty"`
	BaseHash    string `json:"base_hash,omitempty"`
}

// Text implements Textable for SyncStatusOutput.
func (s SyncStatusOutput) Text() string {
	if !s.Initialized {
		return "Sync not initialized. Run 'ari sync pull <remote>' to initialize."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Remote: %s\n", s.Remote))
	if s.LastSync != "" {
		b.WriteString(fmt.Sprintf("Last sync: %s\n", s.LastSync))
	}
	b.WriteString("\n")

	if len(s.TrackedPaths) == 0 {
		b.WriteString("No tracked paths\n")
	} else {
		b.WriteString("Tracked paths:\n")
		for _, p := range s.TrackedPaths {
			indicator := statusIndicator(p.Status)
			b.WriteString(fmt.Sprintf("  %s %s\n", indicator, p.Path))
		}
	}

	if s.HasConflicts {
		b.WriteString(fmt.Sprintf("\nConflicts: %d\n", len(s.Conflicts)))
		for _, c := range s.Conflicts {
			b.WriteString(fmt.Sprintf("  ! %s: %s\n", c.Path, c.Description))
		}
	}

	return b.String()
}

// Headers implements Tabular for SyncStatusOutput.
func (s SyncStatusOutput) Headers() []string {
	return []string{"STATUS", "PATH"}
}

// Rows implements Tabular for SyncStatusOutput.
func (s SyncStatusOutput) Rows() [][]string {
	rows := make([][]string, len(s.TrackedPaths))
	for i, p := range s.TrackedPaths {
		rows[i] = []string{statusIndicator(p.Status), p.Path}
	}
	return rows
}

func statusIndicator(status string) string {
	switch status {
	case "synced":
		return "="
	case "modified":
		return "M"
	case "conflict":
		return "!"
	case "untracked":
		return "?"
	case "added":
		return "+"
	case "deleted":
		return "-"
	default:
		return " "
	}
}

// SyncPullOutput represents the sync pull command output.
type SyncPullOutput struct {
	Remote        string              `json:"remote"`
	Success       bool                `json:"success"`
	FilesUpdated  []SyncFileChange    `json:"files_updated"`
	FilesConflict []SyncConflictEntry `json:"files_conflict,omitempty"`
	HasConflicts  bool                `json:"has_conflicts"`
	UpdatedCount  int                 `json:"updated_count"`
	ConflictCount int                 `json:"conflict_count"`
	Message       string              `json:"message,omitempty"`
}

// SyncFileChange represents a file changed during sync.
type SyncFileChange struct {
	Path      string `json:"path"`
	Action    string `json:"action"` // updated, added, deleted
	OldHash   string `json:"old_hash,omitempty"`
	NewHash   string `json:"new_hash,omitempty"`
	BytesDiff int64  `json:"bytes_diff,omitempty"`
}

// Text implements Textable for SyncPullOutput.
func (p SyncPullOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Pulling from %s...\n\n", p.Remote))

	if len(p.FilesUpdated) > 0 {
		b.WriteString("Updated:\n")
		for _, f := range p.FilesUpdated {
			b.WriteString(fmt.Sprintf("  %s %s\n", actionSymbol(f.Action), f.Path))
		}
	}

	if p.HasConflicts {
		b.WriteString("\nConflicts:\n")
		for _, c := range p.FilesConflict {
			b.WriteString(fmt.Sprintf("  ! %s: %s\n", c.Path, c.Description))
		}
		b.WriteString(fmt.Sprintf("\n%d conflicts detected. Run 'ari sync resolve' to resolve.\n", p.ConflictCount))
	} else if p.Success {
		b.WriteString(fmt.Sprintf("\nPull complete: %d files updated\n", p.UpdatedCount))
	}

	return b.String()
}

func actionSymbol(action string) string {
	switch action {
	case "added":
		return "+"
	case "deleted":
		return "-"
	case "updated":
		return "~"
	default:
		return " "
	}
}

// SyncPushOutput represents the sync push command output.
type SyncPushOutput struct {
	Remote       string           `json:"remote"`
	Success      bool             `json:"success"`
	FilesPushed  []SyncFileChange `json:"files_pushed"`
	PushedCount  int              `json:"pushed_count"`
	Rejected     bool             `json:"rejected"`
	RejectReason string           `json:"reject_reason,omitempty"`
	Message      string           `json:"message,omitempty"`
}

// Text implements Textable for SyncPushOutput.
func (p SyncPushOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Pushing to %s...\n\n", p.Remote))

	if p.Rejected {
		b.WriteString(fmt.Sprintf("Push rejected: %s\n", p.RejectReason))
		return b.String()
	}

	if len(p.FilesPushed) > 0 {
		b.WriteString("Pushed:\n")
		for _, f := range p.FilesPushed {
			b.WriteString(fmt.Sprintf("  %s %s\n", actionSymbol(f.Action), f.Path))
		}
	}

	if p.Success {
		b.WriteString(fmt.Sprintf("\nPush complete: %d files pushed\n", p.PushedCount))
	}

	return b.String()
}

// SyncDiffOutput represents the sync diff command output.
type SyncDiffOutput struct {
	Path          string `json:"path,omitempty"`
	HasChanges    bool   `json:"has_changes"`
	LocalContent  string `json:"local_content,omitempty"`
	RemoteContent string `json:"remote_content,omitempty"`
	UnifiedDiff   string `json:"unified_diff,omitempty"`
	Additions     int    `json:"additions"`
	Deletions     int    `json:"deletions"`
	TotalFiles    int    `json:"total_files"`
	ChangedFiles  int    `json:"changed_files"`
}

// Text implements Textable for SyncDiffOutput.
func (d SyncDiffOutput) Text() string {
	if !d.HasChanges {
		return "No differences found"
	}

	if d.UnifiedDiff != "" {
		return d.UnifiedDiff
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d file(s) changed\n", d.ChangedFiles))
	b.WriteString(fmt.Sprintf("+%d -%d\n", d.Additions, d.Deletions))
	return b.String()
}

// SyncResolveOutput represents the sync resolve command output.
type SyncResolveOutput struct {
	Path           string   `json:"path"`
	Strategy       string   `json:"strategy"`
	Resolved       bool     `json:"resolved"`
	RemainingCount int      `json:"remaining_count"`
	Remaining      []string `json:"remaining,omitempty"`
	Message        string   `json:"message,omitempty"`
}

// Text implements Textable for SyncResolveOutput.
func (r SyncResolveOutput) Text() string {
	var b strings.Builder
	if r.Resolved {
		b.WriteString(fmt.Sprintf("Resolved: %s (strategy: %s)\n", r.Path, r.Strategy))
	} else {
		b.WriteString(fmt.Sprintf("Failed to resolve: %s\n", r.Path))
	}

	if r.RemainingCount > 0 {
		b.WriteString(fmt.Sprintf("\n%d conflicts remaining:\n", r.RemainingCount))
		for _, p := range r.Remaining {
			b.WriteString(fmt.Sprintf("  ! %s\n", p))
		}
	} else if r.Resolved {
		b.WriteString("\nAll conflicts resolved\n")
	}

	return b.String()
}

// SyncHistoryOutput represents the sync history command output.
type SyncHistoryOutput struct {
	Entries []SyncHistoryEntry `json:"entries"`
	Total   int                `json:"total"`
	Limit   int                `json:"limit,omitempty"`
}

// SyncHistoryEntry represents a single sync history entry.
type SyncHistoryEntry struct {
	Timestamp string                 `json:"timestamp"`
	Operation string                 `json:"operation"` // pull, push, resolve
	Remote    string                 `json:"remote,omitempty"`
	Files     []string               `json:"files,omitempty"`
	FileCount int                    `json:"file_count"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Text implements Textable for SyncHistoryOutput.
func (h SyncHistoryOutput) Text() string {
	if len(h.Entries) == 0 {
		return "No sync history"
	}

	var b strings.Builder
	for _, e := range h.Entries {
		status := "OK"
		if !e.Success {
			status = "FAIL"
		}
		b.WriteString(fmt.Sprintf("[%s] %s %s (%d files) %s\n",
			e.Timestamp, e.Operation, e.Remote, e.FileCount, status))
	}
	b.WriteString(fmt.Sprintf("\nTotal: %d entries\n", h.Total))
	return b.String()
}

// Headers implements Tabular for SyncHistoryOutput.
func (h SyncHistoryOutput) Headers() []string {
	return []string{"TIMESTAMP", "OPERATION", "REMOTE", "FILES", "STATUS"}
}

// Rows implements Tabular for SyncHistoryOutput.
func (h SyncHistoryOutput) Rows() [][]string {
	rows := make([][]string, len(h.Entries))
	for i, e := range h.Entries {
		status := "OK"
		if !e.Success {
			status = "FAIL"
		}
		rows[i] = []string{e.Timestamp, e.Operation, e.Remote, fmt.Sprintf("%d", e.FileCount), status}
	}
	return rows
}

// SyncResetOutput represents the sync reset command output.
type SyncResetOutput struct {
	Reset       bool     `json:"reset"`
	Hard        bool     `json:"hard"`
	FilesReset  []string `json:"files_reset,omitempty"`
	StateCleared bool    `json:"state_cleared"`
	Message     string   `json:"message,omitempty"`
}

// Text implements Textable for SyncResetOutput.
func (r SyncResetOutput) Text() string {
	var b strings.Builder

	if r.Hard {
		b.WriteString("Hard reset performed\n")
		if r.StateCleared {
			b.WriteString("  - Sync state cleared\n")
		}
		if len(r.FilesReset) > 0 {
			b.WriteString("  - Files reverted:\n")
			for _, f := range r.FilesReset {
				b.WriteString(fmt.Sprintf("      %s\n", f))
			}
		}
	} else if r.Reset {
		b.WriteString("Sync state reset\n")
	} else {
		b.WriteString("Reset cancelled\n")
	}

	return b.String()
}
