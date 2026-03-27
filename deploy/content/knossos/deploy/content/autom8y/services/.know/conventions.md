---
domain: conventions
generated_at: "2026-03-16T20:07:00Z"
expires_after: "7d"
source_scope:
  - "./*/src/**/*.py"
  - "./*/tests/**/*.py"
  - "./*/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

### Philosophy

The project uses a **domain-scoped exception hierarchy** with machine-readable error codes and HTTP status mappings. Every service that exposes domain logic defines its own base exception class, and all service-specific errors inherit from it. The pattern was explicitly documented as a cross-service convention (referenced in `reconcile-spend/src/reconcile_spend/errors.py` and `account-status-recon/src/account_status_recon/errors.py` as "Follows the established service error hierarchy pattern (.know/conventions.md: ClassVar code + http_status)").

### Error Creation Pattern

**Canonical pattern** (Lambda services and FastAPI services alike):

```python
class ServiceNameError(Exception):
    code: ClassVar[str] = "SERVICE_NAME_ERROR"
    http_status: ClassVar[int] = 500

class FetchError(ServiceNameError):
    code: ClassVar[str] = "FETCH_ERROR"
    http_status: ClassVar[int] = 503
```

Every base exception carries two `ClassVar` attributes:
- `code`: A `SCREAMING_SNAKE_CASE` string identifying the error machine-readably
- `http_status`: The integer HTTP status the error maps to at API boundaries

The `ClassVar` declarations live under `TYPE_CHECKING` in simpler services (reconcile-*, account-status-recon), but are full module-level `ClassVar` annotations with initialized values in more elaborate services (ads).

**Richer base class** (ads service): Base error also carries `message`, `context`, and a `to_dict()` method for serialization to JSON:
- `ads/src/autom8_ads/errors.py`

**Simpler base class** (reconcile-*, account-status-recon): No `to_dict()`, carry only `code` and `http_status`.

**Evidence**:
- `ads/src/autom8_ads/errors.py` — richest pattern, `to_dict()`, 7 exception classes
- `reconcile-spend/src/reconcile_spend/errors.py` — simpler variant, 4 classes
- `reconcile-ads/src/reconcile_ads/errors.py` — 4 classes
- `account-status-recon/src/account_status_recon/errors.py` — 4 classes

### Exception Specialization Patterns

Services define exceptions with structured `__init__` signatures carrying semantic fields:

```python
class FetchError(ReconcileAdsError):
    def __init__(self, source_name: str, method: str, *, time_remaining: float = 0.0) -> None:
        self.source_name = source_name
        super().__init__(f"{source_name}.{method} unavailable (retry in {time_remaining:.0f}s)")
```

This pattern is nearly identical across `reconcile-ads`, `reconcile-spend`, and `account-status-recon` — indicating copy-propagation from a shared template.

For the auth service, exception hierarchies are defined per-subsystem rather than in a central `errors.py`:
- `auth/src/services/credential_vault.py` — `CredentialVaultError` and 4 subclasses
- `auth/src/services/oauth_state.py` — `OAuthStateError` and 3 subclasses
- `auth/src/auth/jwt_handler.py` — `JWTError` and 2 subclasses
- `auth/src/charter/exceptions.py` — `CharterException` and 6 subclasses

### Error Propagation

**Lambda services**: Errors propagate up to the handler, which catches `Exception` generically, logs with `log.exception("event_name", error=str(e))`, and re-raises to trigger Lambda retry/DLQ. Domain errors are caught and converted to appropriate `FetchError` / `ReportError` at orchestrator boundaries.

**FastAPI services**: Domain errors are caught in route handlers and converted to `HTTPException` using `raise HTTPException(status_code=e.http_status, detail=e.to_dict()) from e`. The `from e` chaining is consistent.

```python
# ads/src/autom8_ads/api/launch.py pattern:
except AdsValidationError as e:
    raise HTTPException(status_code=422, detail=e.to_dict()) from e
except AdsError as e:
    logger.error("launch_error", extra={...})
    raise HTTPException(status_code=500, detail=e.to_dict()) from e
```

