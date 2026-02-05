# Team Validation

> Pre-flight checks and troubleshooting for rite deployment.

---

## Pre-Flight Checklist

### Directory Structure
- [ ] Directory exists: `$ROSTER_HOME/rites/{rite-name}/`
- [ ] Name follows pattern: `{domain}-pack` (e.g., `sre`)
- [ ] Agents directory exists with 3-5 `.md` files
- [ ] `workflow.yaml` exists in team root

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
- [ ] `color` is unique within team
- [ ] All 11 sections present (see agent-template.md)

### Integration
- [ ] Command file exists: `.claude/commands/{rite-name}.md`
- [ ] Skill directory exists: `.claude/skills/{rite-name}-ref/`
- [ ] COMMAND_REGISTRY.md updated

### Consultant Knowledge Base (REQUIRED)
- [ ] `ecosystem-map.md` updated with new team
- [ ] `agent-reference.md` updated with new agents
- [ ] `rite-profiles/{rite-name}.md` created
- [ ] Routing files updated (`intent-patterns.md`, `decision-trees.md`)
- [ ] `command-reference.md` updated

### Verification Commands
```bash
# Test team swap
$ROSTER_HOME/swap-rite.sh {rite-name}

# Verify workflow and agents copied
cat .claude/ACTIVE_WORKFLOW.yaml
ls .claude/agents/

# Verify terminal phase
grep -B1 "next: null" $ROSTER_HOME/rites/{rite-name}/workflow.yaml

# Check Consultant knowledge
grep "{rite-name}" .claude/knowledge/consultant/ecosystem-map.md
```

---

## Common Issues

### Team Swap Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| "Team not found" | Directory missing or misnamed | Ensure `$ROSTER_HOME/rites/{domain}-pack/` exists |
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
| `/{rite-name}` not recognized | Command file missing | Create `.claude/commands/{rite-name}.md` |
| `@{rite-name}-ref` not found | Skill missing | Create `.claude/skills/{rite-name}-ref/skill.md` |
| SESSION_CONTEXT stale | Swap script issue | Verify quick-switch updates active_team |

---

## Quick Diagnostic Commands

```bash
# Full team validation
$ROSTER_HOME/swap-rite.sh {rite-name} && \
ls .claude/agents/ && \
cat .claude/ACTIVE_WORKFLOW.yaml

# Check YAML syntax
python3 -c "import yaml; yaml.safe_load(open('workflow.yaml'))"

# Verify agent frontmatter
head -20 agents/*.md | grep -A10 "^---"

# Check command file
cat .claude/commands/{rite-name}.md | head -10
```

---

## Final Sign-Off

- [ ] All checklist items verified
- [ ] Verification commands pass
- [ ] Team tested with sample task
- [ ] Consultant knowledge base synchronized
