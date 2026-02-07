package inscription

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"gopkg.in/yaml.v3"
)

// ManifestLoader handles loading and saving KNOSSOS_MANIFEST.yaml files.
type ManifestLoader struct {
	// ProjectRoot is the root directory of the project.
	ProjectRoot string

	// ManifestPath is the full path to the manifest file.
	ManifestPath string
}

// NewManifestLoader creates a new manifest loader for the given project root.
func NewManifestLoader(projectRoot string) *ManifestLoader {
	return &ManifestLoader{
		ProjectRoot:  projectRoot,
		ManifestPath: DefaultManifestPath(projectRoot),
	}
}

// Load reads and parses the KNOSSOS_MANIFEST.yaml file.
// Returns an error if the file doesn't exist or is invalid.
func (m *ManifestLoader) Load() (*Manifest, error) {
	data, err := os.ReadFile(m.ManifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"manifest file not found",
				map[string]interface{}{"path": m.ManifestPath})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read manifest", err)
	}

	return m.ParseManifest(data)
}

// LoadOrCreate loads the manifest if it exists, or creates a default one.
func (m *ManifestLoader) LoadOrCreate() (*Manifest, error) {
	manifest, err := m.Load()
	if err != nil {
		if errors.IsNotFound(err) {
			return m.CreateDefault()
		}
		return nil, err
	}
	return manifest, nil
}

// ParseManifest parses manifest content from YAML bytes.
func (m *ManifestLoader) ParseManifest(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, errors.NewWithDetails(errors.CodeParseError,
			"failed to parse manifest YAML",
			map[string]interface{}{
				"path":  m.ManifestPath,
				"cause": err.Error(),
			})
	}

	// Validate required fields
	if err := m.ValidateManifest(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// ValidateManifest validates the manifest structure and content.
func (m *ManifestLoader) ValidateManifest(manifest *Manifest) error {
	var issues []string

	// Required fields
	if manifest.SchemaVersion == "" {
		issues = append(issues, "missing required field: schema_version")
	}

	if manifest.InscriptionVersion == "" {
		issues = append(issues, "missing required field: inscription_version")
	}

	if manifest.Regions == nil {
		issues = append(issues, "missing required field: regions")
	}

	// Validate schema version format
	if manifest.SchemaVersion != "" {
		if !isValidSchemaVersion(manifest.SchemaVersion) {
			issues = append(issues, "invalid schema_version format: "+manifest.SchemaVersion+" (expected N.N)")
		}
	}

	// Validate inscription version format (numeric string)
	if manifest.InscriptionVersion != "" {
		if _, err := strconv.Atoi(manifest.InscriptionVersion); err != nil {
			issues = append(issues, "invalid inscription_version format: "+manifest.InscriptionVersion+" (expected numeric)")
		}
	}

	// Validate regions
	for name, region := range manifest.Regions {
		// Validate region name
		if err := ValidateRegionName(name); err != nil {
			issues = append(issues, "invalid region name '"+name+"': "+err.Error())
		}

		// Validate owner
		if region.Owner == "" {
			issues = append(issues, "region '"+name+"' missing required field: owner")
		} else if !region.Owner.IsValid() {
			issues = append(issues, "region '"+name+"' has invalid owner: "+string(region.Owner))
		}

		// Validate source requirement for regenerate regions
		if region.Owner == OwnerRegenerate && region.Source == "" {
			issues = append(issues, "region '"+name+"' with owner 'regenerate' requires source field")
		}

		// Validate hash format if present
		if region.Hash != "" && len(region.Hash) != 64 {
			issues = append(issues, "region '"+name+"' has invalid hash format (expected 64 hex characters)")
		}
	}

	// Validate section_order references existing regions
	for _, section := range manifest.SectionOrder {
		if err := ValidateRegionName(section); err != nil {
			issues = append(issues, "invalid section name in section_order: "+section)
		}
		// Note: We don't require all section_order items to exist in regions
		// because sections may be conditionally included/excluded
	}

	// Validate conditionals
	for name, cond := range manifest.Conditionals {
		if cond.When == "" {
			issues = append(issues, "conditional '"+name+"' missing required field: when")
		}
	}

	if len(issues) > 0 {
		return errors.ErrSchemaInvalid(m.ManifestPath, issues)
	}

	return nil
}

// Save writes the manifest to disk.
func (m *ManifestLoader) Save(manifest *Manifest) error {
	// Ensure directory exists
	dir := filepath.Dir(m.ManifestPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to create manifest directory", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal manifest", err)
	}

	// Write atomically via temp file
	tmpPath := m.ManifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write manifest temp file", err)
	}

	if err := os.Rename(tmpPath, m.ManifestPath); err != nil {
		os.Remove(tmpPath) // Clean up
		return errors.Wrap(errors.CodeGeneralError, "failed to rename manifest file", err)
	}

	return nil
}

