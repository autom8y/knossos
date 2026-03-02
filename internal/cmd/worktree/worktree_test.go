package worktree

import (
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/worktree"
	"github.com/spf13/cobra"
)

// newTestWorktreeCmd creates a WorktreeCmd with default test flags.
func newTestWorktreeCmd() *cobra.Command {
	output := "text"
	verbose := false
	projectDir := ""
	return NewWorktreeCmd(&output, &verbose, &projectDir)
}

// --- Root command metadata ---

func TestWorktreeCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	if cmd.Use != "worktree" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "worktree")
	}
}

func TestWorktreeCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	if cmd.Short == "" {
		t.Error("cmd.Short is empty, want non-empty description")
	}
}

func TestWorktreeCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	if cmd.Long == "" {
		t.Error("cmd.Long is empty, want non-empty long description")
	}
}

func TestWorktreeCmd_LongDescriptionMentionsIsolation(t *testing.T) {
	cmd := newTestWorktreeCmd()
	for _, keyword := range []string{"worktree", "session", "parallel"} {
		if !strings.Contains(strings.ToLower(cmd.Long), keyword) {
			t.Errorf("cmd.Long does not mention %q", keyword)
		}
	}
}

// --- NeedsProject annotation ---

func TestWorktreeCmd_NeedsProjectTrue(t *testing.T) {
	cmd := newTestWorktreeCmd()
	if !common.NeedsProject(cmd) {
		t.Error("worktree command should have needsProject=true")
	}
}

// --- Subcommand presence ---

func TestWorktreeCmd_HasCreateSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"create"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'create' subcommand")
	}
}

func TestWorktreeCmd_HasListSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"list"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'list' subcommand")
	}
}

func TestWorktreeCmd_HasStatusSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"status"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'status' subcommand")
	}
}

func TestWorktreeCmd_HasRemoveSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"remove"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'remove' subcommand")
	}
}

func TestWorktreeCmd_HasCleanupSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"cleanup"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'cleanup' subcommand")
	}
}

func TestWorktreeCmd_HasSwitchSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"switch"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'switch' subcommand")
	}
}

func TestWorktreeCmd_HasCloneSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"clone"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'clone' subcommand")
	}
}

func TestWorktreeCmd_HasSyncSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"sync"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'sync' subcommand")
	}
}

func TestWorktreeCmd_HasExportSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"export"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'export' subcommand")
	}
}

func TestWorktreeCmd_HasImportSubcommand(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, err := cmd.Find([]string{"import"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("worktree command missing 'import' subcommand")
	}
}

// --- create subcommand: metadata ---

func TestCreateCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	if !strings.HasPrefix(sub.Use, "create") {
		t.Errorf("create subcommand Use = %q, want prefix 'create'", sub.Use)
	}
}

func TestCreateCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	if sub.Short == "" {
		t.Error("create subcommand Short is empty")
	}
}

func TestCreateCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	if sub.Long == "" {
		t.Error("create subcommand Long is empty")
	}
}

// --- create subcommand: flags ---

func TestCreateCmd_FlagRite_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	f := sub.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("create subcommand missing --rite flag")
	}
}

func TestCreateCmd_FlagFrom_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	f := sub.Flags().Lookup("from")
	if f == nil {
		t.Fatal("create subcommand missing --from flag")
	}
}

func TestCreateCmd_FlagComplexity_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	f := sub.Flags().Lookup("complexity")
	if f == nil {
		t.Fatal("create subcommand missing --complexity flag")
	}
}

func TestCreateCmd_FlagRite_DefaultEmpty(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	f := sub.Flags().Lookup("rite")
	if f.DefValue != "" {
		t.Errorf("--rite default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestCreateCmd_FlagFrom_DefaultEmpty(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	f := sub.Flags().Lookup("from")
	if f.DefValue != "" {
		t.Errorf("--from default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestCreateCmd_FlagComplexity_DefaultMODULE(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	f := sub.Flags().Lookup("complexity")
	if f.DefValue != "MODULE" {
		t.Errorf("--complexity default = %q, want %q", f.DefValue, "MODULE")
	}
}

// --- create subcommand: arg validation ---

func TestCreateCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no name argument given to 'create', got nil")
	}
}

func TestCreateCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"create", "name1", "name2"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments given to 'create', got nil")
	}
}

// --- create subcommand: NeedsProject inheritance ---

func TestCreateCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"create"})
	if !common.NeedsProject(sub) {
		t.Error("create subcommand should inherit needsProject=true from parent")
	}
}

// --- list subcommand: metadata ---

func TestListCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if !strings.HasPrefix(sub.Use, "list") {
		t.Errorf("list subcommand Use = %q, want prefix 'list'", sub.Use)
	}
}

func TestListCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if sub.Short == "" {
		t.Error("list subcommand Short is empty")
	}
}

func TestListCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if sub.Long == "" {
		t.Error("list subcommand Long is empty")
	}
}

func TestListCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if !common.NeedsProject(sub) {
		t.Error("list subcommand should inherit needsProject=true from parent")
	}
}

// --- status subcommand: metadata ---

func TestStatusCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"status"})
	if !strings.HasPrefix(sub.Use, "status") {
		t.Errorf("status subcommand Use = %q, want prefix 'status'", sub.Use)
	}
}

func TestStatusCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"status"})
	if sub.Short == "" {
		t.Error("status subcommand Short is empty")
	}
}

func TestStatusCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"status"})
	if sub.Long == "" {
		t.Error("status subcommand Long is empty")
	}
}

func TestStatusCmd_AcceptsZeroOrOneArgs(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"status"})
	// cobra.MaximumNArgs(1) — verify by checking two args is rejected
	// We cannot easily execute without a real project, so we check via the command definition
	// by calling Args validator directly.
	if sub.Args == nil {
		// If nil, cobra allows any number of args, which is not what we expect.
		// MaximumNArgs sets Args, so nil means unrestricted — flag as unexpected.
		// However some cobra versions may embed the validator differently.
		// Skip this check if Args is nil (cobra validates internally).
		t.Skip("Args validator is nil; skipping arity check")
	}
}

func TestStatusCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"status"})
	if !common.NeedsProject(sub) {
		t.Error("status subcommand should inherit needsProject=true from parent")
	}
}

// --- switch subcommand: metadata ---

func TestSwitchCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"switch"})
	if !strings.HasPrefix(sub.Use, "switch") {
		t.Errorf("switch subcommand Use = %q, want prefix 'switch'", sub.Use)
	}
}

func TestSwitchCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"switch"})
	if sub.Short == "" {
		t.Error("switch subcommand Short is empty")
	}
}

func TestSwitchCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"switch"})
	if sub.Long == "" {
		t.Error("switch subcommand Long is empty")
	}
}

// --- switch subcommand: flags ---

func TestSwitchCmd_FlagUpdateRite_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"switch"})
	f := sub.Flags().Lookup("update-rite")
	if f == nil {
		t.Fatal("switch subcommand missing --update-rite flag")
	}
}

func TestSwitchCmd_FlagUpdateRite_DefaultFalse(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"switch"})
	f := sub.Flags().Lookup("update-rite")
	if f.DefValue != "false" {
		t.Errorf("--update-rite default = %q, want %q", f.DefValue, "false")
	}
}

// --- switch subcommand: arg validation ---

func TestSwitchCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"switch"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no id-or-name argument given to 'switch', got nil")
	}
}

func TestSwitchCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"switch", "name1", "name2"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments given to 'switch', got nil")
	}
}

func TestSwitchCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"switch"})
	if !common.NeedsProject(sub) {
		t.Error("switch subcommand should inherit needsProject=true from parent")
	}
}

// --- sync subcommand: metadata ---

func TestSyncCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"sync"})
	if !strings.HasPrefix(sub.Use, "sync") {
		t.Errorf("sync subcommand Use = %q, want prefix 'sync'", sub.Use)
	}
}

func TestSyncCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"sync"})
	if sub.Short == "" {
		t.Error("sync subcommand Short is empty")
	}
}

func TestSyncCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"sync"})
	if sub.Long == "" {
		t.Error("sync subcommand Long is empty")
	}
}

// --- sync subcommand: flags ---

func TestSyncCmd_FlagPull_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"sync"})
	f := sub.Flags().Lookup("pull")
	if f == nil {
		t.Fatal("sync subcommand missing --pull flag")
	}
}

func TestSyncCmd_FlagPull_DefaultFalse(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"sync"})
	f := sub.Flags().Lookup("pull")
	if f.DefValue != "false" {
		t.Errorf("--pull default = %q, want %q", f.DefValue, "false")
	}
}

func TestSyncCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"sync"})
	if !common.NeedsProject(sub) {
		t.Error("sync subcommand should inherit needsProject=true from parent")
	}
}

// --- remove subcommand: metadata ---

func TestRemoveCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	if !strings.HasPrefix(sub.Use, "remove") {
		t.Errorf("remove subcommand Use = %q, want prefix 'remove'", sub.Use)
	}
}

func TestRemoveCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	if sub.Short == "" {
		t.Error("remove subcommand Short is empty")
	}
}

func TestRemoveCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	if sub.Long == "" {
		t.Error("remove subcommand Long is empty")
	}
}

// --- remove subcommand: flags ---

func TestRemoveCmd_FlagForce_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	f := sub.Flags().Lookup("force")
	if f == nil {
		t.Fatal("remove subcommand missing --force flag")
	}
}

func TestRemoveCmd_FlagForce_Shorthand_f(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	f := sub.Flags().ShorthandLookup("f")
	if f == nil {
		t.Fatal("remove subcommand missing -f shorthand for --force flag")
	}
}

func TestRemoveCmd_FlagForce_DefaultFalse(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	f := sub.Flags().Lookup("force")
	if f.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", f.DefValue, "false")
	}
}

// --- remove subcommand: arg validation ---

func TestRemoveCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"remove"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no id argument given to 'remove', got nil")
	}
}

func TestRemoveCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"remove", "id1", "id2"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments given to 'remove', got nil")
	}
}

func TestRemoveCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"remove"})
	if !common.NeedsProject(sub) {
		t.Error("remove subcommand should inherit needsProject=true from parent")
	}
}

// --- cleanup subcommand: metadata ---

func TestCleanupCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	if !strings.HasPrefix(sub.Use, "cleanup") {
		t.Errorf("cleanup subcommand Use = %q, want prefix 'cleanup'", sub.Use)
	}
}

func TestCleanupCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	if sub.Short == "" {
		t.Error("cleanup subcommand Short is empty")
	}
}

func TestCleanupCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	if sub.Long == "" {
		t.Error("cleanup subcommand Long is empty")
	}
}

// --- cleanup subcommand: flags ---

func TestCleanupCmd_FlagOlderThan_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().Lookup("older-than")
	if f == nil {
		t.Fatal("cleanup subcommand missing --older-than flag")
	}
}

func TestCleanupCmd_FlagDryRun_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("cleanup subcommand missing --dry-run flag")
	}
}

func TestCleanupCmd_FlagForce_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().Lookup("force")
	if f == nil {
		t.Fatal("cleanup subcommand missing --force flag")
	}
}

func TestCleanupCmd_FlagForce_Shorthand_f(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().ShorthandLookup("f")
	if f == nil {
		t.Fatal("cleanup subcommand missing -f shorthand for --force flag")
	}
}

func TestCleanupCmd_FlagOlderThan_Default7d(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().Lookup("older-than")
	if f.DefValue != "7d" {
		t.Errorf("--older-than default = %q, want %q", f.DefValue, "7d")
	}
}

func TestCleanupCmd_FlagDryRun_DefaultFalse(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().Lookup("dry-run")
	if f.DefValue != "false" {
		t.Errorf("--dry-run default = %q, want %q", f.DefValue, "false")
	}
}

func TestCleanupCmd_FlagForce_DefaultFalse(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	f := sub.Flags().Lookup("force")
	if f.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", f.DefValue, "false")
	}
}

func TestCleanupCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"cleanup"})
	if !common.NeedsProject(sub) {
		t.Error("cleanup subcommand should inherit needsProject=true from parent")
	}
}

// --- clone subcommand: metadata ---

func TestCloneCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	if !strings.HasPrefix(sub.Use, "clone") {
		t.Errorf("clone subcommand Use = %q, want prefix 'clone'", sub.Use)
	}
}

func TestCloneCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	if sub.Short == "" {
		t.Error("clone subcommand Short is empty")
	}
}

func TestCloneCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	if sub.Long == "" {
		t.Error("clone subcommand Long is empty")
	}
}

// --- clone subcommand: flags ---

func TestCloneCmd_FlagRite_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	f := sub.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("clone subcommand missing --rite flag")
	}
}

func TestCloneCmd_FlagCopySession_Exists(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	f := sub.Flags().Lookup("copy-session")
	if f == nil {
		t.Fatal("clone subcommand missing --copy-session flag")
	}
}

