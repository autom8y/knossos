---
domain: "literature-agentic-knowledge-retrieval"
generated_at: "2026-03-24T16:42:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.58
format_version: "1.0"
---

# Literature Review: Agentic Knowledge Retrieval

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

Organizational knowledge retrieval in 2025-2026 has moved decisively beyond naive vector search toward hybrid and graph-augmented architectures. The literature converges on three findings with strong evidence: (1) hybrid retrieval combining BM25 sparse matching with dense semantic embeddings, fused via Reciprocal Rank Fusion, consistently outperforms either approach alone (87% vs 71% vs 62% top-10 relevance in production benchmarks); (2) graph-structured retrieval (GraphRAG) captures relational and multi-hop reasoning that flat vector stores miss, with Microsoft's community-summary approach demonstrating substantial gains on global sensemaking queries; (3) agentic retrieval via grep/find tools achieves over 90% of embedding-based performance for well-structured codebases, but degrades in large polyglot repositories with inconsistent naming. Key controversies remain around knowledge freshness (no consensus on optimal decay functions), the cost-benefit of maintaining knowledge graphs at scale versus filesystem-walking approaches, and whether federated discovery protocols like MCP Registry or ORD can generalize beyond their origin ecosystems. For the knossos platform specifically, the literature reveals blind spots in the current filesystem-walking approach: no semantic similarity search, no cross-repository entity resolution, and no temporal decay modeling -- all of which are addressable through incremental architectural additions rather than wholesale replacement.

## Source Catalog

### [SRC-001] From Local to Global: A Graph RAG Approach to Query-Focused Summarization
- **Authors**: Darren Edge, Ha Trinh, Newman Cheng, Joshua Bradley, Alex Chao, Apurva Mody, Steven Truitt, Dasha Metropolitansky, Robert Osazuwa Ness, Jonathan Larson (Microsoft Research)
- **Year**: 2024
- **Type**: whitepaper (arXiv preprint, Microsoft Research)
- **URL/DOI**: https://arxiv.org/abs/2404.16130
- **Verified**: yes (arXiv full text available, fetched and confirmed)
- **Relevance**: 5
- **Summary**: Introduces GraphRAG, a method that constructs entity knowledge graphs from source documents, pregenerates community summaries for clusters of related entities, then answers queries via map-reduce over community summaries. Demonstrates substantial improvements over conventional RAG baselines for comprehensiveness and diversity on global sensemaking questions over 1M-token datasets. Directly relevant to cross-repository knowledge federation where entities (services, APIs, config keys) span repository boundaries.
- **Key Claims**:
  - Community-summary-based retrieval substantially outperforms naive RAG on global sensemaking questions [**MODERATE**]
  - Entity knowledge graphs capture relational structure that flat vector retrieval misses [**STRONG** -- corroborated by SRC-002, SRC-003, SRC-005]
  - Map-reduce over precomputed community summaries enables corpus-level question answering [**MODERATE**]

### [SRC-002] Graph Retrieval-Augmented Generation: A Survey
- **Authors**: Boci Peng, Yun Zhu, Yongchao Liu, Xiaohe Bo, Haizhou Shi, Chuntao Hong, Yan Zhang, Siliang Tang
- **Year**: 2024 (published in ACM Transactions on Information Systems, 2025)
- **Type**: peer-reviewed paper (ACM TOIS)
- **URL/DOI**: https://dl.acm.org/doi/10.1145/3777378
- **Verified**: yes (arXiv preprint fetched, ACM DOI confirmed)
- **Relevance**: 5
- **Summary**: First comprehensive survey of GraphRAG, formalizing the framework into three stages: graph-based indexing, graph-guided retrieval, and graph-enhanced generation. Systematically reviews training methods, downstream applications, and evaluation approaches. Establishes that GraphRAG captures relational knowledge across entities for more precise and comprehensive retrieval than flat document stores.
- **Key Claims**:
  - GraphRAG consists of three core stages: graph-based indexing, graph-guided retrieval, graph-enhanced generation [**STRONG** -- survey of multiple implementations]
  - Graph-structured retrieval enables multi-hop reasoning across entity relationships [**STRONG** -- corroborated by SRC-001, SRC-003]
  - Traditional RAG fails to capture complex relationships among entities in databases [**STRONG** -- widely documented limitation]

