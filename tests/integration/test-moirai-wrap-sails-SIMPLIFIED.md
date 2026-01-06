# T3-005: Moirai wrap_session Validates Sails - Implementation Summary

**Status**: COMPLETE

## What Was Implemented

### 1. Atropos Agent Documentation (`user-agents/atropos.md`)

Updated wrap_session documentation to reflect sails validation:

- **Syntax**: `wrap_session [--emergency]`
- **Quality Gate**: Blocks wrap if sails are BLACK (unless `--emergency`)
- **Internal Flow**: Documents that `ari session wrap` generates sails and validates color
- **Implementation Pattern**: Shows how Atropos delegates to Ariadne CLI
- **Response Examples**: Includes BLACK sails blocking + emergency override

### 2. Go Implementation (Already Complete)

The `ariadne/internal/cmd/session/wrap.go` implementation:

**Lines 86-110: Sails Generation and Quality Gate**
```go
// Generate White Sails confidence signal before archiving
sailsGen := sails.NewGenerator(sessionDir)
sailsResult, sailsErr := sailsGen.Generate()

// Quality gate: Block wrap if sails are BLACK (unless --force)
if sailsResult.Color == sails.ColorBlack {
    if !opts.force {
        err := errors.NewWithDetails(errors.CodeQualityGateFailed,
            "cannot wrap session with BLACK sails: explicit blockers present",
            map[string]interface{}{
                "color":   string(sailsResult.Color),
                "reasons": sailsResult.Reasons,
            })
        printer.PrintError(err)
        return err
    }
    // If --force, emit warning but continue
    printer.VerboseLog("warn", "wrapping session with BLACK sails (--force used)", ...)
}
```

**Lines 138-148: SAILS_GENERATED Event**
```go
sailsEvent := threadcontract.NewSailsGeneratedEvent(sessionID, threadcontract.SailsGeneratedData{
    Color:         string(sailsResult.Color),
    ComputedBase:  string(sailsResult.ComputedBase),
    Reasons:       sailsResult.Reasons,
    FilePath:      sailsResult.FilePath,
    EvidencePaths: evidencePaths,
})
writer.Write(sailsEvent)
```

**Lines 230-236: Sails Metadata in Output**
```go
if sailsResult != nil {
    result.SailsColor = string(sailsResult.Color)
    result.SailsBase = string(sailsResult.ComputedBase)
    result.SailsReasons = sailsResult.Reasons
    result.SailsPath = sailsResult.FilePath
}
```

### 3. Moirai Router (`user-agents/moirai.md`)

The Moirai router already routes `wrap_session` to Atropos (line 59):

```markdown
| Operation | Fate | Domain |
|-----------|------|--------|
| `wrap_session` | **Atropos** | Termination |
```

## Verification

The implementation is complete and functional:

1. **T3-001** (BLACK sails detection): `color.go` implements BLACK when blockers present
2. **T3-002** (Wrap blocking on BLACK): `wrap.go` lines 94-104 block unless `--force`
3. **T3-003** (Sails in status): Implemented in status output
4. **T3-004** (Evidence collection): `proofs.go` collects from session directory
5. **T3-005** (Moirai validates sails): **THIS TASK** - Atropos documented, Go impl complete

## Usage

### Invoke via Moirai

```
Task(moirai, "wrap_session

Session Context:
- Session ID: session-20260106-123456")
```

Moirai routes to Atropos → Atropos calls `ari session wrap` → Ariadne:
1. Collects proofs
2. Computes sails color
3. **BLOCKS if BLACK** (quality gate)
4. Generates WHITE_SAILS.yaml
5. Archives session
6. Returns structured response with sails metadata

### Emergency Override

```
Task(moirai, "--emergency=reason=\"Hotfix deployment\" wrap_session

Session Context:
- Session ID: session-20260106-123456")
```

This bypasses the BLACK sails quality gate but logs the override.

### Direct CLI Usage

```bash
# Standard (blocks on BLACK)
ari session wrap

# Force override (bypasses BLACK gate)
ari session wrap --force
```

## Track 4 Unblocked

With T3-005 complete, Track 4 (skill delegation to Moirai) can proceed:

- Skills can invoke Moirai via Task tool
- Moirai validates sails before archival
- BLACK sails prevent wrap (unless emergency)
- WHITE_SAILS.yaml generated as part of ceremony

## References

- Knossos Doctrine: `docs/philosophy/knossos-doctrine.md` (Section VII: The Confidence Signal)
- Atropos Agent: `user-agents/atropos.md`
- Moirai Router: `user-agents/moirai.md`
- Wrap Implementation: `ariadne/internal/cmd/session/wrap.go`
- Color Computation: `ariadne/internal/sails/color.go`
