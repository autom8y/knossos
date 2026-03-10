package procession

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"gopkg.in/yaml.v3"
)

// validTemplateYAML is a well-formed template with 5 stations, used as a baseline.
const validTemplateYAML = `
name: security-remediation
description: "Security findings lifecycle: audit, assess, plan, remediate, validate"

stations:
  - name: audit
    rite: security
    goal: "Map attack surface, classify findings by exploitability, produce threat model and pentest report"
    produces: [threat-model, pentest-report]

  - name: assess
    rite: debt-triage
    goal: "Catalog findings, score risk, produce prioritized remediation backlog"
    produces: [debt-inventory, priority-matrix]

  - name: plan
    rite: debt-triage
    goal: "Group findings into sprint-sized tasks with acceptance criteria"
    produces: [sprint-plan]

  - name: remediate
    rite: hygiene
    alt_rite: 10x-dev
    goal: "Execute remediation plan, produce PRs with fixes"
    produces: [remediation-ledger]

  - name: validate
    rite: security
    goal: "Review remediation PRs for security correctness"
    produces: [validation-report]
    loop_to: remediate

artifact_dir: .sos/wip/security-remediation/
`

// mustParseTemplate unmarshals the YAML into a Template without running Validate().
// Used in mutation tests to set up a valid baseline before applying a specific invalid change.
func mustParseTemplate(t *testing.T, data string) Template {
	t.Helper()
	var tmpl Template
	if err := yaml.Unmarshal([]byte(data), &tmpl); err != nil {
		t.Fatalf("mustParseTemplate: yaml.Unmarshal: %v", err)
	}
	return tmpl
}

func TestLoadTemplate_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "security-remediation.yaml")
	if err := os.WriteFile(path, []byte(validTemplateYAML), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tmpl, err := LoadTemplate(path)
	if err != nil {
		t.Fatalf("LoadTemplate() error = %v", err)
	}

	if tmpl.Name != "security-remediation" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "security-remediation")
	}
	if tmpl.Description == "" {
		t.Error("Description should not be empty")
	}
	if len(tmpl.Stations) != 5 {
		t.Errorf("Stations count = %d, want 5", len(tmpl.Stations))
	}
	if tmpl.ArtifactDir != ".sos/wip/security-remediation/" {
		t.Errorf("ArtifactDir = %q, want %q", tmpl.ArtifactDir, ".sos/wip/security-remediation/")
	}

	// Verify last station has loop_to
	last := tmpl.Stations[4]
	if last.LoopTo != "remediate" {
		t.Errorf("Stations[4].LoopTo = %q, want %q", last.LoopTo, "remediate")
	}
}

func TestValidate_ValidationCases(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*Template)
		wantErrFrag string // substring that must appear in the error message; empty means no error
	}{
		{
			name: "missing name",
			mutate: func(tmpl *Template) {
				tmpl.Name = ""
			},
			wantErrFrag: "name",
		},
		{
			name: "invalid name pattern uppercase and space",
			mutate: func(tmpl *Template) {
				tmpl.Name = "Security Remediation"
			},
			wantErrFrag: "pattern",
		},
		{
			name: "single station",
			mutate: func(tmpl *Template) {
				tmpl.Stations = tmpl.Stations[:1]
			},
			wantErrFrag: "at least 2 stations required",
		},
		{
			name: "duplicate station names",
			mutate: func(tmpl *Template) {
				// Rename second station to match the first
				tmpl.Stations[1].Name = "audit"
			},
			wantErrFrag: "duplicate station name",
		},
		{
			name: "invalid loop_to references nonexistent station",
			mutate: func(tmpl *Template) {
				tmpl.Stations[4].LoopTo = "nonexistent"
			},
			wantErrFrag: "loop_to",
		},
		{
			name:        "valid loop_to references existing station",
			mutate:      func(*Template) {},
			wantErrFrag: "", // baseline already has valid loop_to: remediate; no error expected
		},
		{
			name: "empty produces slice",
			mutate: func(tmpl *Template) {
				tmpl.Stations[0].Produces = []string{}
			},
			wantErrFrag: "produces: must have at least 1 element",
		},
		{
			name: "invalid artifact_dir",
			mutate: func(tmpl *Template) {
				tmpl.ArtifactDir = "/tmp/bad/"
			},
			wantErrFrag: "artifact_dir",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse a fresh copy of the valid template for each test case
			tmpl := mustParseTemplate(t, validTemplateYAML)
			tc.mutate(&tmpl)
			err := tmpl.Validate()

			if tc.wantErrFrag == "" {
				if err != nil {
					t.Errorf("Validate() returned unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() returned nil, want error containing %q", tc.wantErrFrag)
			}
			if !strings.Contains(err.Error(), tc.wantErrFrag) {
				t.Errorf("Validate() error = %q, want it to contain %q", err.Error(), tc.wantErrFrag)
			}
		})
	}
}

