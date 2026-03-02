# SPIKE: Release & Distribution Best Practices for an Extensible, Future-Proofed Rite (2026)

**Date**: 2026-03-02
**Author**: Claude Opus 4.6 (spike research)
**Status**: Complete
**Timebox**: Single session
**Rite Under Analysis**: `releaser`

---

## Question and Context

### What are we trying to learn?

From first principles in 2026, what best practices, concepts, decisions, designs, and patterns should the `releaser` rite adopt to ensure it is:

1. **Optimally tooled** -- leveraging the current state of the art in release engineering
2. **Future-proofed** -- aligned with regulatory, security, and ecosystem trajectories
3. **Mildly opinionated** -- standardized enough to enforce quality, flexible enough to not constrain
4. **Appropriately generalized** -- extensible across any ecosystem (Python/uv, Node/npm, Go, Rust/Cargo, and beyond) and any project's release and distribution lifecycle

### What decision will this inform?

Whether the current releaser rite architecture has structural gaps, whether its agent responsibilities need expansion, and what concrete capabilities should be added (as new agents, agent responsibilities, mena skills, or workflow phases) to bring it to 2026 best-practice parity.

### Current State Summary

The releaser rite currently implements a 5-phase sequential workflow:

```
Reconnaissance -> Dependency Analysis -> Release Planning -> Execution -> Verification
```

With 6 agents: `pythia` (orchestrator), `cartographer` (recon), `dependency-resolver` (DAG), `release-planner` (phased plan), `release-executor` (publish/push/PR), `pipeline-monitor` (CI verification).

It supports 4 ecosystems (Python/uv, Node/npm, Go, Rust/Cargo), 3 complexity levels (PATCH/RELEASE/PLATFORM), DAG-branch failure halting, pipeline chain monitoring, and cross-rite routing.

---

## Approach Taken

1. Deep read of all 6 agents, manifest, orchestrator, workflow, and mena skills
2. External research across 15+ topics via web search of reputable 2025-2026 sources
3. Synthesis into 12 capability domains with gap analysis against current rite
4. Comparison matrix scoring current coverage vs. industry standard
5. Prioritized recommendations with implementation guidance

---

## Findings

### Domain 1: Versioning Strategy

**Industry State (2026)**:
- **SemVer** remains the gold standard for libraries/packages (MAJOR.MINOR.PATCH)
- **CalVer** (e.g., 2026.3.0) gaining adoption for applications and internal services
- **Hybrid strategies** common: libraries use SemVer, applications use CalVer
- Automation via Conventional Commits -> automatic version derivation is table stakes
- Tools: `semantic-release`, `release-please`, `commitizen`, `changesets`, `cargo-release`

**Current Rite Coverage**: PARTIAL
- Cartographer detects current versions from manifests
- Release-planner infers publish commands per ecosystem
- No opinion on versioning scheme, no automation of version derivation from commit history
- No changelog generation

**Gap**: The rite detects versions but does not derive them. It has no concept of "what should the next version be?" based on commit analysis. This is intentionally left to human judgment today, but represents a significant automation opportunity.

**Recommendation**: Add a versioning strategy abstraction to the release-planner or as a new responsibility. Support pluggable schemes (SemVer, CalVer, hybrid) configurable per-repo. Changelog generation should be a first-class artifact in the execution phase.

---

### Domain 2: Supply Chain Security (SLSA, SBOM, Signing)

**Industry State (2026)**:
- **SLSA v1.1** is the dominant framework (L1-L3 levels of build provenance)
- **Trusted Publishing via OIDC** is now standard across npm, PyPI, and crates.io -- long-lived tokens deprecated (npm deprecated Classic Tokens December 2025)
- **Sigstore/Cosign** provides keyless signing with OIDC identity binding
- **SBOM generation** mandated by EU CRA and US Executive Order 14028; CISA 2025 minimum elements
- **CycloneDX** (ECMA-424) and **SPDX** (ISO/IEC 5962) are the two standard formats
- Artifact attestation (provenance + SBOM) is becoming a CI/CD table stake

