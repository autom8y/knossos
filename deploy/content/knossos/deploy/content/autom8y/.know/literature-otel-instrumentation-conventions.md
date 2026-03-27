---
name: literature-otel-instrumentation-conventions
domain: otel-instrumentation-conventions
type: literature
created: 2026-03-14
evidence_grade: mixed (A-C depending on source)
sources_evaluated: 34
sources_cited: 28
---

# Literature Review: OpenTelemetry Domain-Specific Semantic Conventions

## Executive Summary

OpenTelemetry semantic conventions have matured into a multi-domain taxonomy covering 20+ technology areas (HTTP, DB, messaging, GenAI, CI/CD, FaaS, etc.), governed by a five-level maturity lifecycle (development, alpha, beta, release_candidate, stable). As of v1.40.0, only HTTP and core database conventions (MySQL, PostgreSQL, MariaDB, MSSQL) have reached stable status; all other domains remain experimental/development. The GenAI semantic conventions (gen_ai.* namespace), introduced in v1.26.0 following Phillip Carter's September 2023 proposal (Issue #327), define a comprehensive attribute set for LLM inference, embeddings, retrieval, tool execution, and agent operations -- but remain at Development maturity. Vendor adoption is strong: Datadog, Elastic, AWS, and others map OTel conventions to their proprietary schemas while adding vendor-specific extensions.

The convention extension model is well-defined: organizations should use reverse-domain prefixes (e.g., `com.acme.shopname`) for custom attributes and avoid squatting on existing OTel namespaces. OTel Weaver, the schema-first CLI tool, now supports custom registries that inherit from the official OTel registry, enabling organizations to define, validate, and code-generate typed constants for domain-specific conventions. This federated model -- where third parties publish their own convention registries -- is a strategic direction endorsed by the 2025 stability proposal. For autom8y, this means the `autom8y-telemetry` SDK can define a `com.autom8y.*` or organization-scoped namespace for scheduling, booking, and reconciliation domain attributes, validated by Weaver and code-generated into typed Python constants.

Test-trace correlation has two mature options in Python: `pytest-otel` (Elastic-affiliated, OTLP export of test spans) and `pytest-opentelemetry` (community, xdist integration, trace-parent propagation). Combined with OpenTelemetry's `InMemorySpanExporter` for assertion-based testing, these tools enable a pattern where the same instrumentation that runs in production is verified in CI, closing the loop between observability-by-design and test-driven development.

## Methodology

