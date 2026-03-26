---
description: "Scan Heuristics Catalog companion for review-ref skill."
---

# Scan Heuristics Catalog

Language-agnostic structural signals for the signal-sifter. All thresholds are starting points -- context may justify exceptions.

## Complexity Signals

| Signal | Threshold | Evidence | Confidence |
|--------|-----------|----------|------------|
| File size (lines) | >500 | `wc -l` | HIGH |
| File size (extreme) | >1000 | `wc -l` | HIGH -- almost always a problem |
| Nesting depth | >4 indent levels | Visual inspection or indent analysis | MEDIUM |
| Function/method count per file | >20 | Grep for function definitions | MEDIUM |
| Deeply nested directories | >4 levels deep | Directory tree | MEDIUM |
| Files per directory | >20 | `ls \| wc -l` | LOW -- may be intentional in flat structures |
| Long parameter lists | >5 params per function | Grep for function signatures | LOW |

### False Positives (Complexity)
- Auto-generated files (migrations, protobuf, swagger) -- check for generation markers
- Configuration files (large YAML/JSON) -- structural, not behavioral complexity
- Test fixtures and data files -- size is not a quality signal here
- Vendored/bundled dependencies -- excluded from review scope

## Testing Signals

| Signal | Threshold | Evidence | Confidence |
|--------|-----------|----------|------------|
| No test directory | absent | `find . -type d -name '*test*'` | HIGH |
| Test-to-source ratio | <0.3 | File count comparison | HIGH |
| No test framework config | absent | Look for jest.config, pytest.ini, etc. | MEDIUM |
| Test files without assertions | any | Grep for assert/expect patterns | MEDIUM |
| Core modules untested | any critical path | Cross-reference entry points with test coverage | HIGH |
| Test file naming mismatch | inconsistent | Compare test file naming conventions | LOW |

### False Positives (Testing)
- Projects with integration-only testing strategy (no unit tests by design)
- Libraries with example-based documentation instead of test suites
- Infrastructure-as-code repositories where tests are in CI pipelines
- Very small projects (<10 files) where test ratio is meaningless

## Dependencies Signals

| Signal | Threshold | Evidence | Confidence |
|--------|-----------|----------|------------|
| Dependency count | >100 entries | Package manifest line count | MEDIUM |
| No lockfile | absent | Look for lock files (package-lock, Gemfile.lock, etc.) | HIGH |
| Multiple package managers | >1 | Multiple manifest types in same project | HIGH |
| Unpinned versions | any `*` or `latest` | Grep version specifiers | MEDIUM |
| Duplicate dependency intent | overlapping libs | Manual cross-reference (e.g., lodash + underscore) | LOW |
| Stale lockfile | lockfile older than manifest | File modification dates | MEDIUM |

### False Positives (Dependencies)
- Monorepos with workspace-level dependency management
- Projects using dependency injection frameworks (appear to have many deps)
- Development-only dependencies counted alongside production deps

## Structure Signals

| Signal | Threshold | Evidence | Confidence |
|--------|-----------|----------|------------|
| Flat project root | >15 files in root | `ls` root directory | MEDIUM |
| No clear entry point | absent | Look for main, index, app, entry files | MEDIUM |
| Mixed concerns in directory | business + infra in same dir | Directory content inspection | MEDIUM |
| Circular directory references | symlinks or cross-imports | Directory tree + import analysis | HIGH |
| No separation of concerns | single directory for all code | Project structure | HIGH |
| Inconsistent directory naming | mixed conventions | Directory listing | LOW |
| README absent at root | absent | File check | MEDIUM |

### False Positives (Structure)
- Single-purpose microservices intentionally have flat structure
- Scripting projects without complex directory hierarchies
- Projects following framework conventions that differ from general expectations

## Hygiene Signals

| Signal | Threshold | Evidence | Confidence |
|--------|-----------|----------|------------|
| TODO/FIXME density | >10 per file or >50 total | `grep -r 'TODO\|FIXME'` | MEDIUM |
| Commented-out code blocks | >10 contiguous lines | Visual pattern matching | MEDIUM |
| Mixed naming conventions | camelCase + snake_case | Identifier pattern analysis | HIGH |
| Dead code indicators | unused exports, unreachable branches | Grep for patterns | LOW |
| Inconsistent file naming | mixed conventions in same directory | `ls` comparison | MEDIUM |
| Debug/console statements | in non-debug code | Grep for print/console/log | LOW |
| Empty catch/except blocks | any | Grep for catch patterns | HIGH |

### False Positives (Hygiene)
- TODO comments with issue tracker references (intentional tracking, not neglect)
- Commented-out code with explicit "keep for reference" notes
- Mixed naming due to language interop (e.g., Python calling C libraries)
- Debug statements in development-only files or debug modules

## Cross-Language Patterns

These patterns apply regardless of programming language:

1. **Entry point clarity**: Every project should have an obvious entry point. If you cannot find it within 30 seconds, that is a Structure signal.

2. **Configuration sprawl**: Multiple config files for the same concern (environment, build, deploy) scattered across directories. Look for `.env`, config/, settings files.

3. **Documentation presence**: README at root, contributing guide for OSS, API docs for libraries. Absence is a signal; quality is not assessed (docs rite scope).

4. **Build system clarity**: Can you determine how to build/run the project within 60 seconds? If not, Structure signal.

5. **Separation of generated vs authored**: Generated code (migrations, bindings, compiled assets) should be clearly separated from authored code.

## Applying Heuristics

### Scan Order
1. Project root structure (entry points, config, README)
2. Directory tree (depth, organization, naming)
3. File-level metrics (sizes, counts per directory)
4. Content signals (TODOs, naming, dead indicators)
5. Dependency analysis (manifests, lockfiles)
6. Testing structure (test directories, ratios, framework config)

### Confidence Guidelines
- **HIGH**: Strong structural signal with clear threshold violation. Report with certainty.
- **MEDIUM**: Contextual signal that may have valid explanations. Report with the evidence.
- **LOW**: Weak signal that might be noise. Include only if multiple LOW signals cluster in the same category.

### When to Skip a Signal
- Explicitly excluded from scope by user request
- File is in a vendored, generated, or third-party directory
- Signal contradicts a higher-confidence signal in the same category
- Project type makes the signal irrelevant (e.g., test ratio for a config-only repo)
