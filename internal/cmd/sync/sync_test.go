package sync

import (
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize"
)

// --- Command metadata ---

func TestSyncCmd_Name(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	if cmd.Use != "sync" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "sync")
	}
}

func TestSyncCmd_ShortDescription(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	if cmd.Short == "" {
		t.Error("cmd.Short is empty, want non-empty description")
	}
}

func TestSyncCmd_LongDescription(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	if cmd.Long == "" {
		t.Error("cmd.Long is empty, want non-empty long description")
	}
}

func TestSyncCmd_LongDescriptionMentionsScopes(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	for _, keyword := range []string{"rite", "user", "org"} {
		if !strings.Contains(strings.ToLower(cmd.Long), keyword) {
			t.Errorf("cmd.Long does not mention scope %q", keyword)
		}
	}
}

// --- NeedsProject annotation ---

func TestSyncCmd_NeedsProjectFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	if common.NeedsProject(cmd) {
		t.Error("sync command should have needsProject=false (user scope works without project)")
	}
}

// --- Flag existence ---

func TestSyncCmd_FlagScope_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("scope")
	if f == nil {
		t.Fatal("sync command missing --scope flag")
	}
}

func TestSyncCmd_FlagRite_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("sync command missing --rite flag")
	}
}

func TestSyncCmd_FlagResource_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("resource")
	if f == nil {
		t.Fatal("sync command missing --resource flag")
	}
}

func TestSyncCmd_FlagDryRun_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("sync command missing --dry-run flag")
	}
}

func TestSyncCmd_FlagKeepOrphans_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("keep-orphans")
	if f == nil {
		t.Fatal("sync command missing --keep-orphans flag")
	}
}

func TestSyncCmd_FlagOrg_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("org")
	if f == nil {
		t.Fatal("sync command missing --org flag")
	}
}

func TestSyncCmd_FlagSource_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("source")
	if f == nil {
		t.Fatal("sync command missing --source flag")
	}
}

func TestSyncCmd_FlagRecover_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("recover")
	if f == nil {
		t.Fatal("sync command missing --recover flag")
	}
}

func TestSyncCmd_FlagOverwriteDiverged_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("overwrite-diverged")
	if f == nil {
		t.Fatal("sync command missing --overwrite-diverged flag")
	}
}

func TestSyncCmd_FlagSoft_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("soft")
	if f == nil {
		t.Fatal("sync command missing --soft flag")
	}
}

func TestSyncCmd_FlagBudget_Exists(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("budget")
	if f == nil {
		t.Fatal("sync command missing --budget flag")
	}
}

// --- Flag defaults ---

func TestSyncCmd_FlagScope_DefaultEmpty(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("scope")
	if f == nil {
		t.Fatal("--scope flag not found")
	}
	if f.DefValue != "" {
		t.Errorf("--scope default = %q, want %q (empty = all)", f.DefValue, "")
	}
}

