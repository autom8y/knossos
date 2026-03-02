package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/spf13/cobra"
)

// newTestAgentCmd creates an AgentCmd with default test flags.
func newTestAgentCmd() *cobra.Command {
	output := "text"
	verbose := false
	projectDir := ""
	return NewAgentCmd(&output, &verbose, &projectDir)
}

// --- Command metadata ---

func TestAgentCmd_Use(t *testing.T) {
	cmd := newTestAgentCmd()
	if cmd.Use != "agent" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "agent")
	}
}

func TestAgentCmd_ShortDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	if cmd.Short == "" {
		t.Error("cmd.Short is empty, want non-empty description")
	}
}

func TestAgentCmd_LongDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	if cmd.Long == "" {
		t.Error("cmd.Long is empty, want non-empty long description")
	}
}

// --- NeedsProject annotation ---

func TestAgentCmd_NeedsProjectTrue(t *testing.T) {
	cmd := newTestAgentCmd()
	if !common.NeedsProject(cmd) {
		t.Error("agent command should have needsProject=true")
	}
}

// --- Subcommand presence ---

func TestAgentCmd_HasValidateSubcommand(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, err := cmd.Find([]string{"validate"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("agent command missing 'validate' subcommand")
	}
}

func TestAgentCmd_HasListSubcommand(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, err := cmd.Find([]string{"list"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("agent command missing 'list' subcommand")
	}
}

func TestAgentCmd_HasNewSubcommand(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, err := cmd.Find([]string{"new"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("agent command missing 'new' subcommand")
	}
}

func TestAgentCmd_HasUpdateSubcommand(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, err := cmd.Find([]string{"update"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("agent command missing 'update' subcommand")
	}
}

// --- validate subcommand: metadata ---

func TestAgentValidateCmd_Use(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	if !strings.HasPrefix(sub.Use, "validate") {
		t.Errorf("validate subcommand Use = %q, want prefix 'validate'", sub.Use)
	}
}

func TestAgentValidateCmd_ShortDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	if sub.Short == "" {
		t.Error("validate subcommand Short is empty")
	}
}

func TestAgentValidateCmd_LongDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	if sub.Long == "" {
		t.Error("validate subcommand Long is empty")
	}
}

// --- validate subcommand: flags ---

func TestAgentValidateCmd_FlagRite_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("validate subcommand missing --rite flag")
	}
}

func TestAgentValidateCmd_FlagStrict_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().Lookup("strict")
	if f == nil {
		t.Fatal("validate subcommand missing --strict flag")
	}
}

func TestAgentValidateCmd_FlagAll_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().Lookup("all")
	if f == nil {
		t.Fatal("validate subcommand missing --all flag")
	}
}

func TestAgentValidateCmd_FlagRite_Shorthand_r(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().ShorthandLookup("r")
	if f == nil {
		t.Fatal("validate subcommand missing -r shorthand for --rite flag")
	}
}

// --- validate subcommand: flag defaults ---

func TestAgentValidateCmd_FlagRite_DefaultEmpty(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().Lookup("rite")
	if f.DefValue != "" {
		t.Errorf("--rite default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestAgentValidateCmd_FlagStrict_DefaultFalse(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().Lookup("strict")
	if f.DefValue != "false" {
		t.Errorf("--strict default = %q, want %q", f.DefValue, "false")
	}
}

func TestAgentValidateCmd_FlagAll_DefaultFalse(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	f := sub.Flags().Lookup("all")
	if f.DefValue != "false" {
		t.Errorf("--all default = %q, want %q", f.DefValue, "false")
	}
}

// --- validate subcommand: execution with empty directory ---

func TestAgentValidateCmd_EmptyProject_ReturnsNoError(t *testing.T) {
	tmpDir := t.TempDir()
	// Create rites dir so collectAllAgents doesn't fail
	if err := os.MkdirAll(filepath.Join(tmpDir, "rites"), 0755); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"validate", "--all"})
	// With no agent files, it should succeed (just print "No agent files found")
	err := cmd.Execute()
	if err != nil {
		t.Errorf("validate on empty project returned error: %v", err)
	}
}

func TestAgentValidateCmd_SpecificFile_NotExist_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"validate", "/nonexistent/path/agent.md"})
	err := cmd.Execute()
	// Error is expected due to validation failure (file not found → errorCount > 0)
	if err == nil {
		t.Error("expected error when validating a non-existent agent file, got nil")
	}
}

// --- validate subcommand: validates a valid agent file ---

