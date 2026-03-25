Novabackup - Modern Python MVP for VM backup/restore

Overview
- Cross-platform CLI + API for VM backup/restore
- Local DB backend (SQLite by default) with optional SQLAlchemy backends (PostgreSQL, MSSQL, Oracle)
- Cloud provider scaffolding (AWS, Azure, GCP) with Mock for CI; real SDKs pluggable
- Web API (FastAPI) with OpenAPI docs
- Lightweight dashboard at /static/dashboard.html
- CI with GitHub Actions; tests cover core, CLI, API, DB, and cloud scaffolding

Why this exists
- Provide a consistent flow for listing VMs, normalizing VM types, backing up, and restoring across local/cloud environments
- Enable migration from a JSON store to a DB-based store with minimal friction
- Provide a path to real cloud backups/restores using AWS/Azure/GCP SDKs

Key components (high level)
- core.py: VM discovery and normalization; cloud providers support via env var NOVABACKUP_CLOUD_PROVIDERS
- backup.py: BackupManager with local DB and cloud path support
- db_sa.py: SQLAlchemy backed store for backups/restores (Postgres/MSSQL/Oracle)
- db.py: small SQLite store used as default
- novabackup/providers/cloud: AWS/Azure/GCP cloud providers + Mock
- novabackup/api.py: FastAPI with /vms, /normalize, /backups, /backups/{backup_id}/restore
- novabackup/cli.py: CLI for backup create/list/restore
- novabackup/static/dashboard.html: minimal dashboard for data visibility
- migrate.py/migrate_cli.py: JSON -> DB migration tooling
- tests/: unit and integration tests including E2E
- docker-compose.yml: quick local multi-service setup (DB/API/UI)
- ROADMAP.md / TODO-PLAN.md: plans and roadmap (in repo)

Getting started (local dev)
- Prerequisites: Git, Python 3.9+, pip
- Linux/macOS:
  - git clone https://github.com/ajjs1ajjs/Backup.git
  - cd Backup
  - python3 -m venv venv
  - source venv/bin/activate
  - pip install -e ".[dev,api]"  # for tests and API
  - pytest -q
  - python -m novabackup list-vms
  - python -m novabackup backup create --vm-id vm1 --dest ./backups --type full --name my-backup
  - python -m novabackup.run_api  # optional: start API
  - http://localhost:8000/docs
- Windows:
  - git clone https://github.com/ajjs1ajjs/Backup.git
  - cd Backup
  - py -m venv venv
  - venv\Scripts\activate
  - pip install -e ".[dev,api]"
  - pytest -q
  - py -m novabackup list-vms
  - py -m novabackup backup create --vm-id vm1 --dest .\backups --type full --name my-backup
  - py -m novabackup.run_api
  - http://localhost:8000/docs

API usage (optional)
- Install API extras: pip install -e ".[api]"
- Start API: python -m novabackup.run_api
- Docs: http://localhost:8000/docs
- Security: You can enable API keys by setting NOVABACKUP_API_KEY; provide header X-API-Key in calls

Cloud providers (prototype)
- AWS, Azure, and GCP are scaffolded with SDK calls ready to be wired up.
- To enable, set NOVABACKUP_CLOUD_PROVIDERS=AWS,AZURE,GCP and provide credentials in environment or profile files
- Cloud endpoints: POST /backups (destination_type=cloud) and POST /backups/{backup_id}/restore (cloud)

Database backends
- Default: SQLite (backups.json)
- SA DB backends: Postgres/MSSQL/Oracle via SA_DBManager
- Environment controls: NOVABACKUP_DATABASE_URL for single database; NOVABACKUP_DATABASE_URL_POSTGRES, NOVABACKUP_DATABASE_URL_MSSQL, NOVABACKUP_DATABASE_URL_ORACLE for CI tests

Migration
- Migrate backups.json to DB via python -m novabackup.migrate_cli or python -m novabackup.migrate

Testing and CI
- Pytest, Mypy, Flake8 in CI
- Cloud tests via moto (AWS) and cloud mocks for Azure/GCP until real credentials are provided

Contribute
- See CONTRIBUTING.md (to be added) and ROADMAP for roadmap

License
- MIT (see LICENSE)