func TestSyncCmd_FlagDryRun_DefaultFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("--dry-run flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--dry-run default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_FlagKeepOrphans_DefaultFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("keep-orphans")
	if f == nil {
		t.Fatal("--keep-orphans flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--keep-orphans default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_FlagRecover_DefaultFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("recover")
	if f == nil {
		t.Fatal("--recover flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--recover default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_FlagSoft_DefaultFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("soft")
	if f == nil {
		t.Fatal("--soft flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--soft default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_FlagBudget_DefaultFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("budget")
	if f == nil {
		t.Fatal("--budget flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--budget default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_FlagOverwriteDiverged_DefaultFalse(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("overwrite-diverged")
	if f == nil {
		t.Fatal("--overwrite-diverged flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--overwrite-diverged default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_FlagRite_DefaultEmpty(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("--rite flag not found")
	}
	if f.DefValue != "" {
		t.Errorf("--rite default = %q, want %q (empty = use ACTIVE_RITE)", f.DefValue, "")
	}
}

func TestSyncCmd_FlagResource_DefaultEmpty(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	f := cmd.Flags().Lookup("resource")
	if f == nil {
		t.Fatal("--resource flag not found")
	}
	if f.DefValue != "" {
		t.Errorf("--resource default = %q, want %q (empty = all resources)", f.DefValue, "")
	}
}

// --- Scope validation (table-driven) ---

func TestSyncScope_IsValid(t *testing.T) {
	tests := []struct {
		scope string
		valid bool
	}{
		{"", true},   // empty = ScopeAll
		{"all", true},
		{"rite", true},
		{"org", true},
		{"user", true},
		{"RITE", false},
		{"User", false},
		{"project", false},
		{"global", false},
		{"invalid-scope", false},
	}
	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			scope := materialize.SyncScope(tt.scope)
			// Empty maps to ScopeAll in the command; test the underlying type directly.
			if tt.scope == "" {
				scope = materialize.ScopeAll
			}
			got := scope.IsValid()
			if got != tt.valid {
				t.Errorf("SyncScope(%q).IsValid() = %v, want %v", tt.scope, got, tt.valid)
			}
		})
	}
}

// TestSyncCmd_InvalidScope_ReturnsError verifies that the RunE function returns
// an error with a clear message for invalid --scope values.
func TestSyncCmd_InvalidScope_ReturnsError(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	cmd.SetArgs([]string{"--scope=bogus"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --scope, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --scope") {
		t.Errorf("error message should mention 'invalid --scope', got: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error message should include the invalid value 'bogus', got: %q", err.Error())
	}
}

func TestSyncCmd_InvalidScope_MentionsValidValues(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	cmd.SetArgs([]string{"--scope=unknown"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --scope, got nil")
	}
	// The error should mention valid options.
	for _, valid := range []string{"rite", "user", "all"} {
		if !strings.Contains(err.Error(), valid) {
			t.Errorf("error message should mention valid scope %q, got: %q", valid, err.Error())
		}
	}
}

// --- Resource validation (table-driven) ---

func TestSyncResource_IsValid(t *testing.T) {
	tests := []struct {
		resource string
		valid    bool
	}{
		{"", true},       // empty = ResourceAll
		{"agents", true},
		{"mena", true},
		{"hooks", true},
		{"AGENTS", false},
		{"Mena", false},
		{"commands", false},
		{"skills", false},
		{"all", false},
		{"invalid-resource", false},
	}
	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			res := materialize.SyncResource(tt.resource)
			got := res.IsValid()
			if got != tt.valid {
				t.Errorf("SyncResource(%q).IsValid() = %v, want %v", tt.resource, got, tt.valid)
			}
		})
	}
}

// TestSyncCmd_InvalidResource_ReturnsError verifies that the RunE function returns
// an error with a clear message for invalid --resource values.
func TestSyncCmd_InvalidResource_ReturnsError(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	cmd.SetArgs([]string{"--resource=commands"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --resource, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --resource") {
		t.Errorf("error message should mention 'invalid --resource', got: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "commands") {
		t.Errorf("error message should include the invalid value 'commands', got: %q", err.Error())
	}
}

func TestSyncCmd_InvalidResource_MentionsValidValues(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""
	cmd := NewSyncCmd(&output, &verbose, &projectDir)
	cmd.SetArgs([]string{"--resource=invalid"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --resource, got nil")
	}
	// The error should mention valid options.
	for _, valid := range []string{"agents", "mena", "hooks"} {
		if !strings.Contains(err.Error(), valid) {
			t.Errorf("error message should mention valid resource %q, got: %q", valid, err.Error())
		}
	}
}

// --- Valid scope and resource values are accepted (RunE proceeds past validation) ---

// TestSyncCmd_ValidScope_PassesValidation verifies that valid --scope values
// pass the scope validation check (they may fail later for other reasons,
// but should not produce an "invalid --scope" error).
func TestSyncCmd_ValidScope_PassesValidation(t *testing.T) {
	validScopes := []string{"rite", "org", "user", "all"}
	for _, scope := range validScopes {
		t.Run(scope, func(t *testing.T) {
			output := "text"
			verbose := false
			projectDir := ""
			cmd := NewSyncCmd(&output, &verbose, &projectDir)
			cmd.SetArgs([]string{"--scope=" + scope})
			err := cmd.Execute()
			// May error for other reasons (no project, no embedded rites, etc.)
			// but must NOT produce an "invalid --scope" error.
			if err != nil && strings.Contains(err.Error(), "invalid --scope") {
				t.Errorf("valid scope %q produced invalid --scope error: %v", scope, err)
			}
		})
	}
}

// TestSyncCmd_ValidResource_PassesValidation verifies that valid --resource values
// pass the resource validation check.
func TestSyncCmd_ValidResource_PassesValidation(t *testing.T) {
	validResources := []string{"agents", "mena", "hooks"}
	for _, resource := range validResources {
		t.Run(resource, func(t *testing.T) {
			output := "text"
			verbose := false
			projectDir := ""
			cmd := NewSyncCmd(&output, &verbose, &projectDir)
			cmd.SetArgs([]string{"--scope=user", "--resource=" + resource})
			err := cmd.Execute()
			// May error for other reasons, but must NOT produce an "invalid --resource" error.
			if err != nil && strings.Contains(err.Error(), "invalid --resource") {
				t.Errorf("valid resource %q produced invalid --resource error: %v", resource, err)
			}
		})
	}
}

// --- formatNum helper ---

func TestFormatNum_BelowThousand(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
	}
	for _, tt := range tests {
		got := formatNum(tt.n)
		if got != tt.want {
			t.Errorf("formatNum(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatNum_AtThousand(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{1000, "1,000"},
		{1001, "1,001"},
		{9999, "9,999"},
		{10000, "10,000"},
		{100000, "100,000"},
	}
	for _, tt := range tests {
		got := formatNum(tt.n)
		if got != tt.want {
			t.Errorf("formatNum(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
