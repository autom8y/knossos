// Package inscription provides the CLAUDE.md templating, ownership tracking,
// and synchronization system for the Knossos platform.
//
// The inscription system manages regions within CLAUDE.md with different ownership:
//   - knossos: Managed by Knossos templates, always synced
//   - satellite: Owned by satellite project, never overwritten
//   - regenerate: Generated from project state (ACTIVE_RITE, agents/)
package inscription

import (
	"time"
)

// OwnerType represents the ownership category for a CLAUDE.md region.
//
// NOTE: provenance.OwnerType is a distinct type with different semantics (file
// ownership: knossos/user/untracked). This type covers CLAUDE.md region ownership
// (knossos/satellite/regenerate). See TENSION-001 in .know/design-constraints.md.
type OwnerType string

const (
	// OwnerKnossos indicates regions managed by Knossos templates.
	// These are always overwritten on sync.
	OwnerKnossos OwnerType = "knossos"

	// OwnerSatellite indicates regions owned by the satellite project.
	// These are never overwritten during sync.
	OwnerSatellite OwnerType = "satellite"

	// OwnerRegenerate indicates regions generated from project state.
	// Content is regenerated from a source (e.g., ACTIVE_RITE, agents/*.md).
	OwnerRegenerate OwnerType = "regenerate"
)

// IsValid checks if the owner type is a known valid value.
func (o OwnerType) IsValid() bool {
	switch o {
	case OwnerKnossos, OwnerSatellite, OwnerRegenerate:
		return true
	default:
		return false
	}
}

// String returns the string representation of the owner type.
func (o OwnerType) String() string {
	return string(o)
}

// MarkerDirective represents the type of KNOSSOS marker directive.
type MarkerDirective string

const (
	// DirectiveStart begins a managed region.
	DirectiveStart MarkerDirective = "START"

	// DirectiveEnd closes a managed region.
	DirectiveEnd MarkerDirective = "END"

	// DirectiveAnchor marks a single-line insertion point.
	DirectiveAnchor MarkerDirective = "ANCHOR"
)

// IsValid checks if the directive is a known valid value.
func (d MarkerDirective) IsValid() bool {
	switch d {
	case DirectiveStart, DirectiveEnd, DirectiveAnchor:
		return true
	default:
		return false
	}
}

// RequiresEnd returns true if this directive type requires a matching END directive.
func (d MarkerDirective) RequiresEnd() bool {
	return d == DirectiveStart
}

// String returns the string representation of the directive.
func (d MarkerDirective) String() string {
	return string(d)
}

// Marker represents a parsed KNOSSOS marker from CLAUDE.md.
type Marker struct {
	// Directive is the operation type: START, END, or ANCHOR.
	Directive MarkerDirective

	// RegionName is the unique identifier for the region (kebab-case).
	RegionName string

	// Options contains key=value pairs from the marker.
	// Examples: regenerate=true, source=ACTIVE_RITE, owner=satellite
	Options map[string]string

	// LineNumber is the 1-indexed line number where the marker appears.
	LineNumber int

	// Raw is the original marker text as found in the file.
	Raw string
}

// GetOption returns the value of an option, or empty string if not present.
func (m *Marker) GetOption(key string) string {
	if m.Options == nil {
		return ""
	}
	return m.Options[key]
}

// HasOption returns true if the option key exists.
func (m *Marker) HasOption(key string) bool {
	if m.Options == nil {
		return false
	}
	_, ok := m.Options[key]
	return ok
}

// Region represents a managed region definition in KNOSSOS_MANIFEST.yaml.
type Region struct {
	// Owner determines sync behavior (knossos, satellite, regenerate).
	Owner OwnerType `yaml:"owner" json:"owner"`

	// Source specifies the data source for regenerate regions.
	// Example: "ACTIVE_RITE+agents", "agents/*.md"
	Source string `yaml:"source,omitempty" json:"source,omitempty"`

	// PreserveOnConflict indicates satellite edits are preserved on conflict.
	// Only applicable when Owner is regenerate.
	PreserveOnConflict bool `yaml:"preserve_on_conflict,omitempty" json:"preserve_on_conflict,omitempty"`

	// Hash is the SHA256 hash of the last synced content (hex string, 64 chars).
	Hash string `yaml:"hash,omitempty" json:"hash,omitempty"`

	// SyncedAt is the timestamp when this region was last synced.
	SyncedAt *time.Time `yaml:"synced_at,omitempty" json:"synced_at,omitempty"`
}

