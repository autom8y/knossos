package validate

import (
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/validation"
	"github.com/spf13/cobra"
)

// newTestCmd is a helper that creates a ValidateCmd with default test flags.
func newTestCmd() *cobra.Command {
	output := "text"
	verbose := false
	projectDir := ""
	return NewValidateCmd(&output, &verbose, &projectDir)
}

// --- Command metadata ---

func TestValidateCmd_Use(t *testing.T) {
	cmd := newTestCmd()
	if cmd.Use != "validate" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "validate")
	}
}

func TestValidateCmd_ShortDescription(t *testing.T) {
	cmd := newTestCmd()
	if cmd.Short == "" {
		t.Error("cmd.Short is empty, want non-empty description")
	}
}

func TestValidateCmd_LongDescription(t *testing.T) {
	cmd := newTestCmd()
	if cmd.Long == "" {
		t.Error("cmd.Long is empty, want non-empty long description")
	}
}

func TestValidateCmd_LongDescriptionMentionsArtifacts(t *testing.T) {
	cmd := newTestCmd()
	for _, keyword := range []string{"prd", "tdd", "adr"} {
		if !strings.Contains(strings.ToLower(cmd.Long), keyword) {
			t.Errorf("cmd.Long does not mention artifact type %q", keyword)
		}
	}
}

// --- NeedsProject annotation ---

func TestValidateCmd_NeedsProjectTrue(t *testing.T) {
	cmd := newTestCmd()
	if !common.NeedsProject(cmd) {
		t.Error("validate command should have needsProject=true")
	}
}

// --- Subcommand presence ---

func TestValidateCmd_HasArtifactSubcommand(t *testing.T) {
	cmd := newTestCmd()
	sub, _, err := cmd.Find([]string{"artifact"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("validate command missing 'artifact' subcommand")
	}
}

func TestValidateCmd_HasHandoffSubcommand(t *testing.T) {
	cmd := newTestCmd()
	sub, _, err := cmd.Find([]string{"handoff"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("validate command missing 'handoff' subcommand")
	}
}

func TestValidateCmd_HasSchemaSubcommand(t *testing.T) {
	cmd := newTestCmd()
	sub, _, err := cmd.Find([]string{"schema"})
	if err != nil || sub == nil || sub == cmd {
		t.Fatal("validate command missing 'schema' subcommand")
	}
}

// --- Artifact subcommand: metadata ---

func TestArtifactCmd_Use(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	if !strings.HasPrefix(sub.Use, "artifact") {
		t.Errorf("artifact subcommand Use = %q, want prefix 'artifact'", sub.Use)
	}
}

func TestArtifactCmd_ShortDescription(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	if sub.Short == "" {
		t.Error("artifact subcommand Short is empty")
	}
}

func TestArtifactCmd_LongDescription(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	if sub.Long == "" {
		t.Error("artifact subcommand Long is empty")
	}
}

// --- Artifact subcommand: flags ---

func TestArtifactCmd_FlagType_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	f := sub.Flags().Lookup("type")
	if f == nil {
		t.Fatal("artifact subcommand missing --type flag")
	}
}

func TestArtifactCmd_FlagType_Shorthand_t(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	f := sub.Flags().ShorthandLookup("t")
	if f == nil {
		t.Fatal("artifact subcommand missing -t shorthand for --type flag")
	}
}

func TestArtifactCmd_FlagType_DefaultEmpty(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	f := sub.Flags().Lookup("type")
	if f == nil {
		t.Fatal("--type flag not found")
	}
	if f.DefValue != "" {
		t.Errorf("--type default = %q, want %q (empty = auto-detect)", f.DefValue, "")
	}
}

// --- Artifact subcommand: requires exactly one arg ---

func TestArtifactCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"artifact"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no file argument given to 'artifact', got nil")
	}
}

func TestArtifactCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"artifact", "file1.md", "file2.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments given to 'artifact', got nil")
	}
}

// --- Handoff subcommand: metadata ---

func TestHandoffCmd_Use(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	if !strings.HasPrefix(sub.Use, "handoff") {
		t.Errorf("handoff subcommand Use = %q, want prefix 'handoff'", sub.Use)
	}
}

