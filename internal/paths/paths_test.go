package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// TestNewResolver verifies that NewResolver stores the project root.
func TestNewResolver(t *testing.T) {
	r := NewResolver("/tmp/myproject")
	if r.ProjectRoot() != "/tmp/myproject" {
		t.Errorf("ProjectRoot() = %q, want %q", r.ProjectRoot(), "/tmp/myproject")
	}
}

// TestResolver_PathMethods verifies every exported path method on Resolver
// returns deterministic, correctly-constructed paths.
func TestResolver_PathMethods(t *testing.T) {
	root := "/tmp/testroot"
	r := NewResolver(root)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ProjectRoot", r.ProjectRoot(), "/tmp/testroot"},
		{"SOSDir", r.SOSDir(), "/tmp/testroot/.sos"},
		{"SessionsDir", r.SessionsDir(), "/tmp/testroot/.sos/sessions"},
		{"LocksDir", r.LocksDir(), "/tmp/testroot/.sos/sessions/.locks"},
		{"HarnessMapDir", r.HarnessMapDir(), "/tmp/testroot/.sos/sessions/.harness-map"},
		{"ArchiveDir", r.ArchiveDir(), "/tmp/testroot/.sos/archive"},
		{"SessionDir", r.SessionDir("s1"), "/tmp/testroot/.sos/sessions/s1"},
		{"SessionContextFile", r.SessionContextFile("s1"), "/tmp/testroot/.sos/sessions/s1/SESSION_CONTEXT.md"},
		{"SessionEventsFile", r.SessionEventsFile("s1"), "/tmp/testroot/.sos/sessions/s1/events.jsonl"},
		{"LockFile", r.LockFile("s1"), "/tmp/testroot/.sos/sessions/.locks/s1.lock"},
		{"CurrentSessionFile", r.CurrentSessionFile(), "/tmp/testroot/.sos/sessions/.current-session"},
		{"ActiveRiteFile", r.ActiveRiteFile(), "/tmp/testroot/.knossos/ACTIVE_RITE"},
		{"ActiveWorkflowFile", r.ActiveWorkflowFile(), "/tmp/testroot/.knossos/ACTIVE_WORKFLOW.yaml"},
		{"KnossosManifestFile", r.KnossosManifestFile(), "/tmp/testroot/.knossos/KNOSSOS_MANIFEST.yaml"},
		{"AgentsDirForChannel/claude", r.AgentsDirForChannel(ClaudeChannel{}), "/tmp/testroot/.claude/agents"},
		{"AgentsDirForChannel/gemini", r.AgentsDirForChannel(GeminiChannel{}), "/tmp/testroot/.gemini/agents"},
		{"AgentsDir", r.AgentsDir(), "/tmp/testroot/" + ClaudeChannel{}.DirName() + "/agents"},
		{"AgentFile", r.AgentFile("potnia.md"), "/tmp/testroot/" + ClaudeChannel{}.DirName() + "/agents/potnia.md"},
		{"ContextFileForChannel/claude", r.ContextFileForChannel(ClaudeChannel{}), "/tmp/testroot/.claude/CLAUDE.md"},
		{"ContextFileForChannel/gemini", r.ContextFileForChannel(GeminiChannel{}), "/tmp/testroot/.gemini/GEMINI.md"},
		{"KnossosDir", r.KnossosDir(), "/tmp/testroot/.knossos"},
		{"RitesDir", r.RitesDir(), "/tmp/testroot/.knossos/rites"},
		{"LedgeDir", r.LedgeDir(), "/tmp/testroot/.ledge"},
		{"LedgeDecisionsDir", r.LedgeDecisionsDir(), "/tmp/testroot/.ledge/decisions"},
		{"LedgeSpecsDir", r.LedgeSpecsDir(), "/tmp/testroot/.ledge/specs"},
		{"LedgeReviewsDir", r.LedgeReviewsDir(), "/tmp/testroot/.ledge/reviews"},
		{"LedgeSpikesDir", r.LedgeSpikesDir(), "/tmp/testroot/.ledge/spikes"},
		{"WipDir", r.WipDir(), "/tmp/testroot/.sos/wip"},
		{"InvocationStateFile", r.InvocationStateFile(), "/tmp/testroot/.knossos/INVOCATION_STATE.yaml"},
		{"KnossosSyncDir", r.KnossosSyncDir(), "/tmp/testroot/.knossos/sync"},
		{"KnossosBackupsDir", r.KnossosBackupsDir(), "/tmp/testroot/.knossos/backups"},
		{"ElCheapoMarkerFile", r.ElCheapoMarkerFile(), "/tmp/testroot/.knossos/.el-cheapo-active"},
		{"WorktreeMetaFile", r.WorktreeMetaFile(), "/tmp/testroot/.knossos/.worktree-meta.json"},
		{"WorktreesDir", r.WorktreesDir(), "/tmp/testroot/.knossos/worktrees"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestResolver_RitePaths verifies rite-related path methods.
// RiteDir does a filesystem check for manifest.yaml in the project rites dir,
// falling back to user rites. We test the fallback case (no manifest on disk).
func TestResolver_RitePaths(t *testing.T) {
	root := t.TempDir()
	r := NewResolver(root)
	rite := "test-rite"

	// Without a project-level manifest.yaml, RiteDir falls back to user rites.
	userRites := UserRitesDir()
	wantRiteDir := filepath.Join(userRites, rite)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"RiteDir_fallback", r.RiteDir(rite), wantRiteDir},
		{"RiteManifestFile", r.RiteManifestFile(rite), filepath.Join(wantRiteDir, "manifest.yaml")},
		{"RiteAgentsDir", r.RiteAgentsDir(rite), filepath.Join(wantRiteDir, "agents")},
		{"RiteSkillsDir", r.RiteSkillsDir(rite), filepath.Join(wantRiteDir, "skills")},
		{"RiteWorkflowFile", r.RiteWorkflowFile(rite), filepath.Join(wantRiteDir, "workflow.yaml")},
		{"RiteOrchestratorFile", r.RiteOrchestratorFile(rite), filepath.Join(wantRiteDir, "orchestrator.yaml")},
		{"RiteContextFile", r.RiteContextFile(rite), filepath.Join(wantRiteDir, "context.yaml")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestResolver_RiteDir_ProjectOverride verifies RiteDir prefers project rites
// when a manifest.yaml exists on disk.
func TestResolver_RiteDir_ProjectOverride(t *testing.T) {
	root := t.TempDir()
	r := NewResolver(root)
	rite := "my-rite"

	// Create project-level satellite rite with manifest.yaml
	riteDir := filepath.Join(root, ".knossos", "rites", rite)
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte("name: my-rite\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := r.RiteDir(rite)
	if got != riteDir {
		t.Errorf("RiteDir(%q) = %q, want project path %q", rite, got, riteDir)
	}
}

// TestIsSessionDir validates session directory name detection with valid,
// invalid, and edge-case inputs.
func TestIsSessionDir(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Valid: meets length >= 32 and "session-" prefix
		{"valid_standard", "session-20260218-120000-abcd1234", true},
		{"valid_longer", "session-20260218-120000-abcdef1234567890", true},
		{"valid_exact_32", "session-20260218-120000-abcd1234", true},

		// Invalid: wrong prefix
		{"invalid_no_prefix", "not-a-session-directory-at-all!!", false},
		{"invalid_uppercase", "SESSION-20260218-120000-abcd1234", false},
		{"invalid_empty", "", false},

		// Edge cases: prefix correct but too short
		{"edge_session_dash_only", "session-", false},
		{"edge_prefix_short", "session-short", false},
		{"edge_31_chars", "session-20260218-120000-abcd123", false},
		{"edge_just_under", "session-20260218-120000-abcd12", false},

		// Edge case: 32+ chars but wrong prefix
		{"long_wrong_prefix", "xession-20260218-120000-abcd1234", false},
		{"long_no_dash", "session020260218-120000-abcd1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSessionDir(tt.input)
			if got != tt.want {
				t.Errorf("IsSessionDir(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestSessionIDFromDir verifies extraction of session ID from a directory path.
func TestSessionIDFromDir(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bare_name", "session-20260218-120000-abcd1234", "session-20260218-120000-abcd1234"},
		{"full_path", "/tmp/root/.sos/sessions/session-20260218-120000-abcd1234", "session-20260218-120000-abcd1234"},
		{"nested", "/a/b/c/my-session", "my-session"},
		{"trailing_slash_removed", "session-id", "session-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SessionIDFromDir(tt.input)
			if got != tt.want {
				t.Errorf("SessionIDFromDir(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestFindProjectRoot verifies project root discovery by walking up directories.
func TestFindProjectRoot(t *testing.T) {
	t.Run("finds_claude_dir", func(t *testing.T) {
		// Create a temp tree: root/{channel}/ and root/a/b/c/
		root := t.TempDir()
		channelDir := filepath.Join(root, ".claude")
		if err := os.MkdirAll(channelDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		nested := filepath.Join(root, "a", "b", "c")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}

		got, err := FindProjectRoot(nested)
		if err != nil {
			t.Fatalf("FindProjectRoot(%q) returned error: %v", nested, err)
		}
		if got != root {
			t.Errorf("FindProjectRoot(%q) = %q, want %q", nested, got, root)
		}
	})

	t.Run("finds_claude_dir_at_start", func(t *testing.T) {
		root := t.TempDir()
		channelDir := filepath.Join(root, ".claude")
		if err := os.MkdirAll(channelDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}

		got, err := FindProjectRoot(root)
		if err != nil {
			t.Fatalf("FindProjectRoot(%q) returned error: %v", root, err)
		}
		if got != root {
			t.Errorf("FindProjectRoot(%q) = %q, want %q", root, got, root)
		}
	})

	t.Run("finds_knossos_dir", func(t *testing.T) {
		// Create a temp tree with only .knossos/ (no channel dir)
		root := t.TempDir()
		knossosDir := filepath.Join(root, ".knossos")
		if err := os.MkdirAll(knossosDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		nested := filepath.Join(root, "a", "b")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}

		got, err := FindProjectRoot(nested)
		if err != nil {
			t.Fatalf("FindProjectRoot(%q) returned error: %v", nested, err)
		}
		if got != root {
			t.Errorf("FindProjectRoot(%q) = %q, want %q", nested, got, root)
		}
	})

	t.Run("prefers_knossos_over_channels", func(t *testing.T) {
		// Both .knossos/ and channel dir exist — .knossos/ is checked first (platform dir)
		root := t.TempDir()
		os.MkdirAll(filepath.Join(root, ".claude"), 0755)
		os.MkdirAll(filepath.Join(root, ".knossos"), 0755)

		got, err := FindProjectRoot(root)
		if err != nil {
			t.Fatalf("FindProjectRoot(%q) returned error: %v", root, err)
		}
		if got != root {
			t.Errorf("FindProjectRoot(%q) = %q, want %q", root, got, root)
		}
	})

	t.Run("finds_gemini_dir", func(t *testing.T) {
		// Create a temp tree with only .gemini/ (no CC channel dir or .knossos/)
		root := t.TempDir()
		geminiDir := filepath.Join(root, ".gemini")
		if err := os.MkdirAll(geminiDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		nested := filepath.Join(root, "a", "b")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}

		got, err := FindProjectRoot(nested)
		if err != nil {
			t.Fatalf("FindProjectRoot(%q) returned error: %v", nested, err)
		}
		if got != root {
			t.Errorf("FindProjectRoot(%q) = %q, want %q", nested, got, root)
		}
	})

	t.Run("error_no_recognized_dir", func(t *testing.T) {
		// Temp dir with no .knossos/, channel dir, or .gemini/ anywhere in its ancestry
		isolated := t.TempDir()
		_, err := FindProjectRoot(isolated)
		if err == nil {
			t.Error("FindProjectRoot() should return error when no recognized directory exists")
		}
	})

	t.Run("file_not_dir_ignored", func(t *testing.T) {
		// .claude exists as a file, not a directory -- should not match
		root := t.TempDir()
		claudePath := filepath.Join(root, ".claude")
		if err := os.WriteFile(claudePath, []byte("not a dir"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := FindProjectRoot(root)
		if err == nil {
			t.Error("FindProjectRoot() should return error when .claude is a file, not a directory")
		}
	})
}

// TestEnsureDir verifies directory creation semantics.
func TestEnsureDir(t *testing.T) {
	t.Run("creates_new_directory", func(t *testing.T) {
		tmp := t.TempDir()
		target := filepath.Join(tmp, "a", "b", "c")

		if err := EnsureDir(target); err != nil {
			t.Fatalf("EnsureDir(%q) returned error: %v", target, err)
		}

		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("Stat(%q) returned error: %v", target, err)
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", target)
		}
	})

	t.Run("noop_existing_directory", func(t *testing.T) {
		tmp := t.TempDir()

		// Should succeed without error on existing dir
		if err := EnsureDir(tmp); err != nil {
			t.Fatalf("EnsureDir(%q) returned error on existing dir: %v", tmp, err)
		}
	})

	t.Run("error_on_file_conflict", func(t *testing.T) {
		tmp := t.TempDir()
		blocker := filepath.Join(tmp, "blocker")
		if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		// Trying to create a dir where a file already exists should fail
		target := filepath.Join(blocker, "subdir")
		if err := EnsureDir(target); err == nil {
			t.Errorf("EnsureDir(%q) should fail when parent is a file", target)
		}
	})
}

// TestReadActiveRite verifies reading the ACTIVE_RITE file.
func TestReadActiveRite(t *testing.T) {
	t.Run("returns_trimmed_content", func(t *testing.T) {
		root := t.TempDir()
		r := NewResolver(root)

		knossosDir := filepath.Join(root, ".knossos")
		if err := os.MkdirAll(knossosDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(r.ActiveRiteFile(), []byte("  10x-dev\n"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		got := r.ReadActiveRite()
		if got != "10x-dev" {
			t.Errorf("ReadActiveRite() = %q, want %q", got, "10x-dev")
		}
	})

	t.Run("returns_empty_on_missing_file", func(t *testing.T) {
		root := t.TempDir()
		r := NewResolver(root)

		got := r.ReadActiveRite()
		if got != "" {
			t.Errorf("ReadActiveRite() = %q, want empty string", got)
		}
	})

	t.Run("returns_empty_string_for_empty_file", func(t *testing.T) {
		root := t.TempDir()
		r := NewResolver(root)

		knossosDir := filepath.Join(root, ".knossos")
		if err := os.MkdirAll(knossosDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(r.ActiveRiteFile(), []byte(""), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		got := r.ReadActiveRite()
		if got != "" {
			t.Errorf("ReadActiveRite() = %q, want empty string", got)
		}
	})
}

// TestXDGDirectories verifies that XDG directory functions return paths
// consistent with the xdg library's current values.
func TestXDGDirectories(t *testing.T) {
	tests := []struct {
		name    string
		got     string
		wantDir string // expected XDG base
	}{
		{"ConfigDir", ConfigDir(), xdg.ConfigHome},
		{"StateDir", StateDir(), xdg.StateHome},
		{"CacheDir", CacheDir(), xdg.CacheHome},
		{"DataDir", DataDir(), xdg.DataHome},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := filepath.Join(tt.wantDir, "knossos")
			if tt.got != want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, want)
			}
		})
	}
}

// TestXDGDirectories_EnvOverride verifies that XDG directory functions
// respect environment variable overrides after xdg.Reload().
func TestXDGDirectories_EnvOverride(t *testing.T) {
	// Save original values for restoration
	origConfig := xdg.ConfigHome
	origState := xdg.StateHome
	origCache := xdg.CacheHome
	origData := xdg.DataHome
	t.Cleanup(func() {
		xdg.ConfigHome = origConfig
		xdg.StateHome = origState
		xdg.CacheHome = origCache
		xdg.DataHome = origData
	})

	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "state"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	xdg.Reload()

	tests := []struct {
		name    string
		fn      func() string
		wantSub string // expected substring in path
	}{
		{"ConfigDir", ConfigDir, filepath.Join(tmp, "config", "knossos")},
		{"StateDir", StateDir, filepath.Join(tmp, "state", "knossos")},
		{"CacheDir", CacheDir, filepath.Join(tmp, "cache", "knossos")},
		{"DataDir", DataDir, filepath.Join(tmp, "data", "knossos")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.wantSub {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.wantSub)
			}
		})
	}
}

// TestUserRitesDir verifies the user rites directory path.
func TestUserRitesDir(t *testing.T) {
	got := UserRitesDir()
	want := filepath.Join(DataDir(), "rites")
	if got != want {
		t.Errorf("UserRitesDir() = %q, want %q", got, want)
	}
}

// TestConfigFile verifies the config file path construction.
func TestConfigFile(t *testing.T) {
	got := ConfigFile("settings.yaml")
	want := filepath.Join(ConfigDir(), "settings.yaml")
	if got != want {
		t.Errorf("ConfigFile(%q) = %q, want %q", "settings.yaml", got, want)
	}
}

// TestEnsureConfigDir verifies config directory creation.
func TestEnsureConfigDir(t *testing.T) {
	// Save and override XDG config to use temp dir
	origConfig := xdg.ConfigHome
	t.Cleanup(func() { xdg.ConfigHome = origConfig })

	tmp := t.TempDir()
	xdg.ConfigHome = filepath.Join(tmp, "xdg-config")

	if err := EnsureConfigDir(); err != nil {
		t.Fatalf("EnsureConfigDir() returned error: %v", err)
	}

	info, err := os.Stat(ConfigDir())
	if err != nil {
		t.Fatalf("Stat(%q) returned error: %v", ConfigDir(), err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", ConfigDir())
	}
}

// TestEnsureStateDir verifies state directory creation.
func TestEnsureStateDir(t *testing.T) {
	origState := xdg.StateHome
	t.Cleanup(func() { xdg.StateHome = origState })

	tmp := t.TempDir()
	xdg.StateHome = filepath.Join(tmp, "xdg-state")

	if err := EnsureStateDir(); err != nil {
		t.Fatalf("EnsureStateDir() returned error: %v", err)
	}

	info, err := os.Stat(StateDir())
	if err != nil {
		t.Fatalf("Stat(%q) returned error: %v", StateDir(), err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", StateDir())
	}
}

// TestUserLevelPaths verifies user-level resource path functions.
func TestUserLevelPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"UserChannelDir/claude", UserChannelDir("claude"), filepath.Join(homeDir, ".claude")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestUserLevelForChannelPaths verifies the channel-parameterized ForChannel variants.
func TestUserLevelForChannelPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	tests := []struct {
		name    string
		channel string
		got     string
		want    string
	}{
		// Claude channel (default behavior)
		{"UserAgentsDirForChannel/claude", "claude", UserAgentsDirForChannel("claude"), filepath.Join(homeDir, ".claude", "agents")},
		{"UserSkillsDirForChannel/claude", "claude", UserSkillsDirForChannel("claude"), filepath.Join(homeDir, ".claude", "skills")},
		{"UserCommandsDirForChannel/claude", "claude", UserCommandsDirForChannel("claude"), filepath.Join(homeDir, ".claude", "commands")},
		{"UserHooksDirForChannel/claude", "claude", UserHooksDirForChannel("claude"), filepath.Join(homeDir, ".claude", "hooks")},
		{"UserProvenanceManifestForChannel/claude", "claude", UserProvenanceManifestForChannel("claude"), filepath.Join(homeDir, ".claude", "USER_PROVENANCE_MANIFEST.yaml")},
		{"OrgProvenanceManifestForChannel/claude", "claude", OrgProvenanceManifestForChannel("claude"), filepath.Join(homeDir, ".claude", "ORG_PROVENANCE_MANIFEST.yaml")},
		// Gemini channel
		{"UserAgentsDirForChannel/gemini", "gemini", UserAgentsDirForChannel("gemini"), filepath.Join(homeDir, ".gemini", "agents")},
		{"UserSkillsDirForChannel/gemini", "gemini", UserSkillsDirForChannel("gemini"), filepath.Join(homeDir, ".gemini", "skills")},
		{"UserCommandsDirForChannel/gemini", "gemini", UserCommandsDirForChannel("gemini"), filepath.Join(homeDir, ".gemini", "commands")},
		{"UserHooksDirForChannel/gemini", "gemini", UserHooksDirForChannel("gemini"), filepath.Join(homeDir, ".gemini", "hooks")},
		{"UserProvenanceManifestForChannel/gemini", "gemini", UserProvenanceManifestForChannel("gemini"), filepath.Join(homeDir, ".gemini", "USER_PROVENANCE_MANIFEST.yaml")},
		{"OrgProvenanceManifestForChannel/gemini", "gemini", OrgProvenanceManifestForChannel("gemini"), filepath.Join(homeDir, ".gemini", "ORG_PROVENANCE_MANIFEST.yaml")},
		// Empty channel defaults to claude
		{"UserAgentsDirForChannel/empty", "", UserAgentsDirForChannel(""), filepath.Join(homeDir, ".claude", "agents")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestForChannelBackwardCompat verifies deprecated functions return same result as ForChannel("claude").
func TestForChannelBackwardCompat(t *testing.T) {
	tests := []struct {
		name       string
		deprecated string
		forChannel string
	}{
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.deprecated != tt.forChannel {
				t.Errorf("%s() = %q, ForChannel(\"claude\") = %q -- backward compat broken",
					tt.name, tt.deprecated, tt.forChannel)
			}
		})
	}
}
