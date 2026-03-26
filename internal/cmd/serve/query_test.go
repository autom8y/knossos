package serve

import (
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/reason"
	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/triage"
	"github.com/autom8y/knossos/internal/trust"
)

func TestNewQueryCmd(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""

	serveCmd := NewServeCmd(&output, &verbose, &projectDir)

	// Find the query subcommand.
	var queryCmd *cobra.Command
	for _, sub := range serveCmd.Commands() {
		if sub.Name() == "query" {
			queryCmd = sub
			break
		}
	}
	if queryCmd == nil {
		t.Fatal("expected 'query' subcommand to be registered on serve")
	}

	if queryCmd.Use != "query [question]" {
		t.Errorf("expected Use 'query [question]', got %q", queryCmd.Use)
	}

	// Verify flags exist.
	orgFlag := queryCmd.Flags().Lookup("org")
	if orgFlag == nil {
		t.Fatal("expected --org flag")
	}
	if orgFlag.DefValue != "" {
		t.Errorf("expected org default empty, got %q", orgFlag.DefValue)
	}

	diagnosticFlag := queryCmd.Flags().Lookup("diagnostic")
	if diagnosticFlag == nil {
		t.Fatal("expected --diagnostic flag")
	}
	if diagnosticFlag.DefValue != "false" {
		t.Errorf("expected diagnostic default false, got %q", diagnosticFlag.DefValue)
	}

	noTriageFlag := queryCmd.Flags().Lookup("no-triage")
	if noTriageFlag == nil {
		t.Fatal("expected --no-triage flag")
	}
	if noTriageFlag.DefValue != "false" {
		t.Errorf("expected no-triage default false, got %q", noTriageFlag.DefValue)
	}
}

