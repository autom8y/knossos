---
domain: "literature-graphrag-hierarchical-retrieval"
generated_at: "2026-03-25T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.65
format_version: "1.0"
---

# Literature Review: GraphRAG, Hierarchical Retrieval, and Knowledge Graph Construction for Structured Engineering Knowledge

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

Graph-based Retrieval-Augmented Generation (GraphRAG) has emerged as a significant evolution of traditional RAG, addressing limitations in multi-hop reasoning, global sensemaking, and relational knowledge retrieval. The literature reveals strong consensus that graph-structured knowledge representations outperform flat vector retrieval for complex reasoning tasks, but substantial controversy exists over whether the indexing cost of full GraphRAG is justified for all use cases. Microsoft's original GraphRAG (2024) established the paradigm of LLM-driven entity extraction, Leiden community detection, and hierarchical community summarization. Subsequent work -- LazyGraphRAG, LightRAG, HyperGraphRAG, RAPTOR -- each addresses specific cost-quality-expressiveness tradeoffs. For structured engineering documentation with domain-typed frontmatter and hierarchical organization, the evidence strongly favors domain-specific KG schemas over generic extraction, with recent work showing 10%+ entity yield improvements from expert-crafted schemas. The field is rapidly evolving, with 2025 benchmarks (GraphRAG-Bench) demonstrating that GraphRAG's advantage is task-dependent rather than universal.

## Source Catalog

### [SRC-001] From Local to Global: A Graph RAG Approach to Query-Focused Summarization
- **Authors**: Darren Edge, Ha Trinh, Newman Cheng, Joshua Bradley, Alex Chao, Apurva Mody, Steven Truitt, Dasha Metropolitansky, Robert Osazuwa Ness, Jonathan Larson
- **Year**: 2024
- **Type**: peer-reviewed paper (arXiv preprint, Microsoft Research)
- **URL/DOI**: https://arxiv.org/abs/2404.16130
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: The foundational Microsoft GraphRAG paper. Introduces the two-stage pipeline: (1) LLM-driven entity knowledge graph extraction from source documents, (2) hierarchical community detection via Leiden algorithm to pregenerate community summaries. At query time, community summaries produce partial responses that are aggregated into final answers. Demonstrated substantial improvements over vector RAG baseline for global sensemaking questions on million-token corpora, particularly in comprehensiveness and diversity of answers.
- **Key Claims**:
  - LLM-driven entity extraction combined with Leiden community detection produces hierarchical knowledge graphs that enable multi-level abstraction for retrieval [**STRONG**]
  - GraphRAG substantially outperforms vector RAG for global sensemaking queries requiring whole-corpus reasoning [**MODERATE**]
  - Community-level summarization enables query-focused summarization at variable granularity [**MODERATE**]
  - The approach scales to million-token corpora [**MODERATE**]

### [SRC-002] Graph Retrieval-Augmented Generation: A Survey
- **Authors**: Boci Peng, Yun Zhu, Yongchao Liu, Xiaohe Bo, Haizhou Shi, Chuntao Hong, Yan Zhang, Siliang Tang
- **Year**: 2024
- **Type**: peer-reviewed paper (ACM Transactions on Information Systems)
- **URL/DOI**: https://arxiv.org/abs/2408.08921
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: The first comprehensive survey formalizing the GraphRAG workflow into three stages: Graph-Based Indexing, Graph-Guided Retrieval, and Graph-Enhanced Generation. Provides taxonomy of methods, downstream applications, evaluation methodologies, and industrial implementations. Published in ACM TOIS, establishing the field's taxonomic vocabulary.
- **Key Claims**:
  - GraphRAG captures relational knowledge that flat vector retrieval misses, enabling more accurate context-aware responses [**STRONG**]
  - The GraphRAG workflow decomposes into three distinct stages (indexing, retrieval, generation), each with its own technique landscape [**STRONG**]
  - Graph-based indexing provides structural relationships among entities that are absent in traditional chunk-based approaches [**MODERATE**]

