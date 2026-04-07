# Backup System

Backup System is a backup management prototype with a .NET 8 server, a React web UI, a C++ agent, and installer scripts for Windows and Linux.

The repository currently provides:
- A REST API for jobs, repositories, backups, restores, reports, settings, agents, hypervisors, and virtual machines
- JWT-based authentication with role policies
- SQLite-backed persistence through EF Core
- A React UI that is built and served from the server's `wwwroot`
- Early-stage agent and restore logic scaffolding

The repository does not currently provide a complete production-ready backup engine. Several advanced features listed in earlier drafts of this README are still partial, stubbed, or under development.

## Architecture

- Server: `src/server/Backup.Server`
- UI: `src/ui`
- Agent: `src/agent/Backup.Agent`
- Installer: `src/installer/Backup.Agent.Installer`
- Contracts: `src/protos`

## Current Status

Implemented at a working application level:
- REST controllers and JWT authentication
- SQLite data model and migrations
- Basic scheduler background services
- Hypervisor and VM inventory endpoints
- Repository CRUD and connection checks
- Reports and settings endpoints

Still incomplete or prototype-level:
- Actual backup execution pipeline
- Real restore orchestration
- End-to-end agent/server runtime communication
- Full cloud and hypervisor integration depth
- Production hardening and deployment validation

## Development

### Requirements

- .NET 8 SDK (`8.0.419` is pinned in `global.json`)
- Node.js 18+
- npm

### Server

```powershell
dotnet restore src/server/Backup.Server/Backup.Server.csproj
dotnet run --project src/server/Backup.Server/Backup.Server.csproj
```

### UI

```powershell
cd src/ui
npm install
npm run build
```

If the UI is built, the server serves the static files from `src/server/Backup.Server/wwwroot`.

### Tests

```powershell
dotnet test src/server/Backup.Server.Tests/Backup.Server.Tests.csproj
dotnet test src/server/Backup.Server.IntegrationTests/Backup.Server.IntegrationTests.csproj
```

On Windows, run these test projects sequentially rather than in parallel because they share server build outputs and can hit file locks.

## Configuration

Main configuration file:
- `src/server/Backup.Server/appsettings.json`

Important sections:
- `ConnectionStrings:DefaultConnection`
- `Jwt`
- `Server`
- `BootstrapAdmin`
- `AllowedOrigins`
- `Encryption`

On first startup the server can generate:
- `jwt.key`
- `data/encryption.key`

These files should not be committed.

## Default Access

Default bootstrap values in development:
- Username: `admin`
- Password: generated at first startup if not configured explicitly

The bootstrap admin is marked to change password on first login.

## Security Notes

- The API now requires authentication by default.
- Anonymous access is limited to login, registration, and first-login password change.
- Self-registration always creates `Viewer` users.
- Emergency password reset is intentionally disabled.
- If `AllowedOrigins` is empty, CORS is limited to local development origins only.

## Repository Hygiene

This repository should only store source files and intentionally versioned assets. Build outputs such as `bin/`, `obj/`, generated publish folders, local database files, and secret keys should stay untracked.

## Documents

- [API docs](API_DOCS.md)
- [Install guide](install.md)
- [Testing guide](TESTING.md)
- [Requirements](requirements.md)
- [Validation](VALIDATION.md)
- [Roadmap](roadmap.md)

## License

MIT