func TestNewQueryCmd_RequiresArgs(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""

	serveCmd := NewServeCmd(&output, &verbose, &projectDir)

	// Find the query subcommand.
	var queryCmd *cobra.Command
	for _, sub := range serveCmd.Commands() {
		if sub.Name() == "query" {
			queryCmd = sub
			break
		}
	}
	if queryCmd == nil {
		t.Fatal("expected 'query' subcommand to be registered on serve")
	}

	// Validate that it requires exactly one argument.
	err := queryCmd.Args(queryCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = queryCmd.Args(queryCmd, []string{"question"})
	if err != nil {
		t.Errorf("expected no error with one arg, got: %v", err)
	}

	err = queryCmd.Args(queryCmd, []string{"question1", "question2"})
	if err == nil {
		t.Error("expected error with two args")
	}
}

func TestConvertTriageResult(t *testing.T) {
	tests := []struct {
		name   string
		input  *triage.TriageResult
		want   *reason.TriageResultInput
		nilOut bool
	}{
		{
			name:   "nil input returns nil",
			input:  nil,
			nilOut: true,
		},
		{
			name: "empty candidates",
			input: &triage.TriageResult{
				RefinedQuery:   "refined query",
				Candidates:     nil,
				ModelCallCount: 1,
			},
			// make([]T, 0) produces a non-nil empty slice.
			want: &reason.TriageResultInput{
				RefinedQuery:   "refined query",
				Candidates:     make([]reason.TriageCandidateInput, 0),
				ModelCallCount: 1,
			},
		},
		{
			name: "full conversion",
			input: &triage.TriageResult{
				RefinedQuery: "How is knossos structured?",
				Candidates: []triage.TriageCandidate{
					{
						QualifiedName:       "autom8y::knossos::architecture",
						RelevanceScore:      0.95,
						EmbeddingSimilarity: 0.0,
						Freshness:           0.85,
						Rationale:           "core architecture doc",
						DomainType:          "architecture",
						RelatedDomains:      []string{"conventions"},
					},
					{
						QualifiedName:  "autom8y::knossos::conventions",
						RelevanceScore: 0.72,
						Freshness:      0.90,
						DomainType:     "conventions",
					},
				},
				ModelCallCount: 2,
			},
			want: &reason.TriageResultInput{
				RefinedQuery: "How is knossos structured?",
				Candidates: []reason.TriageCandidateInput{
					{
						QualifiedName:       "autom8y::knossos::architecture",
						RelevanceScore:      0.95,
						EmbeddingSimilarity: 0.0,
						Freshness:           0.85,
						Rationale:           "core architecture doc",
						DomainType:          "architecture",
						RelatedDomains:      []string{"conventions"},
					},
					{
						QualifiedName:  "autom8y::knossos::conventions",
						RelevanceScore: 0.72,
						Freshness:      0.90,
						DomainType:     "conventions",
					},
				},
				ModelCallCount: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTriageResult(tt.input)
			if tt.nilOut {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if got.RefinedQuery != tt.want.RefinedQuery {
				t.Errorf("RefinedQuery = %q, want %q", got.RefinedQuery, tt.want.RefinedQuery)
			}
			if got.ModelCallCount != tt.want.ModelCallCount {
				t.Errorf("ModelCallCount = %d, want %d", got.ModelCallCount, tt.want.ModelCallCount)
			}
			if len(got.Candidates) != len(tt.want.Candidates) {
				t.Errorf("Candidates len = %d, want %d", len(got.Candidates), len(tt.want.Candidates))
				return
			}
			for i, c := range got.Candidates {
				wantC := tt.want.Candidates[i]
				if c.QualifiedName != wantC.QualifiedName {
					t.Errorf("Candidate[%d].QualifiedName = %q, want %q", i, c.QualifiedName, wantC.QualifiedName)
				}
				if c.RelevanceScore != wantC.RelevanceScore {
					t.Errorf("Candidate[%d].RelevanceScore = %f, want %f", i, c.RelevanceScore, wantC.RelevanceScore)
				}
				if c.Freshness != wantC.Freshness {
					t.Errorf("Candidate[%d].Freshness = %f, want %f", i, c.Freshness, wantC.Freshness)
				}
				if c.DomainType != wantC.DomainType {
					t.Errorf("Candidate[%d].DomainType = %q, want %q", i, c.DomainType, wantC.DomainType)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "-"},
		{"microseconds", 500 * time.Microsecond, "500us"},
		{"milliseconds", 42 * time.Millisecond, "42ms"},
		{"seconds", 3200 * time.Millisecond, "3.2s"},
		{"large seconds", 15 * time.Second, "15.0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestBoolReady(t *testing.T) {
	if boolReady(true) != "ready" {
		t.Errorf("boolReady(true) = %q, want 'ready'", boolReady(true))
	}
	if boolReady(false) != "not available" {
		t.Errorf("boolReady(false) = %q, want 'not available'", boolReady(false))
	}
}

func TestPrintStandardOutput(t *testing.T) {
	// Smoke test: ensure printStandardOutput does not panic for various response shapes.
	tests := []struct {
		name string
		resp *response.ReasoningResponse
	}{
		{
			name: "high tier with citations",
			resp: &response.ReasoningResponse{
				Answer: "Knossos is structured as a Go CLI project.",
				Tier:   trust.TierHigh,
				Confidence: trust.ConfidenceScore{
					Overall:   0.92,
					Freshness: 0.85,
					Retrieval: 0.95,
					Coverage:  0.90,
					Tier:      trust.TierHigh,
				},
				Citations: []response.Citation{
					{QualifiedName: "autom8y::knossos::architecture", Section: "Package Structure"},
				},
			},
		},
		{
			name: "low tier with gap",
			resp: &response.ReasoningResponse{
				Answer: "insufficient knowledge",
				Tier:   trust.TierLow,
				Confidence: trust.ConfidenceScore{
					Overall: 0.2,
					Tier:    trust.TierLow,
				},
				Gap: &trust.GapAdmission{
					Reason:      "no matching domains",
					Suggestions: []string{"run ari registry sync"},
				},
			},
		},
		{
			name: "medium tier no citations",
			resp: &response.ReasoningResponse{
				Answer: "Based on available information...",
				Tier:   trust.TierMedium,
				Confidence: trust.ConfidenceScore{
					Overall: 0.6,
					Tier:    trust.TierMedium,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic. Output goes to os.Stdout.
			printStandardOutput(tt.resp)
		})
	}
}

func TestRunQuery_MissingAPIKey(t *testing.T) {
	// Ensure ANTHROPIC_API_KEY is not set.
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("KNOSSOS_ORG", "")

	output := "text"
	verbose := false
	projectDir := ""

	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &output,
			Verbose:    &verbose,
			ProjectDir: &projectDir,
		},
	}

	opts := queryOptions{}
	err := runQuery(ctx, opts, "test question")
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is not set")
	}
}
