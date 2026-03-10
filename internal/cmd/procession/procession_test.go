package procession

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	procmena "github.com/autom8y/knossos/internal/materialize/procession"
	"github.com/autom8y/knossos/internal/session"
)

// ---- test helpers ----

// testEnv sets up a minimal project + session directory structure.
// Returns projectDir, sessionID, and the path to SESSION_CONTEXT.md.
func testEnv(t *testing.T) (projectDir, sessionID, ctxPath string) {
	t.Helper()
	tmpDir := t.TempDir()
	projectDir = tmpDir
	sessionID = "session-20260310-100000-proctest1"

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")

	for _, dir := range []string{sessionDir, locksDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
	}

	// Write .current-session so GetSessionID resolves correctly
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("write .current-session: %v", err)
	}

	ctxPath = filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	return projectDir, sessionID, ctxPath
}

// writeSession writes a minimal SESSION_CONTEXT.md without procession.
func writeSession(t *testing.T, ctxPath string) {
	t.Helper()
	content := `---
schema_version: "2.3"
session_id: ` + filepath.Base(filepath.Dir(ctxPath)) + `
status: ACTIVE
initiative: Test initiative
complexity: MODULE
active_rite: security
current_phase: requirements
created_at: "2026-03-10T10:00:00Z"
---

# Session: Test initiative
`
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("write SESSION_CONTEXT.md: %v", err)
	}
}

// writeSessionWithProcession writes SESSION_CONTEXT.md with an embedded procession block.
func writeSessionWithProcession(t *testing.T, ctxPath string, proc *session.Procession) {
	t.Helper()

	// Handle nil procession (post-completion state)
	if proc == nil {
		content := `---
schema_version: "2.3"
session_id: ` + filepath.Base(filepath.Dir(ctxPath)) + `
status: ACTIVE
initiative: Test initiative
complexity: MODULE
active_rite: security
current_phase: requirements
created_at: "2026-03-10T10:00:00Z"
---
`
		if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
			t.Fatalf("write session context: %v", err)
		}
		return
	}

	// Build completed_stations YAML block
	completedBlock := ""
	if len(proc.CompletedStations) > 0 {
		var lines []string
		for _, cs := range proc.CompletedStations {
			lines = append(lines, "    - station: "+cs.Station)
			lines = append(lines, "      rite: "+cs.Rite)
			lines = append(lines, "      completed_at: \""+cs.CompletedAt+"\"")
			if len(cs.Artifacts) > 0 {
				lines = append(lines, "      artifacts:")
				for _, a := range cs.Artifacts {
					lines = append(lines, "        - "+a)
				}
			}
		}
		completedBlock = "  completed_stations:\n" + strings.Join(lines, "\n") + "\n"
	}

	nextStationBlock := ""
	if proc.NextStation != "" {
		nextStationBlock += "  next_station: " + proc.NextStation + "\n"
	}
	if proc.NextRite != "" {
		nextStationBlock += "  next_rite: " + proc.NextRite + "\n"
	}

	content := `---
schema_version: "2.3"
session_id: ` + filepath.Base(filepath.Dir(ctxPath)) + `
status: ACTIVE
initiative: Test initiative
complexity: MODULE
active_rite: security
current_phase: requirements
created_at: "2026-03-10T10:00:00Z"
procession:
  id: ` + proc.ID + `
  type: ` + proc.Type + `
  current_station: ` + proc.CurrentStation + `
` + completedBlock + nextStationBlock + `  artifact_dir: ` + proc.ArtifactDir + `
---

# Session: Test initiative
`
	if err := os.WriteFile(ctxPath, []byte(content), 0644); err != nil {
		t.Fatalf("write SESSION_CONTEXT.md with procession: %v", err)
	}
}

