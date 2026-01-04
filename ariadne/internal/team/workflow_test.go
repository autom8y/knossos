package team

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkflow(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "teams", "valid-team", "workflow.yaml")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	workflow, err := LoadWorkflow(absPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	if workflow.Name != "valid-team" {
		t.Errorf("Name = %q, want %q", workflow.Name, "valid-team")
	}

	if workflow.WorkflowType != "sequential" {
		t.Errorf("WorkflowType = %q, want %q", workflow.WorkflowType, "sequential")
	}

	if workflow.EntryPoint.Agent != "agent-a" {
		t.Errorf("EntryPoint.Agent = %q, want %q", workflow.EntryPoint.Agent, "agent-a")
	}

	if len(workflow.Phases) != 2 {
		t.Errorf("len(Phases) = %d, want 2", len(workflow.Phases))
	}
}

func TestWorkflow_PhaseNames(t *testing.T) {
	workflow := &Workflow{
		Phases: []Phase{
			{Name: "phase1"},
			{Name: "phase2"},
			{Name: "phase3"},
		},
	}

	names := workflow.PhaseNames()

	if len(names) != 3 {
		t.Fatalf("len(PhaseNames()) = %d, want 3", len(names))
	}

	expected := []string{"phase1", "phase2", "phase3"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("PhaseNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestWorkflow_AgentNames(t *testing.T) {
	workflow := &Workflow{
		Phases: []Phase{
			{Name: "phase1", Agent: "agent-a"},
			{Name: "phase2", Agent: "agent-b"},
			{Name: "phase3", Agent: "agent-a"}, // duplicate
		},
	}

	agents := workflow.AgentNames()

	// Should deduplicate
	if len(agents) != 2 {
		t.Errorf("len(AgentNames()) = %d, want 2 (deduplicated)", len(agents))
	}

	agentSet := make(map[string]bool)
	for _, a := range agents {
		agentSet[a] = true
	}

	if !agentSet["agent-a"] {
		t.Error("AgentNames() missing agent-a")
	}
	if !agentSet["agent-b"] {
		t.Error("AgentNames() missing agent-b")
	}
}

func TestWorkflow_GetPhase(t *testing.T) {
	workflow := &Workflow{
		Phases: []Phase{
			{Name: "phase1", Agent: "agent-a"},
			{Name: "phase2", Agent: "agent-b"},
		},
	}

	phase := workflow.GetPhase("phase1")
	if phase == nil {
		t.Fatal("GetPhase(phase1) returned nil")
	}
	if phase.Agent != "agent-a" {
		t.Errorf("GetPhase(phase1).Agent = %q, want %q", phase.Agent, "agent-a")
	}

	phase = workflow.GetPhase("nonexistent")
	if phase != nil {
		t.Error("GetPhase(nonexistent) should return nil")
	}
}

func TestWorkflow_GetAgentInfo(t *testing.T) {
	workflow := &Workflow{
		Phases: []Phase{
			{Name: "requirements", Agent: "requirements-analyst", Produces: "prd"},
			{Name: "design", Agent: "architect", Produces: "tdd"},
		},
	}

	infos := workflow.GetAgentInfo()

	if len(infos) != 2 {
		t.Fatalf("len(GetAgentInfo()) = %d, want 2", len(infos))
	}

	// Check first agent
	if infos[0].Name != "requirements-analyst" {
		t.Errorf("infos[0].Name = %q, want %q", infos[0].Name, "requirements-analyst")
	}
	if infos[0].File != "requirements-analyst.md" {
		t.Errorf("infos[0].File = %q, want %q", infos[0].File, "requirements-analyst.md")
	}
	if infos[0].Produces != "prd" {
		t.Errorf("infos[0].Produces = %q, want %q", infos[0].Produces, "prd")
	}

	// Check derived role
	if infos[1].Role != "Evaluates tradeoffs and designs systems" {
		t.Errorf("infos[1].Role = %q (derived from architect)", infos[1].Role)
	}
}

func TestDeriveRoleFromName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"architect", "Evaluates tradeoffs and designs systems"},
		{"principal-engineer", "Transforms designs into production code"},
		{"qa-adversary", "Breaks things so users don't"},
		{"custom-agent", "Custom Agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveRoleFromName(tt.name)
			if got != tt.want {
				t.Errorf("deriveRoleFromName(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
