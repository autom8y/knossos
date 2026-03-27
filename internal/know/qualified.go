// Package know provides shared parsing for .know/ codebase knowledge files.
package know

import (
	"fmt"
	"strings"
)

// QualifiedDomainName is the canonical cross-repo knowledge address.
// Format: "org::repo[/scope]::domain"
//   - org:    GitHub organization name
//   - repo:   GitHub repository name (no "/" -- GitHub constraint)
//   - scope:  path from repo root to .know/ directory (may contain "/", empty for root)
//   - domain: bare domain name (may contain "/" for feat/X, release/X)
//
// The first "/" in the second "::" segment separates repo from scope.
// GitHub repo names cannot contain "/", making this unambiguous.
type QualifiedDomainName struct {
	Org    string
	Repo   string
	Scope  string
	Domain string
}

// Parse parses "org::repo[/scope]::domain" into its components.
func Parse(s string) (QualifiedDomainName, error) {
	if s == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name must not be empty")
	}

	parts := strings.SplitN(s, "::", 3)
	if len(parts) != 3 {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q must have format org::repo::domain", s)
	}

	org, repoSegment, domain := parts[0], parts[1], parts[2]

	if strings.TrimSpace(org) == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: org segment must not be empty", s)
	}
	if strings.TrimSpace(repoSegment) == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: repo segment must not be empty", s)
	}
	if strings.TrimSpace(domain) == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: domain segment must not be empty", s)
	}
	if strings.Contains(domain, "::") {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: domain segment must not contain '::'", s)
	}

	repo, scope, hasScope := strings.Cut(repoSegment, "/")
	if repo == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: repo name must not be empty", s)
	}
	if hasScope && scope == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: scope must not be empty when '/' present", s)
	}

	return QualifiedDomainName{Org: org, Repo: repo, Scope: scope, Domain: domain}, nil
}

// String returns the canonical string form.
func (q QualifiedDomainName) String() string {
	if q.Scope == "" {
		return q.Org + "::" + q.Repo + "::" + q.Domain
	}
	return q.Org + "::" + q.Repo + "/" + q.Scope + "::" + q.Domain
}

// RepoSegment returns the full second "::" segment ("repo" or "repo/scope").
func (q QualifiedDomainName) RepoSegment() string {
	if q.Scope == "" {
		return q.Repo
	}
	return q.Repo + "/" + q.Scope
}

// New creates a root-scope QualifiedDomainName.
func New(org, repo, domain string) QualifiedDomainName {
	return QualifiedDomainName{Org: org, Repo: repo, Domain: domain}
}

// NewScoped creates a scoped QualifiedDomainName. Pass "" for root scope.
func NewScoped(org, repo, scope, domain string) QualifiedDomainName {
	return QualifiedDomainName{Org: org, Repo: repo, Scope: scope, Domain: domain}
}

// RepoFromQualifiedName extracts just the repo name from a qualified name string.
// Handles scoped names: "org::repo/scope::domain" returns "repo".
func RepoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) < 2 {
		return ""
	}
	repo, _, _ := strings.Cut(parts[1], "/")
	return repo
}
