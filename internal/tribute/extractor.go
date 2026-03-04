package tribute

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/session"
	"gopkg.in/yaml.v3"
)

// Extractor extracts data from session files for tribute generation.
type Extractor struct {
	sessionPath string
}

// NewExtractor creates a new Extractor for the given session directory.
func NewExtractor(sessionPath string) *Extractor {
	return &Extractor{sessionPath: sessionPath}
}

// ExtractSessionContext parses SESSION_CONTEXT.md and returns context data.
func (e *Extractor) ExtractSessionContext() (*session.Context, error) {
	contextPath := filepath.Join(e.sessionPath, "SESSION_CONTEXT.md")
	return session.LoadContext(contextPath)
}

// ExtractEvents reads and parses events.jsonl.
// Returns empty slice if file doesn't exist (graceful degradation).
func (e *Extractor) ExtractEvents() ([]EventData, error) {
	eventsPath := filepath.Join(e.sessionPath, "events.jsonl")

	f, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []EventData{}, nil // Graceful degradation
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to open events file", err)
	}
	defer func() { _ = f.Close() }()

	var events []EventData
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event EventData
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip malformed events per TDD spec
			continue
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read events", err)
	}

	return events, nil
}

// ExtractArtifacts extracts artifact_created events from events.
func (e *Extractor) ExtractArtifacts(events []EventData) []Artifact {
	var artifacts []Artifact

	for _, event := range events {
		eventType := event.GetEventType()
		if eventType != "tool.artifact_created" {
			continue
		}

		timestamp := parseTimestamp(event.GetTimestamp())

		artifact := Artifact{
			Type:      event.ArtifactType,
			Path:      event.Path,
			Status:    "Created",
			Timestamp: timestamp,
		}

		// Try to get artifact_type from metadata if not set
		if artifact.Type == "" {
			if meta := event.GetMetadata(); meta != nil {
				if at, ok := meta["artifact_type"].(string); ok {
					artifact.Type = at
				}
			}
		}

		// Infer type from path if still empty
		if artifact.Type == "" {
			artifact.Type = inferArtifactType(artifact.Path)
		}

		// Resolve graduated .ledge/ path for non-code artifact types
		artifact.Path = graduatedArtifactPath(artifact.Type, artifact.Path)

		artifacts = append(artifacts, artifact)
	}

	return artifacts
}

// ExtractDecisions extracts decision events from events.
func (e *Extractor) ExtractDecisions(events []EventData) []Decision {
	var decisions []Decision

	for _, event := range events {
		eventType := event.GetEventType()
		if eventType != "agent.decision" {
			continue
		}

		timestamp := parseTimestamp(event.GetTimestamp())

		decision := Decision{
			Timestamp: timestamp,
			Decision:  event.Decision,
			Rationale: event.Rationale,
			Rejected:  event.Rejected,
			Context:   event.Context,
		}

		// Try to get fields from metadata if not set directly
		if meta := event.GetMetadata(); meta != nil {
			if decision.Decision == "" {
				if d, ok := meta["decision"].(string); ok {
					decision.Decision = d
				}
			}
			if decision.Rationale == "" {
				if r, ok := meta["rationale"].(string); ok {
					decision.Rationale = r
				}
			}
		}

		if decision.Decision != "" {
			decisions = append(decisions, decision)
		}
	}

	return decisions
}

// ExtractPhases extracts phase transition events and builds phase records.
func (e *Extractor) ExtractPhases(events []EventData) []PhaseRecord {
	var phases []PhaseRecord
	var currentPhase *PhaseRecord

	for _, event := range events {
		eventType := event.GetEventType()

		// Handle PHASE_TRANSITIONED events
		if eventType == "PHASE_TRANSITIONED" || eventType == "phase_transitioned" {
			timestamp := parseTimestamp(event.GetTimestamp())

			// Close out current phase if exists
			if currentPhase != nil {
				currentPhase.Duration = timestamp.Sub(currentPhase.StartedAt)
				phases = append(phases, *currentPhase)
			}

			// Get agent from metadata
			agent := ""
			if meta := event.GetMetadata(); meta != nil {
				if a, ok := meta["agent"].(string); ok {
					agent = a
				}
			}

			// Start new phase
			currentPhase = &PhaseRecord{
				Phase:     event.ToPhase,
				StartedAt: timestamp,
				Agent:     agent,
			}
		}
	}

	// Don't forget the last phase
	if currentPhase != nil {
		// Use session end time or now for duration
		phases = append(phases, *currentPhase)
	}

	return phases
}

// ExtractHandoffs extracts handoff events.
func (e *Extractor) ExtractHandoffs(events []EventData) []Handoff {
	var handoffs []Handoff

	// Track prepared handoffs to correlate with executed
	prepared := make(map[string]EventData)

	for _, event := range events {
		eventType := event.GetEventType()

		switch eventType {
		case "agent.handoff_prepared":
			// Store for correlation
			key := event.From + "->" + event.To
			prepared[key] = event

		case "agent.handoff_executed":
			timestamp := parseTimestamp(event.GetTimestamp())
			key := event.From + "->" + event.To

			// Get notes from prepared event if available
			notes := event.Notes
			if prep, ok := prepared[key]; ok {
				if notes == "" {
					notes = prep.Notes
				}
			}

			handoff := Handoff{
				From:      event.From,
				To:        event.To,
				Timestamp: timestamp,
				Notes:     notes,
			}

			handoffs = append(handoffs, handoff)
		}
	}

	return handoffs
}

