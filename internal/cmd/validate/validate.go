// Package validate implements the ari validate commands for artifact validation.
package validate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/validation"
)

// cmdContext holds shared state for validate commands.
type cmdContext struct {
	common.BaseContext
}

// NewValidateCmd creates the validate command group.
func NewValidateCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate artifacts and configurations",
		Long: `Validate workflow artifacts (PRD, TDD, ADR, Test Plans) against schemas
and handoff criteria.

Examples:
  ari validate artifact --type=prd .ledge/specs/PRD-user-auth.md
  ari validate handoff --phase=requirements --artifact=PRD-user-auth
  ari validate schema --file=SESSION_CONTEXT.md`,
	}

	// Add subcommands
	cmd.AddCommand(newArtifactCmd(ctx))
	cmd.AddCommand(newHandoffCmd(ctx))
	cmd.AddCommand(newSchemaCmd(ctx))

	// Validate commands require project context
	common.SetNeedsProject(cmd, true, true)

	return cmd
}

// helper functions for commands

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}

// ArtifactOutput represents the JSON output for artifact validation.
type ArtifactOutput struct {
	Valid        bool                         `json:"valid"`
	ArtifactType string                       `json:"artifact_type,omitempty"`
	FilePath     string                       `json:"file_path"`
	Issues       []validation.ValidationIssue `json:"issues,omitempty"`
	Frontmatter  map[string]any               `json:"frontmatter,omitempty"`
}

// Text implements output.Textable for ArtifactOutput.
func (a ArtifactOutput) Text() string {
	var b strings.Builder

	if a.Valid {
		fmt.Fprintf(&b, "VALID: %s\n", a.FilePath)
		if a.ArtifactType != "" {
			fmt.Fprintf(&b, "  Type: %s\n", a.ArtifactType)
		}
	} else {
		fmt.Fprintf(&b, "INVALID: %s\n", a.FilePath)
		if a.ArtifactType != "" {
			fmt.Fprintf(&b, "  Type: %s\n", a.ArtifactType)
		}
		b.WriteString("  Issues:\n")
		for _, issue := range a.Issues {
			if issue.Field != "" {
				fmt.Fprintf(&b, "    - [%s] %s\n", issue.Field, issue.Message)
			} else {
				fmt.Fprintf(&b, "    - %s\n", issue.Message)
			}
		}
	}

	return b.String()
}