func TestHandoffCmd_ShortDescription(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	if sub.Short == "" {
		t.Error("handoff subcommand Short is empty")
	}
}

func TestHandoffCmd_LongDescription(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	if sub.Long == "" {
		t.Error("handoff subcommand Long is empty")
	}
}

func TestHandoffCmd_LongDescriptionMentionsPhases(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	for _, keyword := range []string{"requirements", "design", "implementation"} {
		if !strings.Contains(strings.ToLower(sub.Long), keyword) {
			t.Errorf("handoff Long does not mention phase %q", keyword)
		}
	}
}

// --- Handoff subcommand: flags ---

func TestHandoffCmd_FlagPhase_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("phase")
	if f == nil {
		t.Fatal("handoff subcommand missing --phase flag")
	}
}

func TestHandoffCmd_FlagArtifact_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("artifact")
	if f == nil {
		t.Fatal("handoff subcommand missing --artifact flag")
	}
}

func TestHandoffCmd_FlagType_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("type")
	if f == nil {
		t.Fatal("handoff subcommand missing --type flag")
	}
}

func TestHandoffCmd_FlagListPhases_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("list-phases")
	if f == nil {
		t.Fatal("handoff subcommand missing --list-phases flag")
	}
}

func TestHandoffCmd_FlagShowCriteria_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("show-criteria")
	if f == nil {
		t.Fatal("handoff subcommand missing --show-criteria flag")
	}
}

func TestHandoffCmd_FlagPhase_DefaultEmpty(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("phase")
	if f.DefValue != "" {
		t.Errorf("--phase default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestHandoffCmd_FlagArtifact_DefaultEmpty(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("artifact")
	if f.DefValue != "" {
		t.Errorf("--artifact default = %q, want %q (empty)", f.DefValue, "")
	}
}

func TestHandoffCmd_FlagListPhases_DefaultFalse(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("list-phases")
	if f.DefValue != "false" {
		t.Errorf("--list-phases default = %q, want %q", f.DefValue, "false")
	}
}

func TestHandoffCmd_FlagShowCriteria_DefaultFalse(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	f := sub.Flags().Lookup("show-criteria")
	if f.DefValue != "false" {
		t.Errorf("--show-criteria default = %q, want %q", f.DefValue, "false")
	}
}

// --- Handoff subcommand: missing required flags ---

func TestHandoffCmd_NoPhase_WithArtifact_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"handoff", "--artifact=docs/PRD-foo.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --phase missing, got nil")
	}
	if !strings.Contains(err.Error(), "--phase") {
		t.Errorf("error should mention '--phase', got: %q", err.Error())
	}
}

func TestHandoffCmd_WithPhase_NoArtifact_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"handoff", "--phase=requirements"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --artifact missing, got nil")
	}
	if !strings.Contains(err.Error(), "--artifact") {
		t.Errorf("error should mention '--artifact', got: %q", err.Error())
	}
}

func TestHandoffCmd_ShowCriteria_MissingPhase_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"handoff", "--show-criteria", "--type=prd"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --phase missing with --show-criteria, got nil")
	}
	if !strings.Contains(err.Error(), "--phase") {
		t.Errorf("error should mention '--phase', got: %q", err.Error())
	}
}

func TestHandoffCmd_ShowCriteria_MissingType_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"handoff", "--show-criteria", "--phase=requirements"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --type missing with --show-criteria, got nil")
	}
	if !strings.Contains(err.Error(), "--type") {
		t.Errorf("error should mention '--type', got: %q", err.Error())
	}
}

func TestHandoffCmd_InvalidPhase_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"handoff", "--phase=bogus", "--artifact=docs/PRD-foo.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --phase value, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "phase") {
		t.Errorf("error should mention 'phase', got: %q", err.Error())
	}
}

// --- Schema subcommand: metadata ---

func TestSchemaCmd_Use(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"schema"})
	if !strings.HasPrefix(sub.Use, "schema") {
		t.Errorf("schema subcommand Use = %q, want prefix 'schema'", sub.Use)
	}
}