// ExtractMetrics calculates metrics from events.
func (e *Extractor) ExtractMetrics(events []EventData) Metrics {
	metrics := Metrics{
		EventsRecorded: len(events),
	}

	for _, event := range events {
		eventType := event.GetEventType()

		switch eventType {
		case "tool.call":
			metrics.ToolCalls++
		case "tool.file_change":
			metrics.FilesModified++
			// Try to get line stats from metadata
			if meta := event.GetMetadata(); meta != nil {
				if added, ok := meta["lines_added"].(float64); ok {
					metrics.LinesAdded += int(added)
				}
				if removed, ok := meta["lines_removed"].(float64); ok {
					metrics.LinesRemoved += int(removed)
				}
			}
		}
	}

	return metrics
}

// ExtractWhiteSails parses WHITE_SAILS.yaml if it exists.
// Returns nil if file doesn't exist (graceful degradation).
func (e *Extractor) ExtractWhiteSails() (*WhiteSailsData, error) {
	sailsPath := filepath.Join(e.sessionPath, "WHITE_SAILS.yaml")

	content, err := os.ReadFile(sailsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Graceful degradation
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read WHITE_SAILS.yaml", err)
	}

	// Parse YAML into a flexible structure
	var raw map[string]any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "failed to parse WHITE_SAILS.yaml", err)
	}

	data := &WhiteSailsData{
		Proofs: make(map[string]ProofStatus),
	}

	// Extract color
	if color, ok := raw["color"].(string); ok {
		data.Color = color
	}

	// Extract computed_base
	if base, ok := raw["computed_base"].(string); ok {
		data.ComputedBase = base
	}

	// Extract proofs
	if proofs, ok := raw["proofs"].(map[string]any); ok {
		for name, proofData := range proofs {
			if proofMap, ok := proofData.(map[string]any); ok {
				proof := ProofStatus{}
				if status, ok := proofMap["status"].(string); ok {
					proof.Status = status
				}
				if summary, ok := proofMap["summary"].(string); ok {
					proof.Summary = summary
				}
				data.Proofs[name] = proof
			}
		}
	}

	return data, nil
}

// ExtractGraduatedArtifacts reads the session artifact registry and returns
// entries that have been (or would be) graduated to .ledge/.
func (e *Extractor) ExtractGraduatedArtifacts(sessionID string, projectRoot string) []GraduatedArtifact {
	registry := artifact.NewRegistry(projectRoot)

	sessionReg, err := registry.LoadSessionRegistry(sessionID)
	if err != nil || len(sessionReg.Artifacts) == 0 {
		return nil
	}

	var graduated []GraduatedArtifact
	for _, entry := range sessionReg.Artifacts {
		category := artifact.LedgeCategoryForType(entry.ArtifactType)
		if category == "" {
			continue
		}
		graduated = append(graduated, GraduatedArtifact{
			Type:          string(entry.ArtifactType),
			OriginalPath:  entry.Path,
			GraduatedPath: registry.GraduatedPath(entry),
			Category:      category,
		})
	}

	return graduated
}

// ExtractNotes extracts relevant notes from SESSION_CONTEXT body.
// Filters out boilerplate sections.
func (e *Extractor) ExtractNotes(body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}

	// Check if it's just default boilerplate
	if isDefaultBoilerplate(body) {
		return ""
	}

	// Extract meaningful content, skipping standard sections
	return strings.TrimSpace(body)
}

// Helper functions

// parseTimestamp parses a timestamp string into time.Time.
func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}

	// Try RFC3339 first (most common)
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t
	}

	// Try RFC3339Nano for milliseconds
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t
	}

	return time.Time{}
}

// graduatedArtifactPath resolves a tribute artifact path to its .ledge/ location.
// Maps the tribute artifact type string to an artifact.ArtifactType for ledge resolution.
// Returns the original path for types that stay in the source tree.
func graduatedArtifactPath(artifactType string, path string) string {
	at := tributeTypeToArtifactType(artifactType)
	category := artifact.LedgeCategoryForType(at)
	if category == "" {
		return path
	}
	filename := filepath.Base(path)
	return filepath.Join(".ledge", category, filename)
}

// tributeTypeToArtifactType maps tribute type strings to artifact.ArtifactType.
func tributeTypeToArtifactType(t string) artifact.ArtifactType {
	switch strings.ToLower(t) {
	case "prd":
		return artifact.TypePRD
	case "tdd":
		return artifact.TypeTDD
	case "adr":
		return artifact.TypeADR
	case "test-plan":
		return artifact.TypeTestPlan
	case "runbook":
		return artifact.TypeRunbook
	case "review":
		return artifact.TypeReview
	case "spike":
		return artifact.TypeSpike
	case "code", "tests":
		return artifact.TypeCode
	default:
		return artifact.TypeCode // unknown types stay in source tree
	}
}

// inferArtifactType infers artifact type from file path.
func inferArtifactType(path string) string {
	lower := strings.ToLower(path)

	switch {
	case strings.Contains(lower, "prd-") || strings.Contains(lower, "/requirements/"):
		return "PRD"
	case strings.Contains(lower, "tdd-") || strings.Contains(lower, "/design/"):
		return "TDD"
	case strings.Contains(lower, "adr-") || strings.Contains(lower, "/decisions/"):
		return "ADR"
	case strings.Contains(lower, "test") || strings.Contains(lower, "_test."):
		return "Tests"
	default:
		return "Code"
	}
}

// isDefaultBoilerplate checks if body is just default session template.
func isDefaultBoilerplate(body string) bool {
	// Check for common default patterns
	defaults := []string{
		"## Artifacts\n- PRD: pending\n- TDD: pending",
		"## Blockers\nNone yet.",
		"## Next Steps\n1. Complete requirements gathering",
	}

	for _, d := range defaults {
		if strings.Contains(body, d) && len(body) < 500 {
			return true
		}
	}

	return false
}
