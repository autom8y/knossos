# CLI Reference - In Progress

**Status**: Under development
**Target**: Comprehensive reference for all 68 `ari` commands across 15 families

## Current Coverage Gap

- **Implemented**: 68 commands
- **Documented**: ~8 commands (mentioned in guides)
- **Gap**: 60 commands (88%)

## Command Families (Pending Documentation)

| Family | Commands | Priority |
|--------|----------|----------|
| session | 11 | HIGH |
| rite | 10 | HIGH |
| worktree | 11 | HIGH |
| sync | 7 | HIGH |
| hook | 6 | MEDIUM |
| handoff | 4 | MEDIUM |
| inscription | 5 | MEDIUM |
| manifest | 4 | MEDIUM |
| artifact | 3 | MEDIUM |
| validate | 3 | MEDIUM |
| sails | 1 | LOW |
| naxos | 1 | LOW |
| tribute | 1 | LOW |

## Temporary Workaround

Use CLI help directly:
```bash
ari --help              # List all commands
ari [command] --help    # Command-specific help
```

## See Also

- [Refinement Recommendations](../../compliance/audits/refinement-recommendations-20260108.md)
- [Ariadne CLI Guide](../guides/ariadne-cli.md)