// writeTemplate writes a minimal security-remediation template YAML to projectDir/processions/.
func writeTemplate(t *testing.T, projectDir string) {
	t.Helper()
	dir := filepath.Join(projectDir, "processions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll processions/: %v", err)
	}
	content := `name: security-remediation
description: "Security findings lifecycle: audit, assess, plan, remediate, validate"
stations:
  - name: audit
    rite: security
    goal: "Map attack surface and produce threat model"
    produces: [threat-model, pentest-report]
  - name: assess
    rite: debt-triage
    goal: "Catalog findings and score risk"
    produces: [debt-inventory, priority-matrix]
  - name: plan
    rite: debt-triage
    goal: "Group findings into sprint-sized tasks"
    produces: [sprint-plan]
  - name: remediate
    rite: hygiene
    alt_rite: 10x-dev
    goal: "Execute remediation plan"
    produces: [remediation-ledger]
  - name: validate
    rite: security
    goal: "Review remediation PRs for security correctness"
    produces: [validation-report]
    loop_to: remediate
artifact_dir: .sos/wip/security-remediation/
`
	if err := os.WriteFile(filepath.Join(dir, "security-remediation.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write template: %v", err)
	}
}

// newTestCtx builds a cmdContext pointing at projectDir with the given sessionID.
func newTestCtx(projectDir, sessionID string, opts ...ProcessionResolveFunc) *cmdContext {
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}
	if len(opts) > 0 {
		ctx.resolveFunc = opts[0]
	}
	return ctx
}

// captureOutput captures stdout while running f.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

// ---- tests ----

// TestCreate_HappyPath verifies that create with a valid template produces the expected output
// and writes the procession block to the session context.
func TestCreate_HappyPath(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeSession(t, ctxPath)
	writeTemplate(t, projectDir)

	ctx := newTestCtx(projectDir, sessionID)
	opts := createOptions{templateName: "security-remediation"}

	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runCreate(ctx, opts)
	})
	if runErr != nil {
		t.Fatalf("runCreate failed: %v", runErr)
	}

	// Parse JSON output
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	// Verify key fields
	if result["type"] != "security-remediation" {
		t.Errorf("type = %v, want security-remediation", result["type"])
	}
	if result["current_station"] != "audit" {
		t.Errorf("current_station = %v, want audit", result["current_station"])
	}
	if result["next_station"] != "assess" {
		t.Errorf("next_station = %v, want assess", result["next_station"])
	}
	if result["next_rite"] != "debt-triage" {
		t.Errorf("next_rite = %v, want debt-triage", result["next_rite"])
	}
	if result["artifact_dir"] != ".sos/wip/security-remediation/" {
		t.Errorf("artifact_dir = %v, want .sos/wip/security-remediation/", result["artifact_dir"])
	}

	// Verify procession ID contains template name and today's date
	id, _ := result["procession_id"].(string)
	if !strings.HasPrefix(id, "security-remediation-") {
		t.Errorf("procession_id = %q, want prefix security-remediation-", id)
	}

	// Verify session context was updated
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if sessCtx.Procession == nil {
		t.Fatal("Procession is nil in session context after create")
	}
	if sessCtx.Procession.Type != "security-remediation" {
		t.Errorf("Procession.Type = %q, want security-remediation", sessCtx.Procession.Type)
	}
	if sessCtx.Procession.CurrentStation != "audit" {
		t.Errorf("Procession.CurrentStation = %q, want audit", sessCtx.Procession.CurrentStation)
	}

	// Verify artifact directory was created
	artifactDir := filepath.Join(projectDir, ".sos", "wip", "security-remediation")
	if _, err := os.Stat(artifactDir); os.IsNotExist(err) {
		t.Errorf("artifact directory not created: %s", artifactDir)
	}
}

// TestCreate_AlreadyHasProcession verifies that create fails when a procession is already active.
func TestCreate_AlreadyHasProcession(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "audit",
		NextStation:    "assess",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	ctx := newTestCtx(projectDir, sessionID)
	opts := createOptions{templateName: "security-remediation"}
	err := runCreate(ctx, opts)
	if err == nil {
		t.Error("expected error when procession already active, got nil")
	}
}

// TestCreate_MissingTemplate verifies that create fails with a clear error for a missing template.
func TestCreate_MissingTemplate(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeSession(t, ctxPath)
	// Do NOT write template file

	ctx := newTestCtx(projectDir, sessionID)
	opts := createOptions{templateName: "nonexistent"}
	err := runCreate(ctx, opts)
	if err == nil {
		t.Error("expected error for missing template, got nil")
	}
}

