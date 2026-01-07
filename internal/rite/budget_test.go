package rite

import (
	"testing"
)

func TestBudgetCalculator_Defaults(t *testing.T) {
	calc := NewBudgetCalculator()

	if calc.DefaultAgentCost != 2000 {
		t.Errorf("DefaultAgentCost = %d, want 2000", calc.DefaultAgentCost)
	}
	if calc.DefaultSkillCost != 1000 {
		t.Errorf("DefaultSkillCost = %d, want 1000", calc.DefaultSkillCost)
	}
	if calc.DefaultWorkflowCost != 500 {
		t.Errorf("DefaultWorkflowCost = %d, want 500", calc.DefaultWorkflowCost)
	}
}

func TestBudgetCalculator_CalculateInvocationCost(t *testing.T) {
	calc := NewBudgetCalculator()

	tests := []struct {
		name     string
		borrowed *BorrowedComponents
		want     int
	}{
		{
			name: "empty",
			borrowed: &BorrowedComponents{},
			want: 0,
		},
		{
			name: "skills only",
			borrowed: &BorrowedComponents{
				Skills: []string{"skill1", "skill2"},
			},
			want: 2000, // 2 * 1000
		},
		{
			name: "agents only",
			borrowed: &BorrowedComponents{
				Agents: []InvokedAgent{
					{Name: "agent1"},
					{Name: "agent2"},
				},
			},
			want: 4000, // 2 * 2000
		},
		{
			name: "mixed",
			borrowed: &BorrowedComponents{
				Skills: []string{"skill1"},
				Agents: []InvokedAgent{{Name: "agent1"}},
			},
			want: 3000, // 1000 + 2000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.CalculateInvocationCost(tt.borrowed)
			if got != tt.want {
				t.Errorf("CalculateInvocationCost() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBudgetCalculator_CalculateRiteCost(t *testing.T) {
	calc := NewBudgetCalculator()

	tests := []struct {
		name     string
		manifest *RiteManifest
		want     int
	}{
		{
			name: "with explicit estimate",
			manifest: &RiteManifest{
				Budget: &BudgetInfo{
					EstimatedTokens: 10000,
				},
			},
			want: 10000,
		},
		{
			name: "with component costs",
			manifest: &RiteManifest{
				Budget: &BudgetInfo{
					AgentsCost:   5000,
					SkillsCost:   3000,
					WorkflowCost: 500,
				},
				Workflow: &WorkflowConfig{Type: "sequential"},
			},
			want: 8500,
		},
		{
			name: "from defaults",
			manifest: &RiteManifest{
				Agents: []AgentRef{{Name: "a1"}, {Name: "a2"}},
				Skills: []SkillRef{{Ref: "s1"}},
				Workflow: &WorkflowConfig{Type: "sequential"},
			},
			want: 5500, // 2*2000 + 1000 + 500
		},
		{
			name: "empty manifest",
			manifest: &RiteManifest{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.CalculateRiteCost(tt.manifest)
			if got != tt.want {
				t.Errorf("CalculateRiteCost() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBudgetCalculator_CheckBudgetWarnings(t *testing.T) {
	calc := NewBudgetCalculator()

	tests := []struct {
		name          string
		state         *InvocationState
		wantWarnings  int
		wantLevel     string
	}{
		{
			name: "under threshold",
			state: &InvocationState{
				Budget: StateBudget{
					TotalTokens: 5000,
					BudgetLimit: 10000,
				},
			},
			wantWarnings: 0,
		},
		{
			name: "at warn threshold",
			state: &InvocationState{
				Budget: StateBudget{
					TotalTokens: 7500,
					BudgetLimit: 10000,
				},
			},
			wantWarnings: 1,
			wantLevel:    "warn",
		},
		{
			name: "at critical threshold",
			state: &InvocationState{
				Budget: StateBudget{
					TotalTokens: 9500,
					BudgetLimit: 10000,
				},
			},
			wantWarnings: 1,
			wantLevel:    "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := calc.CheckBudgetWarnings(tt.state)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("CheckBudgetWarnings() returned %d warnings, want %d",
					len(warnings), tt.wantWarnings)
			}
			if len(warnings) > 0 && warnings[0].Level != tt.wantLevel {
				t.Errorf("Warning level = %q, want %q", warnings[0].Level, tt.wantLevel)
			}
		})
	}
}

func TestBudgetCalculator_CalculateDetailedCost(t *testing.T) {
	calc := NewBudgetCalculator()

	manifest := &RiteManifest{
		Agents: []AgentRef{
			{Name: "agent1"},
			{Name: "agent2"},
		},
		Skills: []SkillRef{
			{Ref: "skill1"},
		},
		Workflow: &WorkflowConfig{Type: "sequential"},
	}

	costs := calc.CalculateDetailedCost(manifest)

	// Should have 2 agents + 1 skill + 1 workflow = 4 entries
	if len(costs) != 4 {
		t.Errorf("len(costs) = %d, want 4", len(costs))
	}

	// Check types
	agentCount := 0
	skillCount := 0
	workflowCount := 0
	for _, cost := range costs {
		switch cost.Type {
		case "agent":
			agentCount++
		case "skill":
			skillCount++
		case "workflow":
			workflowCount++
		}
	}
	if agentCount != 2 {
		t.Errorf("Agent costs = %d, want 2", agentCount)
	}
	if skillCount != 1 {
		t.Errorf("Skill costs = %d, want 1", skillCount)
	}
	if workflowCount != 1 {
		t.Errorf("Workflow costs = %d, want 1", workflowCount)
	}
}

func TestBudgetCalculator_CalculateSummaryCost(t *testing.T) {
	calc := NewBudgetCalculator()

	// With explicit budget
	m1 := &RiteManifest{
		Budget: &BudgetInfo{
			AgentsCost:   5000,
			SkillsCost:   3000,
			WorkflowCost: 1000,
		},
		Workflow: &WorkflowConfig{Type: "sequential"},
	}
	summary1 := calc.CalculateSummaryCost(m1)
	if summary1.TotalCost != 9000 {
		t.Errorf("TotalCost = %d, want 9000", summary1.TotalCost)
	}

	// Without explicit budget (uses defaults)
	m2 := &RiteManifest{
		Agents:   []AgentRef{{Name: "a"}},
		Skills:   []SkillRef{{Ref: "s"}},
		Workflow: &WorkflowConfig{Type: "sequential"},
	}
	summary2 := calc.CalculateSummaryCost(m2)
	if summary2.AgentsCost != 2000 {
		t.Errorf("AgentsCost = %d, want 2000", summary2.AgentsCost)
	}
	if summary2.SkillsCost != 1000 {
		t.Errorf("SkillsCost = %d, want 1000", summary2.SkillsCost)
	}
}
