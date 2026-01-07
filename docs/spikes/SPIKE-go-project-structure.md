# SPIKE: Go Project Structure Best Practices (2025+)

**Date**: 2026-01-07
**Initiative**: knossos-finalization
**Phase**: R&D (research only)
**Session**: session-20260107-164631-8dd6f03a
**Upstream**: SPIKE-knossos-consolidation-architecture.md (Section 8.6)
**Deliverable**: ADR-go-module-structure.md (draft)

---

## Executive Summary

This spike researches Go project layout best practices for 2025+ to inform the Knossos consolidation. The research confirms that the `cmd/` + `internal/` pattern is the idiomatic standard for Go CLI projects. The current `ariadne/` subdirectory structure should be promoted to repository root, and the module path changed from `github.com/autom8y/knossos` to `github.com/autom8y/knossos`.

**Key Finding**: There is no single "official" Go project layout standard, but strong community consensus has emerged around the `cmd/` + `internal/` pattern. The golang-standards/project-layout repository, while explicitly NOT an official standard, documents widely-adopted conventions.

**Recommendation**: Adopt standard Go project layout with `cmd/ari/` + `internal/` at repository root.

---

## 1. Research Areas Covered

| Area | Status | Key Finding |
|------|--------|-------------|
| Standard Go project layouts | Complete | `cmd/` + `internal/` is consensus |
| Alternatives to standard layout | Complete | Flat layout only for small projects |
| Import path migration tooling | Complete | Multiple tools; sed-based is reliable |
| Monorepo vs single-module | Complete | Single-module recommended for knossos |
| CLI tool project structure | Complete | Cobra convention aligns with standard |

---

## 2. Go Project Layout Patterns

### 2.1 The Standard Layout (`cmd/` + `internal/`)

