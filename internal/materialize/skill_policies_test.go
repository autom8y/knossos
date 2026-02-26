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

func fmWithTools(tools []interface{}, skills []interface{}) map[string]interface{} {
	fm := map[string]interface{}{}
	if tools != nil {
		fm["tools"] = tools
	}
	if skills != nil {
		fm["skills"] = skills
	}
	return fm
}

func TestApplySkillPolicies_InjectMatchesAll(t *testing.T) {
	// Policy with no requires_tools matches any agent
	fm := map[string]interface{}{
		"skills": []interface{}{"existing-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "new-skill", Mode: "inject"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	// new-skill should be prepended
	if len(skills) != 2 || skills[0] != "new-skill" || skills[1] != "existing-skill" {
		t.Errorf("expected [new-skill existing-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_InjectRequiresTool_Missing(t *testing.T) {
	// Agent lacks the required tool — skill should NOT be added
	fm := map[string]interface{}{
		"tools": []interface{}{"Read", "Write"},
	}
	policies := []SkillPolicy{
		{Skill: "bash-skill", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills injected, got %v", skills)
	}
}

func TestApplySkillPolicies_InjectRequiresTool_Present(t *testing.T) {
	// Agent has the required tool — skill should be added
	fm := map[string]interface{}{
		"tools": []interface{}{"Bash", "Read"},
	}
	policies := []SkillPolicy{
		{Skill: "bash-skill", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "bash-skill" {
		t.Errorf("expected [bash-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_RequiresNone_Blocked(t *testing.T) {
	// Agent has the blocked tool in disallowedTools — policy should NOT apply
	fm := map[string]interface{}{
		"disallowedTools": []interface{}{"Bash"},
	}
	policies := []SkillPolicy{
		{Skill: "non-bash-skill", Mode: "inject", RequiresNone: []string{"Bash"}},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills (blocked by requires_none), got %v", skills)
	}
}

func TestApplySkillPolicies_RequiresNone_NotBlocked(t *testing.T) {
	// Agent does NOT have the blocked tool — policy should apply
	fm := map[string]interface{}{
		"disallowedTools": []interface{}{"Write"},
	}
	policies := []SkillPolicy{
		{Skill: "non-bash-skill", Mode: "inject", RequiresNone: []string{"Bash"}},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "non-bash-skill" {
		t.Errorf("expected [non-bash-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_ExcludeSpecific(t *testing.T) {
	fm := map[string]interface{}{
		"skill_policy_exclude": []interface{}{"excluded-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "excluded-skill", Mode: "inject"},
		{Skill: "other-skill", Mode: "inject"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "other-skill" {
		t.Errorf("expected [other-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_ExcludeAll(t *testing.T) {
	fm := map[string]interface{}{
		"skill_policy_exclude": "all",
		"skills":               []interface{}{"agent-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "injected-skill", Mode: "inject"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	// Only the original agent skill should remain; nothing injected
	if len(skills) != 1 || skills[0] != "agent-skill" {
		t.Errorf("expected [agent-skill] (exclude all), got %v", skills)
	}
}

func TestApplySkillPolicies_EmptyRequiresTools_MatchesAll(t *testing.T) {
	// No requires_tools means policy matches regardless of agent tools
	fm := map[string]interface{}{}
	policies := []SkillPolicy{
		{Skill: "universal-skill", Mode: "inject"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "universal-skill" {
		t.Errorf("expected [universal-skill], got %v", skills)
	}
}

func TestApplySkillPolicies_Dedup_NoDouble(t *testing.T) {
	// Agent already has the skill — no duplicate should be created
	fm := map[string]interface{}{
		"skills": []interface{}{"ecosystem-ref"},
	}
	policies := []SkillPolicy{
		{Skill: "ecosystem-ref", Mode: "inject"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 {
		t.Errorf("expected dedup to 1 skill, got %v", skills)
	}
}

func TestApplySkillPolicies_MultipleOrdering(t *testing.T) {
	// Multiple policies — first policy's skill comes first
	fm := map[string]interface{}{
		"skills": []interface{}{"agent-skill"},
	}
	policies := []SkillPolicy{
		{Skill: "first-injected", Mode: "inject"},
		{Skill: "second-injected", Mode: "inject"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	// Expected: first-injected, second-injected, agent-skill
	expected := []string{"first-injected", "second-injected", "agent-skill"}
	if !reflect.DeepEqual(skills, expected) {
		t.Errorf("expected %v, got %v", expected, skills)
	}
}

func TestApplySkillPolicies_ReferenceModeSkipped(t *testing.T) {
	// Reference mode policies are skipped in Sprint 2
	fm := map[string]interface{}{}
	policies := []SkillPolicy{
		{Skill: "ref-skill", Mode: "reference"},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 0 {
		t.Errorf("expected no skills injected for reference mode, got %v", skills)
	}
}

func TestApplySkillPolicies_NoopWhenEmpty(t *testing.T) {
	fm := map[string]interface{}{
		"skills": []interface{}{"existing"},
	}
	result := applySkillPolicies(fm, nil)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "existing" {
		t.Errorf("expected unchanged [existing], got %v", skills)
	}
}

func TestApplySkillPolicies_CommaSeperatedTools(t *testing.T) {
	// tools field as comma-separated string
	fm := map[string]interface{}{
		"tools": "Bash, Read, Write",
	}
	policies := []SkillPolicy{
		{Skill: "bash-skill", Mode: "inject", RequiresTools: []string{"Bash"}},
	}
	result := applySkillPolicies(fm, policies)
	skills := toStringSlice(result["skills"])
	if len(skills) != 1 || skills[0] != "bash-skill" {
		t.Errorf("expected [bash-skill] for comma-separated tools, got %v", skills)
	}
}

func TestParseToolsSet_YAMLList(t *testing.T) {
	fm := map[string]interface{}{
		"tools": []interface{}{"Bash", "Read"},
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
	fm := map[string]interface{}{
		"tools": "Bash, Read, Write",
	}
	set := parseToolsSet(fm, "tools")
	if !set["Bash"] || !set["Read"] || !set["Write"] {
		t.Errorf("expected Bash, Read, Write in set, got %v", set)
	}
}

func TestParseToolsSet_Missing(t *testing.T) {
	fm := map[string]interface{}{}
	set := parseToolsSet(fm, "tools")
	if set != nil {
		t.Errorf("expected nil for missing field, got %v", set)
	}
}
