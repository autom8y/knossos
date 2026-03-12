package tour

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Helpers ---

// createTestProject creates a temporary directory structure for testing.
// Returns the project root path and a cleanup function.
func createTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// channel directory
	channelDir := filepath.Join(root, paths.ClaudeChannel{}.DirName())
	require.NoError(t, os.MkdirAll(filepath.Join(channelDir, "agents"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(channelDir, "commands"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(channelDir, "skills"), 0755))

	// .knossos/ directory
	knossosDir := filepath.Join(root, ".knossos")
	require.NoError(t, os.MkdirAll(filepath.Join(knossosDir, "rites"), 0755))

	// .know/ directory
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".know"), 0755))

	// .ledge/ directory
	ledgeDir := filepath.Join(root, ".ledge")
	require.NoError(t, os.MkdirAll(filepath.Join(ledgeDir, "decisions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ledgeDir, "specs"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ledgeDir, "reviews"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ledgeDir, "spikes"), 0755))

	// .sos/ directory
	sosDir := filepath.Join(root, ".sos")
	require.NoError(t, os.MkdirAll(filepath.Join(sosDir, "sessions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sosDir, "archive"), 0755))

	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

func writeSessionContext(t *testing.T, dir, sessionID, status string) {
	t.Helper()
	sessionDir := filepath.Join(dir, sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	content := "---\nstatus: " + status + "\n---\n"
	writeFile(t, filepath.Join(sessionDir, "SESSION_CONTEXT.md"), content)
}

// --- Tour Collection Tests ---

func TestCollectTourFullProject(t *testing.T) {
	// TC-T01: Full project tour
	root := createTestProject(t)

	// Add agents
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "agents", "architect.md"), "# Architect")
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "agents", "engineer.md"), "# Engineer")
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "agents", "qa.md"), "# QA")

	// Add commands and skills
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "commands", "deploy.dro.md"), "deploy")
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "skills", "conventions.lego.md"), "conventions")
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "skills", "standards.lego.md"), "standards")

	// Add settings.json and CLAUDE.md
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "settings.json"), "{}")
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), paths.ClaudeChannel{}.ContextFile()), "# Claude")
	writeFile(t, filepath.Join(root, ".knossos", "ACTIVE_RITE"), "10x-dev")

	// Add rites with manifests
	riteDir := filepath.Join(root, ".knossos", "rites", "10x-dev")
	require.NoError(t, os.MkdirAll(riteDir, 0755))
	writeFile(t, filepath.Join(riteDir, "manifest.yaml"), "name: 10x-dev")

	// Add know domains
	writeFile(t, filepath.Join(root, ".know", "architecture.md"), "# Arch")
	writeFile(t, filepath.Join(root, ".know", "conventions.md"), "# Conv")

	// Add ledge artifacts
	writeFile(t, filepath.Join(root, ".ledge", "decisions", "adr-001.md"), "# ADR")
	writeFile(t, filepath.Join(root, ".ledge", "specs", "prd-001.md"), "# PRD")
	writeFile(t, filepath.Join(root, ".ledge", "specs", "tdd-001.md"), "# TDD")

	// Add sessions
	writeSessionContext(t, filepath.Join(root, ".sos", "sessions"), "session-20260301-120000-abcd1234", "ACTIVE")
	writeSessionContext(t, filepath.Join(root, ".sos", "sessions"), "session-20260301-130000-efgh5678", "PARKED")

	resolver := paths.NewResolver(root)
	tour := collectTour(resolver)

	assert.Equal(t, root, tour.ProjectRoot)
	assert.True(t, tour.Directories.Channel.Exists)
	assert.Equal(t, 3, tour.Directories.Channel.Agents.Count)
	assert.Equal(t, 1, tour.Directories.Channel.Commands.Count)
	assert.Equal(t, 2, tour.Directories.Channel.Skills.Count)
	assert.True(t, tour.Directories.Channel.SettingsJSON)
	assert.True(t, tour.Directories.Channel.ContextFile)
	assert.Equal(t, "10x-dev", tour.Directories.Channel.ActiveRite)

	assert.True(t, tour.Directories.Knossos.Exists)
	assert.Equal(t, 1, tour.Directories.Knossos.Rites.Count)
	assert.Equal(t, []string{"10x-dev"}, tour.Directories.Knossos.Rites.Items)

	assert.True(t, tour.Directories.Know.Exists)
	assert.Equal(t, 2, tour.Directories.Know.Domains.Count)

	assert.True(t, tour.Directories.Ledge.Exists)
	assert.Equal(t, 1, tour.Directories.Ledge.Decisions.Count)
	assert.Equal(t, 2, tour.Directories.Ledge.Specs.Count)
	assert.Equal(t, 0, tour.Directories.Ledge.Reviews.Count)
	assert.Equal(t, 0, tour.Directories.Ledge.Spikes.Count)

	assert.True(t, tour.Directories.SOS.Exists)
	assert.Equal(t, 2, tour.Directories.SOS.Sessions.Count)
	assert.Equal(t, 1, tour.Directories.SOS.Sessions.Active)
	assert.Equal(t, 1, tour.Directories.SOS.Sessions.Parked)
}

