package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/hook/clewcontract"
)

// --- Test helpers ---

// makeSnapshotContext creates a minimal *Context for snapshot tests.
func makeSnapshotContext(phase, status, initiative, rite, complexity string) *Context {
	return &Context{
		SchemaVersion: "2.1",
		SessionID:     "session-20260226-120000-snaptest1",
		Status:        Status(status),
		CreatedAt:     time.Now().UTC(),
		Initiative:    initiative,
		Complexity:    complexity,
		ActiveRite:    rite,
		CurrentPhase:  phase,
		Body:          "\n## Blockers\nNone yet.\n",
	}
}

// makeSnapshotContextWithBlockers creates a context with real blocker content.
func makeSnapshotContextWithBlockers(phase, blocker string) *Context {
	ctx := makeSnapshotContext(phase, "ACTIVE", "test initiative", "ecosystem", "MODULE")
	ctx.Body = fmt.Sprintf("\n## Blockers\n%s\n", blocker)
	return ctx
}

// writeEventsJSONL writes TypedEvents to a temp events.jsonl and returns the path.
func writeEventsJSONL(t *testing.T, events []clewcontract.TypedEvent) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "events.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create events.jsonl: %v", err)
	}
	defer f.Close()

	for _, e := range events {
		line, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("failed to marshal event: %v", err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			t.Fatalf("failed to write event line: %v", err)
		}
	}
	return path
}

// makeTypedDecision creates a decision.recorded TypedEvent.
func makeTypedDecision(decision, rationale string) clewcontract.TypedEvent {
	return clewcontract.NewTypedDecisionRecordedEvent(decision, rationale, nil)
}

// makeTypedDelegation creates an agent.delegated TypedEvent for a named agent.
func makeTypedDelegation(agentName string) clewcontract.TypedEvent {
	return clewcontract.NewTypedAgentDelegatedEvent(clewcontract.SourceHook, agentName, "specialist", "", "")
}

// makeTypedSessionCreated creates a session.created TypedEvent.
func makeTypedSessionCreated(initiative string) clewcontract.TypedEvent {
	return clewcontract.NewTypedSessionCreatedEvent("session-test", initiative, "MODULE", "")
}

// --- Test: GenerateSnapshot ---

// TestGenerateSnapshot_OrchestratorFull verifies orchestrator gets timeline+decisions+blockers.
func TestGenerateSnapshot_OrchestratorFull(t *testing.T) {
	// Build a context with a real blocker.
	ctx := makeSnapshotContextWithBlockers("implementation", "- Safari 15 missing color-mix() support")

	// Write events: 3 curated + 1 decision.
	events := []clewcontract.TypedEvent{
		makeTypedSessionCreated("test initiative"),
		makeTypedDelegation("context-architect"),
		makeTypedDecision("chose CSS variables", "runtime perf"),
		makeTypedDelegation("integration-engineer"),
	}
	eventsPath := writeEventsJSONL(t, events)

	snap, err := GenerateSnapshot(ctx, eventsPath, SnapshotConfig{
		Role:      RoleOrchestrator,
		AgentName: "potnia",
	})
	if err != nil {
		t.Fatalf("GenerateSnapshot returned error: %v", err)
	}

	if snap.Role != RoleOrchestrator {
		t.Errorf("snap.Role = %q, want %q", snap.Role, RoleOrchestrator)
	}
	if len(snap.Timeline) != 4 {
		t.Errorf("snap.Timeline len = %d, want 4", len(snap.Timeline))
	}
	if len(snap.Decisions) != 1 {
		t.Errorf("snap.Decisions len = %d, want 1", len(snap.Decisions))
	}
	if snap.Decisions[0].Decision != "chose CSS variables" {
		t.Errorf("snap.Decisions[0].Decision = %q, want %q", snap.Decisions[0].Decision, "chose CSS variables")
	}
	if snap.Blockers == "" {
		t.Error("snap.Blockers is empty, want blocker content")
	}
}

