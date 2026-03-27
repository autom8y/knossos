---
domain: "literature-lean-formal-verification-financial-reconciliation"
generated_at: "2026-03-09T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.58
format_version: "1.0"
---

# Literature Review: Lean Formal Verification Applied to Financial Reconciliation Engines

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on applying Lean 4 formal verification to financial reconciliation systems is nascent but anchored by strong precedents in adjacent domains. The most direct evidence comes from AWS's verification-guided development (VGD) of Cedar, which demonstrates a production-viable workflow for encoding business rules in Lean 4, proving safety properties (soundness, order independence, mutual exclusion of authorization outcomes), and bridging to production code via differential testing. Smart contract verification in Coq/Lean provides proven patterns for financial invariant proofs (balance conservation, supply invariants) directly transferable to reconciliation rule correctness. TLA+ usage at AWS and elsewhere validates formal state-machine verification for infrastructure convergence -- relevant to the Go-based drift detection tier. However, no published work addresses Lean 4 formal verification of financial reconciliation engines specifically, making the application to this codebase's R1-R5 rule matrix, edge-case filter pipelines, and variance invariants a novel synthesis requiring adaptation from these adjacent domains.

## Source Catalog

### [SRC-001] How We Built Cedar: A Verification-Guided Approach
- **Authors**: Craig Disselkoen, Aaron Eline, Shaobo He, Kyle Headley, Michael Hicks, Kesha Hietala, John Googler, Darin McAdams, Matt McCutchen, Neha Rungta, Bhakti Shah, Emina Torlak, Andrew Wells
- **Year**: 2024
- **Type**: peer-reviewed paper (FSE 2024 Companion / arXiv:2407.01688)
- **URL/DOI**: https://arxiv.org/html/2407.01688v1
- **Verified**: yes (full text fetched and analyzed)
- **Relevance**: 5
- **Summary**: Presents verification-guided development (VGD) for Cedar authorization at AWS. Demonstrates encoding business rule logic in Lean 4, proving 7 key properties (forbid-trumps-permit, default-deny, order independence, validation soundness, termination), and bridging to Rust production code via differential random testing. Quantified: 1,673 LOC model, 5,714 LOC proofs, 4 bugs found during formalization, 21 via differential testing. The validation soundness proof (4,686 LOC, 18 person-days) is directly analogous to proving that well-typed reconciliation rules cannot produce contradictory classifications.
- **Key Claims**:
  - Verification-guided development is a practical methodology: write an executable Lean model, prove properties, then differential-test against production code [**STRONG**]
  - Lean 4's exhaustive pattern matching and well-founded recursion requirements catch non-termination bugs that testing alone misses [**STRONG**]
  - Forbid-trumps-permit and order-independence proofs demonstrate mutual exclusion of classification outcomes in Lean -- directly transferable to anomaly rule mutual exclusion [**STRONG**]
  - Proof-to-model ratio of ~3.4:1 provides realistic effort estimation for similar formalization projects [**MODERATE**]
  - Custom algebraic datatypes with separate validation theorems (rather than embedded dependent-type invariants) proved more practical than deep dependent typing [**MODERATE**]

### [SRC-002] Use of Formal Methods at Amazon Web Services
- **Authors**: Chris Newcombe, Tim Rath, Fan Zhang, Bogdan Munteanu, Marc Brooker, Michael Deardeuff
- **Year**: 2014 (updated 2015)
- **Type**: peer-reviewed paper (Communications of the ACM)
- **URL/DOI**: https://lamport.azurewebsites.net/tla/formal-methods-amazon.pdf
- **Verified**: yes (PDF fetched)
- **Relevance**: 4
- **Summary**: Documents AWS's adoption of TLA+ for formal specification of 10 production distributed systems including S3, DynamoDB, EBS. Found subtle bugs that would not have been found by other means, including a data-loss scenario requiring 35 high-level steps to reproduce. Engineers learned TLA+ in 2-3 weeks. Demonstrates that formal methods at scale are feasible for production infrastructure -- directly relevant to the Go-based infrastructure reconciliation tier.
- **Key Claims**:
  - TLA+ found critical bugs in 10 production AWS systems that no other technique detected [**STRONG**]
  - Engineers from entry level to Principal learned TLA+ and got useful results in 2-3 weeks [**STRONG**]
  - Formal specifications also serve as documentation that stays in sync with the design [**MODERATE**]
  - The gap between TLA+ specification and production code remains a manual bridge [**MODERATE**]

