# SPIKE: Sync Safety Audit — Edge Cases for Pre-Existing .claude/ Content

> Comprehensive audit of ari sync behavior when targeting projects with pre-existing agents, skills, commands, CLAUDE.md, and settings.

**Date**: 2026-03-01
**Session**: session-20260301-202721-88642067
**Verdict**: SAFE — 3-layer protection (provenance + inscription + collision detection)

---

## Executive Summary

The ari sync pipeline is **production-grade for user content safety**. Three independent safety mechanisms work together: provenance ownership tracking, inscription region markers, and collision detection. All user content is preserved by default. The one known risk — silent overwrites of knossos-owned CLAUDE.md sections — is by-design and documented.

**Overall risk for dogfooder #2**: LOW with proper onboarding documentation.

---

## Safety Matrix

| Scenario | What Happens | User Content Safe? |
|----------|-------------|-------------------|
| Sync on project with existing CLAUDE.md | Knossos regions overwritten, satellite regions preserved | YES (if user content in satellite sections) |
| User edits a knossos-owned CLAUDE.md section | Overwritten silently, conflict logged | NO — by design |
| Project has custom agents in .claude/agents/ | Preserved as `owner=user` or `owner=untracked` | YES |
| Sync writes rite agents over user agent with same name | Provenance checked — user-owned files NOT overwritten | YES |
| settings.json has user hooks and MCP servers | Union merge: ari-managed updated, user-defined preserved | YES |
| User-created commands/skills in .claude/ | Preserved (not in rite manifest, treated as user-owned) | YES |
| Double sync (idempotent) | writeIfChanged skips identical content, no spurious writes | YES |
| Rite switch (A → B) | Rite A agents become orphans, KEPT by default | YES |
| Sync fails midway | Atomic file writes, manifests committed last | Recoverable |
| Files in .claude/ not managed by ari (.env, scripts) | Untouched — ari only manages known patterns | YES |

---

## 5 Identified Gaps

### Gap 1: Knossos-Owned CLAUDE.md Sections Silently Overwritten (MEDIUM)

If a user edits `<!-- KNOSSOS:START execution-mode -->` content, the next sync replaces it without confirmation. A conflict is logged but not blocking.

**Mitigation**: Users must put custom content in satellite-owned sections. Document prominently.

### Gap 2: First Sync on Pre-Existing .claude/ (LOW-MEDIUM)

Pre-existing files are treated as `owner=untracked` (safe default). But there's a window where the first sync could create naming conflicts with user files.

**Mitigation**: `ari sync --recover` adopts existing files. Default preserves untracked.

### Gap 3: Mena Destructive Write in Rite Scope (LOW)

Rite-scope mena replaces stale knossos-managed files. Could surprise users expecting accumulation across rite switches.

**Mitigation**: Only knossos-managed mena removed. User-created mena preserved via provenance.

### Gap 4: No Atomic Transaction Across All Files (LOW)

Crash mid-sync leaves partial state. Manifests written last prevents metadata drift.

**Mitigation**: Re-run sync (idempotent). Worst case: missing provenance entries treated as untracked.

### Gap 5: No Reverse Collision Detection (LOW)

Rite-scope can theoretically shadow user-scope cross-rite agents. By design, cross-rite agents live at `~/.claude/agents/` (user level), not project `.claude/agents/`.

**Mitigation**: Enforced by code path separation in materializeAgents.

---

## 3-Layer Safety Architecture

```
Layer 1: PROVENANCE (ownership tracking)
  └─ Every file in .claude/ tracked with OwnerKnossos / OwnerUser / OwnerUntracked
  └─ User-owned files NEVER overwritten by sync
  └─ Orphan detection: stale knossos files → KEEP by default

Layer 2: INSCRIPTION (CLAUDE.md region markers)
  └─ <!-- KNOSSOS:START region owner={knossos|satellite} -->
  └─ Satellite regions preserved across all syncs
  └─ Knossos regions regenerated from templates
  └─ Malformed markers → adopted as satellite (safest)

Layer 3: COLLISION DETECTION (shadow prevention)
  └─ User-scope files checked against rite PROVENANCE_MANIFEST
  └─ Prevents user mena from shadowing rite mena
  └─ Logged and skipped (not deleted)
```

---

## Key Code Paths

| Component | File | Lines | Safety Mechanism |
|-----------|------|-------|-----------------|
| Atomic writes | `internal/materialize/materialize.go` | 194-199 | writeIfChanged: temp + rename |
| CLAUDE.md merge | `internal/inscription/merger.go` | 250-313 | Region ownership + conflict detection |
| Agent protection | `internal/materialize/materialize.go` | 800-986 | Provenance owner check |
| Settings merge | `internal/materialize/hooks/config.go` | 133-198 | ari-managed vs user-defined split |
| MCP merge | `internal/materialize/hooks/mcp.go` | 23-75 | Union merge (add/update, never remove) |
| Collision check | `internal/materialize/userscope/collision.go` | 45-60 | Shadow prevention |
| Orphan handling | `internal/materialize/materialize.go` | 593-696 | Detect + KEEP default + backup |

---

## Dogfooder #2 Onboarding Recommendations

1. **Tell them**: "Don't edit sections between `KNOSSOS:START` and `KNOSSOS:END` markers — add your content in the `user-content` section"
2. **First sync**: Run `ari sync --dry-run` first to preview what will change
3. **If they have existing .claude/**: Use `ari sync --recover` to adopt their files
4. **Settings**: Their hooks and MCP servers will be preserved across syncs
5. **Rite switching**: Their custom agents survive rite switches (kept as orphans)

---

## Test Coverage

| Area | Test File | Coverage |
|------|-----------|----------|
| Inscription merge | `internal/inscription/*_test.go` | Satellite preservation, deprecation, conflicts |
| User scope sync | `internal/materialize/userscope/sync_test.go` | Collision, divergence, dry-run |
| Atomic writes | `internal/materialize/write_test.go` | Skip identical, atomic rename |
| Integration | `internal/materialize/unified_sync_test.go` | Rite switch, orphans, soft mode |
