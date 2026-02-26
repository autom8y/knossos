package materialize

import (
	"reflect"
	"testing"
)

func TestResolveHookDefaults(t *testing.T) {
	tests := []struct {
		name   string
		shared *HookDefaults
		rite   *HookDefaults
		want   *WriteGuardDefaults
	}{
		{
			name: "3-tier cascade: shared base + rite extras merge",
			shared: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					AllowPaths: []string{".wip/", "wip/"},
					Timeout:    3,
				},
			},
			rite: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					ExtraPaths: []string{"docs/ecosystem/"},
				},
			},
			want: &WriteGuardDefaults{
				AllowPaths: []string{".wip/", "wip/", "docs/ecosystem/"},
				Timeout:    3,
			},
		},
		{
			name: "shared-only: no rite overrides",
			shared: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					AllowPaths: []string{".wip/", "wip/"},
					Timeout:    3,
				},
			},
			rite: nil,
			want: &WriteGuardDefaults{
				AllowPaths: []string{".wip/", "wip/"},
				Timeout:    3,
			},
		},
		{
			name:   "no defaults at all: both nil",
			shared: nil,
			rite:   nil,
			want:   nil,
		},
		{
			name:   "rite-only: no shared defaults",
			shared: nil,
			rite: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					AllowPaths: []string{"docs/"},
					Timeout:    5,
				},
			},
			want: &WriteGuardDefaults{
				AllowPaths: []string{"docs/"},
				Timeout:    5,
			},
		},
		{
			name: "deduplication: same path in shared and rite",
			shared: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					AllowPaths: []string{".wip/", "wip/"},
					Timeout:    3,
				},
			},
			rite: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					ExtraPaths: []string{".wip/"}, // duplicate
				},
			},
			want: &WriteGuardDefaults{
				AllowPaths: []string{".wip/", "wip/"},
				Timeout:    3,
			},
		},
		{
			name: "rite timeout overrides shared",
			shared: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					AllowPaths: []string{".wip/"},
					Timeout:    3,
				},
			},
			rite: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					Timeout: 5,
				},
			},
			want: &WriteGuardDefaults{
				AllowPaths: []string{".wip/"},
				Timeout:    5,
			},
		},
		{
			name: "default timeout applied when zero",
			shared: &HookDefaults{
				WriteGuard: &WriteGuardDefaults{
					AllowPaths: []string{".wip/"},
				},
			},
			rite: nil,
			want: &WriteGuardDefaults{
				AllowPaths: []string{".wip/"},
				Timeout:    defaultWriteGuardTimeout,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveHookDefaults(tt.shared, tt.rite)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ResolveHookDefaults() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("ResolveHookDefaults() = nil, want non-nil")
			}
			if !reflect.DeepEqual(got.AllowPaths, tt.want.AllowPaths) {
				t.Errorf("AllowPaths = %v, want %v", got.AllowPaths, tt.want.AllowPaths)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %d, want %d", got.Timeout, tt.want.Timeout)
			}
		})
	}
}