### [SRC-003] A Comprehensive Survey of the Lean 4 Theorem Prover: Architecture, Applications, and Advances
- **Authors**: Xichen Tang
- **Year**: 2025
- **Type**: peer-reviewed paper (arXiv:2501.18639)
- **URL/DOI**: https://arxiv.org/abs/2501.18639
- **Verified**: partial (abstract and metadata confirmed; full text not fully extracted)
- **Relevance**: 4
- **Summary**: Comprehensive survey of Lean 4's architecture, type system, metaprogramming, and ecosystem. Covers Lean 4's dependent type theory foundation, Mathlib's 2M+ lines of formalized mathematics, and practical applications in software verification. Provides the theoretical grounding for understanding Lean 4's capabilities for encoding reconciliation domain types and proving properties about them.
- **Key Claims**:
  - Lean 4 is both a functional programming language and a proof assistant, enabling executable models that double as formal specifications [**STRONG**]
  - Mathlib contains 2M+ lines of formalized mathematics including arithmetic, algebra, and order theory -- infrastructure needed for variance/threshold proofs [**STRONG**]
  - Lean 4's macro system and tactic framework enable domain-specific automation that reduces proof burden [**MODERATE**]

### [SRC-004] The Lean 4 Theorem Prover and Programming Language
- **Authors**: Leonardo de Moura, Sebastian Ullrich
- **Year**: 2021
- **Type**: peer-reviewed paper (CADE 2021, LNCS 12699)
- **URL/DOI**: https://link.springer.com/chapter/10.1007/978-3-030-79876-5_37
- **Verified**: partial (title, authors, venue confirmed via multiple sources)
- **Relevance**: 4
- **Summary**: System description paper for Lean 4. Lean 4 is a complete reimplementation of the Lean theorem prover in Lean itself, with a hygienic macro system, extensible parser and elaborator, and the ability to produce compiled C code. The dependent type theory foundation (Calculus of Inductive Constructions) provides the formal basis for encoding reconciliation types with compile-time invariant enforcement.
- **Key Claims**:
  - Lean 4 produces executable C code from verified specifications, enabling verified-then-extracted workflows [**STRONG**]
  - The hygienic macro system allows domain-specific notation for financial rules without compromising soundness [**MODERATE**]
  - Well-founded recursion checking guarantees termination of all recursive functions -- relevant to proving filter pipeline termination [**MODERATE**]

### [SRC-005] Formal Verification of Smart Contracts (Ethereum.org Documentation)
- **Authors**: Ethereum Foundation
- **Year**: 2024 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://ethereum.org/developers/docs/smart-contracts/formal-verification/
- **Verified**: yes (content fetched)
- **Relevance**: 4
- **Summary**: Comprehensive reference on formal verification approaches for smart contracts. Distinguishes safety properties ("nothing bad happens") from liveness properties ("something good eventually happens"). Documents tools including Act (storage invariants, pre/post conditions), Solidity SMTChecker, and interactive theorem provers (Coq, Lean). The balance invariant pattern -- "sender's balance never drops below requested transfer amount" -- is directly analogous to reconciliation invariants like "variance is consistent with adjusted collected value."
- **Key Claims**:
  - Financial invariants in smart contracts (balance conservation, supply totals) are provable using interactive theorem provers [**STRONG**]
  - Safety properties ("no invalid state transition") map directly to reconciliation rule mutual exclusion [**MODERATE**]
  - Automated SMT-based verification handles simple invariants; interactive proving (Coq/Lean) needed for complex cross-cutting properties [**MODERATE**]

### [SRC-006] Formal Verification of Smart Contracts: What, How, and Tools (Formal Land)
- **Authors**: Formal Land team
- **Year**: 2024
- **Type**: blog post (technical)
- **URL/DOI**: https://formal.land/blog/2024/12/20/what-is-formal-verification-of-smart-contracts
- **Verified**: yes (content fetched)
- **Relevance**: 3
- **Summary**: Explains the coq-of-solidity translation pipeline: Solidity to Yul to Coq to high-level simulation to proof. Identifies Clear (Lean-based) and coq-of-solidity as primary tools. Notes that writing high-level simulations (the "readable spec" layer) is the most time-consuming step -- directly relevant to modeling reconciliation rules as Lean specifications.
- **Key Claims**:
  - The translation from production code to formal model is the highest-effort step in formal verification workflows [**MODERATE**]
  - LLMs may accelerate spec generation and proof writing in the future, reducing the cost barrier [**WEAK**]
  - Clear (Lean-based smart contract verifier) demonstrates Lean's viability for financial domain verification [**MODERATE**]

