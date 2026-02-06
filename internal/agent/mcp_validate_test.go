package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAgentMCPReferences_NoDeclaredServers(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github", "Read", "Write"},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{})

	// Should warn about github server not being declared
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "github")
	assert.Contains(t, warnings[0], "not declared")
}

func TestValidateAgentMCPReferences_AllDeclared(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github", "mcp:terraform", "Read"},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{"github", "terraform"})

	// No warnings - all servers are declared
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPReferences_MixedDeclared(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github", "mcp:custom-server", "Read"},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{"github"})

	// Should warn about custom-server only
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "custom-server")
	assert.NotContains(t, warnings[0], "github")
}

func TestValidateAgentMCPReferences_NoMCPTools(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"Read", "Write", "Edit"},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{"github"})

	// No MCP tools, no warnings
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPReferences_NilAgent(t *testing.T) {
	warnings := ValidateAgentMCPReferences(nil, []string{"github"})
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPReferences_EmptyTools(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{"github"})
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPReferences_WithMethod(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github/create_issue", "mcp:github/create_pr"},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{"github"})

	// Both tools reference same server, should have no warnings
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPReferences_MultipleServersNotDeclared(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github", "mcp:terraform", "mcp:custom"},
	}

	warnings := ValidateAgentMCPReferences(agent, []string{"github"})

	// Should warn about terraform and custom
	assert.Len(t, warnings, 2)
}

func TestValidateAgentMCPToolReferences_DetailedWarnings(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github/create_issue", "mcp:custom-server", "Read"},
	}

	warnings := ValidateAgentMCPToolReferences(agent, []string{"github"})

	// Should have one warning for custom-server tool
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "custom-server")
	assert.Contains(t, warnings[0], "mcp:custom-server")
	assert.Contains(t, warnings[0], "not declared")
}

func TestValidateAgentMCPToolReferences_NoWarningsForDeclared(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github/create_issue", "mcp:terraform", "Read"},
	}

	warnings := ValidateAgentMCPToolReferences(agent, []string{"github", "terraform"})

	// No warnings
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPToolReferences_NilAgent(t *testing.T) {
	warnings := ValidateAgentMCPToolReferences(nil, []string{"github"})
	assert.Empty(t, warnings)
}

func TestValidateAgentMCPToolReferences_MultipleToolsSameServer(t *testing.T) {
	agent := &AgentFrontmatter{
		Name:        "test-agent",
		Description: "test",
		Tools:       []string{"mcp:github/create_issue", "mcp:github/create_pr", "mcp:undeclared"},
	}

	warnings := ValidateAgentMCPToolReferences(agent, []string{"github"})

	// Should only warn about undeclared server
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "undeclared")
}
