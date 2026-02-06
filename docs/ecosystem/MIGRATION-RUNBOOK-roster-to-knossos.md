# Migration Runbook: Roster to Knossos Rename

**Version**: 1.0.0
**Date**: 2026-02-06
**Applies to**: All satellite projects using the Knossos platform (formerly "roster")
**Migration type**: CLEAN BREAK -- no backward compatibility shims
**Estimated time**: 15 minutes for first satellite, 5 minutes each additional

---

## What Changed

The platform formerly known as "roster" has been fully renamed to "knossos". This is a
clean break with no backward compatibility layer. All references to "roster" in code,
configuration, environment variables, and manifests must be updated.

### Breaking Changes Summary

| Category | Old Value | New Value |
|----------|-----------|-----------|
| Default path | `~/Code/roster` | `~/Code/knossos` |
| Primary env var | `ROSTER_HOME` | `KNOSSOS_HOME` |
| Debug env var | `ROSTER_SYNC_DEBUG` | `KNOSSOS_SYNC_DEBUG` |
| Verbose env var | `ROSTER_VERBOSE` | `KNOSSOS_VERBOSE` |
| Dir env var | `ROSTER_DIR` | `KNOSSOS_DIR` |
| Version env var | `ROSTER_SYNC_VERSION` | `KNOSSOS_SYNC_VERSION` |
| Preference env vars | `ROSTER_PREF_*` (14 vars) | `KNOSSOS_PREF_*` |
| Sync executable | `roster-sync` | `knossos-sync` |
| CEM manifest key | `.roster.path`, `.roster.commit`, etc. | `.knossos.path`, `.knossos.commit`, etc. |
| CEM team key | `.team.roster_path` | `.team.knossos_path` |
| Manifest source values | `"roster"`, `"roster-diverged"` | `"knossos"`, `"knossos-diverged"` |
| CLAUDE.md sync markers | `<!-- SYNC: roster-owned -->` | `<!-- SYNC: knossos-owned -->` |
| Shell functions | `is_roster_managed()`, etc. | `is_knossos_managed()`, etc. |

### What the Migration Command Handles Automatically

The `ari migrate roster-to-knossos` command rewrites:

- User manifests at `~/.claude/USER_{AGENT,SKILL,COMMAND,HOOKS}_MANIFEST.json` (source field values)
- CEM manifest at `<project>/.claude/.cem/manifest.json` (top-level keys, team keys, managed_files source values)
- Environment variable detection (advisory -- tells you which `ROSTER_*` vars are set)
- Shell profile migration script generation (optional)

### What You Must Handle Manually

- Renaming the platform directory (`~/Code/roster` to `~/Code/knossos`)
- Updating your shell profile environment variable exports
- Re-syncing each satellite to pick up new sync markers and file content
- Updating any custom scripts or CI/CD that reference `roster-sync` or `ROSTER_*` vars

---

## Pre-Migration Checklist

Complete every item before starting. Do not skip steps.

- [ ] **Verify ari version supports migration**
  ```bash
  ari migrate --help
  ```
  **Verify**: Output includes `roster-to-knossos` as a subcommand. If the command is not
  found, you need to rebuild ari from the latest main branch (see Step 1).

- [ ] **No active Claude Code sessions in any satellite**
  ```bash
  ari session list
  ```
  **Verify**: No sessions show status `ACTIVE`. If active sessions exist, wrap them first:
  ```bash
  ari session wrap
  ```

- [ ] **Note your current ROSTER_* environment variables**
  ```bash
  env | grep ROSTER_
  ```
  **Verify**: Save the output. You will need these values when updating your shell profile.
  If no output is returned, you have no `ROSTER_*` variables to migrate (skip Step 4).

- [ ] **Ensure no uncommitted changes in any satellite's `.claude/` directory**
  ```bash
  cd /path/to/satellite && git status .claude/
  ```
  **Verify**: Working tree is clean for `.claude/` files. Commit or stash changes before
  proceeding.