**Search strategy**: Web searches across OpenTelemetry official documentation (opentelemetry.io), GitHub repositories (open-telemetry/*), vendor documentation (Datadog, Elastic, AWS), community guides (BetterStack, Honeycomb, Chronosphere), PyPI package registries, CNCF blog, and technical blogs. 34 sources evaluated; 28 cited based on relevance and authority.

**Source types**: Official OTel specification and blog posts (Grade A), vendor documentation and core maintainer content (Grade B), community guides and tutorials (Grade C), forum discussions (Grade D).

**Inclusion criteria**: Sources must address semantic convention design, extension patterns, instrumentation library architecture, convention lifecycle, or test-trace correlation. Excluded: general "getting started with OTel" tutorials without convention-specific content, product marketing pages, and sources older than 2022 unless they document foundational decisions.

**Limitation**: OTEP-0248 was referenced in the research prompt but could not be located as a specific OTEP document. The GenAI conventions appear to have evolved through Issue #327 on the semantic-conventions repo rather than through the traditional OTEP process, following the Semantic Conventions Working Group's preferred workflow for convention proposals (as distinct from specification-level OTEPs).

## Findings

### 1. OTel Semantic Conventions Architecture

**Evidence Grade: A (primary specification sources)**

OpenTelemetry semantic conventions v1.40.0 define standardized attribute names, metric names, span names, and event names across 20+ technology domains [1][3]. The conventions are organized by signal type (traces, metrics, logs, events, resources, profiles) and by technology domain.

**Domain taxonomy** (as of v1.40.0) [4]:

| Domain | Status | Key Namespace |
|--------|--------|---------------|
| HTTP | **Stable** (since late 2023) | `http.request.*`, `http.response.*` |
| Database | **Stable** (core: MySQL, PostgreSQL, MariaDB, MSSQL) | `db.*` |
| Messaging | Experimental | `messaging.*` |
| RPC/gRPC | Experimental | `rpc.*` |
| FaaS | Experimental | `faas.*` |
| GenAI/LLM | Development | `gen_ai.*` |
| CI/CD | Experimental (since v1.27.0) | `cicd.*` |
| Cloud Providers | Experimental | `aws.*`, `gcp.*`, `azure.*` |
| Feature Flags | Experimental | `feature_flag.*` |
| GraphQL | Experimental | `graphql.*` |
| System/Hardware | Experimental | `system.*`, `hw.*` |
| Mobile | Experimental | `device.*` |

**Naming rules** [5]:
- All names must be lowercase, valid Unicode
- Dot notation for namespace hierarchy: `{namespace}.{object}.{attribute}`
- Snake_case for multi-word components: `http.response.status_code`
- The `otel.*` namespace is exclusively reserved for OTel specification
- Two attributes MUST NOT share the same name
- Singular names for single entities (`host.name`); plural for arrays (`process.command_args`)

**Stability lifecycle** [6][7]: Conventions progress through five maturity levels: development, alpha, beta, release_candidate, and stable. Group stability MUST NOT change from stable to any lower level. Stable groups MAY reference unstable attributes only with `opt_in` requirement level. Stable instrumentations MUST NOT report unstable convention telemetry by default.

**Key insight for autom8y**: The HTTP stable migration (renaming `http.method` to `http.request.method`) caused breaking changes that required the `OTEL_SEMCONV_STABILITY_OPT_IN` dual-emit pattern [8]. Any custom convention design should anticipate this pattern from the start.

### 2. GenAI Semantic Conventions (gen_ai.*)

**Evidence Grade: A (specification) / B (vendor implementation reports)**

The GenAI semantic conventions emerged from Phillip Carter's Issue #327 (September 15, 2023) proposing semantic conventions for "modern AI (LLMs, vector databases, etc.)" [9]. Carter argued that LLM applications are a "cookie-cutter case for distributed tracing" due to their chain-of-calls architecture with high-cardinality natural language inputs/outputs. The conventions were formally introduced in v1.26.0 and are at Development maturity as of v1.40.0 [10].

**Attribute taxonomy** (from the specification) [11]:

| Category | Key Attributes | Type | Requirement |
|----------|---------------|------|-------------|
| Operation Identity | `gen_ai.operation.name`, `gen_ai.provider.name` | string | Required |
| Request Config | `gen_ai.request.model`, `gen_ai.request.temperature`, `gen_ai.request.max_tokens`, `gen_ai.request.top_p`, `gen_ai.request.top_k` | string/double/int | Conditionally Required / Recommended |
| Response Metadata | `gen_ai.response.id`, `gen_ai.response.model`, `gen_ai.response.finish_reasons` | string/string[] | Recommended |
| Token Usage | `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, `gen_ai.usage.cache_creation.input_tokens`, `gen_ai.usage.cache_read.input_tokens` | int | Recommended |
| Content (Opt-In) | `gen_ai.input.messages`, `gen_ai.output.messages`, `gen_ai.system_instructions`, `gen_ai.tool.definitions` | any | Opt-In (PII sensitive) |
| Tool Execution | `gen_ai.tool.name`, `gen_ai.tool.type`, `gen_ai.tool.call.id`, `gen_ai.tool.call.arguments`, `gen_ai.tool.call.result` | string/any | Recommended/Opt-In |
| Conversation | `gen_ai.conversation.id` | string | Conditionally Required |

**Well-known operation names**: `chat`, `create_agent`, `embeddings`, `execute_tool`, `generate_content`, `invoke_agent`, `retrieval`, `text_completion` [11].

**Well-known provider names**: `anthropic`, `aws.bedrock`, `azure.ai.inference`, `azure.ai.openai`, `cohere`, `deepseek`, `gcp.gemini`, `groq`, `ibm.watsonx.ai`, `mistral_ai`, `openai`, `perplexity`, `x_ai` [11].

**Vendor-specific convention pages** exist for Anthropic, Azure AI Inference, AWS Bedrock, OpenAI, and Model Context Protocol (MCP) [10].

**Agentic systems extension** (Issue #2664, August 2025): Dany Moshkovich proposed extending gen_ai.* for agentic systems with six interconnected domains: tasks, actions, agents, teams, artifacts, and memory [12]. Status: 1 of 10 sub-issues completed; the proposal is in early discussion.

**Vendor adoption**: Datadog natively supports GenAI Semantic Conventions v1.37+ in its LLM Observability product, automatically mapping `gen_ai.request.model`, `gen_ai.usage.*`, and `gen_ai.operation.name` to Datadog's schema [13]. This cross-layer integration correlates GenAI spans with full-stack APM traces.

**Stability transition mechanism**: For existing instrumentations using v1.36.0 or earlier conventions, `OTEL_SEMCONV_STABILITY_OPT_IN=gen_ai_latest_experimental` enables the latest experimental version without the old one [10].

**Key insight for autom8y**: The gen_ai.* namespace is the most recent successful domain extension, demonstrating the full lifecycle from community proposal (Issue #327) to specification inclusion (v1.26.0) to vendor adoption (Datadog v1.37+). This is the template for any autom8y-specific convention proposal. The opt-in content capture pattern (input/output messages are opt-in due to PII) is directly relevant to scheduling domain attributes that may contain client PII.

### 3. Custom Convention Extension Patterns

**Evidence Grade: A (naming specification) / B (vendor docs) / B-C (community guides)**

**Official namespacing guidance** [5]:
- **Company-specific attributes**: Prefix with reverse domain name: `com.acme.shopname`
- **Application-internal attributes**: Prefix with unique application name: `myuniquemapapp.longitude`
- **Critical prohibition**: Do NOT use existing OTel namespace prefixes for custom attributes -- future OTel additions could clash

**Vendor extension patterns**:

**Datadog** [14][13]: Maps OTel semantic conventions to Datadog's proprietary schema. Datadog has a vendor-specific `span.type` attribute inferred from OTel span attributes. Unmapped OTel attributes flow through as custom span tags. Custom attributes are set via `span.setAttribute("key", value)` using the standard OTel API. Datadog's "semantic mapping" documentation provides explicit tables showing OTel-to-Datadog attribute translation for resource attributes, span attributes, and metrics.

**Elastic APM** [15][16]: Elastic donated the Elastic Common Schema (ECS) to OpenTelemetry in April 2023, with the goal of convergence between ECS and OTel Semantic Conventions into a single schema. Unmapped OTel resource attributes become "global labels" in Elastic APM; dots in attribute names are replaced by underscores. The ECS convergence brought mature fields for geo information, threat fields, and security events that OTel lacked. Elastic's approach is the most significant vendor contribution to the OTel convention ecosystem -- it established the precedent that vendor schemas should converge into OTel rather than maintaining parallel taxonomies.

**AWS X-Ray** [17]: X-Ray distinguishes between annotations (indexed, filterable key-value pairs) and metadata (non-indexed, serializable objects). By default, OTel span attributes convert to X-Ray metadata. Specific attributes can be promoted to annotations for filterability. This annotation/metadata distinction is a useful pattern for autom8y: domain attributes that need to be searchable (e.g., `booking.business_id`) should be designed as annotations, not just metadata.

**Organizational standards** (Chronosphere, Honeycomb) [18][19]: Building a "data dictionary" to standardize custom attributes across teams is recommended. Custom attributes should be codified in a shared library to prevent name sprawl and collision. Honeycomb emphasizes that attribute names should "express type, not context" -- e.g., `aws.s3.bucket` rather than `request.bucket_name`. Never encode variables in attribute names; use array types or span events for structured data.

**OTel Weaver for custom registries** [20][21]: OTel Weaver is a CLI tool that validates, generates code from, and evolves semantic convention registries. Organizations can define custom registries via `registry_manifest.yaml` files that extend the official OTel registry through a dependency/inheritance model. Weaver supports:
- Multi-language code generation (Go, Java, Python, etc.) from YAML definitions
- 30+ policy-based validation rules (naming, stability, immutability, collision detection)
- Custom Rego-based policies for organization-specific invariants
- Registry diff detection for breaking changes between versions
- Live compliance checking against running applications
- Telemetry simulation for testing dashboards before live instrumentation

Currently supports two-level registry hierarchies; deeper nesting is on the roadmap.

**Key insight for autom8y**: The Weaver custom registry model is the strategic path for autom8y. Define a `registry_manifest.yaml` that imports the OTel core registry and adds `com.autom8y.*` domain attributes. Weaver generates typed Python constants, validates naming compliance, and detects breaking changes in CI. This is more robust than manually maintaining attribute name strings in the `autom8y-telemetry` SDK.

### 4. Instrumentation Library Design for Business Logic

**Evidence Grade: A (Python SDK docs) / B (vendor guides) / C (community patterns)**

**The boundary between auto and manual instrumentation** [22][23]: Auto-instrumentation handles framework-level telemetry (HTTP server spans, database client spans, Redis/cache spans) through monkey-patching or import hooks. Manual instrumentation is required for business logic: order processing, payment workflows, scheduling operations, reconciliation runs. Most production applications need both -- auto-instrumentation for infrastructure, manual for domain semantics.

**Python instrumentation patterns** [22]:

1. **Context manager** (recommended for operation blocks):
   ```python
   with tracer.start_as_current_span("booking.create") as span:
       span.set_attribute("booking.business_id", business_id)
       span.set_attribute("booking.slot_count", len(slots))
       result = create_booking(...)
   ```

2. **Decorator** (recommended for function-level tracing):
   ```python
   @tracer.start_as_current_span("reconciliation.run")
   def run_reconciliation():
       ...
   ```

3. **Semantic convention constants** (recommended for standardized attributes):
   ```python
   from opentelemetry.semconv.trace import SpanAttributes
   span.set_attribute(SpanAttributes.HTTP_METHOD, "GET")
   ```

**BaseInstrumentor pattern for library authors** [24]: The `BaseInstrumentor` abstract base class provides the standard pattern for creating instrumentation libraries that integrate with the `opentelemetry-instrument` auto-instrumentation CLI:
- `instrumentation_dependencies()` -- declares which package versions are instrumented (requirements.txt format)
- `_instrument()` -- applies monkey-patching or wrapping
- `_uninstrument()` -- reverses instrumentation
- The `instrument()` method must work without optional arguments (compatibility with `opentelemetry-instrument` CLI)

**Library vs. application instrumentation** [22]: Libraries should depend only on `opentelemetry-api` (not the SDK), emitting telemetry only when the consuming application provides a TracerProvider. This allows the autom8y SDK to add instrumentation without forcing SDK dependencies on satellite services.

**Honeycomb guidance on business logic spans** [19]: "Automatic instrumentation does not know your business logic -- it only knows about frameworks and languages." Manual instrumentation should describe "meaningful operations -- placing an order, calculating pricing, processing a job -- in terms that match how the system works in practice." Teams should add instrumentation gradually, guided by real questions and real incidents, not upfront design. Span design should "balance signal and noise" -- avoid creating spans for unbounded loops; emit aggregate metrics instead.

**Key insight for autom8y**: The `autom8y-telemetry` SDK already implements `instrument_app()` and `@instrument_lambda` (auto-instrumentation for FastAPI and Lambda). The next layer is a domain instrumentation module that provides context managers and decorators for business logic spans: `@trace_booking`, `@trace_reconciliation`, `@trace_calendar_sync`, etc. These would automatically apply the `com.autom8y.*` attribute namespace and set domain-specific attributes, bridging the gap identified in the observability posture audit where Lambda services have only shallow root spans (P3-2).

### 5. Convention Versioning and Backward Compatibility

**Evidence Grade: A (specification sources)**

**Signal lifecycle** [25]: OTel signals progress through Development -> Stable -> Deprecated -> Removed. Backward-incompatible changes to stable API packages require a major version bump. Minor version bumps are used for new functionality, development-stage breaking changes, and stability transitions.

**Telemetry Schemas** [26]: OTel uses telemetry schemas to manage convention evolution. Schema files define transformations between versions (primarily attribute/metric/event renames). Schema URLs follow `http[s]://server/path/<version>` and are embedded in OTLP messages at ResourceSpans, ResourceMetrics, ResourceLogs, and InstrumentationLibrary levels. Schema files are immutable once published. Supported transformations are intentionally limited to renames; removals constitute breaking changes.

**OTLP integration**: Schema URLs propagate through the OTLP protocol, enabling backends to automatically translate telemetry from one schema version to another. The OTel Collector can apply schema transformations, allowing backends unaware of schemas to receive data in their expected format.

**Practical migration example**: HTTP conventions renamed `http.method` to `http.request.method` and `http.client_id` to `client.address` in the stable release. The `OTEL_SEMCONV_STABILITY_OPT_IN` environment variable with values like `http` (stable only) or `http/dup` (both old and new) enables gradual migration [8].

**2025 stability proposal** [27]: The OpenTelemetry Governance Committee proposed three major initiatives:
1. **Standardized stability metadata** -- machine-parseable stability indicators across all repos
2. **Federated conventions** -- decouple instrumentation stability from convention maturity; third parties can publish their own convention registries
3. **Epoch releases** -- tested, documented bundles combining specific SDK + instrumentation + Collector versions for enterprise deployment

The proposal explicitly states instrumentation libraries can stabilize independently of semantic conventions, which is critical for the autom8y SDK: `autom8y-telemetry` can declare stable API surfaces even while using experimental convention attributes, provided those experimental attributes are opt-in.

**Key insight for autom8y**: Design the `com.autom8y.*` attribute namespace with schema versioning from day one. Define a `registry_manifest.yaml` with version tracking, and use Weaver's `registry diff` to detect breaking changes in CI before they reach production. The dual-emit pattern (`OTEL_SEMCONV_STABILITY_OPT_IN`) should be adopted for any attribute renames.

### 6. Test-Trace Correlation (pytest-otel)

**Evidence Grade: B (PyPI documentation, GitHub repos)**

Two complementary pytest plugins exist for test-trace correlation in Python:

**pytest-otel** [28]: Elastic-affiliated plugin for reporting test spans via OTLP.
- Requires Python 3.10+
- Configuration via CLI args: `--otel-endpoint`, `--otel-headers`, `--otel-service-name`, `--otel-session-name`, `--otel-traceparent`
- Supports `--otel-exporter-protocol` (grpc or http/protobuf)
- W3C traceparent propagation via `--otel-traceparent` for nesting test runs in CI traces
- Demo configs for Jaeger and Elastic Stack backends

**pytest-opentelemetry** [29]: Community plugin by Chris Guidry.
- Requires Python 3.8+
- Automatic span creation per test run
- Native pytest-xdist integration: "automatically unite [distributed tests] all under one trace"
- `--trace-parent` argument for nesting in larger traces (W3C format)
- Uses project directory name as `service.name` by default
- Respects `OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES`
- Works "even better when testing applications that are themselves instrumented with OpenTelemetry," providing deep visibility into database queries and network requests made during tests

**InMemorySpanExporter for assertion testing** [30]: The OpenTelemetry Python SDK includes `InMemorySpanExporter` for capturing spans during test execution without external collectors. Common pattern:

```python
@pytest.fixture(scope="session", autouse=True)
def configure_otel():
    exporter = InMemorySpanExporter()
    provider = TracerProvider()
    provider.add_span_processor(SimpleSpanProcessor(exporter))
    trace.set_tracer_provider(provider)
    yield exporter
```

This enables assertion patterns:
- Verify span names and attributes
- Assert parent-child span hierarchies
- Validate span duration for performance contracts
- Check span status codes and recorded exceptions

**Trace-based testing patterns** [31]: Trace-based testing validates distributed system behavior by examining actual request flows:
- Contract testing: verify spans and attributes match service interface specs
- Error propagation testing: confirm errors cascade correctly through span hierarchies
- Performance assertion: validate span durations against latency SLOs
- CI/CD integration: spin up collectors during test pipelines, export traces as artifacts

**Key insight for autom8y**: The `InMemorySpanExporter` pattern is directly applicable to the autom8y test suite. A shared pytest fixture in `autom8y-telemetry` could provide `captured_spans` to any satellite test, enabling assertions like "the booking creation span has `com.autom8y.booking.business_id` attribute set." This closes the observability posture gap P3-3 ("No automated trace correlation test"). The `pytest-opentelemetry` plugin with `--trace-parent` can correlate CI test runs with the broader deployment trace when triggered by satellite-dispatch.

### 7. Domain-Specific Vertical Conventions

**Evidence Grade: B (CI/CD conventions) / C (community patterns for other verticals)**

**CI/CD as a model for domain extension** [32]: The CI/CD Observability SIG introduced semantic conventions in v1.27.0, defining the `cicd.*`, `vcs.*`, `deployment.*`, and `artifact.*` namespaces. This is the most recent domain addition to the OTel specification and demonstrates the full lifecycle:
1. Community interest (CI/CD observability SIG formation)
2. OTEP proposal (PR #223, later redirected to semconv WG workflow)
3. Foundational attributes in v1.27.0
4. Ongoing iteration (metrics, additional attributes)

The CI/CD conventions model pipeline runs as traces, with each pipeline stage as a span -- directly analogous to modeling a booking workflow or reconciliation run as a trace.

**No pre-built vertical conventions exist** for healthcare, fintech, scheduling, or e-commerce domains in the OTel specification. The specification intentionally limits its scope to technology-horizontal concerns (HTTP, DB, messaging) and leaves vertical/industry conventions to organizations [3]. This is by design: the specification states that the Semantic Conventions project has a roadmap, and "contributions outside these areas may be delayed or not accepted."

**Community patterns for vertical conventions**:
- Honeycomb recommends defining domain-specific conventions using OTel syntax for documentation and code generation reuse [19]
- Chronosphere recommends building a company-wide "data dictionary" for custom attributes, codified in a shared library [18]
- The reverse-domain prefix pattern (`com.acme.*`) is universally recommended for organization-specific attributes [5]

**Scheduling domain convention design** (derived from OTel patterns):

A hypothetical `com.autom8y.scheduling.*` namespace following OTel naming conventions would include:

| Attribute | Type | OTel Analog |
|-----------|------|-------------|
| `com.autom8y.scheduling.business.id` | string | `db.system.name` (system identifier) |
| `com.autom8y.scheduling.booking.id` | string | `gen_ai.conversation.id` (operation identifier) |
| `com.autom8y.scheduling.calendar.id` | string | `messaging.destination.name` (target resource) |
| `com.autom8y.scheduling.slot.start_time` | string (ISO 8601) | N/A (domain-specific) |
| `com.autom8y.scheduling.slot.duration_minutes` | int | N/A (domain-specific) |
| `com.autom8y.scheduling.operation.name` | string | `gen_ai.operation.name` (operation type) |
| `com.autom8y.scheduling.gcal.sync.status` | string | `db.operation.name` (operation result) |

This follows the `{namespace}.{object}.{property}` pattern from the naming specification [5] and uses the gen_ai.* conventions as a structural template.

**Key insight for autom8y**: No existing OTel vertical convention covers scheduling or booking domains. autom8y must define its own conventions using the reverse-domain prefix pattern, and should model them structurally after gen_ai.* (the most recent successful domain addition). The Weaver custom registry model provides the tooling to formalize, validate, and code-generate these conventions.

## Evidence Grades

- **A**: Primary source -- OTel specification, OTEP documents, official OTel documentation
- **B**: Authoritative secondary -- vendor documentation, core maintainer blogs, CNCF blog posts
- **C**: Community source -- blog posts, tutorials, guides from observability companies
- **D**: Anecdotal -- comments, forum posts, Stack Overflow

## Source Registry

| # | Source | URL | Type | Grade |
|---|--------|-----|------|-------|
| 1 | OTel Semantic Conventions Concepts | https://opentelemetry.io/docs/concepts/semantic-conventions/ | Official docs | A |
| 2 | OTel Semantic Conventions v1.40.0 Spec | https://opentelemetry.io/docs/specs/semconv/ | Specification | A |
| 3 | OTel Semantic Conventions GitHub | https://github.com/open-telemetry/semantic-conventions | Source repo | A |
| 4 | OTel Semconv Domain Taxonomy (v1.40.0) | https://opentelemetry.io/docs/specs/semconv/ | Specification | A |
| 5 | OTel Naming Conventions | https://opentelemetry.io/docs/specs/semconv/general/naming/ | Specification | A |
| 6 | OTel Group Stability Specification | https://opentelemetry.io/docs/specs/semconv/general/group-stability/ | Specification | A |
| 7 | OTel Versioning and Stability | https://opentelemetry.io/docs/specs/otel/versioning-and-stability/ | Specification | A |
| 8 | BetterStack: Missing Guide to OTel Semantic Conventions | https://betterstack.com/community/guides/observability/opentelemetry-semantic-conventions/ | Community guide | C |
| 9 | Issue #327: Semantic Conventions for Modern AI | https://github.com/open-telemetry/semantic-conventions/issues/327 | GitHub issue | A |
| 10 | OTel GenAI Semantic Conventions | https://opentelemetry.io/docs/specs/semconv/gen-ai/ | Specification | A |
| 11 | OTel GenAI Client Span Attributes | https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-spans/ | Specification | A |
| 12 | Issue #2664: GenAI Agentic Systems Conventions | https://github.com/open-telemetry/semantic-conventions/issues/2664 | GitHub issue | A |
| 13 | Datadog LLM Observability + OTel GenAI SemConv | https://www.datadoghq.com/blog/llm-otel-semantic-convention/ | Vendor blog | B |
| 14 | Datadog Semantic Mapping Documentation | https://docs.datadoghq.com/opentelemetry/mapping/semantic_mapping/ | Vendor docs | B |
| 15 | Elastic APM + OpenTelemetry Attributes | https://www.elastic.co/docs/solutions/observability/apm/opentelemetry/attributes | Vendor docs | B |
| 16 | ECS + OTel Convergence Announcement | https://opentelemetry.io/blog/2023/ecs-otel-semconv-convergence/ | Official blog | A |
| 17 | AWS X-Ray Annotations and Metadata (Python) | https://docs.aws.amazon.com/xray/latest/devguide/xray-sdk-python-segment.html | Vendor docs | B |
| 18 | Chronosphere: OTel Attribute Naming Best Practices | https://chronosphere.io/learn/top-3-opentelemetry-attribute-naming-best-practices/ | Vendor blog | B |
| 19 | Honeycomb: Effective Trace Instrumentation with Semantic Conventions | https://www.honeycomb.io/blog/effective-trace-instrumentation-semantic-conventions | Vendor blog | B |
| 20 | OTel Weaver GitHub Repository | https://github.com/open-telemetry/weaver | Source repo | A |
| 21 | OTel Blog: Observability by Design with Weaver | https://opentelemetry.io/blog/2025/otel-weaver/ | Official blog | A |
| 22 | OTel Python Manual Instrumentation Guide | https://opentelemetry.io/docs/languages/python/instrumentation/ | Official docs | A |
| 23 | Cribl: Manual vs Auto Instrumentation | https://cribl.io/blog/manual-vs-auto-instrumentation-opentelemetry-choose-whats-right/ | Vendor blog | C |
| 24 | OTel Python BaseInstrumentor Documentation | https://opentelemetry-python-contrib.readthedocs.io/en/latest/instrumentation/base/instrumentor.html | Official docs | A |
| 25 | OTel Versioning and Stability Specification | https://opentelemetry.io/docs/specs/otel/versioning-and-stability/ | Specification | A |
| 26 | OTel Telemetry Schemas | https://opentelemetry.io/docs/specs/otel/schemas/ | Specification | A |
| 27 | OTel 2025 Stability Proposal Announcement | https://opentelemetry.io/blog/2025/stability-proposal-announcement/ | Official blog | A |
| 28 | pytest-otel PyPI | https://pypi.org/project/pytest-otel/ | Package registry | B |
| 29 | pytest-opentelemetry GitHub | https://github.com/chrisguidry/pytest-opentelemetry | Source repo | B |
| 30 | OTel Python InMemorySpanExporter Source | https://github.com/open-telemetry/opentelemetry-python/blob/main/opentelemetry-sdk/src/opentelemetry/sdk/trace/export/in_memory_span_exporter.py | Source code | A |
| 31 | OneUptime: Trace-Based Testing with OpenTelemetry | https://oneuptime.com/blog/post/2026-01-07-opentelemetry-trace-based-testing/view | Blog post | C |
| 32 | CNCF: OpenTelemetry Expanding into CI/CD Observability | https://www.cncf.io/blog/2024/11/04/opentelemetry-is-expanding-into-ci-cd-observability/ | CNCF blog | B |

## Implications for autom8y

### RD-1: Instrumentation Convention Design

1. **Namespace decision**: Use `com.autom8y.*` (reverse-domain prefix per OTel naming spec [5]) for all custom domain attributes. Do NOT use `scheduling.*` or `booking.*` without prefix -- these could collide with future OTel standard conventions.

2. **Convention registry**: Adopt OTel Weaver [20][21] with a `registry_manifest.yaml` that imports the OTel core registry and defines autom8y-specific attribute groups. This gives type-safe Python constants, CI validation, and automated documentation generation.

3. **Domain attribute design**: Model after gen_ai.* conventions [11]:
   - Required: `com.autom8y.operation.name`, `com.autom8y.service.name`
   - Conditionally Required: `com.autom8y.booking.id`, `com.autom8y.business.id`
   - Recommended: `com.autom8y.calendar.id`, `com.autom8y.slot.start_time`
   - Opt-In (PII): `com.autom8y.client.name`, `com.autom8y.client.phone`

4. **Instrumentation library extension**: Extend `autom8y-telemetry` with domain-specific decorators and context managers that auto-apply the `com.autom8y.*` namespace. This addresses observability posture gap P3-2 (shallow Lambda spans) by providing `@trace_booking`, `@trace_reconciliation`, etc. that create properly-attributed child spans.

### RD-6: Telemetry Convention Standardization

1. **Vendor compatibility matrix**: The autom8y stack uses Grafana Cloud (Tempo + Loki + Mimir). Custom `com.autom8y.*` attributes will flow through as-is to Tempo and be searchable in TraceQL. No vendor-specific mapping is needed (unlike Datadog or Elastic which have proprietary schemas).

2. **Stability versioning**: Adopt the five-level maturity model [6]. Start all `com.autom8y.*` attributes at Development. Promote to Stable only after 2+ sprints of production use with no attribute changes. Use `OTEL_SEMCONV_STABILITY_OPT_IN=com_autom8y_latest` pattern for migrations.

3. **Schema evolution**: Define telemetry schema URLs [26] for autom8y convention versions. Embed schema URLs in OTLP exports from `autom8y-telemetry` to enable future backend-side migration.

### RD-4: Developer Workflow Integration

1. **Test-trace correlation**: Add `pytest-opentelemetry` [29] to the autom8y test stack for CI trace visibility. Its xdist integration is valuable for parallelized satellite test suites.

2. **Assertion-based span testing**: Create a shared pytest fixture in `autom8y-telemetry` providing `InMemorySpanExporter` [30] for verifying domain attribute compliance in unit tests. This closes observability posture gap P3-3.

3. **CI/CD trace correlation**: Use `--trace-parent` in CI pipelines to nest test runs under deployment traces, connecting satellite-dispatch -> satellite-receiver -> test execution in a single trace.

4. **Convention compliance checking**: Use Weaver's `registry live-check` [21] to verify running services emit telemetry conforming to the `com.autom8y.*` convention registry. Integrate into CI as a quality gate alongside existing linting.
