package materialize

import (
	"strings"
	"testing"
)

var testDefaults = &WriteGuardDefaults{
	AllowPaths: []string{".ledge/", ".know/"},
	Timeout:    3,
}

func TestTransformAgentContent_OptIn(t *testing.T) {
	source := `---
name: cruft-cutter
description: Detects temporal debt
tools: Bash, Read, Write
write-guard: true
disallowedTools:
  - Edit
---

# Cruft Cutter

Body content here.
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "cruft-cutter", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// write-guard key must be removed from output
	if strings.Contains(output, "write-guard") {
		t.Error("output should not contain write-guard key")
	}

	// hooks block must be present with correct command
	if !strings.Contains(output, "ari hook agent-guard --agent cruft-cutter --allow-path .ledge/ --allow-path .know/ --output json") {
		t.Errorf("output missing generated write-guard command:\n%s", output)
	}

	// Body must be preserved
	if !strings.Contains(output, "# Cruft Cutter") {
		t.Errorf("output missing body content:\n%s", output)
	}
	if !strings.Contains(output, "Body content here.") {
		t.Errorf("output missing body text:\n%s", output)
	}
}

func TestTransformAgentContent_ExtraPaths(t *testing.T) {
	source := `---
name: risk-assessor
description: Scores debt by risk
write-guard:
  extra-paths:
    - docs/
---

# Risk Assessor
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "risk-assessor", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// Must contain all three paths
	if !strings.Contains(output, "--allow-path .ledge/") {
		t.Error("output missing .ledge/ path")
	}
	if !strings.Contains(output, "--allow-path .know/") {
		t.Error("output missing .know/ path")
	}
	if !strings.Contains(output, "--allow-path docs/") {
		t.Error("output missing docs/ path")
	}

	// write-guard key removed
	if strings.Contains(output, "write-guard") {
		t.Error("output should not contain write-guard key")
	}
}

func TestTransformAgentContent_OptOut(t *testing.T) {
	source := `---
name: some-agent
description: An agent that opts out
write-guard: false
---

# Some Agent
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "some-agent", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// No hooks should be present
	if strings.Contains(output, "agent-guard") {
		t.Errorf("opted-out agent should not have agent-guard hooks:\n%s", output)
	}

	// write-guard key removed
	if strings.Contains(output, "write-guard") {
		t.Error("output should not contain write-guard key")
	}

	// Body preserved
	if !strings.Contains(output, "# Some Agent") {
		t.Error("output missing body")
	}
}

func TestTransformAgentContent_NoWriteGuard(t *testing.T) {
	source := `---
name: integration-engineer
description: Implements changes
tools: Bash, Read, Write, Edit
---

# Integration Engineer
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "integration-engineer", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// Name should be present (injected from agentName param)
	if !strings.Contains(output, "name: integration-engineer") {
		t.Errorf("output missing injected name:\n%s", output)
	}

	// Description and tools should survive (not in blocklist)
	if !strings.Contains(output, "description:") {
		t.Errorf("output missing description:\n%s", output)
	}
	if !strings.Contains(output, "tools:") {
		t.Errorf("output missing tools:\n%s", output)
	}

	// No hooks should be generated (no write-guard directive)
	if strings.Contains(output, "agent-guard") {
		t.Errorf("should not generate hooks when no write-guard present:\n%s", output)
	}

	// Body must be preserved
	if !strings.Contains(output, "# Integration Engineer") {
		t.Errorf("output missing body:\n%s", output)
	}
}

func TestTransformAgentContent_BodyPreservation(t *testing.T) {
	body := `# Cruft Cutter

The fat trimmer. Detects code that was correct at time-of-write.

## Core Responsibilities

- Scan for backwards-compatibility shims
- Identify stale feature flags

` + "```go\nfunc example() {}\n```\n"

	source := "---\nname: cruft-cutter\nwrite-guard: true\n---\n" + body

	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "cruft-cutter", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	// Body must be preserved exactly
	output := string(result)
	if !strings.Contains(output, body) {
		t.Errorf("body not preserved exactly\ngot:\n%s", output)
	}
}

func TestTransformAgentContent_ExistingHooksMerge(t *testing.T) {
	source := `---
name: test-agent
description: Agent with existing hooks
write-guard: true
hooks:
  PostToolUse:
    - matcher: "Write"
      hooks:
        - type: command
          command: "ari hook clew --output json"
          timeout: 5
---

# Test Agent
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// Generated write-guard hook must be present
	if !strings.Contains(output, "agent-guard") {
		t.Error("output missing generated write-guard hook")
	}

	// Existing PostToolUse hook must be preserved
	if !strings.Contains(output, "clew") {
		t.Errorf("existing PostToolUse hook should be preserved:\n%s", output)
	}

	// write-guard key removed
	if strings.Contains(output, "write-guard") {
		t.Error("output should not contain write-guard key")
	}
}

func TestTransformAgentContent_UnknownFieldsSurvive(t *testing.T) {
	source := `---
name: test-agent
description: Agent with custom fields
custom_field: preserved
write-guard: true
---

# Test Agent
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "custom_field") {
		t.Errorf("unknown fields should survive transformation:\n%s", output)
	}
	if !strings.Contains(output, "preserved") {
		t.Errorf("unknown field value should survive:\n%s", output)
	}
}

