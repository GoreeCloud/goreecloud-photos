# GoreeCloud Photos — Repository Notes

**As of:** 2026-09-18

## Verified state

- Repository: `GoreeCloud/goreecloud-photos`.
- Default branch: `main`.
- PR #6 carries the PostgreSQL runtime/readiness slice; final merge evidence is recorded in the GoreeCloud Photos change log and task ledger.
- Product lifecycle: **Experimental**. PR #6 extends the validated prototype with PostgreSQL runtime/readiness integration without advancing the lifecycle.
- Go 1.27.1 is pinned through `go.mod` and exact CI toolchain verification.
- The server defaults to `127.0.0.1:8780`.
- `GET /api/v1/health` is implemented.
- `GET /api/v1/ready` is implemented and fails closed unless both the PostgreSQL core schema and original-media store are healthy.
- The filesystem original-media adapter supports immutable no-overwrite writes, SHA-256 verification, staging, regular-file reads, and readiness probing.
- Initial PostgreSQL schema migrations cover libraries, membership, original objects, assets, device state, upload sessions, ordered synchronization changes, jobs, and idempotency records.
- The synchronized Drive roadmap exists at `GoreeCloud/Feature Roadmap/GoreeCloud Photos/goreecloud-photos.md`.
- The central user-manual copy exists at `GoreeCloud/User Manuals/User Manual — GoreeCloud Photos.md`.
- The Photos-specific active task record exists at `GoreeCloud/Tasks Management/GoreeCloud Photos — Implementation Task List.md`.
- A PostgreSQL runtime adapter is implemented on PR #6. No upload HTTP API, transactional library workflow, authentication/authorization integration, user-facing client, supported deployment, release artifact, production acceptance, or Stable qualification is verified.
- All nine Photos-specific Integral Platform System integrations remain blocked pending runtime implementation and accepted evidence.
- The repository has no GitHub ruleset. The connected integration cannot read the branch-protection endpoint, so branch-protection state is not fully verified.
- Merged topic branches remain visible because the current GitHub connector does not expose branch deletion.

## Immediate repository work

1. Implement durable upload-session persistence and the first resumable upload endpoint family.
2. Atomically commit original-object, Asset, and SyncChange state only after immutable media verification.
3. Extend storage with an S3-compatible backend without changing Photos domain authority.
4. Continue exact-head Go/PostgreSQL integration validation as database-backed domain operations are added.
6. Establish repository-local official visual identity assets.
7. Resolve branch-protection state using an authorized administration path.
8. Keep all nine Integral Platform System integrations blocked until implementation and evidence exist.

## Status discipline

Experimental identifies the validated prototype foundation only. It does not establish ordinary photo backup, supported deployment, Production, Release Candidate, Stable, recovery readiness, or platform-system acceptance.
