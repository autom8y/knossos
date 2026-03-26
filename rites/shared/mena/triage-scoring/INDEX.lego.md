---
name: triage-scoring
description: "Cassandra Protocol triage scoring model for complaint prioritization. Use when: triaging complaints, scoring complaint severity, determining complaint priority, classifying zone impact, cross-referencing against scar tissue, deciding auto-accept vs human-review thresholds. Triggers: triage, scoring, reflect, complaint priority, threshold, zone classification."
---

# Triage Scoring Model (Cassandra Protocol)

> Reference for scoring, prioritizing, and routing complaints filed via the complaint-filing skill. Used by `/reflect` triage pipeline.

## Quick-File vs Deep-File Detection

Before scoring, classify each complaint as **quick-file** or **deep-file**:

| Format | Detection Heuristic | Scoring Path |
|--------|-------------------|--------------|
| **Quick-file** | `filed_by == "drift-detector"` AND `zone` is empty AND `evidence` is nil | Noise-review track (3 dimensions) |
| **Deep-file** | Has `zone`, `effort_estimate`, `evidence`, or `related_scars` | Standard track (6 dimensions) |
| **Agent quick-file** | `filed_by != "drift-detector"` AND no deep-file fields | Standard track (6 dimensions, with zone default = behavior for routing override) |

**Agent quick-file zone handling**: Agent quick-files use two independently conservative defaults that serve different purposes. `zone_impact = 30` (parameter-level score) prevents quick-files from inflating composite scores without evidence. `zone default = behavior` (routing override) forces human review when the fix zone is unknown. These are intentionally decoupled: the score conservatively undersells risk while the routing conservatively oversells it. A filer providing explicit `zone: parameter` metadata overrides both defaults.

### Noise-Review Track (Quick-File Only)

Quick-file complaints use simplified 3-dimension scoring:

| Dimension | Weight | Scoring Rule |
|-----------|--------|--------------|
| **Severity** | 40% | Same as standard (low=20, medium=45, high=70, critical=95) |
| **Recurrence** | 40% | Same as standard (1=15, 2=40, 3-4=65, 5+=90) |
| **Source Diversity** | 20% | Same as standard (1 filer=20, 2=50, 3+=80) |

Zone Impact, Scar-Tissue Match, and Effort-to-Impact are **skipped** (these fields are absent in quick-file format). No zone override applies — quick-file complaints route purely by threshold band.

**Noise-review threshold bands** (same as standard):
- 0-39: auto-reject (typical for single low-severity tool-fallback)
- 40-69: auto-accept
- 70+: escalate to standard track for full 6-dimension scoring

See [scoring-example.lego.md](scoring-example.lego.md) for worked arithmetic on both cases.

## Scoring Dimensions (Standard Track)

Every complaint on the standard track is scored on 6 dimensions. Each dimension produces a 0-100 sub-score. The final triage score is a weighted sum.

| Dimension | Weight | What It Measures |
|-----------|--------|------------------|
| **Severity** | 25% | Impact magnitude from complaint schema (low/medium/high/critical) |
| **Recurrence** | 20% | How many times similar friction has been observed (dedup signal) |
| **Zone Impact** | 20% | Which modification zone is affected (parameter/behavior/structure) |
| **Scar-Tissue Match** | 15% | Whether complaint maps to a known SCAR in `.know/scar-tissue.md` |
| **Effort-to-Impact** | 10% | Ratio of estimated fix effort to expected improvement |
| **Source Diversity** | 10% | Number of distinct filers reporting similar friction |

### Dimension Scoring Rules

**Severity** (from complaint `severity` field):
- `low` = 20, `medium` = 45, `high` = 70, `critical` = 95

**Recurrence** (from dedup count and corroborating notes):
- 1 observation = 15, 2 = 40, 3-4 = 65, 5+ = 90

**Zone Impact** (from complaint `zone` field):

**CRITICAL**: The `zone` field refers to the **modification zone of the proposed fix**, not the system exhibiting the symptom. A complaint about lock behavior whose fix is a skill documentation edit has `zone: parameter` (the fix modifies a parameter/prompt, not the lock system). Documentation-only fixes are always `parameter` zone regardless of the system being documented.

