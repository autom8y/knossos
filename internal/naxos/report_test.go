package naxos

import (
	"strings"
	"testing"
	"time"
)

func TestScanOutput_Headers(t *testing.T) {
	output := ScanOutput{}
	headers := output.Headers()

	expectedHeaders := []string{"SESSION ID", "STATUS", "REASON", "INACTIVE", "SUGGESTED ACTION"}
	if len(headers) != len(expectedHeaders) {
		t.Errorf("len(Headers()) = %d, want %d", len(headers), len(expectedHeaders))
	}
	for i, h := range headers {
		if h != expectedHeaders[i] {
			t.Errorf("Headers()[%d] = %q, want %q", i, h, expectedHeaders[i])
		}
	}
}

func TestScanOutput_Rows(t *testing.T) {
	output := ScanOutput{
		OrphanedSessions: []OrphanedSession{
			{
				SessionID:       "session-20260104-120000-abcd1234",
				Status:          "ACTIVE",
				Reason:          ReasonInactive,
				InactiveFor:     48 * time.Hour,
				SuggestedAction: ActionResume,
			},
			{
				SessionID:       "session-20260103-120000-efgh5678",
				Status:          "PARKED",
				Reason:          ReasonStaleSails,
				InactiveFor:     10 * 24 * time.Hour,
				SuggestedAction: ActionWrap,
			},
		},
	}

	rows := output.Rows()

	if len(rows) != 2 {
		t.Fatalf("len(Rows()) = %d, want 2", len(rows))
	}

	// Check first row
	if !strings.Contains(rows[0][0], "session-20260104") {
		t.Errorf("Row[0][0] should contain session ID, got %q", rows[0][0])
	}
	if rows[0][1] != "ACTIVE" {
		t.Errorf("Row[0][1] = %q, want ACTIVE", rows[0][1])
	}
	if !strings.Contains(rows[0][2], "INACTIVE") {
		t.Errorf("Row[0][2] should contain reason, got %q", rows[0][2])
	}
	if rows[0][4] != "RESUME" {
		t.Errorf("Row[0][4] = %q, want RESUME", rows[0][4])
	}

	// Check second row
	if rows[1][1] != "PARKED" {
		t.Errorf("Row[1][1] = %q, want PARKED", rows[1][1])
	}
	if rows[1][4] != "WRAP" {
		t.Errorf("Row[1][4] = %q, want WRAP", rows[1][4])
	}
}

func TestScanOutput_Rows_TruncatesLongSessionID(t *testing.T) {
	output := ScanOutput{
		OrphanedSessions: []OrphanedSession{
			{
				SessionID:       "session-20260104-120000-abcdefghijklmnopqrstuvwxyz1234567890",
				Status:          "ACTIVE",
				Reason:          ReasonInactive,
				InactiveFor:     48 * time.Hour,
				SuggestedAction: ActionResume,
			},
		},
	}

	rows := output.Rows()
	if len(rows[0][0]) > 38 { // 35 + "..."
		t.Errorf("Session ID should be truncated, got length %d", len(rows[0][0]))
	}
	if !strings.HasSuffix(rows[0][0], "...") {
		t.Errorf("Truncated session ID should end with ..., got %q", rows[0][0])
	}
}

func TestScanOutput_Text_NoOrphans(t *testing.T) {
	output := ScanOutput{
		TotalScanned:     5,
		TotalOrphaned:    0,
		OrphanedSessions: []OrphanedSession{},
		ScannedAt:        "2026-01-06T12:00:00Z",
		Config: ConfigSummary{
			InactiveThreshold:   "24 hours",
			StaleSailsThreshold: "7 days",
		},
	}

	text := output.Text()

	if !strings.Contains(text, "Scanned: 5 sessions") {
		t.Errorf("Text should contain scanned count, got:\n%s", text)
	}
	if !strings.Contains(text, "No orphaned sessions found") {
		t.Errorf("Text should indicate no orphans, got:\n%s", text)
	}
}

func TestScanOutput_Text_WithOrphans(t *testing.T) {
	output := ScanOutput{
		TotalScanned:  10,
		TotalOrphaned: 2,
		OrphanedSessions: []OrphanedSession{
			{
				SessionID:       "session-20260104-120000-abcd1234",
				Status:          "ACTIVE",
				Initiative:      "Test Feature",
				Reason:          ReasonInactive,
				InactiveFor:     48 * time.Hour,
				SuggestedAction: ActionResume,
				AdditionalInfo:  "2 days since last activity",
			},
			{
				SessionID:       "session-20260103-120000-efgh5678",
				Status:          "PARKED",
				Initiative:      "Another Feature",
				Reason:          ReasonStaleSails,
				InactiveFor:     10 * 24 * time.Hour,
				SuggestedAction: ActionWrap,
				SailsColor:      "GRAY",
			},
		},
		ScannedAt: "2026-01-06T12:00:00Z",
		ByReason: ByReasonSummary{
			Inactive:   1,
			StaleSails: 1,
		},
		Config: ConfigSummary{
			InactiveThreshold:   "1 day",
			StaleSailsThreshold: "7 days",
		},
	}

	text := output.Text()

	// Check header
	if !strings.Contains(text, "Naxos Session Scan Report") {
		t.Errorf("Text should contain report title, got:\n%s", text)
	}

	// Check summary
	if !strings.Contains(text, "Scanned: 10 sessions") {
		t.Errorf("Text should contain scanned count, got:\n%s", text)
	}
	if !strings.Contains(text, "Orphaned: 2 sessions") {
		t.Errorf("Text should contain orphaned count, got:\n%s", text)
	}

	// Check by reason breakdown
	if !strings.Contains(text, "Inactive") {
		t.Errorf("Text should contain inactive reason, got:\n%s", text)
	}
	if !strings.Contains(text, "Stale Sails") {
		t.Errorf("Text should contain stale sails reason, got:\n%s", text)
	}

	// Check session details
	if !strings.Contains(text, "session-20260104") {
		t.Errorf("Text should contain session ID, got:\n%s", text)
	}
	if !strings.Contains(text, "Test Feature") {
		t.Errorf("Text should contain initiative, got:\n%s", text)
	}

	// Check action hints
	if !strings.Contains(text, "ari session wrap") {
		t.Errorf("Text should contain wrap action hint, got:\n%s", text)
	}
	if !strings.Contains(text, "ari session resume") {
		t.Errorf("Text should contain resume action hint, got:\n%s", text)
	}
}

