# GoreeCloud Photos — Architecture

**Status:** Phase 0 architecture baseline  
**Lifecycle:** Concept / Planning  
**Implementation status:** Architecture defined; product runtime not yet implemented or accepted.  
**As of:** 2026-09-18

## 1. Architectural objective

GoreeCloud Photos is an original GoreeCloud-controlled photo and video platform. The architecture must preserve original media, keep user-owned information portable, support multiple native clients without creating competing authoritative stores, and keep privacy, security, recovery, and platform-system authority explicit.

The system is divided into five layers:

1. Native and web clients.
2. Photos API and domain services.
3. Background processing.
4. Persistent authoritative and rebuildable storage.
5. GoreeCloud platform integrations and recovery systems.

No optional intelligence, search accelerator, derivative generator, or third-party library may become the only path to original media.

## 2. Initial implementation stack

The Phase 0 implementation target is:

| Area | Initial target | Role |
|---|---|---|
| Server service plane | Go 1.27.1 | API, library logic, upload coordination, synchronization, sharing, administration adapters, worker orchestration |
| Authoritative relational store | PostgreSQL | Libraries, assets, ownership, albums, edits, shares, upload state, sync log, jobs, policy/evidence references |
| Original and derived media | Storage-driver abstraction | Filesystem backend for controlled single-node development; S3-compatible backend for scalable deployments |
| Structured search | PostgreSQL indexes and full-text capabilities first | Avoid a separate search dependency until measured need justifies one |
| Background work | PostgreSQL-backed job queue first | Avoid Redis, Kafka, NATS, or another broker in the initial core unless evidence establishes a requirement |
| Web client | TypeScript + React | First-class web application using the same Photos API |
| Mobile A | Native Android: Kotlin + Jetpack Compose | Required GoreeCloud Android client; may consume shared Android platform libraries |
| Desktop | Native Linux application; Rust + GTK4 target | Required Linux graphical client with native lifecycle, file, accessibility, and desktop integration |
| Mobile B | Reserved native second-mobile target | Platform remains intentionally unassigned until authoritative product scope names it |
| Offline client state | Platform-appropriate local database; SQLite-compatible model where appropriate | Cache, offline library subset, pending operations; never a competing server authority |

Exact package versions other than the already selected Go toolchain must be pinned when source is introduced. A version mentioned in planning does not establish an installed or supported dependency.

## 3. Client architecture

All clients use one Photos domain model and one authoritative server API.

### Web

The web client is a first-class application, not the implementation source for native clients. It may use browser storage for drafts, cached metadata, thumbnails, and pending operations, but server-owned library state remains authoritative.

### Mobile A — Android

Android is the first mandatory mobile target. It must use Android-native background work, media permissions, notifications, lifecycle, accessibility, and device-library interfaces. Camera backup must use platform-supported durable background scheduling and must not claim completion before server verification.

### Desktop — Linux

The Linux desktop client is the first desktop implementation target. It must support native file selection, watch folders, drag-and-drop, notifications, accessibility, offline collections, export, and large-library workflows.

### Mobile B

The repository keeps a separate Mobile B boundary without assuming the operating system. When the platform is authorized, it must be implemented natively for that platform rather than as a wrapper around the web application.

## 4. Server domain boundaries

The initial server is one deployable GoreeCloud Photos service with internal modules. These modules are architectural boundaries, not a requirement to deploy a microservice per module.

### Photos API

Owns versioned external application contracts, request validation, authentication handoff, authorization enforcement, idempotency, pagination, error envelopes, and compatibility.

### Ingestion

Owns untrusted-media admission, upload finalization, checksum verification, quarantine/validation state, media-type validation, and promotion into authoritative original storage.

### Upload Coordinator

Owns resumable upload sessions, part receipt, retry safety, expiration, final checksum calculation, and atomic completion.

### Library

Owns libraries, assets, ownership, archive/trash state, favorites, albums, membership, captions, user descriptions, and library-level relationships.

### Media Processing

Owns rebuildable thumbnails, previews, transcodes, optimized copies, and edit-render derivatives. It must never overwrite the original.

### Metadata

Owns extracted technical metadata and user-editable metadata projections while preserving the immutable original payload.

### Search

Owns structured query execution and, later, optional semantic retrieval. Search indexes are derived state and must be rebuildable.

### Intelligence

Owns optional object, scene, OCR, quality, similarity, landmark, and semantic analysis. Intelligence must be independently disableable and deletable without removing the original.

### People

