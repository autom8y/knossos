# Playbook: API Design & Implementation

> Design-first approach to building service interfaces

## When to Use

- Creating new API endpoints
- Designing service-to-service interfaces
- Building public or partner APIs
- Adding REST/GraphQL/gRPC interfaces
- Versioning existing APIs

## Prerequisites

- Clear understanding of consumers (who will call this API)
- Authentication/authorization requirements identified
- Performance and rate limiting expectations

## Command Sequence

### Phase 1: Initialize

```bash
/10x
```
**Expected output**: Team switched to 10x-dev-pack, roster displayed
**Decision point**: If simple CRUD endpoint, consider `/hotfix`.

### Phase 2: Start Session

```bash
/start "API name/scope" --complexity=SERVICE
```
**Expected output**: Session created, Requirements Analyst invoked
**Decision point**: Adjust complexity level:
- MODULE: Single endpoint or small feature
- SERVICE: Full resource or service interface
- PLATFORM: Cross-cutting API infrastructure

### Phase 3: Requirements

Requirements Analyst captures API contract.

**Expected output**: PRD with consumer needs, use cases, constraints
**Decision point**: Validate with API consumers before design.

### Phase 4: Design

```bash
/architect
```
**Expected output**: TDD with API contract, data models, ADRs for design choices
**Decision point**: Review contract with consumers. Common ADRs:
- REST vs GraphQL vs gRPC
- Versioning strategy (URL vs header)
- Authentication mechanism
- Error response format

### Phase 5: Implementation

```bash
/build
```
**Expected output**: API implementation with tests, OpenAPI spec if applicable
**Decision point**: If blocked, use `/handoff` to clarify requirements.

### Phase 6: Validation

```bash
/qa
```
**Expected output**: Test report including contract tests, load testing results
**Decision point**: Verify against design contract.

### Phase 7: Documentation

Consider routing to doc-team if external API:
```bash
/wrap
# Then if external:
/docs
/start "API documentation"
```
**Expected output**: Session summary, API documentation

### Phase 8: Ship

```bash
/pr
```
**Expected output**: Pull request with API changes

## Variations

- **Internal API**: Skip formal documentation phase
- **Breaking changes**: Requires versioning ADR, migration path
- **GraphQL**: Emphasize schema design in architect phase
- **External/public API**: Add doc-team handoff for reference docs

## Success Criteria

- [ ] API contract defined and approved
- [ ] Implementation matches contract
- [ ] Tests cover happy path and error cases
- [ ] Contract tests validate backwards compatibility
- [ ] Documentation complete (for external APIs)

## Rollback

If API design needs revision:
```bash
/park
# Gather additional consumer feedback
/continue
/handoff architect   # Revise design
```
