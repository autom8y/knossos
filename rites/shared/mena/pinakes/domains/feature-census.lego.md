---
name: feature-census-criteria
description: "Criteria for feature enumeration census. Use when: theoros is scanning project sources to produce a comprehensive feature taxonomy with GENERATE/SKIP recommendations. Triggers: feature census, feature enumeration, feature inventory, feature taxonomy, feature discovery."
---

# Feature Census Criteria

> The theoros scans lightweight project sources and enumerates ALL identifiable features, producing a structured feature inventory with classification and triage recommendations. This is a census -- completeness of enumeration is the grading target, not feature quality.

## Scope

**Target sources** (theoros must scan all of these):
- `rites/*/manifest.yaml` -- rite descriptions reveal major capability domains
- `internal/*/` -- package directory names map to implementation boundaries
- `docs/decisions/ADR-*.md` -- titles and first paragraphs reveal feature decisions
- `.claude/commands/*.md` -- user-facing command surface
- `.claude/agents/*.md` -- agent descriptions reveal feature behaviors
- `INTERVIEW_SYNTHESIS.md` -- high-level feature overview and golden rules
- `.know/*.md` frontmatter -- existing knowledge domains as feature signals

**Observation focus**: Enumerate every identifiable feature in the project. For each feature, produce a structured entry with slug, name, category, source evidence, complexity rating, GENERATE/SKIP recommendation, and confidence score.

**NOTE**: This domain uses knowledge-capture grading. Grade the COMPLETENESS of the enumeration, not the quality of individual features. A = every identifiable feature cataloged with evidence. F = fewer than half of identifiable features documented.

## Output Format

The census theoros must produce per-feature entries with these fields:

| Field | Type | Description |
|-------|------|-------------|
| `slug` | kebab-case string | Unique identifier (e.g., `materialization`, `session-lifecycle`) |
| `name` | human-readable string | Feature display name |
| `category` | string | Grouping (e.g., "Core Platform", "Context Model", "Workflow", "Tooling") |
| `source_evidence` | string list | Which scanned files reference this feature |
| `complexity` | enum | HIGH / MEDIUM / LOW based on package count, ADR count, command count |
| `recommendation` | enum + rationale | GENERATE / SKIP with brief rationale |
| `confidence` | float | 0.0-1.0 how confident theoros is in this classification |

## Triage Heuristics

The theoros uses these rules to produce GENERATE vs SKIP recommendations:

**GENERATE** if any of:
- 1+ ADRs reference the feature
- 10+ implementation files in relevant packages
- User-facing commands exist for the feature
- Multiple rites depend on the feature

**SKIP** if all of:
- Pure utility (fileutil, paths, string helpers)
- No ADRs reference the feature
- Fewer than 5 implementation files
- Single-rite internal detail with no cross-cutting concerns

## Criteria

### Criterion 1: Source Coverage (weight: 30%)

**What to evaluate**: Did the theoros scan all specified source types? Each source type contributes unique feature signals that cannot be found elsewhere. Incomplete scanning produces incomplete census.

**Evidence to collect**:
- Confirm each of the 7 source types was accessed and scanned
- Record per source type: files scanned count, features discovered from that source
- Note any source types that were inaccessible or empty (explicit absence is evidence)
- Flag if any source type was skipped without justification

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | All 7 source types scanned with evidence | Per-source-type scan summary: files accessed, features found, inaccessible sources documented. No source type skipped. |
| B | 6 of 7 source types scanned | One source type missing but justification provided (e.g., no ADRs exist). All other sources fully scanned with evidence. |
| C | 5 of 7 source types scanned | Two source types missing. Remaining sources scanned but per-source evidence incomplete. |
| D | 4 of 7 source types scanned | Three source types missing. Feature enumeration relies on partial evidence base. |
| F | Fewer than 4 source types scanned | More than half the specified sources not scanned. Census built from insufficient evidence. |

---

### Criterion 2: Feature Enumeration (weight: 30%)

**What to evaluate**: Are all identifiable features listed with slugs and names? The census must capture every distinct feature the project implements, not just the obvious ones.

