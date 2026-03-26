# Novabackup

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python Version](https://img.shields.io/badge/python-3.10%2B-blue)](https://www.python.org/)

Production-ready, multi-stage Python backup/restore system with multi-cloud support and RBAC security.

Overview
- A modular backup/restore framework with API, CLI, and UI components.
- Stage 1 focuses on RBAC/OAuth2/JWT with token refresh (in-memory MVP).
- Stage 2 adds cloud orchestration with AWS/Azure/GCP skeletons and a mock provider for CI testing.
- Stage 3 introduces a frontend dashboard; Stage 4 provides production-style Docker/CI; Stage 5+ covers migration and docs.

Table of Contents
- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Architecture](#architecture)
- [Cloud Providers & Orchestrator](#cloud-providers--orchestrator)
- [API Reference](#api-reference)
- [CLI & UI](#cli--ui)
- [Testing](#testing)
- [Deployment](#deployment)
- [Documentation & Roadmap](#documentation--roadmap)
- [Contributing](#contributing)
- [License](#license)

## Features
- Stage 1: RBAC/OAuth2/JWT skeleton with token refresh (in-memory store).
- Stage 2: Cloud orchestration with AWS/Azure/GCP providers and a Mock provider for CI.
- Stage 3: Frontend dashboard (UI).
- Stage 4: Docker/CI pipelines suitable for production.
- Stage 5+: Migration tooling and documentation.
- RBAC is designed to be extended with scopes and rotation; tokens are rotatable via /token/refresh.

## Requirements
- Python 3.10 or newer
- Git
- Optional: Docker and Docker Compose for local development
- Cloud credentials (for real cloud tests; by default Mock provider is used in CI)

## Installation

### Option 1: Using the provided installer scripts (Recommended for first-time users)

#### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat | iex
```

> These scripts will create a virtual environment at `$HOME/.novabackup/venv` (Linux/macOS) or `%USERPROFILE%\.novabackup\venv` (Windows), install the package in development mode with API and dev dependencies, and attempt to fetch the latest source from the repository.

### Option 2: Manual installation

1. Clone the repository
   ```bash
   git clone https://github.com/ajjs1ajjs/Backup.git
   cd Backup
   ```
2. Create a Python virtual environment and install dependencies
   ```bash
   python -m venv venv
   source venv/bin/activate  # on Windows: .\\venv\\Scripts\\activate
   pip install -e ".[dev,api]"
   ```
3. Create a local configuration (optional)
   - The project uses environment variables for sensitive config:
     - NOVABACKUP_JWT_SECRET: secret for JWT signing
     - NOVABACKUP_DATABASE_URL: optional DB URL for DB-backed storage
     - NOVABACKUP_CLOUD_PROVIDER: MOCK to use mock cloud provider in CI, or AWS/AZURE/GCP for real providers when credentials are available

## Quick Start

To get Novabackup up and running in under a minute:

### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | bash && source $HOME/.novabackup/venv/bin/activate && novabackup list-vms
```

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat | iex; & "$USERPROFILE\.novabackup\venv\Scripts\activate"; novabackup list-vms
```

> Note: The `novabackup list-vms` command may return an empty list or an error if no hypervisor is available. This is expected in environments without Hyper-V (Windows) or libvirt (Linux). The installation is successful if the command runs without import errors.

## Usage
- Run the API locally (FastAPI / uvicorn)
  ```bash
  uvicorn novabackup.api:get_app --reload --port 8000
  ```
- Open the docs at http://localhost:8000/docs
- Authenticate with /token (POST) to obtain access and refresh tokens, then call protected endpoints like /vms, /backups, etc.
- Cloud backups (Stage 2): set destination_type to "cloud" and provide cloud_provider, cloud_region, cloud_dest in the create request. By default, Mock provider is used in CI for safe testing.

## API Reference (high level)
- POST /token
- POST /token/refresh
- GET /vms
- GET /normalize/{vm_type}
- POST /backups
- GET /backups
- POST /backups/{backup_id}/restore

## Architecture
- Core: novabackup/core.py provides VM enumeration with optional local and cloud fallbacks.
- API: novabackup/api.py exposes REST endpoints protected by RBAC/JWT.
- Security: novabackup/security.py implements in‑memory RBAC and JWT tokens with refresh support.
- Cloud: novabackup/cloudops.py and provider modules implement cloud orchestration and mock providers for CI.
- Persistence: backups stored either in a DB backend or a simple JSON file for CLI usage.
- UI: Static HTML dashboard and basic frontend scaffolding.

## Cloud Providers & Orchestrator
- Stage 2 uses CloudOrchestrator which aggregates providers: AWS, Azure, GCP and Mock.
- The orchestrator selects a provider based on the provided provider string (e.g., AWS, Azure, GCP, Mock).
- Mock provider is available for CI; real cloud providers can be wired when credentials are present.

## Testing
- Unit and integration tests cover API auth, RBAC, backups, and cloud-backup flows.
- Run tests with pytest: `pytest -q`.
- Tests rely on the Mock cloud provider by default to avoid real cloud credentials in CI.

## Deployment
- Dockerfile provided for building a production-ready image.
- docker-compose.yml for local development (builds from Dockerfile).
- docker-compose-prod.yml template for production deployment (uses external image).
- Production installation scripts are included for Linux and Windows.
- GitHub Actions CI workflow (`.github/workflows/ci.yml`) runs tests and builds Docker image on push/PR to main.

## Documentation & Roadmap
- ROADMAP.md and SUMMARY.md outline ongoing work and milestones for Stage 1 through Stage 7.
- Migration guide: MIGRATION_GUIDE.md
- Release notes: RELEASE_NOTES.md

## Contributing
- Please open issues or submit pull requests with your proposed changes.
- Ensure tests pass locally before opening a PR.

## License
This project is licensed under the MIT License. See the LICENSE file for details.

Notes
- The base code uses Mock cloud providers by default in CI. To test real cloud integration locally, configure appropriate credentials and set NOVABACKUP_CLOUD_PROVIDER to AWS/AZURE/GCP.
- If you want Ukrainian language README as well, I can add a bilingual version or translate this file.