// TestGenerateSnapshot_OrchestratorCapsAt10 verifies orchestrator timeline caps at 10 entries.
func TestGenerateSnapshot_OrchestratorCapsAt10(t *testing.T) {
	ctx := makeSnapshotContext("implementation", "ACTIVE", "test", "ecosystem", "MODULE")

	// Write 15 curated events.
	events := make([]clewcontract.TypedEvent, 15)
	for i := range events {
		events[i] = makeTypedSessionCreated(fmt.Sprintf("entry %d", i))
	}
	eventsPath := writeEventsJSONL(t, events)

	snap, err := GenerateSnapshot(ctx, eventsPath, SnapshotConfig{Role: RoleOrchestrator})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}
	if len(snap.Timeline) != 10 {
		t.Errorf("snap.Timeline len = %d, want 10 (capped)", len(snap.Timeline))
	}
}

// TestGenerateSnapshot_SpecialistScoping verifies specialist scopes to agent + last 5.
func TestGenerateSnapshot_SpecialistScoping(t *testing.T) {
	ctx := makeSnapshotContext("implementation", "ACTIVE", "test", "ecosystem", "MODULE")

	// 10 events, 2 of which mention "context-architect" by name (at positions 2 and 5).
	events := []clewcontract.TypedEvent{
		makeTypedSessionCreated("initiative"),      // 0 - not agent, not recent
		makeTypedDelegation("context-architect"),   // 1 - agent-scoped
		makeTypedSessionCreated("phase 2"),         // 2
		makeTypedDelegation("integration-engineer"), // 3
		makeTypedDelegation("context-architect"),   // 4 - agent-scoped, in recent (last 5)
		makeTypedSessionCreated("phase 3"),         // 5 - recent
		makeTypedDelegation("integration-engineer"), // 6 - recent
		makeTypedSessionCreated("commit"),          // 7 - recent
		makeTypedDelegation("integration-engineer"), // 8 - recent
		makeTypedSessionCreated("phase 4"),         // 9 - recent
	}
	eventsPath := writeEventsJSONL(t, events)

	snap, err := GenerateSnapshot(ctx, eventsPath, SnapshotConfig{
		Role:      RoleSpecialist,
		AgentName: "context-architect",
	})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}

	// recent = events[5..9] = 5 entries
	// agent-scoped outside recent = events[1] (context-architect at index 1, not in recent)
	// events[4] (context-architect) IS in recent (recent starts at index 5... wait: len=10, last 5 = [5,6,7,8,9])
	// So events[4] is NOT in recent. Both events[1] and events[4] are agent-scoped.
	// agent-scoped cap=3, we have 2 — both included.
	// total = 5 recent + 2 agent-scoped = 7 (under hard cap of 8)
	if len(snap.Timeline) != 7 {
		t.Errorf("snap.Timeline len = %d, want 7 (5 recent + 2 agent-scoped)", len(snap.Timeline))
	}
	// No decisions for specialist.
	if len(snap.Decisions) != 0 {
		t.Errorf("snap.Decisions len = %d, want 0 (specialist has no decisions)", len(snap.Decisions))
	}
}

// TestGenerateSnapshot_SpecialistCapsAt8 verifies specialist hard cap at 8 entries.
func TestGenerateSnapshot_SpecialistCapsAt8(t *testing.T) {
	ctx := makeSnapshotContext("implementation", "ACTIVE", "test", "ecosystem", "MODULE")

	// 20 events, 5 mention "context-architect" before the last 5 (recent).
	events := make([]clewcontract.TypedEvent, 20)
	for i := range events {
		if i < 5 {
			events[i] = makeTypedDelegation("context-architect")
		} else {
			events[i] = makeTypedSessionCreated(fmt.Sprintf("entry %d", i))
		}
	}
	eventsPath := writeEventsJSONL(t, events)

	snap, err := GenerateSnapshot(ctx, eventsPath, SnapshotConfig{
		Role:      RoleSpecialist,
		AgentName: "context-architect",
	})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}

	// recent = last 5 (events[15..19])
	// agent-scoped = events[0..4] (5 entries, capped at 3)
	// total = 5 + 3 = 8 (exactly at hard cap)
	if len(snap.Timeline) != 8 {
		t.Errorf("snap.Timeline len = %d, want 8 (5 recent + 3 agent-scoped, hard cap)", len(snap.Timeline))
	}
}

