package materialize

import (
	"reflect"
	"strings"
	"testing"
)

func TestMergeAgentDefaults(t *testing.T) {
	tests := []struct {
		name     string
		defaults map[string]interface{}
		agentFM  map[string]interface{}
		want     map[string]interface{}
	}{
		{
			name: "scalar override: agent has model, default has model, agent wins",
			defaults: map[string]interface{}{
				"model": "sonnet",
			},
			agentFM: map[string]interface{}{
				"model": "opus",
			},
			want: map[string]interface{}{
				"model": "opus",
			},
		},
		{
			name: "scalar default: agent lacks model, default has model, default used",
			defaults: map[string]interface{}{
				"model": "sonnet",
			},
			agentFM: map[string]interface{}{
				"description": "test agent",
			},
			want: map[string]interface{}{
				"description": "test agent",
				"model":       "sonnet",
			},
		},
		{
			name: "scalar override: maxTurns agent wins",
			defaults: map[string]interface{}{
				"maxTurns": 20,
			},
			agentFM: map[string]interface{}{
				"maxTurns": 40,
			},
			want: map[string]interface{}{
				"maxTurns": 40,
			},
		},
		{
			name: "skills scalar: agent wins over default (skill_policies is sole injection mechanism)",
			defaults: map[string]interface{}{
				"skills": []interface{}{"orchestrator-templates"},
			},
			agentFM: map[string]interface{}{
				"skills": []interface{}{"ecosystem-ref"},
			},
			want: map[string]interface{}{
				"skills": []interface{}{"ecosystem-ref"},
			},
		},
		{
			name: "skills scalar: agent wins, default ignored entirely",
			defaults: map[string]interface{}{
				"skills": []interface{}{"orchestrator-templates", "shared-ref"},
			},
			agentFM: map[string]interface{}{
				"skills": []interface{}{"orchestrator-templates", "ecosystem-ref"},
			},
			want: map[string]interface{}{
				"skills": []interface{}{"orchestrator-templates", "ecosystem-ref"},
			},
		},
		{
			name: "skills default only: agent lacks skills, default used",
			defaults: map[string]interface{}{
				"skills": []interface{}{"orchestrator-templates"},
			},
			agentFM: map[string]interface{}{
				"description": "test",
			},
			want: map[string]interface{}{
				"description": "test",
				"skills":      []interface{}{"orchestrator-templates"},
			},
		},
		{
			name: "skills agent only: no default skills, agent preserved",
			defaults: map[string]interface{}{
				"model": "sonnet",
			},
			agentFM: map[string]interface{}{
				"skills": []interface{}{"ecosystem-ref"},
			},
			want: map[string]interface{}{
				"model":  "sonnet",
				"skills": []interface{}{"ecosystem-ref"},
			},
		},
		{
			name: "disallowedTools replace: agent defines its own, default ignored",
			defaults: map[string]interface{}{
				"disallowedTools": []interface{}{"Write", "Edit"},
			},
			agentFM: map[string]interface{}{
				"disallowedTools": []interface{}{"Bash"},
			},
			want: map[string]interface{}{
				"disallowedTools": []interface{}{"Bash"},
			},
		},
		{
			name: "disallowedTools default: agent lacks it, default used",
			defaults: map[string]interface{}{
				"disallowedTools": []interface{}{"Write", "Edit"},
			},
			agentFM: map[string]interface{}{
				"description": "test",
			},
			want: map[string]interface{}{
				"description":     "test",
				"disallowedTools": []interface{}{"Write", "Edit"},
			},
		},
		{
			name: "tools replace: agent defines its own tools list",
			defaults: map[string]interface{}{
				"tools": []interface{}{"Read", "Grep"},
			},
			agentFM: map[string]interface{}{
				"tools": []interface{}{"Bash", "Read", "Write"},
			},
			want: map[string]interface{}{
				"tools": []interface{}{"Bash", "Read", "Write"},
			},
		},
		{
			name: "allowedTools replace: agent defines its own",
			defaults: map[string]interface{}{
				"allowedTools": []interface{}{"Read"},
			},
			agentFM: map[string]interface{}{
				"allowedTools": []interface{}{"Read", "Grep", "Glob"},
			},
			want: map[string]interface{}{
				"allowedTools": []interface{}{"Read", "Grep", "Glob"},
			},
		},
		{
			name: "map deep merge: hooks, agent has PostToolUse, default has PreToolUse, both present",
			defaults: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse": "default-pre-hook",
				},
			},
			agentFM: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PostToolUse": "agent-post-hook",
				},
			},
			want: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse":  "default-pre-hook",
					"PostToolUse": "agent-post-hook",
				},
			},
		},
		{
			name: "map deep merge: hooks conflict, agent key wins",
			defaults: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse": "default-pre-hook",
				},
			},
			agentFM: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse": "agent-pre-hook",
				},
			},
			want: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse": "agent-pre-hook",
				},
			},
		},
		{
			name: "map deep merge: mcpServers",
			defaults: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"github": map[string]interface{}{"url": "https://api.github.com"},
				},
			},
			agentFM: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"slack": map[string]interface{}{"url": "https://api.slack.com"},
				},
			},
			want: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"github": map[string]interface{}{"url": "https://api.github.com"},
					"slack":  map[string]interface{}{"url": "https://api.slack.com"},
				},
			},
		},
		{
			name:     "nil defaults: agent unchanged",
			defaults: nil,
			agentFM: map[string]interface{}{
				"model":       "opus",
				"description": "test",
			},
			want: map[string]interface{}{
				"model":       "opus",
				"description": "test",
			},
		},
		{
			name:     "empty defaults: agent unchanged",
			defaults: map[string]interface{}{},
			agentFM: map[string]interface{}{
				"model": "opus",
			},
			want: map[string]interface{}{
				"model": "opus",
			},
		},
		{
			name: "empty agent frontmatter: defaults applied",
			defaults: map[string]interface{}{
				"model":    "sonnet",
				"maxTurns": 20,
				"skills":   []interface{}{"orchestrator-templates"},
			},
			agentFM: map[string]interface{}{},
			want: map[string]interface{}{
				"model":    "sonnet",
				"maxTurns": 20,
				"skills":   []interface{}{"orchestrator-templates"},
			},
		},
		{
			name: "nil agent frontmatter: defaults applied",
			defaults: map[string]interface{}{
				"model": "sonnet",
			},
			agentFM: nil,
			want: map[string]interface{}{
				"model": "sonnet",
			},
		},
		{
			name: "unknown fields in defaults pass through",
			defaults: map[string]interface{}{
				"customField":  "custom-value",
				"anotherField": 42,
			},
			agentFM: map[string]interface{}{
				"description": "test",
			},
			want: map[string]interface{}{
				"description":  "test",
				"customField":  "custom-value",
				"anotherField": 42,
			},
		},
		{
			name: "mixed scenario: scalars + skills + disallowedTools + maps",
			defaults: map[string]interface{}{
				"model":           "sonnet",
				"maxTurns":        20,
				"skills":          []interface{}{"orchestrator-templates"},
				"disallowedTools": []interface{}{"Write", "Edit"},
				"color":           "blue",
				"hooks": map[string]interface{}{
					"PreToolUse": "default-hook",
				},
			},
			agentFM: map[string]interface{}{
				"model":           "opus",
				"skills":          []interface{}{"ecosystem-ref"},
				"disallowedTools": []interface{}{"Bash"},
				"hooks": map[string]interface{}{
					"PostToolUse": "agent-hook",
				},
			},
			want: map[string]interface{}{
				"model":           "opus",                   // agent wins
				"maxTurns":        20,                       // default (agent lacks)
				"skills":          []interface{}{"ecosystem-ref"}, // agent wins (scalar semantics)
				"disallowedTools": []interface{}{"Bash"},    // agent replaces
				"color":           "blue",                   // default (agent lacks)
				"hooks": map[string]interface{}{ // deep merged
					"PreToolUse":  "default-hook",
					"PostToolUse": "agent-hook",
				},
			},
		},
		{
			name: "scalar types: string, int, bool all handled",
			defaults: map[string]interface{}{
				"memory":         "project",
				"maxTurns":       20,
				"permissionMode": "plan",
			},
			agentFM: map[string]interface{}{
				"memory": "user",
			},
			want: map[string]interface{}{
				"memory":         "user", // agent wins
				"maxTurns":       20,
				"permissionMode": "plan",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeAgentDefaults(tt.defaults, tt.agentFM)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeAgentDefaults() =\n  %v\nwant:\n  %v", got, tt.want)
			}
		})
	}
}

