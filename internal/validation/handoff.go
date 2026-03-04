// Package validation provides handoff validation for phase transitions.
package validation

import (
	"embed"
	"fmt"
	"reflect"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/handoff-criteria.yaml
var handoffCriteriaFS embed.FS

// Phase represents a workflow phase for handoff validation.
type Phase string

const (
	PhaseRequirements   Phase = "requirements"
	PhaseDesign         Phase = "design"
	PhaseImplementation Phase = "implementation"
	PhaseValidation     Phase = "validation"
)

// String returns the string representation of the phase.
func (p Phase) String() string {
	return string(p)
}

// ParsePhase parses a string into a Phase.
func ParsePhase(s string) Phase {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "requirements":
		return PhaseRequirements
	case "design":
		return PhaseDesign
	case "implementation":
		return PhaseImplementation
	case "validation":
		return PhaseValidation
	default:
		return ""
	}
}

// ValidPhases returns all valid phase strings.
func ValidPhases() []string {
	return []string{
		string(PhaseRequirements),
		string(PhaseDesign),
		string(PhaseImplementation),
		string(PhaseValidation),
	}
}

// Criterion represents a single validation criterion.
type Criterion struct {
	// Field is the frontmatter field to validate.
	Field string `yaml:"field" json:"field"`

	// Message is displayed if the criterion is not met.
	Message string `yaml:"message" json:"message"`

	// NonEmpty requires the field to be non-empty (not just present).
	NonEmpty bool `yaml:"non_empty" json:"non_empty,omitempty"`

	// MinItems is the minimum number of items for array fields.
	MinItems *int `yaml:"min_items" json:"min_items,omitempty"`
}

// ArtifactCriteria holds blocking and non-blocking criteria for an artifact type.
type ArtifactCriteria struct {
	Blocking    []Criterion `yaml:"blocking" json:"blocking,omitempty"`
	NonBlocking []Criterion `yaml:"non_blocking" json:"non_blocking,omitempty"`
}

// HandoffCriteria maps phases to artifact types to their criteria.
type HandoffCriteria map[Phase]map[ArtifactType]ArtifactCriteria

// CriterionResult represents the result of evaluating a single criterion.
type CriterionResult struct {
	Criterion Criterion `json:"criterion"`
	Passed    bool      `json:"passed"`
	Message   string    `json:"message,omitempty"`
	Value     any       `json:"value,omitempty"`
}

// HandoffResult contains the result of handoff validation.
type HandoffResult struct {
	// Passed is true if all blocking criteria passed.
	Passed bool `json:"passed"`

	// Phase is the validated phase.
	Phase Phase `json:"phase"`

	// ArtifactType is the validated artifact type.
	ArtifactType ArtifactType `json:"artifact_type"`

	// FilePath is the path to the validated artifact.
	FilePath string `json:"file_path,omitempty"`

	// BlockingResults contains results for blocking criteria.
	BlockingResults []CriterionResult `json:"blocking_results,omitempty"`

	// WarningResults contains results for non-blocking criteria that failed.
	WarningResults []CriterionResult `json:"warning_results,omitempty"`

	// Frontmatter contains the parsed frontmatter data.
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
}

// FailedBlocking returns the blocking criteria that failed.
func (r *HandoffResult) FailedBlocking() []CriterionResult {
	var failed []CriterionResult
	for _, cr := range r.BlockingResults {
		if !cr.Passed {
			failed = append(failed, cr)
		}
	}
	return failed
}

// Warnings returns the non-blocking criteria that failed.
func (r *HandoffResult) Warnings() []CriterionResult {
	var warnings []CriterionResult
	for _, cr := range r.WarningResults {
		if !cr.Passed {
			warnings = append(warnings, cr)
		}
	}
	return warnings
}

// HandoffValidator validates artifacts against handoff criteria.
type HandoffValidator struct {
	criteria HandoffCriteria
}

// NewHandoffValidator creates a new handoff validator.
func NewHandoffValidator() (*HandoffValidator, error) {
	data, err := handoffCriteriaFS.ReadFile("schemas/handoff-criteria.yaml")
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read handoff criteria", err)
	}

	criteria, err := parseCriteria(data)
	if err != nil {
		return nil, err
	}

	return &HandoffValidator{criteria: criteria}, nil
}