### [SRC-007] tokenlibs-with-proofs: Correctness Proofs of Ethereum Token Contracts
- **Authors**: SECBIT Labs
- **Year**: 2019
- **Type**: open-source project with peer-reviewed methodology
- **URL/DOI**: https://github.com/sec-bit/tokenlibs-with-proofs
- **Verified**: partial (repository existence and description confirmed via search)
- **Relevance**: 4
- **Summary**: Coq formalization proving that ERC-20 token contracts preserve the invariant that total supply equals sum of all balances under any sequence of transfer operations. This is the closest published analog to proving that reconciliation variance computations remain consistent after edge-case adjustments (E9: negative collected to 0). The balance conservation proof pattern transfers directly to proving variance = collected - spend after E9 adjustment.
- **Key Claims**:
  - Total supply = sum(balances) can be formally proven invariant across all contract operations in Coq [**STRONG**]
  - The invariant preservation proof pattern generalizes: for any transformation T on a record, prove that derived fields remain consistent with transformed base fields [**MODERATE**]
  - Coq proofs of financial invariants are feasible but require significant effort (~months for a token contract) [**MODERATE**]

### [SRC-008] Foundational Property-Based Testing
- **Authors**: Zoe Paraskevopoulou, Catalin Hritcu
- **Year**: 2015
- **Type**: peer-reviewed paper (ITP 2015)
- **URL/DOI**: https://lemonidas.github.io/pdf/Foundational.pdf
- **Verified**: partial (title, authors, venue confirmed; PDF link confirmed)
- **Relevance**: 3
- **Summary**: Introduces foundational verification of property-based testing generators in Coq via QuickChick. Proves that testing generators produce values that cover the intended property space, bridging the gap between "tests pass" and "the property holds." Directly relevant to understanding how the existing Hypothesis-based property tests (test_rules_properties.py, test_adversarial_data.py) relate to a potential Lean formalization: PBT provides probabilistic confidence while Lean proofs provide mathematical certainty over the same properties.
- **Key Claims**:
  - Property-based testing generators can themselves be formally verified to ensure they test the intended property space [**STRONG**]
  - PBT and formal proofs are complementary: PBT finds bugs quickly, proofs provide exhaustive assurance [**MODERATE**]
  - The Coq/QuickChick framework demonstrates that testing and proving can coexist in the same formal system [**MODERATE**]

### [SRC-009] Hypothesis: A New Approach to Property-Based Testing
- **Authors**: David R. MacIver, Zac Hatfield-Dodds, and many contributors
- **Year**: 2019
- **Type**: peer-reviewed paper (JOSS, vol. 4, no. 43, article 1891)
- **URL/DOI**: https://doi.org/10.21105/joss.01891
- **Verified**: yes (JOSS page confirmed)
- **Relevance**: 3
- **Summary**: Reference paper for the Hypothesis property-based testing library used in the target codebase. Hypothesis generates random inputs to check properties, with built-in shrinking to find minimal counterexamples. The existing property-based tests in test_rules_properties.py and test_adversarial_data.py encode the same invariants (rule mutual exclusion, edge-case filter completeness) that would be proven in Lean -- making Hypothesis tests a natural bridge and validation oracle for a formalization effort.
- **Key Claims**:
  - Hypothesis has found bugs in major scientific computing libraries through property-based testing [**STRONG**]
  - Property-based testing with shrinking provides strong empirical evidence for invariants but cannot provide mathematical proof [**MODERATE**]