- [ ] **Backup satellite projects** (optional but recommended for first migration)
  ```bash
  cp -r /path/to/satellite/.claude /path/to/satellite/.claude.pre-knossos-backup
  ```
  **Verify**: Backup directory exists:
  ```bash
  ls -la /path/to/satellite/.claude.pre-knossos-backup/
  ```

---

## Step 1: Update the Knossos Platform

Pull the latest code and rename the platform directory.

### 1a. Pull latest changes

```bash
cd ~/Code/roster && git pull origin main
```

**Verify**: Pull completes successfully and the `ari migrate` command exists:
```bash
ari migrate roster-to-knossos --help
```
Expected output begins with:
```
Migrates satellite manifests from "roster" to "knossos" branding.
```

If `ari` is not on your PATH, rebuild it:
```bash
cd ~/Code/roster && CGO_ENABLED=0 go install ./cmd/ari
```

### 1b. Rename the platform directory

```bash
mv ~/Code/roster ~/Code/knossos
```

**Verify**: The directory exists at the new path:
```bash
ls ~/Code/knossos/knossos-sync
```
Expected: File exists (no "No such file" error).

### 1c. Set KNOSSOS_HOME if using a non-default path

If your platform directory is NOT at `~/Code/knossos`, you must set `KNOSSOS_HOME`:

```bash
export KNOSSOS_HOME="/your/custom/path/to/knossos"
```

If you use the default path (`~/Code/knossos`), no `KNOSSOS_HOME` export is needed.
The platform resolves to `$HOME/Code/knossos` by default.

**Verify**: The ari binary resolves the correct path:
```bash
ari --version
```
Expected: Version output without errors. If you see "cannot find knossos home" or similar,
your `KNOSSOS_HOME` is not set correctly.

### Rollback for Step 1

```bash
mv ~/Code/knossos ~/Code/roster
export KNOSSOS_HOME=""
# If you had ROSTER_HOME set:
export ROSTER_HOME="$HOME/Code/roster"
```

---

## Step 2: Migrate Manifests (Per Satellite)

Run the migration command in each satellite project. The command is idempotent -- safe to
run multiple times.

Repeat this step for every satellite project.

### 2a. Preview changes (dry-run)

```bash
cd /path/to/satellite
ari migrate roster-to-knossos
```

The command runs in dry-run mode by default. It reports what would change without modifying
any files.

**Verify**: Output shows the manifests that will be updated. Example output:
```
Roster-to-Knossos Migration (dry-run)

User Manifests:
  ~/.claude/USER_AGENT_MANIFEST.json    2 entries rewritten
  ~/.claude/USER_SKILL_MANIFEST.json    1 entry rewritten

CEM Manifest:
  .claude/.cem/manifest.json            4 fields rewritten

Summary: 3 manifests changed, 0 skipped, 7 entries rewritten

Environment variables to update:
  ROSTER_HOME -> KNOSSOS_HOME (current: /Users/you/Code/roster)

Use --apply to execute this migration.
```

If all manifests show "skipped (already migrated)", the satellite has already been
migrated. Skip to Step 3.

### 2b. Execute the migration

```bash
cd /path/to/satellite
ari migrate roster-to-knossos --apply
```

This creates `.roster-backup` backup files before rewriting each manifest.

**Verify**: Output shows successful migration with backup paths:
```
Roster-to-Knossos Migration

User Manifests:
  ~/.claude/USER_AGENT_MANIFEST.json    2 entries rewritten (backup: USER_AGENT_MANIFEST.json.roster-backup)

CEM Manifest:
  .claude/.cem/manifest.json            4 fields rewritten (backup: manifest.json.roster-backup)

Migration complete. 3 manifests updated.
```

### 2c. Verify manifest contents

Check the CEM manifest has been rewritten correctly:

```bash
cat /path/to/satellite/.claude/.cem/manifest.json | python3 -m json.tool | head -20
```

**Verify**: The output shows `"knossos"` as a top-level key (not `"roster"`):
```json
{
    "schema_version": 3,
    "knossos": {
        "path": "/Users/you/Code/knossos",
        "commit": "...",
        "ref": "main",
        "last_sync": "..."
    },
    "team": {
        "name": "your-rite",
        "knossos_path": "/Users/you/Code/knossos/rites/your-rite"
    }
}
```

