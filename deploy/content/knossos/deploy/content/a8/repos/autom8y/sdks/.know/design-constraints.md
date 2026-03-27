---
domain: design-constraints
generated_at: "2026-03-16T15:40:19Z"
expires_after: "7d"
source_scope:
  - "./python/**/*.py"
  - "./python/**/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Duplicated HTTP Auth Client (`autom8y-auth` and `autom8y-core`)

**Location:**
- `python/autom8y-auth/src/autom8y_auth/http_client.py` — `Autom8yClient` class (447 lines)
- `python/autom8y-core/src/autom8y_core/client.py` — `Client` class (463 lines)

**Description:** `Autom8yClient` in `autom8y-auth` is a near-identical copy of `Client` in `autom8y-core`. Both classes implement the same interface: lazy-initialized sync/async httpx clients, Bearer token injection via `TokenManager`, identical method names. The `_compat.py` module in `autom8y-auth` acknowledges this explicitly: it maps `Autom8yClient` → `autom8y_core.Client`, `ClientConfig` → `autom8y_core.Config`, and `BaseServiceClient` → `autom8y_core.BaseClient` as deprecated imports.

**Evidence:** `python/autom8y-auth/src/autom8y_auth/_compat.py` lines 11-16. The migration to `autom8y-core` is incomplete — `Autom8yClient` in `http_client.py` still receives independent maintenance.

### TENSION-002: `autom8y-telemetry` Contains Domain-Specific Instrumentation

**Location:** `python/autom8y-telemetry/src/autom8y_telemetry/__init__.py` lines 95-100, plus:
- `python/autom8y-telemetry/src/autom8y_telemetry/gcal.py`
- `python/autom8y-telemetry/src/autom8y_telemetry/scheduling.py`
- `python/autom8y-telemetry/src/autom8y_telemetry/sms.py`
- `python/autom8y-telemetry/src/autom8y_telemetry/reconciliation.py`

**Description:** The telemetry package is supposed to be a cross-cutting infrastructure primitive. Instead, it directly exports domain decorators: `trace_gcal`, `trace_scheduling`, `trace_sms`, `trace_reconciliation`. This is a layering inversion — `autom8y-telemetry` has knowledge of every instrumented domain.

### TENSION-003: `autom8y-http/resilience/data_service_base.py` Hardcodes Service Identity

**Location:** `python/autom8y-http/src/autom8y_http/resilience/data_service_base.py` line 38

**Description:** `BaseDataServiceClient` in `autom8y-http` (a generic HTTP library) hardcodes `service="autom8y-data"` in the circuit breaker log callback. Any service using `BaseDataServiceClient` will emit circuit breaker logs falsely attributed to "autom8y-data". Additionally, `BaseDataServiceClient.__aenter__` hardcodes `env_prefix="AUTOM8_DATA_"` (line 62).

### TENSION-004: `autom8y-http` Optional Dependency on `autom8y-core` Creates Bidirectional Coupling

**Location:** `python/autom8y-http/pyproject.toml` optional deps `[core]`, and `python/autom8y-http/src/autom8y_http/resilience/`

**Description:** `autom8y-http` → `autom8y-log` → `autom8y-core` (transitive). `autom8y-http[core]` → `autom8y-core` (direct optional). The `resilience` submodule directly imports `autom8y_core.Client` and `autom8y_core.Config`.

### TENSION-005: `autom8y-devx-types` Package Naming Mismatch

**Location:** `python/autom8y-devx-types/pyproject.toml` lines 6 and 24

**Description:** Distribution: `autom8y-devx-types`, import: `autom8_devx_types` (no `y`). `[tool.hatch.build.targets.wheel]` reads `packages = ["src/autom8_devx_types"]`. This mismatch is permanent — changing the import name is a breaking change for all consumers.

### TENSION-006: `autom8y-core` `LoggerProtocol` Versus `autom8y-log` Re-export

**Location:**
- `python/autom8y-core/src/autom8y_core/protocols.py` — canonical definition
- `python/autom8y-log/src/autom8y_log/protocols.py` — re-exports from `autom8y_core.protocols`
- `python/autom8y-log/src/autom8y_log/__init__.py` — exports `LoggerProtocol`

