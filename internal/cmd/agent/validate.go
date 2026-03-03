package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

type validateOptions struct {
	riteName string
	strict   bool
	all      bool
}

func newValidateCmd(ctx *cmdContext) *cobra.Command {
	var opts validateOptions

	cmd := &cobra.Command{
		Use:   "validate [path...]",
		Short: "Validate agent specifications",
		Long: `Validates agent frontmatter against the agent JSON schema.

Examples:
  ari agent validate                              # Validate all agents
  ari agent validate --rite ecosystem            # Validate agents in ecosystem rite
  ari agent validate --strict                    # Strict validation (requires enhanced fields)
  ari agent validate agents/moirai.md             # Validate specific agent file
  ari agent validate rites/*/agents/*.md         # Validate all rite agents

Exit Codes:
  0 - All agents valid
  1 - Validation errors found`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(ctx, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to validate (validates all agents in rite)")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Enable strict validation mode (requires enhanced fields)")
	cmd.Flags().BoolVar(&opts.all, "all", false, "Validate all agents in all rites and agents")

	return cmd
}

func runValidate(ctx *cmdContext, opts validateOptions, paths []string) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	// Create validator
	validator, err := agentpkg.NewAgentValidator()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Determine validation mode
	mode := agentpkg.ValidationModeWarn
	if opts.strict {
		mode = agentpkg.ValidationModeStrict
	}

	// Collect agent paths to validate
	var agentPaths []string

	if len(paths) > 0 {
		// Validate specific paths provided by user
		agentPaths = paths
	} else if opts.riteName != "" {
		// Validate all agents in specific rite
		ritePath := resolver.RiteDir(opts.riteName)
		agentsDir := filepath.Join(ritePath, "agents")
		ritePaths, err := collectAgentsInDir(agentsDir)
		if err != nil {
			printer.PrintError(err)
			return err
		}
		agentPaths = ritePaths
	} else if opts.all || (len(paths) == 0 && opts.riteName == "") {
		// Validate all agents (default behavior)
		allPaths, err := collectAllAgents(resolver)
		if err != nil {
			printer.PrintError(err)
			return err
		}
		agentPaths = allPaths
	}

	if len(agentPaths) == 0 {
		printer.Print("No agent files found to validate")
		return nil
	}

	// Validate all collected agents
	results := make(map[string]*agentpkg.AgentValidationResult)
	var validCount, errorCount, warningCount int

	for _, agentPath := range agentPaths {
		result, err := validator.ValidateAgentFile(agentPath, mode)
		if err != nil {
			// File not found or read error
			printer.VerboseLog("error", "validation failed", map[string]interface{}{
				"path":  agentPath,
				"error": err.Error(),
			})
			errorCount++
			continue
		}

		results[agentPath] = result
		if result.Valid {
			validCount++
		} else {
			errorCount++
		}
		warningCount += len(result.Warnings)
	}

	// Build structured output
	out := buildValidationOutput(results, resolver, len(agentPaths), validCount, errorCount, warningCount)

	if err := printer.Print(out); err != nil {
		return err
	}

	// Return error exit code if any validation failed
	if errorCount > 0 {
		return errors.New(errors.CodeValidationFailed, "Agent validation failed")
	}

	return nil
}

func collectAgentsInDir(dir string) ([]string, error) {
	var paths []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return paths, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	return paths, nil
}

func collectAllAgents(resolver *paths.Resolver) ([]string, error) {
	var allPaths []string

	// Collect agents from all rites
	ritesDir := resolver.RitesDir()
	riteEntries, err := os.ReadDir(ritesDir)
	if err == nil {
		for _, riteEntry := range riteEntries {
			if !riteEntry.IsDir() {
				continue
			}
			agentsDir := filepath.Join(ritesDir, riteEntry.Name(), "agents")
			ritePaths, err := collectAgentsInDir(agentsDir)
			if err == nil {
				allPaths = append(allPaths, ritePaths...)
			}
		}
	}

	// Collect agents from agents/
	projectRoot := resolver.ProjectRoot()
	userAgentsDir := filepath.Join(projectRoot, "agents")
	userPaths, err := collectAgentsInDir(userAgentsDir)
	if err == nil {
		allPaths = append(allPaths, userPaths...)
	}

	return allPaths, nil
}

// agentValidateOutput is the structured output for ari agent validate.
type agentValidateOutput struct {
	Agents       []agentValidateEntry `json:"agents"`
	TotalScanned int                  `json:"total_scanned"`
	Valid        int                  `json:"valid"`
	Errors       int                  `json:"errors"`
	Warnings     int                  `json:"warnings"`
}

type agentValidateEntry struct {
	Path     string                    `json:"path"`
	Status   string                    `json:"status"` // "pass", "warn", "fail"
	Issues   []agentpkg.ValidationIssue `json:"issues,omitempty"`
	Warnings []string                  `json:"warnings,omitempty"`
}

// Text implements output.Textable.
func (v agentValidateOutput) Text() string {
	var b strings.Builder

	for _, entry := range v.Agents {
		switch entry.Status {
		case "pass":
			b.WriteString(fmt.Sprintf("PASS  %s\n", entry.Path))
		case "warn":
			b.WriteString(fmt.Sprintf("WARN  %s\n", entry.Path))
		case "fail":
			b.WriteString(fmt.Sprintf("FAIL  %s\n", entry.Path))
		}

		for _, issue := range entry.Issues {
			if issue.Field != "" {
				b.WriteString(fmt.Sprintf("  ERROR: %s: %s\n", issue.Field, issue.Message))
			} else {
				b.WriteString(fmt.Sprintf("  ERROR: %s\n", issue.Message))
			}
			if issue.Value != nil {
				b.WriteString(fmt.Sprintf("         value: %v\n", issue.Value))
			}
		}

		for _, warning := range entry.Warnings {
			b.WriteString(fmt.Sprintf("  WARN: %s\n", warning))
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Summary: %d agents validated\n", v.TotalScanned))
	b.WriteString(fmt.Sprintf("  Valid: %d\n", v.Valid))
	b.WriteString(fmt.Sprintf("  Errors: %d\n", v.Errors))
	if v.Warnings > 0 {
		b.WriteString(fmt.Sprintf("  Warnings: %d\n", v.Warnings))
	}

	return b.String()
}

func buildValidationOutput(results map[string]*agentpkg.AgentValidationResult, resolver *paths.Resolver, totalScanned, validCount, errorCount, warningCount int) agentValidateOutput {
	projectRoot := resolver.ProjectRoot()

	var entries []agentValidateEntry
	for path, result := range results {
		relPath, err := filepath.Rel(projectRoot, path)
		if err != nil {
			relPath = path
		}

		status := "pass"
		if !result.Valid {
			status = "fail"
		} else if len(result.Warnings) > 0 {
			status = "warn"
		}

		entries = append(entries, agentValidateEntry{
			Path:     relPath,
			Status:   status,
			Issues:   result.Issues,
			Warnings: result.Warnings,
		})
	}

	return agentValidateOutput{
		Agents:       entries,
		TotalScanned: totalScanned,
		Valid:        validCount,
		Errors:       errorCount,
		Warnings:     warningCount,
	}
}
