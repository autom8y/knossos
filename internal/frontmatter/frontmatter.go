// Package frontmatter provides shared frontmatter parsing utilities used by
// both the agent and materialize packages. It extracts YAML frontmatter from
// markdown content delimited by "---\n...\n---\n" and provides the
// FlexibleStringSlice type for YAML fields that accept both comma-separated
// strings and proper YAML lists.
package frontmatter

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

// FlexibleStringSlice is a YAML type that accepts both a comma-separated string
// (e.g., "Bash, Read, Glob") and a proper YAML list (e.g., [Bash, Read, Glob]).
// This handles the common pattern in frontmatter where tools are listed inline.
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

// Parse extracts the YAML frontmatter block from markdown content.
// Frontmatter must be delimited by "---\n" at the start of content and
// a matching "\n---\n" closing delimiter. Returns the raw YAML bytes
// between the delimiters and the remainder of the content after the
// closing delimiter.
//
// Returns errMissingOpen if content does not start with "---\n" and
// errMissingClose if no closing delimiter is found.
func Parse(content []byte) (yamlBytes []byte, body []byte, err error) {
	// Check opening delimiter (handles both \n and \r\n)
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return nil, nil, ErrMissingOpenDelimiter
	}

	// Determine offset past opening delimiter
	openLen := 4 // len("---\n")
	if bytes.HasPrefix(content, []byte("---\r\n")) {
		openLen = 5
	}

	// Find closing delimiter
	rest := content[openLen:]
	closePatterns := [][]byte{
		[]byte("\n---\n"),
		[]byte("\n---\r\n"),
		[]byte("\r\n---\r\n"),
		[]byte("\r\n---\n"),
	}

	for _, pat := range closePatterns {
		if idx := bytes.Index(rest, pat); idx != -1 {
			yamlBytes = rest[:idx]
			body = rest[idx+len(pat):]
			return yamlBytes, body, nil
		}
	}

	return nil, nil, ErrMissingCloseDelimiter
}