// parseCriteria parses the YAML criteria file into HandoffCriteria.
func parseCriteria(data []byte) (HandoffCriteria, error) {
	// Parse into raw structure first
	var raw map[string]map[string]struct {
		Blocking    []Criterion `yaml:"blocking"`
		NonBlocking []Criterion `yaml:"non_blocking"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse handoff criteria YAML", err)
	}

	// Convert to typed structure
	criteria := make(HandoffCriteria)
	for phaseStr, artifactMap := range raw {
		phase := ParsePhase(phaseStr)
		if phase == "" {
			return nil, errors.NewWithDetails(errors.CodeSchemaInvalid,
				"invalid phase in handoff criteria",
				map[string]any{"phase": phaseStr, "valid": ValidPhases()})
		}

		criteria[phase] = make(map[ArtifactType]ArtifactCriteria)
		for artifactStr, ac := range artifactMap {
			artifactType := ParseArtifactType(artifactStr)
			if artifactType == ArtifactTypeUnknown {
				return nil, errors.NewWithDetails(errors.CodeSchemaInvalid,
					"invalid artifact type in handoff criteria",
					map[string]any{"artifact_type": artifactStr, "valid": ValidArtifactTypes()})
			}
			criteria[phase][artifactType] = ArtifactCriteria{
				Blocking:    ac.Blocking,
				NonBlocking: ac.NonBlocking,
			}
		}
	}

	return criteria, nil
}

// GetCriteria returns the criteria for a phase and artifact type.
func (hv *HandoffValidator) GetCriteria(phase Phase, artifactType ArtifactType) (*ArtifactCriteria, error) {
	phaseMap, ok := hv.criteria[phase]
	if !ok {
		return nil, errors.NewWithDetails(errors.CodeSchemaNotFound,
			"no criteria defined for phase",
			map[string]any{"phase": string(phase), "valid": ValidPhases()})
	}

	criteria, ok := phaseMap[artifactType]
	if !ok {
		// Return empty criteria if artifact type not defined for this phase
		return &ArtifactCriteria{}, nil
	}

	return &criteria, nil
}

// ListPhases returns all phases that have criteria defined.
func (hv *HandoffValidator) ListPhases() []Phase {
	phases := make([]Phase, 0, len(hv.criteria))
	for phase := range hv.criteria {
		phases = append(phases, phase)
	}
	return phases
}

// ListArtifactTypes returns all artifact types that have criteria for a phase.
func (hv *HandoffValidator) ListArtifactTypes(phase Phase) []ArtifactType {
	phaseMap, ok := hv.criteria[phase]
	if !ok {
		return nil
	}

	types := make([]ArtifactType, 0, len(phaseMap))
	for artifactType := range phaseMap {
		types = append(types, artifactType)
	}
	return types
}

// ValidateHandoff validates an artifact's frontmatter against handoff criteria.
func (hv *HandoffValidator) ValidateHandoff(phase Phase, artifactType ArtifactType, frontmatter map[string]any) (*HandoffResult, error) {
	result := &HandoffResult{
		Phase:        phase,
		ArtifactType: artifactType,
		Frontmatter:  frontmatter,
		Passed:       true,
	}

	criteria, err := hv.GetCriteria(phase, artifactType)
	if err != nil {
		return nil, err
	}

	// Evaluate blocking criteria
	for _, criterion := range criteria.Blocking {
		cr := evaluateCriterion(criterion, frontmatter)
		result.BlockingResults = append(result.BlockingResults, cr)
		if !cr.Passed {
			result.Passed = false
		}
	}

	// Evaluate non-blocking criteria
	for _, criterion := range criteria.NonBlocking {
		cr := evaluateCriterion(criterion, frontmatter)
		if !cr.Passed {
			result.WarningResults = append(result.WarningResults, cr)
		}
	}

	return result, nil
}

// ValidateHandoffFile validates an artifact file against handoff criteria.
func (hv *HandoffValidator) ValidateHandoffFile(phase Phase, filePath string) (*HandoffResult, error) {
	// Create artifact validator to parse frontmatter
	av, err := NewArtifactValidator()
	if err != nil {
		return nil, err
	}

	// Validate the file to get frontmatter and artifact type
	artifactResult, err := av.ValidateFile(filePath, ArtifactTypeUnknown)
	if err != nil {
		return nil, err
	}

	// Check for frontmatter issues
	if artifactResult.Frontmatter == nil {
		return &HandoffResult{
			Phase:        phase,
			ArtifactType: artifactResult.ArtifactType,
			FilePath:     filePath,
			Passed:       false,
			BlockingResults: []CriterionResult{
				{
					Criterion: Criterion{Message: "Failed to extract frontmatter"},
					Passed:    false,
					Message:   "Artifact has no valid frontmatter",
				},
			},
		}, nil
	}

	// Validate handoff criteria
	result, err := hv.ValidateHandoff(phase, artifactResult.ArtifactType, artifactResult.Frontmatter)
	if err != nil {
		return nil, err
	}

	result.FilePath = filePath
	return result, nil
}

// evaluateCriterion evaluates a single criterion against frontmatter.
func evaluateCriterion(criterion Criterion, frontmatter map[string]any) CriterionResult {
	result := CriterionResult{
		Criterion: criterion,
		Passed:    true,
	}

	value, exists := frontmatter[criterion.Field]
	result.Value = value

	// Check if field exists
	if !exists {
		result.Passed = false
		result.Message = criterion.Message
		return result
	}

	// Check non-empty requirement
	if criterion.NonEmpty && isEmpty(value) {
		result.Passed = false
		result.Message = criterion.Message
		return result
	}

	// Check min_items requirement for arrays
	if criterion.MinItems != nil {
		count := getItemCount(value)
		if count < *criterion.MinItems {
			result.Passed = false
			result.Message = fmt.Sprintf("%s (has %d, needs %d)", criterion.Message, count, *criterion.MinItems)
			return result
		}
	}

	return result
}

// isEmpty checks if a value is empty.
func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// getItemCount returns the count of items for array/slice values.
func getItemCount(value any) int {
	if value == nil {
		return 0
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return v.Len()
	default:
		// Non-array values count as 1 if present
		return 1
	}
}
