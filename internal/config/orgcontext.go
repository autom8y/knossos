package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// OrgContext provides org-level configuration for registry and discovery operations.
// It is an interface to support multi-org scenarios without modifying registry core.
type OrgContext interface {
	// Name returns the organization name.
	Name() string
	// RegistryDir returns the path where the org's registry catalog is persisted.
	// Location: $XDG_DATA_HOME/knossos/registry/{org}/
	RegistryDir() string
	// DataDir returns the org's XDG data directory.
	// Location: $XDG_DATA_HOME/knossos/orgs/{org}/
	DataDir() string
	// Repos returns configured repositories for this org.
	// Sourced from org.yaml; returns empty slice if no repos are configured.
	Repos() []RepoConfig
}

// RepoConfig holds configuration for a single repository within an org.
type RepoConfig struct {
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	DefaultBranch string `yaml:"default_branch"`
}

// orgYAML is the on-disk schema for org.yaml files.
// Only the repos field is consumed here; other fields are preserved on write by
// the org init command.
type orgYAML struct {
	Name  string       `yaml:"name"`
	Repos []RepoConfig `yaml:"repos,omitempty"`
}

// defaultOrgContext implements OrgContext for a named org.
type defaultOrgContext struct {
	name    string
	dataDir string
	repos   []RepoConfig
}

// DefaultOrgContext returns an OrgContext for the currently active org.
// Resolution order: KNOSSOS_ORG env var, then active-org config file.
// Returns an error if no active org is configured.
func DefaultOrgContext() (OrgContext, error) {
	orgName := ActiveOrg()
	if orgName == "" {
		return nil, fmt.Errorf("no active org configured; set KNOSSOS_ORG or run 'ari org set <name>'")
	}
	return NewOrgContext(orgName)
}

// NewOrgContext returns an OrgContext for the named org.
// Reads org.yaml from the org's DataDir to populate Repos().
// If org.yaml does not exist or has no repos field, Repos() returns nil — not an error.
func NewOrgContext(orgName string) (OrgContext, error) {
	if orgName == "" {
		return nil, fmt.Errorf("org name must not be empty")
	}

	dataDir := filepath.Join(XDGDataDir(), "orgs", orgName)
	repos := readReposFromOrgYAML(filepath.Join(dataDir, "org.yaml"))

	return &defaultOrgContext{
		name:    orgName,
		dataDir: dataDir,
		repos:   repos,
	}, nil
}

func (o *defaultOrgContext) Name() string {
	return o.name
}

func (o *defaultOrgContext) RegistryDir() string {
	return RegistryDir(o.name)
}

func (o *defaultOrgContext) DataDir() string {
	return o.dataDir
}

func (o *defaultOrgContext) Repos() []RepoConfig {
	return o.repos
}

// readReposFromOrgYAML reads the repos field from an org.yaml file.
// Returns nil (not an error) if the file is missing or has no repos field.
// This is backward-compatible: existing org.yaml files without repos are valid.
func readReposFromOrgYAML(path string) []RepoConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		// Missing org.yaml is normal for freshly-init'd orgs.
		return nil
	}

	var yml orgYAML
	if err := yaml.Unmarshal(data, &yml); err != nil {
		// Malformed org.yaml — degrade gracefully, return no repos.
		return nil
	}

	return yml.Repos
}