Note the catch order: specific subclass → base class. This is consistent throughout FastAPI routes.

### Error Handling at Boundaries

**No global exception handlers** are registered (`add_exception_handler` / `@app.exception_handler` matches zero files). All exception-to-HTTP translation is done per-route in try/except blocks.

**Circuit breaker pattern**: Several services use a circuit breaker (`CircuitBreakerOpenError` from `autom8y_http`) that wraps `DataServiceClient` calls. On open, the service logs and either re-raises or falls back to staging:
- `pull-payments/src/pull_payments/clients/data_service.py`
- `reconcile-spend/src/reconcile_spend/clients/data_service.py`

**Auth service exception pattern**: Auth service has its own `configure_service_logging` and uses Python's stdlib `logging.getLogger(__name__)` alongside `get_logger` from `autom8y_log`. This is the one service that mixes patterns, as it predates the SDK adoption.

### Error Logging Style

Structured log events use **snake_case event names as the first positional argument**:

```python
log.info("handler_invoked", days_back=days_back)
log.exception("sync_failed", error=str(e))
log.warning("missing_office_phone")
```

Lambda services use `autom8y_log.get_logger(__name__)` (variable named `log`, not `logger`). FastAPI services and devconsole use `logger = get_logger(__name__)`. The auth service is the notable outlier using stdlib logging in some files.

---

## File Organization

### Service Archetypes

Two distinct service archetypes exist:

**Archetype 1: Lambda Worker Service** (reconcile-*, account-status-recon, pull-payments, sms-performance-report, slack-alert, auth-mysql-sync)

Canonical file layout:
```
src/<package_name>/
    __init__.py          # version import + PackageNotFoundError guard
    __main__.py          # CLI/local entrypoint (some services)
    config.py            # Settings class + get_settings() + clear_settings_cache()
    errors.py            # Exception hierarchy (ServiceBaseError + subclasses)
    handler.py           # lambda_handler(@instrument_lambda) — entry point
    orchestrator.py      # async run_* function, coordinates full flow
    fetcher.py           # IO: upstream API calls (reconcile services)
    joiner.py            # Pure logic: data joining/correlation (reconcile services)
    rules.py             # Pure logic: anomaly detection / business rules
    models.py            # Pydantic BaseModel domain types
    metrics.py           # CloudWatch metric emission (record_side_effect)
    readiness.py         # Pipeline freshness gate
    report.py            # Slack Block Kit report builder
    clients/
        data_service.py  # DataServiceClient(BaseDataServiceClient)
        models.py        # Pydantic models for upstream API responses
```

Evidence: `reconcile-spend/src/reconcile_spend/`, `account-status-recon/src/account_status_recon/`, `reconcile-ads/src/reconcile_ads/`, `pull-payments/src/pull_payments/`

**Archetype 2: FastAPI Service** (ads, auth)

```
src/<package_name>/
    __init__.py
    app.py               # create_app() + lifespan()
    config.py            # Settings
    errors.py            # Exception hierarchy
    dependencies.py      # FastAPI Depends() factories (get_*())
    api/
        health.py        # GET /health router
        launch.py        # Route-specific router
    models/
        enums.py
        base.py          # Shared BaseModel base
        <domain>.py      # Domain-specific models
    <domain>/            # Domain logic subpackage
        service.py       # Core business logic class
        ...
    platforms/
        protocol.py      # Protocol definitions
    clients/
        <service>.py     # External HTTP clients
```

Evidence: `ads/src/autom8_ads/`, `auth/src/`

**Archetype 3: Auth-mysql-sync** (non-standard)

Does not follow either archetype. Uses a flat `src/` root (no sub-package) with sub-packages:
```
src/
    main.py
    config.py
    sync/                # orchestrator, reader, writer, transformer, guid_converter
    portover/            # cli, handler
    observability/       # logger, metrics
```