### [SRC-003] Retrieval-Augmented Generation with Graphs (GraphRAG)
- **Authors**: Haoyu Han, Yu Wang, Harry Shomer, Kai Guo, Jiayuan Ding, et al. (18 authors)
- **Year**: 2025
- **Type**: peer-reviewed paper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2501.00309
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: A complementary survey to SRC-002 that defines a five-component GraphRAG framework: query processor, retriever, organizer, generator, and data source. Takes a domain-aware approach, reviewing techniques tailored to different graph types and domains. Emphasizes that different graph structures require dedicated retrieval and generation designs.
- **Key Claims**:
  - Different graph types (knowledge graphs, document graphs, domain-specific graphs) require dedicated GraphRAG designs [**MODERATE**]
  - GraphRAG addresses three critical RAG limitations: complex query understanding, distributed knowledge integration, and system efficiency at scale [**MODERATE**]

### [SRC-004] LazyGraphRAG: Setting a New Standard for Quality and Cost
- **Authors**: Darren Edge, Ha Trinh, Jonathan Larson (Microsoft Research)
- **Year**: 2024
- **Type**: whitepaper (Microsoft Research blog with technical detail)
- **URL/DOI**: https://www.microsoft.com/en-us/research/blog/lazygraphrag-setting-a-new-standard-for-quality-and-cost/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Introduces LazyGraphRAG, which defers LLM usage from indexing time to query time. Uses NLP noun-phrase extraction (not LLM) for concept co-occurrence graphs, then applies iterative deepening with a tunable "relevance test budget" parameter at query time. Achieves indexing costs identical to vector RAG (0.1% of full GraphRAG), while matching GraphRAG Global Search quality at 700x lower query cost. At 4% of GraphRAG query cost, significantly outperforms all baselines on both local and global queries.
- **Key Claims**:
  - LazyGraphRAG indexing costs are 0.1% of full GraphRAG while being identical to vector RAG indexing costs [**MODERATE**]
  - A single "relevance test budget" parameter provides continuous cost-quality tradeoff control [**MODERATE**]
  - At budget=500, LazyGraphRAG matches GraphRAG Global Search quality at >700x lower query cost [**MODERATE**]
  - Combining best-first (similarity) and breadth-first (community coverage) search strategies is more effective than either alone [**WEAK**]
  - NLP noun-phrase extraction can substitute for LLM entity extraction during indexing without quality loss [**WEAK**]

### [SRC-005] RAPTOR: Recursive Abstractive Processing for Tree-Organized Retrieval
- **Authors**: Parth Sarthi, Salman Abdullah, Aditi Tuli, Shubh Khanna, Anna Goldie, Christopher D. Manning
- **Year**: 2024
- **Type**: peer-reviewed paper (ICLR 2024)
- **URL/DOI**: https://arxiv.org/abs/2401.18059
- **Verified**: yes (content fetched, ICLR venue confirmed via proceedings URL)
- **Relevance**: 5
- **Summary**: Introduces recursive tree construction for hierarchical retrieval. Text chunks are embedded, clustered (using GMMs with UMAP dimensionality reduction), and summarized recursively bottom-up, creating a multi-level tree where higher nodes represent increasingly abstract summaries. At inference, retrieval traverses the tree to integrate information at different abstraction levels. Achieved 20% absolute accuracy improvement on QuALITY benchmark when coupled with GPT-4, particularly for complex multi-step reasoning tasks.
- **Key Claims**:
  - Recursive bottom-up clustering and summarization creates multi-level abstraction trees that improve retrieval for complex reasoning tasks [**STRONG**]
  - GMMs with UMAP dimensionality reduction enable soft clustering where text segments can belong to multiple clusters [**MODERATE**]
  - RAPTOR with GPT-4 achieves 20% absolute accuracy improvement on QuALITY benchmark over prior state-of-the-art [**MODERATE**]
  - Hierarchical tree retrieval enables integration of information across different abstraction levels in a single query [**STRONG**]

