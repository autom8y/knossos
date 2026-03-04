package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

// settableFields maps field names to their validator functions.
// Only these fields may be mutated via field-set.
var settableFields = map[string]func(string) error{
	"complexity":  validateComplexity,
	"initiative":  validateNonEmpty,
	"active_rite": validateNonEmpty,
}

// readOnlyRedirects maps read-only field names to actionable redirect messages.
// When a user attempts to set one of these fields, they receive the redirect
// message pointing them to the correct dedicated command.
var readOnlyRedirects = map[string]string{
	"current_phase":  "Use 'ari session transition <phase>' to change phase",
	"status":         "Use lifecycle commands (park/resume/wrap) to change status",
	"session_id":     "Session ID is immutable",
	"created_at":     "Timestamp is immutable",
	"schema_version": "Use 'ari session migrate' to change schema version",
}

// allReadableFields lists all fields that field-get can retrieve, in display order.
var allReadableFields = []string{
	"session_id",
	"status",
	"initiative",
	"complexity",
	"current_phase",
	"active_rite",
	"schema_version",
	"created_at",
}

type fieldGetOptions struct {
	all bool
}

func newFieldSetCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "field-set <key> <value>",
		Short: "Set a session metadata field",
		Long: `Set a validated session frontmatter field and persist it.

Settable fields:
  complexity    PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION (case-sensitive)
  initiative    Any non-empty string (max 200 chars)
  active_rite   Valid rite name (non-empty)

Read-only fields return actionable error redirects:
  current_phase   Use 'ari session transition <phase>'
  status          Use lifecycle commands (park/resume/wrap)
  session_id      Immutable
  created_at      Immutable
  schema_version  Use 'ari session migrate'

Examples:
  ari session field-set complexity SYSTEM
  ari session field-set initiative "dark mode feature"
  ari session field-set active_rite review

Context:
  Update session metadata directly without Moirai coordination.
  For phase changes, use 'ari session transition'. For status, use lifecycle commands.
  Acquires exclusive lock and emits field.updated event automatically.
  Read-only fields return actionable errors pointing to the correct command.
  Prefer this over editing SESSION_CONTEXT.md directly.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldSet(ctx, args[0], args[1])
		},
	}

	return cmd
}

func newFieldGetCmd(ctx *cmdContext) *cobra.Command {
	var opts fieldGetOptions

	cmd := &cobra.Command{
		Use:   "field-get [key]",
		Short: "Get a session metadata field",
		Long: `Get the value of a session frontmatter field.

Use --all to return all fields as a structured snapshot. Any field from
the session frontmatter is readable, including read-only fields.

Examples:
  ari session field-get complexity
  ari session field-get current_phase
  ari session field-get --all
  ari session field-get --all -o json

Context:
  Read any session field. Use --all -o json for structured snapshots.
  For richer context with timeline and decisions, prefer 'ari session status'.
  For role-scoped context injection, use 'ari session context snapshot'.
  Acquires shared lock -- safe for concurrent reads.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := ""
			if len(args) > 0 {
				key = args[0]
			}
			return runFieldGet(ctx, key, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.all, "all", false, "Return all fields as structured output")

	return cmd
}

func runFieldSet(ctx *cmdContext, key, value string) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	// Check if field is read-only — return actionable redirect message
	if redirect, ok := readOnlyRedirects[key]; ok {
		err := errors.New(errors.CodeUsageError, redirect)
		printer.PrintError(err)
		return err
	}

	// Check if field is in the settable set — validate value
	validator, ok := settableFields[key]
	if !ok {
		msg := fmt.Sprintf("unknown field %q: settable fields are complexity, initiative, active_rite", key)
		err := errors.New(errors.CodeUsageError, msg)
		printer.PrintError(err)
		return err
	}

	if err := validator(value); err != nil {
		printer.PrintError(err)
		return err
	}

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock (same pattern as transition.go)
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout, "ari-session-field-set")
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer func() { _ = sessionLock.Release() }()
	emitLockEvent(resolver, sessionID, "ari-session-field-set")

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	// Capture previous value for output
	previousValue := getField(sessCtx, key)

	// Mutate field
	setField(sessCtx, key, value)

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit field.updated event to session log (backplane-only)
	sessionDir := resolver.SessionDir(sessionID)
	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = writer.Close() }()
	writer.Write(clewcontract.Event{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Type:      "field.updated",
		Summary:   fmt.Sprintf("Field updated: %s = %s (was: %s)", key, value, previousValue),
		Meta: map[string]any{
			"session_id":     sessionID,
			"field":          key,
			"value":          value,
			"previous_value": previousValue,
		},
	})
	if err := writer.Flush(); err != nil {
		printer.VerboseLog("warn", "failed to write field.updated event", map[string]any{"error": err.Error()})
	}

	// Output result
	result := output.FieldOutput{
		Key:           key,
		Value:         value,
		PreviousValue: previousValue,
	}

	return printer.PrintSuccess(result)
}

