package rite

import (
	"os"
	"path/filepath"
)

// BudgetCalculator handles token estimation for rite components.
type BudgetCalculator struct {
	// Default costs per component type (tokens)
	DefaultAgentCost   int
	DefaultSkillCost   int
	DefaultWorkflowCost int
	// Bytes per token ratio
	BytesPerToken int
}

// NewBudgetCalculator creates a new budget calculator with defaults.
func NewBudgetCalculator() *BudgetCalculator {
	return &BudgetCalculator{
		DefaultAgentCost:    2000, // Typical agent prompt: 1500-3000 tokens
		DefaultSkillCost:    1000, // Typical skill: 500-2000 tokens
		DefaultWorkflowCost: 500,  // Workflow config: 200-500 tokens
		BytesPerToken:       4,    // Approximate bytes per token
	}
}

// CalculateInvocationCost estimates the token cost of borrowed components.
func (c *BudgetCalculator) CalculateInvocationCost(borrowed *BorrowedComponents) int {
	total := 0

	// Estimate skill costs
	total += len(borrowed.Skills) * c.DefaultSkillCost

	// Estimate agent costs
	total += len(borrowed.Agents) * c.DefaultAgentCost

	return total
}

// CalculateRiteCost estimates the total token cost of a rite.
func (c *BudgetCalculator) CalculateRiteCost(manifest *RiteManifest) int {
	// If budget is specified in manifest, use it
	if manifest.Budget != nil && manifest.Budget.EstimatedTokens > 0 {
		return manifest.Budget.EstimatedTokens
	}

	total := 0

	// Calculate from components
	if manifest.Budget != nil {
		if manifest.Budget.AgentsCost > 0 {
			total += manifest.Budget.AgentsCost
		} else {
			total += len(manifest.Agents) * c.DefaultAgentCost
		}

		if manifest.Budget.SkillsCost > 0 {
			total += manifest.Budget.SkillsCost
		} else {
			total += len(manifest.Skills) * c.DefaultSkillCost
		}

		if manifest.Budget.WorkflowCost > 0 {
			total += manifest.Budget.WorkflowCost
		} else if manifest.HasWorkflow() {
			total += c.DefaultWorkflowCost
		}
	} else {
		// No budget info, use defaults
		total += len(manifest.Agents) * c.DefaultAgentCost
		total += len(manifest.Skills) * c.DefaultSkillCost
		if manifest.HasWorkflow() {
			total += c.DefaultWorkflowCost
		}
	}

	return total
}

// CalculateRiteCostFromDir calculates cost by scanning actual files.
func (c *BudgetCalculator) CalculateRiteCostFromDir(ritePath string) (int, error) {
	total := 0

	// Scan agents directory
	agentsDir := filepath.Join(ritePath, "agents")
	if stat, err := os.Stat(agentsDir); err == nil && stat.IsDir() {
		agentCost, err := c.calculateDirCost(agentsDir, ".md")
		if err == nil {
			total += agentCost
		}
	}

	// Scan skills directory
	skillsDir := filepath.Join(ritePath, "skills")
	if stat, err := os.Stat(skillsDir); err == nil && stat.IsDir() {
		skillCost, err := c.calculateDirCost(skillsDir, "")
		if err == nil {
			total += skillCost
		}
	}

	// Check for workflow
	workflowPath := filepath.Join(ritePath, "workflow.yaml")
	if stat, err := os.Stat(workflowPath); err == nil {
		total += int(stat.Size()) / c.BytesPerToken
	}

	return total, nil
}

// calculateDirCost calculates token cost for files in a directory.
func (c *BudgetCalculator) calculateDirCost(dir, ext string) (int, error) {
	total := 0

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively process subdirectories (for skills)
			subCost, err := c.calculateDirCost(path, ext)
			if err == nil {
				total += subCost
			}
		} else if ext == "" || filepath.Ext(entry.Name()) == ext {
			// Calculate file cost
			stat, err := os.Stat(path)
			if err == nil {
				total += int(stat.Size()) / c.BytesPerToken
			}
		}
	}

	return total, nil
}

// EstimateFileCost estimates the token cost of a single file.
func (c *BudgetCalculator) EstimateFileCost(path string) (int, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(stat.Size()) / c.BytesPerToken, nil
}

// BudgetWarning represents a budget-related warning.
type BudgetWarning struct {
	Level   string `json:"level"`   // "warn", "critical"
	Message string `json:"message"`
	Usage   int    `json:"usage"`
	Limit   int    `json:"limit"`
	Percent float64 `json:"percent"`
}

