// Package main is the entry point for the ari CLI.
// It contains minimal logic - all command implementations are in internal/cmd/.
package main

import (
	"os"

	"github.com/autom8y/knossos"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/cmd/root"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
)

// Version information set at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root.SetVersion(version, commit, date)
	common.SetBuildVersion(version)
	common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)
	common.SetEmbeddedUserAssets(knossos.EmbeddedAgents, knossos.EmbeddedMena)
	common.SetEmbeddedProcessions(knossos.EmbeddedProcessions)
	if err := root.Execute(); err != nil {
		if !errors.IsHandled(err) {
			// Error was not already printed by a command — print it format-aware
			format := output.ParseFormat(root.GetOutputFormat())
			printer := output.NewPrinter(format, os.Stdout, os.Stderr, false)
			printer.PrintError(err)
		}
		os.Exit(errors.GetExitCode(err))
	}
}
