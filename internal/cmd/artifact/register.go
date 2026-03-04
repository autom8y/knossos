package artifact

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/errors"
)

func newRegisterCmd(ctx *cmdContext) *cobra.Command {
	var (
		path           string
		sessionID      string
		taskID         string
		specialist     string
		phase          string
		skipValidation bool
		skipAggregate  bool
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register an artifact to the session registry",
		Long: `Register an artifact to the session registry and optionally aggregate to project registry.

The artifact type and ID are automatically detected from the filename.
Phase and specialist can be specified explicitly or inferred from context.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			registry := ctx.getRegistry()
			aggregator := ctx.getAggregator()
			resolver := ctx.GetResolver()

			// Determine session ID
			sid := sessionID
			if sid == "" {
				var err error
				sid, err = ctx.GetSessionID()
				if err != nil {
					printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
					return err
				}
				if sid == "" {
					err := errors.New(errors.CodeUsageError, "no active session and --session not specified")
					printer.PrintError(err)
					return err
				}
			}

			// Verify file exists
			absPath := path
			if !filepath.IsAbs(path) {
				absPath = filepath.Join(resolver.ProjectRoot(), path)
			}
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				e := errors.NewWithDetails(errors.CodeFileNotFound,
					"artifact file not found",
					map[string]any{"path": path})
				printer.PrintError(e)
				return e
			}

			// Detect artifact type and ID from filename
			artifactType, artifactID, err := detectArtifact(filepath.Base(absPath))
			if err != nil {
				printer.PrintError(err)
				return err
			}

			// Determine phase (default from artifact type if not specified)
			phaseValue := artifact.Phase(phase)
			if phase == "" {
				phaseValue = inferPhaseFromType(artifactType)
			}

			// Determine specialist (default: "unknown" if not specified)
			specialistValue := specialist
			if specialist == "" {
				specialistValue = "unknown"
			}

			// Create entry
			entry := artifact.Entry{
				ArtifactID:       artifactID,
				ArtifactType:     artifactType,
				Path:             path,
				Phase:            phaseValue,
				Specialist:       specialistValue,
				SessionID:        sid,
				TaskID:           taskID,
				RegisteredAt:     time.Now().UTC(),
				Validated:        !skipValidation,
				ValidationIssues: []string{},
			}

			// Register to session
			if err := registry.Register(sid, entry); err != nil {
				printer.PrintError(err)
				return err
			}

			// Aggregate to project unless --skip-aggregate
			aggregated := false
			if !skipAggregate {
				if err := aggregator.AggregateSession(sid); err != nil {
					// Log warning but don't fail
					printer.VerboseLog("warn", "aggregation failed", map[string]any{"error": err.Error()})
				} else {
					aggregated = true
				}
			}

			// Print result
			result := map[string]any{
				"artifact_id":   entry.ArtifactID,
				"artifact_type": entry.ArtifactType,
				"path":          entry.Path,
				"phase":         entry.Phase,
				"specialist":    entry.Specialist,
				"session_id":    entry.SessionID,
				"task_id":       entry.TaskID,
				"registered_at": entry.RegisteredAt.Format(time.RFC3339),
				"validated":     entry.Validated,
				"aggregated":    aggregated,
			}
			_ = printer.Print(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Path to artifact file (required)")
	cmd.Flags().StringVar(&sessionID, "session", "", "Session ID (default: current session)")
	cmd.Flags().StringVar(&taskID, "task", "", "Task ID that produced this artifact")
	cmd.Flags().StringVar(&specialist, "specialist", "", "Agent name (default: unknown)")
	cmd.Flags().StringVar(&phase, "phase", "", "Workflow phase (default: inferred from type)")
	cmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "Skip schema validation")
	cmd.Flags().BoolVar(&skipAggregate, "skip-aggregate", false, "Don't trigger project aggregation")

	_ = cmd.MarkFlagRequired("path")

	return cmd
}

// detectArtifact detects artifact type and ID from filename.
// Expected patterns: PRD-*, TDD-*, ADR-*, TEST-*, etc.
func detectArtifact(filename string) (artifact.ArtifactType, string, error) {
	// Remove extension
	base := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Pattern: TYPE-artifact-id
	artifactIDPattern := regexp.MustCompile(`^([A-Z]+)-([a-z0-9-]+)$`)
	matches := artifactIDPattern.FindStringSubmatch(base)
	if matches == nil {
		return "", "", errors.NewWithDetails(errors.CodeUsageError,
			"invalid artifact filename pattern (expected: TYPE-artifact-id.ext)",
			map[string]any{"filename": filename})
	}

	typePrefix := matches[1]
	artifactID := base

	var artifactType artifact.ArtifactType
	switch typePrefix {
	case "PRD":
		artifactType = artifact.TypePRD
	case "TDD":
		artifactType = artifact.TypeTDD
	case "ADR":
		artifactType = artifact.TypeADR
	case "TEST":
		artifactType = artifact.TypeTestPlan
	case "CODE":
		artifactType = artifact.TypeCode
	case "RUNBOOK":
		artifactType = artifact.TypeRunbook
	default:
		return "", "", errors.NewWithDetails(errors.CodeUsageError,
			"unknown artifact type prefix",
			map[string]any{"prefix": typePrefix, "filename": filename})
	}

	return artifactType, artifactID, nil
}

// inferPhaseFromType infers the workflow phase from artifact type.
func inferPhaseFromType(t artifact.ArtifactType) artifact.Phase {
	switch t {
	case artifact.TypePRD:
		return artifact.PhaseRequirements
	case artifact.TypeTDD, artifact.TypeADR:
		return artifact.PhaseDesign
	case artifact.TypeCode:
		return artifact.PhaseImplementation
	case artifact.TypeTestPlan:
		return artifact.PhaseValidation
	default:
		return artifact.PhaseImplementation
	}
}