func TestScanOutput_Text_IncompleteWrapReason(t *testing.T) {
	output := ScanOutput{
		TotalScanned:  1,
		TotalOrphaned: 1,
		OrphanedSessions: []OrphanedSession{
			{
				SessionID:       "session-20260105-120000-wrap1234",
				Status:          "ACTIVE",
				Initiative:      "Incomplete Wrap Test",
				Reason:          ReasonIncompleteWrap,
				InactiveFor:     1 * time.Hour,
				SuggestedAction: ActionWrap,
				AdditionalInfo:  "Wrap was initiated but never completed",
			},
		},
		ScannedAt: "2026-01-06T12:00:00Z",
		ByReason: ByReasonSummary{
			IncompleteWrap: 1,
		},
		Config: ConfigSummary{
			InactiveThreshold:   "1 day",
			StaleSailsThreshold: "7 days",
		},
	}

	text := output.Text()

	if !strings.Contains(text, "Incomplete Wrap") {
		t.Errorf("Text should contain incomplete wrap reason, got:\n%s", text)
	}
	if !strings.Contains(text, "[x]") {
		t.Errorf("Text should contain incomplete wrap symbol [x], got:\n%s", text)
	}
}

func TestFromScanResult(t *testing.T) {
	now := time.Now().UTC()
	config := ScanConfig{
		InactiveThreshold:   24 * time.Hour,
		StaleSailsThreshold: 7 * 24 * time.Hour,
		IncludeArchived:     false,
	}
	result := NewScanResult(config)
	result.TotalScanned = 5
	result.Add(OrphanedSession{
		SessionID:       "test-session",
		Reason:          ReasonInactive,
		SuggestedAction: ActionResume,
	})

	output := FromScanResult(result)

	if output.TotalScanned != 5 {
		t.Errorf("TotalScanned = %d, want 5", output.TotalScanned)
	}
	if output.TotalOrphaned != 1 {
		t.Errorf("TotalOrphaned = %d, want 1", output.TotalOrphaned)
	}
	if output.ByReason.Inactive != 1 {
		t.Errorf("ByReason.Inactive = %d, want 1", output.ByReason.Inactive)
	}
	if output.Config.InactiveThreshold == "" {
		t.Error("Config.InactiveThreshold should not be empty")
	}
	if output.ScannedAt == "" {
		t.Error("ScannedAt should not be empty")
	}

	// Check timestamp is parseable
	_, err := time.Parse(time.RFC3339, output.ScannedAt)
	if err != nil {
		t.Errorf("ScannedAt is not valid RFC3339: %v", err)
	}
	_ = now // silence unused warning
}

func TestReasonSymbol(t *testing.T) {
	tests := []struct {
		reason OrphanReason
		want   string
	}{
		{ReasonInactive, "[!]"},
		{ReasonStaleSails, "[~]"},
		{ReasonIncompleteWrap, "[x]"},
		{OrphanReason("unknown"), "[?]"},
	}

	for _, tt := range tests {
		got := reasonSymbol(tt.reason)
		if got != tt.want {
			t.Errorf("reasonSymbol(%v) = %q, want %q", tt.reason, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a very long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},       // Edge case: maxLen < 4
		{"hello", 4, "h..."},    // Just enough for ellipsis
	}

	for _, tt := range tests {
		got := truncate(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}

func TestByReasonSummary(t *testing.T) {
	summary := ByReasonSummary{
		Inactive:       3,
		StaleSails:     2,
		IncompleteWrap: 1,
	}

	if summary.Inactive != 3 {
		t.Errorf("Inactive = %d, want 3", summary.Inactive)
	}
	if summary.StaleSails != 2 {
		t.Errorf("StaleSails = %d, want 2", summary.StaleSails)
	}
	if summary.IncompleteWrap != 1 {
		t.Errorf("IncompleteWrap = %d, want 1", summary.IncompleteWrap)
	}
}

func TestConfigSummary(t *testing.T) {
	summary := ConfigSummary{
		InactiveThreshold:   "24 hours",
		StaleSailsThreshold: "7 days",
		IncludeArchived:     true,
	}

	if summary.InactiveThreshold != "24 hours" {
		t.Errorf("InactiveThreshold = %q, want %q", summary.InactiveThreshold, "24 hours")
	}
	if summary.StaleSailsThreshold != "7 days" {
		t.Errorf("StaleSailsThreshold = %q, want %q", summary.StaleSailsThreshold, "7 days")
	}
	if !summary.IncludeArchived {
		t.Error("IncludeArchived should be true")
	}
}
