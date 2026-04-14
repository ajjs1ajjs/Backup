# Roadmap

## Purpose

This roadmap reflects the current codebase more accurately than a feature wishlist.
Its goal is to move the project from a broad MVP into a secure, testable, and production-capable backup platform.

## Current Reality

Working foundations already exist:
- .NET 8 server with REST API, auth, background services, EF Core, and Swagger
- React UI for core operator flows
- integration and unit test coverage for selected backend paths
- C++ agent skeleton with hypervisor-specific modules

Current blockers and gaps:
- security-critical auth shortcuts and predictable bootstrap credentials
- partial mismatch between documented features and implemented behavior
- incomplete agent/runtime orchestration
- production hardening and CI still need maturity
- several features are present as partial implementations or stubs

## Guiding Priorities

1. Remove security blockers before adding more surface area.
2. Stabilize architecture and configuration before scaling features.
3. Make production behavior explicit and testable.
4. Expand backup and recovery depth only after the platform is trustworthy.

## Phase 1: Security Baseline

Priority: Critical

Goals:
- eliminate insecure authentication paths
- make secrets handling production-safe
- reduce exposed attack surface in default deployments

Primary work:
- remove hardcoded admin login paths from the auth service
- replace predictable bootstrap passwords with one-time generated credentials or setup tokens
- move database credentials, JWT keys, and encryption keys out of source-controlled defaults
- restrict Swagger and debug-style endpoints by environment
- add authentication rate limiting, lockout policy, and stronger audit coverage for auth events
- review gRPC agent registration and heartbeat flows for trust and enrollment controls
- replace static-IV secret encryption with authenticated encryption design

Definition of done:
- no source-controlled default admin password or backdoor login remains
- secrets can be provided through environment variables or deployment-specific configuration
- production mode starts with secure defaults
- auth and bootstrap flows are covered by tests

## Phase 2: Architecture Stabilization

Priority: Critical

Goals:
- reduce technical debt that can cause runtime bugs
- align code structure, data model, and documented behavior
- make future feature work safer and easier

Primary work:
- remove `BuildServiceProvider()` usage during service registration
- align EF entities, schema definitions, enums, and runtime behavior
- separate fully supported flows from stubs and partial implementations
- clean up service boundaries across auth, job orchestration, restore, notifications, and agent management
- normalize file encoding and documentation formatting across the repository
- define a clear configuration model for dev, test, and production

Definition of done:
- DI setup has a single coherent container lifecycle
- schema and runtime models do not drift on core entities
- partial features are either completed, hidden, or clearly marked
- the codebase is easier to reason about for new contributors

## Phase 3: Production Hardening

Priority: High

Goals:
- make deployments repeatable, observable, and supportable
- reduce operational surprises in real environments

Primary work:
- add CI for backend build, frontend build, unit tests, and integration tests
- publish environment-specific deployment documentation
- define supported production database path and migration strategy
- add reverse-proxy and TLS deployment guidance
- separate runtime data, logs, and embedded artifacts from source-controlled paths
- add health, metrics, structured logs, and alerting guidance
- validate upgrade, rollback, and bootstrap flows

Definition of done:
- CI enforces a reliable baseline on every change
- production deployment steps are documented and reproducible
- logs, health checks, and failure modes are operationally useful
- the server can be deployed without manual repo surgery

## Phase 4: Backup Engine Completion

Priority: High

Goals:
- make backup execution match the product promise for supported paths
- improve correctness, observability, and operator trust

Primary work:
- complete local backup execution for file and VM scenarios that are already exposed in UI/API
- harden incremental and differential logic where currently partial
- improve repository handling, retention, metadata integrity, and verification
- add checksum validation, cancellation handling, retry strategy, and progress reporting
- make cloud repository support explicit by capability and current limitation

Definition of done:
- supported backup paths complete successfully end-to-end
- backup metadata is reliable enough for restore and reporting flows
- operator-visible status matches actual execution state

## Phase 5: Recovery and Restore Depth

Priority: High

Goals:
- make restore workflows dependable enough for real operational use
- improve confidence in recovery outcomes

Primary work:
- strengthen restore orchestration and state transitions
- improve file-level and instant restore behavior where partially implemented
- add restore validation, verification, and operator feedback in UI
- improve cancel, retry, and failure recovery behavior
- extend integration tests around restore scenarios and invalid inputs

Definition of done:
- restore flows behave predictably under success and failure conditions
- operators can understand what happened and what to do next
- recovery claims in documentation match tested behavior

## Phase 6: Agent Maturity

Priority: Medium

Goals:
- move the agent from skeleton behavior to a secure and reliable execution component
- reduce the gap between server orchestration and host-side execution

Primary work:
- implement a real daemon lifecycle beyond the current heartbeat loop
- formalize agent enrollment, identity, trust, and reconnection behavior
- complete command execution handling for backup and restore operations
- clarify the supported matrix for Hyper-V, VMware, KVM, and database-related agent behavior
- add agent-focused tests and packaging validation

Definition of done:
- agent lifecycle is operationally meaningful
- server-to-agent command paths are secure and testable
- documented agent capabilities match actual supported behavior

## Phase 7: UI and Product Clarity

Priority: Medium

Goals:
- make the UI reflect real system capability
- reduce confusion caused by incomplete workflows

Primary work:
- hide or mark incomplete features that are not yet production-ready
- improve status messaging, validation, and operator guidance
- make dashboards and reports reflect trustworthy backend data
- improve auth, setup, restore, and repository UX where the current flow is ambiguous

Definition of done:
- operators can distinguish supported actions from roadmap items
- UI state and backend state stay aligned
- the product feels more intentional and less aspirational

## Phase 8: Documentation and Release Discipline

Priority: Medium

Goals:
- ensure repo documentation describes the real product, not the intended one
- make releases easier to trust internally and externally

Primary work:
- rewrite release notes to reflect validated capabilities only
- align README, install docs, testing docs, and API docs with actual implementation
- publish a support matrix for platforms, repositories, backup types, and restore types
- document security posture, deployment assumptions, and known limitations

Definition of done:
- no major mismatch remains between docs and code
- release notes describe tested capabilities
- support boundaries are visible to users and maintainers

## Recommended Execution Order

1. Phase 1: Security Baseline
2. Phase 2: Architecture Stabilization
3. Phase 3: Production Hardening
4. Phase 8: Documentation and Release Discipline
5. Phase 4: Backup Engine Completion
6. Phase 5: Recovery and Restore Depth
7. Phase 6: Agent Maturity
8. Phase 7: UI and Product Clarity

## Suggested Milestones

### Milestone A: Safe Internal Build
- Phase 1 complete
- most of Phase 2 complete
- CI baseline started

### Milestone B: Production Candidate
- Phases 1 through 3 complete
- Phase 8 aligned with actual behavior
- core backup and restore paths validated

### Milestone C: Feature Expansion
- Phases 4 through 7 advanced without undermining platform trust

## What Not To Do Yet

- do not market unsupported backup types as complete
- do not broaden agent/platform promises before secure enrollment exists
- do not add more UI surface for features that are still backend stubs
- do not optimize for breadth before security and production basics are solid
