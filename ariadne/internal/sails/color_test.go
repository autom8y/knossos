package sails

import (
	"testing"
	"time"
)

// Test cases from TDD Section 11.2:
// - sails_001: WHITE with all proofs passing
// - sails_002: GRAY with open questions
// - sails_003: GRAY with missing proofs
// - sails_004: BLACK with failing tests
// - sails_005: Spike always GRAY
// - sails_006: Hotfix always GRAY
// - sails_007: Human downgrade override
// - sails_008: QA upgrade gray to white
// - sails_009: Cannot self-upgrade

// TestColor_String verifies Color.String() returns expected values.
func TestColor_String(t *testing.T) {
	tests := []struct {
		color    Color
		expected string
	}{
		{ColorWhite, "WHITE"},
		{ColorGray, "GRAY"},
		{ColorBlack, "BLACK"},
	}

	for _, tt := range tests {
		if got := tt.color.String(); got != tt.expected {
			t.Errorf("Color(%q).String() = %q, want %q", tt.color, got, tt.expected)
		}
	}
}

// TestColor_IsValid verifies Color.IsValid() returns correct values.
func TestColor_IsValid(t *testing.T) {
	tests := []struct {
		color Color
		valid bool
	}{
		{ColorWhite, true},
		{ColorGray, true},
		{ColorBlack, true},
		{Color("YELLOW"), false},
		{Color(""), false},
	}

	for _, tt := range tests {
		if got := tt.color.IsValid(); got != tt.valid {
			t.Errorf("Color(%q).IsValid() = %v, want %v", tt.color, got, tt.valid)
		}
	}
}

// TestProofStatus_IsPassing verifies IsPassing() behavior.
func TestProofStatus_IsPassing(t *testing.T) {
	tests := []struct {
		status  ProofStatus
		passing bool
	}{
		{ProofPass, true},
		{ProofSkip, true},
		{ProofFail, false},
		{ProofUnknown, false},
	}

	for _, tt := range tests {
		if got := tt.status.IsPassing(); got != tt.passing {
			t.Errorf("ProofStatus(%q).IsPassing() = %v, want %v", tt.status, got, tt.passing)
		}
	}
}

// TestModifierType_IsValid verifies ModifierType.IsValid() behavior.
func TestModifierType_IsValid(t *testing.T) {
	tests := []struct {
		modType ModifierType
		valid   bool
	}{
		{ModifierDowngradeToGray, true},
		{ModifierDowngradeToBlack, true},
		{ModifierHumanOverrideGray, true},
		{ModifierType("UPGRADE_TO_WHITE"), false},
		{ModifierType(""), false},
	}

	for _, tt := range tests {
		if got := tt.modType.IsValid(); got != tt.valid {
			t.Errorf("ModifierType(%q).IsValid() = %v, want %v", tt.modType, got, tt.valid)
		}
	}
}

// TestGetRequiredProofs verifies required proofs per complexity level.
func TestGetRequiredProofs(t *testing.T) {
	tests := []struct {
		complexity string
		expected   []string
	}{
		{"PATCH", []string{"tests", "build", "lint"}},
		{"SCRIPT", []string{"tests", "build", "lint"}},
		{"MODULE", []string{"tests", "build", "lint"}},
		{"SERVICE", []string{"tests", "build", "lint"}},
		{"INITIATIVE", []string{"tests", "build", "lint", "adversarial", "integration"}},
		{"MIGRATION", []string{"tests", "build", "lint", "adversarial", "integration"}},
		{"PLATFORM", []string{"tests", "build", "lint", "adversarial", "integration"}},
		{"UNKNOWN", []string{"tests", "build", "lint", "adversarial", "integration"}}, // Defaults to strictest
	}

	for _, tt := range tests {
		got := GetRequiredProofs(tt.complexity)
		if len(got) != len(tt.expected) {
			t.Errorf("GetRequiredProofs(%q) = %v, want %v", tt.complexity, got, tt.expected)
			continue
		}
		for i, exp := range tt.expected {
			if got[i] != exp {
				t.Errorf("GetRequiredProofs(%q)[%d] = %q, want %q", tt.complexity, i, got[i], exp)
			}
		}
	}
}