Check that managed_files source values are updated:
```bash
cat /path/to/satellite/.claude/.cem/manifest.json | python3 -c "
import json, sys
m = json.load(sys.stdin)
for f in m.get('managed_files', []):
    print(f'{f[\"path\"]}: source={f[\"source\"]}')"
```

**Verify**: All source values show `knossos`, not `roster`:
```
.claude/commands: source=knossos
.claude/hooks: source=knossos
.claude/settings.local.json: source=knossos
.claude/CLAUDE.md: source=knossos
```

### 2d. Verify user manifests (run once, applies to all satellites)

```bash
cat ~/.claude/USER_AGENT_MANIFEST.json | python3 -c "
import json, sys
m = json.load(sys.stdin)
for rtype in ['agents', 'skills', 'commands', 'hooks']:
    for name, entry in m.get(rtype, {}).items():
        if 'roster' in entry.get('source', ''):
            print(f'NOT MIGRATED: {rtype}/{name} source={entry[\"source\"]}')"
```

**Verify**: No output (no remaining roster references). If entries appear, re-run
`ari migrate roster-to-knossos --apply`.

### 2e. Run migration for additional satellites

For each additional satellite, repeat steps 2a through 2c:

```bash
cd /path/to/another-satellite
ari migrate roster-to-knossos          # Preview
ari migrate roster-to-knossos --apply  # Execute
```

You can also target a specific satellite without changing directories:

```bash
ari migrate roster-to-knossos --apply --project /path/to/satellite --skip-user
```

The `--skip-user` flag avoids re-migrating user manifests (they only need migration once
since they live at `~/.claude/` globally).

### Rollback for Step 2

Restore manifest backups:

```bash
# CEM manifest
cp /path/to/satellite/.claude/.cem/manifest.json.roster-backup \
   /path/to/satellite/.claude/.cem/manifest.json

# User manifests (if needed -- restores all at once)
for f in ~/.claude/USER_*_MANIFEST.json.roster-backup; do
  cp "$f" "${f%.roster-backup}"
done
```

**Verify**: Restored manifests contain `"roster"` keys:
```bash
grep -l '"roster"' /path/to/satellite/.claude/.cem/manifest.json
```

---

## Step 3: Update Shell Profile

Environment variables in your shell profile must be updated from `ROSTER_*` to `KNOSSOS_*`.

### Option A: Use the generated migration script (recommended)

```bash
ari migrate roster-to-knossos --generate-script > /tmp/migrate-env.sh
```

**Verify**: Review the generated script before running it:
```bash
cat /tmp/migrate-env.sh
```

Expected content (example):
```bash
#!/bin/bash
# Generated by: ari migrate roster-to-knossos
# ...
# Detected variables:
#   ROSTER_HOME -> KNOSSOS_HOME

set -euo pipefail

cp ~/.zshrc ~/.zshrc.pre-knossos-migrate

# ROSTER_HOME -> KNOSSOS_HOME
sed "${SED_INPLACE[@]}" 's/ROSTER_HOME/KNOSSOS_HOME/g' ~/.zshrc
```

If the script looks correct, execute it:

```bash
chmod +x /tmp/migrate-env.sh
/tmp/migrate-env.sh
```

**Verify**: Check that your profile was updated:
```bash
grep KNOSSOS ~/.zshrc   # or ~/.bashrc
```
Expected: Lines containing `KNOSSOS_HOME` (or other `KNOSSOS_*` variables).

```bash
grep ROSTER_ ~/.zshrc   # or ~/.bashrc
```
Expected: No output (no remaining `ROSTER_*` references).

Reload your shell:
```bash
source ~/.zshrc   # or ~/.bashrc
```

### Option B: Manual update

Edit your shell profile (`~/.zshrc`, `~/.bashrc`, or `~/.bash_profile`):

**Before**:
```bash
export ROSTER_HOME="$HOME/Code/roster"
export ROSTER_SYNC_DEBUG=0
export ROSTER_PREF_AUTONOMY_LEVEL="semi-autonomous"
```

