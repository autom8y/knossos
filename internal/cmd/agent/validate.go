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

	// Print results
	printValidationResults(printer, results, resolver)

	// Print summary
	fmt.Fprintf(os.Stdout, "\n")
	fmt.Fprintf(os.Stdout, "Summary: %d agents validated\n", len(agentPaths))
	fmt.Fprintf(os.Stdout, "  Valid: %d\n", validCount)
	fmt.Fprintf(os.Stdout, "  Errors: %d\n", errorCount)
	if warningCount > 0 {
		fmt.Fprintf(os.Stdout, "  Warnings: %d\n", warningCount)
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

func printValidationResults(printer *output.Printer, results map[string]*agentpkg.AgentValidationResult, resolver *paths.Resolver) {
	projectRoot := resolver.ProjectRoot()

	for path, result := range results {
		// Make path relative to project root for display
		relPath, err := filepath.Rel(projectRoot, path)
		if err != nil {
			relPath = path
		}

		// Print status indicator
		if result.Valid {
			if len(result.Warnings) > 0 {
				fmt.Fprintf(os.Stdout, "WARN  %s\n", relPath)
			} else {
				fmt.Fprintf(os.Stdout, "PASS  %s\n", relPath)
			}
		} else {
			fmt.Fprintf(os.Stdout, "FAIL  %s\n", relPath)
		}

		// Print errors
		for _, issue := range result.Issues {
			if issue.Field != "" {
				fmt.Fprintf(os.Stdout, "  ERROR: %s: %s\n", issue.Field, issue.Message)
			} else {
				fmt.Fprintf(os.Stdout, "  ERROR: %s\n", issue.Message)
			}
			if issue.Value != nil {
				fmt.Fprintf(os.Stdout, "         value: %v\n", issue.Value)
			}
		}

		// Print warnings
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stdout, "  WARN: %s\n", warning)
		}
	}
}