// TestIsRequiredProof verifies IsRequiredProof behavior.
func TestIsRequiredProof(t *testing.T) {
	tests := []struct {
		complexity string
		proofName  string
		required   bool
	}{
		{"MODULE", "tests", true},
		{"MODULE", "build", true},
		{"MODULE", "lint", true},
		{"MODULE", "adversarial", false},
		{"MODULE", "integration", false},
		{"INITIATIVE", "adversarial", true},
		{"INITIATIVE", "integration", true},
		{"PLATFORM", "adversarial", true},
		{"SERVICE", "adversarial", false}, // Recommended, not required
	}

	for _, tt := range tests {
		got := IsRequiredProof(tt.complexity, tt.proofName)
		if got != tt.required {
			t.Errorf("IsRequiredProof(%q, %q) = %v, want %v", tt.complexity, tt.proofName, got, tt.required)
		}
	}
}

// sails_001: WHITE with all proofs passing
func TestComputeColor_sails_001_WhiteWithAllProofsPassing(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass, Summary: "47 tests passed"},
			"build": {Status: ProofPass, Summary: "go build succeeded"},
			"lint":  {Status: ProofPass, Summary: "golangci-lint clean"},
		},
		OpenQuestions: []string{},
		Modifiers:     []Modifier{},
	}

	result := ComputeColor(input)

	if result.Color != ColorWhite {
		t.Errorf("sails_001: Color = %q, want %q", result.Color, ColorWhite)
	}
	if result.ComputedBase != ColorWhite {
		t.Errorf("sails_001: ComputedBase = %q, want %q", result.ComputedBase, ColorWhite)
	}
	if len(result.Reasons) == 0 {
		t.Error("sails_001: Reasons should not be empty")
	}
}

// sails_001 variant: WHITE with SKIP status on optional proof
func TestComputeColor_WhiteWithSkippedProof(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofSkip, Summary: "Skipped: no build step needed"},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorWhite {
		t.Errorf("Color = %q, want %q (SKIP should count as passing)", result.Color, ColorWhite)
	}
}

// sails_002: GRAY with open questions
func TestComputeColor_sails_002_GrayWithOpenQuestions(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{
			"How should rate limiting behave under cluster failover?",
			"Need to validate with Production DBA on index strategy",
		},
		Modifiers: []Modifier{},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("sails_002: Color = %q, want %q", result.Color, ColorGray)
	}
	if result.ComputedBase != ColorGray {
		t.Errorf("sails_002: ComputedBase = %q, want %q", result.ComputedBase, ColorGray)
	}

	// Should mention open questions in reasons
	found := false
	for _, reason := range result.Reasons {
		if reason == "open questions present: gray ceiling applied" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sails_002: Reasons should mention open questions, got: %v", result.Reasons)
	}
}

// sails_003: GRAY with missing proofs
func TestComputeColor_sails_003_GrayWithMissingProofs(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			// lint is missing
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("sails_003: Color = %q, want %q", result.Color, ColorGray)
	}
	if result.ComputedBase != ColorGray {
		t.Errorf("sails_003: ComputedBase = %q, want %q", result.ComputedBase, ColorGray)
	}

	// Should mention missing proof in reasons
	found := false
	for _, reason := range result.Reasons {
		if reason == "required proof 'lint' is missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sails_003: Reasons should mention missing lint proof, got: %v", result.Reasons)
	}
}

// sails_003 variant: GRAY with UNKNOWN status
func TestComputeColor_GrayWithUnknownStatus(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofUnknown}, // UNKNOWN = not passing
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("Color = %q, want %q (UNKNOWN should not pass)", result.Color, ColorGray)
	}
}

// sails_004: BLACK with failing tests
func TestComputeColor_sails_004_BlackWithFailingTests(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofFail, Summary: "5 tests failed"},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorBlack {
		t.Errorf("sails_004: Color = %q, want %q", result.Color, ColorBlack)
	}
	if result.ComputedBase != ColorBlack {
		t.Errorf("sails_004: ComputedBase = %q, want %q", result.ComputedBase, ColorBlack)
	}

	// Should mention failing proof in reasons
	found := false
	for _, reason := range result.Reasons {
		if reason == "proof 'tests' has status FAIL" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sails_004: Reasons should mention failing tests, got: %v", result.Reasons)
	}
}