**After**:
```bash
export KNOSSOS_HOME="$HOME/Code/knossos"
export KNOSSOS_SYNC_DEBUG=0
export KNOSSOS_PREF_AUTONOMY_LEVEL="semi-autonomous"
```

Full variable rename table for manual reference:

| Remove | Add |
|--------|-----|
| `export ROSTER_HOME=...` | `export KNOSSOS_HOME=...` |
| `export ROSTER_SYNC_DEBUG=...` | `export KNOSSOS_SYNC_DEBUG=...` |
| `export ROSTER_VERBOSE=...` | `export KNOSSOS_VERBOSE=...` |
| `export ROSTER_DIR=...` | `export KNOSSOS_DIR=...` |
| `export ROSTER_PREF_AUTONOMY_LEVEL=...` | `export KNOSSOS_PREF_AUTONOMY_LEVEL=...` |
| `export ROSTER_PREF_FAILURE_HANDLING=...` | `export KNOSSOS_PREF_FAILURE_HANDLING=...` |
| `export ROSTER_PREF_OUTPUT_FORMAT=...` | `export KNOSSOS_PREF_OUTPUT_FORMAT=...` |
| `export ROSTER_PREF_ORCHESTRATION_MODE=...` | `export KNOSSOS_PREF_ORCHESTRATION_MODE=...` |
| `export ROSTER_PREF_ARTIFACT_VERIFICATION=...` | `export KNOSSOS_PREF_ARTIFACT_VERIFICATION=...` |
| `export ROSTER_PREF_NOTIFICATION_LEVEL=...` | `export KNOSSOS_PREF_NOTIFICATION_LEVEL=...` |
| `export ROSTER_PREF_DEFAULT_BRANCH=...` | `export KNOSSOS_PREF_DEFAULT_BRANCH=...` |
| `export ROSTER_PREF_COMMIT_AUTO_PUSH=...` | `export KNOSSOS_PREF_COMMIT_AUTO_PUSH=...` |
| `export ROSTER_PREF_PR_AUTO_CREATE=...` | `export KNOSSOS_PREF_PR_AUTO_CREATE=...` |
| `export ROSTER_PREF_TEST_BEFORE_COMMIT=...` | `export KNOSSOS_PREF_TEST_BEFORE_COMMIT=...` |
| `export ROSTER_PREF_SESSION_AUTO_PARK=...` | `export KNOSSOS_PREF_SESSION_AUTO_PARK=...` |
| `export ROSTER_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES=...` | `export KNOSSOS_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES=...` |
| `export ROSTER_PREF_EDITOR_INTEGRATION_PRESERVE_CURSOR_POSITION=...` | `export KNOSSOS_PREF_EDITOR_INTEGRATION_PRESERVE_CURSOR_POSITION=...` |

After editing, reload:
```bash
source ~/.zshrc   # or ~/.bashrc
```

### Option C: Fish shell

Fish uses a different syntax. Update `~/.config/fish/config.fish`:

**Before**:
```fish
set -gx ROSTER_HOME "$HOME/Code/roster"
```

**After**:
```fish
set -gx KNOSSOS_HOME "$HOME/Code/knossos"
```

Reload:
```fish
source ~/.config/fish/config.fish
```

### Verify Step 3

```bash
echo "KNOSSOS_HOME=$KNOSSOS_HOME"
env | grep ROSTER_
```

**Verify**:
- `KNOSSOS_HOME` shows your platform path (e.g., `/Users/you/Code/knossos`)
- `env | grep ROSTER_` produces no output (all `ROSTER_*` variables are gone)

### Rollback for Step 3

If you used Option A (generated script), backups were created:
```bash
cp ~/.zshrc.pre-knossos-migrate ~/.zshrc
source ~/.zshrc
```

If you used Option B (manual), reverse the edits in your shell profile.

---

## Step 4: Re-sync Satellites

After manifests and environment variables are updated, re-sync each satellite to update
file content (CLAUDE.md markers, hook scripts, etc.).

### 4a. Standard re-sync

```bash
cd /path/to/satellite
knossos-sync sync
```

**Verify**: Sync completes without errors. Check that CLAUDE.md markers are updated:
```bash
grep "SYNC:" /path/to/satellite/.claude/CLAUDE.md
```

