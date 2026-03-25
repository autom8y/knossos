package know

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOrg   string
		wantRepo  string
		wantDomain string
		wantErr   bool
	}{
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
			name:       "domain with deep slash path",
			input:      "autom8y::payments-service::release/platform-profile",
			wantOrg:    "autom8y",
			wantRepo:   "payments-service",
			wantDomain: "release/platform-profile",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing repo and domain",
			input:   "autom8y",
			wantErr: true,
		},
		{
			name:    "missing domain only",
			input:   "autom8y::knossos",
			wantErr: true,
		},
		{
			name:    "empty org segment",
			input:   "::knossos::architecture",
			wantErr: true,
		},
		{
			name:    "empty repo segment",
			input:   "autom8y::::architecture",
			wantErr: true,
		},
		{
			name:    "empty domain segment",
			input:   "autom8y::knossos::",
			wantErr: true,
		},
		{
			name:    "extra :: in domain rejected",
			input:   "autom8y::knossos::arch::extra",
			wantErr: true,
		},
		{
			name:    "whitespace-only org",
			input:   "   ::knossos::architecture",
			wantErr: true,
		},
		{
			name:    "whitespace-only domain",
			input:   "autom8y::knossos::   ",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error but got none; result=%+v", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tc.input, err)
			}
			if got.Org != tc.wantOrg {
				t.Errorf("Org: got %q, want %q", got.Org, tc.wantOrg)
			}
			if got.Repo != tc.wantRepo {
				t.Errorf("Repo: got %q, want %q", got.Repo, tc.wantRepo)
			}
			if got.Domain != tc.wantDomain {
				t.Errorf("Domain: got %q, want %q", got.Domain, tc.wantDomain)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		q     QualifiedDomainName
		want  string
	}{
		{
			name: "simple",
			q:    QualifiedDomainName{Org: "autom8y", Repo: "knossos", Domain: "architecture"},
			want: "autom8y::knossos::architecture",
		},
		{
			name: "domain with slash",
			q:    QualifiedDomainName{Org: "autom8y", Repo: "knossos", Domain: "feat/materialization"},
			want: "autom8y::knossos::feat/materialization",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.q.String()
			if got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseRoundTrip(t *testing.T) {
	originals := []string{
		"autom8y::knossos::architecture",
		"autom8y::knossos::feat/materialization",
		"my-org::my-repo::release/v2",
	}

	for _, s := range originals {
		q, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", s, err)
		}
		if q.String() != s {
			t.Errorf("round-trip: Parse(%q).String() = %q", s, q.String())
		}
	}
}