**Evidence to collect**:
- Count total features enumerated
- Verify each feature has a unique slug (kebab-case) and human-readable name
- Cross-check: do package names, ADR titles, command names, and rite descriptions all have corresponding feature entries?
- Identify any source references that do not map to a feature entry (potential gaps)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% of identifiable features enumerated | Every package cluster, ADR topic, command group, and rite capability maps to a feature entry. Cross-reference table shows no unaccounted source references. Each entry has unique slug and name. |
| B | 80-89% of identifiable features enumerated | Most features captured. Minor gaps where 1-2 source references lack corresponding feature entries. All entries have slug and name. |
| C | 70-79% of identifiable features enumerated | Core features present but peripheral features (utilities, tooling, workflow features) partially missing. Some entries missing slug or name. |
| D | 60-69% of identifiable features enumerated | Only major features listed. Many package clusters and ADR topics unaccounted for. Slugs or names inconsistent. |
| F | Fewer than 60% of identifiable features enumerated | Fewer than half the identifiable features documented. Census is unreliable for planning purposes. |

---

### Criterion 3: Classification Quality (weight: 20%)

**What to evaluate**: Are categories logical, recommendations justified, and complexity ratings evidence-based? Classification must be grounded in scanned evidence, not assumptions.

**Evidence to collect**:
- Verify categories are internally consistent (features in the same category share meaningful properties)
- For each GENERATE recommendation: confirm at least one triage heuristic is met with evidence
- For each SKIP recommendation: confirm all skip conditions are met with evidence
- For complexity ratings: confirm evidence basis (package count, ADR count, command count cited)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% of entries have evidence-based classification | Every feature entry: category justified, recommendation cites specific heuristic match with file evidence, complexity rating cites counts. Categories are coherent groupings. |
| B | 80-89% of entries have evidence-based classification | Most entries fully classified with evidence. Minor gaps where 1-2 entries lack heuristic citation or complexity evidence. Categories mostly coherent. |
| C | 70-79% of entries have evidence-based classification | Categories present but some are catch-all groupings. Recommendations present but heuristic citations inconsistent. Complexity ratings present but evidence sparse. |
| D | 60-69% of entries have evidence-based classification | Categories applied but not justified. Recommendations present without heuristic evidence. Complexity ratings appear arbitrary. |
| F | Fewer than 60% of entries have evidence-based classification | Classifications absent, unjustified, or internally contradictory. |

---

### Criterion 4: Output Format Compliance (weight: 20%)

**What to evaluate**: Does the output follow the structured per-feature entry format? Downstream consumers (the `/know` dispatch and feature-knowledge theoros) depend on a consistent, parseable census format.

**Evidence to collect**:
- Verify every feature entry contains all 7 required fields (slug, name, category, source_evidence, complexity, recommendation, confidence)
- Confirm slug format (kebab-case, no uppercase, no spaces)
- Confirm complexity values are exactly HIGH / MEDIUM / LOW
- Confirm recommendation values are exactly GENERATE / SKIP with rationale string
- Confirm confidence is a float between 0.0 and 1.0

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 100% of entries conform to format | Every entry has all 7 fields. Slug format valid. Enum values match spec exactly. Confidence values in range. Output is machine-parseable. |
| B | 90-99% of entries conform | 1-2 entries with minor format violations (e.g., missing confidence, slug with underscore). All required fields present in remainder. |
| C | 80-89% of entries conform | Multiple entries with format violations. Some missing fields. Enum values mostly correct but occasional non-standard values. |
| D | 70-79% of entries conform | Format inconsistent across entries. Multiple missing fields. Output requires manual cleanup before downstream use. |
| F | Fewer than 70% of entries conform | Format largely ignored. Entries are free-text rather than structured. Output not usable by downstream consumers. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- Source Coverage: A (midpoint 95%) x 30% = 28.5
- Feature Enumeration: B (midpoint 85%) x 30% = 25.5
- Classification Quality: B (midpoint 85%) x 20% = 17.0
- Output Format Compliance: A (midpoint 95%) x 20% = 19.0
- **Total: 90.0 -> A**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [feature-knowledge-criteria](feature-knowledge.lego.md) -- Per-feature knowledge capture (downstream consumer of this census)
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture (complementary domain)
