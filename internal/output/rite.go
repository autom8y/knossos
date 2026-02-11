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
	// Silent on success per TDD (same as rite switch)
	return ""
}

// OrphanHandleResult tracks how orphans were handled.
type OrphanHandleResult struct {
	Strategy string   `json:"strategy"`
	Agents   []string `json:"agents"`
}

// InscriptionConflictOut represents a sync conflict for output.
type InscriptionConflictOut struct {
	Region    string `json:"region"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Preserved bool   `json:"preserved"`
}

// RiteStatusOutput represents detailed rite status.
type RiteStatusOutput struct {
	Rite           string        `json:"rite"`
	IsActive       bool          `json:"is_active"`
	Path           string        `json:"path"`
	Description    string        `json:"description"`
	WorkflowType   string        `json:"workflow_type"`
	Agents         []AgentStatus `json:"agents"`
	Phases         []string      `json:"phases,omitempty"`
	EntryPoint     string        `json:"entry_point"`
	Orphans        []string      `json:"orphans,omitempty"`
	ManifestValid  bool          `json:"manifest_valid"`
	ClaudeMDSynced bool          `json:"claude_md_synced"`
}

// AgentStatus represents status of an individual agent.
type AgentStatus struct {
	Name      string `json:"name"`
	File      string `json:"file"`
	Role      string `json:"role"`
	Produces  string `json:"produces"`
	Installed bool   `json:"installed"`
}

// Headers implements Tabular for RiteStatusOutput.
func (s RiteStatusOutput) Headers() []string {
	return []string{"PROPERTY", "VALUE"}
}

// Rows implements Tabular for RiteStatusOutput.
func (s RiteStatusOutput) Rows() [][]string {
	activeLabel := "No"
	if s.IsActive {
		activeLabel = "Yes"
	}

	rows := [][]string{
		{"Rite", s.Rite},
		{"Active", activeLabel},
		{"Path", s.Path},
		{"Description", s.Description},
		{"Workflow Type", s.WorkflowType},
		{"Entry Point", s.EntryPoint},
		{"Agents", fmt.Sprintf("%d", len(s.Agents))},
	}

	if len(s.Phases) > 0 {
		rows = append(rows, []string{"Phases", strings.Join(s.Phases, " -> ")})
	}

	if len(s.Orphans) > 0 {
		rows = append(rows, []string{"Orphans", strings.Join(s.Orphans, ", ")})
	}

	manifestStatus := "Valid"
	if !s.ManifestValid {
		manifestStatus = "Invalid"
	}
	rows = append(rows, []string{"Manifest", manifestStatus})

	claudeMDStatus := "Synced"
	if !s.ClaudeMDSynced {
		claudeMDStatus = "Out of sync"
	}
	rows = append(rows, []string{"CLAUDE.md", claudeMDStatus})

	return rows
}

// Text implements Textable for RiteStatusOutput.
func (s RiteStatusOutput) Text() string {
	var b strings.Builder

	activeLabel := ""
	if s.IsActive {
		activeLabel = " (ACTIVE)"
	}
	b.WriteString(fmt.Sprintf("Rite: %s%s\n", s.Rite, activeLabel))
	b.WriteString(fmt.Sprintf("Path: %s\n", s.Path))
	b.WriteString(fmt.Sprintf("Description: %s\n", s.Description))
	b.WriteString(fmt.Sprintf("Workflow: %s\n", s.WorkflowType))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("Agents (%d):\n", len(s.Agents)))
	for _, agent := range s.Agents {
		installed := "[not installed]"
		if agent.Installed {
			installed = "[installed]"
		}
		b.WriteString(fmt.Sprintf("  %-20s %-45s %s\n", agent.Name, agent.Role, installed))
	}
	b.WriteString("\n")

	if len(s.Phases) > 0 {
		b.WriteString(fmt.Sprintf("Phases: %s\n", strings.Join(s.Phases, " -> ")))
	}
	b.WriteString(fmt.Sprintf("Entry Point: %s\n", s.EntryPoint))
	b.WriteString("\n")

	// Status summary
	status := "OK"
	details := []string{}
	if s.ManifestValid {
		details = append(details, "manifest valid")
	} else {
		status = "WARNING"
		details = append(details, "manifest invalid")
	}
	if s.ClaudeMDSynced {
		details = append(details, "CLAUDE.md synced")
	} else {
		status = "WARNING"
		details = append(details, "CLAUDE.md out of sync")
	}
	b.WriteString(fmt.Sprintf("Status: %s (%s)\n", status, strings.Join(details, ", ")))

	return b.String()
}

// RiteSwitchOutput represents rite switch result.
type RiteSwitchOutput struct {
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
}

// Headers implements Tabular for RiteSwitchOutput.
func (s RiteSwitchOutput) Headers() []string {
	return []string{"PROPERTY", "VALUE"}
}

// Rows implements Tabular for RiteSwitchOutput.
func (s RiteSwitchOutput) Rows() [][]string {
	rows := [][]string{
		{"Switched To", s.Rite},
		{"Previous Rite", s.PreviousRite},
		{"Switched At", s.SwitchedAt},
		{"Agents Installed", fmt.Sprintf("%d", len(s.AgentsInstalled))},
	}

	if s.OrphansHandled != nil {
		rows = append(rows, []string{"Orphans Strategy", s.OrphansHandled.Strategy})
	}

	claudeMDStatus := "No"
	if s.ClaudeMDUpdated {
		claudeMDStatus = "Yes"
	}
	rows = append(rows, []string{"CLAUDE.md Updated", claudeMDStatus})

	return rows
}

// Text implements Textable for RiteSwitchOutput.
func (s RiteSwitchOutput) Text() string {
	// Silent on success per TDD
	return ""
}

// RiteSwitchDryRunOutput represents dry-run result.
type RiteSwitchDryRunOutput struct {
	DryRun                 bool     `json:"dry_run"`
	WouldSwitchTo          string   `json:"would_switch_to"`
	CurrentRite            string   `json:"current_rite"`
	WouldInstall           []string `json:"would_install"`
	OrphansDetected        []string `json:"orphans_detected"`
	OrphanStrategyRequired bool     `json:"orphan_strategy_required"`
	SuggestedFlags         []string `json:"suggested_flags,omitempty"`
}

// Headers implements Tabular for RiteSwitchDryRunOutput.
func (s RiteSwitchDryRunOutput) Headers() []string {
	return []string{"PROPERTY", "VALUE"}
}

// Rows implements Tabular for RiteSwitchDryRunOutput.
func (s RiteSwitchDryRunOutput) Rows() [][]string {
	rows := [][]string{
		{"Would Switch To", s.WouldSwitchTo},
		{"Current Rite", s.CurrentRite},
		{"Would Install", strings.Join(s.WouldInstall, ", ")},
	}

	if len(s.OrphansDetected) > 0 {
		rows = append(rows, []string{"Orphans Detected", strings.Join(s.OrphansDetected, ", ")})
		if s.OrphanStrategyRequired {
			rows = append(rows, []string{"Suggested Flags", strings.Join(s.SuggestedFlags, ", ")})
		}
	}

	return rows
}

// Text implements Textable for RiteSwitchDryRunOutput.
func (s RiteSwitchDryRunOutput) Text() string {
	var b strings.Builder
	b.WriteString("[DRY RUN]\n")
	b.WriteString(fmt.Sprintf("Would switch to: %s\n", s.WouldSwitchTo))
	b.WriteString(fmt.Sprintf("Current rite: %s\n", s.CurrentRite))
	b.WriteString(fmt.Sprintf("Would install: %s\n", strings.Join(s.WouldInstall, ", ")))
	if len(s.OrphansDetected) > 0 {
		b.WriteString(fmt.Sprintf("Orphans detected: %s\n", strings.Join(s.OrphansDetected, ", ")))
		if s.OrphanStrategyRequired {
			b.WriteString(fmt.Sprintf("Suggested flags: %s\n", strings.Join(s.SuggestedFlags, ", ")))
		}
	}
	return b.String()
}

// PantheonOutput represents the agent pantheon for a rite.
type PantheonOutput struct {
	Rite   string          `json:"rite"`
	Agents []PantheonAgent `json:"agents"`
	Count  int             `json:"count"`
}

// PantheonAgent represents a single agent in the pantheon.
type PantheonAgent struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description,omitempty"`
	Model       string `json:"model,omitempty"`
}

