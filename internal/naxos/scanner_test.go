package naxos

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

// testSetup creates a temporary test environment with session directories.
type testSetup struct {
	tempDir     string
	projectRoot string
	sessionsDir string
	t           *testing.T
}

func newTestSetup(t *testing.T) *testSetup {
	t.Helper()
	tempDir := t.TempDir()

	projectRoot := tempDir
	sessionsDir := filepath.Join(projectRoot, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	return &testSetup{
		tempDir:     tempDir,
		projectRoot: projectRoot,
		sessionsDir: sessionsDir,
		t:           t,
	}
}

func (ts *testSetup) resolver() *paths.Resolver {
	return paths.NewResolver(ts.projectRoot)
}

// createSession creates a session with the given properties.
func (ts *testSetup) createSession(id, status string, createdAt time.Time, opts ...sessionOption) {
	ts.t.Helper()

	sessionDir := filepath.Join(ts.sessionsDir, id)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		ts.t.Fatalf("Failed to create session dir: %v", err)
	}

	// Build session context
	so := &sessionOpts{
		status:       status,
		createdAt:    createdAt,
		initiative:   "Test Initiative",
		complexity:   "MODULE",
		activeRite:   "test-rite",
		currentPhase: "requirements",
	}
	for _, opt := range opts {
		opt(so)
	}

	content := buildSessionContext(so)
	contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextPath, []byte(content), 0644); err != nil {
		ts.t.Fatalf("Failed to write session context: %v", err)
	}

	// Create WHITE_SAILS.yaml if specified
	if so.sailsColor != "" {
		sailsContent := "color: " + so.sailsColor + "\n"
		sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
		if err := os.WriteFile(sailsPath, []byte(sailsContent), 0644); err != nil {
			ts.t.Fatalf("Failed to write WHITE_SAILS.yaml: %v", err)
		}
	}
}

type sessionOpts struct {
	status       string
	createdAt    time.Time
	initiative   string
	complexity   string
	activeRite   string
	currentPhase string
	parkedAt     *time.Time
	parkedReason string
	sailsColor   string
}

type sessionOption func(*sessionOpts)

func withParked(parkedAt time.Time, reason string) sessionOption {
	return func(so *sessionOpts) {
		so.parkedAt = &parkedAt
		so.parkedReason = reason
	}
}

func withSailsColor(color string) sessionOption {
	return func(so *sessionOpts) {
		so.sailsColor = color
	}
}

func withCurrentPhase(phase string) sessionOption {
	return func(so *sessionOpts) {
		so.currentPhase = phase
	}
}

func buildSessionContext(so *sessionOpts) string {
	content := `---
schema_version: "2.1"
session_id: "` + "test" + `"
status: "` + so.status + `"
created_at: "` + so.createdAt.Format(time.RFC3339) + `"
initiative: "` + so.initiative + `"
complexity: "` + so.complexity + `"
active_rite: "` + so.activeRite + `"
current_phase: "` + so.currentPhase + `"`

	if so.parkedAt != nil {
		content += `
parked_at: "` + so.parkedAt.Format(time.RFC3339) + `"`
	}
	if so.parkedReason != "" {
		content += `
parked_reason: "` + so.parkedReason + `"`
	}

	content += `
---

# Session

Test body.
`
	return content
}

func TestScanner_Scan_NoSessions(t *testing.T) {
	ts := newTestSetup(t)


	scanner := NewScanner(ts.resolver(), DefaultConfig())
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalScanned != 0 {
		t.Errorf("TotalScanned = %d, want 0", result.TotalScanned)
	}
	if result.TotalOrphaned != 0 {
		t.Errorf("TotalOrphaned = %d, want 0", result.TotalOrphaned)
	}
}

func TestScanner_Scan_HealthyActiveSession(t *testing.T) {
	ts := newTestSetup(t)


	// Create a recent active session (should not be flagged)
	now := time.Now().UTC()
	ts.createSession("session-20260106-120000-abcd1234", "ACTIVE", now.Add(-1*time.Hour))

	scanner := NewScanner(ts.resolver(), DefaultConfig())
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalScanned != 1 {
		t.Errorf("TotalScanned = %d, want 1", result.TotalScanned)
	}
	if result.TotalOrphaned != 0 {
		t.Errorf("TotalOrphaned = %d, want 0", result.TotalOrphaned)
	}
}

