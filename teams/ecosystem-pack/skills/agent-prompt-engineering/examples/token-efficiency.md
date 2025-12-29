# QA Adversary Token Efficiency

> Part of [agent-prompt-engineering](../SKILL.md) skill examples

**Problem**: File verification protocol repeated identically. 25 lines duplicated across 5 agents.

## Before (Score: 3.0/5 on Token Efficiency)

```markdown
## File Operations

When writing files, always verify the write succeeded:

1. Use Write tool to create the file
2. Use Read tool to verify contents
3. Compare actual vs expected content
4. If mismatch, retry write operation
5. After 3 failed attempts, report error to user

Important considerations:
- Always use absolute paths
- Check parent directory exists before writing
- Handle permission errors gracefully
- Log all file operations for debugging

[... continues for 25 lines ...]

## File Verification Protocol

For each artifact produced:
1. Write the file
2. Read it back
3. Verify content matches
4. Report verification status

[... 12 more lines ...]
```

**Issues annotated**:
- 25+ lines of file handling repeated in 5 agents
- Total waste: 125 lines across team
- Same content, same wording, same order
- Updates require 5 file edits

## After (Score: 5.0/5 on Token Efficiency)

```markdown
## File Operations

See `file-verification` skill for artifact verification protocol.
```

## Key Improvements

- 1 line replaces 25 lines
- Single source of truth for updates
- Agents reference shared skill
- Consistent verification across team

**Token comparison per agent**: 25 lines -> 1 line
**Token comparison across 5 agents**: 125 lines -> 5 lines (-96%)
