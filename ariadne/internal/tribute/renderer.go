package tribute

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Renderer generates TRIBUTE.md content from a GenerateResult.
type Renderer struct{}

// NewRenderer creates a new Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render generates the complete TRIBUTE.md content.
func (r *Renderer) Render(result *GenerateResult) ([]byte, error) {
	var b strings.Builder

	// 1. YAML Frontmatter
	if err := r.renderFrontmatter(&b, result); err != nil {
		return nil, err
	}

	// 2. Title
	b.WriteString(fmt.Sprintf("# Tribute: %s\n\n", result.Initiative))

	// 3. Session quote
	endDate := result.EndedAt
	if endDate.IsZero() {
		endDate = result.GeneratedAt
	}
	b.WriteString(fmt.Sprintf("> Session `%s` completed on %s\n\n", result.SessionID, endDate.Format("2006-01-02")))

	// 4. Summary section
	r.renderSummary(&b, result)

	// 5. Artifacts Produced (always included, even if empty)
	r.renderArtifacts(&b, result)

	// 6. Decisions Made (conditional)
	if len(result.Decisions) > 0 {
		r.renderDecisions(&b, result)
	}

	// 7. Phase Progression (conditional)
	if len(result.Phases) > 0 {
		r.renderPhases(&b, result)
	}

	// 8. Handoffs (conditional)
	if len(result.Handoffs) > 0 {
		r.renderHandoffs(&b, result)
	}

	// 9. Git Commits (conditional - Phase 2, placeholder)
	if len(result.Commits) > 0 {
		r.renderCommits(&b, result)
	}

	// 10. White Sails Attestation (conditional)
	if result.SailsColor != "" {
		r.renderSails(&b, result)
	}

	// 11. Metrics (always included)
	r.renderMetrics(&b, result)

	// 12. Notes (conditional)
	if result.Notes != "" {
		r.renderNotes(&b, result)
	}

	return []byte(b.String()), nil
}

// renderFrontmatter writes the YAML frontmatter.
func (r *Renderer) renderFrontmatter(b *strings.Builder, result *GenerateResult) error {
	frontmatter := TributeFrontmatter{
		SchemaVersion: SchemaVersion,
		SessionID:     result.SessionID,
		Initiative:    result.Initiative,
		Complexity:    result.Complexity,
		GeneratedAt:   result.GeneratedAt.UTC().Format(time.RFC3339),
	}

	// Calculate duration in hours
	if result.Duration > 0 {
		frontmatter.DurationHours = result.Duration.Hours()
	}

	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return err
	}

	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n\n")

	return nil
}

// renderSummary writes the Summary section.
func (r *Renderer) renderSummary(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Summary\n\n")

	b.WriteString(fmt.Sprintf("**Initiative**: %s\n", result.Initiative))

	// Complexity with estimate
	complexityDesc := complexityDescription(result.Complexity)
	b.WriteString(fmt.Sprintf("**Complexity**: %s%s\n", result.Complexity, complexityDesc))

	// Duration with timestamps
	if !result.StartedAt.IsZero() && !result.EndedAt.IsZero() {
		duration := formatDuration(result.Duration)
		b.WriteString(fmt.Sprintf("**Duration**: %s (%s to %s)\n",
			duration,
			result.StartedAt.UTC().Format(time.RFC3339),
			result.EndedAt.UTC().Format(time.RFC3339)))
	} else if result.Duration > 0 {
		b.WriteString(fmt.Sprintf("**Duration**: %s\n", formatDuration(result.Duration)))
	}

	// Team/Rite
	if result.Team != "" {
		b.WriteString(fmt.Sprintf("**Team/Rite**: %s\n", result.Team))
	}

	// Final Phase
	if result.FinalPhase != "" {
		b.WriteString(fmt.Sprintf("**Final Phase**: %s\n", result.FinalPhase))
	}

	// Confidence Signal
	if result.SailsColor != "" {
		b.WriteString(fmt.Sprintf("**Confidence Signal**: %s\n", result.SailsColor))
	}

	b.WriteString("\n")
}

// renderArtifacts writes the Artifacts Produced section.
func (r *Renderer) renderArtifacts(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Artifacts Produced\n\n")

	if len(result.Artifacts) == 0 {
		b.WriteString("No artifacts recorded.\n\n")
		return
	}

	b.WriteString("| Type | Path | Status |\n")
	b.WriteString("|------|------|--------|\n")

	for _, artifact := range result.Artifacts {
		b.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n",
			artifact.Type, artifact.Path, artifact.Status))
	}

	b.WriteString("\n")
}