// TestStatus_WithProcession verifies status output when a procession is active.
func TestStatus_WithProcession(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "assess",
		NextStation:    "plan",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	ctx := newTestCtx(projectDir, sessionID)
	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runStatus(ctx)
	})
	if runErr != nil {
		t.Fatalf("runStatus failed: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result["procession_id"] != "security-remediation-2026-03-10" {
		t.Errorf("procession_id = %v, want security-remediation-2026-03-10", result["procession_id"])
	}
	if result["current_station"] != "assess" {
		t.Errorf("current_station = %v, want assess", result["current_station"])
	}
	if result["next_station"] != "plan" {
		t.Errorf("next_station = %v, want plan", result["next_station"])
	}
}

// TestStatus_WithoutProcession verifies status returns an error when no procession is active.
func TestStatus_WithoutProcession(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeSession(t, ctxPath)

	ctx := newTestCtx(projectDir, sessionID)
	err := runStatus(ctx)
	if err == nil {
		t.Error("expected error when no procession active, got nil")
	}
}

// TestProceed_AdvanceOneStation verifies that proceed advances to the next station correctly.
func TestProceed_AdvanceOneStation(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "audit",
		NextStation:    "assess",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	// Write a valid handoff artifact for validation
	writeValidHandoff(t, projectDir, ".sos/wip/sr/HANDOFF-audit-to-assess.md")

	ctx := newTestCtx(projectDir, sessionID)
	opts := proceedOptions{artifacts: ".sos/wip/sr/HANDOFF-audit-to-assess.md"}

	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runProceed(ctx, opts)
	})
	if runErr != nil {
		t.Fatalf("runProceed failed: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result["completed_station"] != "audit" {
		t.Errorf("completed_station = %v, want audit", result["completed_station"])
	}
	if result["new_current_station"] != "assess" {
		t.Errorf("new_current_station = %v, want assess", result["new_current_station"])
	}
	if result["next_station"] != "plan" {
		t.Errorf("next_station = %v, want plan", result["next_station"])
	}
	if result["complete"] != false {
		t.Errorf("complete = %v, want false", result["complete"])
	}

	// Verify session context was updated
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if sessCtx.Procession.CurrentStation != "assess" {
		t.Errorf("Procession.CurrentStation = %q after proceed, want assess", sessCtx.Procession.CurrentStation)
	}
	if len(sessCtx.Procession.CompletedStations) != 1 {
		t.Errorf("CompletedStations len = %d, want 1", len(sessCtx.Procession.CompletedStations))
	} else {
		cs := sessCtx.Procession.CompletedStations[0]
		if cs.Station != "audit" {
			t.Errorf("completed station name = %q, want audit", cs.Station)
		}
		if cs.Rite != "security" {
			t.Errorf("completed station rite = %q, want security", cs.Rite)
		}
		if len(cs.Artifacts) != 1 {
			t.Errorf("completed station artifacts len = %d, want 1", len(cs.Artifacts))
		}
	}
}

// TestProceed_FinalStation verifies that proceeding from the final station marks the procession complete.
func TestProceed_FinalStation(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "validate",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	ctx := newTestCtx(projectDir, sessionID)
	opts := proceedOptions{}

	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runProceed(ctx, opts)
	})
	if runErr != nil {
		t.Fatalf("runProceed final station failed: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result["complete"] != true {
		t.Errorf("complete = %v, want true", result["complete"])
	}

	// Procession block should be removed from session context on completion
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if sessCtx.Procession != nil {
		t.Fatal("Procession should be nil after final station completion")
	}
}

// TestRecede_ValidEarlierStation verifies recede repositions to an earlier station.
func TestRecede_ValidEarlierStation(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "validate",
		NextStation:    "",
		NextRite:       "",
		ArtifactDir:    ".sos/wip/security-remediation/",
		CompletedStations: []session.CompletedStation{
			{Station: "audit", Rite: "security", CompletedAt: "2026-03-10T10:00:00Z"},
			{Station: "assess", Rite: "debt-triage", CompletedAt: "2026-03-10T11:00:00Z"},
			{Station: "plan", Rite: "debt-triage", CompletedAt: "2026-03-10T12:00:00Z"},
			{Station: "remediate", Rite: "hygiene", CompletedAt: "2026-03-10T13:00:00Z"},
		},
	})

	ctx := newTestCtx(projectDir, sessionID)
	opts := recedeOptions{to: "remediate"}

	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runRecede(ctx, opts)
	})
	if runErr != nil {
		t.Fatalf("runRecede failed: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result["new_current_station"] != "remediate" {
		t.Errorf("new_current_station = %v, want remediate", result["new_current_station"])
	}
	if result["next_station"] != "validate" {
		t.Errorf("next_station = %v, want validate", result["next_station"])
	}

	// Verify session context
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if sessCtx.Procession.CurrentStation != "remediate" {
		t.Errorf("CurrentStation = %q after recede, want remediate", sessCtx.Procession.CurrentStation)
	}
	// CompletedStations should be unchanged (append-only)
	if len(sessCtx.Procession.CompletedStations) != 4 {
		t.Errorf("CompletedStations len = %d, want 4 (unchanged)", len(sessCtx.Procession.CompletedStations))
	}
}

