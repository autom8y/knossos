package agent

import (
	"testing"
)

func TestParseAgentFrontmatter_Minimal(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	content := []byte(`---
name: context-architect
description: "Infrastructure designer who architects context solutions"
role: "Designs sync/knossos schemas"
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
	if fm.Role != "Designs sync/knossos schemas" {
		t.Errorf("role = %q, want %q", fm.Role, "Designs sync/knossos schemas")
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
	t.Parallel()
	content := []byte(`---
name: moirai
description: "Session lifecycle agent"
tools: Read, Write, Edit, Glob, Grep, Bash, Skill
model: sonnet
color: purple
aliases:
  - fates
  - moirai
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
	if fm.Aliases[0] != "fates" || fm.Aliases[1] != "moirai" {
		t.Errorf("aliases = %v, want [fates moirai]", fm.Aliases)
	}
}

func TestFlexibleStringSlice_CommaSeparated(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	fm := &AgentFrontmatter{
		Tools: FlexibleStringSlice{"Read", "Bash", "Glob"},
	}

	servers := fm.MCPServers()
	if len(servers) != 0 {
		t.Errorf("MCPServers = %v, want empty", servers)
	}
}

func TestValidate_RequiredFields(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
	content := []byte(`name: test
description: "no delimiters"
`)

	_, err := ParseAgentFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for missing delimiter")
	}
}

func TestParseAgentFrontmatter_MissingClosingDelimiter(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestParseAgentFrontmatter_WithCCNativeFields(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: orchestrator
description: "Coordinates ecosystem phases"
type: orchestrator
tools: Read
model: opus
color: purple
maxTurns: 3
skills:
  - ecosystem-ref
  - standards
disallowedTools: Bash, Write, Edit, Glob, Grep, Task
---

# Orchestrator
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.MaxTurns != 3 {
		t.Errorf("maxTurns = %d, want 3", fm.MaxTurns)
	}

	if len(fm.Skills) != 2 {
		t.Fatalf("skills count = %d, want 2", len(fm.Skills))
	}
	if fm.Skills[0] != "ecosystem-ref" {
		t.Errorf("skills[0] = %q, want %q", fm.Skills[0], "ecosystem-ref")
	}
	if fm.Skills[1] != "standards" {
		t.Errorf("skills[1] = %q, want %q", fm.Skills[1], "standards")
	}

	expectedDisallowed := []string{"Bash", "Write", "Edit", "Glob", "Grep", "Task"}
	if len(fm.DisallowedTools) != len(expectedDisallowed) {
		t.Fatalf("disallowedTools count = %d, want %d", len(fm.DisallowedTools), len(expectedDisallowed))
	}
	for i, tool := range fm.DisallowedTools {
		if tool != expectedDisallowed[i] {
			t.Errorf("disallowedTools[%d] = %q, want %q", i, tool, expectedDisallowed[i])
		}
	}
}

func TestParseAgentFrontmatter_MaxTurnsOnly(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: specialist
description: "Domain expert"
tools: Read, Bash
maxTurns: 25
---

# Specialist
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.MaxTurns != 25 {
		t.Errorf("maxTurns = %d, want 25", fm.MaxTurns)
	}
	if len(fm.Skills) != 0 {
		t.Errorf("skills should be empty, got %v", fm.Skills)
	}
	if len(fm.DisallowedTools) != 0 {
		t.Errorf("disallowedTools should be empty, got %v", fm.DisallowedTools)
	}
}

func TestValidate_MaxTurnsNegative(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		MaxTurns:    -1,
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("expected error for negative maxTurns")
	}
	if !containsStr(err.Error(), "maxTurns") {
		t.Errorf("error = %q, want to contain 'maxTurns'", err.Error())
	}
}

func TestValidate_MaxTurnsZero(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		MaxTurns:    0,
	}

	err := fm.Validate()
	if err != nil {
		t.Fatalf("unexpected error for maxTurns=0: %v", err)
	}
}

func TestValidate_MaxTurnsPositive(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		MaxTurns:    15,
	}

	err := fm.Validate()
	if err != nil {
		t.Fatalf("unexpected error for maxTurns=15: %v", err)
	}
}

