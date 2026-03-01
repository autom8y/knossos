# SPIKE: v0.2.0 Release Recovery -- Homebrew Blocked

> Root cause analysis and recovery plan for the v0.2.0 release failure where GoReleaser succeeded at building and publishing the GitHub Release but failed to push the Homebrew formula due to an empty `HOMEBREW_TAP_TOKEN`.

**Date**: 2026-03-01
**Author**: Spike (release recovery)
**Prior Art**: `docs/spikes/SPIKE-distribution-audit-gap-report.md` (predicted this exact failure at GAP-D01)

---

## 1. Question and Context

### What are we trying to learn?

1. What is the exact root cause of the v0.2.0 Homebrew failure?
2. What is the safest recovery path that delivers the Homebrew formula without re-releasing?
3. Should we migrate from deprecated `brews` to `homebrew_casks` as part of recovery?
4. What configuration hardening prevents this class of failure in the future?

### What decision will this inform?

Whether to (a) rerun the failed workflow after setting the token, (b) delete the release and re-release, or (c) manually push the formula and fix the pipeline for v0.3.0.

---

## 2. Root Cause Analysis

### The failure chain

```
Workflow trigger: tag v0.2.0 pushed
    |
    v
GoReleaser v2.14.1 starts
    |
    +-- go mod tidy                          [OK]
    +-- build (4 binaries, CGO_ENABLED=0)    [OK] (~3m44s)
    +-- archive (4 tar.gz + checksums)       [OK]
    +-- homebrew formula generated locally    [OK] (dist/homebrew/Formula/ari.rb written)
    +-- publish: GitHub release              [OK] (mode: replace, 5 assets uploaded)
    +-- publish: homebrew formula             [FAIL]
         |
         +-- Rate limit check                [WARN: could not check]
         +-- GET repos/autom8y/homebrew-tap   [401 Bad credentials]
         |
         v
    FATAL: "homebrew formula: could not get default branch"
    Exit code 1
```

### Root cause: HOMEBREW_TAP_TOKEN is empty

The workflow log explicitly shows the token is empty:

```
env:
  GITHUB_TOKEN: ***
  HOMEBREW_TAP_TOKEN:
```

`GITHUB_TOKEN` was masked (present), `HOMEBREW_TAP_TOKEN` had no value. The secret either:
- Was never created on `autom8y/knossos`
- Was deleted or expired
- Has a name mismatch (case-sensitive)

The GoReleaser config correctly references `{{ .Env.HOMEBREW_TAP_TOKEN }}`, and the workflow correctly passes `HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}`. The wiring is correct; the secret value is missing.

### Why GoReleaser failed entirely (not just Homebrew)

GoReleaser treats the Homebrew formula push as part of the publish pipeline. When `brews` is configured and the token fails, it does **not** degrade gracefully -- it exits with error code 1, marking the entire workflow as failed. The release itself was already published to GitHub before the Homebrew step ran (the publish pipeline is sequential: releases first, then homebrew).

---

## 3. Current State of v0.2.0

| Component | State |
|-----------|-------|
| Git tag `v0.2.0` | Pushed to origin |
| GitHub Release | Published (not draft, not prerelease) |
| Binary assets | 5 assets uploaded (4 platform archives + checksums) |
| Release notes | Populated with changelog |
| Homebrew formula `ari.rb` | NOT delivered to `autom8y/homebrew-tap` |
| `autom8y/homebrew-tap` | Repo exists (public), contains only `.gitignore` + `README.md`, no `Formula/` directory |
| Workflow run 22548648697 | Status: completed, conclusion: failure |

The release is **partially complete** -- fully usable via direct download or `go install`, but the Homebrew channel is broken.

---

## 4. Recovery Options Analysis

### Option A: Rerun the failed workflow

```bash
gh secret set HOMEBREW_TAP_TOKEN --repo autom8y/knossos  # Set the token first
gh run rerun 22548648697 --repo autom8y/knossos
```

