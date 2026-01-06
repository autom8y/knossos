// Package output provides format-aware output printing for Ariadne.
// This file contains rite-domain specific output structures.
package output

import (
	"fmt"
	"strings"
)

// --- Rite Output Structures ---

// RiteListOutput represents the rite list for JSON output.
type RiteListOutput struct {
	Rites      []RiteSummary `json:"rites"`
	Total      int           `json:"total"`
	ActiveRite string        `json:"active_rite,omitempty"`
}

// RiteSummary is a brief rite entry for listing.
type RiteSummary struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Form        string   `json:"form"`
	Agents      []string `json:"agents,omitempty"`
	AgentCount  int      `json:"agent_count"`
	Skills      []string `json:"skills,omitempty"`
	SkillCount  int      `json:"skill_count"`
	Path        string   `json:"path"`
	Source      string   `json:"source"` // "project" or "user"
	Active      bool     `json:"active"`
}

// Headers implements Tabular for RiteListOutput.
func (l RiteListOutput) Headers() []string {
	return []string{"RITE", "FORM", "AGENTS", "SKILLS", "SOURCE", "ACTIVE"}
}

// Rows implements Tabular for RiteListOutput.
func (l RiteListOutput) Rows() [][]string {
	rows := make([][]string, len(l.Rites))
	for i, r := range l.Rites {
		active := ""
		if r.Active {
			active = "*"
		}
		rows[i] = []string{
			r.Name,
			r.Form,
			fmt.Sprintf("%d", r.AgentCount),
			fmt.Sprintf("%d", r.SkillCount),
			r.Source,
			active,
		}
	}
	return rows
}

// Text implements Textable for RiteListOutput.
func (l RiteListOutput) Text() string {
	if len(l.Rites) == 0 {
		return "No rites found"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Total: %d rites (* = active)\n", l.Total))
	return b.String()
}

// RiteInfoOutput represents detailed rite information.
type RiteInfoOutput struct {
	Name            string               `json:"name"`
	DisplayName     string               `json:"display_name,omitempty"`
	Description     string               `json:"description,omitempty"`
	Form            string               `json:"form"`
	Path            string               `json:"path"`
	Source          string               `json:"source"`
	Active          bool                 `json:"active"`
	Agents          []RiteAgentInfo      `json:"agents,omitempty"`
	Skills          []RiteSkillInfo      `json:"skills,omitempty"`
	Workflow        *RiteWorkflowInfo    `json:"workflow,omitempty"`
	Budget          *RiteBudgetInfo      `json:"budget,omitempty"`
	SchemaVersion   string               `json:"schema_version"`
}

// RiteAgentInfo represents agent information.
type RiteAgentInfo struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Role     string `json:"role,omitempty"`
	Produces string `json:"produces,omitempty"`
}

// RiteSkillInfo represents skill information.
type RiteSkillInfo struct {
	Ref      string `json:"ref"`
	Path     string `json:"path,omitempty"`
	External bool   `json:"external"`
}

// RiteWorkflowInfo represents workflow information.
type RiteWorkflowInfo struct {
	Type       string   `json:"type"`
	EntryPoint string   `json:"entry_point"`
	Phases     []string `json:"phases,omitempty"`
}

// RiteBudgetInfo represents budget information.
type RiteBudgetInfo struct {
	EstimatedTokens int `json:"estimated_tokens"`
	AgentsCost      int `json:"agents_cost"`
	SkillsCost      int `json:"skills_cost"`
	WorkflowCost    int `json:"workflow_cost"`
}

