---
description: "Assessment: {project-name} companion for templates skill."
---

# Assessment: {project-name}

## Health Grades

| Category | Grade | Rationale |
|----------|-------|-----------|
| Complexity | {A-F} | {one-line justification} |
| Testing | {A-F} | {one-line justification} |
| Dependencies | {A-F} | {one-line justification} |
| Structure | {A-F} | {one-line justification} |
| Hygiene | {A-F} | {one-line justification} |
| **Overall** | **{A-F}** | **{one-line justification}** |

<!-- Overall grade uses weakest-link model:
     1. Start with median grade across five categories
     2. Any F -> overall cannot exceed D
     3. Any D -> overall cannot exceed C
     4. 3+ categories at C or below -> drop one letter -->

## Validated Findings

<!-- For each finding from the scan, verify by spot-checking the referenced file.
     Dismiss false positives with explanation. Assign severity. -->

### Critical
<!-- Blocks correctness or security. F risk for the category. -->

#### {Finding title}
- **Location**: {path/to/file:line}
- **Scan signal**: {reference to original signal from SCAN-{slug}.md}
- **Verification**: {what you checked to confirm this is real}
- **Description**: {clear explanation of the issue}
- **Impact**: {why this matters}
- **Recommendation**: {actionable next step}
- **Effort**: {small/medium/large}
- **Cross-rite**: {target rite if applicable, or "none"}

### High
<!-- Significant maintainability risk. Drops category by 1-2 grades. -->

### Medium
<!-- Improvement opportunity. Drops category by 0-1 grades. -->

### Low
<!-- Nice-to-have. Informational, no grade impact. -->

## Dismissed Signals

<!-- Scan signals determined to be false positives after verification. -->

| Signal | Location | Reason Dismissed |
|--------|----------|-----------------|
| {signal from scan} | {location} | {why this is a false positive} |

## Patterns Identified

<!-- Cross-cutting themes that emerge from connecting multiple signals.
     These are higher-order observations, not individual findings. -->

1. **{Pattern name}**: {description of the cross-cutting theme, citing 2+ related findings}

## Cross-Rite Routing Recommendations

| Finding | Target Rite | Trigger Signal |
|---------|-------------|----------------|
| {finding title} | {rite name} | {why this rite should investigate} |

## Coverage Gaps

<!-- Areas the signal-sifter did not scan or insufficiently covered.
     If significant, this section triggers a back-route to signal-sifter.
     Mark as BACK-ROUTE REQUESTED if action needed. -->

| Area | Gap Description | Severity |
|------|----------------|----------|
| {area} | {what was missed} | {minor/significant -- significant triggers back-route} |

---
*Produced by pattern-profiler | Review rite | FULL mode*