func TestLoadEmbeddedTemplate_Valid(t *testing.T) {
	memFS := fstest.MapFS{
		"processions/security-remediation.yaml": &fstest.MapFile{
			Data: []byte(validTemplateYAML),
		},
	}

	tmpl, err := LoadEmbeddedTemplate("security-remediation", memFS)
	if err != nil {
		t.Fatalf("LoadEmbeddedTemplate() error = %v", err)
	}

	if tmpl.Name != "security-remediation" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "security-remediation")
	}
	if len(tmpl.Stations) != 5 {
		t.Errorf("Stations count = %d, want 5", len(tmpl.Stations))
	}
}

func TestLoadEmbeddedTemplate_Missing(t *testing.T) {
	memFS := fstest.MapFS{}

	_, err := LoadEmbeddedTemplate("nonexistent", memFS)
	if err == nil {
		t.Fatal("LoadEmbeddedTemplate() should return error for missing template")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want it to mention the template name", err.Error())
	}
}

func TestLoadTemplate_MissingFile(t *testing.T) {
	_, err := LoadTemplate("/nonexistent/path/template.yaml")
	if err == nil {
		t.Fatal("LoadTemplate() should return error for missing file")
	}
}

// --- Helper method tests ---

func TestTemplate_GetStation(t *testing.T) {
	tmpl := mustParseTemplate(t, validTemplateYAML)

	s := tmpl.GetStation("audit")
	if s == nil {
		t.Fatal("GetStation(audit) returned nil")
	}
	if s.Name != "audit" {
		t.Errorf("GetStation name = %q, want %q", s.Name, "audit")
	}
	if s.Rite != "security" {
		t.Errorf("GetStation rite = %q, want %q", s.Rite, "security")
	}

	// Not found returns nil
	if got := tmpl.GetStation("nonexistent"); got != nil {
		t.Errorf("GetStation(nonexistent) should return nil, got %+v", got)
	}
}

func TestTemplate_StationNames(t *testing.T) {
	tmpl := mustParseTemplate(t, validTemplateYAML)

	names := tmpl.StationNames()
	want := []string{"audit", "assess", "plan", "remediate", "validate"}
	if len(names) != len(want) {
		t.Fatalf("StationNames() length = %d, want %d", len(names), len(want))
	}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("StationNames()[%d] = %q, want %q", i, n, want[i])
		}
	}
}

func TestTemplate_NextStation(t *testing.T) {
	tmpl := mustParseTemplate(t, validTemplateYAML)

	tests := []struct {
		current string
		want    string
	}{
		{"audit", "assess"},
		{"assess", "plan"},
		{"plan", "remediate"},
		{"remediate", "validate"},
		{"validate", ""},     // last station
		{"nonexistent", ""}, // not found
	}

	for _, tc := range tests {
		got := tmpl.NextStation(tc.current)
		if got != tc.want {
			t.Errorf("NextStation(%q) = %q, want %q", tc.current, got, tc.want)
		}
	}
}
