// Package manifest provides manifest loading, validation, diffing, and merging for Ariadne.
// It handles Claude Extension Manifests (CEM) and team pack manifests.
package manifest

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/autom8y/ariadne/internal/errors"
	"gopkg.in/yaml.v3"
)

// Format represents the manifest file format.
type Format string

const (
	// FormatJSON is JSON format.
	FormatJSON Format = "json"
	// FormatYAML is YAML format.
	FormatYAML Format = "yaml"
)

// Manifest represents a parsed manifest file.
type Manifest struct {
	Path    string                 `json:"path"`
	Format  Format                 `json:"format"`
	Content map[string]interface{} `json:"content"`
	Raw     []byte                 `json:"-"`
}

// Load reads and parses a manifest from the given path.
// Supports both filesystem paths and git refs (e.g., "HEAD:.claude/manifest.json").
func Load(path string) (*Manifest, error) {
	// Check if this is a git ref
	if isGitRef(path) {
		return LoadFromGitRef(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"manifest file not found",
				map[string]interface{}{"path": path})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read manifest", err)
	}

	format := detectFormat(path)
	content, err := parse(data, format)
	if err != nil {
		return nil, errors.NewWithDetails(CodeParseError,
			"failed to parse manifest",
			map[string]interface{}{
				"path":   path,
				"format": string(format),
				"cause":  err.Error(),
			})
	}

	return &Manifest{
		Path:    path,
		Format:  format,
		Content: content,
		Raw:     data,
	}, nil
}

// LoadFromGitRef loads a manifest from a git reference.
// Ref format: "commit:path" (e.g., "HEAD:.claude/manifest.json")
func LoadFromGitRef(ref string) (*Manifest, error) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return nil, errors.NewWithDetails(errors.CodeUsageError,
			"invalid git ref format",
			map[string]interface{}{"ref": ref, "expected": "commit:path"})
	}

	commit, path := parts[0], parts[1]

	// Execute git show
	cmd := exec.Command("git", "show", commit+":"+path)
	data, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"git ref not found",
				map[string]interface{}{
					"ref":    ref,
					"stderr": string(exitErr.Stderr),
				})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "git show failed", err)
	}

	format := detectFormat(path)
	content, err := parse(data, format)
	if err != nil {
		return nil, errors.NewWithDetails(CodeParseError,
			"failed to parse manifest from git",
			map[string]interface{}{
				"ref":    ref,
				"format": string(format),
				"cause":  err.Error(),
			})
	}

	return &Manifest{
		Path:    ref,
		Format:  format,
		Content: content,
		Raw:     data,
	}, nil
}

// isGitRef checks if a path looks like a git reference.
func isGitRef(path string) bool {
	// Git refs contain a colon and typically start with HEAD, origin, a commit hash, etc.
	if !strings.Contains(path, ":") {
		return false
	}
	// Check if the part before : looks like a git ref (not a Windows drive letter)
	parts := strings.SplitN(path, ":", 2)
	if len(parts) != 2 {
		return false
	}
	// Single letter before colon is likely a Windows path
	if len(parts[0]) == 1 && strings.ContainsAny(parts[0], "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") {
		return false
	}
	return true
}

// detectFormat determines format from file extension.
func detectFormat(path string) Format {
	// For git refs, extract the actual path
	if idx := strings.LastIndex(path, ":"); idx != -1 {
		path = path[idx+1:]
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return FormatYAML
	default:
		return FormatJSON
	}
}

// parse decodes the manifest content.
func parse(data []byte, format Format) (map[string]interface{}, error) {
	var content map[string]interface{}

	switch format {
	case FormatYAML:
		if err := yaml.Unmarshal(data, &content); err != nil {
			return nil, err
		}
	default:
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, err
		}
	}

	return content, nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save(path string) error {
	var data []byte
	var err error

	format := detectFormat(path)
	switch format {
	case FormatYAML:
		data, err = yaml.Marshal(m.Content)
	default:
		data, err = json.MarshalIndent(m.Content, "", "  ")
		if err == nil {
			// Add trailing newline for JSON
			data = append(data, '\n')
		}
	}

	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to encode manifest", err)
	}

	return os.WriteFile(path, data, 0644)
}

// ToJSON converts manifest content to JSON bytes.
func (m *Manifest) ToJSON() ([]byte, error) {
	return json.Marshal(m.Content)
}

// Clone creates a deep copy of the manifest.
func (m *Manifest) Clone() *Manifest {
	// Use JSON round-trip for deep copy
	data, _ := json.Marshal(m.Content)
	var content map[string]interface{}
	json.Unmarshal(data, &content)

	rawCopy := make([]byte, len(m.Raw))
	copy(rawCopy, m.Raw)

	return &Manifest{
		Path:    m.Path,
		Format:  m.Format,
		Content: content,
		Raw:     rawCopy,
	}
}

// Get retrieves a value from the manifest using dot notation (e.g., "teams.default").
func (m *Manifest) Get(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = m.Content

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// GetString retrieves a string value from the manifest.
func (m *Manifest) GetString(path string) string {
	val, ok := m.Get(path)
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetStringSlice retrieves a string slice from the manifest.
func (m *Manifest) GetStringSlice(path string) []string {
	val, ok := m.Get(path)
	if !ok {
		return nil
	}
	if arr, ok := val.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// Exists checks if the manifest was loaded from an existing file.
func (m *Manifest) Exists() bool {
	return m != nil && len(m.Raw) > 0
}