Expected: All markers show `knossos-owned`:
```
<!-- SYNC: knossos-owned -->
```

If you see `roster-owned` markers, use `--force`:
```bash
knossos-sync init --force
```

### 4b. If knossos-sync is not found

The `knossos-sync` script lives in the platform directory. Ensure your PATH includes it
or reference it directly:

```bash
~/Code/knossos/knossos-sync sync
```

Or add to your shell profile:
```bash
export PATH="$KNOSSOS_HOME:$PATH"
```

### 4c. Repeat for each satellite

```bash
for project in /path/to/sat1 /path/to/sat2 /path/to/sat3; do
  echo "=== Syncing $project ==="
  cd "$project" && knossos-sync sync
done
```

**Verify**: Each satellite syncs successfully.

### Rollback for Step 4

Re-syncing is non-destructive (it pulls from the platform). If you need to revert to
roster-era content, restore the full `.claude/` backup:

```bash
rm -rf /path/to/satellite/.claude
cp -r /path/to/satellite/.claude.pre-knossos-backup /path/to/satellite/.claude
```

---

## Step 5: Final Verification

Run these checks after completing all steps for every satellite.

### 5a. Environment

```bash
echo "KNOSSOS_HOME=$KNOSSOS_HOME"
env | grep ROSTER_
which knossos-sync 2>/dev/null || echo "knossos-sync: add to PATH"
```

**Verify**:
- `KNOSSOS_HOME` is set to your platform directory
- No `ROSTER_*` variables remain
- `knossos-sync` is found (or you know its path)

### 5b. Platform directory

```bash
ls ~/Code/knossos/knossos-sync
ls ~/Code/knossos/lib/knossos-home.sh
```

**Verify**: Both files exist. The old `~/Code/roster` directory should not exist:
```bash
ls ~/Code/roster 2>&1
```
Expected: `No such file or directory`

### 5c. CEM manifests (per satellite)

```bash
python3 -c "
import json, sys, glob, os

# Check CEM manifest
cem_path = sys.argv[1]
if os.path.exists(cem_path):
    m = json.load(open(cem_path))
    issues = []
    if 'roster' in m:
        issues.append('Top-level \"roster\" key still present')
    if m.get('team', {}).get('roster_path'):
        issues.append('team.roster_path still present')
    for f in m.get('managed_files', []):
        if f.get('source') == 'roster':
            issues.append(f'managed_files entry {f[\"path\"]} has source=roster')
    if issues:
        print(f'ISSUES in {cem_path}:')
        for i in issues:
            print(f'  - {i}')
    else:
        print(f'OK: {cem_path}')
else:
    print(f'NOT FOUND: {cem_path}')
" /path/to/satellite/.claude/.cem/manifest.json
```

**Verify**: Output shows `OK` for each satellite.

### 5d. User manifests

```bash
python3 -c "
import json, os, glob
home = os.path.expanduser('~')
for path in glob.glob(os.path.join(home, '.claude', 'USER_*_MANIFEST.json')):
    m = json.load(open(path))
    for rtype, entries in m.items():
        if isinstance(entries, dict):
            for name, entry in entries.items():
                src = entry.get('source', '')
                if 'roster' in src:
                    print(f'NOT MIGRATED: {os.path.basename(path)} -> {rtype}/{name} source={src}')
print('User manifest check complete.')
"
```

**Verify**: Output shows only "User manifest check complete." with no NOT MIGRATED lines.

### 5e. CLAUDE.md sync markers (per satellite)

```bash
grep -n "roster" /path/to/satellite/.claude/CLAUDE.md || echo "No roster references found (good)"
```

**Verify**: Output shows "No roster references found (good)" or only references in
user-content sections that you control (e.g., documentation about the migration itself).

---

## Full Rollback Procedure

If you need to completely revert the migration, follow these steps in reverse order.

### Revert Step 4 (re-sync)
```bash
# Restore .claude/ backup for each satellite
rm -rf /path/to/satellite/.claude
cp -r /path/to/satellite/.claude.pre-knossos-backup /path/to/satellite/.claude
```

