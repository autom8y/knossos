package rite

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/autom8y/ariadne/internal/paths"
)

func TestStateManager_LoadSaveRoundTrip(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tempDir)
	manager := NewStateManager(resolver)

	// Create state with invocations
	state := &InvocationState{
		SchemaVersion: "1.0",
		CurrentRite:   "test-rite",
		LastUpdated:   time.Now().UTC(),
		Invocations: []Invocation{
			{
				ID:        "inv-001",
				RiteName:  "other-rite",
				Component: "skills",
				Skills:    []string{"skill1", "skill2"},
				InvokedAt: time.Now().UTC(),
			},
		},
		Budget: StateBudget{
			NativeTokens:   5000,
			BorrowedTokens: 2000,
			TotalTokens:    7000,
			BudgetLimit:    50000,
		},
	}

	// Save
	if err := manager.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	statePath := resolver.InvocationStateFile()
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file was not created")
	}

	// Load
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded state
	if loaded.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", loaded.SchemaVersion, "1.0")
	}
	if loaded.CurrentRite != "test-rite" {
		t.Errorf("CurrentRite = %q, want %q", loaded.CurrentRite, "test-rite")
	}
	if len(loaded.Invocations) != 1 {
		t.Errorf("len(Invocations) = %d, want 1", len(loaded.Invocations))
	}
	if loaded.Invocations[0].ID != "inv-001" {
		t.Errorf("Invocations[0].ID = %q, want %q", loaded.Invocations[0].ID, "inv-001")
	}
	if loaded.Budget.TotalTokens != 7000 {
		t.Errorf("Budget.TotalTokens = %d, want 7000", loaded.Budget.TotalTokens)
	}
}

func TestStateManager_LoadMissing(t *testing.T) {
	tempDir := t.TempDir()
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	resolver := paths.NewResolver(tempDir)
	manager := NewStateManager(resolver)

	// Load should return empty state when file doesn't exist
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if state == nil {
		t.Fatal("Load returned nil state")
	}
	if state.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", state.SchemaVersion, "1.0")
	}
	if len(state.Invocations) != 0 {
		t.Errorf("len(Invocations) = %d, want 0", len(state.Invocations))
	}
}

func TestInvocationState_AddInvocation(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{},
	}

	inv := Invocation{
		ID:       "inv-001",
		RiteName: "test-rite",
	}

	state.AddInvocation(inv)

	if len(state.Invocations) != 1 {
		t.Errorf("len(Invocations) = %d, want 1", len(state.Invocations))
	}
	if state.Invocations[0].ID != "inv-001" {
		t.Errorf("Invocations[0].ID = %q, want %q", state.Invocations[0].ID, "inv-001")
	}
}

func TestInvocationState_RemoveInvocation(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{
			{ID: "inv-001", RiteName: "rite1"},
			{ID: "inv-002", RiteName: "rite2"},
			{ID: "inv-003", RiteName: "rite3"},
		},
	}

	// Remove middle invocation
	removed := state.RemoveInvocation("inv-002")
	if removed == nil {
		t.Fatal("RemoveInvocation returned nil")
	}
	if removed.ID != "inv-002" {
		t.Errorf("Removed ID = %q, want %q", removed.ID, "inv-002")
	}
	if len(state.Invocations) != 2 {
		t.Errorf("len(Invocations) = %d, want 2", len(state.Invocations))
	}

	// Verify remaining invocations
	if state.Invocations[0].ID != "inv-001" || state.Invocations[1].ID != "inv-003" {
		t.Error("Wrong invocations remaining after removal")
	}

	// Remove nonexistent
	notFound := state.RemoveInvocation("inv-999")
	if notFound != nil {
		t.Error("RemoveInvocation(nonexistent) should return nil")
	}
}

func TestInvocationState_RemoveByRite(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{
			{ID: "inv-001", RiteName: "rite-a"},
			{ID: "inv-002", RiteName: "rite-b"},
			{ID: "inv-003", RiteName: "rite-a"},
		},
	}

	removed := state.RemoveByRite("rite-a")
	if len(removed) != 2 {
		t.Errorf("len(removed) = %d, want 2", len(removed))
	}
	if len(state.Invocations) != 1 {
		t.Errorf("len(Invocations) = %d, want 1", len(state.Invocations))
	}
	if state.Invocations[0].RiteName != "rite-b" {
		t.Error("Wrong invocation remaining")
	}
}

func TestInvocationState_RemoveAll(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{
			{ID: "inv-001"},
			{ID: "inv-002"},
		},
	}

	removed := state.RemoveAll()
	if len(removed) != 2 {
		t.Errorf("len(removed) = %d, want 2", len(removed))
	}
	if len(state.Invocations) != 0 {
		t.Errorf("len(Invocations) = %d, want 0", len(state.Invocations))
	}
}

