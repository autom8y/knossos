# Migration Runbook Template

> Satellite owner migration guide.

```markdown
# Migration Runbook: [Change Title]

## Overview
**Affects**: [All satellites | Satellites with X configuration]
**Breaking**: [YES | NO]
**Sync Pipeline Version**: [version]
**Knossos Version**: [version]

[2-3 sentences describing what changed and why satellite owners care]

## Before You Begin
- [ ] Backup `.claude/` directory
- [ ] Verify current ari version: `ari --version`
- [ ] Read this runbook completely before starting

## Current Behavior
[Description with config example]

```json
// Old configuration
{
  "setting": "old-style-value"
}
```

## New Behavior
[Description with config example]

```json
// New configuration
{
  "setting": {
    "new": "nested-style-value"
  }
}
```

## Migration Steps

### 1. Update Knossos
```bash
# Commands to upgrade knossos
cd /path/to/knossos
git pull origin main
./install.sh
ari --version  # Should show vX.Y.Z
```

### 2. Sync knossos changes
```bash
cd /path/to/satellite
ari sync  # Pulls latest knossos changes
```

### 3. Migrate Settings
```bash
# Specific migration commands
mv .claude/settings.json .claude/settings.json.backup
# Apply transformation...
```

### 4. Verify Migration
```bash
# Verification commands
ari sync  # Should complete without errors
# Check specific functionality...
```

**Expected output**: [what success looks like]

### 5. Test Satellite Functionality
- [ ] Hooks fire correctly: `# test command`
- [ ] Skills load: `# test command`
- [ ] Agents register: `# test command`

## Rollback Procedure
If migration fails:

```bash
# Restore backup
cp .claude/settings.json.backup .claude/settings.json
ari sync --force  # Reset to pre-migration state
```

## Troubleshooting

### Issue: [Common problem]
**Symptom**: [error message or behavior]
**Solution**: [fix]

## Compatibility
| Ari | Knossos | Status |
|-----|--------|--------|
| 2.0 | 2.0 | Fully supported |
| 2.0 | 1.9 | Backward compatible |
| 1.9 | 2.0 | Unsupported--upgrade ari first |

## Support
Questions? Issues? [contact info or issue tracker]
```

## Quality Gate

**Migration Runbook complete when:**
- Step-by-step instructions executable
- Rollback procedure tested
- Compatibility matrix complete
- Troubleshooting covers common issues