Owns opt-in face observations, person clusters, user-assigned names, merge/split operations, exclusions, and recognition-derived data.

### Sharing

Owns direct grants, shared albums, collaborative relationships, family-library contribution rules, and external links. Authentication remains an Identity responsibility; Photos owns domain authorization.

### Memories

Owns derived rediscovery collections and dismissal/hide preferences. Memories are derived and must not become an authoritative storage dependency.

### Synchronization

Owns ordered changes, cursors, tombstones, device progress, conflict detection, and resumable incremental synchronization.

### Background Jobs

Owns durable asynchronous execution state. Job state is authoritative only for work orchestration; completed user data must remain recoverable independently from the queue.

## 5. Deployment model

The initial self-hosted server is a modular monolith plus worker processes sharing the same versioned domain contracts and PostgreSQL database.

Recommended initial deployment units:

- photos-api
- photos-worker
- PostgreSQL
- original-media storage
- derived-media storage

This avoids premature distributed-system complexity while preserving internal boundaries that can be separated later if scale or isolation requires it.

The server must support an OCI container image usable with Docker and Podman when implementation begins. Deployment state must remain separate from application source and user data.

## 6. Storage architecture

### Original Media Store

Immutable uploaded originals. An original object is never modified in place.

### Derived Media Store

Thumbnails, previews, optimized copies, transcodes, and rendered edits. All entries must be reproducible from authoritative originals plus authoritative edit/metadata records whenever practical.

### Metadata Store

PostgreSQL records for libraries, assets, ownership, albums, user metadata, edit instructions, shares, synchronization, job state, and policy/evidence references.

### Search Index

Initially PostgreSQL-backed and rebuildable. A later dedicated index is permitted only behind the Search boundary.

### Intelligence Store

Sensitive derived intelligence is logically separated from ordinary library metadata so it can be disabled, revoked, deleted, or rebuilt independently.

### Quarantine/Staging

Untrusted and incomplete uploads remain separate from trusted originals until required validation succeeds.

## 7. Original-media invariant

An original becomes authoritative only after:

1. The upload is complete.
2. The server has calculated the authoritative content checksum.
3. Required ingestion validation has passed.
4. The immutable media object has been durably committed.
5. The corresponding database transaction has committed.
6. The asset can be retrieved using its stable Asset ID.

Client-visible backup state must not advance beyond what this evidence establishes.

## 8. Dependency minimization

The initial architecture deliberately avoids:

- a required Redis deployment;
- a required message broker;
- a required external search engine;
- a required external AI provider;
- GoreeCloud Drive as the Photos database or authoritative media model;
- global cross-user content deduplication;
- wrapper-first native clients.

Optional foundational media tools may be introduced behind worker adapters only after provenance, licensing, security, update, replacement, and sandboxing review.

## 9. Failure and degradation model

Loss of an optional component must degrade safely:

- Search unavailable → browsing by authoritative library state continues.
- Intelligence unavailable → storage, browsing, export, and ordinary search continue.
- Thumbnail corruption → regenerate from original.
- Worker outage → queued jobs remain durable.
- Mesh unavailable → core Photos workflows continue unless a specific cross-application action requires Mesh.
- Observability unavailable → the product does not invent a healthy state.
- Privacy, Identity, Policy, or security evidence required for a sensitive operation unavailable → the sensitive operation fails closed.
- Everkeep unavailable → Photos does not claim recovery readiness; ordinary access to existing media may continue according to local policy.

## 10. Data authority

Authoritative user state belongs to clearly identified Photos domain records and immutable original objects.

Clients may keep local caches and offline subsets, but synchronization never makes a client cache the hidden authority for server-owned library state.

Detected people must never automatically become GoreeCloud Identity users. A face/person identity and an account identity remain separate domains.

## 11. Security boundary

Uploaded media is untrusted input. Parsing, metadata extraction, preview generation, and transcoding are worker operations with bounded resources and restricted filesystem/network authority.

Share-link secrets are never stored in reversible plaintext when a verifier representation can be used.

Administrative interfaces must not expose raw personal media merely to report operational health.

## 12. Architecture evolution

A module may become an independently deployed service only when one or more of these conditions are demonstrated:

- materially different scaling characteristics;
- stronger isolation requirement;
- independent failure domain;
- separate lifecycle or ownership boundary;
- technology requirement that cannot be cleanly satisfied in-process;
- measured performance requirement.

Separation must not weaken portability, authorization, recovery, or evidence traceability.
