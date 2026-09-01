# GoreeCloud Photos

GoreeCloud Photos is a privacy-first, self-hosted photo and video platform for backup, synchronization, organization, memories, search, sharing, migration, preservation, export, and recovery.

**GoreeCloud Photos — powered by Keepsake.**

> **Development status:** Active native development. GoreeCloud-owned code under `native/` is the active product-development authority. The inherited Immich tree remains in this repository for provenance, compatibility research, migration, recovery, and rollback; it is not the approved long-term GoreeCloud application architecture and this repository is not yet production-ready.

## Product boundary

GoreeCloud Photos is the server-backed GoreeCloud photo and video platform. It is distinct from **GoreeCloud Gallery**, which remains a separate device-local and offline-first media viewer and manager. Gallery must not depend on a Photos server to perform its core local functions.

## Keepsake

Keepsake is the first-party GoreeCloud Photos capability family. It organizes native responsibilities including:

- Keepsake Core
- Keepsake Vault
- Keepsake Sync
- Keepsake Memories
- Keepsake Search
- Keepsake Vision
- Keepsake Studio
- Keepsake Share
- Keepsake Places
- Keepsake Import
- Keepsake Export
- Keepsake Recovery
- Keepsake Clients

Keepsake is a capability identity inside GoreeCloud Photos, not a separate application, service, repository, database authority, or permission boundary.

## Current native foundation

The GoreeCloud-owned native server foundation under `native/server/` currently includes:

- authoritative photo/video media records separated from rebuildable derived state;
- owner-scoped media repository reads;
- SHA-256 integrity metadata and duplicate-media rejection foundations;
- a versioned portable export manifest carrying original-media identity, portable paths, integrity information, and album relationships;
- Keepsake Sync state modeling for Pending, Uploading, Synced, and Failed lifecycle states;
- controlled retry/resynchronization transitions, required failure codes, upload-attempt tracking, and timestamp-ordering safeguards;
- focused Go tests and a dedicated Native Server GitHub Actions workflow.

The native foundation does not yet constitute a complete backup service. Persistent upload APIs, resumable transfers, retry scheduling, bandwidth policy, duplicate detection across the transfer pipeline, full import/export execution, search/intelligence, clients, and production deployment remain under development.

## Media authority and privacy

Original photos and videos are authoritative user-owned information. Application indexes, thumbnails, embeddings, caches, and other generated state must remain replaceable or reconstructable unless a future requirement explicitly promotes them to durable authority.

Privacy-sensitive media analysis is intended to be local-first. External analytics, advertising identifiers, behavioral tracking, and third-party usage profiling are not required product architecture.

Do not commit private media, credentials, signing material, tokens, keys, private storage paths, or user-specific media metadata to this repository.

## Development direction

GoreeCloud Photos is developed natively from the ground up under GoreeCloud-controlled application boundaries. Critical mature dependencies may be used only where narrowly justified. The retained Immich tree is transitional reference material for compatibility, migration, recovery, and provenance rather than the permanent GoreeCloud product authority.

See the GoreeCloud repository records under [`docs/goreecloud/`](docs/goreecloud/), including product, architecture, Keepsake identity, upstream provenance, privacy boundary, and native-contract documentation.

## Upstream provenance

This repository preserves its original Immich fork history and attribution. The retained upstream source remains subject to the applicable GNU Affero General Public License version 3 obligations.

See [`docs/goreecloud/UPSTREAM.md`](docs/goreecloud/UPSTREAM.md) and [`LICENSE`](LICENSE).

## Status

GoreeCloud Photos remains active development software. No production deployment, user-media migration, destructive synchronization cutover, or production-service acceptance is implied by the current native foundations.
