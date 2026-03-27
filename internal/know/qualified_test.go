package know

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOrg    string
		wantRepo   string
		wantScope  string
		wantDomain string
		wantErr    bool
	}{
		// Root-scope names
		{
			name:       "simple valid",
			input:      "autom8y::knossos::architecture",
			wantOrg:    "autom8y",
			wantRepo:   "knossos",
			wantDomain: "architecture",
		},
		{
			name:       "domain with slash",
			input:      "autom8y::knossos::feat/materialization",
			wantOrg:    "autom8y",
			wantRepo:   "knossos",
			wantDomain: "feat/materialization",
		},
		{
			name:       "domain with deep path",
			input:      "autom8y::payments-service::release/platform-profile",
			wantOrg:    "autom8y",
			wantRepo:   "payments-service",
			wantDomain: "release/platform-profile",
		},
		// Scoped names
		{
			name:       "single-level scope",
			input:      "autom8y::autom8y/services/ads::architecture",
			wantOrg:    "autom8y",
			wantRepo:   "autom8y",
			wantScope:  "services/ads",
			wantDomain: "architecture",
		},
		{
			name:       "multi-level scope",
			input:      "autom8y::autom8y/sdks/python/autom8y-meta::conventions",
			wantOrg:    "autom8y",
			wantRepo:   "autom8y",
			wantScope:  "sdks/python/autom8y-meta",
			wantDomain: "conventions",
		},
		{
			name:       "scope plus domain with slash",
			input:      "autom8y::autom8y/services/ads::feat/materialization",
			wantOrg:    "autom8y",
			wantRepo:   "autom8y",
			wantScope:  "services/ads",
			wantDomain: "feat/materialization",
		},
		// Error cases
		{name: "empty string", input: "", wantErr: true},
		{name: "missing segments", input: "autom8y", wantErr: true},
		{name: "missing domain", input: "autom8y::knossos", wantErr: true},
		{name: "empty org", input: "::knossos::architecture", wantErr: true},
		{name: "empty repo segment", input: "autom8y::::architecture", wantErr: true},
		{name: "empty domain", input: "autom8y::knossos::", wantErr: true},
		{name: ":: in domain", input: "autom8y::knossos::arch::extra", wantErr: true},
		{name: "whitespace org", input: "   ::knossos::architecture", wantErr: true},
		{name: "whitespace domain", input: "autom8y::knossos::   ", wantErr: true},
		{name: "trailing slash empty scope", input: "autom8y::repo/::domain", wantErr: true},
		{name: "leading slash empty repo", input: "autom8y::/scope::domain", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got %+v", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tc.input, err)
			}
			if got.Org != tc.wantOrg {
				t.Errorf("Org = %q, want %q", got.Org, tc.wantOrg)
			}
			if got.Repo != tc.wantRepo {
				t.Errorf("Repo = %q, want %q", got.Repo, tc.wantRepo)
			}
			if got.Scope != tc.wantScope {
				t.Errorf("Scope = %q, want %q", got.Scope, tc.wantScope)
			}
			if got.Domain != tc.wantDomain {
				t.Errorf("Domain = %q, want %q", got.Domain, tc.wantDomain)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		q    QualifiedDomainName
		want string
	}{
		{
			name: "root scope",
			q:    QualifiedDomainName{Org: "autom8y", Repo: "knossos", Domain: "architecture"},
			want: "autom8y::knossos::architecture",
		},
		{
			name: "domain with slash",
			q:    QualifiedDomainName{Org: "autom8y", Repo: "knossos", Domain: "feat/materialization"},
			want: "autom8y::knossos::feat/materialization",
		},
		{
			name: "scoped",
			q:    QualifiedDomainName{Org: "autom8y", Repo: "autom8y", Scope: "services/ads", Domain: "architecture"},
			want: "autom8y::autom8y/services/ads::architecture",
		},
		{
			name: "deep scope with domain slash",
			q:    QualifiedDomainName{Org: "autom8y", Repo: "autom8y", Scope: "sdks/python/meta", Domain: "feat/x"},
			want: "autom8y::autom8y/sdks/python/meta::feat/x",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.q.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseRoundTrip(t *testing.T) {
	names := []string{
		"autom8y::knossos::architecture",
		"autom8y::knossos::feat/materialization",
		"my-org::my-repo::release/v2",
		"autom8y::autom8y/services/ads::architecture",
		"autom8y::autom8y/sdks/python/autom8y-meta::conventions",
		"autom8y::autom8y/services/ads::feat/materialization",
	}

	for _, s := range names {
		q, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", s, err)
		}
		if got := q.String(); got != s {
			t.Errorf("round-trip: Parse(%q).String() = %q", s, got)
		}
	}
}

func TestRepoSegment(t *testing.T) {
	tests := []struct {
		q    QualifiedDomainName
		want string
	}{
		{QualifiedDomainName{Repo: "knossos"}, "knossos"},
		{QualifiedDomainName{Repo: "autom8y", Scope: "services/ads"}, "autom8y/services/ads"},
	}
	for _, tc := range tests {
		if got := tc.q.RepoSegment(); got != tc.want {
			t.Errorf("RepoSegment() = %q, want %q", got, tc.want)
		}
	}
}

func TestNew(t *testing.T) {
	q := New("autom8y", "knossos", "architecture")
	if q.Scope != "" {
		t.Errorf("New() should have empty scope, got %q", q.Scope)
	}
	if q.String() != "autom8y::knossos::architecture" {
		t.Errorf("New().String() = %q", q.String())
	}
}

func TestNewScoped(t *testing.T) {
	q := NewScoped("autom8y", "autom8y", "services/ads", "architecture")
	if q.Scope != "services/ads" {
		t.Errorf("Scope = %q, want services/ads", q.Scope)
	}
	if q.String() != "autom8y::autom8y/services/ads::architecture" {
		t.Errorf("String() = %q", q.String())
	}
}

func TestRepoFromQualifiedName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"autom8y::knossos::architecture", "knossos"},
		{"autom8y::autom8y/services/ads::architecture", "autom8y"},
		{"invalid", ""},
	}
	for _, tc := range tests {
		if got := RepoFromQualifiedName(tc.input); got != tc.want {
			t.Errorf("RepoFromQualifiedName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
