package tour

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/autom8y/knossos/internal/paths"
)

// collectTour gathers directory data from all five managed directories.
func collectTour(resolver *paths.Resolver) TourOutput {
	return TourOutput{
		ProjectRoot: resolver.ProjectRoot(),
		Directories: TourDirectories{
			Channel: collectChannel(resolver),
			Knossos: collectKnossos(resolver),
			Know:    collectKnow(resolver),
			Ledge:   collectLedge(resolver),
			SOS:     collectSOS(resolver),
		},
	}
}

func collectChannel(resolver *paths.Resolver) ChannelSection {
	channelDir := resolver.ClaudeDir()
	if !dirExists(channelDir) {
		return ChannelSection{Exists: false, Path: ".claude/"}
	}

	section := ChannelSection{
		Exists: true,
		Path:   ".claude/",
	}

	// Count .md files in agents/
	agentsDir := resolver.AgentsDir()
	section.Agents = DirCount{
		Count: countFilesWithSuffix(agentsDir, ".md"),
		Items: listFiles(agentsDir),
	}

	// Count all files in commands/
	commandsDir := filepath.Join(channelDir, "commands")
	section.Commands = DirCount{
		Count: countAllFiles(commandsDir),
	}

	// Count all files in skills/
	skillsDir := filepath.Join(channelDir, "skills")
	section.Skills = DirCount{
		Count: countAllFiles(skillsDir),
	}

	// Check settings.json existence
	section.SettingsJSON = fileExists(filepath.Join(channelDir, "settings.json"))

	// Check CLAUDE.md existence
	section.ClaudeMD = fileExists(filepath.Join(channelDir, "CLAUDE.md"))

	// Read ACTIVE_RITE value
	section.ActiveRite = resolver.ReadActiveRite()

	return section
}

func collectKnossos(resolver *paths.Resolver) KnossosSection {
	knossosDir := resolver.KnossosDir()
	if !dirExists(knossosDir) {
		return KnossosSection{Exists: false, Path: ".knossos/"}
	}

	section := KnossosSection{
		Exists: true,
		Path:   ".knossos/",
	}

	// List rite directories (those with manifest.yaml)
	ritesDir := resolver.RitesDir()
	entries, err := os.ReadDir(ritesDir)
	if err == nil {
		var riteNames []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			manifestPath := filepath.Join(ritesDir, entry.Name(), "manifest.yaml")
			if _, err := os.Stat(manifestPath); err == nil {
				riteNames = append(riteNames, entry.Name())
			}
		}
		sort.Strings(riteNames)
		section.Rites = DirCount{
			Count: len(riteNames),
			Items: riteNames,
		}
	}

	// Count files in templates/ (if exists)
	templatesDir := filepath.Join(knossosDir, "templates")
	if dirExists(templatesDir) {
		section.Templates = DirCount{
			Count: countAllFiles(templatesDir),
		}
	}

	return section
}

func collectKnow(resolver *paths.Resolver) KnowSection {
	knowDir := filepath.Join(resolver.ProjectRoot(), ".know")
	if !dirExists(knowDir) {
		return KnowSection{Exists: false, Path: ".know/"}
	}

	domainNames := listFilesWithSuffix(knowDir, ".md")
	return KnowSection{
		Exists: true,
		Path:   ".know/",
		Domains: DirCount{
			Count: len(domainNames),
			Items: domainNames,
		},
	}
}

func collectLedge(resolver *paths.Resolver) LedgeSection {
	ledgeDir := resolver.LedgeDir()
	if !dirExists(ledgeDir) {
		return LedgeSection{Exists: false, Path: ".ledge/"}
	}

	return LedgeSection{
		Exists: true,
		Path:   ".ledge/",
		Decisions: DirCount{
			Count: countFilesWithSuffix(resolver.LedgeDecisionsDir(), ".md"),
		},
		Specs: DirCount{
			Count: countFilesWithSuffix(resolver.LedgeSpecsDir(), ".md"),
		},
		Reviews: DirCount{
			Count: countFilesWithSuffix(resolver.LedgeReviewsDir(), ".md"),
		},
		Spikes: DirCount{
			Count: countFilesWithSuffix(resolver.LedgeSpikesDir(), ".md"),
		},
	}
}

func collectSOS(resolver *paths.Resolver) SOSSection {
	sosDir := resolver.SOSDir()
	if !dirExists(sosDir) {
		return SOSSection{Exists: false, Path: ".sos/"}
	}

	section := SOSSection{
		Exists: true,
		Path:   ".sos/",
	}

	// Count session directories
	sessionsDir := resolver.SessionsDir()
	if entries, err := os.ReadDir(sessionsDir); err == nil {
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
		section.Sessions = SOSSessions{
			Count:  total,
			Active: active,
			Parked: parked,
		}
	}

	// Count archived sessions
	archiveDir := resolver.ArchiveDir()
	if entries, err := os.ReadDir(archiveDir); err == nil {
		count := 0
		for _, entry := range entries {
			if entry.IsDir() && paths.IsSessionDir(entry.Name()) {
				count++
			}
		}
		section.Archive = DirCount{Count: count}
	}

	return section
}

// --- Helper functions ---

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

// listFiles returns file names (without directory path) in a directory.
// Returns nil if the directory does not exist or is unreadable.
func listFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

// listFilesWithSuffix returns file names matching a suffix, with suffix stripped.
// Returns nil if the directory does not exist or is unreadable.
func listFilesWithSuffix(dir, suffix string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			names = append(names, strings.TrimSuffix(e.Name(), suffix))
		}
	}
	sort.Strings(names)
	return names
}

// dirExists returns true if a path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// fileExists returns true if a path exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
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