**Current Rite Coverage**: MINIMAL
- Pipeline-monitor scans for attestation/signing stages in CI chains (pattern matching)
- Cartographer classifies `attest` and `sign` as `deployment_chain` intermediate stages
- No first-class concept of SBOM generation, provenance attestation, or signing verification
- No OIDC/trusted publishing awareness

**Gap**: This is the single largest gap. The rite treats supply chain security as something that *might happen* in a CI workflow, rather than something the rite *ensures and verifies*. As of 2026, releasing without provenance attestation and SBOM is increasingly non-compliant.

**Recommendation (High Priority)**:
1. Add a **supply-chain-verifier** responsibility (could be a new agent or expansion of pipeline-monitor) that:
   - Verifies SLSA provenance attestations exist after publish
   - Confirms Sigstore/Cosign signatures on published artifacts
   - Validates SBOM generation (CycloneDX or SPDX) as part of release artifacts
   - Checks trusted publishing configuration (OIDC) vs. long-lived tokens
2. Add SLSA level as a configurable gate in the verification phase
3. Add supply chain security status to the verification-report schema

---

### Domain 3: Artifact Management and Promotion

**Industry State (2026)**:
- **Immutable artifacts** promoted across environments (build once, deploy many)
- **OCI registries** becoming universal artifact storage (not just containers) -- Helm charts, Terraform modules, WASM, SBOMs, ML models
- **ORAS** (OCI Registry As Storage) enables pushing any artifact type to OCI registries
- **Artifact promotion** (dev -> staging -> production) with signed, immutable records
- **JFrog Release Bundles**, **GitHub Packages**, **GitLab Package Registry** support multi-format

**Current Rite Coverage**: PARTIAL
- Release-executor publishes to ecosystem-specific registries (npm, PyPI, crates.io, Go tags)
- No concept of environment promotion or immutable artifact lifecycle
- No OCI artifact awareness beyond container-related CI chain detection
- No artifact verification post-publish (only CI status)

**Gap**: The rite publishes artifacts but has no concept of artifact lifecycle management. For projects that need staging -> production promotion, the rite has no model.

**Recommendation (Medium Priority)**:
1. Add an optional **promotion** phase between execution and verification for projects requiring multi-environment artifact promotion
2. Add OCI artifact support as a 5th ecosystem type for projects distributing via OCI registries
3. Add post-publish artifact verification (check registry confirms artifact availability)

---

### Domain 4: Progressive Delivery and Deployment Strategies

**Industry State (2026)**:
- **Canary releases**, **blue-green deployments**, **feature flags** are standard
- **Argo Rollouts** and **Flagger** (Flux) provide Kubernetes-native progressive delivery
- **Feature flags** decouple deployment from release -- code deployed dormant, toggled on
- Release !== Deployment -- modern practice separates "publish to registry" from "deploy to users"

**Current Rite Coverage**: OUT OF SCOPE (by design)
- The rite explicitly routes deployment concerns to the **sre** rite
- Pipeline-monitor tracks deployment chains but does not manage deployments
- Cross-rite routing table correctly sends deployment issues to sre

**Gap**: None. This is correctly out of scope. The rite's boundary -- "publish to registries and verify CI" -- is well-drawn. Progressive delivery belongs in SRE/deployment tooling.

**Recommendation**: No change. Document this boundary explicitly in the releaser-ref skill as a design decision. Ensure the handoff to SRE includes sufficient deployment metadata (artifact digests, promotion history, feature flag state).

---

### Domain 5: Quality Gates and Release Readiness

**Industry State (2026)**:
- **Automated quality gates** at each pipeline stage (security, quality, stability)
- **SLI-based gates** (response time, error rate, throughput) for deployment validation
- **Pre-flight checks** before release (auth, registry availability, dependency freshness)
- **Dry-run simulation** of release plans before execution