**Description:** Two canonical import paths exist: `from autom8y_core.protocols import LoggerProtocol` and `from autom8y_log import LoggerProtocol`. The protocol lives in core but is "owned" by log.

### TENSION-007: `autom8y-gcal` Hardcodes Tenant-Specific SA Identity

**Location:** `python/autom8y-gcal/src/autom8y_gcal/_constants.py` lines 18-22

**Description:** `_EXPECTED_SA_CLIENT_EMAIL` and `_EXPECTED_SA_PROJECT_ID` are hardcoded. Not configurable via environment, constructor, or config file. Adding a second tenant requires modifying and releasing the SDK. The impersonation allowlist (`_APPROVED_IMPERSONATION_TARGETS`) is also hardcoded.

### TENSION-008: Sync/Async Dual Surface Area Throughout

**Location:** All core infrastructure packages:
- `python/autom8y-core/src/autom8y_core/client.py` — sync + async methods
- `python/autom8y-core/src/autom8y_core/token_manager.py` — `get_token()` + `get_token_async()`
- `python/autom8y-auth/src/autom8y_auth/http_client.py` — same dual surface
- `python/autom8y-http/src/autom8y_http/sync.py` — sync wrapper around async client

**Description:** Every HTTP primitive maintains both sync and async method variants. This doubles the API surface and creates maintenance burden: any fix must be applied to both code paths.

### TENSION-009: `ResilienceConfig` Uses Runtime Class Factory

**Location:** `python/autom8y-http/src/autom8y_http/resilience/config.py` lines 49-66, 231-252

**Description:** `ResilienceConfig` is not a class — it is a function that returns dynamically-created Pydantic `BaseSettings` subclasses via `type(...)`. Each unique `env_prefix` gets a new class cached via `@functools.lru_cache`. This bypasses Pydantic's class-level `model_config` mutation problem but means `isinstance(config, ResilienceConfig)` will not work.

## Trade-off Documentation

### TRADEOFF-001: Platform Auth Client Copied to `autom8y-core` (TENSION-001)

**Chosen:** Canonical client migrated to `autom8y-core` and deprecated in `autom8y-auth` via `_compat.py`.
**Rejected:** Keeping `Autom8yClient` canonical in `autom8y-auth` — would have required all non-auth packages to depend on `autom8y-auth` just for HTTP calls.
**Why persists:** Migration incomplete. `Autom8yClient` still exists as a full implementation. Deprecation message says "will be removed in autom8y-auth v1.0.0" — package is at 1.1.1.

### TRADEOFF-002: Domain Span Decorators Bundled in `autom8y-telemetry` (TENSION-002)

**Chosen:** Domain-specific `trace_*` decorators shipped directly in `autom8y-telemetry`.
**Rejected:** Putting each in its domain package — would require each domain to depend on `autom8y-telemetry` for the decorator.
**Why persists:** Any new instrumented domain requires an `autom8y-telemetry` release. A generic `trace_domain(namespace)` factory would eliminate the per-domain files.

### TRADEOFF-003: Hardcoded SA Identity in `autom8y-gcal` (TENSION-007)

**Chosen:** SA client email and project ID hardcoded as constants with no env-var override.
**Rejected:** Making them configurable — introduces credential substitution attack vector.
**Why persists:** Security trade-off. `_constants.py` comment: "Changes require a code change, PR review, and release." Second tenant (`contenteapp.com`) requires a code release.

### TRADEOFF-004: `autom8y-devx-types` Zero-Dependency Constraint (TENSION-005)

**Chosen:** Zero runtime dependencies — enables it to be the bridge between devx console (heavy deps) and domain packages (lightweight).
**Rejected:** Merging type definitions into `autom8y-telemetry` or another package with dependencies.
**Why persists:** The naming mismatch (`autom8y-devx-types` vs `autom8_devx_types`) is the residual cost.

### TRADEOFF-005: `ResilienceConfig` Dynamic Class Factory (TENSION-009)

**Chosen:** Function returning dynamically-created Pydantic subclasses via `lru_cache`.
**Rejected:** Single class with mutated `model_config.env_prefix` — Pydantic's `model_config` is class-level, not instance-level.
**Why persists:** Correct but counter-intuitive. Type checkers see through this partially.

