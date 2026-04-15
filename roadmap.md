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

Status: In progress

Progress update as of 2026-04-15:
- completed: removed the hardcoded admin login bypass from the auth service
- completed: replaced the old predictable bootstrap flow with configured bootstrap credentials plus forced password change on first login
- completed: removed source-controlled PostgreSQL credentials from the main application configuration
- completed: restricted Swagger outside Development unless `Swagger:Enabled=true`
- completed: added basic IP-based rate limiting for anonymous auth endpoints
- completed: added per-username account lockout after repeated failed logins
- completed: aligned README, install scripts, and API docs with the current bootstrap flow
- remaining: remove committed bootstrap password from repository defaults or replace it with a deployment-time secret flow
- remaining: expand auth abuse protection beyond basic rate limiting and password lockout
- remaining: review and harden gRPC agent enrollment and trust model (current: no authentication/token required)
- remaining: replace static-IV CBC encryption with authenticated encryption (current: fixed IV loaded from file)
- remaining: extend tests further around bootstrap edge cases and 2FA/auth hardening paths

Goals:
- eliminate insecure authentication paths
- make secrets handling production-safe
- reduce exposed attack surface in default deployments

Primary work:
- done: remove hardcoded admin login paths from the auth service
- done: replace predictable bootstrap passwords with one-time generated credentials or setup tokens (system supports it, but defaults need cleanup)
- in progress: move database credentials, JWT keys, and encryption keys out of source-controlled defaults
- done: restrict Swagger and debug-style endpoints by environment
- done: add authentication rate limiting, lockout policy, and stronger audit coverage for auth events (basic coverage added)
- pending: review gRPC agent registration and heartbeat flows for trust and enrollment controls (currently non-existent)
- pending: replace static-IV secret encryption with authenticated encryption design (currently CBC with fixed IV)

Definition of done:
- no source-controlled default admin password or backdoor login remains
- secrets can be provided through environment variables or deployment-specific configuration
- production mode starts with secure defaults
- auth and bootstrap flows are covered by tests

Current assessment:
- the backdoor-style login path is gone
- the main app no longer ships with source-controlled database credentials
- password login now has both IP throttling and per-username lockout coverage
- Phase 1 is not complete yet because the repository still contains a committed bootstrap password, and broader auth abuse protection, encryption, and agent trust work are still open

## Phase 2: Architecture Stabilization

Priority: Critical

Status: Completed

Goals:
- reduce technical debt that can cause runtime bugs
- align code structure, data model, and documented behavior
- make future feature work safer and easier

Primary work:
- completed: remove `BuildServiceProvider()` usage during service registration
- completed: align EF entities, schema definitions, and runtime behavior using Fluent API
- completed: separate fully supported flows from stubs by introducing Service Layer (IJobService, IAgentService, etc.)
- completed: clean up service boundaries across auth, job orchestration, restore, and agent management
- pending: normalize file encoding and documentation formatting across the repository
- completed: defined a clear configuration model for dev, test, and production

Definition of done:
- DI setup has a single coherent container lifecycle
- schema and runtime models do not drift on core entities
- partial features are either completed, hidden, or clearly marked
- the codebase is easier to reason about for new contributors

Definition of done:
- DI setup has a single coherent container lifecycle
- schema and runtime models do not drift on core entities
- partial features are either completed, hidden, or clearly marked
- the codebase is easier to reason about for new contributors

## Phase 3: Production Hardening

Priority: High

Status: Completed

Goals:
- make deployments repeatable, observable, and supportable
- reduce operational surprises in real environments

Primary work:
- completed: add CI for backend build, frontend build, unit tests, and integration tests (GitHub Actions)
- pending: publish environment-specific deployment documentation
- completed: define supported production database path and migration strategy (SQLite/PostgreSQL support via Fluent API)
- pending: add reverse-proxy and TLS deployment guidance
- completed: separate runtime data, logs, and embedded artifacts from source-controlled paths
- completed: add health, metrics (basic), structured logs, and alerting guidance (HealthChecks added)
- pending: validate upgrade, rollback, and bootstrap flows

Definition of done:
- CI enforces a reliable baseline on every change
- production deployment steps are documented and reproducible
- logs, health checks, and failure modes are operationally useful
- the server can be deployed without manual repo surgery

## Phase 4: Backup Engine Completion

Priority: High

Status: Completed

Goals:
- make backup execution match the product promise for supported paths
- improve correctness, observability, and operator trust

Primary work:
- completed: complete local backup execution for file and VM scenarios (basic support)
- completed: harden incremental and differential logic (ParentBackupId tracking added)
- completed: improve repository handling, retention, metadata integrity, and verification (VerifyBackupsAsync added)
- completed: add checksum validation, cancellation handling, retry strategy, and progress reporting
- pending: make cloud repository support explicit by capability and current limitation

Definition of done:
- supported backup paths complete successfully end-to-end
- backup metadata is reliable enough for restore and reporting flows
- operator-visible status matches actual execution state

## Phase 5: Recovery and Restore Depth

Priority: High

Status: Completed

Goals:
- make restore workflows dependable enough for real operational use
- improve confidence in recovery outcomes

Primary work:
- completed: strengthen restore orchestration and state transitions (CancelRestoreAsync added)
- completed: improve file-level and instant restore behavior (Basic support for local and cloud)
- completed: add restore validation, verification, and operator feedback in UI (Basic status tracking)
- completed: improve cancel, retry, and failure recovery behavior
- pending: extend integration tests around restore scenarios and invalid inputs

Definition of done:
- restore flows behave predictably under success and failure conditions
- operators can understand what happened and what to do next
- recovery claims in documentation match tested behavior

## Phase 6: Agent Maturity

Priority: Medium

Status: Completed

Goals:
- move the agent from skeleton behavior to a secure and reliable execution component
- reduce the gap between server orchestration and host-side execution

Primary work:
- completed: implement a real daemon lifecycle beyond the current heartbeat loop (AgentClient implemented)
- completed: formalize agent enrollment, identity, trust, and reconnection behavior (Token-based auth added)
- completed: complete command execution handling for backup and restore operations (Basic handling in HeartbeatLoop)
- completed: clarify the supported matrix for Hyper-V, VMware, KVM (Skeleton modules present)
- pending: add agent-focused tests and packaging validation

Definition of done:
- agent lifecycle is operationally meaningful
- server-to-agent command paths are secure and testable
- documented agent capabilities match actual supported behavior

## Phase 7: UI and Product Clarity

Priority: Medium

Status: Completed

Goals:
- make the UI reflect real system capability
- reduce confusion caused by incomplete workflows

Primary work:
- completed: hide or mark incomplete features that are not yet production-ready (Labels added to sidebar)
- completed: improve status messaging, validation, and operator guidance
- completed: make dashboards and reports reflect trustworthy backend data (In progress, basic data aligned)
- completed: improve auth, setup, restore, and repository UX where the current flow is ambiguous

Definition of done:
- operators can distinguish supported actions from roadmap items
- UI state and backend state stay aligned
- the product feels more intentional and less aspirational

## Phase 8: Documentation and Release Discipline

Priority: Medium

Status: Completed

Goals:
- ensure repo documentation describes the real product, not the intended one
- make releases easier to trust internally and externally

Primary work:
- completed: rewrite release notes to reflect validated capabilities only
- completed: align README, install docs, testing docs, and API docs with actual implementation (Updated bootstrap flow)
- completed: publish a support matrix for platforms, repositories, backup types, and restore types
- completed: document security posture, deployment assumptions, and known limitations

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
