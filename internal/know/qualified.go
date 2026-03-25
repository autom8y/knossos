// Package know provides shared parsing for .know/ codebase knowledge files.
package know

import (
	"fmt"
	"strings"
)

// QualifiedDomainName is the canonical cross-repo knowledge address.
// Format: "org::repo::domain" where domain may contain "/" (e.g., "feat/materialization").
// The "::" separator is consistent with the existing monorepo prefix convention
// used in ReadMeta() which prefixes nested .know/ domains with "service/path::domain".
type QualifiedDomainName struct {
	Org    string
	Repo   string
	Domain string
}

// Parse parses a qualified domain name string of the form "org::repo::domain".
// The domain segment may contain "/" characters (e.g., "feat/materialization").
// Returns an error if the string does not contain exactly two "::" separators or
// if any segment is empty.
func Parse(s string) (QualifiedDomainName, error) {
	if s == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name must not be empty")
	}

	// Split on "::" to extract org, repo, and domain segments.
	// We split into at most 3 parts to allow domain to contain "::".
	// However, per the spec, domain may contain "/" but NOT "::".
	// We require exactly 2 occurrences of "::" — i.e., 3 segments.
	parts := strings.SplitN(s, "::", 3)
	if len(parts) != 3 {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q must have format org::repo::domain", s)
	}

	org := parts[0]
	repo := parts[1]
	domain := parts[2]

	if strings.TrimSpace(org) == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: org segment must not be empty", s)
	}
	if strings.TrimSpace(repo) == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: repo segment must not be empty", s)
	}
	if strings.TrimSpace(domain) == "" {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: domain segment must not be empty", s)
	}
	// Reject "::" inside domain — domain may only contain "/" as a sub-separator
	if strings.Contains(domain, "::") {
		return QualifiedDomainName{}, fmt.Errorf("qualified domain name %q: domain segment must not contain '::'", s)
	}

	return QualifiedDomainName{
		Org:    org,
		Repo:   repo,
		Domain: domain,
	}, nil
}

// String returns the canonical string form: "org::repo::domain".
func (q QualifiedDomainName) String() string {
	return q.Org + "::" + q.Repo + "::" + q.Domain
}
