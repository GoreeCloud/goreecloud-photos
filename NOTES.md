# GoreeCloud Photos — Repository Notes

**As of:** 2026-09-18

## Verified state

- Repository: `GoreeCloud/goreecloud-photos`.
- Default branch: `main`; PR #5 is the current Experimental-foundation candidate until merged.
- Product lifecycle candidate: **Experimental**, supported by the validated executable server foundation on the PR head.
- Go 1.27.1 is pinned through `go.mod` and exact CI toolchain verification.
- The server defaults to `127.0.0.1:8780`.
- `GET /api/v1/health` is implemented.
- `GET /api/v1/ready` is implemented and intentionally returns HTTP 503 because the PostgreSQL runtime adapter is not implemented.
- The filesystem original-media adapter supports immutable no-overwrite writes, SHA-256 verification, staging, regular-file reads, and readiness probing.
- Initial PostgreSQL schema migrations cover libraries, membership, original objects, assets, device state, upload sessions, ordered synchronization changes, jobs, and idempotency records.
- The synchronized Drive roadmap exists at `GoreeCloud/Feature Roadmap/GoreeCloud Photos/goreecloud-photos.md`.
- The central user-manual copy exists at `GoreeCloud/User Manuals/User Manual — GoreeCloud Photos.md`.
- The Photos-specific active task record exists at `GoreeCloud/Tasks Management/GoreeCloud Photos — Implementation Task List.md`.
- No upload HTTP API, database runtime adapter, authentication/authorization integration, user-facing client, supported deployment, release artifact, production acceptance, or Stable qualification is verified.
- All nine Photos-specific Integral Platform System integrations remain blocked pending runtime implementation and accepted evidence.
- The repository has no GitHub ruleset. The connected integration cannot read the branch-protection endpoint, so branch-protection state is not fully verified.
- Merged topic branches remain visible because the current GitHub connector does not expose branch deletion.

## Immediate repository work

1. Implement a pinned PostgreSQL runtime adapter and database readiness probe.
2. Implement durable upload-session persistence and the first resumable upload endpoint family.
3. Atomically commit original-object, Asset, and SyncChange state only after immutable media verification.
4. Extend storage with an S3-compatible backend without changing Photos domain authority.
5. Continue exact-head Go CI and add database-backed integration tests when the adapter exists.
6. Establish repository-local official visual identity assets.
7. Resolve branch-protection state using an authorized administration path.
8. Keep all nine Integral Platform System integrations blocked until implementation and evidence exist.

## Status discipline

Experimental identifies the validated prototype foundation only. It does not establish ordinary photo backup, supported deployment, Production, Release Candidate, Stable, recovery readiness, or platform-system acceptance.