func TestScanner_Scan_InactiveSession(t *testing.T) {
	ts := newTestSetup(t)


	// Create an old active session (should be flagged as inactive)
	now := time.Now().UTC()
	ts.createSession("session-20260103-120000-abcd1234", "ACTIVE", now.Add(-48*time.Hour))

	config := DefaultConfig()
	config.InactiveThreshold = 24 * time.Hour

	scanner := NewScanner(ts.resolver(), config)
	// Override now for consistent testing
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalScanned != 1 {
		t.Errorf("TotalScanned = %d, want 1", result.TotalScanned)
	}
	if result.TotalOrphaned != 1 {
		t.Errorf("TotalOrphaned = %d, want 1", result.TotalOrphaned)
	}
	if len(result.OrphanedSessions) != 1 {
		t.Fatalf("Expected 1 orphaned session, got %d", len(result.OrphanedSessions))
	}

	orphan := result.OrphanedSessions[0]
	if orphan.Reason != ReasonInactive {
		t.Errorf("Reason = %v, want %v", orphan.Reason, ReasonInactive)
	}
	if orphan.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", orphan.Status, "ACTIVE")
	}
}

func TestScanner_Scan_StaleSailsSession(t *testing.T) {
	ts := newTestSetup(t)


	// Create a parked session with gray sails that's old
	now := time.Now().UTC()
	parkedAt := now.Add(-10 * 24 * time.Hour) // Parked 10 days ago
	ts.createSession("session-20251227-120000-abcd1234", "PARKED", now.Add(-15*24*time.Hour),
		withParked(parkedAt, ""),
		withSailsColor("GRAY"))

	config := DefaultConfig()
	config.StaleSailsThreshold = 7 * 24 * time.Hour

	scanner := NewScanner(ts.resolver(), config)
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalOrphaned != 1 {
		t.Errorf("TotalOrphaned = %d, want 1", result.TotalOrphaned)
	}
	if len(result.OrphanedSessions) != 1 {
		t.Fatalf("Expected 1 orphaned session, got %d", len(result.OrphanedSessions))
	}

	orphan := result.OrphanedSessions[0]
	if orphan.Reason != ReasonStaleSails {
		t.Errorf("Reason = %v, want %v", orphan.Reason, ReasonStaleSails)
	}
	if orphan.SailsColor != "GRAY" {
		t.Errorf("SailsColor = %q, want %q", orphan.SailsColor, "GRAY")
	}
}

func TestScanner_Scan_ParkedWithWhiteSails_NotFlagged(t *testing.T) {
	ts := newTestSetup(t)


	// Create a parked session with WHITE sails (should not be flagged)
	now := time.Now().UTC()
	parkedAt := now.Add(-10 * 24 * time.Hour)
	ts.createSession("session-20251227-120000-abcd1234", "PARKED", now.Add(-15*24*time.Hour),
		withParked(parkedAt, "Completed work"),
		withSailsColor("WHITE"))

	config := DefaultConfig()
	scanner := NewScanner(ts.resolver(), config)
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalOrphaned != 0 {
		t.Errorf("TotalOrphaned = %d, want 0 (WHITE sails should not be flagged)", result.TotalOrphaned)
	}
}

func TestScanner_Scan_IncompleteWrap(t *testing.T) {
	ts := newTestSetup(t)


	// Create a session with wrap phase but still ACTIVE
	now := time.Now().UTC()
	ts.createSession("session-20260105-120000-abcd1234", "ACTIVE", now.Add(-1*time.Hour),
		withCurrentPhase("wrap"))

	scanner := NewScanner(ts.resolver(), DefaultConfig())
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalOrphaned != 1 {
		t.Errorf("TotalOrphaned = %d, want 1", result.TotalOrphaned)
	}
	if len(result.OrphanedSessions) != 1 {
		t.Fatalf("Expected 1 orphaned session, got %d", len(result.OrphanedSessions))
	}

	orphan := result.OrphanedSessions[0]
	if orphan.Reason != ReasonIncompleteWrap {
		t.Errorf("Reason = %v, want %v", orphan.Reason, ReasonIncompleteWrap)
	}
	if orphan.SuggestedAction != ActionWrap {
		t.Errorf("SuggestedAction = %v, want %v", orphan.SuggestedAction, ActionWrap)
	}
}

func TestScanner_Scan_ArchivedNotIncluded(t *testing.T) {
	ts := newTestSetup(t)


	// Create an archived session
	now := time.Now().UTC()
	ts.createSession("session-20260101-120000-abcd1234", "ARCHIVED", now.Add(-30*24*time.Hour))

	config := DefaultConfig()
	config.IncludeArchived = false

	scanner := NewScanner(ts.resolver(), config)
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Archived sessions should be scanned but not flagged when IncludeArchived=false
	if result.TotalScanned != 1 {
		t.Errorf("TotalScanned = %d, want 1", result.TotalScanned)
	}
	if result.TotalOrphaned != 0 {
		t.Errorf("TotalOrphaned = %d, want 0", result.TotalOrphaned)
	}
}

