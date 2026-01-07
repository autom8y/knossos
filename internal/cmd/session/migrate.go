package session

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	sess "github.com/autom8y/knossos/internal/session"
)

type migrateOptions struct {
	all    bool
	dryRun bool
}

func newMigrateCmd(ctx *cmdContext) *cobra.Command {
	var opts migrateOptions

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate sessions to v2 schema",
		Long:  `Migrates session(s) from v1 to v2.1 schema format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(ctx, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Migrate all v1 sessions")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without applying")

	return cmd
}

func runMigrate(ctx *cmdContext, opts migrateOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	var migrated []output.MigrationResult
	var skipped []output.SkipResult
	var failed []output.FailResult

	// Get sessions to migrate
	var sessionIDs []string

	if opts.all {
		// Scan all sessions
		sessionsDir := resolver.SessionsDir()
		entries, err := os.ReadDir(sessionsDir)
		if err != nil {
			if os.IsNotExist(err) {
				// No sessions directory
				result := output.MigrateOutput{
					Migrated:      migrated,
					Skipped:       skipped,
					Failed:        failed,
					TotalMigrated: 0,
					TotalSkipped:  0,
					TotalFailed:   0,
					DryRun:        opts.dryRun,
				}
				return printer.Print(result)
			}
			err := errors.Wrap(errors.CodeGeneralError, "failed to read sessions directory", err)
			printer.PrintError(err)
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() && paths.IsSessionDir(entry.Name()) {
				sessionIDs = append(sessionIDs, entry.Name())
			}
		}
	} else {
		// Get session from flag or current
		sessionID, err := ctx.GetSessionID()
		if err != nil {
			printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
			return err
		}
		if sessionID == "" {
			err := errors.New(errors.CodeUsageError, "no session specified. Use --session-id or --all")
			printer.PrintError(err)
			return err
		}
		sessionIDs = []string{sessionID}
	}

	// Process each session
	for _, sessionID := range sessionIDs {
		result, skip, fail := migrateSession(ctx, resolver, lockMgr, sessionID, opts.dryRun)
		if result != nil {
			migrated = append(migrated, *result)
		}
		if skip != nil {
			skipped = append(skipped, *skip)
		}
		if fail != nil {
			failed = append(failed, *fail)
		}
	}

	// Output result
	result := output.MigrateOutput{
		Migrated:      migrated,
		Skipped:       skipped,
		Failed:        failed,
		TotalMigrated: len(migrated),
		TotalSkipped:  len(skipped),
		TotalFailed:   len(failed),
		DryRun:        opts.dryRun,
	}

	return printer.Print(result)
}

func migrateSession(ctx *cmdContext, resolver *paths.Resolver, lockMgr *lock.Manager, sessionID string, dryRun bool) (*output.MigrationResult, *output.SkipResult, *output.FailResult) {
	ctxPath := resolver.SessionContextFile(sessionID)

	// Check if file exists
	content, err := os.ReadFile(ctxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &output.SkipResult{
				SessionID: sessionID,
				Reason:    "Session not found",
			}, nil
		}
		return nil, nil, &output.FailResult{
			SessionID: sessionID,
			Error:     err.Error(),
		}
	}

	// Parse to check version
	sessCtx, err := sess.ParseContext(content)
	if err != nil {
		return nil, nil, &output.FailResult{
			SessionID: sessionID,
			Error:     "Failed to parse context: " + err.Error(),
		}
	}

	// Check if already v2
	if sessCtx.SchemaVersion == "2.0" || sessCtx.SchemaVersion == "2.1" {
		return nil, &output.SkipResult{
			SessionID: sessionID,
			Reason:    "Already v2",
		}, nil
	}

	// Determine status from v1 fields
	derivedStatus := deriveV1Status(content)

	if dryRun {
		return &output.MigrationResult{
			SessionID:     sessionID,
			FromVersion:   sessCtx.SchemaVersion,
			ToVersion:     "2.1",
			StatusDerived: derivedStatus,
			FieldsMigrated: []string{
				"Added: schema_version=2.1",
				"Added: status=" + derivedStatus,
			},
		}, nil, nil
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout)
	if err != nil {
		return nil, nil, &output.FailResult{
			SessionID: sessionID,
			Error:     "Failed to acquire lock: " + err.Error(),
		}
	}
	defer sessionLock.Release()

	// Create backup
	backupPath := ctxPath + ".v1.backup"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return nil, nil, &output.FailResult{
			SessionID: sessionID,
			Error:     "Failed to create backup: " + err.Error(),
		}
	}

	// Update context
	sessCtx.SchemaVersion = "2.1"
	sessCtx.Status = sess.Status(derivedStatus)

	// Save migrated context
	if err := sessCtx.Save(ctxPath); err != nil {
		// Restore backup
		os.Rename(backupPath, ctxPath)
		return nil, nil, &output.FailResult{
			SessionID: sessionID,
			Error:     "Failed to save migrated context: " + err.Error(),
		}
	}

	// Emit migration event
	emitter := ctx.getEventEmitter(sessionID)
	if emitter != nil {
		emitter.EmitSchemaMigrated(sessionID, sessCtx.SchemaVersion, "2.1")
	}

	return &output.MigrationResult{
		SessionID:     sessionID,
		FromVersion:   "1.0",
		ToVersion:     "2.1",
		StatusDerived: derivedStatus,
		FieldsMigrated: []string{
			"Added: schema_version=2.1",
			"Added: status=" + derivedStatus,
		},
		BackupPath: filepath.Base(backupPath),
	}, nil, nil
}

// deriveV1Status infers status from v1 session fields.
func deriveV1Status(content []byte) string {
	str := string(content)

	// Check for archived indicators
	if strings.Contains(str, "completed_at:") || strings.Contains(str, "archived_at:") {
		return "ARCHIVED"
	}

	// Check for parked indicators
	if strings.Contains(str, "parked_at:") || strings.Contains(str, "auto_parked_at:") {
		return "PARKED"
	}

	// Default to active
	return "ACTIVE"
}

// Helper to generate timestamps for migration events
func now() time.Time {
	return time.Now().UTC()
}
