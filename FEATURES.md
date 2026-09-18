# GoreeCloud Photos — Current Features

**Lifecycle:** Experimental  
**Implementation status:** Database-backed upload-session persistence validated; no user-facing upload path or ordinary Photos service exists.  
**Verified:** 2026-09-18

## Current implemented product functionality

The current Experimental slice provides:

- Go 1.27.1 server process foundation.
- Loopback-by-default HTTP listener (`127.0.0.1:8780` unless explicitly configured).
- `GET /api/v1/health` returning bounded version/lifecycle process health.
- Fail-closed `GET /api/v1/ready`.
- pgx v5.11.0 PostgreSQL runtime adapter configured through `GC_PHOTOS_DATABASE_URL`.
- Database readiness probing with a bounded timeout.
- Readiness requires the complete core schema; a reachable but unmigrated/incomplete database remains non-ready.
- Readiness succeeds only when both PostgreSQL/schema and original-media storage are healthy.
- Optional filesystem original-media store initialization through `GC_PHOTOS_STORAGE_ROOT`.
- Immutable filesystem writes using staging, SHA-256 calculation/verification, no-overwrite commit semantics, bounded object keys, and regular-file checks.
- Filesystem storage readiness probing.
- PostgreSQL migration baseline for libraries, memberships, original objects, assets, device asset state, upload sessions, per-part upload evidence, ordered synchronization changes, durable jobs, and idempotency records.
- Durable upload-session creation/inspection and per-part receipt persistence in PostgreSQL.
- Per-part checksum evidence with independently retryable, out-of-order part recording.
- Identical part retries are idempotent; conflicting retries are rejected without changing committed evidence.
- Upload-session metadata survives PostgreSQL adapter restart and expired sessions fail closed.
- Real PostgreSQL integration validation that exercises migration up/down, schema-aware readiness, durable upload persistence, retry/conflict behavior, and expiry.
- Automated Go formatting, module reproducibility, vet, unit/integration tests, migration-baseline validation, and binary-build validation.
- Repository/contract baseline validation.

## Verified limitations

The current foundation does **not** provide:

- a media-upload HTTP API, media-byte staging/assembly path, or upload completion-to-Asset transaction;
- authenticated users, GoreeCloud Identity runtime integration, or Photos authorization;
- multi-user library operations;
- asset creation through the API;
- metadata extraction;
- thumbnails or previews;
- synchronization clients;
- web, Android, Linux desktop, or Mobile B clients;
- sharing, search, intelligence, editing, memories, import/export, backup, or restore;
- an S3-compatible storage adapter;
- a supported deployment or production release;
- accepted Privacy Shield, Wardveil Security, Everkeep, Manager, Mesh, Policy, Observability, or Glaze UI integration.

The database and filesystem adapters are Experimental implementation components. Their existence and green CI do not prove that a user's media has been backed up, protected, recovered, or made production-safe.

## Repository and contract capabilities

The repository also contains:

- canonical product specification and roadmap;
- Phase 0 architecture, data model, API, recovery, dependency, platform-integration, and accessibility contracts;
- machine-readable Asset, Sync Change, and Upload Session schemas;
- Platform Contract 0.4 declaration;
- AGPL-3.0-or-later fallback license record.

See `FEATURE-ROADMAP.md` for the next implementation steps.
