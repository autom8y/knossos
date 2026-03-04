package materialize

import (
	"reflect"
	"strings"
	"testing"
)

func TestMergeAgentDefaults(t *testing.T) {
	tests := []struct {
		name     string
		defaults map[string]any
		agentFM  map[string]any
		want     map[string]any
	}{
		{
			name: "scalar override: agent has model, default has model, agent wins",
			defaults: map[string]any{
				"model": "sonnet",
			},
			agentFM: map[string]any{
				"model": "opus",
			},
			want: map[string]any{
				"model": "opus",
			},
		},
		{
			name: "scalar default: agent lacks model, default has model, default used",
			defaults: map[string]any{
				"model": "sonnet",
			},
			agentFM: map[string]any{
				"description": "test agent",
			},
			want: map[string]any{
				"description": "test agent",
				"model":       "sonnet",
			},
		},
		{
			name: "scalar override: maxTurns agent wins",
			defaults: map[string]any{
				"maxTurns": 20,
			},
			agentFM: map[string]any{
				"maxTurns": 40,
			},
			want: map[string]any{
				"maxTurns": 40,
			},
		},
		{
			name: "skills scalar: agent wins over default (skill_policies is sole injection mechanism)",
			defaults: map[string]any{
				"skills": []any{"orchestrator-templates"},
			},
			agentFM: map[string]any{
				"skills": []any{"ecosystem-ref"},
			},
			want: map[string]any{
				"skills": []any{"ecosystem-ref"},
			},
		},
		{
			name: "skills scalar: agent wins, default ignored entirely",
			defaults: map[string]any{
				"skills": []any{"orchestrator-templates", "shared-ref"},
			},
			agentFM: map[string]any{
				"skills": []any{"orchestrator-templates", "ecosystem-ref"},
			},
			want: map[string]any{
				"skills": []any{"orchestrator-templates", "ecosystem-ref"},
			},
		},
		{
			name: "skills default only: agent lacks skills, default used",
			defaults: map[string]any{
				"skills": []any{"orchestrator-templates"},
			},
			agentFM: map[string]any{
				"description": "test",
			},
			want: map[string]any{
				"description": "test",
				"skills":      []any{"orchestrator-templates"},
			},
		},
		{
			name: "skills agent only: no default skills, agent preserved",
			defaults: map[string]any{
				"model": "sonnet",
			},
			agentFM: map[string]any{
				"skills": []any{"ecosystem-ref"},
			},
			want: map[string]any{
				"model":  "sonnet",
				"skills": []any{"ecosystem-ref"},
			},
		},
		{
			name: "disallowedTools replace: agent defines its own, default ignored",
			defaults: map[string]any{
				"disallowedTools": []any{"Write", "Edit"},
			},
			agentFM: map[string]any{
				"disallowedTools": []any{"Bash"},
			},
			want: map[string]any{
				"disallowedTools": []any{"Bash"},
			},
		},
		{
			name: "disallowedTools default: agent lacks it, default used",
			defaults: map[string]any{
				"disallowedTools": []any{"Write", "Edit"},
			},
			agentFM: map[string]any{
				"description": "test",
			},
			want: map[string]any{
				"description":     "test",
				"disallowedTools": []any{"Write", "Edit"},
			},
		},
		{
			name: "tools replace: agent defines its own tools list",
			defaults: map[string]any{
				"tools": []any{"Read", "Grep"},
			},
			agentFM: map[string]any{
				"tools": []any{"Bash", "Read", "Write"},
			},
			want: map[string]any{
				"tools": []any{"Bash", "Read", "Write"},
			},
		},
		{
			name: "allowedTools replace: agent defines its own",
			defaults: map[string]any{
				"allowedTools": []any{"Read"},
			},
			agentFM: map[string]any{
				"allowedTools": []any{"Read", "Grep", "Glob"},
			},
			want: map[string]any{
				"allowedTools": []any{"Read", "Grep", "Glob"},
			},
		},
		{
			name: "map deep merge: hooks, agent has PostToolUse, default has PreToolUse, both present",
			defaults: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": "default-pre-hook",
				},
			},
			agentFM: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": "agent-post-hook",
				},
			},
			want: map[string]any{
				"hooks": map[string]any{
					"PreToolUse":  "default-pre-hook",
					"PostToolUse": "agent-post-hook",
				},
			},
		},
		{
			name: "map deep merge: hooks conflict, agent key wins",
			defaults: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": "default-pre-hook",
				},
			},
			agentFM: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": "agent-pre-hook",
				},
			},
			want: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": "agent-pre-hook",
				},
			},
		},
		{
			name: "map deep merge: mcpServers",
			defaults: map[string]any{
				"mcpServers": map[string]any{
					"github": map[string]any{"url": "https://api.github.com"},
				},
			},
			agentFM: map[string]any{
				"mcpServers": map[string]any{
					"slack": map[string]any{"url": "https://api.slack.com"},
				},
			},
			want: map[string]any{
				"mcpServers": map[string]any{
					"github": map[string]any{"url": "https://api.github.com"},
					"slack":  map[string]any{"url": "https://api.slack.com"},
				},
			},
		},
		{
			name:     "nil defaults: agent unchanged",
			defaults: nil,
			agentFM: map[string]any{
				"model":       "opus",
				"description": "test",
			},
			want: map[string]any{
				"model":       "opus",
				"description": "test",
			},
		},
		{
			name:     "empty defaults: agent unchanged",
			defaults: map[string]any{},
			agentFM: map[string]any{
				"model": "opus",
			},
			want: map[string]any{
				"model": "opus",
			},
		},
		{
			name: "empty agent frontmatter: defaults applied",
			defaults: map[string]any{
				"model":    "sonnet",
				"maxTurns": 20,
				"skills":   []any{"orchestrator-templates"},
			},
			agentFM: map[string]any{},
			want: map[string]any{
				"model":    "sonnet",
				"maxTurns": 20,
				"skills":   []any{"orchestrator-templates"},
			},
		},
		{
			name: "nil agent frontmatter: defaults applied",
			defaults: map[string]any{
				"model": "sonnet",
			},
			agentFM: nil,
			want: map[string]any{
				"model": "sonnet",
			},
		},
		{
			name: "unknown fields in defaults pass through",
			defaults: map[string]any{
				"customField":  "custom-value",
				"anotherField": 42,
			},
			agentFM: map[string]any{
				"description": "test",
			},
			want: map[string]any{
				"description":  "test",
				"customField":  "custom-value",
				"anotherField": 42,
			},
		},
		{
			name: "mixed scenario: scalars + skills + disallowedTools + maps",
			defaults: map[string]any{
				"model":           "sonnet",
				"maxTurns":        20,
				"skills":          []any{"orchestrator-templates"},
				"disallowedTools": []any{"Write", "Edit"},
				"color":           "blue",
				"hooks": map[string]any{
					"PreToolUse": "default-hook",
				},
			},
			agentFM: map[string]any{
				"model":           "opus",
				"skills":          []any{"ecosystem-ref"},
				"disallowedTools": []any{"Bash"},
				"hooks": map[string]any{
					"PostToolUse": "agent-hook",
				},
			},
			want: map[string]any{
				"model":           "opus",                 // agent wins
				"maxTurns":        20,                     // default (agent lacks)
				"skills":          []any{"ecosystem-ref"}, // agent wins (scalar semantics)
				"disallowedTools": []any{"Bash"},          // agent replaces
				"color":           "blue",                 // default (agent lacks)
				"hooks": map[string]any{ // deep merged
					"PreToolUse":  "default-hook",
					"PostToolUse": "agent-hook",
				},
			},
		},
		{
			name: "scalar types: string, int, bool all handled",
			defaults: map[string]any{
				"memory":         "project",
				"maxTurns":       20,
				"permissionMode": "plan",
			},
			agentFM: map[string]any{
				"memory": "user",
			},
			want: map[string]any{
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
	defaults := map[string]any{
		"model":  "sonnet",
		"skills": []any{"base-skill"},
	}
	agentFM := map[string]any{
		"model":  "opus",
		"skills": []any{"agent-skill"},
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
		agentDefaults := map[string]any{
			"model":    "sonnet",
			"maxTurns": 20,
			"skills":   []any{"orchestrator-templates"},
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
		agentDefaults := map[string]any{
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
		agentDefaults := map[string]any{
			"skills": []any{"orchestrator-templates"},
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
		agentDefaults := map[string]any{
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
