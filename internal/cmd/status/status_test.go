package status

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

// setupProject creates a minimal project structure and returns project root + cleanup.
func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude", "agents"), 0755)
	os.MkdirAll(filepath.Join(dir, ".knossos"), 0755)
	return dir
}

func TestCollectClaude_Exists(t *testing.T) {
	root := setupProject(t)
	resolver := paths.NewResolver(root)

	// Write ACTIVE_RITE
	os.WriteFile(filepath.Join(root, ".knossos", "ACTIVE_RITE"), []byte("10x-dev\n"), 0644)

	// Write some agent files
	os.WriteFile(filepath.Join(root, ".claude", "agents", "pythia.md"), []byte("# agent"), 0644)
	os.WriteFile(filepath.Join(root, ".claude", "agents", "architect.md"), []byte("# agent"), 0644)
	os.WriteFile(filepath.Join(root, ".claude", "agents", "README.txt"), []byte("not an agent"), 0644) // should be excluded

	health := collectClaude(resolver)

	if !health.Exists {
		t.Error("expected Exists=true")
	}
	if health.ActiveRite != "10x-dev" {
		t.Errorf("expected ActiveRite=10x-dev, got %q", health.ActiveRite)
	}
	if health.AgentCount != 2 {
		t.Errorf("expected AgentCount=2, got %d", health.AgentCount)
	}
}

func TestCollectClaude_NotExists(t *testing.T) {
	dir := t.TempDir() // no .claude/
	resolver := paths.NewResolver(dir)

	health := collectClaude(resolver)
	if health.Exists {
		t.Error("expected Exists=false when .claude/ missing")
	}
}

func TestCollectKnossos_WithSatellites(t *testing.T) {
	root := setupProject(t)
	resolver := paths.NewResolver(root)

	// Create satellite rites
	ritesDir := filepath.Join(root, ".knossos", "rites")
	os.MkdirAll(filepath.Join(ritesDir, "custom-dev"), 0755)
	os.WriteFile(filepath.Join(ritesDir, "custom-dev", "manifest.yaml"), []byte("name: custom-dev"), 0644)
	os.MkdirAll(filepath.Join(ritesDir, "ml-pipeline"), 0755)
	os.WriteFile(filepath.Join(ritesDir, "ml-pipeline", "manifest.yaml"), []byte("name: ml-pipeline"), 0644)
	// Dir without manifest — should be skipped
	os.MkdirAll(filepath.Join(ritesDir, "invalid"), 0755)

	health := collectKnossos(resolver)

	if !health.Exists {
		t.Error("expected Exists=true")
	}
	if health.SatelliteRiteCount != 2 {
		t.Errorf("expected 2 satellite rites, got %d", health.SatelliteRiteCount)
	}
}

func TestCollectKnossos_NotExists(t *testing.T) {
	dir := t.TempDir()
	resolver := paths.NewResolver(dir)

	health := collectKnossos(resolver)
	if health.Exists {
		t.Error("expected Exists=false when .knossos/ missing")
	}
}

func TestCollectKnow_WithDomains(t *testing.T) {
	root := setupProject(t)
	knowDir := filepath.Join(root, ".know")
	os.MkdirAll(knowDir, 0755)

	// Write a fresh domain file
	freshContent := `---
domain: architecture
generated_at: ` + time.Now().UTC().Format(time.RFC3339) + `
expires_after: 14d
source_hash: ""
confidence: 0.85
format_version: "1.0"
---
# Architecture
`
	os.WriteFile(filepath.Join(knowDir, "architecture.md"), []byte(freshContent), 0644)

	// Write a stale domain file (expired long ago)
	staleContent := `---
domain: scar-tissue
generated_at: 2024-01-01T00:00:00Z
expires_after: 7d
source_hash: ""
confidence: 0.7
format_version: "1.0"
---
# Scar Tissue
`
	os.WriteFile(filepath.Join(knowDir, "scar-tissue.md"), []byte(staleContent), 0644)

	health := collectKnow(root)

	if !health.Exists {
		t.Error("expected Exists=true")
	}
	if health.DomainCount != 2 {
		t.Errorf("expected 2 domains, got %d", health.DomainCount)
	}
	if health.FreshCount != 1 {
		t.Errorf("expected 1 fresh, got %d", health.FreshCount)
	}
	if health.StaleCount != 1 {
		t.Errorf("expected 1 stale, got %d", health.StaleCount)
	}
}