// TestRecede_InvalidStation verifies recede rejects a station that doesn't exist in the template.
func TestRecede_InvalidStation(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "validate",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	ctx := newTestCtx(projectDir, sessionID)
	opts := recedeOptions{to: "nonexistent-station"}
	err := runRecede(ctx, opts)
	if err == nil {
		t.Error("expected error for nonexistent station, got nil")
	}
}

// TestRecede_ForwardStation verifies recede rejects a station that is the same or after current.
func TestRecede_ForwardStation(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "audit",
		NextStation:    "assess",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	ctx := newTestCtx(projectDir, sessionID)
	// "assess" is after "audit" — should be rejected
	opts := recedeOptions{to: "assess"}
	err := runRecede(ctx, opts)
	if err == nil {
		t.Error("expected error when receding to a station after current, got nil")
	}
}

// TestRecede_NoProcession verifies recede fails cleanly when no procession is active.
// This covers the post-completion case: P2 auto-nils Procession on final proceed,
// so recede on a completed procession hits the nil guard.
func TestRecede_NoProcession(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	// Write session without procession (simulates post-completion state)
	writeSessionWithProcession(t, ctxPath, nil)

	ctx := newTestCtx(projectDir, sessionID)
	opts := recedeOptions{to: "audit"}
	err := runRecede(ctx, opts)
	if err == nil {
		t.Fatal("expected error for recede without procession, got nil")
	}
}

// TestAbandon_RemovesProcession verifies abandon clears the procession from the session context.
func TestAbandon_RemovesProcession(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "assess",
		NextStation:    "plan",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	ctx := newTestCtx(projectDir, sessionID)

	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runAbandon(ctx)
	})
	if runErr != nil {
		t.Fatalf("runAbandon failed: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result["procession_id"] != "security-remediation-2026-03-10" {
		t.Errorf("procession_id = %v, want security-remediation-2026-03-10", result["procession_id"])
	}

	// Verify session context: procession should be nil
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if sessCtx.Procession != nil {
		t.Errorf("Procession should be nil after abandon, got %+v", sessCtx.Procession)
	}
	// Session should still be valid
	if sessCtx.Status != session.StatusActive {
		t.Errorf("Session status = %v after abandon, want ACTIVE", sessCtx.Status)
	}
}

// TestAbandon_NoProcession verifies abandon returns an error when nothing is active.
func TestAbandon_NoProcession(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeSession(t, ctxPath)

	ctx := newTestCtx(projectDir, sessionID)
	err := runAbandon(ctx)
	if err == nil {
		t.Error("expected error when no procession active, got nil")
	}
}

// ---- list tests ----

// projectOnlyResolver returns a ProcessionResolveFunc that only scans
// the project's processions/ directory, with no platform/user/org tiers.
// This replaces the isolateKnossosHome anti-pattern with constructor injection.
func projectOnlyResolver() ProcessionResolveFunc {
	return func(projectRoot string, embeddedFS fs.FS) ([]procmena.ResolvedProcession, error) {
		projectDir := ""
		if projectRoot != "" {
			projectDir = filepath.Join(projectRoot, "processions")
		}
		return procmena.ResolveProcessionsWithDirs(projectDir, "", "", "", embeddedFS)
	}
}

// writeTemplate2 writes a second procession template for multi-template tests.
func writeTemplate2(t *testing.T, projectDir string) {
	t.Helper()
	dir := filepath.Join(projectDir, "processions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll processions/: %v", err)
	}
	content := `name: doc-consolidation
description: "Documentation consolidation lifecycle: audit, plan, execute"
stations:
  - name: audit
    rite: docs
    goal: "Audit existing documentation"
    produces: [doc-audit]
  - name: plan
    rite: docs
    goal: "Plan consolidation"
    produces: [consolidation-plan]
  - name: execute
    rite: docs
    goal: "Execute consolidation plan"
    produces: [consolidated-docs]
artifact_dir: .sos/wip/doc-consolidation/
`
	if err := os.WriteFile(filepath.Join(dir, "doc-consolidation.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write template2: %v", err)
	}
}

