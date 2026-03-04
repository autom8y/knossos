package materialize

import (
	"reflect"
	"testing"
)

// --- MergeSkillPolicies tests ---

func TestMergeSkillPolicies_BothEmpty(t *testing.T) {
	result := MergeSkillPolicies(nil, nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMergeSkillPolicies_EmptyShared(t *testing.T) {
	rite := []SkillPolicy{
		{Skill: "foo", Mode: "inject"},
	}
	result := MergeSkillPolicies(nil, rite)
	if len(result) != 1 || result[0].Skill != "foo" {
		t.Errorf("expected [{foo inject}], got %v", result)
	}
}

func TestMergeSkillPolicies_EmptyRite(t *testing.T) {
	shared := []SkillPolicy{
		{Skill: "bar", Mode: "inject"},
	}
	result := MergeSkillPolicies(shared, nil)
	if len(result) != 1 || result[0].Skill != "bar" {
		t.Errorf("expected [{bar inject}], got %v", result)
	}
}

func TestMergeSkillPolicies_NoOverlap(t *testing.T) {
	shared := []SkillPolicy{
		{Skill: "shared-skill", Mode: "inject"},
	}
	rite := []SkillPolicy{
		{Skill: "rite-skill", Mode: "inject"},
	}
	result := MergeSkillPolicies(shared, rite)
	if len(result) != 2 {
		t.Fatalf("expected 2 policies, got %d: %v", len(result), result)
	}
	// Shared comes first, rite appended
	if result[0].Skill != "shared-skill" {
		t.Errorf("expected shared-skill first, got %s", result[0].Skill)
	}
	if result[1].Skill != "rite-skill" {
		t.Errorf("expected rite-skill second, got %s", result[1].Skill)
	}
}

func TestMergeSkillPolicies_RiteOverridesShared(t *testing.T) {
	shared := []SkillPolicy{
		{Skill: "ecosystem-ref", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	rite := []SkillPolicy{
		// Same skill, different mode and conditions — rite wins
		{Skill: "ecosystem-ref", Mode: "reference", RequiresTools: []string{"Read"}},
	}
	result := MergeSkillPolicies(shared, rite)
	if len(result) != 1 {
		t.Fatalf("expected 1 policy (rite overrides shared), got %d: %v", len(result), result)
	}
	if result[0].Mode != "reference" {
		t.Errorf("expected rite policy mode 'reference', got %q", result[0].Mode)
	}
	if !reflect.DeepEqual(result[0].RequiresTools, []string{"Read"}) {
		t.Errorf("expected rite requires_tools [Read], got %v", result[0].RequiresTools)
	}
}

func TestMergeSkillPolicies_OrderPreserved(t *testing.T) {
	shared := []SkillPolicy{
		{Skill: "a", Mode: "inject"},
		{Skill: "b", Mode: "inject"},
		{Skill: "c", Mode: "inject"},
	}
	rite := []SkillPolicy{
		{Skill: "b", Mode: "reference"}, // override b
		{Skill: "d", Mode: "inject"},    // new
	}
	result := MergeSkillPolicies(shared, rite)
	// Expected: a (shared), b (rite override in-place), c (shared), d (rite appended)
	if len(result) != 4 {
		t.Fatalf("expected 4 policies, got %d: %v", len(result), result)
	}
	skills := make([]string, len(result))
	for i, p := range result {
		skills[i] = p.Skill
	}
	expected := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(skills, expected) {
		t.Errorf("expected order %v, got %v", expected, skills)
	}
	if result[1].Mode != "reference" {
		t.Errorf("expected b overridden to 'reference', got %q", result[1].Mode)
	}
}

// --- applySkillPolicies tests ---

func TestApplySkillPolicies_InjectMatchesAll(t *testing.T) {
	// Policy with no requires_tools matches any agent
	fm := map[string]any{
		"skills": []any{"existing-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "new-skill", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	// new-skill should be prepended
	if len(skills) != 2 || skills[0] != "new-skill" || skills[1] != "existing-skill" {
		t.Errorf("expected [new-skill existing-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_InjectRequiresTool_Missing(t *testing.T) {
	// Agent lacks the required tool — skill should NOT be added
	fm := map[string]any{
		"tools": []any{"Read", "Write"},
	}
	policies := []SkillPolicy{
		{Skill: "bash-skill", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills injected, got %v", skills)
	}
}

func TestApplySkillPolicies_InjectRequiresTool_Present(t *testing.T) {
	// Agent has the required tool — skill should be added
	fm := map[string]any{
		"tools": []any{"Bash", "Read"},
	}
	policies := []SkillPolicy{
		{Skill: "bash-skill", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "bash-skill" {
		t.Errorf("expected [bash-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_RequiresNone_Blocked(t *testing.T) {
	// Agent has the blocked tool in disallowedTools — policy should NOT apply
	fm := map[string]any{
		"disallowedTools": []any{"Bash"},
	}
	policies := []SkillPolicy{
		{Skill: "non-bash-skill", Mode: "inject", RequiresNone: []string{"Bash"}},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills (blocked by requires_none), got %v", skills)
	}
}

func TestApplySkillPolicies_RequiresNone_NotBlocked(t *testing.T) {
	// Agent does NOT have the blocked tool — policy should apply
	fm := map[string]any{
		"disallowedTools": []any{"Write"},
	}
	policies := []SkillPolicy{
		{Skill: "non-bash-skill", Mode: "inject", RequiresNone: []string{"Bash"}},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "non-bash-skill" {
		t.Errorf("expected [non-bash-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_ExcludeSpecific(t *testing.T) {
	fm := map[string]any{
		"skill_policy_exclude": []any{"excluded-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "excluded-skill", Mode: "inject"},
		{Skill: "other-skill", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "other-skill" {
		t.Errorf("expected [other-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_ExcludeAll(t *testing.T) {
	fm := map[string]any{
		"skill_policy_exclude": "all",
		"skills":               []any{"agent-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "injected-skill", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	// Only the original agent skill should remain; nothing injected
	if len(skills) != 1 || skills[0] != "agent-skill" {
		t.Errorf("expected [agent-skill] (exclude all), got %v", skills)
	}
}

func TestApplySkillPolicies_EmptyRequiresTools_MatchesAll(t *testing.T) {
	// No requires_tools means policy matches regardless of agent tools
	fm := map[string]any{}
	policies := []SkillPolicy{
		{Skill: "universal-skill", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "universal-skill" {
		t.Errorf("expected [universal-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_Dedup_NoDouble(t *testing.T) {
	// Agent already has the skill — no duplicate should be created
	fm := map[string]any{
		"skills": []any{"ecosystem-ref"},
	}
	policies := []SkillPolicy{
		{Skill: "ecosystem-ref", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 {
		t.Errorf("expected dedup to 1 skill, got %v", skills)
	}
}

func TestApplySkillPolicies_MultipleOrdering(t *testing.T) {
	// Multiple policies — first policy's skill comes first
	fm := map[string]any{
		"skills": []any{"agent-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "first-injected", Mode: "inject"},
		{Skill: "second-injected", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	// Expected: first-injected, second-injected, agent-skill
	expected := []string{"first-injected", "second-injected", "agent-skill"}
	if !reflect.DeepEqual(skills, expected) {
		t.Errorf("expected %v, got %v", expected, skills)
	}
}

func TestApplySkillPolicies_ReferenceMode_AddsComment(t *testing.T) {
	// Reference mode prepends an HTML comment to the body
	fm := map[string]any{}
	body := []byte("# Agent Body\n")
	policies := []SkillPolicy{
		{Skill: "conventions", Mode: "reference"},
	}
	_, resultBody := applySkillPolicies(fm, body, policies)
	bodyStr := string(resultBody)
	expected := "<!-- skill_policies: conventions (invoke via Skill tool when needed) -->\n# Agent Body\n"
	if bodyStr != expected {
		t.Errorf("expected body %q, got %q", expected, bodyStr)
	}
}

func TestApplySkillPolicies_ReferenceMode_ExactFormat(t *testing.T) {
	// Verify the exact comment format produced by reference mode
	fm := map[string]any{}
	policies := []SkillPolicy{
		{Skill: "ecosystem-ref", Mode: "reference"},
	}
	_, resultBody := applySkillPolicies(fm, nil, policies)
	bodyStr := string(resultBody)
	expectedLine := "<!-- skill_policies: ecosystem-ref (invoke via Skill tool when needed) -->\n"
	if bodyStr != expectedLine {
		t.Errorf("expected exact comment line %q, got %q", expectedLine, bodyStr)
	}
}

func TestApplySkillPolicies_DeadReferenceGuard_SkillToolDisallowed(t *testing.T) {
	// Agent has Skill in disallowedTools — reference comment must NOT be added
	fm := map[string]any{
		"disallowedTools": []any{"Skill"},
	}
	body := []byte("# Agent Body\n")
	policies := []SkillPolicy{
		{Skill: "conventions", Mode: "reference"},
	}
	_, resultBody := applySkillPolicies(fm, body, policies)
	if string(resultBody) != "# Agent Body\n" {
		t.Errorf("expected body unchanged (dead reference guard), got %q", string(resultBody))
	}
}

func TestApplySkillPolicies_DeadReferenceGuard_InjectNotAffected(t *testing.T) {
	// Agent has Skill in disallowedTools — inject mode is NOT affected by dead reference guard
	fm := map[string]any{
		"disallowedTools": []any{"Skill"},
	}
	policies := []SkillPolicy{
		{Skill: "conventions", Mode: "inject"},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "conventions" {
		t.Errorf("expected [conventions] injected despite disallowedTools, got %v", skills)
	}
}

func TestApplySkillPolicies_NoopWhenEmpty(t *testing.T) {
	fm := map[string]any{
		"skills": []any{"existing"},
	}
	result, _ := applySkillPolicies(fm, nil, nil)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "existing" {
		t.Errorf("expected unchanged [existing], got %v", skills)
	}
}

func TestApplySkillPolicies_CommaSeperatedTools(t *testing.T) {
	// tools field as comma-separated string
	fm := map[string]any{
		"tools": "Bash, Read, Write",
	}
	policies := []SkillPolicy{
		{Skill: "bash-skill", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	result, _ := applySkillPolicies(fm, nil, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "bash-skill" {
		t.Errorf("expected [bash-skill] for comma-separated tools, got %v", skills)
	}
}

func TestParseToolsSet_YAMLList(t *testing.T) {
	fm := map[string]any{
		"tools": []any{"Bash", "Read"},
	}
	set := parseToolsSet(fm, "tools")
	if !set["Bash"] || !set["Read"] {
		t.Errorf("expected Bash and Read in set, got %v", set)
	}
	if set["Write"] {
		t.Error("Write should not be in set")
	}
}

func TestParseToolsSet_CommaSeparated(t *testing.T) {
	fm := map[string]any{
		"tools": "Bash, Read, Write",
	}
	set := parseToolsSet(fm, "tools")
	if !set["Bash"] || !set["Read"] || !set["Write"] {
		t.Errorf("expected Bash, Read, Write in set, got %v", set)
	}
}

func TestParseToolsSet_Missing(t *testing.T) {
	fm := map[string]any{}
	set := parseToolsSet(fm, "tools")
	if set != nil {
		t.Errorf("expected nil for missing field, got %v", set)
	}
}

// --- Agent override tests ---

func TestApplySkillPolicies_Override_ReferenceToInject(t *testing.T) {
	// Policy is reference, agent override says inject → skill added to skills:, no comment
	fm := map[string]any{
		"skill_policy_override": []any{
			map[string]any{"skill": "conventions", "mode": "inject"},
		},
	}
	body := []byte("# Body\n")
	policies := []SkillPolicy{
		{Skill: "conventions", Mode: "reference"},
	}
	result, resultBody := applySkillPolicies(fm, body, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "conventions" {
		t.Errorf("expected [conventions] injected via override, got %v", skills)
	}
	// No comment should have been added
	if string(resultBody) != "# Body\n" {
		t.Errorf("expected body unchanged (no reference comment), got %q", string(resultBody))
	}
}

func TestApplySkillPolicies_Override_InjectToReference(t *testing.T) {
	// Policy is inject, agent override says reference → comment added, not in skills:
	fm := map[string]any{
		"skill_policy_override": []any{
			map[string]any{"skill": "conventions", "mode": "reference"},
		},
	}
	body := []byte("# Body\n")
	policies := []SkillPolicy{
		{Skill: "conventions", Mode: "inject"},
	}
	result, resultBody := applySkillPolicies(fm, body, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills in frontmatter (overridden to reference), got %v", skills)
	}
	expectedBody := "<!-- skill_policies: conventions (invoke via Skill tool when needed) -->\n# Body\n"
	if string(resultBody) != expectedBody {
		t.Errorf("expected reference comment in body, got %q", string(resultBody))
	}
}

func TestApplySkillPolicies_ExcludeWinsOverOverride(t *testing.T) {
	// Agent both excludes and overrides same skill — exclude wins, skill skipped entirely
	fm := map[string]any{
		"skill_policy_exclude": []any{"conventions"},
		"skill_policy_override": []any{
			map[string]any{"skill": "conventions", "mode": "inject"},
		},
	}
	body := []byte("# Body\n")
	policies := []SkillPolicy{
		{Skill: "conventions", Mode: "reference"},
	}
	result, resultBody := applySkillPolicies(fm, body, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills (exclude wins over override), got %v", skills)
	}
	if string(resultBody) != "# Body\n" {
		t.Errorf("expected body unchanged (excluded), got %q", string(resultBody))
	}
}

func TestApplySkillPolicies_MixedModes(t *testing.T) {
	// Some policies inject, some reference → correct outputs for each
	fm := map[string]any{}
	body := []byte("# Body\n")
	policies := []SkillPolicy{
		{Skill: "inject-skill", Mode: "inject"},
		{Skill: "ref-skill", Mode: "reference"},
		{Skill: "another-inject", Mode: "inject"},
	}
	result, resultBody := applySkillPolicies(fm, body, policies)

	// inject-skill and another-inject should be in frontmatter
	skills := toStringSlice(result["skills"])
	expectedSkills := []string{"inject-skill", "another-inject"}
	if !reflect.DeepEqual(skills, expectedSkills) {
		t.Errorf("expected injected skills %v, got %v", expectedSkills, skills)
	}

	// ref-skill should appear as HTML comment in body
	bodyStr := string(resultBody)
	expectedComment := "<!-- skill_policies: ref-skill (invoke via Skill tool when needed) -->\n"
	if bodyStr != expectedComment+"# Body\n" {
		t.Errorf("expected body with ref comment %q, got %q", expectedComment+"# Body\n", bodyStr)
	}
}

func TestApplySkillPolicies_MultipleReferenceComments_Order(t *testing.T) {
	// Multiple reference policies — all comments prepended in policy order
	fm := map[string]any{}
	body := []byte("# Body\n")
	policies := []SkillPolicy{
		{Skill: "first-ref", Mode: "reference"},
		{Skill: "second-ref", Mode: "reference"},
	}
	_, resultBody := applySkillPolicies(fm, body, policies)
	bodyStr := string(resultBody)
	expected := "<!-- skill_policies: first-ref (invoke via Skill tool when needed) -->\n" +
		"<!-- skill_policies: second-ref (invoke via Skill tool when needed) -->\n" +
		"# Body\n"
	if bodyStr != expected {
		t.Errorf("expected two reference comments in order, got %q", bodyStr)
	}
}