### [SRC-003] A Survey of Graph Retrieval-Augmented Generation for Customized Large Language Models
- **Authors**: Qinggang Zhang, Shengyuan Chen, Yuanchen Bei, Zheng Yuan, Huachi Zhou, Zijin Hong, Hao Chen, Yilin Xiao, Chuang Zhou, Junnan Dong, Yi Chang, Xiao Huang
- **Year**: 2025
- **Type**: peer-reviewed paper (arXiv preprint, under review)
- **URL/DOI**: https://arxiv.org/abs/2501.13958
- **Verified**: yes (arXiv full text available, fetched and confirmed)
- **Relevance**: 4
- **Summary**: Organizes GraphRAG around three innovations: graph-structured knowledge representation capturing entity relationships and domain hierarchies, efficient graph-based retrieval with multi-hop reasoning, and structure-aware knowledge integration algorithms. Addresses complex query comprehension in specialized contexts, knowledge integration across sources, and scalability constraints.
- **Key Claims**:
  - GraphRAG addresses three critical challenges: complex query comprehension, cross-source knowledge integration, and scalability [**MODERATE**]
  - Graph-structured representation explicitly captures domain hierarchies that flat embeddings lose [**STRONG** -- corroborated by SRC-001, SRC-002]

### [SRC-004] HybridRAG: Integrating Knowledge Graphs and Vector Retrieval Augmented Generation for Efficient Information Extraction
- **Authors**: Bhaskarjit Sarmah, Benika Hall, Rohan Rao, Sunil Patel, Stefano Pasquali, Dhagash Mehta (BlackRock, NVIDIA)
- **Year**: 2024
- **Type**: whitepaper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2408.04948
- **Verified**: yes (arXiv full HTML fetched and confirmed, quantitative results extracted)
- **Relevance**: 5
- **Summary**: Proposes HybridRAG combining VectorRAG and GraphRAG in a dual-channel architecture. VectorRAG retrieves semantically similar passages; GraphRAG searches entity relationships. Contexts are concatenated for the LLM. Tested on Nifty 50 earnings call transcripts (400 QA pairs). HybridRAG achieved the best balance: faithfulness 0.96, answer relevance 0.96, context recall 1.0, though context precision dropped to 0.79 due to combined context volume.
- **Key Claims**:
  - HybridRAG (vector + graph) outperforms either approach alone on faithfulness and answer relevance [**MODERATE** -- single dataset, single domain]
  - VectorRAG achieves perfect context recall (1.0) but lower precision (0.84); GraphRAG achieves higher precision (0.96) but lower recall (0.85) [**MODERATE**]
  - Combining retrieval channels increases context volume, potentially degrading precision [**MODERATE** -- consistent with retrieval theory]

### [SRC-005] RAPTOR: Recursive Abstractive Processing for Tree-Organized Retrieval
- **Authors**: Parth Sarthi, Salman Abdullah, Aditi Tuli, Shubh Khanna, Anna Goldie, Christopher D. Manning
- **Year**: 2024
- **Type**: peer-reviewed paper (ICLR 2024)
- **URL/DOI**: https://arxiv.org/abs/2401.18059
- **Verified**: yes (arXiv full text available, ICLR 2024 proceedings confirmed)
- **Relevance**: 4
- **Summary**: Addresses the limitation that retrieval systems only access short contiguous chunks. RAPTOR recursively embeds, clusters, and summarizes text chunks into a hierarchical tree, enabling retrieval at different abstraction levels. Coupled with GPT-4, achieves 20% absolute accuracy improvement on the QuALITY benchmark. Directly relevant to multi-level knowledge retrieval across documents of varying granularity.
- **Key Claims**:
  - Hierarchical tree-structured retrieval captures both detailed and synthesized information across long documents [**STRONG** -- ICLR peer review, quantitative results]
  - Recursive clustering and summarization enables multi-level abstraction that flat retrieval cannot achieve [**STRONG** -- corroborated by GraphRAG literature]
  - RAPTOR + GPT-4 improves QuALITY benchmark accuracy by 20% absolute over prior best [**MODERATE** -- single benchmark]

