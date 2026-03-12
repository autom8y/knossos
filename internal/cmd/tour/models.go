// Package tour implements the ari tour command for directory structure orientation.
package tour

import (
	"fmt"
	"strings"
)

// TourOutput represents the full directory tour result.
type TourOutput struct {
	ProjectRoot string          `json:"project_root"`
	Directories TourDirectories `json:"directories"`
}

// TourDirectories holds all five directory sections.
type TourDirectories struct {
	Channel ChannelSection  `json:"channel"`
	Knossos KnossosSection `json:"knossos"`
	Know    KnowSection    `json:"know"`
	Ledge   LedgeSection   `json:"ledge"`
	SOS     SOSSection     `json:"sos"`
}

// ChannelSection represents channel directory state for tour output.
type ChannelSection struct {
	Exists       bool     `json:"exists"`
	Path         string   `json:"path"`
	Agents       DirCount `json:"agents"`
	Commands     DirCount `json:"commands"`
	Skills       DirCount `json:"skills"`
	SettingsJSON bool     `json:"settings_json"`
	ClaudeMD     bool     `json:"claude_md"`
	ActiveRite   string   `json:"active_rite,omitempty"`
}

// KnossosSection represents .knossos/ directory state for tour output.
type KnossosSection struct {
	Exists    bool     `json:"exists"`
	Path      string   `json:"path"`
	Rites     DirCount `json:"rites"`
	Templates DirCount `json:"templates,omitempty"`
}

// KnowSection represents .know/ directory state for tour output.
type KnowSection struct {
	Exists  bool     `json:"exists"`
	Path    string   `json:"path"`
	Domains DirCount `json:"domains"`
}

// LedgeSection represents .ledge/ directory state for tour output.
type LedgeSection struct {
	Exists    bool     `json:"exists"`
	Path      string   `json:"path"`
	Decisions DirCount `json:"decisions"`
	Specs     DirCount `json:"specs"`
	Reviews   DirCount `json:"reviews"`
	Spikes    DirCount `json:"spikes"`
}

// SOSSection represents .sos/ directory state for tour output.
type SOSSection struct {
	Exists   bool        `json:"exists"`
	Path     string      `json:"path"`
	Sessions SOSSessions `json:"sessions"`
	Archive  DirCount    `json:"archive"`
}

// SOSSessions extends DirCount with active/parked breakdown.
type SOSSessions struct {
	Count  int `json:"count"`
	Active int `json:"active"`
	Parked int `json:"parked"`
}

// DirCount holds a count and optional item list for a directory.
type DirCount struct {
	Count int      `json:"count"`
	Items []string `json:"items,omitempty"`
}

// Text implements output.Textable for TourOutput.
func (t TourOutput) Text() string {
	var b strings.Builder

	b.WriteString("=== Project Tour ===\n")

	// Channel directory
	b.WriteString("\nchannel/\n")
	if !t.Directories.Channel.Exists {
		b.WriteString("  (not found)\n")
	} else {
		c := t.Directories.Channel
		fmt.Fprintf(&b, "  agents/        %d agents\n", c.Agents.Count)
		fmt.Fprintf(&b, "  commands/      %d commands\n", c.Commands.Count)
		fmt.Fprintf(&b, "  skills/        %d skills\n", c.Skills.Count)
		if c.SettingsJSON {
			b.WriteString("  settings.json  present\n")
		} else {
			b.WriteString("  settings.json  missing\n")
		}
		if c.ClaudeMD {
			b.WriteString("  CLAUDE.md      present\n")
		} else {
			b.WriteString("  CLAUDE.md      missing\n")
		}
		if c.ActiveRite != "" {
			fmt.Fprintf(&b, "  ACTIVE_RITE    %s\n", c.ActiveRite)
		} else {
			b.WriteString("  ACTIVE_RITE    (none)\n")
		}
	}

	// .knossos/
	b.WriteString("\n.knossos/\n")
	if !t.Directories.Knossos.Exists {
		b.WriteString("  (not found)\n")
	} else {
		k := t.Directories.Knossos
		if k.Rites.Count > 0 && len(k.Rites.Items) > 0 {
			fmt.Fprintf(&b, "  rites/         %d rites (%s)\n",
				k.Rites.Count, strings.Join(k.Rites.Items, ", "))
		} else {
			fmt.Fprintf(&b, "  rites/         %d rites\n", k.Rites.Count)
		}
		if k.Templates.Count > 0 {
			fmt.Fprintf(&b, "  templates/     %d files\n", k.Templates.Count)
		}
	}

	// .know/
	b.WriteString("\n.know/\n")
	if !t.Directories.Know.Exists {
		b.WriteString("  (not found)\n")
	} else {
		kn := t.Directories.Know
		if kn.Domains.Count > 0 && len(kn.Domains.Items) > 0 {
			fmt.Fprintf(&b, "  %d domain files (%s)\n",
				kn.Domains.Count, strings.Join(kn.Domains.Items, ", "))
		} else {
			fmt.Fprintf(&b, "  %d domain files\n", kn.Domains.Count)
		}
	}

	// .ledge/
	b.WriteString("\n.ledge/\n")
	if !t.Directories.Ledge.Exists {
		b.WriteString("  (not found)\n")
	} else {
		l := t.Directories.Ledge
		fmt.Fprintf(&b, "  decisions/     %s\n", pluralizeFile(l.Decisions.Count))
		fmt.Fprintf(&b, "  specs/         %s\n", pluralizeFile(l.Specs.Count))
		fmt.Fprintf(&b, "  reviews/       %s\n", pluralizeFile(l.Reviews.Count))
		fmt.Fprintf(&b, "  spikes/        %s\n", pluralizeFile(l.Spikes.Count))
	}

	// .sos/
	b.WriteString("\n.sos/\n")
	if !t.Directories.SOS.Exists {
		b.WriteString("  (not found)\n")
	} else {
		s := t.Directories.SOS
		fmt.Fprintf(&b, "  sessions/      %d sessions (%d active, %d parked)\n",
			s.Sessions.Count, s.Sessions.Active, s.Sessions.Parked)
		fmt.Fprintf(&b, "  archive/       %d archived\n", s.Archive.Count)
	}

	return b.String()
}

// pluralizeFile returns "N file" or "N files" based on count.
func pluralizeFile(n int) string {
	if n == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", n)
}
