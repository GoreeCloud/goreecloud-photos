# GoreeCloud Photos

GoreeCloud Photos is the planned GoreeCloud personal and family visual-memory platform for private photo and video backup, synchronization, organization, search, sharing, editing, intelligence, portability, and long-term preservation.

> **Current state:** Experimental. The first executable server foundation is validated, but it is intentionally not ready for ordinary Photos use and no supported client, deployment, release artifact, or production acceptance exists.

## Product direction

GoreeCloud Photos is intended to provide:

- verified photo and video backup;
- a unified chronological library across devices;
- local-first and offline-capable clients;
- albums, favorites, archive, protected media, sharing, and family libraries;
- structured and optional semantic search;
- optional privacy-controlled photo intelligence;
- non-destructive editing;
- portable import and export;
- tested recovery and long-term preservation.

The design principle is that original media remains user-owned, understandable outside the application, and recoverable even if the current application no longer exists.

## Current authority

- [SPECIFICATIONS.md](SPECIFICATIONS.md) — canonical product and technical specification.
- [FEATURE-ROADMAP.md](FEATURE-ROADMAP.md) — implementation-facing roadmap and lifecycle sequence.
- [ARCHITECTURE.md](ARCHITECTURE.md) — Phase 0 system and implementation architecture.
- [DATA-MODEL.md](DATA-MODEL.md) — stable asset, ownership, lifecycle, and synchronization domain model.
- [API.md](API.md) — versioned API, resumable-upload, idempotency, cursor, and conflict contracts.
- [RECOVERY.md](RECOVERY.md) — authoritative/rebuildable data and clean-target recovery model.
- [PLATFORM-INTEGRATIONS.md](PLATFORM-INTEGRATIONS.md) — nine-system GoreeCloud integration boundaries.
- [DEPENDENCIES.md](DEPENDENCIES.md) — selected implementation stack and dependency rules.
- [GLAZE-UI-ACCEPTANCE.md](GLAZE-UI-ACCEPTANCE.md) — Glaze UI 1.5.1 and accessibility acceptance plan.
- [FEATURES.md](FEATURES.md) — current implemented capability state.
- [USER-MANUAL.md](USER-MANUAL.md) — current user-facing availability and usage status.
- [PRIVACY POLICY.md](PRIVACY%20POLICY.md) — current privacy boundary.
- [SECURITY.md](SECURITY.md) — repository-safe security guidance.
- [goreecloud.platform.yaml](goreecloud.platform.yaml) — machine-readable platform declaration.

## Experimental server foundation

The verified Experimental foundation currently includes:

- a Go 1.27.1 service binary;
- loopback-by-default HTTP serving;
- `GET /api/v1/health`;
- fail-closed `GET /api/v1/ready`, which currently returns HTTP 503 because the PostgreSQL runtime adapter is not implemented;
- an immutable filesystem original-media store adapter with SHA-256 verification and overwrite protection;
- an initial PostgreSQL schema/migration baseline for libraries, ownership, assets, originals, upload sessions, synchronization changes, durable jobs, and idempotency records;
- unit tests plus repository, migration, vet, test, and build validation.

PostgreSQL remains the selected authoritative relational store, but no runtime database adapter is implemented yet. TypeScript + React web, native Android, native Linux desktop, S3-compatible storage, and Mobile B remain planned implementation targets rather than current runtime capabilities.

## Platform targets

- **Design system:** Glaze UI V1.5 / 1.5.1 current Stable target.
- **Identity:** GoreeCloud Identity.
- **Privacy:** GoreeCloud Privacy Shield.
- **Security:** Wardveil Security.
- **Resilience and preservation:** Everkeep.
- **Platform Contract:** schema 0.4, using the current nine-system Integral Platform Systems model.

All application-specific platform integrations remain blocked pending implementation and evidence.

## Repository state

The product is being developed as original GoreeCloud-controlled software. It is not intended to be a renamed or permanently architecture-dependent copy of another photo platform.

An Experimental server foundation is verified, but it does not expose a media-upload API and is intentionally not ready because the database adapter is absent. Do not use this repository as evidence that media has been backed up, synchronized, protected, encrypted, indexed, recoverable, or safely deletable from a device.

## License

Unless superseded by an authorized Photos-specific license decision, this repository uses the GoreeCloud default fallback license: **GNU Affero General Public License v3.0 or later (AGPL-3.0-or-later)**. See [LICENSE](LICENSE).

## Status integrity

Documentation, machine-readable contracts, CI validation, design intent, planned integrations, and repository structure do not establish implementation, release, security acceptance, privacy acceptance, recovery acceptance, or production readiness. Those states require independent evidence for the exact code and runtime being evaluated.
