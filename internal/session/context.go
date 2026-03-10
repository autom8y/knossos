package session

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/validation"
	"gopkg.in/yaml.v3"
)

// Strand tracks a child session created via fray.
type Strand struct {
	SessionID string `yaml:"session_id" json:"session_id"`
	Status    string `yaml:"status" json:"status"`
	FrameRef  string `yaml:"frame_ref,omitempty" json:"frame_ref,omitempty"`
	LandedAt  string `yaml:"landed_at,omitempty" json:"landed_at,omitempty"`
}

// Context represents parsed SESSION_CONTEXT.md.
type Context struct {
	SchemaVersion string    `yaml:"schema_version" json:"schema_version"`
	SessionID     string    `yaml:"session_id" json:"session_id"`
	Status        Status    `yaml:"status" json:"status"`
	CreatedAt     time.Time `yaml:"created_at" json:"created_at"`
	Initiative    string    `yaml:"initiative" json:"initiative"`
	Complexity    string    `yaml:"complexity" json:"complexity"`
	ActiveRite    string    `yaml:"active_rite" json:"active_rite"`
	Rite          *string   `yaml:"rite" json:"rite,omitempty"` // null for cross-cutting
	CurrentPhase  string    `yaml:"current_phase" json:"current_phase"`

	// Optional fields
	ParkedAt     *time.Time `yaml:"parked_at,omitempty" json:"parked_at,omitempty"`
	ParkedReason string     `yaml:"parked_reason,omitempty" json:"parked_reason,omitempty"`
	ArchivedAt   *time.Time `yaml:"archived_at,omitempty" json:"archived_at,omitempty"`
	ResumedAt    *time.Time `yaml:"resumed_at,omitempty" json:"resumed_at,omitempty"`

	// Fray fields (session forking)
	FrayedFrom string   `yaml:"frayed_from,omitempty" json:"frayed_from,omitempty"`
	FrayPoint  string   `yaml:"fray_point,omitempty" json:"fray_point,omitempty"` // Phase at fork
	Strands    []Strand `yaml:"strands,omitempty" json:"strands,omitempty"`       // Child sessions

	// v2.3 fields
	FrameRef   string `yaml:"frame_ref,omitempty" json:"frame_ref,omitempty"`
	ClaimedBy  string `yaml:"claimed_by,omitempty" json:"claimed_by,omitempty"`
	ParkSource string `yaml:"park_source,omitempty" json:"park_source,omitempty"`

	// Procession tracks an active cross-rite coordinated workflow.
	// nil when no procession is active (most sessions). Added in v2.3 (additive, omitempty).
	Procession *Procession `yaml:"procession,omitempty" json:"procession,omitempty"`

	// Raw markdown body (after frontmatter)
	Body string `yaml:"-" json:"-"`
}

// FindStrand returns a mutable pointer to the strand with the given session ID,
// or nil if not found.
func (c *Context) FindStrand(id string) *Strand {
	for i := range c.Strands {
		if c.Strands[i].SessionID == id {
			return &c.Strands[i]
		}
	}
	return nil
}

// strandList handles polymorphic YAML deserialization for Strands.
// Accepts either []string (v2.1/2.2) or []Strand (v2.3+).
type strandList []Strand

func (s *strandList) UnmarshalYAML(value *yaml.Node) error {
	// Try []Strand first (new format)
	var strands []Strand
	if err := value.Decode(&strands); err == nil {
		if len(strands) == 0 || strands[0].SessionID != "" {
			*s = strands
			return nil
		}
	}
	// Fall back to []string (old format)
	var ids []string
	if err := value.Decode(&ids); err != nil {
		return err
	}
	result := make([]Strand, len(ids))
	for i, id := range ids {
		result[i] = Strand{SessionID: id, Status: "ACTIVE"}
	}
	*s = result
	return nil
}

// CompletedStation records the completion of a procession station.
// Timestamps are stored as strings (RFC3339) rather than time.Time because
// completed station timestamps are informational/audit-only and never used for
// time arithmetic. This matches how Strand.LandedAt is stored.
type CompletedStation struct {
	Station     string   `yaml:"station"              json:"station"`
	Rite        string   `yaml:"rite"                 json:"rite"`
	CompletedAt string   `yaml:"completed_at"         json:"completed_at"`
	Artifacts   []string `yaml:"artifacts,omitempty"  json:"artifacts,omitempty"`
}

