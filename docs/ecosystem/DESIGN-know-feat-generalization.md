# Context Design: .know/feat/ Criteria Generalization + Seed Freshness Gate

**Status**: Design Complete
**Session**: session-20260303-114830-beb1bf03
**Affects**: Feature census criteria, feature knowledge criteria, `/know` dromenon, INDEX.md source_scope

---

## Problem Summary

Two interrelated problems prevent `.know/feat/` from working in any project except knossos itself:

1. **Feature criteria hardcode knossos paths.** The census criteria reference `rites/*/manifest.yaml`, `internal/*/`, `docs/decisions/ADR-*.md`, `rites/*/mena/**/*.dro.md`, `rites/*/agents/*.md`, `INTERVIEW_SYNTHESIS.md` -- none of which exist in a typical satellite. A Next.js app or Go microservice would find zero useful source material.

2. **Seed freshness is unenforced.** Feature analysis depends on codebase domain seeds (`.know/architecture.md`, etc.) for structural context, but this dependency is advisory. Running `/know --scope=feature` on stale or missing seeds produces degraded feature knowledge silently.

---

## Design Decisions

### Decision 1: Use abstract source categories, not concrete paths

**Chosen approach**: Replace knossos-specific path lists with abstract source categories that the theoros resolves at observation time using language detection and `.know/` seeds.

**Rationale**: The 5 existing codebase domain criteria (architecture, scar-tissue, conventions, design-constraints, test-coverage) already follow this pattern. They define a "Language Detection" section and a "Scope Adaptation" table that maps abstract concepts to language-specific paths. Feature criteria should follow the identical pattern.

**Rejected alternative**: Per-language path templates in the criteria file (e.g., "if Go, scan `internal/*/`; if TS, scan `src/*/`"). This was rejected because it duplicates the scope adaptation table already established by the codebase domain criteria pattern, and requires updating the feature criteria every time a new language is supported.

**Rejected alternative**: Making theoros auto-discover paths with no guidance. This was rejected because it provides no structure for grading -- the census criteria need to define WHAT to scan so they can grade completeness.

### Decision 2: Seven generalized source categories for census

The current census has 7 knossos-specific source types. The generalized version preserves 7 categories but redefines them as universal concepts:

| # | Knossos-specific (current) | Generalized (new) | Rationale |
|---|---|---|---|
| 1 | `rites/*/manifest.yaml` | **Project manifests** -- build config, workspace files, entry point definitions | Every project has manifests (go.mod, package.json, Cargo.toml, workspace configs) |
| 2 | `internal/*/` | **Source directory structure** -- top-level package/module organization | Every project has a source tree; the shape varies by language |
| 3 | `docs/decisions/ADR-*.md` | **Decision records and design docs** -- ADRs, RFCs, design docs in any conventional location | Many projects have decision records; absence is valid (grade accordingly) |
| 4 | `rites/*/mena/**/*.dro.md` | **User-facing commands and scripts** -- CLI commands, scripts, Makefiles, task runners | Every project has some user-facing automation surface |
| 5 | `rites/*/agents/*.md` | **Configuration and workflow definitions** -- CI/CD configs, agent definitions, workflow files | Every project has configuration that reveals feature structure |
| 6 | `INTERVIEW_SYNTHESIS.md` | **Project documentation** -- README, CONTRIBUTING, architecture docs, `.know/` seeds | Every project has some documentation; `.know/` seeds are the richest source |
| 7 | `.know/*.md` frontmatter | **Existing codebase knowledge** -- `.know/*.md` files as feature signals | This is the self-referential category; if seeds exist, they are the best source |

**Rationale**: These 7 categories are exhaustive for feature discovery in any project. Each category maps to observable artifacts regardless of language or framework. The theoros resolves specific paths at observation time.

### Decision 3: Criteria reference `.know/` seeds as primary structural context

**Chosen approach**: The generalized criteria explicitly instruct the theoros to consult `.know/architecture.md` as the structural map for source discovery, rather than hardcoding any source paths.

**Rationale**: The `.know/architecture.md` file already contains the package structure, layer model, and entry points for the specific project -- in a language-agnostic knowledge format. It is the single best source for "where are things in this project." By referencing it, the feature criteria become project-agnostic automatically: whatever `.know/architecture.md` describes for a Go project or a Next.js project, the census theoros follows.