// TestGenerateSnapshot_BackgroundMinimal verifies background has no timeline/decisions/blockers.
func TestGenerateSnapshot_BackgroundMinimal(t *testing.T) {
	ctx := makeSnapshotContextWithBlockers("implementation", "- critical blocker")

	events := []clewcontract.TypedEvent{
		makeTypedSessionCreated("initiative"),
		makeTypedDecision("key decision", "rationale"),
	}
	eventsPath := writeEventsJSONL(t, events)

	snap, err := GenerateSnapshot(ctx, eventsPath, SnapshotConfig{
		Role:      RoleBackground,
		AgentName: "linter",
	})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}

	if len(snap.Timeline) != 0 {
		t.Errorf("snap.Timeline len = %d, want 0 (background)", len(snap.Timeline))
	}
	if len(snap.Decisions) != 0 {
		t.Errorf("snap.Decisions len = %d, want 0 (background)", len(snap.Decisions))
	}
	if snap.Blockers != "" {
		t.Errorf("snap.Blockers = %q, want empty (background)", snap.Blockers)
	}
}

// TestGenerateSnapshot_EmptyEventsFile verifies valid minimal output when events.jsonl absent.
func TestGenerateSnapshot_EmptyEventsFile(t *testing.T) {
	ctx := makeSnapshotContext("requirements", "ACTIVE", "test", "ecosystem", "MODULE")

	// Point to a non-existent events.jsonl.
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, "events.jsonl")

	snap, err := GenerateSnapshot(ctx, eventsPath, SnapshotConfig{Role: RoleOrchestrator})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}

	if snap.Phase != "requirements" {
		t.Errorf("snap.Phase = %q, want %q", snap.Phase, "requirements")
	}
	if len(snap.Timeline) != 0 {
		t.Errorf("snap.Timeline len = %d, want 0 (no events)", len(snap.Timeline))
	}
	if len(snap.Decisions) != 0 {
		t.Errorf("snap.Decisions len = %d, want 0 (no events)", len(snap.Decisions))
	}
}

// --- Test: RenderMarkdown ---

// TestRenderMarkdown_OrchestratorTemplate verifies orchestrator markdown matches spec Section 5.1.
func TestRenderMarkdown_OrchestratorTemplate(t *testing.T) {
	snap := &Snapshot{
		Phase:      "implementation",
		Complexity: "MODULE",
		Status:     "ACTIVE",
		Initiative: "Add dark mode",
		Rite:       "ecosystem",
		Role:       RoleOrchestrator,
		AgentName:  "potnia",
		Timeline: []TimelineEntry{
			{
				Time:     time.Date(0, 1, 1, 14, 3, 0, 0, time.UTC),
				Category: "SESSION ", // 8-char padded
				Summary:  "created: Add dark mode (MODULE)",
			},
		},
		Decisions: []DecisionEntry{
			{Decision: "CSS variables over styled-components", Rationale: "runtime perf"},
		},
		Blockers:    "- Safari 15 missing color-mix()",
		GeneratedAt: time.Now().UTC(),
	}

	md := snap.RenderMarkdown()

	// Must contain the spec header.
	if !strings.Contains(md, "## Session Context (auto-injected)") {
		t.Errorf("missing header: %q", md)
	}
	// Must contain Phase | Complexity | Status line.
	if !strings.Contains(md, "Phase: implementation | Complexity: MODULE | Status: ACTIVE") {
		t.Errorf("missing phase line: %q", md)
	}
	// Orchestrator shows Initiative AND Rite.
	if !strings.Contains(md, "Initiative: Add dark mode | Rite: ecosystem") {
		t.Errorf("missing initiative+rite line: %q", md)
	}
	// Timeline section.
	if !strings.Contains(md, "### Timeline (last 1)") {
		t.Errorf("missing timeline header: %q", md)
	}
	if !strings.Contains(md, "14:03 | SESSION  | created: Add dark mode (MODULE)") {
		t.Errorf("missing timeline entry: %q", md)
	}
	// Decisions section.
	if !strings.Contains(md, "### Decisions") {
		t.Errorf("missing decisions header: %q", md)
	}
	if !strings.Contains(md, "CSS variables over styled-components") {
		t.Errorf("missing decision text: %q", md)
	}
	// Blockers section.
	if !strings.Contains(md, "### Blockers") {
		t.Errorf("missing blockers header: %q", md)
	}
	if !strings.Contains(md, "Safari 15") {
		t.Errorf("missing blocker content: %q", md)
	}
}

