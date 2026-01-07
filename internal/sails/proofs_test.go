package sails

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectExitCode(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantCode int
		wantOK   bool
	}{
		{
			name:     "exit code: N format",
			content:  "Running tests...\nexit code: 0\n",
			wantCode: 0,
			wantOK:   true,
		},
		{
			name:     "exit code N format (no colon)",
			content:  "Running tests...\nexit code 1\n",
			wantCode: 1,
			wantOK:   true,
		},
		{
			name:     "exited with N format",
			content:  "Build process exited with 2\n",
			wantCode: 2,
			wantOK:   true,
		},
		{
			name:     "exited N format (no with)",
			content:  "Process exited 127\n",
			wantCode: 127,
			wantOK:   true,
		},
		{
			name:     "exit status N format",
			content:  "FAIL github.com/test/pkg 0.123s\nexit status 1",
			wantCode: 1,
			wantOK:   true,
		},
		{
			name:     "case insensitive",
			content:  "EXIT CODE: 42",
			wantCode: 42,
			wantOK:   true,
		},
		{
			name:     "no exit code",
			content:  "All tests passed\nDone.",
			wantCode: 0,
			wantOK:   false,
		},
		{
			name:     "empty content",
			content:  "",
			wantCode: 0,
			wantOK:   false,
		},
		{
			name:     "exit code in middle of output",
			content:  "Starting...\nStep 1 complete\nexit code: 0\nCleaning up...",
			wantCode: 0,
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := DetectExitCode(tt.content)
			if ok != tt.wantOK {
				t.Errorf("DetectExitCode() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && code != tt.wantCode {
				t.Errorf("DetectExitCode() code = %d, want %d", code, tt.wantCode)
			}
		})
	}
}

func TestDetectTestSummary(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantPassed  int
		wantFailed  int
		wantSkipped int
		wantOK      bool
	}{
		{
			name:        "pytest style full",
			content:     "collected 10 items\n5 passed, 2 failed, 3 skipped\n",
			wantPassed:  5,
			wantFailed:  2,
			wantSkipped: 3,
			wantOK:      true,
		},
		{
			name:        "pytest style passed only",
			content:     "10 passed\n",
			wantPassed:  10,
			wantFailed:  0,
			wantSkipped: 0,
			wantOK:      true,
		},
		{
			name:        "jest style",
			content:     "Tests: 47 passed, 3 failed, 50 total\n",
			wantPassed:  47,
			wantFailed:  3,
			wantSkipped: 0,
			wantOK:      true,
		},
		{
			name:        "generic N tests passed",
			content:     "Running suite...\n15 tests passed\n",
			wantPassed:  15,
			wantFailed:  0,
			wantSkipped: 0,
			wantOK:      true,
		},
		{
			name:        "go test ok lines",
			content:     "ok  github.com/test/pkg1 0.123s\nok  github.com/test/pkg2 0.456s\n",
			wantPassed:  2,
			wantFailed:  0,
			wantSkipped: 0,
			wantOK:      true,
		},
		{
			name:        "go test mixed",
			content:     "ok  github.com/test/pkg1 0.123s\nFAIL github.com/test/pkg2 0.456s\nok  github.com/test/pkg3 0.789s\n",
			wantPassed:  2,
			wantFailed:  1,
			wantSkipped: 0,
			wantOK:      true,
		},
		{
			name:        "no test summary",
			content:     "Build completed successfully\n",
			wantPassed:  0,
			wantFailed:  0,
			wantSkipped: 0,
			wantOK:      false,
		},
		{
			name:        "empty content",
			content:     "",
			wantPassed:  0,
			wantFailed:  0,
			wantSkipped: 0,
			wantOK:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, failed, skipped, ok := DetectTestSummary(tt.content)
			if ok != tt.wantOK {
				t.Errorf("DetectTestSummary() ok = %v, want %v", ok, tt.wantOK)
			}
			if passed != tt.wantPassed {
				t.Errorf("DetectTestSummary() passed = %d, want %d", passed, tt.wantPassed)
			}
			if failed != tt.wantFailed {
				t.Errorf("DetectTestSummary() failed = %d, want %d", failed, tt.wantFailed)
			}
			if skipped != tt.wantSkipped {
				t.Errorf("DetectTestSummary() skipped = %d, want %d", skipped, tt.wantSkipped)
			}
		})
	}
}