// Text implements Textable for RiteInfoOutput.
func (r RiteInfoOutput) Text() string {
	var b strings.Builder

	activeLabel := ""
	if r.Active {
		activeLabel = " (ACTIVE)"
	}
	b.WriteString(fmt.Sprintf("Rite: %s%s\n", r.Name, activeLabel))
	if r.DisplayName != "" && r.DisplayName != r.Name {
		b.WriteString(fmt.Sprintf("Display Name: %s\n", r.DisplayName))
	}
	b.WriteString(fmt.Sprintf("Form: %s\n", r.Form))
	b.WriteString(fmt.Sprintf("Path: %s\n", r.Path))
	b.WriteString(fmt.Sprintf("Source: %s\n", r.Source))
	if r.Description != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", r.Description))
	}
	b.WriteString("\n")

	if len(r.Agents) > 0 {
		b.WriteString(fmt.Sprintf("Agents (%d):\n", len(r.Agents)))
		for _, agent := range r.Agents {
			role := agent.Role
			if role == "" {
				role = "(no role specified)"
			}
			b.WriteString(fmt.Sprintf("  %-20s %s\n", agent.Name, role))
		}
		b.WriteString("\n")
	}

	if len(r.Skills) > 0 {
		b.WriteString(fmt.Sprintf("Skills (%d):\n", len(r.Skills)))
		for _, skill := range r.Skills {
			source := "local"
			if skill.External {
				source = "external"
			}
			b.WriteString(fmt.Sprintf("  %-20s (%s)\n", skill.Ref, source))
		}
		b.WriteString("\n")
	}

	if r.Workflow != nil {
		b.WriteString(fmt.Sprintf("Workflow: %s\n", r.Workflow.Type))
		b.WriteString(fmt.Sprintf("Entry Point: %s\n", r.Workflow.EntryPoint))
		if len(r.Workflow.Phases) > 0 {
			b.WriteString(fmt.Sprintf("Phases: %s\n", strings.Join(r.Workflow.Phases, " -> ")))
		}
		b.WriteString("\n")
	}

	if r.Budget != nil {
		b.WriteString("Budget:\n")
		b.WriteString(fmt.Sprintf("  Estimated Tokens: %d\n", r.Budget.EstimatedTokens))
		b.WriteString(fmt.Sprintf("  Agents Cost: %d\n", r.Budget.AgentsCost))
		b.WriteString(fmt.Sprintf("  Skills Cost: %d\n", r.Budget.SkillsCost))
		if r.Budget.WorkflowCost > 0 {
			b.WriteString(fmt.Sprintf("  Workflow Cost: %d\n", r.Budget.WorkflowCost))
		}
	}

	return b.String()
}

// RiteCurrentOutput represents the current rite and invocations.
type RiteCurrentOutput struct {
	ActiveRite       string                  `json:"active_rite"`
	NativeAgents     []string                `json:"native_agents,omitempty"`
	NativeSkills     []string                `json:"native_skills,omitempty"`
	Invocations      []InvocationOutput      `json:"invocations,omitempty"`
	BorrowedAgents   []string                `json:"borrowed_agents,omitempty"`
	BorrowedSkills   []string                `json:"borrowed_skills,omitempty"`
	Budget           CurrentBudgetOutput     `json:"budget"`
}

// InvocationOutput represents an active invocation.
type InvocationOutput struct {
	ID        string   `json:"id"`
	RiteName  string   `json:"rite_name"`
	Component string   `json:"component,omitempty"`
	Skills    []string `json:"skills,omitempty"`
	Agents    []string `json:"agents,omitempty"`
	InvokedAt string   `json:"invoked_at"`
}

// CurrentBudgetOutput represents budget status.
type CurrentBudgetOutput struct {
	NativeTokens   int     `json:"native_tokens"`
	BorrowedTokens int     `json:"borrowed_tokens"`
	TotalTokens    int     `json:"total_tokens"`
	BudgetLimit    int     `json:"budget_limit"`
	UsagePercent   float64 `json:"usage_percent"`
}