func TestCollectKnow_NotExists(t *testing.T) {
	dir := t.TempDir()
	health := collectKnow(dir)
	if health.Exists {
		t.Error("expected Exists=false when .know/ missing")
	}
}

func TestCollectLedge_WithArtifacts(t *testing.T) {
	root := setupProject(t)
	resolver := paths.NewResolver(root)

	// Create ledge subdirectories with files
	for _, sub := range []struct {
		dir   string
		count int
	}{
		{"decisions", 3},
		{"specs", 2},
		{"reviews", 1},
		{"spikes", 0},
	} {
		dir := filepath.Join(root, ".ledge", sub.dir)
		os.MkdirAll(dir, 0755)
		for i := 0; i < sub.count; i++ {
			os.WriteFile(filepath.Join(dir, filepath.Base(sub.dir)+"-"+string(rune('0'+i))+".md"),
				[]byte("# artifact"), 0644)
		}
	}

	health := collectLedge(resolver)

	if !health.Exists {
		t.Error("expected Exists=true")
	}
	if health.DecisionCount != 3 {
		t.Errorf("expected 3 decisions, got %d", health.DecisionCount)
	}
	if health.SpecCount != 2 {
		t.Errorf("expected 2 specs, got %d", health.SpecCount)
	}
	if health.ReviewCount != 1 {
		t.Errorf("expected 1 review, got %d", health.ReviewCount)
	}
	if health.SpikeCount != 0 {
		t.Errorf("expected 0 spikes, got %d", health.SpikeCount)
	}
	if health.TotalCount != 6 {
		t.Errorf("expected 6 total, got %d", health.TotalCount)
	}
}

func TestCollectLedge_NotExists(t *testing.T) {
	dir := t.TempDir()
	resolver := paths.NewResolver(dir)

	health := collectLedge(resolver)
	if health.Exists {
		t.Error("expected Exists=false when .ledge/ missing")
	}
}

func TestCollectSOS_WithSessions(t *testing.T) {
	root := setupProject(t)
	resolver := paths.NewResolver(root)

	sessionsDir := filepath.Join(root, ".sos", "sessions")
	archiveDir := filepath.Join(root, ".sos", "archive")
	os.MkdirAll(sessionsDir, 0755)
	os.MkdirAll(archiveDir, 0755)

	// Write current session pointer
	os.WriteFile(filepath.Join(sessionsDir, ".current-session"),
		[]byte("session-20260301-143000-abc12345"), 0644)

	// Active session
	activeDir := filepath.Join(sessionsDir, "session-20260301-143000-abc12345")
	os.MkdirAll(activeDir, 0755)
	os.WriteFile(filepath.Join(activeDir, "SESSION_CONTEXT.md"),
		[]byte("---\nstatus: ACTIVE\n---\n"), 0644)

	// Parked session
	parkedDir := filepath.Join(sessionsDir, "session-20260228-100000-def67890")
	os.MkdirAll(parkedDir, 0755)
	os.WriteFile(filepath.Join(parkedDir, "SESSION_CONTEXT.md"),
		[]byte("---\nstatus: PARKED\n---\n"), 0644)

	// Archived session
	archivedDir := filepath.Join(archiveDir, "session-20260201-090000-ghi11111")
	os.MkdirAll(archivedDir, 0755)

	health := collectSOS(resolver)

	if !health.Exists {
		t.Error("expected Exists=true")
	}
	if health.ActiveCount != 1 {
		t.Errorf("expected 1 active, got %d", health.ActiveCount)
	}
	if health.ParkedCount != 1 {
		t.Errorf("expected 1 parked, got %d", health.ParkedCount)
	}
	if health.ArchivedCount != 1 {
		t.Errorf("expected 1 archived, got %d", health.ArchivedCount)
	}
	if health.TotalCount != 3 {
		t.Errorf("expected 3 total, got %d", health.TotalCount)
	}
	if health.CurrentSession != "session-20260301-143000-abc12345" {
		t.Errorf("expected current session, got %q", health.CurrentSession)
	}
}