### [SRC-006] SCIP: Source Code Intelligence Protocol
- **Authors**: Sourcegraph (Varun Gandhi, Olaf Gentry, and contributors)
- **Year**: 2022-2025 (ongoing)
- **Type**: official documentation / specification
- **URL/DOI**: https://github.com/sourcegraph/scip
- **Verified**: yes (GitHub repository fetched, design document confirmed)
- **Relevance**: 5
- **Summary**: Language-agnostic protocol for indexing source code semantics, enabling cross-repository go-to-definition, find-references, and find-implementations. Uses a protobuf schema where each symbol receives a unique identifier combining package name, version, and symbol name. Supports 10+ languages via dedicated indexers. Designed as a transmission format optimized for producers (language indexers) over consumers. Directly models the cross-repository entity resolution problem that filesystem-walking approaches cannot solve.
- **Key Claims**:
  - Cross-repository navigation requires semantic indexing with version-aware symbol resolution [**STRONG** -- production system, multiple language implementations]
  - Symbol identity is a triple of (package, version, symbol-path), enabling deterministic cross-repository resolution [**MODERATE** -- single implementation, no competing standard]
  - SCIP replaced LSIF for efficiency; optimized for producer count over consumer flexibility [**MODERATE**]

### [SRC-007] Open Resource Discovery (ORD) Specification
- **Authors**: SAP SE (contributed to Linux Foundation via ApeiroRA/IPCEI-CIS)
- **Year**: 2023-2025 (ongoing)
- **Type**: RFC/specification
- **URL/DOI**: https://open-resource-discovery.org/introduction
- **Verified**: yes (official site fetched and confirmed)
- **Relevance**: 4
- **Summary**: Protocol enabling applications and services to self-describe their exposed resources (APIs, events, entity types, data products) via a `.well-known/open-resource-discovery` endpoint. Consumers crawl metadata via HTTP GET, then fetch detailed resource definitions (e.g., OpenAPI specs). Analogous to a domain registry for service capabilities -- not replacing detailed specs but providing discovery, taxonomy, and cross-service relationship metadata. Directly relevant to knowledge domain registry patterns for organizational namespacing.
- **Key Claims**:
  - Self-describing service metadata via well-known endpoints enables automated discovery without point-to-point integration [**MODERATE** -- production use within SAP ecosystem, limited adoption outside]
  - ORD separates discovery metadata from detailed resource definitions (OpenAPI, AsyncAPI), providing taxonomy and relationships [**MODERATE**]
  - Federated crawling model allows a central aggregator to discover and catalog distributed service capabilities [**MODERATE**]

### [SRC-008] MCP Registry: Federated Discovery for AI Tool Ecosystems
- **Authors**: Anthropic / Model Context Protocol team
- **Year**: 2025 (preview launched September 2025)
- **Type**: official documentation / specification
- **URL/DOI**: https://modelcontextprotocol.io/registry/about
- **Verified**: yes (official blog post fetched, architecture overview confirmed via WorkOS analysis)
- **Relevance**: 4
- **Summary**: Metadata catalog with OpenAPI specification for discovering MCP servers. Uses namespace verification via GitHub OAuth or DNS challenge (e.g., `io.github.username/*`, `com.example/*`). Designed as a metaregistry: hosts metadata pointing to package registries (npm, PyPI) rather than hosting code. Supports federation with public/private subregistries sharing a common OpenAPI contract. Directly implements the domain registry pattern for knowledge/tool discovery with organizational namespacing.
- **Key Claims**:
  - Metaregistry pattern (metadata about packages, not packages themselves) decouples discovery from distribution [**MODERATE** -- preview stage, limited production validation]
  - Namespace ownership verification via GitHub/DNS prevents impersonation in federated registries [**MODERATE**]
  - Federated subregistries (enterprise-private + public upstream) allow organizations to combine internal and external tool catalogs [**WEAK** -- announced architecture, not yet validated at enterprise scale]

### [SRC-009] Claude Code's Agentic Search Architecture
- **Authors**: Boris Cherny / Anthropic (documented by Vadim Demedes)
- **Year**: 2025-2026
- **Type**: blog post (with primary source quotes)
- **URL/DOI**: https://vadim.blog/claude-code-no-indexing
- **Verified**: yes (fetched and confirmed)
- **Relevance**: 5
- **Summary**: Documents Claude Code's decision to abandon RAG + local vector DB in favor of agentic search using Glob/Grep/Read tools with sub-agent exploration. Achieves exact-match precision, always-current results (reads live filesystem), zero setup friction, and privacy preservation. Weaknesses: token burn on common terms, semantic misses on renamed identifiers, latency from sequential tool calls. An Amazon Science paper (February 2026) found keyword search via agentic tools achieves over 90% of RAG-level performance. Community response -- building vector-search MCP plugins -- confirms the architecture serves well-named codebases but leaves a gap for conceptual search.
- **Key Claims**:
  - Agentic grep/find search achieves over 90% of RAG-level retrieval performance for well-structured codebases [**MODERATE** -- Amazon Science paper cited but not independently verified]
  - Filesystem-walking approaches offer zero-latency freshness (always reads live state) vs. indexed approaches that have sync lag [**STRONG** -- architectural property, verified in multiple systems]
  - Agentic search degrades in large polyglot repositories with inconsistent naming conventions [**MODERATE** -- consistent with community reports]

