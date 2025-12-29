# Playbook: Ecosystem Migration

> Safe migration of CEM/skeleton infrastructure across satellites

## When to Use

- Upgrading satellites to new skeleton version
- Migrating hook or skill patterns across projects
- Breaking schema changes in settings
- Agent team restructuring affecting multiple projects
- Cross-satellite compatibility testing

## Prerequisites

- Skeleton changes committed and tested locally
- List of affected satellites identified
- Rollback plan documented

## Command Sequence

### Phase 1: Initialize

```bash
/ecosystem
```
**Expected output**: Team switched to ecosystem-pack, roster displayed
**Decision point**: If single-satellite fix, may not need full workflow.

### Phase 2: Start Session

```bash
/start "Migration scope" --complexity=MIGRATION
```
**Expected output**: Session created, context established
**Decision point**: Adjust complexity level:
- PATCH: Single file or setting
- MODULE: Hook, skill, or agent changes
- SYSTEM: Cross-cutting infrastructure
- MIGRATION: Multi-satellite rollout

### Phase 3: Impact Analysis

Ecosystem Analyst assesses migration scope.

**Expected output**: gap-analysis artifact identifying affected satellites and changes
**Decision point**: Are there breaking changes requiring deprecation paths?

### Phase 4: Compatibility Testing

```bash
/handoff compatibility-tester
```
**Expected output**: compatibility-report with satellite test results
**Decision point**: All satellites must pass before proceeding.

### Phase 5: Migration Implementation

```bash
/handoff integration-engineer
```
**Expected output**: Migration scripts, CEM changes, updated manifests
**Decision point**: Test on staging satellite first.

### Phase 6: Documentation

```bash
/handoff documentation-engineer
```
**Expected output**: Migration runbook, changelog, breaking change notices
**Decision point**: For major migrations, create upgrade guide.

### Phase 7: Rollout

Execute migration across satellites:
```bash
# Per satellite:
cd ~/Code/{satellite}
cem sync
```
**Decision point**: Monitor for sync errors, validate each satellite.

### Phase 8: Finalize

```bash
/wrap
```
**Expected output**: Session summary, migration report

## Variations

- **Non-breaking changes**: Skip compatibility testing, direct sync
- **Single satellite**: Skip multi-satellite rollout tracking
- **Emergency fix**: Use `/hotfix` instead

## Success Criteria

- [ ] All affected satellites identified
- [ ] Compatibility verified before rollout
- [ ] Migration scripts tested
- [ ] Documentation updated
- [ ] All satellites successfully synced
- [ ] No rollback required

## Rollback

If migration fails on a satellite:
```bash
# In affected satellite:
cd ~/Code/{satellite}
git checkout .claude/          # Reset to pre-sync state
cem status                     # Verify clean state
# Document failure, fix in skeleton, re-attempt
```
