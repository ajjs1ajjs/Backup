# Novabackup

This project is a Python-based MVP CLI for VM listing and type normalization with optional API.

All-in-one install and run guide (Windows and Linux).

Prerequisites
- Git
- Python 3.9+
- Optional: For API features, extras: fastapi, uvicorn

1) Clone the repository
- Windows:
  - git clone https://your-repo-url/novabackup.git
- Linux:
  - git clone https://your-repo-url/novabackup.git

2) Install and run (local dev)
- Linux/macOS
  - cd novabackup
  - python3 -m venv venv
  - source venv/bin/activate
  - pip install -e .
  - python -m novabackup list-vms
  - python -m novabackup normalize KVM

- Windows
  - cd novabackup
  - py -m venv venv
  - venv\Scripts\activate
  - pip install -e .
  - py -m novabackup list-vms
  - py -m novabackup normalize KVM

3) API (optional)
- To enable API, install extras and run the API server
  - Linux/Windows:
    - pip install -e .[api]
    - python -m novabackup.run_api
  - Access API endpoints:
    - http://localhost:8000/vms
    - http://localhost:8000/normalize/{vm_type}

4) Tests
- pip install -e .[dev]  # if extras are defined, else install pytest
- pytest

-5) One-line installers (example)
- Windows: iwr -Uri https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat -OutFile install.bat; .\\install.bat
- Linux/macOS: curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | bash
Note: Replace the URL with your actual repository URL when published.
- Windows: iwr -Uri https://raw.githubusercontent.com/your-repo/novabackup/main/install.bat -OutFile install.bat; .\\install.bat
- Linux/macOS: curl -fsSL https://raw.githubusercontent.com/your-repo/novabackup/main/install.sh | bash
Note: Replace the URL with your actual repository URL when published.

Notes
- This is a starting MVP. Real provider integrations (Hyper-V, KVM, VMware) can be added via providers modules.
- The roadmap and patch history live in ROADMAP.md and tests/test_core.py as unit tests.
