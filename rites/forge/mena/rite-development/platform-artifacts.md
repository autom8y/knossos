---
description: "Platform Artifacts companion for rite-development skill."
---

# Platform Artifacts

> Knossos structure reference, ari sync behavior, verification commands, and infrastructure maintenance notes.

## Knossos Structure

```
$KNOSSOS_HOME/
├── rites/
│   └── {rite-name}/              # Created by Platform Engineer
│       ├── agents/
│       │   ├── agent-1.md
│       │   ├── agent-2.md
│       │   ├── agent-3.md
│       │   └── agent-4.md
│       └── workflow.yaml
├── cmd/ari/                      # Rite sync binary
└── internal/materialize/         # Sync logic
```

## Verification Commands

```bash
# Check rite exists
ls -la $KNOSSOS_HOME/rites/{rite-name}/

# Count agents
ls $KNOSSOS_HOME/rites/{rite-name}/agents/*.md | wc -l

# Verify workflow
cat $KNOSSOS_HOME/rites/{rite-name}/workflow.yaml

# Test sync
ari sync --rite {rite-name}

# Verify sync worked
cat .knossos/ACTIVE_RITE
ls {channel_dir}/agents/
```

---

## ari sync --rite Reference

Key behaviors to understand:

### Validation Phase
- Checks rite exists in KNOSSOS_HOME/rites/
- Verifies agents/ directory exists
- Counts .md files (requires >= 1)
- Warns if workflow.yaml missing

### Sync Phase
- Clears the channel agents directory
- Copies new agents from knossos
- Copies workflow.yaml to .knossos/ACTIVE_WORKFLOW.yaml
- Preserves global agents from the user channel agents directory

### State Update
- Writes rite name to .knossos/ACTIVE_RITE
- Updates timestamps

### Exit Codes
- 0: Success
- 1: Invalid arguments
- 2: Validation failure
- 3: Backup failure
- 4: Sync failure

### Idempotency
- Detects if same rite already active
- Skips redundant sync operations

---

## Infrastructure Maintenance Notes

### Adding New Rite
1. Create directory structure
2. Deploy files
3. Test sync
4. No code changes needed

### Modifying ari sync
- Binary location: ari (in PATH or ~/bin/ari)
- Source: $KNOSSOS_HOME/cmd/ari and internal/materialize/
- Test thoroughly before changes
- Global agents preserved via sync logic

### Schema Updates
- Agent frontmatter: name, description, tools, model, color
- Workflow: name, workflow_type, entry_point, phases, complexity_levels
- All fields have validation in materialize pipeline
