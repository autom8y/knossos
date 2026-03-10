package procession

import (
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/procession"
)

func TestBuildWorkflowData(t *testing.T) {
	t.Parallel()
	tmpl := &procession.Template{
		Name:        "security-remediation",
		Description: "Security findings lifecycle",
		ArtifactDir: ".sos/wip/security-remediation/",
		Stations: []procession.Station{
			{Name: "audit", Rite: "security", Goal: "Map attack surface", Produces: []string{"threat-model"}},
			{Name: "assess", Rite: "debt-triage", Goal: "Score risk", Produces: []string{"inventory"}},
			{Name: "validate", Rite: "security", Goal: "Review fixes", Produces: []string{"report"}, LoopTo: "assess"},
		},
	}

	data := BuildWorkflowData(tmpl)

	if data.Name != "security-remediation" {
		t.Errorf("expected Name security-remediation, got %s", data.Name)
	}
	if data.SkillName != "security-remediation-ref" {
		t.Errorf("expected SkillName security-remediation-ref, got %s", data.SkillName)
	}
	if data.FirstStation != "audit" {
		t.Errorf("expected FirstStation audit, got %s", data.FirstStation)
	}
	if data.FirstRite != "security" {
		t.Errorf("expected FirstRite security, got %s", data.FirstRite)
	}
	if data.StationCount != 3 {
		t.Errorf("expected StationCount 3, got %d", data.StationCount)
	}
	if data.ArtifactDir != ".sos/wip/security-remediation/" {
		t.Errorf("expected ArtifactDir .sos/wip/security-remediation/, got %s", data.ArtifactDir)
	}
}

func TestBuildWorkflowData_StationTable(t *testing.T) {
	t.Parallel()
	tmpl := &procession.Template{
		Name:        "test",
		Description: "Test",
		ArtifactDir: ".sos/wip/test/",
		Stations: []procession.Station{
			{Name: "alpha", Rite: "security", AltRite: "10x-dev", Goal: "First goal", Produces: []string{"a"}, LoopTo: ""},
			{Name: "beta", Rite: "hygiene", Goal: "Second goal", Produces: []string{"b", "c"}, LoopTo: "alpha"},
		},
	}

	data := BuildWorkflowData(tmpl)

	// Check table header
	if !strings.Contains(data.StationTable, "| # | Station | Rite |") {
		t.Error("station table missing header")
	}
	// Check station rows
	if !strings.Contains(data.StationTable, "| 1 | alpha | security | 10x-dev |") {
		t.Error("station table missing alpha row with alt_rite")
	}
	if !strings.Contains(data.StationTable, "| 2 | beta | hygiene | - |") {
		t.Error("station table missing beta row")
	}
	// Check loop_to rendering
	if !strings.Contains(data.StationTable, "| alpha |") {
		t.Error("station table missing loop_to value for beta")
	}
}

func TestBuildWorkflowData_LongGoalTruncated(t *testing.T) {
	t.Parallel()
	longGoal := strings.Repeat("x", 100)
	tmpl := &procession.Template{
		Name:        "test",
		Description: "Test",
		ArtifactDir: ".sos/wip/test/",
		Stations: []procession.Station{
			{Name: "a", Rite: "security", Goal: longGoal, Produces: []string{"x"}},
			{Name: "b", Rite: "hygiene", Goal: "Short", Produces: []string{"y"}},
		},
	}

	data := BuildWorkflowData(tmpl)

	if strings.Contains(data.StationTable, longGoal) {
		t.Error("expected long goal to be truncated in table")
	}
	if !strings.Contains(data.StationTable, "...") {
		t.Error("expected truncation indicator ...")
	}
}
