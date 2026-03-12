package rite

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/autom8y/knossos/internal/paths"
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
	Rite     string            `json:"rite"`
	Valid    bool              `json:"valid"`
	Checks   []ValidationCheck `json:"checks"`
	Errors   int               `json:"errors"`
	Warnings int               `json:"warnings"`
	Fixable  []string          `json:"fixable,omitempty"`
}

// Validator validates rite integrity.
type Validator struct {
	resolver  *paths.Resolver
	discovery *Discovery
	syncer    Syncer // injected; breaks dependency on materialize package
}

// NewValidator creates a new rite validator.
func NewValidator(resolver *paths.Resolver, syncer Syncer) *Validator {
	return &Validator{
		resolver:  resolver,
		discovery: NewDiscovery(resolver),
		syncer:    syncer,
	}
}

// Validate performs all validation checks on a rite.
func (v *Validator) Validate(riteName string) (*ValidationResult, error) {
	result := &ValidationResult{
		Rite:   riteName,
		Valid:  true,
		Checks: []ValidationCheck{},
	}

	// Get rite info (also validates existence)
	rite, err := v.discovery.Get(riteName)
	if err != nil {
		result.Valid = false
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "RITE_EXISTS",
			Status:  CheckFail,
			Message: "Rite not found: " + riteName,
		})
		result.Errors++
		return result, nil
	}

	// Run all checks
	v.checkRiteExists(result, rite)
	v.checkAgentsDir(result, rite)
	v.checkWorkflowYAML(result, rite)
	v.checkAgentFiles(result, rite)
	v.checkInscriptionSync(result, rite)
	v.checkValidEntryPoint(result, rite)

	// Set overall validity
	result.Valid = result.Errors == 0

	return result, nil
}

// checkRiteExists verifies the rite directory exists.
func (v *Validator) checkRiteExists(result *ValidationResult, rite *Rite) {
	if _, err := os.Stat(rite.Path); os.IsNotExist(err) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "RITE_EXISTS",
			Status:  CheckFail,
			Message: "Rite directory not found",
		})
		result.Errors++
	} else {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "RITE_EXISTS",
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
	switch {
	case workflow.Name == "":
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "WORKFLOW_YAML",
			Status:  CheckWarn,
			Message: "workflow.yaml missing 'name' field",
		})
		result.Warnings++
	case workflow.EntryPoint.Agent == "":
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "WORKFLOW_YAML",
			Status:  CheckWarn,
			Message: "workflow.yaml missing 'entry_point.agent' field",
		})
		result.Warnings++
	default:
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


// checkInscriptionSync verifies inscription file satellite sections match active rite.
// Checks CLAUDE.md (the CC inscription compilation target) for rite alignment.
func (v *Validator) checkInscriptionSync(result *ValidationResult, rite *Rite) {
	// Only check if this is the active rite
	if !rite.Active {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "INSCRIPTION_SYNC",
			Status:  CheckPass,
			Message: "Skipped (rite not active)",
		})
		return
	}

	contextFilePath := v.resolver.ContextFileForChannel(paths.ClaudeChannel{})
	content, err := os.ReadFile(contextFilePath)
	if err != nil {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "INSCRIPTION_SYNC",
			Status:  CheckWarn,
			Message: "Could not read CLAUDE.md: " + err.Error(),
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "INSCRIPTION_SYNC")
		return
	}

	// Check if rite name appears in Quick Start section
	if !strings.Contains(string(content), rite.Name) {
		result.Checks = append(result.Checks, ValidationCheck{
			Check:   "INSCRIPTION_SYNC",
			Status:  CheckWarn,
			Message: "CLAUDE.md does not reference active rite",
		})
		result.Warnings++
		result.Fixable = append(result.Fixable, "INSCRIPTION_SYNC")
		return
	}

	result.Checks = append(result.Checks, ValidationCheck{
		Check:   "INSCRIPTION_SYNC",
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
func (v *Validator) Fix(riteName string) error {
	result, err := v.Validate(riteName)
	if err != nil {
		return err
	}

	for _, fixable := range result.Fixable {
		if fixable == "INSCRIPTION_SYNC" {
			rite, err := v.discovery.Get(riteName)
			if err != nil {
				continue
			}
			if rite.Active && v.syncer != nil {
				if err := v.syncer.SyncRite(riteName, true); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