func TestCollectTourEmptyProject(t *testing.T) {
	// TC-T02: Empty project (.claude only)
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, paths.ClaudeChannel{}.DirName()), 0755))

	resolver := paths.NewResolver(root)
	tour := collectTour(resolver)

	assert.True(t, tour.Directories.Channel.Exists)
	assert.Equal(t, 0, tour.Directories.Channel.Agents.Count)
	assert.False(t, tour.Directories.Knossos.Exists)
	assert.False(t, tour.Directories.Know.Exists)
	assert.False(t, tour.Directories.Ledge.Exists)
	assert.False(t, tour.Directories.SOS.Exists)
}

func TestCollectTourMissingKnossos(t *testing.T) {
	// TC-T03: Missing .knossos
	root := createTestProject(t)
	require.NoError(t, os.RemoveAll(filepath.Join(root, ".knossos")))

	resolver := paths.NewResolver(root)
	tour := collectTour(resolver)

	assert.True(t, tour.Directories.Channel.Exists)
	assert.False(t, tour.Directories.Knossos.Exists)
	assert.True(t, tour.Directories.Know.Exists)
}

func TestCollectTourMissingKnow(t *testing.T) {
	// TC-T04: Missing .know
	root := createTestProject(t)
	require.NoError(t, os.RemoveAll(filepath.Join(root, ".know")))

	resolver := paths.NewResolver(root)
	tour := collectTour(resolver)

	assert.True(t, tour.Directories.Channel.Exists)
	assert.False(t, tour.Directories.Know.Exists)
	assert.True(t, tour.Directories.Ledge.Exists)
}

func TestCollectTourMissingLedge(t *testing.T) {
	// TC-T05: Missing .ledge
	root := createTestProject(t)
	require.NoError(t, os.RemoveAll(filepath.Join(root, ".ledge")))

	resolver := paths.NewResolver(root)
	tour := collectTour(resolver)

	assert.True(t, tour.Directories.Channel.Exists)
	assert.False(t, tour.Directories.Ledge.Exists)
	assert.True(t, tour.Directories.SOS.Exists)
}

func TestCollectTourMissingSOS(t *testing.T) {
	// TC-T06: Missing .sos
	root := createTestProject(t)
	require.NoError(t, os.RemoveAll(filepath.Join(root, ".sos")))

	resolver := paths.NewResolver(root)
	tour := collectTour(resolver)

	assert.True(t, tour.Directories.Channel.Exists)
	assert.False(t, tour.Directories.SOS.Exists)
}

// --- Claude Section Tests ---

func TestClaudeAgentCount(t *testing.T) {
	// TC-T07: Agent count
	root := createTestProject(t)
	agentsDir := filepath.Join(root, paths.ClaudeChannel{}.DirName(), "agents")
	writeFile(t, filepath.Join(agentsDir, "a.md"), "agent")
	writeFile(t, filepath.Join(agentsDir, "b.md"), "agent")
	writeFile(t, filepath.Join(agentsDir, "c.md"), "agent")
	writeFile(t, filepath.Join(agentsDir, "readme.txt"), "not an agent")

	resolver := paths.NewResolver(root)
	section := collectChannel(resolver)
	assert.Equal(t, 3, section.Agents.Count)
}

