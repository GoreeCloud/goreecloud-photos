# GoreeCloud Photos — Repository Notes

**As of:** 2026-09-18

## Verified state

- Repository: `GoreeCloud/goreecloud-photos`.
- Default branch: `main`.
- The canonical product specification exists.
- The repository documentation/governance baseline is established on `main`.
- The synchronized Drive roadmap exists at `GoreeCloud/Feature Roadmap/GoreeCloud Photos/goreecloud-photos.md`.
- The central user-manual copy exists at `GoreeCloud/User Manuals/User Manual — GoreeCloud Photos.md` and is synchronized with the repository user manual.
- The Photos-specific active task record exists at `GoreeCloud/Tasks Management/GoreeCloud Photos — Implementation Task List.md`.
- No supported application or service implementation foundation is verified.
- No release artifact, deployment, production acceptance, or Stable qualification is verified.
- Glaze UI V1.5 / 1.5.1 is the current Stable design-system target.
- Platform Contract schema 0.4 and the current nine-system Integral Platform Systems model are the intended governance baseline.
- All Photos-specific Integral Platform System integrations remain blocked pending implementation and evidence.
- The repository has no GitHub ruleset. The connected integration cannot read the branch-protection endpoint, so branch-protection state is not fully verified.
- Merged documentation topic branches remain visible because the current GitHub connector does not expose branch deletion.

## Immediate repository work

1. Establish the first executable server source tree and pinned Go module from the Phase 0 architecture baseline.
2. Implement the initial PostgreSQL schema and migrations for libraries, assets, originals, uploads, synchronization changes, and durable jobs.
3. Implement the smallest end-to-end upload/storage slice behind the defined API and storage-driver boundaries.
4. Extend CI from repository/contract validation to Go formatting, vetting, unit tests, build validation, and later web/Android/Linux client validation as source is introduced.
5. Establish repository-local official visual identity assets.
6. Resolve branch-protection state using an authorized administration path.
7. Keep all nine Integral Platform System integrations blocked until runtime implementation and evidence exist.

## Status discipline

Do not upgrade lifecycle state merely because documentation exists. Concept/Planning should advance only when implementation evidence supports the applicable lifecycle transition.