- `parameter` = 30, `behavior` = 60, `structure` = 90
- Missing zone for **deep-file** complaints = 45 (default to behavior-level review)
- Missing zone for **agent quick-file** complaints (non-drift-detector, no deep-file fields) = 30 (default to parameter — less restrictive)
- Note: drift-detector quick-file complaints never reach this dimension (they use the noise-review track)

**Scar-Tissue Match** (cross-reference result):
- No match = 20, match to fixed SCAR = 40, match to OPEN scar = 75, regression of fixed SCAR = 95

**Effort-to-Impact** (from `effort_estimate` when present):
- `trivial` fix + high impact = 90, `epic` fix + low impact = 10
- When `effort_estimate` absent: default 50 (neutral)

**Source Diversity** (distinct `filed_by` values for similar complaints):
- 1 filer = 20, 2 filers = 50, 3+ filers = 80

## Threshold Bands

The final weighted score determines the routing action:

| Score Range | Action | Rationale |
|-------------|--------|-----------|
| **0-39** | Auto-reject | Noise, duplicate, stale, or low-impact with high effort |
| **40-69** | Auto-accept (parameter zone only) | Trivial parameter knob changes safe for auto-application |
| **70-84** | Human review required | Behavior-zone changes need human judgment on prompt/routing edits |
| **85-100** | ADR required | Structural implications demand formal architectural decision |

## Zone Interaction Rules

Zone classification overrides threshold-based routing when the zone demands stricter review:

| Zone | Override Behavior |
|------|-------------------|
| `parameter` | Standard threshold routing. Scores 40+ auto-accept. |
| `behavior` | Human review required regardless of score. A behavior complaint scoring 45 still requires human review, not auto-accept. |
| `structure` | ADR required regardless of score. A structure complaint scoring 50 still requires an ADR, not auto-accept or human review. |

The zone override only elevates review level, never reduces it. A parameter complaint scoring 85+ still requires an ADR per the threshold band.

## Cross-Reference Protocol

Check every complaint against `.know/scar-tissue.md` before scoring.

**Step 1: Match detection.** Search the Failure Catalog for SCAR entries matching the complaint's description, tags, or affected file paths. Match on: fix location overlap, category overlap, or behavioral pattern similarity. **RELATED requires mechanism overlap, not just domain overlap.** Two items sharing the same component (e.g., both involve the lock system) but describing different failure mechanisms (e.g., TOCTOU race vs CLI session resolution) are **no-match**, not RELATED.

**Step 2: Scoring adjustment based on match type.**

| Match Type | Score Effect | Notation |
|------------|-------------|----------|
| **Regression** (SCAR fix has regressed) | Scar-tissue dimension = 95 (urgency boost) | Add `scar_ref: SCAR-NNN (REGRESSION)` to triage output |
| **Known-and-fixed** (complaint describes a pattern already fixed by a SCAR) | Scar-tissue dimension = 40 (novelty reduction) | Add `scar_ref: SCAR-NNN (FIXED)` to triage output |
| **Related** (complaint is adjacent to but distinct from a SCAR) | Scar-tissue dimension = 60 (moderate signal) | Add `scar_ref: SCAR-NNN (RELATED)` to triage output |
| **No match** | Scar-tissue dimension = 20 (baseline) | No `scar_ref` notation |

**Step 3: Linkage.** Include `scar_ref` in every triage entry when matched. This creates a bidirectional trace for future scar-tissue regeneration.

## Dedup Constraint

Each complaint belongs to **exactly one** dedup group. Use best-fit by combined tag+title similarity. Multi-group assignment inflates group counts and creates ambiguous routing.

## Triage Output Format

Each triaged complaint produces a summary entry in `.sos/wip/TRIAGE-complaints.md`:

```yaml
- id: COMPLAINT-{id}
  score: {0-100}
  action: auto-reject | auto-accept | human-review | adr-required
  zone: parameter | behavior | structure
  scar_ref: "SCAR-NNN (TYPE)" # if matched
  dimension_scores:
    severity: {0-100}
    recurrence: {0-100}
    zone_impact: {0-100}
    scar_tissue_match: {0-100}
    effort_to_impact: {0-100}
    source_diversity: {0-100}
  rationale: "{one-line explanation of routing decision}"
```

## Companion Reference

| Topic | File | When to Load |
|-------|------|--------------|
| Worked scoring examples | [scoring-example.lego.md](scoring-example.lego.md) | Verifying arithmetic or calibrating intuition |
