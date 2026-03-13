// Package manifest provides manifest loading, validation, diffing, and merging for Ariadne.
// It handles project manifests and rite manifests.
package manifest

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
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
	Path    string         `json:"path"`
	Format  Format         `json:"format"`
	Content map[string]any `json:"content"`
	Raw     []byte         `json:"-"`
}

// Load reads and parses a manifest from the given path.
// Supports both filesystem paths and git refs (e.g., "HEAD:.channel/manifest.json").
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
				map[string]any{"path": path})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read manifest", err)
	}

	format := detectFormat(path)
	content, err := parse(data, format)
	if err != nil {
		return nil, errors.NewWithDetails(CodeParseError,
			"failed to parse manifest",
			map[string]any{
				"path":   path,
				"format": string(format),
				"cause":  err.Error(),
			})
	}

	m := &Manifest{
		Path:    path,
		Format:  format,
		Content: content,
		Raw:     data,
	}

	// Post-parse validation: if the content looks like a rite manifest
	// (has "name" and "agents" keys), run structural validation and log warnings.
	// Per TD-3: warnings only, never block Load().
	if looksLikeRiteManifest(content) {
		if warnings := ValidateRiteManifest(m); len(warnings) > 0 {
			for _, w := range warnings {
				slog.Warn("manifest validation issue", "path", path, "message", w.Message, "location", w.Path)
			}
		}
	}

	return m, nil
}

// LoadFromGitRef loads a manifest from a git reference.
// Ref format: "commit:path" (e.g., "HEAD:.channel/manifest.json")
func LoadFromGitRef(ref string) (*Manifest, error) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return nil, errors.NewWithDetails(errors.CodeUsageError,
			"invalid git ref format",
			map[string]any{"ref": ref, "expected": "commit:path"})
	}

	commit, path := parts[0], parts[1]

	// Execute git show
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "show", commit+":"+path)
	data, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if stderrors.As(err, &exitErr) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"git ref not found",
				map[string]any{
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
			map[string]any{
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
func parse(data []byte, format Format) (map[string]any, error) {
	var content map[string]any

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

	return fileutil.AtomicWriteFile(path, data, 0644)
}

// ToJSON converts manifest content to JSON bytes.
func (m *Manifest) ToJSON() ([]byte, error) {
	return json.Marshal(m.Content)
}

// Clone creates a deep copy of the manifest.
func (m *Manifest) Clone() *Manifest {
	// Use JSON round-trip for deep copy
	data, _ := json.Marshal(m.Content)
	var content map[string]any
	_ = json.Unmarshal(data, &content)

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
func (m *Manifest) Get(path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = m.Content

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
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
	if arr, ok := val.([]any); ok {
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

// looksLikeRiteManifest returns true if the content has keys typical of a rite manifest.
// Used to decide whether to run rite-specific validation from Load().
func looksLikeRiteManifest(content map[string]any) bool {
	_, hasName := content["name"]
	_, hasAgents := content["agents"]
	return hasName && hasAgents
}

// ValidateRiteManifest performs structural validation on a manifest that represents
// a rite manifest. It checks required fields and agent reference format.
//
// Per TD-3: returns warnings (not errors) so that existing manifests with minor
// schema drift do not break ari sync. Callers should log warnings, not abort.
func ValidateRiteManifest(m *Manifest) []ValidationIssue {
	if m == nil || m.Content == nil {
		return nil
	}
	var warnings []ValidationIssue

	// Required field: name
	name, hasName := m.Content["name"]
	if !hasName {
		warnings = append(warnings, ValidationIssue{
			Path:     "$.name",
			Message:  "missing required field 'name'",
			Severity: "warning",
		})
	} else if nameStr, ok := name.(string); ok && nameStr == "" {
		warnings = append(warnings, ValidationIssue{
			Path:     "$.name",
			Message:  "field 'name' must not be empty",
			Severity: "warning",
		})
	}

	// Required field: entry_agent
	entryAgent, hasEntryAgent := m.Content["entry_agent"]
	if !hasEntryAgent {
		warnings = append(warnings, ValidationIssue{
			Path:     "$.entry_agent",
			Message:  "missing required field 'entry_agent'",
			Severity: "warning",
		})
	} else if ea, ok := entryAgent.(string); ok && ea == "" {
		warnings = append(warnings, ValidationIssue{
			Path:     "$.entry_agent",
			Message:  "field 'entry_agent' must not be empty",
			Severity: "warning",
		})
	}

	// Validate agent references: each agent must have a non-empty name
	if agents, ok := m.Content["agents"]; ok {
		if agentList, ok := agents.([]any); ok {
			for i, agent := range agentList {
				if agentMap, ok := agent.(map[string]any); ok {
					agentName, hasAgentName := agentMap["name"]
					if !hasAgentName {
						warnings = append(warnings, ValidationIssue{
							Path:     formatAgentPath(i, "name"),
							Message:  "agent entry missing required field 'name'",
							Severity: "warning",
						})
					} else if n, ok := agentName.(string); ok && n == "" {
						warnings = append(warnings, ValidationIssue{
							Path:     formatAgentPath(i, "name"),
							Message:  "agent 'name' must not be empty",
							Severity: "warning",
						})
					}
				}
			}
		}
	}

	// Validate entry_agent references an agent in the agents list
	if hasEntryAgent && hasName {
		if ea, ok := entryAgent.(string); ok && ea != "" {
			if !agentExistsInList(m.Content, ea) {
				warnings = append(warnings, ValidationIssue{
					Path:     "$.entry_agent",
					Message:  "entry_agent '" + ea + "' not found in agents list",
					Severity: "warning",
				})
			}
		}
	}

	return warnings
}

// formatAgentPath returns a JSON path for an agent list entry.
func formatAgentPath(index int, field string) string {
	return fmt.Sprintf("$.agents[%d].%s", index, field)
}

// agentExistsInList checks if an agent name appears in the manifest's agents list.
func agentExistsInList(content map[string]any, name string) bool {
	agents, ok := content["agents"]
	if !ok {
		return false
	}
	agentList, ok := agents.([]any)
	if !ok {
		return false
	}
	for _, agent := range agentList {
		if agentMap, ok := agent.(map[string]any); ok {
			if n, ok := agentMap["name"].(string); ok && n == name {
				return true
			}
		}
	}
	return false
}
