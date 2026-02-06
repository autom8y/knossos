package agent

import "embed"

//go:embed templates/*.md.tpl
var templateFS embed.FS
