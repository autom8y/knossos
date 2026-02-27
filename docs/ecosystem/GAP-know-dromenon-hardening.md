## Gap Analysis: /know Dromenon Satellite-Readiness Gaps

Confirmation of two gaps identified in `.claude/wip/KNOW-10X-STRATEGY.md` (sections 2.6 and 2.9.1).

### Gap A: Hardcoded Go source_scope

**Affected file**: `rites/shared/mena/know/INDEX.dro.md`
**Affected lines**: 162-165 (Phase 3, step 2 -- frontmatter construction)

The frontmatter assembly statically emits:
```yaml
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
```

No pre-flight step detects project language. TypeScript and Python satellites receive Go-specific scope regardless of their actual codebase.

**Strategy doc assessment confirmed**: Section 2.9.1 (lines 371-384) correctly identifies this as a P1 satellite blocker. The proposed fix -- language detection in pre-flight setting source_scope dynamically based on `go.mod` vs `package.json` vs `pyproject.toml` -- is sound and self-contained.

### Gap B: No custom domain scanning

**Affected file**: `rites/shared/mena/know/INDEX.dro.md`
**Affected lines**: 26-31 (Pre-flight step 2 -- domain registry loading)

Domain discovery loads exclusively from the pinakes registry via `Skill("pinakes")`. There is no instruction to also scan `.know/criteria/*.md` for satellite-authored custom domain criteria files.

**Strategy doc assessment confirmed**: Section 2.6 (lines 265-310) correctly identifies this as a P2 blocker. Option B (`.know/criteria/`) is the right approach: self-contained, no sync pipeline changes, low implementation effort.

### Additional Concern

Pre-flight step 2 (lines 26-31) also has no fallback if `Skill("pinakes")` fails to load or the pinakes INDEX is absent. A satellite that has not synced pinakes skills would hit an unhandled error. This is minor but worth noting for the integration engineer.

### Complexity: PATCH

Both gaps are localized to `INDEX.dro.md` dromenon text. No Go source changes required. No schema changes. No sync pipeline changes.

### Success Criteria

- `/know --all` on a TypeScript satellite produces `source_scope` containing `*.ts` paths, not `*.go`
- `/know --all` discovers and generates domains from `.know/criteria/*.md` in addition to pinakes registry
- `ari knows` reports custom domains alongside built-in domains

### Test Satellites

- knossos (Go, baseline -- verify no regression)
- Any TypeScript satellite (validate dynamic source_scope)
- A satellite with `.know/criteria/database-schema.lego.md` present (validate custom domain discovery)
