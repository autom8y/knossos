package agent

import (
	"testing"
)

func TestParseAgentFrontmatter_Minimal(t *testing.T) {
	content := []byte(`---
name: ecosystem-analyst
description: "Traces ecosystem issues to root causes"
tools: Bash, Glob, Grep, Read, Edit, Write, TodoWrite, Skill
model: opus
color: orange
---

# Ecosystem Analyst
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Name != "ecosystem-analyst" {
		t.Errorf("name = %q, want %q", fm.Name, "ecosystem-analyst")
	}
	if fm.Description != "Traces ecosystem issues to root causes" {
		t.Errorf("description = %q, want %q", fm.Description, "Traces ecosystem issues to root causes")
	}
	if len(fm.Tools) != 8 {
		t.Errorf("tools count = %d, want 8; tools = %v", len(fm.Tools), fm.Tools)
	}
	if fm.Model != "opus" {
		t.Errorf("model = %q, want %q", fm.Model, "opus")
	}
	if fm.Color != "orange" {
		t.Errorf("color = %q, want %q", fm.Color, "orange")
	}
}

func TestParseAgentFrontmatter_Enhanced(t *testing.T) {
	content := []byte(`---
name: context-architect
description: "Infrastructure designer who architects context solutions"
role: "Designs CEM/knossos schemas"
type: specialist
tools:
  - Bash
  - Glob
  - Grep
  - Read
  - Edit
  - Write
model: opus
color: cyan
upstream:
  - source: ecosystem-analyst
    artifact: gap-analysis
downstream:
  - agent: integration-engineer
    condition: "design complete"
    artifact: context-design
produces:
  - artifact: context-design
    format: markdown
contract:
  must_produce:
    - context-design
  must_not:
    - write implementation code
schema_version: "1.0"
---

# Context Architect
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Name != "context-architect" {
		t.Errorf("name = %q, want %q", fm.Name, "context-architect")
	}
	if fm.Type != "specialist" {
		t.Errorf("type = %q, want %q", fm.Type, "specialist")
	}
	if fm.Role != "Designs CEM/knossos schemas" {
		t.Errorf("role = %q, want %q", fm.Role, "Designs CEM/knossos schemas")
	}
	if len(fm.Tools) != 6 {
		t.Errorf("tools count = %d, want 6", len(fm.Tools))
	}

	// Upstream
	if len(fm.Upstream) != 1 {
		t.Fatalf("upstream count = %d, want 1", len(fm.Upstream))
	}
	if fm.Upstream[0].Source != "ecosystem-analyst" {
		t.Errorf("upstream[0].source = %q, want %q", fm.Upstream[0].Source, "ecosystem-analyst")
	}
	if fm.Upstream[0].Artifact != "gap-analysis" {
		t.Errorf("upstream[0].artifact = %q, want %q", fm.Upstream[0].Artifact, "gap-analysis")
	}

	// Downstream
	if len(fm.Downstream) != 1 {
		t.Fatalf("downstream count = %d, want 1", len(fm.Downstream))
	}
	if fm.Downstream[0].Agent != "integration-engineer" {
		t.Errorf("downstream[0].agent = %q, want %q", fm.Downstream[0].Agent, "integration-engineer")
	}

	// Produces
	if len(fm.Produces) != 1 {
		t.Fatalf("produces count = %d, want 1", len(fm.Produces))
	}
	if fm.Produces[0].Artifact != "context-design" {
		t.Errorf("produces[0].artifact = %q, want %q", fm.Produces[0].Artifact, "context-design")
	}

	// Contract
	if fm.Contract == nil {
		t.Fatal("contract is nil")
	}
	if len(fm.Contract.MustProduce) != 1 || fm.Contract.MustProduce[0] != "context-design" {
		t.Errorf("contract.must_produce = %v, want [context-design]", fm.Contract.MustProduce)
	}
	if len(fm.Contract.MustNot) != 1 || fm.Contract.MustNot[0] != "write implementation code" {
		t.Errorf("contract.must_not = %v, want [write implementation code]", fm.Contract.MustNot)
	}

	if fm.SchemaVersion != "1.0" {
		t.Errorf("schema_version = %q, want %q", fm.SchemaVersion, "1.0")
	}
}

func TestParseAgentFrontmatter_WithAliases(t *testing.T) {
	content := []byte(`---
name: moirai
description: "Session lifecycle agent"
tools: Read, Write, Edit, Glob, Grep, Bash, Skill
model: sonnet
color: indigo
aliases:
  - fates
  - state-mate
---

# Moirai
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.Aliases) != 2 {
		t.Fatalf("aliases count = %d, want 2", len(fm.Aliases))
	}
	if fm.Aliases[0] != "fates" || fm.Aliases[1] != "state-mate" {
		t.Errorf("aliases = %v, want [fates state-mate]", fm.Aliases)
	}
}

func TestFlexibleStringSlice_CommaSeparated(t *testing.T) {
	content := []byte(`---
name: test-agent
description: "Test agent"
tools: Bash, Glob, Grep, Read
---

# Test
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Bash", "Glob", "Grep", "Read"}
	if len(fm.Tools) != len(expected) {
		t.Fatalf("tools count = %d, want %d", len(fm.Tools), len(expected))
	}
	for i, tool := range fm.Tools {
		if tool != expected[i] {
			t.Errorf("tools[%d] = %q, want %q", i, tool, expected[i])
		}
	}
}

func TestFlexibleStringSlice_Array(t *testing.T) {
	content := []byte(`---
name: test-agent
description: "Test agent"
tools:
  - Bash
  - Glob
  - Grep
  - Read
---

# Test
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Bash", "Glob", "Grep", "Read"}
	if len(fm.Tools) != len(expected) {
		t.Fatalf("tools count = %d, want %d", len(fm.Tools), len(expected))
	}
	for i, tool := range fm.Tools {
		if tool != expected[i] {
			t.Errorf("tools[%d] = %q, want %q", i, tool, expected[i])
		}
	}
}

func TestMCPToolReferences(t *testing.T) {
	content := []byte(`---
name: mcp-agent
description: "Agent with MCP tools"
tools:
  - Read
  - mcp:github
  - mcp:github/create_issue
---

# MCP Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.Tools) != 3 {
		t.Fatalf("tools count = %d, want 3", len(fm.Tools))
	}
	if fm.Tools[1] != "mcp:github" {
		t.Errorf("tools[1] = %q, want %q", fm.Tools[1], "mcp:github")
	}
	if fm.Tools[2] != "mcp:github/create_issue" {
		t.Errorf("tools[2] = %q, want %q", fm.Tools[2], "mcp:github/create_issue")
	}
}

func TestMCPServers(t *testing.T) {
	fm := &AgentFrontmatter{
		Tools: FlexibleStringSlice{"Read", "mcp:github", "mcp:github/create_issue", "mcp:slack", "Bash"},
	}

	servers := fm.MCPServers()
	if len(servers) != 2 {
		t.Fatalf("MCPServers count = %d, want 2; got %v", len(servers), servers)
	}

	// Check both servers are present (order may vary)
	serverSet := make(map[string]bool)
	for _, s := range servers {
		serverSet[s] = true
	}
	if !serverSet["github"] {
		t.Error("expected github in MCPServers")
	}
	if !serverSet["slack"] {
		t.Error("expected slack in MCPServers")
	}
}

func TestMCPServers_NoMCP(t *testing.T) {
	fm := &AgentFrontmatter{
		Tools: FlexibleStringSlice{"Read", "Bash", "Glob"},
	}

	servers := fm.MCPServers()
	if len(servers) != 0 {
		t.Errorf("MCPServers = %v, want empty", servers)
	}
}

func TestValidate_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		fm      AgentFrontmatter
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing name",
			fm:      AgentFrontmatter{Description: "test"},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "missing description",
			fm:      AgentFrontmatter{Name: "test"},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "valid minimal",
			fm: AgentFrontmatter{
				Name:        "test-agent",
				Description: "A test agent",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fm.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" {
					if got := err.Error(); got != tt.errMsg {
						// Just check it contains the expected message
						if !containsStr(got, tt.errMsg) {
							t.Errorf("error = %q, want to contain %q", got, tt.errMsg)
						}
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidate_InvalidType(t *testing.T) {
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		Type:        "invalid-type",
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !containsStr(err.Error(), "invalid type") {
		t.Errorf("error = %q, want to contain 'invalid type'", err.Error())
	}
}

func TestValidate_InvalidModel(t *testing.T) {
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		Model:       "gpt-4",
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	if !containsStr(err.Error(), "invalid model") {
		t.Errorf("error = %q, want to contain 'invalid model'", err.Error())
	}
}

func TestValidate_InvalidToolReference(t *testing.T) {
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		Tools:       FlexibleStringSlice{"Read", "InvalidTool"},
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid tool")
	}
	if !containsStr(err.Error(), "unknown tool") {
		t.Errorf("error = %q, want to contain 'unknown tool'", err.Error())
	}
}

func TestValidateToolReference(t *testing.T) {
	tests := []struct {
		tool    string
		wantErr bool
	}{
		{"Read", false},
		{"Bash", false},
		{"Glob", false},
		{"Grep", false},
		{"Edit", false},
		{"Write", false},
		{"Task", false},
		{"TodoWrite", false},
		{"WebSearch", false},
		{"WebFetch", false},
		{"Skill", false},
		{"NotebookEdit", false},
		{"AskUserQuestion", false},
		{"mcp:github", false},
		{"mcp:github/create_issue", false},
		{"mcp:my-server", false},
		{"mcp:my-server/some_method", false},
		{"invalid", true},
		{"mcp:", true},       // missing server name
		{"mcp:UPPER", true},  // uppercase not allowed in MCP
		{"mcp:a/B", true},    // uppercase in method
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			err := validateToolReference(tt.tool)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for tool %q", tt.tool)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for tool %q: %v", tt.tool, err)
			}
		})
	}
}

