package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

type listOptions struct {
	riteName string
	all      bool
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agents",
		Long: `Lists agents with their metadata from frontmatter.

Examples:
  ari agent list                  # List all agents
  ari agent list --rite ecosystem # List agents in ecosystem rite`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to list agents from")
	cmd.Flags().BoolVar(&opts.all, "all", false, "List all agents (default)")

	return cmd
}

type agentInfo struct {
	Path        string
	Name        string
	Type        string
	Model       string
	Description string
	Source      string // "rite:ecosystem", "user", etc.
}

func runList(ctx *cmdContext, opts listOptions) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	// Collect agent paths
	var agentPaths []string
	var err error

	if opts.riteName != "" {
		ritePath := resolver.RiteDir(opts.riteName)
		agentsDir := filepath.Join(ritePath, "agents")
		agentPaths, err = collectAgentsInDir(agentsDir)
		if err != nil {
			printer.PrintError(err)
			return err
		}
	} else {
		agentPaths, err = collectAllAgents(resolver)
		if err != nil {
			printer.PrintError(err)
			return err
		}
	}

	if len(agentPaths) == 0 {
		printer.Print("No agents found")
		return nil
	}

	// Parse agent frontmatter
	agents := make([]agentInfo, 0, len(agentPaths))
	for _, agentPath := range agentPaths {
		content, err := os.ReadFile(agentPath)
		if err != nil {
			continue
		}

		fm, err := agentpkg.ParseAgentFrontmatter(content)
		if err != nil {
			continue
		}

		info := agentInfo{
			Path:        agentPath,
			Name:        fm.Name,
			Type:        fm.Type,
			Model:       fm.Model,
			Description: fm.Description,
			Source:      determineSource(agentPath, resolver),
		}

		agents = append(agents, info)
	}

	// Print agents
	printAgentList(agents, resolver)

	return nil
}

func determineSource(agentPath string, resolver *paths.Resolver) string {
	projectRoot := resolver.ProjectRoot()
	ritesDir := resolver.RitesDir()

	// Check if it's a rite agent
	if filepath.HasPrefix(agentPath, ritesDir) {
		relPath, _ := filepath.Rel(ritesDir, agentPath)
		parts := filepath.SplitList(relPath)
		if len(parts) > 0 {
			return "rite:" + filepath.Base(filepath.Dir(filepath.Dir(agentPath)))
		}
	}

	// Check if it's a user agent
	userAgentsDir := filepath.Join(projectRoot, "agents")
	if filepath.HasPrefix(agentPath, userAgentsDir) {
		return "user"
	}

	return "unknown"
}

func printAgentList(agents []agentInfo, resolver *paths.Resolver) {
	projectRoot := resolver.ProjectRoot()

	fmt.Fprintf(os.Stdout, "%-30s %-15s %-10s %-15s %s\n", "AGENT", "TYPE", "MODEL", "SOURCE", "DESCRIPTION")
	fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("-", 100))

	for _, agent := range agents {
		relPath, err := filepath.Rel(projectRoot, agent.Path)
		if err != nil {
			relPath = filepath.Base(agent.Path)
		} else {
			relPath = filepath.Base(relPath)
		}

		name := agent.Name
		if name == "" {
			name = strings.TrimSuffix(relPath, ".md")
		}

		agentType := agent.Type
		if agentType == "" {
			agentType = "-"
		}

		model := agent.Model
		if model == "" {
			model = "-"
		}

		description := agent.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		fmt.Fprintf(os.Stdout, "%-30s %-15s %-10s %-15s %s\n",
			name, agentType, model, agent.Source, description)
	}
}
