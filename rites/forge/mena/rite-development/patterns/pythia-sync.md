---
description: "Pythia Synchronization Pattern companion for patterns skill."
---

# Pythia Synchronization Pattern

> Keep the Pythia agent canonical with all ecosystem changes

## Why This Matters

The Pythia agent (`/consult`) is the ecosystem's meta-navigator. Users rely on it for:
- Rite recommendations
- Command guidance
- Workflow playbooks
- Ecosystem navigation

**If Pythia has stale data, users get wrong guidance.**

---

## Knowledge Base Structure

Pythia knowledge is maintained in rite mena directories and shared skills:

- `mena/navigation/consult/` -- Pythia dromena and reference
- `prompting` skill -- Invocation patterns
- `10x-workflow` skill -- Phase transitions and quality gates
- Rite manifests at `$KNOSSOS_HOME/rites/*/orchestrator.yaml`

---

## Synchronization Matrix

| Change Type | Files to Update |
|-------------|-----------------|
| **New rite** | ecosystem-map, agent-reference, rite-profile (new), intent-patterns, decision-trees, complexity-matrix, command-reference |
| **New agent to existing rite** | agent-reference, rite-profile |
| **New command** | command-reference, ecosystem-map |
| **Workflow change** | rite-profile, agent-reference (if phases change) |
| **Rename rite** | All files referencing old name |
| **Remove rite** | Remove from all files, delete rite-profile |
| **New playbook** | Create in `.ledge/spikes/` |

---

## Step-by-Step: Adding a New Rite

### 1. Update ecosystem-map.md

Add to Rites table:
```markdown
| **{rite}** | `/{rite}` | {N} | {Brief description} |
```

Update counts:
```markdown
**Total Agents**: {new count} across all rites
```

### 2. Update rite manifest

Ensure `rites/{rite}/orchestrator.yaml` contains accurate agent/workflow/complexity data. This is the source of truth for rite routing and capability declaration.

### 3. Verify quick-switch command

Ensure `mena/navigation/{rite}.dro.md` or equivalent quick-switch dromenon exists and routes correctly.

### 4. Verify rite integration

Run `ari rite list` to confirm the new rite appears in the catalog with correct metadata.

---

## Step-by-Step: Modifying Existing Rite

### Agent Added

1. **agent-reference.md**: Add agent to rite section
2. **rite-profiles/{rite}.md**: Update Agents table

### Agent Removed

1. **agent-reference.md**: Remove agent from rite section
2. **rite-profiles/{rite}.md**: Update Agents table

### Workflow Changed

1. **rite-profiles/{rite}.md**: Update Workflow diagram
2. **agent-reference.md**: Update workflow summary if needed

### Complexity Levels Changed

1. **rite-profiles/{rite}.md**: Update Complexity Levels table
2. **routing/complexity-matrix.md**: Update rite section

---

## Step-by-Step: Adding Playbook

### 1. Identify the scenario

Common playbooks cover:
- Feature development
- Bug fixes
- Code quality
- Documentation
- Security
- Performance
- Tech debt
- Incidents

### 2. Create playbook file

Location: playbook files are maintained in rite mena directories or shared skills.

Use format:
```markdown
# Playbook: {Name}

> {One-line description}

## When to Use
- {Trigger 1}
- {Trigger 2}

## Prerequisites
- {Prereq 1}

## Command Sequence

### Phase 1: {Name}
```bash
/{command} {args}
```
**Expected output**: {Description}
**Decision point**: {When to proceed vs adjust}

### Phase 2: {Name}
...

## Variations
- **{Variant}**: {Adjustment}

## Success Criteria
- [ ] {Criterion 1}
- [ ] {Criterion 2}
```

### 3. Update consult-ref skill

Add playbook to the list in `mena/navigation/consult/reference.md`:
```markdown
**Curated Playbooks**: Pre-authored sequences for common scenarios
- `{playbook}.md`
```

---

## Verification Commands

After any sync:

```bash
# Verify rite appears in catalog
ari rite list

# Verify rite manifest exists
cat $KNOSSOS_HOME/rites/{rite-name}/orchestrator.yaml

# Verify quick-switch command exists
ls $KNOSSOS_HOME/mena/navigation/{rite-name}.dro.md 2>/dev/null || echo "No quick-switch"

# Verify rite loads
ari sync --rite {rite-name} --dry-run
```

---

## Common Issues

### Forgot to update ecosystem-map

**Symptom**: `/consult --rite` shows wrong count
**Fix**: Update rite count and add missing rite to table

### Missing rite profile

**Symptom**: `/consult` can't give detailed rite guidance
**Fix**: Create rite-profiles/{rite}.md

### Stale routing patterns

**Symptom**: `/consult "query"` doesn't route to new rite
**Fix**: Add keywords to intent-patterns.md and decision-trees.md

### Wrong agent count

**Symptom**: Pythia says rite has N agents, but it has M
**Fix**: Update agent-reference.md and ecosystem-map.md

---

## Automation Opportunity

Future enhancement: Create a validation script that:
1. Scans knossos for all rites
2. Compares against Pythia knowledge
3. Reports discrepancies
4. Suggests updates

```bash
# Future: $KNOSSOS_HOME/validate-pythia.sh
# Would check all rites are reflected in knowledge base
```

---

## Related

- [INDEX.lego.md](../INDEX.lego.md) - Rite development overview
- [validation/validation.md](../validation/validation.md) - Validation checklist
- [consult](../../../../mena/navigation/consult/INDEX.dro.md) - Pythia reference
