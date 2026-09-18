# GoreeCloud Photos — Repository Notes

**As of:** 2026-09-18

## Verified state

- Repository: `GoreeCloud/goreecloud-photos`.
- Default branch: `main`; PR #7 is the current durable-upload-persistence candidate until merged.
- Product lifecycle: **Experimental**. The candidate extends the validated prototype with durable PostgreSQL upload-session/part metadata without advancing lifecycle.
- Go 1.27.1 is pinned through `go.mod` and exact CI toolchain verification.
- The server defaults to `127.0.0.1:8780`.
- `GET /api/v1/health` is implemented.
- `GET /api/v1/ready` fails closed unless the PostgreSQL schema and original-media store are healthy.
- The pgx v5.11.0 PostgreSQL runtime adapter and schema-aware readiness boundary are implemented.
- The filesystem original-media adapter supports immutable no-overwrite writes, SHA-256 verification, staging, regular-file reads, and readiness probing.
- PostgreSQL migrations cover libraries, membership, original objects, assets, device state, upload sessions, per-part upload receipt evidence, ordered synchronization changes, jobs, and idempotency records.
- PR #7 candidate code persists upload sessions and per-part checksum/length evidence across adapter restart, supports out-of-order parts, treats identical retries idempotently, rejects conflicting retries, and persists expiry state.
- The synchronized Drive roadmap exists at `GoreeCloud/Feature Roadmap/GoreeCloud Photos/goreecloud-photos.md`.
- The central user-manual copy exists at `GoreeCloud/User Manuals/User Manual — GoreeCloud Photos.md`.
- The Photos-specific active task record exists at `GoreeCloud/Tasks Management/GoreeCloud Photos — Implementation Task List.md`.
- No media-upload HTTP API, media-byte staging/assembly path, upload completion-to-Asset transaction, authentication/authorization integration, user-facing client, supported deployment, release artifact, production acceptance, or Stable qualification is verified.
- All nine Photos-specific Integral Platform System integrations remain blocked pending runtime implementation and accepted evidence.
- The repository has no GitHub ruleset. The connected integration cannot read the branch-protection endpoint, so branch-protection state is not fully verified.
- Merged topic branches remain visible because the current GitHub connector does not expose branch deletion.

## Immediate repository work

1. Implement the first resumable upload endpoint family without weakening the existing session/part invariants.
2. Implement media-byte staging/assembly and server-side part checksum verification before accepting persisted receipt evidence from HTTP.
3. Atomically commit immutable OriginalObject, Asset, and SyncChange state only after full upload and ingestion verification.
4. Extend storage with an S3-compatible backend without changing Photos domain authority.
5. Continue exact-head Go/PostgreSQL integration validation as database-backed domain operations are added.
6. Establish repository-local official visual identity assets.
7. Resolve branch-protection state using an authorized administration path.
8. Keep all nine Integral Platform System integrations blocked until implementation and evidence exist.

## Status discipline

Experimental identifies the validated prototype foundation only. Durable upload metadata is not media backup. It does not establish ordinary photo backup, supported deployment, Production, Release Candidate, Stable, recovery readiness, or platform-system acceptance.