**Current Rite Coverage**: GOOD
- Cartographer runs `gh auth status` pre-flight check
- Pipeline-monitor applies timeout-based gates on CI
- Verification report has explicit success criteria (`all_ci_green`, `all_chains_resolved`, `all_deployments_healthy`, `all_versions_consistent`, `zero_manual_intervention`)
- Back-routes enable re-planning when execution discovers conflicts

**Gap**: Missing dry-run / simulation capability. The rite can plan and execute but cannot simulate execution without side effects.

**Recommendation (Medium Priority)**:
1. Add a `--dry-run` mode to the release command that:
   - Runs reconnaissance and dependency analysis fully
   - Generates the release plan
   - Validates all publish commands would succeed (e.g., `npm publish --dry-run`, `cargo publish --dry-run`)
   - Reports what *would* happen without executing
2. Add registry availability checks to pre-flight (can we reach npm, PyPI, crates.io?)
3. Add dependency freshness gate (are declared dependencies available at declared versions?)

---

### Domain 6: Conventional Commits and Changelog Generation

**Industry State (2026)**:
- **Conventional Commits** specification is the standard for machine-readable commit messages
- **Commitizen** (Python and Node) enforces commit conventions and generates changelogs
- **Keep a Changelog** format widely adopted
- Automated changelog from commit history is expected for any published package
- `CHANGELOG.md` is a first-class release artifact alongside the published package

**Current Rite Coverage**: PARTIAL
- `commit-conventions` mena skill exists (legomena)
- Release-executor uses conventional commit format for bump commits (`chore(deps): bump...`)
- No automated changelog generation from commit history
- No changelog as a release artifact

**Gap**: Commit conventions are followed but not leveraged. The information is there (in commit history) but never extracted into a changelog.

**Recommendation (Medium Priority)**:
1. Add changelog generation to the release-executor's responsibilities:
   - Extract conventional commits since last release tag
   - Generate `CHANGELOG.md` entry (or append to existing)
   - Include changelog in the release plan artifacts
2. Make this configurable per-repo (some repos may have their own changelog tooling)

---

### Domain 7: Rollback Strategy

**Industry State (2026)**:
- **Immutable rollback**: revert to previous artifact version (container tag, package version)
- **Feature flag rollback**: disable feature without redeployment
- **Database-aware rollback**: forward-compatible migrations, rollback scripts
- **Automated rollback triggers**: metric-based (error rate, latency thresholds)
- Blue-green enables instant traffic switch back

**Current Rite Coverage**: GOOD
- Release-planner defines rollback boundaries per phase
- Rollback plan with `safe_to_rollback` flag and instructions
- DAG-branch failure halting prevents cascading failures
- Back-routes enable re-execution after partial failure

**Gap**: Rollback plan is defined but never executed by the rite. If a rollback is needed, it requires human intervention and a new release cycle.

**Recommendation (Low Priority)**:
1. Add an optional `rollback-executor` capability (could be a mode of release-executor) that:
   - Reads the rollback plan from the release plan
   - Reverts published packages (where registries support it, e.g., npm unpublish within 72h)
   - Reverts version bump commits (git revert)
   - Creates rollback PRs
2. This should be explicitly gated behind user confirmation (never auto-rollback)

---

### Domain 8: Branching Strategy Awareness

**Industry State (2026)**:
- **Trunk-based development** dominant for continuous delivery teams
- **Git Flow** still used for versioned software with long release cycles
- **Short-lived release branches** (just-in-time) as a hybrid approach
- Feature flags enable trunk-based development without exposing incomplete features

**Current Rite Coverage**: MINIMAL
- Cartographer reads current branch but does not assess branching strategy
- Release-executor pushes to the branch found by cartographer
- No awareness of release branch patterns or trunk-based vs. git-flow implications