### Package Naming

Service directories use `kebab-case` (e.g., `reconcile-spend`). Python package names inside `src/` use `snake_case` (e.g., `reconcile_spend`). The mapping is consistent across all 11 services.

Service Python package names follow the pattern `autom8_<service_abbreviation>` for FastAPI services (e.g., `autom8_ads`, `autom8_devconsole`) and `<domain>_<qualifier>` for Lambda services (e.g., `reconcile_spend`, `pull_payments`, `account_status_recon`).

### Separation Rules

- **One responsibility per file**: `config.py` only defines settings, `errors.py` only defines exceptions, `handler.py` only defines `lambda_handler`, `orchestrator.py` contains the async coordination function.
- **No circular imports**: Inter-module dependencies flow handler → orchestrator → fetcher/rules/report.
- **Generated code**: None observed. `py.typed` marker file exists in `ads/src/autom8_ads/py.typed` to mark the package as typed.
- **Test layout mirrors src**: `tests/` directories mirror the `src/` sub-package structure with `test_<module>.py` naming.

### Configuration Singleton

All Lambda services provide a `get_settings()` cached factory and a `clear_settings_cache()` function at the module level in `config.py`. The pattern is verbatim identical across reconcile-spend, reconcile-ads, account-status-recon, pull-payments, and sms-performance-report:

```python
@lru_cache
def get_settings() -> Settings:
    return Settings()  # type: ignore[call-arg]

def clear_settings_cache() -> None:
    get_settings.cache_clear()
    Autom8yBaseSettings.reset_resolver()
```

Evidence: `reconcile-spend/src/reconcile_spend/config.py:144-153`, `reconcile-ads/src/reconcile_ads/config.py:123-132`

---

## Domain-Specific Idioms

### SDK Stack

Services depend on a set of internal `autom8y_*` packages published to AWS CodeArtifact. These are not PyPI packages. Key SDK packages:

| SDK Package | Purpose | Entry Points |
|---|---|---|
| `autom8y_config` | Pydantic-based settings with secret resolution | `Autom8yBaseSettings`, `Autom8yEnvironment`, `LambdaServiceSettingsMixin` |
| `autom8y_telemetry` | OpenTelemetry instrumentation | `instrument_lambda`, `record_side_effect`, `init_telemetry`, `TelemetryContext` |
| `autom8y_log` | Structured logging | `get_logger(__name__)`, `configure_logging()` |
| `autom8y_http` | HTTP resilience | `CircuitBreakerOpenError`, `BaseDataServiceClient` |
| `autom8y_interop` | Cross-service clients | `DataInsightClient`, `DataPaymentProtocol` |
| `autom8y_slack` | Slack posting | `SlackClient` |
| `autom8y_events` | EventBridge publishing | `DomainEvent`, `EventPublisher` |
| `autom8y_telemetry.conventions` | OTel span attribute name constants | `RECONCILIATION_METRICS_EMITTED`, `RECONCILIATION_CHANNEL`, etc. |

### Settings Pattern: `Autom8yBaseSettings`

All Lambda services use a common MRO: `Settings(LambdaServiceSettingsMixin, Autom8yBaseSettings)`. This is not optional — using bare `BaseSettings` is non-compliant (ads service uses bare `BaseSettings`, which is noted as divergent).

The `_SERVICE_KEY_ALIAS` class variable is used to declare the service-specific secret name. The `service_api_key` field uses `AliasChoices` to accept both the service-specific name and `SERVICE_API_KEY` (the generic canonical):

```python
service_api_key: SecretStr = Field(
    validation_alias=AliasChoices("RECONCILE_SPEND_SERVICE_KEY", "SERVICE_API_KEY"),
)
```

Evidence: `reconcile-spend/src/reconcile_spend/config.py:54-57`, `reconcile-ads/src/reconcile_ads/config.py:58-61`

### Lambda Instrumentation Idiom