// TestList_WithTemplate verifies list returns a single template with correct fields.
func TestList_WithTemplate(t *testing.T) {
	projectDir, sessionID, _ := testEnv(t)
	writeTemplate(t, projectDir)

	ctx := newTestCtx(projectDir, sessionID, projectOnlyResolver())
	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runList(ctx)
	})
	if runErr != nil {
		t.Fatalf("runList failed: %v", runErr)
	}

	var result listOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result.Total != 1 {
		t.Fatalf("total = %d, want 1", result.Total)
	}

	tmpl := result.Templates[0]
	if tmpl.Name != "security-remediation" {
		t.Errorf("name = %q, want security-remediation", tmpl.Name)
	}
	if tmpl.StationCount != 5 {
		t.Errorf("station_count = %d, want 5", tmpl.StationCount)
	}
	if tmpl.EntryRite != "security" {
		t.Errorf("entry_rite = %q, want security", tmpl.EntryRite)
	}
	if tmpl.Source != "project" {
		t.Errorf("source = %q, want project", tmpl.Source)
	}
	expectedStations := []string{"audit", "assess", "plan", "remediate", "validate"}
	if len(tmpl.Stations) != len(expectedStations) {
		t.Errorf("stations = %v, want %v", tmpl.Stations, expectedStations)
	} else {
		for i, s := range tmpl.Stations {
			if s != expectedStations[i] {
				t.Errorf("station[%d] = %q, want %q", i, s, expectedStations[i])
			}
		}
	}
	if tmpl.Description == "" {
		t.Error("description should not be empty")
	}
}

// TestList_EmptyDir verifies list returns empty results when no templates exist.
func TestList_EmptyDir(t *testing.T) {
	projectDir, sessionID, _ := testEnv(t)
	// Do NOT write any templates

	ctx := newTestCtx(projectDir, sessionID, projectOnlyResolver())
	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runList(ctx)
	})
	if runErr != nil {
		t.Fatalf("runList failed: %v", runErr)
	}

	var result listOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result.Total != 0 {
		t.Errorf("total = %d, want 0", result.Total)
	}
	if len(result.Templates) != 0 {
		t.Errorf("templates = %v, want empty", result.Templates)
	}
}

// TestList_MultipleTemplates verifies list returns multiple templates sorted by name.
func TestList_MultipleTemplates(t *testing.T) {
	projectDir, sessionID, _ := testEnv(t)
	writeTemplate(t, projectDir)
	writeTemplate2(t, projectDir)

	ctx := newTestCtx(projectDir, sessionID, projectOnlyResolver())
	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runList(ctx)
	})
	if runErr != nil {
		t.Fatalf("runList failed: %v", runErr)
	}

	var result listOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result.Total != 2 {
		t.Fatalf("total = %d, want 2", result.Total)
	}

	// Should be sorted alphabetically
	if result.Templates[0].Name != "doc-consolidation" {
		t.Errorf("templates[0].name = %q, want doc-consolidation", result.Templates[0].Name)
	}
	if result.Templates[1].Name != "security-remediation" {
		t.Errorf("templates[1].name = %q, want security-remediation", result.Templates[1].Name)
	}

	// Verify second template fields
	docConsolidation := result.Templates[0]
	if docConsolidation.StationCount != 3 {
		t.Errorf("doc-consolidation station_count = %d, want 3", docConsolidation.StationCount)
	}
	if docConsolidation.EntryRite != "docs" {
		t.Errorf("doc-consolidation entry_rite = %q, want docs", docConsolidation.EntryRite)
	}
}