func TestClaudeSettingsJSONPresent(t *testing.T) {
	// TC-T08: settings.json present
	root := createTestProject(t)
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), "settings.json"), "{}")

	resolver := paths.NewResolver(root)
	section := collectChannel(resolver)
	assert.True(t, section.SettingsJSON)
}

func TestClaudeSettingsJSONMissing(t *testing.T) {
	// TC-T09: settings.json missing
	root := createTestProject(t)

	resolver := paths.NewResolver(root)
	section := collectChannel(resolver)
	assert.False(t, section.SettingsJSON)
}

func TestClaudeMDPresent(t *testing.T) {
	// TC-T10: CLAUDE.md present
	root := createTestProject(t)
	writeFile(t, filepath.Join(root, paths.ClaudeChannel{}.DirName(), paths.ClaudeChannel{}.ContextFile()), "# Claude")

	resolver := paths.NewResolver(root)
	section := collectChannel(resolver)
	assert.True(t, section.ContextFile)
}

func TestClaudeActiveRiteValue(t *testing.T) {
	// TC-T11: ACTIVE_RITE value
	root := createTestProject(t)
	writeFile(t, filepath.Join(root, ".knossos", "ACTIVE_RITE"), "10x-dev")

	resolver := paths.NewResolver(root)
	section := collectChannel(resolver)
	assert.Equal(t, "10x-dev", section.ActiveRite)
}

func TestClaudeNoActiveRite(t *testing.T) {
	// TC-T12: No ACTIVE_RITE
	root := createTestProject(t)

	resolver := paths.NewResolver(root)
	section := collectChannel(resolver)
	assert.Empty(t, section.ActiveRite)
}

// --- Knossos Section Tests ---

func TestKnossosRiteListing(t *testing.T) {
	// TC-T13: Rite listing
	root := createTestProject(t)
	ritesDir := filepath.Join(root, ".knossos", "rites")

	// Two valid rites (with manifest.yaml)
	rite1 := filepath.Join(ritesDir, "alpha")
	require.NoError(t, os.MkdirAll(rite1, 0755))
	writeFile(t, filepath.Join(rite1, "manifest.yaml"), "name: alpha")

	rite2 := filepath.Join(ritesDir, "beta")
	require.NoError(t, os.MkdirAll(rite2, 0755))
	writeFile(t, filepath.Join(rite2, "manifest.yaml"), "name: beta")

	// One invalid directory (no manifest.yaml)
	rite3 := filepath.Join(ritesDir, "gamma")
	require.NoError(t, os.MkdirAll(rite3, 0755))

	resolver := paths.NewResolver(root)
	section := collectKnossos(resolver)
	assert.Equal(t, 2, section.Rites.Count)
	assert.Equal(t, []string{"alpha", "beta"}, section.Rites.Items)
}

func TestKnossosTemplatesCount(t *testing.T) {
	// TC-T14: Templates count
	root := createTestProject(t)
	templatesDir := filepath.Join(root, ".knossos", "templates")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	writeFile(t, filepath.Join(templatesDir, "a.yaml"), "template")
	writeFile(t, filepath.Join(templatesDir, "b.yaml"), "template")
	writeFile(t, filepath.Join(templatesDir, "c.yaml"), "template")

	resolver := paths.NewResolver(root)
	section := collectKnossos(resolver)
	assert.Equal(t, 3, section.Templates.Count)
}

// --- Know Section Tests ---

func TestKnowDomainFileListing(t *testing.T) {
	// TC-T15: Domain file listing
	root := createTestProject(t)
	knowDir := filepath.Join(root, ".know")
	writeFile(t, filepath.Join(knowDir, "architecture.md"), "# Arch")
	writeFile(t, filepath.Join(knowDir, "conventions.md"), "# Conv")
	writeFile(t, filepath.Join(knowDir, "scar-tissue.md"), "# Scar")

	resolver := paths.NewResolver(root)
	section := collectKnow(resolver)
	assert.Equal(t, 3, section.Domains.Count)
	assert.Equal(t, []string{"architecture", "conventions", "scar-tissue"}, section.Domains.Items)
}

// --- Ledge Section Tests ---

