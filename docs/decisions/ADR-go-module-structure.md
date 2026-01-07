# ADR-0014: Go Module Structure

| Field | Value |
|-------|-------|
| **Status** | ACCEPTED |
| **Date** | 2026-01-07 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |
| **Related** | ADR-0009-knossos-roster-identity |

## Context

The Knossos platform currently maintains its Go CLI (`ari`) in a subdirectory structure:

```
roster/
  ariadne/
    go.mod         # module github.com/autom8y/knossos
    cmd/ari/
    internal/
```

This structure emerged during rapid development when Ariadne was treated as a separate subproject. The SPIKE-knossos-consolidation-architecture.md alignment session confirmed:

1. **Knossos is the parent project**; Ariadne is its CLI component
2. **Module path should reflect platform identity**: `github.com/autom8y/knossos`
3. **Breaking changes are acceptable** (greenfield restructure)
4. **Repository rename timing**: Restructure first, rename to `knossos` second

### Current State

| Aspect | Current | Issue |
|--------|---------|-------|
| Module path | `github.com/autom8y/knossos` | Identity mismatch |
| Go location | `ariadne/` subdirectory | Non-standard nesting |
| Import paths | 450+ occurrences in 150 files | Requires bulk update |
| Build commands | `cd ariadne && go build` | Extra navigation |

### Forces

- **Go Idiom**: Standard Go projects use `cmd/` + `internal/` at root
- **Platform Identity**: ADR-0009 establishes Knossos as the canonical name
- **Tool Compatibility**: IDE navigation, `go install`, CI/CD all expect root-level Go
- **Migration Cost**: 450+ import path changes, but automatable
- **Breaking Change Tolerance**: Confirmed as acceptable

## Decision

### Structure Change

Adopt standard Go project layout with `cmd/` + `internal/` at repository root:

```
knossos/                          # Repository root (after rename)
  go.mod                          # module github.com/autom8y/knossos
  go.sum
  cmd/
    ari/
      main.go                     # CLI entry point
  internal/
    cmd/                          # Cobra command implementations
      root/
      session/
      rite/
      hook/
      sync/
      ...
    session/                      # Business logic
    rite/
    sync/
    artifact/
    ...
  rites/                          # Non-Go content
    hygiene/
    10x-dev/
    rnd/
    ecosystem/
  templates/                      # Hook templates
  docs/
  schemas/
```

### Module Path

Change from `github.com/autom8y/knossos` to `github.com/autom8y/knossos`.

Rationale:
- Aligns with platform identity (ADR-0009)
- Module name matches repository name (post-rename)
- Single authoritative namespace

### Migration Method

Use `sed` + `find` for bulk import path replacement:

```bash
# Step 1: Move files to root
git mv ariadne/cmd cmd
git mv ariadne/internal internal
git mv ariadne/go.mod go.mod
git mv ariadne/go.sum go.sum
git mv ariadne/test test

# Step 2: Update module declaration
sed -i 's|module github.com/autom8y/knossos|module github.com/autom8y/knossos|' go.mod

# Step 3: Update all imports
find . -type f -name '*.go' -exec sed -i \
  's|github.com/autom8y/knossos|github.com/autom8y/knossos|g' {} \;

# Step 4: Regenerate dependencies
go mod tidy

# Step 5: Verify
go build ./...
go test ./...
```

### What We Are NOT Doing

| Rejected Option | Reason |
|-----------------|--------|
| Multi-module monorepo | Unnecessary complexity; single product |
| Keep `ariadne/` subdirectory | Non-idiomatic; breaks standard tooling |
| Use `pkg/` directory | No external library consumers |
| Use `gopls rename` | Not designed for module path changes |
| Keep current module path | Identity mismatch with platform name |

## Consequences

### Positive

1. **Standard Go Layout**: Matches community expectations and tooling
2. **Simpler Commands**: `go build ./cmd/ari` instead of `cd ariadne && go build ./cmd/ari`
3. **IDE Compatibility**: Better gopls/VSCode/GoLand support at root
4. **Platform Identity**: Module path reflects Knossos identity
5. **CI/CD Simplification**: Standard paths reduce configuration complexity
6. **Future Proofing**: Aligned with repo rename to `knossos`

### Negative

1. **Breaking Change**: All import paths change (450+ occurrences)
2. **One-Time Migration Cost**: 2-4 hours of mechanical work + testing
3. **External References**: Any external code importing `github.com/autom8y/knossos` breaks
4. **CI/CD Updates**: Build paths need adjustment

### Neutral

