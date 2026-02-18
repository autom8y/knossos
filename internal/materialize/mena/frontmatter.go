package mena

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/frontmatter"
)

// FlexibleStringSlice is an alias for the shared frontmatter.FlexibleStringSlice type.
// It accepts both comma-separated strings and YAML lists.
type FlexibleStringSlice = frontmatter.FlexibleStringSlice

// MenaFrontmatter represents the unified frontmatter schema for commands.
// Mena content is either dromena (.dro.md, enacted via /name) or legomena (.lego.md, reference knowledge).
type MenaFrontmatter struct {
	// Identity (required for all)
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Invocation Control
	ArgumentHint           string              `yaml:"argument-hint,omitempty"`
	Triggers               FlexibleStringSlice `yaml:"triggers,omitempty"`
	AllowedTools           FlexibleStringSlice `yaml:"allowed-tools,omitempty"`
	Model                  string              `yaml:"model,omitempty"`
	DisableModelInvocation bool                `yaml:"disable-model-invocation,omitempty"`
	DisallowedTools        FlexibleStringSlice `yaml:"disallowed-tools,omitempty"`
	Context                string              `yaml:"context,omitempty"`

	// Optional Metadata
	Version      string `yaml:"version,omitempty"`
	Deprecated   bool   `yaml:"deprecated,omitempty"`
	DeprecatedBy string `yaml:"deprecated-by,omitempty"`
}

// Validate checks that the frontmatter has required fields and valid values.
func (f *MenaFrontmatter) Validate() error {
	if f.Name == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: name is required")
	}
	if f.Description == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: description is required")
	}
	return nil
}

// readMenaFrontmatterFromDir reads the INDEX file from a filesystem directory,
// parses its YAML frontmatter, and returns the result.
func readMenaFrontmatterFromDir(dirPath string) MenaFrontmatter {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return MenaFrontmatter{}
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			data, err := os.ReadFile(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				return MenaFrontmatter{}
			}
			return ParseMenaFrontmatterBytes(data)
		}
	}
	return MenaFrontmatter{}
}

// readMenaFrontmatterFromFile reads a standalone mena file and parses its
// YAML frontmatter.
func readMenaFrontmatterFromFile(filePath string) MenaFrontmatter {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return MenaFrontmatter{}
	}
	return ParseMenaFrontmatterBytes(data)
}

// ParseMenaFrontmatterBytes extracts YAML frontmatter from raw file bytes.
// Returns a zero-value MenaFrontmatter if no frontmatter delimiters are found
// or if YAML parsing fails. Parse failures are silent (the entry is treated
// as unscoped per EC-7 in the PRD).
func ParseMenaFrontmatterBytes(data []byte) MenaFrontmatter {
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		return MenaFrontmatter{}
	}

	// Find closing delimiter
	var endIndex int
	searchStart := 4
	if bytes.HasPrefix(data, []byte("---\r\n")) {
		searchStart = 5
	}
	if idx := bytes.Index(data[searchStart:], []byte("\n---\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\r\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\r\n---\n")); idx != -1 {
		endIndex = idx
	} else {
		return MenaFrontmatter{}
	}

	var fm MenaFrontmatter
	if err := yaml.Unmarshal(data[searchStart:searchStart+endIndex], &fm); err != nil {
		log.Printf("Warning: malformed YAML frontmatter, treating as unscoped: %v", err)
		return MenaFrontmatter{}
	}
	return fm
}