func TestInvocationState_FindByID(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{
			{ID: "inv-001", RiteName: "rite1"},
			{ID: "inv-002", RiteName: "rite2"},
		},
	}

	found := state.FindByID("inv-002")
	if found == nil {
		t.Fatal("FindByID returned nil")
	}
	if found.RiteName != "rite2" {
		t.Errorf("Found wrong invocation: RiteName = %q, want %q", found.RiteName, "rite2")
	}

	notFound := state.FindByID("inv-999")
	if notFound != nil {
		t.Error("FindByID(nonexistent) should return nil")
	}
}

func TestInvocationState_FindByRite(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{
			{ID: "inv-001", RiteName: "rite-a"},
			{ID: "inv-002", RiteName: "rite-b"},
			{ID: "inv-003", RiteName: "rite-a"},
		},
	}

	found := state.FindByRite("rite-a")
	if len(found) != 2 {
		t.Errorf("len(FindByRite(\"rite-a\")) = %d, want 2", len(found))
	}
}

func TestInvocationState_BudgetMethods(t *testing.T) {
	state := &InvocationState{
		Budget: StateBudget{
			NativeTokens:   5000,
			BorrowedTokens: 2000,
			TotalTokens:    7000,
			BudgetLimit:    10000,
		},
	}

	// BudgetRemaining
	if remaining := state.BudgetRemaining(); remaining != 3000 {
		t.Errorf("BudgetRemaining() = %d, want 3000", remaining)
	}

	// BudgetUsagePercent
	if percent := state.BudgetUsagePercent(); percent != 70.0 {
		t.Errorf("BudgetUsagePercent() = %f, want 70.0", percent)
	}

	// IsBudgetExceeded
	if state.IsBudgetExceeded(2000) {
		t.Error("IsBudgetExceeded(2000) = true, want false")
	}
	if !state.IsBudgetExceeded(4000) {
		t.Error("IsBudgetExceeded(4000) = false, want true")
	}
}

func TestInvocationState_UpdateBudget(t *testing.T) {
	state := &InvocationState{}

	state.UpdateBudget(5000, 3000)

	if state.Budget.NativeTokens != 5000 {
		t.Errorf("NativeTokens = %d, want 5000", state.Budget.NativeTokens)
	}
	if state.Budget.BorrowedTokens != 3000 {
		t.Errorf("BorrowedTokens = %d, want 3000", state.Budget.BorrowedTokens)
	}
	if state.Budget.TotalTokens != 8000 {
		t.Errorf("TotalTokens = %d, want 8000", state.Budget.TotalTokens)
	}
}

func TestInvocationState_GetBorrowedComponents(t *testing.T) {
	state := &InvocationState{
		Invocations: []Invocation{
			{
				Skills: []string{"skill1", "skill2"},
				Agents: []InvokedAgent{{Name: "agent1"}},
			},
			{
				Skills: []string{"skill3"},
				Agents: []InvokedAgent{{Name: "agent2"}, {Name: "agent3"}},
			},
		},
	}

	skills := state.GetBorrowedSkills()
	if len(skills) != 3 {
		t.Errorf("len(GetBorrowedSkills()) = %d, want 3", len(skills))
	}

	agents := state.GetBorrowedAgents()
	if len(agents) != 3 {
		t.Errorf("len(GetBorrowedAgents()) = %d, want 3", len(agents))
	}
}

func TestInvocationState_CleanExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	state := &InvocationState{
		Invocations: []Invocation{
			{ID: "expired", ExpiresAt: &past},
			{ID: "not-expired", ExpiresAt: &future},
			{ID: "no-expiry"},
		},
	}

	removed := state.CleanExpired()
	if len(removed) != 1 {
		t.Errorf("len(removed) = %d, want 1", len(removed))
	}
	if removed[0].ID != "expired" {
		t.Errorf("Removed wrong invocation: %s", removed[0].ID)
	}
	if len(state.Invocations) != 2 {
		t.Errorf("len(Invocations) = %d, want 2", len(state.Invocations))
	}
}

func TestGenerateInvocationID(t *testing.T) {
	id1 := GenerateInvocationID()
	id2 := GenerateInvocationID()

	// Should be unique
	if id1 == id2 {
		t.Error("GenerateInvocationID() returned duplicate IDs")
	}

	// Should start with "inv-"
	if len(id1) < 4 || id1[:4] != "inv-" {
		t.Errorf("ID %q does not start with 'inv-'", id1)
	}
}
