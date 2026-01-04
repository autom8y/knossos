// Package output provides format-aware output printing for Ariadne.
// This file contains team-domain specific output structures.
package output

import (
	"fmt"
	"strings"
)

// --- Team Output Structures ---

// TeamListOutput represents the team list for JSON output.
type TeamListOutput struct {
	Teams      []TeamSummary `json:"teams"`
	Total      int           `json:"total"`
	ActiveTeam string        `json:"active_team,omitempty"`
}

// TeamSummary is a brief team entry for listing.
type TeamSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Agents      []string `json:"agents"`
	AgentCount  int      `json:"agent_count"`
	Path        string   `json:"path"`
	Active      bool     `json:"active"`
}

// Headers implements Tabular for TeamListOutput.
func (l TeamListOutput) Headers() []string {
	return []string{"TEAM", "AGENTS", "DESCRIPTION", "ACTIVE"}
}

// Rows implements Tabular for TeamListOutput.
func (l TeamListOutput) Rows() [][]string {
	rows := make([][]string, len(l.Teams))
	for i, t := range l.Teams {
		active := ""
		if t.Active {
			active = "*"
		}
		desc := t.Description
		if len(desc) > 45 {
			desc = desc[:42] + "..."
		}
		rows[i] = []string{
			t.Name,
			fmt.Sprintf("%d", t.AgentCount),
			desc,
			active,
		}
	}
	return rows
}

// Text implements Textable for TeamListOutput.
func (l TeamListOutput) Text() string {
	if len(l.Teams) == 0 {
		return "No teams found"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Total: %d teams (* = active)\n", l.Total))
	return b.String()
}

// TeamStatusOutput represents detailed team status.
type TeamStatusOutput struct {
	Team           string        `json:"team"`
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

// Text implements Textable for TeamStatusOutput.
func (s TeamStatusOutput) Text() string {
	var b strings.Builder

	activeLabel := ""
	if s.IsActive {
		activeLabel = " (ACTIVE)"
	}
	b.WriteString(fmt.Sprintf("Team: %s%s\n", s.Team, activeLabel))
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

// TeamSwitchOutput represents team switch result.
type TeamSwitchOutput struct {
	Team            string              `json:"team"`
	PreviousTeam    string              `json:"previous_team"`
	SwitchedAt      string              `json:"switched_at"`
	AgentsInstalled []string            `json:"agents_installed"`
	OrphansHandled  *OrphanHandleResult `json:"orphans_handled,omitempty"`
	ClaudeMDUpdated bool                `json:"claude_md_updated"`
	ManifestPath    string              `json:"manifest_path"`
}

// OrphanHandleResult tracks how orphans were handled.
type OrphanHandleResult struct {
	Strategy string   `json:"strategy"`
	Agents   []string `json:"agents"`
}

// Text implements Textable for TeamSwitchOutput.
func (s TeamSwitchOutput) Text() string {
	// Silent on success per TDD
	return ""
}

// TeamSwitchDryRunOutput represents dry-run result.
type TeamSwitchDryRunOutput struct {
	DryRun                 bool     `json:"dry_run"`
	WouldSwitchTo          string   `json:"would_switch_to"`
	CurrentTeam            string   `json:"current_team"`
	WouldInstall           []string `json:"would_install"`
	OrphansDetected        []string `json:"orphans_detected"`
	OrphanStrategyRequired bool     `json:"orphan_strategy_required"`
	SuggestedFlags         []string `json:"suggested_flags,omitempty"`
}

// Text implements Textable for TeamSwitchDryRunOutput.
func (s TeamSwitchDryRunOutput) Text() string {
	var b strings.Builder
	b.WriteString("[DRY RUN]\n")
	b.WriteString(fmt.Sprintf("Would switch to: %s\n", s.WouldSwitchTo))
	b.WriteString(fmt.Sprintf("Current team: %s\n", s.CurrentTeam))
	b.WriteString(fmt.Sprintf("Would install: %s\n", strings.Join(s.WouldInstall, ", ")))
	if len(s.OrphansDetected) > 0 {
		b.WriteString(fmt.Sprintf("Orphans detected: %s\n", strings.Join(s.OrphansDetected, ", ")))
		if s.OrphanStrategyRequired {
			b.WriteString(fmt.Sprintf("Suggested flags: %s\n", strings.Join(s.SuggestedFlags, ", ")))
		}
	}
	return b.String()
}

// TeamValidateOutput represents validation result.
type TeamValidateOutput struct {
	Team     string               `json:"team"`
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

// Text implements Textable for TeamValidateOutput.
func (v TeamValidateOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Validating team: %s\n\n", v.Team))

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
