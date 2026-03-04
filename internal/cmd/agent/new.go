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
)

type newOptions struct {
	archetype   string
	riteName    string
	name        string
	description string
}

func newNewCmd(ctx *cmdContext) *cobra.Command {
	var opts newOptions

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Scaffold a new agent from an archetype",
		Long: `Creates a new agent file from an archetype template.

Archetypes provide default structure, sections, and platform content.
Author sections are marked with TODO comments for you to fill in.

Available archetypes: ` + strings.Join(agentpkg.ListArchetypes(), ", ") + `

Examples:
  ari agent new --archetype specialist --rite rnd --name technology-scout
  ari agent new --archetype reviewer --rite security --name code-reviewer --description "Reviews code for security issues"
  ari agent new --archetype orchestrator --rite ecosystem --name coordinator`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.archetype, "archetype", "a", "", "Archetype to use ("+strings.Join(agentpkg.ListArchetypes(), ", ")+")")
	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to create the agent in")
	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Agent name (kebab-case)")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Agent description (optional)")

	_ = cmd.MarkFlagRequired("archetype")
	_ = cmd.MarkFlagRequired("rite")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func runNew(ctx *cmdContext, opts newOptions) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	// Validate archetype
	archetype, err := agentpkg.GetArchetype(opts.archetype)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Validate rite exists
	riteDir := resolver.RiteDir(opts.riteName)
	if _, statErr := os.Stat(riteDir); os.IsNotExist(statErr) {
		err := errors.NewWithDetails(errors.CodeRiteNotFound,
			fmt.Sprintf("rite %q not found at %s", opts.riteName, riteDir),
			map[string]any{"rite": opts.riteName, "path": riteDir})
		printer.PrintError(err)
		return err
	}

	// Ensure agents directory exists
	agentsDir := filepath.Join(riteDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		wrappedErr := errors.Wrap(errors.CodePermissionDenied,
			fmt.Sprintf("failed to create agents directory: %s", agentsDir), err)
		printer.PrintError(wrappedErr)
		return wrappedErr
	}

	// Check target file does not already exist
	targetPath := filepath.Join(agentsDir, opts.name+".md")
	if _, statErr := os.Stat(targetPath); statErr == nil {
		relPath, _ := filepath.Rel(resolver.ProjectRoot(), targetPath)
		err := errors.NewWithDetails(errors.CodeSessionExists,
			fmt.Sprintf("agent file already exists: %s", relPath),
			map[string]any{"path": targetPath})
		printer.PrintError(err)
		return err
	}

	// Scaffold the agent
	content, err := agentpkg.ScaffoldAgent(archetype, opts.name, opts.riteName, opts.description)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Write the file
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		wrappedErr := errors.Wrap(errors.CodePermissionDenied,
			fmt.Sprintf("failed to write agent file: %s", targetPath), err)
		printer.PrintError(wrappedErr)
		return wrappedErr
	}

	// Print success message
	relPath, relErr := filepath.Rel(resolver.ProjectRoot(), targetPath)
	if relErr != nil {
		relPath = targetPath
	}
	printer.PrintLine(fmt.Sprintf("Created %s -- fill in author sections marked with TODO", relPath))

	return nil
}