**Pros**: Simplest, uses existing infrastructure.

**Risks**:
1. GoReleaser will rebuild everything from scratch and attempt to re-upload release assets. The config has `mode: replace` for release notes, but does NOT have `replace_existing_artifacts: true`. This means the 5 already-uploaded assets may cause `422 already_exists` errors.
2. Even if `mode: replace` handles the release object, asset upload conflicts are a separate concern.
3. Builds may not be byte-identical (different timestamp in ldflags).

**Verdict**: RISKY. Likely to fail on asset upload unless `replace_existing_artifacts: true` is added to `.goreleaser.yaml` first. But adding that requires a new commit on main, and the tag points to the current HEAD. A config change after the tag means the rerun builds from the tagged commit (which lacks the config change).

**Mitigation**: `gh run rerun` replays the workflow against the same commit, so adding `replace_existing_artifacts` to main won't help unless we re-tag.

### Option B: Delete release + re-tag + re-push

```bash
gh release delete v0.2.0 --yes --repo autom8y/knossos
git push origin :refs/tags/v0.2.0   # Delete remote tag
git tag -d v0.2.0                    # Delete local tag
# Fix config if needed, commit
git tag v0.2.0
git push origin v0.2.0              # Triggers fresh release
```

**Pros**: Clean slate. No asset conflicts. Can include config fixes.

**Risks**:
1. Anyone who already downloaded v0.2.0 binaries has artifacts from a deleted release.
2. Checksum mismatch if binaries are rebuilt with different timestamps.
3. Git history pollution (tag movement).
4. Requires the HOMEBREW_TAP_TOKEN to be set before push.

**Verdict**: SAFE but heavy-handed. The "nuclear option."

### Option C: Manually push the formula + fix pipeline for v0.3.0

Since GoReleaser generated `dist/homebrew/Formula/ari.rb` locally during the run, and the release assets are already published, we can:

1. Generate the formula locally with `goreleaser release --skip=publish --clean` or extract it from the failed run.
2. Push `Formula/ari.rb` directly to `autom8y/homebrew-tap`.
3. Fix the pipeline (set token, optionally migrate to `homebrew_casks`) for the next release.

**Pros**: No release disruption. Formula matches the published release exactly. Decouples the fix from the pipeline.

**Cons**: Manual step. The formula generation must match the published assets exactly (checksums).

**Verdict**: SAFEST for v0.2.0 recovery, but requires generating the correct formula.

### Option D: Rerun after setting token (with GoReleaser's built-in replace behavior)

After deeper analysis of the logs: GoReleaser v2.14.1 with `mode: replace` **does** handle existing releases. The key question is asset upload. Looking at the log, all 5 assets uploaded successfully on the first run. On rerun, GoReleaser will attempt to upload them again.

GoReleaser v2 behavior: when it finds the release already exists and `mode` is set, it updates the release. For assets, without `replace_existing_artifacts: true`, it will error on duplicate uploads.

**Verdict**: Will fail on asset upload unless we delete the existing assets first.

---

## 5. Recommended Recovery Path

### Recommended: Option B (Delete + Re-release) -- simplified variant

This is the cleanest path because:
1. The tag doesn't need to move (same commit).
2. We can include a `.goreleaser.yaml` fix in the process.
3. It produces a clean, complete release.

**Step-by-step procedure:**

```bash
# 1. Create and set the PAT
#    Fine-grained: Contents read+write on autom8y/homebrew-tap
#    OR Classic: repo scope
gh secret set HOMEBREW_TAP_TOKEN --repo autom8y/knossos

# 2. Delete only the GitHub Release (keep the tag)
gh release delete v0.2.0 --yes --repo autom8y/knossos
#    This removes the release object and all assets.
#    The git tag remains on the same commit.

# 3. Re-push the tag to trigger the workflow
#    Since the tag already exists, delete and recreate it on the same commit:
git tag -d v0.2.0
git tag v0.2.0
git push origin :refs/tags/v0.2.0
git push origin v0.2.0

# 4. Monitor the workflow
gh run list --repo autom8y/knossos --workflow=release.yml --limit 1
gh run watch <run-id> --repo autom8y/knossos
```