### [SRC-006] HyperGraphRAG: Retrieval-Augmented Generation via Hypergraph-Structured Knowledge Representation
- **Authors**: Haoran Luo, Haihong E, Guanting Chen, Yandan Zheng, Xiaobao Wu, Yikai Guo, Qika Lin, Yu Feng, Zemin Liu, Meina Song, Yifan Zhu, Luu Anh Tuan
- **Year**: 2025
- **Type**: peer-reviewed paper (NeurIPS 2025 main conference)
- **URL/DOI**: https://arxiv.org/abs/2503.21322
- **Verified**: yes (content fetched, NeurIPS venue noted on arxiv)
- **Relevance**: 4
- **Summary**: Addresses the fundamental limitation of binary edges in standard knowledge graphs by introducing hyperedges that connect 3+ entities in n-ary relational facts. Where traditional GraphRAG can only model pairwise relations (entity-relation-entity triples), HyperGraphRAG represents complex multi-entity relationships natively. Evaluated across medicine, agriculture, computer science, and law domains, outperforming both standard RAG and GraphRAG in accuracy and generation quality.
- **Key Claims**:
  - Standard knowledge graph triples cannot adequately represent n-ary relations among 3+ entities that widely exist in real-world knowledge [**STRONG**]
  - Hyperedge-based knowledge representation improves answer accuracy and retrieval efficiency over binary-edge GraphRAG [**MODERATE**]
  - The approach generalizes across diverse domains (medicine, agriculture, CS, law) [**MODERATE**]

### [SRC-007] LightRAG: Simple and Fast Retrieval-Augmented Generation
- **Authors**: Zirui Guo, Lianghao Xia, Yanhua Yu, Tu Ao, Chao Huang
- **Year**: 2024
- **Type**: peer-reviewed paper (EMNLP 2025 Findings)
- **URL/DOI**: https://arxiv.org/abs/2410.05779
- **Verified**: yes (content fetched, EMNLP venue confirmed via ACL anthology)
- **Relevance**: 4
- **Summary**: Proposes a lightweight graph-based RAG system with dual-level retrieval (low-level entity-specific and high-level abstract queries) and an incremental graph update algorithm that avoids full index rebuilds. Integrates graph structures with vector representations for efficient entity and relationship retrieval. Addresses the practical concern of dynamic environments where knowledge bases change frequently.
- **Key Claims**:
  - Dual-level retrieval (specific entity queries + abstract high-level queries) captures both fine-grained and holistic knowledge [**MODERATE**]
  - Incremental graph updates eliminate the need for full index rebuilds when new data arrives [**MODERATE**]
  - Graph structures integrated with vector representations improve retrieval accuracy over flat vector approaches [**WEAK**]

### [SRC-008] GraphRAG on Technical Documents -- Impact of Knowledge Graph Schema
- **Authors**: Henri Scaffidi, Melinda Hodkiewicz, Caitlin Woods, Nicole Roocke
- **Year**: 2025
- **Type**: peer-reviewed paper (TGDK -- Transactions on Graph Data and Knowledge, Dagstuhl)
- **URL/DOI**: https://drops.dagstuhl.de/entities/document/10.4230/TGDK.3.2.3
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Directly tests the impact of domain-specific knowledge graph schemas on Microsoft GraphRAG's performance on technical reports from the minerals industry. Compared four schema approaches (simple 5-class domain schema, expanded 8-class domain schema, auto-generated schema, and schema-less) against baseline RAG. The simple expert-developed 5-class schema extracted ~10% more entities than other options, and both expert schemas produced the most factually correct answers with fewest hallucinations. Demonstrates that schema quality is a critical lever for GraphRAG effectiveness on domain-specific technical documents.
- **Key Claims**:
  - Domain-expert-crafted KG schemas extract more relevant entities than auto-generated or schema-less approaches (~10% improvement with simple 5-class schema) [**MODERATE**]
  - Expert mineral-specific schemas produce fewer hallucinations and more factually correct answers than generic schemas or baseline RAG [**MODERATE**]
  - Schema quality during indexing directly determines retrieval context quality, which cascades to generation quality [**MODERATE**]
  - Simpler domain schemas (5 classes) can outperform more complex schemas (8 classes) by reducing noise [**WEAK**]

