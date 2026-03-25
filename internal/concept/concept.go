// Package concept provides the knossos concept registry.
// Concepts are embedded markdown files with YAML frontmatter that document
// domain-specific terminology (rite, session, mena, dromena, legomena, etc.).
//
// This package was extracted from internal/cmd/explain to resolve TENSION-015:
// internal/search imported internal/cmd/explain across the layer boundary.
// Moving the concept registry to a domain package keeps the import graph clean.
package concept

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
)

//go:embed concepts/*.md
var conceptFS embed.FS

// ConceptEntry holds a fully parsed concept definition.
type ConceptEntry struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	SeeAlso     []string `json:"see_also"`
	Aliases     []string `json:"aliases"`
	HarnessTerm string   `json:"harness_term,omitempty"`
}

// Frontmatter represents the YAML frontmatter parsed from a concept markdown file.
type Frontmatter struct {
	Summary     string   `yaml:"summary"`
	SeeAlso     []string `yaml:"see_also"`
	Aliases     []string `yaml:"aliases"`
	HarnessTerm string   `yaml:"harness_term"`
}

var (
	// registry maps canonical concept names to their parsed entries.
	// Populated by init() from embedded concept files.
	registry map[string]*ConceptEntry

	// aliases maps alias names to canonical concept names.
	// Populated by init() from frontmatter aliases fields.
	aliases map[string]string

	// sortedNames holds all concept names in alphabetical order.
	// Used for listing and error messages.
	sortedNames []string
)

func init() {
	registry = make(map[string]*ConceptEntry)
	aliases = make(map[string]string)

	entries, err := fs.ReadDir(conceptFS, "concepts")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded concepts: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		data, err := conceptFS.ReadFile("concepts/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("failed to read concept %s: %v", name, err))
		}

		c, err := parseConcept(name, data)
		if err != nil {
			panic(fmt.Sprintf("failed to parse concept %s: %v", name, err))
		}

		registry[name] = c

		// Register aliases
		for _, alias := range c.Aliases {
			aliases[strings.ToLower(alias)] = name
		}
	}

	// Build sorted name list
	sortedNames = make([]string, 0, len(registry))
	for name := range registry {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)
}

// parseConcept parses a concept markdown file with YAML frontmatter.
// Format: --- <YAML frontmatter> --- <markdown body>
func parseConcept(name string, data []byte) (*ConceptEntry, error) {
	content := string(data)

	// Validate frontmatter delimiters
	if !strings.HasPrefix(content, "---\n") {
		return nil, errors.New(errors.CodeParseError, "missing opening frontmatter delimiter")
	}

	// Find closing delimiter
	rest := content[4:] // skip "---\n"
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		// Try trailing "---" at end of file (edge case)
		idx = strings.Index(rest, "\n---")
		if idx < 0 {
			return nil, errors.New(errors.CodeParseError, "missing closing frontmatter delimiter")
		}
	}

	frontmatterYAML := rest[:idx]
	body := strings.TrimSpace(rest[idx+4:]) // skip "\n---\n" or "\n---"

	// Parse YAML frontmatter
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
		return nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}

	if fm.Summary == "" {
		return nil, errors.New(errors.CodeParseError, "missing required field: summary")
	}

	// Ensure SeeAlso is non-nil
	if fm.SeeAlso == nil {
		fm.SeeAlso = []string{}
	}

	// Ensure Aliases is non-nil
	if fm.Aliases == nil {
		fm.Aliases = []string{}
	}

	// Compute display name
	displayName := name
	if fm.HarnessTerm != "" {
		displayName = name + " (" + fm.HarnessTerm + ")"
	}

	return &ConceptEntry{
		Name:        name,
		DisplayName: displayName,
		Summary:     fm.Summary,
		Description: body,
		SeeAlso:     fm.SeeAlso,
		Aliases:     fm.Aliases,
		HarnessTerm: fm.HarnessTerm,
	}, nil
}

// LookupConcept resolves a user-provided name to a ConceptEntry.
// Lookup chain: exact match -> alias match -> error with suggestion.
// Input is case-normalized to lowercase.
func LookupConcept(input string) (*ConceptEntry, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))

	// 1. Exact match
	if entry, ok := registry[normalized]; ok {
		return entry, nil
	}

	// 2. Alias match
	if canonical, ok := aliases[normalized]; ok {
		return registry[canonical], nil
	}

	// 3. No match -- compute suggestion
	suggestion := suggestConcept(normalized)
	if suggestion != "" {
		return nil, fmt.Errorf(
			"unknown concept %q. Did you mean %q?\n\nAvailable concepts: %s",
			input, suggestion, strings.Join(sortedNames, ", "))
	}

	return nil, fmt.Errorf(
		"unknown concept %q\n\nAvailable concepts: %s",
		input, strings.Join(sortedNames, ", "))
}

// AllConcepts returns all concept entries sorted alphabetically by name.
func AllConcepts() []*ConceptEntry {
	result := make([]*ConceptEntry, len(sortedNames))
	for i, name := range sortedNames {
		result[i] = registry[name]
	}
	return result
}

// levenshtein computes the Levenshtein edit distance between two strings.
// Standard dynamic programming algorithm, O(m*n) time and space.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use single-row optimization: O(min(m,n)) space
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = minOf3(
				curr[j-1]+1,    // insertion
				prev[j]+1,      // deletion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// suggestConcept returns the best matching concept name for a misspelled input,
// or empty string if no close match exists.
// Threshold: distance <= 3 AND distance < len(input)/2.
func suggestConcept(input string) string {
	bestName := ""
	bestDist := len(input) // start with worst case

	for _, name := range sortedNames {
		dist := levenshtein(input, name)
		if dist < bestDist {
			bestDist = dist
			bestName = name
		}
	}

	// Also check aliases
	for alias, canonical := range aliases {
		dist := levenshtein(input, alias)
		if dist < bestDist {
			bestDist = dist
			bestName = canonical
		}
	}

	// Apply threshold: suggest only if close enough
	if bestDist <= 3 && bestDist < len(input)/2 {
		return bestName
	}

	return ""
}

// minOf3 returns the minimum of three integers.
func minOf3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