### Revert Step 3 (shell profile)
```bash
# If you used the generated script:
cp ~/.zshrc.pre-knossos-migrate ~/.zshrc
source ~/.zshrc

# If manual, re-add ROSTER_* exports and remove KNOSSOS_* exports
```

### Revert Step 2 (manifests)
```bash
# CEM manifest per satellite
cp /path/to/satellite/.claude/.cem/manifest.json.roster-backup \
   /path/to/satellite/.claude/.cem/manifest.json

# User manifests (global, once)
for f in ~/.claude/USER_*_MANIFEST.json.roster-backup; do
  cp "$f" "${f%.roster-backup}"
done
```

### Revert Step 1 (platform directory)
```bash
mv ~/Code/knossos ~/Code/roster
```

### Verify rollback
```bash
echo "ROSTER_HOME=$ROSTER_HOME"
ls ~/Code/roster/knossos-sync   # Note: file was already renamed in-repo
grep '"roster"' /path/to/satellite/.claude/.cem/manifest.json
```

---

## Troubleshooting

### "ari: command not found"

The ari binary needs to be on your PATH or rebuilt after the directory rename.

```bash
cd ~/Code/knossos && CGO_ENABLED=0 go install ./cmd/ari
```

If `go install` puts the binary in `~/go/bin/`, ensure that is in your PATH:
```bash
export PATH="$HOME/go/bin:$PATH"
```

### "knossos-sync: command not found"

The sync script lives in the platform directory. Either add it to PATH:

```bash
export PATH="$KNOSSOS_HOME:$PATH"
```

Or reference it directly:
```bash
~/Code/knossos/knossos-sync sync
```

### "cannot find knossos home" or similar path resolution error

The platform expects to find itself at `$KNOSSOS_HOME` (default: `~/Code/knossos`).

1. Check the directory exists: `ls ~/Code/knossos/`
2. If you use a custom path, set `KNOSSOS_HOME`:
   ```bash
   export KNOSSOS_HOME="/your/custom/path"
   ```
3. Verify: `ls "$KNOSSOS_HOME/knossos-sync"`

### Old `<!-- SYNC: roster-owned -->` markers in CLAUDE.md

Run a force re-init to regenerate all managed files:

```bash
cd /path/to/satellite
knossos-sync init --force
```

**Verify**: `grep "roster-owned" .claude/CLAUDE.md` returns no results.

### Migration command says "already migrated" but manifests still have roster references

This can happen if the manifest was partially hand-edited. Force re-run:

```bash
# Check the actual content
cat .claude/.cem/manifest.json | python3 -m json.tool | grep roster
```

If roster references remain, the JSON structure may not match what the migration tool
expects. Manually edit the manifest or restore from backup and re-run:

```bash
cp .claude/.cem/manifest.json.roster-backup .claude/.cem/manifest.json
ari migrate roster-to-knossos --apply
```

### Permission denied on migrate

Ensure the ari binary is up to date and you have write access to `.claude/` directories:

```bash
ls -la .claude/.cem/manifest.json
ls -la ~/.claude/USER_*_MANIFEST.json
```

### CI/CD pipelines referencing roster-sync or ROSTER_* variables

Update pipeline configuration to use the new names. Common locations:

- `.github/workflows/*.yml` -- env vars and script references
- `Makefile` -- any `roster-sync` calls
- `docker-compose.yml` -- environment variable mappings
- Custom setup scripts -- `ROSTER_HOME` exports

Search your satellite for remaining references:
```bash
grep -r "ROSTER_" /path/to/satellite --include="*.yml" --include="*.yaml" --include="*.sh" --include="Makefile"
grep -r "roster-sync" /path/to/satellite --include="*.yml" --include="*.yaml" --include="*.sh" --include="Makefile"
```

---

## Compatibility Matrix

| Knossos Platform Version | roster-era manifests | knossos-era manifests | Notes |
|--------------------------|---------------------|-----------------------|-------|
| Pre-rename (roster) | Compatible | Not supported | Original state |
| Post-rename (knossos), pre-migrate | NOT COMPATIBLE | Compatible | Must run migration |
| Post-rename (knossos), post-migrate | N/A (migrated) | Compatible | Target state |

