# GoreeCloud Photos

GoreeCloud Photos is the planned GoreeCloud personal and family visual-memory platform for private photo and video backup, synchronization, organization, search, sharing, editing, intelligence, portability, and long-term preservation.

> **Current state:** Concept / planning. The repository contains product and governance documentation, but no supported Photos runtime, client, server, release artifact, deployment, or production acceptance has been established.

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
- [FEATURES.md](FEATURES.md) — current implemented capability state.
- [USER-MANUAL.md](USER-MANUAL.md) — current user-facing availability and usage status.
- [PRIVACY POLICY.md](PRIVACY%20POLICY.md) — current privacy boundary.
- [SECURITY.md](SECURITY.md) — repository-safe security guidance.
- [goreecloud.platform.yaml](goreecloud.platform.yaml) — machine-readable platform declaration.

## Platform targets

- **Design system:** Glaze UI V1.5 / 1.5.1 current Stable target.
- **Identity:** GoreeCloud Identity.
- **Privacy:** GoreeCloud Privacy Shield.
- **Security:** Wardveil Security.
- **Resilience and preservation:** Everkeep.
- **Platform Contract:** schema 0.4, using the current nine-system Integral Platform Systems model.

These are integration targets, not implemented or accepted integrations.

## Repository state

The product is being developed as original GoreeCloud-controlled software. It is not intended to be a renamed or permanently architecture-dependent copy of another photo platform.

No application implementation foundation is currently verified. Do not use this repository as evidence that media has been backed up, synchronized, protected, encrypted, indexed, recoverable, or safely deletable from a device.

## License

Unless superseded by an authorized Photos-specific license decision, this repository uses the GoreeCloud default fallback license: **GNU Affero General Public License v3.0 or later (AGPL-3.0-or-later)**. See [LICENSE](LICENSE).

## Status integrity

Documentation, design intent, planned integrations, and repository structure do not establish implementation, release, security acceptance, privacy acceptance, recovery acceptance, or production readiness. Those states require independent evidence for the exact code and runtime being evaluated.
