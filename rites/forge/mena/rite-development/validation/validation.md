---
description: "Rite Validation companion for validation skill."
---

# Rite Validation

> Pre-flight checks and troubleshooting for rite deployment.

---

## Pre-Flight Checklist

### Directory Structure
- [ ] Directory exists: `$KNOSSOS_HOME/rites/{rite-name}/`
- [ ] Name follows pattern: `{domain}` (e.g., `sre`, `10x-dev`)
- [ ] Agents directory exists with 3-5 `.md` files
- [ ] `workflow.yaml` exists in rite root

### Workflow Configuration
- [ ] `name` matches directory name
- [ ] `workflow_type` is `sequential`
- [ ] `entry_point.agent` matches first phase agent
- [ ] Each phase has: `name`, `agent`, `produces`, `next`
- [ ] Exactly one phase has `next: null` (terminal)
- [ ] No orphan phases (all reachable from entry)
- [ ] 2-4 complexity levels defined with UPPERCASE names
- [ ] Command mapping comments present

### Agent Files
- [ ] YAML frontmatter between `---` markers
- [ ] `name` matches filename (without .md)
- [ ] `description` includes role summary, triggers, example
- [ ] `model` is valid (`opus`, `sonnet`, `haiku`)
- [ ] `color` is unique within the pantheon
- [ ] All 11 sections present (see agent-template.md)

### Integration
- [ ] Command file exists: `{channel_dir}/commands/{rite-name}.md`
- [ ] Skill directory exists: `rites/{rite-name}/mena/{rite-name}-ref/`
- [ ] `ari sync --rite {rite-name}` completes without error

### Pythia Knowledge Base (REQUIRED)
- [ ] agent-curator completed Pythia catalog update
- [ ] `ari rite --list` shows new rite

### Verification Commands
```bash
# Test rite swap
$KNOSSOS_HOME/ari sync --rite {rite-name}

# Verify workflow and agents copied
cat .knossos/ACTIVE_WORKFLOW.yaml
ls {channel_dir}/agents/

# Verify terminal phase
grep -B1 "next: null" $KNOSSOS_HOME/rites/{rite-name}/workflow.yaml

# Check rite appears in catalog
ari rite list
```

---

## Common Issues

### Rite Swap Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| "Rite not found" | Directory missing or misnamed | Ensure `$KNOSSOS_HOME/rites/{domain}/` exists |
| "0 agents, N phases" | Agent files missing `.md` extension | Rename files to `{name}.md` |
| "workflow.yaml not found" | Missing config | Create from `templates/workflow.yaml.template` |

### Workflow Configuration Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Agent/phase count mismatch | `complexity_levels` counted as phases | Normal - verify with `grep -A1 "^phases:"` |
| Phase references non-existent agent | Filename mismatch | Match `agent:` to actual filename (without `.md`) |
| Phase never executes | Orphan phase | Ensure `next` chain connects all phases |

### Agent File Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Model/color not applied | Malformed frontmatter | Check `---` markers, YAML syntax |
| Duplicate colors | Same color across agents | Assign unique colors per role type |
| Wrong model for role | Mismatched capability | Opus for senior, Sonnet for mid-level, Haiku for assessment |

### Command Mapping Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `/architect` wrong agent | Wrong `produces` value | Use `tdd`, `doc-structure`, or `refactor-plan` |
| `/build` wrong agent | Wrong `produces` value | Use `code`, `commits`, or `documentation` |
| `/qa` wrong agent | Terminal phase wrong | Ensure validation phase has `next: null` |

### Integration Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `/{rite-name}` not recognized | Command file missing | Create dromena source at `mena/rite-switching/{rite-name}.dro.md` and run `ari sync` |
| `@{rite-name}-ref` not found | Skill missing | Create `rites/{rite-name}/mena/{rite-name}-ref/INDEX.lego.md` (use skill-ref template) |
| SESSION_CONTEXT stale | Swap script issue | Verify quick-switch updates active_team |

---

## Quick Diagnostic Commands

```bash
# Full rite validation
$KNOSSOS_HOME/ari sync --rite {rite-name} && \
ls {channel_dir}/agents/ && \
cat .knossos/ACTIVE_WORKFLOW.yaml

# Check YAML syntax
python3 -c "import yaml; yaml.safe_load(open('workflow.yaml'))"

# Verify agent frontmatter
head -20 agents/*.md | grep -A10 "^---"

# Check command file
cat {channel_dir}/commands/{rite-name}.md | head -10
```

---

## Final Sign-Off

- [ ] All checklist items verified
- [ ] Verification commands pass
- [ ] Rite tested with sample task
- [ ] Pythia knowledge base synchronized
