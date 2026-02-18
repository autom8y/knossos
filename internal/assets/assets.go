package assets

import "io/fs"

var (
	embeddedRites     fs.FS
	embeddedTemplates fs.FS
	embeddedHooksYAML []byte
	embeddedAgents    fs.FS
	embeddedMena      fs.FS
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

func Rites() fs.FS      { return embeddedRites }
func Templates() fs.FS  { return embeddedTemplates }
func HooksYAML() []byte { return embeddedHooksYAML }
func Agents() fs.FS     { return embeddedAgents }
func Mena() fs.FS       { return embeddedMena }