// sails_004 variant: BLACK takes priority over open questions
func TestComputeColor_BlackTakesPriority(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofFail},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{"This question should not prevent BLACK"},
	}

	result := ComputeColor(input)

	if result.Color != ColorBlack {
		t.Errorf("Color = %q, want %q (BLACK should take priority over GRAY)", result.Color, ColorBlack)
	}
}

// sails_005: Spike always GRAY
func TestComputeColor_sails_005_SpikeAlwaysGray(t *testing.T) {
	input := ColorInput{
		SessionType: "spike",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("sails_005: Color = %q, want %q", result.Color, ColorGray)
	}
	if result.ComputedBase != ColorGray {
		t.Errorf("sails_005: ComputedBase = %q, want %q", result.ComputedBase, ColorGray)
	}

	// Should mention spike ceiling in reasons
	found := false
	for _, reason := range result.Reasons {
		if reason == "session type 'spike' has gray ceiling (spikes never white)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sails_005: Reasons should mention spike ceiling, got: %v", result.Reasons)
	}
}

// sails_005 variant: Case-insensitive spike detection
func TestComputeColor_SpikeCaseInsensitive(t *testing.T) {
	variants := []string{"spike", "SPIKE", "Spike"}

	for _, variant := range variants {
		input := ColorInput{
			SessionType: variant,
			Complexity:  "MODULE",
			Proofs: map[string]Proof{
				"tests": {Status: ProofPass},
				"build": {Status: ProofPass},
				"lint":  {Status: ProofPass},
			},
		}

		result := ComputeColor(input)

		if result.Color != ColorGray {
			t.Errorf("SessionType %q: Color = %q, want %q", variant, result.Color, ColorGray)
		}
	}
}

// sails_006: Hotfix always GRAY
func TestComputeColor_sails_006_HotfixAlwaysGray(t *testing.T) {
	input := ColorInput{
		SessionType: "hotfix",
		Complexity:  "PATCH",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("sails_006: Color = %q, want %q", result.Color, ColorGray)
	}
	if result.ComputedBase != ColorGray {
		t.Errorf("sails_006: ComputedBase = %q, want %q", result.ComputedBase, ColorGray)
	}

	// Should mention hotfix ceiling in reasons
	found := false
	for _, reason := range result.Reasons {
		if reason == "session type 'hotfix' has gray ceiling (expedited gray)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sails_006: Reasons should mention hotfix ceiling, got: %v", result.Reasons)
	}
}

// sails_007: Human downgrade override
func TestComputeColor_sails_007_HumanDowngradeOverride(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
		Modifiers: []Modifier{
			{
				Type:          ModifierDowngradeToGray,
				Justification: "Changed retry logic in payment flow; want senior review",
				AppliedBy:     "human",
				Timestamp:     &now,
			},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("sails_007: Color = %q, want %q", result.Color, ColorGray)
	}
	if result.ComputedBase != ColorWhite {
		t.Errorf("sails_007: ComputedBase = %q, want %q (before modifier)", result.ComputedBase, ColorWhite)
	}
}

