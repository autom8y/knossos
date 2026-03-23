package explain

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/paths"
)

// contextFunc is the signature for project-aware context functions.
type contextFunc func(resolver *paths.Resolver) string

// contextFuncs maps concept names to their context injection functions.
// Concepts not in this map have no project context (inscription, tribute, sails).
var contextFuncs = map[string]contextFunc{
	"rite":     contextRite,
	"agent":    contextAgent,
	"session":  contextSession,
	"mena":     contextMena,
	"dromena":  contextDromena,
	"legomena": contextLegomena,
	"know":     contextKnow,
	"ledge":    contextLedge,
	"sos":      contextSOS,
	"knossos":  contextKnossos,
}

// GetContext returns the project-aware context string for a concept.
// Returns empty string if the concept has no context function,
// the resolver is nil, or the project root is empty.
func GetContext(name string, resolver *paths.Resolver) string {
	if resolver == nil || resolver.ProjectRoot() == "" {
		return ""
	}
	fn, ok := contextFuncs[name]
	if !ok {
		return ""
	}
	return fn(resolver)
}

func contextRite(resolver *paths.Resolver) string {
	rite := resolver.ReadActiveRite()
	if rite != "" {
		return fmt.Sprintf("Your project uses the %s rite.", rite)
	}
	return "No active rite detected."
}

func contextAgent(resolver *paths.Resolver) string {
	count := countMDFilesIn(resolver.AgentsDirForChannel(paths.ClaudeChannel{}))
	return fmt.Sprintf("Your project has %d agents defined.", count)
}

func contextSession(resolver *paths.Resolver) string {
	sessionsDir := resolver.SessionsDir()
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return "Your project has 0 sessions (0 active, 0 parked)."
	}

	var total, active, parked int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !paths.IsSessionDir(entry.Name()) {
			continue
		}
		total++
		contextPath := filepath.Join(sessionsDir, entry.Name(), "SESSION_CONTEXT.md")
		status := readSessionStatus(contextPath)
		switch status {
		case "ACTIVE":
			active++
		case "PARKED":
			parked++
		}
	}

	return fmt.Sprintf("Your project has %d sessions (%d active, %d parked).", total, active, parked)
}

func contextMena(resolver *paths.Resolver) string {
	// Default channel for explain context; explain has no channel flag.
	ch := paths.ClaudeChannel{}
	channelDir := resolver.ChannelDir(ch)
	skillsDir := filepath.Join(channelDir, "skills")
	commandsDir := filepath.Join(channelDir, "commands")
	skillCount := countAllFiles(skillsDir)
	commandCount := countAllFiles(commandsDir)
	total := skillCount + commandCount
	return fmt.Sprintf("Your project has %d mena (%d skills, %d commands).", total, skillCount, commandCount)
}

func contextDromena(resolver *paths.Resolver) string {
	ch := paths.ClaudeChannel{}
	commandsDir := filepath.Join(resolver.ChannelDir(ch), "commands")
	count := countFilesWithSuffix(commandsDir, ".dro.md")
	return fmt.Sprintf("Your project has %d dromena.", count)
}

func contextLegomena(resolver *paths.Resolver) string {
	ch := paths.ClaudeChannel{}
	skillsDir := filepath.Join(resolver.ChannelDir(ch), "skills")
	count := countFilesWithSuffix(skillsDir, ".lego.md")
	return fmt.Sprintf("Your project has %d legomena.", count)
}

func contextKnow(resolver *paths.Resolver) string {
	knowDir := filepath.Join(resolver.ProjectRoot(), ".know")
	if _, err := os.Stat(knowDir); os.IsNotExist(err) {
		return ".know/ directory not found."
	}
	count := countMDFilesIn(knowDir)
	return fmt.Sprintf("Your project has %d knowledge domains in .know/.", count)
}

func contextLedge(resolver *paths.Resolver) string {
	ledgeDir := resolver.LedgeDir()
	if _, err := os.Stat(ledgeDir); os.IsNotExist(err) {
		return ".ledge/ directory not found."
	}

	decisions := countMDFilesIn(resolver.LedgeDecisionsDir())
	specs := countMDFilesIn(resolver.LedgeSpecsDir())
	reviews := countMDFilesIn(resolver.LedgeReviewsDir())
	spikes := countMDFilesIn(resolver.LedgeSpikesDir())
	total := decisions + specs + reviews + spikes

	return fmt.Sprintf("Your project has %d artifacts in .ledge/ (%d decisions, %d specs, %d reviews, %d spikes).",
		total, decisions, specs, reviews, spikes)
}

func contextSOS(resolver *paths.Resolver) string {
	sosDir := resolver.SOSDir()
	if _, err := os.Stat(sosDir); os.IsNotExist(err) {
		return ".sos/ directory not found."
	}
	return ".sos/ directory exists with session state."
}

func contextKnossos(resolver *paths.Resolver) string {
	knossosDir := resolver.KnossosDir()
	if _, err := os.Stat(knossosDir); os.IsNotExist(err) {
		return ".knossos/ directory not found."
	}

	// Count satellite rites
	ritesDir := resolver.RitesDir()
	entries, err := os.ReadDir(ritesDir)
	if err != nil {
		return "Your project has .knossos/ with 0 satellite rites."
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(ritesDir, entry.Name(), "manifest.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			count++
		}
	}

	return fmt.Sprintf("Your project has .knossos/ with %d satellite rites.", count)
}

// --- Helper functions ---

// countMDFilesIn counts .md files in a directory.
// Returns 0 if the directory does not exist or is unreadable.
func countMDFilesIn(dir string) int {
	return countFilesWithSuffix(dir, ".md")
}

// countFilesWithSuffix counts files with a specific suffix in a directory.
// Returns 0 if the directory does not exist or is unreadable.
func countFilesWithSuffix(dir, suffix string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			count++
		}
	}
	return count
}

// countAllFiles counts all non-directory entries in a directory.
// Returns 0 if the directory does not exist or is unreadable.
func countAllFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}

// readSessionStatus reads the status field from SESSION_CONTEXT.md frontmatter.
// Returns empty string on any error.
func readSessionStatus(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	// First line must be "---"
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return ""
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			break
		}
		if value, ok := strings.CutPrefix(line, "status:"); ok {
			return strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}

	return ""
}