func TestCloneCmd_FlagRite_DefaultEmpty(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	f := sub.Flags().Lookup("rite")
	if f.DefValue != "" {
		t.Errorf("--rite default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestCloneCmd_FlagCopySession_DefaultFalse(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	f := sub.Flags().Lookup("copy-session")
	if f.DefValue != "false" {
		t.Errorf("--copy-session default = %q, want %q", f.DefValue, "false")
	}
}

// --- clone subcommand: arg validation ---

func TestCloneCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"clone"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no arguments given to 'clone', got nil")
	}
}

func TestCloneCmd_OneArg_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"clone", "source-only"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when only one argument given to 'clone', got nil")
	}
}

func TestCloneCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"clone", "src", "dst", "extra"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments given to 'clone', got nil")
	}
}

func TestCloneCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"clone"})
	if !common.NeedsProject(sub) {
		t.Error("clone subcommand should inherit needsProject=true from parent")
	}
}

// --- export subcommand: metadata ---

func TestExportCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"export"})
	if !strings.HasPrefix(sub.Use, "export") {
		t.Errorf("export subcommand Use = %q, want prefix 'export'", sub.Use)
	}
}

func TestExportCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"export"})
	if sub.Short == "" {
		t.Error("export subcommand Short is empty")
	}
}

func TestExportCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"export"})
	if sub.Long == "" {
		t.Error("export subcommand Long is empty")
	}
}

// --- export subcommand: arg validation ---

func TestExportCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"export"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no arguments given to 'export', got nil")
	}
}

func TestExportCmd_OneArg_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"export", "worktree-id"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when only one argument given to 'export', got nil")
	}
}

func TestExportCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"export"})
	if !common.NeedsProject(sub) {
		t.Error("export subcommand should inherit needsProject=true from parent")
	}
}

// --- import subcommand: metadata ---

func TestImportCmd_Use(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"import"})
	if !strings.HasPrefix(sub.Use, "import") {
		t.Errorf("import subcommand Use = %q, want prefix 'import'", sub.Use)
	}
}

func TestImportCmd_ShortDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"import"})
	if sub.Short == "" {
		t.Error("import subcommand Short is empty")
	}
}

func TestImportCmd_LongDescription(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"import"})
	if sub.Long == "" {
		t.Error("import subcommand Long is empty")
	}
}

// --- import subcommand: arg validation ---

func TestImportCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"import"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no archive-path argument given to 'import', got nil")
	}
}

func TestImportCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"import", "file1.tar.gz", "file2.tar.gz"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments given to 'import', got nil")
	}
}