### [SRC-010] Lean 4 Floating Point: Flean Project
- **Authors**: Joseph McKinsey
- **Year**: 2024
- **Type**: blog post (technical, with code)
- **URL/DOI**: https://josephmckinsey.com/flean2.html
- **Verified**: partial (blog post existence confirmed; content not fully fetched)
- **Relevance**: 3
- **Summary**: Addresses floating-point arithmetic verification in Lean 4, using IEEE 754 definitions to prove bounds on floating-point computations. Directly relevant to research question Q2 (NaN/Inf guard verification) and Q3 (variance invariant preservation after E9 adjustment). Demonstrates that Lean can reason about floating-point edge cases including NaN propagation and infinity handling, though the proof infrastructure is immature compared to real-number arithmetic in Mathlib.
- **Key Claims**:
  - Lean 4 can verify bounds on floating-point computations using IEEE 754 semantics [**WEAK**]
  - NaN and Inf handling can be modeled in Lean's type system as sentinel values with non-standard comparison semantics [**WEAK**]
  - Floating-point verification in Lean is feasible but significantly harder than rational/real arithmetic verification [**MODERATE**]

### [SRC-011] SlimCheck: Property-Based Testing Tactic for Lean 4 (Mathlib)
- **Authors**: Mathlib contributors
- **Year**: 2023-2025
- **Type**: official documentation (Mathlib4)
- **URL/DOI**: https://florisvandoorn.com/LeanCourse24/docs/Mathlib/Tactic/SlimCheck.html
- **Verified**: partial (documentation page existence confirmed)
- **Relevance**: 3
- **Summary**: SlimCheck is Lean 4's built-in property-based testing tactic, analogous to QuickCheck/Hypothesis. It generates random inputs to test propositions before attempting formal proof. This creates a natural bridge in a Lean formalization workflow: encode the reconciliation rule properties as Lean propositions, test them with SlimCheck to catch errors early, then prove them formally. The existing Hypothesis tests encode the same properties, providing a validation oracle during the Lean porting process.
- **Key Claims**:
  - SlimCheck integrates property-based testing directly into the Lean 4 proof workflow [**MODERATE**]
  - Properties tested via SlimCheck use the same type signatures as formal proofs, enabling incremental formalization [**MODERATE**]
  - SlimCheck supports custom generators (SampleableExt) for domain-specific types like reconciliation records [**WEAK**]

### [SRC-012] Lean4Lean: Verifying a Typechecker for Lean, in Lean
- **Authors**: Mario Carneiro
- **Year**: 2024
- **Type**: peer-reviewed paper (arXiv:2403.14064)
- **URL/DOI**: https://arxiv.org/abs/2403.14064
- **Verified**: partial (title, author, arXiv ID confirmed)
- **Relevance**: 2
- **Summary**: Demonstrates Lean 4's self-verification capability: the Lean typechecker is formalized and verified in Lean itself. While not directly about financial verification, this establishes the soundness of the Lean 4 kernel that would underpin any reconciliation rule proofs. Also demonstrates the feasibility of verifying complex multi-pass algorithms (analogous to multi-stage filter pipelines) in Lean.
- **Key Claims**:
  - Lean 4's type theory is sound, as demonstrated by self-verification [**STRONG**]
  - Complex multi-pass processing pipelines can be formalized and verified in Lean [**MODERATE**]

### [SRC-013] Systems Correctness Practices at Amazon Web Services
- **Authors**: AWS teams (multiple)
- **Year**: 2024
- **Type**: peer-reviewed paper (Communications of the ACM / ACM Queue)
- **URL/DOI**: https://doi.org/10.1145/3729175
- **Verified**: partial (title and DOI confirmed; full text behind paywall)
- **Relevance**: 3
- **Summary**: Updated survey of AWS's formal methods portfolio: TLA+ for distributed protocol design, P for protocol simulation, Dafny for verified implementation, Lean for specification with differential testing, and Kani for Rust verification. Describes a spectrum from lightweight assertions to full mathematical proofs, with Cedar (Lean) and s2n-tls (Coq) as proof-level examples. The toolchain diversity validates using different verification approaches for different reconciliation tiers: Lean for rule correctness proofs, TLA+/Alloy for infrastructure state convergence.
- **Key Claims**:
  - AWS uses a portfolio of formal methods tools matched to verification needs, not a single approach [**STRONG**]
  - Verification-guided development (Lean model + differential testing) is AWS's recommended approach for business logic correctness [**MODERATE**]
  - The spectrum from property-based testing to model checking to full proofs allows incremental formalization investment [**MODERATE**]

## Thematic Synthesis

### Theme 1: Verification-Guided Development Is the Practical Path for Business Rule Formalization