## Abstraction Gap Mapping

### GAP-001: No Domain-Agnostic Span Decorator Factory

**Affected files:** `python/autom8y-telemetry/src/autom8y_telemetry/{gcal,scheduling,sms,reconciliation}.py`
**Description:** All four domain trace decorators have identical structure. A `trace_domain(namespace: str)` factory would eliminate all four files.

### GAP-002: No Unified HTTP Client Interface Across `autom8y-core` and `autom8y-http`

**Affected files:** `python/autom8y-core/src/autom8y_core/client.py`, `python/autom8y-http/src/autom8y_http/client.py`, `python/autom8y-http/src/autom8y_http/resilience/client.py`
**Description:** Three HTTP client implementations with overlapping but non-unified interfaces. No single interface ties all three together.

### GAP-003: Missing Multi-Tenant Abstraction in `autom8y-gcal`

**Affected files:** `python/autom8y-gcal/src/autom8y_gcal/_constants.py`, `auth.py`, `config.py`
**Description:** The SDK is designed for a single tenant. Adding `contenteapp.com` requires modifying three separate hardcoded constraints.

### GAP-004: No Shared `CacheSettings` Sync Contract

**Affected files:** `python/autom8y-cache/src/autom8y_cache/_config_integration.py`, `_settings.py`
**Description:** `Autom8yCacheSettings` adapts to `CacheSettings` (dataclass). Comment notes defaults "must stay in sync" — no enforcement exists.

### GAP-005: Premature Abstraction in `CircuitBreakerRegistry` Group Cap Logic

**Affected files:** `python/autom8y-http/src/autom8y_http/resilience/registry.py` lines 43-49, 171-181
**Description:** Cap warning is one-shot (`_cap_warned = True`). After cap, all rejected groups silently share a single default breaker.

## Load-Bearing Code

### LB-001: `TokenManager._refresh_sync()` and `._refresh_async()` in `autom8y-core`

**Location:** `python/autom8y-core/src/autom8y_core/token_manager.py` lines 293-393
**Must preserve:** (1) `_MAX_RETRY_AFTER = 30.0` cap for Lambda waits, (2) dual sync/async paths, (3) `_RetryableRateLimitError` → `RateLimitedError` transformation, (4) jitter window (`_JITTER_WINDOW = 10.0`) in `_needs_refresh()`.

### LB-002: `GoogleCalendarAuthProvider._validate_sa_identity()` and Dual-Lock Credential Isolation

**Location:** `python/autom8y-gcal/src/autom8y_gcal/auth.py` lines 130-163
**Must not refactor without security review:** (1) allowlist check, (2) separate `_readonly_creds`/`_events_creds` with separate locks, (3) T-60s proactive refresh buffer, (4) double-checked locking in `_refresh_credentials()`.

### LB-003: `autom8y-devx-types` Zero-Dependency Contract

**Location:** `python/autom8y-devx-types/pyproject.toml` line 16
**Frozen:** Must not acquire runtime dependencies.

### LB-004: `BaseClient._owns_client` Ownership Tracking in `autom8y-core`

**Location:** `python/autom8y-core/src/autom8y_core/base_client.py` lines 113-121
**Description:** Tracks whether `BaseClient` owns the underlying `Client` instance. Enables multiple service clients to share a single token cache. If removed, shared-client patterns will either double-close HTTP clients or leak them.

### LB-005: `CircuitBreakerRegistry` Append-Only Dict

**Location:** `python/autom8y-http/src/autom8y_http/resilience/registry.py` lines 53-62, 219-236
**Description:** `_breakers` dict is append-only by design. `get()` reads without locking (safe under GIL). Any change allowing removal or reconfiguration requires adding locking to `get()`.

## Evolution Constraints

### CONSTRAINT-01: `autom8y-devx-types` Module Name Cannot Change

**Scope:** `python/autom8y-devx-types/src/autom8_devx_types/`
**Status:** Frozen. Changing to `autom8y_devx_types` is a breaking change for all consumers.