func TestSchemaCmd_ShortDescription(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"schema"})
	if sub.Short == "" {
		t.Error("schema subcommand Short is empty")
	}
}

func TestSchemaCmd_LongDescription(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"schema"})
	if sub.Long == "" {
		t.Error("schema subcommand Long is empty")
	}
}

func TestSchemaCmd_LongDescriptionMentionsSchemas(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"schema"})
	for _, keyword := range []string{"prd", "tdd", "adr"} {
		if !strings.Contains(strings.ToLower(sub.Long), keyword) {
			t.Errorf("schema Long does not mention schema type %q", keyword)
		}
	}
}

// --- Schema subcommand: requires exactly two positional args ---

func TestSchemaCmd_NoArgs_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"schema"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args given to 'schema', got nil")
	}
}

func TestSchemaCmd_OneArg_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"schema", "prd"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when only one arg given to 'schema', got nil")
	}
}

func TestSchemaCmd_TooManyArgs_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"schema", "prd", "file1.md", "extra.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many args given to 'schema', got nil")
	}
}

func TestSchemaCmd_UnknownSchemaName_ReturnsError(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"schema", "badschema", "somefile.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown schema name, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "schema") {
		t.Errorf("error should mention 'schema', got: %q", err.Error())
	}
}

// --- Schema subcommand: deprecated --schema flag still exists ---

func TestSchemaCmd_DeprecatedSchemaFlag_Exists(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"schema"})
	f := sub.Flags().Lookup("schema")
	if f == nil {
		t.Fatal("schema subcommand missing --schema flag (deprecated flag)")
	}
}

// --- ArtifactOutput.Text() formatting ---

func TestArtifactOutput_Text_ValidNoType(t *testing.T) {
	out := ArtifactOutput{
		Valid:    true,
		FilePath: "docs/requirements/PRD-foo.md",
	}
	text := out.Text()
	if !strings.Contains(text, "VALID") {
		t.Errorf("Text() should contain VALID for valid artifact, got: %q", text)
	}
	if !strings.Contains(text, "PRD-foo.md") {
		t.Errorf("Text() should contain file path, got: %q", text)
	}
}

func TestArtifactOutput_Text_ValidWithType(t *testing.T) {
	out := ArtifactOutput{
		Valid:        true,
		FilePath:     "docs/requirements/PRD-foo.md",
		ArtifactType: "prd",
	}
	text := out.Text()
	if !strings.Contains(text, "VALID") {
		t.Errorf("Text() should contain VALID, got: %q", text)
	}
	if !strings.Contains(text, "prd") {
		t.Errorf("Text() should contain artifact type, got: %q", text)
	}
}