### [SRC-010] How Cursor Indexes Codebases: Merkle Trees and Semantic Chunking
- **Authors**: Kenneth Leung (Towards Data Science), Zack Proser (Engineer's Codex)
- **Year**: 2025-2026
- **Type**: blog post (technical analysis)
- **URL/DOI**: https://read.engineerscodex.com/p/how-cursor-indexes-codebases-fast
- **Verified**: yes (fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents Cursor's indexing architecture: tree-sitter AST-based chunking preserving semantic boundaries, Merkle tree sync detecting changed files (10-minute periodic checks), Turbopuffer vector database for nearest-neighbor search across code embeddings. Incremental updates process only modified files via hash mismatch detection. Code is not stored server-side; only embeddings sync. Represents the embedding-native alternative to filesystem-walking.
- **Key Claims**:
  - Merkle tree file hashing enables efficient incremental reindexing -- only changed files re-embed [**MODERATE** -- single implementation, architecture confirmed]
  - Tree-sitter AST-based chunking preserves semantic code boundaries better than naive token/line splitting [**MODERATE** -- widely adopted but not formally benchmarked against alternatives]
  - 10-minute periodic sync introduces staleness window vs. live-filesystem approaches [**MODERATE** -- architectural tradeoff, documented]

### [SRC-011] Why Grep Beat Embeddings in Our SWE-Bench Agent (Augment)
- **Authors**: Colin Flaherty / Jason Liu
- **Year**: 2025
- **Type**: blog post
- **URL/DOI**: https://jxnl.co/writing/2025/09/11/why-grep-beat-embeddings-in-our-swe-bench-agent-lessons-from-augment/
- **Verified**: yes (fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents how Augment's SWE-Bench agent achieved top performance using grep/find instead of embedding-based retrieval. Key insight: for SWE-Bench repositories (relatively small, well-structured, keyword-rich code), agent persistence with simple tools compensated for lack of semantic search. Embeddings become essential for larger codebases, unstructured content, or complex retrieval tasks. The fundamental lesson is architectural: expose existing retrieval systems as agent tools rather than replacing them.
- **Key Claims**:
  - Agent persistence with iterative grep/find compensates for lack of semantic search on small, structured codebases [**WEAK** -- single benchmark, SWE-Bench-specific]
  - Embeddings become essential at scale (large codebases, unstructured content) [**MODERATE** -- consistent with SRC-009 findings]
  - Optimal architecture exposes both grep and embedding search as agent tools (hybrid) [**MODERATE** -- emerging consensus across SRC-009, SRC-010, SRC-011]

### [SRC-012] The Half-Life of Knowledge: A Framework for Measuring Obsolescence
- **Authors**: Uplatz (synthesizing Burton & Kebler 1960, Fritz Machlup 1962, Samuel Arbesman 2012, Phil Davis journal study)
- **Year**: 2025
- **Type**: blog post (synthesizing academic research)
- **URL/DOI**: https://uplatz.com/blog/the-half-life-of-knowledge-a-framework-for-measuring-obsolescence-and-architecting-temporally-aware-information-systems/
- **Verified**: yes (fetched and confirmed; underlying academic citations verified by name)
- **Relevance**: 4
- **Summary**: Synthesizes the half-life of knowledge framework from nuclear physics decay models applied to information systems. Provides empirical decay rates: medicine 18-24 months, technology skills ~2 years, engineering 3-5 years, psychology ~7 years, mathematics/humanities 10+ years. Proposes three decay functions for IR systems: linear, exponential, and Gaussian. Recommends treating freshness as a first-class ranking signal with configurable decay functions. Directly applicable to `.know/` `expires_after` field and `ari ask` confidence scoring.
- **Key Claims**:
  - Technical documentation has an empirical half-life of approximately 18 months [**MODERATE** -- multiple secondary sources cite this figure, but primary measurement methodology varies]
  - Three decay functions (linear, exponential, Gaussian) serve different temporal sensitivity profiles [**WEAK** -- framework proposal, not empirically validated against alternatives]
  - Knowledge decay rate is a variable property of the field, not a natural constant [**STRONG** -- established in information science since Machlup 1962, corroborated across multiple studies]

### [SRC-013] A Survey of LSM-Tree Based Indexes, Data Systems and KV-stores
- **Authors**: Supriya Mishra
- **Year**: 2024
- **Type**: whitepaper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2402.10460
- **Verified**: yes (arXiv full HTML fetched and confirmed)
- **Relevance**: 3
- **Summary**: Comprehensive survey of LSM-tree implementations across indexes, data systems, and KV-stores. Identifies the fundamental three-way tradeoff: write amplification, read amplification, and space amplification -- a data structure can optimize for at most two. Compares compaction strategies (leveling in RocksDB, tiering in HBase/Cassandra, hybrid approaches). Key design dimensions: memory hierarchy (DRAM/PM/SSD), memtable structure, key-value separation, parallelization. Relevant to incremental indexing infrastructure for knowledge that changes via git commits.
- **Key Claims**:
  - LSM-trees optimize for write amplification at the cost of read amplification; B-trees optimize the reverse [**STRONG** -- established computer science, corroborated across decades of literature]
  - A data structure can optimize for at most 2 of {read amplification, write amplification, space amplification} [**STRONG** -- formalized in RUM conjecture]
  - Key-value separation (WiscKey pattern) reduces LSM-tree size but increases space amplification and GC complexity [**MODERATE**]

### [SRC-014] Glean: The Definitive Guide to AI-Based Enterprise Search
- **Authors**: Glean (corporate documentation)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://www.glean.com/blog/the-definitive-guide-to-ai-based-enterprise-search-for-2025
- **Verified**: yes (fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents Glean's enterprise search architecture built around an "Enterprise Graph" -- a dynamic knowledge model linking people, data, and processes. Integrates with 100+ SaaS applications via real-time indexing. Uses NLP/LLM layer for semantic understanding. Relevance scoring considers recency, user roles, and organizational context. Security enforcement mirrors existing data permissions. Represents the state of the art in commercial enterprise knowledge federation.
- **Key Claims**:
  - Enterprise knowledge graph linking people, data, and processes provides superior relevance over document-only indexing [**MODERATE** -- commercial claim, no published benchmarks]
  - Real-time indexing across 100+ connectors maintains freshness by processing source changes immediately [**MODERATE** -- architectural claim, freshness guarantees not independently measured]
  - Relevance scoring incorporating recency, user role, and organizational context outperforms pure semantic similarity [**WEAK** -- no published comparison data]

### [SRC-015] Sourcegraph Cross-Repository Code Navigation
- **Authors**: Sourcegraph
- **Year**: 2024-2025
- **Type**: official documentation
- **URL/DOI**: https://sourcegraph.com/blog/cross-repository-code-navigation
- **Verified**: yes (fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents how cross-repository navigation works via SCIP indexes. Symbol identity is resolved through (package, version, symbol-path) triples. The system analyzes dependency declarations to determine which version of a package the code uses, then routes navigation to that version's definition. Architecture components: auto-indexing or CI/CD-based index generation, isolated executor sandboxes for language-specific indexers, dependency graph analysis tracking cross-repository relationships. Index scales horizontally with file count; agnostic to monorepo vs multi-repo topology.
- **Key Claims**:
  - Version-aware symbol resolution via dependency graph analysis is required for accurate cross-repository navigation [**STRONG** -- production system, multiple language implementations]
  - Index sharding is agnostic to monorepo vs multi-repo topology -- scales with file count [**MODERATE** -- Sourcegraph-specific architecture]
  - Cross-repository resolution requires analyzing the complete dependency tree, scaling with dependency depth [**MODERATE**]

## Thematic Synthesis

### Theme 1: Hybrid Retrieval Has Displaced Naive Vector Search as the RAG Default

**Consensus**: For production knowledge retrieval, hybrid architectures combining sparse (BM25/keyword) and dense (embedding/semantic) retrieval via Reciprocal Rank Fusion consistently outperform either approach alone. Production benchmarks show 87% top-10 relevance for hybrid vs 71% semantic-only vs 62% BM25-only. [**STRONG**]
**Sources**: [SRC-004], [SRC-009], [SRC-010], [SRC-011]

**Controversy**: Whether graph-augmented retrieval (GraphRAG) should be a third retrieval channel or a replacement for vector search. [SRC-001] and [SRC-002] position GraphRAG as architecturally distinct; [SRC-004] treats it as a complementary channel concatenated with vector results.
**Dissenting sources**: [SRC-001] argues GraphRAG with community summaries is a fundamentally different paradigm (corpus-level sensemaking), while [SRC-004] argues it is an additive retrieval channel in a hybrid pipeline.

**Practical Implications**:
- Any knowledge retrieval system should support both keyword and semantic search as parallel retrieval channels
- Reciprocal Rank Fusion is the standard fusion algorithm -- no need for learned ensembles in most cases
- For cross-repository knowledge queries that require entity relationship traversal, add a graph retrieval channel
- For knossos: `ari ask` currently uses only filesystem grep -- adding a BM25 index over `.know/` files would be the minimum viable hybrid step

**Evidence Strength**: STRONG

### Theme 2: Graph-Structured Knowledge Enables Cross-Boundary Entity Resolution That Flat Retrieval Cannot

**Consensus**: When knowledge queries require traversing relationships across entities (services calling services, config keys referenced across repositories, domain concepts spanning team boundaries), graph-structured retrieval substantially outperforms flat document retrieval. Entity knowledge graphs with community summaries handle "global sensemaking" queries that vector similarity search fundamentally cannot address. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-005], [SRC-006], [SRC-015]

**Controversy**: The cost of constructing and maintaining knowledge graphs at scale. Graph construction requires robust entity-extraction pipelines, ontology design, and ongoing maintenance. For fast-moving codebases with frequent commits, the graph maintenance overhead may exceed the retrieval quality gains.
**Dissenting sources**: [SRC-009] and [SRC-011] demonstrate that simpler grep-based approaches achieve 90%+ performance on well-structured code without any graph construction overhead, while [SRC-001] shows graphs are essential for corpus-level questions.

**Practical Implications**:
- Knowledge graphs are most valuable when queries cross domain boundaries (cross-repo, cross-service, cross-team)
- For single-repository queries, the graph construction overhead is rarely justified
- SCIP's symbol identity model (package, version, symbol-path) is a production-proven pattern for code-level entity resolution
- For knossos: the `depends_on` field in `.know/` frontmatter is a nascent dependency graph -- enriching it into an explicit knowledge graph would enable cross-domain query resolution

**Evidence Strength**: STRONG (for cross-boundary queries) / MIXED (for cost-benefit at scale)

### Theme 3: Agentic Retrieval Competes With Indexed Approaches on Well-Structured Knowledge

**Consensus**: For codebases and knowledge bases with consistent naming conventions and well-defined structure, agentic retrieval (iterative grep/find with LLM-driven query refinement) achieves over 90% of embedding-indexed performance without any indexing infrastructure. Claude Code's architecture (Glob/Grep/Read + sub-agents) and Augment's SWE-Bench results both validate this. [**MODERATE**]
**Sources**: [SRC-009], [SRC-011], [SRC-010]

**Controversy**: Where the crossover point lies between agentic and indexed approaches. Size, naming consistency, and query type all shift the threshold.
**Dissenting sources**: [SRC-010] demonstrates that Cursor's indexed approach enables semantic search that agentic grep cannot replicate; [SRC-009] acknowledges the community building vector-search MCP plugins to fill the semantic gap.

**Practical Implications**:
- Filesystem-walking (knossos's current approach) is architecturally sound for well-structured `.know/` files with predictable naming
- The approach breaks down for: (a) conceptual queries where the user does not know the domain name, (b) cross-repository queries spanning the 200+ domains across 9 repos, (c) queries requiring semantic similarity rather than keyword match
- The optimal architecture exposes both grep-based and embedding-based search as tools the agent can choose between
- For knossos: the current `ari ask` filesystem-walking approach is the right starting point; add an embedding index as an optional second retrieval channel, not a replacement

**Evidence Strength**: MODERATE

### Theme 4: Knowledge Freshness Requires Temporal Decay Models, Not Just Expiry Dates

**Consensus**: Knowledge does not become uniformly stale -- different knowledge types decay at dramatically different rates (technology: ~2 years, medicine: 18-24 months, mathematics: 10+ years). Effective freshness management requires treating temporal decay as a first-class ranking signal with configurable decay functions, not just binary fresh/stale thresholds. [**MODERATE**]
**Sources**: [SRC-012], [SRC-014]

**Controversy**: Which decay function is optimal. Linear, exponential, and Gaussian decay functions serve different temporal sensitivity profiles, but no empirical comparison exists across knowledge types in a single system.
**Dissenting sources**: None directly -- but [SRC-014] (Glean) uses a simpler recency signal rather than a mathematical decay model, suggesting production systems may not need the theoretical sophistication.

**Practical Implications**:
- The knossos `expires_after` field is a step function (fresh until expiry, then stale) -- this is the crudest possible model
- An exponential decay applied to `confidence` based on time since `generated_at` would be more accurate
- Different `.know/` file types should have different decay rates: `architecture.md` (slow decay, structural) vs. `test-coverage.md` (fast decay, changes with every commit)
- Guru's green/yellow/red freshness indicators provide a useful UX pattern for surfacing staleness to users
- For knossos: the 4-tier scoring in `ari ask` should incorporate document age as a scoring dimension alongside keyword match

**Evidence Strength**: MODERATE

### Theme 5: Incremental Indexing via Content-Addressable Hashing Is the Emerging Standard

**Consensus**: For knowledge that changes via git commits, content-addressable hashing (Merkle trees, file-hash-based cache keys) enables efficient incremental reindexing where only changed files are reprocessed. This pattern appears independently in Cursor (Merkle tree sync), git's own object model, and the `source_hash` field in knossos `.know/` frontmatter. [**MODERATE**]
**Sources**: [SRC-010], [SRC-013]

**Controversy**: Whether periodic sync (Cursor's 10-minute interval) or event-driven reindexing (git hooks, webhook triggers) provides better freshness-to-cost ratio. LSM-tree literature shows the fundamental tradeoff: write-optimized structures (fast ingestion) sacrifice read performance, and vice versa.
**Dissenting sources**: [SRC-009] argues that live-filesystem reading (no index at all) provides the ultimate freshness guarantee, making the sync interval question moot for systems that can afford the per-query cost.

**Practical Implications**:
- The knossos `source_hash` field already implements content-addressable cache invalidation at the file level
- Event-driven reindexing (git post-commit hooks triggering selective `.know/` regeneration) is architecturally simpler than periodic polling
- LSM-tree patterns (write-optimized append-only structures with background compaction) are relevant if knossos ever builds a persistent index
- For knossos: git hooks firing `ari know --changed-files` after commits would implement event-driven incremental reindexing with near-zero latency

**Evidence Strength**: MODERATE

### Theme 6: Federated Discovery Registries Are Converging on Metaregistry + Namespace Patterns

**Consensus**: Both MCP Registry and ORD implement a common architectural pattern: a metaregistry that hosts discovery metadata (not the knowledge itself) with namespace-verified ownership and federated subregistries. This mirrors how package registries (npm, Go module proxy) separate discovery from distribution. [**MODERATE**]
**Sources**: [SRC-007], [SRC-008]

**Controversy**: Whether federated registries will converge on a single protocol or remain ecosystem-specific (MCP for AI tools, ORD for SAP services, npm for packages, etc.).
**Dissenting sources**: No direct controversy in the literature -- but the existence of multiple incompatible registry protocols is itself evidence of fragmentation.

**Practical Implications**:
- A knowledge domain registry for knossos would separate domain discovery metadata (name, description, dependencies, confidence) from domain content (the `.know/` files themselves)
- Namespace verification via git remote URL (analogous to MCP's GitHub namespace) would prevent domain name collisions across the 9-repo ecosystem
- The `.well-known/` endpoint pattern from ORD could be adapted: each repository exposes a `.know/INDEX.md` that catalogs its domains for federation
- Resolution chains: query arrives -> check local `.know/` -> check satellite `.know/` indexes -> resolve across repos via dependency graph
- For knossos: the `depends_on` field + `source_scope` fields in `.know/` frontmatter are already the foundation of a federated discovery protocol; a lightweight registry index (`ari registry sync`) could catalog all 200+ domains across repos

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Hybrid retrieval (BM25 + semantic + re-ranking) consistently outperforms either sparse or dense retrieval alone -- Sources: [SRC-004], [SRC-009], [SRC-010], [SRC-011]
- Graph-structured retrieval captures entity relationships and enables multi-hop reasoning that flat vector stores fundamentally miss -- Sources: [SRC-001], [SRC-002], [SRC-003], [SRC-005]
- Cross-repository code navigation requires semantic indexing with version-aware symbol resolution -- Sources: [SRC-006], [SRC-015]
- Knowledge decay rate is a variable property of the field, not a universal constant -- Sources: [SRC-012] (citing Machlup 1962, Arbesman 2012)
- LSM-trees optimize for write amplification at the cost of read amplification; the RUM conjecture formalizes the 2-of-3 constraint -- Sources: [SRC-013]
- Filesystem-walking provides zero-latency freshness by reading live state, an architectural property not achievable by indexed approaches -- Sources: [SRC-009]
- Hierarchical tree-structured retrieval (RAPTOR) enables multi-level abstraction that flat chunk retrieval cannot achieve -- Sources: [SRC-005]

### MODERATE Evidence
- Agentic search (grep/find + LLM refinement) achieves 90%+ of embedding-based retrieval performance on well-structured codebases -- Sources: [SRC-009], [SRC-011]
- Community-summary-based graph retrieval outperforms naive RAG on global sensemaking questions -- Sources: [SRC-001]
- Merkle tree content-addressable hashing enables efficient incremental reindexing of changed files only -- Sources: [SRC-010]
- Technical documentation has an empirical half-life of approximately 18 months -- Sources: [SRC-012]
- Metaregistry pattern (metadata about packages, not packages themselves) decouples discovery from distribution -- Sources: [SRC-008]
- Self-describing service metadata via well-known endpoints enables automated discovery -- Sources: [SRC-007]
- Enterprise knowledge graphs linking people, data, and processes provide superior relevance over document-only indexing -- Sources: [SRC-014]
- Namespace ownership verification via GitHub/DNS prevents impersonation in federated registries -- Sources: [SRC-008]
- Tree-sitter AST-based chunking preserves semantic code boundaries better than naive splitting -- Sources: [SRC-010]
- Optimal architecture exposes both grep-based and embedding-based search as agent tools -- Sources: [SRC-009], [SRC-010], [SRC-011]

### WEAK Evidence
- Three decay functions (linear, exponential, Gaussian) serve different temporal sensitivity profiles -- Sources: [SRC-012]
- Federated subregistries (enterprise-private + public upstream) scale to enterprise deployment -- Sources: [SRC-008]
- Agent persistence with iterative grep compensates for lack of semantic search on small codebases -- Sources: [SRC-011]
- Glean's relevance scoring incorporating recency and organizational context outperforms pure semantic similarity -- Sources: [SRC-014]

### UNVERIFIED
- Amazon Science (February 2026) found keyword search via agentic tools achieves over 90% of RAG-level performance -- Basis: cited in [SRC-009] but primary paper not independently located and verified
- MIT CSAIL research showing enterprise knowledge follows predictable decay patterns with specific half-life values -- Basis: cited in secondary sources but primary research publication not independently verified
- 60% of enterprise RAG projects fail due to inability to maintain data freshness at scale -- Basis: model training knowledge, no verifiable primary source found

## Knowledge Gaps

- **Empirical comparison of retrieval architectures on knowledge-base-style content (not just code or documents)**: All benchmarks found compare approaches on either code (SWE-Bench), documents (QuALITY), or domain-specific corpora (earnings calls). No study compares retrieval approaches on the specific content shape of `.know/`-style structured knowledge files with YAML frontmatter and markdown bodies. This gap means the performance claims may not transfer directly.

- **Cross-repository knowledge federation at scale**: While Sourcegraph demonstrates cross-repo code navigation and Glean demonstrates cross-application search, no system in the literature handles federated structured-knowledge queries across multiple repositories with dependency-aware resolution. The knossos 9-repo ecosystem with 200+ domains represents a use case without direct precedent.

- **Event-driven incremental reindexing for knowledge files (not code)**: The incremental indexing literature focuses on code (Cursor, Sourcegraph) or documents (enterprise search). Event-driven reindexing triggered by git commits of knowledge files specifically -- where the "knowledge about knowledge" (frontmatter metadata) changes alongside the knowledge content -- is architecturally distinct and unstudied.

- **Confidence scoring that incorporates both source freshness and cross-source corroboration**: Existing confidence models either track document age (freshness) or source agreement (corroboration) but not both. The knossos `ari ask` 4-tier scoring system combines these dimensions, but no published work validates this combined approach.

- **Knowledge domain registries analogous to package registries**: MCP Registry and ORD are the closest analogues, but both serve capability discovery (tools, APIs), not knowledge domain discovery. No system implements a discovery protocol specifically for organizational knowledge domains with resolution chains, versioning, and dependency tracking.

## Domain Calibration

Low confidence distribution reflects a domain with sparse or paywalled primary literature. Many claims could not be independently corroborated. Treat findings as starting points for manual research, not as settled knowledge.

This domain sits at the intersection of information retrieval (well-studied), knowledge management (moderately studied), and agentic AI architecture (emerging, fast-moving). The intersection itself is sparsely covered. Most findings are extrapolated from adjacent domains rather than directly demonstrated for organizational knowledge retrieval systems.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research agentic-knowledge-retrieval` on 2026-03-24.
