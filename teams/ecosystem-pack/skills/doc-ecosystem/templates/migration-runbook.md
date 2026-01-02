# Migration Runbook Template

> Satellite owner migration guide.

```markdown
# Migration Runbook: [Change Title]

## Overview
**Affects**: [All satellites | Satellites with X configuration]
**Breaking**: [YES | NO]
**CEM Version**: [version]
**skeleton Version**: [version]

[2-3 sentences describing what changed and why satellite owners care]

## Before You Begin
- [ ] Backup `.claude/` directory
- [ ] Verify current CEM version: `cem --version`
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

### 1. Update CEM
```bash
# Commands to upgrade CEM
cd /path/to/cem
git pull origin main
./install.sh
cem --version  # Should show vX.Y.Z
```

### 2. Update skeleton
```bash
cd /path/to/skeleton
cem sync  # Pulls latest skeleton changes
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
cem sync  # Should complete without errors
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
cem sync --force  # Reset to pre-migration state
```

## Troubleshooting

### Issue: [Common problem]
**Symptom**: [error message or behavior]
**Solution**: [fix]

## Compatibility
| CEM | skeleton | Status |
|-----|----------|--------|
| 2.0 | 2.0 | Fully supported |
| 2.0 | 1.9 | Backward compatible |
| 1.9 | 2.0 | Unsupported--upgrade CEM first |

## Support
Questions? Issues? [contact info or issue tracker]
```

## Quality Gate

**Migration Runbook complete when:**
- Step-by-step instructions executable
- Rollback procedure tested
- Compatibility matrix complete
- Troubleshooting covers common issues
