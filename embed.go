// Package knossos provides embedded rite definitions and templates
// for single-binary distribution.
//
// This file exists at the module root so that //go:embed directives can
// reference rites/, knossos/templates/, and config/ which are adjacent.
// The package is imported by cmd/ari/main.go to wire embedded assets
// into the CLI binary.
package knossos

import "embed"

// EmbeddedRites contains all rite definitions from rites/.
// Access individual rites via fs.Sub(EmbeddedRites, "rites/<name>").
//
//go:embed rites
var EmbeddedRites embed.FS

// EmbeddedTemplates contains inscription templates from knossos/templates/.
// Access section templates via fs.Sub(EmbeddedTemplates, "knossos/templates/sections").
//
//go:embed knossos/templates
var EmbeddedTemplates embed.FS

// EmbeddedHooksYAML contains the hooks configuration for single-binary distribution.
// Used by "ari init" to bootstrap config/hooks.yaml in new projects.
//
//go:embed config/hooks.yaml
var EmbeddedHooksYAML []byte