// CreateDefault creates a default manifest for a new project.
func (m *ManifestLoader) CreateDefault() (*Manifest, error) {
	now := time.Now().UTC()
	manifest := &Manifest{
		SchemaVersion:      DefaultSchemaVersion,
		InscriptionVersion: "1",
		LastSync:           &now,
		TemplatePath:       DefaultTemplatePath,
		Regions:            make(map[string]*Region),
		SectionOrder:       DefaultSectionOrder(),
		Conditionals:       make(map[string]*Conditional),
	}

	// Add default knossos-owned regions
	defaultKnossosRegions := []string{
		"execution-mode",
		"agent-routing",
		"commands",
		"platform-infrastructure",
	}

	for _, name := range defaultKnossosRegions {
		manifest.Regions[name] = &Region{
			Owner: OwnerKnossos,
		}
	}

	// Add default regenerate regions
	manifest.Regions["quick-start"] = &Region{
		Owner:  OwnerRegenerate,
		Source: "ACTIVE_RITE+agents",
	}

	manifest.Regions["agent-configurations"] = &Region{
		Owner:  OwnerRegenerate,
		Source: "agents/*.md",
	}

	// Add user-content region (satellite-owned, user can freely edit)
	manifest.Regions["user-content"] = &Region{
		Owner: OwnerSatellite,
	}

	return manifest, nil
}

// DefaultSectionOrder returns the default section order.
// When removing a section from this list, add its name to DeprecatedRegions().
func DefaultSectionOrder() []string {
	return []string{
		// Core behavior (determines agent mode)
		"execution-mode",

		// Team context (who is available)
		"quick-start",
		"agent-routing",
		"commands",
		"agent-configurations",

		// Infrastructure pointer (how to access platform tools)
		"platform-infrastructure",

		// User customization (edit freely)
		"user-content",
	}
}

// DeprecatedRegions returns region names that were previously part of the
// default section order but have been removed. These are dropped during merge
// instead of being adopted as satellite content.
func DeprecatedRegions() []string {
	return []string{
		"slash-commands", // Removed in v18, absorbed into commands Rosetta Stone
		"navigation",    // Removed in v20, zero unique value (all content duplicated or tautological)
	}
}

// IncrementVersion increments the inscription version and updates last_sync.
func (m *ManifestLoader) IncrementVersion(manifest *Manifest) {
	version, _ := strconv.Atoi(manifest.InscriptionVersion)
	manifest.InscriptionVersion = strconv.Itoa(version + 1)

	now := time.Now().UTC()
	manifest.LastSync = &now
}

// UpdateRegionHash updates a region's hash with the SHA256 of the content.
func (m *ManifestLoader) UpdateRegionHash(manifest *Manifest, regionName string, content string) {
	region := manifest.GetRegion(regionName)
	if region == nil {
		return
	}

	hash := ComputeContentHash(content)
	region.Hash = hash

	now := time.Now().UTC()
	region.SyncedAt = &now
}

// ComputeContentHash computes the SHA256 hash of content as a hex string.
func ComputeContentHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// Exists checks if the manifest file exists.
func (m *ManifestLoader) Exists() bool {
	_, err := os.Stat(m.ManifestPath)
	return err == nil
}

// ToJSON converts the manifest to JSON for schema validation.
func (manifest *Manifest) ToJSON() ([]byte, error) {
	return json.Marshal(manifest)
}

// Clone creates a deep copy of the manifest.
// Returns an error if the manifest cannot be serialized (should not happen in normal operation).
func (manifest *Manifest) Clone() (*Manifest, error) {
	// Use JSON round-trip for deep copy
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, errors.NewWithDetails(errors.CodeGeneralError,
			"clone: failed to marshal manifest",
			map[string]interface{}{"error": err.Error()})
	}
	var clone Manifest
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, errors.NewWithDetails(errors.CodeGeneralError,
			"clone: failed to unmarshal manifest",
			map[string]interface{}{"error": err.Error()})
	}
	return &clone, nil
}