// Headers implements Tabular for PantheonOutput.
func (p PantheonOutput) Headers() []string {
	return []string{"NAME", "MODEL", "ROLE"}
}

// Rows implements Tabular for PantheonOutput.
func (p PantheonOutput) Rows() [][]string {
	rows := make([][]string, len(p.Agents))
	for i, a := range p.Agents {
		model := a.Model
		if model == "" {
			model = "-"
		}
		desc := truncateDescription(a.Description)
		if desc == "" {
			desc = "-"
		}
		rows[i] = []string{a.Name, model, desc}
	}
	return rows
}

// truncateDescription extracts the first sentence from a description.
func truncateDescription(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Join first two lines for descriptions that wrap mid-sentence
	lines := strings.SplitN(s, "\n", 3)
	joined := lines[0]
	if len(lines) > 1 && !strings.HasSuffix(strings.TrimSpace(lines[0]), ".") {
		joined = strings.TrimSpace(lines[0]) + " " + strings.TrimSpace(lines[1])
	}
	// Truncate to first sentence
	if idx := strings.Index(joined, ". "); idx != -1 {
		return joined[:idx+1]
	}
	if strings.HasSuffix(joined, ".") {
		return joined
	}
	// Cap at 80 chars
	if len(joined) > 80 {
		return joined[:77] + "..."
	}
	return joined
}