// Text implements Textable for RiteCurrentOutput.
func (c RiteCurrentOutput) Text() string {
	var b strings.Builder

	if c.ActiveRite == "" {
		b.WriteString("No active rite\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Active Rite: %s\n\n", c.ActiveRite))

	b.WriteString("Native Components:\n")
	if len(c.NativeAgents) > 0 {
		b.WriteString(fmt.Sprintf("  Agents: %s\n", strings.Join(c.NativeAgents, ", ")))
	}
	if len(c.NativeSkills) > 0 {
		b.WriteString(fmt.Sprintf("  Skills: %s\n", strings.Join(c.NativeSkills, ", ")))
	}
	b.WriteString("\n")

	if len(c.Invocations) > 0 {
		b.WriteString("Borrowed Components:\n")
		for _, inv := range c.Invocations {
			b.WriteString(fmt.Sprintf("  From %s (%s):\n", inv.RiteName, inv.ID))
			if len(inv.Skills) > 0 {
				b.WriteString(fmt.Sprintf("    Skills: %s\n", strings.Join(inv.Skills, ", ")))
			}
			if len(inv.Agents) > 0 {
				b.WriteString(fmt.Sprintf("    Agents: %s\n", strings.Join(inv.Agents, ", ")))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("Total Context Budget: ~%d tokens\n", c.Budget.TotalTokens))
	b.WriteString(fmt.Sprintf("  Native: %d tokens\n", c.Budget.NativeTokens))
	b.WriteString(fmt.Sprintf("  Borrowed: %d tokens\n", c.Budget.BorrowedTokens))
	b.WriteString(fmt.Sprintf("  Limit: %d tokens (%.1f%% used)\n", c.Budget.BudgetLimit, c.Budget.UsagePercent))

	return b.String()
}

// RiteInvokeOutput represents invoke result.
type RiteInvokeOutput struct {
	InvokedRite        string   `json:"invoked_rite"`
	Component          string   `json:"component,omitempty"`
	InvocationID       string   `json:"invocation_id,omitempty"`
	BorrowedSkills     []string `json:"borrowed_skills,omitempty"`
	BorrowedAgents     []string `json:"borrowed_agents,omitempty"`
	InscriptionUpdated bool     `json:"inscription_updated"`
	EstimatedTokens    int      `json:"estimated_tokens"`
	DryRun             bool     `json:"dry_run,omitempty"`
}

// Text implements Textable for RiteInvokeOutput.
func (r RiteInvokeOutput) Text() string {
	if r.DryRun {
		var b strings.Builder
		b.WriteString("[DRY RUN]\n")
		b.WriteString(fmt.Sprintf("Would invoke: %s\n", r.InvokedRite))
		if r.Component != "" {
			b.WriteString(fmt.Sprintf("Component: %s\n", r.Component))
		}
		if len(r.BorrowedSkills) > 0 {
			b.WriteString(fmt.Sprintf("Would borrow skills: %s\n", strings.Join(r.BorrowedSkills, ", ")))
		}
		if len(r.BorrowedAgents) > 0 {
			b.WriteString(fmt.Sprintf("Would borrow agents: %s\n", strings.Join(r.BorrowedAgents, ", ")))
		}
		b.WriteString(fmt.Sprintf("Estimated tokens: %d\n", r.EstimatedTokens))
		return b.String()
	}
	// Silent on success for mutations
	return ""
}

// RiteReleaseOutput represents release result.
type RiteReleaseOutput struct {
	ReleasedRites      []string `json:"released_rites"`
	ReleasedSkills     []string `json:"released_skills,omitempty"`
	ReleasedAgents     []string `json:"released_agents,omitempty"`
	InvocationCount    int      `json:"invocation_count"`
	TokensFreed        int      `json:"tokens_freed"`
	InscriptionUpdated bool     `json:"inscription_updated"`
	DryRun             bool     `json:"dry_run,omitempty"`
}

// Text implements Textable for RiteReleaseOutput.
func (r RiteReleaseOutput) Text() string {
	if r.DryRun {
		var b strings.Builder
		b.WriteString("[DRY RUN]\n")
		b.WriteString(fmt.Sprintf("Would release: %s\n", strings.Join(r.ReleasedRites, ", ")))
		b.WriteString(fmt.Sprintf("Invocations: %d\n", r.InvocationCount))
		b.WriteString(fmt.Sprintf("Tokens freed: %d\n", r.TokensFreed))
		return b.String()
	}
	// Silent on success for mutations
	return ""
}

// RiteSwapOutput represents swap result.
type RiteSwapOutput struct {
	Rite               string                   `json:"rite"`
	PreviousRite       string                   `json:"previous_rite"`
	SwitchedAt         string                   `json:"switched_at"`
	AgentsInstalled    []string                 `json:"agents_installed"`
	OrphansHandled     *OrphanHandleResult      `json:"orphans_handled,omitempty"`
	ClaudeMDUpdated    bool                     `json:"claude_md_updated"`
	ManifestPath       string                   `json:"manifest_path"`
	InscriptionSynced  bool                     `json:"inscription_synced,omitempty"`
	InscriptionVersion string                   `json:"inscription_version,omitempty"`
	SyncConflicts      []InscriptionConflictOut `json:"sync_conflicts,omitempty"`
	// Invocations released during swap
	InvocationsReleased int `json:"invocations_released,omitempty"`
}

// Text implements Textable for RiteSwapOutput.
func (s RiteSwapOutput) Text() string {
	// Silent on success per TDD (same as team switch)
	return ""
}
