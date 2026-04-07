# Roadmap

## 1. Backup Engine

- implement full backup execution beyond the current restore-focused MVP
- add stronger backup point creation, metadata, and verification flows
- improve retention and repository lifecycle handling

## 2. Hypervisor Depth

- expand VMware integration beyond the current limited path
- deepen KVM support
- tighten inventory, discovery, and execution flows for managed workloads

## 3. Recovery Features

- improve instant restore and file-level recovery behavior
- add more robust cancel, retry, and progress reporting flows
- extend recovery validation and operator feedback in the UI

## 4. Deployment and CI

- keep CI running unit and integration tests automatically
- improve publish flow for the server and built UI assets
- document production deployment expectations and secrets handling

## 5. UX and Operator Flow

- continue polishing remaining UI screens for consistency
- improve dashboard and reporting usefulness
- reduce confusion around partial or not-yet-implemented workflows

## 6. Verification

- add broader API coverage for auth and edge cases
- add more restore and repository integration scenarios
- validate the full workflow continuously as the backup engine matures