**Gap**: The rite is branch-agnostic (it works with whatever branch is current), which is a reasonable default. However, it cannot advise on or enforce branching best practices.

**Recommendation (Low Priority)**:
1. Add branching strategy detection to cartographer:
   - Detect if repo uses trunk-based (main only), git-flow (develop/main/release branches), or hybrid
   - Record in platform-state-map for downstream awareness
2. Release-planner can use this to adjust merge strategy (e.g., git-flow repos need release branch creation)
3. Keep this informational, not prescriptive -- different projects legitimately use different strategies

---

### Domain 9: Release Observability and DORA Metrics

**Industry State (2026)**:
- **DORA metrics** (Deployment Frequency, Lead Time, Change Failure Rate, Time to Restore) are the standard framework
- 2025 DORA report emphasizes observability foundations for AI-assisted development
- Release event tracking and deployment tracking are table stakes
- Integration with Datadog, New Relic, GitLab analytics for metrics collection

**Current Rite Coverage**: MINIMAL
- Verification report captures timestamps (started/completed) and status
- Execution ledger logs timing per action
- No aggregation into DORA-style metrics
- No historical release tracking across sessions

**Gap**: Each release is a standalone event. There is no cumulative view of release health, frequency, or failure patterns over time.

**Recommendation (Medium Priority)**:
1. Add a release metrics summary to the verification report:
   - Total release duration (reconnaissance -> verification complete)
   - Per-phase duration
   - Publish success rate
   - CI pass rate
2. Consider a persistent release history file (`.ledge/releases/history.yaml`) that accumulates across sessions
3. This enables DORA metric derivation over time without external tooling

---

### Domain 10: Multi-Registry and Cross-Platform Distribution

**Industry State (2026)**:
- Packages published to multiple registries (npm + GitHub Packages, PyPI + private index)
- Go modules distributed via proxy.golang.org + direct
- Cross-platform binary distribution (goreleaser for Go, pyinstaller, pkg for Node)
- GitHub Releases as a distribution channel alongside registry publish

**Current Rite Coverage**: PARTIAL
- One publish command per repo (from justfile or ecosystem default)
- No concept of multi-registry publish
- No GitHub Releases creation
- No binary distribution support

**Gap**: The rite assumes one registry per package. Projects that need to publish to multiple targets (e.g., npm + GitHub Packages) require manual configuration.

**Recommendation (Medium Priority)**:
1. Extend the release-plan schema to support multiple publish targets per repo:
   ```yaml
   publish_targets:
     - registry: npm
       command: "npm publish"
     - registry: github_packages
       command: "npm publish --registry=https://npm.pkg.github.com"
   ```
2. Add GitHub Releases creation as a standard post-publish action (using `gh release create`)
3. Add binary distribution support as an ecosystem extension for Go (`goreleaser`) and Rust (`cargo-dist`)

---

### Domain 11: Release Communication and Coordination

**Industry State (2026)**:
- Slack/Teams notifications on release events
- GitHub deployment status API for tracking
- Release announcements auto-generated from changelog
- Stakeholder notification workflows (PMs, QA, support)

**Current Rite Coverage**: NONE
- No notification mechanism
- No deployment status API integration
- Human-readable `.md` artifacts serve as the communication channel (user reads them)

**Gap**: The rite produces excellent artifacts but does not push notifications. Communication is pull-based (user reads reports) rather than push-based (rite notifies stakeholders).

**Recommendation (Low Priority)**:
1. Add an optional notification phase (or post-verification hook) that:
   - Posts release summary to a configured Slack webhook
   - Creates GitHub deployment status entries
   - Generates release announcement from changelog + verification report
2. This should be entirely opt-in and configured per-project
3. Consider this as a cross-rite capability (shared notification infrastructure)

---

### Domain 12: AI-Assisted Release Engineering