// TestRenderMarkdown_SpecialistTemplate verifies specialist markdown matches spec Section 5.2.
func TestRenderMarkdown_SpecialistTemplate(t *testing.T) {
	snap := &Snapshot{
		Phase:      "implementation",
		Complexity: "MODULE",
		Status:     "ACTIVE",
		Initiative: "Add dark mode",
		Rite:       "ecosystem",
		Role:       RoleSpecialist,
		AgentName:  "integration-engineer",
		Timeline: []TimelineEntry{
			{
				Time:     time.Date(0, 1, 1, 14, 20, 0, 0, time.UTC),
				Category: "AGENT   ", // 8-char padded
				Summary:  "delegated context-architect: solution design",
			},
		},
		Blockers:    "- Safari 15 missing color-mix()",
		GeneratedAt: time.Now().UTC(),
	}

	md := snap.RenderMarkdown()

	// Specialist shows Initiative WITHOUT Rite.
	if strings.Contains(md, "| Rite:") {
		t.Errorf("specialist markdown should NOT contain rite field: %q", md)
	}
	if !strings.Contains(md, "Initiative: Add dark mode") {
		t.Errorf("missing initiative: %q", md)
	}
	// Timeline header says "(recent)" not "(last N)".
	if !strings.Contains(md, "### Timeline (recent)") {
		t.Errorf("missing specialist timeline header: %q", md)
	}
	// No decisions section for specialist.
	if strings.Contains(md, "### Decisions") {
		t.Errorf("specialist markdown should NOT contain decisions section: %q", md)
	}
	// Blockers present.
	if !strings.Contains(md, "### Blockers") {
		t.Errorf("missing blockers: %q", md)
	}
}

// TestRenderMarkdown_BackgroundTemplate verifies background markdown is single-line.
func TestRenderMarkdown_BackgroundTemplate(t *testing.T) {
	snap := &Snapshot{
		Phase:       "implementation",
		Complexity:  "MODULE",
		Status:      "ACTIVE",
		Role:        RoleBackground,
		AgentName:   "linter",
		GeneratedAt: time.Now().UTC(),
	}

	md := snap.RenderMarkdown()

	// Background: header + phase line only.
	if !strings.Contains(md, "## Session Context (auto-injected)") {
		t.Errorf("missing header: %q", md)
	}
	if !strings.Contains(md, "Phase: implementation | Complexity: MODULE | Status: ACTIVE") {
		t.Errorf("missing phase line: %q", md)
	}
	// No Initiative, Timeline, Decisions, Blockers.
	if strings.Contains(md, "Initiative:") {
		t.Errorf("background should not contain Initiative: %q", md)
	}
	if strings.Contains(md, "### Timeline") {
		t.Errorf("background should not contain timeline: %q", md)
	}
	if strings.Contains(md, "### Decisions") {
		t.Errorf("background should not contain decisions: %q", md)
	}
	if strings.Contains(md, "### Blockers") {
		t.Errorf("background should not contain blockers: %q", md)
	}
}

// TestRenderMarkdown_EmptySessionOmitsTimeline verifies empty events omits timeline section.
func TestRenderMarkdown_EmptySessionOmitsTimeline(t *testing.T) {
	snap := &Snapshot{
		Phase:       "requirements",
		Complexity:  "PATCH",
		Status:      "ACTIVE",
		Initiative:  "fix the bug",
		Role:        RoleOrchestrator,
		GeneratedAt: time.Now().UTC(),
		// Timeline empty — fresh session
	}

	md := snap.RenderMarkdown()

	if strings.Contains(md, "### Timeline") {
		t.Errorf("empty timeline should be omitted: %q", md)
	}
	// Header line always present.
	if !strings.Contains(md, "## Session Context (auto-injected)") {
		t.Errorf("missing header: %q", md)
	}
}

