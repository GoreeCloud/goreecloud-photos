# GoreeCloud Photos — Feature Roadmap

**Status:** Active roadmap control  
**Lifecycle:** Experimental  
**As of:** 2026-09-18  
**Canonical specification:** SPECIFICATIONS.md  
**Repository:** GoreeCloud/goreecloud-photos  
**Implementation status:** PostgreSQL runtime and durable upload-session/part persistence validated; ordinary Photos upload, library functionality, and supported deployment remain unimplemented.

> Roadmap entries are planned work unless a later exact revision and evidence explicitly establish implementation and acceptance.

## Phase 0 — Product Foundation

**State:** Architecture/contract baseline plus first executable Experimental server foundation established.

Completed definition work:

- repository information architecture and required root controls;
- initial implementation stack and dependency-minimization policy;
- modular server/client architecture;
- stable Asset and ownership model;
- immutable-original and rebuildable-derivative rules;
- versioned API and resumable-upload contract;
- cursor-based synchronization, idempotency, revisions, tombstones, and conflict model;
- GoreeCloud Identity ownership/authorization boundary;
- Privacy Shield sensitive-processing boundary;
- Wardveil ingestion/security-evidence boundary;
- Everkeep recovery/preservation scope;
- GoreeCloud Manager, Mesh, Policy, and Observability integration architecture;
- Glaze UI 1.5.1 client target and accessibility boundary;
- machine-readable Asset, Sync Change, and Upload Session schemas;
- repository/contract baseline CI validation.

Verified implementation foundation:

- Go 1.27.1 module and server source tree;
- loopback-by-default health service;
- fail-closed readiness reporting;
- immutable filesystem original-media store adapter with SHA-256 verification and overwrite protection;
- initial PostgreSQL migration baseline;
- pgx v5.11.0 PostgreSQL runtime adapter;
- schema-aware database readiness that fails before the core migration and passes after the complete schema is present;
- disposable PostgreSQL 18.6 integration validation covering migration up/down and readiness;
- durable PostgreSQL upload-session and per-part receipt persistence with restart durability, out-of-order parts, identical-retry idempotency, conflicting-retry rejection, and deterministic expiry;
- Go unit/integration tests and pinned exact-head CI covering formatting, module reproducibility, vet, tests, migration validation, and binary build.

Remaining before Phase 1 can advance materially:

- implement the first resumable upload HTTP/API path and media-byte staging/assembly;
- complete upload finalization so verified immutable media, OriginalObject, Asset, and SyncChange state commit with explicit transactional rules;
- atomically create original-object/asset/synchronization records after verified media commit;
- add an S3-compatible storage backend behind the existing storage boundary;
- pin and document any newly introduced implementation dependencies;
- establish repository-local official Photos visual identity assets;
- resolve branch-protection administration using an authorized path;
- expand runtime/security/privacy/recovery evidence as capabilities are introduced.

**Exit condition:** a database-backed, restart-safe upload-to-asset vertical slice exists and is validated without inventing security, privacy, backup, or recovery claims.

## Phase 1 — Core Photo Server

**State:** Experimental foundation in progress.

Implement and verify:

- multi-user libraries;
- stable assets and ownership;
- immutable original media store;
- metadata store;
- resumable upload and transfer verification;
- ingestion validation;
- thumbnail and preview generation;
- basic authentication and authorization integration;
- trash and restore;
- basic Photos API;
- idempotent ingestion and job recovery.

## Phase 2 — Web Library

**State:** Planned.

Implement and verify:

- chronological timeline;
- asset viewer;
- upload;
- albums;
- favorites;
- download;
- basic date and metadata search;
- trash;
- responsive Glaze UI experience.

This is the first intended end-to-end server/client validation.

## Phase 3 — Mobile Backup

**State:** Planned.

Implement and verify:

- camera-library backup;
- background upload;
- selected-folder controls where supported;
- network, battery, charging, and roaming policies;
- resumable transfer recovery;
- backup-state truth;
- safe device-storage cleanup only after verified server durability;
- mobile timeline and offline/download behavior.

## Phase 4 — Desktop Library

**State:** Planned.

Implement and verify:

- folder imports;
- watch folders;
- large-library management;
- bulk operations;
- local cache;
- offline collections;
- large resumable imports;
- export and archival workflows;
- storage diagnostics.

## Phase 5 — Sharing and Family Libraries

**State:** Planned.

Implement and verify:

- direct user sharing;
- shared and collaborative albums;
- personal versus family ownership boundaries;
- household libraries;
- revocable and expiring links;
- password-protected and view/download-controlled links;
- metadata-removal options;
- explicit ownership preservation.

## Phase 6 — Optional Intelligence

**State:** Planned.

Implement as modular, privacy-controlled capabilities:

- object and scene recognition;
- OCR;
- document and screenshot detection;
- pet detection;
- face detection and person clustering;
- duplicate and near-duplicate analysis;
- semantic search;
- suggested albums and memories.

Disabling intelligence must not impair ordinary storage, browsing, export, or recovery.

## Phase 7 — Editing and Memories

**State:** Planned.

Implement and verify:

- non-destructive photo edits;
- edit history and synchronization;
- basic video trim/rotate/mute/color operations;
- memories;
- event clustering;
- generated collections;
- hide/dismiss controls for people, dates, and memory classes.

## Phase 8 — Advanced Privacy and Protected Media

**State:** Planned.

Develop and validate:

- protected media;
- reauthentication requirements;
- search, preview, widget, notification, and memory exclusion;
- sensitive metadata controls;
- secure sharing;
- Sealed Library design;
- client-controlled encryption only after the complete key, thumbnail, metadata, search, sharing, recovery, and succession lifecycle is solved.

## Phase 9 — Preservation, Portability, and Migration

**State:** Planned.

Complete and verify:

- selected and full-library export;
- portable machine-readable manifest;
- album and metadata reconstruction;
- migration from common photo-library exports;
- integrity auditing;
- tested clean-target restore;
- disaster recovery;
- rebuild procedures for derivatives and indexes;
- long-term migration;
- family succession.

## Phase 10 — Stable Acceptance

**State:** Planned.

Require exact evidence for:

- cross-platform behavior;
- accessibility;
- representative-device operation;
- performance and large-library behavior;
- security;
- privacy;
- recovery;
- upgrade and rollback;
- data portability;
- nine-system platform evaluation;
- Glaze UI 1.5.1 conformance;
- release artifact provenance;
- exact-release validation.

## MVP gate

The first usable MVP should include authentication, multi-user libraries, original storage, resumable upload, mobile camera backup, web/mobile timeline, desktop import, thumbnails, albums, favorites, basic metadata search, user-to-user sharing, trash/restore, backup-state indicators, offline/download support, import/export, Privacy Shield integration, Wardveil Security integration, and Everkeep recovery requirements.

Advanced intelligence is explicitly later than trusted storage, synchronization, privacy, security, and recovery foundations.