func TestValidate_DisallowedToolsValid(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:            "test-agent",
		Description:     "A test agent",
		DisallowedTools: FlexibleStringSlice{"Bash", "Write", "Task"},
	}

	err := fm.Validate()
	if err != nil {
		t.Fatalf("unexpected error for valid disallowedTools: %v", err)
	}
}

func TestValidate_DisallowedToolsInvalid(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:            "test-agent",
		Description:     "A test agent",
		DisallowedTools: FlexibleStringSlice{"Read", "InvalidTool"},
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid disallowedTools")
	}
	if !containsStr(err.Error(), "disallowedTools") {
		t.Errorf("error = %q, want to contain 'disallowedTools'", err.Error())
	}
}

func TestValidate_ModelInherit(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		Model:       "inherit",
	}

	err := fm.Validate()
	if err != nil {
		t.Fatalf("unexpected error for model=inherit: %v", err)
	}
}

func TestParseAgentFrontmatter_ModelInherit(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: inherit-agent
description: "Agent that inherits parent model"
tools: Read, Bash
model: inherit
---

# Inherit Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Model != "inherit" {
		t.Errorf("model = %q, want %q", fm.Model, "inherit")
	}

	if err := fm.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestParseAgentFrontmatter_WithMemoryField(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: memory-agent
description: "Agent with memory enabled"
tools: Read, Bash
memory: true
---

# Memory Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = false, want true")
	}
	if fm.Memory.Scope() != "project" {
		t.Errorf("memory.Scope() = %q, want %q", fm.Memory.Scope(), "project")
	}
}

func TestParseAgentFrontmatter_WithPermissionMode(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: bypass-agent
description: "Agent with bypassPermissions"
tools: Read, Bash
permissionMode: bypassPermissions
---

# Bypass Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.PermissionMode != "bypassPermissions" {
		t.Errorf("permissionMode = %q, want %q", fm.PermissionMode, "bypassPermissions")
	}
}

func TestValidate_PermissionMode_Valid(t *testing.T) {
	t.Parallel()
	modes := []string{"default", "plan", "bypassPermissions"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			fm := AgentFrontmatter{
				Name:           "test-agent",
				Description:    "A test agent",
				PermissionMode: mode,
			}
			if err := fm.Validate(); err != nil {
				t.Fatalf("unexpected error for permissionMode %q: %v", mode, err)
			}
		})
	}
}

func TestValidate_PermissionMode_Invalid(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:           "test-agent",
		Description:    "A test agent",
		PermissionMode: "yolo",
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid permissionMode")
	}
	if !containsStr(err.Error(), "permissionMode") {
		t.Errorf("error = %q, want to contain 'permissionMode'", err.Error())
	}
}

func TestParseAgentFrontmatter_WithMcpServers(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: mcp-config-agent
description: "Agent with MCP server configs"
tools: Read
mcpServers:
  - name: github
    url: https://mcp.github.com
  - name: slack
    url: https://mcp.slack.com
---

# MCP Config Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.McpServers) != 2 {
		t.Fatalf("mcpServers count = %d, want 2", len(fm.McpServers))
	}
	if fm.McpServers[0].Name != "github" {
		t.Errorf("mcpServers[0].name = %q, want %q", fm.McpServers[0].Name, "github")
	}
	if fm.McpServers[0].URL != "https://mcp.github.com" {
		t.Errorf("mcpServers[0].url = %q, want %q", fm.McpServers[0].URL, "https://mcp.github.com")
	}
	if fm.McpServers[1].Name != "slack" {
		t.Errorf("mcpServers[1].name = %q, want %q", fm.McpServers[1].Name, "slack")
	}
}

func TestParseAgentFrontmatter_WithHooks(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: hooks-agent
description: "Agent with hooks configuration"
tools: Read
hooks:
  PreToolUse:
    - matcher: Bash
      action: deny
---

# Hooks Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Hooks == nil {
		t.Fatal("hooks should not be nil")
	}
	if _, ok := fm.Hooks["PreToolUse"]; !ok {
		t.Error("hooks should contain PreToolUse key")
	}
}