func TestParseAgentFrontmatter_MissingDelimiter(t *testing.T) {
	content := []byte(`name: test
description: "no delimiters"
`)

	_, err := ParseAgentFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for missing delimiter")
	}
}

func TestParseAgentFrontmatter_MissingClosingDelimiter(t *testing.T) {
	content := []byte(`---
name: test
description: "no closing"
`)

	_, err := ParseAgentFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for missing closing delimiter")
	}
}

func TestParseAgentFrontmatter_MultilineDescription(t *testing.T) {
	content := []byte(`---
name: agent-designer
role: "Designs agent roles and contracts"
description: |
  The rite architecture specialist who designs agent roles.
  Invoke when creating a new rite.
tools: Bash, Glob, Grep, Read, Write, Task, TodoWrite
model: opus
color: purple
---

# Agent Designer
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Name != "agent-designer" {
		t.Errorf("name = %q, want %q", fm.Name, "agent-designer")
	}
	if fm.Description == "" {
		t.Error("description should not be empty for multiline YAML")
	}
	if len(fm.Tools) != 7 {
		t.Errorf("tools count = %d, want 7", len(fm.Tools))
	}
}

func TestParseAgentFrontmatter_NoToolsField(t *testing.T) {
	// Some agents like context-engineer lack a tools field
	content := []byte(`---
name: context-engineer
description: "Context architecture specialist"
model: opus
color: orange
---

# Context Engineer
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Name != "context-engineer" {
		t.Errorf("name = %q, want %q", fm.Name, "context-engineer")
	}
	if len(fm.Tools) != 0 {
		t.Errorf("tools count = %d, want 0 (no tools field)", len(fm.Tools))
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