This creates a formal dependency: feature knowledge REQUIRES fresh codebase seeds. This is already implicitly true (see spike findings). Making it explicit enables the freshness gate (Decision 5).

### Decision 4: Generalized triage heuristics

The current triage heuristics reference knossos concepts ("1+ ADRs", "user-facing dromena", "multiple rites depend on"). The generalized version:

**GENERATE if any of**:
- Decision records (ADRs, RFCs, design docs) reference the feature
- 10+ implementation files in relevant packages/modules
- User-facing commands, API endpoints, or UI components exist for the feature
- Multiple modules/packages depend on the feature (cross-cutting)

**SKIP if all of**:
- Pure utility (string helpers, file utils, common types)
- No decision records reference the feature
- Fewer than 5 implementation files
- Single-module internal detail with no cross-cutting concerns

**Rationale**: The heuristics are structurally identical to the originals but use universal terminology. "ADRs" becomes "decision records", "dromena" becomes "user-facing commands/API endpoints/UI components", "rites" becomes "modules/packages".

### Decision 5: Seed freshness gate in Feature Pre-flight

**Chosen approach**: Insert a seed freshness check in the `/know` dromenon Feature Pre-flight section, between step 3 (ensure directories) and step 4 (load criteria). The gate uses `ari knows --check` semantics (already implemented in Go) but executes within the dromenon context.

**Gate behavior**:
1. Read all `.know/*.md` frontmatter for the 5 codebase domains
2. Check freshness of each (time-based + code-based via source_scope)
3. If ALL fresh: proceed silently
4. If ANY stale or missing: present a gate with options

**Gate output format**:
```
## Seed Freshness Check

Feature analysis requires fresh codebase knowledge as context seeds.

| Domain | Status | Reason |
|--------|--------|--------|
| architecture | FRESH | |
| scar-tissue | STALE | code changed (89b109c -> f14c4ee) |
| conventions | FRESH | |
| design-constraints | MISSING | file not found |
| test-coverage | FRESH | |

**Stale/missing seeds detected.** Feature knowledge built on outdated context will be incomplete.

Options:
1. Run `/know --all` to refresh seeds, then re-invoke `/know --scope=feature` (recommended)
2. Reply "proceed anyway" to continue with current seeds (not recommended)
3. Reply "abort" to cancel
```

**Rationale**: Follows the existing human-gate pattern in the dromenon (the census review gate is precedent). Soft enforcement preserves user agency. The dromenon has access to present options and wait for user response.

### Decision 6: No auto-chaining

**Chosen approach**: The gate suggests running `/know --all` as a separate invocation rather than auto-dispatching it inline.

**Rationale**: Auto-chaining `/know --all` within a `/know --scope=feature` invocation would consume the main thread's context budget with 5 parallel theoros dispatches BEFORE the feature work even begins. The dromenon already runs at ~718 lines of instructions. Adding 5 codebase domain theoros dispatches inline would risk context exhaustion. Separate invocation keeps each run focused.

**Rejected alternative**: Auto-chain with confirmation ("Shall I run /know --all first?"). Rejected because the dromenon cannot invoke itself -- it would need to embed the entire codebase domain pipeline inline, doubling the dromenon size.

### Decision 7: `--force` bypasses the seed gate

**Chosen approach**: When `--force` is passed to `/know --scope=feature`, the seed freshness gate is skipped entirely.

**Rationale**: `--force` already means "skip freshness checks" throughout the `/know` dromenon. The user is explicitly requesting regeneration regardless of state. Requiring fresh seeds when the user said "force" would be contradictory.

### Decision 8: Generalize INDEX.md source_scope frontmatter

**Chosen approach**: Replace the knossos-specific `source_scope` in the INDEX.md frontmatter with language-detected paths, following the same pattern as codebase domain output assembly.