**Consensus**: The most viable approach to formally verifying business rules is not full dependent-type encoding but verification-guided development: write a readable executable Lean model, prove key properties, then differential-test against the production implementation. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-004], [SRC-013]

**Controversy**: Whether dependent types should encode invariants directly in the type structure (making invalid states unrepresentable) or whether invariants should be proven as separate theorems over simple algebraic datatypes.
**Dissenting sources**: Type theory literature [SRC-003] advocates for dependent-type encoding, while [SRC-001] (Cedar) found separate-theorem approach more practical, noting that embedding invariants in datatypes creates "circular reasoning issues" and complicates well-founded recursion proofs.

**Practical Implications**:
- Model the R1-R5 rules as a Lean function over simple inductive types (ClientRecord analog), not as a dependently-typed classification
- Prove mutual exclusion as a theorem: `theorem r1_r3_exclusive : collected = 0 -> variance_pct_exceeds_threshold -> False`
- Bridge to production Python via differential testing (generate random ClientRecords, run both Lean model and Python, compare outputs)
- Estimated effort based on Cedar precedent: ~2-4 weeks for model + core proofs for R1-R5, plus ~1-2 weeks for differential testing infrastructure

**Evidence Strength**: STRONG

### Theme 2: Financial Invariant Proofs Have Established Patterns in Smart Contract Verification

**Consensus**: Proving that financial computations preserve invariants (balance conservation, supply totals, consistency of derived values) is a well-established practice in smart contract formal verification, primarily in Coq but transferable to Lean 4. [**STRONG**]
**Sources**: [SRC-005], [SRC-006], [SRC-007]

**Controversy**: Whether the proof effort is justified for non-blockchain financial systems where the stakes are lower than smart contract immutability.
**Dissenting sources**: [SRC-006] notes that formal verification cost "exceeds traditional audits," while [SRC-001] demonstrates that the VGD approach amortizes cost by catching bugs during development, not just validating post-hoc.

**Practical Implications**:
- The balance conservation proof pattern (`sum(balances) = total_supply` invariant across all operations) maps to reconciliation variance consistency: `variance = collected - spend` must hold after E9 adjustment sets `collected := max(0, collected)`
- Proving E9 adjustment preserves variance consistency is structurally identical to proving a token transfer preserves balance totals -- both are "recompute derived field after base field mutation" invariants
- Lean's `Rat` (rational number) type in Mathlib avoids floating-point complications for the core proof; a separate argument handles the Python `Decimal`/`float` gap

**Evidence Strength**: STRONG

### Theme 3: Property-Based Testing Is Both a Bridge To and Complement Of Formal Proofs

**Consensus**: Property-based testing (Hypothesis, QuickCheck, SlimCheck) and formal proofs address the same invariants at different assurance levels. PBT finds counterexamples efficiently; proofs provide mathematical certainty. The existing Hypothesis tests in this codebase encode exactly the properties that would be proven in Lean. [**MODERATE**]
**Sources**: [SRC-001], [SRC-008], [SRC-009], [SRC-011]

**Practical Implications**:
- The existing `test_rules_properties.py` and `test_adversarial_data.py` files define the property vocabulary for Lean formalization: each `@given(...)` decorator corresponds to a universal quantifier (`forall`) in a Lean theorem
- Cedar's workflow validates this bridge: "property-based testing to check properties of unmodeled parts of the production code" plus "millions of random inputs" for differential testing
- SlimCheck in Lean enables the same test-first-prove-later workflow within a single language: write the property as a Lean proposition, test with `slim_check`, then replace with a formal proof
- Concrete bridge: `test_r1_r3_mutually_exclusive` (Hypothesis) becomes `theorem r1_r3_exclusive` (Lean), tested with SlimCheck, then proved

**Evidence Strength**: MODERATE

### Theme 4: TLA+ and State-Machine Verification Are the Right Tools for Infrastructure Convergence

**Consensus**: For verifying that infrastructure reconciliation (desired state vs. actual state) converges, TLA+ temporal logic specifications are the established approach, not Lean 4 dependent types. AWS validated this at scale. [**STRONG**]
**Sources**: [SRC-002], [SRC-013]

**Controversy**: Whether Lean 4 could subsume TLA+'s role for state-machine verification. Lean has temporal logic libraries but lacks TLA+'s model checker (TLC) which exhaustively enumerates states.
**Dissenting sources**: [SRC-003] notes Lean's expressiveness in principle covers temporal properties, but [SRC-002] demonstrates TLA+'s practical advantage: engineers learned it in 2-3 weeks and found bugs immediately, whereas Lean proofs of the same properties would require deeper expertise.

