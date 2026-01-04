package team

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ClaudeMDUpdater handles CLAUDE.md satellite section updates.
type ClaudeMDUpdater struct {
	path string
}

// NewClaudeMDUpdater creates a new updater for the given CLAUDE.md path.
func NewClaudeMDUpdater(path string) *ClaudeMDUpdater {
	return &ClaudeMDUpdater{path: path}
}

// UpdateForTeam regenerates satellite sections for a team.
func (u *ClaudeMDUpdater) UpdateForTeam(team *Team) error {
	content, err := os.ReadFile(u.path)
	if err != nil {
		return fmt.Errorf("reading CLAUDE.md: %w", err)
	}

	// Load workflow for agent info
	workflowPath := filepath.Join(team.Path, "workflow.yaml")
	workflow, err := LoadWorkflow(workflowPath)
	if err != nil {
		return fmt.Errorf("loading workflow: %w", err)
	}

	// Read agent descriptions from files
	agentInfos := u.loadAgentInfos(team, workflow)

	// Update Quick Start section
	content = u.updateQuickStartSection(content, team, agentInfos)

	// Update Agent Configurations section
	content = u.updateAgentConfigsSection(content, team, agentInfos)

	return os.WriteFile(u.path, content, 0644)
}

// agentFileInfo holds info extracted from agent .md files.
type agentFileInfo struct {
	Name        string
	File        string
	Role        string
	Produces    string
	Description string
}

// loadAgentInfos reads agent metadata from files.
func (u *ClaudeMDUpdater) loadAgentInfos(team *Team, workflow *Workflow) []agentFileInfo {
	// Start with workflow info
	workflowInfos := workflow.GetAgentInfo()

	infos := make([]agentFileInfo, 0, len(workflowInfos))
	agentsDir := filepath.Join(team.Path, "agents")

	for _, wi := range workflowInfos {
		info := agentFileInfo{
			Name:     wi.Name,
			File:     wi.File,
			Role:     wi.Role,
			Produces: wi.Produces,
		}

		// Try to read first line of agent file for description
		agentPath := filepath.Join(agentsDir, wi.File)
		if desc := u.extractAgentDescription(agentPath); desc != "" {
			info.Description = desc
		} else {
			info.Description = info.Role
		}

		infos = append(infos, info)
	}

	return infos
}

// extractAgentDescription reads a brief description from an agent file.
func (u *ClaudeMDUpdater) extractAgentDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDescription := false
	var description strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header lines
		if strings.HasPrefix(line, "#") {
			inDescription = true
			continue
		}

		// Look for first non-empty paragraph after header
		if inDescription && strings.TrimSpace(line) != "" {
			// Get first sentence (up to 80 chars)
			text := strings.TrimSpace(line)
			if len(text) > 80 {
				// Find sentence boundary
				if idx := strings.Index(text[:80], ". "); idx > 0 {
					text = text[:idx]
				} else if idx := strings.Index(text[:80], "."); idx > 0 {
					text = text[:idx]
				} else {
					text = text[:77] + "..."
				}
			}
			description.WriteString(text)
			break
		}
	}

	return description.String()
}

