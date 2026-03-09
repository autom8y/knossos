package naxos

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/fileutil"
)

// TriageArtifactFile is the canonical filename for the triage artifact.
const TriageArtifactFile = "NAXOS_TRIAGE.md"

// triageFrontmatter is the YAML frontmatter written to the artifact file.
// It captures the summary fields so ReadTriageSummary can extract the
// summary_line without parsing the full JSON or markdown table.
type triageFrontmatter struct {
	SchemaVersion string          `yaml:"schema_version"`
	TriagedAt     string          `yaml:"triaged_at"`
	TotalScanned  int             `yaml:"total_scanned"`
	TotalTriaged  int             `yaml:"total_triaged"`
	SummaryLine   string          `yaml:"summary_line"`
	BySeverity    map[string]int  `yaml:"by_severity"`
}

// WriteTriageArtifact writes a NAXOS_TRIAGE.md artifact to sessionsDir.
// The artifact contains YAML frontmatter with summary fields and a markdown
// table of triage entries. Uses atomic write to prevent partial files.
func WriteTriageArtifact(sessionsDir string, result *TriageResult) error {
	content, err := formatArtifact(result)
	if err != nil {
		return fmt.Errorf("naxos: format triage artifact: %w", err)
	}

	path := filepath.Join(sessionsDir, TriageArtifactFile)
	if err := fileutil.AtomicWriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("naxos: write triage artifact %s: %w", path, err)
	}
	return nil
}

// ReadTriageArtifact reads and parses a NAXOS_TRIAGE.md artifact from sessionsDir.
// Returns an error if the file does not exist or cannot be parsed.
func ReadTriageArtifact(sessionsDir string) (*TriageResult, error) {
	path := filepath.Join(sessionsDir, TriageArtifactFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("naxos: read triage artifact %s: %w", path, err)
	}

	fm, entries, err := parseArtifact(data)
	if err != nil {
		return nil, fmt.Errorf("naxos: parse triage artifact %s: %w", path, err)
	}

	triaged, err := time.Parse(time.RFC3339, fm.TriagedAt)
	if err != nil {
		return nil, fmt.Errorf("naxos: parse triaged_at in %s: %w", path, err)
	}

	bySeverity := make(map[Severity]int, len(fm.BySeverity))
	for k, v := range fm.BySeverity {
		bySeverity[Severity(k)] = v
	}

	return &TriageResult{
		Entries:      entries,
		TotalScanned: fm.TotalScanned,
		TotalTriaged: fm.TotalTriaged,
		BySeverity:   bySeverity,
		SummaryLine:  fm.SummaryLine,
		TriagedAt:    triaged,
	}, nil
}

// ReadTriageSummary returns the summary_line from a NAXOS_TRIAGE.md artifact
// without parsing the full file. Returns an empty string if the artifact does
// not exist or is unreadable. Designed to complete in <5ms for hook use.
func ReadTriageSummary(sessionsDir string) string {
	path := filepath.Join(sessionsDir, TriageArtifactFile)
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close() //nolint:errcheck

	// Read the YAML frontmatter block (between the --- delimiters).
	// We do a line-by-line scan to avoid loading large markdown body into memory.
	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	var fmLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if !inFrontmatter {
			if line == "---" {
				inFrontmatter = true
			}
			continue
		}
		if line == "---" {
			break
		}
		fmLines = append(fmLines, line)
	}

	// Extract summary_line without full YAML parse — simple key search.
	for _, l := range fmLines {
		if val, ok := strings.CutPrefix(l, "summary_line: "); ok {
			// Strip optional surrounding quotes.
			return strings.Trim(val, `"`)
		}
	}
	return ""
}

// formatArtifact renders the TriageResult as a NAXOS_TRIAGE.md document.
func formatArtifact(result *TriageResult) ([]byte, error) {
	// Build by_severity map with string keys for YAML portability.
	bySeverity := make(map[string]int, len(result.BySeverity))
	for k, v := range result.BySeverity {
		bySeverity[string(k)] = v
	}

	fm := triageFrontmatter{
		SchemaVersion: "1.0",
		TriagedAt:     result.TriagedAt.UTC().Format(time.RFC3339),
		TotalScanned:  result.TotalScanned,
		TotalTriaged:  result.TotalTriaged,
		SummaryLine:   result.SummaryLine,
		BySeverity:    bySeverity,
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n")
	buf.WriteString("\n## Triage Entries\n\n")
	buf.WriteString("| # | Session ID | Severity | Reason | Inactive | Suggested Action |\n")
	buf.WriteString("|---|------------|----------|--------|----------|------------------|\n")

	for i, entry := range result.Entries {
		sessionID := entry.SessionID
		if len(sessionID) > 35 {
			sessionID = sessionID[:32] + "..."
		}
		buf.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n",
			i+1,
			sessionID,
			entry.Severity,
			entry.Reason,
			FormatDuration(entry.InactiveFor),
			entry.SuggestedAction,
		))
	}

	return buf.Bytes(), nil
}

// parseArtifact extracts the frontmatter and reconstructs TriageEntry slice
// from a NAXOS_TRIAGE.md file. The markdown table rows are parsed back into
// lightweight TriageEntry values; full OrphanedSession fields not stored in
// the table are zeroed (artifact is a summary, not a full reconstruction).
func parseArtifact(data []byte) (triageFrontmatter, []TriageEntry, error) {
	var fm triageFrontmatter

	content := string(data)

	// Extract frontmatter.
	if !strings.HasPrefix(content, "---\n") {
		return fm, nil, fmt.Errorf("missing frontmatter delimiter")
	}
	rest := content[4:]
	endIdx := strings.Index(rest, "\n---\n")
	if endIdx < 0 {
		return fm, nil, fmt.Errorf("unclosed frontmatter block")
	}
	fmBlock := rest[:endIdx]

	if err := yaml.Unmarshal([]byte(fmBlock), &fm); err != nil {
		return fm, nil, fmt.Errorf("unmarshal frontmatter: %w", err)
	}

	// Parse table rows from the body.
	body := rest[endIdx+5:] // skip "\n---\n"
	lines := strings.Split(body, "\n")

	var entries []TriageEntry
	inTable := false
	headerSeen := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			inTable = false
			continue
		}
		inTable = true
		if !headerSeen {
			// First pipe line is the header, second is separator.
			headerSeen = true
			continue
		}
		// Skip separator line (contains ---)
		if strings.Contains(line, "---") {
			continue
		}
		entry, ok := parseTableRow(line)
		if ok {
			entries = append(entries, entry)
		}
	}
	_ = inTable

	return fm, entries, nil
}

// parseTableRow parses a single markdown table row into a TriageEntry.
// Columns: # | Session ID | Severity | Reason | Inactive | Suggested Action
func parseTableRow(line string) (TriageEntry, bool) {
	// Strip leading/trailing pipes and split.
	line = strings.Trim(line, "|")
	cols := strings.Split(line, "|")
	if len(cols) < 6 {
		return TriageEntry{}, false
	}

	trim := func(s string) string { return strings.TrimSpace(s) }

	sessionID := trim(cols[1])
	severity := Severity(trim(cols[2]))
	reason := OrphanReason(trim(cols[3]))
	suggestedAction := SuggestedAction(trim(cols[5]))

	return TriageEntry{
		OrphanedSession: OrphanedSession{
			SessionID:       sessionID,
			Reason:          reason,
			SuggestedAction: suggestedAction,
		},
		Severity: severity,
	}, true
}
