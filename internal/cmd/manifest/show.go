package manifest

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/manifest"
	"github.com/autom8y/knossos/internal/output"
)

type showOptions struct {
	path     string
	schema   bool
	resolved bool
}

func newShowCmd(ctx *cmdContext) *cobra.Command {
	var opts showOptions

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display current effective manifest",
		Long: `Shows the manifest file contents with optional schema information.

Examples:
  ari manifest show
  ari manifest show --schema
  ari manifest show --resolved
  ari manifest show --path {channel}/manifest.json -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.path, "path", "", "Path to manifest file (default: {channel}/manifest.json)")
	cmd.Flags().BoolVar(&opts.schema, "schema", false, "Include schema version and validation status")
	cmd.Flags().BoolVar(&opts.resolved, "resolved", false, "Show resolved values (with defaults applied)")

	return cmd
}

func runShow(ctx *cmdContext, opts showOptions) error {
	printer := ctx.getPrinter()

	// Determine manifest path
	manifestPath := opts.path
	if manifestPath == "" {
		manifestPath = ctx.defaultManifestPath()
	}

	// Try to load the manifest
	m, err := manifest.Load(manifestPath)
	if err != nil {
		// Check if file not found - return graceful response
		if errors.IsNotFound(err) {
			out := output.ManifestShowOutput{
				Path:   manifestPath,
				Exists: false,
			}
			return printer.Print(out)
		}
		return common.PrintAndReturn(printer, err)
	}

	// Build output
	out := output.ManifestShowOutput{
		Path:    m.Path,
		Exists:  true,
		Format:  string(m.Format),
		Content: m.Content,
	}

	// Add schema info if requested
	if opts.schema {
		schemaInfo, err := manifest.GetSchemaInfo(m)
		if err == nil {
			out.Schema = &output.ManifestSchemaInfo{
				Type:    schemaInfo.Type,
				Version: schemaInfo.Version,
				Valid:   true, // Assume valid unless we validate
			}

			// Validate if schema info available
			validator, valErr := ctx.getSchemaValidator()
			if valErr == nil && schemaInfo.Type != "" {
				result, valErr := validator.Validate(m, schemaInfo.Type, false)
				if valErr == nil {
					out.Schema.Valid = result.Valid
				}
			}
		}
	}

	// Apply defaults if resolved requested
	if opts.resolved {
		applyDefaults(out.Content)
	}

	return printer.Print(out)
}

// applyDefaults applies default values to manifest content.
func applyDefaults(content map[string]any) {
	// Apply default paths if not set
	if _, ok := content["paths"]; !ok {
		content["paths"] = map[string]any{}
	}
	paths, _ := content["paths"].(map[string]any)

	// CC-channel-specific wire defaults. These are string constants consumed by
	// the CC harness manifest format, not resolved paths. Other channels (Gemini,
	// Codex) use different manifest schemas and do not share these defaults.
	defaults := map[string]string{
		"sessions": ".sos/sessions",
		"agents":   ".claude/agents",
		"skills":   ".claude/skills",
		"hooks":    ".claude/hooks",
	}
	for k, v := range defaults {
		if _, ok := paths[k]; !ok {
			paths[k] = v
		}
	}

	// Apply default settings if not set
	if _, ok := content["settings"]; !ok {
		content["settings"] = map[string]any{}
	}
	settings, _ := content["settings"].(map[string]any)

	settingDefaults := map[string]any{
		"auto_park_on_stop": true,
		"require_session":   false,
	}
	for k, v := range settingDefaults {
		if _, ok := settings[k]; !ok {
			settings[k] = v
		}
	}
}