**Rationale**: The `source_scope` in frontmatter is used by `internal/know/` for scoped staleness checking. If it references `./rites/*/manifest.yaml`, the staleness check will never trigger in a satellite project (those paths don't exist, so no changes are detected, so the census appears perpetually fresh). The scope must match the actual source paths the census scanned.

---

## Components Affected

### File 1: `rites/shared/mena/pinakes/domains/feature-census.lego.md`

**Change type**: Content rewrite of Scope section, Triage Heuristics section
**Backward compatibility**: COMPATIBLE -- criteria files are consumed by theoros dispatch prompts, not by Go code or schemas. Changing criteria content changes theoros behavior but breaks no interfaces.

**Specific changes**:

1. **Add Language Detection section** (new, after frontmatter). Follow the exact pattern from `architecture.lego.md` lines 12-25: language manifest detection, Scope Adaptation table.

2. **Replace Scope > Target sources** (lines 12-21). Remove knossos-specific paths. Replace with 7 generalized source categories:
   - Project manifests (build config, workspace files)
   - Source directory structure (top-level package/module layout)
   - Decision records and design docs (ADRs, RFCs, in conventional locations)
   - User-facing commands and scripts (CLI, Makefiles, task runners, API route definitions)
   - Configuration and workflow definitions (CI/CD, workflow files, agent definitions)
   - Project documentation (README, CONTRIBUTING, architecture docs)
   - Existing codebase knowledge (`.know/*.md` files)

3. **Add `.know/` dependency instruction** (new, within Scope section): "REQUIRED: Read `.know/architecture.md` first to understand the project's package structure and source layout. Use the architecture knowledge as your navigation map for all other source categories. If `.know/architecture.md` does not exist, fall back to scanning the project root for manifest files and source directories."

4. **Remove knossos NOTE** (line 21): "Scan rite SOURCE artifacts (`rites/`), NOT materialized outputs (`.claude/`)." This is knossos-specific. Replace with: "Scan source files, not build outputs or generated artifacts."

5. **Replace Triage Heuristics** (lines 43-55). Replace knossos-specific terms:
   - "1+ ADRs" -> "Decision records (ADRs, RFCs, design docs) reference the feature"
   - "User-facing dromena (commands)" -> "User-facing commands, API endpoints, or UI components exist"
   - "Multiple rites depend on" -> "Multiple modules/packages depend on the feature"
   - "Single-rite internal detail" -> "Single-module internal detail"

6. **Replace Criterion 1 references to "7 source types"** (lines 59-76): Update grade thresholds to reference "7 source categories" (same count, different names). Update evidence collection to reference the generalized categories.

7. **Replace Criterion 2 cross-check references** (lines 87-95): "package names, ADR titles, command names, and rite descriptions" -> "package/module names, decision record titles, command names, and configuration descriptions"

8. **Replace Criterion 3 heuristic references** (lines 105-107): Update to match the generalized triage heuristics.

### File 2: `rites/shared/mena/pinakes/domains/feature-knowledge.lego.md`

**Change type**: Content rewrite of Scope section, minor evidence reference updates
**Backward compatibility**: COMPATIBLE -- same rationale as File 1.

**Specific changes**:

1. **Add Language Detection section** (new, after frontmatter). Same pattern as architecture.lego.md.

2. **Replace Scope > Target sources** (lines 13-18). Remove knossos-specific paths. Replace with generalized source categories scoped to a single feature:
   - Source code in relevant packages/modules (identified from census source_evidence)
   - Decision records referencing this feature
   - Configuration files affecting this feature's behavior
   - Existing `.know/` files (`architecture.md`, `scar-tissue.md`, `conventions.md`) for structural context
   - Test files covering this feature's functionality
   - Project documentation sections relevant to this feature

3. **Remove knossos NOTE** (line 20): Replace with: "Scan source files, not build outputs or generated artifacts."

4. **Update Criterion 1 evidence references** (lines 33-37):
   - "ADRs in `docs/decisions/`" -> "Decision records (ADRs, RFCs, design docs) referencing this feature"
   - "Spike artifacts in `.ledge/spikes/`" -> "Design exploration artifacts (spikes, prototypes, research docs) if they exist"
   - "INTERVIEW_SYNTHESIS.md sections" -> "Project documentation sections describing the feature's purpose"
   - Keep `.know/` reference as-is (already generic)

5. **Update Criterion 3 evidence references** (lines 75-79):
   - "packages under `internal/` (and `cmd/` if applicable)" -> "packages/modules implementing this feature (identified from census source_evidence and `.know/architecture.md`)"
   - Keep the rest (data flow, API surface, test locations) -- these are already generic.

6. **Update Criterion 4 evidence references** (lines 96-101):
   - "source comments, ADRs, or commit messages" -> "source comments, decision records, or commit messages"
   - "`.know/scar-tissue.md` entries" -- keep as-is (already generic reference)

### File 3: `rites/shared/mena/know/INDEX.dro.md`

**Change type**: Insert seed freshness gate in Feature Pre-flight; update INDEX.md source_scope template
**Backward compatibility**: COMPATIBLE -- additive change to dromenon flow. Existing `--force` bypass preserved.

**Specific changes**:

1. **Insert new step 3.5** between current step 3 (ensure directories) and step 4 (load criteria) in the Feature Pre-flight section (after line 277). Add seed freshness gate:

```markdown
4. **Seed freshness gate** (skip if `--force` is set):
   Check whether the 5 codebase domain seeds are fresh. Feature analysis depends on
   these seeds for structural context.

   Required seed domains: architecture, scar-tissue, conventions, design-constraints, test-coverage.

   For each required domain:
   - If `.know/{domain}.md` does not exist: mark as MISSING
   - If file exists: read YAML frontmatter, check time-based and code-based freshness
     (same algorithm as the codebase domain pipeline's generation queue check in step 3)
   - Time-fresh: `generated_at` + `expires_after` has not passed
   - Code-fresh: `source_hash` matches current HEAD (or no in-scope files changed per source_scope)

   If ALL domains are FRESH: proceed silently to step 5.

   If ANY domain is STALE or MISSING: present the gate:

   ```
   ## Seed Freshness Check

   Feature analysis requires fresh codebase knowledge as context seeds.

   | Domain | Status | Reason |
   |--------|--------|--------|
   | {domain} | {FRESH/STALE/MISSING} | {reason or empty} |
   ...

   **Stale/missing seeds detected.** Feature knowledge built on outdated context will be incomplete.

   Options:
   1. Run `/know --all` to refresh seeds, then re-invoke `/know --scope=feature` (recommended)
   2. Reply "proceed anyway" to continue with current seeds (not recommended)
   3. Reply "abort" to cancel
   ```

   **Wait for user response.** Do NOT continue until the user explicitly responds.
   - If user says "proceed anyway" or equivalent: continue to step 5 with a warning logged.
   - If user says "abort" or equivalent: STOP.
   - If user says anything suggesting refresh: advise them to run `/know --all` and re-invoke.
```

2. **Renumber subsequent steps**: Current step 4 (load criteria) becomes step 5. Current step 5 (route by argument) becomes step 6.

3. **Update INDEX.md source_scope template** (lines 401-419). Replace knossos-specific paths with language-detected paths:

```yaml
source_scope:
  - "{language-detected source glob 1}"
  - "{language-detected source glob 2}"
  - "{language-detected manifest}"
  - "./.know/*.md"
```

Use the same language detection logic already defined in Phase 3 Output Assembly of the codebase domain pipeline (lines 166-170):
- Go: `["./cmd/**/*.go", "./internal/**/*.go", "./go.mod", "./.know/*.md"]`
- TS: `["./src/**/*.ts", "./lib/**/*.ts", "./package.json", "./.know/*.md"]`
- Python: `["./src/**/*.py", "./app/**/*.py", "./pyproject.toml", "./.know/*.md"]`
- Fallback: `["./src/**/*", "./.know/*.md"]`

Always append `./.know/*.md` because the census reads existing `.know/` seeds.

4. **Update census theoros Scope Reminder** (line 387): Change from "Read the **Scope** section in the census criteria above. It defines the target sources to scan. Scan ALL of them." to: "Read the **Scope** section in the census criteria above. It defines the source categories to scan. Resolve each category to actual files in this project using language detection and `.know/architecture.md`. Scan ALL categories."

5. **Update per-feature theoros Scope Reminder** (lines 594-596): Change "Focus on the packages and files relevant to '{slug}'" to "Focus on the packages/modules relevant to '{slug}'. Use `.know/architecture.md` for project navigation."

---

## Backward Compatibility: COMPATIBLE

All changes are content-level modifications to criteria files and dromenon instructions. No Go code changes. No schema changes. No new fields. No interface changes.

**Impact on existing `.know/` files**: None. Existing `.know/architecture.md` (etc.) files are unaffected. Their frontmatter schema is unchanged.

**Impact on existing census/feature files**: If `.know/feat/INDEX.md` or `.know/feat/{slug}.md` files existed (they do not today -- the feature flow has never been populated), they would continue to parse correctly. The source_scope values would differ on regeneration but the Meta struct in `internal/know/know.go` handles any string slice.

**Impact on satellites**: Positive. Satellites that run `/know --scope=feature` will now get criteria that work with their actual project structure instead of failing to find knossos-specific paths.

---

## Integration Test Matrix

| Satellite Type | Test | Expected Outcome |
|---|---|---|
| **knossos itself** | Run `/know --scope=feature --census` | Census theoros scans generalized categories, finds rites/internal/ADRs via `.know/architecture.md` navigation. Source_scope in INDEX.md uses Go-detected paths. Seed gate passes (if seeds fresh) or triggers gate (if stale). |
| **Go microservice** (go.mod, cmd/, internal/) | Run `/know --scope=feature --census` | Census discovers features from Go packages and any docs/ content. No errors from missing rites/ or ADR paths. Source_scope uses Go-detected paths. |
| **Next.js app** (package.json, src/, app/) | Run `/know --scope=feature --census` | Census discovers features from TS modules and package.json scripts. Source_scope uses TS-detected paths. |
| **Python ML pipeline** (pyproject.toml, src/) | Run `/know --scope=feature --census` | Census discovers features from Python modules. Source_scope uses Python-detected paths. |
| **Bare project** (no recognized manifest) | Run `/know --scope=feature --census` | Census uses fallback `./src/**/*`. Theoros reports reduced confidence due to limited source material. |
| **Any project, stale seeds** | Run `/know --scope=feature` with stale `.know/` files | Seed freshness gate triggers, shows stale domains, offers 3 options. Does NOT proceed silently. |
| **Any project, missing seeds** | Run `/know --scope=feature` with no `.know/` files | Seed freshness gate triggers, shows all 5 as MISSING. Recommends `/know --all` first. |
| **Any project, --force** | Run `/know --scope=feature --force` with stale seeds | Seed gate is SKIPPED. Census proceeds without freshness check. |
| **Any project, fresh seeds** | Run `/know --scope=feature` with all seeds fresh | Seed gate passes silently. Census proceeds normally. |

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Census theoros produces lower quality results without knossos-specific path hints | Medium | Medium | The `.know/architecture.md` dependency provides equivalent structural guidance. The theoros has 150 turns to explore the project. Quality may vary by project complexity, not by criteria specificity. |
| Seed freshness gate adds friction to first-time feature analysis | High (every first run) | Low | The gate is soft -- user can "proceed anyway". The friction is intentional: running feature analysis without codebase seeds genuinely produces worse results. |
| Language detection misidentifies polyglot projects | Low | Low | The detection is for source_scope staleness checking only. The census theoros reads the actual project regardless of what source_scope says. Worst case: staleness check produces false positives/negatives for census freshness. |

---

## Open Questions: None

All design decisions are resolved. No TBD items remain.

---

## Handoff to Integration Engineer

The Integration Engineer should implement changes to these 3 files in this order:

1. **`/Users/tomtenuta/Code/knossos/rites/shared/mena/pinakes/domains/feature-census.lego.md`** -- Generalize source categories and triage heuristics (content rewrite, no structural changes)
2. **`/Users/tomtenuta/Code/knossos/rites/shared/mena/pinakes/domains/feature-knowledge.lego.md`** -- Generalize source references (content rewrite, no structural changes)
3. **`/Users/tomtenuta/Code/knossos/rites/shared/mena/know/INDEX.dro.md`** -- Insert seed freshness gate + update INDEX.md source_scope template + update theoros scope reminders

No Go code changes required. No new files required. No schema changes required.
