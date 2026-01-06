// Package team implements team pack discovery, management, and switching for Ariadne.
// This file contains team context types for YAML-based context injection per ADR-0007.
package rite

import (
	"fmt"
	"strings"
)

// RiteContext represents the context injection configuration for a rite.
// This is the Go representation of rites/{rite}/context.yaml files.
type RiteContext struct {
	SchemaVersion string            `yaml:"schema_version" json:"schema_version"`
	TeamName      string            `yaml:"team_name" json:"team_name"` // Keep field name for YAML compatibility
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

// ToMarkdown renders the rite context as a markdown table.
// Format matches the existing bash output:
// | **Key** | Value |
func (rc *RiteContext) ToMarkdown() string {
	if rc == nil || len(rc.ContextRows) == 0 {
		return ""
	}

	var b strings.Builder

	// Write table header (empty header row for compatibility with existing format)
	b.WriteString("| | |\n")
	b.WriteString("|---|---|\n")

	// Write context rows
	for _, row := range rc.ContextRows {
		b.WriteString(fmt.Sprintf("| **%s** | %s |\n", row.Key, row.Value))
	}

	return b.String()
}

// NewTeamContext creates a new RiteContext with default values.
func NewTeamContext(riteName string) *RiteContext {
	return &RiteContext{
		SchemaVersion: "1.0",
		TeamName:      riteName,
		ContextRows:   []ContextRow{},
		Metadata:      make(map[string]string),
	}
}

// AddRow adds a context row to the RiteContext.
func (rc *RiteContext) AddRow(key, value string) {
	rc.ContextRows = append(rc.ContextRows, ContextRow{
		Key:   key,
		Value: value,
	})
}

// GetRow returns the value for a given key, or empty string if not found.
func (rc *RiteContext) GetRow(key string) string {
	for _, row := range rc.ContextRows {
		if row.Key == key {
			return row.Value
		}
	}
	return ""
}

// HasRows returns true if the context has any rows.
func (rc *RiteContext) HasRows() bool {
	return len(rc.ContextRows) > 0
}

// Validate checks that the RiteContext has required fields.
func (rc *RiteContext) Validate() error {
	if rc.TeamName == "" {
		return fmt.Errorf("team_name is required")
	}
	if rc.SchemaVersion == "" {
		return fmt.Errorf("schema_version is required")
	}
	return nil
}