| Component | Requires Migration | Migration Method |
|-----------|--------------------|------------------|
| CEM manifest (`.claude/.cem/manifest.json`) | Yes | `ari migrate roster-to-knossos --apply` |
| User manifests (`~/.claude/USER_*_MANIFEST.json`) | Yes | `ari migrate roster-to-knossos --apply` |
| Shell profile env vars | Yes | Generated script or manual edit |
| Platform directory path | Yes | `mv ~/Code/roster ~/Code/knossos` |
| CLAUDE.md sync markers | Yes | `knossos-sync sync` or `knossos-sync init --force` |
| Custom satellite scripts | Manual | Search and replace `ROSTER_*` and `roster-sync` |

---

## Quick Reference: Multi-Satellite Migration Script

For developers with many satellites, here is a complete migration sequence. Review and
customize the `SATELLITES` array before running.

```bash
#!/bin/bash
set -euo pipefail

# -- Configure these --
SATELLITES=(
  "$HOME/Code/project-alpha"
  "$HOME/Code/project-beta"
  "$HOME/Code/project-gamma"
)

echo "=== Step 1: Rename platform directory ==="
if [[ -d "$HOME/Code/roster" ]]; then
  mv "$HOME/Code/roster" "$HOME/Code/knossos"
  echo "Renamed ~/Code/roster -> ~/Code/knossos"
else
  echo "~/Code/roster not found (already renamed or custom path)"
fi

echo ""
echo "=== Step 2: Migrate user manifests (once) ==="
ari migrate roster-to-knossos --apply --skip-project
echo ""

echo "=== Step 3: Migrate CEM manifests (per satellite) ==="
for sat in "${SATELLITES[@]}"; do
  echo "--- $sat ---"
  if [[ -f "$sat/.claude/.cem/manifest.json" ]]; then
    ari migrate roster-to-knossos --apply --skip-user --project "$sat"
  else
    echo "  No CEM manifest found, skipping"
  fi
  echo ""
done

echo "=== Step 4: Re-sync satellites ==="
for sat in "${SATELLITES[@]}"; do
  echo "--- Syncing $sat ---"
  cd "$sat" && "$HOME/Code/knossos/knossos-sync" sync
  echo ""
done

echo "=== Step 5: Generate env var migration script ==="
ari migrate roster-to-knossos --generate-script > /tmp/migrate-env.sh
echo "Review /tmp/migrate-env.sh, then run it:"
echo "  chmod +x /tmp/migrate-env.sh && /tmp/migrate-env.sh"
echo ""
echo "Migration complete. Remember to:"
echo "  1. Review and run /tmp/migrate-env.sh"
echo "  2. Reload your shell: source ~/.zshrc"
echo "  3. Verify: env | grep ROSTER_ (should be empty)"
```

---

## Backup File Reference

| Backup File | Created By | Location |
|-------------|-----------|----------|
| `manifest.json.roster-backup` | `ari migrate --apply` | `<satellite>/.claude/.cem/` |
| `USER_*_MANIFEST.json.roster-backup` | `ari migrate --apply` | `~/.claude/` |
| `~/.zshrc.pre-knossos-migrate` | Generated migration script | `~/` |
| `~/.bashrc.pre-knossos-migrate` | Generated migration script | `~/` |
| `.claude.pre-knossos-backup` | Manual pre-migration backup | `<satellite>/` |

Backups are NOT overwritten on subsequent migration runs. If a `.roster-backup` file
already exists, the migration tool preserves the original backup.

After confirming migration is successful across all satellites, you may clean up backups:

```bash
# Remove CEM manifest backups
find /path/to/satellites -name "manifest.json.roster-backup" -delete

# Remove user manifest backups
rm -f ~/.claude/USER_*_MANIFEST.json.roster-backup

# Remove shell profile backups
rm -f ~/.zshrc.pre-knossos-migrate ~/.bashrc.pre-knossos-migrate

# Remove manual .claude backups
rm -rf /path/to/satellite/.claude.pre-knossos-backup
```
