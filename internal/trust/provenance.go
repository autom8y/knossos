package trust

import (
	"time"
)

// ProvenanceLink represents a single source citation in a provenance chain.
// Each link traces to a specific .know/ file with its generation metadata.
// Constructable from DomainEntry fields.
type ProvenanceLink struct {
	// QualifiedName is the canonical cross-repo address: "org::repo::domain".
	QualifiedName string

	// GeneratedAt is the timestamp when the source .know/ file was generated.
	GeneratedAt time.Time

	// SourceHash is the git short SHA recorded in the .know/ frontmatter.
	SourceHash string

	// FilePath is the repo-relative path to the .know/ file (e.g., ".know/architecture.md").
	FilePath string

	// Excerpt is an optional section-level excerpt from the source.
	// Empty when the full domain is cited rather than a specific section.
	Excerpt string

	// Domain is the bare domain name (e.g., "architecture", "feat/materialization").
	Domain string

	// Repo is the repository name.
	Repo string

	// FreshnessAtQuery is the freshness score at the time the chain was built.
	// Computed via the decay model. Enables downstream display of per-source staleness.
	FreshnessAtQuery float64
}

// ProvenanceChain is an ordered collection of source citations backing a Clew response.
// The chain is NEVER empty for HIGH or MEDIUM confidence responses.
// It MAY be empty for LOW confidence responses (where GapAdmission carries the information).
type ProvenanceChain struct {
	// Sources is the ordered list of provenance links, most relevant first.
	Sources []ProvenanceLink

	// BuiltAt is the timestamp when the chain was assembled.
	BuiltAt time.Time
}

// ProvenanceLinkInput holds the raw data needed to construct a ProvenanceLink.
// This struct decouples the trust package from DomainEntry's concrete type,
// allowing construction from any source that provides the required fields.
type ProvenanceLinkInput struct {
	QualifiedName string
	GeneratedAt   string // RFC3339 -- parsed internally
	SourceHash    string
	FilePath      string
	Domain        string
	Repo          string
	Excerpt       string
}

// NewProvenanceChain constructs a ProvenanceChain from a slice of inputs.
// Computes freshness for each link using the provided DecayConfig and current time.
// Links with unparseable GeneratedAt timestamps are included with FreshnessAtQuery=0.0
// (the link is still valid provenance; the freshness is just unknown).
func NewProvenanceChain(inputs []ProvenanceLinkInput, decay *DecayConfig, now time.Time) ProvenanceChain {
	chain := ProvenanceChain{
		Sources: make([]ProvenanceLink, 0, len(inputs)),
		BuiltAt: now,
	}

	for _, in := range inputs {
		link := ProvenanceLink{
			QualifiedName: in.QualifiedName,
			SourceHash:    in.SourceHash,
			FilePath:      in.FilePath,
			Excerpt:       in.Excerpt,
			Domain:        in.Domain,
			Repo:          in.Repo,
		}

		generatedAt, err := time.Parse(time.RFC3339, in.GeneratedAt)
		if err != nil {
			link.GeneratedAt = time.Time{} // zero value signals unparseable
			link.FreshnessAtQuery = 0.0
		} else {
			link.GeneratedAt = generatedAt
			link.FreshnessAtQuery = decay.DecayFromString(in.GeneratedAt, in.Domain, now)
		}

		chain.Sources = append(chain.Sources, link)
	}

	return chain
}

// Len returns the number of sources in the chain.
func (pc *ProvenanceChain) Len() int {
	return len(pc.Sources)
}

// IsEmpty returns true if the chain has no sources.
func (pc *ProvenanceChain) IsEmpty() bool {
	return len(pc.Sources) == 0
}

// QualifiedNames returns the list of qualified domain names in the chain.
func (pc *ProvenanceChain) QualifiedNames() []string {
	names := make([]string, len(pc.Sources))
	for i, s := range pc.Sources {
		names[i] = s.QualifiedName
	}
	return names
}

// MinFreshness returns the lowest freshness score across all sources.
// Returns 0.0 for an empty chain.
func (pc *ProvenanceChain) MinFreshness() float64 {
	if len(pc.Sources) == 0 {
		return 0.0
	}
	min := pc.Sources[0].FreshnessAtQuery
	for _, s := range pc.Sources[1:] {
		if s.FreshnessAtQuery < min {
			min = s.FreshnessAtQuery
		}
	}
	return min
}

// MeanFreshness returns the average freshness score across all sources.
// Returns 0.0 for an empty chain.
func (pc *ProvenanceChain) MeanFreshness() float64 {
	if len(pc.Sources) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, s := range pc.Sources {
		sum += s.FreshnessAtQuery
	}
	return sum / float64(len(pc.Sources))
}

// WeightedMeanFreshness returns the position-weighted mean freshness score.
// Sources earlier in the chain (more relevant) receive higher weight.
// Weight for source at position i (0-indexed) = n - i, where n = total sources.
// This gives a linear decay from most-relevant to least-relevant source.
//
// Example with 3 sources [0.9, 0.5, 0.2]:
//
//	weighted = (3*0.9 + 2*0.5 + 1*0.2) / (3+2+1) = (2.7+1.0+0.2) / 6 = 0.65
//
// vs MinFreshness = 0.2, MeanFreshness = 0.533
//
// Returns 0.0 for an empty chain.
func (pc *ProvenanceChain) WeightedMeanFreshness() float64 {
	n := len(pc.Sources)
	if n == 0 {
		return 0.0
	}
	var weightedSum, weightSum float64
	for i, s := range pc.Sources {
		weight := float64(n - i)
		weightedSum += weight * s.FreshnessAtQuery
		weightSum += weight
	}
	if weightSum == 0 {
		return 0.0
	}
	return weightedSum / weightSum
}

// StaleSources returns sources with freshness below the given threshold.
func (pc *ProvenanceChain) StaleSources(threshold float64) []ProvenanceLink {
	var stale []ProvenanceLink
	for _, s := range pc.Sources {
		if s.FreshnessAtQuery < threshold {
			stale = append(stale, s)
		}
	}
	return stale
}
