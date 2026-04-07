# Testing Guide

## Overview

The current automated test contour covers:
- unit tests for scheduler behavior and related service logic
- integration tests for the REST API surface
- authenticated flows for jobs, repositories, restores, agents, and reports

The tests are aligned with the current REST-based server architecture.

## Requirements

- .NET SDK `8.0.419` or compatible with `global.json`
- Node.js 18+ if you want to run UI tests separately

## Run Tests

Run the backend tests from the repository root:

```powershell
dotnet test src/server/Backup.Server.Tests/Backup.Server.Tests.csproj
dotnet test src/server/Backup.Server.IntegrationTests/Backup.Server.IntegrationTests.csproj
```

On Windows, run these commands sequentially. Running both projects in parallel can cause `obj/bin` file-lock issues for the shared server project.

## What Passes Now

Validated locally in this repository:
- `Backup.Server.Tests`: 3 passed
- `Backup.Server.IntegrationTests`: 7 passed

## Integration Test Notes

The integration suite uses `WebApplicationFactory<Program>` and an in-memory SQLite connection for isolated API validation.

Covered flows include:
- authenticated access control
- job creation and execution
- repository creation and listing
- restore request validation
- reports endpoint authorization

## Troubleshooting

If tests fail on Windows with file access errors:

```powershell
Get-Process dotnet -ErrorAction SilentlyContinue | Stop-Process -Force
Remove-Item -LiteralPath src/server/Backup.Server/bin -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath src/server/Backup.Server/obj -Recurse -Force -ErrorAction SilentlyContinue
```

Then rerun one `dotnet test` command at a time.
