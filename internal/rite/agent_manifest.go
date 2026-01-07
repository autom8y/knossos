package rite

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"time"
)

// AgentManifestVersion is the current agent manifest schema version.
const AgentManifestVersion = "1.2"

// AgentManifest represents the AGENT_MANIFEST.json file.
type AgentManifest struct {
	Version     string                `json:"version"`
	GeneratedAt time.Time             `json:"generated_at"`
	ActiveRite  string                `json:"active_rite"`
	Agents      map[string]AgentEntry `json:"agents"`
	Orphans     []string              `json:"orphans"`
}

// AgentEntry tracks the source and state of an installed agent.
type AgentEntry struct {
	Source      string    `json:"source"` // "rite" or "project"
	Origin      string    `json:"origin,omitempty"`
	Checksum    string    `json:"checksum"`
	InstalledAt time.Time `json:"installed_at"`
	Orphaned    bool      `json:"orphaned,omitempty"`
}

// NewEmptyAgentManifest creates a new empty agent manifest.
func NewEmptyAgentManifest() *AgentManifest {
	return &AgentManifest{
		Version:     AgentManifestVersion,
		GeneratedAt: time.Now().UTC(),
		Agents:      make(map[string]AgentEntry),
		Orphans:     []string{},
	}
}

// LoadAgentManifest reads an agent manifest from the given path.
func LoadAgentManifest(path string) (*AgentManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewEmptyAgentManifest(), nil
		}
		return nil, err
	}

	var m AgentManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// Ensure maps are initialized
	if m.Agents == nil {
		m.Agents = make(map[string]AgentEntry)
	}
	if m.Orphans == nil {
		m.Orphans = []string{}
	}

	return &m, nil
}

// Save writes the manifest to the given path.
func (m *AgentManifest) Save(path string) error {
	m.GeneratedAt = time.Now().UTC()

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// DetectOrphans finds agents not belonging to the target rite.
func (m *AgentManifest) DetectOrphans(targetRite string) []string {
	var orphans []string
	for name, entry := range m.Agents {
		if entry.Source == "rite" && entry.Origin != targetRite {
			orphans = append(orphans, name)
		}
	}
	return orphans
}

// AddAgent adds or updates an agent entry.
func (m *AgentManifest) AddAgent(name string, source, riteName, checksum string) {
	m.Agents[name] = AgentEntry{
		Source:      source,
		Origin:      riteName,
		Checksum:    checksum,
		InstalledAt: time.Now().UTC(),
	}
}

// RemoveAgent removes an agent entry.
func (m *AgentManifest) RemoveAgent(name string) {
	delete(m.Agents, name)
}

// MarkOrphaned marks an agent as orphaned.
func (m *AgentManifest) MarkOrphaned(name string) {
	if entry, ok := m.Agents[name]; ok {
		entry.Orphaned = true
		m.Agents[name] = entry
		// Also add to orphans list
		found := false
		for _, o := range m.Orphans {
			if o == name {
				found = true
				break
			}
		}
		if !found {
			m.Orphans = append(m.Orphans, name)
		}
	}
}

// PromoteToProject changes an agent's source from team to project.
func (m *AgentManifest) PromoteToProject(name string) {
	if entry, ok := m.Agents[name]; ok {
		entry.Source = "project"
		entry.Origin = ""
		entry.Orphaned = false
		m.Agents[name] = entry
	}
	// Remove from orphans list
	var newOrphans []string
	for _, o := range m.Orphans {
		if o != name {
			newOrphans = append(newOrphans, o)
		}
	}
	m.Orphans = newOrphans
}

// ClearOrphans removes all orphan markers.
func (m *AgentManifest) ClearOrphans() {
	for name, entry := range m.Agents {
		if entry.Orphaned {
			entry.Orphaned = false
			m.Agents[name] = entry
		}
	}
	m.Orphans = []string{}
}

// SetActiveRite updates the active rite in the manifest.
func (m *AgentManifest) SetActiveRite(riteName string) {
	m.ActiveRite = riteName
}

// GetInstalledAgents returns the list of installed agent filenames.
func (m *AgentManifest) GetInstalledAgents() []string {
	agents := make([]string, 0, len(m.Agents))
	for name := range m.Agents {
		agents = append(agents, name)
	}
	return agents
}

// GetRiteAgents returns agents from a specific rite.
func (m *AgentManifest) GetRiteAgents(riteName string) []string {
	var agents []string
	for name, entry := range m.Agents {
		if entry.Source == "rite" && entry.Origin == riteName {
			agents = append(agents, name)
		}
	}
	return agents
}

// IsFromRite checks if an agent is from a specific rite.
func (m *AgentManifest) IsFromRite(agentName, riteName string) bool {
	entry, ok := m.Agents[agentName]
	if !ok {
		return false
	}
	return entry.Source == "rite" && entry.Origin == riteName
}

// ComputeChecksum calculates a SHA-256 checksum for a file.
func ComputeChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