**Industry State (2026)**:
- **GitHub Agentic Workflows** (technical preview, Feb 2026) automate repo tasks via LLM agents
- AI-assisted CI troubleshooting, test improvements, documentation updates
- Agentic AI for autonomous decision-making in DevOps lifecycle
- Context engineering as a core discipline for reliable AI workflows

**Current Rite Coverage**: EXCELLENT (this IS the model)
- The releaser rite IS an agentic release engineering system
- 6 specialized agents with clear contracts, handoffs, and escalation protocols
- Artifact chain provides structured context flow between agents
- Complexity-gated workflows prevent over-engineering simple releases
- DAG-branch failure halting is sophisticated autonomous decision-making

**Gap**: None. The rite is ahead of the industry curve here. The Knossos agent-based architecture with structured artifact chains, complexity gating, and cross-rite routing is precisely what the industry is moving toward.

**Recommendation**: Continue investing in the agentic model. Key areas to watch:
1. Agent self-evaluation (did the release plan work? what would I change next time?)
2. Learning from release history (pattern matching on failure types)
3. Proactive recommendations (e.g., "this repo hasn't been released in 90 days, dependencies have drifted")

---

## Comparison Matrix

| Domain | Industry Standard (2026) | Current Rite Coverage | Priority | Effort |
|--------|-------------------------|----------------------|----------|--------|
| 1. Versioning Strategy | Automated derivation from commits | Detect only, no derivation | Medium | Medium |
| 2. Supply Chain Security | SLSA + SBOM + Sigstore mandatory | Passive chain detection only | **HIGH** | Large |
| 3. Artifact Promotion | Immutable artifacts, env promotion | Publish to single registry | Medium | Medium |
| 4. Progressive Delivery | Canary/blue-green/feature flags | Correctly out of scope (SRE) | None | N/A |
| 5. Quality Gates | Dry-run, registry checks, SLI gates | Good gates, no dry-run | Medium | Small |
| 6. Changelog Generation | Automated from conventional commits | Conventions followed, no generation | Medium | Small |
| 7. Rollback Strategy | Automated rollback execution | Plan-only, no execution | Low | Medium |
| 8. Branching Awareness | Strategy detection and adaptation | Branch-agnostic (acceptable) | Low | Small |
| 9. Release Observability | DORA metrics, historical tracking | Per-session only | Medium | Medium |
| 10. Multi-Registry Publish | Multiple targets per package | Single registry per package | Medium | Small |
| 11. Release Communication | Push notifications, status APIs | Pull-based artifacts only | Low | Small |
| 12. AI-Assisted Release | Agentic workflows (emerging) | **Ahead of industry** | None | N/A |

---

## Recommendation

### Tier 1: Critical (address in next sprint)

**Supply Chain Security Integration** (Domain 2)
- This is a regulatory and compliance gap, not just a best-practice gap
- SLSA provenance, SBOM generation, and trusted publishing verification should be first-class concerns
- Concrete deliverable: Expand pipeline-monitor or add supply-chain-verifier agent responsibilities; add SLSA/SBOM/signing fields to verification-report schema

### Tier 2: High Value (address in next quarter)

1. **Dry-Run Mode** (Domain 5) -- Low effort, high confidence improvement. Most ecosystem publish commands support `--dry-run`. Wire this through the existing plan-then-execute architecture.

2. **Changelog Generation** (Domain 6) -- Conventional commits are already used. Extracting a changelog is mechanical and produces a valuable release artifact.

3. **Multi-Registry & GitHub Releases** (Domain 10) -- Extend the publish_targets schema to support multiple registries per repo. Add `gh release create` as a standard post-publish action.

### Tier 3: Strategic (address in 6-month horizon)

4. **Automated Version Derivation** (Domain 1) -- Pluggable versioning scheme support (SemVer/CalVer/hybrid) with commit-based version inference.

5. **Release Observability** (Domain 9) -- Persistent release history for DORA metric derivation.

6. **Artifact Promotion** (Domain 3) -- For projects needing staged artifact lifecycle management.

