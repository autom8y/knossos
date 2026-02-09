---
paths:
  - "internal/provenance/**"
---

When modifying files in internal/provenance/:
- Collector interface is thread-safe; defaultCollector uses sync.Mutex
- NullCollector is the no-op implementation for dry-run mode
- Save() uses structurallyEqual() to skip writes when only timestamps change (CC file watcher safety)
- Checksums use sha256: prefix format per ADR-0026
- Divergence detection promotes knossos→user on checksum mismatch; user entries carry forward unchanged
- Manifest validation enforces: schema_version (N.N), non-zero timestamps, valid owner enum, sha256 checksums
- One-way dependency: materialize imports provenance, never the reverse
