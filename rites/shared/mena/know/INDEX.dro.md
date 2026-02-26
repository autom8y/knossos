---
name: know
description: "Generate persistent codebase knowledge via theoros observation. Produces .know/{domain}.md with schema-validated frontmatter and structured knowledge sections."
argument-hint: "[domain] [--force] [--expires=DURATION]"
allowed-tools: Bash, Read, Write, Glob, Grep, Task, Skill
model: opus
context: fork
---

# /know -- Codebase Knowledge Generator

Dispatches theoros to observe and document codebase architecture, producing persistent `.know/` reference files that any CC agent can consume for codebase context.

## Context

This command operates in forked context (transient session). It generates `.know/` files at the project root.

## Pre-flight

1. **Parse arguments**:
   - `domain`: Optional. Default: `"architecture"`. Must be a registered pinakes domain with `codebase` scope.
   - `--force`: Skip expiry check, regenerate even if current.
   - `--expires=DURATION`: Override default expiry (e.g., `--expires=14d`). Default: `"7d"`.

2. **Validate domain exists**:
   - Load pinakes skill: `Skill("pinakes")`
   - Read pinakes INDEX to find the Domain Registry table
   - Verify requested domain appears in the table with scope `codebase`
   - If not found: ERROR "Domain '{domain}' not registered in pinakes or not codebase-scoped. Available codebase domains: {list}"

3. **Check expiry** (skip if `--force`):
   - If `.know/{domain}.md` exists, read its YAML frontmatter
   - Parse `generated_at` and `expires_after`
   - If `generated_at + expires_after > now`: report "Knowledge for '{domain}' is current (generated {date}, expires {date}). Use --force to regenerate." and STOP.
   - If expired or file does not exist: proceed to generation.

4. **Compute source_hash**:
   - Run: `git rev-parse --short HEAD`
   - Store result for frontmatter injection.

5. **Ensure .know/ directory exists**:
   - Run: `mkdir -p .know`

## Phase 1: Criteria Loading

1. Read the domain criteria file from the pinakes source:
   ```
   Read("rites/shared/mena/pinakes/domains/{domain}.lego.md")
   ```
   If that path doesn't exist, fall back to the materialized skill:
   ```
   Read(".claude/skills/pinakes/domains/{domain}.md")
   ```

2. Extract the full criteria content for injection into the theoros dispatch prompt.

## Phase 2: Theoros Dispatch

Dispatch theoros with reframed instructions. The dispatch prompt transforms theoros from auditor to knowledge producer WITHOUT modifying the theoros agent prompt.

```
Task(subagent_type="theoros", prompt="
## Knowledge Observation: {domain}

You are producing a KNOWLEDGE REFERENCE DOCUMENT, not an audit report.

Your domain criteria define WHAT to observe and document in the codebase. Instead of grading compliance against standards, assess the COMPLETENESS of your knowledge capture. The grading rubric measures how thoroughly you documented what exists, not whether what exists meets a standard.

### Reframing Your Audit Protocol

Your audit protocol still applies, but reframed:
- 'What to evaluate' -> 'What to observe and document'
- 'Evidence collection' -> 'Where to look and what to record'
- 'Assign letter grades' -> 'Assess completeness of knowledge capture'
- Grade A = comprehensive knowledge capture with evidence
- Grade F = incomplete or inaccurate knowledge capture

### Domain Criteria

{full_criteria_file_content}

### Output Format

Produce your output in TWO parts:

**Part 1: Knowledge Reference Body**

Structured markdown with one section per criterion. Each section documents what you observed in the codebase. Use specific file paths, type names, and package references. This becomes the body of the .know/ file.

Start with a top-level heading derived from the domain name:
# Codebase {Domain Title}

Then one H2 section per criterion from the criteria file, named EXACTLY as the criterion title appears (e.g., if the criterion is "Criterion 1: Package Structure", the section is "## Package Structure"). Follow the criteria order.

End with:
## Knowledge Gaps
(List anything you could not fully document and why.)

Be thorough. Include file paths, type names, function signatures, and import relationships. This document will be the primary reference for CC agents working in this codebase.

**Part 2: Assessment Metadata**

After the knowledge body, on its own line produce a fenced metadata block:

```metadata
overall_grade: {A-F}
overall_percentage: {N.N}%
confidence: {0.0-1.0}
criteria_grades:
  {criterion_1_snake_case}: {grade}
  {criterion_2_snake_case}: {grade}
  ...one entry per criterion from the criteria file...
