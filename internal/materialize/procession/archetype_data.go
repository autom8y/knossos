package procession

import (
	"fmt"
	"strings"

	"github.com/autom8y/knossos/internal/procession"
)

// ProcessionWorkflowData provides template variables for the
// procession-workflow archetype template.
type ProcessionWorkflowData struct {
	Name         string              // e.g., "security-remediation"
	Description  string              // From template.Description
	Stations     []procession.Station // Full station list (for template access)
	StationTable string              // Pre-rendered markdown table of stations
	ArtifactDir  string              // From template.ArtifactDir
	SkillName    string              // "{name}-ref"
	FirstStation string              // First station name
	FirstRite    string              // First station's rite
	StationCount int                 // Total number of stations
}

// BuildWorkflowData constructs a ProcessionWorkflowData from a validated template.
func BuildWorkflowData(tmpl *procession.Template) *ProcessionWorkflowData {
	data := &ProcessionWorkflowData{
		Name:         tmpl.Name,
		Description:  tmpl.Description,
		Stations:     tmpl.Stations,
		ArtifactDir:  tmpl.ArtifactDir,
		SkillName:    tmpl.Name + "-ref",
		StationCount: len(tmpl.Stations),
	}

	if len(tmpl.Stations) > 0 {
		data.FirstStation = tmpl.Stations[0].Name
		data.FirstRite = tmpl.Stations[0].Rite
	}

	data.StationTable = buildStationTable(tmpl.Stations)
	return data
}

// buildStationTable renders a markdown table of station metadata.
func buildStationTable(stations []procession.Station) string {
	var b strings.Builder
	b.WriteString("| # | Station | Rite | Alt Rite | Goal | Produces | Loop To |\n")
	b.WriteString("|---|---------|------|----------|------|----------|---------|\n")

	for i, s := range stations {
		altRite := "-"
		if s.AltRite != "" {
			altRite = s.AltRite
		}
		loopTo := "-"
		if s.LoopTo != "" {
			loopTo = s.LoopTo
		}
		produces := strings.Join(s.Produces, ", ")

		// Truncate goal for table readability
		goal := s.Goal
		if len(goal) > 60 {
			goal = goal[:57] + "..."
		}

		fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %s | %s |\n",
			i+1, s.Name, s.Rite, altRite, goal, produces, loopTo)
	}
	return b.String()
}