### [SRC-009] When to Use Graphs in RAG: A Comprehensive Analysis for Graph Retrieval-Augmented Generation
- **Authors**: Zhishang Xiang, Chuanjie Wu, Qinggang Zhang, Shengyuan Chen, Zijin Hong, Xiao Huang, Jinsong Su
- **Year**: 2025
- **Type**: peer-reviewed paper (ICLR 2026)
- **URL/DOI**: https://arxiv.org/abs/2506.05690
- **Verified**: yes (content fetched, ICLR venue confirmed via GitHub repo)
- **Relevance**: 4
- **Summary**: Introduces GraphRAG-Bench, a comprehensive benchmark evaluating when graph structures genuinely benefit RAG systems. Tests across four task categories of increasing complexity: fact retrieval, complex reasoning, contextual summarization, and creative generation. Key finding: GraphRAG frequently underperforms vanilla RAG on many real-world tasks, but significantly outperforms on multi-hop reasoning over textual graphs. Provides practical guidelines for when to adopt graph-based approaches.
- **Key Claims**:
  - GraphRAG frequently underperforms vanilla RAG on simple fact retrieval and creative generation tasks [**MODERATE**]
  - GraphRAG significantly outperforms vanilla RAG specifically for multi-hop reasoning tasks on textual graphs [**MODERATE**]
  - The benefit of graph structures is task-dependent, not universal; benchmarking across task types is essential before adopting GraphRAG [**MODERATE**]

### [SRC-010] A Survey of Graph Retrieval-Augmented Generation for Customized Large Language Models
- **Authors**: Qinggang Zhang, Shengyuan Chen, Yuanchen Bei, Zheng Yuan, Huachi Zhou, Zijin Hong, et al.
- **Year**: 2025
- **Type**: peer-reviewed paper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2501.13958
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Survey focusing on GraphRAG for LLM customization in specialized domains. Identifies three core innovations: graph-structured knowledge representation capturing entity relationships and domain hierarchies, graph-based retrieval with multi-hop reasoning, and structure-aware knowledge integration. Emphasizes that different domains require different graph structures and retrieval strategies.
- **Key Claims**:
  - Graph-structured knowledge explicitly captures entity relationships and domain hierarchies that flat text representations lose [**STRONG**]
  - Multi-hop reasoning capability is a primary advantage of graph-based retrieval over vector-based retrieval [**STRONG**]
  - Domain customization of GraphRAG requires domain-specific graph schemas and retrieval strategies [**MODERATE**]

### [SRC-011] Document GraphRAG: Knowledge Graph Enhanced Retrieval Augmented Generation for Document Question Answering Within the Manufacturing Domain
- **Authors**: Simon Knollmeyer, Oguz Caymazer, Daniel Grossmann
- **Year**: 2025
- **Type**: peer-reviewed paper (Electronics journal, MDPI)
- **URL/DOI**: https://www.mdpi.com/2079-9292/14/11/2102
- **Verified**: partial (title and metadata confirmed via search; full text behind publisher access)
- **Relevance**: 4
- **Summary**: Proposes a framework that builds knowledge graphs from a document's intrinsic structure, preserving hierarchical organization and metadata (author, publication year, structural hierarchy). Uses graph-based document structuring and keyword-based semantic linking. Evaluated on SQuAD, HotpotQA, and a manufacturing dataset, showing performance gains over naive RAG, particularly for multi-hop questions that benefit from structured retrieval.
- **Key Claims**:
  - Preserving document intrinsic structure (hierarchies, metadata, sections) in knowledge graphs improves retrieval quality over flat chunking [**MODERATE**]
  - Multi-hop questions benefit most from graph-structured retrieval that preserves document relationships [**MODERATE**]
  - Keyword-based semantic linking combined with structural graph traversal enhances context relevance [**WEAK**]

### [SRC-012] LLM-Empowered Knowledge Graph Construction: A Survey
- **Authors**: Haonan Bian
- **Year**: 2025
- **Type**: peer-reviewed paper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2510.20345
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Surveys the paradigm shift in knowledge graph construction from rule-based/statistical pipelines to LLM-driven generative frameworks. Covers the three-layered pipeline (ontology engineering, knowledge extraction, knowledge fusion) and contrasts schema-based paradigms (structure, normalization, consistency) with schema-free paradigms (flexibility, adaptability, open discovery). Identifies future directions including KG-based reasoning for LLMs, dynamic knowledge memory for agentic systems, and multimodal KG construction.
- **Key Claims**:
  - LLMs have shifted KG construction from rule-based pipelines to language-driven generative frameworks [**MODERATE**]
  - Schema-based KG construction emphasizes structure and consistency; schema-free emphasizes flexibility and open discovery -- the choice is domain-dependent [**MODERATE**]
  - Dynamic knowledge memory for agentic systems is an emerging research direction at the KG-LLM intersection [**WEAK**]