```

The criteria_grades keys should be snake_case versions of the criterion names from the criteria file (e.g., "Package Structure" -> "package_structure", "Error Handling Style" -> "error_handling_style").

The confidence value is your self-assessment of how completely you covered the source scope. 1.0 = every file in scope examined. 0.5 = significant areas unexplored due to scope or turn limits.

### Scope Reminder

Read the **Scope** section in the domain criteria file above. It defines the target files and observation focus. Stay within the defined scope.
")
```

## Phase 3: Output Assembly

After theoros returns:

1. **Parse theoros output**:
   - Extract the knowledge reference body (everything before the ` ```metadata` fence)
   - Extract the metadata block content (between the fences)
   - Parse confidence and grades from the metadata
   - If metadata block is missing or unparseable: use defaults (confidence: 0.5, grade: C)

2. **Construct the .know/ file**:
   Build YAML frontmatter:
   ```yaml
   ---
   domain: {domain}
   generated_at: "{current ISO 8601 UTC timestamp}"
   expires_after: "{expires_duration, default 7d}"
   source_scope:
     - "./cmd/**/*.go"
     - "./internal/**/*.go"
     - "./go.mod"
   generator: theoros
   source_hash: "{git short SHA from pre-flight}"
   confidence: {confidence from theoros metadata}
   format_version: "1.0"
   ---
   ```
   Combine frontmatter + theoros knowledge body (Part 1 only, not the metadata fence).

3. **Write the file**:
   ```
   Write(".know/{domain}.md", frontmatter + body)
   ```

4. **Verify the file**:
   ```
   Read(".know/{domain}.md", limit=20)
   ```
   Confirm frontmatter fields are present and valid. Confirm body sections exist.

## Phase 4: Report

Display summary to the user:

```
## Knowledge Generated: .know/{domain}.md

| Field | Value |
|-------|-------|
| Domain | {domain} |
| Source hash | {source_hash} |
| Confidence | {confidence} |
| Completeness | {grade} ({percentage}%) |
| Expires | {expiry_date} |

### Criteria Completeness
| Criterion | Grade |
|-----------|-------|
| {criterion_1_name} | {grade} |
| {criterion_2_name} | {grade} |
| ... | ... |

This file is consumable by any CC agent via `Read(".know/{domain}.md")`.
Regenerate with: `/know {domain} --force`
```

## Error Handling

| Scenario | Action |
|----------|--------|
| Domain not in pinakes | ERROR with available codebase domains |
| Theoros dispatch fails | ERROR "Knowledge generation failed: {reason}" |
| Theoros output unparseable | Write body as-is with default metadata, WARN about metadata parsing failure |
| .know/ directory not writable | ERROR with permission details |
| Git not available (no source_hash) | Use "unknown" as source_hash, WARN |

## Anti-Patterns

- **Modifying theoros agent prompt**: The reframing happens in THIS dispatch prompt, not in the agent definition. Theoros remains a general-purpose domain auditor.
- **Skipping verification**: Always Read the .know/ file after writing to confirm correctness.
- **Ignoring expiry**: Respect the expiry mechanism. Regenerating current knowledge wastes theoros capacity.
- **Hardcoding architecture-only**: This dromenon accepts any codebase-scoped domain. Architecture is the default, but the mechanism is generic.
