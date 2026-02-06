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
