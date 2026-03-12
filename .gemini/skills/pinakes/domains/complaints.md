---
name: complaints-criteria
description: "Evaluation criteria for the Cassandra complaint pipeline audit. Use when: theoros is auditing the complaints domain, evaluating complaint filing health, schema compliance, and resolution rates. Triggers: complaints audit criteria, complaint pipeline evaluation, cassandra assessment."
---

# Complaints Audit Criteria

> The theoros evaluates the Cassandra complaint pipeline against these standards. Early-stage installations will grade low — that is expected and correct. The value is establishing a baseline to track improvement.

## Scope

**Target files**: `.sos/wip/complaints/*.yaml`

**Supporting context**: `internal/validation/schemas/complaint.schema.json` (required fields and allowed values), `rites/shared/mena/complaint-filing/SKILL.md` (filing protocol)

**Evaluation focus**: Six dimensions of complaint pipeline health — volume, source diversity, severity distribution, tag emergence, schema compliance, and resolution rate.

**Note on early-stage grading**: With 0-10 complaints, most criteria will grade F or D. This is expected and correct. The audit establishes a baseline, not a passing score. A low grade on this domain in early usage is informative, not a failure.

## Criteria

### Criterion 1: Filing Volume (weight: 20%)

**What to evaluate**: Whether complaints are being filed at all. Zero complaints means the mechanism is not being used.

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 30+ complaints | High filing activity. Mechanism is embedded in agent workflows. |
| B | 16-29 complaints | Regular filing activity. Multiple agents contributing. |
| C | 6-15 complaints | Moderate activity. Mechanism is being used. |
| D | 1-5 complaints | Minimal activity. Filed but not yet habitual. |
| F | 0 complaints | No complaints filed. Pipeline not yet active. |

**Evidence collection**: Count YAML files matching `COMPLAINT-*.yaml` in `.sos/wip/complaints/`. Report the count. If directory is absent, grade F with note "directory not created."

---

### Criterion 2: Source Diversity (weight: 20%)

**What to evaluate**: Whether complaints come from multiple filers. Single-source complaints indicate only one agent or hook is exercising the mechanism.

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 5+ distinct filers | Broad adoption across agent types. |
| B | 4 distinct filers | Good coverage. Most active agents contributing. |
| C | 3 distinct filers | Moderate coverage. Some spread. |
| D | 2 distinct filers | Limited diversity. Only 2 sources. |
| F | 0-1 distinct filers | No diversity. Single filer or no files. |

**Evidence collection**: Extract `filed_by` field from each complaint YAML. Count distinct values. List the filers found.

---

### Criterion 3: Severity Distribution (weight: 15%)

**What to evaluate**: Whether complaint severity reflects a healthy mix — not all low (signals under-reporting of real friction) and not all critical (signals alarm fatigue).

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 3-4 severity levels present, no single level > 70% | Healthy distribution reflecting calibrated filing. |
| B | 2-3 severity levels present, no single level > 80% | Reasonable distribution with minor skew. |
| C | 2 severity levels present, or one level 80-89% | Somewhat concentrated. Calibration may be off. |
| D | 1-2 severity levels present, one level 90-99% | Heavily concentrated. Miscalibration likely. |
| F | 0 complaints, or 100% single-severity | No data or completely uniform. |

**Evidence collection**: Extract `severity` field from each complaint YAML. Count by level (low, medium, high, critical). Report distribution and identify the dominant level's percentage.

---

### Criterion 4: Tag Emergence (weight: 15%)

**What to evaluate**: Whether complaints include tags and whether tags show clustering patterns. Empty tags mean no emergent taxonomy is forming.

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 70%+ complaints have non-empty tags AND 2+ tags appear in 3+ complaints | Strong tag emergence. Patterns forming. |
| B | 50-69% complaints have non-empty tags AND at least 1 tag recurring | Moderate emergence. Some clustering visible. |
| C | 30-49% complaints have non-empty tags | Early emergence. Some agents tagging. |
| D | 1-29% complaints have non-empty tags | Minimal tagging. Tags present but rare. |
| F | 0% complaints have non-empty tags, or no complaints | No tag data. |

**Evidence collection**: For each complaint YAML, check `tags` field. Count complaints with non-empty tags. List the most common tags and their occurrence counts.

---

### Criterion 5: Schema Compliance (weight: 15%)

**What to evaluate**: Whether complaint files contain all required fields with valid values. Non-compliant complaints indicate filing protocol drift.

Required fields (quick-file tier): `id`, `filed_by`, `filed_at`, `title`, `severity`, `description`, `status`

Valid values: `severity` in (low, medium, high, critical), `status` in (filed, triaged, accepted, rejected, resolved)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 100% compliance | All files have all required fields with valid values. |
| B | 90-99% compliance | 1 file with minor issue. |
| C | 75-89% compliance | A few files with issues. Protocol partially followed. |
| D | 50-74% compliance | Significant compliance failures. Filing protocol not understood. |
| F | < 50% compliance, or no complaints | Most files non-compliant or no data. |

**Evidence collection**: For each complaint YAML, check for presence of all 7 required fields and that `severity` and `status` values are from the allowed enums. List non-compliant files with specific missing/invalid fields.

---

### Criterion 6: Resolution Rate (weight: 15%)

**What to evaluate**: The ratio of resolved complaints to total complaints. A healthy pipeline moves complaints through triage to resolution.

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 80%+ resolved | Most complaints have been addressed. Active triage loop. |
| B | 60-79% resolved | Good resolution rate. Active but not complete. |
| C | 40-59% resolved | Moderate resolution. Some backlog. |
| D | 20-39% resolved | Low resolution. Backlog accumulating. |
| F | < 20% resolved, or no complaints | Almost no resolution, or pipeline is pre-triage. |

**Note**: In early-stage (Phase 1), 0% resolution is expected. Phase 2 delivers `/reflect` triage capability. Do not penalize a fresh installation for lacking a triage pipeline.

**Evidence collection**: Extract `status` field from each complaint YAML. Count those with `status: resolved`. Compute ratio: resolved / total. List non-resolved complaints by status breakdown.

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores. Example (early-stage baseline):

- Filing Volume: F (0 complaints, midpoint 30%) x 20% = 6.0
- Source Diversity: F (0 distinct filers, midpoint 30%) x 20% = 6.0
- Severity Distribution: F (no data, midpoint 30%) x 15% = 4.5
- Tag Emergence: F (no data, midpoint 30%) x 15% = 4.5
- Schema Compliance: F (no data, midpoint 30%) x 15% = 4.5
- Resolution Rate: F (no data, midpoint 30%) x 15% = 4.5
- **Total: 30.0 -> F (baseline, expected for Phase 1)**

## Related

- [Pinakes INDEX](../INDEX.md) - Full audit system documentation
- [complaint-filing skill](../../../../rites/shared/mena/complaint-filing/SKILL.md) - Filing protocol and schema
- [hooks-criteria](hooks.md) - Hook wiring audit (driftdetect hook contributes to filing volume)
