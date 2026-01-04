package manifest

import (
	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/manifest"
	"github.com/autom8y/ariadne/internal/output"
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
		Long:  `Shows the manifest file contents with optional schema information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.path, "path", "", "Path to manifest file (default: .claude/manifest.json)")
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
		printer.PrintError(err)
		return err
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
func applyDefaults(content map[string]interface{}) {
	// Apply default paths if not set
	if _, ok := content["paths"]; !ok {
		content["paths"] = map[string]interface{}{}
	}
	paths, _ := content["paths"].(map[string]interface{})

	defaults := map[string]string{
		"sessions": ".claude/sessions",
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
		content["settings"] = map[string]interface{}{}
	}
	settings, _ := content["settings"].(map[string]interface{})

	settingDefaults := map[string]interface{}{
		"auto_park_on_stop": true,
		"require_session":   false,
	}
	for k, v := range settingDefaults {
		if _, ok := settings[k]; !ok {
			settings[k] = v
		}
	}
}