func TestAgentValidateCmd_ValidAgentFile_PassesValidation(t *testing.T) {
	tmpDir := t.TempDir()
	agentContent := `---
name: test-agent
description: A test agent for unit testing
tools:
  - Read
  - Bash
---

## Core Responsibilities

Test responsibilities.
`
	agentPath := filepath.Join(tmpDir, "test-agent.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"validate", agentPath})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("validate on valid agent file returned error: %v", err)
	}
}

func TestAgentValidateCmd_InvalidAgentFile_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	// Agent missing required 'name' field
	agentContent := `---
description: A test agent with missing name
tools:
  - Read
---

## Core Responsibilities

Test responsibilities.
`
	agentPath := filepath.Join(tmpDir, "bad-agent.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"validate", agentPath})
	err := cmd.Execute()
	if err == nil {
		t.Error("validate on invalid agent (missing name) should return error, got nil")
	}
}

// --- list subcommand: metadata ---

func TestAgentListCmd_Use(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if !strings.HasPrefix(sub.Use, "list") {
		t.Errorf("list subcommand Use = %q, want prefix 'list'", sub.Use)
	}
}

func TestAgentListCmd_ShortDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if sub.Short == "" {
		t.Error("list subcommand Short is empty")
	}
}

func TestAgentListCmd_LongDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if sub.Long == "" {
		t.Error("list subcommand Long is empty")
	}
}

// --- list subcommand: flags ---

func TestAgentListCmd_FlagRite_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	f := sub.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("list subcommand missing --rite flag")
	}
}

func TestAgentListCmd_FlagAll_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	f := sub.Flags().Lookup("all")
	if f == nil {
		t.Fatal("list subcommand missing --all flag")
	}
}

func TestAgentListCmd_FlagRite_Shorthand_r(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	f := sub.Flags().ShorthandLookup("r")
	if f == nil {
		t.Fatal("list subcommand missing -r shorthand for --rite flag")
	}
}

func TestAgentListCmd_FlagRite_DefaultEmpty(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	f := sub.Flags().Lookup("rite")
	if f.DefValue != "" {
		t.Errorf("--rite default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestAgentListCmd_FlagAll_DefaultFalse(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	f := sub.Flags().Lookup("all")
	if f.DefValue != "false" {
		t.Errorf("--all default = %q, want %q", f.DefValue, "false")
	}
}

// --- list subcommand: execution ---

func TestAgentListCmd_EmptyProject_ReturnsNoError(t *testing.T) {
	tmpDir := t.TempDir()
	// Create rites dir so collectAllAgents doesn't fail
	if err := os.MkdirAll(filepath.Join(tmpDir, "rites"), 0755); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("list on empty project returned error: %v", err)
	}
}

func TestAgentListCmd_WithAgentFile_ReturnsNoError(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "rites", "test-rite", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	agentContent := `---
name: my-agent
description: A test agent
tools:
  - Read
---

# My Agent
`
	if err := os.WriteFile(filepath.Join(agentsDir, "my-agent.md"), []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("list with agent file returned error: %v", err)
	}
}

// --- new subcommand: metadata ---

func TestAgentNewCmd_Use(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	if !strings.HasPrefix(sub.Use, "new") {
		t.Errorf("new subcommand Use = %q, want prefix 'new'", sub.Use)
	}
}

func TestAgentNewCmd_ShortDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	if sub.Short == "" {
		t.Error("new subcommand Short is empty")
	}
}

func TestAgentNewCmd_LongDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	if sub.Long == "" {
		t.Error("new subcommand Long is empty")
	}
}

func TestAgentNewCmd_LongDescriptionMentionsArchetypes(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	// Should mention valid archetypes
	for _, archetype := range agentpkg.ListArchetypes() {
		if !strings.Contains(strings.ToLower(sub.Long), archetype) {
			t.Errorf("new subcommand Long does not mention archetype %q", archetype)
		}
	}
}

// --- new subcommand: flags ---

func TestAgentNewCmd_FlagArchetype_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().Lookup("archetype")
	if f == nil {
		t.Fatal("new subcommand missing --archetype flag")
	}
}

func TestAgentNewCmd_FlagRite_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("new subcommand missing --rite flag")
	}
}

func TestAgentNewCmd_FlagName_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().Lookup("name")
	if f == nil {
		t.Fatal("new subcommand missing --name flag")
	}
}

func TestAgentNewCmd_FlagDescription_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().Lookup("description")
	if f == nil {
		t.Fatal("new subcommand missing --description flag")
	}
}

func TestAgentNewCmd_FlagArchetype_Shorthand_a(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().ShorthandLookup("a")
	if f == nil {
		t.Fatal("new subcommand missing -a shorthand for --archetype flag")
	}
}

func TestAgentNewCmd_FlagRite_Shorthand_r(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().ShorthandLookup("r")
	if f == nil {
		t.Fatal("new subcommand missing -r shorthand for --rite flag")
	}
}