// Procession tracks the state of a cross-rite coordinated workflow within a session.
// Added in schema v2.3 (additive, omitempty). nil when no procession is active.
type Procession struct {
	ID                string             `yaml:"id"                            json:"id"`
	Type              string             `yaml:"type"                          json:"type"`
	CurrentStation    string             `yaml:"current_station"               json:"current_station"`
	CompletedStations []CompletedStation `yaml:"completed_stations,omitempty"  json:"completed_stations,omitempty"`
	NextStation       string             `yaml:"next_station,omitempty"        json:"next_station,omitempty"`
	NextRite          string             `yaml:"next_rite,omitempty"           json:"next_rite,omitempty"`
	ArtifactDir       string             `yaml:"artifact_dir"                  json:"artifact_dir"`
}

// contextYAML is the YAML representation with string timestamps.
type contextYAML struct {
	SchemaVersion string     `yaml:"schema_version"`
	SessionID     string     `yaml:"session_id"`
	Status        string     `yaml:"status"`
	CreatedAt     string     `yaml:"created_at"`
	Initiative    string     `yaml:"initiative"`
	Complexity    string     `yaml:"complexity"`
	ActiveRite    string     `yaml:"active_rite"`
	Rite          *string    `yaml:"rite,omitempty"`
	CurrentPhase  string     `yaml:"current_phase"`
	ParkedAt      string     `yaml:"parked_at,omitempty"`
	ParkedReason  string     `yaml:"parked_reason,omitempty"`
	ArchivedAt    string     `yaml:"archived_at,omitempty"`
	ResumedAt     string     `yaml:"resumed_at,omitempty"`
	FrayedFrom    string     `yaml:"frayed_from,omitempty"`
	FrayPoint     string     `yaml:"fray_point,omitempty"`
	Strands       strandList `yaml:"strands,omitempty"`
	FrameRef      string      `yaml:"frame_ref,omitempty"`
	ClaimedBy     string      `yaml:"claimed_by,omitempty"`
	ParkSource    string      `yaml:"park_source,omitempty"`
	Procession    *Procession `yaml:"procession,omitempty"`
}

// ParseContext parses SESSION_CONTEXT.md content.
func ParseContext(content []byte) (*Context, error) {
	str := string(content)

	// Find frontmatter
	if !strings.HasPrefix(str, "---\n") && !strings.HasPrefix(str, "---\r\n") {
		return nil, errors.New(errors.CodeSchemaInvalid, "no YAML frontmatter found")
	}

	// Find closing delimiter
	endIdx := strings.Index(str[4:], "\n---")
	if endIdx == -1 {
		endIdx = strings.Index(str[4:], "\r\n---")
	}
	if endIdx == -1 {
		return nil, errors.New(errors.CodeSchemaInvalid, "unclosed YAML frontmatter")
	}

	yamlContent := str[4 : endIdx+4]
	body := ""
	afterFrontmatter := endIdx + 4 + 4 // Skip past "\n---"
	if len(str) > afterFrontmatter {
		body = str[afterFrontmatter:]
	}

	// Parse YAML
	var yamlData contextYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid YAML frontmatter", err)
	}

	// Convert to Context
	ctx := &Context{
		SchemaVersion: yamlData.SchemaVersion,
		SessionID:     yamlData.SessionID,
		Status:        NormalizeStatus(yamlData.Status),
		Initiative:    yamlData.Initiative,
		Complexity:    yamlData.Complexity,
		ActiveRite:    yamlData.ActiveRite,
		Rite:          yamlData.Rite,
		CurrentPhase:  yamlData.CurrentPhase,
		ParkedReason:  yamlData.ParkedReason,
		Body:          body,
	}

	// Parse timestamps
	if yamlData.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339, yamlData.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid created_at timestamp", err)
		}
		ctx.CreatedAt = t
	}

	if yamlData.ParkedAt != "" {
		t, err := time.Parse(time.RFC3339, yamlData.ParkedAt)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid parked_at timestamp", err)
		}
		ctx.ParkedAt = &t
	}

	if yamlData.ArchivedAt != "" {
		t, err := time.Parse(time.RFC3339, yamlData.ArchivedAt)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid archived_at timestamp", err)
		}
		ctx.ArchivedAt = &t
	}

	if yamlData.ResumedAt != "" {
		t, err := time.Parse(time.RFC3339, yamlData.ResumedAt)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid resumed_at timestamp", err)
		}
		ctx.ResumedAt = &t
	}

	ctx.FrayedFrom = yamlData.FrayedFrom
	ctx.FrayPoint = yamlData.FrayPoint
	ctx.Strands = []Strand(yamlData.Strands)
	ctx.FrameRef = yamlData.FrameRef
	ctx.ClaimedBy = yamlData.ClaimedBy
	ctx.ParkSource = yamlData.ParkSource
	ctx.Procession = yamlData.Procession

	return ctx, nil
}

