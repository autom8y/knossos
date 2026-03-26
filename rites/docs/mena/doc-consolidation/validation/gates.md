---
description: "Quality Gates Between Phases companion for validation skill."
---

# Quality Gates Between Phases

> Mandatory checkpoints between consolidation phases. No phase proceeds until its gate passes.

## Gate Flow

```
Phase 0        Phase 1        Phase 2        Phase 3        Phase 4
Discovery  --> Extraction --> Synthesis  --> Review     --> Archive
           |              |              |              |
    [Manifest Gate] [Extraction Gate] [Synthesis Gate] [Review Gate] [Archive Gate]
                                                                            |
                                                                     [Validation Gate]
                                                                            |
                                                                        COMPLETE
```

---

## Manifest Gate (Phase 0 → Phase 1)

| ID | Check | Severity |
|----|-------|----------|
| MG-1 | All source files categorized into topics | Blocking |
| MG-2 | No file appears in multiple topics as primary | Blocking |
| MG-3 | No orphan files | Blocking |
| MG-4 | Token estimates provided for all files | Blocking |
| MG-5 | Primary source identified per topic | Blocking |
| MG-6 | Ambiguous mappings flagged for resolution | Warning |
| MG-7 | File hashes computed and recorded | Blocking |
| MG-8 | Topic dependencies form acyclic graph | Blocking |

**On pass**: log timestamp, update checkpoint to `extraction_ready`.
**On fail**: block Phase 1 progression.

---

## Extraction Gate (Phase 1 → Phase 2)

| ID | Check | Severity |
|----|-------|----------|
| EG-1 | All source content accounted for in extraction | Blocking |
| EG-2 | Conflicts explicitly identified and categorized | Blocking |
| EG-3 | Authoritative sources marked per section | Blocking |
| EG-4 | Key concepts extracted with definitions | Blocking |
| EG-5 | Blocking conflicts resolved or escalated | Blocking |
| EG-6 | Source hashes match (no stale sources) | Blocking |
| EG-7 | Extraction token count < 2000 | Warning |
| EG-8 | Synthesis notes provided | Warning |

**On pass**: log timestamp, update checkpoint to `synthesis_ready`.
**On fail**: block Phase 2 progression.

---

## Synthesis Gate (Phase 2 → Phase 3)

| ID | Check | Severity |
|----|-------|----------|
| SG-1 | Every key concept from extraction preserved | Blocking |
| SG-2 | Conflicts resolved in content (not dropped) | Blocking |
| SG-3 | Consolidated doc smaller than sources | Blocking |
| SG-4 | Cross-references use consolidated paths | Blocking |
| SG-5 | Structure follows extraction outline | Warning |
| SG-6 | Code examples included and attributed | Warning |
| SG-7 | Revision history initialized | Warning |
| SG-8 | Self-review completed by synthesis agent | Warning |

**Concept verification script:**

```bash
concepts=$(yq '.shared_concepts[].concept' "$EXTRACTION")
missing=0
for concept in $concepts; do
  grep -qi "$concept" "$CONSOLIDATED" || { echo "MISSING: $concept"; ((missing++)); }
done
[ $missing -eq 0 ] && echo "PASS" || { echo "FAIL: $missing missing"; exit 1; }
```

**On pass**: log timestamp, update checkpoint to `review_ready`.
**On fail**: block Phase 3 progression.

---

## Review Gate (Phase 3 → Phase 4)

| ID | Check | Severity |
|----|-------|----------|
| RG-1 | Review requested from appropriate stakeholder | Blocking |
| RG-2 | Feedback addressed or documented | Blocking |
| RG-3 | Technical accuracy verified | Blocking |
| RG-4 | Approval recorded in checkpoint | Blocking |
| RG-5 | Final version marked in document | Warning |

**On pass**: log timestamp, update checkpoint to `archive_ready`.
**On fail**: block Phase 4 progression.

---

## Archive Gate (Phase 4 → Validation)

| ID | Check | Severity |
|----|-------|----------|
| AG-1 | INDEX.md complete with all mappings | Blocking |
| AG-2 | All originals staged for archive | Blocking |
| AG-3 | Navigation updated to new paths | Blocking |
| AG-4 | No broken links in repository | Blocking |
| AG-5 | Archive directory structure prepared | Blocking |
| AG-6 | Rollback plan documented | Warning |

**On pass**: execute archive operations, proceed to Validation Gate.
**On fail**: do not execute archive operations.

---

## Validation Gate (Phase 4 Complete)

| ID | Check | Severity |
|----|-------|----------|
| VG-1 | Spot-check 3 random sections for info loss | Blocking |
| VG-2 | All skill references updated | Blocking |
| VG-3 | Context file navigation correct | Blocking |
| VG-4 | Token reduction achieved | Blocking |
| VG-5 | Staleness detection functional | Warning |
| VG-6 | Consolidation metrics logged | Warning |

**Spot-check method**: Select 3 random `canonical_sections` from extraction → locate in consolidated doc → verify `key_points` present.

**On pass**: mark consolidation COMPLETE, update manifest status.
**On fail**: escalate for remediation.

---

## Gate Failure Recovery

| Gate | Common Failures | Recovery Action |
|------|-----------------|-----------------|
| Manifest | Orphan files | Add to topic or exclusions |
| Manifest | Dependency cycle | Restructure topic dependencies |
| Extraction | Stale sources | Recompute hashes, re-extract |
| Extraction | Blocking conflict | Escalate to human decision |
| Synthesis | Missing concepts | Update synthesis from extraction |
| Synthesis | Token increase | Further summarization needed |
| Review | Changes requested | Address feedback, re-submit |
| Archive | Broken links | Update references before archive |
| Validation | Info loss detected | Revert archive, fix synthesis |

---

## Gate Log Format

```yaml
gate_log:
  - gate: manifest
    status: passed
    timestamp: "2024-12-25T10:00:00Z"
    agent: "discovery-agent-v1"
    notes: "All 12 files categorized into 4 topics"
```

---

*Part of the [Documentation Consolidation Workflow](../INDEX.lego.md)*
