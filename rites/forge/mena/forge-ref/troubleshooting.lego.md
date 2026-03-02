---
name: forge-ref-troubleshooting
description: "Troubleshooting guide for The Forge rite. Use when: ari sync fails, agent validation fails, Consultant can't find a rite, handoffs don't trigger. Triggers: ari sync error, forge troubleshoot, validation fails, handoff broken."
---

# The Forge: Troubleshooting

## "ari sync --rite fails"

Check:
- Rite directory exists at `$KNOSSOS_HOME/rites/{name}/`
- `agents/` subdirectory has .md files
- `workflow.yaml` exists
- File permissions are correct

## "Agent validation fails"

Check:
- All 11 sections present
- Frontmatter has required fields
- No YAML syntax errors
- Token count under budget

## "Consultant can't find rite"

Check:
- `ecosystem-map.md` updated
- `rite-profiles/{rite}.md` exists
- `intent-patterns.md` has keywords
- `command-reference.md` lists command

## "Handoff doesn't trigger"

Check:
- Handoff criteria are specific
- Next agent is correctly named
- `workflow.yaml` `next` field is correct
