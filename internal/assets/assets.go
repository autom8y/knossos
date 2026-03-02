package assets

import "io/fs"

var (
	embeddedRites     fs.FS
	embeddedTemplates fs.FS
	embeddedHooksYAML []byte
	embeddedAgents    fs.FS
	embeddedMena      fs.FS
	buildVersion      string
)

func SetEmbedded(rites, templates fs.FS, hooksYAML []byte) {
	embeddedRites = rites
	embeddedTemplates = templates
	embeddedHooksYAML = hooksYAML
}

func SetUserAssets(agents, mena fs.FS) {
	embeddedAgents = agents
	embeddedMena = mena
}

// SetBuildVersion stores the binary version string for use during XDG extraction.
// Called from main before any commands execute.
func SetBuildVersion(v string) { buildVersion = v }

// BuildVersion returns the binary version string set via SetBuildVersion.
// Returns "dev" if not set.
func BuildVersion() string {
	if buildVersion == "" {
		return "dev"
	}
	return buildVersion
}

func Rites() fs.FS      { return embeddedRites }
func Templates() fs.FS  { return embeddedTemplates }
func HooksYAML() []byte { return embeddedHooksYAML }
func Agents() fs.FS     { return embeddedAgents }
func Mena() fs.FS       { return embeddedMena }