### [SRC-013] DRIFT Search: Combining Global and Local Search Methods to Improve Quality and Efficiency
- **Authors**: Microsoft Research (GraphRAG team)
- **Year**: 2024
- **Type**: official documentation (Microsoft Research blog + open-source implementation)
- **URL/DOI**: https://www.microsoft.com/en-us/research/blog/introducing-drift-search-combining-global-and-local-search-methods-to-improve-quality-and-efficiency/
- **Verified**: yes (content fetched from Microsoft Research blog and GitHub)
- **Relevance**: 3
- **Summary**: DRIFT (Dynamic Reasoning and Inference with Flexible Traversal) extends Microsoft GraphRAG by combining community-level global search with entity-level local search. Begins with community information via vector search to establish broad query context, then decomposes broad questions into fine-grained follow-ups that dynamically traverse the knowledge graph. Bridges the gap between GraphRAG's global search (comprehensive but expensive) and local search (fast but narrow).
- **Key Claims**:
  - Combining community-level (global) and entity-level (local) search retrieves a higher variety of facts than either approach alone [**WEAK**]
  - Dynamic question decomposition into follow-up queries enables adaptive knowledge graph traversal [**WEAK**]

## Thematic Synthesis

### Theme 1: Hierarchical Knowledge Representation is the Consensus Retrieval Architecture for Complex Reasoning

**Consensus**: Multi-level abstraction hierarchies -- whether via community detection (GraphRAG), recursive tree construction (RAPTOR), or hyperedges (HyperGraphRAG) -- consistently outperform flat vector retrieval for tasks requiring multi-hop reasoning or whole-corpus understanding. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-005], [SRC-006], [SRC-009], [SRC-010]

**Controversy**: Whether the hierarchy should be pre-computed at indexing time (GraphRAG, RAPTOR) or constructed dynamically at query time (LazyGraphRAG, DRIFT). Pre-computed hierarchies offer faster queries but incur high indexing costs and cannot adapt to novel query patterns.
**Dissenting sources**: [SRC-001] argues pre-computed community summaries are essential for global sensemaking, while [SRC-004] argues deferred computation achieves equivalent quality at 0.1% of indexing cost.

**Practical Implications**:
- For engineering knowledge bases with stable, well-structured documents, pre-computed hierarchical indexes are justified
- For rapidly evolving corpora or cost-sensitive deployments, LazyGraphRAG's deferred computation is the pragmatic default
- RAPTOR's recursive tree approach is particularly applicable to document collections with inherent hierarchical structure (e.g., documentation with section/subsection nesting)

**Evidence Strength**: STRONG (hierarchy improves complex reasoning) / MIXED (when to pre-compute vs. defer)

### Theme 2: Domain-Specific Schema Design is a Critical and Under-Appreciated Lever

**Consensus**: The quality of the knowledge graph schema used during entity extraction directly determines downstream retrieval and generation quality. Expert-crafted domain schemas consistently outperform auto-generated and generic schemas. [**MODERATE**]
**Sources**: [SRC-008], [SRC-003], [SRC-010], [SRC-011], [SRC-012]

**Controversy**: Whether to use schema-based (rigid, consistent) or schema-free (flexible, open-discovery) approaches for KG construction. The answer appears domain-dependent: well-defined technical domains benefit from schemas; exploratory domains need flexibility.
**Dissenting sources**: [SRC-012] presents schema-free as a legitimate alternative for flexibility, while [SRC-008] demonstrates clear superiority of domain schemas on technical documents.

**Practical Implications**:
- For structured engineering documentation with known entity types (e.g., tools, patterns, conventions), invest in a curated domain schema before deploying GraphRAG
- Start with a simple schema (5-7 entity classes) rather than an exhaustive one; simpler schemas reduce noise
- Schema design is a human-expert task that cannot yet be reliably automated by LLMs for specialized domains
- Document frontmatter and metadata (types, categories, dates) should be preserved as first-class graph entities, not discarded during extraction