### CONSTRAINT-02: `autom8y-gcal` Impersonation Allowlist Requires Code Release

**Scope:** `python/autom8y-gcal/src/autom8y_gcal/_constants.py`
**Status:** Intentionally frozen — code change required. Security constraint, not an oversight.

### CONSTRAINT-03: `autom8y-auth::Autom8yClient` Deprecation Is Incomplete

**Scope:** `python/autom8y-auth/src/autom8y_auth/http_client.py`
**Status:** In migration — incomplete. Should have been removed at v1.0.0 but package is at v1.1.1.

### CONSTRAINT-04: `autom8y-telemetry` Domain Modules Cannot Be Removed Without Breaking Consumers

**Scope:** `python/autom8y-telemetry/src/autom8y_telemetry/{gcal,scheduling,sms,reconciliation}.py`
**Status:** Load-bearing exports. Can be refactored (factory pattern) if public API signature preserved.

### CONSTRAINT-05: `autom8y-http` `resilience` Submodule Optional Dependency on `autom8y-core`

**Scope:** `python/autom8y-http/src/autom8y_http/resilience/`
**Status:** Deliberate opt-in via `[core]` extra. Frozen mechanism.

### CONSTRAINT-06: `ResilienceConfig` Is a Function, Not a Class

**Scope:** `python/autom8y-http/src/autom8y_http/resilience/config.py`
**Status:** Internal implementation detail. Cannot be changed without breaking API.

## Risk Zone Mapping

### RISK-001: `autom8y-auth::Autom8yClient.close()` Does Not Close Async Client in Sync Context

**Location:** `python/autom8y-auth/src/autom8y_auth/http_client.py` lines 290-295
**Description:** `close()` only closes `_sync_client` and `token_manager.close()`. Does not attempt to close `_async_client`. By contrast, `autom8y-core::Client.close()` has explicit async client cleanup handling. Unguarded resource leak.

### RISK-002: `autom8y-cache::_config_integration.py` Double-Import Guard Is Dead Code

**Location:** `python/autom8y-cache/src/autom8y_cache/_config_integration.py` lines 17-24
**Description:** Module-level `from autom8y_config import Autom8yBaseSettings` will fail before the `try/except ImportError` guard executes. The `CONFIG_AVAILABLE` flag is never `False` in practice. Dead code masquerading as defensive code.

### RISK-003: `BaseDataServiceClient` Service Identity Bleeding Across Satellites

**Location:** `python/autom8y-http/src/autom8y_http/resilience/data_service_base.py` line 38
**Description:** Hardcoded `service="autom8y-data"` and `env_prefix="AUTOM8_DATA_"` contaminates observability for any subclass. A satellite named "ads" using this base class would need to set `AUTOM8_DATA_CB_ENABLED`.

### RISK-004: `autom8y-gcal` Uses `requests` Library (Mixed HTTP Clients)

**Location:** `python/autom8y-gcal/pyproject.toml`
**Description:** `autom8y-gcal` depends on both `httpx` (via `autom8y-http`) and `requests` (for `google-auth` transport). Two HTTP client libraries in flight simultaneously. `requests` cannot be removed — google-auth's refresh mechanism requires `requests.Request` transport.

### RISK-005: `CircuitBreakerRegistry` Cap Warning Is One-Shot

**Location:** `python/autom8y-http/src/autom8y_http/resilience/registry.py` lines 171-181
**Description:** After cap hit, only the first rejection is logged. All subsequent rejections are silent. Could result in hundreds of routes sharing a single circuit breaker state in long-running services.

## Knowledge Gaps

1. **`autom8y-auth` client migration completion status:** Unknown whether satellite services still import `Autom8yClient` from `autom8y_auth.http_client`.

2. **`autom8y-events` relationship to `autom8y-telemetry`**: Whether event publish is instrumented at the satellite layer was not explored.

3. **`autom8y-meta` role in the dependency DAG**: Source files were not read in this observation.

4. **Convention registry in `autom8y-telemetry`**: `conventions/` subdirectory may contain additional evolution constraints or risk zones.

5. **`autom8y-sms-test` purpose**: Relationship to the broader telemetry/instrumentation story was not fully explored.
