---
name: dependencies-criteria
description: "Observation criteria for codebase dependency knowledge capture. Use when: theoros is producing dependency knowledge for .know/, documenting the dependency graph, version health, and vulnerability exposure. Triggers: dependency knowledge criteria, dependency graph observation, go.mod documentation."
---

# Dependencies Observation Criteria

> The theoros observes and documents codebase dependencies -- producing a knowledge reference that enables any CC agent to understand the project's dependency landscape, health signals, and upgrade risks.

## Scope

**Target files**: `go.mod`, `go.sum`, import statements across `./cmd/` and `./internal/`

**Observation focus**: Dependency graph shape, version currency, health signals, and vulnerability exposure that a CC agent needs before adding, upgrading, or removing dependencies.

**NOTE**: This domain uses knowledge-capture grading. Instead of grading compliance ratios ("90% of dependencies are current"), grade the COMPLETENESS of the dependency reference produced. A = comprehensive documentation of the dependency landscape with evidence. F = incomplete or inaccurate documentation.

## Criteria

### Criterion 1: Dependency Graph (weight: 30%)

**What to observe**: Direct and transitive dependencies, their purposes, and how they connect. The knowledge reference must give a reader a mental map of what the project depends on and why.

**Evidence to collect**:
- Parse `go.mod` for all direct dependencies (require block)
- Categorize dependencies by purpose (CLI framework, YAML parsing, testing, templating, etc.)
- Identify the heaviest transitive dependency trees (what pulls in the most sub-dependencies)
- Note any replace directives and their reasons
- Count: total direct deps, total transitive deps, ratio
- Identify standard library vs third-party usage patterns (which packages prefer stdlib)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Every direct dependency documented with purpose and category. Transitive dependency count and heaviest trees identified. Replace directives explained. Stdlib vs third-party patterns noted. An agent could decide whether to add a new dependency. |
| B | 80-89% completeness | Direct dependencies categorized. Most purposes documented. Transitive count present. Minor gaps in replace directive or stdlib analysis. |
| C | 70-79% completeness | Dependencies listed but not all categorized. Transitive analysis missing or superficial. |
| D | 60-69% completeness | Dependencies listed by name only. No categorization or purpose documentation. |
| F | < 60% completeness | Dependency graph not documented or significantly incomplete. |

---

### Criterion 2: Version Currency (weight: 25%)

**What to observe**: How current dependencies are, what the upgrade posture looks like, and where version lag exists. The knowledge reference must tell an agent the project's stance on dependency freshness.

**Evidence to collect**:
- Run or simulate `go list -m -u all` analysis (identify outdated dependencies)
- Categorize version lag: minor behind, major behind, abandoned/archived
- Note any pinned versions and document why (compatibility, breaking changes, etc.)
- Identify the Go version in `go.mod` and whether it's current
- Document any dependency upgrade patterns (regular cadence, ad-hoc, automated)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Currency status documented for all direct dependencies. Outdated packages identified with lag severity. Go version currency noted. Upgrade patterns or lack thereof documented. An agent could plan a dependency upgrade sprint. |
| B | 80-89% completeness | Most dependencies checked for currency. Major lags identified. Go version noted. Minor gaps in upgrade pattern documentation. |
| C | 70-79% completeness | Some currency analysis present but not comprehensive. Go version mentioned but lag not categorized. |
| D | 60-69% completeness | Currency mentioned vaguely ("some deps are old") without specifics. |
| F | < 60% completeness | Version currency not analyzed. |

---

### Criterion 3: Health Signals (weight: 20%)

**What to observe**: Dependency project health — maintenance status, community activity, bus factor, license compatibility. The knowledge reference must flag dependencies that pose risk.

**Evidence to collect**:
- For each direct dependency, note: last release date, maintenance status (active/maintained/archived/abandoned)
- Identify single-maintainer dependencies (bus factor = 1)
- Check license compatibility (MIT, Apache 2.0, BSD, GPL — note any viral licenses)
- Note any dependencies with known deprecation notices or migration paths
- Flag dependencies that are forks or have known alternatives

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Health signals documented for all direct dependencies. Maintenance status, license, and bus factor assessed. Risk flags raised for concerning dependencies. An agent could assess risk before adding a dependency. |
| B | 80-89% completeness | Health signals for most direct dependencies. Major risks identified. Minor gaps in bus factor or license analysis. |
| C | 70-79% completeness | Some health assessment present but not systematic across all dependencies. |
| D | 60-69% completeness | Health mentioned for a few dependencies without systematic assessment. |
| F | < 60% completeness | Dependency health not assessed. |

---

### Criterion 4: Vulnerability Exposure (weight: 15%)

**What to observe**: Known vulnerabilities, security advisory history, and the project's vulnerability management posture. The knowledge reference must tell an agent the security state of dependencies.

**Evidence to collect**:
- Run or simulate `govulncheck` analysis
- Document any known CVEs affecting direct or transitive dependencies
- Note whether the project uses any vulnerability scanning in CI
- Identify dependencies with past vulnerability history (even if currently patched)
- Document the project's dependency security practices

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Vulnerability state documented for all dependencies. Known CVEs listed. CI scanning status noted. Security practices documented. An agent could assess current vulnerability exposure. |
| B | 80-89% completeness | Vulnerability analysis present for most dependencies. Major CVEs noted. Minor gaps in CI or practice documentation. |
| C | 70-79% completeness | Some vulnerability information present but not comprehensive. |
| D | 60-69% completeness | Vulnerability exposure mentioned vaguely without specific CVE or scanning details. |
| F | < 60% completeness | Vulnerability exposure not assessed. |

---

### Criterion 5: Vendoring Strategy (weight: 10%)

**What to observe**: How the project manages dependency availability — vendoring, module proxy, checksum database usage. The knowledge reference must document the build reproducibility posture.

**Evidence to collect**:
- Check for vendor/ directory presence and whether `go mod vendor` is used
- Note GONOSUMCHECK, GONOSUMDB, or GOPROXY settings if documented
- Document whether builds require network access or can build offline
- Note any build constraints related to dependencies (CGO_ENABLED, build tags)
- Check for dependency injection patterns (interfaces wrapping third-party libs)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Vendoring strategy fully documented. Proxy/checksum settings noted. Build reproducibility posture clear. Dependency wrapping patterns documented. An agent could set up a build environment correctly. |
| B | 80-89% completeness | Vendoring and build strategy documented. Minor gaps in proxy or wrapping patterns. |
| C | 70-79% completeness | Basic vendoring information present but build strategy incomplete. |
| D | 60-69% completeness | Vendoring mentioned without detail on build strategy. |
| F | < 60% completeness | Vendoring strategy not documented. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- Dependency Graph: A (midpoint 95%) x 30% = 28.5
- Version Currency: B (midpoint 85%) x 25% = 21.25
- Health Signals: B (midpoint 85%) x 20% = 17.0
- Vulnerability Exposure: C (midpoint 75%) x 15% = 11.25
- Vendoring Strategy: A (midpoint 95%) x 10% = 9.5
- **Total: 87.5 -> B**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture
- [conventions-criteria](conventions.lego.md) -- Codebase conventions knowledge capture