func TestParseProofLog_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.log")

	proof, err := ParseProofLog(path, "test")
	if err != nil {
		t.Fatalf("Expected no error for missing file, got %v", err)
	}
	if proof.Status != ProofUnknown {
		t.Errorf("Expected UNKNOWN status, got %s", proof.Status)
	}
	if proof.Summary != "log file not found" {
		t.Errorf("Unexpected summary: %s", proof.Summary)
	}
}

func TestParseProofLog_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.log")

	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	proof, err := ParseProofLog(path, "test")
	if err != nil {
		t.Fatalf("Expected no error for empty file, got %v", err)
	}
	if proof.Status != ProofUnknown {
		t.Errorf("Expected UNKNOWN status, got %s", proof.Status)
	}
	if proof.Summary != "log file is empty" {
		t.Errorf("Unexpected summary: %s", proof.Summary)
	}
	if proof.EvidencePath != path {
		t.Errorf("Expected evidence path %s, got %s", path, proof.EvidencePath)
	}
}

func TestParseProofLog_TestsPassing(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-output.log")

	content := `=== RUN   TestExample
--- PASS: TestExample (0.00s)
PASS
ok  github.com/test/pkg 0.123s
exit code: 0`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofPass {
		t.Errorf("Expected PASS status, got %s", proof.Status)
	}
	if proof.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", proof.ExitCode)
	}
	if proof.EvidencePath != path {
		t.Errorf("Expected evidence path %s, got %s", path, proof.EvidencePath)
	}
}

func TestParseProofLog_TestsFailing(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-output.log")

	content := `=== RUN   TestExample
--- FAIL: TestExample (0.00s)
    example_test.go:10: assertion failed
FAIL
FAIL github.com/test/pkg 0.123s
exit status 1`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofFail {
		t.Errorf("Expected FAIL status, got %s", proof.Status)
	}
	if proof.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", proof.ExitCode)
	}
}

func TestParseProofLog_BuildSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "build-output.log")

	content := `go build -o bin/app ./cmd/app
exit code: 0`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "build")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofPass {
		t.Errorf("Expected PASS status, got %s", proof.Status)
	}
	if proof.Summary != "build succeeded" {
		t.Errorf("Unexpected summary: %s", proof.Summary)
	}
}

func TestParseProofLog_BuildFailure(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "build-output.log")

	content := `go build -o bin/app ./cmd/app
./cmd/app/main.go:10:5: undefined: foo
exit code: 2`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "build")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofFail {
		t.Errorf("Expected FAIL status, got %s", proof.Status)
	}
	if proof.ExitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", proof.ExitCode)
	}
}

func TestParseProofLog_LintClean(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "lint-output.log")

	content := `golangci-lint run ./...
exit code: 0`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "lint")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofPass {
		t.Errorf("Expected PASS status, got %s", proof.Status)
	}
	if proof.Summary != "lint clean" {
		t.Errorf("Unexpected summary: %s", proof.Summary)
	}
}

func TestParseProofLog_LintIssues(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "lint-output.log")

	content := `golangci-lint run ./...
main.go:10:1: warning: exported function Foo should have comment
main.go:20:5: error: ineffectual assignment
exit code: 1`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "lint")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofFail {
		t.Errorf("Expected FAIL status, got %s", proof.Status)
	}
	if proof.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", proof.ExitCode)
	}
}

func TestCollectProofs_NonexistentDir(t *testing.T) {
	_, err := CollectProofs("/nonexistent/session/dir")
	if err == nil {
		t.Fatal("Expected error for nonexistent directory")
	}
}