func TestResolveWriteGuard(t *testing.T) {
	defaults := &WriteGuardDefaults{
		AllowPaths: []string{".wip/", "wip/"},
		Timeout:    3,
	}

	tests := []struct {
		name      string
		defaults  *WriteGuardDefaults
		agentName string
		agentWG   interface{}
		want      *ResolvedWriteGuard
	}{
		{
			name:      "opt-in with true",
			defaults:  defaults,
			agentName: "cruft-cutter",
			agentWG:   true,
			want: &ResolvedWriteGuard{
				AllowPaths: []string{".wip/", "wip/"},
				Timeout:    3,
				AgentName:  "cruft-cutter",
			},
		},
		{
			name:      "opt-out with false",
			defaults:  defaults,
			agentName: "cruft-cutter",
			agentWG:   false,
			want:      nil,
		},
		{
			name:      "no write-guard key (nil)",
			defaults:  defaults,
			agentName: "cruft-cutter",
			agentWG:   nil,
			want:      nil,
		},
		{
			name:      "no defaults (nil)",
			defaults:  nil,
			agentName: "cruft-cutter",
			agentWG:   true,
			want:      nil,
		},
		{
			name:      "extra-paths additive merge",
			defaults:  defaults,
			agentName: "risk-assessor",
			agentWG: map[string]interface{}{
				"extra-paths": []interface{}{"docs/"},
			},
			want: &ResolvedWriteGuard{
				AllowPaths: []string{".wip/", "wip/", "docs/"},
				Timeout:    3,
				AgentName:  "risk-assessor",
			},
		},
		{
			name:      "extra-paths dedup",
			defaults:  defaults,
			agentName: "risk-assessor",
			agentWG: map[string]interface{}{
				"extra-paths": []interface{}{".wip/"}, // already in defaults
			},
			want: &ResolvedWriteGuard{
				AllowPaths: []string{".wip/", "wip/"},
				Timeout:    3,
				AgentName:  "risk-assessor",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveWriteGuard(tt.defaults, tt.agentName, tt.agentWG)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ResolveWriteGuard() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("ResolveWriteGuard() = nil, want non-nil")
			}
			if !reflect.DeepEqual(got.AllowPaths, tt.want.AllowPaths) {
				t.Errorf("AllowPaths = %v, want %v", got.AllowPaths, tt.want.AllowPaths)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %d, want %d", got.Timeout, tt.want.Timeout)
			}
			if got.AgentName != tt.want.AgentName {
				t.Errorf("AgentName = %q, want %q", got.AgentName, tt.want.AgentName)
			}
		})
	}
}

func TestGenerateWriteGuardHooks(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got := GenerateWriteGuardHooks(nil)
		if got != nil {
			t.Errorf("GenerateWriteGuardHooks(nil) = %v, want nil", got)
		}
	})

	t.Run("generates correct CC-compatible structure", func(t *testing.T) {
		resolved := &ResolvedWriteGuard{
			AllowPaths: []string{".wip/", "wip/", "docs/ecosystem/"},
			Timeout:    3,
			AgentName:  "ecosystem-analyst",
		}
		got := GenerateWriteGuardHooks(resolved)

		// Verify top-level key
		preToolUse, ok := got["PreToolUse"]
		if !ok {
			t.Fatal("missing PreToolUse key")
		}

		// Verify matcher group
		groups, ok := preToolUse.([]interface{})
		if !ok || len(groups) != 1 {
			t.Fatalf("PreToolUse should have 1 matcher group, got %v", preToolUse)
		}

		group, ok := groups[0].(map[string]interface{})
		if !ok {
			t.Fatal("matcher group should be a map")
		}

		if group["matcher"] != "Write" {
			t.Errorf("matcher = %q, want Write", group["matcher"])
		}

		// Verify hook entry
		hooks, ok := group["hooks"].([]interface{})
		if !ok || len(hooks) != 1 {
			t.Fatalf("hooks should have 1 entry, got %v", group["hooks"])
		}

		hook, ok := hooks[0].(map[string]interface{})
		if !ok {
			t.Fatal("hook entry should be a map")
		}

		if hook["type"] != "command" {
			t.Errorf("type = %q, want command", hook["type"])
		}
		if hook["timeout"] != 3 {
			t.Errorf("timeout = %v, want 3", hook["timeout"])
		}

		// Verify command string
		wantCmd := "ari hook agent-guard --agent ecosystem-analyst --allow-path .wip/ --allow-path wip/ --allow-path docs/ecosystem/ --output json"
		if hook["command"] != wantCmd {
			t.Errorf("command = %q, want %q", hook["command"], wantCmd)
		}
	})
}

func TestDedup(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"empty", nil, []string{}},
		{"no duplicates", []string{"a", "b"}, []string{"a", "b"}},
		{"with duplicates preserves order", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedup(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dedup(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
