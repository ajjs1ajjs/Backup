# Requirements

## Runtime and Tooling

- .NET SDK 8.x
- Node.js 18+
- npm
- CMake toolchain for the C++ agent

The repository pins `.NET SDK 8.0.419` in [global.json](/c:/PROJECT/Backup/global.json).

## Current Functional Scope

Implemented or available at a usable level:
- REST API for jobs, repositories, backups, restores, reports, settings, agents, hypervisors, and virtual machines
- JWT authentication with role policies
- SQLite-backed server persistence
- React UI for the main operator workflows
- background restore queue and scheduler services

## Current Product Limitations

Not yet complete or still partial:
- full production-grade backup engine
- advanced replication workflows
- deep VMware and KVM orchestration
- complete file-level and database-native recovery coverage
- full deployment automation and production hardening

## Security Baseline

- authentication is required by default
- self-registration creates `Viewer` users only
- emergency admin password reset is disabled
- empty `AllowedOrigins` falls back to local development origins only

## Testing Baseline

Validated locally:
- unit tests for scheduler behavior
- REST integration tests for core API flows

For local Windows runs, execute backend test projects sequentially to avoid file-lock issues in shared `bin/obj` outputs.