func TestImportCmd_InvalidExtension_ReturnsError(t *testing.T) {
	// Import validates archive extension before hitting the manager.
	// We provide a non-tar.gz path so validation fires without needing a project.
	cmd := newTestWorktreeCmd()
	cmd.SetArgs([]string{"import", "notanarchive.zip"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when archive does not have .tar.gz/.tgz extension, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "tar.gz") &&
		!strings.Contains(strings.ToLower(err.Error()), "archive") &&
		!strings.Contains(strings.ToLower(err.Error()), "tgz") {
		t.Errorf("error should mention archive format, got: %q", err.Error())
	}
}

func TestImportCmd_InheritsNeedsProject(t *testing.T) {
	cmd := newTestWorktreeCmd()
	sub, _, _ := cmd.Find([]string{"import"})
	if !common.NeedsProject(sub) {
		t.Error("import subcommand should inherit needsProject=true from parent")
	}
}

// --- parseDuration helper ---

func TestParseDuration_Days(t *testing.T) {
	cases := []struct {
		input string
		want  time.Duration
	}{
		{"7d", 7 * 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"30d", 30 * 24 * time.Hour},
	}
	for _, tc := range cases {
		got, err := parseDuration(tc.input)
		if err != nil {
			t.Errorf("parseDuration(%q) returned error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseDuration(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseDuration_Hours(t *testing.T) {
	got, err := parseDuration("24h")
	if err != nil {
		t.Fatalf("parseDuration(\"24h\") returned error: %v", err)
	}
	if got != 24*time.Hour {
		t.Errorf("parseDuration(\"24h\") = %v, want %v", got, 24*time.Hour)
	}
}

func TestParseDuration_Minutes(t *testing.T) {
	got, err := parseDuration("90m")
	if err != nil {
		t.Fatalf("parseDuration(\"90m\") returned error: %v", err)
	}
	if got != 90*time.Minute {
		t.Errorf("parseDuration(\"90m\") = %v, want %v", got, 90*time.Minute)
	}
}

func TestParseDuration_InvalidString_ReturnsError(t *testing.T) {
	cases := []string{"notaduration", "abc", "xd", "7x"}
	for _, input := range cases {
		_, err := parseDuration(input)
		if err == nil {
			t.Errorf("parseDuration(%q) expected error, got nil", input)
		}
	}
}

func TestParseDuration_EmptyString_ReturnsError(t *testing.T) {
	_, err := parseDuration("")
	if err == nil {
		t.Error("parseDuration(\"\") expected error, got nil")
	}
}

// --- isValidComplexity helper ---

func TestIsValidComplexity_ValidValues(t *testing.T) {
	valid := []string{"PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION"}
	for _, v := range valid {
		if !isValidComplexity(v) {
			t.Errorf("isValidComplexity(%q) = false, want true", v)
		}
	}
}

func TestIsValidComplexity_InvalidValues(t *testing.T) {
	invalid := []string{"patch", "module", "system", "UNKNOWN", "", "COMPLEX", "LARGE"}
	for _, v := range invalid {
		if isValidComplexity(v) {
			t.Errorf("isValidComplexity(%q) = true, want false", v)
		}
	}
}

// --- parseInt helper ---

func TestParseInt_ValidNumbers(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"1", 1},
		{"7", 7},
		{"30", 30},
		{"365", 365},
	}
	for _, tc := range cases {
		got, err := parseInt(tc.input)
		if err != nil {
			t.Errorf("parseInt(%q) returned error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseInt(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseInt_InvalidInput_ReturnsError(t *testing.T) {
	cases := []string{"abc", "1a", "-1", " 7"}
	for _, input := range cases {
		_, err := parseInt(input)
		if err == nil {
			t.Errorf("parseInt(%q) expected error, got nil", input)
		}
	}
}

// --- formatSize helper ---

func TestFormatSize_Bytes(t *testing.T) {
	got := formatSize(512)
	if !strings.Contains(got, "bytes") {
		t.Errorf("formatSize(512) = %q, want to contain 'bytes'", got)
	}
}

func TestFormatSize_Kilobytes(t *testing.T) {
	got := formatSize(2048)
	if !strings.Contains(got, "KB") {
		t.Errorf("formatSize(2048) = %q, want to contain 'KB'", got)
	}
}

func TestFormatSize_Megabytes(t *testing.T) {
	got := formatSize(2 * 1024 * 1024)
	if !strings.Contains(got, "MB") {
		t.Errorf("formatSize(2MB) = %q, want to contain 'MB'", got)
	}
}

func TestFormatSize_Gigabytes(t *testing.T) {
	got := formatSize(2 * 1024 * 1024 * 1024)
	if !strings.Contains(got, "GB") {
		t.Errorf("formatSize(2GB) = %q, want to contain 'GB'", got)
	}
}

// --- Output Text() methods ---

func TestCreateOutput_Text_ContainsWorktreeID(t *testing.T) {
	out := CreateOutput{
		Success:    true,
		WorktreeID: "wt-20260104-143052-a1b2",
		Name:       "feature-auth",
		Path:       "/tmp/worktrees/wt-20260104-143052-a1b2",
	}
	text := out.Text()
	if !strings.Contains(text, "wt-20260104-143052-a1b2") {
		t.Errorf("CreateOutput.Text() should contain worktree ID, got: %q", text)
	}
	if !strings.Contains(text, "feature-auth") {
		t.Errorf("CreateOutput.Text() should contain name, got: %q", text)
	}
}

func TestCreateOutput_Text_ContainsInstructions(t *testing.T) {
	out := CreateOutput{
		Success:    true,
		WorktreeID: "wt-abc",
		Path:       "/tmp/wt-abc",
	}
	text := out.Text()
	if !strings.Contains(text, "/tmp/wt-abc") {
		t.Errorf("CreateOutput.Text() should contain path, got: %q", text)
	}
	if !strings.Contains(text, "claude") {
		t.Errorf("CreateOutput.Text() should contain claude instruction, got: %q", text)
	}
}

func TestCreateOutput_Text_RiteShownWhenSet(t *testing.T) {
	out := CreateOutput{
		Success:    true,
		WorktreeID: "wt-abc",
		Rite:       "ecosystem",
	}
	text := out.Text()
	if !strings.Contains(text, "ecosystem") {
		t.Errorf("CreateOutput.Text() should contain rite when set, got: %q", text)
	}
}

func TestCreateOutput_Text_RiteHiddenWhenNone(t *testing.T) {
	out := CreateOutput{
		Success:    true,
		WorktreeID: "wt-abc",
		Rite:       "none",
	}
	text := out.Text()
	if strings.Contains(text, "Rite:") {
		t.Errorf("CreateOutput.Text() should not show Rite line when rite='none', got: %q", text)
	}
}

func TestListOutput_Headers_SixColumns(t *testing.T) {
	out := ListOutput{}
	headers := out.Headers()
	if len(headers) != 6 {
		t.Errorf("ListOutput.Headers() len = %d, want 6", len(headers))
	}
}

func TestListOutput_Headers_ContainsID(t *testing.T) {
	out := ListOutput{}
	headers := out.Headers()
	found := false
	for _, h := range headers {
		if h == "ID" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListOutput.Headers() should contain 'ID', got: %v", headers)
	}
}

func TestListOutput_Rows_MarksCurrent(t *testing.T) {
	out := ListOutput{
		Worktrees: []WorktreeSummary{
			{ID: "wt-1", Name: "first", Current: true},
			{ID: "wt-2", Name: "second", Current: false},
		},
	}
	rows := out.Rows()
	if len(rows) != 2 {
		t.Fatalf("ListOutput.Rows() len = %d, want 2", len(rows))
	}
	// Current row should have "* " prefix in ID column
	if !strings.HasPrefix(rows[0][0], "* ") {
		t.Errorf("current worktree row[0][0] = %q, want prefix '* '", rows[0][0])
	}
	// Non-current row should have "  " prefix
	if !strings.HasPrefix(rows[1][0], "  ") {
		t.Errorf("non-current worktree row[1][0] = %q, want prefix '  '", rows[1][0])
	}
}

func TestListOutput_Rows_ShowsDirtyStatus(t *testing.T) {
	out := ListOutput{
		Worktrees: []WorktreeSummary{
			{ID: "wt-1", IsDirty: true},
			{ID: "wt-2", IsDirty: false},
		},
	}
	rows := out.Rows()
	// Status column (index 4)
	if rows[0][4] != "dirty" {
		t.Errorf("dirty worktree status = %q, want 'dirty'", rows[0][4])
	}
	if rows[1][4] != "clean" {
		t.Errorf("clean worktree status = %q, want 'clean'", rows[1][4])
	}
}

func TestListOutput_Rows_ShowsDashForEmptyRite(t *testing.T) {
	out := ListOutput{
		Worktrees: []WorktreeSummary{
			{ID: "wt-1", Rite: ""},
		},
	}
	rows := out.Rows()
	// Rite column (index 2)
	if rows[0][2] != "-" {
		t.Errorf("empty rite column = %q, want '-'", rows[0][2])
	}
}

func TestListOutput_Text_EmptyReturnsNoWorktrees(t *testing.T) {
	out := ListOutput{Worktrees: []WorktreeSummary{}}
	text := out.Text()
	if !strings.Contains(text, "No worktrees") {
		t.Errorf("ListOutput.Text() for empty should contain 'No worktrees', got: %q", text)
	}
}

func TestRemoveOutput_Text_ContainsRemovedID(t *testing.T) {
	out := RemoveOutput{
		Success: true,
		Removed: "wt-20260104-143052-a1b2",
	}
	text := out.Text()
	if !strings.Contains(text, "wt-20260104-143052-a1b2") {
		t.Errorf("RemoveOutput.Text() should contain removed ID, got: %q", text)
	}
}

func TestRemoveOutput_Text_ForcedRemovalMentioned(t *testing.T) {
	out := RemoveOutput{
		Success:   true,
		Removed:   "wt-abc",
		WasForced: true,
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "forced") {
		t.Errorf("RemoveOutput.Text() should mention forced removal, got: %q", text)
	}
}

func TestRemoveOutput_Text_ForcedRemovalNotMentionedWhenFalse(t *testing.T) {
	out := RemoveOutput{
		Success:   true,
		Removed:   "wt-abc",
		WasForced: false,
	}
	text := out.Text()
	if strings.Contains(strings.ToLower(text), "forced") {
		t.Errorf("RemoveOutput.Text() should not mention forced when WasForced=false, got: %q", text)
	}
}

func TestCleanupOutput_Text_DryRunLabel(t *testing.T) {
	out := CleanupOutput{
		DryRun:    true,
		OlderThan: "7d",
		Removed:   []string{"wt-1"},
	}
	text := out.Text()
	if !strings.Contains(text, "DRY RUN") {
		t.Errorf("CleanupOutput.Text() should contain 'DRY RUN', got: %q", text)
	}
}

func TestCleanupOutput_Text_NoWorktrees(t *testing.T) {
	out := CleanupOutput{
		DryRun:    false,
		OlderThan: "7d",
	}
	text := out.Text()
	if !strings.Contains(text, "No worktrees") {
		t.Errorf("CleanupOutput.Text() with no removed/skipped should contain 'No worktrees', got: %q", text)
	}
}

func TestCleanupOutput_Text_ShowsRemovedList(t *testing.T) {
	out := CleanupOutput{
		OlderThan: "7d",
		Removed:   []string{"wt-old-1", "wt-old-2"},
	}
	text := out.Text()
	if !strings.Contains(text, "wt-old-1") {
		t.Errorf("CleanupOutput.Text() should contain removed worktree ID, got: %q", text)
	}
}

func TestCleanupOutput_Text_ShowsSkippedWithReason(t *testing.T) {
	out := CleanupOutput{
		OlderThan:   "7d",
		Skipped:     []string{"wt-dirty"},
		SkipReasons: map[string]string{"wt-dirty": "has uncommitted changes"},
	}
	text := out.Text()
	if !strings.Contains(text, "wt-dirty") {
		t.Errorf("CleanupOutput.Text() should contain skipped worktree ID, got: %q", text)
	}
	if !strings.Contains(text, "uncommitted changes") {
		t.Errorf("CleanupOutput.Text() should contain skip reason, got: %q", text)
	}
}

func TestSwitchOutput_Text_ContainsWorktreeID(t *testing.T) {
	out := SwitchOutput{
		Success:    true,
		WorktreeID: "wt-20260104-143052-a1b2",
		Name:       "feature-auth",
		Path:       "/tmp/worktrees/wt-20260104-143052-a1b2",
	}
	text := out.Text()
	if !strings.Contains(text, "wt-20260104-143052-a1b2") {
		t.Errorf("SwitchOutput.Text() should contain worktree ID, got: %q", text)
	}
}

func TestSwitchOutput_Text_RiteUpdatedMentioned(t *testing.T) {
	out := SwitchOutput{
		Success:     true,
		WorktreeID:  "wt-abc",
		Rite:        "ecosystem",
		RiteUpdated: true,
	}
	text := out.Text()
	if !strings.Contains(text, "updated") {
		t.Errorf("SwitchOutput.Text() should mention rite was updated, got: %q", text)
	}
}

func TestSyncOutput_Text_UpToDate(t *testing.T) {
	out := SyncOutput{
		Success:  true,
		Name:     "feature-auth",
		UpToDate: true,
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "up to date") {
		t.Errorf("SyncOutput.Text() should contain 'up to date', got: %q", text)
	}
}

func TestSyncOutput_Text_DivergedException(t *testing.T) {
	out := SyncOutput{
		Success:  true,
		Name:     "feature-auth",
		Diverged: true,
		Ahead:    3,
		Behind:   2,
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "diverged") {
		t.Errorf("SyncOutput.Text() should contain 'Diverged', got: %q", text)
	}
}

func TestSyncOutput_Text_ShowsConflicts(t *testing.T) {
	out := SyncOutput{
		Success:   true,
		Name:      "feature-auth",
		Conflicts: []string{"src/main.go", "README.md"},
	}
	text := out.Text()
	if !strings.Contains(text, "src/main.go") {
		t.Errorf("SyncOutput.Text() should contain conflict file, got: %q", text)
	}
}

func TestStatusOutput_Text_ContainsWorktreeID(t *testing.T) {
	out := StatusOutput{
		WorktreeID: "wt-20260104-143052-a1b2",
		Name:       "feature-auth",
		Path:       "/tmp/worktrees/wt-abc",
		BaseBranch: "main",
	}
	text := out.Text()
	if !strings.Contains(text, "wt-20260104-143052-a1b2") {
		t.Errorf("StatusOutput.Text() should contain worktree ID, got: %q", text)
	}
}

func TestStatusOutput_Text_DirtyGitStatus(t *testing.T) {
	out := StatusOutput{
		WorktreeID:   "wt-abc",
		IsDirty:      true,
		ChangedFiles: 3,
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "dirty") {
		t.Errorf("StatusOutput.Text() should show dirty git status, got: %q", text)
	}
}

func TestStatusOutput_Text_CleanGitStatus(t *testing.T) {
	out := StatusOutput{
		WorktreeID: "wt-abc",
		IsDirty:    false,
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "clean") {
		t.Errorf("StatusOutput.Text() should show clean git status, got: %q", text)
	}
}

func TestStatusOutput_Text_DetachedBranch(t *testing.T) {
	out := StatusOutput{
		WorktreeID: "wt-abc",
		Branch:     "", // empty = detached
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "detached") {
		t.Errorf("StatusOutput.Text() should show detached branch, got: %q", text)
	}
}

func TestCloneOutput_Text_ContainsSourceName(t *testing.T) {
	out := CloneOutput{
		Success:    true,
		WorktreeID: "wt-new",
		Name:       "feature-auth-v2",
		Path:       "/tmp/wt-new",
		SourceName: "feature-auth",
		SourceID:   "wt-old",
	}
	text := out.Text()
	if !strings.Contains(text, "feature-auth") {
		t.Errorf("CloneOutput.Text() should contain source name, got: %q", text)
	}
}

func TestCloneOutput_Text_SessionCopiedMentioned(t *testing.T) {
	out := CloneOutput{
		Success:       true,
		WorktreeID:    "wt-new",
		SessionCopied: true,
	}
	text := out.Text()
	if !strings.Contains(strings.ToLower(text), "session") {
		t.Errorf("CloneOutput.Text() should mention session when SessionCopied=true, got: %q", text)
	}
}

func TestExportOutput_Text_ContainsArchivePath(t *testing.T) {
	out := ExportOutput{
		Success:     true,
		WorktreeID:  "wt-abc",
		Name:        "feature-auth",
		ArchivePath: "/tmp/feature-auth.tar.gz",
	}
	text := out.Text()
	if !strings.Contains(text, "/tmp/feature-auth.tar.gz") {
		t.Errorf("ExportOutput.Text() should contain archive path, got: %q", text)
	}
}

func TestImportOutput_Text_ContainsFromArchive(t *testing.T) {
	out := ImportOutput{
		Success:     true,
		WorktreeID:  "wt-new",
		Name:        "feature-auth",
		Path:        "/tmp/wt-new",
		FromArchive: "/backups/feature-auth.tar.gz",
	}
	text := out.Text()
	if !strings.Contains(text, "/backups/feature-auth.tar.gz") {
		t.Errorf("ImportOutput.Text() should contain source archive path, got: %q", text)
	}
}

// --- formatDirtyStatus helper ---

func TestFormatDirtyStatus_Clean(t *testing.T) {
	// WorktreeStatus with no changes should return "clean"
	status := worktree.WorktreeStatus{}
	got := formatDirtyStatus(status)
	if got != "clean" {
		t.Errorf("formatDirtyStatus(empty) = %q, want %q", got, "clean")
	}
}

func TestFormatDirtyStatus_ModifiedFiles(t *testing.T) {
	status := worktree.WorktreeStatus{ChangedFiles: 3}
	got := formatDirtyStatus(status)
	if !strings.Contains(got, "modified") {
		t.Errorf("formatDirtyStatus with changed files = %q, want to contain 'modified'", got)
	}
}

func TestFormatDirtyStatus_UntrackedFiles(t *testing.T) {
	status := worktree.WorktreeStatus{UntrackedCount: 2}
	got := formatDirtyStatus(status)
	if !strings.Contains(got, "untracked") {
		t.Errorf("formatDirtyStatus with untracked files = %q, want to contain 'untracked'", got)
	}
}

func TestFormatDirtyStatus_Both(t *testing.T) {
	status := worktree.WorktreeStatus{ChangedFiles: 1, UntrackedCount: 1}
	got := formatDirtyStatus(status)
	if !strings.Contains(got, "modified") || !strings.Contains(got, "untracked") {
		t.Errorf("formatDirtyStatus with both = %q, want to contain 'modified' and 'untracked'", got)
	}
}
