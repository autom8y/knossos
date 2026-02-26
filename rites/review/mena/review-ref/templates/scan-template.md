# Scan Findings: {project-name}

## Scope
- **Target**: {what was scanned -- full repo, specific module, file list}
- **Complexity**: {QUICK|FULL}
- **Date**: {scan date}

## Overview
- **Files**: {count} across {n} directories
- **Languages**: {detected types by file extension}
- **Tests**: {test directory exists? ratio of test files to source files}
- **Dependencies**: {package manager(s) detected, entry count}
- **Entry points**: {identified main/index/app files}

## Raw Signals

<!-- Repeat this block for each signal found. Group by category. -->

### [{CATEGORY}] {Signal title}
- **Location**: {path/to/file:line or directory path}
- **Signal**: {What triggered this finding -- the heuristic that fired}
- **Evidence**: {Quantitative data -- file size, count, ratio, specific content}
- **Confidence**: {HIGH | MEDIUM | LOW}

<!-- Categories: Complexity, Testing, Dependencies, Structure, Hygiene -->

## Noise Log

<!-- Optional: signals considered and dismissed. Helps pattern-profiler understand what was ruled out. -->

| Signal | Location | Reason Dismissed |
|--------|----------|-----------------|
| {signal} | {location} | {why this is noise, not signal} |

## Metrics Summary

| Metric | Value |
|--------|-------|
| Total files scanned | {n} |
| Directories traversed | {n} |
| Signals identified | {n} |
| By category | Complexity: {n}, Testing: {n}, Dependencies: {n}, Structure: {n}, Hygiene: {n} |
| By confidence | HIGH: {n}, MEDIUM: {n}, LOW: {n} |

## Scan Coverage

<!-- Document what was and was not scanned. Helps pattern-profiler identify coverage gaps. -->

| Area | Status | Notes |
|------|--------|-------|
| Project root | {scanned/skipped} | {notes} |
| Source directories | {scanned/skipped} | {notes} |
| Test directories | {scanned/skipped} | {notes} |
| Config/build files | {scanned/skipped} | {notes} |
| Documentation | {scanned/skipped} | {notes} |

---
*Produced by signal-sifter | Review rite | {QUICK|FULL} mode*