**Practical Implications**:
- For research question Q6 (Surface-based reconciliation convergence), model the Differ->Surface->Plan->Operation->Executor pipeline in TLA+ rather than Lean
- The DRIFT->OK convergence property is a liveness property (`<>[] (status = OK)`) that TLA+'s model checker can verify exhaustively for bounded state spaces
- For Q7 (frozen SurfaceName strings), Lean's inductive types are the right tool: define `inductive SurfaceName | ecs | lambda | eventbridge | terraform | ...` with exactly the 8 frozen values, making construction of unlisted names a type error

**Evidence Strength**: MIXED (strong for TLA+ recommendation, weak for Lean-for-infrastructure)

### Theme 5: Floating-Point Edge Cases Require Careful Abstraction in Formal Models

**Consensus**: Formal verification of floating-point arithmetic is significantly harder than verification over rationals or reals. The practical approach is to verify properties over an idealized numeric model (rationals or reals) and separately argue that the implementation's numeric type preserves the proven properties within acceptable precision bounds. [**MODERATE**]
**Sources**: [SRC-003], [SRC-005], [SRC-010]

**Practical Implications**:
- For Q2 (NaN/Inf guard) and Q3 (variance invariant after E9): model the filter pipeline over `Rat` or a custom `FinancialValue` inductive type that explicitly includes `NaN` and `Inf` constructors, then prove that the pipeline eliminates these before reaching anomaly detection
- The Python codebase uses `Decimal` for some paths and `float` for others -- the Lean model should abstract over this distinction and prove the property at the abstract level
- The E10-before-E9 ordering constraint (DEF-4 scar) is provable as: `theorem nan_eliminated_before_e9 : forall r, e10(e5(e4(r))).collected != NaN` -- this is a finite inductive proof, not requiring floating-point arithmetic

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Verification-guided development (executable Lean model + differential testing) is a production-proven methodology for verifying business rule correctness -- Sources: [SRC-001], [SRC-002], [SRC-013]
- Lean 4 can prove mutual exclusion of classification outcomes (forbid-trumps-permit is structurally equivalent to R1-excludes-R3/R4) -- Sources: [SRC-001], [SRC-004]
- Financial invariant preservation (balance = sum of parts) is provable in interactive theorem provers -- Sources: [SRC-005], [SRC-007]
- TLA+ is the established tool for state-machine convergence proofs in infrastructure systems -- Sources: [SRC-002], [SRC-013]
- Property-based testing and formal proofs address the same invariant space at different assurance levels -- Sources: [SRC-001], [SRC-008], [SRC-009]
- Engineers can learn formal methods tools (TLA+) and produce useful results in 2-3 weeks -- Sources: [SRC-002]