// renderDecisions writes the Decisions Made section.
func (r *Renderer) renderDecisions(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Decisions Made\n\n")

	b.WriteString("| Timestamp | Decision | Rationale |\n")
	b.WriteString("|-----------|----------|----------|\n")

	for _, decision := range result.Decisions {
		ts := decision.Timestamp.UTC().Format(time.RFC3339)
		// Truncate long text for table display
		dec := truncate(decision.Decision, 60)
		rat := truncate(decision.Rationale, 60)
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", ts, dec, rat))
	}

	b.WriteString("\n")
}

// renderPhases writes the Phase Progression section.
func (r *Renderer) renderPhases(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Phase Progression\n\n")

	// ASCII timeline
	b.WriteString("```\n")
	for i, phase := range result.Phases {
		duration := formatDuration(phase.Duration)
		if i > 0 {
			b.WriteString(fmt.Sprintf(" -(%s)-> ", duration))
		}
		b.WriteString(phase.Phase)
	}
	b.WriteString("\n```\n\n")

	// Table
	b.WriteString("| Phase | Started | Duration | Agent |\n")
	b.WriteString("|-------|---------|----------|-------|\n")

	for _, phase := range result.Phases {
		started := phase.StartedAt.UTC().Format(time.RFC3339)
		duration := formatDuration(phase.Duration)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			phase.Phase, started, duration, phase.Agent))
	}

	b.WriteString("\n")
}

// renderHandoffs writes the Handoffs section.
func (r *Renderer) renderHandoffs(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Handoffs\n\n")

	b.WriteString("| From | To | Timestamp | Notes |\n")
	b.WriteString("|------|----|-----------|-------|\n")

	for _, handoff := range result.Handoffs {
		ts := handoff.Timestamp.UTC().Format(time.RFC3339)
		notes := truncate(handoff.Notes, 40)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			handoff.From, handoff.To, ts, notes))
	}

	b.WriteString("\n")
}

// renderCommits writes the Git Commits section (Phase 2 placeholder).
func (r *Renderer) renderCommits(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Git Commits\n\n")

	b.WriteString("| Hash | Message | Files Changed |\n")
	b.WriteString("|------|---------|---------------|\n")

	for _, commit := range result.Commits {
		msg := truncate(commit.Message, 50)
		b.WriteString(fmt.Sprintf("| %s | %s | %d |\n",
			commit.ShortHash, msg, commit.FilesChanged))
	}

	b.WriteString("\n")
}

// renderSails writes the White Sails Attestation section.
func (r *Renderer) renderSails(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## White Sails Attestation\n\n")

	b.WriteString(fmt.Sprintf("**Color**: %s\n", result.SailsColor))
	if result.SailsBase != "" && result.SailsBase != result.SailsColor {
		b.WriteString(fmt.Sprintf("**Computed Base**: %s\n", result.SailsBase))
	}

	if len(result.SailsProofs) > 0 {
		b.WriteString("**Proofs**:\n")
		for name, proof := range result.SailsProofs {
			summary := ""
			if proof.Summary != "" {
				summary = fmt.Sprintf(" (%s)", proof.Summary)
			}
			b.WriteString(fmt.Sprintf("- %s: %s%s\n", name, proof.Status, summary))
		}
	}

	b.WriteString("\n")
}

// renderMetrics writes the Metrics section.
func (r *Renderer) renderMetrics(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Metrics\n\n")

	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")

	b.WriteString(fmt.Sprintf("| Tool Calls | %d |\n", result.Metrics.ToolCalls))
	b.WriteString(fmt.Sprintf("| Events Recorded | %d |\n", result.Metrics.EventsRecorded))
	b.WriteString(fmt.Sprintf("| Files Modified | %d |\n", result.Metrics.FilesModified))
	b.WriteString(fmt.Sprintf("| Lines Added | %d |\n", result.Metrics.LinesAdded))
	b.WriteString(fmt.Sprintf("| Lines Removed | %d |\n", result.Metrics.LinesRemoved))

	b.WriteString("\n")
}

// renderNotes writes the Notes section.
func (r *Renderer) renderNotes(b *strings.Builder, result *GenerateResult) {
	b.WriteString("## Notes\n\n")
	b.WriteString(result.Notes)
	b.WriteString("\n")
}

// Helper functions

// complexityDescription returns a parenthetical description for complexity.
func complexityDescription(complexity string) string {
	switch strings.ToUpper(complexity) {
	case "SCRIPT":
		return " (estimated 1-2 hours)"
	case "MODULE":
		return " (estimated 4-8 hours)"
	case "SERVICE":
		return " (estimated 1-2 days)"
	case "SYSTEM":
		return " (estimated 1+ weeks)"
	default:
		return ""
	}
}

// formatDuration formats a duration as "Xh Ym".
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// truncate shortens a string to maxLen with "..." suffix.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