Lambda handler functions are decorated with `@instrument_lambda` from `autom8y_telemetry.aws`. The handler wraps `asyncio.run(orchestrator_function())`. The pattern is:

```python
@instrument_lambda
def lambda_handler(event: dict[str, Any] | None = None, context: Any = None) -> dict[str, Any]:
    event = event or {}
    ...
    try:
        result = asyncio.run(run_reconciliation(period_days))
        ...
        return {"statusCode": 200, "body": json.dumps({...})}
    except Exception as e:
        log.exception("event_name_failed", error=str(e))
        raise  # trigger Lambda retry/DLQ
```

Evidence: `reconcile-spend/src/reconcile_spend/handler.py`, `pull-payments/src/pull_payments/handler.py`, `reconcile-ads/src/reconcile_ads/handler.py`, `account-status-recon/src/account_status_recon/handler.py`, `slack-alert/src/slack_alert/handler.py`

### `record_side_effect` Telemetry

Side-effectful operations (Slack posts, CloudWatch emissions, S3 writes, EventBridge publishes) are annotated with `record_side_effect(span, system=..., operation=..., target=..., payload=...)`. This enables the devconsole mutation summary panel to display side effects.

Evidence: `pull-payments/src/pull_payments/handler.py:60-66`, `reconcile-spend/src/reconcile_spend/orchestrator.py`

### Structured Logging Convention

Log event names are `snake_case` verb phrases describing what happened, not what is happening:
- `"handler_invoked"`, `"sync_failed"`, `"reconciliation_started"`, `"event_published"`, `"cache_cleared"`

Log calls pass keyword arguments for structured context, not f-strings:
- `log.info("handler_invoked", days_back=days_back)` not `log.info(f"handler invoked with {days_back}")`

### `from __future__ import annotations`

Present in 146 of the observed source files. This is universal across the codebase for all non-trivial source files. Not present in test files uniformly.

### Protocol-Based Interfaces

The `ads` service uses `typing.Protocol` with `@runtime_checkable` for adapter contracts rather than ABCs. The auth service uses `ABC` with `@abstractmethod` for OAuth adapters. Both patterns exist in the codebase.

Evidence: `ads/src/autom8_ads/platforms/protocol.py`, `auth/src/services/oauth_adapters/base.py`

### FastAPI Dependency Injection

FastAPI services use `Depends()` with factory functions prefixed `get_*` to inject dependencies. Dependencies are accessed from `request.app.state` in the `get_*` functions. Singletons (config, clients, caches) are initialized in the `lifespan()` context manager and stored in `app.state`:

```python
# ads/src/autom8_ads/app.py pattern
@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    app.state.launch_service = LaunchService(...)
    yield
    # teardown
```

Evidence: `ads/src/autom8_ads/app.py`, `ads/src/autom8_ads/dependencies.py`

### Telemetry Tracer Initialization

Services that emit OTel spans initialize a module-level tracer with a namespaced name:
```python
_tracer = trace.get_tracer("autom8y.reconciliation")
```
The underscore prefix on `_tracer` signals it is a module-private singleton.

Evidence: `reconcile-spend/src/reconcile_spend/handler.py:17`, `pull-payments/src/pull_payments/handler.py:25`

---

## Naming Patterns

### File Names

| Pattern | Convention | Examples |
|---|---|---|
| Service modules | `snake_case.py` | `orchestrator.py`, `data_service.py`, `auth_writer.py` |
| Error modules | Always `errors.py` | All Lambda services |
| Config modules | Always `config.py` | All services |
| Handler modules | Always `handler.py` | All Lambda services |
| Test files | `test_<subject>.py` | `test_orchestrator.py`, `test_errors.py` |

### Class Names

All classes use `PascalCase`. No abbreviation exceptions observed.