// newArtifactCmd creates the artifact validation subcommand.
func newArtifactCmd(ctx *cmdContext) *cobra.Command {
	var artifactType string

	cmd := &cobra.Command{
		Use:   "artifact [file]",
		Short: "Validate an artifact file against its schema",
		Long: `Validate a workflow artifact (PRD, TDD, ADR, Test Plan) against
its corresponding JSON schema.

The artifact type can be auto-detected from:
  1. The frontmatter 'type' field
  2. The filename pattern (PRD-*.md, TDD-*.md, ADR-*.md, TEST-*.md, TP-*.md)
  3. The --type flag (explicit override)

Examples:
  ari validate artifact .ledge/specs/PRD-user-auth.md
  ari validate artifact --type=prd .ledge/specs/PRD-user-auth.md
  ari validate artifact --type=tdd .ledge/specs/TDD-user-auth.md
  ari validate artifact .ledge/decisions/ADR-0001.md`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			filePath := args[0]

			// Resolve to absolute path if needed
			if !filepath.IsAbs(filePath) {
				resolver := ctx.GetResolver()
				filePath = filepath.Join(resolver.ProjectRoot(), filePath)
			}

			// Create validator
			validator, err := validation.NewArtifactValidator()
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			// Parse artifact type if specified
			var aType validation.ArtifactType
			if artifactType != "" {
				aType = validation.ParseArtifactType(artifactType)
				if aType == validation.ArtifactTypeUnknown {
					err := errors.NewWithDetails(errors.CodeUsageError,
						"invalid artifact type",
						map[string]any{
							"type":  artifactType,
							"valid": validation.ValidArtifactTypes(),
						})
					return common.PrintAndReturn(printer, err)
				}
			}

			// Validate the file
			result, err := validator.ValidateFile(filePath, aType)
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			// Create output
			out := ArtifactOutput{
				Valid:        result.Valid,
				ArtifactType: string(result.ArtifactType),
				FilePath:     result.FilePath,
				Issues:       result.Issues,
				Frontmatter:  result.Frontmatter,
			}

			if err := printer.Print(out); err != nil {
				return err
			}

			// Return error if validation failed
			if !result.Valid {
				return errors.ErrSchemaInvalid(result.FilePath, issueMessages(result.Issues))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&artifactType, "type", "t", "", "Artifact type (prd, tdd, adr, test-plan). Auto-detected if not specified.")

	return cmd
}

// issueMessages extracts message strings from validation issues.
func issueMessages(issues []validation.ValidationIssue) []string {
	msgs := make([]string, len(issues))
	for i, issue := range issues {
		if issue.Field != "" {
			msgs[i] = fmt.Sprintf("[%s] %s", issue.Field, issue.Message)
		} else {
			msgs[i] = issue.Message
		}
	}
	return msgs
}

// HandoffOutput represents the JSON output for handoff validation.
type HandoffOutput struct {
	Passed         bool                         `json:"passed"`
	Phase          string                       `json:"phase"`
	ArtifactType   string                       `json:"artifact_type,omitempty"`
	FilePath       string                       `json:"file_path,omitempty"`
	BlockingFailed []validation.CriterionResult `json:"blocking_failed,omitempty"`
	Warnings       []validation.CriterionResult `json:"warnings,omitempty"`
}

// Text implements output.Textable for HandoffOutput.
func (h HandoffOutput) Text() string {
	var b strings.Builder

	if h.Passed {
		fmt.Fprintf(&b, "PASSED: %s handoff for %s\n", h.Phase, h.ArtifactType)
		if h.FilePath != "" {
			fmt.Fprintf(&b, "  File: %s\n", h.FilePath)
		}
	} else {
		fmt.Fprintf(&b, "FAILED: %s handoff for %s\n", h.Phase, h.ArtifactType)
		if h.FilePath != "" {
			fmt.Fprintf(&b, "  File: %s\n", h.FilePath)
		}
		if len(h.BlockingFailed) > 0 {
			b.WriteString("  Blocking failures:\n")
			for _, cr := range h.BlockingFailed {
				fmt.Fprintf(&b, "    - [%s] %s\n", cr.Criterion.Field, cr.Message)
			}
		}
	}

	if len(h.Warnings) > 0 {
		b.WriteString("  Warnings:\n")
		for _, cr := range h.Warnings {
			fmt.Fprintf(&b, "    - [%s] %s\n", cr.Criterion.Field, cr.Message)
		}
	}

	return b.String()
}

// PhaseCriteriaOutput represents the JSON output for listing phase criteria.
type PhaseCriteriaOutput struct {
	Phases []PhaseInfo `json:"phases"`
}

// PhaseInfo contains information about a phase and its artifact types.
type PhaseInfo struct {
	Phase         string   `json:"phase"`
	ArtifactTypes []string `json:"artifact_types"`
}

// Text implements output.Textable for PhaseCriteriaOutput.
func (p PhaseCriteriaOutput) Text() string {
	var b strings.Builder
	b.WriteString("Phases with handoff criteria:\n")
	for _, info := range p.Phases {
		fmt.Fprintf(&b, "  %s:\n", info.Phase)
		for _, at := range info.ArtifactTypes {
			fmt.Fprintf(&b, "    - %s\n", at)
		}
	}
	return b.String()
}

// CriteriaDetailOutput represents the JSON output for showing criteria details.
type CriteriaDetailOutput struct {
	Phase        string                 `json:"phase"`
	ArtifactType string                 `json:"artifact_type"`
	Blocking     []validation.Criterion `json:"blocking,omitempty"`
	NonBlocking  []validation.Criterion `json:"non_blocking,omitempty"`
}

// Text implements output.Textable for CriteriaDetailOutput.
func (c CriteriaDetailOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Handoff criteria for %s/%s:\n", c.Phase, c.ArtifactType)

	if len(c.Blocking) > 0 {
		b.WriteString("  Blocking:\n")
		for _, cr := range c.Blocking {
			fmt.Fprintf(&b, "    - %s: %s\n", cr.Field, cr.Message)
		}
	}

	if len(c.NonBlocking) > 0 {
		b.WriteString("  Non-blocking:\n")
		for _, cr := range c.NonBlocking {
			fmt.Fprintf(&b, "    - %s: %s\n", cr.Field, cr.Message)
		}
	}

	return b.String()
}

// newHandoffCmd creates the handoff criteria validation subcommand.
func newHandoffCmd(ctx *cmdContext) *cobra.Command {
	var (
		phase        string
		artifactPath string
		artifactType string
		listPhases   bool
		showCriteria bool
	)

	cmd := &cobra.Command{
		Use:   "handoff",
		Short: "Validate handoff criteria for phase transitions",
		Long: `Validate that artifacts meet handoff criteria for transitioning
between workflow phases.

The handoff criteria are defined in schemas/handoff-criteria.yaml
and include blocking and non-blocking checks for each artifact type.

Phases: requirements, design, implementation, validation

Examples:
  ari validate handoff --phase=requirements --artifact=.ledge/specs/PRD-user-auth.md
  ari validate handoff --list-phases
  ari validate handoff --phase=requirements --type=prd --show-criteria`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()

			// Create handoff validator
			hv, err := validation.NewHandoffValidator()
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			// Handle --list-phases
			if listPhases {
				phases := hv.ListPhases()
				out := PhaseCriteriaOutput{Phases: make([]PhaseInfo, 0, len(phases))}
				for _, p := range phases {
					artifactTypes := hv.ListArtifactTypes(p)
					typeStrs := make([]string, len(artifactTypes))
					for i, at := range artifactTypes {
						typeStrs[i] = string(at)
					}
					out.Phases = append(out.Phases, PhaseInfo{
						Phase:         string(p),
						ArtifactTypes: typeStrs,
					})
				}
				return printer.Print(out)
			}

			// Handle --show-criteria
			if showCriteria {
				if phase == "" {
					err := errors.New(errors.CodeUsageError, "--phase is required with --show-criteria")
					return common.PrintAndReturn(printer, err)
				}
				if artifactType == "" {
					err := errors.New(errors.CodeUsageError, "--type is required with --show-criteria")
					return common.PrintAndReturn(printer, err)
				}

				p := validation.ParsePhase(phase)
				if p == "" {
					err := errors.NewWithDetails(errors.CodeUsageError,
						"invalid phase",
						map[string]any{"phase": phase, "valid": validation.ValidPhases()})
					return common.PrintAndReturn(printer, err)
				}

				at := validation.ParseArtifactType(artifactType)
				if at == validation.ArtifactTypeUnknown {
					err := errors.NewWithDetails(errors.CodeUsageError,
						"invalid artifact type",
						map[string]any{"type": artifactType, "valid": validation.ValidArtifactTypes()})
					return common.PrintAndReturn(printer, err)
				}

				criteria, err := hv.GetCriteria(p, at)
				if err != nil {
					return common.PrintAndReturn(printer, err)
				}

				out := CriteriaDetailOutput{
					Phase:        phase,
					ArtifactType: artifactType,
					Blocking:     criteria.Blocking,
					NonBlocking:  criteria.NonBlocking,
				}
				return printer.Print(out)
			}

			// Validate handoff - requires --phase and --artifact
			if phase == "" {
				err := errors.New(errors.CodeUsageError, "--phase is required for handoff validation")
				return common.PrintAndReturn(printer, err)
			}
			if artifactPath == "" {
				err := errors.New(errors.CodeUsageError, "--artifact is required for handoff validation")
				return common.PrintAndReturn(printer, err)
			}

			p := validation.ParsePhase(phase)
			if p == "" {
				err := errors.NewWithDetails(errors.CodeUsageError,
					"invalid phase",
					map[string]any{"phase": phase, "valid": validation.ValidPhases()})
				return common.PrintAndReturn(printer, err)
			}

			// Resolve artifact path
			filePath := artifactPath
			if !filepath.IsAbs(filePath) {
				resolver := ctx.GetResolver()
				filePath = filepath.Join(resolver.ProjectRoot(), filePath)
			}

			// Validate handoff
			result, err := hv.ValidateHandoffFile(p, filePath)
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			out := HandoffOutput{
				Passed:         result.Passed,
				Phase:          string(result.Phase),
				ArtifactType:   string(result.ArtifactType),
				FilePath:       result.FilePath,
				BlockingFailed: result.FailedBlocking(),
				Warnings:       result.Warnings(),
			}

			if err := printer.Print(out); err != nil {
				return err
			}

			// Return error if handoff failed
			if !result.Passed {
				failed := result.FailedBlocking()
				msgs := make([]string, len(failed))
				for i, cr := range failed {
					msgs[i] = fmt.Sprintf("[%s] %s", cr.Criterion.Field, cr.Message)
				}
				return errors.ErrSchemaInvalid(result.FilePath, msgs)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Workflow phase to validate for (requirements, design, implementation, validation)")
	cmd.Flags().StringVar(&artifactPath, "artifact", "", "Path to artifact file to validate")
	cmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (prd, tdd, adr, test-plan) for --show-criteria")
	cmd.Flags().BoolVar(&listPhases, "list-phases", false, "List all phases with handoff criteria")
	cmd.Flags().BoolVar(&showCriteria, "show-criteria", false, "Show criteria for a specific phase/type")

	return cmd
}

// SchemaOutput represents the JSON output for schema validation.
type SchemaOutput struct {
	Valid      bool                         `json:"valid"`
	SchemaName string                       `json:"schema_name"`
	FilePath   string                       `json:"file_path"`
	Issues     []validation.ValidationIssue `json:"issues,omitempty"`
}

// Text implements output.Textable for SchemaOutput.
func (s SchemaOutput) Text() string {
	var b strings.Builder

	if s.Valid {
		fmt.Fprintf(&b, "VALID: %s (schema: %s)\n", s.FilePath, s.SchemaName)
	} else {
		fmt.Fprintf(&b, "INVALID: %s (schema: %s)\n", s.FilePath, s.SchemaName)
		b.WriteString("  Issues:\n")
		for _, issue := range s.Issues {
			if issue.Field != "" {
				fmt.Fprintf(&b, "    - [%s] %s\n", issue.Field, issue.Message)
			} else {
				fmt.Fprintf(&b, "    - %s\n", issue.Message)
			}
		}
	}

	return b.String()
}

// newSchemaCmd creates the schema validation subcommand.
func newSchemaCmd(ctx *cmdContext) *cobra.Command {
	var schemaName string

	cmd := &cobra.Command{
		Use:   "schema [schema-name] [file]",
		Short: "Validate a file against a specific schema",
		Long: `Validate a file's YAML frontmatter against a specified JSON schema.

Available schemas:
  prd         - Product Requirements Document
  tdd         - Technical Design Document
  adr         - Architecture Decision Record
  test-plan   - Test Plan

Examples:
  ari validate schema prd .ledge/specs/PRD-user-auth.md
  ari validate schema tdd .ledge/specs/TDD-user-auth.md
  ari validate schema adr .ledge/decisions/ADR-0001.md`,
		Args: common.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			schemaName := args[0]
			filePath := args[1]

			// Resolve to absolute path if needed
			if !filepath.IsAbs(filePath) {
				resolver := ctx.GetResolver()
				filePath = filepath.Join(resolver.ProjectRoot(), filePath)
			}

			// Parse schema name as artifact type
			aType := validation.ParseArtifactType(schemaName)
			if aType == validation.ArtifactTypeUnknown {
				err := errors.NewWithDetails(errors.CodeSchemaNotFound,
					"unknown schema",
					map[string]any{
						"schema": schemaName,
						"valid":  validation.ValidArtifactTypes(),
					})
				return common.PrintAndReturn(printer, err)
			}

			// Create validator
			validator, err := validation.NewArtifactValidator()
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			// Validate the file
			result, err := validator.ValidateFile(filePath, aType)
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			// Create output
			out := SchemaOutput{
				Valid:      result.Valid,
				SchemaName: schemaName,
				FilePath:   result.FilePath,
				Issues:     result.Issues,
			}

			if err := printer.Print(out); err != nil {
				return err
			}

			// Return error if validation failed
			if !result.Valid {
				return errors.ErrSchemaInvalid(result.FilePath, issueMessages(result.Issues))
			}

			return nil
		},
	}

	// Keep the flag for backwards compatibility but it's not used
	cmd.Flags().StringVar(&schemaName, "schema", "", "Schema name (deprecated, use positional argument)")

	return cmd
}