func TestMergeAgentDefaults_DoesNotMutateInputs(t *testing.T) {
	defaults := map[string]interface{}{
		"model":  "sonnet",
		"skills": []interface{}{"base-skill"},
	}
	agentFM := map[string]interface{}{
		"model":  "opus",
		"skills": []interface{}{"agent-skill"},
	}

	// Snapshot originals
	origDefaultModel := defaults["model"]
	origAgentModel := agentFM["model"]

	_ = MergeAgentDefaults(defaults, agentFM)

	// Verify originals unchanged
	if defaults["model"] != origDefaultModel {
		t.Error("defaults map was mutated")
	}
	if agentFM["model"] != origAgentModel {
		t.Error("agentFM map was mutated")
	}
}

func TestTransformAgentContent_WithAgentDefaults(t *testing.T) {
	t.Run("defaults merged before stripping", func(t *testing.T) {
		source := `---
description: Test agent
---

# Test Agent
`
		agentDefaults := map[string]interface{}{
			"model":    "sonnet",
			"maxTurns": 20,
			"skills":   []interface{}{"orchestrator-templates"},
			// knossos-only field in defaults — should be stripped
			"type": "specialist",
		}

		result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", AgentDefaults: agentDefaults})
		if err != nil {
			t.Fatalf("transformAgentContent() error = %v", err)
		}

		output := string(result)

		// CC-consumable fields from defaults should be present
		if !strings.Contains(output, "model: sonnet") {
			t.Errorf("default model should be present:\n%s", output)
		}
		if !strings.Contains(output, "maxTurns: 20") {
			t.Errorf("default maxTurns should be present:\n%s", output)
		}
		if !strings.Contains(output, "orchestrator-templates") {
			t.Errorf("default skills should be present:\n%s", output)
		}

		// knossos-only field from defaults should be stripped
		if strings.Contains(output, "type:") {
			t.Errorf("knossos-only field from defaults should be stripped:\n%s", output)
		}

		// Name should be injected
		if !strings.Contains(output, "name: test-agent") {
			t.Errorf("name should be injected:\n%s", output)
		}
	})

	t.Run("agent values override defaults", func(t *testing.T) {
		source := `---
description: Test agent
model: opus
maxTurns: 40
---

# Test Agent
`
		agentDefaults := map[string]interface{}{
			"model":    "sonnet",
			"maxTurns": 20,
		}

		result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", AgentDefaults: agentDefaults})
		if err != nil {
			t.Fatalf("transformAgentContent() error = %v", err)
		}

		output := string(result)

		if !strings.Contains(output, "model: opus") {
			t.Errorf("agent model should win over default:\n%s", output)
		}
		if !strings.Contains(output, "maxTurns: 40") {
			t.Errorf("agent maxTurns should win over default:\n%s", output)
		}
	})

	t.Run("skills: agent wins over defaults (scalar semantics)", func(t *testing.T) {
		source := `---
description: Test agent
skills:
  - ecosystem-ref
---

# Test Agent
`
		agentDefaults := map[string]interface{}{
			"skills": []interface{}{"orchestrator-templates"},
		}

		result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", AgentDefaults: agentDefaults})
		if err != nil {
			t.Fatalf("transformAgentContent() error = %v", err)
		}

		output := string(result)

		// Agent skill wins; default skill is NOT injected (skill_policies handles injection)
		if strings.Contains(output, "orchestrator-templates") {
			t.Errorf("default skill should NOT be present (agent value wins):\n%s", output)
		}
		if !strings.Contains(output, "ecosystem-ref") {
			t.Errorf("agent skill should be present:\n%s", output)
		}
	})

	t.Run("write-guard from defaults works with hook resolution", func(t *testing.T) {
		source := `---
description: Test agent
write-guard: true
---

# Test Agent
`
		agentDefaults := map[string]interface{}{
			"model": "sonnet",
		}

		result, err := transformAgentContent([]byte(source), &TransformContext{AgentName: "test-agent", WriteGuardDefaults: testDefaults, AgentDefaults: agentDefaults})
		if err != nil {
			t.Fatalf("transformAgentContent() error = %v", err)
		}

		output := string(result)

		// Model from defaults should be present
		if !strings.Contains(output, "model: sonnet") {
			t.Errorf("default model should be present:\n%s", output)
		}

		// Write-guard hooks should be generated (from testDefaults, not agentDefaults)
		if !strings.Contains(output, "agent-guard") {
			t.Errorf("write-guard hooks should be generated:\n%s", output)
		}

		// write-guard key should be stripped
		if strings.Contains(output, "write-guard") {
			t.Errorf("write-guard should be stripped:\n%s", output)
		}
	})
}