func TestTransformAgentContent_StripKnossosFields(t *testing.T) {
	source := `---
name: test-agent
description: Agent with knossos metadata
tools: Bash, Read
type: specialist
role: implementer
color: blue
upstream: architect
downstream: qa-adversary
produces: code
contract: strict
schema_version: "2"
aliases:
  - test
  - tester
---

# Test Agent
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// All knossos-only fields must be stripped
	for _, field := range []string{"type:", "role:", "upstream:", "downstream:", "produces:", "contract:", "schema_version:", "aliases:"} {
		if strings.Contains(output, field) {
			t.Errorf("knossos-only field %q should be stripped from output:\n%s", field, output)
		}
	}

	// CC-consumable fields must survive (including color — CC uses it for subagent UI)
	if !strings.Contains(output, "name: test-agent") {
		t.Errorf("name field should survive:\n%s", output)
	}
	if !strings.Contains(output, "description:") {
		t.Errorf("description field should survive:\n%s", output)
	}
	if !strings.Contains(output, "tools:") {
		t.Errorf("tools field should survive:\n%s", output)
	}
	if !strings.Contains(output, "color: blue") {
		t.Errorf("color field should survive (CC-consumed):\n%s", output)
	}

	// Body preserved
	if !strings.Contains(output, "# Test Agent") {
		t.Errorf("body should be preserved:\n%s", output)
	}
}

func TestTransformAgentContent_NameInjection(t *testing.T) {
	// Source has NO name field — name should be injected from agentName param
	source := `---
description: Agent without explicit name
tools: Bash
---

# Nameless Agent
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "injected-name", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "name: injected-name") {
		t.Errorf("name should be injected from agentName parameter:\n%s", output)
	}

	if !strings.Contains(output, "# Nameless Agent") {
		t.Errorf("body should be preserved:\n%s", output)
	}
}

func TestTransformAgentContent_NameOverride(t *testing.T) {
	// Source has a name field that differs from agentName — agentName wins (filename is source of truth)
	source := `---
name: old-name
description: Agent with mismatched name
---

# Agent
`
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "filename-name", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "name: filename-name") {
		t.Errorf("name should be overridden by agentName parameter:\n%s", output)
	}
	if strings.Contains(output, "old-name") {
		t.Errorf("original name should be replaced:\n%s", output)
	}
}

func TestTransformAgentContent_NilDefaults(t *testing.T) {
	source := `---
name: test-agent
write-guard: true
---

# Test Agent
`
	// With nil defaults, even write-guard: true should produce no hooks
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent"})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	// write-guard key should still be removed (it's a source-only directive)
	// But since defaults are nil, ResolveWriteGuard returns nil, so no hooks generated
	if strings.Contains(output, "agent-guard") {
		t.Error("should not generate hooks when defaults are nil")
	}
}

func TestTransformAgentContent_InvalidFrontmatter(t *testing.T) {
	// No frontmatter delimiters — should pass through unchanged
	source := "# Just a markdown file\n\nNo frontmatter here.\n"
	result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test", WriteGuardDefaults: testDefaults})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}
	if string(result) != source {
		t.Error("content without frontmatter should pass through unchanged")
	}
}

func TestTransformAgentContent_ModelOverride(t *testing.T) {
	// Agent has model: opus — override should force haiku
	source := `---
name: potnia
description: Orchestrator
model: opus
tools: Bash, Read
---

# Potnia
`
	result, err := transformAgentContent([]byte(source), &TransformContext{
		AgentName:     "potnia",
		ModelOverride: "haiku",
	})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "model: haiku") {
		t.Errorf("model override should force haiku:\n%s", output)
	}
	if strings.Contains(output, "model: opus") {
		t.Errorf("original model should be replaced:\n%s", output)
	}
}

func TestTransformAgentContent_ModelOverrideNoExistingModel(t *testing.T) {
	// Agent has no model field — override should inject one
	source := `---
name: worker
description: A worker agent
tools: Bash
---

# Worker
`
	result, err := transformAgentContent([]byte(source), &TransformContext{
		AgentName:     "worker",
		ModelOverride: "haiku",
	})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "model: haiku") {
		t.Errorf("model override should inject model field:\n%s", output)
	}
}

func TestTransformAgentContent_NoModelOverride(t *testing.T) {
	// Empty ModelOverride should leave original model unchanged
	source := `---
name: specialist
description: A specialist
model: sonnet
---

# Specialist
`
	result, err := transformAgentContent([]byte(source), &TransformContext{
		AgentName:     "specialist",
		ModelOverride: "",
	})
	if err != nil {
		t.Fatalf("transformAgentContent() error = %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "model: sonnet") {
		t.Errorf("original model should be preserved when no override:\n%s", output)
	}
}
