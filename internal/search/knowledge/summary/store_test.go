package summary

import (
	"context"
	"fmt"
	"testing"
)

// mockLLMClient implements LLMClient for testing.
type mockLLMClient struct {
	response string
	err      error
	calls    int
}

func (m *mockLLMClient) Complete(_ context.Context, _, _ string, _ int) (string, error) {
	m.calls++
	return m.response, m.err
}

func TestStore_GetSummary(t *testing.T) {
	tests := []struct {
		name          string
		qualifiedName string
		preload       map[string]*DomainSummary
		wantText      string
		wantOK        bool
	}{
		{
			name:          "found",
			qualifiedName: "org::repo::arch",
			preload: map[string]*DomainSummary{
				"org::repo::arch": {
					QualifiedName: "org::repo::arch",
					DomainSummary: "Architecture overview.",
					SourceHash:    "abc123",
				},
			},
			wantText: "Architecture overview.",
			wantOK:   true,
		},
		{
			name:          "not found",
			qualifiedName: "org::repo::missing",
			preload:       map[string]*DomainSummary{},
			wantText:      "",
			wantOK:        false,
		},
		{
			name:          "nil entry",
			qualifiedName: "org::repo::nil",
			preload: map[string]*DomainSummary{
				"org::repo::nil": nil,
			},
			wantText: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStoreFromMap(tt.preload)
			got, ok := s.GetSummary(tt.qualifiedName)
			if ok != tt.wantOK {
				t.Errorf("GetSummary() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.wantText {
				t.Errorf("GetSummary() text = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestStore_NeedsRegeneration(t *testing.T) {
	tests := []struct {
		name          string
		qualifiedName string
		sourceHash    string
		preload       map[string]*DomainSummary
		want          bool
	}{
		{
			name:          "missing domain needs regen",
			qualifiedName: "org::repo::new",
			sourceHash:    "hash1",
			preload:       map[string]*DomainSummary{},
			want:          true,
		},
		{
			name:          "hash match skips regen",
			qualifiedName: "org::repo::arch",
			sourceHash:    "abc123",
			preload: map[string]*DomainSummary{
				"org::repo::arch": {
					QualifiedName: "org::repo::arch",
					DomainSummary: "Existing.",
					SourceHash:    "abc123",
				},
			},
			want: false,
		},
		{
			name:          "hash mismatch needs regen",
			qualifiedName: "org::repo::arch",
			sourceHash:    "def456",
			preload: map[string]*DomainSummary{
				"org::repo::arch": {
					QualifiedName: "org::repo::arch",
					DomainSummary: "Old.",
					SourceHash:    "abc123",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStoreFromMap(tt.preload)
			got := s.NeedsRegeneration(tt.qualifiedName, tt.sourceHash)
			if got != tt.want {
				t.Errorf("NeedsRegeneration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Generate(t *testing.T) {
	tests := []struct {
		name        string
		llmResponse string
		llmErr      error
		sections    map[string]string
		wantErr     bool
		wantDomain  string
		wantSecs    int
	}{
		{
			name: "successful generation with sections",
			llmResponse: `This domain covers the codebase architecture. It describes package structure and data flow patterns.

SECTION: package-structure | Describes the Go package layout and dependency hierarchy.
SECTION: data-flow | Explains how data moves through the pipeline stages.`,
			sections: map[string]string{
				"package-structure": "Package structure content...",
				"data-flow":         "Data flow content...",
			},
			wantErr:    false,
			wantDomain: "This domain covers the codebase architecture. It describes package structure and data flow patterns.",
			wantSecs:   2,
		},
		{
			name: "successful generation without sections",
			llmResponse: `This is a concise architecture overview. It covers key patterns.`,
			sections:    nil,
			wantErr:     false,
			wantDomain:  "This is a concise architecture overview. It covers key patterns.",
			wantSecs:    0,
		},
		{
			name:    "LLM error",
			llmErr:  fmt.Errorf("API timeout"),
			wantErr: true,
		},
		{
			name:    "nil LLM client",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore()

			var client LLMClient
			if tt.name != "nil LLM client" {
				client = &mockLLMClient{response: tt.llmResponse, err: tt.llmErr}
			}

			result, err := s.Generate(
				context.Background(),
				"org::repo::arch",
				"Full content body here...",
				"hash1",
				tt.sections,
				client,
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if result.DomainSummary != tt.wantDomain {
				t.Errorf("DomainSummary = %q, want %q", result.DomainSummary, tt.wantDomain)
			}

			if len(result.SectionSummaries) != tt.wantSecs {
				t.Errorf("SectionSummaries count = %d, want %d", len(result.SectionSummaries), tt.wantSecs)
			}

			if result.SourceHash != "hash1" {
				t.Errorf("SourceHash = %q, want %q", result.SourceHash, "hash1")
			}

			// Verify stored in the store.
			got, ok := s.GetSummary("org::repo::arch")
			if !ok {
				t.Error("summary not stored after Generate()")
			}
			if got != tt.wantDomain {
				t.Errorf("stored summary = %q, want %q", got, tt.wantDomain)
			}
		})
	}
}

func TestStore_Set(t *testing.T) {
	s := NewStore()

	s.Set(&DomainSummary{
		QualifiedName: "org::repo::test",
		DomainSummary: "Test summary.",
		SourceHash:    "hash1",
	})

	if s.Count() != 1 {
		t.Errorf("Count() = %d, want 1", s.Count())
	}

	got, ok := s.GetSummary("org::repo::test")
	if !ok || got != "Test summary." {
		t.Errorf("Set/GetSummary mismatch: got %q, ok=%v", got, ok)
	}

	// Set nil should be a no-op.
	s.Set(nil)
	if s.Count() != 1 {
		t.Errorf("Count() after nil Set = %d, want 1", s.Count())
	}
}

func TestStore_All(t *testing.T) {
	s := NewStore()
	s.Set(&DomainSummary{QualifiedName: "a", DomainSummary: "A", SourceHash: "1"})
	s.Set(&DomainSummary{QualifiedName: "b", DomainSummary: "B", SourceHash: "2"})

	all := s.All()
	if len(all) != 2 {
		t.Errorf("All() len = %d, want 2", len(all))
	}
}