**Evidence Strength**: MODERATE

### Theme 3: The Cost-Quality Tradeoff Defines the Practical GraphRAG Design Space

**Consensus**: Full GraphRAG indexing is prohibitively expensive for many use cases. The field is converging on deferred-computation and lightweight alternatives that preserve most quality benefits at dramatically lower cost. [**MODERATE**]
**Sources**: [SRC-001], [SRC-004], [SRC-007], [SRC-009]

**Practical Implications**:
- LazyGraphRAG achieves comparable quality at 0.1% indexing cost and 700x lower query cost -- it is the cost-efficient default for most deployments
- LightRAG's incremental update capability is essential for dynamic corpora; avoid architectures that require full re-indexing
- Budget the relevance test parameter (LazyGraphRAG) or top-k parameter against SLA requirements; both provide continuous cost-quality dials
- Full GraphRAG is justified only when global sensemaking over stable corpora is the primary use case and cost is not a constraint

**Evidence Strength**: MODERATE

### Theme 4: GraphRAG's Advantage is Task-Dependent, Not Universal

**Consensus**: GraphRAG does not universally outperform vector RAG. Its advantages are concentrated in multi-hop reasoning and global sensemaking tasks, while it may underperform on simple fact retrieval and creative generation. [**MODERATE**]
**Sources**: [SRC-009], [SRC-001], [SRC-004], [SRC-005]

**Practical Implications**:
- Benchmark GraphRAG against vector RAG on your specific task distribution before committing to the architecture
- Hybrid approaches (graph + vector) are likely optimal for mixed workloads
- For engineering knowledge bases, multi-hop queries (e.g., "what conventions affect this pattern?") are the primary use case where GraphRAG adds value
- Simple lookup queries ("what does this config option do?") may be better served by vector RAG

**Evidence Strength**: MODERATE

### Theme 5: N-ary Relations and Hypergraph Representations Address Real Expressiveness Gaps

**Consensus**: Standard binary-edge knowledge graphs (subject-predicate-object triples) cannot adequately represent many real-world knowledge structures, particularly in engineering domains where relationships often involve 3+ entities (e.g., "tool X used in context Y for purpose Z"). [**MODERATE**]
**Sources**: [SRC-006], [SRC-002], [SRC-010]

**Practical Implications**:
- If your domain has prevalent multi-entity relationships (e.g., a pattern that applies to a language in a context for a purpose), consider hypergraph representations over standard KGs
- HyperGraphRAG is the current state-of-the-art for n-ary relational knowledge, but the tooling ecosystem is immature compared to binary KG tooling
- For structured documentation with typed frontmatter, hyperedges could naturally represent the multi-dimensional metadata (domain + type + status + dependencies)

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Hierarchical knowledge representations (community hierarchies, recursive trees) consistently improve retrieval for multi-hop and global reasoning tasks -- Sources: [SRC-001], [SRC-002], [SRC-005], [SRC-009], [SRC-010]
- The GraphRAG workflow decomposes into three distinct stages (indexing, retrieval, generation), each with its own design space -- Sources: [SRC-002], [SRC-003]
- Standard binary-edge knowledge graphs cannot adequately represent n-ary relations among 3+ entities -- Sources: [SRC-006], [SRC-002]
- RAPTOR's recursive bottom-up clustering and summarization creates effective multi-level abstraction trees -- Sources: [SRC-005] (ICLR peer review + independent reimplementations)
- Multi-hop reasoning capability is a primary advantage of graph-based over vector-based retrieval -- Sources: [SRC-010], [SRC-009]
- Graph-structured knowledge captures relational context that flat vector retrieval misses -- Sources: [SRC-002], [SRC-010]

