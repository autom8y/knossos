package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// agentListOutput is the structured output for ari agent list.
type agentListOutput struct {
	Agents []agentListEntry `json:"agents"`
	Total  int              `json:"total"`
}

type agentListEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Model       string `json:"model"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

// Headers implements output.Tabular.
func (l agentListOutput) Headers() []string {
	return []string{"AGENT", "TYPE", "MODEL", "SOURCE", "DESCRIPTION"}
}

// Rows implements output.Tabular.
func (l agentListOutput) Rows() [][]string {
	rows := make([][]string, len(l.Agents))
	for i, a := range l.Agents {
		rows[i] = []string{a.Name, a.Type, a.Model, a.Source, a.Description}
	}
	return rows
}

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
		_ = printer.Print("No agents found")
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

	// Build structured output
	entries := make([]agentListEntry, len(agents))
	for i, a := range agents {
		name := a.Name
		if name == "" {
			relPath, err := filepath.Rel(resolver.ProjectRoot(), a.Path)
			if err != nil {
				relPath = filepath.Base(a.Path)
			} else {
				relPath = filepath.Base(relPath)
			}
			name = strings.TrimSuffix(relPath, ".md")
		}

		agentType := a.Type
		if agentType == "" {
			agentType = "-"
		}

		model := a.Model
		if model == "" {
			model = "-"
		}

		description := a.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		entries[i] = agentListEntry{
			Name:        name,
			Type:        agentType,
			Model:       model,
			Source:      a.Source,
			Description: description,
		}
	}

	return printer.Print(agentListOutput{
		Agents: entries,
		Total:  len(entries),
	})
}

func determineSource(agentPath string, resolver *paths.Resolver) string {
	projectRoot := resolver.ProjectRoot()
	ritesDir := resolver.RitesDir()

	// Check if it's a rite agent
	if strings.HasPrefix(agentPath, ritesDir) {
		relPath, _ := filepath.Rel(ritesDir, agentPath)
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) > 0 {
			return "rite:" + parts[0]
		}
	}

	// Check if it's a user agent
	userAgentsDir := filepath.Join(projectRoot, "agents")
	if strings.HasPrefix(agentPath, userAgentsDir) {
		return "user"
	}

	return "unknown"
}

