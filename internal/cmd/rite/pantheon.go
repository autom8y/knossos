package rite

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

func newPantheonCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pantheon",
		Short: "Display active agents for the current rite",
		Long: `Display the pantheon (set of agents) for the currently active rite.

Shows all agents available in the current rite along with their roles.
This is useful for understanding what specialists are available for delegation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPantheon(ctx)
		},
	}

	return cmd
}

// AgentFrontmatter represents the YAML frontmatter of an agent file.
type AgentFrontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       string   `yaml:"tools"`
	Model       string   `yaml:"model"`
	Color       string   `yaml:"color"`
	Aliases     []string `yaml:"aliases"`
}

func runPantheon(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get active rite
	activeRite := ctx.getActiveRite()
	if activeRite == "" {
		return errors.New(errors.CodeFileNotFound, "no active rite (use 'ari sync --rite=<name>' to activate)")
	}

	// Read agents from .claude/agents/
	agentsDir := filepath.Join(resolver.ChannelDir(paths.ClaudeChannel{}), "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, fmt.Sprintf("failed to read agents directory: %s", agentsDir), err)
	}

	var agents []output.PantheonAgent
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// Read agent file and parse frontmatter
		agentPath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(agentPath)
		if err != nil {
			continue
		}

		agent := output.PantheonAgent{
			File: entry.Name(),
		}

		fm, err := parseFrontmatter(content)
		if err != nil || fm == nil {
			agent.Name = entry.Name()[:len(entry.Name())-3] // Remove .md
		} else {
			agent.Name = fm.Name
			agent.Description = fm.Description
			agent.Model = fm.Model
		}

		agents = append(agents, agent)
	}

	result := output.PantheonOutput{
		Rite:   activeRite,
		Agents: agents,
		Count:  len(agents),
	}

	return printer.Print(result)
}

// parseFrontmatter extracts YAML frontmatter from markdown content.
func parseFrontmatter(content []byte) (*AgentFrontmatter, error) {
	// Find frontmatter delimiters
	str := string(content)
	if len(str) < 3 || str[:3] != "---" {
		return nil, nil
	}

	// Find closing ---
	end := -1
	for i := 3; i < len(str)-3; i++ {
		if str[i:i+4] == "\n---" {
			end = i + 1
			break
		}
	}

	if end == -1 {
		return nil, nil
	}

	// Parse YAML
	var fm AgentFrontmatter
	if err := yaml.Unmarshal([]byte(str[3:end]), &fm); err != nil {
		return nil, err
	}

	return &fm, nil
}