### MODERATE Evidence
- GraphRAG substantially outperforms vector RAG for global sensemaking queries -- Sources: [SRC-001]
- LazyGraphRAG achieves 0.1% of GraphRAG indexing costs with comparable quality -- Sources: [SRC-004]
- Domain-expert-crafted KG schemas outperform auto-generated schemas for technical documents (~10% more entities, fewer hallucinations) -- Sources: [SRC-008]
- Preserving document intrinsic structure in knowledge graphs improves retrieval over flat chunking -- Sources: [SRC-011]
- LLMs have shifted KG construction from rule-based to generative frameworks -- Sources: [SRC-012]
- Schema-based vs. schema-free KG construction is a domain-dependent design choice -- Sources: [SRC-012], [SRC-008]
- HyperGraphRAG outperforms binary-edge GraphRAG in accuracy across multiple domains -- Sources: [SRC-006]
- GraphRAG frequently underperforms vanilla RAG on simple fact retrieval tasks -- Sources: [SRC-009]
- RAPTOR with GPT-4 achieves 20% absolute accuracy improvement on QuALITY benchmark -- Sources: [SRC-005]
- Dual-level retrieval (specific + abstract queries) captures both fine-grained and holistic knowledge -- Sources: [SRC-007]
- Incremental graph updates eliminate the need for full index rebuilds -- Sources: [SRC-007]
- GraphRAG's advantage is task-dependent, not universal -- Sources: [SRC-009]
- GMMs with UMAP enable soft clustering for multi-membership text segments -- Sources: [SRC-005]

### WEAK Evidence
- NLP noun-phrase extraction can substitute for LLM entity extraction during indexing without quality loss -- Sources: [SRC-004]
- Simpler domain schemas (5 classes) can outperform more complex schemas (8 classes) -- Sources: [SRC-008]
- Combining best-first and breadth-first search strategies is more effective than either alone -- Sources: [SRC-004]
- Dynamic question decomposition into follow-up queries enables adaptive KG traversal -- Sources: [SRC-013]
- Keyword-based semantic linking combined with structural graph traversal enhances context relevance -- Sources: [SRC-011]
- Dynamic knowledge memory for agentic systems is an emerging KG-LLM research direction -- Sources: [SRC-012]

### UNVERIFIED
- HyperGraphRAG's specific quantitative improvements over GraphRAG baselines (exact numbers not confirmed from accessible content) -- Basis: model training knowledge + abstract claims
- The optimal community detection granularity (Leiden resolution parameter) for engineering documentation corpora -- Basis: no source directly addresses this for engineering domains
- Whether frontmatter-typed metadata should be represented as node properties, separate entity types, or hyperedge attributes in a GraphRAG pipeline -- Basis: model training knowledge; no source directly compares these approaches

## Knowledge Gaps

- **Frontmatter-aware KG construction**: No source directly addresses how to incorporate structured document frontmatter (YAML metadata, typed headers) into knowledge graph construction pipelines. SRC-011 discusses document structure preservation but at a general level, not for frontmatter-typed engineering documentation specifically. This gap is significant for the knossos use case of `.know/` files with domain-typed YAML frontmatter.

- **Optimal schema design methodology**: SRC-008 demonstrates that expert schemas outperform generic ones, but no source provides a systematic methodology for designing domain-specific KG schemas for engineering knowledge bases. The field lacks empirical guidance on schema complexity (how many entity types, relationship types) for different domain characteristics.

- **GraphRAG for small-to-medium corpora**: Most evaluations use large corpora (100K-1M+ tokens). No source directly evaluates GraphRAG effectiveness on smaller, highly structured knowledge bases (1K-50K tokens) typical of engineering documentation collections. The overhead-to-benefit ratio may differ substantially.

- **Incremental hierarchical updates**: LightRAG addresses incremental graph updates, but no source addresses how to incrementally update hierarchical community structures (as in GraphRAG) or recursive abstraction trees (as in RAPTOR) when documents are added/modified without full re-indexing.

- **Cross-document relationship types for engineering knowledge**: No source evaluates which relationship types (depends-on, supersedes, contradicts, extends, implements) are most valuable for engineering documentation graphs versus general knowledge graphs.

## Domain Calibration

Mixed evidence distribution reflects a rapidly evolving field where foundational architectural claims (hierarchical retrieval, graph-structured knowledge) have strong support, but specific implementation choices (cost tradeoffs, schema design, n-ary representations) are still being empirically validated. The 2024-2025 publication density is exceptionally high, and many claims are based on single-system evaluations rather than independent replication.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research GraphRAG hierarchical retrieval` on 2026-03-25.
