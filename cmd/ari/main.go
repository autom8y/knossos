// Package main is the entry point for the ari CLI.
// It contains minimal logic - all command implementations are in internal/cmd/.
package main

import (
	"fmt"
	"os"

	"github.com/autom8y/knossos"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/cmd/root"
	"github.com/autom8y/knossos/internal/errors"
)

// Version information set at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root.SetVersion(version, commit, date)
	common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)
	common.SetEmbeddedUserAssets(knossos.EmbeddedAgents, knossos.EmbeddedMena)
	if err := root.Execute(); err != nil {
		// Print error to stderr (SilenceErrors is enabled on root cmd)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(errors.GetExitCode(err))
	}
}
