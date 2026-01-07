package naxos

import (
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// Scanner scans session directories for orphaned sessions.
type Scanner struct {
	config   ScanConfig
	resolver *paths.Resolver
	now      func() time.Time // for testing
}

// NewScanner creates a new scanner with the given configuration.
func NewScanner(resolver *paths.Resolver, config ScanConfig) *Scanner {
	return &Scanner{
		config:   config,
		resolver: resolver,
		now:      time.Now,
	}
}

// Scan examines all sessions and returns those flagged for cleanup.
func (s *Scanner) Scan() (*ScanResult, error) {
	result := NewScanResult(s.config)

	// Scan active sessions directory
	sessionsDir := s.resolver.SessionsDir()
	if err := s.scanDirectory(sessionsDir, result); err != nil {
		// Directory might not exist, that's OK
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Optionally scan archive directory
	if s.config.IncludeArchived {
		archiveDir := s.resolver.ArchiveDir()
		if err := s.scanDirectory(archiveDir, result); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	return result, nil
}

// scanDirectory scans a single directory for session subdirectories.
func (s *Scanner) scanDirectory(dir string, result *ScanResult) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip non-session directories
		if !paths.IsSessionDir(entry.Name()) {
			continue
		}

		sessionID := entry.Name()
		sessionDir := filepath.Join(dir, sessionID)
		contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")

		result.TotalScanned++

		// Try to load session context
		ctx, err := session.LoadContext(contextPath)
		if err != nil {
			// Can't load context - might be corrupt, skip
			continue
		}

		// Check for orphan conditions
		orphan := s.checkSession(sessionID, sessionDir, ctx)
		if orphan != nil {
			result.Add(*orphan)
		}
	}

	return nil
}

// checkSession examines a session and returns an OrphanedSession if flagged.
func (s *Scanner) checkSession(sessionID, sessionDir string, ctx *session.Context) *OrphanedSession {
	now := s.now().UTC()

	// Skip archived sessions unless explicitly included
	if ctx.Status == session.StatusArchived && !s.config.IncludeArchived {
		return nil
	}

	// Determine last activity time
	lastActivity := s.determineLastActivity(ctx)

	// Check for inactive sessions (ACTIVE status but no activity)
	if ctx.Status == session.StatusActive {
		inactiveDuration := now.Sub(lastActivity)
		if inactiveDuration > s.config.InactiveThreshold {
			return &OrphanedSession{
				SessionID:       sessionID,
				SessionDir:      sessionDir,
				Status:          string(ctx.Status),
				Initiative:      ctx.Initiative,
				Reason:          ReasonInactive,
				SuggestedAction: s.suggestActionForInactive(ctx),
				Age:             now.Sub(ctx.CreatedAt),
				InactiveFor:     inactiveDuration,
				CreatedAt:       ctx.CreatedAt,
				LastActivity:    lastActivity,
				AdditionalInfo:  formatDuration(inactiveDuration) + " since last activity",
			}
		}
	}

	// Check for incomplete wrap (session marked for archival but not completed)
	if s.isIncompleteWrap(ctx) {
		return &OrphanedSession{
			SessionID:       sessionID,
			SessionDir:      sessionDir,
			Status:          string(ctx.Status),
			Initiative:      ctx.Initiative,
			Reason:          ReasonIncompleteWrap,
			SuggestedAction: ActionWrap,
			Age:             now.Sub(ctx.CreatedAt),
			InactiveFor:     now.Sub(lastActivity),
			CreatedAt:       ctx.CreatedAt,
			LastActivity:    lastActivity,
			AdditionalInfo:  "Wrap was initiated but never completed",
		}
	}

	// Check for stale gray sails (parked sessions with gray sails past threshold)
	if ctx.Status == session.StatusParked {
		sailsColor := s.checkSailsColor(sessionDir)
		if sailsColor == "GRAY" || sailsColor == "" {
			// Check if parked long enough to be considered stale
			parkedAt := ctx.ParkedAt
			if parkedAt != nil {
				staleDuration := now.Sub(*parkedAt)
				if staleDuration > s.config.StaleSailsThreshold {
					return &OrphanedSession{
						SessionID:       sessionID,
						SessionDir:      sessionDir,
						Status:          string(ctx.Status),
						Initiative:      ctx.Initiative,
						Reason:          ReasonStaleSails,
						SuggestedAction: s.suggestActionForStaleSails(ctx),
						Age:             now.Sub(ctx.CreatedAt),
						InactiveFor:     staleDuration,
						CreatedAt:       ctx.CreatedAt,
						LastActivity:    *parkedAt,
						SailsColor:      sailsColor,
						AdditionalInfo:  formatDuration(staleDuration) + " parked with gray/unknown sails",
					}
				}
			}
		}
	}

	return nil
}

// determineLastActivity finds the most recent activity timestamp for a session.
func (s *Scanner) determineLastActivity(ctx *session.Context) time.Time {
	latest := ctx.CreatedAt

	if ctx.ResumedAt != nil && ctx.ResumedAt.After(latest) {
		latest = *ctx.ResumedAt
	}

	if ctx.ParkedAt != nil && ctx.ParkedAt.After(latest) {
		latest = *ctx.ParkedAt
	}

	if ctx.ArchivedAt != nil && ctx.ArchivedAt.After(latest) {
		latest = *ctx.ArchivedAt
	}

	return latest
}

// isIncompleteWrap checks if a session has an incomplete wrap.
func (s *Scanner) isIncompleteWrap(ctx *session.Context) bool {
	// A session is considered to have an incomplete wrap if:
	// 1. Status is ACTIVE but current_phase indicates wrap was started
	// 2. The body contains wrap markers but no archived_at
	if ctx.Status == session.StatusActive && ctx.CurrentPhase == "wrap" {
		return true
	}

	// Check if session context body mentions wrap initiation
	// This is a heuristic - could be refined
	return false
}

// checkSailsColor reads the sails color from a session's sails.yaml if present.
func (s *Scanner) checkSailsColor(sessionDir string) string {
	sailsPath := filepath.Join(sessionDir, "sails.yaml")
	_, err := os.Stat(sailsPath)
	if os.IsNotExist(err) {
		return "" // No sails file = unknown/gray
	}

	// For simplicity, if sails.yaml exists, read it
	// In a real implementation, we'd parse the YAML
	// For now, just return empty to indicate we should check
	data, err := os.ReadFile(sailsPath)
	if err != nil {
		return ""
	}

	// Simple string search for color
	content := string(data)
	if contains(content, "color: WHITE") || contains(content, "color: white") {
		return "WHITE"
	}
	if contains(content, "color: BLACK") || contains(content, "color: black") {
		return "BLACK"
	}
	if contains(content, "color: GRAY") || contains(content, "color: gray") {
		return "GRAY"
	}

	return "GRAY" // Default to gray if we can't determine
}

// suggestActionForInactive determines the best action for an inactive session.
func (s *Scanner) suggestActionForInactive(ctx *session.Context) SuggestedAction {
	// If session has artifacts, suggest wrap
	// If session is very old with no artifacts, suggest delete
	// Otherwise suggest resume

	// Check if session is over 30 days old
	age := s.now().UTC().Sub(ctx.CreatedAt)
	if age > 30*24*time.Hour {
		// Very old inactive session - probably safe to delete
		return ActionDelete
	}

	// Default to resume for moderately old sessions
	return ActionResume
}

// suggestActionForStaleSails determines the best action for stale sails sessions.
func (s *Scanner) suggestActionForStaleSails(ctx *session.Context) SuggestedAction {
	// Parked sessions with stale gray sails should probably be wrapped or resumed
	// If parked reason suggests intentional pause, suggest resume
	if ctx.ParkedReason != "" {
		// Has an explicit park reason, user may want to resume
		return ActionResume
	}

	// No reason given, probably should be wrapped
	return ActionWrap
}

// contains is a simple string contains check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		return pluralize(hours, "hour")
	}
	days := int(d.Hours() / 24)
	return pluralize(days, "day")
}

// pluralize returns "n unit" or "n units" as appropriate.
func pluralize(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return formatInt(n) + " " + unit + "s"
}

// formatInt converts an int to string without importing strconv.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatInt(-n)
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