**Alternative shortcut** (if you want to avoid tag manipulation):

```bash
# 1. Set the token
gh secret set HOMEBREW_TAP_TOKEN --repo autom8y/knossos

# 2. Delete the release AND assets only
gh release delete v0.2.0 --yes --cleanup-tag --repo autom8y/knossos

# 3. Re-tag on the same commit and push
git tag v0.2.0
git push origin v0.2.0
```

### If the rerun approach is preferred (Option A variant)

If you want to try the simpler rerun without re-tagging:

```bash
# 1. Set the token
gh secret set HOMEBREW_TAP_TOKEN --repo autom8y/knossos

# 2. Delete the release assets (but keep the release)
gh release view v0.2.0 --json assets --jq '.assets[].name' --repo autom8y/knossos | \
  xargs -I {} gh release delete-asset v0.2.0 {} --yes --repo autom8y/knossos

# 3. Rerun the workflow
gh run rerun 22548648697 --repo autom8y/knossos

# 4. Monitor
gh run watch <new-run-id> --repo autom8y/knossos
```

This preserves the release object and its URL but rebuilds assets. Binaries will have slightly different build timestamps but identical code.

---

## 6. GoReleaser Deprecation: `brews` -> `homebrew_casks`

### Current warning

```
brews is being phased out in favor of homebrew_casks
```

### Should we migrate now?

**No. Migrate in a separate, non-emergency change.**

Rationale:
1. The recovery should fix one thing: the token. Mixing a config migration into the recovery increases risk.
2. The `brews` deprecation is a warning, not an error. It will work for GoReleaser v2.x.
3. Migration from `brews` to `homebrew_casks` changes the Homebrew artifact type (Formula -> Cask), which has user-facing implications (different install paths, `brew install --cask` vs `brew install`).
4. A formula-to-cask migration requires `tap_migrations.json` in the tap repo for upgrade continuity.

### Migration plan (post-recovery, separate PR)

1. Rename `brews:` to `homebrew_casks:` in `.goreleaser.yaml`
2. Change `directory: Formula` to `directory: Casks` (or omit for default)
3. Update `install:` block syntax (Cask DSL differs from Formula DSL)
4. Add `tap_migrations.json` to `autom8y/homebrew-tap` root:
   ```json
   { "ari": "ari" }
   ```
5. Remove old `Formula/ari.rb` after migration
6. Update docs referencing `brew install` (may need `brew install --cask`)
7. Test with `goreleaser release --snapshot --clean` before cutting a release

### Additional deprecations observed

| Deprecation | Replacement | Severity |
|-------------|-------------|----------|
| `archives.format` | `archives.formats` | Warning |
| `archives.format_overrides.format` | `archives.format_overrides.formats` | Warning |
| `brews` | `homebrew_casks` | Warning |

All three should be addressed in a single config modernization PR.

---

## 7. Configuration Hardening Recommendations

### Immediate (part of recovery)

1. **Set `HOMEBREW_TAP_TOKEN`** as a repository secret on `autom8y/knossos`.

### Short-term (next PR after recovery)

2. **Add `replace_existing_artifacts: true`** to `.goreleaser.yaml` release section:
   ```yaml
   release:
     github:
       owner: autom8y
       name: knossos
     replace_existing_artifacts: true
     mode: replace
   ```
   This prevents future rerun failures from asset conflicts.

3. **Add GoReleaser snapshot validation to CI** (GAP-D09 from prior audit):
   ```yaml
   # In ariadne-tests.yml
   - name: Validate GoReleaser Config
     uses: goreleaser/goreleaser-action@v6
     with:
       args: check
     env:
       GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   ```

