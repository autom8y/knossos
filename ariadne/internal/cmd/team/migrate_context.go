package team

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/team"
)

type migrateContextOptions struct {
	dryRun bool
	force  bool
}

func newMigrateContextCmd(ctx *cmdContext) *cobra.Command {
	var opts migrateContextOptions

	cmd := &cobra.Command{
		Use:   "migrate-context <team>",
		Short: "Migrate team context from bash to YAML",
		Long: `Generates a context.yaml file from an existing context-injection.sh script.

This command:
1. Sources the team's context-injection.sh script in a subshell
2. Captures the output of inject_team_context()
3. Parses the markdown table output into key-value pairs
4. Generates a context.yaml file

Use --dry-run to preview the generated YAML without writing.

Examples:
  ari team migrate-context 10x-dev-pack           # Generate context.yaml
  ari team migrate-context 10x-dev-pack --dry-run # Preview only
  ari team migrate-context ecosystem-pack --force # Overwrite existing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateContext(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.dryRun, "dry-run", "n", false, "Preview generated YAML without writing")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Overwrite existing context.yaml")

	return cmd
}

func runMigrateContext(ctx *cmdContext, teamName string, opts migrateContextOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()
	discovery := ctx.getDiscovery()

	// Verify team exists
	teamInfo, err := discovery.Get(teamName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Check for existing context.yaml
	loader := team.NewContextLoader(resolver)
	if loader.HasContextFile(teamName) && !opts.force {
		err := errors.New(errors.CodeLifecycleViolation,
			fmt.Sprintf("context.yaml already exists for %s. Use --force to overwrite.", teamName))
		printer.PrintError(err)
		return err
	}

	// Find context-injection.sh
	scriptPath := filepath.Join(teamInfo.Path, "context-injection.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		err := errors.New(errors.CodeFileNotFound,
			fmt.Sprintf("No context-injection.sh found at %s", scriptPath))
		printer.PrintError(err)
		return err
	}

	// Execute the script and capture output
	output, err := executeContextScript(scriptPath)
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "Failed to execute context script", err))
		return err
	}

	// Parse the markdown table output
	rows, err := parseMarkdownTable(output)
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeParseError, "Failed to parse script output", err))
		return err
	}

	// Generate TeamContext
	teamCtx := team.NewTeamContext(teamName)
	teamCtx.Description = fmt.Sprintf("Migrated from context-injection.sh")

	// Add workflow info if available
	if workflow, err := team.LoadWorkflow(filepath.Join(teamInfo.Path, "workflow.yaml")); err == nil {
		teamCtx.Description = workflow.Description
	}

	// Add orchestrator info if available
	orchestratorPath := filepath.Join(teamInfo.Path, "orchestrator.yaml")
	if data, err := os.ReadFile(orchestratorPath); err == nil {
		// Quick parse for domain
		if domain := extractYAMLField(data, "domain"); domain != "" {
			teamCtx.Domain = domain
		}
	}

	// Add parsed rows
	for _, row := range rows {
		teamCtx.AddRow(row.Key, row.Value)
	}

	// Add migration metadata
	teamCtx.Metadata["migrated_from"] = "context-injection.sh"
	teamCtx.Metadata["migration_date"] = "auto-generated"

	// Output or save
	if opts.dryRun {
		printer.PrintLine("# Generated context.yaml (dry-run)")
		printer.PrintLine(fmt.Sprintf("# For team: %s", teamName))
		printer.PrintLine("")
		return printer.Print(teamCtx)
	}

	// Save the context
	if err := loader.SaveContext(teamCtx); err != nil {
		printer.PrintError(err)
		return err
	}

	printer.PrintLine(fmt.Sprintf("Generated context.yaml for %s", teamName))
	printer.PrintLine(fmt.Sprintf("Path: %s", loader.GetContextPath(teamName)))
	printer.PrintLine(fmt.Sprintf("Rows: %d", len(rows)))

	return nil
}

// executeContextScript runs the context-injection.sh script and captures output.
func executeContextScript(scriptPath string) (string, error) {
	// Build the bash command to source and execute
	bashCmd := fmt.Sprintf(`
source "%s" 2>/dev/null || exit 1
if declare -f inject_team_context >/dev/null 2>&1; then
    inject_team_context
else
    echo "ERROR: inject_team_context function not found" >&2
    exit 1
fi
`, scriptPath)

	cmd := exec.Command("bash", "-c", bashCmd)

	// Set environment
	cmd.Env = append(os.Environ(),
		"CLAUDE_PROJECT_DIR=.",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%s: %s", err, stderr.String())
		}
		return "", err
	}

	return stdout.String(), nil
}

// parseMarkdownTable parses a markdown table and extracts key-value pairs.
// Expected format:
// | | |
// |---|---|
// | **Key** | Value |
func parseMarkdownTable(input string) ([]team.ContextRow, error) {
	var rows []team.ContextRow

	// Regex to match table rows with bold keys
	// Format: | **Key** | Value |
	rowPattern := regexp.MustCompile(`\|\s*\*\*([^*]+)\*\*\s*\|\s*([^|]+)\s*\|`)

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip header rows
		if strings.Contains(line, "---|") || line == "| | |" {
			continue
		}

		matches := rowPattern.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			if key != "" {
				rows = append(rows, team.ContextRow{
					Key:   key,
					Value: value,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

// extractYAMLField does a simple extraction of a YAML field value.
// This is a simple implementation that doesn't require a full YAML parser.
func extractYAMLField(data []byte, field string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*%s:\s*(.+)$`, field))
	matches := pattern.FindSubmatch(data)
	if len(matches) >= 2 {
		return strings.TrimSpace(string(matches[1]))
	}
	return ""
}
