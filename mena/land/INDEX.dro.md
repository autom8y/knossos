---
name: dion
description: "Synthesize cross-session knowledge from archived sessions via Dionysus agent. Runs ari land synthesize to inventory archives, then dispatches Dionysus to produce .sos/land/{domain}.md files."
argument-hint: "[--domain=DOMAIN] [--force]"
allowed-tools: Bash, Read, Glob, Grep, Task, AskUserQuestion
model: opus
---

# /dion -- Dionysus Knowledge Synthesis

Orchestrates the Dionysus cross-session synthesis pipeline: inventory archived sessions via `ari land synthesize`, then dispatch Dionysus to produce domain-scoped land files in `.sos/land/`.

## Context

This command runs in the main thread (requires Task tool for Dionysus dispatch). The Argus Pattern requires main-thread execution because agents cannot spawn agents -- only the main thread has Task tool access.

Dionysus is a leaf agent. It reads `.sos/archive/` and writes `.sos/land/{domain}.md`. This dromenon is the orchestrator that prepares context and invokes the agent.

## Pre-flight

### 1. Parse Arguments

- `--domain=DOMAIN`: Optional. One of `initiative-history`, `scar-tissue`, `workflow-patterns`, `all`. Default: `all`.
- `--force`: Skip confirmation when existing land files would be overwritten.

If `--domain` is present, validate it against the allowed list. If invalid: ERROR "Invalid domain '{value}'. Must be one of: initiative-history, scar-tissue, workflow-patterns, all." STOP.

### 2. Inventory Archives

Run the CLI to enumerate available archives:

```
Bash("cd {project_root} && ari land synthesize -o json")
```

Parse the JSON output. The shape is:

```json
{
  "status": "ready|no-archive|empty",
  "domain": "all",
  "sessions": [
    {
      "session_id": "session-YYYYMMDD-HHMMSS-XXXXXXXX",
      "initiative": "...",
      "complexity": "PATCH|MODULE|SYSTEM|INITIATIVE",
      "active_rite": "...",
      "created_at": "2026-03-02T11:22:13Z",
      "archived_at": "2026-03-02T11:57:17Z",
      "has_events": true,
      "events_bytes": 6107,
      "has_sails": true
    }
  ],
  "land_dir": ".sos/land",
  "existing_land_files": [
    {
      "domain": "initiative-history",
      "path": ".sos/land/initiative-history.md",
      "generated_at": "2026-03-05T18:00:00Z",
      "sessions_synthesized": 10
    }
  ],
  "message": "..."
}
```

### 3. Handle Empty States

- If `status` is `"no-archive"`: print "No archive directory found. Run `ari session wrap` to archive a completed session first." STOP.
- If `status` is `"empty"`: print "No archived sessions found. Run `ari session wrap` to archive a completed session first." STOP.
- If `sessions` array is empty (defensive check): same message as "empty". STOP.

### 4. Check Existing Land Files

If `existing_land_files` is non-empty AND `--force` is NOT set:

Print a summary of what exists:

```
Existing land files found:
  initiative-history.md (generated: 2026-03-05, sessions: 10)
  scar-tissue.md (generated: 2026-03-05, sessions: 10)

These files will be overwritten by synthesis. Proceed? (Dionysus uses full-rewrite strategy.)
```

Then use `AskUserQuestion` to confirm. If the user declines, STOP.

If `--force` is set, skip confirmation and proceed.

If `existing_land_files` is empty, proceed without confirmation.

### 5. Resolve Source Hash

```
Bash("git rev-parse --short HEAD")
```

Capture the output as `source_hash` (e.g., `ab9fcc6`).

## Execution: Dispatch Dionysus

Extract from the inventory:
- `session_count`: length of the `sessions` array
- `session_ids`: comma-separated list of all `session_id` values
- `domain`: the effective domain (`all` if no `--domain` flag)

Dispatch Dionysus via Task tool:

```
Task("dionysus", "Synthesize {domain} domains from archives in .sos/archive/. Source hash: {source_hash}. Session count: {session_count}. Sessions: {session_ids}.")
```

If `--domain` specifies a single domain (not `all`), include it:

```
Task("dionysus", "Synthesize {domain} domain from archives in .sos/archive/. Source hash: {source_hash}. Session count: {session_count}. Sessions: {session_ids}.")
```

Wait for Dionysus to return. Dionysus will provide a structured summary table with domains, file paths, session counts, confidence scores, and status.

## Post-Synthesis

### 1. Verify Output

After Dionysus returns, verify land files were written:

```
Bash("ls -la .sos/land/*.md 2>/dev/null || echo 'NO_LAND_FILES'")
```

If no land files exist after synthesis, report the error from Dionysus and STOP.

### 2. Display Summary

Print the summary returned by Dionysus (it follows a structured table format). Then append:

```
Land files written to .sos/land/

Next steps:
  - Review land files: Read(".sos/land/{domain}.md")
  - Archive stale sessions first: ari session gc --dry-run
  - Regenerate .know/ with land integration: /know --all --force
  - Run a single domain: /dion --domain=scar-tissue
```

## Error Handling

| Scenario | Action |
|----------|--------|
| `ari land synthesize` not found or fails | ERROR: "ari land synthesize failed. Ensure ari is built and on PATH. Run: CGO_ENABLED=0 go build ./cmd/ari" STOP. |
| `ari land synthesize` returns "stub" status | The installed ari binary is stale. Run: `CGO_ENABLED=0 go build -o ./ari ./cmd/ari && cp ./ari $(which ari)` |
| JSON parse failure | ERROR: "Could not parse inventory JSON. Run `ari land synthesize` manually to diagnose." STOP. |
| Dionysus Task fails | Report the error message. Suggest running with a single domain to isolate the issue. |
| Zero archives | Handled in Pre-flight step 3. |
| Land files exist without --force | Handled in Pre-flight step 4 (user confirmation). |

## Anti-Patterns

- **Running Dionysus without inventory**: Always run `ari land synthesize -o json` first. The inventory provides session count, IDs, and existing land file state that Dionysus needs for context.
- **Passing raw JSON to Dionysus**: Dionysus expects a natural-language prompt with extracted parameters (domain, source_hash, session_count, session_ids). Do NOT paste the full JSON inventory into the Task prompt.
- **Skipping post-synthesis verification**: Always verify land files exist after Dionysus returns. Synthesis can fail silently if archives have unexpected formats.
- **Dispatching multiple Dionysus instances**: Unlike the /radar Argus Pattern, Dionysus handles multi-domain synthesis internally. Dispatch ONE Task, even for `all` domains.