func TestAgentNewCmd_FlagName_Shorthand_n(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().ShorthandLookup("n")
	if f == nil {
		t.Fatal("new subcommand missing -n shorthand for --name flag")
	}
}

func TestAgentNewCmd_FlagDescription_Shorthand_d(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	f := sub.Flags().ShorthandLookup("d")
	if f == nil {
		t.Fatal("new subcommand missing -d shorthand for --description flag")
	}
}

// --- new subcommand: required flags ---

func TestAgentNewCmd_MissingArchetype_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new", "--rite=my-rite", "--name=my-agent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --archetype missing, got nil")
	}
}

func TestAgentNewCmd_MissingRite_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new", "--archetype=specialist", "--name=my-agent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --rite missing, got nil")
	}
}

func TestAgentNewCmd_MissingName_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new", "--archetype=specialist", "--rite=my-rite"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --name missing, got nil")
	}
}

// --- new subcommand: execution ---

func TestAgentNewCmd_InvalidArchetype_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "rites", "test-rite"), 0755); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new",
		"--archetype=invalid-archetype",
		"--rite=test-rite",
		"--name=my-agent",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid archetype, got nil")
	}
}

func TestAgentNewCmd_RiteNotExist_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new",
		"--archetype=specialist",
		"--rite=nonexistent-rite",
		"--name=my-agent",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when rite does not exist, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") &&
		!strings.Contains(strings.ToLower(err.Error()), "rite") {
		t.Errorf("error should mention rite not found, got: %q", err.Error())
	}
}

func TestAgentNewCmd_ValidArgs_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	// RiteDir checks for manifest.yaml in {projectRoot}/.knossos/rites/{name}/
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "my-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create manifest.yaml so RiteDir returns the project-local path
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte("name: my-rite\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new",
		"--archetype=specialist",
		"--rite=my-rite",
		"--name=my-new-agent",
		"--description=A test agent",
	})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("new command returned error: %v", err)
	}

	// Verify the file was created in the rite's agents directory
	expectedPath := filepath.Join(riteDir, "agents", "my-new-agent.md")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Errorf("expected agent file to be created at %s, but it does not exist", expectedPath)
	}
}

func TestAgentNewCmd_AgentAlreadyExists_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	// RiteDir checks for manifest.yaml in {projectRoot}/.knossos/rites/{name}/
	riteDir := filepath.Join(tmpDir, ".knossos", "rites", "my-rite")
	agentsDir := filepath.Join(riteDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create manifest.yaml so RiteDir returns the project-local path
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte("name: my-rite\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the agent file first so the command sees a duplicate
	existingPath := filepath.Join(agentsDir, "existing-agent.md")
	if err := os.WriteFile(existingPath, []byte("---\nname: existing\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"new",
		"--archetype=specialist",
		"--rite=my-rite",
		"--name=existing-agent",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when agent file already exists, got nil")
	}
}

// --- update subcommand: metadata ---

func TestAgentUpdateCmd_Use(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	if !strings.HasPrefix(sub.Use, "update") {
		t.Errorf("update subcommand Use = %q, want prefix 'update'", sub.Use)
	}
}

func TestAgentUpdateCmd_ShortDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	if sub.Short == "" {
		t.Error("update subcommand Short is empty")
	}
}

func TestAgentUpdateCmd_LongDescription(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	if sub.Long == "" {
		t.Error("update subcommand Long is empty")
	}
}

// --- update subcommand: flags ---

func TestAgentUpdateCmd_FlagRite_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().Lookup("rite")
	if f == nil {
		t.Fatal("update subcommand missing --rite flag")
	}
}

func TestAgentUpdateCmd_FlagAll_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().Lookup("all")
	if f == nil {
		t.Fatal("update subcommand missing --all flag")
	}
}

func TestAgentUpdateCmd_FlagDryRun_Exists(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("update subcommand missing --dry-run flag")
	}
}

func TestAgentUpdateCmd_FlagRite_Shorthand_r(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().ShorthandLookup("r")
	if f == nil {
		t.Fatal("update subcommand missing -r shorthand for --rite flag")
	}
}

func TestAgentUpdateCmd_FlagRite_DefaultEmpty(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().Lookup("rite")
	if f.DefValue != "" {
		t.Errorf("--rite default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestAgentUpdateCmd_FlagAll_DefaultFalse(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().Lookup("all")
	if f.DefValue != "false" {
		t.Errorf("--all default = %q, want %q", f.DefValue, "false")
	}
}

func TestAgentUpdateCmd_FlagDryRun_DefaultFalse(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	f := sub.Flags().Lookup("dry-run")
	if f.DefValue != "false" {
		t.Errorf("--dry-run default = %q, want %q", f.DefValue, "false")
	}
}

// --- update subcommand: no target → error ---

func TestAgentUpdateCmd_NoTarget_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"update"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no target specified for update, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "must specify") {
		t.Errorf("error should mention 'must specify', got: %q", err.Error())
	}
}

