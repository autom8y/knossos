---
description: "Pre-Archive Validation Checklist companion for validation skill."
---

# Pre-Archive Validation Checklist

> Comprehensive validation checks before archiving original files. Ensures information preservation, link integrity, and schema compliance.

## Validation Execution Order

```
Phase 1: Schema Compliance (SC-1, SC-2, SC-3)
    |
    v  All pass? Continue : Fix and re-run
    |
Phase 2: Information Preservation (IP-1, IP-2, IP-3, IP-4)
    |
    v  All pass? Continue : Fix and re-run
    |
Phase 3: Cross-Reference Resolution (CR-1, CR-2, CR-3)
    |
    v  All pass? Continue : Fix and re-run
    |
Phase 4: Link Integrity (LI-1, LI-2, LI-3)
    |
    v  All pass? ARCHIVE APPROVED : Fix and re-run
```

---

## Schema Compliance

| ID | Check | Severity | Pass Criteria |
|----|-------|----------|---------------|
| SC-1 | MANIFEST.yaml validity | Blocking | All required fields present; all paths exist; all hashes match |
| SC-2 | Extraction file validity | Blocking | schema_version 1.0; topic.id matches filename; hashes current; canonical_sections non-empty |
| SC-3 | Checkpoint consistency | Significant | status.phase = archive/complete; no blocked state; all sections final |

---

## Information Preservation

| ID | Check | Severity | Pass Criteria |
|----|-------|----------|---------------|
| IP-1 | Key concept coverage | Blocking | 100% of shared_concepts and key_points present in consolidated doc |
| IP-2 | Conflict resolution preservation | Blocking | 0 blocking conflicts unresolved; all decisions applied |
| IP-3 | Source section accounting | Blocking | Every source H1/H2 mapped to consolidated location or explicitly excluded |
| IP-4 | Code example preservation | Significant | All unique code examples preserved or superseded with justification |

**IP-3 audit template:**

| Source File | Section Heading | Status | Destination | Rationale |
|-------------|-----------------|--------|-------------|-----------|
| {path} | {heading} | preserved / moved / removed | {location} | {reason if removed} |

**IP-2 verification table:**

| Conflict ID | Resolution Decision | Applied in Consolidated? | Location |
|-------------|---------------------|--------------------------|----------|
| conflict-001 | {decision} | YES / NO | Section/line |

---

## Cross-Reference Resolution

| ID | Check | Severity | Pass Criteria |
|----|-------|----------|---------------|
| CR-1 | Topic dependency satisfaction | Blocking | All topic dependencies consolidated first; no circular deps |
| CR-2 | Skill reference updates | Blocking | All skill references point to consolidated docs; triggers updated |
| CR-3 | Navigation integrity | Blocking | All context file paths exist; no orphan documents |

---

## Link Integrity

| ID | Check | Severity | Pass Criteria |
|----|-------|----------|---------------|
| LI-1 | Internal cross-references | Blocking | 0 broken internal links |
| LI-2 | External reference validity | Minor | All URLs syntactically valid (HTTP checks advisory) |
| LI-3 | Archive path references | Blocking | 0 references to soon-to-be-archived paths |

**LI-1 automated check:**

```bash
grep -oE '\[.+?\]\(.+?\)' consolidated-doc.md | \
while read link; do
  target=$(echo "$link" | sed 's/.*(\(.*\))/\1/')
  if [[ "$target" == \#* ]]; then
    anchor="${target#\#}"
    grep -qi "^#.*$anchor" consolidated-doc.md || echo "BROKEN: $link"
  elif [[ -f "$target" ]]; then
    echo "OK: $link"
  else
    echo "BROKEN: $link"
  fi
done
```

**LI-3 search:**

```bash
for file in $(cat archive-list.txt); do
  grep -rn "$file" --include="*.md" . | grep -v "^\.archive/"
done
```

---

## Quick Reference Checklist

```markdown
## Pre-Archive Validation - {Topic Name}

**Date**: {YYYY-MM-DD}
**Validator**: {agent_or_person}

### Schema Compliance
- [ ] SC-1: MANIFEST.yaml valid
- [ ] SC-2: Extraction files valid
- [ ] SC-3: Checkpoints consistent

### Information Preservation
- [ ] IP-1: Key concepts covered (100%)
- [ ] IP-2: Conflict resolutions applied
- [ ] IP-3: Source sections accounted
- [ ] IP-4: Code examples preserved

### Cross-Reference Resolution
- [ ] CR-1: Topic dependencies satisfied
- [ ] CR-2: Skill references updated
- [ ] CR-3: Navigation integrity verified

### Link Integrity
- [ ] LI-1: Internal links resolve
- [ ] LI-2: External URLs valid (advisory)
- [ ] LI-3: Archive path references cleared

### Approval
- [ ] All blocking checks pass
- [ ] Significant issues documented
- [ ] Ready for archive

**Approved by**: _______________
**Date**: _______________
```

---

## Failure Modes and Recovery

| Failure | Impact | Recovery |
|---------|--------|----------|
| Missing key concept | Information loss | Re-extract from source; update synthesis |
| Broken internal link | Navigation failure | Update link or add missing anchor |
| Stale source hash | Extraction outdated | Re-run extraction for affected source |
| Unresolved conflict | Ambiguous content | Escalate to human; resolve before archive |
| Orphan file | Lost documentation | Add to topic or explicit exclusion |

---

*Part of the [Documentation Consolidation Workflow](../INDEX.lego.md)*
