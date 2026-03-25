package registry

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world this is a long string", 10, "hello w..."},
		{"hi", 5, "hi"},
		{"exactly", 7, "exactly"},
		{"toolong", 3, "too"},
	}

	for _, tc := range tests {
		got := truncate(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}

func TestDomainListResultOutput_TextEmpty(t *testing.T) {
	r := domainListResultOutput{
		Org:     "autom8y",
		Count:   0,
		Domains: nil,
	}
	text := r.Text()
	if text == "" {
		t.Error("Text() should not be empty")
	}
	if len(text) < 10 {
		t.Errorf("Text() too short: %q", text)
	}
}

func TestDomainListResultOutput_TextWithDomains(t *testing.T) {
	r := domainListResultOutput{
		Org:   "autom8y",
		Count: 1,
		Domains: []domainListOutput{
			{
				Repo:          "knossos",
				Domain:        "architecture",
				QualifiedName: "autom8y::knossos::architecture",
				GeneratedAt:   "2026-03-20T10:00:00Z",
				ExpiresAfter:  "7d",
				SourceHash:    "abc123",
				Confidence:    0.92,
				Stale:         false,
			},
		},
	}
	text := r.Text()
	if text == "" {
		t.Error("Text() should not be empty")
	}
}

func TestRegistryStatusOutput_Text(t *testing.T) {
	s := registryStatusOutput{
		Org:           "autom8y",
		LastSynced:    "2026-03-20T10:00:00Z",
		DomainCount:   5,
		RepoCount:     2,
		StaleCount:    1,
		CatalogPath:   "/tmp/domains.yaml",
		SchemaVersion: "1.0",
	}
	text := s.Text()
	if text == "" {
		t.Error("Text() should not be empty for status output")
	}
}
