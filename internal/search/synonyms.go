package search

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// synonymWeightFactor is the reduction factor for expanded-token scores.
	// 0.6 means expanded tokens score at 60% of original token weights.
	synonymWeightFactor = 0.6

	// maxExpansionsPerToken caps synonym expansions to prevent score inflation.
	maxExpansionsPerToken = 6

	// minTokenLengthForExpansion prevents expanding ambiguous short tokens.
	minTokenLengthForExpansion = 3
)

// SynonymSource provides synonym mappings for query expansion.
// Implementations must be safe for concurrent read access after initialization.
type SynonymSource interface {
	// Expand returns additional tokens that the given token maps to.
	// Returns nil if no synonyms exist for the token.
	// The input token is NOT included in the output.
	Expand(token string) []string
}

// defaultSynonyms is the hardcoded baseline mapping for vocabulary that cannot
// be derived from orchestrator.yaml because it bridges colloquial terms to
// ecosystem concepts.
var defaultSynonyms = map[string][]string{
	"deploy":   {"sre", "operations", "reliability", "infrastructure"},
	"refactor": {"hygiene", "cleanup", "code-quality"},
	"ship":     {"releaser", "release", "publish"},
	"audit":    {"review", "security", "compliance"},
	"test":     {"qa", "validation", "testing"},
	"docs":     {"documentation", "technical-writing"},
	"document": {"documentation", "docs", "technical-writing"},
	"build":    {"forge", "compile", "implementation"},
	"debug":    {"clinic", "diagnose", "troubleshoot"},
	"plan":     {"strategy", "planning", "roadmap"},
	"research": {"rnd", "spike", "investigation"},
}

// StaticSynonymSource provides hardcoded synonym mappings.
type StaticSynonymSource struct {
	synonyms map[string][]string
}

// NewStaticSynonymSource returns the default static synonym mappings.
func NewStaticSynonymSource() *StaticSynonymSource {
	return &StaticSynonymSource{synonyms: defaultSynonyms}
}

// Expand returns additional tokens that the given token maps to.
func (s *StaticSynonymSource) Expand(token string) []string {
	if s == nil || s.synonyms == nil {
		return nil
	}
	return s.synonyms[strings.ToLower(token)]
}

// OrchestratorSynonymSource derives synonym mappings from orchestrator.yaml
// Triggers and UseWhen patterns. Each trigger token maps to the rite name and domain.
type OrchestratorSynonymSource struct {
	synonyms map[string][]string // token -> expansion targets
}

// NewOrchestratorSynonymSource scans orchestrator.yaml files under ritesDir.
// Returns an empty source (not nil) if ritesDir is unreadable.
func NewOrchestratorSynonymSource(ritesDir string) *OrchestratorSynonymSource {
	src := &OrchestratorSynonymSource{
		synonyms: make(map[string][]string),
	}

	dirEntries, err := os.ReadDir(ritesDir)
	if err != nil {
		return src
	}

	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}

		orchPath := filepath.Join(ritesDir, de.Name(), "orchestrator.yaml")
		data, err := os.ReadFile(orchPath)
		if err != nil {
			continue
		}

		var orch orchestratorFile
		if err := yaml.Unmarshal(data, &orch); err != nil {
			continue
		}

		riteName := strings.ToLower(de.Name())
		domain := strings.ToLower(strings.TrimSpace(orch.Rite.Domain))

		// Extract keywords from frontmatter description (Triggers/UseWhen).
		keywords := extractKeywords(orch.Frontmatter.Description)

		// Each keyword token maps to the rite name and domain.
		targets := []string{riteName}
		if domain != "" && domain != riteName {
			targets = append(targets, domain)
		}

		for _, kw := range keywords {
			// Use individual words from multi-word keywords.
			for _, word := range strings.Fields(kw) {
				word = strings.TrimSpace(word)
				if word == "" || word == riteName {
					continue
				}
				src.addMapping(word, targets)
			}
		}
	}

	return src
}

// addMapping adds target expansions for a token, deduplicating.
func (s *OrchestratorSynonymSource) addMapping(token string, targets []string) {
	existing := s.synonyms[token]
	seen := make(map[string]bool, len(existing))
	for _, t := range existing {
		seen[t] = true
	}
	for _, t := range targets {
		if !seen[t] {
			existing = append(existing, t)
			seen[t] = true
		}
	}
	s.synonyms[token] = existing
}

// Expand returns additional tokens that the given token maps to.
func (s *OrchestratorSynonymSource) Expand(token string) []string {
	if s == nil || s.synonyms == nil {
		return nil
	}
	return s.synonyms[strings.ToLower(token)]
}

// CompositeSynonymSource merges multiple SynonymSources.
// Expansions are deduplicated. Earlier sources have lower precedence
// (later sources can add but not remove expansions).
type CompositeSynonymSource struct {
	sources []SynonymSource
}

// NewCompositeSynonymSource merges multiple sources.
func NewCompositeSynonymSource(sources ...SynonymSource) *CompositeSynonymSource {
	return &CompositeSynonymSource{sources: sources}
}

// Expand returns the merged, deduplicated expansions from all sources.
func (c *CompositeSynonymSource) Expand(token string) []string {
	if c == nil || len(c.sources) == 0 {
		return nil
	}

	var result []string
	seen := make(map[string]bool)

	for _, src := range c.sources {
		if src == nil {
			continue
		}
		for _, exp := range src.Expand(token) {
			if !seen[exp] {
				seen[exp] = true
				result = append(result, exp)
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// expandSynonyms returns synonym expansions for a token, subject to guards.
// Returns nil if the token is too short, has no synonyms, or the source is nil.
// Expansions are capped at maxExpansionsPerToken.
func expandSynonyms(token string, source SynonymSource) []string {
	if source == nil {
		return nil
	}
	if len(token) < minTokenLengthForExpansion {
		return nil
	}

	expansions := source.Expand(token)
	if len(expansions) > maxExpansionsPerToken {
		expansions = expansions[:maxExpansionsPerToken]
	}
	return expansions
}