// TestRenderMarkdown_DecisionRationaletruncated verifies rationale truncated to 30 chars.
func TestRenderMarkdown_DecisionRationaletruncated(t *testing.T) {
	snap := &Snapshot{
		Phase:      "implementation",
		Complexity: "MODULE",
		Status:     "ACTIVE",
		Initiative: "test",
		Role:       RoleOrchestrator,
		Decisions: []DecisionEntry{
			{Decision: "chose approach A", Rationale: "this is a very long rationale exceeding thirty characters"},
		},
		GeneratedAt: time.Now().UTC(),
	}

	md := snap.RenderMarkdown()

	if !strings.Contains(md, "### Decisions") {
		t.Errorf("missing decisions section: %q", md)
	}
	// Rationale should be truncated — the full 56-char rationale should not appear.
	if strings.Contains(md, "this is a very long rationale exceeding thirty characters") {
		t.Errorf("rationale should be truncated in markdown: %q", md)
	}
}

// --- Test: RenderJSON ---

// TestRenderJSON_OrchestratorFields verifies orchestrator JSON has all fields.
func TestRenderJSON_OrchestratorFields(t *testing.T) {
	snap := &Snapshot{
		Phase:      "implementation",
		Complexity: "MODULE",
		Status:     "ACTIVE",
		Initiative: "Add dark mode",
		Rite:       "ecosystem",
		Role:       RoleOrchestrator,
		AgentName:  "potnia",
		Timeline: []TimelineEntry{
			{
				Time:     time.Date(0, 1, 1, 14, 3, 0, 0, time.UTC),
				Category: "SESSION ",
				Summary:  "created: Add dark mode (MODULE)",
			},
		},
		Decisions: []DecisionEntry{
			{Decision: "CSS vars", Rationale: "perf"},
		},
		Blockers:    "- Safari 15 issue",
		GeneratedAt: time.Now().UTC(),
	}

	raw, err := snap.RenderJSON()
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	requiredFields := []string{"role", "agent_name", "status", "initiative", "complexity",
		"current_phase", "active_rite", "timeline", "decisions", "blockers", "generated_at"}
	for _, f := range requiredFields {
		if _, ok := out[f]; !ok {
			t.Errorf("missing field %q in orchestrator JSON", f)
		}
	}

	if out["active_rite"] != "ecosystem" {
		t.Errorf("active_rite = %v, want ecosystem", out["active_rite"])
	}
	if out["role"] != "orchestrator" {
		t.Errorf("role = %v, want orchestrator", out["role"])
	}
}

// TestRenderJSON_SpecialistOmitsRiteAndDecisions verifies specialist JSON omits active_rite and decisions.
func TestRenderJSON_SpecialistOmitsRiteAndDecisions(t *testing.T) {
	snap := &Snapshot{
		Phase:      "implementation",
		Complexity: "MODULE",
		Status:     "ACTIVE",
		Initiative: "test",
		Rite:       "ecosystem",
		Role:       RoleSpecialist,
		AgentName:  "integration-engineer",
		Timeline: []TimelineEntry{
			{
				Time:     time.Date(0, 1, 1, 14, 3, 0, 0, time.UTC),
				Category: "SESSION ",
				Summary:  "created (MODULE)",
			},
		},
		Decisions:   []DecisionEntry{{Decision: "hidden decision", Rationale: "internal"}},
		Blockers:    "- some blocker",
		GeneratedAt: time.Now().UTC(),
	}

	raw, err := snap.RenderJSON()
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	// Specialist must NOT have active_rite or decisions.
	if _, ok := out["active_rite"]; ok {
		t.Errorf("specialist JSON should not contain active_rite, got: %v", out["active_rite"])
	}
	if _, ok := out["decisions"]; ok {
		t.Errorf("specialist JSON should not contain decisions, got: %v", out["decisions"])
	}
	// Specialist MUST have timeline and blockers.
	if _, ok := out["timeline"]; !ok {
		t.Errorf("specialist JSON should contain timeline")
	}
	if _, ok := out["blockers"]; !ok {
		t.Errorf("specialist JSON should contain blockers")
	}
}

