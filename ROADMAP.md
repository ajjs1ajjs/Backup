# Roadmap for Novabackup (Python MVP to Production)

Overview
- This document outlines the plan to evolve the current MVP into a production-ready, cloud-enabled backup/restore platform with multi-cloud support (AWS, Azure, Google Cloud), RBAC/OAuth2/JWT authentication, a modern frontend UI, Docker/Compose CI, and migration tooling.

Status snapshot (as of now)
- Cloud scaffolding exists with AWS/Azure/GCP provider skeletons and a cloud orchestrator. Real SDK calls are wired where possible; credentials are required for production use.
- RBAC/OAuth2/JWT scaffolding implemented; token endpoints in API exist; token refresh flow is conceptual and partially implemented with in-memory store for MVP.
- SA DB and migration tooling in place; tests cover migration JSON -> DB, and multiple DB dialects (Postgres, MSSQL, Oracle).
- CLI and API exist with core endpoints: list VMs, normalize VM types, backup creation, listing, and restore (cloud path supported in MVP).
- Frontend/UI scaffold added: dashboard and login flow skeleton.
- Docker/CI scaffolding created: docker-compose-prod.yml, production installers for Linux/Windows, systemd service, and CI flows.
- Documentation: README updated with deployment and migration guidance; migration guide placeholders included.

Phases (detailed plan)
- Phase 1 — Core security and API tokens (RBAC/OAuth2/JWT)
  - Implement full OAuth2 flow with refresh tokens, token rotation, scopes, and roles (admin, user)
  - Harden /token and /token/refresh endpoints; ensure token expiry and rotation rules
  - Integrate RBAC into all sensitive endpoints (backups create/restore restricted to admin or specific roles)
  - Add tests for login, token refresh, and RBAC enforcement

- Phase 2 — Real cloud SDK integrations (AWS/Azure/GCP)
  - Implement AWS: list VM via EC2, backup via EBS snapshot, restore flow
  - Implement Azure: list VMs, snapshot/restore using azure-mgmt-compute
  - Implement GCP: list Compute Engine instances, snapshotting/disks
  - Extend cloud orchestrator to coordinate providers; support multiple providers with a unified API
  - Update CI to cover cloud flows with mock credentials in CI or dedicated test accounts

- Phase 3 — Frontend/UI
  - Build a modern React/Vue frontend; login screen with token handling; dashboard with filters and status
  - UI connected to API endpoints with RBAC enforced on frontend as well

- Phase 4 — CI/CD & Docker
  - Expand docker-compose to include a full prod stack: DB, API, UI
  - Extend GH Actions with multi-DB testing (Postgres, MSSQL, Oracle) and cloud tests (MOCK + real in staging)
  - Implement code quality gates: lint, type checks, security checks

- Phase 5 — Migration docs and release process
  - Provide an explicit MIGRATION_GUIDE for backups.json → DB with steps for all dialects
  - Release notes, versioning strategy, deprecation policy

- Phase 6 — Observability & Audit
  - Add logging, metrics (Prometheus/OpenTelemetry), and audit trail for critical actions

Milestones
- v0.1 MVP: current MVP with CLI + API + SA DB + cloud scaffolding (no real cloud calls yet in prod)
- v0.2 Cloud integration (real SDK calls) + RBAC refactoring
- v0.3 UI & advanced dashboard + Docker prod stack
- v1.0 Production ready: RBAC, audit, observability, and full CI/CD

Risks and mitigations
- Risk: Credentials exposure in CI
  - Mitigation: Use GitHub Secrets; never store secrets in repo
- Risk: Cloud provider API changes
  - Mitigation: Abstract provider interface; wrap real calls with error handling, feature flags
- Risk: RBAC complexity
  - Mitigation: Start with admin/user roles; gradually add more granular roles; include tests

Next steps for the team
- Validate cloud provider integration in a staging/CI environment using cloud test accounts or mocks
- Implement RBAC in UI and API consistently
- Add a robust migration script and comprehensive README migration guide
- Improve dashboard UX and add metrics logging