| Category | Pattern | Examples |
|---|---|---|
| Service base errors | `{ServiceName}Error(Exception)` | `ReconcileSpendError`, `AdsError`, `AccountStatusReconError` |
| Specific errors | `{Concept}Error({BaseError})` | `FetchError`, `CircuitBreakerError`, `LaunchInProgressError` |
| Settings | `Settings(LambdaServiceSettingsMixin, Autom8yBaseSettings)` | Identical name `Settings` in all Lambda services |
| FastAPI settings | `{ServicePascal}Settings(Autom8yBaseSettings)` | `DevconsoleSettings`, `AdsConfig` (inconsistency: ads uses `AdsConfig`) |
| Services/business logic | `{Domain}Service` | `LaunchService`, `CredentialVaultService`, `APIKeyService` |
| Clients | `{Domain}Client` | `DataServiceClient`, `TempoClient`, `AuthenticatedHTTPClient` |
| Orchestrators | `{Domain}Orchestrator` | `SyncOrchestrator` |
| Protocol interfaces | `{Domain}Protocol` | `AdPlatform`, `DataServiceProtocol`, `CircuitBreaker` |

### Variable Names

- Logger variable: `log = get_logger(__name__)` in Lambda services; `logger = get_logger(__name__)` in FastAPI/devconsole
- Tracer variable: `_tracer = trace.get_tracer(...)` (underscore prefix, module-private)
- Settings instance: retrieved via `get_settings()`, not stored at module level in Lambda services
- Router variable (FastAPI): `router = APIRouter(prefix="...")`

### Function Names

| Pattern | Convention | Examples |
|---|---|---|
| Async orchestrator entry | `run_{noun}` or `sync_{noun}` | `run_reconciliation()`, `sync_payments()` |
| Settings factory | `get_settings()` | Identical across all Lambda services |
| Settings cache reset | `clear_settings_cache()` | Identical across all Lambda services |
| FastAPI app factory | `create_app()` | `ads/src/autom8_ads/app.py`, `devconsole` |
| Dependency factories | `get_{resource}()` | `get_launch_service()`, `get_idempotency_cache()` |

### Module/Package Names

- Service Python packages: `snake_case` matching the kebab-case service directory
- Internal SDK packages: `autom8y_<name>` (underscore, not hyphen in Python)
- PyPI package names use hyphens: `autom8y-config`, `autom8y-telemetry`

### Constants

Module-level constants use `SCREAMING_SNAKE_CASE`. In Lambda handlers, a `NAMESPACE` constant declares the CloudWatch metric namespace:
```python
NAMESPACE = "Autom8y/PullPayments"
```

Error `code` values also use `SCREAMING_SNAKE_CASE` strings: `"FETCH_ERROR"`, `"RECONCILE_SPEND_ERROR"`.

### Naming Inconsistencies Observed

1. **Logger variable name**: `log` vs `logger` split along service archetype lines (not a random inconsistency — Lambda services use `log`, FastAPI/devconsole use `logger`).
2. **Settings class name**: Lambda services all name it `Settings`; ads uses `AdsConfig`; devconsole uses `DevconsoleSettings`.
3. **ads package name**: `autom8_ads` (one `8`) vs convention `autom8y_ads` — the service directory is `ads` with Python package `autom8_ads`.

---

## Knowledge Gaps

- **auth service** diverges from Lambda SDK patterns in several places (stdlib logging, no `errors.py`, pre-dates `autom8y_log`/`autom8y_config` adoption). Its conventions may not be representative for new services.
- **sms-performance-report** uses the canonical Lambda archetype but its `src/` layout was not fully read (only config confirmed).
- **devconsole** is a TUI/development tool, not a production service; its conventions (NiceGUI app framework, persistence layer) are not representative for Lambda or API services.
- **Error handling in auth client** (`auth/client/`) follows an independent exception hierarchy (`AuthError` base class) without `ClassVar code` pattern — this is a published SDK, not a service.
- The `autom8y_interop` lazy-import pattern in `reconcile-ads/src/reconcile_ads/fetcher.py` (importing inside try blocks with `# noqa: B904`) was observed but not fully investigated.