// isValidSchemaVersion checks if a version string matches N.N format.
func isValidSchemaVersion(version string) bool {
	// Simple validation: should be like "1.0" or "2.1"
	parts := 0
	hasDigit := false
	for _, c := range version {
		if c == '.' {
			if !hasDigit {
				return false
			}
			parts++
			hasDigit = false
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		} else {
			return false
		}
	}
	return parts == 1 && hasDigit
}

// MergeManifests merges two manifests, preferring values from the overlay.
// Regions are merged by name, with overlay taking precedence.
func MergeManifests(base, overlay *Manifest) (*Manifest, error) {
	if base == nil {
		return overlay.Clone()
	}
	if overlay == nil {
		return base.Clone()
	}

	result, err := base.Clone()
	if err != nil {
		return nil, err
	}

	// Overlay scalar fields
	if overlay.SchemaVersion != "" {
		result.SchemaVersion = overlay.SchemaVersion
	}
	if overlay.InscriptionVersion != "" {
		result.InscriptionVersion = overlay.InscriptionVersion
	}
	if overlay.LastSync != nil {
		result.LastSync = overlay.LastSync
	}
	if overlay.ActiveRite != "" {
		result.ActiveRite = overlay.ActiveRite
	}
	if overlay.TemplatePath != "" {
		result.TemplatePath = overlay.TemplatePath
	}

	// Merge regions
	if overlay.Regions != nil {
		if result.Regions == nil {
			result.Regions = make(map[string]*Region)
		}
		for name, region := range overlay.Regions {
			result.Regions[name] = region
		}
	}

	// Overlay section_order if provided
	if len(overlay.SectionOrder) > 0 {
		result.SectionOrder = overlay.SectionOrder
	}

	// Merge conditionals
	if overlay.Conditionals != nil {
		if result.Conditionals == nil {
			result.Conditionals = make(map[string]*Conditional)
		}
		for name, cond := range overlay.Conditionals {
			result.Conditionals[name] = cond
		}
	}

	return result, nil
}

// GetKnossosRegions returns all regions owned by knossos.
func (manifest *Manifest) GetKnossosRegions() map[string]*Region {
	result := make(map[string]*Region)
	for name, region := range manifest.Regions {
		if region.Owner == OwnerKnossos {
			result[name] = region
		}
	}
	return result
}

// GetSatelliteRegions returns all regions owned by satellite.
func (manifest *Manifest) GetSatelliteRegions() map[string]*Region {
	result := make(map[string]*Region)
	for name, region := range manifest.Regions {
		if region.Owner == OwnerSatellite {
			result[name] = region
		}
	}
	return result
}

// GetRegenerateRegions returns all regions that are regenerated.
func (manifest *Manifest) GetRegenerateRegions() map[string]*Region {
	result := make(map[string]*Region)
	for name, region := range manifest.Regions {
		if region.Owner == OwnerRegenerate {
			result[name] = region
		}
	}
	return result
}

// AddRegion adds a new region to the manifest.
// Returns an error if the region already exists.
func (manifest *Manifest) AddRegion(name string, region *Region) error {
	if err := ValidateRegionName(name); err != nil {
		return err
	}

	if manifest.HasRegion(name) {
		return errors.NewWithDetails(errors.CodeUsageError,
			"region already exists",
			map[string]interface{}{"region": name})
	}

	manifest.SetRegion(name, region)
	return nil
}

// RemoveRegion removes a region from the manifest.
func (manifest *Manifest) RemoveRegion(name string) {
	delete(manifest.Regions, name)

	// Also remove from section_order
	newOrder := make([]string, 0, len(manifest.SectionOrder))
	for _, section := range manifest.SectionOrder {
		if section != name {
			newOrder = append(newOrder, section)
		}
	}
	manifest.SectionOrder = newOrder
}

// SetActiveRite updates the active rite and returns the old value.
func (manifest *Manifest) SetActiveRite(rite string) string {
	old := manifest.ActiveRite
	manifest.ActiveRite = rite
	return old
}

// ContentChanged checks if content has changed by comparing hashes.
func (region *Region) ContentChanged(content string) bool {
	if region.Hash == "" {
		return true
	}
	return region.Hash != ComputeContentHash(content)
}