// TestRenderJSON_BackgroundOmitsTimelineDecisionsBlockers verifies background JSON is minimal.
func TestRenderJSON_BackgroundOmitsTimelineDecisionsBlockers(t *testing.T) {
	snap := &Snapshot{
		Phase:       "implementation",
		Complexity:  "MODULE",
		Status:      "ACTIVE",
		Initiative:  "test",
		Rite:        "ecosystem",
		Role:        RoleBackground,
		AgentName:   "linter",
		GeneratedAt: time.Now().UTC(),
	}

	raw, err := snap.RenderJSON()
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	for _, forbidden := range []string{"active_rite", "timeline", "decisions", "blockers"} {
		if _, ok := out[forbidden]; ok {
			t.Errorf("background JSON should not contain %q, got value: %v", forbidden, out[forbidden])
		}
	}
	// Background must still have core fields.
	for _, required := range []string{"role", "status", "current_phase", "complexity"} {
		if _, ok := out[required]; !ok {
			t.Errorf("background JSON missing required field %q", required)
		}
	}
	if out["role"] != "background" {
		t.Errorf("role = %v, want background", out["role"])
	}
}

// --- Test: extractBlockers ---

// TestExtractBlockers_NoneYet verifies "None yet." returns empty string.
func TestExtractBlockers_NoneYet(t *testing.T) {
	body := "\n## Blockers\nNone yet.\n"
	result := extractBlockers(body)
	if result != "" {
		t.Errorf("extractBlockers('None yet.') = %q, want empty", result)
	}
}

// TestExtractBlockers_None verifies "None" returns empty string.
func TestExtractBlockers_None(t *testing.T) {
	body := "\n## Blockers\nNone\n"
	result := extractBlockers(body)
	if result != "" {
		t.Errorf("extractBlockers('None') = %q, want empty", result)
	}
}

// TestExtractBlockers_WithContent verifies real blocker content is returned.
func TestExtractBlockers_WithContent(t *testing.T) {
	body := "\n## Blockers\n- Safari 15 does not support color-mix()\n\n## Next Steps\n1. Fix it\n"
	result := extractBlockers(body)
	if result == "" {
		t.Error("extractBlockers with real content returned empty")
	}
	if !strings.Contains(result, "Safari 15") {
		t.Errorf("extractBlockers = %q, want Safari 15 content", result)
	}
	// Should NOT include the ## Next Steps section.
	if strings.Contains(result, "Next Steps") {
		t.Errorf("extractBlockers included content beyond ## Blockers section: %q", result)
	}
}

// TestExtractBlockers_AbsentSection verifies missing ## Blockers section returns empty.
func TestExtractBlockers_AbsentSection(t *testing.T) {
	body := "\n## Artifacts\n- PRD: pending\n"
	result := extractBlockers(body)
	if result != "" {
		t.Errorf("extractBlockers (absent section) = %q, want empty", result)
	}
}

// --- Test: extractDecisions ---

// TestExtractDecisions_V3Events verifies decision.recorded events are extracted correctly.
func TestExtractDecisions_V3Events(t *testing.T) {
	events := []clewcontract.TypedEvent{
		makeTypedSessionCreated("test"),
		makeTypedDecision("chose approach A", "speed wins over memory here"),
		makeTypedDelegation("context-architect"),
		makeTypedDecision("use CSS variables", "runtime perf"),
	}

	decisions := extractDecisions(events)
	if len(decisions) != 2 {
		t.Errorf("extractDecisions len = %d, want 2", len(decisions))
	}
	if decisions[0].Decision != "chose approach A" {
		t.Errorf("decisions[0].Decision = %q, want %q", decisions[0].Decision, "chose approach A")
	}
	if decisions[0].Rationale != "speed wins over memory here" {
		t.Errorf("decisions[0].Rationale = %q, want %q", decisions[0].Rationale, "speed wins over memory here")
	}
	if decisions[1].Decision != "use CSS variables" {
		t.Errorf("decisions[1].Decision = %q, want %q", decisions[1].Decision, "use CSS variables")
	}
}

// TestExtractDecisions_NoDecisions verifies empty slice when no decisions exist.
func TestExtractDecisions_NoDecisions(t *testing.T) {
	events := []clewcontract.TypedEvent{
		makeTypedSessionCreated("test"),
		makeTypedDelegation("context-architect"),
	}
	decisions := extractDecisions(events)
	if len(decisions) != 0 {
		t.Errorf("extractDecisions len = %d, want 0", len(decisions))
	}
}

