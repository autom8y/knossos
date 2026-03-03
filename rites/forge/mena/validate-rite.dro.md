---
description: Run validation suite on an existing rite
argument-hint: "<rite-name> [--verbose]"
allowed-tools: Bash, Glob, Grep, Read, Task, TodoWrite
model: opus
---

## Context

Auto-injected by SessionStart hook (project, rite, session context).

## Your Task

Validate the rite: $ARGUMENTS

## Behavior

### 1. Parse Arguments

- `rite-name`: Required. Name of rite to validate
- `--verbose`: Optional. Show detailed check results

### 2. Locate Rite

Check rite exists at:
- `$KNOSSOS_HOME/rites/{rite-name}/`

If not found, report error and suggest checking rite name.

### 3. Invoke Eval Specialist

Use the Task tool to invoke the eval-specialist agent:

```
"Validate the rite at $KNOSSOS_HOME/rites/{rite-name}/.

Run the full validation suite:
1. Structure Validation - Check all files exist
2. Schema Validation - Verify frontmatter and workflow.yaml
3. Logic Validation - Check phase chain and complexity levels
4. Adversarial Testing - Run edge case prompts

Produce an eval-report.md with pass/fail status for each check.

Verbose mode: {verbose}"
```

### 4. Report Results

Display validation results:

```
VALIDATION REPORT: {rite-name}
==============================
Status: {PASS | FAIL | WARNINGS}

Structure:  {✓|✗} All files exist
Schema:     {✓|✗} Frontmatter valid
Logic:      {✓|✗} Workflow coherent
Adversarial: {✓|✗} Edge cases handled

{If verbose, show detailed check results}

{If issues found, list them with severity}
```

## Example Usage

```bash
# Validate security
/validate-rite security

# Validate with detailed output
/validate-rite 10x-dev --verbose

# Validate hygiene rite
/validate-rite hygiene
```

## Reference

Full documentation: `.claude/skills/forge-ref/INDEX.md`
Validation checklist: see `rite-development` skill validation section