func TestLedgePerSubdirectoryCounts(t *testing.T) {
	// TC-T16: Per-subdirectory counts
	root := createTestProject(t)
	writeFile(t, filepath.Join(root, ".ledge", "decisions", "adr-001.md"), "ADR")
	writeFile(t, filepath.Join(root, ".ledge", "decisions", "adr-002.md"), "ADR")
	writeFile(t, filepath.Join(root, ".ledge", "specs", "prd-001.md"), "PRD")

	resolver := paths.NewResolver(root)
	section := collectLedge(resolver)
	assert.Equal(t, 2, section.Decisions.Count)
	assert.Equal(t, 1, section.Specs.Count)
	assert.Equal(t, 0, section.Reviews.Count)
	assert.Equal(t, 0, section.Spikes.Count)
}

// --- SOS Section Tests ---

func TestSOSSessionBreakdown(t *testing.T) {
	// TC-T17: Session breakdown
	root := createTestProject(t)
	sessionsDir := filepath.Join(root, ".sos", "sessions")

	writeSessionContext(t, sessionsDir, "session-20260301-120000-abcd1234", "ACTIVE")
	writeSessionContext(t, sessionsDir, "session-20260301-130000-efgh5678", "ACTIVE")
	writeSessionContext(t, sessionsDir, "session-20260301-140000-ijkl9012", "PARKED")

	resolver := paths.NewResolver(root)
	section := collectSOS(resolver)
	assert.Equal(t, 3, section.Sessions.Count)
	assert.Equal(t, 2, section.Sessions.Active)
	assert.Equal(t, 1, section.Sessions.Parked)
}

func TestSOSArchiveCount(t *testing.T) {
	// TC-T18: Archive count
	root := createTestProject(t)
	archiveDir := filepath.Join(root, ".sos", "archive")

	for _, name := range []string{
		"session-20260201-120000-aaaa1111",
		"session-20260202-120000-bbbb2222",
		"session-20260203-120000-cccc3333",
	} {
		require.NoError(t, os.MkdirAll(filepath.Join(archiveDir, name), 0755))
	}

	resolver := paths.NewResolver(root)
	section := collectSOS(resolver)
	assert.Equal(t, 3, section.Archive.Count)
}

// --- Interface Compliance Tests ---

func TestTourOutputImplementsTextable(t *testing.T) {
	// TC-T19: TourOutput implements Textable
	var _ output.Textable = TourOutput{}
}

// --- Text Output Tests ---

func TestTourTextStartsWithHeader(t *testing.T) {
	// TC-T20: Text starts with header
	tour := TourOutput{
		ProjectRoot: "/tmp/test",
		Directories: TourDirectories{
			Channel: ChannelSection{Exists: true, Path: "channel/"},
		},
	}
	text := tour.Text()
	assert.Contains(t, text, "=== Project Tour ===")
}

func TestTourTextNotFoundDirectories(t *testing.T) {
	// TC-T21: Not-found directories show marker
	tour := TourOutput{
		ProjectRoot: "/tmp/test",
		Directories: TourDirectories{
			Channel:  ChannelSection{Exists: false, Path: "channel/"},
			Knossos: KnossosSection{Exists: false, Path: ".knossos/"},
			Know:    KnowSection{Exists: false, Path: ".know/"},
			Ledge:   LedgeSection{Exists: false, Path: ".ledge/"},
			SOS:     SOSSection{Exists: false, Path: ".sos/"},
		},
	}
	text := tour.Text()
	// Count occurrences of "(not found)"
	count := 0
	for i := 0; i < len(text); i++ {
		if i+11 <= len(text) && text[i:i+11] == "(not found)" {
			count++
		}
	}
	assert.Equal(t, 5, count, "expected 5 (not found) markers")
}

func TestTourTextFileCountsRendered(t *testing.T) {
	// TC-T22: File counts rendered
	tour := TourOutput{
		ProjectRoot: "/tmp/test",
		Directories: TourDirectories{
			Channel: ChannelSection{
				Exists: true,
				Path:   paths.ClaudeChannel{}.DirName() + "/",
				Agents: DirCount{Count: 5},
			},
		},
	}
	text := tour.Text()
	assert.Contains(t, text, "5 agents")
}

func TestTourTextPluralizeFile(t *testing.T) {
	assert.Equal(t, "1 file", pluralizeFile(1))
	assert.Equal(t, "0 files", pluralizeFile(0))
	assert.Equal(t, "3 files", pluralizeFile(3))
}
