package registry

import (
	"testing"
)

// TestRegistryCompleteness verifies every RefKey constant resolves to a
// non-empty entry value in the map.
func TestRegistryCompleteness(t *testing.T) {
	allKeys := []RefKey{
		SkillConventions,
		SkillCommitBehavior,
		SkillAttributionGuard,
		AgentPythia,
		AgentMoirai,
		AgentConsultant,
		AgentContextEngineer,
		CLISessionFieldSet,
		CLISessionLog,
		CLISessionWrap,
		DromenaPark,
	}
	for _, key := range allKeys {
		entry, ok := entries[key]
		if !ok {
			t.Errorf("key %q has no entry in the registry map", key)
			continue
		}
		if entry.Value == "" {
			t.Errorf("key %q has empty Value in registry", key)
		}
	}
}

// TestRef_UnknownKey_Panics verifies that Ref panics for an unregistered key.
func TestRef_UnknownKey_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Ref to panic for unknown key, but it did not")
		}
	}()
	Ref("does.not.exist")
}

// TestRef_KnownKeys verifies Ref returns expected values for a sample of keys.
func TestRef_KnownKeys(t *testing.T) {
	cases := []struct {
		key      RefKey
		expected string
	}{
		{SkillConventions, "conventions"},
		{SkillCommitBehavior, "commit:behavior"},
		{AgentPythia, "pythia"},
		{AgentMoirai, "moirai"},
		{CLISessionWrap, "ari session wrap"},
		{DromenaPark, "/park"},
	}
	for _, tc := range cases {
		got := Ref(tc.key)
		if got != tc.expected {
			t.Errorf("Ref(%q) = %q, want %q", tc.key, got, tc.expected)
		}
	}
}

// TestRecovery_WithHint verifies Recovery returns a non-empty string for
// keys that have a recovery hint registered.
func TestRecovery_WithHint(t *testing.T) {
	hint := Recovery(SkillConventions)
	if hint == "" {
		t.Error("expected non-empty recovery hint for SkillConventions")
	}
}

// TestRecovery_WithoutHint verifies Recovery returns "" for entries that
// have no recovery hint.
func TestRecovery_WithoutHint(t *testing.T) {
	hint := Recovery(AgentPythia)
	if hint != "" {
		t.Errorf("expected empty recovery hint for AgentPythia, got %q", hint)
	}
}

// TestEntriesByCategory_Agents verifies that exactly 4 agent entries exist.
func TestEntriesByCategory_Agents(t *testing.T) {
	agents := EntriesByCategory(CategoryAgent)
	if len(agents) != 4 {
		t.Errorf("expected 4 agent entries, got %d", len(agents))
	}
}

// TestEntriesByCategory_Skills verifies that exactly 2 skill entries exist.
func TestEntriesByCategory_Skills(t *testing.T) {
	skills := EntriesByCategory(CategorySkill)
	if len(skills) != 3 {
		t.Errorf("expected 3 skill entries, got %d", len(skills))
	}
}

// TestThroughlineAgents_AllRegistered verifies that ThroughlineAgents returns
// exactly 4 entries, all with non-empty keys.
func TestThroughlineAgents_AllRegistered(t *testing.T) {
	agents := ThroughlineAgents()
	if len(agents) != 4 {
		t.Errorf("expected 4 throughline agents, got %d", len(agents))
	}
	for name, val := range agents {
		if name == "" {
			t.Error("throughline agent map contains empty key")
		}
		if !val {
			t.Errorf("throughline agent %q has false value (should always be true)", name)
		}
	}
}

// TestTaskDelegation_NoOps verifies the base format with no operations.
func TestTaskDelegation_NoOps(t *testing.T) {
	got := TaskDelegation(AgentMoirai)
	want := `Task(moirai, "<operation>")`
	if got != want {
		t.Errorf("TaskDelegation(AgentMoirai) = %q, want %q", got, want)
	}
}

// TestTaskDelegation_WithOps verifies that operation names are appended.
func TestTaskDelegation_WithOps(t *testing.T) {
	got := TaskDelegation(AgentMoirai, "park", "wrap")
	want := `Task(moirai, "<operation>") -- operations: park, wrap`
	if got != want {
		t.Errorf("TaskDelegation(AgentMoirai, park, wrap) = %q, want %q", got, want)
	}
}

// TestTaskDelegation_OtherAgent verifies TaskDelegation works with AgentPythia.
func TestTaskDelegation_OtherAgent(t *testing.T) {
	got := TaskDelegation(AgentPythia, "start-phase")
	want := `Task(pythia, "<operation>") -- operations: start-phase`
	if got != want {
		t.Errorf("TaskDelegation(AgentPythia, start-phase) = %q, want %q", got, want)
	}
}