// updateQuickStartSection updates the Quick Start table.
func (u *ClaudeMDUpdater) updateQuickStartSection(content []byte, team *Team, agents []agentFileInfo) []byte {
	lines := strings.Split(string(content), "\n")
	var result []string

	inSection := false
	skipToNextSection := false

	for i, line := range lines {
		// Detect section start
		if strings.HasPrefix(line, "## Quick Start") {
			inSection = true
			result = append(result, line)
			continue
		}

		// Detect section end (next ## heading or Agent Routing)
		if inSection && (strings.HasPrefix(line, "## Agent Routing") ||
			(strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "## Quick Start"))) {
			inSection = false
			skipToNextSection = false
			result = append(result, line)
			continue
		}

		// In Quick Start section, replace content after header
		if inSection && !skipToNextSection {
			// Skip blank lines right after header
			if strings.TrimSpace(line) == "" && len(result) > 0 {
				prevLine := result[len(result)-1]
				if strings.HasPrefix(prevLine, "## Quick Start") {
					result = append(result, "")
					continue
				}
			}

			// Look for PRESERVE comment
			if strings.Contains(line, "PRESERVE") {
				result = append(result, line)
				// Generate new content
				result = append(result, u.generateQuickStartContent(team, agents)...)
				skipToNextSection = true
				// Skip until next section
				continue
			}

			// If we see the old table or content, skip it
			if skipToNextSection {
				continue
			}

			// Check if this is the start of team info line
			if strings.HasPrefix(line, "This project uses") {
				// Start of generated content - skip until we find the table or blank lines
				skipToNextSection = true
				// Insert new content
				result = append(result, u.generateQuickStartContent(team, agents)...)
				continue
			}
		}

		// Skip old content if in skip mode
		if skipToNextSection {
			// Only add back when we hit the next section
			if strings.HasPrefix(line, "**New here?") || strings.HasPrefix(line, "## ") {
				skipToNextSection = false
				if !strings.HasPrefix(line, "## ") {
					result = append(result, "")
				}
			}
			// Check if this is page break markers
			if strings.TrimSpace(line) == "" {
				count := 0
				for j := i; j < len(lines) && j < i+20; j++ {
					if strings.TrimSpace(lines[j]) == "" {
						count++
					} else {
						break
					}
				}
				if count >= 5 {
					// This is likely the blank line padding before next section
					skipToNextSection = false
				}
			}
		}

		if !skipToNextSection || strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "**New here?") {
			result = append(result, line)
		}
	}

	return []byte(strings.Join(result, "\n"))
}

// generateQuickStartContent creates the Quick Start table markdown.
func (u *ClaudeMDUpdater) generateQuickStartContent(team *Team, agents []agentFileInfo) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf("This project uses a %d-agent workflow (%s):", team.AgentCount, team.Name))
	lines = append(lines, "")
	lines = append(lines, "| Agent | Role | Produces |")
	lines = append(lines, "| ----- | ---- | -------- |")

	for _, agent := range agents {
		produces := agent.Produces
		if produces == "" {
			produces = "-"
		}
		// Capitalize first letter of produces
		if len(produces) > 0 {
			produces = strings.ToUpper(string(produces[0])) + produces[1:]
		}
		lines = append(lines, fmt.Sprintf("| **%s** | %s | %s |", agent.Name, agent.Role, produces))
	}

	lines = append(lines, "")

	return lines
}

// updateAgentConfigsSection updates the Agent Configurations section.
func (u *ClaudeMDUpdater) updateAgentConfigsSection(content []byte, team *Team, agents []agentFileInfo) []byte {
	lines := strings.Split(string(content), "\n")
	var result []string

	inSection := false
	skipToNextSection := false

	for _, line := range lines {
		// Detect section start
		if strings.HasPrefix(line, "## Agent Configurations") {
			inSection = true
			result = append(result, line)
			continue
		}

		// Detect section end
		if inSection && strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "## Agent Configurations") {
			inSection = false
			skipToNextSection = false
			result = append(result, line)
			continue
		}

		// In section, replace content
		if inSection && !skipToNextSection {
			// Keep header and intro line
			if strings.HasPrefix(line, "Full agent prompts") {
				result = append(result, "")
				result = append(result, line)
				result = append(result, "")
				// Generate new content
				for _, agent := range agents {
					desc := agent.Description
					if len(desc) > 60 {
						// Truncate at word boundary
						if idx := strings.LastIndex(desc[:60], " "); idx > 0 {
							desc = desc[:idx]
						} else {
							desc = desc[:57] + "..."
						}
					}
					result = append(result, fmt.Sprintf("- `%s` - %s", agent.File, desc))
				}
				skipToNextSection = true
				continue
			}
		}

		if !skipToNextSection {
			result = append(result, line)
		}
	}

	return []byte(strings.Join(result, "\n"))
}

// IsSynced checks if CLAUDE.md satellites are in sync with the active team.
func (u *ClaudeMDUpdater) IsSynced(teamName string) bool {
	content, err := os.ReadFile(u.path)
	if err != nil {
		return false
	}

	// Simple check: team name appears in Quick Start section
	return strings.Contains(string(content), teamName)
}
