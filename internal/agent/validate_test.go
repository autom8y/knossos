package agent

import (
	"testing"
)

func newTestValidator(t *testing.T) *AgentValidator {
	t.Helper()
	av, err := NewAgentValidator()
	if err != nil {
		t.Fatalf("failed to create agent validator: %v", err)
	}
	return av
}

func TestArchetypeDefaults_IncludeCCNativeFields(t *testing.T) {
	tests := []struct {
		name                string
		expectedMaxTurns    int
		expectedDisallowed  []string
	}{
		{
			name:               "orchestrator",
			expectedMaxTurns:   3,
			expectedDisallowed: []string{"Bash", "Write", "Edit", "Glob", "Grep", "Task"},
		},
		{
			name:               "specialist",
			expectedMaxTurns:   25,
			expectedDisallowed: nil,
		},
		{
			name:               "reviewer",
			expectedMaxTurns:   15,
			expectedDisallowed: []string{"Task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arch, err := GetArchetype(tt.name)
			if err != nil {
				t.Fatalf("failed to get archetype %q: %v", tt.name, err)
			}

			if arch.Defaults.MaxTurns != tt.expectedMaxTurns {
				t.Errorf("archetype %q: MaxTurns = %d, want %d",
					tt.name, arch.Defaults.MaxTurns, tt.expectedMaxTurns)
			}

			if tt.expectedDisallowed == nil {
				if arch.Defaults.DisallowedTools != nil {
					t.Errorf("archetype %q: DisallowedTools = %v, want nil",
						tt.name, arch.Defaults.DisallowedTools)
				}
			} else {
				if len(arch.Defaults.DisallowedTools) != len(tt.expectedDisallowed) {
					t.Errorf("archetype %q: DisallowedTools count = %d, want %d",
						tt.name, len(arch.Defaults.DisallowedTools), len(tt.expectedDisallowed))
				}
				for i, tool := range arch.Defaults.DisallowedTools {
					if i >= len(tt.expectedDisallowed) {
						break
					}
					if tool != tt.expectedDisallowed[i] {
						t.Errorf("archetype %q: DisallowedTools[%d] = %q, want %q",
							tt.name, i, tool, tt.expectedDisallowed[i])
					}
				}
			}
		})
	}
}

func TestValidateAgentFrontmatter_MinimalValid_WarnMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: ecosystem-analyst
description: "Traces ecosystem issues to root causes and produces gap analysis"
tools: Bash, Glob, Grep, Read, Edit, Write, TodoWrite, Skill
model: opus
color: orange
---

# Ecosystem Analyst
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}
	if result.Frontmatter == nil {
		t.Fatal("frontmatter should not be nil")
	}
	if result.Frontmatter.Name != "ecosystem-analyst" {
		t.Errorf("name = %q, want %q", result.Frontmatter.Name, "ecosystem-analyst")
	}
}

func TestValidateAgentFrontmatter_EnhancedValid_StrictMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: context-architect
description: "Infrastructure designer who architects context solutions and ecosystem patterns"
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

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid in strict mode, got issues: %v", result.Issues)
		for _, w := range result.Warnings {
			t.Logf("warning: %s", w)
		}
	}
}