func TestCollectSOS_NotExists(t *testing.T) {
	dir := t.TempDir()
	resolver := paths.NewResolver(dir)

	health := collectSOS(resolver)
	if health.Exists {
		t.Error("expected Exists=false when .sos/ missing")
	}
}

func TestReadSessionStatus(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"active", "---\nstatus: ACTIVE\n---\n", "ACTIVE"},
		{"parked", "---\nstatus: PARKED\n---\n", "PARKED"},
		{"quoted", "---\nstatus: \"ACTIVE\"\n---\n", "ACTIVE"},
		{"no frontmatter", "# no frontmatter\n", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.name+".md")
			os.WriteFile(path, []byte(tt.content), 0644)
			got := readSessionStatus(path)
			if got != tt.expected {
				t.Errorf("readSessionStatus(%q) = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestReadSessionStatus_MissingFile(t *testing.T) {
	got := readSessionStatus("/nonexistent/path/SESSION_CONTEXT.md")
	if got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

func TestHealthDashboard_Text(t *testing.T) {
	dashboard := HealthDashboard{
		Claude: ClaudeHealth{
			Exists:     true,
			ActiveRite: "10x-dev",
			AgentCount: 5,
			LastSync:   "2026-03-01T14:30:00Z",
		},
		Knossos: KnossosHealth{
			Exists:             true,
			SatelliteRiteCount: 1,
			SatelliteRites:     []string{"custom-dev"},
		},
		Know: KnowHealth{
			Exists:      true,
			DomainCount: 2,
			FreshCount:  1,
			StaleCount:  1,
		},
		Ledge: LedgeHealth{
			Exists:        true,
			DecisionCount: 3,
		},
		SOS: SOSHealth{
			Exists:         true,
			ActiveCount:    1,
			TotalCount:     1,
			CurrentSession: "session-20260301-abc",
		},
		Healthy: true,
	}

	text := dashboard.Text()

	// Verify key sections present
	for _, want := range []string{
		"Project Health Dashboard",
		".claude/",
		"10x-dev",
		"Agents:       5",
		".knossos/",
		"custom-dev",
		".know/",
		"2 (1 fresh, 1 stale)",
		".ledge/",
		"Decisions: 3",
		".sos/",
		"session-20260301-abc",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("Text() missing expected content %q", want)
		}
	}
}

func TestHealthDashboard_JSON(t *testing.T) {
	dashboard := HealthDashboard{
		Claude:  ClaudeHealth{Exists: true, ActiveRite: "10x-dev", AgentCount: 3},
		Knossos: KnossosHealth{Exists: false},
		Know:    KnowHealth{Exists: true, DomainCount: 1, FreshCount: 1},
		Ledge:   LedgeHealth{Exists: false},
		SOS:     SOSHealth{Exists: true, ActiveCount: 1, TotalCount: 1},
		Healthy: true,
	}

	data, err := json.Marshal(dashboard)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify top-level keys
	for _, key := range []string{"claude", "knossos", "know", "ledge", "sos", "healthy"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}

	// Verify claude section
	claude := parsed["claude"].(map[string]any)
	if claude["active_rite"] != "10x-dev" {
		t.Errorf("expected active_rite=10x-dev, got %v", claude["active_rite"])
	}
}

func TestHealthDashboard_Unhealthy(t *testing.T) {
	dashboard := HealthDashboard{
		Claude:  ClaudeHealth{Exists: false},
		Healthy: false,
		Errors:  []string{".claude/ directory not found"},
	}

	text := dashboard.Text()
	if !strings.Contains(text, "(not found)") {
		t.Error("Text() should show (not found) for missing .claude/")
	}
	if !strings.Contains(text, ".claude/ directory not found") {
		t.Error("Text() should show error message")
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"just now", time.Now(), "just now"},
		{"minutes ago", time.Now().Add(-30 * time.Minute), "30m ago"},
		{"hours ago", time.Now().Add(-5 * time.Hour), "5h ago"},
		{"days ago", time.Now().Add(-3 * 24 * time.Hour), "3d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.time)
			if got != tt.want {
				t.Errorf("formatAge() = %q, want %q", got, tt.want)
			}
		})
	}
}