### Tier 4: Nice-to-Have (backlog)

7. **Rollback Execution** (Domain 7) -- Automated rollback from plan.
8. **Branching Strategy Detection** (Domain 8) -- Informational enrichment.
9. **Release Communication** (Domain 11) -- Push notifications via webhooks.

### Not Recommended

- Progressive Delivery (Domain 4) -- Correctly out of scope. Keep the SRE boundary.
- AI-Assisted Release (Domain 12) -- Already ahead of industry. Continue current investment.

---

## Generalization Principles

The following design principles ensure the rite remains extensible across ecosystems:

### 1. Ecosystem as a Plugin, Not a Hardcoded Switch

The current 4-ecosystem matrix (`python_uv`, `node_npm`, `go_mod`, `rust_cargo`) should be treated as the first four implementations of a pluggable ecosystem interface. The interface contract:

```
Ecosystem Interface:
  detect(repo_path) -> bool          # Can this ecosystem handle this repo?
  version(repo_path) -> string       # What is the current version?
  dependencies(repo_path) -> []dep   # What are the cross-repo deps?
  publish_command(repo_path) -> cmd  # How to publish?
  dry_run_command(repo_path) -> cmd  # How to simulate publish?
  version_bump(file, from, to) -> edit  # How to bump a dep version?
  changelog_source(repo_path) -> string # Where to derive changelog?
```

Future ecosystems (Swift/SPM, Elixir/Hex, Java/Maven, .NET/NuGet, PHP/Composer) should plug in without changing agent logic.

### 2. Registry as a Named Target, Not an Implicit Default

Instead of "publish this repo", the model should be "publish this repo to these targets". This enables multi-registry publishing, private registry support, and OCI artifact distribution without architectural changes.

### 3. Security as a Verification Layer, Not an Afterthought

Supply chain security (provenance, signing, SBOM) should be a cross-cutting concern that applies to ALL ecosystems, not an ecosystem-specific bolt-on. The verification phase should check for these regardless of ecosystem.

### 4. Versioning as a Strategy, Not a Value

The rite should understand the *strategy* (SemVer, CalVer, hybrid) not just the current version string. This enables automated version derivation and intelligent bump suggestions.

### 5. Communication as an Optional Sidecar, Not a Core Phase

Release notifications, webhook integrations, and status API updates should be configurable opt-in behaviors, not mandatory workflow phases. Different projects have radically different communication needs.

---

## Follow-Up Actions

1. **ADR**: Write an ADR for supply chain security integration into the releaser rite (SLSA level targeting, SBOM format selection, trusted publishing verification)
2. **Schema update**: Extend `verification-report.yaml` with supply chain security fields
3. **Dry-run spike**: Validate that all 4 ecosystem publish commands support `--dry-run` mode and document the exact flags
4. **Changelog POC**: Build a throwaway POC of conventional-commit-to-changelog extraction for one ecosystem
5. **Ecosystem interface RFC**: Formalize the ecosystem plugin interface to prepare for 5th+ ecosystem support
6. **Release history design**: Design the persistent release history format for DORA metric accumulation

---

## Sources

