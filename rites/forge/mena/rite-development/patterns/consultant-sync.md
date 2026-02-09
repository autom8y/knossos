# Consultant Synchronization Pattern

> Keep the Consultant agent canonical with all ecosystem changes

## Why This Matters

The Consultant agent (`/consult`) is the ecosystem's meta-navigator. Users rely on it for:
- Rite recommendations
- Command guidance
- Workflow playbooks
- Ecosystem navigation

**If the Consultant has stale data, users get wrong guidance.**

---

## Knowledge Base Structure

```
.claude/knowledge/consultant/
├── ecosystem-map.md          # Complete ecosystem overview
├── command-reference.md      # All commands
├── agent-reference.md        # All agents
├── routing/
│   ├── intent-patterns.md    # Natural language → rite/command
│   ├── decision-trees.md     # Structured routing logic
│   └── complexity-matrix.md  # Scope-based selection
├── rite-profiles/
│   └── {rite}.md             # One per rite
└── playbooks/curated/
    └── {scenario}.md         # Common workflow playbooks
```

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
| **New playbook** | Create in playbooks/curated/ |

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

### 2. Update agent-reference.md

Add new section:
```markdown
## {rite} ({N} agents)

| Agent | Model | Phase | Produces |
|-------|-------|-------|----------|
| **{agent-1}** | {model} | {phase} | {artifact} |
| **{agent-2}** | {model} | {phase} | {artifact} |
...

**Workflow**: {phase-1} → {phase-2} → {phase-3} → {phase-4}
```

### 3. Create rite-profiles/{rite}.md

Use template from INDEX.lego.md. Include:
- Overview
- Switch Command
- Agents table
- Workflow diagram
- Complexity Levels
- Best For / Not For
- Quick Start
- Related Commands

### 4. Update routing/intent-patterns.md

Add intent patterns for new rite domain:
```markdown
## {Domain} Intents

| User Says | Likely Need | Recommended |
|-----------|-------------|-------------|
| "{keyword 1}" | {description} | `/{rite}` → `/task` |
| "{keyword 2}" | {description} | `/{rite}` → `/task` |
```

### 5. Update routing/decision-trees.md

Add to Primary Router:
```markdown
├─ {DOMAIN} something?
│   └─ → {rite} (/{rite})
```

Add to Rite Selection Tree:
```markdown
├─ {Domain}
│   └─ → {rite} (/{rite})
```

### 6. Update routing/complexity-matrix.md

Add rite's complexity levels:
```markdown
## {rite} Complexity

| Level | When to Use | Scope |
|-------|-------------|-------|
| **{LEVEL1}** | {description} | {scope} |
| **{LEVEL2}** | {description} | {scope} |
```

### 7. Update command-reference.md

Add to Rite Management section:
```markdown
| `/{rite}` | {rite} |
```

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

Location: `.claude/knowledge/consultant/playbooks/curated/{scenario}.md`

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

Add playbook to the list in `.claude/commands/navigation/consult/INDEX.md`:
```markdown
**Curated Playbooks**: Pre-authored sequences for common scenarios
- `{playbook}.md`
```

---

## Verification Commands

After any sync:

```bash
# Verify ecosystem-map has correct rite count
grep "rites" .claude/knowledge/consultant/ecosystem-map.md

# Verify agent-reference has all rites
grep "## .*" .claude/knowledge/consultant/agent-reference.md

# Verify all rite profiles exist
ls .claude/knowledge/consultant/rite-profiles/

# Verify routing includes rite
grep "{rite-name}" .claude/knowledge/consultant/routing/intent-patterns.md

# Verify command reference
grep "/{rite}" .claude/knowledge/consultant/command-reference.md
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

**Symptom**: Consultant says rite has N agents, but it has M
**Fix**: Update agent-reference.md and ecosystem-map.md

---

## Automation Opportunity

Future enhancement: Create a validation script that:
1. Scans knossos for all rites
2. Compares against Consultant knowledge
3. Reports discrepancies
4. Suggests updates

```bash
# Future: $KNOSSOS_HOME/validate-consultant.sh
# Would check all rites are reflected in knowledge base
```

---

## Related

- [INDEX.lego.md](../INDEX.lego.md) - Rite development overview
- [validation/validation.md](../validation/validation.md) - Validation checklist
- [consult](../../../../mena/navigation/consult/INDEX.dro.md) - Consultant reference
