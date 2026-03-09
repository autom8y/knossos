package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/paths"
)

// --- ReadSessionState Tests ---

func TestReadSessionState_ValidContext(t *testing.T) {
	dir := t.TempDir()
	contextPath := filepath.Join(dir, "SESSION_CONTEXT.md")
	content := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4
status: ACTIVE
created_at: "2026-03-08T14:30:22Z"
initiative: "Ariadne Intelligence -- H4"
complexity: MODULE
active_rite: 10x-dev
rite: 10x-dev
current_phase: implementation
---
# Session content
`
	require.NoError(t, os.WriteFile(contextPath, []byte(content), 0644))

	signals := ReadSessionState(contextPath)
	require.NotNil(t, signals)
	assert.Equal(t, "session-20260308-143022-a1b2c3d4", signals.SessionID)
	assert.Equal(t, "implementation", signals.Phase)
	assert.Equal(t, "10x-dev", signals.Rite)
	assert.Equal(t, "MODULE", signals.Complexity)
	assert.Equal(t, "Ariadne Intelligence -- H4", signals.Initiative)
}

func TestReadSessionState_MissingFile(t *testing.T) {
	signals := ReadSessionState("/nonexistent/path/SESSION_CONTEXT.md")
	assert.Nil(t, signals)
}

func TestReadSessionState_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	contextPath := filepath.Join(dir, "SESSION_CONTEXT.md")
	content := `---
this is not: valid: yaml: [[[
---
`
	require.NoError(t, os.WriteFile(contextPath, []byte(content), 0644))

	signals := ReadSessionState(contextPath)
	assert.Nil(t, signals)
}

func TestReadSessionState_EmptyPhase(t *testing.T) {
	dir := t.TempDir()
	contextPath := filepath.Join(dir, "SESSION_CONTEXT.md")
	content := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4
status: ACTIVE
created_at: "2026-03-08T14:30:22Z"
initiative: "Test"
complexity: TASK
active_rite: 10x-dev
current_phase: ""
---
`
	require.NoError(t, os.WriteFile(contextPath, []byte(content), 0644))

	signals := ReadSessionState(contextPath)
	require.NotNil(t, signals, "should return signals even with empty phase")
	assert.Equal(t, "", signals.Phase)
}

func TestReadSessionState_NoRiteField(t *testing.T) {
	dir := t.TempDir()
	contextPath := filepath.Join(dir, "SESSION_CONTEXT.md")
	content := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4
status: ACTIVE
created_at: "2026-03-08T14:30:22Z"
initiative: "Cross-cutting work"
complexity: INITIATIVE
active_rite: ecosystem
current_phase: design
---
`
	require.NoError(t, os.WriteFile(contextPath, []byte(content), 0644))

	signals := ReadSessionState(contextPath)
	require.NotNil(t, signals)
	assert.Equal(t, "ecosystem", signals.Rite, "should fall back to active_rite when rite is nil")
}

func TestReadSessionState_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	contextPath := filepath.Join(dir, "SESSION_CONTEXT.md")
	content := `# No frontmatter here
Just plain text.
`
	require.NoError(t, os.WriteFile(contextPath, []byte(content), 0644))

	signals := ReadSessionState(contextPath)
	assert.Nil(t, signals)
}

// --- TailReadEvents Tests ---

func TestTailReadEvents_Normal(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")

	var lines []string
	for range 10 {
		evt := map[string]any{
			"ts":      "2026-03-08T14:30:22.000Z",
			"type":    "tool.file_change",
			"summary": "Changed file",
		}
		data, _ := json.Marshal(evt)
		lines = append(lines, string(data))
	}
	// Add an agent task start event.
	agentEvt := map[string]any{
		"ts":      "2026-03-08T14:35:00.000Z",
		"type":    "agent.task_start",
		"summary": "Task started",
		"meta":    map[string]any{"agent": "qa-adversary"},
	}
	data, _ := json.Marshal(agentEvt)
	lines = append(lines, string(data))

	// Add a phase transition event.
	phaseEvt := map[string]any{
		"ts":      "2026-03-08T14:36:00.000Z",
		"type":    "phase.transitioned",
		"summary": "Phase change",
	}
	data, _ = json.Marshal(phaseEvt)
	lines = append(lines, string(data))

	require.NoError(t, os.WriteFile(eventsPath, []byte(strings.Join(lines, "\n")+"\n"), 0644))

	summary := TailReadEvents(eventsPath, 50)
	require.NotNil(t, summary)
	assert.Equal(t, 10, summary.FileChangeCount)
	assert.Equal(t, 1, summary.AgentTasks["qa-adversary"])
	assert.Equal(t, 1, summary.PhaseTransitions)
	assert.Equal(t, "2026-03-08T14:36:00.000Z", summary.LastEventTS)
}

func TestTailReadEvents_LargeFile(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")

	// Create a file larger than tailBufferSize.
	var lines []string
	for range 200 {
		evt := map[string]any{
			"ts":      "2026-03-08T14:30:22.000Z",
			"type":    "tool.file_change",
			"summary": "Changed file " + strings.Repeat("x", 100), // pad to make lines larger
		}
		data, _ := json.Marshal(evt)
		lines = append(lines, string(data))
	}
	content := strings.Join(lines, "\n") + "\n"
	require.NoError(t, os.WriteFile(eventsPath, []byte(content), 0644))

	// Verify file is large enough.
	stat, err := os.Stat(eventsPath)
	require.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(tailBufferSize), "test file should exceed tail buffer")

	summary := TailReadEvents(eventsPath, 50)
	require.NotNil(t, summary)
	// Should have parsed at most 50 events (lineCap).
	assert.LessOrEqual(t, summary.FileChangeCount, 50)
	assert.Greater(t, summary.FileChangeCount, 0)
}

func TestTailReadEvents_MissingFile(t *testing.T) {
	summary := TailReadEvents("/nonexistent/events.jsonl", 50)
	assert.Nil(t, summary)
}

func TestTailReadEvents_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")
	require.NoError(t, os.WriteFile(eventsPath, []byte(""), 0644))

	summary := TailReadEvents(eventsPath, 50)
	assert.Nil(t, summary)
}

func TestTailReadEvents_PartialLastLine(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")

	// Write complete events + a partial line at the end.
	evt := map[string]any{
		"ts":   "2026-03-08T14:30:22.000Z",
		"type": "tool.file_change",
	}
	data, _ := json.Marshal(evt)
	content := string(data) + "\n" + `{"ts":"2026-03-08T14:31:00` // partial
	require.NoError(t, os.WriteFile(eventsPath, []byte(content), 0644))

	summary := TailReadEvents(eventsPath, 50)
	require.NotNil(t, summary)
	assert.Equal(t, 1, summary.FileChangeCount, "should parse complete event, skip partial")
}

func TestTailReadEvents_ConfigurableLineCap(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")

	var lines []string
	for range 20 {
		evt := map[string]any{
			"ts":   "2026-03-08T14:30:22.000Z",
			"type": "tool.file_change",
		}
		data, _ := json.Marshal(evt)
		lines = append(lines, string(data))
	}
	require.NoError(t, os.WriteFile(eventsPath, []byte(strings.Join(lines, "\n")+"\n"), 0644))

	// Cap at 5 lines.
	summary := TailReadEvents(eventsPath, 5)
	require.NotNil(t, summary)
	assert.Equal(t, 5, summary.FileChangeCount, "should respect custom line cap")
}

func TestTailReadEvents_MixedFormats(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")

	// v2 format event.
	v2 := `{"ts":"2026-03-08T14:30:22.000Z","type":"tool.file_change","summary":"Changed"}`
	// v1 legacy format event.
	v1 := `{"timestamp":"2026-03-08T14:31:00.000Z","event":"tool.file_change","summary":"Old format"}`

	content := v2 + "\n" + v1 + "\n"
	require.NoError(t, os.WriteFile(eventsPath, []byte(content), 0644))

	summary := TailReadEvents(eventsPath, 50)
	require.NotNil(t, summary)
	assert.Equal(t, 2, summary.FileChangeCount, "should parse both v1 and v2 events")
}

// --- sessionScoreModifier Tests ---

func TestSessionScoreModifier_NilSignals(t *testing.T) {
	entry := SearchEntry{Name: "test", Domain: DomainAgent}
	mod := sessionScoreModifier(entry, "keyword", 100, nil)
	assert.Equal(t, 0, mod)
}

func TestSessionScoreModifier_PhaseBoost_Implementation(t *testing.T) {
	signals := &SessionSignals{Phase: "implementation"}
	entry := SearchEntry{
		Name:   "principal-engineer",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, phaseBoostAmount, mod, "implementation phase should boost principal-engineer")
}

func TestSessionScoreModifier_PhaseBoost_Requirements(t *testing.T) {
	signals := &SessionSignals{Phase: "requirements"}
	entry := SearchEntry{
		Name:   "requirements-analyst",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, phaseBoostAmount, mod, "requirements phase should boost requirements-analyst")
}

func TestSessionScoreModifier_PhaseBoost_Design(t *testing.T) {
	signals := &SessionSignals{Phase: "design"}
	entry := SearchEntry{
		Name:   "architect",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, phaseBoostAmount, mod, "design phase should boost architect")
}

func TestSessionScoreModifier_PhaseBoost_Validation(t *testing.T) {
	signals := &SessionSignals{Phase: "validation"}
	entry := SearchEntry{
		Name:   "qa-adversary",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, phaseBoostAmount, mod, "validation phase should boost qa-adversary")
}

func TestSessionScoreModifier_PhaseBoost_AllDomains(t *testing.T) {
	// Phase boost should apply to entries in any domain, not just agents.
	signals := &SessionSignals{Phase: "implementation"}
	tests := []struct {
		name   string
		entry  SearchEntry
		expect int
	}{
		{
			name:   "agent domain",
			entry:  SearchEntry{Name: "principal-engineer", Domain: DomainAgent},
			expect: phaseBoostAmount,
		},
		{
			name:   "command domain with build keyword",
			entry:  SearchEntry{Name: "build-tool", Domain: DomainCommand, Keywords: []string{"build"}},
			expect: phaseBoostAmount,
		},
		{
			name:   "dromena with test keyword",
			entry:  SearchEntry{Name: "run-tests", Domain: DomainDromena, Summary: "Run test suite"},
			expect: phaseBoostAmount,
		},
		{
			name:   "unrelated entry",
			entry:  SearchEntry{Name: "session-manager", Domain: DomainConcept, Summary: "Manages sessions"},
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := sessionScoreModifier(tt.entry, "keyword", 100, signals)
			assert.Equal(t, tt.expect, mod)
		})
	}
}

func TestSessionScoreModifier_PhaseBoost_UnknownPhase(t *testing.T) {
	signals := &SessionSignals{Phase: "unknown-phase"}
	entry := SearchEntry{
		Name:   "principal-engineer",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, 0, mod, "unknown phase should produce no boost")
}

func TestSessionScoreModifier_ComplexityPenalty_TASK(t *testing.T) {
	signals := &SessionSignals{Complexity: "TASK"}
	entry := SearchEntry{
		Name:    "orchestration-workflow",
		Domain:  DomainRouting,
		Summary: "Multi-agent orchestration for initiatives",
	}

	mod := sessionScoreModifier(entry, "keyword", 250, signals)
	assert.Equal(t, -complexityPenaltyAmount, mod, "TASK complexity should penalize orchestration routing")
}

func TestSessionScoreModifier_ComplexityPenalty_INITIATIVE(t *testing.T) {
	signals := &SessionSignals{Complexity: "INITIATIVE"}
	entry := SearchEntry{
		Name:    "orchestration-workflow",
		Domain:  DomainRouting,
		Summary: "Multi-agent orchestration for initiatives",
	}

	mod := sessionScoreModifier(entry, "keyword", 250, signals)
	assert.Equal(t, 0, mod, "INITIATIVE complexity should not penalize")
}

func TestSessionScoreModifier_ComplexityPenalty_ExactMatchProtected(t *testing.T) {
	signals := &SessionSignals{Complexity: "TASK"}
	entry := SearchEntry{
		Name:   "orchestration",
		Domain: DomainRouting,
	}

	// Exact match at base score 1000. Penalty of 100 would take it to 900.
	// Tier floor is 900, so modifier should be -100 (floor is met).
	mod := sessionScoreModifier(entry, "exact", 1000, signals)
	assert.GreaterOrEqual(t, 1000+mod, tierFloorExact, "exact match should stay >= tier floor")
}

func TestSessionScoreModifier_ComplexityPenalty_PrefixMatchProtected(t *testing.T) {
	signals := &SessionSignals{Complexity: "TASK"}
	entry := SearchEntry{
		Name:   "orchestration-workflow",
		Domain: DomainRouting,
	}

	// Prefix match at base score 500. Penalty of 100 would take it to 400.
	// But tier floor is 500, so modifier should be 0 (clamp to floor).
	mod := sessionScoreModifier(entry, "prefix", 500, signals)
	assert.GreaterOrEqual(t, 500+mod, tierFloorPrefix, "prefix match should stay >= tier floor")
}

func TestSessionScoreModifier_ComplexityPenalty_NonRoutingDomain(t *testing.T) {
	signals := &SessionSignals{Complexity: "TASK"}
	entry := SearchEntry{
		Name:    "orchestration",
		Domain:  DomainConcept,
		Summary: "Multi-agent orchestration concept",
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, 0, mod, "complexity penalty only applies to routing domain")
}

func TestSessionScoreModifier_ActivityBoost_FileChanges(t *testing.T) {
	signals := &SessionSignals{
		Activity: &ActivitySummary{
			FileChangeCount: 10,
			AgentTasks:      map[string]int{},
		},
	}
	entry := SearchEntry{
		Name:   "principal-engineer",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, activityBoostAmount, mod, "file changes > threshold should boost implementation entries")
}

func TestSessionScoreModifier_ActivityBoost_FileChanges_BelowThreshold(t *testing.T) {
	signals := &SessionSignals{
		Activity: &ActivitySummary{
			FileChangeCount: 3,
			AgentTasks:      map[string]int{},
		},
	}
	entry := SearchEntry{
		Name:   "principal-engineer",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, 0, mod, "file changes below threshold should not boost")
}

func TestSessionScoreModifier_ActivityBoost_QAAgent(t *testing.T) {
	signals := &SessionSignals{
		Activity: &ActivitySummary{
			AgentTasks: map[string]int{"qa-adversary": 3},
		},
	}
	entry := SearchEntry{
		Name:   "qa-adversary",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, activityBoostAmount, mod, "QA agent activity should boost validation entries")
}

func TestSessionScoreModifier_ActivityBoost_ArchitectAgent(t *testing.T) {
	signals := &SessionSignals{
		Activity: &ActivitySummary{
			AgentTasks: map[string]int{"architect": 2},
		},
	}
	entry := SearchEntry{
		Name:   "architect",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, activityBoostAmount, mod, "architect agent activity should boost design entries")
}

func TestSessionScoreModifier_ActivityBoost_SecondaryToPhase(t *testing.T) {
	// When both phase and activity boost target the same entry,
	// phase wins (max, not sum).
	signals := &SessionSignals{
		Phase: "implementation",
		Activity: &ActivitySummary{
			FileChangeCount: 10,
			AgentTasks:      map[string]int{},
		},
	}
	entry := SearchEntry{
		Name:   "principal-engineer",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, phaseBoostAmount, mod, "phase boost should be used when both apply (max, not sum)")
}

func TestSessionScoreModifier_NoDoubleBoost(t *testing.T) {
	signals := &SessionSignals{
		Phase: "implementation",
		Activity: &ActivitySummary{
			FileChangeCount: 10,
			AgentTasks:      map[string]int{},
		},
	}
	entry := SearchEntry{
		Name:   "principal-engineer",
		Domain: DomainAgent,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	// Should be max(150, 75) = 150, NOT 150 + 75 = 225.
	assert.Equal(t, phaseBoostAmount, mod, "should use max, not sum")
	assert.NotEqual(t, phaseBoostAmount+activityBoostAmount, mod, "should NOT sum boosts")
}

func TestSessionScoreModifier_ActivityNil(t *testing.T) {
	signals := &SessionSignals{
		Phase:    "implementation",
		Activity: nil,
	}
	entry := SearchEntry{
		Name:   "session-manager",
		Domain: DomainConcept,
	}

	mod := sessionScoreModifier(entry, "keyword", 100, signals)
	assert.Equal(t, 0, mod, "nil activity with non-matching phase should produce no modification")
}

func TestSessionScoreModifier_PhaseAndComplexity(t *testing.T) {
	// Phase boost + complexity penalty should combine correctly.
	signals := &SessionSignals{
		Phase:      "implementation",
		Complexity: "TASK",
	}
	// A routing entry with implementation keyword AND orchestration keyword.
	entry := SearchEntry{
		Name:     "build-orchestration",
		Domain:   DomainRouting,
		Summary:  "Build orchestration pipeline",
		Keywords: []string{"build", "orchestration"},
	}

	mod := sessionScoreModifier(entry, "keyword", 200, signals)
	// Phase boost: +150 (matches "build" keyword in implementation phase)
	// Complexity penalty: -100 (routing + orchestration in TASK session)
	// Net: +50
	assert.Equal(t, phaseBoostAmount-complexityPenaltyAmount, mod)
}

// --- entryMatchesKeywords Tests ---

func TestEntryMatchesKeywords_NameMatch(t *testing.T) {
	entry := SearchEntry{Name: "principal-engineer"}
	assert.True(t, entryMatchesKeywords(entry, []string{"principal-engineer"}))
}

func TestEntryMatchesKeywords_KeywordMatch(t *testing.T) {
	entry := SearchEntry{
		Name:     "test-runner",
		Keywords: []string{"testing", "coverage"},
	}
	assert.True(t, entryMatchesKeywords(entry, []string{"coverage"}))
}

func TestEntryMatchesKeywords_SummaryMatch(t *testing.T) {
	entry := SearchEntry{
		Name:    "my-tool",
		Summary: "Helps with implementation tasks",
	}
	assert.True(t, entryMatchesKeywords(entry, []string{"implement"}))
}

func TestEntryMatchesKeywords_CaseInsensitive(t *testing.T) {
	entry := SearchEntry{Name: "Principal-Engineer"}
	assert.True(t, entryMatchesKeywords(entry, []string{"principal-engineer"}))
}

func TestEntryMatchesKeywords_NoMatch(t *testing.T) {
	entry := SearchEntry{
		Name:     "session-manager",
		Keywords: []string{"session", "lifecycle"},
		Summary:  "Manages session state",
	}
	assert.False(t, entryMatchesKeywords(entry, []string{"build", "compile", "test"}))
}

// --- CollectParkedSessions Tests ---

func TestCollectParkedSessions_Found(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, "session-20260308-143022-a1b2c3d4-extra")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	contextContent := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4-extra
status: PARKED
created_at: "2026-03-08T14:30:22Z"
initiative: "Ariadne Intelligence -- H1"
complexity: MODULE
active_rite: 10x-dev
current_phase: design
parked_reason: "switching to H4"
---
`
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644))

	resolver := paths.NewResolver(projectRoot)
	entries := CollectParkedSessions(resolver)
	require.Len(t, entries, 1)
	assert.Equal(t, "Ariadne Intelligence -- H1", entries[0].Name)
	assert.Equal(t, DomainSession, entries[0].Domain)
	assert.Contains(t, entries[0].Summary, "Parked session on 10x-dev")
	assert.Contains(t, entries[0].Summary, "switching to H4")
	assert.Contains(t, entries[0].Action, "ari session resume")
	assert.False(t, entries[0].Boosted, "parked sessions should not be boosted")
}

func TestCollectParkedSessions_DomainIsSession(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, "session-20260308-143022-a1b2c3d4-abcd")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	contextContent := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4-abcd
status: PARKED
created_at: "2026-03-08T14:30:22Z"
initiative: "Test"
complexity: TASK
active_rite: ecosystem
current_phase: requirements
---
`
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644))

	resolver := paths.NewResolver(projectRoot)
	entries := CollectParkedSessions(resolver)
	require.Len(t, entries, 1)
	assert.Equal(t, DomainSession, entries[0].Domain)
}

func TestCollectParkedSessions_SkipsActiveSession(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, "session-20260308-143022-a1b2c3d4-efgh")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	contextContent := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4-efgh
status: ACTIVE
created_at: "2026-03-08T14:30:22Z"
initiative: "Active Work"
complexity: MODULE
active_rite: 10x-dev
current_phase: implementation
---
`
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644))

	resolver := paths.NewResolver(projectRoot)
	entries := CollectParkedSessions(resolver)
	assert.Empty(t, entries, "should not include ACTIVE sessions")
}

func TestCollectParkedSessions_UnreadableDir(t *testing.T) {
	resolver := paths.NewResolver("/nonexistent/project")
	entries := CollectParkedSessions(resolver)
	assert.Nil(t, entries)
}

func TestCollectParkedSessions_EmptyInitiative(t *testing.T) {
	projectRoot := t.TempDir()
	sessionsDir := filepath.Join(projectRoot, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, "session-20260308-143022-a1b2c3d4-ijkl")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	contextContent := `---
schema_version: "2.3"
session_id: session-20260308-143022-a1b2c3d4-ijkl
status: PARKED
created_at: "2026-03-08T14:30:22Z"
initiative: ""
complexity: TASK
active_rite: ecosystem
current_phase: requirements
---
`
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644))

	resolver := paths.NewResolver(projectRoot)
	entries := CollectParkedSessions(resolver)
	require.Len(t, entries, 1)
	assert.Equal(t, "session-20260308-143022-a1b2c3d4-ijkl", entries[0].Name, "should use sessionID as fallback")
	assert.Contains(t, entries[0].Summary, "unknown initiative")
}

func TestCollectParkedSessions_NoResolver(t *testing.T) {
	entries := CollectParkedSessions(nil)
	assert.Nil(t, entries)
}

// --- Search Integration Tests ---

func TestSearch_WithSessionSignals_PhaseBoost(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{
				Name:     "principal-engineer",
				Domain:   DomainAgent,
				Summary:  "Implements code according to design specifications",
				Keywords: []string{"implementation", "build", "code"},
			},
			{
				Name:     "requirements-analyst",
				Domain:   DomainAgent,
				Summary:  "Gathers requirements and produces PRD artifacts",
				Keywords: []string{"requirements", "prd", "scope"},
			},
		},
	}

	// Use a query that produces base scores for both entries.
	// "code" matches principal-engineer keywords; "scope" matches requirements-analyst keywords.
	// We test with "code" -- PE should match without session, and get boosted with session.

	// Without session.
	noSessionResults := idx.Search("code", SearchOptions{Limit: 10})

	// With implementation phase: PE should get boosted.
	signals := &SessionSignals{Phase: "implementation"}
	sessionResults := idx.Search("code", SearchOptions{
		Limit:   10,
		Session: signals,
	})

	// Find principal-engineer scores in both.
	var peNoSession, peWithSession int
	for _, r := range noSessionResults {
		if r.Name == "principal-engineer" {
			peNoSession = r.Score
		}
	}
	for _, r := range sessionResults {
		if r.Name == "principal-engineer" {
			peWithSession = r.Score
		}
	}

	require.Greater(t, peNoSession, 0, "PE should have a base score for 'code' query")
	assert.Greater(t, peWithSession, peNoSession,
		"principal-engineer should score higher with implementation phase")
}

func TestSearch_WithoutSession_NoBehaviorChange(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session management"},
		},
	}

	// Nil session should produce identical results to pre-H4.
	results := idx.Search("session", SearchOptions{Limit: 5, Session: nil})
	require.Len(t, results, 1)
	assert.Equal(t, 1000, results[0].Score, "nil session should not change scores")
	assert.Equal(t, "exact", results[0].MatchType)
}

func TestSearch_SessionDomainFilter(t *testing.T) {
	idx := &SearchIndex{
		entries: []SearchEntry{
			{Name: "session", Domain: DomainConcept, Summary: "Session concept"},
			{Name: "parked-work", Domain: DomainSession, Summary: "Parked session"},
		},
	}

	results := idx.Search("session", SearchOptions{
		Limit:   10,
		Domains: []Domain{DomainSession},
	})

	// Only session-domain entries should appear.
	for _, r := range results {
		assert.Equal(t, DomainSession, r.Domain)
	}
}

// --- Table-Driven Score Modifier Tests ---

func TestSessionScoreModifier_Table(t *testing.T) {
	tests := []struct {
		name      string
		entry     SearchEntry
		matchType string
		baseScore int
		signals   *SessionSignals
		wantMod   int
	}{
		{
			name:      "nil signals returns 0",
			entry:     SearchEntry{Name: "test", Domain: DomainAgent},
			matchType: "keyword",
			baseScore: 100,
			signals:   nil,
			wantMod:   0,
		},
		{
			name:      "empty phase returns 0",
			entry:     SearchEntry{Name: "principal-engineer", Domain: DomainAgent},
			matchType: "keyword",
			baseScore: 100,
			signals:   &SessionSignals{Phase: ""},
			wantMod:   0,
		},
		{
			name:      "TASK penalty on routing with orchestration keyword",
			entry:     SearchEntry{Name: "coordinator", Domain: DomainRouting, Keywords: []string{"orchestration"}},
			matchType: "keyword",
			baseScore: 200,
			signals:   &SessionSignals{Complexity: "TASK"},
			wantMod:   -complexityPenaltyAmount,
		},
		{
			name:      "MODULE complexity no penalty",
			entry:     SearchEntry{Name: "coordinator", Domain: DomainRouting, Keywords: []string{"orchestration"}},
			matchType: "keyword",
			baseScore: 200,
			signals:   &SessionSignals{Complexity: "MODULE"},
			wantMod:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sessionScoreModifier(tt.entry, tt.matchType, tt.baseScore, tt.signals)
			assert.Equal(t, tt.wantMod, got)
		})
	}
}