- [Best CI/CD practices for scalable pipelines (2026)](https://www.kellton.com/kellton-tech-blog/continuous-integration-deployment-best-practices-2025)
- [Continuous Deployment in 2025](https://axify.io/blog/continuous-deployment)
- [Release Management Best Practices (2026)](https://www.apwide.com/release-management-best-practices/)
- [SemVer vs CalVer: Choosing the Best Versioning Strategy](https://sensiolabs.com/blog/2025/semantic-vs-calendar-versioning)
- [How to Implement Semantic Versioning Automation](https://oneuptime.com/blog/post/2026-01-25-semantic-versioning-automation/view)
- [NPM Release Automation: Semantic Release vs Release Please vs Changesets](https://oleksiipopov.com/blog/npm-release-automation/)
- [SLSA Framework (JFrog)](https://jfrog.com/learn/grc/slsa-framework/)
- [SLSA Framework (Wiz)](https://www.wiz.io/academy/application-security/slsa-framework)
- [Using Artifact Signing for SLSA Provenance](https://www.endorlabs.com/learn/using-artifact-signing-to-establish-provenance-for-slsa)
- [SLSA Official Site](https://slsa.dev/)
- [Global Alignment on SBOM Standards (OpenSSF)](https://openssf.org/blog/2025/10/22/sboms-in-the-era-of-the-cra-toward-a-unified-and-actionable-framework/)
- [OWASP CycloneDX](https://cyclonedx.org/)
- [npm Trusted Publishing with OIDC](https://github.blog/changelog/2025-07-31-npm-trusted-publishing-with-oidc-is-generally-available/)
- [PyPI Trusted Publishers](https://docs.pypi.org/trusted-publishers/)
- [crates.io Trusted Publishing (RFC)](https://rust-lang.github.io/rfcs/3691-trusted-publishing-cratesio.html)
- [Sigstore Cosign](https://github.com/sigstore/cosign)
- [Cosign Verification of npm Provenance](https://blog.sigstore.dev/cosign-verify-bundles/)
- [OCI Artifacts Explained](https://oneuptime.com/blog/post/2025-12-08-oci-artifacts-explained/view)
- [ORAS (OCI Registry As Storage)](https://oras.land/)
- [Modern Deployment Rollback Techniques (2025)](https://www.featbit.co/articles2025/modern-deploy-rollback-strategies-2025)
- [Modern Rollback Strategies (Octopus)](https://octopus.com/blog/modern-rollback-strategies)
- [Trunk-Based Development vs Gitflow (Mergify)](https://mergify.com/blog/trunk-based-development-vs-gitflow-which-branching-model-actually-works)
- [Trunk-Based Development (Atlassian)](https://www.atlassian.com/continuous-delivery/continuous-integration/trunk-based-development)
- [GitOps in 2025 (CNCF)](https://www.cncf.io/blog/2025/06/09/gitops-in-2025-from-old-school-updates-to-the-modern-way/)
- [What the 2025 DORA Report Teaches Us](https://www.honeycomb.io/blog/what-2025-dora-report-teaches-us-about-observability-platform-quality)
- [DORA Metrics (New Relic)](https://newrelic.com/blog/observability/dora-metrics)
- [DORA Metrics (Datadog)](https://www.datadoghq.com/knowledge-center/dora-metrics/)
- [publib: Unified Multi-Registry Publishing](https://github.com/cdklabs/publib)
- [Package Manager Design Tradeoffs](https://nesbitt.io/2025/12/05/package-manager-tradeoffs.html)
- [Artifact Promotion Patterns](https://oneuptime.com/blog/post/2026-01-30-artifact-promotion/view)
- [Azure DevOps Quality Gates](https://learn.microsoft.com/en-us/azure/devops/pipelines/release/approvals/gates)
- [Conventional Commits Specification](https://www.conventionalcommits.org/en/about/)
- [Commitizen](https://github.com/commitizen-tools/commitizen)
- [GitHub Agentic Workflows (InfoQ, Feb 2026)](https://www.infoq.com/news/2026/02/github-agentic-workflows/)
- [How to Build Reliable AI Workflows (GitHub Blog)](https://github.blog/ai-and-ml/github-copilot/how-to-build-reliable-ai-workflows-with-agentic-primitives-and-context-engineering/)
- [Conventional Changelog](https://github.com/conventional-changelog/conventional-changelog)
- [Monorepo Version, Tag, and Release Strategy](https://medium.com/streamdal/monorepos-version-tag-and-release-strategy-ce26a3fd5a03)
- [Deployment Artifacts Management (MOSS)](https://moss.sh/deployment/deployment-artifacts-management/)