func TestValidateAgentFrontmatter_MissingName(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
description: "An agent without a name"
tools: Read
---

# No Name
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result for missing name")
	}

	foundNameIssue := false
	for _, issue := range result.Issues {
		if issue.Field == "name" || containsStr(issue.Message, "name") {
			foundNameIssue = true
			break
		}
	}
	if !foundNameIssue {
		t.Errorf("expected issue about missing name, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_MissingDescription(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: test-agent
tools: Read
---

# Test
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result for missing description")
	}

	foundDescIssue := false
	for _, issue := range result.Issues {
		if issue.Field == "description" || containsStr(issue.Message, "description") {
			foundDescIssue = true
			break
		}
	}
	if !foundDescIssue {
		t.Errorf("expected issue about missing description, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_MissingTools_WarnMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: context-engineer
description: "Context architecture specialist who optimizes how Claude is leveraged"
model: opus
color: orange
---

# Context Engineer
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid in WARN mode (tools missing is just a warning)
	if !result.Valid {
		t.Errorf("expected valid in warn mode with missing tools, got issues: %v", result.Issues)
	}

	// Should have a warning about missing tools
	foundToolsWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "tools") {
			foundToolsWarning = true
			break
		}
	}
	if !foundToolsWarning {
		t.Errorf("expected warning about missing tools, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_MissingTools_StrictMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: context-engineer
description: "Context architecture specialist who optimizes how Claude is leveraged"
type: specialist
model: opus
color: orange
---

# Context Engineer
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be invalid in STRICT mode (tools required)
	if result.Valid {
		t.Error("expected invalid in strict mode with missing tools")
	}

	foundToolsIssue := false
	for _, issue := range result.Issues {
		if issue.Field == "tools" || containsStr(issue.Message, "tools") {
			foundToolsIssue = true
			break
		}
	}
	if !foundToolsIssue {
		t.Errorf("expected issue about missing tools, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_InvalidToolReference(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: bad-tools
description: "Agent with invalid tool references"
tools:
  - Read
  - FakeTool
---

# Bad Tools
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result for unknown tool")
	}

	foundToolIssue := false
	for _, issue := range result.Issues {
		if containsStr(issue.Message, "unknown tool") || containsStr(issue.Message, "FakeTool") {
			foundToolIssue = true
			break
		}
	}
	if !foundToolIssue {
		t.Errorf("expected issue about unknown tool, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_MissingType_StrictMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: no-type-agent
description: "Agent without type field for strict mode"
tools: Read, Bash
model: opus
---

# No Type Agent
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid in strict mode without type field")
	}

	foundTypeIssue := false
	for _, issue := range result.Issues {
		if containsStr(issue.Message, "type") {
			foundTypeIssue = true
			break
		}
	}
	if !foundTypeIssue {
		t.Errorf("expected issue about missing type, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_OrchestratorWarnings(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: orchestrator
description: "Coordinates ecosystem phases for CEM/knossos infrastructure work"
type: orchestrator
tools: Read, Bash, Write
model: opus
color: purple
---

# Orchestrator
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce warning about non-Read tools for orchestrator
	foundToolWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "orchestrator") && containsStr(w, "Read") {
			foundToolWarning = true
			break
		}
	}
	if !foundToolWarning {
		t.Errorf("expected warning about orchestrator non-Read tools, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_OrchestratorReadOnly(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: orchestrator
description: "Coordinates ecosystem phases for CEM/knossos infrastructure work"
type: orchestrator
tools: Read
model: opus
color: purple
---

# Orchestrator
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}

	// Should NOT have the non-Read tools warning
	for _, w := range result.Warnings {
		if containsStr(w, "orchestrator") && containsStr(w, "Read") {
			t.Errorf("unexpected warning for Read-only orchestrator: %s", w)
		}
	}
}

func TestValidateAgentFrontmatter_ReviewerMustNot_StrictMode(t *testing.T) {
	av := newTestValidator(t)

	// Reviewer without contract.must_not in strict mode
	content := []byte(`---
name: security-reviewer
description: "Final security gate before merge - reviews PRs for security issues"
type: reviewer
tools: Bash, Glob, Grep, Read, Edit, Write
model: opus
color: red
---

# Security Reviewer
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid for reviewer without contract.must_not in strict mode")
	}

	foundContractIssue := false
	for _, issue := range result.Issues {
		if containsStr(issue.Message, "must_not") || containsStr(issue.Field, "contract") {
			foundContractIssue = true
			break
		}
	}
	if !foundContractIssue {
		t.Errorf("expected issue about reviewer must_not, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_ReviewerWithContract(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: security-reviewer
description: "Final security gate before merge - reviews PRs for security issues"
type: reviewer
tools: Bash, Glob, Grep, Read, Edit, Write
model: opus
color: red
contract:
  must_not:
    - approve code without security review
    - skip vulnerability scanning
---

# Security Reviewer
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid for reviewer with contract.must_not, got issues: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_ReviewerWarning_WarnMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: security-reviewer
description: "Final security gate before merge - reviews PRs for security issues"
type: reviewer
tools: Bash, Glob, Grep, Read, Edit, Write
model: opus
color: red
---

# Security Reviewer
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In WARN mode, should be valid but with warning
	if !result.Valid {
		t.Errorf("expected valid in warn mode, got issues: %v", result.Issues)
	}

	foundWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "must_not") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("expected warning about reviewer must_not, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_MCPToolValidation(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: mcp-agent
description: "Agent with MCP tool references for testing validation"
tools:
  - Read
  - mcp:github
  - mcp:github/create_issue
---

# MCP Agent
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid for MCP tools, got issues: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_InvalidMCPFormat(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: bad-mcp
description: "Agent with badly formatted MCP tool reference"
tools:
  - Read
  - mcp:INVALID
---

# Bad MCP
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid for bad MCP format")
	}
}

func TestValidateAgentFrontmatter_ParseError(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`no frontmatter here`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid for missing frontmatter")
	}

	if len(result.Issues) == 0 {
		t.Error("expected issues for parse error")
	}
}

func TestValidateToolReferences(t *testing.T) {
	fm := &AgentFrontmatter{
		Tools: FlexibleStringSlice{"Read", "Bash", "mcp:github", "mcp:slack/post_message"},
	}

	issues, warnings := ValidateToolReferences(fm)

	if len(issues) != 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}

	// Should have warnings about MCP tools
	if len(warnings) != 2 {
		t.Errorf("expected 2 MCP warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateToolReferences_InvalidTool(t *testing.T) {
	fm := &AgentFrontmatter{
		Tools: FlexibleStringSlice{"Read", "NotARealTool"},
	}

	issues, _ := ValidateToolReferences(fm)

	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestValidateAgentFrontmatter_OrchestratorNonOpusModel(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: orchestrator
description: "Coordinates phases with wrong model"
type: orchestrator
tools: Read
model: sonnet
color: purple
---

# Orchestrator
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce warning about non-opus model
	foundModelWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "opus") && containsStr(w, "orchestrator") {
			foundModelWarning = true
			break
		}
	}
	if !foundModelWarning {
		t.Errorf("expected warning about orchestrator model, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_CCNativeFields_Valid(t *testing.T) {
	av := newTestValidator(t)

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

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}

	if result.Frontmatter.MaxTurns != 3 {
		t.Errorf("maxTurns = %d, want 3", result.Frontmatter.MaxTurns)
	}
	if len(result.Frontmatter.Skills) != 2 {
		t.Errorf("skills count = %d, want 2", len(result.Frontmatter.Skills))
	}
	if len(result.Frontmatter.DisallowedTools) != 6 {
		t.Errorf("disallowedTools count = %d, want 6", len(result.Frontmatter.DisallowedTools))
	}
}

func TestValidateAgentFrontmatter_MaxTurnsNegative(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: bad-agent
description: "Agent with negative maxTurns"
tools: Read
maxTurns: -5
---

# Bad Agent
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid for negative maxTurns")
	}

	foundMaxTurnsIssue := false
	for _, issue := range result.Issues {
		if containsStr(issue.Message, "maxTurns") || issue.Field == "maxTurns" {
			foundMaxTurnsIssue = true
			break
		}
	}
	if !foundMaxTurnsIssue {
		t.Errorf("expected issue about maxTurns, got: %v", result.Issues)
	}
}

func TestValidateAgentFrontmatter_MaxTurnsZero_StrictMode(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: specialist
description: "Specialist with unlimited turns"
type: specialist
tools: Read, Bash
maxTurns: 0
---

# Specialist
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid but produce warning
	if !result.Valid {
		t.Errorf("expected valid for maxTurns=0, got issues: %v", result.Issues)
	}

	foundMaxTurnsWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "maxTurns") {
			foundMaxTurnsWarning = true
			break
		}
	}
	if !foundMaxTurnsWarning {
		t.Errorf("expected warning about maxTurns=0, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_DisallowedToolsUnknown(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: test-agent
description: "Agent with unknown disallowed tool"
tools: Read
disallowedTools: Bash, UnknownTool
---

# Test Agent
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid but produce warning
	if !result.Valid {
		t.Errorf("expected valid with unknown disallowedTools, got issues: %v", result.Issues)
	}

	foundDisallowedToolsWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "disallowedTools") && containsStr(w, "UnknownTool") {
			foundDisallowedToolsWarning = true
			break
		}
	}
	if !foundDisallowedToolsWarning {
		t.Errorf("expected warning about unknown disallowedTools, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_OrchestratorHighMaxTurns(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: orchestrator
description: "Orchestrator with too many turns"
type: orchestrator
tools: Read
model: opus
maxTurns: 10
---

# Orchestrator
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundMaxTurnsWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "maxTurns") && containsStr(w, "5") {
			foundMaxTurnsWarning = true
			break
		}
	}
	if !foundMaxTurnsWarning {
		t.Errorf("expected warning about orchestrator maxTurns > 5, got warnings: %v", result.Warnings)
	}
}

func TestValidateAgentFrontmatter_OrchestratorNoDisallowedTools(t *testing.T) {
	av := newTestValidator(t)

	content := []byte(`---
name: orchestrator
description: "Orchestrator without disallowedTools"
type: orchestrator
tools: Read
model: opus
maxTurns: 3
---

# Orchestrator
`)

	result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundDisallowedToolsWarning := false
	for _, w := range result.Warnings {
		if containsStr(w, "disallowedTools") {
			foundDisallowedToolsWarning = true
			break
		}
	}
	if !foundDisallowedToolsWarning {
		t.Errorf("expected warning about missing disallowedTools for orchestrator, got warnings: %v", result.Warnings)
	}
}