func runFieldGet(ctx *cmdContext, key string, opts fieldGetOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	// Validate: must have key or --all, but not neither
	if key == "" && !opts.all {
		err := errors.New(errors.CodeUsageError, "specify a field key or use --all to return all fields")
		printer.PrintError(err)
		return err
	}

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	// Acquire shared lock for consistent read (non-fatal if unavailable)
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Shared, lock.DefaultTimeout, "ari-session-field-get")
	if err != nil {
		// Non-fatal — proceed without lock (same pattern as status.go)
		printer.VerboseLog("warn", "failed to acquire lock", map[string]any{"error": err.Error()})
	} else {
		defer func() { _ = sessionLock.Release() }()
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	if opts.all {
		// Return all fields as structured output
		result := output.FieldAllOutput{
			SessionID:     sessCtx.SessionID,
			Status:        string(sessCtx.Status),
			Initiative:    sessCtx.Initiative,
			Complexity:    sessCtx.Complexity,
			CurrentPhase:  sessCtx.CurrentPhase,
			ActiveRite:    sessCtx.ActiveRite,
			SchemaVersion: sessCtx.SchemaVersion,
			CreatedAt:     sessCtx.CreatedAt.Format(time.RFC3339),
		}
		return printer.Print(result)
	}

	// Validate that the key is a known field
	value, ok := getFieldOk(sessCtx, key)
	if !ok {
		msg := fmt.Sprintf("unknown field %q: valid fields are %v", key, allReadableFields)
		err := errors.New(errors.CodeUsageError, msg)
		printer.PrintError(err)
		return err
	}

	result := output.FieldOutput{
		Key:   key,
		Value: value,
	}

	return printer.Print(result)
}

// setField mutates a session context field by name.
// Only operates on fields in settableFields — callers must validate first.
func setField(ctx *sess.Context, key, value string) {
	switch key {
	case "complexity":
		ctx.Complexity = value
	case "initiative":
		ctx.Initiative = value
	case "active_rite":
		ctx.ActiveRite = value
	}
}

// getField returns the string value of a named field from the session context.
// Returns empty string for unknown keys.
func getField(ctx *sess.Context, key string) string {
	v, _ := getFieldOk(ctx, key)
	return v
}

// getFieldOk returns (value, true) for known fields, ("", false) for unknown keys.
func getFieldOk(ctx *sess.Context, key string) (string, bool) {
	switch key {
	case "session_id":
		return ctx.SessionID, true
	case "status":
		return string(ctx.Status), true
	case "initiative":
		return ctx.Initiative, true
	case "complexity":
		return ctx.Complexity, true
	case "current_phase":
		return ctx.CurrentPhase, true
	case "active_rite":
		return ctx.ActiveRite, true
	case "schema_version":
		return ctx.SchemaVersion, true
	case "created_at":
		return ctx.CreatedAt.Format(time.RFC3339), true
	default:
		return "", false
	}
}

// validateComplexity checks that the value is a recognized complexity level.
// Complexity values are case-sensitive and must be uppercase.
func validateComplexity(value string) error {
	if !isValidComplexity(value) {
		return errors.New(errors.CodeUsageError,
			"invalid complexity: must be PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION")
	}
	return nil
}

// validateNonEmpty checks that the value is not empty and does not exceed 200 chars.
func validateNonEmpty(value string) error {
	if value == "" {
		return errors.New(errors.CodeUsageError, "value must not be empty")
	}
	if len(value) > 200 {
		return errors.New(errors.CodeUsageError, "value must not exceed 200 characters")
	}
	return nil
}