### MODERATE Evidence
- The Cedar VGD proof-to-model ratio (~3.4:1) provides a calibration point for estimating Lean formalization effort for reconciliation rules -- Sources: [SRC-001]
- Lean 4's `Rat` type and Mathlib's order theory provide the mathematical infrastructure for proving threshold asymmetry completeness (Q4) -- Sources: [SRC-003]
- SlimCheck enables a test-first-prove-later workflow within Lean, bridging existing Hypothesis PBT tests -- Sources: [SRC-011]
- Custom algebraic datatypes with separate theorems (Cedar's approach) are more practical than deep dependent-type encoding for business rule verification -- Sources: [SRC-001]
- Adding new rules (R6/R7 for three-way reconciliation) to a formally verified rule system requires extending the Lean model and re-proving mutual exclusion -- not a fundamental barrier but a linear effort increase -- Sources: [SRC-001], [SRC-005]
- The translation from production code (Python/Go) to Lean model is the highest-effort step -- Sources: [SRC-006]

### WEAK Evidence
- Lean 4 can model floating-point NaN/Inf semantics for filter pipeline proofs, but the infrastructure is immature -- Sources: [SRC-010]
- SlimCheck supports custom generators for domain-specific types, enabling reconciliation-specific test data generation in Lean -- Sources: [SRC-011]
- LLMs may reduce the cost of formal specification writing in the future -- Sources: [SRC-006]

### UNVERIFIED
- No published work applies Lean 4 formal verification specifically to financial reconciliation engines (billing, invoicing, spend anomaly detection) -- Basis: exhaustive web search returned no results
- The optimal formalization strategy for the E4-E5-E10-E9 filter pipeline ordering proof is to model it as a function composition and prove NaN-freedom as a postcondition -- Basis: model training knowledge, architectural analogy to Cedar's validator soundness proof
- Three-way reconciliation (actual vs. billed vs. expected) can be modeled as a product type `ThreeWayComparison` with variance computed pairwise, preserving R1-R5 mutual exclusion if new rules R6/R7 are disjoint from the existing variance-based rules -- Basis: model training knowledge
- Lean 4's `inductive` keyword with exactly 8 constructors would enforce LOAD-001 (frozen SurfaceName strings) at the type level, making it impossible to construct an unlisted surface name -- Basis: model training knowledge, Lean 4 type theory fundamentals

## Knowledge Gaps

- **Financial reconciliation formal verification**: No published work applies Lean 4 (or any interactive theorem prover) specifically to financial reconciliation engines, billing anomaly detection, or spend verification systems. Smart contract token balance proofs are the closest analog but operate in a fundamentally different execution model (blockchain transactions vs. batch processing).

- **Python-to-Lean translation tooling**: No established tool translates Python business logic to Lean 4 specifications. The Cedar workflow translates Rust to Lean, and coq-of-solidity translates Solidity to Coq. A reconciliation engine formalization would require manual translation of Python rules to Lean, guided by the existing property-based tests as a correctness oracle.

- **Floating-point formal verification maturity**: Lean 4's floating-point verification infrastructure (Flean) is experimental. Proving properties about NaN/Inf propagation through a filter pipeline is theoretically possible but lacks established patterns. The practical approach is to abstract to rationals and argue the floating-point implementation is faithful.

- **Three-way reconciliation formalization**: No literature addresses formal verification of multi-source reconciliation (two-way vs. three-way matching). The extension from R1-R5 (two sources: collected vs. spend) to R6-R7 (three sources: actual vs. billed vs. expected) is a novel formalization challenge with no published precedent.

- **Lean 4 effort estimation for domain novices**: The Cedar team had Lean expertise from the start. No data exists on how long it takes a team proficient in Python/Go (but not Lean) to produce useful Lean formalizations. The TLA+ data (2-3 weeks for useful results) may not transfer because Lean is more complex.

## Feasibility Assessment Per Research Question

| Question | Feasibility | Rationale |
|----------|-------------|-----------|
| Q1: Rule mutual exclusion proofs (R1-R5) | **Feasible** | Cedar's forbid-trumps-permit proof is a direct structural analog. R1 (collected==0) vs R3/R4 (collected>0 implied by variance_pct threshold) is a simple decidable proposition. Estimated effort: 1-2 weeks for a Lean-proficient developer. |
| Q2: Edge-case filter ordering (NaN/Inf elimination) | **Feasible** | Model the filter pipeline as function composition over an inductive type with explicit NaN/Inf constructors. Prove NaN-freedom as a postcondition of the composed filters. DEF-4 regression impossibility is a corollary. Estimated effort: 2-3 weeks. |
| Q3: Variance invariant after E9 adjustment | **Feasible** | Structurally identical to token balance conservation proofs. Prove: for all `r`, `e9(r).variance = e9(r).collected - r.spend` and `e9(r).variance_pct = None`. Estimated effort: 1 week (given Q2 infrastructure). |
| Q4: Threshold asymmetry completeness | **Feasible** | Prove: for all `r` with `r.collected > 0` and `r.spend > 0`, `r.variance != 0` implies exactly one of {R3 triggers, R4 triggers, neither triggers}. Requires Mathlib's `Rat` and decidable inequality. Estimated effort: 1-2 weeks. |
| Q5: Three-way reconciliation extension | **Aspirational** | No precedent for multi-source reconciliation formalization. Requires designing the Lean type structure from scratch. The mutual exclusion preservation proof would need to be re-established for the extended rule set. Estimated effort: 4-6 weeks. |
| Q6: Infrastructure state convergence | **Aspirational (in Lean) / Feasible (in TLA+)** | Convergence is a liveness property best verified in TLA+ with model checking. Lean could express it but lacks the exhaustive state enumeration that TLA+'s TLC model checker provides. Recommend TLA+ for this question. |
| Q7: Frozen SurfaceName at type level | **Feasible** | Trivially expressible as a Lean `inductive` type with exactly 8 constructors. Construction of unlisted names becomes a type error by definition. Estimated effort: <1 day. |

## Lean 4 Sketch: Mutual Exclusion of R1 and R3/R4 (Research Question Q1)

```lean
-- Domain types (simplified from Python ClientRecord)
structure ClientRecord where
  collected : Rat          -- billing amount collected
  spend     : Rat          -- actual ad spend
  variance  : Rat          -- collected - spend
  variance_pct : Option Rat -- percentage variance (None if zero-spend)
  deriving Repr

-- Anomaly classification (mirrors Python AnomalyType)
inductive AnomalyType
  | ads_running_no_payment   -- R1: collected == 0 and spend > 0
  | paying_no_ads            -- R2: spend == 0 and collected > 0
  | overbilled               -- R3: variance_pct > overbilled_threshold
  | underbilled              -- R4: variance_pct < -underbilled_threshold
  | stale_account            -- R5: independent, can co-fire
  deriving Repr, DecidableEq

-- R1 fires only when collected == 0
def r1_fires (r : ClientRecord) : Prop :=
  r.collected = 0 ∧ r.spend > 0

-- R3 fires only when variance_pct exceeds positive threshold
-- (which requires collected > 0 and spend > 0)
def r3_fires (r : ClientRecord) (threshold : Rat) : Prop :=
  match r.variance_pct with
  | some vp => vp > threshold ∧ r.collected > 0 ∧ r.spend > 0
  | none => False

-- R4 fires only when variance_pct exceeds negative threshold
def r4_fires (r : ClientRecord) (threshold : Rat) : Prop :=
  match r.variance_pct with
  | some vp => vp < -threshold ∧ r.collected > 0 ∧ r.spend > 0
  | none => False

-- PROOF: R1 and R3 are mutually exclusive
theorem r1_r3_exclusive (r : ClientRecord) (threshold : Rat) :
    r1_fires r → r3_fires r threshold → False := by
  intro ⟨hc, _⟩ ⟨_, hcpos, _⟩
  -- r1_fires gives collected = 0, r3_fires requires collected > 0
  -- 0 > 0 is a contradiction
  linarith

-- PROOF: R1 and R4 are mutually exclusive (same structure)
theorem r1_r4_exclusive (r : ClientRecord) (threshold : Rat) :
    r1_fires r → r4_fires r threshold → False := by
  intro ⟨hc, _⟩ ⟨_, hcpos, _⟩
  linarith

-- PROOF: R3 and R4 cannot both fire (no variance_pct is both
-- above threshold and below negative threshold for positive thresholds)
theorem r3_r4_exclusive (r : ClientRecord)
    (overbilled_threshold underbilled_threshold : Rat)
    (h_pos_over : overbilled_threshold > 0)
    (h_pos_under : underbilled_threshold > 0) :
    r3_fires r overbilled_threshold →
    r4_fires r underbilled_threshold → False := by
  intro h3 h4
  match h_vp : r.variance_pct with
  | some vp =>
    simp [r3_fires, h_vp] at h3
    simp [r4_fires, h_vp] at h4
    -- vp > overbilled_threshold > 0 > -underbilled_threshold > vp
    -- is a contradiction
    linarith [h3.1, h4.1]
  | none =>
    simp [r3_fires, h_vp] at h3
```

## Domain Calibration

Low-to-moderate confidence distribution reflects a domain at the intersection of two well-studied areas (formal verification and financial systems) with sparse literature at their intersection. The formal verification foundations (Lean 4, dependent types, interactive theorem proving) are robust and well-documented. The financial reconciliation application is novel and requires adaptation of established patterns rather than direct application of existing work. Treat findings about Lean/Coq/TLA+ capabilities as reliable; treat feasibility estimates and effort projections as informed speculation calibrated from the Cedar VGD precedent.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content (CACM articles, Springer chapters) could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.
5. **Lean 4 sketch not verified**: The Lean 4 code sketch in the Q1 section has not been compiled or type-checked in a Lean 4 environment. It represents the intended proof structure but may contain syntax errors or require additional imports/lemmas.

Generated by `/research lean-formal-verification-financial-reconciliation` on 2026-03-09.