// Conditional represents a conditional inclusion rule for regions.
type Conditional struct {
	// When is the condition expression.
	// Examples: "file_exists('.sos/sessions')", "env_set(VAR)", "always", "never"
	When string `yaml:"when" json:"when"`

	// Include lists regions to include when condition is true.
	Include []string `yaml:"include,omitempty" json:"include,omitempty"`

	// Exclude lists regions to exclude when condition is true.
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// Manifest represents the KNOSSOS_MANIFEST.yaml configuration file.
type Manifest struct {
	// SchemaVersion is the manifest schema version (e.g., "1.0").
	SchemaVersion string `yaml:"schema_version" json:"schema_version"`

	// InscriptionVersion is incremented on each sync operation.
	InscriptionVersion string `yaml:"inscription_version" json:"inscription_version"`

	// LastSync is the timestamp of the last sync operation.
	LastSync *time.Time `yaml:"last_sync,omitempty" json:"last_sync,omitempty"`

	// ActiveRite is the current rite name.
	ActiveRite string `yaml:"active_rite,omitempty" json:"active_rite,omitempty"`

	// TemplatePath is the path to the master template file.
	// Default: "knossos/templates/CLAUDE.md.tpl"
	TemplatePath string `yaml:"template_path,omitempty" json:"template_path,omitempty"`

	// Regions contains region definitions keyed by region name.
	Regions map[string]*Region `yaml:"regions" json:"regions"`

	// SectionOrder is the ordered list of section identifiers.
	SectionOrder []string `yaml:"section_order,omitempty" json:"section_order,omitempty"`

	// Conditionals contains conditional inclusion rules keyed by rule name.
	Conditionals map[string]*Conditional `yaml:"conditionals,omitempty" json:"conditionals,omitempty"`
}

// GetRegion returns the region with the given name, or nil if not found.
func (m *Manifest) GetRegion(name string) *Region {
	if m.Regions == nil {
		return nil
	}
	return m.Regions[name]
}

// SetRegion sets a region in the manifest, creating the map if needed.
func (m *Manifest) SetRegion(name string, region *Region) {
	if m.Regions == nil {
		m.Regions = make(map[string]*Region)
	}
	m.Regions[name] = region
}

// HasRegion returns true if the region exists in the manifest.
func (m *Manifest) HasRegion(name string) bool {
	return m.GetRegion(name) != nil
}

// RegionNames returns all region names in the manifest.
func (m *Manifest) RegionNames() []string {
	if m.Regions == nil {
		return nil
	}
	names := make([]string, 0, len(m.Regions))
	for name := range m.Regions {
		names = append(names, name)
	}
	return names
}

// ParsedRegion represents a region extracted from CLAUDE.md.
type ParsedRegion struct {
	// Name is the region identifier.
	Name string

	// StartMarker is the opening marker.
	StartMarker *Marker

	// EndMarker is the closing marker (nil for ANCHOR directives).
	EndMarker *Marker

	// Content is the text between START and END markers.
	// Empty for ANCHOR directives.
	Content string

	// StartLine is the 1-indexed line number of the START marker.
	StartLine int

	// EndLine is the 1-indexed line number of the END marker (0 for ANCHOR).
	EndLine int
}

// IsAnchor returns true if this is an anchor region (single-line insertion point).
func (pr *ParsedRegion) IsAnchor() bool {
	return pr.StartMarker != nil && pr.StartMarker.Directive == DirectiveAnchor
}

// LineCount returns the number of lines in the region content.
func (pr *ParsedRegion) LineCount() int {
	if pr.Content == "" {
		return 0
	}
	count := 1
	for _, c := range pr.Content {
		if c == '\n' {
			count++
		}
	}
	return count
}

// ParseResult contains the result of parsing CLAUDE.md for KNOSSOS markers.
type ParseResult struct {
	// Regions contains all parsed regions indexed by name.
	Regions map[string]*ParsedRegion

	// Markers contains all markers in order of appearance.
	Markers []*Marker

	// Errors contains any parse errors encountered.
	Errors []ParseError

	// UnmanagedContent is content outside of any managed region.
	UnmanagedContent []UnmanagedSection
}

// GetRegion returns a parsed region by name.
func (pr *ParseResult) GetRegion(name string) *ParsedRegion {
	if pr.Regions == nil {
		return nil
	}
	return pr.Regions[name]
}

// HasErrors returns true if parsing encountered errors.
func (pr *ParseResult) HasErrors() bool {
	return len(pr.Errors) > 0
}

// ParseError represents an error encountered during marker parsing.
type ParseError struct {
	// Line is the 1-indexed line number where the error occurred.
	Line int

	// Column is the 1-indexed column number (0 if not applicable).
	Column int

	// Message describes the error.
	Message string

	// Raw is the problematic text.
	Raw string
}

// Error implements the error interface.
func (pe ParseError) Error() string {
	if pe.Column > 0 {
		return pe.Message
	}
	return pe.Message
}

// UnmanagedSection represents content outside of managed regions.
type UnmanagedSection struct {
	// StartLine is the 1-indexed starting line.
	StartLine int

	// EndLine is the 1-indexed ending line.
	EndLine int

	// Content is the text content.
	Content string
}

// DefaultTemplatePath is the default path to the master CLAUDE.md template.
const DefaultTemplatePath = "knossos/templates/CLAUDE.md.tpl"

// DefaultSchemaVersion is the current manifest schema version.
const DefaultSchemaVersion = "1.0"

// ManifestFileName is the name of the manifest file.
const ManifestFileName = "KNOSSOS_MANIFEST.yaml"

// DefaultManifestPath returns the default path to the manifest file.
// Path: .claude/KNOSSOS_MANIFEST.yaml
func DefaultManifestPath(projectRoot string) string {
	return projectRoot + "/.claude/" + ManifestFileName
}