1. **Internal Structure Preserved**: `internal/cmd/`, `internal/session/`, etc. remain unchanged
2. **Test Organization**: Tests move with their packages
3. **Documentation Updates**: Paths in docs need updating

## Implementation Notes

### Pre-Migration Checklist

- [ ] All tests passing on current structure
- [ ] Git working tree clean
- [ ] CI/CD pipeline documented for path updates
- [ ] Justfile commands inventoried
- [ ] Documentation references cataloged

### Migration Script

```bash
#!/bin/bash
# migrate-module-structure.sh
set -euo pipefail

echo "=== Knossos Module Structure Migration ==="

# Verify clean working tree
if [[ -n $(git status --porcelain) ]]; then
  echo "ERROR: Working tree not clean. Commit or stash changes first."
  exit 1
fi

# Step 1: Move directories
echo "Moving directories to root..."
git mv ariadne/cmd cmd
git mv ariadne/internal internal
git mv ariadne/test test
git mv ariadne/go.mod go.mod
git mv ariadne/go.sum go.sum

# Move remaining files
for f in ariadne/*; do
  if [[ -f "$f" ]]; then
    git mv "$f" .
  fi
done

# Remove empty ariadne directory
rmdir ariadne 2>/dev/null || true

# Step 2: Update module path
echo "Updating module path..."
sed -i '' 's|module github.com/autom8y/knossos|module github.com/autom8y/knossos|' go.mod

# Step 3: Update all imports
echo "Updating import paths ($(find . -name '*.go' | wc -l | tr -d ' ') files)..."
find . -type f -name '*.go' -exec sed -i '' \
  's|github.com/autom8y/knossos|github.com/autom8y/knossos|g' {} \;

# Step 4: Regenerate
echo "Running go mod tidy..."
go mod tidy

# Step 5: Verify
echo "Verifying build..."
go build ./...

echo "Running tests..."
go test ./...

echo "=== Migration Complete ==="
echo "Review changes with: git diff --stat"
echo "Commit with: git commit -m 'feat: migrate to github.com/autom8y/knossos module structure'"
```

### Post-Migration Verification

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (all tests)
- [ ] `go install ./cmd/ari` works
- [ ] No `github.com/autom8y/knossos` references remain
- [ ] Justfile commands updated and working
- [ ] CI/CD pipeline passes
- [ ] Documentation paths updated

### Rollback Procedure

```bash
# If migration fails, revert to previous commit
git reset --hard HEAD~1

# Or if committed but issues found
git revert HEAD
```

## Alternatives Considered

### Alternative 1: Keep `ariadne/` Subdirectory

**Description**: Maintain current structure, only update module path.

**Pros**:
- No file moves required
- Smaller change scope

**Cons**:
- Non-idiomatic Go structure
- Requires `cd ariadne` for all Go commands
- IDE/tooling friction
- Signals "separate project" when it's not

**Rejected**: The nesting adds friction and doesn't match Go conventions.

### Alternative 2: Multi-Module Monorepo

**Description**: Create separate modules for `rites/`, `templates/`, etc.

**Pros**:
- Independent versioning per component
- Dependency isolation

**Cons**:
- Unnecessary complexity (Knossos is single product)
- Requires multi-module tooling scripts
- Go team explicitly recommends against unless necessary

**Rejected**: Single module is correct for single product.

### Alternative 3: Use `github.com/knossos-platform/knossos`

**Description**: Create new GitHub organization for cleaner namespace.

**Pros**:
- Cleaner namespace
- Fresh start

**Cons**:
- Requires GitHub org creation
- External coordination overhead
- Migration from existing org

**Rejected**: Alignment session confirmed `github.com/autom8y` as the namespace.

### Alternative 4: Defer Until Repository Rename

**Description**: Wait for `roster` -> `knossos` rename, do everything at once.

**Pros**:
- Single breaking change event
- Repository name matches module path immediately

**Cons**:
- Delays cleanup
- ADR-0009 requires restructure before rename eligibility
- Compounds risk by bundling changes

**Rejected**: Sequential changes (restructure, then rename) reduces risk.

## Related Decisions

- **ADR-0009**: Knossos-Roster Identity - Establishes platform naming
- **SPIKE-knossos-consolidation-architecture**: Alignment session outcomes

## References

- [Organizing a Go module](https://go.dev/doc/modules/layout) - Official Go documentation
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) - Community conventions
- [Go Project Structure: Practices & Patterns (2025)](https://www.glukhov.org/post/2025/12/go-project-structure/)
- [SPIKE-go-project-structure.md](../spikes/SPIKE-go-project-structure.md) - Research findings

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-01-07 | Technology Scout (R&D) | Initial draft |
