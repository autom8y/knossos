// Package materialize provides frontmatter parsing for command files.
package materialize

import (
	"bytes"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"gopkg.in/yaml.v3"
)

// FlexibleStringSlice is a YAML type that accepts both a comma-separated string
// (e.g., "Bash, Read, Glob") and a proper YAML list (e.g., [Bash, Read, Glob]).
// This handles the common pattern in command frontmatter where tools are listed inline.
type FlexibleStringSlice []string

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (f *FlexibleStringSlice) UnmarshalYAML(value *yaml.Node) error {
	// Try as a sequence first
	if value.Kind == yaml.SequenceNode {
		var slice []string
		if err := value.Decode(&slice); err != nil {
			return err
		}
		*f = slice
		return nil
	}

	// Fall back to comma-separated string
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}

	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	*f = result
	return nil
}

// MenaFrontmatter represents the unified frontmatter schema for commands.
// Mena content is either dromena (.dro.md, enacted via /name) or legomena (.lego.md, reference knowledge).
type MenaFrontmatter struct {
	// Identity (required for all)
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Invocation Control
	ArgumentHint string   `yaml:"argument-hint,omitempty"` // Only for dromena. Usage hint
	Triggers     FlexibleStringSlice `yaml:"triggers,omitempty"`      // Auto-invocation keywords
	AllowedTools FlexibleStringSlice `yaml:"allowed-tools,omitempty"` // Tool restrictions (only for dromena)
	Model        string   `yaml:"model,omitempty"`         // Model selection (only for dromena)

	// Optional Metadata
	Version      string `yaml:"version,omitempty"`       // Semantic version for tracking
	Deprecated   bool   `yaml:"deprecated,omitempty"`    // Mark command as deprecated
	DeprecatedBy string `yaml:"deprecated-by,omitempty"` // Reference to replacement command
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

// ParseMenaFrontmatter extracts frontmatter from a command file.
// Returns error if frontmatter is missing or invalid.
func ParseMenaFrontmatter(content []byte) (*MenaFrontmatter, error) {
	// Find frontmatter delimiters
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return nil, errors.New(errors.CodeParseError, "missing frontmatter delimiter")
	}

	// Find closing delimiter
	var endIndex int
	if idx := bytes.Index(content[4:], []byte("\n---\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(content[4:], []byte("\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(content[4:], []byte("\r\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(content[4:], []byte("\r\n---\n")); idx != -1 {
		endIndex = idx
	} else {
		return nil, errors.New(errors.CodeParseError, "missing closing frontmatter delimiter")
	}

	frontmatterBytes := content[4 : 4+endIndex]

	var fm MenaFrontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &fm); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "invalid frontmatter YAML", err)
	}

	return &fm, nil
}

// DetectMenaType determines content type from file extension convention.
// Files with .dro.md extension are dromena (invokable, project to .claude/commands/).
// Files with .lego.md extension are legomena (reference, project to .claude/skills/).
// Returns "dro" as default for backward compatibility.
func DetectMenaType(filename string) string {
	if strings.Contains(filename, ".dro.") {
		return "dro"
	}
	if strings.Contains(filename, ".lego.") {
		return "lego"
	}
	return "dro" // default for backward compat
}
