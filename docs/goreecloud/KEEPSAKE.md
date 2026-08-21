# Keepsake — GoreeCloud Photos feature identity

## Identity

**Keepsake** is the official first-party feature umbrella for GoreeCloud Photos.

Preferred expression: **GoreeCloud Photos — powered by Keepsake**.

Keepsake names the coherent family of GoreeCloud-owned photo and video capabilities. It does not replace the GoreeCloud Photos application name and it does not replace shared GoreeCloud identities such as Glaze UI, Wardveil Security, Privacy Shield, or Everkeep.

## Purpose

Keepsake reflects the purpose of GoreeCloud Photos: preserve meaningful personal and family media while keeping authoritative originals private, portable, recoverable, understandable, and under user control.

The name intentionally covers both everyday photo-library experiences and the lower-level first-party capabilities required to sustain them over time.

## Capability family

- **Keepsake Core** — users, libraries, assets, metadata, albums, favorites, archive state, Trash, permissions, and primary Photos behavior.
- **Keepsake Vault** — authoritative original-media storage, storage abstraction, integrity, placement, lifecycle handling, and durable-file access.
- **Keepsake Sync** — device backup, upload queues, resumable transfers, duplicate detection, synchronization state, retry behavior, bandwidth controls, and ingestion.
- **Keepsake Memories** — anniversaries, on-this-day experiences, trips, contextual collections, people-oriented memories, and rediscovery.
- **Keepsake Search** — metadata, filename, date, location, indexed, semantic, and visual discovery.
- **Keepsake Vision** — local facial recognition, object recognition, embeddings, OCR, classification, similarity analysis, and other privacy-sensitive media intelligence.
- **Keepsake Studio** — non-destructive editing, adjustments, transformations, video-editing foundations, edit history, and restoration of originals.
- **Keepsake Share** — albums, family and partner sharing, controlled links, expiration controls, download permissions, and sharing-policy enforcement.
- **Keepsake Places** — maps, geographic search, locations, trips, clusters, and location-oriented organization.
- **Keepsake Import** — Google Photos Takeout, Ente, Immich, filesystem, camera, removable-media, phone-media, archive, and portable-export ingestion.
- **Keepsake Export** — portable original-media and metadata export independent of continued Photos operation.
- **Keepsake Recovery** — integrity validation, rebuild tooling, backup verification, database/index reconstruction support, and restoration workflows.
- **Keepsake Clients** — web, Android, Linux desktop, command-line tooling, and future iOS access.

## Shared GoreeCloud identities

Keepsake participates in the broader GoreeCloud platform rather than replacing shared platform identities:

- **GoreeCloud Photos** — application and product identity.
- **Keepsake** — Photos feature-family identity.
- **Glaze UI** — interface and design language.
- **Wardveil Security** — shared security identity and security-control patterns.
- **Privacy Shield** — privacy foundation and user-facing privacy controls.
- **Everkeep** — broader preservation, backup, recovery, and resilience identity where applicable.

## Product-boundary rule

Keepsake applies to GoreeCloud Photos. GoreeCloud Gallery remains the independent offline-first Android local-media application. Optional Gallery/Photos interoperability must not make Gallery dependent on a Photos server, account, or network connection for its core role.

## Fork-to-native rule

New first-party Photos capabilities should use Keepsake naming at product and architecture boundaries while preserving compatibility with inherited Immich internals during the transition. Broad mechanical renaming inside inherited code is not required and should not take priority over safe replacement boundaries, tests, migrations, or data integrity.
