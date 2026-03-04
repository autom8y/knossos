// Package validation provides frontmatter extraction and JSON schema validation for Ariadne.
package validation

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"gopkg.in/yaml.v3"
)

// FrontmatterResult contains the extracted frontmatter and metadata.
type FrontmatterResult struct {
	// Data contains the parsed YAML frontmatter as a map.
	Data map[string]any

	// RawYAML contains the raw YAML content between delimiters.
	RawYAML string

	// StartLine is the line number where frontmatter starts (1-based).
	StartLine int

	// EndLine is the line number where frontmatter ends (1-based).
	EndLine int
}

// ExtractFrontmatter extracts YAML frontmatter from markdown content.
// Frontmatter must be delimited by "---" on its own line at the start of the file.
//
// Returns:
//   - FrontmatterResult with parsed data on success
//   - error with CodeSchemaInvalid for missing/malformed frontmatter
//   - error with CodeParseError for YAML parsing failures
func ExtractFrontmatter(content []byte) (*FrontmatterResult, error) {
	return ExtractFrontmatterFromReader(bytes.NewReader(content))
}

// ExtractFrontmatterFromReader extracts YAML frontmatter from a reader.
func ExtractFrontmatterFromReader(r io.Reader) (*FrontmatterResult, error) {
	scanner := bufio.NewScanner(r)

	// Read first line - must be "---"
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to read file", err)
		}
		return nil, errors.New(errors.CodeSchemaInvalid, "empty file: no frontmatter found")
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		return nil, errors.New(errors.CodeSchemaInvalid, "missing opening '---' delimiter on line 1")
	}

	// Collect YAML content until closing "---"
	var yamlLines []string
	lineNum := 1 // Already read line 1

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for closing delimiter
		if strings.TrimSpace(line) == "---" {
			// Found closing delimiter
			if len(yamlLines) == 0 {
				return nil, errors.New(errors.CodeSchemaInvalid, "empty frontmatter")
			}

			rawYAML := strings.Join(yamlLines, "\n")

			// Parse YAML
			var data map[string]any
			if err := yaml.Unmarshal([]byte(rawYAML), &data); err != nil {
				return nil, errors.Wrap(errors.CodeParseError, "invalid YAML in frontmatter", err)
			}

			// Handle nil result (empty but valid YAML like "---\n---")
			if data == nil {
				data = make(map[string]any)
			}

			return &FrontmatterResult{
				Data:      data,
				RawYAML:   rawYAML,
				StartLine: 1,
				EndLine:   lineNum,
			}, nil
		}

		yamlLines = append(yamlLines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "error reading file", err)
	}

	// Reached end of file without finding closing delimiter
	return nil, errors.New(errors.CodeSchemaInvalid, "unclosed frontmatter: missing closing '---' delimiter")
}

// HasFrontmatter checks if content starts with a frontmatter delimiter.
func HasFrontmatter(content []byte) bool {
	if len(content) < 4 {
		return false
	}

	// Check for "---" followed by newline
	if bytes.HasPrefix(content, []byte("---\n")) {
		return true
	}
	if bytes.HasPrefix(content, []byte("---\r\n")) {
		return true
	}

	return false
}

// BuildFrontmatter creates a markdown file header with YAML frontmatter.
func BuildFrontmatter(data map[string]any) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to marshal frontmatter", err)
	}
	return "---\n" + string(yamlBytes) + "---\n", nil
}
