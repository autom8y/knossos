// Package team implements team pack discovery, management, and switching for Ariadne.
// This file contains team context types for YAML-based context injection per ADR-0007.
package team

import (
	"fmt"
	"strings"
)

// TeamContext represents the context injection configuration for a team.
// This is the Go representation of teams/{team}/context.yaml files.
type TeamContext struct {
	SchemaVersion string            `yaml:"schema_version" json:"schema_version"`
	TeamName      string            `yaml:"team_name" json:"team_name"`
	DisplayName   string            `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Description   string            `yaml:"description,omitempty" json:"description,omitempty"`
	Domain        string            `yaml:"domain,omitempty" json:"domain,omitempty"`
	ContextRows   []ContextRow      `yaml:"context_rows" json:"context_rows"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// ContextRow represents a single key-value pair for context injection.
type ContextRow struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

// ToMarkdown renders the team context as a markdown table.
// Format matches the existing bash output:
// | **Key** | Value |
func (tc *TeamContext) ToMarkdown() string {
	if tc == nil || len(tc.ContextRows) == 0 {
		return ""
	}

	var b strings.Builder

	// Write table header (empty header row for compatibility with existing format)
	b.WriteString("| | |\n")
	b.WriteString("|---|---|\n")

	// Write context rows
	for _, row := range tc.ContextRows {
		b.WriteString(fmt.Sprintf("| **%s** | %s |\n", row.Key, row.Value))
	}

	return b.String()
}

// NewTeamContext creates a new TeamContext with default values.
func NewTeamContext(teamName string) *TeamContext {
	return &TeamContext{
		SchemaVersion: "1.0",
		TeamName:      teamName,
		ContextRows:   []ContextRow{},
		Metadata:      make(map[string]string),
	}
}

// AddRow adds a context row to the TeamContext.
func (tc *TeamContext) AddRow(key, value string) {
	tc.ContextRows = append(tc.ContextRows, ContextRow{
		Key:   key,
		Value: value,
	})
}

// GetRow returns the value for a given key, or empty string if not found.
func (tc *TeamContext) GetRow(key string) string {
	for _, row := range tc.ContextRows {
		if row.Key == key {
			return row.Value
		}
	}
	return ""
}

// HasRows returns true if the context has any rows.
func (tc *TeamContext) HasRows() bool {
	return len(tc.ContextRows) > 0
}

// Validate checks that the TeamContext has required fields.
func (tc *TeamContext) Validate() error {
	if tc.TeamName == "" {
		return fmt.Errorf("team_name is required")
	}
	if tc.SchemaVersion == "" {
		return fmt.Errorf("schema_version is required")
	}
	return nil
}
