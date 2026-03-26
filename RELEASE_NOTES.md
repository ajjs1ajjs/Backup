# Release Notes

## v0.4.0 (Unreleased)

### Features
- Stage 4: Docker/CI production readiness added (Dockerfile, docker-compose.yml, CI workflow).
- Stage 3: Interactive frontend dashboard with login/logout, VM and backup listing, and cloud backup creation form.
- Stage 2: Real cloud provider integration (AWS/Azure/GCP) with CloudOrchestrator and mock provider for CI.
- Stage 1: RBAC/OAuth2/JWT skeleton with token refresh and in-memory store.

### Bug Fixes
- Fixed token refresh endpoint to accept the expected JSON structure.
- Added provider field to Azure and GCP restore responses to match expected schema.
- Updated CloudOrchestrator to select provider by exact class name match.

### Documentation
- Added MIGRATION_GUIDE.md for migrating from JSON store to SQL database.
- Updated README with deployment instructions and environment variable descriptions.

### Dependencies
- Added optional dependencies for cloud providers (boto3, azure-mgmt-compute, google-api-python-client) and database backends (SQLAlchemy, psycopg2-binary, pyodbc, cx_Oracle).
- Added development dependencies for testing (pytest, moto) and code quality (ruff, mypy, flake8) via pyproject.toml.

## v0.3.0
- Initial implementation of Stages 1-3.

## v0.2.0
- Added cloud orchestration and provider skeletons.

## v0.1.0
- Initial MVP with RBAC/OAuth2/JWT skeleton and API.