The [golang-standards/project-layout](https://github.com/golang-standards/project-layout) repository documents the most widely-adopted conventions. Key directories:

| Directory | Purpose | External Import |
|-----------|---------|-----------------|
| `cmd/` | Application entry points | N/A (main packages) |
| `internal/` | Private application code | **Blocked by compiler** |
| `pkg/` | Public library code | Allowed |

**Critical Disclaimer**: This is NOT an official standard defined by the core Go team. It represents community conventions that have emerged through practical experience.

### 2.2 Official Go Team Guidance

The [official Go documentation](https://go.dev/doc/modules/layout) provides this guidance:

```
project-root-directory/
  go.mod
  cmd/
    prog1/
      main.go
    prog2/
      main.go
  internal/
    auth/
      auth.go
    hash/
      hash.go
```

Key recommendations from official docs:
- Use `internal/` to prevent external modules from importing private code
- Use `cmd/` for multiple CLI programs with separate entry points
- Go enforces `internal/` import restrictions at the compiler level

### 2.3 When to Use Each Pattern

| Scenario | Recommended Structure |
|----------|----------------------|
| Simple single-file tool | `main.go` + `go.mod` only |
| Small library | Root package + `go.mod` |
| CLI tool (our case) | `cmd/toolname/` + `internal/` |
| Multiple commands | `cmd/tool1/`, `cmd/tool2/` + `internal/` |
| Server with assets | `cmd/` + `internal/` + asset directories |

### 2.4 Community Best Practices (2025)

Based on [recent analysis](https://www.glukhov.org/post/2025/12/go-project-structure/):

**Do:**
- Start simple; let structure evolve organically
- Favor shallow hierarchies (1-2 levels deep)
- Organize `internal/` by feature/domain, not technical layer
- Place test files alongside implementation (`_test.go` suffix)

**Avoid:**
- Generic package names (`utils`, `helpers`, `common`)
- Over-nesting directories (`internal/services/user/handlers/http/v1/`)
- `src/` directory (Java pattern, not Go idiomatic)
- Circular dependencies
- Mixing business logic with HTTP handlers

### 2.5 Structure Assessment: Current vs Target

**Current Structure** (`ariadne/` subdirectory):
```
roster/
  ariadne/
    go.mod                    # module github.com/autom8y/knossos
    cmd/
      ari/
        main.go
    internal/
      cmd/                    # Command implementations
      session/
      rite/
      sync/
      ...
```

**Target Structure** (root-level Go):
```
knossos/                       # (renamed from roster)
  go.mod                       # module github.com/autom8y/knossos
  cmd/
    ari/
      main.go
  internal/
    cmd/                       # Command implementations
    session/
    rite/
    sync/
    ...
  rites/                       # Non-Go content (rite definitions)
  templates/                   # Non-Go content (hook templates)
  docs/
```

**Assessment**: The target structure aligns perfectly with Go community standards:
- `cmd/ari/` for CLI entry point
- `internal/` for all private code (compiler-enforced)
- No `pkg/` needed (not building public library)
- Non-Go assets at root level

---

## 3. Import Path Migration Tooling

### 3.1 Migration Scope

Current state:
- **Module path**: `github.com/autom8y/knossos`
- **Target path**: `github.com/autom8y/knossos`
- **Files affected**: 150+ Go files
- **Occurrences**: 450+ import statements

### 3.2 Available Tools

| Tool | Approach | Status | Notes |
|------|----------|--------|-------|
| `sed` + `find` | Text-based | **Recommended** | Simple, reliable, well-documented |
| [gofiximports](https://github.com/semk/gofiximports) | Go-aware | Active | Recursive, gofmt-compatible |
| [gorep](https://github.com/novalagung/gorep) | Go-aware | Active | Simple CLI |
| [govers](https://github.com/rogpeppe/govers) | Go-aware | Legacy | Pre-modules, may not work correctly |
| `gopls rename` | LSP-based | Modern | Works for symbols, not module paths |
| `gofmt -r` | AST-based | **Not suitable** | Designed for code transforms, not imports |

### 3.3 Recommended Migration Approach

**Step 1: Update go.mod**
```bash
# Change module declaration
sed -i 's|module github.com/autom8y/knossos|module github.com/autom8y/knossos|' go.mod
```

**Step 2: Update all import statements**
```bash
# Using sed with find (portable, reliable)
find . -type f -name '*.go' -exec sed -i \
  's|github.com/autom8y/knossos|github.com/autom8y/knossos|g' {} \;
```

**Step 3: Regenerate go.sum**
```bash
go mod tidy
```

**Step 4: Verify**
```bash
go build ./...
go test ./...
```

### 3.4 Alternative: gofiximports

```bash
# Install
go install github.com/semk/gofiximports@latest

# Run
gofiximports -dir . \
  -from "github.com/autom8y/knossos" \
  -to "github.com/autom8y/knossos"
```

**Advantages**: Go-aware, respects gofmt, handles edge cases
**Disadvantages**: External dependency, less widely known

### 3.5 Migration Verification Checklist

- [ ] `go.mod` module path updated
- [ ] All `*.go` files import paths updated
- [ ] `go mod tidy` completes without errors
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (all tests)
- [ ] `git diff` shows only expected changes
- [ ] No orphaned `github.com/autom8y/knossos` references remain

---

## 4. Monorepo vs Single-Module Analysis

### 4.1 Decision Context

| Consideration | Single-Module | Multi-Module |
|--------------|---------------|--------------|
| Test execution | `go test ./...` | Per-module scripts |
| Dependency management | Unified | Per-module isolation |
| Versioning | Single version | Independent versions |
| Maintenance burden | Low | High |
| Go team guidance | **Recommended** | "Requires great care" |

### 4.2 Knossos Assessment

Knossos is a **single product** (CLI tool + configuration system). There is no need for:
- Independent versioning of subcomponents
- External consumption of internal packages
- Dependency isolation between components

**Recommendation**: Single-module monorepo is the correct choice.

### 4.3 Content Organization Within Single Module

The `rites/` directory contains non-Go assets (YAML, Markdown, templates). These do NOT require separate modules:

```
knossos/
  go.mod                      # Single module
  cmd/ari/
  internal/
  rites/
    hygiene/                  # YAML + Markdown (not Go)
    10x-dev/
    rnd/
```

This aligns with the "server project" pattern from official docs: Go code in `internal/`, non-Go assets at root.

---

## 5. CLI Tool Project Structure (Cobra/Viper)

### 5.1 Cobra Convention

The [Cobra framework](https://github.com/spf13/cobra) generates this default structure:

```
myapp/
  cmd/
    root.go
    add.go
    list.go
  main.go
```

However, for larger projects, the [recommended pattern](https://www.bytesizego.com/blog/structure-go-cli-app) is:

```
myapp/
  cmd/
    myapp/
      main.go           # Entry point
  internal/
    cmd/                # Cobra command implementations
      root.go
      add.go
      list.go
    services/           # Business logic
    config/             # Configuration handling
```

### 5.2 Current Ariadne Structure (Assessment)

```
ariadne/
  cmd/
    ari/
      main.go           # Entry point (minimal)
  internal/
    cmd/                # Cobra commands (65+ files)
      root/
      session/
      rite/
      hook/
      sync/
      ...
    session/            # Business logic
    rite/
    sync/
    ...
```

**Assessment**: Current structure follows best practices:
- Minimal `main.go` (27 lines)
- Commands in `internal/cmd/`
- Business logic in `internal/{domain}/`
- Clear separation of CLI and business logic

### 5.3 No Changes Needed to Internal Structure

The internal organization is already well-structured. The consolidation only needs to:
1. Move `ariadne/` contents to repository root
2. Update module path in `go.mod`
3. Update all import statements

---

## 6. Findings Summary

### 6.1 Project Structure Decision

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Layout | `cmd/` + `internal/` at root | Go community standard |
| Module path | `github.com/autom8y/knossos` | Platform identity (ADR-0009) |
| `pkg/` directory | Not needed | No external library consumers |
| Content location | `rites/` + `templates/` at root | Non-Go assets alongside Go code |
| Internal organization | Keep current | Already well-structured |

### 6.2 Migration Tooling Decision

| Tool | Decision | Rationale |
|------|----------|-----------|
| Primary | `sed` + `find` | Simple, reliable, no dependencies |
| Validation | `go build`, `go test` | Built-in verification |
| Backup | `gofiximports` | If sed approach has edge cases |

### 6.3 Complexity Assessment

| Migration Step | Complexity | Risk | Automation |
|----------------|------------|------|------------|
| Move files to root | Low | Low | `git mv` |
| Update go.mod | Low | Low | Manual edit |
| Update imports | Medium | Low | `sed` scripted |
| Verify build | Low | Low | `go build ./...` |
| Verify tests | Medium | Medium | `go test ./...` |
| Update CI/CD | Low | Low | Path adjustments |

**Total estimated effort**: 2-4 hours for mechanical changes, plus testing validation.

---

## 7. Open Questions (Resolved)

| Question | Resolution |
|----------|------------|
| Official Go standard? | No official standard; community consensus on `cmd/` + `internal/` |
| Best migration tool? | `sed` + `find` is simplest and most reliable |
| Monorepo structure? | Single module is correct for Knossos |
| Internal reorg needed? | No; current structure is already idiomatic |
| Timing relative to rename? | Do restructure first, rename repo second |

---

## 8. Recommendations

### 8.1 Immediate (Pre-Implementation)

1. **Accept ADR-go-module-structure.md** (drafted below)
2. **Create migration script** for import path updates
3. **Document rollback procedure** (git revert)

### 8.2 Implementation Order

1. Move `ariadne/cmd/` to `cmd/`
2. Move `ariadne/internal/` to `internal/`
3. Move `ariadne/go.mod` to root (edit module path)
4. Move `ariadne/go.sum` to root
5. Move remaining `ariadne/` files (test/, docs/)
6. Run import path migration script
7. `go mod tidy`
8. `go build ./...`
9. `go test ./...`
10. Remove empty `ariadne/` directory
11. Update CI/CD paths
12. Update justfile build commands

### 8.3 Deferred

- Repository rename (`roster` -> `knossos`) - per ADR-0009 criteria
- Terminology migration (`team` -> `rite` in Go code) - bundle with this change

---

## 9. Artifact Verification

| Artifact | Path | Status |
|----------|------|--------|
| This SPIKE | `docs/spikes/SPIKE-go-project-structure.md` | Created |
| ADR draft | `docs/decisions/ADR-go-module-structure.md` | Created (companion) |

---

## 10. Sources

### Official Documentation
- [Organizing a Go module](https://go.dev/doc/modules/layout) - Official Go documentation
- [Migrating to Go Modules](https://go.dev/blog/migrating-to-go-modules) - Go blog

### Community Resources
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) - Community convention repository
- [Go Project Structure: Practices & Patterns (2025)](https://www.glukhov.org/post/2025/12/go-project-structure/) - Recent analysis
- [11 Tips for Structuring Your Go Projects](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects) - Best practices
- [No Nonsense Guide to Go Projects Layout](https://laurentsv.com/blog/2024/10/19/no-nonsense-go-package-layout.html) - Pragmatic guide

### Migration Tooling
- [gofiximports](https://github.com/semk/gofiximports) - Import path replacement tool
- [gorep](https://github.com/novalagung/gorep) - Package name replacement
- [Find and Replace with Sed: Rename Golang Packages](https://wingedrhino.com/2018/07/31/find-and-replace-with-sed-rename-golang-packages/) - sed approach

### CLI Patterns
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Structuring Go Code for CLI Applications](https://www.bytesizego.com/blog/structure-go-cli-app) - CLI structure patterns
- [Building CLI Apps in Go with Cobra & Viper](https://www.glukhov.org/post/2025/11/go-cli-applications-with-cobra-and-viper/) - Cobra/Viper guide

### Go Module Management
- [Go Modules in a Monorepo](https://medium.com/compass-true-north/catching-up-with-the-world-go-modules-in-a-monorepo-c3d1393d6024) - Monorepo patterns
- [Go Module: A Guide for Monorepos (Part 1)](https://engineering.grab.com/go-module-a-guide-for-monorepos-part-1) - Grab engineering

---

**Document Status**: SPIKE COMPLETE
**Next Step**: ADR-go-module-structure.md finalization
**Handoff**: Ready for ecosystem rite implementation
