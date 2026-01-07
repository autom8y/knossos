package rite

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/autom8y/ariadne/internal/paths"
)

// CheckStatus represents the result of a validation check.
type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckFail CheckStatus = "fail"
	CheckWarn CheckStatus = "warn"
)

// ValidationCheck represents a single validation check result.
type ValidationCheck struct {
	Check   string      `json:"check"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
}

// ValidationResult holds the complete validation results.
type ValidationResult struct {
	Team     string            `json:"team"`
	Valid    bool              `json:"valid"`
	Checks   []ValidationCheck `json:"checks"`
	Errors   int               `json:"errors"`
	Warnings int               `json:"warnings"`
	Fixable  []string          `json:"fixable,omitempty"`
}

// Validator validates team pack integrity.
type Validator struct {
	resolver  *paths.Resolver
	discovery *Discovery
}

// NewValidator creates a new team validator.
func NewValidator(resolver *paths.Resolver) *Validator {
	return &Validator{
		resolver:  resolver,
		discovery: NewDiscovery(resolver),
	}
}

// Validate performs all validation checks on a team.
func (v *Validator) Validate(teamName string) (*ValidationResult, error) {
	result := &ValidationResult{
		Team:   teamName,
		Valid:  true,
		Checks: []ValidationCheck{},
	}

	// Get team info (also validates existence)
	team, err := v.discovery.Get(teamName)
	if err != nil {
		result.Valid = false
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "TEAM_EXISTS",
			Status:  CheckFail,
			Message: "Rite not found: " + teamName,
		})
		result.Errors++
		return result, nil
	}

	// Run all checks
	v.checkTeamExists(result, team)
	v.checkAgentsDir(result, team)
	v.checkWorkflowYAML(result, team)
	v.checkAgentFiles(result, team)
	v.checkManifestSync(result, team)
	v.checkClaudeMDSync(result, team)
	v.checkValidEntryPoint(result, team)

	// Set overall validity
	result.Valid = result.Errors == 0

	return result, nil
}

// checkTeamExists verifies the team directory exists.
func (v *Validator) checkTeamExists(result *ValidationResult, rite *Rite) {
	if _, err := os.Stat(rite.Path); os.IsNotExist(err) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "TEAM_EXISTS",
			Status:  CheckFail,
			Message: "Rite directory not found",
		})
		result.Errors++
	} else {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "TEAM_EXISTS",
			Status:  CheckPass,
			Message: "Rite directory found",
		})
	}
}

// checkAgentsDir verifies the agents/ subdirectory exists.
func (v *Validator) checkAgentsDir(result *ValidationResult, rite *Rite) {
	agentsDir := filepath.Join(rite.Path, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "AGENTS_DIR",
			Status:  CheckFail,
			Message: "agents/ directory missing",
		})
		result.Errors++
	} else {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "AGENTS_DIR",
			Status:  CheckPass,
			Message: "agents/ directory exists",
		})
	}
}

// checkWorkflowYAML verifies workflow.yaml exists and is valid.
func (v *Validator) checkWorkflowYAML(result *ValidationResult, rite *Rite) {
	workflowPath := filepath.Join(rite.Path, "workflow.yaml")
	workflow, err := LoadWorkflow(workflowPath)
	if err != nil {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "WORKFLOW_YAML",
			Status:  CheckFail,
			Message: "workflow.yaml invalid: " + err.Error(),
		})
		result.Errors++
		return
	}

	// Check required fields
	if workflow.Name == "" {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "WORKFLOW_YAML",
			Status:  CheckWarn,
			Message: "workflow.yaml missing 'name' field",
		})
		result.Warnings++
	} else if workflow.EntryPoint.Agent == "" {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "WORKFLOW_YAML",
			Status:  CheckWarn,
			Message: "workflow.yaml missing 'entry_point.agent' field",
		})
		result.Warnings++
	} else {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "WORKFLOW_YAML",
			Status:  CheckPass,
			Message: "workflow.yaml is valid YAML",
		})
	}
}

// checkAgentFiles verifies all referenced agents have .md files.
func (v *Validator) checkAgentFiles(result *ValidationResult, rite *Rite) {
	agentsDir := filepath.Join(rite.Path, "agents")

	// Get expected agents from workflow
	workflowPath := filepath.Join(rite.Path, "workflow.yaml")
	workflow, err := LoadWorkflow(workflowPath)
	if err != nil {
		// Already reported in checkWorkflowYAML
		return
	}

	expectedAgents := workflow.AgentNames()
	var missing []string

	for _, agentName := range expectedAgents {
		agentFile := filepath.Join(agentsDir, agentName+".md")
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			missing = append(missing, agentName+".md")
		}
	}

	if len(missing) > 0 {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "AGENT_FILES",
			Status:  CheckFail,
			Message: "Missing agent files: " + strings.Join(missing, ", "),
		})
		result.Errors++
	} else {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "AGENT_FILES",
			Status:  CheckPass,
			Message: "All " + strconv.Itoa(len(expectedAgents)) + " agent files present",
		})
	}
}

// checkManifestSync verifies AGENT_MANIFEST.json matches installed agents.
func (v *Validator) checkManifestSync(result *ValidationResult, rite *Rite) {
	// Only check if this is the active rite
	if !rite.Active {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "MANIFEST_SYNC",
			Status:  CheckPass,
			Message: "Skipped (rite not active)",
		})
		return
	}

	manifest, err := LoadAgentManifest(v.resolver.AgentManifestFile())
	if err != nil {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "MANIFEST_SYNC",
			Status:  CheckWarn,
			Message: "Could not load manifest: " + err.Error(),
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "MANIFEST_SYNC")
		return
	}

	// Check if manifest's active rite matches
	if manifest.ActiveRite != rite.Name {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "MANIFEST_SYNC",
			Status:  CheckWarn,
			Message: "Manifest active rite mismatch: " + manifest.ActiveRite,
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "MANIFEST_SYNC")
		return
	}

	// Check installed agents match manifest
	agentsDir := v.resolver.AgentsDir()
	installedFiles, _ := listAgentFiles(agentsDir)
	manifestAgents := manifest.GetRiteAgents(rite.Name)

	// Quick check for count mismatch
	if len(installedFiles) != len(manifestAgents) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "MANIFEST_SYNC",
			Status:  CheckWarn,
			Message: "Installed agent count differs from manifest",
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "MANIFEST_SYNC")
		return
	}

	result.Checks = append(result.Checks, ValidationCheck{
		Check:   "MANIFEST_SYNC",
		Status:  CheckPass,
		Message: "Manifest matches installed agents",
	})
}

// checkClaudeMDSync verifies CLAUDE.md satellite sections match active rite.
func (v *Validator) checkClaudeMDSync(result *ValidationResult, rite *Rite) {
	// Only check if this is the active rite
	if !rite.Active {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "CLAUDE_MD_SYNC",
			Status:  CheckPass,
			Message: "Skipped (rite not active)",
		})
		return
	}

	claudeMDPath := v.resolver.ClaudeMDFile()
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "CLAUDE_MD_SYNC",
			Status:  CheckWarn,
			Message: "Could not read CLAUDE.md: " + err.Error(),
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "CLAUDE_MD_SYNC")
		return
	}

	// Check if rite name appears in Quick Start section
	if !strings.Contains(string(content), rite.Name) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "CLAUDE_MD_SYNC",
			Status:  CheckWarn,
			Message: "CLAUDE.md does not reference active rite",
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "CLAUDE_MD_SYNC")
		return
	}

	result.Checks = append(result.Checks, ValidationCheck{
		Check:   "CLAUDE_MD_SYNC",
		Status:  CheckPass,
		Message: "CLAUDE.md satellites synced",
	})
}

// checkValidEntryPoint verifies the entry point agent exists.
func (v *Validator) checkValidEntryPoint(result *ValidationResult, rite *Rite) {
	if rite.EntryPoint == "" {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "VALID_ENTRY_POINT",
			Status:  CheckWarn,
			Message: "No entry point defined",
		})
		result.Warnings++
		return
	}

	agentFile := filepath.Join(rite.Path, "agents", rite.EntryPoint+".md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "VALID_ENTRY_POINT",
			Status:  CheckFail,
			Message: "Entry point agent not found: " + rite.EntryPoint,
		})
		result.Errors++
		return
	}

	result.Checks = append(result.Checks, ValidationCheck{
		Check:   "VALID_ENTRY_POINT",
		Status:  CheckPass,
		Message: "Entry point '" + rite.EntryPoint + "' exists",
	})
}

// Fix attempts to repair fixable issues.
func (v *Validator) Fix(teamName string) error {
	result, err := v.Validate(teamName)
	if err != nil {
		return err
	}

	for _, fixable := range result.Fixable {
		switch fixable {
		case "MANIFEST_SYNC":
			// Regenerate manifest from current state
			team, err := v.discovery.Get(teamName)
			if err != nil {
				continue
			}
			if team.Active {
				// Would need to run switch --update
			}
		case "CLAUDE_MD_SYNC":
			// Update CLAUDE.md satellites
			team, err := v.discovery.Get(teamName)
			if err != nil {
				continue
			}
			if team.Active {
				updater := NewClaudeMDUpdater(v.resolver.ClaudeMDFile())
				updater.UpdateForTeam(team)
			}
		}
	}

	return nil
}

// listAgentFiles returns a list of agent filenames in a directory.
func listAgentFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