// LoadContext loads a session context from a file.
func LoadContext(path string) (*Context, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.CodeFileNotFound, "session context file not found")
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read session context", err)
	}
	return ParseContext(content)
}

// Serialize converts the context back to SESSION_CONTEXT.md format.
func (c *Context) Serialize() ([]byte, error) {
	// Build YAML data
	yamlData := contextYAML{
		SchemaVersion: c.SchemaVersion,
		SessionID:     c.SessionID,
		Status:        string(c.Status),
		CreatedAt:     c.CreatedAt.UTC().Format(time.RFC3339),
		Initiative:    c.Initiative,
		Complexity:    c.Complexity,
		ActiveRite:    c.ActiveRite,
		Rite:          c.Rite,
		CurrentPhase:  c.CurrentPhase,
		ParkedReason:  c.ParkedReason,
	}

	if c.ParkedAt != nil {
		yamlData.ParkedAt = c.ParkedAt.UTC().Format(time.RFC3339)
	}
	if c.ArchivedAt != nil {
		yamlData.ArchivedAt = c.ArchivedAt.UTC().Format(time.RFC3339)
	}
	if c.ResumedAt != nil {
		yamlData.ResumedAt = c.ResumedAt.UTC().Format(time.RFC3339)
	}

	yamlData.FrayedFrom = c.FrayedFrom
	yamlData.FrayPoint = c.FrayPoint
	yamlData.Strands = strandList(c.Strands)
	yamlData.FrameRef = c.FrameRef
	yamlData.ClaimedBy = c.ClaimedBy
	yamlData.ParkSource = c.ParkSource
	yamlData.Procession = c.Procession

	// Marshal YAML
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to marshal context", err)
	}

	// Build full content
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n")
	if c.Body != "" {
		b.WriteString(c.Body)
	}

	return []byte(b.String()), nil
}

// Save writes the context to a file.
func (c *Context) Save(path string) error {
	data, err := c.Serialize()
	if err != nil {
		return err
	}
	if err := fileutil.AtomicWriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write session context", err)
	}
	return nil
}

// Validate checks the context against schema requirements.
func (c *Context) Validate() []string {
	data := map[string]any{
		"session_id":    c.SessionID,
		"status":        string(c.Status),
		"created_at":    c.CreatedAt.UTC().Format(time.RFC3339),
		"initiative":    c.Initiative,
		"complexity":    c.Complexity,
		"active_rite":   c.ActiveRite,
		"current_phase": c.CurrentPhase,
	}
	if c.SchemaVersion != "" {
		data["schema_version"] = c.SchemaVersion
	}
	return validation.ValidateSessionFields(data)
}

// NewContext creates a new session context with defaults.
func NewContext(initiative, complexity, rite string) *Context {
	sessionID := GenerateSessionID()
	now := time.Now().UTC()

	ctx := &Context{
		SchemaVersion: "2.3",
		SessionID:     sessionID,
		Status:        StatusActive,
		CreatedAt:     now,
		Initiative:    initiative,
		Complexity:    complexity,
		ActiveRite:    rite,
		CurrentPhase:  "requirements",
		Body:          defaultBody(initiative),
	}

	// Set rite field - null for cross-cutting
	if rite == "" || rite == "none" {
		ctx.Rite = nil
	} else {
		riteCopy := rite
		ctx.Rite = &riteCopy
	}

	return ctx
}

func defaultBody(initiative string) string {
	return fmt.Sprintf(`
# Session: %s

## Artifacts
- PRD: pending
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete requirements gathering
`, initiative)
}
