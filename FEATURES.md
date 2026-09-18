# GoreeCloud Photos — Current Features

**Lifecycle:** Experimental  
**Implementation status:** First executable server foundation validated; not ready for ordinary Photos use.  
**Verified:** 2026-09-18

## Current implemented product functionality

The current Experimental slice provides:

- Go 1.27.1 server process foundation.
- Loopback-by-default HTTP listener (`127.0.0.1:8780` unless explicitly configured).
- `GET /api/v1/health` returning bounded version/lifecycle process health.
- `GET /api/v1/ready` with fail-closed component state.
- Readiness intentionally returns HTTP 503 because the PostgreSQL runtime adapter is not implemented.
- Optional filesystem original-media store initialization through `GC_PHOTOS_STORAGE_ROOT`.
- Immutable filesystem writes using staging, SHA-256 calculation/verification, no-overwrite commit semantics, bounded object keys, and regular-file checks.
- Filesystem storage readiness probing.
- Initial PostgreSQL migration baseline for libraries, memberships, original objects, assets, device asset state, upload sessions, ordered synchronization changes, durable jobs, and idempotency records.
- Automated Go formatting, module reproducibility, vet, unit-test, migration-baseline, and binary-build validation.
- Repository/contract baseline validation.

## Verified limitations

The current foundation does **not** provide:

- a PostgreSQL runtime adapter or a ready service state;
- a media-upload HTTP API;
- authenticated users, GoreeCloud Identity runtime integration, or Photos authorization;
- multi-user library operations;
- asset creation through the API;
- metadata extraction;
- thumbnails or previews;
- synchronization clients;
- web, Android, Linux desktop, or Mobile B clients;
- sharing, search, intelligence, editing, memories, import/export, backup, or restore;
- a supported deployment or production release;
- accepted Privacy Shield, Wardveil Security, Everkeep, Manager, Mesh, Policy, Observability, or Glaze UI integration.

The filesystem store is an Experimental implementation component, not proof that a user's media has been backed up or protected.

## Repository and contract capabilities

The repository also contains:

- canonical product specification and roadmap;
- Phase 0 architecture, data model, API, recovery, dependency, platform-integration, and accessibility contracts;
- machine-readable Asset, Sync Change, and Upload Session schemas;
- Platform Contract 0.4 declaration;
- AGPL-3.0-or-later fallback license record.

See `FEATURE-ROADMAP.md` for the next implementation steps.