// --- update subcommand: execution ---

func TestAgentUpdateCmd_WithAll_EmptyProject_ReturnsNoError(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "rites"), 0755); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"update", "--all"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("update --all on empty project returned error: %v", err)
	}
}

func TestAgentUpdateCmd_DryRun_WithAll_EmptyProject_ReturnsNoError(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "rites"), 0755); err != nil {
		t.Fatal(err)
	}

	output := "text"
	verbose := false
	cmd := NewAgentCmd(&output, &verbose, &tmpDir)
	cmd.SetArgs([]string{"update", "--all", "--dry-run"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("update --all --dry-run on empty project returned error: %v", err)
	}
}

// --- Dispatch layer: subcommand NeedsProject propagation ---

func TestAgentCmd_SubcommandValidate_InheritsNeedsProject(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"validate"})
	if !common.NeedsProject(sub) {
		t.Error("validate subcommand should inherit needsProject=true from parent")
	}
}

func TestAgentCmd_SubcommandList_InheritsNeedsProject(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"list"})
	if !common.NeedsProject(sub) {
		t.Error("list subcommand should inherit needsProject=true from parent")
	}
}

func TestAgentCmd_SubcommandNew_InheritsNeedsProject(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"new"})
	if !common.NeedsProject(sub) {
		t.Error("new subcommand should inherit needsProject=true from parent")
	}
}

func TestAgentCmd_SubcommandUpdate_InheritsNeedsProject(t *testing.T) {
	cmd := newTestAgentCmd()
	sub, _, _ := cmd.Find([]string{"update"})
	if !common.NeedsProject(sub) {
		t.Error("update subcommand should inherit needsProject=true from parent")
	}
}

// --- hasSectionsToUpdate helper ---

func TestHasSectionsToUpdate_WithPlatformSection_ReturnsTrue(t *testing.T) {
	archetype, err := agentpkg.GetArchetype("specialist")
	if err != nil {
		t.Fatalf("failed to get archetype: %v", err)
	}
	// specialist has platform sections
	if !hasSectionsToUpdate(archetype) {
		t.Error("hasSectionsToUpdate(specialist) = false, want true (has platform sections)")
	}
}

func TestHasSectionsToUpdate_WithOrchestratorArchetype_ReturnsTrue(t *testing.T) {
	archetype, err := agentpkg.GetArchetype("orchestrator")
	if err != nil {
		t.Fatalf("failed to get archetype: %v", err)
	}
	if !hasSectionsToUpdate(archetype) {
		t.Error("hasSectionsToUpdate(orchestrator) = false, want true (has platform sections)")
	}
}

// --- countChangedSections helper ---

func TestCountChangedSections_IdenticalSections_ReturnsZero(t *testing.T) {
	// Use a specialist archetype heading so mapSectionsToArchetype assigns a Name.
	// "Core Responsibilities" maps to section name "core-responsibilities" in specialist.
	content := `---
name: test-agent
description: A test agent
type: specialist
tools:
  - Read
---

## Core Responsibilities

Same content here.
`
	sections, err := agentpkg.ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("failed to parse agent sections: %v", err)
	}
	// Comparing identical copies should return zero changes
	changed := countChangedSections(sections, sections)
	if changed != 0 {
		t.Errorf("countChangedSections(identical) = %d, want 0", changed)
	}
}

func TestCountChangedSections_DifferentContent_ReturnsNonZero(t *testing.T) {
	// Use a specialist archetype heading so mapSectionsToArchetype assigns a Name.
	// Without a Name, countChangedSections skips the section entirely.
	original := `---
name: test-agent
description: A test agent
type: specialist
tools:
  - Read
---

## Core Responsibilities

Original content here.
`
	updated := `---
name: test-agent
description: A test agent
type: specialist
tools:
  - Read
---

## Core Responsibilities

Different content here — this is changed.
`
	origSections, err := agentpkg.ParseAgentSections([]byte(original))
	if err != nil {
		t.Fatalf("failed to parse original sections: %v", err)
	}
	updatedSections, err := agentpkg.ParseAgentSections([]byte(updated))
	if err != nil {
		t.Fatalf("failed to parse updated sections: %v", err)
	}

	changed := countChangedSections(origSections, updatedSections)
	if changed == 0 {
		t.Error("countChangedSections with different content = 0, want > 0")
	}
}