func TestScanner_Scan_MultipleOrphans(t *testing.T) {
	ts := newTestSetup(t)


	now := time.Now().UTC()

	// Create multiple orphaned sessions with different reasons
	ts.createSession("session-20260102-120000-aaaa1111", "ACTIVE", now.Add(-48*time.Hour)) // Inactive

	parkedAt := now.Add(-10 * 24 * time.Hour)
	ts.createSession("session-20251227-120000-bbbb2222", "PARKED", now.Add(-15*24*time.Hour),
		withParked(parkedAt, "")) // Stale sails

	ts.createSession("session-20260105-120000-cccc3333", "ACTIVE", now.Add(-1*time.Hour),
		withCurrentPhase("wrap")) // Incomplete wrap

	// Also create a healthy session
	ts.createSession("session-20260106-120000-dddd4444", "ACTIVE", now.Add(-1*time.Hour)) // Healthy

	config := DefaultConfig()
	scanner := NewScanner(ts.resolver(), config)
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalScanned != 4 {
		t.Errorf("TotalScanned = %d, want 4", result.TotalScanned)
	}
	if result.TotalOrphaned != 3 {
		t.Errorf("TotalOrphaned = %d, want 3", result.TotalOrphaned)
	}

	// Check by reason counts
	if result.ByReason[ReasonInactive] != 1 {
		t.Errorf("ByReason[Inactive] = %d, want 1", result.ByReason[ReasonInactive])
	}
	if result.ByReason[ReasonStaleSails] != 1 {
		t.Errorf("ByReason[StaleSails] = %d, want 1", result.ByReason[ReasonStaleSails])
	}
	if result.ByReason[ReasonIncompleteWrap] != 1 {
		t.Errorf("ByReason[IncompleteWrap] = %d, want 1", result.ByReason[ReasonIncompleteWrap])
	}
}

func TestScanner_Scan_CustomThresholds(t *testing.T) {
	ts := newTestSetup(t)


	now := time.Now().UTC()

	// Create a session that's 12 hours old
	ts.createSession("session-20260106-000000-abcd1234", "ACTIVE", now.Add(-12*time.Hour))

	// With default 24h threshold, should NOT be flagged
	config := DefaultConfig()
	scanner := NewScanner(ts.resolver(), config)
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if result.TotalOrphaned != 0 {
		t.Errorf("With 24h threshold, TotalOrphaned = %d, want 0", result.TotalOrphaned)
	}

	// With 6h threshold, SHOULD be flagged
	config.InactiveThreshold = 6 * time.Hour
	scanner = NewScanner(ts.resolver(), config)
	scanner.now = func() time.Time { return now }

	result, err = scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if result.TotalOrphaned != 1 {
		t.Errorf("With 6h threshold, TotalOrphaned = %d, want 1", result.TotalOrphaned)
	}
}

func TestScanner_Scan_NonSessionDirsIgnored(t *testing.T) {
	ts := newTestSetup(t)


	// Create non-session directories that should be ignored
	os.MkdirAll(filepath.Join(ts.sessionsDir, ".locks"), 0755)
	os.MkdirAll(filepath.Join(ts.sessionsDir, ".audit"), 0755)
	os.MkdirAll(filepath.Join(ts.sessionsDir, "not-a-session"), 0755)
	os.WriteFile(filepath.Join(ts.sessionsDir, "some-file.txt"), []byte("test"), 0644)

	scanner := NewScanner(ts.resolver(), DefaultConfig())
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalScanned != 0 {
		t.Errorf("TotalScanned = %d, want 0 (non-session dirs should be ignored)", result.TotalScanned)
	}
}

func TestScanner_SuggestedAction_VeryOldInactive(t *testing.T) {
	ts := newTestSetup(t)


	now := time.Now().UTC()

	// Create a very old inactive session (>30 days)
	ts.createSession("session-20251205-120000-abcd1234", "ACTIVE", now.Add(-35*24*time.Hour))

	scanner := NewScanner(ts.resolver(), DefaultConfig())
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalOrphaned != 1 {
		t.Fatalf("Expected 1 orphaned session, got %d", result.TotalOrphaned)
	}

	orphan := result.OrphanedSessions[0]
	if orphan.SuggestedAction != ActionDelete {
		t.Errorf("SuggestedAction = %v, want %v for very old session", orphan.SuggestedAction, ActionDelete)
	}
}