// Budget warning thresholds
const (
	BudgetWarnPercent     = 75.0  // Warn at 75% usage
	BudgetCriticalPercent = 90.0  // Critical at 90% usage
)

// CheckBudgetWarnings returns any budget warnings.
func (c *BudgetCalculator) CheckBudgetWarnings(state *InvocationState) []BudgetWarning {
	var warnings []BudgetWarning

	percent := state.BudgetUsagePercent()

	if percent >= BudgetCriticalPercent {
		warnings = append(warnings, BudgetWarning{
			Level:   "critical",
			Message: "Context budget critically high. Consider releasing unused invocations.",
			Usage:   state.Budget.TotalTokens,
			Limit:   state.Budget.BudgetLimit,
			Percent: percent,
		})
	} else if percent >= BudgetWarnPercent {
		warnings = append(warnings, BudgetWarning{
			Level:   "warn",
			Message: "Context budget approaching limit. Consider releasing unused invocations.",
			Usage:   state.Budget.TotalTokens,
			Limit:   state.Budget.BudgetLimit,
			Percent: percent,
		})
	}

	return warnings
}

// ComponentCost represents the cost breakdown of a component.
type ComponentCost struct {
	Type      string `json:"type"`      // "agent", "skill", "workflow"
	Name      string `json:"name"`
	Tokens    int    `json:"tokens"`
	Estimated bool   `json:"estimated"` // True if using default estimate
}

// CalculateDetailedCost provides a detailed breakdown of rite costs.
func (c *BudgetCalculator) CalculateDetailedCost(manifest *RiteManifest) []ComponentCost {
	var costs []ComponentCost

	// Agent costs
	for _, agent := range manifest.Agents {
		cost := c.DefaultAgentCost
		estimated := true

		// Try to get actual cost if manifest has budget info
		if manifest.Budget != nil && manifest.Budget.AgentsCost > 0 && len(manifest.Agents) > 0 {
			cost = manifest.Budget.AgentsCost / len(manifest.Agents)
			estimated = false
		}

		costs = append(costs, ComponentCost{
			Type:      "agent",
			Name:      agent.Name,
			Tokens:    cost,
			Estimated: estimated,
		})
	}

	// Skill costs
	for _, skill := range manifest.Skills {
		cost := c.DefaultSkillCost
		estimated := true

		if manifest.Budget != nil && manifest.Budget.SkillsCost > 0 && len(manifest.Skills) > 0 {
			cost = manifest.Budget.SkillsCost / len(manifest.Skills)
			estimated = false
		}

		costs = append(costs, ComponentCost{
			Type:      "skill",
			Name:      skill.Ref,
			Tokens:    cost,
			Estimated: estimated,
		})
	}

	// Workflow cost
	if manifest.HasWorkflow() {
		cost := c.DefaultWorkflowCost
		estimated := true

		if manifest.Budget != nil && manifest.Budget.WorkflowCost > 0 {
			cost = manifest.Budget.WorkflowCost
			estimated = false
		}

		costs = append(costs, ComponentCost{
			Type:      "workflow",
			Name:      "workflow",
			Tokens:    cost,
			Estimated: estimated,
		})
	}

	return costs
}

// SummaryCost provides a summary of rite costs by category.
type SummaryCost struct {
	AgentsCost   int `json:"agents_cost"`
	SkillsCost   int `json:"skills_cost"`
	WorkflowCost int `json:"workflow_cost"`
	TotalCost    int `json:"total_cost"`
}

// CalculateSummaryCost provides a summary cost breakdown.
func (c *BudgetCalculator) CalculateSummaryCost(manifest *RiteManifest) SummaryCost {
	summary := SummaryCost{}

	if manifest.Budget != nil {
		summary.AgentsCost = manifest.Budget.AgentsCost
		summary.SkillsCost = manifest.Budget.SkillsCost
		summary.WorkflowCost = manifest.Budget.WorkflowCost
	}

	// Fall back to estimates if not specified
	if summary.AgentsCost == 0 {
		summary.AgentsCost = len(manifest.Agents) * c.DefaultAgentCost
	}
	if summary.SkillsCost == 0 {
		summary.SkillsCost = len(manifest.Skills) * c.DefaultSkillCost
	}
	if summary.WorkflowCost == 0 && manifest.HasWorkflow() {
		summary.WorkflowCost = c.DefaultWorkflowCost
	}

	summary.TotalCost = summary.AgentsCost + summary.SkillsCost + summary.WorkflowCost

	return summary
}
