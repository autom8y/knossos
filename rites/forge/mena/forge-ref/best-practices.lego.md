---
name: forge-ref-best-practices
description: "Best practices for rite design, agent prompts, and workflows in The Forge. Use when: designing a new rite, writing agent prompts, configuring workflow.yaml. Triggers: rite design, agent prompt, workflow design, best practice."
---

# The Forge: Best Practices

## Rite Design

1. **Start with purpose**: What problem does this rite solve?
2. **3-5 agents**: Less is more. Consolidate related responsibilities.
3. **Clear boundaries**: No overlap between agent domains.
4. **Linear flow**: Avoid circular dependencies.

## Agent Prompts

1. **Strong identity**: First paragraph establishes who they are.
2. **Actionable instructions**: Not vague guidelines.
3. **Realistic examples**: Show actual expected behavior.
4. **Token awareness**: Keep under 4000 tokens.

## Workflows

1. **One terminal**: Exactly one phase with `next: null`.
2. **Reachable phases**: All phases accessible from entry.
3. **Sensible gating**: Lower complexity = fewer phases.
4. **Clear commands**: Map /architect, /build, /qa appropriately.

## Extending The Forge

### Adding New Patterns

1. Create file in `patterns/` directory within this skill
2. Reference in relevant agent prompts
3. Update forge-ref INDEX companion reference

### Adding New Eval Checks

1. Create file in `evals/` directory within this skill
2. Update `eval-specialist.md` to reference it
3. Add to validation checklist

### Modifying Forge Workflow

1. Edit `$KNOSSOS_HOME/rites/forge/workflow.yaml`
2. Update agent handoff criteria if phases change
3. Update forge-ref INDEX companion reference