4. **Consider `skip_upload: auto`** for Homebrew in non-tag builds to prevent accidental formula pushes during snapshot testing.

### Medium-term

5. **Migrate `brews` to `homebrew_casks`** (see section 6 above).
6. **Address all three deprecation warnings** in a single config modernization PR.
7. **Add a pre-release checklist** to the release workflow that validates secrets exist:
   ```yaml
   - name: Validate secrets
     run: |
       if [ -z "$HOMEBREW_TAP_TOKEN" ]; then
         echo "::error::HOMEBREW_TAP_TOKEN is not set"
         exit 1
       fi
     env:
       HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
   ```
   This fails fast instead of building for 4 minutes then failing at the Homebrew step.

---

## 8. PAT Requirements

The `HOMEBREW_TAP_TOKEN` needs specific permissions to push to `autom8y/homebrew-tap`:

### Fine-grained PAT (recommended)

| Setting | Value |
|---------|-------|
| Resource owner | `autom8y` |
| Repository access | `autom8y/homebrew-tap` only |
| Permissions | Contents: Read and write |
| Expiration | 90 days minimum (set a calendar reminder) |

### Classic PAT (alternative)

| Setting | Value |
|---------|-------|
| Scope | `repo` (full control of private repositories) |
| Note | GoReleaser Homebrew tap push |

Fine-grained is preferred because it follows least-privilege: write access only to the homebrew-tap repo, nothing else.

### Token lifecycle risk

PAT expiration is the most likely cause of future recurrence. Mitigations:
- Set the longest reasonable expiration (1 year for fine-grained).
- Add a calendar reminder 2 weeks before expiration.
- Consider using a GitHub App token (no expiration) if available at the org level.

---

## 9. Follow-up Actions

| # | Action | Owner | Priority | Depends On |
|---|--------|-------|----------|------------|
| 1 | Create GitHub PAT with write access to `autom8y/homebrew-tap` | User (manual) | IMMEDIATE | -- |
| 2 | Set `HOMEBREW_TAP_TOKEN` secret on `autom8y/knossos` | User (manual) | IMMEDIATE | #1 |
| 3 | Delete v0.2.0 release and re-tag to trigger fresh release | User (manual) | IMMEDIATE | #2 |
| 4 | Verify Homebrew formula delivered to `autom8y/homebrew-tap` | pipeline-monitor | IMMEDIATE | #3 |
| 5 | Test `brew install autom8y/tap/ari` end-to-end | User (manual) | IMMEDIATE | #4 |
| 6 | Add `replace_existing_artifacts: true` to `.goreleaser.yaml` | PR | SHORT-TERM | -- |
| 7 | Add secret validation step to release workflow | PR | SHORT-TERM | -- |
| 8 | Migrate `brews` to `homebrew_casks` + address all deprecations | PR | MEDIUM-TERM | #5 (v0.2.0 stable) |
| 9 | Add GoReleaser config validation to CI | PR | MEDIUM-TERM | -- |

---

## 10. References

- [GoReleaser Deprecation Notices](https://goreleaser.com/deprecations/) -- `brews` phased out in favor of `homebrew_casks` since v2.10
- [GoReleaser Homebrew Casks](https://goreleaser.com/customization/homebrew_casks/) -- Replacement configuration docs
- [GoReleaser Release Upload Errors](https://goreleaser.com/errors/release-upload/) -- `already_exists` handling and `replace_existing_artifacts`
- [GoReleaser Release Configuration](https://goreleaser.com/customization/release/) -- `mode: replace` behavior
- `docs/spikes/SPIKE-distribution-audit-gap-report.md` -- Prior audit that predicted this failure (GAP-D01)
- `.goreleaser.yaml` -- Current release configuration
- `.github/workflows/release.yml` -- Release automation workflow
- Workflow run 22548648697 logs -- Primary evidence source