func TestArtifactOutput_Text_InvalidWithIssues(t *testing.T) {
	out := ArtifactOutput{
		Valid:        false,
		FilePath:     "docs/requirements/PRD-foo.md",
		ArtifactType: "prd",
		Issues: []validation.ValidationIssue{
			{Field: "title", Message: "title is required"},
			{Message: "missing author"},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "INVALID") {
		t.Errorf("Text() should contain INVALID for invalid artifact, got: %q", text)
	}
	if !strings.Contains(text, "title is required") {
		t.Errorf("Text() should contain issue message, got: %q", text)
	}
	if !strings.Contains(text, "missing author") {
		t.Errorf("Text() should contain issue message without field, got: %q", text)
	}
}

func TestArtifactOutput_Text_IssueWithField_ShowsFieldInBrackets(t *testing.T) {
	out := ArtifactOutput{
		Valid:    false,
		FilePath: "foo.md",
		Issues: []validation.ValidationIssue{
			{Field: "status", Message: "required field missing"},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "[status]") {
		t.Errorf("Text() should show field in brackets for issues with field, got: %q", text)
	}
}

// --- HandoffOutput.Text() formatting ---

func TestHandoffOutput_Text_Passed(t *testing.T) {
	out := HandoffOutput{
		Passed:       true,
		Phase:        "requirements",
		ArtifactType: "prd",
	}
	text := out.Text()
	if !strings.Contains(text, "PASSED") {
		t.Errorf("Text() should contain PASSED, got: %q", text)
	}
	if !strings.Contains(text, "requirements") {
		t.Errorf("Text() should contain phase, got: %q", text)
	}
}

func TestHandoffOutput_Text_Failed(t *testing.T) {
	out := HandoffOutput{
		Passed:       false,
		Phase:        "requirements",
		ArtifactType: "prd",
		BlockingFailed: []validation.CriterionResult{
			{
				Criterion: validation.Criterion{Field: "status", Message: "status must be APPROVED"},
				Message:   "status is DRAFT",
			},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "FAILED") {
		t.Errorf("Text() should contain FAILED, got: %q", text)
	}
	if !strings.Contains(text, "status") {
		t.Errorf("Text() should contain blocking failure field, got: %q", text)
	}
}

func TestHandoffOutput_Text_PassedWithFilePath(t *testing.T) {
	out := HandoffOutput{
		Passed:       true,
		Phase:        "design",
		ArtifactType: "tdd",
		FilePath:     "docs/design/TDD-foo.md",
	}
	text := out.Text()
	if !strings.Contains(text, "TDD-foo.md") {
		t.Errorf("Text() should contain file path when set, got: %q", text)
	}
}

func TestHandoffOutput_Text_WithWarnings(t *testing.T) {
	out := HandoffOutput{
		Passed:       true,
		Phase:        "requirements",
		ArtifactType: "prd",
		Warnings: []validation.CriterionResult{
			{
				Criterion: validation.Criterion{Field: "notes", Message: "notes recommended"},
				Message:   "notes field is empty",
			},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "Warnings") {
		t.Errorf("Text() should contain Warnings section, got: %q", text)
	}
}

// --- PhaseCriteriaOutput.Text() formatting ---

func TestPhaseCriteriaOutput_Text_ContainsPhases(t *testing.T) {
	out := PhaseCriteriaOutput{
		Phases: []PhaseInfo{
			{Phase: "requirements", ArtifactTypes: []string{"prd"}},
			{Phase: "design", ArtifactTypes: []string{"tdd"}},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "requirements") {
		t.Errorf("Text() should contain phase name, got: %q", text)
	}
	if !strings.Contains(text, "prd") {
		t.Errorf("Text() should contain artifact type, got: %q", text)
	}
}

func TestPhaseCriteriaOutput_Text_HasHeader(t *testing.T) {
	out := PhaseCriteriaOutput{Phases: []PhaseInfo{}}
	text := out.Text()
	if text == "" {
		t.Error("Text() should return non-empty string even for empty phases")
	}
}

// --- CriteriaDetailOutput.Text() formatting ---

func TestCriteriaDetailOutput_Text_HasPhaseAndType(t *testing.T) {
	out := CriteriaDetailOutput{
		Phase:        "requirements",
		ArtifactType: "prd",
	}
	text := out.Text()
	if !strings.Contains(text, "requirements") {
		t.Errorf("Text() should contain phase, got: %q", text)
	}
	if !strings.Contains(text, "prd") {
		t.Errorf("Text() should contain artifact type, got: %q", text)
	}
}

func TestCriteriaDetailOutput_Text_ShowsBlockingCriteria(t *testing.T) {
	out := CriteriaDetailOutput{
		Phase:        "requirements",
		ArtifactType: "prd",
		Blocking: []validation.Criterion{
			{Field: "title", Message: "title is required"},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "Blocking") {
		t.Errorf("Text() should contain 'Blocking' section, got: %q", text)
	}
	if !strings.Contains(text, "title") {
		t.Errorf("Text() should contain blocking criterion field, got: %q", text)
	}
}

func TestCriteriaDetailOutput_Text_ShowsNonBlockingCriteria(t *testing.T) {
	out := CriteriaDetailOutput{
		Phase:        "requirements",
		ArtifactType: "prd",
		NonBlocking: []validation.Criterion{
			{Field: "notes", Message: "notes recommended"},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "Non-blocking") {
		t.Errorf("Text() should contain 'Non-blocking' section, got: %q", text)
	}
}

// --- SchemaOutput.Text() formatting ---

func TestSchemaOutput_Text_Valid(t *testing.T) {
	out := SchemaOutput{
		Valid:      true,
		SchemaName: "prd",
		FilePath:   "docs/requirements/PRD-foo.md",
	}
	text := out.Text()
	if !strings.Contains(text, "VALID") {
		t.Errorf("Text() should contain VALID, got: %q", text)
	}
	if !strings.Contains(text, "prd") {
		t.Errorf("Text() should contain schema name, got: %q", text)
	}
	if !strings.Contains(text, "PRD-foo.md") {
		t.Errorf("Text() should contain file path, got: %q", text)
	}
}

func TestSchemaOutput_Text_InvalidWithIssues(t *testing.T) {
	out := SchemaOutput{
		Valid:      false,
		SchemaName: "prd",
		FilePath:   "docs/requirements/PRD-foo.md",
		Issues: []validation.ValidationIssue{
			{Field: "status", Message: "invalid status value"},
			{Message: "unknown error"},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "INVALID") {
		t.Errorf("Text() should contain INVALID, got: %q", text)
	}
	if !strings.Contains(text, "invalid status value") {
		t.Errorf("Text() should contain issue message, got: %q", text)
	}
}

func TestSchemaOutput_Text_IssueWithField_ShowsBrackets(t *testing.T) {
	out := SchemaOutput{
		Valid:      false,
		SchemaName: "adr",
		FilePath:   "docs/decisions/ADR-0001.md",
		Issues: []validation.ValidationIssue{
			{Field: "number", Message: "number is required"},
		},
	}
	text := out.Text()
	if !strings.Contains(text, "[number]") {
		t.Errorf("Text() should show field in brackets, got: %q", text)
	}
}

// --- issueMessages helper ---

func TestIssueMessages_EmptySlice(t *testing.T) {
	msgs := issueMessages(nil)
	if len(msgs) != 0 {
		t.Errorf("issueMessages(nil) = %v, want empty", msgs)
	}
}

func TestIssueMessages_WithFieldAndMessage(t *testing.T) {
	issues := []validation.ValidationIssue{
		{Field: "title", Message: "required"},
		{Field: "", Message: "global issue"},
	}
	msgs := issueMessages(issues)
	if len(msgs) != 2 {
		t.Fatalf("issueMessages len = %d, want 2", len(msgs))
	}
	if !strings.Contains(msgs[0], "[title]") {
		t.Errorf("msgs[0] = %q, want to contain '[title]'", msgs[0])
	}
	if !strings.Contains(msgs[0], "required") {
		t.Errorf("msgs[0] = %q, want to contain 'required'", msgs[0])
	}
	// Issue without field should not have brackets
	if strings.Contains(msgs[1], "[") {
		t.Errorf("msgs[1] = %q, should not contain brackets when field is empty", msgs[1])
	}
	if msgs[1] != "global issue" {
		t.Errorf("msgs[1] = %q, want %q", msgs[1], "global issue")
	}
}

func TestIssueMessages_MultipleIssues(t *testing.T) {
	issues := []validation.ValidationIssue{
		{Field: "a", Message: "err1"},
		{Field: "b", Message: "err2"},
		{Message: "err3"},
	}
	msgs := issueMessages(issues)
	if len(msgs) != 3 {
		t.Fatalf("issueMessages len = %d, want 3", len(msgs))
	}
}

// --- Dispatch layer: subcommand NeedsProject propagation ---

func TestValidateCmd_SubcommandArtifact_InheritsNeedsProject(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"artifact"})
	if !common.NeedsProject(sub) {
		t.Error("artifact subcommand should inherit needsProject=true from parent")
	}
}

func TestValidateCmd_SubcommandHandoff_InheritsNeedsProject(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"handoff"})
	if !common.NeedsProject(sub) {
		t.Error("handoff subcommand should inherit needsProject=true from parent")
	}
}

func TestValidateCmd_SubcommandSchema_InheritsNeedsProject(t *testing.T) {
	cmd := newTestCmd()
	sub, _, _ := cmd.Find([]string{"schema"})
	if !common.NeedsProject(sub) {
		t.Error("schema subcommand should inherit needsProject=true from parent")
	}
}