// TestList_InvalidTemplateSkipped verifies that invalid templates are silently skipped.
func TestList_InvalidTemplateSkipped(t *testing.T) {
	projectDir, sessionID, _ := testEnv(t)
	writeTemplate(t, projectDir) // Valid template

	// Write an invalid YAML file
	dir := filepath.Join(projectDir, "processions")
	invalidContent := `this is not valid: [procession yaml
  broken: {{{`
	if err := os.WriteFile(filepath.Join(dir, "broken.yaml"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("write invalid template: %v", err)
	}

	ctx := newTestCtx(projectDir, sessionID, projectOnlyResolver())
	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runList(ctx)
	})
	if runErr != nil {
		t.Fatalf("runList failed: %v", runErr)
	}

	var result listOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	// Only the valid template should appear
	if result.Total != 1 {
		t.Fatalf("total = %d, want 1 (invalid template should be skipped)", result.Total)
	}
	if result.Templates[0].Name != "security-remediation" {
		t.Errorf("name = %q, want security-remediation", result.Templates[0].Name)
	}
}

// ---- validation tests (proceed with --skip-validation and handoff validation) ----

// writeValidHandoff writes a valid handoff artifact markdown file.
func writeValidHandoff(t *testing.T, dir, name string) string {
	t.Helper()
	content := `---
type: handoff
procession_id: security-remediation-2026-03-10
source_station: audit
source_rite: security
target_station: assess
target_rite: debt-triage
produced_at: "2026-03-10T10:00:00Z"
artifacts:
  - type: threat-model
    path: .sos/wip/sr/threat-model.md
---

# Handoff: audit → assess

Summary of audit findings.
`
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write handoff: %v", err)
	}
	return path
}

// writeInvalidHandoff writes a handoff artifact missing required fields.
func writeInvalidHandoff(t *testing.T, dir, name string) string {
	t.Helper()
	content := `---
type: not-a-handoff
source_station: audit
---

# Bad handoff
`
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write invalid handoff: %v", err)
	}
	return path
}

// TestProceed_ValidationRejectsInvalid verifies proceed fails with invalid handoff artifact.
func TestProceed_ValidationRejectsInvalid(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "audit",
		NextStation:    "assess",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	// Write an invalid handoff artifact
	handoffPath := writeInvalidHandoff(t, projectDir, ".sos/wip/sr/HANDOFF-bad.md")
	// Make path relative to project
	relPath, _ := filepath.Rel(projectDir, handoffPath)

	ctx := newTestCtx(projectDir, sessionID)
	opts := proceedOptions{artifacts: relPath}

	err := runProceed(ctx, opts)
	if err == nil {
		t.Error("expected validation error for invalid handoff, got nil")
	}
}

// TestProceed_ValidationAcceptsValid verifies proceed succeeds with valid handoff artifact.
func TestProceed_ValidationAcceptsValid(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "audit",
		NextStation:    "assess",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	// Write a valid handoff artifact
	handoffPath := writeValidHandoff(t, projectDir, ".sos/wip/sr/HANDOFF-audit-to-assess.md")
	relPath, _ := filepath.Rel(projectDir, handoffPath)

	ctx := newTestCtx(projectDir, sessionID)
	opts := proceedOptions{artifacts: relPath}

	var out string
	var runErr error
	out = captureOutput(func() {
		runErr = runProceed(ctx, opts)
	})
	if runErr != nil {
		t.Fatalf("runProceed with valid handoff failed: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, out)
	}

	if result["completed_station"] != "audit" {
		t.Errorf("completed_station = %v, want audit", result["completed_station"])
	}
}

// TestProceed_SkipValidation verifies --skip-validation bypasses validation.
func TestProceed_SkipValidation(t *testing.T) {
	projectDir, sessionID, ctxPath := testEnv(t)
	writeTemplate(t, projectDir)
	writeSessionWithProcession(t, ctxPath, &session.Procession{
		ID:             "security-remediation-2026-03-10",
		Type:           "security-remediation",
		CurrentStation: "audit",
		NextStation:    "assess",
		NextRite:       "debt-triage",
		ArtifactDir:    ".sos/wip/security-remediation/",
	})

	// Write an invalid handoff artifact
	handoffPath := writeInvalidHandoff(t, projectDir, ".sos/wip/sr/HANDOFF-bad.md")
	relPath, _ := filepath.Rel(projectDir, handoffPath)

	ctx := newTestCtx(projectDir, sessionID)
	opts := proceedOptions{artifacts: relPath, skipValidation: true}

	var runErr error
	captureOutput(func() {
		runErr = runProceed(ctx, opts)
	})
	if runErr != nil {
		t.Fatalf("runProceed with --skip-validation should succeed, got: %v", runErr)
	}
}
