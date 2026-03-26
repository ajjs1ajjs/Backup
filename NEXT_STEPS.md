# Next Steps: Production‑Ready Plan

- Phase 1: RBAC and API token flow (completed baseline)
  - Finalize full OAuth2 with refresh tokens, rotation, scopes, and admin/user roles
  - Solidify token expiry handling and revocation; add tests for login and refresh flows
  - Harden error handling for API responses and retry logic

- Phase 2: Real cloud provider integrations (AWS, Azure, GCP)
  - Implement AWSCloudProvider real flows (list_vms via EC2, backup via EBS snapshots, restore)
  - Implement AzureCloudProvider real flows (VM list, snapshots and restores via Compute API)
  - Implement GCPCloudProvider real flows (Compute Engine instances, snapshots)
  - Extend CloudOrchestrator to coordinate multi-provider backups/restores
  - Expand tests for each provider using real credentials in CI or secure test accounts

- Phase 3: Frontend UI
  - Build a modern SPA (React/Vue/Next) with login, dashboards, filters, and status views
  - Integrate with API for live data and cloud backups

- Phase 4: Docker/CI production readiness
  - Complete docker-compose-prod.yml with DB + API + UI; secure secrets management
  - Expand GH Actions to test multi-DB dialects (PostgreSQL, MSSQL, Oracle) and cloud paths
  - Implement code quality gates (lint, type checks, security checks)

- Phase 5: Migration and Release docs
  - MIGRATION_GUIDE.md with step-by-step path for backups.json → DB across dialects
  - Release notes and versioning plan

- Phase 6: Observability and Audit
  - Logging, metrics (Prometheus/OpenTelemetry), and audit trails for critical actions

Milestones
- MVP: complete RBAC + API skeleton + SA DB + cloud scaffolding + prod scaffolding
- v1.0: full cloud SDK integration, RBAC, UI, Docker CI, and migration docs
