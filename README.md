# Novabackup

A Python MVP for VM backup and restore across Windows Hyper-V and Linux Libvirt, with a DB backend, CLI, REST API, and a lightweight dashboard.

Key features
- Windows Hyper-V provider for VM enumeration and restore workflows
- Linux Libvirt provider for VM enumeration
- SQLite DB backend for backups and restores (with optional SQLAlchemy DB backends for PostgreSQL, MySQL/MariaDB, MSSQL, Oracle)
- CLI (backup create/list/restore) and REST API (FastAPI with Swagger docs)
- Lightweight dashboard served from the API at /static/dashboard.html
- Installers for Linux/macOS and Windows; CI with tests

Getting started (one-click style)
- Prerequisites: Git, Python 3.9+, and optionally OS-specific DB drivers if you plan to use a real DB backend

- Linux/macOS setup
  1) git clone https://github.com/ajjs1ajjs/Backup.git
  2) cd Backup
  3) python3 -m venv venv
  4) source venv/bin/activate
  5) pip install -e ".[dev,api]"
  6) pytest -q
  7) python -m novabackup list-vms
  8) python -m novabackup backup create --vm-id vm1 --dest ./backups --type full --name my-backup
  9) python -m novabackup.run_api  # optional: start API
 10) Open http://localhost:8000/docs for API docs

- Windows setup
  1) git clone https://github.com/ajjs1ajjs/Backup.git
  2) cd Backup
  3) py -m venv venv
  4) .\venv\Scripts\activate
  5) pip install -e ".[dev,api]"
  6) pytest -q
  7) py -m novabackup list-vms
  8) py -m novabackup backup create --vm-id vm1 --dest .\backups --type full --name my-backup
  9) py -m novabackup.run_api
  10) Open http://localhost:8000/docs for API docs

API usage
- The API is built with FastAPI. After starting the API server, you can access:
  - GET /vms
  - GET /normalize/{vm_type}
  - POST /backups
  - GET /backups
  - POST /backups/{backup_id}/restore
  - Docs: http://localhost:8000/docs
  - Optional API Key: set NOVABACKUP_API_KEY and pass header X-API-Key

Database notes
- By default, backups are stored in backups.json (local JSON store).
- You can switch to a real database by setting NOVABACKUP_DATABASE_URL to a SQLAlchemy URL, for example:
  - sqlite:///./novabackup.db
  - postgresql+psycopg2://user:pass@host:5432/novabackup
  - mysql+mysqlconnector://user:pass@host:3306/novabackup
  - mssql+pyodbc://user:pass@host:1433/novabackup?driver=ODBC+Driver+17+for+SQL+Server
  - oracle+cx_oracle://user:pass@host:1521/?service_name=ORCLCDB
Note: You need to install the respective drivers (psycopg2-binary, pyodbc, cx_Oracle).

Migration (JSON → DB)
- A small migration utility moves backups.json into a DB; see novabackup/migrate.py and novabackup/migrate_cli.py for usage.

Tests
- Run: pytest -q

Roadmap and future work
- See ROADMAP.md for the current plan and milestones.

If you need help or want me to tailor the DW (Docs/Wiki) for your team, just say the word.
