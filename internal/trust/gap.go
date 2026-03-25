package trust

import (
	"fmt"
	"strings"
)

// GapAdmission describes what Clew says when confidence is LOW.
// It admits ignorance and provides actionable guidance for the user.
type GapAdmission struct {
	// MissingDomains lists domains that the query requires but the registry does not contain.
	MissingDomains []string

	// StaleDomains lists domains found in the registry but below the freshness threshold.
	StaleDomains []StaleDomainInfo

	// Suggestions is an ordered list of actionable commands the user can run.
	Suggestions []string

	// Reason is a human-readable explanation of why the confidence is LOW.
	Reason string
}

// StaleDomainInfo provides details about a stale domain for display.
type StaleDomainInfo struct {
	// QualifiedName is the full domain address.
	QualifiedName string
	// Domain is the bare domain name.
	Domain string
	// Repo is the repository name.
	Repo string
	// Freshness is the current freshness score (0.0-1.0).
	Freshness float64
	// DaysSinceGenerated is the number of days since the domain was generated.
	DaysSinceGenerated int
}

// NewGapAdmission constructs a GapAdmission from missing and stale domain information.
// Generates actionable suggestions referencing real repos and domains from the registry.
func NewGapAdmission(missingDomains []string, staleDomains []StaleDomainInfo) GapAdmission {
	gap := GapAdmission{
		MissingDomains: missingDomains,
		StaleDomains:   staleDomains,
	}

	// Build reason
	var reasons []string
	if len(missingDomains) > 0 {
		reasons = append(reasons, fmt.Sprintf("no knowledge found for: %s", strings.Join(missingDomains, ", ")))
	}
	if len(staleDomains) > 0 {
		staleNames := make([]string, len(staleDomains))
		for i, sd := range staleDomains {
			staleNames[i] = sd.QualifiedName
		}
		reasons = append(reasons, fmt.Sprintf("stale knowledge in: %s", strings.Join(staleNames, ", ")))
	}
	if len(reasons) == 0 {
		gap.Reason = "insufficient knowledge to answer this question reliably"
	} else {
		gap.Reason = strings.Join(reasons, "; ")
	}

	// Generate suggestions
	gap.Suggestions = generateSuggestions(missingDomains, staleDomains)

	return gap
}

// generateSuggestions creates actionable /know commands from gap information.
// Suggestions reference real repos and domains, not fabricated ones.
func generateSuggestions(missingDomains []string, staleDomains []StaleDomainInfo) []string {
	var suggestions []string

	// For missing domains: suggest generating knowledge
	for _, domain := range missingDomains {
		suggestions = append(suggestions,
			fmt.Sprintf("Run `/know --domain=%s` in the relevant repository to generate this knowledge", domain))
	}

	// For stale domains: suggest regenerating with specific repo context
	for _, sd := range staleDomains {
		suggestions = append(suggestions,
			fmt.Sprintf("Run `/know --domain=%s` in repo %s to refresh (last generated %d days ago)",
				sd.Domain, sd.Repo, sd.DaysSinceGenerated))
	}

	return suggestions
}

// SuggestionFor generates a single actionable suggestion for a specific domain gap.
// Used when a single domain is the bottleneck (e.g., a query about one specific topic).
func SuggestionFor(domain, repo string) string {
	if repo != "" {
		return fmt.Sprintf("Run `/know --domain=%s` in repo %s", domain, repo)
	}
	return fmt.Sprintf("Run `/know --domain=%s` in the relevant repository", domain)
}

// HasGaps returns true if there are any missing or stale domains.
func (ga *GapAdmission) HasGaps() bool {
	return len(ga.MissingDomains) > 0 || len(ga.StaleDomains) > 0
}

// IsEmpty returns true if there are no gaps at all.
func (ga *GapAdmission) IsEmpty() bool {
	return len(ga.MissingDomains) == 0 && len(ga.StaleDomains) == 0
}