func TestScanner_SuggestedAction_RecentInactive(t *testing.T) {
	ts := newTestSetup(t)


	now := time.Now().UTC()

	// Create a recently inactive session (2 days)
	ts.createSession("session-20260104-120000-abcd1234", "ACTIVE", now.Add(-2*24*time.Hour))

	scanner := NewScanner(ts.resolver(), DefaultConfig())
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalOrphaned != 1 {
		t.Fatalf("Expected 1 orphaned session, got %d", result.TotalOrphaned)
	}

	orphan := result.OrphanedSessions[0]
	if orphan.SuggestedAction != ActionResume {
		t.Errorf("SuggestedAction = %v, want %v for recent inactive session", orphan.SuggestedAction, ActionResume)
	}
}

func TestScanner_SuggestedAction_StaleSailsWithReason(t *testing.T) {
	ts := newTestSetup(t)


	now := time.Now().UTC()
	parkedAt := now.Add(-10 * 24 * time.Hour)

	// Parked with explicit reason
	ts.createSession("session-20251227-120000-abcd1234", "PARKED", now.Add(-15*24*time.Hour),
		withParked(parkedAt, "Waiting for external review"),
		withSailsColor("GRAY"))

	scanner := NewScanner(ts.resolver(), DefaultConfig())
	scanner.now = func() time.Time { return now }

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalOrphaned != 1 {
		t.Fatalf("Expected 1 orphaned session, got %d", result.TotalOrphaned)
	}

	orphan := result.OrphanedSessions[0]
	if orphan.SuggestedAction != ActionResume {
		t.Errorf("SuggestedAction = %v, want %v (has park reason)", orphan.SuggestedAction, ActionResume)
	}
}

func TestOrphanReason_String(t *testing.T) {
	tests := []struct {
		reason OrphanReason
		want   string
	}{
		{ReasonInactive, "INACTIVE"},
		{ReasonStaleSails, "STALE_SAILS"},
		{ReasonIncompleteWrap, "INCOMPLETE_WRAP"},
	}

	for _, tt := range tests {
		if got := tt.reason.String(); got != tt.want {
			t.Errorf("OrphanReason.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestOrphanReason_Description(t *testing.T) {
	if desc := ReasonInactive.Description(); desc == "" {
		t.Error("ReasonInactive.Description() should not be empty")
	}
	if desc := ReasonStaleSails.Description(); desc == "" {
		t.Error("ReasonStaleSails.Description() should not be empty")
	}
	if desc := ReasonIncompleteWrap.Description(); desc == "" {
		t.Error("ReasonIncompleteWrap.Description() should not be empty")
	}
}

func TestSuggestedAction_String(t *testing.T) {
	tests := []struct {
		action SuggestedAction
		want   string
	}{
		{ActionWrap, "WRAP"},
		{ActionResume, "RESUME"},
		{ActionDelete, "DELETE"},
	}

	for _, tt := range tests {
		if got := tt.action.String(); got != tt.want {
			t.Errorf("SuggestedAction.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestSuggestedAction_Description(t *testing.T) {
	if desc := ActionWrap.Description(); desc == "" {
		t.Error("ActionWrap.Description() should not be empty")
	}
	if desc := ActionResume.Description(); desc == "" {
		t.Error("ActionResume.Description() should not be empty")
	}
	if desc := ActionDelete.Description(); desc == "" {
		t.Error("ActionDelete.Description() should not be empty")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.InactiveThreshold != 24*time.Hour {
		t.Errorf("InactiveThreshold = %v, want %v", config.InactiveThreshold, 24*time.Hour)
	}
	if config.StaleSailsThreshold != 7*24*time.Hour {
		t.Errorf("StaleSailsThreshold = %v, want %v", config.StaleSailsThreshold, 7*24*time.Hour)
	}
	if config.IncludeArchived != false {
		t.Errorf("IncludeArchived = %v, want false", config.IncludeArchived)
	}
}

func TestScanResult_Add(t *testing.T) {
	result := NewScanResult(DefaultConfig())

	orphan := OrphanedSession{
		SessionID: "test-session",
		Reason:    ReasonInactive,
	}
	result.Add(orphan)

	if result.TotalOrphaned != 1 {
		t.Errorf("TotalOrphaned = %d, want 1", result.TotalOrphaned)
	}
	if result.ByReason[ReasonInactive] != 1 {
		t.Errorf("ByReason[Inactive] = %d, want 1", result.ByReason[ReasonInactive])
	}
	if len(result.OrphanedSessions) != 1 {
		t.Errorf("len(OrphanedSessions) = %d, want 1", len(result.OrphanedSessions))
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Minute, "30m0s"},
		{1 * time.Hour, "1 hour"},
		{2 * time.Hour, "2 hours"},
		{23 * time.Hour, "23 hours"},
		{24 * time.Hour, "1 day"},
		{48 * time.Hour, "2 days"},
		{7 * 24 * time.Hour, "7 days"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.d)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