// --- Test: agentNameInSummary ---

// TestAgentNameInSummary_Match verifies agent name substring match.
func TestAgentNameInSummary_Match(t *testing.T) {
	tests := []struct {
		summary   string
		agentName string
		want      bool
	}{
		{"delegated context-architect: solution design", "context-architect", true},
		{"completed integration-engineer: success", "integration-engineer", true},
		{"created: initiative (MODULE)", "integration-engineer", false},
		{"delegated architect: phase", "context-architect", false}, // substring false-positive prevention
		{"", "context-architect", false},
		{"delegated context-architect: task", "", false},
	}

	for _, tc := range tests {
		got := agentNameInSummary(tc.summary, tc.agentName)
		if got != tc.want {
			t.Errorf("agentNameInSummary(%q, %q) = %v, want %v",
				tc.summary, tc.agentName, got, tc.want)
		}
	}
}

// --- Test: readSnapshotEvents ---

// TestReadSnapshotEvents_SkipsV1V2 verifies v1 and v2 events are skipped.
func TestReadSnapshotEvents_SkipsV1V2(t *testing.T) {
	tmpDir := t.TempDir()
	eventsPath := filepath.Join(tmpDir, "events.jsonl")

	lines := []string{
		// v1 legacy (no "data" field)
		`{"timestamp":"2026-02-26T12:00:00Z","event":"session.created","from":"NONE","to":"ACTIVE"}`,
		// v2 flat (no "data" field)
		`{"ts":"2026-02-26T12:01:00.000Z","type":"agent.delegated","summary":"something"}`,
		// v3 typed (has "data" field)
		`{"ts":"2026-02-26T12:02:00.000Z","type":"decision.recorded","source":"agent","data":{"decision":"chose X","rationale":"speed"}}`,
	}
	if err := os.WriteFile(eventsPath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("failed to write events.jsonl: %v", err)
	}

	events, err := readSnapshotEvents(eventsPath)
	if err != nil {
		t.Fatalf("readSnapshotEvents error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("readSnapshotEvents len = %d, want 1 (only v3)", len(events))
	}
	if len(events) > 0 && events[0].Type != clewcontract.EventTypeDecisionRecorded {
		t.Errorf("events[0].Type = %q, want decision.recorded", events[0].Type)
	}
}

// TestReadSnapshotEvents_NonExistentFile verifies empty slice for missing file.
func TestReadSnapshotEvents_NonExistentFile(t *testing.T) {
	events, err := readSnapshotEvents("/nonexistent/path/events.jsonl")
	if err != nil {
		t.Fatalf("readSnapshotEvents on missing file returned error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("readSnapshotEvents len = %d, want 0 (missing file)", len(events))
	}
}

// --- Test: buildSpecialistTimeline deduplication ---

// TestBuildSpecialistTimeline_AgentScopedAlsoRecent verifies no duplicate when agent-scoped is in recent.
func TestBuildSpecialistTimeline_AgentScopedAlsoRecent(t *testing.T) {
	// 5 total events, last one mentions "context-architect".
	// It is both in recent (last 5) AND agent-scoped. Should appear only once.
	events := []clewcontract.TypedEvent{
		makeTypedSessionCreated("phase 1"),
		makeTypedSessionCreated("phase 2"),
		makeTypedSessionCreated("phase 3"),
		makeTypedSessionCreated("phase 4"),
		makeTypedDelegation("context-architect"), // recent AND agent-scoped
	}

	entries := buildSpecialistTimeline(events, "context-architect")

	// All 5 are recent (len=5, last 5 = all). The agent-scoped entry is in recent.
	// agent-scoped outside recent = none (event[4] is in recent).
	// merged = 0 agent-scoped + 5 recent = 5.
	if len(entries) != 5 {
		t.Errorf("buildSpecialistTimeline len = %d, want 5 (no duplicate)", len(entries))
	}
}

// TestBuildSpecialistTimeline_EmptyEvents verifies nil returned for empty events.
func TestBuildSpecialistTimeline_EmptyEvents(t *testing.T) {
	entries := buildSpecialistTimeline(nil, "context-architect")
	if entries != nil {
		t.Errorf("buildSpecialistTimeline(nil) = %v, want nil", entries)
	}
}