// Text implements Textable for PantheonOutput.
func (p PantheonOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Rite: %s (%d agents)\n", p.Rite, p.Count))
	return b.String()
}

// RiteValidateOutput represents validation result.
type RiteValidateOutput struct {
	Rite     string               `json:"rite"`
	Valid    bool                 `json:"valid"`
	Checks   []ValidationCheckOut `json:"checks"`
	Errors   int                  `json:"errors"`
	Warnings int                  `json:"warnings"`
	Fixable  []string             `json:"fixable,omitempty"`
}

// ValidationCheckOut represents a validation check result.
type ValidationCheckOut struct {
	Check   string `json:"check"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Headers implements Tabular for RiteValidateOutput.
func (v RiteValidateOutput) Headers() []string {
	return []string{"CHECK", "STATUS", "MESSAGE"}
}

// Rows implements Tabular for RiteValidateOutput.
func (v RiteValidateOutput) Rows() [][]string {
	rows := make([][]string, len(v.Checks))
	for i, check := range v.Checks {
		status := "PASS"
		switch check.Status {
		case "fail":
			status = "FAIL"
		case "warn":
			status = "WARN"
		}
		rows[i] = []string{check.Check, status, check.Message}
	}
	return rows
}

// Text implements Textable for RiteValidateOutput.
func (v RiteValidateOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Validating rite: %s\n\n", v.Rite))

	for _, check := range v.Checks {
		status := "[PASS]"
		switch check.Status {
		case "fail":
			status = "[FAIL]"
		case "warn":
			status = "[WARN]"
		}
		b.WriteString(fmt.Sprintf("%s %-18s %s\n", status, check.Check, check.Message))
	}

	b.WriteString("\n")
	result := "VALID"
	if !v.Valid {
		result = "INVALID"
	}
	b.WriteString(fmt.Sprintf("Result: %s (%d errors, %d warnings)\n", result, v.Errors, v.Warnings))

	return b.String()
}
