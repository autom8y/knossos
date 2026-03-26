---
name: rite-development-quick-start
description: "Step-by-step bash commands for creating a new rite from scratch. Use when: ready to scaffold a new rite and need the exact command sequence. Triggers: create rite, scaffold rite, new rite commands."
---

# Rite Creation: Quick Start Commands

## Command Sequence

```bash
# 1. Create directory structure
mkdir -p $KNOSSOS_HOME/rites/{name}/agents

# 2. Copy and fill templates
# - workflow.yaml from templates/workflow.yaml.template
# - agent files from templates/agent-template.md

# 3. Create command and skill
# - {channel_dir}/commands/{name}.md
# - rites/{name}/mena/{name}-ref/INDEX.lego.md

# 4. Sync to project
ari sync --rite {name}

# 5. Validate
$KNOSSOS_HOME/ari sync --rite {name}
```

See [validation/validation.md](validation/validation.md) for full pre-flight checks.

## Notes

- `$KNOSSOS_HOME` is the knossos repo root
- Step 4 and 5 are the same command — run once; the output shows validation status
- After sync, verify with `ari rite --list` that the new rite appears