func TestParseAgentFrontmatter_AllNewCCFields(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: full-cc-agent
description: "Agent with all CC-native fields"
tools: Read, Bash
model: opus
maxTurns: 10
skills:
  - ecosystem-ref
disallowedTools: Task
memory: true
permissionMode: plan
mcpServers:
  - name: github
    url: https://mcp.github.com
hooks:
  PostToolUse:
    - matcher: Write
      action: log
---

# Full CC Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = false, want true")
	}
	if fm.Memory.Scope() != "project" {
		t.Errorf("memory.Scope() = %q, want %q", fm.Memory.Scope(), "project")
	}
	if fm.PermissionMode != "plan" {
		t.Errorf("permissionMode = %q, want %q", fm.PermissionMode, "plan")
	}
	if len(fm.McpServers) != 1 {
		t.Errorf("mcpServers count = %d, want 1", len(fm.McpServers))
	}
	if fm.Hooks == nil {
		t.Error("hooks should not be nil")
	}

	// Validate should pass
	if err := fm.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestParseAgentFrontmatter_MemoryFalse(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: no-memory-agent
description: "Agent with memory explicitly disabled"
tools: Read
memory: false
---

# No Memory Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = true, want false")
	}
	if fm.Memory.Scope() != "" {
		t.Errorf("memory.Scope() = %q, want %q", fm.Memory.Scope(), "")
	}
}

func TestParseAgentFrontmatter_MemoryStringProject(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: project-memory-agent
description: "Agent with explicit project memory scope"
tools: Read
memory: "project"
---

# Project Memory Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = false, want true")
	}
	if fm.Memory.Scope() != "project" {
		t.Errorf("memory.Scope() = %q, want %q", fm.Memory.Scope(), "project")
	}
}

func TestParseAgentFrontmatter_MemoryStringUser(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: user-memory-agent
description: "Agent with user memory scope"
tools: Read
memory: "user"
---

# User Memory Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = false, want true")
	}
	if fm.Memory.Scope() != "user" {
		t.Errorf("memory.Scope() = %q, want %q", fm.Memory.Scope(), "user")
	}
}

func TestParseAgentFrontmatter_MemoryStringLocal(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: local-memory-agent
description: "Agent with local memory scope"
tools: Read
memory: "local"
---

# Local Memory Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = false, want true")
	}
	if fm.Memory.Scope() != "local" {
		t.Errorf("memory.Scope() = %q, want %q", fm.Memory.Scope(), "local")
	}
}

func TestValidate_MemoryScope_Invalid(t *testing.T) {
	t.Parallel()
	fm := AgentFrontmatter{
		Name:        "test-agent",
		Description: "A test agent",
		Memory:      MemoryField("invalid"),
	}

	err := fm.Validate()
	if err == nil {
		t.Fatal("Validate() expected error for invalid memory scope, got nil")
	}
	if !containsStr(err.Error(), "invalid memory scope") {
		t.Errorf("Validate() error = %q, want to contain %q", err.Error(), "invalid memory scope")
	}
}

func TestParseAgentFrontmatter_MemoryAbsent(t *testing.T) {
	t.Parallel()
	content := []byte(`---
name: no-memory-field-agent
description: "Agent without a memory field"
tools: Read
---

# No Memory Field Agent
`)

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Memory.IsEnabled() {
		t.Error("memory.IsEnabled() = true, want false (absent field should be disabled)")
	}
	if fm.Memory.Scope() != "" {
		t.Errorf("memory.Scope() = %q, want %q (absent field should be empty)", fm.Memory.Scope(), "")
	}
}

func TestValidate_MemoryScope_Valid(t *testing.T) {
	t.Parallel()
	scopes := []string{"user", "project", "local"}
	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			t.Parallel()
			fm := AgentFrontmatter{
				Name:        "test-agent",
				Description: "A test agent",
				Memory:      MemoryField(scope),
			}
			if err := fm.Validate(); err != nil {
				t.Errorf("Validate() error = %v, want nil for scope %q", err, scope)
			}
		})
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