func TestCollectProofs_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	proofSet, err := CollectProofs(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// All proofs should be UNKNOWN for missing files
	if proofSet.Tests == nil || proofSet.Tests.Status != ProofUnknown {
		t.Errorf("Expected UNKNOWN tests proof, got %v", proofSet.Tests)
	}
	if proofSet.Build == nil || proofSet.Build.Status != ProofUnknown {
		t.Errorf("Expected UNKNOWN build proof, got %v", proofSet.Build)
	}
	if proofSet.Lint == nil || proofSet.Lint.Status != ProofUnknown {
		t.Errorf("Expected UNKNOWN lint proof, got %v", proofSet.Lint)
	}
}

func TestCollectProofs_AllPassing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test log
	testContent := "ok  github.com/test/pkg 0.123s\nexit code: 0"
	if err := os.WriteFile(filepath.Join(tmpDir, "test-output.log"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	// Create build log
	buildContent := "go build -o bin/app ./cmd/app\nexit code: 0"
	if err := os.WriteFile(filepath.Join(tmpDir, "build-output.log"), []byte(buildContent), 0644); err != nil {
		t.Fatalf("Failed to create build log: %v", err)
	}

	// Create lint log
	lintContent := "golangci-lint run ./...\nexit code: 0"
	if err := os.WriteFile(filepath.Join(tmpDir, "lint-output.log"), []byte(lintContent), 0644); err != nil {
		t.Fatalf("Failed to create lint log: %v", err)
	}

	proofSet, err := CollectProofs(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if proofSet.Tests.Status != ProofPass {
		t.Errorf("Expected PASS for tests, got %s", proofSet.Tests.Status)
	}
	if proofSet.Build.Status != ProofPass {
		t.Errorf("Expected PASS for build, got %s", proofSet.Build.Status)
	}
	if proofSet.Lint.Status != ProofPass {
		t.Errorf("Expected PASS for lint, got %s", proofSet.Lint.Status)
	}
}

func TestCollectProofs_WithOptionalProofs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create required logs (all passing)
	testContent := "ok  github.com/test/pkg 0.123s\nexit code: 0"
	os.WriteFile(filepath.Join(tmpDir, "test-output.log"), []byte(testContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "build-output.log"), []byte("exit code: 0"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "lint-output.log"), []byte("exit code: 0"), 0644)

	// Create optional adversarial log
	adversarialContent := "5 passed, 0 failed\nexit code: 0"
	os.WriteFile(filepath.Join(tmpDir, "adversarial-output.log"), []byte(adversarialContent), 0644)

	// Create optional integration log
	integrationContent := "3 passed\nexit code: 0"
	os.WriteFile(filepath.Join(tmpDir, "integration-output.log"), []byte(integrationContent), 0644)

	proofSet, err := CollectProofs(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if proofSet.Adversarial == nil || proofSet.Adversarial.Status != ProofPass {
		t.Errorf("Expected PASS for adversarial, got %v", proofSet.Adversarial)
	}
	if proofSet.Integration == nil || proofSet.Integration.Status != ProofPass {
		t.Errorf("Expected PASS for integration, got %v", proofSet.Integration)
	}
}

func TestProofSet_HasRequiredProofs(t *testing.T) {
	tests := []struct {
		name     string
		proofSet *ProofSet
		want     bool
	}{
		{
			name:     "nil proofs",
			proofSet: &ProofSet{},
			want:     false,
		},
		{
			name: "all unknown",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofUnknown},
				Build: &ProofItem{Status: ProofUnknown},
				Lint:  &ProofItem{Status: ProofUnknown},
			},
			want: false,
		},
		{
			name: "tests unknown",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofUnknown},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: false,
		},
		{
			name: "all pass",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "tests fail but present",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofFail},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "tests skip",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofSkip},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.proofSet.HasRequiredProofs(); got != tt.want {
				t.Errorf("HasRequiredProofs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofSet_AllPass(t *testing.T) {
	tests := []struct {
		name     string
		proofSet *ProofSet
		want     bool
	}{
		{
			name:     "nil proofs",
			proofSet: &ProofSet{},
			want:     true, // No proofs to fail
		},
		{
			name: "all pass",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "one fail",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofFail},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: false,
		},
		{
			name: "one unknown",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofUnknown},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: false,
		},
		{
			name: "pass and skip",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofSkip},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "with optional proofs passing",
			proofSet: &ProofSet{
				Tests:       &ProofItem{Status: ProofPass},
				Build:       &ProofItem{Status: ProofPass},
				Lint:        &ProofItem{Status: ProofPass},
				Adversarial: &ProofItem{Status: ProofPass},
				Integration: &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "optional proof fails",
			proofSet: &ProofSet{
				Tests:       &ProofItem{Status: ProofPass},
				Build:       &ProofItem{Status: ProofPass},
				Lint:        &ProofItem{Status: ProofPass},
				Adversarial: &ProofItem{Status: ProofFail},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.proofSet.AllPass(); got != tt.want {
				t.Errorf("AllPass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofSet_AnyFail(t *testing.T) {
	tests := []struct {
		name     string
		proofSet *ProofSet
		want     bool
	}{
		{
			name:     "nil proofs",
			proofSet: &ProofSet{},
			want:     false,
		},
		{
			name: "all pass",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: false,
		},
		{
			name: "tests fail",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofFail},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "build fail",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofFail},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: true,
		},
		{
			name: "lint fail",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofFail},
			},
			want: true,
		},
		{
			name: "adversarial fail",
			proofSet: &ProofSet{
				Tests:       &ProofItem{Status: ProofPass},
				Build:       &ProofItem{Status: ProofPass},
				Lint:        &ProofItem{Status: ProofPass},
				Adversarial: &ProofItem{Status: ProofFail},
			},
			want: true,
		},
		{
			name: "unknown is not fail",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofUnknown},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.proofSet.AnyFail(); got != tt.want {
				t.Errorf("AnyFail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofSet_Summary(t *testing.T) {
	tests := []struct {
		name     string
		proofSet *ProofSet
		want     string
	}{
		{
			name:     "nil proofs",
			proofSet: &ProofSet{},
			want:     "no proofs collected",
		},
		{
			name: "all pass",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofPass},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofPass},
			},
			want: "tests: PASS, build: PASS, lint: PASS",
		},
		{
			name: "with optional",
			proofSet: &ProofSet{
				Tests:       &ProofItem{Status: ProofPass},
				Build:       &ProofItem{Status: ProofPass},
				Lint:        &ProofItem{Status: ProofPass},
				Adversarial: &ProofItem{Status: ProofPass},
			},
			want: "tests: PASS, build: PASS, lint: PASS, adversarial: PASS",
		},
		{
			name: "mixed status",
			proofSet: &ProofSet{
				Tests: &ProofItem{Status: ProofFail},
				Build: &ProofItem{Status: ProofPass},
				Lint:  &ProofItem{Status: ProofUnknown},
			},
			want: "tests: FAIL, build: PASS, lint: UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.proofSet.Summary(); got != tt.want {
				t.Errorf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProofCollector_CollectForSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session directory structure
	sessionID := "session-20260105-143022-abc12345"
	sessionDir := filepath.Join(tmpDir, ".claude", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Create test log
	testContent := "ok  github.com/test/pkg 0.123s\nexit code: 0"
	if err := os.WriteFile(filepath.Join(sessionDir, "test-output.log"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	// Create build log
	if err := os.WriteFile(filepath.Join(sessionDir, "build-output.log"), []byte("exit code: 0"), 0644); err != nil {
		t.Fatalf("Failed to create build log: %v", err)
	}

	// Create lint log
	if err := os.WriteFile(filepath.Join(sessionDir, "lint-output.log"), []byte("exit code: 0"), 0644); err != nil {
		t.Fatalf("Failed to create lint log: %v", err)
	}

	collector := NewProofCollector(tmpDir)
	proofSet, err := collector.CollectForSession(sessionID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if proofSet.Tests.Status != ProofPass {
		t.Errorf("Expected PASS for tests, got %s", proofSet.Tests.Status)
	}
	if proofSet.Build.Status != ProofPass {
		t.Errorf("Expected PASS for build, got %s", proofSet.Build.Status)
	}
	if proofSet.Lint.Status != ProofPass {
		t.Errorf("Expected PASS for lint, got %s", proofSet.Lint.Status)
	}
}

func TestCollectProofs_EmptySessionDir(t *testing.T) {
	_, err := CollectProofs("")
	if err == nil {
		t.Fatal("Expected error for empty session directory")
	}
}

func TestCollectProofs_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir")
	os.WriteFile(filePath, []byte("content"), 0644)

	_, err := CollectProofs(filePath)
	if err == nil {
		t.Fatal("Expected error when path is not a directory")
	}
}

func TestParseProofLog_PytestOutput(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-output.log")

	content := `============================= test session starts ==============================
platform linux -- Python 3.11.0, pytest-7.4.0
collected 10 items

test_example.py::test_one PASSED
test_example.py::test_two PASSED
test_example.py::test_three FAILED

=========================== short test summary info ============================
FAILED test_example.py::test_three - AssertionError
========================= 2 passed, 1 failed, 0 skipped =========================
exit code: 1`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	proof, err := ParseProofLog(path, "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if proof.Status != ProofFail {
		t.Errorf("Expected FAIL status, got %s", proof.Status)
	}
	if proof.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", proof.ExitCode)
	}
}

func TestAnalyzeTestOutput_FallbackToPassMarker(t *testing.T) {
	// Test case where there's no summary but PASS marker exists
	content := "Running test...\nPASS\nDone."
	status, _ := analyzeTestOutput(content, 0, false)
	if status != ProofPass {
		t.Errorf("Expected PASS from marker, got %s", status)
	}
}

func TestAnalyzeTestOutput_FallbackToFailMarker(t *testing.T) {
	// Test case where there's no summary but FAIL marker exists
	content := "Running test...\nFAIL: assertion error\nDone."
	status, _ := analyzeTestOutput(content, 0, false)
	if status != ProofFail {
		t.Errorf("Expected FAIL from marker, got %s", status)
	}
}

func TestAnalyzeBuildOutput_ErrorPatterns(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    ProofStatus
	}{
		{
			name:    "error colon pattern",
			content: "main.go:10:5: error: undefined variable\n",
			want:    ProofFail,
		},
		{
			name:    "fatal error",
			content: "fatal error: out of memory\n",
			want:    ProofFail,
		},
		{
			name:    "build failed",
			content: "Build failed: missing dependencies\n",
			want:    ProofFail,
		},
		{
			name:    "compilation failed",
			content: "Compilation failed with 3 errors\n",
			want:    ProofFail,
		},
		{
			name:    "build successful",
			content: "Build successful\nOutput: bin/app\n",
			want:    ProofPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := analyzeBuildOutput(tt.content, 0, false)
			if status != tt.want {
				t.Errorf("analyzeBuildOutput() = %s, want %s", status, tt.want)
			}
		})
	}
}

func TestAnalyzeLintOutput_CleanPatterns(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    ProofStatus
	}{
		{
			name:    "no issues found",
			content: "Linting complete. No issues found.\n",
			want:    ProofPass,
		},
		{
			name:    "no problems",
			content: "ESLint: No problems detected.\n",
			want:    ProofPass,
		},
		{
			name:    "clean output",
			content: "golangci-lint: clean\n",
			want:    ProofPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := analyzeLintOutput(tt.content, 0, false)
			if status != tt.want {
				t.Errorf("analyzeLintOutput() = %s, want %s", status, tt.want)
			}
		})
	}
}
