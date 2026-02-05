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

// CommandFrontmatter represents the unified frontmatter schema for commands.
// Commands can be invokable (user-callable via /name) or reference (auto-loaded patterns).
type CommandFrontmatter struct {
	// Identity (required for all)
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Invocation Control
	Invokable    *bool    `yaml:"invokable,omitempty"`     // Default: true. User-callable via /name
	ArgumentHint string   `yaml:"argument-hint,omitempty"` // Only for invokable=true. Usage hint
	Triggers     FlexibleStringSlice `yaml:"triggers,omitempty"`      // Auto-invocation keywords
	AllowedTools FlexibleStringSlice `yaml:"allowed-tools,omitempty"` // Tool restrictions (only for invokable=true)
	Model        string   `yaml:"model,omitempty"`         // Model selection (only for invokable=true)

	// Classification (for non-invokable)
	Category string `yaml:"category,omitempty"` // reference | template | schema. Required when invokable=false

	// Optional Metadata
	Version      string `yaml:"version,omitempty"`       // Semantic version for tracking
	Deprecated   bool   `yaml:"deprecated,omitempty"`    // Mark command as deprecated
	DeprecatedBy string `yaml:"deprecated-by,omitempty"` // Reference to replacement command
}

// IsInvokable returns whether the command is user-invokable.
// Defaults to true if the Invokable field is not set.
func (f *CommandFrontmatter) IsInvokable() bool {
	if f.Invokable == nil {
		return true // Default is invokable
	}
	return *f.Invokable
}

// Validate checks that the frontmatter has required fields and valid values.
func (f *CommandFrontmatter) Validate() error {
	if f.Name == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: name is required")
	}
	if f.Description == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: description is required")
	}

	// Category is required for non-invokable commands
	if !f.IsInvokable() && f.Category == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: category is required for non-invokable commands")
	}

	// Validate category value
	if f.Category != "" {
		validCategories := map[string]bool{
			"reference": true,
			"template":  true,
			"schema":    true,
		}
		if !validCategories[f.Category] {
			return errors.NewWithDetails(errors.CodeValidationFailed,
				"frontmatter: invalid category value",
				map[string]any{"category": f.Category, "valid": []string{"reference", "template", "schema"}})
		}
	}

	return nil
}

// ParseCommandFrontmatter extracts frontmatter from a command file.
// Returns error if frontmatter is missing or invalid.
func ParseCommandFrontmatter(content []byte) (*CommandFrontmatter, error) {
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

	var fm CommandFrontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &fm); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "invalid frontmatter YAML", err)
	}

	return &fm, nil
}

// BoolPtr is a helper function to create a pointer to a bool value.
// Useful for testing and creating CommandFrontmatter with explicit invokable values.
func BoolPtr(b bool) *bool {
	return &b
}