// sails_007 variant: DOWNGRADE_TO_BLACK
func TestComputeColor_DowngradeToBlack(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		Modifiers: []Modifier{
			{
				Type:          ModifierDowngradeToBlack,
				Justification: "Critical security issue discovered post-review",
				AppliedBy:     "human",
			},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorBlack {
		t.Errorf("Color = %q, want %q", result.Color, ColorBlack)
	}
	if result.ComputedBase != ColorWhite {
		t.Errorf("ComputedBase = %q, want %q (before modifier)", result.ComputedBase, ColorWhite)
	}
}

// sails_007 variant: HUMAN_OVERRIDE_GRAY
func TestComputeColor_HumanOverrideGray(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		Modifiers: []Modifier{
			{
				Type:          ModifierHumanOverrideGray,
				Justification: "Forcing gray for team review process",
				AppliedBy:     "human",
			},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("Color = %q, want %q", result.Color, ColorGray)
	}
}

// sails_008: QA upgrade gray to white
func TestComputeColor_sails_008_QAUpgradeGrayToWhite(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		// Open questions cause gray base
		OpenQuestions: []string{"Initial question that was resolved by QA"},
		QAUpgrade: &QAUpgrade{
			UpgradedAt:              &now,
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/TP-qa-original-session.md",
			AdversarialTestsAdded: []string{
				"tests/integration/rate_limit_failover_test.go",
				"tests/integration/index_edge_cases_test.go",
			},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorWhite {
		t.Errorf("sails_008: Color = %q, want %q", result.Color, ColorWhite)
	}
	if result.ComputedBase != ColorGray {
		t.Errorf("sails_008: ComputedBase = %q, want %q", result.ComputedBase, ColorGray)
	}

	// Should mention QA upgrade in reasons
	found := false
	for _, reason := range result.Reasons {
		if reason == "QA upgrade applied: gray -> white via QA session session-20260106-100000-qa123456" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sails_008: Reasons should mention QA upgrade, got: %v", result.Reasons)
	}
}

// sails_009: Cannot self-upgrade (no upgrade modifier exists)
func TestComputeColor_sails_009_CannotSelfUpgrade(t *testing.T) {
	// Verify that there's no way to upgrade without QA
	// Modifiers can only downgrade
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{"Unresolved question"},
		// No QA upgrade, just modifiers - they cannot upgrade
		Modifiers: []Modifier{
			{
				Type:          ModifierDowngradeToGray, // Even this can't upgrade
				Justification: "Attempting to game the system",
				AppliedBy:     "agent",
			},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("sails_009: Color = %q, want %q (cannot self-upgrade)", result.Color, ColorGray)
	}
}

// sails_009 variant: QA upgrade without constraint log fails
func TestComputeColor_QAUpgradeWithoutConstraintLog(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{"Question"},
		QAUpgrade: &QAUpgrade{
			UpgradedAt:              &now,
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "", // Missing!
			AdversarialTestsAdded:   []string{"test.go"},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("Color = %q, want %q (missing constraint log)", result.Color, ColorGray)
	}
}

// sails_009 variant: QA upgrade without adversarial tests fails
func TestComputeColor_QAUpgradeWithoutAdversarialTests(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{"Question"},
		QAUpgrade: &QAUpgrade{
			UpgradedAt:              &now,
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/log.md",
			AdversarialTestsAdded:   []string{}, // Empty!
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("Color = %q, want %q (no adversarial tests)", result.Color, ColorGray)
	}
}

// Test QA upgrade doesn't apply to already WHITE sessions
func TestComputeColor_QAUpgradeNotAppliedToWhite(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{}, // No questions = WHITE base
		QAUpgrade: &QAUpgrade{
			UpgradedAt:              &now,
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/log.md",
			AdversarialTestsAdded:   []string{"test.go"},
		},
	}

	result := ComputeColor(input)

	// Should still be WHITE, but QA upgrade is not "applied" (base was already WHITE)
	if result.Color != ColorWhite {
		t.Errorf("Color = %q, want %q", result.Color, ColorWhite)
	}
	if result.ComputedBase != ColorWhite {
		t.Errorf("ComputedBase = %q, want %q", result.ComputedBase, ColorWhite)
	}
}

// Test QA upgrade doesn't apply to BLACK sessions
func TestComputeColor_QAUpgradeNotAppliedToBlack(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofFail}, // BLACK
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		QAUpgrade: &QAUpgrade{
			UpgradedAt:              &now,
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/log.md",
			AdversarialTestsAdded:   []string{"test.go"},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorBlack {
		t.Errorf("Color = %q, want %q (cannot upgrade from BLACK)", result.Color, ColorBlack)
	}
}

// Test QA upgrade doesn't apply if color was downgraded to BLACK by modifier
func TestComputeColor_QAUpgradeNotAppliedWhenDowngradedToBlack(t *testing.T) {
	now := time.Now()
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{"Question"}, // GRAY base
		Modifiers: []Modifier{
			{
				Type:          ModifierDowngradeToBlack,
				Justification: "Security issue",
				AppliedBy:     "human",
			},
		},
		QAUpgrade: &QAUpgrade{
			UpgradedAt:              &now,
			QASessionID:             "session-20260106-100000-qa123456",
			ConstraintResolutionLog: "docs/testing/log.md",
			AdversarialTestsAdded:   []string{"test.go"},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorBlack {
		t.Errorf("Color = %q, want %q (downgraded to BLACK blocks QA upgrade)", result.Color, ColorBlack)
	}
}

// Test high complexity requires additional proofs
func TestComputeColor_HighComplexityRequiresAdditionalProofs(t *testing.T) {
	// PLATFORM complexity requires adversarial and integration proofs
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "PLATFORM",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
			// Missing: adversarial, integration
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorGray {
		t.Errorf("Color = %q, want %q (missing adversarial/integration)", result.Color, ColorGray)
	}
}

// Test high complexity with all required proofs
func TestComputeColor_HighComplexityWithAllProofs(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "PLATFORM",
		Proofs: map[string]Proof{
			"tests":       {Status: ProofPass},
			"build":       {Status: ProofPass},
			"lint":        {Status: ProofPass},
			"adversarial": {Status: ProofPass},
			"integration": {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorWhite {
		t.Errorf("Color = %q, want %q", result.Color, ColorWhite)
	}
}

// Test SERVICE complexity doesn't require adversarial (only recommended)
func TestComputeColor_ServiceDoesNotRequireAdversarial(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "SERVICE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
			// adversarial is only recommended, not required
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorWhite {
		t.Errorf("Color = %q, want %q (adversarial not required for SERVICE)", result.Color, ColorWhite)
	}
}

// Test empty session type defaults to standard
func TestComputeColor_EmptySessionTypeDefaults(t *testing.T) {
	input := ColorInput{
		SessionType: "", // Empty
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	result := ComputeColor(input)

	if result.Color != ColorWhite {
		t.Errorf("Color = %q, want %q (empty session type = standard)", result.Color, ColorWhite)
	}
}

// Test multiple modifiers are applied in order
func TestComputeColor_MultipleModifiers(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		Modifiers: []Modifier{
			{
				Type:          ModifierDowngradeToGray,
				Justification: "First downgrade",
				AppliedBy:     "human",
			},
			{
				Type:          ModifierDowngradeToBlack,
				Justification: "Second downgrade to black",
				AppliedBy:     "agent",
			},
		},
	}

	result := ComputeColor(input)

	if result.Color != ColorBlack {
		t.Errorf("Color = %q, want %q (multiple modifiers applied)", result.Color, ColorBlack)
	}
}

// Test ValidateColorInput with invalid proof status
func TestValidateColorInput_InvalidProofStatus(t *testing.T) {
	input := ColorInput{
		Proofs: map[string]Proof{
			"tests": {Status: ProofStatus("INVALID")},
		},
	}

	errors := ValidateColorInput(input)

	if len(errors) == 0 {
		t.Error("Expected validation error for invalid proof status")
	}
}

// Test ValidateColorInput with invalid modifier
func TestValidateColorInput_InvalidModifier(t *testing.T) {
	input := ColorInput{
		Modifiers: []Modifier{
			{
				Type:          ModifierType("UPGRADE"),
				Justification: "",
				AppliedBy:     "unknown",
			},
		},
	}

	errors := ValidateColorInput(input)

	if len(errors) < 3 {
		t.Errorf("Expected at least 3 validation errors, got %d: %v", len(errors), errors)
	}
}

// Test ValidateColorInput with valid input
func TestValidateColorInput_ValidInput(t *testing.T) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		Modifiers: []Modifier{
			{
				Type:          ModifierDowngradeToGray,
				Justification: "Valid justification",
				AppliedBy:     "human",
			},
		},
	}

	errors := ValidateColorInput(input)

	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}
}

// Benchmark color computation
func BenchmarkComputeColor(b *testing.B) {
	input := ColorInput{
		SessionType: "standard",
		Complexity:  "MODULE",
		Proofs: map[string]Proof{
			"tests": {Status: ProofPass},
			"build": {Status: ProofPass},
			"lint":  {Status: ProofPass},
		},
		OpenQuestions: []string{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeColor(input)
	}
}
