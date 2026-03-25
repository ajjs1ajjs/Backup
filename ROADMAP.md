Roadmap: Novabackup Python MVP — Updated Plan

Status overview (as of now)
- Windows: Hyper-V provider enhanced for robust VM enumeration and mapping.
- Linux: Libvirt provider integrated; real data where available.
- DB backend: SQLite with backups and restores; optional shift from JSON to DB via NOVABACKUP_DATABASE_URL.
- API: FastAPI with endpoints /vms, /normalize, /backups, /backups/{backup_id}/restore; API keys scaffold.
- CLI: backup create/list/restore; integrated with DB/provider layers.
- UI: Minimal static dashboard at /static/dashboard.html; data from API.
- CI: GitHub Actions with pytest, mypy, flake8.
- Installers: install.sh (Linux/macOS) and install.bat (Windows).
- Tests: unit tests and integration tests; end-to-end tests added.

Phases (updated plan)
- Phase 1: Real Providers and DB migrations (2–3 weeks)
  - Stabilize Windows Hyper-V provider; finalize Libvirt integration
  - Implement DB migrations: from JSON to DB; add migration tooling
  - Expand tests for providers and DB
- Phase 2: API improvements and security (2–3 weeks)
  - Add API keys/auth, roles, and improved error handling
  - Extend data models (Snapshot, Restore); validation in Pydantic
  - Expand API surface: statuses, progress, more endpoints
  - Improve Swagger docs with examples
- Phase 3: End-to-End testing & UI (2–3 weeks)
  - End-to-end test suite: CLI <-> API <-> DB
  - Improve dashboard UX; optional frontend enhancements
- Phase 4: Packaging & Release readiness (2 weeks)
  - Packaging to wheel/PyPI; release notes and versioning
  - CI optimization; performance/stress tests
- Phase 5: Observability & resilience (ongoing)
  - Metrics (Prometheus/OpenTelemetry), logging, tracing

Milestones (high level)
- Milestone 0: MVP baseline ready (current state).
- Milestone 1: Phase 1 complete — real providers + DB migrations; tests updated.
- Milestone 2: Phase 2 complete — API hardening, docs, authentication.
- Milestone 3: Phase 3 complete — E2E tests and UI improvements.
- Milestone 4: Phase 4 release ready — packaging and CI.

Risks & mitigations
- Provider integration complexity: keep stable provider interfaces; use mocks in CI where needed.
- Windows Hyper-V CI constraints: rely on local lab or mocks; provide test harness.
- Data migration: provide safe migration scripts with dry-run mode and rollback notes.
- Security: API key in dev; plan to implement OAuth2 or token-based in future.

Success criteria
- End-to-end flow validated (CLI -> API -> DB).
- API available with documented endpoints and authentication.
- Packaging and CI stable across Linux/Windows.
- MVP features stable for production-like usage (VM listing, normalize, backups, restore).

Next steps
- Prepare a comprehensive PR description and user guide.
- Implement migration tooling from backups.json to DB with a safe fallback.
- Add a small set of sample providers (mock) to ensure CI stability across hardware.
